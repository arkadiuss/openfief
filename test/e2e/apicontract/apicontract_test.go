package apicontract_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/go-cmp/cmp"
	_ "google.golang.org/genproto/googleapis/api/annotations" // google.api.http, for the synthetic service's text
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/descriptorpb"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/test/e2e/apicontract"
)

// specPath is the read API's generated OpenAPI document, relative to this
// package.
const specPath = "../../../gen/openapi/openfief/views/v1/read_service.openapi.yaml"

// readServiceRoutes returns the read API's routes.
func readServiceRoutes(t *testing.T) []*apicontract.Route {
	t.Helper()
	routes, err := apicontract.Routes(viewsv1.File_openfief_views_v1_read_service_proto.Services().ByName("ReadService"))
	if err != nil {
		t.Fatalf("Routes: %v", err)
	}
	return routes
}

// readRoutes returns the read API's routes by method name.
func readRoutes(t *testing.T) map[string]*apicontract.Route {
	t.Helper()
	byName := map[string]*apicontract.Route{}
	for _, r := range readServiceRoutes(t) {
		byName[r.Name] = r
	}
	return byName
}

func TestReadServiceRoutesMatchTheOpenAPIDocument(t *testing.T) {
	t.Parallel()
	doc, err := openapi3.NewLoader().LoadFromFile(specPath)
	if err != nil {
		t.Fatalf("loading %s: %v", specPath, err)
	}
	routes := readServiceRoutes(t)
	if diff := cmp.Diff(apicontract.RouteOperations(routes), apicontract.Operations(doc)); diff != "" {
		t.Errorf("the OpenAPI document's operations differ from ReadService's routes (-routes +document):\n%s", diff)
	}
	for _, r := range routes {
		if !r.ReadOnly || r.Operation.Method != http.MethodGet {
			t.Errorf("%s: read-only %t, method %s; want a read-only GET", r.Name, r.ReadOnly, r.Operation.Method)
		}
	}
	got := readRoutes(t)["GetAccount"]
	want := apicontract.Operation{Method: http.MethodGet, Path: "/v1/accounts/{accountId}", ID: "openfief.views.v1.ReadService.GetAccount"}
	if got.Operation != want || got.Procedure != "/openfief.views.v1.ReadService/GetAccount" {
		t.Errorf("GetAccount = %+v, %s; want %+v", got.Operation, got.Procedure, want)
	}
}

