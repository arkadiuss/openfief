package answer

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestShowNamesTheNearTargetsWhenNothingMatches(t *testing.T) {
	t.Parallel()
	targets := []askTarget{
		{Kind: ShowAccount, Label: "Demo IBKR", Words: []string{"Demo IBKR"}},
		{Kind: ShowAccount, Label: "Demo XTB", Words: []string{"Demo XTB"}},
		{Kind: ShowInstrument, Label: "Demo Tech Co", Words: []string{"Demo Tech Co", "DTEC"}},
		{Kind: ShowInstrument, Label: "Demo World ETF", Words: []string{"Demo World ETF", "DWLD"}},
		{Kind: ShowInstrument, Label: "Demo Emerging Markets ETF", Words: []string{"Demo Emerging Markets ETF", "DEMG"}},
	}
	labels := func(ts []askTarget) []string {
		out := make([]string, len(ts))
		for i, t := range ts {
			out[i] = t.Label
		}
		return out
	}
	tests := []struct {
		name, term string
		want       []string
	}{
		{name: "two words, held by two names", term: "Demo ETF", want: []string{"Demo World ETF", "Demo Emerging Markets ETF", "Demo IBKR", "Demo XTB", "Demo Tech Co"}},
		{name: "a word of three letters", term: "tech", want: []string{"Demo Tech Co", "Demo IBKR", "Demo XTB", "Demo World ETF", "Demo Emerging Markets ETF"}},
		{name: "words of under three letters are no clue", term: "of it", want: []string{"Demo IBKR", "Demo XTB", "Demo Tech Co", "Demo World ETF", "Demo Emerging Markets ETF"}},
		{name: "no word held keeps the order", term: "zzz", want: []string{"Demo IBKR", "Demo XTB", "Demo Tech Co", "Demo World ETF", "Demo Emerging Markets ETF"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, labels(nearTargets(tt.term, targets))); diff != "" {
				t.Errorf("near targets of %q (-want +got):\n%s", tt.term, diff)
			}
		})
	}
	if got := labels(targets); got[0] != "Demo IBKR" || got[4] != "Demo Emerging Markets ETF" {
		t.Errorf("the targets were reordered: %v", got)
	}
}
