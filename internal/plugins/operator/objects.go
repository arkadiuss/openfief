package operator

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"net"
	"path"
	"slices"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// Labels on every object the operator creates for a plugin. Pods carry them
// too; LabelPlugin selects one plugin's pods.
const (
	LabelPlugin    = "openfief.io/plugin"
	labelName      = "app.kubernetes.io/name"
	labelManagedBy = "app.kubernetes.io/managed-by"
	pluginAppName  = "openfief-plugin"
	managerName    = "openfief-operator"
)

// The runtime contract between the operator and a plugin's container.
const (
	// ContainerName is the plugin container's name.
	ContainerName = "plugin"
	// PortName names the plugin's Connect RPC port.
	PortName = "connect"
	// ConfigKey is the key of the plugin's config in its config ConfigMap.
	ConfigKey = "config.json"
	// EnvConfig holds the plugin's config document; empty when it has none.
	EnvConfig = "OPENFIEF_CONFIG"
	// EnvIdentityFile is the path of the plugin's projected token (D-27).
	EnvIdentityFile = "OPENFIEF_TOKEN_FILE"
	// EnvEgressIdentityFile is the path of the projected token the plugin
	// presents to the egress proxy (D-63); set only when egress is granted.
	EnvEgressIdentityFile = "OPENFIEF_EGRESS_TOKEN_FILE"
	// EnvCoreAddress is the core's "ClusterIP:port"; plugins have no DNS.
	EnvCoreAddress = "OPENFIEF_CORE_ADDRESS"
	// EnvListenPort is the port the plugin serves Connect RPC on.
	EnvListenPort = "OPENFIEF_LISTEN_PORT"
	// AnnotationConfigHash on the pod template is the SHA-256 of the config,
	// so a config change rolls the pod.
	AnnotationConfigHash = "openfief.io/config-sha256"
)

const (
	kindConfigMap     = "ConfigMap"
	identityVolume    = "openfief-token"
	egressVolume      = "openfief-egress-token"
	signingKeysVolume = "openfief-sa-jwks"
	// operatorEnvVars counts the variables the operator may add to the Workload's.
	operatorEnvVars = 5
	// operatorVolumes is the most volumes a plugin pod gets besides its
	// Secrets: two projected tokens and the signing keys.
	operatorVolumes = 3
	// MaxPluginNameLen keeps "plugin-<name>" within the 63 characters of a
	// Service name.
	MaxPluginNameLen = 56
)

// namespaceNameLabel is the label Kubernetes sets on every namespace to its
// name, which a NetworkPolicy's namespace selector matches.
const namespaceNameLabel = "kubernetes.io/metadata.name"

// ErrPluginNameTooLong reports a plugin name longer than MaxPluginNameLen.
var ErrPluginNameTooLong = errors.New("plugin name too long")

// Objects are the Kubernetes objects of one running plugin. The Secrets it
// mounts are the owner's (D-29) and are never among them.
type Objects struct {
	ServiceAccount *corev1.ServiceAccount
	ConfigMap      *corev1.ConfigMap
	Service        *corev1.Service
	NetworkPolicy  *networkingv1.NetworkPolicy
	Deployment     *appsv1.Deployment
}

// BuildObjects maps a Workload onto Kubernetes objects:
//   - a ServiceAccount that mounts no API token, and a pod that mounts only
//     the projected token with the core's audience (D-27);
//   - the granted Secrets, mounted read-only;
//   - the ServiceAccount signing keys ConfigMap, mounted read-only, for
//     verifying the core's token (D-59);
//   - a NetworkPolicy denying all ingress and egress except the Workload's
//     rules, whose peers are the core's and egress proxy's pods, selected by
//     their Services' selectors, on pod ports (D-32);
//   - the config in a ConfigMap, passed as EnvConfig;
//   - a ClusterIP Service for the core to call the plugin.
func BuildObjects(w *plugins.Workload, c *Cluster) (Objects, error) {
	if len(w.Plugin) > MaxPluginNameLen {
		return Objects{}, fmt.Errorf("%w: %q has more than %d characters", ErrPluginNameTooLong, w.Plugin, MaxPluginNameLen)
	}
	np, err := networkPolicy(w, c)
	if err != nil {
		return Objects{}, err
	}
	name := plugins.ServiceAccountName(w.Plugin)
	automount := w.ServiceAccount.AutomountAPIToken
	return Objects{
		ServiceAccount: &corev1.ServiceAccount{
			TypeMeta:                     metav1.TypeMeta{APIVersion: "v1", Kind: "ServiceAccount"},
			ObjectMeta:                   objectMeta(name, w.Plugin),
			AutomountServiceAccountToken: &automount,
		},
		ConfigMap: &corev1.ConfigMap{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: kindConfigMap},
			ObjectMeta: objectMeta(configMapName(w.Plugin), w.Plugin),
			Data:       map[string]string{ConfigKey: string(w.Config)},
		},
		Service: &corev1.Service{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
			ObjectMeta: objectMeta(name, w.Plugin),
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{LabelPlugin: w.Plugin},
				Ports: []corev1.ServicePort{{
					Name: PortName, Protocol: corev1.ProtocolTCP,
					Port: int32(c.Topology.PluginPort), TargetPort: intstr.FromInt32(int32(c.Topology.PluginPort)),
				}},
			},
		},
		NetworkPolicy: np,
		Deployment:    deployment(w, c),
	}, nil
}

