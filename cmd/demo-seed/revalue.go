package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

var (
	// errRevaluationFailed reports a revaluation job that failed.
	errRevaluationFailed = errors.New("a revaluation job failed")
	// errRevaluationTimeout reports revaluations still pending at the
	// deadline.
	errRevaluationTimeout = errors.New("revaluations still pending")
)

// workers is how many revaluations run at once during the load.
const workers = 3

// The end-of-day schedule the workers are given. The load stops them long
// before it: the deployment's own workers own the end of day.
const eodHour, eodMinute = 23, 30

// pollEvery is how often the load checks whether the revaluations are done.
const pollEvery = 500 * time.Millisecond

// revalue works the revaluations the load enqueued on the River queue with
// the real workers (D-36), valuing through the dataset's today, and returns
// once every demo account is valued with nothing pending, or fails after
// cfg.Wait or on the first failed job.
func revalue(ctx context.Context, pool *pgxpool.Pool, l *loader, cfg *config) error {
	client, err := valuationpg.NewClient(pool, l.vstore, l.store, &valuationpg.Config{
		Concurrency: workers,
		EndOfDay:    valuationpg.DailyAt{Hour: eodHour, Minute: eodMinute, Location: time.UTC},
		Today:       func() kernel.Date { return l.ds.Today },
		Logger:      l.log,
		Pack:        pl.NewPack(),
	})
	if err != nil {
		return fmt.Errorf("revaluation client: %w", err)
	}
	failed, cancel := client.Subscribe(river.EventKindJobFailed)
	defer cancel()
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("starting the revaluation workers: %w", err)
	}
	waitErr := waitValued(ctx, l, failed, cfg.Wait)
	if err := client.Stop(context.WithoutCancel(ctx)); err != nil {
		return errors.Join(waitErr, fmt.Errorf("stopping the revaluation workers: %w", err))
	}
	return waitErr
}

// waitValued polls the valuation statuses until every demo account is
// valued through today with no revaluation pending or running.
func waitValued(ctx context.Context, l *loader, failed <-chan *river.Event, wait time.Duration) error {
	deadline := time.NewTimer(wait)
	defer deadline.Stop()
	tick := time.NewTicker(pollEvery)
	defer tick.Stop()
	for {
		statuses, err := l.vstore.Statuses(ctx)
		if err != nil {
			return fmt.Errorf("revaluation status: %w", err)
		}
		if allValued(statuses, l.ds) {
			for _, s := range statuses {
				l.log.Info("account valued", "account", string(s.Account), "valued_through", s.ValuedThrough.String())
			}
			return nil
		}
		select {
		case e := <-failed:
			return fmt.Errorf("%w: job %d (%s): %s", errRevaluationFailed, e.Job.ID, e.Job.Kind, lastError(e.Job))
		case <-deadline.C:
			return fmt.Errorf("%w after %s", errRevaluationTimeout, wait)
		case <-ctx.Done():
			return fmt.Errorf("waiting for the revaluations: %w", ctx.Err())
		case <-tick.C:
		}
	}
}

// allValued reports whether every account of ds has a status that is
// neither pending nor running and is valued through ds.Today.
func allValued(statuses []valuationapp.ValuationStatus, ds *demo.Dataset) bool {
	done := map[string]bool{}
	for _, s := range statuses {
		done[string(s.Account)] = !s.Pending && !s.Running && s.ValuedThrough.Equal(ds.Today)
	}
	for _, a := range ds.Accounts {
		if !done[string(a.ID)] {
			return false
		}
	}
	return true
}

// lastError returns the error of job's latest attempt, or "" if none.
func lastError(job *rivertype.JobRow) string {
	if len(job.Errors) == 0 {
		return ""
	}
	return job.Errors[len(job.Errors)-1].Error
}
