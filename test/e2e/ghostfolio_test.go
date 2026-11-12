//go:build e2e

package e2e

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
	"github.com/arkadiuss/openfief/test/e2e/ghostfoliocheck"
)

// gfReadTimeout bounds one read API call of the check.
const gfReadTimeout = time.Minute

var (
	// errNoBond reports a demo record whose bonds the check never saw.
	errNoBond = errors.New("no treasury bond held in any account")
	// errNoEditable reports owned accounts without an activity holding
	// units, which the check edits.
	errNoEditable = errors.New("no owned activity with units to edit")
)

// TestGhostfolio checks D-200's criterion 9 against a real Ghostfolio
// 3.93.0, in its own shard (task e2e:ghostfolio). It runs Ghostfolio in a
// namespace of its own with its own CloudNativePG database and Redis,
// signs an anonymous user up as an owner does on a fresh Ghostfolio, sets
// the base currency to PLN and puts an activity and a cash balance of the
// owner's own in the account Ghostfolio made. It installs the Ghostfolio
// view with the CLI built from this checkout, the user's security token in
// its Secret and Ghostfolio's Service as its in-cluster service (D-74), with
// no listings, over the demo dataset the suite loads. Then:
//
//   - the first sync creates every owned account, imports every projected
//     activity and sets every projected balance, leaving out only listed
//     instruments, which need a listing; the owned accounts then hold
//     exactly the projection of what the read API answers a script token,
//     each activity naming a transaction of its account, each bond series'
//     units the read API's holding and each owned account's cash the read
//     API's cash in its currency (SCEN-e2e-ghostfolio-holds-projection);
//   - a second sync reports no change and leaves every activity and balance
//     as it was, IDs included (SCEN-e2e-ghostfolio-second-sync-unchanged);
//   - an activity the owner edits by hand in an owned account, its comment
//     kept, is reverted by the next sync, which imports the projected one
//     and deletes the edited one, and the owner's own account still holds
//     exactly their activity and balance (SCEN-e2e-ghostfolio-edit-reverted).
//
// Each sync is the pass a new pod of the view runs at start.
func TestGhostfolio(t *testing.T) {
	t.Parallel()
	if os.Getenv(shardEnv) != shardGhostfolio {
		t.Fatal(errNoShard)
	}
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	deployGhostfolio(t, k)
	waitCoreAvailable(t, k)
	waitSeedComplete(t, k)
	v := newGFView(t, k)
	v.today = v.waitValued(t)
	startGhostfolio(t, k)
	v.owner = signUpOwner(t, k)
	own := v.owner.addOwn(t)
	installView(t, k, reg, v.owner.securityToken)

	pod, err := k.runningPluginPod(t.Context(), gfInstance)
	if err != nil {
		t.Fatal(err)
	}
	first := waitPass(t, k, pod.Name)
	v.checkFirstPass(t, &first)
	before := v.checkHoldsProjection(t)

	// The transactions the projection leaves out stay left out on every pass.
	v.leftOut = first.LeftOut
	second := waitPass(t, k, restartView(t, k))
	if diff := cmp.Diff(ghostfoliocheck.Sync{LeftOut: v.leftOut}, second); diff != "" {
		t.Errorf("the second sync's report (-want +got):\n%s", diff)
	}
	after := v.checkHoldsProjection(t)
	if diff := cmp.Diff(before, after); diff != "" {
		t.Errorf("the owned accounts after the second sync (-before +after):\n%s", diff)
	}

	v.checkEditReverted(t, &after)
	if diff := cmp.Diff(own, v.owner.ownNow(t, own.account), cmp.AllowUnexported(gfOwn{})); diff != "" {
		t.Errorf("the owner's own account (-before +after the syncs):\n%s", diff)
	}
}

// gfView is what TestGhostfolio checks the view against: the read API, as
// a script token reads it, Ghostfolio, as the owner reads it, and the day
// the record is valued through.
type gfView struct {
	k      *kube
	record viewsv1connect.ReadServiceClient
	owner  *gfOwner
	today  kernel.Date
	// leftOut is how many transactions the first pass left out of the
	// projection, which every later pass leaves out too.
	leftOut int
}

// gfState is what the owned accounts hold, a line per activity and per
// balance with its ID.
type gfState struct {
	Activities, Balances []string
}

// newGFView is the check, reading the read API through its Connect client,
// a port-forward to the core's browser port and a script token.
func newGFView(t *testing.T, k *kube) *gfView {
	t.Helper()
	return &gfView{k: k, record: viewsv1connect.NewReadServiceClient(&http.Client{Timeout: gfReadTimeout}, forwardCore(t, k),
		connectrpc.WithInterceptors(withBearer(issueToken(t, k))))}
}

