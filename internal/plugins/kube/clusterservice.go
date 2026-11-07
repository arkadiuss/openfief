package kube

import (
	"context"
	"fmt"
	"maps"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

const (
	// OperatorServiceAccount is the operator's ServiceAccount in the
	// openfief namespace (deploy/operator).
	OperatorServiceAccount = "openfief-operator"
	// labelPlugin marks the objects the CLI creates for one instance outside
	// the openfief namespace; the operator's pods carry the same label.
	labelPlugin = "openfief.io/plugin"
	// labelManagedBy says the CLI created an object.
	labelManagedBy = "app.kubernetes.io/managed-by"
	cliName        = "openfief-cli"
	// namespaceNameLabel is the label Kubernetes sets on every namespace to
	// its name.
	namespaceNameLabel = "kubernetes.io/metadata.name"
)

// ServiceAccessName names the Role and RoleBinding that let the operator
// read one instance's in-cluster service: "openfief-plugin-<instance>".
func ServiceAccessName(instance string) string {
	return "openfief-" + plugins.ServiceAccountName(instance)
}

// ServiceAccess returns the Role and RoleBinding, in the in-cluster
// service's namespace, that let the operator's ServiceAccount get that one
// Service and nothing else (D-74). The operator's own Role stays in the
// openfief namespace (D-15); the owner's install creates these with the
// owner's kubeconfig, and the uninstall deletes them.
func ServiceAccess(instance string, cs plugins.ClusterService) (*rbacv1.Role, *rbacv1.RoleBinding) {
	meta := metav1.ObjectMeta{
		Name: ServiceAccessName(instance), Namespace: cs.Namespace,
		Labels: map[string]string{labelPlugin: instance, labelManagedBy: cliName},
	}
	role := &rbacv1.Role{
		TypeMeta:   metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "Role"},
		ObjectMeta: meta,
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""}, Resources: []string{"services"}, ResourceNames: []string{cs.Service}, Verbs: []string{"get"},
		}},
	}
	binding := &rbacv1.RoleBinding{
		TypeMeta:   metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "RoleBinding"},
		ObjectMeta: *meta.DeepCopy(),
		RoleRef:    rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "Role", Name: meta.Name},
		Subjects:   []rbacv1.Subject{{Kind: "ServiceAccount", Name: OperatorServiceAccount, Namespace: plugins.Namespace}},
	}
	return role, binding
}

// ClusterService reads an in-cluster service the owner picked (D-74): the
// target port behind the Service port it names, its pod selector, and
// whether its namespace has NetworkPolicies of its own.
func (s *Instances) ClusterService(ctx context.Context, cs plugins.ClusterService) (app.ClusterServiceInfo, error) {
	var svc corev1.Service
	if err := s.Client.Get(ctx, client.ObjectKey{Namespace: cs.Namespace, Name: cs.Service}, &svc); err != nil {
		return app.ClusterServiceInfo{}, fmt.Errorf("read service %s: %w", cs, err)
	}
	info := app.ClusterServiceInfo{Selector: maps.Clone(svc.Spec.Selector)}
	found := false
	for _, p := range svc.Spec.Ports {
		if p.Port != int32(cs.Port) {
			continue
		}
		found = true
		info.TargetPort = p.TargetPort.String()
		if p.TargetPort.Type == intstr.Int && p.TargetPort.IntVal == 0 {
			info.TargetPort = strconv.Itoa(int(p.Port))
		}
	}
	if !found {
		return app.ClusterServiceInfo{}, fmt.Errorf("%w: %s has no port %d", app.ErrClusterService, cs, cs.Port)
	}
	if len(info.Selector) == 0 {
		return app.ClusterServiceInfo{}, fmt.Errorf("%w: %s selects no pods", app.ErrClusterService, cs)
	}
	var nps networkingv1.NetworkPolicyList
	if err := s.Client.List(ctx, &nps, client.InNamespace(cs.Namespace), client.Limit(1)); err != nil {
		return app.ClusterServiceInfo{}, fmt.Errorf("list network policies in %s: %w", cs.Namespace, err)
	}
	info.NamespaceHasNetworkPolicies = len(nps.Items) > 0
	return info, nil
}

