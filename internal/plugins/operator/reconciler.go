package operator

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// FieldOwner is the server-side apply field manager of everything the
// operator writes.
const FieldOwner = "openfief-operator"

// Settings name the core's and the egress proxy's Services and ports, the
// port plugins serve on, and the core's ServiceAccount, whose tokens plugins
// accept (D-59).
type Settings struct {
	CoreService        string
	CorePort           string
	EgressProxyService string
	EgressProxyPort    string
	PluginPort         uint16
	CoreServiceAccount string
}

// The operator's Role in the openfief namespace (D-15), generated into
// deploy/operator/role.yaml by `task gen`.
//
// +kubebuilder:rbac:groups=openfief.io,namespace=openfief,resources=plugins;plugingrants,verbs=get;list;watch
// +kubebuilder:rbac:groups=openfief.io,namespace=openfief,resources=plugins/status;plugingrants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openfief.io,namespace=openfief,resources=plugins/finalizers,verbs=update
// +kubebuilder:rbac:groups="",namespace=openfief,resources=serviceaccounts;configmaps;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,namespace=openfief,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,namespace=openfief,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",namespace=openfief,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=coordination.k8s.io,namespace=openfief,resources=leases,verbs=get;list;watch;create;update;patch;delete

// PluginReconciler reconciles one Plugin: it decides which version runs
// (plugins.PlanRollout), applies or removes its objects, republishes the
// egress allow-list ConfigMap and writes the Plugin's status.
type PluginReconciler struct {
	Client   client.Client
	Scheme   *runtime.Scheme
	Verifier ImageVerifier
	Settings Settings
}

// Reconcile brings one Plugin's objects, the allow-list and its status in
// line with its spec, its PluginGrant and the cluster.
func (r *PluginReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Namespace != plugins.Namespace {
		return ctrl.Result{}, nil
	}
	var p v1alpha1.Plugin
	if err := r.Client.Get(ctx, req.NamespacedName, &p); err != nil {
		if apierrors.IsNotFound(err) {
			// Garbage collection removes its objects; drop its allow-list entry.
			return ctrl.Result{}, r.publishAllowList(ctx, req.Name, nil)
		}
		return ctrl.Result{}, fmt.Errorf("get plugin %s: %w", req.Name, err)
	}
	o, err := r.observe(ctx, &p)
	if err != nil {
		return ctrl.Result{}, err
	}
	if o.running() {
		if o.deploymentReady, err = r.apply(ctx, &p, o); err != nil {
			return ctrl.Result{}, err
		}
	} else if err := r.teardown(ctx, &p); err != nil {
		return ctrl.Result{}, err
	}
	var lists []plugins.EgressAllowList
	if o.running() { // apply may have found the cluster unusable
		lists = []plugins.EgressAllowList{{Plugin: p.Name, Hosts: o.rollout.Run.Egress()}}
	}
	if err := r.publishAllowList(ctx, p.Name, lists); err != nil {
		return ctrl.Result{}, err
	}
	return o.result(), r.writeStatus(ctx, &p, o)
}

// verifyRetry is how soon an image whose signature could not be checked,
// such as while its registry is unreachable, is checked again.
const verifyRetry = time.Minute

// result asks for another reconcile when a signature could not be checked.
func (o *observation) result() ctrl.Result {
	unknown := o.signature.Status == metav1.ConditionUnknown ||
		(o.previousSignature != nil && o.previousSignature.Status == metav1.ConditionUnknown)
	if unknown {
		return ctrl.Result{RequeueAfter: verifyRetry}
	}
	return ctrl.Result{}
}

// observe reads the Plugin's grant, checks its current version and decides
// the rollout. Problems that are the owner's to fix are recorded in the
// observation; only API failures are returned.
func (r *PluginReconciler) observe(ctx context.Context, p *v1alpha1.Plugin) (*observation, error) {
	o := &observation{generation: p.Generation}
	current, specErr := pluginFromSpec(p.Name, &p.Spec)
	o.specErr = specErr
	if current != nil {
		o.configErr = ValidateConfig(current.Manifest.ConfigSchema, current.Config)
	}
	o.signature = r.verify(ctx, p.Spec.Image, declaredManifest(current))
	grant, err := r.grant(ctx, p.Name)
	if err != nil {
		return nil, err
	}
	o.grantErr = grant.err
	if o.grantErr == nil {
		r.plan(ctx, p, o, current, grant.grant)
	}
	return o, nil
}

