//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arkadiuss/openfief/internal/demo"
	identitykube "github.com/arkadiuss/openfief/internal/identity/kube"
	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The suite's synthetic connector (test/e2e/connector), whose two images
// task e2e:images pushes to the suite's registry, and the instance
// TestConnectorCredentialsAndUninstall installs.
const (
	// connectorImage carries connectorManifest as its signed label,
	// connectorEgressImage connectorEgressManifest, which asks for the
	// egress host example.org too.
	connectorImage          = "e2e-connector"
	connectorEgressImage    = "e2e-connector-egress"
	connectorManifest       = "connector/manifest.json"
	connectorEgressManifest = "connector/manifest-egress.json"
	// connectorRole is the Secret role both manifests name, and
	// connectorKey the key of that Secret the connector reads.
	connectorRole = "credentials"
	connectorKey  = "token"
	// connectorInstitution is the institution both manifests declare.
	connectorInstitution = "e2e-bank"
	// booksPlugin is the instance that books a synthetic deposit into the
	// demo's XTB account, then is uninstalled.
	booksPlugin = "e2e-books"
	// depositPrefix starts the external ID of the connector's deposit,
	// which ends with its date.
	depositPrefix = "deposit-"
	// syntheticBytes is how many random bytes a synthetic credential has.
	syntheticBytes = 16
	// transactionPages bounds how many pages of an account's transactions
	// a lookup reads.
	transactionPages = 20
	// transactionPageSize is the read API's largest page.
	transactionPageSize = 500
	// dumpTimeout bounds the database dump.
	dumpTimeout = 5 * time.Minute
	// databasePodSelector selects the openfief-db cluster's primary.
	databasePodSelector = "cnpg.io/cluster=openfief-db,cnpg.io/instanceRole=primary"
)

// errNotListed reports a transaction the read API doesn't list.
var errNotListed = errors.New("not listed")

// TestConnectorCredentialsAndUninstall runs a connector with credentials
// the way an owner does and then uninstalls it. The owner creates its
// Secret e2e-books--credentials (D-29, D-89) with a synthetic credential
// made for this run, binds the instance to the demo's XTB account with the
// core's CLI and applies its Plugin and a PluginGrant naming that Secret,
// for the synthetic connector's image signed by the suite's key. The test
// checks that the Secret is mounted into the connector's pod and no other
// pod in the cluster, the core's included, and that the credential is in no
// ConfigMap, no container's environment, command or arguments, no Plugin or
// PluginGrant, and nowhere in the database, once a sync, which reads the
// credential, has booked the connector's deposit
// (SCEN-e2e-credentials-only-in-secrets,
// SCEN-e2e-credentials-mounted-only-in-granted-plugin). `openfief plugin
// uninstall`, the CLI built from this checkout, then removes the Plugin, its
// grant, its Secret, its workload and its allow-list entry, while the
// deposit stays in the read API, marked as from a removed instance (D-58;
// SCEN-e2e-uninstall-keeps-data). A second deposit, booked from the same
// instance under a changed config before the uninstall, stays too, marked
// the same way, and once the instance is gone no pod of it and no Secret
// labelled as its own remains (D-200 criterion 18,
// SCEN-e2e-uninstall-leaves-no-pod-or-secret).
func TestConnectorCredentialsAndUninstall(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	if _, err := os.Stat(cliPath); err != nil {
		t.Fatalf("%v: %v", errNoCLI, err)
	}
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	img, err := reg.image(t.Context(), connectorImage)
	if err != nil {
		t.Fatal(err)
	}
	waitCoreAvailable(t, k)
	waitSeedComplete(t, k)
	t.Cleanup(func() { deleteInstance(t, k, booksPlugin) })

	credential := createCredential(t, k, booksPlugin)
	account := string(demo.XTBAccount)
	// Yesterday, so the deposit is never ahead of the core's today.
	date := time.Now().UTC().AddDate(0, 0, -1).Format(time.DateOnly)
	bindBooks(t, k, account)
	plugin := connectorPlugin(t, booksPlugin, img.cluster, readConnectorManifests(t).first)
	if err := unstructured.SetNestedMap(plugin.Object, map[string]any{"account": account, "date": date}, "spec", "config"); err != nil {
		t.Fatal(err)
	}
	for _, obj := range []*unstructured.Unstructured{connectorGrant(booksPlugin), plugin} {
		if err := k.applyOne(t.Context(), obj); err != nil {
			t.Fatal(err)
		}
	}
	waitPluginReady(t, k, booksPlugin)
	checkMountedOnlyByInstance(t, k, booksPlugin)

	syncDeposit(t, k, booksPlugin)
	token := issueToken(t, k)
	waitDeposit(t, k, token, account, depositPrefix+date, false)
	checkCredentialNowhereElse(t, k, credential)

	// A second deposit, a day earlier, from the same instance: the owner
	// changes the Plugin's config, the operator rolls the pod, and a sync
	// books it next to the first.
	second := time.Now().UTC().AddDate(0, 0, -2).Format(time.DateOnly)
	if err := unstructured.SetNestedMap(plugin.Object, map[string]any{"account": account, "date": second}, "spec", "config"); err != nil {
		t.Fatal(err)
	}
	if err := k.applyOne(t.Context(), plugin); err != nil {
		t.Fatal(err)
	}
	waitPluginReady(t, k, booksPlugin)
	syncUntilDeposit(t, k, booksPlugin, token, account, depositPrefix+second)
	waitDeposit(t, k, token, account, depositPrefix+date, false)

	uninstallBooks(t)
	waitUninstalled(t, k, booksPlugin)
	for _, d := range []string{date, second} {
		waitDeposit(t, k, token, account, depositPrefix+d, true)
	}
}

