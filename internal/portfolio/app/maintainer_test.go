package app_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
)

// The failures the fakes make up; none of them names a note or a figure.
var (
	errBoom          = errors.New("a failure of no class")
	errReadHoldings  = errors.New("the holdings could not be read")
	errAttentionDown = errors.New("attention down")
)

// answerJSON is what a model answers: a note and changes of {target, current,
// proposed, reason}.
func answerJSON(t *testing.T, note string, changes ...[4]string) string {
	t.Helper()
	list := make([]map[string]string, 0, len(changes))
	for _, c := range changes {
		list = append(list, map[string]string{
			"target": c[0], "currentTargetPercent": c[1], "proposedTargetPercent": c[2], "reason": c[3],
		})
	}
	b, err := json.Marshal(map[string]any{"note": note, "changes": list})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// validChanges move 5 points from ETFA to ETFB: 55 + 45 = 100.
func validChanges() [][4]string {
	return [][4]string{
		{"instrument:" + instA, "60", "55", "Overweight and drifting up."},
		{"instrument:" + instB, "40", "45", "Underweight."},
	}
}

// ask has the owner ask and returns the ID of the run it queued.
func ask(t *testing.T, r *maintainerRig) string {
	t.Helper()
	started, err := r.m.Ask(t.Context(), origin, ownerToken)
	if err != nil || !started {
		t.Fatalf("Ask = %v, %v; want a queued run", started, err)
	}
	return r.queue.jobs[len(r.queue.jobs)-1]
}

// work works run id and fails the test on an error.
func work(t *testing.T, r *maintainerRig, id string) {
	t.Helper()
	if err := r.m.Work(t.Context(), id); err != nil {
		t.Fatalf("Work: %v", err)
	}
}

// proposedRun asks, has the model propose validChanges, works the run and
// returns its ID.
func proposedRun(t *testing.T, r *maintainerRig) string {
	t.Helper()
	r.asker.content = answerJSON(t, "Rebalance.", validChanges()...)
	id := ask(t, r)
	work(t, r, id)
	return id
}

// runOf returns the run id as stored.
func runOf(t *testing.T, r *maintainerRig, id string) app.Run {
	t.Helper()
	run, ok, err := r.store.Run(t.Context(), id)
	if err != nil || !ok {
		t.Fatalf("Run %s = %v, %v", id, ok, err)
	}
	return run
}

// overviewOf returns what the session of token is shown.
func overviewOf(t *testing.T, r *maintainerRig, token string) app.Overview {
	t.Helper()
	ov, err := r.m.Overview(t.Context(), origin, token)
	if err != nil {
		t.Fatalf("Overview as %s: %v", token, err)
	}
	return ov
}

// weightsOf lists the weights of m's targets.
func weightsOf(m *portfolio.Model) []string {
	out := make([]string, 0, len(m.Targets))
	for i := range m.Targets {
		out = append(out, m.Targets[i].Weight.String())
	}
	return out
}

func TestAskQueuesOneRunAtATime(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	id := ask(t, r)
	if len(r.queue.jobs) != 1 {
		t.Fatalf("jobs = %v, want one", r.queue.jobs)
	}
	started, err := r.m.Ask(t.Context(), origin, ownerToken)
	if err != nil || started || len(r.queue.jobs) != 1 {
		t.Errorf("a second Ask = %v, %v, jobs %v; want not started, no error and no new job", started, err, r.queue.jobs)
	}
	// Once the run has ended, the owner may ask again.
	r.asker.content = answerJSON(t, "A note.")
	work(t, r, id)
	if started, err = r.m.Ask(t.Context(), origin, ownerToken); err != nil || !started {
		t.Errorf("Ask after the run ended = %v, %v; want started", started, err)
	}
}

func TestAskIsTheOwnersAndNeedsTheMaintainerOn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		token string
		setup func(*maintainerRig)
		want  error
	}{
		{name: "a session that may view everything is refused", token: readerToken, want: app.ErrNotOwner},
		{name: "a limited session is refused", token: limitedToken, want: app.ErrNotOwner},
		{name: "no session", token: "nope", want: app.ErrNoSession},
		{name: "the switch or the toggle off", token: ownerToken, setup: func(r *maintainerRig) { r.gate.on = false }, want: app.ErrMaintainerOff},
		{name: "the model portfolio off", token: ownerToken, setup: func(r *maintainerRig) { r.store.model.Enabled = false }, want: app.ErrMaintainerOff},
		{name: "the model portfolio without targets", token: ownerToken, setup: func(r *maintainerRig) { r.store.model.Targets = nil }, want: app.ErrMaintainerOff},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := newRig(t)
			if tt.setup != nil {
				tt.setup(r)
			}
			started, err := r.m.Ask(t.Context(), origin, tt.token)
			if !errors.Is(err, tt.want) || started {
				t.Errorf("Ask = %v, %v; want %v", started, err, tt.want)
			}
			if len(r.queue.jobs) != 0 || len(r.store.runs) != 0 || len(r.asker.calls) != 0 {
				t.Errorf("a refused Ask queued %v, stored %d runs and called the model %d times", r.queue.jobs, len(r.store.runs), len(r.asker.calls))
			}
		})
	}
}

