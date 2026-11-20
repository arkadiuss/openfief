package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"strings"

	collectormetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/protobuf/proto"

	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

const (
	// maxMetricsReportBytes caps one plugin report: a few counters and a
	// histogram are a few hundred bytes.
	maxMetricsReportBytes = 256 << 10
	// protobufType is the media type of an OTLP/HTTP protobuf body.
	protobufType = "application/x-protobuf"
	// bearerWord starts a bearer Authorization header.
	bearerWord = "bearer "
)

// withPluginMetrics serves the plugin port's one telemetry path, POST
// telemetry.MetricsPath, to view plugins alone, and every other path to h.
// A plugin sends its metrics to the core there, as OTLP/HTTP, and never to a
// collector itself (D-172): the core proves who sent them, takes the service
// name from that proof, applies the allow-list of metrics and attributes
// (telemetry.FilterOTLPMetrics) and forwards what is left to the core's own
// OTLP endpoint, or drops it when the core exports no metrics. Nothing may
// store the answers.
func withPluginMetrics(h http.Handler, callers identityconnect.Callers, fwd *telemetry.MetricsForwarder, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST "+telemetry.MetricsPath, noStore(&pluginMetrics{callers: callers, forward: fwd, log: log}))
	mux.Handle("/", h)
	return mux
}

// withPluginMetricsFromEnv is withPluginMetrics with the forwarder of the
// process environment's OpenTelemetry variables: none when the core exports no
// metrics. A variable it can't use also leaves none: withTracing refused the
// same variables at start, so this is the case of a core started by a test.
func withPluginMetricsFromEnv(h http.Handler, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	fwd, err := pluginMetricsForwarder(os.Getenv)
	if err != nil {
		log.Error("plugin metrics are dropped: the OpenTelemetry variables can't be used",
			slog.String("error", kernel.LogSafe(err, telemetry.ErrBadEndpoint, telemetry.ErrBadProtocol, telemetry.ErrBadHeaders, telemetry.ErrBadExporter, telemetry.ErrBadBool)))
	}
	return withPluginMetrics(h, callers, fwd, log)
}

// pluginMetricsForwarder returns the forwarder of the core's OTLP metrics
// endpoint from getenv: nil when the core exports no metrics.
func pluginMetricsForwarder(getenv func(string) string) (*telemetry.MetricsForwarder, error) {
	f, err := telemetry.MetricsForwarderFromEnv(getenv)
	if err != nil {
		return nil, fmt.Errorf("plugin metrics: %w", err)
	}
	return f, nil
}

// pluginMetrics is the handler of a plugin's metrics report.
type pluginMetrics struct {
	callers identityconnect.Callers
	forward *telemetry.MetricsForwarder
	log     *slog.Logger
}

// ServeHTTP answers a report: 401 without a token, 403 for a caller that is no
// view plugin, 415 for a body that is not protobuf, 413 for one over the cap,
// 400 for one that is no report, and 200 with an empty OTLP answer when it was
// taken, whether it was forwarded or dropped.
func (p *pluginMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	check := p.verify(r.Context(), r.Header.Get("Authorization"))
	if check.status != 0 {
		plain(w, check.status, refusalText(check.status))
		return
	}
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != protobufType {
		plain(w, http.StatusUnsupportedMediaType, "metrics are taken as "+protobufType)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxMetricsReportBytes))
	if err != nil {
		if _, tooBig := errors.AsType[*http.MaxBytesError](err); tooBig {
			plain(w, http.StatusRequestEntityTooLarge, "report too large")
			return
		}
		plain(w, http.StatusBadRequest, "report not read")
		return
	}
	var req collectormetrics.ExportMetricsServiceRequest
	if proto.Unmarshal(body, &req) != nil {
		plain(w, http.StatusBadRequest, "not a metrics report")
		return
	}
	if telemetry.FilterOTLPMetrics(&req, check.plugin) > 0 {
		if forwardErr := p.forward.Forward(r.Context(), &req); forwardErr != nil {
			p.log.LogAttrs(r.Context(), slog.LevelWarn, "forwarding a plugin's metrics failed",
				slog.String("plugin", check.plugin), slog.String("error", kernel.LogSafe(forwardErr, telemetry.ErrForward)))
		}
	}
	// An empty ExportMetricsServiceResponse is an empty body.
	w.Header().Set("Content-Type", protobufType)
	w.WriteHeader(http.StatusOK)
}

// callerCheck is who a report's bearer token proves: the plugin, or the
// status to refuse with.
type callerCheck struct {
	plugin string
	status int
}

// verify returns the view plugin the bearer token of header proves, or the
// status to refuse with: 401 without a token, 403 for one the port refuses,
// 500 when it can't be checked.
func (p *pluginMetrics) verify(ctx context.Context, header string) callerCheck {
	if len(header) <= len(bearerWord) || !strings.EqualFold(header[:len(bearerWord)], bearerWord) {
		return callerCheck{status: http.StatusUnauthorized}
	}
	name, err := p.callers.VerifyCaller(ctx, strings.TrimSpace(header[len(bearerWord):]))
	switch {
	case err == nil:
		return callerCheck{plugin: name}
	case errors.Is(err, identityconnect.ErrCallerRefused):
		return callerCheck{status: http.StatusForbidden}
	}
	p.log.LogAttrs(ctx, slog.LevelError, "verifying a metrics caller failed", slog.String("error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded)))
	return callerCheck{status: http.StatusInternalServerError}
}

// refusalText is the short answer to a refused report.
func refusalText(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "no token"
	case http.StatusForbidden:
		return "refused"
	}
	return "internal error"
}
