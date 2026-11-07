package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	"github.com/arkadiuss/openfief/internal/portfolio"
	portfoliopg "github.com/arkadiuss/openfief/internal/portfolio/postgres"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// registryCOI is the synthetic registry ISIN the rekey connector's
// published data lists for COI1028 (D-214).
const registryCOI = "PL9999000018"

// rekeyConnector is the couponConnector that also delivers, as the
// treasury-bonds connector does, the registry ISIN of each instrument the
// sync request names as provisional whose series its synthetic published
// data lists (treasurybonds.RegistryISINs): COI1028's. It records what
// each sync named as provisional.
type rekeyConnector struct {
	couponConnector

	mu          sync.Mutex
	provisional [][]string
}

func (c *rekeyConnector) Sync(ctx context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	provisional, err := delivery.Provisional(req.Msg)
	if err != nil {
		return nil, fmt.Errorf("sync: %w", connect.NewError(connect.CodeInvalidArgument, err))
	}
	named := make([]string, 0, len(provisional))
	for _, in := range provisional {
		named = append(named, in.String())
	}
	c.mu.Lock()
	c.provisional = append(c.provisional, named)
	c.mu.Unlock()
	resp, err := c.couponConnector.Sync(ctx, req)
	if err != nil {
		return nil, err
	}
	s, err := pl.ParseSeries("COI1028")
	if err != nil {
		return nil, fmt.Errorf("sync: %w", connect.NewError(connect.CodeInternal, err))
	}
	published := &pl.Published{ISINs: []pl.PublishedISIN{{Series: s, ISIN: registryCOI}}}
	resp.Msg.Delivery.RegistryIsins = delivery.RegistryISINs(treasurybonds.RegistryISINs(published, provisional))
	return resp, nil
}

// TestSyncRekeysProvisionalBonds checks the migration of D-214 through the
// real core and Postgres: a bond lot booked under the ISIN earlier versions
// derived from its series code, as alpha installs hold it, with its terms,
// the owner's exemption class, display name and model target, is re-keyed
// to the registry's ISIN the next sync delivers, every one of them; its
// transactions keep their keys, so the coupons the sync after books are
// unchanged; and once re-keyed, no sync names it as provisional again or
// changes anything.
func TestSyncRekeysProvisionalBonds(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	conn := &rekeyConnector{}
	port := serveConnector(t, conn)
	cl := fakeCluster(t, port, "bonds-main")
	saveDemoRegistry(t, dsn)
	seedBondLot(t, dsn)
	from := provisionalCOI(t)
	seedProvisionalExtras(t, dsn, from)
	if _, err := connectorCmd(t, dsn, cl, "bind", "--instance=bonds-main", "--account="+string(demo.BondsAccount)); err != nil {
		t.Fatalf("connector bind: %v", err)
	}
	out, err := connectorCmd(t, dsn, cl, "sync", "--instance=bonds-main")
	if err != nil || !strings.Contains(out, "(transactions): 2 transactions, 2 new") {
		t.Fatalf("connector sync = %q, %v", out, err)
	}
	for range 2 {
		out, err = connectorCmd(t, dsn, cl, "sync", "--instance=bonds-main")
		if err != nil || !strings.Contains(out, "2 transactions, 0 new, 0 new versions, 2 unchanged") {
			t.Fatalf("connector sync after the re-key = %q, %v", out, err)
		}
	}
	conn.checkNamedOnce(t, from)
	checkRekeyed(t, dsn, from)
}

// checkNamedOnce checks that the first of three syncs named in as
// provisional, and the two after it nothing.
func (c *rekeyConnector) checkNamedOnce(t *testing.T, in ledger.Instrument) {
	t.Helper()
	c.mu.Lock()
	named := c.provisional
	c.mu.Unlock()
	want := [][]string{{in.String()}, {}, {}}
	if fmt.Sprint(named) != fmt.Sprint(want) {
		t.Errorf("the syncs named as provisional %v; want %v", named, want)
	}
}

// provisionalCOI is COI1028 under its provisional ISIN, as seedBondLot
// books it.
func provisionalCOI(t *testing.T) ledger.Instrument {
	t.Helper()
	in, err := treasurybonds.ProvisionalInstrument(coiTerms(t).Series)
	if err != nil {
		t.Fatalf("ProvisionalInstrument: %v", err)
	}
	return in
}

