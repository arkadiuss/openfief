//go:build spicedb

package spicedb_test

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/spicedb"
)

// These tests run against a real SpiceDB: `task test:authz` starts the
// pinned one in testing mode, whose datastore is in memory and per preshared
// key, so each test below gets its own, and runs them with the spicedb build
// tag. OPENFIEF_SPICEDB_TEST_ENDPOINT names its gRPC address.

// contractEnv names the environment variable with SpiceDB's address.
const contractEnv = "OPENFIEF_SPICEDB_TEST_ENDPOINT"

// authzSchemaPath is the authorization model, read from where this package
// lives.
const authzSchemaPath = "../../../authz/schema.zed"

// contract is SpiceDB with authz/schema.zed written to a datastore of its
// own, and the adapters over it.
type contract struct {
	t      *testing.T
	client *authzed.Client
	authz  *spicedb.Authorizer
	syncer *spicedb.Syncer
}

func newContract(t *testing.T) *contract {
	t.Helper()
	endpoint := os.Getenv(contractEnv)
	if endpoint == "" {
		t.Fatalf("%s is not set; run these tests with task test:authz", contractEnv)
	}
	// A key of this test's own is its own datastore on a testing server.
	key := "contract-" + strings.NewReplacer("/", "-", " ", "-").Replace(t.Name())
	client, err := authzed.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()), grpcutil.WithInsecureBearerToken(key))
	if err != nil {
		t.Fatalf("authzed.NewClient: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Errorf("Close: %v", closeErr)
		}
	})
	schema, err := os.ReadFile(authzSchemaPath)
	if err != nil {
		t.Fatalf("reading the schema: %v", err)
	}
	if _, err := client.WriteSchema(t.Context(), &v1.WriteSchemaRequest{Schema: string(schema)}); err != nil {
		t.Fatalf("WriteSchema: %v", err)
	}
	return &contract{
		t: t, client: client,
		authz:  spicedb.NewAuthorizer(client.PermissionsServiceClient),
		syncer: spicedb.NewSyncer(client.PermissionsServiceClient),
	}
}

// reconcile has the syncer write d.
func (c *contract) reconcile(d spicedb.Desired) {
	c.t.Helper()
	if err := c.syncer.Reconcile(c.t.Context(), fixed(d)); err != nil {
		c.t.Fatalf("Reconcile: %v", err)
	}
}

// contractAt is when the contract's grants are made.
func contractAt() time.Time { return time.Date(2026, 11, 13, 8, 0, 0, 0, time.UTC) }

// viewerGrant is the grant of target to the holder of kind and id.
func viewerGrant(kind identity.GrantHolderKind, id string, target identity.GrantTarget) identity.Grant {
	return identity.Grant{Holder: identity.GrantHolder{Kind: kind, ID: id}, Target: target, CreatedAt: contractAt()}
}

func accountTarget(id string) identity.GrantTarget {
	return identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: id}
}

func sleeveTarget(id string) identity.GrantTarget {
	return identity.GrantTarget{Kind: identity.GrantTargetSleeve, ID: id}
}

// project passes grants to SpiceDB, as the grants use cases do after they
// commit.
func (c *contract) project(grants ...identity.Grant) {
	c.t.Helper()
	if err := c.syncer.ProjectGrants(c.t.Context(), grants); err != nil {
		c.t.Fatalf("ProjectGrants: %v", err)
	}
}

// unproject removes grants from SpiceDB, as a revocation does.
func (c *contract) unproject(grants ...identity.Grant) {
	c.t.Helper()
	if err := c.syncer.UnprojectGrants(c.t.Context(), grants); err != nil {
		c.t.Fatalf("UnprojectGrants: %v", err)
	}
}

// may asks the authorizer.
func (c *contract) may(caller identity.Caller, a identity.Action, account string) bool {
	c.t.Helper()
	ok, err := c.authz.MayAct(c.t.Context(), caller, a, account)
	if err != nil {
		c.t.Fatalf("MayAct(%+v, %s, %q): %v", caller, a, account, err)
	}
	return ok
}

