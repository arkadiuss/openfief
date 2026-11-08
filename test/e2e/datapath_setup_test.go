//go:build e2e

package e2e

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	identitycli "github.com/arkadiuss/openfief/internal/identity/cli"
	identitykube "github.com/arkadiuss/openfief/internal/identity/kube"
	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// The institution fixture server (test/e2e/fixture), which task e2e:images
// loads into the nodes, as the data path deploys it: an HTTPS Service in a
// namespace of its own, outside openfief, as an owner's in-cluster service
// is (D-74).
const (
	fixtureNamespace = "openfief-e2e-fixture"
	fixtureService   = "fixture"
	fixtureImage     = "openfief.test/fixture:e2e"
	fixtureTLSSecret = "fixture-tls"
	// fixturePort is the Service's port, fixturePodPort the fixture's own.
	fixturePort    = 443
	fixturePodPort = 8443
	// fixtureHost is the Service's DNS name, which its certificate names
	// and the connectors' base URLs use: a plugin has no DNS, so it only
	// names the host TLS verifies (D-74).
	fixtureHost = fixtureService + "." + fixtureNamespace + ".svc"
	// fixtureTimeout bounds the fixture's rollout.
	fixtureTimeout = 5 * time.Minute
	// certValidity is how long the run's CA and certificate are valid.
	certValidity = 6 * time.Hour
	// fixtureUser is the fixture's non-root user, the image's.
	fixtureUser = 65532
)

// The data path's connector instances, each of a first-party connector's
// image in the suite's registry, signed with the run's key.
const (
	dpNBP   = "dp-nbp"
	dpEODHD = "dp-eodhd"
	dpIBKR  = "dp-ibkr"
	dpXTB   = "dp-xtb"
	dpBonds = "dp-bonds"
	// The images task e2e:images builds from plugins/<name>.
	ibkrImage  = "ibkr-flex"
	eodhdImage = "eodhd"
	bondsImage = "treasury-bonds"
)

// The connectors' Secrets (D-89): <instance>--<role>, with these files.
const (
	caBundleRole  = "ca-bundle"
	caBundleFile  = "ca.crt"
	flexRole      = "flex-token"
	flexTokenFile = "token"
	flexQueryFile = "query-id"
	eodhdRole     = "api-token"
	eodhdFile     = "api-token"
)

// The owner's browser (owner.mjs).
const (
	ownerScript = "owner.mjs"
	// ownerReady is the line owner.mjs prints once signed in, and
	// ownerResult starts each answer.
	ownerReady  = "owner: ready"
	ownerResult = "owner-result: "
	// ownerTimeout bounds one form's answer.
	ownerTimeout = 2 * time.Minute
	// ownerLineBytes bounds one line of owner.mjs's output: a page's text.
	ownerLineBytes = 4 << 20
)

// errOwnerGone reports owner.mjs ending before it answered.
var errOwnerGone = errors.New("owner.mjs ended")

// localhostLock is the file whose lock lets one test at a time forward
// localhost:8080 to the core: the origin the core accepts by default, which
// a browser's passkey ceremony needs (D-80).
const localhostLock = "openfief-e2e-localhost-8080.lock"

// lockLocalhostForward waits until no other test of the run forwards
// localhost:8080, and holds that port for t until it ends. TestFirstPasskey
// and the data path's owner take it.
func lockLocalhostForward(t *testing.T) {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(os.TempDir(), localhostLock), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("opening the localhost lock: %v", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("locking localhost:8080: %v", errors.Join(err, f.Close()))
	}
	t.Cleanup(func() {
		if err := errors.Join(syscall.Flock(int(f.Fd()), syscall.LOCK_UN), f.Close()); err != nil {
			t.Errorf("unlocking localhost:8080: %v", err)
		}
	})
}

// fixtureTokens are the synthetic Flex and EODHD tokens the run's fixture
// accepts and nothing else, so a sync shows its connector read its Secret.
type fixtureTokens struct {
	flex, eodhd string
}

