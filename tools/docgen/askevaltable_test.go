package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func standing(model, host, run string, passed bool, failing []evalFailing, cost string) evalResult {
	verdict := verdictPass
	if !passed {
		verdict = verdictFail
	}
	return evalResult{
		Run: run, Date: "2026-11-15", Model: model, Host: host, Verdict: verdict, Failing: failing, Cases: 29,
		Answered: true, CostPer1k: cost, CostSource: "measured", ScorerVersion: 2, ConfigHash: "0123456789abcdef",
		Reliability: &evalReliability{Calls: 100, TransportFailures: 2, Rate: 0.02},
	}
}

func labels(rows []*evalResult) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.label()
	}
	return out
}

func TestStandingRunsSortPassesFirstThenCasesPassedThenCost(t *testing.T) {
	t.Parallel()
	miss := func(n int) []evalFailing {
		f := make([]evalFailing, n)
		for i := range f {
			f[i] = evalFailing{Case: "A1", Tag: "command", Passed: 4, Need: 5}
		}
		return f
	}
	results := []evalResult{
		standing("m/failing-close", "h", "1", false, miss(1), "0.01"),
		standing("m/pass-dear", "h", "1", true, nil, "0.30"),
		standing("m/failing-far", "h", "1", false, miss(3), "0.01"),
		standing("m/pass-cheap", "h", "1", true, nil, "0.10"),
		standing("m/failing-close-cheaper", "h", "1", false, miss(1), "0.005"),
		standing("m/pass-no-cost", "h", "1", true, nil, ""),
	}
	got := labels(standingRuns(results))
	want := []string{
		"m/pass-cheap@h", "m/pass-dear@h", "m/pass-no-cost@h",
		"m/failing-close-cheaper@h", "m/failing-close@h", "m/failing-far@h",
	}
	if !slices.Equal(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

func TestStandingRunsKeepOnlyTheLatestFullRunWithAVerdictOfEachModelAndHost(t *testing.T) {
	t.Parallel()
	latest := standing("m/a", "h", "5", true, nil, "0.1")
	older := standing("m/a", "h", "4", false, nil, "0.1")
	diagnostic := standing("m/a", "h", "6", true, nil, "0.1")
	diagnostic.Diagnostic = []string{"R1"}
	noAnswer := standing("m/b", "", "6", true, nil, "")
	noAnswer.Answered = false
	noVerdict := standing("m/c", "", "6", false, nil, "")
	noVerdict.Verdict = verdictNone
	// Newest first, as loadResults returns them.
	rows := standingRuns([]evalResult{diagnostic, noAnswer, noVerdict, latest, older})
	if len(rows) != 1 || rows[0].Run != "5" {
		t.Fatalf("rows = %v, want only run 5 of m/a@h", labels(rows))
	}
}

func TestResultsTableShowsEveryColumnAndAComment(t *testing.T) {
	t.Parallel()
	safetyMiss := []evalFailing{{Case: "R1", Tag: "safety", Passed: 4, Need: 5}, {Case: "A13", Tag: "command", Passed: 3, Need: 5}}
	results := []evalResult{
		standing("m/a", "crusoe/bf16", "16-2", false, safetyMiss, "0.14667"),
		standing("m/b", "", "7", true, nil, ""),
	}
	history := evalHistory{
		"m/b (default routing)": {
			{Verdict: verdictFail, Reliability: verdictPass, ConfigHash: "x"},
			{Verdict: verdictPass, Reliability: verdictPass, ConfigHash: "y"},
			{Verdict: verdictPass, Reliability: verdictPass, ConfigHash: "y"},
		},
	}
	comments := evalComments{"m/a@crusoe": "passed 14-2; then | single misses"}
	table := resultsTable(results, history, comments)
	for _, want := range []string{
		"| Model | Host | Verdict | Cases passed | Safety | Reliability | USD per 1k questions | Consecutive passes | Last run | Comment |",
		"| m/b | default routing | pass (scorer 2) | 29/29 | 9/9 | 2.0% | n/a | 2 | 2026-11-15 (run 7) |  |",
		"| m/a | crusoe/bf16 | fail (scorer 2) | 27/29 | 8/9 | 2.0% | 0.147 | n/a | 2026-11-15 (run 16-2) | passed 14-2; then \\| single misses |",
	} {
		if !strings.Contains(table, want) {
			t.Errorf("table lacks %q:\n%s", want, table)
		}
	}
}

func TestCommentLookupPrefersTheMostSpecificKeyAndAMissingOneIsEmpty(t *testing.T) {
	t.Parallel()
	r := &evalResult{Model: "m/a", Host: "crusoe/bf16"}
	for name, tc := range map[string]struct {
		comments evalComments
		want     string
	}{
		"full label":    {evalComments{"m/a@crusoe/bf16": "full", "m/a@crusoe": "provider", "m/a": "model"}, "full"},
		"provider slug": {evalComments{"m/a@crusoe": "provider", "m/a": "model"}, "provider"},
		"model alone":   {evalComments{"m/a": " model "}, "model"},
		"other host":    {evalComments{"m/a@deepinfra": "elsewhere"}, ""},
		"none":          {nil, ""},
	} {
		if got := tc.comments.of(r); got != tc.want {
			t.Errorf("%s: comment = %q, want %q", name, got, tc.want)
		}
	}
	if got := (evalComments{"m/a@crusoe": "x"}).of(&evalResult{Model: "m/a"}); got != "" {
		t.Errorf("a default-routing row took a host's comment: %q", got)
	}
}

func TestConsecutivePassesStopAtAFailureOrAnotherConfiguration(t *testing.T) {
	t.Parallel()
	pass := func(hash string) historyEntry {
		return historyEntry{Verdict: verdictPass, Reliability: verdictPass, ConfigHash: hash}
	}
	failedReliability := historyEntry{Verdict: verdictPass, Reliability: verdictFail, ConfigHash: "b"}
	history := evalHistory{
		"two":      {pass("a"), pass("b"), pass("b")},
		"failing":  {pass("a"), pass("a"), {Verdict: verdictFail, Reliability: verdictPass, ConfigHash: "a"}},
		"slow":     {pass("b"), failedReliability},
		"rerun":    {pass("b"), failedReliability, pass("b")},
		"unproven": {},
	}
	for label, want := range map[string]int{"two": 2, "failing": 0, "slow": 0, "rerun": 1} {
		if got, ok := history.consecutivePasses(label); !ok || got != want {
			t.Errorf("%s: passes = %d (ok %v), want %d", label, got, ok, want)
		}
	}
	if _, ok := history.consecutivePasses("unproven"); ok {
		t.Error("a label with no runs has a count")
	}
}

func TestWithTableReplacesOnlyTheFencedText(t *testing.T) {
	t.Parallel()
	guide := "before\n" + tableBegin + "\nold\n" + tableEnd + "\nafter\n"
	got, err := withTable(guide, "| new |\n")
	if err != nil {
		t.Fatal(err)
	}
	want := "before\n" + tableBegin + "\n\n| new |\n\n" + tableEnd + "\nafter\n"
	if got != want {
		t.Errorf("guide = %q, want %q", got, want)
	}
	if again, againErr := withTable(got, "| new |\n"); againErr != nil || again != got {
		t.Error("rewriting is not stable")
	}
	if _, err := withTable("no fences", "x"); err == nil {
		t.Error("withTable() accepted a guide without fences")
	}
}

func TestVoidRunsReadTheSameOnBothPages(t *testing.T) {
	t.Parallel()
	noAnswer := standing("m/a", "", "3", false, nil, "")
	noAnswer.Answered = false
	noVerdict := standing("m/b", "", "6", false, nil, "")
	noVerdict.Verdict = verdictNone
	graded := standing("m/c", "h", "9", false, nil, "0.1")
	graded.ScorerVersion = 0
	page := askEvalPage([]evalResult{noAnswer, noVerdict, graded})
	for _, r := range []*evalResult{&noAnswer, &noVerdict, &graded} {
		if want := "| " + verdictText(r) + " |"; !strings.Contains(page, want) {
			t.Errorf("the results page lacks %q for %s", want, r.label())
		}
	}
	if !noAnswer.void() || !noVerdict.void() || graded.void() {
		t.Error("void() is wrong")
	}
	if got := verdictText(&graded); got != "fail (scorer 1, before versioning)" {
		t.Errorf("verdictText() = %q", got)
	}
	if table := resultsTable([]evalResult{noAnswer, noVerdict}, nil, nil); strings.Contains(table, "m/a") || strings.Contains(table, "m/b") {
		t.Errorf("the table lists a void run:\n%s", table)
	}
}

func TestSafetyCasesMatchTheBenchsCases(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../askbench/testdata/cases.yaml")
	if err != nil {
		t.Fatal(err)
	}
	n := len(regexp.MustCompile(`(?m)^\s*tag:\s*safety\s*$`).FindAll(data, -1))
	if n != safetyCases {
		t.Errorf("the bench has %d safety cases, the table assumes %d", n, safetyCases)
	}
}

func TestCommittedGuideTableIsCurrentAndCommentsAreOneLine(t *testing.T) {
	t.Parallel()
	repo := os.DirFS("../..")
	var out strings.Builder
	if err := runAskEvalTable(repo, []string{"-check"}, &out); err != nil {
		t.Fatalf("the committed guide's table is stale: %v", err)
	}
	comments, err := loadComments(repo)
	if err != nil {
		t.Fatal(err)
	}
	for key, text := range comments {
		if strings.ContainsAny(text, "\n") || text == "" {
			t.Errorf("comment %q is not one line: %q", key, text)
		}
	}
}

func TestStaleGuideTableFailsTheCheck(t *testing.T) {
	t.Parallel()
	results := `[{"run":"1","date":"2026-11-15","model":"m/a","verdict":"pass","cases":29,"answered":true,"configHash":"0123456789abcdef","benchSchema":2}]`
	repo := fstest.MapFS{
		resultsDir + "/2026-11-15-01.json": {Data: []byte(results)},
		guideFile:                          {Data: []byte("x\n" + tableBegin + "\n\nstale\n" + tableEnd + "\n")},
	}
	if err := runAskEvalTable(repo, []string{"-check"}, &strings.Builder{}); err == nil {
		t.Error("a stale table passed the check")
	}
}

func TestARunAtAnotherParallelismIsMarkedOnBothPagesAndImportRecordsIt(t *testing.T) {
	t.Parallel()
	slow := standing("m/a", "", "20", true, nil, "0.1")
	slow.Parallel = 1
	usual := standing("m/b", "", "20", true, nil, "0.2")
	if got := verdictText(&slow); got != "pass (scorer 2, parallel 1)" {
		t.Errorf("verdictText() = %q", got)
	}
	if got := verdictText(&usual); got != "pass (scorer 2)" {
		t.Errorf("verdictText() = %q", got)
	}
	if table := resultsTable([]evalResult{slow, usual}, nil, nil); !strings.Contains(table, "pass (scorer 2, parallel 1)") {
		t.Errorf("the table lacks the parallelism:\n%s", table)
	}
	dir := t.TempDir()
	report := filepath.Join(dir, "report.json")
	if err := os.WriteFile(report, []byte(verdictReport), 0o600); err != nil {
		t.Fatal(err)
	}
	for parallel, wantField := range map[string]bool{"1": true, "3": false} {
		args := []string{"import", "-report", report, "-run", "21", "-date", "2026-11-16", "-out", dir, "-parallel", parallel}
		if err := runAskEval(args, &strings.Builder{}); err != nil {
			t.Fatal(err)
		}
		data, err := fs.ReadFile(os.DirFS(dir), "2026-11-16-21.json")
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Contains(string(data), `"parallel"`); got != wantField {
			t.Errorf("-parallel %s: file has the field = %v, want %v", parallel, got, wantField)
		}
	}
	zero := []string{"import", "-report", report, "-run", "21", "-date", "2026-11-16", "-parallel", "0"}
	if err := runAskEval(zero, &strings.Builder{}); err == nil {
		t.Error("runAskEval() accepted -parallel 0")
	}
}
