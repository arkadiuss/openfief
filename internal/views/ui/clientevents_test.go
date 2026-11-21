package ui

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// num returns a pointer to n, as a field the page sent.
func num(n int64) *int64 {
	p := new(int64)
	*p = n
	return p
}

func TestValidClientEvent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   rawClientEvent
		want clientEvent
		ok   bool
	}{
		{
			"success with time and path",
			rawClientEvent{Kind: "turnstile-success", Ms: num(1234), Path: "visible-fallback"},
			clientEvent{Kind: "turnstile-success", Ms: 1234, Path: "visible-fallback", Line: -1, Status: -1},
			true,
		},
		{
			"error with a code",
			rawClientEvent{Kind: "turnstile-error", Code: "110200", Ms: num(5)},
			clientEvent{Kind: "turnstile-error", Ms: 5, Code: "110200", Line: -1, Status: -1},
			true,
		},
		{
			"the check starting on demand",
			rawClientEvent{Kind: "turnstile-execute", Ms: num(0)},
			clientEvent{Kind: "turnstile-execute", Ms: 0, Line: -1, Status: -1},
			true,
		},
		{
			"a timeout reported as an error with a fixed word",
			rawClientEvent{Kind: "turnstile-error", Code: "timeout", Ms: num(15000), Path: "visible-fallback"},
			clientEvent{Kind: "turnstile-error", Ms: 15000, Code: "timeout", Path: "visible-fallback", Line: -1, Status: -1},
			true,
		},
		{"unknown kind", rawClientEvent{Kind: "keystroke"}, clientEvent{}, false},
		{"empty kind", rawClientEvent{}, clientEvent{}, false},
		{"negative time", rawClientEvent{Kind: "turnstile-success", Ms: num(-2)}, clientEvent{}, false},
		{"time past ten minutes", rawClientEvent{Kind: "turnstile-success", Ms: num(maxClientMs + 1)}, clientEvent{}, false},
		{
			"code with upper case is dropped",
			rawClientEvent{Kind: "turnstile-error", Code: "Bad Code"},
			clientEvent{Kind: "turnstile-error", Ms: -1, Line: -1, Status: -1},
			true,
		},
		{
			"code too long is dropped",
			rawClientEvent{Kind: "turnstile-error", Code: strings.Repeat("a", 49)},
			clientEvent{Kind: "turnstile-error", Ms: -1, Line: -1, Status: -1},
			true,
		},
		{
			"path other than the two words is dropped",
			rawClientEvent{Kind: "turnstile-success", Path: "my-account-9"},
			clientEvent{Kind: "turnstile-success", Ms: -1, Line: -1, Status: -1},
			true,
		},
		{
			"status 503",
			rawClientEvent{Kind: "request-failed", Status: num(503)},
			clientEvent{Kind: "request-failed", Ms: -1, Line: -1, Status: 503},
			true,
		},
		{
			"status 0 is a network failure",
			rawClientEvent{Kind: "request-failed", Status: num(0)},
			clientEvent{Kind: "request-failed", Ms: -1, Line: -1, Status: 0},
			true,
		},
		{"status out of range", rawClientEvent{Kind: "request-failed", Status: num(99999)}, clientEvent{}, false},
		{
			"own script error keeps a plain message",
			rawClientEvent{Kind: "script-error", Src: "/static/ask.js", Line: num(42), Msg: "x is not a function"},
			clientEvent{Kind: "script-error", Ms: -1, Source: "/static/ask.js", Message: "x is not a function", Line: 42, Status: -1},
			true,
		},
		{
			"own script error with a quoted parse error loses its message",
			rawClientEvent{Kind: "script-error", Src: "/static/ask.js", Msg: `Unexpected token 'a', "abc" is not valid JSON`},
			clientEvent{Kind: "script-error", Ms: -1, Source: "/static/ask.js", Line: -1, Status: -1},
			true,
		},
		{
			"own script error with digits loses its message",
			rawClientEvent{Kind: "script-error", Src: "/static/ask.js", Msg: "bad value 12345"},
			clientEvent{Kind: "script-error", Ms: -1, Source: "/static/ask.js", Line: -1, Status: -1},
			true,
		},
		{
			"message is cut to 200 characters",
			rawClientEvent{Kind: "script-error", Src: "/static/ask.js", Msg: strings.Repeat("a", 500)},
			clientEvent{Kind: "script-error", Ms: -1, Source: "/static/ask.js", Message: strings.Repeat("a", maxClientMessage), Line: -1, Status: -1},
			true,
		},
		{
			"script with a query is not a path of ours",
			rawClientEvent{Kind: "script-error", Src: "/static/ask.js?q=secret", Msg: "oops"},
			clientEvent{Kind: "script-error", Ms: -1, Source: "other", Line: -1, Status: -1},
			true,
		},
		{
			"cloudflare script keeps the host and no message",
			rawClientEvent{Kind: "script-error", Src: "https://challenges.cloudflare.com/turnstile/v0/api.js", Msg: "Script error."},
			clientEvent{Kind: "script-error", Ms: -1, Source: "challenges.cloudflare.com", Line: -1, Status: -1},
			true,
		},
		{
			"another host keeps nothing",
			rawClientEvent{Kind: "script-error", Src: "https://evil.example/x.js", Msg: "hello"},
			clientEvent{Kind: "script-error", Ms: -1, Source: "other", Line: -1, Status: -1},
			true,
		},
		{
			"a message on any other kind is never kept",
			rawClientEvent{Kind: "turnstile-error", Msg: "typed text", Src: "/static/ask.js"},
			clientEvent{Kind: "turnstile-error", Ms: -1, Line: -1, Status: -1},
			true,
		},
		{
			"line out of range is dropped",
			rawClientEvent{Kind: "script-error", Line: num(-5)},
			clientEvent{Kind: "script-error", Ms: -1, Source: "other", Line: -1, Status: -1},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := validClientEvent(&tt.in)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("event mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBrowserOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, ua string
		brave    bool
		want     browser
	}{
		{
			"chrome on android", "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.6478.122 Mobile Safari/537.36", false,
			browser{Family: "chrome", Major: 126, Mobile: true},
		},
		{
			"brave on android says so itself", "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36", true,
			browser{Family: "brave", Major: 126, Mobile: true},
		},
		{
			"edge", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36 Edg/126.0.2592.81", false,
			browser{Family: "edge", Major: 126},
		},
		{"firefox", "Mozilla/5.0 (X11; Linux x86_64; rv:127.0) Gecko/20100101 Firefox/127.0", false, browser{Family: "firefox", Major: 127}},
		{
			"safari", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15", false,
			browser{Family: "safari", Major: 17},
		},
		{"version token without Safari", "Foo Version/9", false, browser{Family: "other"}},
		{"curl", "curl/8.4.0", false, browser{Family: "other"}},
		{"empty", "", false, browser{Family: "other"}},
		{"brave claim on a non-chrome browser is ignored", "Mozilla/5.0 (X11; Linux x86_64; rv:127.0) Gecko/20100101 Firefox/127.0", true, browser{Family: "firefox", Major: 127}},
		{"a long version keeps three digits", "Chrome/123456789", false, browser{Family: "chrome", Major: 123}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := browserOf(tt.ua, tt.brave); got != tt.want {
				t.Errorf("browserOf = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAddressLimit(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	l := &addressLimit{now: func() time.Time { return now }, windows: map[string]clientWindow{}}
	for i := range clientReportsPerMinute {
		if !l.allow("198.51.100.1") {
			t.Fatalf("report %d refused, within the limit", i+1)
		}
	}
	if l.allow("198.51.100.1") {
		t.Error("a report past the limit was allowed")
	}
	if !l.allow("198.51.100.2") {
		t.Error("another address was refused for the first one's reports")
	}
	now = now.Add(time.Minute)
	if !l.allow("198.51.100.1") {
		t.Error("the address was still refused a minute later")
	}
}

func TestAddressLimitForgetsOldAddressesAndRefusesNewOnesWhenFull(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	l := &addressLimit{now: func() time.Time { return now }, windows: map[string]clientWindow{}}
	for i := range clientRateAddresses {
		l.windows[strings.Repeat("a", 1)+string(rune('A'+i%26))+strings.Repeat("b", i/26)] = clientWindow{start: now, count: 1}
	}
	if l.allow("203.0.113.9") {
		t.Error("a new address was allowed while every remembered window is current and the table is full")
	}
	now = now.Add(2 * time.Minute)
	if !l.allow("203.0.113.9") {
		t.Error("a new address was refused after the old windows ended")
	}
	if len(l.windows) > clientRateAddresses {
		t.Errorf("the table holds %d addresses, over its bound", len(l.windows))
	}
}

func TestAddressLimitCountsAnIPv6AddressByItsNetwork(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	l := &addressLimit{now: func() time.Time { return now }, windows: map[string]clientWindow{}}
	// A host with a /64 can take any of its 2^64 addresses: a report from a new
	// one must not start a new allowance, or take a new place in the table.
	for i := range clientReportsPerMinute {
		if !l.allow(fmt.Sprintf("2001:db8:1:2::%x", i+1)) {
			t.Fatalf("report %d refused, within the network's limit", i+1)
		}
	}
	if l.allow("2001:db8:1:2:ffff:ffff:ffff:ffff") {
		t.Error("a report from another address of the same /64 was allowed past the network's limit")
	}
	if !l.allow("2001:db8:1:3::1") {
		t.Error("an address of another /64 was refused for the first network's reports")
	}
	if got := len(l.windows); got != 2 {
		t.Errorf("the table holds %d entries for two networks, want 2", got)
	}
}

func TestAddressLimitCountsAnIPv4AddressAsItself(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	l := &addressLimit{now: func() time.Time { return now }, windows: map[string]clientWindow{}}
	for range clientReportsPerMinute {
		l.allow("198.51.100.1")
	}
	if l.allow("::ffff:198.51.100.1") {
		t.Error("the same IPv4 address written as IPv4-mapped IPv6 started a new allowance")
	}
	if !l.allow("198.51.100.2") {
		t.Error("the next IPv4 address was refused for the first one's reports")
	}
}

func TestAddressLimitCapsReportsFromEveryAddressTogether(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	l := &addressLimit{now: func() time.Time { return now }, windows: map[string]clientWindow{}}
	allowed := 0
	for i := range clientReportsPerMinuteTotal + clientReportsPerMinute {
		if l.allow(fmt.Sprintf("198.51.%d.%d", i/250, i%250+1)) {
			allowed++
		}
	}
	if allowed != clientReportsPerMinuteTotal {
		t.Errorf("allowed %d reports from many addresses in a minute, want the cap of %d", allowed, clientReportsPerMinuteTotal)
	}
	now = now.Add(time.Minute)
	if !l.allow("203.0.113.9") {
		t.Error("a report was refused a minute after the cap was reached")
	}
}

func TestAddressLimitRefusalsTakeNothingOfTheSharedCap(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	l := &addressLimit{now: func() time.Time { return now }, windows: map[string]clientWindow{}}
	for range clientReportsPerMinuteTotal + clientReportsPerMinute {
		l.allow("198.51.100.1")
	}
	if !l.allow("198.51.100.2") {
		t.Error("one address past its own limit used up every other address's reports")
	}
}

// clientEventsHarness returns a UI whose log the test reads and whose metrics a
// reader collects.
func clientEventsHarness(t *testing.T, mod func(*Config)) (*harness, *strings.Builder, *sdkmetric.ManualReader) {
	t.Helper()
	logs := &strings.Builder{}
	reader := sdkmetric.NewManualReader()
	meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)).Meter("test")
	h := newHarnessWith(t, slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})), func(c *Config) {
		c.Meter = meter
		if mod != nil {
			mod(c)
		}
	})
	return h, logs, reader
}

