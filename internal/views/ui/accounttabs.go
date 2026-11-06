package ui

import (
	"errors"
	"net/http"
)

// tabParam is the query parameter that picks an account page's tab.
const tabParam = "tab"

// The account page's tabs (docs/dev/design-system.md, M0 screens): its
// holdings, transactions, open lots, cash, the balance assertions that
// reconcile it with its institution, and its settings.
const (
	tabHoldings       = "holdings"
	tabTransactions   = "transactions"
	tabLots           = "lots"
	tabCash           = "cash"
	tabReconciliation = "reconciliation"
	tabSettings       = "settings"
)

// readError is why a page couldn't be read: a call to the core, by its
// Method, or a malformed query, answered with Status.
type readError struct {
	Method string
	Status int
	Err    error
}

// Error names the method or the query's fault.
func (f *readError) Error() string {
	if f.Method != "" {
		return f.Method + ": " + f.Err.Error()
	}
	return f.Err.Error()
}

// Unwrap returns the failure's cause.
func (f *readError) Unwrap() error { return f.Err }

// readFailed answers a page whose reading failed with err: a malformed
// query with its status and reason, a failed call to the core as
// coreFailed does.
func (s *server) readFailed(w http.ResponseWriter, r *http.Request, err error) {
	f, ok := errors.AsType[*readError](err)
	switch {
	case ok && f.Status != 0:
		s.errorPage(w, r, f.Status, f.Err.Error())
	case ok:
		s.coreFailed(w, r, f.Method, f.Err)
	default:
		s.coreFailed(w, r, "page", err)
	}
}

// accountTab is one tab of an account's page.
type accountTab struct {
	Key, Label string
}

// accountTabs are an account page's tabs, in order.
func accountTabs() []accountTab {
	return []accountTab{
		{tabHoldings, "Holdings"},
		{tabTransactions, "Transactions"},
		{tabLots, "Lots"},
		{tabCash, "Cash"},
		{tabReconciliation, "Reconciliation"},
		{tabSettings, "Account settings"},
	}
}

// parseAccountTab returns the tab a query value names, and the holdings for
// an empty or unknown one.
func parseAccountTab(key string) string {
	for _, t := range accountTabs() {
		if t.Key == key {
			return key
		}
	}
	return tabHoldings
}
