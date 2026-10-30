package postgres_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
	"github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// jobDeadline bounds every wait for a job to finish.
const jobDeadline = 30 * time.Second

// memLedger is the ledger's accounts and logs in memory.
type memLedger struct {
	logs map[ledger.AccountID][]ledger.VersionedTransaction
}

func (m *memLedger) Log(_ context.Context, account ledger.AccountID) ([]ledger.VersionedTransaction, error) {
	return m.logs[account], nil
}

func (m *memLedger) Accounts(context.Context) ([]ledger.Account, error) {
	out := make([]ledger.Account, 0, len(m.logs))
	for id := range m.logs {
		out = append(out, ledger.Account{ID: id})
	}
	return out, nil
}

// etfBuy is acc-1 buying 10 units of the synthetic ETF for 990 USD on
// 2026-03-02.
func etfBuy(t *testing.T) *memLedger {
	t.Helper()
	h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: "b1"}, Account: acc1, Date: date(t, "2026-03-02")}
	tx, err := ledger.NewBuy(h, instrument(t, etfISIN, "SYNTH"), kernel.NewQuantity(dec(t, "10")), money(t, "990", "USD"), money(t, "0", "USD"))
	if err != nil {
		t.Fatalf("NewBuy: %v", err)
	}
	return &memLedger{logs: map[ledger.AccountID][]ledger.VersionedTransaction{acc1: {{Transaction: tx, Version: 1}}}}
}

func config(t *testing.T, today string) *postgres.Config {
	t.Helper()
	d := date(t, today)
	return &postgres.Config{
		Concurrency: 2,
		EndOfDay:    postgres.DailyAt{Hour: 23, Minute: 30, Location: time.UTC},
		Today:       func() kernel.Date { return d },
		Logger:      slog.New(slog.DiscardHandler),
		Pack:        pl.NewPack(),
	}
}

// waitForCompleted waits, up to jobDeadline, until n jobs have completed.
func waitForCompleted(t *testing.T, events <-chan *river.Event, n int) {
	t.Helper()
	deadline := time.After(jobDeadline)
	for done := 0; done < n; {
		select {
		case e := <-events:
			if e.Kind == river.EventKindJobCompleted {
				done++
				continue
			}
			t.Fatalf("job %s ended %s: %v", e.Job.Kind, e.Kind, e.Job.Errors)
		case <-deadline:
			t.Fatalf("only %d of %d jobs completed within %s", done, n, jobDeadline)
		}
	}
}

// startClient starts a worker client and returns its job events; it stops
// when the test ends.
func startClient(t *testing.T, client *river.Client[pgx.Tx]) <-chan *river.Event {
	t.Helper()
	events, cancel := client.Subscribe(river.EventKindJobCompleted, river.EventKindJobFailed, river.EventKindJobCancelled)
	t.Cleanup(cancel)
	if err := client.Start(t.Context()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Stop(context.WithoutCancel(t.Context())); err != nil {
			t.Errorf("Stop: %v", err)
		}
	})
	return events
}

// etfPrice is a USD close of the synthetic ETF.
func etfPrice(t *testing.T, on, closing string) valuation.Price {
	t.Helper()
	return valuation.Price{Instrument: instrument(t, etfISIN, "SYNTH").String(), Date: date(t, on), Close: money(t, closing, "USD"), Source: "vendor"}
}

// valuedDays returns acc-1's stored valuations as "date PLN value", or
// "date missing reason".
func valuedDays(t *testing.T, store *postgres.Store) []string {
	t.Helper()
	got, err := store.Valuations(t.Context(), acc1, date(t, "2026-03-01"), date(t, "2026-03-31"))
	if err != nil {
		t.Fatalf("Valuations: %v", err)
	}
	out := make([]string, 0, len(got))
	for i := range got {
		v := got[i].PLN.String()
		if !got[i].Valued() {
			v = "missing " + got[i].Missing.String()
		}
		out = append(out, got[i].Date.String()+" "+v)
	}
	return out
}

// checkStatus compares acc-1's valuation status with an idle one valued
// through through.
func checkStatus(t *testing.T, store *postgres.Store, through, when string) {
	t.Helper()
	got, err := store.Status(t.Context(), acc1)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if diff := cmp.Diff(app.ValuationStatus{Account: acc1, ValuedThrough: date(t, through)}, got); diff != "" {
		t.Errorf("status %s (-want +got):\n%s", when, diff)
	}
}

