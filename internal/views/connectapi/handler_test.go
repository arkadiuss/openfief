package connectapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
)

// specPath is the OpenAPI document generated from the read API's protobuf
// definitions, relative to this package.
const specPath = "../../../gen/openapi/openfief/views/v1/read_service.openapi.yaml"

// api is a handler over the synthetic ledger, with its log.
type api struct {
	handler http.Handler
	log     *bytes.Buffer
}

func newAPI(t *testing.T, l *fakeLedger) api {
	t.Helper()
	var log bytes.Buffer
	h, err := connectapi.New(newViews(t, l, app.WithJurisdiction(pl.NewPack(), fakeTerms{})), fakeVerifier{},
		connectapi.WithLogger(slog.New(slog.NewJSONHandler(&log, nil))),
		connectapi.WithRequestIDs(func() string { return "req-1" }))
	if err != nil {
		t.Fatalf("connectapi.New: %v", err)
	}
	return api{handler: h, log: &log}
}

// serve sends method target with token as bearer token ("" for none).
func (a api) serve(t *testing.T, method, target, token string) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), method, target, http.NoBody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	// The transcoder rewrites the request it serves; validation needs the
	// request as the client sent it.
	a.handler.ServeHTTP(rec, req.Clone(req.Context()))
	return req, rec
}

// contract is the generated OpenAPI document and a router over it.
type contract struct {
	doc    *openapi3.T
	router routers.Router
}

func loadContract(t *testing.T) contract {
	t.Helper()
	doc, err := openapi3.NewLoader().LoadFromFile(specPath)
	if err != nil {
		t.Fatalf("loading %s: %v", specPath, err)
	}
	if invalid := doc.Validate(context.Background()); invalid != nil {
		t.Fatalf("%s is not a valid OpenAPI document: %v", specPath, invalid)
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatalf("routing %s: %v", specPath, err)
	}
	return contract{doc: doc, router: router}
}

// check validates a 200 response rec gave to req against the generated
// OpenAPI document.
func (c contract) check(t *testing.T, req *http.Request, rec *httptest.ResponseRecorder) {
	t.Helper()
	route, params, err := c.router.FindRoute(req)
	if err != nil {
		t.Fatalf("%s %s is not in the OpenAPI document: %v", req.Method, req.URL, err)
	}
	in := &openapi3filter.RequestValidationInput{
		Request: req, PathParams: params, Route: route,
		Options: &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
	}
	out := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: in, Status: rec.Code, Header: rec.Header(),
		Body:    io.NopCloser(bytes.NewReader(rec.Body.Bytes())),
		Options: &openapi3filter.Options{IncludeResponseStatus: true, MultiError: true},
	}
	if err := openapi3filter.ValidateResponse(context.Background(), out); err != nil {
		t.Errorf("%s %s: response does not match the OpenAPI document: %v\n%s", req.Method, req.URL, err, rec.Body)
	}
}