func report(body string) *http.Request {
	r := request(http.MethodPost, clientEventsPath, body)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.6478.122 Mobile Safari/537.36")
	r.Header.Set(openfief.ClientIPHeader, "198.51.100.77")
	return r
}

// metricTotals are the counts the client events' metrics hold.
type metricTotals struct {
	events int64
	tokens uint64
}

// totals returns how many events the counter has counted and how many times to a
// token the histogram has recorded.
func totals(t *testing.T, reader *sdkmetric.ManualReader) metricTotals {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	var got metricTotals
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch d := m.Data.(type) {
			case metricdata.Sum[int64]:
				for _, p := range d.DataPoints {
					got.events += p.Value
				}
			case metricdata.Histogram[float64]:
				for _, p := range d.DataPoints {
					got.tokens += p.Count
				}
			}
		}
	}
	return got
}

func TestClientEventsAreLoggedWithoutTheAddressAndCounted(t *testing.T) {
	t.Parallel()
	h, logs, reader := clientEventsHarness(t, nil)
	rec := h.serve(t, report(`{"events":[
		{"k":"turnstile-success","ms":2500,"path":"visible-fallback"},
		{"k":"turnstile-error","code":"110200","ms":900},
		{"k":"keystroke","msg":"hunter2"},
		{"k":"request-failed","status":503}]}`))
	wantStatus(t, rec, http.StatusNoContent)
	out := logs.String()
	for _, want := range []string{"kind=turnstile-success", "path=visible-fallback", "ms=2500", "kind=turnstile-error", "code=110200", "kind=request-failed", "status=503", "browser=chrome", "major=126", "mobile=true"} {
		if !strings.Contains(out, want) {
			t.Errorf("the log lacks %q:\n%s", want, out)
		}
	}
	for _, never := range []string{"198.51.100.77", "keystroke", "hunter2"} {
		if strings.Contains(out, never) {
			t.Errorf("the log holds %q:\n%s", never, out)
		}
	}
	if got := totals(t, reader); got.events != 3 || got.tokens != 1 {
		t.Errorf("counted %d events and %d times to a token, want 3 and 1 (the unknown kind is not counted)", got.events, got.tokens)
	}
}

func TestClientEventsRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		body string
		want int
	}{
		{"not json", `hello`, http.StatusBadRequest},
		{"too large", `{"events":[{"k":"turnstile-success","code":"` + strings.Repeat("a", maxClientEventsBody) + `"}]}`, http.StatusRequestEntityTooLarge},
		{"empty report", `{"events":[]}`, http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, logs, _ := clientEventsHarness(t, nil)
			wantStatus(t, h.serve(t, report(tt.body)), tt.want)
			if strings.Contains(logs.String(), "client event") {
				t.Errorf("a refused report was logged:\n%s", logs)
			}
		})
	}
}

func TestClientEventsRateLimitPerAddress(t *testing.T) {
	t.Parallel()
	h, _, _ := clientEventsHarness(t, nil)
	for range clientReportsPerMinute {
		wantStatus(t, h.serve(t, report(`{"events":[]}`)), http.StatusNoContent)
	}
	wantStatus(t, h.serve(t, report(`{"events":[]}`)), http.StatusTooManyRequests)
	other := report(`{"events":[]}`)
	other.Header.Set(openfief.ClientIPHeader, "198.51.100.78")
	wantStatus(t, h.serve(t, other), http.StatusNoContent)
}

func TestClientEventsCapTheNumberOfEventsInOneReport(t *testing.T) {
	t.Parallel()
	h, logs, _ := clientEventsHarness(t, nil)
	events := strings.Repeat(`{"k":"turnstile-expired"},`, maxClientEvents+5)
	wantStatus(t, h.serve(t, report(`{"events":[`+strings.TrimSuffix(events, ",")+`]}`)), http.StatusNoContent)
	if got := strings.Count(logs.String(), "kind=turnstile-expired"); got != maxClientEvents {
		t.Errorf("%d events were taken, want %d", got, maxClientEvents)
	}
}

