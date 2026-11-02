package plugins_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// TestManifestInstitutions checks that a connector's manifest label
// carries the institutions it books for, that each is a lower-case label
// named once, that only a connector may declare them, and that two
// manifests differing only in them are not equal, so the operator refuses
// a Plugin whose institutions differ from its image's signed label
// (D-182, D-42).
func TestManifestInstitutions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		label   string
		want    []string
		wantErr []fieldError
	}{
		{name: "one institution", label: `{"capabilities":["connector"],"institutions":["ibkr"]}`, want: []string{"ibkr"}},
		{name: "several", label: `{"capabilities":["connector"],"institutions":["ibkr","ibkr-ie"]}`, want: []string{"ibkr", "ibkr-ie"}},
		{name: "none", label: `{"capabilities":["connector"]}`},
		{
			name: "not a label", label: `{"capabilities":["connector"],"institutions":["IBKR"]}`,
			wantErr: []fieldError{{Field: "institutions[0]", Err: plugins.ErrInvalidName}},
		},
		{
			name: "named twice", label: `{"capabilities":["connector"],"institutions":["xtb","xtb"]}`,
			wantErr: []fieldError{{Field: "institutions[1]", Err: plugins.ErrDuplicate}},
		},
		{
			name: "a view declares one", label: `{"capabilities":["view"],"institutions":["eodhd"]}`,
			wantErr: []fieldError{{Field: "institutions", Err: plugins.ErrInstitutionsNeedConnector}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := plugins.ParseManifestLabel(tt.label)
			if diff := cmp.Diff(tt.wantErr, fieldErrors(err), cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("errors (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, m.Institutions); diff != "" {
				t.Errorf("institutions (-want +got):\n%s", diff)
			}
		})
	}
	a := plugins.Manifest{Capabilities: []plugins.Capability{plugins.CapabilityConnector}, Institutions: []string{"ibkr", "xtb"}}
	b := a
	b.Institutions = []string{"xtb", "ibkr"}
	if !a.Equal(&b) {
		t.Errorf("manifests declaring the same institutions in another order differ")
	}
	b.Institutions = []string{"ibkr", "eodhd"}
	if a.Equal(&b) {
		t.Errorf("manifests declaring other institutions are equal")
	}
}
