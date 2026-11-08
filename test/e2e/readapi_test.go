//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	connectrpc "connectrpc.com/connect"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/testing/protocmp"

	// Every contract's Go package, so that each protobuf file under proto/
	// is registered and its services probed (checkOtherServices).
	_ "github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1"
	_ "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	_ "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	_ "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	_ "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	_ "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	_ "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
	"github.com/arkadiuss/openfief/test/e2e/apicontract"
)

const (
	// readAPISpec is the read API's OpenAPI document as this checkout
	// publishes it, generated from its protobuf (D-50); the core doesn't
	// serve it.
	readAPISpec = "../../gen/openapi/openfief/views/v1/read_service.openapi.yaml"
	// protoDir holds the checkout's contracts.
	protoDir = "../../proto/openfief"
	// connectorPackage is the connector contract's package, which plugins
	// serve and the core calls (D-49): the core serves none of it.
	connectorPackage = "openfief.connector.v1"
	// corePluginPort is the core's plugin port in its pod (D-115).
	corePluginPort = 8081
	// readAPIAttempts is how often one call over Connect and REST is made
	// until both answer the same: the record may change between the two,
	// as the scheduler syncs and other tests install plugins.
	readAPIAttempts = 5
	// readAPIBodyLimit caps one read API response the suite reads.
	readAPIBodyLimit = 8 << 20
	// unissued is a bearer token the core never issued.
	unissued = "e2e-never-issued"
	// The gRPC codes of the google.rpc.Status bodies of the read API's
	// declared errors over REST.
	invalidArgument = 3
	notFound        = 5
	unauthenticated = 16
	// accountMethods and instrumentMethods are how many calls
	// accountCalls makes per account and instrumentCalls per ISIN.
	accountMethods    = 9
	instrumentMethods = 5
	// permissionDenied is Connect's code for a refused caller.
	permissionDenied = "permission_denied"
)

var (
	// errWrongRequest reports a request message of another type than the
	// SDK method takes.
	errWrongRequest = errors.New("request of the wrong type")
	// errWrongResponse reports an SDK response that is no proto message.
	errWrongResponse = errors.New("response is not a proto message")
)

// readAPI is what the read API checks share: the OpenAPI document and a
// router over it, ReadService's descriptor and routes by name in their
// order, the SDK's Connect client, the base URL of a port-forward to the
// core's browser port and the script token.
type readAPI struct {
	doc     *openapi3.T
	router  routers.Router
	service protoreflect.ServiceDescriptor
	routes  map[string]*apicontract.Route
	order   []string
	client  viewsv1connect.ReadServiceClient
	base    string
	token   string
}

// readCall is one call of a ReadService method with its request fields
// by JSON name; off marks the model portfolio, which answers NOT_FOUND
// while it is off (D-52), as it is in the suite.
type readCall struct {
	rpc    string
	fields map[string]string
	off    bool
}

// String names c in a failure.
func (c readCall) String() string {
	keys := make([]string, 0, len(c.fields))
	for k := range c.fields {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+c.fields[k])
	}
	return c.rpc + "(" + strings.Join(parts, ", ") + ")"
}

// restReply is an HTTP answer the suite read.
type restReply struct {
	status int
	header http.Header
	body   []byte
}

