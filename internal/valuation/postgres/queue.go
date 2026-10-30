package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// Queue is the River queue revaluation jobs run on.
const Queue = "valuation"

// revalueLockClass is the first key of the per-account advisory lock that
// orders enqueueing against a worker finishing: "re" in ASCII.
const revalueLockClass int32 = 0x7265

// ErrNoLedgerTx reports an import hook called with a ledger store that
// cannot hand over its database transaction.
var ErrNoLedgerTx = errors.New("ledger store has no database transaction to enqueue in")

// RevalueAccount is the job that revalues one account. Its args name only
// the account: River keeps at most one such job per account waiting or
// running (unique by args), and the date to revalue from is in
// valuation.dirty, where every enqueue lowers it to the earliest (D-36).
type RevalueAccount struct {
	Account string `json:"account" river:"unique"`
}

// Kind names the job kind.
func (RevalueAccount) Kind() string { return "valuation.revalue_account" }

// InsertOpts puts the job on Queue, unique per account among jobs waiting
// or running. Completed jobs don't count, so a change after a run enqueues a
// new one.
func (RevalueAccount) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: Queue,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable, rivertype.JobStatePending, rivertype.JobStateRunning,
				rivertype.JobStateRetryable, rivertype.JobStateScheduled,
			},
		},
	}
}

