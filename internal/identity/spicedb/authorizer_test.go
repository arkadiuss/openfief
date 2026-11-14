package spicedb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/spicedb"
	"github.com/arkadiuss/openfief/internal/kernel"
)

var errSpiceDB = errors.New("spicedb unavailable")

func ref(typ, id string) *v1.SubjectReference {
	return &v1.SubjectReference{Object: &v1.ObjectReference{ObjectType: typ, ObjectId: id}}
}

// TestAuthorizerAsksTheAccountPermission checks the question each holder's
// check puts to SpiceDB (D-164): the account's view or change permission
// for the holder's subject, fully consistent, with IDs SpiceDB takes.
func TestAuthorizerAsksTheAccountPermission(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		caller  identity.Caller
		action  identity.Action
		account string
		want    *v1.CheckPermissionRequest
	}{
		{
			name: "the owner is the one person", caller: identity.Caller{Holder: identity.HolderOwner, ID: "s-1"},
			action: identity.ActionChange, account: "brokerage",
			want: &v1.CheckPermissionRequest{
				Resource:   &v1.ObjectReference{ObjectType: "account", ObjectId: "brokerage"},
				Permission: "change", Subject: ref("person", "owner"),
			},
		},
		{
			name: "a viewer is a person of their own, by their ID", caller: identity.Caller{Holder: identity.HolderViewer, ID: "0123456789abcdef"},
			action: identity.ActionView, account: "bonds",
			want: &v1.CheckPermissionRequest{
				Resource:   &v1.ObjectReference{ObjectType: "account", ObjectId: "bonds"},
				Permission: "view", Subject: ref("person", "0123456789abcdef"),
			},
		},
		{
			name: "a script token by its ID", caller: identity.Caller{Holder: identity.HolderScriptToken, ID: "0123456789abcdef"},
			action: identity.ActionView, account: "bonds",
			want: &v1.CheckPermissionRequest{
				Resource:   &v1.ObjectReference{ObjectType: "account", ObjectId: "bonds"},
				Permission: "view", Subject: ref("script_token", "0123456789abcdef"),
			},
		},
		{
			name: "a view plugin by its name, escaped", caller: identity.Caller{Holder: identity.HolderViewPlugin, ID: "ui.v2"},
			action: identity.ActionView, account: "IKE (PL)",
			want: &v1.CheckPermissionRequest{
				Resource:   &v1.ObjectReference{ObjectType: "account", ObjectId: "IKE=20=28PL=29"},
				Permission: "view", Subject: ref("view_plugin", "ui=2ev2"),
			},
		},
		{
			name: "a demo session by its ID", caller: identity.Caller{Holder: identity.HolderDemoSession, ID: "d1"},
			action: identity.ActionView, account: "bonds",
			want: &v1.CheckPermissionRequest{
				Resource:   &v1.ObjectReference{ObjectType: "account", ObjectId: "bonds"},
				Permission: "view", Subject: ref("demo_session", "d1"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := &fake{check: v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION}
			ok, err := spicedb.NewAuthorizer(f).MayAct(t.Context(), tt.caller, tt.action, tt.account)
			if err != nil || !ok {
				t.Fatalf("MayAct = %v, %v, want true", ok, err)
			}
			if len(f.checks) != 1 {
				t.Fatalf("%d checks, want 1", len(f.checks))
			}
			got := f.checks[0]
			if !got.GetConsistency().GetFullyConsistent() {
				t.Errorf("check is not fully consistent: %v", got.GetConsistency())
			}
			got.Consistency = nil
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("check request (-want +got):\n%s", diff)
			}
		})
	}
}

