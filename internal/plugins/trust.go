package plugins

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Errors of the trust policy (D-45, D-57). An image refused with any of them
// is never deployed, and nothing overrides that.
var (
	// ErrUnsigned: no valid signature was found for the image.
	ErrUnsigned = errors.New("image has no valid signature")
	// ErrUntrustedSigner: the image's valid signatures are all by signers the
	// trust policy does not hold.
	ErrUntrustedSigner = errors.New("image is not signed by a trusted signer")
	// ErrManifestMismatch: the manifest the Plugin declares differs from the
	// signed manifest label of its image (D-42).
	ErrManifestMismatch = errors.New("declared manifest differs from the image's signed manifest label")
	// ErrManifestLabel: the image has no readable ManifestLabel, so what it
	// asks for is unknown (D-42).
	ErrManifestLabel = errors.New("image has no valid manifest label")
	// ErrInvalidSigner reports a malformed trusted signer or signer identity.
	ErrInvalidSigner = errors.New("invalid signer")
)

// GitHubActionsIssuer is the OIDC issuer that Fulcio writes into the
// certificate of a keyless signature made in GitHub Actions.
const GitHubActionsIssuer = "https://token.actions.githubusercontent.com"

// OpenFiefCISubject is the certificate subject of OpenFief's CI signature:
// the reusable images.yml workflow, which ci.yml and release.yml call and
// which pushes and signs images only from main (docs/dev/sdlc.md). Fulcio
// takes the subject of a reusable workflow's certificate from its
// job_workflow_ref, so this names the signing workflow and its ref exactly.
//
// The subject alone does not name OpenFief's CI: images.yml is a reusable
// workflow in a public repository, so any repository's workflow can call it
// and get a certificate with this very subject for a job that builds its own
// code. OpenFiefCI therefore also pins the repository and ref the calling
// run came from, OpenFiefCISourceRepository and OpenFiefCISourceRef.
const OpenFiefCISubject = "https://github.com/arkadiuss/openfief/.github/workflows/images.yml@refs/heads/main"

// OpenFiefCISourceRepository is the repository whose run OpenFief's CI
// signature must come from: Fulcio's source repository URI extension
// (OID 1.3.6.1.4.1.57264.1.12), which for GitHub Actions names the
// repository of the workflow run, the caller of a reusable workflow.
const OpenFiefCISourceRepository = "https://github.com/arkadiuss/openfief"

// OpenFiefCISourceRef is the ref OpenFief's CI signature must come from:
// Fulcio's source repository ref extension (OID 1.3.6.1.4.1.57264.1.14),
// the git ref of the workflow run.
const OpenFiefCISourceRef = "refs/heads/main"

// SignerKind is one of D-45's three kinds of trusted signer.
type SignerKind string

// The kinds of trusted signer (D-45).
const (
	// SignerOpenFiefCI is OpenFief's own CI.
	SignerOpenFiefCI SignerKind = "OpenFiefCI"
	// SignerPublisher is a publisher pinned in the catalog (D-44).
	SignerPublisher SignerKind = "Publisher"
	// SignerOwner is a signer the owner added.
	SignerOwner SignerKind = "Owner"
)

// SignerIdentity is who made a signature. A keyless identity is the OIDC
// issuer and subject that Fulcio certified, the subject given exactly or as
// an RE2 pattern that must match all of it, and optionally the Source the
// signing run must come from; a key identity is a public key in PEM form.
// Exactly one of the two is set.
//
// The same type describes what a verified signature was made by: then the
// subject is the certificate's exact subject and Source the repository and
// ref the certificate names (nil when it names neither), or the key is the
// one that verified the signature.
type SignerIdentity struct {
	Issuer         string
	Subject        string
	SubjectPattern string
	// Source pins the repository and ref of the signing run; nil pins
	// neither. OpenFief's CI always pins both.
	Source    *SignerSource
	PublicKey string
}

// SignerSource is the repository and git ref a keyless signature's run came
// from, as Fulcio certifies them in its source repository URI and source
// repository ref extensions. They matter for a subject that names a reusable
// workflow: Fulcio's subject is then the called workflow, which any
// repository may call, while the source names the caller's run. A field left
// empty in a signer's pin matches any value.
type SignerSource struct {
	// Repository is the https URL of the repository, such as
	// https://github.com/example/plugin.
	Repository string
	// Ref is the git ref, such as refs/heads/main.
	Ref string
}

