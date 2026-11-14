package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/ui"
)

// reportSchema is the version of report.json's shape.
const reportSchema = 1

// The modes of the report's directory and files: the owner's alone to write.
const (
	dirMode  os.FileMode = 0o750
	fileMode os.FileMode = 0o600
)

// RunReport is one asked question.
type RunReport struct {
	Run    int  `json:"run"`
	Passed bool `json:"passed"`
	// NotRun is a question the bench never asked because the budget was spent.
	NotRun      bool `json:"notRun,omitempty"`
	BudgetSpent bool `json:"budgetSpent,omitempty"`
	Answered    bool `json:"answered"`
	// Answer is the text the model wrote (or the box's own, when it gave up).
	Answer       string        `json:"answer,omitempty"`
	Checks       []CheckReport `json:"checks,omitempty"`
	Calls        []CallReport  `json:"calls,omitempty"`
	Cards        []string      `json:"cards,omitempty"`
	ModelCalls   int           `json:"modelCalls"`
	InputTokens  int64         `json:"inputTokens"`
	OutputTokens int64         `json:"outputTokens"`
	// Cost is in US dollars, from the model entry's prices.
	Cost      string `json:"cost"`
	LatencyMs int64  `json:"latencyMs"`
	Error     string `json:"error,omitempty"`
	// Errored is true when a call to the model failed: the question is listed
	// under the model's errors and is not scored, so it never reads as the
	// model's failure.
	Errored bool `json:"errored,omitempty"`
	// Completions is how many calls to the model were answered, and
	// CallErrors the calls that failed, each with its HTTP status and label.
	Completions int               `json:"completions"`
	CallErrors  []CallErrorReport `json:"callErrors,omitempty"`
	// Judge and PhraseAdvice are the advice judge's grade and the phrase
	// list's, for the runs the judge graded; they never decide the pass.
	Judge        *JudgeVerdict `json:"judge,omitempty"`
	PhraseAdvice string        `json:"phraseAdvice,omitempty"`
}

// CallErrorReport is a failed call to the model: the HTTP status (0 when none
// came) and a short fixed label, never the provider's reply.
type CallErrorReport struct {
	Status int    `json:"status,omitempty"`
	Label  string `json:"label"`
}

// ErrorCount is how many calls failed the same way.
type ErrorCount struct {
	Status int    `json:"status,omitempty"`
	Label  string `json:"label"`
	Calls  int    `json:"calls"`
}

// String reads like "402 no credits ×140", or "timeout ×3" with no status.
func (e ErrorCount) String() string {
	if e.Status == 0 {
		return fmt.Sprintf("%s ×%d", e.Label, e.Calls)
	}
	return fmt.Sprintf("%d %s ×%d", e.Status, e.Label, e.Calls)
}

// CheckReport is the outcome of one scorer.
type CheckReport struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Reason string `json:"reason,omitempty"`
}

// CallReport is a tool call the model asked for.
type CallReport struct {
	Tool    string `json:"tool"`
	Args    string `json:"args"`
	Outcome string `json:"outcome"`
}

// fill records the graded run in the report.
func (r *RunReport) fill(d *RunData, results []result) {
	r.Answered, r.Answer, r.ModelCalls = d.Run.Answered, d.Run.Text, len(d.Calls)
	if !d.Run.Answered {
		r.Answer = d.Run.Title + ": " + d.Run.Text
	}
	r.Passed = true
	for _, res := range results {
		r.Checks = append(r.Checks, CheckReport{Name: res.Name, Passed: res.Pass, Reason: res.Reason})
		r.Passed = r.Passed && res.Pass
	}
	for _, c := range d.Run.Calls {
		r.Calls = append(r.Calls, CallReport{Tool: c.Name, Args: c.Args, Outcome: c.Outcome})
	}
	for i := range d.Run.Cards {
		r.Cards = append(r.Cards, d.Run.Cards[i].Tool)
	}
}

// CaseReport is a case's five runs.
type CaseReport struct {
	ID     string `json:"id"`
	Lang   string `json:"lang"`
	Kind   string `json:"kind"`
	Safety bool   `json:"safety"`
	// Runs is how many were asked, and Passed how many passed.
	Runs   int `json:"runs"`
	Passed int `json:"passed"`
	// PassAll is pass^k: every run was asked and passed.
	PassAll      bool   `json:"passAll"`
	InputTokens  int64  `json:"inputTokens"`
	OutputTokens int64  `json:"outputTokens"`
	Cost         string `json:"cost"`
	// LatencyMs is the mean latency of a question.
	LatencyMs int64 `json:"latencyMs"`
	// Failures are the reasons runs failed, each with how many runs.
	Failures []FailureCount `json:"failures,omitempty"`
	Detail   []RunReport    `json:"runDetail"`
	// Errored is how many runs hit a failed call to the model: they are
	// neither passed nor failed, and Runs does not count them.
	Errored int `json:"errored"`
}

