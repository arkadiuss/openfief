//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/catalog"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// The instances the operator trust tests apply (P-10's acceptance item 11,
// D-200), each its own, so the tests run side by side.
const (
	// noGrantPlugin runs the signed NBP FX image with no PluginGrant at
	// first.
	noGrantPlugin = "nogrant-e2e"
	// taggedPlugin names its image by tag next to its digest.
	taggedPlugin = "tagged-e2e"
	// upgradePlugin runs the synthetic connector and is upgraded to its
	// image that asks for an egress host more.
	upgradePlugin = "upgrade-e2e"
	// catalogPlugin runs the NBP FX image the catalog pins, copied to the
	// suite's registry; ownerPlugin the one this checkout built, signed
	// with the suite's key.
	catalogPlugin = "catalog-e2e"
	ownerPlugin   = "owner-e2e"
	// catalogRepo is where the suite copies the catalog's image to.
	catalogRepo = "nbp-fx-catalog"
	// catalogEntry is the catalog entry whose image the suite copies.
	catalogEntry = "nbp-fx"
	// catalogPublisher is that entry's publisher's name.
	catalogPublisher = "openfief"
	// upgradeEgress is the host only the upgrade's manifest asks for, as
	// the operator writes it.
	upgradeEgress = "example.org:443"
	// upgradeAllowList is the instance's allow-list entry once the grant
	// approves it.
	upgradeAllowList = upgradeEgress + "\n"
)

// errNoSignatures reports a catalog image without signatures to copy.
var errNoSignatures = errors.New("no signature in its registry")

// TestOperatorRequiresGrant applies a Plugin of the signed NBP FX image
// with no PluginGrant, as plain kubectl would, and checks that the operator
// verifies its signature yet runs nothing: Ready False with reason NoGrant,
// everything the manifest requests pending approval, no Deployment, no pod,
// no allow-list entry and no deployed version. Applying the grant then
// makes it run, so the grant alone held it back
// (SCEN-e2e-plugin-no-grant-not-deployed).
func TestOperatorRequiresGrant(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	img, err := reg.image(t.Context(), fxImage)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { deleteInstance(t, k, noGrantPlugin) })
	if err := k.applyOne(t.Context(), fxPluginObject(t, noGrantPlugin, img.cluster.String())); err != nil {
		t.Fatal(err)
	}
	// The operator writes every condition at once: once the signature is
	// verified, the status says what that decision ran.
	p := waitCondition(t, k, noGrantPlugin, v1alpha1.ConditionSignatureVerified, metav1.ConditionTrue, operator.ReasonTrustedSigner)
	want := instanceState{
		Ready:     string(metav1.ConditionFalse) + "/" + operator.ReasonNoGrant,
		Approval:  string(metav1.ConditionTrue) + "/" + operator.ReasonPermissionsPending,
		Signature: string(metav1.ConditionTrue) + "/" + operator.ReasonTrustedSigner,
		Pending: &v1alpha1.Permissions{
			Capabilities: []string{"connector"}, Egress: []string{fxEgress + ":443"}, Institutions: &[]string{"nbp"},
		},
	}
	if diff := cmp.Diff(want, stateOf(t, k, p)); diff != "" {
		t.Errorf("plugin %s without a grant (-want +got):\n%s", noGrantPlugin, diff)
	}
	if err := k.applyOne(t.Context(), fxGrant(noGrantPlugin)); err != nil {
		t.Fatal(err)
	}
	waitPluginReady(t, k, noGrantPlugin)
}

