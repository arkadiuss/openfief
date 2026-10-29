package identity

import (
	"errors"
	"fmt"
	"time"
)

// Errors about enrollment codes and ceremonies.
var (
	// ErrMalformedEnrollmentCode reports text that is not an enrollment code.
	ErrMalformedEnrollmentCode = errors.New("malformed enrollment code")
	// ErrEnrollmentCodeUsed reports an enrollment code that was used.
	ErrEnrollmentCodeUsed = errors.New("enrollment code already used")
	// ErrEnrollmentCodeExpired reports an enrollment code used at or after
	// its expiry.
	ErrEnrollmentCodeExpired = errors.New("enrollment code expired")
	// ErrEnrollmentCodeWrongOrigin reports an enrollment code used on
	// another origin than the one it was issued for.
	ErrEnrollmentCodeWrongOrigin = errors.New("enrollment code issued for another origin")
	// ErrInvalidEnrollmentTTL reports a lifetime that is not positive or
	// exceeds MaxEnrollmentTTL.
	ErrInvalidEnrollmentTTL = errors.New("invalid enrollment link lifetime")
	// ErrCeremonyExpired reports a WebAuthn ceremony finished at or after
	// its expiry.
	ErrCeremonyExpired = errors.New("ceremony expired")
	// ErrCeremonyMismatch reports a ceremony finished as the other kind, or
	// on another origin than it began on.
	ErrCeremonyMismatch = errors.New("ceremony does not match")
)

const (
	// EnrollmentPrefix starts every enrollment code.
	EnrollmentPrefix = "ofenroll_"
	// EnrollPath is the built-in UI's enrollment page; the code follows in
	// the URL fragment.
	EnrollPath = "/enroll"
	// DefaultEnrollmentTTL is how long an enrollment link lasts by default.
	DefaultEnrollmentTTL = 15 * time.Minute
	// MaxEnrollmentTTL is the longest an enrollment link may last.
	MaxEnrollmentTTL = 24 * time.Hour
	// CeremonyTTL is how long a WebAuthn ceremony may take, begin to finish.
	CeremonyTTL = 5 * time.Minute
	// CeremonyIDBytes is the number of random bytes in a ceremony ID.
	CeremonyIDBytes = 16
)

// EnrollmentCodeID identifies an enrollment code. It is not secret.
type EnrollmentCodeID string

// EnrollmentCode is a one-time permission to create a passkey on one
// origin (D-60): the CLI issues one for localhost, and an owner session
// one for a new host. Only the hash of its key is stored.
type EnrollmentCode struct {
	ID        EnrollmentCodeID
	Origin    Origin
	CreatedAt time.Time
	ExpiresAt time.Time
	// UsedAt is when a passkey was created with it; zero while unused.
	UsedAt time.Time
}

// NewEnrollmentCode returns a code with id for origin, created at now and
// lasting ttl, which must be positive and at most MaxEnrollmentTTL.
func NewEnrollmentCode(id EnrollmentCodeID, origin Origin, now time.Time, ttl time.Duration) (EnrollmentCode, error) {
	if !isHexID(string(id)) {
		return EnrollmentCode{}, fmt.Errorf("%w: bad ID", ErrMalformedEnrollmentCode)
	}
	if origin.IsZero() {
		return EnrollmentCode{}, fmt.Errorf("%w: no origin", ErrInvalidOrigin)
	}
	if ttl <= 0 || ttl > MaxEnrollmentTTL {
		return EnrollmentCode{}, fmt.Errorf("%w: %s, want more than 0 and at most %s", ErrInvalidEnrollmentTTL, ttl, MaxEnrollmentTTL)
	}
	return EnrollmentCode{ID: id, Origin: origin, CreatedAt: now, ExpiresAt: now.Add(ttl)}, nil
}

