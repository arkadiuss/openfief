package main

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/riverqueue/river/rivertype"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/arkadiuss/openfief/internal/ingest"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

// The synthetic data the traced requests carry (GR-12). None of it, nor
// anything shaped like it, may reach an exported span.
const (
	tracedAmount   = "1234.56"
	tracedRate     = "4.3217"
	tracedQuantity = "17.5"
	tracedAccount  = "U7654321"
	tracedISIN     = "XX00000SYNT3"
	tracedIBAN     = "PL61109010140000071219812874"
	tracedToken    = "synthetic-core-token"
	tracedSession  = "synthetic-session-id"
	tracedInstance = "synthetic-fx"
)

// errTracedRefusal is a connector's refusal that quotes data, as an error
// message may.
var errTracedRefusal = errors.New("account " + tracedAccount + " holds " + tracedQuantity + " of " + tracedISIN)

// deniedKeys are attribute key fragments no exported span may carry.
func deniedKeys() []string {
	return []string{
		"amount", "quantity", "price", "account", "isin", "iban", "token", "session", "cookie", "authorization",
		"password", "secret", "passkey", "credential", "query.text", "statement", "parameter", "args", "url", "path",
		"address", "user_agent", "header", "message", "stacktrace",
	}
}

// deniedValues match values no exported span may carry: the synthetic data
// itself and anything shaped like an amount, an ISIN, an IBAN or a JWT.
func deniedValues() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(regexp.QuoteMeta(tracedAmount)), regexp.MustCompile(regexp.QuoteMeta(tracedRate)),
		regexp.MustCompile(regexp.QuoteMeta(tracedQuantity)), regexp.MustCompile(tracedAccount),
		regexp.MustCompile(tracedToken), regexp.MustCompile(tracedSession),
		regexp.MustCompile(`\d+\.\d{2,}`),                      // an amount or a rate
		regexp.MustCompile(`\b[A-Z]{2}[A-Z0-9]{9}\d\b`),        // an ISIN
		regexp.MustCompile(`\b[A-Z]{2}\d{2}[A-Z0-9]{11,30}\b`), // an IBAN
		regexp.MustCompile(`eyJ[A-Za-z0-9_-]+`),                // a JWT
	}
}

// tracedTokens is the core's token for every instance.
type tracedTokens struct{}

// TokenFor returns the synthetic token.
func (tracedTokens) TokenFor(context.Context, string) (string, error) { return tracedToken, nil }

// Token returns the synthetic token, as a plugin's projected token.
func (tracedTokens) Token() (string, error) { return tracedToken, nil }

// tracedVerifier accepts the synthetic token as the core's.
type tracedVerifier struct{}

// VerifyCore accepts tracedToken alone.
func (tracedVerifier) VerifyCore(_ context.Context, token string) error {
	if token != tracedToken {
		return openfief.ErrNotCore
	}
	return nil
}

// tracedConnector answers a sync with a rate, and refuses the next one with
// a message that quotes data. It records the trace each call ran in.
type tracedConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	traces []trace.TraceID
}

// Sync answers the first call with a synthetic FX rate and refuses the rest.
func (c *tracedConnector) Sync(ctx context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	c.traces = append(c.traces, trace.SpanContextFromContext(ctx).TraceID())
	if len(c.traces) > 1 {
		return nil, refusal()
	}
	return connect.NewResponse(&connectorv1.SyncResponse{
		Delivery: &connectorv1.Delivery{
			BatchId: req.Msg.GetBatchId(), Institution: "nbp",
			FxRates: []*connectorv1.FxRate{{Currency: "EUR", Date: "2026-10-27", Rate: tracedRate}},
		},
		Watermark: "2026-10-27",
	}), nil
}

// refusal is the connector's answer to a second sync.
func refusal() *connect.Error { return connect.NewError(connect.CodeInvalidArgument, errTracedRefusal) }

