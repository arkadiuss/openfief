package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAProviderCostIsReadExactly(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text string
		want string
		ok   bool
	}{
		{"0.0000123", "0.0000123", true},
		{"1.5e-05", "0.000015", true},
		{"0", "0", true},
		{"0.1234567891", "0.1234567891", true},
		{"2", "2", true},
		{"", "", false},
		{"free", "", false},
		{"-0.01", "", false},
	} {
		got, ok := measuredCost(tt.text)
		if ok != tt.ok || (ok && got.String() != tt.want) {
			t.Errorf("measuredCost(%q) = %s, %v; want %s, %v", tt.text, got, ok, tt.want, tt.ok)
		}
	}
}

func TestTheMeasuredCostIsTheCostOfTheRunAndOfTheBudget(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	f.cost, f.provider = "0.001", "Nextbit"
	got := runBench(t, f.url, "-case", "A1")
	m := got.report.Models[0]
	// Five runs of two calls, each 0.001 USD as the reply says, whatever the
	// model entry's prices (1 and 2 USD per million tokens) would make of the tokens.
	if m.Cost != "0.01" || got.report.Budget.Cost != "0.01" {
		t.Errorf("cost %s and budget %s, want the measured 0.01 in both", m.Cost, got.report.Budget.Cost)
	}
	if m.CostSource != "measured" || m.MeasuredCalls != 10 || m.EstimatedCalls != 0 {
		t.Errorf("source %q, %d measured, %d estimated; want measured on all 10", m.CostSource, m.MeasuredCalls, m.EstimatedCalls)
	}
	if m.CostPerQuestion != "0.00200000" || m.CostPer1kQuestions != "2.00000" {
		t.Errorf("per question %s, per 1k %s; want 0.002 and 2", m.CostPerQuestion, m.CostPer1kQuestions)
	}
	if run := got.caseOf(t, "A1").Detail[0]; run.Cost != "0.002" || run.EstimatedCalls != 0 {
		t.Errorf("run = cost %s, estimated %d", run.Cost, run.EstimatedCalls)
	}
}

func TestTheHostsThatServedTheCallsAreListedWithTheirCounts(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	f.cost, f.provider = "0.001", "Nextbit"
	got := runBench(t, f.url, "-case", "A1")
	if diff := cmp.Diff([]ProviderCount{{Provider: "Nextbit", Calls: 10}}, got.report.Models[0].Providers); diff != "" {
		t.Errorf("providers (-want +got):\n%s", diff)
	}
	if run := got.caseOf(t, "A1").Detail[0]; run.Providers["Nextbit"] != 2 {
		t.Errorf("the run's providers = %v, want Nextbit twice", run.Providers)
	}
	summary := got.summary(t)
	for _, want := range []string{"measured cost", "Served by: Nextbit ×10.", "| 1 | test/model | pass | 2.00000 | measured |"} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary.md lacks %q", want)
		}
	}
}

func TestACostFromTheTokenPricesIsSaidSoWhenTheReplyHasNone(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	got := runBench(t, f.url, "-case", "A1")
	m := got.report.Models[0]
	if m.CostSource != "estimated" || m.MeasuredCalls != 0 || m.EstimatedCalls != 10 || m.Cost != "0.0117" || len(m.Providers) != 0 {
		t.Errorf("model = source %q, %d measured, %d estimated, cost %s, providers %v; want 10 calls priced by the entry's tokens",
			m.CostSource, m.MeasuredCalls, m.EstimatedCalls, m.Cost, m.Providers)
	}
	summary := got.summary(t)
	for _, want := range []string{"estimated from the token prices: the replies held no cost", "| 1 | test/model | pass | 2.34000 | estimated |"} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary.md lacks %q", want)
		}
	}
	if strings.Contains(summary, "Served by") {
		t.Error("the summary names hosts that no reply named")
	}
}

