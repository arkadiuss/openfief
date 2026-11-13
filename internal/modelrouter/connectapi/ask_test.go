package connectapi_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
	"github.com/arkadiuss/openfief/internal/modelrouter/connectapi"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

// secretText is in every request and answer of these tests: it must reach a
// model and nothing else (GR-12).
const secretText = "SECRET-QUESTION-AND-ANSWER-TEXT"

// errSecret is an error whose text is the secret.
var errSecret = errors.New(secretText)

// askers admits the session tokens "owner" and "read".
type askers struct{}

func (askers) Asker(_ context.Context, _, token string) (string, error) {
	switch token {
	case "owner", "read":
		return "session-" + token, nil
	}
	return "", app.ErrNoSession
}

// provider is the model provider the router calls: it answers with a script
// and keeps the requests it was sent.
type provider struct {
	mu       sync.Mutex
	err      error
	reply    modelrouter.Result
	requests []*modelrouter.Request
}

func (p *provider) Instance(context.Context, string) error { return nil }

func (p *provider) Describe(context.Context, string) (modelrouter.ProviderInstance, error) {
	return modelrouter.ProviderInstance{}, nil
}

func (p *provider) Models(context.Context, string) ([]modelrouter.ModelEntry, error) {
	return []modelrouter.ModelEntry{{ID: "m", ToolCalling: true}}, nil
}

func (p *provider) Generate(_ context.Context, _, _ string, req *modelrouter.Request) (modelrouter.Result, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.requests = append(p.requests, req)
	return p.reply, p.err
}

// audit keeps the two entries of each call.
type audit struct {
	mu      sync.Mutex
	sent    []app.CallSent
	results []app.CallResult
}

func (a *audit) RecordSent(_ context.Context, c *app.CallSent) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sent = append(a.sent, *c)
	return nil
}

func (a *audit) RecordResult(_ context.Context, c *app.CallResult) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.results = append(a.results, *c)
	return nil
}

type noAttention struct{}

func (noAttention) Raise(context.Context, string, string, string) error { return nil }
func (noAttention) Clear(context.Context, string, string) error         { return nil }

// askFixture is the AskModelService over a router with one model that may
// receive T2 and T0.
type askFixture struct {
	client   modelsv1connect.AskModelServiceClient
	store    *memStore
	provider *provider
	audit    *audit
	logs     *strings.Builder
}

func newAskFixture(t *testing.T) *askFixture {
	t.Helper()
	m := modelrouter.Ref{Instance: "p", Model: "m"}
	store := &memStore{s: modelrouter.Settings{
		Switch: modelrouter.Switch{Enabled: true, Toggles: map[modelrouter.Feature]bool{modelrouter.FeatureAsk: true}},
		Routes: map[modelrouter.Role]modelrouter.Route{modelrouter.RoleAsk: {Role: modelrouter.RoleAsk, Models: []modelrouter.Ref{m}}},
		Policy: modelrouter.DataTierPolicy{modelrouter.DataTierT2: {m}, modelrouter.DataTierT0: {m}},
	}}
	prov := &provider{reply: modelrouter.Result{Content: "the answer", Finish: modelrouter.FinishStop}}
	aud := &audit{}
	n := 0
	router := app.NewRouter(&app.RouterPorts{
		Settings: store, Providers: prov, Audit: aud, Attention: noAttention{},
		CallIDs: func() (string, error) { n++; return fmt.Sprintf("%016x", n), nil },
		Now:     func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) },
	})
	logs := &strings.Builder{}
	path, h := connectapi.NewAsk(app.NewAsk(askers{}, store, router), passThrough{}, slog.New(slog.NewTextHandler(logs, nil)))
	mux := http.NewServeMux()
	mux.Handle(path, h)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &askFixture{
		client: modelsv1connect.NewAskModelServiceClient(srv.Client(), srv.URL), store: store, provider: prov, audit: aud, logs: logs,
	}
}

func userMessage(tier string) *modelsv1.AskMessage {
	return &modelsv1.AskMessage{Message: &modelsv1.Message{Role: modelsv1.MessageRole_MESSAGE_ROLE_USER, Content: secretText}, DataTier: tier}
}

func (f *askFixture) generate(token string, msgs ...*modelsv1.AskMessage) (*modelsv1.GenerateAskResponse, error) {
	resp, err := f.client.GenerateAsk(context.Background(), connect.NewRequest(&modelsv1.GenerateAskRequest{
		Origin: "o", SessionToken: token, Messages: msgs,
		Tools: []*modelsv1.Tool{{Name: "show", Description: "d", ParametersSchemaJson: `{"type":"object"}`}},
	}))
	if err != nil {
		return nil, fmt.Errorf("GenerateAsk: %w", err)
	}
	return resp.Msg, nil
}

