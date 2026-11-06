package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// fakeTerms are the accounts' recorded wrappers and fee rules and the
// owner's settings.
type fakeTerms struct {
	wrappers map[ledger.AccountID]jurisdiction.Wrapper
	fees     map[ledger.AccountID]jurisdiction.FeeRule
	settings jurisdiction.Settings
	format   *jurisdiction.NumberFormat
	classes  map[string]jurisdiction.ExemptionClass
	forms    map[ledger.AccountID]jurisdiction.AccountForm
	paid     map[ledger.AccountID]kernel.Money
}

func (f *fakeTerms) AccountForm(_ context.Context, account ledger.AccountID) (jurisdiction.AccountForm, bool, error) {
	form, ok := f.forms[account]
	return form, ok, nil
}

func (f *fakeTerms) EarlierPayments(_ context.Context, account ledger.AccountID) (kernel.Money, bool, error) {
	paid, ok := f.paid[account]
	return paid, ok, nil
}

func (f *fakeTerms) ExemptionClasses(context.Context) (map[string]jurisdiction.ExemptionClass, error) {
	return f.classes, nil
}

func (f *fakeTerms) TaxWrapper(_ context.Context, account ledger.AccountID) (jurisdiction.Wrapper, bool, error) {
	w, ok := f.wrappers[account]
	return w, ok, nil
}

func (f *fakeTerms) FeeRule(_ context.Context, account ledger.AccountID) (jurisdiction.FeeRule, error) {
	return f.fees[account], nil
}

func (f *fakeTerms) ValueSettings(context.Context) (jurisdiction.Settings, error) {
	return f.settings, nil
}

func (f *fakeTerms) NumberFormat(context.Context) (jurisdiction.NumberFormat, bool, error) {
	if f.format == nil {
		return jurisdiction.NumberFormat{}, false, nil
	}
	return *f.format, true, nil
}

// valuesViews returns the fixture's views with pack and terms, and 100 EUR
// deposited into acc-2 on 2026-03-01.
func valuesViews(t *testing.T, terms *fakeTerms, pack jurisdiction.Pack) *app.Views {
	t.Helper()
	f := newFixture(t)
	deposit, err := ledger.NewDeposit(ledger.TransactionHeader{
		Key: ledger.TransactionKey{Source: "test", ExternalID: "d1"}, Account: "acc-2", Date: date(t, "2026-03-01"),
	}, money(t, "100", "EUR"))
	if err != nil {
		t.Fatal(err)
	}
	f.ledger.txs = map[ledger.AccountID][]ledger.Transaction{"acc-2": {deposit}}
	f.statuses["acc-2"] = f.statuses["acc-1"]
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"), app.WithPortfolio(f.ledger), app.WithJurisdiction(pack, terms))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func ikzeTerms() *fakeTerms {
	return &fakeTerms{wrappers: map[ledger.AccountID]jurisdiction.Wrapper{"acc-2": "ikze"}, fees: map[ledger.AccountID]jurisdiction.FeeRule{
		"acc-2": {Rate: kernel.NewDecimal(1, -2)},
	}}
}

