package app_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// errRead is a file that breaks off while it is read.
var errRead = errors.New("synthetic: the upload broke off")

// brokenReader returns its bytes, then errRead.
type brokenReader struct {
	r io.Reader
}

func (b *brokenReader) Read(p []byte) (int, error) {
	n, err := b.r.Read(p)
	if err != nil {
		return n, errRead
	}
	return n, nil
}

// uploads returns Uploads over the fakes, with the instance flex-main bound
// to acc-1, a 1000-byte limit, batch IDs b-1, b-2, ... and a sniffer that
// names how many bytes it saw.
func uploads(t *testing.T, caller *fakeCaller, l *fakeLedger, rec *fakeRecorder) *app.Uploads {
	t.Helper()
	n := 0
	return &app.Uploads{
		Owners: fakeOwners{err: app.ErrNoSession}, Bindings: fakeAccountBindings{instances: []string{"flex-main"}},
		Connectors: connectors(t, caller, l, &fakeStores{}), Batches: rec, MaxBytes: 1000,
		NewBatchID: func() ingest.BatchID { n++; return ingest.BatchID(fmt.Sprintf("b-%d", n)) },
		Sniff:      func(head []byte) string { return fmt.Sprintf("test/%d-bytes", len(head)) },
	}
}

// fileDelivery is what a connector delivers for a statement of acc-1: a
// deposit, a cash assertion and one ingest problem.
func fileDelivery(t *testing.T) app.Delivery {
	t.Helper()
	return app.Delivery{Institution: "ibkr", Mapping: mapping(t)}
}

func TestUploadsUpload(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, fileName, contents string
		mode                     ingest.Mode
		wantName, wantMedia      string
	}{
		{
			name: "a small statement", fileName: "../statements/flex.xml", contents: "<FlexQueryResponse/>",
			mode: ingest.ModeTransactions, wantName: "flex.xml", wantMedia: "test/20-bytes",
		},
		{
			name: "a statement of exactly the limit", fileName: `C:\exports\flex.xml`, contents: strings.Repeat("x", 1000),
			mode: ingest.ModeOpening, wantName: "flex.xml", wantMedia: "test/512-bytes",
		},
		{
			name: "a name that is no name", fileName: "..", contents: "data",
			mode: ingest.ModeTransactions, wantName: "", wantMedia: "test/4-bytes",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			caller := &fakeCaller{delivery: fileDelivery(t)}
			l := &fakeLedger{summary: ingest.ImportSummary{New: 1}}
			rec := &fakeRecorder{}
			file := app.NamedFile{Name: tt.fileName, MediaType: "application/x-declared", Contents: strings.NewReader(tt.contents)}
			got, err := uploads(t, caller, l, rec).Upload(t.Context(), "https://ui.example.com", "owner", "acc-1", tt.mode, file)
			if err != nil {
				t.Fatalf("Upload: %v", err)
			}
			want := ingest.Batch{
				ID: "b-1", Institution: "ibkr", Source: app.SourceUpload, Mode: tt.mode, Started: march2(t),
				Transactions: 1, Assertions: 1, Imported: ingest.ImportSummary{New: 1}, Accounts: []ledger.AccountID{"acc-1"},
				PerAccount: []ingest.AccountCounts{{Account: "acc-1", Transactions: 1}},
				Problems:   []ingest.Problem{ingest.Unsupported("row 2", "broker interest")},
			}
			if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Upload batch (-want +got):\n%s", diff)
			}
			checkUploaded(t, caller, l, rec, &want, &handedFile{mode: tt.mode, file: tt.wantName + ":" + tt.contents, media: tt.wantMedia})
		})
	}
}

// handedFile is what the connector should have been handed.
type handedFile struct {
	mode        ingest.Mode
	file, media string
}

// checkUploaded checks that the connector got the file as flex-main's batch
// b-1, that the ledger imported it once with keys scoped to the instance,
// seeding only an opening batch, and that the batch was recorded once.
func checkUploaded(t *testing.T, caller *fakeCaller, l *fakeLedger, rec *fakeRecorder, want *ingest.Batch, h *handedFile) {
	t.Helper()
	handed := fakeCaller{instance: "flex-main", id: "b-1", mode: h.mode, file: h.file, media: h.media}
	if diff := cmp.Diff(handed, *caller, cmp.AllowUnexported(fakeCaller{}), cmpopts.IgnoreFields(fakeCaller{}, "delivery")); diff != "" {
		t.Errorf("handed to the connector (-want +got):\n%s", diff)
	}
	if l.calls != 1 || l.seeds != boolInt(h.mode == ingest.ModeOpening) || l.txs[0].Key.Source != "plugin:flex-main:test" {
		t.Errorf("ledger calls = %d, seeds = %d, keys %v; want one import, seeding only an opening batch, keys scoped to the instance",
			l.calls, l.seeds, l.txs)
	}
	if diff := cmp.Diff([]ingest.Batch{*want}, rec.batches, cmpopts.EquateEmpty()); diff != "" || rec.instances[0] != "flex-main" {
		t.Errorf("recorded for %v (-want +got):\n%s", rec.instances, diff)
	}
}

