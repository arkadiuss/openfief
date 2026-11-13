package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
	"github.com/arkadiuss/openfief/internal/portfolio/postgres"
)

const (
	instA = "IETEST000026/ETFA"
	instB = "PLTEST000010/ETFB"

	run1 = "0000000000000001"
	run2 = "0000000000000002"
	run3 = "0000000000000003"
)

var errRecordDown = errors.New("audit down")

// clock is a fixed start the tests move forward.
func clock() time.Time { return time.Date(2026, 11, 13, 6, 0, 0, 0, time.UTC) }

func dec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// sixtyForty saves a model of two instruments and returns the store.
func sixtyForty(t *testing.T) (*postgres.Store, *pgxpool.Pool) {
	t.Helper()
	pool := newPool(t)
	store := postgres.New(pool)
	m := portfolio.Model{Enabled: true, Targets: []portfolio.Target{
		target(t, portfolio.TargetInstrument, instA, "", "60"),
		target(t, portfolio.TargetInstrument, instB, "", "40"),
	}}
	if err := store.SaveModel(t.Context(), m); err != nil {
		t.Fatal(err)
	}
	return store, pool
}

// queueAndStart queues and starts run id.
func queueAndStart(ctx context.Context, t *testing.T, s *postgres.Store, id string, trigger app.RunTrigger, at time.Time) {
	t.Helper()
	if err := s.QueueRun(ctx, &app.Run{ID: id, Trigger: trigger, At: at}); err != nil {
		t.Fatalf("QueueRun %s: %v", id, err)
	}
	if _, ok, err := s.StartRun(ctx, id, at); err != nil || !ok {
		t.Fatalf("StartRun %s = %v, %v", id, ok, err)
	}
}

// proposal is a valid proposal for sixtyForty: 55 and 45.
func proposal(t *testing.T) *portfolio.Proposal {
	t.Helper()
	return &portfolio.Proposal{Changes: []portfolio.Change{
		{Target: "instrument:" + instA, Before: dec(t, "60"), After: dec(t, "55"), Reason: "Trim.", Held: true, Current: dec(t, "63.1"), Drift: dec(t, "3.1")},
		{Target: "instrument:" + instB, Before: dec(t, "40"), After: dec(t, "45"), Reason: "Add.", Held: true, Current: dec(t, "20"), Drift: dec(t, "-20")},
	}}
}

func okFinish(id string, at time.Time, p *portfolio.Proposal) *app.RunFinish {
	return &app.RunFinish{RunID: id, State: app.RunOK, At: at, Model: "local/synth", Note: "A note.", Proposal: p}
}

// finishOK finishes run id with a note and p at at.
func finishOK(ctx context.Context, t *testing.T, s *postgres.Store, id string, at time.Time, p *portfolio.Proposal) []string {
	t.Helper()
	superseded, err := s.FinishRun(ctx, okFinish(id, at, p))
	if err != nil {
		t.Fatalf("FinishRun %s: %v", id, err)
	}
	return superseded
}

// proposalOf returns the first run's proposal as stored, or fails the test.
func proposalOf(ctx context.Context, t *testing.T, s *postgres.Store) (app.StoredProposal, bool) {
	t.Helper()
	p, ok, err := s.Proposal(ctx, run1)
	if err != nil {
		t.Fatalf("Proposal %s: %v", run1, err)
	}
	return p, ok
}

// pendingOf returns the IDs of the proposals waiting for the owner.
func pendingOf(ctx context.Context, t *testing.T, s *postgres.Store) []string {
	t.Helper()
	ids, err := s.PendingProposals(ctx)
	if err != nil {
		t.Fatalf("PendingProposals: %v", err)
	}
	return ids
}

// runOf returns run id as stored, or fails the test.
func runOf(ctx context.Context, t *testing.T, s *postgres.Store, id string) app.Run {
	t.Helper()
	r, ok, err := s.Run(ctx, id)
	if err != nil || !ok {
		t.Fatalf("Run %s = %v, %v", id, ok, err)
	}
	return r
}

// hasNote reports whether any run left a note.
func hasNote(ctx context.Context, t *testing.T, s *postgres.Store) bool {
	t.Helper()
	_, ok, err := s.LatestNote(ctx)
	if err != nil {
		t.Fatalf("LatestNote: %v", err)
	}
	return ok
}

