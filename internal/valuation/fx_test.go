package valuation_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// Synthetic table: Friday 2026-03-06 and Monday 2026-03-09.
func table(t *testing.T) []valuation.FXRate {
	t.Helper()
	return []valuation.FXRate{
		rate(t, "USD", "2026-03-06", "4.0000"),
		rate(t, "EUR", "2026-03-09", "4.3000"),
		rate(t, "USD", "2026-03-09", "4.0100"),
	}
}

func TestRateOnOrBefore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		currency string
		cutoff   string
		want     string // rate date
		wantRate string
		wantErr  error
	}{
		{name: "same-day table", currency: "USD", cutoff: "2026-03-09", want: "2026-03-09", wantRate: "4.01"},
		{name: "weekend takes Friday's table", currency: "USD", cutoff: "2026-03-08", want: "2026-03-06", wantRate: "4"},
		{name: "later table is never used", currency: "USD", cutoff: "2026-03-07", want: "2026-03-06", wantRate: "4"},
		{name: "other currency", currency: "EUR", cutoff: "2026-03-10", want: "2026-03-09", wantRate: "4.3"},
		{name: "before the first table", currency: "USD", cutoff: "2026-03-05", wantErr: valuation.ErrNoRate},
		{name: "currency not in the table", currency: "GBP", cutoff: "2026-03-09", wantErr: valuation.ErrNoRate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := valuation.RateOnOrBefore(table(t), cur(t, tt.currency), date(t, tt.cutoff))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			want := rate(t, tt.currency, tt.want, tt.wantRate)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestRateOnOrBeforeConflicts(t *testing.T) {
	t.Parallel()
	other := rate(t, "USD", "2026-03-09", "4.0200")
	other.Source = "other"
	_, err := valuation.RateOnOrBefore(append(table(t), other), cur(t, "USD"), date(t, "2026-03-09"))
	if !errors.Is(err, valuation.ErrConflictingRates) {
		t.Errorf("error = %v, want ErrConflictingRates", err)
	}
	same := rate(t, "USD", "2026-03-09", "4.01")
	same.Source = "copy"
	got, err := valuation.RateOnOrBefore(append(table(t), same), cur(t, "USD"), date(t, "2026-03-09"))
	if err != nil {
		t.Fatalf("equal duplicate rates: %v", err)
	}
	if diff := cmp.Diff(rate(t, "USD", "2026-03-09", "4.01"), got); diff != "" {
		t.Errorf("first of equal duplicates (-want +got):\n%s", diff)
	}
}

// Polish tax takes the table of the last business day before the event:
// income on Monday 2026-03-09 converts at Friday's 4.0000, not Monday's 4.0100.
func TestTaxRateUsesTheTableBeforeTheEvent(t *testing.T) {
	t.Parallel()
	got, err := valuation.RateOnOrBefore(table(t), cur(t, "USD"), pl.TaxRateCutoff(date(t, "2026-03-09")))
	if err != nil {
		t.Fatalf("RateOnOrBefore: %v", err)
	}
	if diff := cmp.Diff(rate(t, "USD", "2026-03-06", "4"), got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func TestToPLN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		amount  string
		code    string
		rate    string
		mode    kernel.RoundingMode
		want    string
		wantErr error
	}{
		{name: "123.45 USD x 4.0123 = 495.318435, half up", amount: "123.45", code: "USD", rate: "4.0123", mode: kernel.RoundHalfUp, want: "495.32"},
		{name: "123.45 USD x 4.0123 = 495.318435, down", amount: "123.45", code: "USD", rate: "4.0123", mode: kernel.RoundDown, want: "495.31"},
		{name: "10.05 USD x 4.5 = 45.225, tie half up", amount: "10.05", code: "USD", rate: "4.5", mode: kernel.RoundHalfUp, want: "45.23"},
		{name: "10.05 USD x 4.5 = 45.225, tie half even", amount: "10.05", code: "USD", rate: "4.5", mode: kernel.RoundHalfEven, want: "45.22"},
		{name: "negative amount keeps its sign", amount: "-10", code: "USD", rate: "4.0123", mode: kernel.RoundHalfUp, want: "-40.12"},
		{name: "zero rate", amount: "1", code: "USD", rate: "0", mode: kernel.RoundHalfUp, wantErr: valuation.ErrInvalidRate},
		{name: "unknown rounding mode", amount: "1", code: "USD", rate: "4", mode: kernel.RoundingMode(0), wantErr: kernel.ErrUnknownRoundingMode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := valuation.ToPLN(money(t, tt.amount, tt.code), rate(t, tt.code, "2026-03-09", tt.rate), tt.mode)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(money(t, tt.want, "PLN"), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestToPLNNeedsTheAmountsCurrency(t *testing.T) {
	t.Parallel()
	_, err := valuation.ToPLN(money(t, "1", "EUR"), rate(t, "USD", "2026-03-09", "4"), kernel.RoundHalfUp)
	var mismatch *kernel.CurrencyMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("error = %v, want *CurrencyMismatchError", err)
	}
	if diff := cmp.Diff(kernel.CurrencyMismatchError{Left: cur(t, "EUR"), Right: cur(t, "USD")}, *mismatch); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func TestConvertToPLN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		amount  string
		code    string
		cutoff  string
		want    string
		wantErr error
	}{
		{name: "PLN needs no rate, only grosze: 10.005 -> 10.01", amount: "10.005", code: "PLN", cutoff: "2026-01-01", want: "10.01"},
		{name: "USD on Sunday uses Friday: 100 x 4.0000", amount: "100", code: "USD", cutoff: "2026-03-08", want: "400"},
		{name: "EUR on Monday: 2.50 x 4.3000 = 10.75", amount: "2.50", code: "EUR", cutoff: "2026-03-09", want: "10.75"},
		{name: "no rate yet", amount: "1", code: "EUR", cutoff: "2026-03-08", wantErr: valuation.ErrNoRate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := valuation.ConvertToPLN(money(t, tt.amount, tt.code), table(t), date(t, tt.cutoff), kernel.RoundHalfUp)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(money(t, tt.want, "PLN"), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
