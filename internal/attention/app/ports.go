// Package app holds the attention context's use cases (docs/DESIGN.md
// D-130, D-131): checks raising and clearing attention items in the store,
// the owner acknowledging and resolving them, and the outbox's events
// handed to a publisher. Ports are declared here, by their consumer.
package app

import (
	"context"
	"errors"
	"time"

	"github.com/arkadiuss/openfief/internal/attention"
)

// Errors of the attention use cases.
var (
	// ErrNoSession reports a session token that is missing, unknown,
	// expired or revoked.
	ErrNoSession = errors.New("no valid session")
	// ErrNotOwner reports a valid session without owner scope, or an origin
	// the core does not serve the owner on.
	ErrNotOwner = errors.New("only the owner may do this")
	// ErrUnknownItem reports an item ID with no stored item.
	ErrUnknownItem = errors.New("unknown attention item")
)

// Store is the append-only store of attention items. Append writes each
// change's outbox event in the same database transaction as the change
// (D-130).
type Store interface {
	// InTx runs fn with a Store whose calls all run in one database
	// transaction, committed when fn returns nil and rolled back otherwise.
	InTx(ctx context.Context, fn func(tx Store) error) error
	// Lock serializes changes to the items until the transaction ends.
	Lock(ctx context.Context) error
	// Items returns every item of kind, or of every kind when kind is
	// empty, in ID order, each with its history.
	Items(ctx context.Context, kind attention.Kind) ([]attention.Item, error)
	// Item returns the item id with its history; ok is false when there is
	// none.
	Item(ctx context.Context, id int64) (attention.Item, bool, error)
	// Append stores step: a new item when step.ItemID is zero, then its
	// change, and the change's outbox event. It returns the item's ID.
	Append(ctx context.Context, step *attention.Step) (int64, error)
}

// Event is one change to an item as the outbox holds it, numbered from 1
// in the order the changes were written. It is published as
// openfief.attention.v1.AttentionItemChanged.
type Event struct {
	ID      int64
	ItemID  int64
	Kind    attention.Kind
	Subject string
	Change  attention.Change
}

// Outbox reads the events not published yet and records publishing them.
type Outbox interface {
	// Unpublished returns up to limit events not marked published, in ID
	// order.
	Unpublished(ctx context.Context, limit int) ([]Event, error)
	// MarkPublished records that event id was published at at.
	MarkPublished(ctx context.Context, id int64, at time.Time) error
}

// Publisher hands an event to whatever acts on it: notifications today,
// agents in later milestones. A failed Publish is retried with the same
// event, so consumers deduplicate on its ID.
type Publisher interface {
	Publish(ctx context.Context, e *Event) error
}

// Owners checks that a caller is the owner (D-79).
type Owners interface {
	// CheckOwner returns nil for an owner-scoped session token valid on
	// origin, an error wrapping ErrNoSession or ErrNotOwner when it is not,
	// and any other error when it can't tell.
	CheckOwner(ctx context.Context, origin, sessionToken string) error
}

// Check is one check that needs the owner: Find returns every condition
// of Kind it finds now, which Sweep reconciles with the stored items.
type Check struct {
	Kind attention.Kind
	Find func(ctx context.Context) ([]attention.Condition, error)
}
