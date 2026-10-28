package valuation_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

func dec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

func cur(t *testing.T, code string) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", code, err)
	}
	return c
}

func money(t *testing.T, amount, code string) kernel.Money {
	t.Helper()
	return kernel.NewMoney(dec(t, amount), cur(t, code))
}

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

// rate builds a synthetic FX rate; source defaults to "nbp-a".
func rate(t *testing.T, code, on, value string) valuation.FXRate {
	t.Helper()
	return valuation.FXRate{Currency: cur(t, code), Date: date(t, on), Rate: dec(t, value), Source: "nbp-a"}
}
