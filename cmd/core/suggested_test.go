package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// TestSuggestedTransfersRaiseItems checks that a withdrawal from one
// account and a deposit of the same amount into another two days later
// raise one transfer-suggested attention item, a decision tied to both
// accounts (D-105, D-130), which clears once the owner confirms the pair.
func TestSuggestedTransfersRaiseItems(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	c, srv, pool := startCore(t, dsn)
	ctx := t.Context()
	l := ledgerapp.New(ledgerpg.New(pool))
	pair := importSuggestedPair(t, l)
	c.attention.tick(ctx, quiet())
	issued := tokenCmd(t, dsn, "issue", "--name", "suggested-test", "--ttl", "1h")
	secret, _, _ := strings.Cut(issued, "\n")
	var got viewsv1.ListAttentionItemsResponse
	getJSON(t, srv, "/v1/attention-items", secret, &got)
	if len(got.GetItems()) != 1 || got.GetOpen() != 1 {
		t.Fatalf("items = %v, want one open item", got.GetItems())
	}
	if it := got.GetItems()[0]; !suggestedItemOf(it, pair) {
		t.Errorf("item = %v, want the suggested pair, a decision tied to both accounts", it)
	}
	if _, err := l.DecideTransfer(ctx, pair, ledger.TransferConfirmed); err != nil {
		t.Fatalf("DecideTransfer: %v", err)
	}
	c.attention.tick(ctx, quiet())
	getJSON(t, srv, "/v1/attention-items", secret, &got)
	if got.GetOpen() != 0 || got.GetDone() != 1 {
		t.Errorf("after confirming: open %d, done %d; want 0 and 1", got.GetOpen(), got.GetDone())
	}
}

// importSuggestedPair saves the accounts acc-out and acc-in and imports a
// withdrawal of 1500 PLN from acc-out and its deposit into acc-in two days
// later, and returns their pair.
func importSuggestedPair(t *testing.T, l *ledgerapp.Ledger) ledger.TransferPair {
	t.Helper()
	ctx := t.Context()
	for _, id := range []ledger.AccountID{"acc-out", "acc-in"} {
		if err := l.SaveAccount(ctx, ledger.Account{ID: id, Institution: "synthetic", Name: "Synthetic " + string(id)}); err != nil {
			t.Fatalf("SaveAccount: %v", err)
		}
	}
	day, dayErr := kernel.ParseDate("2026-10-20")
	pln, plnErr := kernel.ParseCurrency("PLN")
	if dayErr != nil || plnErr != nil {
		t.Fatalf("ParseDate, ParseCurrency: %v, %v", dayErr, plnErr)
	}
	amount := kernel.NewMoney(kernel.NewDecimal(1500, 0), pln)
	w, wErr := ledger.NewWithdrawal(ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: "w-1"}, Account: "acc-out", Date: day}, amount)
	d, dErr := ledger.NewDeposit(ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: "d-1"}, Account: "acc-in", Date: day.AddDays(2)}, amount)
	if wErr != nil || dErr != nil {
		t.Fatalf("building the legs: %v, %v", wErr, dErr)
	}
	if _, err := l.Import(ctx, []ledger.Transaction{w, d}, nil); err != nil {
		t.Fatalf("Import: %v", err)
	}
	return ledger.TransferPair{Withdrawal: w.Key, Deposit: d.Key}
}

// suggestedItemOf reports whether it is pair's transfer-suggested item, a
// decision tied to acc-out and acc-in.
func suggestedItemOf(it *viewsv1.AttentionItem, pair ledger.TransferPair) bool {
	return it.GetKind() == "transfer-suggested" && it.GetSubject() == pair.String() &&
		it.GetGroup() == viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DECISIONS &&
		slices.Equal(it.GetAccountIds(), []string{"acc-out", "acc-in"})
}
