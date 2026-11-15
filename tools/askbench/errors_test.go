package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/views/ui"
	"github.com/arkadiuss/openfief/plugins/model-openai/openaichat"
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
			_, err := mr.Generate(t.Context(), askedOnce(), nil, 10, "")
			wantCalls := 1 // a 5xx is the provider's and is retried once, the others never
			if tt.status >= 500 {
				wantCalls = 2
			}
			if err == nil || len(mr.calls) != wantCalls {
				t.Fatalf("err = %v, calls = %d, want %d", err, len(mr.calls), wantCalls)
			}
			if c := mr.calls[wantCalls-1]; c.Status != tt.status || c.Label != tt.want {
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
	if _, err := mr.Generate(ctx, askedOnce(), nil, 10, ""); err == nil {
		t.Fatal("no error")
	}
	if c := mr.calls[0]; c.Status != 0 || c.Label != labelTimeout {
		t.Errorf("call = status %d label %q, want 0 %q", c.Status, c.Label, labelTimeout)
	}
}

// errSomething stands for an error that is no call error of the chat client.
var errSomething = errors.New("something")

func TestCallFailureOfAnUnknownErrorIsALocalError(t *testing.T) {
	t.Parallel()
	if got := callFailure(errSomething); got != (CallErrorReport{Label: labelLocal}) {
		t.Errorf("callFailure = %+v, want the label %q and no status", got, labelLocal)
	}
	if got := errorClass(errSomething); got != classLocal {
		t.Errorf("errorClass = %q, want %q", got, classLocal)
	}
}

func TestEveryFailureHasAClassOfTheFixedSet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		err   error
		label string
		class string
	}{
		{"bad reply", openaichat.ErrBadReply, labelBadReply, classBadReply},
		{"wrapped bad reply", fmt.Errorf("x: %w", openaichat.ErrBadReply), labelBadReply, classBadReply},
		{"deadline", context.DeadlineExceeded, labelTimeout, classTimeout},
		{"unknown", errSomething, labelLocal, classLocal},
		{"request not built", fmt.Errorf("%w: encoding", openaichat.ErrRejected), labelLocal, classLocal},
		{"reply not read", fmt.Errorf("%w: the reply could not be read", openaichat.ErrUnavailable), labelLocal, classLocal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := callFailure(tt.err); got != (CallErrorReport{Label: tt.label}) {
				t.Errorf("callFailure = %+v, want the label %q", got, tt.label)
			}
			if got := errorClass(tt.err); got != tt.class {
				t.Errorf("errorClass = %q, want %q", got, tt.class)
			}
		})
	}
}

