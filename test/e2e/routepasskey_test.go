//go:build e2e

package e2e

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

// The second half of the exposure test: the HTTPRoute component's patch of
// the core's auth origins, and a passkey created and used on the route's
// hostname (D-200 criteria 8 and 13, D-60, D-80).
const (
	// gatewayComponentFile is the component, whose patch sets the core's
	// OPENFIEF_AUTH_ORIGINS.
	gatewayComponentFile = "../../deploy/gateway/kustomization.yaml"
	// componentHost is the hostname placeholder the component's route and
	// patch carry, which an overlay replaces with the owner's.
	componentHost = "openfief.invalid"
	// rolloutWait bounds the core's rollout after the patch.
	rolloutWait = 5 * time.Minute
	// holdsCoreDir holds one file per bundle test still running that the
	// route passkey must wait for, since the auth-origin patch restarts the
	// core they talk to; runBundle empties it before the tests start.
	holdsCoreDir = "../../.bin/e2e/holds-core"
	// othersTimeout bounds the wait for the other bundle tests.
	othersTimeout = 15 * time.Minute
	// dirPerm is the mode of holdsCoreDir.
	dirPerm = 0o700
	// hostRulesEnv and trustedKeysEnv tell passkey.mjs where Chromium
	// finds the hostname (its --host-resolver-rules) and which keys it
	// trusts for it (its --ignore-certificate-errors-spki-list): the run's
	// CA and the Gateway's certificate, and no other.
	hostRulesEnv   = "OPENFIEF_E2E_HOST_RULES"
	trustedKeysEnv = "OPENFIEF_E2E_TRUSTED_SPKI"
)

// errNoComponentPatch reports a component without the auth-origin patch.
var errNoComponentPatch = errors.New("deploy/gateway has no patch replacing the core's OPENFIEF_AUTH_ORIGINS")

// holdsCore registers a bundle test that the route passkey must wait for,
// as a file in holdsCoreDir that goes when the test ends. Each calls it
// before t.Parallel, so every registration happens before any parallel test
// resumes, and so before the exposure test waits.
func holdsCore(t *testing.T) {
	t.Helper()
	marker := filepath.Join(holdsCoreDir, t.Name())
	if err := os.WriteFile(marker, nil, filePerm); err != nil {
		t.Fatalf("registering %s with the route passkey: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		if err := os.Remove(marker); err != nil {
			t.Errorf("unregistering %s: %v", t.Name(), err)
		}
	})
}

// resetHoldsCore empties holdsCoreDir before the bundle tests start, so a
// run that crashed earlier leaves nothing to wait for.
func resetHoldsCore() error {
	if err := os.RemoveAll(holdsCoreDir); err != nil {
		return fmt.Errorf("emptying %s: %w", holdsCoreDir, err)
	}
	if err := os.MkdirAll(holdsCoreDir, dirPerm); err != nil {
		return fmt.Errorf("creating %s: %w", holdsCoreDir, err)
	}
	return nil
}

// waitOtherBundleTests waits until every bundle test that called holdsCore
// has ended.
func waitOtherBundleTests(t *testing.T) {
	t.Helper()
	err := waitFor(t.Context(), othersTimeout, func(context.Context) (string, bool, error) {
		left, err := os.ReadDir(holdsCoreDir)
		if err != nil {
			return "", false, fmt.Errorf("reading %s: %w", holdsCoreDir, err)
		}
		names := make([]string, 0, len(left))
		for _, e := range left {
			names = append(names, e.Name())
		}
		return "still running: " + strings.Join(names, ", "), len(left) == 0, nil
	})
	if err != nil {
		t.Fatalf("waiting for the other bundle tests: %v", err)
	}
}

// checkRoutePasskey runs once the other bundle tests have finished: it
// applies the HTTPRoute component's patch to the core's auth origins, with
// the suite's hostname in place of the placeholder as the guide's overlay
// puts it, waits for the core's rollout, and checks the route still serves
// /readyz. Then `openfief enroll --origin https://<hostname>`, the CLI built
// from this checkout, prints a one-time link for the hostname with no
// port-forward, and Chromium, sent to the Gateway's data plane for that
// hostname and trusting only the run's CA and the Gateway's certificate,
// creates a passkey there with its virtual authenticator, keeps the
// __Host- session cookie, signs out and signs in again with it
// (SCEN-e2e-passkey-on-route-hostname).
func checkRoutePasskey(t *testing.T, k *kube, gw *gatewayAccess, certs *gatewayTLSFiles) {
	t.Helper()
	t.Log("waiting for the other bundle tests before the core restarts")
	waitOtherBundleTests(t)
	patchAuthOrigins(t, k)
	waitCoreRolledOut(t, k)
	waitHTTPS(t, gw.client, "/readyz", http.StatusOK, "")

	link := enrollOnRoute(t)
	browser, cancel := context.WithTimeout(t.Context(), browserTimeout)
	defer cancel()
	node := exec.CommandContext(browser, "node", passkeyScript)
	node.Env = append(os.Environ(), linkEnv+"="+link,
		hostRulesEnv+"=MAP "+gatewayHost+" "+gw.addr,
		trustedKeysEnv+"="+strings.Join(certs.keys, ","))
	got, err := node.CombinedOutput()
	t.Logf("%s", got)
	if err != nil {
		t.Errorf("the browser's passkey ceremonies on https://%s: %v", gatewayHost, err)
	}
}

