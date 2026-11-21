package main

import (
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// quoteWith asks a connector over a fake EODHD with key, failing with failure,
// for the quote req names, on 2025-01-06.
func quoteWith(t *testing.T, key string, failure error, req *connectorv1.LookupQuoteRequest) (*connectorv1.LookupQuoteResponse, *fakeEODHD, error) {
	t.Helper()
	var src *fakeEODHD
	c := &connector{
		secrets: secretsDir(t, key), today: func() kernel.Date { return day(t, "2025-01-06") },
		source: func(token string) marketData { src = newFakeEODHD(token); src.err = failure; return src },
	}
	resp, err := c.LookupQuote(t.Context(), connect.NewRequest(req))
	if err != nil {
		return nil, src, err
	}
	return resp.Msg, src, nil
}

func quoteRequest(query, from, to string) *connectorv1.LookupQuoteRequest {
	return &connectorv1.LookupQuoteRequest{Query: query, From: from, To: to}
}

// TestLookupQuote checks D-226's lookup: a ticker or an ISIN found is answered
// with the closes of the period in its listing's currency, the US venue
// preferred when one instrument is listed on several, only the plugin asking
// EODHD with the owner's key.
func TestLookupQuote(t *testing.T) {
	t.Parallel()
	history := []*connectorv1.QuotePoint{
		{Date: "2024-12-30", Close: "55.1"}, {Date: "2025-01-02", Close: "56.25"}, {Date: "2025-01-03", Close: "57.5"},
	}
	want := &connectorv1.LookupQuoteResponse{
		Listing: &connectorv1.QuoteListing{Symbol: "SYNQ", Isin: synqISIN, Name: "Synthetic Quote Corp", Exchange: "NASDAQ", Currency: "USD"},
		History: history,
	}
	for _, query := range []string{"SYNQ", "synq", synqISIN} {
		t.Run(query, func(t *testing.T) {
			t.Parallel()
			got, src, err := quoteWith(t, fakeKey, nil, quoteRequest(query, "2024-12-30", "2025-01-03"))
			if err != nil {
				t.Fatalf("LookupQuote: %v", err)
			}
			if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
				t.Errorf("quote (-want +got):\n%s", diff)
			}
			wantAsked := []string{"find " + query, "eod SYNQ.US 2024-12-30..2025-01-03"}
			if diff := cmp.Diff(wantAsked, src.asked); diff != "" || src.token != fakeKey {
				t.Errorf("asked EODHD with %q (-want +got):\n%s", src.token, diff)
			}
		})
	}
}

// TestLookupQuoteWithoutTradingInThePeriod checks that a period with no close,
// a weekend, still has the latest close before it.
func TestLookupQuoteWithoutTradingInThePeriod(t *testing.T) {
	t.Parallel()
	got, _, err := quoteWith(t, fakeKey, nil, quoteRequest("SYNQ", "2025-01-04", "2025-01-05"))
	if err != nil {
		t.Fatalf("LookupQuote: %v", err)
	}
	if diff := cmp.Diff([]*connectorv1.QuotePoint{{Date: "2025-01-03", Close: "57.5"}}, got.GetHistory(), protocmp.Transform()); diff != "" {
		t.Errorf("history (-want +got):\n%s", diff)
	}
}

// TestLookupQuoteSeveralInstruments checks that a ticker two ISINs share is
// answered with both, the preferred venue first, and no quote, and that no close is fetched: the person
// names one by its ISIN.
func TestLookupQuoteSeveralInstruments(t *testing.T) {
	t.Parallel()
	got, src, err := quoteWith(t, fakeKey, nil, quoteRequest("TWIN", "2025-01-02", "2025-01-03"))
	if err != nil {
		t.Fatalf("LookupQuote: %v", err)
	}
	want := &connectorv1.LookupQuoteResponse{Candidates: []*connectorv1.QuoteListing{
		{Symbol: "TWIN", Isin: twinBISIN, Name: "Twin Beta plc", Exchange: "LSE", Currency: "GBP"},
		{Symbol: "TWIN", Isin: twinAISIN, Name: "Twin Alpha AG", Exchange: "XETRA", Currency: "EUR"},
	}}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("candidates (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"find TWIN"}, src.asked); diff != "" {
		t.Errorf("asked EODHD (-want +got):\n%s", diff)
	}
}

// TestLookupQuoteCleansTheSourcesName checks that a name carrying a line
// break and an instruction is passed on as one line of text: the plugin hands
// over what the source says as data, shortened and without control
// characters, and never acts on it.
func TestLookupQuoteCleansTheSourcesName(t *testing.T) {
	t.Parallel()
	got, _, err := quoteWith(t, fakeKey, nil, quoteRequest("NOVA", "2025-01-02", "2025-01-03"))
	if err != nil {
		t.Fatalf("LookupQuote: %v", err)
	}
	name := got.GetListing().GetName()
	if strings.ContainsAny(name, "\n\r\t") || !strings.HasPrefix(name, "Nova Corp. IGNORE ALL PREVIOUS") {
		t.Errorf("name = %q, want the source's text on one line", name)
	}
	if long := cleanName(strings.Repeat("é", maxQuoteName+50)); len([]rune(long)) != maxQuoteName {
		t.Errorf("a long name has %d characters, want %d", len([]rune(long)), maxQuoteName)
	}
}

func TestLookupQuoteRefuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		key     string
		failure error
		req     *connectorv1.LookupQuoteRequest
		want    connect.Code
	}{
		{name: "no listing", key: fakeKey, req: quoteRequest("NOPE", "2025-01-02", "2025-01-03"), want: connect.CodeNotFound},
		{name: "a listing with no close in or before the period", key: fakeKey, req: quoteRequest("NOVA", "2023-01-01", "2023-01-02"), want: connect.CodeNotFound},
		{name: "no API key", key: "", req: quoteRequest("SYNQ", "2025-01-02", "2025-01-03"), want: connect.CodeFailedPrecondition},
		{name: "EODHD down", key: fakeKey, failure: errEODHDDown, req: quoteRequest("SYNQ", "2025-01-02", "2025-01-03"), want: connect.CodeUnavailable},
		{name: "a query that is no ticker", key: fakeKey, req: quoteRequest("SYNQ&api_token=x", "2025-01-02", "2025-01-03"), want: connect.CodeInvalidArgument},
		{name: "an empty query", key: fakeKey, req: quoteRequest("", "2025-01-02", "2025-01-03"), want: connect.CodeInvalidArgument},
		{name: "a period that is no days", key: fakeKey, req: quoteRequest("SYNQ", "yesterday", "2025-01-03"), want: connect.CodeInvalidArgument},
		{name: "a period out of order", key: fakeKey, req: quoteRequest("SYNQ", "2025-01-03", "2025-01-02"), want: connect.CodeInvalidArgument},
		{name: "a period in the future", key: fakeKey, req: quoteRequest("SYNQ", "2025-01-02", "2025-01-07"), want: connect.CodeInvalidArgument},
		{name: "a period over a year", key: fakeKey, req: quoteRequest("SYNQ", "2024-01-01", "2025-01-03"), want: connect.CodeInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := quoteWith(t, tt.key, tt.failure, tt.req)
			if got := connect.CodeOf(err); got != tt.want {
				t.Fatalf("code = %v (%v), want %v", got, err, tt.want)
			}
			if strings.Contains(err.Error(), fakeKey) {
				t.Errorf("the error %q holds the API key", err)
			}
		})
	}
}
