package ghostfolio_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
)

// Property tests draw synthetic cases from a seeded generator, so every run
// checks the same cases and a failure names the seed that broke.
const (
	propertySeed  = 20261029
	propertyCases = 500
)

// foreignAccount is a Ghostfolio account the view doesn't own.
const foreignAccount = "6f1d3c2b-0000-4000-8000-0000000000ff"

// rng is splitmix64: a tiny deterministic generator, so the tests need no
// math/rand.
type rng struct{ state uint64 }

func (r *rng) next() uint64 {
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// intn returns a number in [0, n), for n > 0.
func (r *rng) intn(n int) int {
	v := int(r.next() >> 1) // fits a 64-bit int
	return v % n
}

// ghostfolioModel is a model of what Ghostfolio does with the calls a plan
// makes: the user's accounts and activities, with IDs from a counter.
type ghostfolioModel struct {
	accounts   []string
	activities []ghostfolio.Activity
	next       int
}

// duplicate is Ghostfolio's import rule (R-15): an activity is skipped when
// one the user has agrees in comment, currency, data source, date, fee,
// quantity, symbol, type and unit price, in any account.
func duplicate(a, b *ghostfolio.Activity) bool {
	return a.Comment == b.Comment && a.Currency.Equal(b.Currency) && a.DataSource == b.DataSource && a.Date == b.Date &&
		a.Fee.Equal(b.Fee) && a.Quantity.Equal(b.Quantity) && a.Symbol == b.Symbol && a.Type == b.Type && a.UnitPrice.Equal(b.UnitPrice)
}

// add stores a with the next ID.
func (m *ghostfolioModel) add(a *ghostfolio.Activity) {
	m.next++
	stored := *a
	stored.ID = fmt.Sprintf("g-%03d", m.next)
	m.activities = append(m.activities, stored)
}

// apply carries out p as Syncer does, in its order, and fails t if the plan
// touches an account that isn't owned.
func (m *ghostfolioModel) apply(t *testing.T, p *ghostfolio.Plan, owned map[string]bool) {
	t.Helper()
	for _, a := range p.CreateAccounts {
		m.accounts = append(m.accounts, a.ID)
	}
	m.remove(t, p.PreDelete, owned)
	for i := range p.Import {
		a := &p.Import[i]
		if !slices.ContainsFunc(m.activities, func(b ghostfolio.Activity) bool { return duplicate(a, &b) }) {
			m.add(a)
		}
	}
	m.remove(t, p.Delete, owned)
	for _, id := range p.Clear {
		if !owned[id] {
			t.Fatalf("bulk delete of account %s, which is not owned", id)
		}
		m.activities = slices.DeleteFunc(m.activities, func(b ghostfolio.Activity) bool { return b.AccountID == id })
	}
}

// remove deletes acts by ID, failing t for one outside the owned accounts.
func (m *ghostfolioModel) remove(t *testing.T, acts []ghostfolio.Activity, owned map[string]bool) {
	t.Helper()
	for i := range acts {
		if !owned[acts[i].AccountID] {
			t.Fatalf("delete of %s in an account not owned", acts[i].ID)
		}
		id := acts[i].ID
		m.activities = slices.DeleteFunc(m.activities, func(b ghostfolio.Activity) bool { return b.ID == id })
	}
}

// in returns m's activities in the accounts keep accepts.
func (m *ghostfolioModel) in(keep func(account string) bool) []ghostfolio.Activity {
	var out []ghostfolio.Activity
	for i := range m.activities {
		if keep(m.activities[i].AccountID) {
			out = append(out, m.activities[i])
		}
	}
	return out
}

// actual is what Syncer reads from m.
func (m *ghostfolioModel) actual(owned map[string]bool) ghostfolio.Actual {
	return ghostfolio.Actual{Accounts: slices.Clone(m.accounts), Activities: m.in(func(id string) bool { return owned[id] })}
}

// generator draws one random case.
type generator struct {
	t     *testing.T
	r     *rng
	start kernel.Date
	d     ghostfolio.Desired
	m     *ghostfolioModel
}

// activity is a random activity with comment c in a random owned account.
func (g *generator) activity(c string) ghostfolio.Activity {
	r := g.r
	return ghostfolio.Activity{
		AccountID: g.d.Accounts[r.intn(len(g.d.Accounts))].ID, Comment: c,
		Type: []ghostfolio.ActivityType{ghostfolio.TypeBuy, ghostfolio.TypeSell}[r.intn(2)],
		Date: g.start.AddDays(r.intn(3)), Currency: cur(g.t, []string{"PLN", "USD"}[r.intn(2)]),
		DataSource: yahoo, Symbol: []string{"ETF1.L", "ETF2.L"}[r.intn(2)],
		Quantity: kernel.NewDecimal(int64(1+r.intn(3)), 0), UnitPrice: kernel.NewDecimal(int64(100+r.intn(3)), -1),
	}
}

// actualFor stores in the model what Ghostfolio holds for desired a: in
// turn nothing, an edited copy, two copies, a copy in another owned
// account, or a itself.
func (g *generator) actualFor(a *ghostfolio.Activity) {
	const ways = 6
	switch g.r.intn(ways) {
	case 0:
	case 1:
		e := *a
		e.Quantity = e.Quantity.Add(kernel.NewDecimal(1, 0))
		g.m.add(&e)
	case 2:
		g.m.add(a)
		g.m.add(a)
	case 3:
		mv := *a
		mv.AccountID = g.d.Accounts[g.r.intn(len(g.d.Accounts))].ID
		g.m.add(&mv)
	default:
		g.m.add(a)
	}
}

// randomCase builds desired activities and a Ghostfolio that differs from
// them in every way a sync meets: missing, edited, copied, moved and stray
// activities in owned accounts, missing owned accounts, and activities of
// the owner's own in an account the view doesn't own.
func randomCase(t *testing.T, r *rng) (ghostfolio.Desired, *ghostfolioModel) {
	t.Helper()
	g := generator{t: t, r: r, start: date(t, "2025-01-02"), m: &ghostfolioModel{accounts: []string{foreignAccount}}}
	for i := range 1 + r.intn(3) {
		id := ghostfolio.AccountUUID(instance, ledger.AccountID(fmt.Sprintf("a%d", i)))
		g.d.Accounts = append(g.d.Accounts, ghostfolio.Account{ID: id})
		if r.intn(4) > 0 {
			g.m.accounts = append(g.m.accounts, id)
		}
	}
	for i := range r.intn(12) {
		a := g.activity(fmt.Sprintf("%stest/%d", ghostfolio.CommentPrefix, i))
		g.d.Activities = append(g.d.Activities, a)
		g.actualFor(&a)
	}
	for i := range r.intn(3) {
		stray := g.activity(fmt.Sprintf("stray %d", i))
		g.m.add(&stray)
	}
	for i := range r.intn(3) {
		mine := g.activity(fmt.Sprintf("mine %d", i))
		mine.AccountID = foreignAccount
		g.m.add(&mine)
	}
	return g.d, g.m
}

// TestPlanIsIdempotent checks, over seeded random cases, that applying a
// plan the way Ghostfolio does and planning again changes nothing (D-64):
// the owned accounts then hold exactly the desired activities, and the
// owner's own activities are untouched.
func TestPlanIsIdempotent(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed}
	isForeign := func(id string) bool { return id == foreignAccount }
	ignoreID := cmpopts.IgnoreFields(ghostfolio.Activity{}, "ID")
	byComment := cmpopts.SortSlices(func(a, b ghostfolio.Activity) bool { return a.Comment < b.Comment })
	for c := range propertyCases {
		desired, m := randomCase(t, r)
		owned := map[string]bool{}
		for _, a := range desired.Accounts {
			owned[a.ID] = true
		}
		foreign := m.in(isForeign)
		actual := m.actual(owned)
		first := ghostfolio.MakePlan(&desired, &actual)
		m.apply(t, &first, owned)
		after := m.actual(owned)
		if again := ghostfolio.MakePlan(&desired, &after); !again.IsEmpty() {
			t.Fatalf("seed %d case %d: the plan after applying %+v is not empty: %+v", propertySeed, c, first, again)
		}
		if diff := cmp.Diff(desired.Activities, after.Activities, ignoreID, byComment, cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("seed %d case %d: owned activities (-desired +after):\n%s", propertySeed, c, diff)
		}
		if diff := cmp.Diff(foreign, m.in(isForeign), cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("seed %d case %d: the owner's own activities changed (-before +after):\n%s", propertySeed, c, diff)
		}
	}
}
