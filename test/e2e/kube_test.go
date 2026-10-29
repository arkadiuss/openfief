//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/yaml"

	"github.com/arkadiuss/openfief/internal/plugins"
)

const (
	// namespace is OpenFief's namespace (deploy/crds).
	namespace = "openfief"
	// fieldManager owns what the suite applies.
	fieldManager = "openfief-e2e"
	// pollEvery is how often a wait checks its condition again.
	pollEvery = 2 * time.Second
	// diagnosticsTimeout bounds the calls that collect diagnostics.
	diagnosticsTimeout = 2 * time.Minute
	// logTail is how many lines of each container's log the diagnostics keep.
	logTail = 300
	// yamlBuffer is the decoder's look-ahead for telling YAML from JSON.
	yamlBuffer = 4096
)

var (
	// errBuild reports a kustomize build that failed.
	errBuild = errors.New("kustomize build failed")
	// errTimeout reports a condition still unmet at its deadline.
	errTimeout = errors.New("timed out")
	// errRetry marks a check's error as a passing state: the check runs again.
	errRetry = errors.New("not yet")
	// errStuck reports a pod that waits for something waiting won't bring.
	errStuck = errors.New("pod cannot start")
	// errNoPod reports a label with no running pod to exec into.
	errNoPod = errors.New("no running pod")
)

// kube is the cluster of the current kubeconfig context, the one
// scripts/assert-local-cluster.sh accepted.
type kube struct {
	config  *rest.Config
	clients kubernetes.Interface
	dynamic dynamic.Interface
	mapper  *restmapper.DeferredDiscoveryRESTMapper
}

// connect returns the cluster of the current kubeconfig context.
func connect() (*kube, error) {
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{},
	)
	cfg, err := loader.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("kubeconfig: %w", err)
	}
	clients, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("kubernetes client: %w", err)
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("dynamic client: %w", err)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(clients.Discovery()))
	return &kube{config: cfg, clients: clients, dynamic: dyn, mapper: mapper}, nil
}

// mustConnect is connect that fails t on an error.
func mustConnect(t *testing.T) *kube {
	t.Helper()
	k, err := connect()
	if err != nil {
		t.Fatal(err)
	}
	return k
}

// build returns what cmd, a kustomize build, prints. Each caller names its
// directory as a constant, so no command line carries a variable.
func build(cmd *exec.Cmd) ([]byte, error) {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w: %s", errBuild, strings.Join(cmd.Args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

// applyBuild applies what cmd, a kustomize build, prints.
func (k *kube) applyBuild(ctx context.Context, cmd *exec.Cmd) error {
	manifests, err := build(cmd)
	if err != nil {
		return err
	}
	return k.apply(ctx, manifests)
}

// apply applies every object in manifests, YAML or JSON, server-side, as
// Flux applies the demo's layers.
func (k *kube) apply(ctx context.Context, manifests []byte) error {
	return k.applyEdited(ctx, manifests, nil)
}

// applyEdited is apply with each object changed by edit first, unless edit
// is nil.
func (k *kube) applyEdited(ctx context.Context, manifests []byte, edit func(*unstructured.Unstructured) error) error {
	dec := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifests), yamlBuffer)
	for {
		var obj unstructured.Unstructured
		err := dec.Decode(&obj.Object)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("decoding manifests: %w", err)
		}
		if len(obj.Object) == 0 {
			continue
		}
		if edit != nil {
			if err := edit(&obj); err != nil {
				return err
			}
		}
		if err := k.applyOne(ctx, &obj); err != nil {
			return err
		}
	}
}

// applyOne applies obj server-side, forcing ownership of its fields.
func (k *kube) applyOne(ctx context.Context, obj *unstructured.Unstructured) error {
	gvk := obj.GroupVersionKind()
	mapping, err := k.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if meta.IsNoMatchError(err) {
		k.mapper.Reset() // a CRD applied earlier in the run
		mapping, err = k.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	}
	if err != nil {
		return fmt.Errorf("mapping %s: %w", gvk, err)
	}
	var res dynamic.ResourceInterface = k.dynamic.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		ns := obj.GetNamespace()
		if ns == "" {
			ns = metav1.NamespaceDefault
		}
		res = k.dynamic.Resource(mapping.Resource).Namespace(ns)
	}
	opts := metav1.ApplyOptions{FieldManager: fieldManager, Force: true}
	if _, err := res.Apply(ctx, obj.GetName(), obj, opts); err != nil {
		return fmt.Errorf("applying %s %s/%s: %w", gvk.Kind, obj.GetNamespace(), obj.GetName(), err)
	}
	return nil
}

