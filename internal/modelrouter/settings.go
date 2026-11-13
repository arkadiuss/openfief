package modelrouter

// Settings are the owner's agents-and-models settings (D-212, D-205,
// D-206): the switch with its toggles, each role's route and the models
// allowed per data tier. The zero value is the default: everything off, no
// routes and no model allowed for any tier.
type Settings struct {
	Switch Switch
	Routes map[Role]Route
	Policy TierPolicy
}
