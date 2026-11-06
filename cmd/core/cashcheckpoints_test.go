package main

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	attentiondomain "github.com/arkadiuss/openfief/internal/attention"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// TestLedgerAttention checks the adapter from the ledger's attention port
// to the attention items (D-158): the ledger's kind is the attention
// context's, a raise becomes the condition, a clear names the kind and
// subject, and the store's failure reaches the ledger.
func TestLedgerAttention(t *testing.T) {
	t.Parallel()
	if ledgerapp.CashCheckpointKind != string(attentiondomain.KindCashCheckpointMismatch) {
		t.Errorf("the ledger raises %q, the attention context knows %q", ledgerapp.CashCheckpointKind, attentiondomain.KindCashCheckpointMismatch)
	}
	const subject = "acc-1/2026-03-31/PLN"
	items := &fakeAttentionItems{}
	a := ledgerAttention{items: items}
	if err := a.Raise(t.Context(), ledgerapp.CashCheckpointKind, subject, "detail"); err != nil {
		t.Fatalf("Raise: %v", err)
	}
	if err := a.Clear(t.Context(), ledgerapp.CashCheckpointKind, subject); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	kind := attentiondomain.KindCashCheckpointMismatch
	want := []attentionCall{
		{Op: "raise", Cond: attentiondomain.Condition{Kind: kind, Subject: subject, Detail: "detail"}},
		{Op: "clear", Cond: attentiondomain.Condition{Kind: kind, Subject: subject}},
	}
	if diff := cmp.Diff(want, items.calls, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("calls (-want +got):\n%s", diff)
	}
	failing := ledgerAttention{items: &fakeAttentionItems{err: errAttentionDown}}
	if err := failing.Raise(t.Context(), ledgerapp.CashCheckpointKind, subject, "detail"); !errors.Is(err, errAttentionDown) {
		t.Errorf("Raise on a failing store = %v, want %v", err, errAttentionDown)
	}
	if err := failing.Clear(t.Context(), ledgerapp.CashCheckpointKind, subject); !errors.Is(err, errAttentionDown) {
		t.Errorf("Clear on a failing store = %v, want %v", err, errAttentionDown)
	}
}

// TestCashCheckpointService runs D-158 through the real core: the owner
// enters the cash a bank showed for an empty account, which the core
// stores as a flagged balance assertion and raises as an attention item at
// once; booking the difference deposits it, the checkpoint then matches
// and the item resolves, and booking again finds nothing to book. Another
// caller and an unknown session are refused.
func TestCashCheckpointService(t *testing.T) {
	t.Parallel()
	oc := ownerSession(t)
	ctx := t.Context()
	accounts := ingestv1connect.NewAccountServiceClient(oc.srv.Client(), oc.srv.URL, connect.WithInterceptors(viewBearer(uiCaller)))
	created, err := accounts.CreateAccount(ctx, connect.NewRequest(&ingestv1.CreateAccountRequest{
		Origin: localUI, SessionToken: oc.session, Name: "Bank", Sleeve: "satellite", Institution: "bank",
	}))
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	id := created.Msg.GetAccount().GetId()
	c := ledgerv1connect.NewCashCheckpointServiceClient(oc.srv.Client(), oc.srv.URL, connect.WithInterceptors(viewBearer(uiCaller)))
	rec, err := c.RecordCashCheckpoint(ctx, connect.NewRequest(&ledgerv1.RecordCashCheckpointRequest{
		Origin: localUI, SessionToken: oc.session, AccountId: id, Date: "2026-03-31", Currency: "PLN", Amount: "250.00",
	}))
	if err != nil || rec.Msg.GetCheckpoint().GetMatched() || rec.Msg.GetCheckpoint().GetDerived() != "0" {
		t.Fatalf("RecordCashCheckpoint = %v, %v; want a mismatch with 0 derived", rec, err)
	}
	checkCheckpointItem(t, oc, id, viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN)
	book := &ledgerv1.BookCashDifferenceRequest{Origin: localUI, SessionToken: oc.session, AccountId: id, Date: "2026-03-31", Currency: "PLN"}
	booked, err := c.BookCashDifference(ctx, connect.NewRequest(book))
	if err != nil || booked.Msg.GetKind() != "deposit" || booked.Msg.GetAmount() != "250" {
		t.Fatalf("BookCashDifference = %v, %v; want a deposit of 250", booked, err)
	}
	checkCheckpointItem(t, oc, id, viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_RESOLVED)
	checkCheckpointRefusals(t, oc, book)
}

// checkCheckpointRefusals checks that the core refuses booking book's
// difference again, from another caller and with an unknown session.
func checkCheckpointRefusals(t *testing.T, oc ownerCore, book *ledgerv1.BookCashDifferenceRequest) {
	t.Helper()
	ctx := t.Context()
	for _, tt := range []struct {
		name, caller, session string
		want                  connect.Code
	}{
		{"booking again", uiCaller, oc.session, connect.CodeFailedPrecondition},
		{"another caller", "wrong", oc.session, connect.CodePermissionDenied},
		{"an unknown session", uiCaller, "ofsess_nope_nope", connect.CodeUnauthenticated},
	} {
		req := &ledgerv1.BookCashDifferenceRequest{
			Origin: book.GetOrigin(), SessionToken: tt.session, AccountId: book.GetAccountId(), Date: book.GetDate(), Currency: book.GetCurrency(),
		}
		c := ledgerv1connect.NewCashCheckpointServiceClient(oc.srv.Client(), oc.srv.URL, connect.WithInterceptors(viewBearer(tt.caller)))
		if _, err := c.BookCashDifference(ctx, connect.NewRequest(req)); connect.CodeOf(err) != tt.want {
			t.Errorf("%s = %v, want %s", tt.name, err, tt.want)
		}
	}
}

// checkCheckpointItem fails t unless the read API lists the cash
// checkpoint item of account id's PLN checkpoint of 2026-03-31 in state.
func checkCheckpointItem(t *testing.T, oc ownerCore, id string, state viewsv1.AttentionItemState) {
	t.Helper()
	var got viewsv1.ListAttentionItemsResponse
	getJSON(t, oc.srv, "/v1/attention-items", oc.session, &got)
	for _, it := range got.GetItems() {
		if it.GetKind() == "cash-checkpoint-mismatch" && it.GetSubject() == id+"/2026-03-31/PLN" {
			if it.GetState() != state {
				t.Errorf("the checkpoint's item is %s, want %s", it.GetState(), state)
			}
			return
		}
	}
	t.Errorf("no cash checkpoint item in %v", got.GetItems())
}
