package gus_test

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/areknoster/hypert"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/gus"
)

// recordGUS switches the hypert tests to record against the real
// api-sdp.stat.gov.pl. Committed code always replays: turn it on only
// locally, following .claude/skills/hypert-recording/SKILL.md, then turn it
// off.
const recordGUS = false

// The recordings under testdata/ come from GUS's public SDP API, which
// holds no account data: CPI is published for everyone, and the requests
// carry no key. Recording drops the Set-Cookie header, the only per-client
// value in a response.
func recordingClient(t *testing.T) *http.Client {
	t.Helper()
	dropCookies := hypert.ResponseTransformFunc(func(r *http.Response) *http.Response {
		r.Header.Del("Set-Cookie")
		return r
	})
	return hypert.TestClient(t, recordGUS,
		hypert.WithResponseTransform(hypert.TransformRespModeOnRecord, dropCookies),
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

// cpi is the CPI reading of year-month with value, a fraction.
func cpi(t *testing.T, year int, month time.Month, value string) valuation.IndexReading {
	t.Helper()
	v, err := kernel.ParseDecimal(value)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", value, err)
	}
	return valuation.IndexReading{Kind: valuation.CPI, Year: year, Month: month, Value: v, Source: gus.Source}
}

// TestCPIFullYear replays 2025, the last year of the COICOP 1999 indicator,
// as recorded on 2026-10-29. September's 102.9 is the 2.9% that, with a
// margin of 1.50%, gives EDO1133 its 4.40% from November 2025 (interest
// tables at https://www.obligacjeskarbowe.pl/tabela-odsetkowa/).
func TestCPIFullYear(t *testing.T) {
	t.Parallel()
	got, err := gus.New(recordingClient(t), gus.DefaultBaseURL).CPI(context.Background(), 2025)
	if err != nil {
		t.Fatalf("CPI: %v", err)
	}
	values := []string{"0.049", "0.049", "0.049", "0.043", "0.04", "0.041", "0.031", "0.029", "0.029", "0.028", "0.025", "0.024"}
	want := make([]valuation.IndexReading, 0, len(values))
	for i, v := range values {
		want = append(want, cpi(t, 2025, time.Month(i+1), v))
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

// TestCPIYearSoFar replays 2026, from the COICOP 2018 indicator, as
// recorded on 2026-10-29: January to September are published.
func TestCPIYearSoFar(t *testing.T) {
	t.Parallel()
	got, err := gus.New(recordingClient(t), gus.DefaultBaseURL).CPI(context.Background(), 2026)
	if err != nil {
		t.Fatalf("CPI: %v", err)
	}
	values := []string{"0.021", "0.021", "0.03", "0.032", "0.031", "0.025", "0.03", "0.034", "0.041"}
	want := make([]valuation.IndexReading, 0, len(values))
	for i, v := range values {
		want = append(want, cpi(t, 2026, time.Month(i+1), v))
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

// TestCPINothingPublished replays 2027, for which GUS answers 404.
func TestCPINothingPublished(t *testing.T) {
	t.Parallel()
	got, err := gus.New(recordingClient(t), gus.DefaultBaseURL).CPI(context.Background(), 2027)
	if err != nil {
		t.Fatalf("CPI: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d readings, want none", len(got))
	}
}

// row is one synthetic value of the CPI indicator for 2025.
func row(period, flag int, value string) string {
	return `{"id-wskaznik":1832,"id-daty":2025,"id-okres":` + strconv.Itoa(period) + `,"id-sposob-prezentacji-miara":5,"id-flaga":` +
		strconv.Itoa(flag) + `,"wartosc":` + value + `}`
}

func TestCPIStopsAtTheFirstMonthNotFinal(t *testing.T) {
	t.Parallel()
	const final = 36 // "nd": not applicable, GUS's flag on a plain value
	tests := []struct {
		name string
		rows []string
		want int
	}{
		{name: "every month final", rows: []string{row(247, final, "104.9"), row(248, final, "104.90000")}, want: 2},
		{name: "a cumulative period is skipped", rows: []string{row(247, final, "104.9"), row(259, final, "104.9")}, want: 1},
		{name: "preliminary data stop the year", rows: []string{row(247, final, "104.9"), row(248, 31, "104.8"), row(249, final, "104.7")}, want: 1},
		{name: "an estimate", rows: []string{row(247, 29, "104.9")}, want: 0},
		{name: "a forecast", rows: []string{row(247, 30, "104.9")}, want: 0},
		{name: "no value", rows: []string{row(247, final, "null")}, want: 0},
		{name: "a month missing", rows: []string{row(247, final, "104.9"), row(249, final, "104.7")}, want: 1},
		{name: "out of order", rows: []string{row(248, final, "104.8"), row(247, final, "104.9")}, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body := "[" + strings.Join(tt.rows, ",") + "]"
			got, err := gus.New(stubClient(http.StatusOK, body), gus.DefaultBaseURL).CPI(context.Background(), 2025)
			if err != nil {
				t.Fatalf("CPI: %v", err)
			}
			if len(got) != tt.want {
				t.Fatalf("got %d readings, want %d", len(got), tt.want)
			}
			for i, r := range got {
				if r.Month != time.Month(i+1) {
					t.Errorf("reading %d is for month %d, want %d", i, r.Month, i+1)
				}
			}
		})
	}
}

func TestCPIRejects(t *testing.T) {
	t.Parallel()
	const final = 36
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr error
	}{
		{name: "server error", status: http.StatusInternalServerError, wantErr: gus.ErrUnexpectedStatus},
		{name: "rate limited", status: http.StatusTooManyRequests, wantErr: gus.ErrUnexpectedStatus},
		{name: "not JSON", status: http.StatusOK, body: "<html>", wantErr: gus.ErrMalformedData},
		{name: "another year", status: http.StatusOK, body: `[{"id-daty":2024,"id-okres":247,"id-sposob-prezentacji-miara":5,"id-flaga":36,"wartosc":104.9}]`, wantErr: gus.ErrMalformedData},
		{name: "month on month", status: http.StatusOK, body: `[{"id-daty":2025,"id-okres":247,"id-sposob-prezentacji-miara":2,"id-flaga":36,"wartosc":100.1}]`, wantErr: gus.ErrMalformedData},
		{name: "a month twice", status: http.StatusOK, body: "[" + row(247, final, "104.9") + "," + row(247, final, "104.8") + "]", wantErr: gus.ErrMalformedData},
		{name: "value as text", status: http.StatusOK, body: "[" + row(247, final, `"n/a"`) + "]", wantErr: gus.ErrMalformedData},
		{name: "zero index", status: http.StatusOK, body: "[" + row(247, final, "0") + "]", wantErr: gus.ErrMalformedData},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := gus.New(stubClient(tt.status, tt.body), gus.DefaultBaseURL).CPI(context.Background(), 2025)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("error (-want +got):\n%s", diff)
			}
		})
	}
}

// stubTransport answers every request with one canned response, so that
// parsing and status handling are tested without a network.
type stubTransport struct {
	status int
	body   string
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: s.status,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Header:     http.Header{},
		Request:    req,
	}, nil
}

func stubClient(status int, body string) *http.Client {
	return &http.Client{Transport: &stubTransport{status: status, body: body}}
}