// listed asks the authorizer for a list.
func (c *contract) listed(caller identity.Caller, a identity.Action) []string {
	c.t.Helper()
	set, err := c.authz.ActableAccounts(c.t.Context(), caller, a)
	if err != nil {
		c.t.Fatalf("ActableAccounts(%+v, %s): %v", caller, a, err)
	}
	return slices.Sorted(maps.Keys(set))
}

func sorted(s []string) []string { return slices.Sorted(slices.Values(s)) }

// contractHolders is a holder of every kind, two of them more than once, and
// nobody.
func contractHolders() []identity.Caller {
	return []identity.Caller{
		{Holder: identity.HolderOwner, ID: "session-1"},
		{Holder: identity.HolderOwner, ID: "session-2"},
		{Holder: identity.HolderScriptToken, ID: "0123456789abcdef"},
		{Holder: identity.HolderViewPlugin, ID: "ui"},
		{Holder: identity.HolderDemoSession, ID: "demo-1"},
		{Holder: identity.HolderDemoSession, ID: "demo-2"},
		{},
	}
}

// compareAnswers checks SpiceDB's answer for h, every action (the zero one
// too) and every account, against identity.MayAct.
func (c *contract) compareAnswers(h identity.Caller, accounts []string) {
	c.t.Helper()
	for _, a := range []identity.Action{identity.ActionView, identity.ActionChange, 0} {
		for _, account := range accounts {
			if got, want := c.may(h, a, account), identity.MayAct(h, a); got != want {
				c.t.Errorf("SpiceDB: %s %q may %s %q = %v, identity.MayAct says %v", h.Holder, h.ID, a, account, got, want)
			}
		}
	}
}

// compareLists checks SpiceDB's list for h against the accounts identity.MayAct
// allows one by one.
func (c *contract) compareLists(h identity.Caller, accounts []string) {
	c.t.Helper()
	for _, a := range []identity.Action{identity.ActionView, identity.ActionChange} {
		var want []string
		if identity.MayAct(h, a) {
			want = sorted(accounts)
		}
		if diff := cmp.Diff(want, c.listed(h, a)); diff != "" {
			c.t.Errorf("SpiceDB lists for %s %q to %s (-want +got):\n%s", h.Holder, h.ID, a, diff)
		}
	}
}

// TestContractSpiceDBMatchesIdentityMayAct is the contract of the
// authorization interface (D-163, D-164): with the household grants written
// for the holders of read access that exist today, SpiceDB answers every
// holder's every action on every account as identity.MayAct does, and lists
// the accounts it may view as MayAct would one by one.
func TestContractSpiceDBMatchesIdentityMayAct(t *testing.T) {
	t.Parallel()
	c := newContract(t)
	accounts := []string{"brokerage", "bonds", "IKE (PL)", "ikze/2"}
	// The household grant (D-209) of the script token and the view plugin that
	// read today, which the cutover writes to the grants table.
	c.reconcile(spicedb.Desired{
		Accounts:     accounts,
		Sleeves:      map[string]string{"brokerage": "core", "bonds": "bonds"}, // two accounts are in no sleeve
		DemoSessions: true,
		Grants: []identity.Grant{
			viewerGrant(identity.GrantHolderScriptToken, "0123456789abcdef", identity.Household()),
			viewerGrant(identity.GrantHolderViewPlugin, "ui", identity.Household()),
		},
	})

	for _, h := range contractHolders() {
		c.compareAnswers(h, accounts)
		c.compareLists(h, accounts)
	}
	// Without a household grant, a script token or view plugin views nothing:
	// the cutover (D-209) writes one for each before SpiceDB answers.
	for _, h := range []identity.Caller{
		{Holder: identity.HolderScriptToken, ID: "fedcba9876543210"},
		{Holder: identity.HolderViewPlugin, ID: "other"},
	} {
		if c.may(h, identity.ActionView, "bonds") {
			t.Errorf("%s %q views an account with no grant", h.Holder, h.ID)
		}
	}
}

