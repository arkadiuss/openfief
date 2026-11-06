package views_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
)

// buyOf is a buy of qty units of in, in account acc on day, keyed id.
func buyOf(t *testing.T, acc, id, day string, in ledger.Instrument, qty string) ledger.Transaction {
	t.Helper()
	q, err := kernel.ParseQuantity(qty)
	if err != nil {
		t.Fatalf("ParseQuantity: %v", err)
	}
	h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: id}, Account: ledger.AccountID(acc), Date: date(t, day)}
	tx, err := ledger.NewBuy(h, in, q, money(t, "100", "EUR"), money(t, "0", "EUR"))
	if err != nil {
		t.Fatalf("NewBuy: %v", err)
	}
	return tx
}

func TestInstrumentTransactionsAndListings(t *testing.T) {
	t.Parallel()
	vwce, vwcg := instrument(t, isinIE, "VWCE"), instrument(t, isinIE, "VWCG")
	other := instrument(t, isinPL, "OTHR")
	h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: "dep"}, Account: "acc-1", Date: date(t, "2026-03-01")}
	deposit, err := ledger.NewDeposit(h, money(t, "500", "EUR"))
	if err != nil {
		t.Fatalf("NewDeposit: %v", err)
	}
	txs := []ledger.Transaction{
		deposit,
		buyOf(t, "acc-1", "b1", "2026-03-02", vwcg, "1"),
		buyOf(t, "acc-2", "b2", "2026-03-03", other, "2"),
		buyOf(t, "acc-2", "b3", "2026-03-04", vwce, "3"),
	}
	got := views.InstrumentTransactions(txs, isinIE)
	keys := make([]string, 0, len(got))
	for i := range got {
		keys = append(keys, got[i].Key.ExternalID)
	}
	if diff := cmp.Diff([]string{"b1", "b3"}, keys); diff != "" {
		t.Errorf("InstrumentTransactions (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]ledger.Instrument{vwce, vwcg}, views.InstrumentListings(txs, isinIE)); diff != "" {
		t.Errorf("InstrumentListings (-want +got):\n%s", diff)
	}
	if views.PostsTo(&txs[0], isinIE) || !views.PostsTo(&txs[2], isinPL) {
		t.Error("PostsTo should follow the holding postings")
	}
}

func TestListingOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want ledger.Instrument
		ok   bool
	}{
		{in: isinIE + "/VWCE", want: instrument(t, isinIE, "VWCE"), ok: true},
		{in: isinIE},
		{in: isinIE + "/"},
		{in: "NOTANISIN/VWCE"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, ok := views.ListingOf(tt.in)
			if ok != tt.ok || got != tt.want {
				t.Errorf("ListingOf = %v, %v; want %v, %v", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestHoldingsOf(t *testing.T) {
	t.Parallel()
	vwce, vwcg, other := instrument(t, isinIE, "VWCE"), instrument(t, isinIE, "VWCG"), instrument(t, isinPL, "OTHR")
	qty := func(s string) kernel.Quantity {
		q, err := kernel.ParseQuantity(s)
		if err != nil {
			t.Fatalf("ParseQuantity: %v", err)
		}
		return q
	}
	holdings := []ledger.Holding{
		{Account: "acc-1", Instrument: other, Quantity: qty("1")},
		{Account: "acc-1", Instrument: vwce, Quantity: qty("3")},
		{Account: "acc-1", Instrument: vwcg, Quantity: qty("2")},
	}
	lots := []ledger.Lot{
		{Instrument: vwce, Opened: date(t, "2026-01-02"), Quantity: qty("1")},
		{Instrument: other, Opened: date(t, "2026-01-03"), Quantity: qty("1")},
		{Instrument: vwce, Opened: date(t, "2026-02-02"), Quantity: qty("2")},
		{Instrument: vwcg, Opened: date(t, "2026-02-03"), Quantity: qty("2")},
	}
	want := []views.InstrumentHolding{
		{Account: "acc-1", Holding: holdings[1], Lots: []ledger.Lot{lots[0], lots[2]}},
		{Account: "acc-1", Holding: holdings[2], Lots: []ledger.Lot{lots[3]}},
	}
	if diff := cmp.Diff(want, views.HoldingsOf("acc-1", holdings, lots, isinIE)); diff != "" {
		t.Errorf("HoldingsOf (-want +got):\n%s", diff)
	}
}

// taxedWrappers returns pack's wrappers whose yearly tax tells assets
// apart, in the pack's order.
func taxedWrappers(t *testing.T, pack jurisdiction.Pack) []jurisdiction.Wrapper {
	t.Helper()
	var taxed []jurisdiction.Wrapper
	for _, w := range pack.Wrappers() {
		if len(pack.ExemptionClasses(w)) > 0 {
			taxed = append(taxed, w)
		}
	}
	if len(taxed) == 0 {
		t.Fatal("the Polish pack has no wrapper with a yearly tax")
	}
	return taxed
}

func TestAssetOf(t *testing.T) {
	t.Parallel()
	bond := views.AssetOf(instrument(t, isinPL, "EDO0336"), &valuation.Price{Close: money(t, "100", "PLN")})
	if bond.Kind != jurisdiction.AssetAccrued || bond.Instrument != isinPL {
		t.Errorf("AssetOf a bond = %+v", bond)
	}
	etf := views.AssetOf(instrument(t, isinIE, "VWCE"), &valuation.Price{Close: money(t, "100", "EUR")})
	if etf.Kind != jurisdiction.AssetListed || etf.Currency.String() != "EUR" {
		t.Errorf("AssetOf an ETF = %+v", etf)
	}
}

func TestInstrumentClasses(t *testing.T) {
	t.Parallel()
	pack := pl.NewPack()
	taxed := taxedWrappers(t, pack)
	bondAsset := views.AssetOf(instrument(t, isinPL, "EDO0336"), &valuation.Price{Close: money(t, "100", "PLN")})
	etf := views.AssetOf(instrument(t, isinIE, "VWCE"), &valuation.Price{Close: money(t, "100", "EUR")})
	tests := []struct {
		name  string
		asset jurisdiction.Asset
		owner map[string]jurisdiction.ExemptionClass
		want  func(w jurisdiction.Wrapper) views.InstrumentClass
	}{
		{
			name: "the pack's class", asset: bondAsset,
			want: func(w jurisdiction.Wrapper) views.InstrumentClass {
				c, known := pack.ExemptionClass(w, &bondAsset)
				return views.InstrumentClass{Wrapper: w, Class: c, Known: known}
			},
		},
		{
			name: "the owner's class wins", asset: etf, owner: map[string]jurisdiction.ExemptionClass{isinIE: pl.Investment},
			want: func(w jurisdiction.Wrapper) views.InstrumentClass {
				return views.InstrumentClass{Wrapper: w, Class: pl.Investment, OwnerChosen: true, Known: true}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			want := make([]views.InstrumentClass, 0, len(taxed))
			for _, w := range taxed {
				want = append(want, tt.want(w))
			}
			if diff := cmp.Diff(want, views.InstrumentClasses(pack, &tt.asset, tt.owner)); diff != "" {
				t.Errorf("InstrumentClasses (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMarkPrimary(t *testing.T) {
	t.Parallel()
	vwce, vwcg := instrument(t, isinIE, "VWCE").String(), instrument(t, isinIE, "VWCG").String()
	p := func(inst, source string) valuation.Price {
		return valuation.Price{Instrument: inst, Date: date(t, "2026-03-02"), Close: money(t, "100", "EUR"), Source: source}
	}
	prices := []valuation.Price{p(vwce, "eodhd"), p(vwce, "ibkr"), p(vwcg, "eodhd")}
	got := views.MarkPrimary(prices, map[string]string{vwce: "eodhd"})
	want := []views.InstrumentPrice{{Price: prices[0], Primary: true}, {Price: prices[1]}, {Price: prices[2]}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("MarkPrimary (-want +got):\n%s", diff)
	}
}