// checkReadAPI checks D-200's criterion 7 on the data path's record,
// through a port-forward to the core's browser port, as a script calls it
// with a script token, and to its plugin port:
//
//   - every resource (SCEN-e2e-read-api-every-resource): the OpenAPI
//     document's operations are exactly ReadService's HTTP routes, the
//     SDK's client has a method for each, and accounts, transactions, tax
//     lots, holdings and valuations of the data path's accounts are there;
//   - the OpenAPI document (SCEN-e2e-read-api-openapi-conformance): every
//     operation, called over REST with the data path's account IDs, ISINs
//     and connector instance, answers 200 with a body the document's
//     schema accepts and that decodes strictly into the response message,
//     and 401 without a token or with one the core never issued;
//   - error responses (SCEN-views-read-api-declared-errors): each 401, the
//     model portfolio's 404 while it is off, a 404 for an account ID no
//     account has and a 400 for a date range that ends before it starts
//     are the responses the operation declares, with a google.rpc.Status
//     body the document's schema accepts and the declared code;
//   - no write (SCEN-e2e-read-api-no-write): the document has only GET
//     operations without a body, every ReadService method is unary with no
//     side effects, every other HTTP method on every path answers 405 with
//     Allow: GET, and every method of every other service under proto/,
//     the owner services (D-79) among them, refuses the script token: 404
//     on the browser port, permission_denied on the plugin port, 404 for
//     the connector contract the core doesn't serve;
//   - Connect equals REST (SCEN-e2e-read-api-connect-equals-rest): each
//     call through the SDK's generated Connect client decodes to the same
//     message as the same call over REST.
func (p *dataPath) checkReadAPI(t *testing.T) {
	t.Helper()
	api := p.readAPI(t)
	calls := p.readCalls(t)
	api.checkEveryResource(t, calls)
	api.checkConformance(t, calls)
	api.checkRefusesTokens(t, calls)
	api.checkErrors(t)
	api.checkOnlyGET(t, calls)
	pod, err := p.k.corePod(t.Context())
	if err != nil {
		t.Fatalf("finding the core's pod: %v", err)
	}
	plugin := "http://" + p.k.forwardPod(t, namespace, podPort{pod: pod, port: corePluginPort})
	api.checkOtherServices(t, plugin)
	api.checkConnectEqualsREST(t, calls)
}

// readAPI loads the OpenAPI document and ReadService's routes and opens a
// port-forward to the core for the SDK's client.
func (p *dataPath) readAPI(t *testing.T) *readAPI {
	t.Helper()
	doc, err := openapi3.NewLoader().LoadFromFile(readAPISpec)
	if err != nil {
		t.Fatalf("loading %s: %v", readAPISpec, err)
	}
	if invalid := doc.Validate(t.Context()); invalid != nil {
		t.Fatalf("%s is not a valid OpenAPI document: %v", readAPISpec, invalid)
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatalf("routing %s: %v", readAPISpec, err)
	}
	api := &readAPI{
		doc: doc, router: router, service: viewsv1.File_openfief_views_v1_read_service_proto.Services().ByName("ReadService"),
		routes: map[string]*apicontract.Route{}, base: forwardCore(t, p.k), token: p.token,
	}
	routes, err := apicontract.Routes(api.service)
	if err != nil {
		t.Fatalf("ReadService's routes: %v", err)
	}
	for _, r := range routes {
		api.routes[r.Name] = r
		api.order = append(api.order, r.Name)
	}
	api.client = viewsv1connect.NewReadServiceClient(http.DefaultClient, api.base,
		connectrpc.WithInterceptors(withBearer(p.token)))
	return api
}

