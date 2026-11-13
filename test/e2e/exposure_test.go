//go:build e2e

package e2e

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/yaml"
)

// The HTTPRoute component (deploy/gateway, D-28) and the Gateway the
// exposure test attaches it to, as an owner's own.
const (
	// gatewayRouteFile and gatewayPolicyFile are the component's route and
	// the NetworkPolicy that admits the Gateway's data plane to the core.
	gatewayRouteFile  = "../../deploy/gateway/httproute.yaml"
	gatewayPolicyFile = "../../deploy/gateway/networkpolicy.yaml"
	// gatewayNamespace and gatewayName are the suite's Gateway; gatewayTLS
	// its certificate's Secret.
	gatewayNamespace = "openfief-e2e-gateway"
	gatewayName      = "openfief-e2e"
	gatewayTLS       = "openfief-e2e-tls"
	// gatewayHost is the owner's hostname in the test: .test never
	// resolves, so the test dials the Gateway itself and only names the
	// host, in TLS and HTTP.
	gatewayHost = "fief.gateway.e2e.test"
	// gatewayTimeout bounds the wait for the Gateway to serve the route.
	gatewayTimeout = 5 * time.Minute
	// certLifetime is how long the run's certificates are valid.
	certLifetime = 24 * time.Hour
	// serialBits is the size of a certificate's random serial number.
	serialBits = 128
	// envoyClass is the GatewayClass task e2e:gateway creates for Envoy
	// Gateway on kind; traefikClass the one k3s's Traefik brings.
	envoyClass   = "openfief-e2e"
	traefikClass = "traefik"
	// databaseCluster is the CloudNativePG cluster the bundle installs.
	databaseCluster = "openfief-db"
)

// gatewayGVR is the Gateway API's Gateway resource.
func gatewayGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "gateway.networking.k8s.io", Version: "v1", Resource: "gateways"}
}

// routeGVRs are the Gateway API's route resources, and its Gateway.
func routeGVRs() []schema.GroupVersionResource {
	return []schema.GroupVersionResource{
		gatewayGVR(),
		{Group: "gateway.networking.k8s.io", Version: "v1", Resource: "httproutes"},
		{Group: "gateway.networking.k8s.io", Version: "v1", Resource: "grpcroutes"},
	}
}

// exposingKinds are the kinds that expose something outside the cluster,
// which the bundle must not hold.
func exposingKinds() []string {
	return []string{"Ingress", "Gateway", "HTTPRoute", "GRPCRoute", "TLSRoute", "TCPRoute", "UDPRoute", "DaemonSet", "StatefulSet"}
}

// gatewayImpl is the Gateway API implementation a cluster runs (task
// e2e:gateway): its GatewayClass, the port its listener must use, where its
// data plane's Service is and which of its ports serves the listener, and
// the peer the component's NetworkPolicy admits.
type gatewayImpl struct {
	class        string
	listenerPort int64
	dataNS       string
	// dataService names the data plane's Service, or dataLabels selects it.
	dataService string
	dataLabels  string
	servicePort intstr.IntOrString
	peer        map[string]any
}