// TestOperatorRefusesTagReference checks both ways an image named by tag is
// refused (D-17): the API server refuses a Plugin whose image is a tag
// alone, as plain kubectl would apply it, so it never exists; one that names
// a tag next to its digest passes the CRD's pattern, and the operator, with
// a grant that approves everything requested, refuses it as InvalidImage and
// runs nothing (SCEN-e2e-plugin-tag-refused).
func TestOperatorRefusesTagReference(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	img, err := reg.image(t.Context(), fxImage)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { deleteInstance(t, k, taggedPlugin) })
	tagged := reg.cluster + "/" + fxImage + ":" + pluginTag

	err = k.applyOne(t.Context(), fxPluginObject(t, taggedPlugin, tagged))
	if !apierrors.IsInvalid(err) {
		t.Fatalf("applying a Plugin with image %s: %v; want the API server to refuse it as invalid", tagged, err)
	}
	if _, getErr := k.getPlugin(t.Context(), taggedPlugin); !apierrors.IsNotFound(getErr) {
		t.Fatalf("plugin %s after the refused apply: %v; want not found", taggedPlugin, getErr)
	}

	both := tagged + "@" + img.cluster.Digest()
	for _, obj := range []*unstructured.Unstructured{fxGrant(taggedPlugin), fxPluginObject(t, taggedPlugin, both)} {
		if err := k.applyOne(t.Context(), obj); err != nil {
			t.Fatal(err)
		}
	}
	p := waitCondition(t, k, taggedPlugin, v1alpha1.ConditionSignatureVerified, metav1.ConditionFalse, operator.ReasonInvalidImage)
	want := instanceState{
		Ready:     string(metav1.ConditionFalse) + "/" + operator.ReasonInvalidSpec,
		Approval:  string(metav1.ConditionUnknown) + "/" + operator.ReasonInvalidSpec,
		Signature: string(metav1.ConditionFalse) + "/" + operator.ReasonInvalidImage,
	}
	if diff := cmp.Diff(want, stateOf(t, k, p)); diff != "" {
		t.Errorf("plugin %s with image %s (-want +got):\n%s", taggedPlugin, both, diff)
	}
}

// TestOperatorUpgradeWaitsForGrant runs the synthetic connector, signed by
// the suite's key, with a grant for exactly what it requests, then applies
// an upgrade to its second image, whose signed manifest asks for the egress
// host example.org too. Until the grant approves the host, the operator
// keeps the first version running, on its own image and without the host
// on the allow-list, and lists the host as pending approval; once the grant
// approves it, the upgrade runs, with the host on the allow-list
// (SCEN-e2e-plugin-upgrade-waits-for-grant).
func TestOperatorUpgradeWaitsForGrant(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	first, err := reg.image(t.Context(), connectorImage)
	if err != nil {
		t.Fatal(err)
	}
	second, err := reg.image(t.Context(), connectorEgressImage)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { deleteInstance(t, k, upgradePlugin) })
	createCredential(t, k, upgradePlugin)
	apply := func(objs ...*unstructured.Unstructured) {
		t.Helper()
		for _, obj := range objs {
			if err := k.applyOne(t.Context(), obj); err != nil {
				t.Fatal(err)
			}
		}
	}
	manifests := readConnectorManifests(t)
	apply(connectorGrant(upgradePlugin), connectorPlugin(t, upgradePlugin, first.cluster, manifests.first))
	waitPluginReady(t, k, upgradePlugin)

	apply(connectorPlugin(t, upgradePlugin, second.cluster, manifests.second))
	p := waitCondition(t, k, upgradePlugin, v1alpha1.ConditionReady, metav1.ConditionTrue, operator.ReasonRunningPrevious)
	want := instanceState{
		Ready:      string(metav1.ConditionTrue) + "/" + operator.ReasonRunningPrevious,
		Approval:   string(metav1.ConditionTrue) + "/" + operator.ReasonPermissionsPending,
		Signature:  string(metav1.ConditionTrue) + "/" + operator.ReasonTrustedSigner,
		Pending:    &v1alpha1.Permissions{Egress: []string{upgradeEgress}},
		Deployed:   first.cluster.String(),
		Deployment: first.cluster.String(),
		Pods:       true,
	}
	if diff := cmp.Diff(want, stateOf(t, k, p)); diff != "" {
		t.Errorf("plugin %s while its upgrade waits for the grant (-want +got):\n%s", upgradePlugin, diff)
	}

	grant := connectorGrant(upgradePlugin)
	if err := unstructured.SetNestedStringSlice(grant.Object, []string{"example.org"}, "spec", "egress"); err != nil {
		t.Fatal(err)
	}
	apply(grant)
	p = waitCondition(t, k, upgradePlugin, v1alpha1.ConditionReady, metav1.ConditionTrue, operator.ReasonRunning)
	want = instanceState{
		Ready:      string(metav1.ConditionTrue) + "/" + operator.ReasonRunning,
		Approval:   string(metav1.ConditionFalse) + "/" + operator.ReasonApproved,
		Signature:  string(metav1.ConditionTrue) + "/" + operator.ReasonTrustedSigner,
		Deployed:   second.cluster.String(),
		Deployment: second.cluster.String(),
		Pods:       true,
		AllowList:  upgradeAllowList,
	}
	if diff := cmp.Diff(want, stateOf(t, k, p)); diff != "" {
		t.Errorf("plugin %s once the grant approves its upgrade (-want +got):\n%s", upgradePlugin, diff)
	}
}

