package kernel_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestQuantityArithmetic(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b string
		op   func(a, b kernel.Quantity) kernel.Quantity
		want string
	}{
		{name: "add fractional shares", a: "10.5", b: "0.25", op: kernel.Quantity.Add, want: "10.75"},
		{name: "sub to zero", a: "3", b: "3.000", op: kernel.Quantity.Sub, want: "0"},
		{name: "neg", a: "7", b: "0", op: func(a, _ kernel.Quantity) kernel.Quantity { return a.Neg() }, want: "-7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(qty(t, tt.want), tt.op(qty(t, tt.a), qty(t, tt.b))); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestQuantityCompare(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		a, b     string
		wantCmp  int
		wantSign int
		wantZero bool
	}{
		{name: "less", a: "1", b: "1.5", wantCmp: -1, wantSign: 1},
		{name: "equal across scales", a: "2.0", b: "2", wantCmp: 0, wantSign: 1},
		{name: "short position", a: "-4", b: "0", wantCmp: -1, wantSign: -1},
		{name: "zero", a: "0", b: "0", wantCmp: 0, wantZero: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := qty(t, tt.a), qty(t, tt.b)
			got := []any{a.Cmp(b), a.Sign(), a.IsZero(), a.Equal(b), a.String()}
			want := []any{tt.wantCmp, tt.wantSign, tt.wantZero, tt.wantCmp == 0, dec(t, tt.a).String()}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Cmp, Sign, IsZero, Equal, String (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseQuantity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{name: "whole", in: "100", want: "100"},
		{name: "fractional", in: "0.125", want: "0.125"},
		{name: "invalid", in: "ten", wantErr: kernel.ErrInvalidDecimal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := kernel.ParseQuantity(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParseQuantity(%q) error (-want +got):\n%s", tt.in, diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got.String()); diff != "" {
				t.Errorf("ParseQuantity(%q) (-want +got):\n%s", tt.in, diff)
			}
		})
	}
}