func TestTarget(t *testing.T) {
	t.Parallel()
	routes := readRoutes(t)
	tests := []struct {
		name, method string
		fields       map[string]string
		want         string
		wantErr      error
	}{
		{name: "no fields", method: "ListAccounts", want: "/v1/accounts"},
		{name: "path variable escaped", method: "GetAccount", fields: map[string]string{"accountId": "acc 1/x"}, want: "/v1/accounts/acc%201%2Fx"},
		{
			name: "query in name order", method: "ListAccountValuations",
			fields: map[string]string{"accountId": "acc-1", "to": "2026-03-13", "from": "2026-03-06"},
			want:   "/v1/accounts/acc-1/valuations?from=2026-03-06&to=2026-03-13",
		},
		{
			name: "two path segments", method: "ListSyncAttempts", fields: map[string]string{"connectorInstance": "dp-ibkr", "pageSize": "5"},
			want: "/v1/connector-instances/dp-ibkr/sync-attempts?pageSize=5",
		},
		{name: "missing path variable", method: "GetInstrument", wantErr: apicontract.ErrMissingPathField},
		{name: "empty path variable", method: "GetInstrument", fields: map[string]string{"isin": ""}, wantErr: apicontract.ErrMissingPathField},
		{name: "unknown field", method: "ListAccounts", fields: map[string]string{"account_id": "acc-1"}, wantErr: apicontract.ErrUnknownField},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := routes[tt.method].Target(tt.fields)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Target error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Target = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFill(t *testing.T) {
	t.Parallel()
	routes := readRoutes(t)
	tests := []struct {
		name, method string
		fields       map[string]string
		want         proto.Message
		wantErr      error
		anyErr       bool
	}{
		{
			name: "enum and date", method: "GetAllocation",
			fields: map[string]string{"date": "2026-03-13", "by": "ALLOCATION_DIMENSION_ASSET_CLASS"},
			want:   &viewsv1.GetAllocationRequest{Date: "2026-03-13", By: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_ASSET_CLASS},
		},
		{
			name: "integer as a string", method: "ListTransactions", fields: map[string]string{"accountId": "acc-1", "pageSize": "50"},
			want: &viewsv1.ListTransactionsRequest{AccountId: "acc-1", PageSize: 50},
		},
		{name: "empty", method: "ListSleeves", want: &viewsv1.ListSleevesRequest{}},
		{name: "unknown field", method: "ListSleeves", fields: map[string]string{"sleeve": "x"}, want: &viewsv1.ListSleevesRequest{}, wantErr: apicontract.ErrUnknownField},
		{name: "unknown enum name", method: "GetAllocation", fields: map[string]string{"by": "NOPE"}, want: &viewsv1.GetAllocationRequest{}, anyErr: true},
		{name: "another method's input", method: "ListAccounts", want: &viewsv1.ListSleevesRequest{}, wantErr: apicontract.ErrWrongMessage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.want.ProtoReflect().New().Interface()
			err := routes[tt.method].Fill(got, tt.fields)
			switch {
			case tt.anyErr:
				if err == nil {
					t.Fatalf("Fill = %v, want an error", got)
				}
				return
			case !errors.Is(err, tt.wantErr):
				t.Fatalf("Fill error = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("Fill (-want +got):\n%s", diff)
			}
		})
	}
}

// syntheticMethod is one method of the synthetic service: its name, its
// options in protobuf text format and whether it streams requests.
type syntheticMethod struct {
	name, options string
	streaming     bool
}

// syntheticRoutes builds a service openfief.e2etest.v1.S with methods,
// each taking a message with a string field account_id and a message
// field inner, and returns Routes of it.
func syntheticRoutes(t *testing.T, methods ...syntheticMethod) ([]*apicontract.Route, error) {
	t.Helper()
	var text strings.Builder
	text.WriteString(`name: "openfief/e2etest/v1/s.proto" package: "openfief.e2etest.v1" syntax: "proto3"
message_type {
  name: "In"
  field { name: "account_id" number: 1 type: TYPE_STRING label: LABEL_OPTIONAL json_name: "accountId" }
  field { name: "inner" number: 2 type: TYPE_MESSAGE label: LABEL_OPTIONAL type_name: ".openfief.e2etest.v1.In" json_name: "inner" }
}
service { name: "S"`)
	for _, m := range methods {
		fmt.Fprintf(&text, `
  method { name: %q input_type: ".openfief.e2etest.v1.In" output_type: ".openfief.e2etest.v1.In" client_streaming: %t options { %s } }`,
			m.name, m.streaming, m.options)
	}
	var file descriptorpb.FileDescriptorProto
	text.WriteString("\n}")
	if err := prototext.Unmarshal([]byte(text.String()), &file); err != nil {
		t.Fatalf("parsing the synthetic service: %v", err)
	}
	fd, err := protodesc.NewFile(&file, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatalf("building the synthetic service: %v", err)
	}
	routes, err := apicontract.Routes(fd.Services().Get(0))
	if err != nil {
		return nil, fmt.Errorf("routes of the synthetic service: %w", err)
	}
	return routes, nil
}

func TestRoutesOfOtherServices(t *testing.T) {
	t.Parallel()
	const noSideEffects = "idempotency_level: NO_SIDE_EFFECTS "
	tests := []struct {
		name         string
		method       syntheticMethod
		want         apicontract.Operation
		wantReadOnly bool
		wantErr      error
	}{
		{
			name:   "read-only GET",
			method: syntheticMethod{name: "Get", options: noSideEffects + `[google.api.http] { get: "/v1/a/{account_id}" }`},
			want:   apicontract.Operation{Method: http.MethodGet, Path: "/v1/a/{accountId}", ID: "openfief.e2etest.v1.S.Get"}, wantReadOnly: true,
		},
		{
			name:   "GET without the option",
			method: syntheticMethod{name: "Get", options: `[google.api.http] { get: "/v1/a" }`},
			want:   apicontract.Operation{Method: http.MethodGet, Path: "/v1/a", ID: "openfief.e2etest.v1.S.Get"},
		},
		{
			name:   "POST",
			method: syntheticMethod{name: "Put", options: `[google.api.http] { post: "/v1/a" body: "*" }`},
			want:   apicontract.Operation{Method: http.MethodPost, Path: "/v1/a", ID: "openfief.e2etest.v1.S.Put"},
		},
		{
			name:   "DELETE",
			method: syntheticMethod{name: "Del", options: `[google.api.http] { delete: "/v1/a" }`},
			want:   apicontract.Operation{Method: http.MethodDelete, Path: "/v1/a", ID: "openfief.e2etest.v1.S.Del"},
		},
		{
			name:   "PUT",
			method: syntheticMethod{name: "Set", options: `[google.api.http] { put: "/v1/a" body: "*" }`},
			want:   apicontract.Operation{Method: http.MethodPut, Path: "/v1/a", ID: "openfief.e2etest.v1.S.Set"},
		},
		{
			name:   "PATCH",
			method: syntheticMethod{name: "Fix", options: `[google.api.http] { patch: "/v1/a" body: "*" }`},
			want:   apicontract.Operation{Method: http.MethodPatch, Path: "/v1/a", ID: "openfief.e2etest.v1.S.Fix"},
		},
		{
			name:   "streaming",
			method: syntheticMethod{name: "Up", options: noSideEffects + `[google.api.http] { get: "/v1/a" }`, streaming: true},
			want:   apicontract.Operation{Method: http.MethodGet, Path: "/v1/a", ID: "openfief.e2etest.v1.S.Up"},
		},
		{name: "no rule", method: syntheticMethod{name: "Get", options: noSideEffects}, wantErr: apicontract.ErrNoHTTPRule},
		{
			name:    "custom verb",
			method:  syntheticMethod{name: "Get", options: `[google.api.http] { custom { kind: "HEAD" path: "/v1/a" } }`},
			wantErr: apicontract.ErrUnsupportedRule,
		},
		{name: "unknown variable", method: syntheticMethod{name: "Get", options: `[google.api.http] { get: "/v1/a/{nope}" }`}, wantErr: apicontract.ErrUnsupportedRule},
		{name: "nested variable", method: syntheticMethod{name: "Get", options: `[google.api.http] { get: "/v1/a/{inner.account_id}" }`}, wantErr: apicontract.ErrUnsupportedRule},
		{name: "unclosed variable", method: syntheticMethod{name: "Get", options: `[google.api.http] { get: "/v1/a/{account_id" }`}, wantErr: apicontract.ErrUnsupportedRule},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			routes, err := syntheticRoutes(t, tt.method)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Routes error = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if routes[0].Operation != tt.want || routes[0].ReadOnly != tt.wantReadOnly {
				t.Errorf("route = %+v read-only %t, want %+v read-only %t", routes[0].Operation, routes[0].ReadOnly, tt.want, tt.wantReadOnly)
			}
			if routes[0].Procedure != "/openfief.e2etest.v1.S/"+tt.method.name {
				t.Errorf("Procedure = %q", routes[0].Procedure)
			}
		})
	}
}