// TestTracingKeepsDataOut drives representative traffic through the core's
// and the SDK's tracing into an in-memory exporter behind telemetry.Redact,
// exactly as an OTLP exporter sits behind it: a connector sync and a
// refused one from the core to a plugin, a plugin's and a browser's calls
// to the core's ports with tokens and session cookies, queries with
// arguments and a failure that quotes one, and a revaluation job enqueued
// and worked. It checks that the trace crosses from the core to the plugin
// and from the job's enqueueing to its work, that the browser can't choose
// the trace, and that no exported span carries any of the data. It is the
// only test in this package that sets OpenTelemetry's globals.
func TestTracingKeepsDataOut(t *testing.T) {
	t.Parallel()
	if err := withTracing(t.Context(), func() error { return nil }); err != nil {
		t.Fatalf("withTracing without an endpoint: %v", err)
	}
	if _, rec := otel.Tracer("test").Start(t.Context(), "off"); rec.IsRecording() {
		t.Error("with no endpoint, the core records spans")
	}
	mem := tracetest.NewInMemoryExporter()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSyncer(telemetry.Redact(mem))))
	otel.SetTextMapPropagator(propagation.TraceContext{})

	ctx, root := otel.Tracer("test").Start(t.Context(), "root")
	conn := syncThroughPlugin(ctx, t)
	forged := callCorePorts(ctx, t)
	runQueries(ctx, t)
	insert := runJob(ctx, t)
	root.End()

	spans := mem.GetSpans()
	for _, got := range conn.traces {
		if got != root.SpanContext().TraceID() {
			t.Errorf("the plugin ran in trace %s, want the core's %s", got, root.SpanContext().TraceID())
		}
	}
	checkLinkedWork(t, spans, insert)
	checkNewBrowserTrace(t, spans, forged)
	for _, name := range []string{"openfief.connector.v1.ConnectorService/Sync", "ingest.sync", "INSERT", "river.insert_many", "river.work/valuation.revalue_account"} {
		if !hasSpan(spans, name) {
			t.Errorf("no span %q was exported", name)
		}
	}
	checkPrivate(t, spans)
}

// syncThroughPlugin serves a tracedConnector with the SDK's connector
// server and syncs it twice with the core's connector client, within a
// span ingest.sync as the scheduler's would be.
func syncThroughPlugin(ctx context.Context, t *testing.T) *tracedConnector {
	t.Helper()
	conn := &tracedConnector{}
	server := openfief.NewConnectorServer("", conn, tracedVerifier{})
	srv := httptest.NewUnstartedServer(server.Handler)
	srv.Config.Protocols = server.Protocols
	srv.Start()
	t.Cleanup(srv.Close)
	client := connectrpc.NewClient(pluginClient(), func(string) (string, error) { return srv.URL, nil }, tracedTokens{})
	call := &ingestapp.SyncCall{ID: "b-1", Mode: ingest.ModeTransactions, Watermark: "2026-10-26", Scope: ingest.MarketDataScope{Indices: []string{"WIG20"}}}
	var first, second error
	telemetry.Run(ctx, "test", "ingest.sync", func(ctx context.Context) error {
		_, first = client.Sync(ctx, tracedInstance, call)
		_, second = client.Sync(ctx, tracedInstance, call)
		if second != nil {
			return fmt.Errorf("refused: %w", second)
		}
		return nil
	}, telemetry.KeyInstance.String(tracedInstance))
	if first != nil {
		t.Fatalf("sync: %v", first)
	}
	if !strings.Contains(fmt.Sprint(second), tracedAccount) {
		t.Fatalf("refused sync = %v, want the connector's refusal", second)
	}
	return conn
}

