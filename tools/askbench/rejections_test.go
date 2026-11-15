package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/views/ui"
	"github.com/arkadiuss/openfief/plugins/model-openai/openaichat"
)

// rejectingModels returns the URL of a fake model that answers every request
// for one of the models with status, and the others as the fake model does.
func rejectingModels(t *testing.T, status int, models ...string) string {
	t.Helper()
	f := newFakeModel(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		var body map[string]any
		if err != nil || json.Unmarshal(raw, &body) != nil {
			http.Error(w, "unreadable request", http.StatusBadRequest)
			return
		}
		if slices.Contains(models, fmt.Sprint(body["model"])) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			say(w, `{"error":"the provider's words about the request"}`)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(raw))
		f.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/v1"
}

// twoEntries are the flags of a run of two models that ask the same questions.
func twoEntries(extra ...string) []string {
	return append([]string{"-model", "test/good,test/bad", "-price", "test/good=1:2,test/bad=1:2", "-case", "R1"}, extra...)
}

// entryOfReport returns the model of the report with the label.
func entryOfReport(t *testing.T, r *Report, label string) *ModelReport {
	t.Helper()
	for i := range r.Models {
		if r.Models[i].Model == label {
			return &r.Models[i]
		}
	}
	t.Fatalf("no model %s in the report", label)
	return nil
}

// callWant is what each question's one failed call of a model should read.
type callWant struct {
	fault, rule, reason string
	status, failures    int
}

// checkEachCall checks that every run of the model's first case ended with one
// failed call of the kind want says, with the request's hash, and is void for
// reason (none when empty).
func checkEachCall(t *testing.T, m *ModelReport, want *callWant) {
	t.Helper()
	runs := m.Cases[0].Detail
	for i := range runs {
		run := &runs[i]
		if len(run.CallErrors) != 1 {
			t.Fatalf("%s run %d has %d call errors, want one", m.Model, run.Run, len(run.CallErrors))
		}
		ce := &run.CallErrors[0]
		if ce.Fault != want.fault || ce.Rule != want.rule || run.VoidReason != want.reason || run.Void != (want.reason != "") || run.TransportFailures != want.failures {
			t.Errorf("%s run %d = %+v (void %v %q, failed calls %d), want the %s's fault by the rule %q, void %q, %d failed calls",
				m.Model, run.Run, *ce, run.Void, run.VoidReason, run.TransportFailures, want.fault, want.rule, want.reason, want.failures)
		}
		if want.status != 0 && ce.Status != want.status {
			t.Errorf("%s run %d status = %d, want %d", m.Model, run.Run, ce.Status, want.status)
		}
	}
}

// checkTexts checks that summary.md and stdout both hold every text, case aside.
func checkTexts(t *testing.T, got benchRun, texts ...string) {
	t.Helper()
	for name, doc := range map[string]string{"summary.md": got.summary(t), "stdout": got.stdout} {
		for _, text := range texts {
			if !strings.Contains(strings.ToLower(doc), strings.ToLower(text)) {
				t.Errorf("%s lacks %q:\n%s", name, text, doc)
			}
		}
	}
}

// checkRefusingHost checks that the host that refused failed all five questions
// and was void for none, and the one that accepted failed none.
func checkRefusingHost(t *testing.T, bad, good *ModelReport) {
	t.Helper()
	if r := bad.Reliability; r.Questions != 5 || r.FailedQuestions != 5 || r.TransportFailures != 5 || !r.Failed || bad.VoidRuns != 0 {
		t.Errorf("the host that refused: reliability %+v, void %d, want five failed questions and no void run", r, bad.VoidRuns)
	}
	if r := good.Reliability; r.FailedQuestions != 0 || r.Failed || good.ErroredRuns != 0 {
		t.Errorf("the host that accepted: reliability %+v, errored %d", r, good.ErroredRuns)
	}
}

