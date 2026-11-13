package ui

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"sync"

	"connectrpc.com/connect"

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
	// failWith, when set, is the error GenerateAsk returns.
	failWith error
	// availableErr, when set, is the error GetAskModel returns.
	availableErr error
	// requests are the conversations GenerateAsk was sent, tiers included.
	requests [][]*modelsv1.AskMessage
	// toolsSeen are the tool names each request offered.
	toolsSeen [][]string
	// sessions are the session tokens each call carried.
	sessions []string
	bearers  []string
	checks   [][]string
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
	return connect.NewResponse(&modelsv1.GetAskModelResponse{Available: f.on}), nil
}

func (f *fakeAskModel) CheckAskTiers(_ context.Context, r *connect.Request[modelsv1.CheckAskTiersRequest]) (*connect.Response[modelsv1.CheckAskTiersResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.checks = append(f.checks, slices.Clone(r.Msg.GetTiers()))
	return connect.NewResponse(&modelsv1.CheckAskTiersResponse{Allowed: f.on && f.mayReceive(r.Msg.GetTiers())}), nil
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
	next := scriptedReply{text: "done"}
	switch {
	case f.reply != nil:
		next = f.reply(r.Msg.GetMessages())
	case len(f.script) > 0:
		next, f.script = f.script[0], f.script[1:]
	}
	resp := &modelsv1.GenerateAskResponse{Content: next.text, ToolCalls: next.calls, Model: "p/m", FinishReason: modelsv1.FinishReason_FINISH_REASON_STOP}
	if len(next.calls) > 0 {
		resp.FinishReason = modelsv1.FinishReason_FINISH_REASON_TOOL_CALLS
	}
	return connect.NewResponse(resp), nil
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
		if !f.allowed[t] {
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
func browserSetUpAskModel(f *fakeAskModel, tiers string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.on, f.allowed = false, map[string]bool{}
	f.reply = nil
	if tiers == "" || tiers == "off" {
		return
	}
	f.on = true
	for t := range strings.SplitSeq(tiers, ",") {
		f.allowed[t] = true
	}
	f.reply = func(msgs []*modelsv1.AskMessage) scriptedReply {
		for _, m := range msgs {
			if m.GetMessage().GetRole() != modelsv1.MessageRole_MESSAGE_ROLE_TOOL {
				continue
			}
			if strings.Contains(m.GetMessage().GetContent(), "withheld") {
				return scriptedReply{text: "I may not read your allocation, so I can't say."}
			}
			return scriptedReply{text: "Most of it is listed.\nThe rest is cash."}
		}
		return scriptedReply{calls: []*modelsv1.ToolCall{toolCall("allocation", `{"by":"asset_class"}`)}}
	}
}

// toolCall is a scripted call of a tool.
func toolCall(name, args string) *modelsv1.ToolCall {
	return &modelsv1.ToolCall{Id: "c1", Name: name, ArgumentsJson: args}
}
