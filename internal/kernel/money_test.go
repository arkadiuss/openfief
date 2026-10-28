package kernel_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestMoneyArithmeticInOneCurrency(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b string
		op   func(a, b kernel.Money) (kernel.Money, error)
		want string
	}{
		{name: "add", a: "10.10", b: "0.20", op: kernel.Money.Add, want: "10.30"},
		{name: "sub below zero", a: "1", b: "2.5", op: kernel.Money.Sub, want: "-1.5"},
		{name: "neg", a: "3", b: "0", op: func(a, _ kernel.Money) (kernel.Money, error) { return a.Neg(), nil }, want: "-3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.op(money(t, tt.a, "PLN"), money(t, tt.b, "PLN"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(money(t, tt.want, "PLN"), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestMoneyMulAndRound(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		price   string
		factor  string
		places  int32
		mode    kernel.RoundingMode
		wantMul string
		want    string
	}{
		{name: "price times fractional quantity, half even", price: "101.37", factor: "3.125", places: 2, mode: kernel.RoundHalfEven, wantMul: "316.78125", want: "316.78"},
		{name: "tie rounds half up", price: "0.5", factor: "0.25", places: 2, mode: kernel.RoundHalfUp, wantMul: "0.125", want: "0.13"},
		{name: "tie rounds half even", price: "0.5", factor: "0.25", places: 2, mode: kernel.RoundHalfEven, wantMul: "0.125", want: "0.12"},
		{name: "round down", price: "10", factor: "0.0333", places: 2, mode: kernel.RoundDown, wantMul: "0.333", want: "0.33"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			product := money(t, tt.price, "USD").Mul(qty(t, tt.factor).Decimal())
			if diff := cmp.Diff(money(t, tt.wantMul, "USD"), product); diff != "" {
				t.Errorf("Mul (-want +got):\n%s", diff)
			}
			got, err := product.Round(tt.places, tt.mode)
			if err != nil {
				t.Fatalf("Round: %v", err)
			}
			if diff := cmp.Diff(money(t, tt.want, "USD"), got); diff != "" {
				t.Errorf("Round (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMoneyNeverMixesCurrencies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		left, right kernel.Currency
	}{
		{name: "PLN and USD", left: cur(t, "PLN"), right: cur(t, "USD")},
		{name: "USD and PLN", left: cur(t, "USD"), right: cur(t, "PLN")},
		{name: "no currency and EUR", left: kernel.Currency{}, right: cur(t, "EUR")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := kernel.NewMoney(dec(t, "1"), tt.left)
			b := kernel.NewMoney(dec(t, "1"), tt.right)
			_, addErr := a.Add(b)
			_, subErr := a.Sub(b)
			_, cmpErr := a.Cmp(b)
			for op, err := range map[string]error{"Add": addErr, "Sub": subErr, "Cmp": cmpErr} {
				var mismatch *kernel.CurrencyMismatchError
				if !errors.As(err, &mismatch) {
					t.Errorf("%s error = %v, want *CurrencyMismatchError", op, err)
					continue
				}
				got := []string{mismatch.Left.String(), mismatch.Right.String()}
				if diff := cmp.Diff([]string{tt.left.String(), tt.right.String()}, got); diff != "" {
					t.Errorf("%s mismatch currencies (-want +got):\n%s", op, diff)
				}
			}
		})
	}
}

func TestMoneyKeepsOriginalCurrency(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		code   string
		amount string
		want   string
	}{
		{name: "EUR", code: "EUR", amount: "12.50", want: "12.5 EUR"},
		{name: "USD", code: "USD", amount: "-3", want: "-3 USD"},
		{name: "PLN", code: "PLN", amount: "0.01", want: "0.01 PLN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := kernel.ParseMoney(tt.amount, cur(t, tt.code))
			if err != nil {
				t.Fatalf("ParseMoney: %v", err)
			}
			sum, err := m.Add(m)
			if err != nil {
				t.Fatalf("Add: %v", err)
			}
			rounded, err := m.Mul(dec(t, "4.0123")).Round(2, kernel.RoundHalfEven)
			if err != nil {
				t.Fatalf("Round: %v", err)
			}
			got := []string{m.String(), m.Currency().String(), sum.Currency().String(), m.Neg().Currency().String(), rounded.Currency().String()}
			want := []string{tt.want, tt.code, tt.code, tt.code, tt.code}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("String and result currencies (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMoneyCompare(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		a, b      string
		wantCmp   int
		wantSign  int
		wantZero  bool
		wantEqual bool
	}{
		{name: "less", a: "1", b: "2", wantCmp: -1, wantSign: 1},
		{name: "equal across scales", a: "2.50", b: "2.5", wantCmp: 0, wantSign: 1, wantEqual: true},
		{name: "negative", a: "-1", b: "-3", wantCmp: 1, wantSign: -1},
		{name: "zero", a: "0", b: "0.00", wantCmp: 0, wantZero: true, wantEqual: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := money(t, tt.a, "PLN"), money(t, tt.b, "PLN")
			c, err := a.Cmp(b)
			if err != nil {
				t.Fatalf("Cmp: %v", err)
			}
			got := []any{c, a.Sign(), a.IsZero(), a.Equal(b), a.Amount().Equal(b.Amount())}
			want := []any{tt.wantCmp, tt.wantSign, tt.wantZero, tt.wantEqual, tt.wantEqual}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Cmp, Sign, IsZero, Equal, Amount (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMoneyEqualNeedsSameCurrency(t *testing.T) {
	t.Parallel()
	if money(t, "1", "PLN").Equal(money(t, "1", "EUR")) {
		t.Error("1 PLN equals 1 EUR, want not equal")
	}
}

func TestMoneyRejectsBadInput(t *testing.T) {
	t.Parallel()
	_, parseErr := kernel.ParseMoney("12,50", cur(t, "PLN"))
	_, roundErr := money(t, "1", "PLN").Round(0, kernel.RoundingMode(99))
	got := []error{parseErr, roundErr}
	want := []error{kernel.ErrInvalidDecimal, kernel.ErrUnknownRoundingMode}
	if diff := cmp.Diff(want, got, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("errors (-want +got):\n%s", diff)
	}
}
