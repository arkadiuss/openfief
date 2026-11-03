package app_test

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Synthetic values a failure may carry, which no log line may show (GR-12):
// an amount, an ISIN and an IBKR-style account number.
const (
	syntheticAmount  = "1234.56"
	syntheticISIN    = "PLTEST000010"
	syntheticAccount = "U1234567"
)

// leaked matches what a log line must never hold: an amount, as digits with
// a decimal point or comma; an ISIN; or an IBKR-style account number.
var leaked = regexp.MustCompile(`\d+[.,]\d+|[A-Z]{2}[A-Z0-9]{9}\d|U\d{7}`)

// valued returns the failures a sync or an upload may meet, each carrying
// the synthetic values.
func valued() []error {
	values := syntheticAmount + " PLN of " + syntheticISIN + " in " + syntheticAccount
	return []error{
		fmt.Errorf("deliver from x batch b-1: %w: posting %s", ledger.ErrInvalidTransaction, values),
		fmt.Errorf("sync x: %w: field close %q", app.ErrInvalidDelivery, values),
		fmt.Errorf("%w: the connector said %s", app.ErrEntryRefused, values),
		fmt.Errorf("storing: %w", &ledger.OversellError{}),
		errors.Join(fmt.Errorf("%w: row %s", errStore, values), ingest.ErrAccountNotBound),
		fmt.Errorf("%w: duplicate key (isin)=(%s), amount %s", errStore, syntheticISIN, syntheticAmount),
	}
}

// lockedBuffer is a bytes.Buffer the scheduler's goroutines may share.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n, err := b.buf.Write(p)
	if err != nil {
		return n, fmt.Errorf("buffer: %w", err)
	}
	return n, nil
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// textLogger logs to w as text, without the time, whose fraction of a
// second looks like an amount.
func textLogger(w *lockedBuffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}))
}

func TestLogSafe(t *testing.T) {
	t.Parallel()
	want := []string{
		ledger.ErrInvalidTransaction.Error(),
		app.ErrInvalidDelivery.Error(),
		app.ErrEntryRefused.Error(),
		"unclassified error of type *ledger.OversellError",
		ingest.ErrAccountNotBound.Error(),
		"unclassified error of type *errors.errorString",
	}
	for i, err := range valued() {
		got := app.LogSafe(err)
		if got != want[i] {
			t.Errorf("LogSafe(failure %d) = %q, want %q", i, got, want[i])
		}
		if leaked.MatchString(got) {
			t.Errorf("LogSafe(failure %d) = %q holds a value", i, got)
		}
	}
	if got := app.LogSafe(nil); got != "" {
		t.Errorf("LogSafe(nil) = %q", got)
	}
	if got := app.LogSafe(fmt.Errorf("parsing: %w", kernel.ErrInvalidDecimal)); got != kernel.ErrInvalidDecimal.Error() {
		t.Errorf("LogSafe(a decimal error) = %q", got)
	}
}

// TestSchedulerLogsNoValues drives syncs that fail with errors carrying an
// amount, an ISIN and an account number, and checks that the scheduler's
// log names the failures without any of them (GR-12).
func TestSchedulerLogsNoValues(t *testing.T) {
	t.Parallel()
	failures := valued()
	var out lockedBuffer
	for _, err := range failures {
		caller, store := &scriptedCaller{errs: []error{err}}, newMemStore()
		s, _ := scheduler(t, caller, store, "xtb-main")
		s.Log = textLogger(&out)
		tick(t, s)
	}
	logged := out.String()
	if got := strings.Count(logged, "connector sync failed"); got != len(failures) {
		t.Fatalf("logged %d failed syncs, want %d:\n%s", got, len(failures), logged)
	}
	if m := leaked.FindAllString(logged, -1); len(m) > 0 {
		t.Errorf("the log holds values %q:\n%s", m, logged)
	}
}
