package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"slices"
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
	// Reasoning, when not empty, is OpenRouter's reasoning control for the
	// model (openaichat.ReasoningOff, Low, Medium or High); empty sends none.
	Reasoning string
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
	// Cost is what the call cost in USD: the provider's own figure when the
	// reply held one (Measured), else the tokens times the model entry's prices.
	Cost     kernel.Decimal
	Measured bool
	// Provider is the host that served the call, "" when the reply named none.
	Provider string
	// Transport is true when the call failed through the host's fault (D-207):
	// no HTTP answer, a timeout, a 5xx, a 429 that its retry did not mend, a
	// reply that is no chat completion to a request the bench checked, or a
	// degenerate generation. Such a call is retried once, a 429 by the chat
	// client before it gets here.
	Transport bool
	// Fault is whose fault a failed call is: faultHost (it counts in the
	// reliability), faultBench (the bench sent a request that is ill-formed, or
	// failed locally: the run is void), or "" (a refused key, no credits or
	// another 4xx, listed as an error).
	Fault string
	// Retry is true for the second attempt of a call that failed at the
	// transport level; Retried is true for the failed first attempt, and
	// Recovered for the one whose retry answered.
	Retry, Retried, Recovered bool
	// Round is what the call sent and got back, for the report's trace.
	Round RoundReport
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

