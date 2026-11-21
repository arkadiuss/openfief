package ui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/app/answer"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// errTierDenied is what a harness's route answers a message of a data tier the
// owner did not allow it to receive.
var errTierDenied = errors.New("no model of the route may receive the data tier")

// The Ask box's conversation as a harness sees it: the same messages, tools
// and replies the page's loop exchanges with the core's router.
type (
	// AskMessage is a message of the Ask box's conversation with its model.
	AskMessage = askMessage
	// AskCall is a tool call the model asked for.
	AskCall = askCall
	// AskReply is the model's next message.
	AskReply = askReply
	// AskTool is a tool the model may call: its name, description and the
	// JSON Schema of its arguments.
	AskTool = askTool
	// AskTurn is an earlier turn of a conversation, as the browser sends it
	// with a follow-up (D-224): the question, the answer's text and the compact
	// summaries of its cards. It is untrusted, as the box treats it.
	AskTurn = askTurn
)

// AskModel is the model side of the Ask box's loop for a harness that has no
// core: it answers one call of the loop with at most maxTokens output tokens.
// A call with a schema asks for the structured answer: a JSON text of the JSON
// Schema, strict structured output for a model that has it and JSON mode for one
// that has not, which the loop checks. The page reaches its model through the
// core's router instead (D-203).
type AskModel interface {
	Generate(ctx context.Context, msgs []AskMessage, tools []AskTool, maxTokens int64, schema string) (AskReply, error)
}

// AskerConfig is what an Asker needs.
type AskerConfig struct {
	// Read answers the tools' reads, as the core's read API does for a session.
	Read viewsv1connect.ReadServiceClient
	// Model answers the loop's calls.
	Model AskModel
	// Today returns the date of the question; nil is the UTC date of the clock.
	Today func() kernel.Date
	// Limits are the limits of one question, D-208's defaults for a zero field.
	Limits AskLimits
	// Group and Decimal are the number format of the owner's settings (D-113)
	// that cards write amounts in; an empty Decimal is the plain format.
	Group, Decimal string
	// MarketQuotes says the install has a market-data connector, as the read
	// API's settings do (D-226): without it the market price tool is not
	// offered.
	MarketQuotes bool
	// DeniedTiers are the data tiers the route's models may not receive, as
	// the owner's lists leave them (D-206); none, the harness's default, is a
	// route that allows every tier. A message of a denied tier is refused as the
	// router refuses it, and a tool whose card is of one is not offered.
	// The box's own fixed text carries no tier, as in the core.
	DeniedTiers []string
	// Opinions gives the harness's model the Ask box's opinions (D-225, D-229): it
	// may say whether it would buy, sell or hold and suggest a thesis, as the
	// owner's own install with opinions on and an eligible route does. Off, the
	// harness is the box as a public demo has it, which declines to recommend
	// (R-7).
	Opinions bool
}

// Asker runs the Ask box's model loop for a harness: the loop and the tools
// the page uses (modelAnswerer), over a read service and a model of the
// harness's own. It skips the commands the box answers itself, so every
// question reaches the model.
type Asker struct {
	answerer *modelAnswerer
	format   numberFormat
	quotes   bool
}

// NewAsker returns an Asker over cfg; it fails with ErrMissingDependency
// without a read client or a model.
func NewAsker(cfg *AskerConfig) (*Asker, error) {
	if cfg.Read == nil || cfg.Model == nil {
		return nil, ErrMissingDependency
	}
	s := &server{read: cfg.Read, today: cfg.Today}
	if s.today == nil {
		s.today = utcToday
	}
	commands := commandAnswerer{s: s}
	return &Asker{
		answerer: &modelAnswerer{
			model: askModelFunc{m: cfg.Model, denied: cfg.DeniedTiers, opinionsOn: cfg.Opinions}, tools: answer.Tools(), run: commands.run, today: s.today,
			log: slog.New(slog.DiscardHandler), limits: cfg.Limits.withDefaults(),
		},
		format: separators(cfg.Group, cfg.Decimal), quotes: cfg.MarketQuotes,
	}, nil
}

// Opinions reports whether the harness gives its model opinions.
func (a *Asker) Opinions() bool {
	f, ok := a.answerer.model.(askModelFunc)
	return ok && f.opinionsOn
}

