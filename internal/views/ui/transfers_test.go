package ui

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"

	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// errTaken is the fake core's reason for refusing the taken pair.
var errTaken = errors.New("test/d1 is already a leg of a confirmed transfer")

// fakeTransfers is an in-memory TransferService over one synthetic pair,
// test/w1 -> test/d1; any other withdrawal is refused as taken.
type fakeTransfers struct {
	ledgerv1connect.UnimplementedTransferServiceHandler

	mu sync.Mutex
	// decision is the pair's current decision; unspecified while it is a
	// suggestion.
	decision ledgerv1.TransferDecision
	// sessions are the session tokens of the decisions, in order.
	sessions []string
}

func (f *fakeTransfers) ConfirmTransfer(_ context.Context, r *connect.Request[ledgerv1.ConfirmTransferRequest]) (*connect.Response[ledgerv1.ConfirmTransferResponse], error) {
	m := r.Msg
	l, err := f.decide(m.GetOrigin(), m.GetSessionToken(), m.GetWithdrawal(), m.GetDeposit(), ledgerv1.TransferDecision_TRANSFER_DECISION_CONFIRMED)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ledgerv1.ConfirmTransferResponse{Link: l}), nil
}

func (f *fakeTransfers) RejectTransfer(_ context.Context, r *connect.Request[ledgerv1.RejectTransferRequest]) (*connect.Response[ledgerv1.RejectTransferResponse], error) {
	m := r.Msg
	l, err := f.decide(m.GetOrigin(), m.GetSessionToken(), m.GetWithdrawal(), m.GetDeposit(), ledgerv1.TransferDecision_TRANSFER_DECISION_REJECTED)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ledgerv1.RejectTransferResponse{Link: l}), nil
}

func (f *fakeTransfers) decide(origin, session string, w, d *ledgerv1.TransactionKey, decision ledgerv1.TransferDecision) (*ledgerv1.TransferLink, error) {
	if origin == "" || session != ownerSession {
		return nil, codeErr(connect.CodePermissionDenied, nil)
	}
	if w.GetExternalId() != "w1" || d.GetExternalId() != "d1" {
		return nil, codeErr(connect.CodeFailedPrecondition, errTaken)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.decision = decision
	f.sessions = append(f.sessions, session)
	return &ledgerv1.TransferLink{Withdrawal: w, Deposit: d, Decision: decision, Version: int64(len(f.sessions))}, nil
}

// ListTransfers lists the synthetic pair under its current decision.
func (f *fakeRead) ListTransfers(_ context.Context, r *connect.Request[viewsv1.ListTransfersRequest]) (*connect.Response[viewsv1.ListTransfersResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	pair := &viewsv1.Transfer{
		WithdrawalAccountId: "acc-1", WithdrawalDate: "2026-03-02", DepositAccountId: "acc-2", DepositDate: "2026-03-04",
		Amount: &viewsv1.Money{Amount: "2500", Currency: "PLN"}, WithdrawalSource: "test", WithdrawalExternalId: "w1",
		DepositSource: "test", DepositExternalId: "d1",
	}
	out := &viewsv1.ListTransfersResponse{}
	f.transfers.mu.Lock()
	defer f.transfers.mu.Unlock()
	switch f.transfers.decision {
	case ledgerv1.TransferDecision_TRANSFER_DECISION_CONFIRMED:
		out.Confirmed = []*viewsv1.Transfer{pair}
	case ledgerv1.TransferDecision_TRANSFER_DECISION_REJECTED:
		out.Rejected = []*viewsv1.Transfer{pair}
	case ledgerv1.TransferDecision_TRANSFER_DECISION_UNSPECIFIED:
		out.Suggested = []*viewsv1.Transfer{pair}
	}
	return connect.NewResponse(out), nil
}

// transferForm is the posted pair w -> test/d1.
func transferPost(path, w string) *http.Request {
	form := url.Values{"withdrawal_source": {"test"}, "withdrawal_id": {w}, "deposit_source": {"test"}, "deposit_id": {"d1"}}
	r := request(http.MethodPost, path, form.Encode())
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

// TestTransfersPage checks the transfers page (D-105): the suggestion with
// its accounts, dates and amount; the owner confirms it, then corrects it
// as not a transfer, and the page lists it under each decision.
func TestTransfersPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/transfers", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Suggested", "acc-1", "acc-2", "2026-03-02", "2026-03-04", "2\u202f500 PLN",
		`action="/transfers/confirm"`, `name="withdrawal_id" value="w1"`, "No confirmed transfers yet.")
	wantNoBody(t, rec, "<script>", "style=")

	for _, step := range []struct {
		path, wantSection string
	}{
		{"/transfers/confirm", "No rejected pairs."},
		{"/transfers/reject", "No confirmed transfers yet."},
	} {
		rec = h.serve(t, withSession(transferPost(step.path, "w1"), ownerSession))
		wantStatus(t, rec, http.StatusSeeOther)
		if got := rec.Header().Get("Location"); got != "/transfers" {
			t.Errorf("POST %s: Location = %q, want /transfers", step.path, got)
		}
		rec = h.serve(t, withSession(request(http.MethodGet, "/transfers", ""), ownerSession))
		wantBody(t, rec, step.wantSection, "No suggested transfers.", "2\u202f500 PLN")
	}
}

// TestTransfersRefusals checks that a read session sees the transfers but
// can't decide, that the core's refusal shows on the page, and that a
// malformed form is refused before the core.
func TestTransfersRefusals(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/transfers", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Waiting for the owner")
	wantNoBody(t, rec, `action="/transfers/confirm"`)

	rec = h.serve(t, withSession(transferPost("/transfers/confirm", "w1"), readSession))
	wantStatus(t, rec, http.StatusForbidden)

	rec = h.serve(t, withSession(transferPost("/transfers/confirm", "w2"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The decision was not recorded: test/d1 is already a leg of a confirmed transfer")

	bad := request(http.MethodPost, "/transfers/reject", "withdrawal_source=test")
	bad.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = h.serve(t, withSession(bad, ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	if len(h.transfers.sessions) != 0 {
		t.Errorf("the core recorded %d decision(s), want none", len(h.transfers.sessions))
	}
}

// TestOverviewLinksSuggestedTransfers checks that the performance section
// counts the suggested transfers and links the transfers page.
func TestOverviewLinksSuggestedTransfers(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `<a href="/transfers">1 suggested transfer(s)</a>`, `<a href="/transfers">Transfers</a>`)
	if strings.Contains(rec.Body.String(), "suggested transfer(s)</a>") && h.transfers.decision != 0 {
		t.Error("the fake changed without a decision")
	}
}
