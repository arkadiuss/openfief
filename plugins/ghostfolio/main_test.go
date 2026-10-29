package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// Synthetic failures of the tests' fakes.
var (
	errUnreachable      = errors.New("unreachable")
	errSyntheticFailure = errors.New("synthetic failure")
)

// TestManifest checks that manifest.json, which the image carries as its
// signed label (D-42), asks for exactly what the view needs: the view
// capability, the Secret with the Ghostfolio security token, and a config
// schema; no fixed egress host, since the owner's Ghostfolio host is not
// known when the image is built, but the host of the config's url (D-75)
// and an in-cluster service for a Ghostfolio in the same cluster (D-74).
func TestManifest(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	got, err := plugins.ParseManifestLabel(string(data))
	if err != nil {
		t.Fatalf("ParseManifestLabel: %v", err)
	}
	want := plugins.Manifest{
		Capabilities:     []plugins.Capability{plugins.CapabilityView},
		Secrets:          []plugins.SecretMount{{Role: "security-token", MountPath: mountDir}},
		EgressFromConfig: []string{"url"},
		ClusterService:   &plugins.ClusterServiceRequest{Description: "a Ghostfolio running in this cluster"},
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(plugins.Manifest{}, "ConfigSchema")); diff != "" {
		t.Errorf("manifest (-want +got):\n%s", diff)
	}
	if len(got.ConfigSchema) == 0 {
		t.Error("the manifest has no config schema")
	}
}

// TestConfigParse checks the config: an https origin, an interval of at
// least five minutes, an hour by default, and listings with valid ISINs,
// none twice.
func TestConfigParse(t *testing.T) {
	t.Parallel()
	isin, err := ledger.ParseISIN("IETEST000026")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		raw     rawConfig
		want    config
		wantErr bool
	}{
		{name: "the URL alone", raw: rawConfig{URL: "https://ghostfolio.example.com/"}, want: config{
			url: "https://ghostfolio.example.com", interval: time.Hour, listings: map[ledger.ISIN]ghostfolio.Listing{},
		}},
		{name: "an interval and a listing", raw: rawConfig{
			URL: "https://ghostfolio.example.com:3443", Interval: "15m",
			Listings: []rawListing{{ISIN: "IETEST000026", DataSource: "YAHOO", Symbol: "TEST.L"}},
		}, want: config{
			url: "https://ghostfolio.example.com:3443", interval: 15 * time.Minute,
			listings: map[ledger.ISIN]ghostfolio.Listing{isin: {DataSource: "YAHOO", Symbol: "TEST.L"}},
		}},
		{name: "no URL, for an in-cluster Ghostfolio", raw: rawConfig{}, want: config{
			interval: time.Hour, listings: map[ledger.ISIN]ghostfolio.Listing{},
		}},
		{name: "plain http", raw: rawConfig{URL: "http://ghostfolio.example.com"}, wantErr: true},
		{name: "a path", raw: rawConfig{URL: "https://example.com/ghostfolio"}, wantErr: true},
		{name: "a user in the URL", raw: rawConfig{URL: "https://someone@example.com"}, wantErr: true},
		{name: "an interval of a minute", raw: rawConfig{URL: "https://example.com", Interval: "1m"}, wantErr: true},
		{name: "not a duration", raw: rawConfig{URL: "https://example.com", Interval: "hourly"}, wantErr: true},
		{name: "a wrong check digit", raw: rawConfig{URL: "https://example.com", Listings: []rawListing{{ISIN: "IETEST000027", DataSource: "YAHOO", Symbol: "X"}}}, wantErr: true},
		{name: "an ISIN twice", raw: rawConfig{URL: "https://example.com", Listings: []rawListing{
			{ISIN: "IETEST000026", DataSource: "YAHOO", Symbol: "X"}, {ISIN: "IETEST000026", DataSource: "YAHOO", Symbol: "Y"},
		}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.raw.parse()
			if tt.wantErr {
				if !errors.Is(err, errBadConfig) {
					t.Errorf("parse = %v, want errBadConfig", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(config{}), cmp.Comparer(func(a, b ledger.ISIN) bool { return a.Equal(b) })); diff != "" {
				t.Errorf("parse (-want +got):\n%s", diff)
			}
		})
	}
}

// TestReadSecretFile checks that the security token is read from the
// Secret, trimmed, and that a missing or empty file is an error that names
// the file only.
func TestReadSecretFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, keyFile), []byte("synthetic-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "empty"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := readSecretFile(dir, keyFile); err != nil || got != "synthetic-token" {
		t.Errorf("readSecretFile = %q, %v; want the token", got, err)
	}
	for _, name := range []string{"empty", "missing", "../escape"} {
		if _, err := readSecretFile(dir, name); !errors.Is(err, errNoToken) {
			t.Errorf("readSecretFile(%s) = %v, want errNoToken", name, err)
		}
	}
}