// Validate reports a malformed identity: neither or both of a keyless
// identity and a public key, a keyless identity without an https issuer, both
// or neither of a subject and a subject pattern, a pattern that does not
// compile, a GitHub-issued identity that doesn't pin both its source
// repository and ref (D-187), a malformed source, or a public key that is
// not one PEM "PUBLIC KEY" block.
func (id SignerIdentity) Validate() error {
	keyless := id.Issuer != "" || id.Subject != "" || id.SubjectPattern != "" || id.Source != nil
	switch {
	case keyless == (id.PublicKey != ""):
		return fmt.Errorf("%w: give either an issuer and subject or a public key", ErrInvalidSigner)
	case keyless:
		return id.validateKeyless()
	}
	_, err := publicKeyDER(id.PublicKey)
	return err
}

// IsKey reports whether id is a public key rather than a keyless identity.
func (id SignerIdentity) IsKey() bool { return id.PublicKey != "" }

// Matches reports whether a signature made by signed counts as one by id:
// the same issuer, a subject equal to id's or matching all of its pattern,
// and a source that id's source pin admits; or the same public key. A
// malformed id matches nothing.
func (id SignerIdentity) Matches(signed SignerIdentity) bool {
	if id.Validate() != nil {
		return false
	}
	if id.IsKey() {
		return samePublicKey(id.PublicKey, signed.PublicKey)
	}
	if signed.IsKey() || signed.Issuer != id.Issuer || signed.Subject == "" || !id.Source.admits(signed.Source) {
		return false
	}
	if id.Subject != "" {
		return signed.Subject == id.Subject
	}
	re, err := regexp.Compile(anchored(id.SubjectPattern))
	return err == nil && re.MatchString(signed.Subject)
}

// String names the identity in messages: the issuer and subject, with the
// source when set, or the public key's SHA-256 fingerprint.
func (id SignerIdentity) String() string {
	if id.IsKey() {
		der, err := publicKeyDER(id.PublicKey)
		if err != nil {
			return "invalid public key"
		}
		sum := sha256.Sum256(der)
		return "public key sha256:" + hex.EncodeToString(sum[:])
	}
	subject := id.Subject
	if subject == "" {
		subject = "~" + id.SubjectPattern
	}
	return keylessName(subject, id.Issuer, id.Source)
}

// keylessName names a keyless signer in messages: its subject and issuer,
// then the source repository and ref, each only when set, so a reusable
// workflow's signature from another repository reads differently from the
// same workflow's signature from its own.
func keylessName(subject, issuer string, source *SignerSource) string {
	s := subject + " (issuer " + issuer
	if source != nil && source.Repository != "" {
		s += ", source " + source.Repository
	}
	if source != nil && source.Ref != "" {
		s += ", ref " + source.Ref
	}
	return s + ")"
}

// validateKeyless checks the issuer, subject, pattern and source of a
// keyless identity. One GitHub issues names a workflow that any repository
// may call, so it must pin the source repository and ref of the signing
// run, as OpenFief's CI does (D-181, D-187).
func (id SignerIdentity) validateKeyless() error {
	if !isHTTPSURL(id.Issuer) {
		return fmt.Errorf("%w: issuer %q is not an https URL", ErrInvalidSigner, id.Issuer)
	}
	if (id.Subject == "") == (id.SubjectPattern == "") {
		return fmt.Errorf("%w: give exactly one of subject and subjectPattern", ErrInvalidSigner)
	}
	if id.SubjectPattern != "" {
		if _, err := regexp.Compile(anchored(id.SubjectPattern)); err != nil {
			return fmt.Errorf("%w: subjectPattern %q: %w", ErrInvalidSigner, id.SubjectPattern, err)
		}
	}
	if id.Issuer == GitHubActionsIssuer && !id.Source.pinsBoth() {
		return fmt.Errorf("%w: a GitHub Actions identity needs sourceRepository and sourceRef (D-187)", ErrInvalidSigner)
	}
	if id.Source != nil {
		return id.Source.validate()
	}
	return nil
}

