package modelrouter

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Errors about references, roles and routes. A message never quotes more
// than the rejected reference, which holds names only.
var (
	// ErrInvalidRef reports a model reference that is not "provider/model".
	ErrInvalidRef = errors.New("invalid model reference")
	// ErrUnknownRole reports a role the router has no route for.
	ErrUnknownRole = errors.New("unknown role")
	// ErrInvalidRoute reports a route with no model, too many, or one twice.
	ErrInvalidRoute = errors.New("invalid route")
)

const (
	// maxProviderLen is the longest provider instance name: a DNS label.
	maxProviderLen = 63
	// maxModelLen is the longest model id.
	maxModelLen = 128
	// MaxRouteModels is the most models a route holds: a primary and its
	// fallbacks.
	MaxRouteModels = 8
	// MaxTierModels is the most models the owner lists for a tier.
	MaxTierModels = 64
	// refSeparator joins a provider and a model in a reference.
	refSeparator = "/"
)

// Ref names a model entry of a provider instance, "provider/model" (D-205):
// the provider is the name of a model-provider plugin instance, a lower-case
// DNS label without "--" (D-89); the model is the entry's id at that
// provider, which may itself hold slashes, such as OpenRouter's
// "openai/gpt-oss-120b".
type Ref struct {
	Provider string
	Model    string
}

// ParseRef parses "provider/model", or returns an error wrapping
// ErrInvalidRef. Only the first slash separates the provider.
func ParseRef(s string) (Ref, error) {
	provider, model, ok := strings.Cut(s, refSeparator)
	r := Ref{Provider: provider, Model: model}
	if !ok || !r.valid() {
		return Ref{}, fmt.Errorf("%w: %q", ErrInvalidRef, s)
	}
	return r, nil
}

// String returns "provider/model".
func (r Ref) String() string { return r.Provider + refSeparator + r.Model }

// IsZero reports whether r names nothing.
func (r Ref) IsZero() bool { return r == Ref{} }

// valid reports whether r's provider is a DNS label without the Secret
// separator and its model is a non-empty id of safe characters.
func (r Ref) valid() bool {
	return validProvider(r.Provider) && r.Model != "" && len(r.Model) <= maxModelLen && !strings.ContainsFunc(r.Model, unsafeModelChar)
}

// validProvider reports whether name is a lower-case DNS label of at most
// maxProviderLen characters without "--".
func validProvider(name string) bool {
	if name == "" || len(name) > maxProviderLen || strings.Contains(name, "--") || name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}
	return !strings.ContainsFunc(name, func(c rune) bool {
		return (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-'
	})
}

// unsafeModelChar reports whether c may not appear in a model id.
func unsafeModelChar(c rune) bool { return !safeModelChar(c) }

// safeModelChar reports whether c may appear in a model id: letters,
// digits and the punctuation providers use in ids and tags.
func safeModelChar(c rune) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	}
	return strings.ContainsRune("._:/@+-", c)
}

// ParseRefs parses each of ss, or returns the first error.
func ParseRefs(ss []string) ([]Ref, error) {
	out := make([]Ref, 0, len(ss))
	for _, s := range ss {
		r, err := ParseRef(s)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

// Role is what the core asks a model to do, with its own primary model and
// fallbacks (D-205).
type Role string

// The roles, a closed set.
const (
	// RoleAsk answers the Ask box's questions beyond its fixed commands
	// (D-203).
	RoleAsk Role = "ask"
	// RoleMaintainer writes the model-portfolio maintainer's notes and
	// proposals (D-204).
	RoleMaintainer Role = "maintainer"
)

// Roles returns every role.
func Roles() []Role { return []Role{RoleAsk, RoleMaintainer} }

// ParseRole parses a role as its constant spells it.
func ParseRole(s string) (Role, error) {
	r := Role(s)
	if !slices.Contains(Roles(), r) {
		return "", fmt.Errorf("%w: %q", ErrUnknownRole, s)
	}
	return r, nil
}

// Feature returns the feature whose toggle gates r: a role runs only when
// its feature is on (D-212).
func (r Role) Feature() Feature {
	if r == RoleMaintainer {
		return FeatureMaintainer
	}
	return FeatureAsk
}

// Route is a role's models in the order they are tried: the primary, then
// the fallbacks (D-205).
type Route struct {
	Role   Role
	Models []Ref
}

// NewRoute returns role's route over models, primary first, or an error
// wrapping ErrUnknownRole or ErrInvalidRoute.
func NewRoute(role Role, models []Ref) (Route, error) {
	if !slices.Contains(Roles(), role) {
		return Route{}, fmt.Errorf("%w: %q", ErrUnknownRole, string(role))
	}
	if len(models) == 0 || len(models) > MaxRouteModels {
		return Route{}, fmt.Errorf("%w: one to %d models", ErrInvalidRoute, MaxRouteModels)
	}
	for i, m := range models {
		if !m.valid() {
			return Route{}, fmt.Errorf("%w: %q", ErrInvalidRef, m.String())
		}
		if slices.Contains(models[:i], m) {
			return Route{}, fmt.Errorf("%w: %q twice", ErrInvalidRoute, m.String())
		}
	}
	return Route{Role: role, Models: slices.Clone(models)}, nil
}
