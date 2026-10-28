package operator_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
)

// service is a synthetic Service with one named port.
func service(name, clusterIP string, selector map[string]string, port *corev1.ServicePort) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "openfief"},
		Spec:       corev1.ServiceSpec{ClusterIP: clusterIP, Selector: selector, Ports: []corev1.ServicePort{*port}},
	}
}

func TestClusterFromServices(t *testing.T) {
	t.Parallel()
	coreRPC := corev1.ServicePort{Name: "connect", Port: 80, TargetPort: intstr.FromInt32(8080)}
	proxyPort := corev1.ServicePort{Name: "proxy", Protocol: corev1.ProtocolTCP, Port: 4750}
	tests := []struct {
		name        string
		core, proxy *corev1.Service
		want        operator.Cluster
		wantErr     error
	}{
		{
			name:  "ClusterIPs, target ports and selectors",
			core:  service("openfief-core", "192.0.2.10", coreSelector(), &coreRPC),
			proxy: service("openfief-egress-proxy", "2001:db8::20", proxySelector(), &proxyPort),
			want: operator.Cluster{
				Topology: plugins.Topology{
					Core:        plugins.Service{Address: "192.0.2.10", Port: 80, TargetPort: 8080},
					EgressProxy: plugins.Service{Address: "2001:db8::20", Port: 4750, TargetPort: 4750},
					PluginPort:  9090,
				},
				CoreSelector: coreSelector(), EgressProxySelector: proxySelector(),
			},
		},
		{
			name:    "headless core",
			core:    service("openfief-core", "None", coreSelector(), &coreRPC),
			proxy:   service("openfief-egress-proxy", "192.0.2.20", proxySelector(), &proxyPort),
			wantErr: operator.ErrServiceUnusable,
		},
		{
			name:    "named target port",
			core:    service("openfief-core", "192.0.2.10", coreSelector(), &corev1.ServicePort{Name: "connect", Port: 80, TargetPort: intstr.FromString("rpc")}),
			proxy:   service("openfief-egress-proxy", "192.0.2.20", proxySelector(), &proxyPort),
			wantErr: operator.ErrServiceUnusable,
		},
		{
			name:    "UDP port",
			core:    service("openfief-core", "192.0.2.10", coreSelector(), &coreRPC),
			proxy:   service("openfief-egress-proxy", "192.0.2.20", proxySelector(), &corev1.ServicePort{Name: "proxy", Protocol: corev1.ProtocolUDP, Port: 4750}),
			wantErr: operator.ErrServiceUnusable,
		},
		{
			name:    "no port of that name",
			core:    service("openfief-core", "192.0.2.10", coreSelector(), &corev1.ServicePort{Name: "http", Port: 80}),
			proxy:   service("openfief-egress-proxy", "192.0.2.20", proxySelector(), &proxyPort),
			wantErr: operator.ErrServiceUnusable,
		},
		{
			name:    "core without a selector",
			core:    service("openfief-core", "192.0.2.10", nil, &coreRPC),
			proxy:   service("openfief-egress-proxy", "192.0.2.20", proxySelector(), &proxyPort),
			wantErr: operator.ErrServiceUnusable,
		},
		{
			name:    "core selector that selects one plugin's pods",
			core:    service("openfief-core", "192.0.2.10", map[string]string{"openfief.io/plugin": "evil"}, &coreRPC),
			proxy:   service("openfief-egress-proxy", "192.0.2.20", proxySelector(), &proxyPort),
			wantErr: operator.ErrSelectorOverlap,
		},
		{
			name: "proxy selector that selects every plugin's pods",
			core: service("openfief-core", "192.0.2.10", coreSelector(), &coreRPC),
			proxy: service("openfief-egress-proxy", "192.0.2.20", map[string]string{
				"app.kubernetes.io/name": "openfief-plugin", "app.kubernetes.io/managed-by": "openfief-operator",
			}, &proxyPort),
			wantErr: operator.ErrSelectorOverlap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := operator.ClusterFromServices(tt.core, tt.proxy, "connect", "proxy", 9090)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ClusterFromServices error mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ClusterFromServices mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
