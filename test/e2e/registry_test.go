//go:build e2e

package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/cosign/cosigntest"
)

// The suite's registry (scripts/e2e-images.sh): task e2e:images pushes the
// plugin images there under pluginTag, and task e2e passes its addresses.
const (
	// registryEnv is where the suite reaches the registry, on loopback.
	registryEnv = "OPENFIEF_E2E_REGISTRY"
	// clusterRegistryEnv is where the cluster reaches it: a private IPv4
	// address on the cluster's Docker network, which the operator's registry
	// client reads over plain HTTP.
	clusterRegistryEnv = "OPENFIEF_E2E_CLUSTER_REGISTRY"
	// pluginTag is the tag of every plugin image in the registry.
	pluginTag = "e2e"
)

// The suite's trusted signer (D-45, D-57).
const (
	// signerName names the suite's key among the owner's trusted signers;
	// a Plugin's status.signature shows it.
	signerName = "e2e-test-key"
	// trustedSignersConfigMap and trustedSignersKey are where deploy/operator
	// reads the owner's trusted signers from.
	trustedSignersConfigMap = "openfief-trusted-signers"
	trustedSignersKey       = "signers.json"
	// operatorDeployment is the operator's Deployment (deploy/operator).
	operatorDeployment = "openfief-operator"
	// signersAnnotation on the operator's pod template is the SHA-256 of the
	// trusted signers, so a run with a new key restarts an operator that a
	// run before it started: it reads them only when it starts.
	signersAnnotation = "openfief.io/e2e-trusted-signers-sha256"
)

var (
	// errNoRegistry reports a run without the registry's addresses.
	errNoRegistry = errors.New(registryEnv + " and " + clusterRegistryEnv + " must be set; run task e2e:images, then task e2e")
	// errClusterRegistry reports a cluster address the operator would read
	// only over HTTPS.
	errClusterRegistry = errors.New(clusterRegistryEnv + " must be a private IPv4 address and port")
)

// pluginImages are the plugin images task e2e:images pushes to the
// registry, which the suite signs: the NBP FX connector and the built-in UI,
// which the operator deploys for the demo's ui instance (editApp).
func pluginImages() []string { return []string{"nbp-fx", uiImage} }

// pluginResource is the operator's Plugin resource.
func pluginResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "openfief.io", Version: "v1alpha1", Resource: "plugins"}
}

// grantResource is the operator's PluginGrant resource.
func grantResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "openfief.io", Version: "v1alpha1", Resource: "plugingrants"}
}

// suiteRegistry is the suite's registry, as the suite and as the cluster
// reach it: two addresses of the same registry.
type suiteRegistry struct {
	local, cluster string
}

// registryFromEnv returns the registry task e2e passes.
func registryFromEnv() (suiteRegistry, error) {
	r := suiteRegistry{local: os.Getenv(registryEnv), cluster: os.Getenv(clusterRegistryEnv)}
	if r.local == "" || r.cluster == "" {
		return suiteRegistry{}, errNoRegistry
	}
	host, _, err := net.SplitHostPort(r.cluster)
	if err != nil {
		return suiteRegistry{}, fmt.Errorf("%w: %q: %w", errClusterRegistry, r.cluster, err)
	}
	if ip := net.ParseIP(host); ip == nil || ip.To4() == nil || !ip.IsPrivate() {
		return suiteRegistry{}, fmt.Errorf("%w: %q", errClusterRegistry, r.cluster)
	}
	return r, nil
}

// registryImage is one image in the suite's registry, pinned by digest.
type registryImage struct {
	// local is the image as the suite reaches it, cluster as the cluster
	// names it.
	local, cluster plugins.ImageRef
}

// image returns plugin image name, pinned to the digest the registry holds
// for its tag.
func (r suiteRegistry) image(ctx context.Context, image string) (registryImage, error) {
	tag, err := name.NewTag(r.local + "/" + image + ":" + pluginTag)
	if err != nil {
		return registryImage{}, fmt.Errorf("image %s: %w", image, err)
	}
	desc, err := remote.Head(tag, remote.WithContext(ctx))
	if err != nil {
		return registryImage{}, fmt.Errorf("image %s in the suite's registry (did task e2e:images run?): %w", tag, err)
	}
	local, err := plugins.ParseImageRef(r.local + "/" + image + "@" + desc.Digest.String())
	if err != nil {
		return registryImage{}, fmt.Errorf("image %s: %w", image, err)
	}
	cluster, err := plugins.ParseImageRef(r.cluster + "/" + image + "@" + desc.Digest.String())
	if err != nil {
		return registryImage{}, fmt.Errorf("image %s: %w", image, err)
	}
	return registryImage{local: local, cluster: cluster}, nil
}

// trustSigner signs every plugin image in the registry with a key made for
// this run, the way an owner signs a plugin with a local key (D-45, D-57),
// and adds the key's public half to the owner's trusted signers. The private
// key never leaves this process. It returns the trusted signers' SHA-256.
func trustSigner(ctx context.Context, k *kube, reg suiteRegistry, logf func(string, ...any)) (string, error) {
	key, err := cosigntest.NewKey()
	if err != nil {
		return "", fmt.Errorf("signing key: %w", err)
	}
	for _, image := range pluginImages() {
		img, imageErr := reg.image(ctx, image)
		if imageErr != nil {
			return "", imageErr
		}
		logf("e2e: signing %s with this run's key", img.local)
		if signErr := key.Sign(img.local, remote.WithContext(ctx)); signErr != nil {
			return "", fmt.Errorf("signing %s: %w", img.local, signErr)
		}
	}
	pub, err := key.PublicKeyPEM()
	if err != nil {
		return "", fmt.Errorf("signing key: %w", err)
	}
	signers, err := json.Marshal(map[string]any{
		"signers": []map[string]string{{"name": signerName, "publicKey": pub}},
	})
	if err != nil {
		return "", fmt.Errorf("trusted signers: %w", err)
	}
	if _, parseErr := plugins.ParseOwnerSigners(signers); parseErr != nil {
		return "", fmt.Errorf("trusted signers: %w", parseErr)
	}
	logf("e2e: adding the key as trusted signer %s (ConfigMap %s)", signerName, trustedSignersConfigMap)
	cm := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "v1", "kind": "ConfigMap",
		"metadata": map[string]any{"name": trustedSignersConfigMap, "namespace": namespace},
		"data":     map[string]any{trustedSignersKey: string(signers)},
	}}
	if err := k.applyOne(ctx, cm); err != nil {
		return "", err
	}
	sum := sha256.Sum256(signers)
	return hex.EncodeToString(sum[:]), nil
}

// editApp returns the changes the suite makes to the demo's application
// layer: the operator's pod template gets the trusted signers' hash, and the
// demo's Plugin ui the built-in UI's image in the suite's registry, which
// the suite signed, as the cluster names it.
func editApp(signersHash string, ui plugins.ImageRef) func(*unstructured.Unstructured) error {
	return func(obj *unstructured.Unstructured) error {
		switch {
		case obj.GetKind() == "Deployment" && obj.GetName() == operatorDeployment:
			err := unstructured.SetNestedField(obj.Object, signersHash, "spec", "template", "metadata", "annotations", signersAnnotation)
			if err != nil {
				return fmt.Errorf("operator: %w", err)
			}
		case obj.GetKind() == "Plugin" && obj.GetName() == uiPlugin:
			if err := unstructured.SetNestedField(obj.Object, ui.String(), "spec", "image"); err != nil {
				return fmt.Errorf("plugin %s: %w", uiPlugin, err)
			}
		}
		return nil
	}
}
