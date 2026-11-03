package identity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
)

// passkeySession is a passkey session issued at sessionAt.
func passkeySession(t *testing.T) identity.Session {
	t.Helper()
	s, err := identity.NewSession("0123456789abcdef", identity.AuthenticatorPasskey, identity.SessionScopeOwner, "localhost", sessionAt(),
		identity.DefaultSessionPolicy())
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return s
}

// TestCheckRecentPasskey checks D-183's window: an action it protects is
// allowed less than five minutes after the session's last passkey check,
// and refused from five minutes on, or with no check at all.
func TestCheckRecentPasskey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		checked time.Duration // after sessionAt; negative: no check
		at      time.Duration
		wantErr error
	}{
		{name: "right after the check", checked: 0, at: 0},
		{name: "just within five minutes", checked: time.Hour, at: time.Hour + identity.PasskeyCheckWindow - time.Second},
		{name: "five minutes after", checked: time.Hour, at: time.Hour + identity.PasskeyCheckWindow, wantErr: identity.ErrPasskeyCheckRequired},
		{name: "long after", checked: 0, at: 6 * time.Hour, wantErr: identity.ErrPasskeyCheckRequired},
		{name: "no check", checked: -1, at: 0, wantErr: identity.ErrPasskeyCheckRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := passkeySession(t)
			if tt.checked >= 0 {
				s.PasskeyCheckedAt = sessionAt().Add(tt.checked)
			}
			if err := s.CheckRecentPasskey(sessionAt().Add(tt.at)); !errors.Is(err, tt.wantErr) {
				t.Errorf("CheckRecentPasskey = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestSessionPasskey checks that signing in with a passkey records a copy
// of it and counts as a check, and that a check never moves back.
func TestSessionPasskey(t *testing.T) {
	t.Parallel()
	s := passkeySession(t)
	id := []byte{1, 2, 3}
	signed, err := s.SignedInWithPasskey(id)
	if err != nil {
		t.Fatalf("SignedInWithPasskey: %v", err)
	}
	want := s
	want.PasskeyID, want.PasskeyCheckedAt = []byte{1, 2, 3}, sessionAt()
	if diff := cmp.Diff(want, signed); diff != "" {
		t.Errorf("SignedInWithPasskey (-want +got):\n%s", diff)
	}
	id[0] = 9 // the session keeps its own copy
	if !signed.SignedInBy([]byte{1, 2, 3}) || signed.SignedInBy([]byte{9, 2, 3}) || s.SignedInBy(nil) {
		t.Error("SignedInBy names the wrong passkey")
	}
	later, laterErr := signed.PasskeyChecked(sessionAt().Add(time.Hour))
	if laterErr != nil || !later.PasskeyCheckedAt.Equal(sessionAt().Add(time.Hour)) {
		t.Errorf("PasskeyChecked an hour later = %s, %v", later.PasskeyCheckedAt, laterErr)
	}
	back, backErr := later.PasskeyChecked(sessionAt())
	if backErr != nil || !back.PasskeyCheckedAt.Equal(sessionAt().Add(time.Hour)) {
		t.Errorf("PasskeyChecked earlier = %s, %v; want the later check kept", back.PasskeyCheckedAt, backErr)
	}
}

// TestDemoSessionHasNoPasskey checks that only a passkey session can record
// a passkey, and only a real one, or a check, so a demo session never passes
// D-183's check.
func TestDemoSessionHasNoPasskey(t *testing.T) {
	t.Parallel()
	demo, err := identity.NewSession("0123456789abcdef", identity.AuthenticatorDemo, identity.SessionScopeRead, "localhost", sessionAt(),
		identity.DefaultSessionPolicy())
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	owner := passkeySession(t)
	if _, nilErr := owner.SignedInWithPasskey(nil); !errors.Is(nilErr, identity.ErrNotPasskeySession) {
		t.Errorf("SignedInWithPasskey(nil) = %v, want %v", nilErr, identity.ErrNotPasskeySession)
	}
	if _, signErr := demo.SignedInWithPasskey([]byte{1}); !errors.Is(signErr, identity.ErrNotPasskeySession) {
		t.Errorf("a demo session signed in with a passkey = %v, want %v", signErr, identity.ErrNotPasskeySession)
	}
	if _, checkErr := demo.PasskeyChecked(sessionAt()); !errors.Is(checkErr, identity.ErrNotPasskeySession) {
		t.Errorf("a demo session checked = %v, want %v", checkErr, identity.ErrNotPasskeySession)
	}
	if recentErr := demo.CheckRecentPasskey(sessionAt()); !errors.Is(recentErr, identity.ErrPasskeyCheckRequired) {
		t.Errorf("a demo session's recent check = %v, want %v", recentErr, identity.ErrPasskeyCheckRequired)
	}
}
