package plugins_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// readingsOutcome is what one plugin and grant pair says about the index
// readings permission (D-186): whether the manifest validates, whether the
// plugin requests it, whether it waits for approval, and whether it is
// enforced.
type readingsOutcome struct {
	ManifestErrs  []fieldError
	Requested     bool
	NeedsApproval bool
	Effective     bool
}

// TestIndexReadingsPermission checks that only a connector whose signed
// manifest declares index readings, such as CPI and the NBP reference
// rate, and whose grant approves that may deliver them (D-186), as D-179
// has it for bond terms: the declaration is requested like any permission,
// waits for approval until the grant holds it, and is enforced only when
// both hold it. A view can't declare it.
func TestIndexReadingsPermission(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		declared bool
		granted  bool
		view     bool
		want     readingsOutcome
	}{
		{name: "declared and approved", declared: true, granted: true, want: readingsOutcome{Requested: true, Effective: true}},
		{name: "declared, not approved", declared: true, want: readingsOutcome{Requested: true, NeedsApproval: true}},
		{name: "approved, not declared", granted: true},
		{
			name: "a view declaring them", declared: true, granted: true, view: true,
			want: readingsOutcome{
				ManifestErrs: []fieldError{{Field: "indexReadings", Err: plugins.ErrDeliveryNeedsConnector}}, Requested: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, g := samplePlugin(), fullGrant()
			p.Manifest.IndexReadings, g.Permissions.IndexReadings = tt.declared, tt.granted
			if tt.view {
				p.Manifest.Capabilities = []plugins.Capability{plugins.CapabilityView}
				g.Permissions.Capabilities = []plugins.Capability{plugins.CapabilityView}
			}
			var got readingsOutcome
			got.ManifestErrs = fieldErrors(p.Manifest.Validate())
			got.Requested = p.Requested().IndexReadings
			if pending, err := plugins.NeedsApproval(&g, &p); err == nil {
				got.NeedsApproval = pending.IndexReadings
			}
			if e, err := plugins.Effective(&p, &g); err == nil {
				got.Effective = e.IndexReadings()
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("index readings (-want +got):\n%s", diff)
			}
		})
	}
}

// TestIndexReadingsLabel checks that a manifest label carries the
// indexReadings declaration, that a view's can't, and that manifests
// differing only in it are not equal (D-186).
func TestIndexReadingsLabel(t *testing.T) {
	t.Parallel()
	m, err := plugins.ParseManifestLabel(`{"capabilities":["connector"],"indexReadings":true}`)
	if err != nil {
		t.Fatalf("ParseManifestLabel: %v", err)
	}
	if !m.IndexReadings {
		t.Errorf("the label's declaration was not read: %+v", m)
	}
	if _, viewErr := plugins.ParseManifestLabel(`{"capabilities":["view"],"indexReadings":true}`); viewErr == nil {
		t.Error("a view's label declaring index readings was accepted")
	}
	if other := (plugins.Manifest{Capabilities: m.Capabilities}); m.Equal(&other) {
		t.Error("manifests differing in index readings are equal")
	}
}

// TestIndexReadingsDiff checks that effective permissions differing only
// in index readings are not equal, and that an upgrade adding or dropping
// them shows in the permissions' diff (D-90, D-186).
func TestIndexReadingsDiff(t *testing.T) {
	t.Parallel()
	p, g := samplePlugin(), fullGrant()
	p.Manifest.IndexReadings, g.Permissions.IndexReadings = true, true
	with, err := plugins.Effective(&p, &g)
	if err != nil {
		t.Fatalf("Effective: %v", err)
	}
	g.Permissions.IndexReadings = false
	without, err := plugins.Effective(&p, &g)
	if err != nil {
		t.Fatalf("Effective: %v", err)
	}
	if with.Equal(without) {
		t.Error("effective permissions differing in index readings are equal")
	}
	from, to := plugins.Permissions{}, plugins.Permissions{IndexReadings: true}
	if to.IsEmpty() {
		t.Error("permissions holding only index readings are empty")
	}
	if added := plugins.DiffPermissions(&from, &to); !added.Added.IndexReadings || added.Removed.IndexReadings {
		t.Errorf("adding index readings: diff %+v", added)
	}
	if removed := plugins.DiffPermissions(&to, &from); removed.Added.IndexReadings || !removed.Removed.IndexReadings {
		t.Errorf("dropping index readings: diff %+v", removed)
	}
}
