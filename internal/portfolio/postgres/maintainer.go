package postgres

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
)

var (
	_ app.RunStore      = (*Store)(nil)
	_ app.ProposalStore = (*Store)(nil)
)

// oneActiveRun is the unique index that lets one run be queued or running at
// a time.
const oneActiveRun = "maintainer_runs_one_active"

// uniqueViolation is PostgreSQL's error code for a unique index refusing a
// row.
const uniqueViolation = "23505"

// QueueRun records a new queued run, unless another is queued or running
// (app.ErrRunInProgress). A run still active after app.StaleAfter is
// abandoned first: closed as failed with class "internal".
func (s *Store) QueueRun(ctx context.Context, run *app.Run) error {
	if _, err := s.db.Exec(ctx, `
		UPDATE portfolio.maintainer_runs SET state = 'failed', error_class = 'internal', updated_at = $1
		WHERE state IN ('queued', 'running') AND updated_at <= $2`, run.At, run.At.Add(-app.StaleAfter)); err != nil {
		return fmt.Errorf("closing abandoned maintainer runs: %w", err)
	}
	var actor *string
	if run.Actor != "" {
		actor = &run.Actor
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO portfolio.maintainer_runs (id, trigger, state, actor, updated_at) VALUES ($1, $2, 'queued', $3, $4)`,
		run.ID, string(run.Trigger), actor, run.At)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == oneActiveRun {
		return app.ErrRunInProgress
	}
	if err != nil {
		return fmt.Errorf("queueing a maintainer run: %w", err)
	}
	return nil
}

// StartRun marks the run id as running and returns it; false when it has
// ended or doesn't exist.
func (s *Store) StartRun(ctx context.Context, id string, at time.Time) (app.Run, bool, error) {
	var (
		trigger string
		actor   *string
	)
	err := s.db.QueryRow(ctx, `
		UPDATE portfolio.maintainer_runs SET state = 'running', updated_at = $2
		WHERE id = $1 AND state IN ('queued', 'running') RETURNING trigger, actor`, id, at).Scan(&trigger, &actor)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.Run{}, false, nil
	}
	if err != nil {
		return app.Run{}, false, fmt.Errorf("starting maintainer run %s: %w", id, err)
	}
	run := app.Run{ID: id, Trigger: app.RunTrigger(trigger), State: app.RunRunning, At: at}
	if actor != nil {
		run.Actor = *actor
	}
	return run, true, nil
}

// runColumns are the columns of a run, in scanRun's order.
const runColumns = `id, trigger, state, updated_at, coalesce(error_class, ''), coalesce(actor, '')`

// scanRun reads one run row.
func scanRun(row pgx.Row) (app.Run, error) {
	var (
		r              app.Run
		trigger, state string
	)
	if err := row.Scan(&r.ID, &trigger, &state, &r.At, &r.Class, &r.Actor); err != nil {
		return app.Run{}, fmt.Errorf("reading a run: %w", err)
	}
	r.Trigger, r.State, r.At = app.RunTrigger(trigger), app.RunState(state), r.At.UTC()
	return r, nil
}

// Run returns the run id, and false when there is none.
func (s *Store) Run(ctx context.Context, id string) (app.Run, bool, error) {
	run, err := scanRun(s.db.QueryRow(ctx, "SELECT "+runColumns+" FROM portfolio.maintainer_runs WHERE id = $1", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return app.Run{}, false, nil
	}
	if err != nil {
		return app.Run{}, false, fmt.Errorf("loading maintainer run %s: %w", id, err)
	}
	return run, true, nil
}

// FinishRun stores how a run ended with its note and proposal in one
// transaction, runs fin.Within before it commits, and marks every earlier
// pending proposal superseded, returning their IDs.
func (s *Store) FinishRun(ctx context.Context, fin *app.RunFinish) ([]string, error) {
	var superseded []string
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		var finishErr error
		superseded, finishErr = finishInTx(ctx, tx, fin)
		return finishErr
	})
	if err != nil {
		return nil, fmt.Errorf("finishing maintainer run %s: %w", fin.RunID, err)
	}
	return superseded, nil
}

// finishInTx locks the active run, records it, stores its ending and its
// proposal, and returns the proposals it superseded.
func finishInTx(ctx context.Context, tx pgx.Tx, fin *app.RunFinish) ([]string, error) {
	var state string
	err := tx.QueryRow(ctx, "SELECT state FROM portfolio.maintainer_runs WHERE id = $1 FOR UPDATE", fin.RunID).Scan(&state)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, app.ErrRunFinal
	case err != nil:
		return nil, fmt.Errorf("locking the run: %w", err)
	case !app.RunState(state).Active():
		return nil, app.ErrRunFinal
	}
	if fin.Within != nil {
		if withinErr := fin.Within(ctx); withinErr != nil {
			return nil, fmt.Errorf("recording the run: %w", withinErr)
		}
	}
	if rowErr := finishRow(ctx, tx, fin); rowErr != nil {
		return nil, rowErr
	}
	if fin.Proposal == nil {
		return nil, nil
	}
	return storeProposal(ctx, tx, fin)
}

// finishRow writes the final state of the run.
func finishRow(ctx context.Context, tx pgx.Tx, fin *app.RunFinish) error {
	var class, model, note *string
	if fin.Class != "" {
		class = &fin.Class
	}
	if fin.State == app.RunOK {
		model, note = &fin.Model, &fin.Note
	}
	if _, err := tx.Exec(ctx, `
		UPDATE portfolio.maintainer_runs SET state = $2, error_class = $3, model = $4, note = $5, updated_at = $6 WHERE id = $1`,
		fin.RunID, string(fin.State), class, model, note, fin.At); err != nil {
		return fmt.Errorf("storing the run's ending: %w", err)
	}
	return nil
}

// storeProposal supersedes the pending proposals and stores fin's.
func storeProposal(ctx context.Context, tx pgx.Tx, fin *app.RunFinish) ([]string, error) {
	rows, err := tx.Query(ctx, `
		UPDATE portfolio.maintainer_proposals SET status = 'superseded' WHERE status = 'pending' RETURNING run_id`)
	if err != nil {
		return nil, fmt.Errorf("superseding the pending proposals: %w", err)
	}
	superseded, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("superseding the pending proposals: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO portfolio.maintainer_proposals (run_id, status, created_at) VALUES ($1, 'pending', $2)`, fin.RunID, fin.At); err != nil {
		return nil, fmt.Errorf("storing the proposal: %w", err)
	}
	batch := &pgx.Batch{}
	for i := range fin.Proposal.Changes {
		c := &fin.Proposal.Changes[i]
		var current, drift *string
		if c.Held {
			cur, dr := c.Current.String(), c.Drift.String()
			current, drift = &cur, &dr
		}
		batch.Queue(`
			INSERT INTO portfolio.maintainer_changes
				(run_id, position, target, before_weight, after_weight, held, current_weight, drift, reason)
			VALUES ($1, $2, $3, $4::numeric, $5::numeric, $6, $7::numeric, $8::numeric, $9)`,
			fin.RunID, i, c.Target, c.Before.String(), c.After.String(), c.Held, current, drift, c.Reason)
	}
	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		return nil, fmt.Errorf("storing the proposal's changes: %w", err)
	}
	slices.Sort(superseded)
	return superseded, nil
}

