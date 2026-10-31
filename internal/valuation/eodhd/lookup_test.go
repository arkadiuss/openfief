package eodhd_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
)

// TestCountryExchange checks that every country with an exchange maps to a
// supported one, and others to none.
func TestCountryExchange(t *testing.T) {
	t.Parallel()
	for country, want := range map[string]string{
		"PL": "WAR", "GB": "LSE", "DE": "XETRA", "FR": "PA", "NL": "AS", "BE": "BR", "PT": "LS", "IE": "IR", "IT": "MI",
		"NO": "OL", "US": "", "UK": "", "": "",
	} {
		got := eodhd.CountryExchange(country)
		if got != want || (got != "" && !eodhd.SupportedExchange(got)) {
			t.Errorf("CountryExchange(%q) = %q, want %q on a supported exchange", country, got, want)
		}
	}
}

// TestFindByCode checks that a code finds its one ISIN on the exchange,
// ignoring case and repeats, and nothing when none or two ISINs match.
func TestFindByCode(t *testing.T) {
	t.Parallel()
	synt, other := isin(t, "XX00000SYNT3"), isin(t, "XX0000FAKEW1")
	xetra := eodhd.Listing{ISIN: synt, Code: "SYNT", Exchange: "XETRA", Name: "Synthetic"}
	tests := []struct {
		name     string
		listings []eodhd.Listing
		code     string
		want     bool
	}{
		{name: "one", listings: []eodhd.Listing{xetra, {ISIN: other, Code: "SYNTX", Exchange: "XETRA"}}, code: "synt", want: true},
		{name: "repeated", listings: []eodhd.Listing{xetra, xetra}, code: "SYNT", want: true},
		{name: "other exchange", listings: []eodhd.Listing{{ISIN: synt, Code: "SYNT", Exchange: "LSE"}}, code: "SYNT"},
		{name: "two ISINs", listings: []eodhd.Listing{xetra, {ISIN: other, Code: "SYNT", Exchange: "XETRA"}}, code: "SYNT"},
		{name: "none", code: "SYNT"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := eodhd.FindByCode(tt.listings, tt.code, "XETRA")
			if ok != tt.want || (ok && got != xetra) {
				t.Errorf("FindByCode = %v, %v; want %v", got, ok, tt.want)
			}
		})
	}
}
