package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// coreToken is what the fake verifier takes for the core's token.
const coreToken = "synthetic-core-token"

type fakeVerifier struct{}

func (fakeVerifier) VerifyCore(_ context.Context, token string) error {
	if token != coreToken {
		return openfief.ErrNotCore
	}
	return nil
}

// coreClient is the generated client, as a struct so helpers return no
// interface.
type coreClient struct {
	connectorv1connect.ConnectorServiceClient
}

// fakeCPI serves synthetic CPI: 2.9% for every month of 2025, and January
// to September 2026 at 4.1%.
type fakeCPI struct{ err error }

func (f fakeCPI) CPI(_ context.Context, year int) ([]valuation.IndexReading, error) {
	if f.err != nil {
		return nil, f.err
	}
	months, value := 12, kernel.NewDecimal(29, -3)
	switch {
	case year == 2026:
		months, value = 9, kernel.NewDecimal(41, -3)
	case year > 2026:
		months = 0
	}
	out := make([]valuation.IndexReading, 0, months)
	for m := time.January; int(m) <= months; m++ {
		out = append(out, valuation.IndexReading{Kind: valuation.CPI, Year: year, Month: m, Value: value})
	}
	return out, nil
}

// fakeRates serves two synthetic reference rate changes: 4.00% from
// 2025-12-04, then 3.75% from 2026-03-05.
type fakeRates struct{ err error }

func (f fakeRates) ReferenceRates(context.Context) ([]pl.RateChange, error) {
	return []pl.RateChange{
		{From: mustDate("2025-12-04"), Rate: kernel.NewDecimal(4, -2)},
		{From: mustDate("2026-03-05"), Rate: kernel.NewDecimal(375, -4)},
	}, f.err
}

// mustDate parses a date that is valid by construction.
func mustDate(s string) kernel.Date {
	d, err := kernel.ParseDate(s)
	if err != nil {
		return kernel.Date{}
	}
	return d
}

// served serves a connector whose today is
// 2026-10-29 and whose index readings start in September 2026, with the SDK
// as the plugin does, and returns a client that calls it as the core.
func served(t *testing.T, cpi treasurybonds.CPISource, rates treasurybonds.ReferenceRateSource) coreClient {
	t.Helper()
	today := mustDate("2026-10-29")
	c := &connector{
		coupons: &treasurybonds.CouponSync{CPI: cpi, Rates: rates},
		indices: &treasurybonds.IndexSync{CPI: cpi, Rates: rates, Start: mustDate("2026-09-01")},
		today:   func() kernel.Date { return today },
	}
	server := openfief.NewConnectorServer("", c, fakeVerifier{})
	srv := httptest.NewUnstartedServer(server.Handler)
	srv.Config.Protocols = server.Protocols
	srv.Start()
	t.Cleanup(srv.Close)
	var p http.Protocols
	p.SetUnencryptedHTTP2(true)
	return coreClient{connectorv1connect.NewConnectorServiceClient(&http.Client{Transport: &http.Transport{Protocols: &p}}, srv.URL)}
}

// asCore returns msg as a request carrying the core's token.
func asCore[T any](msg *T) *connect.Request[T] {
	req := connect.NewRequest(msg)
	req.Header().Set("Authorization", "Bearer "+coreToken)
	return req
}

// edoLot is a synthetic lot of 40 EDO0536 bonds bought on 2026-05-04.
func edoLot() *connectorv1.BondLot {
	return &connectorv1.BondLot{
		Series: "EDO0536", Quantity: "40", PurchaseDate: "2026-05-04", FirstPeriodRate: "0.06", Margin: "0.02",
		EarlyRedemptionFee: &connectorv1.Money{Amount: "2.00", Currency: "PLN"},
	}
}