// deployFixture runs the institution fixture server as an HTTPS Service,
// fixtureHost, with a certificate for it signed by a CA made for this run,
// and returns that CA's certificate, as PEM, for the connectors' extra CA
// bundles. The fixture accepts only tokens. Its namespace goes when t ends.
func deployFixture(t *testing.T, k *kube, tokens fixtureTokens) []byte {
	t.Helper()
	certs, err := fixtureCertificates(time.Now())
	if err != nil {
		t.Fatalf("the fixture's certificates: %v", err)
	}
	t.Cleanup(func() {
		err := k.clients.CoreV1().Namespaces().Delete(context.WithoutCancel(t.Context()), fixtureNamespace, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			t.Errorf("deleting namespace %s: %v", fixtureNamespace, err)
		}
	})
	t.Cleanup(func() { k.dumpFixtureOnFailure(t) })
	for _, obj := range fixtureObjects(certs.cert, certs.key, tokens) {
		if applyErr := k.applyOne(t.Context(), obj); applyErr != nil {
			t.Fatal(applyErr)
		}
	}
	if err := waitFor(t.Context(), fixtureTimeout, k.deploymentAvailable(fixtureNamespace, fixtureService)); err != nil {
		t.Fatalf("waiting for the fixture: %v", err)
	}
	return certs.ca
}

// dumpFixtureOnFailure logs the fixture's pods and their logs, which name
// each request's method and path, when t failed: the suite's diagnostics
// cover OpenFief's namespaces only.
func (k *kube) dumpFixtureOnFailure(t *testing.T) {
	t.Helper()
	if !t.Failed() {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), diagnosticsTimeout)
	defer cancel()
	pods, err := k.clients.CoreV1().Pods(fixtureNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Logf("listing the fixture's pods: %v", err)
		return
	}
	for i := range pods.Items {
		t.Logf("fixture pod %s", podSummary(&pods.Items[i]))
		k.dumpLogs(ctx, t.Logf, &pods.Items[i])
	}
}

// fixtureObjects are the fixture's namespace, TLS Secret, Deployment and
// Service. The fixture runs as the plugins do: non-root, read-only, no
// privilege escalation and no capabilities.
func fixtureObjects(cert, key []byte, tokens fixtureTokens) []*unstructured.Unstructured {
	labels := map[string]any{"app.kubernetes.io/name": fixtureService}
	meta := func(name string) map[string]any {
		return map[string]any{"name": name, "namespace": fixtureNamespace, "labels": labels}
	}
	container := map[string]any{
		"name": "fixture", "image": fixtureImage, "imagePullPolicy": "IfNotPresent",
		"args":  []any{"-flex-token", tokens.flex, "-eodhd-token", tokens.eodhd},
		"ports": []any{map[string]any{"name": "https", "containerPort": int64(fixturePodPort)}},
		"readinessProbe": map[string]any{
			"tcpSocket": map[string]any{"port": int64(fixturePodPort)}, "periodSeconds": int64(2),
		},
		"volumeMounts": []any{map[string]any{"name": "tls", "mountPath": "/etc/fixture/tls", "readOnly": true}},
		"securityContext": map[string]any{
			"runAsNonRoot": true, "runAsUser": int64(fixtureUser), "readOnlyRootFilesystem": true,
			"allowPrivilegeEscalation": false, "capabilities": map[string]any{"drop": []any{"ALL"}},
			"seccompProfile": map[string]any{"type": "RuntimeDefault"},
		},
	}
	return []*unstructured.Unstructured{
		{Object: map[string]any{
			"apiVersion": "v1", "kind": "Namespace", "metadata": map[string]any{"name": fixtureNamespace},
		}},
		{Object: map[string]any{
			"apiVersion": "v1", "kind": "Secret", "metadata": meta(fixtureTLSSecret), "type": "kubernetes.io/tls",
			"stringData": map[string]any{"tls.crt": string(cert), "tls.key": string(key)},
		}},
		{Object: map[string]any{
			"apiVersion": "apps/v1", "kind": "Deployment", "metadata": meta(fixtureService),
			"spec": map[string]any{
				"replicas": int64(1),
				"selector": map[string]any{"matchLabels": labels},
				"template": map[string]any{
					"metadata": map[string]any{"labels": labels},
					"spec": map[string]any{
						"automountServiceAccountToken": false,
						"containers":                   []any{container},
						"volumes": []any{map[string]any{
							"name": "tls", "secret": map[string]any{"secretName": fixtureTLSSecret},
						}},
					},
				},
			},
		}},
		{Object: map[string]any{
			"apiVersion": "v1", "kind": "Service", "metadata": meta(fixtureService),
			"spec": map[string]any{
				"selector": labels,
				"ports": []any{map[string]any{
					"name": "https", "port": int64(fixturePort), "targetPort": int64(fixturePodPort),
				}},
			},
		}},
	}
}