// charge adds the tokens and the cost of an answered call.
func (m *meter) charge(in, out int64, cost kernel.Decimal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inTok += in
	m.outTok += out
	m.cost = m.cost.Add(cost)
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

// measuredCost reads the cost a provider reported, the text of a JSON number
// such as 0.0000123 or 1.5e-05, exactly; it reports false for no figure, one that
// is no number, or a negative one.
func measuredCost(text string) (kernel.Decimal, bool) {
	r, parsed := new(big.Rat).SetString(text)
	if text == "" || !parsed || r.Sign() < 0 {
		return kernel.Decimal{}, false
	}
	const places = 18
	plain := strings.TrimRight(strings.TrimRight(r.FloatString(places), "0"), ".")
	d, err := kernel.ParseDecimal(plain)
	return d, err == nil
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
	req := requestOf(r.model.ID, msgs, tools, maxTokens)
	if schema != "" {
		req.SetAnswerFormat(json.RawMessage(schema), r.model.Structured)
	}
	out, err := r.attempt(ctx, req, len(tools), false)
	if err == nil || !r.retryable(ctx) {
		return out, err
	}
	// A transport failure is retried once, and the retry is a call of its own:
	// it spends the budget and its cost. An answer of any kind, and a 4xx, never
	// are retried; a 429 was already retried once inside the chat client, after
	// Retry-After, so one that is still a 429 is a transport failure with no
	// retry of the bench's own.
	first := len(r.calls) - 1
	retried, retryErr := r.attempt(ctx, req, len(tools), true)
	if errors.Is(retryErr, ErrBudget) {
		return out, err
	}
	r.calls[first].Retried, r.calls[first].Recovered = true, retryErr == nil
	return retried, retryErr
}

// retryable reports whether the call that just failed is worth one more: it
// failed at the transport level, and the question still has time. A 429 is not:
// the chat client retried it once already.
func (r *modelRun) retryable(ctx context.Context) bool {
	if len(r.calls) == 0 {
		return false
	}
	last := &r.calls[len(r.calls)-1]
	return last.Transport && last.Status != http.StatusTooManyRequests && ctx.Err() == nil
}

// attempt makes one call to the model and records it. It fails with ErrBudget,
// recording nothing, when the budget is spent.
func (r *modelRun) attempt(ctx context.Context, req *openaichat.Request, toolsOffered int, retry bool) (ui.AskReply, error) {
	call := ModelCall{ToolsOffered: toolsOffered, MaxTokens: req.MaxTokens, Retry: retry, Round: sentRound(req, len(r.calls)+1, retry)}
	if err := r.meter.allow(); err != nil {
		r.budgetHit = true
		return ui.AskReply{}, err
	}
	if err := validateRequest(req); err != nil {
		// Nothing is sent: a request that is ill-formed says nothing of the host.
		return ui.AskReply{}, r.fail(&call, fmt.Errorf("%w: %w", errInvalidRequest, err))
	}
	started := r.now()
	reply, err := r.client.Complete(ctx, req)
	call.Latency = r.now().Sub(started)
	if err != nil {
		return ui.AskReply{}, r.fail(&call, err)
	}
	if degenerate(reply) {
		call.InputTokens, call.OutputTokens, call.Provider = reply.InputTokens, reply.OutputTokens, reply.Provider
		call.Round.fillReply(reply)
		call.Cost, call.Measured = measuredCost(reply.Cost)
		if !call.Measured {
			call.Cost = costOf(r.model, reply.InputTokens, reply.OutputTokens)
		}
		r.meter.charge(reply.InputTokens, reply.OutputTokens, call.Cost)
		r.cost = r.cost.Add(call.Cost)
		return ui.AskReply{}, r.fail(&call, errDegenerate)
	}
	call.InputTokens, call.OutputTokens, call.Provider = reply.InputTokens, reply.OutputTokens, reply.Provider
	call.Round.fillReply(reply)
	call.Cost, call.Measured = measuredCost(reply.Cost)
	if !call.Measured {
		call.Cost = costOf(r.model, reply.InputTokens, reply.OutputTokens)
	}
	r.meter.charge(reply.InputTokens, reply.OutputTokens, call.Cost)
	r.cost = r.cost.Add(call.Cost)
	r.calls = append(r.calls, call)
	out := ui.AskReply{Content: reply.Content, Model: r.model.ID, OutputTokens: reply.OutputTokens}
	for _, tc := range reply.ToolCalls {
		out.Calls = append(out.Calls, ui.AskCall{ID: tc.ID, Name: tc.Name, Args: tc.Arguments})
	}
	return out, nil
}

// fail records a failed call, with its class, label, status and whose fault it
// is, and returns the error the loop sees.
func (r *modelRun) fail(call *ModelCall, err error) error {
	call.Error = errorClass(err)
	failure := callFailure(err)
	call.Status, call.Label = failure.Status, failure.Label
	call.Fault = faultOf(err)
	call.Transport = call.Fault == faultHost
	r.calls = append(r.calls, *call)
	return fmt.Errorf("the model: %s: %w", call.Error, err)
}

// faultOf says whose fault a failed call is. The host's: no answer, a timeout,
// a 5xx, a 429, a reply that is no chat completion (the request was checked
// before it was sent) and a degenerate generation. The bench's: a request that
// breaks the chat API's rules and an error of its own side of the call. Neither
// for a refused key, no credits and another 4xx.
func faultOf(err error) string {
	switch errorClass(err) {
	case classLocal, classInvalid:
		return faultBench
	case classBadReply, classDegenerate, classTimeout:
		return faultHost
	}
	if isTransportFailure(err) {
		return faultHost
	}
	return ""
}

// harmonyTokens are the control tokens of the harmony format that a host must
// strip: in a reply's content they show a generation that lost its way.
func harmonyTokens() []string {
	return []string{"<|channel|>", "<|message|>", "<|start|>", "<|end|>", "<|return|>", "<|call|>", "<|constrain|>"}
}

// degenerate reports whether a reply is a generation gone wrong and no answer:
// it stopped at the length limit with no text and no tool call, or its content
// holds harmony control tokens.
func degenerate(reply *openaichat.Reply) bool {
	if reply.FinishReason == "length" && strings.TrimSpace(reply.Content) == "" && len(reply.ToolCalls) == 0 {
		return true
	}
	return slices.ContainsFunc(harmonyTokens(), func(tok string) bool { return strings.Contains(reply.Content, tok) })
}

// isTransportFailure reports whether a call failed at the transport level: the
// chat client got no HTTP answer (nothing came, or it timed out), the provider
// answered 5xx, or it answered 429 twice (the client retries a 429 once, after
// Retry-After): a host that rate-limits its callers is not reliable (D-207). Any
// other 4xx, a refused key and a bad reply are the request's or the model's, not
// the host's, and are not.
func isTransportFailure(err error) bool {
	var ce *openaichat.CallError
	if !errors.As(err, &ce) {
		return false
	}
	return ce.Status == 0 || ce.Status == http.StatusTooManyRequests ||
		(ce.Status >= http.StatusInternalServerError && ce.Status < statusCeiling)
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
	labelBadReply      = "bad reply"
	labelDegenerate    = "degenerate reply"
	labelInvalid       = "invalid request"
	labelLocal         = "local error"
	labelFailed        = "failed"
)

// The error classes of a failed call, a fixed set: why it failed, whatever the
// label and the status say. Every failed call has one; none is provider text.
const (
	classRejected    = "rejected"
	classUnavailable = "unavailable"
	classBadReply    = "bad reply"
	classTimeout     = "timeout"
	classLocal       = "local error"
	// classDegenerate is a reply that is a generation gone wrong: it stopped at
	// the length limit with no text, or holds the host's own control tokens.
	classDegenerate = "degenerate reply"
	// classInvalid is a request the bench's own check refused to send.
	classInvalid = "invalid request"
)

// Whose fault a failed call is, a fixed set.
const (
	faultHost  = "host"
	faultBench = "bench"
)

// errInvalidRequest is a request that breaks the chat API's rules; the bench
// never sends one, and the run it belongs to is void.
var errInvalidRequest = errors.New("the request breaks the chat API's rules")

// errDegenerate is a reply that is a generation gone wrong, not an answer.
var errDegenerate = errors.New("the reply is a degenerate generation")

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
	case errors.Is(err, errInvalidRequest):
		return CallErrorReport{Label: labelInvalid}
	case errors.Is(err, errDegenerate):
		return CallErrorReport{Label: labelDegenerate}
	case errors.Is(err, openaichat.ErrBadReply):
		// A 2xx answer that is no chat completion: not a transport failure.
		return CallErrorReport{Label: labelBadReply}
	}
	// An error of the bench's own side of the call: building the request or
	// reading the reply.
	return CallErrorReport{Label: labelLocal}
}

