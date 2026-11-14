package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/ingest"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

var (
	errLastStepBroke = errors.New("the last step broke")
	errImportBroke   = errors.New("the import broke")
	errTookNoID      = errors.New("the database took an account with no ID")
)

// outboxFixture is a migrated database with the outbox over it, whose clock
// stands still.
type outboxFixture struct {
	pool *pgxpool.Pool
	out  *outbox
	at   time.Time
}

func newOutboxFixture(t *testing.T) *outboxFixture {
	t.Helper()
	dsn := emptyDatabase(t)
	pool, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(pool.Close)
	if err = migrate(t.Context(), pool); err != nil {
		t.Fatalf("migrating: %v", err)
	}
	at := time.Date(2026, time.March, 2, 10, 0, 0, 0, time.UTC)
	out, err := newOutbox(pool, func() time.Time { return at })
	if err != nil {
		t.Fatalf("newOutbox: %v", err)
	}
	return &outboxFixture{pool: pool, out: out, at: at}
}

// jobs returns the audit-entry jobs the outbox holds, oldest first.
func (f *outboxFixture) jobs(t *testing.T) []auditEntryArgs {
	t.Helper()
	rows, err := f.pool.Query(t.Context(), `SELECT args FROM river.river_job WHERE kind = $1 ORDER BY id`, auditEntryArgs{}.Kind())
	if err != nil {
		t.Fatalf("listing the audit jobs: %v", err)
	}
	defer rows.Close()
	var out []auditEntryArgs
	for rows.Next() {
		var raw []byte
		var a auditEntryArgs
		if err = rows.Scan(&raw); err != nil {
			t.Fatalf("reading a job: %v", err)
		}
		if err = json.Unmarshal(raw, &a); err != nil {
			t.Fatalf("decoding a job: %v", err)
		}
		out = append(out, a)
	}
	return out
}

// accounts counts the ledger accounts named like prefix.
func (f *outboxFixture) accounts(t *testing.T, prefix string) int {
	t.Helper()
	var n int
	if err := f.pool.QueryRow(t.Context(), `SELECT count(*) FROM ledger.accounts WHERE id LIKE $1`, prefix+"%").Scan(&n); err != nil {
		t.Fatalf("counting accounts: %v", err)
	}
	return n
}

// entry returns a valid entry of the owner deleting account at f's time.
func (f *outboxFixture) entry(t *testing.T, account string) *identity.AuditEntry {
	t.Helper()
	e, err := identity.NewAuditEntry(f.at, identity.Actor{Kind: identity.ActorOwner, ID: "0123456789abcdef"}, identity.ActionAccountDeleted,
		identity.OutcomeOK, identity.Ref{Kind: identity.RefAccount, ID: account})
	if err != nil {
		t.Fatalf("NewAuditEntry: %v", err)
	}
	return &e
}

// refuseAuditJobs makes the database refuse every audit-entry job, as a
// queue that is down would.
func (f *outboxFixture) refuseAuditJobs(t *testing.T) {
	t.Helper()
	if _, err := f.pool.Exec(t.Context(), "ALTER TABLE river.river_job ADD CONSTRAINT no_audit_jobs CHECK (kind <> 'identity.audit_entry') NOT VALID"); err != nil {
		t.Fatalf("breaking the enqueue: %v", err)
	}
}

// insertAccount writes a ledger account through the pool that joins the
// transaction ctx carries.
func insertAccount(ctx context.Context, pool *pgxpool.Pool, id string) error {
	_, err := joined(pool).Exec(ctx, `INSERT INTO ledger.accounts (id, institution, name) VALUES ($1, 'bank', $1)`, id)
	if err != nil {
		return fmt.Errorf("inserting %s: %w", id, err)
	}
	return nil
}

