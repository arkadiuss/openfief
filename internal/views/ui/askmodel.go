package ui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Fixed limits of the Ask box's model calls: each call to the core and each
// tool under a timeout, and the length of what is sent and shown.
const (
	// askCallTimeout bounds one call to the core's router, which may try
	// several models.
	askCallTimeout = 45 * time.Second
	// askToolTimeout bounds one tool, which reads the core a few times.
	askToolTimeout = 20 * time.Second
	// maxModelQuestionRunes caps a question sent to the model.
	maxModelQuestionRunes = 300
	// maxModelAnswerRunes caps the text of a model's answer shown.
	maxModelAnswerRunes = 4000
)

// D-208's limits of one question.
const (
	defaultToolRounds   = 5
	defaultToolCalls    = 10
	defaultOutputTokens = 800
	defaultDeadline     = 2 * time.Minute
)

// AskLimits are the limits of one question the Ask box sends a model (D-203,
// D-208): a bounded tool loop, a token budget and a deadline. A zero field
// takes its default, DefaultAskLimits'.
type AskLimits struct {
	// ToolRounds is how many rounds may call tools. One more round follows
	// with no tools offered, so the model has to answer with what it read.
	ToolRounds int
	// ToolCalls is how many tool calls one question runs in all; a call
	// beyond it gets a refusal and runs nothing.
	ToolCalls int
	// OutputTokens is the most tokens the model may write for one question,
	// all its calls together: each call is limited to what is left, and no
	// call is made once it is spent.
	OutputTokens int64
	// Deadline bounds the whole question.
	Deadline time.Duration
}

// DefaultAskLimits returns the limits D-208 sets: five rounds that may call
// tools then a final one without, ten tool calls, 800 output tokens and two
// minutes.
func DefaultAskLimits() AskLimits {
	return AskLimits{
		ToolRounds: defaultToolRounds, ToolCalls: defaultToolCalls, OutputTokens: defaultOutputTokens, Deadline: defaultDeadline,
	}
}

// withDefaults returns l with each zero field replaced by its default.
func (l AskLimits) withDefaults() AskLimits {
	d := DefaultAskLimits()
	if l.ToolRounds <= 0 {
		l.ToolRounds = d.ToolRounds
	}
	if l.ToolCalls <= 0 {
		l.ToolCalls = d.ToolCalls
	}
	if l.OutputTokens <= 0 {
		l.OutputTokens = d.OutputTokens
	}
	if l.Deadline <= 0 {
		l.Deadline = d.Deadline
	}
	return l
}

// askSystemPrompt is the system message of every question. It holds no
// portfolio data, so it is public (T2): the owner's data comes in only as
// the result of a tool the model calls.
const askSystemPrompt = "You answer questions about the owner's investment portfolio in OpenFief, a portfolio tracker. " +
	"You can only read: the tools are the whole of what you can do, and nothing you say or ask for changes anything. " +
	"Get every figure, date and name from the tools and never guess or compute one the tools don't give. " +
	"If a tool result says it was withheld or failed, say what you could not read and answer only from the rest. " +
	"Answer briefly in plain text, with no markdown, in the language of the question. " +
	"Amounts are in the owner's reporting currency unless a result says otherwise. Today is "

// askMessage is a message of the conversation with the model, with the data
// tier of what it holds (D-206).
type askMessage struct {
	// Role is "system", "user", "assistant" or "tool".
	Role    string
	Content string
	// Calls are the tools an assistant message asked for.
	Calls []askCall
	// CallID is the call a tool message answers.
	CallID string
	// Tier is "T0", "TF", "T1" or "T2".
	Tier string
}

// askCall is a call of a tool the model asked for.
type askCall struct {
	ID, Name string
	// Args is the arguments, a JSON object text.
	Args string
}

// askReply is the model's next message.
type askReply struct {
	Content string
	Calls   []askCall
	// Model is the model that answered, as "provider/model".
	Model string
	// OutputTokens is what the answer used.
	OutputTokens int64
}

