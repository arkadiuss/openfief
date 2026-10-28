package operator_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
)

// Synthetic plugin image; the digest is not of any real image.
const sampleImage = "registry.example.com/openfief/sample-connector@sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"

// sha256 of `{"interval":"24h"}`.
const sampleConfigHash = "d9bdb961470e5cc7001bfce1e9a08fffa2d8ca1a24c035e1dbf8684e7387eaac"

// coreSelector selects the synthetic core's pods.
func coreSelector() map[string]string {
	return map[string]string{"app.kubernetes.io/name": "openfief-core"}
}

// proxySelector selects the synthetic egress proxy's pods.
func proxySelector() map[string]string {
	return map[string]string{"app.kubernetes.io/name": "openfief-egress-proxy"}
}

// cluster is a synthetic cluster with documentation-range ClusterIPs.
func cluster() operator.Cluster {
	return operator.Cluster{
		Topology: plugins.Topology{
			Core:        plugins.Service{Address: "192.0.2.10", Port: 80, TargetPort: 8080},
			EgressProxy: plugins.Service{Address: "192.0.2.20", Port: 4750, TargetPort: 4751},
			PluginPort:  9090,
		},
		CoreSelector:        coreSelector(),
		EgressProxySelector: proxySelector(),
	}
}

// workload compiles the synthetic sample plugin under a grant of egress.
func workload(t *testing.T, egress []plugins.EgressHost) plugins.Workload {
	t.Helper()
	p := &plugins.Plugin{
		Name:  "sample",
		Image: sampleImage,
		Manifest: plugins.Manifest{
			Capabilities: []plugins.Capability{plugins.CapabilityConnector},
			Egress:       []plugins.EgressHost{{Host: "api.example.com", Port: 443}},
			Secrets:      []plugins.SecretMount{{Name: "sample-token", MountPath: "/etc/openfief/token"}},
		},
		Config: []byte(`{"interval":"24h"}`),
	}
	g := &plugins.Grant{Plugin: "sample", Permissions: plugins.Permissions{
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Egress:       egress,
		Secrets:      []string{"sample-token"},
	}}
	e, err := plugins.Effective(p, g)
	if err != nil {
		t.Fatalf("Effective: %v", err)
	}
	c := cluster()
	w, err := plugins.Compile(e, c.Topology)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	return w
}

// ptrTo returns a pointer to a copy of v.
func ptrTo[T any](v T) *T { p := &v; return p }

func labels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name": "openfief-plugin", "app.kubernetes.io/managed-by": "openfief-operator",
		"openfief.io/plugin": "sample",
	}
}

func objectMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{Name: name, Namespace: "openfief", Labels: labels()}
}

func tcp(port int32) []networkingv1.NetworkPolicyPort {
	return []networkingv1.NetworkPolicyPort{{Protocol: ptrTo(corev1.ProtocolTCP), Port: ptrTo(intstr.FromInt32(port))}}
}

func peer(selector map[string]string) []networkingv1.NetworkPolicyPeer {
	return []networkingv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{MatchLabels: selector}}}
}

func networkPolicy(egress ...networkingv1.NetworkPolicyEgressRule) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		TypeMeta:   metav1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "NetworkPolicy"},
		ObjectMeta: objectMeta("plugin-sample"),
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"openfief.io/plugin": "sample"}},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
			Ingress:     []networkingv1.NetworkPolicyIngressRule{{From: peer(coreSelector()), Ports: tcp(9090)}},
			Egress:      egress,
		},
	}
}

func env(withProxy bool) []corev1.EnvVar {
	out := []corev1.EnvVar{}
	if withProxy {
		out = append(out, corev1.EnvVar{Name: "HTTPS_PROXY", Value: "http://192.0.2.20:4750"}, corev1.EnvVar{Name: "NO_PROXY", Value: "192.0.2.10"})
	}
	out = append(out,
		corev1.EnvVar{Name: "OPENFIEF_CONFIG", ValueFrom: &corev1.EnvVarSource{ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "plugin-sample-config"}, Key: "config.json",
		}}},
		corev1.EnvVar{Name: "OPENFIEF_CORE_ADDRESS", Value: "192.0.2.10:80"},
	)
	if withProxy {
		out = append(out, corev1.EnvVar{Name: "OPENFIEF_EGRESS_TOKEN_FILE", Value: "/var/run/secrets/openfief-egress/token"})
	}
	return append(out,
		corev1.EnvVar{Name: "OPENFIEF_LISTEN_PORT", Value: "9090"},
		corev1.EnvVar{Name: "OPENFIEF_TOKEN_FILE", Value: "/var/run/secrets/openfief/token"},
	)
}

// tokenVolume is a projected token volume for audience.
func tokenVolume(name, audience string) corev1.Volume {
	return corev1.Volume{Name: name, VolumeSource: corev1.VolumeSource{Projected: &corev1.ProjectedVolumeSource{
		Sources: []corev1.VolumeProjection{{ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
			Audience: audience, ExpirationSeconds: ptrTo[int64](3600), Path: "token",
		}}},
	}}}
}

// podVolumes are the pod's volumes and the container's mounts: the core
// token, the egress token only with egress granted, and the one Secret.
func podVolumes(withProxy bool) ([]corev1.Volume, []corev1.VolumeMount) {
	vols := []corev1.Volume{tokenVolume("openfief-token", "openfief-core")}
	mounts := []corev1.VolumeMount{{Name: "openfief-token", MountPath: "/var/run/secrets/openfief", ReadOnly: true}}
	if withProxy {
		vols = append(vols, tokenVolume("openfief-egress-token", "openfief-egress"))
		mounts = append(mounts, corev1.VolumeMount{Name: "openfief-egress-token", MountPath: "/var/run/secrets/openfief-egress", ReadOnly: true})
	}
	vols = append(vols, corev1.Volume{Name: "secret-0", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "sample-token"}}})
	mounts = append(mounts, corev1.VolumeMount{Name: "secret-0", MountPath: "/etc/openfief/token", ReadOnly: true})
	return vols, mounts
}

