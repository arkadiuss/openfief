package authz_test

import (
	"os"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/authz"
)

// TestSchemaIsTheFile checks that the embedded schema is schema.zed as it
// stands, the file `task authz:validate` checks, and holds the model's types.
func TestSchemaIsTheFile(t *testing.T) {
	t.Parallel()
	file, err := os.ReadFile("schema.zed")
	if err != nil {
		t.Fatalf("reading schema.zed: %v", err)
	}
	if authz.Schema != string(file) {
		t.Error("authz.Schema differs from schema.zed")
	}
	for _, def := range []string{"definition household", "definition sleeve", "definition account"} {
		if !strings.Contains(authz.Schema, def) {
			t.Errorf("the schema has no %q", def)
		}
	}
}