// DenyAllNetworkPolicy is the NetworkPolicy of a plugin that must not run:
// it denies all ingress and egress, so a pod still terminating reaches
// nothing.
func DenyAllNetworkPolicy(plugin string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		TypeMeta:   metav1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "NetworkPolicy"},
		ObjectMeta: objectMeta(plugins.ServiceAccountName(plugin), plugin),
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{LabelPlugin: plugin}},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
		},
	}
}

// AllowListConfigMap is the plugins.EgressAllowListConfigMap ConfigMap that
// holds the given allow-lists in the egress proxy's format
// (plugins.EncodeAllowLists): one key per plugin, empty for a plugin without
// egress hosts. A plugin without a key may reach no host.
func AllowListConfigMap(lists []plugins.EgressAllowList) (*corev1.ConfigMap, error) {
	data, err := plugins.EncodeAllowLists(lists)
	if err != nil {
		return nil, fmt.Errorf("encode egress allow-lists: %w", err)
	}
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: kindConfigMap},
		ObjectMeta: metav1.ObjectMeta{
			Name: plugins.EgressAllowListConfigMap, Namespace: plugins.Namespace,
			Labels: map[string]string{labelManagedBy: managerName},
		},
		Data: data,
	}, nil
}

func configMapName(plugin string) string { return plugins.ServiceAccountName(plugin) + "-config" }

// podLabels are the labels of a plugin's objects and pods.
func podLabels(plugin string) map[string]string {
	return map[string]string{labelName: pluginAppName, labelManagedBy: managerName, LabelPlugin: plugin}
}

func objectMeta(name, plugin string) metav1.ObjectMeta {
	return metav1.ObjectMeta{Name: name, Namespace: plugins.Namespace, Labels: podLabels(plugin)}
}

func networkPolicy(w *plugins.Workload, c *Cluster) (*networkingv1.NetworkPolicy, error) {
	np := DenyAllNetworkPolicy(w.Plugin)
	errs := make([]error, 0, len(w.NetworkPolicy.Ingress)+len(w.NetworkPolicy.Egress))
	for _, r := range w.NetworkPolicy.Ingress {
		from, err := peer(r.Peer, c)
		errs = append(errs, err)
		np.Spec.Ingress = append(np.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{
			From: from, Ports: tcpPort(r.Port),
		})
	}
	for _, r := range w.NetworkPolicy.Egress {
		to, err := peer(r.Peer, c)
		errs = append(errs, err)
		np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
			To: to, Ports: tcpPort(r.Port),
		})
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return np, nil
}

// peer selects the pods of a NetworkPolicy peer: the core's or the egress
// proxy's in the openfief namespace, or those behind the plugin's granted
// in-cluster service in that service's namespace (D-74).
func peer(p plugins.Peer, c *Cluster) ([]networkingv1.NetworkPolicyPeer, error) {
	var selector map[string]string
	switch p {
	case plugins.PeerCore:
		selector = c.CoreSelector
	case plugins.PeerEgressProxy:
		selector = c.EgressProxySelector
	case plugins.PeerClusterService:
		if len(c.ClusterServiceSelector) == 0 || c.ClusterServiceNamespace == "" || c.ClusterServiceNamespace == plugins.Namespace {
			return nil, fmt.Errorf("%w: peer %q has no usable pod selector", ErrServiceUnusable, p)
		}
		return []networkingv1.NetworkPolicyPeer{{
			NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{namespaceNameLabel: c.ClusterServiceNamespace}},
			PodSelector:       &metav1.LabelSelector{MatchLabels: maps.Clone(c.ClusterServiceSelector)},
		}}, nil
	}
	if len(selector) == 0 || selectsPluginPods(selector) {
		return nil, fmt.Errorf("%w: peer %q has no usable pod selector", ErrServiceUnusable, p)
	}
	return []networkingv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{MatchLabels: selector}}}, nil
}

func tcpPort(port uint16) []networkingv1.NetworkPolicyPort {
	p, tcp := intstr.FromInt32(int32(port)), corev1.ProtocolTCP
	return []networkingv1.NetworkPolicyPort{{Protocol: &tcp, Port: &p}}
}