// boolInt is 1 for true.
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// uploadRefusal is one way Uploads.Upload refuses a file.
type uploadRefusal struct {
	name      string
	session   string
	owners    error
	mode      ingest.Mode
	contents  io.Reader
	instances []string
	bindErr   error
	delivered ledger.AccountID
	callErr   error
	recordErr error
	wantErr   error
	// imported is whether the ledger is still called.
	imported bool
}

// fine is an upload refusal's starting point: everything as in a successful
// upload.
func fine(name string) *uploadRefusal {
	return &uploadRefusal{
		name: name, session: "owner", mode: ingest.ModeTransactions, contents: strings.NewReader("statement"),
		instances: []string{"flex-main"}, delivered: "acc-1",
	}
}

// but returns r changed by change.
func but(r *uploadRefusal, change func(r *uploadRefusal)) *uploadRefusal {
	change(r)
	return r
}

func TestUploadsUploadRefuses(t *testing.T) {
	t.Parallel()
	tests := []*uploadRefusal{
		but(fine("no session"), func(r *uploadRefusal) { r.session, r.owners, r.wantErr = "nope", app.ErrNoSession, app.ErrNoSession }),
		but(fine("a read-scoped session"), func(r *uploadRefusal) { r.session, r.owners, r.wantErr = "reader", app.ErrNotOwner, app.ErrNotOwner }),
		but(fine("an empty file"), func(r *uploadRefusal) { r.contents, r.wantErr = strings.NewReader(""), app.ErrInvalidUpload }),
		but(fine("a file one byte over the limit"), func(r *uploadRefusal) {
			r.contents, r.wantErr = strings.NewReader(strings.Repeat("x", 1001)), app.ErrFileTooLarge
		}),
		but(fine("a file over the limit within its first bytes"), func(r *uploadRefusal) {
			r.contents, r.wantErr = strings.NewReader(strings.Repeat("x", 5000)), app.ErrFileTooLarge
		}),
		but(fine("a file that breaks off"), func(r *uploadRefusal) {
			r.contents, r.wantErr = &brokenReader{r: strings.NewReader(strings.Repeat("x", 700))}, errRead
		}),
		but(fine("a file that breaks off at once"), func(r *uploadRefusal) {
			r.contents, r.wantErr = &brokenReader{r: strings.NewReader("x")}, errRead
		}),
		but(fine("no connector bound to the account"), func(r *uploadRefusal) { r.instances, r.wantErr = nil, app.ErrNoConnector }),
		but(fine("two connectors bound to the account"), func(r *uploadRefusal) {
			r.instances, r.wantErr = []string{"flex-a", "flex-b"}, app.ErrNoConnector
		}),
		but(fine("bindings that can't be read"), func(r *uploadRefusal) { r.bindErr, r.wantErr = errStore, errStore }),
		but(fine("the connector refuses the file"), func(r *uploadRefusal) {
			r.callErr, r.wantErr = fmt.Errorf("%w: not a Flex statement", app.ErrEntryRefused), app.ErrEntryRefused
		}),
		but(fine("the connector takes no files"), func(r *uploadRefusal) {
			r.callErr, r.wantErr = fmt.Errorf("%w: synthetic", app.ErrNoEntries), app.ErrNoEntries
		}),
		but(fine("the connector can't be reached"), func(r *uploadRefusal) {
			r.callErr, r.wantErr = fmt.Errorf("%w: synthetic", app.ErrConnectorUnavailable), app.ErrConnectorUnavailable
		}),
		but(fine("a delivery into another account"), func(r *uploadRefusal) { r.delivered, r.wantErr = "acc-other", ingest.ErrAccountNotBound }),
		but(fine("the history can't record it"), func(r *uploadRefusal) { r.recordErr, r.wantErr, r.imported = errStore, errStore, true }),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := fileDelivery(t)
			if tt.delivered != "acc-1" {
				d.Mapping.Assertions[0].Account = tt.delivered
			}
			caller := &fakeCaller{delivery: d, err: tt.callErr}
			l, rec := &fakeLedger{}, &fakeRecorder{err: tt.recordErr}
			u := uploads(t, caller, l, rec)
			u.Owners, u.Bindings = fakeOwners{err: tt.owners}, fakeAccountBindings{instances: tt.instances, err: tt.bindErr}
			_, err := u.Upload(t.Context(), "https://ui.example.com", tt.session, "acc-1", tt.mode, app.NamedFile{Name: "s.xml", Contents: tt.contents})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Upload error = %v, want %v", err, tt.wantErr)
			}
			if got := l.calls > 0; got != tt.imported || len(rec.batches) != 0 {
				t.Errorf("ledger imported = %v (want %v), recorded %v (want none)", got, tt.imported, rec.batches)
			}
		})
	}
}

