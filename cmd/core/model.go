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
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
	portfolioconnect "github.com/arkadiuss/openfief/internal/portfolio/connectapi"
	portfoliopg "github.com/arkadiuss/openfief/internal/portfolio/postgres"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// ownerSessions checks that a session is the owner's, owner-scoped, on an
// allowed origin, for the owner's use cases: the model portfolio's (D-52)
// and the manual bond lots' (D-9, D-49) (author decision, auth split), and
// asks the authorizer which accounts it may change (D-163). It implements
// each context's Owners port, answering with the errors of the context it
// was made for.
type ownerSessions struct {
	auth    sessions
	origins identity.Origins
	// noSession and notOwner are the context's errors for an invalid
	// session and for a caller who is not the owner.
	noSession, notOwner error
	// authz answers which accounts the owner may change; without it, none.
	authz identityapp.Authorizer
}

// sessions checks a session token; *identityapp.Auth implements it.
type sessions interface {
	Session(ctx context.Context, origin identity.Origin, token string, want identity.SessionScope) (identity.Session, error)
}

// portfolioOwners is ownerSessions answering with the portfolio context's
// errors.
func portfolioOwners(auth *identityapp.Auth, origins identity.Origins) *ownerSessions {
	return &ownerSessions{auth: auth, origins: origins, noSession: portfolioapp.ErrNoSession, notOwner: portfolioapp.ErrNotOwner}
}

// ingestOwners is ownerSessions answering with the ingest context's
// errors, asking authz which accounts the owner may change.
func ingestOwners(auth *identityapp.Auth, origins identity.Origins, authz identityapp.Authorizer) *ownerSessions {
	return &ownerSessions{auth: auth, origins: origins, noSession: ingestapp.ErrNoSession, notOwner: ingestapp.ErrNotOwner, authz: authz}
}

// CheckOwner maps the identity context's answer onto the context's errors:
// an origin off the allow-list or a session without owner scope is
// notOwner, an invalid session noSession.
func (o *ownerSessions) CheckOwner(ctx context.Context, origin, sessionToken string) error {
	_, err := o.session(ctx, origin, sessionToken)
	return err
}

// Changeable returns those of accounts the owner, by the owner-scoped
// session token valid on origin, may change, as the authorizer answers
// (D-163), in their order; it fails as CheckOwner does, and when the
// authorizer can't tell.
func (o *ownerSessions) Changeable(ctx context.Context, origin, sessionToken string, accounts []ledger.AccountID) ([]ledger.AccountID, error) {
	s, err := o.session(ctx, origin, sessionToken)
	if err != nil {
		return nil, err
	}
	out := make([]ledger.AccountID, 0, len(accounts))
	if o.authz == nil {
		return out, nil
	}
	caller := identity.Caller{Holder: identity.HolderOwner, ID: string(s.ID)}
	for _, a := range accounts {
		ok, err := o.authz.MayAct(ctx, caller, identity.ActionChange, string(a))
		if err != nil {
			return nil, fmt.Errorf("authorizing a change of %s: %w", a, err)
		}
		if ok {
			out = append(out, a)
		}
	}
	return out, nil
}

// session returns the owner-scoped session of sessionToken on origin, with
// the identity context's answer mapped as CheckOwner says.
func (o *ownerSessions) session(ctx context.Context, origin, sessionToken string) (identity.Session, error) {
	allowed, err := o.origins.Allowed(origin)
	if err != nil {
		return identity.Session{}, fmt.Errorf("%w: %w", o.notOwner, err)
	}
	s, err := o.auth.Session(ctx, allowed, sessionToken, identity.SessionScopeOwner)
	switch {
	case err == nil:
		return s, nil
	case errors.Is(err, identityapp.ErrInvalidSession):
		return identity.Session{}, fmt.Errorf("%w: %w", o.noSession, err)
	case errors.Is(err, identityapp.ErrForbidden):
		return identity.Session{}, fmt.Errorf("%w: %w", o.notOwner, err)
	}
	return identity.Session{}, fmt.Errorf("checking a session: %w", err)
}

// newModels returns the model portfolio's use cases over pool and vstore,
// with the owner's sessions checked by auth on origins, in the reporting
// currency reporting.
func newModels(pool *pgxpool.Pool, vstore *valuationpg.Store, auth *identityapp.Auth, origins identity.Origins, reporting kernel.Currency) (*portfolioapp.Models, error) {
	models, err := portfolioapp.New(portfoliopg.New(pool), vstore, portfolioOwners(auth, origins), reporting)
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