func TestACostMeasuredOnSomeCallsIsMixedAndEachCallIsCountedByItsSource(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	f.cost, f.costCalls = "0.001", 4
	got := runBench(t, f.url, "-case", "A1")
	m := got.report.Models[0]
	if m.CostSource != "mixed" || m.MeasuredCalls != 4 || m.EstimatedCalls != 6 {
		t.Fatalf("source %q, %d measured, %d estimated; want mixed, 4 and 6", m.CostSource, m.MeasuredCalls, m.EstimatedCalls)
	}
	// Four calls carry the measured 0.001 and six are priced by their tokens;
	// the runs' costs add up to the model's.
	total := decimalOf("0")
	for _, r := range got.caseOf(t, "A1").Detail {
		total = total.Add(decimalOf(r.Cost))
	}
	if total.String() != m.Cost {
		t.Errorf("the runs' costs add to %s, the model's cost is %s", total, m.Cost)
	}
	if !strings.Contains(got.summary(t), "measured on 4 calls and estimated from the token prices on 6") {
		t.Error("the summary does not say the cost is mixed")
	}
}

func TestTheBudgetCountsTheMeasuredCost(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	f.cost = "0.01" // the token prices alone would make a call cost about 0.001
	got := runBench(t, f.url, "-case", "A1", "-budget", "0.03")
	if got.code != exitBudget || !got.report.Budget.Aborted {
		t.Errorf("exit %d, aborted %v; want the budget, spent at 0.01 a call, to end the run", got.code, got.report.Budget.Aborted)
	}
}

func TestTheRankingIsByMeasuredCostPer1kQuestionsAmongThosePassingWithQualityBesideIt(t *testing.T) {
	t.Parallel()
	models := []ModelReport{
		{Model: "dear and good", ConfigHash: "h1", CostPerQuestion: "0.00100000", CostSource: "measured", SafetyPassed: true, Verdict: Verdict{Pass: true}, QualityCases: 28, QualityCasesPassed: 27},
		{Model: "cheapest, unsafe", ConfigHash: "h2", CostPerQuestion: "0.00001000", CostSource: "measured", SafetyPassed: false, Verdict: Verdict{Failing: []VerdictFailure{{Case: "R1", Tag: tagSafety, Passed: 0, Need: 5, Runs: 5}}}, QualityCases: 28, QualityCasesPassed: 28},
		{Model: "cheap and weak", ConfigHash: "h3", CostPerQuestion: "0.00020000", CostSource: "mixed", SafetyPassed: true, Verdict: Verdict{Pass: true}, QualityCases: 28, QualityCasesPassed: 9},
	}
	report := &Report{Models: models, Ranking: ranking(models)}
	want := []RankReport{
		{Verdict: "pass", Model: "cheap and weak", CostPerQuestion: "0.00020000", CostPer1kQuestions: "0.20000", CostSource: "mixed", QualityCasesPassed: 9, QualityCases: 28, ConfigHash: "h3"},
		{Verdict: "pass", Model: "dear and good", CostPerQuestion: "0.00100000", CostPer1kQuestions: "1.00000", CostSource: "measured", QualityCasesPassed: 27, QualityCases: 28, ConfigHash: "h1"},
	}
	if diff := cmp.Diff(want, report.Ranking); diff != "" {
		t.Errorf("ranking (-want +got):\n%s", diff)
	}
	summary := report.Summary()
	for _, s := range []string{
		"| Rank | Model | Verdict | USD per 1k questions | Cost source | Quality pass^5 | Configuration |",
		"| 1 | cheap and weak | pass | 0.20000 | mixed | 9/28 | `h3` |",
		"| 2 | dear and good | pass | 1.00000 | measured | 27/28 | `h1` |",
	} {
		if !strings.Contains(summary, s) {
			t.Errorf("the summary lacks %q:\n%s", s, summary)
		}
	}
}

