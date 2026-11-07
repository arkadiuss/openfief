package main

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// TestManifest checks that manifest.json, which the image carries as its
// signed label (D-42), is a valid manifest asking for exactly what the
// plugin needs: the connector capability and egress to GUS's API and NBP's
// archive, or the hosts of its base URL config fields (D-75), no credential,
// since both are public, only the optional extra CA bundle, and the bond
// series terms it delivers (D-179).
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
		Capabilities:     []plugins.Capability{plugins.CapabilityConnector},
		Egress:           []plugins.EgressHost{{Host: "api-sdp.stat.gov.pl", Port: 443}, {Host: "static.nbp.pl", Port: 443}},
		EgressFromConfig: []string{"gusBaseUrl", "nbpBaseUrl"},
		ClusterService:   &plugins.ClusterServiceRequest{Description: "an in-cluster HTTPS service to call instead of GUS and NBP, at the gusBaseUrl and nbpBaseUrl paths, its certificate issued for their hosts"},
		Secrets:          []plugins.SecretMount{{Role: openfief.CABundleRole, MountPath: openfief.CABundleDir, Optional: true}},
		// It books for the institution every delivery names (D-182).
		Institutions: []string{"treasury-bonds"},
		BondTerms:    true,
		// It delivers CPI and NBP reference rate readings (D-186).
		IndexReadings: true,
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(plugins.Manifest{}, "ConfigSchema")); diff != "" {
		t.Errorf("manifest (-want +got):\n%s", diff)
	}
	if len(got.ConfigSchema) == 0 {
		t.Error("the manifest has no config schema")
	}
	// D-118: one instance serves every bond account, so its config names none.
	if strings.Contains(string(got.ConfigSchema), `"account"`) {
		t.Errorf("the config schema names an account: %s", got.ConfigSchema)
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
			got, err := config{Start: tt.start}.startDate(today)
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