// gatewayOf returns the cluster's Gateway API implementation: Traefik's
// Gateway API provider on k3s, on its websecure entry point, and Envoy
// Gateway elsewhere (kind), whose proxy listens on 443.
func gatewayOf(t *testing.T, k *kube) gatewayImpl {
	t.Helper()
	isK3s, err := k.isK3s(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if isK3s {
		return gatewayImpl{
			class: traefikClass, listenerPort: 8443, dataNS: metav1.NamespaceSystem, dataService: "traefik",
			servicePort: intstr.FromString("websecure"),
			peer:        peer(metav1.NamespaceSystem, "app.kubernetes.io/name", "traefik"),
		}
	}
	const envoyNS = "envoy-gateway-system"
	return gatewayImpl{
		class: envoyClass, listenerPort: 443, dataNS: envoyNS,
		dataLabels: "gateway.envoyproxy.io/owning-gateway-name=" + gatewayName +
			",gateway.envoyproxy.io/owning-gateway-namespace=" + gatewayNamespace,
		servicePort: intstr.FromInt32(443),
		peer:        peer(envoyNS, "app.kubernetes.io/component", "proxy"),
	}
}

// peer is a NetworkPolicy peer: the pods labelled key=value in ns.
func peer(ns, key, value string) map[string]any {
	return map[string]any{
		"namespaceSelector": map[string]any{"matchLabels": map[string]any{"kubernetes.io/metadata.name": ns}},
		"podSelector":       map[string]any{"matchLabels": map[string]any{key: value}},
	}
}

// TestBundleExposure checks how OpenFief installed from the bundle can be
// reached (D-28). First, as installed: neither the bundle nor the openfief
// namespace holds an Ingress, a Gateway API Gateway or route, a DaemonSet
// or StatefulSet; every Service is ClusterIP without external IPs; no pod
// uses the host's network or ports; and every pod belongs to the bundle's
// Deployments, to a plugin the operator runs, or to the openfief-db
// cluster, so nothing runs a tunnel (SCEN-e2e-default-install-exposes-nothing).
// Then the HTTPRoute component's route and NetworkPolicy, with their
// placeholders replaced as docs/guides/expose.md does, attach the core to a
// Gateway of the cluster's implementation that terminates TLS for a test
// hostname with a certificate from a CA made for this run. Dialling the
// Gateway's data plane and naming the hostname, as curl --resolve does,
// the core answers /readyz, refuses the read API without a token, and
// passes the built-in UI's login page through, all over HTTPS verified
// against that CA (SCEN-e2e-httproute-https-through-gateway). Last, once
// every other bundle test has finished, since it restarts the core they
// use, it applies the component's patch of the core's auth origins and
// creates and uses a passkey on the route's hostname through the Gateway
// (checkRoutePasskey, SCEN-e2e-passkey-on-route-hostname). Then, signed
// in on localhost, the owner's browser creates a passkey for the route's
// hostname through the settings page's one-time link and uses it there
// (checkSettingsLinkPasskey, SCEN-e2e-passkey-second-host-settings-link).
func TestBundleExposure(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitCoreAvailable(t, k)
	checkBundleExposesNothing(t)
	checkNamespaceExposesNothing(t, k)

	impl := gatewayOf(t, k)
	certs := gatewayCerts(t)
	installGateway(t, k, &impl, certs.secret)
	applyRouteComponent(t, k, &impl)
	waitPluginReady(t, k, uiPlugin)
	gw := gatewayClient(t, k, &impl, certs.ca)
	waitHTTPS(t, gw.client, "/readyz", http.StatusOK, "")
	waitHTTPS(t, gw.client, "/v1/accounts", http.StatusUnauthorized, "")
	waitHTTPS(t, gw.client, "/login", http.StatusOK, loginButton)
	checkRoutePasskey(t, k, gw, &certs)
	checkSettingsLinkPasskey(t, k, gw, &certs)
}

// bundleObject is what the exposure test reads of an object in the bundle.
type bundleObject struct {
	Kind     string            `json:"kind"`
	Metadata metav1.ObjectMeta `json:"metadata"`
	Spec     struct {
		Type        corev1.ServiceType `json:"type"`
		ExternalIPs []string           `json:"externalIPs"`
	} `json:"spec"`
}

// checkBundleExposesNothing checks that no file of the bundle holds a kind
// that exposes anything, or a Service that isn't ClusterIP.
func checkBundleExposesNothing(t *testing.T) {
	t.Helper()
	b, err := readBundle(os.Getenv(bundleEnv))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range [][]byte{b.crds, b.install, b.database} {
		objs := bundleObjects(t, file)
		for i := range objs {
			obj := &objs[i]
			switch {
			case slices.Contains(exposingKinds(), obj.Kind):
				t.Errorf("the bundle holds %s %s", obj.Kind, obj.Metadata.Name)
			case obj.Kind == "Service" && !isClusterIP(obj.Spec.Type, obj.Spec.ExternalIPs):
				t.Errorf("the bundle's Service %s is %s with external IPs %v", obj.Metadata.Name, obj.Spec.Type, obj.Spec.ExternalIPs)
			}
		}
	}
}

// bundleObjects decodes every object in file, a YAML stream of the bundle.
func bundleObjects(t *testing.T, file []byte) []bundleObject {
	t.Helper()
	var objs []bundleObject
	decoder := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(file)))
	for {
		doc, err := decoder.Read()
		if errors.Is(err, io.EOF) {
			return objs
		}
		if err != nil {
			t.Fatalf("reading the bundle: %v", err)
		}
		var obj bundleObject
		if err := yaml.Unmarshal(doc, &obj); err != nil {
			t.Fatalf("decoding the bundle: %v", err)
		}
		objs = append(objs, obj)
	}
}