// withBearer sends token as the bearer token of every call.
func withBearer(token string) connectrpc.UnaryInterceptorFunc {
	return func(next connectrpc.UnaryFunc) connectrpc.UnaryFunc {
		return func(ctx context.Context, req connectrpc.AnyRequest) (connectrpc.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}

// readCalls lists the calls the checks make: every method without an
// account or instrument, each account-scoped method on each data path
// account, each instrument method on each of its ISINs, and the
// allocation by every dimension and value kind. Dates are the IBKR week's.
func (p *dataPath) readCalls(t *testing.T) []readCall {
	t.Helper()
	week := map[string]string{"from": goLive, "to": weekEnd}
	global := []readCall{
		{rpc: "ListSleeves"},
		{rpc: "ListSleeveKinds"},
		{rpc: "GetSettings"},
		{rpc: "ListYearlyTaxes", fields: map[string]string{"year": goLive[:4]}},
		{rpc: "ListAccounts"},
		{rpc: "ListConnectorBindings"},
		{rpc: "ListValuationTotals", fields: week},
		{rpc: "ListValuationStatuses"},
		{rpc: "ListPriceChecks"},
		{rpc: "GetPerformance"},
		{rpc: "GetPerformance", fields: week},
		{rpc: "ListTransfers"},
		{rpc: "GetModelPortfolio", off: true},
		{rpc: "ListPlugins"},
		{rpc: "ListAttentionItems"},
		{rpc: "ListActivity"},
		{rpc: "ListSyncStatuses"},
		{rpc: "ListSyncAttempts", fields: map[string]string{"connectorInstance": dpIBKR}},
	}
	return slices.Concat(global, accountCalls(p.ibkr, p.xtb, p.bonds), instrumentCalls(isinG, isinH, isinJ), allocationCalls(t))
}

// accountCalls calls each account-scoped method on each of accounts.
func accountCalls(accounts ...string) []readCall {
	calls := make([]readCall, 0, accountMethods*len(accounts))
	for _, account := range accounts {
		a := map[string]string{"accountId": account}
		on := map[string]string{"accountId": account, "date": weekEnd}
		week := map[string]string{"accountId": account, "from": goLive, "to": weekEnd}
		calls = append(calls,
			readCall{rpc: "GetAccount", fields: a}, readCall{rpc: "ListTransactions", fields: a},
			readCall{rpc: "ListHoldings", fields: on}, readCall{rpc: "ListLots", fields: on}, readCall{rpc: "ListCash", fields: on},
			readCall{rpc: "ListAssertionChecks", fields: a}, readCall{rpc: "ListAccountValuations", fields: week},
			readCall{rpc: "GetValuationStatus", fields: a}, readCall{rpc: "GetPerformance", fields: week})
	}
	return calls
}

// instrumentCalls calls each instrument method on each of isins.
func instrumentCalls(isins ...string) []readCall {
	calls := make([]readCall, 0, instrumentMethods*len(isins))
	for _, isin := range isins {
		i := map[string]string{"isin": isin}
		calls = append(calls,
			readCall{rpc: "GetInstrument", fields: i}, readCall{rpc: "ListInstrumentTransactions", fields: i},
			readCall{rpc: "ListInstrumentPrices", fields: map[string]string{"isin": isin, "from": goLive, "to": weekEnd}},
			readCall{rpc: "ListInstrumentHoldings", fields: map[string]string{"isin": isin, "date": weekEnd}},
			readCall{rpc: "ListPriceChecks", fields: i})
	}
	return calls
}

// allocationCalls asks for the allocation at the week's end by every
// dimension, and by asset class in every value kind.
func allocationCalls(t *testing.T) []readCall {
	t.Helper()
	dimensions, values := enumNames(t, viewsv1.AllocationDimension(0).Descriptor()), enumNames(t, viewsv1.ValueKind(0).Descriptor())
	calls := make([]readCall, 0, len(dimensions)+len(values))
	for _, by := range dimensions {
		calls = append(calls, readCall{rpc: "GetAllocation", fields: map[string]string{"date": weekEnd, "by": by}})
	}
	for _, value := range values {
		calls = append(calls, readCall{rpc: "GetAllocation", fields: map[string]string{
			"date": weekEnd, "by": viewsv1.AllocationDimension_ALLOCATION_DIMENSION_ASSET_CLASS.String(), "value": value,
		}})
	}
	return calls
}

// enumNames returns e's value names other than its zero value.
func enumNames(t *testing.T, e protoreflect.EnumDescriptor) []string {
	t.Helper()
	var names []string
	for i := range e.Values().Len() {
		if v := e.Values().Get(i); v.Number() != 0 {
			names = append(names, string(v.Name()))
		}
	}
	if len(names) == 0 {
		t.Fatalf("%s has no values", e.FullName())
	}
	return names
}

// checkEveryResource checks that the OpenAPI document's operations are
// exactly ReadService's routes, and that the SDK's client and the calls
// cover every method.
func (a *readAPI) checkEveryResource(t *testing.T, calls []readCall) {
	t.Helper()
	routes := make([]*apicontract.Route, 0, len(a.order))
	for _, name := range a.order {
		routes = append(routes, a.routes[name])
	}
	if diff := cmp.Diff(apicontract.RouteOperations(routes), apicontract.Operations(a.doc)); diff != "" {
		t.Errorf("the OpenAPI document's operations differ from ReadService's routes (-routes +document):\n%s", diff)
	}
	called := map[string]bool{}
	for _, c := range calls {
		if a.routes[c.rpc] == nil {
			t.Fatalf("%s: ReadService has no such method", c)
		}
		called[c.rpc] = true
	}
	sdk := sdkMethods(a.client)
	for _, name := range a.order {
		if !called[name] {
			t.Errorf("no call of %s: add one to readCalls", name)
		}
		if sdk[name] == nil {
			t.Errorf("sdkMethods has no %s of the SDK's ReadServiceClient", name)
		}
	}
	if len(sdk) != len(a.routes) {
		t.Errorf("sdkMethods has %d methods, ReadService %d", len(sdk), len(a.routes))
	}
}

// checkConformance calls each of calls over REST: it answers 200 with a
// body the OpenAPI document accepts for that operation and that decodes
// strictly into the response message, and D-200's resources hold the
// data path's record; the model portfolio, off, answers 404.
func (a *readAPI) checkConformance(t *testing.T, calls []readCall) {
	t.Helper()
	for _, c := range calls {
		target := a.target(t, c)
		r := a.rest(t, http.MethodGet, target, a.token)
		switch {
		case c.off:
			a.checkError(t, c, target, r, declaredError{http.StatusNotFound, "ModelPortfolioOff", notFound})
		case r.status != http.StatusOK:
			t.Errorf("GET %s: status %d, body %q; want 200", target, r.status, excerpt(string(r.body)))
		default:
			a.validate(t, target, r)
			m, err := a.decodeREST(c, r.body)
			if err != nil {
				t.Errorf("GET %s: %v", target, err)
				continue
			}
			checkHoldsData(t, c, m.msg)
		}
	}
}

// checkRefusesTokens checks that every operation answers 401 without a
// token and with one the core never issued, as the operation declares it.
func (a *readAPI) checkRefusesTokens(t *testing.T, calls []readCall) {
	t.Helper()
	for _, c := range callPerMethod(calls) {
		target := a.target(t, c)
		for _, token := range []string{"", unissued} {
			r := a.rest(t, http.MethodGet, target, token)
			a.checkError(t, c, target, r, declaredError{http.StatusUnauthorized, "Unauthorized", unauthenticated})
		}
	}
}

// checkErrors checks a 404 for an account ID no account has and a 400 for
// a date range that ends before it starts, each as its operation declares
// it.
func (a *readAPI) checkErrors(t *testing.T) {
	t.Helper()
	for _, e := range []struct {
		call readCall
		want declaredError
	}{
		{
			readCall{rpc: "GetAccount", fields: map[string]string{"accountId": "e2e-no-such-account"}},
			declaredError{http.StatusNotFound, "NotFound", notFound},
		},
		{
			readCall{rpc: "ListValuationTotals", fields: map[string]string{"from": weekEnd, "to": goLive}},
			declaredError{http.StatusBadRequest, "BadRequest", invalidArgument},
		},
	} {
		target := a.target(t, e.call)
		a.checkError(t, e.call, target, a.rest(t, http.MethodGet, target, a.token), e.want)
	}
}

// declaredError is an error response the OpenAPI document declares: its
// HTTP status, its name in components.responses and the gRPC code of its
// google.rpc.Status body.
type declaredError struct {
	status   int
	response string
	code     int32
}

// checkError checks r, the answer to GET target, a call of c: it has
// want's status, which c's operation declares as want's response, and a
// body the document's schema accepts that decodes strictly into a
// google.rpc.Status with want's code.
func (a *readAPI) checkError(t *testing.T, c readCall, target string, r restReply, want declaredError) {
	t.Helper()
	if r.status != want.status {
		t.Errorf("GET %s: status %d, body %q; want %d", target, r.status, excerpt(string(r.body)), want.status)
		return
	}
	if got, err := apicontract.ErrorResponse(a.doc, a.routes[c.rpc].Operation, want.status); err != nil || got != want.response {
		t.Errorf("GET %s: the document declares %d as %q (%v); want %q", target, want.status, got, err, want.response)
	}
	a.validate(t, target, r)
	s, err := apicontract.RESTStatus(r.body)
	if err != nil {
		t.Errorf("GET %s: %v", target, err)
		return
	}
	if s.Code != want.code {
		t.Errorf("GET %s: code %d (%s); want %d", target, s.Code, s.Message, want.code)
	}
}

// callPerMethod returns the first of calls of each method.
func callPerMethod(calls []readCall) []readCall {
	seen := map[string]bool{}
	var first []readCall
	for _, c := range calls {
		if !seen[c.rpc] {
			seen[c.rpc] = true
			first = append(first, c)
		}
	}
	return first
}

// targetPerMethod returns the REST target of the first of calls of each
// method.
func (a *readAPI) targetPerMethod(t *testing.T, calls []readCall) []string {
	t.Helper()
	first := callPerMethod(calls)
	targets := make([]string, 0, len(first))
	for _, c := range first {
		targets = append(targets, a.target(t, c))
	}
	return targets
}

// validate checks a response to GET target against the OpenAPI document.
func (a *readAPI) validate(t *testing.T, target string, r restReply) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, a.base+target, http.NoBody)
	if err != nil {
		t.Fatalf("GET %s: %v", target, err)
	}
	route, params, err := a.router.FindRoute(req)
	if err != nil {
		t.Errorf("GET %s is not in the OpenAPI document: %v", target, err)
		return
	}
	in := &openapi3filter.RequestValidationInput{
		Request: req, PathParams: params, Route: route,
		Options: &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
	}
	if err := openapi3filter.ValidateRequest(t.Context(), in); err != nil {
		t.Errorf("GET %s: the request doesn't match the OpenAPI document: %v", target, err)
	}
	out := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: in, Status: r.status, Header: r.header, Body: io.NopCloser(bytes.NewReader(r.body)),
		Options: &openapi3filter.Options{IncludeResponseStatus: true, MultiError: true},
	}
	if err := openapi3filter.ValidateResponse(t.Context(), out); err != nil {
		t.Errorf("GET %s: the response doesn't match the OpenAPI document: %v", target, err)
	}
}

