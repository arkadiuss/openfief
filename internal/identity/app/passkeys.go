package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
)

// Started is a WebAuthn ceremony the browser is to carry out: the ID to
// finish it with, the options for the browser, and when it expires.
type Started struct {
	CeremonyID string
	Options    []byte
	ExpiresAt  time.Time
}

// RegistrationGrant is what allows creating a passkey: exactly one of a
// one-time enrollment code issued for the origin, or an owner session on
// the origin's host (D-60).
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
	code, err := a.checkGrant(ctx, o, grant)
	if err != nil {
		return Started{}, err
	}
	owner, err := a.ensureOwner(ctx)
	if err != nil {
		return Started{}, err
	}
	keys, err := a.hostPasskeys(ctx, o)
	if err != nil {
		return Started{}, err
	}
	begun, err := a.ceremonies.BeginRegistration(o, &owner, keys)
	if err != nil {
		return Started{}, fmt.Errorf("beginning a registration: %w", err)
	}
	c, err := a.saveCeremony(ctx, identity.CeremonyRegistration, o, begun.State, code)
	if err != nil {
		return Started{}, err
	}
	return Started{CeremonyID: c.ID, Options: begun.Options, ExpiresAt: c.ExpiresAt}, nil
}

// Registered is a passkey just created, and, when an enrollment code
// allowed it, the owner session that comes with it.
type Registered struct {
	Passkey identity.Passkey
	// Session is set only for a registration begun with an enrollment code.
	Session *NewSession
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
	if name == "" {
		name = o.RPID()
	}
	if nameErr := identity.CheckPasskeyName(name); nameErr != nil {
		return Registered{}, fmt.Errorf("%w: %w", ErrInvalidRequest, nameErr)
	}
	c, err := a.takeCeremony(ctx, ceremonyID, identity.CeremonyRegistration, o)
	if err != nil {
		return Registered{}, err
	}
	owner, err := a.store.Owner(ctx)
	if err != nil {
		return Registered{}, fmt.Errorf("loading the owner: %w", err)
	}
	p, err := a.ceremonies.FinishRegistration(o, &owner, c.State, credential)
	if err != nil {
		return Registered{}, fmt.Errorf("%w: %w", ErrRegistrationFailed, err)
	}
	now := a.nowSecond()
	p.Name, p.RPID, p.CreatedAt = name, o.RPID(), now
	if err := p.Validate(); err != nil {
		return Registered{}, fmt.Errorf("%w: %w", ErrRegistrationFailed, err)
	}
	return a.storeRegistration(ctx, &c, &owner, &p)
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
	c, err := a.saveCeremony(ctx, identity.CeremonyLogin, o, begun.State, "")
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
	a := p.auth
	c, err := a.takeCeremony(ctx, proof.CeremonyID, identity.CeremonyLogin, origin)
	if err != nil {
		return Authenticated{}, err
	}
	owner, err := a.store.Owner(ctx)
	if errors.Is(err, ErrNoOwner) {
		return Authenticated{}, fmt.Errorf("%w: no passkey exists", ErrAuthenticationFailed)
	}
	if err != nil {
		return Authenticated{}, fmt.Errorf("loading the owner: %w", err)
	}
	keys, err := a.hostPasskeys(ctx, origin)
	if err != nil {
		return Authenticated{}, err
	}
	got, err := a.ceremonies.FinishLogin(origin, &owner, keys, c.State, proof.Credential)
	if err != nil {
		return Authenticated{}, fmt.Errorf("%w: %w", ErrAuthenticationFailed, err)
	}
	used, err := usedPasskey(keys, origin, &got, a.nowSecond())
	if err != nil {
		return Authenticated{}, err
	}
	if err := a.store.UpdatePasskeyUse(ctx, &used); err != nil {
		return Authenticated{}, fmt.Errorf("recording a passkey's use: %w", err)
	}
	return Authenticated{Owner: owner, Scope: identity.SessionScopeOwner, Authenticator: identity.AuthenticatorPasskey}, nil
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

// Passkeys returns every passkey, on every host, for an owner session.
func (a *Auth) Passkeys(ctx context.Context, origin, token string) ([]identity.Passkey, error) {
	if _, _, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner); err != nil {
		return nil, err
	}
	keys, err := a.store.Passkeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing passkeys: %w", err)
	}
	return keys, nil
}

