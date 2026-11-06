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
	// broken makes a confirmed pair one a correction broke (D-131).
	broken bool
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

// decided reports whether the owner decided the synthetic pair.
func (f *fakeTransfers) decided() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.decision != ledgerv1.TransferDecision_TRANSFER_DECISION_UNSPECIFIED
}

// suggestedItem is the transfer-suggested attention item of the synthetic
// pair, tied to both its accounts (D-105, D-130).
func suggestedItem() *viewsv1.AttentionItem {
	return &viewsv1.AttentionItem{
		Id: 9, Kind: kindTransferSuggested, Subject: "test/w1 -> test/d1", Detail: "suggested",
		State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DECISIONS,
		AccountId: "acc-1", AccountIds: []string{"acc-1", "acc-2"},
	}
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
		if f.transfers.broken {
			pair.Amount = &viewsv1.Money{Amount: "2499.50", Currency: "PLN"}
			out.Confirmed, out.Broken = nil, []*viewsv1.Transfer{pair}
		}
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

// TestInboxTransfers checks the transfers in the Inbox (D-105, D-196): the
// suggestion's attention item with its accounts, dates and amount among the
// decisions, counted in the navigation and confirmed in place; the item
// done once the pair is decided; the confirmed pair in the done view,
// corrected there as not a transfer, back to the done view.
func TestInboxTransfers(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.attention.add(suggestedItem())
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Decisions · 1", "Confirm a transfer between your accounts", "acc-1", "acc-2", "2026-03-02", "2026-03-04",
		"2\u202f500 PLN", `action="/inbox/transfers/confirm"`, `name="withdrawal_id" value="w1"`, "Not a transfer",
		`title="3 open"`)
	wantNoBody(t, rec, "<script>", "style=")

	rec = h.serve(t, withSession(transferPost("/inbox/transfers/confirm", "w1"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/inbox" {
		t.Errorf("confirm: Location = %q, want /inbox", got)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), ownerSession))
	wantBody(t, rec, "Decisions · 0", "No transfers to confirm")
	wantNoBody(t, rec, "Confirm a transfer between your accounts")
	rec = h.serve(t, withSession(request(http.MethodGet, "/inbox?view=done", ""), ownerSession))
	wantBody(t, rec, "Transfers you decided: 1 confirmed, 0 rejected", `<span class="pill pill-ok">Confirmed</span>`, "2\u202f500 PLN")

	rec = h.serve(t, withSession(rejectFromDone(), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/inbox?view=done" {
		t.Errorf("reject from the done view: Location = %q, want /inbox?view=done", got)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/inbox?view=done", ""), ownerSession))
	wantBody(t, rec, "Transfers you decided: 0 confirmed, 1 rejected")
}

// rejectFromDone posts a rejection of test/w1 -> test/d1 from the done
// view.
func rejectFromDone() *http.Request {
	form := url.Values{"withdrawal_source": {"test"}, "withdrawal_id": {"w1"}, "deposit_source": {"test"}, "deposit_id": {"d1"}, "view": {"done"}}
	r := request(http.MethodPost, "/inbox/transfers/reject", form.Encode())
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

// TestTransfersRefusals checks that a read session sees the transfers but
// can't decide, that the core's refusal shows in the Inbox, and that a
// malformed form is refused before the core.
func TestTransfersRefusals(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.attention.add(suggestedItem())
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Waiting for the owner")
	wantNoBody(t, rec, `action="/inbox/transfers/confirm"`)

	rec = h.serve(t, withSession(transferPost("/inbox/transfers/confirm", "w1"), readSession))
	wantStatus(t, rec, http.StatusForbidden)

	rec = h.serve(t, withSession(transferPost("/inbox/transfers/confirm", "w2"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The decision was not recorded: test/d1 is already a leg of a confirmed transfer")

	bad := request(http.MethodPost, "/inbox/transfers/reject", "withdrawal_source=test")
	bad.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = h.serve(t, withSession(bad, ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	if len(h.transfers.sessions) != 0 {
		t.Errorf("the core recorded %d decision(s), want none", len(h.transfers.sessions))
	}
}

// TestOverviewLinksSuggestedTransfers checks that the performance section
// counts the suggested transfers and links the Inbox.
func TestOverviewLinksSuggestedTransfers(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `<a href="/inbox">1 suggested transfer(s)</a>`, `<a href="/inbox">Inbox`)
	if strings.Contains(rec.Body.String(), "suggested transfer(s)</a>") && h.transfers.decision != 0 {
		t.Error("the fake changed without a decision")
	}
}
