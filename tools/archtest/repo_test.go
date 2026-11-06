package archtest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arkadiuss/openfief/internal/plugins/catalog"
	"github.com/arkadiuss/openfief/tools/archtest"
)

// TestArchitecture runs every rule on the repository and fails on a
// violation neither baseline.txt nor exceptions.txt lists, and on a listed
// one that no longer happens: the baseline only shrinks.
func TestArchitecture(t *testing.T) {
	t.Parallel()
	repo, err := archtest.Load(&archtest.Config{
		Root:     "../..",
		Module:   "github.com/arkadiuss/openfief",
		Patterns: []string{"./...", "./sdk/go/..."},
		Walk:     []string{"../../internal", "../../cmd", "../../plugins", "../../sdk", "../../tools", "../../test", "../../proto"},
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	c, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	digests := make([]string, 0, len(c.Entries))
	for i := range c.Entries {
		digests = append(digests, c.Entries[i].Image.Digest())
	}
	baseline := readList(t, "baseline.txt")
	exceptions := readList(t, "exceptions.txt")
	fresh, stale := archtest.Compare(archtest.Check(repo, digests), baseline, exceptions)
	for _, v := range fresh {
		t.Errorf("%s\n  fix the code; the baseline only shrinks, and is never edited to get green", v)
	}
	for _, e := range stale {
		t.Errorf("[%s] %s no longer happens: delete line %d of the baseline or exceptions (it only shrinks)", e.Rule, e.Key, e.Line)
	}
}

// readList reads one of this package's lists.
func readList(t *testing.T, name string) []archtest.Entry {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean(name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	entries, err := archtest.ParseList(string(data))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return entries
}
