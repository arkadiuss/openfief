package openaichat

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestDataCollectionJoinsTheProviderObjectOnOpenRouterOnly checks that a data
// collection setting lands in the same `provider` object as the pinned hosts
// and require_parameters, alone when the request needs nothing else, and
// reaches no other server.
func TestDataCollectionJoinsTheProviderObjectOnOpenRouterOnly(t *testing.T) {
	t.Parallel()
	plain := &Request{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}}
	withTools := &Request{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}, Tools: []Tool{{Name: "t"}}}
	const openRouter, other = "https://openrouter.ai/api/v1", "https://api.openai.com/v1"
	tests := []struct {
		name         string
		client       Client
		req          *Request
		wantProvider any
	}{
		{
			name: "deny alone", client: Client{BaseURL: openRouter, DataCollection: DataCollectionDeny}, req: plain,
			wantProvider: map[string]any{"data_collection": "deny"},
		},
		{
			name: "allow alone", client: Client{BaseURL: openRouter, DataCollection: DataCollectionAllow}, req: plain,
			wantProvider: map[string]any{"data_collection": "allow"},
		},
		{
			name: "deny with tools", client: Client{BaseURL: openRouter, DataCollection: DataCollectionDeny}, req: withTools,
			wantProvider: map[string]any{"require_parameters": true, "data_collection": "deny"},
		},
		{
			name:         "deny with pinned hosts",
			client:       Client{BaseURL: openRouter, Route: []string{"nextbit/bf16"}, DataCollection: DataCollectionDeny},
			req:          withTools,
			wantProvider: map[string]any{"order": []any{"nextbit/bf16"}, "allow_fallbacks": false, "require_parameters": true, "data_collection": "deny"},
		},
		{name: "an unknown value sends nothing", client: Client{BaseURL: openRouter, DataCollection: "maybe"}, req: plain},
		{
			name: "unset keeps the usual body", client: Client{BaseURL: openRouter}, req: withTools,
			wantProvider: map[string]any{"require_parameters": true},
		},
		{name: "another server is told nothing", client: Client{BaseURL: other, DataCollection: DataCollectionDeny}, req: withTools},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body := onTheWire(t, tt.client.RequestBody(tt.req))
			if diff := cmp.Diff(tt.wantProvider, body["provider"]); diff != "" {
				t.Errorf("provider (-want +got):\n%s", diff)
			}
		})
	}
}
