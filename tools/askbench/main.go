// Command askbench runs the Ask box bench (D-217, D-218, D-219): a fixed set of
// questions in English and Polish over the demo's synthetic portfolio, each
// asked five times of a model through the Ask box's own loop and tools, graded
// by code against what the read API answers. It is a development tool, not
// part of the product; `task ask:bench` runs it:
//
//	askbench [-model ids] [-out dir] [-budget usd] [-max-calls n] [-judge] ...
//
// The model is reached the way the core's model-openai plugin reaches it, with
// the key in the environment variable OPENROUTER_API_KEY; the key goes only to
// the provider and is in no report, log or error. The bench exits 0 when every
// safety case (refusals and injections) passed on all five runs for every
// model, 1 when one did not, 2 when it could not run (a bad argument or key, or
// a model that no call reached: bad key, no credits, no network) and 3 when its
// budget ran out before every question was asked. A call that fails while
// others answer is an error listed in the report, not a model failure. Answer cases are reported, not gated.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	gevalopenai "github.com/datar-psa/goeval/openai"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/ui"
)

// keyEnv names the environment variable of the provider key: the same key the
// model router's OpenRouter provider holds.
const keyEnv = "OPENROUTER_API_KEY"

// The exit codes.
const (
	exitOK      = 0
	exitUnsafe  = 1
	exitError   = 2
	exitBudget  = 3
	httpTimeout = time.Minute
	// defaultBudget is the most the run may spend, in US dollars: far above
	// what the two candidates cost for the 28 cases five times each.
	defaultBudget   = "1.00"
	defaultMaxCalls = 2000
	defaultParallel = 4
	defaultOut      = "ask-bench-report"
	// hashShown is how many characters of a configuration hash the console shows.
	hashShown = 12
)

// ErrUsage reports arguments the bench can't use.
var ErrUsage = errors.New("usage")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	code := run(ctx, os.Args[1:], os.Getenv, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

// config is the arguments of a run, read.
type config struct {
	opts       Options
	out        string
	judgeModel string
}

// run runs the bench and returns its exit code.
func run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	cfg, err := parseArgs(args, getenv, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return exitOK
	}
	if err != nil {
		return failed(stderr, err)
	}
	b, err := NewBench(&cfg.opts)
	if err != nil {
		return failed(stderr, err)
	}
	report, err := b.Run(ctx)
	if err != nil {
		return failed(stderr, err)
	}
	if err := report.WriteTo(cfg.out); err != nil {
		return failed(stderr, err)
	}
	if _, err := io.WriteString(stdout, resultText(report, cfg.out)); err != nil {
		return exitError
	}
	switch {
	case report.CouldNotRun():
		return exitError
	case report.Budget.Aborted:
		return exitBudget
	case report.Failed():
		return exitUnsafe
	}
	return exitOK
}

// failed says why the bench could not run, on stderr, and returns its exit code.
func failed(stderr io.Writer, err error) int {
	if _, werr := io.WriteString(stderr, "askbench: "+err.Error()+"\n"); werr != nil {
		return exitError
	}
	return exitError
}

// resultText says in a few lines how the run went and where the report is.
func resultText(r *Report, dir string) string {
	var b strings.Builder
	for i := range r.Models {
		m := &r.Models[i]
		verdict := "every safety case passed on all runs"
		switch {
		case m.CouldNotRun:
			verdict = "COULD NOT RUN: no call reached the model, so nothing was scored"
		case !m.SafetyPassed:
			verdict = "A SAFETY CASE DID NOT PASS ON ALL RUNS"
		case m.ErroredRuns > 0:
			verdict = "no safety case failed, but it is not proven: some calls failed"
		}
		fmt.Fprintf(&b, "%s: %s; %s USD; configuration %s\n", m.Model, verdict, m.Cost, m.ConfigHash[:hashShown])
		if len(m.Errors) > 0 {
			fmt.Fprintf(&b, "%s: failed calls (errors, not model failures): %s\n", m.Model, errorsText(m.Errors))
		}
	}
	if r.Budget.Aborted {
		fmt.Fprintf(&b, "the budget ran out (%s of %s USD, %d of %d calls): the run was aborted\n", r.Budget.Cost, r.Budget.MaxCost, r.Budget.Calls, r.Budget.MaxCalls)
	}
	fmt.Fprintf(&b, "report: %s/report.json, %s/summary.md\n", dir, dir)
	return b.String()
}

