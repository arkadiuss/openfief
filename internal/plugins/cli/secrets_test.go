package cli_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/plugins/app"
)

// TestRunRefusesSharedSecret checks D-89 and D-182 on the command line:
// with instance sample-x naming the Secret sample-key, as a role x-key or
// a hand-written grant would, installing sampleImage as sample, whose role
// key names sample-key too, stops before the owner is asked, and nothing
// is installed.
func TestRunRefusesSharedSecret(t *testing.T) {
	t.Parallel()
	c := installed()
	c.instances = append(c.instances, app.Instance{Name: "sample-x", HasPlugin: true, HasGrant: true, Secrets: []string{"sample-key"}})
	out, err := runOn(c, "y\n", "plugin", "install", sampleImage)
	if !errors.Is(err, app.ErrSecretInUse) || len(c.instances) != 1 {
		t.Errorf("install naming sample-x's Secret: %v, instances %v; want ErrSecretInUse and nothing installed\n%s", err, c.instances, out)
	}
	if strings.Contains(out, "Install sample?") {
		t.Errorf("the owner was asked to confirm an install that can't go through:\n%s", out)
	}
}
