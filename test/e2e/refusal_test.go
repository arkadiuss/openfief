//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/cosign"
	"github.com/arkadiuss/openfief/internal/plugins/cosign/cosigntest"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// The instances TestOperatorRefusesUntrustedImages applies, each named like
// the image it runs in the suite's registry.
const (
	// refusedUnsigned runs an image nobody signed.
	refusedUnsigned = "refused-unsigned"
	// refusedUntrusted runs an image signed by a key made for the test that
	// is not among the owner's trusted signers.
	refusedUntrusted = "refused-untrusted"
	// untrustedSignerName names that key in the suite's own check that its
	// signature is valid; the operator never hears of it.
	untrustedSignerName = "e2e-untrusted-key"
	// refusalTimeout bounds the wait for the operator to decide a Plugin's
	// signature, and for its event.
	refusalTimeout = 3 * time.Minute
)

// errVerifiedAnyway reports a signature the operator accepted although no
// trusted signer made it: waiting won't change that.
var errVerifiedAnyway = errors.New("signature verified, want refused")

// TestOperatorRefusesUntrustedImages applies Plugins the way an owner does,
// with a PluginGrant that approves everything the manifest requests, for two
// images the operator must never deploy (D-57): one nobody signed, and one
// signed by a key the owner does not trust, whose signature the suite first
// checks is valid and covers the image's manifest label. For both the
// operator records SignatureVerified False with reason Unsigned, since it
// counts only signatures that verify as a trusted signer's (D-45, D-56),
// and Ready False with reason SignatureNotVerified; it creates no
// Deployment and no pod, publishes no allow-list entry, records nothing as
// deployed, and records the refusal as a Warning event on the Plugin
// (SCEN-e2e-plugin-unsigned-refused, SCEN-e2e-plugin-untrusted-signer-refused).
// TestOperatorDeploysSignedPlugin checks that the same operator deploys an
// image the trusted signer signed.
func TestOperatorRefusesUntrustedImages(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("unsigned image", func(t *testing.T) {
		t.Parallel()
		img := pushUnsignedImage(t, reg, refusedUnsigned)
		checkRefused(t, k, refusedUnsigned, img.cluster)
	})
	t.Run("image signed by an untrusted key", func(t *testing.T) {
		t.Parallel()
		img := pushUnsignedImage(t, reg, refusedUntrusted)
		signUntrusted(t, img.local)
		checkRefused(t, k, refusedUntrusted, img.cluster)
	})
}

// fxManifestLabel is the NBP FX connector's manifest as an image's manifest
// label (D-42).
func fxManifestLabel(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(fxManifest)
	if err != nil {
		t.Fatalf("reading the manifest: %v", err)
	}
	var label bytes.Buffer
	if err := json.Compact(&label, data); err != nil {
		t.Fatalf("compacting the manifest: %v", err)
	}
	return label.String()
}

// pushUnsignedImage pushes a synthetic image to repository repo of the
// suite's registry, under the suite's tag: one layer naming repo, so every
// repository gets its own digest, and the NBP FX connector's manifest
// label, so only its signature can make the operator refuse it. Nothing
// signs it.
func pushUnsignedImage(t *testing.T, reg suiteRegistry, repo string) registryImage {
	t.Helper()
	img, err := mutate.Config(empty.Image, v1.Config{Labels: map[string]string{plugins.ManifestLabel: fxManifestLabel(t)}})
	if err != nil {
		t.Fatalf("image config: %v", err)
	}
	img, err = mutate.AppendLayers(img, static.NewLayer([]byte("openfief e2e synthetic layer "+repo), types.DockerLayer))
	if err != nil {
		t.Fatalf("image layer: %v", err)
	}
	tag, err := name.NewTag(reg.local + "/" + repo + ":" + pluginTag)
	if err != nil {
		t.Fatalf("image %s: %v", repo, err)
	}
	if writeErr := remote.Write(tag, img, remote.WithContext(t.Context())); writeErr != nil {
		t.Fatalf("pushing %s: %v", tag, writeErr)
	}
	pinned, err := reg.image(t.Context(), repo)
	if err != nil {
		t.Fatal(err)
	}
	return pinned
}

