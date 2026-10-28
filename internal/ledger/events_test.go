package ledger_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestTransactionConstructors(t *testing.T) {
	t.Parallel()
	h := header(t, "e-1", "2026-03-02")
	in := instrument(t, isinA, "TSTA")
	usd, pln := cur(t, "USD"), cur(t, "PLN")
	m := func(amount, code string) kernel.Money { return money(t, amount, code) }
	ten, zeroQty := qty(t, "10"), qty(t, "0")
	opened, afterGoLive := date(t, "2019-05-17"), date(t, "2026-03-03")
	minusOne, minusTen := qty(t, "-1"), qty(t, "-10")
	none := kernel.Money{}
	type posting = ledger.Posting
	tests := []struct {
		name     string
		build    func() (ledger.Transaction, error)
		wantKind ledger.TransactionKind
		want     []posting
		wantErr  error
	}{
		{
			name: "buy with a fee in the account's base currency",
			build: func() (ledger.Transaction, error) {
				return ledger.NewBuy(h, in, ten, m("1500.25", "USD"), m("4", "PLN"))
			},
			wantKind: ledger.KindBuy,
			want: []posting{
				{Book: ledger.HoldingBook(in), Amount: m("1500.25", "USD"), Quantity: ten},
				{Book: ledger.CashBook(usd), Amount: m("-1500.25", "USD")},
				{Book: ledger.FeeBook(pln), Amount: m("4", "PLN")},
				{Book: ledger.CashBook(pln), Amount: m("-4", "PLN")},
			},
		},
		{
			name:     "buy without a fee",
			build:    func() (ledger.Transaction, error) { return ledger.NewBuy(h, in, ten, m("1500", "USD"), none) },
			wantKind: ledger.KindBuy,
			want: []posting{
				{Book: ledger.HoldingBook(in), Amount: m("1500", "USD"), Quantity: ten},
				{Book: ledger.CashBook(usd), Amount: m("-1500", "USD")},
			},
		},
		{
			name: "sell with a fee",
			build: func() (ledger.Transaction, error) {
				return ledger.NewSell(h, in, ten, m("1700", "USD"), m("1.5", "USD"))
			},
			wantKind: ledger.KindSell,
			want: []posting{
				{Book: ledger.HoldingBook(in), Amount: m("-1700", "USD"), Quantity: minusTen},
				{Book: ledger.CashBook(usd), Amount: m("1700", "USD")},
				{Book: ledger.FeeBook(usd), Amount: m("1.5", "USD")},
				{Book: ledger.CashBook(usd), Amount: m("-1.5", "USD")},
			},
		},
		{
			name:     "dividend with withholding tax",
			build:    func() (ledger.Transaction, error) { return ledger.NewDividend(h, in, m("100", "USD"), m("15", "USD")) },
			wantKind: ledger.KindDividend,
			want: []posting{
				{Book: ledger.HoldingBook(in), Amount: m("-100", "USD"), Quantity: zeroQty},
				{Book: ledger.WithholdingTaxBook(usd), Amount: m("15", "USD")},
				{Book: ledger.CashBook(usd), Amount: m("85", "USD")},
			},
		},
		{
			name:     "dividend without withholding tax",
			build:    func() (ledger.Transaction, error) { return ledger.NewDividend(h, in, m("100", "USD"), none) },
			wantKind: ledger.KindDividend,
			want: []posting{
				{Book: ledger.HoldingBook(in), Amount: m("-100", "USD")},
				{Book: ledger.CashBook(usd), Amount: m("100", "USD")},
			},
		},
		{
			name:     "dividend fully withheld",
			build:    func() (ledger.Transaction, error) { return ledger.NewDividend(h, in, m("10", "USD"), m("10", "USD")) },
			wantKind: ledger.KindDividend,
			want: []posting{
				{Book: ledger.HoldingBook(in), Amount: m("-10", "USD")},
				{Book: ledger.WithholdingTaxBook(usd), Amount: m("10", "USD")},
			},
		},
		{
			name:     "account fee",
			build:    func() (ledger.Transaction, error) { return ledger.NewFee(h, m("10", "PLN")) },
			wantKind: ledger.KindFee,
			want: []posting{
				{Book: ledger.FeeBook(pln), Amount: m("10", "PLN")},
				{Book: ledger.CashBook(pln), Amount: m("-10", "PLN")},
			},
		},
		{
			name: "FX conversion from USD to PLN with a fee",
			build: func() (ledger.Transaction, error) {
				return ledger.NewFXConversion(h, m("1000", "USD"), m("3712.40", "PLN"), m("2", "USD"))
			},
			wantKind: ledger.KindFXConversion,
			want: []posting{
				{Book: ledger.CashBook(usd), Amount: m("-1000", "USD")},
				{Book: ledger.FXBook(usd), Amount: m("1000", "USD")},
				{Book: ledger.FXBook(pln), Amount: m("-3712.40", "PLN")},
				{Book: ledger.CashBook(pln), Amount: m("3712.40", "PLN")},
				{Book: ledger.FeeBook(usd), Amount: m("2", "USD")},
				{Book: ledger.CashBook(usd), Amount: m("-2", "USD")},
			},
		},
		{
			name:     "deposit",
			build:    func() (ledger.Transaction, error) { return ledger.NewDeposit(h, m("5000", "PLN")) },
			wantKind: ledger.KindDeposit,
			want: []posting{
				{Book: ledger.CashBook(pln), Amount: m("5000", "PLN")},
				{Book: ledger.EquityBook(pln), Amount: m("-5000", "PLN")},
			},
		},
		{
			name:     "withdrawal",
			build:    func() (ledger.Transaction, error) { return ledger.NewWithdrawal(h, m("200", "PLN")) },
			wantKind: ledger.KindWithdrawal,
			want: []posting{
				{Book: ledger.CashBook(pln), Amount: m("-200", "PLN")},
				{Book: ledger.EquityBook(pln), Amount: m("200", "PLN")},
			},
		},
		{
			name: "opening lot",
			build: func() (ledger.Transaction, error) {
				return ledger.NewOpeningLot(h, in, ten, m("812.34", "USD"), opened)
			},
			wantKind: ledger.KindOpeningLot,
			want: []posting{
				{Book: ledger.HoldingBook(in), Amount: m("812.34", "USD"), Quantity: ten},
				{Book: ledger.EquityBook(usd), Amount: m("-812.34", "USD")},
			},
		},
		{name: "buy of zero units", build: func() (ledger.Transaction, error) { return ledger.NewBuy(h, in, zeroQty, m("1", "USD"), none) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "buy at negative cost", build: func() (ledger.Transaction, error) { return ledger.NewBuy(h, in, ten, m("-1", "USD"), none) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "buy with a negative fee", build: func() (ledger.Transaction, error) { return ledger.NewBuy(h, in, ten, m("1", "USD"), m("-1", "USD")) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "buy without cost currency", build: func() (ledger.Transaction, error) { return ledger.NewBuy(h, in, ten, none, none) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "sell of negative units", build: func() (ledger.Transaction, error) { return ledger.NewSell(h, in, minusOne, m("1", "USD"), none) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "sell with negative proceeds", build: func() (ledger.Transaction, error) { return ledger.NewSell(h, in, ten, m("-1", "USD"), none) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "dividend with more tax than gross", build: func() (ledger.Transaction, error) { return ledger.NewDividend(h, in, m("10", "USD"), m("11", "USD")) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "dividend with tax in another currency", build: func() (ledger.Transaction, error) { return ledger.NewDividend(h, in, m("10", "USD"), m("1", "PLN")) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "dividend with negative tax", build: func() (ledger.Transaction, error) { return ledger.NewDividend(h, in, m("10", "USD"), m("-1", "USD")) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "zero dividend", build: func() (ledger.Transaction, error) { return ledger.NewDividend(h, in, m("0", "USD"), none) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "zero fee", build: func() (ledger.Transaction, error) { return ledger.NewFee(h, m("0", "PLN")) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "FX conversion to the same currency", build: func() (ledger.Transaction, error) {
			return ledger.NewFXConversion(h, m("1", "USD"), m("1", "USD"), none)
		}, wantErr: ledger.ErrInvalidTransaction},
		{name: "FX conversion of nothing", build: func() (ledger.Transaction, error) {
			return ledger.NewFXConversion(h, m("0", "USD"), m("1", "PLN"), none)
		}, wantErr: ledger.ErrInvalidTransaction},
		{name: "FX conversion to nothing", build: func() (ledger.Transaction, error) {
			return ledger.NewFXConversion(h, m("1", "USD"), m("0", "PLN"), none)
		}, wantErr: ledger.ErrInvalidTransaction},
		{name: "FX conversion with a negative fee", build: func() (ledger.Transaction, error) {
			return ledger.NewFXConversion(h, m("1", "USD"), m("4", "PLN"), m("-1", "USD"))
		}, wantErr: ledger.ErrInvalidTransaction},
		{name: "negative deposit", build: func() (ledger.Transaction, error) { return ledger.NewDeposit(h, m("-1", "PLN")) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "zero withdrawal", build: func() (ledger.Transaction, error) { return ledger.NewWithdrawal(h, m("0", "PLN")) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "opening lot of no units", build: func() (ledger.Transaction, error) { return ledger.NewOpeningLot(h, in, zeroQty, m("1", "USD"), opened) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "opening lot at negative cost", build: func() (ledger.Transaction, error) { return ledger.NewOpeningLot(h, in, ten, m("-1", "USD"), opened) }, wantErr: ledger.ErrInvalidTransaction},
		{name: "opening lot opened after go-live", build: func() (ledger.Transaction, error) {
			return ledger.NewOpeningLot(h, in, ten, m("1", "USD"), afterGoLive)
		}, wantErr: ledger.ErrInvalidTransaction},
		{name: "invalid header", build: func() (ledger.Transaction, error) {
			return ledger.NewDeposit(ledger.TransactionHeader{}, m("1", "PLN"))
		}, wantErr: ledger.ErrInvalidTransaction},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.build()
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			checkBuilt(t, &got, tt.wantKind, tt.want)
			if got.Kind == ledger.KindOpeningLot && !got.LotOpened.Equal(opened) {
				t.Errorf("LotOpened = %s, want %s", got.LotOpened, opened)
			}
		})
	}
}

// checkBuilt compares a constructor's transaction with the kind and postings
// it should have, and checks that it took the header of every test case.
func checkBuilt(t *testing.T, got *ledger.Transaction, wantKind ledger.TransactionKind, want []ledger.Posting) {
	t.Helper()
	if diff := cmp.Diff(want, got.Postings); diff != "" {
		t.Errorf("postings (-want +got):\n%s", diff)
	}
	gotHeader := []any{got.Kind, got.Key.ExternalID, got.Account, got.Date.String()}
	if diff := cmp.Diff([]any{wantKind, "e-1", testAccount, "2026-03-02"}, gotHeader); diff != "" {
		t.Errorf("kind, key, account, date (-want +got):\n%s", diff)
	}
}

func TestKindNames(t *testing.T) {
	t.Parallel()
	got := []string{}
	for k := ledger.TransactionKind(0); k <= ledger.KindOpeningLot+1; k++ {
		got = append(got, k.String())
	}
	want := []string{
		"TransactionKind(0)", "buy", "sell", "dividend", "fee", "fx-conversion",
		"deposit", "withdrawal", "opening-lot", "TransactionKind(9)",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TransactionKind.String (-want +got):\n%s", diff)
	}
	key := ledger.TransactionKey{Source: "broker-a", ExternalID: "42"}
	if diff := cmp.Diff("broker-a/42", key.String()); diff != "" {
		t.Errorf("TransactionKey.String (-want +got):\n%s", diff)
	}
}
