package answer

import (
	"cmp"
	"context"
	"fmt"
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
)

// Card is a command's answer to one question (D-194): a title, a sentence, a
// small table and one link to the place it summarizes, or, for a question the
// box doesn't understand, the commands it does. It is the data of an answer,
// not its drawing: the page renders it as HTML (the UI) and the MCP server
// serialises it as JSON (CardResult). Nothing in a card changes anything.
type Card struct {
	Title string
	Text  string
	Rows  []Row
	// Link is the one "Open …" link to the place the answer summarizes;
	// nil when there is none.
	Link *Link
	// Help are the commands the box understands, for a question it
	// doesn't; empty otherwise.
	Help []Help
}

// Row is one row of an answer's table: what it is about, a detail
// and a figure, and where it links, if anywhere.
type Row struct {
	Label, Detail, Figure, Href string
}

// Link is an answer's link to the page it summarizes.
type Link struct {
	Href, Text string
}

// Reader is the part of the core's read API the answerers read: the views'
// ReadServiceClient implements it. Every call carries the caller's own token,
// so the caller's grants limit what a card shows.
type Reader interface {
	StatusReader
	CatalogReader
	ValuationReader
	QuoteReader
}

// StatusReader reads the owner's settings, what needs the owner and the
// connectors' syncs.
type StatusReader interface {
	GetSettings(ctx context.Context, req *connect.Request[viewsv1.GetSettingsRequest]) (*connect.Response[viewsv1.GetSettingsResponse], error)
	ListAttentionItems(ctx context.Context, req *connect.Request[viewsv1.ListAttentionItemsRequest]) (*connect.Response[viewsv1.ListAttentionItemsResponse], error)
	ListSyncStatuses(ctx context.Context, req *connect.Request[viewsv1.ListSyncStatusesRequest]) (*connect.Response[viewsv1.ListSyncStatusesResponse], error)
	ListSyncAttempts(ctx context.Context, req *connect.Request[viewsv1.ListSyncAttemptsRequest]) (*connect.Response[viewsv1.ListSyncAttemptsResponse], error)
}

// CatalogReader reads what a show command can name: the accounts, the sleeves,
// the holdings and the instruments.
type CatalogReader interface {
	ListAccounts(ctx context.Context, req *connect.Request[viewsv1.ListAccountsRequest]) (*connect.Response[viewsv1.ListAccountsResponse], error)
	ListSleeves(ctx context.Context, req *connect.Request[viewsv1.ListSleevesRequest]) (*connect.Response[viewsv1.ListSleevesResponse], error)
	ListHoldings(ctx context.Context, req *connect.Request[viewsv1.ListHoldingsRequest]) (*connect.Response[viewsv1.ListHoldingsResponse], error)
	GetInstrument(ctx context.Context, req *connect.Request[viewsv1.GetInstrumentRequest]) (*connect.Response[viewsv1.GetInstrumentResponse], error)
	ListInstrumentHoldings(ctx context.Context, req *connect.Request[viewsv1.ListInstrumentHoldingsRequest]) (*connect.Response[viewsv1.ListInstrumentHoldingsResponse], error)
}

// ValuationReader reads the stored valuations and what is made of them.
type ValuationReader interface {
	ListValuationStatuses(ctx context.Context, req *connect.Request[viewsv1.ListValuationStatusesRequest]) (*connect.Response[viewsv1.ListValuationStatusesResponse], error)
	ListAccountValuations(ctx context.Context, req *connect.Request[viewsv1.ListAccountValuationsRequest]) (*connect.Response[viewsv1.ListAccountValuationsResponse], error)
	ListValuationTotals(ctx context.Context, req *connect.Request[viewsv1.ListValuationTotalsRequest]) (*connect.Response[viewsv1.ListValuationTotalsResponse], error)
	GetAllocation(ctx context.Context, req *connect.Request[viewsv1.GetAllocationRequest]) (*connect.Response[viewsv1.GetAllocationResponse], error)
	GetPerformance(ctx context.Context, req *connect.Request[viewsv1.GetPerformanceRequest]) (*connect.Response[viewsv1.GetPerformanceResponse], error)
	GetPeriodChange(ctx context.Context, req *connect.Request[viewsv1.GetPeriodChangeRequest]) (*connect.Response[viewsv1.GetPeriodChangeResponse], error)
}

