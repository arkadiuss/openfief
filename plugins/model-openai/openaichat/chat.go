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
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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

// CallError is a failed call's error as a program reads it: errors.As finds
// it in any error of Complete that came from the provider's HTTP status or from
// no answer. Its text and its sentinel (ErrRejected, ErrUnavailable) are those
// of the error it wraps, so what a user is shown does not change. It carries
// the status code and nothing of the provider's reply or the request.
type CallError struct {
	// Status is the HTTP status the provider answered, 0 when it gave none.
	Status int
	// Timeout is true when the call ran out of time without an answer.
	Timeout bool
	err     error
	// retryAfter is the answer's Retry-After header, for the retry of a 429.
	retryAfter string
}

// Error is the text of the wrapped error.
func (e *CallError) Error() string { return e.err.Error() }

// Unwrap returns the wrapped error, which wraps the sentinel.
func (e *CallError) Unwrap() error { return e.err }

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
	// Route, when not empty, pins the hosts OpenRouter may use, in this order
	// (its provider slugs), with no fallback to others: `provider.order`,
	// `allow_fallbacks` false and `require_parameters`. Nothing sets it in the
	// plugin; the Ask box bench does, to compare a model on one host.
	Route []string
	// Reasoning, when not empty, sets OpenRouter's `reasoning` request
	// parameter: ReasoningOff sends {"enabled": false}, and ReasoningLow,
	// ReasoningMedium or ReasoningHigh send {"effort": ...}. Empty sends
	// nothing, the model's own default. Nothing sets it in the plugin; the Ask
	// box bench does, to compare a model with its thinking off or turned down.
	Reasoning string
	// DataCollection, when not empty, sets OpenRouter's
	// `provider.data_collection`: DataCollectionDeny lets no host serve a
	// request that may store it or train on it, DataCollectionAllow refuses
	// none. It joins whatever else the request's `provider` object holds
	// (Route, require_parameters).
	DataCollection string
	// IncludeCost asks OpenRouter for the call's exact cost: `usage` with
	// `include` true, which puts usage.cost (USD) and the serving `provider`
	// in the reply. Nothing sets it in the plugin; the bench does.
	IncludeCost bool
	// Wait pauses for d, or until ctx ends, between a 429 and its one retry;
	// nil waits on a timer. Now is the clock Retry-After's HTTP date is read
	// against; nil is time.Now. Tests set both so that none sleeps.
	Wait func(ctx context.Context, d time.Duration) error
	Now  func() time.Time
}

// The wait before the one retry of a 429 (D-207): the provider's Retry-After
// capped at maxRetryAfter, or defaultRetryWait when it names none.
const (
	maxRetryAfter    = 10 * time.Second
	defaultRetryWait = time.Second
)

// The values of Client.Reasoning.
const (
	ReasoningOff    = "off"
	ReasoningLow    = "low"
	ReasoningMedium = "medium"
	ReasoningHigh   = "high"
)

// The values of Client.DataCollection.
const (
	DataCollectionAllow = "allow"
	DataCollectionDeny  = "deny"
)

// ValidDataCollection reports whether v is a value of Client.DataCollection.
func ValidDataCollection(v string) bool { return v == DataCollectionAllow || v == DataCollectionDeny }

// ValidReasoning reports whether v is a value of Client.Reasoning.
func ValidReasoning(v string) bool {
	switch v {
	case ReasoningOff, ReasoningLow, ReasoningMedium, ReasoningHigh:
		return true
	}
	return false
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
	// ResponseSchema, when set, asks for a JSON answer of this JSON Schema,
	// strictly; JSONMode asks for a JSON object only (response_format
	// json_object), for a model that has no structured output. SetAnswerFormat
	// sets one or the other.
	ResponseSchema json.RawMessage
	JSONMode       bool
	MaxTokens      int64
}

