package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeTestRoute writes a demo route of one model, test/model, to a new directory
// under dir and returns it.
func writeTestRoute(t *testing.T, dir string) string {
	t.Helper()
	route := filepath.Join(dir, "route")
	if err := os.Mkdir(route, dirMode); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"settings.json": `{"routes":{"ask":["prov/test/model"]}}`,
		"prov.yaml":     "kind: Plugin\nmetadata: {name: prov}\nspec:\n  config:\n    models:\n      - {id: test/model, inputPrice: '1', outputPrice: '2'}\n",
	} {
		if err := os.WriteFile(filepath.Join(route, name), []byte(content), fileMode); err != nil {
			t.Fatal(err)
		}
	}
	return route
}

func TestAWholeRunIsAddedToTheHistoryAndAGivenRouteMemberIsGated(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	dir := t.TempDir()
	route := writeTestRoute(t, dir)
	hist := filepath.Join(dir, "history.json")
	first := runBench(t, f.url, "-history", hist, "-demo-route", route)
	second := runBench(t, f.url, "-history", hist, "-demo-route", route)
	h, err := LoadHistory(hist)
	if err != nil {
		t.Fatal(err)
	}
	runs := h.Runs["test/model"]
	if len(runs) != 2 || runs[0].ConfigHash != first.report.Models[0].ConfigHash || runs[1].Date != time.Now().UTC().Format(time.DateOnly) {
		t.Fatalf("runs = %+v, want the two runs recorded with the report's hash and today's date", runs)
	}
	checkMemberStreakAndExit(t, &second)
}

// checkMemberStreakAndExit checks the second of two runs: the model is a member
// with two runs in its streak, and the run fails exactly when the member did.
func checkMemberStreakAndExit(t *testing.T, run *benchRun) {
	t.Helper()
	m := run.report.Models[0]
	if m.Streak == nil || !m.Streak.Member || m.Streak.Passes+m.Streak.Fails != 2 {
		t.Errorf("streak = %+v, want two runs of a member", m.Streak)
	}
	failed := !m.Verdict.Pass || m.Reliability.Failed
	if failed != (run.code == exitRoute || run.code == exitUnsafe) {
		t.Errorf("exit %d with verdict pass %v: a failing member must fail the run", run.code, m.Verdict.Pass)
	}
}
