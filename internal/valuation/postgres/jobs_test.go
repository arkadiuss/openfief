package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/arkadiuss/openfief/internal/valuation/postgres"
)

var errRegister = errors.New("a worker that can't be registered")

// otherArgs is the job of another context that shares the client.
type otherArgs struct{}

func (otherArgs) Kind() string { return "other.job" }

// otherJob is a valid job of another context on its own queue, with a
// periodic insert, which registers its worker once.
func otherJob(registered *int) postgres.Job {
	return postgres.Job{
		Queue: "other", MaxWorkers: 1,
		Register: func(w *river.Workers) error {
			*registered++
			return river.AddWorkerSafely(w, river.WorkFunc(func(context.Context, *river.Job[otherArgs]) error { return nil }))
		},
		Periodic: []*river.PeriodicJob{river.NewPeriodicJob(postgres.DailyAt{Hour: 6, Location: time.UTC},
			func() (river.JobArgs, *river.InsertOpts) { return otherArgs{}, nil }, &river.PeriodicJobOpts{ID: "other.job"})},
	}
}

func TestNewClientTakesOtherContextsJobs(t *testing.T) {
	t.Parallel()
	t.Run("a job on a queue of its own is registered once", func(t *testing.T) {
		t.Parallel()
		registered := 0
		cfg := config(t, "2026-03-02")
		cfg.Jobs = []postgres.Job{otherJob(&registered)}
		// A pool connects on first use: building the client never touches the network.
		pool, err := pgxpool.New(t.Context(), "postgres://openfief@127.0.0.1:1/none")
		if err != nil {
			t.Fatalf("pool: %v", err)
		}
		t.Cleanup(pool.Close)
		client, err := postgres.NewClient(pool, nil, etfBuy(t), cfg)
		if err != nil || client == nil {
			t.Fatalf("NewClient = %v, %v", client, err)
		}
		if registered != 1 {
			t.Errorf("the job's workers were registered %d times, want once", registered)
		}
	})
	tests := []struct {
		name    string
		jobs    func(registered *int) []postgres.Job
		wantErr error
	}{
		{name: "the revaluations' queue", jobs: func(r *int) []postgres.Job {
			j := otherJob(r)
			j.Queue = postgres.Queue
			return []postgres.Job{j}
		}, wantErr: postgres.ErrInvalidConfig},
		{name: "no queue", jobs: func(r *int) []postgres.Job {
			j := otherJob(r)
			j.Queue = ""
			return []postgres.Job{j}
		}, wantErr: postgres.ErrInvalidConfig},
		{name: "no workers", jobs: func(r *int) []postgres.Job {
			j := otherJob(r)
			j.MaxWorkers = 0
			return []postgres.Job{j}
		}, wantErr: postgres.ErrInvalidConfig},
		{name: "nothing to register", jobs: func(r *int) []postgres.Job {
			j := otherJob(r)
			j.Register = nil
			return []postgres.Job{j}
		}, wantErr: postgres.ErrInvalidConfig},
		{name: "two jobs on one queue", jobs: func(r *int) []postgres.Job {
			return []postgres.Job{otherJob(r), otherJob(r)}
		}, wantErr: postgres.ErrInvalidConfig},
		{name: "a worker that fails to register", jobs: func(*int) []postgres.Job {
			return []postgres.Job{{Queue: "other", MaxWorkers: 1, Register: func(*river.Workers) error { return errRegister }}}
		}, wantErr: errRegister},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			registered := 0
			cfg := config(t, "2026-03-02")
			cfg.Jobs = tt.jobs(&registered)
			if _, err := postgres.NewClient(nil, nil, etfBuy(t), cfg); !errors.Is(err, tt.wantErr) {
				t.Errorf("NewClient error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