// isClusterIP reports whether a Service of type typ with externalIPs is
// reachable only inside the cluster; an empty type defaults to ClusterIP.
func isClusterIP(typ corev1.ServiceType, externalIPs []string) bool {
	return (typ == "" || typ == corev1.ServiceTypeClusterIP) && len(externalIPs) == 0
}

// checkNamespaceExposesNothing checks the openfief namespace as installed:
// no Ingress, Gateway or route, only ClusterIP Services, no pod on the
// host's network or ports, and no workload but the bundle's, the plugins'
// and the database's.
func checkNamespaceExposesNothing(t *testing.T, k *kube) {
	t.Helper()
	found := slices.Concat(k.routesIn(t), k.exposedServices(t), k.foreignWorkloads(t))
	if len(found) > 0 {
		t.Errorf("the default install exposes: %s", strings.Join(found, "; "))
	}
}

// routesIn names every Ingress and Gateway API Gateway or route in the
// openfief namespace.
func (k *kube) routesIn(t *testing.T) []string {
	t.Helper()
	ctx := t.Context()
	ingresses, err := k.clients.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing Ingresses: %v", err)
	}
	found := make([]string, 0, len(ingresses.Items))
	for i := range ingresses.Items {
		found = append(found, "Ingress "+ingresses.Items[i].Name)
	}
	for _, gvr := range routeGVRs() {
		list, listErr := k.dynamic.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
		switch {
		case apierrors.IsNotFound(listErr) || meta.IsNoMatchError(listErr):
			continue // the cluster has no such resource
		case listErr != nil:
			t.Fatalf("listing %s: %v", gvr.Resource, listErr)
		}
		for i := range list.Items {
			found = append(found, gvr.Resource+" "+list.Items[i].GetName())
		}
	}
	return found
}

// exposedServices names every Service in the openfief namespace that isn't
// ClusterIP without external IPs.
func (k *kube) exposedServices(t *testing.T) []string {
	t.Helper()
	services, err := k.clients.CoreV1().Services(namespace).List(t.Context(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing Services: %v", err)
	}
	var found []string
	for i := range services.Items {
		s := &services.Items[i]
		if !isClusterIP(s.Spec.Type, s.Spec.ExternalIPs) {
			found = append(found, fmt.Sprintf("Service %s of type %s with external IPs %v", s.Name, s.Spec.Type, s.Spec.ExternalIPs))
		}
	}
	return found
}

// foreignWorkloads names every DaemonSet and StatefulSet in the openfief
// namespace, and every pod there that runs on the host's network or ports,
// or that belongs to neither the bundle's Deployments and Jobs, a Deployment
// a Plugin owns, nor the openfief-db cluster.
func (k *kube) foreignWorkloads(t *testing.T) []string {
	t.Helper()
	ctx := t.Context()
	daemonSets, err := k.clients.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing DaemonSets: %v", err)
	}
	statefulSets, err := k.clients.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing StatefulSets: %v", err)
	}
	pods, err := k.clients.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing pods: %v", err)
	}
	found := make([]string, 0, len(daemonSets.Items)+len(statefulSets.Items))
	for i := range daemonSets.Items {
		found = append(found, "DaemonSet "+daemonSets.Items[i].Name)
	}
	for i := range statefulSets.Items {
		found = append(found, "StatefulSet "+statefulSets.Items[i].Name)
	}
	allowed := k.installDeployments(t)
	jobs := bundleJobs(t)
	for i := range pods.Items {
		p := &pods.Items[i]
		if p.Spec.HostNetwork || usesHostPort(p) {
			found = append(found, "pod "+p.Name+" on the host's network or ports")
		}
		if !k.ownedByInstall(t, p, allowed, jobs) {
			found = append(found, "pod "+p.Name+" of no workload the install runs")
		}
	}
	return found
}