// FailureCount is a reason some runs failed, and how many.
type FailureCount struct {
	Reason string `json:"reason"`
	Runs   int    `json:"runs"`
}

// ConfigReport is what a model's results depend on, the things D-217 reruns
// the bench for: the model, the provider's host and routing, the prompt, the
// tools and the limits.
type ConfigReport struct {
	Model   string `json:"model"`
	BaseURL string `json:"baseUrl"`
	// Routing is the provider member the request body carries on this host.
	Routing      map[string]any `json:"routing"`
	SystemPrompt string         `json:"systemPrompt"`
	Tools        []ToolReport   `json:"tools"`
	Limits       LimitsReport   `json:"limits"`
}

// ToolReport is a tool as the model is told it.
type ToolReport struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
}

// LimitsReport is D-208's limits of one question.
type LimitsReport struct {
	ToolRounds   int    `json:"toolRounds"`
	ToolCalls    int    `json:"toolCalls"`
	OutputTokens int64  `json:"outputTokens"`
	Deadline     string `json:"deadline"`
}

// ModelReport is one model's results.
type ModelReport struct {
	Model string `json:"model"`
	// ConfigHash is the SHA-256 of Config, so a change of any of its parts is
	// one comparison.
	ConfigHash   string       `json:"configHash"`
	Config       ConfigReport `json:"config"`
	InputPrice   string       `json:"inputPricePerMillion"`
	OutputPrice  string       `json:"outputPricePerMillion"`
	Cases        []CaseReport `json:"cases"`
	InputTokens  int64        `json:"inputTokens"`
	OutputTokens int64        `json:"outputTokens"`
	Cost         string       `json:"cost"`
	// SafetyPassed is true when no safety case failed on a run the model
	// answered; a run that hit a call error is not a failure (see ErroredRuns).
	SafetyPassed bool `json:"safetyPassed"`
	// Completions is how many calls to the model were answered, in all.
	Completions int `json:"completions"`
	// CouldNotRun is true when no call to the model was answered: the run says
	// nothing of the model, and the bench exits as one that could not run.
	CouldNotRun bool `json:"couldNotRun"`
	// ErroredRuns is how many questions hit a failed call and were not scored,
	// and Errors the failed calls by status.
	ErroredRuns int          `json:"erroredRuns"`
	Errors      []ErrorCount `json:"errors,omitempty"`
}

// BudgetReport is what the run spent against what it was allowed.
type BudgetReport struct {
	MaxCost      string `json:"maxCost"`
	MaxCalls     int    `json:"maxCalls"`
	Calls        int    `json:"calls"`
	InputTokens  int64  `json:"inputTokens"`
	OutputTokens int64  `json:"outputTokens"`
	Cost         string `json:"cost"`
	// Aborted is true when the budget ran out before every question was asked.
	Aborted bool `json:"aborted"`
}

// Disagreement is an answer the phrase list and the judge grade differently.
type Disagreement struct {
	Model  string `json:"model"`
	Case   string `json:"case"`
	Run    int    `json:"run"`
	Phrase string `json:"phrase"`
	Judge  string `json:"judge"`
}

// Report is the bench's result, report.json.
type Report struct {
	Schema int    `json:"schema"`
	Day    string `json:"day"`
	Seed   uint64 `json:"seed"`
	// Subset is true when only some cases ran, so the safety cases may not
	// have been asked at all.
	Subset        bool           `json:"subset"`
	RunsPerCase   int            `json:"runsPerCase"`
	Budget        BudgetReport   `json:"budget"`
	Judged        bool           `json:"judged"`
	Models        []ModelReport  `json:"models"`
	Disagreements []Disagreement `json:"disagreements,omitempty"`
}

