package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// ErrInvalidConfig reports revaluation settings that cannot be used.
var ErrInvalidConfig = errors.New("invalid revaluation settings")

// maxRounds is how many times one RevalueAccount job revalues before it
// gives way: each round picks up changes that arrived during the last one.
// After that, the job is snoozed and runs again at once, without using up an
// attempt.
const maxRounds = 5

const (
	hoursPerDay    = 24
	minutesPerHour = 60
)

// Config is what D-36 makes configurable. What triggers a revaluation is
// not.
type Config struct {
	// Concurrency is how many revaluations run at once: the MaxWorkers of
	// Queue in this process.
	Concurrency int
	// EndOfDay is when the end-of-day revaluation of every account runs.
	EndOfDay DailyAt
	// BondBasis is the bond basis for totals (D-25); zero is the default,
	// net if redeemed. Apply it with app.InputRecorder.ApplyBondBasis at
	// start-up, which revalues everything when it changed.
	BondBasis valuation.Basis
	// Today returns the date revaluations value through.
	Today func() kernel.Date
	// Logger receives River's logs; nil is slog's default.
	Logger *slog.Logger
}

// DailyAt is a River schedule that fires once a day at Hour:Minute in
// Location.
type DailyAt struct {
	Hour, Minute int
	Location     *time.Location
}

// Next returns the first time after current at Hour:Minute in Location.
func (d DailyAt) Next(current time.Time) time.Time {
	t := current.In(d.Location)
	next := time.Date(t.Year(), t.Month(), t.Day(), d.Hour, d.Minute, 0, 0, d.Location)
	if !next.After(t) {
		next = time.Date(t.Year(), t.Month(), t.Day()+1, d.Hour, d.Minute, 0, 0, d.Location)
	}
	return next
}

// validate checks c.
func (c *Config) validate() error {
	switch {
	case c.Concurrency < 1:
		return fmt.Errorf("%w: concurrency %d, want at least 1", ErrInvalidConfig, c.Concurrency)
	case c.EndOfDay.Location == nil:
		return fmt.Errorf("%w: end-of-day schedule without a time zone", ErrInvalidConfig)
	case c.EndOfDay.Hour < 0 || c.EndOfDay.Hour >= hoursPerDay || c.EndOfDay.Minute < 0 || c.EndOfDay.Minute >= minutesPerHour:
		return fmt.Errorf("%w: end of day at %d:%02d", ErrInvalidConfig, c.EndOfDay.Hour, c.EndOfDay.Minute)
	case c.Today == nil:
		return fmt.Errorf("%w: no date source", ErrInvalidConfig)
	}
	return nil
}

// TodayIn returns the current date in loc, read from now.
func TodayIn(loc *time.Location, now func() time.Time) func() kernel.Date {
	return func() kernel.Date {
		t := now().In(loc)
		d, err := kernel.NewDate(t.Year(), t.Month(), t.Day())
		if err != nil {
			return kernel.Date{}
		}
		return d
	}
}

// Ledger is what the workers read from the ledger context: its accounts and
// their logs. ledger/app.Ledger and the ledger's store provide them.
type Ledger interface {
	app.AccountLister
	app.TransactionLog
}

// NewClient returns a River client that works revaluations on Queue with
// cfg.Concurrency workers and inserts the end-of-day job on cfg.EndOfDay.
// Start it to run them; enqueueing needs only the Store.
func NewClient(pool *pgxpool.Pool, store *Store, l Ledger, cfg *Config) (*river.Client[pgx.Tx], error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	revaluer, err := app.NewRevaluer(l, store, cfg.BondBasis, cfg.Today)
	if err != nil {
		return nil, fmt.Errorf("revaluation client: %w", err)
	}
	workers := river.NewWorkers()
	eod := &EndOfDayWorker{store: store, recorder: app.NewInputRecorder(app.NewPlanner(l, l), store)}
	if err = errors.Join(
		river.AddWorkerSafely(workers, &RevalueWorker{store: store, revaluer: revaluer}),
		river.AddWorkerSafely(workers, eod),
	); err != nil {
		return nil, fmt.Errorf("revaluation client: %w", err)
	}
	c, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Schema:  RiverSchema,
		Logger:  cfg.Logger,
		Queues:  map[string]river.QueueConfig{Queue: {MaxWorkers: cfg.Concurrency}},
		Workers: workers,
		PeriodicJobs: []*river.PeriodicJob{river.NewPeriodicJob(cfg.EndOfDay,
			func() (river.JobArgs, *river.InsertOpts) { return EndOfDay{}, nil },
			&river.PeriodicJobOpts{ID: EndOfDay{}.Kind()})},
	})
	if err != nil {
		return nil, fmt.Errorf("revaluation client: %w", err)
	}
	return c, nil
}