// validate reports a source pin that pins nothing, a repository that is not
// an https URL, or a ref that is not a git ref under refs/.
func (s *SignerSource) validate() error {
	switch {
	case s.Repository == "" && s.Ref == "":
		return fmt.Errorf("%w: a source needs sourceRepository or sourceRef", ErrInvalidSigner)
	case s.Repository != "" && !isHTTPSURL(s.Repository):
		return fmt.Errorf("%w: sourceRepository %q is not an https URL", ErrInvalidSigner, s.Repository)
	case s.Ref != "" && (!strings.HasPrefix(s.Ref, "refs/") || strings.ContainsAny(s.Ref, " \t\n")):
		return fmt.Errorf("%w: sourceRef %q is not a git ref such as refs/heads/main", ErrInvalidSigner, s.Ref)
	}
	return nil
}

// admits reports whether a signature whose run came from signed counts for
// the pin s: no pin admits any source, and a pin needs each field it sets
// to be equal.
func (s *SignerSource) admits(signed *SignerSource) bool {
	if s == nil {
		return true
	}
	if signed == nil {
		return false
	}
	return (s.Repository == "" || s.Repository == signed.Repository) && (s.Ref == "" || s.Ref == signed.Ref)
}

// pinsBoth reports whether s pins both a repository and a ref.
func (s *SignerSource) pinsBoth() bool {
	return s != nil && s.Repository != "" && s.Ref != ""
}

// isHTTPSURL reports whether s is https:// and a host, with no whitespace.
func isHTTPSURL(s string) bool {
	host, isHTTPS := strings.CutPrefix(s, "https://")
	return isHTTPS && host != "" && !strings.ContainsAny(host, " \t\n")
}

// samePublicKey reports whether two PEM public keys are the same key.
func samePublicKey(a, b string) bool {
	want, wantErr := publicKeyDER(a)
	got, gotErr := publicKeyDER(b)
	return wantErr == nil && gotErr == nil && bytes.Equal(want, got)
}

// anchored makes an RE2 pattern match whole strings only.
func anchored(pattern string) string { return `^(?:` + pattern + `)$` }

// publicKeyDER returns the DER bytes of a PEM public key, which must be one
// "PUBLIC KEY" block and nothing else.
func publicKeyDER(s string) ([]byte, error) {
	block, rest := pem.Decode([]byte(s))
	if block == nil || block.Type != "PUBLIC KEY" || strings.TrimSpace(string(rest)) != "" {
		return nil, fmt.Errorf("%w: publicKey must be one PEM PUBLIC KEY block", ErrInvalidSigner)
	}
	return block.Bytes, nil
}

// Signer is a trusted signer: a name for messages, its kind and its identity.
type Signer struct {
	Name     string
	Kind     SignerKind
	Identity SignerIdentity
}

// Validate reports a signer without a name or kind, with a malformed
// identity, or of kind SignerOpenFiefCI without a keyless identity that pins
// both the source repository and the source ref: its subject names a
// reusable workflow anyone can call, so only those name OpenFief's CI.
func (s *Signer) Validate() error {
	switch {
	case s.Name == "":
		return fmt.Errorf("%w: a signer needs a name", ErrInvalidSigner)
	case s.Kind != SignerOpenFiefCI && s.Kind != SignerPublisher && s.Kind != SignerOwner:
		return fmt.Errorf("%w: signer %q: unknown kind %q", ErrInvalidSigner, s.Name, s.Kind)
	case s.Kind == SignerOpenFiefCI && !s.Identity.Source.pinsBoth():
		return fmt.Errorf("%w: signer %q: OpenFief's CI needs a keyless identity with sourceRepository and sourceRef", ErrInvalidSigner, s.Name)
	}
	if err := s.Identity.Validate(); err != nil {
		return fmt.Errorf("signer %q: %w", s.Name, err)
	}
	return nil
}