// fixtureCertificates makes the run's CA and, signed by it, the fixture's
// certificate for fixtureHost, valid from an hour before now for
// certValidity, and returns the CA's certificate, the fixture's and its
// key, as PEM. The CA's key is dropped once it signed.
func fixtureCertificates(now time.Time) (fixturePEM, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fixturePEM{}, fmt.Errorf("CA key: %w", err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "OpenFief e2e fixture CA"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(certValidity),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		return fixturePEM{}, fmt.Errorf("CA certificate: %w", err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return fixturePEM{}, fmt.Errorf("CA certificate: %w", err)
	}
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fixturePEM{}, fmt.Errorf("fixture key: %w", err)
	}
	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: fixtureHost},
		DNSNames:  []string{fixtureHost, fixtureHost + ".cluster.local"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(certValidity),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		return fixturePEM{}, fmt.Errorf("fixture certificate: %w", err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(leafKey)
	if err != nil {
		return fixturePEM{}, fmt.Errorf("fixture key: %w", err)
	}
	return fixturePEM{
		ca:   pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}),
		cert: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafDER}),
		key:  pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
	}, nil
}

// fixturePEM is the run's CA certificate, and the fixture's certificate and
// key, as PEM.
type fixturePEM struct {
	ca, cert, key []byte
}

// dataPathInstance is one connector instance the data path installs: its
// image in the suite's registry, its config, its Secrets by role with their
// files, and whether its endpoint is the fixture's in-cluster Service.
type dataPathInstance struct {
	name, image    string
	config         map[string]any
	secrets        map[string]map[string]string
	clusterService bool
}

// dataPathInstances are the data path's connectors, each calling the
// fixture's Service at its own path over HTTPS, trusting the run's CA from
// its <instance>--ca-bundle, but the XTB parser, which calls nothing.
func dataPathInstances(ca []byte, tokens fixtureTokens) []dataPathInstance {
	base := "https://" + fixtureHost
	bundle := map[string]string{caBundleFile: string(ca)}
	return []dataPathInstance{
		{
			name: dpNBP, image: fxImage, clusterService: true,
			config:  map[string]any{"baseUrl": base + "/nbp-datapath", "start": "2026-01-01"},
			secrets: map[string]map[string]string{caBundleRole: bundle},
		},
		{
			name: dpEODHD, image: eodhdImage, clusterService: true,
			config: map[string]any{"baseUrl": base + "/eodhd"},
			secrets: map[string]map[string]string{
				caBundleRole: bundle, eodhdRole: {eodhdFile: tokens.eodhd},
			},
		},
		{
			name: dpIBKR, image: ibkrImage, clusterService: true,
			config: map[string]any{"baseUrl": base + "/ibkr/AccountManagement/FlexWebService"},
			secrets: map[string]map[string]string{
				caBundleRole: bundle, flexRole: {flexTokenFile: tokens.flex, flexQueryFile: flexOpening},
			},
		},
		{
			name: dpBonds, image: bondsImage, clusterService: true,
			// GUS, NBP's archive, gov.pl and dane.gov.pl are the fixture's,
			// which withholds the Ministry of Finance's data until the bond
			// lots are booked under their provisional ISINs (D-214). The
			// first sync reads CPI from 2025 on, the readings the data path's
			// series need, so it takes one request a year.
			config: map[string]any{
				"start":      "2025-01-01",
				"gusBaseUrl": base + "/gus", "nbpBaseUrl": base + "/nbp-static",
				"govPlBaseUrl": base + "/gov", "daneGovPlBaseUrl": base + "/dane",
			},
			secrets: map[string]map[string]string{caBundleRole: bundle},
		},
		{name: dpXTB, image: xtbImage},
	}
}

