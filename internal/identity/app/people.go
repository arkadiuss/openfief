package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
)

// InviteViewer issues, for an owner session on the allowed origin text with
// a passkey check in the last five minutes (D-183), a one-time link that
// invites a viewer (D-65, D-210): the viewer is called name, and starts with
// the grants for targets, which must exist. Whoever opens the link enrolls
// their own passkey, which creates the person, binds the passkey, makes the
// grants and signs them in with a read session; the link is then spent. It
// lasts ttl, or identity.DefaultEnrollmentTTL when ttl is zero, and works on
// target, another allowed origin, or on origin itself when target is empty.
func (a *Auth) InviteViewer(ctx context.Context, origin, token, name string, targets []identity.GrantTarget, target string, ttl time.Duration) (EnrollmentLink, error) {
	o, s, err := a.recentOwner(ctx, origin, token)
	if err != nil {
		return EnrollmentLink{}, err
	}
	invite := identity.ViewerInvite{Name: name, Targets: targets}
	if validErr := invite.Validate(); validErr != nil {
		return EnrollmentLink{}, fmt.Errorf("%w: %w", ErrInvalidRequest, validErr)
	}
	switch {
	case len(targets) > 0 && a.grants == nil:
		return EnrollmentLink{}, errNoGrants
	case len(targets) > 0:
		if invite.Targets, err = a.grants.checkTargets(ctx, targets); err != nil {
			return EnrollmentLink{}, err
		}
	}
	t := o
	if target != "" {
		if t, err = a.origin(target); err != nil {
			return EnrollmentLink{}, err
		}
	}
	return issueLink(ctx, a.store, a.random, a.now(), a.cfg.Origins, t, ttl, ownerActor(&s), &invite)
}

// PersonAccess is a person with what they may view and how they sign in
// (D-210).
type PersonAccess struct {
	Person identity.Person
	// Grants are the grants in force; none for the owner, who holds every
	// account without one.
	Grants []identity.Grant
	// Passkeys is how many of the person's passkeys work.
	Passkeys int
	// ActiveSessions is how many of their sessions are still valid.
	ActiveSessions int
}

// People returns the owner and every viewer, revoked ones included, for an
// owner session: each with the grants in force, the passkeys that work and
// the sessions still valid. The owner comes first.
func (a *Auth) People(ctx context.Context, origin, token string) ([]PersonAccess, error) {
	if _, _, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner); err != nil {
		return nil, err
	}
	people, err := a.store.People(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing people: %w", err)
	}
	keys, err := a.store.PasskeyCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting passkeys: %w", err)
	}
	sessions, err := a.store.ActiveSessionCounts(ctx, a.now().UTC())
	if err != nil {
		return nil, fmt.Errorf("counting sessions: %w", err)
	}
	var grants []identity.Grant
	if a.grants != nil {
		if grants, err = a.grants.List(ctx, identity.GrantMatch{}, true); err != nil {
			return nil, err
		}
	}
	out := make([]PersonAccess, 0, len(people))
	for i := range people {
		p := people[i]
		p.Handle = nil
		pa := PersonAccess{Person: p, Passkeys: keys[p.ID], ActiveSessions: sessions[p.ID]}
		if p.IsViewer() {
			pa.Grants = grantsOf(grants, p.GrantHolder())
		}
		out = append(out, pa)
	}
	return out, nil
}

// Revocation is what revoking a viewer ended.
type Revocation struct {
	Sessions, Passkeys, Grants int
}