func TestNothingReachingAModelCouldNotRunAndExitsTwo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status int
		count  string
		label  string
	}{
		{402, "402 no credits (unavailable) ×10", "no credits"},
		{401, "401 key refused (unavailable) ×10", "key refused"},
		{429, "429 rate limited (unavailable) ×10", "rate limited"},
		{500, "500 provider error (unavailable) ×10", "provider error"},
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
		checkAllErrored(t, &m.Cases[i], CallErrorReport{Status: status, Label: label, Class: classUnavailable, Fault: faultHostFor(status)})
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
	f.failFirst, f.failStatus = 10, 429 // two requests a question: the call and the client's retry
	got := runBench(t, f.url, "-case", "A1,R1")
	if got.code != exitOK {
		t.Fatalf("exit %d, want %d; stdout %q", got.code, exitOK, got.stdout)
	}
	m := got.report.Models[0]
	if m.CouldNotRun || m.Completions == 0 || m.ErroredRuns != 5 {
		t.Fatalf("model = couldNotRun %v completions %d errored %d", m.CouldNotRun, m.Completions, m.ErroredRuns)
	}
	a1, r1 := got.caseOf(t, "A1"), got.caseOf(t, "R1")
	checkAllErrored(t, &a1, CallErrorReport{Status: 429, Label: "rate limited", Class: classUnavailable, Fault: faultHost})
	if a1.PassAll || !r1.PassAll || r1.Errored != 0 {
		t.Errorf("A1 pass^5 %v and R1 = %+v, want A1 not passed and R1 scored and passed", a1.PassAll, r1)
	}
	checkListsErrors(t, got, "429 rate limited (unavailable) ×5")
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
	f.failFirst, f.failStatus = 10, 500 // each question fails twice: a call and its retry
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

// badFinalReplyModel plays A3 of run 16-1: it calls the value tool in round 1,
// says nothing after its result in round 2, and answers the final round with a
// 2xx reply that is no chat completion (an error object with no choices).
func badFinalReplyModel(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)
	var body map[string]any
	if err != nil || json.Unmarshal(raw, &body) != nil {
		http.Error(w, "unreadable request", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	switch {
	case len(conversationOf(body).results) == 0:
		say(w, toolCallsReply([2]string{"value", "{}"}))
	case !strings.Contains(string(raw), `"response_format"`):
		say(w, textReply(""))
	default:
		say(w, `{"error":{"message":"provider words about the final round"}}`)
	}
}

// checkBadReplyRun checks a run of A1 against the host that answers its final
// round with a bad reply. The final round's request is well-formed
// (validateRequest passed it), so the bad reply is the host's: retried once,
// both failures in the reliability, the run not void.
func checkBadReplyRun(t *testing.T, run *RunReport) {
	t.Helper()
	want := CallErrorReport{Label: labelBadReply, Class: classBadReply, Fault: faultHost}
	if len(run.Rounds) != 4 || run.Attempts != 4 || len(run.CallErrors) != 1 || run.CallErrors[0] != want || run.TransportFailures != 2 || run.Void {
		t.Fatalf("run %d = rounds %d, attempts %d, call errors %+v, transport failures %d, void %v; want four rounds, one %+v, two transport failures, not void",
			run.Run, len(run.Rounds), run.Attempts, run.CallErrors, run.TransportFailures, run.Void, want)
	}
	for i := 2; i < len(run.Rounds); i++ {
		if last := &run.Rounds[i]; last.Answered || strings.Join(last.Roles, ",") != "system,user,assistant,tool,user" {
			t.Errorf("run %d final round = %s", run.Run, last.summary())
		}
	}
}

// checkBadReplyOutputs checks the model's counts and the texts of a run whose
// calls all ended in bad replies, and that no provider text is in any output.
func checkBadReplyOutputs(t *testing.T, got benchRun) {
	t.Helper()
	m := got.report.Models[0]
	if len(m.Errors) != 1 || m.Errors[0].String() != "bad reply (bad reply) ×5" || m.Reliability.TransportFailures != 10 || m.VoidRuns != 0 {
		t.Errorf("errors = %v, reliability %+v, void %d", m.Errors, m.Reliability, m.VoidRuns)
	}
	for name, text := range map[string]string{"stdout": got.stdout, "summary.md": got.summary(t)} {
		if !strings.Contains(text, "bad reply (bad reply) ×5") || !strings.Contains(text, "ad replies") {
			t.Errorf("%s lacks the bad replies: %q", name, text)
		}
	}
	checkNoProviderWords(t, got)
	if raw, err := readOut(got.dir, "report.json"); err != nil || strings.Contains(string(raw), "provider words") {
		t.Errorf("report.json (err %v) holds the provider's words", err)
	}
}

func TestABadReplyInTheFinalRoundIsReportedWithItsClassAndCountedApart(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(badFinalReplyModel))
	t.Cleanup(srv.Close)
	got := runBench(t, srv.URL+"/v1", "-case", "A1", "-parallel", "1")
	a1 := got.caseOf(t, "A1")
	if a1.Errored != 5 || a1.Passed != 0 {
		t.Fatalf("A1 = %+v, want five errored runs", a1)
	}
	for i := range a1.Detail {
		checkBadReplyRun(t, &a1.Detail[i])
	}
	checkBadReplyOutputs(t, got)
}

// askedOnce is a conversation of one question.
func askedOnce() []ui.AskMessage {
	return []ui.AskMessage{{Role: "user", Content: "How much is my portfolio worth?"}}
}

// faultHostFor is whose fault a failed call of status is: the host's for a 429
// and a 5xx, nobody's for the other refusals (a key, credits).
func faultHostFor(status int) string {
	if status == http.StatusTooManyRequests || status >= http.StatusInternalServerError {
		return faultHost
	}
	return ""
}

func TestValidateRequestAcceptsTheLoopsConversations(t *testing.T) {
	t.Parallel()
	call := ui.AskCall{ID: "c1", Name: "value", Args: "{}"}
	tests := []struct {
		name string
		msgs []ui.AskMessage
	}{
		{"the first round", []ui.AskMessage{{Role: "system", Content: "s"}, {Role: "user", Content: "q"}}},
		// A3 of run 16-1: round 3, after an empty reply to a tool's result.
		{"the final round after an empty reply", []ui.AskMessage{
			{Role: "system", Content: "s"},
			{Role: "user", Content: "q"},
			{Role: "assistant", Calls: []ui.AskCall{call}},
			{Role: "tool", Content: "card", CallID: "c1"},
			{Role: "user", Content: "Answer the user's question in this schema."},
		}},
		{"the final round after plain words", []ui.AskMessage{
			{Role: "system", Content: "s"},
			{Role: "user", Content: "q"},
			{Role: "assistant", Content: "words"},
			{Role: "user", Content: "Put it in the schema."},
		}},
		{"a retry after the final round", []ui.AskMessage{
			{Role: "system", Content: "s"},
			{Role: "user", Content: "q"},
			{Role: "assistant", Calls: []ui.AskCall{call}},
			{Role: "tool", Content: "card", CallID: "c1"},
			{Role: "user", Content: "Answer."},
			{Role: "assistant", Content: "{bad"},
			{Role: "user", Content: "Fix it."},
		}},
		{"two calls answered in turn", []ui.AskMessage{
			{Role: "user", Content: "q"},
			{Role: "assistant", Calls: []ui.AskCall{call, {ID: "c2", Name: "value", Args: "{}"}}},
			{Role: "tool", Content: "a", CallID: "c2"},
			{Role: "tool", Content: "b", CallID: "c1"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateRequest(requestOf("m", tt.msgs, nil, 10)); err != nil {
				t.Errorf("validateRequest = %v, want a valid request", err)
			}
		})
	}
}

func TestValidateRequestRefusesWhatTheChatAPIRefuses(t *testing.T) {
	t.Parallel()
	call := ui.AskCall{ID: "c1", Name: "value", Args: "{}"}
	tests := []struct {
		name string
		msgs []ui.AskMessage
	}{
		{"no messages", nil},
		{"an assistant message with neither text nor calls", []ui.AskMessage{{Role: "user", Content: "q"}, {Role: "assistant"}, {Role: "user", Content: "x"}}},
		{"a tool call never answered", []ui.AskMessage{{Role: "user", Content: "q"}, {Role: "assistant", Calls: []ui.AskCall{call}}}},
		{"a tool call answered after another message", []ui.AskMessage{{Role: "user", Content: "q"}, {Role: "assistant", Calls: []ui.AskCall{call}}, {Role: "user", Content: "x"}, {Role: "tool", Content: "r", CallID: "c1"}}},
		{"a tool message that answers nothing", []ui.AskMessage{{Role: "user", Content: "q"}, {Role: "tool", Content: "r", CallID: "c9"}}},
		{"a tool call with no id", []ui.AskMessage{{Role: "user", Content: "q"}, {Role: "assistant", Calls: []ui.AskCall{{Name: "value"}}}, {Role: "tool", Content: "r"}}},
		{"a system message after the first", []ui.AskMessage{{Role: "user", Content: "q"}, {Role: "system", Content: "s"}, {Role: "user", Content: "x"}}},
		{"a conversation that ends with the assistant", []ui.AskMessage{{Role: "user", Content: "q"}, {Role: "assistant", Content: "a"}}},
		{"a user message with no text", []ui.AskMessage{{Role: "user", Content: "  "}}},
		{"an unknown role", []ui.AskMessage{{Role: "robot", Content: "q"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateRequest(requestOf("m", tt.msgs, nil, 10)); err == nil {
				t.Error("validateRequest accepted a request the chat API refuses")
			}
		})
	}
}

func TestAnInvalidRequestIsNeverSentAndIsTheBenchsFault(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	mr := &modelRun{client: newChatClient(http.DefaultClient, f.url, testKey), model: &Model{ID: "m"}, meter: newMeter(decimalOf("1"), 10), now: time.Now}
	bad := []ui.AskMessage{{Role: "user", Content: "q"}, {Role: "assistant"}}
	if _, err := mr.Generate(t.Context(), bad, nil, 10, ""); err == nil {
		t.Fatal("Generate accepted an invalid request")
	}
	if len(f.requests) != 0 || len(mr.calls) != 1 {
		t.Fatalf("%d requests sent and %d calls recorded, want none sent and one recorded", len(f.requests), len(mr.calls))
	}
	if c := mr.calls[0]; c.Fault != faultBench || c.Label != labelInvalid || c.Error != classInvalid || c.Transport || c.Retried {
		t.Errorf("call = %+v, want the bench's fault, invalid request, no transport failure, no retry", c)
	}
}

func TestEveryFailedCallIsInOneBucket(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"a reply that is no chat completion", openaichat.ErrBadReply, faultHost},
		{"a degenerate generation", errDegenerate, faultHost},
		{"a deadline", context.DeadlineExceeded, faultHost},
		{"an invalid request", errInvalidRequest, faultBench},
		{"an unknown error", errSomething, faultBench},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := faultOf(tt.err); got != tt.want {
				t.Errorf("faultOf = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestADegenerateGenerationIsTheHostsFaultAndRetriedOnce(t *testing.T) {
	t.Parallel()
	replies := map[string]string{
		"length with no text": completionOf("", "length"),
		"harmony tokens":      completionOf("<|channel|>final<|message|>hi", "stop"),
	}
	for name, reply := range replies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				say(w, reply)
			}))
			t.Cleanup(srv.Close)
			mr := &modelRun{client: newChatClient(http.DefaultClient, srv.URL+"/v1", testKey), model: &Model{ID: "m"}, meter: newMeter(decimalOf("1"), 10), now: time.Now}
			if _, err := mr.Generate(t.Context(), askedOnce(), nil, 10, ""); err == nil {
				t.Fatal("Generate took a degenerate generation for an answer")
			}
			if len(mr.calls) != 2 || !mr.calls[0].Retried || !mr.calls[1].Transport || mr.calls[1].Fault != faultHost || mr.calls[1].Error != classDegenerate {
				t.Errorf("calls = %+v, want two degenerate calls, the host's fault, the first retried", mr.calls)
			}
			if mr.calls[1].InputTokens != 10 {
				t.Errorf("the degenerate call's tokens were not counted: %+v", mr.calls[1])
			}
		})
	}
}