// checkHoldsData checks that m, c's answer, holds the data path's record
// where c is one of the resources D-200's criterion 7 names: accounts are
// listed, and each data path account has transactions, lots and holdings
// at the week's end and valuations in it, as the totals do.
func checkHoldsData(t *testing.T, c readCall, m proto.Message) {
	t.Helper()
	var n int
	switch r := m.(type) {
	case *viewsv1.ListAccountsResponse:
		n = len(r.GetAccounts())
	case *viewsv1.ListTransactionsResponse:
		n = len(r.GetTransactions())
	case *viewsv1.ListLotsResponse:
		n = len(r.GetLots())
	case *viewsv1.ListHoldingsResponse:
		n = len(r.GetHoldings())
	case *viewsv1.ListAccountValuationsResponse:
		n = len(r.GetValuations())
	case *viewsv1.ListValuationTotalsResponse:
		n = len(r.GetTotals())
	default:
		return
	}
	if n == 0 {
		t.Errorf("%s: nothing listed; the data path booked some", c)
	}
}

// checkOnlyGET checks that nothing in the read API writes: the OpenAPI
// document has only GET operations without a body, each ReadService
// method is a unary GET with no side effects, and every other method on
// each operation's path answers 405 with Allow: GET, with the script
// token and without one.
func (a *readAPI) checkOnlyGET(t *testing.T, calls []readCall) {
	t.Helper()
	for path, item := range a.doc.Paths.Map() {
		for method, op := range item.Operations() {
			if method != http.MethodGet || op.RequestBody != nil {
				t.Errorf("%s %s: the read API has only GET operations without a body (D-6, D-7)", method, path)
			}
		}
	}
	for _, name := range a.order {
		if r := a.routes[name]; !r.ReadOnly || r.Operation.Method != http.MethodGet {
			t.Errorf("%s: read-only %t, %s; want a unary GET with no side effects", name, r.ReadOnly, r.Operation.Method)
		}
	}
	for _, target := range a.targetPerMethod(t, calls) {
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
			a.checkNotAllowed(t, method, target)
		}
	}
}