// callCorePorts calls the core's plugin port as a plugin does and its
// browser port as a browser does, with a forged traceparent, and returns
// the trace the browser tried to choose.
func callCorePorts(ctx context.Context, t *testing.T) trace.TraceID {
	t.Helper()
	ok := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	s := &server{}
	s.browser.Store(&ok)
	s.plugin.Store(&ok)
	plugin := httptest.NewServer(telemetry.ServerHandler(s.pluginHandler()))
	t.Cleanup(plugin.Close)
	browser := httptest.NewServer(telemetry.PublicHandler(s.browserHandler()))
	t.Cleanup(browser.Close)

	send(ctx, t, openfief.CoreHTTPClient(tracedTokens{}), plugin.URL+"/openfief.views.v1.ReadService/ListAccounts", nil)
	forged := trace.TraceID{0x4b, 0xf9, 0x2f, 0x35, 0x77, 0xb3, 0x4d, 0xa6, 0xa3, 0xce, 0x92, 0x9d, 0x0e, 0x0e, 0x47, 0x36}
	send(context.WithoutCancel(ctx), t, http.DefaultClient,
		browser.URL+"/v1/accounts/"+tracedAccount+"/holdings?isin="+tracedISIN+"&iban="+tracedIBAN, http.Header{
			"Cookie":        {"openfief_session=" + tracedSession},
			"Authorization": {"Bearer " + tracedToken},
			"Traceparent":   {"00-" + forged.String() + "-00f067aa0ba902b7-01"},
		})
	return forged
}

// send GETs url with header through client.
func send(ctx context.Context, t *testing.T, client *http.Client, url string, header http.Header) {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	maps.Copy(req.Header, header)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d", resp.StatusCode)
	}
}

// runQueries runs a query with arguments and one whose failure quotes one
// through the core's query tracer, and, when task test:pg provides a
// database, a real one through newPool.
func runQueries(ctx context.Context, t *testing.T) {
	t.Helper()
	tracer := queryTracer()
	qctx := tracer.TraceQueryStart(ctx, nil, pgx.TraceQueryStartData{
		SQL:  "INSERT INTO ledger.postings (account, amount, isin) VALUES ($1, $2, $3)",
		Args: []any{tracedAccount, tracedAmount, tracedISIN},
	})
	tracer.TraceQueryEnd(qctx, nil, pgx.TraceQueryEndData{Err: &pgconn.PgError{
		Severity: "ERROR", Code: "22P02", Message: `invalid input syntax for type numeric: "` + tracedAmount + `"`,
		Detail: "Key (account)=(" + tracedAccount + ") already exists.",
	}})
	pool, err := newPool(ctx, "postgres://synthetic@192.0.2.1:5432/openfief?connect_timeout=1")
	if err != nil {
		t.Fatalf("newPool: %v", err)
	}
	if pool.Config().ConnConfig.Tracer == nil {
		t.Error("newPool's connections have no query tracer")
	}
	pool.Close()
	dsn := os.Getenv(dsnVar)
	if dsn == "" {
		return // the unit run checks the tracer alone
	}
	db, err := newPool(ctx, dsn)
	if err != nil {
		t.Fatalf("newPool: %v", err)
	}
	defer db.Close()
	var got string
	if err := db.QueryRow(ctx, "SELECT $1::numeric::text || $2", tracedAmount, tracedIBAN).Scan(&got); err != nil {
		t.Fatalf("query: %v", err)
	}
}