// exec runs command in container of pod in ns with stdin and returns its
// stdout, also when it fails. Errors never carry stdin, which may hold a
// token.
func (k *kube) exec(ctx context.Context, ns, pod, container string, stdin []byte, command ...string) ([]byte, error) {
	req := k.clients.CoreV1().RESTClient().Post().Resource("pods").Namespace(ns).Name(pod).SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container, Command: command, Stdin: stdin != nil, Stdout: true, Stderr: true,
		}, scheme.ParameterCodec)
	ex, err := remotecommand.NewWebSocketExecutor(k.config, "GET", req.URL().String())
	if err != nil {
		return nil, fmt.Errorf("exec in %s/%s: %w", ns, pod, err)
	}
	var stdout, stderr bytes.Buffer
	opts := remotecommand.StreamOptions{Stdout: &stdout, Stderr: &stderr}
	if stdin != nil {
		opts.Stdin = bytes.NewReader(stdin)
	}
	if err := ex.StreamWithContext(ctx, opts); err != nil {
		return stdout.Bytes(), fmt.Errorf("exec %s in %s/%s: %w: %s", command[0], ns, pod, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

// runningPod returns the name of a running, ready pod in ns with label
// app.kubernetes.io/name=name.
func (k *kube) runningPod(ctx context.Context, ns, name string) (string, error) {
	pods, err := k.clients.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=" + name})
	if err != nil {
		return "", fmt.Errorf("listing %s pods: %w", name, err)
	}
	for i := range pods.Items {
		if p := &pods.Items[i]; p.DeletionTimestamp == nil && podReady(p) {
			return p.Name, nil
		}
	}
	return "", fmt.Errorf("%w: %s/%s", errNoPod, ns, name)
}

// conditionTrue reports whether the object of resource ns/name has a status
// condition of type cond that is True.
func (k *kube) conditionTrue(ctx context.Context, resource schema.GroupVersionResource, ns, name, cond string) (bool, error) {
	obj, err := k.dynamic.Resource(resource).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("%w: getting %s %s/%s: %w", errRetry, resource.Resource, ns, name, err)
	}
	conditions, _, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil {
		return false, fmt.Errorf("%s %s/%s: conditions: %w", resource.Resource, ns, name, err)
	}
	for _, c := range conditions {
		m, ok := c.(map[string]any)
		if ok && m["type"] == cond {
			return m["status"] == string(metav1.ConditionTrue), nil
		}
	}
	return false, nil
}

// retry calls f until it succeeds or timeout passes.
func retry(ctx context.Context, timeout time.Duration, f func() error) error {
	return waitFor(ctx, timeout, func(context.Context) (string, bool, error) {
		if err := f(); err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		return "", true, nil
	})
}

// check reports the current state of a condition in words and whether it
// holds. An error wrapping errRetry is a state too; any other error means
// the condition can no longer hold.
type check func(ctx context.Context) (state string, done bool, err error)

// waitFor checks c every pollEvery until it holds, it fails, or timeout
// passes; then the error carries the last state seen.
func waitFor(ctx context.Context, timeout time.Duration, c check) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	tick := time.NewTicker(pollEvery)
	defer tick.Stop()
	for {
		state, done, err := c(ctx)
		switch {
		case errors.Is(err, errRetry):
			state = err.Error()
		case err != nil:
			return err
		case done:
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w after %s; last state: %s", errTimeout, timeout, state)
		case <-tick.C:
		}
	}
}