// Caller is who asks: the token every read carries, a session's or a script
// token's, and the number format of the owner's settings (D-113), which the
// cards write amounts in.
type Caller struct {
	Token  string
	Format NumberFormat
}

// bearer returns msg as a request that carries the caller's token as its
// bearer token: the read API takes the owner's session like a script token
// (D-6).
func bearer[T any](msg *T, c Caller) *connect.Request[T] {
	req := connect.NewRequest(msg)
	req.Header().Set("Authorization", "Bearer "+c.Token)
	return req
}

// Commands answers the Ask box's commands (Parse) from the read API, with the
// caller's token, as every page reads it. The page's loop, its model's tools
// and the MCP server's tools all run it. It never changes anything, and never
// logs or sends a question anywhere (GR-12).
type Commands struct {
	read  Reader
	today func() kernel.Date
	links Links
}

// NewCommands returns Commands that read with read, take the date of a
// question from today, and link cards to the pages links names; nil links is
// NoLinks. Without a clock today is the UTC date.
func NewCommands(read Reader, today func() kernel.Date, links Links) *Commands {
	if today == nil {
		today = UTCToday
	}
	if links == nil {
		links = NoLinks{}
	}
	return &Commands{read: read, today: today, links: links}
}

// Today returns the date of a question: the clock the Commands were given.
func (cm *Commands) Today() kernel.Date { return cm.today() }

// Answer parses question and answers its command; a question none of the
// commands matches gets the commands the box understands.
func (cm *Commands) Answer(ctx context.Context, c Caller, question string) (Card, error) {
	q, ok := Parse(question, cm.today())
	if !ok {
		return UnknownAnswer(), nil
	}
	return cm.Run(ctx, c, &q)
}

// Run answers the command q understood, reading with the caller's token. It is
// what the model's tools call too (D-203): the same read-only answerers, with
// the asking person's own session. A query that is no command gets the
// commands the box understands.
func (cm *Commands) Run(ctx context.Context, c Caller, q *Query) (Card, error) {
	switch q.Command {
	case CommandUnknown:
		return UnknownAnswer(), nil
	case CommandQuote:
		return cm.quote(ctx, c, q)
	case CommandCalc: // the figures are the arguments: nothing is read
		return calcCard(q.Expression, c.Format), nil
	case CommandNeeds, CommandShow, CommandSync, CommandAllocation, CommandPerformance, CommandValue:
	}
	date, err := cm.latestDate(ctx, c)
	if err != nil {
		return Card{}, err
	}
	return cm.dispatch(ctx, c, q, date)
}

// dispatch answers one of the commands that read the valued portfolio, on date,
// the latest valued day.
func (cm *Commands) dispatch(ctx context.Context, c Caller, q *Query, date kernel.Date) (Card, error) {
	switch q.Command {
	case CommandNeeds:
		return cm.needs(ctx, c)
	case CommandShow:
		return cm.show(ctx, c, q, date)
	case CommandSync:
		return cm.sync(ctx, c, q.Term)
	case CommandAllocation:
		return cm.allocation(ctx, c, ParseDimension(q.Dimension), date)
	case CommandPerformance:
		return cm.performance(ctx, c, q.Period, date)
	case CommandValue:
		return cm.valueAnswer(ctx, c, q, date)
	case CommandUnknown, CommandQuote, CommandCalc:
	}
	return UnknownAnswer(), nil
}

// UnknownAnswer answers a question the box doesn't understand with the
// commands it does.
func UnknownAnswer() Card {
	return Card{
		Title: "I can't answer that yet",
		Text:  "The Ask box understands a few commands and answers them from your own record; it holds no AI model. Try one of these.",
		Help:  KnownCommands(),
	}
}

// latestDate returns the latest day any account is valued through, or
// today when none is.
func (cm *Commands) latestDate(ctx context.Context, c Caller) (kernel.Date, error) {
	statuses, err := cm.read.ListValuationStatuses(ctx, bearer(&viewsv1.ListValuationStatusesRequest{}, c))
	if err != nil {
		return kernel.Date{}, fmt.Errorf("ListValuationStatuses: %w", err)
	}
	return LatestValued(statuses.Msg.GetStatuses(), cm.today()), nil
}

