//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// The plugin instance TestOperatorDeploysSignedPlugin installs: the
// first-party NBP FX connector (plugins/nbp-fx), from the suite's registry.
const (
	// fxPlugin is the instance's name; the operator names its objects
	// plugin-<name>.
	fxPlugin = "nbp-e2e"
	// fxImage is its image in the suite's registry.
	fxImage = "nbp-fx"
	// fxManifest is the manifest the image's signed label carries, which the
	// Plugin must declare.
	fxManifest = "../../plugins/nbp-fx/manifest.json"
	// fxStart is the instance's config: a first sync from a day after today
	// asks NBP for nothing, so the suite never calls NBP's real API
	// (docs/dev/testing.md), while the core still calls the connector.
	fxStart = "2999-12-31"
	// fxEgress is the host the manifest requests and the grant approves.
	fxEgress = "api.nbp.pl"
	// fxGrantOnly is a host the grant approves that the manifest never
	// requests: the operator enforces only what both hold (D-16), so it is
	// never on the allow-list.
	fxGrantOnly = "example.org"
	// fxAllowList is the instance's entry in the allow-list ConfigMap.
	fxAllowList = "api.nbp.pl:443\n"
	// fxTokenSeconds is the projected tokens' lifetime the operator asks for.
	fxTokenSeconds = 3600
	// pluginPort is the port the operator gives every plugin's Service
	// (deploy/operator's --plugin-port).
	pluginPort = 8080
	// pluginTimeout bounds the wait for the instance to run, from the
	// operator's own start.
	pluginTimeout = 5 * time.Minute
	// syncTimeout bounds the retries of the core's sync of the instance.
	syncTimeout = 2 * time.Minute
	// uninstallTimeout bounds the wait for the instance's objects to go.
	uninstallTimeout = 2 * time.Minute
)

var (
	// errNotVerified reports a plugin whose image signature the operator
	// refused: waiting won't change that.
	errNotVerified = errors.New("signature not verified")
	// errStillThere reports an object that should be gone.
	errStillThere = errors.New("still there")
)

// TestOperatorDeploysSignedPlugin installs a plugin the way an owner does,
// with a Plugin and a PluginGrant, from an image built from this checkout and
// signed with the suite's key, a trusted signer (D-45, D-57), and checks
// what the operator makes of it: the signature verified and reported in the
// status (D-56); a workload with the hardened security context and only
// projected tokens (D-27); a NetworkPolicy that lets the core call the
// plugin and no other pod (D-59); an egress allow-list with exactly the
// host both the manifest and the grant hold (D-16, D-32); and, once the
// Plugin is deleted, no workload and no allow-list entry
// (SCEN-e2e-plugin-signature-verified, SCEN-e2e-plugin-workload-hardened,
// SCEN-e2e-plugin-only-core-calls, SCEN-e2e-plugin-egress-allow-list,
// SCEN-e2e-plugin-uninstalled).
func TestOperatorDeploysSignedPlugin(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	img, err := reg.image(t.Context(), fxImage)
	if err != nil {
		t.Fatal(err)
	}
	installFX(t, k, img.cluster)
	p := waitPluginReady(t, k, fxPlugin)
	checkSignature(t, p)
	checkWorkload(t, k, img.cluster)
	waitCoreAvailable(t, k)
	checkCoreSyncs(t, k)
	checkOthersRefused(t, k)
	checkAllowList(t, k)
	uninstallFX(t, k)
}

// installFX applies the instance's PluginGrant and Plugin: the manifest is
// the image's own, read from the checkout.
func installFX(t *testing.T, k *kube, image plugins.ImageRef) {
	t.Helper()
	data, err := os.ReadFile(fxManifest)
	if err != nil {
		t.Fatalf("reading the manifest: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decoding the manifest: %v", err)
	}
	grant := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "openfief.io/v1alpha1", "kind": "PluginGrant",
		"metadata": map[string]any{"name": fxPlugin, "namespace": namespace},
		"spec": map[string]any{
			"capabilities": []any{"connector"},
			"egress":       []any{fxEgress, fxGrantOnly},
		},
	}}
	plugin := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "openfief.io/v1alpha1", "kind": "Plugin",
		"metadata": map[string]any{"name": fxPlugin, "namespace": namespace},
		"spec": map[string]any{
			"image":    image.String(),
			"manifest": manifest,
			"config":   map[string]any{"start": fxStart},
		},
	}}
	for _, obj := range []*unstructured.Unstructured{grant, plugin} {
		if err := k.applyOne(t.Context(), obj); err != nil {
			t.Fatal(err)
		}
	}
}