// installDataPath creates each data path instance's Secrets but its CA
// bundle, labelled with it as the owner labels them (D-182), installs it
// with the openfief CLI built from this checkout, `openfief plugin install
// <image> --name <instance> --config <file> --extra-ca --cluster-service
// <fixture> --yes`, sets its CA bundle with `openfief plugin ca-bundle set
// <instance> <pem-file>`, which creates the Secret and stamps its checksum,
// and waits until each runs with the stamped bundle. Each is uninstalled
// with the CLI when t ends, which deletes its Secrets and the Role in the
// fixture's namespace.
func installDataPath(t *testing.T, k *kube, reg suiteRegistry, instances []dataPathInstance) {
	t.Helper()
	for i := range instances {
		in := &instances[i]
		img, err := reg.image(t.Context(), in.image)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { uninstallWithCLI(t, in.name) })
		for role, files := range in.secrets {
			if role != caBundleRole {
				createInstanceSecret(t, k, in.name, role, files)
			}
		}
		installWithCLI(t, in, img.cluster)
		if files, ok := in.secrets[caBundleRole]; ok {
			setCABundleWithCLI(t, in.name, []byte(files[caBundleFile]))
		}
	}
	for i := range instances {
		in := &instances[i]
		p := waitPluginReady(t, k, in.name)
		if files, ok := in.secrets[caBundleRole]; ok {
			checkCABundleStamped(t, p)
			waitCABundleRolled(t, k, in.name, p.Spec.CABundleSHA256, []byte(files[caBundleFile]))
		}
	}
}

// setCABundleWithCLI runs `openfief plugin ca-bundle set <instance>
// <pem-file>` with bundle, as the owner does with their kubeconfig.
func setCABundleWithCLI(t *testing.T, instance string, bundle []byte) {
	t.Helper()
	file := filepath.Join(t.TempDir(), instance+"-ca.pem")
	if err := os.WriteFile(file, bundle, 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), cliTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, cliPath)
	// The arguments are the suite's own: a name and a file it wrote.
	cmd.Args = append(cmd.Args, "plugin", "ca-bundle", "set", instance, file)
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "stamped on Plugin "+instance) {
		t.Fatalf("openfief plugin ca-bundle set %s: %v\n%s\nwant it stamped", instance, err, out)
	}
}

// checkCABundleStamped checks that the operator, which never reads the
// CA bundle Secret, reports the checksum the CLI stamped.
func checkCABundleStamped(t *testing.T, p *v1alpha1.Plugin) {
	t.Helper()
	c := apimeta.FindStatusCondition(p.Status.Conditions, v1alpha1.ConditionCABundleMissing)
	if c == nil || c.Status != metav1.ConditionFalse || c.Reason != operator.ReasonCABundleChecksumStamped {
		t.Errorf("plugin %s: condition %s: %+v; want False, reason %s", p.Name, v1alpha1.ConditionCABundleMissing, c, operator.ReasonCABundleChecksumStamped)
	}
}

