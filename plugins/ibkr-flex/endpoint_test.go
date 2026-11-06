package main

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// TestConfigBaseURL checks the baseUrl config field end to end without a
// cluster: the manifest's schema takes an https base URL and refuses plain
// http and the rest the SDK refuses; the install asks the grant for the
// configured host beside the fixed ndcdyn.interactivebrokers.com (D-75); and the connector
// calls the configured base URL, or IBKR's real one without it.
func TestConfigBaseURL(t *testing.T) {
	t.Parallel()
	m := readManifest(t)
	tests := []struct {
		name      string
		config    string
		wantBase  string
		wantHosts []plugins.ConfigHost
		refused   bool
	}{
		{name: "no config: IBKR's Flex Web Service", config: `{}`, wantBase: defaultBaseURL},
		{
			name: "the default written out", config: `{"baseUrl":"https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService"}`, wantBase: defaultBaseURL,
			wantHosts: []plugins.ConfigHost{{Field: "baseUrl", Host: plugins.EgressHost{Host: "ndcdyn.interactivebrokers.com", Port: 443}}},
		},
		{
			name: "a mirror with a port and a path", config: `{"baseUrl":"https://flex.example.org:8443/flex/"}`,
			wantBase:  "https://flex.example.org:8443/flex",
			wantHosts: []plugins.ConfigHost{{Field: "baseUrl", Host: plugins.EgressHost{Host: "flex.example.org", Port: 8443}}},
		},
		{name: "plain http", config: `{"baseUrl":"http://flex.example.org"}`, refused: true},
		{name: "a user", config: `{"baseUrl":"https://owner@flex.example.org"}`, refused: true},
		{name: "a query", config: `{"baseUrl":"https://flex.example.org/?key=x"}`, refused: true},
		{name: "a fragment", config: `{"baseUrl":"https://flex.example.org/#x"}`, refused: true},
		{name: "not a string", config: `{"baseUrl":443}`, refused: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !checkConfig(t, &m, tt.config, tt.refused, tt.wantHosts) {
				return
			}
			var cfg config
			if jsonErr := json.Unmarshal([]byte(tt.config), &cfg); jsonErr != nil {
				t.Fatal(jsonErr)
			}
			base, err := openfief.BaseURL(cfg.BaseURL, defaultBaseURL)
			if err != nil || base != tt.wantBase {
				t.Errorf("BaseURL = %q, %v; want %q", base, err, tt.wantBase)
			}
		})
	}
}

// checkConfig checks config against m's schema, which must refuse it when
// refused, and the hosts it asks the grant for (D-75); it reports whether
// the config was accepted.
func checkConfig(t *testing.T, m *plugins.Manifest, config string, refused bool, wantHosts []plugins.ConfigHost) bool {
	t.Helper()
	err := operator.ValidateConfig(m.ConfigSchema, []byte(config))
	if (err != nil) != refused {
		t.Fatalf("ValidateConfig(%s) = %v, want refused %v", config, err, refused)
	}
	if refused {
		return false
	}
	hosts, err := plugins.ConfigHosts(m, []byte(config))
	if err != nil {
		t.Fatalf("ConfigHosts: %v", err)
	}
	if diff := cmp.Diff(wantHosts, hosts); diff != "" {
		t.Errorf("config hosts the grant approves (-want +got):\n%s", diff)
	}
	return true
}

// readManifest reads manifest.json as the image's signed label carries it.
func readManifest(t *testing.T) plugins.Manifest {
	t.Helper()
	data, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	m, err := plugins.ParseManifestLabel(string(data))
	if err != nil {
		t.Fatalf("ParseManifestLabel: %v", err)
	}
	return m
}

// TestFlexEndpoints checks that the default base URL gives exactly the
// endpoints IBKR documents, and that a mirror's gives them below it.
func TestFlexEndpoints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, base, wantSend, wantGet string
	}{
		{
			name: "IBKR's", base: defaultBaseURL,
			wantSend: ibkrflex.DefaultSendRequestURL, wantGet: ibkrflex.DefaultGetStatementURL,
		},
		{
			name: "a mirror", base: "https://flex.example.org/flex",
			wantSend: "https://flex.example.org/flex/SendRequest", wantGet: "https://flex.example.org/flex/GetStatement",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := flexEndpoints(tt.base)
			if got.sendRequest != tt.wantSend || got.getStatement != tt.wantGet {
				t.Errorf("flexEndpoints(%s) = %+v, want %s and %s", tt.base, got, tt.wantSend, tt.wantGet)
			}
			if flexFetch(http.DefaultClient, tt.base) == nil {
				t.Error("flexFetch returned no fetch")
			}
		})
	}
}
