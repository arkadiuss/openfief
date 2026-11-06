package ui

import (
	"cmp"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The overview's limits (D-193): how many open items Needs you shows and
// how many entries Recent activity lists.
const (
	needsShown    = 3
	activityShown = 8
)

// heroView is the overview's value card (D-193): the portfolio's gross
// value large, with net and liquidation (D-77), the change since the end of
// last month and the time-weighted return this year (D-61).
type heroView struct {
	Date                    string
	Gross, Net, Liquidation *viewsv1.Money
	// MonthChange is the gross value's change since the last valued day
	// before this month, and MonthPercent the same in percent; nil and
	// empty when either day has no gross value.
	MonthChange  *viewsv1.Money
	MonthPercent string
	// YearReturn is the time-weighted return since the last day of last
	// year, or the first valued day after it, in percent; empty without
	// valued days.
	YearReturn string
}

// sleeveGroup is one sleeve's accounts on the overview (D-193): its name,
// the sum of its accounts' gross values and the accounts.
type sleeveGroup struct {
	ID, Name string
	Total    *viewsv1.Money
	Accounts []accountRow
}

// activityEntry is one line of Recent activity (D-193): its day, what
// happened and the accounts it links to.
type activityEntry struct {
	Day, Text string
	Links     []inboxLink
}

// allocationBar is one bar of the allocation (D-193): its group, value and
// share, its width out of barScale, and where it links, if anywhere.
type allocationBar struct {
	Key, Share, Width, Href string
	Value                   *viewsv1.Money
	Series                  int
}

// askBox is the overview's Ask box (D-194): the question asked, if any,
// its answer, and the suggestions drawn from the current state.
type askBox struct {
	Question    string
	Answer      *answerCard
	Suggestions []string
}

// statusLine is the overview's sentence of status (D-193): the day
// everything is valued through, how many items are open, and the
// connector instances whose sync keeps failing.
type statusLine struct {
	ValuedThrough string
	Open          int
	Failing       []string
}

// Sentence writes the status as one sentence.
func (st statusLine) Sentence() string {
	var b strings.Builder
	if st.ValuedThrough == "" {
		b.WriteString("Nothing is valued yet.")
	} else {
		b.WriteString("Everything is valued through " + st.ValuedThrough + ".")
	}
	switch st.Open {
	case 0:
		b.WriteString(" Nothing needs you.")
	default:
		b.WriteString(" " + plural(st.Open, "thing needs you.", "things need you."))
	}
	for _, inst := range st.Failing {
		b.WriteString(" The " + inst + " sync keeps failing.")
	}
	return b.String()
}

// sumMoney returns the sum of ms, every one in the same currency; nil for
// none, a sum that isn't a decimal, or mixed currencies.
func sumMoney(ms []*viewsv1.Money) *viewsv1.Money {
	if len(ms) == 0 {
		return nil
	}
	var sum kernel.Decimal
	for _, m := range ms {
		d, err := kernel.ParseDecimal(m.GetAmount())
		if err != nil || m.GetCurrency() != ms[0].GetCurrency() {
			return nil
		}
		sum = sum.Add(d)
	}
	return &viewsv1.Money{Amount: sum.String(), Currency: ms[0].GetCurrency()}
}

// moneyChange is a change of money and the same in percent, written in a
// number format; the zero change is none.
type moneyChange struct {
	Money   *viewsv1.Money
	Percent string
}

// diffMoney returns to - from, both in one currency, and the change in
// percent of from, signed, half-even to one place in f; none when either
// isn't a decimal, the currencies differ or from is zero.
func diffMoney(from, to *viewsv1.Money, f numberFormat) moneyChange {
	a, errA := kernel.ParseDecimal(from.GetAmount())
	b, errB := kernel.ParseDecimal(to.GetAmount())
	if errA != nil || errB != nil || from.GetCurrency() != to.GetCurrency() || a.IsZero() {
		return moneyChange{}
	}
	return moneyChange{
		Money:   &viewsv1.Money{Amount: b.Sub(a).String(), Currency: to.GetCurrency()},
		Percent: returnPercent(a.String(), b.String(), f),
	}
}

// monthChange returns the change of the gross value from the last of
// totals dated before date's month to the one on date, both in the
// reporting currency, and in percent; none without both.
func monthChange(totals []*viewsv1.DateTotal, date kernel.Date, f numberFormat) moneyChange {
	start := monthStart(date).String()
	var base, last *viewsv1.DateTotal
	for _, t := range totals {
		if t.GetDate() < start && t.GetValueReporting() != nil {
			base = t
		}
		if t.GetDate() == date.String() {
			last = t
		}
	}
	if base == nil || last == nil {
		return moneyChange{}
	}
	return diffMoney(base.GetValueReporting(), last.GetValueReporting(), f)
}

// yearReturn returns the time-weighted return of days, a performance
// series, since the last of them dated before date's year, or its first
// day, to its last, in percent; empty without days.
func yearReturn(days []*viewsv1.PerformanceDay, date kernel.Date, f numberFormat) string {
	if len(days) == 0 {
		return ""
	}
	start := yearStart(date.Year()).String()
	base := days[0]
	for _, d := range days {
		if d.GetDate() < start {
			base = d
		}
	}
	return returnPercent(base.GetIndex(), days[len(days)-1].GetIndex(), f)
}

// groupBySleeve returns rows grouped by their sleeve, in the order of
// sleeves, each with the sum of its accounts' gross values; accounts of a
// sleeve sleeves doesn't name come last, under their sleeve's ID.
func groupBySleeve(rows []accountRow, sleeves []*viewsv1.Sleeve) []sleeveGroup {
	var out []sleeveGroup
	index := map[string]int{}
	for _, sl := range sleeves {
		index[sl.GetId()] = len(out)
		out = append(out, sleeveGroup{ID: sl.GetId(), Name: sl.GetName()})
	}
	for _, row := range rows {
		id := row.Account.GetSleeve()
		i, ok := index[id]
		if !ok {
			i, index[id] = len(out), len(out)
			out = append(out, sleeveGroup{ID: id, Name: cmp.Or(id, "No sleeve")})
		}
		out[i].Accounts = append(out[i].Accounts, row)
	}
	out = slices.DeleteFunc(out, func(g sleeveGroup) bool { return len(g.Accounts) == 0 })
	for i := range out {
		var gross []*viewsv1.Money
		for _, row := range out[i].Accounts {
			if v := row.Total.GetValueReporting(); v != nil {
				gross = append(gross, v)
			}
		}
		out[i].Total = sumMoney(gross)
	}
	return out
}

// accountPill returns an account's status pill on the overview (D-193):
// its connector's sync keeps failing, then how many of its items are open,
// then whether it is valued, and through when.
func accountPill(row *accountRow, failing, open int) statusPill {
	switch {
	case failing > 0:
		return statusPill{"Sync failing", pillDanger}
	case open > 0:
		return statusPill{strconv.Itoa(open) + " to check", "pill-warn"}
	case row.Status.GetValuedThrough() == "":
		return statusPill{"Not valued yet", "pill-neutral"}
	}
	return statusPill{"Valued " + row.Status.GetValuedThrough(), pillOK}
}

// activityLine writes one entry of the recent activity (D-193), naming
// each account by its name in names.
func activityLine(a *viewsv1.Activity, names map[string]string, f numberFormat) activityEntry {
	day, _, _ := strings.Cut(a.GetAt(), "T")
	e := activityEntry{Day: day}
	inst := a.GetConnectorInstance()
	switch a.GetKind() {
	case viewsv1.ActivityKind_ACTIVITY_KIND_IMPORT, viewsv1.ActivityKind_ACTIVITY_KIND_SYNC:
		b := a.GetBatch()
		what := "Synced"
		if a.GetKind() == viewsv1.ActivityKind_ACTIVITY_KIND_IMPORT {
			what = "Imported"
		}
		e.Text = what + " · " + inst + " · " + plural(int(b.GetNewTransactions()), "new transaction", "new transactions")
		if b.GetUnsupported() > 0 {
			e.Text += ", " + strconv.Itoa(int(b.GetUnsupported())) + " left out"
		}
		e.Links = batchLinks(b, names)
	case viewsv1.ActivityKind_ACTIVITY_KIND_SYNC_FAILED:
		e.Text = "Sync failed · " + inst + " · " + enumLabel(a.GetAttempt().GetErrorClass().String(), "SYNC_ERROR_CLASS_")
	case viewsv1.ActivityKind_ACTIVITY_KIND_COUPON:
		c := a.GetChange()
		e.Text = "Coupon booked · " + instrumentLabel(c.GetInstrument()) + " · " + f.money(c.GetAmount())
		e.Links = accountLinks([]string{c.GetAccountId()}, names)
	case viewsv1.ActivityKind_ACTIVITY_KIND_CORRECTION:
		c := a.GetChange()
		e.Text = "Corrected · " + enumLabel(c.GetKind().String(), "TRANSACTION_KIND_") + " of " + c.GetDate()
		if c.GetDeleted() {
			e.Text = "Deleted · " + enumLabel(c.GetKind().String(), "TRANSACTION_KIND_") + " of " + c.GetDate()
		}
		e.Links = accountLinks([]string{c.GetAccountId()}, names)
	case viewsv1.ActivityKind_ACTIVITY_KIND_UNSPECIFIED:
		e.Text = "Something happened"
	}
	return e
}

// batchLinks links each account a batch booked into, of those the caller
// may view, to its page, named by names with its own new transactions
// when the read API counts them per account (D-199).
func batchLinks(b *viewsv1.ActivityBatch, names map[string]string) []inboxLink {
	if len(b.GetAccounts()) == 0 {
		return accountLinks(b.GetAccountIds(), names)
	}
	out := make([]inboxLink, 0, len(b.GetAccounts()))
	for _, c := range b.GetAccounts() {
		id := c.GetAccountId()
		out = append(out, inboxLink{
			Href: "/accounts/" + url.PathEscape(id), Text: cmp.Or(names[id], id) + ": " + strconv.Itoa(int(c.GetNewTransactions())) + " new",
		})
	}
	return out
}

// accountLinks links each of ids to its account's page, named by names.
func accountLinks(ids []string, names map[string]string) []inboxLink {
	out := make([]inboxLink, 0, len(ids))
	for _, id := range ids {
		if id != "" {
			out = append(out, inboxLink{Href: "/accounts/" + url.PathEscape(id), Text: cmp.Or(names[id], id)})
		}
	}
	return out
}

// allocationBars returns a's groups as bars (D-193), each a link to what
// is behind it (D-195): a sleeve's accounts on the overview, an
// institution's on the accounts page, an asset class's or a currency's
// holdings.
func allocationBars(a *allocationView) []allocationBar {
	out := make([]allocationBar, 0, len(a.Groups))
	for i, g := range a.Groups {
		bar := allocationBar{Key: g.Key, Value: g.Value, Share: g.Share, Width: strings.TrimSuffix(g.Share, "%"), Series: i%seriesCount + 1}
		switch a.Dimension.Key {
		case dimSleeve:
			bar.Href = "#sleeve-" + url.PathEscape(g.Key)
		case dimInstitution:
			bar.Href = accountsPath
		case dimAssetClass, dimCurrency:
			bar.Href = holdingsLink(a.Dimension.Key, g.Key, a.Date)
		}
		out = append(out, bar)
	}
	return out
}

// seriesCount is how many series colors the design system draws (D-30).
const seriesCount = 5

// overParam picks what the overview's chart card draws: the value
// (default) or the performance; the chart parameter picks the value's kind.
const (
	overParam       = "over"
	overPerformance = "performance"
	overAllocation  = "allocation"
)

// homeData reads what the overview shows besides its charts (D-193): the
// open items with the first few for Needs you, the syncs that keep
// failing, which accounts each instance is bound to, the sleeves and the
// recent activity, into data.
func (s *server) homeData(r *http.Request, sess session, data *overviewData) error {
	ctx := r.Context()
	items, err := s.read.ListAttentionItems(ctx, bearer(&viewsv1.ListAttentionItemsRequest{View: viewsv1.AttentionView_ATTENTION_VIEW_OPEN}, sess))
	if err != nil {
		return fmt.Errorf("ListAttentionItems: %w", err)
	}
	syncs, err := s.read.ListSyncStatuses(ctx, bearer(&viewsv1.ListSyncStatusesRequest{}, sess))
	if err != nil {
		return fmt.Errorf("ListSyncStatuses: %w", err)
	}
	bindings, err := s.read.ListConnectorBindings(ctx, bearer(&viewsv1.ListConnectorBindingsRequest{}, sess))
	if err != nil {
		return fmt.Errorf("ListConnectorBindings: %w", err)
	}
	sleeves, err := s.read.ListSleeves(ctx, bearer(&viewsv1.ListSleevesRequest{}, sess))
	if err != nil {
		return fmt.Errorf("ListSleeves: %w", err)
	}
	activity, err := s.read.ListActivity(ctx, bearer(&viewsv1.ListActivityRequest{PageSize: activityShown}, sess))
	if err != nil {
		return fmt.Errorf("ListActivity: %w", err)
	}
	data.Open = int(items.Msg.GetOpen())
	data.Needs = needsOf(items.Msg.GetItems())
	data.Failing = failingSyncs(syncs.Msg.GetStatuses())
	failingAccounts := failingPerAccount(data.Failing, bindings.Msg.GetBindings())
	openOf := openPerAccount(items.Msg.GetItems())
	names := map[string]string{}
	for i := range data.Accounts {
		row := &data.Accounts[i]
		names[row.Account.GetId()] = row.Account.GetName()
		row.Pill = accountPill(row, failingAccounts[row.Account.GetId()], openOf[row.Account.GetId()])
	}
	data.Sleeves = groupBySleeve(data.Accounts, sleeves.Msg.GetSleeves())
	f := requestFormat(r)
	for _, a := range activity.Msg.GetActivities() {
		data.Activity = append(data.Activity, activityLine(a, names, f))
	}
	data.Ask.Suggestions = askSuggestions(&askState{Open: data.Open, Failing: data.Failing, Accounts: accountNames(data.Accounts), Sleeves: sleeveNames(data.Sleeves)})
	return nil
}

// needsOf returns the first needsShown of items, the open ones, as Needs
// you shows them.
func needsOf(items []*viewsv1.AttentionItem) []inboxItem {
	out := make([]inboxItem, 0, needsShown)
	for i, it := range items {
		if i == needsShown {
			break
		}
		out = append(out, inboxItem{Item: it, Title: itemTitle(it), Pill: itemPill(it), Links: itemLinks(it)})
	}
	return out
}

// failingSyncs returns the connector instances among statuses whose sync
// keeps failing, in order.
func failingSyncs(statuses []*viewsv1.SyncStatus) []string {
	var out []string
	for _, st := range statuses {
		if st.GetKeepsFailing() {
			out = append(out, st.GetConnectorInstance())
		}
	}
	return out
}

// failingPerAccount counts the instances in failing each account is bound
// to, by bindings.
func failingPerAccount(failing []string, bindings []*viewsv1.ConnectorBinding) map[string]int {
	out := map[string]int{}
	for _, b := range bindings {
		if slices.Contains(failing, b.GetConnectorInstance()) {
			out[b.GetAccountId()]++
		}
	}
	return out
}

// openPerAccount counts items, the open ones, per account they are tied
// to.
func openPerAccount(items []*viewsv1.AttentionItem) map[string]int {
	out := map[string]int{}
	for _, it := range items {
		for _, a := range it.GetAccountIds() {
			out[a]++
		}
	}
	return out
}

// accountNames returns the names of rows' accounts, in order.
func accountNames(rows []accountRow) []string {
	out := make([]string, 0, len(rows))
	for i := range rows {
		out = append(out, rows[i].Account.GetName())
	}
	return out
}

// sleeveNames returns the names of groups, in order.
func sleeveNames(groups []sleeveGroup) []string {
	out := make([]string, 0, len(groups))
	for i := range groups {
		out = append(out, groups[i].Name)
	}
	return out
}
