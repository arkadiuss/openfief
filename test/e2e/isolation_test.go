//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/arkadiuss/openfief/internal/plugins"
	pluginskube "github.com/arkadiuss/openfief/internal/plugins/kube"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/smokescreen"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// The plugin instances TestPluginIsolation installs, and where it finds
// what it probes.
const (
	// isoXTB runs the first-party XTB connector (plugins/xtb), a file-only
	// connector whose manifest requests no egress host.
	isoXTB = "xtb-iso"
	// isoEgress runs the NBP FX connector like fxPlugin: its grant approves
	// api.nbp.pl, which the manifest requests, and example.org, which it
	// doesn't, and its config asks NBP for nothing.
	isoEgress = "egress-iso"
	// xtbImage is the XTB connector's image in the suite's registry.
	xtbImage = "xtb"
	// xtbManifest is the manifest the XTB image's signed label carries.
	xtbManifest = "../../plugins/xtb/manifest.json"
	// xtbInstitution is the institution the XTB manifest declares (D-185).
	xtbInstitution = "xtb"
	// coreDeployment and coreService are the core's (deploy/core).
	coreDeployment = "openfief-core"
	coreService    = "openfief-core"
	// proxyService is the egress proxy's (deploy/egress-proxy).
	proxyService = "egress-proxy"
	// clientContainer is the client pods' container (testdata/clients.yaml).
	clientContainer = "curl"
)

// How TestPluginIsolation probes from inside a pod.
const (
	// probePrefix starts the name of the ephemeral container the test adds
	// to a pod to probe from it; a time suffix keeps it unique in a cluster
	// that a run before used.
	probePrefix = "iso-probe-"
	// probeUser is the numeric user the probe runs as: the curl image names
	// its user, which runAsNonRoot can't check.
	probeUser = 65532
	// probeSeconds is how long a probe container stays up.
	probeSeconds = "3600"
	// probeMaxTime bounds one curl call in seconds, name lookup included.
	probeMaxTime = "20"
	// probeWriteOut is what curl writes after each call: the HTTP status,
	// the proxy's answer to CONNECT, curl's exit code and its error.
	probeWriteOut = "\n%{http_code} %{http_connect} %{exitcode} %{errormsg}\n"
	// probeTimeout bounds the start of a probe container and the retries of
	// one call whose exec stream failed.
	probeTimeout = 3 * time.Minute
	// tokenSeconds is the lifetime of the tokens the test requests, the
	// shortest the API server issues.
	tokenSeconds = 600
	// proxyLogTimeout bounds the wait for the proxy to log the test's
	// requests.
	proxyLogTimeout = 2 * time.Minute
	// canary is in the path, query, header and body of a request the test
	// sends through the proxy; no proxy log line may hold it (GR-12).
	canary = "openfief-iso-canary"
)

// curl's exit codes that the probes tell apart.
const (
	curlCouldntResolve = 6
	curlCouldntConnect = 7
	curlTimedOut       = 28
	curlProxyError     = 97
)

var (
	// errNoProbeOutput reports curl output without the write-out line.
	errNoProbeOutput = errors.New("no write-out line in curl's output")
	// errProbeStopped reports a probe container that ended.
	errProbeStopped = errors.New("probe container stopped")
	// errNoEndpoint reports a Service with no ready endpoint address.
	errNoEndpoint = errors.New("no ready endpoint")
	// errNoPort reports a Service without the named port.
	errNoPort = errors.New("no such port")
	// resolvingTimedOut is curl's error for a name lookup that timed out.
	resolvingTimedOut = regexp.MustCompile(`^Resolving timed out`)
)

// TestPluginIsolation checks plugin isolation, egress through the proxy and
// the tokens on both sides of the core's calls (D-200 criteria 10, 16 and
// 17). It installs the XTB connector, which has no egress, and an NBP FX
// connector granted api.nbp.pl, and probes from inside their pods, and the
// core's, with an ephemeral container: it shares the pod's network, so the
// pod's NetworkPolicy applies to it as to the plugin. A pod no policy
// restricts, the suite's outsider, shows each blocked destination reachable
// otherwise. It checks that each plugin runs in its own pod; that the XTB
// pod reaches nothing but the core's plugin port, neither the proxy nor an
// outside address nor DNS, and that the operator gave it no proxy rule, no
// egress token and an empty allow-list (SCEN-e2e-xtb-no-egress); that the
// proxy refuses the granted plugin's CONNECT to hosts outside its effective
// grant and a CONNECT without a token (SCEN-e2e-egress-outside-grant-refused);
// that a direct connection and a DNS lookup from the plugin pods fail
// (SCEN-e2e-egress-bypass-fails); that the proxy logs each of those
// requests with host, decision and instance and nothing of a path, query,
// header, body or token, and allowed nothing outside an allow-list
// (SCEN-e2e-egress-proxy-log); that a connector answers the core's valid
// token for its instance and refuses a call without one, with a malformed
// one, with one for another instance and with a plugin's own, while no other
// pod reaches it at all (SCEN-e2e-connector-refuses-bad-core-token); and
// that the core's plugin port refuses the same kinds of bad tokens on the
// read API and the AuthService, and its browser port a view plugin's token
// (SCEN-e2e-core-refuses-bad-plugin-tokens).
func TestPluginIsolation(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	xtb, err := reg.image(t.Context(), xtbImage)
	if err != nil {
		t.Fatal(err)
	}
	fx, err := reg.image(t.Context(), fxImage)
	if err != nil {
		t.Fatal(err)
	}
	installXTB(t, k, xtb.cluster)
	// isoEgress asks for the exempt host as its baseUrl's host, and isoPrivate
	// for another Service's, in the same private range (egressallowed_test.go).
	installFXAt(t, k, isoEgress, fx.cluster, egressFixtureHost)
	installFXAt(t, k, isoPrivate, fx.cluster, egressOtherHost)
	waitPluginReady(t, k, isoXTB)
	waitPluginReady(t, k, isoEgress)
	waitPluginReady(t, k, isoPrivate)
	waitPluginReady(t, k, uiPlugin)
	waitCoreAvailable(t, k)
	env := newIsolation(t, k)
	checkOwnPods(t, k, map[string]plugins.ImageRef{isoXTB: xtb.cluster, isoEgress: fx.cluster})
	checkXTBObjects(t, k, env)
	checkProbes(t, k, xtbNoEgress(env))
	checkProbes(t, k, outsideGrant(env))
	checkProbes(t, k, bypasses(env))
	checkProbes(t, k, exemptionProbes(env))
	checkProxyLog(t, k, env)
	checkProbes(t, k, connectorTokens(env))
	checkProbes(t, k, corePortTokens(env))
	removeInstance(t, k, isoXTB)
	removeInstance(t, k, isoEgress)
	removeInstance(t, k, isoPrivate)
}

