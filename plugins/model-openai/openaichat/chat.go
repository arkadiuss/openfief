// Package openaichat is the OpenAI-compatible chat completions client of the
// model-openai plugin (D-205): the request and reply wire format, the status
// classes and OpenRouter's routing preference. The plugin serves the core with
// it, and the Ask box bench (tools/askbench) sends its requests through the
// same code, so both put the same bytes on the wire.
package openaichat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// maxReplyBytes caps what is read of the provider's reply.
const maxReplyBytes = 4 << 20

// Errors of a call to the provider. None quotes the provider's reply or the
// request: either may hold the owner's data (GR-12).
var (
	// ErrRejected reports a request the provider refused as malformed or
	// unknown, such as a model it does not serve (HTTP 400, 404, 413, 422).
	ErrRejected = errors.New("the provider rejected the request")
	// ErrUnavailable reports a provider that could not answer: unreachable,
	// slow, failing, rate-limiting, or refusing the key.
	ErrUnavailable = errors.New("the provider is unavailable")
	// ErrBadReply reports an answer that is not a chat completion.
	ErrBadReply = errors.New("the provider's reply is not a chat completion")
)

// Client speaks the OpenAI-compatible chat completions protocol: the
// first wire protocol (D-205), which OpenAI, OpenRouter, Ollama, vLLM, LM
// Studio and others serve. HTTP must reach only the provider (the egress
// proxy and the endpoint's confinement), and Key is read per call so a
// rotated Secret works at once; an empty key sends no Authorization header.
type Client struct {
	// HTTP sends the requests.
	HTTP *http.Client
	// BaseURL is the provider's API root, such as https://api.openai.com/v1.
	BaseURL string
	// Key returns the API key of a call.
	Key func() (string, error)
}

// Message is a message of the conversation.
type Message struct {
	Role       string
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
}

// ToolCall is a call of a function the model asked for.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// Tool is a function the model may call, with its JSON Schema.
type Tool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// Request is what one POST /chat/completions asks.
type Request struct {
	Model    string
	Messages []Message
	Tools    []Tool
	// ResponseSchema, when set, asks for a JSON answer of this JSON Schema.
	ResponseSchema json.RawMessage
	MaxTokens      int64
}

// Reply is the part of a completion the plugin reads.
type Reply struct {
	Content      string
	ToolCalls    []ToolCall
	FinishReason string
	InputTokens  int64
	OutputTokens int64
}

// object is a JSON object whose members are read by name. The OpenAI wire
// format names members in snake case, which no struct tag here may spell
// (the repository's tagliatelle rule is camel case), so the format is
// written and read through object and the names below.
type object map[string]json.RawMessage

// get decodes the member name into v and reports whether it was there and
// decoded.
func (o object) get(name string, v any) bool {
	raw, ok := o[name]
	return ok && json.Unmarshal(raw, v) == nil
}

// Names of the wire format that more than one place spells.
const (
	// typeFunction is the type of a tool and of a tool call.
	typeFunction = "function"
	keyType      = "type"
	keyName      = "name"
)

// body returns the request as the JSON object the protocol takes.
func (r *Request) body() map[string]any {
	msgs := make([]map[string]any, 0, len(r.Messages))
	for i := range r.Messages {
		msgs = append(msgs, r.Messages[i].body())
	}
	body := map[string]any{"model": r.Model, "messages": msgs}
	if len(r.Tools) > 0 {
		tools := make([]map[string]any, 0, len(r.Tools))
		for i := range r.Tools {
			tools = append(tools, r.Tools[i].body())
		}
		body["tools"] = tools
	}
	if len(r.ResponseSchema) > 0 {
		body["response_format"] = map[string]any{
			keyType:       "json_schema",
			"json_schema": map[string]any{keyName: "response", "strict": true, "schema": r.ResponseSchema},
		}
	}
	if r.MaxTokens > 0 {
		body["max_tokens"] = r.MaxTokens
	}
	return body
}

// body returns the message as the JSON object the protocol takes.
func (m *Message) body() map[string]any {
	out := map[string]any{"role": m.Role, "content": m.Content}
	if m.ToolCallID != "" {
		out["tool_call_id"] = m.ToolCallID
	}
	if len(m.ToolCalls) > 0 {
		calls := make([]map[string]any, 0, len(m.ToolCalls))
		for _, tc := range m.ToolCalls {
			calls = append(calls, map[string]any{
				"id": tc.ID, keyType: typeFunction, "function": map[string]any{keyName: tc.Name, "arguments": tc.Arguments},
			})
		}
		out["tool_calls"] = calls
	}
	return out
}

