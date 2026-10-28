package identity

import (
	"errors"
	"fmt"
	"time"
)

// Errors about sessions. Checks return them wrapped.
var (
	// ErrMalformedSession reports text that is not a session token.
	ErrMalformedSession = errors.New("malformed session token")
	// ErrSessionExpired reports a session used at or after its absolute
	// expiry.
	ErrSessionExpired = errors.New("session expired")
	// ErrSessionIdle reports a session not used for longer than its idle
	// timeout.
	ErrSessionIdle = errors.New("session expired after being idle")
	// ErrSessionRevoked reports a session the owner logged out of.
	ErrSessionRevoked = errors.New("session revoked")
	// ErrSessionWrongHost reports a session used on another host than the
	// one it was issued on.
	ErrSessionWrongHost = errors.New("session issued on another host")
	// ErrSessionScope reports a session used for something its scope does
	// not allow.
	ErrSessionScope = errors.New("session scope does not allow this")
	// ErrInvalidSessionPolicy reports session lifetimes that are not
	// positive, or an idle timeout longer than the absolute lifetime.
	ErrInvalidSessionPolicy = errors.New("invalid session lifetimes")
)

const (
	// SessionPrefix starts every session token, so a leaked one is easy to
	// recognize and tell from a script token.
	SessionPrefix = "ofsess_"
	// DefaultSessionIdle is how long a session lasts unused by default.
	DefaultSessionIdle = 12 * time.Hour
	// DefaultSessionMax is how long a session lasts at most by default.
	DefaultSessionMax = 7 * 24 * time.Hour
	// sessionTouchEvery is how stale a session's last use may get before
	// it is stored again, so that every request does not write.
	sessionTouchEvery = time.Minute
)

// SessionScope is what a session may do (author decision, demo read-only
// sessions).
type SessionScope int

// The session scopes. The zero SessionScope is not a scope.
const (
	// SessionScopeRead may only read: the read API, WhoAmI and Logout.
	SessionScopeRead SessionScope = iota + 1
	// SessionScopeOwner may do everything the owner may: read, and manage
	// passkeys, script tokens and enrollment links.
	SessionScopeOwner
)

// String returns the scope's name: "read" or "owner".
func (s SessionScope) String() string {
	switch s {
	case SessionScopeRead:
		return "read"
	case SessionScopeOwner:
		return "owner"
	}
	return fmt.Sprintf("SessionScope(%d)", int(s))
}

// ParseSessionScope parses a scope name as String writes it.
func ParseSessionScope(s string) (SessionScope, error) {
	for _, sc := range []SessionScope{SessionScopeRead, SessionScopeOwner} {
		if s == sc.String() {
			return sc, nil
		}
	}
	return 0, fmt.Errorf("%w: unknown scope %q", ErrSessionScope, s)
}

// Includes reports whether a session with scope s may do what want allows:
// owner includes read.
func (s SessionScope) Includes(want SessionScope) bool {
	switch want {
	case SessionScopeRead:
		return s == SessionScopeRead || s == SessionScopeOwner
	case SessionScopeOwner:
		return s == SessionScopeOwner
	}
	return false
}

// Authenticator is how the owner proved who they are before a session was
// issued (author decision, auth split). Each authenticator gives one scope.
type Authenticator int

// The authenticators. The zero Authenticator is none.
const (
	// AuthenticatorPasskey is a WebAuthn passkey (D-26): owner scope.
	AuthenticatorPasskey Authenticator = iota + 1
	// AuthenticatorDemo needs no credential and gives read scope, on demo
	// databases only.
	AuthenticatorDemo
)

// String returns the authenticator's name: "passkey" or "demo".
func (a Authenticator) String() string {
	switch a {
	case AuthenticatorPasskey:
		return "passkey"
	case AuthenticatorDemo:
		return "demo"
	}
	return fmt.Sprintf("Authenticator(%d)", int(a))
}

// ParseAuthenticator parses an authenticator name as String writes it.
func ParseAuthenticator(s string) (Authenticator, error) {
	for _, a := range []Authenticator{AuthenticatorPasskey, AuthenticatorDemo} {
		if s == a.String() {
			return a, nil
		}
	}
	return 0, fmt.Errorf("%w: unknown authenticator %q", ErrMalformedSession, s)
}

// Scope returns the scope a session from a gets: owner for a passkey, read
// for the demo authenticator.
func (a Authenticator) Scope() (SessionScope, error) {
	switch a {
	case AuthenticatorPasskey:
		return SessionScopeOwner, nil
	case AuthenticatorDemo:
		return SessionScopeRead, nil
	}
	return 0, fmt.Errorf("%w: %s", ErrSessionScope, a)
}