// historyOf answers HasTransactionsFrom from a fixed set of accounts that
// hold transactions, as accounts started from their whole history do
// (D-136).
type historyOf []ledger.AccountID

func (h historyOf) HasTransactionsFrom(_ context.Context, account ledger.AccountID, _ string) (bool, error) {
	return slices.Contains(h, account), nil
}

// openingCash is broker account id's seeding: its opening cash of 100 USD
// on day, keyed as a connector keys it.
func openingCash(t *testing.T, id ingest.BrokerAccountID, day kernel.Date) ingest.BrokerEntries {
	t.Helper()
	usd, err := kernel.ParseMoney("100", mustCurrency(t, "USD"))
	if err != nil {
		t.Fatalf("ParseMoney: %v", err)
	}
	h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "flex", ExternalID: ledger.OpeningCashPrefix + string(id)}, Account: "named-by-broker-account", Date: day}
	dep, err := ledger.NewDeposit(h, usd)
	if err != nil {
		t.Fatalf("NewDeposit: %v", err)
	}
	dep.Account = ""
	return ingest.BrokerEntries{Account: id, Transactions: []ledger.Transaction{dep}}
}

// TestUploadsMultiAccountOpening checks D-140 on a statement of four
// linked broker accounts uploaded for acc-1, which hasn't started: the
// upload runs in acc-1's mode, opening, and seeds only the accounts not
// seeded yet. acc-2, started by an earlier opening upload (D-160), and
// acc-3, started from its whole history (D-136), keep their start: their
// seeding is left out as skipped problems, not refused. The broker account
// that isn't mapped is reported. The owner must be allowed to change every
// account the statement may book (D-163).
func TestUploadsMultiAccountOpening(t *testing.T) {
	t.Parallel()
	day := march2(t)
	d := app.Delivery{Institution: "ibkr", Mapping: ingest.Mapping{
		BrokerAccounts: []ingest.BrokerAccountID{"U0000001", "U0000002", "U0000003", "U0000004"},
		ByBrokerAccount: []ingest.BrokerEntries{
			openingCash(t, "U0000001", day), openingCash(t, "U0000002", day), openingCash(t, "U0000003", day),
			openingCash(t, "U0000004", day),
		},
	}}
	l := &fakeLedger{}
	u := uploads(t, &fakeCaller{delivery: d}, l, &fakeRecorder{})
	u.Owners, u.History = fakeOwners{}, historyOf{"acc-3"}
	u.Connectors.Bindings = &fakeBindings{accounts: []ledger.AccountID{"acc-1", "acc-2", "acc-3"}}
	u.Connectors.Starts = &memStarts{starts: map[ledger.AccountID]kernel.Date{"acc-2": day}}
	u.Connectors.Brokers = &fakeBrokers{accounts: []ingest.BrokerAccount{
		{Instance: "flex-main", ID: "U0000001", Account: "acc-1"},
		{Instance: "flex-main", ID: "U0000002", Account: "acc-2"},
		{Instance: "flex-main", ID: "U0000003", Account: "acc-3"},
		{Instance: "flex-main", ID: "U0000004"},
	}}
	b, err := u.Upload(t.Context(), "https://ui.example.com", "owner", "acc-1", 0, app.NamedFile{Name: "s.xml", Contents: strings.NewReader("flex")})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	booked := make([]ledger.AccountID, 0, len(l.txs))
	for i := range l.txs {
		booked = append(booked, l.txs[i].Account)
	}
	got := []any{b.Mode, booked, l.seeds, b.Unmapped}
	want := []any{ingest.ModeOpening, []ledger.AccountID{"acc-1"}, 1, []ingest.BrokerAccountID{"U0000004"}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mode, seeded accounts, seeds, unmapped (-want +got):\n%s", diff)
	}
	wantSkipped := []ingest.Problem{
		ingest.Skipped("flex:acc-2/"+ledger.OpeningCashPrefix+"U0000002", "the account is already seeded"),
		ingest.Skipped("flex:acc-3/"+ledger.OpeningCashPrefix+"U0000003", "the account is already seeded"),
	}
	if diff := cmp.Diff(wantSkipped, b.Problems); diff != "" {
		t.Errorf("problems (-want +got):\n%s", diff)
	}

	u.Owners = fakeOwners{denied: []ledger.AccountID{"acc-3"}}
	if _, err := u.Upload(t.Context(), "https://ui.example.com", "owner", "acc-1", 0, app.NamedFile{Name: "s.xml", Contents: strings.NewReader("flex")}); !errors.Is(err, app.ErrNotOwner) {
		t.Errorf("Upload of a statement that may book an account the owner may not change = %v, want ErrNotOwner", err)
	}
}
