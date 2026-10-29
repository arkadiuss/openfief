package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
)

func TestPrepareUpgrade(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := installSamples(t)
	in := installer(t)
	up, err := in.PrepareUpgrade(ctx, store, app.UpgradeRequest{Instance: "sample-a", Version: unpinnedImage})
	if err != nil {
		t.Fatalf("PrepareUpgrade: %v", err)
	}
	if up.From.Image != pinnedImage || up.Plugin.Image != unpinnedImage || up.Plugin.Name != "sample-a" {
		t.Errorf("upgrade from %s to %s as %s, want sample-a from the pinned to the unpinned image", up.From.Image, up.Plugin.Image, up.Plugin.Name)
	}
	if !up.Diff.IsEmpty() || up.Decision.Trust != plugins.Unverified {
		t.Errorf("diff %+v, trust %s; want the same permissions, unverified", up.Diff, up.Decision.Trust)
	}
	if err := app.Upgrade(ctx, store, &up); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
	if got := store.plugins["sample-a"].Image; got != unpinnedImage {
		t.Errorf("image after the upgrade = %s, want %s", got, unpinnedImage)
	}
	if diff := cmp.Diff(up.Plugin.Requested(), store.grants["sample-a"].Permissions); diff != "" {
		t.Errorf("grant after the upgrade (-want +got):\n%s", diff)
	}
}

func TestPrepareUpgradeRefuses(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := installSamples(t)
	in := installer(t)
	refusals := []struct {
		name    string
		req     app.UpgradeRequest
		wantErr error
	}{
		{name: "not installed", req: app.UpgradeRequest{Instance: "nope", Version: pinnedImage}, wantErr: app.ErrNotInstalled},
		{name: "an untrusted image", req: app.UpgradeRequest{Instance: "sample-b", Version: strangerImage}, wantErr: app.ErrNotTrusted},
		{name: "a config the schema refuses", req: app.UpgradeRequest{Instance: "sample-b", Version: pinnedImage, Config: []byte(`{"bad":1}`)}, wantErr: app.ErrConfig},
	}
	for _, tt := range refusals {
		if _, err := in.PrepareUpgrade(ctx, store, tt.req); !errors.Is(err, tt.wantErr) {
			t.Errorf("%s: error = %v, want %v", tt.name, err, tt.wantErr)
		}
	}
}
