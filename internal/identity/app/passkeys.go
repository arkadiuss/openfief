package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
)

// Started is a WebAuthn ceremony the browser is to carry out: the ID to
// finish it with, the options for the browser, and when it expires.
type Started struct {
	CeremonyID string
	Options    []byte
	ExpiresAt  time.Time
	// InvitedName is the name the owner gave the viewer a viewer's
	// enrollment code invites; empty for any other ceremony.
	InvitedName string
}

// RegistrationGrant is what allows creating a passkey: exactly one of a
// one-time enrollment code issued for the origin, or an owner session on
// the origin's host (D-60) with a passkey check in the last five minutes
// (D-183).
type RegistrationGrant struct {
	EnrollmentCode string
	SessionToken   string
}

// BeginRegistration starts creating a passkey on the allowed origin text,
// as grant allows, creating the owner on first use.
func (a *Auth) BeginRegistration(ctx context.Context, origin string, grant RegistrationGrant) (Started, error) {
	o, err := a.origin(origin)
	if err != nil {
		return Started{}, err
	}
	code, session, invite, err := a.checkGrant(ctx, o, grant)
	if err != nil {
		return Started{}, err
	}
	owner, err := a.ensureOwner(ctx)
	if err != nil {
		return Started{}, err
	}
	person := identity.OwnerPerson(&owner)
	var viewer *identity.Person
	if invite != nil {
		v, newErr := a.newViewer(invite.Name)
		if newErr != nil {
			return Started{}, newErr
		}
		person, viewer = v, &v
	}
	keys, err := a.hostPasskeys(ctx, o)
	if err != nil {
		return Started{}, err
	}
	begun, err := a.ceremonies.BeginRegistration(o, &person, keys)
	if err != nil {
		return Started{}, fmt.Errorf("beginning a registration: %w", err)
	}
	c, err := a.saveCeremony(ctx, identity.CeremonyRegistration, o, begun.State, code, session, viewer)
	if err != nil {
		return Started{}, err
	}
	st := Started{CeremonyID: c.ID, Options: begun.Options, ExpiresAt: c.ExpiresAt}
	if viewer != nil {
		st.InvitedName = viewer.Name
	}
	return st, nil
}

// newViewer returns a viewer called name with a random ID and user handle,
// created now: the person a viewer's enrollment link will create.
func (a *Auth) newViewer(name string) (identity.Person, error) {
	id, err := a.randomHex(identity.IDBytes)
	if err != nil {
		return identity.Person{}, fmt.Errorf("creating a viewer: %w", err)
	}
	handle := make([]byte, identity.OwnerHandleBytes)
	if _, readErr := io.ReadFull(a.random, handle); readErr != nil {
		return identity.Person{}, fmt.Errorf("creating a viewer: random handle: %w", readErr)
	}
	p, err := identity.NewViewer(id, name, handle, a.nowSecond())
	if err != nil {
		return identity.Person{}, fmt.Errorf("creating a viewer: %w", err)
	}
	return p, nil
}

// Registered is a passkey just created, and, when an enrollment code
// allowed it, the owner session that comes with it.
type Registered struct {
	Passkey identity.Passkey
	// Session is set only for a registration begun with an enrollment code.
	Session *NewSession
	// NotProjected is set when a viewer's grants are stored and audited but
	// could not be passed on to the authorization model: the registration
	// stands, and the next start's reconcile, or granting again, repairs it.
	NotProjected error
}

// FinishRegistration verifies credential for the registration ceremony
// ceremonyID on the allowed origin text and stores the passkey as name
// (empty names it after the host). A registration begun with an enrollment
// code uses the code up and logs the owner in.
func (a *Auth) FinishRegistration(ctx context.Context, origin, ceremonyID string, credential []byte, name string) (Registered, error) {
	o, err := a.origin(origin)
	if err != nil {
		return Registered{}, err
	}
	name, err = passkeyName(name, o)
	if err != nil {
		return Registered{}, err
	}
	c, err := a.takeCeremony(ctx, ceremonyID, identity.CeremonyRegistration, o)
	if err != nil {
		return Registered{}, err
	}
	owner, err := a.store.Owner(ctx)
	if err != nil {
		return Registered{}, fmt.Errorf("loading the owner: %w", err)
	}
	if c.EnrollmentCode == "" && c.Session == "" {
		// Begun before the audit log named who begins a registration: begin again.
		return Registered{}, fmt.Errorf("%w: the registration names neither a code nor a session", ErrCeremony)
	}
	person, invite, err := a.registrationPerson(ctx, &c, &owner)
	if err != nil {
		return Registered{}, err
	}
	p, err := a.verifiedPasskey(o, &person, c.State, credential, name)
	if err != nil {
		return Registered{}, err
	}
	if invite != nil {
		p.PersonID = person.ID
		return a.storeViewerRegistration(ctx, &c, &person, invite, &p)
	}
	return a.storeRegistration(ctx, &c, &owner, &p)
}

