package answer

import "strings"

// askTarget is something a show command can name (D-194): an account, a
// sleeve or an instrument, its ID (an account's or a sleeve's ID, an
// ISIN), the label an answer gives it and every word it is found by, such
// as its name, its ID, a ticker or an ISIN.
type askTarget struct {
	Kind  ShowKind
	ID    string
	Label string
	Words []string
}

// matchTarget returns the target among targets that term names, of the
// kind asked for unless ShowAny: one with a word equal to term, case
// aside, first; else the first with a word that starts with it; else the
// first with a word that holds it. ok is false when none does.
func matchTarget(term string, kind ShowKind, targets []askTarget) (askTarget, bool) {
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
			if kind != ShowAny && t.Kind != kind {
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
		if t, ok := matchTarget(w, ShowAny, targets); ok {
			return t.ID, true
		}
	}
	return "", false
}

// Help is one command the Ask box understands, with an example.
type Help struct {
	Command, Example string
}

// KnownCommands are the commands the Ask box understands (D-194), each
// with an example, for a question it doesn't.
func KnownCommands() []Help {
	return []Help{
		{"What needs me", "What needs me?"},
		{"Show an account, a sleeve or an instrument", "Show bonds"},
		{"Why a sync fails", "Why is xtb failing?"},
		{"Allocation by asset class, currency, institution or sleeve", "Allocation by currency"},
		{"Performance this year, this month, last 12 months, last year, in a year or a month, between two dates, since a date or all", "Performance this year"},
		{"Value on a date, today, yesterday or at the end of last month, or its change in a month, a year or between two dates", "Value on 2026-01-31"},
	}
}
