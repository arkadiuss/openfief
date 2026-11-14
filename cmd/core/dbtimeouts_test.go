package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDBTimeoutsFromEnv checks the session timeouts of the core's database
// connections (D-211): a sane default for each, the environment's value when it
// is set, zero to turn one off, and a value that is no duration refused.
func TestDBTimeoutsFromEnv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		env     map[string]string
		want    dbTimeouts
		wantErr error
	}{
		{name: "defaults", want: dbTimeouts{Statement: 2 * time.Minute, IdleInTransaction: 2 * time.Minute}},
		{
			name: "both set", env: map[string]string{dbStatementTimeoutEnv: "90s", dbIdleInTransactionTimeoutEnv: "30s"},
			want: dbTimeouts{Statement: 90 * time.Second, IdleInTransaction: 30 * time.Second},
		},
		{
			name: "one set", env: map[string]string{dbStatementTimeoutEnv: "10m"},
			want: dbTimeouts{Statement: 10 * time.Minute, IdleInTransaction: 2 * time.Minute},
		},
		{
			name: "off", env: map[string]string{dbStatementTimeoutEnv: "0", dbIdleInTransactionTimeoutEnv: "0s"},
			want: dbTimeouts{},
		},
		{name: "not a duration", env: map[string]string{dbStatementTimeoutEnv: "soon"}, wantErr: errBadDBTimeout},
		{name: "negative", env: map[string]string{dbIdleInTransactionTimeoutEnv: "-1s"}, wantErr: errBadDBTimeout},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := dbTimeoutsFromEnv(func(k string) string { return tt.env[k] })
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("dbTimeoutsFromEnv error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("timeouts (-want +got):\n%s", diff)
			}
		})
	}
}

// TestPoolConfigSetsTheTimeoutsOnEveryConnection checks that the pool's
// connections start with both timeouts in milliseconds, whichever form the
// connection string takes, and that a timeout the connection string itself
// names, the owner's own database setting, wins over the core's.
func TestPoolConfigSetsTheTimeoutsOnEveryConnection(t *testing.T) {
	t.Parallel()
	timeouts := dbTimeouts{Statement: 2 * time.Minute, IdleInTransaction: 45 * time.Second}
	tests := []struct {
		name string
		dsn  string
		want map[string]string
	}{
		{
			name: "a URL", dsn: "postgres://synthetic@192.0.2.1:5432/openfief",
			want: map[string]string{"statement_timeout": "120000", "idle_in_transaction_session_timeout": "45000"},
		},
		{
			name: "key and value", dsn: "host=192.0.2.1 user=synthetic dbname=openfief",
			want: map[string]string{"statement_timeout": "120000", "idle_in_transaction_session_timeout": "45000"},
		},
		{
			name: "the connection string names one", dsn: "postgres://synthetic@192.0.2.1:5432/openfief?statement_timeout=5000",
			want: map[string]string{"statement_timeout": "5000", "idle_in_transaction_session_timeout": "45000"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg, err := poolConfig(tt.dsn, timeouts)
			if err != nil {
				t.Fatalf("poolConfig: %v", err)
			}
			got := map[string]string{}
			for _, k := range []string{"statement_timeout", "idle_in_transaction_session_timeout"} {
				got[k] = cfg.ConnConfig.RuntimeParams[k]
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("runtime parameters (-want +got):\n%s", diff)
			}
			if cfg.ConnConfig.Tracer == nil {
				t.Error("the pool's connections have no query tracer")
			}
		})
	}
}

// show returns the value of the session setting name on pool.
func show(t *testing.T, pool *pgxpool.Pool, name string) string {
	t.Helper()
	var got string
	if err := pool.QueryRow(t.Context(), `SELECT current_setting($1)`, name).Scan(&got); err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return got
}

// TestPoolStartsEveryConnectionWithTheTimeouts checks over a real database
// (D-211) that a pool from newPool starts every connection with the default
// timeouts, and with the environment's value for the one it sets.
func TestPoolStartsEveryConnectionWithTheTimeouts(t *testing.T) {
	dsn := emptyDatabase(t)
	t.Setenv(dbIdleInTransactionTimeoutEnv, "45s")
	pool, err := newPool(t.Context(), dsn)
	if err != nil {
		t.Fatalf("newPool: %v", err)
	}
	t.Cleanup(pool.Close)
	if got := show(t, pool, "statement_timeout"); got != "2min" {
		t.Errorf("statement_timeout = %q, want the default 2min", got)
	}
	if got := show(t, pool, "idle_in_transaction_session_timeout"); got != "45s" {
		t.Errorf("idle_in_transaction_session_timeout = %q, want the environment's 45s", got)
	}
}

// idleTx is a transaction the database may end under the test.
type idleTx struct{ pgx.Tx }

// idleTransaction begins a transaction on pool that has written a table, and
// closes it when the test ends: the database ends an idle one, but the pool only
// gets its connection back when the transaction is closed on this side too.
func idleTransaction(t *testing.T, pool *pgxpool.Pool) *idleTx {
	t.Helper()
	tx, err := pool.Begin(t.Context())
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	t.Cleanup(func() {
		if rollbackErr := tx.Rollback(context.WithoutCancel(t.Context())); rollbackErr != nil {
			t.Logf("closing the ended transaction: %v", rollbackErr)
		}
	})
	if _, err = tx.Exec(t.Context(), `CREATE TABLE public.idle_probe (id int)`); err != nil {
		t.Fatalf("writing in the transaction: %v", err)
	}
	return &idleTx{Tx: tx}
}

// endedByIdleTimeout reports whether err is the database ending a transaction
// for idling past idle_in_transaction_session_timeout, or the connection it
// closed.
func endedByIdleTimeout(err error) bool {
	if err == nil {
		return false
	}
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return !ok || pgErr.Code == "25P03"
}

// TestIdleTransactionIsEndedByTheDatabase checks that a transaction left idle
// past the idle timeout is ended by the database, which rolls it back, so a
// change whose process hangs holds no lock for long.
func TestIdleTransactionIsEndedByTheDatabase(t *testing.T) {
	dsn := emptyDatabase(t)
	t.Setenv(dbIdleInTransactionTimeoutEnv, "300ms")
	pool, err := newPool(t.Context(), dsn)
	if err != nil {
		t.Fatalf("newPool: %v", err)
	}
	t.Cleanup(pool.Close)
	tx := idleTransaction(t, pool)
	// The transaction sits idle while another connection works for longer than
	// the idle timeout.
	if _, err = pool.Exec(t.Context(), `SELECT pg_sleep(1)`); err != nil {
		t.Fatalf("working elsewhere: %v", err)
	}
	_, err = tx.Exec(t.Context(), `INSERT INTO public.idle_probe VALUES (1)`)
	if !endedByIdleTimeout(err) {
		t.Errorf("the idle transaction went on or failed with %v, want it ended by 25P03 idle_in_transaction_session_timeout", err)
	}
	var left int
	if err = pool.QueryRow(t.Context(), `SELECT count(*) FROM pg_tables WHERE tablename = 'idle_probe'`).Scan(&left); err != nil || left != 0 {
		t.Errorf("the ended transaction left %d tables, %v; want it rolled back", left, err)
	}
}