func TestImportBondLots(t *testing.T) {
	t.Parallel()
	client := served(t, fakeCPI{}, fakeRates{})
	// Bought after go-live: a deposit and a buy on the purchase date (D-98).
	resp, err := client.ImportBondLots(t.Context(), asCore(&connectorv1.ImportBondLotsRequest{BatchId: "b-1", AccountId: "acc-bonds", GoLive: "2026-05-01", Lots: []*connectorv1.BondLot{edoLot()}}))
	if err != nil {
		t.Fatalf("ImportBondLots: %v", err)
	}
	edo := &connectorv1.Instrument{Isin: "PLEDO0536000", Symbol: "EDO0536"}
	pln := func(amount string) *connectorv1.Money { return &connectorv1.Money{Amount: amount, Currency: "PLN"} }
	cash := &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_CASH, Currency: "PLN"}
	key := func(id string) *connectorv1.TransactionKey {
		return &connectorv1.TransactionKey{Source: "treasury-bonds:acc-bonds", ExternalId: "lot:EDO0536:2026-05-04:" + id}
	}
	want := &connectorv1.Delivery{
		BatchId: "b-1", Institution: "treasury-bonds",
		Transactions: []*connectorv1.Transaction{
			{
				Key: key("deposit"), AccountId: "acc-bonds", Kind: connectorv1.TransactionKind_TRANSACTION_KIND_DEPOSIT, Date: "2026-05-04",
				Description: "Cash for treasury bonds EDO0536",
				Postings: []*connectorv1.Posting{
					{Book: cash, Amount: pln("4000")},
					{Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_EQUITY, Currency: "PLN"}, Amount: pln("-4000")},
				},
			},
			{
				Key: key("buy"), AccountId: "acc-bonds", Kind: connectorv1.TransactionKind_TRANSACTION_KIND_BUY, Date: "2026-05-04",
				Description: "Treasury bonds EDO0536",
				Postings: []*connectorv1.Posting{
					{Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_HOLDING, Instrument: edo}, Amount: pln("4000"), Quantity: "40"},
					{Book: cash, Amount: pln("-4000")},
				},
			},
		},
		BondTerms: []*connectorv1.BondTerms{{Instrument: edo, FirstPeriodRate: "0.06", Margin: "0.02", EarlyRedemptionFee: pln("2")}},
	}
	if diff := cmp.Diff(want, resp.Msg.GetDelivery(), protocmp.Transform()); diff != "" {
		t.Errorf("ImportBondLots delivery (-want +got):\n%s", diff)
	}
}

func TestImportBondLotsRefuses(t *testing.T) {
	t.Parallel()
	client := served(t, fakeCPI{}, fakeRates{})
	with := func(change func(l *connectorv1.BondLot)) []*connectorv1.BondLot {
		l := edoLot()
		change(l)
		return []*connectorv1.BondLot{l}
	}
	tests := []struct {
		name   string
		goLive string
		lots   []*connectorv1.BondLot
	}{
		{name: "no go-live date", lots: []*connectorv1.BondLot{edoLot()}},
		{name: "an unknown bond type", goLive: "2026-06-01", lots: with(func(l *connectorv1.BondLot) { l.Series = "XYZ0536" })},
		{name: "a purchase date of another series", goLive: "2026-06-01", lots: with(func(l *connectorv1.BondLot) { l.PurchaseDate = "2026-06-04" })},
		{name: "bought after today", goLive: "2026-06-01", lots: with(func(l *connectorv1.BondLot) { l.Series, l.PurchaseDate = "EDO1136", "2026-11-04" })},
		{name: "a fraction of a bond", goLive: "2026-06-01", lots: with(func(l *connectorv1.BondLot) { l.Quantity = "2.5" })},
		{name: "no fee", goLive: "2026-06-01", lots: with(func(l *connectorv1.BondLot) { l.EarlyRedemptionFee = nil })},
		{name: "a rate that isn't a number", goLive: "2026-06-01", lots: with(func(l *connectorv1.BondLot) { l.FirstPeriodRate = "6%" })},
		{name: "one series and day twice", goLive: "2026-06-01", lots: []*connectorv1.BondLot{edoLot(), edoLot()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.ImportBondLots(t.Context(), asCore(&connectorv1.ImportBondLotsRequest{BatchId: "b-1", AccountId: "acc-bonds", GoLive: tt.goLive, Lots: tt.lots}))
			if connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Errorf("ImportBondLots = %v, want InvalidArgument", err)
			}
		})
	}
	if _, err := client.ImportBondLots(t.Context(), connect.NewRequest(&connectorv1.ImportBondLotsRequest{BatchId: "b-2"})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("ImportBondLots without the core's token: %v, want Unauthenticated", err)
	}
	// D-118: the core names the account with the lots; the config names none.
	noAccount := &connectorv1.ImportBondLotsRequest{BatchId: "b-3", GoLive: "2026-06-01", Lots: []*connectorv1.BondLot{edoLot()}}
	if _, err := client.ImportBondLots(t.Context(), asCore(noAccount)); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("ImportBondLots without an account: %v, want InvalidArgument", err)
	}
}