func deployment(withProxy bool) *appsv1.Deployment {
	vols, mounts := podVolumes(withProxy)
	return &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: objectMeta("plugin-sample"),
		Spec: appsv1.DeploymentSpec{
			Replicas: ptrTo[int32](1),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"openfief.io/plugin": "sample"}},
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels(), Annotations: map[string]string{"openfief.io/config-sha256": sampleConfigHash}},
				Spec: corev1.PodSpec{
					ServiceAccountName:           "plugin-sample",
					AutomountServiceAccountToken: ptrTo(false),
					EnableServiceLinks:           ptrTo(false),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: ptrTo(true), SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
					},
					Containers: []corev1.Container{{
						Name: "plugin", Image: sampleImage, ImagePullPolicy: corev1.PullIfNotPresent,
						Ports:        []corev1.ContainerPort{{Name: "connect", ContainerPort: 9090, Protocol: corev1.ProtocolTCP}},
						Env:          env(withProxy),
						VolumeMounts: mounts,
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: ptrTo(false), Privileged: ptrTo(false), ReadOnlyRootFilesystem: ptrTo(true),
							RunAsNonRoot: ptrTo(true), Capabilities: &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
						},
					}},
					Volumes: vols,
				},
			},
		},
	}
}

func objects(withProxy bool) operator.Objects {
	egress := []networkingv1.NetworkPolicyEgressRule{{To: peer(coreSelector()), Ports: tcp(8080)}}
	if withProxy {
		egress = append(egress, networkingv1.NetworkPolicyEgressRule{To: peer(proxySelector()), Ports: tcp(4751)})
	}
	return operator.Objects{
		ServiceAccount: &corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ServiceAccount"}, ObjectMeta: objectMeta("plugin-sample"),
			AutomountServiceAccountToken: ptrTo(false),
		},
		ConfigMap: &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"}, ObjectMeta: objectMeta("plugin-sample-config"),
			Data: map[string]string{"config.json": `{"interval":"24h"}`},
		},
		Service: &corev1.Service{
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Service"}, ObjectMeta: objectMeta("plugin-sample"),
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"openfief.io/plugin": "sample"},
				Ports:    []corev1.ServicePort{{Name: "connect", Protocol: corev1.ProtocolTCP, Port: 9090, TargetPort: intstr.FromInt32(9090)}},
			},
		},
		NetworkPolicy: networkPolicy(egress...),
		Deployment:    deployment(withProxy),
	}
}

func TestBuildObjects(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		egress []plugins.EgressHost
		want   operator.Objects
	}{
		{
			name:   "granted egress adds the proxy rule, proxy environment and egress token",
			egress: []plugins.EgressHost{{Host: "api.example.com", Port: 443}},
			want:   objects(true),
		},
		{
			name: "no granted egress reaches only the core",
			want: objects(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := workload(t, tt.egress)
			c := cluster()
			got, err := operator.BuildObjects(&w, &c)
			if err != nil {
				t.Fatalf("BuildObjects: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("BuildObjects mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBuildObjectsRefuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		plugin  string
		cluster func(*operator.Cluster)
		want    error
	}{
		{
			name: "name too long for a Service", plugin: strings.Repeat("a", 57),
			cluster: func(*operator.Cluster) {}, want: operator.ErrPluginNameTooLong,
		},
		{
			name: "core without a selector", plugin: "sample",
			cluster: func(c *operator.Cluster) { c.CoreSelector = nil }, want: operator.ErrServiceUnusable,
		},
		{
			name: "proxy selector that selects plugin pods", plugin: "sample",
			cluster: func(c *operator.Cluster) {
				c.EgressProxySelector = map[string]string{"app.kubernetes.io/managed-by": "openfief-operator"}
			},
			want: operator.ErrServiceUnusable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := workload(t, []plugins.EgressHost{{Host: "api.example.com", Port: 443}})
			w.Plugin = tt.plugin
			c := cluster()
			tt.cluster(&c)
			got, err := operator.BuildObjects(&w, &c)
			if diff := cmp.Diff(tt.want, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("BuildObjects error mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(operator.Objects{}, got); diff != "" {
				t.Errorf("BuildObjects returned objects with its error:\n%s", diff)
			}
		})
	}
}

func TestDenyAllNetworkPolicy(t *testing.T) {
	t.Parallel()
	want := networkPolicy()
	want.Spec.Ingress = nil
	if diff := cmp.Diff(want, operator.DenyAllNetworkPolicy("sample")); diff != "" {
		t.Errorf("DenyAllNetworkPolicy mismatch (-want +got):\n%s", diff)
	}
}

func TestAllowListConfigMap(t *testing.T) {
	t.Parallel()
	got, err := operator.AllowListConfigMap([]plugins.EgressAllowList{
		{Plugin: "sample", Hosts: []plugins.EgressHost{{Host: "data.example.org", Port: 8443}, {Host: "api.example.com", Port: 443}}},
		{Plugin: "xtb"},
	})
	if err != nil {
		t.Fatalf("AllowListConfigMap: %v", err)
	}
	want := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "openfief-egress-allowlist", Namespace: "openfief",
			Labels: map[string]string{"app.kubernetes.io/managed-by": "openfief-operator"},
		},
		Data: map[string]string{"sample": "api.example.com:443\ndata.example.org:8443\n", "xtb": ""},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("AllowListConfigMap mismatch (-want +got):\n%s", diff)
	}
}
