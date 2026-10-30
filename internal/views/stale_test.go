package views_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// staleLog is one account's synthetic history: ibkr-old, now removed,
// bought 10 ETF1 on 2026-03-02; ibkr-main bought 5 more on 2026-03-03 and
// sold 12 on 2026-03-04, consuming the old lot and part of the new one; an
// upload, not a connector, bought ETF2.
func staleLog(t *testing.T) []ledger.Transaction {
	t.Helper()
	etf1, etf2 := instrument(t, isinIE, "ETF1"), instrument(t, isinPL, "ETF2")
	usd := money(t, "1", "USD").Currency()
	zero := kernel.NewMoney(kernel.Decimal{}, usd)
	h := func(source, id, day string) ledger.TransactionHeader {
		return ledger.TransactionHeader{Key: ledger.TransactionKey{Source: source, ExternalID: id}, Account: "acc-1", Date: date(t, day)}
	}
	units := func(n int64) kernel.Quantity { return kernel.NewQuantity(kernel.NewDecimal(n, 0)) }
	old, err1 := ledger.NewBuy(h("plugin:ibkr-old:ibkr", "b1", "2026-03-02"), etf1, units(10), money(t, "1000", "USD"), zero)
	more, err2 := ledger.NewBuy(h("plugin:ibkr-main:ibkr", "b2", "2026-03-03"), etf1, units(5), money(t, "500", "USD"), zero)
	sell, err3 := ledger.NewSell(h("plugin:ibkr-main:ibkr", "s1", "2026-03-04"), etf1, units(12), money(t, "1300", "USD"), zero)
	other, err4 := ledger.NewBuy(h("upload", "b3", "2026-03-02"), etf2, units(1), money(t, "10", "USD"), zero)
	if err := errors.Join(err1, err2, err3, err4); err != nil {
		t.Fatal(err)
	}
	return []ledger.Transaction{old, other, more, sell}
}

func TestRemovedBehindHoldings(t *testing.T) {
	t.Parallel()
	txs := staleLog(t)
	installed := views.InstalledOf([]views.Plugin{{Instance: "ibkr-main"}})
	etf1 := instrument(t, isinIE, "ETF1")
	got := views.RemovedBehindHoldings(txs, installed, date(t, "2026-03-04"))
	if diff := cmp.Diff(map[ledger.Instrument][]string{etf1: {"ibkr-old"}}, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
	if got := views.RemovedBehindHoldings(txs, views.Installed{}, date(t, "2026-03-04")); len(got) != 0 {
		t.Errorf("with the installed instances unknown: %v", got)
	}
	if got := views.RemovedBehindHoldings(txs, installed, date(t, "2026-03-01")); len(got) != 0 {
		t.Errorf("before any transaction: %v", got)
	}
}

func TestRemovedBehindLots(t *testing.T) {
	t.Parallel()
	txs := staleLog(t)
	etf1 := instrument(t, isinIE, "ETF1")
	b2 := ledger.TransactionKey{Source: "plugin:ibkr-main:ibkr", ExternalID: "b2"}
	b1 := ledger.TransactionKey{Source: "plugin:ibkr-old:ibkr", ExternalID: "b1"}
	tests := []struct {
		name      string
		installed views.Installed
		on        string
		want      map[ledger.TransactionKey][]string
	}{
		{
			name: "the old lot is open: it came from the removed instance", installed: views.InstalledOf([]views.Plugin{{Instance: "ibkr-main"}}), on: "2026-03-03",
			want: map[ledger.TransactionKey][]string{b1: {"ibkr-old"}},
		},
		{
			name: "the sale consumed the old lot; the new one is current", installed: views.InstalledOf([]views.Plugin{{Instance: "ibkr-main"}}), on: "2026-03-04",
			want: map[ledger.TransactionKey][]string{},
		},
		{
			name: "the new lot is marked when the sale that cut it came from a removed instance", installed: views.InstalledOf(nil), on: "2026-03-04",
			want: map[ledger.TransactionKey][]string{b2: {"ibkr-main"}},
		},
		{name: "nothing is marked while the instances are unknown", on: "2026-03-04", want: map[ledger.TransactionKey][]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := views.RemovedBehindLots(txs, etf1, tt.installed, date(t, tt.on))
			if err != nil {
				t.Fatalf("RemovedBehindLots: %v", err)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
	short := txs[3:]
	if _, err := views.RemovedBehindLots(short, etf1, views.InstalledOf(nil), date(t, "2026-03-04")); err == nil {
		t.Error("a sale of more than held was replayed")
	}
}
