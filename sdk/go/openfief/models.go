package openfief

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

// NewModelProviderServer returns a server on addr, such as ListenAddress
// gives, that serves svc's model-provider contract (D-205) to the core only
// (RequireCore). Like NewConnectorServer it speaks HTTP/1.1 and HTTP/2
// without TLS and reads at most MaxMessageBytes of a message, which is far
// above a request to a model.
func NewModelProviderServer(addr string, svc modelsv1connect.ModelProviderServiceHandler, v CoreVerifier, opts ...connect.HandlerOption) *http.Server {
	opts = append([]connect.HandlerOption{connect.WithReadMaxBytes(MaxMessageBytes)}, opts...)
	path, h := modelsv1connect.NewModelProviderServiceHandler(svc, opts...)
	return newCoreServer(addr, path, h, v)
}

// ServeModelProvider serves svc's model-provider contract to the core until
// ctx is done, the way ServeConnector serves a connector: on ListenAddress,
// behind the CoreVerifier NewCoreVerifier builds from the plugin's
// environment, with the plugin's tracing started, and a graceful stop.
func ServeModelProvider(ctx context.Context, svc modelsv1connect.ModelProviderServiceHandler, opts ...connect.HandlerOption) error {
	addr, err := ListenAddress()
	if err != nil {
		return err
	}
	v, err := NewCoreVerifier()
	if err != nil {
		return err
	}
	return withTelemetry(ctx, func() error {
		return serveUntilDone(ctx, NewModelProviderServer(addr, svc, v, opts...))
	})
}