// verifiedPasskey verifies credential against state for person on o and
// returns the passkey it creates, named name and made now.
func (a *Auth) verifiedPasskey(o identity.Origin, person *identity.Person, state, credential []byte, name string) (identity.Passkey, error) {
	p, err := a.ceremonies.FinishRegistration(o, person, state, credential)
	if err != nil {
		return identity.Passkey{}, fmt.Errorf("%w: %w", ErrRegistrationFailed, err)
	}
	p.Name, p.RPID, p.CreatedAt = name, o.RPID(), a.nowSecond()
	if err := p.Validate(); err != nil {
		return identity.Passkey{}, fmt.Errorf("%w: %w", ErrRegistrationFailed, err)
	}
	return p, nil
}

// passkeyName returns name, or the host of o when it is empty, checked.
func passkeyName(name string, o identity.Origin) (string, error) {
	if name == "" {
		name = o.RPID()
	}
	if err := identity.CheckPasskeyName(name); err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	return name, nil
}

// registrationPerson returns the person the registration of c creates a
// passkey for: the owner, or the viewer the ceremony was begun for, with the
// invitation of the link it began with.
func (a *Auth) registrationPerson(ctx context.Context, c *identity.Ceremony, owner *identity.Owner) (identity.Person, *identity.ViewerInvite, error) {
	if c.PersonID == "" {
		return identity.OwnerPerson(owner), nil, nil
	}
	code, _, err := a.store.EnrollmentCode(ctx, c.EnrollmentCode)
	if err != nil || code.Viewer == nil {
		return identity.Person{}, nil, fmt.Errorf("%w: the viewer's link is gone", ErrCeremony)
	}
	person, err := identity.NewViewer(c.PersonID, code.Viewer.Name, c.PersonHandle, a.nowSecond())
	if err != nil {
		return identity.Person{}, nil, fmt.Errorf("%w: %w", ErrRegistrationFailed, err)
	}
	return person, code.Viewer, nil
}

// BeginLogin starts a passkey login on the allowed origin text.
func (a *Auth) BeginLogin(ctx context.Context, origin string) (Started, error) {
	o, err := a.origin(origin)
	if err != nil {
		return Started{}, err
	}
	begun, err := a.ceremonies.BeginLogin(o)
	if err != nil {
		return Started{}, fmt.Errorf("beginning a login: %w", err)
	}
	c, err := a.saveCeremony(ctx, identity.CeremonyLogin, o, begun.State, "", "", nil)
	if err != nil {
		return Started{}, err
	}
	return Started{CeremonyID: c.ID, Options: begun.Options, ExpiresAt: c.ExpiresAt}, nil
}

// Passkey returns the passkey authenticator (D-26): it finishes a login
// ceremony and proves the owner, with owner scope.
func (a *Auth) Passkey() PasskeyAuthenticator { return PasskeyAuthenticator{auth: a} }

// PasskeyAuthenticator is the passkey authenticator (see Auth.Passkey).
type PasskeyAuthenticator struct {
	auth *Auth
}

