package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// maxListedFailing is how many failing cases a table cell lists before it
	// only counts them.
	maxListedFailing = 8
	percentFactor    = 100
)

// askEvalPage renders the page the Ask box links to (D-222): each model and
// host with its latest verdict, then every run, newest first, from the
// committed results (internal/views/ui/evalresults), so the page and the box's "not
// evaluated" note read the same files.
func askEvalPage(results []evalResult) string {
	var b strings.Builder
	b.WriteString("---\ntitle: How models did in the Ask box evaluation\n" +
		"description: Each model and host the Ask box's eval has covered, with its verdict, failing cases, reliability, cost and run dates, regenerated from the bench's committed results.\nweight: 40\n---\n\n")
	b.WriteString("Generated on every build from the bench's results, committed under [internal/views/ui/evalresults](../../../../internal/views/ui/evalresults/). " +
		"The guide leads with a table of one row per model and host, with comments, and has the method, what each tag asks and what we learned: [How models did on the Ask box evaluation](../../../../docs/guides/ask-box-evaluation.md). " +
		"the bar is D-220, the reliability rule D-207 and the demo route D-221, and the cases and grading are D-217 and D-219. " +
		"All of the data is synthetic: the bench asks over the demo's seeded record, never a real account. A model the eval has not covered is named as not evaluated in the Ask box.\n")

	b.WriteString("\n## Where each model stands\n\n")
	b.WriteString("The latest run of every case for each model and host. A host passes the eval when its verdict is pass and its reliability holds; it joins the demo route after two such runs in a row (D-221), and none has.\n\n")
	b.WriteString("| Model@host | Reasoning | Latest verdict | Failing cases | Reliability | USD per 1k questions | Run date | Config |\n|---|---|---|---|---|---|---|---|\n")
	seen := map[string]bool{}
	for i := range results {
		r := &results[i]
		if !r.stands() || seen[r.label()+"|"+r.Reasoning] {
			continue
		}
		seen[r.label()+"|"+r.Reasoning] = true
		fmt.Fprintf(&b, "| %s | %s | `%s` |\n", strings.Join(resultCells(r), " | "), r.Date, short(r.ConfigHash))
	}

	b.WriteString("\n## Every run\n\n")
	b.WriteString("Newest first. *Failing cases* give runs passed over runs needed (D-220); *unscored* runs are ones whose call failed. A run marked *void* says nothing of its model: no call of it got an answer, or it ran before the verdict existed. A run marked *diagnostic* asked only some cases, which cannot prove safety, so it carries no verdict. Void and diagnostic runs are left out of the table above and of the guide's. Runs before the verdict existed (D-220) show how many cases were below pass^5. Cost is measured, or marked estimated or mixed where tokens were priced from the listing. The scorer version beside each verdict says which scoring rules graded it (the bench's `ScorerVersion`, bumped whenever a scoring rule changes), so verdicts of different versions are not comparable; a run graded before versioning shows scorer 1, and a run graded again with the stored answers and no model call shows the version that re-graded it. The bench's report schema and the configuration hash (model, host, reasoning, structured-output flag, prompt, tool schemas and limits) are in the last columns: the same hash is the same configuration.\n\n")
	b.WriteString("| Run | Date | Model@host | Reasoning | Verdict | Failing cases | Reliability | USD per 1k questions | Bench schema | Config |\n|---|---|---|---|---|---|---|---|---|---|\n")
	for i := range results {
		r := &results[i]
		fmt.Fprintf(&b, "| %s | %s | %s | %d | `%s` |\n", r.Run, r.Date, strings.Join(resultCells(r), " | "), r.BenchSchema, short(r.ConfigHash))
	}
	return b.String()
}

// resultCells are a result's cells, from the model to its cost.
func resultCells(r *evalResult) []string {
	reasoning := r.Reasoning
	if reasoning == "" {
		reasoning = "default"
	}
	host := r.label()
	if r.Host == "" {
		host += " (default routing)"
	}
	verdict := verdictText(r)
	if !r.full() {
		verdict += " (diagnostic: " + strings.Join(r.Diagnostic, ", ") + " only)"
	}
	return []string{tableCell(host), reasoning, verdict, failingText(r), reliabilityText(r), costText(r)}
}

// verdictText is a run's verdict as both pages show it: "pass (scorer 2)", "fail
// (scorer 1, before versioning)", or for a void run what made it void, with no
// scorer version, since nothing was graded.
func verdictText(r *evalResult) string {
	switch {
	case r.VoidCredentials > 0:
		return "void (credentials/credit)"
	case r.Void > 0:
		return "void (bench fault)"
	case !r.Answered:
		return "void (no answer: every call failed, rate limited)"
	case r.Verdict == verdictNone:
		return "void (no verdict: before D-220)"
	}
	text := r.Verdict + " (" + r.scorerText()
	if r.Parallel > 0 && r.Parallel != defaultParallel {
		text += fmt.Sprintf(", parallel %d", r.Parallel)
	}
	return text + ")"
}

func failingText(r *evalResult) string {
	switch {
	case !r.Answered:
		return "none answered"
	case r.Verdict == verdictNone:
		return strconv.Itoa(r.UnderBar) + " of " + strconv.Itoa(r.Cases) + " cases under pass^5"
	case len(r.Failing) == 0:
		return "none"
	case len(r.Failing) > maxListedFailing:
		return strconv.Itoa(len(r.Failing)) + " cases, most with runs unscored"
	}
	parts := make([]string, len(r.Failing))
	for i, f := range r.Failing {
		parts[i] = fmt.Sprintf("%s (%s) %d/%d", f.Case, f.Tag, f.Passed, f.Need)
		if f.Unscored > 0 {
			parts[i] += fmt.Sprintf(" +%d unscored", f.Unscored)
		}
	}
	return strings.Join(parts, ", ")
}

func reliabilityText(r *evalResult) string {
	text := "not recorded"
	if rel := r.Reliability; rel != nil {
		text = fmt.Sprintf("%.1f%% transport failures (%d of %d calls)", rel.Rate*percentFactor, rel.TransportFailures, rel.Calls)
		if rel.Questions > 0 {
			text = fmt.Sprintf("%.1f%% of questions failed (%d of %d; %d failed calls)", rel.Rate*percentFactor, rel.FailedQuestions, rel.Questions, rel.TransportFailures)
		}
		if rel.Failed {
			text += ", fails"
		}
	}
	if r.Unscored > 0 && r.Answered {
		text += fmt.Sprintf("; %d runs unscored", r.Unscored)
	}
	return text
}

func costText(r *evalResult) string {
	if r.CostPer1k == "" {
		return notAvailable
	}
	f, err := strconv.ParseFloat(r.CostPer1k, 64)
	if err != nil {
		return notAvailable
	}
	text := strconv.FormatFloat(f, 'f', 3, 64)
	if r.CostSource != "measured" && r.CostSource != "" {
		text += " (" + r.CostSource + ")"
	}
	return text
}

func short(hash string) string {
	const n = 8
	if len(hash) > n {
		return hash[:n]
	}
	return hash
}
