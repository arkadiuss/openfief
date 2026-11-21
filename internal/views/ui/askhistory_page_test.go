package ui

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/views/app/answer"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// asksAfter asks question the way the page does: a POST whose body carries the
// question and the earlier turns, as JSON in the history field.
func (h *harness) asksAfter(t *testing.T, question string, earlier []askTurn) *httptest.ResponseRecorder {
	t.Helper()
	body := url.Values{askParam: {question}}
	if earlier != nil {
		raw, err := json.Marshal(earlier)
		if err != nil {
			t.Fatal(err)
		}
		body.Set(historyField, string(raw))
	}
	return h.serve(t, htmx(withSession(form("/ask", body.Encode()), ownerSession)))
}

func TestAFollowUpReachesTheModelWithTheEarlierTurnsAndAddsItsOwn(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.askModel.turnOn("T0", "T2")
	h.askModel.say(
		scriptedReply{calls: []*modelsv1.ToolCall{toolCall("performance", `{"period":"last_year"}`)}},
		scriptedReply{text: answerText("Last year it was listed.", "c1")},
	)
	earlier := []askTurn{{Question: "How did I do this year?", Answer: "Listed.", Cards: []string{"Performance; this year"}}}
	rec := h.asksAfter(t, "and last year?", earlier)
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Last year it was listed.", `class="bubble"`, "and last year?", "data-turn=")
	first := firstUser(h.askModel.requests[0])
	for _, want := range []string{"How did I do this year?", "untrusted", "New question: and last year?"} {
		if !strings.Contains(first, want) {
			t.Errorf("the model's first message lacks %q: %s", want, first)
		}
	}
	// The turn the page keeps with this answer is the question, the text and the card, and no HTML.
	body := rec.Body.String()
	start := strings.Index(body, `data-turn="`)
	end := strings.Index(body[start+len(`data-turn="`):], `"`)
	var turn askTurn
	raw := strings.NewReplacer("&#34;", `"`, "&amp;", "&").Replace(body[start+len(`data-turn="`) : start+len(`data-turn="`)+end])
	if err := json.Unmarshal([]byte(raw), &turn); err != nil {
		t.Fatalf("data-turn %q: %v", raw, err)
	}
	if turn.Question != "and last year?" || turn.Answer != "Last year it was listed." || len(turn.Cards) != 1 || strings.Contains(raw, "<") {
		t.Errorf("turn = %+v (%s), want the new question, its text and one compact card", turn, raw)
	}
}

func TestACommandInsideAConversationAnswersAtOnceAndJoinsTheThread(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.askModel.turnOn("T0", "T2")
	earlier := []askTurn{{Question: "How did I do?", Answer: "Listed."}}
	rec := h.asksAfter(t, "allocation by currency", earlier)
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Answered from your data · no AI model", "By currency", "data-turn=")
	if n := h.askModel.generated(); n != 0 {
		t.Errorf("the model was asked %d times for a command inside a conversation", n)
	}
}

func TestTheServerCapsAndDistrustsWhatTheBrowserSent(t *testing.T) {
	t.Parallel()
	earlier := make([]askTurn, 0, 12)
	for range 12 {
		earlier = append(earlier, askTurn{Question: "marker-question", Answer: repeated("w", 900), Cards: []string{repeated("c", 900)}})
	}
	h := newHarness(t)
	h.askModel.turnOn("T0", "T2")
	rec := h.asksAfter(t, "next question?", earlier)
	wantStatus(t, rec, http.StatusOK)
	got := firstUser(h.askModel.requests[0])
	if n := strings.Count(got, "marker-question"); n < 1 || n > maxHistoryTurns {
		t.Errorf("the model got %d earlier turns, want 1 to %d", n, maxHistoryTurns)
	}
	if n := strings.Count(got, "w"); n > maxHistoryRunes {
		t.Errorf("the earlier turns hold %d characters, want at most %d", n, maxHistoryRunes)
	}

	// A history that is not an array of turns is dropped, and the question goes alone.
	h2 := newHarness(t)
	h2.askModel.turnOn("T0", "T2")
	rec = h2.serve(t, htmx(withSession(form("/ask", url.Values{askParam: {"plain question?"}, historyField: {`{"q":"x"}`}}.Encode()), ownerSession)))
	wantStatus(t, rec, http.StatusOK)
	if got := firstUser(h2.askModel.requests[0]); got != "plain question?" {
		t.Errorf("with a history that is no array the model's message is %q, want the question alone", got)
	}

	// A body over its cap is refused before anything is read of it.
	h3 := newHarness(t)
	h3.askModel.turnOn("T0", "T2")
	rec = h3.serve(t, htmx(withSession(form("/ask", url.Values{askParam: {"x?"}, historyField: {repeated("a", maxAskBody+1)}}.Encode()), ownerSession)))
	wantStatus(t, rec, http.StatusRequestEntityTooLarge)
	if n := h3.askModel.generated(); n != 0 {
		t.Errorf("the model was asked %d times for a body over the cap", n)
	}
}