// Authenticate implements Authenticator: it verifies the login assertion
// against the passkeys of origin's host only (D-60), refuses a signature
// counter that went backwards, and records the passkey's use.
func (p PasskeyAuthenticator) Authenticate(ctx context.Context, origin identity.Origin, proof Proof) (Authenticated, error) {
	owner, person, used, err := p.auth.assertPasskey(ctx, origin, identity.CeremonyLogin, proof)
	if err != nil {
		return Authenticated{}, err
	}
	scope := identity.SessionScopeOwner
	if person.IsViewer() {
		// A viewer only reads: every call that changes something needs an owner session.
		scope = identity.SessionScopeRead
	}
	return Authenticated{
		Owner: owner, Person: person.ID, Scope: scope, Authenticator: identity.AuthenticatorPasskey, Passkey: used.CredentialID,
	}, nil
}

// assertPasskey finishes the ceremony of proof, of kind, on origin: it
// verifies the assertion against the passkeys of origin's host only (D-60),
// refuses a signature counter that went backwards, and records the
// passkey's use. It returns the owner, the person the passkey belongs to and
// the passkey used. A viewer who was revoked can't sign in: their passkeys are
// disabled, and are refused here too.
func (a *Auth) assertPasskey(ctx context.Context, origin identity.Origin, kind identity.CeremonyKind, proof Proof) (identity.Owner, identity.Person, identity.Passkey, error) {
	c, err := a.takeCeremony(ctx, proof.CeremonyID, kind, origin)
	if err != nil {
		return identity.Owner{}, identity.Person{}, identity.Passkey{}, err
	}
	owner, err := a.loginOwner(ctx)
	if err != nil {
		return identity.Owner{}, identity.Person{}, identity.Passkey{}, err
	}
	people, err := a.loginPeople(ctx, &owner)
	if err != nil {
		return identity.Owner{}, identity.Person{}, identity.Passkey{}, err
	}
	keys, err := a.hostPasskeys(ctx, origin)
	if err != nil {
		return identity.Owner{}, identity.Person{}, identity.Passkey{}, err
	}
	got, err := a.ceremonies.FinishLogin(origin, people, keys, c.State, proof.Credential)
	if err != nil {
		return identity.Owner{}, identity.Person{}, identity.Passkey{}, fmt.Errorf("%w: %w", ErrAuthenticationFailed, err)
	}
	used, err := usedPasskey(keys, origin, &got, a.nowSecond())
	if err != nil {
		return identity.Owner{}, identity.Person{}, identity.Passkey{}, err
	}
	i := slices.IndexFunc(people, func(p identity.Person) bool { return p.ID == used.Person() && p.Active() })
	if i < 0 {
		return identity.Owner{}, identity.Person{}, identity.Passkey{}, fmt.Errorf("%w: the passkey's person is revoked", ErrAuthenticationFailed)
	}
	if err := a.store.UpdatePasskeyUse(ctx, &used); err != nil {
		return identity.Owner{}, identity.Person{}, identity.Passkey{}, fmt.Errorf("recording a passkey's use: %w", err)
	}
	return owner, people[i], used, nil
}

// loginOwner returns the owner a login is for; with no passkey made yet it
// fails with an error wrapping ErrAuthenticationFailed.
func (a *Auth) loginOwner(ctx context.Context) (identity.Owner, error) {
	owner, err := a.store.Owner(ctx)
	if errors.Is(err, ErrNoOwner) {
		return identity.Owner{}, fmt.Errorf("%w: no passkey exists", ErrAuthenticationFailed)
	}
	if err != nil {
		return identity.Owner{}, fmt.Errorf("loading the owner: %w", err)
	}
	return owner, nil
}

// loginPeople returns the people a login may name by their user handle, the
// owner with theirs.
func (a *Auth) loginPeople(ctx context.Context, owner *identity.Owner) ([]identity.Person, error) {
	people, err := a.store.People(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing people: %w", err)
	}
	for i := range people {
		if people[i].Kind == identity.PersonOwner {
			people[i] = identity.OwnerPerson(owner)
		}
	}
	return people, nil
}

// BeginPasskeyCheck starts a passkey check in the owner session of token on
// the allowed origin text (D-183): an assertion like a login's, which
// FinishPasskeyCheck records on the session. A read session, such as a demo
// session, can't begin one.
func (a *Auth) BeginPasskeyCheck(ctx context.Context, origin, token string) (Started, error) {
	o, _, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner)
	if err != nil {
		return Started{}, err
	}
	begun, err := a.ceremonies.BeginLogin(o)
	if err != nil {
		return Started{}, fmt.Errorf("beginning a passkey check: %w", err)
	}
	c, err := a.saveCeremony(ctx, identity.CeremonyPasskeyCheck, o, begun.State, "", "", nil)
	if err != nil {
		return Started{}, err
	}
	return Started{CeremonyID: c.ID, Options: begun.Options, ExpiresAt: c.ExpiresAt}, nil
}

