package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
)

// Error classes of the auth use cases. Every error they return wraps one of
// these, or is a store failure; the API adapter maps them to status codes.
var (
	// ErrInvalidSession reports a session token that is malformed, unknown,
	// not matching, revoked, expired or from another host: log in again.
	ErrInvalidSession = errors.New("invalid session")
	// ErrForbidden reports a request that is understood but not allowed: an
	// origin off the allow-list, a session without the scope, an invalid
	// enrollment code, or demo sessions while they are off.
	ErrForbidden = errors.New("forbidden")
	// ErrAuthenticationFailed reports a login whose assertion does not
	// verify, names an unknown passkey or a passkey of another host, or
	// whose signature counter went backwards.
	ErrAuthenticationFailed = errors.New("authentication failed")
	// ErrRegistrationFailed reports a new credential that does not verify.
	ErrRegistrationFailed = errors.New("registration failed")
	// ErrCeremony reports a ceremony that is unknown, used, expired, or
	// finished as another kind or on another origin: begin again.
	ErrCeremony = errors.New("ceremony not usable")
	// ErrInvalidRequest reports a bad argument, such as a token name or a
	// lifetime out of range.
	ErrInvalidRequest = errors.New("invalid request")
	// ErrNotFound reports a passkey or script token that does not exist.
	ErrNotFound = errors.New("not found")
)

// AuthConfig is what the auth use cases are configured with.
type AuthConfig struct {
	// Origins is the allow-list of origins the owner's browser may use.
	Origins identity.Origins
	// Sessions is how long sessions last.
	Sessions identity.SessionPolicy
	// DemoSessions turns the demo authenticator on; it works only on a
	// database with the demo marker as well.
	DemoSessions bool
}

// Auth is the authentication use cases (author decision, auth split):
// authenticators prove who is at the browser, always the owner (GR-2), and
// with what scope; Auth then issues an opaque, revocable session, and lets
// an owner session manage passkeys, script tokens and enrollment links.
type Auth struct {
	store      AuthStore
	ceremonies Ceremonies
	tokens     *Tokens
	demo       DemoMarker
	cfg        AuthConfig
	now        func() time.Time
	random     io.Reader
}

// NewAuth returns the auth use cases over store, ceremonies, the script
// token use cases and the demo marker, reading the time from now and
// secrets from random; a nil random is crypto/rand.Reader.
func NewAuth(store AuthStore, ceremonies Ceremonies, tokens *Tokens, demo DemoMarker, cfg AuthConfig, now func() time.Time, random io.Reader) (*Auth, error) {
	if err := cfg.Sessions.Validate(); err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}
	if len(cfg.Origins.List()) == 0 {
		return nil, fmt.Errorf("auth: %w: no allowed origins", identity.ErrInvalidOrigin)
	}
	if random == nil {
		random = rand.Reader
	}
	return &Auth{store: store, ceremonies: ceremonies, tokens: tokens, demo: demo, cfg: cfg, now: now, random: random}, nil
}

// Authenticated is what an authenticator proved: the owner, with the scope
// the session gets.
type Authenticated struct {
	Owner         identity.Owner
	Scope         identity.SessionScope
	Authenticator identity.Authenticator
}

// Proof is what the browser sends to finish an authentication: the
// ceremony and its credential, for a passkey; nothing, for the demo.
type Proof struct {
	CeremonyID string
	Credential []byte
}

// Authenticator proves who is at the browser on origin (author decision,
// auth split). Every authenticator resolves to the single owner.
type Authenticator interface {
	Authenticate(ctx context.Context, origin identity.Origin, proof Proof) (Authenticated, error)
}

// NewSession is a session just issued, with its token, shown only once.
type NewSession struct {
	Session identity.Session
	Token   string
}

// Login runs authenticator on the allowed origin text with proof and, when
// it succeeds, issues a session on that origin's host.
func (a *Auth) Login(ctx context.Context, authenticator Authenticator, origin string, proof Proof) (NewSession, error) {
	o, err := a.origin(origin)
	if err != nil {
		return NewSession{}, err
	}
	got, err := authenticator.Authenticate(ctx, o, proof)
	if err != nil {
		return NewSession{}, fmt.Errorf("authenticating: %w", err)
	}
	return a.issueSession(ctx, got, o)
}

