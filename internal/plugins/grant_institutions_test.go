package plugins_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// institutionPair is the sample connector declaring institutions, under a
// grant of everything else that approves approved.
func institutionPair(declared, approved []string) (plugins.Plugin, plugins.Grant) {
	p, g := samplePlugin(), fullGrant()
	p.Manifest.Institutions = declared
	g.Permissions.Institutions = approved
	return p, g
}

// TestInstitutionsNeedApproval checks D-185: the institutions a connector
// declares are a permission its grant approves, like its hostnames; the
// connector books only for those both name, and an upgrade's diff shows
// the institutions it adds and drops.
func TestInstitutionsNeedApproval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		declared      []string
		approved      []string
		wantPending   []string
		wantEffective []string
	}{
		{name: "approved", declared: []string{"ibkr"}, approved: []string{"ibkr"}, wantEffective: []string{"ibkr"}},
		{name: "none approved", declared: []string{"ibkr", "xtb"}, wantPending: []string{"ibkr", "xtb"}},
		{name: "one of two approved", declared: []string{"xtb", "ibkr"}, approved: []string{"ibkr"}, wantPending: []string{"xtb"}, wantEffective: []string{"ibkr"}},
		{name: "grant approves more", declared: []string{"ibkr"}, approved: []string{"ibkr", "nbp"}, wantEffective: []string{"ibkr"}},
		{name: "none declared", approved: []string{"ibkr"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, g := institutionPair(tt.declared, tt.approved)
			pending, err := plugins.NeedsApproval(&g, &p)
			if err != nil {
				t.Fatalf("NeedsApproval: %v", err)
			}
			if diff := cmp.Diff(plugins.Permissions{Institutions: tt.wantPending}, pending); diff != "" {
				t.Errorf("NeedsApproval() (-want +got):\n%s", diff)
			}
			e, err := plugins.Effective(&p, &g)
			if err != nil {
				t.Fatalf("Effective: %v", err)
			}
			if diff := cmp.Diff(tt.wantEffective, e.Institutions()); diff != "" {
				t.Errorf("Effective().Institutions() (-want +got):\n%s", diff)
			}
		})
	}
}

// TestInstitutionsInGrantAndDiff checks that a grant's institutions are
// validated like a manifest's, and that an upgrade's diff names the
// institutions it changes.
func TestInstitutionsInGrantAndDiff(t *testing.T) {
	t.Parallel()
	g := fullGrant()
	g.Permissions.Institutions = []string{"IBKR", "nbp", "nbp"}
	if err := g.Validate(); !errors.Is(err, plugins.ErrInvalidName) {
		t.Errorf("Validate() of bad institutions = %v, want a refusal", err)
	}
	from := plugins.Permissions{Institutions: []string{"ibkr"}}
	to := plugins.Permissions{Institutions: []string{"xtb", "ibkr"}}
	d := plugins.DiffPermissions(&from, &to)
	if diff := cmp.Diff(plugins.PermissionsDiff{Added: plugins.Permissions{Institutions: []string{"xtb"}}}, d); diff != "" {
		t.Errorf("DiffPermissions() (-want +got):\n%s", diff)
	}
	if d.IsEmpty() {
		t.Error("a diff adding an institution is empty")
	}
}

// TestApprovedInstitutions checks how a PluginGrant approves institutions:
// the list it names, even an empty one, or, for a grant made before D-185
// that names none, those of the version that runs under it.
func TestApprovedInstitutions(t *testing.T) {
	t.Parallel()
	running, _ := institutionPair([]string{"ibkr"}, nil)
	tests := []struct {
		name     string
		approved *[]string
		deployed *plugins.Plugin
		want     []string
	}{
		{name: "named", approved: &[]string{"xtb"}, deployed: &running, want: []string{"xtb"}},
		{name: "named empty approves none", approved: &[]string{}, deployed: &running, want: []string{}},
		{name: "before D-185, a version runs", deployed: &running, want: []string{"ibkr"}},
		{name: "before D-185, nothing runs"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, plugins.ApprovedInstitutions(tt.approved, tt.deployed)); diff != "" {
				t.Errorf("ApprovedInstitutions() (-want +got):\n%s", diff)
			}
		})
	}
}

// TestSharedInstitutions checks the install's warning (D-185): each
// institution another instance already books for, with that instance,
// never the instance itself.
func TestSharedInstitutions(t *testing.T) {
	t.Parallel()
	holders := []plugins.InstitutionHolder{
		{Instance: "ibkr-main", Institutions: []string{"ibkr"}},
		{Instance: "ibkr-joint", Institutions: []string{"ibkr", "ibkr-ie"}},
		{Instance: "nbp", Institutions: []string{"nbp"}},
		{Instance: "new", Institutions: []string{"ibkr"}},
	}
	got := plugins.SharedInstitutions("new", []string{"ibkr-ie", "ibkr", "xtb"}, holders)
	want := []plugins.SharedInstitution{
		{Institution: "ibkr", Instance: "ibkr-joint"},
		{Institution: "ibkr", Instance: "ibkr-main"},
		{Institution: "ibkr-ie", Instance: "ibkr-joint"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("SharedInstitutions() (-want +got):\n%s", diff)
	}
	if got := plugins.SharedInstitutions("new", []string{"xtb"}, holders); got != nil {
		t.Errorf("SharedInstitutions() for an institution no one books for = %v, want none", got)
	}
}