func TestWorkSendsTheFiguresOnceAsThePromptSays(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	id := ask(t, r)
	r.asker.content = answerJSON(t, "ETFA is five points over.")
	work(t, r, id)
	if len(r.asker.calls) != 1 {
		t.Fatalf("the model was called %d times, want once", len(r.asker.calls))
	}
	got := r.asker.calls[0]
	wantFigures := "Portfolio as of 2026-11-12. Weights are percent of the whole portfolio's value in PLN.\n" +
		"Targets, one per line: key | target weight | weight held now | drift in percentage points.\n" +
		"instrument:" + instA + " | 60 | 70 | +10\n" +
		"instrument:" + instB + " | 40 | 20 | -20\n" +
		"Held outside the targets (cash and other holdings): 10\n"
	if diff := cmp.Diff(wantFigures, got.Figures); diff != "" {
		t.Errorf("figures sent (-want +got):\n%s", diff)
	}
	if got.Schema != portfolio.ResponseSchema() || got.Actor != "00000000000000a1" {
		t.Errorf("schema or actor wrong: actor %q", got.Actor)
	}
	if strings.Contains(got.Instructions, "70") {
		t.Error("the instructions, which go as public data, hold a figure")
	}
}

func TestWorkStoresTheNoteWithItsAuditEntry(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	id := ask(t, r)
	r.asker.content = answerJSON(t, "ETFA is five points over.")
	work(t, r, id)
	note := overviewOf(t, r, readerToken).Note
	if note == nil || note.Text != "ETFA is five points over." || note.Model != "local/synth" || note.Trigger != app.TriggerOwner {
		t.Fatalf("the note = %+v, want the model's text from local/synth, asked by the owner", note)
	}
	if want := []string{"run:ok"}; !cmp.Equal(want, r.audit.events()) {
		t.Errorf("audit = %v, want %v", r.audit.events(), want)
	}
	if len(r.decisions.open) != 0 {
		t.Errorf("a note alone raised %v", r.decisions.open)
	}
}

// ending is one way a run ends, and what it leaves.
type ending struct {
	name         string
	content      string
	askErr       error
	holdingsErr  error
	zeroTotal    bool
	wantState    app.RunState
	wantClass    string
	wantNote     bool
	wantProposal bool
	wantAudit    []string
}

func TestWorkEndsEveryRunInAState(t *testing.T) {
	t.Parallel()
	dropped := answerJSON(t, "Rebalance.", [4]string{"instrument:" + instA, "60", "55", "Trim."})
	tests := []ending{
		{
			name: "a note and a valid proposal", content: answerJSON(t, "Rebalance.", validChanges()...),
			wantState: app.RunOK, wantNote: true, wantProposal: true, wantAudit: []string{"run:ok"},
		},
		{
			name: "a proposal that breaks a rule is dropped whole, the note stays", content: dropped,
			wantState: app.RunOK, wantNote: true, wantAudit: []string{"dropped:refused", "run:ok"},
		},
		{
			name: "an answer that is no note is a failed run", content: "not json", wantState: app.RunFailed,
			wantClass: app.ClassInvalidOutput, wantAudit: []string{"run:failed:invalid_output"},
		},
		{
			name: "the router refusing", askErr: &app.AskError{Class: app.ClassDenied}, wantState: app.RunRefused,
			wantClass: app.ClassDenied, wantAudit: []string{"run:refused:denied"},
		},
		{
			name: "every model down", askErr: &app.AskError{Class: "unavailable"}, wantState: app.RunFailed,
			wantClass: "unavailable", wantAudit: []string{"run:failed:unavailable"},
		},
		{
			name: "a request the provider rejected", askErr: &app.AskError{Class: "rejected"}, wantState: app.RunFailed,
			wantClass: "rejected", wantAudit: []string{"run:failed:rejected"},
		},
		{
			name: "an error of no class", askErr: errBoom, wantState: app.RunFailed,
			wantClass: app.ClassInternal, wantAudit: []string{"run:failed:internal"},
		},
		{
			name: "nothing valued yet sends nothing", holdingsErr: app.ErrNoHoldings, wantState: app.RunFailed,
			wantClass: app.ClassNoFigures, wantAudit: []string{"run:failed:no_figures"},
		},
		{
			name: "a portfolio worth nothing sends nothing", zeroTotal: true, wantState: app.RunFailed,
			wantClass: app.ClassNoFigures, wantAudit: []string{"run:failed:no_figures"},
		},
		{
			name: "a failed read of the holdings", holdingsErr: errReadHoldings, wantState: app.RunFailed,
			wantClass: app.ClassInternal, wantAudit: []string{"run:failed:internal"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkEnding(t, &tt)
		})
	}
}