// TestAuthorizerAnswers checks that only SpiceDB's HAS_PERMISSION allows,
// and that a caller or action that is none never reaches SpiceDB.
func TestAuthorizerAnswers(t *testing.T) {
	t.Parallel()
	owner := identity.Caller{Holder: identity.HolderOwner, ID: "s"}
	tests := []struct {
		name      string
		check     v1.CheckPermissionResponse_Permissionship
		caller    identity.Caller
		action    identity.Action
		want      bool
		wantAsked int
	}{
		{name: "has permission", check: v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, caller: owner, action: identity.ActionView, want: true, wantAsked: 1},
		{name: "no permission", check: v1.CheckPermissionResponse_PERMISSIONSHIP_NO_PERMISSION, caller: owner, action: identity.ActionView, wantAsked: 1},
		{name: "conditional is not enough", check: v1.CheckPermissionResponse_PERMISSIONSHIP_CONDITIONAL_PERMISSION, caller: owner, action: identity.ActionView, wantAsked: 1},
		{name: "unspecified is not enough", check: v1.CheckPermissionResponse_PERMISSIONSHIP_UNSPECIFIED, caller: owner, action: identity.ActionChange, wantAsked: 1},
		{name: "nobody is refused without a call", check: v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, caller: identity.Caller{}, action: identity.ActionView},
		{name: "no action is refused without a call", check: v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, caller: owner},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := &fake{check: tt.check}
			got, err := spicedb.NewAuthorizer(f).MayAct(t.Context(), tt.caller, tt.action, "bonds")
			if err != nil {
				t.Fatalf("MayAct: %v", err)
			}
			if got != tt.want {
				t.Errorf("MayAct = %v, want %v", got, tt.want)
			}
			if len(f.checks) != tt.wantAsked {
				t.Errorf("%d checks, want %d", len(f.checks), tt.wantAsked)
			}
		})
	}
}

// holders is one caller of each kind.
func holders() []identity.Caller {
	return []identity.Caller{
		{Holder: identity.HolderOwner, ID: "s"},
		{Holder: identity.HolderScriptToken, ID: "t"},
		{Holder: identity.HolderViewPlugin, ID: "ui"},
		{Holder: identity.HolderDemoSession, ID: "d"},
	}
}

// TestAuthorizerFailsClosed checks that no failure of SpiceDB, a timeout
// included, ever allows: not for an owner, who holds every account when it
// answers, either.
func TestAuthorizerFailsClosed(t *testing.T) {
	t.Parallel()
	failures := []struct {
		name string
		f    func() *fake
		want error
	}{
		{"error", func() *fake {
			return &fake{err: errSpiceDB, check: v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION}
		}, errSpiceDB},
		{"timeout", func() *fake { return &fake{block: true} }, context.DeadlineExceeded},
	}
	for _, c := range holders() {
		for _, a := range []identity.Action{identity.ActionView, identity.ActionChange} {
			for _, fl := range failures {
				t.Run(c.Holder.String()+" "+a.String()+" "+fl.name, func(t *testing.T) {
					t.Parallel()
					ok, err := spicedb.NewAuthorizer(fl.f(), spicedb.WithTimeout(10*time.Millisecond)).MayAct(t.Context(), c, a, "bonds")
					if ok || !errors.Is(err, fl.want) {
						t.Errorf("MayAct = %v, %v, want false and %v", ok, err, fl.want)
					}
				})
			}
		}
	}
}

// TestAuthorizerRefusesAnIDItCantTake checks that an empty account or an
// empty holder ID is a refusal with an error, and no call.
func TestAuthorizerRefusesAnIDItCantTake(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		caller  identity.Caller
		account string
	}{
		{"an empty account", identity.Caller{Holder: identity.HolderOwner, ID: "s"}, ""},
		{"a holder with no ID", identity.Caller{Holder: identity.HolderScriptToken}, "bonds"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := &fake{check: v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION}
			ok, err := spicedb.NewAuthorizer(f).MayAct(t.Context(), tt.caller, identity.ActionView, tt.account)
			if ok || !errors.Is(err, spicedb.ErrInvalidID) || len(f.checks) != 0 {
				t.Errorf("MayAct = %v, %v after %d checks, want false, ErrInvalidID and no call", ok, err, len(f.checks))
			}
		})
	}
}

