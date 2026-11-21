package app

import (
	"context"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/thesis"
)

// Theses are the draft theses use cases (D-229), each checked against the
// owner's session: a session that is not the owner's reads and writes
// nothing.
type Theses struct {
	store  Store
	owners Owners
	audit  Auditor
	ids    IDs
	now    func() time.Time
}

// New returns the use cases over store, with the owner's session checked by
// owners, every save and edit entered in the audit log through audit, new
// thesis IDs made by ids and the time taken from now.
func New(store Store, owners Owners, audit Auditor, ids IDs, now func() time.Time) *Theses {
	return &Theses{store: store, owners: owners, audit: audit, ids: ids, now: now}
}

// Save stores draft as a new thesis, version 1, and returns it. A draft that
// breaks a rule is an error wrapping thesis.ErrInvalid.
func (t *Theses) Save(ctx context.Context, origin, sessionToken string, draft *thesis.Draft) (thesis.Thesis, error) {
	session, err := t.owner(ctx, origin, sessionToken)
	if err != nil {
		return thesis.Thesis{}, err
	}
	d, err := thesis.New(draft)
	if err != nil {
		return thesis.Thesis{}, fmt.Errorf("saving a draft thesis: %w", err)
	}
	id, err := t.ids()
	if err != nil {
		return thesis.Thesis{}, fmt.Errorf("making a thesis ID: %w", err)
	}
	if err := t.store.Save(ctx, id, 0, t.now().UTC(), &d, t.recording(session, id, 1, false)); err != nil {
		return thesis.Thesis{}, fmt.Errorf("saving a draft thesis: %w", err)
	}
	return t.Get(ctx, origin, sessionToken, id)
}

// Edit stores draft as the version after base of the thesis id, keeping the
// earlier versions, and returns the thesis. base is the version the owner
// edited; when the thesis has a newer one the error wraps thesis.ErrStale and
// nothing is stored. A draft that breaks a rule wraps thesis.ErrInvalid, and
// a thesis that does not exist thesis.ErrNotFound.
func (t *Theses) Edit(ctx context.Context, origin, sessionToken, id string, base int, draft *thesis.Draft) (thesis.Thesis, error) {
	session, err := t.owner(ctx, origin, sessionToken)
	if err != nil {
		return thesis.Thesis{}, err
	}
	if !validID(id) || base < 1 {
		return thesis.Thesis{}, fmt.Errorf("editing a draft thesis: %w", thesis.ErrNotFound)
	}
	d, err := thesis.New(draft)
	if err != nil {
		return thesis.Thesis{}, fmt.Errorf("editing a draft thesis: %w", err)
	}
	if err := t.store.Save(ctx, id, base, t.now().UTC(), &d, t.recording(session, id, base+1, true)); err != nil {
		return thesis.Thesis{}, fmt.Errorf("editing a draft thesis: %w", err)
	}
	return t.Get(ctx, origin, sessionToken, id)
}

// Get returns the thesis id with every version, or an error wrapping
// thesis.ErrNotFound.
func (t *Theses) Get(ctx context.Context, origin, sessionToken, id string) (thesis.Thesis, error) {
	if _, err := t.owner(ctx, origin, sessionToken); err != nil {
		return thesis.Thesis{}, err
	}
	if !validID(id) {
		return thesis.Thesis{}, fmt.Errorf("reading a draft thesis: %w", thesis.ErrNotFound)
	}
	got, err := t.store.Get(ctx, id)
	if err != nil {
		return thesis.Thesis{}, fmt.Errorf("reading a draft thesis: %w", err)
	}
	return got, nil
}

// List returns the draft theses as their latest versions, the most recently
// changed first.
func (t *Theses) List(ctx context.Context, origin, sessionToken string) ([]thesis.Summary, error) {
	if _, err := t.owner(ctx, origin, sessionToken); err != nil {
		return nil, err
	}
	got, err := t.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing the draft theses: %w", err)
	}
	return got, nil
}

// owner checks that the session on origin is the owner's and returns its ID.
func (t *Theses) owner(ctx context.Context, origin, sessionToken string) (string, error) {
	id, err := t.owners.OwnerSession(ctx, origin, sessionToken)
	if err != nil {
		return "", fmt.Errorf("the draft theses: %w", err)
	}
	return id, nil
}

// recording returns the hook that writes the audit entry of a change inside
// the transaction that stores it.
func (t *Theses) recording(session, id string, version int, edited bool) func(context.Context) error {
	return func(ctx context.Context) error {
		c := &Change{At: t.now().UTC(), Session: session, Thesis: id, Version: version, Edited: edited}
		if err := t.audit.RecordChange(ctx, c); err != nil {
			return fmt.Errorf("recording a change of a draft thesis: %w", err)
		}
		return nil
	}
}

// idLen is the length of a thesis ID: 16 hex digits.
const idLen = 16

// validID reports whether id has the shape of a thesis ID, so a malformed one
// is not found without asking the store.
func validID(id string) bool {
	if len(id) != idLen {
		return false
	}
	for _, c := range id {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
