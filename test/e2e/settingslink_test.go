//go:build e2e

package e2e

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// The settings page's way to a second hostname (D-200 criterion 8, D-60):
// secondhost.mjs drives it, with the second origin in secondOriginEnv.
const (
	secondHostScript = "secondhost.mjs"
	secondOriginEnv  = "OPENFIEF_E2E_SECOND_ORIGIN"
)

// checkSettingsLinkPasskey runs after checkRoutePasskey, once the core
// accepts https://gatewayHost next to http://localhost:8080. `openfief
// enroll`, as the docs give it (enrollLink), forwards localhost:8080 to the
// core and prints a one-time link for localhost; Chromium, also sent to the
// Gateway's data plane for gatewayHost and trusting only the run's CA and
// the Gateway's certificate, creates a passkey on localhost with it, then,
// signed in there, has the settings page's "Passkey on another host" form
// create a one-time link for https://gatewayHost, confirming with the
// passkey first when the page asks (D-183). On gatewayHost it creates a
// passkey with that link, keeps the __Host- session cookie, signs out and
// signs in again with it, and the settings page there lists passkeys on
// both hosts (SCEN-e2e-passkey-second-host-settings-link).
func checkSettingsLinkPasskey(t *testing.T, k *kube, gw *gatewayAccess, certs *gatewayTLSFiles) {
	t.Helper()
	lockLocalhostForward(t)
	link := enrollLink(t, k)
	waitHTTP(t)
	browser, cancel := context.WithTimeout(t.Context(), browserTimeout)
	defer cancel()
	node := exec.CommandContext(browser, "node", secondHostScript)
	node.Env = append(os.Environ(), linkEnv+"="+link, secondOriginEnv+"=https://"+gatewayHost,
		hostRulesEnv+"=MAP "+gatewayHost+" "+gw.addr,
		trustedKeysEnv+"="+strings.Join(certs.keys, ","))
	got, err := node.CombinedOutput()
	t.Logf("%s", got)
	if err != nil {
		t.Errorf("the browser's passkey on https://%s through the settings page's link: %v", gatewayHost, err)
	}
}
