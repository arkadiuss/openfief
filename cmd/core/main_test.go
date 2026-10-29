package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// testDSN is a placeholder connection string; nothing connects to it.
const testDSN = "postgres://placeholder@localhost/openfief"

// warsaw returns the default time zone.
func warsaw(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}
	return loc
}

// defaultServe returns serve's configuration with every flag at its
// default.
func defaultServe(t *testing.T) serveConfig {
	t.Helper()
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	return serveConfig{
		Listen: ":8080", Reporting: pln,
		EndOfDay: valuationpg.DailyAt{Hour: 23, Minute: 30, Location: warsaw(t)},
		Workers:  2, BondBasis: valuation.NetIfRedeemed, EnsureDefaults: true, ShutdownTimeout: 25 * time.Second,
		Sync: syncConfig{
			Policy:         ingest.SyncPolicy{Interval: 6 * time.Hour, Jitter: 30 * time.Minute, RetryFirst: 5 * time.Minute, RetryMax: time.Hour},
			Timeout:        5 * time.Minute,
			ServiceAccount: "openfief-core",
		},
		Auth: authConfig{Origins: mustOrigins(t, "http://localhost:8080"), Sessions: identity.DefaultSessionPolicy()},
	}
}

// mustOrigins returns the allow-list of texts.
func mustOrigins(t *testing.T, texts ...string) identity.Origins {
	t.Helper()
	o, err := identity.NewOrigins(texts)
	if err != nil {
		t.Fatalf("NewOrigins: %v", err)
	}
	return o
}

// mustOrigin parses text.
func mustOrigin(t *testing.T, text string) identity.Origin {
	t.Helper()
	o, err := identity.ParseOrigin(text)
	if err != nil {
		t.Fatalf("ParseOrigin: %v", err)
	}
	return o
}