// checkNotAllowed checks that method target answers 405 with Allow: GET,
// with the script token and without one.
func (a *readAPI) checkNotAllowed(t *testing.T, method, target string) {
	t.Helper()
	for _, token := range []string{a.token, ""} {
		r := a.rest(t, method, target, token)
		if r.status != http.StatusMethodNotAllowed || r.header.Get("Allow") != http.MethodGet {
			t.Errorf("%s %s (with the token: %t): status %d, Allow %q; want 405, Allow GET",
				method, target, token != "", r.status, r.header.Get("Allow"))
		}
	}
}

// checkOtherServices calls every method of every service under proto/
// other than ReadService over Connect with the script token, on the
// browser port and on the plugin port at plugin: none succeeds. The
// browser port serves none of them (404); the plugin port refuses the
// script token with permission_denied, since only a view plugin acting
// for an owner session calls the owner services (D-79, D-115); the
// connector contract, which plugins serve, is 404 on both.
func (a *readAPI) checkOtherServices(t *testing.T, plugin string) {
	t.Helper()
	probed := 0
	for _, sd := range contractServices(t) {
		if sd.FullName() == a.service.FullName() {
			continue
		}
		want := permissionDenied
		if sd.ParentFile().Package() == connectorPackage {
			want = ""
		}
		for i := range sd.Methods().Len() {
			m := sd.Methods().Get(i)
			probed++
			a.checkRefused(t, a.base, m, "")
			a.checkRefused(t, plugin, m, want)
		}
	}
	if probed == 0 {
		t.Error("no service besides ReadService found under proto/")
	}
}