// RevokePerson revokes the viewer id for an owner session (D-210): their
// grants are removed from the authorization model first, so a failure there
// changes nothing; then, in one transaction, the person is revoked, their
// sessions end, their passkeys are disabled and their grants end, with an
// audit entry for the person, each passkey and each grant (D-211). The person
// stays, with their ID. It needs no passkey check: it only takes access away.
// An unknown or revoked viewer fails with an error wrapping ErrNotFound; the
// owner, with ErrInvalidRequest.
func (a *Auth) RevokePerson(ctx context.Context, origin, token, id string) (Revocation, error) {
	_, s, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner)
	if err != nil {
		return Revocation{}, err
	}
	person, err := a.activeViewer(ctx, id)
	if err != nil {
		return Revocation{}, err
	}
	match := identity.GrantMatch{Holder: person.GrantHolder()}
	if a.grants != nil {
		inForce, listErr := a.grants.store.Grants(ctx, match, true)
		if listErr != nil {
			return Revocation{}, fmt.Errorf("loading the grants to revoke: %w", listErr)
		}
		if unprojectErr := a.grants.unproject(ctx, inForce); unprojectErr != nil {
			return Revocation{}, unprojectErr
		}
	}
	var out Revocation
	err = within(ctx, a.store, func(ctx context.Context) error {
		var revokeErr error
		out, revokeErr = a.revokeStored(ctx, ownerActor(&s), &person, match)
		return revokeErr
	})
	if err != nil {
		return Revocation{}, fmt.Errorf("revoking a viewer: %w", err)
	}
	return out, nil
}

