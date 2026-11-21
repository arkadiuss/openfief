package ui

import (
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// opinionsJSON is the structured answer a model writes with opinions on.
func opinionsJSON(kind, text string, cards []string, t *structuredThesis) string {
	if cards == nil {
		cards = []string{}
	}
	if t == nil {
		t = &structuredThesis{Instruments: []string{}, Invalidation: []string{}, Sources: []string{}}
	}
	b, err := json.Marshal(struct {
		Kind       string           `json:"kind"`
		Text       string           `json:"text"`
		Cards      []string         `json:"cards"`
		Candidates []string         `json:"candidates"`
		Thesis     structuredThesis `json:"thesis"`
	}{kind, text, cards, []string{}, *t})
	if err != nil {
		panic(err)
	}
	return string(b)
}

// goodThesis is a thesis the box accepts when a card names EXMP.
func goodThesis() *structuredThesis {
	return &structuredThesis{
		Statement: "Services revenue keeps growing faster than hardware.", Instruments: []string{"EXMP"}, Horizon: horizonLongTerm,
		Invalidation: []string{"Services growth falls below hardware growth for two quarters."},
		Sources:      []string{"The holdings card, Example Corp.", "The market price card for EXMP."},
	}
}

// errCoreDown is the error of a core that can't be reached.
var errCoreDown = errors.New("core down")

// opinionsFixture is a loop over a route that may receive T0, T1 and T2, whose
// cards name EXMP and OTHR, with opinions in state.
func opinionsFixture(t *testing.T, state opinionsState) *loopFixture {
	t.Helper()
	f := newLoopFixture(t, tierT0, tierT1, tierT2)
	f.showing("Example Corp (EXMP)", "Other Corp (OTHR)")
	f.port.opinionsAre = state
	return f
}

// readThenSay is a script: the model reads a card, then answers with content.
func readThenSay(content string) []askReply {
	return []askReply{
		calling(askCall{ID: "1", Name: "value", Args: `{}`}),
		{Content: content, Model: "p/m"},
	}
}

// thesisReplies is a script: the model reads a card, answers in plain text, so the box
// asks for the structured answer, and then suggests a thesis.
func thesisReplies() []askReply {
	return []askReply{
		calling(askCall{ID: "1", Name: "value", Args: `{}`}),
		{Content: "Example Corp looks interesting.", Model: "p/m"},
		{Content: opinionsJSON(answerKindThesis, "A thesis from your holdings.", []string{"c1"}, goodThesis()), Model: "p/m"},
	}
}

func TestWithOpinionsOnTheModelSuggestsAThesisTheBoxShowsAsACard(t *testing.T) {
	t.Parallel()
	f := opinionsFixture(t, opinionsOn)
	f.port.replies = thesisReplies()
	got, err := f.ask("suggest a thesis on EXMP")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != answerKindThesis || !got.Opinions || got.Thesis == nil {
		t.Fatalf("card = kind %q, opinions %v, thesis %v; want a thesis answer given with opinions on", got.Kind, got.Opinions, got.Thesis)
	}
	want := &thesisCard{
		Statement: goodThesis().Statement, Horizon: "Long-term", HorizonKey: horizonLongTerm, Instruments: []string{"EXMP"},
		Invalidation: goodThesis().Invalidation, Sources: goodThesis().Sources,
	}
	if diff := cmp.Diff(want, got.Thesis); diff != "" {
		t.Errorf("the thesis card (-want +got):\n%s", diff)
	}
	if len(got.Used) != 1 {
		t.Errorf("the answer shows %d cards, want the one it names", len(got.Used))
	}
}

// onlySchema fails t if a call asked for a schema other than want.
func onlySchema(t *testing.T, schemas []string, want string) {
	t.Helper()
	for i, schema := range schemas {
		if schema != "" && schema != want {
			t.Errorf("call %d asked for a schema other than the expected one", i)
		}
	}
}

func TestWithOpinionsOnTheModelGetsTheOpinionsPromptAndSchemaAndT1IsChecked(t *testing.T) {
	t.Parallel()
	f := opinionsFixture(t, opinionsOn)
	f.port.replies = thesisReplies()
	if _, err := f.ask("suggest a thesis on EXMP"); err != nil {
		t.Fatal(err)
	}
	system := f.port.requests[0][0]
	if system.Role != roleSystem || system.Tier != tierInstructions || system.Content != askSystemPromptFor(mustDate(t, askToday), true) {
		t.Errorf("the first message = %+v, want the opinions prompt, which names no data", system)
	}
	onlySchema(t, f.port.schemas, askOpinionsSchemaJSON)
	if !slices.Contains(f.port.schemas, askOpinionsSchemaJSON) {
		t.Errorf("no call asked for the opinions schema: %v", f.port.schemas)
	}
	// The thesis is T1 data in the conversation's tiers: the route is checked for it with the question and the card.
	for _, tiers := range f.port.permitted {
		if !slices.Contains(tiers, tierT1) {
			t.Errorf("a card was checked for tiers %v, want T1 among them in an opinions question", tiers)
		}
	}
}

func TestWithOpinionsOffTheModelKeepsTheR7PromptAndSchema(t *testing.T) {
	t.Parallel()
	f := opinionsFixture(t, opinionsOff)
	f.port.replies = readThenSay(answerJSON(answerKindRefuse, "I can't recommend what to buy.", nil, nil, nil))
	got, err := f.ask("should I buy EXMP?")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != answerKindRefuse || got.Opinions || got.Thesis != nil || got.OpinionsNote != "" {
		t.Errorf("card = %+v, want the plain refusal of R-7 with no opinions note", got)
	}
	if first := f.port.requests[0][0].Content; first != askSystemPrompt(mustDate(t, askToday)) || !strings.Contains(first, "advice on what to buy, sell or hold") {
		t.Errorf("the prompt is not the one that declines advice")
	}
	onlySchema(t, f.port.schemas, askAnswerSchemaJSON)
}

func TestAThesisAnswerIsNoAnswerWithOpinionsOff(t *testing.T) {
	t.Parallel()
	// It is retried, then the box says the model did not reply in its form.
	f := opinionsFixture(t, opinionsOff)
	f.port.replies = slices.Repeat(readThenSay(opinionsJSON(answerKindThesis, "t", []string{"c1"}, goodThesis()))[1:], 3)
	got, err := f.ask("suggest a thesis on EXMP")
	if err != nil {
		t.Fatal(err)
	}
	if got.Thesis != nil || got.Kind == answerKindThesis || got.Title != "The model couldn't answer" {
		t.Errorf("a thesis with opinions off became %+v, want the box's own failure card", got)
	}
}

func TestAThesisMemberOnAPlainAnswerIsIgnoredWithOpinionsOffAndDroppedWithThemOn(t *testing.T) {
	t.Parallel()
	for _, state := range []opinionsState{opinionsOff, opinionsOn} {
		f := opinionsFixture(t, state)
		f.port.replies = readThenSay(opinionsJSON(answerKindAnswer, "You hold Example Corp.", []string{"c1"}, goodThesis()))
		got, err := f.ask("what do I hold?")
		if err != nil {
			t.Fatal(err)
		}
		if got.Kind != answerKindAnswer || got.Thesis != nil {
			t.Errorf("state %v: card = kind %q thesis %v, want a plain answer with no thesis shown", state, got.Kind, got.Thesis)
		}
	}
}

func TestARouteThatKeepsOpinionsOffSaysSoOnARefusalOnly(t *testing.T) {
	t.Parallel()
	f := opinionsFixture(t, opinionsIneligible)
	f.port.replies = readThenSay(answerJSON(answerKindRefuse, "I can't say whether to buy.", nil, nil, nil))
	got, err := f.ask("should I buy EXMP?")
	if err != nil {
		t.Fatal(err)
	}
	if got.OpinionsNote != opinionsIneligibleNote || got.Opinions {
		t.Errorf("card = note %q, opinions %v; want the note that says why and no opinions", got.OpinionsNote, got.Opinions)
	}
	if first := f.port.requests[0][0].Content; first != askSystemPrompt(mustDate(t, askToday)) {
		t.Error("an ineligible route's model got a prompt that allows opinions")
	}
	f = opinionsFixture(t, opinionsIneligible)
	f.port.replies = readThenSay(answerText("You hold Example Corp.", "c1"))
	if got, err = f.ask("what do I hold?"); err != nil || got.OpinionsNote != "" {
		t.Errorf("an ordinary answer = %+v, %v; want no note about opinions", got, err)
	}
}

func TestAnUnreachableCoreMeansNoOpinions(t *testing.T) {
	t.Parallel()
	f := opinionsFixture(t, opinionsOn)
	f.port.opinionsErr = errCoreDown
	f.port.replies = readThenSay(answerJSON(answerKindRefuse, "I can't recommend.", nil, nil, nil))
	got, err := f.ask("should I buy EXMP?")
	if err != nil {
		t.Fatal(err)
	}
	if got.Opinions || f.port.requests[0][0].Content != askSystemPrompt(mustDate(t, askToday)) {
		t.Error("opinions were given on a guess: the box never gives one without the core's word")
	}
}

func TestOpinionsPromptAllowsOpinionsAndNothingMore(t *testing.T) {
	t.Parallel()
	day := mustDate(t, askToday)
	off, on := askSystemPromptFor(day, false), askSystemPromptFor(day, true)
	if off != askSystemPrompt(day) {
		t.Error("the prompt with opinions off is not the plain prompt")
	}
	for _, want := range []string{
		"five fields: kind, text, cards, candidates and thesis", `"thesis" when you suggest a thesis`,
		"Refuse only these two things:", "buying or selling for the owner, starting an agent or a fief, or leasing money",
		"only from the cards you read and the prices the market price tool returned", "never from memory",
		"an AI model's and not advice", "You cannot act", "answer with kind \"refuse\" and say you can only give an opinion",
		"only instruments a card you read names", "\"long_term\" or \"short_term\"", "never holds an amount, a lease size, a share of the portfolio or an order",
		"never your opinion or a thesis", askPromptCanary,
	} {
		if !strings.Contains(on, want) {
			t.Errorf("the opinions prompt lacks %q", want)
		}
	}
	for _, gone := range []string{"four fields", "Refuse only these three things", "advice on what to buy, sell or hold"} {
		if strings.Contains(on, gone) {
			t.Errorf("the opinions prompt still says %q", gone)
		}
	}
	// A public demo's and the default prompt keep declining advice, whatever the other says.
	for _, want := range []string{"advice on what to buy, sell or hold", "four fields"} {
		if !strings.Contains(off, want) {
			t.Errorf("the plain prompt lacks %q", want)
		}
	}
	if strings.Contains(off, "thesis") {
		t.Error("the plain prompt mentions a thesis: a demo's model is told nothing of it")
	}
}

func TestNoneOfTheOpinionsSentencesIsPublic(t *testing.T) {
	t.Parallel()
	day := mustDate(t, askToday)
	// None of the opinions sentences is public: the box never says its instructions.
	if got, plain := askPromptNonPublicFor(day, true), askPromptNonPublic(day); len(got) <= len(plain) {
		t.Errorf("%d non-public sentences with opinions on, %d without: the opinions sentences are not all non-public", len(got), len(plain))
	}
	for _, p := range askPromptPartsFor(day, true) {
		if p.Public && p.Text != askPromptParts(day)[0].Text {
			t.Errorf("a public sentence the plain prompt lacks: %q", p.Text)
		}
	}
}

// schemaMatches fails t unless the object schema raw names, and requires, exactly
// the JSON fields of typ.
func schemaMatches(t *testing.T, name string, raw json.RawMessage, typ reflect.Type) {
	t.Helper()
	var schema struct {
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
		Extra      bool                       `json:"additionalProperties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("the %s schema is not JSON: %v", name, err)
	}
	fields := jsonFieldsOf(typ, "")
	named := slices.Collect(maps.Keys(schema.Properties))
	slices.Sort(fields)
	slices.Sort(named)
	slices.Sort(schema.Required)
	if diff := cmp.Diff(fields, named); diff != "" {
		t.Errorf("%s fields against properties (-struct +schema):\n%s", name, diff)
	}
	if diff := cmp.Diff(fields, schema.Required); diff != "" {
		t.Errorf("%s fields against required (-struct +schema):\n%s", name, diff)
	}
	if schema.Extra {
		t.Errorf("the %s schema allows other fields", name)
	}
}

func TestOpinionsSchemaIsTheStructsFields(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(AskAnswerSchemaFor(true))
	schemaMatches(t, "answer", raw, reflect.TypeFor[structuredAnswer]())
	var schema struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	schemaMatches(t, "thesis", schema.Properties["thesis"], reflect.TypeFor[structuredThesis]())
}

func TestTheThesisSchemaHasNoFieldForALeaseSizeAndFitsTheCoresLimit(t *testing.T) {
	t.Parallel()
	fields := jsonFieldsOf(reflect.TypeFor[structuredThesis](), "")
	for _, banned := range []string{"lease", "amount", "size", "budget", "order", "weight"} {
		if slices.ContainsFunc(fields, func(f string) bool { return strings.Contains(f, banned) }) {
			t.Errorf("the thesis has a field naming %q: it has no lease size or amount", banned)
		}
	}
	const coresLimit = 4096 // internal/modelrouter/connectapi's maxAskSchemaBytes
	if len(askOpinionsSchemaJSON) > coresLimit {
		t.Errorf("the opinions schema is %d bytes, over the core's limit of %d", len(askOpinionsSchemaJSON), coresLimit)
	}
	if AskAnswerSchemaFor(false) != askAnswerSchemaJSON {
		t.Error("the schema with opinions off is not the plain one")
	}
}

func TestAThesisSuggestionIsCheckedAgainstTheCardsTheModelRead(t *testing.T) {
	t.Parallel()
	const cards = "card 1\nexample corp (exmp)\nother corp (othr)\n"
	long := strings.Repeat("x", maxThesisStatementRunes+1)
	tests := []struct {
		name string
		edit func(t *structuredThesis)
		want error
	}{
		{"a complete thesis", func(*structuredThesis) {}, nil},
		{"an instrument in another case", func(t *structuredThesis) { t.Instruments = []string{"Exmp"} }, nil},
		{"two instruments the cards name", func(t *structuredThesis) { t.Instruments = []string{"EXMP", "OTHR"} }, nil},
		{"an instrument no card names", func(t *structuredThesis) { t.Instruments = []string{"EXMP", "MADEUP"} }, errThesisInstrument},
		{"no statement", func(t *structuredThesis) { t.Statement = "  " }, errThesisFields},
		{"a statement too long", func(t *structuredThesis) { t.Statement = long }, errThesisFields},
		{"no instrument", func(t *structuredThesis) { t.Instruments = nil }, errThesisFields},
		{"too many instruments", func(t *structuredThesis) {
			t.Instruments = nil
			for i := range maxThesisInstruments + 1 {
				t.Instruments = append(t.Instruments, "EXMP"+strings.Repeat("x", i))
			}
		}, errThesisFields},
		{"no horizon", func(t *structuredThesis) { t.Horizon = "" }, errThesisFields},
		{"a horizon that is neither", func(t *structuredThesis) { t.Horizon = "medium_term" }, errThesisFields},
		{"short term", func(t *structuredThesis) { t.Horizon = horizonShortTerm }, nil},
		{"no invalidation condition", func(t *structuredThesis) { t.Invalidation = nil }, errThesisFields},
		{"no source", func(t *structuredThesis) { t.Sources = []string{" "} }, errThesisFields},
		{"a control character", func(t *structuredThesis) { t.Sources = []string{"a\x07b"} }, errThesisText},
		{"a lease size in the statement", func(t *structuredThesis) { t.Statement = "Put 5000 PLN into EXMP." }, errThesisSize},
		{"a share of the portfolio", func(t *structuredThesis) { t.Invalidation = []string{"Hold 5% of the portfolio."} }, errThesisSize},
		{"the word lease", func(t *structuredThesis) { t.Sources = []string{"A lease of one tenth."} }, errThesisSize},
		{"a percentage that is no share", func(t *structuredThesis) { t.Statement = "Revenue grew 12% in 2026." }, nil},
		{"a price quoted from a card", func(t *structuredThesis) { t.Sources = []string{"The market price card: closed at 46.15 USD."} }, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			th := goodThesis()
			tt.edit(th)
			th.tidy()
			if err := th.check(cards); !errors.Is(err, tt.want) {
				t.Errorf("check() = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestAThesisAnswerNeedsAThesisAndACardAndNothingQuotesTheReply(t *testing.T) {
	t.Parallel()
	read := []answerCard{{Title: "Holdings", Rows: []answerRow{{Label: "Example Corp (EXMP)"}}}}
	conv := &askConversation{opinions: opinionsOn, used: read}
	tests := []struct {
		name    string
		content string
		want    error
	}{
		{"a thesis with a card", opinionsJSON(answerKindThesis, "Here.", []string{"c1"}, goodThesis()), nil},
		{"a thesis naming no card", opinionsJSON(answerKindThesis, "Here.", nil, goodThesis()), errThesisMissing},
		{"a thesis answer with an empty thesis", opinionsJSON(answerKindThesis, "Here.", []string{"c1"}, nil), errThesisMissing},
		{"a thesis naming an instrument no card names", opinionsJSON(answerKindThesis, "Here.", []string{"c1"}, func() *structuredThesis {
			th := goodThesis()
			th.Instruments = []string{"GHOST"}
			return th
		}()), errThesisInstrument},
		{"an unknown kind", opinionsJSON("recommend", "Buy.", nil, nil), errAnswerKind},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, err := conv.parse(tt.content)
			if !errors.Is(err, tt.want) {
				t.Fatalf("parse() error = %v, want %v", err, tt.want)
			}
			if err != nil && strings.Contains(err.Error(), "GHOST") {
				t.Errorf("the error quotes the reply: %v", err)
			}
			if err == nil && a.Thesis == nil {
				t.Error("a valid thesis answer has no thesis")
			}
		})
	}
}

func TestAThesisInTheHistoryIsSummarisedForTheNextQuestion(t *testing.T) {
	t.Parallel()
	card := answerCard{Question: "suggest a thesis", Model: "p/m", Kind: answerKindThesis, Text: "Here is one.", Thesis: cardOfThesis(goodThesis())}
	turn, ok := turnOf(&card)
	if !ok {
		t.Fatal("a thesis answer adds no turn")
	}
	for _, want := range []string{"Here is one.", "Suggested thesis (Long-term)", "EXMP", "Wrong if", "market price card for EXMP"} {
		if !strings.Contains(turn.Answer, want) {
			t.Errorf("the turn's answer lacks %q: %q", want, turn.Answer)
		}
	}
}

func TestEveryModelAnswerSaysItIsNotAdviceAndOpinionsSaySMore(t *testing.T) {
	t.Parallel()
	if got := AskNote(false); got != "Not investment advice." {
		t.Errorf("AskNote(false) = %q", got)
	}
	got := AskNote(true)
	for _, want := range []string{"Not investment advice. ", "AI model's opinion", "never acts"} {
		if !strings.Contains(got, want) {
			t.Errorf("AskNote(true) = %q, lacks %q", got, want)
		}
	}
	if !strings.Contains(AskOpinionsIneligibleNote(), "Allow one on the Models page.") {
		t.Error("the ineligible note does not say what to do")
	}
	if !strings.Contains(thesisCardNote, "not investment advice") || !strings.Contains(thesisCardNote, "no lease size") {
		t.Error("the thesis card's note does not say it is not advice and holds no lease size")
	}
}

func TestAskCheckThesisIsTheLoopsCheck(t *testing.T) {
	t.Parallel()
	cards := []AskRunCard{{Title: "Holdings", Rows: []AskRunRow{{Label: "Example Corp (EXMP)"}}}}
	good := &AskThesis{
		Statement: "Services keep growing.", Instruments: []string{"exmp"}, Horizon: horizonShortTerm,
		Invalidation: []string{"Growth stalls."}, Sources: []string{"The holdings card."},
	}
	if err := AskCheckThesis(good, cards); err != nil {
		t.Errorf("AskCheckThesis(good) = %v", err)
	}
	bad := *good
	bad.Instruments = []string{"GHOST"}
	if err := AskCheckThesis(&bad, cards); !errors.Is(err, errThesisInstrument) {
		t.Errorf("AskCheckThesis(an instrument no card names) = %v", err)
	}
	if !AskThesisMentionsSize("a lease of 10% of the portfolio") || AskThesisMentionsSize("EXMP") {
		t.Error("AskThesisMentionsSize does not tell a lease size from an instrument")
	}
	if l := AskThesisLimitsOf(); l.Instruments != 10 || l.StatementRunes != 1000 {
		t.Errorf("AskThesisLimitsOf() = %+v", l)
	}
}
