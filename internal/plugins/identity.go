package plugins

import "strings"

const (
	// Namespace is the only namespace OpenFief's operator and plugins use (D-15).
	Namespace = "openfief"
	// CoreAudience is the audience of the projected ServiceAccount token a
	// plugin presents to the core (D-27).
	CoreAudience = "openfief-core"
	// EgressAudience is the audience of the projected ServiceAccount token a
	// plugin presents to the egress proxy (D-63). A separate audience keeps
	// the proxy from replaying a plugin's token to the core.
	EgressAudience = "openfief-egress"

	serviceAccountPrefix  = "plugin-"
	serviceAccountUserFmt = "system:serviceaccount:" + Namespace + ":"
)

// ServiceAccountName returns the name of the ServiceAccount that is plugin
// identity for the named plugin: "plugin-<name>" in the openfief namespace.
func ServiceAccountName(plugin string) string { return serviceAccountPrefix + plugin }

// PluginForServiceAccount maps an authenticated Kubernetes username, as a
// TokenReview of a projected token reports it, back to a plugin name. It
// accepts only "system:serviceaccount:openfief:plugin-<name>" with a valid
// plugin name; any other identity is not a plugin.
func PluginForServiceAccount(username string) (string, bool) {
	sa, ok := strings.CutPrefix(username, serviceAccountUserFmt)
	if !ok {
		return "", false
	}
	name, ok := strings.CutPrefix(sa, serviceAccountPrefix)
	if !ok || !validPluginName(name) {
		return "", false
	}
	return name, true
}