func TestAskServiceGeneratesThroughTheRouter(t *testing.T) {
	t.Parallel()
	f := newAskFixture(t)
	f.provider.reply = modelrouter.Result{
		Content: "looking", Finish: modelrouter.FinishToolCalls,
		ToolCalls: []modelrouter.ToolCall{{ID: "call-1", Name: "show", ArgumentsJSON: `{"name":"x"}`}},
	}
	assistant := &modelsv1.AskMessage{DataTier: "T2", Message: &modelsv1.Message{
		Role: modelsv1.MessageRole_MESSAGE_ROLE_ASSISTANT, ToolCalls: []*modelsv1.ToolCall{{Id: "c0", Name: "show", ArgumentsJson: "{}"}},
	}}
	tool := &modelsv1.AskMessage{DataTier: "T0", Message: &modelsv1.Message{Role: modelsv1.MessageRole_MESSAGE_ROLE_TOOL, ToolCallId: "c0", Content: `{"title":"t"}`}}
	got, err := f.generate("read", userMessage("T2"), assistant, tool)
	if err != nil {
		t.Fatal(err)
	}
	want := &modelsv1.GenerateAskResponse{
		Content: "looking", FinishReason: modelsv1.FinishReason_FINISH_REASON_TOOL_CALLS, Model: "p/m",
		ToolCalls: []*modelsv1.ToolCall{{Id: "call-1", Name: "show", ArgumentsJson: `{"name":"x"}`}},
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("response (-want +got):\n%s", diff)
	}
	sent := f.provider.requests[0]
	wantMsgs := []modelrouter.Message{
		{Role: modelrouter.MessageUser, Content: secretText, DataTier: modelrouter.DataTierT2},
		{Role: modelrouter.MessageAssistant, DataTier: modelrouter.DataTierT2, ToolCalls: []modelrouter.ToolCall{{ID: "c0", Name: "show", ArgumentsJSON: "{}"}}},
		{Role: modelrouter.MessageTool, ToolCallID: "c0", Content: `{"title":"t"}`, DataTier: modelrouter.DataTierT0},
	}
	if diff := cmp.Diff(wantMsgs, sent.Messages); diff != "" {
		t.Errorf("messages the model got (-want +got):\n%s", diff)
	}
	if len(sent.Tools) != 1 || sent.Tools[0].Name != "show" {
		t.Errorf("tools the model got = %+v", sent.Tools)
	}
	if len(f.audit.sent) != 1 || f.audit.sent[0].Role != modelrouter.RoleAsk || f.audit.sent[0].Actor.ID != "session-read" {
		t.Errorf("audit sent = %+v, want one entry of the ask role for the reader's session", f.audit.sent)
	}
}

func TestAskServiceRefusalsAndFailuresByKind(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		token string
		msgs  []*modelsv1.AskMessage
		mod   func(f *askFixture)
		want  connect.Code
	}{
		{name: "no session", token: "nobody", msgs: []*modelsv1.AskMessage{userMessage("T2")}, want: connect.CodeUnauthenticated},
		{name: "a message with no tier", token: "owner", msgs: []*modelsv1.AskMessage{userMessage("")}, want: connect.CodeInvalidArgument},
		{name: "a tier that is none", token: "owner", msgs: []*modelsv1.AskMessage{userMessage("T9")}, want: connect.CodeInvalidArgument},
		{name: "no messages", token: "owner", want: connect.CodeInvalidArgument},
		{name: "a message with no role", token: "owner", msgs: []*modelsv1.AskMessage{{DataTier: "T2", Message: &modelsv1.Message{}}}, want: connect.CodeInvalidArgument},
		{
			name: "no model may receive the tier", token: "owner", msgs: []*modelsv1.AskMessage{userMessage("T2")},
			mod:  func(f *askFixture) { f.store.s.Policy = modelrouter.DataTierPolicy{} },
			want: connect.CodePermissionDenied,
		},
		{
			name: "the switch is off", token: "owner", msgs: []*modelsv1.AskMessage{userMessage("T2")},
			mod:  func(f *askFixture) { f.store.s.Switch.Enabled = false },
			want: connect.CodeFailedPrecondition,
		},
		{
			name: "every model is down", token: "owner", msgs: []*modelsv1.AskMessage{userMessage("T2")},
			mod:  func(f *askFixture) { f.provider.err = fmt.Errorf("%w: %s", app.ErrProviderUnavailable, secretText) },
			want: connect.CodeUnavailable,
		},
		{
			name: "the provider rejects the request", token: "owner", msgs: []*modelsv1.AskMessage{userMessage("T2")},
			mod:  func(f *askFixture) { f.provider.err = fmt.Errorf("%w: %s", app.ErrProviderRejected, secretText) },
			want: connect.CodeInvalidArgument,
		},
		{
			name: "something else", token: "owner", msgs: []*modelsv1.AskMessage{userMessage("T2")},
			mod:  func(f *askFixture) { f.provider.err = errSecret },
			want: connect.CodeInternal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newAskFixture(t)
			if tt.mod != nil {
				tt.mod(f)
			}
			_, err := f.generate(tt.token, tt.msgs...)
			if got := connect.CodeOf(err); got != tt.want {
				t.Fatalf("code = %v (%v), want %v", got, err, tt.want)
			}
			f.checkNoText(t, err)
		})
	}
}

