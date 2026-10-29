package openfief_test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// Synthetic values; no real cluster's.
const (
	testIssuer   = "https://kubernetes.default.svc.cluster.local"
	coreSubject  = "system:serviceaccount:openfief:openfief-core"
	testInstance = "nbp"
)

// testNow is the fixed time every token is checked at.
func testNow() time.Time { return time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC) }

// testKeys are locally generated signing keys, made once per test binary.
type testKeys struct {
	rsa, rsaOther *rsa.PrivateKey
	ec            *ecdsa.PrivateKey
}

func newTestKeys(t *testing.T) testKeys {
	t.Helper()
	rsa1, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	rsa2, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ec key: %v", err)
	}
	return testKeys{rsa: rsa1, rsaOther: rsa2, ec: ec}
}

// writeKeys writes the issuer and a JWKS of the given keys into dir, as the
// operator's ConfigMap mount holds them.
func writeKeys(t *testing.T, dir, issuer string, keys ...jose.JSONWebKey) {
	t.Helper()
	data, err := json.Marshal(jose.JSONWebKeySet{Keys: keys})
	if err != nil {
		t.Fatalf("marshal jwks: %v", err)
	}
	writeFile(t, filepath.Join(dir, openfief.KeysFile), data)
	writeFile(t, filepath.Join(dir, openfief.IssuerFile), []byte(issuer+"\n"))
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func publicJWK(key crypto.PublicKey, kid string) jose.JSONWebKey {
	return jose.JSONWebKey{Key: key, KeyID: kid, Use: "sig"}
}

// claims are a core token's claims, valid at testNow().
func claims() jwt.Claims {
	return jwt.Claims{
		Issuer: testIssuer, Subject: coreSubject, Audience: jwt.Audience{openfief.InstanceAudience(testInstance)},
		IssuedAt: jwt.NewNumericDate(testNow().Add(-time.Minute)), NotBefore: jwt.NewNumericDate(testNow().Add(-time.Minute)),
		Expiry: jwt.NewNumericDate(testNow().Add(10 * time.Minute)),
	}
}

// sign serializes c signed with key under alg, with key ID kid when not empty.
func sign(t *testing.T, alg jose.SignatureAlgorithm, key any, kid string, c *jwt.Claims) string {
	t.Helper()
	opts := (&jose.SignerOptions{}).WithType("JWT")
	if kid != "" {
		opts = opts.WithHeader("kid", kid)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: alg, Key: key}, opts)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}
	token, err := jwt.Signed(signer).Claims(c).Serialize()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return token
}

// unsigned builds a token with alg "none" and the given claims.
func unsigned(t *testing.T, kid string, c *jwt.Claims) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": "none", "kid": kid})
	if err != nil {
		t.Fatalf("header: %v", err)
	}
	payload, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("payload: %v", err)
	}
	enc := base64.RawURLEncoding
	return enc.EncodeToString(header) + "." + enc.EncodeToString(payload) + "."
}

// checkRefusal fails the test unless err is nil when want is, and otherwise
// wraps ErrNotCore and want without containing token.
func checkRefusal(t *testing.T, err, want error, token string) {
	t.Helper()
	switch {
	case want == nil && err != nil:
		t.Fatalf("VerifyCore = %v, want nil", err)
	case want == nil:
	case !errors.Is(err, openfief.ErrNotCore) || !errors.Is(err, want):
		t.Fatalf("VerifyCore = %v, want ErrNotCore and %v", err, want)
	case token != "" && strings.Contains(err.Error(), token):
		t.Errorf("error %q contains the token", err)
	}
}

func newVerifier(dir string) *openfief.JWKSVerifier {
	return &openfief.JWKSVerifier{
		Dir: dir, Audience: openfief.InstanceAudience(testInstance), Subject: coreSubject,
		Now: testNow,
	}
}

