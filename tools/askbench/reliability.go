package main

import "fmt"

// maxTransportFailurePercent is the most a host may fail at the transport level
// in a run (D-207): above it the host fails reliability for that run, whatever
// its verdict.
const maxTransportFailurePercent = 5

// percent is the factor from a ratio to a percentage.
const percent = 100

// Reliability is how a host held up in one run of a model@host: the share of
// its calls that failed at the transport level (no HTTP answer, a timeout, a
// 5xx or a 429 that its retry did not mend), counted before any retry hides them. Every call counts, retries
// included, since each is a call made.
type Reliability struct {
	// Calls is how many calls were made and TransportFailures how many of them
	// failed at the transport level.
	Calls             int `json:"calls"`
	TransportFailures int `json:"transportFailures"`
	// Retried is how many failed calls were retried and Recovered how many of
	// those retries answered.
	Retried   int `json:"retried"`
	Recovered int `json:"recovered"`
	// Rate is TransportFailures over Calls, 0 for a run with no calls.
	Rate float64 `json:"rate"`
	// Failed is true when the rate exceeds the threshold, 5%. It is the zero
	// value for a host that held up.
	Failed bool `json:"failed"`
}

// add counts the calls of runs.
func (r *Reliability) add(runs []RunReport) {
	for i := range runs {
		r.Calls += runs[i].Attempts
		r.TransportFailures += runs[i].TransportFailures
		r.Retried += len(runs[i].Retries)
		for _, rt := range runs[i].Retries {
			if rt.Recovered {
				r.Recovered++
			}
		}
	}
}

// finish computes the rate and the pass or fail from the counts, exactly: the
// threshold is compared in integers.
func (r *Reliability) finish() {
	r.Rate = 0
	if r.Calls > 0 {
		r.Rate = float64(r.TransportFailures) / float64(r.Calls)
	}
	r.Failed = r.TransportFailures*percent > r.Calls*maxTransportFailurePercent
}

// text is the reliability as one cell: "pass (0.8%, 2/250)" or "fail (7.5%, 9/120)".
func (r *Reliability) text() string {
	word := verdictPass
	if r.Failed {
		word = verdictFail
	}
	return fmt.Sprintf("%s (%.1f%%, %d/%d)", word, r.Rate*percent, r.TransportFailures, r.Calls)
}