// installDeployments returns the names of the Deployments the install
// runs: the bundle's and those a Plugin owns.
func (k *kube) installDeployments(t *testing.T) map[string]bool {
	t.Helper()
	allowed := map[string]bool{}
	bundled := bundleDeployments(t)
	for i := range bundled {
		allowed[bundled[i].Name] = true
	}
	deployments, err := k.clients.AppsV1().Deployments(namespace).List(t.Context(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing Deployments: %v", err)
	}
	for i := range deployments.Items {
		d := &deployments.Items[i]
		if owner := metav1.GetControllerOf(d); owner != nil && owner.Kind == "Plugin" {
			allowed[d.Name] = true
		}
	}
	return allowed
}

// bundleJobs returns the names of the Jobs the bundle's install runs: the one
// that generates SpiceDB's credentials (deploy/spicedb).
func bundleJobs(t *testing.T) map[string]bool {
	t.Helper()
	b, err := readBundle(os.Getenv(bundleEnv))
	if err != nil {
		t.Fatal(err)
	}
	jobs := map[string]bool{}
	objs := bundleObjects(t, b.install)
	for i := range objs {
		if objs[i].Kind == "Job" {
			jobs[objs[i].Metadata.Name] = true
		}
	}
	return jobs
}

// usesHostPort reports whether a container of p binds a port of its node.
func usesHostPort(p *corev1.Pod) bool {
	containers := slices.Concat(p.Spec.InitContainers, p.Spec.Containers)
	for i := range containers {
		for _, port := range containers[i].Ports {
			if port.HostPort != 0 {
				return true
			}
		}
	}
	return false
}

// ownedByInstall reports whether pod p belongs to an allowed Deployment,
// through its ReplicaSet, to one of the bundle's Jobs, or to the openfief-db
// cluster, directly or through one of its Jobs.
func (k *kube) ownedByInstall(t *testing.T, p *corev1.Pod, allowed, jobs map[string]bool) bool {
	t.Helper()
	owner := metav1.GetControllerOf(p)
	if owner == nil {
		return false
	}
	switch owner.Kind {
	case "Cluster":
		return owner.Name == databaseCluster
	case "Job":
		return jobs[owner.Name] || k.jobOfDatabase(t, owner.Name)
	case "ReplicaSet":
		return allowed[k.deploymentOf(t, owner.Name)]
	}
	return false
}

// jobOfDatabase reports whether Job name belongs to the openfief-db
// cluster; one already gone counts, as CloudNativePG removes its finished
// Jobs.
func (k *kube) jobOfDatabase(t *testing.T, name string) bool {
	t.Helper()
	job, err := k.clients.BatchV1().Jobs(namespace).Get(t.Context(), name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return true
	}
	if err != nil {
		t.Fatalf("getting Job %s: %v", name, err)
	}
	owner := metav1.GetControllerOf(job)
	return owner != nil && owner.Kind == "Cluster" && owner.Name == databaseCluster
}

// deploymentOf returns the Deployment that owns ReplicaSet name; empty for
// none.
func (k *kube) deploymentOf(t *testing.T, name string) string {
	t.Helper()
	rs, err := k.clients.AppsV1().ReplicaSets(namespace).Get(t.Context(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting ReplicaSet %s: %v", name, err)
	}
	if owner := metav1.GetControllerOf(rs); owner != nil && owner.Kind == "Deployment" {
		return owner.Name
	}
	return ""
}

// gatewayTLSFiles are the run's CA, which the test trusts, the Gateway's
// certificate for gatewayHost, followed by the CA's, with its key, as its
// TLS Secret holds them, and the keys of both in the form Chromium trusts
// them for the run (spkiHash).
type gatewayTLSFiles struct {
	ca     *x509.CertPool
	secret map[string][]byte
	keys   []string
}

// gatewayCerts makes a CA for this run and a certificate it signs for
// gatewayHost. The keys never leave this process but in the Gateway's
// Secret.
func gatewayCerts(t *testing.T) gatewayTLSFiles {
	t.Helper()
	now := time.Now()
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber: serial(t), Subject: pkix.Name{CommonName: "OpenFief e2e CA"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(certLifetime),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	leaf := &x509.Certificate{
		SerialNumber: serial(t), Subject: pkix.Name{CommonName: gatewayHost}, DNSNames: []string{gatewayHost},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(certLifetime),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, leaf, caCert, &key.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	leafCert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(caCert)
	chain := slices.Concat(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}))
	return gatewayTLSFiles{ca: pool, keys: []string{spkiHash(caCert), spkiHash(leafCert)}, secret: map[string][]byte{
		corev1.TLSCertKey:       chain,
		corev1.TLSPrivateKeyKey: pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}),
	}}
}