func TestJWKSVerifier(t *testing.T) {
	t.Parallel()
	keys := newTestKeys(t)
	with := func(edit func(*jwt.Claims)) *jwt.Claims {
		c := claims()
		edit(&c)
		return &c
	}
	valid := claims()
	tests := []struct {
		name    string
		token   func(t *testing.T) string
		wantErr error
	}{
		{name: "the core's RSA-signed token for this instance", token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsa, "rsa-1", &valid)
		}},
		{name: "the core's ECDSA-signed token for this instance", token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.ES256, keys.ec, "ec-1", &valid)
		}},
		{name: "expired within the leeway", token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsa, "rsa-1", with(func(c *jwt.Claims) { c.Expiry = jwt.NewNumericDate(testNow().Add(-30 * time.Second)) }))
		}},
		{name: "one of several audiences", token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsa, "rsa-1", with(func(c *jwt.Claims) {
				c.Audience = jwt.Audience{"openfief-plugin:other", openfief.InstanceAudience(testInstance)}
			}))
		}},
		{name: "another instance's audience", wantErr: jwt.ErrInvalidAudience, token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsa, "rsa-1", with(func(c *jwt.Claims) { c.Audience = jwt.Audience{"openfief-plugin:nbp2"} }))
		}},
		{name: "a token for the core's own audience", wantErr: jwt.ErrInvalidAudience, token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsa, "rsa-1", with(func(c *jwt.Claims) { c.Audience = jwt.Audience{"openfief-core"} }))
		}},
		{name: "another ServiceAccount", wantErr: jwt.ErrInvalidSubject, token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsa, "rsa-1", with(func(c *jwt.Claims) { c.Subject = "system:serviceaccount:openfief:plugin-other" }))
		}},
		{name: "another issuer", wantErr: jwt.ErrInvalidIssuer, token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsa, "rsa-1", with(func(c *jwt.Claims) { c.Issuer = "https://issuer.example.com" }))
		}},
		{name: "expired", wantErr: jwt.ErrExpired, token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsa, "rsa-1", with(func(c *jwt.Claims) { c.Expiry = jwt.NewNumericDate(testNow().Add(-2 * time.Minute)) }))
		}},
		{name: "not valid yet", wantErr: jwt.ErrNotValidYet, token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsa, "rsa-1", with(func(c *jwt.Claims) { c.NotBefore = jwt.NewNumericDate(testNow().Add(5 * time.Minute)) }))
		}},
		{name: "no expiry", wantErr: openfief.ErrNoExpiry, token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsa, "rsa-1", with(func(c *jwt.Claims) { c.Expiry = nil }))
		}},
		{name: "signed by a key not in the set", wantErr: openfief.ErrUnknownKey, token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsaOther, "rsa-2", &valid)
		}},
		{name: "another key under a known key ID", wantErr: openfief.ErrNotCore, token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsaOther, "rsa-1", &valid)
		}},
		{name: "no key ID", wantErr: openfief.ErrUnknownKey, token: func(t *testing.T) string {
			t.Helper()
			return sign(t, jose.RS256, keys.rsa, "", &valid)
		}},
		{name: "alg none", wantErr: openfief.ErrNotCore, token: func(t *testing.T) string {
			t.Helper()
			return unsigned(t, "rsa-1", &valid)
		}},
		{name: "HMAC keyed with the public key", wantErr: openfief.ErrNotCore, token: func(t *testing.T) string {
			t.Helper()
			pub, err := json.Marshal(publicJWK(&keys.rsa.PublicKey, "rsa-1"))
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			return sign(t, jose.HS256, pub, "rsa-1", &valid)
		}},
		{name: "not a token", wantErr: openfief.ErrNotCore, token: func(*testing.T) string { return "synthetic-not-a-jwt" }},
	}
	dir := t.TempDir()
	writeKeys(t, dir, testIssuer, publicJWK(&keys.rsa.PublicKey, "rsa-1"), publicJWK(&keys.ec.PublicKey, "ec-1"))
	v := newVerifier(dir)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			token := tt.token(t)
			err := v.VerifyCore(t.Context(), token)
			checkRefusal(t, err, tt.wantErr, token)
		})
	}
}

