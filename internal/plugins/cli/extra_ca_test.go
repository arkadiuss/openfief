package cli_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// caImage is a synthetic connector, signed by OpenFief's CI, whose manifest
// asks for the optional CA bundle Secret role.
const caImage = "registry.example.com/example/rates@sha256:d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4"

// TestRunInstallExtraCA checks that --extra-ca makes the installed Plugin
// expect an extra CA bundle and the summary say what follows from it, that
// without it the summary says the Secret isn't mounted while the grant
// still approves it, and that a manifest without the CA bundle role refuses
// the flag.
func TestRunInstallExtraCA(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		args        []string
		wantExtraCA bool
		wantOut     string
		wantErr     error
	}{
		{
			name: "expected", args: []string{"plugin", "install", caImage, "--name", "rates", "--extra-ca", "--yes"},
			wantExtraCA: true, wantOut: "CABundleMissing condition while it doesn't exist, and restarts the plugin when it changes",
		},
		{
			name: "not expected", args: []string{"plugin", "install", caImage, "--name", "rates", "--yes"},
			wantOut: "not mounted: install with --extra-ca to trust extra CAs from it",
		},
		{
			name: "a manifest without the role", args: []string{"plugin", "install", sampleImage, "--name", "view", "--extra-ca", "--yes"},
			wantErr: plugins.ErrExtraCANotRequested,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := installed()
			out, err := runOn(c, "", tt.args...)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Run: %v, want %v\n%s", err, tt.wantErr, out)
			}
			checkExtraCAInstall(t, c, out, tt.wantErr == nil, tt.wantExtraCA, tt.wantOut)
		})
	}
}

// checkExtraCAInstall checks what an install of instance rates left in c and
// printed: nothing on a refusal; otherwise a Plugin expecting a CA bundle
// when wantExtraCA, a grant that approves the bundle's Secret either way,
// and wantOut in out.
func checkExtraCAInstall(t *testing.T, c *fakeCluster, out string, installed, wantExtraCA bool, wantOut string) {
	t.Helper()
	if !installed {
		if len(c.plugins) != 0 {
			t.Errorf("installed %v after an error", c.plugins)
		}
		return
	}
	if !strings.Contains(out, wantOut) {
		t.Errorf("output without %q:\n%s", wantOut, out)
	}
	if p := c.plugins["rates"]; p.ExtraCA != wantExtraCA {
		t.Errorf("Plugin ExtraCA = %v, want %v", p.ExtraCA, wantExtraCA)
	}
	if g := c.grants["rates"]; !slices.Equal(g.Permissions.Secrets, []string{"rates--ca-bundle"}) {
		t.Errorf("grant Secrets %v, want rates--ca-bundle either way", g.Permissions.Secrets)
	}
}
