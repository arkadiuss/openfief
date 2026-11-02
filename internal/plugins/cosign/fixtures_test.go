package cosign

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

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
	protorekor "github.com/sigstore/protobuf-specs/gen/pb-go/rekor/v1"
	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/fulcio/certificate"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/sign"
	"github.com/sigstore/sigstore-go/pkg/testing/ca"
	"github.com/sigstore/sigstore-go/pkg/tlog"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// bundleMediaType is the media type cosign v3 gives a signature bundle.
const bundleMediaType = "application/vnd.dev.sigstore.bundle.v0.3+json"

// Synthetic plugin manifests, as image labels.
const (
	labelView      = `{"capabilities":["view"]}`
	labelConnector = `{"capabilities":["connector"],"egress":["api.example.com"]}`
)

// testRegistry is an in-memory OCI registry served over HTTP.
type testRegistry struct {
	host string
}

// newTestRegistry serves an in-memory registry, with or without the
// Referrers API (GHCR has none, so cosign uses the referrers tag). wrap, if
// not nil, sits in front of the registry.
func newTestRegistry(t *testing.T, referrersAPI bool, wrap func(http.Handler) http.Handler) *testRegistry {
	t.Helper()
	h := registry.New(registry.WithReferrersSupport(referrersAPI))
	if wrap != nil {
		h = wrap(h)
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &testRegistry{host: strings.TrimPrefix(srv.URL, "http://")}
}

// testImage is an image built in a test.
type testImage struct{ v1.Image }

// labelledImage is a one-layer image whose config carries label as its
// manifest label; an empty label means none.
func labelledImage(t *testing.T, label string) testImage {
	t.Helper()
	cfg := v1.Config{Labels: map[string]string{"org.opencontainers.image.title": "sample"}}
	if label != "" {
		cfg.Labels[plugins.ManifestLabel] = label
	}
	img, err := mutate.Config(empty.Image, cfg)
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	img, err = mutate.AppendLayers(img, static.NewLayer([]byte("synthetic layer "+label), types.DockerLayer))
	if err != nil {
		t.Fatalf("layer: %v", err)
	}
	return testImage{img}
}

// push writes img to the registry and returns its digest-pinned reference.
func (r *testRegistry) push(t *testing.T, img testImage) plugins.ImageRef {
	t.Helper()
	d, err := img.Digest()
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	ref := r.host + "/openfief/sample@" + d.String()
	if err := remote.Write(mustDigest(t, ref), img.Image); err != nil {
		t.Fatalf("push image: %v", err)
	}
	return mustImageRef(t, ref)
}

// pushIndex writes a multi-platform index of imgs, plus a BuildKit-style
// attestation manifest, and returns its reference.
func (r *testRegistry) pushIndex(t *testing.T, imgs ...testImage) plugins.ImageRef {
	t.Helper()
	var idx v1.ImageIndex = empty.Index
	arch := []string{"amd64", "arm64"}
	for i, img := range imgs {
		idx = mutate.AppendManifests(idx, mutate.IndexAddendum{
			Add: img.Image, Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: "linux", Architecture: arch[i%len(arch)]}},
		})
	}
	attestation := labelledImage(t, "")
	idx = mutate.AppendManifests(idx, mutate.IndexAddendum{Add: attestation.Image, Descriptor: v1.Descriptor{
		Platform:    &v1.Platform{OS: "unknown", Architecture: "unknown"},
		Annotations: map[string]string{"vnd.docker.reference.type": attestationReference},
	}})
	d, err := idx.Digest()
	if err != nil {
		t.Fatalf("index digest: %v", err)
	}
	ref := r.host + "/openfief/sample@" + d.String()
	if err := remote.WriteIndex(mustDigest(t, ref), idx); err != nil {
		t.Fatalf("push index: %v", err)
	}
	return mustImageRef(t, ref)
}

// rawManifest is a manifest pushed exactly as written.
type rawManifest struct{ raw []byte }

