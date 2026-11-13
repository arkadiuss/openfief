// Package modelrouter is the modelrouter bounded context: which models may be called,
// for what and with which data (docs/DESIGN.md D-205, D-206, D-212).
//
// It holds the one switch for agents and models with a toggle per feature
// (Switch, D-212), the references that name a model at a provider instance
// (Ref), the roles that need a model and each role's primary model and
// fallbacks (Role, Route, D-205), the data tiers every piece of data sent to
// a model carries and the models the owner allows per tier (DataTier, DataTierPolicy,
// D-206), and Plan, the one decision the model router makes before a call:
// which models, in order, may be sent a request, or why none may.
//
// The package is pure: no I/O, no clock, no logging. A request's text is the
// owner's data and is never part of an error (GR-12).
package modelrouter