// stuckReasons are container waiting reasons that never clear by waiting:
// the pod needs a fixed image or manifest.
func stuckReasons() []string {
	return []string{
		"CreateContainerConfigError", "CreateContainerError", "InvalidImageName",
		"ErrImagePull", "ImagePullBackOff", "ErrImageNeverPull", "CrashLoopBackOff",
	}
}

// podsState summarizes the pods in ns with label app.kubernetes.io/name=name,
// and fails when one of them is stuck.
func (k *kube) podsState(ctx context.Context, ns, name string) (string, error) {
	pods, err := k.clients.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=" + name})
	if err != nil {
		return "", fmt.Errorf("%w: listing %s pods: %w", errRetry, name, err)
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

// podState describes p's phase and waiting containers, and fails when one
// of them is stuck.
func podState(p *corev1.Pod) (string, error) {
	part := p.Name + " " + string(p.Status.Phase)
	for i := range p.Status.ContainerStatuses {
		w := p.Status.ContainerStatuses[i].State.Waiting
		if w == nil {
			continue
		}
		part += " " + w.Reason
		if slices.Contains(stuckReasons(), w.Reason) {
			return part, fmt.Errorf("%w: %s: %s: %s", errStuck, p.Name, w.Reason, w.Message)
		}
	}
	return part, nil
}

// podReady reports whether p's Ready condition is true.
func podReady(p *corev1.Pod) bool {
	for _, c := range p.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

// diagnoseOnFailure logs the cluster's state at the end of t if t failed.
func diagnoseOnFailure(t *testing.T, k *kube) {
	t.Helper()
	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), diagnosticsTimeout)
		defer cancel()
		k.dumpDiagnostics(ctx, t.Logf)
	})
}

// diagnosedNamespaces are the namespaces whose pods the diagnostics show in
// full.
func diagnosedNamespaces() []string {
	return []string{namespace, "cnpg-system", clientNamespace}
}

// dumpDiagnostics logs what explains a failure: every pod and event in the
// cluster, the network and plugin objects in OpenFief's namespace, the spec
// and status of each pod in OpenFief's, CloudNativePG's and the clients'
// namespaces that is not ready, and the logs of every pod there: the core,
// the operator, the plugins, the demo loader, the database and the
// CloudNativePG operator among them.
func (k *kube) dumpDiagnostics(ctx context.Context, logf func(string, ...any)) {
	pods, err := k.clients.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		logf("diagnostics: listing pods: %v", err)
		return
	}
	var b strings.Builder
	for i := range pods.Items {
		p := &pods.Items[i]
		fmt.Fprintf(&b, "%s/%s ready=%t node=%s %s\n", p.Namespace, p.Name, podReady(p), p.Spec.NodeName, podSummary(p))
	}
	logf("diagnostics: pods\n%s", b.String())
	k.dumpEvents(ctx, logf)
	k.dumpNetwork(ctx, logf)
	k.dumpPlugins(ctx, logf)
	for i := range pods.Items {
		k.dumpPod(ctx, logf, &pods.Items[i])
	}
}

// podSummary is podState's description, with why the pod is stuck if it is.
func podSummary(p *corev1.Pod) string {
	state, err := podState(p)
	if err != nil {
		return state + " (" + err.Error() + ")"
	}
	return state
}

// dumpPod logs p's logs if it is Traefik or in a diagnosed namespace, and
// also its spec and status if it is the core or not ready.
func (k *kube) dumpPod(ctx context.Context, logf func(string, ...any), p *corev1.Pod) {
	name := p.Labels["app.kubernetes.io/name"]
	traefik := p.Namespace == metav1.NamespaceSystem && name == "traefik"
	if !traefik && !slices.Contains(diagnosedNamespaces(), p.Namespace) {
		return
	}
	if name == "openfief-core" || (!podReady(p) && p.Status.Phase != corev1.PodSucceeded) {
		p.ManagedFields = nil
		dumpYAML(logf, fmt.Sprintf("pod %s/%s (ready=%t), spec and status", p.Namespace, p.Name, podReady(p)), p, nil)
	}
	k.dumpLogs(ctx, logf, p)
}