func lookedUp(ids ...string) []*v1.LookupResourcesResponse {
	out := make([]*v1.LookupResourcesResponse, 0, len(ids))
	for _, id := range ids {
		out = append(out, &v1.LookupResourcesResponse{
			ResourceObjectId: id,
			Permissionship:   v1.LookupPermissionship_LOOKUP_PERMISSIONSHIP_HAS_PERMISSION,
		})
	}
	return out
}

// scriptToken is the caller the list tests ask for.
func scriptToken() identity.Caller {
	return identity.Caller{Holder: identity.HolderScriptToken, ID: "t"}
}

// TestAuthorizerActableAccounts checks the list answer: one
// LookupResources call, IDs decoded, and a conditional result left out.
func TestAuthorizerActableAccounts(t *testing.T) {
	t.Parallel()
	f := &fake{lookup: append(lookedUp("bonds", "IKE=20=28PL=29"), &v1.LookupResourcesResponse{
		ResourceObjectId: "maybe", Permissionship: v1.LookupPermissionship_LOOKUP_PERMISSIONSHIP_CONDITIONAL_PERMISSION,
	})}
	got, err := spicedb.NewAuthorizer(f).ActableAccounts(t.Context(), scriptToken(), identity.ActionView)
	if err != nil {
		t.Fatalf("ActableAccounts: %v", err)
	}
	if diff := cmp.Diff(map[string]struct{}{"bonds": {}, "IKE (PL)": {}}, got); diff != "" {
		t.Errorf("accounts (-want +got):\n%s", diff)
	}
	if len(f.lookups) != 1 {
		t.Fatalf("%d lookups, want 1", len(f.lookups))
	}
	l := f.lookups[0]
	if l.GetResourceObjectType() != "account" || l.GetPermission() != "view" || !l.GetConsistency().GetFullyConsistent() ||
		l.GetSubject().GetObject().GetObjectType() != "script_token" || l.GetSubject().GetObject().GetObjectId() != "t" {
		t.Errorf("lookup request = %v", l)
	}
}

// TestAuthorizerActableAccountsFailsClosed checks that any failure of the
// lookup returns no accounts at all, with the error.
func TestAuthorizerActableAccountsFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		f    *fake
		want error
	}{
		{"a stream that fails", &fake{lookup: lookedUp("bonds"), lookupErr: errSpiceDB}, errSpiceDB},
		{"a call that fails", &fake{err: errSpiceDB}, errSpiceDB},
		{"a timeout", &fake{block: true}, context.DeadlineExceeded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := spicedb.NewAuthorizer(tt.f, spicedb.WithTimeout(10*time.Millisecond)).ActableAccounts(t.Context(), scriptToken(), identity.ActionView)
			if got != nil || !errors.Is(err, tt.want) {
				t.Errorf("ActableAccounts = %v, %v, want none and %v", got, err, tt.want)
			}
		})
	}
}

// TestAuthorizerListsNothingForNobody checks that a caller that is no
// holder lists no account, without a call.
func TestAuthorizerListsNothingForNobody(t *testing.T) {
	t.Parallel()
	f := &fake{lookup: lookedUp("bonds")}
	got, err := spicedb.NewAuthorizer(f).ActableAccounts(t.Context(), identity.Caller{}, identity.ActionView)
	if err != nil || len(got) != 0 || len(f.lookups) != 0 {
		t.Errorf("ActableAccounts = %v, %v after %d lookups, want none", got, err, len(f.lookups))
	}
}

