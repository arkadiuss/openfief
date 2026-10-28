package app

import (
	"context"
	"errors"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
)

// Errors the AuthStore returns, wrapped.
var (
	// ErrNoOwner reports an installation where no passkey was ever created.
	ErrNoOwner = errors.New("no owner yet")
	// ErrUnknownPasskey reports a credential ID that is not stored.
	ErrUnknownPasskey = errors.New("unknown passkey")
	// ErrLastPasskey reports removing the only passkey left: the owner could
	// then only get back in with core enroll-link.
	ErrLastPasskey = errors.New("the last passkey can't be removed")
	// ErrUnknownCeremony reports a ceremony ID that is not stored: never
	// begun, already finished, or cleaned up after it expired.
	ErrUnknownCeremony = errors.New("unknown ceremony")
	// ErrUnknownEnrollmentCode reports an enrollment code ID that is not
	// stored.
	ErrUnknownEnrollmentCode = errors.New("unknown enrollment code")
	// ErrUnknownSession reports a session ID that is not stored.
	ErrUnknownSession = errors.New("unknown session")
)

// AuthStore keeps the owner, passkeys, ceremonies, enrollment codes and
// sessions. Enrollment codes and sessions are stored with the SHA-256 hash
// of their key, never the key. *identity/postgres.Store implements it.
type AuthStore interface {
	OwnerStore
	PasskeyStore
	CeremonyStore
	EnrollmentStore
	SessionStore
}

// OwnerStore keeps the owner.
type OwnerStore interface {
	// Owner returns the owner, or an error wrapping ErrNoOwner.
	Owner(ctx context.Context) (identity.Owner, error)
	// EnsureOwner stores candidate as the owner unless there is one, and
	// returns the stored owner.
	EnsureOwner(ctx context.Context, candidate identity.Owner) (identity.Owner, error)
}

// PasskeyStore keeps passkeys.
type PasskeyStore interface {
	// SavePasskey stores a new passkey.
	SavePasskey(ctx context.Context, p *identity.Passkey) error
	// Passkeys returns every passkey, in creation then ID order.
	Passkeys(ctx context.Context) ([]identity.Passkey, error)
	// UpdatePasskeyUse stores p's sign count, backup state and last use.
	UpdatePasskeyUse(ctx context.Context, p *identity.Passkey) error
	// DeletePasskey deletes the passkey with credential ID id, failing with
	// an error wrapping ErrUnknownPasskey, or ErrLastPasskey when it is the
	// only one, atomically.
	DeletePasskey(ctx context.Context, id []byte) error
}

// CeremonyStore keeps WebAuthn ceremonies between begin and finish.
type CeremonyStore interface {
	// SaveCeremony stores a ceremony, deleting those expired at its start.
	SaveCeremony(ctx context.Context, c *identity.Ceremony) error
	// TakeCeremony deletes and returns the ceremony id, so it is used once,
	// or fails with an error wrapping ErrUnknownCeremony.
	TakeCeremony(ctx context.Context, id string) (identity.Ceremony, error)
}

// EnrollmentStore keeps enrollment codes.
type EnrollmentStore interface {
	// SaveEnrollmentCode stores a new code with its key's hash.
	SaveEnrollmentCode(ctx context.Context, c *identity.EnrollmentCode, keyHash []byte) error
	// EnrollmentCode returns the code id and its key's hash, or an error
	// wrapping ErrUnknownEnrollmentCode.
	EnrollmentCode(ctx context.Context, id identity.EnrollmentCodeID) (identity.EnrollmentCode, []byte, error)
	// UseEnrollmentCode marks the code id used at at, failing with an error
	// wrapping identity.ErrEnrollmentCodeUsed when it already was,
	// atomically, so a code is used once even under concurrent calls.
	UseEnrollmentCode(ctx context.Context, id identity.EnrollmentCodeID, at time.Time) error
}

// SessionStore keeps sessions.
type SessionStore interface {
	// SaveSession stores a new session with its key's hash.
	SaveSession(ctx context.Context, s *identity.Session, keyHash []byte) error
	// Session returns the session id and its key's hash, or an error
	// wrapping ErrUnknownSession.
	Session(ctx context.Context, id identity.SessionID) (identity.Session, []byte, error)
	// TouchSession stores at as the session's last use.
	TouchSession(ctx context.Context, id identity.SessionID, at time.Time) error
	// RevokeSession marks the session revoked at at, keeping an earlier
	// revocation.
	RevokeSession(ctx context.Context, id identity.SessionID, at time.Time) error
}

// Ceremonies runs the WebAuthn side of passkey ceremonies for one origin at
// a time (D-26, D-29: a vetted library, no home-grown cryptography).
// *identity/webauthn.Ceremonies implements it. Every error from a Finish
// method that a bad credential causes wraps ErrCredentialRefused.
type Ceremonies interface {
	// BeginRegistration returns the creation options for the browser and
	// the state to keep, for owner on origin's host, excluding the passkeys
	// the owner has there already.
	BeginRegistration(origin identity.Origin, owner *identity.Owner, existing []identity.Passkey) (Begun, error)
	// FinishRegistration verifies credential against state and returns the
	// new passkey, without a name or creation time.
	FinishRegistration(origin identity.Origin, owner *identity.Owner, state, credential []byte) (identity.Passkey, error)
	// BeginLogin returns the request options for a discoverable-credential
	// login on origin's host and the state to keep.
	BeginLogin(origin identity.Origin) (Begun, error)
	// FinishLogin verifies the assertion credential against state, the
	// owner's handle and passkeys, and returns what it proved.
	FinishLogin(origin identity.Origin, owner *identity.Owner, passkeys []identity.Passkey, state, credential []byte) (Assertion, error)
}

// Begun is a ceremony's begin step: the options for the browser, and the
// relying party's state to keep until the finish step.
type Begun struct {
	Options []byte
	State   []byte
}

// Assertion is what a verified login assertion carries.
type Assertion struct {
	CredentialID []byte
	SignCount    uint32
	BackupState  bool
}

// ErrCredentialRefused reports a WebAuthn credential or assertion that does
// not verify.
var ErrCredentialRefused = errors.New("credential refused")

// DemoMarker reports whether the database is a demo database: the marker
// the demo loader writes (cmd/demo-seed). The identity context reads it only
// through this port.
type DemoMarker interface {
	DemoMarked(ctx context.Context) (bool, error)
}
