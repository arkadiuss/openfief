//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/plugins"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// TestGrantsLimitWhatAScriptTokenViews checks D-209 on the core that answers from SpiceDB: a script token with no
// grant views nothing; granted one sleeve it views that sleeve's accounts alone, in the totals and the allocation
// too, and an account created in the sleeve afterwards is covered by the same grant, while every other account
// answers 404; revoked, the next call views nothing; granted one account it views that account alone; and the audit
// log shows each grant and revocation, naming the token and its target by ID and nothing else (D-211)
// (SCEN-e2e-grants-limit-what-a-script-token-views).
//
// It is not parallel, on purpose: it creates an account in a sleeve and deletes it again, which the tests that
// expect the demo's three accounts must not see, so Go runs it before any parallel test starts. The owner's session
// is a virtual authenticator's passkey made through an enrollment link, as a browser's would be.
//
//nolint:paralleltest // see above: the test creates an account the other tests must not see
func TestGrantsLimitWhatAScriptTokenViews(t *testing.T) {
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitCoreAvailable(t, k)
	waitSeedComplete(t, k)
	waitSpiceDBAvailable(t, k)
	waitPluginReady(t, k, uiPlugin)
	waitLogLine(t, k, coreDeployment, reconciledLine)
	owner := newOwnerAPI(t, k)
	token := issueTokenWith(t, k)
	secret, tokenID := token.secret, token.id
	household := issueToken(t, k)
	viewed := func(ids []string) bool { return slices.Equal(ids, demoAccountIDs()) }
	waitAccounts(t, k, household, viewed)
	checkCutoverKeptTheUI(t, k, owner)

	// With no grant the token views nothing: no account lists, and each answers as one that doesn't exist.
	if got := waitAccounts(t, k, secret, func([]string) bool { return true }); len(got) != 0 {
		t.Fatalf("a token with no grant lists %v, want nothing", got)
	}
	expectStatus(t, k, secret, accountPath(demo.XTBAccount), http.StatusNotFound, "an account of a token with no grant")

	// One sleeve: the defensive sleeve's account alone.
	sleeveKind, tokenKind := identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE, identityv1.GrantHolderKind_GRANT_HOLDER_KIND_SCRIPT_TOKEN
	owner.grant(t, tokenKind, tokenID, sleeveKind, string(demo.DefensiveSleeve))
	bonds := []string{string(demo.BondsAccount)}
	if got := waitAccounts(t, k, secret, func(ids []string) bool { return len(ids) > 0 }); !slices.Equal(got, bonds) {
		t.Fatalf("a token granted the defensive sleeve lists %v, want %v", got, bonds)
	}
	expectStatus(t, k, secret, accountPath(demo.BondsAccount), http.StatusOK, "the granted sleeve's account")
	expectStatus(t, k, secret, accountPath(demo.XTBAccount), http.StatusNotFound, "an account outside the granted sleeve")
	expectStatus(t, k, secret, accountPath(demo.IBKRAccount), http.StatusNotFound, "an account outside the granted sleeve")
	checkSleeveTotals(t, k, secret, household)

	// An account created in the sleeve later is covered by the same grant.
	later := owner.createAccount(t, "E2E later account", string(demo.DefensiveSleeve))
	t.Cleanup(func() { owner.deleteAccount(t, later) })
	if got := waitAccounts(t, k, secret, func(ids []string) bool { return len(ids) == len(bonds)+1 }); !slices.Equal(got, slices.Sorted(slices.Values(append(slices.Clone(bonds), later)))) {
		t.Fatalf("after an account was created in the sleeve the token lists %v, want %v and %s", got, bonds, later)
	}
	expectStatus(t, k, secret, accountPath(ledger.AccountID(later)), http.StatusOK, "an account created in the granted sleeve")

	// Revoked, the next call views nothing.
	owner.revoke(t, tokenKind, tokenID, sleeveKind, string(demo.DefensiveSleeve))
	if got := waitAccounts(t, k, secret, func(ids []string) bool { return len(ids) == 0 }); len(got) != 0 {
		t.Fatalf("a token whose grant was revoked lists %v, want nothing", got)
	}
	expectStatus(t, k, secret, accountPath(demo.BondsAccount), http.StatusNotFound, "an account of a revoked grant")

	// One account.
	owner.grant(t, tokenKind, tokenID, identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, string(demo.XTBAccount))
	if got := waitAccounts(t, k, secret, func(ids []string) bool { return len(ids) > 0 }); !slices.Equal(got, []string{string(demo.XTBAccount)}) {
		t.Fatalf("a token granted one account lists %v, want it alone", got)
	}
	expectStatus(t, k, secret, accountPath(demo.IBKRAccount), http.StatusNotFound, "an account next to the granted one")
	owner.revoke(t, tokenKind, tokenID, identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, string(demo.XTBAccount))

	checkGrantAudit(t, owner, tokenID)
}

// checkCutoverKeptTheUI checks that the UI plugin, installed before the core started on SpiceDB, still lists every
// demo account with its own token: the core gave it the household grant at its start (D-209), as it did every
// script token and view plugin that held read access.
func checkCutoverKeptTheUI(t *testing.T, k *kube, owner *ownerAPI) {
	t.Helper()
	token := k.token(t, plugins.ServiceAccountName(uiPlugin), plugins.CoreAudience)
	err := waitFor(t.Context(), spicedbTimeout, func(ctx context.Context) (string, bool, error) {
		ids, err := owner.accountsOf(ctx, token)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("%d accounts", len(ids)), slices.Equal(ids, demoAccountIDs()), nil
	})
	if err != nil {
		t.Fatalf("waiting for the UI plugin to list every demo account: %v", err)
	}
}

