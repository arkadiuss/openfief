package main

import (
	"bytes"
	"crypto/rand"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	identitykube "github.com/arkadiuss/openfief/internal/identity/kube"
)

// storedSecret returns the Secret name of the openfief namespace as client
// holds it, failing t without it.
func storedSecret(t *testing.T, client *fake.Clientset, name string) *corev1.Secret {
	t.Helper()
	obj, err := client.Tracker().Get(corev1.SchemeGroupVersion.WithResource("secrets"), "openfief", name)
	if err != nil {
		t.Fatalf("Secret %s: %v", name, err)
	}
	s, ok := obj.(*corev1.Secret)
	if !ok {
		t.Fatalf("Secret %s is a %T", name, obj)
	}
	return s
}

// createdSecrets returns the names of the Secrets client was asked to
// create, and fails t for every other call: a create of a Secret in
// openfief is the only call the credentials Job may make (D-216).
func createdSecrets(t *testing.T, client *fake.Clientset) []string {
	t.Helper()
	var names []string
	for _, a := range client.Actions() {
		c, isCreate := a.(k8stesting.CreateAction)
		if !isCreate || a.GetVerb() != "create" || a.GetResource().Resource != "secrets" || a.GetNamespace() != "openfief" {
			t.Errorf("unexpected API call: %s %s in %q", a.GetVerb(), a.GetResource().Resource, a.GetNamespace())
			continue
		}
		if s, isSecret := c.GetObject().(*corev1.Secret); isSecret {
			names = append(names, s.Name)
		}
	}
	return names
}

// TestInitCredentials runs the command's body over a fake API server that
// records every call (D-216): it creates the SpiceDB preshared key's Secret
// and the role's basic-auth Secret in openfief and nothing else, never gets,
// lists, updates, patches or deletes one, keeps one that exists, and logs
// names and outcomes, never a value.
func TestInitCredentials(t *testing.T) {
	t.Parallel()
	owners := &corev1.Secret{}
	owners.Name, owners.Namespace = "openfief-spicedb-key", "openfief"
	owners.Data = map[string][]byte{"key": []byte("owner-fed")}
	tests := []struct {
		name      string
		objects   []runtime.Object
		wantLog   []string
		wantOwner bool // the key's Secret is the owner's, and stays so
	}{
		{
			name:    "a fresh install creates both",
			wantLog: []string{`"secret":"openfief-spicedb-key","result":"created"`, `"secret":"openfief-db-spicedb","result":"created"`},
		},
		{
			name:    "an owner-fed Secret is kept",
			objects: []runtime.Object{owners}, wantOwner: true,
			wantLog: []string{`"secret":"openfief-spicedb-key","result":"kept"`, `"secret":"openfief-db-spicedb","result":"created"`},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewClientset(tc.objects...)
			var logs bytes.Buffer
			log := slog.New(slog.NewJSONHandler(&logs, nil))
			if err := initCredentials(t.Context(), identitykube.NewSecretCreator(client, "openfief"), rand.Reader, log); err != nil {
				t.Fatalf("initCredentials: %v", err)
			}
			if diff := cmp.Diff([]string{"openfief-spicedb-key", "openfief-db-spicedb"}, createdSecrets(t, client)); diff != "" {
				t.Errorf("Secrets asked for (-want +got):\n%s", diff)
			}
			checkLog(t, client, logs.String(), tc.wantLog)
			if tc.wantOwner && string(storedSecret(t, client, "openfief-spicedb-key").Data["key"]) != "owner-fed" {
				t.Error("the owner's Secret was changed")
			}
		})
	}
}

// checkLog fails t unless log holds each of want, and if it holds a value of a
// Secret client stores.
func checkLog(t *testing.T, client *fake.Clientset, log string, want []string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(log, w) {
			t.Errorf("the log lacks %s:\n%s", w, log)
		}
	}
	for _, name := range []string{"openfief-spicedb-key", "openfief-db-spicedb"} {
		for k, v := range storedSecret(t, client, name).Data {
			if len(v) >= 32 && strings.Contains(log, string(v)) {
				t.Errorf("the log holds the value of %s.%s", name, k)
			}
		}
	}
}