func TestParseArgs(t *testing.T) {
	t.Parallel()
	def := defaultServe(t)
	utc := def
	utc.EndOfDay = valuationpg.DailyAt{Hour: 22, Minute: 5, Location: time.UTC}
	utc.Listen, utc.Workers, utc.BondBasis, utc.EnsureDefaults, utc.ShutdownTimeout = "127.0.0.1:9090", 4, valuation.GrossAccrued, false, time.Second
	pod := def
	pod.Connectors = true
	tuned := pod
	tuned.Connectors = false
	tuned.Sync = syncConfig{
		Policy:  ingest.SyncPolicy{Interval: time.Hour, Jitter: 0, RetryFirst: time.Minute, RetryMax: 2 * time.Minute},
		Timeout: time.Minute, ServiceAccount: "core-sa",
	}
	authed := def
	authed.Auth = authConfig{
		Origins:      mustOrigins(t, "https://dev.example.com", "http://localhost:9000"),
		Sessions:     identity.SessionPolicy{Idle: time.Hour, Max: 48 * time.Hour},
		DemoSessions: true,
	}
	tests := []struct {
		name    string
		args    []string
		dsn     string
		pod     bool // KUBERNETES_SERVICE_HOST is set
		want    command
		wantErr error // nil with wantFail: any error
		fail    bool
	}{
		{name: "no arguments serve", args: nil, dsn: testDSN, want: command{Name: cmdServe, DSN: testDSN, Serve: def}},
		{name: "serve", args: []string{"serve"}, dsn: testDSN, want: command{Name: cmdServe, DSN: testDSN, Serve: def}},
		{
			name: "serve flags",
			args: []string{
				"--listen=127.0.0.1:9090", "--time-zone=UTC", "--end-of-day=22:05", "--workers=4",
				"--bond-basis=gross_accrued", "--ensure-defaults=false", "--shutdown-timeout=1s",
			},
			dsn: testDSN, want: command{Name: cmdServe, DSN: testDSN, Serve: utc},
		},
		{name: "in a pod, connectors are on", dsn: testDSN, pod: true, want: command{Name: cmdServe, DSN: testDSN, Serve: pod}},
		{
			name: "connector flags", pod: true, dsn: testDSN,
			args: []string{
				"--connectors=false", "--sync-every=1h", "--sync-jitter=0s",
				"--sync-retry-first=1m", "--sync-retry-max=2m", "--sync-timeout=1m", "--service-account=core-sa",
			},
			want: command{Name: cmdServe, DSN: testDSN, Serve: tuned},
		},
		{
			name: "auth flags", dsn: testDSN,
			args: []string{
				"--auth-origin=https://dev.example.com", "--auth-origin", "http://localhost:9000",
				"--session-idle=1h", "--session-max=48h", "--demo-sessions",
			},
			want: command{Name: cmdServe, DSN: testDSN, Serve: authed},
		},
		{name: "plain http off localhost", args: []string{"--auth-origin=http://dev.example.com"}, dsn: testDSN, wantErr: identity.ErrInvalidOrigin},
		{name: "idle above max", args: []string{"--session-idle=10h", "--session-max=1h"}, dsn: testDSN, wantErr: identity.ErrInvalidSessionPolicy},
		{
			name: "enroll-link", args: []string{"enroll-link"}, dsn: testDSN,
			want: command{Name: cmdEnrollLink, DSN: testDSN, Enroll: enrollArgs{Origin: mustOrigin(t, "http://localhost:8080"), TTL: 15 * time.Minute}},
		},
		{
			name: "enroll-link for a host", args: []string{"enroll-link", "--host=dev.example.com", "--ttl=1h"}, dsn: testDSN,
			want: command{Name: cmdEnrollLink, DSN: testDSN, Enroll: enrollArgs{Origin: mustOrigin(t, "https://dev.example.com"), TTL: time.Hour}},
		},
		{
			name: "enroll-link for an origin", args: []string{"enroll-link", "--host=https://dev.example.com:8443"}, dsn: testDSN,
			want: command{Name: cmdEnrollLink, DSN: testDSN, Enroll: enrollArgs{Origin: mustOrigin(t, "https://dev.example.com:8443"), TTL: 15 * time.Minute}},
		},
		{name: "enroll-link for an IP", args: []string{"enroll-link", "--host=192.0.2.1"}, dsn: testDSN, wantErr: identity.ErrInvalidOrigin},
		{name: "enroll-link with an argument", args: []string{"enroll-link", "now"}, dsn: testDSN, wantErr: errUsage},
		{name: "retry max below first", args: []string{"--sync-retry-first=1h", "--sync-retry-max=1m"}, dsn: testDSN, wantErr: ingest.ErrInvalidSyncPolicy},
		{name: "zero sync interval", args: []string{"--sync-every=0s"}, dsn: testDSN, wantErr: ingest.ErrInvalidSyncPolicy},
		{
			name: "connector bind", args: []string{"connector", "bind", "--instance", "xtb-main", "--account", "acc-1"}, dsn: testDSN,
			want: command{Name: cmdBind, DSN: testDSN, Connector: connectorArgs{Instance: "xtb-main", Account: "acc-1"}},
		},
		{
			name: "connector sync", args: []string{"connector", "sync", "--instance=xtb-main"}, dsn: testDSN,
			want: command{Name: cmdSync, DSN: testDSN, Connector: connectorArgs{Instance: "xtb-main"}},
		},
		{
			name: "connector sync in a mode", args: []string{"connector", "sync", "--instance=xtb-main", "--mode=transactions"}, dsn: testDSN,
			want: command{Name: cmdSync, DSN: testDSN, Connector: connectorArgs{Instance: "xtb-main", Mode: ingest.ModeTransactions}},
		},
		{name: "connector without subcommand", args: []string{"connector"}, dsn: testDSN, wantErr: errUsage},
		{name: "connector sync with no instance flag", args: []string{"connector", "sync", "--mode=opening"}, dsn: testDSN, wantErr: errNoInstance},
		{name: "connector bind without an account", args: []string{"connector", "bind", "--instance=x"}, dsn: testDSN, wantErr: errNoAccount},
		{name: "connector bind with a mode", args: []string{"connector", "bind", "--instance=x", "--account=a", "--mode=opening"}, dsn: testDSN, wantErr: errUsage},
		{name: "connector sync with an account", args: []string{"connector", "sync", "--instance=x", "--account=a"}, dsn: testDSN, wantErr: errUsage},
		{name: "connector sync in an unknown mode", args: []string{"connector", "sync", "--instance=x", "--mode=full"}, dsn: testDSN, wantErr: errBadMode},
		{name: "connector sync with an argument", args: []string{"connector", "sync", "--instance=x", "now"}, dsn: testDSN, wantErr: errUsage},
		{name: "migrate", args: []string{"migrate"}, dsn: testDSN, want: command{Name: cmdMigrate, DSN: testDSN}},
		{name: "token issue", args: []string{"token", "issue", "--name", "demo", "--ttl", "168h"}, dsn: testDSN, want: command{Name: cmdTokenIssue, DSN: testDSN, TokenName: "demo", TTL: 168 * time.Hour}},
		{name: "token issue default lifetime", args: []string{"token", "issue", "--name=x"}, dsn: testDSN, want: command{Name: cmdTokenIssue, DSN: testDSN, TokenName: "x", TTL: 720 * time.Hour}},
		{name: "token list", args: []string{"token", "list"}, dsn: testDSN, want: command{Name: cmdTokenList, DSN: testDSN}},
		{name: "token revoke", args: []string{"token", "revoke", "0123456789abcdef"}, dsn: testDSN, want: command{Name: cmdTokenRevoke, DSN: testDSN, TokenID: "0123456789abcdef"}},
		{name: "no DSN", args: []string{"migrate"}, wantErr: errNoDSN},
		{name: "unknown command", args: []string{"frobnicate"}, dsn: testDSN, wantErr: errUsage},
		{name: "token without subcommand", args: []string{"token"}, dsn: testDSN, wantErr: errUsage},
		{name: "revoke without ID", args: []string{"token", "revoke"}, dsn: testDSN, wantErr: errUsage},
		{name: "revoke with two IDs", args: []string{"token", "revoke", "a", "b"}, dsn: testDSN, wantErr: errUsage},
		{name: "migrate with an argument", args: []string{"migrate", "now"}, dsn: testDSN, wantErr: errUsage},
		{name: "list with an argument", args: []string{"token", "list", "all"}, dsn: testDSN, wantErr: errUsage},
		{name: "issue with an argument", args: []string{"token", "issue", "--name=x", "extra"}, dsn: testDSN, wantErr: errUsage},
		{name: "serve with an argument", args: []string{"serve", "extra"}, dsn: testDSN, wantErr: errUsage},
		{name: "unknown flag", args: []string{"--verbose"}, dsn: testDSN, fail: true},
		{name: "bad lifetime", args: []string{"token", "issue", "--ttl=soon"}, dsn: testDSN, fail: true},
		{name: "bad currency", args: []string{"--reporting-currency=zloty"}, dsn: testDSN, fail: true},
		{name: "market basis for bonds", args: []string{"--bond-basis=market"}, dsn: testDSN, wantErr: valuation.ErrBasisNotForBonds},
		{name: "unknown basis", args: []string{"--bond-basis=par"}, dsn: testDSN, wantErr: valuation.ErrUnknownBasis},
		{name: "unknown time zone", args: []string{"--time-zone=Mars/Olympus"}, dsn: testDSN, fail: true},
		{name: "end of day hour", args: []string{"--end-of-day=24:00"}, dsn: testDSN, wantErr: errBadEndOfDay},
		{name: "end of day minute", args: []string{"--end-of-day=23:60"}, dsn: testDSN, wantErr: errBadEndOfDay},
		{name: "end of day short", args: []string{"--end-of-day=9:30"}, dsn: testDSN, wantErr: errBadEndOfDay},
		{name: "end of day no colon", args: []string{"--end-of-day=2330"}, dsn: testDSN, wantErr: errBadEndOfDay},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseArgs(tt.args, func(k string) string {
				switch {
				case k == dsnEnv:
					return tt.dsn
				case k == inClusterEnv && tt.pod:
					return "192.0.2.1"
				}
				return ""
			})
			wantFail := tt.fail || tt.wantErr != nil
			if (err != nil) != wantFail || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Fatalf("parseArgs error = %v, want %v (fail: %v)", err, tt.wantErr, wantFail)
			}
			if diff := cmp.Diff(tt.want, got, cmp.Comparer(sameLocation)); diff != "" {
				t.Errorf("parseArgs (-want +got):\n%s", diff)
			}
		})
	}
}

