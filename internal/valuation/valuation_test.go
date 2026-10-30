package valuation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

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
		{name: "net if redeemed is no basis since D-77", in: "net_if_redeemed", wantErr: valuation.ErrUnknownBasis},
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
