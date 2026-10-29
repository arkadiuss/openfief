package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// fakePlugins answers the installed plugins from fixed records, or fails.
type fakePlugins struct {
	records []views.PluginRecord
	fail    error
}

func (f fakePlugins) InstalledPlugins(context.Context) ([]views.PluginRecord, error) {
	return f.records, f.fail
}

// installedViews returns the fixture's views reading the installed plugins
// from plugins, with acc-1 holding one transaction from each of the
// connector instances ibkr-main and ibkr-old and one from no instance.
func installedViews(t *testing.T, plugins app.Plugins) *app.Views {
	t.Helper()
	f := newFixture(t)
	day := date(t, "2026-03-02")
	f.ledger.txs = map[ledger.AccountID][]ledger.Transaction{"acc-1": {
		{Key: ledger.TransactionKey{Source: ingest.ConnectorKeySource("ibkr-main", "ibkr"), ExternalID: "1"}, Account: "acc-1", Kind: ledger.KindDeposit, Date: day},
		{Key: ledger.TransactionKey{Source: ingest.ConnectorKeySource("ibkr-old", "ibkr"), ExternalID: "2"}, Account: "acc-1", Kind: ledger.KindDeposit, Date: day},
		{Key: ledger.TransactionKey{Source: "seed", ExternalID: "3"}, Account: "acc-1", Kind: ledger.KindDeposit, Date: day},
	}}
	var opts []app.Option
	if plugins != nil {
		opts = append(opts, app.WithPlugins(plugins))
	}
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), opts...)
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	return v
}

// record is a running, CI-signed plugin record named instance.
func record(instance string) views.PluginRecord {
	return views.PluginRecord{
		Instance: instance, Image: "registry.example.com/openfief/" + instance + "@sha256:0f0f",
		Effective: []string{"connector"}, Requested: []string{"connector"},
		Ready:     views.Condition{Status: "True", Reason: "Running"},
		Signature: views.Condition{Status: "True", Reason: "SignedByTrustedSigner"},
		Trust:     "Unverified", SignerKind: "OpenFiefCI", Signer: "openfief-ci",
	}
}

func TestPlugins(t *testing.T) {
	t.Parallel()
	ui, ibkr := record("ui"), record("ibkr-main")
	v := installedViews(t, fakePlugins{records: []views.PluginRecord{ui, ibkr}})
	got, err := v.Plugins(context.Background())
	if err != nil {
		t.Fatalf("Plugins: %v", err)
	}
	want := []views.Plugin{views.PluginOf(&ibkr), views.PluginOf(&ui)}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Plugins (-want +got):\n%s", diff)
	}
	if got[0].Trust != views.TrustOpenFiefCI {
		t.Errorf("trust = %v, want signed by OpenFief's CI", got[0].Trust)
	}
}

func TestPluginsUnknown(t *testing.T) {
	t.Parallel()
	notSynced := fakePlugins{fail: fmt.Errorf("%w: not synced", app.ErrPluginsUnknown)}
	tests := []struct {
		name    string
		plugins app.Plugins
		want    error
	}{
		{"no Plugins port, outside a cluster", nil, app.ErrPluginsUnknown},
		{"the cluster not read yet", notSynced, app.ErrPluginsUnknown},
		{"the port fails otherwise", fakePlugins{fail: errStore}, errStore},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := installedViews(t, tt.plugins)
			if _, err := v.Plugins(context.Background()); !errors.Is(err, tt.want) {
				t.Errorf("Plugins error = %v, want %v", err, tt.want)
			}
			page, err := v.Transactions(context.Background(), "acc-1", nil, views.DefaultPageSize)
			if err != nil {
				t.Fatalf("Transactions: %v, want the page without removed marks", err)
			}
			want := []views.Origin{{Instance: "ibkr-main"}, {Instance: "ibkr-old"}, {}}
			if diff := cmp.Diff(want, page.Origins); diff != "" {
				t.Errorf("origins while unknown (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTransactionsMarkRemovedInstances(t *testing.T) {
	t.Parallel()
	v := installedViews(t, fakePlugins{records: []views.PluginRecord{record("ibkr-main"), record("ui")}})
	page, err := v.Transactions(context.Background(), "acc-1", nil, views.DefaultPageSize)
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	want := []views.Origin{{Instance: "ibkr-main"}, {Instance: "ibkr-old", Removed: true}, {}}
	if diff := cmp.Diff(want, page.Origins); diff != "" {
		t.Errorf("origins (-want +got):\n%s", diff)
	}
	if len(page.Transactions) != len(want) {
		t.Errorf("%d transactions, want every one kept (D-58)", len(page.Transactions))
	}
}