// installXTB applies the PluginGrant and Plugin of XTB connector instance
// isoXTB with image: the manifest is the connector's own, read from the
// checkout, and the grant approves its capability and institution and no
// egress host, as the CLI writes it for a manifest that requests none.
func installXTB(t *testing.T, k *kube, image plugins.ImageRef) {
	t.Helper()
	data, err := os.ReadFile(xtbManifest)
	if err != nil {
		t.Fatalf("reading the manifest: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decoding the manifest: %v", err)
	}
	grant := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "openfief.io/v1alpha1", "kind": "PluginGrant",
		"metadata": map[string]any{"name": isoXTB, "namespace": namespace},
		"spec": map[string]any{
			"capabilities": []any{"connector"},
			"institutions": []any{xtbInstitution},
		},
	}}
	plugin := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "openfief.io/v1alpha1", "kind": "Plugin",
		"metadata": map[string]any{"name": isoXTB, "namespace": namespace},
		"spec": map[string]any{
			"image":    image.String(),
			"manifest": manifest,
			"config":   map[string]any{},
		},
	}}
	for _, obj := range []*unstructured.Unstructured{grant, plugin} {
		if err := k.applyOne(t.Context(), obj); err != nil {
			t.Fatal(err)
		}
	}
}

// probeAt is a container to run curl in.
type probeAt struct {
	ns, pod, container string
}

// isolation is what TestPluginIsolation probes from and to: the probe
// containers, the addresses (plugins have no DNS, so ClusterIPs), and the
// tokens it requested from the API server, which it never logs.
type isolation struct {
	xtb, egress, private, core, outsider, client probeAt
	// proxy is the egress proxy's URL; coreBrowser and corePlugin the
	// core's two ports; connector the XTB instance's Service and
	// otherPlugin the granted instance's.
	proxy, coreBrowser, corePlugin, connector, otherPlugin string
	// apiServer is the Kubernetes API server's own address, outside the
	// pod network: a destination a pod without a policy reaches.
	apiServer string
	// coreSelector and proxySelector are the core's and the egress proxy's
	// pod selectors, as their Services give them; corePort and proxyPort
	// the pod ports plugins dial.
	coreSelector, proxySelector map[string]string
	corePort, proxyPort         int32
	tokens                      isoTokens
}

// isoTokens are ServiceAccount tokens the test requests through
// TokenRequest, as the kubelet would project them.
type isoTokens struct {
	// egress is isoEgress's token for the egress proxy, private isoPrivate's.
	egress, private string
	// coreForXTB and coreForEgress are the core's tokens for each instance
	// (D-59).
	coreForXTB, coreForEgress string
	// xtbForXTB is the XTB plugin's own token with its instance's audience:
	// not the core's. xtbForCore is its token for the core, a connector's.
	xtbForXTB, xtbForCore string
	// uiForCore is the built-in UI's token for the core, a view plugin's;
	// uiForEgress its token for the egress proxy.
	uiForCore, uiForEgress string
}

// newIsolation starts the probe containers and reads the addresses and
// tokens TestPluginIsolation needs.
func newIsolation(t *testing.T, k *kube) *isolation {
	t.Helper()
	ctx := t.Context()
	image := clientImage(t, k)
	env := &isolation{
		outsider: probeAt{ns: clientNamespace, pod: outsiderPod, container: clientContainer},
		client:   probeAt{ns: clientNamespace, pod: clientPod, container: clientContainer},
	}
	for _, p := range []struct {
		to       *probeAt
		instance string
	}{{&env.xtb, isoXTB}, {&env.egress, isoEgress}, {&env.private, isoPrivate}} {
		pod, err := k.runningPluginPod(ctx, p.instance)
		if err != nil {
			t.Fatal(err)
		}
		*p.to = startProbe(t, k, pod.Name, image)
	}
	core, err := k.runningPod(ctx, namespace, coreDeployment)
	if err != nil {
		t.Fatal(err)
	}
	env.core = startProbe(t, k, core, image)
	env.readAddresses(t, k)
	env.tokens = requestTokens(t, k)
	return env
}

