package ui

import (
	"cmp"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The Ask box's limits: how many rows an answer's table shows and how
// many sync attempts an answer about a sync reads.
const (
	answerRows     = 6
	answerAttempts = 5
	// twrPlaces is how many decimal places a return in percent shows.
	twrPlaces = 1
)

// answerCard is the Ask box's answer to one question (D-194): a title, a
// sentence, a small table and one link to the place it summarizes, or, for
// a question the box doesn't understand, the commands it does. It is the
// seam an agent fills in M2 (P-57): whatever answers a question, the page
// shows its card, and nothing in the card changes anything.
type answerCard struct {
	Question string
	Title    string
	Text     string
	Rows     []answerRow
	// Link is the one "Open …" link to the place the answer summarizes;
	// nil when there is none.
	Link *answerLink
	// Help are the commands the box understands, for a question it
	// doesn't; empty otherwise.
	Help []askHelp
}

// answerRow is one row of an answer's table: what it is about, a detail
// and a figure, and where it links, if anywhere.
type answerRow struct {
	Label, Detail, Figure, Href string
}

// answerLink is an answer's link to the page it summarizes.
type answerLink struct {
	Href, Text string
}

// answerer answers a question asked in the Ask box for a session. M0's is
// commandAnswerer, which knows a fixed set of commands and reads the read
// API alone; an M2 agent can answer through the same cards (P-57). An
// answerer never changes anything, and never logs or sends the question
// anywhere (GR-12).
type answerer interface {
	answer(r *http.Request, sess session, question string) (answerCard, error)
}

// commandAnswerer answers the Ask box's commands (parseAsk) from the read
// API, with the session's token, as every page reads it.
type commandAnswerer struct {
	s *server
}

// answer parses question and answers its command; a question none of the
// commands matches gets the commands the box understands.
func (a commandAnswerer) answer(r *http.Request, sess session, question string) (answerCard, error) {
	q, ok := parseAsk(question, a.s.today())
	if !ok {
		return unknownAnswer(question), nil
	}
	date, err := a.s.latestDate(r, sess)
	if err != nil {
		return answerCard{}, err
	}
	var card answerCard
	switch q.Command {
	case askNeeds:
		card, err = a.needs(r, sess)
	case askShow:
		card, err = a.show(r, sess, &q, date)
	case askSync:
		card, err = a.sync(r, sess, q.Term)
	case askAllocation:
		card, err = a.allocation(r, sess, parseDimension(q.Dimension), date)
	case askPerformance:
		card, err = a.performance(r, sess, q.Period, date)
	case askValue:
		card, err = a.value(r, sess, q.Date, date)
	case askUnknown:
		return unknownAnswer(question), nil
	}
	card.Question = question
	return card, err
}

// unknownAnswer answers a question the box doesn't understand with the
// commands it does.
func unknownAnswer(question string) answerCard {
	return answerCard{
		Question: question, Title: "I can't answer that yet",
		Text: "The Ask box understands a few commands and answers them from your own record; it holds no AI model. Try one of these.",
		Help: knownCommands(),
	}
}

// latestDate returns the latest day any account is valued through, or
// today when none is.
func (s *server) latestDate(r *http.Request, sess session) (kernel.Date, error) {
	statuses, err := s.read.ListValuationStatuses(r.Context(), bearer(&viewsv1.ListValuationStatusesRequest{}, sess))
	if err != nil {
		return kernel.Date{}, fmt.Errorf("ListValuationStatuses: %w", err)
	}
	return asOfDate("", statuses.Msg.GetStatuses(), s.today())
}

// needs answers what needs the owner: the open attention items, the
// Inbox's first ones.
func (a commandAnswerer) needs(r *http.Request, sess session) (answerCard, error) {
	resp, err := a.s.read.ListAttentionItems(r.Context(), bearer(&viewsv1.ListAttentionItemsRequest{View: viewsv1.AttentionView_ATTENTION_VIEW_OPEN}, sess))
	if err != nil {
		return answerCard{}, fmt.Errorf("ListAttentionItems: %w", err)
	}
	n := int(resp.Msg.GetOpen())
	card := answerCard{
		Title: plural(n, "thing needs you", "things need you"),
		Text:  "The Inbox gathers every check that found something for you, grouped by what you do about it.",
		Link:  &answerLink{Href: inboxPath, Text: "Open the Inbox (" + strconv.Itoa(n) + ")"},
	}
	if n == 0 {
		card.Title, card.Text = "Nothing needs you", "No check has found anything new for you to look at."
	}
	for i, it := range resp.Msg.GetItems() {
		if i == answerRows {
			break
		}
		card.Rows = append(card.Rows, answerRow{
			Label: itemTitle(it), Detail: groupLabel(it.GetGroup()), Figure: itemPill(it).Word, Href: inboxPath + "#item-" + strconv.FormatInt(it.GetId(), 10),
		})
	}
	return card, nil
}

// groupLabel names an attention item's group as the Inbox does.
func groupLabel(g viewsv1.AttentionItemGroup) string {
	return newInboxGroups()[groupIndex(g)].Title
}

// plural writes n with one or many, "1 thing needs you" or "3 things
// need you".
func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}

