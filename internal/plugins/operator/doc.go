// Package operator is the plugin operator's Kubernetes adapter (D-13): it
// reconciles Plugin and PluginGrant resources in the openfief namespace
// (D-15) into each plugin's ServiceAccount, config ConfigMap, Service,
// NetworkPolicy and Deployment, and publishes every running plugin's egress
// allow-list in the plugins.EgressAllowListConfigMap ConfigMap for the egress
// proxy (D-32). SigningKeysPublisher publishes the cluster's ServiceAccount
// token-signing public keys in the plugins.SigningKeysConfigMap ConfigMap,
// which every plugin mounts to verify the core's token (D-59).
//
// Every decision is the domain's: the resources are translated into
// plugins.Plugin and plugins.Grant, plugins.PlanRollout picks the version that
// runs (D-16), plugins.Compile turns its effective permissions into a
// plugins.Workload, and BuildObjects maps that Workload onto Kubernetes
// objects. The operator never creates Secret data (D-29) and never creates an
// Ingress (D-28).
//
// Images are deployed only after an ImageVerifier confirms their cosign
// signature (D-17) and that their signed manifest label is the manifest the
// Plugin declares (D-42). The operator's verifier is PolicyVerifier: it reads
// the image through a SignedImageSource (the cosign adapter) and lets
// plugins.TrustPolicy decide (D-45, D-56). Its outcome is the
// SignatureVerified condition: True with reason Verified for a verified
// plugin, True with reason SignedByTrustedSigner for any other trusted
// signer, False when refused, Unknown when the registry could not be read,
// which is retried. Nothing deploys an image the policy refuses (D-57).
package operator