func TestREST(t *testing.T) {
	t.Parallel()
	c := loadContract(t)
	tests := []struct {
		target string
		want   int
	}{
		{"/v1/sleeves", http.StatusOK},
		{"/v1/accounts", http.StatusOK},
		{"/v1/accounts/acc-1", http.StatusOK},
		{"/v1/accounts/acc-1/transactions", http.StatusOK},
		{"/v1/accounts/acc-1/transactions?pageSize=1", http.StatusOK},
		{"/v1/accounts/acc-1/holdings?date=2026-03-02", http.StatusOK},
		{"/v1/accounts/acc-1/lots?date=2026-03-02", http.StatusOK},
		{"/v1/accounts/acc-1/cash?date=2026-03-02", http.StatusOK},
		{"/v1/accounts/acc-1/balance-assertions", http.StatusOK},
		{"/v1/accounts/acc-1/balance-assertions?status=ASSERTION_STATUS_FLAGGED", http.StatusOK},
		{"/v1/accounts/acc-1/valuations?from=2026-03-01&to=2026-03-02", http.StatusOK},
		{"/v1/accounts/acc-1/valuation-status", http.StatusOK},
		{"/v1/valuations?from=2026-03-01&to=2026-03-02", http.StatusOK},
		{"/v1/valuation-status", http.StatusOK},
		{"/v1/allocation?date=2026-03-02&by=ALLOCATION_DIMENSION_ASSET_CLASS", http.StatusOK},
		{"/v1/allocation?date=2026-03-02&by=ALLOCATION_DIMENSION_SLEEVE", http.StatusOK},
		{"/v1/performance", http.StatusOK},
		{"/v1/performance?from=2026-03-01&to=2026-03-04", http.StatusOK},
		{"/v1/settings", http.StatusOK},
		{"/v1/yearly-taxes?year=2026", http.StatusOK},
		{"/v1/allocation?date=2026-03-02&by=ALLOCATION_DIMENSION_ASSET_CLASS&value=VALUE_KIND_NET", http.StatusOK},
		{"/v1/allocation?date=2026-03-02&by=ALLOCATION_DIMENSION_ASSET_CLASS&value=VALUE_KIND_LIQUIDATION", http.StatusOK},
		{"/v1/yearly-taxes?year=0", http.StatusBadRequest},
		{"/v1/accounts/nope", http.StatusNotFound},
		{"/v1/accounts/nope/holdings?date=2026-03-02", http.StatusNotFound},
		{"/v1/accounts/nope/valuation-status", http.StatusNotFound},
		{"/v1/nothing-here", http.StatusNotFound},
		{"/v1/accounts/acc-1/holdings", http.StatusBadRequest},
		{"/v1/accounts/acc-1/holdings?date=2026-02-30", http.StatusBadRequest},
		{"/v1/accounts/acc-1/lots?date=yesterday", http.StatusBadRequest},
		{"/v1/accounts/acc-1/transactions?pageSize=501", http.StatusBadRequest},
		{"/v1/accounts/acc-1/transactions?pageToken=bogus", http.StatusBadRequest},
		{"/v1/accounts/acc-1/valuations?from=2026-03-02&to=2026-03-01", http.StatusBadRequest},
		{"/v1/valuations?from=2000-01-01&to=2026-03-01", http.StatusBadRequest},
		{"/v1/allocation?date=2026-03-02", http.StatusBadRequest},
		{"/v1/performance?from=yesterday", http.StatusBadRequest},
		{"/v1/performance?from=2026-03-02&to=2026-03-01", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			t.Parallel()
			req, rec := newAPI(t, newLedger(t)).serve(t, http.MethodGet, tt.target, goodToken)
			if rec.Code != tt.want {
				t.Fatalf("GET %s = %d, want %d\n%s", tt.target, rec.Code, tt.want, rec.Body)
			}
			if rec.Code == http.StatusOK {
				c.check(t, req, rec)
			}
		})
	}
}

func TestRESTInternalError(t *testing.T) {
	t.Parallel()
	l := newLedger(t)
	l.fail = errStore
	a := newAPI(t, l)
	for _, target := range []string{"/v1/sleeves", "/v1/accounts/acc-1/cash?date=2026-03-02", "/v1/allocation?date=2026-03-02&by=ALLOCATION_DIMENSION_CURRENCY"} {
		_, rec := a.serve(t, http.MethodGet, target, goodToken)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("GET %s = %d, want 500", target, rec.Code)
		}
		if body := rec.Body.String(); strings.Contains(body, "store unavailable") || !strings.Contains(body, "req-1") {
			t.Errorf("GET %s: body %s should name the request ID and not the error", target, body)
		}
	}
	if !strings.Contains(a.log.String(), `"request_id":"req-1"`) {
		t.Errorf("a failure was not logged under its request ID:\n%s", a.log)
	}
}

