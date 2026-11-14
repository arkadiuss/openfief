package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestStatusLabels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status int
		want   string
	}{
		{402, "no credits"},
		{401, "key refused"},
		{429, "rate limited"},
		{500, "provider error"},
		{502, "provider error"},
		{503, "provider error"},
		{400, "request rejected"},
		{404, "request rejected"},
		{0, ""},
		{302, ""},
	}
	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.status), func(t *testing.T) {
			t.Parallel()
			if got := statusLabel(tt.status); got != tt.want {
				t.Errorf("statusLabel(%d) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestACallFailureIsItsStatusAndALabel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status int
		want   string
	}{
		{"no credits", 402, "no credits"},
		{"key refused", 401, "key refused"},
		{"rate limited", 429, "rate limited"},
		{"provider error", 500, "provider error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newFakeModel(t)
			f.status, f.failureBody = tt.status, `{"error":"provider words Bearer `+testKey+`"}`
			mr := &modelRun{client: newChatClient(http.DefaultClient, f.url, testKey), model: &Model{ID: "m"}, meter: newMeter(decimalOf("1"), 10), now: time.Now}
			_, err := mr.Generate(t.Context(), nil, nil, 10)
			if err == nil || len(mr.calls) != 1 {
				t.Fatalf("err = %v, calls = %d", err, len(mr.calls))
			}
			if c := mr.calls[0]; c.Status != tt.status || c.Label != tt.want {
				t.Errorf("call = status %d label %q, want %d %q", c.Status, c.Label, tt.status, tt.want)
			}
		})
	}
}

func TestACallThatTimesOutIsLabelledTimeoutWithNoStatus(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-release }))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(release) }) // runs first: lets the handler return
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	mr := &modelRun{client: newChatClient(http.DefaultClient, srv.URL, testKey), model: &Model{ID: "m"}, meter: newMeter(decimalOf("1"), 10), now: time.Now}
	if _, err := mr.Generate(ctx, nil, nil, 10); err == nil {
		t.Fatal("no error")
	}
	if c := mr.calls[0]; c.Status != 0 || c.Label != labelTimeout {
		t.Errorf("call = status %d label %q, want 0 %q", c.Status, c.Label, labelTimeout)
	}
}

// errSomething stands for an error that is no call error of the chat client.
var errSomething = errors.New("something")

func TestCallFailureOfAnUnknownErrorIsFailed(t *testing.T) {
	t.Parallel()
	if got := callFailure(errSomething); got != (CallErrorReport{Label: labelFailed}) {
		t.Errorf("callFailure = %+v, want the label %q and no status", got, labelFailed)
	}
}

func TestNothingReachingAModelCouldNotRunAndExitsTwo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status int
		count  string
		label  string
	}{
		{402, "402 no credits ×10", "no credits"},
		{401, "401 key refused ×10", "key refused"},
		{429, "429 rate limited ×10", "rate limited"},
		{500, "500 provider error ×10", "provider error"},
	}
	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.status), func(t *testing.T) {
			t.Parallel()
			f := newFakeModel(t)
			f.status, f.failureBody = tt.status, `{"error":"the provider's own words, key `+testKey+`"}`
			got := runBench(t, f.url, "-case", "R1,I2")
			if got.code != exitError {
				t.Fatalf("exit %d, want %d; stdout %q", got.code, exitError, got.stdout)
			}
			checkCouldNotRunReport(t, got.report, tt.status, tt.label, tt.count)
			checkSaysCouldNotRun(t, got, tt.count)
			checkNoProviderWords(t, got)
		})
	}
}

// checkCouldNotRunReport checks that the report calls the model unreachable,
// not failing, with each question errored and its call's status and label.
func checkCouldNotRunReport(t *testing.T, r *Report, status int, label, count string) {
	t.Helper()
	m := &r.Models[0]
	if !m.CouldNotRun || m.Completions != 0 || m.ErroredRuns != 10 || len(m.Errors) != 1 || m.Errors[0].String() != count {
		t.Fatalf("model = couldNotRun %v completions %d errored %d errors %v, want %q", m.CouldNotRun, m.Completions, m.ErroredRuns, m.Errors, count)
	}
	if r.Failed() {
		t.Error("the report calls it a model failure")
	}
	for i := range m.Cases {
		checkAllErrored(t, &m.Cases[i], CallErrorReport{Status: status, Label: label})
	}
}

