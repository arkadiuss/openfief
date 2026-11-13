package ui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// secretAsk is in every question, answer and tool result of these tests; it
// may reach the model and the page, and never a log line or an error (GR-12).
const secretAsk = "SECRET-ASK-TEXT"

// errSecret is an error whose text is the secret.
var errSecret = errors.New(secretAsk)

// scriptedPort is an askModelPort that answers from a script and keeps what
// it was sent. Its model may receive the tiers in allowed. It keeps the token
// limit of each call and says each call spent `used` tokens.
type scriptedPort struct {
	mu          sync.Mutex
	unavailable error
	allowed     map[string]bool
	replies     []askReply
	generateErr error
	permitErr   error
	availCalls  int
	requests    [][]askMessage
	offered     []int
	maxTokens   []int64
	used        int64
	refused     int
	permitted   [][]string
}

func (p *scriptedPort) available(context.Context, session) (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.availCalls++
	return p.unavailable == nil, p.unavailable
}

func (p *scriptedPort) permits(_ context.Context, _ session, tiers []string) (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.permitted = append(p.permitted, slices.Clone(tiers))
	if p.permitErr != nil {
		return false, p.permitErr
	}
	for _, t := range tiers {
		if !p.allowed[t] {
			return false, nil
		}
	}
	return true, nil
}

func (p *scriptedPort) generate(_ context.Context, _ session, msgs []askMessage, tools []askTool, maxTokens int64) (askReply, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.requests = append(p.requests, slices.Clone(msgs))
	p.offered = append(p.offered, len(tools))
	p.maxTokens = append(p.maxTokens, maxTokens)
	for i := range msgs {
		// The router refuses a conversation with a tier no model of the
		// route may receive before it sends anything.
		if !p.allowed[msgs[i].Tier] {
			p.refused++
			return askReply{}, codeErr(connect.CodePermissionDenied, nil)
		}
	}
	if p.generateErr != nil {
		return askReply{}, p.generateErr
	}
	next := askReply{Content: "done", Model: "p/m"}
	if len(p.replies) > 0 {
		next, p.replies = p.replies[0], p.replies[1:]
	}
	next.OutputTokens = p.used
	return next, nil
}

// calling is a model reply that asks for tools.
func calling(calls ...askCall) askReply { return askReply{Calls: calls, Model: "p/m"} }

// loopFixture is a modelAnswerer over a scripted port, with a tool table
// whose sync_status card is one fief's (TF) so a route that may receive T0
// and T2 but not TF is refused that card and gets the others, and a runner that makes a card
// per command and counts what it ran.
type loopFixture struct {
	port *scriptedPort
	a    *modelAnswerer
	logs *strings.Builder
	mu   sync.Mutex
	ran  []askCommand
	fail map[askCommand]error
}

func newLoopFixture(t *testing.T, allowed ...string) *loopFixture {
	t.Helper()
	f := &loopFixture{port: &scriptedPort{allowed: map[string]bool{}}, logs: &strings.Builder{}, fail: map[askCommand]error{}}
	for _, tier := range allowed {
		f.port.allowed[tier] = true
	}
	tools := askTools()
	for i := range tools {
		if tools[i].Name == "sync_status" {
			tools[i].Tier = tierTF
		}
	}
	f.a = &modelAnswerer{
		limits: DefaultAskLimits(),
		model:  f.port, tools: tools, today: func() kernel.Date { return mustDate(t, askToday) },
		log: slog.New(slog.NewTextHandler(f.logs, nil)),
		run: func(_ *http.Request, _ session, q *askQuery) (answerCard, error) {
			f.mu.Lock()
			defer f.mu.Unlock()
			f.ran = append(f.ran, q.Command)
			if err := f.fail[q.Command]; err != nil {
				return answerCard{}, err
			}
			return answerCard{
				Title: fmt.Sprintf("Card %d", q.Command), Text: secretAsk, Rows: []answerRow{{Label: "row", Figure: "1 000,00 PLN"}},
				Link: &answerLink{Href: "/somewhere", Text: "Open it"},
			}, nil
		},
	}
	return f
}