// show answers a show command: the account, sleeve or instrument the
// question names among the session's, on date.
func (a commandAnswerer) show(r *http.Request, sess session, q *askQuery, date kernel.Date) (answerCard, error) {
	c, err := a.s.askTargets(r, sess, date)
	if err != nil {
		return answerCard{}, err
	}
	targets, accounts, sleeves := c.Targets, c.Accounts, c.Sleeves
	t, ok := matchTarget(q.Term, q.Show, targets)
	if !ok {
		return answerCard{
			Title: "Nothing here is called “" + q.Term + "”",
			Text:  "Show finds an account, a sleeve or an instrument by its name, ticker or ISIN.",
			Rows:  targetRows(targets),
		}, nil
	}
	switch t.Kind {
	case showAccount:
		return a.s.accountAnswer(r, sess, accounts[t.ID], sleeves, date)
	case showSleeve:
		return a.s.sleeveAnswer(r, sess, t, accounts, date)
	case showInstrument, showAny:
	}
	return a.s.instrumentAnswer(r, sess, t, accounts, date)
}

// targetRows lists the first targets a show command can name, as rows.
func targetRows(targets []askTarget) []answerRow {
	var out []answerRow
	for i, t := range targets {
		if i == answerRows {
			break
		}
		out = append(out, answerRow{Label: t.Label, Detail: [...]string{"", "account", dimSleeve, kindInstrument}[t.Kind]})
	}
	return out
}

// askCatalog is what a show command can name: its targets, the accounts
// by ID and the sleeves' names by ID.
type askCatalog struct {
	Targets  []askTarget
	Accounts map[string]*viewsv1.Account
	Sleeves  map[string]string
}

// askTargets returns what a show command can name, on date: every
// account, sleeve and instrument the session's accounts hold.
func (s *server) askTargets(r *http.Request, sess session, date kernel.Date) (askCatalog, error) {
	ctx := r.Context()
	accounts, err := s.read.ListAccounts(ctx, bearer(&viewsv1.ListAccountsRequest{}, sess))
	if err != nil {
		return askCatalog{}, fmt.Errorf("ListAccounts: %w", err)
	}
	sleeves, err := s.read.ListSleeves(ctx, bearer(&viewsv1.ListSleevesRequest{}, sess))
	if err != nil {
		return askCatalog{}, fmt.Errorf("ListSleeves: %w", err)
	}
	byID, names := map[string]*viewsv1.Account{}, map[string]string{}
	var targets []askTarget
	for _, a := range accounts.Msg.GetAccounts() {
		byID[a.GetId()] = a
		targets = append(targets, askTarget{Kind: showAccount, ID: a.GetId(), Label: a.GetName(), Words: []string{a.GetName(), a.GetId(), a.GetInstitution()}})
	}
	for _, sl := range sleeves.Msg.GetSleeves() {
		names[sl.GetId()] = sl.GetName()
		targets = append(targets, askTarget{Kind: showSleeve, ID: sl.GetId(), Label: sl.GetName(), Words: []string{sl.GetName(), sl.GetId(), sl.GetKind()}})
	}
	seen := map[string]bool{}
	for _, a := range accounts.Msg.GetAccounts() {
		holdings, hErr := s.read.ListHoldings(ctx, bearer(&viewsv1.ListHoldingsRequest{AccountId: a.GetId(), Date: date.String()}, sess))
		if hErr != nil {
			return askCatalog{}, fmt.Errorf("ListHoldings: %w", hErr)
		}
		for _, h := range holdings.Msg.GetHoldings() {
			in := h.GetInstrument()
			if seen[in.GetIsin()] {
				continue
			}
			seen[in.GetIsin()] = true
			targets = append(targets, askTarget{Kind: showInstrument, ID: in.GetIsin(), Label: instrumentLabel(in), Words: []string{in.GetName(), in.GetSymbol(), in.GetIsin()}})
		}
	}
	return askCatalog{Targets: targets, Accounts: byID, Sleeves: names}, nil
}

