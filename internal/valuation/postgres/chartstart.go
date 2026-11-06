package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// chartStartKey is the valuation.settings key of the portfolio's charts'
// start (D-167); an account's is chartStartKey, a colon and its ID.
const chartStartKey = "chart_start"

// chartStartKeyOf returns the settings key of account's charts' start, the
// portfolio's for an empty account.
func chartStartKeyOf(account ledger.AccountID) string {
	if account == "" {
		return chartStartKey
	}
	return chartStartKey + ":" + string(account)
}

// ChartStart returns the charts' start the owner stored for account, or for
// the portfolio when account is empty (D-167), and the zero date when none
// is stored.
func (s *Store) ChartStart(ctx context.Context, account ledger.AccountID) (kernel.Date, error) {
	text, err := s.setting(ctx, chartStartKeyOf(account))
	if err != nil || text == "" {
		return kernel.Date{}, err
	}
	d, err := kernel.ParseDate(text)
	if err != nil {
		return kernel.Date{}, fmt.Errorf("%w: charts' start: %w", ErrCorrupt, err)
	}
	return d, nil
}

// SaveChartStart stores start as the charts' start of account, or of the
// portfolio when account is empty; a zero start removes the setting.
func (s *Store) SaveChartStart(ctx context.Context, account ledger.AccountID, start kernel.Date) error {
	key := chartStartKeyOf(account)
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if !start.IsZero() {
			return putSetting(ctx, tx, key, start.String())
		}
		if _, err := tx.Exec(ctx, `DELETE FROM valuation.settings WHERE key = $1`, key); err != nil {
			return fmt.Errorf("clearing the setting %s: %w", key, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("saving the charts' start: %w", err)
	}
	return nil
}
