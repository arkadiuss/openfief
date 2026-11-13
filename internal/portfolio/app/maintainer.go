package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// Errors of the model-portfolio maintainer's use cases (D-204).
var (
	// ErrNotReader reports a valid session that may not read the maintainer's
	// note: it speaks of the whole portfolio, so only the owner and a session
	// that may view every account read it (D-163).
	ErrNotReader = errors.New("this session may not read the whole portfolio")
	// ErrMaintainerOff reports a maintainer that can't run: the models switch
	// or the maintainer's toggle is off (D-212), or the model portfolio is off
	// or empty (D-52).
	ErrMaintainerOff = errors.New("the maintainer is off")
	// ErrRunInProgress reports a run asked for while another is queued or
	// running for the model portfolio: one at a time.
	ErrRunInProgress = errors.New("a maintainer run is already queued or running")
	// ErrUnknownProposal reports a proposal no run made.
	ErrUnknownProposal = errors.New("unknown proposal")
	// ErrProposalDecided reports a proposal the owner already applied or
	// dismissed, or that a later run superseded.
	ErrProposalDecided = errors.New("the proposal is already decided")
)

// FailingRuns is how many weekly runs in a row must end without a note
// before the Inbox says so (D-196). One is a hiccup of a provider the route's
// fallbacks and next week's run cover; two is a fortnight with no note, which
// is the owner's to look at: a route with no model that may receive the data,
// a key that stopped working or a provider that is gone.
const FailingRuns = 2

// Attention kinds the maintainer raises (D-196), both among the decisions.
const (
	// KindTargetProposal is a proposed set of target changes waiting for the
	// owner to apply or dismiss (D-204). Subject: the proposal's ID, which is
	// its run's.
	KindTargetProposal = "model-target-proposal"
	// KindMaintainerFailing is a maintainer whose weekly run has failed
	// FailingRuns times in a row. Subject: "maintainer". It clears when a run
	// leaves a note, or when the maintainer is turned off.
	KindMaintainerFailing = "maintainer-failing"
)

// FailingSubject is the subject of the KindMaintainerFailing item.
const FailingSubject = "maintainer"

// RunTrigger says what started a run.
type RunTrigger string

// The triggers.
const (
	// TriggerWeekly is the weekly schedule.
	TriggerWeekly RunTrigger = "weekly"
	// TriggerOwner is the owner asking.
	TriggerOwner RunTrigger = "owner"
)

// RunState is where a run stands.
type RunState string

// The states of a run. Queued and running runs are active; the others are
// final.
const (
	RunQueued  RunState = "queued"
	RunRunning RunState = "running"
	// RunOK is a run that left a note.
	RunOK RunState = "ok"
	// RunRefused is a run the router refused: the switch, the toggle, the
	// route or the data tiers did not allow it.
	RunRefused RunState = "refused"
	// RunFailed is any other run that left no note.
	RunFailed RunState = "failed"
)

// Active reports whether s is queued or running.
func (s RunState) Active() bool { return s == RunQueued || s == RunRunning }

// Error classes a run ends in, besides the model router's (identity's
// ModelErrorClasses): the router's refusal, a usable answer that was not,
// and no figures to send.
const (
	ClassDenied        = "denied"
	ClassInvalidOutput = "invalid_output"
	ClassNoFigures     = "no_figures"
	ClassInternal      = "internal"
)

// Run is one run of the maintainer.
type Run struct {
	ID      string
	Trigger RunTrigger
	State   RunState
	// At is when the run was queued, started or ended, as its state says.
	At time.Time
	// Class is the class a refused or failed run ended in.
	Class string
	// Actor is the owner's session that asked for the run, empty for the
	// schedule; the audit entries of its model calls name it.
	Actor string
}

// Note is the maintainer's latest note.
type Note struct {
	RunID   string
	Text    string
	Model   string
	At      time.Time
	Trigger RunTrigger
}

// ProposalStatus is where a proposal stands.
type ProposalStatus string

// The statuses.
const (
	ProposalPending    ProposalStatus = "pending"
	ProposalApplied    ProposalStatus = "applied"
	ProposalDismissed  ProposalStatus = "dismissed"
	ProposalSuperseded ProposalStatus = "superseded"
)