// signUntrusted signs image with a key made for this test only, which the
// owner's trusted signers don't hold, and checks that the signature is
// valid, so the operator must refuse the image for its signer alone. The
// private key never leaves this process.
func signUntrusted(t *testing.T, image plugins.ImageRef) {
	t.Helper()
	key, err := cosigntest.NewKey()
	if err != nil {
		t.Fatalf("untrusted key: %v", err)
	}
	t.Logf("e2e: signing %s with a key the owner does not trust", image)
	if signErr := key.Sign(image, remote.WithContext(t.Context())); signErr != nil {
		t.Fatalf("signing %s: %v", image, signErr)
	}
	pub, err := key.PublicKeyPEM()
	if err != nil {
		t.Fatalf("untrusted key: %v", err)
	}
	checkSignedBy(t, image, pub)
}

// checkSignedBy checks, with the operator's own signature check trusting
// only the key whose public half is pub, that image carries a valid
// signature by that key over its digest and its manifest label.
func checkSignedBy(t *testing.T, image plugins.ImageRef, pub string) {
	t.Helper()
	policy, err := plugins.NewTrustPolicy(plugins.Catalog{}, []plugins.Signer{{
		Name: untrustedSignerName, Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{PublicKey: pub},
	}})
	if err != nil {
		t.Fatalf("policy trusting the untrusted key: %v", err)
	}
	trustRoot, err := cosign.TrustedRoot()
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := cosign.New(&policy, trustRoot)
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	signed, err := verifier.Inspect(t.Context(), image)
	if err != nil {
		t.Fatalf("inspecting %s: %v", image, err)
	}
	declared, err := plugins.ParseManifestLabel(fxManifestLabel(t))
	if err != nil {
		t.Fatalf("manifest label: %v", err)
	}
	d, err := policy.Decide(&signed, &declared)
	if err != nil {
		t.Fatalf("the untrusted key's signature on %s does not verify: %v", image, err)
	}
	if d.Signer.Name != untrustedSignerName {
		t.Fatalf("%s verified as signed by %q; want %q", image, d.Signer.Name, untrustedSignerName)
	}
}

// refusal is what checkRefused compares of a refused Plugin and the
// cluster around it.
type refusal struct {
	// Ready, Approval and Signature are the Plugin's conditions as
	// "<status>/<reason>".
	Ready, Approval, Signature string
	// SignatureMessageNamesRefusal is whether the SignatureVerified
	// message says the image has no valid signature.
	SignatureMessageNamesRefusal bool
	// Deployed, Effective and Signer are whether the status records a
	// deployed version, effective permissions and a signer.
	Deployed, Effective, Signer bool
	// Deployment is whether the plugin's Deployment exists; Pods counts its
	// pods; AllowList is whether it has an allow-list entry.
	Deployment bool
	Pods       int
	AllowList  bool
}

// checkRefused applies instance running image with a grant that
// approves everything it requests, waits until the operator has decided its
// signature for the current generation, checks that it refused the image
// and runs nothing, waits for the refusal's Warning event, and removes the
// instance again.
func checkRefused(t *testing.T, k *kube, instance string, image plugins.ImageRef) {
	t.Helper()
	installFX(t, k, instance, image)
	p := waitSignatureDecided(t, k, instance)
	want := refusal{
		Ready:     string(metav1.ConditionFalse) + "/" + operator.ReasonSignatureNotVerified,
		Approval:  string(metav1.ConditionFalse) + "/" + operator.ReasonApproved,
		Signature: string(metav1.ConditionFalse) + "/" + operator.ReasonUnsigned, SignatureMessageNamesRefusal: true,
	}
	if diff := cmp.Diff(want, refusalOf(t, k, p)); diff != "" {
		t.Errorf("plugin %s (-want +got):\n%s", instance, diff)
	}
	waitRefusalEvent(t, k, instance)
	removeInstance(t, k, instance)
}

// waitSignatureDecided waits until Plugin instance's SignatureVerified condition
// is False for its current generation, failing at once when it is True, and
// returns the Plugin.
func waitSignatureDecided(t *testing.T, k *kube, instance string) *v1alpha1.Plugin {
	t.Helper()
	var decided *v1alpha1.Plugin
	err := waitFor(t.Context(), refusalTimeout, func(ctx context.Context) (string, bool, error) {
		p, err := k.getPlugin(ctx, instance)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		sig := meta.FindStatusCondition(p.Status.Conditions, v1alpha1.ConditionSignatureVerified)
		if sig == nil || sig.ObservedGeneration != p.Generation {
			return conditionsState(p), false, nil
		}
		switch sig.Status {
		case metav1.ConditionTrue:
			return "", false, fmt.Errorf("%w: %s: %s", errVerifiedAnyway, sig.Reason, sig.Message)
		case metav1.ConditionFalse:
			decided = p
			return conditionsState(p), true, nil
		case metav1.ConditionUnknown:
		}
		return conditionsState(p), false, nil
	})
	if err != nil {
		t.Fatalf("waiting for the operator to decide plugin %s's signature: %v", instance, err)
	}
	return decided
}

