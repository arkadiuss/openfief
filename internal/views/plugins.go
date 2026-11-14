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
	conditionTrue         = "True"
	conditionFalse        = "False"
	signatureVerified     = "Verified"
	signerOpenFiefCI      = "OpenFiefCI"
	signerPublisher       = "Publisher"
	signerOwner           = "Owner"
	reasonUntrustedSigner = "UntrustedSigner"
	phaseRemoving         = "Removing"
	phasePending          = "Pending"
	imageDigestPrefix     = "@"
	imageDigestAlgoSep    = ":"
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
// none, and whether it is a market-data connector: its manifest declares it
// (spec.manifest.marketData) and the grant approves it
// (status.effective.marketData) (D-123).
type PluginRecord struct {
	Instance string
	Image    string
	Deleting bool
	// Generation is metadata.generation: the spec's, which the operator's
	// apply of a new spec raises, and which the audit log keeps as the
	// generation of an install or an upgrade (D-211). Zero when unknown.
	Generation int64
	Requested  []string
	Effective  []string
	Ready      Condition
	Signature  Condition
	// BrowserPath is the browser path prefix the grant approves
	// (status.effective.browserPaths, D-81): set for the view plugin the
	// core proxies browsers to, which reads with the signed-in person's
	// session, and empty for every other plugin.
	BrowserPath string
	Trust       string
	SignerKind  string
	Signer      string
	// UntrustedSigners is status.untrustedSigners: who signed an image the
	// operator refused as signed by no trusted signer (D-169).
	UntrustedSigners []string
	// RequestsMarketData: the manifest declares a market-data connector.
	RequestsMarketData bool
	// MarketData: the running version is one the grant approves.
	MarketData bool
	// RequestsBondTerms and RequestsIndexLevels: the manifest declares that
	// it delivers them; BondTerms and IndexLevels: the grant approves that
	// for the running version (spec.manifest and status.effective, D-179).
	RequestsBondTerms, BondTerms     bool
	RequestsIndexLevels, IndexLevels bool
	// RequestsIndexReadings and IndexReadings: the same of the index
	// readings bond terms depend on (D-186).
	RequestsIndexReadings, IndexReadings bool
	// Institutions are the institutions its manifest declares it books
	// for (D-182); ApprovedInstitutions those of them its grant approves,
	// the ones it books for (status.effective, D-185).
	Institutions         []string
	ApprovedInstitutions []string
}

// Plugin is an installed plugin instance as the read API shows it (D-46,
// D-48, D-56): its name, image and image digest, the capabilities it may
// use now and those it requests, how far its image is trusted and by which
// signer, why (the SignatureVerified reason), and its phase, the Ready
// condition's reason, with whether it runs. MarketData says that each sync
// tells it every instrument the owner holds and the model's targets, and
// RequestsMarketData that its manifest asks for that (D-93, D-123).
// Institutions are those it books for (D-182).
type Plugin struct {
	Instance     string
	Image        string
	Digest       string
	Capabilities []string
	Requested    []string
	Trust        SignatureTrust
	Signer       string
	TrustReason  string
	// UntrustedSigners names who signed an image refused because no
	// trusted signer did (D-169); empty otherwise.
	UntrustedSigners []string
	Phase            string
	Ready            bool
	// RequestsMarketData: its manifest declares a market-data connector.
	RequestsMarketData bool
	// MarketData: it is one, as the grant approves (D-123).
	MarketData bool
	// BondTerms and IndexLevels: the core stores the treasury bond series
	// terms and market index levels it delivers, as the grant approves;
	// RequestsBondTerms and RequestsIndexLevels: its manifest declares them
	// (D-179).
	RequestsBondTerms, BondTerms     bool
	RequestsIndexLevels, IndexLevels bool
	// IndexReadings: the core stores the index readings it delivers, such
	// as CPI and the NBP reference rate, as the grant approves;
	// RequestsIndexReadings: its manifest declares them (D-186).
	RequestsIndexReadings, IndexReadings bool
	// Institutions are the institutions its manifest declares it books
	// for (D-182); ApprovedInstitutions those of them its grant approves.
	// The core refuses a delivery naming any but the approved ones (D-185).
	Institutions         []string
	ApprovedInstitutions []string
}

// PluginOf returns the Plugin r describes. The signer is named only when a
// trusted signer signed the image, the untrusted signers only when the
// image is refused for them (D-169). The phase is Ready's reason, Removing
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

		RequestsMarketData: r.RequestsMarketData,
		MarketData:         r.MarketData,
		Institutions:       slices.Clone(r.Institutions),

		ApprovedInstitutions: slices.Clone(r.ApprovedInstitutions),

		RequestsBondTerms: r.RequestsBondTerms, BondTerms: r.BondTerms,
		RequestsIndexLevels: r.RequestsIndexLevels, IndexLevels: r.IndexLevels,

		RequestsIndexReadings: r.RequestsIndexReadings, IndexReadings: r.IndexReadings,
	}
	if p.Trust.Signed() {
		p.Signer = r.Signer
	}
	if p.Trust == TrustRefused && r.Signature.Reason == reasonUntrustedSigner {
		p.UntrustedSigners = slices.Clone(r.UntrustedSigners)
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