func (m rawManifest) RawManifest() ([]byte, error)        { return m.raw, nil }
func (m rawManifest) MediaType() (types.MediaType, error) { return types.OCIManifestSchema1, nil }

// attach stores a bundle as a referrer of image, the way cosign v3 does: an
// OCI manifest with an empty config, the bundle as its one layer, the bundle
// media type as its artifact type and the image as its subject.
func (r *testRegistry) attach(t *testing.T, image plugins.ImageRef, bundleJSON []byte) {
	t.Helper()
	subject, err := remote.Head(mustDigest(t, image.String()))
	if err != nil {
		t.Fatalf("head subject: %v", err)
	}
	repo := mustDigest(t, image.String()).Context()
	cfg := static.NewLayer([]byte("{}"), emptyConfigMediaType)
	layer := static.NewLayer(bundleJSON, bundleMediaType)
	for _, l := range []v1.Layer{cfg, layer} {
		if writeErr := remote.WriteLayer(repo, l); writeErr != nil {
			t.Fatalf("push blob: %v", writeErr)
		}
	}
	m := v1.Manifest{
		SchemaVersion: 2, MediaType: types.OCIManifestSchema1, ArtifactType: bundleMediaType,
		Config:  descriptorOf(t, cfg),
		Layers:  []v1.Descriptor{descriptorOf(t, layer)},
		Subject: &v1.Descriptor{MediaType: subject.MediaType, Digest: subject.Digest, Size: subject.Size},
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal referrer: %v", err)
	}
	h, _, err := v1.SHA256(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("referrer digest: %v", err)
	}
	if err := remote.Put(repo.Digest(h.String()), rawManifest{raw}); err != nil {
		t.Fatalf("push referrer: %v", err)
	}
}

func descriptorOf(t *testing.T, l v1.Layer) v1.Descriptor {
	t.Helper()
	d, err := l.Digest()
	if err != nil {
		t.Fatalf("layer digest: %v", err)
	}
	size, err := l.Size()
	if err != nil {
		t.Fatalf("layer size: %v", err)
	}
	mt, err := l.MediaType()
	if err != nil {
		t.Fatalf("layer media type: %v", err)
	}
	return v1.Descriptor{MediaType: mt, Digest: d, Size: size}
}

func mustDigest(t *testing.T, ref string) name.Digest {
	t.Helper()
	d, err := name.NewDigest(ref)
	if err != nil {
		t.Fatalf("digest reference %q: %v", ref, err)
	}
	return d
}

func mustImageRef(t *testing.T, ref string) plugins.ImageRef {
	t.Helper()
	img, err := plugins.ParseImageRef(ref)
	if err != nil {
		t.Fatalf("image reference %q: %v", ref, err)
	}
	return img
}

// statement is the in-toto statement `cosign sign` signs for an image.
func statement(t *testing.T, image plugins.ImageRef, predicateType string) []byte {
	t.Helper()
	s := map[string]any{
		"_type":         "https://in-toto.io/Statement/v1",
		"subject":       []any{map[string]any{"digest": map[string]string{"sha256": strings.TrimPrefix(image.Digest(), sha256Prefix)}}},
		"predicateType": predicateType,
		"predicate":     map[string]any{},
	}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("statement: %v", err)
	}
	return b
}

