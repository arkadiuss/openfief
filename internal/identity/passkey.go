package identity

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"time"
	"unicode"
	"unicode/utf8"
)

// Errors about the owner and passkeys.
var (
	// ErrInvalidOwner reports an owner whose user handle is not
	// OwnerHandleBytes bytes.
	ErrInvalidOwner = errors.New("invalid owner")
	// ErrInvalidPasskey reports a passkey without a credential ID, public
	// key or host, or with a bad name.
	ErrInvalidPasskey = errors.New("invalid passkey")
	// ErrPasskeyWrongHost reports a passkey used on another host than the
	// one it was created on (D-60).
	ErrPasskeyWrongHost = errors.New("passkey belongs to another host")
	// ErrSignCountRegressed reports an assertion whose signature counter is
	// not above the stored one: a sign that the authenticator was cloned.
	ErrSignCountRegressed = errors.New("passkey signature counter went backwards")
)

const (
	// OwnerHandleBytes is the length of the owner's random WebAuthn user
	// handle.
	OwnerHandleBytes = 32
	// maxPasskeyNameLen is the longest passkey name, in characters.
	maxPasskeyNameLen = 64
	// maxCredentialIDLen is the longest WebAuthn credential ID.
	maxCredentialIDLen = 1023
)

// Owner is the one person an installation has (GR-2). Its user handle, the
// WebAuthn user ID every passkey carries, is random and never shown.
type Owner struct {
	Handle    []byte
	CreatedAt time.Time
}

// NewOwner returns the owner with handle, created at now.
func NewOwner(handle []byte, now time.Time) (Owner, error) {
	if len(handle) != OwnerHandleBytes {
		return Owner{}, fmt.Errorf("%w: want a %d-byte handle", ErrInvalidOwner, OwnerHandleBytes)
	}
	return Owner{Handle: slices.Clone(handle), CreatedAt: now}, nil
}

// Passkey is one of the owner's WebAuthn credentials (D-26). It works only
// on the host it was created on, its relying party ID (D-60).
type Passkey struct {
	CredentialID []byte
	Name         string
	RPID         string
	// PublicKey is the credential's COSE public key.
	PublicKey         []byte
	AttestationType   string
	AttestationFormat string
	AAGUID            []byte
	SignCount         uint32
	Transports        []string
	UserVerified      bool
	BackupEligible    bool
	BackupState       bool
	CreatedAt         time.Time
	// LastUsedAt is the last login with it; zero if it was never used.
	LastUsedAt time.Time
}

// ID returns the credential ID in unpadded base64url, as the browser and
// the API show it.
func (p *Passkey) ID() string { return base64.RawURLEncoding.EncodeToString(p.CredentialID) }

// ParsePasskeyID decodes an ID written by Passkey.ID.
func ParsePasskeyID(id string) ([]byte, error) {
	b, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil || len(b) == 0 || len(b) > maxCredentialIDLen {
		return nil, fmt.Errorf("%w: bad ID", ErrInvalidPasskey)
	}
	return b, nil
}

// Validate checks that p has a credential ID, a public key, a host and a
// name of 1 to 64 printable characters.
func (p *Passkey) Validate() error {
	switch {
	case len(p.CredentialID) == 0 || len(p.CredentialID) > maxCredentialIDLen:
		return fmt.Errorf("%w: bad credential ID", ErrInvalidPasskey)
	case len(p.PublicKey) == 0:
		return fmt.Errorf("%w: no public key", ErrInvalidPasskey)
	case !validHost(p.RPID):
		return fmt.Errorf("%w: bad host", ErrInvalidPasskey)
	}
	return CheckPasskeyName(p.Name)
}

// CheckPasskeyName checks a passkey name: 1 to 64 printable characters.
func CheckPasskeyName(name string) error {
	if name == "" || utf8.RuneCountInString(name) > maxPasskeyNameLen || !utf8.ValidString(name) {
		return fmt.Errorf("%w: want a name of 1 to %d characters", ErrInvalidPasskey, maxPasskeyNameLen)
	}
	for _, r := range name {
		if !unicode.IsPrint(r) {
			return fmt.Errorf("%w: the name has unprintable characters", ErrInvalidPasskey)
		}
	}
	return nil
}

// CheckHost returns ErrPasskeyWrongHost unless p belongs to host rpID.
func (p *Passkey) CheckHost(rpID string) error {
	if p.RPID != rpID {
		return fmt.Errorf("%w: %s", ErrPasskeyWrongHost, p.RPID)
	}
	return nil
}

// Used returns p after a login at now whose assertion carried signCount and
// backupState. Following WebAuthn's signature counter rule, a counter that
// is not above the stored one fails with ErrSignCountRegressed, unless both
// are zero (authenticators that keep no counter, such as synced passkeys).
func (p *Passkey) Used(signCount uint32, backupState bool, now time.Time) (Passkey, error) {
	if (signCount != 0 || p.SignCount != 0) && signCount <= p.SignCount {
		return Passkey{}, fmt.Errorf("%w: stored %d, got %d", ErrSignCountRegressed, p.SignCount, signCount)
	}
	q := *p
	q.CredentialID = slices.Clone(p.CredentialID)
	q.SignCount, q.BackupState, q.LastUsedAt = signCount, backupState, now
	return q, nil
}

// SameCredential reports whether p and q are the same credential.
func (p *Passkey) SameCredential(q *Passkey) bool { return bytes.Equal(p.CredentialID, q.CredentialID) }
