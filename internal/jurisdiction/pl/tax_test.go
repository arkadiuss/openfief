package pl_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
)

func TestBelkaTax(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		interest string
		wrapper  pl.TaxWrapper
		want     string
		wantErr  error
	}{
		{name: "19% of 10.20 = 1.938, rounds to 1.94", interest: "10.20", wrapper: pl.Taxable, want: "1.94"},
		{name: "19% of 0.05 = 0.0095, rounds up to 0.01", interest: "0.05", wrapper: pl.Taxable, want: "0.01"},
		{name: "19% of 0.02 = 0.0038, rounds to 0.00", interest: "0.02", wrapper: pl.Taxable, want: "0"},
		{name: "19% of 0.50 = 0.095, tie rounds half up to 0.10", interest: "0.50", wrapper: pl.Taxable, want: "0.10"},
		{name: "no interest, no tax", interest: "0", wrapper: pl.Taxable, want: "0"},
		{name: "loss after fee, no tax", interest: "-1.86", wrapper: pl.Taxable, want: "0"},
		{name: "EDO0524 example: 19% of 2.16 = 0.4104, withheld 0.41", interest: "2.16", wrapper: pl.Taxable, want: "0.41"},
		{name: "ROR0623 example: 19% of a 0.44 coupon = 0.0836, withheld 0.08", interest: "0.44", wrapper: pl.Taxable, want: "0.08"},
		{name: "DOR0624 example: 19% of a 0.46 coupon = 0.0874, withheld 0.09", interest: "0.46", wrapper: pl.Taxable, want: "0.09"},
		{name: "COI0423 example: 19% of 9.75 = 1.8525, withheld 1.85", interest: "9.75", wrapper: pl.Taxable, want: "1.85"},
		{name: "IKE exempt", interest: "10.20", wrapper: pl.IKE, want: "0"},
		{name: "IKZE withholds nothing", interest: "10.20", wrapper: pl.IKZE, want: "0"},
		{name: "unknown wrapper", interest: "10.20", wrapper: pl.TaxWrapper(0), wantErr: pl.ErrUnknownTaxWrapper},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pl.BelkaTax(plnMoney(t, tt.interest), tt.wrapper)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(plnMoney(t, tt.want), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestBelkaRate(t *testing.T) {
	t.Parallel()
	if diff := cmp.Diff(dec(t, "0.19"), pl.BelkaRate()); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func TestTaxRateCutoff(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		event string
		want  string
	}{
		{name: "income on Tuesday: Monday's table at the latest", event: "2026-03-10", want: "2026-03-09"},
		{name: "income on Monday: Sunday, so Friday's table is the latest", event: "2026-03-09", want: "2026-03-08"},
		{name: "never the event day's own table", event: "2026-01-02", want: "2026-01-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(date(t, tt.want), pl.TaxRateCutoff(date(t, tt.event))); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