// OpenFiefCI is OpenFief's CI as a trusted signer: keyless, GitHub Actions,
// exactly OpenFiefCISubject, from a run of OpenFiefCISourceRepository at
// OpenFiefCISourceRef.
func OpenFiefCI() Signer {
	return Signer{
		Name: "openfief-ci", Kind: SignerOpenFiefCI,
		Identity: SignerIdentity{
			Issuer: GitHubActionsIssuer, Subject: OpenFiefCISubject,
			Source: &SignerSource{Repository: OpenFiefCISourceRepository, Ref: OpenFiefCISourceRef},
		},
	}
}

// Trust is how far an image is trusted once a trusted signer signed it
// (D-56).
type Trust string

// The trust levels of a deployable image.
const (
	// Verified: the catalog pins the image's digest and its publisher signed
	// it (D-44).
	Verified Trust = "Verified"
	// Unverified: a trusted signer signed it, but the catalog does not pin
	// it to that signer (D-43, D-45).
	Unverified Trust = "Unverified"
)

// SignedImage is what the registry holds for an image, checked
// cryptographically by an adapter: the identities whose signatures over the
// image digest verified, and the manifest from the image's signed manifest
// label (D-42). Untrusted names the signers of the image's other signatures,
// those no trusted signer made (D-169); they never make an image deployable,
// and only say why one is refused.
type SignedImage struct {
	Image      ImageRef
	Signatures []SignerIdentity
	Manifest   Manifest
	Untrusted  []UntrustedSigner
}

// UntrustedSigner is who made a signature over an image when the trust
// policy holds no such signer (D-169). A keyless signature names the issuer
// and subject of a certificate that verified against the Sigstore trust
// root, and the source the certificate names, if any; a
// signature made with a key names the SHA-256 fingerprint of that key's DER
// form, as the signature's key hint gives it, which nothing can check
// without the key itself. Exactly one of the two is set.
type UntrustedSigner struct {
	Issuer    string
	Subject   string
	Source    *SignerSource
	KeySHA256 string
}

// String names the signer as SignerIdentity.String names a trusted one, so
// the owner can compare it with their trusted signers: the subject and
// issuer with the source repository and ref, or "public key sha256:" and
// the key's fingerprint.
func (u UntrustedSigner) String() string {
	if u.KeySHA256 != "" {
		return "public key sha256:" + u.KeySHA256
	}
	return keylessName(u.Subject, u.Issuer, u.Source)
}

// UntrustedSignerError is ErrUntrustedSigner for one image: the signers of
// its valid signatures, none of whom the trust policy holds (D-169).
type UntrustedSignerError struct {
	Image   ImageRef
	Signers []string
}

// Error names the image and its signers.
func (e *UntrustedSignerError) Error() string {
	return ErrUntrustedSigner.Error() + ": " + e.Image.String() + " is signed by " + strings.Join(e.Signers, ", ")
}

// Is makes the error match ErrUntrustedSigner.
func (e *UntrustedSignerError) Is(target error) bool { return target == ErrUntrustedSigner }

// TrustDecision is why an image may be deployed: its trust level and the
// signer that earned it.
type TrustDecision struct {
	Trust  Trust
	Signer Signer
}

// TrustPolicy is every signer the operator trusts (D-45): OpenFief's CI, each
// catalog publisher and each signer the owner added, with the catalog that
// decides which images are verified. There is no way to trust an unsigned
// image (D-57).
type TrustPolicy struct {
	signers []Signer
	catalog Catalog
}

// NewTrustPolicy builds the policy from the catalog and the owner's signers.
// OpenFief's CI is always trusted. It fails on an invalid catalog, an invalid
// owner signer, or an owner signer whose kind is not SignerOwner.
//
// A publisher who signs several catalog entries, or an owner signer listed
// twice, is trusted once, where it first appears: two signers are the same
// when their name, kind and identity are all equal (sameSigner). Signers that
// share only a name stay apart, because trust follows the identity and the
// name is only for messages; Catalog.Validate doesn't require names to be
// unique, so a publisher who signs with two identities keeps both. An owner
// signer with a publisher's identity stays apart too: its kind differs.
func NewTrustPolicy(catalog Catalog, owner []Signer) (TrustPolicy, error) {
	p := TrustPolicy{signers: []Signer{OpenFiefCI()}, catalog: catalog}
	for i := range catalog.Entries {
		p.signers = appendSigner(p.signers, &catalog.Entries[i].Publisher)
	}
	errs := []error{catalog.Validate()}
	for i := range owner {
		s := owner[i]
		if s.Kind != SignerOwner {
			errs = append(errs, fmt.Errorf("%w: owner signer %q has kind %q", ErrInvalidSigner, s.Name, s.Kind))
			continue
		}
		if err := s.Validate(); err != nil {
			errs = append(errs, err)
			continue
		}
		p.signers = appendSigner(p.signers, &s)
	}
	if err := errors.Join(errs...); err != nil {
		return TrustPolicy{}, err
	}
	return p, nil
}

