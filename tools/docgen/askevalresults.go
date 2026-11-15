package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"slices"
	"strings"
)

// The trimmed result files of the Ask box bench's runs (D-222), one file per
// run, are committed beside the UI that embeds them to say whether the
// answering model was evaluated; docgen renders the results page from the
// same files, so the two cannot disagree.
const resultsDir = "internal/views/ui/evalresults"

// Verdicts of a run. None is a run that has no D-220 verdict: one from before
// the verdict existed, or that no call answered.
const (
	verdictPass = "pass"
	verdictFail = "fail"
	verdictNone = "none"
)

// evalFailing is a case that missed its D-220 bar.
type evalFailing struct {
	Case     string `json:"case"`
	Tag      string `json:"tag"`
	Passed   int    `json:"passed"`
	Need     int    `json:"need"`
	Unscored int    `json:"unscored,omitempty"`
}

// evalReliability is D-207's measure of a host in one run.
type evalReliability struct {
	Calls             int     `json:"calls"`
	TransportFailures int     `json:"transportFailures"`
	Rate              float64 `json:"rate"`
	Failed            bool    `json:"failed"`
}

// evalResult is one model on one host in one bench run.
type evalResult struct {
	// Run names the run, such as "9-1"; Date is its UTC day, YYYY-MM-DD.
	Run  string `json:"run"`
	Date string `json:"date"`
	// Model is the OpenRouter model id and Host the provider slug it was
	// pinned to, empty for default routing.
	Model     string `json:"model"`
	Host      string `json:"host,omitempty"`
	Reasoning string `json:"reasoning,omitempty"`
	Verdict   string `json:"verdict"`
	// Diagnostic lists the cases of a run of some cases only (-case), which
	// proves nothing of safety and so of the verdict.
	Diagnostic []string      `json:"diagnostic,omitempty"`
	Failing    []evalFailing `json:"failing,omitempty"`
	// Cases is how many cases the run scored; UnderBar how many missed pass^5
	// in a run with no verdict.
	Cases    int `json:"cases"`
	UnderBar int `json:"underBar,omitempty"`
	// Unscored is the runs a failed call ended.
	Unscored int `json:"unscored,omitempty"`
	// Void is how many of those runs the bench's own fault ended: the result
	// is void, and the model is run again after the fix.
	Void int `json:"void,omitempty"`
	// Answered is false when no call of the model got an answer.
	Answered    bool             `json:"answered"`
	Reliability *evalReliability `json:"reliability,omitempty"`
	// CostPer1k is the USD per 1,000 questions as a decimal text; CostSource
	// is measured, estimated or mixed.
	CostPer1k  string `json:"costPer1k,omitempty"`
	CostSource string `json:"costSource,omitempty"`
	ConfigHash string `json:"configHash"`
	// BenchSchema is the report.json schema of the bench that wrote the run.
	BenchSchema int `json:"benchSchema"`
	// ScorerVersion is the version of the scoring rules that graded the run;
	// a file without one is from before versioning and reads as 1.
	ScorerVersion int `json:"scorerVersion,omitempty"`
	// Parallel is how many questions the run asked at once when that was not
	// D-208's 3; the bench's report does not carry it, so `askeval import
	// -parallel` records it. A file without it was made at 3.
	Parallel int `json:"parallel,omitempty"`
}

// unversionedScorer is the scorer version a result file without one reads as.
const unversionedScorer = 1

// scorerText is the scorer version that graded the run, as the page shows it
// beside the verdict: "scorer 2", and for a file from before versioning "scorer
// 1, before versioning".
func (r *evalResult) scorerText() string {
	if r.ScorerVersion <= 0 {
		return fmt.Sprintf("scorer %d, before versioning", unversionedScorer)
	}
	return fmt.Sprintf("scorer %d", r.ScorerVersion)
}

// label is "model@host", or the model alone for default routing.
func (r *evalResult) label() string {
	if r.Host == "" {
		return r.Model
	}
	return r.Model + "@" + r.Host
}

// full reports whether the run asked every case, so that its verdict stands.
func (r *evalResult) full() bool { return len(r.Diagnostic) == 0 }

// loadResults reads every result file of repo, newest first: by date, and
// within a date the files' order reversed, so a later run comes first.
func loadResults(repo fs.FS) ([]evalResult, error) {
	names, err := fs.Glob(repo, resultsDir+"/*.json")
	if err != nil {
		return nil, fmt.Errorf("list results: %w", err)
	}
	slices.Sort(names)
	var all []evalResult
	for _, name := range names {
		data, readErr := fs.ReadFile(repo, name)
		if readErr != nil {
			return nil, fmt.Errorf("read %s: %w", name, readErr)
		}
		var rs []evalResult
		if jsonErr := json.Unmarshal(data, &rs); jsonErr != nil {
			return nil, fmt.Errorf("decode %s: %w", name, jsonErr)
		}
		all = append(all, rs...)
	}
	slices.Reverse(all)
	slices.SortStableFunc(all, func(a, b evalResult) int { return strings.Compare(b.Date, a.Date) })
	return all, nil
}
