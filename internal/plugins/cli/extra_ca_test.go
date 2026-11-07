package cli_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/cli"
)

// caImage is a synthetic connector, signed by OpenFief's CI, whose manifest
// asks for the optional CA bundle Secret role.
const caImage = "registry.example.com/example/rates@sha256:d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4d4"

// TestRunInstallExtraCA checks that --extra-ca makes the installed Plugin
// expect an extra CA bundle and the summary say how to set it, that
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
			wantExtraCA: true, wantOut: "set it with `openfief plugin ca-bundle set rates <pem-file>`",
		},
		{
			name: "not expected", args: []string{"plugin", "install", caImage, "--name", "rates", "--yes"},
			wantOut: "not mounted: install or upgrade with --extra-ca to trust extra CAs from it",
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

// testCertPEM is a self-signed CA certificate made for the test, as PEM:
// no key or certificate is checked in.
func testCertPEM(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "ca.example.com"},
		NotBefore: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), NotAfter: time.Date(2036, 1, 1, 0, 0, 0, 0, time.UTC),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("certificate: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

// runWithFiles runs the command line once against c, as runOn does, with
// files added to what it may read.
func runWithFiles(c *fakeCluster, files map[string][]byte, in string, args ...string) (string, error) {
	var out, errOut bytes.Buffer
	e := env(c, in, &out, &errOut)
	read := e.ReadFile
	e.ReadFile = func(path string) ([]byte, error) {
		if data, ok := files[path]; ok {
			return data, nil
		}
		return read(path)
	}
	if err := cli.Run(context.Background(), args, e); err != nil {
		return out.String() + errOut.String(), fmt.Errorf("run: %w", err)
	}
	return out.String() + errOut.String(), nil
}

// caFixture is a cluster where rates expects a CA bundle and view, an
// instance of a manifest without the role, doesn't.
func caFixture(t *testing.T) *fakeCluster {
	t.Helper()
	c := installed()
	for _, args := range [][]string{
		{"plugin", "install", caImage, "--name", "rates", "--extra-ca", "--yes"},
		{"plugin", "install", sampleImage, "--name", "view", "--yes"},
	} {
		if out, err := runOn(c, "", args...); err != nil {
			t.Fatalf("install: %v\n%s", err, out)
		}
	}
	return c
}

// caFiles are the files a run may read: a CA certificate, a private key
// and text.
func caFiles(t *testing.T) map[string][]byte {
	t.Helper()
	return map[string][]byte{
		"ca.pem":   testCertPEM(t),
		"key.pem":  pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("synthetic")}),
		"note.txt": []byte("not a certificate"),
	}
}

// TestRunCABundleSet checks that `ca-bundle set` writes the Secret and then
// stamps its checksum on the Plugin; that a file that isn't PEM
// certificates, an instance that doesn't expect a bundle and one that isn't
// installed are refused before anything changes; and that --print needs no
// cluster, writes the Secret and the stamp, and says to keep them in step.
func TestRunCABundleSet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		args      []string
		noCluster bool
		wantErr   error
		wantOut   []string
		// wantWrites are the cluster writes, in order.
		wantWrites []string
	}{
		{
			name: "set", args: []string{"plugin", "ca-bundle", "set", "rates", "ca.pem"},
			wantOut:    []string{"Set rates--ca-bundle: 1 certificate(s), sha256 ", "stamped on Plugin rates as spec.caBundleSha256"},
			wantWrites: []string{"secret rates", "stamp rates"},
		},
		{name: "a private key", args: []string{"plugin", "ca-bundle", "set", "rates", "key.pem"}, wantErr: plugins.ErrCABundleNotCertificate},
		{name: "not PEM", args: []string{"plugin", "ca-bundle", "set", "rates", "note.txt"}, wantErr: plugins.ErrCABundleNotPEM},
		{name: "no such file", args: []string{"plugin", "ca-bundle", "set", "rates", "nope.pem"}, wantErr: fs.ErrNotExist},
		{name: "an instance expecting none", args: []string{"plugin", "ca-bundle", "set", "view", "ca.pem"}, wantErr: plugins.ErrExtraCANotExpected},
		{name: "an instance not installed", args: []string{"plugin", "ca-bundle", "set", "nope", "ca.pem"}, wantErr: app.ErrNotInstalled},
		{
			name: "--print needs no cluster", args: []string{"plugin", "ca-bundle", "set", "rates", "ca.pem", "--print"}, noCluster: true,
			wantOut: []string{"kind: Secret\nname: rates--ca-bundle\ncaBundleSha256: ", "Keep the checksum in step with the Secret", "the operator never reads the Secret"},
		},
		{name: "--print still checks the file", args: []string{"plugin", "ca-bundle", "set", "rates", "key.pem", "--print"}, noCluster: true, wantErr: plugins.ErrCABundleNotCertificate},
		{name: "no file", args: []string{"plugin", "ca-bundle", "set", "rates"}, wantErr: cli.ErrUsage},
		{name: "no subcommand", args: []string{"plugin", "ca-bundle", "rates", "ca.pem"}, wantErr: cli.ErrUsage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, files := caFixture(t), caFiles(t)
			target := c
			if tt.noCluster {
				target = nil
			}
			out, err := runWithFiles(target, files, "", tt.args...)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Run: %v, want %v\n%s", err, tt.wantErr, out)
			}
			checkCABundleSet(t, c, files["ca.pem"], out, tt.wantOut, tt.wantWrites)
		})
	}
}

