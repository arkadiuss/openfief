package ui

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

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
}

// Asker runs the Ask box's model loop for a harness: the loop and the tools
// the page uses (modelAnswerer), over a read service and a model of the
// harness's own. It skips the commands the box answers itself, so every
// question reaches the model.
type Asker struct {
	answerer *modelAnswerer
	format   numberFormat
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
			model: askModelFunc{m: cfg.Model}, tools: askTools(), run: commands.run, today: s.today,
			log: slog.New(slog.DiscardHandler), limits: cfg.Limits.withDefaults(),
		},
		format: separators(cfg.Group, cfg.Decimal),
	}, nil
}

// askModelFunc is askModelPort over an AskModel: a harness's route allows
// every data tier, so a model is always available and every tool's card may
// be sent.
type askModelFunc struct {
	m AskModel
}

var _ askModelPort = askModelFunc{}

func (askModelFunc) available(context.Context, session) (bool, error) { return true, nil }

func (askModelFunc) permits(context.Context, session, []string) (bool, error) { return true, nil }

// admit lets every question go: a harness has no budget and no limit.
func (askModelFunc) admit(context.Context, session, string) (askAdmission, error) {
	return askAdmission{Kind: admitAllowed}, nil
}

func (f askModelFunc) generate(ctx context.Context, _ session, msgs []askMessage, tools []askTool, maxTokens int64, answer bool) (askReply, error) {
	schema := ""
	if answer {
		schema = askAnswerSchemaJSON
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
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/ask", http.NoBody)
	if err != nil {
		return AskRun{}, fmt.Errorf("starting the question: %w", err)
	}
	r = withFormat(r, a.format)
	card, conv, err := a.answerer.converse(r, session{Token: "ask-bench", Origin: "ask-bench"}, question)
	if err != nil {
		return AskRun{}, fmt.Errorf("asking: %w", err)
	}
	run := AskRun{
		Answered: card.Source == "" && len(card.Help) == 0, Title: card.Title, Text: card.Text, Kind: card.Kind,
		Candidates: card.Candidates, Suggestions: card.Suggestions, Invalid: conv.invalid,
		Withheld: len(conv.withheld), Rounds: conv.rounds, Read: make([]AskRunCard, len(conv.used)),
	}
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
func AskTools() []AskTool { return askTools() }

// AskArgDefaults returns the value each argument of the named tool has when the
// model leaves it out, latest being the latest valued day: what the tool's
// description says in words ("this_year when omitted", "the latest valued day
// when no date is given"). A call that gives an argument its default is the call
// without it. It is nil for a tool with no defaults and for an unknown tool.
func AskArgDefaults(tool string, latest kernel.Date) map[string]any {
	for _, t := range askTools() {
		if t.Name == tool && t.Defaults != nil {
			return t.Defaults(latest)
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

// AskSystemPrompt returns the system message of a question asked on today.
func AskSystemPrompt(today kernel.Date) string { return askSystemPrompt(today) }

// AskPromptCanary returns the token the system prompt tells the model never to
// write: an answer that holds any trace of it has repeated the prompt.
func AskPromptCanary() string { return askPromptCanary }

// AskPromptNonPublic returns the sentences of the system prompt for a question
// asked on today that the box does not say of itself, which an answer must not
// repeat.
func AskPromptNonPublic(today kernel.Date) []string { return askPromptNonPublic(today) }

// AskCommand is one of the commands the Ask box understands, with an
// example, as it suggests them to a question it can't answer (D-194).
type AskCommand struct {
	Command, Example string
}

// AskCommands returns the commands the Ask box understands.
func AskCommands() []AskCommand {
	known := knownCommands()
	out := make([]AskCommand, len(known))
	for i, k := range known {
		out[i] = AskCommand(k)
	}
	return out
}
