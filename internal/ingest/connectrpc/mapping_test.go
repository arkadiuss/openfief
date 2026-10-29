package connectrpc

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

const testISIN = "PLTEST000010"

// testBondISIN is the derived ISIN of the bond series EDO0536.
const testBondISIN = "PLEDO0536000"

// testDelivery is a synthetic delivery for batch b-1: one buy of 2 units for
// 100.50 USD, a holding assertion, a skipped row, a price, an FX rate, an
// index reading, an index level, the terms of a bond series and a
// cross-check price.
func testDelivery() *connectorv1.Delivery {
	instrument := &connectorv1.Instrument{Isin: testISIN, Symbol: "TST"}
	holding := &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_HOLDING, Instrument: instrument}
	return &connectorv1.Delivery{
		BatchId: "b-1", Institution: "xtb",
		Transactions: []*connectorv1.Transaction{{
			Key:       &connectorv1.TransactionKey{Source: "acc-1", ExternalId: "T1"},
			AccountId: "acc-1", Kind: connectorv1.TransactionKind_TRANSACTION_KIND_BUY, Date: "2026-03-02",
			Description: "buy",
			Postings: []*connectorv1.Posting{
				{Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_CASH, Currency: "USD"}, Amount: &connectorv1.Money{Amount: "-100.50", Currency: "USD"}},
				{Book: holding, Amount: &connectorv1.Money{Amount: "100.50", Currency: "USD"}, Quantity: "2"},
			},
		}},
		BalanceAssertions: []*connectorv1.BalanceAssertion{{AccountId: "acc-1", Date: "2026-03-02", Book: holding, Balance: "2"}},
		Problems:          []*connectorv1.Problem{{Kind: connectorv1.ProblemKind_PROBLEM_KIND_SKIPPED, Ref: "row 3", Reason: "summary row"}},
		Prices: []*connectorv1.Price{{
			Instrument: instrument, Date: "2026-03-02", Close: &connectorv1.Money{Amount: "50.25", Currency: "USD"},
		}},
		FxRates:       []*connectorv1.FxRate{{Currency: "USD", Date: "2026-03-02", Rate: "3.9876"}},
		IndexReadings: []*connectorv1.IndexReading{{Index: connectorv1.Index_INDEX_CPI, Date: "2026-02-01", Value: "0.049"}},
		IndexLevels:   []*connectorv1.IndexLevel{{Index: "FAKE20", Date: "2026-03-02", Level: "1522.60"}},
		BondTerms: []*connectorv1.BondTerms{{
			Instrument:      &connectorv1.Instrument{Isin: testBondISIN, Symbol: "EDO0536"},
			FirstPeriodRate: "0.06", Margin: "0.02", EarlyRedemptionFee: &connectorv1.Money{Amount: "2.00", Currency: "PLN"},
		}},
		CrossCheckPrices: []*connectorv1.Price{{
			Instrument: instrument, Date: "2026-03-02", Close: &connectorv1.Money{Amount: "50.30", Currency: "USD"},
		}},
	}
}

// testBondTerms are testDelivery's bond terms in the domain.
func testBondTerms(t *testing.T) ingest.BondTerms {
	t.Helper()
	isin, err := ledger.ParseISIN(testBondISIN)
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	series, err := pl.ParseSeries("EDO0536")
	if err != nil {
		t.Fatalf("ParseSeries: %v", err)
	}
	return ingest.BondTerms{
		Instrument: ledger.Instrument{ISIN: isin, Symbol: "EDO0536"},
		Terms: pl.Terms{
			Series: series, FirstPeriodRate: mustDecimal(t, "0.06"), Margin: mustDecimal(t, "0.02"),
			EarlyRedemptionFee: mustMoney(t, "2.00", "PLN"),
		},
	}
}

func mustDecimal(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal: %v", err)
	}
	return d
}

func mustDate(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	return d
}

func mustMoney(t *testing.T, amount, currency string) kernel.Money {
	t.Helper()
	c, err := kernel.ParseCurrency(currency)
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	return kernel.NewMoney(mustDecimal(t, amount), c)
}

func testInstrument(t *testing.T) ledger.Instrument {
	t.Helper()
	isin, err := ledger.ParseISIN(testISIN)
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	return ledger.Instrument{ISIN: isin, Symbol: "TST"}
}

