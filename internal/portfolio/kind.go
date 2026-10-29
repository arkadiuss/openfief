package portfolio

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidKind reports a sleeve kind without a well-formed key, a name or
// a known default role.
var ErrInvalidKind = errors.New("invalid sleeve kind")

// ErrUnknownKind reports a sleeve whose kind is not among the stored kinds.
var ErrUnknownKind = errors.New("unknown sleeve kind")

// ErrKindRoleChange reports an attempt to change a stored kind's default
// role. A kind can be renamed; its default role never changes.
var ErrKindRoleChange = errors.New("a sleeve kind's default role cannot change")

// ErrRoleOverride reports a sleeve whose role differs from its kind's
// default role in a way the owner did not explicitly ask for.
var ErrRoleOverride = errors.New("sleeve role does not match its kind")

// maxKindKeyLen bounds a kind key, which appears in URLs and the read API.
const maxKindKeyLen = 64

// KindKey identifies a sleeve kind: lowercase letters and digits in words
// joined by single hyphens, such as "agent-pool".
type KindKey string

// The built-in kinds' keys.
const (
	KindEmergencyFund KindKey = "emergency-fund"
	KindBonds         KindKey = "bonds"
	KindRetirement    KindKey = "retirement"
	KindCore          KindKey = "core"
	KindSatellite     KindKey = "satellite"
	KindAgentPool     KindKey = "agent-pool"
)

// Validate reports whether k is a well-formed key: 1 to 64 characters of
// lowercase ASCII letters and digits, in words joined by single hyphens.
func (k KindKey) Validate() error {
	s := string(k)
	if s == "" || len(s) > maxKindKeyLen {
		return fmt.Errorf("%w: key %q must have 1 to %d characters", ErrInvalidKind, s, maxKindKeyLen)
	}
	for word := range strings.SplitSeq(s, "-") {
		if word == "" {
			return fmt.Errorf("%w: key %q has an empty word", ErrInvalidKind, s)
		}
		for _, c := range word {
			if (c < 'a' || c > 'z') && (c < '0' || c > '9') {
				return fmt.Errorf("%w: key %q may hold only a-z, 0-9 and hyphens", ErrInvalidKind, s)
			}
		}
	}
	return nil
}

// Kind is a sleeve kind: a named category of sleeve with the role a sleeve
// of the kind gets by default. The built-in kinds come from DefaultKinds;
// the owner can add more and rename any of them.
type Kind struct {
	Key  KindKey
	Name string
	// DefaultRole is the role a new sleeve of this kind gets. It is fixed
	// when the kind is created.
	DefaultRole Role
}

// Validate reports whether k has a well-formed key, a name and a known
// default role.
func (k Kind) Validate() error {
	if err := k.Key.Validate(); err != nil {
		return err
	}
	if k.Name == "" {
		return fmt.Errorf("%w %s: empty name", ErrInvalidKind, k.Key)
	}
	switch k.DefaultRole {
	case RoleDefensive, RoleOffensive:
		return nil
	}
	return fmt.Errorf("%w %s: default role %s", ErrInvalidKind, k.Key, k.DefaultRole)
}

// DefaultKinds returns the built-in kinds, in key order. Only agent-pool and
// satellite are offensive: agent-pool is what fiefs are leased from (GR-6),
// and satellite holds the owner's own higher-risk positions.
func DefaultKinds() []Kind {
	return []Kind{
		{Key: KindAgentPool, Name: "Agent pool", DefaultRole: RoleOffensive},
		{Key: KindBonds, Name: "Bonds", DefaultRole: RoleDefensive},
		{Key: KindCore, Name: "Core", DefaultRole: RoleDefensive},
		{Key: KindEmergencyFund, Name: "Emergency fund", DefaultRole: RoleDefensive},
		{Key: KindRetirement, Name: "Retirement", DefaultRole: RoleDefensive},
		{Key: KindSatellite, Name: "Satellite", DefaultRole: RoleOffensive},
	}
}

