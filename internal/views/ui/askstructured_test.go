package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/views/app/answer"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
)

func TestParseStructuredAnswer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
		cards   int
		want    structuredAnswer
		err     error
	}{
		{
			name: "an answer", content: answerText("It is 10.", "c1"), cards: 1,
			want: structuredAnswer{Kind: "answer", Text: "It is 10.", Cards: []string{"c1"}, Candidates: []string{}},
		},
		{name: "the text is trimmed", content: answerText("  It is 10.  "), want: structuredAnswer{Kind: "answer", Text: "It is 10.", Cards: []string{}, Candidates: []string{}}},
		{
			name: "a refusal's suggestions are ignored", content: answerJSON("refuse", "No.", nil, []string{"What needs me?", "Allocation by currency", "Show bonds"}, nil),
			want: structuredAnswer{Kind: "refuse", Text: "No.", Cards: []string{}, Candidates: []string{}},
		},
		{
			name: "a clarifying answer with its choices", content: answerJSON("clarify", "Which?", nil, nil, []string{"Demo World ETF", "Demo Emerging Markets ETF"}),
			want: structuredAnswer{Kind: "clarify", Text: "Which?", Cards: []string{}, Candidates: []string{"Demo World ETF", "Demo Emerging Markets ETF"}},
		},
		// A slip in suggestions never fails a reply: the box ignores them.
		{name: "suggestions that are no commands", content: answerJSON("refuse", "No.", nil, []string{"Write me a poem", "buy more bonds"}, nil), want: structuredAnswer{Kind: "refuse", Text: "No.", Cards: []string{}, Candidates: []string{}}},
		{name: "too many suggestions", content: answerJSON("refuse", "No.", nil, slices.Repeat([]string{"What needs me?"}, 20), nil), want: structuredAnswer{Kind: "refuse", Text: "No.", Cards: []string{}, Candidates: []string{}}},
		{name: "suggestions of the wrong type", content: `{"kind":"refuse","text":"No.","cards":[],"suggestions":"What needs me?","candidates":[]}`, want: structuredAnswer{Kind: "refuse", Text: "No.", Cards: []string{}, Candidates: []string{}}},
		{name: "suggestions that are an object", content: `{"kind":"refuse","text":"No.","cards":[],"suggestions":{"a":1},"candidates":[]}`, want: structuredAnswer{Kind: "refuse", Text: "No.", Cards: []string{}, Candidates: []string{}}},
		{name: "no suggestions member", content: `{"kind":"refuse","text":"No.","cards":[],"candidates":[]}`, want: structuredAnswer{Kind: "refuse", Text: "No.", Cards: []string{}, Candidates: []string{}}},
		{name: "null lists are empty", content: `{"kind":"answer","text":"x","cards":null,"suggestions":null,"candidates":null}`, want: structuredAnswer{Kind: "answer", Text: "x"}},
		// What is not an answer.
		{name: "not JSON", content: "It is 10.", err: errAnswerNotJSON},
		{name: "empty", content: "", err: errAnswerNotJSON},
		{name: "a JSON array", content: `["a"]`, err: errAnswerNotJSON},
		{name: "a field the schema lacks", content: `{"kind":"answer","text":"x","cards":[],"suggestions":[],"candidates":[],"extra":1}`, err: errAnswerNotJSON},
		{name: "a field of the wrong type", content: `{"kind":"answer","text":"x","cards":"c1","suggestions":[],"candidates":[]}`, err: errAnswerNotJSON},
		{name: "text after the object", content: answerText("x") + ` {"kind":"answer"}`, err: errAnswerNotJSON},
		{name: "a code fence", content: "```json\n" + answerText("x") + "\n```", err: errAnswerNotJSON},
		{name: "an unknown kind", content: answerJSON("maybe", "x", nil, nil, nil), err: errAnswerKind},
		{name: "no kind", content: `{"text":"x","cards":[],"suggestions":[],"candidates":[]}`, err: errAnswerKind},
		{name: "empty text", content: answerText("   "), err: errAnswerNoText},
		{name: "a card that was not read", content: answerText("x", "c2"), cards: 1, err: errAnswerUnknownCard},
		{name: "a card when none was read", content: answerText("x", "c1"), err: errAnswerUnknownCard},
		{name: "c0", content: answerText("x", "c0"), cards: 3, err: errAnswerUnknownCard},
		{name: "c01", content: answerText("x", "c01"), cards: 3, err: errAnswerUnknownCard},
		{name: "a capital C", content: answerText("x", "C1"), cards: 3, err: errAnswerUnknownCard},
		{name: "a card's title instead of its id", content: answerText("x", "By currency"), cards: 3, err: errAnswerUnknownCard},
		{name: "a clarifying answer with no choices", content: answerJSON("clarify", "Which?", nil, nil, nil), err: errAnswerNoChoices},
		{name: "an empty choice", content: answerJSON("clarify", "Which?", nil, nil, []string{"a", " "}), err: errAnswerNoChoices},
		{name: "a choice that is a paragraph", content: answerJSON("clarify", "Which?", nil, nil, []string{strings.Repeat("a", maxCandidateRunes+1)}), err: errAnswerNoChoices},
		{name: "too many choices", content: answerJSON("clarify", "Which?", nil, nil, slices.Repeat([]string{"a"}, maxAnswerCandidates+1)), err: errAnswerTooMany},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseStructuredAnswer(tt.content, tt.cards)
			if !errors.Is(err, tt.err) {
				t.Fatalf("error = %v, want %v", err, tt.err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("answer (-want +got):\n%s", diff)
			}
		})
	}
}

