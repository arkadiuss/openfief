// Package app holds the modelrouter context's use cases (docs/DESIGN.md D-205,
// D-206, D-212): the owner's settings of the one switch, the routes and the
// models allowed per data tier, and the model router, which every call to a
// model goes through. Ports are declared here, by their consumer.
package app

import (
	"context"
	"errors"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/modelrouter"
)

// Errors of the models use cases.
var (
	// ErrNoSession reports a session token that is missing, unknown,
	// expired or revoked.
	ErrNoSession = errors.New("no valid session")
	// ErrNotOwner reports a valid session without owner scope, or an origin
	// the core does not serve the owner on.
	ErrNotOwner = errors.New("only the owner may do this")
	// ErrInvalidInput reports a feature, role, tier or reference the
	// settings do not accept.
	ErrInvalidInput = errors.New("invalid input")
	// ErrOpinionsOnDemo reports an attempt to turn opinions on in a public
	// demo, which never gives one (D-225, R-7).
	ErrOpinionsOnDemo = errors.New("a public demo never gives opinions")
	// ErrProviderUnavailable reports a provider instance that can't be
	// reached, is down, is too slow or refused the plugin's key: the router
	// tries the route's next model.
	ErrProviderUnavailable = errors.New("model provider unavailable")
	// ErrProviderMissing reports a route that names a provider instance no
	// plugin has: the router tries the route's next model and raises an
	// attention item for the broken entry (D-196).
	ErrProviderMissing = errors.New("model provider instance not found")
	// ErrNotModelProvider reports a route that names a plugin instance that
	// is not a model provider, such as a connector: the router treats it
	// like a missing one.
	ErrNotModelProvider = errors.New("plugin instance is not a model provider")
	// ErrPluginsNotFound reports that the core has not found the plugins
	// yet, so it can't say whether a provider instance exists. It wraps
	// ErrProviderUnavailable in a call, and makes the route check skip its
	// round, so items raised earlier are not cleared on a restart.
	ErrPluginsNotFound = errors.New("the core has not found the plugins yet")
	// ErrProviderRejected reports a request the provider or its plugin
	// rejected, such as an unknown model or a malformed message list: the
	// router does not try the next model with the same request.
	ErrProviderRejected = errors.New("model provider rejected the request")
	// ErrAuditUnavailable reports that a call's audit entry could not be
	// written, so the call was refused (D-211: the log fails closed).
	ErrAuditUnavailable = errors.New("the audit entry could not be written")
	// ErrNoAnswer reports that every candidate model of the route failed.
	ErrNoAnswer = errors.New("no model of the route answered")
	// ErrBusy reports that the Ask box's calls at once are all taken and none
	// freed within the queue wait (D-208): the router refused the call.
	ErrBusy = errors.New("too many model calls at once")
	// ErrUsageUnavailable reports that the day's usage could not be read or
	// recorded, so the call was refused or its answer withheld: the budget
	// fails closed like the audit log.
	ErrUsageUnavailable = errors.New("the day's model usage could not be read or recorded")
)

// Owners checks that a session is the owner's (D-66, D-79); the identity
// context answers it in cmd/core, with errors wrapping ErrNoSession or
// ErrNotOwner.
type Owners interface {
	CheckOwner(ctx context.Context, origin, sessionToken string) error
	// OwnerSession checks the same and returns the session's ID, which the
	// audit entry of a change names as its actor.
	OwnerSession(ctx context.Context, origin, sessionToken string) (string, error)
}

// SettingsStore keeps the owner's agents-and-models settings.
// *postgres.Store implements it.
type SettingsStore interface {
	// Settings returns the stored settings; the zero Settings, everything
	// off, when none was ever saved.
	Settings(ctx context.Context) (modelrouter.Settings, error)
	//
	// Each Save runs within, when it is not nil, inside the transaction that
	// stores the change, before it commits: the audit entry of the change.
	// When within fails nothing is stored.
	//
	// SaveEnabled turns the one switch on or off.
	SaveEnabled(ctx context.Context, on bool, within func(context.Context) error) error
	// SaveOpinions turns the Ask box's opinions on or off (D-225).
	SaveOpinions(ctx context.Context, on bool, within func(context.Context) error) error
	// SaveToggle turns one feature's toggle on or off.
	SaveToggle(ctx context.Context, f modelrouter.Feature, on bool, within func(context.Context) error) error
	// SaveRoute replaces role's route; no models clear it.
	SaveRoute(ctx context.Context, role modelrouter.Role, route []modelrouter.Ref, within func(context.Context) error) error
	// SaveDataTierModels replaces tier's allowed models; none clears it.
	SaveDataTierModels(ctx context.Context, tier modelrouter.DataTier, allowed []modelrouter.Ref, within func(context.Context) error) error
}

