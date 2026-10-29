package openfief

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

// shutdownTimeout bounds how long calls in flight may run on after
// ServeConnector's context is done.
const shutdownTimeout = 20 * time.Second

// ServeConnector serves svc's connector contract to the core until ctx is
// done (D-49, D-59): on ListenAddress, behind the CoreVerifier
// NewCoreVerifier builds from the plugin's environment. It then stops
// taking calls and gives the ones in flight up to 20 seconds to finish. It
// returns nil after such a shutdown and the error otherwise.
func ServeConnector(ctx context.Context, svc connectorv1connect.ConnectorServiceHandler, opts ...connect.HandlerOption) error {
	addr, err := ListenAddress()
	if err != nil {
		return err
	}
	v, err := NewCoreVerifier()
	if err != nil {
		return err
	}
	return serveUntilDone(ctx, NewConnectorServer(addr, svc, v, opts...))
}

// serveUntilDone runs srv until it fails or ctx is done, then shuts it down.
func serveUntilDone(ctx context.Context, srv *http.Server) error {
	done := make(chan error, 1)
	go func() { done <- srv.ListenAndServe() }()
	select {
	case err := <-done:
		return fmt.Errorf("serve connector: %w", err)
	case <-ctx.Done():
	}
	stop, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(stop); err != nil {
		return fmt.Errorf("shut down connector: %w", err)
	}
	if err := <-done; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve connector: %w", err)
	}
	return nil
}
