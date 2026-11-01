package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	attentionapp "github.com/arkadiuss/openfief/internal/attention/app"
	attentionconnect "github.com/arkadiuss/openfief/internal/attention/connectapi"
	attentionpg "github.com/arkadiuss/openfief/internal/attention/postgres"
	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// attentionEvery is how often the core runs the checks that raise
// attention items and publishes the outbox's events (D-130).
const attentionEvery = time.Minute

// attention is the wired attention context: the items with every check
// that needs the owner, and the relay that publishes the outbox's events.
type attention struct {
	items *attentionapp.Items
	relay *attentionapp.Relay
}

// attentionOwners is ownerSessions answering with the attention context's
// errors.
func attentionOwners(auth *identityapp.Auth, origins identity.Origins) ownerSessions {
	return ownerSessions{auth: auth, origins: origins, noSession: attentionapp.ErrNoSession, notOwner: attentionapp.ErrNotOwner}
}

// newAttention wires the attention items over pool (D-130, D-131) with the
// core's checks: flagged balance assertions and price checks, import
// problems, unmapped broker accounts, unconfirmed tickers, the missing
// marginal rate and broken transfers. Its events go to a log publisher.
func newAttention(pool *pgxpool.Pool, l *ledgerapp.Ledger, vstore *valuationpg.Store, checks *valuationapp.CrossChecks,
	owners attentionapp.Owners, now func() time.Time, log *slog.Logger,
) *attention {
	store := attentionpg.New(pool)
	imports := ingestpg.New(pool)
	items := attentionapp.New(store, owners, now,
		attentionapp.FlaggedAssertions(l),
		attentionapp.FlaggedPriceChecks(checks),
		attentionapp.ImportProblems(imports),
		attentionapp.UnmappedBrokerAccounts(imports),
		attentionapp.UnconfirmedTickers(imports),
		attentionapp.MissingSettings(l, vstore, jurisdictionPack()),
		attentionapp.BrokenTransfers(l),
	)
	return &attention{items: items, relay: &attentionapp.Relay{Outbox: store, Publisher: attentionapp.LogPublisher{Log: log}, Now: now}}
}

// tick runs every check once and publishes what the outbox holds; a
// failure is logged and the next tick tries again.
func (a *attention) tick(ctx context.Context, log *slog.Logger) {
	if err := a.items.Sweep(ctx); err != nil {
		log.Warn("an attention check failed", "error", err)
	}
	if _, err := a.relay.Publish(ctx); err != nil {
		log.Warn("publishing attention events failed", "error", err)
	}
}

// run ticks at once and then every every, until ctx ends.
func (a *attention) run(ctx context.Context, every time.Duration, log *slog.Logger) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		a.tick(ctx, log)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// withAttentionService returns h with the AttentionService mounted next to
// it (D-130, D-79): its calls go to the attention items, admitted like the
// AuthService's, view plugins only, every other path to h. Nothing may
// store its responses.
func withAttentionService(h http.Handler, items *attentionapp.Items, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := attentionconnect.NewAttention(items, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, noStore(handler))
	mux.Handle("/", h)
	return mux
}