func (f *loopFixture) ask(question string) (answerCard, error) {
	return f.a.ask(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ask", http.NoBody), session{Token: "t", Origin: "o"}, question)
}

func titles(cards []answerCard) []string {
	out := make([]string, 0, len(cards))
	for i := range cards {
		out = append(out, cards[i].Title)
	}
	return out
}

// TestModelLoopShowsExactlyTheCardsUsed checks that the answer is the model's
// text with the cards its tool calls produced, in the order it called them,
// and no other.
func TestModelLoopShowsExactlyTheCardsUsed(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{
		calling(askCall{ID: "1", Name: "allocation", Args: `{"by":"currency"}`}, askCall{ID: "2", Name: "value", Args: `{}`}),
		calling(askCall{ID: "3", Name: "show", Args: `{"name":"bonds"}`}),
		{Content: "  Mostly PLN.  ", Model: "p/m"},
	}
	got, err := f.ask("how am I doing?")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"Mostly PLN.", "p/m", "how am I doing?", ""}, []string{got.Text, got.Model, got.Question, got.Title}); diff != "" {
		t.Errorf("answer's text, model, question and title, want the model's trimmed text under its name (-want +got):\n%s", diff)
	}
	wantCmds := []askCommand{askAllocation, askValue, askShow}
	if diff := cmp.Diff(wantCmds, f.ran); diff != "" {
		t.Errorf("commands run (-want +got):\n%s", diff)
	}
	wantTitles := []string{fmt.Sprintf("Card %d", askAllocation), fmt.Sprintf("Card %d", askValue), fmt.Sprintf("Card %d", askShow)}
	if diff := cmp.Diff(wantTitles, titles(got.Used)); diff != "" {
		t.Errorf("cards shown (-want +got):\n%s", diff)
	}
	if len(got.Withheld) != 0 || len(f.port.requests) != 3 {
		t.Fatalf("withheld %v and the model asked %d times, want none and 3", got.Withheld, len(f.port.requests))
	}
	// Every card the model was sent is one the page shows, and the other way round.
	shown := make([]string, 0, len(got.Used))
	for i := range got.Used {
		shown = append(shown, cardResult(&got.Used[i]))
	}
	sent := make([]string, 0, 4)
	for _, m := range toolMessages(f.port.requests[2]) {
		sent = append(sent, m.Content)
	}
	if diff := cmp.Diff(shown, sent); diff != "" {
		t.Errorf("tool results sent against cards shown (-shown +sent):\n%s", diff)
	}
}

// TestModelLoopTagsEveryMessageWithATier checks the tiers the router gets
// (D-206): the system prompt and the question are public, each tool result
// carries its tool's tier, a refusal carries none of the data's, and what the
// model said after it read the portfolio carries the most sensitive tier.
func TestModelLoopTagsEveryMessageWithATier(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{
		calling(askCall{ID: "1", Name: "value", Args: `{}`}, askCall{ID: "2", Name: "nope", Args: `{}`}),
		{Content: "ok", Model: "p/m"},
	}
	if _, err := f.ask("value?"); err != nil {
		t.Fatal(err)
	}
	want := []string{"system:T2", "user:T0", "assistant:T0", "tool:T0", "tool:T2"}
	if diff := cmp.Diff(want, roleTiers(f.port.requests[1])); diff != "" {
		t.Errorf("messages of the second round (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"system:T2", "user:T0"}, roleTiers(f.port.requests[0])); diff != "" {
		t.Errorf("messages of the first round (-want +got):\n%s", diff)
	}
}

// roleTiers writes each message as its role and tier.
func roleTiers(msgs []askMessage) []string {
	out := make([]string, 0, len(msgs))
	for i := range msgs {
		out = append(out, msgs[i].Role+":"+msgs[i].Tier)
	}
	return out
}

// TestModelLoopRefusesEachToolResultAboveWhatTheRouteAllows checks D-206 per
// tool result: with a model that may receive T0 and public data but not one
// fief's data (TF), the card of the TF tool is withheld, the model is told so
// without any of its data, the page names the card, and a T0 card still goes
// through. The question is T0, so a route with no T0 is refused before this
// (TestModelFailuresAreCardsThatNameTheCommands).
func TestModelLoopRefusesEachToolResultAboveWhatTheRouteAllows(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{
		calling(askCall{ID: "1", Name: "what_needs_me", Args: `{}`}, askCall{ID: "2", Name: "sync_status", Args: `{"connector":"xtb"}`}),
		{Content: "I could not read the sync status.", Model: "p/m"},
	}
	got, err := f.ask("what is wrong?")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]withheldCard{{Label: "Sync status", Tier: tierTF, Data: tierWords(tierTF)}}, got.Withheld); diff != "" {
		t.Errorf("withheld (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{fmt.Sprintf("Card %d", askNeeds)}, titles(got.Used)); diff != "" {
		t.Errorf("cards shown (-want +got):\n%s", diff)
	}
	tools := toolMessages(f.port.requests[1])
	if len(tools) != 2 {
		t.Fatalf("%d tool results sent, want 2", len(tools))
	}
	shown, refused := tools[0], tools[1]
	if !isRefusalWithoutData(&refused) {
		t.Errorf("the refused tool result = %+v, want a refusal with no data, tagged public", refused)
	}
	if !strings.Contains(shown.Content, secretAsk) || shown.Tier != tierT0 {
		t.Errorf("the shown tool result = %+v, want the card's data", shown)
	}
	// The router is asked about the conversation's tiers together, per result.
	wantAsked := [][]string{{tierT0}, {tierT0, tierTF}}
	if diff := cmp.Diff(wantAsked, f.port.permitted); diff != "" {
		t.Errorf("tiers asked about (-want +got):\n%s", diff)
	}
}