// FinishPasskeyCheck verifies the assertion credential for the passkey
// check ceremonyID in the owner session of token on the allowed origin
// text, against the passkeys of its host, and records the check on the
// session, so the actions D-183 protects are allowed in it for the next
// identity.PasskeyCheckWindow. It returns the session after the check.
func (a *Auth) FinishPasskeyCheck(ctx context.Context, origin, token, ceremonyID string, credential []byte) (identity.Session, error) {
	o, s, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner)
	if err != nil {
		return identity.Session{}, err
	}
	_, person, _, assertErr := a.assertPasskey(ctx, o, identity.CeremonyPasskeyCheck, Proof{CeremonyID: ceremonyID, Credential: credential})
	if assertErr != nil {
		return identity.Session{}, assertErr
	}
	if person.ID != s.Person() {
		return identity.Session{}, fmt.Errorf("%w: the passkey is another person's", ErrAuthenticationFailed)
	}
	checked, err := s.PasskeyChecked(a.nowSecond())
	if err != nil {
		return identity.Session{}, fmt.Errorf("%w: %w", ErrForbidden, err)
	}
	if err := a.store.PasskeyCheckSession(ctx, s.ID, checked.PasskeyCheckedAt); err != nil {
		return identity.Session{}, fmt.Errorf("recording a passkey check: %w", err)
	}
	return checked, nil
}

// usedPasskey returns the passkey among keys that got names, after its use
// at now, refusing one of another host and a counter that went backwards.
func usedPasskey(keys []identity.Passkey, origin identity.Origin, got *Assertion, now time.Time) (identity.Passkey, error) {
	for i := range keys {
		if !bytes.Equal(keys[i].CredentialID, got.CredentialID) {
			continue
		}
		if err := keys[i].CheckHost(origin.RPID()); err != nil {
			return identity.Passkey{}, fmt.Errorf("%w: %w", ErrAuthenticationFailed, err)
		}
		used, err := keys[i].Used(got.SignCount, got.BackupState, now)
		if err != nil {
			return identity.Passkey{}, fmt.Errorf("%w: %w", ErrAuthenticationFailed, err)
		}
		return used, nil
	}
	return identity.Passkey{}, fmt.Errorf("%w: unknown passkey", ErrAuthenticationFailed)
}

// Passkeys returns the owner's passkeys, on every host, for an owner session:
// a viewer's are shown as a count with the viewer, never listed here.
func (a *Auth) Passkeys(ctx context.Context, origin, token string) ([]identity.Passkey, error) {
	if _, _, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner); err != nil {
		return nil, err
	}
	keys, err := a.store.Passkeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing passkeys: %w", err)
	}
	return slices.DeleteFunc(keys, func(k identity.Passkey) bool { return k.Person() != identity.OwnerPersonID }), nil
}

// RemovePasskey removes the passkey id (Passkey.ID) for an owner session
// with a passkey check in the last five minutes (D-183), and revokes every
// session that passkey signed in, so a session opened with a lost or stolen
// passkey ends with it. The last passkey can't be removed. It reports
// whether the session of token was among those revoked.
func (a *Auth) RemovePasskey(ctx context.Context, origin, token, id string) (bool, error) {
	_, s, err := a.recentOwner(ctx, origin, token)
	if err != nil {
		return false, err
	}
	credID, err := identity.ParsePasskeyID(id)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	err = within(ctx, a.store, func(ctx context.Context) error {
		if deleteErr := a.store.DeletePasskey(ctx, credID, a.now().UTC()); deleteErr != nil {
			return fmt.Errorf("deleting the passkey: %w", deleteErr)
		}
		return a.record(ctx, ownerActor(&s), identity.ActionPasskeyRemoved, identity.OutcomeOK, identity.Ref{Kind: identity.RefPasskey, ID: id})
	})
	switch {
	case errors.Is(err, ErrUnknownPasskey):
		return false, fmt.Errorf("%w: %w", ErrNotFound, err)
	case err != nil:
		return false, fmt.Errorf("removing a passkey: %w", err)
	}
	return s.SignedInBy(credID), nil
}

