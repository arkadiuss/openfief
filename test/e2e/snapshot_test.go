//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The upgrade run compares what the published release stored with what the upgraded install holds (D-161). It reads
// the database over the primary's local socket, as the connector test's dump does, and compares by content hash:
// every row of every table becomes the md5 of its text, so the test keeps and prints hashes and counts, never a
// value, and a row that changed by one digit is a different hash.
//
// A row of the database is stored data, derived data or ephemeral. Stored data is compared as it was: the same rows,
// unchanged, nothing added or lost. The exceptions are named here, each with why:
//   - attention.* is append-only and the new core's first checks may raise more items and add changes to the old
//     ones, so the published rows must all still be there (grows), and more may be.
//   - What the valuation rules derive is rewritten when the rules' version changes, which the new core does on its
//     first start by revaluing every account (valuation.SourceRulesVersion, the jurisdiction pack's version, D-161
//     counts derived values as recomputed, not migrated): holding_valuations' net, liquidation and tax columns,
//     account_status' revalued_at, and the settings row rules_version. Everything else in those tables is compared,
//     the rows' keys and market values included, so a lost or moved valuation still fails.
//   - ingest.tickers lost three columns to the migration that moved a ticker's confirmation to its institution
//     (ingest 0015, D-153): isin, currency and confirmed_at now live in ingest.ticker_confirmations, which
//     checkTickerConfirmations compares with what the seed confirmed. Every other column of the table is compared.
//   - Sessions, ceremonies and enrollment codes are ephemeral, the revaluation queue is a work queue, and the
//     migrations' own bookkeeping records the new migrations.
//
// snapshotSchemas are the schemas the contexts own; River's tables and the migrations' are left out.
func snapshotSchemas() []string {
	return []string{"attention", "demo", "identity", "ingest", "ledger", "portfolio", "valuation"}
}

// ephemeralTables are not compared at all.
func ephemeralTables() []string {
	return []string{"identity.sessions", "identity.ceremonies", "identity.enrollment_codes", "valuation.dirty"}
}

// derivedColumns are left out of a table's rows: what the valuation rules derive, and what a migration moved.
func derivedColumns() map[string][]string {
	return map[string][]string{
		"valuation.holding_valuations": {"net", "liquidation", "tax_cost", "sale_cost", "net_missing", "liquidation_missing"},
		"valuation.account_status":     {"revalued_at"},
		"ingest.tickers":               {"isin", "currency", "confirmed_at"},
	}
}

// derivedRows leaves a row out of a table: the recorded rules version.
func derivedRows() map[string]string {
	return map[string]string{"valuation.settings": `key <> 'rules_version'`}
}

// growingTables only gain rows: every row of before must be in after, and more may be.
func growingTables() []string {
	return []string{"attention.items", "attention.changes", "attention.outbox", "attention.published"}
}

// tableSpec is what to hash of one table: the columns, and a row filter, or "" for every row.
type tableSpec struct {
	name, columns, where string
}

// query returns the SELECT that prints the table's rows as their hashes, one line each: name|hash.
func (s tableSpec) query() string {
	q := fmt.Sprintf(`SELECT '%s', md5(r::text) FROM (SELECT %s FROM %s`, s.name, s.columns, s.name)
	if s.where != "" {
		q += " WHERE " + s.where
	}
	return q + ") r;\n"
}

// snapshot is the hashed rows of the tables it names.
type snapshot struct {
	tables []tableSpec
	// rows maps a table to its rows' hashes and how many rows have each.
	rows map[string]map[string]int
}

// count returns the number of rows the snapshot holds of table.
func (s *snapshot) count(table string) int {
	n := 0
	for _, c := range s.rows[table] {
		n += c
	}
	return n
}

// surplus returns how many rows of a have no equal row in b: the rows a holds more of than b.
func surplus(a, b map[string]int) int {
	n := 0
	for hash, count := range a {
		if d := count - b[hash]; d > 0 {
			n += d
		}
	}
	return n
}

