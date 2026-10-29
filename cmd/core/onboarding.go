package main

import (
	"log/slog"
	"net/http"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	ingestconnect "github.com/arkadiuss/openfief/internal/ingest/connectapi"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
	portfolioconnect "github.com/arkadiuss/openfief/internal/portfolio/connectapi"
)

// newSleeves returns the owner's sleeve entries (D-71, D-79) over the
// ledger, which keeps sleeves and kinds, with the owner's sessions checked
// by auth on origins.
func newSleeves(l *ledgerapp.Ledger, auth *identityapp.Auth, origins identity.Origins) *portfolioapp.Sleeves {
	return &portfolioapp.Sleeves{Owners: portfolioOwners(auth, origins), Store: l}
}

// newAccounts returns the owner's account entries (D-72, D-79) over d: the
// ledger's accounts, and the ingest schema's connector bindings and broker
// accounts.
func newAccounts(d *connectorDeps, owners ingestapp.Owners) *ingestapp.Accounts {
	store := ingestpg.New(d.pool)
	return &ingestapp.Accounts{Owners: owners, Ledger: d.ledger, Bindings: store, Brokers: store, Now: d.now}
}

// withOnboardingServices returns h with the SleeveService and the
// AccountService mounted next to it (D-71, D-72, D-73), admitted like the
// AuthService's, view plugins only; every other path goes to h. Nothing may
// store their responses.
func withOnboardingServices(h http.Handler, sleeves *portfolioapp.Sleeves, accounts *ingestapp.Accounts,
	callers identityconnect.Callers, log *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()
	sleevePath, sleeveHandler := portfolioconnect.NewSleeves(sleeves, identityconnect.NewCallerInterceptor(callers, log), log)
	accountPath, accountHandler := ingestconnect.NewAccounts(accounts, identityconnect.NewCallerInterceptor(callers, log), log)
	mux.Handle(sleevePath, noStore(sleeveHandler))
	mux.Handle(accountPath, noStore(accountHandler))
	mux.Handle("/", h)
	return mux
}

// noStore serves h with responses no cache may keep or sniff.
func noStore(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		h.ServeHTTP(w, r)
	})
}