// isRefusalWithoutData reports whether m says a card was withheld, holds
// none of its data and is tagged public.
func isRefusalWithoutData(m *askMessage) bool {
	return strings.Contains(m.Content, "withheld") && !strings.Contains(m.Content, secretAsk) && !strings.Contains(m.Content, "1 000,00") && m.Tier == tierT2
}

func toolMessages(msgs []askMessage) []askMessage {
	var out []askMessage
	for _, m := range msgs {
		if m.Role == "tool" {
			out = append(out, m)
		}
	}
	return out
}

// TestModelLoopIsBounded checks the loop's limits (D-203, D-208): up to five
// calls of the model that offer the tools and one last that offers none so
// the model must answer, and at most ten tools run, the calls beyond that
// refused.
func TestModelLoopIsBounded(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	one := askCall{ID: "c", Name: "value", Args: `{}`}
	for range 10 {
		f.port.replies = append(f.port.replies, calling(one, one, one))
	}
	got, err := f.ask("endless")
	if err != nil {
		t.Fatal(err)
	}
	limits := DefaultAskLimits()
	// Every round offers the tools but the last, which must answer with what it has.
	wantOffered := make([]int, 0, limits.ToolRounds+1)
	for range limits.ToolRounds {
		wantOffered = append(wantOffered, len(f.a.tools))
	}
	wantOffered = append(wantOffered, 0)
	if diff := cmp.Diff(wantOffered, f.port.offered); diff != "" {
		t.Errorf("tools offered per round (-want +got):\n%s", diff)
	}
	if len(f.ran) != limits.ToolCalls {
		t.Errorf("%d tools ran, want at most %d for one question", len(f.ran), limits.ToolCalls)
	}
	if got.Title != "The model didn't finish" || got.Model != "" {
		t.Errorf("answer = %+v, want the card of a model that didn't finish", got)
	}
	// The calls over the limit still got an answer: a refusal the model can read.
	if !slices.ContainsFunc(toolMessages(f.port.requests[limits.ToolRounds]), func(m askMessage) bool {
		return strings.Contains(m.Content, "too many tool calls")
	}) {
		t.Errorf("no tool call was refused for the limit")
	}
}

// TestAskLimitsDefaultsAndOverrides checks D-208's defaults, which a zero
// field takes, and that a configured limit replaces only its own field.
func TestAskLimitsDefaultsAndOverrides(t *testing.T) {
	t.Parallel()
	want := AskLimits{ToolRounds: 5, ToolCalls: 10, OutputTokens: 800, Deadline: 2 * time.Minute}
	tests := []struct {
		name string
		in   AskLimits
		want AskLimits
	}{
		{name: "defaults", in: AskLimits{}, want: want},
		{name: "one field", in: AskLimits{ToolCalls: 3}, want: AskLimits{ToolRounds: 5, ToolCalls: 3, OutputTokens: 800, Deadline: 2 * time.Minute}},
		{name: "negative is unset", in: AskLimits{ToolRounds: -1, OutputTokens: -5}, want: want},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.in.withDefaults()); diff != "" {
				t.Errorf("limits (-want +got):\n%s", diff)
			}
		})
	}
	if diff := cmp.Diff(want, DefaultAskLimits()); diff != "" {
		t.Errorf("DefaultAskLimits (-want +got):\n%s", diff)
	}
}

