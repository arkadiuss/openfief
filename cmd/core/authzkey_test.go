package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/spicedb"
)

// waitDeadline bounds a wait for the gate in a test; it is met at once.
const waitDeadline = 10 * time.Second

// gateFor returns a gate over a key file in a fresh directory, which is its
// cfg.KeyFile, with the log's output in out.
func gateFor(t *testing.T, out *bytes.Buffer) *keyGate {
	t.Helper()
	path := filepath.Join(t.TempDir(), "key")
	cfg := authzConfig{Mode: authzSpiceDB, Endpoint: "127.0.0.1:1", KeyFile: path, Insecure: true, Timeout: time.Second}
	// A line per log call, as the production handler writes.
	log := slog.New(slog.NewTextHandler(out, nil))
	return newKeyGate(&cfg, time.Millisecond, log)
}

// requireRefused fails t unless az, over g, refuses every check and every
// change with the pending error, and g is not connected.
func requireRefused(t *testing.T, g *keyGate, az authorization, when string) {
	t.Helper()
	owner := identity.Caller{Holder: identity.HolderOwner, ID: "s"}
	ok, err := az.authorizer.MayAct(t.Context(), owner, identity.ActionView, "brokerage")
	if ok || !errors.Is(err, errSpiceDBKeyPending) {
		t.Errorf("%s: MayAct = %v, %v, want false and errSpiceDBKeyPending", when, ok, err)
	}
	if err := az.syncer.ProjectGrants(t.Context(), nil); !errors.Is(err, errSpiceDBKeyPending) {
		t.Errorf("%s: ProjectGrants = %v, want errSpiceDBKeyPending", when, err)
	}
	if err := az.syncer.AccountPlaced(t.Context(), "brokerage", "core"); !errors.Is(err, errSpiceDBKeyPending) {
		t.Errorf("%s: AccountPlaced = %v, want errSpiceDBKeyPending", when, err)
	}
	if g.isConnected() {
		t.Errorf("%s: connected", when)
	}
}

// TestKeyGateRefusesUntilTheKeyIsThere checks that a core without
// SpiceDB's key refuses every check, closed, whoever asks, and every change
// it would pass on (D-164), whether the file is missing, empty or only
// whitespace.
func TestKeyGateRefusesUntilTheKeyIsThere(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	g := gateFor(t, &out)
	path := g.cfg.KeyFile
	az := gatedAuthorization(g)
	if g.step() {
		t.Fatal("step connected without a key file")
	}
	requireRefused(t, g, az, "no file")
	for _, content := range []string{"", "  \n\t\n"} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if g.step() {
			t.Fatalf("step connected with the file holding %q", content)
		}
		requireRefused(t, g, az, fmt.Sprintf("file holding %q", content))
	}
}