// enqueueTx enqueues revals on tx: for each account, in account order, it
// takes the account's revaluation lock, lowers its pending from-date to the
// earliest (LEAST), and inserts a RevalueAccount job unless one is already
// waiting or running.
func (s *Store) enqueueTx(ctx context.Context, tx pgx.Tx, revals []valuation.Revaluation) error {
	coalesced := valuation.Coalesce(revals)
	params := make([]river.InsertManyParams, 0, len(coalesced))
	for _, r := range coalesced {
		if err := lockAccount(ctx, tx, r.Account); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO valuation.dirty (account_id, from_date) VALUES ($1, $2)
			ON CONFLICT (account_id) DO UPDATE SET from_date = LEAST(valuation.dirty.from_date, EXCLUDED.from_date)`,
			string(r.Account), r.From.String())
		if err != nil {
			return fmt.Errorf("enqueueing a revaluation of %s: %w", r.Account, err)
		}
		params = append(params, river.InsertManyParams{Args: RevalueAccount{Account: string(r.Account)}})
	}
	if len(params) == 0 {
		return nil
	}
	if _, err := s.insert.InsertManyTx(ctx, tx, params); err != nil {
		return fmt.Errorf("enqueueing revaluations: %w", err)
	}
	return nil
}

// lockAccount takes account's revaluation lock until tx ends.
func lockAccount(ctx context.Context, tx pgx.Tx, account ledger.AccountID) error {
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1, hashtext($2))", revalueLockClass, string(account)); err != nil {
		return fmt.Errorf("locking the revaluations of %s: %w", account, err)
	}
	return nil
}

// inTx runs write and then enqueues revals, in one database transaction.
func (s *Store) inTx(ctx context.Context, revals []valuation.Revaluation, write func(tx pgx.Tx) error) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := write(tx); err != nil {
			return err
		}
		return s.enqueueTx(ctx, tx, revals)
	})
	if err != nil {
		return fmt.Errorf("valuation transaction: %w", err)
	}
	return nil
}

// Enqueue enqueues revals in a database transaction of its own.
func (s *Store) Enqueue(ctx context.Context, revals []valuation.Revaluation) error {
	return s.inTx(ctx, revals, func(pgx.Tx) error { return nil })
}

// ImportHook returns the ledger import hook that enqueues a revaluation of
// every account an import changed, from its earliest affected date, in the
// import's own database transaction (D-36). The ledger store must hand that
// transaction over with WithTx, as ledger/postgres.Store does.
func (s *Store) ImportHook() ledgerapp.ImportHook {
	return func(ctx context.Context, tx ledgerapp.Store, sum *ledgerapp.ImportSummary) error {
		revals := app.ForImport(sum.EarliestAffected)
		if len(revals) == 0 {
			return nil
		}
		ledgerTx, ok := tx.(interface {
			WithTx(fn func(tx pgx.Tx) error) error
		})
		if !ok {
			return ErrNoLedgerTx
		}
		return ledgerTx.WithTx(func(ptx pgx.Tx) error { return s.enqueueTx(ctx, ptx, revals) })
	}
}

// SavePrices upserts prices and enqueues revals in one transaction.
func (s *Store) SavePrices(ctx context.Context, prices []valuation.Price, revals []valuation.Revaluation) error {
	err := s.inTx(ctx, revals, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}
		for _, p := range prices {
			batch.Queue(`
				INSERT INTO valuation.prices (instrument, date, close, currency, source) VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (instrument, date, source) DO UPDATE SET close = EXCLUDED.close, currency = EXCLUDED.currency`,
				p.Instrument, p.Date.String(), p.Close.Amount().String(), p.Close.Currency().String(), p.Source)
		}
		return tx.SendBatch(ctx, batch).Close()
	})
	if err != nil {
		return fmt.Errorf("saving prices: %w", err)
	}
	return nil
}

// SaveFXRates upserts rates and enqueues revals in one transaction.
func (s *Store) SaveFXRates(ctx context.Context, rates []valuation.FXRate, revals []valuation.Revaluation) error {
	err := s.inTx(ctx, revals, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}
		for _, r := range rates {
			batch.Queue(`
				INSERT INTO valuation.fx_rates (currency, date, rate, source) VALUES ($1, $2, $3, $4)
				ON CONFLICT (currency, date, source) DO UPDATE SET rate = EXCLUDED.rate`,
				r.Currency.String(), r.Date.String(), r.Rate.String(), r.Source)
		}
		return tx.SendBatch(ctx, batch).Close()
	})
	if err != nil {
		return fmt.Errorf("saving FX rates: %w", err)
	}
	return nil
}

// SaveIndexLevels upserts market index levels in one transaction. They
// cause no revaluation.
func (s *Store) SaveIndexLevels(ctx context.Context, levels []valuation.IndexLevel) error {
	err := s.inTx(ctx, nil, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}
		for _, l := range levels {
			batch.Queue(`
				INSERT INTO valuation.index_levels (index, date, level, source) VALUES ($1, $2, $3, $4)
				ON CONFLICT (index, date) DO UPDATE SET level = EXCLUDED.level, source = EXCLUDED.source`,
				l.Index, l.Date.String(), l.Level.String(), l.Source)
		}
		return tx.SendBatch(ctx, batch).Close()
	})
	if err != nil {
		return fmt.Errorf("saving index levels: %w", err)
	}
	return nil
}

// SaveIndexReadings upserts readings and enqueues revals in one transaction.
func (s *Store) SaveIndexReadings(ctx context.Context, readings []valuation.IndexReading, revals []valuation.Revaluation) error {
	err := s.inTx(ctx, revals, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}
		for _, r := range readings {
			batch.Queue(`
				INSERT INTO valuation.index_readings (kind, year, month, value, source) VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (kind, year, month, source) DO UPDATE SET value = EXCLUDED.value`,
				r.Kind.String(), r.Year, int(r.Month), r.Value.String(), r.Source)
		}
		return tx.SendBatch(ctx, batch).Close()
	})
	if err != nil {
		return fmt.Errorf("saving index readings: %w", err)
	}
	return nil
}

// SaveBondTerms upserts the series terms of the bond with isin and enqueues
// revals in one transaction.
func (s *Store) SaveBondTerms(ctx context.Context, isin ledger.ISIN, terms *pl.Terms, revals []valuation.Revaluation) error {
	err := s.inTx(ctx, revals, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO valuation.bond_terms (isin, series, first_period_rate, margin, early_redemption_fee)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (isin) DO UPDATE SET series = EXCLUDED.series, first_period_rate = EXCLUDED.first_period_rate,
			    margin = EXCLUDED.margin, early_redemption_fee = EXCLUDED.early_redemption_fee`,
			isin.String(), terms.Series.String(), terms.FirstPeriodRate.String(), terms.Margin.String(),
			terms.EarlyRedemptionFee.Amount().String())
		if err != nil {
			return fmt.Errorf("upserting: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("saving bond terms of %s: %w", isin, err)
	}
	return nil
}
