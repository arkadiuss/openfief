package connectapi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/go-cmp/cmp"
)

// errWrongResponse reports an operation's response that is not the one
// the document must declare.
var errWrongResponse = errors.New("not the declared error response")

// restError is an error's body over REST: the google.rpc.Status Vanguard
// writes, with the numeric gRPC code.
type restError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Details []json.RawMessage `json:"details"`
}

// TestRESTErrorsMatchContract checks that the read API's errors over REST
// are the responses the OpenAPI document declares for their operation and
// status (D-200#7): 401 without a valid token, 404 for an unknown ID in the
// path and for the model portfolio while it is off (D-52), and 400 for an
// invalid argument. Each body validates against the document and is the
// error's status, its fixed text and no details.
func TestRESTErrorsMatchContract(t *testing.T) {
	t.Parallel()
	c := loadContract(t)
	tests := []struct {
		name, target, token string
		api                 func(*testing.T) api
		wantStatus          int
		want                restError
		wantChallenge       string
	}{
		{
			"no token", "/v1/accounts", "", ledgerAPI, http.StatusUnauthorized,
			restError{16, "a bearer token is required", []json.RawMessage{}},
			"Bearer",
		},
		{
			"a token the core never issued", "/v1/accounts/acc-1/holdings?date=2026-03-02", "synthetic-other-token", ledgerAPI, http.StatusUnauthorized,
			restError{16, "the bearer token is not valid", []json.RawMessage{}},
			`Bearer error="invalid_token"`,
		},
		{
			"unknown account", "/v1/accounts/nope", goodToken, ledgerAPI, http.StatusNotFound,
			restError{5, "no such account", []json.RawMessage{}},
			"",
		},
		{
			"unknown account's holdings", "/v1/accounts/nope/holdings?date=2026-03-02", goodToken, ledgerAPI, http.StatusNotFound,
			restError{5, "no such account", []json.RawMessage{}},
			"",
		},
		{
			"unknown instrument", "/v1/instruments/XX0000000001", goodToken, uiv2API, http.StatusNotFound,
			restError{5, "no such instrument", []json.RawMessage{}},
			"",
		},
		{
			"unknown connector instance", "/v1/connector-instances/nbp-fx/sync-attempts", goodToken, uiv2API, http.StatusNotFound,
			restError{5, "no such connector instance", []json.RawMessage{}},
			"",
		},
		{
			"model portfolio off", "/v1/model-portfolio", goodToken, modelOffAPI, http.StatusNotFound,
			restError{5, "no model portfolio", []json.RawMessage{}},
			"",
		},
		{
			"date range ending before it starts", "/v1/valuations?from=2026-03-02&to=2026-03-01", goodToken, ledgerAPI, http.StatusBadRequest,
			restError{3, "the date range ends before it starts or is too long", []json.RawMessage{}},
			"",
		},
		{
			"malformed date", "/v1/accounts/acc-1/lots?date=yesterday", goodToken, ledgerAPI, http.StatusBadRequest,
			restError{3, "a request field is missing or malformed", []json.RawMessage{}},
			"",
		},
		{
			"page token the API never returned", "/v1/accounts/acc-1/transactions?pageToken=bogus", goodToken, ledgerAPI, http.StatusBadRequest,
			restError{3, "the page token is not one this API returned", []json.RawMessage{}},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req, rec := tt.api(t).serve(t, http.MethodGet, tt.target, tt.token)
			if rec.Code != tt.wantStatus {
				t.Fatalf("GET %s = %d, want %d\n%s", tt.target, rec.Code, tt.wantStatus, rec.Body)
			}
			c.check(t, req, rec)
			dec := json.NewDecoder(bytes.NewReader(rec.Body.Bytes()))
			dec.DisallowUnknownFields()
			var got restError
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("decoding %s: %v", rec.Body, err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GET %s body (-want +got):\n%s", tt.target, diff)
			}
			if got := rec.Header().Get("WWW-Authenticate"); got != tt.wantChallenge {
				t.Errorf("WWW-Authenticate = %q, want %q", got, tt.wantChallenge)
			}
		})
	}
}