func TestWorkerRevaluesEndToEnd(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	l := etfBuy(t)
	client, err := postgres.NewClient(pool, store, l, config(t, "2026-03-04"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	recorder := app.NewInputRecorder(app.NewPlanner(l, l), store, pl.NewPack())
	usd, err := kernel.ParseCurrency("USD")
	if err != nil {
		t.Fatal(err)
	}
	err = recorder.RecordFXRates(ctx, []valuation.FXRate{{Currency: usd, Date: date(t, "2026-03-02"), Rate: dec(t, "4"), Source: "nbp-a"}})
	if err == nil {
		err = recorder.RecordPrices(ctx, []valuation.Price{etfPrice(t, "2026-03-03", "100")})
	}
	if err != nil {
		t.Fatalf("recording inputs: %v", err)
	}
	checkPending(t, pool, []string{"acc-1 2026-03-02 -"}, []string{"acc-1 available"})

	events := startClient(t, client)
	waitForCompleted(t, events, 1)
	want := []string{"2026-03-02 missing price", "2026-03-03 4000 PLN", "2026-03-04 4000 PLN"}
	if diff := cmp.Diff(want, valuedDays(t, store)); diff != "" {
		t.Errorf("valuations (-want +got):\n%s", diff)
	}
	checkStatus(t, store, "2026-03-01", "after the run")
	checkPending(t, pool, nil, nil)

	// The missing price arrives: its own revaluation closes the gap.
	if err := recorder.RecordPrices(ctx, []valuation.Price{etfPrice(t, "2026-03-02", "99")}); err != nil {
		t.Fatalf("RecordPrices: %v", err)
	}
	waitForCompleted(t, events, 1)
	checkStatus(t, store, "2026-03-04", "after the missing price")

	// The end-of-day job enqueues every account again from its
	// valued-through date.
	if _, err := client.Insert(ctx, postgres.EndOfDay{}, nil); err != nil {
		t.Fatalf("Insert end of day: %v", err)
	}
	waitForCompleted(t, events, 2)
	want = []string{"2026-03-02 3960 PLN", "2026-03-03 4000 PLN", "2026-03-04 4000 PLN"}
	if diff := cmp.Diff(want, valuedDays(t, store)); diff != "" {
		t.Errorf("valuations after end of day (-want +got):\n%s", diff)
	}
	checkPending(t, pool, nil, nil)
}

func TestDailyAt(t *testing.T) {
	t.Parallel()
	warsaw, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Skipf("no time zone database: %v", err)
	}
	at := postgres.DailyAt{Hour: 23, Minute: 30, Location: warsaw}
	tests := []struct {
		name    string
		current time.Time
		want    time.Time
	}{
		{name: "later the same day", current: time.Date(2026, 3, 2, 12, 0, 0, 0, warsaw), want: time.Date(2026, 3, 2, 23, 30, 0, 0, warsaw)},
		{name: "exactly then: the next day", current: time.Date(2026, 3, 2, 23, 30, 0, 0, warsaw), want: time.Date(2026, 3, 3, 23, 30, 0, 0, warsaw)},
		{name: "a UTC instant past the time in Warsaw", current: time.Date(2026, 3, 2, 22, 45, 0, 0, time.UTC), want: time.Date(2026, 3, 3, 23, 30, 0, 0, warsaw)},
		{name: "across the switch to summer time", current: time.Date(2026, 3, 29, 0, 0, 0, 0, warsaw), want: time.Date(2026, 3, 29, 23, 30, 0, 0, warsaw)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := at.Next(tt.current); !got.Equal(tt.want) {
				t.Errorf("Next(%s) = %s, want %s", tt.current, got, tt.want)
			}
		})
	}
}

func TestNewClientChecksTheConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		change  func(c *postgres.Config)
		wantErr error
	}{
		{name: "no workers", change: func(c *postgres.Config) { c.Concurrency = 0 }, wantErr: postgres.ErrInvalidConfig},
		{name: "no time zone", change: func(c *postgres.Config) { c.EndOfDay.Location = nil }, wantErr: postgres.ErrInvalidConfig},
		{name: "hour 24", change: func(c *postgres.Config) { c.EndOfDay.Hour = 24 }, wantErr: postgres.ErrInvalidConfig},
		{name: "no date source", change: func(c *postgres.Config) { c.Today = nil }, wantErr: postgres.ErrInvalidConfig},
		{name: "no jurisdiction pack", change: func(c *postgres.Config) { c.Pack = nil }, wantErr: postgres.ErrInvalidConfig},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := config(t, "2026-03-02")
			tt.change(cfg)
			if _, err := postgres.NewClient(nil, nil, etfBuy(t), cfg); !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestTodayIn(t *testing.T) {
	t.Parallel()
	warsaw, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Skipf("no time zone database: %v", err)
	}
	now := func() time.Time { return time.Date(2026, 3, 2, 23, 30, 0, 0, time.UTC) }
	if got := postgres.TodayIn(warsaw, now)(); !got.Equal(date(t, "2026-03-03")) {
		t.Errorf("TodayIn = %s, want 2026-03-03", got)
	}
}