// TestKeyGateConnectsWithTheKey checks that the gate turns to SpiceDB's
// answers when the key appears in the file.
func TestKeyGateConnectsWithTheKey(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	g := gateFor(t, &out)
	path := g.cfg.KeyFile
	az := gatedAuthorization(g)
	if err := os.WriteFile(path, []byte(testKey+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !g.step() || !g.isConnected() {
		t.Fatal("the gate did not connect with the key in the file")
	}
	// SpiceDB at the placeholder address is down: the check is refused with its error, not the pending one.
	owner := identity.Caller{Holder: identity.HolderOwner, ID: "s"}
	if ok, err := az.authorizer.MayAct(t.Context(), owner, identity.ActionView, "brokerage"); ok || err == nil || errors.Is(err, errSpiceDBKeyPending) {
		t.Errorf("MayAct after the key = %v, %v, want false and SpiceDB's error", ok, err)
	}
	if !g.wait(t.Context()) {
		t.Error("wait = false after the key")
	}
	if err := g.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

// TestKeyGateAccountLister checks that the list of accounts a caller may act
// on is refused too, while the key is pending.
func TestKeyGateAccountLister(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	g := gateFor(t, &out)
	got, err := gatedAuthorizer{g}.ActableAccounts(t.Context(), identity.Caller{Holder: identity.HolderOwner, ID: "s"}, identity.ActionView)
	if len(got) != 0 || !errors.Is(err, errSpiceDBKeyPending) {
		t.Errorf("ActableAccounts = %v, %v, want none and errSpiceDBKeyPending", got, err)
	}
}

// TestKeyGateRunConnects checks that run, polling the file, connects once the
// key appears, and stops with its context.
func TestKeyGateRunConnects(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	g := gateFor(t, &out)
	path := g.cfg.KeyFile
	go g.run(t.Context())
	if err := os.WriteFile(path, []byte(testKey), 0o600); err != nil {
		t.Fatal(err)
	}
	waited, stop := context.WithTimeout(t.Context(), waitDeadline)
	defer stop()
	if !g.wait(waited) {
		t.Fatal("the gate did not connect after the key appeared")
	}
}

// TestKeyGateNeverLogsTheKey checks that the wait is logged once and that no
// line of the log, while waiting or after connecting, shows the key.
func TestKeyGateNeverLogsTheKey(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	g := gateFor(t, &out)
	path := g.cfg.KeyFile
	for range 3 {
		g.step()
	}
	if err := os.WriteFile(path, []byte(testKey), 0o600); err != nil {
		t.Fatal(err)
	}
	if !g.step() {
		t.Fatal("step did not connect")
	}
	logged := out.String()
	if n := strings.Count(logged, "waiting for SpiceDB's preshared key"); n != 1 {
		t.Errorf("the wait is logged %d times, want once:\n%s", n, logged)
	}
	if strings.Contains(logged, testKey) {
		t.Errorf("the log shows the preshared key:\n%s", logged)
	}
}

// TestNotReadyWithoutTheKey checks that /readyz reports not ready, with the
// core otherwise wired, until the key is there, and that the health endpoint
// stays up.
func TestNotReadyWithoutTheKey(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	g := gateFor(t, &out)
	path := g.cfg.KeyFile
	s := &server{}
	api := http.NotFoundHandler()
	s.browser.Store(&api)
	s.ready(&core{access: access{authorization: gatedAuthorization(g)}})

	get := func(p string) (int, string) {
		rec := httptest.NewRecorder()
		s.browserHandler().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, p, http.NoBody))
		return rec.Code, strings.TrimSpace(rec.Body.String())
	}
	if code, body := get("/readyz"); code != http.StatusServiceUnavailable || body != "waiting for SpiceDB's preshared key" {
		t.Errorf("GET /readyz without the key = %d %q, want 503 waiting for the key", code, body)
	}
	if code, _ := get("/healthz"); code != http.StatusOK {
		t.Errorf("GET /healthz without the key = %d, want 200", code)
	}
	if err := os.WriteFile(path, []byte(testKey), 0o600); err != nil {
		t.Fatal(err)
	}
	if !g.step() {
		t.Fatal("step did not connect")
	}
	if ready := s.authzReady.Load(); ready == nil || !(*ready)() {
		t.Error("the authorizer is not ready after the key appeared")
	}
}

// TestReadKeyFile checks which files hold a key: those with a value, trimmed.
func TestReadKeyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tests := []struct {
		name string
		// present is whether the file exists, with content.
		present bool
		content string
		want    secret
	}{
		{name: "absent"},
		{name: "empty", present: true},
		{name: "whitespace", present: true, content: " \n\t "},
		{name: "a key", present: true, content: testKey, want: testKey},
		{name: "a key with a newline", present: true, content: testKey + "\n", want: testKey},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(dir, strings.ReplaceAll(tt.name, " ", "-"))
			if tt.present {
				if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if got := readKeyFile(path); got != tt.want {
				t.Errorf("readKeyFile = %q, want %q", string(got), string(tt.want))
			}
		})
	}
}