// runJob enqueues a revaluation job and works it through the core's River
// plugins, as two processes would, and returns the enqueueing span.
func runJob(ctx context.Context, t *testing.T) trace.SpanContext {
	t.Helper()
	plugins := jobTracing()
	inserter, ok := plugins[0].(rivertype.JobInsertMiddleware)
	worker, isWorker := plugins[0].(rivertype.WorkerMiddleware)
	if !ok || !isWorker {
		t.Fatalf("job tracing %T is not insert and worker middleware", plugins[0])
	}
	args := []byte(`{"account":"` + tracedAccount + `","amount":"` + tracedAmount + `"}`)
	params := []*rivertype.JobInsertParams{{Kind: "valuation.revalue_account", Queue: "valuation", EncodedArgs: args}}
	var insert trace.SpanContext
	if _, err := inserter.InsertMany(ctx, params, func(ctx context.Context) ([]*rivertype.JobInsertResult, error) {
		insert = trace.SpanContextFromContext(ctx)
		return []*rivertype.JobInsertResult{}, nil
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	job := &rivertype.JobRow{ID: 1, Kind: params[0].Kind, Queue: params[0].Queue, Attempt: 1, EncodedArgs: args, Metadata: params[0].Metadata}
	// The worker runs elsewhere, outside any span.
	err := worker.Work(trace.ContextWithSpanContext(ctx, trace.SpanContext{}), job, func(ctx context.Context) error {
		telemetry.Run(ctx, "test", "valuation.revalue", func(context.Context) error { return errTracedRefusal })
		return errTracedRefusal
	})
	if !errors.Is(err, errTracedRefusal) {
		t.Fatalf("work = %v, want the worker's error", err)
	}
	return insert
}

// checkLinkedWork checks that the job's work span links to the span that
// enqueued it.
func checkLinkedWork(t *testing.T, spans tracetest.SpanStubs, insert trace.SpanContext) {
	t.Helper()
	for i := range spans {
		s := &spans[i]
		if s.Name != "river.work/valuation.revalue_account" {
			continue
		}
		if len(s.Links) != 1 || s.Links[0].SpanContext.SpanID() != insert.SpanID() || s.Links[0].SpanContext.TraceID() != insert.TraceID() {
			t.Errorf("work span links = %+v, want the enqueueing span %s", s.Links, insert.SpanID())
		}
		return
	}
	t.Error("no work span")
}

// checkNewBrowserTrace checks that the browser port started its own trace
// rather than the one the request forged.
func checkNewBrowserTrace(t *testing.T, spans tracetest.SpanStubs, forged trace.TraceID) {
	t.Helper()
	for i := range spans {
		if s := &spans[i]; s.SpanContext.TraceID() == forged {
			t.Errorf("span %q continues the browser's forged trace", s.Name)
		}
	}
}

// hasSpan reports whether spans has one called name.
func hasSpan(spans tracetest.SpanStubs, name string) bool {
	for i := range spans {
		if spans[i].Name == name {
			return true
		}
	}
	return false
}

// checkPrivate checks every exported span's name, attributes, events,
// links and status against deniedKeys and deniedValues.
func checkPrivate(t *testing.T, spans tracetest.SpanStubs) {
	t.Helper()
	if len(spans) == 0 {
		t.Fatal("no span was exported")
	}
	for i := range spans {
		s := &spans[i]
		checkText(t, s.Name, "name", s.Name)
		checkText(t, s.Name, "status", s.Status.Description)
		checkAttrs(t, s.Name, s.Attributes)
		for _, e := range s.Events {
			checkText(t, s.Name, "event", e.Name)
			checkAttrs(t, s.Name, e.Attributes)
		}
		for _, l := range s.Links {
			checkAttrs(t, s.Name, l.Attributes)
		}
	}
}

// checkAttrs checks attrs' keys and values.
func checkAttrs(t *testing.T, span string, attrs []attribute.KeyValue) {
	t.Helper()
	for _, kv := range attrs {
		key := strings.ToLower(string(kv.Key))
		for _, denied := range deniedKeys() {
			if strings.Contains(key, denied) {
				t.Errorf("span %q exports attribute %q", span, kv.Key)
			}
		}
		checkText(t, span, string(kv.Key), kv.Value.String())
	}
}

// checkText checks one exported text against deniedValues.
func checkText(t *testing.T, span, field, text string) {
	t.Helper()
	for _, re := range deniedValues() {
		if re.MatchString(text) {
			t.Errorf("span %q exports %s %q, which matches %s", span, field, text, re)
		}
	}
}
