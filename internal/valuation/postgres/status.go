package postgres

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// statusSQL lists, per account with a valued-through date, pending work or
// a live RevalueAccount job: the date, whether a change is not yet picked up,
// whether pending work exists at all, and whether a job is waiting or
// running.
const statusSQL = `
WITH jobs AS (
    SELECT args->>'account' AS account_id,
           bool_or(state = 'running') AS running,
           bool_or(state IN ('available', 'pending', 'scheduled', 'retryable')) AS waiting
    FROM river.river_job
    WHERE kind = $1 AND state IN ('available', 'pending', 'scheduled', 'retryable', 'running')
    GROUP BY 1
), accounts AS (
    SELECT account_id FROM valuation.account_status
    UNION SELECT account_id FROM valuation.dirty
    UNION SELECT account_id FROM jobs
)
SELECT a.account_id,
       coalesce(s.valued_through::text, ''),
       coalesce(d.from_date IS NOT NULL, false),
       d.account_id IS NOT NULL,
       coalesce(j.waiting, false),
       coalesce(j.running, false)
FROM accounts a
LEFT JOIN valuation.account_status s USING (account_id)
LEFT JOIN valuation.dirty d USING (account_id)
LEFT JOIN jobs j USING (account_id)`

// Statuses returns the valuation status of every account that was ever
// revalued or has a revaluation pending or running, in account order
// (D-36): its valued-through date, zero if never valued; Pending when a
// change is not yet picked up, a job is waiting, or pending work has no
// running job (such as a failed one); Running when a job runs now.
func (s *Store) Statuses(ctx context.Context) ([]app.ValuationStatus, error) {
	rows, err := s.pool.Query(ctx, statusSQL, RevalueAccount{}.Kind())
	if err != nil {
		return nil, fmt.Errorf("loading valuation statuses: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanStatus)
	if err != nil {
		return nil, fmt.Errorf("loading valuation statuses: %w", err)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Account < out[j].Account })
	return out, nil
}

// Status returns account's valuation status, as Statuses does; an account
// never revalued and with nothing pending has a zero date and neither flag.
func (s *Store) Status(ctx context.Context, account ledger.AccountID) (app.ValuationStatus, error) {
	all, err := s.Statuses(ctx)
	if err != nil {
		return app.ValuationStatus{}, err
	}
	for _, st := range all {
		if st.Account == account {
			return st, nil
		}
	}
	return app.ValuationStatus{Account: account}, nil
}

// scanStatus reads one row of statusSQL.
func scanStatus(row pgx.CollectableRow) (app.ValuationStatus, error) {
	var account, through string
	var changed, dirty, waiting, running bool
	if err := row.Scan(&account, &through, &changed, &dirty, &waiting, &running); err != nil {
		return app.ValuationStatus{}, fmt.Errorf("reading a valuation status: %w", err)
	}
	var d kernel.Date
	if through != "" {
		var err error
		if d, err = kernel.ParseDate(through); err != nil {
			return app.ValuationStatus{}, fmt.Errorf("%w: valued-through date of %s: %w", ErrCorrupt, account, err)
		}
	}
	return app.ValuationStatus{
		Account: ledger.AccountID(account), ValuedThrough: d,
		Pending: changed || waiting || (dirty && !running), Running: running,
	}, nil
}
