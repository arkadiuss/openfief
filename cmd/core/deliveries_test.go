package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	attentiondomain "github.com/arkadiuss/openfief/internal/attention"
	attentionapp "github.com/arkadiuss/openfief/internal/attention/app"
	attentionpg "github.com/arkadiuss/openfief/internal/attention/postgres"
	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest"
	ingestkube "github.com/arkadiuss/openfief/internal/ingest/kube"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// termsConnector is depositConnector whose deliveries also carry one
// treasury bond series' terms.
type termsConnector struct {
	depositConnector
}

func (c termsConnector) Sync(ctx context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	res, err := c.depositConnector.Sync(ctx, req)
	if err != nil {
		return nil, err
	}
	res.Msg.Delivery.BondTerms = []*connectorv1.BondTerms{{
		Instrument: &connectorv1.Instrument{Isin: "PL0000000014", Symbol: "EDO0336"}, FirstPeriodRate: "0.0560", Margin: "0.02",
		EarlyRedemptionFee: &connectorv1.Money{Amount: "2.00", Currency: "PLN"},
	}}
	return res, nil
}

// approveBondTerms sets whether instance's Plugin resource says its grant
// approves the bond terms it delivers (status.effective.bondTerms).
func approveBondTerms(t *testing.T, cl *cluster, instance string, approved bool) {
	t.Helper()
	plugins := cl.dynamic.Resource(ingestkube.PluginResource()).Namespace(cl.namespace)
	u, err := plugins.Get(t.Context(), instance, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("reading the Plugin %s: %v", instance, err)
	}
	if err := unstructured.SetNestedField(u.Object, approved, "status", "effective", "bondTerms"); err != nil {
		t.Fatalf("setting status.effective.bondTerms: %v", err)
	}
	if _, err := plugins.Update(t.Context(), u, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("updating the Plugin %s: %v", instance, err)
	}
}

// storedTermSources returns the stored bond terms as "isin source".
func storedTermSources(t *testing.T, dsn string) []string {
	t.Helper()
	pool, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	defer pool.Close()
	rows, err := pool.Query(t.Context(), `SELECT isin || ' ' || source FROM valuation.bond_terms ORDER BY isin`)
	if err != nil {
		t.Fatalf("reading bond terms: %v", err)
	}
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatalf("reading bond terms: %v", err)
		}
		out = append(out, strings.TrimSpace(s))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("reading bond terms: %v", err)
	}
	return out
}

// deliveryItems returns the states of the stored attention items of
// D-179's kind, as "subject state".
func deliveryItems(t *testing.T, dsn string) []string {
	t.Helper()
	var items []attentiondomain.Item
	err := withConnectorDeps(t.Context(), dsn, quiet(), func(d *connectorDeps) error {
		var err error
		items, err = attentionapp.New(attentionpg.New(d.pool), nil, time.Now).List(t.Context())
		if err != nil {
			return fmt.Errorf("listing attention items: %w", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("reading attention items: %v", err)
	}
	var out []string
	for i := range items {
		if items[i].Kind == ingest.KindUndeclaredDeliveries {
			out = append(out, items[i].Subject+" "+items[i].State().String())
		}
	}
	return out
}

// syncUndeclared syncs bonds-main twice while its grant doesn't approve bond
// terms, and checks that each sync imports the deposit with one problem,
// that no terms are stored and that one attention item is open.
func syncUndeclared(t *testing.T, dsn string, cl *cluster) {
	t.Helper()
	for range 2 {
		out, err := connectorCmd(t, dsn, cl, "sync", "--instance=bonds-main", "--mode=transactions")
		if err != nil || !strings.Contains(out, "1 transactions") || !strings.Contains(out, "1 problems") {
			t.Fatalf("connector sync = %q, %v; want the deposit imported and one problem", out, err)
		}
	}
	if got := storedTermSources(t, dsn); len(got) != 0 {
		t.Errorf("undeclared bond terms were stored: %v", got)
	}
	if got := deliveryItems(t, dsn); len(got) != 1 || got[0] != "bonds-main open" {
		t.Errorf("attention items = %v, want one open for bonds-main", got)
	}
}

// TestUndeclaredDeliveriesRaiseAttention checks D-179 through the real core
// and PostgreSQL: a sync whose delivery carries bond terms its instance's
// grant doesn't approve drops them, imports the rest, and keeps one open
// attention item for the instance however often it happens; once the
// grant approves them, the next sync stores them with the instance as
// their source and resolves the item.
func TestUndeclaredDeliveriesRaiseAttention(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	port := serveConnector(t, termsConnector{depositConnector{account: string(demo.IBKRAccount)}})
	cl := fakeCluster(t, port, "bonds-main")
	approveBondTerms(t, cl, "bonds-main", false)
	saveDemoRegistry(t, dsn)
	if _, err := connectorCmd(t, dsn, cl, "bind", "--instance=bonds-main", "--account="+string(demo.IBKRAccount)); err != nil {
		t.Fatalf("connector bind: %v", err)
	}
	syncUndeclared(t, dsn, cl)

	approveBondTerms(t, cl, "bonds-main", true)
	if out, err := connectorCmd(t, dsn, cl, "sync", "--instance=bonds-main", "--mode=transactions"); err != nil {
		t.Fatalf("connector sync once approved = %q, %v", out, err)
	}
	if got := storedTermSources(t, dsn); len(got) != 1 || got[0] != "PL0000000014 bonds-main" {
		t.Errorf("stored bond terms = %v, want PL0000000014 from bonds-main", got)
	}
	if got := deliveryItems(t, dsn); len(got) != 1 || got[0] != "bonds-main resolved" {
		t.Errorf("attention items = %v, want bonds-main's resolved", got)
	}
}
