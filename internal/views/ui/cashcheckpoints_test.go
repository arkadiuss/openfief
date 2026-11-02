package ui

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
)

// The fake core's refusals: an amount that isn't a decimal, and a
// checkpoint with nothing to book.
var (
	errNotDecimal = errors.New(`malformed cash checkpoint: amount "1,5"`)
	errMatches    = errors.New("the cash checkpoint matches the derived cash")
)

// fakeCheckpoints is the core's CashCheckpointService: it records what the
// owner sent, refuses an amount of "1,5" and a booking for any date but
// 2026-03-31.
type fakeCheckpoints struct {
	ledgerv1connect.UnimplementedCashCheckpointServiceHandler

	mu       sync.Mutex
	recorded []*ledgerv1.RecordCashCheckpointRequest
	booked   []*ledgerv1.BookCashDifferenceRequest
}

func (f *fakeCheckpoints) RecordCashCheckpoint(_ context.Context, r *connect.Request[ledgerv1.RecordCashCheckpointRequest]) (*connect.Response[ledgerv1.RecordCashCheckpointResponse], error) {
	m := r.Msg
	if m.GetOrigin() == "" || m.GetSessionToken() != ownerSession {
		return nil, codeErr(connect.CodePermissionDenied, nil)
	}
	if m.GetAmount() == "1,5" {
		return nil, codeErr(connect.CodeInvalidArgument, errNotDecimal)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.recorded = append(f.recorded, m)
	return connect.NewResponse(&ledgerv1.RecordCashCheckpointResponse{}), nil
}

func (f *fakeCheckpoints) BookCashDifference(_ context.Context, r *connect.Request[ledgerv1.BookCashDifferenceRequest]) (*connect.Response[ledgerv1.BookCashDifferenceResponse], error) {
	m := r.Msg
	if m.GetOrigin() == "" || m.GetSessionToken() != ownerSession {
		return nil, codeErr(connect.CodePermissionDenied, nil)
	}
	if m.GetDate() != "2026-03-31" {
		return nil, codeErr(connect.CodeFailedPrecondition, errMatches)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.booked = append(f.booked, m)
	return connect.NewResponse(&ledgerv1.BookCashDifferenceResponse{Kind: "deposit", Amount: "20.5", Currency: "EUR"}), nil
}

// formPost is a form posted to path.
func formPost(path string, form url.Values) *http.Request {
	r := request(http.MethodPost, path, form.Encode())
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

// TestCashCheckpointsOnTheAccountPage checks D-158 on the account page:
// the owner's checkpoints with their differences, the one that differs
// offering to book it as a deposit and the matching one nothing; the
// owner's form for a new checkpoint; and a read session seeing the
// checkpoints without any form.
func TestCashCheckpointsOnTheAccountPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `id="cash-checkpoints"`, "Cash checkpoints", "2026-03-31", "120,50 EUR", "20,50 EUR", "differs",
		`action="/accounts/acc-synthetic-1/cash-checkpoints/book"`, "Book as a deposit", "2026-02-28", "matches", "nothing to book",
		`action="/accounts/acc-synthetic-1/cash-checkpoints"`, `name="date" value="2026-04-01"`, "Enter cash checkpoint", "cash checkpoint</span>")
	rec = h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Cash checkpoints", "20,50 EUR", "differs")
	wantNoBody(t, rec, "Book as a deposit", "Enter cash checkpoint")
}

// TestCashCheckpointForms checks that the owner's checkpoint and booking
// reach the core as posted, each back to the account's checkpoints; that a
// refusal shows the core's reason with the form as posted; that a read
// session is refused before the core; and that a malformed form never
// reaches it.
func TestCashCheckpointForms(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	path := "/accounts/" + accountID + "/cash-checkpoints"
	rec := h.serve(t, withSession(formPost(path, url.Values{"date": {"2026-03-31"}, "amount": {"120.50"}, "currency": {"eur"}}), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/accounts/acc-synthetic-1#cash-checkpoints" {
		t.Errorf("Location = %q", got)
	}
	rec = h.serve(t, withSession(formPost(path+"/book", url.Values{"date": {"2026-03-31"}, "currency": {"EUR"}}), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)

	rec = h.serve(t, withSession(formPost(path, url.Values{"date": {"2026-03-31"}, "amount": {"1,5"}, "currency": {"EUR"}}), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The cash checkpoint was not recorded: malformed cash checkpoint", `value="1,5"`)
	rec = h.serve(t, withSession(formPost(path+"/book", url.Values{"date": {"2026-02-28"}, "currency": {"EUR"}}), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The difference was not booked: the cash checkpoint matches the derived cash")

	rec = h.serve(t, withSession(formPost(path, url.Values{"date": {"2026-03-31"}, "amount": {"1"}, "currency": {"EUR"}}), readSession))
	wantStatus(t, rec, http.StatusForbidden)
	rec = h.serve(t, withSession(formPost(path, url.Values{"date": {"2026-03-31"}}), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	rec = h.serve(t, withSession(formPost(path+"/book", url.Values{"currency": {"EUR"}}), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)

	h.checkpoints.mu.Lock()
	defer h.checkpoints.mu.Unlock()
	if len(h.checkpoints.recorded) != 1 || h.checkpoints.recorded[0].GetCurrency() != "EUR" || h.checkpoints.recorded[0].GetAmount() != "120.50" {
		t.Errorf("recorded = %v, want the one 120.50 EUR checkpoint", h.checkpoints.recorded)
	}
	got := make([]string, 0, len(h.checkpoints.booked))
	for _, b := range h.checkpoints.booked {
		got = append(got, b.GetAccountId()+" "+b.GetDate()+" "+b.GetCurrency())
	}
	if diff := cmp.Diff([]string{"acc-synthetic-1 2026-03-31 EUR"}, got); diff != "" {
		t.Errorf("booked (-want +got):\n%s", diff)
	}
}