// activeViewer returns the viewer id, or an error wrapping ErrInvalidRequest
// for the owner or an ID of no shape a viewer has, and ErrNotFound for one
// that is unknown or revoked already.
func (a *Auth) activeViewer(ctx context.Context, id string) (identity.Person, error) {
	holder := identity.GrantHolder{Kind: identity.GrantHolderPerson, ID: id}
	if id == identity.OwnerPersonID {
		return identity.Person{}, fmt.Errorf("%w: the owner can't be revoked", ErrInvalidRequest)
	}
	if err := holder.Validate(); err != nil {
		return identity.Person{}, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	person, err := a.store.Person(ctx, id)
	if errors.Is(err, ErrUnknownPerson) || (err == nil && !person.Active()) {
		return identity.Person{}, fmt.Errorf("%w: no such viewer", ErrNotFound)
	}
	if err != nil {
		return identity.Person{}, fmt.Errorf("loading the viewer: %w", err)
	}
	return person, nil
}

// revokeStored revokes person in the store, inside the transaction ctx
// carries, and records it: the person, each passkey disabled and each grant
// ended, by actor.
func (a *Auth) revokeStored(ctx context.Context, actor identity.Actor, person *identity.Person, match identity.GrantMatch) (Revocation, error) {
	keys, sessions, err := a.store.RevokePerson(ctx, person.ID, a.now().UTC())
	if errors.Is(err, ErrUnknownPerson) {
		return Revocation{}, fmt.Errorf("%w: no such viewer", ErrNotFound)
	}
	if err != nil {
		return Revocation{}, fmt.Errorf("revoking the viewer: %w", err)
	}
	out := Revocation{Sessions: sessions, Passkeys: len(keys)}
	if recordErr := a.record(ctx, actor, identity.ActionPersonRevoked, identity.OutcomeOK, person.Ref()); recordErr != nil {
		return Revocation{}, recordErr
	}
	for i := range keys {
		passkey := identity.Ref{Kind: identity.RefPasskey, ID: keys[i].ID()}
		if recordErr := a.record(ctx, actor, identity.ActionPasskeyRemoved, identity.OutcomeOK, passkey, person.Ref()); recordErr != nil {
			return Revocation{}, recordErr
		}
	}
	if a.grants == nil {
		return out, nil
	}
	ended, err := a.grants.revokeStored(ctx, actor, match)
	out.Grants = len(ended)
	return out, err
}

// storeViewerRegistration creates the viewer c's registration was begun for,
// with the passkey p, and their grants, uses up the enrollment code and signs
// the viewer in with a read session, all in one transaction with the audit
// entries that record them (D-211): the link's use, the person's creation,
// the passkey's, each grant and the sign-in. The grants are then passed on to
// the authorization model; a failure there is Registered.NotProjected.
func (a *Auth) storeViewerRegistration(ctx context.Context, c *identity.Ceremony, person *identity.Person, invite *identity.ViewerInvite, p *identity.Passkey) (Registered, error) {
	targets := invite.Targets
	if len(targets) > 0 {
		if a.grants == nil {
			return Registered{}, errNoGrants
		}
		var err error
		if targets, err = a.grants.stillThere(ctx, targets); err != nil {
			return Registered{}, err
		}
	}
	var (
		out  Registered
		made []identity.Grant
	)
	err := within(ctx, a.store, func(ctx context.Context) error {
		actor := identity.Actor{Kind: identity.ActorEnrollmentCode, ID: string(c.EnrollmentCode)}
		if err := a.enrollViewerRows(ctx, c, actor, person, p); err != nil {
			return err
		}
		var grantErr error
		if made, grantErr = a.grantViewer(ctx, actor, person, targets); grantErr != nil {
			return grantErr
		}
		// Creating the passkey verified the viewer on it, so it signs the session in.
		s, err := a.issueSession(ctx, &Authenticated{
			Person: person.ID, Scope: identity.SessionScopeRead, Authenticator: identity.AuthenticatorPasskey, Passkey: p.CredentialID,
		}, c.Origin)
		if err != nil {
			return err
		}
		out = Registered{Passkey: *p, Session: &s}
		return nil
	})
	if err != nil {
		return Registered{}, err
	}
	if len(made) > 0 {
		out.NotProjected = a.grants.project(ctx, made)
	}
	return out, nil
}

// enrollViewerRows uses up c's enrollment code and stores the viewer person
// and their passkey p, with the entries of the link's use, the person's
// creation and the passkey's, inside the transaction ctx carries.
func (a *Auth) enrollViewerRows(ctx context.Context, c *identity.Ceremony, actor identity.Actor, person *identity.Person, p *identity.Passkey) error {
	useErr := a.store.UseEnrollmentCode(ctx, c.EnrollmentCode, p.CreatedAt)
	if errors.Is(useErr, identity.ErrEnrollmentCodeUsed) {
		return fmt.Errorf("%w: %w", ErrForbidden, useErr)
	}
	if useErr != nil {
		return fmt.Errorf("using an enrollment code: %w", useErr)
	}
	if err := a.store.SavePerson(ctx, person); err != nil {
		return fmt.Errorf("storing the viewer: %w", err)
	}
	if err := a.store.SavePasskey(ctx, p); err != nil {
		return fmt.Errorf("storing a passkey: %w", err)
	}
	code := identity.Ref{Kind: identity.RefEnrollmentCode, ID: string(c.EnrollmentCode)}
	passkey := identity.Ref{Kind: identity.RefPasskey, ID: p.ID()}
	for _, e := range []struct {
		action identity.AuditAction
		refs   []identity.Ref
	}{
		{identity.ActionEnrollmentLinkUsed, []identity.Ref{code, passkey, person.Ref()}},
		{identity.ActionPersonCreated, []identity.Ref{person.Ref(), code}},
		{identity.ActionPasskeyAdded, []identity.Ref{passkey, person.Ref()}},
	} {
		if err := a.record(ctx, actor, e.action, identity.OutcomeOK, e.refs...); err != nil {
			return err
		}
	}
	return nil
}

// grantViewer makes the grant of each of targets to person, with an entry by
// actor, inside the transaction ctx carries, and returns those it made.
func (a *Auth) grantViewer(ctx context.Context, actor identity.Actor, person *identity.Person, targets []identity.GrantTarget) ([]identity.Grant, error) {
	var made []identity.Grant
	for _, t := range targets {
		grant, err := identity.NewGrant(person.GrantHolder(), t, a.now().UTC().Truncate(time.Microsecond))
		if err != nil {
			return nil, fmt.Errorf("building a grant: %w", err)
		}
		created, err := a.grants.saveGrant(ctx, actor, &grant)
		if err != nil {
			return nil, err
		}
		if created {
			made = append(made, grant)
		}
	}
	return made, nil
}

// stillThere returns those of targets that exist now: a sleeve or an account
// deleted since the owner invited the viewer is not granted.
func (g *Grants) stillThere(ctx context.Context, targets []identity.GrantTarget) ([]identity.GrantTarget, error) {
	out := make([]identity.GrantTarget, 0, len(targets))
	for _, t := range targets {
		ok, err := g.dir.TargetExists(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("looking for the %s: %w", t.Kind, err)
		}
		if ok {
			out = append(out, t)
		}
	}
	return out, nil
}
