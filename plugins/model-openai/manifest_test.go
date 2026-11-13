package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// manifestFile is the part of manifest.json the test reads.
type manifestFile struct {
	Capabilities     []string `json:"capabilities"`
	EgressFromConfig []string `json:"egressFromConfig"`
	Secrets          []struct {
		Role      string `json:"role"`
		MountPath string `json:"mountPath"`
		Optional  bool   `json:"optional"`
	} `json:"secrets"`
}

// TestManifest checks that manifest.json, which the image carries as its
// signed label (D-42), asks for exactly what the plugin needs: the
// model-provider capability, a host from its baseUrl config field (D-75) and
// none fixed, the optional API key and CA bundle Secrets, and no data tier.
// internal/plugins checks that it is a valid manifest the operator accepts.
func TestManifest(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var m manifestFile
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if !reflect.DeepEqual(m.Capabilities, []string{"model-provider"}) {
		t.Errorf("capabilities = %v, want model-provider only", m.Capabilities)
	}
	if !reflect.DeepEqual(m.EgressFromConfig, []string{"baseUrl"}) {
		t.Errorf("egressFromConfig = %v, want baseUrl only", m.EgressFromConfig)
	}
	wantSecrets := map[string]string{"api-key": keyMount, "ca-bundle": "/etc/openfief/ca-bundle"}
	if len(m.Secrets) != len(wantSecrets) {
		t.Fatalf("secrets = %v, want %v", m.Secrets, wantSecrets)
	}
	for _, s := range m.Secrets {
		if wantSecrets[s.Role] != s.MountPath || !s.Optional {
			t.Errorf("secret %+v, want an optional one at %q", s, wantSecrets[s.Role])
		}
	}
}

// TestManifestAsksForNothingFixed checks that the manifest names no fixed
// egress host, no data tier and nothing a connector declares: a model
// provider reaches only the host of its config and delivers nothing.
func TestManifestAsksForNothingFixed(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("manifest is not JSON: %v", err)
	}
	for _, key := range []string{"egress", "dataTiers", "institutions", "marketData"} {
		if _, ok := m[key]; ok {
			t.Errorf("manifest has %q", key)
		}
	}
}
