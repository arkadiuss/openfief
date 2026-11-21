// Package app holds the thesis context's use cases (docs/DESIGN.md D-229):
// saving a draft thesis, editing it into a new version, reading one, its
// versions and the list. All of it is the owner's; no other caller, viewer,
// script token or view plugin included, reads or writes a draft. Ports are
// declared here, by their consumer.
package app

import (
	"context"
	"errors"
	"time"

	"github.com/arkadiuss/openfief/internal/thesis"
)

// Errors of the thesis use cases, besides those of the thesis package.
var (
	// ErrNoSession reports a session token that is missing, unknown, expired
	// or revoked.
	ErrNoSession = errors.New("no valid session")
	// ErrNotOwner reports a valid session without owner scope, or an origin
	// the core does not serve the owner on.
	ErrNotOwner = errors.New("only the owner may do this")
)

// Owners checks that a session is the owner's (D-66, D-79) and returns its
// ID, which the audit entry of a change names as its actor; the identity
// context answers it in cmd/core, with errors wrapping ErrNoSession or
// ErrNotOwner.
type Owners interface {
	OwnerSession(ctx context.Context, origin, sessionToken string) (string, error)
}

// Store keeps the draft theses and every version of each. *postgres.Store
// implements it.
type Store interface {
	// Save stores draft as version base+1 of thesis id; base 0 creates the
	// thesis, as its version 1. It stores nothing and returns an error
	// wrapping thesis.ErrStale when base is not the thesis's latest version
	// (a create finds the id taken), and thesis.ErrNotFound for an edit of
	// a thesis that does not exist. within runs inside the transaction that
	// stores the version, before it commits: the audit entry of the change,
	// and when it fails nothing is stored.
	Save(ctx context.Context, id string, base int, at time.Time, draft *thesis.Draft, within func(context.Context) error) error
	// Get returns the thesis with every version, or an error wrapping
	// thesis.ErrNotFound.
	Get(ctx context.Context, id string) (thesis.Thesis, error)
	// List returns every thesis as its latest version, the most recently
	// changed first.
	List(ctx context.Context) ([]thesis.Summary, error)
}

// Change is what the audit log keeps of a save or an edit (D-211, D-229):
// who, when, which thesis and which version. It holds no text (GR-12).
type Change struct {
	At time.Time
	// Session is the owner's session that made the change.
	Session string
	Thesis  string
	// Version is the version the change stored: 1 for a save.
	Version int
	// Edited is false for a save, which stores version 1, and true for an
	// edit, which stores a later one.
	Edited bool
}

// Auditor appends the entry of a change to the audit log, inside the
// transaction that stores it; cmd/core implements it over the identity
// context's log. A change whose entry can't be written is not made.
type Auditor interface {
	RecordChange(ctx context.Context, c *Change) error
}

// IDs makes a thesis's ID: 16 lower-case hex digits from a random source.
type IDs func() (string, error)
