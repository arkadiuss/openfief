package portfolio

import (
	"errors"
	"fmt"
)

// ErrInvalidSleeve reports a sleeve without an ID, a name or a known role.
var ErrInvalidSleeve = errors.New("invalid sleeve")

// Role is what a sleeve is for. Every sleeve has exactly one.
type Role int

// The roles. The zero Role is not a role.
const (
	// RoleDefensive marks a sleeve that is never available to agents.
	RoleDefensive Role = iota + 1
	// RoleOffensive marks the sleeve that fiefs are leased from (GR-6).
	RoleOffensive
)

// ParseRole parses "defensive" or "offensive".
func ParseRole(s string) (Role, error) {
	for _, r := range []Role{RoleDefensive, RoleOffensive} {
		if r.String() == s {
			return r, nil
		}
	}
	return 0, fmt.Errorf("%w: unknown role %q", ErrInvalidSleeve, s)
}

// String returns "defensive" or "offensive", or a placeholder for an invalid
// role.
func (r Role) String() string {
	switch r {
	case RoleDefensive:
		return "defensive"
	case RoleOffensive:
		return "offensive"
	}
	return fmt.Sprintf("Role(%d)", int(r))
}

// SleeveID identifies a sleeve. Accounts refer to their sleeve by it.
type SleeveID string

// Sleeve is a part of the portfolio with one role and one kind. Its role
// defaults to its kind's default role and is stored on the sleeve, so that
// renaming a kind never changes what agents may lease from (GR-6).
type Sleeve struct {
	ID   SleeveID
	Name string
	Kind KindKey
	Role Role
	// RoleOverride records the owner's explicit choice of a role other than
	// its kind's default role (D-71). CheckSleeve requires it for exactly
	// that case and refuses it otherwise.
	RoleOverride bool
}

// Validate reports whether s has an ID, a name and one known role.
func (s Sleeve) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("%w: empty ID", ErrInvalidSleeve)
	}
	if s.Name == "" {
		return fmt.Errorf("%w %s: empty name", ErrInvalidSleeve, s.ID)
	}
	switch s.Role {
	case RoleDefensive, RoleOffensive:
		return nil
	}
	return fmt.Errorf("%w %s: %s", ErrInvalidSleeve, s.ID, s.Role)
}