// seedProvisionalExtras stores what the owner and connectors may keep of
// in besides its lot and terms: its exemption class (D-132), a display
// name (D-197) and the model portfolio's target (D-52).
func seedProvisionalExtras(t *testing.T, dsn string, in ledger.Instrument) {
	t.Helper()
	err := withConnectorDeps(t.Context(), dsn, quiet(), func(d *connectorDeps) error {
		ctx := t.Context()
		vstore, err := valuationpg.NewStore(d.pool)
		if err != nil {
			return fmt.Errorf("setup: %w", err)
		}
		if err := vstore.SaveExemptionClass(ctx, in.ISIN.String(), jurisdiction.ExemptionClass("savings")); err != nil {
			return fmt.Errorf("setup: %w", err)
		}
		names := &ledgerapp.Names{Store: ledgerpg.New(d.pool)}
		if err := names.Deliver(ctx, "bonds-main", []ledger.InstrumentName{{ISIN: in.ISIN, Name: "COI1028 synthetic"}}); err != nil {
			return fmt.Errorf("setup: %w", err)
		}
		target := portfolio.Target{Kind: portfolio.TargetInstrument, Name: in.String(), Weight: kernel.NewDecimal(100, 0)}
		if err := portfoliopg.New(d.pool).SaveModel(ctx, portfolio.Model{Enabled: true, Targets: []portfolio.Target{target}}); err != nil {
			return fmt.Errorf("setup: %w", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seeding: %v", err)
	}
}

// checkRekeyed checks that nothing the core stores of from is left under
// it: every current transaction of the bond account posts to the
// registry's instrument under its old key, and its terms, exemption class,
// display name and model target moved with it.
func checkRekeyed(t *testing.T, dsn string, from ledger.Instrument) {
	t.Helper()
	isin, parseErr := ledger.ParseISIN(registryCOI)
	if parseErr != nil {
		t.Fatalf("ParseISIN: %v", parseErr)
	}
	to := ledger.Instrument{ISIN: isin, Symbol: from.Symbol}
	err := withConnectorDeps(t.Context(), dsn, quiet(), func(d *connectorDeps) error {
		if err := checkRekeyedLedger(t, d, to); err != nil {
			return err
		}
		return checkRekeyedStores(t, d, from, to)
	})
	if err != nil {
		t.Fatalf("checking: %v", err)
	}
}

// checkRekeyedLedger checks that every current transaction of the bond
// account posts to the registry's instrument alone, under its old key.
func checkRekeyedLedger(t *testing.T, d *connectorDeps, to ledger.Instrument) error {
	t.Helper()
	txs, err := d.ledger.CurrentTransactions(t.Context(), demo.BondsAccount)
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}
	keys := make([]string, 0, len(txs))
	for i := range txs {
		if !postsOnlyTo(&txs[i], to) {
			t.Errorf("%s posts to another instrument than %s", txs[i].Key, to)
		}
		keys = append(keys, txs[i].Key.ExternalID)
	}
	wantKeys := "lot:COI1028:2024-10-10,lot:COI1028:2024-10-10:coupon:2025-10-10,lot:COI1028:2024-10-10:coupon:2026-10-10"
	if strings.Join(keys, ",") != wantKeys {
		t.Errorf("current keys %v, want %s", keys, wantKeys)
	}
	return nil
}

// postsOnlyTo reports whether every holding tx posts to is in.
func postsOnlyTo(tx *ledger.Transaction, in ledger.Instrument) bool {
	for _, p := range tx.Postings {
		if p.Book.Kind == ledger.BookHolding && p.Book.Instrument != in {
			return false
		}
	}
	return true
}

// checkRekeyedStores checks that the terms, exemption class, display
// name and model target of from moved to to.
func checkRekeyedStores(t *testing.T, d *connectorDeps, from, to ledger.Instrument) error {
	t.Helper()
	ctx := t.Context()
	vstore, err := valuationpg.NewStore(d.pool)
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}
	terms, err := vstore.AllBondTerms(ctx)
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}
	classes, err := vstore.ExemptionClasses(ctx)
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}
	if _, ok := terms[to.ISIN]; !ok || len(terms) != 1 || classes[to.ISIN.String()] != "savings" || len(classes) != 1 {
		t.Errorf("terms %v and classes %v; want both under %s alone", terms, classes, to.ISIN)
	}
	return checkRekeyedNameAndModel(t, d, from, to)
}

// checkRekeyedNameAndModel checks that the display name and the model
// target of from moved to to.
func checkRekeyedNameAndModel(t *testing.T, d *connectorDeps, from, to ledger.Instrument) error {
	t.Helper()
	ctx := t.Context()
	names, err := (&ledgerapp.Names{Store: ledgerpg.New(d.pool)}).Current(ctx)
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}
	if names[to.ISIN] != "COI1028 synthetic" {
		t.Errorf("names %v; want %s named as %s was", names, to.ISIN, from.ISIN)
	}
	model, err := portfoliopg.New(d.pool).Model(ctx)
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}
	if len(model.Targets) != 1 || model.Targets[0].Name != to.String() {
		t.Errorf("model targets %v; want %s", model.Targets, to)
	}
	return nil
}