// checkRefused calls m at base with the script token and an empty
// message: a unary method as Connect's unary JSON, a streaming one as one
// enveloped message. It wants the Connect code want, or 404 when want is
// empty.
func (a *readAPI) checkRefused(t *testing.T, base string, m protoreflect.MethodDescriptor, want string) {
	t.Helper()
	contentType, body := "application/json", []byte("{}")
	if m.IsStreamingClient() || m.IsStreamingServer() {
		var err error
		if body, err = apicontract.Envelope(body); err != nil {
			t.Fatal(err)
		}
		contentType = "application/connect+json"
	}
	url := base + apicontract.Procedure(m)
	r := send(t, http.MethodPost, url, a.token, contentType, body)
	if want == "" {
		if r.status != http.StatusNotFound {
			t.Errorf("POST %s with the script token: status %d, body %q; want 404", url, r.status, excerpt(string(r.body)))
		}
		return
	}
	code, err := apicontract.ConnectErrorCode(r.header.Get("Content-Type"), r.body)
	if err != nil || code != want {
		t.Errorf("POST %s with the script token: status %d, code %q (%v); want %s", url, r.status, code, err, want)
	}
}

// contractServices returns the services of every protobuf file under
// proto/, failing t for a file no Go package registers.
func contractServices(t *testing.T) []protoreflect.ServiceDescriptor {
	t.Helper()
	root := filepath.Dir(protoDir)
	var services []protoreflect.ServiceDescriptor
	err := filepath.WalkDir(protoDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".proto" {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		fd, err := protoregistry.GlobalFiles.FindFileByPath(filepath.ToSlash(rel))
		if err != nil {
			t.Errorf("%s is not registered: import its Go package in readapi_test.go: %v", rel, err)
			return nil
		}
		for i := range fd.Services().Len() {
			services = append(services, fd.Services().Get(i))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", protoDir, err)
	}
	return services
}

// checkConnectEqualsREST makes each call through the SDK's Connect client
// and over REST; both decode to the same message. A pair that differs is
// made again, up to readAPIAttempts times, since the record may change
// between the two. The model portfolio, off, is NOT_FOUND over Connect
// and 404 over REST.
func (a *readAPI) checkConnectEqualsREST(t *testing.T, calls []readCall) {
	t.Helper()
	sdk := sdkMethods(a.client)
	for _, c := range calls {
		invoke := sdk[c.rpc]
		if invoke == nil {
			continue
		}
		fill := func(m proto.Message) error { return a.routes[c.rpc].Fill(m, c.fields) }
		var diff string
		for range readAPIAttempts {
			if diff = a.compare(t, c, invoke, fill); diff == "" {
				break
			}
		}
		if diff != "" {
			t.Errorf("%s: Connect and REST differ in each of %d attempts; the last (-REST +Connect):\n%s", c, readAPIAttempts, diff)
		}
	}
}

// compare makes c over REST and Connect once and describes how they
// differ; "" is the same answer.
func (a *readAPI) compare(t *testing.T, c readCall, invoke sdkMethod, fill func(proto.Message) error) string {
	t.Helper()
	r := a.rest(t, http.MethodGet, a.target(t, c), a.token)
	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()
	got, err := invoke(ctx, fill)
	if c.off {
		if code := connectrpc.CodeOf(err); r.status != http.StatusNotFound || code != connectrpc.CodeNotFound {
			return fmt.Sprintf("REST status %d, Connect %s (%v); want 404 and not_found", r.status, code, err)
		}
		return ""
	}
	switch {
	case err != nil:
		return fmt.Sprintf("Connect failed: %v", err)
	case r.status != http.StatusOK:
		return fmt.Sprintf("REST status %d, body %q", r.status, excerpt(string(r.body)))
	}
	want, err := a.decodeREST(c, r.body)
	if err != nil {
		return err.Error()
	}
	return cmp.Diff(want.msg, got, protocmp.Transform())
}

// decoded is a REST body decoded into its response message.
type decoded struct {
	msg proto.Message
}

// decodeREST decodes a REST body of c strictly into c's response message:
// a field the message lacks is an error.
func (a *readAPI) decodeREST(c readCall, body []byte) (decoded, error) {
	md := a.service.Methods().ByName(protoreflect.Name(c.rpc))
	if md == nil {
		return decoded{}, fmt.Errorf("%w: ReadService has no %s", errWrongResponse, c.rpc)
	}
	mt, err := protoregistry.GlobalTypes.FindMessageByName(md.Output().FullName())
	if err != nil {
		return decoded{}, fmt.Errorf("%s's response type: %w", c.rpc, err)
	}
	m := mt.New().Interface()
	if err := protojson.Unmarshal(body, m); err != nil {
		return decoded{}, fmt.Errorf("decoding %s's REST body strictly: %w", c.rpc, err)
	}
	return decoded{msg: m}, nil
}

// target is c's REST path and query.
func (a *readAPI) target(t *testing.T, c readCall) string {
	t.Helper()
	target, err := a.routes[c.rpc].Target(c.fields)
	if err != nil {
		t.Fatalf("%s: %v", c, err)
	}
	return target
}

// rest sends method target to the core's browser port with token as the
// bearer token unless it is empty.
func (a *readAPI) rest(t *testing.T, method, target, token string) restReply {
	t.Helper()
	return send(t, method, a.base+target, token, "", nil)
}

// send sends method url with token as the bearer token unless it is
// empty, and body as contentType over Connect unless body is nil. The
// token never appears in a failure.
func send(t *testing.T, method, url, token, contentType string, body []byte) restReply {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()
	var reader io.Reader = http.NoBody
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Connect-Protocol-Version", "1")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("%s %s: closing the body: %v", method, url, closeErr)
		}
	}()
	reply, err := io.ReadAll(io.LimitReader(resp.Body, readAPIBodyLimit))
	if err != nil {
		t.Fatalf("%s %s: reading the body: %v", method, url, err)
	}
	return restReply{status: resp.StatusCode, header: resp.Header, body: reply}
}

