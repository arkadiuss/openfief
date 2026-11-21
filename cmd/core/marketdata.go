package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	"github.com/arkadiuss/openfief/internal/ingest"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationconnect "github.com/arkadiuss/openfief/internal/valuation/connectapi"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
)

// valuationOwners is ownerSessions answering with the valuation context's
// errors.
func valuationOwners(auth *identityapp.Auth, origins identity.Origins) *ownerSessions {
	return &ownerSessions{auth: auth, origins: origins, noSession: valuationapp.ErrNoSession, notOwner: valuationapp.ErrNotOwner}
}

// newRankings returns the owner's ranking of market-data sources (D-94)
// over pool, l and vstore, with the owner's sessions checked by owners;
// nil owners serves no owner, as the connector commands need none.
func newRankings(pool *pgxpool.Pool, l *ledgerapp.Ledger, vstore *valuationpg.Store, owners *ownerSessions) *valuationapp.Rankings {
	planner := valuationapp.NewPlanner(l, ledgerpg.New(joined(pool)))
	if owners == nil {
		return valuationapp.NewRankings(planner, vstore, nil)
	}
	return valuationapp.NewRankings(planner, vstore, owners).WithAudit(owners)
}

// withMarketDataService returns h with the MarketDataService mounted next
// to it (D-94, D-79): its calls go to the rankings' handler, admitted like
// the AuthService's, view plugins only, every other path to h. Nothing may
// store its responses.
func withMarketDataService(h http.Handler, rankings *valuationapp.Rankings, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := valuationconnect.New(rankings, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		handler.ServeHTTP(w, r)
	}))
	mux.Handle("/", h)
	return mux
}

// sourceInputs records a delivery's market-data sources through the
// rankings (D-94).
type sourceInputs struct {
	rankings *valuationapp.Rankings
}

// StoreSources records sources.
func (s sourceInputs) StoreSources(ctx context.Context, _ ingest.BatchID, sources []valuation.Source) error {
	if err := s.rankings.RecordSources(ctx, sources); err != nil {
		return fmt.Errorf("valuation: %w", err)
	}
	return nil
}

// priceRanking is the owner's ranking of the sources of prices (D-94) as the
// quote lookup asks for it.
type priceRanking struct {
	store *valuationpg.Store
}

var _ ingestapp.PriceRanking = priceRanking{}

// RankedPriceSources returns the names of the sources of prices, the
// highest-ranked first.
func (p priceRanking) RankedPriceSources(ctx context.Context) ([]string, error) {
	r, err := p.store.SourceRanking(ctx, valuation.KindPrices)
	if err != nil {
		return nil, fmt.Errorf("ranking of the sources of prices: %w", err)
	}
	return r.Names(), nil
}

// heldSources finds the source an instrument the owner holds is valued from,
// through the sources of the valuation's price series (D-94).
type heldSources struct {
	store *valuationpg.Store
}

var _ ingestapp.HeldSources = heldSources{}

// ValuationSource returns the connector instance whose prices value the
// instrument query names, a ticker or an ISIN.
func (h heldSources) ValuationSource(ctx context.Context, query string) (string, error) {
	series, err := h.store.SeriesSources(ctx)
	if err != nil {
		return "", fmt.Errorf("sources of the price series: %w", err)
	}
	source, _ := valuation.PriceSourceOfQuery(series, query)
	return source, nil
}

// marketQuoteOption has the views look quotes up through the owner's
// market-data connector, the highest-ranked source of prices (D-226, D-94): the
// connector client entries, the running instances and what their manifests
// declare, which D-123 lets be asked. The core makes no request of its own: the
// connector plugin alone reaches its provider, under its egress grant.
//
// A demo answers from its synthetic data alone (D-226): with --demo-sessions
// the views get no source, so a lookup is a failed precondition and the read
// API says quotes are not available, even where an installed connector holds
// a provider key.
func marketQuoteOption(cfg *serveConfig, entries *entryCaller, vstore *valuationpg.Store, now func() time.Time) viewsapp.Option {
	if cfg.Auth.DemoSessions {
		return viewsapp.WithMarketQuotes(&viewsapp.MarketQuotesConfig{})
	}
	return viewsapp.WithMarketQuotes(&viewsapp.MarketQuotesConfig{
		Source: &ingestapp.Quotes{Lookup: entries, Instances: entries, Declarations: entries, Ranking: priceRanking{store: vstore}, Held: heldSources{store: vstore}},
		Today:  valuationpg.TodayIn(cfg.EndOfDay.Location, now), Now: now,
	})
}
