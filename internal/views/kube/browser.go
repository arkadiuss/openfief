package kube

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"
	"sync/atomic"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// The operator's runtime contract for a plugin's Service
// (internal/plugins/operator): its labels and the name of its port.
const (
	labelPlugin    = "openfief.io/plugin"
	labelName      = "app.kubernetes.io/name"
	pluginAppName  = "openfief-plugin"
	portName       = "connect"
	capabilityView = "view"
)

// ErrNoAddress reports a view plugin instance without a Service the core
// can call.
var ErrNoAddress = errors.New("view plugin instance has no address")

// Browser follows the view plugins that serve browsers (D-81): the Plugin
// resources of one namespace whose running version may act as a view and
// has a granted browser path, and the Services the operator creates for
// plugins, with informers. The core's ServiceAccount needs get, list and
// watch on both, which connector discovery already uses.
type Browser struct {
	plugins  informers.GenericInformer
	services cache.SharedIndexInformer
	factory  dynamicinformer.DynamicSharedInformerFactory
	core     informers.SharedInformerFactory
	started  atomic.Bool
}

// NewBrowser returns a Browser of namespace over the dynamic client dyn
// and client, relisting every resync.
func NewBrowser(dyn dynamic.Interface, client kubernetes.Interface, namespace string, resync time.Duration) *Browser {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dyn, resync, namespace, nil)
	core := informers.NewSharedInformerFactoryWithOptions(client, resync, informers.WithNamespace(namespace),
		informers.WithTweakListOptions(func(o *metav1.ListOptions) { o.LabelSelector = labelName + "=" + pluginAppName }))
	return &Browser{
		plugins:  factory.ForResource(PluginResource()),
		services: core.Core().V1().Services().Informer(),
		factory:  factory,
		core:     core,
	}
}

// Start starts the watches until ctx ends and waits until the caches have
// synced. Until then BrowserRoutes fails with app.ErrViewsUnknown.
func (b *Browser) Start(ctx context.Context) error {
	b.factory.Start(ctx.Done())
	b.core.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), b.plugins.Informer().HasSynced, b.services.HasSynced) {
		return fmt.Errorf("syncing the view plugin caches: %w", context.Cause(ctx))
	}
	b.started.Store(true)
	return nil
}

// BrowserRoutes returns the route of every view plugin that serves
// browsers now (BrowserRouteOf), in instance order, or an error wrapping
// app.ErrViewsUnknown before Start synced the caches.
func (b *Browser) BrowserRoutes(context.Context) ([]views.BrowserRoute, error) {
	if !b.started.Load() {
		return nil, fmt.Errorf("%w: the plugin cache has not synced", app.ErrViewsUnknown)
	}
	objs, err := b.plugins.Lister().List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("listing plugins: %w", err)
	}
	var routes []views.BrowserRoute
	for _, o := range objs {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			continue
		}
		if r, ok := BrowserRouteOf(u); ok {
			routes = append(routes, r)
		}
	}
	slices.SortFunc(routes, func(a, c views.BrowserRoute) int { return compareStrings(a.Instance, c.Instance) })
	return routes, nil
}

// BaseURL returns the URL the core calls instance at: http:// and its
// Service's ClusterIP and connect port, so no DNS lookup is needed.
func (b *Browser) BaseURL(instance string) (string, error) {
	for _, o := range b.services.GetStore().List() {
		svc, ok := o.(*corev1.Service)
		if !ok || svc.Labels[labelPlugin] != instance {
			continue
		}
		if addr, ok := serviceAddress(svc); ok {
			return "http://" + addr, nil
		}
	}
	return "", fmt.Errorf("%w: %s", ErrNoAddress, instance)
}

// BrowserRouteOf returns the browser route of a Plugin resource, and
// whether it has one: it must not be being deleted, its Ready condition
// must be True, view must be among its effective capabilities, and its
// effective permissions must hold a browser path, the one its manifest
// declares and its grant approves (D-81).
func BrowserRouteOf(u *unstructured.Unstructured) (views.BrowserRoute, bool) {
	r := RecordOf(u)
	if r.Deleting || r.Ready.Status != string(metav1.ConditionTrue) || !slices.Contains(r.Effective, capabilityView) {
		return views.BrowserRoute{}, false
	}
	paths := nestedStrings(u, "status", "effective", "browserPaths")
	if len(paths) == 0 || paths[0] == "" {
		return views.BrowserRoute{}, false
	}
	return views.BrowserRoute{Instance: r.Instance, Prefix: paths[0]}, true
}

// serviceAddress returns a plugin Service's "ClusterIP:port" for its
// connect port, or false when it has no ClusterIP or no such port.
func serviceAddress(svc *corev1.Service) (string, bool) {
	ip := svc.Spec.ClusterIP
	if ip == "" || ip == corev1.ClusterIPNone {
		return "", false
	}
	for _, p := range svc.Spec.Ports {
		if p.Name == portName && p.Port > 0 {
			return net.JoinHostPort(ip, strconv.Itoa(int(p.Port))), true
		}
	}
	return "", false
}

// compareStrings orders a before b.
func compareStrings(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}
