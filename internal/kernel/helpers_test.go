package kernel_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
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

func qty(t *testing.T, s string) kernel.Quantity {
	t.Helper()
	return kernel.NewQuantity(dec(t, s))
}

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}
