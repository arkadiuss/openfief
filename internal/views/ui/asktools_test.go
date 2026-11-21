package ui

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/app/answer"
)

func TestMostSensitive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		tiers map[string]bool
		want  string
	}{
		{nil, "T2"},
		{map[string]bool{"T2": true}, "T2"},
		{map[string]bool{"T2": true, "T0": true}, "T0"},
		{map[string]bool{"T1": true, "TF": true}, "TF"},
	}
	for _, tt := range tests {
		if got := mostSensitive(tt.tiers); got != tt.want {
			t.Errorf("mostSensitive(%v) = %s, want %s", tt.tiers, got, tt.want)
		}
	}
	if got := tierList(map[string]bool{"T2": true, "T0": true, "T1": true}); !slices.Equal(got, []string{"T0", "T1", "T2"}) {
		t.Errorf("tierList = %v, want most sensitive first", got)
	}
}

// TestAskArgDefaultsSayWhatLeavingAnArgumentOutMeans checks the defaults a
// harness reads (AskArgDefaults): a call that gives an argument its default
// asks the same command as the call without it, and the value tool's default is
// the latest valued day, which the caller names.
func TestAskArgDefaultsSayWhatLeavingAnArgumentOutMeans(t *testing.T) {
	t.Parallel()
	today := mustDate(t, askToday)
	for _, tool := range answer.Tools() {
		defaults := AskArgDefaults(tool.Name, today)
		if (tool.Defaults == nil) != (defaults == nil) {
			t.Errorf("%s: defaults %v with a table of %v", tool.Name, defaults, tool.Defaults == nil)
		}
		// The market price tool needs a query to ask anything, so its default is checked with one (answer's tests).
		if len(defaults) == 0 || tool.Name == "value" || tool.Name == "market_price" {
			continue
		}
		given, err := json.Marshal(defaults)
		if err != nil {
			t.Fatal(err)
		}
		withDefault, errWith := tool.Query(string(given), today)
		without, errWithout := tool.Query(`{}`, today)
		if errWith != nil || errWithout != nil {
			t.Fatalf("%s: %v, %v", tool.Name, errWith, errWithout)
		}
		if diff := cmp.Diff(without, withDefault); diff != "" {
			t.Errorf("%s: the default %s asks something else than leaving it out (-without +with):\n%s", tool.Name, given, diff)
		}
	}
}

func TestAskArgDefaultsAreTheDocumentedOnes(t *testing.T) {
	t.Parallel()
	today := mustDate(t, askToday)
	tests := []struct {
		tool   string
		latest kernel.Date
		want   map[string]any
	}{
		{tool: "value", latest: today, want: map[string]any{"date": askToday}},
		{tool: "value", latest: today.AddDays(-1), want: map[string]any{"date": "2026-03-30"}},
		{tool: "value", latest: kernel.Date{}},
		{tool: "allocation", latest: today, want: map[string]any{"by": "asset_class"}},
		{tool: "performance", latest: today, want: map[string]any{"period": "this_year"}},
		{tool: "market_price", latest: today, want: map[string]any{"period": "month"}},
		{tool: "show", latest: today},
		{tool: "sync_status", latest: today},
		{tool: "what_needs_me", latest: today},
		{tool: "calc", latest: today},
		{tool: "no_such_tool", latest: today},
	}
	for _, tt := range tests {
		if diff := cmp.Diff(tt.want, AskArgDefaults(tt.tool, tt.latest)); diff != "" {
			t.Errorf("%s defaults (-want +got):\n%s", tt.tool, diff)
		}
	}
}

// TestTheBoxAnswersItselfWhatItsRoutedAnswererDoes checks the one rule a
// harness reads (AskAnswersItself, D-203): the fixed commands, an empty
// question and one too long for a model go to the box's own answerer, and
// anything else to the model.
func TestTheBoxAnswersItselfWhatItsRoutedAnswererDoes(t *testing.T) {
	t.Parallel()
	today := mustDate(t, askToday)
	tests := []struct {
		q    string
		want bool
	}{
		{q: "What needs me?", want: true},
		{q: "Show me the Demo World ETF", want: true},
		{q: "Why is xtb failing?", want: true},
		{q: "Value on 2026-01-31", want: true},
		{q: "", want: true},
		{q: "   ", want: true},
		{q: strings.Repeat("a ", maxModelQuestionRunes), want: true},
		{q: "Who holds the Demo World ETF, and at what price?"},
		{q: "Say hello"},
		{q: "Why did my value drop in November 2025?"},
	}
	for _, tt := range tests {
		if got := AskAnswersItself(tt.q, today); got != tt.want {
			t.Errorf("AskAnswersItself(%q) = %v, want %v", tt.q, got, tt.want)
		}
	}
}
