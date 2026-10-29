package plugins

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// TokenMountPath is the directory of a plugin's projected token.
	TokenMountPath = ReservedMountPath + "/openfief"
	// TokenPath is the token's file name inside TokenMountPath and
	// EgressTokenMountPath.
	TokenPath = "token"
	// EgressTokenMountPath is the directory of the projected token a plugin
	// presents to the egress proxy (D-63).
	EgressTokenMountPath = ReservedMountPath + "/openfief-egress"
	// TokenExpirationSeconds is the projected token's requested lifetime; the
	// kubelet refreshes it before it expires.
	TokenExpirationSeconds int64 = 3600

	// EnvHTTPSProxy points a plugin's HTTPS clients at the egress proxy.
	EnvHTTPSProxy = "HTTPS_PROXY"
	// EnvNoProxy keeps calls to the core off the egress proxy.
	EnvNoProxy = "NO_PROXY"

	// SigningKeysConfigMap is the ConfigMap in which the operator publishes
	// the cluster's ServiceAccount token-signing public keys and issuer
	// (D-59). Plugins can't reach the Kubernetes API, so every plugin gets it
	// mounted read-only, to verify the core's token with.
	SigningKeysConfigMap = "openfief-sa-jwks"
	// SigningKeysFile is the key of the JSON Web Key Set in
	// SigningKeysConfigMap, as the API server serves it at /openid/v1/jwks.
	SigningKeysFile = "jwks.json"
	// SigningKeysIssuerFile is the key of the token issuer in
	// SigningKeysConfigMap.
	SigningKeysIssuerFile = "issuer"
	// SigningKeysMountPath is where a plugin gets SigningKeysConfigMap.
	SigningKeysMountPath = ReservedMountPath + "/openfief-sa-jwks"

	// EnvInstance names the plugin instance (D-48): the core's token
	// carries its audience, "openfief-plugin:<instance>" (D-59).
	EnvInstance = "OPENFIEF_INSTANCE"
	// EnvCoreSubject is the core's Kubernetes username, the subject of the
	// token the core presents to a plugin (D-59).
	EnvCoreSubject = "OPENFIEF_CORE_SUBJECT"
	// EnvSigningKeysDir is the directory of the mounted signing keys.
	EnvSigningKeysDir = "OPENFIEF_SA_JWKS_DIR"
)

// Service is a cluster Service a plugin talks to. Address is the Service's
// ClusterIP, never a DNS name, because plugins have no DNS (D-32). Port is
// the Service port a client dials; TargetPort is the pod port behind it,
// which is what NetworkPolicy matches.
type Service struct {
	Address    string
	Port       uint16
	TargetPort uint16
}

// Topology is where the core and the egress proxy are, the port every
// plugin serves Connect RPC on, and the core's ServiceAccount, whose tokens
// plugins accept (D-59).
type Topology struct {
	Core               Service
	EgressProxy        Service
	PluginPort         uint16
	CoreServiceAccount string
}

// Peer is a party in a plugin's NetworkPolicy.
type Peer string

// The only peers a plugin may talk to (D-32).
const (
	PeerCore        Peer = "core"
	PeerEgressProxy Peer = "egress-proxy"
)

// NetworkRule allows TCP traffic between a plugin and Peer on Port, a pod
// port of whichever side receives the connection.
type NetworkRule struct {
	Peer Peer
	Port uint16
}

// NetworkPolicy describes a plugin pod's NetworkPolicy. Both directions are
// default-deny: only the listed rules are allowed. No rule ever allows DNS.
type NetworkPolicy struct {
	Ingress []NetworkRule
	Egress  []NetworkRule
}

// ServiceAccount is a plugin's identity. It has no Role or RoleBinding: a
// plugin never talks to the Kubernetes API, so its API token is not mounted.
type ServiceAccount struct {
	Name              string
	Namespace         string
	AutomountAPIToken bool
}

// ProjectedToken is a projected ServiceAccount token volume: the one a
// plugin presents to the core (D-27) or the one it presents to the egress
// proxy (D-63).
type ProjectedToken struct {
	Audience          string
	ExpirationSeconds int64
	MountPath         string
	Path              string
}

// SecretVolume mounts one Secret from the openfief namespace into the plugin.
type SecretVolume struct {
	SecretName string
	MountPath  string
	ReadOnly   bool
}

// ConfigMapVolume mounts one ConfigMap from the openfief namespace into the
// plugin, read-only.
type ConfigMapVolume struct {
	ConfigMapName string
	MountPath     string
}

// EnvVar is one environment variable of the plugin container.
type EnvVar struct {
	Name  string
	Value string
}