// plan decides the rollout and which version's spec runs. The previous
// version runs only if its image is still verified.
func (r *PluginReconciler) plan(ctx context.Context, p *v1alpha1.Plugin, o *observation, current *plugins.Plugin, grant *plugins.Grant) {
	previous := previousVersion(p)
	o.rollout, o.planErr = plugins.PlanRollout(plugins.RolloutInput{
		Current:           current,
		CurrentDeployable: o.specErr == nil && o.configErr == nil && o.signature.Verified(),
		Previous:          previous,
		Grant:             grant,
	})
	switch {
	case o.planErr != nil || o.rollout.Run == nil:
	case o.rollout.Outcome == plugins.RolloutPrevious:
		if v := r.verify(ctx, p.Status.Deployed.Image, declaredManifest(previous)); !v.Verified() {
			o.previousSignature = &v
			return
		}
		o.runSpec = p.Status.Deployed
	default:
		o.runSpec = &p.Spec
	}
}

// previousVersion is the version that runs now, from the Plugin's status;
// nil when none runs or the record is unreadable.
func previousVersion(p *v1alpha1.Plugin) *plugins.Plugin {
	if p.Status.Deployed == nil {
		return nil
	}
	previous, err := pluginFromSpec(p.Name, p.Status.Deployed)
	if err != nil {
		return nil
	}
	return previous
}

// verify checks an image reference's signature and, when declared is not
// nil, that its signed manifest label is declared (D-42); an unparsable
// reference is never verified.
func (r *PluginReconciler) verify(ctx context.Context, image string, declared *plugins.Manifest) Verification {
	ref, err := plugins.ParseImageRef(image)
	if err != nil {
		return Verification{Status: metav1.ConditionFalse, Reason: ReasonInvalidImage, Message: err.Error()}
	}
	return r.Verifier.VerifyImage(ctx, ref, declared)
}

// declaredManifest is a valid plugin version's manifest; nil when the
// version is invalid, which never runs.
func declaredManifest(p *plugins.Plugin) *plugins.Manifest {
	if p == nil {
		return nil
	}
	return &p.Manifest
}

type grantResult struct {
	grant *plugins.Grant
	err   error
}

// grant reads the PluginGrant of the plugin's name; no grant is a nil grant.
func (r *PluginReconciler) grant(ctx context.Context, name string) (grantResult, error) {
	var g v1alpha1.PluginGrant
	err := r.Client.Get(ctx, types.NamespacedName{Namespace: plugins.Namespace, Name: name}, &g)
	switch {
	case apierrors.IsNotFound(err):
		return grantResult{}, nil
	case err != nil:
		return grantResult{}, fmt.Errorf("get plugin grant %s: %w", name, err)
	}
	grant, err := grantFromResource(&g)
	return grantResult{grant: grant, err: err}, nil
}

// cluster reads the core's and the egress proxy's Services.
func (r *PluginReconciler) cluster(ctx context.Context) (Cluster, error) {
	var core, proxy corev1.Service
	s := r.Settings
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: plugins.Namespace, Name: s.CoreService}, &core); err != nil {
		return Cluster{}, fmt.Errorf("core service %s: %w", s.CoreService, err)
	}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: plugins.Namespace, Name: s.EgressProxyService}, &proxy); err != nil {
		return Cluster{}, fmt.Errorf("egress proxy service %s: %w", s.EgressProxyService, err)
	}
	c, err := ClusterFromServices(&core, &proxy, s.CorePort, s.EgressProxyPort, s.PluginPort)
	c.Topology.CoreServiceAccount = s.CoreServiceAccount
	return c, err
}

// apply compiles the running version's workload and applies its objects,
// the NetworkPolicy before the Deployment. It reports whether the Deployment
// is available. When the cluster cannot be read, nothing runs.
func (r *PluginReconciler) apply(ctx context.Context, p *v1alpha1.Plugin, o *observation) (bool, error) {
	c, err := r.cluster(ctx)
	var objs Objects
	if err == nil {
		var w plugins.Workload
		if w, err = plugins.Compile(o.rollout.Run, c.Topology); err == nil {
			objs, err = BuildObjects(&w, &c)
		}
	}
	if err != nil {
		o.clusterErr, o.runSpec = err, nil
		return false, r.teardown(ctx, p)
	}
	for _, obj := range []client.Object{objs.ServiceAccount, objs.ConfigMap, objs.NetworkPolicy, objs.Service} {
		if err := r.applyOwned(ctx, p, obj); err != nil {
			return false, err
		}
	}
	if err := r.applyOwned(ctx, p, objs.Deployment); err != nil {
		return false, err
	}
	return deploymentAvailable(objs.Deployment), nil
}

