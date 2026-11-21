package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/ui"
)

// runsPerCase is how many times each case is asked: a case passes only if all
// of them pass (pass^5, D-217).
const runsPerCase = 5

// ErrRun reports a bench that can't start or finish its run.
var ErrRun = errors.New("the bench's run")

// Options are what a run is told.
type Options struct {
	Models  []Model
	BaseURL string
	// APIKey is the provider key. It goes only to the provider, in a request
	// header: no report, log or error holds it.
	APIKey string
	// Runs is how many times each case is asked.
	Runs     int
	Parallel int
	MaxCost  kernel.Decimal
	MaxCalls int
	Limits   ui.AskLimits
	// Cases, when not empty, are the ids of the cases to run.
	Cases []string
	// Opinions runs the opinions set (D-225, D-229) instead of the eval's cases:
	// its own seed and cases, graded with the same loop, recorded in no history.
	Opinions bool
	// Judge, when set, grades the advice cases beside the phrase list.
	Judge *adviceJudge
	// HTTP sends the requests to the model.
	HTTP *http.Client
	Now  Clock
}

// task is one question to ask: a model's run of a case.
type task struct {
	model int
	c     *Case
	run   int
}

// Bench is a prepared run: the portfolio, its read service, the cases and the
// text the scorers need.
type Bench struct {
	opts  Options
	seed  *Seed
	pf    *Portfolio
	cases []Case
	// total is how many cases the bench has, of which cases are the ones asked.
	total int
	vocab []string
	names []string
	// classes are the gross values by asset class on the bench's day.
	classes map[string]kernel.Decimal
	prompt  []string
	meter   *meter
}

// benchInputs builds the seed, the portfolio and the cases the bench asks of
// opts: the eval's, or with opts.Opinions the opinions set's.
func benchInputs(opts *Options) (*Seed, *Portfolio, []Case, error) {
	build, load := BuildSeed, LoadCases
	if opts.Opinions {
		build = func(bool) (*Seed, error) { return BuildOpinionsSeed() }
		load = LoadOpinionCases
	}
	seed, err := build(true)
	if err != nil {
		return nil, nil, nil, err
	}
	pf, err := NewPortfolio(seed)
	if err != nil {
		return nil, nil, nil, err
	}
	cases, err := load(pf)
	return seed, pf, cases, err
}

// pickCases returns the cases the ids name, in the order of ids.
func pickCases(cases []Case, ids []string) ([]Case, error) {
	var kept []Case
	for _, id := range ids {
		i := slices.IndexFunc(cases, func(c Case) bool { return c.ID == id })
		if i < 0 {
			return nil, fmt.Errorf("%w: no case %q", ErrRun, id)
		}
		kept = append(kept, cases[i])
	}
	return kept, nil
}

// NewBench builds the portfolio and reads the cases for opts.
func NewBench(opts *Options) (*Bench, error) {
	seed, pf, cases, err := benchInputs(opts)
	if err != nil {
		return nil, err
	}
	total := len(cases)
	if len(opts.Cases) > 0 {
		if cases, err = pickCases(cases, opts.Cases); err != nil {
			return nil, err
		}
	}
	o := *opts
	o.Limits = resolveLimits(o.Limits)
	if o.Runs <= 0 {
		o.Runs = runsPerCase
	}
	if o.Parallel <= 0 {
		o.Parallel = 1
	}
	classes, err := classTotals(seed, pf)
	if err != nil {
		return nil, err
	}
	return &Bench{
		opts: o, seed: seed, pf: pf, cases: cases, total: total, vocab: vocabularyOf(seed), names: realNamesOf(seed), prompt: ui.AskPromptNonPublicFor(seed.Today, opts.Opinions),
		classes: classes,
		meter:   newMeter(o.MaxCost, o.MaxCalls),
	}, nil
}

// Run asks every case of every model Runs times and returns the report. It
// stops asking, cleanly, when the budget is spent: the report then says so and
// counts the questions it never asked as not run.
func (b *Bench) Run(ctx context.Context) (*Report, error) {
	var tasks []task
	for mi := range b.opts.Models {
		for ci := range b.cases {
			for r := 1; r <= b.opts.Runs; r++ {
				tasks = append(tasks, task{model: mi, c: &b.cases[ci], run: r})
			}
		}
	}
	results, stopped := b.work(ctx, tasks)
	resolveRejections(tasks, results)
	return b.report(tasks, results, stopped)
}

// work asks the tasks, Parallel at a time, and returns their results in order
// and whether the budget stopped the run: after that no question is asked.
func (b *Bench) work(ctx context.Context, tasks []task) ([]RunReport, bool) {
	results := make([]RunReport, len(tasks))
	var stopped atomic.Bool
	var wg sync.WaitGroup
	queue := make(chan int)
	for range b.opts.Parallel {
		wg.Go(func() {
			for i := range queue {
				results[i] = b.askUnlessStopped(ctx, &tasks[i], &stopped)
			}
		})
	}
	for i := range tasks {
		queue <- i
	}
	close(queue)
	wg.Wait()
	return results, stopped.Load()
}

// askUnlessStopped asks t, or reports it not run once the budget has stopped
// the run; a question that finds the budget spent stops it. A question the box
// answers itself with one of its commands is not asked at all: in production
// the matcher answers it before the model (D-203), so the run is tagged as the
// matcher's, whatever the budget.
func (b *Bench) askUnlessStopped(ctx context.Context, t *task, stopped *atomic.Bool) RunReport {
	if ui.AskAnswersItself(t.c.Question, b.seed.Today) {
		return RunReport{Run: t.run, Matcher: true}
	}
	if stopped.Load() {
		return RunReport{Run: t.run, NotRun: true}
	}
	rep := b.ask(ctx, t)
	if rep.BudgetSpent {
		stopped.Store(true)
	}
	return rep
}