func TestJWKSVerifierKeySet(t *testing.T) {
	t.Parallel()
	keys := newTestKeys(t)
	valid := claims()
	tests := []struct {
		name    string
		write   func(t *testing.T, dir string)
		wantErr error
	}{
		{name: "no files", write: func(*testing.T, string) {}, wantErr: openfief.ErrNoSigningKeys},
		{name: "no issuer", wantErr: openfief.ErrNoSigningKeys, write: func(t *testing.T, dir string) {
			t.Helper()
			writeKeys(t, dir, "  ", publicJWK(&keys.rsa.PublicKey, "rsa-1"))
		}},
		{name: "not JSON", wantErr: openfief.ErrNoSigningKeys, write: func(t *testing.T, dir string) {
			t.Helper()
			writeKeys(t, dir, testIssuer)
			writeFile(t, filepath.Join(dir, openfief.KeysFile), []byte("{"))
		}},
		{name: "an empty set", wantErr: openfief.ErrNoSigningKeys, write: func(t *testing.T, dir string) {
			t.Helper()
			writeKeys(t, dir, testIssuer)
		}},
		{name: "only a symmetric key", wantErr: openfief.ErrNoSigningKeys, write: func(t *testing.T, dir string) {
			t.Helper()
			writeKeys(t, dir, testIssuer, jose.JSONWebKey{Key: []byte("synthetic-shared-secret-32-bytes"), KeyID: "rsa-1"})
		}},
		{name: "only an encryption key", wantErr: openfief.ErrNoSigningKeys, write: func(t *testing.T, dir string) {
			t.Helper()
			writeKeys(t, dir, testIssuer, jose.JSONWebKey{Key: &keys.rsa.PublicKey, KeyID: "rsa-1", Use: "enc"})
		}},
		{name: "a key without an ID", wantErr: openfief.ErrNoSigningKeys, write: func(t *testing.T, dir string) {
			t.Helper()
			writeKeys(t, dir, testIssuer, jose.JSONWebKey{Key: &keys.rsa.PublicKey})
		}},
		{name: "an unknown key type beside a good key", write: func(t *testing.T, dir string) {
			t.Helper()
			good, err := json.Marshal(publicJWK(&keys.rsa.PublicKey, "rsa-1"))
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			writeKeys(t, dir, testIssuer)
			writeFile(t, filepath.Join(dir, openfief.KeysFile), []byte(`{"keys":[{"kty":"synthetic","kid":"x"},`+string(good)+`]}`))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			tt.write(t, dir)
			err := newVerifier(dir).VerifyCore(t.Context(), sign(t, jose.RS256, keys.rsa, "rsa-1", &valid))
			checkRefusal(t, err, tt.wantErr, "")
		})
	}
}

func TestJWKSVerifierRotatedKeys(t *testing.T) {
	t.Parallel()
	keys := newTestKeys(t)
	valid := claims()
	dir := t.TempDir()
	writeKeys(t, dir, testIssuer, publicJWK(&keys.rsa.PublicKey, "rsa-1"))
	v := newVerifier(dir)
	oldToken := sign(t, jose.RS256, keys.rsa, "rsa-1", &valid)
	newToken := sign(t, jose.RS256, keys.rsaOther, "rsa-2", &valid)
	if err := v.VerifyCore(t.Context(), oldToken); err != nil {
		t.Fatalf("before rotation: VerifyCore(old) = %v", err)
	}
	if err := v.VerifyCore(t.Context(), newToken); !errors.Is(err, openfief.ErrUnknownKey) {
		t.Fatalf("before rotation: VerifyCore(new) = %v, want ErrUnknownKey", err)
	}
	writeKeys(t, dir, testIssuer, publicJWK(&keys.rsaOther.PublicKey, "rsa-2"))
	if err := v.VerifyCore(t.Context(), newToken); err != nil {
		t.Fatalf("after rotation: VerifyCore(new) = %v", err)
	}
	if err := v.VerifyCore(t.Context(), oldToken); !errors.Is(err, openfief.ErrUnknownKey) {
		t.Fatalf("after rotation: VerifyCore(old) = %v, want ErrUnknownKey", err)
	}
}

