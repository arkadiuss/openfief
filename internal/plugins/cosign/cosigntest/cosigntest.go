// Package cosigntest builds signed plugin images for tests of code that
// checks them with package cosign: an in-memory OCI registry served over
// HTTP, images carrying a manifest label, and cosign bundles signed with a
// fresh key and attached as OCI referrers, the way `cosign sign --key`
// stores them. Everything it makes is synthetic. Key also signs images in a
// real registry, such as the end-to-end suite's (test/e2e).
package cosigntest

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	protocommon "github.com/sigstore/protobuf-specs/gen/pb-go/common/v1"
	protodsse "github.com/sigstore/protobuf-specs/gen/pb-go/dsse"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/cosign"
)

const (
	// bundleMediaType is the media type cosign v3 gives a signature bundle.
	bundleMediaType = "application/vnd.dev.sigstore.bundle.v0.3+json"
	// emptyConfigMediaType is an OCI artifact manifest's config.
	emptyConfigMediaType = "application/vnd.oci.empty.v1+json"
	// ociSchemaVersion is the schemaVersion of every OCI manifest.
	ociSchemaVersion = 2
	// inTotoPayloadType is the DSSE payload type of an in-toto statement.
	inTotoPayloadType = "application/vnd.in-toto+json"
)

// Registry is an in-memory OCI registry with the Referrers API, served over
// plain HTTP on a loopback address for the length of a test.
type Registry struct {
	// Host is the registry's host and port.
	Host string
}

// NewRegistry serves a registry until the test ends.
func NewRegistry(t *testing.T) *Registry {
	t.Helper()
	srv := httptest.NewServer(registry.New(registry.WithReferrersSupport(true), registry.Logger(log.New(io.Discard, "", 0))))
	t.Cleanup(srv.Close)
	return &Registry{Host: strings.TrimPrefix(srv.URL, "http://")}
}

// Push writes a one-layer image to repository repo whose config carries
// label as its plugin manifest label, and returns its digest-pinned
// reference.
func (r *Registry) Push(t *testing.T, repo, label string) plugins.ImageRef {
	t.Helper()
	img, err := mutate.Config(empty.Image, v1.Config{Labels: map[string]string{plugins.ManifestLabel: label}})
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	img, err = mutate.AppendLayers(img, static.NewLayer([]byte("synthetic layer "+label), types.DockerLayer))
	if err != nil {
		t.Fatalf("layer: %v", err)
	}
	d, err := img.Digest()
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	ref := imageRef(t, r.Host+"/"+repo+"@"+d.String())
	if err := remote.Write(digest(t, ref.String()), img); err != nil {
		t.Fatalf("push image: %v", err)
	}
	return ref
}

// SignWithKey signs image with a fresh Key, attaching the bundle as a
// referrer, and returns the key's public half in PEM form.
func (r *Registry) SignWithKey(t *testing.T, image plugins.ImageRef) string {
	t.Helper()
	key, err := NewKey()
	if err != nil {
		t.Fatal(err)
	}
	if signErr := key.Sign(image); signErr != nil {
		t.Fatal(signErr)
	}
	pub, err := key.PublicKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
	return pub
}

// Key is a fresh ECDSA P-256 signing key, like one `cosign
// generate-key-pair` makes, that exists only in memory: nothing writes it
// anywhere, so it signs only during the run that made it. It signs with the
// standard library rather than sigstore-go's sign package, which would put
// golang.org/x/crypto/openpgp into every binary that links this package.
type Key struct {
	private *ecdsa.PrivateKey
}

// NewKey generates a Key.
func NewKey() (*Key, error) {
	private, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("keypair: %w", err)
	}
	return &Key{private: private}, nil
}

// PublicKeyPEM returns the key's public half as one PEM PUBLIC KEY block,
// the form a trusted signer's publicKey takes (plugins.ParseOwnerSigners).
func (k *Key) PublicKeyPEM() (string, error) {
	der, err := x509.MarshalPKIXPublicKey(&k.private.PublicKey)
	if err != nil {
		return "", fmt.Errorf("public key: %w", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})), nil
}

