package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// RecordedChanges returns, newest first by when each was recorded and then
// by row ID, up to limit versions the read API's activity shows (D-193):
// every coupon booked, the first version of an interest transaction, and
// every correction, a version after the first, deletions included. Unless
// before is zero, only versions recorded before it, or at it with a row ID
// below beforeID, are returned: those after the last one of a previous
// page. Each version carries its postings and is checked against its
// fingerprint, as the log is.
func (s *Store) RecordedChanges(ctx context.Context, before time.Time, beforeID int64, limit int) ([]app.RecordedVersion, error) {
	var bound *time.Time
	if !before.IsZero() {
		bound = &before
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, recorded_at FROM ledger.transaction_versions
		WHERE (version > 1 OR kind = 'interest')
			AND ($1::timestamptz IS NULL OR (recorded_at, id) < ($1, $2))
		ORDER BY recorded_at DESC, id DESC LIMIT $3`, bound, beforeID, limit)
	if err != nil {
		return nil, fmt.Errorf("loading recorded changes: %w", err)
	}
	var ids []int64
	var times []time.Time
	var id int64
	var at time.Time
	if _, scanErr := pgx.ForEachRow(rows, []any{&id, &at}, func() error {
		ids, times = append(ids, id), append(times, at.UTC())
		return nil
	}); scanErr != nil {
		return nil, fmt.Errorf("loading recorded changes: %w", scanErr)
	}
	versions, err := s.load(ctx, selectVersions+` WHERE id = ANY($1) ORDER BY recorded_at DESC, id DESC`, ids)
	if err != nil {
		return nil, fmt.Errorf("loading recorded changes: %w", err)
	}
	if len(versions) != len(ids) {
		return nil, fmt.Errorf("%w: %d recorded changes loaded as %d versions", ErrCorruptLog, len(ids), len(versions))
	}
	out := make([]app.RecordedVersion, len(ids))
	for i := range ids {
		out[i] = app.RecordedVersion{VersionedTransaction: versions[i], ID: ids[i], Recorded: times[i]}
	}
	return out, nil
}