// TestContractGrantsFollowTheSchema checks, through the adapters, what
// authz/validation.yaml asserts offline (D-209): a sleeve grant covers an
// account placed in the sleeve later and stops when it is taken out, a
// household grant covers an account added later, and a viewer grant only
// ever gives view.
func TestContractGrantsFollowTheSchema(t *testing.T) {
	t.Parallel()
	c := newContract(t)
	view, change := identity.ActionView, identity.ActionChange
	sleeveToken := identity.Caller{Holder: identity.HolderScriptToken, ID: "00000000000000aa"}
	householdPerson := identity.Caller{Holder: identity.HolderScriptToken, ID: "00000000000000bb"}
	accountPlugin := identity.Caller{Holder: identity.HolderViewPlugin, ID: "one-account"}
	owner := identity.Caller{Holder: identity.HolderOwner, ID: "s"}

	d := spicedb.Desired{Accounts: []string{"brokerage", "bonds"}, Sleeves: map[string]string{"brokerage": "core"}}
	c.reconcile(d)
	sleeveGrant := viewerGrant(identity.GrantHolderScriptToken, "00000000000000aa", sleeveTarget("core"))
	householdGrant := viewerGrant(identity.GrantHolderScriptToken, "00000000000000bb", identity.Household())
	accountGrant := viewerGrant(identity.GrantHolderViewPlugin, "one-account", accountTarget("bonds"))
	c.project(sleeveGrant, householdGrant, accountGrant)

	check := func(step string, who identity.Caller, a identity.Action, account string, want bool) {
		t.Helper()
		if got := c.may(who, a, account); got != want {
			t.Errorf("%s: %s %q may %s %s = %v, want %v", step, who.Holder, who.ID, a, account, got, want)
		}
	}
	check("sleeve grant", sleeveToken, view, "brokerage", true)
	check("sleeve grant stops at its accounts", sleeveToken, view, "bonds", false)
	check("household grant", householdPerson, view, "brokerage", true)
	check("household grant", householdPerson, view, "bonds", true)
	check("account grant", accountPlugin, view, "bonds", true)
	check("account grant stops at its account", accountPlugin, view, "brokerage", false)
	for _, who := range []identity.Caller{sleeveToken, householdPerson, accountPlugin} {
		check("a viewer grant never changes", who, change, "brokerage", false)
		check("a viewer grant never changes", who, change, "bonds", false)
	}

	// An account added later: placed in the sleeve through the syncer, as the
	// registry does, and one that is in none.
	if err := c.syncer.AccountPlaced(t.Context(), "later", "core"); err != nil {
		t.Fatalf("AccountPlaced: %v", err)
	}
	if err := c.syncer.AccountPlaced(t.Context(), "unplaced", ""); err != nil {
		t.Fatalf("AccountPlaced without a sleeve: %v", err)
	}
	check("sleeve grant covers a later account", sleeveToken, view, "later", true)
	check("sleeve grant skips an account in no sleeve", sleeveToken, view, "unplaced", false)
	check("household grant covers a later account", householdPerson, view, "later", true)
	check("household grant covers an account in no sleeve", householdPerson, view, "unplaced", true)
	check("the owner changes a later account", owner, change, "later", true)
	check("the owner changes an account in no sleeve", owner, change, "unplaced", true)
	check("a later account is not an account grant's", accountPlugin, view, "later", false)

	// Taken out of its sleeve, an account is the sleeve grant's no more, and
	// the owner keeps it.
	if err := c.syncer.AccountUnplaced(t.Context(), "later"); err != nil {
		t.Fatalf("AccountUnplaced: %v", err)
	}
	check("a removed sleeve tuple removes access", sleeveToken, view, "later", false)
	check("the household grant stays", householdPerson, view, "later", true)
	check("the owner keeps the account", owner, change, "later", true)

	// Moved to another sleeve, the first sleeve's grant ends with it.
	playGrant := viewerGrant(identity.GrantHolderViewPlugin, "one-account", sleeveTarget("play"))
	c.project(playGrant)
	if err := c.syncer.AccountPlaced(t.Context(), "brokerage", "play"); err != nil {
		t.Fatalf("AccountPlaced move: %v", err)
	}
	check("a move ends the old sleeve's grant", sleeveToken, view, "brokerage", false)
	check("a move starts the new sleeve's grant", accountPlugin, view, "brokerage", true)
	check("a sleeve grant never changes", accountPlugin, change, "brokerage", false)

	// A revoked grant: the relationship is removed, and the holder views no more
	// of what only that grant gave it. The other grants stay.
	c.unproject(householdGrant)
	check("a revoked household grant ends", householdPerson, view, "bonds", false)
	check("a revoked household grant ends for a later account too", householdPerson, view, "later", false)
	check("another holder's grant stays", accountPlugin, view, "bonds", true)
	c.project(householdGrant) // granted again
	check("a grant made again", householdPerson, view, "bonds", true)

	// A deleted account: reconcile removes its household, so nobody views it.
	d = spicedb.Desired{Accounts: []string{"bonds"}, Grants: []identity.Grant{householdGrant, accountGrant}}
	c.reconcile(d)
	check("a deleted account is nobody's", owner, view, "brokerage", false)
	check("the household grant of another account stays", householdPerson, view, "bonds", true)
	check("a grant not in the table is removed by the reconcile", sleeveToken, view, "bonds", false)
}

