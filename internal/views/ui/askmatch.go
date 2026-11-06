package ui

import (
	"strings"
)

// maxSuggestions is how many suggestion chips the Ask box shows.
const maxSuggestions = 5

// askTarget is something a show command can name (D-194): an account, a
// sleeve or an instrument, its ID (an account's or a sleeve's ID, an
// ISIN), the label an answer gives it and every word it is found by, such
// as its name, its ID, a ticker or an ISIN.
type askTarget struct {
	Kind  showKind
	ID    string
	Label string
	Words []string
}

// matchTarget returns the target among targets that term names, of the
// kind asked for unless showAny: one with a word equal to term, case
// aside, first; else the first with a word that starts with it; else the
// first with a word that holds it. ok is false when none does.
func matchTarget(term string, kind showKind, targets []askTarget) (askTarget, bool) {
	term = strings.ToLower(strings.TrimSpace(term))
	if term == "" {
		return askTarget{}, false
	}
	tests := []func(word string) bool{
		func(w string) bool { return w == term },
		func(w string) bool { return strings.HasPrefix(w, term) },
		func(w string) bool { return strings.Contains(w, term) },
	}
	for _, test := range tests {
		for _, t := range targets {
			if kind != showAny && t.Kind != kind {
				continue
			}
			for _, w := range t.Words {
				if test(strings.ToLower(w)) {
					return t, true
				}
			}
		}
	}
	return askTarget{}, false
}

// matchInstance returns the connector instance among instances that term
// names: one equal to it first, case aside, then one that starts with it
// or holds it, so "xtb" names "xtb-main". ok is false when none does.
func matchInstance(term string, instances []string) (string, bool) {
	words := strings.Fields(strings.ToLower(term))
	targets := make([]askTarget, 0, len(instances))
	for _, inst := range instances {
		targets = append(targets, askTarget{ID: inst, Words: []string{inst}})
	}
	for _, w := range append([]string{strings.Join(words, " ")}, words...) {
		if t, ok := matchTarget(w, showAny, targets); ok {
			return t.ID, true
		}
	}
	return "", false
}

// askState is what the Ask box's suggestion chips are drawn from: how
// many items are open, the connector instances whose sync keeps failing,
// and the accounts' and sleeves' names, in their order.
type askState struct {
	Open     int
	Failing  []string
	Accounts []string
	Sleeves  []string
}

// askSuggestions returns at most maxSuggestions questions drawn from st,
// each one the box understands: what needs the owner while anything is
// open, why each failing sync fails, the first account and sleeve, then
// the allocation by currency and this year's performance.
func askSuggestions(st *askState) []string {
	var out []string
	if st.Open > 0 {
		out = append(out, "What needs me?")
	}
	for _, inst := range st.Failing {
		out = append(out, "Why is "+inst+" failing?")
	}
	if len(st.Accounts) > 0 {
		out = append(out, "Show "+st.Accounts[0])
	}
	if len(st.Sleeves) > 0 {
		out = append(out, "Show sleeve "+st.Sleeves[0])
	}
	out = append(out, "Allocation by currency", "Performance this year", "Value at end of last month")
	if len(out) > maxSuggestions {
		out = out[:maxSuggestions]
	}
	return out
}

// askHelp is one command the Ask box understands, with an example.
type askHelp struct {
	Command, Example string
}

// knownCommands are the commands the Ask box understands (D-194), each
// with an example, for a question it doesn't.
func knownCommands() []askHelp {
	return []askHelp{
		{"What needs me", "What needs me?"},
		{"Show an account, a sleeve or an instrument", "Show bonds"},
		{"Why a sync fails", "Why is xtb failing?"},
		{"Allocation by asset class, currency, institution or sleeve", "Allocation by currency"},
		{"Performance this year, this month, last 12 months, last year, in a year, since a date or all", "Performance this year"},
		{"Value on a date, today, yesterday or at the end of last month", "Value on 2026-01-31"},
	}
}
