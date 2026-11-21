package ui

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"connectrpc.com/connect"
)

// With opinions on (D-225), the Ask box's model may suggest a thesis (D-229): a
// statement, its instruments, a horizon, the conditions that would invalidate
// it and its sources, and no lease size. The page shows it as a card the owner
// may save as a draft thesis. The checks here are the box's own, applied to
// what the model wrote before the page draws it; the core checks the draft again
// when it is saved (the thesis context's own rules, of which these are a copy
// that tools/askbench's tests keep in step), and a draft is only ever text: it
// has no agent, no fief and no lease.

// answerKindThesis is a thesis suggestion, a kind of answer only with opinions on.
const answerKindThesis = "thesis"

// The horizons of a thesis suggestion, as the contract spells them.
const (
	horizonLongTerm  = "long_term"
	horizonShortTerm = "short_term"
)

// The limits of a thesis suggestion, in characters and items: the thesis
// context's own.
const (
	maxThesisStatementRunes    = 1000
	maxThesisInstruments       = 10
	maxThesisInstrumentRunes   = 60
	maxThesisInvalidation      = 8
	maxThesisInvalidationRunes = 300
	maxThesisSources           = 10
	maxThesisSourceRunes       = 200
)

// askOpinionsSchemaJSON is the JSON Schema of the structured answer with
// opinions on: askAnswerSchemaJSON with the kind "thesis" and a thesis member
// that is empty for every other kind. It names no lease size or amount: the
// thesis has no field for one.
const askOpinionsSchemaJSON = `{"type":"object","properties":{` +
	`"kind":{"type":"string","enum":["answer","clarify","refuse","thesis"]},` +
	`"text":{"type":"string"},` +
	`"cards":{"type":"array","items":{"type":"string"}},` +
	`"candidates":{"type":"array","items":{"type":"string"}},` +
	`"thesis":{"type":"object","properties":{` +
	`"statement":{"type":"string"},` +
	`"instruments":{"type":"array","items":{"type":"string"}},` +
	`"horizon":{"type":"string","enum":["","long_term","short_term"]},` +
	`"invalidation":{"type":"array","items":{"type":"string"}},` +
	`"sources":{"type":"array","items":{"type":"string"}}},` +
	`"required":["statement","instruments","horizon","invalidation","sources"],"additionalProperties":false}},` +
	`"required":["kind","text","cards","candidates","thesis"],"additionalProperties":false}`

// opinionsState is whether the Ask box may give opinions and suggest a thesis in
// a question (D-225, D-229), as the core says it.
type opinionsState int

// The states of opinions.
const (
	// opinionsOff: the box declines to recommend (R-7) and says nothing more. It
	// is every question's state unless the core says otherwise: opinions off,
	// the person not the owner, a demo, the switch off, or the core unreachable.
	opinionsOff opinionsState = iota
	// opinionsOn: the model may give opinions and suggest a thesis.
	opinionsOn
	// opinionsIneligible: the owner turned opinions on but no model of the route
	// may receive T0, T1 and T2 together; they stay off and the box says why.
	opinionsIneligible
)

// opinionsOf asks the core whether the model may give opinions in a question of
// sess. Anything but an answer is off: the box never gives an opinion on a guess.
func (a *modelAnswerer) opinionsOf(ctx context.Context, sess session) opinionsState {
	state, err := a.model.opinions(ctx, sess)
	if err != nil {
		a.log.WarnContext(ctx, "ask opinions check failed", "code", connect.CodeOf(err).String())
		return opinionsOff
	}
	return state
}

// checkThesis returns why a thesis answer is not valid, or nil. A "thesis"
// answer names at least one card it read and suggests a thesis the box accepts
// (structuredThesis.check) whose instruments the cards it read name; every
// other kind's thesis is dropped, since the page shows only a thesis answer's.
func (a *structuredAnswer) checkThesis(read []answerCard) error {
	if a.Kind != answerKindThesis {
		a.Thesis = nil
		return nil
	}
	if a.Thesis.empty() || len(a.Cards) == 0 {
		return errThesisMissing
	}
	a.Thesis.tidy()
	return a.Thesis.check(cardsHaystack(read))
}

// structuredThesis is the thesis a model suggests, as it writes it.
type structuredThesis struct {
	Statement    string   `json:"statement"`
	Instruments  []string `json:"instruments"`
	Horizon      string   `json:"horizon"`
	Invalidation []string `json:"invalidation"`
	Sources      []string `json:"sources"`
}