func deployment(w *plugins.Workload, c *Cluster) *appsv1.Deployment {
	hash := sha256.Sum256(w.Config)
	automount, serviceLinks, nonRoot, replicas := w.ServiceAccount.AutomountAPIToken, false, true, int32(1)
	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      podLabels(w.Plugin),
			Annotations: map[string]string{AnnotationConfigHash: hex.EncodeToString(hash[:])},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName:           w.ServiceAccount.Name,
			AutomountServiceAccountToken: &automount,
			EnableServiceLinks:           &serviceLinks,
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot:   &nonRoot,
				SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
			},
			Containers: []corev1.Container{container(w, c)},
			Volumes:    volumes(w),
		},
	}
	return &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: objectMeta(plugins.ServiceAccountName(w.Plugin), w.Plugin),
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{LabelPlugin: w.Plugin}},
			// Recreate: an old pod never runs beside a new one with other permissions.
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
			Template: pod,
		},
	}
}

func container(w *plugins.Workload, c *Cluster) corev1.Container {
	yes, no := true, false
	core := c.Topology.Core
	env := make([]corev1.EnvVar, 0, len(w.Env)+operatorEnvVars)
	env = append(env,
		corev1.EnvVar{Name: EnvCoreAddress, Value: net.JoinHostPort(core.Address, strconv.FormatUint(uint64(core.Port), 10))},
		corev1.EnvVar{Name: EnvListenPort, Value: strconv.FormatUint(uint64(c.Topology.PluginPort), 10)},
		corev1.EnvVar{Name: EnvIdentityFile, Value: path.Join(w.Token.MountPath, w.Token.Path)},
		corev1.EnvVar{Name: EnvConfig, ValueFrom: &corev1.EnvVarSource{ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: configMapName(w.Plugin)}, Key: ConfigKey,
		}}},
	)
	mounts := make([]corev1.VolumeMount, 0, operatorVolumes+len(w.Secrets))
	mounts = append(mounts,
		corev1.VolumeMount{Name: identityVolume, MountPath: w.Token.MountPath, ReadOnly: true},
		corev1.VolumeMount{Name: signingKeysVolume, MountPath: w.SigningKeys.MountPath, ReadOnly: true},
	)
	if t := w.EgressToken; t != nil {
		env = append(env, corev1.EnvVar{Name: EnvEgressIdentityFile, Value: path.Join(t.MountPath, t.Path)})
		mounts = append(mounts, corev1.VolumeMount{Name: egressVolume, MountPath: t.MountPath, ReadOnly: true})
	}
	for _, e := range w.Env {
		env = append(env, corev1.EnvVar{Name: e.Name, Value: e.Value})
	}
	slices.SortFunc(env, func(a, b corev1.EnvVar) int { return cmp.Compare(a.Name, b.Name) })
	for i, s := range w.Secrets {
		mounts = append(mounts, corev1.VolumeMount{Name: secretVolume(i), MountPath: s.MountPath, ReadOnly: s.ReadOnly})
	}
	return corev1.Container{
		Name:            ContainerName,
		Image:           w.Image.String(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{{
			Name: PortName, ContainerPort: int32(c.Topology.PluginPort), Protocol: corev1.ProtocolTCP,
		}},
		Env:          env,
		VolumeMounts: mounts,
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: &no,
			Privileged:               &no,
			ReadOnlyRootFilesystem:   &yes,
			RunAsNonRoot:             &yes,
			Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
		},
	}
}

func volumes(w *plugins.Workload) []corev1.Volume {
	vs := make([]corev1.Volume, 0, operatorVolumes+len(w.Secrets))
	vs = append(vs, tokenVolume(identityVolume, w.Token), corev1.Volume{
		Name: signingKeysVolume,
		VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: w.SigningKeys.ConfigMapName},
		}},
	})
	if w.EgressToken != nil {
		vs = append(vs, tokenVolume(egressVolume, *w.EgressToken))
	}
	for i, s := range w.Secrets {
		vs = append(vs, corev1.Volume{
			Name:         secretVolume(i),
			VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: s.SecretName}},
		})
	}
	return vs
}

// tokenVolume is a projected ServiceAccount token volume named name.
func tokenVolume(name string, t plugins.ProjectedToken) corev1.Volume {
	expiration := t.ExpirationSeconds
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{Projected: &corev1.ProjectedVolumeSource{
			Sources: []corev1.VolumeProjection{{ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
				Audience: t.Audience, ExpirationSeconds: &expiration, Path: t.Path,
			}}},
		}},
	}
}

// secretVolume names the volume of the i-th Secret; Secret names can be too
// long for a volume name.
func secretVolume(i int) string { return "secret-" + strconv.Itoa(i) }