// checkGrant checks that grant allows creating a passkey on o, and returns
// the enrollment code's ID when it is a code, or the owner session's ID when
// it is a session.
func (a *Auth) checkGrant(ctx context.Context, o identity.Origin, grant RegistrationGrant) (identity.EnrollmentCodeID, identity.SessionID, *identity.ViewerInvite, error) {
	switch {
	case grant.EnrollmentCode != "" && grant.SessionToken != "":
		return "", "", nil, fmt.Errorf("%w: an enrollment code or a session, not both", ErrInvalidRequest)
	case grant.EnrollmentCode != "":
		c, err := a.enrollmentCode(ctx, o, grant.EnrollmentCode)
		if err != nil {
			return "", "", nil, err
		}
		return c.ID, "", c.Viewer, nil
	case grant.SessionToken != "":
		s, err := a.Session(ctx, o, grant.SessionToken, identity.SessionScopeOwner)
		if err != nil {
			return "", "", nil, err
		}
		if err := s.CheckRecentPasskey(a.now().UTC()); err != nil {
			return "", "", nil, fmt.Errorf("%w: %w", ErrPasskeyCheckRequired, err)
		}
		return "", s.ID, nil, nil
	}
	return "", "", nil, fmt.Errorf("%w: creating a passkey needs an enrollment code or an owner session", ErrForbidden)
}

// enrollmentCode returns the stored code text stands for when it may be
// used now on o; otherwise an error wrapping ErrForbidden.
func (a *Auth) enrollmentCode(ctx context.Context, o identity.Origin, text string) (identity.EnrollmentCode, error) {
	id, key, err := identity.ParseEnrollmentCode(text)
	if err != nil {
		return identity.EnrollmentCode{}, fmt.Errorf("%w: %w", ErrForbidden, err)
	}
	c, hash, err := a.store.EnrollmentCode(ctx, id)
	if errors.Is(err, ErrUnknownEnrollmentCode) {
		return identity.EnrollmentCode{}, fmt.Errorf("%w: unknown enrollment code", ErrForbidden)
	}
	if err != nil {
		return identity.EnrollmentCode{}, fmt.Errorf("checking an enrollment code: %w", err)
	}
	if !identity.KeyMatches(key, hash) {
		return identity.EnrollmentCode{}, fmt.Errorf("%w: enrollment code does not match", ErrForbidden)
	}
	if err := c.Check(a.now().UTC(), o); err != nil {
		// A link that was spent, has expired or is used on another origin is refused and recorded,
		// by its ID: only someone who holds its key gets this far (D-211).
		refusal := fmt.Errorf("%w: %w", ErrForbidden, err)
		actor := identity.Actor{Kind: identity.ActorEnrollmentCode, ID: string(c.ID)}
		if auditErr := a.record(ctx, actor, identity.ActionEnrollmentLinkUsed, identity.OutcomeRefused, identity.Ref{Kind: identity.RefEnrollmentCode, ID: string(c.ID)}); auditErr != nil {
			return identity.EnrollmentCode{}, fmt.Errorf("%w (the refusal could not be audited: %w)", refusal, auditErr)
		}
		return identity.EnrollmentCode{}, refusal
	}
	return c, nil
}