// Check returns nil when c may be used at now on origin. Otherwise, in this
// order: ErrEnrollmentCodeUsed, ErrEnrollmentCodeExpired,
// ErrEnrollmentCodeWrongOrigin.
func (c *EnrollmentCode) Check(now time.Time, origin Origin) error {
	switch {
	case !c.UsedAt.IsZero():
		return ErrEnrollmentCodeUsed
	case !now.Before(c.ExpiresAt):
		return ErrEnrollmentCodeExpired
	case origin != c.Origin:
		return ErrEnrollmentCodeWrongOrigin
	}
	return nil
}

// FormatEnrollmentCode returns the code id with key: EnrollmentPrefix, the
// ID, "_" and the key in unpadded base64url.
func FormatEnrollmentCode(id EnrollmentCodeID, key []byte) string {
	return formatSecret(EnrollmentPrefix, string(id), key)
}

// ParseEnrollmentCode splits a code written by FormatEnrollmentCode into
// its ID and key. Its errors never contain the code.
func ParseEnrollmentCode(code string) (EnrollmentCodeID, []byte, error) {
	parts, err := parseSecret(EnrollmentPrefix, code, ErrMalformedEnrollmentCode)
	if err != nil {
		return "", nil, err
	}
	return EnrollmentCodeID(parts.id), parts.key, nil
}

// EnrollmentURL returns the enrollment link for code on origin:
// origin, EnrollPath and the code in the fragment, which browsers never
// send to a server.
func EnrollmentURL(origin Origin, code string) string {
	return origin.String() + EnrollPath + "#code=" + code
}

// CeremonyKind is which WebAuthn ceremony a Ceremony is.
type CeremonyKind int

// The ceremony kinds. The zero CeremonyKind is none.
const (
	// CeremonyRegistration creates a passkey.
	CeremonyRegistration CeremonyKind = iota + 1
	// CeremonyLogin logs in with one.
	CeremonyLogin
)

// String returns "registration" or "login".
func (k CeremonyKind) String() string {
	switch k {
	case CeremonyRegistration:
		return "registration"
	case CeremonyLogin:
		return "login"
	}
	return fmt.Sprintf("CeremonyKind(%d)", int(k))
}

// ParseCeremonyKind parses a kind as String writes it.
func ParseCeremonyKind(s string) (CeremonyKind, error) {
	for _, k := range []CeremonyKind{CeremonyRegistration, CeremonyLogin} {
		if s == k.String() {
			return k, nil
		}
	}
	return 0, fmt.Errorf("%w: unknown kind %q", ErrCeremonyMismatch, s)
}

// Ceremony is a WebAuthn ceremony between its begin and finish steps: the
// relying party's state (the challenge and what it expects), kept by the
// core so the browser can't change it, and used once.
type Ceremony struct {
	ID     string
	Kind   CeremonyKind
	Origin Origin
	// State is the WebAuthn library's ceremony state, opaque here.
	State []byte
	// EnrollmentCode is the code a registration began with; empty when an
	// owner session began it.
	EnrollmentCode EnrollmentCodeID
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

// NewCeremony returns a ceremony of kind on origin with state, begun at now
// and lasting CeremonyTTL.
func NewCeremony(id string, kind CeremonyKind, origin Origin, state []byte, now time.Time) Ceremony {
	return Ceremony{ID: id, Kind: kind, Origin: origin, State: state, CreatedAt: now, ExpiresAt: now.Add(CeremonyTTL)}
}

// Check returns nil when c may be finished at now as kind on origin;
// otherwise ErrCeremonyExpired or ErrCeremonyMismatch.
func (c *Ceremony) Check(now time.Time, kind CeremonyKind, origin Origin) error {
	switch {
	case !now.Before(c.ExpiresAt):
		return ErrCeremonyExpired
	case c.Kind != kind:
		return fmt.Errorf("%w: a %s ceremony finished as %s", ErrCeremonyMismatch, c.Kind, kind)
	case c.Origin != origin:
		return fmt.Errorf("%w: begun on another origin", ErrCeremonyMismatch)
	}
	return nil
}