// waitCABundleRolled waits until every pod of plugin instance carries the
// stamped checksum, which must be the SHA-256 of bundle, and is ready: the
// stamp rolled the pod that may have started before the Secret existed.
func waitCABundleRolled(t *testing.T, k *kube, instance, sum string, bundle []byte) {
	t.Helper()
	if want := plugins.CABundleChecksum(bundle); sum != want {
		t.Fatalf("plugin %s: spec.caBundleSha256 %q, want the SHA-256 of the bundle, %q", instance, sum, want)
	}
	err := waitFor(t.Context(), pluginTimeout, func(ctx context.Context) (string, bool, error) {
		pods, err := k.clients.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: operator.LabelPlugin + "=" + instance})
		if err != nil {
			return "", false, fmt.Errorf("%w: listing the pods of plugin %s: %w", errRetry, instance, err)
		}
		rolled := len(pods.Items) > 0
		for i := range pods.Items {
			pod := &pods.Items[i]
			ready := false
			for _, c := range pod.Status.Conditions {
				ready = ready || (c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue)
			}
			rolled = rolled && ready && pod.Annotations[operator.AnnotationCABundleHash] == sum
		}
		return fmt.Sprintf("%d pods, rolled %v", len(pods.Items), rolled), rolled, nil
	})
	if err != nil {
		t.Fatalf("waiting for plugin %s to run with its CA bundle: %v", instance, err)
	}
}

// createInstanceSecret creates instance's Secret of role with files.
func createInstanceSecret(t *testing.T, k *kube, instance, role string, files map[string]string) {
	t.Helper()
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: instance + "--" + role, Namespace: namespace,
			Labels: map[string]string{plugins.OwnerInstanceLabel: instance},
		},
		StringData: files,
	}
	if _, err := k.clients.CoreV1().Secrets(namespace).Create(t.Context(), s, metav1.CreateOptions{FieldManager: fieldManager}); err != nil {
		t.Fatalf("creating Secret %s: %v", s.Name, err)
	}
}

// installWithCLI installs in from image with the openfief CLI, as the owner
// does with their kubeconfig, its config from a file and the fixture's
// Service as its in-cluster service when it calls one.
func installWithCLI(t *testing.T, in *dataPathInstance, image plugins.ImageRef) {
	t.Helper()
	config := ""
	if in.config != nil {
		data, err := json.Marshal(in.config)
		if err != nil {
			t.Fatal(err)
		}
		config = filepath.Join(t.TempDir(), in.name+".json")
		if err := os.WriteFile(config, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(t.Context(), cliTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, cliPath)
	// The arguments are the suite's own: names, a digest and a file it wrote.
	cmd.Args = append(cmd.Args, installArgs(in, image, config)...)
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "Installed "+in.name+".") {
		t.Fatalf("openfief plugin install %s: %v\n%s\nwant it installed", in.name, err, out)
	}
}

// installArgs are the arguments of `openfief plugin install` for in from
// image, with its config from the file config, if any, --extra-ca when it
// has a CA bundle Secret, and the fixture's Service as its in-cluster
// service when it calls one.
func installArgs(in *dataPathInstance, image plugins.ImageRef, config string) []string {
	args := []string{"plugin", "install", image.String(), "--name", in.name, "--yes"}
	if config != "" {
		args = append(args, "--config", config)
	}
	if _, ok := in.secrets[caBundleRole]; ok {
		// It trusts the run's CA from its <instance>--ca-bundle.
		args = append(args, "--extra-ca")
	}
	if in.clusterService {
		args = append(args, "--cluster-service", fmt.Sprintf("%s/%s:%d", fixtureNamespace, fixtureService, fixturePort))
	}
	return args
}

// uninstallWithCLI runs `openfief plugin uninstall <instance> --yes`.
func uninstallWithCLI(t *testing.T, instance string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cliTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, cliPath)
	cmd.Args = append(cmd.Args, uninstallArgs(instance)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("openfief plugin uninstall %s: %v\n%s", instance, err, out)
	}
}

// uninstallArgs are the arguments of `openfief plugin uninstall` for
// instance, without asking.
func uninstallArgs(instance string) []string {
	return []string{"plugin", "uninstall", instance, "--yes"}
}