// askModelFunc is askModelPort over an AskModel: a harness's route allows
// every data tier but those it lists as denied, so a model is always available
// and a tool's card may be sent unless its tier is denied.
type askModelFunc struct {
	m      AskModel
	denied []string
	// opinionsOn is whether the harness's model may give opinions.
	opinionsOn bool
}

var _ askModelPort = askModelFunc{}

func (askModelFunc) available(context.Context, session) (bool, error) { return true, nil }

// opinions is the harness's setting: on when the harness gives opinions and its
// route allows T0, T1 and T2, which a harness that denies one of them does not.
func (f askModelFunc) opinions(context.Context, session) (opinionsState, error) {
	switch {
	case !f.opinionsOn:
		return opinionsOff, nil
	case slices.ContainsFunc([]string{tierT0, tierT1, tierT2}, func(t string) bool { return slices.Contains(f.denied, t) }):
		return opinionsIneligible, nil
	}
	return opinionsOn, nil
}

func (f askModelFunc) permits(_ context.Context, _ session, tiers []string) (bool, error) {
	return !slices.ContainsFunc(tiers, func(t string) bool { return slices.Contains(f.denied, t) }), nil
}

// admit lets every question go: a harness has no budget and no limit.
func (askModelFunc) admit(context.Context, session, string) (askAdmission, error) {
	return askAdmission{Kind: admitAllowed}, nil
}

func (f askModelFunc) generate(ctx context.Context, _ session, msgs []askMessage, tools []askTool, maxTokens int64, schema string) (askReply, error) {
	for i := range msgs {
		if slices.Contains(f.denied, msgs[i].Tier) {
			// As the router refuses a request with data no model on the route may receive (D-206).
			return askReply{}, fmt.Errorf("the harness's route: %w", connect.NewError(connect.CodePermissionDenied, errTierDenied))
		}
	}
	reply, err := f.m.Generate(ctx, msgs, tools, maxTokens, schema)
	if err != nil {
		return askReply{}, fmt.Errorf("the harness's model: %w", err)
	}
	return reply, nil
}

// AskRun is what one question did: the answer, the cards it read and every
// call the model made.
type AskRun struct {
	// Answered is true when the model gave a valid structured answer; false
	// when the box gave up (Title says why) and Text is its own explanation.
	Answered bool
	Title    string
	Text     string
	// Kind is the answer's kind, "answer", "clarify" or "refuse"; empty when
	// the box gave up. Candidates are a clarifying answer's choices and
	// Suggestions a refusing answer's commands, as the page shows them.
	Kind        string
	Candidates  []string
	Suggestions []string
	// Opinions is whether the model could give opinions in this question, and
	// Thesis the thesis a "thesis" answer suggested, as the page shows it; nil
	// for any other answer. OpinionsNote is what the page says when opinions
	// were on in Settings and the route kept them off.
	Opinions     bool
	Thesis       *AskThesis
	OpinionsNote string
	// SchemaError says why the model's reply was not a structured answer, when
	// the box gave up for it; Invalid counts the replies that were not one, a
	// retried one included.
	SchemaError string
	Invalid     int
	// Cards are the cards the answer names, the ones the page shows under it,
	// in the order it names them; Read are all the cards the model's tool calls
	// produced, in the order the tools ran.
	Cards []AskRunCard
	Read  []AskRunCard
	// Withheld counts the cards the route's data tiers kept from the model.
	Withheld int
	// Calls are the tool calls the model asked for, in order.
	Calls []AskRunCall
	// Rounds is how many times the model was called.
	Rounds int
}

// AskThesis is a thesis suggestion as the page shows it, with the horizon as
// the contract spells it: "long_term" or "short_term".
type AskThesis struct {
	Statement    string
	Instruments  []string
	Horizon      string
	Invalidation []string
	Sources      []string
}

// AskRunCard is a card the answer shows, as the model read it.
type AskRunCard struct {
	// Tool is the tool that made it.
	Tool  string
	Title string
	Text  string
	Rows  []AskRunRow
}

// AskRunRow is a row of a card.
type AskRunRow struct {
	Label, Detail, Figure string
}