// TestTheAnswerSchemaIsTheStructsFields keeps the one definition one: the JSON
// Schema names exactly the fields of the answer's struct, all required, nothing
// else allowed, with the kinds the page knows.
func TestTheAnswerSchemaIsTheStructsFields(t *testing.T) {
	t.Parallel()
	var schema struct {
		Type       string                     `json:"type"`
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
		Extra      bool                       `json:"additionalProperties"`
	}
	if err := json.Unmarshal([]byte(AskAnswerSchema()), &schema); err != nil {
		t.Fatalf("the schema is not JSON: %v", err)
	}
	var fields []string
	for field := range reflect.TypeFor[structuredAnswer]().Fields() {
		fields = append(fields, strings.Split(field.Tag.Get("json"), ",")[0])
	}
	named := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		named = append(named, name)
	}
	slices.Sort(fields)
	slices.Sort(named)
	slices.Sort(schema.Required)
	if diff := cmp.Diff(fields, named); diff != "" {
		t.Errorf("fields against schema properties (-struct +schema):\n%s", diff)
	}
	if diff := cmp.Diff(fields, schema.Required); diff != "" {
		t.Errorf("fields against required (-struct +schema):\n%s", diff)
	}
	if schema.Type != "object" || schema.Extra {
		t.Errorf("schema = %+v, want an object that allows no other field", schema)
	}
	var kind struct {
		Enum []string `json:"enum"`
	}
	if err := json.Unmarshal(schema.Properties["kind"], &kind); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{answerKindAnswer, answerKindClarify, answerKindRefuse}, kind.Enum); diff != "" {
		t.Errorf("kinds (-want +got):\n%s", diff)
	}
}

func TestCardIDsRoundTrip(t *testing.T) {
	t.Parallel()
	for n := 1; n <= 12; n++ {
		if got, ok := cardNumber(cardID(n)); !ok || got != n {
			t.Errorf("cardNumber(cardID(%d)) = %d, %v", n, got, ok)
		}
	}
	for _, bad := range []string{"", "c", "c0", "c-1", "c1x", "x1", "c1000", "C1", " c1"} {
		if _, ok := cardNumber(bad); ok {
			t.Errorf("cardNumber(%q) is an id", bad)
		}
	}
}