func TestDeliveryFrom(t *testing.T) {
	t.Parallel()
	got, err := deliveryFrom(testDelivery(), "b-1", "w-2")
	if err != nil {
		t.Fatalf("deliveryFrom: %v", err)
	}
	day := mustDate(t, "2026-03-02")
	usd := mustMoney(t, "100.50", "USD")
	holding := ledger.HoldingBook(testInstrument(t))
	want := app.Delivery{
		Institution: "xtb", Watermark: "w-2",
		Prices:  []valuation.Price{{Instrument: testInstrument(t).String(), Date: day, Close: mustMoney(t, "50.25", "USD")}},
		FXRates: []valuation.FXRate{{Currency: usd.Currency(), Date: day, Rate: mustDecimal(t, "3.9876")}},
		IndexReadings: []ingest.IndexReading{{
			Index: ingest.IndexCPI, Date: mustDate(t, "2026-02-01"), Value: mustDecimal(t, "0.049"),
		}},
		IndexLevels: []valuation.IndexLevel{{Index: "FAKE20", Date: day, Level: mustDecimal(t, "1522.60")}},
		BondTerms:   []ingest.BondTerms{testBondTerms(t)},
		CrossCheckPrices: []valuation.CrossCheckPrice{
			{Instrument: testInstrument(t).String(), Date: day, Close: mustMoney(t, "50.30", "USD")},
		},
		Mapping: ingest.Mapping{
			Transactions: []ledger.Transaction{{
				Key: ledger.TransactionKey{Source: "acc-1", ExternalID: "T1"}, Account: "acc-1",
				Kind: ledger.KindBuy, Date: day, Description: "buy",
				Postings: []ledger.Posting{
					{Book: ledger.CashBook(usd.Currency()), Amount: usd.Neg()},
					{Book: holding, Amount: usd, Quantity: kernel.NewQuantity(mustDecimal(t, "2"))},
				},
			}},
			Assertions: []ledger.BalanceAssertion{{Account: "acc-1", Date: day, Book: holding, Balance: mustDecimal(t, "2")}},
			Problems:   []ingest.Problem{ingest.Skipped("row 3", "summary row")},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("deliveryFrom (-want +got):\n%s", diff)
	}
	for i := range got.Mapping.Transactions {
		if err := got.Mapping.Transactions[i].Validate(); err != nil {
			t.Errorf("translated transaction is not valid: %v", err)
		}
	}
}

func TestDeliveryFromRejects(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(r *connectorv1.Delivery)
	}{
		{name: "another batch", edit: func(r *connectorv1.Delivery) { r.BatchId = "b-0" }},
		{name: "bad price", edit: func(r *connectorv1.Delivery) { r.Prices[0].Close.Amount = "0" }},
		{name: "zero cross-check price", edit: func(r *connectorv1.Delivery) { r.CrossCheckPrices[0].Close.Amount = "0" }},
		{name: "cross-check price with a bad ISIN", edit: func(r *connectorv1.Delivery) { r.CrossCheckPrices[0].Instrument.Isin = "XX0000000000" }},
		{name: "bad FX rate", edit: func(r *connectorv1.Delivery) { r.FxRates[0].Rate = "-1" }},
		{name: "bad index reading", edit: func(r *connectorv1.Delivery) { r.IndexReadings[0].Index = 0 }},
		{name: "index level without an index", edit: func(r *connectorv1.Delivery) { r.IndexLevels[0].Index = "" }},
		{name: "zero index level", edit: func(r *connectorv1.Delivery) { r.IndexLevels[0].Level = "0" }},
		{name: "index level with a bad date", edit: func(r *connectorv1.Delivery) { r.IndexLevels[0].Date = "2026-3-2" }},
		{name: "exponent index level", edit: func(r *connectorv1.Delivery) { r.IndexLevels[0].Level = "1.5e3" }},
		{name: "exponent amount", edit: func(r *connectorv1.Delivery) { r.Transactions[0].Postings[0].Amount.Amount = "-1.005e2" }},
		{name: "grouped amount", edit: func(r *connectorv1.Delivery) { r.Transactions[0].Postings[0].Amount.Amount = "1,000.50" }},
		{name: "missing amount", edit: func(r *connectorv1.Delivery) { r.Transactions[0].Postings[0].Amount = nil }},
		{name: "lower-case currency", edit: func(r *connectorv1.Delivery) { r.Transactions[0].Postings[0].Amount.Currency = "usd" }},
		{name: "bad quantity", edit: func(r *connectorv1.Delivery) { r.Transactions[0].Postings[1].Quantity = "two" }},
		{name: "date with time", edit: func(r *connectorv1.Delivery) { r.Transactions[0].Date = "2026-03-02T10:00:00Z" }},
		{name: "bad lot open date", edit: func(r *connectorv1.Delivery) { r.Transactions[0].LotOpened = "02.03.2026" }},
		{name: "no transaction kind", edit: func(r *connectorv1.Delivery) {
			r.Transactions[0].Kind = connectorv1.TransactionKind_TRANSACTION_KIND_UNSPECIFIED
		}},
		{name: "no book kind", edit: func(r *connectorv1.Delivery) { r.Transactions[0].Postings[0].Book.Kind = 0 }},
		{name: "wrong ISIN check digit", edit: func(r *connectorv1.Delivery) {
			r.Transactions[0].Postings[1].Book.Instrument = &connectorv1.Instrument{Isin: "PLTEST000011", Symbol: "TST"}
		}},
		{name: "no venue symbol", edit: func(r *connectorv1.Delivery) {
			r.Transactions[0].Postings[1].Book.Instrument = &connectorv1.Instrument{Isin: testISIN}
		}},
		{name: "bad assertion balance", edit: func(r *connectorv1.Delivery) { r.BalanceAssertions[0].Balance = "" }},
		{name: "bad assertion date", edit: func(r *connectorv1.Delivery) { r.BalanceAssertions[0].Date = "" }},
		{name: "bad assertion book", edit: func(r *connectorv1.Delivery) { r.BalanceAssertions[0].Book = nil }},
		{name: "no problem kind", edit: func(r *connectorv1.Delivery) { r.Problems[0].Kind = 0 }},
		{name: "bond terms of no series", edit: func(r *connectorv1.Delivery) { r.BondTerms[0].Instrument.Symbol = "TST" }},
		{name: "bond terms with a bad ISIN", edit: func(r *connectorv1.Delivery) { r.BondTerms[0].Instrument.Isin = "PLEDO0536001" }},
		{name: "bond terms with an exponent rate", edit: func(r *connectorv1.Delivery) { r.BondTerms[0].FirstPeriodRate = "6e-2" }},
		{name: "bond terms without a margin", edit: func(r *connectorv1.Delivery) { r.BondTerms[0].Margin = "" }},
		{name: "bond terms without a fee", edit: func(r *connectorv1.Delivery) { r.BondTerms[0].EarlyRedemptionFee = nil }},
		{name: "bond terms with a fee in EUR", edit: func(r *connectorv1.Delivery) { r.BondTerms[0].EarlyRedemptionFee.Currency = "EUR" }},
		{name: "bond terms with a negative fee", edit: func(r *connectorv1.Delivery) { r.BondTerms[0].EarlyRedemptionFee.Amount = "-2" }},
		{name: "transaction naming both accounts", edit: func(r *connectorv1.Delivery) { r.Transactions[0].BrokerAccount = "U0000001" }},
		{name: "assertion naming both accounts", edit: func(r *connectorv1.Delivery) { r.BalanceAssertions[0].BrokerAccount = "U0000001" }},
		{name: "broker account with a space", edit: func(r *connectorv1.Delivery) {
			r.Transactions[0].AccountId, r.Transactions[0].BrokerAccount = "", "U000 0001"
		}},
		{name: "listed broker account empty", edit: func(r *connectorv1.Delivery) {
			r.BrokerAccounts = []*connectorv1.BrokerAccount{{Id: ""}}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := testDelivery()
			tt.edit(r)
			if _, err := deliveryFrom(r, "b-1", ""); !errors.Is(err, ErrInvalidMessage) {
				t.Errorf("deliveryFrom error = %v, want %v", err, ErrInvalidMessage)
			}
		})
	}
}

