package connectapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"testing"
)

// Synthetic values a failure may carry, which no log line may show (GR-12):
// an amount, an ISIN and an IBKR-style account number.
const valuedText = "1234.56 PLN of PLTEST000010 in U1234567"

// leakedValue matches what a log line must never hold: an amount, as digits
// with a decimal point or comma; an ISIN; or an IBKR-style account number.
var leakedValue = regexp.MustCompile(`\d+[.,]\d+|[A-Z]{2}[A-Z0-9]{9}\d|U\d{7}`)

// errStoreDown is a store's failure that the failures below wrap.
var errStoreDown = errors.New("store down")

// valuedFailures are failures that reach a sink's log line, each carrying
// the synthetic values: a store's error quoting a row, joined errors, and a
// deadline that wraps them.
func valuedFailures() []error {
	return []error{
		fmt.Errorf("loading %s: %w", valuedText, errStoreDown),
		errors.Join(fmt.Errorf("row %s: %w", valuedText, errStoreDown), errStoreDown),
		fmt.Errorf("%w: while reading %s", context.DeadlineExceeded, valuedText),
	}
}

// textLog returns a logger writing text to buf, without the time, whose
// fraction of a second looks like an amount.
func textLog(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}))
}

// checkNoValues fails t when logged holds fewer than want lines or any of
// the synthetic values.
func checkNoValues(t *testing.T, logged string, want int) {
	t.Helper()
	if got := strings.Count(logged, "\n"); got < want {
		t.Fatalf("logged %d lines, want %d:\n%s", got, want, logged)
	}
	if m := leakedValue.FindAllString(logged, -1); len(m) > 0 {
		t.Errorf("the log holds values %q:\n%s", m, logged)
	}
}

// TestFailuresLogNoValues drives every log sink of this package's services
// with failures that carry an amount, an ISIN and an account number, and
// checks that the log names each failure by its kind only (GR-12).
func TestFailuresLogNoValues(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	n := 0
	for _, err := range valuedFailures() {
		s := &sleeveService{log: textLog(&buf)}
		if got := s.fail(t.Context(), err); got == nil {
			t.Error("a failure mapped to no error")
		}
		n++
	}
	for _, err := range valuedFailures() {
		s := &service{log: textLog(&buf)}
		if got := s.fail(t.Context(), err); got == nil {
			t.Error("a failure mapped to no error")
		}
		n++
	}
	checkNoValues(t, buf.String(), n)
}
