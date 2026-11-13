package spicedb

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	// ErrInvalidID reports an empty ID, which SpiceDB can't take as an object.
	ErrInvalidID = errors.New("empty ID")
	// errBadEscape reports an object ID whose escape is cut short or not hex.
	errBadEscape = errors.New("bad escape")
)

// The object types and relations of authz/schema.zed.
const (
	typePerson      = "person"
	typeScriptToken = "script_token"
	typeViewPlugin  = "view_plugin"
	typeDemoSession = "demo_session"
	typeHousehold   = "household"
	typeAccount     = "account"
	typeSleeve      = "sleeve"

	relOwner     = "owner"
	relReader    = "reader"
	relViewer    = "viewer"
	relHousehold = "household"
	relSleeve    = "sleeve"

	// householdID is the one household an installation serves (D-65).
	householdID = "home"
	// ownerID is the owner as a person: until viewers have passkeys of their
	// own (D-65), the one person there is.
	ownerID = "owner"
	// wildcardID is SpiceDB's wildcard object ID: every object of a type.
	wildcardID = "*"
)

// ObjectID returns id as a SpiceDB object ID, which takes only letters,
// digits and /_|-=+. A byte it doesn't take, and "=", becomes "=" and two
// lowercase hex digits, so two IDs never share an object and an ID that is
// already plain, such as "brokerage", is unchanged. An empty ID fails with
// ErrInvalidID.
func ObjectID(id string) (string, error) {
	if id == "" {
		return "", ErrInvalidID
	}
	var b strings.Builder
	for i := range len(id) {
		c := id[i]
		if plain(c) {
			b.WriteByte(c)
			continue
		}
		fmt.Fprintf(&b, "=%02x", c)
	}
	return b.String(), nil
}

// plain reports whether SpiceDB takes c in an object ID as itself, which
// ObjectID does for all but "=".
func plain(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	}
	return strings.IndexByte("/_|-+", c) >= 0
}

// idOf returns the ID that ObjectID turned into object; a malformed escape
// fails.
func idOf(object string) (string, error) {
	var b strings.Builder
	for i := 0; i < len(object); i++ {
		if object[i] != '=' {
			b.WriteByte(object[i])
			continue
		}
		if i+3 > len(object) {
			return "", fmt.Errorf("object ID %q: %w", object, errBadEscape)
		}
		n, err := strconv.ParseUint(object[i+1:i+3], 16, 8)
		if err != nil {
			return "", fmt.Errorf("object ID %q: %w", object, errBadEscape)
		}
		b.WriteByte(byte(n))
		i += 2
	}
	return b.String(), nil
}
