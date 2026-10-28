package plugins_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// Synthetic discovery documents in the shape the API server serves them;
// the key values are placeholders, not real keys.
const (
	openIDConfig = `{"issuer":"https://kubernetes.default.svc.cluster.local","jwks_uri":"https://192.0.2.1:6443/openid/v1/jwks",` +
		`"response_types_supported":["id_token"],"subject_types_supported":["public"],"id_token_signing_alg_values_supported":["RS256"]}`
	rsaKey = `{"use":"sig","kty":"RSA","kid":"synthetic-rsa","alg":"RS256","n":"c3ludGhldGlj","e":"AQAB"}`
	ecKey  = `{"use":"sig","kty":"EC","kid":"synthetic-ec","crv":"P-256","alg":"ES256","x":"c3ludGhldGlj","y":"c3ludGhldGlj"}`
)

func TestParseSigningKeys(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  string
		jwks    string
		want    plugins.SigningKeys
		wantErr error
	}{
		{
			name: "RSA and EC keys, compacted", config: openIDConfig,
			jwks: "{\n  \"keys\": [" + rsaKey + ",\n " + ecKey + "]\n}\n",
			want: plugins.SigningKeys{
				Issuer: "https://kubernetes.default.svc.cluster.local",
				JWKS:   []byte(`{"keys":[` + rsaKey + `,` + ecKey + `]}`),
			},
		},
		{name: "no issuer", config: `{"jwks_uri":"https://192.0.2.1/openid/v1/jwks"}`, jwks: `{"keys":[` + rsaKey + `]}`, wantErr: plugins.ErrInvalidSigningKeys},
		{name: "issuer with a newline", config: `{"issuer":"https://a.example\nb"}`, jwks: `{"keys":[` + rsaKey + `]}`, wantErr: plugins.ErrInvalidSigningKeys},
		{name: "configuration not JSON", config: `<html>`, jwks: `{"keys":[` + rsaKey + `]}`, wantErr: plugins.ErrInvalidSigningKeys},
		{name: "key set not JSON", config: openIDConfig, jwks: `{"keys":`, wantErr: plugins.ErrInvalidSigningKeys},
		{name: "empty key set", config: openIDConfig, jwks: `{"keys":[]}`, wantErr: plugins.ErrInvalidSigningKeys},
		{name: "symmetric key", config: openIDConfig, jwks: `{"keys":[{"kty":"oct","kid":"s","k":"c3ludGhldGlj"}]}`, wantErr: plugins.ErrInvalidSigningKeys},
		{
			name: "private RSA key", config: openIDConfig,
			jwks: `{"keys":[{"kty":"RSA","kid":"r","n":"c3ludGhldGlj","e":"AQAB","d":"c3ludGhldGlj"}]}`, wantErr: plugins.ErrInvalidSigningKeys,
		},
		{name: "key without an ID", config: openIDConfig, jwks: `{"keys":[{"kty":"RSA","n":"c3ludGhldGlj","e":"AQAB"}]}`, wantErr: plugins.ErrInvalidSigningKeys},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := plugins.ParseSigningKeys([]byte(tt.config), []byte(tt.jwks))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParseSigningKeys() error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ParseSigningKeys() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSigningKeysData(t *testing.T) {
	t.Parallel()
	k := plugins.SigningKeys{Issuer: "https://issuer.example", JWKS: []byte(`{"keys":[` + rsaKey + `]}`)}
	want := map[string]string{"jwks.json": `{"keys":[` + rsaKey + `]}`, "issuer": "https://issuer.example"}
	if diff := cmp.Diff(want, k.Data()); diff != "" {
		t.Errorf("Data() (-want +got):\n%s", diff)
	}
}