// sameLocation compares time zones by name.
func sameLocation(a, b *time.Location) bool { return a.String() == b.String() }

func TestListTokens(t *testing.T) {
	t.Parallel()
	created := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)
	toks := []identity.ScriptToken{
		{ID: "0123456789abcdef", Name: "active", Scope: identity.ScopeRead, CreatedAt: created, ExpiresAt: created.Add(720 * time.Hour)},
		{ID: "1123456789abcdef", Name: "old", Scope: identity.ScopeRead, CreatedAt: created, ExpiresAt: created.Add(time.Hour)},
		{ID: "2123456789abcdef", Name: "gone", Scope: identity.ScopeRead, CreatedAt: created, ExpiresAt: created.Add(720 * time.Hour), RevokedAt: created.Add(time.Hour)},
	}
	var out bytes.Buffer
	if err := listTokens(&out, toks, created.Add(24*time.Hour)); err != nil {
		t.Fatalf("listTokens: %v", err)
	}
	want := `ID                NAME    SCOPE  CREATED               EXPIRES               STATUS
0123456789abcdef  active  read   2026-10-01T08:00:00Z  2026-10-31T08:00:00Z  active
1123456789abcdef  old     read   2026-10-01T08:00:00Z  2026-10-01T09:00:00Z  expired
2123456789abcdef  gone    read   2026-10-01T08:00:00Z  2026-10-31T08:00:00Z  revoked
`
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("listTokens (-want +got):\n%s", diff)
	}
}

