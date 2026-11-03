package plugins_test

import (
	"errors"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// TestMayMountSecrets: an instance may mount a Secret no other instance
// mounts; one another mounts is refused, naming the holder, unless the
// instance mounts it too and is the older, as when two came to mount it at
// once (D-89, D-182). D-89's names keep valid instances apart, so the
// instances here stand for any two that name one Secret.
func TestMayMountSecrets(t *testing.T) {
	t.Parallel()
	early := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	late := early.Add(time.Hour)
	ibkr := plugins.SecretHolder{Instance: "ibkr", Created: early, Secrets: []string{"ibkr--flex-token"}}
	tests := []struct {
		name     string
		self     plugins.SecretHolder
		mounting []string
		held     []plugins.SecretHolder
		wantErr  error
	}{
		{
			name: "no one mounts it",
			self: plugins.SecretHolder{Instance: "ibkr-flex", Created: late, Secrets: []string{"ibkr--flex-token"}},
		},
		{
			name: "another instance mounts it",
			self: plugins.SecretHolder{Instance: "ibkr-flex", Created: late, Secrets: []string{"ibkr--flex-token"}},
			held: []plugins.SecretHolder{ibkr}, wantErr: plugins.ErrSecretHeld,
		},
		{
			name: "the holder itself", self: ibkr, mounting: ibkr.Secrets, held: []plugins.SecretHolder{ibkr},
		},
		{
			name: "both mount it and self is older", self: ibkr, mounting: ibkr.Secrets,
			held: []plugins.SecretHolder{{Instance: "ibkr-flex", Created: late, Secrets: []string{"ibkr--flex-token"}}},
		},
		{
			name:     "both mount it and self is younger",
			self:     plugins.SecretHolder{Instance: "ibkr-flex", Created: late, Secrets: []string{"ibkr--flex-token"}},
			mounting: []string{"ibkr--flex-token"}, held: []plugins.SecretHolder{ibkr}, wantErr: plugins.ErrSecretHeld,
		},
		{
			name: "other Secrets", self: plugins.SecretHolder{Instance: "eodhd", Created: late, Secrets: []string{"eodhd--api-token"}},
			held: []plugins.SecretHolder{ibkr},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := plugins.MayMountSecrets(tt.self, tt.mounting, tt.held); !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Errorf("MayMountSecrets = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