// readAddresses reads the Services' ClusterIPs, ports and selectors and the
// API server's own address.
func (env *isolation) readAddresses(t *testing.T, k *kube) {
	t.Helper()
	core := k.getService(t, coreService)
	proxy := k.getService(t, proxyService)
	connector := k.getService(t, plugins.ServiceAccountName(isoXTB))
	other := k.getService(t, plugins.ServiceAccountName(isoEgress))
	env.coreBrowser = "http://" + serviceAddress(t, core, "http")
	env.corePlugin = "http://" + serviceAddress(t, core, "connect")
	env.proxy = "http://" + serviceAddress(t, proxy, "proxy")
	env.connector = "http://" + serviceAddress(t, connector, operator.PortName)
	env.otherPlugin = "http://" + serviceAddress(t, other, operator.PortName)
	env.coreSelector, env.proxySelector = core.Spec.Selector, proxy.Spec.Selector
	env.corePort, env.proxyPort = podTargetPort(t, core, "connect"), podTargetPort(t, proxy, "proxy")
	addr, err := k.apiServerAddress(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	env.apiServer = "https://" + addr + "/"
}

// requestTokens requests every token TestPluginIsolation sends.
func requestTokens(t *testing.T, k *kube) isoTokens {
	t.Helper()
	d, err := k.clients.AppsV1().Deployments(namespace).Get(t.Context(), coreDeployment, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting the core Deployment: %v", err)
	}
	coreSA := d.Spec.Template.Spec.ServiceAccountName
	xtbSA, uiSA := plugins.ServiceAccountName(isoXTB), plugins.ServiceAccountName(uiPlugin)
	return isoTokens{
		egress:        k.token(t, plugins.ServiceAccountName(isoEgress), plugins.EgressAudience),
		private:       k.token(t, plugins.ServiceAccountName(isoPrivate), plugins.EgressAudience),
		coreForXTB:    k.token(t, coreSA, openfief.InstanceAudience(isoXTB)),
		coreForEgress: k.token(t, coreSA, openfief.InstanceAudience(isoEgress)),
		xtbForXTB:     k.token(t, xtbSA, openfief.InstanceAudience(isoXTB)),
		xtbForCore:    k.token(t, xtbSA, plugins.CoreAudience),
		uiForCore:     k.token(t, uiSA, plugins.CoreAudience),
		uiForEgress:   k.token(t, uiSA, plugins.EgressAudience),
	}
}

// token requests a token of ServiceAccount sa in OpenFief's namespace with
// audience, as the kubelet projects one.
func (k *kube) token(t *testing.T, sa, audience string) string {
	t.Helper()
	seconds := int64(tokenSeconds)
	req := &authenticationv1.TokenRequest{Spec: authenticationv1.TokenRequestSpec{
		Audiences: []string{audience}, ExpirationSeconds: &seconds,
	}}
	resp, err := k.clients.CoreV1().ServiceAccounts(namespace).CreateToken(t.Context(), sa, req, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("requesting a token of %s for %s: %v", sa, audience, err)
	}
	return resp.Status.Token
}

// clientImage returns the image of the suite's client pods, pinned by
// digest in testdata/clients.yaml, which the probes run too.
func clientImage(t *testing.T, k *kube) string {
	t.Helper()
	p, err := k.clients.CoreV1().Pods(clientNamespace).Get(t.Context(), clientPod, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting client pod %s: %v", clientPod, err)
	}
	return p.Spec.Containers[0].Image
}

// getService returns Service name of OpenFief's namespace.
func (k *kube) getService(t *testing.T, name string) *corev1.Service {
	t.Helper()
	s, err := k.clients.CoreV1().Services(namespace).Get(t.Context(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting Service %s: %v", name, err)
	}
	return s
}

// namedPort returns the port of s named name.
func namedPort(t *testing.T, s *corev1.Service, name string) corev1.ServicePort {
	t.Helper()
	i := slices.IndexFunc(s.Spec.Ports, func(p corev1.ServicePort) bool { return p.Name == name })
	if i < 0 {
		t.Fatalf("Service %s: %v: %s", s.Name, errNoPort, name)
	}
	return s.Spec.Ports[i]
}

// serviceAddress returns "ClusterIP:port" of s's port name.
func serviceAddress(t *testing.T, s *corev1.Service, name string) string {
	t.Helper()
	return net.JoinHostPort(s.Spec.ClusterIP, strconv.Itoa(int(namedPort(t, s, name).Port)))
}

// podTargetPort returns the pod port behind s's port name.
func podTargetPort(t *testing.T, s *corev1.Service, name string) int32 {
	t.Helper()
	return namedPort(t, s, name).TargetPort.IntVal
}

// apiServerAddress returns "address:port" of the Kubernetes API server
// itself, from the EndpointSlices of Service kubernetes, which a pod's
// packets reach after the Service's DNAT.
func (k *kube) apiServerAddress(ctx context.Context) (string, error) {
	eps, err := k.clients.DiscoveryV1().EndpointSlices(metav1.NamespaceDefault).List(ctx,
		metav1.ListOptions{LabelSelector: "kubernetes.io/service-name=kubernetes"})
	if err != nil {
		return "", fmt.Errorf("listing the API server's endpoints: %w", err)
	}
	for i := range eps.Items {
		s := &eps.Items[i]
		if len(s.Ports) == 0 || s.Ports[0].Port == nil || len(s.Endpoints) == 0 || len(s.Endpoints[0].Addresses) == 0 {
			continue
		}
		return net.JoinHostPort(s.Endpoints[0].Addresses[0], strconv.Itoa(int(*s.Ports[0].Port))), nil
	}
	return "", fmt.Errorf("%w: Service default/kubernetes", errNoEndpoint)
}

// startProbe adds an ephemeral container running image, with the
// restricted security context OpenFief's namespace enforces, to pod of
// OpenFief's namespace, and waits until it runs. It shares the pod's
// network, so whatever it reaches, the pod reaches.
func startProbe(t *testing.T, k *kube, pod, image string) probeAt {
	t.Helper()
	ctx := t.Context()
	name := probePrefix + strconv.FormatInt(time.Now().UnixNano(), 36)
	pods := k.clients.CoreV1().Pods(namespace)
	p, err := pods.Get(ctx, pod, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting pod %s: %v", pod, err)
	}
	p.Spec.EphemeralContainers = append(p.Spec.EphemeralContainers, corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name: name, Image: image, Command: []string{"sleep", probeSeconds},
			SecurityContext: probeSecurityContext(),
		},
	})
	if _, err = pods.UpdateEphemeralContainers(ctx, pod, p, metav1.UpdateOptions{FieldManager: fieldManager}); err != nil {
		t.Fatalf("adding a probe container to pod %s: %v", pod, err)
	}
	err = waitFor(ctx, probeTimeout, func(ctx context.Context) (string, bool, error) {
		current, getErr := pods.Get(ctx, pod, metav1.GetOptions{})
		if getErr != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, getErr)
		}
		return ephemeralRunning(current, name)(ctx)
	})
	if err != nil {
		t.Fatalf("waiting for the probe container in pod %s: %v", pod, err)
	}
	return probeAt{ns: namespace, pod: pod, container: name}
}

