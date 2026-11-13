package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

var _ app.PersonStore = (*Store)(nil)

// inviteTarget is a grant target as a viewer's enrollment link stores it, in
// the column viewer_grants.
type inviteTarget struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// personColumns are the columns scanPerson reads, in its order.
const personColumns = `id, kind, name, handle, created_at, revoked_at`

// scanPerson reads a row of personColumns.
func scanPerson(row pgx.Row) (identity.Person, error) {
	var (
		p       identity.Person
		kind    string
		revoked *time.Time
	)
	if err := row.Scan(&p.ID, &kind, &p.Name, &p.Handle, &p.CreatedAt, &revoked); err != nil {
		return identity.Person{}, fmt.Errorf("scanning a person: %w", err)
	}
	p.Kind = identity.PersonKind(kind)
	p.CreatedAt = p.CreatedAt.UTC()
	if revoked != nil {
		p.RevokedAt = revoked.UTC()
	}
	return p, nil
}

// Person returns the person id, or an error wrapping app.ErrUnknownPerson. The
// owner's handle is the owner's (identity.owner), which the caller fills in.
func (s *Store) Person(ctx context.Context, id string) (identity.Person, error) {
	p, err := scanPerson(s.db(ctx).QueryRow(ctx, `SELECT `+personColumns+` FROM identity.persons WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.Person{}, fmt.Errorf("%w: %s", app.ErrUnknownPerson, id)
	}
	if err != nil {
		return identity.Person{}, fmt.Errorf("loading person %s: %w", id, err)
	}
	return p, nil
}

// People returns every person, the owner first and then the viewers in the
// order they were created.
func (s *Store) People(ctx context.Context) ([]identity.Person, error) {
	rows, err := s.db(ctx).Query(ctx, `SELECT `+personColumns+` FROM identity.persons ORDER BY (kind = 'owner') DESC, created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("listing people: %w", err)
	}
	people, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (identity.Person, error) { return scanPerson(row) })
	if err != nil {
		return nil, fmt.Errorf("listing people: %w", err)
	}
	return people, nil
}

// SavePerson stores a new viewer.
func (s *Store) SavePerson(ctx context.Context, p *identity.Person) error {
	if err := p.Validate(); err != nil {
		return fmt.Errorf("saving a person: %w", err)
	}
	if !p.IsViewer() {
		return fmt.Errorf("saving a person: %w: only viewers are stored here", identity.ErrInvalidPerson)
	}
	_, err := s.db(ctx).Exec(ctx, `INSERT INTO identity.persons (id, kind, name, handle, created_at) VALUES ($1, 'viewer', $2, $3, $4)`,
		p.ID, p.Name, p.Handle, p.CreatedAt)
	if err != nil {
		return fmt.Errorf("storing person %s: %w", p.ID, err)
	}
	return nil
}

// RevokePerson revokes the active viewer id as of at: the person, their
// sessions and their passkeys, in one transaction, and returns the passkeys
// it disabled (in creation order) and how many sessions it ended. A viewer
// already revoked, the owner or an unknown ID fails with an error wrapping
// app.ErrUnknownPerson.
func (s *Store) RevokePerson(ctx context.Context, id string, at time.Time) ([]identity.Passkey, int, error) {
	var (
		keys     []identity.Passkey
		sessions int
	)
	err := s.Atomically(ctx, func(ctx context.Context) error {
		tx := s.db(ctx)
		tag, err := tx.Exec(ctx, `UPDATE identity.persons SET revoked_at = GREATEST($2, created_at)
			WHERE id = $1 AND kind = 'viewer' AND revoked_at IS NULL`, id, at)
		if err != nil {
			return fmt.Errorf("revoking person %s: %w", id, err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("%w: %s", app.ErrUnknownPerson, id)
		}
		rows, err := tx.Query(ctx, `UPDATE identity.passkeys SET disabled_at = GREATEST($2, created_at)
			WHERE person_id = $1 AND disabled_at IS NULL RETURNING `+webauthnColumns, id, at)
		if err != nil {
			return fmt.Errorf("disabling the passkeys of person %s: %w", id, err)
		}
		if keys, err = pgx.CollectRows(rows, scanPasskey); err != nil {
			return fmt.Errorf("disabling the passkeys of person %s: %w", id, err)
		}
		tag, err = tx.Exec(ctx, `UPDATE identity.sessions SET revoked_at = GREATEST($2, created_at)
			WHERE person_id = $1 AND revoked_at IS NULL`, id, at)
		if err != nil {
			return fmt.Errorf("ending the sessions of person %s: %w", id, err)
		}
		sessions = int(tag.RowsAffected())
		return nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("revoking a person: %w", err)
	}
	return keys, sessions, nil
}

// PasskeyCounts returns how many passkeys that work each person has, by
// person ID (the owner's is "owner").
func (s *Store) PasskeyCounts(ctx context.Context) (map[string]int, error) {
	return s.counts(ctx, `SELECT person_id, count(*) FROM identity.passkeys WHERE disabled_at IS NULL GROUP BY person_id`)
}

// ActiveSessionCounts returns how many sessions of each person are valid at
// at: not revoked, not past their absolute expiry or their idle timeout.
func (s *Store) ActiveSessionCounts(ctx context.Context, at time.Time) (map[string]int, error) {
	return s.counts(ctx, `SELECT person_id, count(*) FROM identity.sessions
		WHERE revoked_at IS NULL AND expires_at > $1 AND last_seen_at + make_interval(secs => idle_seconds) > $1 GROUP BY person_id`, at)
}

// counts runs a query of (person ID, count) rows.
func (s *Store) counts(ctx context.Context, query string, args ...any) (map[string]int, error) {
	rows, err := s.db(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("counting per person: %w", err)
	}
	out := map[string]int{}
	for rows.Next() {
		var (
			id string
			n  int
		)
		if err := rows.Scan(&id, &n); err != nil {
			return nil, fmt.Errorf("counting per person: %w", err)
		}
		out[id] = n
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("counting per person: %w", err)
	}
	return out, nil
}
