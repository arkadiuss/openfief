package plugins_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// marketDataOutcome is what one plugin and grant pair says about the
// market-data permission (D-123): whether the manifest validates, whether
// the plugin requests it, whether it waits for approval, and whether it
// is enforced.
type marketDataOutcome struct {
	ManifestErrs  []fieldError
	Requested     bool
	NeedsApproval bool
	Effective     bool
}

// TestMarketData checks that only a connector whose manifest declares it a
// market-data connector, and whose grant approves that, is one (D-123):
// the declaration is requested like any permission, waits for approval
// until the grant holds it, and is enforced only when both hold it. A
// view can't declare it.
func TestMarketData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		declared bool
		granted  bool
		view     bool
		want     marketDataOutcome
	}{
		{name: "declared and approved", declared: true, granted: true, want: marketDataOutcome{Requested: true, Effective: true}},
		{name: "declared, not approved", declared: true, want: marketDataOutcome{Requested: true, NeedsApproval: true}},
		{name: "approved, not declared", granted: true, want: marketDataOutcome{}},
		{name: "neither", want: marketDataOutcome{}},
		{
			name: "a view declaring it", declared: true, granted: true, view: true,
			want: marketDataOutcome{ManifestErrs: []fieldError{{Field: "marketData", Err: plugins.ErrMarketDataNeedsConnector}}, Requested: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, g := samplePlugin(), fullGrant()
			p.Manifest.MarketData, g.Permissions.MarketData = tt.declared, tt.granted
			if tt.view {
				p.Manifest.Capabilities = []plugins.Capability{plugins.CapabilityView}
				g.Permissions.Capabilities = []plugins.Capability{plugins.CapabilityView}
			}
			var got marketDataOutcome
			got.ManifestErrs = fieldErrors(p.Manifest.Validate())
			got.Requested = p.Requested().MarketData
			if pending, err := plugins.NeedsApproval(&g, &p); err == nil {
				got.NeedsApproval = pending.MarketData
			}
			if e, err := plugins.Effective(&p, &g); err == nil {
				got.Effective = e.MarketData()
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("market data (-want +got):\n%s", diff)
			}
		})
	}
}

// TestMarketDataLabelAndDiff checks that a manifest label carries the
// declaration (marketData), that two manifests differing only in it are
// not equal, and that an upgrade adding or dropping it shows in the
// permissions' diff (D-90, D-123).
func TestMarketDataLabelAndDiff(t *testing.T) {
	t.Parallel()
	m, err := plugins.ParseManifestLabel(`{"capabilities":["connector"],"marketData":true}`)
	if err != nil {
		t.Fatalf("ParseManifestLabel: %v", err)
	}
	if !m.MarketData {
		t.Error("the label's marketData was not read")
	}
	if _, err := plugins.ParseManifestLabel(`{"capabilities":["view"],"marketData":true}`); err == nil {
		t.Error("a view's label declaring market data was accepted")
	}
	plain := plugins.Manifest{Capabilities: m.Capabilities}
	if m.Equal(&plain) {
		t.Error("manifests differing in marketData are equal")
	}
	from, to := plugins.Permissions{}, plugins.Permissions{MarketData: true}
	added := plugins.DiffPermissions(&from, &to)
	if !added.Added.MarketData || added.Removed.MarketData || added.IsEmpty() {
		t.Errorf("adding market data: diff %+v", added)
	}
	removed := plugins.DiffPermissions(&to, &from)
	if removed.Added.MarketData || !removed.Removed.MarketData {
		t.Errorf("dropping market data: diff %+v", removed)
	}
}