func TestClientEventsToggle(t *testing.T) {
	t.Parallel()
	off, logs, _ := clientEventsHarness(t, func(c *Config) { c.DisableClientEvents = true })
	wantStatus(t, off.serve(t, report(`{"events":[{"k":"turnstile-success","ms":5}]}`)), http.StatusNoContent)
	if strings.Contains(logs.String(), "client event") {
		t.Errorf("a report was logged with reporting off:\n%s", logs)
	}
	// Off, the page loads no reporting script; on, it loads it before the others.
	for _, tt := range []struct {
		name string
		h    *harness
		want bool
	}{{"off", off, false}, {"on", func() *harness { h, _, _ := clientEventsHarness(t, nil); return h }(), true}} {
		page := tt.h.serve(t, request(http.MethodGet, "/login", ""))
		got := strings.Contains(page.Body.String(), `<script src="/static/events.js"></script>`)
		if got != tt.want {
			t.Errorf("%s: page loads events.js = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestClientEventsNeedNoSessionAndKeepTheSameOriginRule(t *testing.T) {
	t.Parallel()
	h, _, _ := clientEventsHarness(t, nil)
	cross := report(`{"events":[]}`)
	cross.Header.Set("Sec-Fetch-Site", "cross-site")
	wantStatus(t, h.serve(t, cross), http.StatusForbidden)
}