// getPlugin returns Plugin name of the openfief namespace.
func (k *kube) getPlugin(ctx context.Context, name string) (*v1alpha1.Plugin, error) {
	u, err := k.dynamic.Resource(pluginResource()).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting plugin %s: %w", name, err)
	}
	var p v1alpha1.Plugin
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &p); err != nil {
		return nil, fmt.Errorf("decoding plugin %s: %w", name, err)
	}
	return &p, nil
}

// waitPluginReady waits until Plugin name of the current generation is
// Ready, failing at once when the operator refuses its signature or its pod
// can't start, and returns it.
func waitPluginReady(t *testing.T, k *kube, name string) *v1alpha1.Plugin {
	t.Helper()
	var ready *v1alpha1.Plugin
	err := waitFor(t.Context(), pluginTimeout, func(ctx context.Context) (string, bool, error) {
		p, err := k.getPlugin(ctx, name)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		state := conditionsState(p)
		sig := meta.FindStatusCondition(p.Status.Conditions, v1alpha1.ConditionSignatureVerified)
		if sig != nil && sig.ObservedGeneration == p.Generation && sig.Status == metav1.ConditionFalse {
			return state, false, fmt.Errorf("%w: %s: %s", errNotVerified, sig.Reason, sig.Message)
		}
		r := meta.FindStatusCondition(p.Status.Conditions, v1alpha1.ConditionReady)
		if r != nil && r.ObservedGeneration == p.Generation && r.Status == metav1.ConditionTrue {
			ready = p
			return state, true, nil
		}
		pods, err := k.pluginPodsState(ctx, name)
		return state + "; " + pods, false, err
	})
	if err != nil {
		t.Fatalf("waiting for plugin %s to be ready: %v", name, err)
	}
	return ready
}

// conditionsState describes a Plugin's conditions in a line.
func conditionsState(p *v1alpha1.Plugin) string {
	parts := make([]string, 0, len(p.Status.Conditions))
	for _, c := range p.Status.Conditions {
		parts = append(parts, fmt.Sprintf("%s=%s (%s: %s)", c.Type, c.Status, c.Reason, c.Message))
	}
	if len(parts) == 0 {
		return "no conditions yet"
	}
	return strings.Join(parts, "; ")
}

// pluginPodsState summarizes plugin name's pods, and fails when one of
// them is stuck.
func (k *kube) pluginPodsState(ctx context.Context, name string) (string, error) {
	pods, err := k.clients.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: operator.LabelPlugin + "=" + name})
	if err != nil {
		return "", fmt.Errorf("%w: listing the pods of plugin %s: %w", errRetry, name, err)
	}
	parts := make([]string, 0, len(pods.Items))
	for i := range pods.Items {
		part, err := podState(&pods.Items[i])
		if err != nil {
			return part, err
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return "no pods yet", nil
	}
	return strings.Join(parts, "; "), nil
}

// checkSignature checks that the operator verified the image as signed by
// the suite's key, an owner's trusted signer, which makes the plugin
// unverified but runnable (D-44, D-45), and that it runs exactly what the
// manifest requests and the grant approves (D-16).
func checkSignature(t *testing.T, p *v1alpha1.Plugin) {
	t.Helper()
	sig := meta.FindStatusCondition(p.Status.Conditions, v1alpha1.ConditionSignatureVerified)
	if sig == nil || sig.Status != metav1.ConditionTrue || sig.Reason != operator.ReasonTrustedSigner {
		t.Errorf("condition %s: %+v; want True, reason %s", v1alpha1.ConditionSignatureVerified, sig, operator.ReasonTrustedSigner)
	}
	want := &v1alpha1.SignatureStatus{
		Trust: string(plugins.Unverified), SignerKind: string(plugins.SignerOwner), Signer: signerName,
	}
	if diff := cmp.Diff(want, p.Status.Signature); diff != "" {
		t.Errorf("status.signature (-want +got):\n%s", diff)
	}
	effective := &v1alpha1.Permissions{Capabilities: []string{"connector"}, Egress: []string{fxEgress + ":443"}}
	if diff := cmp.Diff(effective, p.Status.Effective); diff != "" {
		t.Errorf("status.effective (-want +got):\n%s", diff)
	}
}

// workload is what checkWorkload compares of a plugin's running pod: who it
// runs as, what it may do, and which tokens it holds.
type workload struct {
	ServiceAccount           string
	AutomountToken           string
	ServiceLinks             string
	PodRunAsNonRoot          string
	Seccomp                  corev1.SeccompProfileType
	Image                    string
	Privileged               string
	AllowPrivilegeEscalation string
	ReadOnlyRootFilesystem   string
	RunAsNonRoot             string
	DropCapabilities         []corev1.Capability
	AddCapabilities          []corev1.Capability
	// TokenAudiences are the audiences of the projected ServiceAccount
	// tokens, sorted; TokenLifetimes their requested lifetimes.
	TokenAudiences []string
	TokenLifetimes []int64
	// OtherVolumes are the volumes that are neither projected tokens nor
	// ConfigMaps, such as a Secret or an automounted API token.
	OtherVolumes []string
}