// accountPath is the read API's path of an account.
func accountPath(id ledger.AccountID) string { return "/v1/accounts/" + url.PathEscape(string(id)) }

// expectStatus waits until GET path with token answers want.
func expectStatus(t *testing.T, k *kube, token, path string, want int, what string) {
	t.Helper()
	err := waitFor(t.Context(), spicedbTimeout, func(ctx context.Context) (string, bool, error) {
		r, err := call(ctx, k, clientPod, path, token)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		return fmt.Sprintf("HTTP %d", r.status), r.status == want, nil
	})
	if err != nil {
		t.Fatalf("GET %s (%s): %v, want %d", path, what, err, want)
	}
}

// checkSleeveTotals checks that the allocation by sleeve, which totals the accounts a caller may view, holds the
// granted sleeve alone for the sleeve-granted token while the household's token has every sleeve, so the total is
// computed over the accounts the grant covers (D-163).
func checkSleeveTotals(t *testing.T, k *kube, sleeveToken, householdToken string) {
	t.Helper()
	date := warsawToday(t)()[0]
	path := "/v1/allocation?" + url.Values{"by": {viewsv1.AllocationDimension_ALLOCATION_DIMENSION_SLEEVE.String()}, "date": {date}}.Encode()
	groups := func(token string) []string { return sleeveGroups(t, k, path, token) }
	if diff := cmp.Diff([]string{string(demo.DefensiveSleeve)}, groups(sleeveToken)); diff != "" {
		t.Errorf("the sleeve-granted token's allocation by sleeve (-want +got):\n%s", diff)
	}
	want := []string{string(demo.DefensiveSleeve), string(demo.OffensiveSleeve)}
	if diff := cmp.Diff(want, groups(householdToken)); diff != "" {
		t.Errorf("the household token's allocation by sleeve (-want +got):\n%s", diff)
	}
}

// sleeveGroups waits until GET path, an allocation by sleeve, answers token
// with at least one group, and returns the groups' keys, sorted.
func sleeveGroups(t *testing.T, k *kube, path, token string) []string {
	t.Helper()
	var keys []string
	err := waitFor(t.Context(), valuedTimeout, func(ctx context.Context) (string, bool, error) {
		r, err := call(ctx, k, clientPod, path, token)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		if r.status != http.StatusOK {
			return fmt.Sprintf("HTTP %d", r.status), false, nil
		}
		var allocation viewsv1.GetAllocationResponse
		if err := protojson.Unmarshal(r.body, &allocation); err != nil {
			return "", false, fmt.Errorf("decoding the allocation: %w", err)
		}
		keys = keys[:0]
		for _, g := range allocation.GetGroups() {
			keys = append(keys, g.GetKey())
		}
		slices.Sort(keys)
		return fmt.Sprintf("%d groups", len(keys)), len(keys) > 0, nil
	})
	if err != nil {
		t.Fatalf("waiting for the allocation by sleeve: %v", err)
	}
	return slices.Clone(keys)
}

// checkGrantAudit checks that the audit log, read as the owner, shows each grant and revocation of the token,
// created before ended, naming the token, then its target, by kind and ID and nothing else.
func checkGrantAudit(t *testing.T, owner *ownerAPI, tokenID string) {
	t.Helper()
	line := func(e *identityv1.AuditEntry) string {
		words := make([]string, 0, 1+len(e.GetRefs()))
		words = append(words, e.GetActor().GetKind().String())
		for _, r := range e.GetRefs() {
			words = append(words, r.GetKind().String()+":"+r.GetId())
		}
		return strings.Join(words, " ")
	}
	token := " AUDIT_REF_KIND_SCRIPT_TOKEN:" + tokenID
	var created, revoked []string
	for _, e := range slices.Backward(owner.auditEntries(t, identityv1.AuditAction_AUDIT_ACTION_GRANT_CREATED)) {
		if len(e.GetRefs()) == 2 && e.GetRefs()[0].GetId() == tokenID {
			created = append(created, line(e))
		}
	}
	for _, e := range slices.Backward(owner.auditEntries(t, identityv1.AuditAction_AUDIT_ACTION_GRANT_REVOKED)) {
		if len(e.GetRefs()) == 2 && e.GetRefs()[0].GetId() == tokenID {
			revoked = append(revoked, line(e))
		}
	}
	const (
		actor   = "AUDIT_ACTOR_KIND_OWNER"
		sleeve  = " AUDIT_REF_KIND_SLEEVE:" + string(demo.DefensiveSleeve)
		account = " AUDIT_REF_KIND_ACCOUNT:" + string(demo.XTBAccount)
	)
	if diff := cmp.Diff([]string{actor + token + sleeve, actor + token + account}, created); diff != "" {
		t.Errorf("grant_created entries of the token (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{actor + token + sleeve, actor + token + account}, revoked); diff != "" {
		t.Errorf("grant_revoked entries of the token (-want +got):\n%s", diff)
	}
}
