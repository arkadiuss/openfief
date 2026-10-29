//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"slices"
	"testing"
)

const (
	// uiPlugin is the demo's instance of the built-in UI
	// (deploy/demo/plugins/ui.yaml), which the operator deploys from the
	// suite's signed image (editApp).
	uiPlugin = "ui"
	// uiImage is the UI's image in the suite's registry.
	uiImage = "ui"
	// uiBrowserPath is the browser path its manifest declares and its grant
	// approves (D-81).
	uiBrowserPath = "/"
	// loginButton is on the UI's login page.
	loginButton = "Sign in with passkey"
)

// TestUIReady checks that the operator deploys the built-in UI, built from
// this checkout and signed with the suite's key, and runs it with its
// browser path / granted (SCEN-e2e-ui-ready).
func TestUIReady(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	p := waitPluginReady(t, k, uiPlugin)
	if p.Status.Effective == nil || !slices.Equal(p.Status.Effective.BrowserPaths, []string{uiBrowserPath}) {
		t.Errorf("plugin %s effective permissions = %+v, want the browser path %s", uiPlugin, p.Status.Effective, uiBrowserPath)
	}
}

// TestCoreProxiesUI checks that a browser reaches the built-in UI through
// the core, the only entry point for browsers (D-81): from the client pod,
// which only the core admits, / sends a browser without a session to /login,
// and /login answers 200 with the login page, which the core proxied to the
// UI with its own token (SCEN-e2e-ui-through-core).
func TestCoreProxiesUI(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitCoreAvailable(t, k)
	waitPluginReady(t, k, uiPlugin)
	checkLogin(t, k, coreURL)
}

// TestTraefikRoutesToUI checks the demo's route to the built-in UI through
// k3s's bundled Traefik: Traefik sends every path to the core, and the core
// proxies the UI's (SCEN-e2e-ui-login-through-traefik). Only k3s bundles
// Traefik, so on kind there is nothing to check.
func TestTraefikRoutesToUI(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	skipUnlessK3s(t, k)
	waitPluginReady(t, k, uiPlugin)
	waitTraefik(t, k)
	checkLogin(t, k, traefikURL, "Host: "+routeHost)
}

// checkLogin checks, from the client pod, that base/ sends a browser
// without a session to /login and that base/login answers 200 with the
// login page. The core reads the UI's route from its Plugin, so the first
// answers may still be 503 or 404 while it catches up.
func checkLogin(t *testing.T, k *kube, base string, headers ...string) {
	t.Helper()
	var root reply
	err := waitFor(t.Context(), traefikTimeout, func(ctx context.Context) (string, bool, error) {
		r, getErr := get(ctx, k, clientPod, base+"/", headers...)
		root = r
		return fmt.Sprintf("status %d", r.status), r.status == http.StatusSeeOther, getErr
	})
	if err != nil {
		t.Fatalf("GET %s/: %v", base, err)
	}
	r, err := get(t.Context(), k, clientPod, base+"/login", headers...)
	if err != nil {
		t.Fatal(err)
	}
	if r.status != http.StatusOK || !bytes.Contains(r.body, []byte(loginButton)) {
		t.Fatalf("GET %s/login: status %d, body %q; want 200 with %q (GET / was %d)", base, r.status, r.body, loginButton, root.status)
	}
}

// skipUnlessK3s skips t on a cluster without k3s's bundled Traefik.
func skipUnlessK3s(t *testing.T, k *kube) {
	t.Helper()
	isK3s, err := k.isK3s(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !isK3s {
		t.Skip("not k3s: no bundled Traefik; the demo's routes are checked on k3d")
	}
}

// waitTraefik waits until k3s's Traefik is available.
func waitTraefik(t *testing.T, k *kube) {
	t.Helper()
	if err := waitFor(t.Context(), traefikTimeout, k.deploymentAvailable("kube-system", "traefik")); err != nil {
		t.Fatalf("waiting for Traefik: %v", err)
	}
}