func TestNeitherTheQuestionNorTheConversationIsLogged(t *testing.T) {
	t.Parallel()
	var logged strings.Builder
	h := newHarnessLogging(t, slog.New(slog.NewTextHandler(&logged, &slog.HandlerOptions{Level: slog.LevelDebug})))
	h.askModel.turnOn("T0", "T2")
	h.askModel.say(
		scriptedReply{calls: []*modelsv1.ToolCall{toolCall("show", `{"name":"secret-new-question"}`)}},
		scriptedReply{text: answerText("answer holding secret-new-answer", "c1")},
	)
	earlier := []askTurn{{Question: "secret-old-question", Answer: "secret-old-answer", Cards: []string{"secret-old-card"}}}
	rec := h.asksAfter(t, "what is secret-new-question?", earlier)
	wantStatus(t, rec, http.StatusOK)
	// Then the model fails, and a command and an unreadable history are asked as well: each is logged by its code alone, or not at all.
	h.askModel.mu.Lock()
	h.askModel.failWith = codeErr(connect.CodeInternal, nil)
	h.askModel.mu.Unlock()
	h.asksAfter(t, "what is secret-new-question?", earlier)
	h.asksAfter(t, "allocation by currency", earlier)
	h.serve(t, htmx(withSession(form("/ask", url.Values{askParam: {"secret-new-question?"}, historyField: {"secret-old-question{"}}.Encode()), ownerSession)))
	if strings.Contains(logged.String(), "secret") {
		t.Errorf("the log holds a question, an answer or an earlier turn: %q", logged.String())
	}
}

// TestTheBoxsToolsAreTheMCPServersToolbox checks D-224's last sentence: the
// tools the page's model is offered are exactly the shared toolbox the MCP
// server offers (D-213): the same names, descriptions and argument schemas,
// in the same order, and no tool that changes anything.
func TestTheBoxsToolsAreTheMCPServersToolbox(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.askModel.turnOn("T0", "T2")
	rec := h.asksAfter(t, "how diversified am I?", []askTurn{{Question: "hello", Answer: "Hello."}})
	wantStatus(t, rec, http.StatusOK)
	box, err := answer.NewToolbox(&answer.ToolboxConfig{Read: viewsv1connect.NewReadServiceClient(http.DefaultClient, "http://core.invalid")})
	if err != nil {
		t.Fatal(err)
	}
	offered := make([]answer.ToolInfo, 0, len(h.askModel.toolDefs[0]))
	for _, tool := range h.askModel.toolDefs[0] {
		offered = append(offered, answer.ToolInfo{Name: tool.GetName(), Description: tool.GetDescription(), Schema: tool.GetParametersSchemaJson()})
	}
	if diff := cmp.Diff(box.Tools(), offered); diff != "" {
		t.Errorf("the model's tools differ from the shared toolbox's, which MCP serves (-toolbox +model):\n%s", diff)
	}
	if len(offered) == 0 {
		t.Fatal("the model was offered no tools")
	}
}

func TestAViewerAsksByPostLikeAnyone(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, htmx(withSession(form("/ask", url.Values{askParam: {"what needs me?"}}.Encode()), viewerSession)))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `class="bubble"`)
}

func TestAPostedQuestionNeedsTheSameOriginAsEveryPost(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	r := htmx(withSession(form("/ask", url.Values{askParam: {"what needs me?"}}.Encode()), ownerSession))
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	wantStatus(t, h.serve(t, r), http.StatusForbidden)
}

// TestTheCheckForAFollowUpIsTheSameAsForAFirstQuestion keeps the Turnstile
// flow (D-208) as it was: a follow-up that goes to the model with no token is
// told a check is needed, and with a good token in its body goes on with its
// history. The check reads its token from the body of a POST as from the
// query of a GET.
func TestTheCheckForAFollowUpIsTheSameAsForAFirstQuestion(t *testing.T) {
	t.Parallel()
	h, _, _ := turnstileHarness(t)
	earlier := []askTurn{{Question: "How did I do this year?", Answer: "Listed."}}
	rec := h.asksAfter(t, "and last year?", earlier)
	wantStatus(t, rec, http.StatusOK)
	if got := rec.Header().Get(hxTrigger); got != checkNeededTrigger {
		t.Errorf("HX-Trigger = %q, want %q for a follow-up with no token", got, checkNeededTrigger)
	}
	if n := h.askModel.generated(); n != 0 {
		t.Errorf("the model was asked %d times before the check passed", n)
	}
	raw, err := json.Marshal(earlier)
	if err != nil {
		t.Fatal(err)
	}
	body := url.Values{askParam: {"and last year?"}, historyField: {string(raw)}, turnstileField: {"ok-follow-up"}}
	rec = h.serve(t, htmx(withSession(form("/ask", body.Encode()), ownerSession)))
	wantStatus(t, rec, http.StatusOK)
	if got := rec.Header().Get(hxTrigger); got != "" {
		t.Errorf("HX-Trigger = %q after a good token, want none", got)
	}
	if got := firstUser(h.askModel.requests[0]); !strings.Contains(got, "How did I do this year?") {
		t.Errorf("the model's first message %q lacks the earlier turn", got)
	}
}