func TestARejectionAnotherEntryAcceptedIsTheHostsFault(t *testing.T) {
	t.Parallel()
	for _, status := range []int{http.StatusBadRequest, http.StatusRequestEntityTooLarge, http.StatusUnprocessableEntity} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			t.Parallel()
			got := runBench(t, rejectingModels(t, status, "test/bad"), twoEntries()...)
			bad, good := entryOfReport(t, got.report, "test/bad"), entryOfReport(t, got.report, "test/good")
			checkEachCall(t, bad, &callWant{fault: faultHost, rule: ruleAcceptedElsewhere, status: status, failures: 1})
			checkRefusingHost(t, bad, good)
			if got.code == exitCredentials {
				t.Errorf("exit %d: a host's refusal is no credentials fault", got.code)
			}
			checkTexts(t, got, "accepted elsewhere")
		})
	}
}

func TestARejectionNoEntryAcceptedIsTheBenchsFaultAndVoidsTheRun(t *testing.T) {
	t.Parallel()
	for _, status := range []int{http.StatusBadRequest, http.StatusRequestEntityTooLarge, http.StatusUnprocessableEntity} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			t.Parallel()
			got := runBench(t, rejectingModels(t, status, "test/good", "test/bad"), twoEntries()...)
			for i := range got.report.Models {
				m := &got.report.Models[i]
				if m.VoidRuns != 5 || m.VoidCredentials != 0 || m.Reliability.Questions != 0 || m.Reliability.Failed || m.Reliability.TransportFailures != 0 {
					t.Errorf("%s: void %d, reliability %+v, want five bench-void runs counted against no host", m.Model, m.VoidRuns, m.Reliability)
				}
				checkEachCall(t, m, &callWant{fault: faultBench, rule: ruleAcceptedNowhere, reason: voidBench, status: status})
			}
			if hashes := requestHashesOf(got.report); len(hashes) != 1 || hashes[0] == "" {
				t.Errorf("the two entries' identical requests have hashes %v, want one", hashes)
			}
			checkTexts(t, got, "accepted nowhere", "requestcheck.go", "void (bench fault)")
		})
	}
}

// requestHashesOf lists the distinct hashes of the failed calls of a report.
func requestHashesOf(r *Report) []string {
	var out []string
	for mi := range r.Models {
		for ci := range r.Models[mi].Cases {
			runs := r.Models[mi].Cases[ci].Detail
			for ri := range runs {
				for _, ce := range runs[ri].CallErrors {
					if !slices.Contains(out, ce.RequestHash) {
						out = append(out, ce.RequestHash)
					}
				}
			}
		}
	}
	return out
}

func TestARejectionOfTheOnlyEntryIsAcceptedNowhere(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	f.failFirst, f.failStatus = 1, http.StatusBadRequest
	got := runBench(t, f.url, "-case", "R1")
	m := got.report.Models[0]
	if m.VoidRuns != 1 || m.Reliability.Questions != 4 || m.Reliability.FailedQuestions != 0 || got.code == exitCredentials {
		t.Errorf("void %d, reliability %+v, exit %d, want one void run left out of the four questions", m.VoidRuns, m.Reliability, got.code)
	}
	if ce := m.Cases[0].Detail[0].CallErrors[0]; ce.Rule != ruleAcceptedNowhere || ce.Fault != faultBench {
		t.Errorf("call error = %+v", ce)
	}
}

func TestARejectionIsDecidedByTheRequestNotTheModelsName(t *testing.T) {
	t.Parallel()
	// Both entries get the same question, so the same request hash despite the
	// different model ids, and the routing members never enter the hash.
	client := newChatClient(http.DefaultClient, "https://openrouter.ai/api/v1", "k")
	req := func(model string) *openaichat.Request {
		return &openaichat.Request{Model: model, Messages: []openaichat.Message{{Role: "user", Content: "How much is my portfolio worth?"}}}
	}
	routed := routedClient(http.DefaultClient, "https://openrouter.ai/api/v1", "k", []string{"a", "b"})
	if requestHash(client, req("m/one")) != requestHash(routed, req("m/two")) {
		t.Error("equivalent requests to other hosts and models have different hashes")
	}
	other := req("m/one")
	other.Messages[0].Content = "another question"
	if requestHash(client, req("m/one")) == requestHash(client, other) {
		t.Error("different requests have the same hash")
	}
}