// Session returns the session of token when it may be used now on origin's
// host for what want allows, recording its use. Otherwise it fails with an
// error wrapping ErrInvalidSession or, for a valid session without the
// scope, ErrForbidden. A zero origin skips the host check.
func (a *Auth) Session(ctx context.Context, origin identity.Origin, token string, want identity.SessionScope) (identity.Session, error) {
	id, key, err := identity.ParseSessionToken(token)
	if err != nil {
		return identity.Session{}, fmt.Errorf("%w: %w", ErrInvalidSession, err)
	}
	s, hash, err := a.store.Session(ctx, id)
	if errors.Is(err, ErrUnknownSession) {
		return identity.Session{}, fmt.Errorf("%w: unknown", ErrInvalidSession)
	}
	if err != nil {
		return identity.Session{}, fmt.Errorf("checking a session: %w", err)
	}
	if !identity.KeyMatches(key, hash) {
		return identity.Session{}, fmt.Errorf("%w: key does not match", ErrInvalidSession)
	}
	now := a.now().UTC()
	rpID := ""
	if !origin.IsZero() {
		rpID = origin.RPID()
	}
	if err := s.Check(now, rpID, want); err != nil {
		if errors.Is(err, identity.ErrSessionScope) {
			return identity.Session{}, fmt.Errorf("%w: %w", ErrForbidden, err)
		}
		return identity.Session{}, fmt.Errorf("%w: %w", ErrInvalidSession, err)
	}
	touched, store := s.Touched(now)
	if store {
		if err := a.store.TouchSession(ctx, s.ID, touched.LastSeenAt); err != nil {
			return identity.Session{}, fmt.Errorf("recording a session's use: %w", err)
		}
	}
	return touched, nil
}

// Logout revokes the session of token, of any scope.
func (a *Auth) Logout(ctx context.Context, origin, token string) error {
	_, s, err := a.sessionOn(ctx, origin, token, identity.SessionScopeRead)
	if err != nil {
		return err
	}
	if err := a.store.RevokeSession(ctx, s.ID, a.now().UTC()); err != nil {
		return fmt.Errorf("logging out: %w", err)
	}
	return nil
}

// WhoAmI returns the owner and the session of token, of any scope.
func (a *Auth) WhoAmI(ctx context.Context, origin, token string) (identity.Owner, identity.Session, error) {
	_, s, err := a.sessionOn(ctx, origin, token, identity.SessionScopeRead)
	if err != nil {
		return identity.Owner{}, identity.Session{}, err
	}
	owner, err := a.store.Owner(ctx)
	if err != nil {
		return identity.Owner{}, identity.Session{}, fmt.Errorf("loading the owner: %w", err)
	}
	return owner, s, nil
}

// Principal is who a read API bearer token stands for: always the owner,
// through a script token or a session.
type Principal struct {
	// Session is the session's ID; empty for a script token.
	Session identity.SessionID
	// ScriptToken is the script token's ID; empty for a session.
	ScriptToken identity.ScriptTokenID
}

