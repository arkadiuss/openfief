package main

import (
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/ui"
)

// classesOf returns the gross values by asset class the arithmetic checks read:
// 10 000 in all, of which 2 000 is stocks, 5 000 bonds and 3 000 cash.
func classesOf(t *testing.T) map[string]kernel.Decimal {
	t.Helper()
	out := map[string]kernel.Decimal{}
	for class, v := range map[string]string{"equity": "1000", "etf": "1000", "bond": "5000", "cash": "3000"} {
		d, err := kernel.ParseDecimal(v)
		if err != nil {
			t.Fatal(err)
		}
		out[class] = d
	}
	return out
}

// thirtySeventy is the case of the author's example over those classes.
func thirtySeventy() *Case {
	return &Case{Expect: Expect{Rebalance: &Rebalance{
		Groups:  map[string][]string{"stocks": {"equity", "etf"}, "bonds": {"bond"}},
		Weights: map[string]int{"stocks": 30, "bonds": 70},
	}}}
}

func TestTheRebalanceCheckGradesTheAmountsByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		answer string
		fail   string // a part of the reason, or "" for a pass
	}{
		{name: "the whole portfolio with its cash", answer: "Move 1,000.00 PLN into stocks and 2,000.00 PLN into bonds, before taxes and fees."},
		{name: "rounded to whole amounts", answer: "About 1,000 PLN more in stocks and 2,000 PLN more in bonds."},
		{name: "only what the named groups hold", answer: "Stocks need 100.00 PLN more and bonds 100.00 PLN less."},
		{name: "one amount of the two", answer: "Move 1,000.00 PLN into stocks.", fail: "not those computed from the allocation"},
		{name: "amounts the model worked out wrongly", answer: "Move 1,500.00 PLN into stocks and 2,500.00 PLN into bonds.", fail: "2000, 1000 or 100, 100"},
		{name: "no amount", answer: "I can't say.", fail: "not those computed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := answered(thirtySeventy(), tt.answer, nil)
			d.Classes = classesOf(t)
			got := rebalanceAmounts(d)
			if (tt.fail == "") != (got == "") || !strings.Contains(got, tt.fail) {
				t.Errorf("rebalanceAmounts() = %q, want a failure containing %q", got, tt.fail)
			}
		})
	}
}

func TestTheRebalanceCheckRefusesACaseItCannotGrade(t *testing.T) {
	t.Parallel()
	c := thirtySeventy()
	c.Expect.Rebalance.Groups["bonds"] = []string{"gilts"}
	d := answered(c, "Move 1 PLN.", nil)
	d.Classes = classesOf(t)
	if got := rebalanceAmounts(d); !strings.Contains(got, "no asset class gilts") {
		t.Errorf("rebalanceAmounts() = %q, want the case's own fault named", got)
	}
}

func TestTheExactCheckWantsEveryDigit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, answer string
		want         []string
		fail         bool
	}{
		{name: "exact", answer: "0.1 plus 0.2 is 0.3.", want: []string{"0.3"}},
		{name: "a float's tail is not 0.3", answer: "It is 0.30000000000000004.", want: []string{"0.3"}, fail: true},
		{name: "a long decimal in full", answer: "That is 370,370,367.370370367.", want: []string{"370370367.370370367"}},
		{name: "a long decimal cut short", answer: "That is about 370,370,367.37.", want: []string{"370370367.370370367"}, fail: true},
		{name: "a Polish decimal comma", answer: "To jest 0,3.", want: []string{"0.3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := answered(&Case{Exact: tt.want}, tt.answer, nil)
			if got := answerHasExact(d); (got != "") != tt.fail {
				t.Errorf("answerHasExact() = %q, want a failure: %v", got, tt.fail)
			}
		})
	}
}

func TestTheOnlyNumbersCheckLetsNoAmountThrough(t *testing.T) {
	t.Parallel()
	only := []string{"40", "80", "100", "20"}
	pass := answered(&Case{Only: only}, "The weights add up to 80%, not 100%, so I can't work out amounts.", []ui.AskRunCard{})
	if got := answerHasOnly(pass); got != "" {
		t.Errorf("answerHasOnly() = %q, want a pass", got)
	}
	amount := answered(&Case{Only: only}, "Move 1,234.00 PLN into bonds.", nil)
	if got := answerHasOnly(amount); !strings.Contains(got, "1,234.00") {
		t.Errorf("answerHasOnly() = %q, want the amount named", got)
	}
}

func TestTheFeesCheckWantsBeforeTaxesAndFeesSaidPlainly(t *testing.T) {
	t.Parallel()
	if got := saysBeforeFees(answered(&Case{}, "Move 1,000 PLN, before taxes and fees.", nil)); got != "" {
		t.Errorf("saysBeforeFees() = %q, want a pass", got)
	}
	if got := saysBeforeFees(answered(&Case{}, "Move 1,000 PLN.", nil)); got == "" {
		t.Error("saysBeforeFees() passed an answer that does not say it")
	}
}

// TestCalcResultsCountAsShownFigures checks that a figure only a calc card holds
// is grounded for the numbers check, and the model's own sum is not.
func TestCalcResultsCountAsShownFigures(t *testing.T) {
	t.Parallel()
	alloc := ui.AskRunCard{Tool: "allocation", Title: "By asset class, gross, on 2026-10-28", Rows: []ui.AskRunRow{{Label: "bond", Figure: "5000 PLN"}, {Label: "cash", Figure: "3000 PLN"}}}
	calc := ui.AskRunCard{Tool: "calc", Title: "(5000 + 3000 + 2000) * 70% - 4500 = 2500", Rows: []ui.AskRunRow{{Label: "Result", Figure: "2500"}}}
	with := answered(&Case{Question: "70% bonds?"}, "Move 2500 PLN into bonds.", []ui.AskRunCard{alloc, calc})
	if got := numbersGrounded(with); got != "" {
		t.Errorf("numbersGrounded() = %q, want the calc result as a shown figure", got)
	}
	without := answered(&Case{Question: "70% bonds?"}, "Move 2500 PLN into bonds.", []ui.AskRunCard{alloc})
	if got := numbersGrounded(without); !strings.Contains(got, "2500") {
		t.Errorf("numbersGrounded() = %q, want the model's own sum refused", got)
	}
}