// keylessBundle signs stmt in the virtual Sigstore as subject from issuer,
// logged at integratedTime, and returns the v0.3 bundle cosign would store.
func keylessBundle(t *testing.T, vs *ca.VirtualSigstore, subject, issuer string, stmt []byte, integratedTime time.Time) []byte {
	t.Helper()
	entity, err := vs.AttestAtTime(subject, issuer, stmt, integratedTime, true)
	if err != nil {
		t.Fatalf("attest: %v", err)
	}
	vc, err := entity.VerificationContent()
	if err != nil {
		t.Fatalf("verification content: %v", err)
	}
	sc, err := entity.SignatureContent()
	if err != nil {
		t.Fatalf("signature content: %v", err)
	}
	env, ok := sc.(*bundle.Envelope)
	if !ok {
		t.Fatalf("signature content is %T, want a DSSE envelope", sc)
	}
	entries, err := entity.TlogEntries()
	if err != nil {
		t.Fatalf("tlog entries: %v", err)
	}
	tsrs, err := entity.Timestamps()
	if err != nil {
		t.Fatalf("timestamps: %v", err)
	}
	pb := &protobundle.Bundle{
		MediaType: bundleMediaType,
		VerificationMaterial: &protobundle.VerificationMaterial{
			Content:                   &protobundle.VerificationMaterial_Certificate{Certificate: &protocommon.X509Certificate{RawBytes: vc.Certificate().Raw}},
			TlogEntries:               []*protorekor.TransparencyLogEntry{loggedEntry(t, vs, entries[0])},
			TimestampVerificationData: &protobundle.TimestampVerificationData{},
		},
		Content: &protobundle.Bundle_DsseEnvelope{DsseEnvelope: dsseProto(t, env)},
	}
	for _, tsr := range tsrs {
		pb.VerificationMaterial.TimestampVerificationData.Rfc3161Timestamps = append(
			pb.VerificationMaterial.TimestampVerificationData.Rfc3161Timestamps, &protocommon.RFC3161SignedTimestamp{SignedTimestamp: tsr},
		)
	}
	return marshalBundle(t, pb)
}

// loggedEntry is a virtual Rekor entry as a bundle carries it: with its
// kind, inclusion proof and signed entry timestamp (inclusion promise), so
// its integrated time counts as an observer timestamp.
func loggedEntry(t *testing.T, vs *ca.VirtualSigstore, e *tlog.Entry) *protorekor.TransparencyLogEntry {
	t.Helper()
	tle := e.TransparencyLogEntry()
	logID, err := vs.RekorLogID()
	if err != nil {
		t.Fatalf("log id: %v", err)
	}
	set, err := vs.RekorSignPayload(tlog.RekorPayload{
		Body:           base64.StdEncoding.EncodeToString(tle.GetCanonicalizedBody()),
		IntegratedTime: tle.GetIntegratedTime(),
		LogIndex:       tle.GetLogIndex(),
		LogID:          logID,
	})
	if err != nil {
		t.Fatalf("sign entry timestamp: %v", err)
	}
	tle.KindVersion = &protorekor.KindVersion{Kind: "dsse", Version: "0.0.1"}
	tle.InclusionPromise = &protorekor.InclusionPromise{SignedEntryTimestamp: set}
	return tle
}

func dsseProto(t *testing.T, env *bundle.Envelope) *protodsse.Envelope {
	t.Helper()
	raw := env.RawEnvelope()
	payload, err := base64.StdEncoding.DecodeString(raw.Payload)
	if err != nil {
		t.Fatalf("payload: %v", err)
	}
	out := &protodsse.Envelope{Payload: payload, PayloadType: raw.PayloadType}
	for _, s := range raw.Signatures {
		sig, err := base64.StdEncoding.DecodeString(s.Sig)
		if err != nil {
			t.Fatalf("signature: %v", err)
		}
		out.Signatures = append(out.Signatures, &protodsse.Signature{Sig: sig, Keyid: s.KeyID})
	}
	return out
}

// keySigned is a bundle signed with a key, the key in PEM form, the key
// hint the bundle carries, and the key's fingerprint (hex SHA-256 of its
// DER form), which the hint encodes.
type keySigned struct {
	bundle      []byte
	publicKey   string
	hint        []byte
	fingerprint string
}

// keyBundle signs stmt with a fresh key, as `cosign sign --key` does
// without a transparency log.
func keyBundle(t *testing.T, stmt []byte) keySigned {
	t.Helper()
	return signWithKey(t, stmt, nil)
}

