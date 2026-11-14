package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// Auditor keeps the append-only audit log (D-211). *identity/postgres.Store
// implements it.
type Auditor interface {
	// AppendAudit adds e to the log, inside the transaction ctx carries (see
	// Atomic) or, without one, on its own, and sets e.Seq. It fails without
	// writing when e is not valid.
	AppendAudit(ctx context.Context, e *identity.AuditEntry) error
	// AuditEntries returns the page f selects, newest first.
	AuditEntries(ctx context.Context, f *identity.AuditFilter) (AuditPage, error)
}

// Atomic runs a function in one transaction of the core's database, so a
// change of state and its audit entry commit together or not at all: the
// log fails closed (D-211). Every store method called with the ctx fn gets
// joins the transaction; a nested call joins the outer one.
type Atomic interface {
	Atomically(ctx context.Context, fn func(ctx context.Context) error) error
}

// AuditPage is one page of audit entries, newest first.
type AuditPage struct {
	Entries []identity.AuditEntry
	// Next is the cursor for the page after this one (AuditFilter.Before);
	// zero when this is the last.
	Next int64
	// Persons maps the ID of a viewer's session that is an actor on this
	// page to the person it is of, so the log can name them; a session that
	// is no longer known is left out. Only an ID, never a name.
	Persons map[identity.SessionID]string
}

// record appends the entry for action by actor with outcome, naming refs, as
// of now. Called inside Atomic.Atomically, it commits or rolls back with the
// change it records.
func (a *Auth) record(ctx context.Context, actor identity.Actor, action identity.AuditAction, outcome identity.AuditOutcome, refs ...identity.Ref) error {
	e, err := identity.NewAuditEntry(a.now(), actor, action, outcome, refs...)
	if err != nil {
		return fmt.Errorf("building an audit entry: %w", err)
	}
	if err := a.store.AppendAudit(ctx, &e); err != nil {
		return fmt.Errorf("appending an audit entry: %w", err)
	}
	return nil
}

// within runs fn in one transaction of store (see Atomic), wrapping its
// error.
func within(ctx context.Context, store Atomic, fn func(ctx context.Context) error) error {
	if err := store.Atomically(ctx, fn); err != nil {
		return fmt.Errorf("in one transaction: %w", err)
	}
	return nil
}

// ownerActor is whoever session s is of: the owner, or a viewer.
func ownerActor(s *identity.Session) identity.Actor {
	if s.OfViewer() {
		return identity.Actor{Kind: identity.ActorViewer, ID: string(s.ID)}
	}
	return identity.Actor{Kind: identity.ActorOwner, ID: string(s.ID)}
}

// refused returns err, after writing the refused sign-in it stands for to the
// audit log, when it is a refusal: a credential, a ceremony or an origin the
// core would not accept. Nobody is signed in, so there is no change of state
// for the entry to share a transaction with: the entry is written on its own,
// and the sign-in is refused whether or not the write succeeds. A failed
// write is joined to the refusal, so the caller can log it, and never turns
// a refusal into a sign-in.
func (a *Auth) refused(ctx context.Context, err error) error {
	if !errors.Is(err, ErrAuthenticationFailed) && !errors.Is(err, ErrForbidden) && !errors.Is(err, ErrCeremony) {
		return err
	}
	if auditErr := a.record(ctx, identity.Actor{Kind: identity.ActorAnonymous}, identity.ActionSignIn, identity.OutcomeRefused); auditErr != nil {
		return fmt.Errorf("%w (the refusal could not be audited: %w)", err, auditErr)
	}
	return err
}

// AuditEntries returns the page of the audit log f selects, newest first,
// for an owner session (D-211: Settings shows it to the owner, filterable).
// It is not on the read API, so no script token or view plugin reads it
// (D-137), and reading it is itself no audit entry. An invalid filter fails
// with an error wrapping ErrInvalidRequest.
func (a *Auth) AuditEntries(ctx context.Context, origin, token string, f *identity.AuditFilter) (AuditPage, error) {
	if _, _, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner); err != nil {
		return AuditPage{}, err
	}
	if err := f.Validate(); err != nil {
		return AuditPage{}, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	page, err := a.store.AuditEntries(ctx, f)
	if err != nil {
		return AuditPage{}, fmt.Errorf("listing audit entries: %w", err)
	}
	page.Persons = map[identity.SessionID]string{}
	for i := range page.Entries {
		actor := page.Entries[i].Actor
		id := identity.SessionID(actor.ID)
		if _, known := page.Persons[id]; actor.Kind != identity.ActorViewer || known {
			continue
		}
		s, _, sessionErr := a.store.Session(ctx, id)
		switch {
		case errors.Is(sessionErr, ErrUnknownSession):
		case sessionErr != nil:
			return AuditPage{}, fmt.Errorf("naming the viewer of a session: %w", sessionErr)
		default:
			page.Persons[id] = s.Person()
		}
	}
	return page, nil
}

// RecordOwnerChange has change, a change of state another context makes for
// the owner of session, kept together with its audit entry (D-211): the entry
// of action naming refs is appended first, in a transaction, then change runs,
// and the entry is kept only if change returns nil. So a change the log can't
// take an entry for is not made, and a change that fails leaves no entry.
// change runs with the caller's context, on the connections of its own store;
// the transaction opened here holds the entry alone, so a commit that fails
// after change succeeded leaves a change with no entry, and returns an error.
//
// A failed append is an error wrapping kernel.ErrAuditUnavailable, and an
// entry the domain refuses (an unknown action, a reference of the wrong shape)
// one wrapping identity.ErrInvalidAuditEntry, either way with change not
// called. An error of change is returned as it is, wrapped.
func (a *Auth) RecordOwnerChange(ctx context.Context, session identity.SessionID, action identity.AuditAction, refs []identity.Ref,
	change func(context.Context) error,
) error {
	e, err := identity.NewAuditEntry(a.now(), identity.Actor{Kind: identity.ActorOwner, ID: string(session)}, action, identity.OutcomeOK, refs...)
	if err != nil {
		return fmt.Errorf("building the audit entry of %s: %w", action, err)
	}
	err = a.store.Atomically(ctx, func(tx context.Context) error {
		if appendErr := a.store.AppendAudit(tx, &e); appendErr != nil {
			return fmt.Errorf("%w: %w", kernel.ErrAuditUnavailable, appendErr)
		}
		return change(ctx)
	})
	if err != nil {
		return fmt.Errorf("recording the change %s: %w", action, err)
	}
	return nil
}
