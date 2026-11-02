package plugins_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// deliveriesOutcome is what one plugin and grant pair says about the bond
// terms and index levels permissions (D-179), each as [bond terms, index
// levels]: whether the manifest validates, whether the plugin requests
// them, whether they wait for approval, and whether they are enforced.
type deliveriesOutcome struct {
	ManifestErrs  []fieldError
	Requested     [2]bool
	NeedsApproval [2]bool
	Effective     [2]bool
}

// TestDeliveries checks that only a connector whose signed manifest
// declares bond terms, or index levels, and whose grant approves that may
// deliver them (D-179): each declaration is requested like any
// permission, waits for approval until the grant holds it, and is
// enforced only when both hold it. A view can't declare either.
func TestDeliveries(t *testing.T) {
	t.Parallel()
	both := [2]bool{true, true}
	tests := []struct {
		name     string
		declared [2]bool
		granted  [2]bool
		view     bool
		want     deliveriesOutcome
	}{
		{name: "both declared and approved", declared: both, granted: both, want: deliveriesOutcome{Requested: both, Effective: both}},
		{
			name: "both declared, bond terms approved", declared: both, granted: [2]bool{true, false},
			want: deliveriesOutcome{Requested: both, NeedsApproval: [2]bool{false, true}, Effective: [2]bool{true, false}},
		},
		{name: "declared, not approved", declared: both, want: deliveriesOutcome{Requested: both, NeedsApproval: both}},
		{name: "approved, not declared", granted: both},
		{
			name: "a view declaring them", declared: both, granted: both, view: true,
			want: deliveriesOutcome{ManifestErrs: []fieldError{
				{Field: "bondTerms", Err: plugins.ErrDeliveryNeedsConnector}, {Field: "indexLevels", Err: plugins.ErrDeliveryNeedsConnector},
			}, Requested: both},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, g := samplePlugin(), fullGrant()
			p.Manifest.BondTerms, p.Manifest.IndexLevels = tt.declared[0], tt.declared[1]
			g.Permissions.BondTerms, g.Permissions.IndexLevels = tt.granted[0], tt.granted[1]
			if tt.view {
				p.Manifest.Capabilities = []plugins.Capability{plugins.CapabilityView}
				g.Permissions.Capabilities = []plugins.Capability{plugins.CapabilityView}
			}
			var got deliveriesOutcome
			got.ManifestErrs = fieldErrors(p.Manifest.Validate())
			r := p.Requested()
			got.Requested = [2]bool{r.BondTerms, r.IndexLevels}
			if pending, err := plugins.NeedsApproval(&g, &p); err == nil {
				got.NeedsApproval = [2]bool{pending.BondTerms, pending.IndexLevels}
			}
			if e, err := plugins.Effective(&p, &g); err == nil {
				got.Effective = [2]bool{e.BondTerms(), e.IndexLevels()}
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("deliveries (-want +got):\n%s", diff)
			}
		})
	}
}

// TestDeliveriesLabel checks that a manifest label carries both
// declarations (bondTerms, indexLevels), and that manifests differing only
// in one are not equal (D-179).
func TestDeliveriesLabel(t *testing.T) {
	t.Parallel()
	m, err := plugins.ParseManifestLabel(`{"capabilities":["connector"],"bondTerms":true,"indexLevels":true}`)
	if err != nil {
		t.Fatalf("ParseManifestLabel: %v", err)
	}
	if !m.BondTerms || !m.IndexLevels {
		t.Errorf("the label's declarations were not read: %+v", m)
	}
	if _, err := plugins.ParseManifestLabel(`{"capabilities":["view"],"bondTerms":true}`); err == nil {
		t.Error("a view's label declaring bond terms was accepted")
	}
	for _, other := range []plugins.Manifest{
		{Capabilities: m.Capabilities, BondTerms: true}, {Capabilities: m.Capabilities, IndexLevels: true},
	} {
		if m.Equal(&other) {
			t.Errorf("manifests differing in a declaration are equal: %+v", other)
		}
	}
}

// TestDeliveriesDiff checks that an upgrade adding or dropping bond terms or
// index levels shows in the permissions' diff (D-90, D-179).
func TestDeliveriesDiff(t *testing.T) {
	t.Parallel()
	from, to := plugins.Permissions{}, plugins.Permissions{BondTerms: true, IndexLevels: true}
	added := plugins.DiffPermissions(&from, &to)
	if !added.Added.BondTerms || !added.Added.IndexLevels || added.Removed.BondTerms || added.IsEmpty() {
		t.Errorf("adding bond terms and index levels: diff %+v", added)
	}
	removed := plugins.DiffPermissions(&to, &from)
	if removed.Added.IndexLevels || !removed.Removed.BondTerms || !removed.Removed.IndexLevels {
		t.Errorf("dropping bond terms and index levels: diff %+v", removed)
	}
}
