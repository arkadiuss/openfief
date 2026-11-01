package views_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

func TestHistoryOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		events []event
		// first and seeded are day offsets; seeded -1 means not seeded.
		first, seeded int
	}{
		{
			name:   "whole history: a deposit and a buy",
			events: []event{{kind: "buy", acc: "a", day: 3, amount: "100", cur: "PLN", qty: "1"}, {kind: "deposit", acc: "a", day: 1, amount: "100", cur: "PLN"}},
			first:  1, seeded: -1,
		},
		{
			name: "seeded with an opening lot, later deposits",
			events: []event{
				{kind: "deposit", acc: "a", day: 9, amount: "50", cur: "PLN"},
				{kind: "opening-lot", acc: "a", day: 2, amount: "100", cur: "PLN", qty: "1"},
			},
			first: 2, seeded: 2,
		},
		{
			name: "opening cash entered on two days: the later one counts",
			events: []event{
				{kind: "opening-cash", acc: "a", day: 4, amount: "10", cur: "PLN"},
				{kind: "opening-cash", acc: "a", day: 6, amount: "10", cur: "USD"},
			},
			first: 4, seeded: 6,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			txs := make([]ledger.Transaction, len(tc.events))
			for i := range tc.events {
				txs[i] = tc.events[i].build(t, i)
			}
			want := views.AccountHistory{First: perfDay(t, tc.first)}
			if tc.seeded >= 0 {
				want.Seeded, want.SeededOn = true, perfDay(t, tc.seeded)
			}
			if diff := cmp.Diff(want, views.HistoryOf(txs)); diff != "" {
				t.Errorf("HistoryOf (-want +got):\n%s", diff)
			}
		})
	}
}

// history is an account's history as day offsets; seeded -1 means not
// seeded.
type history struct{ first, seeded int }

func TestChartStart(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		accounts []history
		want     int
		found    bool
	}{
		{name: "no accounts"},
		{name: "one account with its whole history: its first day", accounts: []history{{first: 5, seeded: -1}}, want: 5, found: true},
		{name: "one seeded account: its seeding day", accounts: []history{{first: 5, seeded: 5}}, want: 5, found: true},
		{
			name:     "the latest seeding of several",
			accounts: []history{{first: 2, seeded: 2}, {first: 10, seeded: 10}, {first: 4, seeded: 4}},
			want:     10, found: true,
		},
		{
			name:     "an account with its whole history from earlier doesn't move it",
			accounts: []history{{first: 0, seeded: -1}, {first: 7, seeded: 7}},
			want:     7, found: true,
		},
		{
			name:     "an account opened later with its whole history doesn't move it",
			accounts: []history{{first: 3, seeded: 3}, {first: 40, seeded: -1}},
			want:     3, found: true,
		},
		{
			name:     "no seeded account: the first day of any",
			accounts: []history{{first: 9, seeded: -1}, {first: 1, seeded: -1}},
			want:     1, found: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			accounts := make([]views.AccountHistory, len(tc.accounts))
			for i, h := range tc.accounts {
				accounts[i] = views.AccountHistory{First: perfDay(t, h.first)}
				if h.seeded >= 0 {
					accounts[i].Seeded, accounts[i].SeededOn = true, perfDay(t, h.seeded)
				}
			}
			got, found := views.ChartStart(accounts)
			want := kernel.Date{}
			if tc.found {
				want = perfDay(t, tc.want)
			}
			if found != tc.found || !got.Equal(want) {
				t.Errorf("ChartStart = %s, %t; want %s, %t", got, found, want, tc.found)
			}
		})
	}
}

func TestChosenStart(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		// chosen is a day offset; -1 means none chosen.
		chosen, want int
	}{
		{name: "none chosen: the default", chosen: -1, want: 10},
		{name: "a day after the default", chosen: 20, want: 20},
		{name: "a day between the first day and the default", chosen: 6, want: 6},
		{name: "a day before the first day starts on the first day", chosen: 1, want: 4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var chosen kernel.Date
			if tc.chosen >= 0 {
				chosen = perfDay(t, tc.chosen)
			}
			got := views.ChosenStart(chosen, perfDay(t, 4), perfDay(t, 10))
			if want := perfDay(t, tc.want); !got.Equal(want) {
				t.Errorf("ChosenStart = %s, want %s", got, want)
			}
		})
	}
}
