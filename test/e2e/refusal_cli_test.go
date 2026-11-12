//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// The instances and images TestCLIRefusalAgreesWithOperator uses, each its
// own, so the subtests run side by side and apart from
// TestOperatorRefusesUntrustedImages.
const (
	cliRefusedUnsigned  = "cli-refused-unsigned"
	cliRefusedUntrusted = "cli-refused-untrusted"
	cliRefusedTagged    = "cli-refused-tagged"
	cliRefusedTagDigest = "cli-refused-tagdigest"
)

// TestCLIRefusalAgreesWithOperator checks that `openfief plugin install`,
// the CLI built from this checkout, previews the check the operator makes
// itself (D-56), so an owner who applies a Plugin with plain kubectl gets
// the verdict the CLI would have given. For each image the operator must
// never deploy, the CLI exits non-zero, applies nothing, no Plugin and no
// PluginGrant, and prints the refusal the operator then records when the
// same image is applied directly: an image nobody signed (Unsigned), one
// signed by a key the owner doesn't trust (UntrustedSigner, naming the
// key), one named by tag alone, which the API server refuses as invalid
// while the CLI says the image isn't pinned by digest, and one that names a
// tag next to its digest (InvalidImage). For the first, second and fourth,
// the kubectl-applied Plugin exists, and its status says why it isn't
// deployed (D-200 criterion 11, SCEN-e2e-cli-refusal-matches-operator).
func TestCLIRefusalAgreesWithOperator(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	if _, err := os.Stat(cliPath); err != nil {
		t.Fatalf("%v: %v", errNoCLI, err)
	}
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("unsigned image", func(t *testing.T) {
		t.Parallel()
		img := pushUnsignedImage(t, reg, cliRefusedUnsigned)
		checkCLIAgreesWithOperator(t, k, cliRefusedUnsigned, img.cluster.String(), refusedAs{
			reason: operator.ReasonUnsigned, err: plugins.ErrUnsigned,
		})
	})
	t.Run("image signed by an untrusted key", func(t *testing.T) {
		t.Parallel()
		img := pushUnsignedImage(t, reg, cliRefusedUntrusted)
		signer := signUntrusted(t, img.local)
		checkCLIAgreesWithOperator(t, k, cliRefusedUntrusted, img.cluster.String(), refusedAs{
			reason: operator.ReasonUntrustedSigner, err: plugins.ErrUntrustedSigner, signers: []string{signer},
		})
	})
	t.Run("image named by tag alone", func(t *testing.T) {
		t.Parallel()
		tagged := reg.cluster + "/" + fxImage + ":" + pluginTag
		t.Cleanup(func() { deleteInstance(t, k, cliRefusedTagged) })
		out := mustRefuseWithCLI(t, k, cliRefusedTagged, tagged)
		if !bytes.Contains(out, []byte(plugins.ErrImageNotPinned.Error())) {
			t.Errorf("openfief plugin install %s printed:\n%s\nwant the refusal %q", tagged, out, plugins.ErrImageNotPinned)
		}
		// The API server refuses the same Plugin as invalid, so it never exists
		// for the operator to judge: both refuse, and neither leaves a Plugin.
		err := k.applyOne(t.Context(), fxPluginObject(t, cliRefusedTagged, tagged))
		if !apierrors.IsInvalid(err) {
			t.Fatalf("applying a Plugin with image %s: %v; want the API server to refuse it as invalid", tagged, err)
		}
		checkNothingApplied(t, k, cliRefusedTagged)
	})
	t.Run("image naming a tag next to its digest", func(t *testing.T) {
		t.Parallel()
		img, err := reg.image(t.Context(), fxImage)
		if err != nil {
			t.Fatal(err)
		}
		both := reg.cluster + "/" + fxImage + ":" + pluginTag + "@" + img.cluster.Digest()
		checkCLIAgreesWithOperator(t, k, cliRefusedTagDigest, both, refusedAs{
			reason: operator.ReasonInvalidImage, err: plugins.ErrInvalidImageRef,
		})
	})
}