// owner is the owner's browser, signed in to the built-in UI with a passkey
// made for this test (owner.mjs).
type owner struct {
	stdin   io.WriteCloser
	answers chan string
}

// ownerRequest is one form owner.mjs sends from the signed-in page: Form
// URL-encoded, or Multipart in its order.
type ownerRequest struct {
	Path      string            `json:"path"`
	Form      map[string]string `json:"form,omitempty"`
	Multipart []ownerPart       `json:"multipart,omitempty"`
}

// ownerPart is a multipart form's field, or a file named Filename with its
// content in Base64.
type ownerPart struct {
	Name     string `json:"name"`
	Value    string `json:"value,omitempty"`
	Filename string `json:"filename,omitempty"`
	Base64   string `json:"base64,omitempty"`
}

// ownerAnswer is what the UI answered a form with, after redirects.
type ownerAnswer struct {
	Status     int    `json:"status"`
	Redirected bool   `json:"redirected"`
	Path       string `json:"path"`
	Body       string `json:"body"`
}

// signInOwner signs the owner in to the built-in UI as the docs give it:
// `openfief enroll` runs core enroll-link in the core's pod and forwards
// localhost:8080 to the core (D-80), and owner.mjs opens the one-time link
// in Chromium, whose virtual authenticator creates a passkey, and keeps
// the session. It holds localhostForward until t ends, when the browser
// and the forward stop.
func signInOwner(t *testing.T, k *kube) *owner {
	t.Helper()
	lockLocalhostForward(t)
	ctx, cancel := context.WithCancel(context.WithoutCancel(t.Context()))
	t.Cleanup(cancel)
	link := enrollLink(t, k)
	waitHTTP(t, localCore+"/readyz", http.StatusOK)
	var logs bytes.Buffer
	browser := exec.CommandContext(ctx, "node", ownerScript)
	browser.Env = append(os.Environ(), linkEnv+"="+link)
	browser.Stderr = &logs
	stdin, err := browser.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := browser.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := browser.Start(); err != nil {
		t.Fatalf("starting owner.mjs: %v", err)
	}
	t.Cleanup(func() { stopBrowser(t, browser, cancel, &logs) })
	o := &owner{stdin: stdin, answers: make(chan string)}
	t.Cleanup(func() {
		if err := stdin.Close(); err != nil {
			t.Logf("closing owner.mjs's input: %v", err)
		}
	})
	ready := make(chan bool, 1)
	go o.read(stdout, ready)
	select {
	case ok := <-ready:
		if !ok {
			t.Fatalf("%v before signing in:\n%s", errOwnerGone, logs.String())
		}
	case <-time.After(browserTimeout):
		t.Fatalf("owner.mjs did not sign in within %s:\n%s", browserTimeout, logs.String())
	}
	return o
}

// enrollLink runs `openfief enroll` as TestFirstPasskey does and returns
// the one-time link it printed; it forwards localhost:8080 to the core
// until t ends.
func enrollLink(t *testing.T, k *kube) string {
	t.Helper()
	ctx, cancel := context.WithCancel(context.WithoutCancel(t.Context()))
	out, lines := io.Pipe()
	enrolled := make(chan error, 1)
	go func() {
		access := &identitykube.CoreAccess{Config: k.config, Client: k.clients, Namespace: namespace}
		err := identitycli.Run(ctx, nil, &identitycli.Env{
			Out:     lines,
			Connect: func(string, string) (identitycli.Core, error) { return access, nil },
		})
		enrolled <- errors.Join(err, lines.Close())
	}()
	t.Cleanup(func() {
		cancel()
		if err := <-enrolled; err != nil {
			t.Logf("openfief enroll: %v", err)
		}
	})
	link, err := readLink(ctx, out)
	if err != nil {
		t.Fatal(err)
	}
	return link
}

