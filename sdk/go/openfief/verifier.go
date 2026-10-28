package openfief

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

// The environment and files a plugin verifies the core's token with (D-59).
// The operator sets the variables in every plugin and mounts the
// ServiceAccount signing keys it publishes, read-only, at DefaultKeysDir.
const (
	// EnvInstance holds the plugin instance's name (D-48); the core's
	// token must carry InstanceAudience of it.
	EnvInstance = "OPENFIEF_INSTANCE"
	// EnvCoreSubject holds the core's Kubernetes username,
	// "system:serviceaccount:openfief:<core ServiceAccount>": the subject
	// of the core's token.
	EnvCoreSubject = "OPENFIEF_CORE_SUBJECT"
	// EnvKeysDir holds the directory of the cluster's ServiceAccount
	// token-signing keys: KeysFile and IssuerFile.
	EnvKeysDir = "OPENFIEF_SA_JWKS_DIR"
	// DefaultKeysDir is where the operator mounts the keys.
	DefaultKeysDir = "/var/run/secrets/openfief-sa-jwks"
	// KeysFile is the public keys as a JSON Web Key Set (RFC 7517), as the
	// API server serves them at /openid/v1/jwks.
	KeysFile = "jwks.json"
	// IssuerFile is the issuer of the cluster's ServiceAccount tokens, as
	// the API server's /.well-known/openid-configuration names it.
	IssuerFile = "issuer"
)

// coreTokenAlgorithms are the signature algorithms Kubernetes signs
// ServiceAccount tokens with: RSA, or ECDSA on P-256, P-384 or P-521. Any
// other algorithm, "none" and HMAC included, is refused before any key is
// tried.
func coreTokenAlgorithms() []jose.SignatureAlgorithm {
	return []jose.SignatureAlgorithm{jose.RS256, jose.ES256, jose.ES384, jose.ES512}
}

// Verifier errors. Each is wrapped in ErrNotCore when VerifyCore refuses a
// token.
var (
	// ErrNoSigningKeys reports a key directory without a usable public key
	// or issuer.
	ErrNoSigningKeys = errors.New("no ServiceAccount signing keys")
	// ErrNoInstance reports a missing OPENFIEF_INSTANCE.
	ErrNoInstance = errors.New("no plugin instance name (OPENFIEF_INSTANCE is not set)")
	// ErrNoCoreSubject reports a missing OPENFIEF_CORE_SUBJECT.
	ErrNoCoreSubject = errors.New("no core subject (OPENFIEF_CORE_SUBJECT is not set)")
	// ErrUnknownKey reports a token signed with no key in the key set.
	ErrUnknownKey = errors.New("token is not signed with a cluster key")
	// ErrNoExpiry reports a token without an expiry.
	ErrNoExpiry = errors.New("token has no expiry")
)

// JWKSVerifier is the SDK's CoreVerifier (D-59). It accepts a ServiceAccount
// token signed by the cluster, issued by its issuer to the core (Subject),
// with this instance's audience (Audience), and not expired. Plugins can't
// reach the Kubernetes API, so it reads the cluster's public signing keys
// and issuer from files the operator mounts; it reads them on every call
// and parses them again only when they changed, so rotated keys are used as
// soon as the kubelet updates the mount.
//
// Set the fields before the first call and don't change them after.
type JWKSVerifier struct {
	// Dir holds KeysFile and IssuerFile. Required.
	Dir string
	// Audience is InstanceAudience of this instance. Required.
	Audience string
	// Subject is the core's Kubernetes username. Required.
	Subject string
	// Now returns the current time; nil means time.Now.
	Now func() time.Time

	mu     sync.Mutex
	raw    []byte
	keys   keySet
	loaded bool
}

// keySet is the parsed content of the key directory.
type keySet struct {
	issuer string
	keys   []jose.JSONWebKey
}

// NewCoreVerifier returns a JWKSVerifier for this plugin from its
// environment: OPENFIEF_INSTANCE, OPENFIEF_CORE_SUBJECT, and
// OPENFIEF_SA_JWKS_DIR or else DefaultKeysDir.
func NewCoreVerifier() (*JWKSVerifier, error) {
	instance, subject := os.Getenv(EnvInstance), os.Getenv(EnvCoreSubject)
	switch {
	case instance == "":
		return nil, ErrNoInstance
	case subject == "":
		return nil, ErrNoCoreSubject
	}
	dir := os.Getenv(EnvKeysDir)
	if dir == "" {
		dir = DefaultKeysDir
	}
	return &JWKSVerifier{Dir: dir, Audience: InstanceAudience(instance), Subject: subject}, nil
}

