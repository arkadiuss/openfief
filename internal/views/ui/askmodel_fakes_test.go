package ui

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

// scriptedReply is what the fake model says next: a text, or tool calls.
type scriptedReply struct {
	text  string
	calls []*modelsv1.ToolCall
}

// fakeAskModel is an in-memory AskModelService standing for the core's router
// and its model. It is off until a test turns it on, as the core is with
// nothing saved. It admits only the plugin's token, refuses what the core
// would (a tier its model may not receive), says what it was sent, and
// answers from a script.
type fakeAskModel struct {
	modelsv1connect.UnimplementedAskModelServiceHandler

	mu sync.Mutex
	// on makes the Ask box have a model (the switch, the toggle and a route).
	on bool
	// allowed are the data tiers the model may receive; the zero value
	// allows none.
	allowed map[string]bool
	// script is what the model says, in order; past its end it says "done".
	script []scriptedReply
	// reply, when set, answers every GenerateAsk instead of the script, from
	// the conversation so far.
	reply func(msgs []*modelsv1.AskMessage) scriptedReply
	// model, when set, is the "provider/model" reference that answers; "p/m"
	// otherwise.
	model string
	// failWith, when set, is the error GenerateAsk returns.
	failWith error
	// availableErr, when set, is the error GetAskModel returns.
	availableErr error
	// requests are the conversations GenerateAsk was sent, tiers included.
	requests [][]*modelsv1.AskMessage
	// toolsSeen are the tool names each request offered, and schemas the
	// response schema each carried ("" for none).
	toolsSeen [][]string
	schemas   []string
	// toolDefs are the tools each request offered, with their descriptions and
	// argument schemas.
	toolDefs [][]*modelsv1.Tool
	// sessions are the session tokens each call carried.
	sessions []string
	bearers  []string
	checks   [][]string
	// denyChecks are tiers CheckAskTiers refuses though GenerateAsk took
	// them: a tool above what the route allows, as a fief's own data (TF)
	// would be for a route that may receive T0.
	denyChecks map[string]bool
	// opinions is what GetAskModel says of opinions (D-225): off unless set.
	opinions modelsv1.AskOpinions
	// admission is the decision AskAdmission answers for every question (the
	// zero value allows), resetsAt the time it names, admitErr the error
	// AdmitAskQuestion returns instead, and admitted the client addresses of the
	// questions it was asked about, in order.
	admission modelsv1.AskAdmission
	resetsAt  time.Time
	admitErr  error
	admitted  []string
}

