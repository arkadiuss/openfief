// Package kube is the ingest context's Kubernetes adapter: Discovery finds
// the connector instances that run in the cluster (docs/DESIGN.md D-48,
// D-49) and the address of each one's Service, by watching the operator's
// Plugin resources and the Services it creates for them.
//
// It reads the Plugin resource as unstructured data, only the fields it
// needs, so the ingest context doesn't depend on the plugins context's
// Kubernetes types. The names it relies on are the operator's runtime
// contract: the plugin label, the app name label and the connect port.
package kube

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// The operator's runtime contract (internal/plugins/operator).
const (
	// LabelPlugin names the plugin an object belongs to.
	LabelPlugin = "openfief.io/plugin"
	// LabelName and PluginAppName mark every object of every plugin.
	LabelName     = "app.kubernetes.io/name"
	PluginAppName = "openfief-plugin"
	// PortName names the Connect RPC port of a plugin's Service.
	PortName = "connect"
	// CapabilityConnector is the capability of a connector plugin.
	CapabilityConnector = "connector"
)

// PluginResource returns the operator's Plugin custom resource.
func PluginResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "openfief.io", Version: "v1alpha1", Resource: "plugins"}
}

// Discovery errors.
var (
	// ErrNotStarted reports a Discovery used before Start synced it.
	ErrNotStarted = errors.New("connector discovery not started")
	// ErrNoAddress reports an instance without a Service the core can call.
	ErrNoAddress = errors.New("connector instance has no address")
)

// Discovery follows the cluster's Plugin resources and plugin Services in
// one namespace with informers, so listing reads a cache that watches keep
// current. The core's ServiceAccount needs get, list and watch on both.
type Discovery struct {
	plugins  informers.GenericInformer
	services cache.SharedIndexInformer
	factory  dynamicinformer.DynamicSharedInformerFactory
	core     informers.SharedInformerFactory
	started  bool
}

// NewDiscovery returns a Discovery of namespace over the dynamic client dyn
// and client, which relists every resync.
func NewDiscovery(dyn dynamic.Interface, client kubernetes.Interface, namespace string, resync time.Duration) *Discovery {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dyn, resync, namespace, nil)
	core := informers.NewSharedInformerFactoryWithOptions(client, resync, informers.WithNamespace(namespace),
		informers.WithTweakListOptions(func(o *metav1.ListOptions) { o.LabelSelector = LabelName + "=" + PluginAppName }))
	return &Discovery{
		plugins:  factory.ForResource(PluginResource()),
		services: core.Core().V1().Services().Informer(),
		factory:  factory,
		core:     core,
	}
}

// Start starts the watches until ctx ends and waits until the caches have
// synced.
func (d *Discovery) Start(ctx context.Context) error {
	d.factory.Start(ctx.Done())
	d.core.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), d.plugins.Informer().HasSynced, d.services.HasSynced) {
		return fmt.Errorf("syncing the plugin caches: %w", context.Cause(ctx))
	}
	d.started = true
	return nil
}

// ConnectorInstances returns the names of the connector instances that run
// now, in name order: Plugins whose running version may act as a connector
// and that are Ready, with a Service the core can call.
func (d *Discovery) ConnectorInstances(context.Context) ([]string, error) {
	if !d.started {
		return nil, ErrNotStarted
	}
	objs, err := d.plugins.Lister().List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("listing plugins: %w", err)
	}
	var names []string
	for _, o := range objs {
		u, ok := o.(*unstructured.Unstructured)
		if !ok || !RunningConnector(u) {
			continue
		}
		if _, err := d.BaseURL(u.GetName()); err == nil {
			names = append(names, u.GetName())
		}
	}
	slices.Sort(names)
	return names, nil
}

// BaseURL returns the URL the core calls instance at: http:// and its
// Service's ClusterIP and connect port, so no DNS lookup is needed.
func (d *Discovery) BaseURL(instance string) (string, error) {
	for _, o := range d.services.GetStore().List() {
		svc, ok := o.(*corev1.Service)
		if !ok || svc.Labels[LabelPlugin] != instance {
			continue
		}
		if addr, ok := ServiceAddress(svc); ok {
			return "http://" + addr, nil
		}
	}
	return "", fmt.Errorf("%w: %s", ErrNoAddress, instance)
}

// RunningConnector reports whether a Plugin resource runs a version that
// may act as a connector: not being deleted, its Ready condition True, and
// connector among its effective capabilities.
func RunningConnector(u *unstructured.Unstructured) bool {
	if u.GetDeletionTimestamp() != nil {
		return false
	}
	caps, _, err := unstructured.NestedStringSlice(u.Object, "status", "effective", "capabilities")
	if err != nil || !slices.Contains(caps, CapabilityConnector) {
		return false
	}
	conds, _, err := unstructured.NestedSlice(u.Object, "status", "conditions")
	if err != nil {
		return false
	}
	for _, c := range conds {
		m, ok := c.(map[string]any)
		if ok && m["type"] == "Ready" && m["status"] == string(metav1.ConditionTrue) {
			return true
		}
	}
	return false
}

// ServiceAddress returns a plugin Service's "ClusterIP:port" for its connect
// port, or false when it has no ClusterIP or no such port.
func ServiceAddress(svc *corev1.Service) (string, bool) {
	ip := svc.Spec.ClusterIP
	if ip == "" || ip == corev1.ClusterIPNone {
		return "", false
	}
	for _, p := range svc.Spec.Ports {
		if p.Name == PortName && p.Port > 0 {
			return net.JoinHostPort(ip, strconv.Itoa(int(p.Port))), true
		}
	}
	return "", false
}
