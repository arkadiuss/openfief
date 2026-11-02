package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

var _ app.AccountStarts = (*Store)(nil)

// RecordStarts records the go-live date of each account an opening upload
// started (D-160). An account started before keeps its date.
func (s *Store) RecordStarts(ctx context.Context, starts map[ledger.AccountID]kernel.Date) error {
	accounts, days := make([]string, 0, len(starts)), make([]string, 0, len(starts))
	for a, d := range starts {
		accounts, days = append(accounts, string(a)), append(days, d.String())
	}
	if len(accounts) == 0 {
		return nil
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO ingest.account_starts (account_id, go_live)
		SELECT a, d::date FROM unnest($1::text[], $2::text[]) AS s(a, d)
		ON CONFLICT (account_id) DO NOTHING`, accounts, days)
	if err != nil {
		return fmt.Errorf("recording account starts: %w", err)
	}
	return nil
}

// Start returns the go-live date an opening upload recorded for account,
// and false when none did.
func (s *Store) Start(ctx context.Context, account ledger.AccountID) (kernel.Date, bool, error) {
	var day string
	err := s.pool.QueryRow(ctx, `SELECT go_live::text FROM ingest.account_starts WHERE account_id = $1`, string(account)).Scan(&day)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return kernel.Date{}, false, nil
	case err != nil:
		return kernel.Date{}, false, fmt.Errorf("loading the start of %s: %w", account, err)
	}
	d, err := kernel.ParseDate(day)
	if err != nil {
		return kernel.Date{}, false, fmt.Errorf("the start of %s: %w", account, err)
	}
	return d, true, nil
}
