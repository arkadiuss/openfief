package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	attentiondomain "github.com/arkadiuss/openfief/internal/attention"
	attentionapp "github.com/arkadiuss/openfief/internal/attention/app"
	attentionpg "github.com/arkadiuss/openfief/internal/attention/postgres"
	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// undeclaredConnector is depositConnector whose deliveries also carry a
// cross-check price, which its manifest doesn't declare: the fake cluster's
// Plugin resources declare nothing.
type undeclaredConnector struct {
	depositConnector
}

func (c undeclaredConnector) Sync(ctx context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	res, err := c.depositConnector.Sync(ctx, req)
	if err != nil {
		return nil, err
	}
	res.Msg.Delivery.CrossCheckPrices = []*connectorv1.Price{{
		Instrument: &connectorv1.Instrument{Isin: "PLTEST000010", Symbol: "TST"}, Date: "2026-10-27",
		Close: &connectorv1.Money{Amount: "50.30", Currency: "USD"},
	}}
	return res, nil
}

// storedAttentionItems returns the attention items stored in dsn.
func storedAttentionItems(t *testing.T, dsn string) []attentiondomain.Item {
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
	return items
}

// TestUndeclaredCrossCheckPricesRaiseAttention checks D-141 through the
// real core and PostgreSQL: a sync whose delivery carries cross-check
// prices the instance's manifest doesn't declare drops them, imports the
// rest of the delivery, and raises one open attention item for the
// instance, which a sync that drops them again leaves as it is.
func TestUndeclaredCrossCheckPricesRaiseAttention(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	port := serveConnector(t, undeclaredConnector{depositConnector{account: string(demo.IBKRAccount)}})
	cl := fakeCluster(t, port, "ibkr-main")
	saveDemoRegistry(t, dsn)
	if _, err := connectorCmd(t, dsn, cl, "bind", "--instance=ibkr-main", "--account="+string(demo.IBKRAccount)); err != nil {
		t.Fatalf("connector bind: %v", err)
	}
	for range 2 {
		out, err := connectorCmd(t, dsn, cl, "sync", "--instance=ibkr-main", "--mode=transactions")
		if err != nil || !strings.Contains(out, "1 transactions") || !strings.Contains(out, "1 problems") {
			t.Fatalf("connector sync = %q, %v; want the deposit imported and one problem", out, err)
		}
	}
	checkUndeclaredItem(t, storedAttentionItems(t, dsn))
}

// checkUndeclaredItem checks that items are one open item of ibkr-main's
// undeclared cross-check prices, raised once.
func checkUndeclaredItem(t *testing.T, items []attentiondomain.Item) {
	t.Helper()
	if len(items) != 1 {
		t.Fatalf("attention items %+v; want one", items)
	}
	it := items[0]
	if it.Kind != ingest.KindUndeclaredCrossCheckPrices || it.Subject != "ibkr-main" || it.State() != attentiondomain.StateOpen ||
		it.Current().Detail != ingest.UndeclaredCrossCheckPricesDetail || len(it.Changes) != 1 {
		t.Errorf("attention item %+v; want one open item of ibkr-main raised once", it)
	}
}