// SessionPolicy is how long sessions last: Idle without use, Max at most.
type SessionPolicy struct {
	Idle time.Duration
	Max  time.Duration
}

// DefaultSessionPolicy returns 12 hours idle and 7 days at most.
func DefaultSessionPolicy() SessionPolicy {
	return SessionPolicy{Idle: DefaultSessionIdle, Max: DefaultSessionMax}
}

// Validate checks that both lifetimes are positive and Idle is at most Max.
func (p SessionPolicy) Validate() error {
	if p.Idle <= 0 || p.Max <= 0 || p.Idle > p.Max {
		return fmt.Errorf("%w: idle %s, max %s", ErrInvalidSessionPolicy, p.Idle, p.Max)
	}
	return nil
}

// SessionID identifies a session. It is not secret.
type SessionID string

// Session is the owner's login on one host: an opaque, revocable bearer
// token the core issued after an authenticator proved who the owner is
// (author decision, auth split). It holds no secret: only the hash of the
// token's key is stored.
type Session struct {
	ID            SessionID
	Scope         SessionScope
	Authenticator Authenticator
	// RPID is the host the session was issued on, and the only one it
	// works on.
	RPID       string
	CreatedAt  time.Time
	LastSeenAt time.Time
	// ExpiresAt is the absolute expiry, CreatedAt plus the policy's Max.
	ExpiresAt time.Time
	// IdleTimeout is the policy's Idle when the session was issued.
	IdleTimeout time.Duration
	// RevokedAt is when the owner logged out; zero while they did not.
	RevokedAt time.Time
}

// NewSession returns a session with id and scope, from authenticator a, on
// host rpID, issued at now under policy. An authenticator grants at most
// its own scope (Authenticator.Scope).
func NewSession(id SessionID, a Authenticator, scope SessionScope, rpID string, now time.Time, policy SessionPolicy) (Session, error) {
	if !isHexID(string(id)) {
		return Session{}, fmt.Errorf("%w: bad ID", ErrMalformedSession)
	}
	most, err := a.Scope()
	if err != nil {
		return Session{}, err
	}
	if !most.Includes(scope) {
		return Session{}, fmt.Errorf("%w: a %s session can't have %s scope", ErrSessionScope, a, scope)
	}
	if err := policy.Validate(); err != nil {
		return Session{}, err
	}
	if !validHost(rpID) {
		return Session{}, fmt.Errorf("%w: bad host", ErrMalformedSession)
	}
	return Session{
		ID: id, Scope: scope, Authenticator: a, RPID: rpID,
		CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(policy.Max), IdleTimeout: policy.Idle,
	}, nil
}

// IdleExpiresAt returns when the session expires if it is not used again:
// LastSeenAt plus the idle timeout, but never after ExpiresAt.
func (s *Session) IdleExpiresAt() time.Time {
	idle := s.LastSeenAt.Add(s.IdleTimeout)
	if idle.After(s.ExpiresAt) {
		return s.ExpiresAt
	}
	return idle
}

// Check returns nil when s may be used at now on host rpID for what want
// allows. Otherwise, in this order: ErrSessionRevoked, ErrSessionExpired,
// ErrSessionIdle, ErrSessionWrongHost, ErrSessionScope.
func (s *Session) Check(now time.Time, rpID string, want SessionScope) error {
	switch {
	case !s.RevokedAt.IsZero():
		return ErrSessionRevoked
	case !now.Before(s.ExpiresAt):
		return ErrSessionExpired
	case !now.Before(s.LastSeenAt.Add(s.IdleTimeout)):
		return ErrSessionIdle
	case rpID != "" && rpID != s.RPID:
		return ErrSessionWrongHost
	case !s.Scope.Includes(want):
		return fmt.Errorf("%w: %s session, %s needed", ErrSessionScope, s.Scope, want)
	}
	return nil
}

// Touched returns s used at now, and whether that is worth storing: only
// when its last use is a minute or more old, so that a burst of requests
// writes once.
func (s *Session) Touched(now time.Time) (Session, bool) {
	t := *s
	if now.Sub(s.LastSeenAt) < sessionTouchEvery {
		return t, false
	}
	t.LastSeenAt = now
	return t, true
}

// FormatSessionToken returns the token of session id with key:
// SessionPrefix, the ID, "_" and the key in unpadded base64url.
func FormatSessionToken(id SessionID, key []byte) string {
	return formatSecret(SessionPrefix, string(id), key)
}

// ParseSessionToken splits a token written by FormatSessionToken into the
// session's ID and key. Its errors never contain the token.
func ParseSessionToken(token string) (SessionID, []byte, error) {
	parts, err := parseSecret(SessionPrefix, token, ErrMalformedSession)
	if err != nil {
		return "", nil, err
	}
	return SessionID(parts.id), parts.key, nil
}
