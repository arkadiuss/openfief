package ui

import "github.com/arkadiuss/openfief/internal/views/app/answer"

// The data tiers of D-206 (answer.TierT0 and the others), which the Ask box
// tags every message it sends a model with; the core's router decides which
// models may receive it.
const (
	tierT0 = answer.TierT0
	tierTF = answer.TierTF
	tierT1 = answer.TierT1
	tierT2 = answer.TierT2
)

// tierWords describes a data tier in words, for the notice of a card the
// model was not allowed to see.
func tierWords(tier string) string {
	switch tier {
	case tierT0:
		return "whole-portfolio holdings, amounts, transactions or tax data"
	case tierTF:
		return "one fief's own lease, cash, positions or orders"
	case tierT1:
		return "a thesis"
	case tierT2:
		return "public data"
	}
	return ""
}

// tierOrder returns the data tiers, most sensitive first, to pick the tier a
// message that holds data of several carries.
func tierOrder() []string { return []string{tierT0, tierTF, tierT1, tierT2} }

// mostSensitive returns the first of tierOrder among tiers, or T2 when it
// holds none.
func mostSensitive(tiers map[string]bool) string {
	for _, t := range tierOrder() {
		if tiers[t] {
			return t
		}
	}
	return tierT2
}
