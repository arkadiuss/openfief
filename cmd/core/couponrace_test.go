package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	"github.com/arkadiuss/openfief/internal/kernel"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// gate holds a connector call once it has started, until the test lets it
// go on: started receives once the call is in, and the call waits for
// release.
type gate struct {
	started chan struct{}
	release chan struct{}
}

func newGate() *gate { return &gate{started: make(chan struct{}), release: make(chan struct{})} }

// pass reports the call's start to g and waits for its release or ctx;
// a nil g lets the call through.
func (g *gate) pass(ctx context.Context) error {
	if g == nil {
		return nil
	}
	g.started <- struct{}{}
	select {
	case <-g.release:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("held at the gate: %w", ctx.Err())
	}
}

// gatedCoupons is couponConnector whose sync, after reading the request,
// and whose bond lot import can each be held at a gate, so a test can run
// a lot deletion and a sync of the same instance at the same time. It
// counts the syncs that reached it.
type gatedCoupons struct {
	couponConnector

	gmu    sync.Mutex
	syncs  int
	sync   *gate
	delete *gate
}

func (c *gatedCoupons) Sync(ctx context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	c.gmu.Lock()
	c.syncs++
	syncGate := c.sync
	c.gmu.Unlock()
	// The request, with the lots it was told about, is read before the gate.
	res, err := c.couponConnector.Sync(ctx, req)
	if gerr := syncGate.pass(ctx); gerr != nil {
		return nil, gerr
	}
	return res, err
}

func (c *gatedCoupons) ImportBondLots(ctx context.Context, req *connect.Request[connectorv1.ImportBondLotsRequest]) (*connect.Response[connectorv1.ImportBondLotsResponse], error) {
	c.gmu.Lock()
	deleteGate := c.delete
	c.gmu.Unlock()
	if err := deleteGate.pass(ctx); err != nil {
		return nil, err
	}
	return bondsConnector{}.ImportBondLots(ctx, req)
}

// syncsSeen returns how many syncs reached the connector.
func (c *gatedCoupons) syncsSeen() int {
	c.gmu.Lock()
	defer c.gmu.Unlock()
	return c.syncs
}

// result is what a call run in the background returned.
type result struct {
	out   string
	batch ingest.Batch
	err   error
}

// syncBonds syncs bonds-main as `core connector sync` does, over cl.
func syncBonds(ctx context.Context, dsn string, cl *cluster) result {
	cmd, err := parseArgs([]string{"connector", "sync", "--instance=bonds-main"}, func(k string) string {
		if k == dsnEnv {
			return dsn
		}
		return ""
	})
	if err != nil {
		return result{err: err}
	}
	var out bytes.Buffer
	err = withConnectorDeps(ctx, cmd.DSN, quiet(), func(d *connectorDeps) error {
		return connectorCommand(ctx, &cmd, d, cl, &out)
	})
	return result{out: out.String(), err: err}
}

// deleteBondLot deletes the bond account's lot that key names through the
// core's bond lot use case and the connector on port.
func deleteBondLot(ctx context.Context, dsn string, port int32, key ingest.BondLotKey) result {
	var res result
	err := withConnectorDeps(ctx, dsn, quiet(), func(d *connectorDeps) error {
		d.entries = &entryCaller{}
		d.entries.set(connectrpc.NewClient(connectrpc.H2CClient(), func(string) (string, error) {
			return fmt.Sprintf("http://127.0.0.1:%d", port), nil
		}, staticTokens{}))
		// bonds-main books for the treasury bond registry (D-182).
		d.entries.setDiscovery(bondsDiscovery{})
		res.batch, res.err = newBondLots(d, everyAccountOwner{}).Delete(ctx, localUI, "session", demo.BondsAccount, key)
		return nil
	})
	if err != nil {
		res.err = err
	}
	return res
}

