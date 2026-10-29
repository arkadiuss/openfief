package postgres_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
	"github.com/arkadiuss/openfief/internal/valuation/postgres"
)

var errImportFails = errors.New("synthetic import failure")

func TestMigrateTwice(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	if err := postgres.Migrate(t.Context(), pool); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
	n := countRows(t, pool, `
		SELECT count(*) FROM information_schema.tables
		WHERE (table_schema, table_name) IN (('river', 'river_job'), ('valuation', 'holding_valuations'), ('valuation', 'dirty'))`)
	if n != 3 {
		t.Errorf("tables found = %d, want 3", n)
	}
}

// pricesOf returns one price on 2026-03-02 per close, each of its own
// synthetic instrument.
func pricesOf(t *testing.T, closes []string) []valuation.Price {
	t.Helper()
	etf := instrument(t, etfISIN, "SYNTH").String()
	out := make([]valuation.Price, 0, len(closes))
	for i, c := range closes {
		out = append(out, valuation.Price{
			Instrument: etf + strconv.Itoa(i), Date: date(t, "2026-03-02"), Close: money(t, c, "USD"), Source: "vendor",
		})
	}
	return out
}

func TestSaveEnqueuesInTheSameTransaction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		closes     []string
		wantErr    bool
		wantPrices int
		wantDirty  []string
		wantJobs   []string
	}{
		{
			name: "the job commits with the price", closes: []string{"100"}, wantPrices: 1,
			wantDirty: []string{"acc-1 2026-03-02 -"}, wantJobs: []string{"acc-1 available"},
		},
		{name: "a failing write rolls the job back", closes: []string{"100", "-1"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store, pool := newStore(t)
			err := store.SavePrices(t.Context(), pricesOf(t, tt.closes), []valuation.Revaluation{reval(t, acc1, "2026-03-02")})
			if (err != nil) != tt.wantErr {
				t.Fatalf("SavePrices error = %v, want error %v", err, tt.wantErr)
			}
			if n := countRows(t, pool, "SELECT count(*) FROM valuation.prices"); n != tt.wantPrices {
				t.Errorf("prices stored = %d, want %d", n, tt.wantPrices)
			}
			checkPending(t, pool, tt.wantDirty, tt.wantJobs)
		})
	}
}

func TestEnqueueCoalescesFromTheEarliestDate(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	for _, batch := range [][]valuation.Revaluation{
		{reval(t, acc1, "2026-03-10")},
		{reval(t, acc1, "2026-03-01"), reval(t, acc2, "2026-03-07")},
		{reval(t, acc1, "2026-03-05"), reval(t, acc1, "2026-03-03")},
	} {
		if err := store.Enqueue(t.Context(), batch); err != nil {
			t.Fatalf("Enqueue: %v", err)
		}
	}
	checkPending(t, pool, []string{"acc-1 2026-03-01 -", "acc-2 2026-03-07 -"}, []string{"acc-1 available", "acc-2 available"})
}

// ledgerTx stands in for the ledger's store inside an import: the hook
// reaches the import's database transaction only through WithTx.
type ledgerTx struct {
	ledgerapp.Store

	tx pgx.Tx
}

func (l *ledgerTx) WithTx(fn func(tx pgx.Tx) error) error { return fn(l.tx) }

func TestImportHookEnqueuesInTheImportsTransaction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		fail      error
		wantDirty []string
		wantJobs  []string
	}{
		{name: "the jobs commit with the import", wantDirty: []string{"acc-1 2026-03-02 -", "acc-2 2026-03-04 -"}, wantJobs: []string{"acc-1 available", "acc-2 available"}},
		{name: "a failed import rolls them back", fail: errImportFails},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store, pool := newStore(t)
			sum := &ledgerapp.ImportSummary{EarliestAffected: map[ledger.AccountID]kernel.Date{
				acc2: date(t, "2026-03-04"), acc1: date(t, "2026-03-02"),
			}}
			err := pgx.BeginFunc(t.Context(), pool, func(tx pgx.Tx) error {
				if err := store.ImportHook()(t.Context(), &ledgerTx{tx: tx}, sum); err != nil {
					return err
				}
				return tt.fail
			})
			if diff := cmp.Diff(tt.fail, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("import (-want +got):\n%s", diff)
			}
			checkPending(t, pool, tt.wantDirty, tt.wantJobs)
		})
	}
}

// plainStore is a ledger store that cannot hand over a transaction.
type plainStore struct{ ledgerapp.Store }

