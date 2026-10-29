package cli_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/identity/cli"
)

// routeOrigin is the owner's own route in these tests.
const routeOrigin = "https://fief.example.org"

// TestEnrollOrigin checks D-80's --origin: for an origin the core allows,
// the command prints the link for that origin and forwards no port.
func TestEnrollOrigin(t *testing.T) {
	t.Parallel()
	core := &fakeCore{linked: make(chan struct{}), allowed: []string{"http://localhost:8080", routeOrigin}}
	var out syncBuffer
	err := cli.Run(t.Context(), []string{"--origin", routeOrigin, "--ttl", "1h"},
		&cli.Env{Out: &out, Connect: func(string, string) (cli.Core, error) { return core, nil }})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	core.mu.Lock()
	defer core.mu.Unlock()
	if core.port != 0 || core.host != routeOrigin || core.ttl != time.Hour {
		t.Errorf("forwarded port %d, link for %q lasting %v; want no forward, %q, 1h", core.port, core.host, core.ttl, routeOrigin)
	}
	text := out.String()
	for _, want := range []string{syntheticLink, "first-run wizard", "passkey on " + routeOrigin} {
		if !strings.Contains(text, want) {
			t.Errorf("output lacks %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "Ctrl-C") || strings.Contains(text, "Forwarding") {
		t.Errorf("output speaks of a port-forward:\n%s", text)
	}
}

func TestEnrollOriginRefused(t *testing.T) {
	t.Parallel()
	local := []string{"http://localhost:8080"}
	tests := []struct {
		name    string
		args    []string
		core    *fakeCore
		wantErr error
	}{
		{name: "an origin the core doesn't allow", args: []string{"--origin", routeOrigin}, core: &fakeCore{allowed: local}, wantErr: cli.ErrOriginNotAllowed},
		{name: "--origin with --port", args: []string{"--origin", routeOrigin, "--port", "9090"}, core: &fakeCore{allowed: local}, wantErr: cli.ErrUsage},
		{name: "a host, not an origin", args: []string{"--origin", "fief.example.org"}, core: &fakeCore{allowed: local}, wantErr: cli.ErrUsage},
		{name: "the core's origins can't be read", args: []string{"--origin", routeOrigin}, core: &fakeCore{allowedErr: errSynthetic}, wantErr: errSynthetic},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var out syncBuffer
			err := cli.Run(t.Context(), tt.args, &cli.Env{Out: &out, Connect: func(string, string) (cli.Core, error) { return tt.core, nil }})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Run = %v, want %v", err, tt.wantErr)
			}
			tt.core.mu.Lock()
			defer tt.core.mu.Unlock()
			if tt.core.host != "" || tt.core.port != 0 || strings.Contains(out.String(), syntheticLink) {
				t.Errorf("issued a link for %q or forwarded port %d after a refusal:\n%s", tt.core.host, tt.core.port, out.String())
			}
		})
	}
}
