package jobs_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"github.com/arkadiuss/openfief/internal/portfolio/jobs"
)

var errQuote = errors.New("a failure that quotes ETFA 63.1 and a note")

// fakeRunner records the calls and fails with err.
type fakeRunner struct {
	weekly int
	worked []string
	err    error
}

func (f *fakeRunner) Weekly(context.Context) (string, error) {
	f.weekly++
	return "", f.err
}

func (f *fakeRunner) Work(_ context.Context, id string) error {
	f.worked = append(f.worked, id)
	return f.err
}

func job(runID string) *river.Job[jobs.RunMaintainer] {
	return &river.Job[jobs.RunMaintainer]{JobRow: &rivertype.JobRow{Attempt: 1}, Args: jobs.RunMaintainer{RunID: runID}}
}

func newWorker(t *testing.T) (*jobs.Worker, *jobs.Late) {
	t.Helper()
	late := &jobs.Late{}
	w, _, err := jobs.New(late, jobs.WeeklyTime(time.UTC), slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return w, late
}

func TestWeeklyAtNext(t *testing.T) {
	t.Parallel()
	warsaw, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatal(err)
	}
	at := func(loc *time.Location, s string) time.Time {
		tt, err := time.ParseInLocation("2006-01-02 15:04", s, loc)
		if err != nil {
			t.Fatal(err)
		}
		return tt
	}
	tests := []struct {
		name string
		loc  *time.Location
		from string
		want string
	}{
		{name: "from a Friday to the next Monday", loc: time.UTC, from: "2026-11-13 12:00", want: "2026-11-16 06:30"},
		{name: "from Monday before the time to that day", loc: time.UTC, from: "2026-11-16 05:00", want: "2026-11-16 06:30"},
		{name: "from Monday at the time to the next week", loc: time.UTC, from: "2026-11-16 06:30", want: "2026-11-23 06:30"},
		{name: "from Monday after the time to the next week", loc: time.UTC, from: "2026-11-16 09:00", want: "2026-11-23 06:30"},
		{name: "across the end of a month", loc: time.UTC, from: "2026-11-30 07:00", want: "2026-12-07 06:30"},
		{name: "in the core's time zone", loc: warsaw, from: "2026-11-13 12:00", want: "2026-11-16 06:30"},
		{name: "across the clock change", loc: warsaw, from: "2027-03-22 07:00", want: "2027-03-29 06:30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := jobs.WeeklyTime(tt.loc).Next(at(tt.loc, tt.from))
			if want := at(tt.loc, tt.want); !got.Equal(want) {
				t.Errorf("Next(%s) = %s, want %s", tt.from, got, want)
			}
		})
	}
}

func TestWorkerRunsTheRunOrTheWeeklyTurn(t *testing.T) {
	t.Parallel()
	w, late := newWorker(t)
	r := &fakeRunner{}
	late.Set(r)
	if err := w.Work(t.Context(), job("00112233445566aa")); err != nil {
		t.Fatalf("Work of an owner's run: %v", err)
	}
	if err := w.Work(t.Context(), job("")); err != nil {
		t.Fatalf("Work of the weekly turn: %v", err)
	}
	if len(r.worked) != 1 || r.worked[0] != "00112233445566aa" || r.weekly != 1 {
		t.Errorf("worked %v, weekly %d; want the owner's run once and the weekly turn once", r.worked, r.weekly)
	}
}

func TestWorkerWaitsUntilTheMaintainerIsBuilt(t *testing.T) {
	t.Parallel()
	w, _ := newWorker(t)
	err := w.Work(t.Context(), job("00112233445566aa"))
	if _, ok := errors.AsType[*rivertype.JobSnoozeError](err); !ok {
		t.Errorf("Work before Set = %v, want a snooze", err)
	}
}

func TestWorkerHandsRiverNoText(t *testing.T) {
	t.Parallel()
	w, late := newWorker(t)
	late.Set(&fakeRunner{err: errQuote})
	err := w.Work(t.Context(), job("00112233445566aa"))
	if err == nil {
		t.Fatal("a failed run was a success")
	}
	if strings.Contains(err.Error(), "ETFA") || strings.Contains(err.Error(), "63.1") || strings.Contains(err.Error(), "note") {
		t.Errorf("the job's error quotes the failure: %v", err)
	}
	if _, ok := errors.AsType[*jobs.JobError](err); !ok {
		t.Errorf("error = %T, want *jobs.JobError", err)
	}
}

func TestPeriodicWeeklyJobHasOneAttemptAndNoRun(t *testing.T) {
	t.Parallel()
	args, opts := jobs.WeeklyJob()
	if args.RunID != "" {
		t.Errorf("the schedule's job = %+v, want the weekly turn (no run ID)", args)
	}
	if opts == nil || opts.MaxAttempts != 1 {
		t.Errorf("the schedule's insert options = %+v, want one attempt", opts)
	}
}

func TestNewNeedsWhatItRuns(t *testing.T) {
	t.Parallel()
	if _, _, err := jobs.New(nil, jobs.WeeklyTime(time.UTC), slog.New(slog.DiscardHandler)); !errors.Is(err, jobs.ErrInvalidJob) {
		t.Errorf("New without a runner: %v, want ErrInvalidJob", err)
	}
	if _, _, err := jobs.New(&jobs.Late{}, nil, slog.New(slog.DiscardHandler)); !errors.Is(err, jobs.ErrInvalidJob) {
		t.Errorf("New without a schedule: %v, want ErrInvalidJob", err)
	}
}
