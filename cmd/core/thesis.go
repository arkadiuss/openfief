package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	thesisapp "github.com/arkadiuss/openfief/internal/thesis/app"
	thesisconnect "github.com/arkadiuss/openfief/internal/thesis/connectapi"
	thesispg "github.com/arkadiuss/openfief/internal/thesis/postgres"
)

// thesisAudit is the draft theses' Auditor over the identity context's audit
// log (D-211, D-229): a save is an entry of the action thesis_saved and an
// edit one of thesis_edited, each naming the thesis and its version, never its
// text, enqueued in the outbox inside the transaction that stores the version,
// so a version the log can't take is not stored.
type thesisAudit struct {
	out entryEnqueuer
}

var _ thesisapp.Auditor = thesisAudit{}

// RecordChange enqueues the entry of a save or an edit of a draft thesis.
func (a thesisAudit) RecordChange(ctx context.Context, c *thesisapp.Change) error {
	action := identity.ActionThesisSaved
	if c.Edited {
		action = identity.ActionThesisEdited
	}
	e, err := identity.NewAuditEntry(c.At, identity.Actor{Kind: identity.ActorOwner, ID: c.Session}, action, identity.OutcomeOK,
		identity.Ref{Kind: identity.RefThesis, ID: c.Thesis}, identity.Ref{Kind: identity.RefThesisVersion, ID: strconv.Itoa(c.Version)})
	if err != nil {
		return fmt.Errorf("building the audit entry of a draft thesis: %w", err)
	}
	if err := a.out.Enqueue(ctx, &e); err != nil {
		return fmt.Errorf("enqueueing the audit entry of a draft thesis: %w", err)
	}
	return nil
}

// newTheses wires the thesis context over pool: the drafts in the thesis
// schema, owner-only (D-229), every save and edit audited in its transaction.
func newTheses(pool *pgxpool.Pool, out *outbox, auth *identityapp.Auth, origins identity.Origins, now func() time.Time) *thesisapp.Theses {
	owners := &ownerSessions{auth: auth, origins: origins, noSession: thesisapp.ErrNoSession, notOwner: thesisapp.ErrNotOwner}
	return thesisapp.New(thesispg.New(joined(pool)), owners, thesisAudit{out: out}, newCallID, now)
}

// withThesisService returns h with the ThesisService mounted next to it
// (D-229): its calls go to the draft theses' handler, admitted like the
// ModelSettingsService's, view plugins only, every other path to h. It is on
// the plugin port and not on the read API, so no script token or demo session
// reaches it. Nothing may store its responses.
func withThesisService(h http.Handler, theses *thesisapp.Theses, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := thesisconnect.New(theses, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, noStore(handler))
	mux.Handle("/", h)
	return mux
}
