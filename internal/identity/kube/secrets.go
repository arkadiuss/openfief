package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arkadiuss/openfief/internal/identity"
)

// SecretCreator creates the platform's Secrets in one namespace (D-216)
// with the ServiceAccount of the pod it runs in, which holds the one right
// `create` on secrets: it never gets, lists, updates, patches or deletes
// one, and a Secret that exists answers 409 and stays as it is.
type SecretCreator struct {
	client    kubernetes.Interface
	namespace string
}

// NewSecretCreator returns a SecretCreator for namespace over client.
func NewSecretCreator(client kubernetes.Interface, namespace string) *SecretCreator {
	return &SecretCreator{client: client, namespace: namespace}
}

// CreateIfMissing creates s and reports true; when the API server answers
// that the Secret already exists, it reports false and does nothing more.
// Any other failure is returned, without s's data.
func (c *SecretCreator) CreateIfMissing(ctx context.Context, s identity.PlatformSecret) (bool, error) {
	data := make(map[string][]byte, len(s.Data))
	for k, v := range s.Data {
		data[k] = []byte(v)
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: s.Name, Namespace: c.namespace, Labels: s.Labels},
		Type:       corev1.SecretType(s.Type),
		Data:       data,
	}
	_, err := c.client.CoreV1().Secrets(c.namespace).Create(ctx, secret, metav1.CreateOptions{})
	switch {
	case err == nil:
		return true, nil
	case apierrors.IsAlreadyExists(err):
		return false, nil
	}
	// An API error names the object and the reason, never the data of a Secret.
	return false, fmt.Errorf("creating Secret %s: %w", s.Name, err)
}
