package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// historySchema is the version of the history file's shape.
const historySchema = 1

// historyKeep is how many runs of one model@host the history keeps, the newest:
// a streak never needs more, and the committed file stays small.
const historyKeep = 52

// consecutiveRuns is D-221's bar: a model@host joins the demo route after this
// many consecutive passing runs and leaves it after this many consecutive
// failing weekly runs.
const consecutiveRuns = 2

// HistoryEntry is one run of a model@host: when it was made, its verdict
// (D-220), its host's reliability (D-207) and the configuration hash of the run,
// so a run of another configuration is never counted with it.
type HistoryEntry struct {
	// Date is the day the run was made, YYYY-MM-DD in UTC.
	Date string `json:"date"`
	// Verdict and Reliability are "pass" or "fail".
	Verdict     string `json:"verdict"`
	Reliability string `json:"reliability"`
	// TransportFailureRate is the run's rate, for the reader of the file.
	TransportFailureRate float64 `json:"transportFailureRate"`
	ConfigHash           string  `json:"configHash"`
	// ScorerVersion is the version of the scoring rules that graded the run
	// (ScorerVersion); an entry without one is from before versioning. It is
	// for the reader and does not end a streak; the entry of a run that was
	// re-scored (rescore.go) is edited to the version that graded it.
	ScorerVersion int `json:"scorerVersion,omitempty"`
}

// passed is whether the run passed D-221's bar: the verdict and the reliability.
func (e *HistoryEntry) passed() bool {
	return e.Verdict == verdictPass && e.Reliability == verdictPass
}

// History is each model@host's runs, oldest first, keyed by its label (the
// model id with its hosts). It is a small file in git: the bench rewrites it
// only when -history names it, and a person commits the result after review.
type History struct {
	Schema int                       `json:"schema"`
	Runs   map[string][]HistoryEntry `json:"runs"`
}

// NewHistory returns a history with no runs.
func NewHistory() *History {
	return &History{Schema: historySchema, Runs: map[string][]HistoryEntry{}}
}

// LoadHistory reads the history at path; a file that does not exist is an empty
// history, a file that does not read is an error.
func LoadHistory(path string) (*History, error) {
	raw, err := os.ReadFile(filepath.Clean(path))
	if errors.Is(err, fs.ErrNotExist) {
		return NewHistory(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: reading the history %s: %w", ErrRun, path, err)
	}
	h := NewHistory()
	if err := json.Unmarshal(raw, h); err != nil {
		return nil, fmt.Errorf("%w: the history %s is no history: %w", ErrRun, path, err)
	}
	if h.Schema != historySchema {
		return nil, fmt.Errorf("%w: the history %s has schema %d, want %d", ErrRun, path, h.Schema, historySchema)
	}
	if h.Runs == nil {
		h.Runs = map[string][]HistoryEntry{}
	}
	return h, nil
}

// Save writes the history to path, labels in order, so a diff shows only what a
// run added.
func (h *History) Save(path string) error {
	raw, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: encoding the history: %w", ErrRun, err)
	}
	if err := os.WriteFile(filepath.Clean(path), append(raw, '\n'), fileMode); err != nil {
		return fmt.Errorf("%w: writing the history %s: %w", ErrRun, path, err)
	}
	return nil
}

// Add records a run of label, keeping the newest historyKeep.
func (h *History) Add(label string, e *HistoryEntry) {
	h.Runs[label] = append(h.Runs[label], *e)
	if runs := h.Runs[label]; len(runs) > historyKeep {
		h.Runs[label] = runs[len(runs)-historyKeep:]
	}
}

// Streak is how many runs in a row, the newest, the same way: Passes when the
// latest run passed, Fails when it did not, and the other is zero.
type Streak struct {
	Passes int `json:"passes"`
	Fails  int `json:"fails"`
}

// Eligible is whether the model@host may join the demo route (D-221): two
// consecutive passes, verdict and reliability both.
func (s Streak) Eligible() bool { return s.Passes >= consecutiveRuns }

// Failing is whether it has failed two consecutive runs, the bar for leaving
// the route (D-221).
func (s Streak) Failing() bool { return s.Fails >= consecutiveRuns }

// StreakOf returns the streak of label: the newest runs that agree with the
// latest on passing and on the configuration hash. A run of another
// configuration is a different model@host in effect, so it ends the streak.
func (h *History) StreakOf(label string) Streak {
	runs := h.Runs[label]
	if len(runs) == 0 {
		return Streak{}
	}
	last := &runs[len(runs)-1]
	n := 0
	for i := len(runs) - 1; i >= 0; i-- {
		if runs[i].ConfigHash != last.ConfigHash || runs[i].passed() != last.passed() {
			break
		}
		n++
	}
	if last.passed() {
		return Streak{Passes: n}
	}
	return Streak{Fails: n}
}

