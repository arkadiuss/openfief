package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// dropFirst returns a server that closes the connection of the first n requests
// without any HTTP answer, a transport failure, and hands the rest to f. seen
// counts the requests it received.
func dropFirst(t *testing.T, f *fakeModel, n int64) (string, *atomic.Int64) {
	t.Helper()
	seen := &atomic.Int64{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if seen.Add(1) > n {
			f.ServeHTTP(w, r)
			return
		}
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("the server cannot hijack a connection")
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Errorf("hijack: %v", err)
			return
		}
		if err := conn.Close(); err != nil {
			t.Errorf("close: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/v1", seen
}

func TestATransportErrorThenAnAnswerIsRetriedAndScored(t *testing.T) {
	t.Parallel()
	url, seen := dropFirst(t, newFakeModel(t), 1)
	got := runBench(t, url, "-case", "R1")
	r1 := got.caseOf(t, "R1")
	if r1.Errored != 0 || r1.Runs != 5 || !r1.PassAll {
		t.Fatalf("R1 = errored %d runs %d passAll %v, want the retried question scored and passed", r1.Errored, r1.Runs, r1.PassAll)
	}
	first := r1.Detail[0]
	if len(first.Retries) != 1 || !first.Retries[0].Recovered || first.Retries[0].Label != labelNoAnswer || first.Attempts != 2 || first.Errored {
		t.Errorf("first run = retries %+v attempts %d errored %v, want one recovered retry of a call with no answer", first.Retries, first.Attempts, first.Errored)
	}
	checkRecoveredRetry(t, got.report, int(seen.Load()))
}

// checkRecoveredRetry checks the report of a run whose one failed call was
// retried and answered: the reliability counts, no error, and the budget's calls.
func checkRecoveredRetry(t *testing.T, r *Report, calls int) {
	t.Helper()
	m := r.Models[0]
	if m.Reliability.TransportFailures != 1 || m.Reliability.Retried != 1 || m.Reliability.Recovered != 1 || m.Reliability.Calls != calls {
		t.Errorf("reliability = %+v, want 1 failure of the %d calls, retried and recovered", m.Reliability, calls)
	}
	if m.ErroredRuns != 0 || len(m.Errors) != 0 {
		t.Errorf("errored runs %d, errors %v, want none: the retry answered", m.ErroredRuns, m.Errors)
	}
	if r.Budget.Calls != calls {
		t.Errorf("budget calls = %d, want %d: the retry is a call", r.Budget.Calls, calls)
	}
}

func TestTwoTransportErrorsLeaveTheQuestionNotScored(t *testing.T) {
	t.Parallel()
	url, seen := dropFirst(t, newFakeModel(t), 2)
	got := runBench(t, url, "-case", "R1")
	r1 := got.caseOf(t, "R1")
	if r1.Errored != 1 || r1.Runs != 4 || seen.Load() != 6 {
		t.Fatalf("R1 = errored %d runs %d with %d requests, want the failed question not scored and 6 requests: two for it, one for each other", r1.Errored, r1.Runs, seen.Load())
	}
	if first := &r1.Detail[0]; !failedTwice(first) {
		t.Errorf("first run = %+v, want two failed attempts, a retry that did not recover and the one error that ended the attempt", *first)
	}
}

// failedTwice is whether a run made two attempts, both failing at the transport
// level, retried the first without recovery and listed one error.
func failedTwice(r *RunReport) bool {
	retried := len(r.Retries) == 1 && !r.Retries[0].Recovered
	return r.Errored && r.Attempts == 2 && r.TransportFailures == 2 && retried && len(r.CallErrors) == 1 && r.CallErrors[0].Label == labelNoAnswer
}

func TestA4xxIsNeverRetried(t *testing.T) {
	t.Parallel()
	for _, status := range []int{400, 401, 402, 404} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			t.Parallel()
			f := newFakeModel(t)
			f.failFirst, f.failStatus = 1, status
			got := runBench(t, f.url, "-case", "R1")
			first := got.caseOf(t, "R1").Detail[0]
			if first.Attempts != 1 || len(first.Retries) != 0 || !first.Errored {
				t.Errorf("first run = attempts %d retries %v errored %v, want one attempt, no retry, errored", first.Attempts, first.Retries, first.Errored)
			}
			if f.seen != 5 {
				t.Errorf("requests = %d, want 5: one a question", f.seen)
			}
			if r := got.report.Models[0].Reliability; r.TransportFailures != 0 || r.Failed {
				t.Errorf("reliability = %+v, want a 4xx not counted against the host", r)
			}
		})
	}
}