// currentKeys returns the external IDs of the bond account's current
// transactions.
func currentKeys(t *testing.T, dsn string) []string {
	t.Helper()
	var keys []string
	err := withConnectorDeps(t.Context(), dsn, quiet(), func(d *connectorDeps) error {
		txs, err := d.ledger.CurrentTransactions(t.Context(), demo.BondsAccount)
		if err != nil {
			return fmt.Errorf("current transactions: %w", err)
		}
		for i := range txs {
			keys = append(keys, txs[i].Key.ExternalID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("reading the ledger: %v", err)
	}
	return keys
}

// bondRace is the bond account with its COI1028 lot bound to bonds-main on
// a fresh database, served by a gated connector.
type bondRace struct {
	dsn  string
	conn *gatedCoupons
	port int32
	cl   *cluster
	key  ingest.BondLotKey
}

// newBondRace sets up a bondRace.
func newBondRace(t *testing.T) bondRace {
	t.Helper()
	r := bondRace{dsn: emptyDatabase(t), conn: &gatedCoupons{}}
	r.port = serveConnector(t, r.conn)
	r.cl = fakeCluster(t, r.port, "bonds-main")
	saveDemoRegistry(t, r.dsn)
	seedBondLot(t, r.dsn)
	if _, err := connectorCmd(t, r.dsn, r.cl, "bind", "--instance=bonds-main", "--account="+string(demo.BondsAccount)); err != nil {
		t.Fatalf("connector bind: %v", err)
	}
	purchased, err := kernel.ParseDate("2024-10-10")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	r.key = ingest.BondLotKey{Series: coiTerms(t).Series, Purchased: purchased}
	return r
}

// TestBondLotDeletionDuringSync reproduces D-154's race through the real
// core and PostgreSQL: a sync has read the bond account's open lot and is
// about to book its coupons when the owner deletes the lot. Without the
// sync claim the deletion went through, and the sync then booked the two
// coupons of the deleted lot, cash that was never paid. Now the deletion
// is refused while the sync runs, and once the sync has booked them,
// deleting the lot deletes them with it.
func TestBondLotDeletionDuringSync(t *testing.T) {
	t.Parallel()
	r := newBondRace(t)
	dsn, conn, port, cl, key := r.dsn, r.conn, r.port, r.cl, r.key
	syncGate := newGate()
	conn.gmu.Lock()
	conn.sync = syncGate
	conn.gmu.Unlock()
	synced := make(chan result, 1)
	go func() { synced <- syncBonds(t.Context(), dsn, cl) }()
	<-syncGate.started

	if res := deleteBondLot(t.Context(), dsn, port, key); !errors.Is(res.err, ingestapp.ErrLotSyncing) {
		t.Errorf("deleting the lot while a sync that read it runs = %+v, want ErrLotSyncing", res)
	}
	close(syncGate.release)
	if res := <-synced; res.err != nil || !strings.Contains(res.out, "2 transactions, 2 new") {
		t.Fatalf("the sync = %q, %v; want the lot's two coupons booked", res.out, res.err)
	}

	res := deleteBondLot(t.Context(), dsn, port, key)
	if res.err != nil || res.batch.Imported.NewVersions != 3 {
		t.Fatalf("deleting the lot after the sync = %+v; want the opening lot and its two coupons superseded", res)
	}
	if keys := currentKeys(t, dsn); len(keys) != 0 {
		t.Errorf("after deleting the lot, current transactions %v; want none", keys)
	}
}

// TestSyncDuringBondLotDeletion checks the other order: while the owner's
// deletion of the lot runs, a sync of the instance is refused before it
// reads the lots, so it can't book coupons of the lot being deleted; the
// next sync, after the deletion, finds no open lot and books nothing.
func TestSyncDuringBondLotDeletion(t *testing.T) {
	t.Parallel()
	r := newBondRace(t)
	dsn, conn, port, cl, key := r.dsn, r.conn, r.port, r.cl, r.key
	deleteGate := newGate()
	conn.gmu.Lock()
	conn.delete = deleteGate
	conn.gmu.Unlock()
	deleted := make(chan result, 1)
	go func() { deleted <- deleteBondLot(t.Context(), dsn, port, key) }()
	<-deleteGate.started

	if res := syncBonds(t.Context(), dsn, cl); !errors.Is(res.err, ingestapp.ErrSyncRunning) {
		t.Errorf("a sync while the lot is being deleted = %q, %v; want ErrSyncRunning", res.out, res.err)
	}
	if n := conn.syncsSeen(); n != 0 {
		t.Errorf("%d syncs reached the connector during the deletion; want none", n)
	}
	close(deleteGate.release)
	if res := <-deleted; res.err != nil || res.batch.Imported.NewVersions != 1 {
		t.Fatalf("deleting the lot = %+v; want its opening lot superseded", res)
	}

	if res := syncBonds(t.Context(), dsn, cl); res.err != nil || !strings.Contains(res.out, "0 transactions") {
		t.Fatalf("the sync after the deletion = %q, %v; want nothing booked", res.out, res.err)
	}
	if keys := currentKeys(t, dsn); len(keys) != 0 {
		t.Errorf("after the deletion and a sync, current transactions %v; want none", keys)
	}
}
