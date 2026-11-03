package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

// couponConnector books coupons as the treasury-bonds connector does
// (D-117), from what the sync request says the bound accounts hold, with
// a synthetic CPI of 2.9% for August 2025, as of 2026-10-29. It records
// the accounts it was told about.
type couponConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	mu   sync.Mutex
	held []ingest.HeldAccount
}

func (c *couponConnector) Sync(_ context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	held, terms, err := delivery.Held(req.Msg)
	if err != nil {
		return nil, fmt.Errorf("sync: %w", connect.NewError(connect.CodeInvalidArgument, err))
	}
	c.mu.Lock()
	c.held = held
	c.mu.Unlock()
	today, err := kernel.ParseDate("2026-10-29")
	if err != nil {
		return nil, fmt.Errorf("sync: %w", connect.NewError(connect.CodeInternal, err))
	}
	due, _ := treasurybonds.DueCoupons(held, terms, today)
	cpi := []valuation.IndexReading{{Kind: valuation.CPI, Year: 2025, Month: time.August, Value: kernel.NewDecimal(29, -3), Source: "gus-cpi"}}
	m, err := treasurybonds.BookCoupons(due, cpi)
	if err != nil {
		return nil, fmt.Errorf("sync: %w", connect.NewError(connect.CodeInternal, err))
	}
	d, err := delivery.FromMapping(req.Msg.GetBatchId(), string(treasurybonds.Institution), &m)
	if err != nil {
		return nil, fmt.Errorf("sync: %w", connect.NewError(connect.CodeInternal, err))
	}
	return connect.NewResponse(&connectorv1.SyncResponse{Delivery: d, Watermark: "w"}), nil
}

// TestSyncBooksBondCoupons checks, through the real core and Postgres,
// that a sync of the treasury-bonds instance carries what its bound bond
// account holds, with the series terms (D-117), and that the coupons it
// delivers are booked as interest once, however often it syncs (D-101).
func TestSyncBooksBondCoupons(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	conn := &couponConnector{}
	port := serveConnector(t, conn)
	cl := fakeCluster(t, port, "bonds-main")
	saveDemoRegistry(t, dsn)
	seedBondLot(t, dsn)
	if _, err := connectorCmd(t, dsn, cl, "bind", "--instance=bonds-main", "--account="+string(demo.BondsAccount)); err != nil {
		t.Fatalf("connector bind: %v", err)
	}
	out, err := connectorCmd(t, dsn, cl, "sync", "--instance=bonds-main")
	// Year 1 (2025-10-10) at 6.55% and year 2 (2026-10-10) at 2.9% + 1.50%.
	if err != nil || !strings.Contains(out, "(transactions): 2 transactions, 2 new") {
		t.Fatalf("connector sync = %q, %v", out, err)
	}
	out, err = connectorCmd(t, dsn, cl, "sync", "--instance=bonds-main")
	if err != nil || !strings.Contains(out, "2 transactions, 0 new, 0 new versions, 2 unchanged") {
		t.Fatalf("second connector sync = %q, %v", out, err)
	}
	conn.mu.Lock()
	held := conn.held
	conn.mu.Unlock()
	if len(held) != 1 || held[0].Account != demo.BondsAccount || len(held[0].Lots) != 1 || held[0].GoLive.String() != "2025-01-02" {
		t.Errorf("the sync told the connector %+v; want the bond account live since 2025-01-02 with its lot", held)
	}
	checkCoupons(t, dsn)
}

// seedBondLot books 10 COI1028 bonds bought on 2024-10-10 as an opening
// lot of the demo bond account on 2025-01-02, with the series' terms
// (first year 6.55%, margin 1.50%), as the treasury-bonds mapping does.
func seedBondLot(t *testing.T, dsn string) {
	t.Helper()
	err := withConnectorDeps(t.Context(), dsn, quiet(), func(d *connectorDeps) error {
		day := func(s string) kernel.Date {
			v, err := kernel.ParseDate(s)
			if err != nil {
				t.Fatalf("ParseDate: %v", err)
			}
			return v
		}
		lot := ingest.BondLot{Terms: coiTerms(t), Quantity: kernel.NewQuantity(kernel.NewDecimal(10, 0)), Purchased: day("2024-10-10")}
		e := ingest.BondLotEntries{Account: demo.BondsAccount, GoLive: day("2025-01-02"), Lots: []ingest.BondLot{lot}}
		m, terms, err := treasurybonds.Mapper{}.Map(&e, day("2026-10-29"))
		if err != nil {
			return fmt.Errorf("setup: %w", err)
		}
		if _, err := d.ledger.AddOpeningLots(t.Context(), m.Transactions, nil); err != nil {
			return fmt.Errorf("setup: %w", err)
		}
		for i := range terms {
			terms[i].Source = "bonds-main" // the instance that delivered them (D-179)
		}
		if err := (valuationInputs{recorder: d.recorder}).StoreBondTerms(t.Context(), "setup", terms); err != nil {
			return fmt.Errorf("setup: %w", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seeding a bond lot: %v", err)
	}
}

// checkCoupons checks the two coupons in the ledger: interest, keyed per
// lot and coupon date under the instance, with Belka tax withheld in the
// taxable account.
func checkCoupons(t *testing.T, dsn string) {
	t.Helper()
	err := withConnectorDeps(t.Context(), dsn, quiet(), func(d *connectorDeps) error {
		txs, err := d.ledger.CurrentTransactions(t.Context(), demo.BondsAccount)
		if err != nil {
			return fmt.Errorf("check: %w", err)
		}
		var got []string
		for i := range txs {
			if txs[i].Kind == ledger.KindInterest {
				got = append(got, txs[i].Key.ExternalID+" "+txs[i].Date.String()+" "+cashOf(&txs[i]))
			}
		}
		want := []string{
			"lot:COI1028:2024-10-10:coupon:2025-10-10 2025-10-10 53.05",
			"lot:COI1028:2024-10-10:coupon:2026-10-10 2026-10-10 35.64",
		}
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("interest booked:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("checking: %v", err)
	}
}

// coiTerms are COI1028's synthetic terms: 6.55% the first year, 1.50%
// margin, a 2.00 PLN fee.
func coiTerms(t *testing.T) pl.Terms {
	t.Helper()
	s, err := pl.ParseSeries("COI1028")
	if err != nil {
		t.Fatalf("ParseSeries: %v", err)
	}
	fee, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	return pl.Terms{
		Series: s, FirstPeriodRate: kernel.NewDecimal(655, -4), Margin: kernel.NewDecimal(15, -3),
		EarlyRedemptionFee: kernel.NewMoney(kernel.NewDecimal(2, 0), fee),
	}
}

// cashOf returns the amount t posts to cash.
func cashOf(t *ledger.Transaction) string {
	for _, p := range t.Postings {
		if p.Book.Kind == ledger.BookCash {
			return p.Amount.Amount().String()
		}
	}
	return ""
}
