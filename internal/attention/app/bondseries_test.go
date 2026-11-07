package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/app"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// bondBuy is a synthetic buy of qty bonds of the instrument isin/code in
// account on day.
func bondBuy(t *testing.T, account ledger.AccountID, isin, code, on, qty string) ledger.Transaction {
	t.Helper()
	in := instrumentOf(t, isin, code)
	q, err := kernel.ParseQuantity(qty)
	if err != nil {
		t.Fatalf("ParseQuantity: %v", err)
	}
	h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "treasury-bonds:" + string(account), ExternalID: "lot:" + code + ":" + on}, Account: account, Date: day(t, on)}
	tx, err := ledger.NewBuy(h, in, q, pln(t, qty+"00"), pln(t, "0"))
	if err != nil {
		t.Fatalf("NewBuy: %v", err)
	}
	return tx
}

// instrumentOf is the synthetic instrument isin/code.
func instrumentOf(t *testing.T, isin, code string) ledger.Instrument {
	t.Helper()
	i, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", isin, err)
	}
	return ledger.Instrument{ISIN: i, Symbol: code}
}

// TestIncompleteBondSeries checks D-214's check on 2026-11-07: a series
// held for seven days or more under its provisional ISIN, without terms,
// or both, is one condition, its subject the ISIN it is held under, since
// the first day any account held it; one held for less is silent, and one
// with the registry's ISIN and terms, one no longer held and an
// instrument that is no treasury bond are none.
func TestIncompleteBondSeries(t *testing.T) {
	t.Parallel()
	const a, b = ledger.AccountID("bonds-1"), ledger.AccountID("bonds-2")
	terms := map[ledger.ISIN]pl.Terms{}
	bondTerms(t, terms, "PLTOS1129009", "TOS1129", pl.FeeFromTable)
	bondTerms(t, terms, "PLEDO1236006", "EDO1236", pl.FeeFromTable)
	bondTerms(t, terms, "PL9999000026", "EDO0536", pl.FeeFromTable)
	redeemed, err := ledger.NewRedemption(
		ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "treasury-bonds:bonds-1", ExternalID: "lot:OTS0826:redemption"}, Account: a, Date: day(t, "2026-08-31")},
		instrumentOf(t, "PLOTS0826001", "OTS0826"), kernel.NewQuantity(kernel.NewDecimal(5, 0)), pln(t, "501.25"), pln(t, "0"),
	)
	if err != nil {
		t.Fatalf("NewRedemption: %v", err)
	}
	l := &fakeLedger{
		accounts: []ledger.Account{{ID: a}, {ID: b}},
		txs: map[ledger.AccountID][]ledger.Transaction{
			a: {
				bondBuy(t, a, "PLCOI1130004", "COI1130", "2026-10-20", "3"),
				bondBuy(t, a, "PL9999000018", "EDO1136", "2026-10-31", "2"),
				bondBuy(t, a, "PLEDO1236006", "EDO1236", "2026-11-03", "1"),
				bondBuy(t, a, "PLOTS0826001", "OTS0826", "2026-05-31", "5"), redeemed,
				bondBuy(t, a, "PL9999000034", "SYNT", "2026-01-02", "4"),
			},
			b: {
				bondBuy(t, b, "PLCOI1130004", "COI1130", "2026-10-01", "1"),
				bondBuy(t, b, "PLTOS1129009", "TOS1129", "2026-09-01", "6"),
				bondBuy(t, b, "PL9999000026", "EDO0536", "2026-05-04", "7"),
			},
		},
	}
	check := app.IncompleteBondSeries(l, &fakeBondTerms{terms: terms}, func() kernel.Date { return day(t, "2026-11-07") })
	got, err := check.Find(t.Context())
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	kind := attention.KindBondSeriesIncomplete
	const terms14 = "; its holding is not valued until they arrive, or until you enter its terms with a lot"
	want := []attention.Condition{
		{Kind: kind, Subject: "PL9999000018", Detail: "EDO1136, held since 2026-10-31, still has no series terms: the Ministry of Finance's data doesn't list them yet" + terms14},
		{Kind: kind, Subject: "PLCOI1130004", Detail: "COI1130, held since 2026-10-01, still has no ISIN and no series terms: the Ministry of Finance's data doesn't list them yet" + terms14},
		{Kind: kind, Subject: "PLTOS1129009", Detail: "TOS1129, held since 2026-09-01, still has no ISIN: the Ministry of Finance's data doesn't list it yet, so it is held under a provisional one"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Find (-want +got):\n%s", diff)
	}
	if check.Kind != kind || attention.GroupOf(kind) != attention.GroupDecisions {
		t.Errorf("the check's kind %s is in group %s, want %s among the decisions", check.Kind, attention.GroupOf(kind), kind)
	}
	if s := attention.ScopeOf(&attention.Item{Kind: kind, Subject: "PLCOI1130004"}); s.Instrument != "PLCOI1130004" {
		t.Errorf("ScopeOf = %+v, want the instrument", s)
	}
}
