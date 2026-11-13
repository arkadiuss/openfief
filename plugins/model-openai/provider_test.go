package main

import (
	"net/http"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

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