// heldRequest is a sync of two bound accounts (D-117): a taxable one live
// since 2025-06-01 with 10 COI1028 bonds bought on 2024-10-10 (first year
// 6.55%, margin 1.50%) and 5 EDO0536 bonds, which pay no coupons; an IKE
// one live since 2026-08-20 with 100 ROR0827 bonds bought that day (first
// month 4.25%, no margin); and one not seeded yet.
func heldRequest(mode connectorv1.Mode) *connectorv1.SyncRequest {
	pln := func(amount string) *connectorv1.Money { return &connectorv1.Money{Amount: amount, Currency: "PLN"} }
	coi := bondInstrument("COI1028")
	edo := &connectorv1.Instrument{Isin: "PLEDO0536000", Symbol: "EDO0536"}
	ror := bondInstrument("ROR0827")
	return &connectorv1.SyncRequest{
		BatchId: "b-1", Mode: mode,
		Accounts: []*connectorv1.AccountHoldings{
			{
				AccountId: "acc-bonds", GoLive: "2025-06-01", TaxWrapper: "taxable",
				Holdings: []*connectorv1.Holding{{Instrument: coi, Quantity: "10"}, {Instrument: edo, Quantity: "5"}},
				Lots: []*connectorv1.OpenLot{
					{Instrument: coi, Quantity: "10", Opened: "2024-10-10", Cost: pln("1000"), Opening: true},
					{Instrument: edo, Quantity: "5", Opened: "2026-05-04", Cost: pln("500")},
				},
			},
			{
				AccountId: "acc-bonds-ike", GoLive: "2026-08-20", TaxWrapper: "ike",
				Lots: []*connectorv1.OpenLot{{Instrument: ror, Quantity: "100", Opened: "2026-08-20", Cost: pln("10000"), Opening: true}},
			},
			{AccountId: "acc-new", TaxWrapper: "taxable"},
		},
		BondTerms: []*connectorv1.BondTerms{
			{Instrument: coi, FirstPeriodRate: "0.0655", Margin: "0.015", EarlyRedemptionFee: pln("2.00")},
			{Instrument: edo, FirstPeriodRate: "0.06", Margin: "0.02", EarlyRedemptionFee: pln("2.00")},
			{Instrument: ror, FirstPeriodRate: "0.0425", Margin: "0", EarlyRedemptionFee: pln("0.50")},
		},
	}
}

// bondInstrument is the instrument of series code, as the connector derives
// it (D-99); the zero instrument for a code that is no series.
func bondInstrument(code string) *connectorv1.Instrument {
	s, err := pl.ParseSeries(code)
	if err != nil {
		return &connectorv1.Instrument{}
	}
	in, err := treasurybonds.Instrument(s)
	if err != nil {
		return &connectorv1.Instrument{}
	}
	return &connectorv1.Instrument{Isin: in.ISIN.String(), Symbol: in.Symbol}
}

// couponLine is what TestSyncBooksCoupons compares of a delivered
// transaction: its account, key, date and postings by book kind.
type couponLine struct {
	Account, ID, Date string
	Postings          map[string]string
}

