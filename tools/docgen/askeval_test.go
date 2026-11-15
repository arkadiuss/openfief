package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const verdictReport = `{"schema":1,"subset":false,"models":[
 {"model":"openai/gpt-oss-120b@crusoe/bf16","modelId":"openai/gpt-oss-120b","route":["crusoe/bf16"],"reasoning":"low",
  "configHash":"0123456789abcdef","cost":"0.02","inputTokens":900,"costPer1kQuestions":"0.13","costSource":"measured",
  "cases":[{"id":"A1","passAll":true},{"id":"A2","passAll":false}],
  "verdict":{"pass":false,"failing":[{"case":"A2","tag":"command","passed":4,"need":5,"runs":5}]},
  "reliability":{"calls":10,"transportFailures":0,"rate":0,"failed":false},"completions":10},
 {"model":"google/gemma-4-31b-it:free","modelId":"google/gemma-4-31b-it:free","configHash":"fedcba9876543210","cost":"0","inputTokens":0,
  "cases":[{"id":"A1","passAll":false}],"completions":0}]}`

const oldReport = `{"schema":1,"subset":true,"models":[
 {"model":"qwen/qwen3.7-flash","configHash":"aa00bb11cc22dd33","cost":"0.014","inputTokens":284532,
  "cases":[{"id":"A1","passAll":true},{"id":"R1","passAll":false}]}]}`

func trimmed(t *testing.T, report string) []evalResult {
	t.Helper()
	rs, err := trimReport([]byte(report), "9-1", "2026-11-14")
	if err != nil {
		t.Fatalf("trimReport() error = %v", err)
	}
	return rs
}

func TestTrimReportKeepsHostFailingCasesAndReliability(t *testing.T) {
	t.Parallel()
	r := trimmed(t, verdictReport)[0]
	got := []string{r.Model, r.Host, r.Verdict, r.Failing[0].Case, r.CostPer1k, r.Date}
	want := []string{"openai/gpt-oss-120b", "crusoe/bf16", verdictFail, "A2", "0.13", "2026-11-14"}
	if !slices.Equal(got, want) || r.Reliability == nil || !r.Answered {
		t.Errorf("result = %+v", r)
	}
}

func TestTrimReportOfAModelNoCallAnsweredHasNoVerdict(t *testing.T) {
	t.Parallel()
	if r := trimmed(t, verdictReport)[1]; r.Answered || r.Verdict != verdictNone {
		t.Errorf("result = %+v", r)
	}
}

func TestTrimReportOfAnOldSubsetReportIsDiagnosticWithNoVerdict(t *testing.T) {
	t.Parallel()
	r := trimmed(t, oldReport)[0]
	if r.Verdict != verdictNone || r.UnderBar != 1 || len(r.Diagnostic) != 2 || r.CostSource != "estimated" || r.Host != "" {
		t.Errorf("result = %+v", r)
	}
}

func TestRunAskEvalImportWritesASortableFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	report := filepath.Join(dir, "report.json")
	if err := os.WriteFile(report, []byte(verdictReport), 0o600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := runAskEval([]string{"import", "-report", report, "-run", "9-1", "-date", "2026-11-14", "-out", dir}, &out)
	if err != nil {
		t.Fatalf("runAskEval() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "2026-11-14-09-1.json")); err != nil {
		t.Errorf("result file: %v", err)
	}
	if err := runAskEval([]string{"import", "-report", report, "-run", "x", "-date", "2026-11-14"}, &out); err == nil {
		t.Error("runAskEval() accepted a bad run name")
	}
}

func TestAskEvalPage(t *testing.T) {
	t.Parallel()
	results := []evalResult{
		{Run: "7", Date: "2026-11-14", Model: "m/a", Host: "h/bf16", Reasoning: "low", Verdict: verdictPass, Answered: true, Cases: 29, CostPer1k: "0.127", CostSource: "measured", ConfigHash: "0123456789abcdef", BenchSchema: 1},
		{Run: "6", Date: "2026-11-14", Model: "m/b", Verdict: verdictNone, Answered: true, Cases: 28, UnderBar: 4, ConfigHash: "ffff0000ffff0000", BenchSchema: 1},
		{Run: "3", Date: "2026-11-14", Model: "m/c", Verdict: verdictNone, Diagnostic: []string{"R1"}, ConfigHash: "abcd", BenchSchema: 1},
	}
	page := askEvalPage(results)
	for _, want := range []string{"title: How models did in the Ask box evaluation", "synthetic", "m/a@h/bf16", "`01234567`", "4 of 28 cases under pass^5", "diagnostic: R1 only", "no answer"} {
		if !strings.Contains(page, want) {
			t.Errorf("page lacks %q", want)
		}
	}
	standing, _, _ := strings.Cut(page, "## Every run")
	if strings.Contains(standing, "m/b") || strings.Contains(standing, "m/c") {
		t.Errorf("the standing table lists a run with no verdict: %s", standing)
	}
}

func TestCommittedResultsLoadNewestFirst(t *testing.T) {
	t.Parallel()
	results, err := loadResults(os.DirFS("../.."))
	if err != nil {
		t.Fatalf("loadResults() error = %v", err)
	}
	if len(results) == 0 {
		t.Fatal("no results committed")
	}
	for i := range results {
		r := &results[i]
		if r.Model == "" || r.Date == "" || r.ConfigHash == "" || r.Verdict == "" {
			t.Errorf("result %d is incomplete: %+v", i, r)
		}
		if i > 0 && results[i-1].Date < r.Date {
			t.Errorf("results are not newest first at %d", i)
		}
	}
}
