package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	"github.com/arkadiuss/openfief/internal/ingest"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationconnect "github.com/arkadiuss/openfief/internal/valuation/connectapi"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// valuationOwners is ownerSessions answering with the valuation context's
// errors.
func valuationOwners(auth *identityapp.Auth, origins identity.Origins) ownerSessions {
	return ownerSessions{auth: auth, origins: origins, noSession: valuationapp.ErrNoSession, notOwner: valuationapp.ErrNotOwner}
}

// newRankings returns the owner's ranking of market-data sources (D-94)
// over pool, l and vstore, with the owner's sessions checked by owners;
// nil owners serves no owner, as the connector commands need none.
func newRankings(pool *pgxpool.Pool, l *ledgerapp.Ledger, vstore *valuationpg.Store, owners valuationapp.Owners) *valuationapp.Rankings {
	return valuationapp.NewRankings(valuationapp.NewPlanner(l, ledgerpg.New(pool)), vstore, owners)
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