// needs answers what needs the owner: the open attention items, the
// Inbox's first ones.
func (cm *Commands) needs(ctx context.Context, c Caller) (Card, error) {
	resp, err := cm.read.ListAttentionItems(ctx, bearer(&viewsv1.ListAttentionItemsRequest{View: viewsv1.AttentionView_ATTENTION_VIEW_OPEN}, c))
	if err != nil {
		return Card{}, fmt.Errorf("ListAttentionItems: %w", err)
	}
	n := int(resp.Msg.GetOpen())
	card := Card{
		Title: Plural(n, "thing needs you", "things need you"),
		Text:  "The Inbox gathers every check that found something for you, grouped by what you do about it.",
		Link:  &Link{Href: cm.links.Inbox(), Text: "Open the Inbox (" + strconv.Itoa(n) + ")"},
	}
	if n == 0 {
		card.Title, card.Text = "Nothing needs you", "No check has found anything new for you to look at."
	}
	for i, it := range resp.Msg.GetItems() {
		if i == answerRows {
			break
		}
		card.Rows = append(card.Rows, Row{
			Label: ItemTitle(it), Detail: GroupTitle(it.GetGroup()), Figure: itemWord(it), Href: cm.links.InboxItem(it.GetId()),
		})
	}
	return card, nil
}

// Plural writes n with one or many, "1 thing needs you" or "3 things
// need you".
func Plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}

// show answers a show command: the account, sleeve or instrument the
// question names among the session's, on date.
func (cm *Commands) show(ctx context.Context, c Caller, q *Query, date kernel.Date) (Card, error) {
	cat, err := cm.askTargets(ctx, c, date)
	if err != nil {
		return Card{}, err
	}
	targets, accounts, sleeves := cat.Targets, cat.Accounts, cat.Sleeves
	t, ok := matchTarget(q.Term, q.Show, targets)
	if !ok {
		return Card{
			Title: "Nothing here is called “" + q.Term + "”",
			Text:  "Show finds an account, a sleeve or an instrument by its name, ticker or ISIN.",
			Rows:  targetRows(nearTargets(q.Term, targets)),
		}, nil
	}
	switch t.Kind {
	case ShowAccount:
		return cm.accountAnswer(ctx, c, accounts[t.ID], sleeves, date)
	case ShowSleeve:
		return cm.sleeveAnswer(ctx, c, t, accounts, date)
	case ShowInstrument, ShowAny:
	}
	return cm.instrumentAnswer(ctx, c, t, accounts, date)
}

// nearTargets returns targets, those that hold the most words of term first, each
// score in its order, so a name that matches none but whose words match several
// lists those several first: it is the choice a model is asked to put to the
// person. Words are of at least three letters.
func nearTargets(term string, targets []askTarget) []askTarget {
	var words []string
	for w := range strings.FieldsSeq(strings.ToLower(term)) {
		if len([]rune(w)) >= minNearWord {
			words = append(words, w)
		}
	}
	score := func(t askTarget) int {
		n := 0
		for _, w := range words {
			if slices.ContainsFunc(t.Words, func(tw string) bool { return strings.Contains(strings.ToLower(tw), w) }) {
				n++
			}
		}
		return n
	}
	out := slices.Clone(targets)
	slices.SortStableFunc(out, func(a, b askTarget) int { return cmp.Compare(score(b), score(a)) })
	return out
}

// minNearWord is the shortest word of a name that makes a target near it.
const minNearWord = 3