// SetAnswerFormat asks for a JSON answer of schema: strictly (json_schema) when
// the model takes structured output, else as a JSON object (JSON mode) that the
// caller checks against the schema.
func (r *Request) SetAnswerFormat(schema json.RawMessage, structured bool) {
	r.ResponseSchema, r.JSONMode = nil, true
	if structured {
		r.ResponseSchema, r.JSONMode = schema, false
	}
}

// Reply is the part of a completion the plugin reads.
type Reply struct {
	Content      string
	ToolCalls    []ToolCall
	FinishReason string
	InputTokens  int64
	OutputTokens int64
	// CachedTokens is how many of the input tokens the host served from its
	// prompt cache, as usage.prompt_tokens_details.cached_tokens (or the
	// usage's own cached_tokens) reports them; 0 when the reply has none.
	CachedTokens int64
	// Cost is the call's cost in USD as the provider reported it (OpenRouter's
	// usage.cost), as the number's own text; empty when the reply has none.
	Cost string
	// Provider is the host that served the call (OpenRouter's `provider`),
	// empty when the reply names none.
	Provider string
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
	switch {
	case len(r.ResponseSchema) > 0:
		body["response_format"] = map[string]any{
			keyType:       "json_schema",
			"json_schema": map[string]any{keyName: "response", "strict": true, "schema": r.ResponseSchema},
		}
	case r.JSONMode:
		body["response_format"] = map[string]any{keyType: "json_object"}
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
// request that sends tools or a response format also asks
// `provider.require_parameters`, so the router only picks a host that supports
// every parameter of the request, tools and format included: without it a host
// that silently drops them can answer with no tool call, which the Ask box
// would read as an answer from data it never read, or with text that is not the
// structured answer it asked for. Host pinning (`order`, `allow_fallbacks`) is
// set only when the client has a Route, which pins the hosts for every request.
func (c *Client) RequestBody(req *Request) map[string]any {
	body := req.body()
	if c.OnOpenRouter() {
		if provider := c.providerPreferences(req); len(provider) > 0 {
			body["provider"] = provider
		}
		c.addOpenRouterOptions(body)
	}
	return body
}

// OnOpenRouter reports whether the client's base URL is OpenRouter's: its
// host is openrouter.ai or one of its subdomains. Other servers get no
// `provider`, `reasoning` or `usage` member: OpenAI's own API refuses a request
// member it does not know, so each is sent only where it is part of the
// protocol.
func (c *Client) OnOpenRouter() bool {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return host == openRouterHost || strings.HasSuffix(host, "."+openRouterHost)
}

// providerPreferences returns OpenRouter's `provider` object for req, one
// object for every preference: the pinned hosts with no fallback and
// require_parameters when the client has a Route, require_parameters alone for
// a request that sends tools or a response format, and data_collection when
// the client sets it, whatever else the request holds.
func (c *Client) providerPreferences(req *Request) map[string]any {
	provider := map[string]any{}
	switch {
	case len(c.Route) > 0:
		provider["order"], provider["allow_fallbacks"], provider["require_parameters"] = c.Route, false, true
	case len(req.Tools) > 0 || len(req.ResponseSchema) > 0 || req.JSONMode:
		provider["require_parameters"] = true
	}
	if ValidDataCollection(c.DataCollection) {
		provider["data_collection"] = c.DataCollection
	}
	return provider
}

// addOpenRouterOptions adds the reasoning and usage members the client's
// options ask for. Both are OpenRouter's: other servers may refuse them.
func (c *Client) addOpenRouterOptions(body map[string]any) {
	switch {
	case c.Reasoning == ReasoningOff:
		body["reasoning"] = map[string]any{"enabled": false}
	case ValidReasoning(c.Reasoning):
		body["reasoning"] = map[string]any{"effort": c.Reasoning}
	}
	if c.IncludeCost {
		body["usage"] = map[string]any{"include": true}
	}
}

// post sends req to the provider and returns the body of its 2xx answer. A 429
// is retried once, after the wait retryWait gives, when that wait fits in the
// time the call has left (D-207): the question's deadline stays the limit. Any
// other status, a 4xx included, is never retried here.
func (c *Client) post(ctx context.Context, req *Request) ([]byte, error) {
	body, err := json.Marshal(c.RequestBody(req))
	if err != nil {
		return nil, fmt.Errorf("%w: encoding", ErrRejected)
	}
	raw, err := c.send(ctx, body)
	var ce *CallError
	if err == nil || !errors.As(err, &ce) || ce.Status != http.StatusTooManyRequests {
		return raw, err
	}
	wait := c.retryWait(ce.retryAfter)
	if deadline, ok := ctx.Deadline(); ok && c.now().Add(wait).After(deadline) {
		return nil, err
	}
	if c.pause(ctx, wait) != nil {
		return nil, err
	}
	return c.send(ctx, body)
}

// retryWait is how long to wait before retrying a 429 whose Retry-After header
// is retryAfter: the seconds or the HTTP date it holds, never past maxRetryAfter
// and never negative; defaultRetryWait when it is empty or neither form.
func (c *Client) retryWait(retryAfter string) time.Duration {
	retryAfter = strings.TrimSpace(retryAfter)
	wait := defaultRetryWait
	if secs, err := strconv.ParseUint(retryAfter, 10, 31); err == nil {
		wait = time.Duration(secs) * time.Second
	} else if at, err := http.ParseTime(retryAfter); err == nil {
		wait = max(at.Sub(c.now()), 0)
	}
	return min(wait, maxRetryAfter)
}

// now is the client's clock.
func (c *Client) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// pause waits d, or until ctx ends, and returns why it stopped early.
func (c *Client) pause(ctx context.Context, d time.Duration) error {
	if c.Wait != nil {
		return c.Wait(ctx, d)
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("waiting to retry: %w", ctx.Err())
	}
}

// send makes one POST of body and returns the body of its 2xx answer. The error
// of a non-2xx answer keeps the answer's Retry-After header.
func (c *Client) send(ctx context.Context, body []byte) ([]byte, error) {
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
		return nil, noAnswer(err)
	}
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, maxReplyBytes+1))
	closeErr := resp.Body.Close()
	if statusErr := statusError(resp.StatusCode); statusErr != nil {
		if ce, ok := errors.AsType[*CallError](statusErr); ok {
			ce.retryAfter = resp.Header.Get("Retry-After")
		}
		return nil, statusErr
	}
	if readErr != nil || closeErr != nil || len(raw) > maxReplyBytes {
		return nil, fmt.Errorf("%w: the reply could not be read", ErrUnavailable)
	}
	return raw, nil
}

