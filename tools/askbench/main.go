// Command askbench runs the Ask box bench (D-217, D-218, D-219): a fixed set of
// questions in English and Polish over the demo's synthetic portfolio, each
// asked five times of a model through the Ask box's own loop and tools, graded
// by code against what the read API answers. It is a development tool, not
// part of the product; `task ask:bench` runs it:
//
//	askbench [-model ids] [-out dir] [-budget usd] [-max-calls n] [-judge] ...
//	askbench -rescore report.json -out dir
//
// The second form grades the stored answers of a report again by the scoring
// rules of this build (ScorerVersion), calls no model and needs no key.
//
// The model is reached the way the core's model-openai plugin reaches it, with
// the key in the environment variable OPENROUTER_API_KEY; the key goes only to
// the provider and is in no report, log or error. The bench exits 0 when every
// safety property (the refusals, and the injection cases' safety: no payload
// followed, no extra call, no echo) passed on all five runs for every model, 1
// when one did not, 2 when it could not run (a bad argument or key, or
// a model that no call reached: no network), 3 when its budget ran out before
// every question was asked, 4 when a member of the demo route (-demo-route,
// D-221) failed its verdict or reliability and 5 when a refused key (401) or no
// credits (402) voided a run, which says nothing of any model and comes before
// the other codes. A call that fails while others answer is an error listed in
// the report, not a model failure. Answer cases are reported, not gated.
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
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	gevalopenai "github.com/datar-psa/goeval/openai"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/ui"
	"github.com/arkadiuss/openfief/plugins/model-openai/openaichat"
)

// keyEnv names the environment variable of the provider key: the same key the
// model router's OpenRouter provider holds.
const keyEnv = "OPENROUTER_API_KEY"

// The exit codes.
const (
	exitOK     = 0
	exitUnsafe = 1
	exitError  = 2
	exitBudget = 3
	// exitRoute is the exit code of a member of the demo route whose run failed
	// its verdict or its reliability, when no safety case failed (D-221).
	exitRoute = 4
	// exitCredentials is the exit code of a run that a refused key (401) or no
	// credits (402) voided: it says nothing of any model or host, so the weekly
	// workflow fails and the author sees it (D-221). It comes before every other
	// code, since the results it would read are not the run's.
	exitCredentials = 5
	httpTimeout     = time.Minute
	// defaultBudget is the most the run may spend, in US dollars: far above
	// what the two candidates cost for the 44 cases five times each.
	defaultBudget   = "1.00"
	defaultMaxCalls = 2000
	defaultParallel = 3 // D-208 caps the box at 3 model calls at once, and one question makes one call at a time
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
	// history is the path of the history file the run reads and rewrites, "" for
	// none; members are the labels of the demo route's models, nil for none.
	history string
	members []string
	// rescore is the report.json to grade again, "" for a run of the bench.
	rescore string
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
	if cfg.rescore != "" {
		return runRescore(ctx, cfg, stdout, stderr)
	}
	b, err := NewBench(&cfg.opts)
	if err != nil {
		return failed(stderr, err)
	}
	report, err := b.Run(ctx)
	if err != nil {
		return failed(stderr, err)
	}
	if err := recordRun(report, cfg, cfg.opts.Now()); err != nil {
		return failed(stderr, err)
	}
	if err := report.WriteTo(cfg.out); err != nil {
		return failed(stderr, err)
	}
	if _, err := io.WriteString(stdout, resultText(report, cfg.out)); err != nil {
		return exitError
	}
	return exitCode(report)
}

// runRescore grades the report again with this build's scoring rules, makes no
// model call, writes the new report.json and summary.md into the output
// directory and says what changed. Its exit code follows the report like a run's.
func runRescore(ctx context.Context, cfg *config, stdout, stderr io.Writer) int {
	report, err := LoadReport(cfg.rescore)
	if err != nil {
		return failed(stderr, err)
	}
	if err := Rescore(ctx, report); err != nil {
		return failed(stderr, err)
	}
	if err := report.WriteTo(cfg.out); err != nil {
		return failed(stderr, err)
	}
	if _, err := io.WriteString(stdout, rescoreText(report, cfg.out)); err != nil {
		return exitError
	}
	return exitCode(report)
}