// targetRows lists the first targets a show command can name, as rows.
func targetRows(targets []askTarget) []Row {
	var out []Row
	for i, t := range targets {
		if i == answerRows {
			break
		}
		out = append(out, Row{Label: t.Label, Detail: [...]string{"", wordAccount, DimSleeve, wordInstrument}[t.Kind]})
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
func (cm *Commands) askTargets(ctx context.Context, c Caller, date kernel.Date) (askCatalog, error) {
	accounts, err := cm.read.ListAccounts(ctx, bearer(&viewsv1.ListAccountsRequest{}, c))
	if err != nil {
		return askCatalog{}, fmt.Errorf("ListAccounts: %w", err)
	}
	sleeves, err := cm.read.ListSleeves(ctx, bearer(&viewsv1.ListSleevesRequest{}, c))
	if err != nil {
		return askCatalog{}, fmt.Errorf("ListSleeves: %w", err)
	}
	byID, names := map[string]*viewsv1.Account{}, map[string]string{}
	var targets []askTarget
	for _, a := range accounts.Msg.GetAccounts() {
		byID[a.GetId()] = a
		targets = append(targets, askTarget{Kind: ShowAccount, ID: a.GetId(), Label: a.GetName(), Words: []string{a.GetName(), a.GetId(), a.GetInstitution()}})
	}
	for _, sl := range sleeves.Msg.GetSleeves() {
		names[sl.GetId()] = sl.GetName()
		targets = append(targets, askTarget{Kind: ShowSleeve, ID: sl.GetId(), Label: sl.GetName(), Words: []string{sl.GetName(), sl.GetId(), sl.GetKind()}})
	}
	seen := map[string]bool{}
	for _, a := range accounts.Msg.GetAccounts() {
		holdings, hErr := cm.read.ListHoldings(ctx, bearer(&viewsv1.ListHoldingsRequest{AccountId: a.GetId(), Date: date.String()}, c))
		if hErr != nil {
			return askCatalog{}, fmt.Errorf("ListHoldings: %w", hErr)
		}
		for _, h := range holdings.Msg.GetHoldings() {
			in := h.GetInstrument()
			if seen[in.GetIsin()] {
				continue
			}
			seen[in.GetIsin()] = true
			targets = append(targets, askTarget{Kind: ShowInstrument, ID: in.GetIsin(), Label: InstrumentLabel(in), Words: []string{in.GetName(), in.GetSymbol(), in.GetIsin()}})
		}
	}
	return askCatalog{Targets: targets, Accounts: byID, Sleeves: names}, nil
}

// accountAnswer answers a show command naming account a: its three values
// and its largest holdings on date.
func (cm *Commands) accountAnswer(ctx context.Context, c Caller, a *viewsv1.Account, sleeves map[string]string, date kernel.Date) (Card, error) {
	f := c.Format
	vals, err := cm.read.ListAccountValuations(ctx, bearer(&viewsv1.ListAccountValuationsRequest{AccountId: a.GetId(), From: date.String(), To: date.String()}, c))
	if err != nil {
		return Card{}, fmt.Errorf("ListAccountValuations: %w", err)
	}
	card := Card{
		Title: a.GetName() + ": not valued on " + date.String(),
		Text:  strings.Join(nonEmpty(a.GetInstitution(), sleeves[a.GetSleeve()], WrapperLabel(a.GetTaxWrapper())), " · ") + ". Holdings at their stored valuations, in the reporting currency.",
		Link:  &Link{Href: cm.links.Account(a.GetId()), Text: "Open " + a.GetName()},
	}
	if totals := vals.Msg.GetTotals(); len(totals) > 0 {
		card.Title = a.GetName() + ": " + f.Money(totals[len(totals)-1].GetValueReporting()) + " gross"
	}
	for i, v := range vals.Msg.GetValuations() {
		if i == answerRows {
			break
		}
		card.Rows = append(card.Rows, Row{Label: InstrumentLabel(v.GetInstrument()), Detail: EnumLabel(v.GetBasis().String(), "VALUATION_BASIS_"), Figure: f.Money(v.GetValueReporting())})
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
func (cm *Commands) sleeveAnswer(ctx context.Context, c Caller, t askTarget, accounts map[string]*viewsv1.Account, date kernel.Date) (Card, error) {
	f := c.Format
	card := Card{
		Title: t.Label + ": no accounts", Text: "Each account in the sleeve, gross with its cash, in the reporting currency.",
		Link: &Link{Href: cm.links.Sleeve(t.ID), Text: "Open " + t.Label + " on the overview"},
	}
	var sum []*viewsv1.Money
	for _, a := range sortedAccounts(accounts) {
		if a.GetSleeve() != t.ID {
			continue
		}
		vals, err := cm.read.ListAccountValuations(ctx, bearer(&viewsv1.ListAccountValuationsRequest{AccountId: a.GetId(), From: date.String(), To: date.String()}, c))
		if err != nil {
			return Card{}, fmt.Errorf("ListAccountValuations: %w", err)
		}
		row := Row{Label: a.GetName(), Detail: a.GetInstitution(), Figure: "not valued", Href: cm.links.Account(a.GetId())}
		if totals := vals.Msg.GetTotals(); len(totals) > 0 {
			gross := totals[len(totals)-1].GetValueReporting()
			row.Figure, sum = f.Money(gross), append(sum, gross)
		}
		card.Rows = append(card.Rows, row)
	}
	if len(card.Rows) > 0 {
		card.Title = t.Label + ": " + Plural(len(card.Rows), "account", "accounts")
		if total := SumMoney(sum); total != nil {
			card.Title += ", " + f.Money(total) + " gross"
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
func (cm *Commands) instrumentAnswer(ctx context.Context, c Caller, t askTarget, accounts map[string]*viewsv1.Account, date kernel.Date) (Card, error) {
	in, err := cm.read.GetInstrument(ctx, bearer(&viewsv1.GetInstrumentRequest{Isin: t.ID}, c))
	if err != nil {
		return Card{}, fmt.Errorf("GetInstrument: %w", err)
	}
	held, err := cm.read.ListInstrumentHoldings(ctx, bearer(&viewsv1.ListInstrumentHoldingsRequest{Isin: t.ID, Date: date.String()}, c))
	if err != nil {
		return Card{}, fmt.Errorf("ListInstrumentHoldings: %w", err)
	}
	f := c.Format
	d := in.Msg.GetInstrument()
	card := Card{
		Title: strings.Join(nonEmpty(d.GetName(), t.ID), " · "),
		Text:  "ISIN " + d.GetIsin() + ", " + strings.ReplaceAll(d.GetAssetClass(), "_", " ") + ".",
	}
	for _, l := range d.GetListings() {
		if p := l.GetLatestPrice(); p != nil {
			card.Text += " " + l.GetInstrument().GetSymbol() + " last priced " + f.Money(p.GetClose()) + " on " + p.GetDate() + " from " + l.GetPriceSource() + "."
		}
	}
	for _, h := range held.Msg.GetHoldings() {
		name := h.GetAccountId()
		if a, ok := accounts[name]; ok {
			name = a.GetName()
		}
		card.Rows = append(card.Rows, Row{Label: name, Detail: h.GetInstrument().GetSymbol(), Figure: f.Amount(h.GetQuantity()) + " units", Href: cm.links.Account(h.GetAccountId())})
	}
	card.Link = &Link{Href: cm.links.Instrument(d.GetIsin()), Text: "Open " + card.Title}
	return card, nil
}

// sync answers why a sync fails: the instance term names, its status and
// its latest attempts, each with its error class, never its text (D-198).
func (cm *Commands) sync(ctx context.Context, c Caller, term string) (Card, error) {
	statuses, err := cm.read.ListSyncStatuses(ctx, bearer(&viewsv1.ListSyncStatusesRequest{}, c))
	if err != nil {
		return Card{}, fmt.Errorf("ListSyncStatuses: %w", err)
	}
	var instances []string
	for _, st := range statuses.Msg.GetStatuses() {
		instances = append(instances, st.GetConnectorInstance())
	}
	inst, ok := matchInstance(term, instances)
	if !ok {
		card := Card{Title: "No connector called “" + term + "” has synced", Text: "These have:", Link: &Link{Href: cm.links.Plugins(), Text: openPlugins}}
		for _, st := range statuses.Msg.GetStatuses() {
			card.Rows = append(card.Rows, syncStatusRow(st))
		}
		return card, nil
	}
	resp, err := cm.read.ListSyncAttempts(ctx, bearer(&viewsv1.ListSyncAttemptsRequest{ConnectorInstance: inst, PageSize: answerAttempts}, c))
	if err != nil && connect.CodeOf(err) != connect.CodeNotFound {
		return Card{}, fmt.Errorf("ListSyncAttempts: %w", err)
	}
	return syncAnswer(inst, resp.Msg, cm.today(), cm.links), nil
}

// syncAnswer turns an instance's status and latest attempts into its answer,
// today being today: when it started failing and for how long, how many attempts
// in a row failed and with which error class, the last good sync and the next
// try, then each of the latest attempts with its time and its error class.
func syncAnswer(inst string, resp *viewsv1.ListSyncAttemptsResponse, today kernel.Date, links Links) Card {
	st := resp.GetStatus()
	card := Card{
		Title: "The " + inst + " sync is working",
		Text:  "Last good sync: " + orNone(st.GetLastSuccessAt()) + ". Next try: " + orNone(st.GetNextSyncAt()) + ". Every attempt keeps its error class, never its text.",
		Link:  &Link{Href: links.Plugins(), Text: openPlugins},
	}
	if st.GetFailures() > 0 {
		card.Text = failureSummary(st, today) + " " + card.Text
	}
	switch {
	case st.GetKeepsFailing():
		card.Title = "The " + inst + " sync keeps failing since " + st.GetFailingSince()
		card.Link = &Link{Href: links.Inbox(), Text: "Open the failing sync in the Inbox"}
	case st.GetFailures() > 0:
		card.Title = "The " + inst + " sync failed " + Plural(int(st.GetFailures()), "time", "times") + " in a row"
	}
	for _, at := range resp.GetAttempts() {
		detail, outcome := at.GetMode()+" sync", "went through"
		if at.GetFailed() {
			detail, outcome = at.GetMode()+" sync failed, error class", errorClassLabel(at.GetErrorClass())
		}
		card.Rows = append(card.Rows, Row{Label: at.GetStartedAt(), Detail: detail, Figure: outcome})
	}
	return card
}

// errorClassLabel names a sync's error class in words: "timeout", "unavailable".
func errorClassLabel(c viewsv1.SyncErrorClass) string {
	return EnumLabel(c.String(), "SYNC_ERROR_CLASS_")
}

// failureSummary says how long a sync that failed has been failing, how many
// attempts in a row failed and with which error class, today being today. It
// reads like "Failing for 3 days, since 2026-10-25: 4 failed attempts in a row,
// the latest with the error class unavailable".
func failureSummary(st *viewsv1.SyncStatus, today kernel.Date) string {
	out := "Failing"
	if since := st.GetFailingSince(); since != "" {
		day := since[:min(len(since), len(isoDay))]
		if d, err := kernel.ParseDate(day); err == nil && !d.After(today) {
			out += " for " + days(d.DaysUntil(today)) + ", since " + day
		} else {
			out += " since " + since
		}
	}
	out += ": " + Plural(int(st.GetFailures()), "failed attempt", "failed attempts") + " in a row"
	if c := st.GetLastErrorClass(); c != viewsv1.SyncErrorClass_SYNC_ERROR_CLASS_UNSPECIFIED {
		out += ", the latest with the error class " + errorClassLabel(c)
	}
	return out + "."
}

// isoDay is a day written as 2026-10-28: the first characters of a time stamp.
const isoDay = "2006-01-02"

// days writes a number of days: "less than a day", "1 day", "3 days".
func days(n int) string {
	if n < 1 {
		return "less than a day"
	}
	return Plural(n, "day", "days")
}

// syncStatusRow is one instance's status as a row.
func syncStatusRow(st *viewsv1.SyncStatus) Row {
	state := "working"
	switch {
	case st.GetKeepsFailing():
		state = "keeps failing"
	case st.GetFailures() > 0:
		state = "failed " + strconv.Itoa(int(st.GetFailures())) + "×"
	}
	return Row{Label: st.GetConnectorInstance(), Detail: "last attempt " + orNone(st.GetLastAttemptAt()), Figure: state}
}

// orNone is s, or "none yet" when it is empty.
func orNone(s string) string {
	if s == "" {
		return "none yet"
	}
	return s
}

// allocation answers the allocation by d, gross, on date.
func (cm *Commands) allocation(ctx context.Context, c Caller, d Dimension, date kernel.Date) (Card, error) {
	resp, err := cm.read.GetAllocation(ctx, bearer(&viewsv1.GetAllocationRequest{Date: date.String(), By: d.Enum, Value: viewsv1.ValueKind_VALUE_KIND_GROSS}, c))
	if err != nil {
		return Card{}, fmt.Errorf("GetAllocation: %w", err)
	}
	f := c.Format
	total := resp.Msg.GetTotalReporting()
	card := Card{
		Title: "By " + strings.ToLower(d.Label) + ", gross, on " + date.String(),
		Text:  "Holdings at their stored valuations plus cash, in the reporting currency: " + f.Money(total) + " in all.",
		Link:  &Link{Href: cm.links.Allocation(d.Key), Text: "Open the allocation"},
	}
	for i, g := range resp.Msg.GetGroups() {
		if i == answerRows {
			break
		}
		card.Rows = append(card.Rows, Row{Label: g.GetKey(), Detail: Share(g.GetValueReporting().GetAmount(), total.GetAmount()), Figure: f.Money(g.GetValueReporting())})
	}
	return card, nil
}

// performance answers the time-weighted return over p, up to date, the
// latest valued day: every period, named or bounded, resolves to its days
// (periodRange) and is answered by the read API's GetPeriodChange, so one
// convention gives one card (D-61, D-194).
func (cm *Commands) performance(ctx context.Context, c Caller, p Period, date kernel.Date) (Card, error) {
	s := periodRange(p, date)
	if s.From.IsZero() && p.Kind == PeriodAll {
		from, err := cm.trackingStart(ctx, c, date)
		if err != nil {
			return Card{}, err
		}
		if from.IsZero() {
			return Card{Title: "No valued days " + s.Label, Text: "The time-weighted return needs valued days."}, nil
		}
		s.From = from
	}
	return cm.periodPerformance(ctx, c, s, date)
}

// trackingStart returns the day the portfolio's tracking began, the charts'
// default start (D-138), as the read API names it; zero when nothing is
// recorded. It asks for one day of the series, only for that start.
func (cm *Commands) trackingStart(ctx context.Context, c Caller, date kernel.Date) (kernel.Date, error) {
	resp, err := cm.read.GetPerformance(ctx, bearer(&viewsv1.GetPerformanceRequest{From: OptionalString(date), To: date.String()}, c))
	if err != nil {
		return kernel.Date{}, fmt.Errorf("GetPerformance: %w", err)
	}
	if resp.Msg.GetDefaultFrom() == "" {
		return kernel.Date{}, nil
	}
	from, err := kernel.ParseDate(resp.Msg.GetDefaultFrom())
	if err != nil {
		return kernel.Date{}, fmt.Errorf("GetPerformance default start: %w", err)
	}
	return from, nil
}

// valueAnswer answers a value command: its change over a period when it asks
// for one, else the three values on its day.
func (cm *Commands) valueAnswer(ctx context.Context, c Caller, q *Query, latest kernel.Date) (Card, error) {
	if q.Change != nil {
		return cm.periodValue(ctx, c, boundedSpan(*q.Change), latest)
	}
	return cm.value(ctx, c, q.Date, latest)
}

// value answers the portfolio's three values on day, or on latest, the
// latest valued day, when day is zero.
func (cm *Commands) value(ctx context.Context, c Caller, day, latest kernel.Date) (Card, error) {
	if day.IsZero() {
		day = latest
	}
	resp, err := cm.read.ListValuationTotals(ctx, bearer(&viewsv1.ListValuationTotalsRequest{From: day.String(), To: day.String()}, c))
	if err != nil {
		return Card{}, fmt.Errorf("ListValuationTotals: %w", err)
	}
	link := &Link{Href: cm.links.Overview(day.String()), Text: "Open the overview on " + day.String()}
	totals := resp.Msg.GetTotals()
	if len(totals) == 0 {
		return Card{Title: "Nothing is valued on " + day.String(), Text: "No account has a valuation that day.", Link: link}, nil
	}
	f := c.Format
	t := totals[len(totals)-1]
	for _, on := range totals {
		if on.GetDate() == day.String() {
			t = on
		}
	}
	return Card{
		Title: "Value on " + day.String() + ": " + f.Money(t.GetValueReporting()) + " gross",
		Text:  "Holdings at their stored valuations plus cash, in the reporting currency.",
		Rows: []Row{
			{Label: "Gross", Figure: f.Money(t.GetValueReporting())},
			{Label: "Net", Detail: "after the taxes due at the planned exit", Figure: orMissing(f.Money(t.GetNetReporting()))},
			{Label: "Liquidation", Detail: "if everything were sold that day", Figure: orMissing(f.Money(t.GetLiquidationReporting()))},
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

// openPlugins is the words of a link to the installed plugins.
const openPlugins = "Open the plugins"