// stopBrowser waits for owner.mjs, which ends with its input, and stops it
// with cancel if it doesn't within browserTimeout, logging its errors.
func stopBrowser(t *testing.T, browser *exec.Cmd, cancel context.CancelFunc, logs *bytes.Buffer) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- browser.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Logf("owner.mjs: %v\n%s", err, logs.String())
		}
	case <-time.After(browserTimeout):
		cancel()
		t.Logf("owner.mjs did not end with its input: %v\n%s", <-done, logs.String())
	}
}

// read reads owner.mjs's output: it reports on ready whether the owner
// signed in, then sends each answer's JSON to o.answers, and closes it when
// the output ends.
func (o *owner) read(stdout io.Reader, ready chan<- bool) {
	defer close(o.answers)
	s := bufio.NewScanner(stdout)
	s.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), ownerLineBytes)
	signedIn := false
	for s.Scan() {
		line := s.Text()
		switch {
		case !signedIn && line == ownerReady:
			signedIn = true
			ready <- true
		case signedIn && strings.HasPrefix(line, ownerResult):
			o.answers <- strings.TrimPrefix(line, ownerResult)
		}
	}
	if !signedIn {
		ready <- false
	}
}

// post sends req from the owner's signed-in page and returns the UI's
// answer.
func (o *owner) post(t *testing.T, req *ownerRequest) ownerAnswer {
	t.Helper()
	line, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := o.stdin.Write(append(line, '\n')); err != nil {
		t.Fatalf("POST %s: %v", req.Path, err)
	}
	select {
	case got, ok := <-o.answers:
		if !ok {
			t.Fatalf("POST %s: %v", req.Path, errOwnerGone)
		}
		var a ownerAnswer
		if err := json.Unmarshal([]byte(got), &a); err != nil {
			t.Fatalf("POST %s: decoding the answer: %v", req.Path, err)
		}
		return a
	case <-time.After(ownerTimeout):
		t.Fatalf("POST %s: no answer within %s", req.Path, ownerTimeout)
	}
	return ownerAnswer{}
}

// postForm sends a form that redirects on success, as every owner form of
// the UI but the upload does, and fails t unless it was taken.
func (o *owner) postForm(t *testing.T, path string, form map[string]string) {
	t.Helper()
	a := o.post(t, &ownerRequest{Path: path, Form: form})
	if a.Status != http.StatusOK || !a.Redirected {
		t.Fatalf("POST %s: status %d, redirected %t, at %s; want it taken and redirected:\n%s", path, a.Status, a.Redirected, a.Path, excerpt(a.Body))
	}
}

// statementFile is a statement the owner uploads: its file name and
// content, and the opening cash the owner enters with it, with its
// currency, if any (D-136).
type statementFile struct {
	name                  string
	content               []byte
	openingCash, currency string
}

// upload uploads f to account through the UI's upload form, in the mode
// the core chooses, and fails t unless the UI shows the import.
func (o *owner) upload(t *testing.T, account string, f *statementFile) {
	t.Helper()
	parts := []ownerPart{{Name: "mode", Value: "auto"}}
	if f.openingCash != "" {
		parts = append(parts, ownerPart{Name: "opening_cash", Value: f.openingCash}, ownerPart{Name: "opening_cash_currency", Value: f.currency})
	}
	parts = append(parts, ownerPart{Name: "statement", Filename: f.name, Base64: base64.StdEncoding.EncodeToString(f.content)})
	path := "/accounts/" + account + "/statements"
	a := o.post(t, &ownerRequest{Path: path, Multipart: parts})
	if a.Status != http.StatusOK {
		t.Fatalf("POST %s with %s: status %d; want the import shown:\n%s", path, f.name, a.Status, excerpt(a.Body))
	}
}

// excerpt is the start of a page's text, for a failure message.
func excerpt(body string) string {
	const maxExcerpt = 2000
	if len(body) > maxExcerpt {
		return body[:maxExcerpt] + "..."
	}
	return body
}