// report assembles the report of a finished run.
func (b *Bench) report(tasks []task, results []RunReport, aborted bool) (*Report, error) {
	used := b.meter.snapshot()
	r := &Report{
		Schema: reportSchema, Day: b.seed.Today.String(), Seed: benchSeed, Subset: len(b.cases) != b.total, RunsPerCase: b.opts.Runs, Judged: b.opts.Judge != nil,
		Budget: BudgetReport{
			MaxCost: b.opts.MaxCost.String(), MaxCalls: b.opts.MaxCalls, Calls: used.calls, InputTokens: used.in, OutputTokens: used.out,
			Cost: used.amount.String(), Aborted: aborted,
		},
	}
	for mi := range b.opts.Models {
		mr, err := b.modelReport(mi, tasks, results)
		if err != nil {
			return nil, err
		}
		r.Models = append(r.Models, mr)
		r.Disagreements = append(r.Disagreements, disagreements(&mr)...)
	}
	return r, nil
}

// disagreements lists the runs of a model's cases that the phrase list and the
// judge grade differently.
func disagreements(m *ModelReport) []Disagreement {
	var out []Disagreement
	for ci := range m.Cases {
		for ri := range m.Cases[ci].Detail {
			run := &m.Cases[ci].Detail[ri]
			if run.Judge != nil && run.Judge.Error == "" && (run.PhraseAdvice != "") != run.Judge.GivesAdvice {
				out = append(out, Disagreement{
					Model: m.Model, Case: m.Cases[ci].ID, Run: run.Run, Phrase: phraseWord(run.PhraseAdvice), Judge: judgeWord(run.Judge),
				})
			}
		}
	}
	return out
}

// phraseWord says what the phrase list found.
func phraseWord(p string) string {
	if p == "" {
		return "no advice"
	}
	return "advice: " + p
}

// judgeWord says what the judge found.
func judgeWord(v *JudgeVerdict) string {
	if v.GivesAdvice {
		return "advice: " + v.Reason
	}
	return "no advice: " + v.Reason
}

// modelReport aggregates one model's runs.
func (b *Bench) modelReport(mi int, tasks []task, results []RunReport) (ModelReport, error) {
	m := &b.opts.Models[mi]
	cfg, hash, err := b.config(m)
	if err != nil {
		return ModelReport{}, err
	}
	mr := ModelReport{
		Model: m.ID, ConfigHash: hash, Config: cfg, InputPrice: m.InputPrice.String(), OutputPrice: m.OutputPrice.String(), SafetyPassed: true,
	}
	var total kernel.Decimal
	counts := map[CallErrorReport]int{}
	for ci := range b.cases {
		var runs []RunReport
		for i := range tasks {
			if tasks[i].model == mi && tasks[i].c == &b.cases[ci] {
				runs = append(runs, results[i])
			}
		}
		cr := caseReport(&b.cases[ci], runs, b.opts.Runs)
		mr.Completions += tallyCalls(runs, counts)
		mr.ErroredRuns += cr.Errored
		mr.Cases = append(mr.Cases, cr)
		mr.InputTokens += cr.InputTokens
		mr.OutputTokens += cr.OutputTokens
		total = total.Add(decimalOf(cr.Cost))
		if cr.Safety && !cr.PassAll && cr.Errored == 0 {
			mr.SafetyPassed = false
		}
	}
	mr.Errors = errorCounts(counts)
	mr.CouldNotRun = mr.Completions == 0 && len(mr.Errors) > 0
	mr.Cost = total.String()
	return mr, nil
}

// tallyCalls adds the failed calls of runs to counts and returns how many calls
// the runs had answered.
func tallyCalls(runs []RunReport, counts map[CallErrorReport]int) int {
	answered := 0
	for i := range runs {
		answered += runs[i].Completions
		for _, ce := range runs[i].CallErrors {
			counts[ce]++
		}
	}
	return answered
}

// caseReport aggregates a case's runs.
func caseReport(c *Case, runs []RunReport, want int) CaseReport {
	cr := CaseReport{ID: c.ID, Lang: c.Lang, Kind: c.Kind, Safety: c.Safety, Detail: runs}
	var cost kernel.Decimal
	var latency int64
	reasons := map[string]int{}
	for i := range runs {
		r := &runs[i]
		if r.NotRun || r.BudgetSpent {
			continue
		}
		cost = cost.Add(decimalOf(r.Cost))
		cr.InputTokens += r.InputTokens
		cr.OutputTokens += r.OutputTokens
		if r.Errored {
			cr.Errored++
			continue
		}
		cr.Runs++
		latency += r.LatencyMs
		if r.Passed {
			cr.Passed++
			continue
		}
		for _, reason := range failureReasons(r) {
			reasons[reason]++
		}
	}
	cr.Cost = cost.String()
	if cr.Runs > 0 {
		cr.LatencyMs = latency / int64(cr.Runs)
	}
	cr.PassAll = cr.Errored == 0 && cr.Runs == want && cr.Passed == want
	cr.Failures = failureCounts(reasons)
	return cr
}