// syncUntilDeposit syncs instance until the read API lists the deposit with
// externalID in account. A sync right after a config change may reach the
// pod the operator is replacing, which books the old deposit again, so the
// sync is repeated until the new one is there.
func syncUntilDeposit(t *testing.T, k *kube, instance, token, account, externalID string) {
	t.Helper()
	err := waitFor(t.Context(), syncTimeout, func(ctx context.Context) (string, bool, error) {
		pod, err := k.corePod(ctx)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		out, err := k.exec(ctx, namespace, pod, "core", nil,
			"/core", "connector", "sync", "--instance", instance, "--mode", "transactions")
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		_, err = findTransaction(ctx, k, token, account, externalID)
		switch {
		case errors.Is(err, errNotListed):
			return fmt.Sprintf("synced (%s), deposit %s not listed", bytes.TrimSpace(out), externalID), false, nil
		case err != nil:
			return "", false, err
		}
		return "deposit listed", true, nil
	})
	if err != nil {
		t.Fatalf("syncing instance %s until deposit %s is listed in account %s: %v", instance, externalID, account, err)
	}
}

// credentialSecret is the Secret instance mounts for the connector's
// credentials role, <instance>--<role> (D-89).
func credentialSecret(instance string) string { return instance + "--" + connectorRole }

// syntheticCredential returns a random credential made for this run: no
// real institution issued it, and it is never logged.
func syntheticCredential(t *testing.T) string {
	t.Helper()
	b := make([]byte, syntheticBytes)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("synthetic credential: %v", err)
	}
	return "e2e-synthetic-" + hex.EncodeToString(b)
}

// connectorGrant is a PluginGrant for instance of the synthetic connector
// that approves exactly what its first manifest requests: the connector
// capability, its institution and its credentials Secret.
func connectorGrant(instance string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "openfief.io/v1alpha1", "kind": "PluginGrant",
		"metadata": map[string]any{"name": instance, "namespace": namespace},
		"spec": map[string]any{
			"capabilities": []any{"connector"},
			"institutions": []any{connectorInstitution},
			"secrets":      []any{credentialSecret(instance)},
		},
	}}
}