func TestSyncBooksCoupons(t *testing.T) {
	t.Parallel()
	client := served(t, fakeCPI{}, fakeRates{})
	resp, err := client.Sync(t.Context(), asCore(heldRequest(connectorv1.Mode_MODE_TRANSACTIONS)))
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	got := make([]couponLine, 0, len(resp.Msg.GetDelivery().GetTransactions()))
	for _, tx := range resp.Msg.GetDelivery().GetTransactions() {
		if tx.GetKind() != connectorv1.TransactionKind_TRANSACTION_KIND_INTEREST {
			t.Errorf("transaction %s is %v, want interest (D-101)", tx.GetKey().GetExternalId(), tx.GetKind())
		}
		line := couponLine{Account: tx.GetAccountId(), ID: tx.GetKey().GetSource() + " " + tx.GetKey().GetExternalId(), Date: tx.GetDate(), Postings: map[string]string{}}
		for _, p := range tx.GetPostings() {
			line.Postings[p.GetBook().GetKind().String()] = p.GetAmount().GetAmount()
		}
		got = append(got, line)
	}
	holding, cash, tax := "BOOK_KIND_HOLDING", "BOOK_KIND_CASH", "BOOK_KIND_WITHHOLDING_TAX"
	want := []couponLine{
		// Year 1 at 6.55%: 10 × 6.55 = 65.50, Belka 19% = 12.445 -> 12.45 withheld.
		{
			Account: "acc-bonds", ID: "treasury-bonds:acc-bonds lot:COI1028:2024-10-10:coupon:2025-10-10", Date: "2025-10-10",
			Postings: map[string]string{holding: "-65.5", tax: "12.45", cash: "53.05"},
		},
		// Year 2 at August 2025's CPI 2.9% + 1.50% = 4.40%: 44.00, tax 8.36.
		{
			Account: "acc-bonds", ID: "treasury-bonds:acc-bonds lot:COI1028:2024-10-10:coupon:2026-10-10", Date: "2026-10-10",
			Postings: map[string]string{holding: "-44", tax: "8.36", cash: "35.64"},
		},
		// Month 1 at 4.25%: 100 × 0.354 -> 0.35 a bond, no tax in IKE.
		{
			Account: "acc-bonds-ike", ID: "treasury-bonds:acc-bonds-ike lot:ROR0827:2026-08-20:coupon:2026-09-20", Date: "2026-09-20",
			Postings: map[string]string{holding: "-35", cash: "35"},
		},
		// Month 2 at the reference rate in force in mid-August 2026, 3.75%: 0.3125 -> 0.31 a bond.
		{
			Account: "acc-bonds-ike", ID: "treasury-bonds:acc-bonds-ike lot:ROR0827:2026-08-20:coupon:2026-10-20", Date: "2026-10-20",
			Postings: map[string]string{holding: "-31", cash: "31"},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("coupons (-want +got):\n%s", diff)
	}
	if len(resp.Msg.GetDelivery().GetIndexReadings()) == 0 {
		t.Error("the sync with coupons delivered no index readings")
	}

	// An opening batch seeds; it books no coupons.
	opening, err := client.Sync(t.Context(), asCore(heldRequest(connectorv1.Mode_MODE_OPENING)))
	if err != nil {
		t.Fatalf("opening Sync: %v", err)
	}
	if n := len(opening.Msg.GetDelivery().GetTransactions()); n != 0 {
		t.Errorf("an opening sync delivered %d transactions, want none", n)
	}

	// Holdings that don't parse refuse the sync.
	bad := heldRequest(connectorv1.Mode_MODE_TRANSACTIONS)
	bad.Accounts[0].Lots[0].Opened = "10.10.2024"
	if _, err := client.Sync(t.Context(), asCore(bad)); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("Sync with a malformed lot = %v, want InvalidArgument", err)
	}
}

func TestSyncDeliversIndexReadings(t *testing.T) {
	t.Parallel()
	client := served(t, fakeCPI{}, fakeRates{})
	resp, err := client.Sync(t.Context(), asCore(&connectorv1.SyncRequest{BatchId: "b-1", Mode: connectorv1.Mode_MODE_TRANSACTIONS}))
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	cpi, ref := connectorv1.Index_INDEX_CPI, connectorv1.Index_INDEX_REFERENCE_RATE
	// September's CPI is the last published. The reference rate's day for
	// September 2026 is 18 August, for October 17 September and for
	// November 19 October, all over; December's, 17 November, is not.
	want := &connectorv1.SyncResponse{
		Delivery: &connectorv1.Delivery{BatchId: "b-1", Institution: "treasury-bonds", IndexReadings: []*connectorv1.IndexReading{
			{Index: cpi, Date: "2026-09-01", Value: "0.041"},
			{Index: ref, Date: "2026-09-01", Value: "0.0375"},
			{Index: ref, Date: "2026-10-01", Value: "0.0375"},
			{Index: ref, Date: "2026-11-01", Value: "0.0375"},
		}},
		Watermark: "cpi=2026-09;reference-rate=2026-11",
	}
	if diff := cmp.Diff(want, resp.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("Sync (-want +got):\n%s", diff)
	}
	resp, err = client.Sync(t.Context(), asCore(&connectorv1.SyncRequest{BatchId: "b-2", Mode: connectorv1.Mode_MODE_TRANSACTIONS, Watermark: want.GetWatermark()}))
	if err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	again := &connectorv1.SyncResponse{
		Delivery:  &connectorv1.Delivery{BatchId: "b-2", Institution: "treasury-bonds"},
		Watermark: "cpi=2026-09;reference-rate=2026-11",
	}
	if diff := cmp.Diff(again, resp.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("second Sync (-want +got):\n%s", diff)
	}
}