// TestOutboxChangeThatIsMadeLeavesItsWritesAndOneJob checks over a real
// database (D-211) that a change writing two accounts of the ledger leaves
// both and one job, with the change's time.
func TestOutboxChangeThatIsMadeLeavesItsWritesAndOneJob(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	err := f.out.Within(t.Context(), f.entry(t, "made-1"), func(ctx context.Context) error {
		return errors.Join(insertAccount(ctx, f.pool, "made-1"), insertAccount(ctx, f.pool, "made-2"))
	})
	if err != nil || f.accounts(t, "made-") != 2 {
		t.Fatalf("Within = %v with %d accounts, want both written", err, f.accounts(t, "made-"))
	}
	jobs := f.jobs(t)
	if len(jobs) != 1 || !jobs[0].At.Equal(f.at) || jobs[0].Action != "account_deleted" {
		t.Errorf("jobs = %+v, want the one entry with the change's time", jobs)
	}
}

// TestOutboxChangeThatFailsLeavesNothing checks that a change that fails after
// its writes leaves neither them nor a job.
func TestOutboxChangeThatFailsLeavesNothing(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	err := f.out.Within(t.Context(), f.entry(t, "gone-1"), func(ctx context.Context) error {
		if err := insertAccount(ctx, f.pool, "gone-1"); err != nil {
			return err
		}
		return errLastStepBroke
	})
	if !errors.Is(err, errLastStepBroke) || f.accounts(t, "gone-") != 0 || len(f.jobs(t)) != 0 {
		t.Errorf("Within = %v with %d accounts and %d jobs, want the error, no account and no job", err, f.accounts(t, "gone-"), len(f.jobs(t)))
	}
}

// TestOutboxStepOfOneStoreThatFailsRollsTheWholeChangeBack checks that a
// statement of a second store that the database refuses rolls back what the
// first wrote, leaves no job and no transaction open: the next change goes
// through.
func TestOutboxStepOfOneStoreThatFailsRollsTheWholeChangeBack(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	err := f.out.Within(t.Context(), f.entry(t, "step-1"), func(ctx context.Context) error {
		if err := insertAccount(ctx, f.pool, "step-1"); err != nil {
			return err
		}
		// A statement of another context's table that the database refuses: no name.
		if _, err := joined(f.pool).Exec(ctx, `INSERT INTO portfolio.sleeve_kinds (key, name, default_role) VALUES ('step-kind', NULL, 'core')`); err != nil {
			return fmt.Errorf("the portfolio step: %w", err)
		}
		return nil
	})
	if err == nil || f.accounts(t, "step-") != 0 || len(f.jobs(t)) != 0 {
		t.Fatalf("Within = %v with %d accounts and %d jobs, want an error and nothing kept", err, f.accounts(t, "step-"), len(f.jobs(t)))
	}
	next := f.out.Within(t.Context(), f.entry(t, "after-1"), func(ctx context.Context) error { return insertAccount(ctx, f.pool, "after-1") })
	if next != nil {
		t.Errorf("the change after a failed one: %v", next)
	}
}

// TestOutboxStepThatFailsAndIsHandledCostsNothingElse checks that a statement
// that fails inside a savepoint the change rolled back itself leaves the rest
// of the change as it was.
func TestOutboxStepThatFailsAndIsHandledCostsNothingElse(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	err := f.out.Within(t.Context(), f.entry(t, "sp-1"), func(ctx context.Context) error {
		if err := insertAccount(ctx, f.pool, "sp-1"); err != nil {
			return err
		}
		tx, err := joined(f.pool).Begin(ctx)
		if err != nil {
			return fmt.Errorf("opening the savepoint: %w", err)
		}
		if _, execErr := tx.Exec(ctx, `INSERT INTO ledger.accounts (id, institution, name) VALUES (NULL, 'bank', 'x')`); execErr == nil {
			return errTookNoID
		}
		if err = tx.Rollback(ctx); err != nil {
			return fmt.Errorf("rolling the savepoint back: %w", err)
		}
		return nil
	})
	if err != nil || f.accounts(t, "sp-") != 1 {
		t.Errorf("Within = %v with %d accounts, want the change made around the handled failure", err, f.accounts(t, "sp-"))
	}
}

