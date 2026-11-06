package telemetry

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// procedurePrefix starts the path of every OpenFief Connect procedure, such
// as /openfief.views.v1.ReadService/ListAccounts.
const procedurePrefix = "/openfief."

// ServerHandler returns h in a server span per request, continuing the
// trace the caller propagated: for a port only OpenFief's own processes
// reach, such as the core's plugin port or a plugin's. GET /healthz and
// GET /readyz are not traced.
func ServerHandler(h http.Handler) http.Handler {
	return otelhttp.NewHandler(h, "", otelhttp.WithSpanNameFormatter(spanName), otelhttp.WithFilter(traced))
}

// PublicHandler is ServerHandler for a port browsers and scripts reach,
// the core's browser port: each request starts a new trace, linked to the
// caller's, so no caller can choose a trace or force sampling.
func PublicHandler(h http.Handler) http.Handler {
	return otelhttp.NewHandler(h, "", otelhttp.WithSpanNameFormatter(spanName), otelhttp.WithFilter(traced),
		otelhttp.WithPublicEndpointFn(func(*http.Request) bool { return true }))
}

// Transport returns base in a client span per request that propagates the
// trace in a W3C traceparent header: for calls between OpenFief's own
// processes, never to the internet. A nil base is http.DefaultTransport.
func Transport(base http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(base, otelhttp.WithSpanNameFormatter(spanName))
}

// spanName names a request's span by its Connect procedure, such as
// openfief.connector.v1.ConnectorService/Sync, or else by its method
// alone: never by a path that may hold an ID.
func spanName(_ string, r *http.Request) string {
	if p, ok := procedure(r.URL.Path); ok {
		return p
	}
	return r.Method
}

// procedure returns the Connect procedure path names, without its leading
// slash, when it is one of OpenFief's: a package starting openfief., a
// service and a method.
func procedure(path string) (string, bool) {
	rest, ok := strings.CutPrefix(path, procedurePrefix)
	if !ok {
		return "", false
	}
	service, method, ok := strings.Cut(rest, "/")
	if !ok || service == "" || method == "" || strings.Contains(method, "/") {
		return "", false
	}
	return path[1:], true
}

// traced reports whether a request gets a span: every one but the health
// probes'.
func traced(r *http.Request) bool {
	return r.Method != http.MethodGet || (r.URL.Path != "/healthz" && r.URL.Path != "/readyz")
}