// probeSecurityContext is the restricted Pod Security Standard's container
// security context.
func probeSecurityContext() *corev1.SecurityContext {
	yes, no, user := true, false, int64(probeUser)
	return &corev1.SecurityContext{
		RunAsNonRoot: &yes, RunAsUser: &user, RunAsGroup: &user,
		AllowPrivilegeEscalation: &no, ReadOnlyRootFilesystem: &yes,
		Capabilities:   &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
		SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
	}
}

// ephemeralRunning checks that ephemeral container name of p runs, and
// fails when it is stuck or stopped.
func ephemeralRunning(p *corev1.Pod, name string) check {
	return func(context.Context) (string, bool, error) {
		for i := range p.Status.EphemeralContainerStatuses {
			s := &p.Status.EphemeralContainerStatuses[i]
			switch {
			case s.Name != name:
			case s.State.Running != nil:
				return "running", true, nil
			case s.State.Terminated != nil:
				return "", false, fmt.Errorf("%w: %s: %s", errProbeStopped, name, s.State.Terminated.Reason)
			case s.State.Waiting != nil && slices.Contains(stuckReasons(), s.State.Waiting.Reason):
				return "", false, fmt.Errorf("%w: %s: %s: %s", errStuck, name, s.State.Waiting.Reason, s.State.Waiting.Message)
			case s.State.Waiting != nil:
				return s.State.Waiting.Reason, false, nil
			}
		}
		return "no status yet", false, nil
	}
}

// checkOwnPods checks that each instance runs in a pod of its own, with
// the plugin's container alone, running its image (D-200).
func checkOwnPods(t *testing.T, k *kube, images map[string]plugins.ImageRef) {
	t.Helper()
	seen := map[string]string{}
	for instance, image := range images {
		pods, err := k.clients.CoreV1().Pods(namespace).List(t.Context(),
			metav1.ListOptions{LabelSelector: operator.LabelPlugin + "=" + instance})
		if err != nil {
			t.Fatalf("listing plugin %s's pods: %v", instance, err)
		}
		if len(pods.Items) != 1 {
			t.Errorf("plugin %s: %d pods; want 1", instance, len(pods.Items))
			continue
		}
		p := &pods.Items[0]
		if other, ok := seen[p.Name]; ok {
			t.Errorf("plugins %s and %s share pod %s", instance, other, p.Name)
		}
		seen[p.Name] = instance
		got := make([]string, 0, len(p.Spec.Containers))
		for i := range p.Spec.Containers {
			got = append(got, p.Spec.Containers[i].Name+" "+p.Spec.Containers[i].Image)
		}
		if diff := cmp.Diff([]string{operator.ContainerName + " " + image.String()}, got); diff != "" {
			t.Errorf("plugin %s: pod %s's containers (-want +got):\n%s", instance, p.Name, diff)
		}
	}
}

// pluginNetwork is what checkXTBObjects compares of what the operator made
// for a plugin.
type pluginNetwork struct {
	// Egress describes the NetworkPolicy's egress rules, one line per peer
	// and its ports; EgressPolicy is whether the policy restricts egress.
	Egress       []string
	EgressPolicy bool
	// Env are the proxy and egress token variables of the plugin's container.
	Env []string
	// TokenAudiences are the audiences of the pod's projected tokens.
	TokenAudiences []string
	// Listed is whether the plugin has an allow-list entry, and AllowList
	// the entry: the operator lists a plugin without egress hosts with an
	// empty one, so the proxy knows it has none. Effective are the egress
	// hosts its status says are effective.
	Listed    bool
	AllowList string
	Effective []string
}

// checkXTBObjects checks what the operator made for the XTB instance next
// to the granted one: the XTB's NetworkPolicy lets it reach the core's
// plugin port alone, and it has no proxy settings, no egress token, no
// host on its allow-list and no effective egress host, while the granted
// instance has each of them (D-32, D-63, D-200).
func checkXTBObjects(t *testing.T, k *kube, env *isolation) {
	t.Helper()
	core := fmt.Sprintf("pods %s: TCP/%d", selectorString(env.coreSelector), env.corePort)
	proxy := fmt.Sprintf("pods %s: TCP/%d", selectorString(env.proxySelector), env.proxyPort)
	want := map[string]*pluginNetwork{
		isoXTB: {
			Egress: []string{core}, EgressPolicy: true,
			TokenAudiences: []string{plugins.CoreAudience}, Listed: true,
		},
		isoEgress: {
			Egress: []string{core, proxy}, EgressPolicy: true,
			Env:            []string{plugins.EnvHTTPSProxy, plugins.EnvNoProxy, operator.EnvEgressIdentityFile},
			TokenAudiences: []string{plugins.CoreAudience, plugins.EgressAudience},
			Listed:         true, AllowList: fxAllowList, Effective: []string{fxEgress + ":443"},
		},
	}
	for instance, w := range want {
		if diff := cmp.Diff(w, networkOf(t, k, instance), cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("plugin %s's network (-want +got):\n%s", instance, diff)
		}
	}
}

