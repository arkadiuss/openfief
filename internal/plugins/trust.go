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
const OpenFiefCISubject = "https://github.com/arkadiuss/openfief/.github/workflows/images.yml@refs/heads/main"

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
// an RE2 pattern that must match all of it; a key identity is a public key
// in PEM form. Exactly one of the two is set.
//
// The same type describes what a verified signature was made by: then the
// subject is the certificate's exact subject, or the key is the one that
// verified the signature.
type SignerIdentity struct {
	Issuer         string
	Subject        string
	SubjectPattern string
	PublicKey      string
}

// Validate reports a malformed identity: neither or both of a keyless
// identity and a public key, a keyless identity without an https issuer, both
// or neither of a subject and a subject pattern, a pattern that does not
// compile, or a public key that is not one PEM "PUBLIC KEY" block.
func (id SignerIdentity) Validate() error {
	keyless := id.Issuer != "" || id.Subject != "" || id.SubjectPattern != ""
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
// the same issuer and a subject equal to id's or matching all of its
// pattern, or the same public key. A malformed id matches nothing.
func (id SignerIdentity) Matches(signed SignerIdentity) bool {
	if id.Validate() != nil {
		return false
	}
	if id.IsKey() {
		want, wantErr := publicKeyDER(id.PublicKey)
		got, gotErr := publicKeyDER(signed.PublicKey)
		return wantErr == nil && gotErr == nil && bytes.Equal(want, got)
	}
	if signed.IsKey() || signed.Issuer != id.Issuer || signed.Subject == "" {
		return false
	}
	if id.Subject != "" {
		return signed.Subject == id.Subject
	}
	re, err := regexp.Compile(anchored(id.SubjectPattern))
	return err == nil && re.MatchString(signed.Subject)
}

// String names the identity in messages: the issuer and subject, or the
// public key's SHA-256 fingerprint.
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
	return subject + " (issuer " + id.Issuer + ")"
}

// validateKeyless checks the issuer, subject and pattern of a keyless
// identity.
func (id SignerIdentity) validateKeyless() error {
	host, isHTTPS := strings.CutPrefix(id.Issuer, "https://")
	if !isHTTPS || host == "" || strings.ContainsAny(host, " \t\n") {
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
	return nil
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

// Validate reports a signer without a name or kind, or with a malformed
// identity.
func (s *Signer) Validate() error {
	switch {
	case s.Name == "":
		return fmt.Errorf("%w: a signer needs a name", ErrInvalidSigner)
	case s.Kind != SignerOpenFiefCI && s.Kind != SignerPublisher && s.Kind != SignerOwner:
		return fmt.Errorf("%w: signer %q: unknown kind %q", ErrInvalidSigner, s.Name, s.Kind)
	}
	if err := s.Identity.Validate(); err != nil {
		return fmt.Errorf("signer %q: %w", s.Name, err)
	}
	return nil
}

// OpenFiefCI is OpenFief's CI as a trusted signer: keyless, GitHub Actions,
// exactly OpenFiefCISubject.
func OpenFiefCI() Signer {
	return Signer{
		Name: "openfief-ci", Kind: SignerOpenFiefCI,
		Identity: SignerIdentity{Issuer: GitHubActionsIssuer, Subject: OpenFiefCISubject},
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
// label (D-42).
type SignedImage struct {
	Image      ImageRef
	Signatures []SignerIdentity
	Manifest   Manifest
}

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
func NewTrustPolicy(catalog Catalog, owner []Signer) (TrustPolicy, error) {
	p := TrustPolicy{signers: []Signer{OpenFiefCI()}, catalog: catalog}
	for i := range catalog.Entries {
		p.signers = append(p.signers, catalog.Entries[i].Publisher)
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
		p.signers = append(p.signers, s)
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
// signer signed it. It fails with ErrUnsigned, ErrUntrustedSigner or
// ErrManifestMismatch.
func (p *TrustPolicy) Decide(img *SignedImage, declared *Manifest) (TrustDecision, error) {
	if len(img.Signatures) == 0 {
		return TrustDecision{}, fmt.Errorf("%w: %s", ErrUnsigned, img.Image)
	}
	d, ok := p.trusted(img)
	if !ok {
		names := make([]string, 0, len(img.Signatures))
		for _, s := range img.Signatures {
			names = append(names, s.String())
		}
		return TrustDecision{}, fmt.Errorf("%w: %s is signed by %s", ErrUntrustedSigner, img.Image, strings.Join(names, ", "))
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

// signedBy reports whether any of img's signatures was made by id.
func signedBy(img *SignedImage, id SignerIdentity) bool {
	return slices.ContainsFunc(img.Signatures, id.Matches)
}