// envelope is msg in a Connect streaming envelope with flags.
func envelope(t *testing.T, flags byte, msg string) []byte {
	t.Helper()
	b, err := apicontract.Envelope([]byte(msg))
	if err != nil {
		t.Fatalf("Envelope: %v", err)
	}
	b[0] = flags
	return b
}

func TestConnectErrorCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, contentType string
		body              []byte
		want              string
		wantErr           error
	}{
		{name: "unary", contentType: "application/json", body: []byte(`{"code":"permission_denied","message":"caller refused"}`), want: "permission_denied"},
		{name: "unary without a code", contentType: "application/json", body: []byte(`{"accounts":[]}`), wantErr: apicontract.ErrNoConnectError},
		{name: "unary not JSON", contentType: "text/plain", body: []byte("404 page not found"), wantErr: apicontract.ErrNoConnectError},
		{
			name: "stream ends with an error", contentType: "application/connect+json",
			body: envelope(t, 0x02, `{"error":{"code":"permission_denied","message":"caller refused"}}`), want: "permission_denied",
		},
		{
			name: "stream with a message first", contentType: "application/connect+json",
			body: slices.Concat(envelope(t, 0, `{"ok":true}`), envelope(t, 0x02, `{"error":{"code":"unauthenticated"}}`)), want: "unauthenticated",
		},
		{name: "stream ends without an error", contentType: "application/connect+json", body: envelope(t, 0x02, `{}`), want: ""},
		{name: "stream without an end", contentType: "application/connect+proto", body: envelope(t, 0, `x`), wantErr: apicontract.ErrNoConnectError},
		{name: "truncated stream", contentType: "application/connect+json", body: envelope(t, 0x02, `{"error":{}}`)[:8], wantErr: apicontract.ErrNoConnectError},
		{name: "end of stream not JSON", contentType: "application/connect+json", body: envelope(t, 0x02, `nope`), wantErr: apicontract.ErrNoConnectError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := apicontract.ConnectErrorCode(tt.contentType, tt.body)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ConnectErrorCode error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ConnectErrorCode = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnvelope(t *testing.T) {
	t.Parallel()
	got, err := apicontract.Envelope([]byte("{}"))
	if err != nil {
		t.Fatalf("Envelope: %v", err)
	}
	if want := []byte{0, 0, 0, 0, 2, '{', '}'}; !cmp.Equal(got, want) {
		t.Errorf("Envelope = %v, want %v", got, want)
	}
}

func TestRESTStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, body string
		want       apicontract.Status
		wantErr    error
	}{
		{"not found", `{"code":5,"message":"no such account","details":[]}`, apicontract.Status{Code: 5, Message: "no such account", Details: []json.RawMessage{}}, nil},
		{"with a detail", `{"code":3,"message":"bad","details":[{"@type":"type.googleapis.com/google.rpc.BadRequest"}]}`, apicontract.Status{
			Code: 3, Message: "bad", Details: []json.RawMessage{json.RawMessage(`{"@type":"type.googleapis.com/google.rpc.BadRequest"}`)},
		}, nil},
		{"Connect's string code", `{"code":"not_found","message":"no such account"}`, apicontract.Status{}, apicontract.ErrNoStatus},
		{"no details", `{"code":5,"message":"no such account"}`, apicontract.Status{}, apicontract.ErrNoStatus},
		{"null details", `{"code":5,"message":"no such account","details":null}`, apicontract.Status{}, apicontract.ErrNoStatus},
		{"no message", `{"code":5,"details":[]}`, apicontract.Status{}, apicontract.ErrNoStatus},
		{"another field", `{"code":5,"message":"m","details":[],"error":"x"}`, apicontract.Status{}, apicontract.ErrNoStatus},
		{"not JSON", `404 page not found`, apicontract.Status{}, apicontract.ErrNoStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := apicontract.RESTStatus([]byte(tt.body))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RESTStatus(%s) error = %v, want %v", tt.body, err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("RESTStatus(%s) (-want +got):\n%s", tt.body, diff)
			}
		})
	}
}

