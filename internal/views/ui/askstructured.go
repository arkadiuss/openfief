package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// The Ask box's model ends a question with a structured answer: a JSON object
// that says what kind of answer it is, what the person reads, which of the cards
// it read the answer rests on, which of the box's own commands the person could
// ask instead, and, when a name could be more than one thing, the names to choose
// from. The page draws from these fields and from nothing else the model wrote,
// so what it shows is what was checked.

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
	maxAnswerSuggestions = 6
	maxAnswerCandidates  = 8
	maxCandidateRunes    = 120
	maxAnswerCards       = 10
)

// askAnswerSchemaJSON is the JSON Schema of the structured answer, the one
// definition: strict structured output sends it as the response format, and JSON
// mode shows the shape in the prompt and the reply is checked against it.
const askAnswerSchemaJSON = `{"type":"object","properties":{` +
	`"kind":{"type":"string","enum":["answer","clarify","refuse"]},` +
	`"text":{"type":"string"},` +
	`"cards":{"type":"array","items":{"type":"string"}},` +
	`"suggestions":{"type":"array","items":{"type":"string"}},` +
	`"candidates":{"type":"array","items":{"type":"string"}}},` +
	`"required":["kind","text","cards","suggestions","candidates"],"additionalProperties":false}`

// structuredAnswer is the model's final answer as it writes it.
type structuredAnswer struct {
	Kind string `json:"kind"`
	// Text is what the person reads.
	Text string `json:"text"`
	// Cards are the ids of the cards the model read that the answer rests on.
	Cards []string `json:"cards"`
	// Suggestions are commands of the box the person could ask instead.
	Suggestions []string `json:"suggestions"`
	// Candidates are the names a clarifying answer asks the person to choose
	// from.
	Candidates []string `json:"candidates"`
}

// Why a reply is not a structured answer. None quotes the reply: it is the
// person's data (GR-12), and the reason goes back to the model.
var (
	errAnswerNotJSON     = errors.New("it is not one JSON object with the five fields")
	errAnswerKind        = errors.New("kind is none of answer, clarify and refuse")
	errAnswerNoText      = errors.New("text is empty")
	errAnswerUnknownCard = errors.New("cards names a card you did not read")
	errAnswerBadCommand  = errors.New("a suggestion is not one of the box's commands")
	errAnswerNoChoices   = errors.New("a clarifying answer lists the names to choose from in candidates")
	errAnswerTooMany     = errors.New("a list holds more than the box shows")
)

// parseStructuredAnswer reads content as a structured answer of a question
// asked on today after the model read cards cards, and checks it: the kind is
// one of the three, the text says something, every card named is one shown in
// this question, every suggestion is a command the box understands, a
// clarifying answer lists its candidates, and no list is longer than the page
// shows. A reply that breaks one is not an answer.
func parseStructuredAnswer(content string, cards int, today kernel.Date) (structuredAnswer, error) {
	var a structuredAnswer
	dec := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&a); err != nil || dec.More() {
		return structuredAnswer{}, errAnswerNotJSON
	}
	a.Text = strings.TrimSpace(a.Text)
	if err := a.check(cards, today); err != nil {
		return structuredAnswer{}, err
	}
	return a, nil
}

// check returns why a is not a valid answer, or nil.
func (a *structuredAnswer) check(cards int, today kernel.Date) error {
	switch a.Kind {
	case answerKindAnswer, answerKindClarify, answerKindRefuse:
	default:
		return errAnswerKind
	}
	switch {
	case a.Text == "":
		return errAnswerNoText
	case len(a.Cards) > maxAnswerCards || len(a.Suggestions) > maxAnswerSuggestions || len(a.Candidates) > maxAnswerCandidates:
		return errAnswerTooMany
	}
	return a.checkReferences(cards, today)
}

// checkReferences returns why what a points to is not valid, or nil: the cards
// it names are among the cards read, its suggestions are the box's commands, and
// its candidates are names, at least one for a clarifying answer.
func (a *structuredAnswer) checkReferences(cards int, today kernel.Date) error {
	switch {
	case slices.ContainsFunc(a.Cards, func(id string) bool { return !cardRead(id, cards) }):
		return errAnswerUnknownCard
	case slices.ContainsFunc(a.Suggestions, func(s string) bool { return !IsAskCommand(s, today) }):
		return errAnswerBadCommand
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
func cardID(n int) string { return "c" + strconv.Itoa(n) }

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
	_, ok := parseAsk(text, today)
	return ok
}

// AskAnswerSchema returns the JSON Schema the Ask box's final answer follows.
func AskAnswerSchema() string { return askAnswerSchemaJSON }

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

// answerExamples returns the commands the prompt offers the model as examples
// of what to suggest, each one the box understands.
func answerExamples() string {
	known := knownCommands()
	examples := make([]string, len(known))
	for i, k := range known {
		examples[i] = fmt.Sprintf("%q", k.Example)
	}
	return strings.Join(examples, ", ")
}
