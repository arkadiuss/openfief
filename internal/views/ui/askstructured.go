package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/app/answer"
)

// The Ask box's model ends a question with a structured answer: a JSON object
// that says what kind of answer it is, what the person reads, which of the cards
// it read the answer rests on and, when a name could be more than one thing, the
// names to choose from. The commands a refusal offers instead are the box's own,
// never the model's (D-194). The page draws from these fields and from nothing
// else the model wrote, so what it shows is what was checked.

// The kinds of an answer.
const (
	// answerKindAnswer answers from the cards the model read.
	answerKindAnswer = "answer"
	// answerKindClarify asks which of several things a name means.
	answerKindClarify = "clarify"
	// answerKindRefuse declines: a change, advice, or a question outside the
	// portfolio.
	answerKindRefuse = "refuse"
)

// The most a structured answer may hold, so a reply that rambles fails instead
// of filling the page.
const (
	maxAnswerCandidates = 8
	maxCandidateRunes   = 120
	maxAnswerCards      = 10
)

// askAnswerSchemaJSON is the JSON Schema of the structured answer, the one
// definition: strict structured output sends it as the response format, and JSON
// mode shows the shape in the prompt and the reply is checked against it. It has
// no suggestions: the box attaches its own commands to a refusal (D-194).
const askAnswerSchemaJSON = `{"type":"object","properties":{` +
	`"kind":{"type":"string","enum":["answer","clarify","refuse"]},` +
	`"text":{"type":"string"},` +
	`"cards":{"type":"array","items":{"type":"string"}},` +
	`"candidates":{"type":"array","items":{"type":"string"}}},` +
	`"required":["kind","text","cards","candidates"],"additionalProperties":false}`

// structuredAnswer is the model's final answer as it writes it.
type structuredAnswer struct {
	Kind string `json:"kind"`
	// Text is what the person reads.
	Text string `json:"text"`
	// Cards are the ids of the cards the model read that the answer rests on.
	Cards []string `json:"cards"`
	// Candidates are the names a clarifying answer asks the person to choose
	// from.
	Candidates []string `json:"candidates"`
}

// Why a reply is not a structured answer. None quotes the reply: it is the
// person's data (GR-12), and the reason goes back to the model.
var (
	errAnswerNotJSON     = errors.New("it is not one JSON object with the four fields")
	errAnswerKind        = errors.New("kind is none of answer, clarify and refuse")
	errAnswerNoText      = errors.New("text is empty")
	errAnswerUnknownCard = errors.New("cards names a card you did not read")
	errAnswerNoChoices   = errors.New("a clarifying answer lists the names to choose from in candidates")
	errAnswerTooMany     = errors.New("a list holds more than the box shows")
)

// parseStructuredAnswer reads content as a structured answer of a question
// after the model read cards cards, and checks it: the kind is one of the
// three, the text says something, every card named is one shown in this
// question, a clarifying answer lists its candidates, and no list is longer than
// the page shows. A reply that breaks one is not an answer. A "suggestions"
// member, whatever it holds, is ignored: the box offers the commands itself
// (D-194), so a slip in it never fails a correct reply.
func parseStructuredAnswer(content string, cards int) (structuredAnswer, error) {
	var members map[string]json.RawMessage
	dec := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	if err := dec.Decode(&members); err != nil || dec.More() {
		return structuredAnswer{}, errAnswerNotJSON
	}
	delete(members, answerSuggestionsMember)
	raw, err := json.Marshal(members)
	if err != nil {
		return structuredAnswer{}, errAnswerNotJSON
	}
	var a structuredAnswer
	strict := json.NewDecoder(bytes.NewReader(raw))
	strict.DisallowUnknownFields()
	if err := strict.Decode(&a); err != nil {
		return structuredAnswer{}, errAnswerNotJSON
	}
	a.Text = strings.TrimSpace(a.Text)
	if err := a.check(cards); err != nil {
		return structuredAnswer{}, err
	}
	return a, nil
}

// answerSuggestionsMember is the member older replies still write: the commands
// of a refusal, which the box now chooses (D-194) and the parser ignores.
const answerSuggestionsMember = "suggestions"

