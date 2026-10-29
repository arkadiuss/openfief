// Package kube is the views context's Kubernetes adapter: Plugins reads the
// installed plugin instances (docs/DESIGN.md D-46, D-48) and how far each
// image is trusted (D-56) from the operator's Plugin resources, for the read
// API to list them and to mark transactions of removed instances (D-58).
//
// It reads the Plugin resource as unstructured data, only the fields it
// shows, so the views context doesn't depend on the plugins context's
// Kubernetes types; a test checks the field names against them. It only
// gets, lists and watches Plugins: the core has no right to change a plugin
// or a grant (D-46).
package kube

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// The operator's runtime contract (internal/plugins/operator): its
// condition types.
const (
	conditionReady     = "Ready"
	conditionSignature = "SignatureVerified"
)

// PluginResource returns the operator's Plugin custom resource.
func PluginResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "openfief.io", Version: "v1alpha1", Resource: "plugins"}
}

// Plugins follows the Plugin resources of one namespace with an informer,
// so listing reads a cache that a watch keeps current. The core's
// ServiceAccount needs get, list and watch on plugins, which connector
// discovery already uses.
type Plugins struct {
	informer informers.GenericInformer
	factory  dynamicinformer.DynamicSharedInformerFactory
	started  atomic.Bool
}

// NewPlugins returns Plugins of namespace over the dynamic client dyn,
// relisting every resync.
func NewPlugins(dyn dynamic.Interface, namespace string, resync time.Duration) *Plugins {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dyn, resync, namespace, nil)
	return &Plugins{informer: factory.ForResource(PluginResource()), factory: factory}
}

// Start starts the watch until ctx ends and waits until the cache has
// synced. Until then InstalledPlugins fails with app.ErrPluginsUnknown.
func (p *Plugins) Start(ctx context.Context) error {
	p.factory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), p.informer.Informer().HasSynced) {
		return fmt.Errorf("syncing the plugin cache: %w", context.Cause(ctx))
	}
	p.started.Store(true)
	return nil
}

// InstalledPlugins returns every Plugin resource in the namespace, those
// being deleted included, or an error wrapping app.ErrPluginsUnknown before
// Start synced the cache.
func (p *Plugins) InstalledPlugins(context.Context) ([]views.PluginRecord, error) {
	if !p.started.Load() {
		return nil, fmt.Errorf("%w: the plugin cache has not synced", app.ErrPluginsUnknown)
	}
	objs, err := p.informer.Lister().List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("listing plugins: %w", err)
	}
	out := make([]views.PluginRecord, 0, len(objs))
	for _, o := range objs {
		if u, ok := o.(*unstructured.Unstructured); ok {
			out = append(out, RecordOf(u))
		}
	}
	return out, nil
}

// RecordOf reads what the views show of a Plugin resource: its name,
// spec.image, spec.manifest.capabilities, status.effective.capabilities,
// its Ready and SignatureVerified conditions and status.signature. A field
// that is missing or of another type reads as empty.
func RecordOf(u *unstructured.Unstructured) views.PluginRecord {
	r := views.PluginRecord{
		Instance:   u.GetName(),
		Deleting:   u.GetDeletionTimestamp() != nil,
		Image:      nestedString(u, "spec", "image"),
		Requested:  nestedStrings(u, "spec", "manifest", "capabilities"),
		Effective:  nestedStrings(u, "status", "effective", "capabilities"),
		Trust:      nestedString(u, "status", "signature", "trust"),
		SignerKind: nestedString(u, "status", "signature", "signerKind"),
		Signer:     nestedString(u, "status", "signature", "signer"),
	}
	conds, found, err := unstructured.NestedSlice(u.Object, "status", "conditions")
	if err != nil || !found {
		return r
	}
	for _, c := range conds {
		m, ok := c.(map[string]any)
		if !ok {
			continue
		}
		cond := views.Condition{Status: stringField(m, "status"), Reason: stringField(m, "reason")}
		switch m["type"] {
		case conditionReady:
			r.Ready = cond
		case conditionSignature:
			r.Signature = cond
		}
	}
	return r
}

// nestedString returns the string at fields of u, or "" when there is
// none or it isn't a string.
func nestedString(u *unstructured.Unstructured, fields ...string) string {
	s, found, err := unstructured.NestedString(u.Object, fields...)
	if err != nil || !found {
		return ""
	}
	return s
}

// nestedStrings returns the strings at fields of u, or nil when there are
// none or they aren't all strings.
func nestedStrings(u *unstructured.Unstructured, fields ...string) []string {
	s, found, err := unstructured.NestedStringSlice(u.Object, fields...)
	if err != nil || !found {
		return nil
	}
	return s
}

// stringField returns m[key] when it is a string, and "" otherwise.
func stringField(m map[string]any, key string) string {
	s, ok := m[key].(string)
	if !ok {
		return ""
	}
	return s
}