// askModelPort is the core's router, as the Ask box uses it (D-203): every
// call goes through the switch, the route, the data tiers and the audit log
// there. It is declared here, by its consumer; askModelClient implements it
// over the AskModelService.
type askModelPort interface {
	// available reports whether the Ask box has a model for sess.
	available(ctx context.Context, sess session) (bool, error)
	// permits reports whether a model of the route may receive data of every
	// tier together, with tools.
	permits(ctx context.Context, sess session, tiers []string) (bool, error)
	// generate asks the route's models for the next message, of at most
	// maxTokens output tokens.
	generate(ctx context.Context, sess session, msgs []askMessage, tools []askTool, maxTokens int64) (askReply, error)
}

// toolRunner runs a command as a tool: the commands' own answerer, reading
// with the asking person's session (never the plugin's own identity), so
// their grants limit what the model sees.
type toolRunner func(r *http.Request, sess session, q *askQuery) (answerCard, error)

// modelAnswerer answers the questions beyond the Ask box's commands with a
// model (D-203). The model may only call the commands' read-only answerers
// as tools, and the page shows the cards its calls produced next to its
// text. It holds no question or answer beyond one request, logs neither
// (GR-12) and changes nothing.
type modelAnswerer struct {
	model askModelPort
	tools []askTool
	run   toolRunner
	today func() kernel.Date
	log   *slog.Logger
	// limits are the limits of one question, with no zero field.
	limits AskLimits
}

// ask answers question with the model: the tool loop of at most
// limits.ToolRounds rounds that may call tools and a last one that may not,
// each tool result checked against the data tiers the route's models may
// receive before it joins the conversation. A card the route's models may not
// receive is withheld, and the model is told so.
//
// The question is the owner's own words, which may carry amounts or holdings,
// so it is T0 like a tool's result (D-206): a route whose models may not
// receive T0 is refused by the router before anything is sent.
func (a *modelAnswerer) ask(r *http.Request, sess session, question string) (answerCard, error) {
	ctx, cancel := context.WithTimeout(r.Context(), a.limits.Deadline)
	defer cancel()
	r = r.WithContext(ctx)
	conv := &askConversation{
		a: a, r: r, sess: sess, tiers: map[string]bool{tierT0: true},
		msgs: []askMessage{
			{Role: "system", Content: askSystemPrompt + a.today().String() + ".", Tier: tierT2},
			{Role: "user", Content: question, Tier: tierT0},
		},
	}
	for round := 1; round <= a.limits.ToolRounds+1; round++ {
		tools := a.tools
		if round > a.limits.ToolRounds {
			tools = nil
		}
		if conv.tokens >= a.limits.OutputTokens {
			return overBudgetAnswer(question), nil
		}
		reply, err := conv.generate(ctx, tools)
		if err != nil {
			return a.failure(r, question, err)
		}
		conv.tokens += reply.OutputTokens
		if len(reply.Calls) == 0 {
			return conv.answer(question, &reply), nil
		}
		conv.msgs = append(conv.msgs, askMessage{Role: "assistant", Content: reply.Content, Calls: reply.Calls, Tier: mostSensitive(conv.tiers)})
		for _, call := range reply.Calls {
			if err := conv.callTool(ctx, call); err != nil {
				return a.failure(r, question, err)
			}
		}
	}
	return unfinishedAnswer(question), nil
}

// askConversation is one question's conversation with the model: its
// messages, the data tiers they hold, the cards its tool calls produced and
// those withheld.
type askConversation struct {
	a    *modelAnswerer
	r    *http.Request
	sess session
	msgs []askMessage
	// tiers are the data tiers the messages hold so far.
	tiers map[string]bool
	calls int
	// tokens are the output tokens the model has used so far.
	tokens   int64
	used     []answerCard
	withheld []withheldCard
}