// sdkMethod calls one method of the SDK's generated client with a request
// fill sets the fields of.
type sdkMethod func(ctx context.Context, fill func(proto.Message) error) (proto.Message, error)

// unary adapts one generated client method to sdkMethod.
func unary[Req, Res any](f func(context.Context, *connectrpc.Request[Req]) (*connectrpc.Response[Res], error)) sdkMethod {
	return func(ctx context.Context, fill func(proto.Message) error) (proto.Message, error) {
		req := new(Req)
		m, ok := any(req).(proto.Message)
		if !ok {
			return nil, fmt.Errorf("%w: %T", errWrongRequest, req)
		}
		if err := fill(m); err != nil {
			return nil, err
		}
		resp, err := f(ctx, connectrpc.NewRequest(req))
		if err != nil {
			return nil, fmt.Errorf("calling over Connect: %w", err)
		}
		out, ok := any(resp.Msg).(proto.Message)
		if !ok {
			return nil, fmt.Errorf("%w: %T", errWrongResponse, resp.Msg)
		}
		return out, nil
	}
}

// sdkMethods is every method of the SDK's generated ReadServiceClient by
// name; checkEveryResource fails when ReadService has one it lacks.
func sdkMethods(c viewsv1connect.ReadServiceClient) map[string]sdkMethod {
	return map[string]sdkMethod{
		"ListSleeves":                unary(c.ListSleeves),
		"ListSleeveKinds":            unary(c.ListSleeveKinds),
		"GetSettings":                unary(c.GetSettings),
		"ListYearlyTaxes":            unary(c.ListYearlyTaxes),
		"ListAccounts":               unary(c.ListAccounts),
		"ListConnectorBindings":      unary(c.ListConnectorBindings),
		"GetAccount":                 unary(c.GetAccount),
		"ListTransactions":           unary(c.ListTransactions),
		"ListHoldings":               unary(c.ListHoldings),
		"ListLots":                   unary(c.ListLots),
		"ListCash":                   unary(c.ListCash),
		"ListAssertionChecks":        unary(c.ListAssertionChecks),
		"ListAccountValuations":      unary(c.ListAccountValuations),
		"GetValuationStatus":         unary(c.GetValuationStatus),
		"ListValuationTotals":        unary(c.ListValuationTotals),
		"ListValuationStatuses":      unary(c.ListValuationStatuses),
		"ListPriceChecks":            unary(c.ListPriceChecks),
		"GetAllocation":              unary(c.GetAllocation),
		"GetPerformance":             unary(c.GetPerformance),
		"ListTransfers":              unary(c.ListTransfers),
		"GetModelPortfolio":          unary(c.GetModelPortfolio),
		"ListPlugins":                unary(c.ListPlugins),
		"ListAttentionItems":         unary(c.ListAttentionItems),
		"ListActivity":               unary(c.ListActivity),
		"ListSyncStatuses":           unary(c.ListSyncStatuses),
		"ListSyncAttempts":           unary(c.ListSyncAttempts),
		"GetInstrument":              unary(c.GetInstrument),
		"ListInstrumentPrices":       unary(c.ListInstrumentPrices),
		"ListInstrumentHoldings":     unary(c.ListInstrumentHoldings),
		"ListInstrumentTransactions": unary(c.ListInstrumentTransactions),
	}
}