// accountAnswer answers a show command naming account a: its three values
// and its largest holdings on date.
func (s *server) accountAnswer(r *http.Request, sess session, a *viewsv1.Account, sleeves map[string]string, date kernel.Date) (answerCard, error) {
	f := requestFormat(r)
	vals, err := s.read.ListAccountValuations(r.Context(), bearer(&viewsv1.ListAccountValuationsRequest{AccountId: a.GetId(), From: date.String(), To: date.String()}, sess))
	if err != nil {
		return answerCard{}, fmt.Errorf("ListAccountValuations: %w", err)
	}
	card := answerCard{
		Title: a.GetName() + ": not valued on " + date.String(),
		Text:  strings.Join(nonEmpty(a.GetInstitution(), sleeves[a.GetSleeve()], wrapperLabel(a.GetTaxWrapper())), " · ") + ". Holdings at their stored valuations, in the reporting currency.",
		Link:  &answerLink{Href: "/accounts/" + url.PathEscape(a.GetId()), Text: "Open " + a.GetName()},
	}
	if totals := vals.Msg.GetTotals(); len(totals) > 0 {
		card.Title = a.GetName() + ": " + f.money(totals[len(totals)-1].GetValueReporting()) + " gross"
	}
	for i, v := range vals.Msg.GetValuations() {
		if i == answerRows {
			break
		}
		card.Rows = append(card.Rows, answerRow{Label: instrumentLabel(v.GetInstrument()), Detail: enumLabel(v.GetBasis().String(), "VALUATION_BASIS_"), Figure: f.money(v.GetValueReporting())})
	}
	return card, nil
}

// nonEmpty returns those of parts that aren't empty, in order.
func nonEmpty(parts ...string) []string {
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// sleeveAnswer answers a show command naming sleeve t: its accounts with
// each one's gross value on date.
func (s *server) sleeveAnswer(r *http.Request, sess session, t askTarget, accounts map[string]*viewsv1.Account, date kernel.Date) (answerCard, error) {
	f := requestFormat(r)
	card := answerCard{
		Title: t.Label + ": no accounts", Text: "Each account in the sleeve, gross with its cash, in the reporting currency.",
		Link: &answerLink{Href: "/#sleeve-" + url.PathEscape(t.ID), Text: "Open " + t.Label + " on the overview"},
	}
	var sum []*viewsv1.Money
	for _, a := range sortedAccounts(accounts) {
		if a.GetSleeve() != t.ID {
			continue
		}
		vals, err := s.read.ListAccountValuations(r.Context(), bearer(&viewsv1.ListAccountValuationsRequest{AccountId: a.GetId(), From: date.String(), To: date.String()}, sess))
		if err != nil {
			return answerCard{}, fmt.Errorf("ListAccountValuations: %w", err)
		}
		row := answerRow{Label: a.GetName(), Detail: a.GetInstitution(), Figure: "not valued", Href: "/accounts/" + url.PathEscape(a.GetId())}
		if totals := vals.Msg.GetTotals(); len(totals) > 0 {
			gross := totals[len(totals)-1].GetValueReporting()
			row.Figure, sum = f.money(gross), append(sum, gross)
		}
		card.Rows = append(card.Rows, row)
	}
	if len(card.Rows) > 0 {
		card.Title = t.Label + ": " + plural(len(card.Rows), "account", "accounts")
		if total := sumMoney(sum); total != nil {
			card.Title += ", " + f.money(total) + " gross"
		}
	}
	return card, nil
}

// sortedAccounts returns accounts in name then ID order.
func sortedAccounts(accounts map[string]*viewsv1.Account) []*viewsv1.Account {
	out := make([]*viewsv1.Account, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, a)
	}
	slices.SortFunc(out, func(a, b *viewsv1.Account) int {
		return cmp.Or(cmp.Compare(a.GetName(), b.GetName()), cmp.Compare(a.GetId(), b.GetId()))
	})
	return out
}