// SettingChange is what the audit log keeps of a change to the settings
// (D-211, D-212): who made it, when, and which setting, as the identity
// context spells it ("switch:on", "feature:ask:off", "route:ask",
// "tier:T0", "opinions:on"). It holds no model list.
type SettingChange struct {
	At time.Time
	// Session is the owner's session that made the change.
	Session string
	Setting string
}

// SettingsAuditor appends the entry of a change to the settings to the
// audit log. cmd/core implements it over the identity context's log. A
// change whose entry can't be written is not made.
type SettingsAuditor interface {
	RecordChange(ctx context.Context, c *SettingChange) error
}

// ProviderDirectory lists the model-provider plugin instances that run now,
// in name order. cmd/core answers it over the plugin discovery; before the
// core has found the plugins it returns an error wrapping ErrPluginsNotFound.
type ProviderDirectory interface {
	ModelProviders(ctx context.Context) ([]string, error)
}

// ProviderInstances reaches the model-provider plugins by their instance name, the
// first part of a modelrouter.Ref (D-205). connectrpc.Client implements it. It
// reports a provider that can't be reached or is down as an error wrapping
// ErrProviderUnavailable, an instance no plugin has as one wrapping
// ErrProviderMissing, a plugin that is not a model provider as one wrapping
// ErrNotModelProvider, and a request it rejected as one wrapping
// ErrProviderRejected.
type ProviderInstances interface {
	// Instance reports whether instance is a running model-provider
	// instance, without calling it: nil, or an error wrapping
	// ErrProviderMissing, ErrNotModelProvider or ErrProviderUnavailable (it
	// exists but does not run yet, which is not a broken route).
	Instance(ctx context.Context, instance string) error
	// Describe returns where the provider's models run and its retention
	// terms.
	Describe(ctx context.Context, provider string) (modelrouter.ProviderInstance, error)
	// Models lists the provider's model entries.
	Models(ctx context.Context, provider string) ([]modelrouter.ModelEntry, error)
	// Generate asks the provider's model entry for an answer.
	Generate(ctx context.Context, provider, model string, req *modelrouter.Request) (modelrouter.Result, error)
}

// ActorKind is who a model call is made for.
type ActorKind string

// The actor kinds of a model call.
const (
	// ActorOwner is a call made for the owner in a session, such as an Ask
	// box question; the actor's ID is the session's ID.
	ActorOwner ActorKind = "owner"
	// ActorViewer is a call made for a viewer in a session, such as an Ask
	// box question the viewer asked; the actor's ID is the session's ID.
	ActorViewer ActorKind = "viewer"
	// ActorSystem is a call the core makes on its own, such as the
	// maintainer's weekly note.
	ActorSystem ActorKind = "system"
)

// Actor is who a model call is made for. It names the party by ID only.
type Actor struct {
	Kind ActorKind
	ID   string
}

// Outcome is how a model call ended, as its result entry records it.
type Outcome string

// The outcomes of a model call.
const (
	// OutcomeOK is a call a model answered.
	OutcomeOK Outcome = "ok"
	// OutcomeRefused is a call the router would not send: the switch or the
	// role's toggle is off, no model may receive the data, or the data is
	// untagged.
	OutcomeRefused Outcome = "refused"
	// OutcomeFailed is a call that came back with an error, or never
	// reached its model; its ErrorClass says which.
	OutcomeFailed Outcome = "failed"
)

// ErrorClass is why a failed call failed, a closed set that names no text
// (GR-12): the audit log keeps it as the entry's error class.
type ErrorClass string

// The error classes of a failed call.
const (
	// ClassUnavailable: the provider could not be reached, was too slow or
	// refused the plugin's key. The router tried the next model.
	ClassUnavailable ErrorClass = "unavailable"
	// ClassRejected: the provider or its plugin rejected the request.
	ClassRejected ErrorClass = "rejected"
	// ClassCanceled: the caller gave up.
	ClassCanceled ErrorClass = "canceled"
	// ClassMisconfigured: the route names a provider instance that is
	// missing or is not a model provider. The router tried the next model.
	ClassMisconfigured ErrorClass = "misconfigured"
	// ClassInternal: anything else.
	ClassInternal ErrorClass = "internal"
	// ClassBudget: the router refused the call because the day's budget is
	// spent (D-208), or because the call could not be priced against it.
	ClassBudget ErrorClass = "budget"
	// ClassBusy: the router refused the call because the Ask box's calls at
	// once were all taken (D-208).
	ClassBusy ErrorClass = "busy"
)