// keyBundleWithHint signs stmt with a fresh key under the given key hint,
// as a forger claiming someone else's key would.
func keyBundleWithHint(t *testing.T, stmt, hint []byte) []byte {
	t.Helper()
	return signWithKey(t, stmt, &sign.EphemeralKeypairOptions{Hint: hint}).bundle
}

// signWithKey signs stmt with a fresh key made with opts.
func signWithKey(t *testing.T, stmt []byte, opts *sign.EphemeralKeypairOptions) keySigned {
	t.Helper()
	kp, err := sign.NewEphemeralKeypair(opts)
	if err != nil {
		t.Fatalf("keypair: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(kp.GetPublicKey())
	if err != nil {
		t.Fatalf("public key: %v", err)
	}
	sum := sha256.Sum256(der)
	pb, err := sign.Bundle(&sign.DSSEData{Data: stmt, PayloadType: "application/vnd.in-toto+json"}, kp, sign.BundleOptions{})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	pemKey, err := kp.GetPublicKeyPem()
	if err != nil {
		t.Fatalf("public key: %v", err)
	}
	return keySigned{bundle: marshalBundle(t, pb), publicKey: pemKey, hint: kp.GetHint(), fingerprint: hex.EncodeToString(sum[:])}
}

func marshalBundle(t *testing.T, pb *protobundle.Bundle) []byte {
	t.Helper()
	b, err := protojson.Marshal(pb)
	if err != nil {
		t.Fatalf("marshal bundle: %v", err)
	}
	return b
}

// tamperConfig rewrites every image config blob the registry serves: the
// first occurrence of from becomes to, the way an attacker who controls the
// registry would widen a signed manifest label.
func tamperConfig(from, to string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || !strings.Contains(r.URL.Path, "/blobs/") {
				next.ServeHTTP(w, r)
				return
			}
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)
			body := rec.Body.Bytes()
			if bytes.Contains(body, []byte(plugins.ManifestLabel)) {
				body = bytes.Replace(body, []byte(from), []byte(to), 1)
			}
			maps.Copy(w.Header(), rec.Header())
			w.Header().Del("Content-Length")
			w.WriteHeader(rec.Code)
			if _, err := w.Write(body); err != nil {
				return
			}
		})
	}
}

// githubCA is a test Fulcio CA whose certificates carry what Fulcio writes
// for a GitHub Actions run: the OIDC issuer and the source repository and
// ref of the run, which the virtual Sigstore's own CA can't add.
type githubCA struct {
	authority *root.FulcioCertificateAuthority
	key       *ecdsa.PrivateKey
}

func newGitHubCA(t *testing.T) *githubCA {
	t.Helper()
	rootCert, rootKey, err := ca.GenerateRootCa()
	if err != nil {
		t.Fatalf("root CA: %v", err)
	}
	intermediate, key, err := ca.GenerateFulcioIntermediate(rootCert, rootKey)
	if err != nil {
		t.Fatalf("intermediate CA: %v", err)
	}
	return &githubCA{authority: &root.FulcioCertificateAuthority{Root: rootCert, Intermediates: []*x509.Certificate{intermediate}}, key: key}
}

// withCA trusts what its TrustedMaterial trusts plus one more Fulcio CA.
type withCA struct {
	root.TrustedMaterial

	extra root.CertificateAuthority
}

func (w withCA) FulcioCertificateAuthorities() []root.CertificateAuthority {
	return append(slices.Clone(w.TrustedMaterial.FulcioCertificateAuthorities()), w.extra)
}

// githubRun is what a GitHub Actions run's certificate names: the subject
// (the workflow that ran, the called one for a reusable workflow), the
// issuer, and the repository and ref of the run (empty: no extension).
type githubRun struct {
	subject, issuer, repository, ref string
}

// openFiefCIRun is OpenFief's CI: its images workflow, run from its own
// repository's main branch.
func openFiefCIRun() githubRun {
	return githubRun{
		subject: plugins.OpenFiefCISubject, issuer: plugins.GitHubActionsIssuer,
		repository: plugins.OpenFiefCISourceRepository, ref: plugins.OpenFiefCISourceRef,
	}
}