// askConv asks question of the fixture and returns the card with the
// conversation, whose trace of the loop the tests read.
func (f *loopFixture) askConv(question string) (answerCard, *askConversation, error) {
	return f.a.converse(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ask", http.NoBody), session{Token: "t", Origin: "o"}, question)
}

func TestTheModelsToolResultsCarryTheIDsItNamesCardsBy(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{
		calling(askCall{ID: "1", Name: "value", Args: `{}`}, askCall{ID: "2", Name: "nope", Args: `{}`}, askCall{ID: "3", Name: "allocation", Args: `{}`}),
		say("x", "c2"),
	}
	if _, err := f.ask("q"); err != nil {
		t.Fatal(err)
	}
	results := toolMessages(f.port.requests[1])
	if len(results) != 3 || !strings.HasPrefix(results[0].Content, `{"id":"c1",`) || strings.Contains(results[1].Content, `"id"`) || !strings.HasPrefix(results[2].Content, `{"id":"c2",`) {
		t.Errorf("tool results = %+v, want ids on the two cards that ran, in order, and none on the failure", results)
	}
}

func TestTheAnswerShowsTheCardsItNamesInTheOrderItNamesThem(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{
		calling(askCall{ID: "1", Name: "allocation", Args: `{}`}, askCall{ID: "2", Name: "value", Args: `{}`}, askCall{ID: "3", Name: "show", Args: `{"name":"x"}`}),
		say("x", "c3", "c1", "c3"),
	}
	got, conv, err := f.askConv("q")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{fmt.Sprintf("Card %d", answer.CommandShow), fmt.Sprintf("Card %d", answer.CommandAllocation)}, titles(got.Used)); diff != "" {
		t.Errorf("cards shown (-want +got):\n%s", diff)
	}
	if len(conv.used) != 3 || !slices.Equal(conv.cited, []int{2, 0}) {
		t.Errorf("read %d cards and cited %v, want 3 and [2 0]", len(conv.used), conv.cited)
	}
}

func TestAnAnswerShowsNoCardItDoesNotName(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{calling(askCall{ID: "1", Name: "value", Args: `{}`}), say("I read it and name nothing.")}
	got, err := f.ask("q")
	if err != nil || len(got.Used) != 0 {
		t.Errorf("answer = %+v, %v; want no card: the model named none", got, err)
	}
}

func TestTheKindDecidesWhatThePageDraws(t *testing.T) {
	t.Parallel()
	choices, junk := []string{"Demo World ETF", "Demo Emerging Markets ETF"}, []string{"Write me a poem", "Show bonds"}
	commands := refusalSuggestions(mustDate(t, askToday))
	tests := []struct {
		name          string
		reply         askReply
		wantKind      string
		wantChoices   []string
		wantCommands  []string
		wantTextShown string
	}{
		{name: "an answer", reply: sayKind("answer", "It is 10.", nil, nil), wantKind: "answer", wantTextShown: "It is 10."},
		{name: "a clarifying answer shows its choices", reply: sayKind("clarify", "Which?", nil, choices), wantKind: "clarify", wantChoices: choices, wantTextShown: "Which?"},
		{name: "a refusal shows the box's commands", reply: sayKind("refuse", "No.", nil, nil), wantKind: "refuse", wantCommands: commands, wantTextShown: "No."},
		{name: "a refusal shows the box's commands, not the model's", reply: sayKind("refuse", "No.", junk, nil), wantKind: "refuse", wantCommands: commands, wantTextShown: "No."},
		{name: "an answer's choices and commands are not drawn", reply: sayKind("answer", "It is 10.", junk, choices), wantKind: "answer", wantTextShown: "It is 10."},
		{name: "a refusal's choices are not drawn", reply: sayKind("refuse", "No.", nil, choices), wantKind: "refuse", wantCommands: commands, wantTextShown: "No."},
		{name: "a clarifying answer's commands are not drawn", reply: sayKind("clarify", "Which?", junk, choices), wantKind: "clarify", wantChoices: choices, wantTextShown: "Which?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newLoopFixture(t, tierT0, tierT2)
			f.showing(choices...)
			f.port.replies = []askReply{calling(askCall{ID: "1", Name: "show", Args: `{"name":"etf"}`}), tt.reply}
			got, err := f.ask("q")
			if err != nil {
				t.Fatal(err)
			}
			if got.Kind != tt.wantKind || got.Text != tt.wantTextShown || !slices.Equal(got.Candidates, tt.wantChoices) || !slices.Equal(got.Suggestions, tt.wantCommands) {
				t.Errorf("answer = kind %q text %q candidates %v suggestions %v", got.Kind, got.Text, got.Candidates, got.Suggestions)
			}
		})
	}
}