// Sign signs image as `cosign sign --key` does without a transparency log,
// a v0.3 bundle holding a DSSE envelope around the in-toto statement, and
// attaches the bundle as a referrer of image in its registry. The options
// set how the registry is reached.
func (k *Key) Sign(image plugins.ImageRef, opts ...remote.Option) error {
	der, err := x509.MarshalPKIXPublicKey(&k.private.PublicKey)
	if err != nil {
		return fmt.Errorf("public key: %w", err)
	}
	payload, err := statement(image)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(pae(inTotoPayloadType, payload))
	sig, err := ecdsa.SignASN1(rand.Reader, k.private, sum[:])
	if err != nil {
		return fmt.Errorf("sign: %w", err)
	}
	keyHash := sha256.Sum256(der)
	hint := base64.StdEncoding.EncodeToString(keyHash[:])
	data, err := protojson.Marshal(&protobundle.Bundle{
		MediaType: bundleMediaType,
		VerificationMaterial: &protobundle.VerificationMaterial{
			Content: &protobundle.VerificationMaterial_PublicKey{PublicKey: &protocommon.PublicKeyIdentifier{Hint: hint}},
		},
		Content: &protobundle.Bundle_DsseEnvelope{DsseEnvelope: &protodsse.Envelope{
			Payload: payload, PayloadType: inTotoPayloadType,
			Signatures: []*protodsse.Signature{{Sig: sig}},
		}},
	})
	if err != nil {
		return fmt.Errorf("marshal bundle: %w", err)
	}
	return attach(image, data, opts)
}

// pae is DSSE's pre-authentication encoding of a payload, what a DSSE
// signature signs.
func pae(payloadType string, payload []byte) []byte {
	return []byte(fmt.Sprintf("DSSEv1 %d %s %d %s", len(payloadType), payloadType, len(payload), payload))
}

// statement is the in-toto statement `cosign sign` signs for an image.
func statement(image plugins.ImageRef) ([]byte, error) {
	b, err := json.Marshal(map[string]any{
		"_type":         "https://in-toto.io/Statement/v1",
		"subject":       []any{map[string]any{"digest": map[string]string{"sha256": strings.TrimPrefix(image.Digest(), "sha256:")}}},
		"predicateType": cosign.SignPredicateType,
		"predicate":     map[string]any{},
	})
	if err != nil {
		return nil, fmt.Errorf("statement: %w", err)
	}
	return b, nil
}

// rawManifest is a manifest pushed exactly as written.
type rawManifest struct{ raw []byte }

func (m rawManifest) RawManifest() ([]byte, error)        { return m.raw, nil }
func (m rawManifest) MediaType() (types.MediaType, error) { return types.OCIManifestSchema1, nil }

// attach stores a bundle as a referrer of image, as cosign v3 does: an OCI
// manifest with an empty config, the bundle as its one layer and the image
// as its subject. A registry without the Referrers API gets the referrers
// tag instead.
func attach(image plugins.ImageRef, bundleJSON []byte, opts []remote.Option) error {
	d, err := name.NewDigest(image.String())
	if err != nil {
		return fmt.Errorf("digest reference %q: %w", image, err)
	}
	subject, err := remote.Head(d, opts...)
	if err != nil {
		return fmt.Errorf("head subject: %w", err)
	}
	cfg := static.NewLayer([]byte("{}"), emptyConfigMediaType)
	layer := static.NewLayer(bundleJSON, bundleMediaType)
	for _, l := range []v1.Layer{cfg, layer} {
		if writeErr := remote.WriteLayer(d.Context(), l, opts...); writeErr != nil {
			return fmt.Errorf("push blob: %w", writeErr)
		}
	}
	cfgDesc, err := descriptor(cfg)
	if err != nil {
		return err
	}
	layerDesc, err := descriptor(layer)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(v1.Manifest{
		SchemaVersion: ociSchemaVersion, MediaType: types.OCIManifestSchema1, ArtifactType: bundleMediaType,
		Config:  cfgDesc,
		Layers:  []v1.Descriptor{layerDesc},
		Subject: &v1.Descriptor{MediaType: subject.MediaType, Digest: subject.Digest, Size: subject.Size},
	})
	if err != nil {
		return fmt.Errorf("marshal referrer: %w", err)
	}
	h, _, err := v1.SHA256(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("referrer digest: %w", err)
	}
	if err := remote.Put(d.Context().Digest(h.String()), rawManifest{raw}, opts...); err != nil {
		return fmt.Errorf("push referrer: %w", err)
	}
	return nil
}

// descriptor describes a layer.
func descriptor(l v1.Layer) (v1.Descriptor, error) {
	d, err := l.Digest()
	if err != nil {
		return v1.Descriptor{}, fmt.Errorf("layer digest: %w", err)
	}
	size, err := l.Size()
	if err != nil {
		return v1.Descriptor{}, fmt.Errorf("layer size: %w", err)
	}
	mt, err := l.MediaType()
	if err != nil {
		return v1.Descriptor{}, fmt.Errorf("layer media type: %w", err)
	}
	return v1.Descriptor{MediaType: mt, Digest: d, Size: size}, nil
}

func digest(t *testing.T, ref string) name.Digest {
	t.Helper()
	d, err := name.NewDigest(ref)
	if err != nil {
		t.Fatalf("digest reference %q: %v", ref, err)
	}
	return d
}

func imageRef(t *testing.T, ref string) plugins.ImageRef {
	t.Helper()
	img, err := plugins.ParseImageRef(ref)
	if err != nil {
		t.Fatalf("image reference %q: %v", ref, err)
	}
	return img
}