// instrumentAnswer answers a show command naming instrument t: what the
// read API knows of it and who holds it on date.
func (s *server) instrumentAnswer(r *http.Request, sess session, t askTarget, accounts map[string]*viewsv1.Account, date kernel.Date) (answerCard, error) {
	ctx := r.Context()
	in, err := s.read.GetInstrument(ctx, bearer(&viewsv1.GetInstrumentRequest{Isin: t.ID}, sess))
	if err != nil {
		return answerCard{}, fmt.Errorf("GetInstrument: %w", err)
	}
	held, err := s.read.ListInstrumentHoldings(ctx, bearer(&viewsv1.ListInstrumentHoldingsRequest{Isin: t.ID, Date: date.String()}, sess))
	if err != nil {
		return answerCard{}, fmt.Errorf("ListInstrumentHoldings: %w", err)
	}
	f := requestFormat(r)
	d := in.Msg.GetInstrument()
	card := answerCard{
		Title: strings.Join(nonEmpty(d.GetName(), t.ID), " · "),
		Text:  "ISIN " + d.GetIsin() + ", " + strings.ReplaceAll(d.GetAssetClass(), "_", " ") + ".",
	}
	for _, l := range d.GetListings() {
		if p := l.GetLatestPrice(); p != nil {
			card.Text += " " + l.GetInstrument().GetSymbol() + " last priced " + f.money(p.GetClose()) + " on " + p.GetDate() + " from " + l.GetPriceSource() + "."
		}
	}
	for _, h := range held.Msg.GetHoldings() {
		name := h.GetAccountId()
		if a, ok := accounts[name]; ok {
			name = a.GetName()
		}
		card.Rows = append(card.Rows, answerRow{Label: name, Detail: h.GetInstrument().GetSymbol(), Figure: f.amount(h.GetQuantity()) + " units", Href: "/accounts/" + url.PathEscape(h.GetAccountId())})
	}
	card.Link = &answerLink{Href: instrumentPath(d.GetIsin()), Text: "Open " + card.Title}
	return card, nil
}

// sync answers why a sync fails: the instance term names, its status and
// its latest attempts, each with its error class, never its text (D-198).
func (a commandAnswerer) sync(r *http.Request, sess session, term string) (answerCard, error) {
	statuses, err := a.s.read.ListSyncStatuses(r.Context(), bearer(&viewsv1.ListSyncStatusesRequest{}, sess))
	if err != nil {
		return answerCard{}, fmt.Errorf("ListSyncStatuses: %w", err)
	}
	var instances []string
	for _, st := range statuses.Msg.GetStatuses() {
		instances = append(instances, st.GetConnectorInstance())
	}
	inst, ok := matchInstance(term, instances)
	if !ok {
		card := answerCard{Title: "No connector called “" + term + "” has synced", Text: "These have:", Link: &answerLink{Href: pluginsPath, Text: openPlugins}}
		for _, st := range statuses.Msg.GetStatuses() {
			card.Rows = append(card.Rows, syncStatusRow(st))
		}
		return card, nil
	}
	resp, err := a.s.read.ListSyncAttempts(r.Context(), bearer(&viewsv1.ListSyncAttemptsRequest{ConnectorInstance: inst, PageSize: answerAttempts}, sess))
	if err != nil && connect.CodeOf(err) != connect.CodeNotFound {
		return answerCard{}, fmt.Errorf("ListSyncAttempts: %w", err)
	}
	return syncAnswer(inst, resp.Msg), nil
}