// TestModelLoopStopsAtTheOutputTokenLimit checks D-208's 800 tokens per
// question: each call is limited to what is left of them, and once the
// calls have used them all no more are made, however many rounds remain.
func TestModelLoopStopsAtTheOutputTokenLimit(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.used = 300
	one := askCall{ID: "c", Name: "value", Args: `{}`}
	for range 10 {
		f.port.replies = append(f.port.replies, calling(one))
	}
	got, err := f.ask("endless")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]int64{800, 500, 200}, f.port.maxTokens); diff != "" {
		t.Errorf("token limit of each call (-want +got):\n%s", diff)
	}
	if got.Title != "The model didn't finish" {
		t.Errorf("answer = %+v, want the card of a model that spent its tokens", got)
	}
}

// TestModelLoopLimitsAreConfigurable checks that the limits a loop is given
// are the ones it keeps, not the defaults.
func TestModelLoopLimitsAreConfigurable(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.a.limits = AskLimits{ToolRounds: 1, ToolCalls: 1, OutputTokens: 50, Deadline: time.Minute}
	one := askCall{ID: "c", Name: "value", Args: `{}`}
	f.port.replies = []askReply{calling(one, one), {Content: "ok", Model: "p/m"}}
	if _, err := f.ask("two tools"); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]int{len(f.a.tools), 0}, f.port.offered); diff != "" {
		t.Errorf("tools offered per round (-want +got):\n%s", diff)
	}
	if len(f.ran) != 1 {
		t.Errorf("%d tools ran, want the configured 1", len(f.ran))
	}
	if diff := cmp.Diff([]int64{50, 50}, f.port.maxTokens); diff != "" {
		t.Errorf("token limit of each call (-want +got):\n%s", diff)
	}
}

func TestModelLoopBadCallsGetAFailureAndRunNothing(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{
		calling(
			askCall{ID: "1", Name: "delete_everything", Args: `{}`},
			askCall{ID: "2", Name: "show", Args: `{"name":"` + secretAsk + `","kind":"nonsense"}`},
			askCall{ID: "3", Name: "value", Args: `not json`},
		),
		{Content: "I could not.", Model: "p/m"},
	}
	got, err := f.ask("do something")
	if err != nil {
		t.Fatal(err)
	}
	if len(f.ran) != 0 || len(got.Used) != 0 {
		t.Errorf("ran %v and showed %d cards, want nothing: no call was valid", f.ran, len(got.Used))
	}
	for _, m := range toolMessages(f.port.requests[1]) {
		if !strings.HasPrefix(m.Content, `{"error":`) || m.Tier != tierT2 || strings.Contains(m.Content, secretAsk) {
			t.Errorf("tool result %+v, want a failure that holds none of the arguments", m)
		}
	}
	if got.Text != "I could not." || got.Model != "p/m" {
		t.Errorf("answer = %+v", got)
	}
}

