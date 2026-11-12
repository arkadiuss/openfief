package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

var _ app.Auditor = (*Store)(nil)

// refJSON is how a reference is stored in the refs column.
type refJSON struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// AppendAudit adds e to the audit log, in the transaction ctx carries (see
// Atomically) or, without one, on its own. It validates e first and fails
// without writing when it is no record the domain accepts. The table gives
// the core's role INSERT and SELECT only (migration 0004).
func (s *Store) AppendAudit(ctx context.Context, e *identity.AuditEntry) error {
	if err := e.Validate(); err != nil {
		return fmt.Errorf("appending an audit entry: %w", err)
	}
	refs := make([]refJSON, 0, len(e.Refs))
	for _, r := range e.Refs {
		refs = append(refs, refJSON{Kind: string(r.Kind), ID: r.ID})
	}
	blob, err := json.Marshal(refs)
	if err != nil {
		return fmt.Errorf("appending an audit entry: %w", err)
	}
	err = s.db(ctx).QueryRow(ctx, `
		INSERT INTO identity.audit_log (at, actor_kind, actor_id, action, outcome, refs)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		e.At, string(e.Actor.Kind), e.Actor.ID, string(e.Action), string(e.Outcome), blob).Scan(&e.Seq)
	if err != nil {
		return fmt.Errorf("appending an audit entry (%s): %w", e.Action, err)
	}
	return nil
}

// AuditEntries returns the page of entries f selects, newest first, and the
// cursor of the next page (zero on the last). The filter must be valid.
func (s *Store) AuditEntries(ctx context.Context, f *identity.AuditFilter) (app.AuditPage, error) {
	if err := f.Validate(); err != nil {
		return app.AuditPage{}, fmt.Errorf("listing audit entries: %w", err)
	}
	size := f.PageSize()
	q := auditQuery(f)
	rows, err := s.db(ctx).Query(ctx, q.text, q.args...)
	if err != nil {
		return app.AuditPage{}, fmt.Errorf("listing audit entries: %w", err)
	}
	entries, err := pgx.CollectRows(rows, scanAudit)
	if err != nil {
		return app.AuditPage{}, fmt.Errorf("listing audit entries: %w", err)
	}
	page := app.AuditPage{Entries: entries}
	if len(entries) > size {
		page.Entries = entries[:size]
		page.Next = entries[size-1].Seq
	}
	return page, nil
}

// auditSQL is a statement and its arguments.
type auditSQL struct {
	text string
	args []any
}

// auditQuery returns the SELECT for the page f asks for. The conditions are
// fixed text with numbered placeholders; the values only ever travel as
// arguments.
func auditQuery(f *identity.AuditFilter) auditSQL {
	var (
		where []string
		args  []any
	)
	add := func(cond string, v any) {
		args = append(args, v)
		where = append(where, strings.ReplaceAll(cond, "$$", "$"+strconv.Itoa(len(args))))
	}
	if f.ActorKind != "" {
		add("actor_kind = $$", string(f.ActorKind))
	}
	if f.ActorID != "" {
		add("actor_id = $$", f.ActorID)
	}
	if f.Action != "" {
		add("action = $$", string(f.Action))
	}
	if !f.From.IsZero() {
		add("at >= $$", f.From)
	}
	if !f.To.IsZero() {
		add("at < $$", f.To)
	}
	if f.Before != 0 {
		add("id < $$", f.Before)
	}
	query := `SELECT id, at, actor_kind, actor_id, action, outcome, refs FROM identity.audit_log`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	// One more than the page, to know whether another follows.
	args = append(args, f.PageSize()+1)
	return auditSQL{text: query + " ORDER BY id DESC LIMIT $" + strconv.Itoa(len(args)), args: args}
}

// scanAudit reads a row of the audit_log columns AuditEntries selects.
func scanAudit(row pgx.CollectableRow) (identity.AuditEntry, error) {
	var (
		e                          identity.AuditEntry
		actorKind, action, outcome string
		blob                       []byte
		at                         time.Time
	)
	if err := row.Scan(&e.Seq, &at, &actorKind, &e.Actor.ID, &action, &outcome, &blob); err != nil {
		return identity.AuditEntry{}, fmt.Errorf("scanning an audit entry: %w", err)
	}
	var refs []refJSON
	if err := json.Unmarshal(blob, &refs); err != nil {
		return identity.AuditEntry{}, fmt.Errorf("%w: audit entry %d: refs", errBadRow, e.Seq)
	}
	e.At, e.Actor.Kind = at.UTC(), identity.ActorKind(actorKind)
	e.Action, e.Outcome = identity.AuditAction(action), identity.AuditOutcome(outcome)
	for _, r := range refs {
		e.Refs = append(e.Refs, identity.Ref{Kind: identity.RefKind(r.Kind), ID: r.ID})
	}
	if err := e.Validate(); err != nil {
		return identity.AuditEntry{}, fmt.Errorf("%w: audit entry %d: %w", errBadRow, e.Seq, err)
	}
	return e, nil
}
