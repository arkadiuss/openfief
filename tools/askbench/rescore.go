package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/ui"
)

// Re-scoring (D-217, D-219, D-220): the same answers graded by the scoring
// rules of this build, with no call to a model. What a grade reads of a run is
// the answer, the calls the model made, the cards it read, the box's own title
// when it gave up, and each call's tools offered, output tokens and output
// limit; the rest, the portfolio's names and numbers and the case's expected
// calls, is the bench's own and is rebuilt from the seed and cases.yaml.
//
// report.json of schema 1 does not hold the cards' text (its `cards` names only
// the tools that made them) nor the output limit each call carried
// (`maxTokens`). From schema 2 on the bench stores both. For a report of schema
// 1 the cards are rebuilt, not invented: the run's stored calls are replayed
// through the box's own loop over the bench's seed, which makes the same cards
// (the cards are a function of the call and the seed, and the model's words do
// not enter them), and the rebuild is trusted only when the replay ends every
// call as the run did and the cards the answer named can be told apart. A run
// for which it cannot is kept as graded, and counted. The output limit cannot
// be rebuilt, so the Limits check of a schema-1 run is carried over as it was
// graded. Both are counted in the report's `rescore` section.

// maxRescoreNotes is how many of the reasons a run was kept as graded a report
// lists.
const maxRescoreNotes = 10

// RescoreReport says how a report was re-graded.
type RescoreReport struct {
	// From is the scorer version of the report that was re-graded, SourceSchema
	// its report.json schema.
	From         int `json:"from"`
	SourceSchema int `json:"sourceSchema"`
	// Runs is how many runs were graded again and Changed how many of them
	// passed or failed differently now.
	Runs    int `json:"runs"`
	Changed int `json:"changed"`
	// RebuiltCards is how many runs had their cards rebuilt by replaying their
	// calls (a report of schema 1 does not hold them), CarriedChecks how many
	// check results were carried over as they were graded (the Limits check of
	// such a run, which needs the output limit each call carried), and Kept how
	// many runs were kept as graded because their cards could not be rebuilt,
	// with the reasons in Notes.
	RebuiltCards  int      `json:"rebuiltCards"`
	CarriedChecks int      `json:"carriedChecks"`
	Kept          int      `json:"kept"`
	Notes         []string `json:"notes,omitempty"`
	// DerivedFaults is how many failed calls of the report had no fault stored
	// (a report before scorer version 3) and had it derived from their status and
	// label; Undecided how many of those are a 4xx that only the request's hash,
	// which that report did not store, could decide: they count in no host's
	// reliability and void nothing. The reliability of every model is counted
	// again per question from the stored call failures.
	DerivedFaults int `json:"derivedFaults,omitempty"`
	Undecided     int `json:"undecided,omitempty"`
}

// LoadReport reads a bench's report.json.
func LoadReport(path string) (*Report, error) {
	raw, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("%w: reading %s: %w", ErrRun, path, err)
	}
	var r Report
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("%w: decoding %s: %w", ErrRun, path, err)
	}
	return &r, nil
}

// Rescore grades the runs of the report again by this build's scoring rules,
// in place, and rebuilds everything that follows from the grades: each case's
// counts, each model's totals and verdict, and the ranking. It calls no model.
// The streaks and the demo route's section come from the history and are not
// kept: they are not what the new grades say.
func Rescore(ctx context.Context, rep *Report) error {
	b, err := NewBench(&Options{})
	if err != nil {
		return err
	}
	info := &RescoreReport{From: scorerOf(rep), SourceSchema: rep.Schema}
	for mi := range rep.Models {
		if err := b.rescoreModel(ctx, rep, &rep.Models[mi], info); err != nil {
			return err
		}
	}
	rep.Ranking = ranking(rep.Models)
	rep.Route = nil
	rep.Schema, rep.ScorerVersion, rep.Rescore = reportSchema, ScorerVersion, info
	return nil
}

