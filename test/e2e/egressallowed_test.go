//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/egress"
	"github.com/arkadiuss/openfief/internal/plugins/smokescreen"
)

// The egress proxy's test-only exemption (D-215) and the fixture it exempts,
// which let D-200's criterion 16 be checked on an allowed connection: the
// proxy refuses private addresses, and a cluster's Services are private, so
// without one no connection through the proxy could succeed in the suite.
//
// The suite's install runs the proxy with -refused-range-exempt-host naming
// egressFixtureHost at its Service's ClusterIP, and the e2e overlay's
// NetworkPolicy patch (overlays/openfief) lets the proxy reach that fixture's
// pods on their port and nothing else. The ClusterIP is only known once the
// Service exists, so the suite creates the fixture before it applies the
// application layer (installEgressFixture) and adds the flag to the proxy's
// Deployment (editApp). The demo's base never sets it, and the release
// bundle's test fails if the bundle does (scripts/tests/release-bundle.test.sh).
const (
	egressFixtureNamespace = "openfief-e2e-egress"
	// egressFixtureService is the exempt host's Service and the fixture's
	// Deployment; egressOtherService is a second Service of the same pods,
	// in the same private range, which the exemption does not cover: only
	// the proxy's own check, not the NetworkPolicy, keeps the proxy from it.
	egressFixtureService = "egress-fixture"
	egressOtherService   = "egress-other"
	egressFixtureHost    = egressFixtureService + "." + egressFixtureNamespace + ".svc.cluster.local"
	egressOtherHost      = egressOtherService + "." + egressFixtureNamespace + ".svc.cluster.local"
	// isoPrivate is the connector instance granted egressOtherHost.
	isoPrivate = "egress-iso-private"
	// forbiddenText is in the proxy's log line of a connection it refused
	// for the range of the address a name resolved to.
	forbiddenText = "forbidden destination address"
)

// skipsEgressFixture reports the Ghostfolio shard, whose images don't hold
// the fixture's: it has no exemption.
func skipsEgressFixture() bool { return os.Getenv(shardEnv) == shardGhostfolio }

// installEgressFixture runs the fixture the exempt host names, in a
// namespace of its own, and returns the exemption for its Service's
// ClusterIP, or none in the Ghostfolio shard.
func installEgressFixture(ctx context.Context, k *kube, logf func(string, ...any)) (egress.Exemption, error) {
	if skipsEgressFixture() {
		return egress.Exemption{}, nil
	}
	logf("e2e: starting the egress proxy's exempt fixture")
	certs, err := fixtureCertificates(time.Now())
	if err != nil {
		return egress.Exemption{}, fmt.Errorf("the egress fixture's certificates: %w", err)
	}
	for _, obj := range egressFixtureObjects(certs.cert, certs.key) {
		if applyErr := k.applyOne(ctx, obj); applyErr != nil {
			return egress.Exemption{}, applyErr
		}
	}
	if err = waitFor(ctx, fixtureTimeout, k.deploymentAvailable(egressFixtureNamespace, egressFixtureService)); err != nil {
		return egress.Exemption{}, fmt.Errorf("waiting for the egress fixture: %w", err)
	}
	svc, err := k.clients.CoreV1().Services(egressFixtureNamespace).Get(ctx, egressFixtureService, metav1.GetOptions{})
	if err != nil {
		return egress.Exemption{}, fmt.Errorf("getting the egress fixture's Service: %w", err)
	}
	exempt, err := egress.ParseExemption(egressFixtureHost + "=" + svc.Spec.ClusterIP)
	if err != nil {
		return egress.Exemption{}, fmt.Errorf("the egress fixture's exemption: %w", err)
	}
	return exempt, nil
}

