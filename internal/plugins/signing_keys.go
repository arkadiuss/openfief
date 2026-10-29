package plugins

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidSigningKeys reports a cluster issuer or key set that can't be
// published for plugins: not JSON, no issuer, no public signing key, or a
// key with private material.
var ErrInvalidSigningKeys = errors.New("invalid ServiceAccount signing keys")

// privateKeyMembers are the JSON Web Key members (RFC 7518 section 6) that
// carry private or symmetric key material; a published key never has one.
func privateKeyMembers() []string { return []string{"d", "p", "q", "dp", "dq", "qi", "oth", "k"} }

// SigningKeys are the cluster's ServiceAccount token issuer and the public
// keys that verify its tokens, as the operator publishes them for plugins in
// SigningKeysConfigMap (D-59). JWKS is a compact JSON Web Key Set.
type SigningKeys struct {
	Issuer string
	JWKS   []byte
}

// ParseSigningKeys reads the API server's OpenID provider configuration
// (/.well-known/openid-configuration) for its issuer, and its key set
// (/openid/v1/jwks). It accepts only a set of RSA and EC keys, each with a
// key ID and no private material, and at least one of them, so a plugin is
// never handed a key it could be fooled with.
func ParseSigningKeys(openIDConfig, jwks []byte) (SigningKeys, error) {
	var cfg struct {
		Issuer string `json:"issuer"`
	}
	if err := json.Unmarshal(openIDConfig, &cfg); err != nil {
		return SigningKeys{}, fmt.Errorf("%w: openid configuration: %w", ErrInvalidSigningKeys, err)
	}
	if cfg.Issuer == "" || strings.ContainsAny(cfg.Issuer, " \t\r\n") {
		return SigningKeys{}, fmt.Errorf("%w: issuer %q", ErrInvalidSigningKeys, cfg.Issuer)
	}
	var set struct {
		Keys []map[string]json.RawMessage `json:"keys"`
	}
	if err := json.Unmarshal(jwks, &set); err != nil {
		return SigningKeys{}, fmt.Errorf("%w: key set: %w", ErrInvalidSigningKeys, err)
	}
	if len(set.Keys) == 0 {
		return SigningKeys{}, fmt.Errorf("%w: the key set is empty", ErrInvalidSigningKeys)
	}
	for i, k := range set.Keys {
		if err := checkPublicKey(k); err != nil {
			return SigningKeys{}, fmt.Errorf("%w: keys[%d]: %w", ErrInvalidSigningKeys, i, err)
		}
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, jwks); err != nil {
		return SigningKeys{}, fmt.Errorf("%w: key set: %w", ErrInvalidSigningKeys, err)
	}
	return SigningKeys{Issuer: cfg.Issuer, JWKS: compact.Bytes()}, nil
}

// Errors of one key in a set.
var (
	errKeyType       = errors.New("key type is not RSA or EC")
	errNoKeyID       = errors.New("no key ID")
	errPrivateMember = errors.New("carries private key material")
)

// checkPublicKey accepts an RSA or EC key with a key ID and no private
// member.
func checkPublicKey(k map[string]json.RawMessage) error {
	var kty, kid string
	if json.Unmarshal(k["kty"], &kty) != nil || (kty != "RSA" && kty != "EC") {
		return errKeyType
	}
	if json.Unmarshal(k["kid"], &kid) != nil || kid == "" {
		return errNoKeyID
	}
	for _, m := range privateKeyMembers() {
		if _, ok := k[m]; ok {
			return fmt.Errorf("%w: %q", errPrivateMember, m)
		}
	}
	return nil
}

// Data returns the keys as SigningKeysConfigMap holds them: the key set
// under SigningKeysFile and the issuer under SigningKeysIssuerFile.
func (k SigningKeys) Data() map[string]string {
	return map[string]string{SigningKeysFile: string(k.JWKS), SigningKeysIssuerFile: k.Issuer}
}