// rescoreModel grades a model's runs again and totals them.
func (b *Bench) rescoreModel(ctx context.Context, rep *Report, m *ModelReport, info *RescoreReport) error {
	limits, err := limitsOf(&m.Config.Limits)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrRun, m.Model, err)
	}
	m.SafetyPassed, m.SafetyCases, m.SafetyCasesPassed, m.QualityCases, m.QualityCasesPassed, m.MatcherCases = true, 0, 0, 0, 0, 0
	m.Streak = nil
	for ci := range m.Cases {
		cr := &m.Cases[ci]
		i := slices.IndexFunc(b.cases, func(c Case) bool { return c.ID == cr.ID })
		if i < 0 {
			return fmt.Errorf("%w: %s: the report has case %s, which the bench no longer has", ErrRun, m.Model, cr.ID)
		}
		for ri := range cr.Detail {
			cr.Detail[ri].deriveFaults(info)
			if err := b.rescoreRun(ctx, &runKey{c: &b.cases[i], limits: limits, schema: rep.Schema}, &cr.Detail[ri], info); err != nil {
				return fmt.Errorf("%w: %s %s run %d: %w", ErrRun, m.Model, cr.ID, cr.Detail[ri].Run, err)
			}
		}
		fresh := caseReport(&b.cases[i], cr.Detail, rep.RunsPerCase)
		m.Cases[ci] = fresh
		m.addTotals(&fresh)
	}
	m.Verdict = evalVerdict(m.Cases, rep.RunsPerCase, m.CouldNotRun)
	m.recountReliability()
	return nil
}

// recountReliability counts the model's void runs and its reliability again from
// its runs' stored call failures: per question (D-207), whatever the report
// counted before.
func (m *ModelReport) recountReliability() {
	m.VoidRuns, m.VoidCredentials, m.Reliability = 0, 0, Reliability{}
	for ci := range m.Cases {
		m.VoidRuns += m.Cases[ci].Void
		m.VoidCredentials += m.Cases[ci].VoidCredentials
		m.Reliability.add(m.Cases[ci].Detail)
	}
	m.Reliability.finish()
}

// deriveFaults gives the failed calls of a stored run the fault a report before
// scorer version 3 did not store, from what it did store: a refused key and no
// credits void the run, a 404 is the host's, no answer, a timeout, a 5xx, a 429
// and a bad or degenerate reply are the host's as they were counted, an invalid
// request and a local error are the bench's. Another 4xx cannot be decided, for
// that needs the hash of the request, and is counted and left out. Then it
// settles the run again.
func (run *RunReport) deriveFaults(info *RescoreReport) {
	for i := range run.CallErrors {
		ce := &run.CallErrors[i]
		if ce.Fault != "" {
			continue
		}
		derived := legacyFault(ce)
		if derived.fault == "" {
			info.Undecided++
			continue
		}
		ce.Fault, ce.Rule = derived.fault, derived.rule
		info.DerivedFaults++
		if derived.fault == faultHost && (ce.Status == http.StatusNotFound || ce.Label == labelBadReply || ce.Label == labelDegenerate || ce.Label == labelFailed) {
			run.TransportFailures++ // a 404 and a bad, degenerate or bare failed reply were no failure of the host before
		}
	}
	run.settle()
}

// derivedFault is the fault and the rule given to a failed call that had none.
type derivedFault struct{ fault, rule string }

// legacyFault is the fault and the rule of a failed call that has none stored,
// empty for a call that cannot be decided from its status, label and class.
func legacyFault(ce *CallErrorReport) derivedFault {
	switch {
	case ce.Status == http.StatusUnauthorized || ce.Status == http.StatusPaymentRequired:
		return derivedFault{faultCredentials, ruleCredentials}
	case ce.Status == http.StatusNotFound:
		return derivedFault{faultHost, ruleNotFound}
	case ce.Status == http.StatusTooManyRequests || ce.Status >= http.StatusInternalServerError:
		return derivedFault{faultHost, ""}
	case ce.Status != 0:
		return derivedFault{}
	}
	switch ce.Label {
	case labelInvalid, labelLocal:
		return derivedFault{faultBench, ""}
	}
	return derivedFault{faultHost, ""}
}

// runKey is what a run's re-grading needs besides the run: its case, its
// model's limits and the schema of the report it is in.
type runKey struct {
	c      *Case
	limits ui.AskLimits
	schema int
}

// limitsOf reads the limits a model's run was asked under, from its report.
func limitsOf(l *LimitsReport) (ui.AskLimits, error) {
	deadline, err := time.ParseDuration(l.Deadline)
	if err != nil {
		return ui.AskLimits{}, fmt.Errorf("the report's deadline %q: %w", l.Deadline, err)
	}
	return ui.AskLimits{ToolRounds: l.ToolRounds, ToolCalls: l.ToolCalls, OutputTokens: l.OutputTokens, Deadline: deadline}, nil
}