// TestContractReconcileWritesGrantsAndIsIdempotent checks that reconciling
// writes the grants of the table, removes one SpiceDB holds that the table
// does not, and writes nothing the second time, against the real SpiceDB's
// reads.
func TestContractReconcileWritesGrantsAndIsIdempotent(t *testing.T) {
	t.Parallel()
	c := newContract(t)
	tokenGrant := viewerGrant(identity.GrantHolderScriptToken, "00000000000000aa", identity.Household())
	personGrant := viewerGrant(identity.GrantHolderPerson, "00000000000000cc", accountTarget("a"))
	ended := viewerGrant(identity.GrantHolderViewPlugin, "ui", sleeveTarget("core"))
	c.project(ended) // SpiceDB still holds a grant the table ended
	d := spicedb.Desired{Accounts: []string{"a", "b"}, Sleeves: map[string]string{"a": "core"}, DemoSessions: true, Grants: []identity.Grant{tokenGrant, personGrant}}
	c.reconcile(d)
	c.reconcile(d)
	c.reconcile(d)

	var got []string
	for _, f := range []*v1.RelationshipFilter{{ResourceType: "household"}, {ResourceType: "account"}, {ResourceType: "sleeve"}} {
		stream, err := c.client.ReadRelationships(context.Background(), &v1.ReadRelationshipsRequest{
			Consistency: &v1.Consistency{Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true}}, RelationshipFilter: f,
		})
		if err != nil {
			t.Fatalf("ReadRelationships: %v", err)
		}
		for {
			r, err := stream.Recv()
			if err != nil {
				break
			}
			rel := r.GetRelationship()
			got = append(got, fmt.Sprintf("%s:%s#%s@%s:%s", rel.GetResource().GetObjectType(), rel.GetResource().GetObjectId(),
				rel.GetRelation(), rel.GetSubject().GetObject().GetObjectType(), rel.GetSubject().GetObject().GetObjectId()))
		}
	}
	want := []string{
		"account:a#household@household:home", "account:a#sleeve@sleeve:core", "account:a#viewer@person:00000000000000cc",
		"account:b#household@household:home",
		"household:home#owner@person:owner", "household:home#reader@demo_session:*", "household:home#reader@script_token:00000000000000aa",
	}
	if diff := cmp.Diff(sorted(want), sorted(got)); diff != "" {
		t.Errorf("relationships (-want +got):\n%s", diff)
	}
}