// modelOf returns the stored model portfolio's weights.
func modelOf(ctx context.Context, t *testing.T, s *postgres.Store) []string {
	t.Helper()
	m, err := s.Model(ctx)
	if err != nil {
		t.Fatalf("Model: %v", err)
	}
	out := make([]string, 0, len(m.Targets))
	for i := range m.Targets {
		out = append(out, m.Targets[i].Weight.String())
	}
	return out
}

func TestOneRunIsQueuedOrRunningAtATime(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	at := clock()
	if err := store.QueueRun(ctx, &app.Run{ID: run1, Trigger: app.TriggerOwner, At: at, Actor: "00000000000000a1"}); err != nil {
		t.Fatalf("QueueRun: %v", err)
	}
	if err := store.QueueRun(ctx, &app.Run{ID: run2, Trigger: app.TriggerWeekly, At: at.Add(time.Minute)}); !errors.Is(err, app.ErrRunInProgress) {
		t.Fatalf("a second QueueRun while one is queued: %v, want ErrRunInProgress", err)
	}
	run, ok, err := store.StartRun(ctx, run1, at.Add(time.Minute))
	if err != nil || !ok || run.Trigger != app.TriggerOwner || run.Actor != "00000000000000a1" {
		t.Fatalf("StartRun = %+v, %v, %v", run, ok, err)
	}
	if err := store.QueueRun(ctx, &app.Run{ID: run2, Trigger: app.TriggerWeekly, At: at.Add(2 * time.Minute)}); !errors.Is(err, app.ErrRunInProgress) {
		t.Fatalf("QueueRun while one runs: %v, want ErrRunInProgress", err)
	}
	finishOK(ctx, t, store, run1, at.Add(3*time.Minute), nil)
	if err := store.QueueRun(ctx, &app.Run{ID: run2, Trigger: app.TriggerWeekly, At: at.Add(4 * time.Minute)}); err != nil {
		t.Errorf("QueueRun after the run ended: %v", err)
	}
}

func TestAFinalRunIsNeitherFinishedNorStartedAgain(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	at := clock()
	queueAndStart(ctx, t, store, run1, app.TriggerOwner, at)
	finishOK(ctx, t, store, run1, at.Add(time.Minute), nil)
	if _, err := store.FinishRun(ctx, okFinish(run1, at.Add(2*time.Minute), nil)); !errors.Is(err, app.ErrRunFinal) {
		t.Errorf("FinishRun of a final run: %v, want ErrRunFinal", err)
	}
	if _, ok, err := store.StartRun(ctx, run1, at); err != nil || ok {
		t.Errorf("StartRun of a final run = %v, %v; want false", ok, err)
	}
}

func TestRunStuckForHalfAnHourIsAbandoned(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	at := clock()
	queueAndStart(ctx, t, store, run1, app.TriggerWeekly, at)
	if err := store.QueueRun(ctx, &app.Run{ID: run2, Trigger: app.TriggerWeekly, At: at.Add(app.StaleAfter - time.Second)}); !errors.Is(err, app.ErrRunInProgress) {
		t.Fatalf("QueueRun just before the limit: %v, want ErrRunInProgress", err)
	}
	if err := store.QueueRun(ctx, &app.Run{ID: run2, Trigger: app.TriggerWeekly, At: at.Add(app.StaleAfter)}); err != nil {
		t.Fatalf("QueueRun at the limit: %v", err)
	}
	if got := runOf(ctx, t, store, run1); got.State != app.RunFailed || got.Class != app.ClassInternal {
		t.Errorf("the abandoned run = %+v; want failed/internal", got)
	}
}

func TestFinishStoresTheNoteWhole(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	queueAndStart(ctx, t, store, run1, app.TriggerOwner, clock())
	at := clock().Add(time.Minute)
	if superseded := finishOK(ctx, t, store, run1, at, proposal(t)); len(superseded) != 0 {
		t.Fatalf("FinishRun superseded %v", superseded)
	}
	note, ok, err := store.LatestNote(ctx)
	want := app.Note{RunID: run1, Text: "A note.", Model: "local/synth", At: at, Trigger: app.TriggerOwner}
	if err != nil || !ok || cmp.Diff(want, note) != "" {
		t.Errorf("LatestNote = %+v, %v, %v; want %+v", note, ok, err, want)
	}
}

