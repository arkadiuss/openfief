package connectrpc_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"testing"

	"google.golang.org/protobuf/encoding/protowire"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
)

// heapCaseEnv names the answer TestAnswerHeapChild serves, in the child
// process TestAnswerHeap starts for it.
const heapCaseEnv = "OPENFIEF_ANSWER_HEAP_CASE"

// answerHeapCases are hostile and largest-allowed connector answers, each
// with how far the core's heap may grow while it handles one, with Go's
// default GC: the core's memory limit is 256 MiB (deploy/core). Measured:
// about 31, 63 and 147 MiB. Before the answer was counted on the wire, the
// first case, at the old 32 MiB byte cap, grew the heap by 1.8 GiB.
func answerHeapCases() []answerHeapCase {
	return []answerHeapCase{
		{
			// The whole byte cap of empty postings, two bytes each.
			name:    "empty postings filling the byte cap",
			answer:  func() []byte { return emptyPostings((connectrpc.MaxDeliveryBytes - 64) / 2) },
			wantErr: connectrpc.ErrDeliveryTooLarge,
			maxMiB:  64,
		},
		{
			// As many empty postings as the wire count lets through: decoded,
			// then refused by the item cap.
			name:    "most empty postings decoded",
			answer:  func() []byte { return emptyPostings(connectrpc.MaxDeliveryMessages - 3) },
			wantErr: connectrpc.ErrDeliveryTooLarge,
			maxMiB:  96,
		},
		{
			// The largest price backfill the core takes, decoded and translated.
			name:   "largest price backfill",
			answer: func() []byte { return prices(connectrpc.MaxDeliveryItems) },
			maxMiB: 192,
		},
	}
}

// answerHeapCase is one connector answer and the most the core's heap may
// grow handling it.
type answerHeapCase struct {
	name    string
	answer  func() []byte
	wantErr error
	maxMiB  uint64
}

// TestAnswerHeap: whatever a connector answers, the core's heap grows a
// bounded amount handling it (D-184). Each case runs alone in a child
// process, so no other test's allocations count.
func TestAnswerHeap(t *testing.T) {
	t.Parallel()
	self, err := os.Executable()
	if err != nil {
		t.Fatalf("test binary: %v", err)
	}
	for _, tc := range answerHeapCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.CommandContext(t.Context(), self, "-test.run=^TestAnswerHeapChild$", "-test.count=1", "-test.v")
			cmd.Env = append(os.Environ(), heapCaseEnv+"="+tc.name)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("child: %v\n%s", err, out)
			}
			at := bytes.Index(out, []byte("heap-growth="))
			if at < 0 {
				t.Fatalf("child printed no heap growth:\n%s", out)
			}
			var grown uint64
			if _, scanErr := fmt.Sscanf(string(out[at:]), "heap-growth=%d", &grown); scanErr != nil {
				t.Fatalf("child output %q: %v", out, scanErr)
			}
			t.Logf("%s: peak heap grew %d MiB", tc.name, grown>>20)
			if grown > tc.maxMiB<<20 {
				t.Errorf("handling the answer grew the heap %d MiB, want at most %d MiB", grown>>20, tc.maxMiB)
			}
		})
	}
}

// TestAnswerHeapChild serves the case heapCaseEnv names to a sync and
// logs how far the heap grew during it. Outside TestAnswerHeap's child process
// it has nothing to do.
func TestAnswerHeapChild(t *testing.T) {
	t.Parallel()
	name := os.Getenv(heapCaseEnv)
	if name == "" {
		return
	}
	for _, tc := range answerHeapCases() {
		if tc.name != name {
			continue
		}
		body := tc.answer()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/proto")
			w.Header().Set("Content-Length", strconv.Itoa(len(body)))
			if _, err := w.Write(body); err != nil {
				return // the client stopped reading
			}
		}))
		t.Cleanup(srv.Close)
		c := connectrpc.NewClient(srv.Client(), func(string) (string, error) { return srv.URL, nil }, fakeTokens{})
		var err error
		grown := heapGrowth(func() {
			_, err = c.Sync(t.Context(), "md-main", &app.SyncCall{ID: "b-1", Mode: ingest.ModeTransactions})
		})
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("Sync error = %v, want %v", err, tc.wantErr)
		}
		t.Logf("heap-growth=%d", grown)
		return
	}
	t.Fatalf("no case %q", name)
}