// TestOutboxEnqueueThatIsRefusedFailsTheChange checks that an enqueue the
// database refuses fails the change with ErrAuditUnavailable and rolls its
// writes back.
func TestOutboxEnqueueThatIsRefusedFailsTheChange(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	f.refuseAuditJobs(t)
	called := false
	err := f.out.Within(t.Context(), f.entry(t, "refused-1"), func(ctx context.Context) error {
		called = true
		return insertAccount(ctx, f.pool, "refused-1")
	})
	if !called || !errors.Is(err, kernel.ErrAuditUnavailable) || f.accounts(t, "refused-") != 0 {
		t.Errorf("Within = %v (change run %v) with %d accounts, want ErrAuditUnavailable and the change rolled back", err, called, f.accounts(t, "refused-"))
	}
}

// TestOutboxEnqueueNeedsTheTransactionOfAChange checks that an entry can't be
// enqueued outside a change's transaction: it would be written whether or not
// the change was.
func TestOutboxEnqueueNeedsTheTransactionOfAChange(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	if err := f.out.Enqueue(t.Context(), f.entry(t, "x-1")); !errors.Is(err, kernel.ErrAuditUnavailable) || !errors.Is(err, ErrNoChangeTx) {
		t.Errorf("Enqueue outside a transaction = %v, want ErrAuditUnavailable and ErrNoChangeTx", err)
	}
}

// syncCase is one sync of TestSyncEntryOnlyWhenTheSyncChangedState and the
// entry it leaves, as "<kind>:<ID>" references; none for no entry.
type syncCase struct {
	name     string
	change   ingestapp.SyncChange
	err      error
	wantRefs []string
}

// syncCases are the syncs the test runs.
func syncCases() []syncCase {
	many := make([]ledger.AccountID, 0, 30)
	for i := range 30 {
		many = append(many, ledger.AccountID(fmt.Sprintf("acc-%02d", i)))
	}
	return []syncCase{
		{name: "nothing new", change: ingestapp.SyncChange{Instance: "ibkr-main", Batch: "b-0123456789abcdef"}},
		{name: "new transaction versions", change: ingestapp.SyncChange{
			Instance: "ibkr-main", Batch: "b-0123456789abcdef", Transactions: 3, Accounts: []ledger.AccountID{"ibkr"},
		}, wantRefs: []string{"plugin:ibkr-main", "import_batch:b-0123456789abcdef", "account:ibkr", "sync_count:transactions:3"}},
		{
			name: "prices that changed", change: ingestapp.SyncChange{Instance: "prices-eu", Batch: "b-1", Prices: 12},
			wantRefs: []string{"plugin:prices-eu", "import_batch:b-1", "sync_count:prices:12"},
		},
		{
			name: "bond terms that changed", change: ingestapp.SyncChange{Instance: "bonds-main", Batch: "b-2", BondTerms: 2},
			wantRefs: []string{"plugin:bonds-main", "import_batch:b-2", "sync_count:bond_terms:2"},
		},
		{
			name: "FX rates that changed", change: ingestapp.SyncChange{Instance: "nbp-fx", Batch: "b-6", FXRates: 4},
			wantRefs: []string{"plugin:nbp-fx", "import_batch:b-6", "sync_count:fx_rates:4"},
		},
		{
			name: "index levels that changed", change: ingestapp.SyncChange{Instance: "indices", Batch: "b-7", IndexLevels: 250},
			wantRefs: []string{"plugin:indices", "import_batch:b-7", "sync_count:index_levels:250"},
		},
		{
			name: "index readings that changed", change: ingestapp.SyncChange{Instance: "gus-cpi", Batch: "b-8", IndexReadings: 1},
			wantRefs: []string{"plugin:gus-cpi", "import_batch:b-8", "sync_count:index_readings:1"},
		},
		{name: "every kind", change: ingestapp.SyncChange{
			Instance: "ibkr-main", Batch: "b-3", Transactions: 1, Prices: 2, FXRates: 3, IndexReadings: 4, IndexLevels: 5, BondTerms: 6,
			Accounts: []ledger.AccountID{"a", "b"},
		}, wantRefs: []string{
			"plugin:ibkr-main", "import_batch:b-3", "account:a", "account:b",
			"sync_count:transactions:1", "sync_count:prices:2", "sync_count:fx_rates:3", "sync_count:index_readings:4",
			"sync_count:index_levels:5", "sync_count:bond_terms:6",
		}},
		{name: "more accounts than an entry names", change: ingestapp.SyncChange{
			Instance: "ibkr-main", Batch: "b-4", Transactions: 30, Accounts: many,
		}, wantRefs: []string{"plugin:ibkr-main", "import_batch:b-4", "sync_count:transactions:30"}},
		{name: "a sync that fails", err: errImportBroke, change: ingestapp.SyncChange{Instance: "ibkr-main", Batch: "b-5", Transactions: 1}},
	}
}