// connectorPlugin is a Plugin for instance of the synthetic connector's
// image, declaring manifest, with no config.
func connectorPlugin(t *testing.T, instance string, image plugins.ImageRef, manifest map[string]any) *unstructured.Unstructured {
	t.Helper()
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "openfief.io/v1alpha1", "kind": "Plugin",
		"metadata": map[string]any{"name": instance, "namespace": namespace},
		"spec":     map[string]any{"image": image.String(), "manifest": manifest},
	}}
}

// bindBooks binds booksPlugin to account with the core's own CLI inside its
// pod, as an owner would with kubectl exec, so its deliveries may write
// there.
func bindBooks(t *testing.T, k *kube, account string) {
	t.Helper()
	pod, err := k.corePod(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := k.exec(t.Context(), namespace, pod, "core", nil,
		"/core", "connector", "bind", "--instance", booksPlugin, "--account", account); err != nil {
		t.Fatalf("core connector bind --instance %s: %v", booksPlugin, err)
	}
}

// corePod returns a ready core pod, as `openfief enroll` finds it.
func (k *kube) corePod(ctx context.Context) (string, error) {
	access := &identitykube.CoreAccess{Config: k.config, Client: k.clients, Namespace: namespace}
	pod, err := access.CorePod(ctx)
	if err != nil {
		return "", fmt.Errorf("finding the core's pod: %w", err)
	}
	return pod, nil
}

// uninstallBooks runs `openfief plugin uninstall e2e-books --yes` with the
// CLI built from this checkout, as the owner does with their kubeconfig.
func uninstallBooks(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), cliTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, cliPath, "plugin", "uninstall", booksPlugin, "--yes")
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("openfief plugin uninstall %s: %v\n%s", booksPlugin, err, out.String())
	}
	if !strings.Contains(out.String(), "Uninstalled "+booksPlugin+".") {
		t.Fatalf("openfief plugin uninstall %s printed:\n%s\nwant it uninstalled", booksPlugin, out.String())
	}
}

// syncDeposit syncs instance in transactions mode with the core's CLI until
// the delivery holds the connector's one deposit: the connector reads its
// credential on every sync, so a sync fails while it isn't mounted. The
// core's scheduler may sync the instance at the same moment, which makes a
// sync fail, so a failure is tried again until the deadline.
func syncDeposit(t *testing.T, k *kube, instance string) {
	t.Helper()
	err := waitFor(t.Context(), syncTimeout, func(ctx context.Context) (string, bool, error) {
		pod, err := k.corePod(ctx)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		out, err := k.exec(ctx, namespace, pod, "core", nil,
			"/core", "connector", "sync", "--instance", instance, "--mode", "transactions")
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		if !bytes.Contains(out, []byte(": 1 transactions,")) {
			return "", false, fmt.Errorf("%w: unexpected output %q", errRetry, out)
		}
		return string(out), true, nil
	})
	if err != nil {
		t.Fatalf("core connector sync --instance %s: %v", instance, err)
	}
}

// waitDeposit waits until the read API lists the connector's deposit with
// externalID in account, from booksPlugin, with its removed-instance mark
// as removed says.
func waitDeposit(t *testing.T, k *kube, token, account, externalID string, removed bool) {
	t.Helper()
	err := waitFor(t.Context(), pluginTimeout, func(ctx context.Context) (string, bool, error) {
		tx, err := findTransaction(ctx, k, token, account, externalID)
		switch {
		case errors.Is(err, errNotListed):
			return "deposit not listed", false, nil
		case err != nil:
			return "", false, err
		case tx.GetConnectorInstance() != booksPlugin:
			return "", false, fmt.Errorf("%w: deposit from instance %q, want %q", errStillThere, tx.GetConnectorInstance(), booksPlugin)
		}
		state := fmt.Sprintf("deposit listed, connector_instance_removed=%t", tx.GetConnectorInstanceRemoved())
		return state, tx.GetConnectorInstanceRemoved() == removed, nil
	})
	if err != nil {
		t.Fatalf("waiting for the deposit %s in account %s with connector_instance_removed=%t: %v", externalID, account, removed, err)
	}
}

