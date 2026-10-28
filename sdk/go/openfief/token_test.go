package openfief_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// writeToken writes token to the file at path, as the kubelet does when it
// rotates a projected token.
func writeToken(t *testing.T, path, token string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(token), 0o600); err != nil {
		t.Fatalf("write token: %v", err)
	}
}

func TestTokenFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
		noFile  bool
		want    string
		wantErr error
	}{
		{name: "token", content: "synthetic-token-1", want: "synthetic-token-1"},
		{name: "trailing newline is trimmed", content: "synthetic-token-2\n", want: "synthetic-token-2"},
		{name: "empty file", content: " \n", wantErr: openfief.ErrEmptyToken},
		{name: "no file", noFile: true, wantErr: fs.ErrNotExist},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "token")
			if !tt.noFile {
				writeToken(t, path, tt.content)
			}
			got, err := openfief.TokenFile{Path: path}.Token()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Token() error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Token() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTokenFileRereadsRotatedToken(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "token")
	src := openfief.TokenFile{Path: path}
	for _, token := range []string{"synthetic-token-1", "synthetic-token-2"} {
		writeToken(t, path, token)
		got, err := src.Token()
		if err != nil {
			t.Fatalf("Token(): %v", err)
		}
		if got != token {
			t.Errorf("Token() = %q, want the rotated %q", got, token)
		}
	}
}
