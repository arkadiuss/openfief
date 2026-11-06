package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// appendAttempt appends a to the sync history in tx: with its error class
// and failures for a failed attempt, none for a successful one.
func appendAttempt(ctx context.Context, tx pgx.Tx, a *ingest.SyncAttempt) error {
	var class *string
	if a.Failed() {
		name := a.Class.String()
		class = &name
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO ingest.sync_attempts (instance, batch_id, mode, started_at, ended_at, error_class, failures, next_sync_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		a.Instance, string(a.Batch), a.Mode.String(), a.Started, a.Ended, class, a.Failures, a.NextSync); err != nil {
		return fmt.Errorf("recording the attempt: %w", err)
	}
	return nil
}

// SyncAttempts returns the sync attempts q selects, newest first: by when
// they ended, then by ID.
func (s *Store) SyncAttempts(ctx context.Context, q *ingest.AttemptQuery) ([]ingest.SyncAttempt, error) {
	var before *time.Time
	if !q.Before.IsZero() {
		before = &q.Before
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, instance, batch_id, mode, started_at, ended_at, coalesce(error_class, ''), failures, next_sync_at
		FROM ingest.sync_attempts
		WHERE ($1 = '' OR instance = $1) AND (NOT $2 OR error_class IS NOT NULL)
			AND ($3::timestamptz IS NULL OR (ended_at, id) < ($3, $4))
		ORDER BY ended_at DESC, id DESC LIMIT $5`,
		q.Instance, q.FailedOnly, before, q.BeforeID, q.Limit)
	if err != nil {
		return nil, fmt.Errorf("loading sync attempts: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanAttempt)
	if err != nil {
		return nil, fmt.Errorf("loading sync attempts: %w", err)
	}
	return out, nil
}

// scanAttempt reads one sync attempt.
func scanAttempt(row pgx.CollectableRow) (ingest.SyncAttempt, error) {
	var a ingest.SyncAttempt
	var batch, mode, class string
	if err := row.Scan(&a.ID, &a.Instance, &batch, &mode, &a.Started, &a.Ended, &class, &a.Failures, &a.NextSync); err != nil {
		return ingest.SyncAttempt{}, fmt.Errorf("reading a sync attempt: %w", err)
	}
	var err error
	if a.Mode, err = parseMode(mode); err != nil {
		return ingest.SyncAttempt{}, err
	}
	if a.Class, err = ingest.ParseSyncErrorClass(class); err != nil {
		return ingest.SyncAttempt{}, fmt.Errorf("sync attempt %d: %w", a.ID, err)
	}
	a.Batch = ingest.BatchID(batch)
	a.Started, a.Ended, a.NextSync = a.Started.UTC(), a.Ended.UTC(), a.NextSync.UTC()
	return a, nil
}

// SyncStatuses returns where each instance's syncs stand, in instance
// order, from its sync history: every instance that was ever claimed for a
// sync, those with no attempt yet included. A failure counts toward
// Failures and FailingSince only after the instance's latest success.
func (s *Store) SyncStatuses(ctx context.Context) ([]ingest.SyncStatus, error) {
	rows, err := s.pool.Query(ctx, `
		WITH last_ok AS (
			SELECT instance, max(id) AS id, max(ended_at) AS ended_at FROM ingest.sync_attempts
			WHERE error_class IS NULL GROUP BY instance
		), failing AS (
			SELECT a.instance, count(*) AS failures, min(a.started_at) AS since FROM ingest.sync_attempts a
			LEFT JOIN last_ok o ON o.instance = a.instance
			WHERE a.error_class IS NOT NULL AND a.id > coalesce(o.id, 0) GROUP BY a.instance
		), latest AS (
			SELECT DISTINCT ON (instance) instance, ended_at, coalesce(error_class, '') AS error_class
			FROM ingest.sync_attempts ORDER BY instance, id DESC
		)
		SELECT i.instance, l.ended_at, o.ended_at, coalesce(f.failures, 0), f.since, coalesce(l.error_class, ''),
			i.next_sync_at
		FROM ingest.connector_instances i
		LEFT JOIN latest l ON l.instance = i.instance
		LEFT JOIN last_ok o ON o.instance = i.instance
		LEFT JOIN failing f ON f.instance = i.instance
		ORDER BY i.instance`)
	if err != nil {
		return nil, fmt.Errorf("loading sync statuses: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanStatus)
	if err != nil {
		return nil, fmt.Errorf("loading sync statuses: %w", err)
	}
	return out, nil
}

// scanStatus reads one instance's sync status.
func scanStatus(row pgx.CollectableRow) (ingest.SyncStatus, error) {
	var st ingest.SyncStatus
	var last, ok, since, next *time.Time
	var class string
	if err := row.Scan(&st.Instance, &last, &ok, &st.Failures, &since, &class, &next); err != nil {
		return ingest.SyncStatus{}, fmt.Errorf("reading a sync status: %w", err)
	}
	st.LastAttempt, st.LastSuccess, st.FailingSince, st.NextSync = utc(last), utc(ok), utc(since), utc(next)
	if st.Failures > 0 {
		c, err := ingest.ParseSyncErrorClass(class)
		if err != nil {
			return ingest.SyncStatus{}, fmt.Errorf("sync status of %s: %w", st.Instance, err)
		}
		st.LastClass = c
	}
	return st, nil
}

// utc returns *t in UTC, or the zero time for nil.
func utc(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.UTC()
}

// CompletedBatches returns the completed import batches q selects, every
// source alike, newest first: by when they finished, then by ID in byte
// order.
func (s *Store) CompletedBatches(ctx context.Context, q *ingest.BatchQuery) ([]ingest.BatchRecord, error) {
	var before *time.Time
	if !q.Before.IsZero() {
		before = &q.Before
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, instance, source, mode, finished_at, institution, transactions, new, new_versions, unchanged,
			failed_assertions, problems, coalesce(unsupported, 0), coalesce(accounts, '{}')
		FROM ingest.import_batches
		WHERE state = 'completed'
			AND ($1::timestamptz IS NULL OR (finished_at, id COLLATE "C") < ($1, $2::text COLLATE "C"))
		ORDER BY finished_at DESC, id COLLATE "C" DESC LIMIT $3`,
		before, string(q.BeforeID), q.Limit)
	if err != nil {
		return nil, fmt.Errorf("loading completed batches: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanBatchRecord)
	if err != nil {
		return nil, fmt.Errorf("loading completed batches: %w", err)
	}
	return out, nil
}

// scanBatchRecord reads one completed batch.
func scanBatchRecord(row pgx.CollectableRow) (ingest.BatchRecord, error) {
	var b ingest.BatchRecord
	var id, mode, institution string
	var accounts []string
	if err := row.Scan(&id, &b.Instance, &b.Source, &mode, &b.Finished, &institution, &b.Transactions, &b.New,
		&b.NewVersions, &b.Unchanged, &b.FailedAssertions, &b.Problems, &b.Unsupported, &accounts); err != nil {
		return ingest.BatchRecord{}, fmt.Errorf("reading a batch: %w", err)
	}
	for _, a := range accounts {
		b.Accounts = append(b.Accounts, ledger.AccountID(a))
	}
	m, err := parseMode(mode)
	if err != nil {
		return ingest.BatchRecord{}, err
	}
	b.ID, b.Mode, b.Institution, b.Finished = ingest.BatchID(id), m, ingest.InstitutionID(institution), b.Finished.UTC()
	return b, nil
}