// checkEnding works a run as e says and checks what it left.
func checkEnding(t *testing.T, e *ending) {
	t.Helper()
	r := newRig(t)
	r.asker.content, r.asker.err, r.holdings.err = e.content, e.askErr, e.holdingsErr
	if e.zeroTotal {
		r.holdings.held.Total = mustDec(t, "0")
	}
	id := ask(t, r)
	work(t, r, id)
	run := runOf(t, r, id)
	if run.State != e.wantState || run.Class != e.wantClass {
		t.Errorf("run = %s/%q, want %s/%q", run.State, run.Class, e.wantState, e.wantClass)
	}
	if (e.holdingsErr != nil || e.zeroTotal) && len(r.asker.calls) != 0 {
		t.Error("the model was called with no figures to send")
	}
	checkLeft(t, r, e, id)
}

// checkLeft checks the note, the proposal, the audit entries and the Inbox item
// run id left.
func checkLeft(t *testing.T, r *maintainerRig, e *ending, id string) {
	t.Helper()
	if _, has := r.store.notes[id]; has != e.wantNote {
		t.Errorf("a note stored = %v, want %v", has, e.wantNote)
	}
	if _, has := r.store.proposals[id]; has != e.wantProposal {
		t.Errorf("a proposal stored = %v, want %v", has, e.wantProposal)
	}
	if diff := cmp.Diff(e.wantAudit, r.audit.events()); diff != "" {
		t.Errorf("audit (-want +got):\n%s", diff)
	}
	if _, open := r.decisions.open[app.KindTargetProposal+"/"+id]; open != e.wantProposal {
		t.Errorf("a proposal item open = %v, want %v", open, e.wantProposal)
	}
}

func TestWorkWithTheMaintainerTurnedOffMeanwhileCallsNoModel(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	id := ask(t, r)
	r.gate.on = false
	work(t, r, id)
	if len(r.asker.calls) != 0 {
		t.Error("the model was called with the switch off")
	}
	if run := runOf(t, r, id); run.State != app.RunRefused || run.Class != app.ClassDenied {
		t.Errorf("run = %s/%q, want refused/denied", run.State, run.Class)
	}
}

func TestWorkFailsClosedWhenTheAuditLogCannotBeWritten(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.asker.content = answerJSON(t, "Rebalance.", validChanges()...)
	id := ask(t, r)
	r.audit.err = errAuditDown
	if err := r.m.Work(t.Context(), id); !errors.Is(err, errAuditDown) {
		t.Fatalf("Work error = %v, want the audit's", err)
	}
	if len(r.store.notes) != 0 || len(r.store.proposals) != 0 || len(r.decisions.open) != 0 {
		t.Errorf("a run that could not be recorded stored a note %v, a proposal %v or an item %v", r.store.notes, r.store.proposals, r.decisions.open)
	}
	// The job is retried once the log is back: the run it left running is taken over.
	r.audit.err = nil
	work(t, r, id)
	if _, ok := r.store.proposals[id]; !ok {
		t.Error("the retried run left no proposal")
	}
}