// LatestNote returns the note of the latest run that left one.
func (s *Store) LatestNote(ctx context.Context) (app.Note, bool, error) {
	var (
		n       app.Note
		trigger string
	)
	err := s.db.QueryRow(ctx, `
		SELECT id, note, model, updated_at, trigger FROM portfolio.maintainer_runs
		WHERE state = 'ok' ORDER BY updated_at DESC, id DESC LIMIT 1`).Scan(&n.RunID, &n.Text, &n.Model, &n.At, &trigger)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.Note{}, false, nil
	}
	if err != nil {
		return app.Note{}, false, fmt.Errorf("loading the latest note: %w", err)
	}
	n.Trigger, n.At = app.RunTrigger(trigger), n.At.UTC()
	return n, true, nil
}

// LatestRun returns the latest run, active or final.
func (s *Store) LatestRun(ctx context.Context) (app.Run, bool, error) {
	run, err := scanRun(s.db.QueryRow(ctx, "SELECT "+runColumns+" FROM portfolio.maintainer_runs ORDER BY updated_at DESC, id DESC LIMIT 1"))
	if errors.Is(err, pgx.ErrNoRows) {
		return app.Run{}, false, nil
	}
	if err != nil {
		return app.Run{}, false, fmt.Errorf("loading the latest run: %w", err)
	}
	return run, true, nil
}

