package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
)

// marketDataInstance is a discovery that finds one running instance, eodhd,
// whose manifest declares a market-data connector and whose grant approves it.
type marketDataInstance struct{}

func (marketDataInstance) ConnectorInstances(context.Context) ([]string, error) {
	return []string{"eodhd"}, nil
}

func (marketDataInstance) Declarations(context.Context, string) (ingest.Declarations, error) {
	return ingest.Declarations{MarketData: true}, nil
}

// TestADemoLooksNoQuotesUp checks D-226 on the demo: with --demo-sessions the
// views have no market quote source, so a lookup is refused before any
// connector is asked and the read API says quotes are not available, even
// though a market-data connector is installed and approved, as a demo's
// visitors must never spend a provider key or reach a real provider. The store
// is nil: a lookup that went on to the ranking of sources would panic, which
// is how this fails if the demo ever gets a source.
func TestADemoLooksNoQuotesUp(t *testing.T) {
	t.Parallel()
	cfg := &serveConfig{}
	cfg.Auth.DemoSessions = true
	cfg.EndOfDay.Location = time.UTC
	entries := &entryCaller{}
	entries.setDiscovery(marketDataInstance{})
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	now := func() time.Time { return time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC) }
	views, err := viewsapp.New(nil, nil, nil, nil, pln, pln, marketQuoteOption(cfg, entries, nil, now))
	if err != nil {
		t.Fatalf("viewsapp.New: %v", err)
	}
	if views.MarketQuotesAvailable(t.Context()) {
		t.Error("the demo says quotes are available")
	}
	if _, err := views.MarketQuote(t.Context(), "NVDA", kernel.Date{}, kernel.Date{}); !errors.Is(err, viewsapp.ErrQuotesUnavailable) {
		t.Errorf("MarketQuote on the demo = %v, want ErrQuotesUnavailable", err)
	}
}
