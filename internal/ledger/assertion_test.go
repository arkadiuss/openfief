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

func TestCheckBalanceAssertion(t *testing.T) {
	t.Parallel()
	log := holdingLog(t)
	a, b := instrument(t, isinA, "TSTA"), instrument(t, isinB, "TSTB")
	cash := func(on, amount, code string) ledger.BalanceAssertion {
		return ledger.CashAssertion(testAccount, date(t, on), money(t, amount, code))
	}
	units := func(on string, in ledger.Instrument, n string) ledger.BalanceAssertion {
		return ledger.HoldingAssertion(testAccount, date(t, on), in, qty(t, n))
	}
	feeAssertion := cash("2026-03-06", "0", "USD")
	feeAssertion.Book = ledger.FeeBook(cur(t, "USD"))
	noAccount := cash("2026-03-06", "0", "USD")
	noAccount.Account = ""
	noDate := cash("2026-03-06", "0", "USD")
	noDate.Date = kernel.Date{}
	noCurrency := ledger.BalanceAssertion{Account: testAccount, Date: date(t, "2026-03-06"), Book: ledger.Book{Kind: ledger.BookCash}}
	tests := []struct {
		name        string
		assertion   ledger.BalanceAssertion
		wantErr     error
		wantDerived string // set when the error is an *AssertionMismatchError
	}{
		{name: "cash matches", assertion: cash("2026-03-06", "1287", "USD")},
		{name: "cash matches in another scale", assertion: cash("2026-03-06", "1287.00", "USD")},
		{name: "cash in another currency", assertion: cash("2026-03-07", "5", "EUR")},
		{name: "cash of a currency never held", assertion: cash("2026-03-07", "0", "PLN")},
		{name: "cash mismatch names both balances", assertion: cash("2026-03-05", "1287", "USD"), wantDerived: "1280"},
		{name: "units match", assertion: units("2026-03-05", a, "11")},
		{name: "units of a sold-out holding", assertion: units("2026-03-07", b, "0")},
		{name: "broker reports the uncorrected row", assertion: units("2026-03-02", a, "13"), wantDerived: "15"},
		{name: "fee ledger account", assertion: feeAssertion, wantErr: ledger.ErrInvalidAssertion},
		{name: "no account", assertion: noAccount, wantErr: ledger.ErrInvalidAssertion},
		{name: "no date", assertion: noDate, wantErr: ledger.ErrInvalidAssertion},
		{name: "cash without currency", assertion: noCurrency, wantErr: ledger.ErrInvalidAssertion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ledger.CheckBalanceAssertion(log, &tt.assertion)
			if tt.wantDerived == "" {
				if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
					t.Errorf("CheckBalanceAssertion error (-want +got):\n%s", diff)
				}
				return
			}
			checkMismatch(t, err, &tt.assertion, dec(t, tt.wantDerived))
		})
	}
}

// checkMismatch expects an *AssertionMismatchError carrying the assertion
// and the derived balance.
func checkMismatch(t *testing.T, err error, assertion *ledger.BalanceAssertion, derived kernel.Decimal) {
	t.Helper()
	var mismatch *ledger.AssertionMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("error = %v, want an *AssertionMismatchError", err)
	}
	want := &ledger.AssertionMismatchError{Assertion: *assertion, Derived: derived}
	if diff := cmp.Diff(want, mismatch); diff != "" {
		t.Errorf("AssertionMismatchError (-want +got):\n%s", diff)
	}
	msg := err.Error()
	if !strings.Contains(msg, assertion.Balance.String()) || !strings.Contains(msg, derived.String()) {
		t.Errorf("Error() = %q, want both balances", msg)
	}
}

func TestCashOn(t *testing.T) {
	t.Parallel()
	log := holdingLog(t)
	tests := []struct {
		name string
		on   string
		code string
		want string
	}{
		{name: "before the deposit", on: "2026-02-28", code: "USD", want: "0"},
		{name: "after the deposit", on: "2026-03-01", code: "USD", want: "2000"},
		{name: "after the buy at its corrected cost", on: "2026-03-02", code: "USD", want: "800"},
		{name: "after the sale and the dividend", on: "2026-03-06", code: "USD", want: "1287"},
		{name: "another currency", on: "2026-03-07", code: "EUR", want: "5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.CashOn(log, testAccount, cur(t, tt.code), date(t, tt.on))
			if err != nil {
				t.Fatalf("CashOn: %v", err)
			}
			if diff := cmp.Diff(money(t, tt.want, tt.code), got); diff != "" {
				t.Errorf("CashOn (-want +got):\n%s", diff)
			}
		})
	}
	invalid := append([]ledger.VersionedTransaction(nil), log...)
	invalid[0].Version = 0
	if _, err := ledger.CashOn(invalid, testAccount, cur(t, "USD"), date(t, "2026-03-06")); !errors.Is(err, ledger.ErrInvalidLog) {
		t.Errorf("CashOn on an invalid log = %v, want ErrInvalidLog", err)
	}
	assertion := ledger.CashAssertion(testAccount, date(t, "2026-03-06"), money(t, "1287", "USD"))
	if err := ledger.CheckBalanceAssertion(invalid, &assertion); !errors.Is(err, ledger.ErrInvalidLog) {
		t.Errorf("CheckBalanceAssertion on an invalid log = %v, want ErrInvalidLog", err)
	}
}