// networkOf reads a pluginNetwork of instance from the cluster.
func networkOf(t *testing.T, k *kube, instance string) *pluginNetwork {
	t.Helper()
	ctx := t.Context()
	name := plugins.ServiceAccountName(instance)
	np, err := k.clients.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting plugin %s's NetworkPolicy: %v", instance, err)
	}
	pod, err := k.runningPluginPod(ctx, instance)
	if err != nil {
		t.Fatal(err)
	}
	cm, err := k.clients.CoreV1().ConfigMaps(namespace).Get(ctx, plugins.EgressAllowListConfigMap, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting the allow-list: %v", err)
	}
	p, err := k.getPlugin(ctx, instance)
	if err != nil {
		t.Fatal(err)
	}
	n := &pluginNetwork{
		Egress:         describeEgress(np.Spec.Egress),
		EgressPolicy:   slices.Contains(np.Spec.PolicyTypes, networkingv1.PolicyTypeEgress),
		Env:            egressEnv(pod),
		TokenAudiences: workloadOf(pod).TokenAudiences,
	}
	n.AllowList, n.Listed = cm.Data[instance]
	if p.Status.Effective != nil {
		n.Effective = p.Status.Effective.Egress
	}
	return n
}

// egressEnv lists the plugin container's variables that route or prove
// egress, sorted.
func egressEnv(pod *corev1.Pod) []string {
	names := []string{operator.EnvEgressIdentityFile, plugins.EnvHTTPSProxy, plugins.EnvNoProxy}
	var got []string
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Name != operator.ContainerName {
			continue
		}
		for _, e := range pod.Spec.Containers[i].Env {
			if slices.Contains(names, e.Name) {
				got = append(got, e.Name)
			}
		}
	}
	sort.Strings(got)
	return got
}

// describeEgress describes NetworkPolicy egress rules, one line per peer
// with the rule's ports.
func describeEgress(rules []networkingv1.NetworkPolicyEgressRule) []string {
	var lines []string
	for _, r := range rules {
		ports := make([]string, 0, len(r.Ports))
		for _, p := range r.Ports {
			proto := corev1.ProtocolTCP
			if p.Protocol != nil {
				proto = *p.Protocol
			}
			ports = append(ports, fmt.Sprintf("%s/%s", proto, p.Port.String()))
		}
		peers := r.To
		if len(peers) == 0 {
			peers = []networkingv1.NetworkPolicyPeer{{}}
		}
		for _, peer := range peers {
			lines = append(lines, describePeer(peer)+": "+strings.Join(ports, ","))
		}
	}
	return lines
}

// describePeer describes a NetworkPolicy peer; an empty one is anywhere.
func describePeer(p networkingv1.NetworkPolicyPeer) string {
	var parts []string
	if p.NamespaceSelector != nil {
		parts = append(parts, "namespaces "+selectorString(p.NamespaceSelector.MatchLabels))
	}
	if p.PodSelector != nil {
		parts = append(parts, "pods "+selectorString(p.PodSelector.MatchLabels))
	}
	if p.IPBlock != nil {
		parts = append(parts, "ipBlock "+p.IPBlock.CIDR)
	}
	if len(parts) == 0 {
		return "anywhere"
	}
	return strings.Join(parts, " ")
}