func TestModelLoopToolReadFailures(t *testing.T) {
	t.Parallel()
	t.Run("a failed read is a failure the model reads, with none of the error", func(t *testing.T) {
		t.Parallel()
		f := newLoopFixture(t, tierT0, tierT2)
		f.fail[askValue] = fmt.Errorf("GetThing: %w", connect.NewError(connect.CodeNotFound, errSecret))
		f.port.replies = []askReply{calling(askCall{ID: "1", Name: "value", Args: `{}`}), {Content: "no", Model: "p/m"}}
		got, err := f.ask("value?")
		if err != nil {
			t.Fatal(err)
		}
		tool := toolMessages(f.port.requests[1])[0]
		if tool.Content != `{"error":"the data could not be read"}` || len(got.Used) != 0 {
			t.Errorf("tool result %q, cards %d; want the plain failure and no card", tool.Content, len(got.Used))
		}
	})
	t.Run("a session that ended ends the question", func(t *testing.T) {
		t.Parallel()
		f := newLoopFixture(t, tierT0, tierT2)
		f.fail[askValue] = connect.NewError(connect.CodeUnauthenticated, nil)
		f.port.replies = []askReply{calling(askCall{ID: "1", Name: "value", Args: `{}`})}
		if _, err := f.ask("value?"); connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("error = %v, want Unauthenticated up to the page", err)
		}
	})
	t.Run("a router that fails the tier check ends the question", func(t *testing.T) {
		t.Parallel()
		f := newLoopFixture(t, tierT0, tierT2)
		f.port.permitErr = connect.NewError(connect.CodeUnavailable, nil)
		f.port.replies = []askReply{calling(askCall{ID: "1", Name: "value", Args: `{}`})}
		got, err := f.ask("value?")
		if err != nil || got.Title != "The model didn't answer" {
			t.Errorf("answer = %+v, %v; want the card of a model that didn't answer", got, err)
		}
	})
}

// TestModelQuestionIsT0 checks the author's correction of D-206 for the Ask
// box: the question is the owner's own words, which may carry amounts or
// holdings, so it is T0 like a tool's result. A route whose models may not
// receive T0 is refused, with nothing sent and no tool run, however public
// the rest is; one that may receive it is asked.
func TestModelQuestionIsT0(t *testing.T) {
	t.Parallel()
	testQuestionRefusedWithoutT0(t)
	t.Run("a route with T0 is asked, the question tagged T0", func(t *testing.T) {
		t.Parallel()
		f := newLoopFixture(t, tierT0, tierT2)
		got, err := f.ask("what is my value?")
		if err != nil || got.Text != "done" {
			t.Fatalf("answer = %+v, %v", got, err)
		}
		if diff := cmp.Diff([]string{"system:T2", "user:T0"}, roleTiers(f.port.requests[0])); diff != "" {
			t.Errorf("messages (-want +got):\n%s", diff)
		}
	})
}

// testQuestionRefusedWithoutT0 is the refusal half of TestModelQuestionIsT0.
func testQuestionRefusedWithoutT0(t *testing.T) {
	t.Helper()
	t.Run("a route without T0 is refused before anything is sent", func(t *testing.T) {
		t.Parallel()
		f := newLoopFixture(t, tierT2)
		got, err := f.ask("I hold 1 000,00 PLN of X, is that too much?")
		if err != nil {
			t.Fatal(err)
		}
		if got.Title != "No model may see this question" || got.Text == "" || len(got.Help) == 0 {
			t.Errorf("answer = %+v, want the card that names the commands", got)
		}
		if !strings.Contains(got.Text, "T0") || !strings.Contains(got.Text, "Nothing was sent") {
			t.Errorf("text = %q, want it to say the question is T0 and nothing was sent", got.Text)
		}
		if f.port.refused != 1 || len(f.ran) != 0 || len(f.port.permitted) != 0 {
			t.Errorf("refused %d times, ran %v, asked about tiers %v; want one refusal and nothing else", f.port.refused, f.ran, f.port.permitted)
		}
	})
}

// TestModelFailuresAreCardsThatNameTheCommands checks what the person sees
// when the router won't or can't answer, by the core's Connect code alone.
func TestModelFailuresAreCardsThatNameTheCommands(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		err       error
		wantTitle string
		wantLog   bool
	}{
		{"no model may see the question", connect.NewError(connect.CodePermissionDenied, errSecret), "No model may see this question", false},
		{"every model is down", connect.NewError(connect.CodeUnavailable, errSecret), "The model didn't answer", false},
		{"the model timed out", fmt.Errorf("GenerateAsk: %w", context.DeadlineExceeded), "The model didn't answer", false},
		{"the switch went off", connect.NewError(connect.CodeFailedPrecondition, errSecret), "I can't answer that yet", false},
		{"the core failed", connect.NewError(connect.CodeInternal, errSecret), "The model couldn't answer", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newLoopFixture(t, tierT0, tierT2)
			f.port.generateErr = tt.err
			got, err := f.ask("anything " + secretAsk)
			if err != nil || got.Title != tt.wantTitle {
				t.Fatalf("answer = %+v, %v; want the card %q", got, err, tt.wantTitle)
			}
			// A failure names no model, shows no card and, but for today's
			// no-model answer, lists the commands to fall back on.
			if got.Model != "" || len(got.Used) != 0 || (tt.wantTitle != "I can't answer that yet" && len(got.Help) == 0) {
				t.Errorf("the card = %+v, want no model, no cards and the commands", got)
			}
			logged := f.logs.String()
			if strings.Contains(logged, secretAsk) || tt.wantLog != strings.Contains(logged, "ask model call failed") {
				t.Errorf("log = %q, want a line only for a core failure: %v, and never the question", logged, tt.wantLog)
			}
		})
	}
}