func TestAReplyInARoundWithToolsThatIsAnAnswerEndsTheQuestionAtOnce(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{calling(askCall{ID: "1", Name: "value", Args: `{}`}), say("It is 10.", "c1")}
	got, conv, err := f.askConv("q")
	if err != nil || got.Text != "It is 10." || conv.rounds != 2 {
		t.Fatalf("answer = %+v, %v after %d calls, want the answer after two", got, err, conv.rounds)
	}
	// Neither call was held to the schema: the model was offered the tools and wrote the answer itself.
	if diff := cmp.Diff([]bool{false, false}, f.port.answers); diff != "" {
		t.Errorf("calls that asked for the answer (-want +got):\n%s", diff)
	}
}

func TestPlainTextInARoundWithToolsIsAskedForAgainAsTheAnswer(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{
		calling(askCall{ID: "1", Name: "value", Args: `{}`}),
		{Content: "Plain words.", Model: "p/m"},
		say("Plain words.", "c1"),
	}
	got, conv, err := f.askConv("q")
	if err != nil || got.Text != "Plain words." || len(got.Used) != 1 || conv.rounds != 3 || conv.invalid != 0 {
		t.Fatalf("answer = %+v, %v after %d calls and %d invalid; want the answer after three and none invalid", got, err, conv.rounds, conv.invalid)
	}
	// The third call is the final round: no tools, the structured answer asked for.
	if diff := cmp.Diff([]int{len(f.a.tools), len(f.a.tools), 0}, f.port.offered); diff != "" {
		t.Errorf("tools offered per call (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]bool{false, false, true}, f.port.answers); diff != "" {
		t.Errorf("calls that asked for the answer (-want +got):\n%s", diff)
	}
	// It was told its plain words were not the answer, with no data in the nudge.
	if diff := cmp.Diff([]string{"system:T2", "user:T0", "assistant:T0", "tool:T0", "assistant:T0", "user:T2"}, roleTiers(f.port.requests[2])); diff != "" {
		t.Errorf("messages of the final round (-want +got):\n%s", diff)
	}
	if last := f.port.requests[2][5]; last.Content != finalText {
		t.Errorf("the nudge = %q, want %q", last.Content, finalText)
	}
}

func TestAnInvalidFinalAnswerGetsOneRetryWithTheReason(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.a.limits = AskLimits{ToolRounds: 1, ToolCalls: 10, OutputTokens: 800, Deadline: DefaultAskLimits().Deadline}
	f.port.replies = []askReply{
		calling(askCall{ID: "1", Name: "value", Args: `{}`}),
		{Content: `{"kind":"answer","text":"x","cards":["c7"],"suggestions":[],"candidates":[]}`, Model: "p/m"},
		say("Mended.", "c1"),
	}
	got, conv, err := f.askConv("q")
	if err != nil || got.Text != "Mended." || conv.rounds != 3 || conv.invalid != 1 {
		t.Fatalf("answer = %+v, %v after %d calls and %d invalid; want the mended answer after three and one invalid", got, err, conv.rounds, conv.invalid)
	}
	if diff := cmp.Diff([]bool{false, true, true}, f.port.answers); diff != "" {
		t.Errorf("calls that asked for the answer (-want +got):\n%s", diff)
	}
	retry := f.port.requests[2]
	want := []askMessage{
		{Role: roleAssistant, Content: `{"kind":"answer","text":"x","cards":["c7"],"suggestions":[],"candidates":[]}`, Tier: tierT0},
		{Role: roleUser, Content: retryText(errAnswerUnknownCard), Tier: tierT2},
	}
	// The model's own reply, then the reason in fixed words, tagged public.
	if diff := cmp.Diff(want, retry[len(retry)-2:]); diff != "" {
		t.Errorf("the last two messages of the retry (-want +got):\n%s", diff)
	}
	if !strings.Contains(retryText(errAnswerUnknownCard), "a card you did not read") {
		t.Errorf("the reason = %q", retryText(errAnswerUnknownCard))
	}
}

func TestTwoInvalidAnswersAreAnErrorTheBoxCallsCouldNotAnswer(t *testing.T) {
	t.Parallel()
	plain := func(text string) askReply { return askReply{Content: text, Model: "p/m"} }
	tests := []struct {
		name      string
		toolRound int
		replies   []askReply
	}{
		{name: "plain text in the first round, then two invalid answers", toolRound: 5, replies: []askReply{plain("words"), plain("more words"), plain("{}")}},
		{name: "an invalid answer in the final round, then another", toolRound: 1, replies: []askReply{calling(askCall{ID: "1", Name: "value", Args: `{}`}), plain(""), plain("{")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newLoopFixture(t, tierT0, tierT2)
			f.a.limits = AskLimits{ToolRounds: tt.toolRound, ToolCalls: 10, OutputTokens: 800, Deadline: DefaultAskLimits().Deadline}
			f.port.replies = tt.replies
			got, conv, err := f.askConv("secret " + secretAsk)
			if err != nil {
				t.Fatal(err)
			}
			wantCouldNotAnswer(t, f, &got, conv)
		})
	}
}

func wantCouldNotAnswer(t *testing.T, f *loopFixture, got *answerCard, conv *askConversation) {
	t.Helper()
	if got.Title != "The model couldn't answer" || got.Model != "" || got.Source == "" || len(got.Help) == 0 || len(got.Used) != 0 {
		t.Errorf("answer = %+v, want the card of a model that could not answer, with the commands", got)
	}
	if strings.Contains(got.Text, secretAsk) || strings.Contains(f.logs.String(), secretAsk) {
		t.Error("the question reached the card's text or a log")
	}
	wantTwoInvalid(t, f, conv)
}

// wantTwoInvalid checks the loop's record of two invalid answers: an error, two
// invalid replies, three calls, the last two asking for the answer.
func wantTwoInvalid(t *testing.T, f *loopFixture, conv *askConversation) {
	t.Helper()
	if conv.schemaErr == nil || conv.invalid != 2 || conv.rounds != 3 {
		t.Errorf("schema error %v, %d invalid, %d calls; want an error, 2 and 3", conv.schemaErr, conv.invalid, conv.rounds)
	}
	if diff := cmp.Diff([]bool{false, true, true}, f.port.answers); diff != "" {
		t.Errorf("calls that asked for the answer (-want +got):\n%s", diff)
	}
}

func TestEveryCallThatAsksForTheAnswerOffersNoTools(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.replies = []askReply{
		calling(askCall{ID: "1", Name: "value", Args: `{}`}),
		{Content: "plain", Model: "p/m"},
		{Content: "still plain", Model: "p/m"},
		say("ok", "c1"),
	}
	if _, err := f.ask("q"); err != nil {
		t.Fatal(err)
	}
	for i, asked := range f.port.answers {
		if asked && f.port.offered[i] != 0 {
			t.Errorf("call %d asked for the answer and offered %d tools", i+1, f.port.offered[i])
		}
	}
}

func TestTheRetryIsSpentFromTheQuestionsTokens(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.port.used = 400
	f.port.replies = []askReply{{Content: "plain", Model: "p/m"}, {Content: "plain", Model: "p/m"}, {Content: "plain", Model: "p/m"}}
	got, err := f.ask("q")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]int64{800, 400}, f.port.maxTokens); diff != "" {
		t.Errorf("token limit of each call (-want +got):\n%s", diff)
	}
	if got.Title != "The model didn't finish" {
		t.Errorf("answer = %+v, want the card of a model that spent its tokens before the retry", got)
	}
}

