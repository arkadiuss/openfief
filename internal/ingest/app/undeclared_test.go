package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// deliveryAttention records the attention items raised and cleared, as
// "raise kind subject" and "clear kind subject".
type deliveryAttention struct {
	calls   []string
	details []string
}

func (f *deliveryAttention) Raise(_ context.Context, kind, subject, detail string) error {
	f.calls = append(f.calls, "raise "+kind+" "+subject)
	f.details = append(f.details, detail)
	return nil
}

func (f *deliveryAttention) Clear(_ context.Context, kind, subject string) error {
	f.calls = append(f.calls, "clear "+kind+" "+subject)
	return nil
}

// TestConnectorsUndeclaredDeliveries checks D-179 and D-186 at the ingest
// path: bond series terms, index levels and index readings from an
// instance that may not deliver them, because its signed manifest doesn't declare them or its grant
// doesn't approve that, are dropped as one ingest problem per kind and
// raise the instance's one attention item while the rest of the delivery
// is imported; those it may deliver are stored with the instance as their
// source, and a delivery that drops nothing clears the item.
func TestConnectorsUndeclaredDeliveries(t *testing.T) {
	t.Parallel()
	in := pricingInstrument(t, "PL0000000014", "EDO0336")
	series, err := pl.ParseSeries("EDO0336")
	if err != nil {
		t.Fatal(err)
	}
	terms := ingest.BondTerms{Instrument: in, Terms: pl.Terms{Series: series, EarlyRedemptionFee: pl.Nominal()}, Source: "spoofed"}
	level := valuation.IndexLevel{Index: "FAKE20", Date: march2(t), Level: kernel.NewDecimal(1500, 0), Source: "spoofed"}
	reading := delivery(t).IndexReadings[0]
	stored := reading
	stored.Source = "bonds-1"
	kind := ingest.KindUndeclaredDeliveries
	tests := []struct {
		name         string
		declared     ingest.Declarations
		wantTerms    []ingest.BondTerms
		wantLevels   []valuation.IndexLevel
		wantReadings []ingest.IndexReading
		wantSources  []valuation.Source
		wantCalls    []string
		problems     int
	}{
		{
			name:      "none declared: all dropped, the item raised",
			wantCalls: []string{"raise " + kind + " bonds-1"}, problems: 3,
		},
		{
			name: "bond terms declared: stored with the instance, the levels and readings dropped", declared: ingest.Declarations{BondTerms: true},
			wantTerms: []ingest.BondTerms{{Instrument: in, Terms: terms.Terms, Source: "bonds-1"}},
			wantCalls: []string{"raise " + kind + " bonds-1"}, problems: 2,
		},
		{
			name: "index readings declared: stored with the instance and ranked as a source, the rest dropped", declared: ingest.Declarations{IndexReadings: true},
			wantReadings: []ingest.IndexReading{stored},
			wantSources:  []valuation.Source{{Kind: valuation.KindIndexReadings, Name: "bonds-1", Institution: "xtb"}},
			wantCalls:    []string{"raise " + kind + " bonds-1"}, problems: 2,
		},
		{
			name: "all declared: all stored with the instance, the item cleared", declared: ingest.Declarations{BondTerms: true, IndexLevels: true, IndexReadings: true},
			wantTerms:    []ingest.BondTerms{{Instrument: in, Terms: terms.Terms, Source: "bonds-1"}},
			wantLevels:   []valuation.IndexLevel{{Index: "FAKE20", Date: level.Date, Level: level.Level, Source: "bonds-1"}},
			wantReadings: []ingest.IndexReading{stored},
			wantSources:  []valuation.Source{{Kind: valuation.KindIndexReadings, Name: "bonds-1", Institution: "xtb"}},
			wantCalls:    []string{"clear " + kind + " bonds-1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := delivery(t)
			d.Prices, d.FXRates = nil, nil
			d.BondTerms, d.IndexLevels, d.IndexReadings = []ingest.BondTerms{terms}, []valuation.IndexLevel{level}, []ingest.IndexReading{reading}
			stores, attention := &fakeStores{}, &deliveryAttention{}
			c := connectors(t, &fakeCaller{delivery: d}, &fakeLedger{}, stores)
			c.Declarations = fakeDeclarations{byInstance: map[string]ingest.Declarations{"bonds-1": tt.declared}}
			c.Attention = attention
			got, err := c.Sync(t.Context(), "bonds-1", &app.SyncCall{ID: "b-1", Mode: ingest.ModeTransactions})
			if err != nil {
				t.Fatalf("Sync: %v", err)
			}
			if diff := cmp.Diff(
				[]any{tt.wantTerms, tt.wantLevels, tt.wantReadings, tt.wantSources, tt.wantCalls},
				[]any{stores.terms, stores.levels, stores.readings, stores.sources, attention.calls},
			); diff != "" {
				t.Errorf("stored terms, levels, readings, sources and attention calls (-want +got):\n%s", diff)
			}
			wantProblems := ingest.UndeclaredDeliveries(tt.declared, 1, 1, 1).Problems
			got.Batch.Transactions -= len(d.Mapping.Transactions)
			if diff := cmp.Diff([]any{tt.problems, append(mapping(t).Problems, wantProblems...), 0},
				[]any{len(wantProblems), got.Batch.Problems, got.Batch.Transactions}); diff != "" {
				t.Errorf("problems, batch problems and transactions not imported (-want +got):\n%s", diff)
			}
		})
	}
}

// TestConnectorsUndeclaredDeliveriesItem checks that the attention item
// stays one per instance with a detail that doesn't change while the same
// kinds keep being dropped, so an acknowledged item isn't reopened, and
// that the instance's next delivery that drops nothing clears it (D-179).
func TestConnectorsUndeclaredDeliveriesItem(t *testing.T) {
	t.Parallel()
	attention := &deliveryAttention{}
	undeclared := delivery(t)
	clean := delivery(t)
	clean.IndexLevels, clean.IndexReadings = nil, nil
	for i, d := range []app.Delivery{undeclared, undeclared, clean} {
		c := connectors(t, &fakeCaller{delivery: d}, &fakeLedger{}, &fakeStores{})
		c.Declarations, c.Attention = fakeDeclarations{}, attention
		if _, err := c.Sync(t.Context(), "prices-1", &app.SyncCall{ID: ingest.BatchID("b-" + string(rune('1'+i))), Mode: ingest.ModeTransactions}); err != nil {
			t.Fatalf("Sync %d: %v", i, err)
		}
	}
	kind := ingest.KindUndeclaredDeliveries
	want := []string{"raise " + kind + " prices-1", "raise " + kind + " prices-1", "clear " + kind + " prices-1"}
	if diff := cmp.Diff(want, attention.calls); diff != "" {
		t.Errorf("attention calls (-want +got):\n%s", diff)
	}
	if len(attention.details) != 2 || attention.details[0] != attention.details[1] || attention.details[0] == "" {
		t.Errorf("details %q; want the same detail both times", attention.details)
	}
}
