package delivery_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// lotMessage is a synthetic lot of 40 EDO0536 bonds bought on 2026-05-04.
func lotMessage() *connectorv1.BondLot {
	return &connectorv1.BondLot{
		Series: "EDO0536", Quantity: "40", PurchaseDate: "2026-05-04", FirstPeriodRate: "0.068", Margin: "0.02",
		EarlyRedemptionFee: &connectorv1.Money{Amount: "2.00", Currency: "PLN"},
	}
}

// edoTerms are lotMessage's series terms.
func edoTerms(t *testing.T) pl.Terms {
	t.Helper()
	series, err := pl.ParseSeries("EDO0536")
	if err != nil {
		t.Fatalf("ParseSeries: %v", err)
	}
	return pl.Terms{
		Series: series, FirstPeriodRate: dec(t, "0.068"), Margin: dec(t, "0.02"),
		EarlyRedemptionFee: kernel.NewMoney(dec(t, "2.00"), pl.Nominal().Currency()),
	}
}

func TestBondLots(t *testing.T) {
	t.Parallel()
	// D-97: rates or a fee left empty are the connector's to fill in.
	noRates, noFee := lotMessage(), lotMessage()
	noRates.FirstPeriodRate, noRates.Margin = "", ""
	noFee.EarlyRedemptionFee = nil
	got, err := delivery.BondLots([]*connectorv1.BondLot{lotMessage(), noRates, noFee})
	if err != nil {
		t.Fatalf("BondLots: %v", err)
	}
	purchased, err := kernel.ParseDate("2026-05-04")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	q := kernel.NewQuantity(dec(t, "40"))
	withoutRates, withoutFee := edoTerms(t), edoTerms(t)
	withoutRates.FirstPeriodRate, withoutRates.Margin = kernel.Decimal{}, kernel.Decimal{}
	withoutFee.EarlyRedemptionFee = kernel.Money{}
	want := []ingest.BondLot{
		{Terms: edoTerms(t), Quantity: q, Purchased: purchased},
		{Terms: withoutRates, PublishedRates: true, Quantity: q, Purchased: purchased},
		{Terms: withoutFee, ReviewedFee: true, Quantity: q, Purchased: purchased},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BondLots (-want +got):\n%s", diff)
	}
}

func TestBondLotsRefuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(l *connectorv1.BondLot)
	}{
		{name: "no series", edit: func(l *connectorv1.BondLot) { l.Series = "" }},
		{name: "unknown bond type", edit: func(l *connectorv1.BondLot) { l.Series = "XYZ0536" }},
		{name: "quantity not a number", edit: func(l *connectorv1.BondLot) { l.Quantity = "forty" }},
		{name: "date not YYYY-MM-DD", edit: func(l *connectorv1.BondLot) { l.PurchaseDate = "04.05.2026" }},
		{name: "rate not a number", edit: func(l *connectorv1.BondLot) { l.FirstPeriodRate = "6.8%" }},
		{name: "a rate without the margin", edit: func(l *connectorv1.BondLot) { l.Margin = "" }},
		{name: "a margin without the rate", edit: func(l *connectorv1.BondLot) { l.FirstPeriodRate = "" }},
		{name: "margin not a number", edit: func(l *connectorv1.BondLot) { l.Margin = "2%" }},
		{name: "fee without a currency", edit: func(l *connectorv1.BondLot) { l.EarlyRedemptionFee.Currency = "" }},
		{name: "fee not a number", edit: func(l *connectorv1.BondLot) { l.EarlyRedemptionFee.Amount = "two" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := lotMessage()
			tt.edit(l)
			if _, err := delivery.BondLots([]*connectorv1.BondLot{l}); !cmp.Equal(delivery.ErrMalformedLot, err, cmpopts.EquateErrors()) {
				t.Errorf("BondLots error = %v, want ErrMalformedLot", err)
			}
		})
	}
}