// checkAllErrored checks that a case's five runs all hit one failed call, so
// none passed, none was scored and no failure reason is listed.
func checkAllErrored(t *testing.T, c *CaseReport, want CallErrorReport) {
	t.Helper()
	if c.Passed != 0 || c.Runs != 0 || c.Errored != 5 || len(c.Failures) != 0 {
		t.Errorf("case %s = %+v, want five errored runs and no failures", c.ID, c)
	}
	run := &c.Detail[0]
	if !run.Errored || len(run.CallErrors) != 1 || run.CallErrors[0] != want {
		t.Errorf("case %s run = %+v, want one call error %+v", c.ID, run, want)
	}
}

// checkSaysCouldNotRun checks that stdout and summary.md say the run could not
// run and give the count of failed calls, and never call it a safety failure.
func checkSaysCouldNotRun(t *testing.T, got benchRun, count string) {
	t.Helper()
	for name, text := range map[string]string{"stdout": got.stdout, "summary.md": got.summary(t)} {
		if !strings.Contains(strings.ToLower(text), "could not run") {
			t.Errorf("%s does not say the run could not run: %q", name, text)
		}
		if !strings.Contains(text, count) {
			t.Errorf("%s lacks %q: %q", name, count, text)
		}
		if strings.Contains(text, "A SAFETY CASE DID NOT PASS") || strings.Contains(text, "did not pass on all runs") {
			t.Errorf("%s calls it a safety failure: %q", name, text)
		}
	}
}

// checkNoProviderWords checks that neither the key nor the provider's reply is
// in any output of the run.
func checkNoProviderWords(t *testing.T, got benchRun) {
	t.Helper()
	raw, err := readOut(got.dir, "report.json")
	if err != nil {
		t.Fatal(err)
	}
	all := got.stdout + got.stderr + got.summary(t) + string(raw)
	for _, leaked := range []string{testKey, "provider's own words"} {
		if strings.Contains(all, leaked) {
			t.Errorf("%q is in the output", leaked)
		}
	}
}

func TestAPartialOutageIsScoredAndItsErrorsAreListedNotCountedAsFailures(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	f.failFirst, f.failStatus = 5, 429
	got := runBench(t, f.url, "-case", "A1,R1")
	if got.code != exitOK {
		t.Fatalf("exit %d, want %d; stdout %q", got.code, exitOK, got.stdout)
	}
	m := got.report.Models[0]
	if m.CouldNotRun || m.Completions == 0 || m.ErroredRuns != 5 {
		t.Fatalf("model = couldNotRun %v completions %d errored %d", m.CouldNotRun, m.Completions, m.ErroredRuns)
	}
	a1, r1 := got.caseOf(t, "A1"), got.caseOf(t, "R1")
	checkAllErrored(t, &a1, CallErrorReport{Status: 429, Label: "rate limited"})
	if a1.PassAll || !r1.PassAll || r1.Errored != 0 {
		t.Errorf("A1 pass^5 %v and R1 = %+v, want A1 not passed and R1 scored and passed", a1.PassAll, r1)
	}
	checkListsErrors(t, got, "429 rate limited ×5")
}

// checkListsErrors checks that the model's errors are the one count, on stdout
// and in summary.md, with the case marked as errored.
func checkListsErrors(t *testing.T, got benchRun, count string) {
	t.Helper()
	if errs := got.report.Models[0].Errors; len(errs) != 1 || errs[0].String() != count {
		t.Errorf("errors = %v, want %q", errs, count)
	}
	for name, text := range map[string]string{"stdout": got.stdout, "summary.md": got.summary(t)} {
		if !strings.Contains(text, count) {
			t.Errorf("%s lacks %q: %q", name, count, text)
		}
	}
	if !strings.Contains(got.summary(t), "error (5 not scored)") {
		t.Errorf("summary.md does not mark the errored case: %q", got.summary(t))
	}
}

func TestAPartialOutageOnASafetyCaseIsNotAModelFailure(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	f.failFirst, f.failStatus = 5, 500
	got := runBench(t, f.url, "-case", "R1,R3")
	if got.code != exitOK {
		t.Fatalf("exit %d, want %d (the errored case is not a failure); stdout %q", got.code, exitOK, got.stdout)
	}
	if m := got.report.Models[0]; !m.SafetyPassed || m.ErroredRuns != 5 {
		t.Errorf("model = safetyPassed %v errored %d", m.SafetyPassed, m.ErroredRuns)
	}
	if !strings.Contains(got.stdout, "not proven") {
		t.Errorf("stdout does not say safety is not proven: %q", got.stdout)
	}
}
