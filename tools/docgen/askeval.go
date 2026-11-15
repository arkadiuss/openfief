package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var errAskEvalUsage = errors.New("usage: docgen askeval import -report FILE -run NAME -date YYYY-MM-DD [-out DIR] [-parallel N] | askeval table [-check]")

const (
	// runsPerCase is how many times the bench asks each case; a report that
	// lacks the number of questions is priced over cases times this.
	runsPerCase = 5
	perThousand = 1000
	// defaultParallel is how many questions the bench asks at once, D-208's cap
	// of model calls at once, and what a result file without a parallelism was
	// made at.
	defaultParallel = 3
	costPlaces      = 5
)

var (
	runNameRE = regexp.MustCompile(`^\d+(-[a-z0-9]+)*$`)
	dateRE    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// importArgs are the flags of `docgen askeval import`.
type importArgs struct {
	report, run, date, out string
	parallel               int
}

func parseImportArgs(args []string) (importArgs, error) {
	var a importArgs
	flags := flag.NewFlagSet("askeval import", flag.ContinueOnError)
	flags.StringVar(&a.report, "report", "", "the bench's report.json")
	flags.StringVar(&a.run, "run", "", "the run's name, a number and an optional label, such as 9-1 or 12-trace")
	flags.StringVar(&a.date, "date", "", "the run's UTC day, YYYY-MM-DD")
	flags.IntVar(&a.parallel, "parallel", defaultParallel, "questions the bench asked at once (its -parallel; the report does not carry it)")
	flags.StringVar(&a.out, "out", resultsDir, "folder of result files")
	if err := flags.Parse(args); err != nil {
		return a, fmt.Errorf("askeval import: %w", err)
	}
	if a.parallel < 1 || a.report == "" || !runNameRE.MatchString(a.run) || !dateRE.MatchString(a.date) {
		return a, errAskEvalUsage
	}
	return a, nil
}

// runAskEval handles `docgen askeval import`: it trims a bench's report.json
// to the committed result file of the run (internal/views/ui/evalresults).
func runAskEval(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "import" {
		return errAskEvalUsage
	}
	a, err := parseImportArgs(args[1:])
	if err != nil {
		return err
	}
	data, err := os.ReadFile(a.report)
	if err != nil {
		return fmt.Errorf("read report: %w", err)
	}
	results, err := trimReport(data, a.run, a.date)
	if err != nil {
		return err
	}
	stampParallel(results, a.parallel)
	body, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("encode results: %w", err)
	}
	name := filepath.Join(a.out, a.date+"-"+padRun(a.run)+".json")
	if err := os.WriteFile(name, append(body, '\n'), filePerm); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	if _, err := fmt.Fprintf(stdout, "askeval: wrote %d results to %s\n", len(results), name); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	return nil
}

// stampParallel records how many questions the run asked at once, unless it
// is D-208's 3, which a result file leaves out.
func stampParallel(results []evalResult, parallel int) {
	if parallel == defaultParallel {
		return
	}
	for i := range results {
		results[i].Parallel = parallel
	}
}

// padRun zero-pads the run's number so that file names sort in run order.
func padRun(run string) string {
	num, rest, _ := strings.Cut(run, "-")
	n, err := strconv.Atoi(num)
	if err != nil {
		return run
	}
	s := fmt.Sprintf("%02d", n)
	if rest != "" {
		s += "-" + rest
	}
	return s
}

// benchReport is the part of the bench's report.json the trim reads. Older
// reports lack the verdict, reliability and cost fields, which then stay empty.
type benchReport struct {
	Schema int `json:"schema"`
	// ScorerVersion is the version of the scoring rules that graded the report,
	// 0 for a report that predates versioning.
	ScorerVersion int          `json:"scorerVersion"`
	Subset        bool         `json:"subset"`
	Models        []benchModel `json:"models"`
}

type benchCase struct {
	ID      string `json:"id"`
	PassAll bool   `json:"passAll"`
}