// rescoreText says in a few lines what the re-grading did and where its report is.
func rescoreText(r *Report, dir string) string {
	var b strings.Builder
	info := r.Rescore
	fmt.Fprintf(&b, "re-scored %d runs from scorer version %d to %d: %d passed or failed differently; %d had their cards rebuilt from their calls, %d checks were carried over, %d runs were kept as graded\n",
		info.Runs, info.From, r.ScorerVersion, info.Changed, info.RebuiltCards, info.CarriedChecks, info.Kept)
	for _, note := range info.Notes {
		fmt.Fprintf(&b, "kept: %s\n", note)
	}
	if info.DerivedFaults > 0 || info.Undecided > 0 {
		fmt.Fprintf(&b, "faults: %d failed calls had none stored and got theirs from their status and label; %d 4xx could not be decided without the request's hash and count in no reliability\n", info.DerivedFaults, info.Undecided)
	}
	for i := range r.Models {
		m := &r.Models[i]
		fmt.Fprintf(&b, "%s: %s; reliability %s\n", m.Model, verdictText(&m.Verdict), m.Reliability.text())
		b.WriteString(faultLines(m))
	}
	fmt.Fprintf(&b, "report: %s/report.json, %s/summary.md\n", dir, dir)
	return b.String()
}

// exitCode is the bench's exit code for a finished report: a run voided by its
// key or credits first, then a model no call reached or a budget that ran out, then a safety property that failed,
// then a member of the demo route that failed its verdict or reliability (D-221,
// which fails the weekly workflow so the author sees it).
func exitCode(report *Report) int {
	switch {
	case report.CredentialsVoid():
		return exitCredentials
	case report.CouldNotRun():
		return exitError
	case report.Budget.Aborted:
		return exitBudget
	case report.Failed():
		return exitUnsafe
	case report.Route != nil && len(report.Route.Failing) > 0:
		return exitRoute
	}
	return exitOK
}

// recordRun applies the demo route and the history to the report: it reads the
// history when -history names one, adds the run to it, rewrites the file, and
// fills the report's route section. A run given neither has none.
func recordRun(r *Report, cfg *config, now time.Time) error {
	if cfg.history == "" && len(cfg.members) == 0 {
		return nil
	}
	var h *History
	if cfg.history != "" {
		var err error
		if h, err = LoadHistory(cfg.history); err != nil {
			return err
		}
	}
	r.applyHistory(h, cfg.members, now.UTC().Format(time.DateOnly))
	if h != nil {
		return h.Save(cfg.history)
	}
	return nil
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
		fmt.Fprintf(&b, "%s: %s; %s USD; configuration %s; scorer version %d\n", m.Model, verdict, m.Cost, m.ConfigHash[:hashShown], r.ScorerVersion)
		if len(m.Errors) > 0 {
			fmt.Fprintf(&b, "%s: failed calls (errors, not model failures): %s\n", m.Model, errorsText(m.Errors))
			if n := badReplies(m.Errors); n > 0 {
				fmt.Fprintf(&b, "%s: bad replies (a 2xx answer that is no chat completion; the host's fault, in the reliability): %d\n", m.Model, n)
			}
			b.WriteString(faultLines(m))
		}
	}
	if r.Budget.Aborted {
		fmt.Fprintf(&b, "the budget ran out (%s of %s USD, %d of %d calls): the run was aborted\n", r.Budget.Cost, r.Budget.MaxCost, r.Budget.Calls, r.Budget.MaxCalls)
	}
	if r.Route != nil {
		b.WriteString(routeText(r.Route))
	}
	fmt.Fprintf(&b, "report: %s/report.json, %s/summary.md\n", dir, dir)
	return b.String()
}

// faultLines says on stdout what a model's failed calls mean beyond their
// counts: a 404 loudly, the 4xx the evidence decided, and the void runs.
func faultLines(m *ModelReport) string {
	var b strings.Builder
	if n := notFoundCalls(m.Errors); n > 0 {
		fmt.Fprintf(&b, "%s: 404: the pin may be stale (%d calls found no route; the host's fault, in the reliability)\n", m.Model, n)
	}
	if n := callsByRule(m.Errors, ruleAcceptedElsewhere); n > 0 {
		fmt.Fprintf(&b, "%s: %d rejected calls were accepted elsewhere: the host's fault, in the reliability\n", m.Model, n)
	}
	if n := callsByRule(m.Errors, ruleAcceptedNowhere); n > 0 {
		fmt.Fprintf(&b, "%s: %d rejected calls were accepted nowhere: the bench's fault, the request check should catch them (requestcheck.go)\n", m.Model, n)
	}
	if m.VoidCredentials > 0 {
		fmt.Fprintf(&b, "%s: void (credentials/credit): %d runs, the key was refused or the credits are gone; exit %d\n", m.Model, m.VoidCredentials, exitCredentials)
	}
	if bench := m.VoidRuns - m.VoidCredentials; bench > 0 {
		fmt.Fprintf(&b, "%s: void (bench fault): %d runs, to be run again after the fix\n", m.Model, bench)
	}
	return b.String()
}

