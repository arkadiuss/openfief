package main

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// TestManifest checks that manifest.json, which the image carries as its
// signed label (D-42), is a valid manifest asking for exactly what the
// plugin needs: the connector capability and egress to GUS's API and NBP's
// archive, and no secret, since both are public.
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
		Egress:       []plugins.EgressHost{{Host: "api-sdp.stat.gov.pl", Port: 443}, {Host: "static.nbp.pl", Port: 443}},
	}
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

func TestStartDate(t *testing.T) {
	t.Parallel()
	today := mustDate("2026-10-29")
	tests := []struct {
		name    string
		start   string
		want    string
		wantErr bool
	}{
		{name: "default: January 13 years back", want: "2013-01-01"},
		{name: "configured", start: "2018-06-01", want: "2018-06-01"},
		{name: "not a date", start: "2018-06", wantErr: true},
		{name: "before the holiday calendar", start: "1991-12-01", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := config{Account: "acc-bonds", Start: tt.start}.startDate(today)
			if (err != nil) != tt.wantErr {
				t.Fatalf("startDate = %v, want an error: %t", err, tt.wantErr)
			}
			if tt.start == "1991-12-01" && !errors.Is(err, errStartTooEarly) {
				t.Errorf("startDate = %v, want errStartTooEarly", err)
			}
			if err == nil && got.String() != tt.want {
				t.Errorf("startDate = %s, want %s", got, tt.want)
			}
		})
	}
}