// boolField describes an optional boolean: "unset" when the API server
// defaults it.
func boolField(b *bool) string {
	if b == nil {
		return "unset"
	}
	return strconv.FormatBool(*b)
}

// workloadOf extracts a workload from a plugin pod.
func workloadOf(p *corev1.Pod) workload {
	w := workload{
		ServiceAccount: p.Spec.ServiceAccountName,
		AutomountToken: boolField(p.Spec.AutomountServiceAccountToken),
		ServiceLinks:   boolField(p.Spec.EnableServiceLinks),
	}
	if sc := p.Spec.SecurityContext; sc != nil {
		w.PodRunAsNonRoot = boolField(sc.RunAsNonRoot)
		if sc.SeccompProfile != nil {
			w.Seccomp = sc.SeccompProfile.Type
		}
	}
	for i := range p.Spec.Containers {
		if c := &p.Spec.Containers[i]; c.Name == operator.ContainerName {
			w.addContainer(c)
		}
	}
	w.addVolumes(p.Spec.Volumes)
	return w
}

// addContainer records the plugin container's image and security context.
func (w *workload) addContainer(c *corev1.Container) {
	w.Image = c.Image
	sc := c.SecurityContext
	if sc == nil {
		return
	}
	w.Privileged, w.AllowPrivilegeEscalation = boolField(sc.Privileged), boolField(sc.AllowPrivilegeEscalation)
	w.ReadOnlyRootFilesystem, w.RunAsNonRoot = boolField(sc.ReadOnlyRootFilesystem), boolField(sc.RunAsNonRoot)
	if sc.Capabilities != nil {
		w.DropCapabilities, w.AddCapabilities = sc.Capabilities.Drop, sc.Capabilities.Add
	}
}

// addVolumes records the pod's projected tokens and the volumes that are
// neither tokens nor ConfigMaps.
func (w *workload) addVolumes(volumes []corev1.Volume) {
	for i := range volumes {
		v := &volumes[i]
		switch {
		case v.ConfigMap != nil:
		case v.Projected != nil && len(v.Projected.Sources) == 1 && v.Projected.Sources[0].ServiceAccountToken != nil:
			token := v.Projected.Sources[0].ServiceAccountToken
			w.TokenAudiences = append(w.TokenAudiences, token.Audience)
			if token.ExpirationSeconds != nil {
				w.TokenLifetimes = append(w.TokenLifetimes, *token.ExpirationSeconds)
			}
		default:
			w.OtherVolumes = append(w.OtherVolumes, v.Name)
		}
	}
	slices.Sort(w.TokenAudiences)
}

// checkWorkload checks the running pod of the instance: its own
// ServiceAccount with no API token, the restricted security context, the
// signed image, and two projected tokens, one for the core and one for the
// egress proxy (D-27, D-63).
func checkWorkload(t *testing.T, k *kube, image plugins.ImageRef) {
	t.Helper()
	pod, err := k.runningPluginPod(t.Context(), fxPlugin)
	if err != nil {
		t.Fatal(err)
	}
	want := workload{
		ServiceAccount: plugins.ServiceAccountName(fxPlugin), AutomountToken: "false", ServiceLinks: "false",
		PodRunAsNonRoot: "true", Seccomp: corev1.SeccompProfileTypeRuntimeDefault, Image: image.String(),
		Privileged: "false", AllowPrivilegeEscalation: "false", ReadOnlyRootFilesystem: "true", RunAsNonRoot: "true",
		DropCapabilities: []corev1.Capability{"ALL"},
		TokenAudiences:   []string{plugins.CoreAudience, plugins.EgressAudience},
		TokenLifetimes:   []int64{fxTokenSeconds, fxTokenSeconds},
	}
	if diff := cmp.Diff(want, workloadOf(pod)); diff != "" {
		t.Errorf("pod %s (-want +got):\n%s", pod.Name, diff)
	}
}

// runningPluginPod returns the running, ready pod of plugin name.
func (k *kube) runningPluginPod(ctx context.Context, name string) (*corev1.Pod, error) {
	pods, err := k.clients.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: operator.LabelPlugin + "=" + name})
	if err != nil {
		return nil, fmt.Errorf("listing the pods of plugin %s: %w", name, err)
	}
	for i := range pods.Items {
		if p := &pods.Items[i]; p.DeletionTimestamp == nil && podReady(p) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("%w: plugin %s", errNoPod, name)
}

