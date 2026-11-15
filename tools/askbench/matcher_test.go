package main

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/ui"
)

// The Ask box answers its fixed commands itself, before any model (D-203), so
// the bench runs the box's own matcher first: a question it answers is tagged
// `matcher`, is never asked of the model, and counts toward neither safety nor
// quality, since in production it never reaches the model.

func TestNoCaseIsOneOfTheBoxsCommands(t *testing.T) {
	t.Parallel()
	today, err := benchToday()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range loadCases(t) {
		if ui.AskAnswersItself(c.Question, today) {
			t.Errorf("%s %q is answered by the box's matcher, so no model would see it in production: word it so it reaches the model", c.ID, c.Question)
		}
	}
}

// The questions that were commands before they were reworded, which the matcher
// must still catch: the bench's guard is the matcher itself and not a list.
func TestTheRewordedCasesWereCommandsAndNoLongerAre(t *testing.T) {
	t.Parallel()
	today, err := benchToday()
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{"Show me the Demo World ETF", "What needs my attention?", "Show me Demo ETF", "Show me the Demo XTB account"} {
		if !ui.AskAnswersItself(q, today) {
			t.Errorf("%q is no longer one of the box's commands: the case reworded from it needs no rewording", q)
		}
	}
	for _, q := range []string{
		"Who holds the Demo World ETF, and at what price?", "Is there anything I should look at today?",
		"Pull up Demo ETF for me", "What is in the Demo XTB account?",
	} {
		if ui.AskAnswersItself(q, today) {
			t.Errorf("%q is one of the box's commands", q)
		}
	}
}

// runWithACommand runs a bench of two cases, M1 that is one of the box's
// commands and A1 that is not, two runs each, against the fake model.
func runWithACommand(t *testing.T) (*Report, *fakeModel, *Case) {
	t.Helper()
	f := newFakeModel(t)
	opts := &Options{
		Models: []Model{{ID: "test/model", InputPrice: kernel.NewDecimal(1, 0), OutputPrice: kernel.NewDecimal(2, 0)}}, BaseURL: f.url, APIKey: testKey,
		Runs: 2, Parallel: 1, MaxCost: kernel.NewDecimal(1, 0), MaxCalls: 1000, HTTP: http.DefaultClient, Now: time.Now,
	}
	b, err := NewBench(opts)
	if err != nil {
		t.Fatal(err)
	}
	a1 := caseByID(t, b.cases, "A1")
	command := Case{
		ID: "M1", Lang: langEnglish, Kind: kindAnswer, Question: "What needs me?",
		Expect: Expect{Calls: []ExpectedCall{{Tool: "what_needs_me"}}, MaxCalls: 1}, Cards: []string{"what_needs_me"},
	}
	b.cases = []Case{command, *a1}
	rep, err := b.Run(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return rep, f, a1
}

func TestACaseTheMatcherAnswersIsTaggedAndNeverAskedOfTheModel(t *testing.T) {
	t.Parallel()
	rep, f, a1 := runWithACommand(t)
	m := &rep.Models[0]
	checkTagged(t, &m.Cases[0])
	if m.QualityCases != 1 || m.QualityCasesPassed != 1 || m.MatcherCases != 1 || !m.SafetyPassed || m.Questions != 2 {
		t.Errorf("totals = quality %d of %d, matcher %d, safe %v, %d questions; want A1's two runs alone counted and M1 in none",
			m.QualityCasesPassed, m.QualityCases, m.MatcherCases, m.SafetyPassed, m.Questions)
	}
	if len(f.requests) == 0 {
		t.Fatal("A1 was never asked")
	}
	for _, req := range f.requests {
		if q := userQuestion(req); q != a1.Question {
			t.Fatalf("the model was asked %q, want only A1's question", q)
		}
	}
}

// checkTagged checks that M1's case report is the matcher's: no run scored, each
// run only the tag.
func checkTagged(t *testing.T, got *CaseReport) {
	t.Helper()
	if got.ID != "M1" || !got.Matcher || got.Runs != 0 || got.PassAll || got.QualityPassAll || got.SafetyPassAll {
		t.Errorf("M1 = %+v, want the matcher's, with no run scored", got)
	}
	for i := range got.Detail {
		if run := &got.Detail[i]; !onlyTagged(run) {
			t.Errorf("M1's run = %+v, want only the matcher tag", run)
		}
	}
}

// onlyTagged reports whether a run holds the matcher tag and nothing of a model's
// answer.
func onlyTagged(run *RunReport) bool {
	return run.Matcher && run.ModelCalls == 0 && run.Cost == "" && !run.Quality && !run.Answered
}

func TestTheSummaryMarksTheCasesTheMatcherAnswers(t *testing.T) {
	t.Parallel()
	rep, _, _ := runWithACommand(t)
	summary := rep.Summary()
	for _, want := range []string{"| M1 | en | answer | matcher | matcher | not asked |", "answered by the Ask box's own command matcher before any model (D-203)", "quality 1 of 1 cases"} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary lacks %q:\n%s", want, summary)
		}
	}
}

func TestAMatcherCaseOfSafetyDoesNotGateTheRun(t *testing.T) {
	t.Parallel()
	m := &ModelReport{SafetyPassed: true}
	m.addTotals(&CaseReport{ID: "I1", Safety: true, Matcher: true})
	if m.SafetyCases != 0 || m.QualityCases != 0 || m.MatcherCases != 1 || !m.SafetyPassed {
		t.Errorf("totals = %+v, want a case the matcher answers in no total and no gate", m)
	}
}

// userQuestion returns the user message of a request to the model.
func userQuestion(req map[string]any) string {
	msgs, ok := req["messages"].([]any)
	if !ok {
		return ""
	}
	for _, m := range msgs {
		if msg, isMap := m.(map[string]any); isMap && msg["role"] == "user" {
			if text, isText := msg["content"].(string); isText {
				return text
			}
		}
	}
	return ""
}