// check returns why a is not a valid answer, or nil.
func (a *structuredAnswer) check(cards int) error {
	switch a.Kind {
	case answerKindAnswer, answerKindClarify, answerKindRefuse:
	default:
		return errAnswerKind
	}
	switch {
	case a.Text == "":
		return errAnswerNoText
	case len(a.Cards) > maxAnswerCards || len(a.Candidates) > maxAnswerCandidates:
		return errAnswerTooMany
	}
	return a.checkReferences(cards)
}

// checkReferences returns why what a points to is not valid, or nil: the cards
// it names are among the cards read, and its candidates are names, at least one
// for a clarifying answer.
func (a *structuredAnswer) checkReferences(cards int) error {
	switch {
	case slices.ContainsFunc(a.Cards, func(id string) bool { return !cardRead(id, cards) }):
		return errAnswerUnknownCard
	case slices.ContainsFunc(a.Candidates, notAName), a.Kind == answerKindClarify && len(a.Candidates) == 0:
		return errAnswerNoChoices
	}
	return nil
}

// notAName reports whether a candidate is empty or too long to be a name.
func notAName(c string) bool {
	return strings.TrimSpace(c) == "" || utf8.RuneCountInString(c) > maxCandidateRunes
}

// cardRead reports whether id names one of the first n cards of the question.
func cardRead(id string, n int) bool {
	number, ok := cardNumber(id)
	return ok && number <= n
}

// cardID is the id a model reads on the n-th card (counting from one) its tools
// returned in this question, and names in its answer.
func cardID(n int) string { return answer.CardID(n) }

// cardNumber returns the number a card id names, ok false for an id that is not
// c1, c2 and so on.
func cardNumber(id string) (int, bool) {
	rest, ok := strings.CutPrefix(id, "c")
	if !ok || rest == "" || rest[0] == '0' || len(rest) > 3 {
		return 0, false
	}
	n, err := strconv.Atoi(rest)
	return n, err == nil && n > 0
}

// IsAskCommand reports whether text is one of the commands the Ask box
// understands on today (D-194): the box answers it itself, at once, without a
// model.
func IsAskCommand(text string, today kernel.Date) bool {
	_, ok := answer.Parse(text, today)
	return ok
}

// AskAnswerSchema returns the JSON Schema the Ask box's final answer follows.
func AskAnswerSchema() string { return askAnswerSchemaJSON }

// AskRefusalSuggestions returns the commands the box offers with a refusal on
// today, the ones a harness expects to find on a refusing answer (D-194).
func AskRefusalSuggestions(today kernel.Date) []string { return refusalSuggestions(today) }

// AskStructuredAnswer is a structured answer as a harness reads it.
type AskStructuredAnswer = structuredAnswer

// retryText is what the person's model is told when its reply is not an answer:
// the fixed reason and what to send, nothing of the reply. It fixes the form,
// not the content (D-205, D-217): the retry keeps the reply's meaning.
func retryText(err error) string {
	return "Your reply was not accepted: " + err.Error() +
		". Fix only its form: reply again with only the JSON object, without changing its meaning or adding anything."
}

// finalText is what the model is told when it answered in plain text, or in a
// reply that was no valid answer, while the box wants the structured answer. It
// asks for the same reply in the schema's form and nothing new (D-205, D-217): a
// model told to "reply now" had turned a correct refusal into a greeting.
const finalText = "Put your previous reply into this schema without changing its meaning or adding anything."

// finalAnswerText is what the model is told when its last reply was empty (it
// only called tools, or stopped after a tool's result without a word): there is
// no reply to restate, and asking for one made models refuse. It asks for the
// answer itself, in the schema, from the cards read.
const finalAnswerText = "Answer the user's question in this schema, using only the cards above."

// refusalSuggestions are the commands the box offers with a refusal (D-194):
// the examples of its own commands that need no name from the person's record,
// each one it understands, in the order of knownCommands. They are the box's, not
// the model's, so a refusal always has them and never a command the box doesn't
// answer. The commands are English: the matcher understands no other language, so
// a Polish chip would only be answered with the commands.
func refusalSuggestions(today kernel.Date) []string {
	var out []string
	for _, k := range answer.KnownCommands() {
		if q, ok := answer.Parse(k.Example, today); ok && q.Command != answer.CommandShow && q.Command != answer.CommandSync {
			out = append(out, k.Example)
		}
	}
	return out
}