func TestAccountTerms(t *testing.T) {
	t.Parallel()
	v := valuesViews(t, ikzeTerms(), pl.NewPack())
	ctx := reader(t)
	got1, err1 := v.AccountTerms(ctx, "acc-1")
	got2, err2 := v.AccountTerms(ctx, "acc-2")
	if err := errors.Join(err1, err2); err != nil {
		t.Fatal(err)
	}
	got := []string{string(got1.Wrapper), string(got2.Wrapper), got2.Fee.Rate.String()}
	if diff := cmp.Diff([]string{"taxable", "ikze", "0.01"}, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func TestSettings(t *testing.T) {
	t.Parallel()
	terms := ikzeTerms()
	v := valuesViews(t, terms, pl.NewPack())
	ctx := reader(t)
	summary := func(v *app.Views) []any {
		t.Helper()
		s, err := v.Settings(ctx)
		if err != nil {
			t.Fatalf("Settings: %v", err)
		}
		return []any{s.ReportingCurrency.String(), s.NumberFormat.String(), s.NumberFormatDefault, len(s.Wrappers)}
	}
	if diff := cmp.Diff([]any{"PLN", "space-comma", true, 4}, summary(v)); diff != "" {
		t.Errorf("the pack's default (-want +got):\n%s", diff)
	}
	terms.format = &jurisdiction.NumberFormat{Group: jurisdiction.SeparatorComma, Decimal: jurisdiction.SeparatorDot}
	if diff := cmp.Diff([]any{"PLN", "comma-dot", false, 4}, summary(v)); diff != "" {
		t.Errorf("the owner's choice (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]any{"PLN", "space-dot", true, 0}, summary(newFixture(t).views(t))); diff != "" {
		t.Errorf("without a jurisdiction (-want +got):\n%s", diff)
	}
}

// TestTotalsWithIKZECash checks acc-2's totals: 100 EUR at 4.30 is 430 PLN
// gross, 387 net after IKZE's 10%, and no liquidation total without the
// owner's marginal rate; 0.32 gives 292.40.
func TestTotalsWithIKZECash(t *testing.T) {
	t.Parallel()
	terms := ikzeTerms()
	v := valuesViews(t, terms, pl.NewPack())
	got, err := v.AccountValuations(reader(t), "acc-2", date(t, "2026-03-01"), date(t, "2026-03-01"))
	if err != nil {
		t.Fatalf("AccountValuations: %v", err)
	}
	want := []views.DateTotal{{
		Date: date(t, "2026-03-01"), Reporting: money(t, "430.00", "PLN"), ByCurrency: []kernel.Money{money(t, "100", "EUR")},
		Net: money(t, "387.00", "PLN"), Liquidation: money(t, "0", "PLN"), LiquidationMissing: true, Cash: money(t, "430.00", "PLN"),
		NetOffset: money(t, "0", "PLN"), LiquidationOffset: money(t, "0", "PLN"),
	}}
	if diff := cmp.Diff(want, got.Totals); diff != "" {
		t.Errorf("totals (-want +got):\n%s", diff)
	}
	terms.settings = jurisdiction.Settings{MarginalRate: dec(t, "0.32"), HasMarginalRate: true}
	got, err = v.AccountValuations(reader(t), "acc-2", date(t, "2026-03-01"), date(t, "2026-03-01"))
	if err != nil || got.Totals[0].LiquidationMissing || !got.Totals[0].Liquidation.Equal(money(t, "292.40", "PLN")) {
		t.Errorf("with the marginal rate: %+v, %v", got.Totals, err)
	}
}

// TestAllocationOfAKind splits the fixture's net value: acc-2's IKZE cash
// counts at 387 PLN; its liquidation value is left out without the rate.
func TestAllocationOfAKind(t *testing.T) {
	t.Parallel()
	v := valuesViews(t, ikzeTerms(), pl.NewPack())
	ctx := reader(t)
	net, err := v.Allocation(ctx, date(t, "2026-03-02"), views.BySleeve, views.Net)
	if err != nil {
		t.Fatalf("Allocation(net): %v", err)
	}
	want := []views.AllocationGroup{{Key: "core", Reporting: money(t, "1620.00", "PLN")}, {Key: "offense", Reporting: money(t, "387.00", "PLN")}}
	if diff := cmp.Diff(want, net.Allocation.Groups); diff != "" || net.Kind != views.Net {
		t.Errorf("net groups (-want +got):\n%s", diff)
	}
	liq, err := v.Allocation(ctx, date(t, "2026-03-02"), views.BySleeve, views.Liquidation)
	if err != nil {
		t.Fatalf("Allocation(liquidation): %v", err)
	}
	last := liq.Unvalued[len(liq.Unvalued)-1]
	if last.Account != "acc-2" || !last.Cash.Equal(money(t, "100", "EUR")) || last.Missing != "marginal_rate" {
		t.Errorf("liquidation unvalued = %+v", liq.Unvalued)
	}
}

// TestYearlyTaxes checks OKI's yearly tax (D-127): none in force in 2026
// under the real act; under the act moved to 2026 at 1%, acc-2 in OKI,
// opened on 2026-03-01 with 100 EUR and valued through 2026-03-02, holds
// foreign cash, which no exemption covers, worth 430.00 PLN at the rate of
// the day before each of its two days: 430.00 averaged and taxed, 4.30,
// rounded to 4 złoty (D-144), an estimate through 2026-03-02 (D-146).
// acc-1, in OKI too and holding a USD fund, has no USD rate before
// 2026-03-02, so its tax can't be converted.
func TestYearlyTaxes(t *testing.T) {
	t.Parallel()
	ctx := reader(t)
	notInForce, err := valuesViews(t, ikzeTerms(), pl.NewPack()).YearlyTaxes(ctx, 2026)
	if err != nil || len(notInForce) != 1 || notInForce[0].Wrapper != "oki" || notInForce[0].InForce {
		t.Errorf("YearlyTaxes without an act = %+v, %v", notInForce, err)
	}
	terms := &fakeTerms{wrappers: map[ledger.AccountID]jurisdiction.Wrapper{"acc-2": "oki"}}
	act := pl.CurrentOKIAct()
	act.InForceFrom, act.Rates = date(t, "2026-01-01"), map[int]kernel.Decimal{2026: dec(t, "0.01")}
	got, err := valuesViews(t, terms, &pl.Pack{OKI: act}).YearlyTaxes(ctx, 2026)
	if err != nil || len(got) != 1 {
		t.Fatalf("YearlyTaxes = %+v, %v", got, err)
	}
	g := got[0]
	summary := []any{g.Days, g.Through.String(), g.Estimate, g.Average.String(), g.Taxed.String(), g.Tax.String(), len(g.Unclassified)}
	if diff := cmp.Diff([]any{2, "2026-03-02", true, "430 PLN", "430 PLN", "4 PLN", 0}, summary); diff != "" {
		t.Errorf("YearlyTaxes under the act (-want +got):\n%s", diff)
	}
}

// TestYearlyTaxesWithoutARate checks acc-1 in OKI, holding a USD fund
// with no USD rate before 2026-03-02: its yearly tax can't be converted.
func TestYearlyTaxesWithoutARate(t *testing.T) {
	t.Parallel()
	ctx := reader(t)
	act := pl.CurrentOKIAct()
	act.InForceFrom, act.Rates = date(t, "2026-01-01"), map[int]kernel.Decimal{2026: dec(t, "0.01")}
	f := newFixture(t)
	deposit, err := ledger.NewDeposit(ledger.TransactionHeader{
		Key: ledger.TransactionKey{Source: "test", ExternalID: "d0"}, Account: "acc-1", Date: date(t, "2026-03-01"),
	}, money(t, "10", "PLN"))
	if err != nil {
		t.Fatal(err)
	}
	f.ledger.txs = map[ledger.AccountID][]ledger.Transaction{"acc-1": {deposit}}
	oki := &fakeTerms{wrappers: map[ledger.AccountID]jurisdiction.Wrapper{"acc-1": "oki"}}
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"), app.WithPortfolio(f.ledger), app.WithJurisdiction(&pl.Pack{OKI: act}, oki))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.YearlyTaxes(ctx, 2026); !errors.Is(err, views.ErrTaxConversion) {
		t.Errorf("YearlyTaxes without a USD rate: %v, want ErrTaxConversion", err)
	}
}

func TestHoldingsAndLotsMarkRemovedInstances(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	etf1 := instrument(t, isinIE, "ETF1")
	usd := cur(t, "USD")
	buy, err := ledger.NewBuy(ledger.TransactionHeader{
		Key: ledger.TransactionKey{Source: ingest.ConnectorKeySource("ibkr-old", "ibkr"), ExternalID: "b1"}, Account: "acc-1", Date: f.on,
	}, etf1, kernel.NewQuantity(kernel.NewDecimal(3, 0)), money(t, "90", "USD"), kernel.NewMoney(kernel.Decimal{}, usd))
	if err != nil {
		t.Fatal(err)
	}
	f.ledger.txs = map[ledger.AccountID][]ledger.Transaction{"acc-1": {buy}}
	f.ledger.lots = map[ledger.AccountID][]ledger.Lot{"acc-1": {{Instrument: etf1, Opened: f.on, Quantity: kernel.NewQuantity(kernel.NewDecimal(3, 0)), Cost: money(t, "90", "USD"), Origin: buy.Key}}}
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"), app.WithPortfolio(f.ledger),
		app.WithPlugins(fakePlugins{records: []views.PluginRecord{record("ibkr-main")}}))
	if err != nil {
		t.Fatal(err)
	}
	holdings, err := v.Holdings(reader(t), "acc-1", f.on)
	if diff := cmp.Diff(map[ledger.Instrument][]string{etf1: {"ibkr-old"}}, holdings.Removed); err != nil || diff != "" {
		t.Errorf("holdings marks, %v (-want +got):\n%s", err, diff)
	}
	lots, err := v.Lots(reader(t), "acc-1", f.on)
	if diff := cmp.Diff(map[ledger.TransactionKey][]string{buy.Key: {"ibkr-old"}}, lots.Removed); err != nil || diff != "" {
		t.Errorf("lot marks, %v (-want +got):\n%s", err, diff)
	}
}
