package postgres_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/riverqueue/river"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// leakedValue matches what neither a log line nor a stored job error may
// hold (GR-12): an amount, as digits with a decimal point or comma; an
// ISIN; or an IBKR-style account number.
var leakedValue = regexp.MustCompile(`\d+[.,]\d+|[A-Z]{2}[A-Z0-9]{9}\d|U\d{7}`)

// errLedgerDown is the failure failingLedger's valued errors wrap.
var errLedgerDown = errors.New("ledger down")

// failingLedger is a ledger whose logs can't be read: each read fails with
// err, which carries synthetic values.
type failingLedger struct {
	memLedger

	err error
}

func (f *failingLedger) Log(context.Context, ledger.AccountID) ([]ledger.VersionedTransaction, error) {
	return nil, f.err
}

// lockedBuffer is a log sink River's goroutines may write concurrently.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n, err := b.buf.Write(p)
	if err != nil {
		return n, fmt.Errorf("log buffer: %w", err)
	}
	return n, nil
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// valuedFailures are revaluation failures that carry synthetic values: a
// store's error quoting a row, and an oversell naming the instrument, the
// disposal and both quantities.
func valuedFailures(t *testing.T) []error {
	t.Helper()
	return []error{
		fmt.Errorf("reading 1234.56 PLN of PLTEST000010 in U1234567: %w", errLedgerDown),
		&ledger.OversellError{
			Account: acc1, Instrument: instrument(t, etfISIN, "SYNTH"), Date: date(t, "2026-03-02"),
			Disposal:  ledger.TransactionKey{Source: "test", ExternalID: "U1234567"},
			Requested: kernel.NewQuantity(dec(t, "12.5")), Available: kernel.NewQuantity(dec(t, "10.25")),
		},
	}
}

// TestFailedJobsKeepNoValues checks GR-12 on the revaluation workers'
// path: a revaluation that fails with an error carrying an amount, an ISIN
// and an account number reaches River's log and the job's stored errors
// (river_job.errors) named by its kind only.
func TestFailedJobsKeepNoValues(t *testing.T) {
	t.Parallel()
	for i, failure := range valuedFailures(t) {
		t.Run(fmt.Sprintf("failure %d", i), func(t *testing.T) {
			t.Parallel()
			checkFailedJobKeepsNoValues(t, failure)
		})
	}
}

// checkFailedJobKeepsNoValues runs one revaluation that fails with
// failure through real River workers, and checks that River's log and the
// job's stored errors name the job and hold none of failure's values.
func checkFailedJobKeepsNoValues(t *testing.T, failure error) {
	t.Helper()
	store, pool := newStore(t)
	ctx := t.Context()
	logged := &lockedBuffer{}
	cfg := config(t, "2026-03-04")
	cfg.Logger = slog.New(slog.NewTextHandler(logged, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: withoutTime}))
	client, err := postgres.NewClient(pool, store, &failingLedger{memLedger: *etfBuy(t), err: failure}, cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if enqueueErr := store.Enqueue(ctx, []valuation.Revaluation{{Account: acc1, From: date(t, "2026-03-02")}}); enqueueErr != nil {
		t.Fatalf("Enqueue: %v", enqueueErr)
	}
	waitForFailed(t, startClient(t, client))
	var stored string
	err = pool.QueryRow(ctx, `SELECT array_to_string(ARRAY(SELECT e->>'error' FROM unnest(errors) e), ' | ') FROM river.river_job WHERE kind = $1`,
		postgres.RevalueAccount{}.Kind()).Scan(&stored)
	if err != nil {
		t.Fatalf("reading the job's errors: %v", err)
	}
	lines := errorLines(logged.String())
	if !strings.Contains(stored, "revaluation job") || !strings.Contains(lines, "revaluation job") {
		t.Errorf("the failed job is not named: stored %q, logged:\n%s", stored, logged.String())
	}
	for _, text := range []string{stored, lines} {
		if m := leakedValue.FindAllString(text, -1); len(m) > 0 {
			t.Errorf("values %q kept in:\n%s", m, text)
		}
	}
}

// withoutTime drops a log line's time, whose fraction of a second reads as
// an amount.
func withoutTime(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}

// waitForFailed waits, up to jobDeadline, until a job fails.
func waitForFailed(t *testing.T, events <-chan *river.Event) {
	t.Helper()
	deadline := time.After(jobDeadline)
	for {
		select {
		case e := <-events:
			if e.Kind == river.EventKindJobFailed {
				return
			}
		case <-deadline:
			t.Fatalf("no job failed within %s", jobDeadline)
		}
	}
}

// errorLines returns the log lines that carry an error attribute, River's
// failed-job lines among them; the others can hold durations, which read
// as amounts.
func errorLines(logged string) string {
	var out []string
	for line := range strings.Lines(logged) {
		if strings.Contains(line, "error=") || strings.Contains(line, "err=") {
			out = append(out, line)
		}
	}
	return strings.Join(out, "")
}