// parseArgs reads the arguments and the key of a run.
func parseArgs(args []string, getenv func(string) string, stderr io.Writer) (*config, error) {
	fs := flag.NewFlagSet("askbench", flag.ContinueOnError)
	fs.SetOutput(stderr)
	models := fs.String("model", "", "comma-separated OpenRouter model ids; default D-207's two candidates, qwen/qwen3.7-flash and openai/gpt-oss-120b")
	prices := fs.String("price", "", "comma-separated id=input:output prices in USD per million tokens, for a model not among the candidates")
	baseURL := fs.String("base-url", defaultBaseURL, "the OpenAI-compatible API root")
	out := fs.String("out", defaultOut, "directory for report.json and summary.md")
	budget := fs.String("budget", defaultBudget, "the most to spend on the model, in USD; the run aborts cleanly past it")
	maxCalls := fs.Int("max-calls", defaultMaxCalls, "the most model calls to make; the run aborts cleanly past it")
	parallel := fs.Int("parallel", defaultParallel, "questions asked at once")
	cases := fs.String("case", "", "comma-separated ids of the cases to run; default all (a subset does not prove the safety cases)")
	judge := fs.Bool("judge", false, "also grade the advice cases with a judge model, reported only; costs calls beyond the Ask calls")
	judgeModel := fs.String("judge-model", "openai/gpt-oss-120b", "the judge's model id")
	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUsage, err)
	}
	key := getenv(keyEnv)
	if key == "" {
		return nil, fmt.Errorf("%w: %s is not set", ErrUsage, keyEnv)
	}
	list, err := modelsFrom(*models, *prices)
	if err != nil {
		return nil, err
	}
	limit, err := kernel.ParseDecimal(*budget)
	if err != nil || limit.Sign() <= 0 {
		return nil, fmt.Errorf("%w: -budget is a positive amount of USD", ErrUsage)
	}
	hc := &http.Client{Timeout: httpTimeout}
	cfg := &config{out: *out, judgeModel: *judgeModel, opts: Options{
		Models: list, BaseURL: *baseURL, APIKey: key, Parallel: *parallel, MaxCost: limit, MaxCalls: *maxCalls,
		Limits: ui.AskLimits{}, Cases: splitList(*cases), HTTP: hc, Now: time.Now,
	}}
	if *judge {
		cfg.opts.Judge = &adviceJudge{gen: gevalopenai.NewGenerator(*judgeModel,
			gevalopenai.WithBaseURL(*baseURL), gevalopenai.WithAPIKey(key), gevalopenai.WithHTTPClient(hc))}
	}
	return cfg, nil
}

// splitList splits a comma-separated list, dropping blanks.
func splitList(s string) []string {
	var out []string
	for part := range strings.SplitSeq(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// modelsFrom returns the models to ask: those named, or the candidates, each
// with the price the -price flag gives or the candidates' table.
func modelsFrom(names, prices string) ([]Model, error) {
	known := map[string]Model{}
	for _, m := range candidateModels() {
		known[m.ID] = m
	}
	for _, p := range splitList(prices) {
		m, err := parsePrice(p)
		if err != nil {
			return nil, err
		}
		known[m.ID] = m
	}
	ids := splitList(names)
	if len(ids) == 0 {
		for _, m := range candidateModels() {
			ids = append(ids, m.ID)
		}
	}
	out := make([]Model, 0, len(ids))
	for _, id := range ids {
		m, ok := known[id]
		if !ok {
			return nil, fmt.Errorf("%w: no price for %s: give one with -price %s=input:output", ErrModels, id, id)
		}
		out = append(out, m)
	}
	return out, nil
}

// parsePrice reads id=input:output, the prices in USD per million tokens.
func parsePrice(s string) (Model, error) {
	i := strings.LastIndex(s, "=")
	in, out, ok := strings.Cut(s[i+1:], ":")
	if i <= 0 || !ok {
		return Model{}, fmt.Errorf("%w: -price wants id=input:output, got %q", ErrUsage, s)
	}
	inPrice, err1 := kernel.ParseDecimal(in)
	outPrice, err2 := kernel.ParseDecimal(out)
	if err1 != nil || err2 != nil || inPrice.Sign() < 0 || outPrice.Sign() < 0 {
		return Model{}, fmt.Errorf("%w: -price %q: prices are non-negative decimals", ErrUsage, s)
	}
	return Model{ID: s[:i], InputPrice: inPrice, OutputPrice: outPrice}, nil
}
