package ui

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// holdingsPath is the holdings behind one allocation bar of an asset class
// or a currency (D-195).
const holdingsPath = "/holdings"

// assetCash is the allocation's asset class of cash.
const assetCash = "cash"

// holdingsLink is the holdings behind the allocation group key of
// dimension d on date.
func holdingsLink(d, key, date string) string {
	return holdingsPath + "?" + url.Values{"by": {d}, "key": {key}, dateParam: {date}}.Encode()
}

// filteredHolding is one holding behind an allocation bar, with its stored
// valuation, in an account.
type filteredHolding struct {
	AccountID, Account string
	Valuation          *viewsv1.Valuation
}

// filteredCash is one account's cash in one currency behind an allocation
// bar.
type filteredCash struct {
	AccountID, Account string
	Balance            *viewsv1.Money
}

// holdingsData is what the holdings behind an allocation bar show (D-195):
// every valued holding and cash balance of the group, in the accounts the
// caller may view, each holding linked to its instrument.
type holdingsData struct {
	Date      string
	Dimension dimension
	Key       string
	Holdings  []filteredHolding
	Cash      []filteredCash
	// Total is the holdings' gross value in the reporting currency; nil
	// without holdings.
	Total *viewsv1.Money
}

// Label names the group: an asset class in words, or a currency code.
func (d *holdingsData) Label() string {
	if d.Dimension.Key == dimAssetClass {
		return enumLabel(d.Key, "")
	}
	return d.Key
}

// Crumbs is the page's breadcrumb, through the overview's allocation.
func (d *holdingsData) Crumbs() []crumb {
	back := "/?" + url.Values{"by": {d.Dimension.Key}, dateParam: {d.Date}, overParam: {overAllocation}}.Encode() + "#over-time"
	return []crumb{overviewCrumb(), {Href: back, Text: "Allocation by " + enumLabel(d.Dimension.Label, "")}, {Text: d.Label()}}
}

// inGroup reports whether a holding valued as v falls in the group key of
// dimension d, its instrument's asset class being class.
func inGroup(d, key string, v *viewsv1.Valuation, class string) bool {
	if d == dimCurrency {
		return v.GetValue().GetCurrency() == key
	}
	return class == key
}

// holdingsPage shows the holdings behind one allocation bar of an asset
// class or a currency on a date (D-195).
func (s *server) holdingsPage(w http.ResponseWriter, r *http.Request, sess session) {
	q := r.URL.Query()
	d := parseDimension(q.Get("by"))
	if (d.Key != dimAssetClass && d.Key != dimCurrency) || q.Get("key") == "" {
		s.errorPage(w, r, http.StatusBadRequest, "Holdings are listed by asset class or currency, with the group's key.")
		return
	}
	date, err := asOfDate(q.Get(dateParam), nil, s.today())
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, err.Error())
		return
	}
	data := &holdingsData{Date: date.String(), Dimension: d, Key: q.Get("key")}
	if method, readErr := s.filteredHoldings(r, sess, date, data); readErr != nil {
		s.coreFailed(w, r, method, readErr)
		return
	}
	s.show(w, r, http.StatusOK, "holdings", page{Title: "Holdings · " + data.Label(), Session: &sess, Data: data})
}

// filteredHoldings reads, for every account, its valued holdings and its
// cash on date, and keeps those in data's group. It returns the method
// that failed with its error.
func (s *server) filteredHoldings(r *http.Request, sess session, date kernel.Date, data *holdingsData) (string, error) {
	ctx := r.Context()
	accounts, err := s.read.ListAccounts(ctx, bearer(&viewsv1.ListAccountsRequest{}, sess))
	if err != nil {
		return "ListAccounts", fmt.Errorf("ListAccounts: %w", err)
	}
	classes := map[string]string{}
	var gross []*viewsv1.Money
	for _, a := range accounts.Msg.GetAccounts() {
		vals, valErr := s.read.ListAccountValuations(ctx, bearer(&viewsv1.ListAccountValuationsRequest{AccountId: a.GetId(), From: date.String(), To: date.String()}, sess))
		if valErr != nil {
			return "ListAccountValuations", fmt.Errorf("ListAccountValuations: %w", valErr)
		}
		for _, v := range vals.Msg.GetValuations() {
			class, classErr := s.assetClass(r, sess, v.GetInstrument().GetIsin(), classes)
			if classErr != nil {
				return "GetInstrument", classErr
			}
			if inGroup(data.Dimension.Key, data.Key, v, class) {
				data.Holdings = append(data.Holdings, filteredHolding{AccountID: a.GetId(), Account: a.GetName(), Valuation: v})
				gross = append(gross, v.GetValueReporting())
			}
		}
		if cashErr := s.filteredCash(r, sess, a, date, data); cashErr != nil {
			return "ListCash", cashErr
		}
	}
	data.Total = sumMoney(gross)
	return "", nil
}

// assetClass returns the asset class of the instrument with ISIN isin, as
// the read API names it, remembering each in classes.
func (s *server) assetClass(r *http.Request, sess session, isin string, classes map[string]string) (string, error) {
	if c, ok := classes[isin]; ok {
		return c, nil
	}
	in, err := s.read.GetInstrument(r.Context(), bearer(&viewsv1.GetInstrumentRequest{Isin: isin}, sess))
	if err != nil {
		return "", fmt.Errorf("GetInstrument: %w", err)
	}
	classes[isin] = in.Msg.GetInstrument().GetAssetClass()
	return classes[isin], nil
}

// filteredCash adds account a's cash on date that falls in data's group:
// every balance for the cash asset class, the balance in the currency for
// a currency.
func (s *server) filteredCash(r *http.Request, sess session, a *viewsv1.Account, date kernel.Date, data *holdingsData) error {
	if data.Dimension.Key == dimAssetClass && data.Key != assetCash {
		return nil
	}
	cash, err := s.read.ListCash(r.Context(), bearer(&viewsv1.ListCashRequest{AccountId: a.GetId(), Date: date.String()}, sess))
	if err != nil {
		return fmt.Errorf("ListCash: %w", err)
	}
	for _, b := range cash.Msg.GetBalances() {
		if data.Dimension.Key == dimAssetClass || b.GetCurrency() == data.Key {
			data.Cash = append(data.Cash, filteredCash{AccountID: a.GetId(), Account: a.GetName(), Balance: b})
		}
	}
	return nil
}
