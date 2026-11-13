package connectrpc_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
	"github.com/arkadiuss/openfief/internal/modelrouter/connectrpc"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

// errProviderText is the text of a provider's reply, which no error of the
// client may quote; errNoService and errNoToken are failures to find the
// instance.
var (
	errProviderText = errors.New("SECRET-TEXT-FROM-THE-PROVIDER")
	errNoService    = errors.New("no service")
	errNoToken      = errors.New("no token")
)

// plugin is a fake model-provider plugin.
type plugin struct {
	modelsv1connect.UnimplementedModelProviderServiceHandler

	mu       sync.Mutex
	auth     string
	gotModel string
	gotReq   *modelsv1.GenerateRequest
	err      error
}

func (p *plugin) Describe(_ context.Context, req *connect.Request[modelsv1.DescribeRequest]) (*connect.Response[modelsv1.DescribeResponse], error) {
	p.seen(req.Header())
	return connect.NewResponse(&modelsv1.DescribeResponse{
		Protocol: modelsv1.WireProtocol_WIRE_PROTOCOL_OPENAI_COMPATIBLE,
		Location: modelsv1.Location_LOCATION_OWNER_HARDWARE, Retention: &modelsv1.Retention{Days: 30, MayTrain: false},
	}), nil
}

func (p *plugin) ListModels(_ context.Context, req *connect.Request[modelsv1.ListModelsRequest]) (*connect.Response[modelsv1.ListModelsResponse], error) {
	p.seen(req.Header())
	return connect.NewResponse(&modelsv1.ListModelsResponse{Models: []*modelsv1.ModelEntry{
		{Id: "openai/gpt-oss-120b", InputPrice: "0.15", OutputPrice: "0.60", PriceCurrency: "USD", ContextTokens: 131072, MaxOutputTokens: 8192, ToolCalling: true, StructuredOutput: true},
		{Id: "qwen3:8b", ToolCalling: true},
	}}), nil
}

func (p *plugin) Generate(_ context.Context, req *connect.Request[modelsv1.GenerateRequest]) (*connect.Response[modelsv1.GenerateResponse], error) {
	p.seen(req.Header())
	p.mu.Lock()
	p.gotReq = req.Msg
	p.gotModel = req.Msg.GetModel()
	err := p.err
	p.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&modelsv1.GenerateResponse{
		Content: "ok", FinishReason: modelsv1.FinishReason_FINISH_REASON_TOOL_CALLS,
		ToolCalls: []*modelsv1.ToolCall{{Id: "c1", Name: "show_account", ArgumentsJson: `{"id":"a"}`}},
		Usage:     &modelsv1.Usage{InputTokens: 11, OutputTokens: 7},
	}), nil
}

func (p *plugin) seen(h http.Header) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.auth = h.Get("Authorization")
}

type tokens struct{ err error }

func (t tokens) TokenFor(_ context.Context, instance string) (string, error) {
	return "token-for-" + instance, t.err
}

func newClient(t *testing.T, p *plugin, tok tokens) *connectrpc.Client {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(modelsv1connect.NewModelProviderServiceHandler(p))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return connectrpc.NewClient(srv.Client(), func(string) (string, error) { return srv.URL, nil }, tok)
}

func TestDescribe(t *testing.T) {
	t.Parallel()
	p := &plugin{}
	got, err := newClient(t, p, tokens{}).Describe(t.Context(), "ollama")
	if err != nil {
		t.Fatal(err)
	}
	want := modelrouter.Provider{Location: modelrouter.LocationOwnerHardware, RetentionDays: 30}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("provider (-want +got):\n%s", diff)
	}
	if p.auth != "Bearer token-for-ollama" {
		t.Errorf("Authorization = %q, want the core's token for the instance", p.auth)
	}
}

func TestModels(t *testing.T) {
	t.Parallel()
	got, err := newClient(t, &plugin{}, tokens{}).Models(t.Context(), "openrouter")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d entries", len(got))
	}
	if e := got[0]; !pricedEntry(&e) {
		t.Errorf("entry = %+v", e)
	}
	if got[1].HasPrice || got[1].StructuredOut {
		t.Errorf("a local entry = %+v, want no price and no structured output", got[1])
	}
}

// pricedEntry reports whether e is the plugin's priced entry, translated.
func pricedEntry(e *modelrouter.Entry) bool {
	return e.ID == "openai/gpt-oss-120b" && e.HasPrice && e.InputPrice.String() == "0.15" && e.OutputPrice.String() == "0.6" &&
		e.PriceCurrency.String() == "USD" && e.ContextTokens == 131072 && e.MaxOutputTokens == 8192 && e.ToolCalling && e.StructuredOut
}

