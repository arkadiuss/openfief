package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// The egress proxy's pods and container, as deploy/egress-proxy names them.
const (
	EgressProxySelector  = "app.kubernetes.io/name=egress-proxy"
	EgressProxyContainer = "egress-proxy"
)

// EgressProxyLog returns the logs of the egress proxy's container in each of
// its pods in plugins.Namespace, one after another: everything each pod has
// logged since it started.
func EgressProxyLog(ctx context.Context, clients kubernetes.Interface) ([]byte, error) {
	pods, err := clients.CoreV1().Pods(plugins.Namespace).List(ctx, metav1.ListOptions{LabelSelector: EgressProxySelector})
	if err != nil {
		return nil, fmt.Errorf("listing the egress proxy's pods: %w", err)
	}
	var all []byte
	for i := range pods.Items {
		out, err := clients.CoreV1().Pods(plugins.Namespace).GetLogs(pods.Items[i].Name,
			&corev1.PodLogOptions{Container: EgressProxyContainer}).DoRaw(ctx)
		if err != nil {
			return nil, fmt.Errorf("logs of %s: %w", pods.Items[i].Name, err)
		}
		all = append(all, out...)
		if len(out) > 0 && out[len(out)-1] != '\n' {
			all = append(all, '\n')
		}
	}
	return all, nil
}

// GrantedEgress returns the hosts each PluginGrant in plugins.Namespace
// approves, by instance. A grant host that doesn't parse is an error: the
// operator would refuse it too.
func GrantedEgress(ctx context.Context, c client.Reader) (map[string][]plugins.EgressHost, error) {
	var grants v1alpha1.PluginGrantList
	if err := c.List(ctx, &grants, client.InNamespace(plugins.Namespace)); err != nil {
		return nil, fmt.Errorf("list plugin grants: %w", err)
	}
	granted := make(map[string][]plugins.EgressHost, len(grants.Items))
	for i := range grants.Items {
		g := &grants.Items[i]
		hosts := make([]plugins.EgressHost, 0, len(g.Spec.Egress))
		for _, s := range g.Spec.Egress {
			h, err := plugins.ParseEgressHost(s)
			if err != nil {
				return nil, fmt.Errorf("plugin grant %s: %w", g.Name, err)
			}
			hosts = append(hosts, h)
		}
		granted[g.Name] = hosts
	}
	return granted, nil
}