func TestSuggestionsAreCommandsTheBoxUnderstands(t *testing.T) {
	t.Parallel()
	today := mustDate(t, askToday)
	for _, c := range answer.KnownCommands() {
		if !IsAskCommand(c.Example, today) {
			t.Errorf("the example %q of %q is no command", c.Example, c.Command)
		}
	}
	for _, text := range []string{"buy more bonds", "", "tell me a joke"} {
		if IsAskCommand(text, today) {
			t.Errorf("%q is a command", text)
		}
	}
	// The box offers these with a refusal itself, never the model's words.
	got := refusalSuggestions(today)
	if diff := cmp.Diff([]string{"What needs me?", "Allocation by currency", "Performance this year", "Value on 2026-01-31"}, got); diff != "" {
		t.Errorf("a refusal's suggestions (-want +got):\n%s", diff)
	}
	for _, s := range got {
		if !IsAskCommand(s, today) {
			t.Errorf("the suggestion %q is no command the box understands", s)
		}
	}
	if strings.Contains(askSystemPrompt(today), "suggestions") {
		t.Error("the prompt still asks the model for suggestions")
	}
}

func TestThePromptSaysWhatEachFieldIsFor(t *testing.T) {
	t.Parallel()
	prompt := askSystemPrompt(mustDate(t, askToday))
	for _, want := range []string{
		"JSON object", "kind, text, cards and candidates", `"answer"`, `"clarify"`, `"refuse"`,
		"never guess", "candidates", "the box adds the commands", "advice",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt lacks %q", want)
		}
	}
}

