package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// TestManifest checks that manifest.json, which the image carries as its
// signed label (D-42), is a valid manifest asking for exactly what the
// plugin needs: the connector capability, api.nbp.pl, the host of its
// baseUrl config field (D-75) and the optional extra CA bundle, nothing
// else.
func TestManifest(t *testing.T) {
	t.Parallel()
	got := readManifest(t)
	want := plugins.Manifest{
		Capabilities:     []plugins.Capability{plugins.CapabilityConnector},
		Egress:           []plugins.EgressHost{{Host: "api.nbp.pl", Port: 443}},
		EgressFromConfig: []string{"baseUrl"},
		Secrets:          []plugins.SecretMount{{Role: openfief.CABundleRole, MountPath: openfief.CABundleDir, Optional: true}},
		// It books for the institution every delivery names (D-182).
		Institutions: []string{institution},
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(plugins.Manifest{}, "ConfigSchema")); diff != "" {
		t.Errorf("manifest (-want +got):\n%s", diff)
	}
	if len(got.ConfigSchema) == 0 {
		t.Error("the manifest has no config schema")
	}
}

func TestStartDate(t *testing.T) {
	t.Parallel()
	today := day(t, "2025-04-30")
	tests := []struct {
		name    string
		cfg     config
		want    string
		wantErr bool
	}{
		{name: "configured", cfg: config{Start: "2024-12-31"}, want: "2024-12-31"},
		{name: "default: the 93 days up to today", want: "2025-01-28"},
		{name: "not a date", cfg: config{Start: "31.12.2024"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.cfg.startDate(today)
			if (err != nil) != tt.wantErr {
				t.Fatalf("startDate error = %v, want error %v", err, tt.wantErr)
			}
			if err == nil && got.String() != tt.want {
				t.Errorf("startDate = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestToday(t *testing.T) {
	t.Parallel()
	if today().IsZero() {
		t.Error("today() is the zero date")
	}
}