// entryOf is the history entry of a model's run, and whether the run counts: a
// run that could not run proves nothing of the model, and a run of some cases,
// or one the budget cut short, is not D-221's run.
func entryOf(m *ModelReport, date string) (HistoryEntry, bool) {
	if m.CouldNotRun || m.VoidRuns > 0 {
		return HistoryEntry{}, false
	}
	e := HistoryEntry{Date: date, Verdict: verdictFail, Reliability: verdictFail, TransportFailureRate: m.Reliability.Rate, ConfigHash: m.ConfigHash}
	if m.Verdict.Pass {
		e.Verdict = verdictPass
	}
	if !m.Reliability.Failed {
		e.Reliability = verdictPass
	}
	return e, true
}

// StreakReport is a model@host's streak as the report holds it.
type StreakReport struct {
	Passes int `json:"passes"`
	Fails  int `json:"fails"`
	// Eligible is true after two consecutive passes (D-221).
	Eligible bool `json:"eligible"`
	// Member is true when the label is a member of the demo route.
	Member bool `json:"member"`
}

// RouteReport is what the run says of the demo route (D-221), read-only: the bench
// never edits the route, and a change of it is a reviewed change.
type RouteReport struct {
	// Members are the labels of the demo route's models, read from its config.
	Members []string `json:"members,omitempty"`
	// Eligible are the labels of this run with two consecutive passes, which may
	// join the route; Joinable is the part of them that is not a member yet.
	Eligible []string `json:"eligible,omitempty"`
	Joinable []string `json:"joinable,omitempty"`
	// Removable are the members with two consecutive failing runs, to take off
	// the route in a reviewed change.
	Removable []string `json:"removable,omitempty"`
	// Failing are the members whose run fails its verdict or its reliability: one
	// is enough to fail the weekly workflow.
	Failing []RouteFailure `json:"failing,omitempty"`
	// Recorded is true when the run was added to the history.
	Recorded bool `json:"recorded"`
}

// RouteFailure is a member of the demo route that failed its run, and why.
type RouteFailure struct {
	Model  string `json:"model"`
	Reason string `json:"reason"`
}

// applyHistory records the report's runs in h (when h is not nil and the run is
// a whole one), sets each model's streak and fills the report's route section
// from the members. It never changes the route.
func (r *Report) applyHistory(h *History, members []string, date string) {
	route := &RouteReport{Members: members}
	whole := !r.Subset && !r.Budget.Aborted
	for i := range r.Models {
		m := &r.Models[i]
		if h == nil {
			continue
		}
		if e, ok := entryOf(m, date); ok && whole {
			e.ScorerVersion = scorerOf(r)
			h.Add(m.Model, &e)
			route.Recorded = true
		}
		route.note(m, h.StreakOf(m.Model), slices.Contains(members, m.Model))
	}
	route.Failing = routeGate(r, members)
	r.Route = route
}

// note sets the model's streak and lists it as eligible, joinable or removable.
func (rr *RouteReport) note(m *ModelReport, s Streak, member bool) {
	m.Streak = &StreakReport{Passes: s.Passes, Fails: s.Fails, Eligible: s.Eligible(), Member: member}
	if s.Eligible() {
		rr.Eligible = append(rr.Eligible, m.Model)
		if !member {
			rr.Joinable = append(rr.Joinable, m.Model)
		}
	}
	if member && s.Failing() {
		rr.Removable = append(rr.Removable, m.Model)
	}
}

// routeGate is the weekly workflow's gate (D-221): the members of the demo
// route among the report's models whose run fails the verdict (D-220) or the
// reliability (D-207), each with why. One is enough to fail the workflow. A run
// of some cases only proves nothing of a verdict, so it is not gated.
func routeGate(r *Report, members []string) []RouteFailure {
	if r.Subset {
		return nil
	}
	var out []RouteFailure
	for i := range r.Models {
		m := &r.Models[i]
		if !slices.Contains(members, m.Model) {
			continue
		}
		var why []string
		if !m.Verdict.Pass {
			why = append(why, "verdict "+oneLine(verdictText(&m.Verdict)))
		}
		if m.Reliability.Failed {
			why = append(why, "reliability "+m.Reliability.text())
		}
		if len(why) > 0 {
			out = append(out, RouteFailure{Model: m.Model, Reason: strings.Join(why, "; ")})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Model < out[j].Model })
	return out
}

// routeText says what the run found of the demo route, for the console and the
// summary: one line for each thing worth the author's eye.
func routeText(rr *RouteReport) string {
	var b strings.Builder
	if len(rr.Members) > 0 {
		fmt.Fprintf(&b, "demo route members: %s\n", strings.Join(rr.Members, ", "))
	}
	for _, f := range rr.Failing {
		fmt.Fprintf(&b, "ROUTE MEMBER FAILED: %s: %s\n", f.Model, f.Reason)
	}
	if len(rr.Removable) > 0 {
		fmt.Fprintf(&b, "route members with %d consecutive failing runs (remove in a reviewed change): %s\n", consecutiveRuns, strings.Join(rr.Removable, ", "))
	}
	if len(rr.Joinable) > 0 {
		fmt.Fprintf(&b, "eligible to join the route (%d consecutive passes): %s\n", consecutiveRuns, strings.Join(rr.Joinable, ", "))
	}
	return b.String()
}
