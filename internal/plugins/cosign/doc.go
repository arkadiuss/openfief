// Package cosign checks plugin images in their registry (D-17, D-42, D-56):
// it finds an image's cosign signatures, verifies them with sigstore-go
// against the bundled Sigstore trust root and the trust policy's signers, and
// reads the image's manifest label. Whether a verified image may run is the
// domain's decision (plugins.TrustPolicy.Decide).
//
// # Signature format
//
// OpenFief's CI signs with `cosign sign` from cosign v3 (images.yml, `task
// image:sign`), which by default writes a Sigstore bundle (v0.3): a DSSE
// envelope around an in-toto statement whose predicate type is
// SignPredicateType and whose subject is the image digest. The bundle is
// stored as an OCI referrer of the image: a manifest whose subject is the
// image and whose one layer is the bundle, found through the registry's
// Referrers API or, on registries without it such as GHCR, the
// "sha256-<hex>" referrers tag. The older format, a ".sig" tag with the
// certificate in annotations, is not read: CI does not produce it, and
// owners signing their own plugins (D-45) use the same cosign defaults.
//
// # What is verified
//
// A keyless signature must chain to a Fulcio CA in the trust root, carry an
// SCT from a trusted CT log, be in a trusted Rekor log, and have an observer
// timestamp (Rekor's integrated time or an RFC 3161 timestamp) inside the
// certificate's validity. Its certificate's OIDC issuer and subject must be
// one of the policy's keyless signers, checked by sigstore-go. A signature
// with an owner's key must verify with that key; its log entry, if any, is
// not required. Every signature must be over the exact image digest. No
// signature, cryptography or certificate check is home-grown (D-29).
//
// # Trust root
//
// trusted_root.json is the Sigstore public-good trusted root, embedded so
// verification works offline (D-56): bundles carry their certificates and log
// proofs, so verifying needs no call to Fulcio, Rekor or a TUF mirror. It is
// refreshed with `task trustroot:update`, which fetches it through TUF from
// sigstore-go's embedded TUF root and checks it before writing.
//
// # Network
//
// The operator reads from the image's registry only: the referrers listing or
// tag, the referrer manifests and bundle blobs, the image (or index)
// manifest and the image config blobs. It needs HTTPS egress to every
// registry plugins come from, and to the blob storage those registries
// redirect to (for GHCR, ghcr.io and pkg-containers.githubusercontent.com).
// It authenticates anonymously, so plugin images must be public.
package cosign
