package ledger_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Synthetic ISINs with valid check digits; no real instrument is implied.
const (
	isinA = "PLTEST000010"
	isinB = "IETEST000026"
	isinC = "XSFAKE000034"
)

const testAccount ledger.AccountID = "acc-1"

func dec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

func cur(t *testing.T, code string) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", code, err)
	}
	return c
}

func money(t *testing.T, amount, code string) kernel.Money {
	t.Helper()
	return kernel.NewMoney(dec(t, amount), cur(t, code))
}

func qty(t *testing.T, s string) kernel.Quantity {
	t.Helper()
	return kernel.NewQuantity(dec(t, s))
}

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func instrument(t *testing.T, isin, symbol string) ledger.Instrument {
	t.Helper()
	i, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", isin, err)
	}
	return ledger.Instrument{ISIN: i, Symbol: symbol}
}

// must returns a function that unwraps a constructor's result and fails the
// test on its error: must(t)(ledger.NewBuy(...)).
func must(t *testing.T) func(ledger.Transaction, error) ledger.Transaction {
	t.Helper()
	return func(tx ledger.Transaction, err error) ledger.Transaction {
		t.Helper()
		if err != nil {
			t.Fatalf("building transaction: %v", err)
		}
		return tx
	}
}

// header returns a header in testAccount from source "test".
func header(t *testing.T, id, day string) ledger.TransactionHeader {
	t.Helper()
	return ledger.TransactionHeader{
		Key:     ledger.TransactionKey{Source: "test", ExternalID: id},
		Account: testAccount,
		Date:    date(t, day),
	}
}
