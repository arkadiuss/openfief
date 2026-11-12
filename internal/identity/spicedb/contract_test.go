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

// grant writes the relationship resource#relation@subject, each "type:id",
// as the owner's grants are.
func (c *contract) grant(resource, relation, subject string) {
	c.t.Helper()
	res, sub := splitRef(resource), splitRef(subject)
	_, err := c.client.WriteRelationships(c.t.Context(), &v1.WriteRelationshipsRequest{Updates: []*v1.RelationshipUpdate{{
		Operation: v1.RelationshipUpdate_OPERATION_TOUCH,
		Relationship: &v1.Relationship{
			Resource: &v1.ObjectReference{ObjectType: res[0], ObjectId: res[1]},
			Relation: relation,
			Subject:  ref(sub[0], sub[1]),
		},
	}}})
	if err != nil {
		c.t.Fatalf("writing %s#%s@%s: %v", resource, relation, subject, err)
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
	c.reconcile(spicedb.Desired{
		Accounts:     accounts,
		Sleeves:      map[string]string{"brokerage": "core", "bonds": "bonds"}, // two accounts are in no sleeve
		DemoSessions: true,
	})
	// The household grant (D-209) of the script token and the view plugin that
	// read today.
	c.grant("household:home", "reader", "script_token:0123456789abcdef")
	c.grant("household:home", "reader", "view_plugin:ui")

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
	c.grant("sleeve:core", "viewer", "script_token:00000000000000aa")
	c.grant("household:home", "reader", "script_token:00000000000000bb")
	c.grant("account:bonds", "viewer", "view_plugin:one-account")

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
	c.grant("sleeve:play", "viewer", "view_plugin:one-account")
	if err := c.syncer.AccountPlaced(t.Context(), "brokerage", "play"); err != nil {
		t.Fatalf("AccountPlaced move: %v", err)
	}
	check("a move ends the old sleeve's grant", sleeveToken, view, "brokerage", false)
	check("a move starts the new sleeve's grant", accountPlugin, view, "brokerage", true)
	check("a sleeve grant never changes", accountPlugin, change, "brokerage", false)

	// A deleted account: reconcile removes its household, so nobody views it.
	c.reconcile(spicedb.Desired{Accounts: []string{"bonds"}})
	check("a deleted account is nobody's", owner, view, "brokerage", false)
	check("the household grant of another account stays", householdPerson, view, "bonds", true)
}

// TestContractReconcileLeavesGrantsAndIsIdempotent checks that reconciling
// twice writes nothing the second time and removes no grant, against the
// real SpiceDB's reads.
func TestContractReconcileLeavesGrantsAndIsIdempotent(t *testing.T) {
	t.Parallel()
	c := newContract(t)
	d := spicedb.Desired{Accounts: []string{"a", "b"}, Sleeves: map[string]string{"a": "core"}, DemoSessions: true}
	c.reconcile(d)
	c.grant("household:home", "reader", "script_token:00000000000000aa")
	c.grant("account:a", "viewer", "person:viewer")
	c.reconcile(d)
	c.reconcile(d)

	var got []string
	for _, f := range []*v1.RelationshipFilter{{ResourceType: "household"}, {ResourceType: "account"}} {
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
		"account:a#household@household:home", "account:a#sleeve@sleeve:core", "account:a#viewer@person:viewer",
		"account:b#household@household:home",
		"household:home#owner@person:owner", "household:home#reader@demo_session:*", "household:home#reader@script_token:00000000000000aa",
	}
	if diff := cmp.Diff(sorted(want), sorted(got)); diff != "" {
		t.Errorf("relationships (-want +got):\n%s", diff)
	}
}
