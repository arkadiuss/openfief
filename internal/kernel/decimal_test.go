package kernel_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestParseDecimal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{name: "integer", in: "42", want: "42"},
		{name: "fraction", in: "-12.345", want: "-12.345"},
		{name: "trailing zeros dropped", in: "1.500", want: "1.5"},
		{name: "many digits stay exact", in: "123456789012345678901234567890.123456789", want: "123456789012345678901234567890.123456789"},
		{name: "empty", in: "", wantErr: kernel.ErrInvalidDecimal},
		{name: "comma separator", in: "1,5", wantErr: kernel.ErrInvalidDecimal},
		{name: "text", in: "abc", wantErr: kernel.ErrInvalidDecimal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := kernel.ParseDecimal(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParseDecimal(%q) error (-want +got):\n%s", tt.in, diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got.String()); diff != "" {
				t.Errorf("ParseDecimal(%q) (-want +got):\n%s", tt.in, diff)
			}
		})
	}
}

func TestDecimalArithmeticIsExact(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b string
		op   func(a, b kernel.Decimal) kernel.Decimal
		want string
	}{
		{name: "0.1 + 0.2", a: "0.1", b: "0.2", op: kernel.Decimal.Add, want: "0.3"},
		{name: "1 - 0.9", a: "1", b: "0.9", op: kernel.Decimal.Sub, want: "0.1"},
		{name: "neg", a: "2.5", op: func(a, _ kernel.Decimal) kernel.Decimal { return a.Neg() }, want: "-2.5"},
		{name: "mul keeps every digit", a: "1.005", b: "4.1234", op: kernel.Decimal.Mul, want: "4.144017"},
		{name: "NewDecimal", op: func(_, _ kernel.Decimal) kernel.Decimal { return kernel.NewDecimal(1234, -2) }, want: "12.34"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var a, b kernel.Decimal
			if tt.a != "" {
				a = dec(t, tt.a)
			}
			if tt.b != "" {
				b = dec(t, tt.b)
			}
			if diff := cmp.Diff(dec(t, tt.want), tt.op(a, b)); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestDecimalRound(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		places  int32
		mode    kernel.RoundingMode
		want    string
		wantErr error
	}{
		{name: "half even, tie to even down", in: "2.345", places: 2, mode: kernel.RoundHalfEven, want: "2.34"},
		{name: "half even, tie to even up", in: "2.355", places: 2, mode: kernel.RoundHalfEven, want: "2.36"},
		{name: "half even, negative tie", in: "-2.345", places: 2, mode: kernel.RoundHalfEven, want: "-2.34"},
		{name: "half up, tie away from zero", in: "2.345", places: 2, mode: kernel.RoundHalfUp, want: "2.35"},
		{name: "half up, negative tie away from zero", in: "-2.345", places: 2, mode: kernel.RoundHalfUp, want: "-2.35"},
		{name: "down truncates", in: "2.349", places: 2, mode: kernel.RoundDown, want: "2.34"},
		{name: "down truncates toward zero", in: "-2.349", places: 2, mode: kernel.RoundDown, want: "-2.34"},
		{name: "zero places", in: "7.5", places: 0, mode: kernel.RoundHalfEven, want: "8"},
		{name: "unknown mode", in: "1", places: 0, mode: kernel.RoundingMode(0), wantErr: kernel.ErrUnknownRoundingMode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := dec(t, tt.in).Round(tt.places, tt.mode)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Round error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(dec(t, tt.want), got); diff != "" {
				t.Errorf("Round(%s, %d) (-want +got):\n%s", tt.in, tt.places, diff)
			}
		})
	}
}

func TestDecimalCompare(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		a, b     string
		wantCmp  int
		wantSign int
		wantZero bool
	}{
		{name: "less", a: "1.5", b: "2", wantCmp: -1, wantSign: 1},
		{name: "equal across scales", a: "1.50", b: "1.5", wantCmp: 0, wantSign: 1},
		{name: "greater negative", a: "-1", b: "-2", wantCmp: 1, wantSign: -1},
		{name: "zero", a: "0.00", b: "0", wantCmp: 0, wantSign: 0, wantZero: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := dec(t, tt.a), dec(t, tt.b)
			got := []any{a.Cmp(b), a.Sign(), a.IsZero(), a.Equal(b)}
			want := []any{tt.wantCmp, tt.wantSign, tt.wantZero, tt.wantCmp == 0}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Cmp, Sign, IsZero, Equal (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDecimalDiv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		a, b    string
		places  int32
		mode    kernel.RoundingMode
		want    string
		wantErr error
	}{
		{name: "exact", a: "10", b: "4", places: 2, mode: kernel.RoundHalfEven, want: "2.5"},
		{name: "a third, half even", a: "1", b: "3", places: 4, mode: kernel.RoundHalfEven, want: "0.3333"},
		{name: "two thirds, half even rounds up", a: "2", b: "3", places: 2, mode: kernel.RoundHalfEven, want: "0.67"},
		{name: "two thirds, down truncates", a: "2", b: "3", places: 2, mode: kernel.RoundDown, want: "0.66"},
		{name: "half even, tie to even down", a: "0.125", b: "1", places: 2, mode: kernel.RoundHalfEven, want: "0.12"},
		{name: "half even, tie to even up", a: "0.375", b: "1", places: 2, mode: kernel.RoundHalfEven, want: "0.38"},
		{name: "half up, tie away from zero", a: "0.125", b: "1", places: 2, mode: kernel.RoundHalfUp, want: "0.13"},
		{name: "half up, negative tie away from zero", a: "-1", b: "8", places: 2, mode: kernel.RoundHalfUp, want: "-0.13"},
		{name: "half even, negative divisor", a: "2", b: "-3", places: 2, mode: kernel.RoundHalfEven, want: "-0.67"},
		{name: "down, negative toward zero", a: "-2", b: "3", places: 2, mode: kernel.RoundDown, want: "-0.66"},
		{name: "zero places", a: "7", b: "2", places: 0, mode: kernel.RoundHalfEven, want: "4"},
		{name: "zero dividend", a: "0", b: "7", places: 2, mode: kernel.RoundHalfUp, want: "0"},
		{name: "division by zero", a: "1", b: "0", places: 2, mode: kernel.RoundHalfEven, wantErr: kernel.ErrDivisionByZero},
		{name: "unknown mode", a: "1", b: "3", places: 2, mode: kernel.RoundingMode(0), wantErr: kernel.ErrUnknownRoundingMode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := dec(t, tt.a).Div(dec(t, tt.b), tt.places, tt.mode)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Div error (-want +got):\n%s", diff)
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
