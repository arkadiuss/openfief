//go:build e2e

package e2e

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"

	identitykube "github.com/arkadiuss/openfief/internal/identity/kube"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The suite's bundle run (task e2e runs it before the demo run): TestMain
// installs OpenFief from the release bundle that task e2e rendered with
// scripts/release-bundle.sh from the images task e2e:images pushed to the
// suite's registry, each pinned by digest, as an owner installs a release
// (docs/guides/install.md): CloudNativePG from the bundle's pinned release
// manifest, the CRDs file, the bundle's openfief-db cluster, then the
// install file, all unchanged. Between the CRDs and the install it adds the
// suite's key to the owner's trusted signers, as an owner who signs plugins
// with their own key does (D-45), since this checkout's UI image can't carry
// the release pipeline's signature. Only the TestBundle tests run against
// it, and TestMain removes the install again afterwards, so the demo run
// starts from an empty namespace.
const (
	// bundleEnv names the directory with the bundle's files; set, it makes
	// TestMain install the bundle instead of the demo's stack.
	bundleEnv = "OPENFIEF_E2E_BUNDLE"
	// removeTimeout bounds the removal of OpenFief's namespace.
	removeTimeout = 5 * time.Minute
	// forwardTimeout bounds the start of a port-forward to the core.
	forwardTimeout = 30 * time.Second
	// requestTimeout bounds one HTTP call through the port-forward.
	requestTimeout = 30 * time.Second
	// bundleBodyLimit bounds how much of a reply the tests read.
	bundleBodyLimit = 1 << 20
)

var (
	// errBundleFiles reports a bundle directory without exactly one file
	// of each kind the bundle holds.
	errBundleFiles = errors.New("the bundle needs exactly one openfief-crds-*.yaml, openfief-install-*.yaml, " +
		"openfief-db-*.yaml and cnpg-*.yaml")
	// errNotPinned reports an image in the bundle that isn't pinned by digest.
	errNotPinned = errors.New("image not pinned by digest")
)

// bundle is a rendered release bundle: its CRDs, its install manifests,
// its database cluster and CloudNativePG's release manifest.
type bundle struct {
	crds, install, database, cnpg []byte
}

// readBundle reads the bundle's files from dir.
func readBundle(dir string) (bundle, error) {
	crds, crdsErr := readOne(dir, "openfief-crds-*.yaml")
	install, installErr := readOne(dir, "openfief-install-*.yaml")
	database, databaseErr := readOne(dir, "openfief-db-*.yaml")
	cnpg, cnpgErr := readOne(dir, "cnpg-*.yaml")
	if err := errors.Join(crdsErr, installErr, databaseErr, cnpgErr); err != nil {
		return bundle{}, err
	}
	return bundle{crds: crds, install: install, database: database, cnpg: cnpg}, nil
}

// readOne reads the one file in dir that matches pattern.
func readOne(dir, pattern string) ([]byte, error) {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", pattern, err)
	}
	if len(matches) != 1 {
		return nil, fmt.Errorf("%w: %d files match %s in %s", errBundleFiles, len(matches), pattern, dir)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("opening the bundle's directory: %w", err)
	}
	data, err := root.ReadFile(filepath.Base(matches[0]))
	if closeErr := root.Close(); err == nil && closeErr != nil {
		err = closeErr
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", matches[0], err)
	}
	return data, nil
}

// runBundle is TestMain's bundle run: it installs the bundle in dir, runs
// the tests, removes the install whatever happened, and returns the exit
// code.
func runBundle(m *testing.M, dir string) int {
	k, err := connect()
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), installTimeout)
		err = installBundle(ctx, k, dir, logStderrf)
		cancel()
	}
	if err == nil {
		err = resetHoldsCore()
	}
	code := 1
	if err != nil {
		logStderrf("e2e: installing the release bundle failed: %v", err)
		if k != nil {
			dctx, dcancel := context.WithTimeout(context.Background(), diagnosticsTimeout)
			k.dumpDiagnostics(dctx, logStderrf)
			dcancel()
		}
	} else {
		code = m.Run()
	}
	if k != nil {
		ctx, cancel := context.WithTimeout(context.Background(), removeTimeout)
		if rmErr := k.removeNamespace(ctx, logStderrf); rmErr != nil {
			logStderrf("e2e: removing the bundle's install: %v", rmErr)
			code = 1
		}
		cancel()
	}
	return code
}

