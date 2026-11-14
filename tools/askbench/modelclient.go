package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/ui"
	"github.com/arkadiuss/openfief/plugins/model-openai/openaichat"
)

// defaultBaseURL is OpenRouter's API root, where D-207's two candidates run.
const defaultBaseURL = "https://openrouter.ai/api/v1"

// ErrBudget reports that the run's budget is spent: no more model calls are
// made.
var ErrBudget = errors.New("the budget is spent")

// ErrModels reports a model list the bench can't use.
var ErrModels = errors.New("the bench's models")

// Model is a model the bench asks and what it costs: the id on the provider
// and its prices in US dollars per million tokens, as a model entry of the
// model-openai plugin holds them (D-205).
type Model struct {
	ID          string
	InputPrice  kernel.Decimal
	OutputPrice kernel.Decimal
	// Route, when not empty, pins the OpenRouter hosts the model is asked on, by
	// their provider slugs, in this order and with no fallback to others.
	Route []string
	// Structured is whether the model's entry has the structured-output flag
	// (D-205): the structured answer is then asked of it strictly (json_schema),
	// else as a JSON object (JSON mode) that the loop checks and retries once.
	Structured bool
}

// modelLabel names a run of the model: its id, and, when it is pinned to hosts, the
// id with them (google/gemma-4-26b-a4b-it@primeintellect), so one model on two
// hosts is two entries of the report.
func modelLabel(m *Model) string {
	if len(m.Route) == 0 {
		return m.ID
	}
	return m.ID + "@" + strings.Join(m.Route, "|")
}

// candidateModels are D-207's two candidates for the demo's model, with the
// prices OpenRouter lists for them (R-30): Qwen3.7 Flash and gpt-oss-120b. The
// second is priced at the plugin's documented entry, the dearer hosts', so a
// budget over it holds on any host.
func candidateModels() []Model {
	return []Model{
		priced("qwen/qwen3.7-flash", "0.03", "0.13"),
		{ID: "openai/gpt-oss-120b", InputPrice: priced("", "0.15", "0").InputPrice, OutputPrice: priced("", "0", "0.60").OutputPrice, Structured: true},
	}
}

// structuredModels are the models whose entries have the structured-output flag
// by default: gpt-oss and gemma-4 do on most hosts. -structured and -json-mode
// set it for others, or take it away.
func structuredModels() []string {
	return []string{"openai/gpt-oss-120b", "google/gemma-4-26b-a4b-it", "google/gemma-4-31b-it"}
}

// priced returns the model with the prices, decimal texts of constants: a price
// that does not read is zero, which the budget would never stop.
func priced(id, input, output string) Model {
	in, errIn := kernel.ParseDecimal(input)
	out, errOut := kernel.ParseDecimal(output)
	if errIn != nil || errOut != nil {
		return Model{ID: id}
	}
	return Model{ID: id, InputPrice: in, OutputPrice: out}
}

// ModelCall is one call of the loop to the model.
type ModelCall struct {
	// ToolsOffered is how many tools the call offered; the last round offers none.
	ToolsOffered int
	// MaxTokens is the output limit the call carried.
	MaxTokens                 int64
	InputTokens, OutputTokens int64
	Latency                   time.Duration
	// Error is the class of a failed call, "" for one that answered.
	Error string
	// Status is the HTTP status of a failed call, 0 when none came.
	Status int
	// Label is the short fixed label of a failed call's status or timeout:
	// never any text of the provider's reply.
	Label string
}

// meter counts what a run spends, across the runs that share it, and says when
// the budget is spent. Its methods are safe for concurrent use.
type meter struct {
	mu       sync.Mutex
	maxCost  kernel.Decimal
	maxCalls int
	calls    int
	inTok    int64
	outTok   int64
	cost     kernel.Decimal
}

// newMeter returns a meter with the budget: a most to spend and a most of
// calls to make.
func newMeter(maxCost kernel.Decimal, maxCalls int) *meter {
	return &meter{maxCost: maxCost, maxCalls: maxCalls}
}

