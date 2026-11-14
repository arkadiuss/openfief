package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// run is a history entry for a test: a pass or a fail of the hash.
func hrun(pass bool, hash string) HistoryEntry {
	e := HistoryEntry{Date: "2026-11-01", Verdict: "fail", Reliability: verdictPass, ConfigHash: hash}
	if pass {
		e.Verdict = verdictPass
	}
	return e
}

func historyOf(label string, runs ...HistoryEntry) *History {
	h := NewHistory()
	for i := range runs {
		h.Add(label, &runs[i])
	}
	return h
}

func TestTheStreakCountsTheLatestRunsThatAgree(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		runs     []HistoryEntry
		want     Streak
		eligible bool
		failing  bool
	}{
		{"no runs", nil, Streak{}, false, false},
		{"one pass", []HistoryEntry{hrun(true, "a")}, Streak{Passes: 1}, false, false},
		{"pass, pass is eligible", []HistoryEntry{hrun(true, "a"), hrun(true, "a")}, Streak{Passes: 2}, true, false},
		{"fail, pass restarts at one pass", []HistoryEntry{hrun(false, "a"), hrun(true, "a")}, Streak{Passes: 1}, false, false},
		{"pass, fail is one fail", []HistoryEntry{hrun(true, "a"), hrun(false, "a")}, Streak{Fails: 1}, false, false},
		{"fail, fail is flagged", []HistoryEntry{hrun(true, "a"), hrun(false, "a"), hrun(false, "a")}, Streak{Fails: 2}, false, true},
		{"a configuration change resets the passes", []HistoryEntry{hrun(true, "a"), hrun(true, "a"), hrun(true, "b")}, Streak{Passes: 1}, false, false},
		{"a configuration change resets the fails", []HistoryEntry{hrun(false, "a"), hrun(false, "a"), hrun(false, "b")}, Streak{Fails: 1}, false, false},
		{"two passes of the new configuration are eligible", []HistoryEntry{hrun(true, "a"), hrun(true, "b"), hrun(true, "b")}, Streak{Passes: 2}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := historyOf("m", tt.runs...).StreakOf("m")
			if s != tt.want || s.Eligible() != tt.eligible || s.Failing() != tt.failing {
				t.Errorf("streak = %+v eligible %v failing %v, want %+v %v %v", s, s.Eligible(), s.Failing(), tt.want, tt.eligible, tt.failing)
			}
		})
	}
}

func TestARunPassesOnlyWhenTheVerdictAndTheReliabilityBoth(t *testing.T) {
	t.Parallel()
	flaky := hrun(true, "a")
	flaky.Reliability = "fail"
	if s := historyOf("m", hrun(true, "a"), flaky).StreakOf("m"); s != (Streak{Fails: 1}) {
		t.Errorf("streak = %+v, want one fail: a passing verdict on an unreliable host is not a pass", s)
	}
}

func TestTheHistoryKeepsTheNewestRunsOfALabel(t *testing.T) {
	t.Parallel()
	h := NewHistory()
	for i := range historyKeep + 10 {
		h.Add("m", &HistoryEntry{Date: strconv.Itoa(i), Verdict: verdictPass, Reliability: verdictPass, ConfigHash: "a"})
	}
	runs := h.Runs["m"]
	if len(runs) != historyKeep || runs[len(runs)-1].Date != strconv.Itoa(historyKeep+9) || runs[0].Date != "10" {
		t.Errorf("kept %d runs from %s to %s", len(runs), runs[0].Date, runs[len(runs)-1].Date)
	}
}

func TestTheHistoryRoundTripsAndAMissingFileIsEmpty(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "history.json")
	empty, err := LoadHistory(path)
	if err != nil || len(empty.Runs) != 0 {
		t.Fatalf("missing file: %+v, %v", empty, err)
	}
	h := historyOf("google/gemma@nextbit/bf16", hrun(true, "abc"), hrun(false, "abc"))
	h.Add("other", &HistoryEntry{Date: "2026-11-02", Verdict: verdictPass, Reliability: verdictPass, ConfigHash: "def"})
	if saveErr := h.Save(path); saveErr != nil {
		t.Fatal(saveErr)
	}
	got, err := LoadHistory(path)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(h, got); diff != "" {
		t.Errorf("history (-want +got):\n%s", diff)
	}
	raw, err := os.ReadFile(filepath.Clean(path))
	if err != nil || !strings.Contains(string(raw), `"configHash": "abc"`) || !strings.HasSuffix(string(raw), "}\n") {
		t.Errorf("the file = %q, %v", raw, err)
	}
}

