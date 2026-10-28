package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	"github.com/arkadiuss/openfief/internal/kernel"
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
	portfolioconnect "github.com/arkadiuss/openfief/internal/portfolio/connectapi"
	portfoliopg "github.com/arkadiuss/openfief/internal/portfolio/postgres"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// ownerSessions checks that a session is the owner's, owner-scoped, on an
// allowed origin, for the model portfolio's use cases (D-52, author
// decision, auth split). It implements portfolioapp.Owners.
type ownerSessions struct {
	auth    *identityapp.Auth
	origins identity.Origins
}

// CheckOwner maps the identity context's answer onto the portfolio
// context's errors: an origin off the allow-list or a session without owner
// scope is portfolioapp.ErrNotOwner, an invalid session
// portfolioapp.ErrNoSession.
func (o ownerSessions) CheckOwner(ctx context.Context, origin, sessionToken string) error {
	allowed, err := o.origins.Allowed(origin)
	if err != nil {
		return fmt.Errorf("%w: %w", portfolioapp.ErrNotOwner, err)
	}
	_, err = o.auth.Session(ctx, allowed, sessionToken, identity.SessionScopeOwner)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, identityapp.ErrInvalidSession):
		return fmt.Errorf("%w: %w", portfolioapp.ErrNoSession, err)
	case errors.Is(err, identityapp.ErrForbidden):
		return fmt.Errorf("%w: %w", portfolioapp.ErrNotOwner, err)
	}
	return fmt.Errorf("checking a session: %w", err)
}

// newModels returns the model portfolio's use cases over pool and vstore,
// with the owner's sessions checked by auth on origins, in the reporting
// currency reporting.
func newModels(pool *pgxpool.Pool, vstore *valuationpg.Store, auth *identityapp.Auth, origins identity.Origins, reporting kernel.Currency) (*portfolioapp.Models, error) {
	models, err := portfolioapp.New(portfoliopg.New(pool), vstore, ownerSessions{auth: auth, origins: origins}, reporting)
	if err != nil {
		return nil, fmt.Errorf("model portfolio: %w", err)
	}
	return models, nil
}

// withModelService returns h with the ModelPortfolioService mounted next to
// it (D-52): its calls go to the model portfolio's handler, admitted like
// the AuthService's, view plugins only, every other path to h. Nothing may
// store its responses.
func withModelService(h http.Handler, models *portfolioapp.Models, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := portfolioconnect.New(models, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		handler.ServeHTTP(w, r)
	}))
	mux.Handle("/", h)
	return mux
}
