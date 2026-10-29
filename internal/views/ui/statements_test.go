package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

var (
	errNotFlex    = errors.New("ibkr-flex: not a Flex statement")
	errFlexDown   = errors.New("the connector can't be reached")
	errNoHeaderUp = errors.New("no header")
)

// fakeStatements is an in-memory StatementService. Like the core, it wants
// the plugin's token and an owner-scoped session; it refuses a file that
// starts with "date," as the connector would, answers a file named
// down.xml as if the connector were down, and imports anything else as one
// batch with three ingest problems, listing two.
type fakeStatements struct {
	ingestv1connect.UnimplementedStatementServiceHandler

	mu      sync.Mutex
	bearers []string
	headers []*ingestv1.StatementHeader
	sizes   []int
	head    []byte
}

func (f *fakeStatements) UploadStatement(_ context.Context, stream *connect.ClientStream[ingestv1.UploadStatementRequest]) (*connect.Response[ingestv1.UploadStatementResponse], error) {
	f.locked(func() { f.bearers = append(f.bearers, stream.RequestHeader().Get("Authorization")) })
	if stream.RequestHeader().Get("Authorization") != "Bearer "+pluginToken {
		return nil, codeErr(connect.CodePermissionDenied, nil)
	}
	if !stream.Receive() || stream.Msg().GetHeader() == nil {
		return nil, codeErr(connect.CodeInvalidArgument, errNoHeaderUp)
	}
	h := stream.Msg().GetHeader()
	if err := owner(h.GetSessionToken()); err != nil {
		return nil, err
	}
	var head []byte
	size := 0
	for stream.Receive() {
		chunk := stream.Msg().GetChunk()
		size += len(chunk)
		if len(head) < 64 {
			head = append(head, chunk[:min(len(chunk), 64-len(head))]...)
		}
	}
	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("fake core: %w", err)
	}
	f.locked(func() { f.headers, f.sizes, f.head = append(f.headers, h), append(f.sizes, size), head })
	switch {
	case bytes.HasPrefix(head, []byte("date,")):
		return nil, codeErr(connect.CodeInvalidArgument, errNotFlex)
	case h.GetFileName() == "down.xml":
		return nil, codeErr(connect.CodeUnavailable, errFlexDown)
	}
	return connect.NewResponse(&ingestv1.UploadStatementResponse{
		BatchId: "b-7", Institution: "ibkr", Transactions: 10, Assertions: 4, New: 9, Unchanged: 1, FailedAssertions: 1,
		ProblemsTotal: 3, Problems: []*ingestv1.Problem{
			{Kind: ingestv1.ProblemKind_PROBLEM_KIND_UNSUPPORTED, Ref: "FlexStatement#1/Trades#8[tradeID=9000106]", Reason: `asset category "OPT" is not supported`},
			{Kind: ingestv1.ProblemKind_PROBLEM_KIND_SKIPPED, Ref: "FlexStatement#2", Reason: "statement of an IBKR account that feeds no ledger account"},
		},
	}), nil
}

// locked runs fn with the fake's lock held.
func (f *fakeStatements) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

// uploadPart is one part of a posted upload form: a field, or a file when
// fileName is set.
type uploadPart struct {
	name, fileName string
	body           io.Reader
}

// uploadRequest posts parts as a multipart form to the account's
// statements, streamed so a large body is never held in memory.
func uploadRequest(parts ...uploadPart) *http.Request {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		for _, p := range parts {
			var w io.Writer
			var err error
			if p.fileName != "" {
				w, err = mw.CreateFormFile(p.name, p.fileName)
			} else {
				w, err = mw.CreateFormField(p.name)
			}
			if err == nil {
				_, err = io.Copy(w, p.body)
			}
			if err != nil {
				pw.CloseWithError(err)
				return
			}
		}
		pw.CloseWithError(mw.Close())
	}()
	r := request(http.MethodPost, "/accounts/"+accountID+"/statements", "")
	r.Body = pr
	r.ContentLength = -1
	r.Header.Set("Content-Type", mw.FormDataContentType())
	return r
}

// modePart and filePart are the form's two parts.
func modePart(mode string) uploadPart { return uploadPart{name: "mode", body: strings.NewReader(mode)} }

func filePart(name string, body io.Reader) uploadPart {
	return uploadPart{name: "statement", fileName: name, body: body}
}

// flexLike is a synthetic statement of n bytes that starts like a Flex one.
func flexLike(n int) io.Reader {
	return io.MultiReader(strings.NewReader("<FlexQueryResponse>"), io.LimitReader(zeros{}, int64(n-len("<FlexQueryResponse>"))))
}

// zeros reads as endless zero bytes.
type zeros struct{}

func (zeros) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}

