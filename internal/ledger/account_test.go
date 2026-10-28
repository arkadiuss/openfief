package ledger_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestAccountValidate(t *testing.T) {
	t.Parallel()
	valid := ledger.Account{ID: "acc-1", Institution: "broker-a", Name: "Main", Sleeve: "core"}
	tests := []struct {
		name    string
		edit    func(a *ledger.Account)
		wantErr error
	}{
		{name: "valid", edit: func(*ledger.Account) {}},
		{name: "no sleeve", edit: func(a *ledger.Account) { a.Sleeve = "" }, wantErr: ledger.ErrInvalidAccount},
		{name: "no ID", edit: func(a *ledger.Account) { a.ID = "" }, wantErr: ledger.ErrInvalidAccount},
		{name: "no institution", edit: func(a *ledger.Account) { a.Institution = "" }, wantErr: ledger.ErrInvalidAccount},
		{name: "no name", edit: func(a *ledger.Account) { a.Name = "" }, wantErr: ledger.ErrInvalidAccount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := valid
			tt.edit(&a)
			if diff := cmp.Diff(tt.wantErr, a.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBookValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		kind       ledger.BookKind
		currency   string
		instrument bool
		wantString string
		wantErr    error
	}{
		{name: "cash", kind: ledger.BookCash, currency: "USD", wantString: "cash:USD"},
		{name: "holding", kind: ledger.BookHolding, instrument: true, wantString: "holding:" + isinA + "/TSTA"},
		{name: "fee", kind: ledger.BookFee, currency: "PLN", wantString: "fee:PLN"},
		{name: "withholding tax", kind: ledger.BookWithholdingTax, currency: "USD", wantString: "withholding-tax:USD"},
		{name: "fx", kind: ledger.BookFX, currency: "EUR", wantString: "fx:EUR"},
		{name: "equity", kind: ledger.BookEquity, currency: "PLN", wantString: "equity:PLN"},
		{name: "cash without currency", kind: ledger.BookCash, wantString: "cash:", wantErr: ledger.ErrInvalidBook},
		{name: "cash with an instrument", kind: ledger.BookCash, currency: "USD", instrument: true, wantString: "cash:USD", wantErr: ledger.ErrInvalidBook},
		{name: "holding with a currency", kind: ledger.BookHolding, currency: "USD", instrument: true, wantString: "holding:" + isinA + "/TSTA", wantErr: ledger.ErrInvalidBook},
		{name: "holding without instrument", kind: ledger.BookHolding, wantString: "holding:/", wantErr: ledger.ErrInvalidBook},
		{name: "unknown kind", kind: ledger.BookKind(0), currency: "USD", wantString: "BookKind(0):USD", wantErr: ledger.ErrInvalidBook},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := ledger.Book{Kind: tt.kind}
			if tt.currency != "" {
				l.Currency = cur(t, tt.currency)
			}
			if tt.instrument {
				l.Instrument = instrument(t, isinA, "TSTA")
			}
			if diff := cmp.Diff(tt.wantErr, l.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantString, l.String()); diff != "" {
				t.Errorf("String() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBookConstructors(t *testing.T) {
	t.Parallel()
	usd := cur(t, "USD")
	in := instrument(t, isinA, "TSTA")
	got := []ledger.Book{
		ledger.CashBook(usd), ledger.HoldingBook(in), ledger.FeeBook(usd),
		ledger.WithholdingTaxBook(usd), ledger.FXBook(usd), ledger.EquityBook(usd),
	}
	want := []ledger.Book{
		{Kind: ledger.BookCash, Currency: usd},
		{Kind: ledger.BookHolding, Instrument: in},
		{Kind: ledger.BookFee, Currency: usd},
		{Kind: ledger.BookWithholdingTax, Currency: usd},
		{Kind: ledger.BookFX, Currency: usd},
		{Kind: ledger.BookEquity, Currency: usd},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("constructors (-want +got):\n%s", diff)
	}
	for _, l := range got {
		if err := l.Validate(); err != nil {
			t.Errorf("%s.Validate() = %v", l, err)
		}
	}
}
