package identity

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInvalidGrant reports a grant that names no holder or target the owner
// could give or take access of. Validation returns it wrapped, and a message
// never quotes the rejected ID, which could be a secret that landed in the
// wrong field (GR-12).
var ErrInvalidGrant = errors.New("invalid grant")

// HouseholdID is the ID of the one household an installation serves (D-65):
// the target of a household grant.
const HouseholdID = "home"

// GrantHolderKind is what kind of party holds a grant.
type GrantHolderKind string

// The kinds of holder a grant names, a closed set (D-209). The demo
// authenticator's sessions hold the household grant by the model itself
// (D-68), never as a stored grant, and the owner needs none.
const (
	// GrantHolderPerson is a viewer, a person with a passkey of their own
	// (D-65), identified by their ID.
	GrantHolderPerson GrantHolderKind = "person"
	// GrantHolderScriptToken is a script token (D-27), by its ID.
	GrantHolderScriptToken GrantHolderKind = "script_token"
	// GrantHolderViewPlugin is a view plugin, by its instance name (D-27).
	GrantHolderViewPlugin GrantHolderKind = "view_plugin"
)

// GrantTargetKind is what a grant lets its holder view.
type GrantTargetKind string

// The kinds of target of a grant, a closed set (D-209).
const (
	// GrantTargetAccount is one account, by its ID.
	GrantTargetAccount GrantTargetKind = "account"
	// GrantTargetSleeve is one sleeve, by its ID: every account in it, those
	// placed there later too.
	GrantTargetSleeve GrantTargetKind = "sleeve"
	// GrantTargetHousehold is the whole household: every account, present and
	// future.
	GrantTargetHousehold GrantTargetKind = "household"
)

// GrantHolder is who holds a grant: a kind and an ID among holders of that
// kind.
type GrantHolder struct {
	Kind GrantHolderKind
	ID   string
}

// GrantTarget is what a grant covers: a kind and an ID among targets of that
// kind. A household grant's ID is HouseholdID.
type GrantTarget struct {
	Kind GrantTargetKind
	ID   string
}

// Grant is the owner's permission for a person, a script token or a view
// plugin to view an account, a sleeve or the whole household (D-209). A
// grant only ever lets its holder view: nothing here, or in SpiceDB's model,
// turns one into a permission to change. Revoking keeps the grant, with the
// time it ended, so the history of who could view what stays.
type Grant struct {
	Holder    GrantHolder
	Target    GrantTarget
	CreatedAt time.Time
	// RevokedAt is when the grant ended; zero while it is in force.
	RevokedAt time.Time
}

// The longest a plugin instance's name and an account's or a sleeve's ID are.
const (
	maxPluginInstanceLen = 63
	maxRecordIDLen       = 128
)

// isPluginInstanceName reports whether s is a Kubernetes DNS label, which a
// plugin instance's name is: lowercase letters, digits and dashes, starting
// and ending with a letter or digit. A grant names a holder by such an ID
// alone, so what reaches the audit log is an identifier and nothing a person
// typed freely.
func isPluginInstanceName(s string) bool {
	if s == "" || len(s) > maxPluginInstanceLen || strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return false
	}
	return strings.IndexFunc(s, func(r rune) bool { return !isLowerAlnum(r) && r != '-' }) < 0
}

// isRecordID reports whether s is an account's or a sleeve's ID: the key made
// from its name (portfolio.KeyFromName), or an ID a seed or an import gave
// it, with letters, digits, '.', '_' and '-', starting with a letter or
// digit.
func isRecordID(s string) bool {
	if s == "" || len(s) > maxRecordIDLen || !isAlnum(rune(s[0])) {
		return false
	}
	return strings.IndexFunc(s, func(r rune) bool { return !isAlnum(r) && r != '.' && r != '_' && r != '-' }) < 0
}

