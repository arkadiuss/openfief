package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// recordedNames is an app.InstrumentNamer that records what it is handed.
type recordedNames struct {
	sources []string
	names   []ledger.InstrumentName
}

func (r *recordedNames) Deliver(_ context.Context, source string, names []ledger.InstrumentName) error {
	r.sources = append(r.sources, source)
	r.names = append(r.names, names...)
	return nil
}

// pendingStore is fakeTickerStore with one pending ticker.
type pendingStore struct {
	fakeTickerStore

	pending []ingest.PendingTicker
}

func (p *pendingStore) PendingTickers(context.Context) ([]ingest.PendingTicker, error) {
	return p.pending, nil
}

// TestTickersConfirmNamesTheInstrument checks that confirming a ticker's
// ISIN names the instrument by the institution's name of the ticker, as
// delivered by its instance (D-197).
func TestTickersConfirmNamesTheInstrument(t *testing.T) {
	t.Parallel()
	isin, err := ledger.ParseISIN("PLTESTA00002")
	if err != nil {
		t.Fatal(err)
	}
	store := &pendingStore{pending: []ingest.PendingTicker{
		{Instance: "xtb-main", Ticker: ingest.Ticker{Symbol: "OTHER.PL", Name: "Other SA"}},
		{Instance: "xtb-main", Ticker: ingest.Ticker{Symbol: "TESTA.PL", Name: "Synthetic SA"}},
	}}
	names := &recordedNames{}
	tickers := app.Tickers{Owners: fakeOwners{}, Store: store, Ledger: &fakeLedger{}, Names: names}
	if _, err := tickers.Confirm(t.Context(), "https://ui.example.com", "owner",
		ingest.TickerConfirmation{Instance: "xtb-main", Symbol: "TESTA.PL", ISIN: isin}); err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	want := []ledger.InstrumentName{{ISIN: isin, Name: "Synthetic SA"}}
	if diff := cmp.Diff(want, names.names, cmp.Comparer(ledger.ISIN.Equal)); diff != "" || len(names.sources) != 1 || names.sources[0] != "xtb-main" {
		t.Errorf("names (-want +got):\n%s\nsources %v", diff, names.sources)
	}
}

// TestConnectorsSyncNamesInstruments checks that a sync's instrument names
// are handed to Names as delivered by the instance (D-197).
func TestConnectorsSyncNamesInstruments(t *testing.T) {
	t.Parallel()
	isin, err := ledger.ParseISIN("PLTESTA00002")
	if err != nil {
		t.Fatal(err)
	}
	d := delivery(t)
	d.InstrumentNames = []ledger.InstrumentName{{ISIN: isin, Name: "SYNTHETIC ETF"}}
	names := &recordedNames{}
	c := connectors(t, &fakeCaller{delivery: d}, &fakeLedger{}, &fakeStores{})
	c.Names = names
	if _, err := c.Sync(t.Context(), "ibkr-main", &app.SyncCall{ID: "b-1", Mode: ingest.ModeTransactions}); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if diff := cmp.Diff(d.InstrumentNames, names.names, cmp.Comparer(ledger.ISIN.Equal)); diff != "" || len(names.sources) != 1 || names.sources[0] != "ibkr-main" {
		t.Errorf("names (-want +got):\n%s\nsources %v", diff, names.sources)
	}
}