// compare returns what differs between s, taken before the upgrade, and after: rows of s that are gone or changed
// and, but for the growing tables, rows that were added or changed. It prints counts only.
func (s *snapshot) compare(after *snapshot) []string {
	var diffs []string
	for _, spec := range s.tables {
		was, now := s.rows[spec.name], after.rows[spec.name]
		if lost := surplus(was, now); lost > 0 {
			diffs = append(diffs, fmt.Sprintf("%s: %d of %d rows are gone or changed", spec.name, lost, s.count(spec.name)))
		}
		if added := surplus(now, was); added > 0 && !slices.Contains(growingTables(), spec.name) {
			diffs = append(diffs, fmt.Sprintf("%s: %d rows were added or changed (%d before, %d after)", spec.name, added, s.count(spec.name), after.count(spec.name)))
		}
	}
	return diffs
}

// same reports whether s and other hold the same rows, but for rows the growing tables gained.
func (s *snapshot) same(other *snapshot) bool {
	return len(s.compare(other)) == 0 && len(other.compare(s)) == 0
}

// attentionCounts is how many attention items the published release's core has raised, and how many events its
// outbox holds and has published.
type attentionCounts struct {
	items, outbox, published int
}

// attention reads the counts.
func attention(ctx context.Context, k *kube) (attentionCounts, error) {
	rows, err := queryLinesCtx(ctx, k, `SELECT (SELECT count(*) FROM attention.items) || '|' || (SELECT count(*) FROM attention.outbox)
		|| '|' || (SELECT count(*) FROM attention.published)`)
	if err != nil {
		return attentionCounts{}, err
	}
	var c attentionCounts
	if _, scanErr := fmt.Sscanf(strings.Join(rows, ""), "%d|%d|%d", &c.items, &c.outbox, &c.published); scanErr != nil {
		return attentionCounts{}, fmt.Errorf("%w: %q: %w", errQuery, rows, scanErr)
	}
	return c, nil
}

// settler follows the seeded data until it is at rest.
type settler struct {
	tables []tableSpec
	last   *snapshot
	equal  int
}

// observe takes a snapshot and counts the ones in a row that equal it.
func (s *settler) observe(ctx context.Context, k *kube) error {
	if s.tables == nil {
		tables, err := listTables(ctx, k)
		if err != nil {
			return err
		}
		s.tables = tables
	}
	snap, err := takeSnapshot(ctx, k, s.tables)
	if err != nil {
		return fmt.Errorf("%w: %w", errRetry, err)
	}
	if s.last != nil && s.last.same(snap) {
		s.equal++
	} else {
		s.equal = 0
	}
	s.last = snap
	return nil
}

// check returns the wait's condition: done once the core's attention items are raised and published and
// stableSnapshots snapshots in a row were equal.
func (s *settler) check(k *kube) check {
	return func(ctx context.Context) (string, bool, error) {
		c, err := attention(ctx, k)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		state := fmt.Sprintf("%d attention items, %d events, %d published", c.items, c.outbox, c.published)
		if c.items == 0 || c.outbox != c.published {
			return state, false, nil
		}
		if err := s.observe(ctx, k); err != nil {
			return "", false, err
		}
		return fmt.Sprintf("%s; %d equal snapshots", state, s.equal), s.equal >= stableSnapshots-1, nil
	}
}

// settledSnapshot waits until the published release's core has raised its attention items, published them and the
// data has stopped changing (stableSnapshots equal snapshots in a row), and returns it. The core runs its checks
// once a minute, so a snapshot taken at once would race them.
func settledSnapshot(t *testing.T, k *kube) *snapshot {
	t.Helper()
	var s settler
	err := waitFor(t.Context(), settleTimeout, s.check(k))
	if err != nil {
		t.Fatalf("waiting for the seeded data to settle: %v", err)
	}
	return s.last
}

// listTables returns what to hash: every base table of the contexts' schemas, less the ephemeral ones, with the
// derived columns and rows left out.
func listTables(ctx context.Context, k *kube) ([]tableSpec, error) {
	lines, err := queryLinesCtx(ctx, k, listTablesQuery())
	if err != nil {
		return nil, err
	}
	return tablesFrom(lines)
}