// generate asks the router for the next message, offering tools and no more
// output tokens than the question has left.
func (c *askConversation) generate(ctx context.Context, tools []askTool) (askReply, error) {
	ctx, cancel := context.WithTimeout(ctx, askCallTimeout)
	defer cancel()
	return c.a.model.generate(ctx, c.sess, c.msgs, tools, c.a.limits.OutputTokens-c.tokens)
}

// callTool runs one tool call of the model and adds its result to the
// conversation: the card's data when the route's models may receive it, a
// refusal when not, a failure when the tool or its arguments are wrong. It
// returns only an error that ends the question, such as a session that is no
// longer valid.
func (c *askConversation) callTool(ctx context.Context, call askCall) error {
	c.calls++
	tool, ok := c.a.tool(call.Name)
	if !ok {
		c.addResult(&call, toolFailure("there is no tool called %s", call.Name), tierT2)
		return nil
	}
	if c.calls > c.a.limits.ToolCalls {
		c.addResult(&call, toolFailure("too many tool calls for one question; answer with what you have"), tierT2)
		return nil
	}
	q, err := tool.query(call.Args, c.a.today())
	if err != nil {
		c.addResult(&call, toolFailure("%s", err.Error()), tierT2)
		return nil
	}
	toolCtx, cancel := context.WithTimeout(ctx, askToolTimeout)
	defer cancel()
	card, err := c.a.run(c.r.WithContext(toolCtx), c.sess, &q)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeUnauthenticated || ctx.Err() != nil {
			return err
		}
		// What failed is not the model's to see: its result says only that
		// the read did not work.
		c.addResult(&call, toolFailure("the data could not be read"), tierT2)
		return nil
	}
	tiers := c.tiers
	next := map[string]bool{tool.Tier: true}
	for t := range tiers {
		next[t] = true
	}
	allowed, err := c.a.model.permits(ctx, c.sess, tierList(next))
	if err != nil {
		return err
	}
	if !allowed {
		c.withheld = append(c.withheld, withheldCard{Label: tool.Label, Tier: tool.Tier, Data: tierWords(tool.Tier)})
		c.addResult(&call, toolFailure("withheld: this model may not receive %s", tierWords(tool.Tier)), tierT2)
		return nil
	}
	c.tiers = next
	c.used = append(c.used, card)
	c.addResult(&call, cardResult(&card), tool.Tier)
	return nil
}

// addResult appends the result of call to the conversation.
func (c *askConversation) addResult(call *askCall, content, tier string) {
	c.msgs = append(c.msgs, askMessage{Role: "tool", Content: content, CallID: call.ID, Tier: tier})
}

// answer is the model's final text with the cards it read and those it was
// refused.
func (c *askConversation) answer(question string, reply *askReply) answerCard {
	text := strings.TrimSpace(reply.Content)
	if text == "" {
		text = "The model returned no text."
	}
	if utf8.RuneCountInString(text) > maxModelAnswerRunes {
		text = string([]rune(text)[:maxModelAnswerRunes]) + "…"
	}
	return answerCard{Question: question, Text: text, Model: reply.Model, Used: c.used, Withheld: c.withheld}
}

// tool returns the tool called name.
func (a *modelAnswerer) tool(name string) (askTool, bool) {
	i := slices.IndexFunc(a.tools, func(t askTool) bool { return t.Name == name })
	if i < 0 {
		return askTool{}, false
	}
	return a.tools[i], true
}

// tierList returns the tiers of set in tierOrder.
func tierList(set map[string]bool) []string {
	var out []string
	for _, t := range tierOrder() {
		if set[t] {
			out = append(out, t)
		}
	}
	return out
}