// wasGraded reports whether a stored run was graded: the matcher's, one never
// asked and one a failed call ended were not.
func wasGraded(run *RunReport) bool {
	return !run.Matcher && !run.NotRun && !run.BudgetSpent && !run.Errored && run.Error == ""
}

// rescoreRun grades one run again; a run that was not graded stays as it is.
func (b *Bench) rescoreRun(ctx context.Context, k *runKey, run *RunReport, info *RescoreReport) error {
	if !wasGraded(run) {
		return nil
	}
	d := b.runDataOf(k, run)
	cards := b.cardsOf(ctx, k, run)
	if cards.why != "" {
		info.keep(k.c.ID, run.Run, cards.why)
		return nil
	}
	d.Run.Cards, d.Run.Read = cards.cards, cards.read
	results, err := regrade(ctx, k, run, d, info)
	if err != nil {
		return err
	}
	was := run.Passed
	run.Checks, run.Calls, run.Cards, run.Safety = nil, nil, nil, nil
	run.fill(d, results)
	info.Runs++
	if cards.rebuilt {
		info.RebuiltCards++
	}
	if was != run.Passed {
		info.Changed++
	}
	return nil
}

// regrade grades a stored run by the current rules. A report of schema 1 did not
// store the output limit each call carried, so its Limits check is carried over
// as it was graded.
func regrade(ctx context.Context, k *runKey, run *RunReport, d *RunData, info *RescoreReport) ([]result, error) {
	results, err := grade(ctx, d)
	if err != nil || k.schema >= storedCardsSchema {
		return results, err
	}
	return carryChecks(results, run.Checks, []string{"Limits"}, info)
}

// keep counts a run kept as graded and notes why, up to a few notes.
func (info *RescoreReport) keep(id string, run int, why string) {
	info.Kept++
	if len(info.Notes) < maxRescoreNotes {
		info.Notes = append(info.Notes, fmt.Sprintf("%s run %d: %s", id, run, why))
	}
}

// cardsOf returns the run's cards: those the report holds (schema 2), or, for a
// report of schema 1, the ones its calls rebuild.
func (b *Bench) cardsOf(ctx context.Context, k *runKey, run *RunReport) replayed {
	if k.schema >= storedCardsSchema {
		return replayed{cards: askRunCards(run.CardData), read: askRunCards(run.ReadData)}
	}
	got := b.replayCards(ctx, k, run)
	got.rebuilt = got.why == ""
	return got
}

// runDataOf returns what the scorers read of a stored run but its cards: its
// answer, calls and rounds, with the bench's own seed-derived data.
func (b *Bench) runDataOf(k *runKey, run *RunReport) *RunData {
	d := &RunData{
		Case: k.c, Limits: k.limits, Duration: time.Duration(run.LatencyMs) * time.Millisecond,
		NonPublic: b.prompt, Canary: ui.AskPromptCanary(), Vocabulary: b.vocab, RealNames: b.names, Today: b.seed.Today,
	}
	d.Run = ui.AskRun{
		Answered: run.Answered, Text: run.Answer, Title: run.Title, Kind: run.Kind, Candidates: run.Candidates, Suggestions: run.Suggestions,
		SchemaError: run.SchemaError, Invalid: run.Invalid,
	}
	if !run.Answered && run.Title == "" {
		// Schema 1 kept the box's title and text as one: "title: text".
		d.Run.Title, d.Run.Text, _ = strings.Cut(run.Answer, ": ")
	}
	for _, c := range run.Calls {
		d.Run.Calls = append(d.Run.Calls, ui.AskRunCall{Name: c.Tool, Args: c.Args, Outcome: c.Outcome})
	}
	for i := range run.Rounds {
		r := &run.Rounds[i]
		d.Calls = append(d.Calls, ModelCall{ToolsOffered: r.ToolsOffered, MaxTokens: r.MaxTokens, OutputTokens: r.CompletionTokens, Round: *r})
	}
	return d
}

