package main

import (
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
)

// drainAudit runs the audit worker over every audit-entry job the outbox holds
// that is not finished, as the core's River client does in the background, and
// marks each job completed: the entries are in the log when it returns, without
// waiting for the client. The client may write the same entry meanwhile, which
// is the point of keying an entry by its job: it is written once.
func drainAudit(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	type job struct {
		id   int64
		args auditEntryArgs
	}
	rows, err := pool.Query(t.Context(), `
		SELECT id, args FROM river.river_job
		WHERE kind = $1 AND state IN ('available', 'retryable', 'scheduled', 'pending', 'running') ORDER BY id`, auditEntryArgs{}.Kind())
	if err != nil {
		t.Fatalf("listing the audit jobs: %v", err)
	}
	var jobs []job
	for rows.Next() {
		var (
			j   job
			raw []byte
		)
		if err := rows.Scan(&j.id, &raw); err != nil {
			t.Fatalf("reading an audit job: %v", err)
		}
		if err := json.Unmarshal(raw, &j.args); err != nil {
			t.Fatalf("decoding the audit job %d: %v", j.id, err)
		}
		jobs = append(jobs, j)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("listing the audit jobs: %v", err)
	}
	w := &auditWorker{log: identitypg.New(pool)}
	for i := range jobs {
		if err := w.write(t.Context(), jobs[i].id, &jobs[i].args); err != nil {
			t.Fatalf("writing the entry of job %d: %v", jobs[i].id, err)
		}
		if _, err := pool.Exec(t.Context(), `UPDATE river.river_job SET state = 'completed', finalized_at = now() WHERE id = $1`, jobs[i].id); err != nil {
			t.Fatalf("completing job %d: %v", jobs[i].id, err)
		}
	}
}
