package ui

import (
	"cmp"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// connectorCapability is the capability of a plugin that feeds accounts.
const connectorCapability = "connector"

// accountsData is what the account forms show (D-72): the accounts with
// the connector instances bound to each, the broker accounts the
// instances reported, mapped and unmapped, the sleeves and connector
// instances to pick from, the page the forms return to, and the core's
// reason when it refused one.
type accountsData struct {
	Owner     bool
	Accounts  []accountEntry
	Sleeves   []*viewsv1.Sleeve
	Instances []string
	// InstancesUnknown says the core can't read its cluster, so the
	// instances offered are only those it has seen.
	InstancesUnknown bool
	Brokers          []*viewsv1.BrokerAccount
	// Wrappers are the tax wrappers an account can have, the pack's
	// default first (D-76).
	Wrappers []string
	Next     string
	Error    string
}

// accountEntry is one account with the connector instances bound to it.
type accountEntry struct {
	Account   *viewsv1.Account
	Instances []string
}

// Unmapped returns the broker accounts no account is mapped to: their
// transactions are not booked until the owner picks an account for them.
func (d *accountsData) Unmapped() []*viewsv1.BrokerAccount {
	var out []*viewsv1.BrokerAccount
	for _, b := range d.Brokers {
		if b.GetAccountId() == "" {
			out = append(out, b)
		}
	}
	return out
}

// loadAccounts reads what the account forms show.
func (s *server) loadAccounts(r *http.Request, sess session, next string) (*accountsData, error) {
	ctx := r.Context()
	accounts, err := s.read.ListAccounts(ctx, bearer(&viewsv1.ListAccountsRequest{}, sess))
	if err != nil {
		return nil, fmt.Errorf("ListAccounts: %w", err)
	}
	sleeves, err := s.read.ListSleeves(ctx, bearer(&viewsv1.ListSleevesRequest{}, sess))
	if err != nil {
		return nil, fmt.Errorf("ListSleeves: %w", err)
	}
	bindings, err := s.read.ListConnectorBindings(ctx, bearer(&viewsv1.ListConnectorBindingsRequest{}, sess))
	if err != nil {
		return nil, fmt.Errorf("ListConnectorBindings: %w", err)
	}
	brokers, err := s.read.ListBrokerAccounts(ctx, bearer(&viewsv1.ListBrokerAccountsRequest{}, sess))
	if err != nil {
		return nil, fmt.Errorf("ListBrokerAccounts: %w", err)
	}
	data := &accountsData{
		Owner: sess.Owner(), Sleeves: sleeves.Msg.GetSleeves(), Brokers: brokers.Msg.GetBrokerAccounts(), Next: next,
		Wrappers: sess.Settings.GetTaxWrappers(),
	}
	for _, a := range accounts.Msg.GetAccounts() {
		e := accountEntry{Account: a}
		for _, b := range bindings.Msg.GetBindings() {
			if b.GetAccountId() == a.GetId() {
				e.Instances = append(e.Instances, b.GetConnectorInstance())
			}
		}
		data.Accounts = append(data.Accounts, e)
	}
	data.Instances, data.InstancesUnknown = s.connectorInstances(r, sess, bindings.Msg.GetBindings(), data.Brokers)
	return data, nil
}

// connectorInstances returns the installed connector instances, in name
// order, to bind accounts to. When the core can't read its cluster it
// returns those the core has seen, bound or reporting broker accounts, and
// true.
func (s *server) connectorInstances(r *http.Request, sess session, bindings []*viewsv1.ConnectorBinding, brokers []*viewsv1.BrokerAccount) ([]string, bool) {
	var out []string
	plugins, err := s.read.ListPlugins(r.Context(), bearer(&viewsv1.ListPluginsRequest{}, sess))
	if err == nil {
		for _, p := range plugins.Msg.GetPlugins() {
			if slices.Contains(p.GetCapabilities(), connectorCapability) || slices.Contains(p.GetRequestedCapabilities(), connectorCapability) {
				out = append(out, p.GetInstance())
			}
		}
		slices.Sort(out)
		return out, false
	}
	for _, b := range bindings {
		out = append(out, b.GetConnectorInstance())
	}
	for _, b := range brokers {
		out = append(out, b.GetConnectorInstance())
	}
	slices.Sort(out)
	return slices.Compact(out), true
}

// accountsPage shows the accounts, the connector instances bound to each
// and the broker accounts, mapped and unmapped (D-72), with the owner's
// forms to create and change accounts and bind their connectors.
func (s *server) accountsPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderAccounts(w, r, sess, http.StatusOK, accountsPath, "")
}