// storeRegistration uses up the enrollment code of c, if any, stores p and,
// for a code, logs owner in on c's origin, all in one transaction with the
// audit entries that record them (D-211): the code's use and the passkey's
// creation, by whoever began the registration, and the sign-in.
func (a *Auth) storeRegistration(ctx context.Context, c *identity.Ceremony, owner *identity.Owner, p *identity.Passkey) (Registered, error) {
	var out Registered
	err := within(ctx, a.store, func(ctx context.Context) error {
		actor := identity.Actor{Kind: identity.ActorOwner, ID: string(c.Session)}
		if c.EnrollmentCode != "" {
			actor = identity.Actor{Kind: identity.ActorEnrollmentCode, ID: string(c.EnrollmentCode)}
			if err := a.useCode(ctx, c, p, actor); err != nil {
				return err
			}
		}
		if err := a.store.SavePasskey(ctx, p); err != nil {
			return fmt.Errorf("storing a passkey: %w", err)
		}
		passkey := identity.Ref{Kind: identity.RefPasskey, ID: p.ID()}
		if err := a.record(ctx, actor, identity.ActionPasskeyAdded, identity.OutcomeOK, passkey); err != nil {
			return err
		}
		out = Registered{Passkey: *p}
		if c.EnrollmentCode == "" {
			return nil
		}
		// Creating the passkey verified the owner on it, so it signs the session
		// in and counts as its first passkey check (D-183).
		s, err := a.issueSession(ctx, &Authenticated{
			Owner: *owner, Scope: identity.SessionScopeOwner, Authenticator: identity.AuthenticatorPasskey, Passkey: p.CredentialID,
		}, c.Origin)
		if err != nil {
			return err
		}
		out.Session = &s
		return nil
	})
	if err != nil {
		return Registered{}, err
	}
	return out, nil
}

// useCode uses up the enrollment code of c for creating p and records it,
// naming the code and the passkey, as actor's.
func (a *Auth) useCode(ctx context.Context, c *identity.Ceremony, p *identity.Passkey, actor identity.Actor) error {
	err := a.store.UseEnrollmentCode(ctx, c.EnrollmentCode, p.CreatedAt)
	if errors.Is(err, identity.ErrEnrollmentCodeUsed) {
		return fmt.Errorf("%w: %w", ErrForbidden, err)
	}
	if err != nil {
		return fmt.Errorf("using an enrollment code: %w", err)
	}
	code := identity.Ref{Kind: identity.RefEnrollmentCode, ID: string(c.EnrollmentCode)}
	return a.record(ctx, actor, identity.ActionEnrollmentLinkUsed, identity.OutcomeOK, code, identity.Ref{Kind: identity.RefPasskey, ID: p.ID()})
}

// hostPasskeys returns the passkeys of o's host.
func (a *Auth) hostPasskeys(ctx context.Context, o identity.Origin) ([]identity.Passkey, error) {
	all, err := a.store.Passkeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing passkeys: %w", err)
	}
	var out []identity.Passkey
	for i := range all {
		if all[i].RPID == o.RPID() {
			out = append(out, all[i])
		}
	}
	return out, nil
}

// saveCeremony stores a new ceremony of kind on o with state, begun with
// the enrollment code or the owner session, for a registration.
func (a *Auth) saveCeremony(ctx context.Context, kind identity.CeremonyKind, o identity.Origin, state []byte, code identity.EnrollmentCodeID, session identity.SessionID, viewer *identity.Person) (identity.Ceremony, error) {
	id, err := a.randomHex(identity.CeremonyIDBytes)
	if err != nil {
		return identity.Ceremony{}, fmt.Errorf("beginning a %s: %w", kind, err)
	}
	c := identity.NewCeremony(id, kind, o, state, a.now().UTC())
	c.EnrollmentCode, c.Session = code, session
	if viewer != nil {
		c.PersonID, c.PersonHandle = viewer.ID, viewer.Handle
	}
	if err := a.store.SaveCeremony(ctx, &c); err != nil {
		return identity.Ceremony{}, fmt.Errorf("beginning a %s: %w", kind, err)
	}
	return c, nil
}

// takeCeremony takes the ceremony id out of the store, so it is used once,
// and checks that it may be finished now as kind on o.
func (a *Auth) takeCeremony(ctx context.Context, id string, kind identity.CeremonyKind, o identity.Origin) (identity.Ceremony, error) {
	c, err := a.store.TakeCeremony(ctx, id)
	if errors.Is(err, ErrUnknownCeremony) {
		return identity.Ceremony{}, fmt.Errorf("%w: %w", ErrCeremony, err)
	}
	if err != nil {
		return identity.Ceremony{}, fmt.Errorf("finishing a %s: %w", kind, err)
	}
	if err := c.Check(a.now().UTC(), kind, o); err != nil {
		return identity.Ceremony{}, fmt.Errorf("%w: %w", ErrCeremony, err)
	}
	return c, nil
}