// isLowerAlnum reports whether r is an ASCII lowercase letter or a digit.
func isLowerAlnum(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') }

// isAlnum reports whether r is an ASCII letter or a digit.
func isAlnum(r rune) bool { return isLowerAlnum(r) || (r >= 'A' && r <= 'Z') }

// NewGrant returns the grant of target to holder, made at at, or an error
// wrapping ErrInvalidGrant.
func NewGrant(holder GrantHolder, target GrantTarget, at time.Time) (Grant, error) {
	g := Grant{Holder: holder, Target: target, CreatedAt: at.UTC()}
	if err := g.Validate(); err != nil {
		return Grant{}, err
	}
	return g, nil
}

// Validate returns nil when g names a holder and a target by IDs of their
// kinds' shapes, and a revocation, when there is one, that is not before the
// grant. Otherwise it wraps ErrInvalidGrant, without quoting an ID.
func (g *Grant) Validate() error {
	if err := g.Holder.Validate(); err != nil {
		return err
	}
	if err := g.Target.Validate(); err != nil {
		return err
	}
	if g.CreatedAt.IsZero() {
		return fmt.Errorf("%w: no time", ErrInvalidGrant)
	}
	if !g.RevokedAt.IsZero() && g.RevokedAt.Before(g.CreatedAt) {
		return fmt.Errorf("%w: revoked before it was made", ErrInvalidGrant)
	}
	return nil
}

// Active reports whether g is in force: not revoked.
func (g *Grant) Active() bool { return g.RevokedAt.IsZero() }

// Revoked returns g, ended at at: the time of a revocation never precedes the
// grant's own.
func (g *Grant) Revoked(at time.Time) Grant {
	out := *g
	at = at.UTC()
	if at.Before(g.CreatedAt) {
		at = g.CreatedAt
	}
	out.RevokedAt = at
	return out
}

// Validate returns nil when h's ID has the shape its kind's IDs have.
func (h GrantHolder) Validate() error {
	switch h.Kind {
	case GrantHolderPerson, GrantHolderScriptToken:
		if !isHexID(h.ID) {
			return fmt.Errorf("%w: a %s holder is its ID", ErrInvalidGrant, h.Kind)
		}
	case GrantHolderViewPlugin:
		if !isPluginInstanceName(h.ID) {
			return fmt.Errorf("%w: a view plugin holder is its instance name", ErrInvalidGrant)
		}
	default:
		return fmt.Errorf("%w: unknown holder kind", ErrInvalidGrant)
	}
	return nil
}

// Validate returns nil when t's ID has the shape its kind's IDs have: the
// household's is HouseholdID.
func (t GrantTarget) Validate() error {
	switch t.Kind {
	case GrantTargetAccount, GrantTargetSleeve:
		if !isRecordID(t.ID) || looksLikeSecret(t.ID) {
			return fmt.Errorf("%w: an account or sleeve target is its ID", ErrInvalidGrant)
		}
	case GrantTargetHousehold:
		if t.ID != HouseholdID {
			return fmt.Errorf("%w: the household's ID is %q", ErrInvalidGrant, HouseholdID)
		}
	default:
		return fmt.Errorf("%w: unknown target kind", ErrInvalidGrant)
	}
	return nil
}

// Household returns the target of a household grant.
func Household() GrantTarget { return GrantTarget{Kind: GrantTargetHousehold, ID: HouseholdID} }

// Ref names h in an audit entry, by kind and ID.
func (h GrantHolder) Ref() Ref {
	switch h.Kind {
	case GrantHolderPerson:
		return Ref{Kind: RefPerson, ID: h.ID}
	case GrantHolderViewPlugin:
		return Ref{Kind: RefViewPlugin, ID: h.ID}
	case GrantHolderScriptToken:
	}
	return Ref{Kind: RefScriptToken, ID: h.ID}
}

// Ref names t in an audit entry, by kind and ID.
func (t GrantTarget) Ref() Ref {
	switch t.Kind {
	case GrantTargetSleeve:
		return Ref{Kind: RefSleeve, ID: t.ID}
	case GrantTargetHousehold:
		return Ref{Kind: RefHousehold, ID: t.ID}
	case GrantTargetAccount:
	}
	return Ref{Kind: RefAccount, ID: t.ID}
}

// GrantMatch selects the grants of one holder, of one target, or of both.
// A zero side selects every holder or target. The zero GrantMatch selects
// every grant.
type GrantMatch struct {
	Holder GrantHolder
	Target GrantTarget
}

// Matches reports whether g has the holder and the target m names, a zero
// side naming any.
func (m GrantMatch) Matches(g *Grant) bool {
	return (m.Holder == GrantHolder{} || g.Holder == m.Holder) && (m.Target == GrantTarget{} || g.Target == m.Target)
}

// Validate returns nil when each side m names is valid.
func (m GrantMatch) Validate() error {
	if m.Holder != (GrantHolder{}) {
		if err := m.Holder.Validate(); err != nil {
			return err
		}
	}
	if m.Target != (GrantTarget{}) {
		if err := m.Target.Validate(); err != nil {
			return err
		}
	}
	return nil
}
