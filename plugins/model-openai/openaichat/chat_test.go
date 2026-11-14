package openaichat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// fakeKey is a synthetic API key.
const fakeKey = "synthetic-key-0000"

// fakeProvider is an OpenAI-compatible server: it records the last request
// and answers with status and body.
type fakeProvider struct {
	mu     sync.Mutex
	auth   string
	path   string
	body   map[string]any
	status int
	reply  string
}

func (f *fakeProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)
	var body map[string]any
	if err != nil || json.Unmarshal(raw, &body) != nil {
		http.Error(w, "unreadable request", http.StatusBadRequest)
		return
	}
	f.mu.Lock()
	f.auth, f.path, f.body = r.Header.Get("Authorization"), r.URL.Path, body
	status, reply := f.status, f.reply
	f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := io.WriteString(w, reply); err != nil {
		return // the client went away; the test sees its own error
	}
}

// newFake starts a fake provider that answers 200 with reply.
func newFake(t *testing.T, reply string) (*fakeProvider, *Client) {
	t.Helper()
	f := &fakeProvider{status: http.StatusOK, reply: reply}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	return f, &Client{HTTP: srv.Client(), BaseURL: srv.URL + "/v1", Key: func() (string, error) { return fakeKey, nil }}
}

const textReply = `{"choices":[{"message":{"role":"assistant","content":"synthetic answer"},"finish_reason":"stop"}],
"usage":{"prompt_tokens":11,"completion_tokens":7}}`

func TestCompleteSendsTheChatProtocol(t *testing.T) {
	t.Parallel()
	f, c := newFake(t, textReply)
	req := &Request{
		Model:          "synthetic-8b",
		Messages:       []Message{{Role: "system", Content: "be brief"}, {Role: "user", Content: "hi"}},
		Tools:          []Tool{{Name: "lookup", Parameters: json.RawMessage(`{"type":"object"}`)}},
		ResponseSchema: json.RawMessage(`{"type":"object"}`),
		MaxTokens:      64,
	}
	reply, err := c.Complete(t.Context(), req)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if reply.Content != "synthetic answer" || reply.InputTokens != 11 || reply.OutputTokens != 7 || reply.FinishReason != "stop" {
		t.Errorf("reply = %+v", reply)
	}
	if f.path != "/v1/chat/completions" || f.auth != "Bearer "+fakeKey {
		t.Errorf("path %q, authorization %q", f.path, f.auth)
	}
	want := map[string]any{
		"model":      "synthetic-8b",
		"max_tokens": float64(64),
		"messages": []any{
			map[string]any{"role": "system", "content": "be brief"},
			map[string]any{"role": "user", "content": "hi"},
		},
		"tools": []any{map[string]any{"type": "function", "function": map[string]any{
			"name": "lookup", "parameters": map[string]any{"type": "object"},
		}}},
		"response_format": map[string]any{"type": "json_schema", "json_schema": map[string]any{
			"name": "response", "strict": true, "schema": map[string]any{"type": "object"},
		}},
	}
	if diff := cmp.Diff(want, f.body); diff != "" {
		t.Errorf("request as the provider saw it (-want +got):\n%s", diff)
	}
}

// TestRequestBodyRequiresParametersOnOpenRouterWithTools checks the request
// body per base URL and per whether it sends tools: only OpenRouter, and only
// with tools, gets provider.require_parameters, so a host that would drop the
// tools is never picked; a server that refuses unknown members, such as
// OpenAI's own, never sees the member.
func TestRequestBodyRequiresParametersOnOpenRouterWithTools(t *testing.T) {
	t.Parallel()
	tools := []Tool{{Name: "lookup", Parameters: json.RawMessage(`{"type":"object"}`)}}
	tests := []struct {
		name    string
		baseURL string
		tools   []Tool
		want    bool
	}{
		{name: "OpenRouter with tools", baseURL: "https://openrouter.ai/api/v1", tools: tools, want: true},
		{name: "OpenRouter host in capitals", baseURL: "https://OpenRouter.ai/api/v1", tools: tools, want: true},
		{name: "OpenRouter without tools", baseURL: "https://openrouter.ai/api/v1", want: false},
		{name: "OpenAI with tools", baseURL: "https://api.openai.com/v1", tools: tools, want: false},
		{name: "Ollama with tools", baseURL: "http://ollama.models.svc:11434/v1", tools: tools, want: false},
		{name: "a host that only ends in the name", baseURL: "https://notopenrouter.ai/v1", tools: tools, want: false},
		{name: "OpenRouter's name in the path", baseURL: "https://example.com/openrouter.ai/v1", tools: tools, want: false},
		{name: "an unparsable base URL", baseURL: "://", tools: tools, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{BaseURL: tt.baseURL}
			body := c.RequestBody(&Request{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}, Tools: tt.tools})
			got, has := body["provider"]
			if has != tt.want {
				t.Fatalf("provider member present = %v, want %v", has, tt.want)
			}
			if tt.want {
				if diff := cmp.Diff(map[string]any{"require_parameters": true}, got); diff != "" {
					t.Errorf("provider (-want +got):\n%s", diff)
				}
			}
			raw, err := json.Marshal(body)
			if err != nil || !json.Valid(raw) {
				t.Errorf("the body does not encode: %v", err)
			}
		})
	}
}