// A 429 is retried once inside the chat client, after Retry-After; a bench call
// that got through on the retry is one call, and one that is still a 429 is a
// transport failure, the one 4xx that counts against the host (D-207), with no
// retry of the bench's own.
func TestA429ThatTheClientsRetryMendsIsOneAnsweredCall(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	f.failFirst, f.failStatus = 1, 429
	got := runBench(t, f.url, "-case", "R1")
	first := got.caseOf(t, "R1").Detail[0]
	if first.Errored || first.Attempts != 1 || len(first.Retries) != 0 || first.TransportFailures != 0 {
		t.Errorf("first run = %+v, want one answered call: the client's retry hid the 429", first)
	}
	if r := got.report.Models[0].Reliability; r.TransportFailures != 0 || r.Failed {
		t.Errorf("reliability = %+v, want a mended 429 not counted", r)
	}
}

func TestAStill429IsATransportFailureTheBenchDoesNotRetry(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	f.failFirst, f.failStatus = 2, 429 // the first call and the client's retry
	got := runBench(t, f.url, "-case", "R1")
	first := got.caseOf(t, "R1").Detail[0]
	if !first.Errored || first.Attempts != 1 || len(first.Retries) != 0 || first.TransportFailures != 1 {
		t.Errorf("first run = %+v, want one failed call counted as a transport failure and no bench retry", first)
	}
	if f.seen != 6 { // two requests for the first question, one for each of the other four
		t.Errorf("requests = %d, want 6", f.seen)
	}
	if r := got.report.Models[0].Reliability; r.TransportFailures != 1 || r.Calls != 5 || !r.Failed {
		t.Errorf("reliability = %+v, want 1 of 5 calls a transport failure, over the threshold", r)
	}
}

func TestA5xxIsRetriedOnceLikeATransportError(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	f.failFirst, f.failStatus = 1, 503
	got := runBench(t, f.url, "-case", "R1")
	first := got.caseOf(t, "R1").Detail[0]
	if first.Errored || len(first.Retries) != 1 || first.Retries[0].Status != 503 || !first.Retries[0].Recovered {
		t.Errorf("first run = errored %v retries %+v, want a recovered retry of the 503", first.Errored, first.Retries)
	}
}

func TestAnAnswerIsNeverRetried(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	got := runBench(t, f.url, "-case", "R1")
	m := got.report.Models[0]
	if m.Reliability.Retried != 0 || m.Reliability.TransportFailures != 0 || m.Reliability.Failed || m.Reliability.Calls == 0 {
		t.Errorf("reliability = %+v, want calls and no failure", m.Reliability)
	}
}

func TestTheRetryIsACallOfTheBudget(t *testing.T) {
	t.Parallel()
	url, _ := dropFirst(t, newFakeModel(t), 1)
	// One call is allowed: the retry finds the budget spent, so the question is
	// not retried and stays an error, and the run is stopped.
	got := runBench(t, url, "-case", "R1", "-max-calls", "1")
	first := got.caseOf(t, "R1").Detail[0]
	if len(first.Retries) != 0 || first.Attempts != 1 || !first.Errored {
		t.Errorf("first run = attempts %d retries %v errored %v, want no retry once the budget is spent", first.Attempts, first.Retries, first.Errored)
	}
}

func TestAHostIsUnreliableAboveFivePercentTransportFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		failures, calls int
		failed          bool
	}{
		{0, 0, false},
		{0, 100, false},
		{5, 100, false}, // exactly 5% passes
		{6, 100, true},
		{1, 20, false},
		{2, 39, true}, // 5.13%
		{2, 40, false},
		{1, 1, true},
	}
	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.failures)+"of"+strconv.Itoa(tt.calls), func(t *testing.T) {
			t.Parallel()
			r := Reliability{Calls: tt.calls, TransportFailures: tt.failures}
			r.finish()
			if r.Failed != tt.failed {
				t.Errorf("%d of %d: failed = %v (rate %.4f), want %v", tt.failures, tt.calls, r.Failed, r.Rate, tt.failed)
			}
		})
	}
}