// body returns the tool as the JSON object the protocol takes.
func (t *Tool) body() map[string]any {
	fn := map[string]any{keyName: t.Name}
	if t.Description != "" {
		fn["description"] = t.Description
	}
	if len(t.Parameters) > 0 {
		fn["parameters"] = t.Parameters
	}
	return map[string]any{keyType: typeFunction, "function": fn}
}

// Complete sends req and returns the first choice.
func (c *Client) Complete(ctx context.Context, req *Request) (*Reply, error) {
	raw, err := c.post(ctx, req)
	if err != nil {
		return nil, err
	}
	return parseReply(raw)
}

// openRouterHost is the host of OpenRouter's API, the one OpenAI-compatible
// server that takes routing preferences in a request's `provider` object.
const openRouterHost = "openrouter.ai"

// RequestBody returns the JSON object req is sent as. On OpenRouter a
// request that sends tools also asks `provider.require_parameters`, so the
// router only picks a host that supports every parameter of the request,
// tools included: without it a host that silently drops them can answer
// with no tool call, which the Ask box would read as an answer from data it
// never read. Host pinning (`order`, `allow_fallbacks`) is not set here.
func (c *Client) RequestBody(req *Request) map[string]any {
	body := req.body()
	if len(req.Tools) > 0 && c.onOpenRouter() {
		body["provider"] = map[string]any{"require_parameters": true}
	}
	return body
}

// onOpenRouter reports whether the client's base URL is OpenRouter's: its
// host is openrouter.ai or one of its subdomains. Other servers get no
// `provider` object: OpenAI's own API refuses a request member it does not
// know, so the member is sent only where it is part of the protocol.
func (c *Client) onOpenRouter() bool {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return host == openRouterHost || strings.HasSuffix(host, "."+openRouterHost)
}

// post sends req to the provider and returns the body of its 2xx answer.
func (c *Client) post(ctx context.Context, req *Request) ([]byte, error) {
	body, err := json.Marshal(c.RequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("%w: encoding", ErrRejected)
	}
	hr, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: building the request", ErrUnavailable)
	}
	hr.Header.Set("Content-Type", "application/json")
	hr.Header.Set("Accept", "application/json")
	key, err := c.Key()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if key != "" {
		hr.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := c.HTTP.Do(hr)
	if err != nil {
		// A transport error can quote the URL, never the body; the key and
		// the request are not in it.
		return nil, fmt.Errorf("%w: no answer from the provider", ErrUnavailable)
	}
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, maxReplyBytes+1))
	closeErr := resp.Body.Close()
	if statusErr := statusError(resp.StatusCode); statusErr != nil {
		return nil, statusErr
	}
	if readErr != nil || closeErr != nil || len(raw) > maxReplyBytes {
		return nil, fmt.Errorf("%w: the reply could not be read", ErrUnavailable)
	}
	return raw, nil
}

// parseReply reads the first choice and the usage of a chat completion.
func parseReply(raw []byte) (*Reply, error) {
	var top object
	var choices []object
	if json.Unmarshal(raw, &top) != nil || !top.get("choices", &choices) || len(choices) == 0 {
		return nil, ErrBadReply
	}
	var out Reply
	var message object
	if !choices[0].get("message", &message) {
		return nil, ErrBadReply
	}
	var content *string
	if message.get("content", &content) && content != nil {
		out.Content = *content
	}
	var calls []object
	message.get("tool_calls", &calls)
	for _, c := range calls {
		var fn object
		var tc ToolCall
		c.get("id", &tc.ID)
		if c.get("function", &fn) {
			fn.get("name", &tc.Name)
			fn.get("arguments", &tc.Arguments)
		}
		out.ToolCalls = append(out.ToolCalls, tc)
	}
	choices[0].get("finish_reason", &out.FinishReason)
	var usage object
	if top.get("usage", &usage) {
		usage.get("prompt_tokens", &out.InputTokens)
		usage.get("completion_tokens", &out.OutputTokens)
	}
	return &out, nil
}

// statusError classifies an HTTP status: a 2xx is nil, a refusal of the
// request itself is ErrRejected, and anything else, a refused key, a rate
// limit or a failing provider, is ErrUnavailable so the router tries the
// route's next model.
func statusError(code int) error {
	switch {
	case code >= 200 && code < 300:
		return nil
	case code == http.StatusBadRequest, code == http.StatusNotFound, code == http.StatusRequestEntityTooLarge,
		code == http.StatusUnprocessableEntity:
		return fmt.Errorf("%w: HTTP %d", ErrRejected, code)
	}
	return fmt.Errorf("%w: HTTP %d", ErrUnavailable, code)
}
