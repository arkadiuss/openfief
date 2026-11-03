package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
)

// Connections reads how connector instances feed accounts: the connector
// bindings. Broker accounts are the owner's AccountService's to list (D-121).
// *ingest/postgres.Store implements it.
type Connections interface {
	// Bindings returns every connector binding, in instance then account
	// order.
	Bindings(ctx context.Context) ([]ingest.Binding, error)
}

// WithConnections reads connector bindings from c.
// Without it there are none.
func WithConnections(c Connections) Option {
	return func(v *Views) { v.connections = c }
}

// ConnectorBindings returns every connector binding to an account the
// caller may view (D-163), in instance then account order.
func (v *Views) ConnectorBindings(ctx context.Context) ([]ingest.Binding, error) {
	if v.connections == nil {
		return nil, nil
	}
	b, err := v.connections.Bindings(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading connector bindings: %w", err)
	}
	return v.viewableBindings(ctx, b)
}
