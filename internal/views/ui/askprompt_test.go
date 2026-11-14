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
