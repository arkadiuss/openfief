// Package postgres stores attention items in PostgreSQL, in the attention
// schema that Migrate creates (docs/DESIGN.md D-130): the items, their
// append-only history, and the core's transactional outbox, whose event
// for a change is written in the same database transaction as the change.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/app"
)

// Topic is the outbox topic of an attention item's change.
const Topic = "openfief.attention.v1.AttentionItemChanged"

// changeLock is the advisory lock key that serializes changes to the
// items: "attnch" in ASCII.
const changeLock int64 = 0x6174746e6368

// ErrBadRow reports a stored row this package can't read.
var ErrBadRow = errors.New("bad attention row")

// Store is app.Store and app.Outbox in PostgreSQL. New returns one over a
// pool, where each call commits on its own; InTx passes one whose calls
// all run in one transaction.
type Store struct {
	db db
	tx pgx.Tx // nil outside InTx
}

// db is what Store runs its statements on: a *pgxpool.Pool or a pgx.Tx.
type db interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

var (
	_ app.Store  = (*Store)(nil)
	_ app.Outbox = (*Store)(nil)
)

// New returns the store over pool. Run Migrate first.
func New(pool *pgxpool.Pool) *Store { return &Store{db: pool} }

// InTx runs fn with a Store whose calls all run in one database
// transaction, committed when fn returns nil and rolled back otherwise.
// Inside InTx it runs fn in the same transaction.
func (s *Store) InTx(ctx context.Context, fn func(tx app.Store) error) error {
	if s.tx != nil {
		return fn(s)
	}
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error { return fn(&Store{db: tx, tx: tx}) })
	if err != nil {
		return fmt.Errorf("attention transaction: %w", err)
	}
	return nil
}

// Lock takes the change lock until the transaction ends; outside InTx it
// ends at once.
func (s *Store) Lock(ctx context.Context) error {
	if _, err := s.db.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", changeLock); err != nil {
		return fmt.Errorf("locking attention items: %w", err)
	}
	return nil
}

// Items returns every item of kind, or of every kind when kind is empty,
// in ID order, each with its history.
func (s *Store) Items(ctx context.Context, kind attention.Kind) ([]attention.Item, error) {
	return s.items(ctx, `WHERE ($1 = '' OR i.kind = $1)`, string(kind))
}

// Item returns the item id with its history; ok is false when there is
// none.
func (s *Store) Item(ctx context.Context, id int64) (attention.Item, bool, error) {
	items, err := s.items(ctx, `WHERE i.id = $1`, id)
	if err != nil || len(items) == 0 {
		return attention.Item{}, false, err
	}
	return items[0], true, nil
}

// Append stores step: a new item when step.ItemID is zero, then its change
// and the change's outbox event, all in one database transaction (InTx's,
// or its own). It returns the item's ID.
func (s *Store) Append(ctx context.Context, step *attention.Step) (int64, error) {
	if s.tx != nil {
		return s.append(ctx, step)
	}
	var id int64
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		var err error
		id, err = (&Store{db: tx, tx: tx}).append(ctx, step)
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("attention transaction: %w", err)
	}
	return id, nil
}

// payload is an outbox event's payload: AttentionItemChanged's JSON form
// without its event_id, which is the outbox row's ID.
type payload struct {
	ItemID  int64  `json:"itemId,string"`
	Kind    string `json:"kind"`
	Subject string `json:"subject"`
	Seq     int    `json:"seq"`
	State   string `json:"state"`
	Reason  string `json:"reason"`
	Detail  string `json:"detail"`
	At      string `json:"at"`
}

// eventPayload returns the payload of step's change to item id. Its
// subject is the one read access may see (attention.ReadSubject): an
// unmapped broker account's event names only its connector instance, never
// the broker account number (D-121, D-165). The item keeps its full
// subject for the owner.
func eventPayload(id int64, step *attention.Step) payload {
	c := &step.Change
	return payload{
		ItemID: id, Kind: string(step.Kind), Subject: attention.ReadSubject(step.Kind, step.Subject), Seq: c.Seq, State: c.State.String(),
		Reason: c.Reason.String(), Detail: c.Detail, At: c.At.UTC().Format(time.RFC3339),
	}
}

