package connectrpc_test

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

// errBrokerDown stands for a connector's own reason, which the core never
// keeps.
var errBrokerDown = errors.New("synthetic broker outage")

// answeringConnector answers every sync with code.
type answeringConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	code connect.Code
}

func (a answeringConnector) Sync(context.Context, *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	return nil, refuse(a.code, errBrokerDown)
}

// TestClientSyncErrorClasses checks that a failed Sync wraps what the
// scheduler classifies it by: an instance with no address or a connector
// answering UNAVAILABLE can't be reached, and any other answer is the
// connector's own error.
func TestClientSyncErrorClasses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		conn     connectorv1connect.ConnectorServiceHandler
		instance string
		want     ingest.SyncErrorClass
	}{
		{name: "no address", conn: &fakeConnector{}, instance: "nowhere", want: ingest.SyncErrorUnavailable},
		{name: "unavailable", conn: answeringConnector{code: connect.CodeUnavailable}, instance: "xtb-main", want: ingest.SyncErrorUnavailable},
		{name: "unimplemented", conn: failingConnector{}, instance: "xtb-main", want: ingest.SyncErrorConnector},
		{name: "permission denied", conn: answeringConnector{code: connect.CodePermissionDenied}, instance: "xtb-main", want: ingest.SyncErrorConnector},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newClient(t, tt.conn)
			_, err := c.Sync(t.Context(), tt.instance, &app.SyncCall{ID: "b-1", Mode: ingest.ModeTransactions})
			if got := app.SyncErrorClassOf(err); got != tt.want {
				t.Errorf("class of %v = %v, want %v", err, got, tt.want)
			}
		})
	}
}
