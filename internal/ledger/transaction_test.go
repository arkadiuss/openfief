package ledger_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestTransactionValidate(t *testing.T) {
	t.Parallel()
	in := instrument(t, isinA, "TSTA")
	usd, pln := cur(t, "USD"), cur(t, "PLN")
	base := ledger.Transaction{
		Key:     ledger.TransactionKey{Source: "test", ExternalID: "t-1"},
		Account: testAccount,
		Kind:    ledger.KindBuy,
		Date:    date(t, "2026-03-02"),
		Postings: []ledger.Posting{
			{Book: ledger.HoldingBook(in), Amount: money(t, "1000.50", "USD"), Quantity: qty(t, "10")},
			{Book: ledger.CashBook(usd), Amount: money(t, "-1000.50", "USD")},
			{Book: ledger.FeeBook(pln), Amount: money(t, "4", "PLN")},
			{Book: ledger.CashBook(pln), Amount: money(t, "-4", "PLN")},
		},
	}
	cashOneCentShort := money(t, "-1000.49", "USD")
	plnFeeTooHigh := money(t, "4.01", "PLN")
	usdInPLNCash := money(t, "-1000.50", "USD")
	someQty := qty(t, "1")
	zeroUSD := money(t, "0", "USD")
	usdCentShort, plnCentOver := money(t, "0.01", "USD"), money(t, "0.01", "PLN")
	goLive, earlier, later := date(t, "2026-03-02"), date(t, "2019-05-17"), date(t, "2026-03-03")
	tests := []struct {
		name    string
		edit    func(tx *ledger.Transaction)
		wantErr error
		wantSum *kernel.Money // set when the error is an *UnbalancedError
	}{
		{name: "balanced in two currencies", edit: func(*ledger.Transaction) {}},
		{name: "one cent short in USD", edit: func(tx *ledger.Transaction) { tx.Postings[1].Amount = cashOneCentShort }, wantSum: &usdCentShort},
		{name: "PLN unbalanced while USD balances", edit: func(tx *ledger.Transaction) { tx.Postings[2].Amount = plnFeeTooHigh }, wantSum: &plnCentOver},
		{name: "a single posting", edit: func(tx *ledger.Transaction) { tx.Postings = tx.Postings[:1] }, wantErr: ledger.ErrInvalidTransaction},
		{name: "no source", edit: func(tx *ledger.Transaction) { tx.Key.Source = "" }, wantErr: ledger.ErrInvalidTransaction},
		{name: "no external ID", edit: func(tx *ledger.Transaction) { tx.Key.ExternalID = "" }, wantErr: ledger.ErrInvalidTransaction},
		{name: "no account", edit: func(tx *ledger.Transaction) { tx.Account = "" }, wantErr: ledger.ErrInvalidTransaction},
		{name: "no kind", edit: func(tx *ledger.Transaction) { tx.Kind = 0 }, wantErr: ledger.ErrInvalidTransaction},
		{name: "no date", edit: func(tx *ledger.Transaction) { tx.Date = kernel.Date{} }, wantErr: ledger.ErrInvalidTransaction},
		{name: "buy with a lot open date", edit: func(tx *ledger.Transaction) { tx.LotOpened = earlier }, wantErr: ledger.ErrInvalidTransaction},
		{name: "opening lot without a lot open date", edit: func(tx *ledger.Transaction) { tx.Kind = ledger.KindOpeningLot }, wantErr: ledger.ErrInvalidTransaction},
		{name: "opening lot opened before go-live", edit: func(tx *ledger.Transaction) { tx.Kind, tx.LotOpened = ledger.KindOpeningLot, earlier }},
		{name: "opening lot opened on go-live", edit: func(tx *ledger.Transaction) { tx.Kind, tx.LotOpened = ledger.KindOpeningLot, goLive }},
		{name: "opening lot opened after go-live", edit: func(tx *ledger.Transaction) { tx.Kind, tx.LotOpened = ledger.KindOpeningLot, later }, wantErr: ledger.ErrInvalidTransaction},
		{name: "USD posted to PLN cash", edit: func(tx *ledger.Transaction) { tx.Postings[3].Amount = usdInPLNCash }, wantErr: ledger.ErrInvalidTransaction},
		{name: "quantity on cash", edit: func(tx *ledger.Transaction) { tx.Postings[1].Quantity = someQty }, wantErr: ledger.ErrInvalidTransaction},
		{name: "amount without currency", edit: func(tx *ledger.Transaction) { tx.Postings[0].Amount = kernel.Money{} }, wantErr: ledger.ErrInvalidTransaction},
		{name: "empty posting", edit: func(tx *ledger.Transaction) {
			tx.Postings = append(tx.Postings, ledger.Posting{Book: ledger.CashBook(usd), Amount: zeroUSD})
		}, wantErr: ledger.ErrInvalidTransaction},
		{name: "invalid ledger account", edit: func(tx *ledger.Transaction) { tx.Postings[1].Book = ledger.Book{} }, wantErr: ledger.ErrInvalidTransaction},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tx := base
			tx.Postings = append([]ledger.Posting(nil), base.Postings...)
			tt.edit(&tx)
			checkValidateError(t, tx.Validate(), tt.wantErr, tt.wantSum)
		})
	}
}

// checkValidateError compares err with wantErr, or, when wantSum is set,
// expects an *UnbalancedError with that sum.
func checkValidateError(t *testing.T, err, wantErr error, wantSum *kernel.Money) {
	t.Helper()
	if wantSum == nil {
		if diff := cmp.Diff(wantErr, err, cmpopts.EquateErrors()); diff != "" {
			t.Errorf("Validate() error (-want +got):\n%s", diff)
		}
		return
	}
	var unbalanced *ledger.UnbalancedError
	if !errors.As(err, &unbalanced) {
		t.Fatalf("Validate() = %v, want an *UnbalancedError", err)
	}
	if diff := cmp.Diff(*wantSum, unbalanced.Sum); diff != "" {
		t.Errorf("UnbalancedError.Sum (-want +got):\n%s", diff)
	}
	if !strings.Contains(err.Error(), wantSum.Currency().String()) {
		t.Errorf("Error() = %q, want it to name %s", err, wantSum.Currency())
	}
}