// completionOf is a chat completion of 10 and 5 tokens with content and finish.
func completionOf(content, finish string) string {
	return `{"choices":[{"message":{"content":` + jsonText(content) + `},"finish_reason":"` + finish + `"}],"usage":{"prompt_tokens":10,"completion_tokens":5}}`
}

func TestALengthStopWithAToolCallIsNotDegenerate(t *testing.T) {
	t.Parallel()
	if degenerate(&openaichat.Reply{FinishReason: "length", ToolCalls: []openaichat.ToolCall{{ID: "c", Name: "value"}}}) {
		t.Error("a reply with a tool call is degenerate")
	}
	if degenerate(&openaichat.Reply{FinishReason: "stop"}) {
		t.Error("an empty reply that stopped is degenerate: the loop answers it with the final round")
	}
}

func TestAVoidRunHasNoVerdictInTheHistoryAndSaysSo(t *testing.T) {
	t.Parallel()
	if _, counted := entryOf(&ModelReport{VoidRuns: 1}, "2026-11-15"); counted {
		t.Error("a model with a void run has a place in the history")
	}
	c := &CaseReport{Runs: 4, Passed: 4, Errored: 1, Void: 1}
	if got := passCellOf(c, true, 4); got != "void (bench fault, 1 not scored)" {
		t.Errorf("cell = %q", got)
	}
	var b strings.Builder
	writeModel(&b, &ModelReport{Model: "m", ConfigHash: "h", Errors: []ErrorCount{{Label: labelInvalid, Class: classInvalid, Calls: 1}}, ErroredRuns: 1, VoidRuns: 1, SafetyPassed: true})
	if !strings.Contains(b.String(), "void (bench fault)") || !strings.Contains(b.String(), "run again after the fix") {
		t.Errorf("summary does not say the runs are void: %q", b.String())
	}
}
