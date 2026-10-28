package pl_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
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

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func plnMoney(t *testing.T, amount string) kernel.Money {
	t.Helper()
	c, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	return kernel.NewMoney(dec(t, amount), c)
}

func series(t *testing.T, code string) pl.Series {
	t.Helper()
	s, err := pl.ParseSeries(code)
	if err != nil {
		t.Fatalf("ParseSeries(%q): %v", code, err)
	}
	return s
}

// indices builds period index readings from "period=reading" pairs.
func indices(t *testing.T, readings map[int]string) pl.Indices {
	t.Helper()
	idx := pl.Indices{}
	for n, r := range readings {
		idx[n] = dec(t, r)
	}
	return idx
}