// AskRunCall is a tool call of the model: the tool, its arguments as the
// model wrote them, and how it ended: "ran", "unknown-tool", "over-limit",
// "bad-arguments", "read-failed" or "withheld".
type AskRunCall struct {
	Name, Args, Outcome string
}

// Ask asks question, in a conversation of its own, and returns what it did.
// It fails only for a question that can't be started.
func (a *Asker) Ask(ctx context.Context, question string) (AskRun, error) {
	return a.AskAfter(ctx, question, nil)
}

// AskAfter asks question as a follow-up of the earlier turns, as the page's
// browser would send them (D-224): they are capped and shown to the model as an
// untrusted transcript, the way the page does.
func (a *Asker) AskAfter(ctx context.Context, question string, earlier []AskTurn) (AskRun, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/ask", http.NoBody)
	if err != nil {
		return AskRun{}, fmt.Errorf("starting the question: %w", err)
	}
	r = withHistory(withFormat(r, a.format), earlier)
	sess := session{Token: "ask-bench", Origin: "ask-bench", Settings: &viewsv1.GetSettingsResponse{MarketQuotes: a.quotes}}
	card, conv, err := a.answerer.converse(r, sess, question)
	if err != nil {
		return AskRun{}, fmt.Errorf("asking: %w", err)
	}
	run := AskRun{
		Answered: card.Source == "" && len(card.Help) == 0, Title: card.Title, Text: card.Text, Kind: card.Kind,
		Candidates: card.Candidates, Suggestions: card.Suggestions, Invalid: conv.invalid,
		Withheld: len(conv.withheld), Rounds: conv.rounds, Read: make([]AskRunCard, len(conv.used)),
		Opinions: conv.opinions == opinionsOn, OpinionsNote: card.OpinionsNote,
	}
	run.Thesis = card.Thesis.forRun()
	if !run.Answered && conv.schemaErr != nil {
		run.SchemaError = conv.schemaErr.Error()
	}
	for i := range conv.used {
		run.Read[i] = runCard(&conv.used[i])
	}
	for _, t := range conv.trace {
		run.Calls = append(run.Calls, AskRunCall{Name: t.Name, Args: t.Args, Outcome: t.Outcome})
		if t.Card >= 0 {
			run.Read[t.Card].Tool = t.Name
		}
	}
	for _, i := range conv.cited {
		run.Cards = append(run.Cards, run.Read[i])
	}
	return run, nil
}

// runCard returns card as a harness reads it.
func runCard(card *answerCard) AskRunCard {
	out := AskRunCard{Title: card.Title, Text: card.Text}
	for _, r := range card.Rows {
		out.Rows = append(out.Rows, AskRunRow{Label: r.Label, Detail: r.Detail, Figure: r.Figure})
	}
	return out
}

// AskTools returns the tools the Ask box offers its model, in the order the
// model is told them.
func AskTools() []AskTool { return answer.Tools() }

// AskArgDefaults returns the value each argument of the named tool has when the
// model leaves it out, latest being the latest valued day: what the tool's
// description says in words ("this_year when omitted", "the latest valued day
// when no date is given"). A call that gives an argument its default is the call
// without it. It is nil for a tool with no defaults and for an unknown tool.
func AskArgDefaults(tool string, latest kernel.Date) map[string]any {
	for _, t := range answer.Tools() {
		if t.Name == tool && t.Defaults != nil {
			return t.Defaults(latest)
		}
	}
	return nil
}

// AskArgReplacedBy returns, for each argument of the named tool that has a
// default, the other arguments that replace it: a call that gives one of them
// has no use for the default, since the tool takes the argument given instead
// (a value asked over a period has no date). It is nil for a tool with none
// and for an unknown tool.
func AskArgReplacedBy(tool string) map[string][]string {
	for _, t := range answer.Tools() {
		if t.Name == tool {
			return t.ReplacedBy
		}
	}
	return nil
}

// AskAnswersItself reports whether the Ask box answers question on today itself,
// at once, without asking its model (D-203): the fixed commands, and a question
// the box does not send a model. A harness that grades the model leaves such a
// question out, since in production it never reaches it.
func AskAnswersItself(question string, today kernel.Date) bool {
	return boxAnswersItself(question, today)
}

