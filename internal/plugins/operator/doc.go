// Package operator is the plugin operator's Kubernetes adapter (D-13): it
// reconciles Plugin and PluginGrant resources in the openfief namespace
// (D-15) into each plugin's ServiceAccount, config ConfigMap, Service,
// NetworkPolicy and Deployment, and publishes every running plugin's egress
// allow-list in the plugins.EgressAllowListConfigMap ConfigMap for the egress
// proxy (D-32).
//
// Every decision is the domain's: the resources are translated into
// plugins.Plugin and plugins.Grant, plugins.PlanRollout picks the version that
// runs (D-16), plugins.Compile turns its effective permissions into a
// plugins.Workload, and BuildObjects maps that Workload onto Kubernetes
// objects. The operator never creates Secret data (D-29) and never creates an
// Ingress (D-28).
//
// Images are deployed only after an ImageVerifier confirms their cosign
// signature (D-17). Verification is not implemented yet: the default
// UnimplementedVerifier reports every image as unverified, so the operator
// deploys no plugin until a real verifier replaces it.
package operator
