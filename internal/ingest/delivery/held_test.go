package delivery_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// heldSync is a sync request whose one account holds one lot of EDO0536,
// with its terms; change edits it before it is read.
func heldSync(change func(*connectorv1.SyncRequest)) *connectorv1.SyncRequest {
	edo := &connectorv1.Instrument{Isin: "PLEDO0536000", Symbol: "EDO0536"}
	pln := func(a string) *connectorv1.Money { return &connectorv1.Money{Amount: a, Currency: "PLN"} }
	req := &connectorv1.SyncRequest{
		Accounts: []*connectorv1.AccountHoldings{{
			AccountId: "acc-bonds", GoLive: "2026-05-15", TaxWrapper: "taxable",
			Holdings: []*connectorv1.Holding{{Instrument: edo, Quantity: "40"}},
			Lots:     []*connectorv1.OpenLot{{Instrument: edo, Quantity: "40", Opened: "2026-05-04", Cost: pln("4000")}},
		}},
		BondTerms: []*connectorv1.BondTerms{{Instrument: edo, FirstPeriodRate: "0.06", Margin: "0.02", EarlyRedemptionFee: pln("2")}},
	}
	if change != nil {
		change(req)
	}
	return req
}

func TestHeld(t *testing.T) {
	t.Parallel()
	held, terms, err := delivery.Held(heldSync(nil))
	if err != nil {
		t.Fatalf("Held: %v", err)
	}
	parts := make([]string, 0, 1+len(held)+len(terms))
	parts = append(parts, fmt.Sprint(len(held), " ", len(terms)))
	for _, a := range held {
		parts = append(parts, fmt.Sprint(a.Account, " ", a.GoLive, " ", a.Wrapper, " ", len(a.Holdings), " ", a.Lots[0].Opened, " ", a.Lots[0].Quantity))
	}
	for _, bt := range terms {
		parts = append(parts, bt.Terms.Series.String())
	}
	if got, want := strings.Join(parts, " "), "1 1 acc-bonds 2026-05-15 taxable 1 2026-05-04 40 EDO0536"; got != want {
		t.Errorf("Held = %q, want %q", got, want)
	}
	none, _, err := delivery.Held(&connectorv1.SyncRequest{})
	if err != nil || len(none) != 0 {
		t.Errorf("Held of a market-data sync = %v, %v; want nothing", none, err)
	}
}

func TestHeldRefuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		change func(*connectorv1.SyncRequest)
	}{
		{name: "no account ID", change: func(r *connectorv1.SyncRequest) { r.Accounts[0].AccountId = "" }},
		{name: "a go-live that is no date", change: func(r *connectorv1.SyncRequest) { r.Accounts[0].GoLive = "15.05.2026" }},
		{name: "a holding with a bad ISIN", change: func(r *connectorv1.SyncRequest) {
			r.Accounts[0].Holdings[0].Instrument = &connectorv1.Instrument{Isin: "PLEDO0536009"}
		}},
		{name: "a lot without cost", change: func(r *connectorv1.SyncRequest) { r.Accounts[0].Lots[0].Cost = nil }},
		{name: "a fraction that is no number", change: func(r *connectorv1.SyncRequest) { r.Accounts[0].Lots[0].Quantity = "forty" }},
		{name: "terms of no series", change: func(r *connectorv1.SyncRequest) {
			r.BondTerms[0].Instrument = &connectorv1.Instrument{Isin: "US0378331005", Symbol: "AAPL"}
		}},
		{name: "terms without a fee", change: func(r *connectorv1.SyncRequest) { r.BondTerms[0].EarlyRedemptionFee = nil }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, _, err := delivery.Held(heldSync(tt.change)); !errors.Is(err, delivery.ErrMalformedHoldings) {
				t.Errorf("Held = %v, want %v", err, delivery.ErrMalformedHoldings)
			}
		})
	}
}