func TestBondTerms(t *testing.T) {
	t.Parallel()
	isin, err := ledger.ParseISIN("PLEDO0536000")
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	made, err := kernel.ParseDate("2026-10-26")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	published := edoTerms(t)
	published.Attribution = &pl.Attribution{
		Publisher: "Ministerstwo Finansów", Dataset: "Dane dotyczące obligacji detalicznych", License: "CC BY 4.0",
		URL: "https://www.example.org/attachment/x", Made: made, Processing: "read",
		Fetched: time.Date(2026, time.October, 29, 7, 0, 0, 0, time.FixedZone("CET", 3600)),
	}
	in := ledger.Instrument{ISIN: isin, Symbol: "EDO0536"}
	got := delivery.BondTerms([]ingest.BondTerms{{Instrument: in, Terms: edoTerms(t)}, {Instrument: in, Terms: published}})
	want := []*connectorv1.BondTerms{
		{
			Instrument:      &connectorv1.Instrument{Isin: "PLEDO0536000", Symbol: "EDO0536"},
			FirstPeriodRate: "0.068", Margin: "0.02", EarlyRedemptionFee: &connectorv1.Money{Amount: "2", Currency: "PLN"},
		},
		{
			Instrument:      &connectorv1.Instrument{Isin: "PLEDO0536000", Symbol: "EDO0536"},
			FirstPeriodRate: "0.068", Margin: "0.02", EarlyRedemptionFee: &connectorv1.Money{Amount: "2", Currency: "PLN"},
			// D-97: the attribution travels with terms read from published data.
			Attribution: &connectorv1.BondTermsAttribution{
				Publisher: "Ministerstwo Finansów", Dataset: "Dane dotyczące obligacji detalicznych", License: "CC BY 4.0",
				Url: "https://www.example.org/attachment/x", Made: "2026-10-26", Fetched: "2026-10-29T06:00:00Z", Processing: "read",
			},
		},
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("BondTerms (-want +got):\n%s", diff)
	}
}

func TestBondLotEntries(t *testing.T) {
	t.Parallel()
	exchanged := lotMessage()
	exchanged.Purchase = connectorv1.BondPurchase_BOND_PURCHASE_EXCHANGE
	got, err := delivery.BondLotEntries(&connectorv1.ImportBondLotsRequest{
		GoLive: "2026-06-01", Lots: []*connectorv1.BondLot{exchanged}, AccountId: "acc-bonds",
		Deleted: []*connectorv1.BondLotKey{{Series: "EDO0536", PurchaseDate: "2026-05-10"}},
		Booked: []*connectorv1.BookedTransaction{{
			Key:  &connectorv1.TransactionKey{Source: "treasury-bonds:acc-bonds", ExternalId: "lot:EDO0536:2026-05-10:redemption"},
			Date: "2036-05-10",
		}},
	})
	if err != nil {
		t.Fatalf("BondLotEntries: %v", err)
	}
	day := func(s string) kernel.Date {
		d, err := kernel.ParseDate(s)
		if err != nil {
			t.Fatalf("ParseDate: %v", err)
		}
		return d
	}
	terms := edoTerms(t)
	want := ingest.BondLotEntries{
		Account: "acc-bonds",
		GoLive:  day("2026-06-01"),
		Lots:    []ingest.BondLot{{Terms: terms, Quantity: kernel.NewQuantity(dec(t, "40")), Purchased: day("2026-05-04"), Purchase: ingest.PurchaseExchange}},
		Deleted: []ingest.BondLotKey{{Series: terms.Series, Purchased: day("2026-05-10")}},
		Booked: []ledger.TransactionHeader{{
			Key:     ledger.TransactionKey{Source: "treasury-bonds:acc-bonds", ExternalID: "lot:EDO0536:2026-05-10:redemption"},
			Account: "acc-bonds", Date: day("2036-05-10"),
		}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BondLotEntries (-want +got):\n%s", diff)
	}
	for name, req := range map[string]*connectorv1.ImportBondLotsRequest{
		"a malformed go-live":                  {GoLive: "1.6.2026", Lots: []*connectorv1.BondLot{lotMessage()}},
		"a deleted series":                     {GoLive: "2026-06-01", Deleted: []*connectorv1.BondLotKey{{Series: "EDO536", PurchaseDate: "2026-05-10"}}},
		"a deleted date":                       {GoLive: "2026-06-01", Deleted: []*connectorv1.BondLotKey{{Series: "EDO0536", PurchaseDate: "10.5.2026"}}},
		"an unknown purchase":                  {GoLive: "2026-06-01", Lots: []*connectorv1.BondLot{withPurchase(9)}},
		"a booked transaction without its key": {Booked: []*connectorv1.BookedTransaction{{Date: "2025-10-10"}}},
		"a booked transaction without its date": {Booked: []*connectorv1.BookedTransaction{{
			Key: &connectorv1.TransactionKey{Source: "treasury-bonds:acc-bonds", ExternalId: "x"},
		}}},
	} {
		if _, err := delivery.BondLotEntries(req); !errors.Is(err, delivery.ErrMalformedLot) {
			t.Errorf("%s: %v, want ErrMalformedLot", name, err)
		}
	}
}

// withPurchase returns the synthetic lot message bought by purchase p.
func withPurchase(p connectorv1.BondPurchase) *connectorv1.BondLot {
	l := lotMessage()
	l.Purchase = p
	return l
}
