package connectapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// Synthetic ISINs of a quote: one instrument SYNQ, and TWIN under two.
const (
	isinSYNQ  = "XX00000SYNQ9"
	isinTwinA = "XX00000TWIN8"
	isinTwinB = "XX0000TWINB5"
)

// quoteSource answers by query: SYNQ is quoted from "eodhd", TWIN names two
// listings, and any other query fails with err or finds nothing.
type quoteSource struct {
	t         *testing.T
	err       error
	available bool
	asked     []ingest.QuoteRequest
}

func (s *quoteSource) Quote(_ context.Context, req ingest.QuoteRequest) (ingest.QuoteAnswer, error) {
	s.asked = append(s.asked, req)
	if s.err != nil {
		return ingest.QuoteAnswer{}, s.err
	}
	usd := cur(s.t, "USD")
	listing := func(isin, symbol, name string) ingest.QuoteListing {
		i, err := ledger.ParseISIN(isin)
		if err != nil {
			s.t.Fatal(err)
		}
		return ingest.QuoteListing{Symbol: symbol, ISIN: i, Name: name, Exchange: "NASDAQ", Currency: usd}
	}
	switch req.Query {
	case "SYNQ":
		q := ingest.Quote{
			Listing: listing(isinSYNQ, "SYNQ", "Synthetic Quote Corp"),
			History: []ingest.QuotePoint{
				{Date: date(s.t, "2026-03-01"), Close: money(s.t, "56.25", "USD")},
				{Date: date(s.t, "2026-03-02"), Close: money(s.t, "57.5", "USD")},
			},
		}
		return ingest.QuoteAnswer{Source: "eodhd", Quote: &q}, nil
	case "TWIN":
		return ingest.QuoteAnswer{Source: "eodhd", Candidates: []ingest.QuoteListing{listing(isinTwinA, "TWIN", "Twin Alpha AG"), listing(isinTwinB, "TWIN", "Twin Beta plc")}}, nil
	}
	return ingest.QuoteAnswer{}, fmt.Errorf("x: %w", ingestapp.ErrQuoteNotFound)
}

func (s *quoteSource) Available(context.Context) (bool, error) { return s.available, nil }

// quoteAPI is the read API over a quote source, on 2026-03-02, or with none
// when src is nil.
func quoteAPI(t *testing.T, src *quoteSource) api {
	t.Helper()
	var opts []app.Option
	if src != nil {
		opts = append(opts, app.WithMarketQuotes(&app.MarketQuotesConfig{Source: src, Today: func() kernel.Date { return date(t, "2026-03-02") }}))
	}
	var log bytes.Buffer
	h, err := connectapi.New(newViews(t, newLedger(t), opts...), fakeVerifier{},
		connectapi.WithLogger(slog.New(slog.NewJSONHandler(&log, nil))),
		connectapi.WithRequestIDs(func() string { return "req-1" }))
	if err != nil {
		t.Fatalf("connectapi.New: %v", err)
	}
	return api{handler: h, log: &log}
}

// serveAPI serves a over a test server, closed with the test.
func serveAPI(t *testing.T, a api) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(a.handler)
	t.Cleanup(srv.Close)
	return srv
}

// TestConnectGetMarketQuote checks the capability's shape over Connect: the
// listing, the latest close and its day, the closes in order, the connector
// instance that answered and the tier T2, for the owner and a script token alike.
func TestConnectGetMarketQuote(t *testing.T) {
	t.Parallel()
	for _, token := range []string{goodToken, scriptToken} {
		t.Run(token, func(t *testing.T) {
			t.Parallel()
			src := &quoteSource{t: t, available: true}
			srv := serveAPI(t, quoteAPI(t, src))
			c := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)
			got, err := c.GetMarketQuote(context.Background(), authed(token, &viewsv1.GetMarketQuoteRequest{Query: "SYNQ"}))
			if err != nil {
				t.Fatalf("GetMarketQuote: %v", err)
			}
			want := &viewsv1.GetMarketQuoteResponse{
				Source: "eodhd", DataTier: "T2",
				Quote: &viewsv1.MarketQuote{
					Listing: &viewsv1.MarketQuoteListing{
						Instrument: &viewsv1.Instrument{Isin: isinSYNQ, Symbol: "SYNQ", Name: "Synthetic Quote Corp"}, Exchange: "NASDAQ", Currency: "USD",
					},
					Latest: &viewsv1.Money{Amount: "57.5", Currency: "USD"}, Date: "2026-03-02", Source: "eodhd", DataTier: "T2",
					History: []*viewsv1.MarketQuotePoint{
						{Date: "2026-03-01", Close: &viewsv1.Money{Amount: "56.25", Currency: "USD"}},
						{Date: "2026-03-02", Close: &viewsv1.Money{Amount: "57.5", Currency: "USD"}},
					},
				},
			}
			if diff := cmp.Diff(want, got.Msg, protocmp.Transform()); diff != "" {
				t.Errorf("quote (-want +got):\n%s", diff)
			}
			if len(src.asked) != 1 || src.asked[0].From.String() != "2026-02-01" || src.asked[0].To.String() != "2026-03-02" {
				t.Errorf("the source was asked %+v, want the last 30 days up to 2026-03-02", src.asked)
			}
		})
	}
}