// dumpNetwork logs what decides whether traffic reaches the core: the
// NetworkPolicies in OpenFief's namespace, the core Service's
// EndpointSlices, which Traefik routes to, and the Ingresses.
func (k *kube) dumpNetwork(ctx context.Context, logf func(string, ...any)) {
	policies, err := k.clients.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	dumpYAML(logf, "NetworkPolicies in "+namespace, policies, err)
	endpoints, err := k.clients.DiscoveryV1().EndpointSlices(namespace).List(ctx,
		metav1.ListOptions{LabelSelector: "kubernetes.io/service-name=openfief-core"})
	dumpYAML(logf, "EndpointSlices of Service openfief-core", endpoints, err)
	ingresses, err := k.clients.NetworkingV1().Ingresses(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	dumpYAML(logf, "Ingresses", ingresses, err)
}

// dumpPlugins logs what decides whether a plugin runs: the Plugins and
// PluginGrants with their status, the operator's Deployments for them, the
// egress allow-lists and the trusted signers (public keys only).
func (k *kube) dumpPlugins(ctx context.Context, logf func(string, ...any)) {
	for _, r := range []schema.GroupVersionResource{pluginResource(), grantResource()} {
		list, err := k.dynamic.Resource(r).Namespace(namespace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for i := range list.Items {
				list.Items[i].SetManagedFields(nil)
			}
		}
		dumpYAML(logf, r.Resource+" in "+namespace, list, err)
	}
	deployments, err := k.clients.AppsV1().Deployments(namespace).List(ctx,
		metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=openfief-plugin"})
	if err == nil {
		for i := range deployments.Items {
			deployments.Items[i].ManagedFields = nil
		}
	}
	dumpYAML(logf, "plugin Deployments", deployments, err)
	for _, name := range []string{plugins.EgressAllowListConfigMap, trustedSignersConfigMap} {
		cm, err := k.clients.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
		if err == nil {
			cm.ManagedFields = nil
		}
		dumpYAML(logf, "ConfigMap "+name, cm, err)
	}
}

// dumpYAML logs v as YAML under title, or err.
func dumpYAML(logf func(string, ...any), title string, v any, err error) {
	if err != nil {
		logf("diagnostics: %s: %v", title, err)
		return
	}
	out, err := yaml.Marshal(v)
	if err != nil {
		logf("diagnostics: encoding %s: %v", title, err)
		return
	}
	logf("diagnostics: %s\n%s", title, out)
}

// dumpEvents logs every event in the cluster, oldest first.
func (k *kube) dumpEvents(ctx context.Context, logf func(string, ...any)) {
	events, err := k.clients.CoreV1().Events(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		logf("diagnostics: listing events: %v", err)
		return
	}
	slices.SortFunc(events.Items, func(a, b corev1.Event) int {
		return a.LastTimestamp.Compare(b.LastTimestamp.Time)
	})
	var b strings.Builder
	for i := range events.Items {
		e := &events.Items[i]
		fmt.Fprintf(&b, "%s %s %s/%s %s: %s\n", e.LastTimestamp.Format(time.RFC3339), e.Type,
			e.InvolvedObject.Kind, e.InvolvedObject.Name, e.Reason, e.Message)
	}
	logf("diagnostics: events\n%s", b.String())
}

// dumpLogs logs the tail of each container's log in p, and of its previous
// run when it restarted.
func (k *kube) dumpLogs(ctx context.Context, logf func(string, ...any), p *corev1.Pod) {
	tail := int64(logTail)
	for i := range p.Status.ContainerStatuses {
		cs := &p.Status.ContainerStatuses[i]
		for _, previous := range []bool{false, true} {
			if previous && cs.RestartCount == 0 {
				continue
			}
			out, err := k.clients.CoreV1().Pods(p.Namespace).GetLogs(p.Name, &corev1.PodLogOptions{
				Container: cs.Name, TailLines: &tail, Previous: previous,
			}).DoRaw(ctx)
			if err != nil {
				logf("diagnostics: logs of %s/%s %s (previous=%t): %v", p.Namespace, p.Name, cs.Name, previous, err)
				continue
			}
			logf("diagnostics: logs of %s/%s %s (previous=%t)\n%s", p.Namespace, p.Name, cs.Name, previous, out)
		}
	}
}
