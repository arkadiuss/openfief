package valuation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

func TestBasisNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    valuation.Basis
		wantErr error
	}{
		{name: "market", in: "market", want: valuation.Market},
		{name: "gross accrued", in: "gross_accrued", want: valuation.GrossAccrued},
		{name: "net if redeemed", in: "net_if_redeemed", want: valuation.NetIfRedeemed},
		{name: "wrong case", in: "Market", wantErr: valuation.ErrUnknownBasis},
		{name: "empty", in: "", wantErr: valuation.ErrUnknownBasis},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := valuation.ParseBasis(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if got != tt.want || got.String() != tt.in {
				t.Errorf("ParseBasis(%q) = %v (%q), want %v", tt.in, got, got.String(), tt.want)
			}
		})
	}
	if got := valuation.Basis(9).String(); got != "Basis(9)" {
		t.Errorf("Basis(9).String() = %q", got)
	}
}

func TestMarketValuation(t *testing.T) {
	t.Parallel()
	// 12.5 units x 101.37 USD = 1267.125 USD, exact and still in USD.
	p := valuation.Price{Instrument: "SYNTH-ETF", Date: date(t, "2026-03-09"), Close: money(t, "101.37", "USD"), Source: "test-vendor"}
	got, err := valuation.MarketValuation("SYNTH-ETF", kernel.NewQuantity(dec(t, "12.5")), &p)
	if err != nil {
		t.Fatalf("MarketValuation: %v", err)
	}
	want := valuation.Valuation{
		Instrument: "SYNTH-ETF", Date: date(t, "2026-03-09"), Value: money(t, "1267.125", "USD"),
		Basis: valuation.Market, Source: "test-vendor",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
	if _, err := valuation.MarketValuation("OTHER", kernel.NewQuantity(dec(t, "1")), &p); !cmp.Equal(valuation.ErrNoPrice, err, cmpopts.EquateErrors()) {
		t.Errorf("price of another instrument: error = %v, want ErrNoPrice", err)
	}
}

func TestBondBasis(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		configured valuation.Basis
		want       valuation.Basis
		wantErr    error
	}{
		{name: "not configured: net if redeemed (D-25)", configured: 0, want: valuation.NetIfRedeemed},
		{name: "configured gross accrued", configured: valuation.GrossAccrued, want: valuation.GrossAccrued},
		{name: "configured net if redeemed", configured: valuation.NetIfRedeemed, want: valuation.NetIfRedeemed},
		{name: "retail bonds have no market price", configured: valuation.Market, wantErr: valuation.ErrBasisNotForBonds},
		{name: "unknown basis", configured: valuation.Basis(9), wantErr: valuation.ErrUnknownBasis},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := valuation.BondBasis(tt.configured)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if got != tt.want {
				t.Errorf("BondBasis(%v) = %v, want %v", tt.configured, got, tt.want)
			}
		})
	}
}

// The synthetic EDO lot of the pl tests: 10 bonds, 6% first year, 2.00 PLN
// fee, bought 2024-03-15, valued 2024-09-15: gross 1030.20, net 1008.26.
func TestBondValuation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		configured valuation.Basis
		wantBasis  valuation.Basis
		wantValue  string
		wantErr    error
	}{
		{name: "default counts net if redeemed", configured: 0, wantBasis: valuation.NetIfRedeemed, wantValue: "1008.26"},
		{name: "gross accrued when configured", configured: valuation.GrossAccrued, wantBasis: valuation.GrossAccrued, wantValue: "1030.20"},
		{name: "market is rejected", configured: valuation.Market, wantErr: valuation.ErrBasisNotForBonds},
	}
	s, err := pl.ParseSeries("EDO0334")
	if err != nil {
		t.Fatal(err)
	}
	terms := pl.Terms{Series: s, FirstPeriodRate: dec(t, "0.06"), Margin: dec(t, "0.02"), EarlyRedemptionFee: money(t, "2.00", "PLN")}
	on := date(t, "2024-09-15")
	bought := pl.Purchase{Date: date(t, "2024-03-15"), Price: pl.Nominal()}
	lot, err := pl.ValueLot(&terms, nil, kernel.NewQuantity(dec(t, "10")), bought, pl.Taxable, on)
	if err != nil {
		t.Fatalf("ValueLot: %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := valuation.BondValuation("EDO0334", on, &lot, tt.configured, "statutory")
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			want := valuation.Valuation{Instrument: "EDO0334", Date: on, Value: money(t, tt.wantValue, "PLN"), Basis: tt.wantBasis, Source: "statutory"}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
