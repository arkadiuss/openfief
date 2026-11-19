package ui

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/views/app/answer"
)

// askToday is the synthetic day the Ask tests ask on.
const askToday = "2026-03-31"

func TestAskSuggestions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		st   askState
		want []string
	}{
		{
			name: "an empty install",
			want: []string{"Allocation by currency", "Performance this year", "Value at end of last month"},
		},
		{
			name: "open items and a failing sync first, five at most",
			st:   askState{Open: 2, Failing: []string{"xtb-main"}, Accounts: []string{"Synthetic brokerage"}, Sleeves: []string{"Core"}},
			want: []string{"What needs me?", "Why is xtb-main failing?", "Show Synthetic brokerage", "Show sleeve Core", "Allocation by currency"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := askSuggestions(&tt.st)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("askSuggestions (-want +got):\n%s", diff)
			}
			for _, q := range got {
				if _, ok := answer.Parse(q, mustDate(t, askToday)); !ok {
					t.Errorf("suggestion %q is not a command the box understands", q)
				}
			}
		})
	}
}

func TestReturnPercent(t *testing.T) {
	t.Parallel()
	f := numberFormat{Group: narrowSpace, Decimal: ","}
	tests := []struct{ first, last, want string }{
		{"100", "109.4", "+9,4%"},
		{"100", "96.85", "-3,2%"},
		{"105", "105", "+0,0%"},
		{"0", "5", ""},
		{"x", "5", ""},
	}
	for _, tt := range tests {
		if got := returnPercent(tt.first, tt.last, f); got != tt.want {
			t.Errorf("returnPercent(%s, %s) = %q, want %q", tt.first, tt.last, got, tt.want)
		}
	}
}