// CheckKind validates k as the owner saves it over stored, the kinds stored
// so far: a new key adds a kind, and an existing key renames it. It fails
// with ErrKindRoleChange when k would change a stored kind's default role,
// because the sleeves of that kind were checked against the old one.
func CheckKind(stored []Kind, k Kind) error {
	if err := k.Validate(); err != nil {
		return err
	}
	old, ok := findKind(stored, k.Key)
	if ok && old.DefaultRole != k.DefaultRole {
		return fmt.Errorf("%w: %s is %s", ErrKindRoleChange, k.Key, old.DefaultRole)
	}
	return nil
}

// findKind returns the kind with key among kinds.
func findKind(kinds []Kind, key KindKey) (Kind, bool) {
	for _, k := range kinds {
		if k.Key == key {
			return k, true
		}
	}
	return Kind{}, false
}

// CheckSleeve validates s against kinds, the stored kinds, and returns it
// as it is stored: with its kind's default role when s has no role. The
// sleeve's kind must be among kinds. Its role may differ from the kind's
// default role only with RoleOverride set: the owner's explicit choice
// (D-71), whichever way it goes, so that no capital moves into or out of
// agents' reach (GR-6) without the owner asking for exactly that.
// RoleOverride on a sleeve whose role is its kind's is an error, so it
// always means the one thing.
func CheckSleeve(kinds []Kind, s Sleeve) (Sleeve, error) {
	k, ok := findKind(kinds, s.Kind)
	if s.Kind == "" || !ok {
		return Sleeve{}, fmt.Errorf("%w %q for sleeve %s", ErrUnknownKind, s.Kind, s.ID)
	}
	if s.Role == 0 {
		s.Role = k.DefaultRole
	}
	if err := s.Validate(); err != nil {
		return Sleeve{}, err
	}
	differs := s.Role != k.DefaultRole
	switch {
	case differs && !s.RoleOverride:
		return Sleeve{}, fmt.Errorf("%w: sleeve %s is %s but kind %s is %s; a role other than the kind's needs the owner's explicit choice",
			ErrRoleOverride, s.ID, s.Role, s.Kind, k.DefaultRole)
	case !differs && s.RoleOverride:
		return Sleeve{}, fmt.Errorf("%w: sleeve %s has a role override but its role is its kind's", ErrRoleOverride, s.ID)
	}
	return s, nil
}

// DefaultSleeves returns the sleeves an install starts with, in ID order:
// Bonds, Core and Satellite, each with its kind's default role.
func DefaultSleeves() []Sleeve {
	return []Sleeve{
		{ID: "bonds", Name: "Bonds", Kind: KindBonds, Role: RoleDefensive},
		{ID: "core", Name: "Core", Kind: KindCore, Role: RoleDefensive},
		{ID: "satellite", Name: "Satellite", Kind: KindSatellite, Role: RoleOffensive},
	}
}

// MissingDefaults returns the default kinds that are not stored yet, by
// key, and, when no sleeve is stored at all, the default sleeves. Saving
// them gives a new install every built-in kind and its three starting
// sleeves (D-71). A default kind the owner renamed is stored and so is
// left alone; once any sleeve is stored the install has started, so a
// default sleeve the owner changed, or one the demo dataset replaced with
// its own, is never added again.
func MissingDefaults(kinds []Kind, sleeves []Sleeve) ([]Kind, []Sleeve) {
	var missingKinds []Kind
	for _, k := range DefaultKinds() {
		if _, ok := findKind(kinds, k.Key); !ok {
			missingKinds = append(missingKinds, k)
		}
	}
	if len(sleeves) > 0 {
		return missingKinds, nil
	}
	return missingKinds, DefaultSleeves()
}

// KindBySleeve maps each sleeve's ID to its kind, so that an allocation
// grouped by sleeve kind (D-30) can find an account's kind from its sleeve.
func KindBySleeve(sleeves []Sleeve) map[SleeveID]KindKey {
	out := make(map[SleeveID]KindKey, len(sleeves))
	for _, s := range sleeves {
		out[s.ID] = s.Kind
	}
	return out
}
