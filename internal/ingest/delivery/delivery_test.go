package delivery_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

func dec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

func cur(t *testing.T, s string) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency(s)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", s, err)
	}
	return c
}

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

// synthetic is a synthetic instrument; its ISIN's check digit is valid.
func synthetic(t *testing.T) ledger.Instrument {
	t.Helper()
	isin, err := ledger.ParseISIN("IETEST000026")
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	return ledger.Instrument{ISIN: isin, Symbol: "TSTB"}
}

func TestFromMapping(t *testing.T) {
	t.Parallel()
	eur := cur(t, "EUR")
	in := synthetic(t)
	m := ingest.Mapping{
		Transactions: []ledger.Transaction{
			{
				Key:     ledger.TransactionKey{Source: "ibkr-flex:acc-ibkr", ExternalID: "opening-lot:1"},
				Account: "acc-ibkr", Kind: ledger.KindOpeningLot, Date: date(t, "2026-03-13"), LotOpened: date(t, "2025-11-03"),
				Description: "opening lot",
				Postings: []ledger.Posting{
					{Book: ledger.HoldingBook(in), Amount: kernel.NewMoney(dec(t, "160.50"), eur), Quantity: kernel.NewQuantity(dec(t, "1.5"))},
					{Book: ledger.Book{Kind: ledger.BookEquity, Currency: eur}, Amount: kernel.NewMoney(dec(t, "-160.50"), eur)},
				},
			},
			{
				Key:     ledger.TransactionKey{Source: "ibkr-flex:acc-ibkr", ExternalID: "fee:2"},
				Account: "acc-ibkr", Kind: ledger.KindFee, Date: date(t, "2026-03-14"),
				Postings: []ledger.Posting{
					{Book: ledger.Book{Kind: ledger.BookFee, Currency: eur}, Amount: kernel.NewMoney(dec(t, "0.0001"), eur)},
					{Book: ledger.CashBook(eur), Amount: kernel.NewMoney(dec(t, "-0.0001"), eur)},
				},
			},
		},
		Assertions: []ledger.BalanceAssertion{
			{Account: "acc-ibkr", Date: date(t, "2026-03-13"), Book: ledger.HoldingBook(in), Balance: dec(t, "4.5")},
			{Account: "acc-ibkr", Date: date(t, "2026-03-13"), Book: ledger.CashBook(eur), Balance: dec(t, "-1.20")},
		},
		Problems: []ingest.Problem{ingest.Unsupported("row#4", "option trade"), ingest.Skipped("row#5", "summary row")},
	}
	holding := &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_HOLDING, Instrument: &connectorv1.Instrument{Isin: "IETEST000026", Symbol: "TSTB"}}
	want := &connectorv1.Delivery{
		BatchId: "b-1", Institution: "ibkr",
		Transactions: []*connectorv1.Transaction{
			{
				Key:       &connectorv1.TransactionKey{Source: "ibkr-flex:acc-ibkr", ExternalId: "opening-lot:1"},
				AccountId: "acc-ibkr", Kind: connectorv1.TransactionKind_TRANSACTION_KIND_OPENING_LOT,
				Date: "2026-03-13", LotOpened: "2025-11-03", Description: "opening lot",
				Postings: []*connectorv1.Posting{
					{Book: holding, Amount: &connectorv1.Money{Amount: "160.5", Currency: "EUR"}, Quantity: "1.5"},
					{Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_EQUITY, Currency: "EUR"}, Amount: &connectorv1.Money{Amount: "-160.5", Currency: "EUR"}},
				},
			},
			{
				Key:       &connectorv1.TransactionKey{Source: "ibkr-flex:acc-ibkr", ExternalId: "fee:2"},
				AccountId: "acc-ibkr", Kind: connectorv1.TransactionKind_TRANSACTION_KIND_FEE, Date: "2026-03-14",
				Postings: []*connectorv1.Posting{
					{Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_FEE, Currency: "EUR"}, Amount: &connectorv1.Money{Amount: "0.0001", Currency: "EUR"}},
					{Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_CASH, Currency: "EUR"}, Amount: &connectorv1.Money{Amount: "-0.0001", Currency: "EUR"}},
				},
			},
		},
		BalanceAssertions: []*connectorv1.BalanceAssertion{
			{AccountId: "acc-ibkr", Date: "2026-03-13", Book: holding, Balance: "4.5"},
			{AccountId: "acc-ibkr", Date: "2026-03-13", Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_CASH, Currency: "EUR"}, Balance: "-1.2"},
		},
		Problems: []*connectorv1.Problem{
			{Kind: connectorv1.ProblemKind_PROBLEM_KIND_UNSUPPORTED, Ref: "row#4", Reason: "option trade"},
			{Kind: connectorv1.ProblemKind_PROBLEM_KIND_SKIPPED, Ref: "row#5", Reason: "summary row"},
		},
	}
	got, err := delivery.FromMapping("b-1", "ibkr", &m)
	if err != nil {
		t.Fatalf("FromMapping: %v", err)
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("FromMapping (-want +got):\n%s", diff)
	}
}

