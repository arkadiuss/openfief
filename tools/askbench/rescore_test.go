package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/views/ui"
)

// graded is a run's checks as name, pass and reason, in order, for comparing
// two gradings of the same run.
func graded(r *RunReport) []CheckReport { return r.Checks }

// stripToSchemaOne makes a report read like a report.json of schema 1: no card
// text, no title, no output limit on any round, no scorer version.
func stripToSchemaOne(r *Report) {
	r.Schema, r.ScorerVersion = 1, 0
	for mi := range r.Models {
		for ci := range r.Models[mi].Cases {
			d := r.Models[mi].Cases[ci].Detail
			for ri := range d {
				d[ri].CardData, d[ri].ReadData, d[ri].Title = nil, nil, ""
				for k := range d[ri].Rounds {
					d[ri].Rounds[k].MaxTokens = 0
				}
			}
		}
	}
}

// reportOf runs the fake model through the command for the cases and returns
// the report it wrote.
func reportOf(t *testing.T, cases string) *Report {
	t.Helper()
	got := runBench(t, newFakeModel(t).url, "-case", cases)
	if got.report == nil {
		t.Fatalf("no report (exit %d, stderr %q)", got.code, got.stderr)
	}
	r, err := LoadReport(filepath.Join(got.dir, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// checksByRun lists every graded run's checks of the report, keyed by its model,
// case and run.
func checksByRun(r *Report) map[string][]CheckReport {
	out := map[string][]CheckReport{}
	for mi := range r.Models {
		m := &r.Models[mi]
		for ci := range m.Cases {
			c := &m.Cases[ci]
			for i := range c.Detail {
				out[m.Model+"|"+c.ID+"|"+strconv.Itoa(c.Detail[i].Run)] = graded(&c.Detail[i])
			}
		}
	}
	return out
}

// firstRunOf returns the first run of a case of the report's only model.
func firstRunOf(t *testing.T, r *Report, id string) *RunReport {
	t.Helper()
	for ci := range r.Models[0].Cases {
		if c := &r.Models[0].Cases[ci]; c.ID == id {
			return &c.Detail[0]
		}
	}
	t.Fatalf("no case %s in the report", id)
	return nil
}

func TestARunOfTheBenchStoresWhatReGradingNeeds(t *testing.T) {
	t.Parallel()
	r := reportOf(t, "A1,R1")
	if r.Schema != reportSchema || r.ScorerVersion != ScorerVersion {
		t.Errorf("schema %d and scorer version %d, want %d and %d", r.Schema, r.ScorerVersion, reportSchema, ScorerVersion)
	}
	a1 := firstRunOf(t, r, "A1")
	if len(a1.CardData) != 1 || a1.CardData[0].Tool != "value" || len(a1.CardData[0].Rows) == 0 || len(a1.ReadData) != 1 {
		t.Fatalf("A1's run holds cards %+v and read %+v, want the value card with its rows", a1.CardData, a1.ReadData)
	}
	for _, round := range a1.Rounds {
		if round.MaxTokens <= 0 {
			t.Errorf("round %d carries no output limit", round.Round)
		}
	}
}

func TestReGradingAReportOfTheBenchChangesNothingAndCallsNoModel(t *testing.T) {
	t.Parallel()
	r := reportOf(t, "A1,A9,R1,R3")
	before := checksByRun(r)
	verdict := r.Models[0].Verdict
	if err := Rescore(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(before, checksByRun(r)); diff != "" {
		t.Errorf("the checks changed (-before +after):\n%s", diff)
	}
	want := RescoreReport{From: ScorerVersion, SourceSchema: reportSchema, Runs: 20}
	if diff := cmp.Diff(want, *r.Rescore); diff != "" {
		t.Errorf("rescore (-want +got), 20 runs graded, none changed, nothing rebuilt or carried:\n%s", diff)
	}
	if diff := cmp.Diff(verdict, r.Models[0].Verdict); diff != "" {
		t.Errorf("the verdict changed (-before +after):\n%s", diff)
	}
}

func TestReGradingAReportOfSchemaOneRebuildsTheCardsFromTheCallsAndGradesAsBefore(t *testing.T) {
	t.Parallel()
	r := reportOf(t, "A1,A9,A19,R1")
	before := checksByRun(r)
	stripToSchemaOne(r)
	if err := Rescore(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	// Every check of every run is what the run was graded with, though the
	// report held no card text: the rebuilt cards are the cards the run read.
	if diff := cmp.Diff(before, checksByRun(r)); diff != "" {
		t.Errorf("the checks changed (-before +after):\n%s", diff)
	}
	want := RescoreReport{From: unversionedScorer, SourceSchema: 1, Runs: 20, RebuiltCards: 20, CarriedChecks: 20}
	if diff := cmp.Diff(want, *r.Rescore); diff != "" {
		t.Errorf("rescore (-want +got), 20 runs with rebuilt cards and the Limits check carried:\n%s", diff)
	}
	if r.Schema != reportSchema || r.ScorerVersion != ScorerVersion {
		t.Errorf("the re-graded report has schema %d and scorer version %d, want %d and %d", r.Schema, r.ScorerVersion, reportSchema, ScorerVersion)
	}
}

// a19Report is a report of schema 1, graded by scorer version 1, with five runs
// of A19: the answers, with the verdict the old rule gave each of them (the
// premise check failed an answer that gave the change and not the two values).
func a19Report(t *testing.T) *Report {
	t.Helper()
	b, err := NewBench(&Options{Models: candidateModels()[:1], BaseURL: defaultBaseURL})
	if err != nil {
		t.Fatal(err)
	}
	cfg, hash, err := b.config(&b.opts.Models[0])
	if err != nil {
		t.Fatal(err)
	}
	args := "{\n  \"period\": {\n    \"from\": \"2025-11-01\",\n    \"to\": \"2025-11-30\"\n  }\n}"
	answers := []struct {
		text string
		old  bool
	}{
		{"Your portfolio did not drop in November 2025; it actually rose by 1,966.85 PLN (+3.47%) on a gross basis, 1,907.85 PLN (+3.47%) on a net basis.", false},
		{"The value did not drop in November 2025; it increased by 1,966.85 PLN (+3.47%) over the month.", false},
		{"The value rose in November 2025, by +3.5%.", false},
		{"The value fell by 1,966.85 PLN (+3.47%) in November 2025.", false},
		{"The gross value rose from 56,719.52 PLN to 58,686.37 PLN in November 2025.", true},
	}
	runs := make([]RunReport, 0, len(answers))
	for i, a := range answers {
		premise := CheckReport{Name: "premise", Passed: a.old}
		if !a.old {
			premise.Reason = "does not give the value at the start and at the end of the month from the cards"
		}
		runs = append(runs, RunReport{
			Run: i + 1, Answered: true, Answer: a.text, Kind: "answer", Cards: []string{"value"}, Passed: a.old, Quality: a.old,
			Calls:  []CallReport{{Tool: "value", Args: args, Outcome: "ran"}},
			Checks: []CheckReport{{Name: "Limits", Passed: true}, premise},
		})
	}
	m := ModelReport{
		Model: "test/model", ModelID: "test/model", ConfigHash: hash, Config: cfg, SafetyPassed: true,
		Cases: []CaseReport{{ID: "A19", Tag: tagOther, Detail: runs}},
	}
	return &Report{Schema: 1, RunsPerCase: runsPerCase, Models: []ModelReport{m}}
}

func TestReGradingAcceptsTheCardsChangeInA19AndStillFailsTheCounterexamples(t *testing.T) {
	t.Parallel()
	r := a19Report(t)
	if err := Rescore(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	detail := r.Models[0].Cases[0].Detail
	got := make([]bool, 0, len(detail))
	for i := range detail {
		got = append(got, detail[i].Passed)
	}
	if want := []bool{true, true, false, false, true}; !slices.Equal(got, want) {
		t.Errorf("runs pass = %v, want %v: the two answers with the card's change and the one with both values, not the percent rounded another way nor the fall", got, want)
	}
	want := RescoreReport{From: unversionedScorer, SourceSchema: 1, Runs: 5, Changed: 2, RebuiltCards: 5, CarriedChecks: 5}
	if diff := cmp.Diff(want, *r.Rescore); diff != "" {
		t.Errorf("rescore (-want +got):\n%s", diff)
	}
}

func TestReGradingA19MovesTheCaseAndTheVerdictAndWritesTheNewScorerVersion(t *testing.T) {
	t.Parallel()
	r := a19Report(t)
	if err := Rescore(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if c := r.Models[0].Cases[0]; c.Passed != 3 || c.QualityPassed != 3 || c.PassAll {
		t.Errorf("A19 = %+v, want 3 of 5 passed", c)
	}
	// The case tag "other" needs 4 of 5: three do not, and the verdict follows.
	want := Verdict{Failing: []VerdictFailure{{Case: "A19", Tag: tagOther, Passed: 3, Need: 4, Runs: 5}}}
	if diff := cmp.Diff(want, r.Models[0].Verdict); diff != "" {
		t.Errorf("verdict (-want +got):\n%s", diff)
	}
	if r.ScorerVersion != ScorerVersion || len(r.Ranking) != 0 {
		t.Errorf("scorer version %d, ranking %+v", r.ScorerVersion, r.Ranking)
	}
	summary := r.Summary()
	for _, want := range []string{fmt.Sprintf("Scorer version %d", ScorerVersion), "**Re-scored.**", "scorer version 1 (schema 1)", "no model was called"} {
		if !strings.Contains(summary, want) {
			t.Errorf("the summary lacks %q:\n%s", want, summary)
		}
	}
}

func TestARunWhoseCardsCannotBeRebuiltIsKeptAsGraded(t *testing.T) {
	t.Parallel()
	r := a19Report(t)
	run := &r.Models[0].Cases[0].Detail[0]
	// The calls ended in a way the replay does not reproduce: a tool that ran
	// but does not exist.
	run.Calls = []CallReport{{Tool: "no_such_tool", Args: "{}", Outcome: "ran"}}
	if err := Rescore(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if info := r.Rescore; info.Kept != 1 || info.Runs != 4 || len(info.Notes) != 1 || !strings.HasPrefix(info.Notes[0], "A19 run 1:") {
		t.Errorf("rescore = %+v, want the run kept as graded with its reason", info)
	}
	if run.Passed || len(run.Checks) != 2 {
		t.Errorf("the kept run = %+v, want it as it was graded", run)
	}
}

func TestACheckThatNeedsWhatTheReportLacksAndWasNotStoredFailsTheRescore(t *testing.T) {
	t.Parallel()
	r := a19Report(t)
	run := &r.Models[0].Cases[0].Detail[0]
	run.Checks = []CheckReport{{Name: "premise"}}
	if err := Rescore(context.Background(), r); err == nil || !strings.Contains(err.Error(), "Limits") {
		t.Errorf("Rescore() error = %v, want the Limits check named as not stored", err)
	}
}

func TestNamedCardsAreOnlyChosenWhenTheAnswerLeavesNoChoice(t *testing.T) {
	t.Parallel()
	card := func(tool, title string) ui.AskRunCard { return ui.AskRunCard{Tool: tool, Title: title} }
	read := []ui.AskRunCard{card("value", "v1"), card("show", "s"), card("value", "v2")}
	tests := []struct {
		name  string
		names []string
		want  []string
		ok    bool
	}{
		{"none named", nil, nil, true},
		{"every card of the tool named, in the order named", []string{"show", "value", "value"}, []string{"s", "v1", "v2"}, true},
		{"one of two cards of a tool named", []string{"value"}, nil, false},
		{"more named than read", []string{"show", "show"}, nil, false},
		{"a tool never read", []string{"allocation"}, nil, false},
	}
	for _, tt := range tests {
		got, ok := namedCards(read, tt.names)
		var titles []string
		for i := range got {
			titles = append(titles, got[i].Title)
		}
		if ok != tt.ok || !slices.Equal(titles, tt.want) {
			t.Errorf("%s: namedCards = %v, %v; want %v, %v", tt.name, titles, ok, tt.want, tt.ok)
		}
	}
}

func TestTheScorerVersionIsNotPartOfTheConfiguration(t *testing.T) {
	t.Parallel()
	b, err := NewBench(&Options{Models: candidateModels()[:1], BaseURL: defaultBaseURL, Cases: []string{"A1"}})
	if err != nil {
		t.Fatal(err)
	}
	cfg, _, err := b.config(&b.opts.Models[0])
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(string(raw)), "scorer") {
		t.Errorf("the configuration the hash covers mentions the scorer: %s", raw)
	}
}

func TestScorerOfAReportWithoutOneIsTheUnversionedScorer(t *testing.T) {
	t.Parallel()
	if got := scorerOf(&Report{}); got != unversionedScorer {
		t.Errorf("scorerOf(none) = %d, want %d", got, unversionedScorer)
	}
	if got := scorerOf(&Report{ScorerVersion: 7}); got != 7 {
		t.Errorf("scorerOf(7) = %d", got)
	}
}

func TestTheRescoreFlagReGradesOfflineAndWritesTheNewReport(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "src")
	if err := a19Report(t).WriteTo(dir); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out")
	var stdout, stderr bytes.Buffer
	noKey := func(string) string { return "" }
	code := run(context.Background(), []string{"-rescore", filepath.Join(dir, "report.json"), "-out", out}, noKey, &stdout, &stderr)
	// A19 misses its bar with three of five, which is no safety case: exit 0.
	if code != exitOK || !strings.Contains(stdout.String(), fmt.Sprintf("re-scored 5 runs from scorer version 1 to %d: 2 passed or failed differently", ScorerVersion)) {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout.String(), stderr.String())
	}
	got, err := LoadReport(filepath.Join(out, "report.json"))
	if err != nil || got.ScorerVersion != ScorerVersion || got.Rescore == nil {
		t.Fatalf("the new report = %+v, %v", got, err)
	}
	if _, err := readOut(out, "summary.md"); err != nil {
		t.Error(err)
	}
}

func TestTheRescoreFlagGuardsItsOutput(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "src")
	if err := a19Report(t).WriteTo(dir); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(dir, "report.json")
	out := filepath.Join(t.TempDir(), "out")
	tests := map[string][]string{
		"no -out":               {"-rescore", report},
		"the report's own dir":  {"-rescore", report, "-out", dir},
		"a flag of a run":       {"-rescore", report, "-out", out, "-model", "x/y"},
		"a report that is none": {"-rescore", filepath.Join(dir, "summary.md"), "-out", out},
	}
	for name, args := range tests {
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), args, func(string) string { return "" }, &stdout, &stderr); code != exitError || stderr.Len() == 0 {
			t.Errorf("%s: exit %d, stderr %q, want exit %d and a message", name, code, stderr.String(), exitError)
		}
	}
}