// checkCLIAgreesWithOperator installs image as instance with the CLI, which
// must refuse it and apply nothing, then applies the same image directly,
// with a grant that approves everything requested, and checks that the
// operator's Plugin exists, is not Ready, deploys nothing and records the
// refusal as want says, the one the CLI printed.
func checkCLIAgreesWithOperator(t *testing.T, k *kube, instance, image string, want refusedAs) {
	t.Helper()
	t.Cleanup(func() { deleteInstance(t, k, instance) })
	out := mustRefuseWithCLI(t, k, instance, image)
	if !namesRefusal(string(out), want) {
		t.Errorf("openfief plugin install %s printed:\n%s\nwant the refusal %q and its untrusted signers %v", image, out, want.err, want.signers)
	}

	if err := k.applyOne(t.Context(), fxGrant(instance)); err != nil {
		t.Fatal(err)
	}
	if err := k.applyOne(t.Context(), fxPluginObject(t, instance, image)); err != nil {
		t.Fatalf("applying Plugin %s with plain kubectl: %v; want the API server to take it, the operator to judge it", instance, err)
	}
	checkOperatorRefusal(t, k, instance, want)
}

// checkOperatorRefusal waits until the operator has decided Plugin
// instance's signature and checks that it records the refusal as want says,
// is not Ready and deployed nothing.
func checkOperatorRefusal(t *testing.T, k *kube, instance string, want refusedAs) {
	t.Helper()
	p := waitSignatureDecided(t, k, instance)
	sig := meta.FindStatusCondition(p.Status.Conditions, v1alpha1.ConditionSignatureVerified)
	if sig.Reason != want.reason {
		t.Errorf("plugin %s: SignatureVerified reason %q, want %q like the CLI's refusal", instance, sig.Reason, want.reason)
	}
	if !namesRefusal(sig.Message, want) {
		t.Errorf("plugin %s: SignatureVerified message %q does not give the refusal %q the CLI printed, naming %v", instance, sig.Message, want.err, want.signers)
	}
	if ready := meta.FindStatusCondition(p.Status.Conditions, v1alpha1.ConditionReady); ready == nil || ready.Status != metav1.ConditionFalse {
		t.Errorf("plugin %s: Ready condition %+v, want False", instance, ready)
	}
	if p.Status.Deployed != nil {
		t.Errorf("plugin %s records a deployed version %+v; want none", instance, p.Status.Deployed)
	}
	if left, err := k.instanceDeployed(t.Context(), instance); err != nil || len(left) > 0 {
		t.Errorf("plugin %s deployed something (%v): %v; want nothing", instance, err, left)
	}
}

// mustRefuseWithCLI runs `openfief plugin install <image> --name <instance>
// --yes` and checks that the CLI exits non-zero, having been run, and that
// it applied no Plugin and no PluginGrant. It returns the CLI's output.
func mustRefuseWithCLI(t *testing.T, k *kube, instance, image string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), cliTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, cliPath)
	// The arguments are the suite's own: an image it pushed, a name and a flag.
	cmd.Args = append(cmd.Args, "plugin", "install", image, "--name", instance, "--yes")
	out, err := cmd.CombinedOutput()
	if exitErr, ok := errors.AsType[*exec.ExitError](err); !ok || exitErr.ExitCode() == 0 {
		t.Fatalf("openfief plugin install %s: %v\n%s\nwant it to run and exit non-zero", image, err, out)
	}
	checkNothingApplied(t, k, instance)
	return out
}

// checkNothingApplied checks that neither Plugin nor PluginGrant instance
// exists.
func checkNothingApplied(t *testing.T, k *kube, instance string) {
	t.Helper()
	if _, err := k.getPlugin(t.Context(), instance); !apierrors.IsNotFound(err) {
		t.Errorf("plugin %s: %v; want not found", instance, err)
	}
	_, err := k.dynamic.Resource(grantResource()).Namespace(namespace).Get(t.Context(), instance, metav1.GetOptions{})
	if !apierrors.IsNotFound(err) {
		t.Errorf("plugin grant %s: %v; want not found", instance, err)
	}
}

// instanceDeployed names the Deployment and pods the operator made for
// instance, if any.
func (k *kube) instanceDeployed(ctx context.Context, instance string) ([]string, error) {
	var found []string
	_, err := k.clients.AppsV1().Deployments(namespace).Get(ctx, plugins.ServiceAccountName(instance), metav1.GetOptions{})
	switch {
	case err == nil:
		found = append(found, "Deployment")
	case !apierrors.IsNotFound(err):
		return nil, fmt.Errorf("getting the Deployment: %w", err)
	}
	pods, err := k.clients.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: operator.LabelPlugin + "=" + instance})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}
	for i := range pods.Items {
		found = append(found, "pod "+pods.Items[i].Name)
	}
	return found, nil
}