// StoredProposal is a proposal with its run's note and where it stands.
type StoredProposal struct {
	ID        string
	Status    ProposalStatus
	CreatedAt time.Time
	// DecidedAt is zero while pending.
	DecidedAt time.Time
	Model     string
	Note      string
	Proposal  portfolio.Proposal
}

// RunFinish is how a run ended, for FinishRun to store whole: the note and
// proposal come with an OK run and nothing else does.
type RunFinish struct {
	RunID string
	State RunState
	Class string
	At    time.Time
	// Model, Note and Proposal are set for an OK run; Proposal only when the
	// model proposed changes that passed every rule.
	Model    string
	Note     string
	Proposal *portfolio.Proposal
	// Within runs inside the database transaction that stores the finish,
	// before it commits: the audit entries of the run. When it fails nothing
	// is stored.
	Within func(ctx context.Context) error
}

// RunStore keeps the maintainer's runs. *portfolio/postgres.Store implements
// it.
type RunStore interface {
	// QueueRun records a new queued run, unless another is queued or
	// running, which is ErrRunInProgress. A run still active after
	// StaleAfter is abandoned first: it is closed as failed with class
	// "internal", so a lost job never blocks the maintainer.
	QueueRun(ctx context.Context, run *Run) error
	// StartRun marks the run id as running and returns it, and reports false
	// when it has ended. A run still running is taken over: its job is
	// retried only after the earlier attempt ended.
	StartRun(ctx context.Context, id string, at time.Time) (Run, bool, error)
	// Run returns the run id, and false when there is none.
	Run(ctx context.Context, id string) (Run, bool, error)
	// FinishRun stores how a run ended, its note and its proposal, in one
	// transaction, and marks every earlier pending proposal superseded; it
	// returns their IDs. A run already final is ErrRunFinal.
	FinishRun(ctx context.Context, fin *RunFinish) (superseded []string, err error)
	// LatestRun returns the latest run, active or final.
	LatestRun(ctx context.Context) (Run, bool, error)
	// RecentWeekly returns the states of the latest limit final weekly runs,
	// newest first.
	RecentWeekly(ctx context.Context, limit int) ([]RunState, error)
}

// ProposalStore keeps the notes the maintainer wrote and the proposals it
// made. *portfolio/postgres.Store implements it.
type ProposalStore interface {
	// LatestNote returns the note of the latest run that left one.
	LatestNote(ctx context.Context) (Note, bool, error)
	// PendingProposals returns the IDs of the proposals waiting for the
	// owner.
	PendingProposals(ctx context.Context) ([]string, error)
	// Proposal returns the proposal id with its note, and false when none.
	Proposal(ctx context.Context, id string) (StoredProposal, bool, error)
	// ApplyProposal applies the pending proposal id in one transaction: it
	// loads the model portfolio and calls apply with it and the proposal;
	// what apply returns replaces the model's targets, the proposal becomes
	// applied at at, and within runs before it commits. An error from apply
	// or within stores nothing. A proposal that is not pending is
	// ErrProposalDecided, one that does not exist ErrUnknownProposal.
	ApplyProposal(ctx context.Context, id string, at time.Time,
		apply func(model *portfolio.Model, p *portfolio.Proposal) (portfolio.Model, error),
		within func(ctx context.Context) error) error
	// DismissProposal marks the pending proposal id dismissed at at, and
	// runs within before it commits, with the same errors as ApplyProposal.
	DismissProposal(ctx context.Context, id string, at time.Time, within func(ctx context.Context) error) error
}

// ErrRunFinal reports a finish for a run that already ended.
var ErrRunFinal = errors.New("the maintainer run is already final")

// StaleAfter is how long a queued or running run may stay active before the
// next one treats it as lost (its job died with the process).
const StaleAfter = 30 * time.Minute

// Gate says whether the maintainer may run: the one models switch and the
// maintainer's toggle (D-212). cmd/core answers it over the models
// settings. With either off nothing runs and nothing calls a model.
type Gate interface {
	MaintainerOn(ctx context.Context) (bool, error)
}

