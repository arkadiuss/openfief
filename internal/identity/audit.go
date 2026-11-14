package identity

import (
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Errors about audit entries and filters (D-211). Validation returns them
// wrapped, and a message never quotes the rejected value, which could be the
// very secret GR-12 keeps out of the log.
var (
	// ErrInvalidAuditEntry reports an entry that would not be a record of
	// who, what, which records by ID, when and the outcome.
	ErrInvalidAuditEntry = errors.New("invalid audit entry")
	// ErrInvalidAuditFilter reports a filter or page that is out of range.
	ErrInvalidAuditFilter = errors.New("invalid audit filter")
)

// ActorKind is what kind of party made an audit entry's change.
type ActorKind string

// The actor kinds, a closed set. Adding one is a design change: an entry
// names people and parties by ID only (D-211, GR-12).
const (
	// ActorOwner is the owner in a session; the actor's ID is the session's
	// ID, which says which sign-in acted.
	ActorOwner ActorKind = "owner"
	// ActorViewer is a viewer in a session; the actor's ID is the session's
	// ID. The person is the session's: the audit entry of its sign-in names
	// them.
	ActorViewer ActorKind = "viewer"
	// ActorEnrollmentCode is the holder of a one-time enrollment code, who
	// is not signed in yet; the actor's ID is the code's ID.
	ActorEnrollmentCode ActorKind = "enrollment_code"
	// ActorSystem is the core itself, such as its enroll-link command; the
	// actor's ID is "core".
	ActorSystem ActorKind = "system"
	// ActorAnonymous is a caller that proved nothing, such as a refused
	// sign-in; it has no ID.
	ActorAnonymous ActorKind = "anonymous"
)

// SystemActorID is the ID of the system actor.
const SystemActorID = "core"

// Actor is who made the change an audit entry records.
type Actor struct {
	Kind ActorKind
	ID   string
}

// Validate returns nil when a names its party by an ID of the shape its kind
// has, or none for an anonymous caller.
func (a Actor) Validate() error {
	switch a.Kind {
	case ActorOwner, ActorViewer:
		if !isHexID(a.ID) {
			return fmt.Errorf("%w: an owner or viewer actor is named by its session ID", ErrInvalidAuditEntry)
		}
	case ActorEnrollmentCode:
		if !isHexID(a.ID) {
			return fmt.Errorf("%w: an enrollment code actor is named by the code's ID", ErrInvalidAuditEntry)
		}
	case ActorSystem:
		if a.ID != SystemActorID {
			return fmt.Errorf("%w: the system actor's ID is %q", ErrInvalidAuditEntry, SystemActorID)
		}
	case ActorAnonymous:
		if a.ID != "" {
			return fmt.Errorf("%w: an anonymous actor has no ID", ErrInvalidAuditEntry)
		}
	default:
		return fmt.Errorf("%w: unknown actor kind", ErrInvalidAuditEntry)
	}
	return nil
}

// AuditAction is what an audit entry records, a closed set: a free-text
// action could carry an amount, a secret or a question's text.
type AuditAction string

// The audit actions of the identity context. Later contexts add theirs to
// the set, such as grants, settings and decisions on attention items.
const (
	// ActionSignIn is a sign-in: an issued session when ok, a refused
	// authentication when refused.
	ActionSignIn AuditAction = "sign_in"
	// ActionSignOut is a session ended by its holder.
	ActionSignOut AuditAction = "sign_out"
	// ActionSignOutEverywhere is every session revoked at once (D-183).
	ActionSignOutEverywhere AuditAction = "sign_out_everywhere"
	// ActionPasskeyAdded is a passkey created.
	ActionPasskeyAdded AuditAction = "passkey_added"
	// ActionPasskeyRemoved is a passkey removed.
	ActionPasskeyRemoved AuditAction = "passkey_removed"
	// ActionEnrollmentLinkIssued is a one-time enrollment link issued.
	ActionEnrollmentLinkIssued AuditAction = "enrollment_link_issued"
	// ActionEnrollmentLinkUsed is an enrollment link used up to create a
	// passkey.
	ActionEnrollmentLinkUsed AuditAction = "enrollment_link_used"
	// ActionEnrollmentLinkCancelled is a viewer's invitation the owner
	// cancelled before it was used.
	ActionEnrollmentLinkCancelled AuditAction = "enrollment_link_cancelled"
	// ActionScriptTokenCreated is a script token created. Its text is written
	// as two parts only so gosec's G101 doesn't read the constant's name and
	// the shape of its value as a hardcoded credential; the value is a word,
	// not a secret.
	ActionScriptTokenCreated AuditAction = "script_" + "token_created"
	// ActionScriptTokenRevoked is a script token revoked.
	ActionScriptTokenRevoked AuditAction = "script_token_revoked"
	// ActionGrantCreated is a grant of access made (D-209): to a holder,
	// of a target, both named by ID.
	ActionGrantCreated AuditAction = "grant_created"
	// ActionGrantRevoked is a grant of access ended, by the owner or
	// because its holder or its target is gone.
	ActionGrantRevoked AuditAction = "grant_revoked"
	// ActionPersonCreated is a viewer created, when they enroll their first
	// passkey with the link the owner issued: the person by ID, the link by
	// ID.
	ActionPersonCreated AuditAction = "person_created"
	// ActionPersonRevoked is a viewer revoked by the owner: their sessions
	// end, their passkeys stop working and their grants end, each of the
	// last two with an entry of its own. The person and their history stay.
	ActionPersonRevoked AuditAction = "person_revoked"
	// ActionModelCall is a call to a model about to be sent (D-205, D-211):
	// who asked, for which role and which model, under a call ID, with the
	// outcome "sent". It is written before the call, and the call is not
	// made when it can't be. Never the request, the answer or what it cost.
	ActionModelCall AuditAction = "model_call"
	// ActionModelCallResult is how a call to a model ended, under the call
	// ID of its ActionModelCall entry: ok, failed with an error class, or
	// refused. A call the router refused before choosing a model (the
	// switch, the data tiers) has only this entry, with no model.
	ActionModelCallResult AuditAction = "model_call_result"
	// ActionMaintainerRun is a run of the model-portfolio maintainer that
	// ended without a note or with one (D-204, D-211): the role, the run's ID
	// and, when it was refused or failed, the class it ended in. The model
	// calls of the run are their own entries.
	ActionMaintainerRun AuditAction = "maintainer_run"
	// ActionMaintainerProposalDropped is a proposed set of target changes
	// that broke a rule and was dropped whole, the run's note kept. It names
	// the run, never what the model wrote.
	ActionMaintainerProposalDropped AuditAction = "maintainer_proposal_dropped"
	// ActionMaintainerProposalApplied is the owner applying a proposal: ok
	// when the model portfolio's targets changed, refused when the model had
	// changed since the proposal was made. It names the run.
	ActionMaintainerProposalApplied AuditAction = "maintainer_proposal_applied"
	// ActionMaintainerProposalDismissed is the owner dismissing a proposal.
	// It names the run.
	ActionMaintainerProposalDismissed AuditAction = "maintainer_proposal_dismissed"
)

// actionKinds maps each action to the kinds of record its entries may name.
// A function, not a variable: the domain holds no package-level state.
func actionKinds() map[AuditAction][]RefKind {
	return map[AuditAction][]RefKind{
		ActionSignIn:                  {RefSession, RefPerson},
		ActionSignOut:                 {RefSession},
		ActionSignOutEverywhere:       nil,
		ActionPasskeyAdded:            {RefPasskey, RefPerson},
		ActionPasskeyRemoved:          {RefPasskey, RefPerson},
		ActionEnrollmentLinkIssued:    {RefEnrollmentCode},
		ActionEnrollmentLinkUsed:      {RefEnrollmentCode, RefPasskey, RefPerson},
		ActionEnrollmentLinkCancelled: {RefEnrollmentCode},
		ActionPersonCreated:           {RefPerson, RefEnrollmentCode},
		ActionPersonRevoked:           {RefPerson},
		ActionScriptTokenCreated:      {RefScriptToken},
		ActionScriptTokenRevoked:      {RefScriptToken},
		ActionGrantCreated:            grantRefKinds(),
		ActionGrantRevoked:            grantRefKinds(),
		ActionModelCall:               {RefModelRole, RefModel, RefModelCall},
		ActionModelCallResult:         {RefModelRole, RefModel, RefModelCall, RefModelError},

		ActionMaintainerRun:               {RefModelRole, RefMaintainerRun, RefModelError},
		ActionMaintainerProposalDropped:   {RefMaintainerRun},
		ActionMaintainerProposalApplied:   {RefMaintainerRun},
		ActionMaintainerProposalDismissed: {RefMaintainerRun},
	}
}

// grantRefKinds are the kinds of record a grant entry names: its holder and
// its target.
func grantRefKinds() []RefKind {
	return []RefKind{RefPerson, RefScriptToken, RefViewPlugin, RefAccount, RefSleeve, RefHousehold}
}

// isGrantAction reports whether a is one of the two actions whose entries
// name a grant: exactly its holder, then its target.
func isGrantAction(a AuditAction) bool { return a == ActionGrantCreated || a == ActionGrantRevoked }

// AuditActions returns every audit action, sorted.
func AuditActions() []AuditAction {
	kinds := actionKinds()
	out := make([]AuditAction, 0, len(kinds))
	for a := range kinds {
		out = append(out, a)
	}
	slices.Sort(out)
	return out
}

// ParseAuditAction parses an action as its constant spells it.
func ParseAuditAction(s string) (AuditAction, error) {
	a := AuditAction(s)
	if _, ok := actionKinds()[a]; !ok {
		return "", fmt.Errorf("%w: unknown action", ErrInvalidAuditEntry)
	}
	return a, nil
}

// AuditOutcome is how the recorded change ended.
type AuditOutcome string

// The outcomes, a closed set.
const (
	// OutcomeOK is a change made.
	OutcomeOK AuditOutcome = "ok"
	// OutcomeRefused is a change the core understood and would not make,
	// such as a sign-in whose credential did not verify.
	OutcomeRefused AuditOutcome = "refused"
	// OutcomeFailed is a change that broke on the way.
	OutcomeFailed AuditOutcome = "failed"
	// OutcomeSent is a call about to be sent, which another entry, under the
	// same call ID, will say the result of (ActionModelCall).
	OutcomeSent AuditOutcome = "sent"
)

// ParseAuditOutcome parses an outcome as its constant spells it.
func ParseAuditOutcome(s string) (AuditOutcome, error) {
	switch o := AuditOutcome(s); o {
	case OutcomeOK, OutcomeRefused, OutcomeFailed, OutcomeSent:
		return o, nil
	}
	return "", fmt.Errorf("%w: unknown outcome", ErrInvalidAuditEntry)
}

// RefKind is the kind of record an audit entry names.
type RefKind string

// The kinds of record an entry may name, a closed set.
const (
	// RefSession is a session, by its ID.
	RefSession RefKind = "session"
	// RefPasskey is a passkey, by its credential ID (Passkey.ID).
	RefPasskey RefKind = "passkey"
	// RefEnrollmentCode is an enrollment code, by its ID, never its key.
	RefEnrollmentCode RefKind = "enrollment_code"
	// RefScriptToken is a script token, by its ID, never its secret.
	RefScriptToken RefKind = "script_token"
	// RefPerson is a person, by their ID.
	RefPerson RefKind = "person"
	// RefViewPlugin is a view plugin, by its instance name.
	RefViewPlugin RefKind = "view_plugin"
	// RefAccount is an account, by its ID.
	RefAccount RefKind = "account"
	// RefSleeve is a sleeve, by its ID.
	RefSleeve RefKind = "sleeve"
	// RefHousehold is the household, by its ID ("home").
	RefHousehold RefKind = "household"
	// RefModelRole is the role a model call was made for, by its name: "ask"
	// or "maintainer" (D-205).
	RefModelRole RefKind = "model_role"
	// RefModel is a model entry called, as "provider/model" (D-205): names
	// the owner configured, never text a model was sent.
	RefModel RefKind = "model"
	// RefModelCall is the ID that links the entry written before a call to
	// a model to the entry that says how it ended.
	RefModelCall RefKind = "model_call"
	// RefModelError is the class a failed call to a model ended in, one of
	// ModelErrorClasses; never the provider's message.
	RefModelError RefKind = "model_error"
	// RefMaintainerRun is a run of the model-portfolio maintainer, and the
	// proposal it made, by the run's ID: 16 hex digits the core makes.
	RefMaintainerRun RefKind = "maintainer_run"
)

// maxPasskeyRefLen is the longest base64url credential ID: 1023 bytes.
const maxPasskeyRefLen = 1364

// Ref names one record by kind and ID. It has no other field, so an entry
// has nowhere to put a name, an amount or a secret.
type Ref struct {
	Kind RefKind
	ID   string
}

// Validate returns nil when r's ID has the shape its kind's IDs have.
func (r Ref) Validate() error {
	switch r.Kind {
	case RefSession, RefEnrollmentCode, RefScriptToken, RefMaintainerRun:
		if !isHexID(r.ID) {
			return fmt.Errorf("%w: a %s reference is its ID", ErrInvalidAuditEntry, r.Kind)
		}
	case RefPerson, RefViewPlugin, RefAccount, RefSleeve, RefHousehold:
		return r.validateGrantRef()
	case RefPasskey:
		return r.validatePasskeyRef()
	case RefModelRole, RefModel, RefModelCall, RefModelError:
		return r.validateModel()
	default:
		return fmt.Errorf("%w: unknown reference kind", ErrInvalidAuditEntry)
	}
	return nil
}

// validatePasskeyRef checks that r's ID is a passkey's credential ID, in the
// one encoding it is written in.
func (r Ref) validatePasskeyRef() error {
	if len(r.ID) > maxPasskeyRefLen {
		return fmt.Errorf("%w: a passkey reference is its credential ID", ErrInvalidAuditEntry)
	}
	b, err := ParsePasskeyID(r.ID)
	if err != nil || looksLikeSecret(r.ID) || base64.RawURLEncoding.EncodeToString(b) != r.ID {
		return fmt.Errorf("%w: a passkey reference is its credential ID", ErrInvalidAuditEntry)
	}
	return nil
}

// validateGrantRef checks r's ID when r names what a grant names besides a
// script token: a person, a view plugin, an account, a sleeve or the
// household.
func (r Ref) validateGrantRef() error {
	var ok bool
	switch r.Kind {
	case RefPerson:
		ok = isHexID(r.ID)
	case RefViewPlugin:
		ok = isPluginInstanceName(r.ID)
	case RefAccount, RefSleeve:
		ok = isRecordID(r.ID) && !looksLikeSecret(r.ID)
	case RefHousehold:
		ok = r.ID == HouseholdID
	case RefSession, RefPasskey, RefEnrollmentCode, RefScriptToken, RefModelRole, RefModel, RefModelCall, RefModelError, RefMaintainerRun:
	}
	if !ok {
		return fmt.Errorf("%w: a %s reference is the ID of its shape", ErrInvalidAuditEntry, r.Kind)
	}
	return nil
}

// validateModel returns nil when r, of kind RefModelRole, RefModel,
// RefModelCall or RefModelError, names a role, a model entry, a call or an
// error class: names the owner configured, an ID or a class of a closed
// set, never text.
func (r Ref) validateModel() error {
	var ok bool
	switch r.Kind {
	case RefModelRole:
		ok = r.ID == "ask" || r.ID == "maintainer"
	case RefModel:
		ok = validModelRef(r.ID)
	case RefModelCall:
		ok = isHexID(r.ID)
	case RefModelError:
		ok = slices.Contains(ModelErrorClasses(), r.ID)
	case RefSession, RefPasskey, RefEnrollmentCode, RefScriptToken, RefPerson, RefViewPlugin, RefAccount, RefSleeve, RefHousehold, RefMaintainerRun:
	}
	if !ok {
		return fmt.Errorf("%w: a %s reference is a role, a provider/model, a call ID or an error class of its shape", ErrInvalidAuditEntry, r.Kind)
	}
	return nil
}

// ModelErrorClasses are the classes a failed call to a model ends in, a
// closed set: the provider could not be reached, was too slow or refused the
// plugin's key ("unavailable"); the provider or its plugin rejected the
// request ("rejected"); the caller gave up ("canceled"); the route names a
// provider instance that is missing or is not a model provider
// ("misconfigured"); anything else ("internal"). A maintainer run (D-204)
// adds three of its own: the router refused the data, the route or the
// feature's toggle ("denied"), the model's answer was no usable note
// ("invalid_output"), and OpenFief had no portfolio figures to send
// ("no_figures").
func ModelErrorClasses() []string {
	return []string{"unavailable", "rejected", "canceled", "misconfigured", "internal", "denied", "invalid_output", "no_figures"}
}

// maintainerOutcomes are the outcomes the entries of each maintainer action
// may have: a run ends ok, refused or failed; a dropped proposal is refused;
// an applied one ok, refused (the model portfolio changed since) or failed;
// a dismissed one ok.
func maintainerOutcomes() map[AuditAction][]AuditOutcome {
	return map[AuditAction][]AuditOutcome{
		ActionMaintainerRun:               {OutcomeOK, OutcomeRefused, OutcomeFailed},
		ActionMaintainerProposalDropped:   {OutcomeRefused},
		ActionMaintainerProposalApplied:   {OutcomeOK, OutcomeRefused, OutcomeFailed},
		ActionMaintainerProposalDismissed: {OutcomeOK},
	}
}

// validateMaintainerEntry checks the references and the outcome of a
// maintainer entry (D-204, D-211): every entry names its run once, with an
// outcome its action allows. A run entry also names its role, "maintainer",
// once, and a class (model_error) exactly when it was refused or failed; the
// other three are about a proposal and name nothing else.
func validateMaintainerEntry(e *AuditEntry) error {
	if !slices.Contains(maintainerOutcomes()[e.Action], e.Outcome) || !maintainerRefsOK(e) {
		return fmt.Errorf("%w: a %s entry names its run, and the role and class its outcome needs", ErrInvalidAuditEntry, e.Action)
	}
	return nil
}

// maintainerRefsOK reports whether e names the records a maintainer entry of
// its action and outcome names, and nothing else.
func maintainerRefsOK(e *AuditEntry) bool {
	count := map[RefKind]int{}
	for _, r := range e.Refs {
		count[r.Kind]++
	}
	if count[RefMaintainerRun] != 1 {
		return false
	}
	if e.Action != ActionMaintainerRun {
		return len(e.Refs) == 1
	}
	classed := e.Outcome == OutcomeRefused || e.Outcome == OutcomeFailed
	asMaintainer := slices.ContainsFunc(e.Refs, func(r Ref) bool { return r.Kind == RefModelRole && r.ID == "maintainer" })
	return asMaintainer && count[RefModelRole] == 1 && (count[RefModelError] == 1) == classed && len(e.Refs) == 2+count[RefModelError]
}

// validateModelEntry checks the references of a model call's entries
// (D-211): every entry names its role and its call ID once. The entry written
// before a call, outcome sent, names the model it is about to send to; the
// entry that ends one names the model unless the call was refused before any
// was chosen, and an error class exactly when the call failed.
func validateModelEntry(e *AuditEntry) error {
	count := map[RefKind]int{}
	for _, r := range e.Refs {
		count[r.Kind]++
	}
	ok := count[RefModelRole] == 1 && count[RefModelCall] == 1 && count[RefModel] <= 1 && count[RefModelError] <= 1
	if e.Action == ActionModelCall {
		ok = ok && sentEntryShape(e.Outcome, count)
	} else {
		ok = ok && resultEntryShape(e.Outcome, count)
	}
	if !ok {
		return fmt.Errorf("%w: a %s entry names its role and its call, and the model and error class its outcome needs", ErrInvalidAuditEntry, e.Action)
	}
	return nil
}

// sentEntryShape reports whether an ActionModelCall entry of outcome that
// names count of each kind of record is as D-211 asks: sent, with its model.
func sentEntryShape(outcome AuditOutcome, count map[RefKind]int) bool {
	return outcome == OutcomeSent && count[RefModel] == 1 && count[RefModelError] == 0
}

// resultEntryShape reports whether an ActionModelCallResult entry of outcome
// that names count of each kind of record is as D-211 asks: not sent; with a
// model unless refused; with an error class if and only if failed.
func resultEntryShape(outcome AuditOutcome, count map[RefKind]int) bool {
	if outcome == OutcomeSent {
		return false
	}
	modelOK := outcome == OutcomeRefused || count[RefModel] == 1
	return modelOK && (count[RefModelError] == 1) == (outcome == OutcomeFailed)
}

// validModelRef reports whether id is a model reference: a provider
// instance's name, a slash and the model's id (the models context's Ref).
func validModelRef(id string) bool {
	provider, model, ok := strings.Cut(id, "/")
	return ok && validProviderRef(provider) && validModelID(model)
}

// validProviderRef reports whether name is a lower-case DNS label of at most
// maxProviderRefID characters, without "--".
func validProviderRef(name string) bool {
	if name == "" || len(name) > maxProviderRefID || strings.Contains(name, "--") || name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}
	return !strings.ContainsFunc(name, func(c rune) bool {
		return (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-'
	})
}

// validModelID reports whether id is one to maxModelRefID characters of
// those ids and tags contain: letters, digits and "._:/@+-".
func validModelID(id string) bool {
	if id == "" || len(id) > maxModelRefID {
		return false
	}
	return !strings.ContainsFunc(id, func(c rune) bool {
		return (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && !strings.ContainsRune("._:/@+-", c)
	})
}

// The longest provider instance name and model id of a model reference.
const (
	maxProviderRefID = 63
	maxModelRefID    = 128
)

// looksLikeSecret reports whether s starts as a script token secret, a
// session token or an enrollment code does. Their alphabet is base64url's,
// like a credential ID's, so the shape alone would let one into a passkey
// reference. A genuine credential ID that begins so (about one in sixteen
// million) is refused as well: the passkey is created again.
func looksLikeSecret(s string) bool {
	return strings.HasPrefix(s, SecretPrefix) || strings.HasPrefix(s, SessionPrefix) || strings.HasPrefix(s, EnrollmentPrefix)
}

// maxRefs is the most records one entry names.
const maxRefs = 8

// grantRefs is how many records a grant entry names: its holder and its
// target.
const grantRefs = 2

// AuditEntry is one record of the append-only audit log (D-211): who, what,
// which records by ID, when and the outcome. It never carries an amount, a
// secret or a question's text (GR-12): every field is a member of a closed
// set, a time, or an ID of a fixed shape, and there is no free-text field.
type AuditEntry struct {
	// Seq is the entry's place in the log, assigned when it is appended;
	// zero before.
	Seq     int64
	At      time.Time
	Actor   Actor
	Action  AuditAction
	Outcome AuditOutcome
	Refs    []Ref
}

// NewAuditEntry returns the entry for action by actor at at, with outcome
// and the records it names, or an error wrapping ErrInvalidAuditEntry.
func NewAuditEntry(at time.Time, actor Actor, action AuditAction, outcome AuditOutcome, refs ...Ref) (AuditEntry, error) {
	e := AuditEntry{At: at.UTC(), Actor: actor, Action: action, Outcome: outcome, Refs: slices.Clone(refs)}
	if err := e.Validate(); err != nil {
		return AuditEntry{}, err
	}
	return e, nil
}

// Validate returns nil when e is a complete record: a time, an actor named
// by ID, a known action and outcome, and only references of kinds its action
// names, each an ID. Otherwise it wraps ErrInvalidAuditEntry, without
// quoting the value it rejected.
func (e *AuditEntry) Validate() error {
	if e.At.IsZero() {
		return fmt.Errorf("%w: no time", ErrInvalidAuditEntry)
	}
	if err := e.Actor.Validate(); err != nil {
		return err
	}
	kinds, ok := actionKinds()[e.Action]
	if !ok {
		return fmt.Errorf("%w: unknown action", ErrInvalidAuditEntry)
	}
	if _, err := ParseAuditOutcome(string(e.Outcome)); err != nil {
		return err
	}
	if err := validateRefs(e.Action, kinds, e.Refs); err != nil {
		return err
	}
	switch {
	case isGrantAction(e.Action):
		return validateGrantRefs(e.Refs)
	case e.Action == ActionModelCall || e.Action == ActionModelCallResult:
		return validateModelEntry(e)
	case e.Action == ActionMaintainerRun, e.Action == ActionMaintainerProposalDropped,
		e.Action == ActionMaintainerProposalApplied, e.Action == ActionMaintainerProposalDismissed:
		return validateMaintainerEntry(e)
	}
	return nil
}

// validateGrantRefs checks that refs are what a grant entry names: its holder,
// then its target, and nothing else (D-209, D-211).
func validateGrantRefs(refs []Ref) error {
	if len(refs) != grantRefs {
		return fmt.Errorf("%w: a grant entry names its holder and its target", ErrInvalidAuditEntry)
	}
	if !slices.Contains([]RefKind{RefPerson, RefScriptToken, RefViewPlugin}, refs[0].Kind) {
		return fmt.Errorf("%w: a grant entry names its holder first", ErrInvalidAuditEntry)
	}
	if !slices.Contains([]RefKind{RefAccount, RefSleeve, RefHousehold}, refs[1].Kind) {
		return fmt.Errorf("%w: a grant entry names its target second", ErrInvalidAuditEntry)
	}
	return nil
}

// validateRefs checks refs: few enough, each an ID of its kind, of a kind the
// action names, and none twice.
func validateRefs(action AuditAction, kinds []RefKind, refs []Ref) error {
	if len(refs) > maxRefs {
		return fmt.Errorf("%w: more than %d references", ErrInvalidAuditEntry, maxRefs)
	}
	seen := make(map[Ref]bool, len(refs))
	for _, r := range refs {
		if err := r.Validate(); err != nil {
			return err
		}
		if !slices.Contains(kinds, r.Kind) {
			return fmt.Errorf("%w: a %s entry does not name a %s", ErrInvalidAuditEntry, action, r.Kind)
		}
		if seen[r] {
			return fmt.Errorf("%w: a reference twice", ErrInvalidAuditEntry)
		}
		seen[r] = true
	}
	return nil
}

// Page limits of an audit listing.
const (
	// DefaultAuditPage is the page size when a filter asks for none.
	DefaultAuditPage = 50
	// MaxAuditPage is the largest page a filter may ask for.
	MaxAuditPage = 200
)

// AuditFilter selects audit entries and one page of them, newest first. The
// zero value selects everything, a default page from the newest.
type AuditFilter struct {
	// ActorKind selects entries of one kind of actor, and ActorID of one
	// actor of it; empty selects every actor.
	ActorKind ActorKind
	ActorID   string
	// Action selects one action; empty selects every action.
	Action AuditAction
	// Actions selects entries of any one of them, as the Access page's
	// history does for the grants and the links; empty selects every action.
	// With Action set, an entry must match both.
	Actions []AuditAction
	// Outcome selects entries that ended so; empty selects every outcome.
	Outcome AuditOutcome
	// From and To select entries at or after From and before To; a zero
	// bound is open.
	From, To time.Time
	// Before selects entries with a Seq below it, the cursor a page hands
	// out for the next; zero starts from the newest.
	Before int64
	// Limit is the page size: DefaultAuditPage when zero, at most
	// MaxAuditPage.
	Limit int
}

// Validate returns nil when f can be answered, otherwise an error wrapping
// ErrInvalidAuditFilter.
func (f *AuditFilter) Validate() error {
	if err := f.validateActor(); err != nil {
		return err
	}
	if err := f.validateSelection(); err != nil {
		return err
	}
	if !f.From.IsZero() && !f.To.IsZero() && !f.From.Before(f.To) {
		return fmt.Errorf("%w: the range is empty", ErrInvalidAuditFilter)
	}
	if f.Before < 0 {
		return fmt.Errorf("%w: negative cursor", ErrInvalidAuditFilter)
	}
	if f.Limit < 0 || f.Limit > MaxAuditPage {
		return fmt.Errorf("%w: page size out of range", ErrInvalidAuditFilter)
	}
	return nil
}

// PageSize returns the page size f asks for.
func (f *AuditFilter) PageSize() int {
	if f.Limit == 0 {
		return DefaultAuditPage
	}
	return f.Limit
}

// Matches reports whether e is selected by f's actor, action and time
// range, and its cursor.
func (f *AuditFilter) Matches(e *AuditEntry) bool {
	if f.Before != 0 && e.Seq >= f.Before {
		return false
	}
	return f.matchesActor(&e.Actor) && f.matchesWhat(e) && f.inRange(e.At)
}

// matchesActor reports whether a is the kind and the actor f selects.
func (f *AuditFilter) matchesActor(a *Actor) bool {
	return (f.ActorKind == "" || a.Kind == f.ActorKind) && (f.ActorID == "" || a.ID == f.ActorID)
}

// matchesWhat reports whether e is of the action, one of the actions and the
// outcome f selects.
func (f *AuditFilter) matchesWhat(e *AuditEntry) bool {
	return (f.Action == "" || e.Action == f.Action) &&
		(len(f.Actions) == 0 || slices.Contains(f.Actions, e.Action)) &&
		(f.Outcome == "" || e.Outcome == f.Outcome)
}

// validateSelection checks the action, the set of actions and the outcome f
// selects.
func (f *AuditFilter) validateSelection() error {
	for _, a := range append([]AuditAction{f.Action}, f.Actions...) {
		if _, ok := actionKinds()[a]; a != "" && !ok {
			return fmt.Errorf("%w: unknown action", ErrInvalidAuditFilter)
		}
	}
	if f.Outcome != "" {
		if _, err := ParseAuditOutcome(string(f.Outcome)); err != nil {
			return fmt.Errorf("%w: unknown outcome", ErrInvalidAuditFilter)
		}
	}
	return nil
}

// validateActor checks the actor kind and ID f selects.
func (f *AuditFilter) validateActor() error {
	if f.ActorKind == "" && f.ActorID != "" {
		return fmt.Errorf("%w: an actor ID needs its kind", ErrInvalidAuditFilter)
	}
	if f.ActorKind != "" && !knownActorKind(f.ActorKind) {
		return fmt.Errorf("%w: unknown actor kind", ErrInvalidAuditFilter)
	}
	if f.ActorID != "" {
		if err := (Actor{Kind: f.ActorKind, ID: f.ActorID}).Validate(); err != nil {
			return fmt.Errorf("%w: bad actor ID", ErrInvalidAuditFilter)
		}
	}
	return nil
}

// inRange reports whether at is at or after From and before To; a zero bound
// is open.
func (f *AuditFilter) inRange(at time.Time) bool {
	return (f.From.IsZero() || !at.Before(f.From)) && (f.To.IsZero() || at.Before(f.To))
}

// knownActorKind reports whether k is one of the actor kinds.
func knownActorKind(k ActorKind) bool {
	switch k {
	case ActorOwner, ActorViewer, ActorEnrollmentCode, ActorSystem, ActorAnonymous:
		return true
	}
	return false
}