// errorClass names why a call failed, from the fixed set of classes, never
// quoting the provider or the key. Every error has one: an error that is none
// of the chat client's is a local error.
func errorClass(err error) string {
	var ce *openaichat.CallError
	found := errors.As(err, &ce)
	switch {
	case errors.Is(err, errInvalidRequest):
		return classInvalid
	case errors.Is(err, errDegenerate):
		return classDegenerate
	case found && ce.Timeout, errors.Is(err, context.DeadlineExceeded):
		return classTimeout
	case errors.Is(err, openaichat.ErrBadReply):
		return classBadReply
	case found && errors.Is(err, openaichat.ErrRejected):
		return classRejected
	case found && errors.Is(err, openaichat.ErrUnavailable):
		return classUnavailable
	}
	// Neither an answer of the provider nor its absence: the request could not
	// be built or the reply could not be read.
	return classLocal
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

// benchClient is the chat client of a run of model: pinned to its hosts, with
// its reasoning control, and asking OpenRouter for each call's exact cost. The
// last two reach OpenRouter only (openaichat); the plugin sets neither.
func benchClient(hc *http.Client, baseURL, key string, m *Model) *openaichat.Client {
	c := routedClient(hc, baseURL, key, m.Route)
	c.Reasoning, c.IncludeCost = m.Reasoning, true
	return c
}

// routingProbe returns the request body members a request with one tool gets at
// baseURL for model m: the provider routing (OpenRouter's require_parameters,
// and the order and no fallback of a route), the reasoning control and the
// usage accounting the bench's client adds for the host. The config hash
// covers them.
func routingProbe(baseURL string, m *Model) map[string]any {
	c := benchClient(nil, baseURL, "", m)
	body := c.RequestBody(&openaichat.Request{Model: "m", Tools: []openaichat.Tool{{Name: "t"}}})
	out := map[string]any{}
	for _, key := range []string{"provider", "reasoning", "usage"} {
		if v, ok := body[key]; ok {
			out[key] = v
		}
	}
	return out
}
