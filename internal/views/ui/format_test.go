package ui

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

func TestFormatAmount(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, in, want string }{
		{name: "small", in: "12.50", want: "12.50"},
		{name: "thousands", in: "1234.56", want: "1 234.56"},
		{name: "millions, negative", in: "-1234567.8", want: "-1 234 567.8"},
		{name: "plus sign dropped", in: "+1000", want: "1 000"},
		{name: "exactly three digits", in: "999", want: "999"},
		{name: "many decimals kept", in: "0.000001", want: "0.000001"},
		{name: "not a decimal, shown as is", in: "n/a", want: "n/a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := formatAmount(tt.in); got != tt.want {
				t.Errorf("formatAmount(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
	if got := formatMoney(nil); got != "" {
		t.Errorf("formatMoney(nil) = %q", got)
	}
}

func TestShare(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, part, total, want string }{
		{name: "two thirds, half-even to one place", part: "1000.00", total: "1500.00", want: "66.7%"},
		{name: "whole", part: "10", total: "10", want: "100%"},
		{name: "tie rounds to even", part: "1", total: "800", want: "0.1%"},
		{name: "negative part", part: "-50", total: "200", want: "-25%"},
		{name: "zero total shows nothing", part: "1", total: "0", want: ""},
		{name: "not a decimal", part: "x", total: "1", want: ""},
		{name: "total not a decimal", part: "1", total: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := share(tt.part, tt.total); got != tt.want {
				t.Errorf("share(%q, %q) = %q, want %q", tt.part, tt.total, got, tt.want)
			}
		})
	}
}