// TestOperatorVerifiedOnlyWhenCatalogPinned copies the NBP FX image the
// embedded catalog pins, with its signatures from OpenFief's CI, from GHCR
// into the suite's registry at the same digest, and applies it and the NBP
// FX image this checkout built, signed by the suite's key, an owner's
// trusted signer, each as a Plugin without a grant, so nothing runs. The
// operator verifies both, offline against its bundled trust root, and its
// status calls only the catalog's image Verified, signed by the catalog's
// publisher; the owner's is Unverified, signed by the owner's signer
// (D-44, D-56; SCEN-e2e-plugin-verified-when-catalog-pinned).
func TestOperatorVerifiedOnlyWhenCatalogPinned(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	pinned := copyCatalogImage(t, reg)
	owner, err := reg.image(t.Context(), fxImage)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		deleteInstance(t, k, catalogPlugin)
		deleteInstance(t, k, ownerPlugin)
	})
	for _, obj := range []*unstructured.Unstructured{bareFXPlugin(t, catalogPlugin, pinned), fxPluginObject(t, ownerPlugin, owner.cluster.String())} {
		if err := k.applyOne(t.Context(), obj); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []struct {
		instance, reason string
		want             *v1alpha1.SignatureStatus
	}{
		{
			instance: catalogPlugin, reason: operator.ReasonVerifiedPlugin,
			want: &v1alpha1.SignatureStatus{
				Trust: string(plugins.Verified), SignerKind: string(plugins.SignerPublisher), Signer: catalogPublisher,
			},
		},
		{
			instance: ownerPlugin, reason: operator.ReasonTrustedSigner,
			want: &v1alpha1.SignatureStatus{
				Trust: string(plugins.Unverified), SignerKind: string(plugins.SignerOwner), Signer: signerName,
			},
		},
	} {
		p := waitCondition(t, k, c.instance, v1alpha1.ConditionSignatureVerified, metav1.ConditionTrue, c.reason)
		if diff := cmp.Diff(c.want, p.Status.Signature); diff != "" {
			t.Errorf("plugin %s: status.signature (-want +got):\n%s", c.instance, diff)
		}
	}
}

// copyCatalogImage copies the image the embedded catalog pins for
// catalogEntry, every platform's image and every signature referring to it,
// from its registry into repository catalogRepo of the suite's registry,
// and returns it there at the same digest. Only the digest is pinned
// (plugins.CatalogEntry), so the copy is the catalog's image.
func copyCatalogImage(t *testing.T, reg suiteRegistry) registryImage {
	t.Helper()
	src := catalogImage(t)
	from, err := name.NewDigest(src.String())
	if err != nil {
		t.Fatalf("catalog image %s: %v", src, err)
	}
	to, err := name.NewDigest(reg.local + "/" + catalogRepo + "@" + src.Digest())
	if err != nil {
		t.Fatalf("copy of %s: %v", src, err)
	}
	t.Logf("e2e: copying the catalog's %s image %s, with its signatures, to %s", catalogEntry, src, to)
	if copyErr := copyManifest(t.Context(), from, to); copyErr != nil {
		t.Fatal(copyErr)
	}
	if copyErr := copyReferrers(t.Context(), from, to); copyErr != nil {
		t.Fatal(copyErr)
	}
	cluster, err := plugins.ParseImageRef(reg.cluster + "/" + catalogRepo + "@" + src.Digest())
	if err != nil {
		t.Fatal(err)
	}
	local, err := plugins.ParseImageRef(to.String())
	if err != nil {
		t.Fatal(err)
	}
	return registryImage{local: local, cluster: cluster}
}

// catalogImage is the image the embedded catalog pins for catalogEntry.
func catalogImage(t *testing.T) plugins.ImageRef {
	t.Helper()
	c, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	for i := range c.Entries {
		if c.Entries[i].Name == catalogEntry {
			return c.Entries[i].Image
		}
	}
	t.Fatalf("the embedded catalog has no entry %s", catalogEntry)
	return plugins.ImageRef{}
}

