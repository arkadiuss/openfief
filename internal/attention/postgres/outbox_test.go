package postgres_test

import (
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/postgres"
)

// outboxRow is one outbox event as stored: its whole payload and the
// subject in it.
type outboxRow struct {
	Payload, Subject string
}

// outboxRows reads every outbox row.
func outboxRows(t *testing.T, pool *pgxpool.Pool) []outboxRow {
	t.Helper()
	rows, err := pool.Query(t.Context(), `SELECT payload::text, payload->>'subject' FROM attention.outbox ORDER BY id`)
	if err != nil {
		t.Fatalf("reading the outbox: %v", err)
	}
	out, err := pgx.CollectRows(rows, pgx.RowToStructByPos[outboxRow])
	if err != nil {
		t.Fatalf("reading the outbox: %v", err)
	}
	return out
}

// TestOutboxHoldsNoBrokerAccountNumber checks D-121 at the outbox: the
// event of an unmapped broker account's item names only its connector
// instance, as read access sees it (attention.ReadSubject, D-165), so no
// outbox row holds the synthetic broker account number, while the item
// keeps its full subject for the owner.
func TestOutboxHoldsNoBrokerAccountNumber(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	s := postgres.New(pool)
	ctx := t.Context()
	const account = "U0000001"
	step := raise(attention.KindBrokerAccountUnmapped, "ibkr-main/"+account)
	id, err := s.Append(ctx, &step)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	rows := outboxRows(t, pool)
	if len(rows) != 1 {
		t.Fatalf("outbox holds %d rows, want 1", len(rows))
	}
	if r := rows[0]; strings.Contains(r.Payload, account) || r.Subject != "ibkr-main" {
		t.Errorf("outbox payload %s; want the subject ibkr-main and no broker account number", r.Payload)
	}
	if it, ok, err := s.Item(ctx, id); err != nil || !ok || it.Subject != step.Subject {
		t.Errorf("Item = %+v, %v, %v; want the full subject kept for the owner", it, ok, err)
	}
}