// TestCompleteSendsRequireParametersToOpenRouterOnTheWire checks the member
// reaches the provider in the posted JSON, as the member the protocol names.
func TestCompleteSendsRequireParametersToOpenRouterOnTheWire(t *testing.T) {
	t.Parallel()
	f, c := newFake(t, textReply)
	c.HTTP = &http.Client{Transport: rewriteHost{to: c.BaseURL, rt: c.HTTP.Transport}}
	c.BaseURL = "https://openrouter.ai/api/v1"
	req := &Request{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}, Tools: []Tool{{Name: "lookup"}}}
	if _, err := c.Complete(t.Context(), req); err != nil {
		t.Fatalf("complete: %v", err)
	}
	want := map[string]any{"require_parameters": true}
	if diff := cmp.Diff(want, f.body["provider"]); diff != "" {
		t.Errorf("provider as the server saw it (-want +got):\n%s", diff)
	}
}

// rewriteHost sends every request to the fake server at to, whatever its
// address, so a test can name OpenRouter's base URL without any network.
type rewriteHost struct {
	to string
	rt http.RoundTripper
}

func (r rewriteHost) RoundTrip(req *http.Request) (*http.Response, error) {
	target, err := url.Parse(r.to)
	if err != nil {
		return nil, fmt.Errorf("parsing the fake's address: %w", err)
	}
	req = req.Clone(req.Context())
	req.URL.Scheme, req.URL.Host = target.Scheme, target.Host
	return r.rt.RoundTrip(req) //nolint:wrapcheck // a test helper: the client's own error
}

func TestCompleteReadsToolCalls(t *testing.T) {
	t.Parallel()
	_, c := newFake(t, `{"choices":[{"message":{"content":null,"tool_calls":[
		{"id":"call-1","type":"function","function":{"name":"lookup","arguments":"{\"q\":1}"}}]},"finish_reason":"tool_calls"}],
		"usage":{"prompt_tokens":5,"completion_tokens":2}}`)
	reply, err := c.Complete(t.Context(), &Request{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	want := &Reply{
		ToolCalls: []ToolCall{{ID: "call-1", Name: "lookup", Arguments: `{"q":1}`}}, FinishReason: "tool_calls", InputTokens: 5, OutputTokens: 2,
	}
	if diff := cmp.Diff(want, reply); diff != "" {
		t.Errorf("reply (-want +got):\n%s", diff)
	}
}

func TestCompleteSendsToolCallsAndResults(t *testing.T) {
	t.Parallel()
	f, c := newFake(t, textReply)
	_, err := c.Complete(t.Context(), &Request{Model: "m", Messages: []Message{
		{Role: "assistant", ToolCalls: []ToolCall{{ID: "c0", Name: "lookup", Arguments: "{}"}}},
		{Role: "tool", Content: "result", ToolCallID: "c0"},
	}})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	want := []any{
		map[string]any{"role": "assistant", "content": "", "tool_calls": []any{map[string]any{
			"id": "c0", "type": "function", "function": map[string]any{"name": "lookup", "arguments": "{}"},
		}}},
		map[string]any{"role": "tool", "content": "result", "tool_call_id": "c0"},
	}
	if diff := cmp.Diff(want, f.body["messages"]); diff != "" {
		t.Errorf("messages as the provider saw them (-want +got):\n%s", diff)
	}
}

func TestCompleteWithoutAKeySendsNoAuthorization(t *testing.T) {
	t.Parallel()
	f, c := newFake(t, textReply)
	c.Key = func() (string, error) { return "", nil }
	if _, err := c.Complete(t.Context(), &Request{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}}); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if f.auth != "" {
		t.Errorf("authorization = %q, want none", f.auth)
	}
}

func TestCompleteErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status int
		reply  string
		want   error
	}{
		{name: "bad request", status: 400, reply: `{"error":"bad"}`, want: ErrRejected},
		{name: "unknown model", status: 404, reply: `{}`, want: ErrRejected},
		{name: "unprocessable", status: 422, reply: `{}`, want: ErrRejected},
		{name: "key refused", status: 401, reply: `{}`, want: ErrUnavailable},
		{name: "forbidden", status: 403, reply: `{}`, want: ErrUnavailable},
		{name: "rate limited", status: 429, reply: `{}`, want: ErrUnavailable},
		{name: "provider failing", status: 502, reply: `{}`, want: ErrUnavailable},
		{name: "not JSON", status: 200, reply: `<html>`, want: ErrBadReply},
		{name: "no choices", status: 200, reply: `{"choices":[]}`, want: ErrBadReply},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, c := newFake(t, tt.reply)
			f.status = tt.status
			_, err := c.Complete(t.Context(), &Request{Model: "m", Messages: []Message{{Role: "user", Content: "synthetic question"}}})
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if strings.Contains(err.Error(), "synthetic question") || strings.Contains(err.Error(), fakeKey) {
				t.Errorf("error %q quotes the request or the key", err)
			}
		})
	}
}

