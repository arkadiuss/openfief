// Package jobs runs the portfolio context's River jobs: the model-portfolio
// maintainer's runs (D-204), the weekly one on a schedule and the owner's on
// request. The jobs ride on the one River client the revaluation workers use
// (internal/valuation/postgres, D-36), since River keeps a single leader per
// schema that enqueues every periodic job.
package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
)

// Queue is the River queue the maintainer's jobs run on, apart from the
// revaluations so a model call never holds one up.
const Queue = "portfolio"

// RunTimeout bounds one run: the router's own timeouts sit well inside it,
// and app.StaleAfter, after which a run that never ended is told lost, sits
// well outside it.
const RunTimeout = 10 * time.Minute

const (
	// daysInWeek is how many days a week has.
	daysInWeek = 7
	// weeklyHour and weeklyMinute are when the weekly run is due, Monday.
	weeklyHour, weeklyMinute = 6, 30
	// maxAttempts is how many times a job is tried: an owner's run only when
	// its worker fails, the weekly turn once for each of the week's attempts.
	maxAttempts = app.WeeklyAttempts
)

// The delays before the weekly turn's retries (D-198's pattern): the second
// attempt a quarter of an hour after the first ended, the third an hour after
// the second.
const (
	firstRetryDelay  = 15 * time.Minute
	secondRetryDelay = time.Hour
)

// WeeklyTime is when the maintainer's weekly run is due: Monday at 06:30 in
// the core's time zone, after the end-of-day revaluation has valued the
// week's last day.
func WeeklyTime(loc *time.Location) WeeklyAt {
	return WeeklyAt{Weekday: time.Monday, Hour: weeklyHour, Minute: weeklyMinute, Location: loc}
}

// WeeklyAt is a River schedule that fires once a week.
type WeeklyAt struct {
	Weekday      time.Weekday
	Hour, Minute int
	Location     *time.Location
}

// Next returns the first time after current that is Weekday at Hour:Minute
// in Location.
func (w WeeklyAt) Next(current time.Time) time.Time {
	t := current.In(w.Location)
	days := (int(w.Weekday) - int(t.Weekday()) + daysInWeek) % daysInWeek
	next := time.Date(t.Year(), t.Month(), t.Day()+days, w.Hour, w.Minute, 0, 0, w.Location)
	if !next.After(t) {
		next = time.Date(t.Year(), t.Month(), t.Day()+days+daysInWeek, w.Hour, w.Minute, 0, 0, w.Location)
	}
	return next
}

// RunMaintainer is the job that works one run of the maintainer. A job with
// a run ID works the run the owner asked for, already queued; one without is
// the weekly turn, which queues its own run, or does nothing when another is
// queued or running.
type RunMaintainer struct {
	RunID string `json:"runId,omitempty"`
}

// Kind names the job kind.
func (RunMaintainer) Kind() string { return "portfolio.run_maintainer" }

// InsertOpts puts the job on Queue, with three attempts: the retry of an
// owner's run takes over the run its earlier attempt left running, and the
// weekly turn's are its week's attempts.
func (RunMaintainer) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: Queue, MaxAttempts: maxAttempts}
}

// Runner is what a job calls: *app.Maintainer.
type Runner interface {
	Weekly(ctx context.Context, final bool) (string, error)
	Work(ctx context.Context, runID string) error
}

// Late is the maintainer a worker calls, set once the core has built it: the
// River client is built before the views the maintainer reads, and starts
// only after. Until it is set a job waits.
type Late struct {
	runner atomic.Pointer[Runner]
}

// Set makes r the maintainer the workers call.
func (l *Late) Set(r Runner) { l.runner.Store(&r) }

// Worker works RunMaintainer jobs.
type Worker struct {
	river.WorkerDefaults[RunMaintainer]

	late *Late
	log  *slog.Logger
}

// Timeout bounds a run to RunTimeout.
func (*Worker) Timeout(*river.Job[RunMaintainer]) time.Duration { return RunTimeout }

