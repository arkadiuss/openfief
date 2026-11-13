package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/portfolio"
)

// available reports whether the maintainer can run: the switch and the
// toggle are on, and the model portfolio is on with targets.
func (m *Maintainer) available(ctx context.Context) (bool, error) {
	on, err := m.p.Gate.MaintainerOn(ctx)
	if err != nil {
		return false, fmt.Errorf("reading the models settings: %w", err)
	}
	if !on {
		return false, nil
	}
	model, err := m.p.Model.Model(ctx)
	if err != nil {
		return false, fmt.Errorf("loading the model portfolio: %w", err)
	}
	return model.Enabled && len(model.Targets) > 0, nil
}

// queue records a new queued run.
func (m *Maintainer) queue(ctx context.Context, t RunTrigger, actor string) (string, error) {
	id, err := m.p.IDs()
	if err != nil {
		return "", fmt.Errorf("making a run ID: %w", err)
	}
	run := Run{ID: id, Trigger: t, State: RunQueued, At: m.p.Now().UTC(), Actor: actor}
	if err := m.p.Runs.QueueRun(ctx, &run); err != nil {
		return "", fmt.Errorf("queueing a maintainer run: %w", err)
	}
	return id, nil
}

// owner checks the session is the owner's and returns its ID.
func (m *Maintainer) owner(ctx context.Context, origin, sessionToken string) (string, error) {
	r, err := m.p.Sessions.Reader(ctx, origin, sessionToken)
	if err != nil {
		return "", fmt.Errorf("the maintainer: %w", err)
	}
	if !r.Owner {
		return "", ErrNotOwner
	}
	return r.SessionID, nil
}

// attempt does the work of a started run and returns how it ended.
func (m *Maintainer) attempt(ctx context.Context, run *Run) *RunFinish {
	at := m.p.Now().UTC()
	fin := &RunFinish{RunID: run.ID, At: at}
	fail := func(class string) *RunFinish {
		fin.State, fin.Class = RunFailed, class
		if class == ClassDenied {
			fin.State = RunRefused
		}
		return fin
	}
	if on, err := m.available(ctx); err != nil || !on {
		return fail(offClass(err))
	}
	model, err := m.p.Model.Model(ctx)
	if err != nil {
		return fail(ClassInternal)
	}
	held, err := m.p.Holdings.Holdings(ctx)
	if err != nil {
		return fail(holdingsClass(err))
	}
	drift, err := portfolio.DriftOf(&model, held.Values, held.Total)
	if err != nil {
		return fail(ClassNoFigures)
	}
	prompt := portfolio.BuildPrompt(&drift, held.AsOf, held.Currency)
	res, err := m.p.Asker.Ask(ctx, &AskRequest{
		Instructions: prompt.Instructions, Figures: prompt.Figures, Schema: portfolio.ResponseSchema(), Actor: run.Actor,
	})
	if err != nil {
		return fail(askClass(err))
	}
	answer, err := portfolio.ParseAnswer(res.Content, &model, &drift)
	if err != nil {
		return fail(ClassInvalidOutput)
	}
	fin.State, fin.Model, fin.Note, fin.Proposal = RunOK, res.Model, answer.Note, answer.Proposal
	if answer.Dropped {
		fin.Within = m.auditDropped(run.ID, at)
	}
	return fin
}

// offClass is the class a run ends in when the maintainer can't run: denied
// when it was turned off, internal when the core couldn't tell.
func offClass(err error) string {
	if err != nil {
		return ClassInternal
	}
	return ClassDenied
}

// holdingsClass is the class a failed read of the holdings ends a run in.
func holdingsClass(err error) string {
	if errors.Is(err, ErrNoHoldings) {
		return ClassNoFigures
	}
	return ClassInternal
}

// askClass is the class a failed ask ends a run in.
func askClass(err error) string {
	if ae, ok := errors.AsType[*AskError](err); ok {
		return ae.Class
	}
	return ClassInternal
}

// auditDropped returns the hook that writes the audit entry of a dropped
// proposal inside the transaction that stores the run.
func (m *Maintainer) auditDropped(runID string, at time.Time) func(context.Context) error {
	return func(ctx context.Context) error {
		return m.p.Audit.Record(ctx, &AuditRecord{Event: AuditDropped, Outcome: AuditRefused, RunID: runID, At: at})
	}
}

// finish stores how a run ended with its audit entry, then settles the
// Inbox: the proposal's item raised, the superseded ones cleared, and the
// failing item raised or cleared.
func (m *Maintainer) finish(ctx context.Context, fin *RunFinish) error {
	dropped := fin.Within
	fin.Within = func(ctx context.Context) error {
		if dropped != nil {
			if err := dropped(ctx); err != nil {
				return err
			}
		}
		return m.p.Audit.Record(ctx, &AuditRecord{
			Event: AuditRun, Outcome: outcomeOf(fin.State), RunID: fin.RunID, Class: fin.Class, At: fin.At,
		})
	}
	superseded, err := m.p.Runs.FinishRun(ctx, fin)
	if err != nil {
		return fmt.Errorf("storing the maintainer run: %w", err)
	}
	for _, id := range superseded {
		if err := m.p.Decisions.Clear(ctx, KindTargetProposal, id); err != nil {
			return fmt.Errorf("clearing a superseded proposal's item: %w", err)
		}
	}
	return m.settle(ctx, fin.RunID)
}