// Workload is the domain description of everything the operator creates for
// one plugin (D-11): its pod's image, identity, token, Secret mounts and
// environment, its NetworkPolicy, and its egress allow-list entry. Every list
// is sorted, so the same input always compiles to the same Workload.
type Workload struct {
	Plugin          string
	Namespace       string
	Image           ImageRef
	Capabilities    []Capability
	DataTiers       []DataTier
	ServiceAccount  ServiceAccount
	Token           ProjectedToken
	EgressToken     *ProjectedToken // nil unless some egress host is granted
	Secrets         []SecretVolume
	SigningKeys     ConfigMapVolume
	Env             []EnvVar
	NetworkPolicy   NetworkPolicy
	EgressAllowList EgressAllowList
	Config          []byte
}

// Compile turns effective permissions into a Workload:
//   - ingress only from the core, on the plugin port;
//   - egress to the core, plus the egress proxy only when some egress host is
//     granted; never DNS (D-32);
//   - HTTPS_PROXY set to the proxy only when some egress host is granted, with
//     NO_PROXY set to the core;
//   - the granted Secrets, mounted read-only;
//   - a ServiceAccount without Kubernetes API access, and a projected token
//     with audience CoreAudience (D-27), in the openfief namespace (D-15);
//   - only when some egress host is granted, a second projected token with
//     audience EgressAudience for the egress proxy (D-63), so the proxy never
//     sees a token the core would accept;
//   - the cluster's ServiceAccount signing keys, mounted read-only, with the
//     instance's name and the core's username in the environment, so the
//     plugin can verify the core's token (D-59).
func Compile(e *EffectivePermissions, t Topology) (Workload, error) {
	if e == nil || e.plugin == "" {
		return Workload{}, fmt.Errorf("%w: empty effective permissions", ErrNoCapabilityGranted)
	}
	if err := t.validate(); err != nil {
		return Workload{}, err
	}
	w := Workload{
		Plugin:         e.plugin,
		Namespace:      Namespace,
		Image:          e.image,
		Capabilities:   e.Capabilities(),
		DataTiers:      e.DataTiers(),
		ServiceAccount: ServiceAccount{Name: ServiceAccountName(e.plugin), Namespace: Namespace},
		Token: ProjectedToken{
			Audience: CoreAudience, ExpirationSeconds: TokenExpirationSeconds,
			MountPath: TokenMountPath, Path: TokenPath,
		},
		NetworkPolicy: NetworkPolicy{
			Ingress: []NetworkRule{{Peer: PeerCore, Port: t.PluginPort}},
			Egress:  []NetworkRule{{Peer: PeerCore, Port: t.Core.TargetPort}},
		},
		EgressAllowList: EgressAllowList{Plugin: e.plugin, Hosts: e.Egress()},
		Config:          e.Config(),
		SigningKeys:     ConfigMapVolume{ConfigMapName: SigningKeysConfigMap, MountPath: SigningKeysMountPath},
	}
	for _, s := range e.secrets {
		w.Secrets = append(w.Secrets, SecretVolume{SecretName: s.Name, MountPath: s.MountPath, ReadOnly: true})
	}
	if len(e.egress) > 0 {
		w.NetworkPolicy.Egress = append(w.NetworkPolicy.Egress, NetworkRule{Peer: PeerEgressProxy, Port: t.EgressProxy.TargetPort})
		w.EgressToken = &ProjectedToken{
			Audience: EgressAudience, ExpirationSeconds: TokenExpirationSeconds,
			MountPath: EgressTokenMountPath, Path: TokenPath,
		}
		w.Env = []EnvVar{
			{Name: EnvHTTPSProxy, Value: "http://" + t.EgressProxy.hostPort()},
			{Name: EnvNoProxy, Value: t.Core.Address},
		}
	}
	w.Env = append(w.Env,
		EnvVar{Name: EnvCoreSubject, Value: serviceAccountUserFmt + t.CoreServiceAccount},
		EnvVar{Name: EnvInstance, Value: e.plugin},
		EnvVar{Name: EnvSigningKeysDir, Value: SigningKeysMountPath},
	)
	return w, nil
}

func (t Topology) validate() error {
	var errs []error
	if t.PluginPort == 0 {
		errs = append(errs, fieldErr("pluginPort", ErrInvalidTopology, "0"))
	}
	if !validPluginName(t.CoreServiceAccount) {
		errs = append(errs, fieldErr("coreServiceAccount", ErrInvalidTopology, t.CoreServiceAccount))
	}
	errs = append(errs, t.Core.validate("core")...)
	errs = append(errs, t.EgressProxy.validate("egressProxy")...)
	return errors.Join(errs...)
}

func (s Service) validate(field string) []error {
	var errs []error
	if s.Address == "" || strings.ContainsAny(s.Address, "/@[] \t\n") {
		errs = append(errs, fieldErr(field+".address", ErrInvalidTopology, s.Address))
	}
	if s.Port == 0 || s.TargetPort == 0 {
		errs = append(errs, fieldErr(field+".port", ErrInvalidTopology, s.hostPort()))
	}
	return errs
}

// hostPort returns "address:port", bracketing an IPv6 address.
func (s Service) hostPort() string {
	host := s.Address
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return host + ":" + strconv.FormatUint(uint64(s.Port), 10)
}
