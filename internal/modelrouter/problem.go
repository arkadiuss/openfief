package modelrouter

// AttentionKind is the kind of the attention item a broken route entry raises
// (D-196): the Inbox lists it among the decisions, as a missing setting. The
// attention context spells the same string.
const AttentionKind = "model-route-broken"

// ProblemReason is why a route entry is broken.
type ProblemReason string

// The reasons a route entry is broken.
const (
	// ProblemMissing: no plugin instance has the entry's provider name.
	ProblemMissing ProblemReason = "missing"
	// ProblemNotModelProvider: a plugin instance has the name, and it is not
	// a model provider, such as a connector.
	ProblemNotModelProvider ProblemReason = "not-model-provider"
)

// RouteProblem is one broken entry of a role's route: the role, the model
// entry the route names and why the route can't use it. A provider that is
// only slow or down is no problem: the owner has nothing to fix.
type RouteProblem struct {
	Role   Role
	Model  Ref
	Reason ProblemReason
}

// Subject is what makes the entry's attention item unique: the route and the
// entry, "<role>/<provider>/<model>", so it names the role and the provider
// instance.
func (p RouteProblem) Subject() string { return string(p.Role) + refSeparator + p.Model.String() }

// Detail is the item's text: which route, which provider instance and what is
// wrong with it. It holds names only (GR-12).
func (p RouteProblem) Detail() string {
	why := "no plugin has that name"
	if p.Reason == ProblemNotModelProvider {
		why = "that plugin is not a model provider"
	}
	return "the " + string(p.Role) + " route names provider instance " + p.Model.Instance + ": " + why
}