func TestTheClientAsksTheSchemaOnlyForTheAnswer(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.askModel.turnOn("T0", "T2")
	h.askModel.say(
		scriptedReply{calls: []*modelsv1.ToolCall{toolCall("value", `{}`)}},
		scriptedReply{text: "plain words"},
		scriptedReply{text: answerText("It is 10.", "c1")},
	)
	rec := h.asks(t, ownerSession, "how am I doing?")
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "It is 10.", "Answered by p/m")
	h.askModel.mu.Lock()
	defer h.askModel.mu.Unlock()
	if diff := cmp.Diff([]string{"", "", askAnswerSchemaJSON}, h.askModel.schemas); diff != "" {
		t.Errorf("response schema of each call (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([][]string{{"allocation", "performance", "show", "sync_status", "value", "what_needs_me"}, {"allocation", "performance", "show", "sync_status", "value", "what_needs_me"}, nil},
		[][]string{sortedCopy(h.askModel.toolsSeen[0]), sortedCopy(h.askModel.toolsSeen[1]), h.askModel.toolsSeen[2]}); diff != "" {
		t.Errorf("tools of each call (-want +got):\n%s", diff)
	}
}

// TestAskBoxDrawsEachKindOfAnswer checks the page for the three kinds of the
// model's structured answer: an answer with its cards, a clarifying answer whose
// candidates are choices that ask the box's own show command, and a refusal whose
// suggestions are command chips; and a reply that is no answer, as an error.
func TestAskBoxDrawsEachKindOfAnswer(t *testing.T) {
	t.Parallel()
	t.Run("a clarifying answer draws its candidates as choices", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		h.askModel.turnOn("T0", "T2")
		h.askModel.say(scriptedReply{calls: []*modelsv1.ToolCall{toolCall("show", `{"name":"Synthetic"}`)}},
			scriptedReply{text: answerJSON("clarify", "Which fund do you mean?", nil, nil, []string{"Synthetic Fund", "Synthetic brokerage", "Demo <b>Fund</b>"})})
		rec := h.asks(t, ownerSession, "I mean the fund")
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec,
			"Which fund do you mean?", "From p/m · it asks which you mean", `class="chips ask-choices"`, `aria-label="Which one do you mean"`,
			`href="/?q=Show&#43;Synthetic&#43;Fund#ask"`, `hx-get="/ask?q=Show&#43;Synthetic&#43;Fund"`, `>Synthetic Fund</a>`, `>Synthetic brokerage</a>`,
		)
		// A name no card the model read holds is dropped, and never drawn.
		wantNoBody(t, rec, `class="chips ask-suggestions"`, "Answered by", "Demo <b>Fund</b>", "Demo &lt;b&gt;Fund")
	})
	t.Run("a refusal draws the box's own commands as chips, not the model's", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		h.askModel.turnOn("T0", "T2")
		h.askModel.say(scriptedReply{text: answerJSON("refuse", "I can only read, so I can't sell anything.", nil, []string{"Write me a poem", "Allocation by currency"}, nil)})
		rec := h.asks(t, ownerSession, "sell everything")
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec,
			"I can only read, so I can&#39;t sell anything.", "From p/m · it can't do that", `class="chips ask-suggestions"`, `aria-label="Commands you can ask instead"`,
			`href="/?q=What&#43;needs&#43;me%3F#ask"`, `hx-get="/ask?q=Allocation&#43;by&#43;currency"`, `>Allocation by currency</a>`, `>Performance this year</a>`,
		)
		wantNoBody(t, rec, `class="chips ask-choices"`, "Answered by", `class="used-card"`, "Write me a poem")
	})
	t.Run("a refusal's chip is answered by the box itself, at once", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		h.askModel.turnOn("T0", "T2")
		rec := h.asks(t, ownerSession, "Allocation by currency")
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec, "Answered from your data · no AI model")
		if n := h.askModel.generated(); n != 0 {
			t.Errorf("the model was asked %d times for a chip", n)
		}
	})
	t.Run("a model that does not reply in the form is an error the page calls could not answer", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		h.askModel.turnOn("T0", "T2")
		h.askModel.say(scriptedReply{text: "plain"}, scriptedReply{text: "plain again"}, scriptedReply{text: "{}"})
		rec := h.asks(t, ownerSession, "how diversified am I?")
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec, "The model couldn&#39;t answer", "reply in the form the box needs", "The model gave no answer", "What needs me")
		wantNoBody(t, rec, "Answered by", "From p/m", "plain again")
	})
}

