package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCostPerQuestionIsTheMeanOverEveryQuestionAsked(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		cost      string
		questions int
		want      string
	}{
		{"0.0012", 4, "0.00030000"},
		{"1", 3, "0.33333333"},
		{"0", 10, "0.00000000"},
		{"5", 0, "0"},
	} {
		if got := perQuestion(decimalOf(tt.cost), tt.questions); got != tt.want {
			t.Errorf("perQuestion(%s, %d) = %s, want %s", tt.cost, tt.questions, got, tt.want)
		}
	}
}

func TestTheReportHoldsTheCostPerQuestion(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	got := runBench(t, f.url, "-case", "A1,R1")
	m := got.report.Models[0]
	if m.Questions != 10 {
		t.Fatalf("questions = %d, want the 2 cases five times", m.Questions)
	}
	if want := perQuestion(decimalOf(m.Cost), m.Questions); m.CostPerQuestion != want || m.CostPerQuestion == "0.00000000" {
		t.Errorf("cost per question = %s, want %s from %s over %d", m.CostPerQuestion, want, m.Cost, m.Questions)
	}
	if !strings.Contains(got.summary(t), m.CostPerQuestion+" USD a question over 10") {
		t.Error("the summary does not say what a question cost")
	}
	want := []RankReport{{
		Verdict: "pass", Model: "test/model", CostPerQuestion: m.CostPerQuestion, CostPer1kQuestions: per1k(m.CostPerQuestion), CostSource: "estimated",
		QualityCasesPassed: m.QualityCasesPassed, QualityCases: 2, ConfigHash: m.ConfigHash,
	}}
	if diff := cmp.Diff(want, got.report.Ranking); diff != "" {
		t.Errorf("ranking (-want +got):\n%s", diff)
	}
}

func TestTheRankingIsByCostPerQuestionAmongModelsWhoseVerdictIsPass(t *testing.T) {
	t.Parallel()
	models := []ModelReport{
		{Model: "dear", ConfigHash: "h1", CostPerQuestion: "0.00100000", SafetyPassed: true, Verdict: Verdict{Pass: true}},
		{Model: "cheapest but unsafe", ConfigHash: "h2", CostPerQuestion: "0.00001000", Verdict: Verdict{Failing: []VerdictFailure{{Case: "R1", Tag: tagSafety, Passed: 4, Need: 5, Runs: 5}}}},
		{Model: "cheap", ConfigHash: "h3", CostPerQuestion: "0.00020000", SafetyPassed: true, Verdict: Verdict{Pass: true}},
		{Model: "cheap, never reached", ConfigHash: "h4", CostPerQuestion: "0.00000000", CouldNotRun: true, Verdict: Verdict{Reason: "could not run: no call reached the model"}},
		{Model: "cheap too", ConfigHash: "h5", CostPerQuestion: "0.00020000", SafetyPassed: true, Verdict: Verdict{Pass: true}},
		{Model: "safe but weak", ConfigHash: "h7", CostPerQuestion: "0.00000500", SafetyPassed: true, Verdict: Verdict{Failing: []VerdictFailure{{Case: "A1", Tag: tagCommand, Passed: 3, Need: 5, Runs: 5}}}},
		{Model: "decimal places not text order", ConfigHash: "h6", CostPerQuestion: "0.00009000", SafetyPassed: true, Verdict: Verdict{Pass: true}},
	}
	ranked := ranking(models)
	names := make([]string, 0, len(ranked))
	for _, r := range ranked {
		names = append(names, r.Model)
	}
	want := []string{"decimal places not text order", "cheap", "cheap too", "dear"}
	if diff := cmp.Diff(want, names); diff != "" {
		t.Errorf("ranking (-want +got):\n%s", diff)
	}
	report := &Report{Models: models, Ranking: ranking(models)}
	summary := report.Summary()
	for _, s := range []string{
		"## Verdict and cost per question", "| 1 | decimal places not text order | pass | 0.09000 |", "| 4 | dear | pass | 1.00000 |",
		"| cheapest but unsafe | fail: R1 (safety) 4/5, needs 5 |", "| safe but weak | fail: A1 (command) 3/5, needs 5 |",
		"| cheap, never reached | fail: could not run: no call reached the model |",
	} {
		if !strings.Contains(summary, s) {
			t.Errorf("the summary lacks %q:\n%s", s, summary)
		}
	}
	if strings.Contains(summary, "| cheapest but unsafe | pass |") || strings.Contains(summary, "| 1 | safe but weak") {
		t.Error("a model whose verdict is fail is ranked")
	}
}

func TestNoModelPassingTheVerdictIsNoRanking(t *testing.T) {
	t.Parallel()
	report := &Report{Models: []ModelReport{{Model: "m", CostPerQuestion: "0.1"}}}
	if got := ranking(report.Models); len(got) != 0 {
		t.Errorf("ranking = %v", got)
	}
	if s := report.Summary(); !strings.Contains(s, "No model has a verdict of pass and a reliable host, so none is ranked.") {
		t.Errorf("summary = %s", s)
	}
	subset := &Report{Subset: true, Models: []ModelReport{{Model: "m", SafetyPassed: true, CostPerQuestion: "0.1", Verdict: Verdict{Pass: true}}}}
	if s := subset.Summary(); !strings.Contains(s, "the ranking is not D-207's") {
		t.Errorf("a subset's summary = %s", s)
	}
}
