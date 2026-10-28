package operator

import (
	"context"
	"fmt"
	"slices"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// SetupWithManager registers the reconciler: it runs on changes to a
// Plugin, its objects and its PluginGrant, and reconciles every Plugin when
// the core's or egress proxy's Service or the allow-list ConfigMap changes.
func (r *PluginReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		Named("plugin").
		For(&v1alpha1.Plugin{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Watches(&v1alpha1.PluginGrant{}, handler.EnqueueRequestsFromMapFunc(samePlugin)).
		Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(r.allPluginsIfNamed(r.Settings.CoreService, r.Settings.EgressProxyService))).
		Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(r.allPluginsIfNamed(plugins.EgressAllowListConfigMap))).
		Complete(r)
	if err != nil {
		return fmt.Errorf("set up plugin controller: %w", err)
	}
	return nil
}

// samePlugin maps a PluginGrant to the Plugin of the same name.
func samePlugin(_ context.Context, o client.Object) []reconcile.Request {
	return []reconcile.Request{{NamespacedName: types.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()}}}
}

// allPluginsIfNamed maps an object with one of the given names to every
// Plugin in the openfief namespace.
func (r *PluginReconciler) allPluginsIfNamed(names ...string) handler.MapFunc {
	return func(ctx context.Context, o client.Object) []reconcile.Request {
		if o.GetNamespace() != plugins.Namespace || !slices.Contains(names, o.GetName()) {
			return nil
		}
		var list v1alpha1.PluginList
		if err := r.Client.List(ctx, &list, client.InNamespace(plugins.Namespace)); err != nil {
			ctrl.LoggerFrom(ctx).Error(err, "list plugins")
			return nil
		}
		reqs := make([]reconcile.Request, 0, len(list.Items))
		for i := range list.Items {
			reqs = append(reqs, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&list.Items[i])})
		}
		return reqs
	}
}

// GrantReconciler records on each PluginGrant whether it is well-formed.
// Enforcing grants is the PluginReconciler's job.
type GrantReconciler struct {
	Client client.Client
}

// Reconcile sets a PluginGrant's Valid condition.
func (r *GrantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var g v1alpha1.PluginGrant
	if err := r.Client.Get(ctx, req.NamespacedName, &g); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("get plugin grant %s: %w", req.Name, err)
	}
	st := grantStatus(&g)
	if equality.Semantic.DeepEqual(st, g.Status) {
		return ctrl.Result{}, nil
	}
	g.Status = st
	if err := r.Client.Status().Update(ctx, &g); err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, fmt.Errorf("update status of plugin grant %s: %w", g.Name, err)
	}
	return ctrl.Result{}, nil
}

// grantStatus is a PluginGrant's status: Valid, or why not.
func grantStatus(g *v1alpha1.PluginGrant) v1alpha1.PluginGrantStatus {
	st := v1alpha1.PluginGrantStatus{ObservedGeneration: g.Generation, Conditions: slices.Clone(g.Status.Conditions)}
	c := metav1.Condition{
		Type: v1alpha1.ConditionValid, Status: metav1.ConditionTrue, Reason: "Valid",
		Message: "the grant is well-formed", ObservedGeneration: g.Generation,
	}
	if _, err := grantFromResource(g); err != nil {
		c.Status, c.Reason, c.Message = metav1.ConditionFalse, ReasonInvalidGrant, truncate(err.Error())
	}
	meta.SetStatusCondition(&st.Conditions, c)
	return st
}

// SetupWithManager registers the reconciler for PluginGrants.
func (r *GrantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := ctrl.NewControllerManagedBy(mgr).Named("plugingrant").For(&v1alpha1.PluginGrant{}).Complete(r); err != nil {
		return fmt.Errorf("set up plugin grant controller: %w", err)
	}
	return nil
}
