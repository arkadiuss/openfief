package main

import (
	"log/slog"
	"net/http"

	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationconnect "github.com/arkadiuss/openfief/internal/valuation/connectapi"
)

// withPriceCheckService returns h with the PriceCheckService mounted next
// to it (D-110, D-112, D-79): its calls go to the cross-check use cases,
// admitted like the AuthService's, view plugins only, every other path to
// h. Nothing may store its responses.
func withPriceCheckService(h http.Handler, checks *valuationapp.CrossChecks, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := valuationconnect.NewPriceChecks(checks, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		handler.ServeHTTP(w, r)
	}))
	mux.Handle("/", h)
	return mux
}