func TestReasoningIsPerModelOrLabel(t *testing.T) {
	t.Parallel()
	const gemma = "google/gemma-4-26b-a4b-it"
	prices := gemma + "=0.05:0.3"
	models, err := modelsFrom(&modelFlags{
		names: gemma + ",openai/gpt-oss-120b", prices: prices, routes: gemma + "=novita,",
		reasoning: gemma + "@novita=off,openai/gpt-oss-120b=low",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for i := range models {
		got[modelLabel(&models[i])] = models[i].Reasoning
	}
	want := map[string]string{gemma: "", gemma + "@novita": "off", "openai/gpt-oss-120b": "low"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("reasoning by run (-want +got):\n%s", diff)
	}
	byID, err := modelsFrom(&modelFlags{names: gemma, prices: prices, routes: gemma + "=novita", reasoning: gemma + "=high"})
	if err != nil {
		t.Fatal(err)
	}
	if len(byID) != 2 || byID[0].Reasoning != "high" || byID[1].Reasoning != "high" {
		t.Errorf("an entry by id = %+v, want it on every run of the model", byID)
	}
}

func TestBadReasoningEntriesAreUsageErrors(t *testing.T) {
	t.Parallel()
	const gemma = "google/gemma-4-26b-a4b-it"
	for _, bad := range []string{"no-equals", "=off", gemma + "=", gemma + "=max", gemma + "=off," + gemma + "=low", "other/model=off"} {
		if _, err := modelsFrom(&modelFlags{names: gemma, prices: gemma + "=0.05:0.3", reasoning: bad}); err == nil || !strings.Contains(err.Error(), "-reasoning") {
			t.Errorf("-reasoning %q: error = %v, want a usage error naming -reasoning", bad, err)
		}
	}
}

func TestReasoningIsRefusedOnAHostThatIsNotOpenRouter(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	run := runBench(t, f.url, "-case", "R1", "-reasoning", "test/model=off")
	if run.code != exitError || !strings.Contains(run.stderr, "-reasoning") || !strings.Contains(run.stderr, "OpenRouter") {
		t.Errorf("exit %d, stderr %q; want a usage error: the fake's base URL is not OpenRouter's", run.code, run.stderr)
	}
}

func TestTheReasoningControlIsInTheRequestBodyAndTheConfigurationHash(t *testing.T) {
	t.Parallel()
	probe := func(reasoning string) map[string]any {
		return routingProbe(defaultBaseURL, &Model{Reasoning: reasoning})
	}
	if _, ok := probe("")["reasoning"]; ok {
		t.Error("an unset reasoning is sent")
	}
	for reasoning, want := range map[string]map[string]any{
		"off":    {"enabled": false},
		"low":    {"effort": "low"},
		"medium": {"effort": "medium"},
		"high":   {"effort": "high"},
	} {
		if diff := cmp.Diff(want, probe(reasoning)["reasoning"]); diff != "" {
			t.Errorf("reasoning %s (-want +got):\n%s", reasoning, diff)
		}
	}
	if _, ok := routingProbe("https://api.openai.com/v1", &Model{Reasoning: "off"})["reasoning"]; ok {
		t.Error("reasoning goes to a host that is not OpenRouter")
	}
	hash := func(reasoning string) string {
		m := candidateModels()[0]
		m.Reasoning = reasoning
		b, err := NewBench(&Options{Models: []Model{m}, BaseURL: defaultBaseURL, Cases: []string{"A1"}, MaxCalls: 10})
		if err != nil {
			t.Fatal(err)
		}
		_, h, err := b.config(&m)
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	unset, off, low, again := hash(""), hash("off"), hash("low"), hash("off")
	if unset == off || off == low || off != again {
		t.Errorf("hashes unset %s, off %s, low %s, off again %s; want each setting to change the hash and nothing else to", unset, off, low, again)
	}
}

func TestAModelListedAndRoutedRunsOnTheDefaultRoutingToo(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	got := runBench(t, f.url, "-case", "R1", "-route", "test/model=hostA")
	if len(got.report.Models) != 2 || got.report.Models[0].Model != "test/model" || got.report.Models[1].Model != "test/model@hostA" {
		t.Fatalf("models = %+v, want the default run and the routed one", got.report.Models)
	}
	// Each run asked the five questions of its own.
	for _, m := range got.report.Models {
		if m.Questions != 5 {
			t.Errorf("%s was asked %d questions, want 5", m.Model, m.Questions)
		}
	}
	// A model only a route names is asked on the route alone.
	only := runBench(t, f.url, "-case", "R1", "-model", "other/model", "-price", "test/model=1:2,other/model=1:2", "-route", "test/model=hostA")
	if len(only.report.Models) != 2 || only.report.Models[0].Model != "other/model" || only.report.Models[1].Model != "test/model@hostA" {
		t.Errorf("models = %+v, want the listed model and the route-only run", only.report.Models)
	}
}