// IngressRule is the NetworkPolicy the owner adds in the in-cluster
// service's namespace when that namespace has NetworkPolicies of its own
// (D-74): it admits the instance's pods, from the openfief namespace, to
// the service's pods on the service's target port.
func IngressRule(instance string, cs plugins.ClusterService, info *app.ClusterServiceInfo) *networkingv1.NetworkPolicy {
	tcp, port := corev1.ProtocolTCP, intstr.Parse(info.TargetPort)
	return &networkingv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "NetworkPolicy"},
		ObjectMeta: metav1.ObjectMeta{
			Name: ServiceAccessName(instance), Namespace: cs.Namespace,
			Labels: map[string]string{labelPlugin: instance},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: maps.Clone(info.Selector)},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
			Ingress: []networkingv1.NetworkPolicyIngressRule{{
				From: []networkingv1.NetworkPolicyPeer{{
					NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{namespaceNameLabel: plugins.Namespace}},
					PodSelector:       &metav1.LabelSelector{MatchLabels: map[string]string{labelPlugin: instance}},
				}},
				Ports: []networkingv1.NetworkPolicyPort{{Protocol: &tcp, Port: &port}},
			}},
		},
	}
}

// IngressRuleYAML renders IngressRule as YAML for the owner to apply.
func IngressRuleYAML(instance string, cs plugins.ClusterService, info *app.ClusterServiceInfo) ([]byte, error) {
	doc, err := cleanYAML(IngressRule(instance, cs, info))
	if err != nil {
		return nil, err
	}
	return append([]byte("---\n"), doc...), nil
}

// createServiceAccess creates the Role and RoleBinding of ServiceAccess; ones
// that already exist are left as they are.
func (s *Instances) createServiceAccess(ctx context.Context, instance string, cs plugins.ClusterService) error {
	role, binding := ServiceAccess(instance, cs)
	for _, obj := range []client.Object{role, binding} {
		if err := s.Client.Create(ctx, obj); err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("create %T %s/%s: %w", obj, obj.GetNamespace(), obj.GetName(), err)
		}
	}
	return nil
}

// clusterServiceOfPlugin returns the in-cluster service the named Plugin
// names, and whether it names one: not when it is gone, names none, or
// names one so malformed that no Role was ever created for it.
func (s *Instances) clusterServiceOfPlugin(ctx context.Context, name string) (plugins.ClusterService, bool, error) {
	var p v1alpha1.Plugin
	err := s.Client.Get(ctx, client.ObjectKey{Namespace: plugins.Namespace, Name: name}, &p)
	switch {
	case apierrors.IsNotFound(err):
		return plugins.ClusterService{}, false, nil
	case err != nil:
		return plugins.ClusterService{}, false, fmt.Errorf("look up %q: %w", name, err)
	case p.Spec.ClusterService == nil:
		return plugins.ClusterService{}, false, nil
	}
	cs, csErr := clusterServiceOf(p.Spec.ClusterService)
	return cs, csErr == nil, nil
}

// clusterServiceOf reads an in-cluster service reference of a Plugin.
func clusterServiceOf(r *v1alpha1.ClusterServiceRef) (plugins.ClusterService, error) {
	cs, err := plugins.ParseClusterService(r.Namespace + "/" + r.Service + ":" + strconv.Itoa(int(r.Port)))
	if err != nil {
		return plugins.ClusterService{}, fmt.Errorf("in-cluster service: %w", err)
	}
	return cs, nil
}

// clusterServiceRef writes an in-cluster service as the API's reference.
func clusterServiceRef(cs plugins.ClusterService) v1alpha1.ClusterServiceRef {
	return v1alpha1.ClusterServiceRef{Namespace: cs.Namespace, Service: cs.Service, Port: int32(cs.Port)}
}

// deleteServiceAccess deletes the Role and RoleBinding of ServiceAccess;
// ones already gone are skipped.
func (s *Instances) deleteServiceAccess(ctx context.Context, instance string, cs plugins.ClusterService) []error {
	role, binding := ServiceAccess(instance, cs)
	var errs []error
	for _, obj := range []client.Object{binding, role} {
		if err := s.Client.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			errs = append(errs, fmt.Errorf("delete %T %s/%s: %w", obj, obj.GetNamespace(), obj.GetName(), err))
		}
	}
	return errs
}
