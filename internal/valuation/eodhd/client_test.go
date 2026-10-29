package eodhd_test

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/areknoster/hypert"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
)

// recordEODHD would switch the hypert tests to record against the real
// eodhd.com. It stays off: the recordings under testdata/ are SYNTHETIC,
// hand-written in the response shape EODHD documents for /api/search and
// /api/eod, with made-up ISINs, tickers and prices, because recording needs
// an API key. Re-recording follows .claude/skills/hypert-recording/SKILL.md.
const recordEODHD = false

// fakeKey is the API key the tests send; recordings hold "SANITIZED".
const fakeKey = "synthetic-test-key"

// recordingClient replays testdata/<test>/. The API key is sanitized in
// the stored request and in the replayed one, so a recording never holds
// it and replay still checks every other query parameter.
func recordingClient(t *testing.T) *http.Client {
	t.Helper()
	return hypert.TestClient(t, recordEODHD,
		hypert.WithRequestSanitizer(hypert.ComposedRequestSanitizer(
			hypert.DefaultRequestSanitizer(), hypert.SanitizerQueryParams("api_token"),
		)),
		hypert.WithRequestValidator(hypert.ComposedRequestValidator(
			hypert.PathValidator(), hypert.MethodValidator(), hypert.QueryParamsValidator(),
			hypert.SchemeValidator(), userAgentAndHeadersValidator(),
		)))
}

// userAgentAndHeadersValidator is hypert's HeadersValidator, which skips
// the recorded User-Agent because Go's transport usually adds it; the
// adapter sets its own, so it is compared here instead.
func userAgentAndHeadersValidator() hypert.RequestValidatorFunc {
	return hypert.RequestValidatorFunc(func(t hypert.T, recorded, got hypert.RequestData) error {
		if r, g := recorded.Headers.Get("User-Agent"), got.Headers.Get("User-Agent"); r != g {
			t.Errorf("expected header 'User-Agent' to be '%s', got '%s'", r, g)
		}
		got.Headers = got.Headers.Clone()
		got.Headers.Del("User-Agent")
		return hypert.HeadersValidator().Validate(t, recorded, got)
	})
}

