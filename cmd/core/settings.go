package main

import (
	"log/slog"
	"net/http"

	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
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
// defaults and the owner's sessions checked by owners.
func newSettings(recorder *valuationapp.InputRecorder, vstore *valuationpg.Store, owners valuationapp.Owners, cfg *serveConfig) *valuationapp.Settings {
	return valuationapp.NewSettings(recorder, vstore, owners, jurisdictionPack(), cfg.Reporting)
}

// withSettingsService returns h with the SettingsService mounted next to
// it (D-79): its calls go to the settings' handler, admitted like the
// AuthService's, view plugins only, every other path to h. Nothing may
// store its responses.
func withSettingsService(h http.Handler, settings *valuationapp.Settings, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := valuationconnect.NewSettings(settings, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, noStore(handler))
	mux.Handle("/", h)
	return mux
}
