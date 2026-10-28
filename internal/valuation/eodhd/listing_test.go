package eodhd_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
)

func TestSupportedExchange(t *testing.T) {
	t.Parallel()
	for code, want := range map[string]bool{
		"WAR": true, "LSE": true, "XETRA": true, "PA": true, "AS": true, "BR": true, "LS": true, "IR": true, "MI": true, "OL": true,
		"F": false, "US": false, "INDX": false, "war": false, "": false,
	} {
		if got := eodhd.SupportedExchange(code); got != want {
			t.Errorf("SupportedExchange(%q) = %v, want %v", code, got, want)
		}
	}
}

func TestPick(t *testing.T) {
	t.Parallel()
	synt, other := isin(t, "XX00000SYNT3"), isin(t, "XX0000FAKEW1")
	xetra := eodhd.Listing{ISIN: synt, Code: "SYNT", Exchange: "XETRA", Currency: "EUR"}
	lse := eodhd.Listing{ISIN: synt, Code: "SYNT", Exchange: "LSE", Currency: "GBX"}
	frankfurt := eodhd.Listing{ISIN: synt, Code: "SYN1", Exchange: "F", Currency: "EUR"}
	war := eodhd.Listing{ISIN: other, Code: "FAKEW", Exchange: "WAR", Currency: "PLN"}
	tests := []struct {
		name     string
		listings []eodhd.Listing
		exchange string
		want     eodhd.Listing
		wantErr  error
	}{
		{name: "the only supported listing", listings: []eodhd.Listing{frankfurt, xetra, war}, want: xetra},
		{name: "the exchange asked for", listings: []eodhd.Listing{xetra, lse, frankfurt}, exchange: "LSE", want: lse},
		{name: "a repeated ticker is one listing", listings: []eodhd.Listing{xetra, xetra}, want: xetra},
		{name: "two supported exchanges, none asked for", listings: []eodhd.Listing{xetra, lse}, wantErr: eodhd.ErrAmbiguous},
		{name: "only on an unsupported exchange", listings: []eodhd.Listing{frankfurt}, wantErr: eodhd.ErrNotListed},
		{name: "not on the exchange asked for", listings: []eodhd.Listing{xetra}, exchange: "WAR", wantErr: eodhd.ErrNotListed},
		{name: "an unsupported exchange asked for", listings: []eodhd.Listing{frankfurt}, exchange: "F", wantErr: eodhd.ErrNotListed},
		{name: "only another ISIN", listings: []eodhd.Listing{war}, wantErr: eodhd.ErrNotListed},
		{name: "nothing found", wantErr: eodhd.ErrNotListed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := eodhd.Pick(tt.listings, synt, tt.exchange)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Pick error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Pick (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTicker(t *testing.T) {
	t.Parallel()
	if got := (eodhd.Listing{Code: "FAKEW", Exchange: "WAR"}).Ticker(); got != "FAKEW.WAR" {
		t.Errorf("Ticker = %q, want FAKEW.WAR", got)
	}
}

func TestQuote(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		currency string
		closing  string
		want     string
		wantErr  error
	}{
		{name: "euro", currency: "EUR", closing: "101.075", want: "101.075 EUR"},
		{name: "zloty", currency: "PLN", closing: "42.5", want: "42.5 PLN"},
		{name: "pence become pounds exactly", currency: "GBX", closing: "8412.5", want: "84.125 GBP"},
		{name: "pounds stay pounds", currency: "GBP", closing: "84.125", want: "84.125 GBP"},
		{name: "no currency", currency: "", closing: "1", wantErr: eodhd.ErrQuoteCurrency},
		{name: "not a code", currency: "Euro", closing: "1", wantErr: eodhd.ErrQuoteCurrency},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			closing, err := kernel.ParseDecimal(tt.closing)
			if err != nil {
				t.Fatalf("ParseDecimal: %v", err)
			}
			got, err := eodhd.Listing{Code: "SYNT", Exchange: "LSE", Currency: tt.currency}.Quote(closing)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Quote error (-want +got):\n%s", diff)
			}
			if err == nil && got.String() != tt.want {
				t.Errorf("Quote = %s, want %s", got, tt.want)
			}
		})
	}
}
