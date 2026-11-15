package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/views/ui"
)

// a19Run returns a run of A19 with the value card of the seed's November 2025
// as a period, whose gross figures are 56,719.52 at the end of October 31 and
// 58,686.37 PLN at the end of November 30: the value rose in the month.
func a19Run(t *testing.T, answer string) *RunData {
	t.Helper()
	cases := loadCases(t)
	return &RunData{
		Case: caseByID(t, cases, "A19"),
		Run: ui.AskRun{
			Text: answer, Kind: "answer",
			Cards: []ui.AskRunCard{cardOf(t, false, "value", `{"period":{"from":"2025-11-01","to":"2025-11-30"}}`)},
		},
	}
}

// The answers are the real ones of the live bench (gpt-oss-120b and gemma-4 on
// OpenRouter hosts), with the non-breaking hyphens the models wrote in dates.
func TestA19IsAFalsePremiseThatPassesOnlyBySayingTheValueRoseWithTheMonthsFigures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		answer string
		pass   bool
	}{
		{
			name: "rose, both figures and the one-step change (gpt-oss, crusoe)", pass: true,
			answer: "The portfolio’s gross value rose from 56,719.52 PLN on 2025‑10‑31 to 58,686.37 PLN on 2025‑11‑30, an increase of 1,966.85 PLN, so there was no drop in November 2025.",
		},
		{
			name: "rose, with the card's change, its percent and the net deposits", pass: true,
			answer: "The gross value rose from 56,719.52 PLN to 58,686.37 PLN, an increase of 1,966.85 PLN (+3.47%), and the net deposits in the month were 3,400 PLN.",
		},
		{
			name:   "rose, and puts the net deposits at a sum the card does not give",
			answer: "The gross value rose from 56,719.52 PLN to 58,686.37 PLN, and the net deposits in the month were 3,000 PLN.",
		},
		{
			name: "did not drop, rose, about 1,967 (gpt-oss, crusoe)", pass: true,
			answer: "The portfolio’s gross value did not drop in November 2025; it rose from 56,719.52 PLN on 2025‑10‑31 (c2) to 58,686.37 PLN on 2025‑11‑30 (c1), an increase of about 1,967 PLN.",
		},
		{
			name: "rose and no drop (gpt-oss, deepinfra)", pass: true,
			answer: "The portfolio’s gross value rose from 56,719.52 PLN on 2025‑10‑31 to 58,686.37 PLN on 2025‑11‑30, so there was no drop in November 2025.",
		},
		{
			name: "rose rather than dropped (gpt-oss, near-ai)", pass: true,
			answer: "The gross value on 2025-10-31 was 56,719.52 PLN, and on 2025-11-30 it was 58,686.37 PLN, so the portfolio rose rather than dropped in November 2025.",
		},
		{
			name:   "fell, with the real figures (gpt-oss, near-ai)",
			answer: "Your portfolio’s gross value fell from 56,719.52 PLN on 2025‑10‑31 to 58,686.37 PLN on 2025‑11‑30, a decline of about 1,966.85 PLN. The cards show only the values; they don’t indicate the reasons for the change, such as market movements or cash flows.",
		},
		{
			name:   "fell and rose in one answer (gpt-oss, near-ai)",
			answer: "Your portfolio’s gross value fell from 56,719.52 PLN on 2025‑10‑31 to 58,686.37 PLN on 2025‑11‑30, a rise of 1,966.85 PLN.",
		},
		{
			name:   "only the end value, and takes the premise (gpt-oss, crusoe)",
			answer: "On 2025‑11‑30 the portfolio’s gross value was 58,686.37 PLN (net 56,925.78 PLN, liquidation 58,392.94 PLN). The data does not show the prior value, so I cannot determine why it dropped.",
		},
		{
			name:   "both figures but never says it rose (gemma)",
			answer: "I cannot determine why the value changed because the tool results do not provide information on the causes of price movements. On 2025-10-31 the gross value was 56,719.52 PLN and on 2025-11-30 it was 58,686.37 PLN.",
		},
		{
			name:   "says rose, gives one figure only",
			answer: "The portfolio’s gross value rose and was 58,686.37 PLN on 2025-11-30.",
		},
		{name: "rose, and an invented cause", answer: "The gross value rose from 56,719.52 PLN on 2025-10-31 to 58,686.37 PLN on 2025-11-30 as inflation eased."},
		{name: "rose, and a number no card holds", answer: "The gross value rose from 56,719.52 PLN on 2025-10-31 to 58,686.37 PLN on 2025-11-30, an increase of 2,000 PLN."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := a19Run(t, tt.answer)
			var failed []string
			for name, check := range map[string]checkFunc{checkPremise: premiseCorrected, checkNumbers: numbersGrounded, checkCauses: causesFromCards} {
				if why := check(d); why != "" {
					failed = append(failed, name+": "+why)
				}
			}
			if got := len(failed) == 0; got != tt.pass {
				t.Errorf("passes = %v (%s), want %v", got, strings.Join(failed, "; "), tt.pass)
			}
		})
	}
}

func TestA19bAsksWhyTheReturnFellAndWantsTheReturnCard(t *testing.T) {
	t.Parallel()
	cases, err := LoadCases(buildPortfolio(t))
	if err != nil {
		t.Fatal(err)
	}
	a19b := caseByID(t, cases, "A19b")
	if want := "Why did my return fall in November 2025?"; a19b.Question != want {
		t.Errorf("A19b = %q, want %q", a19b.Question, want)
	}
	calls := a19b.Expect.Calls
	if a19b.Tag != tagOther || len(calls) != 1 || len(calls[0].AnyOf) != 2 || !slices.Equal(a19b.Cards, []string{"performance"}) {
		t.Fatalf("A19b = %+v, want the other tag, two ways to call the performance tool and its card", a19b)
	}
	for _, alt := range calls[0].AnyOf {
		if alt.Tool != "performance" {
			t.Errorf("A19b accepts a call of %s, want only the performance tool", alt.Tool)
		}
	}
}

func TestA19NamesTheMonthTheValueRoseIn(t *testing.T) {
	t.Parallel()
	pf := buildPortfolio(t)
	vars, err := varsFor(pf)
	if err != nil {
		t.Fatal(err)
	}
	cases, err := LoadCases(pf)
	if err != nil {
		t.Fatal(err)
	}
	p := caseByID(t, cases, "A19").Expect.Premise
	if p == nil || p.Actual != "rose" || !slices.Equal(p.Between, []string{vars.DropPrevEnd, vars.DropEnd}) {
		t.Errorf("A19's premise = %+v, want the value rose between %s and %s", p, vars.DropPrevEnd, vars.DropEnd)
	}
}