// checkCABundleSet checks what a `ca-bundle set` of rates left in c and
// printed: wantOut in out, the cluster writes wantWrites in order and,
// when there were any, the Secret holding the PEM file bundle and the
// Plugin stamped with its checksum; otherwise no stamp.
func checkCABundleSet(t *testing.T, c *fakeCluster, bundle []byte, out string, wantOut, wantWrites []string) {
	t.Helper()
	for _, s := range wantOut {
		if !strings.Contains(out, s) {
			t.Errorf("output without %q:\n%s", s, out)
		}
	}
	if diff := cmp.Diff(wantWrites, c.caWrites); diff != "" {
		t.Errorf("cluster writes (-want +got):\n%s", diff)
	}
	wantStamp := ""
	if wantWrites != nil {
		wantStamp = plugins.CABundleChecksum(bundle)
		if !bytes.Equal(c.caBundles["rates"], bundle) {
			t.Errorf("Secret holds %q, want the PEM file", c.caBundles["rates"])
		}
	}
	if got := c.plugins["rates"].CABundleSHA256; got != wantStamp {
		t.Errorf("stamp %q, want %q", got, wantStamp)
	}
}

// TestRunCABundleUnset checks that `ca-bundle unset` asks first, then
// deletes the Secret and clears the stamp, and that declining changes
// nothing.
func TestRunCABundleUnset(t *testing.T) {
	t.Parallel()
	c, files := caFixture(t), caFiles(t)
	if out, err := runWithFiles(c, files, "", "plugin", "ca-bundle", "set", "rates", "ca.pem"); err != nil {
		t.Fatalf("set: %v\n%s", err, out)
	}
	c.caWrites = nil
	if out, err := runOn(c, "n\n", "plugin", "ca-bundle", "unset", "rates"); !errors.Is(err, cli.ErrAborted) || c.caWrites != nil {
		t.Errorf("declined unset: %v, writes %v; want nothing changed\n%s", err, c.caWrites, out)
	}
	out, err := runOn(c, "y\n", "plugin", "ca-bundle", "unset", "rates")
	if err != nil || !strings.Contains(out, "Deletes Secret rates--ca-bundle and clears spec.caBundleSha256") {
		t.Fatalf("unset: %v\n%s", err, out)
	}
	if diff := cmp.Diff([]string{"delete rates", "stamp rates"}, c.caWrites); diff != "" {
		t.Errorf("cluster writes (-want +got):\n%s", diff)
	}
	if _, ok := c.caBundles["rates"]; ok || c.plugins["rates"].CABundleSHA256 != "" {
		t.Errorf("after unset: Secret kept %v, stamp %q; want both gone", ok, c.plugins["rates"].CABundleSHA256)
	}
	if _, err := runOn(c, "", "plugin", "ca-bundle", "unset", "nope", "--yes"); !errors.Is(err, app.ErrNotInstalled) {
		t.Errorf("unset of an instance not installed: %v, want %v", err, app.ErrNotInstalled)
	}
}

// TestRunUpgradeExtraCA checks that `upgrade --extra-ca` makes an installed
// instance expect a bundle without reinstalling, and says how to set it;
// that a plain upgrade keeps it and its stamp; that --no-extra-ca stops
// expecting it and clears the stamp; and that --extra-ca for a manifest
// without the role, or both flags at once, is refused.
func TestRunUpgradeExtraCA(t *testing.T) {
	t.Parallel()
	c, files := caFixture(t), caFiles(t)
	stamp := plugins.CABundleChecksum(files["ca.pem"])
	steps := []struct {
		name string
		args []string
		// wantErr, then whether rates expects a bundle, its stamp and what
		// the summary says.
		wantErr     error
		wantExtraCA bool
		wantStamp   string
		wantOut     string
	}{
		{
			name: "--no-extra-ca", args: []string{"plugin", "upgrade", "rates", caImage, "--no-extra-ca", "--yes"},
			wantOut: "Extra CA:     no longer expected",
		},
		{
			name: "--extra-ca", args: []string{"plugin", "upgrade", "rates", caImage, "--extra-ca", "--yes"}, wantExtraCA: true,
			wantOut: "Extra CA:     expected from now on: set it with `openfief plugin ca-bundle set rates <pem-file>`",
		},
		{
			name: "set", args: []string{"plugin", "ca-bundle", "set", "rates", "ca.pem"}, wantExtraCA: true, wantStamp: stamp,
			wantOut: "stamped on Plugin rates",
		},
		{
			name: "a plain upgrade keeps both", args: []string{"plugin", "upgrade", "rates", caImage, "--yes"}, wantExtraCA: true, wantStamp: stamp,
			wantOut: "Extra CA:     expected, set (sha256 " + stamp + ")",
		},
		{
			name: "--no-extra-ca clears the stamp", args: []string{"plugin", "upgrade", "rates", caImage, "--no-extra-ca", "--yes"},
			wantOut: "Upgraded rates.",
		},
		{
			name: "--extra-ca without the role", args: []string{"plugin", "upgrade", "view", sampleImage, "--extra-ca", "--yes"},
			wantErr: plugins.ErrExtraCANotRequested,
		},
		{
			name: "both flags", args: []string{"plugin", "upgrade", "rates", caImage, "--extra-ca", "--no-extra-ca", "--yes"},
			wantErr: cli.ErrUsage,
		},
	}
	for _, st := range steps {
		out, err := runWithFiles(c, files, "", st.args...)
		got := c.plugins["rates"]
		if !errors.Is(err, st.wantErr) || got.ExtraCA != st.wantExtraCA || got.CABundleSHA256 != st.wantStamp || !strings.Contains(out, st.wantOut) {
			t.Errorf("%s: %v, ExtraCA %v, stamp %q; want %v, %v, %q and %q in:\n%s",
				st.name, err, got.ExtraCA, got.CABundleSHA256, st.wantErr, st.wantExtraCA, st.wantStamp, st.wantOut, out)
		}
	}
}