// selectorString writes labels as sorted key=value pairs.
func selectorString(labels map[string]string) string {
	pairs := make([]string, 0, len(labels))
	for k, v := range labels {
		pairs = append(pairs, k+"="+v)
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}

// probeResult is what curl reported of one call.
type probeResult struct {
	// status is the HTTP status, connect the proxy's answer to CONNECT;
	// 0 when there was none.
	status, connect int
	// exit is curl's exit code and message its error.
	exit    int
	message string
}

// String describes r for a failure message.
func (r probeResult) String() string {
	return fmt.Sprintf("HTTP %d, CONNECT %d, curl exit %d %q", r.status, r.connect, r.exit, r.message)
}

// parseProbe reads the write-out line curl ends its output with.
func parseProbe(out []byte) (probeResult, error) {
	s := strings.TrimSpace(string(out))
	line := s[strings.LastIndexByte(s, '\n')+1:]
	const minFields, allFields = 3, 4
	fields := strings.SplitN(line, " ", allFields)
	if len(fields) < minFields {
		return probeResult{}, errNoProbeOutput
	}
	var r probeResult
	var err error
	for i, n := range []*int{&r.status, &r.connect, &r.exit} {
		if *n, err = strconv.Atoi(fields[i]); err != nil {
			return probeResult{}, fmt.Errorf("%w: %w", errNoProbeOutput, err)
		}
	}
	if len(fields) == allFields {
		r.message = fields[3]
	}
	return r, nil
}

// outcome is what a probe must report.
type outcome struct {
	desc string
	ok   func(probeResult) bool
}

// noConnection is a probe that could not connect: no answer, and not for
// lack of a name.
func noConnection() outcome {
	return outcome{desc: "no connection", ok: func(r probeResult) bool {
		return r.status == 0 && r.connect == 0 && !dnsFailed(r) &&
			slices.Contains([]int{curlCouldntConnect, curlTimedOut, curlProxyError}, r.exit)
	}}
}

// lookupFails is a probe whose name lookup failed.
func lookupFails() outcome {
	return outcome{desc: "a failed name lookup", ok: func(r probeResult) bool { return r.status == 0 && dnsFailed(r) }}
}

// dnsFailed reports whether curl failed to look up the name.
func dnsFailed(r probeResult) bool {
	return r.exit == curlCouldntResolve || (r.exit == curlTimedOut && resolvingTimedOut.MatchString(r.message))
}

// answered is a probe that got an HTTP answer of any status.
func answered() outcome {
	return outcome{desc: "an HTTP answer", ok: func(r probeResult) bool { return r.status != 0 }}
}

// wantStatus is a probe answered with HTTP status code.
func wantStatus(code int) outcome {
	return outcome{desc: fmt.Sprintf("HTTP %d", code), ok: func(r probeResult) bool { return r.status == code }}
}

// connectRefused is a CONNECT the proxy refused, which Smokescreen answers
// with 407.
func connectRefused() outcome {
	return outcome{desc: "CONNECT refused with 407", ok: func(r probeResult) bool {
		return r.connect == http.StatusProxyAuthRequired
	}}
}

// probeCase is one curl call: from where, with which arguments and
// headers, and what it must report. Headers go in on stdin, so tokens are
// never in a command line or an error message.
type probeCase struct {
	name string
	from probeAt
	args []string
	// headers are sent to the server, proxyHeaders to the proxy.
	headers, proxyHeaders []string
	want                  outcome
}

// curlArgs returns the case's arguments and, when it has headers, the
// arguments that read them from stdin.
func (c *probeCase) curlArgs() []string {
	args := slices.Clone(c.args)
	switch {
	case len(c.proxyHeaders) > 0:
		args = append(args, "--proxy-header", "@-")
	case len(c.headers) > 0:
		args = append(args, "-H", "@-")
	}
	return args
}

// stdin returns the case's header lines, one per line, or nil without
// any.
func (c *probeCase) stdin() []byte {
	lines := c.headers
	if len(c.proxyHeaders) > 0 {
		lines = c.proxyHeaders
	}
	if len(lines) == 0 {
		return nil
	}
	var b bytes.Buffer
	for _, l := range lines {
		b.WriteString(l + "\n")
	}
	return b.Bytes()
}

// checkProbes runs each case and checks what it reports.
func checkProbes(t *testing.T, k *kube, cases []probeCase) {
	t.Helper()
	for i := range cases {
		c := &cases[i]
		if r := mustProbe(t, k, c); !c.want.ok(r) {
			t.Errorf("%s (from %s/%s): %s; want %s", c.name, c.from.ns, c.from.pod, r, c.want.desc)
		}
	}
}

// mustProbe runs c's curl call, retrying only when the exec stream itself
// fails, never on curl's answer.
func mustProbe(t *testing.T, k *kube, c *probeCase) probeResult {
	t.Helper()
	args, stdin := c.curlArgs(), c.stdin()
	var got probeResult
	err := waitFor(t.Context(), probeTimeout, func(ctx context.Context) (string, bool, error) {
		r, err := k.curl(ctx, c.from, stdin, args...)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		got = r
		return "", true, nil
	})
	if err != nil {
		t.Fatalf("%s: %v", c.name, err)
	}
	return got
}

// curl runs curl with args in at, with stdin, and returns what it reported.
// curl's own failure, such as no connection, is a result, not an error.
func (k *kube) curl(ctx context.Context, at probeAt, stdin []byte, args ...string) (probeResult, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	argv := append([]string{
		"curl", "-sS", "-o", "/dev/null", "--connect-timeout", connectTimeout, "--max-time", probeMaxTime,
		"-w", probeWriteOut,
	}, args...)
	out, err := k.exec(ctx, at.ns, at.pod, at.container, stdin, argv...)
	r, perr := parseProbe(out)
	if perr != nil {
		return probeResult{}, fmt.Errorf("curl in %s/%s: %w (%w)", at.ns, at.pod, perr, err)
	}
	return r, nil
}

// bearer is an Authorization header with token.
func bearer(token string) string { return "Authorization: Bearer " + token }

// proxyAuth is the Proxy-Authorization header an HTTP client sends for a
// proxy URL with token as its password (D-63).
func proxyAuth(token string) string {
	return "Proxy-Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte("plugin:"+token))
}

// xtbNoEgress are the XTB pod's probes: the core's plugin port is the one
// destination it reaches; the proxy, the core's browser port, another
// plugin, the API server's address and DNS are not, while the outsider
// reaches the API server and looks its name up (D-200 criterion 10).
func xtbNoEgress(env *isolation) []probeCase {
	return []probeCase{
		{name: "XTB pod to the core's plugin port", from: env.xtb, args: []string{env.corePlugin + "/v1/accounts"}, want: wantStatus(http.StatusUnauthorized)},
		{name: "XTB pod to the egress proxy", from: env.xtb, args: []string{"-p", "-x", env.proxy, "https://" + fxGrantOnly + "/"}, want: noConnection()},
		{name: "XTB pod to the core's browser port", from: env.xtb, args: []string{env.coreBrowser + "/healthz"}, want: noConnection()},
		{name: "XTB pod to another plugin", from: env.xtb, args: []string{env.otherPlugin + "/"}, want: noConnection()},
		{name: "XTB pod to the API server's address", from: env.xtb, args: []string{"-k", env.apiServer}, want: noConnection()},
		{name: "XTB pod looks up a cluster name", from: env.xtb, args: []string{"-k", apiServerName}, want: lookupFails()},
		{name: "XTB pod looks up an outside name", from: env.xtb, args: []string{"https://" + fxGrantOnly + "/"}, want: lookupFails()},
		{name: "outsider to the API server's address", from: env.outsider, args: []string{"-k", env.apiServer}, want: answered()},
		{name: "outsider looks up a cluster name", from: env.outsider, args: []string{"-k", apiServerName}, want: answered()},
	}
}

// outsideGrant are the granted plugin's CONNECTs that the proxy must
// refuse: a host its grant names but its manifest doesn't request, an IP
// address and an in-cluster service, a CONNECT without a token and one with
// the token of a plugin granted no host, and a plain HTTP request whose
// path, query, header and body hold the canary. That the
// proxy answers shows the pod reaches it (D-200 criteria 10 and 16).
func outsideGrant(env *isolation) []probeCase {
	auth := []string{proxyAuth(env.tokens.egress)}
	connect := func(name, url string, auth []string) probeCase {
		return probeCase{
			name: name, from: env.egress, args: []string{"-p", "-x", env.proxy, url},
			proxyHeaders: auth, want: connectRefused(),
		}
	}
	return []probeCase{
		connect("CONNECT to a host granted but not requested", "https://"+fxGrantOnly+"/", auth),
		connect("CONNECT to an IP address", "https://"+refusedIP+"/", auth),
		connect("CONNECT to an in-cluster service", coreURL+"/healthz", auth),
		connect("CONNECT without a token", "https://"+fxGrantOnly+"/", nil),
		connect("CONNECT with an ungranted plugin's token", "https://"+fxGrantOnly+"/", []string{proxyAuth(env.tokens.uiForEgress)}),
		{
			name: "plain HTTP request through the proxy", from: env.egress,
			args: []string{
				"-x", env.proxy, "-H", "X-Openfief-Canary: " + canary, "--data", canary,
				"http://" + fxGrantOnly + "/" + canary + "?" + canary + "=1",
			},
			proxyHeaders: auth, want: wantStatus(http.StatusProxyAuthRequired),
		},
	}
}

// Destinations the probes name.
const (
	// refusedIP is an address from TEST-NET-1 (RFC 5737), which nothing
	// serves.
	refusedIP = "192.0.2.1"
	// apiServerName is the API server's Service by its cluster DNS name.
	apiServerName = "https://kubernetes.default.svc.cluster.local/"
)

// bypasses are the granted plugin's attempts around the proxy: a direct
// connection to the API server's address and name lookups fail, while its
// pod reaches the proxy, shown by outsideGrant (D-200 criteria 10 and 16).
func bypasses(env *isolation) []probeCase {
	return []probeCase{
		{name: "granted plugin to the API server's address", from: env.egress, args: []string{"-k", env.apiServer}, want: noConnection()},
		{name: "granted plugin to an IP address", from: env.egress, args: []string{"-k", "https://" + refusedIP + "/"}, want: noConnection()},
		{name: "granted plugin looks up an outside name", from: env.egress, args: []string{"https://" + fxGrantOnly + "/"}, want: lookupFails()},
		{name: "granted plugin looks up a cluster name", from: env.egress, args: []string{"-k", apiServerName}, want: lookupFails()},
	}
}

// connectorTokens are calls to the XTB connector: from the core, its valid
// token for the instance passes the check, and the connector answers
// Unimplemented for a call it doesn't serve, while no token, a malformed
// one, the core's token for another instance and the plugin's own token
// are refused with 401; from any other pod, even with the valid token, no
// connection is made (D-59, D-200 criterion 17).
func connectorTokens(env *isolation) []probeCase {
	url := env.connector + connectorv1connect.ConnectorServiceImportBondLotsProcedure
	call := func(name string, from probeAt, token string, want outcome) probeCase {
		headers := []string{"Content-Type: application/json"}
		if token != "" {
			headers = append(headers, bearer(token))
		}
		return probeCase{name: name, from: from, args: []string{"--data", "{}", url}, headers: headers, want: want}
	}
	tk := env.tokens
	return []probeCase{
		call("core with its token for the instance", env.core, tk.coreForXTB, wantStatus(http.StatusNotImplemented)),
		call("core without a token", env.core, "", wantStatus(http.StatusUnauthorized)),
		call("core with a malformed token", env.core, "not-a-token", wantStatus(http.StatusUnauthorized)),
		call("core with its token for another instance", env.core, tk.coreForEgress, wantStatus(http.StatusUnauthorized)),
		call("core with the plugin's own token", env.core, tk.xtbForXTB, wantStatus(http.StatusUnauthorized)),
		call("client pod with the core's token", env.client, tk.coreForXTB, noConnection()),
		call("outsider with the core's token", env.outsider, tk.coreForXTB, noConnection()),
		call("another plugin with the core's token", env.egress, tk.coreForXTB, noConnection()),
	}
}

// corePortTokens are calls to the core from a connector's pod, on the
// plugin port, and from the client pod, on the browser port: the plugin
// port's read API and AuthService serve a view plugin's token and refuse no
// token, a malformed one, a connector's own and a view plugin's token for
// another audience; the browser port refuses a view plugin's token and a
// malformed one (D-69, D-83, D-115, D-200 criterion 17).
func corePortTokens(env *isolation) []probeCase {
	accounts := env.corePlugin + "/v1/accounts"
	login := env.corePlugin + identityv1connect.AuthServiceGetLoginMethodsProcedure
	read := func(name string, from probeAt, url, token string, want outcome) probeCase {
		var headers []string
		if token != "" {
			headers = []string{bearer(token)}
		}
		return probeCase{name: name, from: from, args: []string{url}, headers: headers, want: want}
	}
	auth := func(name, token string, want outcome) probeCase {
		headers := []string{"Content-Type: application/json"}
		if token != "" {
			headers = append(headers, bearer(token))
		}
		// The login page asks on an allowed origin, as the UI does: a call
		// without one is refused as forbidden whatever its token (D-60).
		body := `{"origin":"` + localCore + `"}`
		return probeCase{name: name, from: env.xtb, args: []string{"--data", body, login}, headers: headers, want: want}
	}
	tk := env.tokens
	return []probeCase{
		read("read API, view plugin's token", env.xtb, accounts, tk.uiForCore, wantStatus(http.StatusOK)),
		read("read API, no token", env.xtb, accounts, "", wantStatus(http.StatusUnauthorized)),
		read("read API, malformed token", env.xtb, accounts, "not-a-token", wantStatus(http.StatusUnauthorized)),
		read("read API, connector's own token", env.xtb, accounts, tk.xtbForCore, wantStatus(http.StatusUnauthorized)),
		read("read API, view plugin's egress token", env.xtb, accounts, tk.uiForEgress, wantStatus(http.StatusUnauthorized)),
		auth("AuthService, view plugin's token", tk.uiForCore, wantStatus(http.StatusOK)),
		auth("AuthService, no token", "", wantStatus(http.StatusForbidden)),
		auth("AuthService, malformed token", "not-a-token", wantStatus(http.StatusForbidden)),
		auth("AuthService, connector's own token", tk.xtbForCore, wantStatus(http.StatusForbidden)),
		auth("AuthService, view plugin's egress token", tk.uiForEgress, wantStatus(http.StatusForbidden)),
		read("browser port, malformed token", env.client, coreURL+"/v1/accounts", "not-a-token", wantStatus(http.StatusUnauthorized)),
		read("browser port, view plugin's token", env.client, coreURL+"/v1/accounts", tk.uiForCore, wantStatus(http.StatusUnauthorized)),
	}
}

// wantedLine is a decision line checkProxyLog waits for.
type wantedLine struct {
	role, proxyType, host string
}

// checkProxyLog waits until the egress proxy has logged a refusal for each
// request outsideGrant sent, by instance (no role for the CONNECT without
// a token), proxy type and host, then audits the whole log of the run as
// the demo and the release smoke test do (smokescreen.AuditLog): every
// decision line holds only who asked, where to and the decision, with a
// bare host:port, and every allowed request went to a host on its
// instance's allow-list; and no line holds the canary or a token (D-200
// criterion 16, GR-12).
func checkProxyLog(t *testing.T, k *kube, env *isolation) {
	t.Helper()
	wanted := []wantedLine{
		{isoEgress, "connect", fxGrantOnly + ":443"},
		{isoEgress, "connect", refusedIP + ":443"},
		{isoEgress, "connect", strings.TrimPrefix(coreURL, "http://")},
		{"", "connect", fxGrantOnly + ":443"},
		{uiPlugin, "connect", fxGrantOnly + ":443"},
		{isoEgress, "http", fxGrantOnly},
	}
	var log []byte
	err := waitFor(t.Context(), proxyLogTimeout, func(ctx context.Context) (string, bool, error) {
		var err error
		if log, err = pluginskube.EgressProxyLog(ctx, k.clients); err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		missing := missingLines(smokescreen.DecisionLines(log), wanted)
		lacking := exemptionLogProblems(log)
		return fmt.Sprintf("no refusal logged for %v; %v", missing, lacking), len(missing) == 0 && len(lacking) == 0, nil
	})
	if err != nil {
		t.Fatalf("waiting for the egress proxy's decision lines: %v", err)
	}
	audit := smokescreen.AuditLog(log, allowListHosts(t, k))
	for _, p := range audit.Problems {
		t.Errorf("the egress proxy's log: %s", p)
	}
	checkAllowedAudit(t, audit)
	tk := env.tokens
	for name, secret := range map[string]string{
		"the canary": canary, "the egress token": tk.egress, "its Proxy-Authorization header": proxyAuth(tk.egress),
		"the UI's egress token": tk.uiForEgress, "the second instance's egress token": tk.private,
	} {
		if bytes.Contains(log, []byte(secret)) {
			t.Errorf("the egress proxy's log holds %s (GR-12)", name)
		}
	}
}

// missingLines returns the wanted refusals that no line of lines records.
func missingLines(lines []smokescreen.DecisionLine, wanted []wantedLine) []wantedLine {
	var missing []wantedLine
	for _, w := range wanted {
		found := slices.ContainsFunc(lines, func(l smokescreen.DecisionLine) bool {
			return l["role"] == w.role && l["proxy_type"] == w.proxyType && l["requested_host"] == w.host && l["allow"] == false
		})
		if !found {
			missing = append(missing, w)
		}
	}
	return missing
}

// allowListHosts returns each instance's hosts on its current allow-list,
// the effective hosts the operator gives the proxy: stricter than its grant
// alone, which the demo's and the release smoke test's audits read.
func allowListHosts(t *testing.T, k *kube) map[string][]plugins.EgressHost {
	t.Helper()
	cm, err := k.clients.CoreV1().ConfigMaps(namespace).Get(t.Context(), plugins.EgressAllowListConfigMap, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting the allow-list: %v", err)
	}
	lists, err := plugins.DecodeAllowLists(cm.Data)
	if err != nil {
		t.Fatalf("decoding the allow-list: %v", err)
	}
	hosts := map[string][]plugins.EgressHost{}
	for _, l := range lists {
		hosts[l.Plugin] = append(hosts[l.Plugin], l.Hosts...)
	}
	return hosts
}
