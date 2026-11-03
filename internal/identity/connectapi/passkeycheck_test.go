package connectapi_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/identity"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

func (*sessionStore) SaveCeremony(context.Context, *identity.Ceremony) error { return nil }

func (s *sessionStore) RevokeSessions(_ context.Context, at time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for id := range s.sessions {
		ss := s.sessions[id]
		if ss.RevokedAt.IsZero() {
			ss.RevokedAt = at
			s.sessions[id] = ss
			n++
		}
	}
	return n, nil
}

// staleOwnerSession plants in store an owner session that a passkey signed
// in an hour before the service's clock, so its passkey check is stale, and
// returns its token.
func staleOwnerSession(t *testing.T, store *sessionStore) string {
	t.Helper()
	at := time.Date(2026, 10, 28, 11, 0, 0, 0, time.UTC)
	s, err := identity.NewSession("0123456789abcdef", identity.AuthenticatorPasskey, identity.SessionScopeOwner, "localhost", at,
		identity.DefaultSessionPolicy())
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if s, err = s.SignedInWithPasskey([]byte{1, 2, 3}); err != nil {
		t.Fatalf("SignedInWithPasskey: %v", err)
	}
	key := bytes.Repeat([]byte{0x42}, identity.KeyBytes)
	if err := store.SaveSession(t.Context(), &s, identity.HashKey(key)); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
	return identity.FormatSessionToken(s.ID, key)
}

// wantPasskeyCheck fails unless err is FAILED_PRECONDITION with the
// PasskeyCheckRequired detail naming the five-minute window (D-183).
func wantPasskeyCheck(t *testing.T, what string, err error) {
	t.Helper()
	ce, ok := errors.AsType[*connect.Error](err)
	if !ok || ce.Code() != connect.CodeFailedPrecondition {
		t.Fatalf("%s = %v, want %s", what, err, connect.CodeFailedPrecondition)
	}
	for _, d := range ce.Details() {
		msg, valueErr := d.Value()
		if valueErr != nil {
			t.Fatalf("%s: detail: %v", what, valueErr)
		}
		if req, isCheck := msg.(*identityv1.PasskeyCheckRequired); isCheck {
			if req.GetWindow().AsDuration() != identity.PasskeyCheckWindow {
				t.Errorf("%s: window %s, want %s", what, req.GetWindow().AsDuration(), identity.PasskeyCheckWindow)
			}
			return
		}
	}
	t.Errorf("%s: no PasskeyCheckRequired detail in %v", what, ce.Details())
}

// TestPasskeyCheckOverConnect checks that the actions D-183 protects fail
// with FAILED_PRECONDITION and the PasskeyCheckRequired detail in a session
// without a recent passkey check, that a passkey check begins in it, and
// that signing out everywhere revokes every session.
func TestPasskeyCheckOverConnect(t *testing.T) {
	t.Parallel()
	c, store := newServed(t, viewCaller)
	ctx := t.Context()
	token := staleOwnerSession(t, store)
	_, err := c.CreateScriptToken(ctx, connect.NewRequest(&identityv1.CreateScriptTokenRequest{Origin: local, SessionToken: token, Name: "x"}))
	wantPasskeyCheck(t, "CreateScriptToken", err)
	_, err = c.CreateEnrollmentLink(ctx, connect.NewRequest(&identityv1.CreateEnrollmentLinkRequest{Origin: local, SessionToken: token}))
	wantPasskeyCheck(t, "CreateEnrollmentLink", err)
	_, err = c.RemovePasskey(ctx, connect.NewRequest(&identityv1.RemovePasskeyRequest{Origin: local, SessionToken: token, PasskeyId: "AQID"}))
	wantPasskeyCheck(t, "RemovePasskey", err)
	_, err = c.BeginRegistration(ctx, connect.NewRequest(&identityv1.BeginRegistrationRequest{
		Origin: local, Authorization: &identityv1.BeginRegistrationRequest_SessionToken{SessionToken: token},
	}))
	wantPasskeyCheck(t, "BeginRegistration", err)

	begun, err := c.BeginPasskeyCheck(ctx, connect.NewRequest(&identityv1.BeginPasskeyCheckRequest{Origin: local, SessionToken: token}))
	if err != nil || begun.Msg.GetCeremonyId() == "" || begun.Msg.GetOptionsJson() == "" {
		t.Errorf("BeginPasskeyCheck = %v, %v", begun, err)
	}
	_, err = c.FinishPasskeyCheck(ctx, connect.NewRequest(&identityv1.FinishPasskeyCheckRequest{
		Origin: local, SessionToken: token, CeremonyId: "nope", CredentialJson: "{}",
	}))
	wantCode(t, "FinishPasskeyCheck of an unknown ceremony", err, connect.CodeFailedPrecondition)

	demo := demoSession(t, c)
	_, err = c.BeginPasskeyCheck(ctx, connect.NewRequest(&identityv1.BeginPasskeyCheckRequest{Origin: local, SessionToken: demo}))
	wantCode(t, "BeginPasskeyCheck with a read session", err, connect.CodePermissionDenied)
	_, err = c.SignOutEverywhere(ctx, connect.NewRequest(&identityv1.SignOutEverywhereRequest{Origin: local, SessionToken: demo}))
	wantCode(t, "SignOutEverywhere with a read session", err, connect.CodePermissionDenied)
	out, err := c.SignOutEverywhere(ctx, connect.NewRequest(&identityv1.SignOutEverywhereRequest{Origin: local, SessionToken: token}))
	if err != nil || out.Msg.GetRevokedSessions() != 2 {
		t.Errorf("SignOutEverywhere = %v, %v; want 2 sessions revoked", out, err)
	}
	for _, tok := range []string{token, demo} {
		_, err = c.WhoAmI(ctx, connect.NewRequest(&identityv1.WhoAmIRequest{Origin: local, SessionToken: tok}))
		wantCode(t, "WhoAmI after signing out everywhere", err, connect.CodeUnauthenticated)
	}
}