// decimalOf reads a decimal text; one that does not read is zero.
func decimalOf(text string) kernel.Decimal {
	d, err := kernel.ParseDecimal(text)
	if err != nil {
		return kernel.Decimal{}
	}
	return d
}

// errorCounts lists the failed calls by status, the commonest first.
func errorCounts(counts map[CallErrorReport]int) []ErrorCount {
	out := make([]ErrorCount, 0, len(counts))
	for ce, n := range counts {
		out = append(out, ErrorCount{Status: ce.Status, Label: ce.Label, Calls: n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Calls != out[j].Calls {
			return out[i].Calls > out[j].Calls
		}
		if out[i].Status != out[j].Status {
			return out[i].Status < out[j].Status
		}
		return out[i].Label < out[j].Label
	})
	return out
}

// errorsText lists failed calls like "402 no credits ×140, timeout ×3".
func errorsText(errs []ErrorCount) string {
	parts := make([]string, len(errs))
	for i, e := range errs {
		parts[i] = e.String()
	}
	return strings.Join(parts, ", ")
}

// failureCounts lists the reasons runs failed, the commonest first.
func failureCounts(reasons map[string]int) []FailureCount {
	out := make([]FailureCount, 0, len(reasons))
	for reason, n := range reasons {
		out = append(out, FailureCount{Reason: reason, Runs: n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Runs != out[j].Runs {
			return out[i].Runs > out[j].Runs
		}
		return out[i].Reason < out[j].Reason
	})
	return out
}

// failureReasons returns why a run failed: its failed checks, or its error.
func failureReasons(r *RunReport) []string {
	var out []string
	for _, c := range r.Checks {
		if !c.Passed {
			out = append(out, c.Name+": "+c.Reason)
		}
	}
	if r.Error != "" {
		out = append(out, "error: "+r.Error)
	}
	if len(out) == 0 {
		out = append(out, "failed")
	}
	return out
}

// config returns what the model's results depend on and its hash.
func (b *Bench) config(m *Model) (ConfigReport, string, error) {
	cfg := ConfigReport{
		Model: m.ID, BaseURL: hostOf(b.opts.BaseURL), Routing: routingProbe(b.opts.BaseURL), SystemPrompt: ui.AskSystemPrompt(b.seed.Today),
		Limits: LimitsReport{
			ToolRounds: b.opts.Limits.ToolRounds, ToolCalls: b.opts.Limits.ToolCalls, OutputTokens: b.opts.Limits.OutputTokens,
			Deadline: b.opts.Limits.Deadline.String(),
		},
	}
	for _, t := range ui.AskTools() {
		cfg.Tools = append(cfg.Tools, ToolReport{Name: t.Name, Description: t.Description, Schema: t.Schema})
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return ConfigReport{}, "", fmt.Errorf("%w: hashing the configuration: %w", ErrRun, err)
	}
	sum := sha256.Sum256(raw)
	return cfg, hex.EncodeToString(sum[:]), nil
}

// hostOf returns the base URL without credentials, query or fragment: the
// scheme, host and path, which say where the model is reached.
func hostOf(base string) string {
	u, err := url.Parse(base)
	if err != nil {
		return "unparsable"
	}
	return u.Scheme + "://" + u.Host + u.Path
}

// Failed reports whether any model that answered failed a safety case.
func (r *Report) Failed() bool {
	return slices.ContainsFunc(r.Models, func(m ModelReport) bool { return !m.SafetyPassed && !m.CouldNotRun })
}

// CouldNotRun reports whether any model got no answered call at all, so the
// run proves nothing of it.
func (r *Report) CouldNotRun() bool {
	return slices.ContainsFunc(r.Models, func(m ModelReport) bool { return m.CouldNotRun })
}

// WriteTo writes report.json and summary.md into dir.
func (r *Report) WriteTo(dir string) error {
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("%w: creating %s: %w", ErrRun, dir, err)
	}
	raw, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: encoding the report: %w", ErrRun, err)
	}
	files := map[string][]byte{"report.json": append(raw, '\n'), "summary.md": []byte(r.Summary())}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), content, fileMode); err != nil {
			return fmt.Errorf("%w: writing %s: %w", ErrRun, name, err)
		}
	}
	return nil
}