func TestWorkRetriedAfterTheItemFailedStillRaisesIt(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.asker.content = answerJSON(t, "Rebalance.", validChanges()...)
	id := ask(t, r)
	r.decisions.err = errAttentionDown
	if err := r.m.Work(t.Context(), id); err == nil {
		t.Fatal("Work succeeded with the Inbox down")
	}
	if p := r.store.proposals[id]; p == nil || p.Status != app.ProposalPending {
		t.Fatalf("the proposal was not kept pending: %+v", p)
	}
	r.decisions.err = nil
	work(t, r, id)
	if _, open := r.decisions.open[app.KindTargetProposal+"/"+id]; !open {
		t.Error("the retry did not raise the proposal's item")
	}
	if len(r.asker.calls) != 1 {
		t.Errorf("the retry asked the model again: %d calls", len(r.asker.calls))
	}
}

func TestANewProposalSupersedesThePendingOne(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	first := proposedRun(t, r)
	r.advance(time.Hour)
	second := proposedRun(t, r)
	if st := r.store.proposals[first].Status; st != app.ProposalSuperseded {
		t.Errorf("the first proposal is %s, want superseded", st)
	}
	if _, open := r.decisions.open[app.KindTargetProposal+"/"+first]; open {
		t.Error("the superseded proposal's item stays open")
	}
	if _, open := r.decisions.open[app.KindTargetProposal+"/"+second]; !open {
		t.Error("the new proposal has no item")
	}
	if got := overviewOf(t, r, ownerToken).Pending; got != 1 {
		t.Errorf("Overview pending = %d, want 1", got)
	}
}

func TestApplyChangesTheTargetsAndTradesNothing(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	id := proposedRun(t, r)
	r.audit.records = nil
	got, err := r.m.Apply(t.Context(), origin, ownerToken, id)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got.Status != app.ProposalApplied || got.DecidedAt.IsZero() {
		t.Errorf("proposal = %s decided %v, want applied", got.Status, got.DecidedAt)
	}
	if diff := cmp.Diff([]string{"55", "45"}, weightsOf(&r.store.model)); diff != "" {
		t.Errorf("targets after Apply (-want +got):\n%s", diff)
	}
	if want := []string{"applied:ok"}; !cmp.Equal(want, r.audit.events()) || r.audit.records[0].Actor != "00000000000000a1" {
		t.Errorf("audit = %v, want applied:ok by the owner's session", r.audit.events())
	}
	if _, open := r.decisions.open[app.KindTargetProposal+"/"+id]; open {
		t.Error("the applied proposal's item stays open")
	}
	if _, err := r.m.Apply(t.Context(), origin, ownerToken, id); !errors.Is(err, app.ErrProposalDecided) {
		t.Errorf("Apply twice error = %v, want ErrProposalDecided", err)
	}
}

func TestApplyOfAProposalTheModelOutgrewIsRefusedAndAudited(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	id := proposedRun(t, r)
	// The owner edits the model meanwhile.
	r.store.model.Targets[0].Weight = mustDec(t, "50")
	r.store.model.Targets[1].Weight = mustDec(t, "50")
	r.audit.records = nil
	if _, err := r.m.Apply(t.Context(), origin, ownerToken, id); !errors.Is(err, portfolio.ErrStaleProposal) {
		t.Fatalf("Apply error = %v, want ErrStaleProposal", err)
	}
	if got := r.store.model.Targets[0].Weight.String(); got != "50" {
		t.Errorf("a refused Apply changed a target to %s", got)
	}
	if want := []string{"applied:refused"}; !cmp.Equal(want, r.audit.events()) {
		t.Errorf("audit = %v, want %v", r.audit.events(), want)
	}
	if p := r.store.proposals[id]; p.Status != app.ProposalPending {
		t.Errorf("a refused Apply left the proposal %s, want pending so the owner can dismiss it", p.Status)
	}
}

// decideWith applies or dismisses proposal id as the owner, by kind.
func decideWith(t *testing.T, r *maintainerRig, kind, id string) error {
	t.Helper()
	ctx := t.Context()
	if kind == "apply" {
		if _, err := r.m.Apply(ctx, origin, ownerToken, id); err != nil {
			return fmt.Errorf("applying: %w", err)
		}
		return nil
	}
	if _, err := r.m.Dismiss(ctx, origin, ownerToken, id); err != nil {
		return fmt.Errorf("dismissing: %w", err)
	}
	return nil
}

