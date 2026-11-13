// Package app holds the modelrouter context's use cases (docs/DESIGN.md D-205,
// D-206, D-212): the owner's settings of the one switch, the routes and the
// models allowed per data tier, and the model router, which every call to a
// model goes through. Ports are declared here, by their consumer.
package app

import (
	"context"
	"errors"
	"time"

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
	// ErrProviderUnavailable reports a provider instance that can't be
	// reached, is down, is too slow or refused the plugin's key: the router
	// tries the route's next model.
	ErrProviderUnavailable = errors.New("model provider unavailable")
	// ErrProviderRejected reports a request the provider or its plugin
	// rejected, such as an unknown model or a malformed message list: the
	// router does not try the next model with the same request.
	ErrProviderRejected = errors.New("model provider rejected the request")
	// ErrAuditUnavailable reports that a call's audit entry could not be
	// written, so the call was refused (D-211: the log fails closed).
	ErrAuditUnavailable = errors.New("the audit entry could not be written")
	// ErrNoAnswer reports that every candidate model of the route failed.
	ErrNoAnswer = errors.New("no model of the route answered")
)

// Owners checks that a session is the owner's (D-66, D-79); the identity
// context answers it in cmd/core, with errors wrapping ErrNoSession or
// ErrNotOwner.
type Owners interface {
	CheckOwner(ctx context.Context, origin, sessionToken string) error
}

// SettingsStore keeps the owner's agents-and-models settings.
// *postgres.Store implements it.
type SettingsStore interface {
	// Settings returns the stored settings; the zero Settings, everything
	// off, when none was ever saved.
	Settings(ctx context.Context) (modelrouter.Settings, error)
	// SaveEnabled turns the one switch on or off.
	SaveEnabled(ctx context.Context, on bool) error
	// SaveToggle turns one feature's toggle on or off.
	SaveToggle(ctx context.Context, f modelrouter.Feature, on bool) error
	// SaveRoute replaces role's route; no models clear it.
	SaveRoute(ctx context.Context, role modelrouter.Role, route []modelrouter.Ref) error
	// SaveTierModels replaces tier's allowed models; none clears it.
	SaveTierModels(ctx context.Context, tier modelrouter.Tier, allowed []modelrouter.Ref) error
}

// Providers reaches the model-provider plugins by their instance name, the
// first part of a modelrouter.Ref (D-205). connectrpc.Client implements it. It
// reports a provider that can't be reached or is down as an error wrapping
// ErrProviderUnavailable, and a request it rejected as one wrapping
// ErrProviderRejected.
type Providers interface {
	// Describe returns where the provider's models run and its retention
	// terms.
	Describe(ctx context.Context, provider string) (modelrouter.Provider, error)
	// Models lists the provider's model entries.
	Models(ctx context.Context, provider string) ([]modelrouter.Entry, error)
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
	// ActorSystem is a call the core makes on its own, such as the
	// maintainer's weekly note.
	ActorSystem ActorKind = "system"
)

// Actor is who a model call is made for. It names the party by ID only.
type Actor struct {
	Kind ActorKind
	ID   string
}

// Outcome is how a recorded model call ended.
type Outcome string

// The outcomes of a recorded model call.
const (
	// OutcomeCalled is a call sent to a model: recorded before it is sent.
	OutcomeCalled Outcome = "ok"
	// OutcomeRefused is a call the router would not send: the switch or the
	// role's toggle is off, no model may receive the data, or the data is
	// untagged.
	OutcomeRefused Outcome = "refused"
	// OutcomeFailed is a call that came back with an error.
	OutcomeFailed Outcome = "failed"
)

// CallRecord is what the audit log keeps of one model call (D-211): who,
// what, which model by name, when and the outcome. It has no field for a
// question, an answer, a tool's argument or an amount (GR-12), nor for a
// token count: usage is the budget's to keep (D-208).
type CallRecord struct {
	At    time.Time
	Actor Actor
	Role  modelrouter.Role
	// Model is the entry called; zero for a refusal, where none was chosen.
	Model   modelrouter.Ref
	Outcome Outcome
}

// Auditor appends a model call to the append-only audit log. cmd/core
// implements it over the identity context's log.
type Auditor interface {
	RecordCall(ctx context.Context, rec *CallRecord) error
}
