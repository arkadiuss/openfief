package kernel_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Errors for TestLogSafe: two known rules and a store's failure that the
// failures wrap with values.
var (
	errRuleFailed = errors.New("the rule that failed")
	errOtherRule  = errors.New("another known rule")
	errQuotedRow  = errors.New("store down")
)

// TestLogSafe checks that an error is described for a log line by the known
// errors in its chain only, or else by its innermost error's type, never by
// a value its text carries (GR-12).
func TestLogSafe(t *testing.T) {
	t.Parallel()
	known, other := errRuleFailed, errOtherRule
	values := "1234.56 PLN of PLTEST000010 in U1234567"
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "nil", err: nil, want: ""},
		{name: "a known error wrapped with values", err: fmt.Errorf("posting %s: %w", values, known), want: "the rule that failed"},
		{
			name: "two known errors joined", err: errors.Join(fmt.Errorf("%s: %w", values, known), other),
			want: "the rule that failed; another known rule",
		},
		{
			name: "an unknown error quoting a row", err: fmt.Errorf("row %s: %w", values, errQuotedRow),
			want: "unclassified error of type *errors.errorString",
		},
		{
			name: "an unknown typed error", err: fmt.Errorf("wrapping: %w", &kernel.CurrencyMismatchError{}),
			want: "unclassified error of type *kernel.CurrencyMismatchError",
		},
		{
			name: "unknown errors joined: the first one's type", err: errors.Join(fmt.Errorf("%s: %w", values, &kernel.CurrencyMismatchError{}), errQuotedRow),
			want: "unclassified error of type *kernel.CurrencyMismatchError",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := kernel.LogSafe(tt.err, known, other); got != tt.want {
				t.Errorf("LogSafe() = %q, want %q", got, tt.want)
			}
		})
	}
}