// TestAccountUploadForm checks that only the owner sees the "Upload
// statement" form, on any account's page, as a multipart form that sends
// the mode before the file.
func TestAccountUploadForm(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `<h2 id="upload-title">Upload statement</h2>`, `action="/accounts/`+accountID+`/statements"`,
		`enctype="multipart/form-data"`, `name="statement" type="file"`)
	if body := rec.Body.String(); strings.Index(body, `name="mode"`) > strings.Index(body, `name="statement"`) {
		t.Error("the form sends the file before the mode; the UI reads the mode first")
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, "Upload statement")
}

// TestUploadStatement checks that the posted file is streamed to the core
// whole, in chunks, with the plugin's token, the owner's session, the
// account, the file's name and the mode, and that the page shows the
// import's counts and ingest problems.
func TestUploadStatement(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(uploadRequest(modePart("opening"), filePart("flex.xml", flexLike(300000))), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Statement imported", "Import batch b-7 from ibkr: 10 transactions (9 new, 0 new versions, 1 unchanged)",
		"4 balance assertions, 1 of them flagged", "3 rows were left out of the ledger; the first 2 are listed",
		"FlexStatement#1/Trades#8[tradeID=9000106]", "asset category &#34;OPT&#34; is not supported", "<td>skipped</td>")
	var headers []*ingestv1.StatementHeader
	var sizes []int
	h.statements.locked(func() { headers, sizes = h.statements.headers, h.statements.sizes })
	want := []*ingestv1.StatementHeader{{
		Origin: testOrigin, SessionToken: ownerSession, AccountId: accountID, FileName: "flex.xml", Mode: ingestv1.Mode_MODE_OPENING,
	}}
	if diff := cmp.Diff(want, headers, protocmp.Transform()); diff != "" || len(sizes) != 1 || sizes[0] != 300000 {
		t.Errorf("uploads handed to the core: %v bytes (want 300000) (-want +got):\n%s", sizes, diff)
	}
	if bearers := seen(h.statements, func() []string { return h.statements.bearers }); bearers[0] != "Bearer "+pluginToken {
		t.Errorf("StatementService callers = %v, want the plugin's token", bearers)
	}
}

// TestUploadStatementRefusals checks the core's refusals and the forms the
// UI refuses itself, which never reach the core.
func TestUploadStatementRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		session string
		parts   []uploadPart
		status  int
		body    string
		reached bool
	}{
		{
			name: "the connector refuses the file", session: ownerSession, reached: true,
			parts:  []uploadPart{modePart("transactions"), filePart("export.csv", strings.NewReader("date,amount\n"))},
			status: http.StatusBadRequest, body: "The statement was not imported: ibkr-flex: not a Flex statement",
		},
		{
			name: "the connector is down", session: ownerSession, reached: true,
			parts:  []uploadPart{modePart("transactions"), filePart("down.xml", flexLike(100))},
			status: http.StatusServiceUnavailable, body: "its connector can&#39;t be reached now",
		},
		{
			name: "a read session", session: readSession,
			parts:  []uploadPart{modePart("transactions"), filePart("flex.xml", flexLike(100))},
			status: http.StatusForbidden, body: "Only the owner",
		},
		{
			name: "the file before the mode", session: ownerSession,
			parts:  []uploadPart{filePart("flex.xml", flexLike(100)), modePart("transactions")},
			status: http.StatusBadRequest, body: "The upload form is malformed",
		},
		{
			name: "an unknown mode", session: ownerSession,
			parts:  []uploadPart{modePart("everything"), filePart("flex.xml", flexLike(100))},
			status: http.StatusBadRequest, body: "The upload form is malformed",
		},
		{
			name: "no file", session: ownerSession,
			parts:  []uploadPart{modePart("transactions")},
			status: http.StatusBadRequest, body: "The upload form is malformed",
		},
		{
			name: "a file over 64 MiB", session: ownerSession, reached: true,
			parts:  []uploadPart{modePart("transactions"), filePart("big.xml", flexLike(maxUploadBody+1))},
			status: http.StatusRequestEntityTooLarge, body: "the file is larger than 64 MiB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			rec := h.serve(t, withSession(uploadRequest(tt.parts...), tt.session))
			wantStatus(t, rec, tt.status)
			wantBody(t, rec, tt.body)
			if reached := len(seen(h.statements, func() []string { return h.statements.bearers })) > 0; reached != tt.reached {
				t.Errorf("reached the core = %v, want %v", reached, tt.reached)
			}
			var sizes []int
			h.statements.locked(func() { sizes = h.statements.sizes })
			if tt.status == http.StatusRequestEntityTooLarge && len(sizes) != 0 {
				t.Errorf("the core completed an upload of %v bytes; want the upload stopped", sizes)
			}
		})
	}
}