// resolveLimits returns the limits with D-208's defaults for each zero field,
// as the loop applies them, so the scorers hold the run to what the loop did.
func resolveLimits(l ui.AskLimits) ui.AskLimits {
	d := ui.DefaultAskLimits()
	if l.ToolRounds > 0 {
		d.ToolRounds = l.ToolRounds
	}
	if l.ToolCalls > 0 {
		d.ToolCalls = l.ToolCalls
	}
	if l.OutputTokens > 0 {
		d.OutputTokens = l.OutputTokens
	}
	if l.Deadline > 0 {
		d.Deadline = l.Deadline
	}
	return d
}

// ask runs one task: a question in a conversation of its own, graded.
func (b *Bench) ask(ctx context.Context, t *task) RunReport {
	model := &b.opts.Models[t.model]
	mr := &modelRun{client: benchClient(b.opts.HTTP, b.opts.BaseURL, b.opts.APIKey, model), model: model, meter: b.meter, now: b.opts.Now}
	style := t.c.numberFormat()
	asker, err := ui.NewAsker(&ui.AskerConfig{
		Read: newReadClient(b.seed, b.pf), Model: mr, Today: func() kernel.Date { return b.seed.Today },
		Limits: b.opts.Limits, Group: style.Group, Decimal: style.Decimal, MarketQuotes: true, DeniedTiers: t.c.Route.DenyTiers, Opinions: t.c.Opinions,
	})
	if err != nil {
		return RunReport{Run: t.run, Error: errorText(err)}
	}
	started := b.opts.Now()
	run, err := asker.AskAfter(ctx, t.c.Question, t.c.earlier())
	d := &RunData{
		Case: t.c, Run: run, Calls: mr.calls, Limits: b.opts.Limits, Duration: b.opts.Now().Sub(started),
		NonPublic: b.prompt, Canary: ui.AskPromptCanary(), Vocabulary: b.vocab, RealNames: b.names, Today: b.seed.Today, Classes: b.classes,
	}
	rep := RunReport{Run: t.run, Cost: mr.cost.String(), LatencyMs: d.Duration.Milliseconds()}
	for i := range mr.calls {
		rep.addCall(&mr.calls[i])
		rep.Rounds = append(rep.Rounds, mr.calls[i].Round)
	}
	rep.addOutcomes(mr.calls)
	if err != nil {
		rep.Error = errorText(err)
		return rep
	}
	if rep.Errored {
		return rep
	}
	if mr.budgetHit {
		rep.BudgetSpent = true
		return rep
	}
	if gradeErr := b.grade(ctx, d, &rep); gradeErr != nil {
		rep.Error = errorText(gradeErr)
	}
	return rep
}

// addOutcomes counts the answered calls of the run and lists the calls that
// ended it in failure, and marks the run void when a fault of the bench's or of
// the credentials is one.
func (r *RunReport) addOutcomes(calls []ModelCall) {
	for i := range calls {
		c := &calls[i]
		switch {
		case c.Error == "":
			r.Completions++
		case !c.Retried:
			// A failed call that was retried is told by its retry's outcome, and in
			// the reliability counts: only the call that ended the attempt is an error.
			ce := CallErrorReport{Status: c.Status, Label: c.Label, Class: c.Error, Fault: c.Fault, Rule: c.Rule}
			if isEvidenceRejection(c.Status) {
				ce.RequestHash = c.RequestHash
			}
			r.CallErrors = append(r.CallErrors, ce)
		}
	}
	r.settle()
}

// settle decides from the run's call errors whether a failed call ended it and
// whether it is void: a bench's fault or a refused key or no credits make it
// void, the credentials' reason winning, since it says that nothing of the run
// can be trusted. A void run is no host's or model's. It is called again when
// the evidence of the whole run has decided a 4xx (resolveRejections).
func (r *RunReport) settle() {
	r.Errored = len(r.CallErrors) > 0
	r.Void, r.VoidReason = false, ""
	for i := range r.CallErrors {
		switch r.CallErrors[i].Fault {
		case faultCredentials:
			r.Void, r.VoidReason = true, voidCredentials
		case faultBench:
			if r.VoidReason == "" {
				r.Void, r.VoidReason = true, voidBench
			}
		}
	}
}

// addCall adds a call of the run to its tokens, its estimated-cost count and the
// hosts that served it.
func (r *RunReport) addCall(c *ModelCall) {
	r.InputTokens += c.InputTokens
	r.OutputTokens += c.OutputTokens
	r.Attempts++
	if c.Error != "" && c.Fault == faultHost {
		r.TransportFailures++
	}
	if c.Retried {
		r.Retries = append(r.Retries, RetryReport{Status: c.Status, Label: c.Label, Recovered: c.Recovered})
	}
	if c.Error != "" {
		return
	}
	if c.RequestHash != "" {
		r.accepted = append(r.accepted, c.RequestHash)
	}
	if !c.Measured {
		r.EstimatedCalls++
	}
	if c.Provider != "" {
		if r.Providers == nil {
			r.Providers = map[string]int{}
		}
		r.Providers[c.Provider]++
	}
}

// grade scores the run's answer and fills the report with the results.
func (b *Bench) grade(ctx context.Context, d *RunData, rep *RunReport) error {
	results, err := grade(ctx, d)
	if err != nil {
		return err
	}
	rep.fill(d, results)
	if b.opts.Judge != nil && d.Case.Kind == kindRefusalAdvice && d.Run.Answered {
		v := b.opts.Judge.verdict(ctx, d.Case.Question, d.Run.Text)
		rep.Judge = &v
		rep.PhraseAdvice = adviceOf(d)
	}
	return nil
}