// AskSystemPrompt returns the system message of a question asked on today, with
// opinions off.
func AskSystemPrompt(today kernel.Date) string { return askSystemPrompt(today) }

// AskSystemPromptFor returns the system message of a question asked on today,
// with opinions on or off (D-225).
func AskSystemPromptFor(today kernel.Date, opinions bool) string {
	return askSystemPromptFor(today, opinions)
}

// AskAnswerSchemaFor returns the JSON Schema the Ask box's final answer follows,
// with opinions on or off.
func AskAnswerSchemaFor(opinions bool) string {
	if opinions {
		return askOpinionsSchemaJSON
	}
	return askAnswerSchemaJSON
}

// AskNote returns the sentence the page says under a model's answer about
// advice, with opinions on or off (D-225): every model answer carries it.
func AskNote(opinions bool) string { return (&modelNote{Opinions: opinions}).Advice() }

// AskOpinionsIneligibleNote is what the box says when the owner turned opinions
// on and the route is all that keeps them off.
func AskOpinionsIneligibleNote() string { return opinionsIneligibleNote }

// AskThesisMentionsSize reports whether text names an amount, a lease size or a
// share of the portfolio, which a thesis suggestion never holds (D-229): the
// rule the box applies to what a model suggests.
func AskThesisMentionsSize(text string) bool { return thesisMentionsSize(text) }

// AskThesisLimits are the most a thesis suggestion may hold: the statement's
// characters, and the items and each item's characters of its instruments,
// invalidation conditions and sources.
type AskThesisLimits struct {
	StatementRunes                  int
	Instruments, InstrumentRunes    int
	Invalidation, InvalidationRunes int
	Sources, SourceRunes            int
}

// AskThesisLimitsOf returns the limits the box holds a thesis suggestion to.
func AskThesisLimitsOf() AskThesisLimits {
	return AskThesisLimits{
		StatementRunes: maxThesisStatementRunes, Instruments: maxThesisInstruments, InstrumentRunes: maxThesisInstrumentRunes,
		Invalidation: maxThesisInvalidation, InvalidationRunes: maxThesisInvalidationRunes, Sources: maxThesisSources, SourceRunes: maxThesisSourceRunes,
	}
}

// AskCheckThesis returns why a thesis suggestion is not one the box shows, or
// nil, given the text of the cards the model read: the same checks the page's
// loop makes (a harness grades by them).
func AskCheckThesis(t *AskThesis, cards []AskRunCard) error {
	st := &structuredThesis{
		Statement: t.Statement, Instruments: t.Instruments, Horizon: t.Horizon, Invalidation: t.Invalidation, Sources: t.Sources,
	}
	st.tidy()
	read := make([]answerCard, len(cards))
	for i := range cards {
		read[i] = answerCard{Title: cards[i].Title, Text: cards[i].Text}
		for _, r := range cards[i].Rows {
			read[i].Rows = append(read[i].Rows, answerRow{Label: r.Label, Detail: r.Detail, Figure: r.Figure})
		}
	}
	return st.check(cardsHaystack(read))
}

// AskPromptCanary returns the token the system prompt tells the model never to
// write: an answer that holds any trace of it has repeated the prompt.
func AskPromptCanary() string { return askPromptCanary }

// AskPromptNonPublic returns the sentences of the system prompt for a question
// asked on today that the box does not say of itself, which an answer must not
// repeat.
func AskPromptNonPublic(today kernel.Date) []string { return askPromptNonPublic(today) }

// AskPromptNonPublicFor is AskPromptNonPublic with opinions on or off.
func AskPromptNonPublicFor(today kernel.Date, opinions bool) []string {
	return askPromptNonPublicFor(today, opinions)
}

// AskCommand is one of the commands the Ask box understands, with an
// example, as it suggests them to a question it can't answer (D-194).
type AskCommand struct {
	Command, Example string
}

// AskCommands returns the commands the Ask box understands.
func AskCommands() []AskCommand {
	known := answer.KnownCommands()
	out := make([]AskCommand, len(known))
	for i, k := range known {
		out[i] = AskCommand(k)
	}
	return out
}
