package kube

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// CABundleAccess returns the Role and RoleBinding, in the openfief
// namespace, that let the operator's ServiceAccount get one instance's CA
// bundle Secret, <instance>--ca-bundle, and no other Secret: the operator's
// own Role reads none. The owner's install creates them for an instance
// that expects an extra CA bundle, so the operator can report the Secret
// missing and restart the plugin when it changes; the uninstall deletes
// them.
func CABundleAccess(instance string) (*rbacv1.Role, *rbacv1.RoleBinding) {
	meta := metav1.ObjectMeta{
		Name: plugins.CABundleAccessName(instance), Namespace: plugins.Namespace,
		Labels: map[string]string{labelPlugin: instance, labelManagedBy: cliName},
	}
	role := &rbacv1.Role{
		TypeMeta:   metav1.TypeMeta{APIVersion: rbacAPIVersion, Kind: kindRole},
		ObjectMeta: meta,
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""}, Resources: []string{"secrets"},
			ResourceNames: []string{plugins.SecretName(instance, plugins.CABundleRole)}, Verbs: []string{"get"},
		}},
	}
	binding := &rbacv1.RoleBinding{
		TypeMeta:   metav1.TypeMeta{APIVersion: rbacAPIVersion, Kind: kindRoleBinding},
		ObjectMeta: *meta.DeepCopy(),
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacGroup, Kind: kindRole, Name: meta.Name},
		Subjects:   []rbacv1.Subject{{Kind: "ServiceAccount", Name: OperatorServiceAccount, Namespace: plugins.Namespace}},
	}
	return role, binding
}

// createCABundleAccess creates the Role and RoleBinding of CABundleAccess;
// ones that already exist are left as they are.
func (s *Instances) createCABundleAccess(ctx context.Context, instance string) error {
	role, binding := CABundleAccess(instance)
	for _, obj := range []client.Object{role, binding} {
		if err := s.Client.Create(ctx, obj); err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("create %T %s/%s: %w", obj, obj.GetNamespace(), obj.GetName(), err)
		}
	}
	return nil
}

// deleteCABundleAccess deletes the Role and RoleBinding of CABundleAccess;
// ones already gone, or never created, are skipped.
func (s *Instances) deleteCABundleAccess(ctx context.Context, instance string) []error {
	role, binding := CABundleAccess(instance)
	var errs []error
	for _, obj := range []client.Object{binding, role} {
		if err := s.Client.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			errs = append(errs, fmt.Errorf("delete %T %s/%s: %w", obj, obj.GetNamespace(), obj.GetName(), err))
		}
	}
	return errs
}