// TestObjectID checks that IDs SpiceDB takes stay as they are and others
// are escaped so that no two IDs share an object.
func TestObjectID(t *testing.T) {
	t.Parallel()
	tests := []struct{ id, want string }{
		{"brokerage", "brokerage"},
		{"acc-1_x/y|z+w", "acc-1_x/y|z+w"},
		{"a b", "a=20b"},
		{"a=b", "a=3db"},
		{"a=3db", "a=3d3db"},
		{"ikę", "ik=c4=99"},
		{"ui.v2", "ui=2ev2"},
	}
	seen := map[string]string{}
	for _, tt := range tests {
		got, err := spicedb.ObjectID(tt.id)
		if err != nil {
			t.Fatalf("ObjectID(%q): %v", tt.id, err)
		}
		if got != tt.want {
			t.Errorf("ObjectID(%q) = %q, want %q", tt.id, got, tt.want)
		}
		if other, dup := seen[got]; dup {
			t.Errorf("ObjectID(%q) and ObjectID(%q) are both %q", tt.id, other, got)
		}
		seen[got] = tt.id
	}
	if _, err := spicedb.ObjectID(""); !errors.Is(err, spicedb.ErrInvalidID) {
		t.Errorf("ObjectID(\"\") = %v, want ErrInvalidID", err)
	}
}

// TestAuthorizerUnavailableIsMarked checks that a SpiceDB that fails or is too
// slow, in a check, a lookup or a ping, is reported as
// kernel.ErrAuthzUnavailable, which the APIs answer as 503; that a refused ID
// is not, since SpiceDB was never asked; and that a client that went away is
// not blamed on SpiceDB.
func TestAuthorizerUnavailableIsMarked(t *testing.T) {
	t.Parallel()
	owner := identity.Caller{Holder: identity.HolderOwner, ID: "s"}
	check := func(ctx context.Context, z *spicedb.Authorizer) error {
		_, err := z.MayAct(ctx, owner, identity.ActionView, "bonds")
		return fmt.Errorf("check: %w", err)
	}
	lookup := func(ctx context.Context, z *spicedb.Authorizer) error {
		_, err := z.ActableAccounts(ctx, owner, identity.ActionView)
		return fmt.Errorf("lookup: %w", err)
	}
	ping := func(ctx context.Context, z *spicedb.Authorizer) error { return fmt.Errorf("ping: %w", z.Ping(ctx)) }
	tests := []struct {
		name string
		call func(ctx context.Context, z *spicedb.Authorizer) error
		f    *fake
	}{
		{"a check that fails", check, &fake{err: errSpiceDB}},
		{"a check that times out", check, &fake{block: true}},
		{"a lookup that fails", lookup, &fake{err: errSpiceDB}},
		{"a lookup that times out", lookup, &fake{block: true}},
		{"a lookup stream that fails", lookup, &fake{lookup: lookedUp("bonds"), lookupErr: errSpiceDB}},
		{"a ping that fails", ping, &fake{err: errSpiceDB}},
		{"a ping that times out", ping, &fake{block: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.call(t.Context(), spicedb.NewAuthorizer(tt.f, spicedb.WithTimeout(10*time.Millisecond)))
			if !errors.Is(err, kernel.ErrAuthzUnavailable) {
				t.Errorf("error = %v, want kernel.ErrAuthzUnavailable", err)
			}
		})
	}
	t.Run("a refused ID", func(t *testing.T) {
		t.Parallel()
		_, err := spicedb.NewAuthorizer(&fake{}).MayAct(t.Context(), owner, identity.ActionView, "")
		if err == nil || errors.Is(err, kernel.ErrAuthzUnavailable) {
			t.Errorf("error = %v, want a refusal that is not unavailable: SpiceDB was never asked", err)
		}
	})
	t.Run("a client that went away", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		_, err := spicedb.NewAuthorizer(&fake{block: true}).MayAct(ctx, owner, identity.ActionView, "bonds")
		if err == nil || errors.Is(err, kernel.ErrAuthzUnavailable) {
			t.Errorf("error = %v, want one that is not unavailable", err)
		}
	})
	t.Run("a ping that is answered", func(t *testing.T) {
		t.Parallel()
		f := &fake{}
		if err := spicedb.NewAuthorizer(f).Ping(t.Context()); err != nil || len(f.checks) != 1 {
			t.Errorf("Ping = %v after %d checks, want nil after one", err, len(f.checks))
		}
	})
}
