package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	"github.com/arkadiuss/openfief/internal/views"
)

// ErrUnknownAccount reports a query for an account that is not stored.
var ErrUnknownAccount = errors.New("unknown account")

// cashRounding is how cash converted to the reporting currency for the
// allocation is rounded to its minor unit: as the valuation context rounds
// holding values, so cash and holdings add up alike.
const cashRounding = valuation.ValueRounding

// UnsupportedReportingCurrencyError reports a reporting currency (D-62) the
// views cannot convert to. FX rates are quoted against PLN only (D-20), and
// cross rates for any other reporting currency are undecided (T-12), so PLN is
// the only one supported.
type UnsupportedReportingCurrencyError struct {
	Currency kernel.Currency
}

// Error names the currency.
func (e *UnsupportedReportingCurrencyError) Error() string {
	return fmt.Sprintf("reporting currency %s is not supported: only %s", e.Currency, valuation.PLN())
}

// Views runs the read API's queries.
type Views struct {
	ledger      Ledger
	valuations  Valuations
	statuses    ValuationStatuses
	rates       FXRates
	reporting   kernel.Currency
	models      ModelReports
	plugins     Plugins
	priceChecks PriceChecks
	connections Connections
	pack        jurisdiction.Pack
	terms       Terms
}

// New returns the queries over the given ports, with totals, valuations and
// the allocation in the reporting currency reporting (D-62), configured by
// opts. A reporting currency other than PLN is an
// *UnsupportedReportingCurrencyError.
func New(l Ledger, v Valuations, s ValuationStatuses, r FXRates, reporting kernel.Currency, opts ...Option) (*Views, error) {
	if !reporting.Equal(valuation.PLN()) {
		return nil, &UnsupportedReportingCurrencyError{Currency: reporting}
	}
	out := &Views{ledger: l, valuations: v, statuses: s, rates: r, reporting: reporting}
	for _, o := range opts {
		o(out)
	}
	return out, nil
}

// ReportingCurrency returns the currency totals, valuations and the
// allocation are reported in.
func (v *Views) ReportingCurrency() kernel.Currency { return v.reporting }

// Sleeves returns every sleeve, in ID order.
func (v *Views) Sleeves(ctx context.Context) ([]portfolio.Sleeve, error) {
	s, err := v.ledger.Sleeves(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading sleeves: %w", err)
	}
	return s, nil
}

// Accounts returns every account, in ID order.
func (v *Views) Accounts(ctx context.Context) ([]ledger.Account, error) {
	a, err := v.ledger.Accounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading accounts: %w", err)
	}
	return a, nil
}

// Account returns the account id, or an error wrapping ErrUnknownAccount.
func (v *Views) Account(ctx context.Context, id ledger.AccountID) (ledger.Account, error) {
	a, err := v.ledger.Account(ctx, id)
	if err != nil {
		return ledger.Account{}, accountErr("loading account", err)
	}
	return a, nil
}

// Transactions returns one page of the current versions of account's
// transactions, in date then key order, starting after the cursor after, or
// from the first when after is nil, each with its origin: the connector
// instance that brought it in and whether that instance is removed (D-58).
func (v *Views) Transactions(ctx context.Context, account ledger.AccountID, after *views.TransactionCursor, size int) (views.TransactionPage, error) {
	txs, err := v.ledger.CurrentTransactions(ctx, account)
	if err != nil {
		return views.TransactionPage{}, accountErr("loading transactions", err)
	}
	page, err := views.PageTransactions(txs, after, size)
	if err != nil {
		return views.TransactionPage{}, fmt.Errorf("paging transactions: %w", err)
	}
	page.Origins = v.installed(ctx).Origins(page.Transactions)
	return page, nil
}

// Holdings returns account's non-zero holdings at the end of on, each with
// the removed connector instances behind it (D-91).
func (v *Views) Holdings(ctx context.Context, account ledger.AccountID, on kernel.Date) (MarkedHoldings, error) {
	h, err := v.ledger.HoldingsOn(ctx, account, on)
	if err != nil {
		return MarkedHoldings{}, accountErr("loading holdings", err)
	}
	return v.markHoldings(ctx, account, on, h)
}

// Lots returns account's open lots at the end of on, each with the removed
// connector instances behind it (D-91).
func (v *Views) Lots(ctx context.Context, account ledger.AccountID, on kernel.Date) (MarkedLots, error) {
	l, err := v.ledger.LotsOn(ctx, account, on)
	if err != nil {
		return MarkedLots{}, accountErr("loading lots", err)
	}
	return v.markLots(ctx, account, on, l)
}

