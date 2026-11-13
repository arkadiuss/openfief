package modelrouter

// Settings are the owner's agents-and-models settings (D-212, D-205,
// D-206): the switch with its toggles, each role's route and the models
// allowed per data tier. The zero value is the default: everything off, no
// routes and no model allowed for any tier.
type Settings struct {
	Switch Switch
	Routes map[Role]Route
	Policy DataTierPolicy
}

// HasModel reports whether role has a model it can be asked: the one switch
// and the role's toggle are on and its route names a model. It says nothing
// of the data tiers, which decide per request (D-206); the Ask box uses it
// to know whether a question beyond its commands has anywhere to go (D-203).
func (s Settings) HasModel(role Role) bool {
	return s.Switch.Allows(role.Feature()) && len(s.Routes[role].Models) > 0
}
