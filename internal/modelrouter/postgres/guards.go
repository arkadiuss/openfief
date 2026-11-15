package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
)

var (
	_ app.UsageStore    = (*Store)(nil)
	_ app.QuestionStore = (*Store)(nil)
)

// Spent returns what the calls of day cost in currency: the sum of the day's
// rows, zero for a day with none.
func (s *Store) Spent(ctx context.Context, day kernel.Date, currency kernel.Currency) (kernel.Decimal, error) {
	var text string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(cost), 0)::text FROM modelrouter.daily_usage WHERE day = $1::date AND currency = $2`,
		day.String(), currency.String()).Scan(&text)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("reading the day's usage: %w", err)
	}
	d, err := kernel.ParseDecimal(text)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("%w: a day's cost: %w", ErrCorrupt, err)
	}
	return d, nil
}

// Record adds one call's usage to its day's row for the model, in one
// statement, so calls at the same moment, in this process or another replica,
// all count.
func (s *Store) Record(ctx context.Context, u *app.UsageRecord) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO modelrouter.daily_usage (day, provider, model, currency, calls, input_tokens, output_tokens, cost)
		VALUES ($1::date, $2, $3, $4, 1, $5, $6, $7::numeric)
		ON CONFLICT (day, provider, model, currency) DO UPDATE SET
			calls = daily_usage.calls + 1,
			input_tokens = daily_usage.input_tokens + EXCLUDED.input_tokens,
			output_tokens = daily_usage.output_tokens + EXCLUDED.output_tokens,
			cost = daily_usage.cost + EXCLUDED.cost`,
		u.Day.String(), u.Model.Instance, u.Model.Model, u.Currency.String(),
		max(u.Usage.InputTokens, 0), max(u.Usage.OutputTokens, 0), u.Cost.String())
	if err != nil {
		return fmt.Errorf("recording the call's usage: %w", err)
	}
	return nil
}

// Admit decides whether the question q may go to a model and records it when
// it may. In one transaction it takes an advisory lock for the session and,
// after it, for the address (always in that order, so two questions can't
// wait on each other), reads each subject's questions of the last day, asks
// modelrouter.DecideQuestion about each, and inserts the question only when
// both allow it; so two questions of one subject at the same moment can't both
// take its last place, on this replica or another. Questions older than the
// longest window are deleted as it goes.
func (s *Store) Admit(ctx context.Context, q *app.QuestionAsked) (app.QuestionRuling, error) {
	var ruling app.QuestionRuling
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var admitErr error
		ruling, admitErr = admitIn(ctx, tx, q)
		return admitErr
	})
	if err != nil {
		return app.QuestionRuling{}, fmt.Errorf("admitting a question: %w", err)
	}
	return ruling, nil
}

// admitIn is Admit inside transaction tx.
func admitIn(ctx context.Context, tx pgx.Tx, q *app.QuestionAsked) (app.QuestionRuling, error) {
	key, err := addressKey(ctx, tx, q.Address)
	if err != nil {
		return app.QuestionRuling{}, err
	}
	if lockErr := lockSubject(ctx, tx, "session", q.Session); lockErr != nil {
		return app.QuestionRuling{}, lockErr
	}
	if key != "" {
		if lockErr := lockSubject(ctx, tx, "address", key); lockErr != nil {
			return app.QuestionRuling{}, lockErr
		}
	}
	since := q.At.Add(-modelrouter.QuestionWindow)
	if _, err = tx.Exec(ctx, "DELETE FROM modelrouter.ask_questions WHERE asked_at < $1", since); err != nil {
		return app.QuestionRuling{}, fmt.Errorf("forgetting old questions: %w", err)
	}
	ruling, err := rule(ctx, tx, q, key, since)
	if err != nil || !ruling.Allowed {
		return ruling, err
	}
	var stored *string
	if key != "" {
		stored = &key
	}
	_, err = tx.Exec(ctx, "INSERT INTO modelrouter.ask_questions (asked_at, session_id, address_key) VALUES ($1, $2, $3)",
		q.At, q.Session, stored)
	if err != nil {
		return app.QuestionRuling{}, fmt.Errorf("recording the question: %w", err)
	}
	return ruling, nil
}

// subjectQuery is one subject of the limits, a session or an address, and the
// column of ask_questions its questions are found by: a constant of this
// package, never a caller's text.
type subjectQuery struct {
	subject       app.QuestionSubject
	column, value string
}

// rule asks modelrouter.DecideQuestion about the question's session and, when
// its address is known, about its address, and says which one stopped it first.
func rule(ctx context.Context, tx pgx.Tx, q *app.QuestionAsked, key string, since time.Time) (app.QuestionRuling, error) {
	subjects := []subjectQuery{{app.SubjectSession, "session_id", q.Session}}
	if key != "" {
		subjects = append(subjects, subjectQuery{app.SubjectAddress, "address_key", key})
	}
	for _, sub := range subjects {
		times, err := questionTimes(ctx, tx, sub.column, sub.value, since)
		if err != nil {
			return app.QuestionRuling{}, err
		}
		if v := modelrouter.DecideQuestion(times, q.At, q.Limits); !v.Allowed {
			return app.QuestionRuling{Subject: sub.subject, ResetsAt: v.ResetsAt}, nil
		}
	}
	return app.QuestionRuling{Allowed: true}, nil
}

// addressKey returns the salted SHA-256 of address in hex, as the database
// computes it, or "" for an unknown address. The address itself is only ever a
// parameter of this one statement.
func addressKey(ctx context.Context, tx pgx.Tx, address string) (string, error) {
	if address == "" {
		return "", nil
	}
	var key string
	err := tx.QueryRow(ctx, `
		SELECT encode(sha256(convert_to($1 || salt, 'UTF8')), 'hex') FROM modelrouter.guard_salt`, address).Scan(&key)
	if err != nil {
		return "", fmt.Errorf("keying an address: %w", err)
	}
	return key, nil
}

// lockSubject takes the transaction-scoped advisory lock of one subject of the
// limits, kind and name hashed to the lock's key.
func lockSubject(ctx context.Context, tx pgx.Tx, kind, name string) error {
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 0))", "modelrouter.ask:"+kind+":"+name); err != nil {
		return fmt.Errorf("locking a subject of the limits: %w", err)
	}
	return nil
}

// questionTimes returns when the subject whose column is value asked since
// since. column is a constant of this package, never a caller's text.
func questionTimes(ctx context.Context, tx pgx.Tx, column, value string, since time.Time) ([]time.Time, error) {
	rows, err := tx.Query(ctx, "SELECT asked_at FROM modelrouter.ask_questions WHERE "+column+" = $1 AND asked_at >= $2", value, since)
	if err != nil {
		return nil, fmt.Errorf("reading a subject's questions: %w", err)
	}
	times, err := pgx.CollectRows(rows, pgx.RowTo[time.Time])
	if err != nil {
		return nil, fmt.Errorf("reading a subject's questions: %w", err)
	}
	return times, nil
}