// Cash returns account's non-zero cash at the end of on, one balance per
// currency.
func (v *Views) Cash(ctx context.Context, account ledger.AccountID, on kernel.Date) ([]kernel.Money, error) {
	c, err := v.ledger.CashOn(ctx, account, on)
	if err != nil {
		return nil, accountErr("loading cash", err)
	}
	return c, nil
}

// AssertionChecks returns account's balance assertion checks, only those
// with status when status is not zero.
func (v *Views) AssertionChecks(ctx context.Context, account ledger.AccountID, status ledgerapp.AssertionStatus) ([]ledgerapp.AssertionCheck, error) {
	checks, err := v.ledger.AssertionChecks(ctx, account)
	if err != nil {
		return nil, accountErr("loading balance assertions", err)
	}
	if status == 0 {
		return checks, nil
	}
	out := checks[:0:0]
	for i := range checks {
		if checks[i].Status == status {
			out = append(out, checks[i])
		}
	}
	return out, nil
}

// accountErr wraps err from a query about one account, adding
// ErrUnknownAccount when the ledger does not store the account.
func accountErr(what string, err error) error {
	if errors.Is(err, ledgerapp.ErrUnknownAccount) {
		return fmt.Errorf("%s: %w: %w", what, ErrUnknownAccount, err)
	}
	return fmt.Errorf("%s: %w", what, err)
}

// AccountValuations is an account's valued holding valuations in a date
// range with their totals per date, cash included.
type AccountValuations struct {
	Valuations []valuation.HoldingValuation
	Totals     []views.DateTotal
}

// AccountValuations returns account's stored valuations dated from to to,
// both inclusive, each with its three values (D-77), and its totals per
// date with its cash (D-106).
func (v *Views) AccountValuations(ctx context.Context, account ledger.AccountID, from, to kernel.Date) (AccountValuations, error) {
	if err := views.ValidateRange(from, to); err != nil {
		return AccountValuations{}, fmt.Errorf("account valuations: %w", err)
	}
	acc, err := v.Account(ctx, account)
	if err != nil {
		return AccountValuations{}, err
	}
	stored, err := v.valuations.Valuations(ctx, account, from, to)
	if err != nil {
		return AccountValuations{}, fmt.Errorf("loading valuations: %w", err)
	}
	totals, err := v.totals(ctx, []ledger.Account{acc}, from, to)
	if err != nil {
		return AccountValuations{}, err
	}
	return AccountValuations{Valuations: valued(stored), Totals: totals}, nil
}

// ValuationTotals returns the whole portfolio's totals per date from to to,
// its three values (D-77) summed over every account's stored valuations
// and cash (D-106).
func (v *Views) ValuationTotals(ctx context.Context, from, to kernel.Date) ([]views.DateTotal, error) {
	if err := views.ValidateRange(from, to); err != nil {
		return nil, fmt.Errorf("valuation totals: %w", err)
	}
	accounts, err := v.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	return v.totals(ctx, accounts, from, to)
}

// ValuationStatus returns the date account is valued through and whether a
// revaluation is pending or running (D-36).
func (v *Views) ValuationStatus(ctx context.Context, account ledger.AccountID) (valuationapp.ValuationStatus, error) {
	if _, err := v.Account(ctx, account); err != nil {
		return valuationapp.ValuationStatus{}, err
	}
	s, err := v.statuses.Status(ctx, account)
	if err != nil {
		return valuationapp.ValuationStatus{}, fmt.Errorf("loading valuation status: %w", err)
	}
	s.Account = account
	return s, nil
}

// ValuationStatuses returns every account's valuation status, in account ID
// order.
func (v *Views) ValuationStatuses(ctx context.Context) ([]valuationapp.ValuationStatus, error) {
	accounts, err := v.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]valuationapp.ValuationStatus, 0, len(accounts))
	for i := range accounts {
		s, err := v.statuses.Status(ctx, accounts[i].ID)
		if err != nil {
			return nil, fmt.Errorf("loading valuation status: %w", err)
		}
		s.Account = accounts[i].ID
		out = append(out, s)
	}
	return out, nil
}

// valued returns the holding valuations in vals that lack no input. Those
// that do are left out: the valuation status says how far an account is
// valued.
func valued(vals []valuation.HoldingValuation) []valuation.HoldingValuation {
	out := make([]valuation.HoldingValuation, 0, len(vals))
	for i := range vals {
		if vals[i].Valued() {
			out = append(out, vals[i])
		}
	}
	return out
}

// AllocationOn is the portfolio's allocation of one value kind on a date,
// with the holdings and cash left out of it because they lack that value.
type AllocationOn struct {
	Date       kernel.Date
	Kind       views.ValueKind
	Allocation views.Allocation
	Unvalued   []views.Unvalued
}