func TestFinishStoresTheProposalWithItsFigures(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	queueAndStart(ctx, t, store, run1, app.TriggerOwner, clock())
	at := clock().Add(time.Minute)
	finishOK(ctx, t, store, run1, at, proposal(t))
	got, ok := proposalOf(ctx, t, store)
	if !ok {
		t.Fatal("no proposal stored")
	}
	want := app.StoredProposal{ID: run1, Status: app.ProposalPending, CreatedAt: at, Model: "local/synth", Note: "A note.", Proposal: *proposal(t)}
	if diff := cmp.Diff(want, got, cmp.Comparer(func(a, b kernel.Decimal) bool { return a.Equal(b) })); diff != "" {
		t.Errorf("Proposal (-want +got):\n%s", diff)
	}
	pending, err := store.PendingProposals(ctx)
	if err != nil || len(pending) != 1 || pending[0] != run1 {
		t.Errorf("PendingProposals = %v, %v", pending, err)
	}
}

func TestFinishSupersedesEarlierPendingProposals(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	at := clock()
	queueAndStart(ctx, t, store, run1, app.TriggerOwner, at)
	finishOK(ctx, t, store, run1, at.Add(time.Minute), proposal(t))
	queueAndStart(ctx, t, store, run2, app.TriggerWeekly, at.Add(time.Hour))
	superseded := finishOK(ctx, t, store, run2, at.Add(time.Hour+time.Minute), proposal(t))
	if len(superseded) != 1 || superseded[0] != run1 {
		t.Fatalf("FinishRun superseded %v; want the first", superseded)
	}
	if p, _ := proposalOf(ctx, t, store); p.Status != app.ProposalSuperseded {
		t.Errorf("first proposal = %s, want superseded", p.Status)
	}
	if pending := pendingOf(ctx, t, store); len(pending) != 1 || pending[0] != run2 {
		t.Errorf("pending = %v, want only the second", pending)
	}
}

func TestFinishStoresNothingWhenTheRecordFails(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	queueAndStart(ctx, t, store, run1, app.TriggerOwner, clock())
	fin := okFinish(run1, clock().Add(time.Minute), proposal(t))
	fin.Within = func(context.Context) error { return errRecordDown }
	if _, err := store.FinishRun(ctx, fin); !errors.Is(err, errRecordDown) {
		t.Fatalf("FinishRun error = %v, want the record's", err)
	}
	if hasNote(ctx, t, store) {
		t.Error("a note was stored although its record failed")
	}
	if _, ok := proposalOf(ctx, t, store); ok {
		t.Error("a proposal was stored although its record failed")
	}
	if run := runOf(ctx, t, store, run1); run.State != app.RunRunning {
		t.Errorf("the run is %s, want still running so its job can be retried", run.State)
	}
}

func TestRefusedAndFailedRunsLeaveNoNote(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	for i, fin := range []*app.RunFinish{
		{RunID: run1, State: app.RunRefused, Class: app.ClassDenied},
		{RunID: run2, State: app.RunFailed, Class: "unavailable"},
	} {
		at := clock().Add(time.Duration(i) * time.Hour)
		queueAndStart(ctx, t, store, fin.RunID, app.TriggerWeekly, at)
		fin.At = at.Add(time.Minute)
		if _, err := store.FinishRun(ctx, fin); err != nil {
			t.Fatalf("FinishRun %s: %v", fin.RunID, err)
		}
	}
	if hasNote(ctx, t, store) {
		t.Error("a run that left no note has one")
	}
	got, err := store.RecentWeekly(ctx, 5)
	if err != nil || cmp.Diff([]app.RunState{app.RunFailed, app.RunRefused}, got) != "" {
		t.Errorf("RecentWeekly = %v, %v; want failed, refused (newest first)", got, err)
	}
	latest, ok, err := store.LatestRun(ctx)
	if err != nil || !ok || latest.ID != run2 || latest.Class != "unavailable" {
		t.Errorf("LatestRun = %+v, %v, %v", latest, ok, err)
	}
}