// Signers returns every trusted signer: OpenFief's CI first, then the
// catalog's publishers, then the owner's signers.
func (p *TrustPolicy) Signers() []Signer { return append([]Signer(nil), p.signers...) }

// Decide checks a signed image against the policy and, when declared is not
// nil, the Plugin's declared manifest against the image's signed manifest
// label (D-42). The image is Verified when a catalog entry pins its digest
// and that entry's publisher signed it, Unverified when any other trusted
// signer signed it. It fails with ErrUnsigned when the image has no valid
// signature at all, with an *UntrustedSignerError naming the signers when
// its only valid signatures are by signers the policy doesn't hold (D-169),
// or with ErrManifestMismatch.
func (p *TrustPolicy) Decide(img *SignedImage, declared *Manifest) (TrustDecision, error) {
	if len(img.Signatures) == 0 {
		if len(img.Untrusted) == 0 {
			return TrustDecision{}, fmt.Errorf("%w: %s", ErrUnsigned, img.Image)
		}
		names := make([]string, 0, len(img.Untrusted))
		for _, u := range img.Untrusted {
			names = appendNew(names, u.String())
		}
		return TrustDecision{}, &UntrustedSignerError{Image: img.Image, Signers: names}
	}
	d, ok := p.trusted(img)
	if !ok {
		names := make([]string, 0, len(img.Signatures))
		for _, s := range img.Signatures {
			names = appendNew(names, s.String())
		}
		return TrustDecision{}, &UntrustedSignerError{Image: img.Image, Signers: names}
	}
	if declared != nil && !declared.Equal(&img.Manifest) {
		return TrustDecision{}, fmt.Errorf("%w: %s", ErrManifestMismatch, img.Image)
	}
	return d, nil
}

// trusted finds the best trusted signature on img: a catalog pin first,
// then the first trusted signer in policy order.
func (p *TrustPolicy) trusted(img *SignedImage) (TrustDecision, bool) {
	for i := range p.catalog.Entries {
		e := &p.catalog.Entries[i]
		if e.Image.Digest() == img.Image.Digest() && signedBy(img, e.Publisher.Identity) {
			return TrustDecision{Trust: Verified, Signer: e.Publisher}, true
		}
	}
	for _, s := range p.signers {
		if signedBy(img, s.Identity) {
			return TrustDecision{Trust: Unverified, Signer: s}, true
		}
	}
	return TrustDecision{}, false
}

// appendNew appends s to names unless names already holds it.
func appendNew(names []string, s string) []string {
	if slices.Contains(names, s) {
		return names
	}
	return append(names, s)
}

// appendSigner appends s to signers unless signers already holds the same
// signer.
func appendSigner(signers []Signer, s *Signer) []Signer {
	for i := range signers {
		if sameSigner(&signers[i], s) {
			return signers
		}
	}
	return append(signers, *s)
}

// sameSigner reports whether a and b have the same name, kind and identity,
// comparing the identity's source by value rather than by pointer.
func sameSigner(a, b *Signer) bool {
	x, y := &a.Identity, &b.Identity
	return a.Name == b.Name && a.Kind == b.Kind &&
		x.Issuer == y.Issuer && x.Subject == y.Subject && x.SubjectPattern == y.SubjectPattern &&
		x.PublicKey == y.PublicKey && sameSource(x.Source, y.Source)
}

// sameSource reports whether a and b are both nil or pin the same repository
// and ref.
func sameSource(a, b *SignerSource) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// signedBy reports whether any of img's signatures was made by id.
func signedBy(img *SignedImage, id SignerIdentity) bool {
	return slices.ContainsFunc(img.Signatures, id.Matches)
}