// egressFixtureObjects are the fixture's namespace, TLS Secret, Deployment
// and its two Services, egressFixtureService and egressOtherService. The
// fixture is the institution fixture's image, run as it is there: non-root,
// read-only, no privilege escalation and no capabilities. Its certificate
// is for another name; the probes don't verify it.
func egressFixtureObjects(cert, key []byte) []*unstructured.Unstructured {
	labels := map[string]any{"app.kubernetes.io/name": egressFixtureService}
	meta := func(name string) map[string]any {
		return map[string]any{"name": name, "namespace": egressFixtureNamespace, "labels": labels}
	}
	service := func(name string) *unstructured.Unstructured {
		return &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "v1", "kind": "Service", "metadata": meta(name),
			"spec": map[string]any{
				"selector": labels,
				"ports": []any{map[string]any{
					"name": "https", "port": int64(fixturePort), "targetPort": int64(fixturePodPort),
				}},
			},
		}}
	}
	container := map[string]any{
		"name": "fixture", "image": fixtureImage, "imagePullPolicy": "IfNotPresent",
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
	const certsName = "egress-fixture-tls"
	return []*unstructured.Unstructured{
		{Object: map[string]any{
			"apiVersion": "v1", "kind": "Namespace", "metadata": map[string]any{"name": egressFixtureNamespace},
		}},
		{Object: map[string]any{
			"apiVersion": "v1", "kind": "Secret", "metadata": meta(certsName), "type": "kubernetes.io/tls",
			"stringData": map[string]any{"tls.crt": string(cert), "tls.key": string(key)},
		}},
		{Object: map[string]any{
			"apiVersion": "apps/v1", "kind": "Deployment", "metadata": meta(egressFixtureService),
			"spec": map[string]any{
				"replicas": int64(1),
				"selector": map[string]any{"matchLabels": labels},
				"template": map[string]any{
					"metadata": map[string]any{"labels": labels},
					"spec": map[string]any{
						"automountServiceAccountToken": false,
						"containers":                   []any{container},
						"volumes": []any{map[string]any{
							"name": "tls", "secret": map[string]any{"secretName": certsName},
						}},
					},
				},
			},
		}},
		service(egressFixtureService),
		service(egressOtherService),
	}
}

// errProxyContainer reports an egress proxy Deployment that doesn't have
// the one container with arguments the suite adds the exemption to.
var errProxyContainer = errors.New("the egress proxy Deployment must have one container")

// exemptProxy adds the exemption to the egress proxy's arguments.
func exemptProxy(obj *unstructured.Unstructured, exempt egress.Exemption) error {
	containers, found, err := unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
	if err != nil {
		return fmt.Errorf("%w: %w", errProxyContainer, err)
	}
	if !found || len(containers) != 1 {
		return fmt.Errorf("%w: found %d", errProxyContainer, len(containers))
	}
	c, ok := containers[0].(map[string]any)
	if !ok {
		return fmt.Errorf("%w: it is a %T", errProxyContainer, containers[0])
	}
	args, ok := c["args"].([]any)
	if !ok {
		return fmt.Errorf("%w: its args are a %T", errProxyContainer, c["args"])
	}
	c["args"] = append(args, "-refused-range-exempt-host="+exempt.String())
	containers[0] = c
	if err := unstructured.SetNestedSlice(obj.Object, containers, "spec", "template", "spec", "containers"); err != nil {
		return fmt.Errorf("egress proxy: %w", err)
	}
	return nil
}

// installFXAt applies the PluginGrant and Plugin of an NBP FX connector
// instance name whose baseUrl is https://host, so that its manifest requests
// host through egressFromConfig and its grant approves it.
func installFXAt(t *testing.T, k *kube, name string, image plugins.ImageRef, host string) {
	t.Helper()
	grant, plugin := fxGrant(name), fxPluginObject(t, name, image.String())
	egressHosts, _, err := unstructured.NestedStringSlice(grant.Object, "spec", "egress")
	if err != nil {
		t.Fatal(err)
	}
	if err = unstructured.SetNestedStringSlice(grant.Object, append(egressHosts, host), "spec", "egress"); err != nil {
		t.Fatal(err)
	}
	if err = unstructured.SetNestedField(plugin.Object, "https://"+host, "spec", "config", "baseUrl"); err != nil {
		t.Fatal(err)
	}
	for _, obj := range []*unstructured.Unstructured{grant, plugin} {
		if err = k.applyOne(t.Context(), obj); err != nil {
			t.Fatal(err)
		}
	}
}

