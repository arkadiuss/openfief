package openfief

import (
	"context"
	"errors"
	"fmt"
	"os"

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