func TestConnectGetMarketQuoteNamesSeveralListings(t *testing.T) {
	t.Parallel()
	srv := serveAPI(t, quoteAPI(t, &quoteSource{t: t, available: true}))
	c := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)
	got, err := c.GetMarketQuote(context.Background(), authed(goodToken, &viewsv1.GetMarketQuoteRequest{Query: "TWIN"}))
	if err != nil {
		t.Fatalf("GetMarketQuote: %v", err)
	}
	isins := make([]string, 0, len(got.Msg.GetCandidates()))
	for _, l := range got.Msg.GetCandidates() {
		isins = append(isins, l.GetInstrument().GetIsin())
	}
	if diff := cmp.Diff([]string{isinTwinA, isinTwinB}, isins); diff != "" || got.Msg.GetQuote() != nil || got.Msg.GetDataTier() != "T2" {
		t.Errorf("candidates (-want +got):\n%s\nquote = %v, tier %q", diff, got.Msg.GetQuote(), got.Msg.GetDataTier())
	}
}

// TestRESTGetMarketQuote checks the capability under /v1 and the error each
// failure has, by status, with a text that quotes nothing asked.
func TestRESTGetMarketQuote(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		src    *quoteSource
		target string
		token  string
		want   int
	}{
		{name: "a quote", src: &quoteSource{t: t}, target: "/v1/market-quotes/SYNQ?from=2026-02-20&to=2026-03-02", token: goodToken, want: http.StatusOK},
		{name: "a ticker with a dot", src: &quoteSource{t: t}, target: "/v1/market-quotes/BRK.B", token: goodToken, want: http.StatusNotFound},
		{name: "no token", src: &quoteSource{t: t}, target: "/v1/market-quotes/SYNQ", want: http.StatusUnauthorized},
		{name: "a date that is none", src: &quoteSource{t: t}, target: "/v1/market-quotes/SYNQ?from=yesterday", token: goodToken, want: http.StatusBadRequest},
		{name: "a period in the future", src: &quoteSource{t: t}, target: "/v1/market-quotes/SYNQ?to=2026-03-03", token: goodToken, want: http.StatusBadRequest},
		{name: "a period over a year", src: &quoteSource{t: t}, target: "/v1/market-quotes/SYNQ?from=2025-01-01", token: goodToken, want: http.StatusBadRequest},
		{name: "a query that is no ticker", src: &quoteSource{t: t}, target: "/v1/market-quotes/SYNQ%3Fx", token: goodToken, want: http.StatusBadRequest},
		{name: "a query of dots alone", src: &quoteSource{t: t}, target: "/v1/market-quotes/..%2E", token: goodToken, want: http.StatusBadRequest},
		{name: "a query with two dots inside", src: &quoteSource{t: t}, target: "/v1/market-quotes/SYNQ..X", token: goodToken, want: http.StatusBadRequest},
		{name: "no such instrument", src: &quoteSource{t: t}, target: "/v1/market-quotes/NOPE", token: goodToken, want: http.StatusNotFound},
		{name: "the source is down", src: &quoteSource{t: t, err: fmt.Errorf("x: %w", ingestapp.ErrConnectorUnavailable)}, target: "/v1/market-quotes/SYNQ", token: goodToken, want: http.StatusServiceUnavailable},
		{name: "no market-data connector", src: &quoteSource{t: t, err: fmt.Errorf("x: %w", ingestapp.ErrNoQuoteSource)}, target: "/v1/market-quotes/SYNQ", token: goodToken, want: http.StatusBadRequest},
		{name: "no source at all, as on a demo", src: nil, target: "/v1/market-quotes/SYNQ", token: goodToken, want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, rec := quoteAPI(t, tt.src).serve(t, http.MethodGet, tt.target, tt.token)
			if rec.Code != tt.want {
				t.Fatalf("status = %d (%s), want %d", rec.Code, rec.Body.String(), tt.want)
			}
			if tt.want >= http.StatusBadRequest && strings.Contains(strings.ToLower(rec.Body.String()), "x?") {
				t.Errorf("the error body quotes the query: %s", rec.Body.String())
			}
			if tt.want == http.StatusOK {
				checkQuoteBody(t, rec.Body.Bytes())
			}
		})
	}
}

