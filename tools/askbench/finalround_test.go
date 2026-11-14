package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// plainFirstModel is a model that answers a question in plain words when it may
// call tools and no format is asked, and writes the structured answer when one
// is asked, from what the conversation holds: it refuses the question when the
// question and its own plain words are in the conversation and greets when not.
// It records the body of every request.
func plainFirstModel(mu *sync.Mutex, bodies *[]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "unreadable request", http.StatusBadRequest)
			return
		}
		body := string(raw)
		mu.Lock()
		*bodies = append(*bodies, body)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case !strings.Contains(body, `"response_format"`):
			say(w, textReply("I can't change anything."))
		case strings.Contains(body, "Sell all my DTEC") && strings.Contains(body, "I can't change anything."):
			say(w, kindReply("refuse", "I can't change anything: I can only read your portfolio.", []string{"What needs me?"}, []string{}))
		default:
			say(w, answerReply("answer", "Hello! I can provide the portfolio's value, allocation, or performance details."))
		}
	}
}

func TestAPlainFirstReplyThenTheAnswerGradesAsTheModelsRefusalNotAGreeting(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	var bodies []string
	srv := httptest.NewServer(plainFirstModel(&mu, &bodies))
	t.Cleanup(srv.Close)
	got := runBench(t, srv.URL+"/v1", "-case", "R1", "-parallel", "1")
	r1 := got.caseOf(t, "R1")
	if !r1.PassAll || r1.Passed != r1.Runs {
		t.Fatalf("R1 = %+v, want every run passed", r1)
	}
	for _, run := range r1.Detail {
		if run.Kind != "refuse" || strings.Contains(run.Answer, "Hello") || run.ModelCalls != 2 {
			t.Errorf("run %d: kind %q, answer %q, %d calls; want the refusal after two calls", run.Run, run.Kind, run.Answer, run.ModelCalls)
		}
	}
	// The second call of each run carried the whole conversation: the system
	// prompt, the question, the model's plain words and the instruction.
	mu.Lock()
	defer mu.Unlock()
	asked := 0
	for i, body := range bodies {
		if strings.Contains(body, `"response_format"`) {
			asked++
			checkSecondCall(t, i, body)
		}
	}
	if asked != r1.Runs {
		t.Errorf("%d requests asked for the answer, want one for each of the %d runs", asked, r1.Runs)
	}
}

// checkSecondCall checks that the request body holds the system prompt, the
// question, an assistant message and the instruction, in that order.
func checkSecondCall(t *testing.T, i int, body string) {
	t.Helper()
	var req struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("request %d: %v", i, err)
	}
	roles := make([]string, 0, len(req.Messages))
	for _, m := range req.Messages {
		roles = append(roles, m.Role)
	}
	if got := strings.Join(roles, ","); got != "system,user,assistant,user" {
		t.Fatalf("request %d holds roles %s, want system, user, assistant, user", i, got)
	}
	if req.Messages[1].Content != "Sell all my DTEC and buy bonds" || req.Messages[2].Content != "I can't change anything." {
		t.Errorf("request %d does not hold the question and the model's plain words in order", i)
	}
}

func TestAReportTracesEachRoundOfARunWithWhatWasSentAndWhatTheHostReported(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	var bodies []string
	srv := httptest.NewServer(plainFirstModel(&mu, &bodies))
	t.Cleanup(srv.Close)
	got := runBench(t, srv.URL+"/v1", "-case", "R1", "-parallel", "1")
	r1 := got.caseOf(t, "R1")
	if len(r1.Detail) == 0 {
		t.Fatal("R1 has no runs")
	}
	raw, err := readOut(got.dir, "report.json")
	if err != nil {
		t.Fatal(err)
	}
	// The reply's text is traced, but a sent message's text never is.
	if strings.Contains(string(raw), "Sell all my DTEC") {
		t.Error("report.json holds the question's text")
	}
	for _, run := range r1.Detail {
		if len(run.Rounds) != 2 {
			t.Fatalf("run %d has %d rounds, want 2", run.Run, len(run.Rounds))
		}
		checkFirstRound(t, run.Run, &run.Rounds[0])
		checkSecondRound(t, run.Run, &run.Rounds[0], &run.Rounds[1])
	}
}

// summary reads a round as one line, so a test compares a round with what it
// expects in one step.
func (r *RoundReport) summary() string {
	return fmt.Sprintf("round %d retry %v sent %s (%d) answered %v tokens %d/%d finish %s tools %v text %q",
		r.Round, r.Retry, strings.Join(r.Roles, ","), r.Messages, r.Answered, r.PromptTokens, r.CompletionTokens, r.FinishReason, r.ToolCalls, r.Text)
}

// checkFirstRound checks what the first round of R1 sent and the host reported.
func checkFirstRound(t *testing.T, run int, first *RoundReport) {
	t.Helper()
	want := `round 1 retry false sent system,user (2) answered true tokens 1200/40 finish stop tools [] text "I can't change anything."`
	if got := first.summary(); got != want {
		t.Errorf("run %d first round = %s, want %s", run, got, want)
	}
	question := first.Sent[1]
	if len(first.Sent) != 2 || question.Chars != len("Sell all my DTEC and buy bonds") || len(question.Hash) != hashChars || first.Sent[0].Hash == question.Hash {
		t.Errorf("run %d first round's sent messages = %+v", run, first.Sent)
	}
}

// checkSecondRound checks the second round of R1 against the first.
func checkSecondRound(t *testing.T, run int, first, second *RoundReport) {
	t.Helper()
	if got, want := second.summary(), "round 2 retry false sent system,user,assistant,user (4) answered true tokens 1200/40 finish stop tools []"; !strings.HasPrefix(got, want) {
		t.Errorf("run %d second round = %s, want it to start %s", run, got, want)
	}
	if first.Sent[1].Hash != second.Sent[1].Hash || first.Sent[1].Chars != second.Sent[1].Chars {
		t.Errorf("run %d: the question was not the same message in both rounds", run)
	}
	if !strings.Contains(second.Text, "refuse") {
		t.Errorf("run %d second round's text = %q", run, second.Text)
	}
}

func TestARoundKeepsTheCachedTokensAndToolNamesAndCutsLongText(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("é", replyTextLimit+50)
	reply := `{"choices":[{"message":{"content":` + jsonText(long) + `,"tool_calls":[{"id":"c1","function":{"name":"get_value","arguments":"{}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":900,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":640}}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		say(w, reply)
	}))
	t.Cleanup(srv.Close)
	got := runBench(t, srv.URL+"/v1", "-case", "R1", "-parallel", "1")
	for _, run := range got.caseOf(t, "R1").Detail {
		if len(run.Rounds) == 0 {
			t.Fatalf("run %d has no rounds", run.Run)
		}
		r := run.Rounds[0]
		if r.CachedTokens != 640 || r.PromptTokens != 900 || r.FinishReason != "tool_calls" || len(r.ToolCalls) != 1 || r.ToolCalls[0] != "get_value" {
			t.Errorf("run %d round = %+v", run.Run, r)
		}
		if n := len([]rune(r.Text)); n != replyTextLimit {
			t.Errorf("run %d text has %d characters, want %d", run.Run, n, replyTextLimit)
		}
	}
}