// Why a thesis suggestion is not accepted. None quotes the reply (GR-12).
var (
	errThesisMissing    = errors.New("kind thesis needs thesis filled in, and cards naming the cards you read")
	errThesisFields     = errors.New("thesis needs a statement, 1 to 10 instruments, a horizon of long_term or short_term, 1 to 8 invalidation conditions and 1 to 10 sources, each within its length")
	errThesisInstrument = errors.New("thesis names an instrument that no card you read names")
	errThesisSize       = errors.New("a thesis holds no amount, lease size, share of the portfolio or order")
	errThesisText       = errors.New("thesis holds a control character")
)

// empty reports whether t holds nothing: what every kind but "thesis" carries.
func (t *structuredThesis) empty() bool {
	return t == nil || (strings.TrimSpace(t.Statement) == "" && t.Horizon == "" && len(t.Instruments) == 0 && len(t.Invalidation) == 0 && len(t.Sources) == 0)
}

// tidy trims the thesis's text and drops an item that is empty or an
// instrument a line before names, the way the thesis context tidies a draft.
func (t *structuredThesis) tidy() {
	t.Statement = strings.TrimSpace(t.Statement)
	t.Instruments = dedupeFold(tidyItems(t.Instruments))
	t.Invalidation = tidyItems(t.Invalidation)
	t.Sources = tidyItems(t.Sources)
}

// tidyItems trims each item, makes each run of white space one space and drops
// the empty ones.
func tidyItems(in []string) []string {
	var out []string
	for _, s := range in {
		if s = strings.Join(strings.Fields(s), " "); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// dedupeFold drops an item that an earlier one equals without regard to case.
func dedupeFold(in []string) []string {
	var out []string
	for _, s := range in {
		if !slices.ContainsFunc(out, func(o string) bool { return strings.EqualFold(o, s) }) {
			out = append(out, s)
		}
	}
	return out
}

// check returns why t is not a thesis the box shows, or nil: every field there
// and within its limit, a horizon of the two, no control character, no amount or
// lease size in any of its text, and every instrument one that the text of a
// card the model read names (cards, in lower case, joined).
func (t *structuredThesis) check(cards string) error {
	if t.fieldsOutside() {
		return errThesisFields
	}
	text := append(slices.Concat(t.Instruments, t.Invalidation, t.Sources), t.Statement)
	if slices.ContainsFunc(text, hasControlRune) {
		return errThesisText
	}
	if slices.ContainsFunc(text, thesisMentionsSize) {
		return errThesisSize
	}
	for _, inst := range t.Instruments {
		if !strings.Contains(cards, strings.ToLower(inst)) {
			return errThesisInstrument
		}
	}
	return nil
}

// fieldsOutside reports whether a field of t is missing or beyond its limit: no
// statement or one too long, a horizon that is neither of the two, and a list with
// none, too many or an item too long.
func (t *structuredThesis) fieldsOutside() bool {
	return t.Statement == "" || utf8.RuneCountInString(t.Statement) > maxThesisStatementRunes ||
		(t.Horizon != horizonLongTerm && t.Horizon != horizonShortTerm) ||
		itemsOutside(t.Instruments, maxThesisInstruments, maxThesisInstrumentRunes) ||
		itemsOutside(t.Invalidation, maxThesisInvalidation, maxThesisInvalidationRunes) ||
		itemsOutside(t.Sources, maxThesisSources, maxThesisSourceRunes)
}

// itemsOutside reports whether a list has none, more than most, or an item longer
// than runes characters.
func itemsOutside(items []string, most, runes int) bool {
	return len(items) == 0 || len(items) > most || slices.ContainsFunc(items, func(s string) bool { return utf8.RuneCountInString(s) > runes })
}

// hasControlRune reports whether s holds a control character other than a line
// break.
func hasControlRune(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return r != '\n' && unicode.IsControl(r) }) >= 0
}

