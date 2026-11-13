package app_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
)

// Sessions the fake knows: the owner's, one that may view every account (the
// household's reader), and one that may not.
const (
	ownerToken   = "ofsess_owner"
	readerToken  = "ofsess_reader"
	limitedToken = "ofsess_limited"
)

var errAuditDown = errors.New("audit down")

// memStore is the maintainer's store in memory, with the semantics of the
// PostgreSQL one: one active run at a time, a finish stored whole or not at
// all, a proposal decided once.
type memStore struct {
	model     portfolio.Model
	runs      []*memRun
	proposals map[string]*app.StoredProposal
	notes     map[string]string // run ID -> note
	models    map[string]string // run ID -> model
}

type memRun struct{ app.Run }

func newMemStore(m portfolio.Model) *memStore {
	return &memStore{model: m, proposals: map[string]*app.StoredProposal{}, notes: map[string]string{}, models: map[string]string{}}
}

func (s *memStore) Model(context.Context) (portfolio.Model, error) { return s.model, nil }

func (s *memStore) SaveModel(_ context.Context, m portfolio.Model) error {
	s.model = m
	return nil
}

func (s *memStore) QueueRun(_ context.Context, run *app.Run) error {
	for _, r := range s.runs {
		if r.State.Active() {
			if run.At.Sub(r.At) < app.StaleAfter {
				return app.ErrRunInProgress
			}
			r.State, r.Class = app.RunFailed, app.ClassInternal
		}
	}
	s.runs = append(s.runs, &memRun{*run})
	return nil
}

func (s *memStore) StartRun(_ context.Context, id string, at time.Time) (app.Run, bool, error) {
	r := s.find(id)
	if r == nil || !r.State.Active() {
		return app.Run{}, false, nil
	}
	r.State, r.At = app.RunRunning, at
	return r.Run, true, nil
}

func (s *memStore) Run(_ context.Context, id string) (app.Run, bool, error) {
	if r := s.find(id); r != nil {
		return r.Run, true, nil
	}
	return app.Run{}, false, nil
}

func (s *memStore) FinishRun(ctx context.Context, fin *app.RunFinish) ([]string, error) {
	r := s.find(fin.RunID)
	if r == nil || !r.State.Active() {
		return nil, app.ErrRunFinal
	}
	if fin.Within != nil {
		if err := fin.Within(ctx); err != nil {
			return nil, fmt.Errorf("recording the run: %w", err)
		}
	}
	r.State, r.Class, r.At = fin.State, fin.Class, fin.At
	var superseded []string
	if fin.State == app.RunOK {
		s.notes[r.ID], s.models[r.ID] = fin.Note, fin.Model
	}
	if fin.Proposal != nil {
		for id, p := range s.proposals {
			if p.Status == app.ProposalPending {
				p.Status = app.ProposalSuperseded
				superseded = append(superseded, id)
			}
		}
		s.proposals[r.ID] = &app.StoredProposal{
			ID: r.ID, Status: app.ProposalPending, CreatedAt: fin.At, Model: fin.Model, Note: fin.Note, Proposal: *fin.Proposal,
		}
	}
	slices.Sort(superseded)
	return superseded, nil
}

func (s *memStore) LatestNote(context.Context) (app.Note, bool, error) {
	for _, r := range slices.Backward(s.runs) {
		if note, ok := s.notes[r.ID]; ok {
			return app.Note{RunID: r.ID, Text: note, Model: s.models[r.ID], At: r.At, Trigger: r.Trigger}, true, nil
		}
	}
	return app.Note{}, false, nil
}

func (s *memStore) LatestRun(context.Context) (app.Run, bool, error) {
	if len(s.runs) == 0 {
		return app.Run{}, false, nil
	}
	return s.runs[len(s.runs)-1].Run, true, nil
}

func (s *memStore) RecentWeekly(_ context.Context, limit int) ([]app.RunState, error) {
	var out []app.RunState
	for i := len(s.runs) - 1; i >= 0 && len(out) < limit; i-- {
		if r := s.runs[i]; r.Trigger == app.TriggerWeekly && !r.State.Active() {
			out = append(out, r.State)
		}
	}
	return out, nil
}

func (s *memStore) PendingProposals(context.Context) ([]string, error) {
	var out []string
	for id, p := range s.proposals {
		if p.Status == app.ProposalPending {
			out = append(out, id)
		}
	}
	slices.Sort(out)
	return out, nil
}

func (s *memStore) Proposal(_ context.Context, id string) (app.StoredProposal, bool, error) {
	if p, ok := s.proposals[id]; ok {
		return *p, true, nil
	}
	return app.StoredProposal{}, false, nil
}

func (s *memStore) ApplyProposal(ctx context.Context, id string, at time.Time,
	apply func(*portfolio.Model, *portfolio.Proposal) (portfolio.Model, error), within func(context.Context) error,
) error {
	p, err := s.pending(id)
	if err != nil {
		return err
	}
	next, err := apply(&s.model, &p.Proposal)
	if err != nil {
		return err
	}
	if err := within(ctx); err != nil {
		return err
	}
	s.model, p.Status, p.DecidedAt = next, app.ProposalApplied, at
	return nil
}

func (s *memStore) DismissProposal(ctx context.Context, id string, at time.Time, within func(context.Context) error) error {
	p, err := s.pending(id)
	if err != nil {
		return err
	}
	if err := within(ctx); err != nil {
		return err
	}
	p.Status, p.DecidedAt = app.ProposalDismissed, at
	return nil
}

func (s *memStore) find(id string) *memRun {
	for _, r := range s.runs {
		if r.ID == id {
			return r
		}
	}
	return nil
}

