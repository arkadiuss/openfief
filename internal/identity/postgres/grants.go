package postgres

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

var _ app.GrantStore = (*Store)(nil)

// SaveGrant stores g as in force, in the transaction ctx carries or on its
// own. It stores nothing and returns false when a grant of g's target to g's
// holder is in force already: the partial unique index grants_in_force
// decides, so two concurrent calls make one grant.
func (s *Store) SaveGrant(ctx context.Context, g *identity.Grant) (bool, error) {
	if err := g.Validate(); err != nil {
		return false, fmt.Errorf("saving a grant: %w", err)
	}
	tag, err := s.db(ctx).Exec(ctx, `
		INSERT INTO identity.grants (holder_kind, holder_id, target_kind, target_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (holder_kind, holder_id, target_kind, target_id) WHERE revoked_at IS NULL DO NOTHING`,
		string(g.Holder.Kind), g.Holder.ID, string(g.Target.Kind), g.Target.ID, g.CreatedAt)
	if err != nil {
		return false, fmt.Errorf("saving a grant: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

// grantColumns are the columns scanGrant reads, in its order.
const grantColumns = `holder_kind, holder_id, target_kind, target_id, created_at, revoked_at`

// grantFilter is the WHERE condition of the holder and target m selects, over
// the arguments $1 to $4: a zero side matches every row.
const grantFilter = `($1::text = '' OR (holder_kind = $1 AND holder_id = $2)) AND ($3::text = '' OR (target_kind = $3 AND target_id = $4))`

// RevokeGrants ends, as of at, every grant in force that m selects, and
// returns them as ended, in creation order. The time of an ending never
// precedes the grant's own creation.
func (s *Store) RevokeGrants(ctx context.Context, m identity.GrantMatch, at time.Time) ([]identity.Grant, error) {
	rows, err := s.db(ctx).Query(ctx, `
		UPDATE identity.grants SET revoked_at = GREATEST($5, created_at)
		WHERE revoked_at IS NULL AND `+grantFilter+`
		RETURNING `+grantColumns+`, id`,
		string(m.Holder.Kind), m.Holder.ID, string(m.Target.Kind), m.Target.ID, at)
	if err != nil {
		return nil, fmt.Errorf("revoking grants: %w", err)
	}
	type endedGrant struct {
		grant identity.Grant
		id    int64
	}
	ended, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (endedGrant, error) {
		var id int64
		g, scanErr := scanGrant(row, &id)
		return endedGrant{grant: g, id: id}, scanErr
	})
	if err != nil {
		return nil, fmt.Errorf("revoking grants: %w", err)
	}
	// UPDATE ... RETURNING has no order of its own.
	slices.SortFunc(ended, func(x, y endedGrant) int {
		return cmp.Or(x.grant.CreatedAt.Compare(y.grant.CreatedAt), cmp.Compare(x.id, y.id))
	})
	out := make([]identity.Grant, 0, len(ended))
	for i := range ended {
		out = append(out, ended[i].grant)
	}
	return out, nil
}

// Grants returns the grants m selects, oldest first and in ID order: those in
// force only, or the ended ones too.
func (s *Store) Grants(ctx context.Context, m identity.GrantMatch, inForceOnly bool) ([]identity.Grant, error) {
	rows, err := s.db(ctx).Query(ctx, `
		SELECT `+grantColumns+` FROM identity.grants
		WHERE `+grantFilter+` AND (NOT $5::boolean OR revoked_at IS NULL)
		ORDER BY created_at, id`,
		string(m.Holder.Kind), m.Holder.ID, string(m.Target.Kind), m.Target.ID, inForceOnly)
	if err != nil {
		return nil, fmt.Errorf("listing grants: %w", err)
	}
	grants, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (identity.Grant, error) { return scanGrant(row) })
	if err != nil {
		return nil, fmt.Errorf("listing grants: %w", err)
	}
	return grants, nil
}

// GrantCutoverDone reports whether the cutover to SpiceDB was recorded.
func (s *Store) GrantCutoverDone(ctx context.Context) (bool, error) {
	var done bool
	if err := s.db(ctx).QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM identity.grant_cutover)`).Scan(&done); err != nil {
		return false, fmt.Errorf("reading the grant cutover: %w", err)
	}
	return done, nil
}

// MarkGrantCutover records that the cutover ran at at, and returns false when
// it was recorded already.
func (s *Store) MarkGrantCutover(ctx context.Context, at time.Time) (bool, error) {
	tag, err := s.db(ctx).Exec(ctx, `INSERT INTO identity.grant_cutover (done_at) VALUES ($1) ON CONFLICT DO NOTHING`, at)
	if err != nil {
		return false, fmt.Errorf("recording the grant cutover: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

// scanGrant reads a row of grantColumns, then into extra any further
// columns, and checks the grant it holds is one the domain accepts.
func scanGrant(row pgx.Row, extra ...any) (identity.Grant, error) {
	var (
		holderKind, targetKind string
		g                      identity.Grant
		created                time.Time
		revoked                *time.Time
	)
	dest := append([]any{&holderKind, &g.Holder.ID, &targetKind, &g.Target.ID, &created, &revoked}, extra...)
	if err := row.Scan(dest...); err != nil {
		return identity.Grant{}, fmt.Errorf("scanning a grant: %w", err)
	}
	g.Holder.Kind, g.Target.Kind = identity.GrantHolderKind(holderKind), identity.GrantTargetKind(targetKind)
	g.CreatedAt = created.UTC()
	if revoked != nil {
		g.RevokedAt = revoked.UTC()
	}
	if err := g.Validate(); err != nil {
		return identity.Grant{}, fmt.Errorf("%w: a grant: %w", errBadRow, err)
	}
	return g, nil
}