func TestListings(t *testing.T) {
	t.Parallel()
	c := eodhd.New(recordingClient(t), eodhd.DefaultBaseURL, fakeKey)
	synt := isin(t, "XX00000SYNT3")
	got, err := c.Listings(t.Context(), synt)
	if err != nil {
		t.Fatalf("Listings: %v", err)
	}
	// The result without an ISIN, a fuzzy match on the name, is left out.
	want := []eodhd.Listing{
		{ISIN: synt, Code: "SYNT", Exchange: "XETRA", Currency: "EUR"},
		{ISIN: synt, Code: "SYNT", Exchange: "LSE", Currency: "GBX"},
		{ISIN: synt, Code: "SYN1", Exchange: "F", Currency: "EUR"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Listings (-want +got):\n%s", diff)
	}
}

func TestListingsNone(t *testing.T) {
	t.Parallel()
	c := eodhd.New(recordingClient(t), eodhd.DefaultBaseURL, fakeKey)
	got, err := c.Listings(t.Context(), isin(t, "XX000000NON6"))
	if err != nil || len(got) != 0 {
		t.Errorf("Listings = %v, %v; want none, no error", got, err)
	}
}

func TestCloses(t *testing.T) {
	t.Parallel()
	c := eodhd.New(recordingClient(t), eodhd.DefaultBaseURL, fakeKey)
	got, err := c.Closes(t.Context(), "SYNT.XETRA", date(t, "2025-01-02"), date(t, "2025-01-07"))
	if err != nil {
		t.Fatalf("Closes: %v", err)
	}
	want := []eodhd.Close{
		closing(t, "2025-01-02", "100.52"), closing(t, "2025-01-03", "101.075"),
		closing(t, "2025-01-06", "101.37"), closing(t, "2025-01-07", "102"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Closes (-want +got):\n%s", diff)
	}
}

func TestClosesWholeHistory(t *testing.T) {
	t.Parallel()
	c := eodhd.New(recordingClient(t), eodhd.DefaultBaseURL, fakeKey)
	// A zero from sends no from parameter: EODHD's whole history.
	got, err := c.Closes(t.Context(), "FAKE20.INDX", kernel.Date{}, date(t, "2001-01-04"))
	if err != nil {
		t.Fatalf("Closes: %v", err)
	}
	want := []eodhd.Close{
		closing(t, "2001-01-02", "1505.12"), closing(t, "2001-01-03", "1522.6"), closing(t, "2001-01-04", "1519.99"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Closes (-want +got):\n%s", diff)
	}
}

func TestClosesUnknownTicker(t *testing.T) {
	t.Parallel()
	c := eodhd.New(recordingClient(t), eodhd.DefaultBaseURL, fakeKey)
	_, err := c.Closes(t.Context(), "NOPE.WAR", date(t, "2025-01-02"), date(t, "2025-01-07"))
	if !errors.Is(err, eodhd.ErrNotFound) {
		t.Errorf("Closes error = %v, want ErrNotFound", err)
	}
}

func TestClosesParsesExactly(t *testing.T) {
	t.Parallel()
	// More digits than a float64 holds, an exponent, a day out of order, and
	// days without a positive close.
	body := `[{"date":"2025-01-03","close":4.10120000000000000001},{"date":"2025-01-02","close":2.54E-1},` +
		`{"date":"2025-01-06","close":null},{"date":"2025-01-07","close":0},{"date":"2025-01-08","close":-1}]`
	c := eodhd.New(stubClient(http.StatusOK, body), eodhd.DefaultBaseURL, fakeKey)
	got, err := c.Closes(t.Context(), "SYNT.XETRA", date(t, "2025-01-02"), date(t, "2025-01-08"))
	if err != nil {
		t.Fatalf("Closes: %v", err)
	}
	want := []eodhd.Close{closing(t, "2025-01-02", "0.254"), closing(t, "2025-01-03", "4.10120000000000000001")}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Closes (-want +got):\n%s", diff)
	}
}

func TestClientRejects(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr error
	}{
		{name: "wrong API key", status: http.StatusUnauthorized, body: "Unauthenticated", wantErr: eodhd.ErrUnexpectedStatus},
		{name: "plan limit used up", status: http.StatusPaymentRequired, wantErr: eodhd.ErrUnexpectedStatus},
		{name: "server error", status: http.StatusInternalServerError, wantErr: eodhd.ErrUnexpectedStatus},
		{name: "not JSON", status: http.StatusOK, body: "<html>", wantErr: eodhd.ErrMalformedResponse},
		{name: "an error object", status: http.StatusOK, body: `{"error":"x"}`, wantErr: eodhd.ErrMalformedResponse},
		{name: "bad date", status: http.StatusOK, body: `[{"date":"02.01.2025","close":1}]`, wantErr: eodhd.ErrMalformedResponse},
		{name: "close as text", status: http.StatusOK, body: `[{"date":"2025-01-02","close":"n/a"}]`, wantErr: eodhd.ErrMalformedResponse},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := eodhd.New(stubClient(tt.status, tt.body), eodhd.DefaultBaseURL, fakeKey)
			_, err := c.Closes(t.Context(), "SYNT.XETRA", date(t, "2025-01-02"), date(t, "2025-01-02"))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClientNeverReturnsTheKey(t *testing.T) {
	t.Parallel()
	failing := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errNetwork
	})}
	c := eodhd.New(failing, eodhd.DefaultBaseURL, fakeKey)
	_, closesErr := c.Closes(t.Context(), "SYNT.XETRA", date(t, "2025-01-02"), date(t, "2025-01-02"))
	_, listErr := c.Listings(t.Context(), isin(t, "XX00000SYNT3"))
	for _, err := range []error{closesErr, listErr} {
		if !errors.Is(err, errNetwork) {
			t.Errorf("error = %v, want the transport's error", err)
		}
		if err != nil && strings.Contains(err.Error(), fakeKey) {
			t.Errorf("error %q contains the API key", err)
		}
		if _, ok := errors.AsType[*url.Error](err); ok {
			t.Errorf("error %q still carries the request URL", err)
		}
	}
}

var errNetwork = errors.New("synthetic: connection refused")

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// stubClient answers every request with one canned response.
func stubClient(status int, body string) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}, Request: req,
		}, nil
	})}
}

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func isin(t *testing.T, s string) ledger.ISIN {
	t.Helper()
	i, err := ledger.ParseISIN(s)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", s, err)
	}
	return i
}

func closing(t *testing.T, on, value string) eodhd.Close {
	t.Helper()
	d, err := kernel.ParseDecimal(value)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", value, err)
	}
	return eodhd.Close{Date: date(t, on), Close: d}
}
