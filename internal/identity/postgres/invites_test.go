package postgres_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/postgres"
)

// seededInvites are the codes seedInvites stores.
type seededInvites struct {
	// alice and bob are pending at now; the others are not.
	alice, bob identity.EnrollmentCode
	// expired is a viewer's invitation that ran out, owner an owner's own link,
	// used an invitation that was used.
	expired, owner, used identity.EnrollmentCode
	now                  time.Time
}

// seedInvites stores a viewer's invitations, an expired one, an owner's link
// and a used one in store, and returns them with a time at which two are pending.
func seedInvites(t *testing.T, store *postgres.Store) seededInvites {
	t.Helper()
	key := bytes.Repeat([]byte{4}, identity.KeyBytes)
	save := func(id string, created time.Time, ttl time.Duration, viewer *identity.ViewerInvite) identity.EnrollmentCode {
		c, err := identity.NewEnrollmentCode(identity.EnrollmentCodeID(id), localOrigin(t), created, ttl)
		if err != nil {
			t.Fatalf("NewEnrollmentCode: %v", err)
		}
		if viewer != nil {
			if err = c.WithViewer(viewer); err != nil {
				t.Fatalf("WithViewer: %v", err)
			}
		}
		if err = store.SaveEnrollmentCode(t.Context(), &c, identity.HashKey(key)); err != nil {
			t.Fatalf("SaveEnrollmentCode: %v", err)
		}
		return c
	}
	at := issuedAt()
	s := seededInvites{now: at.Add(30 * time.Minute)}
	s.alice = save("00000000000000a1", at, time.Hour, &identity.ViewerInvite{Name: "Alice", Targets: []identity.GrantTarget{{Kind: identity.GrantTargetSleeve, ID: "core"}}})
	s.bob = save("00000000000000b2", at.Add(time.Second), 24*time.Hour, &identity.ViewerInvite{Name: "Bob"})
	s.expired = save("00000000000000c3", at, 15*time.Minute, &identity.ViewerInvite{Name: "Carol"})
	s.owner = save("00000000000000d4", at, time.Hour, nil)
	s.used = save("00000000000000e5", at, time.Hour, &identity.ViewerInvite{Name: "Erin"})
	if err := store.UseEnrollmentCode(t.Context(), s.used.ID, at.Add(time.Minute)); err != nil {
		t.Fatalf("UseEnrollmentCode: %v", err)
	}
	return s
}

// TestPendingInvites checks the query behind the Access page's pending
// invitations: only a viewer's code not used, not cancelled and not expired is
// listed, oldest first and without a key hash.
func TestPendingInvites(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	s := seedInvites(t, store)
	got, err := store.PendingInvites(t.Context(), s.now)
	if err != nil {
		t.Fatalf("PendingInvites: %v", err)
	}
	if diff := cmp.Diff([]identity.EnrollmentCode{s.alice, s.bob}, got); diff != "" {
		t.Errorf("PendingInvites (-want +got):\n%s", diff)
	}
}

// wantCancelRefused checks that cancelling a code already cancelled, expired,
// an owner's link, used or unknown fails as an unknown enrollment code.
func (s *seededInvites) wantCancelRefused(t *testing.T, store *postgres.Store) {
	t.Helper()
	for name, id := range map[string]identity.EnrollmentCodeID{
		"cancelled already": s.alice.ID, "expired": s.expired.ID, "an owner link": s.owner.ID, "used": s.used.ID, "unknown": "fedcba9876543210",
	} {
		if err := store.CancelEnrollmentCode(t.Context(), id, s.now); !errors.Is(err, app.ErrUnknownEnrollmentCode) {
			t.Errorf("CancelEnrollmentCode of one %s = %v, want %v", name, err, app.ErrUnknownEnrollmentCode)
		}
	}
}

// TestCancelEnrollmentCode checks that cancelling cancels only a pending
// viewer's invitation, once, and that a cancelled code is no longer listed, can
// no longer be used and checks as cancelled.
func TestCancelEnrollmentCode(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	s := seedInvites(t, store)
	if err := store.CancelEnrollmentCode(ctx, s.alice.ID, s.now); err != nil {
		t.Fatalf("CancelEnrollmentCode: %v", err)
	}
	s.wantCancelRefused(t, store)
	if err := store.UseEnrollmentCode(ctx, s.alice.ID, s.now); !errors.Is(err, identity.ErrEnrollmentCodeUsed) {
		t.Errorf("UseEnrollmentCode of a cancelled code = %v, want %v", err, identity.ErrEnrollmentCodeUsed)
	}
	code, _, err := store.EnrollmentCode(ctx, s.alice.ID)
	if err != nil || !code.CancelledAt.Equal(s.now) || !errors.Is(code.Check(s.now, localOrigin(t)), identity.ErrEnrollmentCodeCancelled) {
		t.Errorf("EnrollmentCode after cancelling = %+v, %v, want CancelledAt set and ErrEnrollmentCodeCancelled", code, err)
	}
	left, err := store.PendingInvites(ctx, s.now)
	if err != nil || len(left) != 1 || left[0].ID != s.bob.ID {
		t.Errorf("PendingInvites after cancelling Alice = %v, %v, want Bob's alone", left, err)
	}
}
