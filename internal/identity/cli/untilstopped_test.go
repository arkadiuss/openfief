package cli

import (
	"context"
	"errors"
	"testing"
)

var errSyntheticForward = errors.New("synthetic: forward failed")

// TestUntilStopped covers the moment both a stop and the forward's end are
// ready: select picks between them at random, so each case repeats until
// both branches have surely been taken.
func TestUntilStopped(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		stopped bool  // the owner pressed Ctrl-C before the forward ended
		ferr    error // what the forward returned
		wantErr error
	}{
		{name: "stopped, forward ended cleanly", stopped: true, ferr: nil, wantErr: nil},
		{name: "stopped, forward ended with an error", stopped: true, ferr: errSyntheticForward, wantErr: errSyntheticForward},
		{name: "forward ended on its own", stopped: false, ferr: nil, wantErr: errForwardEnded},
		{name: "forward failed on its own", stopped: false, ferr: errSyntheticForward, wantErr: errForwardEnded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for range 200 {
				err := stopOnce(t, tt.stopped, tt.ferr)
				if (tt.wantErr == nil) != (err == nil) || !errors.Is(err, tt.wantErr) {
					t.Fatalf("untilStopped: %v, want %v", err, tt.wantErr)
				}
			}
		})
	}
}

// stopOnce runs untilStopped once with the forward's end already waiting,
// and the context cancelled first when stopped.
func stopOnce(t *testing.T, stopped bool, ferr error) error {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	if stopped {
		cancel()
	}
	forwarded := make(chan error, 1)
	forwarded <- ferr
	return untilStopped(ctx, forwarded)
}
