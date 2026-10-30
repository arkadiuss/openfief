package ghostfolio_test

import (
	"cmp"
	"fmt"
	"slices"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
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

// bondSeries returns the series codes of the random cases' bonds.
func bondSeries() []string { return []string{"EDO0136", "COI0128"} }

// legacySymbol is the symbol an earlier version gave the bond series with
// code series, before symbols were keyed by instance and user.
func legacySymbol(series string) string { return "legacy-" + series }

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
// makes: the user's accounts, activities, MANUAL assets (symbol to name)
// and cash balances, with IDs from a counter.
type ghostfolioModel struct {
	accounts   []string
	activities []ghostfolio.Activity
	assets     map[string]string
	balances   []ghostfolio.Balance
	next       int
}

// newModel is a Ghostfolio whose user has only the account foreignAccount.
func newModel() ghostfolioModel {
	return ghostfolioModel{accounts: []string{foreignAccount}, assets: map[string]string{}}
}

// duplicate is Ghostfolio's import rule (R-15): an activity is skipped when
// one the user has agrees in comment, currency, data source, date, fee,
// quantity, symbol, type and unit price, in any account.
func duplicate(a, b *ghostfolio.Activity) bool {
	return a.Comment == b.Comment && a.Currency.Equal(b.Currency) && a.DataSource == b.DataSource && a.Date == b.Date &&
		a.Fee.Equal(b.Fee) && a.Quantity.Equal(b.Quantity) && a.Symbol == b.Symbol && a.Type == b.Type && a.UnitPrice.Equal(b.UnitPrice)
}

// add stores a with the next ID and, unless it exists, the MANUAL asset it
// uses, named name.
func (m *ghostfolioModel) add(a *ghostfolio.Activity, name string) {
	m.next++
	stored := *a
	stored.ID = fmt.Sprintf("g-%03d", m.next)
	m.activities = append(m.activities, stored)
	if m.assets == nil {
		m.assets = map[string]string{}
	}
	if _, ok := m.assets[a.Symbol]; a.DataSource == ghostfolio.DataSourceManual && !ok {
		m.assets[a.Symbol] = name
	}
}

// addBalance stores b with the next ID.
func (m *ghostfolioModel) addBalance(b ghostfolio.Balance) {
	m.next++
	b.ID = fmt.Sprintf("b-%03d", m.next)
	m.balances = append(m.balances, b)
}

// importOne imports a as Ghostfolio does, with names giving the names of
// the assets sent along. It fails t when a would go onto a new MANUAL asset
// while the user has another asset of its name: Ghostfolio's import would
// reuse that one, and the dry run would refuse the import.
func (m *ghostfolioModel) importOne(t *testing.T, a *ghostfolio.Activity, names map[string]string) {
	t.Helper()
	if _, ok := m.assets[a.Symbol]; a.DataSource == ghostfolio.DataSourceManual && !ok {
		for sym, name := range m.assets {
			if name == names[a.Symbol] {
				t.Fatalf("the import of %q would reuse the asset %s of the same name", a.Comment, sym)
			}
		}
	}
	if !slices.ContainsFunc(m.activities, func(b ghostfolio.Activity) bool { return duplicate(a, &b) }) {
		m.add(a, names[a.Symbol])
	}
}

// apply carries out p as Syncer does, in its order, and fails t if the plan
// touches an account that isn't owned.
func (m *ghostfolioModel) apply(t *testing.T, p *ghostfolio.Plan, owned map[string]bool) {
	t.Helper()
	for _, a := range p.CreateAccounts {
		m.accounts = append(m.accounts, a.ID)
	}
	m.remove(t, p.PreDelete, owned)
	names := map[string]string{}
	for _, as := range p.Assets {
		names[as.Symbol] = as.Name
	}
	for i := range p.Import {
		m.importOne(t, &p.Import[i], names)
	}
	m.remove(t, p.Delete, owned)
	for _, id := range p.Clear {
		if !owned[id] {
			t.Fatalf("bulk delete of account %s, which is not owned", id)
		}
		m.activities = slices.DeleteFunc(m.activities, func(b ghostfolio.Activity) bool { return b.AccountID == id })
		m.dropUnused()
	}
	m.applyBalances(t, p, owned)
}

// applyBalances sets and deletes the balances p names: a set replaces the
// balance of its account and date.
func (m *ghostfolioModel) applyBalances(t *testing.T, p *ghostfolio.Plan, owned map[string]bool) {
	t.Helper()
	for _, b := range p.SetBalances {
		if !owned[b.AccountID] {
			t.Fatalf("balance set in account %s, which is not owned", b.AccountID)
		}
		m.balances = slices.DeleteFunc(m.balances, func(o ghostfolio.Balance) bool { return o.AccountID == b.AccountID && o.Date == b.Date })
		m.addBalance(b)
	}
	for _, b := range p.DeleteBalances {
		if !owned[b.AccountID] {
			t.Fatalf("delete of balance %s in an account not owned", b.ID)
		}
		m.balances = slices.DeleteFunc(m.balances, func(o ghostfolio.Balance) bool { return o.ID == b.ID })
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
		m.dropUnused()
	}
}

// dropUnused deletes the MANUAL assets no activity uses, as Ghostfolio
// deletes an asset with its last activity (R-15).
func (m *ghostfolioModel) dropUnused() {
	for sym := range m.assets {
		if !slices.ContainsFunc(m.activities, func(a ghostfolio.Activity) bool { return a.Symbol == sym }) {
			delete(m.assets, sym)
		}
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

// balancesIn returns m's balances in the accounts keep accepts.
func (m *ghostfolioModel) balancesIn(keep func(account string) bool) []ghostfolio.Balance {
	var out []ghostfolio.Balance
	for _, b := range m.balances {
		if keep(b.AccountID) {
			out = append(out, b)
		}
	}
	return out
}

// actual is what Syncer reads from m.
func (m *ghostfolioModel) actual(owned map[string]bool) ghostfolio.Actual {
	isOwned := func(id string) bool { return owned[id] }
	return ghostfolio.Actual{Accounts: slices.Clone(m.accounts), Activities: m.in(isOwned), Balances: m.balancesIn(isOwned)}
}

// generator draws one random case.
type generator struct {
	t     *testing.T
	r     *rng
	start kernel.Date
	d     ghostfolio.Desired
	m     *ghostfolioModel
	// names are the bond assets' names by symbol.
	names map[string]string
}

// activity is a random activity with comment c in a random owned account:
// a listed instrument's, or, when bonds is set, possibly a bond's.
func (g *generator) activity(c string, bonds bool) ghostfolio.Activity {
	r := g.r
	a := ghostfolio.Activity{
		AccountID: g.d.Accounts[r.intn(len(g.d.Accounts))].ID, Comment: c,
		Type: []ghostfolio.ActivityType{ghostfolio.TypeBuy, ghostfolio.TypeSell}[r.intn(2)],
		Date: g.start.AddDays(r.intn(3)), Currency: cur(g.t, []string{"PLN", "USD"}[r.intn(2)]),
		DataSource: yahoo, Symbol: []string{"ETF1.L", "ETF2.L"}[r.intn(2)],
		Quantity: kernel.NewDecimal(int64(1+r.intn(3)), 0), UnitPrice: kernel.NewDecimal(int64(100+r.intn(3)), -1),
	}
	if bonds && r.intn(3) == 0 {
		all := bondSeries()
		series := all[r.intn(len(all))]
		a.DataSource, a.Symbol, a.Currency = ghostfolio.DataSourceManual, bondSym(series), cur(g.t, "PLN")
		g.names[a.Symbol] = series
	}
	return a
}

// actualFor stores in the model what Ghostfolio holds for desired a: in
// turn nothing, an edited copy, two copies, a copy in another owned
// account, a copy on the asset an earlier version gave a bond, or a itself.
func (g *generator) actualFor(a *ghostfolio.Activity) {
	const ways = 7
	name := g.names[a.Symbol]
	switch g.r.intn(ways) {
	case 0:
	case 1:
		e := *a
		e.Quantity = e.Quantity.Add(kernel.NewDecimal(1, 0))
		g.m.add(&e, name)
	case 2:
		g.m.add(a, name)
		g.m.add(a, name)
	case 3:
		mv := *a
		mv.AccountID = g.d.Accounts[g.r.intn(len(g.d.Accounts))].ID
		g.m.add(&mv, name)
	case 4:
		old := *a
		if old.DataSource == ghostfolio.DataSourceManual {
			old.Symbol = legacySymbol(name)
		}
		g.m.add(&old, name)
	default:
		g.m.add(a, name)
	}
}

// balances draws the owned accounts' desired balances, and what Ghostfolio
// holds for them: nothing, another value or the same, plus a balance on a
// date that should have none, and one in the owner's own account.
func (g *generator) balances() {
	r := g.r
	for _, acc := range g.d.Accounts {
		exists := slices.Contains(g.m.accounts, acc.ID)
		for k := range r.intn(3) {
			b := ghostfolio.Balance{AccountID: acc.ID, Date: g.start.AddDays(2 * k), Value: kernel.NewDecimal(int64(10*r.intn(5)), 0)}
			g.d.Balances = append(g.d.Balances, b)
			switch {
			case !exists || r.intn(3) == 0:
			case r.intn(2) == 0:
				b.Value = b.Value.Add(kernel.NewDecimal(1, 0))
				g.m.addBalance(b)
			default:
				g.m.addBalance(b)
			}
		}
		if exists && r.intn(3) == 0 {
			g.m.addBalance(ghostfolio.Balance{AccountID: acc.ID, Date: g.start.AddDays(5), Value: kernel.NewDecimal(3, 0)})
		}
	}
	g.m.addBalance(ghostfolio.Balance{AccountID: foreignAccount, Date: g.start, Value: kernel.NewDecimal(7, 0)})
}

// assets returns the desired MANUAL assets: those the desired activities
// use, in symbol order.
func (g *generator) assets() []ghostfolio.AssetProfile {
	var out []ghostfolio.AssetProfile
	for i := range g.d.Activities {
		sym := g.d.Activities[i].Symbol
		if g.d.Activities[i].DataSource == ghostfolio.DataSourceManual && !slices.ContainsFunc(out, func(a ghostfolio.AssetProfile) bool { return a.Symbol == sym }) {
			out = append(out, ghostfolio.AssetProfile{Symbol: sym, Name: g.names[sym], Currency: cur(g.t, "PLN")})
		}
	}
	slices.SortFunc(out, func(a, b ghostfolio.AssetProfile) int { return cmp.Compare(a.Symbol, b.Symbol) })
	return out
}

// randomCase builds desired activities and balances and a Ghostfolio that
// differs from them in every way a sync meets: missing, edited, copied,
// moved and stray activities in owned accounts, activities on the assets
// an earlier version gave bonds, missing, different and stray balances,
// missing owned accounts, and activities and balances of the owner's own
// in an account the view doesn't own.
func randomCase(t *testing.T, r *rng) (ghostfolio.Desired, *ghostfolioModel) {
	t.Helper()
	m := newModel()
	g := generator{t: t, r: r, start: date(t, "2025-01-02"), m: &m, names: map[string]string{}}
	for i := range 1 + r.intn(3) {
		id := ghostfolio.AccountUUID(instance, ledger.AccountID(fmt.Sprintf("a%d", i)))
		g.d.Accounts = append(g.d.Accounts, ghostfolio.Account{ID: id})
		if r.intn(4) > 0 {
			g.m.accounts = append(g.m.accounts, id)
		}
	}
	for i := range r.intn(12) {
		a := g.activity(fmt.Sprintf("%stest/%d", ghostfolio.CommentPrefix, i), true)
		g.d.Activities = append(g.d.Activities, a)
		g.actualFor(&a)
	}
	for i := range r.intn(3) {
		stray := g.activity(fmt.Sprintf("stray %d", i), true)
		g.m.add(&stray, g.names[stray.Symbol])
	}
	for i := range r.intn(3) {
		mine := g.activity(fmt.Sprintf("mine %d", i), false)
		mine.AccountID = foreignAccount
		g.m.add(&mine, "")
	}
	g.balances()
	g.d.Assets = g.assets()
	return g.d, g.m
}

// TestPlanIsIdempotent checks, over seeded random cases, that applying a
// plan the way Ghostfolio does and planning again changes nothing (D-64,
// D-108): the owned accounts then hold exactly the desired activities and
// balances, every bond has its asset, no import trips over an asset an
// earlier version left, and the owner's own activities and balances are
// untouched.
func TestPlanIsIdempotent(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed}
	isForeign := func(id string) bool { return id == foreignAccount }
	for c := range propertyCases {
		desired, m := randomCase(t, r)
		owned := map[string]bool{}
		for _, a := range desired.Accounts {
			owned[a.ID] = true
		}
		foreign, foreignBalances := m.in(isForeign), m.balancesIn(isForeign)
		actual := m.actual(owned)
		first := ghostfolio.MakePlan(&desired, &actual)
		m.apply(t, &first, owned)
		after := m.actual(owned)
		if again := ghostfolio.MakePlan(&desired, &after); !again.IsEmpty() {
			t.Fatalf("seed %d case %d: the plan after applying %+v is not empty: %+v", propertySeed, c, first, again)
		}
		if msg := converged(&desired, &after, m); msg != "" {
			t.Fatalf("seed %d case %d: %s", propertySeed, c, msg)
		}
		if diff := gocmp.Diff(foreign, m.in(isForeign), cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("seed %d case %d: the owner's own activities changed (-before +after):\n%s", propertySeed, c, diff)
		}
		if diff := gocmp.Diff(foreignBalances, m.balancesIn(isForeign), cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("seed %d case %d: the owner's own balances changed (-before +after):\n%s", propertySeed, c, diff)
		}
	}
}

// converged returns what differs between desired and what the owned
// accounts hold after a sync, with the assets of m, or "" when nothing does.
func converged(desired *ghostfolio.Desired, after *ghostfolio.Actual, m *ghostfolioModel) string {
	byComment := cmpopts.SortSlices(func(a, b ghostfolio.Activity) bool { return a.Comment < b.Comment })
	byDate := cmpopts.SortSlices(func(a, b ghostfolio.Balance) bool {
		return cmp.Or(cmp.Compare(a.AccountID, b.AccountID), a.Date.Compare(b.Date)) < 0
	})
	ignoreID := cmpopts.IgnoreFields(ghostfolio.Activity{}, "ID")
	if diff := gocmp.Diff(desired.Activities, after.Activities, ignoreID, byComment, cmpopts.EquateEmpty()); diff != "" {
		return "owned activities (-desired +after):\n" + diff
	}
	ignoreBalanceID := cmpopts.IgnoreFields(ghostfolio.Balance{}, "ID")
	if diff := gocmp.Diff(desired.Balances, after.Balances, ignoreBalanceID, byDate, cmpopts.EquateEmpty()); diff != "" {
		return "owned balances (-desired +after):\n" + diff
	}
	for _, as := range desired.Assets {
		if name, ok := m.assets[as.Symbol]; !ok || name != as.Name {
			return fmt.Sprintf("asset %s is %q, %v; want %q", as.Symbol, name, ok, as.Name)
		}
	}
	return ""
}