// renderAccounts shows the account forms on the page next names, the
// accounts page or the wizard's step, with message as the core's refusal.
func (s *server) renderAccounts(w http.ResponseWriter, r *http.Request, sess session, status int, next, message string) {
	if strings.HasPrefix(next, "/setup") {
		s.renderSetup(w, r, sess, status, stepAccounts, message)
		return
	}
	data, err := s.loadAccounts(r, sess, accountsPath)
	if err != nil {
		s.coreFailed(w, r, "ListAccounts", err)
		return
	}
	data.Error = message
	s.show(w, r, status, "accounts", page{Title: "Accounts", Session: &sess, Data: data})
}

// accountEntryDone sends the browser back where the form came from, or,
// when the core refused the entry, shows the form's page again with the
// reason.
func (s *server) accountEntryDone(w http.ResponseWriter, r *http.Request, sess session, method string, err error) {
	next := returnTo(r.PostFormValue("next"), accountsPath)
	if err == nil {
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}
	if status, refused := refusalStatus(connect.CodeOf(err)); refused {
		s.renderAccounts(w, r, sess, status, next, connectMessage(err))
		return
	}
	s.coreFailed(w, r, method, err)
}

// brokerPick is a posted broker account choice: the instance that
// reported it and its ID.
type brokerPick struct {
	Instance, ID string
}

// brokerChoice reads a posted broker account choice, "instance/id" as the
// form offers it; an empty choice is neither. An instance name never holds
// a slash, so the first one splits them.
func brokerChoice(v string) brokerPick {
	instance, id, _ := strings.Cut(v, "/")
	return brokerPick{Instance: instance, ID: id}
}

// createAccount hands a new account to the core (CreateAccount): its name,
// institution, sleeve, tax wrapper (D-76), sale fee rule (D-78), and the connector instance that feeds it, or a
// broker account that instance reported, picked from a list so the owner
// never types an account number (D-72).
func (s *server) createAccount(w http.ResponseWriter, r *http.Request, sess session) {
	instance := r.PostFormValue("instance")
	broker := ""
	if choice := r.PostFormValue("broker"); choice != "" {
		pick := brokerChoice(choice)
		instance, broker = pick.Instance, pick.ID
	}
	_, err := s.accounts.CreateAccount(r.Context(), connect.NewRequest(&ingestv1.CreateAccountRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Name: strings.TrimSpace(r.PostFormValue("name")),
		Institution: strings.TrimSpace(r.PostFormValue("institution")), Sleeve: r.PostFormValue("sleeve"),
		ConnectorInstance: instance, BrokerAccount: broker,
		TaxWrapper: r.PostFormValue("tax_wrapper"), FeeRule: postedFeeRule(r),
	}))
	s.accountEntryDone(w, r, sess, "CreateAccount", err)
}

// updateAccount hands an account's new name, sleeve, tax wrapper or sale
// fee rule to the core (UpdateAccount).
func (s *server) updateAccount(w http.ResponseWriter, r *http.Request, sess session) {
	_, err := s.accounts.UpdateAccount(r.Context(), connect.NewRequest(&ingestv1.UpdateAccountRequest{
		Origin: sess.Origin, SessionToken: sess.Token, AccountId: r.PathValue("id"),
		Name: strings.TrimSpace(r.PostFormValue("name")), Sleeve: r.PostFormValue("sleeve"),
		TaxWrapper: r.PostFormValue("tax_wrapper"), FeeRule: postedFeeRule(r),
	}))
	s.accountEntryDone(w, r, sess, "UpdateAccount", err)
}