func TestRecentWeeklyIgnoresOwnerRunsAndActiveOnes(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	triggers := map[string]app.RunTrigger{run1: app.TriggerWeekly, run2: app.TriggerOwner, run3: app.TriggerWeekly}
	for i, id := range []string{run1, run2, run3} {
		at := clock().Add(time.Duration(i) * time.Hour)
		queueAndStart(ctx, t, store, id, triggers[id], at)
		if id == run3 {
			break // still running
		}
		if _, err := store.FinishRun(ctx, &app.RunFinish{RunID: id, State: app.RunFailed, Class: "internal", At: at.Add(time.Minute)}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := store.RecentWeekly(ctx, 5)
	if err != nil || cmp.Diff([]app.RunState{app.RunFailed}, got) != "" {
		t.Errorf("RecentWeekly = %v, %v; want only the one final weekly run", got, err)
	}
}

// applyProposal is the use case's apply, as the store takes it.
func applyProposal(m *portfolio.Model, p *portfolio.Proposal) (portfolio.Model, error) {
	next, err := p.Apply(m)
	if err != nil {
		return portfolio.Model{}, fmt.Errorf("applying: %w", err)
	}
	return next, nil
}

// recorded is a within that succeeds and says it ran.
func recorded(ran *bool) func(context.Context) error {
	return func(context.Context) error {
		*ran = true
		return nil
	}
}

// pendingProposal queues, runs and finishes run1 with the valid proposal.
func pendingProposal(ctx context.Context, t *testing.T, s *postgres.Store) {
	t.Helper()
	queueAndStart(ctx, t, s, run1, app.TriggerOwner, clock())
	finishOK(ctx, t, s, run1, clock().Add(time.Minute), proposal(t))
}

func TestApplyProposalChangesTheTargetsInOneStep(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	pendingProposal(ctx, t, store)
	ran := false
	at := clock().Add(time.Hour)
	if err := store.ApplyProposal(ctx, run1, at, applyProposal, recorded(&ran)); err != nil {
		t.Fatalf("ApplyProposal: %v", err)
	}
	if !ran {
		t.Error("the decision was not recorded before the commit")
	}
	if diff := cmp.Diff([]string{"55", "45"}, modelOf(ctx, t, store)); diff != "" {
		t.Errorf("model after Apply (-want +got):\n%s", diff)
	}
	if m, err := store.Model(ctx); err != nil || !m.Enabled {
		t.Errorf("the model after Apply = %+v, %v; want it still on", m, err)
	}
	if p, _ := proposalOf(ctx, t, store); p.Status != app.ProposalApplied || !p.DecidedAt.Equal(at) {
		t.Errorf("proposal = %s decided %v, want applied at %v", p.Status, p.DecidedAt, at)
	}
}

func TestADecidedProposalCannotBeDecidedAgain(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	pendingProposal(ctx, t, store)
	ran := false
	at := clock().Add(time.Hour)
	if err := store.ApplyProposal(ctx, run1, at, applyProposal, recorded(&ran)); err != nil {
		t.Fatalf("ApplyProposal: %v", err)
	}
	if err := store.ApplyProposal(ctx, run1, at, applyProposal, recorded(&ran)); !errors.Is(err, app.ErrProposalDecided) {
		t.Errorf("a second Apply: %v, want ErrProposalDecided", err)
	}
	if err := store.DismissProposal(ctx, run1, at, recorded(&ran)); !errors.Is(err, app.ErrProposalDecided) {
		t.Errorf("Dismiss after Apply: %v, want ErrProposalDecided", err)
	}
	if err := store.ApplyProposal(ctx, "ffffffffffffffff", at, applyProposal, nil); !errors.Is(err, app.ErrUnknownProposal) {
		t.Errorf("Apply of an unknown proposal: %v, want ErrUnknownProposal", err)
	}
}

func TestApplyProposalChangesNothingWhenItCannotBeRecorded(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	pendingProposal(ctx, t, store)
	err := store.ApplyProposal(ctx, run1, clock().Add(time.Hour), applyProposal, func(context.Context) error { return errRecordDown })
	if !errors.Is(err, errRecordDown) {
		t.Fatalf("Apply with the record down: %v, want its error", err)
	}
	if diff := cmp.Diff([]string{"60", "40"}, modelOf(ctx, t, store)); diff != "" {
		t.Errorf("a decision that could not be recorded changed the targets (-want +got):\n%s", diff)
	}
	if p, _ := proposalOf(ctx, t, store); p.Status != app.ProposalPending {
		t.Errorf("proposal = %s, want pending", p.Status)
	}
}

func TestApplyProposalChangesNothingWhenItIsStale(t *testing.T) {
	t.Parallel()
	store, _ := sixtyForty(t)
	ctx := t.Context()
	pendingProposal(ctx, t, store)
	// The owner edits the model; the proposal is stale.
	edited := portfolio.Model{Enabled: true, Targets: []portfolio.Target{
		target(t, portfolio.TargetInstrument, instA, "", "50"), target(t, portfolio.TargetInstrument, instB, "", "50"),
	}}
	if err := store.SaveModel(ctx, edited); err != nil {
		t.Fatal(err)
	}
	ran := false
	at := clock().Add(time.Hour)
	if err := store.ApplyProposal(ctx, run1, at, applyProposal, recorded(&ran)); !errors.Is(err, portfolio.ErrStaleProposal) {
		t.Fatalf("Apply of a stale proposal: %v, want ErrStaleProposal", err)
	}
	if ran {
		t.Error("a refused Apply recorded a decision")
	}
	if diff := cmp.Diff([]string{"50", "50"}, modelOf(ctx, t, store)); diff != "" {
		t.Errorf("a refused Apply changed the targets (-want +got):\n%s", diff)
	}
	if p, _ := proposalOf(ctx, t, store); p.Status != app.ProposalPending {
		t.Errorf("proposal = %s, want pending: the owner dismisses a stale one", p.Status)
	}
	if err := store.DismissProposal(ctx, run1, at, recorded(&ran)); err != nil {
		t.Fatalf("Dismiss: %v", err)
	}
	if p, _ := proposalOf(ctx, t, store); p.Status != app.ProposalDismissed || p.DecidedAt.IsZero() {
		t.Errorf("proposal = %s, want dismissed", p.Status)
	}
}

func TestMaintainerTablesRefuseWhatTheRulesForbid(t *testing.T) {
	t.Parallel()
	_, pool := sixtyForty(t)
	bad := []struct{ name, sql string }{
		{"an ok run without a note", `INSERT INTO portfolio.maintainer_runs (id, trigger, state, updated_at, model) VALUES ('00000000000000e1', 'weekly', 'ok', now(), 'a/b')`},
		{"a failed run without a class", `INSERT INTO portfolio.maintainer_runs (id, trigger, state, updated_at) VALUES ('00000000000000e2', 'weekly', 'failed', now())`},
		{"a class outside the closed set", `INSERT INTO portfolio.maintainer_runs (id, trigger, state, updated_at, error_class) VALUES ('00000000000000e3', 'weekly', 'failed', now(), 'the provider said hello')`},
		{"an ID that is not 16 hex digits", `INSERT INTO portfolio.maintainer_runs (id, trigger, state, updated_at) VALUES ('nope', 'weekly', 'queued', now())`},
		{"an unknown trigger", `INSERT INTO portfolio.maintainer_runs (id, trigger, state, updated_at) VALUES ('00000000000000e4', 'cron', 'queued', now())`},
		{"a note of 2001 characters", `INSERT INTO portfolio.maintainer_runs (id, trigger, state, updated_at, model, note) VALUES ('00000000000000e5', 'weekly', 'ok', now(), 'a/b', repeat('a', 2001))`},
	}
	for _, tt := range bad {
		if _, err := pool.Exec(t.Context(), tt.sql); err == nil {
			t.Errorf("%s was stored", tt.name)
		}
	}
}

func TestMaintainerChangesRefuseWhatTheRulesForbid(t *testing.T) {
	t.Parallel()
	store, pool := sixtyForty(t)
	ctx := t.Context()
	pendingProposal(ctx, t, store)
	for _, sql := range []string{
		`INSERT INTO portfolio.maintainer_changes (run_id, position, target, before_weight, after_weight, held, reason) VALUES ('` + run1 + `', 9, 'index:X', 10, 10.005, false, 'x')`,
		`INSERT INTO portfolio.maintainer_changes (run_id, position, target, before_weight, after_weight, held, reason) VALUES ('` + run1 + `', 9, 'index:X', 10, 11, false, E'two\nlines')`,
	} {
		if _, err := pool.Exec(ctx, sql); err == nil {
			t.Errorf("a malformed change was stored: %s", sql)
		}
	}
}