// Work runs the job's run, or the weekly turn. A failure reaches River as a
// *JobError, without any text of a note or a figure (GR-12).
func (w *Worker) Work(ctx context.Context, job *river.Job[RunMaintainer]) error {
	p := w.late.runner.Load()
	if p == nil {
		return fmt.Errorf("the maintainer is not built yet: %w", river.JobSnooze(time.Minute))
	}
	r := *p
	var err error
	if job.Args.RunID == "" {
		_, err = r.Weekly(ctx, job.Attempt >= job.MaxAttempts)
	} else {
		err = r.Work(ctx, job.Args.RunID)
	}
	if err == nil {
		return nil
	}
	safe := kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded, app.ErrRunInProgress, app.ErrMaintainerOff)
	if failed, ok := errors.AsType[*app.AttemptFailedError](err); ok {
		// The class is one of a fixed set and names no text of a note.
		safe = failed.Error()
	}
	w.log.LogAttrs(ctx, slog.LevelError, "maintainer job failed", slog.String("error", safe), slog.Int("attempt", job.Attempt))
	return &JobError{Safe: safe}
}

// NextRetry spaces the weekly turn's attempts: a quarter of an hour after the
// first, an hour after the second. An owner's run keeps River's default
// backoff, since it is retried only when the core failed to store it.
func (*Worker) NextRetry(job *river.Job[RunMaintainer]) time.Time {
	if job.Args.RunID != "" {
		return time.Time{}
	}
	return time.Now().Add(RetryDelay(job.Attempt))
}

// RetryDelay is how long the weekly turn waits after attempt number attempt
// (1 for the first) failed.
func RetryDelay(attempt int) time.Duration {
	if attempt <= 1 {
		return firstRetryDelay
	}
	return secondRetryDelay
}

// JobError is what the worker hands River for a failed job: the failure
// described without any value it carried (GR-12), since River logs a failed
// job's error in full and keeps it in river_job.errors.
type JobError struct {
	Safe string
}

// Error names the job and the failure by its kind only.
func (e *JobError) Error() string { return "maintainer job: " + e.Safe }

// ErrInvalidJob reports a job built without what it needs.
var ErrInvalidJob = errors.New("invalid maintainer job")

// New returns the worker over late and the periodic job that asks for the
// weekly turn on schedule; cmd/core hands both to the River client.
func New(late *Late, weekly river.PeriodicSchedule, log *slog.Logger) (*Worker, *river.PeriodicJob, error) {
	if late == nil || weekly == nil || log == nil {
		return nil, nil, fmt.Errorf("%w: a runner, a schedule and a logger", ErrInvalidJob)
	}
	periodic := river.NewPeriodicJob(weekly, func() (river.JobArgs, *river.InsertOpts) {
		args, opts := WeeklyJob()
		return args, opts
	}, &river.PeriodicJobOpts{ID: RunMaintainer{}.Kind()})
	return &Worker{late: late, log: log}, periodic, nil
}

// WeeklyJob is what the schedule inserts: the weekly turn, with the week's
// attempts (app.WeeklyAttempts), the later ones after NextRetry's delays. A
// week whose attempts all end without a note raises the failing item.
func WeeklyJob() (RunMaintainer, *river.InsertOpts) {
	return RunMaintainer{}, &river.InsertOpts{MaxAttempts: app.WeeklyAttempts}
}

// Queued enqueues the job of a run the owner asked for: it is the
// app.QueuedRun port over the River client.
type Queued struct {
	Client *river.Client[pgx.Tx]
}

// Enqueue puts the job of run runID on the queue.
func (q Queued) Enqueue(ctx context.Context, runID string) error {
	if _, err := q.Client.Insert(ctx, RunMaintainer{RunID: runID}, nil); err != nil {
		return fmt.Errorf("inserting the maintainer's job: %w", err)
	}
	return nil
}