func TestCompleteUnreachableProviderIsUnavailable(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.NotFoundHandler())
	addr := srv.URL
	srv.Close()
	c := &Client{HTTP: http.DefaultClient, BaseURL: addr, Key: func() (string, error) { return fakeKey, nil }}
	_, err := c.Complete(t.Context(), &Request{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error = %v, want %v", err, ErrUnavailable)
	}
	if strings.Contains(err.Error(), addr) {
		t.Errorf("error %q names the URL", err)
	}
}

func TestCompleteErrorsCarryTheStatusAndNothingElse(t *testing.T) {
	t.Parallel()
	for _, status := range []int{400, 401, 402, 404, 429, 500, 503} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			t.Parallel()
			f, c := newFake(t, `{"error":"secret provider text `+fakeKey+`"}`)
			f.status = status
			_, err := c.Complete(t.Context(), &Request{Model: "m"})
			var ce *CallError
			if !errors.As(err, &ce) || ce.Status != status || ce.Timeout {
				t.Fatalf("error = %#v, want a CallError of status %d", err, status)
			}
			if want := fmt.Sprintf("HTTP %d", status); !strings.HasSuffix(err.Error(), want) {
				t.Errorf("error text %q does not end in %q: the text users see changed", err, want)
			}
			if strings.Contains(err.Error(), "secret provider text") || strings.Contains(err.Error(), fakeKey) {
				t.Errorf("error %q quotes the provider's reply or the key", err)
			}
		})
	}
}

func TestCompleteNoAnswerCarriesNoStatus(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.NotFoundHandler())
	addr := srv.URL
	srv.Close()
	c := &Client{HTTP: http.DefaultClient, BaseURL: addr, Key: func() (string, error) { return fakeKey, nil }}
	_, err := c.Complete(t.Context(), &Request{Model: "m"})
	var ce *CallError
	if !errors.As(err, &ce) || ce.Status != 0 || ce.Timeout {
		t.Fatalf("error = %#v, want a CallError with no status and no timeout", err)
	}
}

// errKeyUnreadable stands for a key the caller could not read.
var errKeyUnreadable = errors.New("the key is unreadable")

func TestCompleteKeyErrorIsUnavailable(t *testing.T) {
	t.Parallel()
	_, c := newFake(t, textReply)
	c.Key = func() (string, error) { return "", errKeyUnreadable }
	if _, err := c.Complete(t.Context(), &Request{Model: "m"}); !errors.Is(err, ErrUnavailable) || !errors.Is(err, errKeyUnreadable) {
		t.Fatalf("error = %v, want unavailable for the key", err)
	}
}
