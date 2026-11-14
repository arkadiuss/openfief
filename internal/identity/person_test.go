package identity_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
)

func TestNewViewer(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
	handle := bytes.Repeat([]byte{3}, identity.OwnerHandleBytes)
	tests := []struct {
		name    string
		id      string
		person  string
		handle  []byte
		wantErr bool
	}{
		{"a viewer", "0123456789abcdef", "Alice", handle, false},
		{"a name of 64 characters", "0123456789abcdef", strings.Repeat("a", 64), handle, false},
		{"the owner's ID", identity.OwnerPersonID, "Alice", handle, true},
		{"an ID that is not hex", "not-a-hex-id-xyz", "Alice", handle, true},
		{"no name", "0123456789abcdef", "", handle, true},
		{"a name of 65 characters", "0123456789abcdef", strings.Repeat("a", 65), handle, true},
		{"an unprintable name", "0123456789abcdef", "Al\x00ice", handle, true},
		{"a short handle", "0123456789abcdef", "Alice", handle[:5], true},
		{"no handle", "0123456789abcdef", "Alice", nil, true},
	}
	for _, tt := range tests {
		p, err := identity.NewViewer(tt.id, tt.person, tt.handle, now)
		if (err != nil) != tt.wantErr || (err != nil && !errors.Is(err, identity.ErrInvalidPerson)) {
			t.Errorf("NewViewer(%s) = %v, want error %v", tt.name, err, tt.wantErr)
		}
		if err == nil && (!p.IsViewer() || !p.Active()) {
			t.Errorf("NewViewer(%s) = %+v, want an active viewer", tt.name, p)
		}
	}
}

func TestPersonRevoked(t *testing.T) {
	t.Parallel()
	created := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
	v, err := identity.NewViewer("0123456789abcdef", "Alice", bytes.Repeat([]byte{3}, identity.OwnerHandleBytes), created)
	if err != nil {
		t.Fatal(err)
	}
	if got := v.Revoked(created.Add(time.Hour)); got.Active() || !got.RevokedAt.Equal(created.Add(time.Hour)) {
		t.Errorf("Revoked = %+v, want revoked an hour after", got)
	}
	if got := v.Revoked(created.Add(-time.Hour)); !got.RevokedAt.Equal(created) {
		t.Errorf("a revocation before creation = %v, want clamped to creation", got.RevokedAt)
	}
	owner := identity.OwnerPerson(&identity.Owner{Handle: bytes.Repeat([]byte{1}, identity.OwnerHandleBytes), CreatedAt: created})
	if got := owner.Revoked(created.Add(time.Hour)); !got.Active() {
		t.Error("the owner was revoked")
	}
}

func TestViewerSessionReadsOnly(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
	mk := func(scope identity.SessionScope) identity.Session {
		s, err := identity.NewSession("0123456789abcdef", identity.AuthenticatorPasskey, scope, "localhost", now, identity.DefaultSessionPolicy())
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	read, owner := mk(identity.SessionScopeRead), mk(identity.SessionScopeOwner)
	if v, err := read.ForViewer("fedcba9876543210"); err != nil || !v.OfViewer() || v.Check(now, "localhost", identity.SessionScopeOwner) == nil {
		t.Errorf("ForViewer(read) = %+v, %v, want a viewer's session the owner scope refuses", v, err)
	}
	if _, err := owner.ForViewer("fedcba9876543210"); !errors.Is(err, identity.ErrSessionScope) {
		t.Errorf("ForViewer(owner scope) = %v, want %v", err, identity.ErrSessionScope)
	}
	if _, err := read.ForViewer("owner"); !errors.Is(err, identity.ErrSessionScope) {
		t.Errorf("ForViewer(the owner's ID) = %v, want %v", err, identity.ErrSessionScope)
	}
	if owner.OfViewer() || read.Person() != identity.OwnerPersonID {
		t.Error("a session with no person is not the owner's")
	}
}

func TestMayActViewerHoldsNothingByItself(t *testing.T) {
	t.Parallel()
	c := identity.Caller{Holder: identity.HolderViewer, ID: "0123456789abcdef"}
	for _, a := range []identity.Action{identity.ActionView, identity.ActionChange} {
		if identity.MayAct(c, a) {
			t.Errorf("MayAct(viewer, %s) = true: a viewer holds only their grants", a)
		}
	}
}