// TestServeAuthorization checks which authorization serve gets: the gate,
// only with SpiceDB and no key in the environment, which wins over the file.
func TestServeAuthorization(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		cfg      authzConfig
		wantGate bool
	}{
		{name: "the policy", cfg: authzConfig{Mode: authzPolicy}},
		{name: "the key from the environment", cfg: authzConfig{Mode: authzSpiceDB, Endpoint: "127.0.0.1:1", Key: testKey, Insecure: true, Timeout: time.Second}},
		{name: "no key in the environment", cfg: authzConfig{Mode: authzSpiceDB, Endpoint: "127.0.0.1:1", KeyFile: "/nonexistent/key", Insecure: true, Timeout: time.Second}, wantGate: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, stop := context.WithTimeout(t.Context(), waitDeadline)
			defer stop()
			az, err := serveAuthorization(ctx, &tt.cfg, slog.New(slog.DiscardHandler))
			if err != nil {
				t.Fatalf("serveAuthorization: %v", err)
			}
			if (az.gate != nil) != tt.wantGate {
				t.Errorf("gate = %v, want one: %v", az.gate != nil, tt.wantGate)
			}
			if az.closer != nil {
				if err := az.closer.Close(); err != nil {
					t.Errorf("Close: %v", err)
				}
			}
		})
	}
}

// TestRequireKey checks the commands' key: the environment's, else the
// file's, else an error, since a command can't wait.
func TestRequireKey(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "key")
	cfg := authzConfig{Mode: authzSpiceDB, KeyFile: path}
	if err := cfg.requireKey(); !errors.Is(err, errNoSpiceDBKey) {
		t.Errorf("requireKey without a file = %v, want errNoSpiceDBKey", err)
	}
	if err := os.WriteFile(path, []byte(testKey+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := cfg.requireKey(); err != nil || cfg.Key != testKey {
		t.Errorf("requireKey with the file = %v, key %q, want nil and the file's", err, string(cfg.Key))
	}
	policy := authzConfig{Mode: authzPolicy}
	if err := policy.requireKey(); err != nil {
		t.Errorf("requireKey with the policy = %v, want nil", err)
	}
}

// TestReadKeyFileFollowsTheKubeletsLinks checks that the key is read through
// the links the kubelet makes in a mounted Secret: key to ..data/key, and
// ..data to the directory of the current version.
func TestReadKeyFileFollowsTheKubeletsLinks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "..v1"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "..v1", "key"), []byte(testKey), 0o600); err != nil {
		t.Fatal(err)
	}
	for link, target := range map[string]string{"..data": "..v1", "key": filepath.Join("..data", "key")} {
		if err := os.Symlink(target, filepath.Join(dir, link)); err != nil {
			t.Fatal(err)
		}
	}
	if got := readKeyFile(filepath.Join(dir, "key")); got != testKey {
		t.Errorf("readKeyFile through the kubelet's links = %q, want the key", string(got))
	}
}

// TestStartReconcileWaitsForTheKey checks that, while the key is to come, the
// reconcile writes nothing to SpiceDB, and that it runs once the gate has
// connected.
func TestStartReconcileWaitsForTheKey(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	g := gateFor(t, &out)
	w := &writeLog{}
	az := authorization{syncer: spicedb.NewSyncer(w), gate: g}
	load := func(context.Context) (spicedb.Desired, error) {
		return spicedb.Desired{Accounts: []string{"bonds"}, Sleeves: map[string]string{"bonds": "bonds"}}, nil
	}
	// The reconcile's log line, read through a pipe that closes at the deadline.
	logged, logWriter := io.Pipe()
	ctx, stop := context.WithTimeout(t.Context(), waitDeadline)
	defer stop()
	defer context.AfterFunc(ctx, func() {
		if err := logWriter.Close(); err != nil {
			t.Error(err)
		}
	})()
	az.startReconcile(t.Context(), load, slog.New(slog.NewTextHandler(logWriter, nil)))
	if n := w.count(); n != 0 {
		t.Fatalf("%d writes to SpiceDB before the key was there, want none", n)
	}
	close(g.connected)
	lines := bufio.NewScanner(logged)
	reconciled := false
	for lines.Scan() {
		if strings.Contains(lines.Text(), "authorization reconciled with SpiceDB") {
			reconciled = true
			break
		}
	}
	if !reconciled {
		t.Fatal("the reconcile did not run after the key was there")
	}
	if w.count() == 0 {
		t.Error("the reconcile wrote nothing")
	}
}

// count returns how many relationships have been written.
func (w *writeLog) count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.writes)
}