// runCertificate is a sign.CertificateProvider that certifies a keypair as
// run, from gh.
type runCertificate struct {
	gh  *githubCA
	run githubRun
}

func (c runCertificate) GetCertificate(_ context.Context, kp sign.Keypair, _ *sign.CertificateProviderOptions) ([]byte, error) {
	san, err := url.Parse(c.run.subject)
	if err != nil {
		return nil, fmt.Errorf("subject: %w", err)
	}
	exts := []pkix.Extension{}
	for _, e := range []struct {
		oid   asn1.ObjectIdentifier
		value string
	}{
		{certificate.OIDIssuerV2, c.run.issuer},
		{certificate.OIDSourceRepositoryURI, c.run.repository},
		{certificate.OIDSourceRepositoryRef, c.run.ref},
	} {
		if e.value == "" {
			continue
		}
		value, marshalErr := asn1.MarshalWithParams(e.value, "utf8")
		if marshalErr != nil {
			return nil, fmt.Errorf("extension %s: %w", e.oid, marshalErr)
		}
		exts = append(exts, pkix.Extension{Id: e.oid, Value: value})
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:    big.NewInt(1),
		URIs:            []*url.URL{san},
		NotBefore:       now.Add(-time.Minute),
		NotAfter:        now.Add(10 * time.Minute),
		KeyUsage:        x509.KeyUsageDigitalSignature,
		ExtKeyUsage:     []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		ExtraExtensions: exts,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, c.gh.authority.Intermediates[0], kp.GetPublicKey(), c.gh.key)
	if err != nil {
		return nil, fmt.Errorf("leaf certificate: %w", err)
	}
	return der, nil
}

// githubBundle signs stmt as cosign does in the GitHub Actions run run,
// with a certificate from gh, logged and timestamped in vs at
// integratedTime, and returns the v0.3 bundle cosign would store.
func githubBundle(t *testing.T, vs *ca.VirtualSigstore, gh *githubCA, run githubRun, stmt []byte, integratedTime time.Time) []byte {
	t.Helper()
	kp, err := sign.NewEphemeralKeypair(nil)
	if err != nil {
		t.Fatalf("keypair: %v", err)
	}
	pb, err := sign.Bundle(&sign.DSSEData{Data: stmt, PayloadType: "application/vnd.in-toto+json"}, kp,
		sign.BundleOptions{CertificateProvider: runCertificate{gh: gh, run: run}})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	b, err := bundle.NewBundle(pb)
	if err != nil {
		t.Fatalf("bundle: %v", err)
	}
	env, err := b.Envelope()
	if err != nil {
		t.Fatalf("envelope: %v", err)
	}
	raw := env.RawEnvelope()
	sig, err := base64.StdEncoding.DecodeString(raw.Signatures[0].Sig)
	if err != nil {
		t.Fatalf("signature: %v", err)
	}
	leaf, err := x509.ParseCertificate(pb.GetVerificationMaterial().GetCertificate().GetRawBytes())
	if err != nil {
		t.Fatalf("leaf certificate: %v", err)
	}
	entry, err := vs.GenerateTlogEntry(leaf, raw, sig, integratedTime.Unix(), true)
	if err != nil {
		t.Fatalf("log entry: %v", err)
	}
	tsr, err := vs.TimestampResponse(sig)
	if err != nil {
		t.Fatalf("timestamp: %v", err)
	}
	pb.VerificationMaterial.TlogEntries = []*protorekor.TransparencyLogEntry{loggedEntry(t, vs, entry)}
	pb.VerificationMaterial.TimestampVerificationData = &protobundle.TimestampVerificationData{
		Rfc3161Timestamps: []*protocommon.RFC3161SignedTimestamp{{SignedTimestamp: tsr}},
	}
	return marshalBundle(t, pb)
}
