package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/plugins/model-openai/openaichat"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
)

// newProvider is the provider over a fake with the config's two models: a
// plain one and one with tools, structured output and limits.
func newProvider(t *testing.T, reply string) (*fakeProvider, *provider) {
	t.Helper()
	f, c := newFake(t, reply)
	raw := rawConfig{
		BaseURL: "https://models.example.org/v1", Location: "third_party_cloud", Retention: rawRetention{Days: 30},
		Models: []rawModel{
			{ID: "plain"},
			{ID: "capable", ToolCalling: true, StructuredOutput: true, MaxOutputTokens: 100, InputPrice: "1.5", OutputPrice: "3", PriceCurrency: "USD"},
		},
	}
	cfg, err := raw.parse()
	if err != nil {
		t.Fatal(err)
	}
	return f, &provider{cfg: cfg, chat: c}
}

func userMessage(text string) *modelsv1.Message {
	return &modelsv1.Message{Role: modelsv1.MessageRole_MESSAGE_ROLE_USER, Content: text}
}

func TestDescribeAndListModels(t *testing.T) {
	t.Parallel()
	_, p := newProvider(t, textReply)
	d, err := p.Describe(t.Context(), connect.NewRequest(&modelsv1.DescribeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if d.Msg.GetProtocol() != modelsv1.WireProtocol_WIRE_PROTOCOL_OPENAI_COMPATIBLE ||
		d.Msg.GetLocation() != modelsv1.Location_LOCATION_THIRD_PARTY_CLOUD || d.Msg.GetRetention().GetDays() != 30 {
		t.Errorf("Describe = %v", d.Msg)
	}
	l, err := p.ListModels(t.Context(), connect.NewRequest(&modelsv1.ListModelsRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if got := l.Msg.GetModels(); len(got) != 2 || got[1].GetId() != "capable" || got[1].GetOutputPrice() != "3" || !got[1].GetToolCalling() {
		t.Errorf("ListModels = %v", got)
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()
	f, p := newProvider(t, `{"choices":[{"message":{"content":null,"tool_calls":[
		{"id":"call-1","type":"function","function":{"name":"lookup","arguments":"{\"q\":1}"}}]},"finish_reason":"tool_calls"}],
		"usage":{"prompt_tokens":5,"completion_tokens":2}}`)
	resp, err := p.Generate(t.Context(), connect.NewRequest(&modelsv1.GenerateRequest{
		Model: "capable", Messages: []*modelsv1.Message{
			{Role: modelsv1.MessageRole_MESSAGE_ROLE_SYSTEM, Content: "s"},
			userMessage("u"),
			{Role: modelsv1.MessageRole_MESSAGE_ROLE_ASSISTANT, ToolCalls: []*modelsv1.ToolCall{{Id: "c0", Name: "lookup", ArgumentsJson: "{}"}}},
			{Role: modelsv1.MessageRole_MESSAGE_ROLE_TOOL, Content: "result", ToolCallId: "c0"},
		},
		Tools:              []*modelsv1.Tool{{Name: "lookup", Description: "d", ParametersSchemaJson: `{"type":"object"}`}},
		ResponseSchemaJson: `{"type":"object"}`,
	}))
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := &modelsv1.GenerateResponse{
		FinishReason: modelsv1.FinishReason_FINISH_REASON_TOOL_CALLS,
		ToolCalls:    []*modelsv1.ToolCall{{Id: "call-1", Name: "lookup", ArgumentsJson: `{"q":1}`}},
		Usage:        &modelsv1.Usage{InputTokens: 5, OutputTokens: 2},
	}
	if diff := cmp.Diff(want, resp.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("response (-want +got):\n%s", diff)
	}
	// The entry's limit applies when the request names none.
	if f.body["max_tokens"] != float64(100) || count(f.body["messages"]) != 4 || count(f.body["tools"]) != 1 || f.body["response_format"] == nil {
		t.Errorf("request as the provider saw it = %+v", f.body)
	}
}

// count is the length of a decoded JSON array, or -1 for anything else.
func count(v any) int {
	list, ok := v.([]any)
	if !ok {
		return -1
	}
	return len(list)
}

func TestGenerateAsksForTheStructuredAnswerByTheEntrysFlag(t *testing.T) {
	t.Parallel()
	strict := map[string]any{"type": "json_schema", "json_schema": map[string]any{
		"name": "response", "strict": true, "schema": map[string]any{"type": "object"},
	}}
	tests := []struct {
		name     string
		model    string
		fallback bool
		want     map[string]any
	}{
		{name: "an entry with structured output is held to the schema", model: "capable", want: strict},
		{name: "the fallback changes nothing for it", model: "capable", fallback: true, want: strict},
		{name: "an entry without it gets JSON mode when the request allows it", model: "plain", fallback: true, want: map[string]any{"type": "json_object"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, p := newProvider(t, textReply)
			if _, err := p.Generate(t.Context(), connect.NewRequest(&modelsv1.GenerateRequest{
				Model: tt.model, Messages: []*modelsv1.Message{userMessage("u")}, ResponseSchemaJson: `{"type":"object"}`, JsonModeFallback: tt.fallback,
			})); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			if diff := cmp.Diff(tt.want, f.body["response_format"]); diff != "" {
				t.Errorf("response_format as the provider saw it (-want +got):\n%s", diff)
			}
		})
	}
}

// TestGenerateRefusesAStrictRequestForAnEntryWithoutTheFlagAndAsksNothingForAFallbackAlone
// checks the two edges: a request that does not allow JSON mode is refused for an
// entry without structured output, as before, with no call; the fallback with no
// schema asks for no format at all.
func TestGenerateRefusesAStrictRequestForAnEntryWithoutTheFlagAndAsksNothingForAFallbackAlone(t *testing.T) {
	t.Parallel()
	f, p := newProvider(t, textReply)
	_, err := p.Generate(t.Context(), connect.NewRequest(&modelsv1.GenerateRequest{
		Model: "plain", Messages: []*modelsv1.Message{userMessage("u")}, ResponseSchemaJson: `{"type":"object"}`,
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument || f.path != "" {
		t.Fatalf("code = %v (%v), call made: %v; want a refusal and no call", connect.CodeOf(err), err, f.path != "")
	}
	if _, err := p.Generate(t.Context(), connect.NewRequest(&modelsv1.GenerateRequest{
		Model: "plain", Messages: []*modelsv1.Message{userMessage("u")}, JsonModeFallback: true,
	})); err != nil {
		t.Fatal(err)
	}
	if f.body["response_format"] != nil {
		t.Errorf("response_format = %v, want none", f.body["response_format"])
	}
}

func TestGenerateText(t *testing.T) {
	t.Parallel()
	_, p := newProvider(t, textReply)
	resp, err := p.Generate(t.Context(), connect.NewRequest(&modelsv1.GenerateRequest{
		Model: "plain", Messages: []*modelsv1.Message{userMessage("u")},
	}))
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if resp.Msg.GetContent() != "synthetic answer" || resp.Msg.GetFinishReason() != modelsv1.FinishReason_FINISH_REASON_STOP {
		t.Errorf("response = %v", resp.Msg)
	}
}

func TestGenerateRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		req  *modelsv1.GenerateRequest
	}{
		{name: "a model that is not an entry", req: &modelsv1.GenerateRequest{Model: "other", Messages: []*modelsv1.Message{userMessage("u")}}},
		{name: "no messages", req: &modelsv1.GenerateRequest{Model: "plain"}},
		{name: "ends with the assistant", req: &modelsv1.GenerateRequest{Model: "plain", Messages: []*modelsv1.Message{
			userMessage("u"), {Role: modelsv1.MessageRole_MESSAGE_ROLE_ASSISTANT, Content: "a"},
		}}},
		{name: "a message of no role", req: &modelsv1.GenerateRequest{Model: "plain", Messages: []*modelsv1.Message{{Content: "u"}}}},
		{name: "tools for a model without them", req: &modelsv1.GenerateRequest{
			Model: "plain", Messages: []*modelsv1.Message{userMessage("u")}, Tools: []*modelsv1.Tool{{Name: "t"}},
		}},
		{name: "a schema for a model without structured output", req: &modelsv1.GenerateRequest{
			Model: "plain", Messages: []*modelsv1.Message{userMessage("u")}, ResponseSchemaJson: `{}`,
		}},
		{name: "a tool without a name", req: &modelsv1.GenerateRequest{
			Model: "capable", Messages: []*modelsv1.Message{userMessage("u")}, Tools: []*modelsv1.Tool{{}},
		}},
		{name: "more output than the entry allows", req: &modelsv1.GenerateRequest{
			Model: "capable", Messages: []*modelsv1.Message{userMessage("u")}, MaxOutputTokens: 101,
		}},
		{name: "negative output", req: &modelsv1.GenerateRequest{
			Model: "plain", Messages: []*modelsv1.Message{userMessage("u")}, MaxOutputTokens: -1,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, p := newProvider(t, textReply)
			_, err := p.Generate(t.Context(), connect.NewRequest(tt.req))
			if connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Fatalf("code = %v (%v), want invalid argument", connect.CodeOf(err), err)
			}
			if f.path != "" {
				t.Error("the provider was called for a request the plugin refuses")
			}
		})
	}
}

func TestGenerateMapsProviderFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status int
		reply  string
		want   connect.Code
	}{
		{name: "key refused", status: http.StatusUnauthorized, reply: `{}`, want: connect.CodeUnavailable},
		{name: "provider down", status: http.StatusServiceUnavailable, reply: `{}`, want: connect.CodeUnavailable},
		{name: "request rejected", status: http.StatusBadRequest, reply: `{}`, want: connect.CodeInvalidArgument},
		{name: "reply not a completion", status: http.StatusOK, reply: `{"choices":[]}`, want: connect.CodeInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, p := newProvider(t, tt.reply)
			f.status = tt.status
			_, err := p.Generate(t.Context(), connect.NewRequest(&modelsv1.GenerateRequest{
				Model: "plain", Messages: []*modelsv1.Message{userMessage("synthetic question")},
			}))
			if got := connect.CodeOf(err); got != tt.want {
				t.Fatalf("code = %v (%v), want %v", got, err, tt.want)
			}
		})
	}
}

// TestAnEntrysHostsAndReasoningReachOnlyItsOwnRequestsOnOpenRouter checks the
// body of the request each entry of an OpenRouter instance sends: the pinned
// hosts with no fallback, and the reasoning setting, of that entry only, and
// require_parameters when the request needs it; an entry with none sends what
// it always did, and an instance of another server sends neither.
func TestAnEntrysHostsAndReasoningReachOnlyItsOwnRequestsOnOpenRouter(t *testing.T) {
	t.Parallel()
	const openRouter = "https://openrouter.ai/api/v1"
	raw := openRouterConfig()
	cfg, err := raw.parse()
	if err != nil {
		t.Fatal(err)
	}
	other := validConfig()
	other.Models = []rawModel{{ID: "vendor/synthetic-26b"}}
	otherCfg, err := other.parse()
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
		name          string
		cfg           config
		base          string
		req           *openaichat.Request
		wantProvider  any
		wantReasoning any
	}{
		{
			name: "a pinned entry with tools", cfg: cfg, base: openRouter, req: withTools("vendor/synthetic-26b"),
			wantProvider: map[string]any{"order": []any{"nextbit/bf16"}, "allow_fallbacks": false, "require_parameters": true},
		},
		{
			name: "a pinned entry without tools is still pinned", cfg: cfg, base: openRouter, req: plain("vendor/synthetic-26b"),
			wantProvider: map[string]any{"order": []any{"nextbit/bf16"}, "allow_fallbacks": false, "require_parameters": true},
		},
		{
			name: "a pinned entry with reasoning", cfg: cfg, base: openRouter, req: withTools("vendor/synthetic-120b"),
			wantProvider:  map[string]any{"order": []any{"groq", "crusoe"}, "allow_fallbacks": false, "require_parameters": true},
			wantReasoning: map[string]any{"effort": "low"},
		},
		{
			name: "an entry with no options keeps the usual body", cfg: cfg, base: openRouter, req: withTools("vendor/unpinned"),
			wantProvider: map[string]any{"require_parameters": true},
		},
		{name: "an entry with no options and no tools sends nothing", cfg: cfg, base: openRouter, req: plain("vendor/unpinned")},
		{name: "another server is told nothing", cfg: otherCfg, base: "https://models.example.org/v1", req: plain("vendor/synthetic-26b")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := &openaichat.Client{BaseURL: tt.base}
			client := tt.cfg.chatFor(base, tt.req.Model)
			body := wireBody(t, client.RequestBody(tt.req))
			if diff := cmp.Diff(tt.wantProvider, body["provider"]); diff != "" {
				t.Errorf("provider (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantReasoning, body["reasoning"]); diff != "" {
				t.Errorf("reasoning (-want +got):\n%s", diff)
			}
			if base.Route != nil || base.Reasoning != "" {
				t.Errorf("the shared client was changed: %+v", base)
			}
		})
	}
}

// wireBody returns body as the provider decodes it from the wire.
func wireBody(t *testing.T, body map[string]any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