// Holdings is what the portfolio holds, valued, as of a date: the input of
// the weights. cmd/core reads it from the views of the owner's accounts.
type Holdings struct {
	// AsOf is the date every account is valued through.
	AsOf kernel.Date
	// Currency is the reporting currency Total and Values are in.
	Currency kernel.Currency
	// Total is the value of the whole portfolio, cash included.
	Total kernel.Decimal
	// Values is the value of each instrument held, by "ISIN/symbol".
	Values map[string]kernel.Decimal
}

// HoldingsReader reads the valued holdings. An error wrapping
// ErrNoHoldings says there is nothing valued to weigh yet.
type HoldingsReader interface {
	Holdings(ctx context.Context) (Holdings, error)
}

// ErrNoHoldings reports a portfolio with nothing valued yet.
var ErrNoHoldings = errors.New("nothing is valued yet")

// AskRequest is what the maintainer sends its model.
type AskRequest struct {
	// Instructions hold no data of the portfolio (data tier T2).
	Instructions string
	// Figures are whole-portfolio data (data tier T0, D-206).
	Figures string
	// Schema is the JSON Schema of the answer.
	Schema string
	// Actor is the session of the owner who asked, empty for the schedule.
	Actor string
}

// AskResult is a model's answer.
type AskResult struct {
	// Content is the answer text, the JSON of the schema.
	Content string
	// Model is the model that answered, as "provider/model".
	Model string
}

// AskError is a failed or refused ask: Class is one of the classes a run
// ends in. It carries nothing the model or the provider wrote.
type AskError struct {
	Class string
}

// Error names the class.
func (e *AskError) Error() string { return "the maintainer's model call ended: " + e.Class }

// Asker sends a request through the model router for the maintainer's role.
// cmd/core answers it over the router: a refusal of the switch, the toggle,
// the route or the data tiers comes back as an *AskError of class "denied",
// any other failure as one of its class.
type Asker interface {
	Ask(ctx context.Context, req *AskRequest) (AskResult, error)
}

// Decisions raises and clears the maintainer's items in the Inbox's
// decisions (D-196). cmd/core answers it over the attention items.
type Decisions interface {
	Raise(ctx context.Context, kind, subject, detail string) error
	Clear(ctx context.Context, kind, subject string) error
}

// AuditEvent says what a maintainer audit record is about.
type AuditEvent string

// The events, one for each audit action of D-211 the maintainer adds.
const (
	AuditRun       AuditEvent = "run"
	AuditDropped   AuditEvent = "dropped"
	AuditApplied   AuditEvent = "applied"
	AuditDismissed AuditEvent = "dismissed"
)

// AuditOutcome is how what an audit record is about ended.
type AuditOutcome string

// The outcomes: a refused run is a refused record, and so is a proposal
// that no longer fits the model.
const (
	AuditOK      AuditOutcome = "ok"
	AuditRefused AuditOutcome = AuditOutcome(RunRefused)
	AuditFailed  AuditOutcome = AuditOutcome(RunFailed)
)

// AuditRecord is one entry of the audit log (D-211), with no text of a note,
// a reason or a figure.
type AuditRecord struct {
	Event   AuditEvent
	Outcome AuditOutcome
	RunID   string
	// Class is the class a refused or failed run ended in.
	Class string
	// Actor is the owner's session ID; empty for the core itself.
	Actor string
	At    time.Time
}

// Auditor appends to the audit log; cmd/core answers it over the identity
// context's. A record that can't be written fails what it records.
type Auditor interface {
	Record(ctx context.Context, r *AuditRecord) error
}

// QueuedRun hands a queued run to the job queue.
type QueuedRun interface {
	// Enqueue puts the job that works the run on the queue.
	Enqueue(ctx context.Context, runID string) error
}

// Reader is who a session is to the maintainer's note.
type Reader struct {
	// SessionID is the session's ID.
	SessionID string
	// Owner is true for an owner-scoped session.
	Owner bool
	// ViewsAll is true when the session may view every account (D-163): the
	// owner always does.
	ViewsAll bool
}

