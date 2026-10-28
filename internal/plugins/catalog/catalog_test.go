package catalog_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/catalog"
)

func TestLoad(t *testing.T) {
	t.Parallel()
	c, err := catalog.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, err := plugins.NewTrustPolicy(c, nil); err != nil {
		t.Errorf("the embedded catalog does not make a trust policy: %v", err)
	}
	seen := map[string]bool{}
	for _, e := range c.Entries {
		if seen[e.Image.Digest()] {
			t.Errorf("digest %s is pinned twice", e.Image.Digest())
		}
		seen[e.Image.Digest()] = true
	}
}
