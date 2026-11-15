package main

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/plugins/model-openai/openaichat"
)

// dataCollectionConfig is an OpenRouter config whose first entry denies data
// collection and is pinned, whose second denies it and nothing else, and
// whose third sets nothing.
func dataCollectionConfig() rawConfig {
	raw := validConfig()
	raw.BaseURL = "https://openrouter.ai/api/v1"
	raw.Location = "third_party_cloud"
	raw.Models = []rawModel{
		{ID: "vendor/pinned-deny", Hosts: []string{"nextbit/bf16"}, DataCollection: "deny"},
		{ID: "vendor/deny", DataCollection: "deny"},
		{ID: "vendor/plain"},
	}
	return raw
}

func TestParseConfigDataCollectionOfAnEntryOnOpenRouter(t *testing.T) {
	t.Parallel()
	raw := dataCollectionConfig()
	got, err := raw.parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want := map[string]routing{
		"vendor/pinned-deny": {hosts: []string{"nextbit/bf16"}, dataCollection: "deny"},
		"vendor/deny":        {dataCollection: "deny"},
	}
	if diff := cmp.Diff(want, got.routing, cmp.AllowUnexported(routing{})); diff != "" {
		t.Errorf("routing (-want +got):\n%s", diff)
	}
}

func TestParseConfigRefusesDataCollection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(r *rawConfig)
	}{
		{name: "on another server", edit: func(r *rawConfig) { r.BaseURL = "https://api.openai.com/v1" }},
		{name: "an unknown value", edit: func(r *rawConfig) { r.Models[1].DataCollection = "maybe" }},
		{name: "a value in the wrong case", edit: func(r *rawConfig) { r.Models[1].DataCollection = "Deny" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := dataCollectionConfig()
			tt.edit(&raw)
			if _, err := raw.parse(); !errors.Is(err, errBadConfig) {
				t.Errorf("parse error = %v, want it to wrap errBadConfig", err)
			}
		})
	}
}

// TestAnEntrysDataCollectionReachesTheBodyOfItsOwnRequestsOnly checks the
// `provider` object each entry sends: data_collection merged with the pinned
// hosts and require_parameters, for that entry alone.
func TestAnEntrysDataCollectionReachesTheBodyOfItsOwnRequestsOnly(t *testing.T) {
	t.Parallel()
	raw := dataCollectionConfig()
	cfg, err := raw.parse()
	if err != nil {
		t.Fatal(err)
	}
	withTools := func(model string) *openaichat.Request {
		return &openaichat.Request{Model: model, Messages: []openaichat.Message{{Role: "user", Content: "q"}}, Tools: []openaichat.Tool{{Name: "t"}}}
	}
	plain := func(model string) *openaichat.Request {
		return &openaichat.Request{Model: model, Messages: []openaichat.Message{{Role: "user", Content: "q"}}}
	}
	tests := []struct {
		name         string
		req          *openaichat.Request
		wantProvider any
	}{
		{
			name: "pinned and denied, with tools", req: withTools("vendor/pinned-deny"),
			wantProvider: map[string]any{"order": []any{"nextbit/bf16"}, "allow_fallbacks": false, "require_parameters": true, "data_collection": "deny"},
		},
		{
			name: "pinned and denied, without tools", req: plain("vendor/pinned-deny"),
			wantProvider: map[string]any{"order": []any{"nextbit/bf16"}, "allow_fallbacks": false, "require_parameters": true, "data_collection": "deny"},
		},
		{
			name: "denied alone, with tools", req: withTools("vendor/deny"),
			wantProvider: map[string]any{"require_parameters": true, "data_collection": "deny"},
		},
		{name: "denied alone, without tools", req: plain("vendor/deny"), wantProvider: map[string]any{"data_collection": "deny"}},
		{name: "an entry with none keeps the usual body", req: withTools("vendor/plain"), wantProvider: map[string]any{"require_parameters": true}},
		{name: "an entry with none and no tools sends nothing", req: plain("vendor/plain")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := &openaichat.Client{BaseURL: "https://openrouter.ai/api/v1"}
			body := wireBody(t, cfg.chatFor(base, tt.req.Model).RequestBody(tt.req))
			if diff := cmp.Diff(tt.wantProvider, body["provider"]); diff != "" {
				t.Errorf("provider (-want +got):\n%s", diff)
			}
			if base.DataCollection != "" {
				t.Errorf("the shared client was changed: %+v", base)
			}
		})
	}
}