// serial is a random certificate serial number.
func serial(t *testing.T) *big.Int {
	t.Helper()
	n, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), serialBits))
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// installGateway creates the suite's Gateway in its own namespace, with
// cert as its TLS Secret: one HTTPS listener for gatewayHost on the port
// impl's data plane listens on, admitting routes from the openfief
// namespace alone. The namespace goes again when the test ends.
func installGateway(t *testing.T, k *kube, impl *gatewayImpl, cert map[string][]byte) {
	t.Helper()
	ctx := t.Context()
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: gatewayNamespace}}
	if _, err := k.clients.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); err != nil {
		t.Fatalf("creating namespace %s: %v", gatewayNamespace, err)
	}
	t.Cleanup(func() {
		err := k.clients.CoreV1().Namespaces().Delete(context.WithoutCancel(ctx), gatewayNamespace, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			t.Errorf("deleting namespace %s: %v", gatewayNamespace, err)
		}
	})
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: gatewayTLS, Namespace: gatewayNamespace},
		Type:       corev1.SecretTypeTLS, Data: cert,
	}
	if _, err := k.clients.CoreV1().Secrets(gatewayNamespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		t.Fatalf("creating the Gateway's TLS Secret: %v", err)
	}
	gw := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1", "kind": "Gateway",
		"metadata": map[string]any{"name": gatewayName, "namespace": gatewayNamespace},
		"spec": map[string]any{
			"gatewayClassName": impl.class,
			"listeners": []any{map[string]any{
				"name": "https", "protocol": "HTTPS", "port": impl.listenerPort, "hostname": gatewayHost,
				"tls": map[string]any{"mode": "Terminate", "certificateRefs": []any{map[string]any{"kind": "Secret", "name": gatewayTLS}}},
				"allowedRoutes": map[string]any{"namespaces": map[string]any{
					"from":     "Selector",
					"selector": map[string]any{"matchLabels": map[string]any{"kubernetes.io/metadata.name": namespace}},
				}},
			}},
		},
	}}
	if err := k.applyOne(ctx, gw); err != nil {
		t.Fatal(err)
	}
}

