package cli_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/cli"
)

// sampleV2 is a synthetic next version of sample, signed by OpenFief's CI,
// that also asks for egress to api.example.com.
const sampleV2 = "registry.example.com/example/sample@sha256:b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2"

// runOn runs the command line once against c with the owner's input in.
func runOn(c *fakeCluster, in string, args ...string) (string, error) {
	var out, errOut bytes.Buffer
	if err := cli.Run(context.Background(), args, env(c, in, &out, &errOut)); err != nil {
		return out.String() + errOut.String(), fmt.Errorf("run: %w", err)
	}
	return out.String() + errOut.String(), nil
}

func TestRunUpgrade(t *testing.T) {
	t.Parallel()
	c := installed()
	if _, err := runOn(c, "", "plugin", "install", sampleImage, "--yes"); err != nil {
		t.Fatalf("install: %v", err)
	}
	out, err := runOn(c, "n\n", "plugin", "upgrade", "sample", sampleV2)
	if !errors.Is(err, cli.ErrAborted) || c.plugins["sample"].Image != sampleImage {
		t.Fatalf("declined upgrade: %v, image %s; want nothing changed", err, c.plugins["sample"].Image)
	}
	for _, s := range []string{
		"From:         " + sampleImage, "To:           " + sampleV2,
		"Its requested permissions change", "  + egress host   api.example.com:443", "Upgrade sample? [y/N] ",
	} {
		if !strings.Contains(out, s) {
			t.Errorf("summary lacks %q:\n%s", s, out)
		}
	}
	if out, err = runOn(c, "y\n", "plugin", "upgrade", "sample", sampleV2); err != nil || !strings.Contains(out, "Upgraded sample.") {
		t.Fatalf("confirmed upgrade: %v\n%s", err, out)
	}
	want := plugins.Permissions{
		Capabilities: []plugins.Capability{plugins.CapabilityView},
		Egress:       []plugins.EgressHost{{Host: "api.example.com", Port: 443}},
		Secrets:      []string{"sample-key"},
	}
	if diff := cmp.Diff(want, c.grants["sample"].Permissions); diff != "" || c.plugins["sample"].Image != sampleV2 {
		t.Errorf("after the upgrade image %s, grant (-want +got):\n%s", c.plugins["sample"].Image, diff)
	}
	checkSameVersion(t, c)
}

// checkSameVersion asks to upgrade sample to the version it runs, and
// declines: the summary says nothing changes.
func checkSameVersion(t *testing.T, c *fakeCluster) {
	t.Helper()
	out, err := runOn(c, "", "plugin", "upgrade", "sample", sampleV2)
	if !errors.Is(err, cli.ErrAborted) || !strings.Contains(out, "It requests the same permissions as the installed version.") {
		t.Errorf("same version again: %v\n%s", err, out)
	}
}

func TestRunUpgradeRefuses(t *testing.T) {
	t.Parallel()
	c := installed()
	if _, err := runOn(c, "", "plugin", "install", sampleImage, "--yes"); err != nil {
		t.Fatalf("install: %v", err)
	}
	tests := []struct {
		name    string
		args    []string
		wantErr error
	}{
		{name: "an instance that isn't installed", args: []string{"plugin", "upgrade", "nope", sampleV2, "--yes"}, wantErr: app.ErrNotInstalled},
		{name: "a tag", args: []string{"plugin", "upgrade", "sample", "registry.example.com/example/sample:2", "--yes"}, wantErr: plugins.ErrImageNotPinned},
		{name: "a version the catalog doesn't have", args: []string{"plugin", "upgrade", "sample", "2026.11.0", "--yes"}, wantErr: plugins.ErrNoUpgradeVersion},
		{name: "installing over the instance", args: []string{"plugin", "install", sampleV2, "--yes"}, wantErr: app.ErrInstanceExists},
		{name: "no version", args: []string{"plugin", "upgrade", "sample"}, wantErr: cli.ErrUsage},
	}
	for _, tt := range tests {
		if out, err := runOn(c, "", tt.args...); !errors.Is(err, tt.wantErr) {
			t.Errorf("%s: error = %v, want %v\n%s", tt.name, err, tt.wantErr, out)
		}
	}
	if c.plugins["sample"].Image != sampleImage {
		t.Errorf("a refused upgrade changed the image to %s", c.plugins["sample"].Image)
	}
}