// refusalOf reads what checkRefused compares: the Plugin's status, its
// Deployment, its pods and its allow-list entry. The operator writes the
// status after it created or removed the plugin's objects, so they are
// what that decision left.
func refusalOf(t *testing.T, k *kube, p *v1alpha1.Plugin) refusal {
	t.Helper()
	ctx := t.Context()
	state := func(cond string) string {
		if c := meta.FindStatusCondition(p.Status.Conditions, cond); c != nil {
			return string(c.Status) + "/" + c.Reason
		}
		return "missing"
	}
	r := refusal{
		Ready: state(v1alpha1.ConditionReady), Approval: state(v1alpha1.ConditionApprovalRequired),
		Signature: state(v1alpha1.ConditionSignatureVerified),
		Deployed:  p.Status.Deployed != nil, Effective: p.Status.Effective != nil, Signer: p.Status.Signature != nil,
	}
	if sig := meta.FindStatusCondition(p.Status.Conditions, v1alpha1.ConditionSignatureVerified); sig != nil {
		r.SignatureMessageNamesRefusal = strings.Contains(sig.Message, plugins.ErrUnsigned.Error())
	}
	_, err := k.clients.AppsV1().Deployments(namespace).Get(ctx, plugins.ServiceAccountName(p.Name), metav1.GetOptions{})
	switch {
	case err == nil:
		r.Deployment = true
	case !apierrors.IsNotFound(err):
		t.Fatalf("getting plugin %s's Deployment: %v", p.Name, err)
	}
	pods, err := k.clients.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: operator.LabelPlugin + "=" + p.Name})
	if err != nil {
		t.Fatalf("listing plugin %s's pods: %v", p.Name, err)
	}
	r.Pods = len(pods.Items)
	cm, err := k.clients.CoreV1().ConfigMaps(namespace).Get(ctx, plugins.EgressAllowListConfigMap, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting the allow-list: %v", err)
	}
	_, r.AllowList = cm.Data[p.Name]
	return r
}

// waitRefusalEvent waits for the Warning event the operator records on
// Plugin instance when it refuses its signature.
func waitRefusalEvent(t *testing.T, k *kube, instance string) {
	t.Helper()
	selector := fields.Set{
		"involvedObject.kind": "Plugin", "involvedObject.name": instance,
		"type": corev1.EventTypeWarning, "reason": operator.ReasonUnsigned,
	}.String()
	err := waitFor(t.Context(), refusalTimeout, func(ctx context.Context) (string, bool, error) {
		events, err := k.clients.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{FieldSelector: selector})
		if err != nil {
			return "", false, fmt.Errorf("%w: listing events: %w", errRetry, err)
		}
		for i := range events.Items {
			if strings.Contains(events.Items[i].Message, plugins.ErrUnsigned.Error()) {
				return events.Items[i].Message, true, nil
			}
		}
		return fmt.Sprintf("%d matching events, none naming the refusal", len(events.Items)), false, nil
	})
	if err != nil {
		t.Fatalf("waiting for the refusal event of plugin %s (%s): %v", instance, selector, err)
	}
}

// removeInstance deletes Plugin instance and its PluginGrant; the cluster's
// garbage collector removes what the operator made for it.
func removeInstance(t *testing.T, k *kube, instance string) {
	t.Helper()
	ctx := t.Context()
	if err := k.dynamic.Resource(pluginResource()).Namespace(namespace).Delete(ctx, instance, metav1.DeleteOptions{}); err != nil {
		t.Errorf("deleting plugin %s: %v", instance, err)
	}
	if err := k.dynamic.Resource(grantResource()).Namespace(namespace).Delete(ctx, instance, metav1.DeleteOptions{}); err != nil {
		t.Errorf("deleting plugin grant %s: %v", instance, err)
	}
}