// TestNewSyncer checks the wiring: without the core's address, the
// instance's name or the egress proxy the plugin refuses to start.
func TestNewSyncer(t *testing.T) {
	t.Setenv(openfief.EnvCoreAddress, "")
	cfg := config{url: "https://ghostfolio.example.com", interval: time.Hour}
	if _, err := newSyncer(&cfg, func(string) string { return "ghostfolio" }); !errors.Is(err, openfief.ErrNoCoreAddress) {
		t.Errorf("newSyncer without a core address = %v", err)
	}
	t.Setenv(openfief.EnvCoreAddress, "192.0.2.10:8080")
	if _, err := newSyncer(&cfg, func(string) string { return "" }); !errors.Is(err, openfief.ErrNoInstance) {
		t.Errorf("newSyncer without an instance = %v", err)
	}
	t.Setenv(openfief.EnvHTTPSProxy, "")
	if _, err := newSyncer(&cfg, func(string) string { return "ghostfolio" }); err == nil {
		t.Error("newSyncer without the egress proxy succeeded")
	}
	inCluster := func(k string) string {
		if k == plugins.EnvClusterServiceAddress {
			return "192.0.2.30:3333"
		}
		return "ghostfolio"
	}
	if _, err := newSyncer(&cfg, inCluster); !errors.Is(err, errTwoGhostfolios) {
		t.Errorf("newSyncer with a url and an in-cluster service = %v, want errTwoGhostfolios", err)
	}
	// An in-cluster Ghostfolio needs no egress proxy (D-74).
	if _, err := newSyncer(&config{interval: time.Hour}, inCluster); err != nil {
		t.Errorf("newSyncer for an in-cluster Ghostfolio: %v", err)
	}
	onlyInstance := func(k string) string {
		if k == openfief.EnvInstance {
			return "ghostfolio"
		}
		return ""
	}
	if _, err := newSyncer(&config{interval: time.Hour}, onlyInstance); !errors.Is(err, errNoGhostfolio) {
		t.Errorf("newSyncer with neither = %v, want errNoGhostfolio", err)
	}
}

// emptyRecord is a read API with no account.
type emptyRecord struct{ ghostfolio.Record }

func (emptyRecord) ListAccounts(context.Context, *connect.Request[viewsv1.ListAccountsRequest]) (*connect.Response[viewsv1.ListAccountsResponse], error) {
	return connect.NewResponse(&viewsv1.ListAccountsResponse{}), nil
}

// TestNewSyncerWith checks that the syncer logs in with the token from the
// Secret: a missing token fails the sync before anything is sent.
func TestNewSyncerWith(t *testing.T) {
	t.Parallel()
	sent := 0
	hc := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		sent++
		return nil, errUnreachable
	})}
	s := newSyncerWith(&config{url: "https://ghostfolio.example.com"}, "ghostfolio", emptyRecord{}, hc, t.TempDir())
	if _, err := s.Sync(t.Context()); !errors.Is(err, errNoToken) || sent != 0 {
		t.Errorf("Sync without a token = %v after %d requests, want errNoToken and none", err, sent)
	}
}

// roundTripFunc is an http.RoundTripper made of a function.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// countingSyncer reports each sync's number on calls, and fails every
// other sync.
type countingSyncer struct {
	n     int
	calls chan int
}

func (c *countingSyncer) Sync(context.Context) (ghostfolio.Report, error) {
	c.n++
	select {
	case c.calls <- c.n:
	default:
	}
	if c.n%2 == 0 {
		return ghostfolio.Report{}, errSyntheticFailure
	}
	return ghostfolio.Report{Imported: 1}, nil
}

// TestLoop checks that the loop syncs at once, keeps going after a failed
// sync, and stops when its context ends.
func TestLoop(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	s := &countingSyncer{calls: make(chan int, 1)}
	done := make(chan struct{})
	go func() {
		loop(ctx, s, time.Millisecond, slog.New(slog.DiscardHandler))
		close(done)
	}()
	deadline := time.After(10 * time.Second)
	for want := 1; want <= 3; {
		select {
		case n := <-s.calls:
			if n >= want {
				want = n + 1
			}
		case <-deadline:
			t.Fatal("the loop didn't sync three times")
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the loop didn't stop")
	}
}