func TestA404CountsAgainstTheHostAndSaysThePinMayBeStale(t *testing.T) {
	t.Parallel()
	got := runBench(t, rejectingModels(t, http.StatusNotFound, "test/bad"), twoEntries()...)
	bad := entryOfReport(t, got.report, "test/bad")
	checkEachCall(t, bad, &callWant{fault: faultHost, rule: ruleNotFound, status: http.StatusNotFound, failures: 1})
	if r := bad.Reliability; r.FailedQuestions != 5 || r.TransportFailures != 5 || r.Calls != 5 || !r.Failed {
		t.Errorf("reliability = %+v, want a host that fails every question with no retry", r)
	}
	checkTexts(t, got, "404: the pin may be stale")
	if got.code == exitCredentials {
		t.Errorf("exit %d", got.code)
	}
}

// checkCredentialsVoid checks a model whose first run a refused key or no
// credits ended: void for the credentials, counted against no host.
func checkCredentialsVoid(t *testing.T, m *ModelReport) {
	t.Helper()
	run := &m.Cases[0].Detail[0]
	if !run.Void || run.VoidReason != voidCredentials || run.CallErrors[0].Fault != faultCredentials || run.CallErrors[0].Rule != ruleCredentials {
		t.Errorf("first run = %+v, want void for the credentials", run)
	}
	want := Reliability{Questions: 4, Calls: 5}
	want.finish()
	if m.VoidRuns != 1 || m.VoidCredentials != 1 || m.Reliability != want {
		t.Errorf("void %d/%d, reliability %+v, want the run counted against no host: %+v", m.VoidRuns, m.VoidCredentials, m.Reliability, want)
	}
	if _, counted := entryOf(m, "2026-11-15"); counted {
		t.Error("a run void for its credentials has a place in the history")
	}
}

func TestARefusedKeyOrNoCreditsVoidsTheRunAndExitsWithItsOwnCode(t *testing.T) {
	t.Parallel()
	for _, status := range []int{http.StatusUnauthorized, http.StatusPaymentRequired} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			t.Parallel()
			f := newFakeModel(t)
			f.failFirst, f.failStatus = 1, status
			got := runBench(t, f.url, "-case", "R1")
			checkCredentialsVoid(t, &got.report.Models[0])
			if got.code != exitCredentials || !got.report.CredentialsVoid() {
				t.Errorf("exit %d, want %d", got.code, exitCredentials)
			}
			checkTexts(t, got, "void (credentials/credit)")
		})
	}
}

func TestTheCredentialsExitCodeIsDistinctAndComesFirst(t *testing.T) {
	t.Parallel()
	others := []int{exitOK, exitUnsafe, exitError, exitBudget, exitRoute}
	if slices.Contains(others, exitCredentials) {
		t.Fatalf("exit %d is not distinct from %v", exitCredentials, others)
	}
	r := &Report{Models: []ModelReport{{VoidRuns: 1, VoidCredentials: 1, CouldNotRun: true}}, Budget: BudgetReport{Aborted: true}}
	if got := exitCode(r); got != exitCredentials {
		t.Errorf("exit %d, want %d before the could-not-run and budget codes", got, exitCredentials)
	}
	r = &Report{Models: []ModelReport{{VoidRuns: 1, SafetyPassed: true}}}
	if got := exitCode(r); got != exitOK {
		t.Errorf("exit %d: a bench-void run alone is no credentials fault", got)
	}
}

func TestADeterministicBadReplyIsOneFailedQuestionNotTwo(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(badFinalReplyModel))
	t.Cleanup(srv.Close)
	got := runBench(t, srv.URL+"/v1", "-case", "A1", "-parallel", "1")
	m := got.report.Models[0]
	// Each question's final round fails twice, the call and its retry, and the
	// question counts once.
	if r := m.Reliability; r.Questions != 5 || r.FailedQuestions != 5 || r.TransportFailures != 10 || r.Calls != 20 || r.Rate != 1 || !r.Failed {
		t.Errorf("reliability = %+v, want 5 of 5 questions failed by 10 failed calls", r)
	}
	if want := "fail (100.0% of questions, 5 of 5; 10 failed calls)"; m.Reliability.text() != want {
		t.Errorf("text = %q, want %q", m.Reliability.text(), want)
	}
	if want := "Reliability (D-207): fail (100.0% of questions, 5 of 5; 10 failed calls)"; !strings.Contains(got.summary(t), want) {
		t.Errorf("summary lacks %q:\n%s", want, got.summary(t))
	}
}

