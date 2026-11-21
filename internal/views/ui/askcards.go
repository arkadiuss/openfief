package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arkadiuss/openfief/internal/views/app/answer"
)

// The Ask box's commands, their questions and cards are the answer package's:
// the page's loop, the model's tools and the MCP server's tools all run the
// same answerers. The page adds what only it draws, below.
type (
	// askQuery is a question the Ask box understood: its command and what the
	// command names.
	askQuery = answer.Query
	// askTool is one tool the Ask box's model may call (D-203).
	askTool = answer.Tool
	// answerRow is one row of an answer's table.
	answerRow = answer.Row
	// answerLink is an answer's link to the page it summarizes.
	answerLink = answer.Link
	// askHelp is one command the Ask box understands, with an example.
	askHelp = answer.Help
)

// answerCard is the Ask box's answer to one question (D-194), as the page
// shows it: the commands' card (answer.Card), a title, a sentence, a small
// table and one link to the place it summarizes, or, for a question the box
// doesn't understand, the commands it does; and, when a model answered, what
// the page says about the model. It is the seam an agent fills in M2 (P-57):
// whatever answers a question, the page shows its card, and nothing in the
// card changes anything.
type answerCard struct {
	Question string
	Title    string
	Text     string
	Rows     []answerRow
	// Link is the one "Open …" link to the place the answer summarizes;
	// nil when there is none.
	Link *answerLink
	// Help are the commands the box understands, for a question it
	// doesn't; empty otherwise.
	Help []askHelp
	// Model is the model that wrote Text, as "provider/model", when a model
	// answered (D-203); empty for the commands' own answers.
	Model string
	// Note is what a model's answer says under it: its name, that it can be
	// wrong and the link to the eval (D-222); nil for the commands' own
	// answers and for a card about a model that did not answer.
	Note *modelNote
	// Source, when set, is what the card's header says in place of "Answered
	// from your data": a card about a model that did not answer.
	Source string
	// Kind is the kind of a model's structured answer (answerKindAnswer,
	// answerKindClarify or answerKindRefuse); empty for any other card.
	Kind string
	// Candidates are the names a clarifying answer asks the person to choose
	// from, each asked as a show command; Suggestions are the commands a
	// refusing answer offers instead, each a chip: the candidates are names a
	// card the model read holds, the suggestions the box's own commands.
	Candidates  []string
	Suggestions []string
	// Used are the cards the model read and named in its answer, shown under
	// its text: each one its tool calls produced and it was sent.
	Used []answerCard
	// CheckNeeded is set on the card of a model question that came with no
	// Turnstile token: the response then tells the page to run the check and
	// ask again (checkNeededTrigger).
	CheckNeeded bool
	// pass is the Turnstile pass cookie the response sets, issued or renewed for
	// a model question that went to the model (D-208); the page never sees it.
	pass *http.Cookie
	// Turn is what the answer adds to the conversation, as the JSON the page
	// keeps with it and sends back with the next question (D-224); empty for
	// a card that adds nothing.
	Turn string
	// Withheld are the cards the model asked for that the router would not
	// send it, because no model of the route may receive their data tier
	// (D-206).
	Withheld []withheldCard
}

// withheldCard is a card the model asked for and did not get: the tool that
// makes it, as the person knows it, and the data tier that kept it back. It
// holds none of the card's data.
type withheldCard struct {
	Label string
	Tier  string
	// Data is what the tier is, in words.
	Data string
}

// cardOf returns the page's card for a card of the commands.
func cardOf(c *answer.Card) answerCard {
	return answerCard{Title: c.Title, Text: c.Text, Rows: c.Rows, Link: c.Link, Help: c.Help}
}

// answerer answers a question asked in the Ask box for a session. M0's is
// commandAnswerer, which knows a fixed set of commands and reads the read
// API alone; an M2 agent can answer through the same cards (P-57). An
// answerer never changes anything, and never logs or sends the question
// anywhere (GR-12).
type answerer interface {
	answer(r *http.Request, sess session, question string) (answerCard, error)
}

// commandAnswerer answers the Ask box's commands (answer.Parse) from the read
// API, with the session's token, as every page reads it, and links the cards
// to the pages.
type commandAnswerer struct {
	s *server
}

// commands returns the answerers over the server's read API and clock.
func (a commandAnswerer) commands() *answer.Commands {
	return answer.NewCommands(a.s.read, a.s.today, pageLinks{})
}

// callerOf returns who asks: the session's token and the number format its
// request carries.
func callerOf(r *http.Request, sess session) answer.Caller {
	return answer.Caller{Token: sess.Token, Format: answer.NumberFormat(requestFormat(r))}
}

// answer parses question and answers its command; a question none of the
// commands matches gets the commands the box understands.
func (a commandAnswerer) answer(r *http.Request, sess session, question string) (answerCard, error) {
	card, err := a.commands().Answer(r.Context(), callerOf(r, sess), question)
	if err != nil {
		return answerCard{}, fmt.Errorf("answering a question: %w", err)
	}
	out := cardOf(&card)
	out.Question = question
	return out, nil
}

// run answers the command q understood, reading with sess. It is what the
// model's tools call too (D-203): the same read-only answerers, with the
// asking person's own session. A query that is no command gets the commands
// the box understands.
func (a commandAnswerer) run(r *http.Request, sess session, q *askQuery) (answerCard, error) {
	card, err := a.commands().Run(r.Context(), callerOf(r, sess), q)
	if err != nil {
		return answerCard{}, fmt.Errorf("running a command: %w", err)
	}
	return cardOf(&card), nil
}

// pageLinks is where the cards' links go: the UI's pages.
type pageLinks struct{}

var _ answer.Links = pageLinks{}

// Inbox is the Inbox.
func (pageLinks) Inbox() string { return inboxPath }

// InboxItem is one item of the Inbox.
func (pageLinks) InboxItem(id int64) string { return inboxPath + "#item-" + strconv.FormatInt(id, 10) }

// Plugins is the installed plugins' page.
func (pageLinks) Plugins() string { return pluginsPath }

// Account is an account's page.
func (pageLinks) Account(id string) string { return "/accounts/" + url.PathEscape(id) }

// Sleeve is a sleeve on the overview.
func (pageLinks) Sleeve(id string) string { return "/#sleeve-" + url.PathEscape(id) }

// Instrument is an instrument's page.
func (pageLinks) Instrument(isin string) string { return instrumentPath(isin) }

// Allocation is the overview's allocation by the dimension.
func (pageLinks) Allocation(dimension string) string {
	return "/?" + url.Values{"by": {dimension}, overParam: {overAllocation}}.Encode() + "#over-time"
}

// Overview is the overview on the day.
func (pageLinks) Overview(day string) string {
	return "/?" + url.Values{dateParam: {day}}.Encode()
}

// PerformanceChart is the performance chart from the day.
func (pageLinks) PerformanceChart(from string) string {
	return "/?" + url.Values{overParam: {overPerformance}, startParam: {from}}.Encode() + "#over-time"
}

// unknownAnswer answers a question the box doesn't understand with the
// commands it does.
func unknownAnswer(question string) answerCard {
	c := answer.UnknownAnswer()
	out := cardOf(&c)
	out.Question = question
	return out
}

// cardResult returns card as the JSON text of a tool's result, with the id the
// model names it by: the card's data and nothing the page needs to draw it.
func cardResult(card *answerCard, id string) string {
	return answer.CardResult(&answer.Card{Title: card.Title, Text: card.Text, Rows: card.Rows}, id)
}