// installBundle installs the bundle in dir on a cluster without OpenFief:
// it removes OpenFief's namespace if an earlier run left it, then applies,
// in the order the install guide gives, the bundle's CloudNativePG manifest,
// its CRDs, the owner's trusted signer (the suite's key), with which it
// builds the CLI (buildCLI), its database and its install.
func installBundle(ctx context.Context, k *kube, dir string, logf func(string, ...any)) error {
	b, err := readBundle(dir)
	if err != nil {
		return err
	}
	reg, err := registryFromEnv()
	if err != nil {
		return err
	}
	steps := []func() error{
		func() error { return k.removeNamespace(ctx, logf) },
		func() error {
			logf("e2e: installing CloudNativePG from the release bundle in %s", dir)
			return applyCNPG(ctx, k, b.cnpg)
		},
		func() error {
			logf("e2e: applying the release bundle's CRDs from %s", dir)
			return k.apply(ctx, b.crds)
		},
		func() error {
			signer, trustErr := trustSigner(ctx, k, reg, logf)
			if trustErr != nil {
				return trustErr
			}
			return buildCLI(ctx, reg, signer, logf)
		},
		func() error { return applyDatabase(ctx, k, logf, "the release bundle's openfief-db", b.database) },
		func() error {
			logf("e2e: applying the release bundle's install from %s", dir)
			return k.apply(ctx, b.install)
		},
	}
	for _, step := range steps {
		if stepErr := step(); stepErr != nil {
			return stepErr
		}
	}
	return nil
}

// removeNamespace deletes OpenFief's namespace, with everything in it, and
// waits until it is gone. A namespace that doesn't exist is fine.
func (k *kube) removeNamespace(ctx context.Context, logf func(string, ...any)) error {
	namespaces := k.clients.CoreV1().Namespaces()
	err := namespaces.Delete(ctx, namespace, metav1.DeleteOptions{})
	switch {
	case apierrors.IsNotFound(err):
		return nil
	case err != nil:
		return fmt.Errorf("deleting namespace %s: %w", namespace, err)
	}
	logf("e2e: removing namespace %s", namespace)
	err = waitFor(ctx, removeTimeout, func(ctx context.Context) (string, bool, error) {
		ns, getErr := namespaces.Get(ctx, namespace, metav1.GetOptions{})
		switch {
		case apierrors.IsNotFound(getErr):
			return "gone", true, nil
		case getErr != nil:
			return "", false, fmt.Errorf("%w: %w", errRetry, getErr)
		}
		return string(ns.Status.Phase), false, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for namespace %s to be removed: %w", namespace, err)
	}
	return nil
}

// bundleDeployments returns the Deployments in the install file of the
// bundle that TestMain installed.
func bundleDeployments(t *testing.T) []appsv1.Deployment {
	t.Helper()
	b, err := readBundle(os.Getenv(bundleEnv))
	if err != nil {
		t.Fatal(err)
	}
	var deployments []appsv1.Deployment
	decoder := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(b.install)))
	for {
		doc, readErr := decoder.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			t.Fatalf("reading the install file: %v", readErr)
		}
		var d appsv1.Deployment
		if err := yaml.Unmarshal(doc, &d); err != nil {
			t.Fatalf("decoding the install file: %v", err)
		}
		if d.Kind == "Deployment" {
			deployments = append(deployments, d)
		}
	}
	return deployments
}

// TestBundlePinsImageDigests checks that every Deployment of the bundle's
// install runs its images by digest, from the suite's registry, where task
// e2e:images pushed this checkout's images, and that the cluster runs
// exactly those (SCEN-e2e-bundle-pinned-images).
func TestBundlePinsImageDigests(t *testing.T) {
	holdsCore(t)
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	deployments := bundleDeployments(t)
	names := make([]string, 0, len(deployments))
	for i := range deployments {
		d := &deployments[i]
		names = append(names, d.Name)
		live, getErr := k.clients.AppsV1().Deployments(namespace).Get(t.Context(), d.Name, metav1.GetOptions{})
		if getErr != nil {
			t.Fatalf("getting Deployment %s: %v", d.Name, getErr)
		}
		want := containerImages(d)
		for _, image := range want {
			if !strings.HasPrefix(image, reg.cluster+"/") || !strings.Contains(image, "@sha256:") {
				t.Errorf("Deployment %s: %v: %s, want %s/<name>@sha256:<digest>", d.Name, errNotPinned, image, reg.cluster)
			}
		}
		if got := containerImages(live); !slices.Equal(got, want) {
			t.Errorf("Deployment %s runs %v, the bundle names %v", d.Name, got, want)
		}
	}
	slices.Sort(names)
	if want := []string{"egress-proxy", "openfief-core", "openfief-operator"}; !slices.Equal(names, want) {
		t.Errorf("the bundle's Deployments are %v, want %v", names, want)
	}
}

