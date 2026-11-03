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

// fakeConnections holds one binding of
// ibkr-main, or fails with fail.
type fakeConnections struct {
	fail error
}

func (f fakeConnections) Bindings(context.Context) ([]ingest.Binding, error) {
	return []ingest.Binding{{Instance: "ibkr-main", Account: "acc-1"}}, f.fail
}

func TestConnections(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"), app.WithConnections(fakeConnections{}))
	if err != nil {
		t.Fatal(err)
	}
	ctx := reader(t)
	kinds, err := v.SleeveKinds(ctx)
	if diff := cmp.Diff(portfolio.DefaultKinds(), kinds); err != nil || diff != "" {
		t.Errorf("SleeveKinds = %v (-want +got):\n%s", err, diff)
	}
	checkConnections(t, v)
}

// checkConnections checks v's bindings over
// fakeConnections.
func checkConnections(t *testing.T, v *app.Views) {
	t.Helper()
	ctx := reader(t)
	bindings, err := v.ConnectorBindings(ctx)
	if diff := cmp.Diff([]ingest.Binding{{Instance: "ibkr-main", Account: "acc-1"}}, bindings); err != nil || diff != "" {
		t.Errorf("ConnectorBindings = %v (-want +got):\n%s", err, diff)
	}
}

func TestConnectionsWithoutThePortOrFailing(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := reader(t)
	none := f.views(t)
	if b, err := none.ConnectorBindings(ctx); err != nil || len(b) != 0 {
		t.Errorf("ConnectorBindings without the port = %v, %v; want none", b, err)
	}
	failing, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"), app.WithConnections(fakeConnections{fail: errStore}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := failing.ConnectorBindings(ctx); !cmp.Equal(err, errStore, cmpopts.EquateErrors()) {
		t.Errorf("ConnectorBindings with the store down = %v, want %v", err, errStore)
	}
	f.ledger.fail = errStore
	if _, err := none.SleeveKinds(ctx); !cmp.Equal(err, errStore, cmpopts.EquateErrors()) {
		t.Errorf("SleeveKinds with the store down = %v, want %v", err, errStore)
	}
}