// listTablesQuery lists the columns of every base table of the contexts' schemas, as schema.table|column.
func listTablesQuery() string {
	schemas := "'" + strings.Join(snapshotSchemas(), "', '") + "'"
	return `SELECT c.table_schema || '.' || c.table_name || '|' || c.column_name
		FROM information_schema.columns c JOIN information_schema.tables t USING (table_schema, table_name)
		WHERE t.table_type = 'BASE TABLE' AND c.table_schema IN (` + schemas + `) AND c.table_name <> 'schema_migrations'
		ORDER BY c.table_schema, c.table_name, c.ordinal_position`
}

// tablesFrom turns the lines of listTablesQuery into what to hash.
func tablesFrom(lines []string) ([]tableSpec, error) {
	var specs []tableSpec
	for _, line := range lines {
		table, column, ok := strings.Cut(line, "|")
		if !ok {
			return nil, fmt.Errorf("%w: %q", errQuery, line)
		}
		if slices.Contains(ephemeralTables(), table) || slices.Contains(derivedColumns()[table], column) {
			continue
		}
		if n := len(specs); n > 0 && specs[n-1].name == table {
			specs[n-1].columns += `, "` + column + `"`
			continue
		}
		specs = append(specs, tableSpec{name: table, columns: `"` + column + `"`, where: derivedRows()[table]})
	}
	return specs, nil
}

// takeSnapshot hashes the rows of tables in one database call.
func takeSnapshot(ctx context.Context, k *kube, tables []tableSpec) (*snapshot, error) {
	var script strings.Builder
	script.WriteString("SET timezone TO 'UTC';\n")
	for _, spec := range tables {
		script.WriteString(spec.query())
	}
	out, err := k.psql(ctx, script.String())
	if err != nil {
		return nil, err
	}
	snap := &snapshot{tables: tables, rows: make(map[string]map[string]int, len(tables))}
	for _, spec := range tables {
		snap.rows[spec.name] = map[string]int{}
	}
	for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		table, hash, ok := strings.Cut(line, "|")
		if !ok || snap.rows[table] == nil {
			return nil, fmt.Errorf("%w: %q", errQuery, line)
		}
		snap.rows[table][hash]++
	}
	return snap, nil
}

// psql runs script in psql, as the postgres user over the local socket of the openfief-db cluster's primary, against
// the openfief database, and returns what it prints without headers or alignment, columns joined by "|".
func (k *kube) psql(ctx context.Context, script string) (string, error) {
	pods, err := k.clients.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: databasePodSelector})
	if err != nil {
		return "", fmt.Errorf("%w: listing the database's pods: %w", errRetry, err)
	}
	for i := range pods.Items {
		p := &pods.Items[i]
		if p.DeletionTimestamp != nil || !podReady(p) {
			continue
		}
		out, execErr := k.exec(ctx, namespace, p.Name, "postgres", []byte(script),
			"psql", "-X", "-q", "-A", "-t", "-F", "|", "-v", "ON_ERROR_STOP=1", "-d", ownerDB)
		if execErr != nil {
			return "", execErr
		}
		return string(out), nil
	}
	return "", fmt.Errorf("%w: %w: %s", errRetry, errNoPrimary, databasePodSelector)
}

// queryLinesCtx returns the lines the one query prints.
func queryLinesCtx(ctx context.Context, k *kube, query string) ([]string, error) {
	out, err := k.psql(ctx, query+";\n")
	if err != nil {
		return nil, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// queryLines is queryLinesCtx for a test, failing it on an error.
func queryLines(t *testing.T, k *kube, query string) []string {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), kubectlTimeout)
	defer cancel()
	var lines []string
	err := waitFor(ctx, kubectlTimeout, func(ctx context.Context) (string, bool, error) {
		var qErr error
		lines, qErr = queryLinesCtx(ctx, k, query)
		return "querying", qErr == nil, qErr
	})
	if err != nil {
		t.Fatalf("querying the database: %v", err)
	}
	return lines
}

// queryOne returns the one line the query prints.
func queryOne(t *testing.T, k *kube, query string) string {
	t.Helper()
	lines := queryLines(t, k, query)
	if len(lines) != 1 {
		t.Fatalf("%v: %q answered %d lines, want 1", errQuery, query, len(lines))
	}
	return lines[0]
}
