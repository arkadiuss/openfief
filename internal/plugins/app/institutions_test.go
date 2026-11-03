package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
)

// TestSharedInstitutions checks the install's and upgrade's warning
// (D-185): the institutions a plugin books for that another installed
// instance already books for, with that instance; none for a plugin that
// declares none, and never the instance itself.
func TestSharedInstitutions(t *testing.T) {
	t.Parallel()
	store := &fakeStore{instances: []app.Instance{
		{Name: "ibkr-main", Institutions: []string{"ibkr"}},
		{Name: "ibkr-joint", Institutions: []string{"ibkr"}},
		{Name: "nbp", Institutions: []string{"nbp"}},
	}}
	p := plugins.Plugin{Name: "ibkr-main", Manifest: plugins.Manifest{Institutions: []string{"ibkr", "ibkr-ie"}}}
	got, err := app.SharedInstitutions(context.Background(), store, &p)
	if err != nil {
		t.Fatalf("SharedInstitutions: %v", err)
	}
	if diff := cmp.Diff([]plugins.SharedInstitution{{Institution: "ibkr", Instance: "ibkr-joint"}}, got); diff != "" {
		t.Errorf("SharedInstitutions() (-want +got):\n%s", diff)
	}
	p.Manifest.Institutions = nil
	if got, err := app.SharedInstitutions(context.Background(), store, &p); err != nil || got != nil {
		t.Errorf("SharedInstitutions() of a plugin that declares none = %v, %v; want none", got, err)
	}
}
