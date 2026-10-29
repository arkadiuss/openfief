//go:build e2e

package e2e

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	identitycli "github.com/arkadiuss/openfief/internal/identity/cli"
	identitykube "github.com/arkadiuss/openfief/internal/identity/kube"
)

const (
	// enrollTimeout bounds `openfief enroll` until it prints the link, and
	// the wait for the port-forward to answer.
	enrollTimeout = 2 * time.Minute
	// browserTimeout bounds the browser's passkey ceremonies.
	browserTimeout = 3 * time.Minute
	// passkeyScript drives Chromium through the ceremonies; it reads the
	// one-time link from linkEnv.
	passkeyScript = "passkey.mjs"
	linkEnv       = "OPENFIEF_E2E_ENROLL_LINK"
	// localCore is where `openfief enroll` forwards the core by default: the
	// origin the core allows by default (deploy/core's --auth-origin).
	localCore = "http://localhost:8080"
)

// errNoLink reports `openfief enroll` ending before it printed a link.
var errNoLink = errors.New("openfief enroll ended without printing a link")

// TestFirstPasskey runs the owner's first sign-in as the docs describe it:
// `openfief enroll` runs core enroll-link in the core's pod through
// Kubernetes exec and forwards localhost:8080 to the core (D-80), the only
// entry point for browsers, which proxies the built-in UI (D-81); then
// Chromium, with its virtual authenticator, opens the one-time link,
// creates the passkey, keeps the __Host- session cookie on
// http://localhost (D-85), signs out and signs in again with the passkey
// (SCEN-e2e-first-passkey-through-core). No flag weakens the origin check:
// http://localhost:8080 is the core's default allowed origin.
func TestFirstPasskey(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitCoreAvailable(t, k)
	waitPluginReady(t, k, uiPlugin)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	out, lines := io.Pipe()
	done := make(chan error, 1)
	go func() {
		access := &identitykube.CoreAccess{Config: k.config, Client: k.clients, Namespace: namespace}
		err := identitycli.Run(ctx, nil, &identitycli.Env{
			Out:     lines,
			Connect: func(string, string) (identitycli.Core, error) { return access, nil },
		})
		done <- errors.Join(err, lines.Close())
	}()
	link, err := readLink(ctx, out)
	if err != nil {
		cancel()
		t.Fatalf("%v (openfief enroll: %v)", err, <-done)
	}
	waitHTTP(t, localCore+"/readyz", http.StatusOK)
	browser, bcancel := context.WithTimeout(ctx, browserTimeout)
	defer bcancel()
	cmd := exec.CommandContext(browser, "node", passkeyScript)
	cmd.Env = append(os.Environ(), linkEnv+"="+link)
	got, err := cmd.CombinedOutput()
	t.Logf("%s", got)
	if err != nil {
		t.Errorf("the browser's passkey ceremonies: %v", err)
	}
	cancel()
	if err := <-done; err != nil {
		t.Errorf("openfief enroll: %v", err)
	}
}

// readLink reads `openfief enroll`'s output until the one-time link for
// localCore and returns it, then drains the rest in the background.
func readLink(ctx context.Context, out io.Reader) (string, error) {
	found := make(chan string, 1)
	go func() {
		s := bufio.NewScanner(out)
		for s.Scan() {
			if line := strings.TrimSpace(s.Text()); strings.HasPrefix(line, localCore+"/enroll#code=") {
				found <- line
				break
			}
		}
		close(found)
		for s.Scan() { // keep the writer from blocking
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, enrollTimeout)
	defer cancel()
	select {
	case link, ok := <-found:
		if !ok {
			return "", errNoLink
		}
		return link, nil
	case <-ctx.Done():
		return "", fmt.Errorf("waiting for openfief enroll's link: %w", ctx.Err())
	}
}

// waitHTTP waits until GET url, from this machine, answers want.
func waitHTTP(t *testing.T, url string, want int) {
	t.Helper()
	err := waitFor(t.Context(), enrollTimeout, func(ctx context.Context) (string, bool, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return "", false, fmt.Errorf("request: %w", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		if err := resp.Body.Close(); err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		return resp.Status, resp.StatusCode == want, nil
	})
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
}