// applyRouteComponent applies the HTTPRoute component's route and
// NetworkPolicy, read from deploy/gateway, with their placeholders replaced
// as docs/guides/expose.md replaces them: the suite's Gateway and hostname,
// and impl's data plane as the peer the policy admits.
func applyRouteComponent(t *testing.T, k *kube, impl *gatewayImpl) {
	t.Helper()
	routeYAML, routeErr := os.ReadFile(gatewayRouteFile)
	policyYAML, policyErr := os.ReadFile(gatewayPolicyFile)
	if err := errors.Join(routeErr, policyErr); err != nil {
		t.Fatalf("reading the HTTPRoute component: %v", err)
	}
	route, policy := decodeObject(t, routeYAML), decodeObject(t, policyYAML)
	parent := onlyItem(t, route, "spec", "parentRefs")
	parent["name"], parent["namespace"] = gatewayName, gatewayNamespace
	rule := onlyItem(t, policy, "spec", "ingress")
	rule["from"] = []any{impl.peer}
	err := errors.Join(
		unstructured.SetNestedSlice(route.Object, []any{parent}, "spec", "parentRefs"),
		unstructured.SetNestedStringSlice(route.Object, []string{gatewayHost}, "spec", "hostnames"),
		unstructured.SetNestedSlice(policy.Object, []any{rule}, "spec", "ingress"),
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, obj := range []*unstructured.Unstructured{route, policy} {
		if err := k.applyOne(t.Context(), obj); err != nil {
			t.Fatal(err)
		}
	}
}

// decodeObject decodes the one object in data, a YAML document.
func decodeObject(t *testing.T, data []byte) *unstructured.Unstructured {
	t.Helper()
	var obj map[string]any
	if err := yaml.Unmarshal(data, &obj); err != nil {
		t.Fatalf("decoding the HTTPRoute component: %v", err)
	}
	return &unstructured.Unstructured{Object: obj}
}

// onlyItem returns the one item of obj's list at fields, which must be an
// object.
func onlyItem(t *testing.T, obj *unstructured.Unstructured, fields ...string) map[string]any {
	t.Helper()
	items, _, err := unstructured.NestedSlice(obj.Object, fields...)
	if err != nil || len(items) != 1 {
		t.Fatalf("%s %s has %s %v (%v); want one", obj.GetKind(), obj.GetName(), strings.Join(fields, "."), items, err)
	}
	item, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("%s %s's %s holds a %T", obj.GetKind(), obj.GetName(), strings.Join(fields, "."), items[0])
	}
	return item
}

// gatewayAccess is the way to the Gateway's data plane: the local address
// of a port-forward to one of its pods, and an HTTPS client that sends
// every request for gatewayHost there.
type gatewayAccess struct {
	client *http.Client
	addr   string
}

// gatewayClient returns the way to impl's data plane, through a
// port-forward to one of its pods, with a client that trusts only ca: what
// curl --resolve does, without DNS or the runner's hosts file.
func gatewayClient(t *testing.T, k *kube, impl *gatewayImpl, ca *x509.CertPool) *gatewayAccess {
	t.Helper()
	var target podPort
	err := waitFor(t.Context(), gatewayTimeout, func(ctx context.Context) (string, bool, error) {
		var findErr error
		target, findErr = k.dataPlane(ctx, impl)
		if findErr != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, findErr)
		}
		return target.pod, true, nil
	})
	if err != nil {
		t.Fatalf("waiting for the Gateway's data plane: %v", err)
	}
	addr := k.forwardPod(t, impl.dataNS, target)
	dialer := &net.Dialer{}
	return &gatewayAccess{addr: addr, client: &http.Client{
		Timeout: requestTimeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, addr)
			},
			TLSClientConfig: &tls.Config{RootCAs: ca, ServerName: gatewayHost, MinVersion: tls.VersionTLS12},
		},
		// The test checks each answer itself, the UI's redirects included.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// podPort is a pod and one of its container ports.
type podPort struct {
	pod  string
	port int32
}

// dataPlane finds impl's data plane: a ready pod behind its Service and the
// container port its HTTPS listener's Service port leads to.
func (k *kube) dataPlane(ctx context.Context, impl *gatewayImpl) (podPort, error) {
	svc, err := k.dataService(ctx, impl)
	if err != nil {
		return podPort{}, err
	}
	i := slices.IndexFunc(svc.Spec.Ports, func(p corev1.ServicePort) bool {
		if impl.servicePort.Type == intstr.String {
			return p.Name == impl.servicePort.StrVal
		}
		return p.Port == impl.servicePort.IntVal
	})
	if i < 0 {
		return podPort{}, fmt.Errorf("%w: Service %s/%s has no port %s", errNoPod, svc.Namespace, svc.Name, impl.servicePort.String())
	}
	target := svc.Spec.Ports[i].TargetPort
	selector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: svc.Spec.Selector})
	pods, err := k.clients.CoreV1().Pods(impl.dataNS).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return podPort{}, fmt.Errorf("listing the data plane's pods: %w", err)
	}
	for j := range pods.Items {
		p := &pods.Items[j]
		if p.DeletionTimestamp == nil && podReady(p) {
			if port, ok := containerPort(p, target); ok {
				return podPort{pod: p.Name, port: port}, nil
			}
		}
	}
	return podPort{}, fmt.Errorf("%w: data plane %s/%s", errNoPod, impl.dataNS, selector)
}