func (s *memStore) pending(id string) (*app.StoredProposal, error) {
	p, ok := s.proposals[id]
	switch {
	case !ok:
		return nil, app.ErrUnknownProposal
	case p.Status != app.ProposalPending:
		return nil, app.ErrProposalDecided
	}
	return p, nil
}

// fakeGate is the switch and toggle.
type fakeGate struct{ on bool }

func (g *fakeGate) MaintainerOn(context.Context) (bool, error) { return g.on, nil }

// fakeHoldings answers the valued holdings, or an error.
type fakeHoldings struct {
	held app.Holdings
	err  error
}

func (h *fakeHoldings) Holdings(context.Context) (app.Holdings, error) { return h.held, h.err }

// fakeAsker answers with content or an error, and keeps what it was asked.
type fakeAsker struct {
	content string
	model   string
	err     error
	calls   []app.AskRequest
}

func (a *fakeAsker) Ask(_ context.Context, req *app.AskRequest) (app.AskResult, error) {
	a.calls = append(a.calls, *req)
	if a.err != nil {
		return app.AskResult{}, a.err
	}
	return app.AskResult{Content: a.content, Model: a.model}, nil
}

// fakeDecisions keeps the open items of the Inbox by kind and subject.
type fakeDecisions struct {
	open map[string]string
	err  error
}

func newFakeDecisions() *fakeDecisions { return &fakeDecisions{open: map[string]string{}} }

func (d *fakeDecisions) Raise(_ context.Context, kind, subject, detail string) error {
	if d.err != nil {
		return d.err
	}
	d.open[kind+"/"+subject] = detail
	return nil
}

func (d *fakeDecisions) Clear(_ context.Context, kind, subject string) error {
	if d.err != nil {
		return d.err
	}
	delete(d.open, kind+"/"+subject)
	return nil
}

// fakeAudit keeps the records, or fails.
type fakeAudit struct {
	records []app.AuditRecord
	err     error
}

func (a *fakeAudit) Record(_ context.Context, r *app.AuditRecord) error {
	if a.err != nil {
		return a.err
	}
	a.records = append(a.records, *r)
	return nil
}

// events lists the audit records as "event:outcome[:class]".
func (a *fakeAudit) events() []string {
	out := make([]string, 0, len(a.records))
	for _, r := range a.records {
		e := string(r.Event) + ":" + string(r.Outcome)
		if r.Class != "" {
			e += ":" + r.Class
		}
		out = append(out, e)
	}
	return out
}

// fakeQueue keeps the job IDs enqueued.
type fakeQueue struct {
	jobs []string
	err  error
}

func (q *fakeQueue) Enqueue(_ context.Context, runID string) error {
	if q.err != nil {
		return q.err
	}
	q.jobs = append(q.jobs, runID)
	return nil
}

// fakeSessions knows the three tokens above.
type fakeSessions struct{}

func (fakeSessions) Reader(_ context.Context, _, token string) (app.Reader, error) {
	switch token {
	case ownerToken:
		return app.Reader{SessionID: "00000000000000a1", Owner: true, ViewsAll: true}, nil
	case readerToken:
		return app.Reader{SessionID: "00000000000000b2", ViewsAll: true}, nil
	case limitedToken:
		return app.Reader{SessionID: "00000000000000c3"}, nil
	}
	return app.Reader{}, app.ErrNoSession
}

// maintainerRig is a Maintainer over the fakes, with the clock a test moves.
type maintainerRig struct {
	m         *app.Maintainer
	store     *memStore
	gate      *fakeGate
	holdings  *fakeHoldings
	asker     *fakeAsker
	decisions *fakeDecisions
	audit     *fakeAudit
	queue     *fakeQueue
	now       time.Time
	nextID    int
}

// The two instruments of the maintainer's model.
const (
	instA = "IETEST000026/ETFA"
	instB = "PLTEST000010/ETFB"
)

// newRig returns a rig with the maintainer on, a valid model of 60/40 and a
// portfolio holding 70% and 20% of its value in the two instruments.
func newRig(t *testing.T) *maintainerRig {
	t.Helper()
	model := portfolio.Model{Enabled: true, Targets: []portfolio.Target{
		{Kind: portfolio.TargetInstrument, Name: instA, Weight: mustDec(t, "60")},
		{Kind: portfolio.TargetInstrument, Name: instB, Weight: mustDec(t, "40")},
	}}
	r := &maintainerRig{
		store: newMemStore(model), gate: &fakeGate{on: true}, asker: &fakeAsker{model: "local/synth"},
		decisions: newFakeDecisions(), audit: &fakeAudit{}, queue: &fakeQueue{},
		now: time.Date(2026, 11, 13, 6, 0, 0, 0, time.UTC),
		holdings: &fakeHoldings{held: app.Holdings{
			AsOf: mustDate(t, "2026-11-12"), Currency: mustCurrency(t, "PLN"), Total: mustDec(t, "10000"),
			Values: map[string]kernel.Decimal{instA: mustDec(t, "7000"), instB: mustDec(t, "2000")},
		}},
	}
	r.m = app.NewMaintainer(&app.MaintainerPorts{
		Runs: r.store, Proposals: r.store, Model: r.store, Gate: r.gate, Holdings: r.holdings, Asker: r.asker, Decisions: r.decisions,
		Audit: r.audit, Queue: r.queue, Sessions: fakeSessions{},
		IDs: func() (string, error) {
			r.nextID++
			return fmt.Sprintf("%016x", r.nextID), nil
		},
		Now: func() time.Time { return r.now },
	})
	return r
}

// advance moves the rig's clock.
func (r *maintainerRig) advance(d time.Duration) { r.now = r.now.Add(d) }