func TestABadHistoryFileIsAnErrorNotAnEmptyHistory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for name, content := range map[string]string{"junk.json": "not json", "schema.json": `{"schema":9,"runs":{}}`} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), fileMode); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadHistory(path); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

// modelRep is a model's report for a test: its verdict, reliability and hash.
func modelRep(label, hash string, pass, reliable bool) ModelReport {
	m := ModelReport{Model: label, ConfigHash: hash, Verdict: Verdict{Pass: pass}}
	if !pass {
		m.Verdict.Failing = []VerdictFailure{{Case: "A1", Tag: tagCommand, Passed: 3, Need: 5, Runs: 5}}
	}
	m.Reliability = Reliability{Calls: 100, TransportFailures: 1, Rate: 0.01}
	if !reliable {
		m.Reliability = Reliability{Calls: 100, TransportFailures: 9, Rate: 0.09, Failed: true}
	}
	return m
}

func TestTwoPassingRunsMakeAModelEligibleAndTwoFailingOnesFlagAMemberForRemoval(t *testing.T) {
	t.Parallel()
	h := historyOf("joiner", hrun(true, "j1"))
	h.Add("member", &HistoryEntry{Date: "2026-11-01", Verdict: "fail", Reliability: verdictPass, ConfigHash: "m1"})
	r := &Report{Models: []ModelReport{modelRep("joiner", "j1", true, true), modelRep("member", "m1", false, true), modelRep("new", "n1", true, true)}}
	r.applyHistory(h, []string{"member"}, "2026-11-08")
	if strings.Join(r.Route.Joinable, ",") != "joiner" || strings.Join(r.Route.Eligible, ",") != "joiner" {
		t.Errorf("eligible %v joinable %v, want joiner only: new has one pass", r.Route.Eligible, r.Route.Joinable)
	}
	if strings.Join(r.Route.Removable, ",") != "member" {
		t.Errorf("removable = %v, want member: it failed two runs in a row", r.Route.Removable)
	}
	if got, want := r.Models[0].Streak, (&StreakReport{Passes: 2, Eligible: true}); got == nil || *got != *want {
		t.Errorf("joiner streak = %+v, want %+v", got, want)
	}
	if got, want := r.Models[1].Streak, (&StreakReport{Fails: 2, Member: true}); got == nil || *got != *want {
		t.Errorf("member streak = %+v, want %+v", got, want)
	}
	checkRecordedRun(t, h, r.Route)
}

// checkRecordedRun checks that the run of "new" was added with the day, the
// verdict and the hash, and that the route report says so.
func checkRecordedRun(t *testing.T, h *History, route *RouteReport) {
	t.Helper()
	runs := h.Runs["new"]
	if len(runs) != 1 || runs[0].Date != "2026-11-08" || runs[0].Verdict != verdictPass || runs[0].ConfigHash != "n1" {
		t.Errorf("new's runs = %+v, want this run recorded", runs)
	}
	if !route.Recorded {
		t.Error("the route report does not say the run was recorded")
	}
}

func TestAConfigurationChangeStartsTheStreakAgainInTheReport(t *testing.T) {
	t.Parallel()
	h := historyOf("m", hrun(true, "old"), hrun(true, "old"))
	r := &Report{Models: []ModelReport{modelRep("m", "new", true, true)}}
	r.applyHistory(h, nil, "2026-11-08")
	if s := r.Models[0].Streak; s.Passes != 1 || s.Eligible || len(r.Route.Eligible) != 0 {
		t.Errorf("streak = %+v eligible %v, want one pass: the configuration changed", s, r.Route.Eligible)
	}
}