// waitValued waits until every demo account is valued through today, so
// the record the view projects no longer changes, and returns that day.
func (v *gfView) waitValued(t *testing.T) kernel.Date {
	t.Helper()
	today := warsawToday(t)
	err := waitFor(t.Context(), valuedTimeout, func(ctx context.Context) (string, bool, error) {
		resp, err := v.record.ListValuationStatuses(ctx, connectrpc.NewRequest(&viewsv1.ListValuationStatusesRequest{}))
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		return valuationState(resp.Msg), allValued(resp.Msg, demoAccountIDs(), today()), nil
	})
	if err != nil {
		t.Fatalf("waiting for every demo account to be valued through today: %v", err)
	}
	d, err := kernel.ParseDate(today()[0])
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// project returns the view's projection of what the read API answers now,
// for the owner's Ghostfolio user, and the transactions it leaves out.
func (v *gfView) project(t *testing.T) ([]ghostfolio.SourceAccount, ghostfolio.Desired, []ghostfolio.Problem) {
	t.Helper()
	accounts, err := ghostfolio.ReadRecord(t.Context(), v.record)
	if err != nil {
		t.Fatalf("reading the record: %v", err)
	}
	u := v.owner.user
	desired, problems := ghostfolio.Project(&ghostfolio.Source{
		Instance: gfInstance, User: u.ID, Currency: u.BaseCurrency, Accounts: accounts,
	})
	return accounts, desired, problems
}

// checkFirstPass checks the first sync's report against the projection: it
// created every owned account, imported every activity, set every balance,
// changed nothing else, and left out only transactions of listed
// instruments, of which the demo has some.
func (v *gfView) checkFirstPass(t *testing.T, pass *ghostfoliocheck.Sync) {
	t.Helper()
	_, desired, problems := v.project(t)
	want := ghostfoliocheck.Sync{LeftOut: len(problems), Report: ghostfoliocheck.Report{
		AccountsCreated: len(desired.Accounts), Imported: len(desired.Activities), BalancesSet: len(desired.Balances),
	}}
	if diff := cmp.Diff(want, *pass); diff != "" {
		t.Errorf("the first sync's report (-want +got):\n%s", diff)
	}
	if len(desired.Activities) == 0 || len(problems) == 0 {
		t.Errorf("the projection has %d activities and leaves out %d transactions; want both", len(desired.Activities), len(problems))
	}
	for _, p := range problems {
		if p.Reason != ghostfolio.ProblemUnlisted {
			t.Errorf("transaction %s left out: %v; want only unlisted instruments left out", p.Transaction, p.Reason)
		}
	}
}

// checkHoldsProjection checks that the owned accounts hold exactly the
// projection of the record, and what the read API answers on its own, and
// returns what they hold.
func (v *gfView) checkHoldsProjection(t *testing.T) gfState {
	t.Helper()
	accounts, desired, _ := v.project(t)
	actual := v.actual(t, &desired)
	if plan := ghostfolio.MakePlan(&desired, &actual); !plan.IsEmpty() {
		t.Errorf("the owned accounts differ from the projection: %d accounts to create, %d activities to import, "+
			"%d and %d to delete, %d accounts to clear, %d balances to set, %d to delete",
			len(plan.CreateAccounts), len(plan.Import), len(plan.PreDelete), len(plan.Delete), len(plan.Clear),
			len(plan.SetBalances), len(plan.DeleteBalances))
	}
	acts, bs := checkActivities(actual.Activities), checkBalances(actual.Balances)
	v.checkComments(t, accounts, acts)
	v.checkBonds(t, accounts, acts)
	v.checkCash(t, accounts, &desired, bs)
	return gfState{Activities: ghostfoliocheck.Describe(acts), Balances: ghostfoliocheck.DescribeBalances(bs)}
}

// actual reads which accounts the user has and what the owned ones hold.
func (v *gfView) actual(t *testing.T, desired *ghostfolio.Desired) ghostfolio.Actual {
	t.Helper()
	c := v.owner.client
	ids, err := c.Accounts(t.Context())
	if err != nil {
		t.Fatalf("listing Ghostfolio's accounts: %v", err)
	}
	act := ghostfolio.Actual{Accounts: ids}
	var owned []string
	for _, a := range desired.Accounts {
		if slices.Contains(ids, a.ID) {
			owned = append(owned, a.ID)
		}
	}
	if len(owned) == 0 {
		return act
	}
	if act.Activities, err = c.Activities(t.Context(), owned); err != nil {
		t.Fatalf("listing the owned accounts' activities: %v", err)
	}
	for _, id := range owned {
		bs, err := c.Balances(t.Context(), id)
		if err != nil {
			t.Fatalf("listing an owned account's balances: %v", err)
		}
		act.Balances = append(act.Balances, bs...)
	}
	return act
}

// checkComments checks that each activity is in the owned account of a
// record account, never in one that holds only cash, and that its comment
// names one of that account's current transactions (D-4).
func (v *gfView) checkComments(t *testing.T, accounts []ghostfolio.SourceAccount, acts []ghostfoliocheck.Activity) {
	t.Helper()
	keys := map[string]map[string]bool{}
	for i := range accounts {
		a := &accounts[i]
		owned := map[string]bool{}
		for j := range a.Transactions {
			owned[a.Transactions[j].Key.String()] = true
		}
		keys[ghostfolio.AccountUUID(gfInstance, a.ID)] = owned
	}
	for i := range acts {
		a := &acts[i]
		key, ok := ghostfoliocheck.CommentKey(a.Comment)
		if !ok || !keys[a.AccountID][key] {
			t.Errorf("activity %s in account %s: comment %q names no current transaction of its account", a.ID, a.AccountID, a.Comment)
		}
	}
}

// checkBonds checks that, for each treasury bond the read API says an
// account holds today, its owned account's activities on the series'
// MANUAL asset hold the same units (D-64), and that the demo holds one.
func (v *gfView) checkBonds(t *testing.T, accounts []ghostfolio.SourceAccount, acts []ghostfoliocheck.Activity) {
	t.Helper()
	held := 0
	for i := range accounts {
		id := accounts[i].ID
		resp, err := v.record.ListHoldings(t.Context(), connectrpc.NewRequest(&viewsv1.ListHoldingsRequest{
			AccountId: string(id), Date: v.today.String(),
		}))
		if err != nil {
			t.Fatalf("listing the holdings of %s: %v", id, err)
		}
		for _, h := range resp.Msg.GetHoldings() {
			in, err := ledger.ParseISIN(h.GetInstrument().GetIsin())
			if err != nil {
				t.Fatalf("holding %v: %v", h.GetInstrument(), err)
			}
			series := h.GetInstrument().GetSymbol()
			if views.AssetClassOf(ledger.Instrument{ISIN: in, Symbol: series}) != views.AssetTreasuryBond {
				continue
			}
			held++
			want := decimal(t, h.GetQuantity())
			symbol := ghostfolio.BondSymbol(gfInstance, v.owner.user.ID, series)
			if got := ghostfoliocheck.NetUnits(acts, ghostfolio.AccountUUID(gfInstance, id), symbol); !got.Equal(want) {
				t.Errorf("account %s, series %s: Ghostfolio holds %s units, the read API %s", id, series, got, want)
			}
		}
	}
	if held == 0 {
		t.Error(errNoBond)
	}
}

// checkCash checks that each owned account's cash today is the read API's
// cash of its account in its currency, zero for a currency the account no
// longer holds (D-108, D-133), and that no currency of the cash lacks an
// owned account.
func (v *gfView) checkCash(t *testing.T, accounts []ghostfolio.SourceAccount, desired *ghostfolio.Desired, bs []ghostfoliocheck.Balance) {
	t.Helper()
	currencies := map[string]kernel.Currency{}
	for _, a := range desired.Accounts {
		currencies[a.ID] = a.Currency
	}
	for i := range accounts {
		id := accounts[i].ID
		cash := v.cashOf(t, id)
		for owned, amount := range cash {
			if _, ok := currencies[owned]; !ok && !amount.IsZero() {
				t.Errorf("account %s: no owned account holds its cash %s", id, amount)
			}
		}
		for owned := range cashAccountsOf(id, desired, v.owner.user.BaseCurrency) {
			if got := ghostfoliocheck.Latest(bs, owned, v.today); !got.Equal(cash[owned]) {
				t.Errorf("account %s: owned account %s (%s) holds %s cash, the read API %s", id, owned, currencies[owned], got, cash[owned])
			}
		}
	}
}

// cashOf returns the read API's cash of account today, each amount under
// the owned account that holds that currency: the account's own for the
// base currency, one of its own for any other (D-133).
func (v *gfView) cashOf(t *testing.T, account ledger.AccountID) map[string]kernel.Decimal {
	t.Helper()
	resp, err := v.record.ListCash(t.Context(), connectrpc.NewRequest(&viewsv1.ListCashRequest{AccountId: string(account), Date: v.today.String()}))
	if err != nil {
		t.Fatalf("listing the cash of %s: %v", account, err)
	}
	cash := map[string]kernel.Decimal{}
	for _, m := range resp.Msg.GetBalances() {
		c, err := kernel.ParseCurrency(m.GetCurrency())
		if err != nil {
			t.Fatalf("cash of %s: %v", account, err)
		}
		owned := ghostfolio.CashAccountUUID(gfInstance, account, c)
		if c.Equal(v.owner.user.BaseCurrency) {
			owned = ghostfolio.AccountUUID(gfInstance, account)
		}
		cash[owned] = decimal(t, m.GetAmount())
	}
	return cash
}

// cashAccountsOf returns the owned accounts that hold account's cash: its
// own and one per other currency of its cash (D-133).
func cashAccountsOf(account ledger.AccountID, desired *ghostfolio.Desired, base kernel.Currency) map[string]bool {
	out := map[string]bool{ghostfolio.AccountUUID(gfInstance, account): true}
	for _, a := range desired.Accounts {
		if !a.Currency.Equal(base) && a.ID == ghostfolio.CashAccountUUID(gfInstance, account, a.Currency) {
			out[a.ID] = true
		}
	}
	return out
}

// checkEditReverted edits one owned activity by hand, as the owner would in
// Ghostfolio's UI, its comment kept, and checks that the next sync imports
// the projected activity and deletes the edited one, and nothing else, and
// that the owned accounts then hold exactly the projection again: every
// activity as it was before the edit, but the reverted one under a new ID.
func (v *gfView) checkEditReverted(t *testing.T, before *gfState) {
	t.Helper()
	_, desired, _ := v.project(t)
	actual := v.actual(t, &desired)
	edited, ok := ghostfoliocheck.Editable(checkActivities(actual.Activities))
	if !ok {
		t.Fatal(errNoEditable)
	}
	v.owner.send(t, http.MethodPut, "/api/v1/activities/"+edited.ID, ghostfoliocheck.Edited(&edited), nil)

	pass := waitPass(t, v.k, restartView(t, v.k))
	if diff := cmp.Diff(ghostfoliocheck.Sync{Report: ghostfoliocheck.Report{Imported: 1, Deleted: 1}, LeftOut: v.leftOut}, pass); diff != "" {
		t.Errorf("the sync after the edit, report (-want +got):\n%s", diff)
	}
	after := v.checkHoldsProjection(t)
	gone := ghostfoliocheck.Describe([]ghostfoliocheck.Activity{edited})
	removed, added := linesNotIn(before.Activities, after.Activities), linesNotIn(after.Activities, before.Activities)
	if diff := cmp.Diff(gone, removed); diff != "" {
		t.Errorf("activities gone after the revert (-want +got):\n%s", diff)
	}
	if len(added) != 1 {
		t.Errorf("activities new after the revert: %v; want the reverted one alone", added)
	}
	if diff := cmp.Diff(before.Balances, after.Balances); diff != "" {
		t.Errorf("the owned balances after the revert (-before +after):\n%s", diff)
	}
}

// linesNotIn returns the lines of from that in lacks.
func linesNotIn(from, in []string) []string {
	var out []string
	for _, l := range from {
		if !slices.Contains(in, l) {
			out = append(out, l)
		}
	}
	return out
}

// checkActivities are acts as ghostfoliocheck reads them.
func checkActivities(acts []ghostfolio.Activity) []ghostfoliocheck.Activity {
	out := make([]ghostfoliocheck.Activity, 0, len(acts))
	for i := range acts {
		a := &acts[i]
		out = append(out, ghostfoliocheck.Activity{
			ID: a.ID, AccountID: a.AccountID, Comment: a.Comment, Type: string(a.Type), Date: a.Date, Currency: a.Currency,
			DataSource: a.DataSource, Symbol: a.Symbol, Quantity: a.Quantity, UnitPrice: a.UnitPrice, Fee: a.Fee,
		})
	}
	return out
}

// checkBalances are bs as ghostfoliocheck reads them.
func checkBalances(bs []ghostfolio.Balance) []ghostfoliocheck.Balance {
	out := make([]ghostfoliocheck.Balance, 0, len(bs))
	for i := range bs {
		b := &bs[i]
		out = append(out, ghostfoliocheck.Balance{ID: b.ID, AccountID: b.AccountID, Date: b.Date, Value: b.Value})
	}
	return out
}