func TestApplyAndDismissFailClosedOnTheAuditLog(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"apply", "dismiss"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			r := newRig(t)
			id := proposedRun(t, r)
			r.audit.err = errAuditDown
			if err := decideWith(t, r, kind, id); !errors.Is(err, errAuditDown) {
				t.Fatalf("error = %v, want the audit's", err)
			}
			if got := r.store.model.Targets[0].Weight.String(); got != "60" {
				t.Errorf("a decision the log could not record changed a target to %s", got)
			}
			if p := r.store.proposals[id]; p.Status != app.ProposalPending {
				t.Errorf("the proposal is %s, want pending", p.Status)
			}
		})
	}
}

func TestDismissLeavesTheModelAndIsAudited(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	id := proposedRun(t, r)
	r.audit.records = nil
	got, err := r.m.Dismiss(t.Context(), origin, ownerToken, id)
	if err != nil || got.Status != app.ProposalDismissed {
		t.Fatalf("Dismiss = %+v, %v", got, err)
	}
	if w := r.store.model.Targets[0].Weight.String(); w != "60" {
		t.Errorf("Dismiss changed a target to %s", w)
	}
	if want := []string{"dismissed:ok"}; !cmp.Equal(want, r.audit.events()) {
		t.Errorf("audit = %v, want %v", r.audit.events(), want)
	}
	if _, open := r.decisions.open[app.KindTargetProposal+"/"+id]; open {
		t.Error("the dismissed proposal's item stays open")
	}
}

func TestProposalCallsAreTheOwnersAlone(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	id := proposedRun(t, r)
	ctx := t.Context()
	for _, token := range []string{readerToken, limitedToken} {
		_, getErr := r.m.Proposal(ctx, origin, token, id)
		_, applyErr := r.m.Apply(ctx, origin, token, id)
		_, dismissErr := r.m.Dismiss(ctx, origin, token, id)
		for op, err := range map[string]error{"Proposal": getErr, "Apply": applyErr, "Dismiss": dismissErr} {
			if !errors.Is(err, app.ErrNotOwner) {
				t.Errorf("%s as %s: %v, want ErrNotOwner", op, token, err)
			}
		}
	}
	if _, err := r.m.Apply(ctx, origin, "nope", id); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("Apply with no session: %v, want ErrNoSession", err)
	}
	if _, err := r.m.Proposal(ctx, origin, ownerToken, "ffffffffffffffff"); !errors.Is(err, app.ErrUnknownProposal) {
		t.Errorf("Proposal of an unknown ID: %v, want ErrUnknownProposal", err)
	}
	if p := r.store.proposals[id]; p.Status != app.ProposalPending || r.store.model.Targets[0].Weight.String() != "60" {
		t.Error("a refused caller changed something")
	}
}

func TestOverviewByReader(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	proposedRun(t, r)
	owner := overviewOf(t, r, ownerToken)
	if !owner.Available || owner.Note == nil || owner.Pending != 1 {
		t.Errorf("owner overview = %+v; want available, a note, one pending", owner)
	}
	reader := overviewOf(t, r, readerToken)
	if reader.Note == nil || reader.Pending != 0 || reader.LastFailure != nil {
		t.Errorf("reader overview = %+v; want the note and none of the owner's", reader)
	}
	r.gate.on = false
	if off := overviewOf(t, r, ownerToken); off.Available || off.Note == nil {
		t.Errorf("overview with the switch off = %+v; want unavailable, the old note still shown", off)
	}
}

func TestOverviewRefusesWhoMayNotReadTheWholePortfolio(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	if _, err := r.m.Overview(t.Context(), origin, limitedToken); !errors.Is(err, app.ErrNotReader) {
		t.Errorf("limited overview error = %v, want ErrNotReader", err)
	}
	if _, err := r.m.Overview(t.Context(), origin, "nope"); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("overview with no session: %v, want ErrNoSession", err)
	}
}

func TestOverviewShowsRunningAndTheOwnersLastFailure(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	id := ask(t, r)
	if ov := overviewOf(t, r, ownerToken); !ov.Running {
		t.Fatalf("overview while queued = %+v; want running", ov)
	}
	r.asker.err = &app.AskError{Class: "unavailable"}
	work(t, r, id)
	ov := overviewOf(t, r, ownerToken)
	if ov.Running || ov.LastFailure == nil || ov.LastFailure.Class != "unavailable" || ov.Note != nil {
		t.Errorf("owner overview after a failure = %+v", ov)
	}
	if rd := overviewOf(t, r, readerToken); rd.LastFailure != nil {
		t.Error("a failure is shown to a reader who is not the owner")
	}
}