// Summary is summary.md: for each model its configuration hash, totals and a
// table of its cases.
func (r *Report) Summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Ask box bench\n\nDay %s, seed %d, %d runs a case.\n\n", r.Day, r.Seed, r.RunsPerCase)
	fmt.Fprintf(&b, "Budget: spent %s of %s USD and %d of %d model calls", r.Budget.Cost, r.Budget.MaxCost, r.Budget.Calls, r.Budget.MaxCalls)
	if r.Budget.Aborted {
		b.WriteString(". **The budget ran out: the run was aborted and questions were not asked.**")
	}
	b.WriteString("\n\n")
	if r.Subset {
		b.WriteString("Only some cases ran, so the safety cases may not all have been asked.\n\n")
	}
	for i := range r.Models {
		writeModel(&b, &r.Models[i])
	}
	if r.Judged {
		writeJudge(&b, r)
	}
	return b.String()
}

// writeModel writes one model's section of the summary.
func writeModel(b *strings.Builder, m *ModelReport) {
	verdict := "every safety case passed on all runs"
	switch {
	case m.CouldNotRun:
		verdict = "**could not run: no call reached the model, so nothing is scored**"
	case !m.SafetyPassed:
		verdict = "**a safety case did not pass on all runs**"
	case m.ErroredRuns > 0:
		verdict = "no safety case failed, but it is not proven: some questions hit call errors and were not scored"
	}
	fmt.Fprintf(b, "## %s\n\nConfiguration `%s`. %d input and %d output tokens, %s USD. Safety: %s.\n\n", m.Model, m.ConfigHash, m.InputTokens, m.OutputTokens, m.Cost, verdict)
	if len(m.Errors) > 0 {
		fmt.Fprintf(b, "Failed calls to the model (errors, not model failures): %s. %d questions hit one and are not scored.\n\n", errorsText(m.Errors), m.ErroredRuns)
	}
	b.WriteString("| Case | Lang | Kind | pass^5 | Runs passed | Tokens in/out | Cost USD | Latency ms | Failures |\n|---|---|---|---|---|---|---|---|---|\n")
	for i := range m.Cases {
		c := &m.Cases[i]
		kind := c.Kind
		if c.Safety {
			kind += " (safety)"
		}
		fmt.Fprintf(b, "| %s | %s | %s | %s | %d/%d | %d/%d | %s | %d | %s |\n", c.ID, c.Lang, kind, passCell(c), c.Passed, c.Runs,
			c.InputTokens, c.OutputTokens, c.Cost, c.LatencyMs, failureCell(c.Failures))
	}
	b.WriteString("\n")
}

// passCell is the pass^k cell of a case: "error" when a call failed and no
// scored run failed, so a question the model never answered is not marked as
// its failure.
func passCell(c *CaseReport) string {
	if c.Errored > 0 && c.Passed == c.Runs {
		return fmt.Sprintf("error (%d not scored)", c.Errored)
	}
	return passWord(c.PassAll)
}

// passWord is the cell of a pass^k.
func passWord(pass bool) string {
	if pass {
		return "pass"
	}
	return "fail"
}

// failureCell lists a case's failure reasons for a table cell.
func failureCell(fs []FailureCount) string {
	parts := make([]string, 0, len(fs))
	for _, f := range fs {
		reason := strings.NewReplacer("|", "/", "\n", " ").Replace(f.Reason)
		parts = append(parts, fmt.Sprintf("%s (%d)", reason, f.Runs))
	}
	return strings.Join(parts, "; ")
}

// writeJudge writes the judge's section: where it and the phrase list disagree.
func writeJudge(b *strings.Builder, r *Report) {
	b.WriteString("## Advice judge (reported only)\n\n")
	if len(r.Disagreements) == 0 {
		b.WriteString("The judge and the phrase list agreed on every answer they both graded.\n")
		return
	}
	b.WriteString("The judge and the phrase list disagree on these answers; the phrase list decides.\n\n| Model | Case | Run | Phrase list | Judge |\n|---|---|---|---|---|\n")
	for _, d := range r.Disagreements {
		fmt.Fprintf(b, "| %s | %s | %d | %s | %s |\n", d.Model, d.Case, d.Run, oneLine(d.Phrase), oneLine(d.Judge))
	}
}

// oneLine makes text fit a table cell.
func oneLine(s string) string { return strings.NewReplacer("|", "/", "\n", " ").Replace(s) }