// copyReferrers copies every manifest that refers to from, such as its
// signatures, to the repository of to, each at its own digest.
func copyReferrers(ctx context.Context, from, to name.Digest) error {
	refs, err := remote.Referrers(from, remote.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("listing what refers to %s: %w", from, err)
	}
	index, err := refs.IndexManifest()
	if err != nil {
		return fmt.Errorf("listing what refers to %s: %w", from, err)
	}
	if len(index.Manifests) == 0 {
		return fmt.Errorf("%w: %s", errNoSignatures, from)
	}
	for i := range index.Manifests {
		d := index.Manifests[i].Digest.String()
		if copyErr := copyManifest(ctx, from.Context().Digest(d), to.Context().Digest(d)); copyErr != nil {
			return copyErr
		}
	}
	return nil
}

// copyManifest copies the manifest from names, an image or an index with
// every image it lists, or a signature with its bundle, to to, keeping its
// digest; a manifest with a subject also lands on the registry's
// referrers tag when the registry has no Referrers API.
func copyManifest(ctx context.Context, from, to name.Digest) error {
	desc, err := remote.Get(from, remote.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("reading %s: %w", from, err)
	}
	if desc.MediaType.IsIndex() {
		idx, idxErr := desc.ImageIndex()
		if idxErr != nil {
			return fmt.Errorf("reading %s: %w", from, idxErr)
		}
		if writeErr := remote.WriteIndex(to, idx, remote.WithContext(ctx)); writeErr != nil {
			return fmt.Errorf("copying %s to %s: %w", from, to, writeErr)
		}
		return nil
	}
	img, err := desc.Image()
	if err != nil {
		return fmt.Errorf("reading %s: %w", from, err)
	}
	if writeErr := remote.Write(to, img, remote.WithContext(ctx)); writeErr != nil {
		return fmt.Errorf("copying %s to %s: %w", from, to, writeErr)
	}
	return nil
}

// bareFXPlugin is a Plugin of the NBP FX connector image as the cluster
// names it, declaring the manifest its signed label carries, read from its
// linux/amd64 image, with no config.
func bareFXPlugin(t *testing.T, instance string, img registryImage) *unstructured.Unstructured {
	t.Helper()
	ref, err := name.NewDigest(img.local.String())
	if err != nil {
		t.Fatal(err)
	}
	platform := v1.Platform{OS: "linux", Architecture: "amd64"}
	image, err := remote.Image(ref, remote.WithContext(t.Context()), remote.WithPlatform(platform))
	if err != nil {
		t.Fatalf("reading %s: %v", ref, err)
	}
	cfg, err := image.ConfigFile()
	if err != nil {
		t.Fatalf("reading %s's config: %v", ref, err)
	}
	var manifest map[string]any
	if err := json.Unmarshal([]byte(cfg.Config.Labels[plugins.ManifestLabel]), &manifest); err != nil {
		t.Fatalf("%s's manifest label: %v", ref, err)
	}
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "openfief.io/v1alpha1", "kind": "Plugin",
		"metadata": map[string]any{"name": instance, "namespace": namespace},
		"spec":     map[string]any{"image": img.cluster.String(), "manifest": manifest},
	}}
}

// instanceState is what the trust tests compare of an instance: its
// conditions as "<status>/<reason>", what waits for approval, the image of
// its deployed version and of its Deployment's plugin container (empty when
// there is none), whether it has pods, and its allow-list entry.
type instanceState struct {
	Ready, Approval, Signature string
	Pending                    *v1alpha1.Permissions
	Deployed, Deployment       string
	Pods                       bool
	AllowList                  string
}

// stateOf reads p's instanceState: its status, as the operator wrote it
// after it created or removed the plugin's objects, and those objects.
func stateOf(t *testing.T, k *kube, p *v1alpha1.Plugin) instanceState {
	t.Helper()
	ctx := t.Context()
	cond := func(c string) string {
		if got := meta.FindStatusCondition(p.Status.Conditions, c); got != nil {
			return string(got.Status) + "/" + got.Reason
		}
		return "missing"
	}
	s := instanceState{
		Ready: cond(v1alpha1.ConditionReady), Approval: cond(v1alpha1.ConditionApprovalRequired),
		Signature: cond(v1alpha1.ConditionSignatureVerified), Pending: p.Status.PendingApproval,
	}
	if p.Status.Deployed != nil {
		s.Deployed = p.Status.Deployed.Image
	}
	d, err := k.clients.AppsV1().Deployments(namespace).Get(ctx, plugins.ServiceAccountName(p.Name), metav1.GetOptions{})
	switch {
	case err == nil:
		for i := range d.Spec.Template.Spec.Containers {
			if c := &d.Spec.Template.Spec.Containers[i]; c.Name == operator.ContainerName {
				s.Deployment = c.Image
			}
		}
	case !apierrors.IsNotFound(err):
		t.Fatalf("getting plugin %s's Deployment: %v", p.Name, err)
	}
	pods, err := k.clients.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: operator.LabelPlugin + "=" + p.Name})
	if err != nil {
		t.Fatalf("listing plugin %s's pods: %v", p.Name, err)
	}
	s.Pods = len(pods.Items) > 0
	cm, err := k.clients.CoreV1().ConfigMaps(namespace).Get(ctx, plugins.EgressAllowListConfigMap, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting the allow-list: %v", err)
	}
	s.AllowList = cm.Data[p.Name]
	return s
}