// containerImages returns the images of d's init and app containers, in
// order.
func containerImages(d *appsv1.Deployment) []string {
	spec := &d.Spec.Template.Spec
	images := make([]string, 0, len(spec.InitContainers)+len(spec.Containers))
	for i := range spec.InitContainers {
		images = append(images, spec.InitContainers[i].Image)
	}
	for i := range spec.Containers {
		images = append(images, spec.Containers[i].Image)
	}
	return images
}

// TestBundleBecomesReady checks that each Deployment of the bundle's install
// becomes available: the operator, the egress proxy and the core
// (SCEN-e2e-bundle-ready).
func TestBundleBecomesReady(t *testing.T) {
	holdsCore(t)
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	for _, d := range bundleDeployments(t) {
		available := k.deploymentAvailable(namespace, d.Name)
		err := waitFor(t.Context(), coreTimeout, func(ctx context.Context) (string, bool, error) {
			if state, done, err := available(ctx); done || (err != nil && !errors.Is(err, errRetry)) {
				return state, done, err
			}
			state, err := k.podsState(ctx, namespace, d.Spec.Selector.MatchLabels["app.kubernetes.io/name"])
			return state, false, err
		})
		if err != nil {
			t.Fatalf("waiting for Deployment %s to be available: %v", d.Name, err)
		}
	}
}

// TestBundleReadAPI checks, through a port-forward to the core as
// `openfief enroll` opens one, that the core installed from the bundle
// answers /readyz, lists accounts with a script token issued by its own CLI,
// and refuses the read API without one (SCEN-e2e-bundle-read-api).
func TestBundleReadAPI(t *testing.T) {
	holdsCore(t)
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitCoreAvailable(t, k)
	base := forwardCore(t, k)
	token := issueToken(t, k)

	if r := bundleGet(t, base+"/readyz", ""); r.status != http.StatusOK {
		t.Fatalf("GET /readyz: status %d, body %q; want 200", r.status, r.body)
	}
	r := bundleGet(t, base+"/v1/accounts", token)
	if r.status != http.StatusOK {
		t.Fatalf("GET /v1/accounts: status %d, body %q; want 200", r.status, r.body)
	}
	var accounts viewsv1.ListAccountsResponse
	if err := protojson.Unmarshal(r.body, &accounts); err != nil {
		t.Fatalf("decoding the accounts: %v", err)
	}
	if r := bundleGet(t, base+"/v1/accounts", ""); r.status != http.StatusUnauthorized {
		t.Fatalf("GET /v1/accounts without a token: status %d, body %q; want 401", r.status, r.body)
	}
}

// forwardCore forwards a free local port to the core's pod until t ends and
// returns its base URL.
func forwardCore(t *testing.T, k *kube) string {
	t.Helper()
	l, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("finding a free port: %v", err)
	}
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("listener address %v is not TCP", l.Addr())
	}
	port := addr.Port
	if err := l.Close(); err != nil {
		t.Fatalf("freeing port %d: %v", port, err)
	}
	access := &identitykube.CoreAccess{Config: k.config, Client: k.clients, Namespace: namespace}
	ctx, cancel := context.WithCancel(context.WithoutCancel(t.Context()))
	ready := make(chan struct{})
	done := make(chan error, 1)
	go func() { done <- access.Forward(ctx, port, ready) }()
	t.Cleanup(func() {
		cancel()
		if err := <-done; err != nil {
			t.Errorf("port-forward to the core: %v", err)
		}
	})
	select {
	case <-ready:
	case err := <-done:
		t.Fatalf("port-forward to the core: %v", err)
	case <-time.After(forwardTimeout):
		t.Fatalf("port-forward to the core: not listening after %s", forwardTimeout)
	}
	return fmt.Sprintf("http://127.0.0.1:%d", port)
}

// bundleGet sends GET url with token as the bearer token unless it is empty,
// and returns the status and body. The token never appears in a failure.
func bundleGet(t *testing.T, url, token string) reply {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("GET %s: closing the body: %v", url, closeErr)
		}
	}()
	body, err := io.ReadAll(io.LimitReader(resp.Body, bundleBodyLimit))
	if err != nil {
		t.Fatalf("GET %s: reading the body: %v", url, err)
	}
	return reply{status: resp.StatusCode, body: body}
}
