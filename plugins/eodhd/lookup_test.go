package main

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// lookupWith asks a connector over a fake EODHD with key, failing with
// failure, to look tickers up.
func lookupWith(t *testing.T, key string, failure error, tickers []*connectorv1.Ticker) (*connectorv1.LookupInstrumentsResponse, *fakeEODHD, error) {
	t.Helper()
	var src *fakeEODHD
	c := &connector{
		secrets: secretsDir(t, key), today: func() kernel.Date { return day(t, "2025-01-06") },
		source: func(token string) marketData { src = newFakeEODHD(token); src.err = failure; return src },
	}
	resp, err := c.LookupInstruments(t.Context(), connect.NewRequest(&connectorv1.LookupInstrumentsRequest{Tickers: tickers}))
	if err != nil {
		return nil, src, err
	}
	return resp.Msg, src, nil
}

// TestLookupInstruments checks D-116's lookup: a ticker's code found once
// on its country's exchange is proposed with its name; one found under two
// ISINs, one of a country with no supported exchange and one EODHD doesn't
// list get no proposal, and only supported exchanges are searched.
func TestLookupInstruments(t *testing.T) {
	t.Parallel()
	tickers := []*connectorv1.Ticker{
		{Symbol: "FAKEW.PL", Code: "FAKEW", Country: "PL"},
		{Symbol: "SYNT.DE", Code: "SYNT", Country: "DE"},
		{Symbol: "SYN.UK", Code: "SYN", Country: "GB"},
		{Symbol: "NON.US", Code: "NON", Country: "US"},
		{Symbol: "GONE.PL", Code: "GONE", Country: "PL"},
	}
	got, src, err := lookupWith(t, fakeKey, nil, tickers)
	if err != nil {
		t.Fatalf("LookupInstruments: %v", err)
	}
	want := &connectorv1.LookupInstrumentsResponse{Proposals: []*connectorv1.InstrumentProposal{
		{Symbol: "FAKEW.PL", Isin: fakewISIN, Name: "Fake Warsaw SA", Currency: "PLN"},
		{Symbol: "SYNT.DE", Isin: syntISIN, Name: "Synthetic Test ETF", Currency: "EUR"},
	}}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("proposals (-want +got):\n%s", diff)
	}
	wantAsked := []string{"search FAKEW on WAR", "search SYNT on XETRA", "search SYN on LSE", "search GONE on WAR"}
	if diff := cmp.Diff(wantAsked, src.asked); diff != "" || src.token != fakeKey {
		t.Errorf("asked EODHD with %q (-want +got):\n%s", src.token, diff)
	}
}

func TestLookupInstrumentsRefuses(t *testing.T) {
	t.Parallel()
	many := make([]*connectorv1.Ticker, maxLookupTickers+1)
	for i := range many {
		many[i] = &connectorv1.Ticker{Symbol: "FAKEW.PL", Code: "FAKEW", Country: "PL"}
	}
	one := many[:1]
	tests := []struct {
		name    string
		key     string
		failure error
		tickers []*connectorv1.Ticker
		want    connect.Code
	}{
		{name: "no API key", key: "", tickers: one, want: connect.CodeFailedPrecondition},
		{name: "EODHD down", key: fakeKey, failure: errEODHDDown, tickers: one, want: connect.CodeUnavailable},
		{name: "too many", key: fakeKey, tickers: many, want: connect.CodeInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, _, err := lookupWith(t, tt.key, tt.failure, tt.tickers); connect.CodeOf(err) != tt.want {
				t.Errorf("LookupInstruments = %v, want %v", err, tt.want)
			}
		})
	}
}
