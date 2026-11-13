package openfief_test

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

var (
	errRule   = errors.New("a rule of the plugin")
	errOther  = errors.New("another rule")
	errSecret = errors.New("SECRET-REPLY-TEXT")
)

func TestLogSafe(t *testing.T) {
	t.Parallel()
	secret := errSecret
	tests := []struct {
		name  string
		err   error
		known []error
		want  string
	}{
		{name: "nil", want: ""},
		{name: "a known error, wrapped around a secret", err: fmt.Errorf("%w: %w", errRule, secret), known: []error{errRule}, want: "a rule of the plugin"},
		{name: "two known errors", err: fmt.Errorf("%w: %w", errRule, errOther), known: []error{errRule, errOther}, want: "a rule of the plugin; another rule"},
		{name: "an unknown error is its type only", err: fmt.Errorf("reading: %w", secret), want: "unclassified error of type *errors.errorString"},
		{name: "the type of the innermost error", err: fmt.Errorf("opening: %w", &fs.PathError{Op: "open", Path: "/secret", Err: fs.ErrNotExist}), want: "unclassified error of type *errors.errorString"},
		{name: "joined errors, the first one", err: errors.Join(secret, errOther), known: nil, want: "unclassified error of type *errors.errorString"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := openfief.LogSafe(tt.err, tt.known...); got != tt.want {
				t.Errorf("LogSafe = %q, want %q", got, tt.want)
			}
		})
	}
}
