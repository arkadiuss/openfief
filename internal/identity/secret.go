package identity

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// formatSecret returns prefix, the ID, "_" and key in unpadded base64url:
// the form of every bearer secret the identity context issues (script
// tokens, sessions, enrollment codes).
func formatSecret(prefix, id string, key []byte) string {
	return prefix + id + secretSep + base64.RawURLEncoding.EncodeToString(key)
}

// parseSecret splits a secret written by formatSecret with prefix into its
// ID and key, failing with an error wrapping malformed. The ID must be
// IDBytes bytes in lowercase hex and the key KeyBytes bytes. Its errors
// never contain the secret.
func parseSecret(prefix, secret string, malformed error) (secretParts, error) {
	rest, ok := strings.CutPrefix(secret, prefix)
	if !ok {
		return secretParts{}, fmt.Errorf("%w: no %s prefix", malformed, prefix)
	}
	id, keyText, ok := strings.Cut(rest, secretSep)
	if !ok {
		return secretParts{}, fmt.Errorf("%w: no key", malformed)
	}
	if !isHexID(id) {
		return secretParts{}, fmt.Errorf("%w: bad ID", malformed)
	}
	key, err := base64.RawURLEncoding.DecodeString(keyText)
	if err != nil || len(key) != KeyBytes {
		return secretParts{}, fmt.Errorf("%w: want a %d-byte base64url key", malformed, KeyBytes)
	}
	return secretParts{id: id, key: key}, nil
}

// secretParts are a parsed secret's ID and key.
type secretParts struct {
	id  string
	key []byte
}

// isHexID reports whether s is IDBytes bytes in lowercase hex.
func isHexID(s string) bool {
	if len(s) != hex.EncodedLen(IDBytes) || strings.ToLower(s) != s {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}
