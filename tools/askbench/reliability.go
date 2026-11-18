package main

import "fmt"

// maxTransportFailurePercent is the most of its questions a host may fail in a
// run (D-207): above it the host fails reliability for that run, whatever its
// verdict.
const maxTransportFailurePercent = 5

// percent is the factor from a ratio to a percentage.
const percent = 100

// Why a run is void, the two reasons of RunReport.VoidReason.
const (
	voidBench       = "bench fault"
	voidCredentials = "credentials/credit"
)

// Reliability is how a host held up in one run of a model@host (D-207). The rule
// counts questions, not calls: a question is failed when, after its retry, it
// still has no usable answer because of the host's fault (no answer, a timeout,
// a 5xx, a 429, a bad or degenerate reply, a 404, or a 4xx that an identical
// request accepted elsewhere shows the host cannot serve). A question counts
// once however many calls it took, so a host that always sends a bad reply to
// one question fails one question and not two. The calls are counted too, for
// diagnosis: each call made is one, retries included.
type Reliability struct {
	// Questions is how many questions were asked of the host and counted: all
	// but those the matcher answers, the budget stopped and a void run, which is
	// no host's. FailedQuestions is how many of them ended with no usable answer
	// through the host's fault.
	Questions       int `json:"questions"`
	FailedQuestions int `json:"failedQuestions"`
	// Calls is how many calls were made and TransportFailures how many of them
	// failed through the host's fault, counted before any retry hides them.
	Calls             int `json:"calls"`
	TransportFailures int `json:"transportFailures"`
	// Retried is how many failed calls were retried and Recovered how many of
	// those retries answered.
	Retried   int `json:"retried"`
	Recovered int `json:"recovered"`
	// Rate is FailedQuestions over Questions, 0 for a run with no questions.
	Rate float64 `json:"rate"`
	// Failed is true when the rate exceeds the threshold, 5%. It is the zero
	// value for a host that held up.
	Failed bool `json:"failed"`
}

// add counts the calls and the questions of runs.
func (r *Reliability) add(runs []RunReport) {
	for i := range runs {
		run := &runs[i]
		r.Calls += run.Attempts
		r.TransportFailures += run.TransportFailures
		r.Retried += len(run.Retries)
		for _, rt := range run.Retries {
			if rt.Recovered {
				r.Recovered++
			}
		}
		if run.countsForReliability() {
			r.Questions++
			if run.failedByHost() {
				r.FailedQuestions++
			}
		}
	}
}

// countsForReliability reports whether the run is a question that was asked of
// the host: not one the matcher answered, the budget stopped, or a fault that is
// no host's voided.
func (r *RunReport) countsForReliability() bool {
	return !r.Matcher && !r.NotRun && !r.BudgetSpent && !r.Void
}

// failedByHost reports whether the question ended with no usable answer through
// the host's fault: a call that ended it, after any retry, failed as the host's.
func (r *RunReport) failedByHost() bool {
	for i := range r.CallErrors {
		if r.CallErrors[i].Fault == faultHost {
			return true
		}
	}
	return false
}

// finish computes the rate and the pass or fail from the counts, exactly: the
// threshold is compared in integers.
func (r *Reliability) finish() {
	r.Rate = 0
	if r.Questions > 0 {
		r.Rate = float64(r.FailedQuestions) / float64(r.Questions)
	}
	r.Failed = r.FailedQuestions*percent > r.Questions*maxTransportFailurePercent
}

// text is the reliability as one cell: "pass (0.3% of questions, 2 of 145; 3
// failed calls)" or "fail (7.5% of questions, 9 of 120; 11 failed calls)".
func (r *Reliability) text() string {
	word := verdictPass
	if r.Failed {
		word = verdictFail
	}
	return fmt.Sprintf("%s (%.1f%% of questions, %d of %d; %d failed calls)", word, r.Rate*percent, r.FailedQuestions, r.Questions, r.TransportFailures)
}