func TestRESTUnauthenticated(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, header, wantChallenge string
		want                        int
	}{
		{"no token", "", "Bearer", http.StatusUnauthorized},
		{"other scheme", "Basic " + goodToken, "Bearer", http.StatusUnauthorized},
		{"empty bearer", "Bearer    ", "Bearer", http.StatusUnauthorized},
		{"unknown token", "Bearer synthetic-other-token", `Bearer error="invalid_token"`, http.StatusUnauthorized},
		{"token that cannot be checked", "Bearer " + brokenToken, "", http.StatusInternalServerError},
		{"scheme in any case", "bearer " + goodToken, "", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/accounts", http.NoBody)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			newAPI(t, newLedger(t)).handler.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("GET /v1/accounts = %d, want %d", rec.Code, tt.want)
			}
			if got := rec.Header().Get("WWW-Authenticate"); got != tt.wantChallenge {
				t.Errorf("WWW-Authenticate = %q, want %q", got, tt.wantChallenge)
			}
			if tt.want != http.StatusOK && strings.Contains(rec.Body.String(), "acc-1") {
				t.Errorf("an unauthenticated response leaked data: %s", rec.Body)
			}
		})
	}
}

func TestRESTReadOnly(t *testing.T) {
	t.Parallel()
	methods := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	for _, target := range []string{"/v1/accounts", "/v1/accounts/acc-1/transactions", "/v1/allocation"} {
		for _, method := range methods {
			t.Run(method+" "+target, func(t *testing.T) {
				t.Parallel()
				// Even without a token: a write is refused before anything else.
				for _, token := range []string{goodToken, ""} {
					_, rec := newAPI(t, newLedger(t)).serve(t, method, target, token)
					if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != http.MethodGet {
						t.Errorf("%s %s = %d, Allow %q; want 405, Allow GET", method, target, rec.Code, rec.Header().Get("Allow"))
					}
				}
			})
		}
	}
}