// carryChecks puts the stored result of each named check in place of its new
// one, which cannot be graded from the report.
func carryChecks(results []result, stored []CheckReport, names []string, info *RescoreReport) ([]result, error) {
	for i := range results {
		if !slices.Contains(names, results[i].Name) {
			continue
		}
		j := slices.IndexFunc(stored, func(c CheckReport) bool { return c.Name == results[i].Name })
		if j < 0 {
			return nil, fmt.Errorf("%w: the check %s cannot be graded from the report and was not stored", ErrRun, results[i].Name)
		}
		results[i] = result{Name: stored[j].Name, Pass: stored[j].Passed, Reason: stored[j].Reason, Safety: stored[j].Safety}
		info.CarriedChecks++
	}
	return results, nil
}

// replayModel is a model that asks for the stored calls of a run in one round
// and then answers naming the cards of the calls that ran: it reproduces the
// run's tool outcomes and cards, not its words.
type replayModel struct {
	calls []ui.AskRunCall
}

var _ ui.AskModel = replayModel{}

// Generate asks for the stored calls, then answers.
func (m replayModel) Generate(_ context.Context, msgs []ui.AskMessage, _ []ui.AskTool, _ int64, _ string) (ui.AskReply, error) {
	if len(m.calls) == 0 || msgs[len(msgs)-1].Role == chatTool {
		var ids []string
		for i, c := range m.calls {
			if c.Outcome == "ran" {
				ids = append(ids, fmt.Sprintf(`"c%d"`, i+1))
			}
		}
		answer := `{"kind":"answer","text":"replay","cards":[` + strings.Join(ids, ",") + `],"suggestions":[],"candidates":[]}`
		return ui.AskReply{Content: answer, Model: "replay"}, nil
	}
	reply := ui.AskReply{Model: "replay"}
	for i, c := range m.calls {
		reply.Calls = append(reply.Calls, ui.AskCall{ID: fmt.Sprintf("c%d", i+1), Name: c.Name, Args: c.Args})
	}
	return reply, nil
}

// replayed are a run's cards: the ones its answer named and all the ones its
// calls read, whether they were rebuilt by a replay, and, when they could not be
// had, why.
type replayed struct {
	cards, read []ui.AskRunCard
	rebuilt     bool
	why         string
}

// replayCards rebuilds the cards of a run of a report of schema 1 by replaying
// its calls over the bench's seed. A rebuild is trusted only when every call
// ends as the run's did and the named cards can be told apart; otherwise why
// says why not.
func (b *Bench) replayCards(ctx context.Context, k *runKey, run *RunReport) replayed {
	stored := make([]ui.AskRunCall, len(run.Calls))
	for i, c := range run.Calls {
		stored[i] = ui.AskRunCall{Name: c.Tool, Args: c.Args, Outcome: c.Outcome}
	}
	style := k.c.numberFormat()
	asker, err := ui.NewAsker(&ui.AskerConfig{
		Read: newReadClient(b.seed, b.pf), Model: replayModel{calls: stored}, Today: func() kernel.Date { return b.seed.Today },
		Limits: k.limits, Group: style.Group, Decimal: style.Decimal,
	})
	if err != nil {
		return replayed{why: "replay: " + err.Error()}
	}
	out, err := asker.AskAfter(ctx, k.c.Question, k.c.earlier())
	if err != nil {
		return replayed{why: "replay: " + err.Error()}
	}
	if !slices.EqualFunc(out.Calls, stored, func(a, b ui.AskRunCall) bool { return a.Name == b.Name && a.Outcome == b.Outcome }) {
		return replayed{why: "the replayed calls did not end as the run's did"}
	}
	cards, ok := namedCards(out.Read, run.Cards)
	if !ok {
		return replayed{why: "the cards the answer named cannot be told from the cards read"}
	}
	return replayed{cards: cards, read: out.Read}
}

// namedCards returns the cards of read that an answer naming the tools names,
// in the order it names them, when that is certain: every card of a tool the
// answer names is named, so which cards they are does not depend on the model's
// words. Cards of the same tool are taken in the order they were read.
func namedCards(read []ui.AskRunCard, names []string) ([]ui.AskRunCard, bool) {
	queue := map[string][]ui.AskRunCard{}
	for i := range read {
		queue[read[i].Tool] = append(queue[read[i].Tool], read[i])
	}
	asked := map[string]int{}
	for _, n := range names {
		asked[n]++
	}
	for tool, n := range asked {
		if len(queue[tool]) != n {
			return nil, false
		}
	}
	var out []ui.AskRunCard
	for _, n := range names {
		out = append(out, queue[n][0])
		queue[n] = queue[n][1:]
	}
	return out, true
}