// Authenticate checks a read API bearer token: a session of any scope (read
// is the least), or a read-scoped script token (D-27). An invalid token
// fails with an error wrapping ErrInvalidToken; a store failure with
// another error. Errors never contain the token.
func (a *Auth) Authenticate(ctx context.Context, token string) (Principal, error) {
	if strings.HasPrefix(token, identity.SessionPrefix) {
		s, err := a.Session(ctx, identity.Origin{}, token, identity.SessionScopeRead)
		if errors.Is(err, ErrInvalidSession) || errors.Is(err, ErrForbidden) {
			return Principal{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
		}
		if err != nil {
			return Principal{}, err
		}
		return Principal{Session: s.ID}, nil
	}
	if err := a.tokens.Verify(ctx, token, identity.ScopeRead); err != nil {
		return Principal{}, err
	}
	id, _, err := identity.ParseSecret(token)
	if err != nil {
		return Principal{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	return Principal{ScriptToken: id}, nil
}

// LoginMethods is what the login page can offer on an origin.
type LoginMethods struct {
	// PasskeyRegistered is whether a passkey exists for the origin's host.
	PasskeyRegistered bool
	// DemoSessions is whether the demo authenticator would issue a session.
	DemoSessions bool
}

// LoginMethods returns what the login page can offer on the allowed origin
// text.
func (a *Auth) LoginMethods(ctx context.Context, origin string) (LoginMethods, error) {
	o, err := a.origin(origin)
	if err != nil {
		return LoginMethods{}, err
	}
	keys, err := a.store.Passkeys(ctx)
	if err != nil {
		return LoginMethods{}, fmt.Errorf("listing passkeys: %w", err)
	}
	var m LoginMethods
	for i := range keys {
		if keys[i].RPID == o.RPID() {
			m.PasskeyRegistered = true
		}
	}
	m.DemoSessions, err = a.demoAvailable(ctx)
	return m, err
}

// Demo returns the demo authenticator: a read-scoped session for the owner
// without credentials (author decision, demo read-only sessions), only when
// the core runs with demo sessions on AND the database carries the demo
// marker. Otherwise it fails with ErrForbidden.
func (a *Auth) Demo() DemoAuthenticator { return DemoAuthenticator{auth: a} }

// DemoAuthenticator is the demo authenticator (see Auth.Demo).
type DemoAuthenticator struct {
	auth *Auth
}

// Authenticate implements Authenticator: the owner, read scope, when demo
// sessions are available.
func (d DemoAuthenticator) Authenticate(ctx context.Context, _ identity.Origin, _ Proof) (Authenticated, error) {
	ok, err := d.auth.demoAvailable(ctx)
	if err != nil {
		return Authenticated{}, err
	}
	if !ok {
		return Authenticated{}, fmt.Errorf("%w: demo sessions are off", ErrForbidden)
	}
	owner, err := d.auth.ensureOwner(ctx)
	if err != nil {
		return Authenticated{}, err
	}
	return Authenticated{Owner: owner, Scope: identity.SessionScopeRead, Authenticator: identity.AuthenticatorDemo}, nil
}

// issueSession stores a new session for got on o's host.
func (a *Auth) issueSession(ctx context.Context, got Authenticated, o identity.Origin) (NewSession, error) {
	id, err := a.randomHex(identity.IDBytes)
	if err != nil {
		return NewSession{}, fmt.Errorf("issuing a session: %w", err)
	}
	key := make([]byte, identity.KeyBytes)
	if _, readErr := io.ReadFull(a.random, key); readErr != nil {
		return NewSession{}, fmt.Errorf("issuing a session: random key: %w", readErr)
	}
	s, err := identity.NewSession(identity.SessionID(id), got.Authenticator, got.Scope, o.RPID(), a.nowSecond(), a.cfg.Sessions)
	if err != nil {
		return NewSession{}, fmt.Errorf("issuing a session: %w", err)
	}
	if err := a.store.SaveSession(ctx, &s, identity.HashKey(key)); err != nil {
		return NewSession{}, fmt.Errorf("issuing a session: %w", err)
	}
	return NewSession{Session: s, Token: identity.FormatSessionToken(s.ID, key)}, nil
}

// sessionOn checks the allowed origin text and the session of token on it
// for want.
func (a *Auth) sessionOn(ctx context.Context, origin, token string, want identity.SessionScope) (identity.Origin, identity.Session, error) {
	o, err := a.origin(origin)
	if err != nil {
		return identity.Origin{}, identity.Session{}, err
	}
	s, err := a.Session(ctx, o, token, want)
	if err != nil {
		return identity.Origin{}, identity.Session{}, err
	}
	return o, s, nil
}

// origin parses text and checks it against the allow-list; a bad or
// disallowed origin is ErrForbidden.
func (a *Auth) origin(text string) (identity.Origin, error) {
	o, err := a.cfg.Origins.Allowed(text)
	if err != nil {
		return identity.Origin{}, fmt.Errorf("%w: %w", ErrForbidden, err)
	}
	return o, nil
}

// randomHex returns n random bytes in lowercase hex.
func (a *Auth) randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(a.random, b); err != nil {
		return "", fmt.Errorf("random ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// nowSecond returns now in UTC without its fraction of a second, as stored.
func (a *Auth) nowSecond() time.Time { return a.now().UTC().Truncate(time.Second) }

// demoAvailable reports whether both conditions for demo sessions hold:
// the configuration turns them on, and the database carries the demo
// marker. The marker is read only when the configuration allows them.
func (a *Auth) demoAvailable(ctx context.Context) (bool, error) {
	if !a.cfg.DemoSessions || a.demo == nil {
		return false, nil
	}
	marked, err := a.demo.DemoMarked(ctx)
	if err != nil {
		return false, fmt.Errorf("reading the demo marker: %w", err)
	}
	return marked, nil
}

// ensureOwner returns the owner, creating them with a random handle on
// first use.
func (a *Auth) ensureOwner(ctx context.Context) (identity.Owner, error) {
	owner, err := a.store.Owner(ctx)
	if err == nil {
		return owner, nil
	}
	if !errors.Is(err, ErrNoOwner) {
		return identity.Owner{}, fmt.Errorf("loading the owner: %w", err)
	}
	handle := make([]byte, identity.OwnerHandleBytes)
	if _, readErr := io.ReadFull(a.random, handle); readErr != nil {
		return identity.Owner{}, fmt.Errorf("creating the owner: random handle: %w", readErr)
	}
	candidate, err := identity.NewOwner(handle, a.nowSecond())
	if err != nil {
		return identity.Owner{}, fmt.Errorf("creating the owner: %w", err)
	}
	owner, err = a.store.EnsureOwner(ctx, candidate)
	if err != nil {
		return identity.Owner{}, fmt.Errorf("creating the owner: %w", err)
	}
	return owner, nil
}