func TestOnlyAWholeRunThatRanIsRecorded(t *testing.T) {
	t.Parallel()
	could := modelRep("could not", "h", false, true)
	could.CouldNotRun = true
	tests := []struct {
		name   string
		report Report
		want   int
	}{
		{"whole", Report{Models: []ModelReport{modelRep("m", "h", true, true)}}, 1},
		{"subset", Report{Subset: true, Models: []ModelReport{modelRep("m", "h", true, true)}}, 0},
		{"budget stopped", Report{Budget: BudgetReport{Aborted: true}, Models: []ModelReport{modelRep("m", "h", true, true)}}, 0},
		{"could not run", Report{Models: []ModelReport{could}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := NewHistory()
			tt.report.applyHistory(h, nil, "2026-11-08")
			n := 0
			for _, runs := range h.Runs {
				n += len(runs)
			}
			if n != tt.want || tt.report.Route.Recorded != (tt.want > 0) {
				t.Errorf("recorded %d runs (route says %v), want %d", n, tt.report.Route.Recorded, tt.want)
			}
		})
	}
}

func TestTheGateFailsAMemberWhoseVerdictOrReliabilityFailsAndNoOther(t *testing.T) {
	t.Parallel()
	r := &Report{Models: []ModelReport{
		modelRep("good", "h", true, true),
		modelRep("weak", "h", false, true),
		modelRep("flaky", "h", true, false),
		modelRep("both", "h", false, false),
		modelRep("outsider", "h", false, false),
	}}
	got := routeGate(r, []string{"good", "weak", "flaky", "both"})
	names := make([]string, len(got))
	reasons := map[string]string{}
	for i, f := range got {
		names[i], reasons[f.Model] = f.Model, f.Reason
	}
	if strings.Join(names, ",") != "both,flaky,weak" {
		t.Fatalf("failing = %v, want both, flaky and weak, not good or the outsider", names)
	}
	for model, wants := range map[string][2]string{
		"weak":  {"verdict fail: A1 (command) 3/5", "reliability"},
		"flaky": {"reliability fail (9.0%, 9/100)", "verdict"},
	} {
		if !strings.Contains(reasons[model], wants[0]) || strings.Contains(reasons[model], wants[1]) {
			t.Errorf("%s: %q, want %q and not %q", model, reasons[model], wants[0], wants[1])
		}
	}
	if !strings.Contains(reasons["both"], "verdict") || !strings.Contains(reasons["both"], "reliability") {
		t.Errorf("both: %q", reasons["both"])
	}
	if subset := (&Report{Subset: true, Models: r.Models}); len(routeGate(subset, []string{"weak"})) != 0 {
		t.Error("a run of some cases is gated")
	}
	if len(routeGate(r, nil)) != 0 {
		t.Error("a run with no route members is gated")
	}
}

func TestAFailingMemberIsExitFourAndSafetyAndTheBudgetComeFirst(t *testing.T) {
	t.Parallel()
	failing := &Report{Models: []ModelReport{modelRep("m", "h", false, true)}}
	failing.Models[0].SafetyPassed = true
	failing.applyHistory(nil, []string{"m"}, "2026-11-08")
	if got := exitCode(failing); got != exitRoute {
		t.Errorf("exit = %d, want %d", got, exitRoute)
	}
	unsafe := &Report{Models: []ModelReport{modelRep("m", "h", false, true)}}
	unsafe.applyHistory(nil, []string{"m"}, "2026-11-08")
	if got := exitCode(unsafe); got != exitUnsafe {
		t.Errorf("exit = %d, want %d: a failed safety property is the stronger signal", got, exitUnsafe)
	}
	passing := &Report{Models: []ModelReport{modelRep("m", "h", true, true)}}
	passing.Models[0].SafetyPassed = true
	passing.applyHistory(nil, []string{"m"}, "2026-11-08")
	if got := exitCode(passing); got != exitOK {
		t.Errorf("exit = %d, want %d", got, exitOK)
	}
	if got := exitCode(&Report{Models: passing.Models}); got != exitOK {
		t.Errorf("a report with no route section exits %d, want %d", got, exitOK)
	}
}

