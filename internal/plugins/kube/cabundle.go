package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// CABundleSecret is the Secret that holds an instance's extra CA bundle:
// <instance>--ca-bundle in the openfief namespace, the PEM under
// plugins.CABundleKey, labelled as the instance's (D-182), so uninstall
// deletes it.
func CABundleSecret(instance string, pem []byte) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name: plugins.SecretName(instance, plugins.CABundleRole), Namespace: plugins.Namespace,
			Labels: map[string]string{plugins.OwnerInstanceLabel: instance},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{plugins.CABundleKey: pem},
	}
}

// CABundleYAML renders, for GitOps, what `openfief plugin ca-bundle set`
// applies: the CA bundle Secret, then a partial Plugin that stamps its
// checksum, to merge into the instance's Plugin (for example as a
// kustomize patch). The owner keeps the two in step.
func CABundleYAML(instance string, pem []byte, checksum string) ([]byte, error) {
	stamp := map[string]any{
		"apiVersion": v1alpha1.GroupVersion().String(),
		"kind":       "Plugin",
		"metadata":   map[string]any{"name": instance, "namespace": plugins.Namespace},
		"spec":       map[string]any{"caBundleSha256": checksum},
	}
	var out []byte
	for _, obj := range []any{CABundleSecret(instance, pem), stamp} {
		doc, err := cleanYAML(obj)
		if err != nil {
			return nil, err
		}
		out = append(out, "---\n"...)
		out = append(out, doc...)
	}
	return out, nil
}

// PutCABundle creates the instance's CA bundle Secret, or replaces the data
// of the one that exists, which must be labelled as the instance's (D-182):
// a Secret of that name labelled otherwise, or not at all, is left alone
// and fails with app.ErrSecretNotOwned. It writes the Secret, never reads
// back more than it needs to update it, and logs nothing of it.
func (s *Instances) PutCABundle(ctx context.Context, instance string, pem []byte) error {
	want := CABundleSecret(instance, pem)
	var cur corev1.Secret
	err := s.Client.Get(ctx, client.ObjectKeyFromObject(want), &cur)
	switch {
	case apierrors.IsNotFound(err):
		if createErr := s.Client.Create(ctx, want); createErr != nil {
			return fmt.Errorf("create Secret %q: %w", want.Name, createErr)
		}
		return nil
	case err != nil:
		return fmt.Errorf("read Secret %q: %w", want.Name, err)
	case cur.Labels[plugins.OwnerInstanceLabel] != instance:
		return fmt.Errorf("%w: %s, label %s=%s to replace it", app.ErrSecretNotOwned, want.Name, plugins.OwnerInstanceLabel, instance)
	}
	cur.Data, cur.StringData = want.Data, nil
	if err := s.Client.Update(ctx, &cur); err != nil {
		return fmt.Errorf("update Secret %q: %w", want.Name, err)
	}
	return nil
}

// DeleteCABundle deletes the instance's CA bundle Secret when it is
// labelled as the instance's; one already gone is not an error.
func (s *Instances) DeleteCABundle(ctx context.Context, instance string) error {
	return s.deleteOwnedSecret(ctx, instance, plugins.SecretName(instance, plugins.CABundleRole))
}

// StampCABundle sets spec.caBundleSha256 on the instance's Plugin, or
// clears it for an empty checksum, with a merge patch that touches nothing
// else. A missing Plugin fails with app.ErrNotInstalled.
func (s *Instances) StampCABundle(ctx context.Context, instance, checksum string) error {
	var p v1alpha1.Plugin
	err := s.Client.Get(ctx, client.ObjectKey{Namespace: plugins.Namespace, Name: instance}, &p)
	switch {
	case apierrors.IsNotFound(err):
		return fmt.Errorf("%w: %q", app.ErrNotInstalled, instance)
	case err != nil:
		return fmt.Errorf("read plugin %q: %w", instance, err)
	}
	base := p.DeepCopy()
	p.Spec.CABundleSHA256 = checksum
	if err := s.Client.Patch(ctx, &p, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("stamp plugin %q: %w", instance, err)
	}
	return nil
}