type benchVerdict struct {
	Pass    bool          `json:"pass"`
	Failing []evalFailing `json:"failing"`
}

type benchModel struct {
	Model       string           `json:"model"`
	ModelID     string           `json:"modelId"`
	Route       []string         `json:"route"`
	Reasoning   string           `json:"reasoning"`
	ConfigHash  string           `json:"configHash"`
	Cost        string           `json:"cost"`
	InputTokens int              `json:"inputTokens"`
	CostPer1k   string           `json:"costPer1kQuestions"`
	CostSource  string           `json:"costSource"`
	Cases       []benchCase      `json:"cases"`
	Verdict     *benchVerdict    `json:"verdict"`
	Reliability *evalReliability `json:"reliability"`
	Completions *int             `json:"completions"`
	Errored     int              `json:"erroredRuns"`
	VoidRuns    int              `json:"voidRuns"`
	VoidCreds   int              `json:"voidCredentials"`
	CouldNot    bool             `json:"couldNotRun"`
	Questions   int              `json:"questions"`
}

// trimReport turns a report into the results of its models.
func trimReport(data []byte, run, date string) ([]evalResult, error) {
	var rep benchReport
	if err := json.Unmarshal(data, &rep); err != nil {
		return nil, fmt.Errorf("decode report: %w", err)
	}
	out := make([]evalResult, 0, len(rep.Models))
	for i := range rep.Models {
		r := trimModel(&rep.Models[i], rep.Subset)
		r.Run, r.Date, r.BenchSchema, r.ScorerVersion = run, date, rep.Schema, rep.ScorerVersion
		out = append(out, r)
	}
	return out, nil
}

// trimModel is one model@host of a report.
func trimModel(m *benchModel, subset bool) evalResult {
	r := evalResult{
		Model: m.ModelID, Host: strings.Join(m.Route, ","), Reasoning: m.Reasoning, Verdict: verdictNone,
		Cases: len(m.Cases), Unscored: m.Errored, Void: m.VoidRuns, VoidCredentials: m.VoidCreds, ConfigHash: m.ConfigHash, Reliability: m.Reliability,
		CostPer1k: m.CostPer1k, CostSource: m.CostSource, Answered: m.answered(),
	}
	if r.Model == "" {
		r.Model, _, _ = strings.Cut(m.Model, "@")
		r.Model = strings.TrimSuffix(r.Model, " (default routing)")
	}
	if subset {
		for _, c := range m.Cases {
			r.Diagnostic = append(r.Diagnostic, c.ID)
		}
	}
	if m.Verdict != nil && r.Answered {
		r.Verdict = verdictFail
		if m.Verdict.Pass {
			r.Verdict = verdictPass
		}
		r.Failing = m.Verdict.Failing
	} else {
		for _, c := range m.Cases {
			if !c.PassAll {
				r.UnderBar++
			}
		}
	}
	if r.CostPer1k == "" {
		est := m.estimatedCost()
		r.CostPer1k, r.CostSource = est.per1k, est.source
	}
	return r
}

// answered reports whether any call of the model got an answer.
func (m *benchModel) answered() bool {
	if m.CouldNot {
		return false
	}
	if m.Completions != nil {
		return *m.Completions > 0
	}
	return m.InputTokens > 0
}

// estimatedCost prices 1,000 questions from the run's total cost, for a
// report that predates the bench's own per-question figure.
func (m *benchModel) estimatedCost() costEstimate {
	total, err := strconv.ParseFloat(m.Cost, 64)
	if err != nil || total <= 0 {
		return costEstimate{}
	}
	questions := m.Questions
	if questions == 0 {
		questions = len(m.Cases) * runsPerCase
	}
	return costEstimate{per1k: strconv.FormatFloat(total/float64(questions)*perThousand, 'f', costPlaces, 64), source: "estimated"}
}

// costEstimate is USD per 1,000 questions as a decimal text, and where it
// comes from; both are empty when the report has no cost.
type costEstimate struct{ per1k, source string }
