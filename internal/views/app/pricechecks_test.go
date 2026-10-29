package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views/app"
)

var errChecksDown = errors.New("synthetic price check failure")

// fakePriceChecks answers fixed price checks at a fixed threshold, or
// fails.
type fakePriceChecks struct {
	checks    []valuation.PriceCheck
	threshold kernel.Decimal
	fail      error
}

func (f fakePriceChecks) PriceChecks(context.Context) ([]valuation.PriceCheck, error) {
	return f.checks, f.fail
}

func (f fakePriceChecks) Threshold() kernel.Decimal { return f.threshold }

// syntheticChecks are a matched and a flagged check of a synthetic ETF.
func syntheticChecks(t *testing.T) []valuation.PriceCheck {
	t.Helper()
	in := instrument(t, isinIE, "ETF1").String()
	check := func(on, closing, mark, diff string, status valuation.PriceCheckStatus) valuation.PriceCheck {
		return valuation.PriceCheck{
			Price:      valuation.Price{Instrument: in, Date: date(t, on), Close: money(t, closing, "USD"), Source: "eodhd"},
			CrossCheck: valuation.CrossCheckPrice{Instrument: in, Date: date(t, on), Close: money(t, mark, "USD"), Source: "ibkr-main"},
			Difference: dec(t, diff), Status: status,
		}
	}
	return []valuation.PriceCheck{
		check("2026-03-02", "100", "100.2", "0.002", valuation.PriceMatched),
		check("2026-03-03", "100", "103", "0.03", valuation.PriceFlagged),
	}
}

func TestPriceChecks(t *testing.T) {
	t.Parallel()
	checks := syntheticChecks(t)
	tests := []struct {
		name    string
		source  app.PriceChecks
		status  valuation.PriceCheckStatus
		want    app.PriceCheckList
		wantErr error
	}{
		{
			name: "every check", source: fakePriceChecks{checks: checks, threshold: dec(t, "0.01")},
			want: app.PriceCheckList{Threshold: dec(t, "0.01"), Checks: checks},
		},
		{
			name: "only flagged", source: fakePriceChecks{checks: checks, threshold: dec(t, "0.01")}, status: valuation.PriceFlagged,
			want: app.PriceCheckList{Threshold: dec(t, "0.01"), Checks: checks[1:]},
		},
		{
			name: "only matched", source: fakePriceChecks{checks: checks, threshold: dec(t, "0.01")}, status: valuation.PriceMatched,
			want: app.PriceCheckList{Threshold: dec(t, "0.01"), Checks: checks[:1]},
		},
		{name: "no prices compared in this core", want: app.PriceCheckList{}},
		{name: "the checks can't be read", source: fakePriceChecks{fail: errChecksDown}, wantErr: errChecksDown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newFixture(t)
			var opts []app.Option
			if tt.source != nil {
				opts = append(opts, app.WithPriceChecks(tt.source))
			}
			v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), opts...)
			if err != nil {
				t.Fatalf("app.New: %v", err)
			}
			got, err := v.PriceChecks(t.Context(), tt.status)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("PriceChecks() error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("PriceChecks() (-want +got):\n%s", diff)
			}
		})
	}
}