func TestGenerateTranslatesBothWays(t *testing.T) {
	t.Parallel()
	p := &plugin{}
	req := &modelrouter.Request{
		Messages: []modelrouter.Message{
			{Role: modelrouter.MessageSystem, Content: "be brief", Tier: modelrouter.TierT2},
			{Role: modelrouter.MessageUser, Content: "q", Tier: modelrouter.TierT0},
			{Role: modelrouter.MessageAssistant, ToolCalls: []modelrouter.ToolCall{{ID: "c0", Name: "x", ArgumentsJSON: "{}"}}, Tier: modelrouter.TierT0},
			{Role: modelrouter.MessageTool, ToolCallID: "c0", Content: "r", Tier: modelrouter.TierT0},
		},
		Tools:              []modelrouter.Tool{{Name: "show_account", Description: "d", ParametersSchemaJSON: `{"type":"object"}`}},
		ResponseSchemaJSON: `{"type":"object"}`, MaxOutputTokens: 800,
	}
	got, err := newClient(t, p, tokens{}).Generate(t.Context(), "openrouter", "openai/gpt-oss-120b", req)
	if err != nil {
		t.Fatal(err)
	}
	want := modelrouter.Result{
		Content: "ok", Finish: modelrouter.FinishToolCalls,
		ToolCalls: []modelrouter.ToolCall{{ID: "c1", Name: "show_account", ArgumentsJSON: `{"id":"a"}`}},
		Usage:     modelrouter.Usage{InputTokens: 11, OutputTokens: 7},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("result (-want +got):\n%s", diff)
	}
	wantReq := &modelsv1.GenerateRequest{
		Model: "openai/gpt-oss-120b",
		Messages: []*modelsv1.Message{
			{Role: modelsv1.MessageRole_MESSAGE_ROLE_SYSTEM, Content: "be brief"},
			{Role: modelsv1.MessageRole_MESSAGE_ROLE_USER, Content: "q"},
			{Role: modelsv1.MessageRole_MESSAGE_ROLE_ASSISTANT, ToolCalls: []*modelsv1.ToolCall{{Id: "c0", Name: "x", ArgumentsJson: "{}"}}},
			{Role: modelsv1.MessageRole_MESSAGE_ROLE_TOOL, ToolCallId: "c0", Content: "r"},
		},
		Tools:              []*modelsv1.Tool{{Name: "show_account", Description: "d", ParametersSchemaJson: `{"type":"object"}`}},
		ResponseSchemaJson: `{"type":"object"}`, MaxOutputTokens: 800,
	}
	if diff := cmp.Diff(wantReq, p.gotReq, protocmp.Transform()); diff != "" {
		t.Errorf("the plugin received (-want +got):\n%s", diff)
	}
}

func TestGenerateErrorClasses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		code         connect.Code
		wantRejected bool
	}{
		{"invalid argument is a rejection", connect.CodeInvalidArgument, true},
		{"not found (an unknown model) is a rejection", connect.CodeNotFound, true},
		{"unavailable falls back", connect.CodeUnavailable, false},
		{"a deadline falls back", connect.CodeDeadlineExceeded, false},
		{"an internal error falls back", connect.CodeInternal, false},
		{"a refused core token falls back", connect.CodeUnauthenticated, false},
	}
	secret := errProviderText.Error()
	req := &modelrouter.Request{Messages: []modelrouter.Message{{Role: modelrouter.MessageUser, Content: "q", Tier: modelrouter.TierT2}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &plugin{err: connect.NewError(tt.code, fmt.Errorf("%w", errProviderText))}
			_, err := newClient(t, p, tokens{}).Generate(t.Context(), "p", "m", req)
			if err == nil {
				t.Fatal("no error")
			}
			if got := errors.Is(err, app.ErrProviderRejected); got != tt.wantRejected {
				t.Errorf("rejected = %v, want %v (%v)", got, tt.wantRejected, err)
			}
			if got := errors.Is(err, app.ErrProviderUnavailable); got == tt.wantRejected {
				t.Errorf("unavailable = %v, want %v (%v)", got, !tt.wantRejected, err)
			}
			if strings.Contains(err.Error(), secret) {
				t.Errorf("the error quotes the provider's reply: %v", err)
			}
		})
	}
}

func TestAnInstanceWithoutAnAddressOrTokenIsUnavailable(t *testing.T) {
	t.Parallel()
	req := &modelrouter.Request{Messages: []modelrouter.Message{{Role: modelrouter.MessageUser, Content: "q", Tier: modelrouter.TierT2}}}
	noAddr := connectrpc.NewClient(http.DefaultClient, func(string) (string, error) { return "", errNoService }, tokens{})
	if _, err := noAddr.Generate(t.Context(), "p", "m", req); !errors.Is(err, app.ErrProviderUnavailable) {
		t.Errorf("no address: %v", err)
	}
	noToken := newClient(t, &plugin{}, tokens{err: errNoToken})
	if _, err := noToken.Generate(t.Context(), "p", "m", req); !errors.Is(err, app.ErrProviderUnavailable) {
		t.Errorf("no token: %v", err)
	}
	if _, err := noToken.Models(t.Context(), "p"); !errors.Is(err, app.ErrProviderUnavailable) {
		t.Errorf("no token, Models: %v", err)
	}
}