// TestModelEndedSessionEndsTheQuestion checks that a session the core no
// longer accepts is the page's to handle, not a card.
func TestModelEndedSessionEndsTheQuestion(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.generateErr = connect.NewError(connect.CodeUnauthenticated, errSecret)
	if _, err := f.ask("anything"); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("error = %v, want Unauthenticated up to the page", err)
	}
}

// TestModelAnswerNeverLogsTheQuestionOrTheAnswer runs a whole question with a
// secret in the question, the model's tool arguments and answer, and a card,
// and checks no log line holds it (GR-12).
func TestModelAnswerNeverLogsTheQuestionOrTheAnswer(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.fail[askShow] = errSecret
	f.port.replies = []askReply{
		calling(askCall{ID: "1", Name: "value", Args: `{}`}, askCall{ID: "2", Name: "show", Args: `{"name":"` + secretAsk + `"}`}),
		{Content: "the answer is " + secretAsk, Model: "p/m"},
	}
	got, err := f.ask("my question is " + secretAsk)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.Text, secretAsk) {
		t.Errorf("answer text = %q", got.Text)
	}
	if f.logs.Len() != 0 {
		t.Errorf("a normal question logged %q, want nothing", f.logs.String())
	}
}

func TestModelAnswerText(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	long := strings.Repeat("é", maxModelAnswerRunes+10)
	tests := []struct {
		name, reply, want string
	}{
		{"empty", "  \n ", "The model returned no text."},
		{"trimmed", "  hello  ", "hello"},
		{"cut at the limit", long, strings.Repeat("é", maxModelAnswerRunes) + "…"},
	}
	for _, tt := range tests {
		f.port.replies = []askReply{{Content: tt.reply, Model: "p/m"}}
		got, err := f.ask("q")
		if err != nil || got.Text != tt.want {
			t.Errorf("%s: text = %q, %v; want %q", tt.name, got.Text, err, tt.want)
		}
	}
}

// TestRoutedAnswererFallsBackToTheCommandsAlone checks the Ask box with no
// model behaves as it did before there was one (D-203): the model is not
// asked, and a question beyond the commands gets the commands.
func TestRoutedAnswererFallsBackToTheCommandsAlone(t *testing.T) {
	t.Parallel()
	today := func() kernel.Date { return mustDate(t, askToday) }
	tests := []struct {
		name        string
		question    string
		unavailable error
		noModel     bool
	}{
		{name: "no model client", question: "buy more bonds", noModel: true},
		{name: "the switch is off, the toggle or the route is missing", question: "buy more bonds", unavailable: errUnavailable},
		{name: "a question longer than the box takes", question: strings.Repeat("a ", maxModelQuestionRunes)},
		{name: "an empty question", question: "  "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newLoopFixture(t, tierT0, tierT2)
			f.port.unavailable = tt.unavailable
			a := routedAnswerer{commands: commandAnswerer{s: &server{today: today}}, model: f.a, today: today, log: f.a.log}
			if tt.noModel {
				a.model = nil
			}
			got, err := a.answer(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ask", http.NoBody), session{Token: "t"}, tt.question)
			if err != nil || got.Title != "I can't answer that yet" || len(got.Help) == 0 || got.Model != "" || f.port.generateCalls() != 0 {
				t.Errorf("answer = %+v, %v after %d model calls; want today's \"I can't answer that yet\" with the commands and no call", got, err, f.port.generateCalls())
			}
		})
	}
}

// errUnavailable is a core that can't say whether the Ask box has a model.
var errUnavailable = connect.NewError(connect.CodeUnavailable, nil)

func (p *scriptedPort) generateCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.requests)
}