// findTransaction pages through account's transactions in the read API
// for the one with externalID, failing with errNotListed when it isn't
// listed.
func findTransaction(ctx context.Context, k *kube, token, account, externalID string) (*viewsv1.Transaction, error) {
	page := ""
	for range transactionPages {
		txs, err := transactionsPage(ctx, k, token, account, page)
		if err != nil {
			return nil, err
		}
		if i := slices.IndexFunc(txs.GetTransactions(), func(tx *viewsv1.Transaction) bool {
			return tx.GetExternalId() == externalID
		}); i >= 0 {
			return txs.GetTransactions()[i], nil
		}
		if page = txs.GetNextPageToken(); page == "" {
			break
		}
	}
	return nil, fmt.Errorf("%w: transaction %s in account %s", errNotListed, externalID, account)
}

// transactionsPage reads one page of account's transactions from the read
// API through the client pod.
func transactionsPage(ctx context.Context, k *kube, token, account, page string) (*viewsv1.ListTransactionsResponse, error) {
	path := fmt.Sprintf("/v1/accounts/%s/transactions?page_size=%d", account, transactionPageSize)
	if page != "" {
		path += "&page_token=" + page
	}
	r, err := call(ctx, k, clientPod, path, token)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errRetry, err)
	}
	if r.status != http.StatusOK {
		return nil, fmt.Errorf("%w: GET %s: status %d, body %q", errRetry, path, r.status, r.body)
	}
	var txs viewsv1.ListTransactionsResponse
	if err := protojson.Unmarshal(r.body, &txs); err != nil {
		return nil, fmt.Errorf("decoding the transactions: %w", err)
	}
	return &txs, nil
}

