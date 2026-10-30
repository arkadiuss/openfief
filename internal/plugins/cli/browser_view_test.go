package cli_test

import (
	"errors"
	"testing"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
)

// galleryImage is a synthetic view plugin, signed by OpenFief's CI, that
// declares the browser path /gallery.
const galleryImage = "registry.example.com/example/gallery@sha256:c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3"

// withUI is a cluster where the instance ui holds the browser path /.
func withUI() *fakeCluster {
	c := installed()
	c.instances = append(c.instances, app.Instance{Name: "ui", HasPlugin: true, HasGrant: true, BrowserPath: "/"})
	return c
}

// TestRunRefusesSecondBrowserView checks D-114 on the command line: with the
// built-in UI holding a browser path, neither installing nor upgrading to a
// version that declares one gets as far as the owner's confirmation, and
// nothing changes; with no view installed, the same install goes through.
func TestRunRefusesSecondBrowserView(t *testing.T) {
	t.Parallel()
	c := withUI()
	out, err := runOn(c, "y\n", "plugin", "install", galleryImage)
	if !errors.Is(err, plugins.ErrBrowserPathHeld) || len(c.instances) != 1 {
		t.Errorf("install a second view: %v, instances %v; want ErrBrowserPathHeld and nothing installed\n%s", err, c.instances, out)
	}
	if _, err := runOn(c, "", "plugin", "install", sampleImage, "--yes"); err != nil {
		t.Fatalf("install sample, no browser path: %v", err)
	}
	if out, err := runOn(c, "y\n", "plugin", "upgrade", "sample", galleryImage); !errors.Is(err, plugins.ErrBrowserPathHeld) || c.plugins["sample"].Image != sampleImage {
		t.Errorf("upgrade to a view: %v, image %s; want ErrBrowserPathHeld and no change\n%s", err, c.plugins["sample"].Image, out)
	}
	alone := installed()
	if out, err := runOn(alone, "", "plugin", "install", galleryImage, "--yes"); err != nil {
		t.Errorf("install the only view: %v\n%s", err, out)
	}
}