// TestSyncEntryOnlyWhenTheSyncChangedState checks the entry of a sync the core
// started (D-198, D-211): it is the system's, names the instance, the batch,
// the accounts and counts, never an amount, and is enqueued in the transaction
// of the sync's writes only when they changed state, so a sync that brought
// nothing new leaves none and a sync that fails leaves none.
func TestSyncEntryOnlyWhenTheSyncChangedState(t *testing.T) {
	t.Parallel()
	for _, tt := range syncCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newOutboxFixture(t)
			err := f.out.AuditedSync(t.Context(), tt.change.Instance, func(context.Context) (ingestapp.SyncChange, error) {
				return tt.change, tt.err
			})
			if !errors.Is(err, tt.err) {
				t.Fatalf("AuditedSync = %v, want %v", err, tt.err)
			}
			checkSyncJobs(t, f.jobs(t), tt.wantRefs, f.at)
		})
	}
}

// checkSyncJobs fails t unless jobs holds no entry, when wantRefs is nil, or
// the one entry of the system's sync, made at the sync's time, that names
// wantRefs.
func checkSyncJobs(t *testing.T, jobs []auditEntryArgs, wantRefs []string, when time.Time) {
	t.Helper()
	if wantRefs == nil {
		if len(jobs) != 0 {
			t.Fatalf("jobs = %+v, want none: nothing changed", jobs)
		}
		return
	}
	if len(jobs) != 1 {
		t.Fatalf("jobs = %+v, want one entry", jobs)
	}
	checkSyncJob(t, &jobs[0], wantRefs, when)
}

// checkSyncJob fails t unless j is the entry of the system's sync, made at the
// sync's time, that names wantRefs and no amount.
func checkSyncJob(t *testing.T, j *auditEntryArgs, wantRefs []string, when time.Time) {
	t.Helper()
	if j.ActorKind != "system" || j.ActorID != "core" || j.Action != "connector_synced" || j.Outcome != "ok" || !j.At.Equal(when) {
		t.Errorf("entry = %+v, want the system's connector_synced, ok, at the sync's time", j)
	}
	refs := refTexts(j.Refs)
	if diff := cmp.Diff(wantRefs, refs); diff != "" {
		t.Errorf("refs (-want +got):\n%s", diff)
	}
	if strings.ContainsAny(strings.Join(refs, " "), "€$") {
		t.Errorf("refs %v carry an amount", refs)
	}
}

// refTexts returns refs as "<kind>:<ID>".
func refTexts(refs []auditRef) []string {
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		out = append(out, r.Kind+":"+r.ID)
	}
	return out
}

// TestSyncEntryIsAValidEntry checks the entry the sync builds is one the log
// accepts, whatever it names, and survives the job that carries it.
func TestSyncEntryIsAValidEntry(t *testing.T) {
	t.Parallel()
	c := ingestapp.SyncChange{
		Instance: "ibkr-main", Batch: ingest.BatchID("b-0123456789abcdef"), Transactions: 1, Prices: 1, FXRates: 1, IndexReadings: 1,
		IndexLevels: 1, BondTerms: 1, Accounts: []ledger.AccountID{"ibkr"},
	}
	e, err := syncEntry(time.Date(2026, time.March, 2, 10, 0, 0, 0, time.UTC), "ibkr-main", &c)
	if err != nil {
		t.Fatalf("syncEntry: %v", err)
	}
	if validateErr := e.Validate(); validateErr != nil {
		t.Errorf("Validate: %v", validateErr)
	}
	args := auditArgsOf(&e)
	back, err := entryOf(&args)
	if err != nil || cmp.Diff(e, back) != "" {
		t.Errorf("the job's entry = %+v, %v; want the entry it was made from", back, err)
	}
}