func TestJWKSVerifierWithoutAudience(t *testing.T) {
	t.Parallel()
	keys := newTestKeys(t)
	valid := claims()
	dir := t.TempDir()
	writeKeys(t, dir, testIssuer, publicJWK(&keys.rsa.PublicKey, "rsa-1"))
	v := &openfief.JWKSVerifier{Dir: dir, Subject: coreSubject, Now: testNow}
	if err := v.VerifyCore(t.Context(), sign(t, jose.RS256, keys.rsa, "rsa-1", &valid)); !errors.Is(err, openfief.ErrNotCore) {
		t.Fatalf("VerifyCore without an audience = %v, want ErrNotCore", err)
	}
}

func TestRequireCoreWithJWKSVerifier(t *testing.T) {
	t.Parallel()
	keys := newTestKeys(t)
	valid, other := claims(), claims()
	other.Audience = jwt.Audience{"openfief-plugin:other"}
	dir := t.TempDir()
	writeKeys(t, dir, testIssuer, publicJWK(&keys.ec.PublicKey, "ec-1"))
	h := openfief.RequireCore(newVerifier(dir), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	tests := []struct {
		name string
		auth string
		want int
	}{
		{name: "the core's token", auth: "Bearer " + sign(t, jose.ES256, keys.ec, "ec-1", &valid), want: http.StatusNoContent},
		{name: "another instance's token", auth: "Bearer " + sign(t, jose.ES256, keys.ec, "ec-1", &other), want: http.StatusUnauthorized},
		{name: "no token", want: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", http.NoBody)
			if tt.auth != "" {
				req.Header.Set("Authorization", tt.auth)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

// TestNewCoreVerifier sets the process environment, so it can't run in
// parallel.
func TestNewCoreVerifier(t *testing.T) {
	t.Setenv(openfief.EnvInstance, testInstance)
	t.Setenv(openfief.EnvCoreSubject, coreSubject)
	t.Setenv(openfief.EnvKeysDir, "")
	v, err := openfief.NewCoreVerifier()
	if err != nil {
		t.Fatalf("NewCoreVerifier: %v", err)
	}
	if v.Dir != openfief.DefaultKeysDir || v.Audience != "openfief-plugin:nbp" || v.Subject != coreSubject {
		t.Errorf("NewCoreVerifier = %q, %q, %q; want the default directory, the instance's audience and the core", v.Dir, v.Audience, v.Subject)
	}
	t.Setenv(openfief.EnvKeysDir, "/synthetic/keys")
	if v, err = openfief.NewCoreVerifier(); err != nil || v.Dir != "/synthetic/keys" {
		t.Errorf("NewCoreVerifier with %s = %v, %v; want /synthetic/keys", openfief.EnvKeysDir, v, err)
	}
	t.Setenv(openfief.EnvCoreSubject, "")
	if _, err := openfief.NewCoreVerifier(); !errors.Is(err, openfief.ErrNoCoreSubject) {
		t.Errorf("NewCoreVerifier without a subject = %v, want %v", err, openfief.ErrNoCoreSubject)
	}
	t.Setenv(openfief.EnvInstance, "")
	if _, err := openfief.NewCoreVerifier(); !errors.Is(err, openfief.ErrNoInstance) {
		t.Errorf("NewCoreVerifier without an instance = %v, want %v", err, openfief.ErrNoInstance)
	}
}