// Allocation returns the portfolio's value of kind k (D-77) at the end of on
// in the reporting currency, grouped by d (D-30): every account's holdings
// at their stored valuation on that date, and its cash converted at the FX
// rate for that date and valued in the account's tax wrapper.
func (v *Views) Allocation(ctx context.Context, on kernel.Date, d views.Dimension, k views.ValueKind) (AllocationOn, error) {
	accounts, err := v.Accounts(ctx)
	if err != nil {
		return AllocationOn{}, err
	}
	settings, err := v.valueSettings(ctx)
	if err != nil {
		return AllocationOn{}, err
	}
	var items []views.AllocationItem
	out := AllocationOn{Date: on, Kind: k}
	for i := range accounts {
		got, unvalued, itemErr := v.allocationItems(ctx, &accounts[i], on, k, &settings)
		if itemErr != nil {
			return AllocationOn{}, itemErr
		}
		items = append(items, got...)
		out.Unvalued = append(out.Unvalued, unvalued...)
	}
	if out.Allocation, err = views.Allocate(items, d, v.reporting); err != nil {
		return AllocationOn{}, fmt.Errorf("allocating: %w", err)
	}
	return out, nil
}

// allocationItems returns a's holdings and cash at the end of on that have
// the value of kind k, and those that don't.
func (v *Views) allocationItems(ctx context.Context, a *ledger.Account, on kernel.Date, k views.ValueKind, settings *jurisdiction.Settings) ([]views.AllocationItem, []views.Unvalued, error) {
	holdings, err := v.ledger.HoldingsOn(ctx, a.ID, on)
	if err != nil {
		return nil, nil, fmt.Errorf("loading holdings: %w", err)
	}
	vals, err := v.valuations.Valuations(ctx, a.ID, on, on)
	if err != nil {
		return nil, nil, fmt.Errorf("loading valuations: %w", err)
	}
	items, unvalued, err := views.ValueHoldings(*a, holdings, vals, on, k)
	if err != nil {
		return nil, nil, fmt.Errorf("valuing holdings: %w", err)
	}
	cashItems, cashUnvalued, err := v.cashItems(ctx, a, on, k, settings)
	if err != nil {
		return nil, nil, err
	}
	return append(items, cashItems...), append(unvalued, cashUnvalued...), nil
}

// cashItems returns a's cash at the end of on at its value of kind k in
// a's tax wrapper, and the cash without that value.
func (v *Views) cashItems(ctx context.Context, a *ledger.Account, on kernel.Date, k views.ValueKind, settings *jurisdiction.Settings) ([]views.AllocationItem, []views.Unvalued, error) {
	cash, err := v.ledger.CashOn(ctx, a.ID, on)
	if err != nil {
		return nil, nil, fmt.Errorf("loading cash: %w", err)
	}
	terms, err := v.AccountTerms(ctx, a.ID)
	if err != nil {
		return nil, nil, err
	}
	var items []views.AllocationItem
	var unvalued []views.Unvalued
	for _, c := range cash {
		converted, err := v.cashInReporting(ctx, c, on)
		if err != nil {
			return nil, nil, fmt.Errorf("converting cash to %s: %w", v.reporting, err)
		}
		values, err := v.cashValues(converted, terms.Wrapper, settings)
		if err != nil {
			return nil, nil, err
		}
		value, ok, err := views.CashValue(&values, k)
		if err != nil {
			return nil, nil, fmt.Errorf("valuing cash: %w", err)
		}
		if !ok {
			unvalued = append(unvalued, views.Unvalued{Account: a.ID, Cash: c, Missing: valuation.MissingMarginalRate.String()})
			continue
		}
		items = append(items, views.AllocationItem{Account: *a, AssetClass: views.AssetCash, Value: c, Reporting: value})
	}
	return items, unvalued, nil
}

// cashInReporting converts cash c to the reporting currency at the FX rate
// for on. New admits only PLN as the reporting currency, so the stored rates
// to PLN convert; PLN cash needs no rate.
func (v *Views) cashInReporting(ctx context.Context, c kernel.Money, on kernel.Date) (kernel.Money, error) {
	var rates []valuation.FXRate
	if !c.Currency().Equal(v.reporting) {
		var err error
		if rates, err = v.rates.FXRates(ctx, c.Currency(), on); err != nil {
			return kernel.Money{}, fmt.Errorf("loading FX rates: %w", err)
		}
	}
	converted, err := valuation.ConvertToPLN(c, rates, on, cashRounding)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("converting cash: %w", err)
	}
	return converted, nil
}
