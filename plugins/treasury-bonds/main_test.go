package main

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// TestManifest checks that manifest.json, which the image carries as its
// signed label (D-42), is a valid manifest asking for exactly what the
// plugin needs: the connector capability, and no egress host or secret.
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
	want := plugins.Manifest{Capabilities: []plugins.Capability{plugins.CapabilityConnector}}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(plugins.Manifest{}, "ConfigSchema")); diff != "" {
		t.Errorf("manifest (-want +got):\n%s", diff)
	}
	if len(got.ConfigSchema) == 0 {
		t.Error("the manifest has no config schema")
	}
}

func TestDateIn(t *testing.T) {
	t.Parallel()
	loc, err := time.LoadLocation(warsaw)
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}
	// 23:30 UTC on 28 October is already 29 October in Warsaw (UTC+1).
	instant := time.Date(2026, time.October, 28, 23, 30, 0, 0, time.UTC)
	if got := dateIn(instant, loc).String(); got != "2026-10-29" {
		t.Errorf("dateIn = %s, want 2026-10-29", got)
	}
}