// TestOpenAPIDeclaresErrors checks that every operation of the generated
// document declares 400 and 401, every operation with a path parameter
// declares 404, and GetModelPortfolio declares 404 for the model portfolio
// being off, each a google.rpc.Status (D-200#7).
func TestOpenAPIDeclaresErrors(t *testing.T) {
	t.Parallel()
	c := loadContract(t)
	for path, item := range c.doc.Paths.Map() {
		for method, op := range item.Operations() {
			for status, name := range declaredErrors(path, op.OperationID) {
				if err := isStatusResponse(op.Responses.Value(status), name); err != nil {
					t.Errorf("%s %s: response %s: %v", method, path, status, err)
				}
			}
		}
	}
}

// declaredErrors returns the responses, by status, the operation id at
// path must declare.
func declaredErrors(path, id string) map[string]string {
	want := map[string]string{"400": "BadRequest", "401": "Unauthorized"}
	if strings.ContainsRune(path, '{') {
		want["404"] = "NotFound"
	}
	if id == "openfief.views.v1.ReadService.GetModelPortfolio" {
		want["404"] = "ModelPortfolioOff"
	}
	return want
}

// isStatusResponse returns why ref is not a reference to the response
// name with a google.rpc.Status body, or nil.
func isStatusResponse(ref *openapi3.ResponseRef, name string) error {
	if ref == nil || ref.Ref != "#/components/responses/"+name {
		return fmt.Errorf("%w: %v, want a reference to %s", errWrongResponse, ref, name)
	}
	schema := ref.Value.Content.Get("application/json").Schema.Value
	if len(schema.AllOf) == 0 || schema.AllOf[0].Ref != "#/components/schemas/google.rpc.Status" {
		return fmt.Errorf("%w: %s is not a google.rpc.Status", errWrongResponse, name)
	}
	return nil
}

// ledgerAPI serves the read API over newLedger.
func ledgerAPI(t *testing.T) api {
	t.Helper()
	return newAPI(t, newLedger(t))
}

// modelOffAPI serves the read API with the model portfolio off.
func modelOffAPI(t *testing.T) api {
	t.Helper()
	return api{handler: modelHandler(t, false), log: &bytes.Buffer{}}
}

// TestOpenAPIRefusesWrongErrors checks that the document's error responses
// constrain their bodies, so the checks above can fail: a status whose body
// carries another code, or no google.rpc.Status at all, does not validate.
func TestOpenAPIRefusesWrongErrors(t *testing.T) {
	t.Parallel()
	c := loadContract(t)
	tests := []struct {
		name, target, body string
		status             int
	}{
		{"404 with permission denied's code", "/v1/accounts/nope", `{"code":7,"message":"no","details":[]}`, http.StatusNotFound},
		{"401 with not found's code", "/v1/accounts", `{"code":5,"message":"no","details":[]}`, http.StatusUnauthorized},
		{"400 with unauthenticated's code", "/v1/valuations", `{"code":16,"message":"no","details":[]}`, http.StatusBadRequest},
		{"model portfolio off with internal's code", "/v1/model-portfolio", `{"code":13,"message":"no","details":[]}`, http.StatusNotFound},
		{"a code that is no number", "/v1/accounts/nope", `{"code":"not_found","message":"no","details":[]}`, http.StatusNotFound},
		{"details that are no list", "/v1/accounts/nope", `{"code":5,"message":"no","details":{}}`, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.target, http.NoBody)
			rec := httptest.NewRecorder()
			rec.Header().Set("Content-Type", "application/json")
			rec.WriteHeader(tt.status)
			if _, err := rec.WriteString(tt.body); err != nil {
				t.Fatal(err)
			}
			if err := c.validate(req, rec); err == nil {
				t.Errorf("GET %s = %d %s validated against the document", tt.target, tt.status, tt.body)
			}
		})
	}
}