// deploymentAvailable reports whether d's current generation has an
// available, updated replica.
func deploymentAvailable(d *appsv1.Deployment) bool {
	s := d.Status
	return s.ObservedGeneration >= d.Generation && s.UpdatedReplicas >= 1 && s.AvailableReplicas >= 1
}

// teardown removes a plugin that must not run: its NetworkPolicy becomes
// deny-all, then its Deployment, Service, config and ServiceAccount go,
// which also invalidates its tokens.
func (r *PluginReconciler) teardown(ctx context.Context, p *v1alpha1.Plugin) error {
	if err := r.applyOwned(ctx, p, DenyAllNetworkPolicy(p.Name)); err != nil {
		return err
	}
	name := plugins.ServiceAccountName(p.Name)
	om := metav1.ObjectMeta{Name: name, Namespace: plugins.Namespace}
	cm := metav1.ObjectMeta{Name: configMapName(p.Name), Namespace: plugins.Namespace}
	for _, obj := range []client.Object{
		&appsv1.Deployment{ObjectMeta: om}, &corev1.Service{ObjectMeta: om},
		&corev1.ConfigMap{ObjectMeta: cm}, &corev1.ServiceAccount{ObjectMeta: om},
	} {
		err := r.Client.Delete(ctx, obj)
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("delete %s: %w", obj.GetName(), err)
		}
	}
	return nil
}

// applyOwned server-side applies obj, controlled by p, and reads the result
// back into obj.
func (r *PluginReconciler) applyOwned(ctx context.Context, p *v1alpha1.Plugin, obj client.Object) error {
	if err := controllerutil.SetControllerReference(p, obj, r.Scheme); err != nil {
		return fmt.Errorf("owner of %s: %w", obj.GetName(), err)
	}
	return r.applyObject(ctx, obj)
}

// applyObject server-side applies obj as FieldOwner.
func (r *PluginReconciler) applyObject(ctx context.Context, obj client.Object) error {
	return r.applyObjectAs(ctx, obj, FieldOwner)
}

// applyObjectAs server-side applies obj as the given field manager and reads
// the result back into obj.
func (r *PluginReconciler) applyObjectAs(ctx context.Context, obj client.Object, owner string) error {
	return applyAs(ctx, r.Client, obj, owner)
}

// applyAs server-side applies obj with c as the given field manager and
// reads the result back into obj.
func applyAs(ctx context.Context, c client.Client, obj client.Object, owner string) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return fmt.Errorf("convert %s: %w", obj.GetName(), err)
	}
	// Fields a typed object always carries but never sets.
	unstructured.RemoveNestedField(u, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(u, "spec", "template", "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(u, "status")
	applied := &unstructured.Unstructured{Object: u}
	if err := c.Apply(ctx, client.ApplyConfigurationFromUnstructured(applied), client.FieldOwner(owner), client.ForceOwnership); err != nil {
		return fmt.Errorf("apply %s %s: %w", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName(), err)
	}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(applied.Object, obj); err != nil {
		return fmt.Errorf("read back %s: %w", obj.GetName(), err)
	}
	return nil
}

// publishAllowList sets self's entry in the allow-list ConfigMap to lists,
// which holds self's allow-list while self runs, and is empty to remove the
// entry. Each plugin's entry has its own
// server-side apply field manager, so a reconcile changes only its own
// plugin's key and never acts on another plugin's possibly stale state.
func (r *PluginReconciler) publishAllowList(ctx context.Context, self string, lists []plugins.EgressAllowList) error {
	cm, err := AllowListConfigMap(lists)
	if err != nil {
		return err
	}
	return r.applyObjectAs(ctx, cm, allowListFieldOwner(self))
}

// allowListFieldOwner is the field manager of one plugin's allow-list entry.
func allowListFieldOwner(plugin string) string { return FieldOwner + "/allowlist/" + plugin }

// writeStatus records the observation, if it changed anything.
func (r *PluginReconciler) writeStatus(ctx context.Context, p *v1alpha1.Plugin, o *observation) error {
	st := pluginStatus(o, p.Status.Conditions)
	if equality.Semantic.DeepEqual(st, p.Status) {
		return nil
	}
	p.Status = st
	if err := r.Client.Status().Update(ctx, p); err != nil {
		return fmt.Errorf("update status of plugin %s: %w", p.Name, err)
	}
	return nil
}