// Unpublished returns up to limit events not marked published, in ID
// order.
func (s *Store) Unpublished(ctx context.Context, limit int) ([]app.Event, error) {
	rows, err := s.db.Query(ctx, `
		SELECT o.id, o.payload FROM attention.outbox o
		WHERE NOT EXISTS (SELECT 1 FROM attention.published p WHERE p.event_id = o.id)
		ORDER BY o.id LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("reading the outbox: %w", err)
	}
	events, err := pgx.CollectRows(rows, scanEvent)
	if err != nil {
		return nil, fmt.Errorf("reading the outbox: %w", err)
	}
	return events, nil
}

// scanEvent reads one outbox row.
func scanEvent(row pgx.CollectableRow) (app.Event, error) {
	var (
		id  int64
		raw []byte
		p   payload
	)
	if err := row.Scan(&id, &raw); err != nil {
		return app.Event{}, fmt.Errorf("scanning an event: %w", err)
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return app.Event{}, fmt.Errorf("%w: event %d: %w", ErrBadRow, id, err)
	}
	state, err := parseState(p.State)
	if err != nil {
		return app.Event{}, err
	}
	reason, err := parseReason(p.Reason)
	if err != nil {
		return app.Event{}, err
	}
	at, err := time.Parse(time.RFC3339, p.At)
	if err != nil {
		return app.Event{}, fmt.Errorf("%w: event %d: %w", ErrBadRow, id, err)
	}
	return app.Event{
		ID: id, ItemID: p.ItemID, Kind: attention.Kind(p.Kind), Subject: p.Subject,
		Change: attention.Change{Seq: p.Seq, State: state, Reason: reason, Detail: p.Detail, At: at.UTC()},
	}, nil
}

// MarkPublished records that event id was published at at; marking it
// again keeps the first record.
func (s *Store) MarkPublished(ctx context.Context, id int64, at time.Time) error {
	if _, err := s.db.Exec(ctx, `
		INSERT INTO attention.published (event_id, published_at) VALUES ($1, $2) ON CONFLICT (event_id) DO NOTHING`,
		id, at); err != nil {
		return fmt.Errorf("marking event %d published: %w", id, err)
	}
	return nil
}

// items returns the items the where clause, over arg, selects.
func (s *Store) items(ctx context.Context, where string, arg any) ([]attention.Item, error) {
	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.kind, i.subject, c.seq, c.state, c.reason, c.detail, c.at
		FROM attention.items i JOIN attention.changes c ON c.item_id = i.id `+where+`
		ORDER BY i.id, c.seq`, arg)
	if err != nil {
		return nil, fmt.Errorf("loading attention items: %w", err)
	}
	defer rows.Close()
	var out []attention.Item
	for rows.Next() {
		var (
			id                    int64
			kind, subject         string
			state, reason, detail string
			c                     attention.Change
		)
		if scanErr := rows.Scan(&id, &kind, &subject, &c.Seq, &state, &reason, &detail, &c.At); scanErr != nil {
			return nil, fmt.Errorf("scanning an attention item: %w", scanErr)
		}
		if c.State, err = parseState(state); err != nil {
			return nil, err
		}
		if c.Reason, err = parseReason(reason); err != nil {
			return nil, err
		}
		c.Detail, c.At = detail, c.At.UTC()
		if len(out) == 0 || out[len(out)-1].ID != id {
			out = append(out, attention.Item{ID: id, Kind: attention.Kind(kind), Subject: subject})
		}
		out[len(out)-1].Changes = append(out[len(out)-1].Changes, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("loading attention items: %w", err)
	}
	return out, nil
}

// append is Append inside a transaction.
func (s *Store) append(ctx context.Context, step *attention.Step) (int64, error) {
	c := &step.Change
	id := step.ItemID
	if id == 0 {
		if err := s.db.QueryRow(ctx, `
			INSERT INTO attention.items (kind, subject, created_at) VALUES ($1, $2, $3) RETURNING id`,
			string(step.Kind), step.Subject, c.At).Scan(&id); err != nil {
			return 0, fmt.Errorf("storing attention item %s %q: %w", step.Kind, step.Subject, err)
		}
	}
	if _, err := s.db.Exec(ctx, `
		INSERT INTO attention.changes (item_id, seq, state, reason, detail, at) VALUES ($1, $2, $3, $4, $5, $6)`,
		id, c.Seq, c.State.String(), c.Reason.String(), c.Detail, c.At); err != nil {
		return 0, fmt.Errorf("storing change %d of attention item %d: %w", c.Seq, id, err)
	}
	payload, err := json.Marshal(eventPayload(id, step))
	if err != nil {
		return 0, fmt.Errorf("encoding the event of attention item %d: %w", id, err)
	}
	if _, err := s.db.Exec(ctx, `
		INSERT INTO attention.outbox (topic, item_id, seq, payload, created_at) VALUES ($1, $2, $3, $4, $5)`,
		Topic, id, c.Seq, payload, c.At); err != nil {
		return 0, fmt.Errorf("storing the event of attention item %d: %w", id, err)
	}
	return id, nil
}

// parseState reads a stored state.
func parseState(s string) (attention.State, error) {
	for _, st := range []attention.State{attention.StateOpen, attention.StateAcknowledged, attention.StateResolved} {
		if st.String() == s {
			return st, nil
		}
	}
	return 0, fmt.Errorf("%w: state %q", ErrBadRow, s)
}

// parseReason reads a stored reason.
func parseReason(s string) (attention.Reason, error) {
	for _, r := range []attention.Reason{
		attention.ReasonRaised, attention.ReasonChanged, attention.ReasonReopened,
		attention.ReasonCleared, attention.ReasonAcknowledged, attention.ReasonResolved,
	} {
		if r.String() == s {
			return r, nil
		}
	}
	return 0, fmt.Errorf("%w: reason %q", ErrBadRow, s)
}
