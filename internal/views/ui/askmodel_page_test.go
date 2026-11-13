package ui

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
)

// asks sends question to the Ask box as an htmx request of session.
func (h *harness) asks(t *testing.T, session, question string) *httptest.ResponseRecorder {
	t.Helper()
	return h.serve(t, htmx(withSession(request(http.MethodGet, "/ask?"+url.Values{"q": {question}}.Encode(), ""), session)))
}

func toolNames(t *testing.T, msgs []*modelsv1.AskMessage) []string {
	t.Helper()
	var out []string
	for _, m := range msgs {
		if m.GetMessage().GetRole() == modelsv1.MessageRole_MESSAGE_ROLE_TOOL {
			out = append(out, m.GetDataTier()+" "+m.GetMessage().GetContent())
		}
	}
	return out
}

// TestAskBoxAnswersBeyondItsCommandsWithAModel checks D-203 through the
// page: with a model configured, a question that is no command goes to it, it
// calls the same answerers as tools, reads with the asking person's own
// session, and the page shows its text and exactly the cards it read.
func TestAskBoxAnswersBeyondItsCommandsWithAModel(t *testing.T) {
	t.Parallel()
	for _, session := range []string{ownerSession, readSession} {
		t.Run(session, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.askModel.turnOn("T0", "T2")
			h.askModel.say(
				scriptedReply{calls: []*modelsv1.ToolCall{toolCall("allocation", `{"by":"currency"}`)}},
				scriptedReply{text: "Two thirds of it is listed."},
			)
			rec := h.asks(t, session, "how diversified am I?")
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec,
				`class="bubble"`, "how diversified am I?", "Two thirds of it is listed.",
				"Answered by p/m · it read only the cards below", `class="used-cards"`, `class="used-card"`,
				"By currency, gross, on 2026-03-31", "66.7%", `href="/?by=currency`,
			)
			wantNoBody(t, rec, "no AI model", "withheld", "<html")
			if got := h.askModel.generated(); got != 2 {
				t.Fatalf("the model was asked %d times, want 2", got)
			}
			if diff := cmp.Diff([]string{"allocation", "performance", "show", "sync_status", "value", "what_needs_me"}, sortedCopy(h.askModel.toolsSeen[0])); diff != "" {
				t.Errorf("tools offered (-want +got):\n%s", diff)
			}
			second := h.askModel.requests[1]
			results := toolNames(t, second)
			if len(results) != 1 || !strings.HasPrefix(results[0], "T0 ") || !strings.Contains(results[0], `"label":"listed"`) {
				t.Errorf("tool results the model got = %v, want the allocation card as T0 data", results)
			}
			wantAll(t, "a call to the core carried session", h.askModel.sessions, session)
			wantAll(t, "a call to the router was admitted with", h.askModel.bearers, "Bearer "+pluginToken)
			reads := seen(h.read, func() []string { return h.read.bearers })
			if len(reads) == 0 {
				t.Fatal("the tool read nothing")
			}
			// Their grants limit what the model sees: it reads as they do.
			wantAll(t, "a tool read with", reads, "Bearer "+session)
		})
	}
}

// wantAll fails when any of got is not want.
func wantAll(t *testing.T, what string, got []string, want string) {
	t.Helper()
	for _, g := range got {
		if g != want {
			t.Errorf("%s %q, want %q", what, g, want)
		}
	}
}

func sortedCopy(s []string) []string {
	out := slices.Clone(s)
	slices.Sort(out)
	return out
}

// TestAskBoxCommandsAnswerAtOnceWithAModel checks that a model being
// configured changes nothing for the commands: they answer themselves, as
// before, and the router is not even asked.
func TestAskBoxCommandsAnswerAtOnceWithAModel(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.askModel.turnOn("T0", "T2")
	for _, q := range []string{"What needs me?", "allocation by currency", "value on 2026-03-31", "show synthetic brokerage"} {
		rec := h.asks(t, ownerSession, q)
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec, "Answered from your data · no AI model")
		wantNoBody(t, rec, "Answered by", `class="used-card"`)
	}
	if n := h.askModel.generated(); n != 0 {
		t.Errorf("the model was asked %d times for commands", n)
	}
	h.askModel.mu.Lock()
	defer h.askModel.mu.Unlock()
	if len(h.askModel.sessions) != 0 {
		t.Errorf("the router was called %d times for commands, want none", len(h.askModel.sessions))
	}
}

// TestAskBoxWithholdsACardAboveWhatTheRouteAllows checks D-206 through the
// page: a model allowed public data only is refused the allocation's data,
// the page names the withheld card without drawing it, and the model's
// answer is what it could say without it.
func TestAskBoxWithholdsACardAboveWhatTheRouteAllows(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.askModel.turnOn("T2")
	h.askModel.say(
		scriptedReply{calls: []*modelsv1.ToolCall{toolCall("allocation", `{}`)}},
		scriptedReply{text: "I am not allowed to see your allocation."},
	)
	rec := h.asks(t, ownerSession, "how diversified am I?")
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		"I am not allowed to see your allocation.", "Answered by p/m · it read nothing from your data",
		`class="withheld"`, "Withheld from the model: the Allocation card", "(T0)",
	)
	wantNoBody(t, rec, `class="used-card"`, "By asset class", "66.7%", "1 000,00")
	results := toolNames(t, h.askModel.requests[1])
	if len(results) != 1 || !strings.HasPrefix(results[0], "T2 ") || !strings.Contains(results[0], "withheld") || strings.Contains(results[0], "listed") {
		t.Errorf("what the model got for the card = %v, want a refusal tagged public with none of its data", results)
	}
	if want := [][]string{{"T0", "T2"}}; !cmp.Equal(want, h.askModel.checks) {
		t.Errorf("tier checks = %v, want %v", h.askModel.checks, want)
	}
}

