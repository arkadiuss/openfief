//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// The openfief CLI the bundle run builds from this checkout (buildCLI), as
// the release builds it, and the files that build reads, all relative to
// this package and kept in the gitignored .bin.
const (
	cliPath = "../../.bin/e2e/openfief"
	// cliCatalog is the run's catalog, which the CLI embeds instead of
	// catalogFile, and cliOverlay the go build overlay that says so.
	cliCatalog  = "../../.bin/e2e/catalog.json"
	cliOverlay  = "../../.bin/e2e/catalog-overlay.json"
	catalogFile = "../../internal/plugins/catalog/plugins.json"
	cliPackage  = "../../cmd/openfief"
	// cliTimeout bounds building the CLI and `openfief plugin install`.
	cliTimeout = 5 * time.Minute
	// cliStopDelay bounds how long `openfief enroll` may take to stop
	// after Ctrl-C.
	cliStopDelay = 10 * time.Second
	// installedUI is what `openfief plugin install` prints once it created
	// the UI's Plugin and PluginGrant.
	installedUI = "Installed " + uiPlugin + "."
	// filePerm is the mode of the files buildCLI writes.
	filePerm = 0o600
)

// errNoCLI reports a run without the CLI built from this checkout.
var errNoCLI = errors.New(cliPath + " is missing; the bundle run's TestMain builds it")

// buildCLI builds the openfief CLI from this checkout into cliPath the way
// the release workflow builds it from its release commit (D-88): with a
// catalog that pins the built-in UI's image in the suite's registry, here
// at the digest the suite pushed and as the cluster names it, to its
// publisher, here the suite's key. The checkout's own catalog stays
// untouched: go build reads the run's through an overlay. So `openfief
// plugin install ui` works as the install guide gives it.
func buildCLI(ctx context.Context, reg suiteRegistry, signer suiteSigner, logf func(string, ...any)) error {
	ui, err := reg.image(ctx, uiImage)
	if err != nil {
		return err
	}
	catalog, err := json.Marshal(map[string]any{"plugins": []any{map[string]any{
		"name": uiPlugin, "image": ui.cluster.String(),
		"publisher": map[string]string{"name": signerName, "publicKey": signer.publicKey},
	}}})
	if err != nil {
		return fmt.Errorf("the run's catalog: %w", err)
	}
	if _, parseErr := plugins.ParseCatalog(catalog); parseErr != nil {
		return fmt.Errorf("the run's catalog: %w", parseErr)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{catalogFile: cliCatalog}})
	if err != nil {
		return fmt.Errorf("the build overlay: %w", err)
	}
	if err := errors.Join(os.WriteFile(cliCatalog, catalog, filePerm), os.WriteFile(cliOverlay, overlay, filePerm)); err != nil {
		return fmt.Errorf("writing the run's catalog: %w", err)
	}
	logf("e2e: building the openfief CLI from this checkout into %s, its catalog pinning %s", cliPath, ui.cluster)
	ctx, cancel := context.WithTimeout(ctx, cliTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-trimpath", "-overlay", cliOverlay, "-o", cliPath, cliPackage)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if _, err := build(cmd); err != nil {
		return err
	}
	return nil
}

// TestBundleFirstRun follows the owner's first run on a bundle install
// (docs/guides/install.md, steps 6 and 7) with the openfief CLI built from
// this checkout, run as its own process with the kubeconfig the suite uses:
// `openfief plugin install ui` finds the built-in UI in the CLI's catalog,
// verifies its image in the suite's registry against the catalog's
// publisher, shows what its signed manifest requests and, once answered
// yes, creates its Plugin and PluginGrant; the operator then runs it with
// the browser path / (D-81). `openfief enroll` forwards localhost:8080 to
// the core and prints the first passkey's one-time link (D-80), and
// Chromium, with its virtual authenticator, creates the passkey through the
// core and the UI, keeps the __Host- session cookie, signs out and signs in
// again (D-85); Ctrl-C then stops the forward cleanly
// (SCEN-e2e-bundle-first-run). Only downloading the CLI stays outside the
// suite (docs/dev/testing.md).
func TestBundleFirstRun(t *testing.T) {
	holdsCore(t)
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	if _, err := os.Stat(cliPath); err != nil {
		t.Fatalf("%v: %v", errNoCLI, err)
	}
	waitCoreAvailable(t, k)

	out := installUIWithCLI(t)
	if !strings.Contains(out, "Install "+uiPlugin+"? [y/N]") || !strings.Contains(out, installedUI) {
		t.Fatalf("openfief plugin install printed:\n%s\nwant its confirmation and %q", out, installedUI)
	}
	p := waitPluginReady(t, k, uiPlugin)
	if p.Status.Effective == nil || !slices.Equal(p.Status.Effective.BrowserPaths, []string{uiBrowserPath}) {
		t.Errorf("plugin %s effective permissions = %+v, want the browser path %s", uiPlugin, p.Status.Effective, uiBrowserPath)
	}

	enrollWithCLI(t)
}

// installUIWithCLI runs `openfief plugin install ui` with the CLI built from
// this checkout, answering its question yes, and returns what it printed;
// it fails t when the CLI fails.
func installUIWithCLI(t *testing.T) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), cliTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, cliPath, "plugin", "install", uiPlugin)
	cmd.Stdin = strings.NewReader("y\n")
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("openfief plugin install %s: %v\n%s", uiPlugin, err, out.String())
	}
	return out.String()
}

// enrollWithCLI runs `openfief enroll` with the CLI built from this
// checkout until the browser has created the first passkey and signed in
// with it, then stops it as the owner does, with Ctrl-C, and checks it
// exited cleanly.
func enrollWithCLI(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	cmd := exec.CommandContext(ctx, cliPath, "enroll")
	cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
	cmd.WaitDelay = cliStopDelay
	// Its output goes through a pipe that readLink drains to the end, so
	// Wait, which waits for the copy into it, never races a reader.
	stdout, lines := io.Pipe()
	cmd.Stdout = lines
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting openfief enroll: %v", err)
	}
	wait := func() error {
		err := cmd.Wait()
		return errors.Join(err, lines.Close())
	}
	link, err := readLink(ctx, stdout)
	if err != nil {
		cancel()
		t.Fatalf("%v (openfief enroll: %v; stderr: %s)", err, wait(), stderr.String())
	}
	waitHTTP(t)
	browser, bcancel := context.WithTimeout(ctx, browserTimeout)
	defer bcancel()
	node := exec.CommandContext(browser, "node", passkeyScript)
	node.Env = append(os.Environ(), linkEnv+"="+link)
	got, err := node.CombinedOutput()
	t.Logf("%s", got)
	if err != nil {
		t.Errorf("the browser's passkey ceremonies: %v", err)
	}
	cancel()
	// Wait reports the cancelled context even when the CLI exits cleanly on
	// Ctrl-C, so its exit status decides.
	if err := wait(); err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("openfief enroll after Ctrl-C: %v; stderr: %s", err, stderr.String())
	}
	if !cmd.ProcessState.Success() {
		t.Errorf("openfief enroll exited with %s after Ctrl-C; stderr: %s", cmd.ProcessState, stderr.String())
	}
}