// waitCondition waits until Plugin instance's condition cond, for its
// current generation, has status and reason, and returns the Plugin. It
// fails at once when a pod of the instance can't start.
func waitCondition(t *testing.T, k *kube, instance, cond string, status metav1.ConditionStatus, reason string) *v1alpha1.Plugin {
	t.Helper()
	var got *v1alpha1.Plugin
	err := waitFor(t.Context(), pluginTimeout, func(ctx context.Context) (string, bool, error) {
		p, err := k.getPlugin(ctx, instance)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		c := meta.FindStatusCondition(p.Status.Conditions, cond)
		if c != nil && c.ObservedGeneration == p.Generation && c.Status == status && c.Reason == reason {
			got = p
			return conditionsState(p), true, nil
		}
		pods, err := k.pluginPodsState(ctx, instance)
		return conditionsState(p) + "; " + pods, false, err
	})
	if err != nil {
		t.Fatalf("waiting for plugin %s's %s to be %s/%s: %v", instance, cond, status, reason, err)
	}
	return got
}

// deleteInstance deletes Plugin instance, its PluginGrant and its
// credentials Secret, whichever exist; the cluster's garbage collector
// removes what the operator made for it.
func deleteInstance(t *testing.T, k *kube, instance string) {
	t.Helper()
	ctx := context.WithoutCancel(t.Context())
	errs := []error{
		k.dynamic.Resource(pluginResource()).Namespace(namespace).Delete(ctx, instance, metav1.DeleteOptions{}),
		k.dynamic.Resource(grantResource()).Namespace(namespace).Delete(ctx, instance, metav1.DeleteOptions{}),
		k.clients.CoreV1().Secrets(namespace).Delete(ctx, credentialSecret(instance), metav1.DeleteOptions{}),
	}
	for _, err := range errs {
		if err != nil && !apierrors.IsNotFound(err) {
			t.Errorf("removing instance %s: %v", instance, err)
		}
	}
}

// createCredential creates instance's credentials Secret, as the owner
// creates it (D-29), with a synthetic credential made for this run, which
// it returns. It is never logged.
func createCredential(t *testing.T, k *kube, instance string) string {
	t.Helper()
	credential := syntheticCredential(t)
	s := &corev1.Secret{
		// Labelled with its instance, as the owner labels it (D-182): uninstall
		// deletes only the Secrets that label names as the instance's own.
		ObjectMeta: metav1.ObjectMeta{
			Name:      credentialSecret(instance),
			Namespace: namespace,
			Labels:    map[string]string{plugins.OwnerInstanceLabel: instance},
		},
		StringData: map[string]string{connectorKey: credential},
	}
	if _, err := k.clients.CoreV1().Secrets(namespace).Create(t.Context(), s, metav1.CreateOptions{FieldManager: fieldManager}); err != nil {
		t.Fatalf("creating Secret %s: %v", s.Name, err)
	}
	return credential
}

// connectorManifests are the synthetic connector's two manifests, as a
// Plugin's spec.manifest: its first image's, and its second's, which asks
// for an egress host more.
type connectorManifests struct {
	first, second map[string]any
}

// readConnectorManifests reads the synthetic connector's manifests, the
// labels its images carry.
func readConnectorManifests(t *testing.T) connectorManifests {
	t.Helper()
	first, firstErr := os.ReadFile(connectorManifest)
	second, secondErr := os.ReadFile(connectorEgressManifest)
	if err := errors.Join(firstErr, secondErr); err != nil {
		t.Fatalf("reading the connector's manifests: %v", err)
	}
	var m connectorManifests
	if err := errors.Join(json.Unmarshal(first, &m.first), json.Unmarshal(second, &m.second)); err != nil {
		t.Fatalf("decoding the connector's manifests: %v", err)
	}
	return m
}