// syncAnswer turns an instance's status and latest attempts into its
// answer.
func syncAnswer(inst string, resp *viewsv1.ListSyncAttemptsResponse) answerCard {
	st := resp.GetStatus()
	card := answerCard{
		Title: "The " + inst + " sync is working",
		Text:  "Last good sync: " + orNone(st.GetLastSuccessAt()) + ". Next try: " + orNone(st.GetNextSyncAt()) + ". Every attempt keeps its error class, never its text.",
		Link:  &answerLink{Href: pluginsPath, Text: openPlugins},
	}
	switch {
	case st.GetKeepsFailing():
		card.Title = "The " + inst + " sync keeps failing since " + st.GetFailingSince()
		card.Link = &answerLink{Href: inboxPath, Text: "Open the failing sync in the Inbox"}
	case st.GetFailures() > 0:
		card.Title = "The " + inst + " sync failed " + plural(int(st.GetFailures()), "time", "times") + " in a row"
	}
	for _, at := range resp.GetAttempts() {
		outcome := "went through"
		if at.GetFailed() {
			outcome = enumLabel(at.GetErrorClass().String(), "SYNC_ERROR_CLASS_")
		}
		card.Rows = append(card.Rows, answerRow{Label: at.GetStartedAt(), Detail: at.GetMode() + " sync", Figure: outcome})
	}
	return card
}

// syncStatusRow is one instance's status as a row.
func syncStatusRow(st *viewsv1.SyncStatus) answerRow {
	state := "working"
	switch {
	case st.GetKeepsFailing():
		state = "keeps failing"
	case st.GetFailures() > 0:
		state = "failed " + strconv.Itoa(int(st.GetFailures())) + "×"
	}
	return answerRow{Label: st.GetConnectorInstance(), Detail: "last attempt " + orNone(st.GetLastAttemptAt()), Figure: state}
}

// orNone is s, or "none yet" when it is empty.
func orNone(s string) string {
	if s == "" {
		return "none yet"
	}
	return s
}

// allocation answers the allocation by d, gross, on date.
func (a commandAnswerer) allocation(r *http.Request, sess session, d dimension, date kernel.Date) (answerCard, error) {
	resp, err := a.s.read.GetAllocation(r.Context(), bearer(&viewsv1.GetAllocationRequest{Date: date.String(), By: d.Enum, Value: viewsv1.ValueKind_VALUE_KIND_GROSS}, sess))
	if err != nil {
		return answerCard{}, fmt.Errorf("GetAllocation: %w", err)
	}
	f := requestFormat(r)
	total := resp.Msg.GetTotalReporting()
	card := answerCard{
		Title: "By " + strings.ToLower(d.Label) + ", gross, on " + date.String(),
		Text:  "Holdings at their stored valuations plus cash, in the reporting currency: " + f.money(total) + " in all.",
		Link:  &answerLink{Href: "/?" + url.Values{"by": {d.Key}, overParam: {overAllocation}}.Encode() + "#over-time", Text: "Open the allocation"},
	}
	for i, g := range resp.Msg.GetGroups() {
		if i == answerRows {
			break
		}
		card.Rows = append(card.Rows, answerRow{Label: g.GetKey(), Detail: share(g.GetValueReporting().GetAmount(), total.GetAmount()), Figure: f.money(g.GetValueReporting())})
	}
	return card, nil
}

