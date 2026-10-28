// Package plugins is the plugins bounded context: what a plugin requests
// (Plugin, Manifest), what the owner approved (Grant), what is enforced
// (EffectivePermissions, D-16), and how that compiles into a Workload: the
// domain description of a plugin's NetworkPolicy, Secret mounts, RBAC,
// projected token and egress allow-list (D-11, D-27, D-32).
//
// Everything here is pure. Kubernetes objects are built from a Workload by
// the operator's adapter; the egress proxy's address checks, which need
// net/netip, live in the egress subpackage.
package plugins
