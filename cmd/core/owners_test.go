package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// fakeSessions knows the owner-scoped session "owner" (ID s1) and the
// read-scoped "read".
type fakeSessions struct{}

func (fakeSessions) Session(_ context.Context, _ identity.Origin, token string, _ identity.SessionScope) (identity.Session, error) {
	switch token {
	case "owner":
		return identity.Session{ID: "s1", Scope: identity.SessionScopeOwner}, nil
	case "read":
		return identity.Session{}, fmt.Errorf("%w: read scope", identityapp.ErrForbidden)
	}
	return identity.Session{}, fmt.Errorf("%w: unknown", identityapp.ErrInvalidSession)
}

// recordingAuthorizer answers as identityapp.Policy except for the account
// denied, or fails with fail, and records whom it was asked about.
type recordingAuthorizer struct {
	denied string
	fail   error
	asked  []identity.Caller
}

func (r *recordingAuthorizer) MayAct(_ context.Context, c identity.Caller, a identity.Action, account string) (bool, error) {
	r.asked = append(r.asked, c)
	if r.fail != nil {
		return false, r.fail
	}
	return identity.MayAct(c, a) && account != r.denied, nil
}

// errAuthzDown is an authorizer that can't tell.
var errAuthzDown = errors.New("authorizer down")

// TestOwnerSessionsChangeable checks that the owner services' port asks
// the authorizer about the owner of the session, for a change of each
// account (D-163), and refuses without a valid owner session.
func TestOwnerSessionsChangeable(t *testing.T) {
	t.Parallel()
	origins, err := identity.NewOrigins([]string{"http://localhost:8080"})
	if err != nil {
		t.Fatalf("NewOrigins: %v", err)
	}
	accounts := []ledger.AccountID{"acc-1", "acc-2"}
	const local = "http://localhost:8080"
	tests := []struct {
		name    string
		token   string
		origin  string
		authz   identityapp.Authorizer
		want    []ledger.AccountID
		wantErr error
	}{
		{name: "every account", token: "owner", origin: local, authz: &recordingAuthorizer{}, want: accounts},
		{name: "one denied", token: "owner", origin: local, authz: &recordingAuthorizer{denied: "acc-2"}, want: []ledger.AccountID{"acc-1"}},
		{name: "no authorizer changes none", token: "owner", origin: local, want: []ledger.AccountID{}},
		{name: "the authorizer can't tell", token: "owner", origin: local, authz: &recordingAuthorizer{fail: errAuthzDown}, wantErr: errAuthzDown},
		{name: "a read session", token: "read", origin: local, authz: &recordingAuthorizer{}, wantErr: ingestapp.ErrNotOwner},
		{name: "no session", token: "nope", origin: local, authz: &recordingAuthorizer{}, wantErr: ingestapp.ErrNoSession},
		{name: "an origin off the list", token: "owner", origin: "https://other.example.com", authz: &recordingAuthorizer{}, wantErr: ingestapp.ErrNotOwner},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o := &ownerSessions{auth: fakeSessions{}, origins: origins, noSession: ingestapp.ErrNoSession, notOwner: ingestapp.ErrNotOwner, authz: tt.authz}
			got, err := o.Changeable(t.Context(), tt.origin, tt.token, accounts)
			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				t.Fatalf("Changeable error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Changeable mismatch (-want +got):\n%s", diff)
			}
			checkAskedOwner(t, tt.authz)
		})
	}
}

// checkAskedOwner checks that a, when it is a recordingAuthorizer, was
// asked only about the owner of session s1.
func checkAskedOwner(t *testing.T, a identityapp.Authorizer) {
	t.Helper()
	r, ok := a.(*recordingAuthorizer)
	if !ok {
		return
	}
	for _, c := range r.asked {
		if c != (identity.Caller{Holder: identity.HolderOwner, ID: "s1"}) {
			t.Errorf("the authorizer was asked about %+v, want the owner of session s1", c)
		}
	}
}
