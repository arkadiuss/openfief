package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	ingestconnect "github.com/arkadiuss/openfief/internal/ingest/connectapi"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
	portfolioconnect "github.com/arkadiuss/openfief/internal/portfolio/connectapi"
)

// newSleeves returns the owner's sleeve entries (D-71, D-79) over the
// portfolio registry r, which keeps sleeves and kinds, with the owner's
// sessions checked by auth on origins.
func newSleeves(r *portfolioapp.Registry, auth *identityapp.Auth, origins identity.Origins) *portfolioapp.Sleeves {
	owners := portfolioOwners(auth, origins)
	return &portfolioapp.Sleeves{Owners: owners, Store: r, Deletions: r, Audit: owners}
}

// newAccounts returns the owner's account entries (D-72, D-79) over d: the
// ledger's accounts, the ingest schema's connector bindings and broker
// accounts, and terms, the accounts' tax wrappers and fee rules (D-76,
// D-78) in the Polish pack, which it also forgets with a deleted account
// (D-120).
func newAccounts(d *connectorDeps, owners ingestapp.Owners, terms accountTerms) *ingestapp.Accounts {
	store := ingestpg.New(d.pool)
	accounts := &ingestapp.Accounts{
		Owners: owners, Ledger: d.ledger, Bindings: store, Brokers: store, Terms: terms, Pack: jurisdictionPack(), Now: d.now,
		Deletions: d.ledger, Connections: store, Forget: terms, Payouts: ledgerImporter{ledger: d.ledger},
		Sleeves: d.sleeves, PayoutSettings: store,
	}
	// The owners the core builds enter each change in the audit log; the tools that build the use cases
	// without a core pass owners that can't, and enter none.
	if a, ok := owners.(ingestapp.ChangeAuditor); ok {
		accounts.Audit = a
	}
	return accounts
}

// withOnboardingServices returns h with the SleeveService and the
// AccountService mounted next to it (D-71, D-72, D-73), admitted like the
// AuthService's, view plugins only; every other path goes to h. Nothing may
// store their responses. While a demo deployment's data is loading, they
// refuse every change (refuseWhileDemoLoads).
func withOnboardingServices(h http.Handler, sleeves *portfolioapp.Sleeves, accounts *ingestapp.Accounts,
	demo demoLoads, callers identityconnect.Callers, log *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()
	sleevePath, sleeveHandler := portfolioconnect.NewSleeves(sleeves, identityconnect.NewCallerInterceptor(callers, log), log)
	accountPath, accountHandler := ingestconnect.NewAccounts(accounts, identityconnect.NewCallerInterceptor(callers, log), log)
	mux.Handle(sleevePath, noStore(refuseWhileDemoLoads(sleeveHandler, demo, log)))
	mux.Handle(accountPath, noStore(refuseWhileDemoLoads(accountHandler, demo, log)))
	mux.Handle("/", h)
	return mux
}

// errDemoLoading is the core's reason for refusing a change while a demo
// deployment's data is loading.
var errDemoLoading = errors.New("the demo data is still loading, and a demo is not set up by hand")

// errDemoMarker is the core's answer when it can't read the demo marker.
var errDemoMarker = errors.New("reading the demo marker failed")

// demoLoads reports whether a demo deployment's data is still loading;
// *identityapp.Auth implements it.
type demoLoads interface {
	DemoLoading(ctx context.Context) (bool, error)
}

// onboardingReads are the procedures of the onboarding services that
// change nothing, which refuseWhileDemoLoads lets through.
func onboardingReads() map[string]bool {
	return map[string]bool{ingestv1connect.AccountServiceListBrokerAccountsProcedure: true}
}

// refuseWhileDemoLoads serves the onboarding services' handler h, but
// while demo says a demo deployment's data is loading it refuses every
// procedure that changes something with FailedPrecondition: the first-run
// wizard's sleeves and accounts would be data without the demo marker, and
// the demo loader would then refuse the database as a real one. When the
// demo marker can't be read, it refuses as internal.
func refuseWhileDemoLoads(h http.Handler, demo demoLoads, log *slog.Logger) http.Handler {
	reads := onboardingReads()
	errs := connect.NewErrorWriter()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !reads[r.URL.Path] {
			loading, err := demo.DemoLoading(r.Context())
			if err != nil {
				log.ErrorContext(r.Context(), "reading the demo marker failed", "path", r.URL.Path)
				writeConnectError(w, r, errs, connect.NewError(connect.CodeInternal, errDemoMarker), log)
				return
			}
			if loading {
				writeConnectError(w, r, errs, connect.NewError(connect.CodeFailedPrecondition, errDemoLoading), log)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

// writeConnectError answers r with err in r's protocol, logging a failed
// write.
func writeConnectError(w http.ResponseWriter, r *http.Request, errs *connect.ErrorWriter, err *connect.Error, log *slog.Logger) {
	if writeErr := errs.Write(w, r, err); writeErr != nil {
		log.DebugContext(r.Context(), "writing a refusal failed", "path", r.URL.Path)
	}
}

// noStore serves h with responses no cache may keep or sniff.
func noStore(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		h.ServeHTTP(w, r)
	})
}
