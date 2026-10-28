package nbp_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/areknoster/hypert"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
	"github.com/arkadiuss/openfief/internal/valuation/nbp"
)

// recordNBP switches the hypert tests to record against the real
// api.nbp.pl. Committed code always replays: turn it on only locally,
// following .claude/skills/hypert-recording/SKILL.md, then turn it off.
const recordNBP = false

// The recordings under testdata/ come from the public api.nbp.pl, which
// holds no account data: table A is published for everyone. Recording drops
// the WAF's Set-Cookie header, the only per-client value in a response.
func recordingClient(t *testing.T) *http.Client {
	t.Helper()
	dropCookies := hypert.ResponseTransformFunc(func(r *http.Response) *http.Response {
		r.Header.Del("Set-Cookie")
		return r
	})
	return hypert.TestClient(t, recordNBP,
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

func TestRatesTableA(t *testing.T) {
	t.Parallel()
	c := nbp.New(recordingClient(t), nbp.DefaultBaseURL)
	// Thursday 2 to Tuesday 7 January 2025: the weekend and Epiphany
	// (Monday 6) have no table.
	got, err := c.Rates(context.Background(), date(t, "2025-01-02"), date(t, "2025-01-07"))
	if err != nil {
		t.Fatalf("Rates: %v", err)
	}
	perDate := map[string]int{}
	byKey := map[string]valuation.FXRate{}
	for _, r := range got {
		perDate[r.Date.String()]++
		byKey[r.Currency.String()+" "+r.Date.String()] = r
	}
	wantPerDate := map[string]int{"2025-01-02": 33, "2025-01-03": 33, "2025-01-07": 33}
	if diff := cmp.Diff(wantPerDate, perDate); diff != "" {
		t.Errorf("rates per table date (-want +got):\n%s", diff)
	}
	// Spot checks, copied from the recorded response text.
	for _, want := range []valuation.FXRate{
		rate(t, "USD", "2025-01-02", "4.1219"),
		rate(t, "EUR", "2025-01-03", "4.2718"),
		rate(t, "HUF", "2025-01-07", "0.010226"),
		rate(t, "IDR", "2025-01-07", "0.00025276"),
		rate(t, "XDR", "2025-01-02", "5.3541"),
	} {
		key := want.Currency.String() + " " + want.Date.String()
		if diff := cmp.Diff(want, byKey[key]); diff != "" {
			t.Errorf("%s (-want +got):\n%s", key, diff)
		}
	}
}

func TestRatesNoData(t *testing.T) {
	t.Parallel()
	c := nbp.New(recordingClient(t), nbp.DefaultBaseURL)
	// Saturday 4 to Monday 6 January 2025 (Epiphany): NBP answers 404.
	got, err := c.Rates(context.Background(), date(t, "2025-01-04"), date(t, "2025-01-06"))
	if err != nil {
		t.Fatalf("Rates: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d rates, want none", len(got))
	}
}

func TestRatesKeepsEveryDigit(t *testing.T) {
	t.Parallel()
	// More digits than a float64 holds: a float would round them.
	body := `[{"table":"A","no":"001/A/NBP/2025","effectiveDate":"2025-01-02","rates":[` +
		`{"currency":"dolar amerykański","code":"USD","mid":4.10120000000000000001},` +
		`{"currency":"rupia indonezyjska","code":"IDR","mid":2.54E-4}]}]`
	c := nbp.New(stubClient(http.StatusOK, body), nbp.DefaultBaseURL)
	got, err := c.Rates(context.Background(), date(t, "2025-01-02"), date(t, "2025-01-02"))
	if err != nil {
		t.Fatalf("Rates: %v", err)
	}
	want := []valuation.FXRate{
		rate(t, "USD", "2025-01-02", "4.10120000000000000001"),
		rate(t, "IDR", "2025-01-02", "0.000254"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func TestRatesRejects(t *testing.T) {
	t.Parallel()
	const head = `[{"table":"A","no":"001/A/NBP/2025","effectiveDate":"2025-01-02","rates":[`
	tests := []struct {
		name    string
		from    string
		to      string
		status  int
		body    string
		wantErr error
	}{
		{name: "range over 93 days", from: "2025-01-01", to: "2025-04-04", wantErr: app.ErrInvalidRange},
		{name: "end before start", from: "2025-01-03", to: "2025-01-02", wantErr: app.ErrInvalidRange},
		{name: "server error", status: http.StatusInternalServerError, body: "", wantErr: nbp.ErrUnexpectedStatus},
		{name: "range limit refused", status: http.StatusBadRequest, body: "400 BadRequest", wantErr: nbp.ErrUnexpectedStatus},
		{name: "not JSON", status: http.StatusOK, body: "<html>", wantErr: nbp.ErrMalformedTable},
		{name: "another table", status: http.StatusOK, body: `[{"table":"B","effectiveDate":"2025-01-02","rates":[]}]`, wantErr: nbp.ErrMalformedTable},
		{name: "bad date", status: http.StatusOK, body: `[{"table":"A","effectiveDate":"02.01.2025","rates":[]}]`, wantErr: nbp.ErrMalformedTable},
		{name: "bad currency", status: http.StatusOK, body: head + `{"code":"usd","mid":4.1}]}]`, wantErr: nbp.ErrMalformedTable},
		{name: "zero rate", status: http.StatusOK, body: head + `{"code":"USD","mid":0}]}]`, wantErr: nbp.ErrMalformedTable},
		{name: "rate as text", status: http.StatusOK, body: head + `{"code":"USD","mid":"n/a"}]}]`, wantErr: nbp.ErrMalformedTable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			from, to := "2025-01-02", "2025-01-02"
			if tt.from != "" {
				from, to = tt.from, tt.to
			}
			c := nbp.New(stubClient(tt.status, tt.body), nbp.DefaultBaseURL)
			_, err := c.Rates(context.Background(), date(t, from), date(t, to))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRatesSendsItsOwnUserAgent(t *testing.T) {
	t.Parallel()
	// api.nbp.pl refuses Go's default User-Agent with 403.
	st := &stubTransport{status: http.StatusNotFound}
	c := nbp.New(&http.Client{Transport: st}, nbp.DefaultBaseURL)
	if _, err := c.Rates(context.Background(), date(t, "2025-01-04"), date(t, "2025-01-06")); err != nil {
		t.Fatalf("Rates: %v", err)
	}
	if got := st.userAgent; got != "openfief" {
		t.Errorf("User-Agent %q, want %q", got, "openfief")
	}
}

// stubTransport answers every request with one canned response, so that
// parsing and status handling are tested without a network.
type stubTransport struct {
	status    int
	body      string
	userAgent string
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s.userAgent = req.Header.Get("User-Agent")
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

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func rate(t *testing.T, code, on, value string) valuation.FXRate {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", code, err)
	}
	d, err := kernel.ParseDecimal(value)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", value, err)
	}
	return valuation.FXRate{Currency: c, Date: date(t, on), Rate: d, Source: nbp.Source}
}
