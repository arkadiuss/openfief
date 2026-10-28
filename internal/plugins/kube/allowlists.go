package kube

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// ErrNotSynced reports that the allow-list ConfigMap could not be read before
// the context ended.
var ErrNotSynced = errors.New("egress allow-list ConfigMap not synced")

// AllowLists is app.AllowListSource over the plugins.EgressAllowListConfigMap
// ConfigMap in plugins.Namespace, kept current by a watch. It fails closed:
// while the ConfigMap is missing or cannot be decoded, no plugin has a list.
type AllowLists struct {
	lists  atomic.Pointer[map[string]plugins.EgressAllowList]
	report func(error)
}

// WatchAllowLists starts watching the allow-list ConfigMap until ctx ends and
// returns once the first read is in. report receives each ConfigMap that
// fails to decode; it must not block.
func WatchAllowLists(ctx context.Context, client kubernetes.Interface, report func(error)) (*AllowLists, error) {
	a := &AllowLists{report: report}
	a.lists.Store(&map[string]plugins.EgressAllowList{})
	factory := informers.NewSharedInformerFactoryWithOptions(client, 0,
		informers.WithNamespace(plugins.Namespace),
		informers.WithTweakListOptions(func(o *metav1.ListOptions) {
			o.FieldSelector = fields.OneTermEqualSelector("metadata.name", plugins.EgressAllowListConfigMap).String()
		}),
	)
	informer := factory.Core().V1().ConfigMaps().Informer()
	if _, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    a.update,
		UpdateFunc: func(_, obj any) { a.update(obj) },
		DeleteFunc: a.remove,
	}); err != nil {
		return nil, fmt.Errorf("watch egress allow-lists: %w", err)
	}
	factory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		return nil, ErrNotSynced
	}
	return a, nil
}

// AllowList implements app.AllowListSource.
func (a *AllowLists) AllowList(plugin string) (plugins.EgressAllowList, bool) {
	l, ok := (*a.lists.Load())[plugin]
	return l, ok
}

// update replaces the lists with the ConfigMap's, or with none if it does not decode.
func (a *AllowLists) update(obj any) {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok || cm.Name != plugins.EgressAllowListConfigMap {
		return
	}
	decoded, err := plugins.DecodeAllowLists(cm.Data)
	if err != nil {
		a.lists.Store(&map[string]plugins.EgressAllowList{})
		a.report(fmt.Errorf("ConfigMap %s/%s: %w", cm.Namespace, cm.Name, err))
		return
	}
	lists := make(map[string]plugins.EgressAllowList, len(decoded))
	for _, l := range decoded {
		lists[l.Plugin] = l
	}
	a.lists.Store(&lists)
}

// remove drops every list when the ConfigMap is deleted.
func (a *AllowLists) remove(obj any) {
	if cm, ok := obj.(*corev1.ConfigMap); ok && cm.Name != plugins.EgressAllowListConfigMap {
		return
	}
	a.lists.Store(&map[string]plugins.EgressAllowList{})
}
