package modelrouter

// OpinionsStatus is whether the Ask box may give opinions to a person who
// asks it now (D-225, D-229), and when it may not, whether the owner turned
// them on and the route is what keeps them off.
type OpinionsStatus string

// The statuses of opinions.
const (
	// OpinionsOff: the Ask box gives none, and says nothing about it: the
	// owner left them off (the default), the one switch or the Ask box's
	// toggle is off, the person asking is not the owner, or the core is a
	// public demo, which never gives one (R-7).
	OpinionsOff OpinionsStatus = "off"
	// OpinionsOn: the Ask box may say whether it would buy, sell or hold and
	// suggest a thesis, grounded in the cards it shows and the public data it
	// fetched.
	OpinionsOn OpinionsStatus = "on"
	// OpinionsRouteIneligible: the owner turned opinions on for the owner's
	// own install, but no model of the Ask box's route is allowed to receive
	// the question and cards (T0), a thesis's text (T1) and public data (T2)
	// together, so opinions stay off in effect and the box says why.
	OpinionsRouteIneligible OpinionsStatus = "route-ineligible"
)

// OpinionsTiers are the data tiers a request of the Ask box with opinions on
// holds: the person's question and the cards it reads (T0), the thesis a model
// writes (T1) and the prices it fetches (T2). The Ask box's route needs a
// model the owner allowed for all three together before opinions work
// (D-206, D-229).
func OpinionsTiers() DataTierSet {
	return DataTierSet{DataTierT0, DataTierT1, DataTierT2}
}

// OpinionsInput is what Opinions decides over.
type OpinionsInput struct {
	// Settings are the stored settings: the switch, its toggles and the
	// owner's opinions setting.
	Settings Settings
	// Owner is whether the person asking holds the owner's session. A viewer's
	// session never gets opinions: they are the owner's, on the owner's own
	// install (D-225), and a viewer is another person.
	Owner bool
	// Demo is whether the core is a public demo (D-68).
	Demo bool
	// RouteEligible is whether a model of the Ask box's route may receive
	// OpinionsTiers together, with tools. It is only read when everything else
	// says opinions are on.
	RouteEligible func() (bool, error)
}

// Opinions decides whether the Ask box may give opinions now. They are off
// unless the owner turned them on, the one switch and the Ask box's toggle
// allow AI features (D-212), the asker is the owner and the core is no demo
// (R-7); then they work only when the route has a model allowed for every tier
// they use. RouteEligible's error is returned as it is.
func Opinions(in *OpinionsInput) (OpinionsStatus, error) {
	if in.Demo || !in.Owner || !in.Settings.Opinions || !in.Settings.Switch.Allows(FeatureAsk) || !in.Settings.HasModel(RoleAsk) {
		return OpinionsOff, nil
	}
	ok, err := in.RouteEligible()
	switch {
	case err != nil:
		return OpinionsOff, err
	case !ok:
		return OpinionsRouteIneligible, nil
	}
	return OpinionsOn, nil
}
