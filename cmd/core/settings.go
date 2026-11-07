package main

import (
	"log/slog"
	"net/http"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationconnect "github.com/arkadiuss/openfief/internal/valuation/connectapi"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// accountTerms joins valuation's input recorder, which records an
// account's tax wrapper and fee rule and revalues it (D-76, D-78), and its
// store, which reads them back, into the account entries' AccountTerms.
type accountTerms struct {
	*valuationapp.InputRecorder
	*valuationpg.Store
}

// newSettings returns the owner's settings use cases (D-78, D-113) over
// recorder and vstore, with the reporting currency, the Polish pack's
// defaults, the bonds' early-redemption fees in vstore, which the owner
// confirms or overrides (D-97), and the owner's sessions checked by owners.
func newSettings(recorder *valuationapp.InputRecorder, vstore *valuationpg.Store, owners valuationapp.Owners, cfg *serveConfig) *valuationapp.Settings {
	return valuationapp.NewSettings(recorder, vstore, owners, jurisdictionPack(), cfg.Reporting).WithBondFees(vstore)
}

// newChartStarts returns the owner's setting of the charts' start (D-167)
// over vstore and the accounts l lists, with the owner's sessions checked
// by auth on origins and the accounts the owner may change asked of authz
// (D-163).
func newChartStarts(auth *identityapp.Auth, origins identity.Origins, authz identityapp.Authorizer, vstore *valuationpg.Store, l *ledgerapp.Ledger) *valuationapp.ChartStarts {
	owners := &ownerSessions{auth: auth, origins: origins, noSession: valuationapp.ErrNoSession, notOwner: valuationapp.ErrNotOwner, authz: authz}
	return valuationapp.NewChartStarts(owners, vstore, l)
}

// withSettingsService returns h with the SettingsService mounted next to
// it (D-79): its calls go to the settings' and the charts' start's
// handler, admitted like the AuthService's, view plugins only, every other
// path to h. Nothing may store its responses.
func withSettingsService(h http.Handler, settings *valuationapp.Settings, starts *valuationapp.ChartStarts, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := valuationconnect.NewSettings(settings, starts, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, noStore(handler))
	mux.Handle("/", h)
	return mux
}