// outcomeOf is the audit outcome of a final run state.
func outcomeOf(s RunState) AuditOutcome {
	switch s {
	case RunOK:
		return AuditOK
	case RunRefused:
		return AuditRefused
	case RunQueued, RunRunning, RunFailed:
	}
	return AuditFailed
}

// settle makes the Inbox say what the final run id left: its proposal's item
// when it is still pending, and the failing item as the latest weekly runs
// decide. A run that is not final, or unknown, changes nothing.
func (m *Maintainer) settle(ctx context.Context, runID string) error {
	run, ok, err := m.p.Runs.Run(ctx, runID)
	if err != nil {
		return fmt.Errorf("loading the maintainer run: %w", err)
	}
	if !ok || run.State.Active() {
		return nil
	}
	p, found, err := m.p.Proposals.Proposal(ctx, runID)
	if err != nil {
		return fmt.Errorf("loading the proposal: %w", err)
	}
	if found && p.Status == ProposalPending {
		detail := fmt.Sprintf("The maintainer proposes changing %d target weight(s). Apply or dismiss it.", len(p.Proposal.Changes))
		if err := m.p.Decisions.Raise(ctx, KindTargetProposal, runID, detail); err != nil {
			return fmt.Errorf("raising the proposal's item: %w", err)
		}
	}
	return m.settleFailing(ctx, &run)
}

// settleFailing raises the failing item when the latest FailingRuns weekly
// runs all ended without a note, and clears it when a run left one.
func (m *Maintainer) settleFailing(ctx context.Context, run *Run) error {
	if run.State == RunOK {
		return m.clearFailing(ctx)
	}
	if run.Trigger != TriggerWeekly {
		return nil
	}
	recent, err := m.p.Runs.RecentWeekly(ctx, FailingRuns)
	if err != nil {
		return fmt.Errorf("loading the latest weekly runs: %w", err)
	}
	if len(recent) < FailingRuns {
		return nil
	}
	if slices.Contains(recent, RunOK) {
		return nil
	}
	detail := fmt.Sprintf("The maintainer's weekly run has ended without a note %d times in a row; the last one ended as %q.", FailingRuns, run.Class)
	if err := m.p.Decisions.Raise(ctx, KindMaintainerFailing, FailingSubject, detail); err != nil {
		return fmt.Errorf("raising the failing item: %w", err)
	}
	return nil
}

// standing returns what every reader is shown: whether the maintainer can run,
// the latest note and whether a run is under way; and the latest run, nil when
// there is none.
func (m *Maintainer) standing(ctx context.Context) (Overview, *Run, error) {
	var out Overview
	var err error
	if out.Available, err = m.available(ctx); err != nil {
		return Overview{}, nil, err
	}
	note, ok, err := m.p.Proposals.LatestNote(ctx)
	if err != nil {
		return Overview{}, nil, fmt.Errorf("loading the maintainer's note: %w", err)
	}
	if ok {
		out.Note = &note
	}
	latest, ok, err := m.p.Runs.LatestRun(ctx)
	if err != nil || !ok {
		return out, nil, wrapLatest(err)
	}
	out.Running = latest.State.Active()
	return out, &latest, nil
}

// wrapLatest describes a failed read of the latest run; nil is nil.
func wrapLatest(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("loading the latest maintainer run: %w", err)
}

// pending counts the proposals waiting for the owner.
func (m *Maintainer) pending(ctx context.Context) (int, error) {
	ids, err := m.p.Proposals.PendingProposals(ctx)
	if err != nil {
		return 0, fmt.Errorf("loading the pending proposals: %w", err)
	}
	return len(ids), nil
}

// clearFailing clears the failing item.
func (m *Maintainer) clearFailing(ctx context.Context) error {
	if err := m.p.Decisions.Clear(ctx, KindMaintainerFailing, FailingSubject); err != nil {
		return fmt.Errorf("clearing the maintainer's failing item: %w", err)
	}
	return nil
}

// settled clears the Inbox item of proposal id, now decided, and returns the
// proposal as stored.
func (m *Maintainer) settled(ctx context.Context, id string) (StoredProposal, error) {
	if err := m.p.Decisions.Clear(ctx, KindTargetProposal, id); err != nil {
		return StoredProposal{}, fmt.Errorf("clearing the proposal's item: %w", err)
	}
	p, ok, err := m.p.Proposals.Proposal(ctx, id)
	if err != nil || !ok {
		return StoredProposal{}, fmt.Errorf("loading the proposal: %w", errors.Join(err, ErrUnknownProposal))
	}
	return p, nil
}