// AdmitAskQuestion lets every question go unless the test says otherwise.
func (f *fakeAskModel) AdmitAskQuestion(_ context.Context, r *connect.Request[modelsv1.AdmitAskQuestionRequest]) (*connect.Response[modelsv1.AdmitAskQuestionResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.admitted = append(f.admitted, r.Msg.GetClientAddress())
	if f.admitErr != nil {
		return nil, f.admitErr
	}
	out := &modelsv1.AdmitAskQuestionResponse{Admission: modelsv1.AskAdmission_ASK_ADMISSION_ALLOWED}
	if f.admission != modelsv1.AskAdmission_ASK_ADMISSION_UNSPECIFIED {
		out.Admission = f.admission
	}
	if !f.resetsAt.IsZero() {
		out.ResetsAt = timestamppb.New(f.resetsAt)
	}
	return connect.NewResponse(out), nil
}

func (f *fakeAskModel) GetAskModel(_ context.Context, r *connect.Request[modelsv1.GetAskModelRequest]) (*connect.Response[modelsv1.GetAskModelResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if f.availableErr != nil {
		return nil, f.availableErr
	}
	// As the core decides: opinions are the owner's own session's alone (D-225).
	opinions := f.opinions
	if r.Msg.GetSessionToken() != ownerSession {
		opinions = modelsv1.AskOpinions_ASK_OPINIONS_UNSPECIFIED
	}
	return connect.NewResponse(&modelsv1.GetAskModelResponse{Available: f.on, Opinions: opinions}), nil
}

func (f *fakeAskModel) CheckAskTiers(_ context.Context, r *connect.Request[modelsv1.CheckAskTiersRequest]) (*connect.Response[modelsv1.CheckAskTiersResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.checks = append(f.checks, slices.Clone(r.Msg.GetTiers()))
	allowed := f.on && f.mayReceive(r.Msg.GetTiers())
	for _, t := range r.Msg.GetTiers() {
		allowed = allowed && !f.denyChecks[t]
	}
	return connect.NewResponse(&modelsv1.CheckAskTiersResponse{Allowed: allowed}), nil
}

func (f *fakeAskModel) GenerateAsk(_ context.Context, r *connect.Request[modelsv1.GenerateAskRequest]) (*connect.Response[modelsv1.GenerateAskResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if f.failWith != nil {
		return nil, f.failWith
	}
	if !f.on {
		return nil, codeErr(connect.CodeFailedPrecondition, nil)
	}
	var tiers []string
	for _, m := range r.Msg.GetMessages() {
		tiers = append(tiers, m.GetDataTier())
	}
	if !f.mayReceive(tiers) {
		return nil, codeErr(connect.CodePermissionDenied, nil)
	}
	f.requests = append(f.requests, r.Msg.GetMessages())
	var names []string
	for _, t := range r.Msg.GetTools() {
		names = append(names, t.GetName())
	}
	f.toolsSeen = append(f.toolsSeen, names)
	f.toolDefs = append(f.toolDefs, r.Msg.GetTools())
	f.schemas = append(f.schemas, r.Msg.GetResponseSchemaJson())
	next := scriptedReply{text: answerText("done")}
	switch {
	case f.reply != nil:
		next = f.reply(r.Msg.GetMessages())
	case len(f.script) > 0:
		next, f.script = f.script[0], f.script[1:]
	}
	resp := &modelsv1.GenerateAskResponse{Content: next.text, ToolCalls: next.calls, Model: f.modelRef(), FinishReason: modelsv1.FinishReason_FINISH_REASON_STOP}
	if len(next.calls) > 0 {
		resp.FinishReason = modelsv1.FinishReason_FINISH_REASON_TOOL_CALLS
	}
	return connect.NewResponse(resp), nil
}

// modelRef is the reference of the model that answers.
func (f *fakeAskModel) modelRef() string {
	if f.model != "" {
		return f.model
	}
	return "p/m"
}

func (f *fakeAskModel) turnOn(tiers ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.on, f.allowed = true, map[string]bool{}
	for _, t := range tiers {
		f.allowed[t] = true
	}
}

func (f *fakeAskModel) say(rs ...scriptedReply) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.script = append(f.script, rs...)
}

// generated is how many times the model was asked.
func (f *fakeAskModel) generated() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.requests)
}

func (f *fakeAskModel) admit(h http.Header, session string) error {
	f.bearers = append(f.bearers, h.Get("Authorization"))
	f.sessions = append(f.sessions, session)
	if h.Get("Authorization") != "Bearer "+pluginToken {
		return codeErr(connect.CodePermissionDenied, nil)
	}
	if _, err := scopeOf(session); err != nil {
		return err
	}
	return nil
}

func (f *fakeAskModel) mayReceive(tiers []string) bool {
	for _, t := range tiers {
		if t != tierInstructions && !f.allowed[t] {
			return false
		}
	}
	return true
}

// browserSetUpAskModel makes the fake model what the browser checks ask
// about (D-203): it takes the tiers listed in tiers, comma separated, and is
// gone for "off" or none. A conversation with no tool result gets a call of
// the allocation tool; one with a result, or a refusal, gets the text a model
// would write from it.
func browserSetUpAskModel(f *fakeAskModel, tiers string, opinions bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.on, f.allowed = false, map[string]bool{}
	f.reply = nil
	f.opinions = modelsv1.AskOpinions_ASK_OPINIONS_UNSPECIFIED
	if opinions {
		// With opinions on (D-225), the owner's own session gets them; a question about a thesis is answered with one.
		f.opinions = modelsv1.AskOpinions_ASK_OPINIONS_ON
	}
	if tiers == "" || tiers == "off" {
		return
	}
	f.on = true
	for t := range strings.SplitSeq(tiers, ",") {
		f.allowed[t] = true
	}
	f.reply = browserReply
}

// browserReply is the fake model's answer to a conversation of the browser checks:
// one that asks for a change is refused, with commands to ask instead; one that
// asks for a thesis gets one; one that names the fund is asked about, with the
// choices; any other reads the allocation.
func browserReply(msgs []*modelsv1.AskMessage) scriptedReply {
	switch question := firstUser(msgs); {
	case strings.Contains(question, "sell"):
		return scriptedReply{text: answerJSON(answerKindRefuse, "I can only read, so I can't sell anything.", nil, nil, nil)}
	case strings.Contains(question, "thesis"):
		return thesisFundReply(msgs)
	case strings.Contains(question, "fund"):
		return fundReply(msgs)
	}
	for _, m := range msgs {
		if m.GetMessage().GetRole() != modelsv1.MessageRole_MESSAGE_ROLE_TOOL {
			continue
		}
		if strings.Contains(m.GetMessage().GetContent(), "withheld") {
			return scriptedReply{text: answerText("I may not read your allocation, so I can't say.")}
		}
		return scriptedReply{text: answerText("Most of it is listed.\nThe rest is cash.", "c1")}
	}
	return scriptedReply{calls: []*modelsv1.ToolCall{toolCall("allocation", `{"by":"asset_class"}`)}}
}

// fundReply is the fake model's answer to a question that names the fund: it
// reads the show card, then offers the names the card holds.
func fundReply(msgs []*modelsv1.AskMessage) scriptedReply {
	if !hasToolResult(msgs) {
		return scriptedReply{calls: []*modelsv1.ToolCall{toolCall("show", `{"name":"Synthetic"}`)}}
	}
	return scriptedReply{text: answerJSON(answerKindClarify, "Which one do you mean?", nil, nil, []string{"Synthetic brokerage", "Synthetic Fund"})}
}

// thesisFundReply is the fake model's answer to a question that asks for a thesis
// (D-229): it reads the show card of the fund, then suggests a thesis on it.
func thesisFundReply(msgs []*modelsv1.AskMessage) scriptedReply {
	if !hasToolResult(msgs) {
		return scriptedReply{calls: []*modelsv1.ToolCall{toolCall("show", `{"name":"Synthetic Fund","kind":"instrument"}`)}}
	}
	th := &structuredThesis{
		Statement: "Synthetic Fund keeps compounding at the pace its holders have seen.", Instruments: []string{"Synthetic Fund"}, Horizon: horizonLongTerm,
		Invalidation: []string{"Its fees double.", "It stops being listed."}, Sources: []string{"The show card for Synthetic Fund."},
	}
	return scriptedReply{text: opinionsJSON(answerKindThesis, "Here is a thesis on the fund, from its card.", []string{"c1"}, th)}
}

// hasToolResult reports whether the conversation holds a tool's result.
func hasToolResult(msgs []*modelsv1.AskMessage) bool {
	for _, m := range msgs {
		if m.GetMessage().GetRole() == modelsv1.MessageRole_MESSAGE_ROLE_TOOL {
			return true
		}
	}
	return false
}

// firstUser is the first user message of a conversation: the question.
func firstUser(msgs []*modelsv1.AskMessage) string {
	for _, m := range msgs {
		if m.GetMessage().GetRole() == modelsv1.MessageRole_MESSAGE_ROLE_USER {
			return m.GetMessage().GetContent()
		}
	}
	return ""
}

// toolCall is a scripted call of a tool.
func toolCall(name, args string) *modelsv1.ToolCall {
	return &modelsv1.ToolCall{Id: "c1", Name: name, ArgumentsJson: args}
}