// checkCoreSyncs syncs the instance with the core's own CLI inside its pod,
// as an owner would with kubectl exec: the core finds the instance's
// Service, calls it with a token for that instance through both
// NetworkPolicies, and the connector answers with an empty delivery. The
// core's scheduler may be syncing the instance at the same moment, which
// makes a sync fail, so a failure is tried again until the deadline.
func checkCoreSyncs(t *testing.T, k *kube) {
	t.Helper()
	err := waitFor(t.Context(), syncTimeout, func(ctx context.Context) (string, bool, error) {
		pod, err := k.runningPod(ctx, namespace, "openfief-core")
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		out, err := k.exec(ctx, namespace, pod, "core", nil, "/core", "connector", "sync", "--instance", fxPlugin)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		if !bytes.Contains(out, []byte(": 0 transactions,")) {
			return "", false, fmt.Errorf("%w: unexpected output %q", errRetry, out)
		}
		return string(out), true, nil
	})
	if err != nil {
		t.Fatalf("core connector sync --instance %s: %v", fxPlugin, err)
	}
}

// checkOthersRefused checks that no pod but the core reaches the instance:
// neither the client pod, which a policy admits to the core, nor the
// outsider (D-59). The core's sync showed the plugin listening.
func checkOthersRefused(t *testing.T, k *kube) {
	t.Helper()
	url := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d/", plugins.ServiceAccountName(fxPlugin), namespace, pluginPort)
	for _, pod := range []string{clientPod, outsiderPod} {
		if r := answerAt(t, k, pod, url); r.status != 0 {
			t.Errorf("GET %s from %s: status %d; want no connection", url, pod, r.status)
		}
	}
}

// checkAllowList checks the instance's entry in the egress proxy's
// allow-list: the one host both the manifest and the grant hold.
func checkAllowList(t *testing.T, k *kube) {
	t.Helper()
	cm, err := k.clients.CoreV1().ConfigMaps(namespace).Get(t.Context(), plugins.EgressAllowListConfigMap, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting the allow-list: %v", err)
	}
	got, ok := cm.Data[fxPlugin]
	if !ok || got != fxAllowList {
		t.Errorf("allow-list entry %s: %q (present %t); want %q", fxPlugin, got, ok, fxAllowList)
	}
}

// uninstallFX deletes the instance's Plugin and waits until its workload,
// Service, NetworkPolicy and allow-list entry are gone, then deletes its
// grant.
func uninstallFX(t *testing.T, k *kube) {
	t.Helper()
	ctx := t.Context()
	if err := k.dynamic.Resource(pluginResource()).Namespace(namespace).Delete(ctx, fxPlugin, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("deleting plugin %s: %v", fxPlugin, err)
	}
	objects := plugins.ServiceAccountName(fxPlugin)
	err := waitFor(ctx, uninstallTimeout, func(ctx context.Context) (string, bool, error) {
		var left []string
		_, err := k.clients.AppsV1().Deployments(namespace).Get(ctx, objects, metav1.GetOptions{})
		left = appendIfPresent(left, "Deployment", err)
		_, err = k.clients.CoreV1().Services(namespace).Get(ctx, objects, metav1.GetOptions{})
		left = appendIfPresent(left, "Service", err)
		_, err = k.clients.NetworkingV1().NetworkPolicies(namespace).Get(ctx, objects, metav1.GetOptions{})
		left = appendIfPresent(left, "NetworkPolicy", err)
		cm, err := k.clients.CoreV1().ConfigMaps(namespace).Get(ctx, plugins.EgressAllowListConfigMap, metav1.GetOptions{})
		if err != nil {
			return "", false, fmt.Errorf("%w: getting the allow-list: %w", errRetry, err)
		}
		if _, ok := cm.Data[fxPlugin]; ok {
			left = append(left, "allow-list entry")
		}
		return fmt.Sprintf("%s: %s", errStillThere, strings.Join(left, ", ")), len(left) == 0, nil
	})
	if err != nil {
		t.Fatalf("waiting for plugin %s's objects to go: %v", fxPlugin, err)
	}
	if err := k.dynamic.Resource(grantResource()).Namespace(namespace).Delete(ctx, fxPlugin, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("deleting plugin grant %s: %v", fxPlugin, err)
	}
}

// appendIfPresent appends kind to left unless err says it is not found;
// any other error counts as present, to be checked again.
func appendIfPresent(left []string, kind string, err error) []string {
	if apierrors.IsNotFound(err) {
		return left
	}
	return append(left, kind)
}