// connectAllowed is a CONNECT the proxy let through: 200.
func connectAllowed() outcome {
	return outcome{desc: "CONNECT allowed with 200", ok: func(r probeResult) bool { return r.connect == 200 }}
}

// connectBadGateway is a CONNECT the proxy refused for the address its host
// resolved to, which Smokescreen answers with 502.
func connectBadGateway() outcome {
	return outcome{desc: "CONNECT refused with 502", ok: func(r probeResult) bool { return r.connect == 502 }}
}

// exemptionProbes are the CONNECTs of criterion 16 on an allowed connection
// and its limit (D-215): the granted instance reaches the exempt host, and
// another instance, granted another host of the same private range, is
// refused. The target is a host name, so the proxy resolves it.
func exemptionProbes(env *isolation) []probeCase {
	connect := func(name string, from probeAt, token, host string, want outcome) probeCase {
		return probeCase{
			name: name, from: from, args: []string{"-k", "-p", "-x", env.proxy, "https://" + host + "/"},
			proxyHeaders: []string{proxyAuth(token)}, want: want,
		}
	}
	return []probeCase{
		connect("CONNECT to the exempt host, granted to the instance", env.egress, env.tokens.egress, egressFixtureHost, connectAllowed()),
		connect("CONNECT to a granted host in the same range, not exempt", env.private, env.tokens.private, egressOtherHost, connectBadGateway()),
	}
}

// exemptionLogProblems returns what the proxy's log lacks of D-215's
// scenario, for the lines of log: the start warning naming the exempt host;
// an allowed decision line of isoEgress for the exempt host that holds no
// error; and a line of isoPrivate for the other host whose error says its
// address was refused. Each is a line the proxy logs before it answers, so
// they can be waited for.
func exemptionLogProblems(log []byte) []string {
	var problems []string
	if !exemptionWarned(log) {
		problems = append(problems, "no start warning of the exemption naming "+egressFixtureHost)
	}
	lines := smokescreen.DecisionLines(log)
	allowed := slices.ContainsFunc(lines, func(l smokescreen.DecisionLine) bool {
		return l.Text("role") == isoEgress && l.Text("requested_host") == egressFixtureHost+":443" &&
			l["allow"] == true && l["error"] == nil && l.Text("proxy_type") == "connect"
	})
	if !allowed {
		problems = append(problems, "no allowed decision line without an error for "+isoEgress+" to "+egressFixtureHost)
	}
	refused := slices.ContainsFunc(lines, func(l smokescreen.DecisionLine) bool {
		return l.Text("role") == isoPrivate && l.Text("requested_host") == egressOtherHost+":443" &&
			strings.Contains(l.Text("error"), forbiddenText)
	})
	if !refused {
		problems = append(problems, "no decision line of "+isoPrivate+" to "+egressOtherHost+" refused for its address")
	}
	return problems
}

// exemptionWarned reports whether a line of log is the proxy's start warning
// of the exemption, naming the exempt host.
func exemptionWarned(log []byte) bool {
	for raw := range bytes.SplitSeq(log, []byte("\n")) {
		if bytes.Contains(raw, []byte(smokescreen.ExemptionWarning)) && bytes.Contains(raw, []byte(`"`+egressFixtureHost+`"`)) {
			return true
		}
	}
	return false
}

// checkAllowedAudit checks AuditLog's answer on the run's log after the
// exemption's CONNECTs: it found no problem, which includes every allowed
// line being on its instance's allow-list, and the allowed set holds the
// exempt host for isoEgress.
func checkAllowedAudit(t *testing.T, audit smokescreen.Audit) {
	t.Helper()
	if len(audit.Allowed) == 0 {
		t.Error("the egress proxy's log has no allowed destination, want the exempt host")
	}
	if !slices.ContainsFunc(audit.Allowed, func(d smokescreen.Destination) bool {
		return d.Instance == isoEgress && d.Host == egressFixtureHost+":443"
	}) {
		t.Errorf("the egress proxy's allowed destinations %+v don't hold %s to %s:443", audit.Allowed, isoEgress, egressFixtureHost)
	}
}
