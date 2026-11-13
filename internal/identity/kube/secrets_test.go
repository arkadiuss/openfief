package kube_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/kube"
)

var errNotAllowed = errors.New("not allowed")

// secretsResource is the Secrets resource.
func secretsResource() schema.GroupVersionResource {
	return corev1.SchemeGroupVersion.WithResource("secrets")
}

// callsOf returns the verbs and resources client was asked for, in order,
// as "verb resource".
func callsOf(client *fake.Clientset) []string {
	actions := client.Actions()
	got := make([]string, 0, len(actions))
	for _, a := range actions {
		got = append(got, a.GetVerb()+" "+a.GetResource().Resource)
	}
	return got
}

// existing is the Secret an owner made first.
func existing(name string) *corev1.Secret {
	s := &corev1.Secret{}
	s.Name, s.Namespace = name, "openfief"
	s.Data = map[string][]byte{"key": []byte("the owner's own")}
	return s
}

// stored returns the Secret name of the openfief namespace as client holds
// it, failing t without it.
func stored(t *testing.T, client *fake.Clientset, name string) *corev1.Secret {
	t.Helper()
	obj, err := client.Tracker().Get(secretsResource(), "openfief", name)
	if err != nil {
		t.Fatalf("Secret %s: %v", name, err)
	}
	s, ok := obj.(*corev1.Secret)
	if !ok {
		t.Fatalf("Secret %s is a %T", name, obj)
	}
	return s
}

func TestSecretCreatorCreatesOnly(t *testing.T) {
	t.Parallel()
	client := fake.NewClientset()
	c := kube.NewSecretCreator(client, "openfief")

	created, err := c.CreateIfMissing(t.Context(), identity.PlatformSecrets("K", "P")[1])
	if err != nil || !created {
		t.Fatalf("CreateIfMissing = %v, %v, want true, nil", created, err)
	}
	secret := stored(t, client, "openfief-db-spicedb")
	if secret.Type != corev1.SecretTypeBasicAuth || secret.Labels["cnpg.io/reload"] != "true" ||
		string(secret.Data["username"]) != "spicedb" || string(secret.Data["password"]) != "P" {
		t.Errorf("the stored Secret has type %q, labels %v", secret.Type, secret.Labels)
	}
	if diff := cmp.Diff([]string{"create secrets"}, callsOf(client)); diff != "" {
		t.Errorf("API calls (-want +got):\n%s", diff)
	}
}

func TestSecretCreatorKeepsWhatExists(t *testing.T) {
	t.Parallel()
	client := fake.NewClientset(existing("openfief-spicedb-key"))
	c := kube.NewSecretCreator(client, "openfief")

	created, err := c.CreateIfMissing(t.Context(), identity.PlatformSecrets("K", "P")[0])
	if err != nil || created {
		t.Fatalf("CreateIfMissing = %v, %v, want false, nil", created, err)
	}
	// Only the create was asked for, and the owner's value is as it was.
	if diff := cmp.Diff([]string{"create secrets"}, callsOf(client)); diff != "" {
		t.Errorf("API calls (-want +got):\n%s", diff)
	}
	if v := string(stored(t, client, "openfief-spicedb-key").Data["key"]); v != "the owner's own" {
		t.Errorf("the owner's Secret was changed")
	}
}

func TestSecretCreatorFailures(t *testing.T) {
	t.Parallel()
	forbidden := fmt.Errorf("the API server: %w", apierrors.NewForbidden(schema.GroupResource{Resource: "secrets"}, "openfief-spicedb-key", errNotAllowed))
	client := fake.NewClientset()
	client.PrependReactor("create", "secrets", func(k8stesting.Action) (bool, runtime.Object, error) { return true, nil, forbidden })
	c := kube.NewSecretCreator(client, "openfief")

	created, err := c.CreateIfMissing(t.Context(), identity.PlatformSecrets("K", "P")[0])
	if created || !apierrors.IsForbidden(err) {
		t.Fatalf("CreateIfMissing = %v, %v, want false and the API's refusal", created, err)
	}
}

// createdNames returns the names of the Secrets client was asked to create,
// and reports every other call as a failure of t: a create of a Secret in
// the openfief namespace is the only call the credentials Job may make.
func createdNames(t *testing.T, client *fake.Clientset) []string {
	t.Helper()
	var names []string
	for _, a := range client.Actions() {
		c, isCreate := a.(k8stesting.CreateAction)
		if !isCreate || a.GetVerb() != "create" || a.GetResource().Resource != "secrets" || a.GetNamespace() != "openfief" || a.GetSubresource() != "" {
			t.Errorf("unexpected API call: %s %s in %q", a.GetVerb(), a.GetResource().Resource, a.GetNamespace())
			continue
		}
		if s, isSecret := c.GetObject().(*corev1.Secret); isSecret {
			names = append(names, s.Name)
		}
	}
	return names
}

// TestInitCreatesExactlyTheTwoSecrets runs the use case over the adapter and a
// fake API server (D-216): it creates the two named Secrets in the openfief
// namespace and nothing else, and every call it makes is a create of a
// Secret, never a get, list, watch, update, patch or delete.
func TestInitCreatesExactlyTheTwoSecrets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		objects      []runtime.Object
		wantOutcomes []app.PlatformOutcome
	}{
		{
			name:         "a fresh install",
			wantOutcomes: []app.PlatformOutcome{{Name: "openfief-spicedb-key", Created: true}, {Name: "openfief-db-spicedb", Created: true}},
		},
		{
			name:         "the owner fed both first",
			objects:      []runtime.Object{existing("openfief-spicedb-key"), existing("openfief-db-spicedb")},
			wantOutcomes: []app.PlatformOutcome{{Name: "openfief-spicedb-key"}, {Name: "openfief-db-spicedb"}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewClientset(tc.objects...)
			outcomes, err := app.InitPlatformCredentials(t.Context(), kube.NewSecretCreator(client, "openfief"), counter{})
			if err != nil {
				t.Fatalf("InitPlatformCredentials: %v", err)
			}
			if diff := cmp.Diff(tc.wantOutcomes, outcomes); diff != "" {
				t.Errorf("outcomes (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff([]string{"openfief-spicedb-key", "openfief-db-spicedb"}, createdNames(t, client)); diff != "" {
				t.Errorf("created Secrets (-want +got):\n%s", diff)
			}
			// Two Secrets exist, and both are in openfief.
			for _, name := range []string{"openfief-spicedb-key", "openfief-db-spicedb"} {
				stored(t, client, name)
			}
		})
	}
}

// counter is a deterministic io.Reader: its bytes count up from 1.
type counter struct{}

func (counter) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(i + 1)
	}
	return len(p), nil
}
