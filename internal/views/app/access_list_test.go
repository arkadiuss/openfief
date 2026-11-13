package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/identity"
)

// listingAuthorizer answers as denyingAuthorizer and also lists the
// accounts a caller may view in one call, as SpiceDB's LookupResources does
// (D-164); it counts the calls of each kind.
type listingAuthorizer struct {
	denyingAuthorizer

	accounts []string
	lists    int
	checks   int
}

func (l *listingAuthorizer) MayAct(ctx context.Context, c identity.Caller, a identity.Action, account string) (bool, error) {
	l.checks++
	return l.denyingAuthorizer.MayAct(ctx, c, a, account)
}

func (l *listingAuthorizer) ActableAccounts(_ context.Context, c identity.Caller, a identity.Action) (map[string]struct{}, error) {
	l.lists++
	if l.fail != nil {
		return nil, l.fail
	}
	out := map[string]struct{}{}
	if identity.MayAct(c, a) {
		for _, id := range l.accounts {
			out[id] = struct{}{}
		}
	}
	return out, nil
}

// TestViewsListAccountsInOneCall checks that an authorizer that can list
// the accounts a caller may view is asked once for every list, total and
// allocation, with the same results as asking about each account (D-163).
func TestViewsListAccountsInOneCall(t *testing.T) {
	t.Parallel()
	ctx := reader(t)
	authz := &listingAuthorizer{accounts: []string{"acc-1"}}
	got, err := withAuthorizer(t, newFixture(t), authz).Accounts(ctx)
	if err != nil {
		t.Fatalf("Accounts: %v", err)
	}
	want, err := withAuthorizer(t, newFixture(t), &denyingAuthorizer{denied: "acc-2"}).Accounts(ctx) // asked per account
	if err != nil {
		t.Fatalf("Accounts per account: %v", err)
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("Accounts by list differs from by check (-want +got):\n%s", diff)
	}
	if authz.lists != 1 || authz.checks != 0 {
		t.Errorf("%d lists and %d checks for Accounts, want 1 and 0", authz.lists, authz.checks)
	}
}

// TestViewsRefuseWhenTheListFails checks that an authorizer that can't list
// leaves a caller with an error, never every account.
func TestViewsRefuseWhenTheListFails(t *testing.T) {
	t.Parallel()
	authz := &listingAuthorizer{accounts: []string{"acc-1", "acc-2"}}
	authz.fail = errAuthz
	accounts, err := withAuthorizer(t, newFixture(t), authz).Accounts(reader(t))
	if err == nil || len(accounts) != 0 {
		t.Errorf("Accounts = %v, %v, want an error and none", accounts, err)
	}
}