// RecentWeekly returns the states of the latest limit final weekly runs,
// newest first.
func (s *Store) RecentWeekly(ctx context.Context, limit int) ([]app.RunState, error) {
	rows, err := s.db.Query(ctx, `
		SELECT state FROM portfolio.maintainer_runs
		WHERE trigger = 'weekly' AND state IN ('ok', 'refused', 'failed') ORDER BY updated_at DESC, id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("loading the latest weekly runs: %w", err)
	}
	states, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("loading the latest weekly runs: %w", err)
	}
	out := make([]app.RunState, 0, len(states))
	for _, st := range states {
		out = append(out, app.RunState(st))
	}
	return out, nil
}

// PendingProposals returns the IDs of the proposals waiting for the owner,
// oldest first.
func (s *Store) PendingProposals(ctx context.Context) ([]string, error) {
	rows, err := s.db.Query(ctx, "SELECT run_id FROM portfolio.maintainer_proposals WHERE status = 'pending' ORDER BY created_at, run_id")
	if err != nil {
		return nil, fmt.Errorf("loading the pending proposals: %w", err)
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("loading the pending proposals: %w", err)
	}
	return ids, nil
}

// Proposal returns the proposal id with its run's note and model.
func (s *Store) Proposal(ctx context.Context, id string) (app.StoredProposal, bool, error) {
	p := app.StoredProposal{ID: id}
	var (
		status  string
		decided *time.Time
	)
	err := s.db.QueryRow(ctx, `
		SELECT p.status, p.created_at, p.decided_at, r.model, coalesce(r.note, '')
		FROM portfolio.maintainer_proposals p JOIN portfolio.maintainer_runs r ON r.id = p.run_id
		WHERE p.run_id = $1`, id).Scan(&status, &p.CreatedAt, &decided, &p.Model, &p.Note)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.StoredProposal{}, false, nil
	}
	if err != nil {
		return app.StoredProposal{}, false, fmt.Errorf("loading proposal %s: %w", id, err)
	}
	p.Status, p.CreatedAt = app.ProposalStatus(status), p.CreatedAt.UTC()
	if decided != nil {
		p.DecidedAt = decided.UTC()
	}
	rows, err := s.db.Query(ctx, `
		SELECT target, before_weight::text, after_weight::text, held, coalesce(current_weight::text, ''), coalesce(drift::text, ''), reason
		FROM portfolio.maintainer_changes WHERE run_id = $1 ORDER BY position`, id)
	if err != nil {
		return app.StoredProposal{}, false, fmt.Errorf("loading proposal %s's changes: %w", id, err)
	}
	if p.Proposal.Changes, err = pgx.CollectRows(rows, scanChange); err != nil {
		return app.StoredProposal{}, false, fmt.Errorf("loading proposal %s's changes: %w", id, err)
	}
	return p, true, nil
}

// scanChange reads one change row.
func scanChange(row pgx.CollectableRow) (portfolio.Change, error) {
	var (
		c                             portfolio.Change
		before, after, current, drift string
	)
	if err := row.Scan(&c.Target, &before, &after, &c.Held, &current, &drift, &c.Reason); err != nil {
		return portfolio.Change{}, fmt.Errorf("reading a change: %w", err)
	}
	var errs [4]error
	c.Before, errs[0] = kernel.ParseDecimal(before)
	c.After, errs[1] = kernel.ParseDecimal(after)
	if c.Held {
		c.Current, errs[2] = kernel.ParseDecimal(current)
		c.Drift, errs[3] = kernel.ParseDecimal(drift)
	}
	if err := errors.Join(errs[:]...); err != nil {
		return portfolio.Change{}, fmt.Errorf("%w: a proposed change: %w", ErrCorrupt, err)
	}
	return c, nil
}

// lockPending locks proposal id's row and returns its status: pending, or
// app.ErrProposalDecided, or app.ErrUnknownProposal.
func lockPending(ctx context.Context, tx pgx.Tx, id string) error {
	var status string
	err := tx.QueryRow(ctx, "SELECT status FROM portfolio.maintainer_proposals WHERE run_id = $1 FOR UPDATE", id).Scan(&status)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return app.ErrUnknownProposal
	case err != nil:
		return fmt.Errorf("locking proposal %s: %w", id, err)
	case status != string(app.ProposalPending):
		return app.ErrProposalDecided
	}
	return nil
}

// decide marks proposal id with status at at.
func decide(ctx context.Context, tx pgx.Tx, id string, status app.ProposalStatus, at time.Time) error {
	if _, err := tx.Exec(ctx, "UPDATE portfolio.maintainer_proposals SET status = $2, decided_at = $3 WHERE run_id = $1", id, string(status), at); err != nil {
		return fmt.Errorf("marking proposal %s %s: %w", id, status, err)
	}
	return nil
}

// ApplyProposal applies the pending proposal id in one transaction: apply
// gets the model portfolio and the proposal and returns the model to store,
// within runs before the commit, and the proposal becomes applied at at.
func (s *Store) ApplyProposal(ctx context.Context, id string, at time.Time,
	apply func(*portfolio.Model, *portfolio.Proposal) (portfolio.Model, error), within func(context.Context) error,
) error {
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		return applyInTx(ctx, tx, id, at, apply, within)
	})
	if err != nil {
		return fmt.Errorf("applying proposal %s: %w", id, err)
	}
	return nil
}

// applyInTx locks the pending proposal id, makes the model it gives of the
// stored one, records the decision and marks the proposal applied.
func applyInTx(ctx context.Context, tx pgx.Tx, id string, at time.Time,
	apply func(*portfolio.Model, *portfolio.Proposal) (portfolio.Model, error), within func(context.Context) error,
) error {
	if err := lockPending(ctx, tx, id); err != nil {
		return err
	}
	inTx := &Store{db: tx, tx: tx}
	p, ok, err := inTx.Proposal(ctx, id)
	switch {
	case err != nil:
		return err
	case !ok:
		return app.ErrUnknownProposal
	}
	model, err := inTx.Model(ctx)
	if err != nil {
		return err
	}
	next, err := apply(&model, &p.Proposal)
	if err != nil {
		return err
	}
	if saveErr := inTx.SaveModel(ctx, next); saveErr != nil {
		return saveErr
	}
	if within != nil {
		if withinErr := within(ctx); withinErr != nil {
			return fmt.Errorf("recording the decision: %w", withinErr)
		}
	}
	return decide(ctx, tx, id, app.ProposalApplied, at)
}

// DismissProposal marks the pending proposal id dismissed at at, running
// within before the commit.
func (s *Store) DismissProposal(ctx context.Context, id string, at time.Time, within func(context.Context) error) error {
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if err := lockPending(ctx, tx, id); err != nil {
			return err
		}
		if within != nil {
			if err := within(ctx); err != nil {
				return fmt.Errorf("recording the decision: %w", err)
			}
		}
		return decide(ctx, tx, id, app.ProposalDismissed, at)
	})
	if err != nil {
		return fmt.Errorf("dismissing proposal %s: %w", id, err)
	}
	return nil
}