func TestPlainDecimal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"0", true},
		{"-12.34", true},
		{"100.50", true},
		{"0.049", true},
		{"", false},
		{"-", false},
		{".5", false},
		{"5.", false},
		{"+1", false},
		{"1e3", false},
		{"1,000", false},
		{" 1", false},
		{"1.2.3", false},
		{"--1", false},
		{"1-", false},
		{"NaN", false},
	}
	for _, tt := range tests {
		if got := plainDecimal(tt.in); got != tt.want {
			t.Errorf("plainDecimal(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestValuationInputsFrom(t *testing.T) {
	t.Parallel()
	in := testInstrument(t)
	price, err := priceFrom("p", &connectorv1.Price{
		Instrument: &connectorv1.Instrument{Isin: testISIN, Symbol: "TST"}, Date: "2026-03-02",
		Close: &connectorv1.Money{Amount: "12.34", Currency: "PLN"},
	})
	if err != nil {
		t.Fatalf("priceFrom: %v", err)
	}
	wantPrice := valuation.Price{Instrument: in.String(), Date: mustDate(t, "2026-03-02"), Close: mustMoney(t, "12.34", "PLN")}
	if diff := cmp.Diff(wantPrice, price); diff != "" {
		t.Errorf("priceFrom (-want +got):\n%s", diff)
	}
	rate, err := fxRateFrom("r", &connectorv1.FxRate{Currency: "USD", Date: "2026-03-02", Rate: "3.9876"})
	if err != nil {
		t.Fatalf("fxRateFrom: %v", err)
	}
	wantRate := valuation.FXRate{Currency: mustMoney(t, "1", "USD").Currency(), Date: mustDate(t, "2026-03-02"), Rate: mustDecimal(t, "3.9876")}
	if diff := cmp.Diff(wantRate, rate); diff != "" {
		t.Errorf("fxRateFrom (-want +got):\n%s", diff)
	}
	reading, err := indexReadingFrom("i", &connectorv1.IndexReading{Index: connectorv1.Index_INDEX_REFERENCE_RATE, Date: "2026-07-03", Value: "0.0575"})
	if err != nil {
		t.Fatalf("indexReadingFrom: %v", err)
	}
	wantReading := ingest.IndexReading{Index: ingest.IndexReferenceRate, Date: mustDate(t, "2026-07-03"), Value: mustDecimal(t, "0.0575")}
	if diff := cmp.Diff(wantReading, reading); diff != "" {
		t.Errorf("indexReadingFrom (-want +got):\n%s", diff)
	}
}

func TestValuationInputsFromRejects(t *testing.T) {
	t.Parallel()
	usd := &connectorv1.Money{Amount: "1", Currency: "USD"}
	instrument := &connectorv1.Instrument{Isin: testISIN, Symbol: "TST"}
	tests := []struct {
		name string
		err  func() error
	}{
		{"zero close", func() error {
			_, err := priceFrom("p", &connectorv1.Price{Instrument: instrument, Date: "2026-03-02", Close: &connectorv1.Money{Amount: "0", Currency: "USD"}})
			return err
		}},
		{"price without instrument", func() error {
			_, err := priceFrom("p", &connectorv1.Price{Date: "2026-03-02", Close: usd})
			return err
		}},
		{"price without date", func() error {
			_, err := priceFrom("p", &connectorv1.Price{Instrument: instrument, Close: usd})
			return err
		}},
		{"negative rate", func() error {
			_, err := fxRateFrom("r", &connectorv1.FxRate{Currency: "USD", Date: "2026-03-02", Rate: "-3.9"})
			return err
		}},
		{"rate without currency", func() error {
			_, err := fxRateFrom("r", &connectorv1.FxRate{Date: "2026-03-02", Rate: "3.9"})
			return err
		}},
		{"rate without date", func() error {
			_, err := fxRateFrom("r", &connectorv1.FxRate{Currency: "USD", Rate: "3.9"})
			return err
		}},
		{"no index", func() error {
			_, err := indexReadingFrom("i", &connectorv1.IndexReading{Date: "2026-07-03", Value: "0.0575"})
			return err
		}},
		{"reading without date", func() error {
			_, err := indexReadingFrom("i", &connectorv1.IndexReading{Index: connectorv1.Index_INDEX_CPI, Value: "0.049"})
			return err
		}},
		{"percent reading", func() error {
			_, err := indexReadingFrom("i", &connectorv1.IndexReading{Index: connectorv1.Index_INDEX_CPI, Date: "2026-02-01", Value: "4.9%"})
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(ErrInvalidMessage, tt.err(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDeliveryFromMissing(t *testing.T) {
	t.Parallel()
	if _, err := deliveryFrom(nil, "b-1", ""); !errors.Is(err, ErrInvalidMessage) {
		t.Errorf("deliveryFrom(nil) error = %v, want %v", err, ErrInvalidMessage)
	}
}

func TestModeTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      ingest.Mode
		want    connectorv1.Mode
		wantErr error
	}{
		{in: ingest.ModeOpening, want: connectorv1.Mode_MODE_OPENING},
		{in: ingest.ModeTransactions, want: connectorv1.Mode_MODE_TRANSACTIONS},
		{in: 0, want: connectorv1.Mode_MODE_UNSPECIFIED, wantErr: ingest.ErrInvalidBatch},
	}
	for _, tt := range tests {
		got, err := modeTo(tt.in)
		if got != tt.want || !errors.Is(err, tt.wantErr) {
			t.Errorf("modeTo(%v) = %v, %v; want %v, %v", tt.in, got, err, tt.want, tt.wantErr)
		}
	}
}