// thesisSizePatterns are the patterns of what a lease's size looks like in words:
// the word "lease", a share of the portfolio, a sleeve or a budget, a position
// size, and an amount or a percentage that a verb of sizing (allocate, invest,
// put, stake, commit, spend, deploy) puts to work. A price or a figure of the
// company's own, such as "closed at 46.15 USD" or "grew 12%", is no lease size
// and passes. A function, not a variable: the domain holds no package-level
// state.
func thesisSizePatterns() []*regexp.Regexp {
	const unit = `(?:%|pln|zł|zl|eur|usd|gbp|chf|€|\$|£)`
	const verb = `\b(?:allocat\w+|invest(?:ing|s|ed)?|put(?:ting)?|stak(?:e|es|ing)|commit(?:s|ting)?|spend(?:ing)?|deploy(?:s|ing)?|przeznacz\w*|zainwestuj\w*|zaalokuj\w*)`
	return []*regexp.Regexp{
		regexp.MustCompile(`(?i)\blease[sd]?\b|\bdzier[żz]aw\w*`),
		regexp.MustCompile(`(?i)` + verb + `[^.;\n]{0,40}?\d[\d .,]*\s?` + unit),
		regexp.MustCompile(`(?i)` + verb + `[^.;\n]{0,40}?(?:€|\$|£)\s?\d`),
		regexp.MustCompile(`(?i)\d+(?:[.,]\d+)?\s?%\s+(?:of\s+)?(?:the\s+|your\s+|my\s+|our\s+)?(?:portfolio|sleeve|satellite|fief|lease|budget|pool|capital|net worth|portfel\w*)`),
		regexp.MustCompile(`(?i)\bposition size\b|\bwielko[śs][ćc] pozycji\b`),
	}
}

// thesisMentionsSize reports whether s names a lease's size or an amount of
// money, which a thesis never holds (D-229).
func thesisMentionsSize(s string) bool {
	return slices.ContainsFunc(thesisSizePatterns(), func(p *regexp.Regexp) bool { return p.MatchString(s) })
}

// cardsHaystack is the text of the cards the model read, in lower case: their
// titles, texts and the labels, details and figures of their rows, in which an
// instrument of a thesis must be named.
func cardsHaystack(cards []answerCard) string {
	var b strings.Builder
	for i := range cards {
		b.WriteString(strings.ToLower(cards[i].Title + "\n" + cards[i].Text))
		for _, row := range cards[i].Rows {
			b.WriteString("\n" + strings.ToLower(row.Label+"\n"+row.Detail+"\n"+row.Figure))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// thesisCard is a thesis suggestion as the page shows it: its fields in words
// and as the form posts them, and whether the owner may save it as a draft.
// Its field names are those of a saved version, so one fragment draws both.
type thesisCard struct {
	Statement    string
	Horizon      string
	HorizonKey   string
	ShortTerm    bool
	Instruments  []string
	Invalidation []string
	Sources      []string
	// CanSave is true when the page can save it: the core serves draft theses
	// and the person is the owner.
	CanSave bool
}

// cardOfThesis returns the card of a thesis the box checked.
func cardOfThesis(t *structuredThesis) *thesisCard {
	c := &thesisCard{
		Statement: t.Statement, HorizonKey: t.Horizon, Horizon: "Long-term", Instruments: t.Instruments,
		Invalidation: t.Invalidation, Sources: t.Sources,
	}
	if t.Horizon == horizonShortTerm {
		c.Horizon, c.ShortTerm = "Short-term", true
	}
	return c
}

// forRun returns the thesis as a harness reads it; nil for no thesis.
func (c *thesisCard) forRun() *AskThesis {
	if c == nil {
		return nil
	}
	return &AskThesis{Statement: c.Statement, Instruments: c.Instruments, Horizon: c.HorizonKey, Invalidation: c.Invalidation, Sources: c.Sources}
}

// summary is the thesis in one line, as the conversation's history carries it
// for the next question.
func (c *thesisCard) summary() string {
	return fmt.Sprintf("Suggested thesis (%s): %s Instruments: %s. Wrong if: %s Sources: %s",
		c.Horizon, c.Statement, strings.Join(c.Instruments, ", "), strings.Join(c.Invalidation, "; "), strings.Join(c.Sources, "; "))
}

// The words the page says around a thesis suggestion and an opinion.
const (
	// thesisCardNote sits under the suggestion, next to the button that saves it.
	thesisCardNote = "A suggestion from an AI model, not investment advice. It holds no lease size: how much an agent may hold is yours to decide, " +
		"and saving it only keeps the text as a draft on the Theses page."
	// opinionsIneligibleNote is what the box says when the owner turned opinions
	// on and the route is all that keeps them off (D-206, D-225).
	opinionsIneligibleNote = "Opinions are on in Settings, but no model on the Ask box's route is allowed to receive your question and cards (T0), " +
		"a thesis's text (T1) and prices (T2) together, so this answer gives none. Allow one on the Models page."
	// notAdviceNote is the sentence every model answer carries (D-222), and with
	// opinions on it is followed by one that says what an opinion is, so the line
	// of D-222 stays whole.
	notAdviceNote         = "Not investment advice."
	opinionsNotAdviceNote = notAdviceNote + " This is an AI model's opinion from your cards and public prices, and it never acts."
)
