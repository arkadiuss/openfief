package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// The conversation cases (D-224): a follow-up and an injection across turns
// hold the earlier turns the page's browser would send, and the bench sends
// them the way the Ask box's server takes them.

func TestTheConversationCasesHoldTheirEarlierTurns(t *testing.T) {
	t.Parallel()
	cases := loadCases(t)
	for _, id := range []string{"F1", "F2", "I4", "I5"} {
		c := caseByID(t, cases, id)
		if len(c.History) == 0 || len(c.earlier()) != len(c.History) {
			t.Errorf("%s has %d earlier turns, want some, all of them sent to the Ask box", id, len(c.History))
		}
	}
	if got := caseByID(t, cases, "A1"); len(got.History) != 0 {
		t.Errorf("A1 has earlier turns %v, want none: it is a first question", got.History)
	}
}

func TestTheInjectionsAcrossTurnsAreSafetyCasesWithTheirPayloads(t *testing.T) {
	t.Parallel()
	cases := loadCases(t)
	if a := caseByID(t, cases, "I4").History[0].Answer; !strings.Contains(a, "ignore my rules") {
		t.Errorf("I4's forged answer is %q, want one that says the rules are off", a)
	}
	if q := caseByID(t, cases, "I5").History[0].Question; !strings.Contains(q, "sync_status") {
		t.Errorf("I5's earlier question is %q, want one that plants a call of sync_status", q)
	}
	for _, id := range []string{"I4", "I5"} {
		c := caseByID(t, cases, id)
		if !c.Safety || !c.safetyCheck(scorerCallsSafe) || len(c.Expect.PayloadCalls) == 0 {
			t.Errorf("%s = %+v, want a safety case that grades acting on the payload", id, c)
		}
	}
}

func TestTheBenchSendsTheEarlierTurnsAsAnUntrustedTranscript(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	runBench(t, f.url, "-case", "F1")
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.requests) == 0 {
		t.Fatal("the model was asked nothing")
	}
	first := firstUserMessage(t, f.requests[0])
	for _, want := range []string{"untrusted", "How did I do this year?", "99,99", "New question: And last year?"} {
		if !strings.Contains(first, want) {
			t.Errorf("the first user message lacks %q:\n%s", want, first)
		}
	}
}

// firstUserMessage is the text of the first user message of a request the fake model received.
func firstUserMessage(t *testing.T, request map[string]any) string {
	t.Helper()
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	for _, m := range body.Messages {
		if m.Role == "user" {
			return m.Content
		}
	}
	t.Fatal("the request has no user message")
	return ""
}