// weekly runs the schedule's turn and returns the run it worked.
func weekly(t *testing.T, r *maintainerRig) string {
	t.Helper()
	id, err := r.m.Weekly(t.Context(), true)
	if err != nil {
		t.Fatalf("Weekly: %v", err)
	}
	return id
}

func TestWeeklyRunsOnTheSchedule(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.asker.content = answerJSON(t, "Weekly note.", validChanges()...)
	if id := weekly(t, r); id == "" {
		t.Fatal("Weekly ran nothing")
	}
	ov := overviewOf(t, r, ownerToken)
	if ov.Note == nil || ov.Note.Trigger != app.TriggerWeekly || ov.Pending != 1 {
		t.Errorf("overview = %+v, want a weekly note and a pending proposal", ov)
	}
	if got := r.asker.calls[0].Actor; got != "" {
		t.Errorf("the schedule's call names the actor %q, want none (the core)", got)
	}
}

func TestWeeklyWithTheMaintainerOffDoesNothingAtAll(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.gate.on = false
	r.decisions.open[app.KindMaintainerFailing+"/"+app.FailingSubject] = "stale"
	if id := weekly(t, r); id != "" {
		t.Errorf("Weekly ran %s with the switch off", id)
	}
	if len(r.store.runs) != 0 || len(r.asker.calls) != 0 || len(r.audit.records) != 0 {
		t.Errorf("an off maintainer stored %d runs, called the model %d times, wrote %d audit records", len(r.store.runs), len(r.asker.calls), len(r.audit.records))
	}
	if len(r.decisions.open) != 0 {
		t.Errorf("a failing item stays open for a maintainer that is off: %v", r.decisions.open)
	}
}

func TestWeeklyDoesNothingWhileARunIsActive(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	ask(t, r)
	if id := weekly(t, r); id != "" {
		t.Errorf("Weekly ran %s while the owner's run was queued", id)
	}
	if len(r.asker.calls) != 0 || len(r.store.runs) != 1 {
		t.Errorf("Weekly called the model %d times and stored %d runs", len(r.asker.calls), len(r.store.runs))
	}
}

func TestWeeklyTakesOverARunThatWasLost(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	ask(t, r)
	r.advance(app.StaleAfter)
	r.asker.content = answerJSON(t, "Weekly note.")
	if id := weekly(t, r); id == "" {
		t.Error("Weekly stays blocked by a run that was lost")
	}
	if run := r.store.runs[0]; run.State != app.RunFailed || run.Class != app.ClassInternal {
		t.Errorf("the lost run = %s/%q, want failed/internal", run.State, run.Class)
	}
}

// failingItem is the key of the maintainer's failing item.
const failingItem = app.KindMaintainerFailing + "/" + app.FailingSubject

// failAttempt has the model fail in class and runs one attempt of the
// schedule's turn, which must report an *app.AttemptFailedError of that class.
func failAttempt(t *testing.T, r *maintainerRig, class string, final bool) {
	t.Helper()
	r.asker.err = &app.AskError{Class: class}
	_, err := r.m.Weekly(t.Context(), final)
	failed, ok := errors.AsType[*app.AttemptFailedError](err)
	if !ok || failed.Class != class || failed.Final != final {
		t.Fatalf("Weekly(final=%v) = %v, want an *AttemptFailedError of class %q", final, err, class)
	}
}

func TestTheFailingItemNeedsEveryAttemptOfTheWeekToFail(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	for attempt := 1; attempt < app.WeeklyAttempts; attempt++ {
		failAttempt(t, r, "unavailable", false)
		if _, open := r.decisions.open[failingItem]; open {
			t.Fatalf("attempt %d of %d raised the item", attempt, app.WeeklyAttempts)
		}
		r.advance(time.Hour)
	}
	failAttempt(t, r, "unavailable", true)
	detail, open := r.decisions.open[failingItem]
	if !open {
		t.Fatal("a week whose every attempt failed raised no item")
	}
	if !strings.Contains(detail, "unavailable") || strings.Contains(detail, "ETF") {
		t.Errorf("detail = %q, want the class and no figure or name", detail)
	}
}

