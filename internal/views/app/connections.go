package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// Connections reads how connector instances feed accounts: the connector
// bindings and the broker accounts the instances reported (D-72).
// *ingest/postgres.Store implements it.
type Connections interface {
	// Bindings returns every connector binding, in instance then account
	// order.
	Bindings(ctx context.Context) ([]ingest.Binding, error)
	// AllBrokerAccounts returns every instance's broker accounts, in
	// instance then ID order.
	AllBrokerAccounts(ctx context.Context) ([]ingest.BrokerAccount, error)
}

// WithConnections reads connector bindings and broker accounts from c.
// Without it there are none.
func WithConnections(c Connections) Option {
	return func(v *Views) { v.connections = c }
}

// SleeveKinds returns every sleeve kind, in key order (D-71).
func (v *Views) SleeveKinds(ctx context.Context) ([]portfolio.Kind, error) {
	k, err := v.ledger.Kinds(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading sleeve kinds: %w", err)
	}
	return k, nil
}

// ConnectorBindings returns every connector binding, in instance then
// account order.
func (v *Views) ConnectorBindings(ctx context.Context) ([]ingest.Binding, error) {
	if v.connections == nil {
		return nil, nil
	}
	b, err := v.connections.Bindings(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading connector bindings: %w", err)
	}
	return b, nil
}

// BrokerAccounts returns the broker accounts connector instances reported
// (D-72), in instance then ID order, mapped and unmapped; with unmapped
// only those not mapped to an OpenFief account, whose transactions are not
// booked.
func (v *Views) BrokerAccounts(ctx context.Context, unmapped bool) ([]ingest.BrokerAccount, error) {
	if v.connections == nil {
		return nil, nil
	}
	all, err := v.connections.AllBrokerAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading broker accounts: %w", err)
	}
	if !unmapped {
		return all, nil
	}
	out := make([]ingest.BrokerAccount, 0, len(all))
	for i := range all {
		if !all[i].Mapped() {
			out = append(out, all[i])
		}
	}
	return out, nil
}
