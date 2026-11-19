package openfief

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

// pluginService is a plugin's service name in its spans when neither
// OTEL_SERVICE_NAME nor OPENFIEF_INSTANCE names it.
const pluginService = "openfief-plugin"

// StartTelemetry sets up the plugin's tracing from the standard
// OpenTelemetry variables (telemetry.ParseConfig), naming the service after
// the plugin's instance. Export stays off unless the plugin's environment
// names an OTLP endpoint, which the operator doesn't set; the trace the
// core propagates continues through the plugin either way. ServeConnector
// calls it; a view plugin that serves its own port calls it itself and
// calls the returned function to send what is buffered before it exits.
func StartTelemetry(ctx context.Context) (func(context.Context) error, error) {
	service := os.Getenv(EnvInstance)
	if service == "" {
		service = pluginService
	}
	t, err := telemetry.StartFromEnv(ctx, service)
	if err != nil {
		return nil, fmt.Errorf("plugin telemetry: %w", err)
	}
	return t.Shutdown, nil
}

// metricsReportTimeout bounds one report of a plugin's metrics to the core.
const metricsReportTimeout = 10 * time.Second

// StartMetrics sends the plugin's metrics to the core, every minute, with
// the plugin's projected token tokens, and returns the function that sends
// what is buffered before the plugin exits. The core applies its allow-list
// to them (D-172) and exports them with its own, where its OTLP endpoint is
// set; the plugin never reaches a collector itself, and its metrics go
// nowhere when the core exports none. Only a view plugin calls it: the core
// takes reports from view plugins alone. The instruments are the ones the
// plugin creates with OpenTelemetry's global meter, before or after this
// call. OTEL_SDK_DISABLED=true or OTEL_METRICS_EXPORTER=none turns it off.
func StartMetrics(ctx context.Context, tokens TokenSource) (func(context.Context) error, error) {
	if telemetry.MetricsOff(os.Getenv) {
		return func(context.Context) error { return nil }, nil
	}
	coreURL, err := CoreURL()
	if err != nil {
		return nil, fmt.Errorf("plugin metrics: %w", err)
	}
	service := os.Getenv(EnvInstance)
	if service == "" {
		service = pluginService
	}
	client := coreClient(tokens, false)
	client.Timeout = metricsReportTimeout
	t, err := telemetry.StartMetricsTo(ctx, coreURL+telemetry.MetricsPath, client, service)
	if err != nil {
		return nil, fmt.Errorf("plugin metrics: %w", err)
	}
	return t.Shutdown, nil
}

// withTelemetry runs serve with the plugin's tracing started, and shuts it
// down after serve returns, within the shutdown timeout.
func withTelemetry(ctx context.Context, serve func() error) error {
	stop, err := StartTelemetry(ctx)
	if err != nil {
		return err
	}
	serveErr := serve()
	flush, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	return errors.Join(serveErr, stop(flush))
}
