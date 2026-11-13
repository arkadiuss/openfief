package identity

import (
	"encoding/hex"
	"errors"
	"fmt"
)

// The platform's own credentials (docs/DESIGN.md D-216): the ones nothing
// else makes. The first is the preshared key that SpiceDB and the core share
// (D-164); the second is the password of SpiceDB's role in the database,
// which CloudNativePG sets from the Secret (D-14) but generates only for the
// owner of a database it creates.
const (
	// PresharedKeyName is the Secret of SpiceDB's preshared key; its one
	// key is PresharedKeyField.
	PresharedKeyName  = "openfief-spicedb-key"
	PresharedKeyField = "key"
	// RoleAuthName is the kubernetes.io/basic-auth Secret of the PostgreSQL
	// role spicedb, as CloudNativePG's managed roles take it.
	RoleAuthName      = "openfief-db-spicedb"
	RoleUsernameField = "username"
	RolePasswordField = "password"
	// RoleUser is the PostgreSQL role the password is for.
	RoleUser = "spicedb"

	// PlatformCredentialBytes is the entropy of each generated value: 32
	// random bytes, 256 bits.
	PlatformCredentialBytes = 32

	// TypeOpaque and TypeBasicAuth are the Kubernetes Secret
	// types the platform's Secrets have.
	TypeOpaque    = "Opaque"
	TypeBasicAuth = "kubernetes.io/basic-auth"
)

// ErrShortCredential reports random bytes too few for a credential.
var ErrShortCredential = errors.New("not enough random bytes for a credential")

// PlatformSecret is one Secret the platform's credentials Job creates: its
// name, Kubernetes type, labels and string data. It holds a credential, so
// it is never logged or put into an error.
type PlatformSecret struct {
	Name   string
	Type   string
	Labels map[string]string
	Data   map[string]string
}

// PlatformCredential returns random as a credential: lowercase hex, which
// needs no escaping in a connection string or a header. It fails with
// ErrShortCredential below PlatformCredentialBytes bytes.
func PlatformCredential(random []byte) (string, error) {
	if len(random) < PlatformCredentialBytes {
		return "", fmt.Errorf("%w: got %d bytes, want %d", ErrShortCredential, len(random), PlatformCredentialBytes)
	}
	return hex.EncodeToString(random), nil
}

// PlatformSecrets returns the Secrets to create, in order: SpiceDB's
// preshared key, holding presharedKey, and the role's basic-auth Secret,
// holding rolePassword. They are exactly these two (D-216): nothing else is
// created.
func PlatformSecrets(presharedKey, rolePassword string) []PlatformSecret {
	labels := func(extra ...string) map[string]string {
		l := map[string]string{
			"app.kubernetes.io/name":    "openfief-spicedb",
			"app.kubernetes.io/part-of": "openfief",
		}
		for _, k := range extra {
			l[k] = "true"
		}
		return l
	}
	return []PlatformSecret{
		{
			Name:   PresharedKeyName,
			Type:   TypeOpaque,
			Labels: labels(),
			Data:   map[string]string{PresharedKeyField: presharedKey},
		},
		{
			Name: RoleAuthName,
			Type: TypeBasicAuth,
			// cnpg.io/reload makes CloudNativePG watch the Secret for the role's password.
			Labels: labels("cnpg.io/reload"),
			Data:   map[string]string{RoleUsernameField: RoleUser, RolePasswordField: rolePassword},
		},
	}
}