func TestOpenAPIIsReadOnly(t *testing.T) {
	t.Parallel()
	c := loadContract(t)
	if c.doc.Paths.Len() == 0 {
		t.Fatal("the OpenAPI document has no paths")
	}
	for path, item := range c.doc.Paths.Map() {
		for method, op := range item.Operations() {
			if method != http.MethodGet {
				t.Errorf("%s %s: the read API has only GET operations (D-6, D-7)", method, path)
			}
			if op.RequestBody != nil {
				t.Errorf("%s %s takes a request body: data never enters through the read API (D-7)", method, path)
			}
		}
		if !strings.HasPrefix(path, "/v1/") {
			t.Errorf("%s is not versioned under /v1", path)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	t.Parallel()
	a := newAPI(t, newLedger(t))
	want := map[string]string{
		"Cache-Control": "no-store", "X-Content-Type-Options": "nosniff", "X-Frame-Options": "DENY",
		"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'", "Referrer-Policy": "no-referrer",
		"X-Request-ID": "req-1",
	}
	for _, tc := range []struct {
		method, target, token string
		status                int
	}{
		{http.MethodGet, "/v1/accounts", goodToken, http.StatusOK},
		{http.MethodGet, "/v1/nothing-here", goodToken, http.StatusNotFound},
		{http.MethodGet, "/v1/accounts", "", http.StatusUnauthorized},
		{http.MethodPost, "/v1/accounts", goodToken, http.StatusMethodNotAllowed},
	} {
		_, rec := a.serve(t, tc.method, tc.target, tc.token)
		if rec.Code != tc.status {
			t.Errorf("%s %s = %d, want %d", tc.method, tc.target, rec.Code, tc.status)
		}
		for k, v := range want {
			if got := rec.Header().Get(k); got != v {
				t.Errorf("%s %s: %s = %q, want %q", tc.method, tc.target, k, got, v)
			}
		}
	}
}

func TestMoneyIsADecimalString(t *testing.T) {
	t.Parallel()
	_, rec := newAPI(t, newLedger(t)).serve(t, http.MethodGet, "/v1/accounts/acc-1/valuations?from=2026-03-02&to=2026-03-02", goodToken)
	var body struct {
		ReportingCurrency string `json:"reportingCurrency"`
		Valuations        []struct {
			Value          map[string]any `json:"value"`
			ValueReporting map[string]any `json:"valueReporting"`
		} `json:"valuations"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding %s: %v", rec.Body, err)
	}
	if len(body.Valuations) != 2 || body.ReportingCurrency != "PLN" {
		t.Fatalf("got %d valuations in %q, want 2 in PLN:\n%s", len(body.Valuations), body.ReportingCurrency, rec.Body)
	}
	if got := body.Valuations[1].Value; got["amount"] != "31.5" || got["currency"] != "USD" {
		t.Errorf("value = %v, want 31.5 USD: an exact decimal string in its original currency", got)
	}
	if got := body.Valuations[1].ValueReporting; got["amount"] != "126" || got["currency"] != "PLN" {
		t.Errorf("valueReporting = %v, want 126 PLN as a string", got)
	}
}

func TestLogNamesNoData(t *testing.T) {
	t.Parallel()
	l := newLedger(t)
	a := newAPI(t, l)
	for _, target := range []string{
		"/v1/accounts/acc-1/holdings?date=2026-03-02", "/v1/accounts/acc-1/cash?date=2026-03-02",
		"/v1/accounts/nope", "/v1/allocation?date=2026-03-02&by=ALLOCATION_DIMENSION_INSTITUTION",
	} {
		a.serve(t, http.MethodGet, target, goodToken)
	}
	l.fail = errStore
	a.serve(t, http.MethodGet, "/v1/accounts/acc-1/lots?date=2026-03-02", goodToken)
	a.serve(t, http.MethodGet, "/v1/accounts", brokenToken)
	logged := a.log.String()
	if !strings.Contains(logged, `"request_id":"req-1"`) {
		t.Fatalf("requests are not logged by request ID:\n%s", logged)
	}
	for _, secret := range []string{"acc-1", "nope", "broker-a", "Synthetic", "EDO0334", isinIE, "100.00", "1043.21", "USD", goodToken, brokenToken, "store unavailable", "/v1/"} {
		if strings.Contains(logged, secret) {
			t.Errorf("the log contains %q:\n%s", secret, logged)
		}
	}
}

// amount is a Money message's JSON.
type amount struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// performanceDay is a PerformanceDay message's JSON, without in transit.
type performanceDay struct {
	Date     string `json:"date"`
	Index    string `json:"index"`
	Value    amount `json:"valueReporting"`
	Holdings amount `json:"holdingsReporting"`
	Cash     amount `json:"cashReporting"`
	Flow     amount `json:"flowReporting"`
}

// performanceBody is a GetPerformanceResponse message's JSON.
type performanceBody struct {
	From               string           `json:"from"`
	To                 string           `json:"to"`
	Gap                string           `json:"gap"`
	RevaluationPending bool             `json:"revaluationPending"`
	ReportingCurrency  string           `json:"reportingCurrency"`
	Days               []performanceDay `json:"days"`
}

func TestRESTPerformance(t *testing.T) {
	t.Parallel()
	_, rec := newAPI(t, newLedger(t)).serve(t, http.MethodGet, "/v1/performance?to=2026-03-04", goodToken)
	var body performanceBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding %s: %v", rec.Body, err)
	}
	// Go-live is 2026-03-02, valued through that day: 1043.21 + 126 PLN of
	// holdings, -90 USD cash at 4.00, and the opening lot flowing in at its
	// valuation; 03-03 is not valued.
	pln := func(a string) amount { return amount{Amount: a, Currency: "PLN"} }
	want := performanceBody{
		From: "2026-03-02", To: "2026-03-04", Gap: "PERFORMANCE_GAP_UNVALUED", RevaluationPending: true, ReportingCurrency: "PLN",
		Days: []performanceDay{{"2026-03-02", "100", pln("809.21"), pln("1169.21"), pln("-360"), pln("1043.21")}},
	}
	if diff := cmp.Diff(want, body); diff != "" {
		t.Errorf("GET /v1/performance mismatch (-want +got):\n%s", diff)
	}
}