// bindAccount binds a connector instance to an account, or maps a broker
// account an instance reported to it (BindAccount), as `core connector
// bind` does.
func (s *server) bindAccount(w http.ResponseWriter, r *http.Request, sess session) {
	instance, broker := r.PostFormValue("instance"), ""
	if choice := r.PostFormValue("broker"); choice != "" {
		pick := brokerChoice(choice)
		instance, broker = pick.Instance, pick.ID
	}
	_, err := s.accounts.BindAccount(r.Context(), connect.NewRequest(&ingestv1.BindAccountRequest{
		Origin: sess.Origin, SessionToken: sess.Token, AccountId: r.PostFormValue("account"),
		ConnectorInstance: instance, BrokerAccount: broker,
	}))
	s.accountEntryDone(w, r, sess, "BindAccount", err)
}

// sleeveChoice is what an account form's sleeve select shows: its ID, the
// sleeves and the one chosen.
type sleeveChoice struct {
	ID      string
	Sleeves []*viewsv1.Sleeve
	Current string
}

// sleeveSelect returns the sleeve select with id over sleeves, current
// chosen.
func sleeveSelect(id string, sleeves []*viewsv1.Sleeve, current string) sleeveChoice {
	return sleeveChoice{ID: id, Sleeves: sleeves, Current: current}
}

// fromPercent turns a percentage the owner typed, such as "0,29" or
// "0.29", into the decimal fraction the core takes, "0.0029". Text that
// isn't a decimal goes to the core as typed, which refuses it with its
// reason.
func fromPercent(typed string) string {
	typed = strings.ReplaceAll(strings.TrimSpace(typed), ",", ".")
	d, err := kernel.ParseDecimal(typed)
	if err != nil {
		return typed
	}
	return d.Mul(kernel.NewDecimal(1, -2)).String()
}

// toPercent writes a decimal fraction as the percentage a form field
// shows, without the sign: "0.0029" is "0.29"; empty for zero or none.
func toPercent(fraction string) string {
	if isZero(fraction) {
		return ""
	}
	return strings.TrimSuffix(percent(fraction), "%")
}

// postedFeeRule reads the fee rule fields of an account form (D-78): the
// sale rate and conversion rate in percent, and the minimum with its
// currency. All empty is nil: no rule on a new account, the account's own
// on a change.
func postedFeeRule(r *http.Request) *ingestv1.FeeRule {
	rate, conversion := r.PostFormValue("fee_rate"), r.PostFormValue("fee_conversion")
	minimum := strings.ReplaceAll(strings.TrimSpace(r.PostFormValue("fee_minimum")), ",", ".")
	currency := strings.ToUpper(strings.TrimSpace(r.PostFormValue("fee_minimum_currency")))
	if strings.TrimSpace(rate) == "" && strings.TrimSpace(conversion) == "" && minimum == "" {
		return nil
	}
	rule := &ingestv1.FeeRule{Rate: cmp.Or(fromPercent(rate), "0"), ConversionRate: cmp.Or(fromPercent(conversion), "0")}
	if minimum != "" && !isZero(minimum) {
		rule.Minimum, rule.MinimumCurrency = minimum, currency
	}
	return rule
}

// termsForm is what an account form's tax wrapper and fee rule fields
// show: their ID prefix, the wrappers to pick from, and the account's own
// terms, or none for a new account.
type termsForm struct {
	Prefix     string
	Wrappers   []string
	Wrapper    string
	Rate       string
	Minimum    string
	Currency   string
	Conversion string
}

// termsFields returns the tax wrapper and fee rule fields with IDs under
// prefix over wrappers, filled with acc's terms; nil acc is a new account.
func termsFields(prefix string, wrappers []string, acc *viewsv1.Account) termsForm {
	rule := acc.GetFeeRule()
	f := termsForm{
		Prefix: prefix, Wrappers: wrappers, Wrapper: acc.GetTaxWrapper(),
		Rate: toPercent(rule.GetRate()), Conversion: toPercent(rule.GetConversionRate()),
	}
	if !isZero(rule.GetMinimum().GetAmount()) {
		f.Minimum, f.Currency = rule.GetMinimum().GetAmount(), rule.GetMinimum().GetCurrency()
	}
	return f
}