// checkQuoteBody checks a REST answer for SYNQ: the latest day, the source and tier T2.
func checkQuoteBody(t *testing.T, raw []byte) {
	t.Helper()
	var body struct {
		Quote struct {
			Date     string `json:"date"`
			DataTier string `json:"dataTier"`
			Source   string `json:"source"`
		} `json:"quote"`
		DataTier string `json:"dataTier"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if body.Quote.Date != "2026-03-02" || body.Quote.Source != "eodhd" || body.Quote.DataTier != "T2" || body.DataTier != "T2" {
		t.Errorf("response = %s, want the latest day, the source and tier T2", raw)
	}
}

// TestMarketQuoteWithoutASourceIsAFailedPrecondition checks the code of a
// quote asked where none can be looked up, over Connect.
func TestMarketQuoteWithoutASourceIsAFailedPrecondition(t *testing.T) {
	t.Parallel()
	srv := serveAPI(t, quoteAPI(t, nil))
	c := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)
	_, err := c.GetMarketQuote(context.Background(), authed(goodToken, &viewsv1.GetMarketQuoteRequest{Query: "SYNQ"}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("code = %v (%v), want FailedPrecondition", connect.CodeOf(err), err)
	}
}

// TestSettingsTellWhetherQuotesAreAvailable checks that GetSettings says
// whether a quote can be asked, which a client that offers the lookup, the Ask
// box, reads.
func TestSettingsTellWhetherQuotesAreAvailable(t *testing.T) {
	t.Parallel()
	for name, src := range map[string]*quoteSource{"a connector runs": {t: t, available: true}, "none runs": {t: t}, "a demo": nil} {
		srv := serveAPI(t, quoteAPI(t, src))
		c := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)
		got, err := c.GetSettings(context.Background(), authed(goodToken, &viewsv1.GetSettingsRequest{}))
		if err != nil {
			t.Fatalf("%s: GetSettings: %v", name, err)
		}
		if want := src != nil && src.available; got.Msg.GetMarketQuotes() != want {
			t.Errorf("%s: market_quotes = %v, want %v", name, got.Msg.GetMarketQuotes(), want)
		}
	}
}

// TestConnectGetMarketQuoteRefusesAViewPluginAndADemoSession checks that the
// callers who may not read quotes are refused with PermissionDenied and the
// source is never asked (D-226).
func TestConnectGetMarketQuoteRefusesAViewPluginAndADemoSession(t *testing.T) {
	t.Parallel()
	for _, token := range []string{pluginToken, demoToken} {
		t.Run(token, func(t *testing.T) {
			t.Parallel()
			src := &quoteSource{t: t, available: true}
			srv := serveAPI(t, quoteAPI(t, src))
			c := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)
			_, err := c.GetMarketQuote(context.Background(), authed(token, &viewsv1.GetMarketQuoteRequest{Query: "SYNQ"}))
			if connect.CodeOf(err) != connect.CodePermissionDenied {
				t.Errorf("code = %v (%v), want PermissionDenied", connect.CodeOf(err), err)
			}
			if len(src.asked) != 0 {
				t.Errorf("the source was asked %d times", len(src.asked))
			}
		})
	}
}