// heapGrowth returns how much heap the process obtained while f ran: the
// runtime never returns heap address space, so the growth of HeapSys is
// the peak heap f needed beyond what was free before it. Stacks come out
// of the heap, so they are counted with it.
func heapGrowth(f func()) uint64 {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)
	end, start := after.HeapSys+after.StackSys, before.HeapSys+before.StackSys
	return max(end, start) - start
}

// emptyPostings is a sync answer for batch b-1 whose one transaction holds
// n empty postings.
func emptyPostings(n int) []byte {
	return syncAnswer(3, field(nil, 7, nil), n)
}

// prices is a sync answer for batch b-1 of n copies of one valid price.
func prices(n int) []byte {
	instrument := field(field(nil, 1, []byte("PLTEST000010")), 2, []byte("TST"))
	closing := field(field(nil, 1, []byte("12.34")), 2, []byte("PLN"))
	p := field(field(field(nil, 1, instrument), 2, []byte("2024-01-02")), 3, closing)
	return syncAnswer(0, field(nil, 6, p), n)
}

// field appends field num of wire type bytes holding v to b.
func field(b []byte, num protowire.Number, v []byte) []byte {
	return protowire.AppendBytes(protowire.AppendTag(b, num, protowire.BytesType), v)
}

// syncAnswer is a SyncResponse whose delivery, for batch b-1, holds n
// copies of item, nested in one message of field wrap unless wrap is zero.
// It is written once into a buffer of its exact size, so building it leaves
// no garbage for the measured call to reuse.
func syncAnswer(wrap protowire.Number, item []byte, n int) []byte {
	batch := field(nil, 1, []byte("b-1"))
	items := len(item) * n
	delivery := len(batch) + items
	if wrap != 0 {
		delivery += protowire.SizeTag(wrap) + protowire.SizeBytes(items) - items
	}
	b := make([]byte, 0, protowire.SizeTag(1)+protowire.SizeBytes(delivery))
	b = protowire.AppendVarint(protowire.AppendTag(b, 1, protowire.BytesType), size(delivery))
	b = append(b, batch...)
	if wrap != 0 {
		b = protowire.AppendVarint(protowire.AppendTag(b, wrap, protowire.BytesType), size(items))
	}
	for range n {
		b = append(b, item...)
	}
	return b
}

// size is n, a length, as a varint's value.
func size(n int) uint64 {
	if n < 0 {
		return 0
	}
	return uint64(n)
}

// TestClientTooLargeIsNotRefusal: an entry call whose answer the core
// refuses to decode as too large fails with ErrDeliveryTooLarge, not as
// the connector refusing the entries.
func TestClientTooLargeIsNotRefusal(t *testing.T) {
	t.Parallel()
	body := emptyPostings(connectrpc.MaxDeliveryMessages)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/proto")
		if _, err := w.Write(body); err != nil {
			return // the client stopped reading
		}
	}))
	t.Cleanup(srv.Close)
	c := connectrpc.NewClient(srv.Client(), func(string) (string, error) { return srv.URL, nil }, fakeTokens{})
	_, err := c.ImportBondLots(t.Context(), "bonds-main", "b-1", &ingest.BondLotEntries{Account: "acc-1"})
	if !errors.Is(err, connectrpc.ErrDeliveryTooLarge) || errors.Is(err, app.ErrEntryRefused) {
		t.Errorf("ImportBondLots error = %v, want %v and not %v", err, connectrpc.ErrDeliveryTooLarge, app.ErrEntryRefused)
	}
}