// declaredErrors returns the responses, by status, r's operation must
// declare.
func declaredErrors(r *apicontract.Route) map[int]string {
	want := map[int]string{http.StatusBadRequest: "BadRequest", http.StatusUnauthorized: "Unauthorized"}
	if strings.Contains(r.Operation.Path, "{") {
		want[http.StatusNotFound] = "NotFound"
	}
	if r.Name == "GetModelPortfolio" {
		want[http.StatusNotFound] = "ModelPortfolioOff"
	}
	return want
}

// TestErrorResponse walks the generated document's error responses: every
// operation declares 400 and 401, an operation with a path parameter 404,
// GetModelPortfolio its own 404, and nothing else outside the default.
func TestErrorResponse(t *testing.T) {
	t.Parallel()
	doc, err := openapi3.NewLoader().LoadFromFile(specPath)
	if err != nil {
		t.Fatalf("loading %s: %v", specPath, err)
	}
	routes := readRoutes(t)
	for _, r := range readServiceRoutes(t) {
		want := declaredErrors(r)
		for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError} {
			var wantErr error
			if want[status] == "" {
				wantErr = apicontract.ErrUndeclaredStatus
			}
			if got, err := apicontract.ErrorResponse(doc, r.Operation, status); !errors.Is(err, wantErr) || got != want[status] {
				t.Errorf("ErrorResponse(%s, %d) = %q, %v; want %q, %v", r.Name, status, got, err, want[status], wantErr)
			}
		}
	}
	gone := apicontract.Operation{Method: http.MethodGet, Path: "/v1/nothing-here"}
	if _, err := apicontract.ErrorResponse(doc, gone, http.StatusNotFound); !errors.Is(err, apicontract.ErrNoOperation) {
		t.Errorf("ErrorResponse of an operation the document lacks = %v, want %v", err, apicontract.ErrNoOperation)
	}
	post := apicontract.Operation{Method: http.MethodPost, Path: routes["ListAccounts"].Operation.Path}
	if _, err := apicontract.ErrorResponse(doc, post, http.StatusBadRequest); !errors.Is(err, apicontract.ErrNoOperation) {
		t.Errorf("ErrorResponse of POST %s = %v, want %v", post.Path, err, apicontract.ErrNoOperation)
	}
}
