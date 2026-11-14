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
	total  int
	vocab  []string
	prompt []string
	meter  *meter
}

// NewBench builds the portfolio and reads the cases for opts.
func NewBench(opts *Options) (*Bench, error) {
	seed, err := BuildSeed(true)
	if err != nil {
		return nil, err
	}
	pf, err := NewPortfolio(seed)
	if err != nil {
		return nil, err
	}
	cases, err := LoadCases(pf)
	if err != nil {
		return nil, err
	}
	total := len(cases)
	if len(opts.Cases) > 0 {
		var kept []Case
		for _, id := range opts.Cases {
			i := slices.IndexFunc(cases, func(c Case) bool { return c.ID == id })
			if i < 0 {
				return nil, fmt.Errorf("%w: no case %q", ErrRun, id)
			}
			kept = append(kept, cases[i])
		}
		cases = kept
	}
	o := *opts
	o.Limits = resolveLimits(o.Limits)
	if o.Runs <= 0 {
		o.Runs = runsPerCase
	}
	if o.Parallel <= 0 {
		o.Parallel = 1
	}
	return &Bench{
		opts: o, seed: seed, pf: pf, cases: cases, total: total, vocab: vocabularyOf(seed), prompt: promptTexts(seed.Today),
		meter: newMeter(o.MaxCost, o.MaxCalls),
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
// the run; a question that finds the budget spent stops it.
func (b *Bench) askUnlessStopped(ctx context.Context, t *task, stopped *atomic.Bool) RunReport {
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
	mr := &modelRun{client: newChatClient(b.opts.HTTP, b.opts.BaseURL, b.opts.APIKey), model: model, meter: b.meter, now: b.opts.Now}
	style := t.c.numberFormat()
	asker, err := ui.NewAsker(&ui.AskerConfig{
		Read: newReadClient(b.seed, b.pf), Model: mr, Today: func() kernel.Date { return b.seed.Today },
		Limits: b.opts.Limits, Group: style.Group, Decimal: style.Decimal,
	})
	if err != nil {
		return RunReport{Run: t.run, Error: errorText(err)}
	}
	started := b.opts.Now()
	run, err := asker.Ask(ctx, t.c.Question)
	d := &RunData{
		Case: t.c, Run: run, Calls: mr.calls, Limits: b.opts.Limits, Duration: b.opts.Now().Sub(started),
		Prompt: b.prompt, Vocabulary: b.vocab,
	}
	rep := RunReport{Run: t.run, Cost: mr.cost.String(), LatencyMs: d.Duration.Milliseconds()}
	for _, c := range mr.calls {
		rep.InputTokens += c.InputTokens
		rep.OutputTokens += c.OutputTokens
	}
	for _, c := range mr.calls {
		if c.Error == "" {
			rep.Completions++
			continue
		}
		rep.CallErrors = append(rep.CallErrors, CallErrorReport{Status: c.Status, Label: c.Label})
	}
	rep.Errored = len(rep.CallErrors) > 0
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
		rep.PhraseAdvice = adviceIn(d.Run.Text)
	}
	return nil
}