// CallSent is what the audit log keeps of a call about to be sent (D-211),
// written before it: who, for which role and model by name, under which call
// ID, and when. It has no field for a question, an answer, a tool's argument
// or an amount (GR-12), nor for a token count: usage is the budget's to keep
// (D-208).
type CallSent struct {
	At     time.Time
	Actor  Actor
	Role   modelrouter.Role
	Model  modelrouter.Ref
	CallID string
}

// CallResult is what the audit log keeps of how a call ended, under the call
// ID of its CallSent: the same fields, the outcome and, for a failure, the
// error class. A call refused before a model was chosen has only a
// CallResult, with a zero Model.
type CallResult struct {
	At      time.Time
	Actor   Actor
	Role    modelrouter.Role
	Model   modelrouter.Ref
	CallID  string
	Outcome Outcome
	Class   ErrorClass
}

// Auditor appends the two entries of a model call to the append-only audit
// log. cmd/core implements it over the identity context's log.
type Auditor interface {
	// RecordSent writes the entry that comes before a call; when it fails,
	// the call is not made.
	RecordSent(ctx context.Context, c *CallSent) error
	// RecordResult writes the entry that says how a call ended.
	RecordResult(ctx context.Context, c *CallResult) error
}

// CallIDs makes the ID that links a call's two audit entries: 16 lower-case
// hex digits from a random source.
type CallIDs func() (string, error)

// UsageRecord is what one answered call used, for the day's budget (D-208):
// the day it is counted in, the model entry, the tokens the provider reported
// and what they cost at the entry's price. It holds no text (GR-12).
type UsageRecord struct {
	Day      kernel.Date
	Model    modelrouter.Ref
	Usage    modelrouter.Usage
	Cost     kernel.Decimal
	Currency kernel.Currency
}

// UsageStore keeps the day's usage across restarts and across the core's
// replicas: the budget is the sum of what every call of the day cost.
// *postgres.Store implements it.
type UsageStore interface {
	// Spent returns what the calls of day cost in currency.
	Spent(ctx context.Context, day kernel.Date, currency kernel.Currency) (kernel.Decimal, error)
	// Record adds one call's usage to its day, atomically.
	Record(ctx context.Context, u *UsageRecord) error
}

// QuestionAsked is one question of the Ask box that goes to a model: when,
// which session asked and, when it is known, from which address.
type QuestionAsked struct {
	At time.Time
	// Session is the asking session's ID.
	Session string
	// Address is the asker's IP address as the core's front door found it
	// (the trusted proxy's header, or the connection); empty when unknown,
	// which leaves the address limit out.
	Address string
	Limits  modelrouter.QuestionLimits
}

// QuestionSubject is whom a limit counts: a session or an address.
type QuestionSubject string

// The subjects a limit counts.
const (
	SubjectSession QuestionSubject = "session"
	SubjectAddress QuestionSubject = "address"
)

// QuestionRuling is a store's decision on one question.
type QuestionRuling struct {
	// Allowed is true when the question was recorded and may go on.
	Allowed bool
	// Subject is whose limit stopped it, when it was not allowed.
	Subject QuestionSubject
	// ResetsAt is when that subject may ask again.
	ResetsAt time.Time
}

// QuestionStore counts the Ask box's questions to a model per session and per
// address, in the database, so the counts survive a restart and every replica
// of the core sees the same ones. *postgres.Store implements it.
type QuestionStore interface {
	// Admit decides, with the subjects' earlier questions of the last day and
	// modelrouter.DecideQuestion, whether q may go to the model, and records q
	// when it may, as one step that two questions of one subject at the same
	// moment can't both pass the last place of. A question that is not allowed
	// is not recorded.
	Admit(ctx context.Context, q *QuestionAsked) (QuestionRuling, error)
}

// AttentionRaiser raises and clears the attention items of broken routes
// (D-196), one condition at a time, with the kind and subject the
// modelrouter package builds (RouteProblem). cmd/core wires it to the
// attention items, so neither context imports the other.
type AttentionRaiser interface {
	Raise(ctx context.Context, kind, subject, detail string) error
	Clear(ctx context.Context, kind, subject string) error
}