func TestImportHookNeedsTheLedgersTransaction(t *testing.T) {
	t.Parallel()
	store, err := postgres.NewStore(nil)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	hook := store.ImportHook()
	if err := hook(t.Context(), plainStore{}, &ledgerapp.ImportSummary{}); err != nil {
		t.Errorf("an import that changed nothing: %v", err)
	}
	sum := &ledgerapp.ImportSummary{EarliestAffected: map[ledger.AccountID]kernel.Date{acc1: date(t, "2026-03-02")}}
	if err := hook(t.Context(), plainStore{}, sum); !cmp.Equal(postgres.ErrNoLedgerTx, err, cmpopts.EquateErrors()) {
		t.Errorf("error = %v, want ErrNoLedgerTx", err)
	}
}

func TestSaveRulesVersionRevaluesOnlyOnAChange(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	revals := []valuation.Revaluation{reval(t, acc1, "2026-03-02")}
	steps := []struct {
		version     string
		wantChanged bool
	}{
		{version: "pl/1", wantChanged: true},
		{version: "pl/1"},
		{version: "pl/2", wantChanged: true},
	}
	for i, st := range steps {
		if _, err := pool.Exec(t.Context(), "DELETE FROM valuation.dirty; DELETE FROM river.river_job"); err != nil {
			t.Fatal(err)
		}
		changed, err := store.SaveRulesVersion(t.Context(), st.version, revals)
		if err != nil {
			t.Fatalf("step %d: SaveRulesVersion: %v", i, err)
		}
		if jobs := liveJobs(t, pool); changed != st.wantChanged || (len(jobs) == 1) != st.wantChanged {
			t.Errorf("step %d (%s): changed = %v with jobs %v, want %v", i, st.version, changed, jobs, st.wantChanged)
		}
	}
}

// statusStep is one change TestStatuses makes, and the statuses after it.
type statusStep struct {
	name    string
	enqueue []valuation.Revaluation
	sql     string
	want    []app.ValuationStatus
}

func TestStatuses(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	if err := store.ReplaceValuations(ctx, acc1, date(t, "2026-03-02"), nil, date(t, "2026-03-03")); err != nil {
		t.Fatalf("ReplaceValuations: %v", err)
	}
	valued := app.ValuationStatus{Account: acc1, ValuedThrough: date(t, "2026-03-03")}
	waiting := valued
	waiting.Pending = true
	steps := []statusStep{
		{name: "valued, nothing pending", want: []app.ValuationStatus{valued}},
		{
			name: "changes wait", enqueue: []valuation.Revaluation{reval(t, acc1, "2026-03-03"), reval(t, acc2, "2026-03-01")},
			want: []app.ValuationStatus{waiting, {Account: acc2, Pending: true}},
		},
		{
			name: "acc-2 runs", sql: `UPDATE valuation.dirty SET claimed_from = from_date, from_date = NULL WHERE account_id = 'acc-2';
				UPDATE river.river_job SET state = 'running', attempt = 1, attempted_at = now() WHERE args->>'account' = 'acc-2'`,
			want: []app.ValuationStatus{waiting, {Account: acc2, Running: true}},
		},
		{
			name: "acc-2 failed and waits for a retry",
			sql:  `UPDATE river.river_job SET state = 'retryable', scheduled_at = now() + interval '1 hour' WHERE args->>'account' = 'acc-2'`,
			want: []app.ValuationStatus{waiting, {Account: acc2, Pending: true}},
		},
	}
	for _, st := range steps {
		applyStatusStep(t, store, pool, &st)
		got, err := store.Statuses(ctx)
		if err != nil {
			t.Fatalf("%s: Statuses: %v", st.name, err)
		}
		if diff := cmp.Diff(st.want, got); diff != "" {
			t.Errorf("%s (-want +got):\n%s", st.name, diff)
		}
	}
	got, err := store.Status(ctx, "acc-9")
	if err != nil || !cmp.Equal(app.ValuationStatus{Account: "acc-9"}, got) {
		t.Errorf("Status of an unknown account = %+v, %v", got, err)
	}
}

// applyStatusStep enqueues st's revaluations and runs its SQL.
func applyStatusStep(t *testing.T, store *postgres.Store, pool *pgxpool.Pool, st *statusStep) {
	t.Helper()
	if len(st.enqueue) > 0 {
		if err := store.Enqueue(t.Context(), st.enqueue); err != nil {
			t.Fatalf("%s: Enqueue: %v", st.name, err)
		}
	}
	if st.sql != "" {
		if _, err := pool.Exec(t.Context(), st.sql); err != nil {
			t.Fatalf("%s: %v", st.name, err)
		}
	}
}