// Sessions checks a session on its origin. cmd/core answers it over the
// identity context, with errors wrapping ErrNoSession.
type Sessions interface {
	Reader(ctx context.Context, origin, sessionToken string) (Reader, error)
}

// MaintainerPorts is what a Maintainer reads and writes.
type MaintainerPorts struct {
	Runs      RunStore
	Proposals ProposalStore
	Model     ModelStore
	Gate      Gate
	Holdings  HoldingsReader
	Asker     Asker
	Decisions Decisions
	Audit     Auditor
	Queue     QueuedRun
	Sessions  Sessions
	// IDs makes a run's ID: 16 hex digits.
	IDs func() (string, error)
	Now func() time.Time
}

// Maintainer is the model-portfolio maintainer (D-204): OpenFief computes
// the weights and the drift, the maintainer's model writes a note about them
// and may propose target changes, which wait in the Inbox for the owner.
// It holds no tools, trades nothing, and keeps no prompt or answer beyond
// the run: only the note and the checked proposal are stored, and no text of
// either reaches a log, an audit entry or an error (GR-12).
type Maintainer struct {
	p MaintainerPorts
}

// NewMaintainer returns the maintainer over ports.
func NewMaintainer(p *MaintainerPorts) *Maintainer { return &Maintainer{p: *p} }

// Overview is what the model portfolio's page shows of the maintainer.
type Overview struct {
	// Available is true when the maintainer can run.
	Available bool
	// Note is the latest note, when there is one.
	Note *Note
	// Running is true while a run is queued or running.
	Running bool
	// LastFailure and Pending are the owner's: the latest run when it ended
	// without a note, and how many proposals wait.
	LastFailure *Run
	Pending     int
}

// Overview returns the maintainer's standing for the session: the owner's,
// or one that may view every account. Anyone else is ErrNotReader.
func (m *Maintainer) Overview(ctx context.Context, origin, sessionToken string) (Overview, error) {
	r, err := m.p.Sessions.Reader(ctx, origin, sessionToken)
	if err != nil {
		return Overview{}, fmt.Errorf("the maintainer: %w", err)
	}
	if !r.Owner && !r.ViewsAll {
		return Overview{}, ErrNotReader
	}
	out, latest, err := m.standing(ctx)
	if err != nil || !r.Owner {
		return out, err
	}
	if latest != nil && (latest.State == RunRefused || latest.State == RunFailed) {
		out.LastFailure = latest
	}
	if out.Pending, err = m.pending(ctx); err != nil {
		return Overview{}, err
	}
	return out, nil
}

// Ask queues a run for the owner's session. It is ErrMaintainerOff when the
// maintainer can't run, nothing is queued and no model is called, and
// ErrRunInProgress when one is already queued or running; started is true
// when a run was queued. It returns before the run happens.
func (m *Maintainer) Ask(ctx context.Context, origin, sessionToken string) (bool, error) {
	asker, err := m.owner(ctx, origin, sessionToken)
	if err != nil {
		return false, err
	}
	on, err := m.available(ctx)
	switch {
	case err != nil:
		return false, err
	case !on:
		return false, ErrMaintainerOff
	}
	id, err := m.queue(ctx, TriggerOwner, asker)
	if errors.Is(err, ErrRunInProgress) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := m.p.Queue.Enqueue(ctx, id); err != nil {
		return false, fmt.Errorf("queueing the maintainer's job: %w", err)
	}
	return true, nil
}

// Weekly is the weekly schedule's turn: it queues a run of its own and
// works it. With the maintainer off it does nothing at all, and clears a
// standing "failing" item; while another run is queued or running it
// does nothing either. It returns the ID of the run it worked, empty when
// it ran none.
func (m *Maintainer) Weekly(ctx context.Context) (string, error) {
	on, err := m.available(ctx)
	if err != nil {
		return "", err
	}
	if !on {
		if clearErr := m.p.Decisions.Clear(ctx, KindMaintainerFailing, FailingSubject); clearErr != nil {
			return "", fmt.Errorf("clearing the maintainer's failing item: %w", clearErr)
		}
		return "", nil
	}
	id, err := m.queue(ctx, TriggerWeekly, "")
	if errors.Is(err, ErrRunInProgress) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, m.Work(ctx, id)
}

