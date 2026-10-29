package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
)

// errStoreDown is a store that can't be reached.
var errStoreDown = errors.New("store down")

// fakeOwnerTokens accepts one token as the owner's and fails every check
// when down.
type fakeOwnerTokens struct {
	owner string
	down  bool
}

func (f fakeOwnerTokens) Authenticate(_ context.Context, token string) (identityapp.Principal, error) {
	switch {
	case f.down:
		return identityapp.Principal{}, errStoreDown
	case token == f.owner:
		return identityapp.Principal{ScriptToken: "t1"}, nil
	}
	return identityapp.Principal{}, fmt.Errorf("%w: unknown", identityapp.ErrInvalidToken)
}

// fakeViewCallers accepts one token as a view plugin's, and fails every
// check when down.
type fakeViewCallers struct {
	view string
	down bool
}

func (f fakeViewCallers) VerifyCaller(_ context.Context, token string) (string, error) {
	switch {
	case f.down:
		return "", errStoreDown
	case token == f.view:
		return "ghostfolio", nil
	}
	return "", identityconnect.ErrCallerRefused
}

// TestReadTokens checks who may read the read API (D-27): the owner's
// sessions and script tokens as before, and now a view plugin proven by
// its projected token, the callers the AuthService accepts; anyone else is
// refused as an invalid token, and a check that can't be made fails as
// such, never as a refusal or an acceptance.
func TestReadTokens(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		tokens    readTokens
		token     string
		wantErr   error
		wantOther bool
	}{
		{name: "the owner's token", tokens: readTokens{auth: fakeOwnerTokens{owner: "own"}, plugins: fakeViewCallers{view: "view"}}, token: "own"},
		{name: "a view plugin's token", tokens: readTokens{auth: fakeOwnerTokens{owner: "own"}, plugins: fakeViewCallers{view: "view"}}, token: "view"},
		{name: "neither", tokens: readTokens{auth: fakeOwnerTokens{owner: "own"}, plugins: fakeViewCallers{view: "view"}}, token: "other", wantErr: viewsapp.ErrInvalidToken},
		{name: "no plugin callers", tokens: readTokens{auth: fakeOwnerTokens{owner: "own"}}, token: "view", wantErr: viewsapp.ErrInvalidToken},
		{name: "the owner's store is down", tokens: readTokens{auth: fakeOwnerTokens{down: true}, plugins: fakeViewCallers{view: "view"}}, token: "view", wantOther: true},
		{name: "the plugin check fails", tokens: readTokens{auth: fakeOwnerTokens{owner: "own"}, plugins: fakeViewCallers{down: true}}, token: "view", wantOther: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.tokens.VerifyToken(t.Context(), tt.token)
			switch {
			case tt.wantOther:
				if err == nil || errors.Is(err, viewsapp.ErrInvalidToken) {
					t.Errorf("VerifyToken = %v, want a failure that is not an invalid token", err)
				}
			case !errors.Is(err, tt.wantErr) || (tt.wantErr == nil) != (err == nil):
				t.Errorf("VerifyToken = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
