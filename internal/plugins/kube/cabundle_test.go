package kube_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/kube"
)

// caSum is a synthetic CA bundle checksum.
const caSum = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"

// caYAML is what `ca-bundle set rates --print` writes for the bundle
// "PEM": the Secret, then the Plugin's stamp.
const caYAML = `---
apiVersion: v1
data:
  ca.crt: UEVN
kind: Secret
metadata:
  labels:
    openfief.io/instance: rates
  name: rates--ca-bundle
  namespace: openfief
type: Opaque
---
apiVersion: openfief.io/v1alpha1
kind: Plugin
metadata:
  name: rates
  namespace: openfief
spec:
  caBundleSha256: ` + caSum + `
`

// caPlan is an install of instance rates, a connector whose manifest asks
// for the optional CA bundle Secret, expecting a bundle.
func caPlan(t *testing.T) (plugins.Plugin, plugins.Grant) {
	t.Helper()
	img, err := plugins.ParseImageRef(sampleImage)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	signed := plugins.SignedImage{Image: img, Manifest: plugins.Manifest{
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Secrets:      []plugins.SecretMount{{Role: plugins.CABundleRole, MountPath: "/etc/openfief/ca-bundle", Optional: true}},
	}}
	p, g, err := plugins.PlanInstall("rates", &signed, nil, nil)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := p.SetExtraCA(true); err != nil {
		t.Fatalf("SetExtraCA: %v", err)
	}
	return p, g
}

func TestCABundleYAML(t *testing.T) {
	t.Parallel()
	got, err := kube.CABundleYAML("rates", []byte("PEM"), caSum)
	if err != nil {
		t.Fatalf("CABundleYAML: %v", err)
	}
	if diff := cmp.Diff(caYAML, string(got)); diff != "" {
		t.Errorf("YAML (-want +got):\n%s", diff)
	}
}

// TestInstancesPutCABundle checks that the CLI's store creates the CA
// bundle Secret labelled as the instance's, replaces its data on a
// rotation, deletes it, and leaves alone and reports one labelled as
// another instance's.
func TestInstancesPutCABundle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := fakeCluster(t)
	for _, pem := range []string{"first", "second"} {
		if err := store.PutCABundle(ctx, "rates", []byte(pem)); err != nil {
			t.Fatalf("PutCABundle(%s): %v", pem, err)
		}
		checkCABundleSecret(t, store, pem)
	}
	for range 2 { // the second time it is already gone
		if err := store.DeleteCABundle(ctx, "rates"); err != nil {
			t.Fatalf("DeleteCABundle: %v", err)
		}
		checkCABundleSecret(t, store, "")
	}
	foreign := fakeCluster(t, ownedSecret("rates--ca-bundle", "other"))
	if err := foreign.PutCABundle(ctx, "rates", []byte("first")); !errors.Is(err, app.ErrSecretNotOwned) {
		t.Errorf("PutCABundle over another instance's Secret: %v, want %v", err, app.ErrSecretNotOwned)
	}
}

// TestInstancesStampCABundle checks that the store stamps and clears the
// checksum on the Plugin without touching the rest of its spec, and that
// stamping a Plugin that isn't installed fails with app.ErrNotInstalled.
func TestInstancesStampCABundle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := fakeCluster(t)
	p, g := caPlan(t)
	if err := store.Create(ctx, &p, &g); err != nil {
		t.Fatalf("create: %v", err)
	}
	for _, sum := range []string{caSum, ""} {
		if err := store.StampCABundle(ctx, "rates", sum); err != nil {
			t.Fatalf("StampCABundle(%q): %v", sum, err)
		}
		got, _, err := store.Get(ctx, "rates")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		want := p
		want.CABundleSHA256 = sum
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Plugin after stamping %q (-want +got):\n%s", sum, diff)
		}
	}
	if err := store.StampCABundle(ctx, "other", caSum); !errors.Is(err, app.ErrNotInstalled) {
		t.Errorf("StampCABundle of no Plugin: %v, want %v", err, app.ErrNotInstalled)
	}
}

// checkCABundleSecret fails t unless rates's CA bundle Secret holds exactly
// pem under ca.crt, labelled as rates's; for an empty pem, unless it is
// gone.
func checkCABundleSecret(t *testing.T, store *kube.Instances, pem string) {
	t.Helper()
	var s corev1.Secret
	err := store.Client.Get(context.Background(), client.ObjectKey{Namespace: "openfief", Name: "rates--ca-bundle"}, &s)
	if pem == "" {
		if !apierrors.IsNotFound(err) {
			t.Errorf("Secret after delete: %v, want not found", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("get Secret: %v", err)
	}
	if diff := cmp.Diff(map[string][]byte{"ca.crt": []byte(pem)}, s.Data); diff != "" {
		t.Errorf("Secret data (-want +got):\n%s", diff)
	}
	if s.Labels[plugins.OwnerInstanceLabel] != "rates" {
		t.Errorf("Secret labels %v, want %s=rates", s.Labels, plugins.OwnerInstanceLabel)
	}
}

// TestInstancesKeepCABundleStamp checks that the checksum stamped on a
// Plugin survives the resource the CLI writes, as spec.caBundleSha256,
// and reads back for an upgrade.
func TestInstancesKeepCABundleStamp(t *testing.T) {
	t.Parallel()
	p, g := caPlan(t)
	if err := p.StampCABundle(caSum); err != nil {
		t.Fatalf("StampCABundle: %v", err)
	}
	plugin, grant := kube.Resources(&p, &g)
	if plugin.Spec.CABundleSHA256 != caSum {
		t.Errorf("spec.caBundleSha256 = %q, want %q", plugin.Spec.CABundleSHA256, caSum)
	}
	got, _, err := fakeCluster(t, plugin, grant).Get(context.Background(), "rates")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.CABundleSHA256 != caSum {
		t.Errorf("stamp read back = %q, want %q", got.CABundleSHA256, caSum)
	}
}
