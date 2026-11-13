package identity

import (
	"errors"
	"fmt"
	"slices"
	"time"
	"unicode"
	"unicode/utf8"
)

// ErrInvalidPerson reports a person without a valid ID, kind or name, or a
// viewer without a WebAuthn user handle.
var ErrInvalidPerson = errors.New("invalid person")

// OwnerPersonID is the owner's ID among the people: the one person an
// installation always has (D-65), the subject SpiceDB knows as person:owner.
// A viewer's ID is IDBytes random bytes in hex, so no viewer can take it.
const OwnerPersonID = "owner"

// maxPersonNameLen is the longest name of a person, in characters.
const maxPersonNameLen = 64

// PersonKind says whether a person is the owner or a viewer.
type PersonKind string

// The kinds of person.
const (
	// PersonOwner holds every account and alone changes anything (D-65).
	PersonOwner PersonKind = "owner"
	// PersonViewer is someone the owner invited to view the accounts they
	// grant them (D-65, D-209): never more.
	PersonViewer PersonKind = "viewer"
)

// Person is someone with passkeys of their own: the owner, and each viewer
// the owner enrolls (D-65). A person is never deleted: a revoked viewer
// keeps their ID, and with it the history that names them (Q8, D-211).
type Person struct {
	// ID is OwnerPersonID, or a viewer's random ID.
	ID   string
	Kind PersonKind
	// Name is what the owner calls them; the owner's is "Owner". It is
	// shown to the owner and to the person, never written to the audit log,
	// which names people by ID.
	Name string
	// Handle is the WebAuthn user handle their passkeys carry: random and
	// never shown.
	Handle    []byte
	CreatedAt time.Time
	// RevokedAt is when the owner revoked a viewer; zero while they are
	// active, and always for the owner.
	RevokedAt time.Time
}

// OwnerPerson returns the owner as a person, named "Owner".
func OwnerPerson(o *Owner) Person {
	return Person{ID: OwnerPersonID, Kind: PersonOwner, Name: "Owner", Handle: slices.Clone(o.Handle), CreatedAt: o.CreatedAt}
}

// NewViewer returns the viewer with ID id, called name, whose passkeys carry
// handle, created at now.
func NewViewer(id, name string, handle []byte, now time.Time) (Person, error) {
	p := Person{ID: id, Kind: PersonViewer, Name: name, Handle: slices.Clone(handle), CreatedAt: now.UTC()}
	if err := p.Validate(); err != nil {
		return Person{}, err
	}
	return p, nil
}

// Validate checks that p's ID has the shape its kind's has, its name is 1 to
// 64 printable characters, and a viewer has a user handle of OwnerHandleBytes
// bytes.
func (p *Person) Validate() error {
	switch p.Kind {
	case PersonOwner:
		if p.ID != OwnerPersonID {
			return fmt.Errorf("%w: the owner's ID is %q", ErrInvalidPerson, OwnerPersonID)
		}
	case PersonViewer:
		if !isHexID(p.ID) {
			return fmt.Errorf("%w: a viewer's ID is %d random bytes in hex", ErrInvalidPerson, IDBytes)
		}
		if len(p.Handle) != OwnerHandleBytes {
			return fmt.Errorf("%w: want a %d-byte user handle", ErrInvalidPerson, OwnerHandleBytes)
		}
	default:
		return fmt.Errorf("%w: unknown kind", ErrInvalidPerson)
	}
	return CheckPersonName(p.Name)
}

// CheckPersonName checks a person's name: 1 to 64 printable characters.
func CheckPersonName(name string) error {
	if name == "" || utf8.RuneCountInString(name) > maxPersonNameLen || !utf8.ValidString(name) {
		return fmt.Errorf("%w: want a name of 1 to %d characters", ErrInvalidPerson, maxPersonNameLen)
	}
	for _, r := range name {
		if !unicode.IsPrint(r) {
			return fmt.Errorf("%w: the name has unprintable characters", ErrInvalidPerson)
		}
	}
	return nil
}

// IsViewer reports whether p is a viewer.
func (p *Person) IsViewer() bool { return p.Kind == PersonViewer }

// Active reports whether p may sign in: the owner always, a viewer until the
// owner revokes them.
func (p *Person) Active() bool { return p.RevokedAt.IsZero() }

// Revoked returns p, revoked at at: a revocation never precedes the person's
// own creation. The owner can't be revoked: p is returned as it is.
func (p *Person) Revoked(at time.Time) Person {
	out := *p
	if p.Kind != PersonViewer {
		return out
	}
	at = at.UTC()
	if at.Before(p.CreatedAt) {
		at = p.CreatedAt
	}
	out.RevokedAt = at
	return out
}

// GrantHolder returns p as the holder of grants (D-209).
func (p *Person) GrantHolder() GrantHolder { return GrantHolder{Kind: GrantHolderPerson, ID: p.ID} }

// Ref names p in an audit entry, by ID.
func (p *Person) Ref() Ref { return Ref{Kind: RefPerson, ID: p.ID} }

// ViewerInvite is what the owner chose for a viewer when they issued the
// enrollment link: the name they call them and the grants the viewer starts
// with, made when the link is used.
type ViewerInvite struct {
	Name    string
	Targets []GrantTarget
}

// Validate checks the invite's name and each target's shape.
func (v *ViewerInvite) Validate() error {
	if err := CheckPersonName(v.Name); err != nil {
		return err
	}
	for _, t := range v.Targets {
		if err := t.Validate(); err != nil {
			return err
		}
	}
	return nil
}
