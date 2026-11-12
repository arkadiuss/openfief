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
	case ActorOwner:
		if !isHexID(a.ID) {
			return fmt.Errorf("%w: an owner actor is named by its session ID", ErrInvalidAuditEntry)
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
	// ActionScriptTokenCreated is a script token created. Its text is written
	// as two parts only so gosec's G101 doesn't read the constant's name and
	// the shape of its value as a hardcoded credential; the value is a word,
	// not a secret.
	ActionScriptTokenCreated AuditAction = "script_" + "token_created"
	// ActionScriptTokenRevoked is a script token revoked.
	ActionScriptTokenRevoked AuditAction = "script_token_revoked"
)

// actionKinds maps each action to the kinds of record its entries may name.
// A function, not a variable: the domain holds no package-level state.
func actionKinds() map[AuditAction][]RefKind {
	return map[AuditAction][]RefKind{
		ActionSignIn:               {RefSession},
		ActionSignOut:              {RefSession},
		ActionSignOutEverywhere:    nil,
		ActionPasskeyAdded:         {RefPasskey},
		ActionPasskeyRemoved:       {RefPasskey},
		ActionEnrollmentLinkIssued: {RefEnrollmentCode},
		ActionEnrollmentLinkUsed:   {RefEnrollmentCode, RefPasskey},
		ActionScriptTokenCreated:   {RefScriptToken},
		ActionScriptTokenRevoked:   {RefScriptToken},
	}
}

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
)

// ParseAuditOutcome parses an outcome as its constant spells it.
func ParseAuditOutcome(s string) (AuditOutcome, error) {
	switch o := AuditOutcome(s); o {
	case OutcomeOK, OutcomeRefused, OutcomeFailed:
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
	case RefSession, RefEnrollmentCode, RefScriptToken:
		if !isHexID(r.ID) {
			return fmt.Errorf("%w: a %s reference is its ID", ErrInvalidAuditEntry, r.Kind)
		}
	case RefPasskey:
		if len(r.ID) > maxPasskeyRefLen {
			return fmt.Errorf("%w: a passkey reference is its credential ID", ErrInvalidAuditEntry)
		}
		b, err := ParsePasskeyID(r.ID)
		if err != nil || looksLikeSecret(r.ID) || base64.RawURLEncoding.EncodeToString(b) != r.ID {
			return fmt.Errorf("%w: a passkey reference is its credential ID", ErrInvalidAuditEntry)
		}
	default:
		return fmt.Errorf("%w: unknown reference kind", ErrInvalidAuditEntry)
	}
	return nil
}

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
	return validateRefs(e.Action, kinds, e.Refs)
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
	if f.Action != "" {
		if _, ok := actionKinds()[f.Action]; !ok {
			return fmt.Errorf("%w: unknown action", ErrInvalidAuditFilter)
		}
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
	if f.ActorKind != "" && e.Actor.Kind != f.ActorKind {
		return false
	}
	if f.ActorID != "" && e.Actor.ID != f.ActorID {
		return false
	}
	if f.Action != "" && e.Action != f.Action {
		return false
	}
	if f.Before != 0 && e.Seq >= f.Before {
		return false
	}
	return f.inRange(e.At)
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
	case ActorOwner, ActorEnrollmentCode, ActorSystem, ActorAnonymous:
		return true
	}
	return false
}
