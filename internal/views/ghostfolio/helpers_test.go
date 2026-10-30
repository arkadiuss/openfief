package ghostfolio_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
)

// Synthetic ISINs with valid check digits; no real instrument is implied.
// isinPL with a series code as its symbol is a treasury bond.
const (
	isinPL = "PLTEST000010"
	isinIE = "IETEST000026"
	isinXS = "XSTEST000013"
)

// The view instance of the tests and the Ghostfolio data source of the
// listed instrument.
const (
	instance = "ghostfolio"
	yahoo    = "YAHOO"
)

// gfUser is the synthetic Ghostfolio user of the offline tests.
const gfUser = "5a1e0f6c-0000-4000-8000-00000000c0de"

// bondSym and feeSym are the MANUAL assets' symbols of the tests' instance
// and user.
func bondSym(series string) string { return ghostfolio.BondSymbol(instance, gfUser, series) }

func feeSym(t *testing.T, code string) string {
	t.Helper()
	return ghostfolio.FeeSymbol(instance, gfUser, cur(t, code))
}

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
	m, err := kernel.ParseMoney(amount, cur(t, code))
	if err != nil {
		t.Fatalf("ParseMoney(%q): %v", amount, err)
	}
	return m
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

func qty(t *testing.T, s string) kernel.Quantity {
	t.Helper()
	q, err := kernel.ParseQuantity(s)
	if err != nil {
		t.Fatalf("ParseQuantity(%q): %v", s, err)
	}
	return q
}

// header is the header of transaction ext of account on day.
func header(t *testing.T, account, ext, day string) ledger.TransactionHeader {
	t.Helper()
	return ledger.TransactionHeader{
		Key: ledger.TransactionKey{Source: "test/" + account, ExternalID: ext}, Account: ledger.AccountID(account), Date: date(t, day),
	}
}

// mustTx returns a function that returns the transaction a ledger
// constructor built, failing t when it returned an error.
func mustTx(t *testing.T) func(ledger.Transaction, error) ledger.Transaction {
	t.Helper()
	return func(tx ledger.Transaction, err error) ledger.Transaction {
		t.Helper()
		if err != nil {
			t.Fatalf("building a transaction: %v", err)
		}
		return tx
	}
}

// comment is the comment of the activity of transaction ext of account.
func comment(account, ext string) string {
	return ghostfolio.CommentPrefix + "test/" + account + "/" + ext
}

// activity is a desired activity in the owned account of account.
func activity(t *testing.T, account, ext string, typ ghostfolio.ActivityType, day, currency, source, symbol, quantity, unit, fee string) ghostfolio.Activity {
	t.Helper()
	return ghostfolio.Activity{
		AccountID: ghostfolio.AccountUUID(instance, ledger.AccountID(account)), Comment: comment(account, ext), Type: typ,
		Date: date(t, day), Currency: cur(t, currency), DataSource: source, Symbol: symbol,
		Quantity: dec(t, quantity), UnitPrice: dec(t, unit), Fee: dec(t, fee),
	}
}
