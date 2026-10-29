package operator

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"net/netip"
	"slices"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// Errors about the cluster around the plugins.
var (
	// ErrServiceUnusable reports a core or egress proxy Service the operator
	// cannot derive a plugin's network rules from.
	ErrServiceUnusable = errors.New("service unusable for plugin network rules")
	// ErrSelectorOverlap reports a core or egress proxy selector that would
	// also select plugin pods, which would let one plugin pass for the core.
	ErrSelectorOverlap = errors.New("service selector also selects plugin pods")
)

// Cluster is where the core and the egress proxy are: their ClusterIPs and
// ports (plugins have no DNS, D-32), and the pod selectors of their Services,
// which are the peers of every plugin's NetworkPolicy.
type Cluster struct {
	Topology            plugins.Topology
	CoreSelector        map[string]string
	EgressProxySelector map[string]string
	// ClusterServiceNamespace and ClusterServiceSelector select the pods
	// behind the plugin's granted in-cluster service (D-74), whose address
	// is Topology.ClusterService; empty when it has none.
	ClusterServiceNamespace string
	ClusterServiceSelector  map[string]string
}

// AddClusterService adds the plugin's granted in-cluster service to c: svc
// is that Service, read from its namespace, and port the Service port the
// grant names. The Service needs a ClusterIP, a non-empty selector and a TCP
// port of that number with a numeric targetPort, because NetworkPolicy
// matches pod ports (D-74). On an error c is unchanged.
func (c *Cluster) AddClusterService(svc *corev1.Service, port uint16) error {
	addr, err := netip.ParseAddr(svc.Spec.ClusterIP)
	if err != nil {
		return fmt.Errorf("%w: %s/%s: clusterIP %q is not an address", ErrServiceUnusable, svc.Namespace, svc.Name, svc.Spec.ClusterIP)
	}
	if len(svc.Spec.Selector) == 0 {
		return fmt.Errorf("%w: %s/%s: no pod selector", ErrServiceUnusable, svc.Namespace, svc.Name)
	}
	i := slices.IndexFunc(svc.Spec.Ports, func(p corev1.ServicePort) bool { return p.Port == int32(port) })
	if i < 0 {
		return fmt.Errorf("%w: %s/%s: no port %d", ErrServiceUnusable, svc.Namespace, svc.Name, port)
	}
	target, err := servicePort(svc, svc.Spec.Ports[i].Name)
	if err != nil {
		return err
	}
	target.Address = addr.String()
	c.Topology.ClusterService = &target
	c.ClusterServiceNamespace = svc.Namespace
	c.ClusterServiceSelector = maps.Clone(svc.Spec.Selector)
	return nil
}

// ClusterFromServices reads a Cluster from the core's and the egress proxy's
// Services. corePort and proxyPort name the Service port plugins dial; the
// port's targetPort must be a number, because NetworkPolicy matches pod
// ports. Each Service needs a ClusterIP and a non-empty selector that selects
// no plugin pod.
func ClusterFromServices(core, proxy *corev1.Service, corePort, proxyPort string, pluginPort uint16) (Cluster, error) {
	coreSvc, coreErr := serviceOf(core, corePort)
	proxySvc, proxyErr := serviceOf(proxy, proxyPort)
	if err := errors.Join(coreErr, proxyErr); err != nil {
		return Cluster{}, err
	}
	return Cluster{
		Topology:            plugins.Topology{Core: coreSvc, EgressProxy: proxySvc, PluginPort: pluginPort},
		CoreSelector:        maps.Clone(core.Spec.Selector),
		EgressProxySelector: maps.Clone(proxy.Spec.Selector),
	}, nil
}

// serviceOf returns the address and ports of the named port of s.
func serviceOf(s *corev1.Service, portName string) (plugins.Service, error) {
	addr, err := netip.ParseAddr(s.Spec.ClusterIP)
	if err != nil {
		return plugins.Service{}, fmt.Errorf("%w: %s: clusterIP %q is not an address", ErrServiceUnusable, s.Name, s.Spec.ClusterIP)
	}
	if len(s.Spec.Selector) == 0 {
		return plugins.Service{}, fmt.Errorf("%w: %s: no pod selector", ErrServiceUnusable, s.Name)
	}
	if selectsPluginPods(s.Spec.Selector) {
		return plugins.Service{}, fmt.Errorf("%w: %s", ErrSelectorOverlap, s.Name)
	}
	svc, err := servicePort(s, portName)
	if err != nil {
		return plugins.Service{}, err
	}
	svc.Address = addr.String()
	return svc, nil
}

// servicePort returns the port and numeric target port of s's TCP port
// named portName, without an address.
func servicePort(s *corev1.Service, portName string) (plugins.Service, error) {
	i := slices.IndexFunc(s.Spec.Ports, func(p corev1.ServicePort) bool { return p.Name == portName })
	if i < 0 {
		return plugins.Service{}, fmt.Errorf("%w: %s: no port named %q", ErrServiceUnusable, s.Name, portName)
	}
	p := s.Spec.Ports[i]
	t := p.TargetPort
	if t.Type == intstr.Int && t.IntVal == 0 {
		t = intstr.FromInt32(p.Port)
	}
	if (p.Protocol != corev1.ProtocolTCP && p.Protocol != "") || t.Type != intstr.Int {
		return plugins.Service{}, fmt.Errorf("%w: %s: port %q must be TCP with a numeric targetPort", ErrServiceUnusable, s.Name, portName)
	}
	port, okPort := toPort(p.Port)
	target, okTarget := toPort(t.IntVal)
	if !okPort || !okTarget {
		return plugins.Service{}, fmt.Errorf("%w: %s: port %q is out of range", ErrServiceUnusable, s.Name, portName)
	}
	return plugins.Service{Port: port, TargetPort: target}, nil
}

// toPort converts a Kubernetes port number, refusing 0 and out-of-range ones.
func toPort(p int32) (uint16, bool) {
	if p <= 0 || p > math.MaxUint16 {
		return 0, false
	}
	return uint16(p), true
}

// selectsPluginPods reports whether a Service selector matches the pods of
// some plugin: every label it requires is one plugin pods carry, with the
// value they carry (any value for the plugin-name label).
func selectsPluginPods(selector map[string]string) bool {
	fixed := podLabels("")
	for k, v := range selector {
		if k == LabelPlugin {
			continue
		}
		if have, ok := fixed[k]; !ok || have != v {
			return false
		}
	}
	return true
}