// A slip in a refusal's suggestions, which the box no longer asks of the model,
// neither fails the refusal nor forces another round (D-194): the commands are
// the box's own.
func TestARefusalWithAnythingInSuggestionsIsAcceptedAtOnceWithTheBoxsCommands(t *testing.T) {
	t.Parallel()
	for name, content := range map[string]string{
		"commands the box does not know": answerJSON("refuse", "I can't sell.", nil, []string{"Write me a poem"}, nil),
		"too many":                       answerJSON("refuse", "I can't sell.", nil, slices.Repeat([]string{"What needs me?"}, 30), nil),
		"the wrong type":                 `{"kind":"refuse","text":"I can't sell.","cards":[],"suggestions":"What needs me?","candidates":[]}`,
		"none":                           `{"kind":"refuse","text":"I can't sell.","cards":[],"candidates":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			f := newLoopFixture(t, tierT0, tierT2)
			f.port.replies = []askReply{{Content: content, Model: "p/m"}}
			got, conv, err := f.askConv("Sell all my DTEC and buy bonds")
			if err != nil || got.Kind != answerKindRefuse || conv.rounds != 1 || conv.invalid != 0 {
				t.Fatalf("answer = %+v, %v after %d calls and %d invalid; want the refusal at once", got, err, conv.rounds, conv.invalid)
			}
			if diff := cmp.Diff(refusalSuggestions(mustDate(t, askToday)), got.Suggestions); diff != "" {
				t.Errorf("suggestions (-want +got):\n%s", diff)
			}
		})
	}
}

// A clarifying answer offers names from the cards the model read: one no card
// holds is dropped, and an answer left with none is no answer, so the model is
// asked once more.
func TestAClarifyingAnswerKeepsOnlyTheNamesACardHolds(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	f.showing("Demo World ETF", "Demo Emerging Markets ETF")
	f.port.replies = []askReply{
		calling(askCall{ID: "1", Name: "show", Args: `{"name":"etf"}`}),
		sayKind("clarify", "Which?", nil, []string{"demo world etf", "Made Up Fund"}),
	}
	got, conv, err := f.askConv("etf")
	if err != nil || got.Kind != answerKindClarify || conv.invalid != 0 {
		t.Fatalf("answer = %+v, %v with %d invalid; want the clarifying answer", got, err, conv.invalid)
	}
	if diff := cmp.Diff([]string{"demo world etf"}, got.Candidates); diff != "" {
		t.Errorf("candidates (-want +got):\n%s", diff)
	}

	g := newLoopFixture(t, tierT0, tierT2)
	g.showing("Demo World ETF")
	g.port.replies = []askReply{
		calling(askCall{ID: "1", Name: "show", Args: `{"name":"etf"}`}),
		sayKind("clarify", "Which?", nil, []string{"Made Up Fund"}),
		sayKind("clarify", "Which?", nil, []string{"Demo World ETF"}),
	}
	got, conv, err = g.askConv("etf")
	if err != nil || conv.rounds != 3 || !slices.Equal(got.Candidates, []string{"Demo World ETF"}) {
		t.Fatalf("answer = %+v, %v after %d calls; want a third call, then the name the card holds", got, err, conv.rounds)
	}
	if again := g.port.requests[len(g.port.requests)-1]; again[len(again)-1].Content != finalText {
		t.Errorf("the model was told %q, want the final round's instruction %q", again[len(again)-1].Content, finalText)
	}
}