func TestParseDimension(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want viewsv1.AllocationDimension
	}{
		{in: "", want: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_ASSET_CLASS},
		{in: "asset-class", want: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_ASSET_CLASS},
		{in: "currency", want: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_CURRENCY},
		{in: "institution", want: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_INSTITUTION},
		{in: "sleeve", want: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_SLEEVE},
		{in: "broker", want: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_ASSET_CLASS},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			if got := parseDimension(tt.in).Enum; got != tt.want {
				t.Errorf("parseDimension(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestLabels(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, got, want string }{
		{name: "cash book", got: bookLabel(&viewsv1.Book{Kind: viewsv1.BookKind_BOOK_KIND_WITHHOLDING_TAX, Currency: "USD"}), want: "withholding tax USD"},
		{name: "holding book", got: bookLabel(&viewsv1.Book{Kind: viewsv1.BookKind_BOOK_KIND_HOLDING, Instrument: syntheticInstrument()}), want: "SYN (XX0000000001)"},
		{name: "cash balance", got: balanceLabel(&viewsv1.Book{Kind: viewsv1.BookKind_BOOK_KIND_CASH, Currency: "USD"}, "1234.50"), want: "1\u202f234.50 USD"},
		{name: "holding balance is a quantity", got: balanceLabel(&viewsv1.Book{Kind: viewsv1.BookKind_BOOK_KIND_HOLDING, Instrument: syntheticInstrument()}, "12"), want: "12"},
		{name: "enum", got: enumLabel(viewsv1.TransactionKind_TRANSACTION_KIND_OPENING_LOT.String(), "TRANSACTION_KIND_"), want: "opening lot"},
		{name: "scope", got: scopeLabel(identityv1.SessionScope_SESSION_SCOPE_OWNER), want: "owner"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestAsOfDate(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-04-01")
	statuses := []*viewsv1.ValuationStatus{
		{AccountId: "a", ValuedThrough: "2026-03-20"},
		{AccountId: "b", ValuedThrough: "2026-03-31"},
		{AccountId: "c"},
	}
	tests := []struct {
		name     string
		query    string
		statuses []*viewsv1.ValuationStatus
		want     string
		wantErr  bool
	}{
		{name: "query wins", query: "2026-01-15", statuses: statuses, want: "2026-01-15"},
		{name: "latest valued-through date", statuses: statuses, want: "2026-03-31"},
		{name: "nothing valued: today", statuses: statuses[2:], want: "2026-04-01"},
		{name: "malformed query", query: "15.01.2026", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := asOfDate(tt.query, tt.statuses, today)
			if (err != nil) != tt.wantErr {
				t.Fatalf("asOfDate error = %v, want error %t", err, tt.wantErr)
			}
			if err == nil && got.String() != tt.want {
				t.Errorf("asOfDate = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestBuildCharts checks the chart data: exact decimal strings, groups in
// the read API's order, and the value-over-time range from its first date.
func TestBuildCharts(t *testing.T) {
	t.Parallel()
	alloc := buildAllocation(&viewsv1.GetAllocationResponse{
		Date: "2026-03-31", TotalReporting: pln("300.10"),
		Groups: []*viewsv1.AllocationGroup{{Key: "PLN", ValueReporting: pln("200.05")}, {Key: "EUR", ValueReporting: pln("100.05")}},
	}, parseDimension("currency"))
	wantAlloc := `{"kind":"bar","label":"Value by Currency","currency":"PLN","labels":["PLN","EUR"],"values":["200.05","100.05"]}`
	if alloc.Chart != wantAlloc {
		t.Errorf("allocation chart = %s, want %s", alloc.Chart, wantAlloc)
	}
	if got := []string{alloc.Groups[0].Share, alloc.Groups[1].Share}; got[0] != "66.7%" || got[1] != "33.3%" {
		t.Errorf("shares = %v", got)
	}
	perf := &viewsv1.GetPerformanceResponse{ReportingCurrency: "PLN", Gap: viewsv1.PerformanceGap_PERFORMANCE_GAP_UNVALUED, Days: []*viewsv1.PerformanceDay{
		{Date: "2026-01-02", Index: "100", ValueReporting: pln("10.00")}, {Date: "2026-01-03", Index: "110", ValueReporting: pln("11.00")},
	}}
	hist := buildPerformance(perf, nil)
	wantHist := performanceView{
		From: "2026-01-02", To: "2026-01-03", Currency: "PLN", Points: 2, Value: pln("11.00"), Gap: "the next day is not fully valued yet",
		Chart: `{"kind":"line","label":"Portfolio (time-weighted return)","currency":"index","labels":["2026-01-02","2026-01-03"],"values":["100","110"]}`,
	}
	if diff := cmp.Diff(wantHist, hist, protocmp.Transform()); diff != "" {
		t.Errorf("performance (-want +got):\n%s", diff)
	}
	// The model's line is aligned with the portfolio's days; a day it lacks
	// is empty.
	withModel := buildPerformance(perf, &viewsv1.GetModelPortfolioResponse{Days: []*viewsv1.BacktestDay{{Date: "2026-01-03", Index: "100"}}})
	if want := `"series":[{"label":"Model portfolio (backtest)","values":["","100"]}]`; !strings.Contains(withModel.Chart, want) || withModel.Model == nil {
		t.Errorf("chart with the model = %s, want %s", withModel.Chart, want)
	}
	if empty := buildPerformance(&viewsv1.GetPerformanceResponse{}, nil); empty.Points != 0 || empty.From != "" || empty.Model != nil {
		t.Errorf("empty performance = %+v", empty)
	}
}

func TestJoinValuations(t *testing.T) {
	t.Parallel()
	other := &viewsv1.Instrument{Isin: "XX0000000002", Symbol: "OTH"}
	val := &viewsv1.Valuation{Instrument: syntheticInstrument(), Value: pln("5.00")}
	got := joinValuations(
		[]*viewsv1.Holding{{Instrument: syntheticInstrument(), Quantity: "1"}, {Instrument: other, Quantity: "2"}},
		[]*viewsv1.Valuation{val},
	)
	want := []holdingRow{
		{Holding: &viewsv1.Holding{Instrument: syntheticInstrument(), Quantity: "1"}, Valuation: val},
		{Holding: &viewsv1.Holding{Instrument: other, Quantity: "2"}},
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("joinValuations (-want +got):\n%s", diff)
	}
}

func TestRequestOrigin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, host, proto, wantOrigin string
		wantLocal                     bool
	}{
		{name: "https behind a proxy", host: testHost, proto: "https", wantOrigin: testOrigin},
		{name: "port-forward to localhost", host: "localhost:8080", wantOrigin: localOrigin, wantLocal: true},
		{name: "loopback address", host: "127.0.0.1:8080", wantOrigin: "http://127.0.0.1:8080", wantLocal: true},
		{name: "plain http elsewhere", host: testHost, wantOrigin: "http://" + testHost},
		{name: "unknown proto ignored", host: testHost, proto: "ftp", wantOrigin: "http://" + testHost},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
			if err != nil {
				t.Fatal(err)
			}
			r.Host = tt.host
			if tt.proto != "" {
				r.Header.Set("X-Forwarded-Proto", tt.proto)
			}
			origin := requestOrigin(r)
			if origin != tt.wantOrigin || localHTTP(origin) != tt.wantLocal {
				t.Errorf("origin %q, local %t; want %q, %t", origin, localHTTP(origin), tt.wantOrigin, tt.wantLocal)
			}
		})
	}
}

// TestSessionCookie checks the session cookie's attributes: the __Host-
// name, HttpOnly, Secure, SameSite=Strict and Path=/.
func TestSessionCookie(t *testing.T) {
	t.Parallel()
	c := sessionCookie("t", time.Unix(0, 0), -1)
	got := []any{c.Name, c.Path, c.Secure, c.HttpOnly, c.SameSite, c.MaxAge}
	want := []any{"__Host-openfief_session", "/", true, true, http.SameSiteStrictMode, -1}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func TestTokenDays(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in     string
		want   int
		wantOK bool
	}{
		{in: "", want: 30, wantOK: true},
		{in: "1", want: 1, wantOK: true},
		{in: "365", want: 365, wantOK: true},
		{in: "0"},
		{in: "366"},
		{in: "7.5"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, ok := tokenDays(tt.in)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("tokenDays(%q) = %d, %t; want %d, %t", tt.in, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

// TestUTCToday checks the default clock gives a date.
func TestUTCToday(t *testing.T) {
	t.Parallel()
	if utcToday().IsZero() {
		t.Error("utcToday() is the zero date")
	}
}