// failure turns an error that ended a question into its answer: an error of
// the person's own session goes up to the page, which sends them to log in;
// every other one is a card that says what to do, with the commands the box
// still understands. It logs the failing call's Connect code only, never the
// question or any answer (GR-12).
func (a *modelAnswerer) failure(r *http.Request, question string, err error) (answerCard, error) {
	code := connect.CodeOf(err)
	switch {
	case code == connect.CodeUnauthenticated:
		return answerCard{}, err
	case r.Context().Err() != nil && !errors.Is(r.Context().Err(), context.DeadlineExceeded):
		return answerCard{}, err // the person left
	case code == connect.CodeFailedPrecondition:
		// The switch or the toggle went off, or the route was cleared, since
		// the box last looked: it behaves as it does with no model.
		return unknownAnswer(question), nil
	case code == connect.CodePermissionDenied:
		return modelCard(question, "No model may see this question",
			"A question may carry amounts or holdings, so it is T0 data, and the owner has not allowed any model on the Ask box's route to receive T0. Nothing was sent. The box's commands still answer from your own record."), nil
	case code == connect.CodeUnavailable, code == connect.CodeDeadlineExceeded, errors.Is(err, context.DeadlineExceeded):
		return modelCard(question, "The model didn't answer",
			"Every model on the route was down or too slow. Try again in a moment; the commands answer from your own record."), nil
	}
	a.log.WarnContext(r.Context(), "ask model call failed", "code", code.String())
	return modelCard(question, "The model couldn't answer",
		"Something went wrong asking it. The commands answer from your own record."), nil
}

// overBudgetAnswer is the answer of a model that spent the question's output
// tokens on tool calls.
func overBudgetAnswer(question string) answerCard {
	return modelCard(question, "The model didn't finish",
		"It used up the words one question may take before it answered. Ask something narrower, or use a command.")
}

// unfinishedAnswer is the answer of a model that kept asking for tools
// through the last round, which offers none.
func unfinishedAnswer(question string) answerCard {
	return modelCard(question, "The model didn't finish",
		"It kept asking for more data than one question may read. Ask something narrower, or use a command.")
}

// modelCard is a card about the model that lists the commands the box
// understands.
func modelCard(question, title, text string) answerCard {
	return answerCard{Question: question, Title: title, Text: text, Help: knownCommands(), Source: "The model gave no answer"}
}

// routedAnswerer is the Ask box's answerer (D-203): the fixed commands
// answer themselves, at once; with a model configured anything else goes to
// the model; without one it gets the commands the box understands, as it
// did before there was a model.
type routedAnswerer struct {
	commands commandAnswerer
	model    *modelAnswerer
	today    func() kernel.Date
	log      *slog.Logger
}

// answer answers question with its command, or with the model.
func (a routedAnswerer) answer(r *http.Request, sess session, question string) (answerCard, error) {
	if a.model == nil {
		return a.commands.answer(r, sess, question)
	}
	if _, isCommand := parseAsk(question, a.today()); isCommand {
		return a.commands.answer(r, sess, question)
	}
	if q := normalizeQuestion(question); q == "" || utf8.RuneCountInString(q) > maxModelQuestionRunes {
		return a.commands.answer(r, sess, question)
	}
	ok, err := a.model.model.available(r.Context(), sess)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeUnauthenticated {
			return answerCard{}, fmt.Errorf("asking whether the Ask box has a model: %w", err)
		}
		// Without an answer the box behaves as it does with no model.
		a.log.WarnContext(r.Context(), "ask model check failed", "code", connect.CodeOf(err).String())
		return a.commands.answer(r, sess, question)
	}
	if !ok {
		return a.commands.answer(r, sess, question)
	}
	return a.model.ask(r, sess, question)
}

// askHasModel reports whether the Ask box has a model for sess, so the box
// can say where a question beyond its commands goes. It is false without a
// model client and when the core can't say: the box then claims nothing it
// can't keep.
func (s *server) askHasModel(r *http.Request, sess session) bool {
	if s.askModel == nil {
		return false
	}
	ok, err := s.askModel.available(r.Context(), sess)
	if err != nil {
		s.log.WarnContext(r.Context(), "ask model check failed", "code", connect.CodeOf(err).String())
		return false
	}
	return ok
}