func TestHealthBeforeReady(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		draining bool
		path     string
		want     int
		wantBody string
	}{
		{name: "alive while starting", path: "/healthz", want: http.StatusOK, wantBody: "ok"},
		{name: "not ready while starting", path: "/readyz", want: http.StatusServiceUnavailable, wantBody: "starting"},
		{name: "no API while starting", path: "/v1/accounts", want: http.StatusServiceUnavailable, wantBody: "starting"},
		{name: "alive while draining", draining: true, path: "/healthz", want: http.StatusOK, wantBody: "ok"},
		{name: "not ready while draining", draining: true, path: "/readyz", want: http.StatusServiceUnavailable, wantBody: "shutting down"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &server{}
			s.draining.Store(tt.draining)
			rec := httptest.NewRecorder()
			s.handler().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, http.NoBody))
			if rec.Code != tt.want || strings.TrimSpace(rec.Body.String()) != tt.wantBody {
				t.Errorf("GET %s = %d %q, want %d %q", tt.path, rec.Code, rec.Body.String(), tt.want, tt.wantBody)
			}
			if got := rec.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
		})
	}
}

func TestRunRefusesUnknownCommand(t *testing.T) {
	t.Parallel()
	err := run(context.Background(), &command{Name: "nope"}, &bytes.Buffer{}, nil)
	if diff := cmp.Diff(errUsage, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("run (-want +got):\n%s", diff)
	}
}
