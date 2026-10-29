package views

import (
	"slices"
	"strconv"
	"strings"
)

// SignatureTrust is how far an installed plugin's image is trusted, as the
// operator recorded it (D-45, D-56). The numbers match the read API's
// SignatureTrust enum.
type SignatureTrust int

// The trust levels of an installed plugin.
const (
	// TrustVerified: the catalog pins the image and its publisher signed it
	// (D-44).
	TrustVerified SignatureTrust = iota + 1
	// TrustOpenFiefCI: OpenFief's CI signed it; it is not a verified plugin.
	TrustOpenFiefCI
	// TrustPublisher: a catalog publisher signed it, but the catalog does
	// not pin this image to them.
	TrustPublisher
	// TrustOwner: a signer the owner added signed it (D-45).
	TrustOwner
	// TrustSigned: a trusted signer signed it, and the status does not say
	// which kind, as an operator older than status.signature writes it.
	TrustSigned
	// TrustRefused: no trusted signer signed it, or its manifest label does
	// not match; it is never deployed (D-57).
	TrustRefused
	// TrustNotChecked: the operator has not checked it yet, or could not
	// read its registry.
	TrustNotChecked
)

// The operator's status values the trust level is read from
// (internal/plugins/operator): condition statuses, the trust of
// status.signature and its signer kinds.
const (
	conditionTrue      = "True"
	conditionFalse     = "False"
	signatureVerified  = "Verified"
	signerOpenFiefCI   = "OpenFiefCI"
	signerPublisher    = "Publisher"
	signerOwner        = "Owner"
	phaseRemoving      = "Removing"
	phasePending       = "Pending"
	imageDigestPrefix  = "@"
	imageDigestAlgoSep = ":"
)

// Condition is one status condition of a Plugin: its status, True, False
// or Unknown, and its reason. Both are empty when the condition is absent.
type Condition struct {
	Status string
	Reason string
}

// PluginRecord is what the operator's Plugin resource says about one
// installed instance: its name and image, whether it is being deleted, the
// capabilities it requests and those it may use now, its Ready and
// SignatureVerified conditions, and the trusted signature on its image
// (status.signature: trust, signer kind and signer), empty while there is
// none.
type PluginRecord struct {
	Instance   string
	Image      string
	Deleting   bool
	Requested  []string
	Effective  []string
	Ready      Condition
	Signature  Condition
	Trust      string
	SignerKind string
	Signer     string
}

// Plugin is an installed plugin instance as the read API shows it (D-46,
// D-48, D-56): its name, image and image digest, the capabilities it may
// use now and those it requests, how far its image is trusted and by which
// signer, why (the SignatureVerified reason), and its phase, the Ready
// condition's reason, with whether it runs.
type Plugin struct {
	Instance     string
	Image        string
	Digest       string
	Capabilities []string
	Requested    []string
	Trust        SignatureTrust
	Signer       string
	TrustReason  string
	Phase        string
	Ready        bool
}

// PluginOf returns the Plugin r describes. The signer is named only when a
// trusted signer signed the image. The phase is Ready's reason, Removing
// while the resource is being deleted, and Pending before the operator
// wrote a Ready condition.
func PluginOf(r *PluginRecord) Plugin {
	p := Plugin{
		Instance:     r.Instance,
		Image:        r.Image,
		Digest:       ImageDigest(r.Image),
		Capabilities: slices.Clone(r.Effective),
		Requested:    slices.Clone(r.Requested),
		Trust:        TrustOf(r),
		TrustReason:  r.Signature.Reason,
		Phase:        r.Ready.Reason,
		Ready:        r.Ready.Status == conditionTrue,
	}
	if p.Trust.Signed() {
		p.Signer = r.Signer
	}
	switch {
	case r.Deleting:
		p.Phase, p.Ready = phaseRemoving, false
	case p.Phase == "":
		p.Phase = phasePending
	}
	return p
}

// TrustOf returns the trust level r's SignatureVerified condition and
// status.signature give: True is verified, or signed by the kind of signer
// status.signature names; False is refused; Unknown or no condition is not
// checked.
func TrustOf(r *PluginRecord) SignatureTrust {
	switch r.Signature.Status {
	case conditionTrue:
	case conditionFalse:
		return TrustRefused
	default:
		return TrustNotChecked
	}
	if r.Trust == signatureVerified {
		return TrustVerified
	}
	switch r.SignerKind {
	case signerOpenFiefCI:
		return TrustOpenFiefCI
	case signerPublisher:
		return TrustPublisher
	case signerOwner:
		return TrustOwner
	}
	if r.Signature.Reason == signatureVerified {
		return TrustVerified
	}
	return TrustSigned
}

// String names the trust level in snake case, as the read API's enum
// does after its prefix.
func (t SignatureTrust) String() string {
	switch t {
	case TrustVerified:
		return "verified"
	case TrustOpenFiefCI:
		return "openfief_ci"
	case TrustPublisher:
		return "publisher"
	case TrustOwner:
		return "owner"
	case TrustSigned:
		return "trusted_signer"
	case TrustRefused:
		return "refused"
	case TrustNotChecked:
		return "not_checked"
	}
	return "SignatureTrust(" + strconv.Itoa(int(t)) + ")"
}

// Signed reports whether t means a trusted signer signed the image.
func (t SignatureTrust) Signed() bool {
	switch t {
	case TrustVerified, TrustOpenFiefCI, TrustPublisher, TrustOwner, TrustSigned:
		return true
	case TrustRefused, TrustNotChecked:
	}
	return false
}

// ImageDigest returns the digest of an image reference pinned by digest,
// such as "sha256:…" from "repository@sha256:…", or "" when it has none.
func ImageDigest(image string) string {
	_, digest, ok := strings.Cut(image, imageDigestPrefix)
	if !ok || !strings.Contains(digest, imageDigestAlgoSep) {
		return ""
	}
	return digest
}