// RevalueWorker works RevalueAccount jobs: it revalues the account from its
// pending from-date until no change is left, then completes the job in the
// same transaction that clears the pending date, so that no change is ever
// left without a job.
type RevalueWorker struct {
	river.WorkerDefaults[RevalueAccount]

	store    *Store
	revaluer *app.Revaluer
}

// Work revalues job's account.
func (w *RevalueWorker) Work(ctx context.Context, job *river.Job[RevalueAccount]) error {
	account := ledger.AccountID(job.Args.Account)
	for range maxRounds {
		from, ok, err := w.store.claim(ctx, account)
		if err != nil {
			return err
		}
		if ok {
			if _, revalueErr := w.revaluer.Revalue(ctx, account, from); revalueErr != nil {
				return fmt.Errorf("revaluation job: %w", revalueErr)
			}
		}
		done, err := w.store.finish(ctx, account, job)
		if err != nil || done {
			return err
		}
	}
	return fmt.Errorf("revaluation job: changes kept arriving: %w", river.JobSnooze(0))
}

// claim takes account's pending from-date for a revaluation: the earlier of
// the date not yet picked up and the date of an earlier claim that did not
// finish. It reports false when nothing is pending.
func (s *Store) claim(ctx context.Context, account ledger.AccountID) (kernel.Date, bool, error) {
	var from string
	err := s.pool.QueryRow(ctx, `
		UPDATE valuation.dirty SET claimed_from = LEAST(claimed_from, from_date), from_date = NULL
		WHERE account_id = $1 RETURNING claimed_from::text`, string(account)).Scan(&from)
	if errors.Is(err, pgx.ErrNoRows) {
		return kernel.Date{}, false, nil
	}
	if err != nil {
		return kernel.Date{}, false, fmt.Errorf("claiming the revaluation of %s: %w", account, err)
	}
	d, err := kernel.ParseDate(from)
	if err != nil {
		return kernel.Date{}, false, fmt.Errorf("%w: pending date of %s: %w", ErrCorrupt, account, err)
	}
	return d, true, nil
}

// finish ends a round of job: under account's revaluation lock, it clears
// the pending row and completes job, in one transaction, unless a change
// arrived since the claim; then it reports false and the caller goes again.
func (s *Store) finish(ctx context.Context, account ledger.AccountID, job *river.Job[RevalueAccount]) (bool, error) {
	done := false
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := lockAccount(ctx, tx, account); err != nil {
			return err
		}
		var changed bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS (SELECT FROM valuation.dirty WHERE account_id = $1 AND from_date IS NOT NULL)`,
			string(account)).Scan(&changed)
		if err != nil {
			return fmt.Errorf("checking for new changes: %w", err)
		}
		if changed {
			return nil
		}
		if _, err := tx.Exec(ctx, `DELETE FROM valuation.dirty WHERE account_id = $1`, string(account)); err != nil {
			return fmt.Errorf("clearing the pending revaluation: %w", err)
		}
		if _, err := river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job); err != nil {
			return fmt.Errorf("completing the job: %w", err)
		}
		done = true
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("finishing the revaluation of %s: %w", account, err)
	}
	return done, nil
}

// EndOfDay is the periodic job that revalues every account from its
// valued-through date.
type EndOfDay struct{}

// Kind names the job kind.
func (EndOfDay) Kind() string { return "valuation.end_of_day" }

// InsertOpts puts the job on Queue.
func (EndOfDay) InsertOpts() river.InsertOpts { return river.InsertOpts{Queue: Queue} }

// EndOfDayWorker works EndOfDay jobs: it enqueues the revaluation of every
// account (app.InputRecorder.EndOfDay).
type EndOfDayWorker struct {
	river.WorkerDefaults[EndOfDay]

	store    *Store
	recorder *app.InputRecorder
}

// Work enqueues the end-of-day revaluations.
func (w *EndOfDayWorker) Work(ctx context.Context, _ *river.Job[EndOfDay]) error {
	if err := w.recorder.EndOfDay(ctx, w.store); err != nil {
		return fmt.Errorf("end-of-day job: %w", err)
	}
	return nil
}