// patchAuthOrigins applies the component's patch of the core Deployment to
// the running core, its hostname placeholder replaced with gatewayHost.
func patchAuthOrigins(t *testing.T, k *kube) {
	t.Helper()
	ops := componentCorePatch(t)
	replaced := false
	for _, op := range ops {
		if value, ok := op["value"].(string); ok && op["op"] == "replace" && strings.Contains(value, "https://"+componentHost) {
			op["value"] = strings.ReplaceAll(value, "https://"+componentHost, "https://"+gatewayHost)
			replaced = true
		}
	}
	if !replaced {
		t.Fatal(errNoComponentPatch)
	}
	patch, err := json.Marshal(ops)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("patching the core's auth origins as deploy/gateway does, for https://%s", gatewayHost)
	_, err = k.clients.AppsV1().Deployments(namespace).Patch(t.Context(), coreDeployment, types.JSONPatchType, patch,
		metav1.PatchOptions{FieldManager: fieldManager})
	if err != nil {
		t.Fatalf("patching the core's auth origins: %v", err)
	}
}

// componentCorePatch returns the JSON patch operations deploy/gateway's
// component applies to the core Deployment.
func componentCorePatch(t *testing.T) []map[string]any {
	t.Helper()
	raw, err := os.ReadFile(gatewayComponentFile)
	if err != nil {
		t.Fatalf("reading the HTTPRoute component: %v", err)
	}
	var component struct {
		Patches []struct {
			Target struct{ Kind, Name string } `json:"target"`
			Patch  string                      `json:"patch"`
		} `json:"patches"`
	}
	if decodeErr := yaml.Unmarshal(raw, &component); decodeErr != nil {
		t.Fatalf("decoding the HTTPRoute component: %v", decodeErr)
	}
	var ops []map[string]any
	for _, p := range component.Patches {
		if p.Target.Kind == "Deployment" && p.Target.Name == coreDeployment {
			if decodeErr := yaml.Unmarshal([]byte(p.Patch), &ops); decodeErr != nil {
				t.Fatalf("decoding the component's patch: %v", decodeErr)
			}
		}
	}
	return ops
}

// waitCoreRolledOut waits until the core Deployment's newest generation
// runs alone: every replica updated and available, and no other core pod
// left, terminating ones included, so the CLI's exec reaches the new one.
func waitCoreRolledOut(t *testing.T, k *kube) {
	t.Helper()
	err := waitFor(t.Context(), rolloutWait, func(ctx context.Context) (string, bool, error) {
		d, err := k.clients.AppsV1().Deployments(namespace).Get(ctx, coreDeployment, metav1.GetOptions{})
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		want := int32(1)
		if d.Spec.Replicas != nil {
			want = *d.Spec.Replicas
		}
		pods, err := k.clients.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=" + coreDeployment})
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		s := d.Status
		state := fmt.Sprintf("generation %d observed %d, %d updated, %d available, %d pods", d.Generation, s.ObservedGeneration,
			s.UpdatedReplicas, s.AvailableReplicas, len(pods.Items))
		done := s.ObservedGeneration >= d.Generation && s.UpdatedReplicas == want && s.AvailableReplicas == want &&
			s.Replicas == want && len(pods.Items) == int(want)
		return state, done, nil
	})
	if err != nil {
		t.Fatalf("waiting for the core's rollout after the auth-origin patch: %v", err)
	}
}

// enrollOnRoute runs `openfief enroll --origin https://gatewayHost` with the
// CLI built from this checkout and returns the one-time link it prints.
func enrollOnRoute(t *testing.T) string {
	t.Helper()
	if _, err := os.Stat(cliPath); err != nil {
		t.Fatalf("%v: %v", errNoCLI, err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), enrollTimeout)
	defer cancel()
	origin := "https://" + gatewayHost
	cmd := exec.CommandContext(ctx, cliPath, "enroll", "--origin", origin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("openfief enroll --origin %s: %v; stderr: %s", origin, err, stderr.String())
	}
	s := bufio.NewScanner(&stdout)
	for s.Scan() {
		if line := strings.TrimSpace(s.Text()); strings.HasPrefix(line, origin+"/enroll#code=") {
			return line
		}
	}
	t.Fatalf("openfief enroll --origin %s printed no link for it; stderr: %s", origin, stderr.String())
	return ""
}

// spkiHash is Chromium's form of a key it trusts for a run: the base64 of
// the SHA-256 of the certificate's SubjectPublicKeyInfo.
func spkiHash(cert *x509.Certificate) string {
	sum := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	return base64.StdEncoding.EncodeToString(sum[:])
}