// VerifyCore accepts token only if it is the core's token for this instance;
// otherwise it returns an error wrapping ErrNotCore, without the token.
func (v *JWKSVerifier) VerifyCore(_ context.Context, token string) error {
	if err := v.verify(token); err != nil {
		return fmt.Errorf("%w: %w", ErrNotCore, err)
	}
	return nil
}

// verify checks the token's signature, then its claims.
func (v *JWKSVerifier) verify(token string) error {
	if v.Audience == "" || v.Subject == "" {
		return fmt.Errorf("%w: verifier has no audience or subject", ErrNoSigningKeys)
	}
	ks, err := v.load()
	if err != nil {
		return err
	}
	tok, err := jwt.ParseSigned(token, coreTokenAlgorithms())
	if err != nil {
		return fmt.Errorf("parse token: %w", err)
	}
	claims, err := ks.claims(tok)
	if err != nil {
		return err
	}
	if claims.Expiry == nil {
		return ErrNoExpiry
	}
	now := time.Now
	if v.Now != nil {
		now = v.Now
	}
	expected := jwt.Expected{Issuer: ks.issuer, Subject: v.Subject, AnyAudience: jwt.Audience{v.Audience}, Time: now()}
	if err := claims.ValidateWithLeeway(expected, jwt.DefaultLeeway); err != nil {
		return fmt.Errorf("token claims: %w", err)
	}
	return nil
}

// claims verifies tok's signature with the key its header names and returns
// its claims.
func (ks *keySet) claims(tok *jwt.JSONWebToken) (*jwt.Claims, error) {
	if len(tok.Headers) != 1 || tok.Headers[0].KeyID == "" {
		return nil, fmt.Errorf("%w: no key ID", ErrUnknownKey)
	}
	kid := tok.Headers[0].KeyID
	for i := range ks.keys {
		if ks.keys[i].KeyID != kid {
			continue
		}
		var c jwt.Claims
		if err := tok.Claims(ks.keys[i].Key, &c); err != nil {
			return nil, fmt.Errorf("token signature: %w", err)
		}
		return &c, nil
	}
	return nil, fmt.Errorf("%w: key ID %q", ErrUnknownKey, kid)
}

// load reads the key directory and parses it again if it changed.
func (v *JWKSVerifier) load() (keySet, error) {
	keys, keysErr := readInDir(filepath.Join(v.Dir, KeysFile))
	issuer, issuerErr := readInDir(filepath.Join(v.Dir, IssuerFile))
	if err := errors.Join(keysErr, issuerErr); err != nil {
		return keySet{}, fmt.Errorf("%w: %w", ErrNoSigningKeys, err)
	}
	raw := bytes.Join([][]byte{issuer, keys}, []byte{0})
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.loaded && bytes.Equal(raw, v.raw) {
		return v.keys, nil
	}
	ks, err := parseKeySet(strings.TrimSpace(string(issuer)), keys)
	if err != nil {
		return keySet{}, err
	}
	v.raw, v.keys, v.loaded = raw, ks, true
	return ks, nil
}

// parseKeySet keeps the set's public RSA and ECDSA signing keys that have a
// key ID; a key it can't parse, a private or symmetric key, and a key for
// encryption are left out. A set with no key left is an error.
func parseKeySet(issuer string, data []byte) (keySet, error) {
	if issuer == "" {
		return keySet{}, fmt.Errorf("%w: empty issuer", ErrNoSigningKeys)
	}
	var doc struct {
		Keys []json.RawMessage `json:"keys"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return keySet{}, fmt.Errorf("%w: %s: %w", ErrNoSigningKeys, KeysFile, err)
	}
	ks := keySet{issuer: issuer}
	for _, raw := range doc.Keys {
		var k jose.JSONWebKey
		if json.Unmarshal(raw, &k) != nil || !usableKey(&k) {
			continue
		}
		ks.keys = append(ks.keys, k)
	}
	if len(ks.keys) == 0 {
		return keySet{}, fmt.Errorf("%w: %s holds no public signing key", ErrNoSigningKeys, KeysFile)
	}
	return ks, nil
}

// usableKey reports whether k is a valid public RSA or ECDSA key with an ID
// that may sign.
func usableKey(k *jose.JSONWebKey) bool {
	if k.KeyID == "" || (k.Use != "" && k.Use != "sig") || !k.Valid() {
		return false
	}
	switch k.Key.(type) {
	case *rsa.PublicKey, *ecdsa.PublicKey:
		return true
	}
	return false
}
