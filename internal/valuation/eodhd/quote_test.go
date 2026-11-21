package eodhd_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
)

func TestExchangeCode(t *testing.T) {
	t.Parallel()
	for venue, want := range map[string]string{
		"NASDAQ": "US", "NYSE": "US", "nyse arca": "US", "US": "US", "OTC": "US",
		"XETRA": "XETRA", "LSE": "LSE", "WAR": "WAR", "": "",
	} {
		if got := eodhd.ExchangeCode(venue); got != want {
			t.Errorf("ExchangeCode(%q) = %q, want %q", venue, got, want)
		}
	}
}

func TestQuoteTicker(t *testing.T) {
	t.Parallel()
	l := eodhd.Listing{Code: "SYNQ", Exchange: "NASDAQ"}
	if got := l.QuoteTicker(); got != "SYNQ.US" {
		t.Errorf("QuoteTicker = %q, want SYNQ.US, the code its end-of-day API takes", got)
	}
}

// TestMatchQuery checks that a ticker or an ISIN names one listing, the US
// venue first, or several ISINs the person must choose between, or none.
func TestMatchQuery(t *testing.T) {
	t.Parallel()
	synq, twinA, twinB := isin(t, "XX00000SYNQ9"), isin(t, "XX00000TWIN8"), isin(t, "XX0000TWINB5")
	nasdaq := eodhd.Listing{ISIN: synq, Code: "SYNQ", Exchange: "NASDAQ", Currency: "USD"}
	xetra := eodhd.Listing{ISIN: synq, Code: "SYNQ", Exchange: "XETRA", Currency: "EUR"}
	lse := eodhd.Listing{ISIN: synq, Code: "SYNQ", Exchange: "LSE", Currency: "GBX"}
	frankfurt := eodhd.Listing{ISIN: synq, Code: "SYNQ", Exchange: "F", Currency: "EUR"}
	alpha := eodhd.Listing{ISIN: twinA, Code: "TWIN", Exchange: "XETRA", Currency: "EUR"}
	beta := eodhd.Listing{ISIN: twinB, Code: "TWIN", Exchange: "LSE", Currency: "GBX"}
	fuzzy := eodhd.Listing{ISIN: twinA, Code: "TWINS", Exchange: "XETRA", Currency: "EUR"}
	tests := []struct {
		name     string
		listings []eodhd.Listing
		query    string
		want     eodhd.Listing
		wantList []eodhd.Listing
		wantErr  error
	}{
		{name: "the US venue wins", listings: []eodhd.Listing{xetra, nasdaq, lse}, query: "SYNQ", want: nasdaq},
		{name: "the case of a ticker doesn't matter", listings: []eodhd.Listing{nasdaq}, query: "synq", want: nasdaq},
		{name: "an ISIN", listings: []eodhd.Listing{lse, xetra}, query: "XX00000SYNQ9", want: lse},
		{name: "a supported exchange before another", listings: []eodhd.Listing{frankfurt, xetra}, query: "SYNQ", want: xetra},
		{name: "a repeated end-of-day ticker is one listing", listings: []eodhd.Listing{nasdaq, {ISIN: synq, Code: "SYNQ", Exchange: "NYSE"}}, query: "SYNQ", want: nasdaq},
		{name: "two ISINs under one ticker", listings: []eodhd.Listing{alpha, beta}, query: "TWIN", wantList: []eodhd.Listing{beta, alpha}, wantErr: eodhd.ErrAmbiguous},
		{name: "a fuzzy match on the name is no match", listings: []eodhd.Listing{fuzzy}, query: "TWIN", wantErr: eodhd.ErrNotListed},
		{name: "nothing found", query: "NOPE", wantErr: eodhd.ErrNotListed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			match, err := eodhd.MatchQuery(tt.listings, tt.query)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("MatchQuery error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, match.Listing); diff != "" {
				t.Errorf("MatchQuery (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantList, match.Candidates); diff != "" {
				t.Errorf("MatchQuery candidates (-want +got):\n%s", diff)
			}
		})
	}
}

// TestMatchQueryCapsTheCandidates checks that a ticker many ISINs share is
// answered with at most ten, one per ISIN.
func TestMatchQueryCapsTheCandidates(t *testing.T) {
	t.Parallel()
	listings := make([]eodhd.Listing, 0, 24)
	for n := range 12 {
		listings = append(listings, eodhd.Listing{ISIN: isin(t, syntheticISIN(n)), Code: "MANY", Exchange: "XETRA"}, eodhd.Listing{ISIN: isin(t, syntheticISIN(n)), Code: "MANY", Exchange: "LSE"})
	}
	match, err := eodhd.MatchQuery(listings, "MANY")
	if !errors.Is(err, eodhd.ErrAmbiguous) || len(match.Candidates) != 10 {
		t.Fatalf("MatchQuery = %d candidates, %v; want 10 and ErrAmbiguous", len(match.Candidates), err)
	}
}

// syntheticISIN returns the n-th synthetic ISIN, XX0000000nnC with its check
// digit C, as ISO 6166 computes it.
func syntheticISIN(n int) string {
	body := fmt.Sprintf("XX0000000%02d", n)
	var digits []int
	for _, r := range body {
		if r >= 'A' {
			v := int(r-'A') + 10
			digits = append(digits, v/10, v%10)
			continue
		}
		digits = append(digits, int(r-'0'))
	}
	sum := 0
	for i := range digits {
		d := digits[len(digits)-1-i]
		if i%2 == 0 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return body + strconv.Itoa((10-sum%10)%10)
}
