package app_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

var errAPIDown = errors.New("API server unavailable")

// secretShelf is an in-memory app.SecretCreator: names in exists answer
// "already there"; failOn names a Secret whose creation fails.
type secretShelf struct {
	exists map[string]bool
	failOn string
	made   []identity.PlatformSecret
}

func (s *secretShelf) CreateIfMissing(_ context.Context, secret identity.PlatformSecret) (bool, error) {
	if secret.Name == s.failOn {
		return false, errAPIDown
	}
	if s.exists[secret.Name] {
		return false, nil
	}
	s.made = append(s.made, secret)
	return true, nil
}

// madeValues returns what the shelf was asked to create, as name to the one
// credential each Secret holds.
func (s *secretShelf) madeValues() map[string]string {
	got := map[string]string{}
	for i := range s.made {
		got[s.made[i].Name] = s.made[i].Data[identity.PresharedKeyField] + s.made[i].Data[identity.RolePasswordField]
	}
	return got
}

// twoValues is the random source of 32 bytes of 0x11 and then 32 of 0x22.
func twoValues() *bytes.Reader {
	return bytes.NewReader(append(bytes.Repeat([]byte{0x11}, 32), bytes.Repeat([]byte{0x22}, 32)...))
}

// wantValue is what twoValues becomes for the Secret name: the key's value
// first, the role's second, as hex.
func wantValue(name string) string {
	if name == identity.PresharedKeyName {
		return strings.Repeat("11", 32)
	}
	return strings.Repeat("22", 32)
}

func TestInitPlatformCredentials(t *testing.T) {
	t.Parallel()
	key, role := identity.PresharedKeyName, identity.RoleAuthName
	tests := []struct {
		name         string
		shelf        secretShelf
		wantOutcomes []app.PlatformOutcome
		wantMade     map[string]string
		wantErr      error
	}{
		{
			name:         "both missing, both created",
			wantOutcomes: []app.PlatformOutcome{{Name: key, Created: true}, {Name: role, Created: true}},
			wantMade:     map[string]string{key: wantValue(key), role: wantValue(role)},
		},
		{
			name:         "an owner-fed key is kept",
			shelf:        secretShelf{exists: map[string]bool{key: true}},
			wantOutcomes: []app.PlatformOutcome{{Name: key}, {Name: role, Created: true}},
			wantMade:     map[string]string{role: wantValue(role)},
		},
		{
			name:         "both exist, nothing created",
			shelf:        secretShelf{exists: map[string]bool{key: true, role: true}},
			wantOutcomes: []app.PlatformOutcome{{Name: key}, {Name: role}},
			wantMade:     map[string]string{},
		},
		{
			name:         "a failure stops at the first Secret",
			shelf:        secretShelf{failOn: role},
			wantOutcomes: []app.PlatformOutcome{{Name: key, Created: true}},
			wantMade:     map[string]string{key: wantValue(key)},
			wantErr:      errAPIDown,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			shelf := tc.shelf
			outcomes, err := app.InitPlatformCredentials(t.Context(), &shelf, twoValues())
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.wantOutcomes, outcomes); diff != "" {
				t.Errorf("outcomes (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantMade, shelf.madeValues()); diff != "" {
				t.Errorf("created Secrets and their values (-want +got):\n%s", diff)
			}
			if err != nil && (strings.Contains(err.Error(), wantValue(key)) || strings.Contains(err.Error(), wantValue(role))) {
				t.Errorf("the error holds a credential: %v", err)
			}
		})
	}
}

func TestInitPlatformCredentialsNeedsRandomBytes(t *testing.T) {
	t.Parallel()
	for _, n := range []int{0, 31, 40, 63} {
		shelf := &secretShelf{}
		outcomes, err := app.InitPlatformCredentials(t.Context(), shelf, bytes.NewReader(make([]byte, n)))
		if err == nil {
			t.Errorf("%d random bytes: error = nil, want a failure", n)
		}
		if len(outcomes) != 0 || len(shelf.made) != 0 {
			t.Errorf("%d random bytes: %d outcomes and %d Secrets created, want none", n, len(outcomes), len(shelf.made))
		}
	}
}
