package operator_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
)

const modelOpenAIManifest = "../../../plugins/model-openai/manifest.json"

// TestModelProviderManifest checks that the first-party OpenAI-compatible
// model-provider plugin's manifest (D-205) is one the operator accepts: a
// valid manifest whose only capability is model-provider, with no fixed
// egress host. TestModelProviderConfig checks its schema.
func TestModelProviderManifest(t *testing.T) {
	t.Parallel()
	m := readModelOpenAIManifest(t)
	if diff := cmp.Diff([]plugins.Capability{plugins.CapabilityModelProvider}, m.Capabilities); diff != "" {
		t.Errorf("capabilities (-want +got):\n%s", diff)
	}
	if len(m.Egress) != 0 {
		t.Errorf("fixed egress hosts = %v, want none: the host comes from the config", m.Egress)
	}
}

// readModelOpenAIManifest parses the plugin's manifest as the operator does.
func readModelOpenAIManifest(t *testing.T) plugins.Manifest {
	t.Helper()
	data, err := os.ReadFile(modelOpenAIManifest)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	m, err := plugins.ParseManifestLabel(string(data))
	if err != nil {
		t.Fatalf("ParseManifestLabel: %v", err)
	}
	return m
}

// TestModelProviderConfig checks the manifest's schema against provider
// instances' configs and the hosts the grant approves for them (D-75).
func TestModelProviderConfig(t *testing.T) {
	t.Parallel()
	m := readModelOpenAIManifest(t)
	tests := []struct {
		name      string
		config    string
		refused   bool
		wantHosts []plugins.ConfigHost
	}{
		{
			name:      "a cloud provider with a priced model",
			config:    `{"baseUrl":"https://models.example.org/v1","location":"third_party_cloud","retention":{"days":30,"mayTrain":false},"models":[{"id":"vendor/synthetic","inputPrice":"0.15","outputPrice":"0.6","priceCurrency":"USD","toolCalling":true}]}`,
			wantHosts: []plugins.ConfigHost{{Field: "baseUrl", Host: plugins.EgressHost{Host: "models.example.org", Port: 443}}},
		},
		{
			name:      "a provider on a port",
			config:    `{"baseUrl":"https://llm.example.org:8443/v1","timeout":"90s","location":"owner_hardware","retention":{"days":0,"mayTrain":false},"models":[{"id":"synthetic-8b"}]}`,
			wantHosts: []plugins.ConfigHost{{Field: "baseUrl", Host: plugins.EgressHost{Host: "llm.example.org", Port: 8443}}},
		},
		{
			name: "plain http", refused: true,
			config: `{"baseUrl":"http://models.example.org","location":"owner_hardware","retention":{"days":0,"mayTrain":false},"models":[{"id":"m"}]}`,
		},
		{
			name: "no base URL", refused: true,
			config: `{"location":"owner_hardware","retention":{"days":0,"mayTrain":false},"models":[{"id":"m"}]}`,
		},
		{
			name: "no location", refused: true,
			config: `{"baseUrl":"https://models.example.org","retention":{"days":0,"mayTrain":false},"models":[{"id":"m"}]}`,
		},
		{
			name: "no retention terms", refused: true,
			config: `{"baseUrl":"https://models.example.org","location":"third_party_cloud","models":[{"id":"m"}]}`,
		},
		{
			name: "no models", refused: true,
			config: `{"baseUrl":"https://models.example.org","location":"owner_hardware","retention":{"days":0,"mayTrain":false},"models":[]}`,
		},
		{
			name: "a key in the config", refused: true,
			config: `{"baseUrl":"https://models.example.org","apiKey":"synthetic","location":"owner_hardware","retention":{"days":0,"mayTrain":false},"models":[{"id":"m"}]}`,
		},
		{
			name: "a price as a number", refused: true,
			config: `{"baseUrl":"https://models.example.org","location":"owner_hardware","retention":{"days":0,"mayTrain":false},"models":[{"id":"m","inputPrice":1.5}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := operator.ValidateConfig(m.ConfigSchema, []byte(tt.config))
			if (err != nil) != tt.refused {
				t.Fatalf("ValidateConfig = %v, want refused %v", err, tt.refused)
			}
			if tt.refused {
				return
			}
			hosts, err := plugins.ConfigHosts(&m, []byte(tt.config))
			if err != nil {
				t.Fatalf("ConfigHosts: %v", err)
			}
			if diff := cmp.Diff(tt.wantHosts, hosts); diff != "" {
				t.Errorf("config hosts the grant approves (-want +got):\n%s", diff)
			}
		})
	}
}