// allow reserves a call, or fails with ErrBudget when the budget is spent.
func (m *meter) allow() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.calls >= m.maxCalls || m.cost.Cmp(m.maxCost) >= 0 {
		return ErrBudget
	}
	m.calls++
	return nil
}

// charge adds the tokens of an answered call, priced by model.
func (m *meter) charge(model *Model, in, out int64) kernel.Decimal {
	spent := costOf(model, in, out)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inTok += in
	m.outTok += out
	m.cost = m.cost.Add(spent)
	return spent
}

// spent is what a meter has counted.
type spent struct {
	calls  int
	in     int64
	out    int64
	amount kernel.Decimal
}

// snapshot returns what has been spent.
func (m *meter) snapshot() spent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return spent{calls: m.calls, in: m.inTok, out: m.outTok, amount: m.cost}
}

// costOf prices tokens: each count times its price per million tokens, exact.
func costOf(model *Model, in, out int64) kernel.Decimal {
	perToken := kernel.NewDecimal(1, -6)
	return kernel.NewDecimal(in, 0).Mul(model.InputPrice).Mul(perToken).Add(kernel.NewDecimal(out, 0).Mul(model.OutputPrice).Mul(perToken))
}

// modelRun is the model of one question, ui.AskModel over the plugin's own
// chat client (openaichat), so a request carries what the plugin sends:
// OpenRouter's require_parameters with tools, the output limit as max_tokens.
// It records each call. One question, one modelRun: not for concurrent use.
type modelRun struct {
	client *openaichat.Client
	model  *Model
	meter  *meter
	now    Clock
	calls  []ModelCall
	// budgetHit is set when a call was refused because the budget is spent.
	budgetHit bool
	// cost is what this question's calls came to.
	cost kernel.Decimal
}

var _ ui.AskModel = (*modelRun)(nil)

// Generate sends the conversation to the model and returns its next message. A
// call with a schema asks for the structured answer, as the plugin does: strictly
// when the model's entry has the flag, else as a JSON object.
func (r *modelRun) Generate(ctx context.Context, msgs []ui.AskMessage, tools []ui.AskTool, maxTokens int64, schema string) (ui.AskReply, error) {
	call := ModelCall{ToolsOffered: len(tools), MaxTokens: maxTokens}
	if err := r.meter.allow(); err != nil {
		r.budgetHit = true
		return ui.AskReply{}, err
	}
	started := r.now()
	req := requestOf(r.model.ID, msgs, tools, maxTokens)
	if schema != "" {
		req.SetAnswerFormat(json.RawMessage(schema), r.model.Structured)
	}
	reply, err := r.client.Complete(ctx, req)
	call.Latency = r.now().Sub(started)
	if err != nil {
		call.Error = errorClass(err)
		failure := callFailure(err)
		call.Status, call.Label = failure.Status, failure.Label
		r.calls = append(r.calls, call)
		return ui.AskReply{}, fmt.Errorf("the model: %s: %w", call.Error, err)
	}
	call.InputTokens, call.OutputTokens = reply.InputTokens, reply.OutputTokens
	r.cost = r.cost.Add(r.meter.charge(r.model, reply.InputTokens, reply.OutputTokens))
	r.calls = append(r.calls, call)
	out := ui.AskReply{Content: reply.Content, Model: r.model.ID, OutputTokens: reply.OutputTokens}
	for _, tc := range reply.ToolCalls {
		out.Calls = append(out.Calls, ui.AskCall{ID: tc.ID, Name: tc.Name, Args: tc.Arguments})
	}
	return out, nil
}

// The labels of a failed call, a fixed set: a status and one of these say what
// went wrong, and nothing the provider wrote.
const (
	labelNoCredits     = "no credits"
	labelKeyRefused    = "key refused"
	labelRateLimited   = "rate limited"
	labelProviderError = "provider error"
	labelRejected      = "request rejected"
	labelNoAnswer      = "no answer"
	labelFailed        = "failed"
)

