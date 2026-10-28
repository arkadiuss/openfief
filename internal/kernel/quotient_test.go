package kernel_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestDecimalQuo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		a, b    string
		places  int32
		mode    kernel.RoundingMode
		want    string
		wantErr error
	}{
		{name: "exact 1 / 4", a: "1", b: "4", places: 2, mode: kernel.RoundHalfEven, want: "0.25"},
		{name: "2 / 3 half-up", a: "2", b: "3", places: 2, mode: kernel.RoundHalfUp, want: "0.67"},
		{name: "2 / 3 down", a: "2", b: "3", places: 2, mode: kernel.RoundDown, want: "0.66"},
		{name: "-2 / 3 half-up away from zero", a: "-2", b: "3", places: 2, mode: kernel.RoundHalfUp, want: "-0.67"},
		{name: "2 / -3 down toward zero", a: "2", b: "-3", places: 2, mode: kernel.RoundDown, want: "-0.66"},
		{name: "tie 0.125 half-up", a: "1", b: "8", places: 2, mode: kernel.RoundHalfUp, want: "0.13"},
		{name: "tie 0.125 half-even to 2", a: "1", b: "8", places: 2, mode: kernel.RoundHalfEven, want: "0.12"},
		{name: "tie 0.375 half-even to 8", a: "3", b: "8", places: 2, mode: kernel.RoundHalfEven, want: "0.38"},
		{name: "tie -0.125 half-even", a: "-1", b: "8", places: 2, mode: kernel.RoundHalfEven, want: "-0.12"},
		{name: "above tie half-even rounds up", a: "1.2501", b: "10", places: 2, mode: kernel.RoundHalfEven, want: "0.13"},
		{name: "bond accrual 100 x 0.0535 x 100 / 365", a: "535", b: "365", places: 2, mode: kernel.RoundHalfUp, want: "1.47"},
		{name: "zero places", a: "7", b: "2", places: 0, mode: kernel.RoundHalfEven, want: "4"},
		{name: "division by zero", a: "1", b: "0", places: 2, mode: kernel.RoundHalfUp, wantErr: kernel.ErrDivisionByZero},
		{name: "unknown mode", a: "1", b: "4", places: 2, mode: kernel.RoundingMode(99), wantErr: kernel.ErrUnknownRoundingMode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := dec(t, tt.a).Quo(dec(t, tt.b), tt.places, tt.mode)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Quo error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(dec(t, tt.want), got); diff != "" {
				t.Errorf("%s / %s (-want +got):\n%s", tt.a, tt.b, diff)
			}
		})
	}
}