// TestAskBoxBehavesAsBeforeWithoutAModel checks the fallbacks: with no
// switch, toggle or route, or when the core can't say, a question beyond the
// commands gets today's "I can't answer that yet" and the model is not asked.
func TestAskBoxBehavesAsBeforeWithoutAModel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mod  func(f *fakeAskModel)
	}{
		{"nothing configured", func(*fakeAskModel) {}},
		{"the core can't say", func(f *fakeAskModel) {
			f.turnOn("T0", "T2")
			f.availableErr = codeErr(connect.CodeUnavailable, nil)
		}},
		{"the switch went off after the box looked", func(f *fakeAskModel) {
			f.turnOn("T0", "T2")
			f.failWith = codeErr(connect.CodeFailedPrecondition, nil)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			tt.mod(h.askModel)
			rec := h.asks(t, ownerSession, "buy more bonds")
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec, "Answered from your data · no AI model", "I can&#39;t answer that yet", "What needs me", "Allocation by currency",
				"The Ask box understands a few commands and answers them from your own record; it holds no AI model.")
			wantNoBody(t, rec, "Answered by")
			if n := h.askModel.generated(); n != 0 {
				t.Errorf("the model was asked %d times", n)
			}
		})
	}
}

func TestAskBoxCardsForAModelThatCannotAnswer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"no model may see the question", codeErr(connect.CodePermissionDenied, nil), "No model may see this question"},
		{"every model is down", codeErr(connect.CodeUnavailable, nil), "The model didn&#39;t answer"},
		{"the core failed", codeErr(connect.CodeInternal, nil), "The model couldn&#39;t answer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.askModel.turnOn("T0", "T2")
			h.askModel.failWith = tt.err
			rec := h.asks(t, ownerSession, "how diversified am I?")
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec, tt.want, "answer from your own record", "The model gave no answer", `class="ask-help"`, "What needs me")
			wantNoBody(t, rec, "Answered by", `class="used-card"`)
		})
	}
	t.Run("an ended session goes to the login page", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		h.askModel.turnOn("T0", "T2")
		h.askModel.failWith = codeErr(connect.CodeUnauthenticated, nil)
		rec := h.asks(t, ownerSession, "how diversified am I?")
		wantStatus(t, rec, http.StatusUnauthorized)
		if got := rec.Header().Get(hxRedirect); got != "/login" {
			t.Errorf("%s = %q, want /login", hxRedirect, got)
		}
	})
}

// TestAskBoxSaysWhereAQuestionGoes checks that the overview's Ask box says a
// question beyond the commands goes to a model only when it does.
func TestAskBoxSaysWhereAQuestionGoes(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	off := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantBody(t, off, "The box holds no AI model.", `maxlength="200"`)
	wantNoBody(t, off, "goes to the model")
	h.askModel.turnOn("T0", "T2")
	on := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantBody(t, on, "Any other question goes to the model the owner set up", "never logs the question or the answer", `maxlength="300"`)
	wantNoBody(t, on, "holds no AI model")
	page := h.serve(t, withSession(request(http.MethodGet, "/?q=anything+else", ""), ownerSession))
	wantBody(t, page, `value="anything else"`, `class="bubble"`, "Answered by p/m")
}

// TestAskBoxNeverLogsTheQuestionOrAnswerWithAModel checks that a question
// that goes to a model whose router fails is logged by its code alone
// (GR-12), and that a question that succeeds logs nothing.
func TestAskBoxNeverLogsTheQuestionOrAnswerWithAModel(t *testing.T) {
	t.Parallel()
	var logged strings.Builder
	h := newHarnessLogging(t, slog.New(slog.NewTextHandler(&logged, nil)))
	h.askModel.turnOn("T0", "T2")
	h.askModel.say(
		scriptedReply{calls: []*modelsv1.ToolCall{toolCall("show", `{"name":"secret-question"}`)}},
		scriptedReply{text: "answer holding secret-answer"},
	)
	rec := h.asks(t, ownerSession, "what is secret-question?")
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "secret-answer")
	if strings.Contains(logged.String(), "secret") {
		t.Errorf("log = %q, want no question and no answer", logged.String())
	}
	h.askModel.mu.Lock()
	h.askModel.failWith = codeErr(connect.CodeInternal, nil)
	h.askModel.mu.Unlock()
	rec = h.asks(t, ownerSession, "what is secret-question?")
	wantStatus(t, rec, http.StatusOK)
	if strings.Contains(logged.String(), "secret") || !strings.Contains(logged.String(), "ask model call failed") {
		t.Errorf("log = %q, want the failure by its code without the question", logged.String())
	}
}
