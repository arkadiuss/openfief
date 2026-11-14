package main

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
)

// validConfig is a config with one local model.
func validConfig() rawConfig {
	return rawConfig{
		BaseURL: "https://models.example.org/v1", Location: "owner_hardware",
		Retention: rawRetention{Days: 0},
		Models:    []rawModel{{ID: "synthetic-8b", ToolCalling: true}},
	}
}

func TestParseConfigLocalModel(t *testing.T) {
	t.Parallel()
	raw := validConfig()
	got, err := raw.parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.timeout != defaultTimeout || got.location != modelsv1.Location_LOCATION_OWNER_HARDWARE {
		t.Errorf("timeout %v, location %v", got.timeout, got.location)
	}
	if e, ok := got.entry("synthetic-8b"); !ok || e.GetInputPrice() != "" || !e.GetToolCalling() {
		t.Errorf("entry = %v, %v; want a local model with tools and no price", e, ok)
	}
}

func TestParseConfigCloudModel(t *testing.T) {
	t.Parallel()
	raw := validConfig()
	raw.Location = "third_party_cloud"
	raw.Timeout = "90s"
	raw.Retention = rawRetention{Days: 30, MayTrain: true}
	raw.Models = []rawModel{{ID: "vendor/synthetic-120b", InputPrice: "0.15", OutputPrice: "0.6", PriceCurrency: "USD", ContextTokens: 131072}}
	got, err := raw.parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	e, ok := got.entry("vendor/synthetic-120b")
	if !ok || e.GetInputPrice() != "0.15" || e.GetPriceCurrency() != "USD" {
		t.Errorf("entry = %v, %v", e, ok)
	}
	if got.timeout != 90*time.Second || got.retention.GetDays() != 30 || !got.retention.GetMayTrain() {
		t.Errorf("timeout %v, retention %v", got.timeout, got.retention)
	}
}

func TestParseConfigRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(r *rawConfig)
	}{
		{name: "no location", edit: func(r *rawConfig) { r.Location = "" }},
		{name: "an unknown location", edit: func(r *rawConfig) { r.Location = "moon" }},
		{name: "negative retention", edit: func(r *rawConfig) { r.Retention.Days = -1 }},
		{name: "no models", edit: func(r *rawConfig) { r.Models = nil }},
		{name: "too many models", edit: func(r *rawConfig) {
			r.Models = nil
			for range maxModels + 1 {
				r.Models = append(r.Models, rawModel{ID: "m" + strconv.Itoa(len(r.Models))})
			}
		}},
		{name: "a model twice", edit: func(r *rawConfig) { r.Models = append(r.Models, r.Models[0]) }},
		{name: "a model without an id", edit: func(r *rawConfig) { r.Models = []rawModel{{}} }},
		{name: "an input price without an output price", edit: func(r *rawConfig) {
			r.Models = []rawModel{{ID: "m", InputPrice: "1", PriceCurrency: "USD"}}
		}},
		{name: "a price without a currency", edit: func(r *rawConfig) {
			r.Models = []rawModel{{ID: "m", InputPrice: "1", OutputPrice: "2"}}
		}},
		{name: "a price that is not a decimal", edit: func(r *rawConfig) {
			r.Models = []rawModel{{ID: "m", InputPrice: "1e3", OutputPrice: "2", PriceCurrency: "USD"}}
		}},
		{name: "a negative limit", edit: func(r *rawConfig) { r.Models = []rawModel{{ID: "m", ContextTokens: -1}} }},
		{name: "a timeout that is not a duration", edit: func(r *rawConfig) { r.Timeout = "soon" }},
		{name: "a timeout above the limit", edit: func(r *rawConfig) { r.Timeout = "10m" }},
		{name: "a zero timeout", edit: func(r *rawConfig) { r.Timeout = "0s" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := validConfig()
			tt.edit(&raw)
			if _, err := raw.parse(); !errors.Is(err, errBadConfig) {
				t.Errorf("parse error = %v, want it to wrap errBadConfig", err)
			}
		})
	}
}

// openRouterConfig is a config for OpenRouter with one pinned, reasoning model.
func openRouterConfig() rawConfig {
	raw := validConfig()
	raw.BaseURL = "https://openrouter.ai/api/v1"
	raw.Location = "third_party_cloud"
	raw.Models = []rawModel{
		{ID: "vendor/synthetic-26b", Hosts: []string{"nextbit/bf16"}},
		{ID: "vendor/synthetic-120b", Hosts: []string{"groq", "crusoe"}, Reasoning: "low"},
		{ID: "vendor/unpinned"},
	}
	return raw
}

func TestParseConfigHostsAndReasoningOfAnEntryOnOpenRouter(t *testing.T) {
	t.Parallel()
	raw := openRouterConfig()
	got, err := raw.parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want := map[string]routing{
		"vendor/synthetic-26b":  {hosts: []string{"nextbit/bf16"}},
		"vendor/synthetic-120b": {hosts: []string{"groq", "crusoe"}, reasoning: "low"},
	}
	if diff := cmp.Diff(want, got.routing, cmp.AllowUnexported(routing{})); diff != "" {
		t.Errorf("routing (-want +got):\n%s", diff)
	}
}

func TestParseConfigRefusesHostsAndReasoning(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(r *rawConfig)
	}{
		{name: "hosts on another server", edit: func(r *rawConfig) {
			r.BaseURL = "https://api.openai.com/v1"
		}},
		{name: "reasoning on another server", edit: func(r *rawConfig) {
			r.BaseURL = "https://api.openai.com/v1"
			r.Models = []rawModel{{ID: "m", Reasoning: "low"}}
		}},
		{name: "a host that is not a slug", edit: func(r *rawConfig) { r.Models[0].Hosts = []string{"Next Bit"} }},
		{name: "an empty host", edit: func(r *rawConfig) { r.Models[0].Hosts = []string{""} }},
		{name: "a host listed twice", edit: func(r *rawConfig) { r.Models[0].Hosts = []string{"groq", "groq"} }},
		{name: "too many hosts", edit: func(r *rawConfig) {
			r.Models[0].Hosts = nil
			for i := range maxHosts + 1 {
				r.Models[0].Hosts = append(r.Models[0].Hosts, "host"+strconv.Itoa(i))
			}
		}},
		{name: "a host that is too long", edit: func(r *rawConfig) { r.Models[0].Hosts = []string{strings.Repeat("a", maxHostLen+1)} }},
		{name: "an unknown reasoning setting", edit: func(r *rawConfig) { r.Models[1].Reasoning = "extreme" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := openRouterConfig()
			tt.edit(&raw)
			if _, err := raw.parse(); !errors.Is(err, errBadConfig) {
				t.Errorf("parse error = %v, want it to wrap errBadConfig", err)
			}
		})
	}
}