// checkMountedOnlyByInstance checks that, of every pod in the openfief
// namespace, the only one that can see a Secret of it, only instance's own
// pods use its credentials Secret, and that they do.
func checkMountedOnlyByInstance(t *testing.T, k *kube, instance string) {
	t.Helper()
	pods, err := k.clients.CoreV1().Pods(namespace).List(t.Context(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing pods: %v", err)
	}
	secret := credentialSecret(instance)
	var users, own []string
	for i := range pods.Items {
		p := &pods.Items[i]
		if p.Labels[operator.LabelPlugin] == instance && p.DeletionTimestamp == nil {
			own = append(own, p.Name)
		}
		if volumesUseSecret(p.Spec.Volumes, secret) || containersUseSecret(slices.Concat(p.Spec.InitContainers, p.Spec.Containers), secret) {
			users = append(users, p.Name)
		}
	}
	slices.Sort(users)
	slices.Sort(own)
	if len(own) == 0 || !slices.Equal(users, own) {
		t.Errorf("pods using Secret %s: %v; want exactly instance %s's pods %v", secret, users, instance, own)
	}
}

// volumesUseSecret reports whether volumes mount Secret name, whole or as
// a projected source.
func volumesUseSecret(volumes []corev1.Volume, name string) bool {
	return slices.ContainsFunc(volumes, func(v corev1.Volume) bool {
		if v.Secret != nil && v.Secret.SecretName == name {
			return true
		}
		return v.Projected != nil && slices.ContainsFunc(v.Projected.Sources, func(s corev1.VolumeProjection) bool {
			return s.Secret != nil && s.Secret.Name == name
		})
	})
}

// containersUseSecret reports whether a container takes environment
// variables from Secret name.
func containersUseSecret(containers []corev1.Container, name string) bool {
	return slices.ContainsFunc(containers, func(c corev1.Container) bool {
		return slices.ContainsFunc(c.EnvFrom, func(e corev1.EnvFromSource) bool {
			return e.SecretRef != nil && e.SecretRef.Name == name
		}) || slices.ContainsFunc(c.Env, func(e corev1.EnvVar) bool {
			return e.ValueFrom != nil && e.ValueFrom.SecretKeyRef != nil && e.ValueFrom.SecretKeyRef.Name == name
		})
	})
}

// checkCredentialNowhereElse checks that the credential, which lives in its
// Secret, appears in no ConfigMap, no pod's spec, which holds every
// container's environment, command and arguments, in any namespace, no
// Plugin or PluginGrant, and nowhere in a dump of every database of the
// openfief-db cluster. It names where it found it, never the credential.
func checkCredentialNowhereElse(t *testing.T, k *kube, credential string) {
	t.Helper()
	needle := []byte(credential)
	found := slices.Concat(k.configMapsHolding(t, needle), k.podsHolding(t, needle), k.pluginsHolding(t, needle))
	if bytes.Contains(dumpDatabase(t, k), needle) {
		found = append(found, "the database")
	}
	if len(found) > 0 {
		t.Errorf("the synthetic credential is outside its Secret, in: %s", strings.Join(found, "; "))
	}
}

// configMapsHolding names every ConfigMap in the cluster with needle in a
// value.
func (k *kube) configMapsHolding(t *testing.T, needle []byte) []string {
	t.Helper()
	cms, err := k.clients.CoreV1().ConfigMaps(metav1.NamespaceAll).List(t.Context(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing ConfigMaps: %v", err)
	}
	var found []string
	for i := range cms.Items {
		cm := &cms.Items[i]
		text := slices.ContainsFunc(slices.Collect(maps.Values(cm.Data)), func(v string) bool {
			return strings.Contains(v, string(needle))
		})
		binary := slices.ContainsFunc(slices.Collect(maps.Values(cm.BinaryData)), func(v []byte) bool {
			return bytes.Contains(v, needle)
		})
		if text || binary {
			found = append(found, "ConfigMap "+cm.Namespace+"/"+cm.Name)
		}
	}
	return found
}

// podsHolding names every pod in the cluster with needle in its spec.
func (k *kube) podsHolding(t *testing.T, needle []byte) []string {
	t.Helper()
	return k.podsHoldingIn(t, metav1.NamespaceAll, needle)
}

// podsHoldingIn names every pod of ns with needle in its spec.
func (k *kube) podsHoldingIn(t *testing.T, ns string, needle []byte) []string {
	t.Helper()
	pods, err := k.clients.CoreV1().Pods(ns).List(t.Context(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing pods: %v", err)
	}
	var found []string
	for i := range pods.Items {
		p := &pods.Items[i]
		spec, err := json.Marshal(&p.Spec)
		if err != nil {
			t.Fatalf("encoding pod %s/%s: %v", p.Namespace, p.Name, err)
		}
		if bytes.Contains(spec, needle) {
			found = append(found, "pod "+p.Namespace+"/"+p.Name)
		}
	}
	return found
}

// pluginsHolding names the kinds of Plugin and PluginGrant whose objects
// hold needle.
func (k *kube) pluginsHolding(t *testing.T, needle []byte) []string {
	t.Helper()
	var found []string
	for _, gvr := range []schema.GroupVersionResource{pluginResource(), grantResource()} {
		list, err := k.dynamic.Resource(gvr).Namespace(namespace).List(t.Context(), metav1.ListOptions{})
		if err != nil {
			t.Fatalf("listing %s: %v", gvr.Resource, err)
		}
		data, err := list.MarshalJSON()
		if err != nil {
			t.Fatalf("encoding %s: %v", gvr.Resource, err)
		}
		if bytes.Contains(data, needle) {
			found = append(found, gvr.Resource)
		}
	}
	return found
}

// dumpDatabase returns a data-only dump of every database of the
// openfief-db cluster, taken inside its primary's postgres container as the
// postgres user, over the local socket.
func dumpDatabase(t *testing.T, k *kube) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), dumpTimeout)
	defer cancel()
	pods, err := k.clients.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: databasePodSelector})
	if err != nil {
		t.Fatalf("listing the database's pods: %v", err)
	}
	for i := range pods.Items {
		if p := &pods.Items[i]; p.DeletionTimestamp == nil && podReady(p) {
			out, err := k.exec(ctx, namespace, p.Name, "postgres", nil, "pg_dumpall", "--data-only")
			if err != nil {
				t.Fatalf("dumping the database: %v", err)
			}
			return out
		}
	}
	t.Fatalf("%v: %s", errNoPod, databasePodSelector)
	return nil
}