func TestTheRateIsTransportFailedCallsOverCallsMade(t *testing.T) {
	t.Parallel()
	runs := []RunReport{
		{Attempts: 3, TransportFailures: 1, Retries: []RetryReport{{Label: labelNoAnswer, Recovered: true}}},
		{Attempts: 2, TransportFailures: 2, Retries: []RetryReport{{Label: labelTimeout}}},
		{Attempts: 5},
	}
	var r Reliability
	r.add(runs)
	r.finish()
	if r.Calls != 10 || r.TransportFailures != 3 || r.Retried != 2 || r.Recovered != 1 || r.Rate != 0.3 || !r.Failed {
		t.Errorf("reliability = %+v, want 3 of 10 calls, 2 retried, 1 recovered, rate 0.3, failed", r)
	}
}

func TestTheRankingLeavesOutAnUnreliableHostWhateverItsVerdict(t *testing.T) {
	t.Parallel()
	models := []ModelReport{
		{Model: "cheap@flaky", ConfigHash: "h1", CostPerQuestion: "0.00001000", Verdict: Verdict{Pass: true}, Reliability: Reliability{Calls: 100, TransportFailures: 9, Rate: 0.09, Failed: true}},
		{Model: "dear@steady", ConfigHash: "h2", CostPerQuestion: "0.00100000", Verdict: Verdict{Pass: true}, Reliability: Reliability{Calls: 100, TransportFailures: 2, Rate: 0.02}},
	}
	got := ranking(models)
	if len(got) != 1 || got[0].Model != "dear@steady" || got[0].TransportFailureRate != 0.02 {
		t.Fatalf("ranking = %+v, want only dear@steady", got)
	}
	summary := (&Report{Models: models, Ranking: got}).Summary()
	for _, want := range []string{"| cheap@flaky | pass | fail (9.0%, 9/100) |", "| dear@steady | pass | pass (2.0%, 2/100) |", "| 1 | dear@steady | pass |", "more than 5%"} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary lacks %q:\n%s", want, summary)
		}
	}
	if strings.Contains(summary, "| 2 |") {
		t.Errorf("summary ranks the unreliable host:\n%s", summary)
	}
	if !strings.Contains(summary, "**Reliability (D-207): fail (9.0%, 9/100).**") {
		t.Errorf("the model's section lacks its reliability:\n%s", summary)
	}
}

func TestNoReliableModelPassingIsNoRanking(t *testing.T) {
	t.Parallel()
	models := []ModelReport{{Model: "m", CostPerQuestion: "0.1", Verdict: Verdict{Pass: true}, Reliability: Reliability{Calls: 10, TransportFailures: 10, Rate: 1, Failed: true}}}
	if got := ranking(models); len(got) != 0 {
		t.Errorf("ranking = %+v, want none", got)
	}
}

func TestAModelRunRetriesATransportErrorAndRecordsBoth(t *testing.T) {
	t.Parallel()
	url, _ := dropFirst(t, newFakeModel(t), 1)
	mr := &modelRun{client: newChatClient(http.DefaultClient, url, testKey), model: &Model{ID: "m"}, meter: newMeter(decimalOf("1"), 10), now: time.Now}
	if _, err := mr.Generate(t.Context(), askedOnce(), nil, 10, ""); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(mr.calls) != 2 {
		t.Fatalf("calls = %d, want the failed call and its retry", len(mr.calls))
	}
	if c := mr.calls[0]; !c.Transport || !c.Retried || !c.Recovered || c.Retry {
		t.Errorf("first call = %+v, want a transport failure, retried and recovered", c)
	}
	if c := mr.calls[1]; c.Error != "" || !c.Retry {
		t.Errorf("second call = %+v, want an answered retry", c)
	}
	if got := mr.meter.snapshot().calls; got != 2 {
		t.Errorf("budget calls = %d, want 2", got)
	}
}