func TestEveryAttemptIsKeptWithItsClass(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	classes := []string{"unavailable", "rejected", "unavailable"}
	for i, class := range classes {
		failAttempt(t, r, class, i == len(classes)-1)
		r.advance(time.Hour)
	}
	if len(r.store.runs) != len(classes) {
		t.Fatalf("%d runs kept, want one for each of the %d attempts", len(r.store.runs), len(classes))
	}
	for i, run := range r.store.runs {
		if run.Trigger != app.TriggerWeekly || run.State != app.RunFailed || run.Class != classes[i] {
			t.Errorf("attempt %d = %s/%s/%q, want weekly/failed/%q", i+1, run.Trigger, run.State, run.Class, classes[i])
		}
	}
	for _, rec := range r.audit.records {
		if rec.Event == app.AuditRun && rec.Class == "" {
			t.Errorf("an attempt's audit entry has no class: %+v", rec)
		}
	}
}

func TestAWeekThatSucceedsOnALaterAttemptRaisesNothing(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	failAttempt(t, r, "unavailable", false)
	r.advance(15 * time.Minute)
	r.asker.err, r.asker.content = nil, answerJSON(t, "Second try.")
	weekly(t, r)
	if _, open := r.decisions.open[failingItem]; open {
		t.Error("a week that left a note raised the failing item")
	}
	if ov := overviewOf(t, r, ownerToken); ov.Note == nil {
		t.Error("the later attempt left no note")
	}
}

func TestAnOwnersFailedRunDoesNotCount(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.asker.err = &app.AskError{Class: "unavailable"}
	for range app.WeeklyAttempts + 1 {
		work(t, r, ask(t, r))
		r.advance(time.Hour)
	}
	if _, open := r.decisions.open[failingItem]; open {
		t.Error("owner-asked failures raised the item")
	}
	// A scheduled week that fails in full still raises it, whatever the owner's runs did before.
	for attempt := 1; attempt <= app.WeeklyAttempts; attempt++ {
		failAttempt(t, r, "unavailable", attempt == app.WeeklyAttempts)
		r.advance(time.Hour)
	}
	if _, open := r.decisions.open[failingItem]; !open {
		t.Error("a failed week raised no item")
	}
}

func TestANoteClearsTheFailingItem(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	failAttempt(t, r, "unavailable", true)
	if _, open := r.decisions.open[failingItem]; !open {
		t.Fatal("a failed week raised no item")
	}
	// The owner's good run clears it.
	r.advance(time.Hour)
	r.asker.err, r.asker.content = nil, answerJSON(t, "Back.")
	work(t, r, ask(t, r))
	if _, open := r.decisions.open[failingItem]; open {
		t.Fatal("a run with a note left the failing item open")
	}
	// The next failed week raises it again, and the next scheduled success clears it.
	r.advance(7 * 24 * time.Hour)
	failAttempt(t, r, "unavailable", true)
	if _, open := r.decisions.open[failingItem]; !open {
		t.Fatal("a second failed week did not raise the item again")
	}
	r.advance(7 * 24 * time.Hour)
	r.asker.err = nil
	weekly(t, r)
	if _, open := r.decisions.open[failingItem]; open {
		t.Error("a scheduled run with a note left the failing item open")
	}
}

func TestRefusedWeeklyAttemptsCountAsFailures(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	for attempt := 1; attempt <= app.WeeklyAttempts; attempt++ {
		failAttempt(t, r, app.ClassDenied, attempt == app.WeeklyAttempts)
		r.advance(time.Hour)
	}
	if _, open := r.decisions.open[failingItem]; !open {
		t.Error("a week whose every attempt the router refused raised no item")
	}
	if run := r.store.runs[0]; run.State != app.RunRefused {
		t.Errorf("the first attempt = %s, want refused", run.State)
	}
}

func TestARetryAfterTheMaintainerWasTurnedOffRunsNothingAndClearsTheItem(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	failAttempt(t, r, "unavailable", false)
	r.gate.on = false
	r.decisions.open[failingItem] = "stale"
	calls := len(r.asker.calls)
	if id, err := r.m.Weekly(t.Context(), true); id != "" || err != nil {
		t.Errorf("Weekly with the maintainer off = %q, %v; want nothing", id, err)
	}
	if len(r.asker.calls) != calls || len(r.decisions.open) != 0 {
		t.Errorf("a retry of an off maintainer called the model or left an item: %v", r.decisions.open)
	}
}
