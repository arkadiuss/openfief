package main

import (
	"encoding/json"
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