// noAnswer is the error of a call that got no HTTP answer, from the transport's
// error cause, which is not quoted: it can name the URL. It says only whether
// the call ran out of time.
func noAnswer(cause error) error {
	var netErr net.Error
	timeout := errors.Is(cause, context.DeadlineExceeded) || (errors.As(cause, &netErr) && netErr.Timeout())
	return &CallError{Timeout: timeout, err: fmt.Errorf("%w: no answer from the provider", ErrUnavailable)}
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
		out.Cost = usage.costText()
		out.CachedTokens = usage.cachedTokens()
	}
	top.get("provider", &out.Provider)
	return &out, nil
}

// cachedTokens returns how many input tokens the host served from its cache:
// OpenAI's and OpenRouter's prompt_tokens_details.cached_tokens, else a
// top-level cached_tokens; 0 when the usage has neither.
func (o object) cachedTokens() int64 {
	var details object
	var n int64
	if o.get("prompt_tokens_details", &details) && details.get("cached_tokens", &n) {
		return n
	}
	if o.get("cached_tokens", &n) {
		return n
	}
	return 0
}

// costText returns the usage's cost as the number's own text, "" when it has none.
func (o object) costText() string {
	var cost json.Number
	if !o.get("cost", &cost) {
		return ""
	}
	return cost.String()
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
		return &CallError{Status: code, err: fmt.Errorf("%w: HTTP %d", ErrRejected, code)}
	}
	return &CallError{Status: code, err: fmt.Errorf("%w: HTTP %d", ErrUnavailable, code)}
}
