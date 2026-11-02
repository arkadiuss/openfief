package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/rivertype"
	"github.com/riverqueue/rivercontrib/otelriver"

	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

const (
	// coreService names the core in its spans unless OTEL_SERVICE_NAME
	// does.
	coreService = "openfief-core"
	// flushTimeout bounds sending the spans still buffered at exit.
	flushTimeout = 5 * time.Second
)

// withTracing runs fn with the core's tracing started from the standard
// OpenTelemetry variables (telemetry.ParseConfig): off, with nothing sent,
// unless an OTLP endpoint is set. After fn it sends the spans still
// buffered, for up to flushTimeout. A variable it can't use fails before fn
// runs.
func withTracing(ctx context.Context, fn func() error) error {
	tel, err := telemetry.StartFromEnv(ctx, coreService)
	if err != nil {
		return fmt.Errorf("core tracing: %w", err)
	}
	runErr := fn()
	flush, cancel := context.WithTimeout(context.WithoutCancel(ctx), flushTimeout)
	defer cancel()
	return errors.Join(runErr, tel.Shutdown(flush))
}

// newPool connects to dsn with queryTracer on every connection.
func newPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting: %w", err)
	}
	cfg.ConnConfig.Tracer = queryTracer()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connecting: %w", err)
	}
	return pool, nil
}

// queryTracer records each query, batch and copy that runs inside a traced
// operation as a span named by its SQL operation, such as SELECT or INSERT:
// never the statement, its arguments or the connection's details (GR-12).
// A query outside any span, such as River polling its queue, records
// nothing.
func queryTracer() *otelpgx.Tracer {
	return otelpgx.NewTracer(otelpgx.WithDisableSQLStatementInAttributes(),
		otelpgx.WithDisableConnectionDetailsInAttributes(), otelpgx.WithDisableAcquireTracer())
}

// jobTracing returns the River plugins every client of the revaluation
// queue installs (D-36): spans for inserting jobs and for working them,
// named by the job's kind, with the inserting span's trace context stored
// in the job's metadata so the work span links back to the change that
// queued it. The spans carry the job's kind, queue, attempt and outcome,
// never its args. With tracing off they record nothing.
func jobTracing() []rivertype.Plugin {
	return []rivertype.Plugin{otelriver.NewMiddleware(&otelriver.MiddlewareConfig{
		EnableTracePropagation:      true,
		EnableWorkSpanJobKindSuffix: true,
	})}
}

// pluginClient is the HTTP client the core calls connector instances with
// (D-49): HTTP/2 without TLS, each call in a client span that propagates
// the core's trace to the plugin in a W3C traceparent header.
func pluginClient() *http.Client {
	c := connectrpc.H2CClient()
	c.Transport = telemetry.Transport(c.Transport)
	return c
}
