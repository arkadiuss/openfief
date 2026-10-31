package pl_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestCouponDates(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, series, purchased string
		want                    []string
		wantErr                 error
	}{
		{
			// COI runs four yearly periods; the fourth one's interest comes with the redemption.
			name: "COI pays three yearly coupons", series: "COI0129", purchased: "2025-01-10",
			want: []string{"2026-01-10", "2027-01-10", "2028-01-10"},
		},
		{
			name: "ROR pays eleven monthly coupons", series: "ROR0126", purchased: "2025-01-31",
			want: []string{
				"2025-02-28", "2025-03-31", "2025-04-30", "2025-05-31", "2025-06-30", "2025-07-31",
				"2025-08-31", "2025-09-30", "2025-10-31", "2025-11-30", "2025-12-31",
			},
		},
		{name: "EDO capitalizes and pays no coupon", series: "EDO0535", purchased: "2025-05-05"},
		{name: "OTS pays its interest at maturity only", series: "OTS0825", purchased: "2025-05-05"},
		{
			name: "a purchase in another month than the series names", series: "COI0129", purchased: "2025-02-10",
			wantErr: pl.ErrSeriesMismatch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pl.CouponDates(series(t, tt.series), date(t, tt.purchased))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CouponDates error = %v, want %v", err, tt.wantErr)
			}
			var want []kernel.Date
			for _, d := range tt.want {
				want = append(want, date(t, d))
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("CouponDates (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCoupon(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, series, first, margin, purchased string
		idx                                    map[int]string
		period                                 int
		wantDate, wantPerBond                  string
		wantErr                                error
	}{
		{
			// N × r = 100 × 0.065 = 6.50, as the published COI first-year coupons.
			name: "COI first year at the first-period rate", series: "COI0129", first: "0.065", margin: "0.0125",
			purchased: "2025-01-10", period: 1, wantDate: "2026-01-10", wantPerBond: "6.50",
		},
		{
			// r2 = CPI 4.03% + margin 1.25% = 5.28%: 100 × 0.0528 = 5.28.
			name: "COI second year at CPI plus margin", series: "COI0129", first: "0.065", margin: "0.0125",
			idx: map[int]string{2: "0.0403"}, purchased: "2025-01-10", period: 2, wantDate: "2027-01-10", wantPerBond: "5.28",
		},
		{
			// A negative CPI counts as zero: the margin alone, 100 × 0.0125 = 1.25.
			name: "COI deflation floors the index at zero", series: "COI0129", first: "0.065", margin: "0.0125",
			idx: map[int]string{3: "-0.01"}, purchased: "2025-01-10", period: 3, wantDate: "2028-01-10", wantPerBond: "1.25",
		},
		{
			// 100 × 0.0575 / 12 = 0.479166... -> 0.48, whatever the month's day count.
			name: "ROR first month: r over twelve, rounded half up", series: "ROR0126", first: "0.0575", margin: "0",
			purchased: "2025-01-31", period: 1, wantDate: "2025-02-28", wantPerBond: "0.48",
		},
		{
			// 100 × 0.0525 / 12 = 0.4375 -> 0.44 (half up).
			name: "ROR later month at the reference rate", series: "ROR0126", first: "0.0575", margin: "0",
			idx: map[int]string{4: "0.0525"}, purchased: "2025-01-31", period: 4, wantDate: "2025-05-31", wantPerBond: "0.44",
		},
		{
			name: "the last period is paid at maturity, not as a coupon", series: "COI0129", first: "0.065", margin: "0.0125",
			idx: map[int]string{4: "0.03"}, purchased: "2025-01-10", period: 4, wantErr: pl.ErrInvalidTerms,
		},
		{
			name: "a capitalized series pays none", series: "EDO0535", first: "0.0625", margin: "0.02",
			purchased: "2025-05-05", period: 1, wantErr: pl.ErrInvalidTerms,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			terms := pl.Terms{
				Series: series(t, tt.series), FirstPeriodRate: dec(t, tt.first), Margin: dec(t, tt.margin),
				EarlyRedemptionFee: plnMoney(t, "0.70"),
			}
			got, err := terms.Coupon(indices(t, tt.idx), date(t, tt.purchased), tt.period)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Coupon error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			want := pl.CouponPayment{Period: tt.period, Date: date(t, tt.wantDate), PerBond: plnMoney(t, tt.wantPerBond)}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Coupon (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCouponWithoutItsReading(t *testing.T) {
	t.Parallel()
	terms := pl.Terms{
		Series: series(t, "COI0129"), FirstPeriodRate: dec(t, "0.065"), Margin: dec(t, "0.0125"), EarlyRedemptionFee: plnMoney(t, "0.70"),
	}
	_, err := terms.Coupon(pl.Indices{}, date(t, "2025-01-10"), 2)
	if _, ok := errors.AsType[*pl.MissingIndexError](err); !ok {
		t.Errorf("Coupon error = %v, want a *MissingIndexError", err)
	}
}