func TestAQuestionTheRetryMendedIsNotFailedAndOneThatBeatsTheHostIsFailedOnce(t *testing.T) {
	t.Parallel()
	hostFailed := []CallErrorReport{{Label: labelNoAnswer, Class: classUnavailable, Fault: faultHost}}
	benchFailed := []CallErrorReport{{Label: labelInvalid, Class: classInvalid, Fault: faultBench}}
	tests := []struct {
		name   string
		run    RunReport
		counts bool
		failed bool
	}{
		{name: "mended by its retry", run: RunReport{Attempts: 2, TransportFailures: 1, Retries: []RetryReport{{Label: labelNoAnswer, Recovered: true}}}, counts: true},
		{name: "failed after its retry", run: RunReport{Attempts: 2, TransportFailures: 2, Errored: true, CallErrors: hostFailed, Retries: []RetryReport{{Label: labelNoAnswer}}}, counts: true, failed: true},
		{name: "void for the bench", run: RunReport{Errored: true, Void: true, VoidReason: voidBench, CallErrors: benchFailed}},
		{name: "answered by the matcher", run: RunReport{Matcher: true}},
		{name: "never asked", run: RunReport{NotRun: true}},
		{name: "stopped by the budget", run: RunReport{BudgetSpent: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.run.countsForReliability(); got != tt.counts {
				t.Errorf("counts = %v, want %v", got, tt.counts)
			}
			if got := tt.run.failedByHost(); got != tt.failed {
				t.Errorf("failed by the host = %v, want %v", got, tt.failed)
			}
		})
	}
}

func TestTheRequestCheckRefusesWhatEveryHostRefuses(t *testing.T) {
	t.Parallel()
	msgs := []ui.AskMessage{{Role: "user", Content: "q"}}
	long := strings.Repeat("x", maxToolName+1)
	tests := []struct {
		name  string
		tools []ui.AskTool
		valid bool
	}{
		{name: "a plain tool", tools: []ui.AskTool{{Name: "what_needs_me", Schema: `{"type":"object"}`}}, valid: true},
		{name: "a tool with no parameters", tools: []ui.AskTool{{Name: "value-2"}}, valid: true},
		{name: "a name with a space", tools: []ui.AskTool{{Name: "what needs me"}}},
		{name: "an empty name", tools: []ui.AskTool{{Name: ""}}},
		{name: "a name too long", tools: []ui.AskTool{{Name: long}}},
		{name: "a name used twice", tools: []ui.AskTool{{Name: "a"}, {Name: "a"}}},
		{name: "parameters that are no object", tools: []ui.AskTool{{Name: "a", Schema: `[1]`}}},
		{name: "parameters that are no JSON", tools: []ui.AskTool{{Name: "a", Schema: `{`}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateRequest(requestOf("m", msgs, tt.tools, 10))
			if (err == nil) != tt.valid {
				t.Errorf("validateRequest = %v, want valid = %v", err, tt.valid)
			}
		})
	}
	req := requestOf("m", msgs, nil, 10)
	req.SetAnswerFormat([]byte(`"text"`), true)
	if validateRequest(req) == nil {
		t.Error("a response schema that is no object was accepted")
	}
}

func TestEveryToolTheBenchOffersPassesTheRequestCheck(t *testing.T) {
	t.Parallel()
	req := requestOf("m", []ui.AskMessage{{Role: "user", Content: "q"}}, ui.AskTools(), 10)
	if err := validateRequest(req); err != nil {
		t.Errorf("the box's own tools fail the request check: %v", err)
	}
}