// checkNoText checks that neither err nor the log holds the secret text
// (GR-12) and that every audit entry is the ask role's.
func (f *askFixture) checkNoText(t *testing.T, err error) {
	t.Helper()
	if strings.Contains(err.Error(), secretText) || strings.Contains(f.logs.String(), secretText) {
		t.Errorf("the error %q or the log %q holds the question or the answer (GR-12)", err, f.logs.String())
	}
	for _, e := range f.audit.sent {
		if e.Role != modelrouter.RoleAsk {
			t.Errorf("audit entry for role %s", e.Role)
		}
	}
}

func TestAskServiceRefusesTooMuch(t *testing.T) {
	t.Parallel()
	f := newAskFixture(t)
	msgs := make([]*modelsv1.AskMessage, 65)
	for i := range msgs {
		msgs[i] = userMessage("T2")
	}
	if _, err := f.generate("owner", msgs...); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("65 messages: code = %v, want InvalidArgument", connect.CodeOf(err))
	}
	if len(f.provider.requests) != 0 {
		t.Errorf("a model was called with %d requests", len(f.provider.requests))
	}
}

// available asks whether the Ask box has a model for the session token.
func (f *askFixture) available(token string) (bool, connect.Code) {
	resp, err := f.client.GetAskModel(context.Background(), connect.NewRequest(&modelsv1.GetAskModelRequest{Origin: "o", SessionToken: token}))
	if err != nil {
		return false, connect.CodeOf(err)
	}
	return resp.Msg.GetAvailable(), 0
}

// mayReceive asks whether every one of tiers may go to a model together.
func (f *askFixture) mayReceive(tiers ...string) (bool, connect.Code) {
	resp, err := f.client.CheckAskTiers(context.Background(), connect.NewRequest(&modelsv1.CheckAskTiersRequest{Origin: "o", SessionToken: "read", Tiers: tiers}))
	if err != nil {
		return false, connect.CodeOf(err)
	}
	return resp.Msg.GetAllowed(), 0
}

func TestGetAskModel(t *testing.T) {
	t.Parallel()
	f := newAskFixture(t)
	if got, code := f.available("read"); !got || code != 0 {
		t.Errorf("GetAskModel = %v, %v; want available", got, code)
	}
	if _, code := f.available("nobody"); code != connect.CodeUnauthenticated {
		t.Errorf("GetAskModel without a session: code = %v, want Unauthenticated", code)
	}
	f.store.s.Switch.Enabled = false
	if got, code := f.available("read"); got || code != 0 {
		t.Errorf("GetAskModel with the switch off = %v, %v; want unavailable", got, code)
	}
	if len(f.audit.sent)+len(f.audit.results) != 0 {
		t.Errorf("the checks wrote audit entries %v %v", f.audit.sent, f.audit.results)
	}
}

func TestCheckAskTiers(t *testing.T) {
	t.Parallel()
	f := newAskFixture(t)
	tests := []struct {
		tiers []string
		want  bool
		code  connect.Code
	}{
		{[]string{"T2", "T0"}, true, 0},
		{[]string{"T1"}, false, 0},
		{[]string{"T9"}, false, connect.CodeInvalidArgument},
	}
	for _, tt := range tests {
		if got, code := f.mayReceive(tt.tiers...); got != tt.want || code != tt.code {
			t.Errorf("CheckAskTiers(%v) = %v, %v; want %v, %v", tt.tiers, got, code, tt.want, tt.code)
		}
	}
	if len(f.audit.sent)+len(f.audit.results) != 0 || len(f.provider.requests) != 0 {
		t.Errorf("the checks wrote audit entries or called a model")
	}
}