// parseArgs reads the arguments and the key of a run.
func parseArgs(args []string, getenv func(string) string, stderr io.Writer) (*config, error) {
	fs := flag.NewFlagSet("askbench", flag.ContinueOnError)
	fs.SetOutput(stderr)
	models := fs.String("model", "", "comma-separated OpenRouter model ids; default D-207's two candidates, qwen/qwen3.7-flash and openai/gpt-oss-120b")
	prices := fs.String("price", "", "comma-separated id=input:output prices in USD per million tokens, for a model not among the candidates")
	routes := fs.String("route", "", "comma-separated model=provider1|provider2 entries: pin the model to those OpenRouter hosts (their provider slugs), in that order, with no fallback; one model may have several entries, each a run of its own labelled model@provider1|provider2; a model also listed in -model keeps its default-routing run beside them")
	reasoning := fs.String("reasoning", "", "comma-separated model=off|low|medium|high entries (the model's id or its label): OpenRouter's reasoning control for that run, {\"enabled\": false} for off and {\"effort\": ...} for the rest; nothing is sent for a model without an entry; OpenRouter base URLs only")
	structured := fs.String("structured", "", "comma-separated model ids whose entry has the structured-output flag, asked for the answer strictly (gpt-oss and gemma-4 are by default)")
	jsonMode := fs.String("json-mode", "", "comma-separated model ids whose entry has no structured-output flag, asked for the answer as a JSON object, whatever the default")
	baseURL := fs.String("base-url", defaultBaseURL, "the OpenAI-compatible API root")
	out := fs.String("out", defaultOut, "directory for report.json and summary.md")
	budget := fs.String("budget", defaultBudget, "the most to spend on the model, in USD; the run aborts cleanly past it")
	maxCalls := fs.Int("max-calls", defaultMaxCalls, "the most model calls to make; the run aborts cleanly past it")
	parallel := fs.Int("parallel", defaultParallel, "questions asked at once (default 3, D-208's cap of model calls at once: more would test a load the box refuses)")
	cases := fs.String("case", "", "comma-separated ids of the cases to run; default all (a subset does not prove the safety cases)")
	historyPath := fs.String("history", "", "the path of the history file (tools/askbench/history.json): the run's verdict and reliability are added to it and the file is rewritten, and each model@host's consecutive passes and fails (D-221) are reported; never written without this flag")
	demoRoute := fs.String("demo-route", "", "the demo route's config directory (deploy/demo/model), read only: its models are asked when -model and -route name none, and a member whose run fails its verdict or reliability makes the bench exit 4; the route is never edited")
	judge := fs.Bool("judge", false, "also grade the advice cases with a judge model, reported only; costs calls beyond the Ask calls")
	judgeModel := fs.String("judge-model", "openai/gpt-oss-120b", "the judge's model id")
	rescore := fs.String("rescore", "", "a report.json to grade again with the current scoring rules (ScorerVersion), making no model call and needing no key; the new report.json and summary.md go into -out, which is required and must not hold the report")
	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUsage, err)
	}
	if *rescore != "" {
		return rescoreArgs(fs, *rescore, *out)
	}
	key := getenv(keyEnv)
	if key == "" {
		return nil, fmt.Errorf("%w: %s is not set", ErrUsage, keyEnv)
	}
	chosen, err := chooseModels(&modelFlags{names: *models, prices: *prices, routes: *routes, reasoning: *reasoning, structured: splitList(*structured), jsonMode: splitList(*jsonMode)}, *demoRoute)
	if err != nil {
		return nil, err
	}
	if *reasoning != "" && !newChatClient(nil, *baseURL, "").OnOpenRouter() {
		return nil, fmt.Errorf("%w: -reasoning is OpenRouter's request parameter and -base-url is not OpenRouter's", ErrUsage)
	}
	limit, err := kernel.ParseDecimal(*budget)
	if err != nil || limit.Sign() <= 0 {
		return nil, fmt.Errorf("%w: -budget is a positive amount of USD", ErrUsage)
	}
	hc := &http.Client{Timeout: httpTimeout}
	cfg := &config{out: *out, judgeModel: *judgeModel, history: *historyPath, members: memberLabels(chosen.members), opts: Options{
		Models: chosen.list, BaseURL: *baseURL, APIKey: key, Parallel: *parallel, MaxCost: limit, MaxCalls: *maxCalls,
		Limits: ui.AskLimits{}, Cases: splitList(*cases), HTTP: hc, Now: time.Now,
	}}
	if *judge {
		cfg.opts.Judge = &adviceJudge{gen: gevalopenai.NewGenerator(*judgeModel,
			gevalopenai.WithBaseURL(*baseURL), gevalopenai.WithAPIKey(key), gevalopenai.WithHTTPClient(hc))}
	}
	return cfg, nil
}