// errOutage is a synthetic outage of GUS or NBP.
var errOutage = errors.New("synthetic outage")

func TestSyncRefuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		cpi       fakeCPI
		rates     fakeRates
		watermark string
		asCore    bool
		want      connect.Code
	}{
		{name: "a watermark it never returned", watermark: "2026-10-28", asCore: true, want: connect.CodeInvalidArgument},
		{name: "GUS is down", cpi: fakeCPI{err: errOutage}, asCore: true, want: connect.CodeUnavailable},
		{name: "NBP is down", rates: fakeRates{err: errOutage}, asCore: true, want: connect.CodeUnavailable},
		{name: "not the core", want: connect.CodeUnauthenticated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msg := &connectorv1.SyncRequest{BatchId: "b-1", Mode: connectorv1.Mode_MODE_TRANSACTIONS, Watermark: tt.watermark}
			req := connect.NewRequest(msg)
			if tt.asCore {
				req = asCore(msg)
			}
			_, err := served(t, tt.cpi, tt.rates).Sync(t.Context(), req)
			if got := connect.CodeOf(err); got != tt.want {
				t.Errorf("Sync = %v, want %s", err, tt.want)
			}
		})
	}
}

func TestImportBondLotsBeforeGoLiveAndDeleted(t *testing.T) {
	t.Parallel()
	client := served(t, fakeCPI{}, fakeRates{})
	lot := edoLot()
	lot.Purchase = connectorv1.BondPurchase_BOND_PURCHASE_EXCHANGE
	resp, err := client.ImportBondLots(t.Context(), asCore(&connectorv1.ImportBondLotsRequest{
		BatchId: "b-1", AccountId: "acc-bonds", GoLive: "2026-06-01", Lots: []*connectorv1.BondLot{lot},
		Deleted: []*connectorv1.BondLotKey{{Series: "EDO0536", PurchaseDate: "2026-05-10"}},
	}))
	if err != nil {
		t.Fatalf("ImportBondLots: %v", err)
	}
	edo := &connectorv1.Instrument{Isin: "PLEDO0536000", Symbol: "EDO0536"}
	pln := func(amount string) *connectorv1.Money { return &connectorv1.Money{Amount: amount, Currency: "PLN"} }
	key := func(id string) *connectorv1.TransactionKey {
		return &connectorv1.TransactionKey{Source: "treasury-bonds:acc-bonds", ExternalId: id}
	}
	// Bought before go-live by exchange at 99.90 PLN a bond: an opening lot on go-live (D-98, D-100); the
	// deleted lot of 2026-05-10, also before go-live, is a deletion of its opening lot.
	want := []*connectorv1.Transaction{
		{
			Key: key("lot:EDO0536:2026-05-04"), AccountId: "acc-bonds", Kind: connectorv1.TransactionKind_TRANSACTION_KIND_OPENING_LOT,
			Date: "2026-06-01", LotOpened: "2026-05-04", Description: "Treasury bonds EDO0536 bought before go-live",
			Postings: []*connectorv1.Posting{
				{Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_HOLDING, Instrument: edo}, Amount: pln("3996"), Quantity: "40"},
				{Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_EQUITY, Currency: "PLN"}, Amount: pln("-3996")},
			},
		},
		{
			Key: key("lot:EDO0536:2026-05-10"), AccountId: "acc-bonds", Kind: connectorv1.TransactionKind_TRANSACTION_KIND_DELETION,
			Date: "2026-06-01", Description: "Deleted treasury bonds EDO0536",
		},
	}
	if diff := cmp.Diff(want, resp.Msg.GetDelivery().GetTransactions(), protocmp.Transform()); diff != "" {
		t.Errorf("ImportBondLots transactions (-want +got):\n%s", diff)
	}
}
