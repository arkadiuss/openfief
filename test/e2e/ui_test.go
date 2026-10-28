//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const (
	// uiDeployment is the built-in UI's Deployment, as the operator names a
	// plugin instance's workload (plugin-<name>); overlays/openfief runs a
	// stand-in of it.
	uiDeployment = "plugin-ui"
	// uiTimeout bounds the wait for the UI to be available.
	uiTimeout = 3 * time.Minute
	// loginButton is on the UI's login page.
	loginButton = "Sign in with passkey"
)

// TestUIReady checks that the built-in UI's pod, built from this checkout,
// becomes ready (SCEN-e2e-ui-ready).
func TestUIReady(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitUIAvailable(t, k)
}

// TestTraefikRoutesToUI checks the demo's route to the built-in UI through
// k3s's bundled Traefik and the ui-from-traefik policy: / sends a browser
// without a session to /login, and /login answers 200 with the login page
// (SCEN-e2e-ui-login-through-traefik). WebAuthn needs a browser and is not
// exercised. Only k3s bundles Traefik, so on kind there is nothing to check.
func TestTraefikRoutesToUI(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	skipUnlessK3s(t, k)
	waitUIAvailable(t, k)
	waitTraefik(t, k)
	var root reply
	err := waitFor(t.Context(), traefikTimeout, func(ctx context.Context) (string, bool, error) {
		r, getErr := get(ctx, k, clientPod, traefikURL+"/", "Host: "+routeHost)
		root = r
		return fmt.Sprintf("status %d", r.status), r.status == http.StatusSeeOther, getErr
	})
	if err != nil {
		t.Fatalf("GET / through Traefik: %v", err)
	}
	r, err := get(t.Context(), k, clientPod, traefikURL+"/login", "Host: "+routeHost)
	if err != nil {
		t.Fatal(err)
	}
	if r.status != http.StatusOK || !bytes.Contains(r.body, []byte(loginButton)) {
		t.Fatalf("GET /login through Traefik: status %d, body %q; want 200 with %q (GET / was %d)", r.status, r.body, loginButton, root.status)
	}
}

// waitUIAvailable waits until the UI's Deployment is available, failing at
// once if its pod can't start.
func waitUIAvailable(t *testing.T, k *kube) {
	t.Helper()
	available := k.deploymentAvailable(namespace, uiDeployment)
	err := waitFor(t.Context(), uiTimeout, func(ctx context.Context) (string, bool, error) {
		if state, done, err := available(ctx); done || (err != nil && !errors.Is(err, errRetry)) {
			return state, done, err
		}
		state, err := k.podsState(ctx, namespace, "openfief-plugin")
		return state, false, err
	})
	if err != nil {
		t.Fatalf("waiting for the UI Deployment to be available: %v", err)
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