// waitUninstalled waits until instance's Plugin, PluginGrant, credentials
// Secret, Deployment, pods, Service, NetworkPolicy, ServiceAccount and
// allow-list entry are all gone, and no Secret labelled as the instance's
// is left, of whatever role.
func waitUninstalled(t *testing.T, k *kube, instance string) {
	t.Helper()
	if err := untilUninstalled(t.Context(), k, instance); err != nil {
		t.Fatalf("waiting for instance %s to be uninstalled: %v", instance, err)
	}
}

// untilUninstalled is waitUninstalled's wait, for a ctx that outlives the
// test: a Cleanup runs after t.Context() is canceled.
func untilUninstalled(ctx context.Context, k *kube, instance string) error {
	return waitFor(ctx, uninstallTimeout, func(ctx context.Context) (string, bool, error) {
		left, err := k.instanceLeft(ctx, instance)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("%s: %s", errStillThere, strings.Join(left, ", ")), len(left) == 0, nil
	})
}

// instanceLeft names what is left of instance.
func (k *kube) instanceLeft(ctx context.Context, instance string) ([]string, error) {
	objects := plugins.ServiceAccountName(instance)
	var left []string
	_, err := k.dynamic.Resource(pluginResource()).Namespace(namespace).Get(ctx, instance, metav1.GetOptions{})
	left = appendIfPresent(left, "Plugin", err)
	_, err = k.dynamic.Resource(grantResource()).Namespace(namespace).Get(ctx, instance, metav1.GetOptions{})
	left = appendIfPresent(left, "PluginGrant", err)
	_, err = k.clients.CoreV1().Secrets(namespace).Get(ctx, credentialSecret(instance), metav1.GetOptions{})
	left = appendIfPresent(left, "Secret", err)
	_, err = k.clients.AppsV1().Deployments(namespace).Get(ctx, objects, metav1.GetOptions{})
	left = appendIfPresent(left, "Deployment", err)
	_, err = k.clients.CoreV1().Services(namespace).Get(ctx, objects, metav1.GetOptions{})
	left = appendIfPresent(left, "Service", err)
	_, err = k.clients.NetworkingV1().NetworkPolicies(namespace).Get(ctx, objects, metav1.GetOptions{})
	left = appendIfPresent(left, "NetworkPolicy", err)
	_, err = k.clients.CoreV1().ServiceAccounts(namespace).Get(ctx, objects, metav1.GetOptions{})
	left = appendIfPresent(left, "ServiceAccount", err)
	cm, err := k.clients.CoreV1().ConfigMaps(namespace).Get(ctx, plugins.EgressAllowListConfigMap, metav1.GetOptions{})
	switch {
	case apierrors.IsNotFound(err):
	case err != nil:
		return nil, fmt.Errorf("%w: getting the allow-list: %w", errRetry, err)
	default:
		if _, ok := cm.Data[instance]; ok {
			left = append(left, "allow-list entry")
		}
	}
	// The Deployment going is not its pods going, and a Secret of any role,
	// the CA bundle's too, goes with the instance when it carries the
	// instance's label (D-182).
	pods, err := k.clients.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: operator.LabelPlugin + "=" + instance})
	if err != nil {
		return nil, fmt.Errorf("%w: listing pods: %w", errRetry, err)
	}
	for i := range pods.Items {
		left = append(left, "pod "+pods.Items[i].Name)
	}
	secrets, err := k.clients.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{LabelSelector: plugins.OwnerInstanceLabel + "=" + instance})
	if err != nil {
		return nil, fmt.Errorf("%w: listing Secrets: %w", errRetry, err)
	}
	for i := range secrets.Items {
		left = append(left, "labelled Secret "+secrets.Items[i].Name)
	}
	return left, nil
}