// performance answers the time-weighted return over p, up to date.
func (a commandAnswerer) performance(r *http.Request, sess session, p askPeriod, date kernel.Date) (answerCard, error) {
	days := periodRange(p, date)
	label := days.Label
	resp, err := a.s.read.GetPerformance(r.Context(), bearer(&viewsv1.GetPerformanceRequest{From: optionalString(days.From), To: days.To.String()}, sess))
	if err != nil {
		return answerCard{}, fmt.Errorf("GetPerformance: %w", err)
	}
	series := resp.Msg.GetDays()
	if len(series) == 0 {
		return answerCard{Title: "No valued days " + label, Text: "The time-weighted return needs valued days."}, nil
	}
	first, last := series[0], series[len(series)-1]
	f := requestFormat(r)
	return answerCard{
		Title: "Time-weighted return " + label + ": " + returnPercent(first.GetIndex(), last.GetIndex(), f),
		Text:  "From " + first.GetDate() + " to " + last.GetDate() + ". Deposits, withdrawals and the transfers you confirmed don't count as return.",
		Rows: []answerRow{
			{Label: "Start", Detail: first.GetDate(), Figure: f.money(first.GetValueReporting())},
			{Label: "End", Detail: last.GetDate(), Figure: f.money(last.GetValueReporting())},
		},
		Link: &answerLink{Href: "/?" + url.Values{overParam: {overPerformance}, startParam: {first.GetDate()}}.Encode() + "#over-time", Text: "Show the performance chart"},
	}, nil
}

// returnPercent is the return from index first to index last in percent,
// signed, half-even to one place, in the number format f: "+9,4%"; empty
// when either isn't a decimal or first is zero.
func returnPercent(first, last string, f numberFormat) string {
	a, errA := kernel.ParseDecimal(first)
	b, errB := kernel.ParseDecimal(last)
	if errA != nil || errB != nil || a.IsZero() {
		return ""
	}
	q, err := b.Sub(a).Mul(kernel.NewDecimal(hundred, 0)).Quo(a, twrPlaces, kernel.RoundHalfEven)
	if err != nil {
		return ""
	}
	sign, digits := "", q.String()
	if q.Sign() >= 0 {
		sign = "+"
	}
	if !strings.Contains(digits, ".") {
		digits += ".0"
	}
	return sign + f.amount(digits) + "%"
}

// value answers the portfolio's three values on day, or on latest, the
// latest valued day, when day is zero.
func (a commandAnswerer) value(r *http.Request, sess session, day, latest kernel.Date) (answerCard, error) {
	if day.IsZero() {
		day = latest
	}
	resp, err := a.s.read.ListValuationTotals(r.Context(), bearer(&viewsv1.ListValuationTotalsRequest{From: day.String(), To: day.String()}, sess))
	if err != nil {
		return answerCard{}, fmt.Errorf("ListValuationTotals: %w", err)
	}
	link := &answerLink{Href: "/?" + url.Values{dateParam: {day.String()}}.Encode(), Text: "Open the overview on " + day.String()}
	totals := resp.Msg.GetTotals()
	if len(totals) == 0 {
		return answerCard{Title: "Nothing is valued on " + day.String(), Text: "No account has a valuation that day.", Link: link}, nil
	}
	f := requestFormat(r)
	t := totals[len(totals)-1]
	for _, on := range totals {
		if on.GetDate() == day.String() {
			t = on
		}
	}
	return answerCard{
		Title: "Value on " + day.String() + ": " + f.money(t.GetValueReporting()) + " gross",
		Text:  "Holdings at their stored valuations plus cash, in the reporting currency.",
		Rows: []answerRow{
			{Label: "Gross", Figure: f.money(t.GetValueReporting())},
			{Label: "Net", Detail: "after the taxes due at the planned exit", Figure: orMissing(f.money(t.GetNetReporting()))},
			{Label: "Liquidation", Detail: "if everything were sold that day", Figure: orMissing(f.money(t.GetLiquidationReporting()))},
		},
		Link: link,
	}, nil
}

// orMissing is s, or "missing" when it is empty.
func orMissing(s string) string {
	if s == "" {
		return "missing"
	}
	return s
}
