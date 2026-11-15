package main

// ScorerVersion is the version of the bench's scoring rules, written into
// report.json (`scorerVersion`), into each trimmed result file and shown beside
// every verdict on the results page, so that two verdicts can be compared only
// when they were graded by the same rules.
//
// Bump it whenever any scoring rule changes, that is, whenever the same answer
// could be graded differently: a check's logic or its word lists, a case's
// expected calls, cards, checks or premise (cases.yaml), a tag, the verdict's
// bar (verdict.go) or what a check reads. Then re-score the runs whose verdicts
// the change could move (`askbench -rescore`), which writes the new version
// into the re-graded report. Do not bump it for a change that cannot alter a
// grade: a refactor, a new report field, a doc.
//
// It is no part of the configuration hash. The hash says what the model sees
// (the model, its host and routing, its reasoning control, the system prompt,
// the tools' schemas and D-208's limits), and the grader is not in it: a change
// of scorer leaves a model@host's hash, and its D-221 streak of the same
// configuration, as it was.
//
// Versions: 1 is every grading before versioning (runs 2 to 14-2 as first
// graded, whose rules changed from run to run); 2 is A19 accepting the start
// and end values or the card's change (D-217, D-219).
const ScorerVersion = 2

// unversionedScorer is the version a report without one reads as.
const unversionedScorer = 1

// scorerOf returns the scorer version of a report: its own, or 1 for a report
// written before versioning.
func scorerOf(r *Report) int {
	if r.ScorerVersion <= 0 {
		return unversionedScorer
	}
	return r.ScorerVersion
}