// RemovePasskey removes the passkey id (Passkey.ID) for an owner session;
// the last one can't be removed.
func (a *Auth) RemovePasskey(ctx context.Context, origin, token, id string) error {
	if _, _, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner); err != nil {
		return err
	}
	credID, err := identity.ParsePasskeyID(id)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	err = a.store.DeletePasskey(ctx, credID)
	switch {
	case errors.Is(err, ErrUnknownPasskey):
		return fmt.Errorf("%w: %w", ErrNotFound, err)
	case err != nil:
		return fmt.Errorf("removing a passkey: %w", err)
	}
	return nil
}

// checkGrant checks that grant allows creating a passkey on o, and returns
// the enrollment code's ID when it is a code.
func (a *Auth) checkGrant(ctx context.Context, o identity.Origin, grant RegistrationGrant) (identity.EnrollmentCodeID, error) {
	switch {
	case grant.EnrollmentCode != "" && grant.SessionToken != "":
		return "", fmt.Errorf("%w: an enrollment code or a session, not both", ErrInvalidRequest)
	case grant.EnrollmentCode != "":
		c, err := a.enrollmentCode(ctx, o, grant.EnrollmentCode)
		if err != nil {
			return "", err
		}
		return c.ID, nil
	case grant.SessionToken != "":
		if _, err := a.Session(ctx, o, grant.SessionToken, identity.SessionScopeOwner); err != nil {
			return "", err
		}
		return "", nil
	}
	return "", fmt.Errorf("%w: creating a passkey needs an enrollment code or an owner session", ErrForbidden)
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
		return identity.EnrollmentCode{}, fmt.Errorf("%w: %w", ErrForbidden, err)
	}
	return c, nil
}

// storeRegistration uses up the enrollment code of c, if any, stores p and,
// for a code, logs owner in on c's origin.
func (a *Auth) storeRegistration(ctx context.Context, c *identity.Ceremony, owner *identity.Owner, p *identity.Passkey) (Registered, error) {
	if c.EnrollmentCode != "" {
		err := a.store.UseEnrollmentCode(ctx, c.EnrollmentCode, p.CreatedAt)
		if errors.Is(err, identity.ErrEnrollmentCodeUsed) {
			return Registered{}, fmt.Errorf("%w: %w", ErrForbidden, err)
		}
		if err != nil {
			return Registered{}, fmt.Errorf("using an enrollment code: %w", err)
		}
	}
	if err := a.store.SavePasskey(ctx, p); err != nil {
		return Registered{}, fmt.Errorf("storing a passkey: %w", err)
	}
	out := Registered{Passkey: *p}
	if c.EnrollmentCode == "" {
		return out, nil
	}
	s, err := a.issueSession(ctx, Authenticated{Owner: *owner, Scope: identity.SessionScopeOwner, Authenticator: identity.AuthenticatorPasskey}, c.Origin)
	if err != nil {
		return Registered{}, err
	}
	out.Session = &s
	return out, nil
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

// saveCeremony stores a new ceremony of kind on o with state.
func (a *Auth) saveCeremony(ctx context.Context, kind identity.CeremonyKind, o identity.Origin, state []byte, code identity.EnrollmentCodeID) (identity.Ceremony, error) {
	id, err := a.randomHex(identity.CeremonyIDBytes)
	if err != nil {
		return identity.Ceremony{}, fmt.Errorf("beginning a %s: %w", kind, err)
	}
	c := identity.NewCeremony(id, kind, o, state, a.now().UTC())
	c.EnrollmentCode = code
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
