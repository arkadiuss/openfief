package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// fakeConnections holds one binding and two synthetic broker accounts of
// ibkr-main, one unmapped, or fails with fail.
type fakeConnections struct {
	fail error
}

func (f fakeConnections) Bindings(context.Context) ([]ingest.Binding, error) {
	return []ingest.Binding{{Instance: "ibkr-main", Account: "acc-1"}}, f.fail
}

func (f fakeConnections) AllBrokerAccounts(context.Context) ([]ingest.BrokerAccount, error) {
	return []ingest.BrokerAccount{
		{Instance: "ibkr-main", ID: "U0000001", Institution: "ibkr", Account: "acc-1"},
		{Instance: "ibkr-main", ID: "U0000002", Institution: "ibkr"},
	}, f.fail
}

func TestConnections(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), app.WithConnections(fakeConnections{}))
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	kinds, err := v.SleeveKinds(ctx)
	if diff := cmp.Diff(portfolio.DefaultKinds(), kinds); err != nil || diff != "" {
		t.Errorf("SleeveKinds = %v (-want +got):\n%s", err, diff)
	}
	checkConnections(t, v)
}

// checkConnections checks v's bindings and broker accounts over
// fakeConnections.
func checkConnections(t *testing.T, v *app.Views) {
	t.Helper()
	ctx := t.Context()
	bindings, err := v.ConnectorBindings(ctx)
	if diff := cmp.Diff([]ingest.Binding{{Instance: "ibkr-main", Account: "acc-1"}}, bindings); err != nil || diff != "" {
		t.Errorf("ConnectorBindings = %v (-want +got):\n%s", err, diff)
	}
	all, err := v.BrokerAccounts(ctx, false)
	if err != nil || len(all) != 2 {
		t.Errorf("BrokerAccounts = %v, %v; want both", all, err)
	}
	unmapped, err := v.BrokerAccounts(ctx, true)
	if err != nil || len(unmapped) != 1 || unmapped[0].ID != "U0000002" {
		t.Errorf("unmapped BrokerAccounts = %v, %v; want only U0000002", unmapped, err)
	}
}

func TestConnectionsWithoutThePortOrFailing(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := t.Context()
	none := f.views(t)
	if b, err := none.ConnectorBindings(ctx); err != nil || len(b) != 0 {
		t.Errorf("ConnectorBindings without the port = %v, %v; want none", b, err)
	}
	if b, err := none.BrokerAccounts(ctx, true); err != nil || len(b) != 0 {
		t.Errorf("BrokerAccounts without the port = %v, %v; want none", b, err)
	}
	failing, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), app.WithConnections(fakeConnections{fail: errStore}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := failing.ConnectorBindings(ctx); !cmp.Equal(err, errStore, cmpopts.EquateErrors()) {
		t.Errorf("ConnectorBindings with the store down = %v, want %v", err, errStore)
	}
	if _, err := failing.BrokerAccounts(ctx, false); !cmp.Equal(err, errStore, cmpopts.EquateErrors()) {
		t.Errorf("BrokerAccounts with the store down = %v, want %v", err, errStore)
	}
	f.ledger.fail = errStore
	if _, err := none.SleeveKinds(ctx); !cmp.Equal(err, errStore, cmpopts.EquateErrors()) {
		t.Errorf("SleeveKinds with the store down = %v, want %v", err, errStore)
	}
}
