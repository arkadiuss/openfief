package main

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/valuation/gus"

	"github.com/arkadiuss/openfief/internal/valuation/nbp"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// TestConfigBaseURLs checks the gusBaseUrl, nbpBaseUrl, govPlBaseUrl and
// daneGovPlBaseUrl config fields end to end without a cluster: the
// manifest's schema takes https base URLs and refuses plain http and the
// rest the SDK refuses; the install asks the grant for each configured host
// beside the fixed hosts (D-75); and the connector calls the configured base
// URLs, or the real ones without them.
func TestConfigBaseURLs(t *testing.T) {
	t.Parallel()
	m := readManifest(t)
	tests := []struct {
		name              string
		config            string
		wantGUS, wantNBP  string
		wantGov, wantDane string
		wantHosts         []plugins.ConfigHost
		refused           bool
	}{
		{name: "no config: the real GUS and NBP", config: `{}`, wantGUS: gus.DefaultBaseURL, wantNBP: nbp.DefaultArchiveURL},
		{
			name:    "the ministry's data mirrored",
			config:  `{"govPlBaseUrl":"https://mf.example.org/gov","daneGovPlBaseUrl":"https://dane.example.org"}`,
			wantGUS: gus.DefaultBaseURL, wantNBP: nbp.DefaultArchiveURL,
			wantGov: "https://mf.example.org/gov", wantDane: "https://dane.example.org",
			wantHosts: []plugins.ConfigHost{
				{Field: "govPlBaseUrl", Host: plugins.EgressHost{Host: "mf.example.org", Port: 443}},
				{Field: "daneGovPlBaseUrl", Host: plugins.EgressHost{Host: "dane.example.org", Port: 443}},
			},
		},
		{name: "plain http for gov.pl", config: `{"govPlBaseUrl":"http://mf.example.org"}`, refused: true},
		{name: "a fragment for dane.gov.pl", config: `{"daneGovPlBaseUrl":"https://dane.example.org/#x"}`, refused: true},
		{
			name:    "both mirrored",
			config:  `{"gusBaseUrl":"https://cpi.example.org/gus/","nbpBaseUrl":"https://rates.example.org:8443"}`,
			wantGUS: "https://cpi.example.org/gus", wantNBP: "https://rates.example.org:8443",
			wantHosts: []plugins.ConfigHost{
				{Field: "gusBaseUrl", Host: plugins.EgressHost{Host: "cpi.example.org", Port: 443}},
				{Field: "nbpBaseUrl", Host: plugins.EgressHost{Host: "rates.example.org", Port: 8443}},
			},
		},
		{
			name: "one mirrored", config: `{"nbpBaseUrl":"https://rates.example.org"}`,
			wantGUS: gus.DefaultBaseURL, wantNBP: "https://rates.example.org",
			wantHosts: []plugins.ConfigHost{{Field: "nbpBaseUrl", Host: plugins.EgressHost{Host: "rates.example.org", Port: 443}}},
		},
		{name: "plain http for GUS", config: `{"gusBaseUrl":"http://cpi.example.org"}`, refused: true},
		{name: "plain http for NBP", config: `{"nbpBaseUrl":"http://rates.example.org"}`, refused: true},
		{name: "a user", config: `{"nbpBaseUrl":"https://owner@rates.example.org"}`, refused: true},
		{name: "a query", config: `{"gusBaseUrl":"https://cpi.example.org/?k=x"}`, refused: true},
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
			cpi, rates := indexClients(t, &cfg, tt.wantGUS, tt.wantNBP)
			published := ministryClient(t, &cfg, or(tt.wantGov, treasurybonds.DefaultGovPLBaseURL), or(tt.wantDane, treasurybonds.DefaultDaneBaseURL))
			c := newConnector(cpi, rates, published, mustDate("2026-01-01"), nil)
			if c.indices == nil || c.coupons == nil || c.published == nil {
				t.Errorf("newConnector = %+v, want the index and coupon syncs and the published data", c)
			}
		})
	}
}

// indexClients checks cfg's base URLs of GUS and NBP against wantGUS and
// wantNBP, and returns clients of them.
func indexClients(t *testing.T, cfg *config, wantGUS, wantNBP string) (*gus.Client, *nbp.Archive) {
	t.Helper()
	gusURL, gusErr := openfief.BaseURL(cfg.GUSBaseURL, gus.DefaultBaseURL)
	nbpURL, nbpErr := openfief.BaseURL(cfg.NBPBaseURL, nbp.DefaultArchiveURL)
	if gusErr != nil || nbpErr != nil || gusURL != wantGUS || nbpURL != wantNBP {
		t.Errorf("base URLs = %q, %q (%v, %v); want %q, %q", gusURL, nbpURL, gusErr, nbpErr, wantGUS, wantNBP)
	}
	return gus.New(http.DefaultClient, gusURL), nbp.NewArchive(http.DefaultClient, nbpURL)
}

// ministryClient checks cfg's base URLs of gov.pl and dane.gov.pl against
// wantGov and wantDane, and returns a client of the ministry's data at them.
func ministryClient(t *testing.T, cfg *config, wantGov, wantDane string) *treasurybonds.MinistryClient {
	t.Helper()
	govURL, govErr := openfief.BaseURL(cfg.GovPLBaseURL, treasurybonds.DefaultGovPLBaseURL)
	daneURL, daneErr := openfief.BaseURL(cfg.DaneBaseURL, treasurybonds.DefaultDaneBaseURL)
	if govErr != nil || daneErr != nil || govURL != wantGov || daneURL != wantDane {
		t.Errorf("ministry base URLs = %q, %q (%v, %v); want %q, %q", govURL, daneURL, govErr, daneErr, wantGov, wantDane)
	}
	return treasurybonds.NewMinistryClient(treasurybonds.MinistryConfig{GovPLBaseURL: govURL, DaneBaseURL: daneURL})
}

// or is s, or def when s is empty.
func or(s, def string) string {
	if s == "" {
		return def
	}
	return s
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