// rescoreArgs reads the arguments of a re-grading: the report and an explicit
// output directory other than the report's own, and no flag that only a run of
// the bench uses.
func rescoreArgs(fs *flag.FlagSet, report, out string) (*config, error) {
	given := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { given[f.Name] = true })
	for _, name := range []string{"model", "route", "reasoning", "structured", "json-mode", "price", "history", "demo-route", "judge", "budget", "max-calls", "case", "base-url"} {
		if given[name] {
			return nil, fmt.Errorf("%w: -%s is a flag of a run, and -rescore makes none", ErrUsage, name)
		}
	}
	if !given["out"] {
		return nil, fmt.Errorf("%w: -rescore needs -out, the directory for the new report", ErrUsage)
	}
	if filepath.Clean(filepath.Dir(report)) == filepath.Clean(out) {
		return nil, fmt.Errorf("%w: -out is the directory of the report to re-score; it would overwrite it", ErrUsage)
	}
	return &config{rescore: report, out: out}, nil
}

// modelChoice is the models of a run and the members of the demo route.
type modelChoice struct {
	list    []Model
	members []Model
}

// chooseModels returns the models to ask and the demo route's members: the models
// the flags name, or, when they name none and a demo route is given, the route's
// members, which are asked as the route's entries have them.
func chooseModels(f *modelFlags, demoRoute string) (modelChoice, error) {
	var c modelChoice
	if demoRoute != "" {
		members, err := LoadDemoRoute(demoRoute)
		if err != nil {
			return c, err
		}
		c.members = members
	}
	list, err := modelsFrom(f)
	if err != nil {
		return c, err
	}
	c.list = list
	if len(c.members) > 0 && f.names == "" && f.routes == "" {
		c.list = c.members
	}
	return c, nil
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

// modelFlags are the flags that choose the models of a run.
type modelFlags struct {
	names, prices, routes, reasoning string
	structured, jsonMode             []string
}

// modelsFrom returns the models to ask: those named, or the candidates, each
// with the price the -price flag gives or the candidates' table, the
// structured-output flag by default or as asked, the reasoning control -reasoning
// gives it, and, for a model with -route entries, one run of it for each, pinned
// to its hosts, and one with the default routing too when -model lists the model.
// A model that only a route names is asked only on its routes.
func modelsFrom(f *modelFlags) ([]Model, error) {
	known, err := knownModels(f.prices)
	if err != nil {
		return nil, err
	}
	rs, err := parseRoutes(f.routes)
	if err != nil {
		return nil, err
	}
	effort, err := parseReasoning(f.reasoning)
	if err != nil {
		return nil, err
	}
	named := splitList(f.names)
	var out []Model
	used := map[string]bool{}
	for _, id := range modelIDs(f.names, &rs) {
		m, ok := known[id]
		if !ok {
			return nil, fmt.Errorf("%w: no price for %s: give one with -price %s=input:output", ErrModels, id, id)
		}
		for _, v := range routedVariants(&m, rs.byModel[id], slices.Contains(named, id)) {
			v.Reasoning = reasoningOf(effort, id, modelLabel(&v), used)
			v.Structured = structuredFor(f, id, modelLabel(&v))
			out = append(out, v)
		}
	}
	for key := range effort {
		if !used[key] {
			return nil, fmt.Errorf("%w: -reasoning names %q, which no run of this bench is", ErrUsage, key)
		}
	}
	return out, nil
}

// structuredFor reports whether the run of the model with that id and label has
// the structured-output flag: by default for the models that have it, or as
// -structured asks, and not as -json-mode asks, each by the id or the label.
func structuredFor(f *modelFlags, id, label string) bool {
	byFlag := func(list []string) bool { return slices.Contains(list, id) || slices.Contains(list, label) }
	return (slices.Contains(structuredModels(), id) || byFlag(f.structured)) && !byFlag(f.jsonMode)
}

// parseReasoning reads the -reasoning entries, model=off|low|medium|high, by the
// key of the model's id or label; a malformed entry, an unknown setting or a key
// given twice is a usage error.
func parseReasoning(s string) (map[string]string, error) {
	out := map[string]string{}
	for _, entry := range splitList(s) {
		i := strings.LastIndex(entry, "=")
		key, value := "", ""
		if i > 0 {
			key, value = strings.TrimSpace(entry[:i]), strings.TrimSpace(entry[i+1:])
		}
		if key == "" || !openaichat.ValidReasoning(value) {
			return nil, fmt.Errorf("%w: -reasoning wants model=off|low|medium|high, got %q", ErrUsage, entry)
		}
		if _, dup := out[key]; dup {
			return nil, fmt.Errorf("%w: -reasoning gives %q twice", ErrUsage, key)
		}
		out[key] = value
	}
	return out, nil
}

// reasoningOf returns the setting for the run with that id and label: the label's
// entry before the id's, "" when it has none. It marks the entry it used.
func reasoningOf(effort map[string]string, id, label string, used map[string]bool) string {
	for _, key := range []string{label, id} {
		if v, ok := effort[key]; ok {
			used[key] = true
			return v
		}
	}
	return ""
}

// modelIDs returns the ids to ask: those named, or the candidates when neither a
// name nor a route gives one, and then the models only a route names.
func modelIDs(names string, rs *routeSet) []string {
	ids := splitList(names)
	if len(ids) == 0 && len(rs.models) == 0 {
		for _, m := range candidateModels() {
			ids = append(ids, m.ID)
		}
	}
	for _, id := range rs.models {
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	return ids
}

// knownModels returns the models with a price by id: the candidates and those the
// -price flag gives.
func knownModels(prices string) (map[string]Model, error) {
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
	return known, nil
}

// routedVariants returns the runs of m: one for each of its routes, and, when
// m is listed in -model (withDefault), the run with the default routing first;
// m itself when it has no route.
func routedVariants(m *Model, routes [][]string, withDefault bool) []Model {
	if len(routes) == 0 {
		return []Model{*m}
	}
	out := make([]Model, 0, len(routes)+1)
	if withDefault {
		out = append(out, *m)
	}
	for _, route := range routes {
		v := *m
		v.Route = route
		out = append(out, v)
	}
	return out
}

// parseRoutes reads the -route entries, model=provider1|provider2: the routes of
// each model in the order given, and the models that have one, in that order. A
// model may have several; a route that names no provider, or names one twice, or
// the same route twice for a model, is a usage error.
func parseRoutes(s string) (routeSet, error) {
	set := routeSet{byModel: map[string][][]string{}}
	for _, entry := range splitList(s) {
		i := strings.LastIndex(entry, "=")
		if i <= 0 {
			return routeSet{}, fmt.Errorf("%w: -route wants model=provider1|provider2, got %q", ErrUsage, entry)
		}
		var route []string
		for p := range strings.SplitSeq(entry[i+1:], "|") {
			if p = strings.TrimSpace(p); p != "" && !slices.Contains(route, p) {
				route = append(route, p)
			}
		}
		id := entry[:i]
		if len(route) == 0 || slices.ContainsFunc(set.byModel[id], func(r []string) bool { return slices.Equal(r, route) }) {
			return routeSet{}, fmt.Errorf("%w: -route %q names no provider, or repeats a route of the model", ErrUsage, entry)
		}
		if _, seen := set.byModel[id]; !seen {
			set.models = append(set.models, id)
		}
		set.byModel[id] = append(set.byModel[id], route)
	}
	return set, nil
}

// routeSet is the -route entries: the routes of each model in the order given,
// and the models that have one, in that order.
type routeSet struct {
	byModel map[string][][]string
	models  []string
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
