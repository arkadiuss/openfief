package views_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/views"
)

// item is an allocation item: an account at institution in sleeve, holding
// value in currency worth pln.
type item struct {
	institution, sleeve string
	class               views.AssetClass
	value, currency     string
	pln                 string
}

func (i *item) build(t *testing.T) views.AllocationItem {
	t.Helper()
	return views.AllocationItem{
		Account:    ledger.Account{ID: "acc", Institution: i.institution, Name: "Synthetic", Sleeve: portfolio.SleeveID(i.sleeve)},
		AssetClass: i.class, Value: money(t, i.value, i.currency), Reporting: money(t, i.pln, "PLN"),
	}
}

func TestAllocate(t *testing.T) {
	t.Parallel()
	items := []item{
		{"broker-a", "core", views.AssetListed, "100", "USD", "400.00"},
		{"broker-a", "core", views.AssetCash, "50", "USD", "200.00"},
		{"broker-b", "offense", views.AssetListed, "300", "PLN", "300.00"},
		{"bond-registry", "core", views.AssetTreasuryBond, "1000.50", "PLN", "1000.50"},
		{"broker-b", "offense", views.AssetCash, "100", "EUR", "430.00"},
	}
	tests := []struct {
		by    views.Dimension
		want  [][2]string
		total string
	}{
		{views.ByAssetClass, [][2]string{{"treasury_bond", "1000.50"}, {"listed", "700.00"}, {"cash", "630.00"}}, "2330.50"},
		{views.ByCurrency, [][2]string{{"PLN", "1300.50"}, {"USD", "600.00"}, {"EUR", "430.00"}}, "2330.50"},
		{views.ByInstitution, [][2]string{{"bond-registry", "1000.50"}, {"broker-b", "730.00"}, {"broker-a", "600.00"}}, "2330.50"},
		{views.BySleeve, [][2]string{{"core", "1600.50"}, {"offense", "730.00"}}, "2330.50"},
	}
	for _, tt := range tests {
		t.Run(tt.by.String(), func(t *testing.T) {
			t.Parallel()
			in := make([]views.AllocationItem, 0, len(items))
			for i := range items {
				in = append(in, items[i].build(t))
			}
			got, err := views.Allocate(in, tt.by, pln(t))
			if err != nil {
				t.Fatalf("Allocate() error = %v", err)
			}
			want := views.Allocation{By: tt.by, Total: money(t, tt.total, "PLN")}
			for _, g := range tt.want {
				want.Groups = append(want.Groups, views.AllocationGroup{Key: g[0], Reporting: money(t, g[1], "PLN")})
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Allocate() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAllocateEdges(t *testing.T) {
	t.Parallel()
	t.Run("no items is zero PLN and no groups", func(t *testing.T) {
		t.Parallel()
		got, err := views.Allocate(nil, views.BySleeve, pln(t))
		if err != nil {
			t.Fatalf("Allocate() error = %v", err)
		}
		want := views.Allocation{By: views.BySleeve, Total: money(t, "0", "PLN"), Groups: []views.AllocationGroup{}}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Allocate() mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("equal groups are in key order", func(t *testing.T) {
		t.Parallel()
		in := []views.AllocationItem{(&item{"b", "s", views.AssetCash, "1", "PLN", "1"}).build(t), (&item{"a", "s", views.AssetCash, "1", "PLN", "1"}).build(t)}
		got, err := views.Allocate(in, views.ByInstitution, pln(t))
		if err != nil {
			t.Fatalf("Allocate() error = %v", err)
		}
		if got.Groups[0].Key != "a" || got.Groups[1].Key != "b" {
			t.Errorf("Allocate() groups = %v, want a then b", got.Groups)
		}
	})
	t.Run("a reporting value in another currency is an error", func(t *testing.T) {
		t.Parallel()
		bad := (&item{"a", "s", views.AssetCash, "1", "USD", "1"}).build(t)
		bad.Reporting = money(t, "1", "USD")
		if _, err := views.Allocate([]views.AllocationItem{bad}, views.ByCurrency, pln(t)); !errors.Is(err, views.ErrNotReportingCurrency) {
			t.Errorf("Allocate() error = %v, want %v", err, views.ErrNotReportingCurrency)
		}
	})
	t.Run("an unknown dimension is an error", func(t *testing.T) {
		t.Parallel()
		in := []views.AllocationItem{(&item{"a", "s", views.AssetCash, "1", "PLN", "1"}).build(t)}
		if _, err := views.Allocate(in, views.Dimension(9), pln(t)); !errors.Is(err, views.ErrUnknownDimension) {
			t.Errorf("Allocate() error = %v, want %v", err, views.ErrUnknownDimension)
		}
	})
}

func TestAllocateInAnyReportingCurrency(t *testing.T) {
	t.Parallel()
	eur := money(t, "0", "EUR").Currency()
	in := (&item{"a", "s", views.AssetCash, "1", "USD", "0"}).build(t)
	in.Reporting = money(t, "0.90", "EUR")
	got, err := views.Allocate([]views.AllocationItem{in}, views.ByCurrency, eur)
	if err != nil || !got.Total.Equal(money(t, "0.90", "EUR")) {
		t.Errorf("Allocate(EUR) = %v, %v, want a total of 0.90 EUR", got, err)
	}
}

func TestParseDimension(t *testing.T) {
	t.Parallel()
	for _, d := range []views.Dimension{views.ByAssetClass, views.ByCurrency, views.ByInstitution, views.BySleeve} {
		got, err := views.ParseDimension(d.String())
		if err != nil || got != d {
			t.Errorf("ParseDimension(%q) = %v, %v, want %v", d.String(), got, err, d)
		}
	}
	if _, err := views.ParseDimension("broker"); !errors.Is(err, views.ErrUnknownDimension) {
		t.Errorf("ParseDimension(broker) error = %v, want %v", err, views.ErrUnknownDimension)
	}
	if got := views.Dimension(0).String(); got != "Dimension(0)" {
		t.Errorf("Dimension(0).String() = %q", got)
	}
}

func TestAssetClassOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		isin, symbol string
		want         views.AssetClass
	}{
		{isinPL, "EDO0334", views.AssetTreasuryBond},
		{isinPL, "PKN", views.AssetListed},
		{isinIE, "EDO0334", views.AssetListed},
		{isinIE, "ETF1", views.AssetListed},
	}
	for _, tt := range tests {
		t.Run(tt.isin+"/"+tt.symbol, func(t *testing.T) {
			t.Parallel()
			if got := views.AssetClassOf(instrument(t, tt.isin, tt.symbol)); got != tt.want {
				t.Errorf("AssetClassOf() = %v, want %v", got, tt.want)
			}
		})
	}
	names := map[views.AssetClass]string{views.AssetCash: "cash", views.AssetTreasuryBond: "treasury_bond", views.AssetListed: "listed", 0: "AssetClass(0)"}
	for c, want := range names {
		if got := c.String(); got != want {
			t.Errorf("AssetClass(%d).String() = %q, want %q", int(c), got, want)
		}
	}
}

func TestValueHoldings(t *testing.T) {
	t.Parallel()
	acc := ledger.Account{ID: "acc-1", Institution: "broker-a", Name: "Synthetic", Sleeve: "core"}
	on := date(t, "2026-03-02")
	bond, etf, other := instrument(t, isinPL, "EDO0334"), instrument(t, isinIE, "ETF1"), instrument(t, isinIE, "ETF2")
	holdings := []ledger.Holding{
		{Account: acc.ID, Instrument: bond, Date: on, Quantity: kernel.NewQuantity(kernel.NewDecimal(10, 0))},
		{Account: acc.ID, Instrument: etf, Date: on, Quantity: kernel.NewQuantity(kernel.NewDecimal(3, 0))},
		{Account: acc.ID, Instrument: other, Date: on, Quantity: kernel.NewQuantity(kernel.NewDecimal(1, 0))},
	}
	stored := vals(t, []val{
		{isinPL, "EDO0334", "2026-03-02", "1050.00", "PLN", "1050.00"},
		{isinIE, "ETF1", "2026-03-02", "30", "USD", "120.00"},
		// Another date, another account or a missing input do not value the holdings on.
		{isinIE, "ETF2", "2026-03-01", "10", "USD", "40.00"},
		{isinIE, "ETF2", "2026-03-02", "missing", "USD", "missing"},
	})
	moved := vals(t, []val{{isinIE, "ETF2", "2026-03-02", "10", "USD", "40.00"}})[0]
	moved.Account = "acc-2"
	stored = append(stored, moved)

	items, unvalued, err := views.ValueHoldings(acc, holdings, stored, on, views.Gross)
	if err != nil {
		t.Fatalf("ValueHoldings: %v", err)
	}
	wantItems := []views.AllocationItem{
		{Account: acc, AssetClass: views.AssetTreasuryBond, Value: money(t, "1050.00", "PLN"), Reporting: money(t, "1050.00", "PLN")},
		{Account: acc, AssetClass: views.AssetListed, Value: money(t, "30", "USD"), Reporting: money(t, "120.00", "PLN")},
	}
	if diff := cmp.Diff(wantItems, items); diff != "" {
		t.Errorf("ValueHoldings() items mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]views.Unvalued{{Account: acc.ID, Holding: holdings[2], Missing: "price"}}, unvalued); diff != "" {
		t.Errorf("ValueHoldings() unvalued mismatch (-want +got):\n%s", diff)
	}
}