// trailOf returns the audit log of dsn after the outbox's pending entries were
// written, as "<actor> <action> <refs>" lines, oldest first.
func trailOf(t *testing.T, dsn string) []string {
	t.Helper()
	pool, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	defer pool.Close()
	drainAudit(t, pool)
	rows, err := pool.Query(t.Context(), `SELECT actor_kind || ' ' || action || ' ' || refs::text FROM identity.audit_log ORDER BY id`)
	if err != nil {
		t.Fatalf("reading the audit log: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var line string
		if err = rows.Scan(&line); err != nil {
			t.Fatalf("reading an entry: %v", err)
		}
		out = append(out, line)
	}
	return out
}

// TestSyncDeliveryThatRunsPastItsTimeoutFailsWhole checks over a real database
// (D-18, D-198, D-211) that a delivery whose transaction has a statement that
// runs past the delivery timeout fails, with an error that says so and that the
// attempt records as the class delivery-timeout, and writes nothing: not the
// rows its earlier statements wrote, and no entry.
func TestSyncDeliveryThatRunsPastItsTimeoutFailsWhole(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	f.out.deliveryTimeout = 200 * time.Millisecond
	err := f.out.AuditedSync(t.Context(), "ibkr-main", f.delivery("slow-1", 5))
	if !errors.Is(err, ingestapp.ErrDeliveryTimeout) {
		t.Fatalf("AuditedSync = %v, want an error wrapping ErrDeliveryTimeout", err)
	}
	if class := ingestapp.SyncErrorClassOf(err); class != ingest.SyncErrorDeliveryTimeout {
		t.Errorf("the attempt's class = %v, want delivery-timeout", class)
	}
	if n, jobs := f.accounts(t, "slow-"), f.jobs(t); n != 0 || len(jobs) != 0 {
		t.Errorf("a delivery that timed out left %d accounts and entries %+v, want nothing", n, jobs)
	}
}

// delivery returns a sync's writes that store the account name and then run a
// statement for sleep seconds, reporting a change of one transaction.
func (f *outboxFixture) delivery(name string, sleep float64) func(context.Context) (ingestapp.SyncChange, error) {
	change := ingestapp.SyncChange{Instance: "ibkr-main", Batch: "b-0123456789abcdef", Transactions: 1, Accounts: []ledger.AccountID{"ibkr"}}
	return func(ctx context.Context) (ingestapp.SyncChange, error) {
		if err := insertAccount(ctx, f.pool, name); err != nil {
			return ingestapp.SyncChange{}, err
		}
		if _, err := joined(f.pool).Exec(ctx, `SELECT pg_sleep($1)`, sleep); err != nil {
			return ingestapp.SyncChange{}, fmt.Errorf("a statement that runs %vs: %w", sleep, err)
		}
		return change, nil
	}
}

// TestSyncDeliveryInTimeCommitsAndLeavesNoTimeoutBehind checks that a delivery
// whose statements finish in time commits as before, and that the timeout, set
// for the transaction alone, does not outlive it.
func TestSyncDeliveryInTimeCommitsAndLeavesNoTimeoutBehind(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	f.out.deliveryTimeout = 5 * time.Second
	if err := f.out.AuditedSync(t.Context(), "ibkr-main", f.delivery("quick-1", 0.01)); err != nil {
		t.Fatalf("AuditedSync of a delivery in time: %v", err)
	}
	if n, jobs := f.accounts(t, "quick-"), len(f.jobs(t)); n != 1 || jobs != 1 {
		t.Errorf("a delivery in time left %d accounts and %d entries, want one each", n, jobs)
	}
	var after string
	if err := f.pool.QueryRow(t.Context(), `SHOW statement_timeout`).Scan(&after); err != nil || after != "0" {
		t.Errorf("statement_timeout after the delivery = %q, %v; want the connection's own, 0", after, err)
	}
}