// statusLabel returns the fixed label of an HTTP status, "" when it has none.
func statusLabel(status int) string {
	switch {
	case status == http.StatusPaymentRequired:
		return labelNoCredits
	case status == http.StatusUnauthorized:
		return labelKeyRefused
	case status == http.StatusTooManyRequests:
		return labelRateLimited
	case status >= http.StatusInternalServerError && status < statusCeiling:
		return labelProviderError
	case status >= http.StatusBadRequest && status < http.StatusInternalServerError:
		return labelRejected
	}
	return ""
}

// statusCeiling is the first number that is no HTTP status.
const statusCeiling = 600

// callFailure returns the HTTP status of a failed call (0 when none came) and
// its label, read from the chat client's CallError.
func callFailure(err error) CallErrorReport {
	var ce *openaichat.CallError
	found := errors.As(err, &ce)
	switch {
	case found && ce.Status != 0:
		if l := statusLabel(ce.Status); l != "" {
			return CallErrorReport{Status: ce.Status, Label: l}
		}
		return CallErrorReport{Status: ce.Status, Label: labelFailed}
	case found && ce.Timeout, errors.Is(err, context.DeadlineExceeded):
		return CallErrorReport{Label: labelTimeout}
	case found:
		return CallErrorReport{Label: labelNoAnswer}
	}
	return CallErrorReport{Label: labelFailed}
}

// errorClass names why a call failed, never quoting the provider or the key.
func errorClass(err error) string {
	switch {
	case errors.Is(err, openaichat.ErrRejected):
		return "rejected"
	case errors.Is(err, openaichat.ErrBadReply):
		return "bad reply"
	case errors.Is(err, openaichat.ErrUnavailable):
		return "unavailable"
	case errors.Is(err, context.DeadlineExceeded):
		return labelTimeout
	}
	return "failed"
}

// requestOf translates the loop's conversation into the chat client's request,
// as the plugin does with the core's: the same roles, the tools' schemas as
// parameters, and the output limit as max_tokens.
func requestOf(model string, msgs []ui.AskMessage, tools []ui.AskTool, maxTokens int64) *openaichat.Request {
	req := &openaichat.Request{Model: model, MaxTokens: maxTokens}
	for i := range msgs {
		m := &msgs[i]
		wm := openaichat.Message{Role: m.Role, Content: m.Content, ToolCallID: m.CallID}
		for _, c := range m.Calls {
			wm.ToolCalls = append(wm.ToolCalls, openaichat.ToolCall{ID: c.ID, Name: c.Name, Arguments: c.Args})
		}
		req.Messages = append(req.Messages, wm)
	}
	for i := range tools {
		t := &tools[i]
		req.Tools = append(req.Tools, openaichat.Tool{Name: t.Name, Description: t.Description, Parameters: json.RawMessage(t.Schema)})
	}
	return req
}

// newChatClient returns the plugin's chat client for baseURL with key. The key
// is read per call and never appears in an error (the client's errors name a
// status, not a request).
func newChatClient(hc *http.Client, baseURL, key string) *openaichat.Client {
	return &openaichat.Client{HTTP: hc, BaseURL: strings.TrimRight(baseURL, "/"), Key: func() (string, error) { return key, nil }}
}

// routedClient is newChatClient pinned to the hosts of route.
func routedClient(hc *http.Client, baseURL, key string, route []string) *openaichat.Client {
	c := newChatClient(hc, baseURL, key)
	c.Route = route
	return c
}

// routingProbe returns the request body a request with one tool gets at
// baseURL, pinned to route: it holds the provider routing member (OpenRouter's
// require_parameters, and the order and no fallback of a route) the plugin's
// client adds for the host, which the config hash covers.
func routingProbe(baseURL string, route []string) map[string]any {
	c := routedClient(nil, baseURL, "", route)
	body := c.RequestBody(&openaichat.Request{Model: "m", Tools: []openaichat.Tool{{Name: "t"}}})
	if provider, ok := body["provider"]; ok {
		return map[string]any{"provider": provider}
	}
	return map[string]any{}
}
