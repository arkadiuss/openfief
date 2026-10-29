package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/cli"
	"github.com/arkadiuss/openfief/internal/plugins/cosign/cosigntest"
	"github.com/arkadiuss/openfief/internal/plugins/kube"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// sampleLabel is a synthetic connector's manifest label.
const sampleLabel = `{"capabilities":["connector"],"egress":["api.example.com"],` +
	`"secrets":[{"name":"sample-credentials","mountPath":"/etc/openfief/sample"}],` +
	`"configSchema":{"type":"object","properties":{"start":{"type":"string"}},"additionalProperties":false}}`

// world is a registry with a key-signed and an unsigned image, a fake
// cluster, and the owner's signers file trusting that key.
type world struct {
	signed, unsigned plugins.ImageRef
	signersFile      string
	signersJSON      string
	dir              string
}

func newWorld(t *testing.T) *world {
	t.Helper()
	reg := cosigntest.NewRegistry(t)
	w := &world{
		signed:   reg.Push(t, "owner/sample", sampleLabel),
		unsigned: reg.Push(t, "owner/unsigned", `{"capabilities":["view"]}`),
		dir:      t.TempDir(),
	}
	key := reg.SignWithKey(t, w.signed)
	doc, err := json.Marshal(map[string]any{"signers": []any{map[string]string{"name": "my-key", "publicKey": key}}})
	if err != nil {
		t.Fatalf("signers: %v", err)
	}
	w.signersJSON = string(doc)
	w.signersFile = w.file(t, "signers.json", w.signersJSON)
	return w
}

// file writes a file in the world's directory and returns its path.
func (w *world) file(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(w.dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

// runCLI runs the real command line against c, answering in.
func runCLI(t *testing.T, c client.Client, in string, args ...string) (string, error) {
	t.Helper()
	env, err := newEnv()
	if err != nil {
		t.Fatalf("env: %v", err)
	}
	var out bytes.Buffer
	env.In, env.Out, env.Err = strings.NewReader(in), &out, &out
	env.Cluster = func(string, string) (cli.Cluster, error) { return cluster{&kube.Instances{Client: c}}, nil }
	err = cli.Run(context.Background(), args, env)
	return out.String(), err
}

// fakeClient is a fake API server holding objs.
func fakeClient(t *testing.T, objs ...client.Object) *kube.Instances {
	t.Helper()
	scheme, err := newScheme()
	if err != nil {
		t.Fatalf("scheme: %v", err)
	}
	return &kube.Instances{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()}
}

func TestInstallVerifiesSignatures(t *testing.T) {
	t.Parallel()
	w := newWorld(t)
	good := w.file(t, "good.json", `{"start":"2025-01-02"}`)
	bad := w.file(t, "bad.json", `{"start":1}`)
	signersCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Namespace: plugins.Namespace, Name: kube.TrustedSignersConfigMap},
		Data:       map[string]string{kube.TrustedSignersKey: w.signersJSON},
	}
	tests := []struct {
		name    string
		objs    []client.Object
		args    []string
		wantErr error
	}{
		{name: "owner's key from a file", args: []string{w.signed.String(), "--trusted-signers", w.signersFile, "--config", good}},
		{name: "owner's key from the cluster", objs: []client.Object{signersCM}, args: []string{w.signed.String(), "--config", good}},
		{name: "key not trusted counts as unsigned", args: []string{w.signed.String()}, wantErr: plugins.ErrUnsigned},
		{name: "unsigned", args: []string{w.unsigned.String(), "--trusted-signers", w.signersFile}, wantErr: plugins.ErrUnsigned},
		{name: "config the schema rejects", args: []string{w.signed.String(), "--trusted-signers", w.signersFile, "--config", bad}, wantErr: app.ErrConfig},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := fakeClient(t, tt.objs...)
			out, err := runCLI(t, store.Client, "y\n", append([]string{"plugin", "install", "--name", "sample"}, tt.args...)...)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v\n%s", err, tt.wantErr, out)
			}
			if err != nil {
				return
			}
			checkInstalled(t, store, w.signed.String(), out)
		})
	}
}

// checkInstalled checks that sample runs image with a grant approving
// exactly its signed manifest.
func checkInstalled(t *testing.T, store *kube.Instances, image, out string) {
	t.Helper()
	if !strings.Contains(out, "signed by trusted signer my-key (Owner") || !strings.Contains(out, "Installed sample.") {
		t.Errorf("output:\n%s", out)
	}
	var p v1alpha1.Plugin
	var g v1alpha1.PluginGrant
	key := client.ObjectKey{Namespace: plugins.Namespace, Name: "sample"}
	if err := errors.Join(store.Client.Get(context.Background(), key, &p), store.Client.Get(context.Background(), key, &g)); err != nil {
		t.Fatalf("get: %v", err)
	}
	if p.Spec.Image != image || string(p.Spec.Config.Raw) != `{"start":"2025-01-02"}` {
		t.Errorf("plugin spec = %+v", p.Spec)
	}
	want := v1alpha1.Permissions{Capabilities: []string{"connector"}, Egress: []string{"api.example.com:443"}, Secrets: []string{"sample-credentials"}}
	if diff := cmp.Diff(want, g.Spec); diff != "" {
		t.Errorf("grant (-want +got):\n%s", diff)
	}
}

func TestPrint(t *testing.T) {
	t.Parallel()
	w := newWorld(t)
	store := fakeClient(t)
	out, err := runCLI(t, store.Client, "", "plugin", "install", w.signed.String(), "--print", "--trusted-signers", w.signersFile)
	if err != nil {
		t.Fatalf("print: %v\n%s", err, out)
	}
	if !strings.Contains(out, "kind: PluginGrant") || !strings.Contains(out, "name: sample\n") {
		t.Errorf("printed:\n%s", out)
	}
	listed, err := store.List(context.Background())
	if err != nil || len(listed) != 0 {
		t.Errorf("--print changed the cluster: %+v (error %v)", listed, err)
	}
}

func TestInstallListUninstall(t *testing.T) {
	t.Parallel()
	w := newWorld(t)
	store := fakeClient(t)
	for _, args := range [][]string{
		{"install", w.signed.String(), "--trusted-signers", w.signersFile, "--yes"},
		{"list"},
		{"uninstall", "sample", "--yes"},
	} {
		out, err := runCLI(t, store.Client, "", append([]string{"plugin"}, args...)...)
		if err != nil || !strings.Contains(out, "sample") {
			t.Fatalf("%s: %v\n%s", args[0], err, out)
		}
	}
	for _, obj := range []client.Object{&v1alpha1.Plugin{}, &v1alpha1.PluginGrant{}} {
		err := store.Client.Get(context.Background(), client.ObjectKey{Namespace: plugins.Namespace, Name: "sample"}, obj)
		if !apierrors.IsNotFound(err) {
			t.Errorf("%T after uninstall: %v", obj, err)
		}
	}
}