// dataService returns impl's data plane Service, by its name or labels.
func (k *kube) dataService(ctx context.Context, impl *gatewayImpl) (*corev1.Service, error) {
	services := k.clients.CoreV1().Services(impl.dataNS)
	if impl.dataService != "" {
		svc, err := services.Get(ctx, impl.dataService, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("getting the data plane's Service: %w", err)
		}
		return svc, nil
	}
	list, err := services.List(ctx, metav1.ListOptions{LabelSelector: impl.dataLabels})
	if err != nil {
		return nil, fmt.Errorf("listing the data plane's Services: %w", err)
	}
	if len(list.Items) != 1 {
		return nil, fmt.Errorf("%w: %d data plane Services %s", errNoPod, len(list.Items), impl.dataLabels)
	}
	return &list.Items[0], nil
}

// containerPort resolves a Service's target port on pod p: a number as it
// is, a name to the container port of that name.
func containerPort(p *corev1.Pod, target intstr.IntOrString) (int32, bool) {
	if target.Type == intstr.Int {
		return target.IntVal, true
	}
	for i := range p.Spec.Containers {
		for _, cp := range p.Spec.Containers[i].Ports {
			if cp.Name == target.StrVal {
				return cp.ContainerPort, true
			}
		}
	}
	return 0, false
}

// forwardPod forwards a free local port to target, a pod in ns and its port,
// until t ends, and returns its local address.
func (k *kube) forwardPod(t *testing.T, ns string, target podPort) string {
	t.Helper()
	pod := target.pod
	transport, upgrader, err := spdy.RoundTripperFor(k.config)
	if err != nil {
		t.Fatalf("port-forward: %v", err)
	}
	u := k.clients.CoreV1().RESTClient().Post().Resource("pods").Namespace(ns).Name(pod).SubResource("portforward").URL()
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, u)
	stop, ready := make(chan struct{}), make(chan struct{})
	fw, err := portforward.NewOnAddresses(dialer, []string{"127.0.0.1"}, []string{"0:" + strconv.Itoa(int(target.port))},
		stop, ready, io.Discard, io.Discard)
	if err != nil {
		t.Fatalf("port-forward to %s/%s: %v", ns, pod, err)
	}
	done := make(chan error, 1)
	go func() { done <- fw.ForwardPorts() }()
	t.Cleanup(func() {
		close(stop)
		if fwErr := <-done; fwErr != nil {
			t.Errorf("port-forward to %s/%s: %v", ns, pod, fwErr)
		}
	})
	select {
	case <-ready:
	case fwErr := <-done:
		t.Fatalf("port-forward to %s/%s: %v", ns, pod, fwErr)
	case <-time.After(forwardTimeout):
		t.Fatalf("port-forward to %s/%s: not listening after %s", ns, pod, forwardTimeout)
	}
	ports, err := fw.GetPorts()
	if err != nil || len(ports) != 1 {
		t.Fatalf("port-forward to %s/%s: ports %v (%v)", ns, pod, ports, err)
	}
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(int(ports[0].Local)))
}

// waitHTTPS sends GET https://gatewayHost/path through client until it
// answers status with a body holding want, until the Gateway serves the
// route or the deadline passes.
func waitHTTPS(t *testing.T, client *http.Client, path string, status int, want string) {
	t.Helper()
	url := "https://" + gatewayHost + path
	err := waitFor(t.Context(), gatewayTimeout, func(ctx context.Context) (string, bool, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return "", false, fmt.Errorf("GET %s: %w", url, err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return "", false, fmt.Errorf("%w: GET %s: %w", errRetry, url, err)
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, bundleBodyLimit))
		if closeErr := resp.Body.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return "", false, fmt.Errorf("%w: GET %s: reading the body: %w", errRetry, url, err)
		}
		state := fmt.Sprintf("status %d", resp.StatusCode)
		return state, resp.StatusCode == status && bytes.Contains(body, []byte(want)), nil
	})
	if err != nil {
		t.Fatalf("GET %s through the Gateway: %v; want status %d with %q", url, err, status, want)
	}
}