// Work works the queued run id: it computes the weights and drift, sends the
// figures to the maintainer's model through the router, checks the answer
// strictly and stores the note and, when it passed every rule, the proposal,
// raising its item in the Inbox. A run that has ended is not run again;
// one that already left a proposal has its Inbox item made sure of, so a
// job retried after a failure to raise it still does. Every ending is a
// final state with its audit entry: the router's refusals and the model's
// failures end the run without a note, never with part of one.
func (m *Maintainer) Work(ctx context.Context, runID string) error {
	run, started, err := m.p.Runs.StartRun(ctx, runID, m.p.Now().UTC())
	if err != nil {
		return fmt.Errorf("starting maintainer run: %w", err)
	}
	if !started {
		return m.settle(ctx, runID)
	}
	return m.finish(ctx, m.attempt(ctx, &run))
}

// Proposal returns the proposal id with the figures behind it, for the
// owner.
func (m *Maintainer) Proposal(ctx context.Context, origin, sessionToken, id string) (StoredProposal, error) {
	if _, err := m.owner(ctx, origin, sessionToken); err != nil {
		return StoredProposal{}, err
	}
	p, ok, err := m.p.Proposals.Proposal(ctx, id)
	if err != nil {
		return StoredProposal{}, fmt.Errorf("loading the proposal: %w", err)
	}
	if !ok {
		return StoredProposal{}, ErrUnknownProposal
	}
	return p, nil
}

// Apply changes the model portfolio's targets to proposal id's, for the
// owner: nothing is traded. The weights are checked again against the model
// as it is now (portfolio.Proposal.Apply): a model the owner changed since is
// a refused entry and an error wrapping portfolio.ErrStaleProposal, and
// nothing changes. A decided proposal is ErrProposalDecided. The audit entry
// is written inside the transaction that changes the targets, so a change
// the log can't record is not made.
func (m *Maintainer) Apply(ctx context.Context, origin, sessionToken, id string) (StoredProposal, error) {
	actor, err := m.owner(ctx, origin, sessionToken)
	if err != nil {
		return StoredProposal{}, err
	}
	at := m.p.Now().UTC()
	err = m.p.Proposals.ApplyProposal(ctx, id, at,
		func(model *portfolio.Model, p *portfolio.Proposal) (portfolio.Model, error) {
			return p.Apply(model)
		},
		func(ctx context.Context) error {
			return m.p.Audit.Record(ctx, &AuditRecord{Event: AuditApplied, Outcome: AuditOK, RunID: id, Actor: actor, At: at})
		})
	if errors.Is(err, portfolio.ErrStaleProposal) || errors.Is(err, portfolio.ErrInvalidProposal) {
		// The refusal stands whether or not it was recorded; a failure to
		// record it is joined to it.
		if recErr := m.p.Audit.Record(ctx, &AuditRecord{Event: AuditApplied, Outcome: AuditRefused, RunID: id, Actor: actor, At: at}); recErr != nil {
			err = errors.Join(err, recErr)
		}
		return StoredProposal{}, fmt.Errorf("applying the proposal: %w", err)
	}
	if err != nil {
		return StoredProposal{}, fmt.Errorf("applying the proposal: %w", err)
	}
	return m.settled(ctx, id)
}

// Dismiss closes proposal id for the owner and leaves the model portfolio as
// it is; its entry is written inside the transaction, like Apply's.
func (m *Maintainer) Dismiss(ctx context.Context, origin, sessionToken, id string) (StoredProposal, error) {
	actor, err := m.owner(ctx, origin, sessionToken)
	if err != nil {
		return StoredProposal{}, err
	}
	at := m.p.Now().UTC()
	err = m.p.Proposals.DismissProposal(ctx, id, at, func(ctx context.Context) error {
		return m.p.Audit.Record(ctx, &AuditRecord{Event: AuditDismissed, Outcome: AuditOK, RunID: id, Actor: actor, At: at})
	})
	if err != nil {
		return StoredProposal{}, fmt.Errorf("dismissing the proposal: %w", err)
	}
	return m.settled(ctx, id)
}
