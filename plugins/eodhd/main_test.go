package main

import (
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
)

// TestManifest checks that manifest.json, which the image carries as its
// signed label (D-42), is a valid manifest asking for exactly what the
// plugin needs: the connector capability, eodhd.com and the Secret with the
// owner's API key (D-22), nothing else.
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
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Egress:       []plugins.EgressHost{{Host: "eodhd.com", Port: 443}},
		Secrets:      []plugins.SecretMount{{Name: "eodhd-credentials", MountPath: keyMount}},
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(plugins.Manifest{}, "ConfigSchema")); diff != "" {
		t.Errorf("manifest (-want +got):\n%s", diff)
	}
	if len(got.ConfigSchema) == 0 {
		t.Error("the manifest has no config schema")
	}
}

func TestConfigParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		raw     rawConfig
		wantErr bool
	}{
		{name: "empty: nothing to fetch, whole history"},
		{name: "instruments and indices", raw: rawConfig{
			Start:       "2020-01-02",
			Instruments: []rawInstrument{{ISIN: syntISIN, Exchange: "XETRA", Symbol: "SYNT"}, {ISIN: fakewISIN}},
			Indices:     []rawIndex{{Name: "FAKE20", Ticker: "FAKE20.INDX"}},
		}},
		{name: "start not a date", raw: rawConfig{Start: "02.01.2020"}, wantErr: true},
		{name: "bad check digit", raw: rawConfig{Instruments: []rawInstrument{{ISIN: "XX00000SYNT4"}}}, wantErr: true},
		{name: "unsupported exchange", raw: rawConfig{Instruments: []rawInstrument{{ISIN: syntISIN, Exchange: "F"}}}, wantErr: true},
		{name: "an ISIN twice", raw: rawConfig{Instruments: []rawInstrument{{ISIN: syntISIN, Exchange: "LSE"}, {ISIN: syntISIN, Exchange: "XETRA"}}}, wantErr: true},
		{name: "an index without a ticker", raw: rawConfig{Indices: []rawIndex{{Name: "FAKE20"}}}, wantErr: true},
		{name: "an index name twice", raw: rawConfig{Indices: []rawIndex{{Name: "X", Ticker: "A.INDX"}, {Name: "X", Ticker: "B.INDX"}}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.raw.parse()
			if errors.Is(err, errBadConfig) != tt.wantErr {
				t.Fatalf("parse error = %v, want an invalid config error %v", err, tt.wantErr)
			}
			complete := len(got.instruments) == len(tt.raw.Instruments) && len(got.indices) == len(tt.raw.Indices) &&
				(tt.raw.Start == "") == got.start.IsZero()
			if err == nil && !complete {
				t.Errorf("parse = %+v, want every instrument and index, and the start", got)
			}
		})
	}
}

func TestNewConnector(t *testing.T) {
	t.Parallel()
	c := newConnector(http.DefaultClient, config{})
	if _, ok := c.source(fakeKey).(*eodhd.Client); !ok || c.secrets != keyMount || c.today == nil {
		t.Errorf("newConnector = %+v, want EODHD's client, the mounted Secret and the clock", c)
	}
}

func TestToday(t *testing.T) {
	t.Parallel()
	if today().IsZero() {
		t.Error("today() is the zero date")
	}
}
