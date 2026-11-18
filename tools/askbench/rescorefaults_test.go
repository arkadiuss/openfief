package main

import (
	"slices"
	"strings"
	"testing"
)

// legacyReliabilityReport is a report of a run of R1 as the bench wrote it
// before scorer version 3: five questions, four of which a failed call ended,
// with no fault stored for any 4xx and none of the new fields.
func legacyReliabilityReport(t *testing.T) *Report {
	t.Helper()
	r := reportOf(t, "R1")
	r.Schema, r.ScorerVersion = 2, 2
	runs := r.Models[0].Cases[0].Detail
	ended := func(run *RunReport, ce CallErrorReport) {
		*run = RunReport{Run: run.Run, Errored: true, Attempts: 1, CallErrors: []CallErrorReport{ce}}
	}
	// A timeout that its retry did not mend, as the old bench counted it.
	runs[0] = RunReport{
		Run: 1, Errored: true, Attempts: 2, TransportFailures: 2, Retries: []RetryReport{{Label: labelTimeout}},
		CallErrors: []CallErrorReport{{Label: labelTimeout}},
	}
	ended(&runs[1], CallErrorReport{Status: 402, Label: "no credits"})
	ended(&runs[2], CallErrorReport{Status: 400, Label: "request rejected", Class: classRejected})
	ended(&runs[3], CallErrorReport{Status: 404, Label: "request rejected", Class: classRejected})
	return r
}

// checkDerivedFaults checks the faults a re-grading derived for the four runs of
// legacyReliabilityReport that a failed call ended, and which of them are void.
func checkDerivedFaults(t *testing.T, runs []RunReport) {
	t.Helper()
	got := make([]string, 4)
	for i := range got {
		ce := &runs[i].CallErrors[0]
		got[i] = ce.Fault + "/" + ce.Rule
	}
	want := []string{"host/", "credentials/" + ruleCredentials, "/", "host/" + ruleNotFound}
	if !slices.Equal(got, want) {
		t.Errorf("derived faults = %q, want %q", got, want)
	}
	if !runs[1].Void || runs[1].VoidReason != voidCredentials || runs[0].Void || runs[2].Void {
		t.Errorf("void = %v %v %v, want only the refused credits", runs[0].Void, runs[1].Void, runs[2].Void)
	}
}

func TestReGradingCountsReliabilityPerQuestionFromTheStoredCallFailures(t *testing.T) {
	t.Parallel()
	r := legacyReliabilityReport(t)
	hash := r.Models[0].ConfigHash
	if err := Rescore(t.Context(), r); err != nil {
		t.Fatal(err)
	}
	m := &r.Models[0]
	checkDerivedFaults(t, m.Cases[0].Detail)
	// Four questions count: the credits' run is void and the 400 is undecided, so
	// it is no host's failure; the timeout and the 404 are.
	want := Reliability{Questions: 4, FailedQuestions: 2, Calls: 6, TransportFailures: 3, Retried: 1, Rate: 0.5, Failed: true}
	if m.Reliability != want || m.VoidRuns != 1 || m.VoidCredentials != 1 {
		t.Errorf("reliability = %+v, void %d/%d, want %+v and one void run", m.Reliability, m.VoidRuns, m.VoidCredentials, want)
	}
	if r.Rescore.DerivedFaults != 3 || r.Rescore.Undecided != 1 {
		t.Errorf("rescore = %+v, want three derived faults and one 4xx it could not decide", r.Rescore)
	}
	if got := []int{r.Schema, r.ScorerVersion, exitCode(r)}; !slices.Equal(got, []int{reportSchema, ScorerVersion, exitCredentials}) || m.ConfigHash != hash {
		t.Errorf("schema, scorer and exit = %v, hash kept %v", got, m.ConfigHash == hash)
	}
	checkAll(t, r.Summary(), "counted again per question", "which it did not store", "of questions, 2 of 4; 3 failed calls")
	checkAll(t, rescoreText(r, "out"), "faults: 3 failed calls", "1 4xx could not be decided")
}

// checkAll checks that doc holds every text.
func checkAll(t *testing.T, doc string, texts ...string) {
	t.Helper()
	for _, text := range texts {
		if !strings.Contains(doc, text) {
			t.Errorf("the text lacks %q:\n%s", text, doc)
		}
	}
}

func TestReGradingAReportOfTheBenchKeepsItsReliability(t *testing.T) {
	t.Parallel()
	r := reportOf(t, "R1")
	before := r.Models[0].Reliability
	if err := Rescore(t.Context(), r); err != nil {
		t.Fatal(err)
	}
	if r.Models[0].Reliability != before || r.Rescore.DerivedFaults != 0 || r.Rescore.Undecided != 0 {
		t.Errorf("reliability = %+v, rescore %+v, want %+v and nothing derived", r.Models[0].Reliability, r.Rescore, before)
	}
}
