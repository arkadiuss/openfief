package identity_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
)

func TestPlatformCredential(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		random  []byte
		want    string
		wantErr error
	}{
		{name: "32 bytes are 64 lowercase hex digits", random: bytes.Repeat([]byte{0xab}, 32), want: strings.Repeat("ab", 32)},
		{name: "more bytes keep every one", random: append(bytes.Repeat([]byte{0x01}, 32), 0xff), want: strings.Repeat("01", 32) + "ff"},
		{name: "one byte short is refused", random: make([]byte, 31), wantErr: identity.ErrShortCredential},
		{name: "no bytes is refused", wantErr: identity.ErrShortCredential},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := identity.PlatformCredential(tc.random)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("credential = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPlatformSecrets(t *testing.T) {
	t.Parallel()
	want := []identity.PlatformSecret{
		{
			Name: "openfief-spicedb-key",
			Type: "Opaque",
			Labels: map[string]string{
				"app.kubernetes.io/name": "openfief-spicedb", "app.kubernetes.io/part-of": "openfief",
			},
			Data: map[string]string{"key": "K"},
		},
		{
			Name: "openfief-db-spicedb",
			Type: "kubernetes.io/basic-auth",
			Labels: map[string]string{
				"app.kubernetes.io/name": "openfief-spicedb", "app.kubernetes.io/part-of": "openfief", "cnpg.io/reload": "true",
			},
			Data: map[string]string{"username": "spicedb", "password": "P"},
		},
	}
	if diff := cmp.Diff(want, identity.PlatformSecrets("K", "P")); diff != "" {
		t.Errorf("PlatformSecrets mismatch (-want +got):\n%s", diff)
	}
}