func TestFromMappingRefuses(t *testing.T) {
	t.Parallel()
	eur := cur(t, "EUR")
	tests := []struct {
		name string
		m    ingest.Mapping
	}{
		{name: "zero transaction kind", m: ingest.Mapping{Transactions: []ledger.Transaction{{Account: "a"}}}},
		{name: "zero ledger account kind", m: ingest.Mapping{Transactions: []ledger.Transaction{{
			Kind: ledger.KindFee, Postings: []ledger.Posting{{Amount: kernel.NewMoney(dec(t, "1"), eur)}},
		}}}},
		{name: "zero ledger account kind in an assertion", m: ingest.Mapping{Assertions: []ledger.BalanceAssertion{{Account: "a"}}}},
		{name: "zero problem kind", m: ingest.Mapping{Problems: []ingest.Problem{{Ref: "r"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := delivery.FromMapping("b", "i", &tt.m)
			if diff := cmp.Diff(delivery.ErrUnknownValue, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("FromMapping error (-want +got):\n%s", diff)
			}
		})
	}
}

// TestFromMappingIBKRStatement maps the ibkrflex adapter's synthetic
// statement in both modes and checks that every transaction, assertion and
// problem is written, in order, with plain decimal strings.
func TestFromMappingIBKRStatement(t *testing.T) {
	t.Parallel()
	for _, mode := range []ingest.Mode{ingest.ModeOpening, ingest.ModeTransactions} {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			src, err := os.ReadFile(filepath.Join("..", "ibkrflex", "testdata", "activity.xml"))
			if err != nil {
				t.Fatalf("read statement: %v", err)
			}
			m, err := ibkrflex.Mapper{Accounts: map[string]ledger.AccountID{"U0000000": "acc-ibkr"}}.Map(bytes.NewReader(src), mode)
			if err != nil {
				t.Fatalf("Map: %v", err)
			}
			d, err := delivery.FromMapping("b-2", "ibkr", &m)
			if err != nil {
				t.Fatalf("FromMapping: %v", err)
			}
			checkWritten(t, &m, d)
		})
	}
}

// checkWritten fails t unless d holds every transaction, assertion and
// problem of m, in order, with each posting's amount as m's decimal string.
func checkWritten(t *testing.T, m *ingest.Mapping, d *connectorv1.Delivery) {
	t.Helper()
	got := []int{len(d.GetTransactions()), len(d.GetBalanceAssertions()), len(d.GetProblems())}
	if want := []int{len(m.Transactions), len(m.Assertions), len(m.Problems)}; !cmp.Equal(want, got) || got[0] == 0 {
		t.Fatalf("counts = %v, want %v and some transactions", got, want)
	}
	for i, tx := range d.GetTransactions() {
		if tx.GetKey().GetExternalId() != m.Transactions[i].Key.ExternalID {
			t.Errorf("transactions[%d] key = %q, want %q", i, tx.GetKey().GetExternalId(), m.Transactions[i].Key.ExternalID)
		}
		for j, p := range tx.GetPostings() {
			if want := m.Transactions[i].Postings[j].Amount.Amount().String(); p.GetAmount().GetAmount() != want {
				t.Errorf("transactions[%d].postings[%d] amount = %q, want %q", i, j, p.GetAmount().GetAmount(), want)
			}
		}
	}
}

func TestFXRates(t *testing.T) {
	t.Parallel()
	rates := []valuation.FXRate{
		{Currency: cur(t, "USD"), Date: date(t, "2025-01-02"), Rate: dec(t, "4.1219"), Source: "nbp-a"},
		{Currency: cur(t, "HUF"), Date: date(t, "2025-01-02"), Rate: dec(t, "0.010350"), Source: "nbp-a"},
	}
	want := []*connectorv1.FxRate{
		{Currency: "USD", Date: "2025-01-02", Rate: "4.1219"},
		{Currency: "HUF", Date: "2025-01-02", Rate: "0.01035"},
	}
	if diff := cmp.Diff(want, delivery.FXRates(rates), protocmp.Transform()); diff != "" {
		t.Errorf("FXRates (-want +got):\n%s", diff)
	}
	if got := delivery.FXRates(nil); got == nil || len(got) != 0 {
		t.Errorf("FXRates(nil) = %v, want an empty list", got)
	}
}

func TestMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      connectorv1.Mode
		want    ingest.Mode
		wantErr error
	}{
		{in: connectorv1.Mode_MODE_OPENING, want: ingest.ModeOpening},
		{in: connectorv1.Mode_MODE_TRANSACTIONS, want: ingest.ModeTransactions},
		{in: connectorv1.Mode_MODE_UNSPECIFIED, wantErr: delivery.ErrNoMode},
		{in: connectorv1.Mode(42), wantErr: delivery.ErrNoMode},
	}
	for _, tt := range tests {
		t.Run(tt.in.String(), func(t *testing.T) {
			t.Parallel()
			got, err := delivery.Mode(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Mode error (-want +got):\n%s", diff)
			}
			if got != tt.want {
				t.Errorf("Mode = %v, want %v", got, tt.want)
			}
		})
	}
}