func TestTheSummaryShowsTheStreakAndTheRoute(t *testing.T) {
	t.Parallel()
	h := historyOf("member", hrun(false, "mmmmmmmmmmmm"))
	r := &Report{Models: []ModelReport{modelRep("joiner", "jjjjjjjjjjjj", true, true), modelRep("member", "mmmmmmmmmmmm", false, true)}}
	h.Add("joiner", &HistoryEntry{Date: "2026-11-01", Verdict: verdictPass, Reliability: verdictPass, ConfigHash: "jjjjjjjjjjjj"})
	r.applyHistory(h, []string{"member"}, "2026-11-08")
	summary := r.Summary()
	for _, want := range []string{
		"Consecutive runs |", "| 2 consecutive passes (eligible for the route) |", "| 2 consecutive fails |", "## Demo route (D-221)",
		"ROUTE MEMBER FAILED: member: verdict fail", "remove in a reviewed change): member", "eligible to join the route (2 consecutive passes): joiner", "never edits it",
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary lacks %q:\n%s", want, summary)
		}
	}
	text := resultText(r, "out")
	if !strings.Contains(text, "ROUTE MEMBER FAILED: member") || !strings.Contains(text, "eligible to join the route") {
		t.Errorf("console text = %q", text)
	}
}

func TestTheDemosRouteIsReadFromItsConfigAsTheLabelsTheReportKeysBy(t *testing.T) {
	t.Parallel()
	models, err := LoadDemoRoute("../../deploy/demo/model")
	if err != nil {
		t.Fatal(err)
	}
	if len(models) == 0 {
		t.Fatal("no route members")
	}
	m := models[0]
	if got := modelLabel(&m); got != "google/gemma-4-26b-a4b-it@nextbit/bf16" || !m.Structured || m.InputPrice.Sign() == 0 || m.OutputPrice.Sign() == 0 {
		t.Errorf("first member = %q structured %v prices %v/%v", got, m.Structured, m.InputPrice, m.OutputPrice)
	}
}

func TestADemoRouteThatNamesNoEntryIsAUsageError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), fileMode); err != nil {
			t.Fatal(err)
		}
	}
	write("settings.json", `{"routes":{"ask":["prov/a/model"]}}`)
	write("prov.yaml", "kind: Plugin\nmetadata: {name: prov}\nspec:\n  config:\n    models:\n      - {id: b/other, inputPrice: '1', outputPrice: '2'}\n")
	if _, err := LoadDemoRoute(dir); err == nil || !strings.Contains(err.Error(), `"prov/a/model"`) {
		t.Errorf("err = %v, want the entry named", err)
	}
	write("prov.yaml", "kind: Plugin\nmetadata: {name: prov}\nspec:\n  config:\n    models:\n      - {id: a/model, inputPrice: '1', outputPrice: '2', hosts: [h1, h2], reasoning: low}\n---\nkind: Service\nmetadata: {name: ignored}\n")
	models, err := LoadDemoRoute(dir)
	if err != nil || len(models) != 1 || modelLabel(&models[0]) != "a/model@h1|h2" || models[0].Reasoning != "low" {
		t.Errorf("models = %+v, err %v", models, err)
	}
}

func TestTheHistoryIsWrittenOnlyWhenAskedAndTheRouteFlagReadsTheDemo(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	dir := t.TempDir()
	untouched := filepath.Join(dir, "history.json")
	// No -history: nothing is written, whatever else runs.
	runBench(t, f.url, "-case", "R1")
	if _, err := os.Stat(untouched); err == nil {
		t.Fatal("a history appeared without -history")
	}
	hist := filepath.Join(dir, "asked.json")
	got := runBench(t, f.url, "-case", "R1", "-history", hist)
	if got.report.Route == nil || got.report.Route.Recorded {
		t.Errorf("route = %+v, want a subset run not recorded", got.report.Route)
	}
	h, err := LoadHistory(hist)
	if err != nil || len(h.Runs) != 0 {
		t.Errorf("history = %+v, %v, want empty after a run of some cases", h, err)
	}
	if !strings.Contains(got.summary(t), "not in the history") {
		t.Errorf("summary does not say the run was not recorded:\n%s", got.summary(t))
	}
	bad := runBench(t, f.url, "-case", "R1", "-demo-route", filepath.Join(dir, "nowhere"))
	if bad.code != exitError {
		t.Errorf("exit = %d, want %d for a route that cannot be read", bad.code, exitError)
	}
}
