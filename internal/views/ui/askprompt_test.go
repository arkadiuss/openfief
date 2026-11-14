package ui

import (
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestPromptHoldsTheCanaryAndTheDay(t *testing.T) {
	t.Parallel()
	day, err := kernel.ParseDate("2026-10-28")
	if err != nil {
		t.Fatal(err)
	}
	prompt := askSystemPrompt(day)
	if !strings.Contains(prompt, askPromptCanary) {
		t.Error("the system prompt does not hold the canary")
	}
	if !strings.HasSuffix(prompt, "Today is 2026-10-28.") {
		t.Errorf("the system prompt ends %q, want the day", prompt[len(prompt)-30:])
	}
	if AskSystemPrompt(day) != prompt {
		t.Error("the exported prompt is not the page's")
	}
}

func TestPromptSplitsWhatTheBoxSaysOfItselfFromTheRest(t *testing.T) {
	t.Parallel()
	day := kernel.Date{}
	parts := askPromptParts(day)
	var public, rest []string
	for _, p := range parts {
		if p.Public {
			public = append(public, p.Text)
		} else {
			rest = append(rest, p.Text)
		}
	}
	if len(public) != 1 || !strings.Contains(public[0], "investment portfolio") {
		t.Errorf("public sentences = %q, want the one self-description", public)
	}
	got := AskPromptNonPublic(day)
	if strings.Join(got, "|") != strings.Join(rest, "|") {
		t.Errorf("non-public = %q, want %q", got, rest)
	}
	if strings.Contains(strings.Join(got, " "), public[0]) {
		t.Error("a public sentence is among the non-public ones")
	}
	// The prompt is the sentences, in order, and nothing else.
	all := make([]string, 0, len(parts))
	for _, p := range parts {
		all = append(all, p.Text)
	}
	if askSystemPrompt(day) != strings.Join(all, " ") {
		t.Error("the prompt is not its sentences in order")
	}
}

func TestPromptDeclinesExactlyChangesAdviceAndQuestionsOutsideThePortfolio(t *testing.T) {
	t.Parallel()
	const prefix = "Refuse only these three things: "
	var rules []string
	for _, p := range askPromptParts(kernel.Date{}) {
		if strings.HasPrefix(p.Text, prefix) {
			rules = append(rules, p.Text)
		}
	}
	if len(rules) != 1 {
		t.Fatalf("%d decline rules, want one", len(rules))
	}
	got := strings.Split(strings.TrimSuffix(strings.TrimPrefix(rules[0], prefix), "."), "; ")
	want := []string{"changing anything", "advice on what to buy, sell or hold", "a question that is not about the portfolio"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("the rule names %q, want exactly %q", got, want)
	}
	if other := sentencesNamingADeclinedClass(rules[0]); len(other) > 0 {
		t.Errorf("sentences besides the rule name a declined class: %q", other)
	}
}

// sentencesNamingADeclinedClass returns the sentences of the prompt other than
// rule that name one of the declined things.
func sentencesNamingADeclinedClass(rule string) []string {
	var out []string
	for _, p := range askPromptParts(kernel.Date{}) {
		if p.Text != rule && (strings.Contains(p.Text, "buy, sell or hold") || strings.Contains(p.Text, "changing anything")) {
			out = append(out, p.Text)
		}
	}
	return out
}

func TestPromptSaysWhatToAnswerAndHowToDecline(t *testing.T) {
	t.Parallel()
	prompt := askSystemPrompt(kernel.Date{})
	for _, s := range []string{"Greet back when you are greeted", "\"why\" questions too", "one friendly sentence"} {
		if !strings.Contains(prompt, s) {
			t.Errorf("the prompt lacks %q", s)
		}
	}
}
