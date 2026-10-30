package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestConnectorsRefuseUnboundAccounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		bindings    *fakeBindings
		edit        func(d *app.Delivery)
		wantErr     error
		wantImports int
		wantLookups int
	}{
		{name: "bound account imports", bindings: &fakeBindings{accounts: []ledger.AccountID{"acc-1"}}, wantImports: 1, wantLookups: 1},
		{name: "no binding refuses the whole delivery", bindings: &fakeBindings{}, wantErr: ingest.ErrAccountNotBound, wantLookups: 1},
		{name: "binding to another account refuses it", bindings: &fakeBindings{accounts: []ledger.AccountID{"acc-2"}}, wantErr: ingest.ErrAccountNotBound, wantLookups: 1},
		{name: "a failing binding store imports nothing", bindings: &fakeBindings{err: errStore}, wantErr: errStore, wantLookups: 1},
		{
			name: "a delivery without transactions or assertions needs no binding", bindings: &fakeBindings{},
			edit:        func(d *app.Delivery) { d.Mapping = ingest.Mapping{} },
			wantImports: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := delivery(t)
			if tt.edit != nil {
				tt.edit(&d)
			}
			l, stores := &fakeLedger{}, &fakeStores{}
			c := connectors(t, &fakeCaller{delivery: d}, l, stores)
			c.Bindings = tt.bindings
			_, err := c.Sync(t.Context(), "xtb-main", &app.SyncCall{ID: "b-1", Mode: ingest.ModeTransactions})
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Sync error (-want +got):\n%s", diff)
			}
			if l.calls != tt.wantImports || tt.bindings.calls != tt.wantLookups {
				t.Errorf("ledger imports = %d, binding lookups = %d; want %d and %d", l.calls, tt.bindings.calls, tt.wantImports, tt.wantLookups)
			}
			if tt.wantErr != nil && stores.calls != 0 {
				t.Errorf("a refused delivery stored %d valuation inputs", stores.calls)
			}
		})
	}
}

func TestConnectorsSeedOnlyInOpeningMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		mode                 ingest.Mode
		mapping              func(t *testing.T) ingest.Mapping
		wantErr              error
		wantCalls, wantSeeds int
	}{
		{name: "opening batch seeds", mode: ingest.ModeOpening, mapping: openingMapping, wantCalls: 1, wantSeeds: 1},
		{name: "transactions batch imports", mode: ingest.ModeTransactions, mapping: mapping, wantCalls: 1},
		{name: "transactions batch with opening cash is refused", mode: ingest.ModeTransactions, mapping: openingMapping, wantErr: app.ErrSeedingOutsideOpening},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := delivery(t)
			d.Mapping = tt.mapping(t)
			l := &fakeLedger{}
			_, err := connectors(t, &fakeCaller{delivery: d}, l, &fakeStores{}).Sync(t.Context(), "xtb-main", &app.SyncCall{ID: "b-1", Mode: tt.mode})
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Sync error (-want +got):\n%s", diff)
			}
			if l.calls != tt.wantCalls || l.seeds != tt.wantSeeds {
				t.Errorf("ledger calls = %d, seeds = %d; want %d and %d", l.calls, l.seeds, tt.wantCalls, tt.wantSeeds)
			}
		})
	}
}
