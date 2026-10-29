package cli_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/identity/cli"
)

// syntheticLink is what the fake core's enroll-link prints.
const syntheticLink = "http://localhost:8080/enroll#code=ofenroll_synthetic"

var (
	errSynthetic = errors.New("synthetic failure")
	errNoCluster = errors.New("synthetic: no cluster")
)

// fakeCore records what the command asked of the core.
type fakeCore struct {
	mu          sync.Mutex
	forwardErr  error // returned at once instead of forwarding
	linkErr     error
	port        int
	host        string
	ttl         time.Duration
	forwardDone bool
	// linked is closed once EnrollLink was called.
	linked chan struct{}
}

func (f *fakeCore) Forward(ctx context.Context, localPort int, ready chan struct{}) error {
	f.mu.Lock()
	f.port = localPort
	f.mu.Unlock()
	if f.forwardErr != nil {
		return f.forwardErr
	}
	close(ready)
	<-ctx.Done()
	f.mu.Lock()
	f.forwardDone = true
	f.mu.Unlock()
	return nil
}

func (f *fakeCore) EnrollLink(_ context.Context, host string, ttl time.Duration) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.host, f.ttl = host, ttl
	close(f.linked)
	return syntheticLink, f.linkErr
}

// syncBuffer is a bytes.Buffer safe for one writer and one reader.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, err := s.b.Write(p)
	if err != nil {
		return n, fmt.Errorf("buffer: %w", err)
	}
	return n, nil
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func TestEnroll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		args     []string
		wantPort int
		wantHost string
		wantTTL  time.Duration
		wantWarn bool
	}{
		{name: "defaults", wantPort: 8080, wantHost: "localhost:8080", wantTTL: 15 * time.Minute},
		{name: "another port and ttl", args: []string{"--port", "9090", "--ttl", "1h"}, wantPort: 9090, wantHost: "localhost:9090", wantTTL: time.Hour, wantWarn: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			core := &fakeCore{linked: make(chan struct{})}
			var out syncBuffer
			ctx, cancel := context.WithCancel(t.Context())
			done := make(chan error, 1)
			go func() {
				done <- cli.Run(ctx, tt.args, &cli.Env{Out: &out, Connect: func(string, string) (cli.Core, error) { return core, nil }})
			}()
			<-core.linked
			cancel() // the owner presses Ctrl-C
			if err := <-done; err != nil {
				t.Fatalf("Run: %v", err)
			}
			core.mu.Lock()
			defer core.mu.Unlock()
			if core.port != tt.wantPort || core.host != tt.wantHost || core.ttl != tt.wantTTL || !core.forwardDone {
				t.Errorf("forwarded %d, link for %q lasting %v, forward stopped %v; want %d, %q, %v, stopped",
					core.port, core.host, core.ttl, core.forwardDone, tt.wantPort, tt.wantHost, tt.wantTTL)
			}
			checkOutput(t, out.String(), tt.wantHost, tt.wantWarn)
		})
	}
}

// checkOutput checks what `openfief enroll` told the owner.
func checkOutput(t *testing.T, text, host string, warn bool) {
	t.Helper()
	for _, want := range []string{syntheticLink, "first-run wizard", "Ctrl-C", "http://" + host} {
		if !strings.Contains(text, want) {
			t.Errorf("output lacks %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "--auth-origin") != warn {
		t.Errorf("output warns about --auth-origin: %v, want %v:\n%s", !warn, warn, text)
	}
}

func TestEnrollFails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		core    *fakeCore
		connect error
		wantErr error
	}{
		{name: "an unknown flag", args: []string{"--host", "x"}, core: &fakeCore{}, wantErr: cli.ErrUsage},
		{name: "an argument", args: []string{"now"}, core: &fakeCore{}, wantErr: cli.ErrUsage},
		{name: "not a port", args: []string{"--port", "70000"}, core: &fakeCore{}, wantErr: cli.ErrUsage},
		{name: "no cluster", core: &fakeCore{}, connect: errNoCluster, wantErr: errNoCluster},
		{name: "the forward fails", core: &fakeCore{forwardErr: errSynthetic}, wantErr: errSynthetic},
		{name: "no link", core: &fakeCore{linkErr: errSynthetic, linked: make(chan struct{})}, wantErr: errSynthetic},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var out syncBuffer
			err := cli.Run(t.Context(), tt.args, &cli.Env{Out: &out, Connect: func(string, string) (cli.Core, error) {
				if tt.connect != nil {
					return nil, tt.connect
				}
				return tt.core, nil
			}})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Run = %v, want %v", err, tt.wantErr)
			}
			if strings.Contains(out.String(), syntheticLink) {
				t.Errorf("printed a link after a failure:\n%s", out.String())
			}
		})
	}
}
