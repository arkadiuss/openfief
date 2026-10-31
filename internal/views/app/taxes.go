package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
)

// YearlyTax is a wrapper's yearly tax (D-127) with the instruments it
// counted in the class the pack gives them without knowing it, because the
// owner hasn't classified them yet (D-132), in ISIN order.
type YearlyTax struct {
	jurisdiction.YearlyTax

	Unclassified []string
}

// YearlyTaxes returns, for year, the yearly tax of every tax wrapper the
// pack taxes on its accounts' value (D-127), such as OKI's: under a law in
// force that year, from each account's stored valuations and cash on the
// days of the year it was held, from its first transaction through the
// year's last day or the earliest date the wrapper's accounts are valued
// through, with its deposits and withdrawals; a law not in force gives the
// tax with InForce false. None without a jurisdiction.
func (v *Views) YearlyTaxes(ctx context.Context, year int) ([]YearlyTax, error) {
	if v.pack == nil {
		return nil, nil
	}
	var out []YearlyTax
	for _, w := range v.pack.Wrappers() {
		t, err := v.pack.YearlyTax(w, year, nil)
		if errors.Is(err, jurisdiction.ErrNoYearlyTax) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("yearly tax of %s: %w", w, err)
		}
		tax := YearlyTax{YearlyTax: t}
		if t.InForce {
			if tax, err = v.yearlyTax(ctx, w, year); err != nil {
				return nil, err
			}
		}
		out = append(out, tax)
	}
	return out, nil
}

// yearlyTax computes w's tax for year from its accounts' records.
func (v *Views) yearlyTax(ctx context.Context, w jurisdiction.Wrapper, year int) (YearlyTax, error) {
	first, err1 := kernel.NewDate(year, time.January, 1)
	next, err2 := kernel.NewDate(year+1, time.January, 1)
	if err := errors.Join(err1, err2); err != nil {
		return YearlyTax{}, fmt.Errorf("year %d: %w", year, err)
	}
	in, err := v.accountsIn(ctx, w)
	if err != nil {
		return YearlyTax{}, err
	}
	rec, err := v.recorded(ctx, in)
	if err != nil {
		return YearlyTax{}, err
	}
	classes, err := v.terms.ExemptionClasses(ctx)
	if err != nil {
		return YearlyTax{}, fmt.Errorf("loading exemption classes: %w", err)
	}
	to := earlier(next.AddDays(-1), rec.through)
	var accounts []jurisdiction.TaxAccount
	unclassified := map[string]bool{}
	for i := range in {
		a, missing, accErr := v.taxAccount(ctx, &in[i], w, first, to, classes)
		if accErr != nil {
			return YearlyTax{}, accErr
		}
		accounts = append(accounts, a)
		for _, isin := range missing {
			unclassified[isin] = true
		}
	}
	t, err := v.pack.YearlyTax(w, year, accounts)
	if err != nil {
		return YearlyTax{}, fmt.Errorf("yearly tax of %s: %w", w, err)
	}
	out := YearlyTax{YearlyTax: t}
	for isin := range unclassified {
		out.Unclassified = append(out.Unclassified, isin)
	}
	slices.Sort(out.Unclassified)
	return out, nil
}

// taxAccount returns one account's days in the year from first through to,
// from its first transaction on, and the instruments it holds that no one
// classified.
func (v *Views) taxAccount(ctx context.Context, acc *ledger.Account, w jurisdiction.Wrapper, first, to kernel.Date, classes map[string]jurisdiction.ExemptionClass) (jurisdiction.TaxAccount, []string, error) {
	txs, err := v.ledger.CurrentTransactions(ctx, acc.ID)
	if err != nil {
		return jurisdiction.TaxAccount{}, nil, accountErr("loading transactions", err)
	}
	if len(txs) == 0 {
		return jurisdiction.TaxAccount{Account: string(acc.ID)}, nil, nil
	}
	from := first
	if opened := firstDate(txs); opened.After(from) {
		from = opened
	}
	if to.Before(from) {
		return jurisdiction.TaxAccount{Account: string(acc.ID)}, nil, nil
	}
	vals, err := v.valuations.Valuations(ctx, acc.ID, from, to)
	if err != nil {
		return jurisdiction.TaxAccount{}, nil, fmt.Errorf("loading valuations: %w", err)
	}
	rates, err := v.taxRates(ctx, txs, vals, to)
	if err != nil {
		return jurisdiction.TaxAccount{}, nil, err
	}
	in := views.TaxYearAccount{
		Account: acc.ID, Wrapper: w, From: from, To: to, Vals: vals, Cash: views.CashOnDays(txs, from, to),
		Flows: views.ExternalFlows(txs), Rates: rates, Classes: classes,
	}
	a, missing, err := views.TaxDays(&in, v.pack)
	if err != nil {
		return jurisdiction.TaxAccount{}, nil, fmt.Errorf("yearly tax of %s: %w", acc.ID, err)
	}
	return a, missing, nil
}

// taxRates returns the FX rates of the currencies of txs' cash and of
// vals' prices through to.
func (v *Views) taxRates(ctx context.Context, txs []ledger.Transaction, vals []valuation.HoldingValuation, to kernel.Date) ([]valuation.FXRate, error) {
	rates, err := v.cashRates(ctx, txs, to)
	if err != nil {
		return nil, err
	}
	seen := map[kernel.Currency]bool{v.reporting: true}
	for i := range rates {
		seen[rates[i].Currency] = true
	}
	for i := range vals {
		c := vals[i].Value.Currency()
		if !vals[i].Valued() || seen[c] {
			continue
		}
		seen[c] = true
		more, err := v.rates.FXRates(ctx, c, to)
		if err != nil {
			return nil, fmt.Errorf("loading FX rates: %w", err)
		}
		rates = append(rates, more...)
	}
	return rates, nil
}

// accountsIn returns the accounts in wrapper w.
func (v *Views) accountsIn(ctx context.Context, w jurisdiction.Wrapper) ([]ledger.Account, error) {
	accounts, err := v.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	var in []ledger.Account
	for i := range accounts {
		terms, termsErr := v.AccountTerms(ctx, accounts[i].ID)
		if termsErr != nil {
			return nil, termsErr
		}
		if terms.Wrapper == w {
			in = append(in, accounts[i])
		}
	}
	return in, nil
}

// totals returns the totals per date of the range from..to of accounts'
// stored valuations and cash, their net and liquidation values counted as
// the jurisdiction's wrappers count tax on a total (views.TaxTotals,
// D-128); without a jurisdiction, the sums of the rows.
func (v *Views) totals(ctx context.Context, accounts []ledger.Account, from, to kernel.Date) ([]views.DateTotal, error) {
	settings, err := v.valueSettings(ctx)
	if err != nil {
		return nil, err
	}
	var vals []valuation.HoldingValuation
	var cash []views.DayCash
	taxed := make([]views.TaxedAccount, 0, len(accounts))
	for i := range accounts {
		stored, loadErr := v.valuations.Valuations(ctx, accounts[i].ID, from, to)
		if loadErr != nil {
			return nil, fmt.Errorf("loading valuations: %w", loadErr)
		}
		vals = append(vals, stored...)
		days, cashErr := v.cashDays(ctx, accounts[i].ID, from, to, &settings)
		if cashErr != nil {
			return nil, cashErr
		}
		cash = append(cash, days...)
		if v.pack != nil {
			a, taxErr := v.taxedAccount(ctx, accounts[i].ID, to, stored, days)
			if taxErr != nil {
				return nil, taxErr
			}
			taxed = append(taxed, a)
		}
	}
	totals, err := views.TotalsByDate(vals, cash, v.reporting)
	if err != nil {
		return nil, fmt.Errorf("totalling valuations: %w", err)
	}
	if v.pack == nil {
		return totals, nil
	}
	if err := views.TaxTotals(totals, taxed, v.pack, &settings); err != nil {
		return nil, fmt.Errorf("taxing totals: %w", err)
	}
	return totals, nil
}

// taxedAccount returns account's rows with what its total's tax needs: its
// wrapper, its sales' gains and, for a wrapper that taxes what is
// withdrawn above what was paid in, its payments.
func (v *Views) taxedAccount(ctx context.Context, account ledger.AccountID, to kernel.Date, stored []valuation.HoldingValuation, cash []views.DayCash) (views.TaxedAccount, error) {
	terms, err := v.AccountTerms(ctx, account)
	if err != nil {
		return views.TaxedAccount{}, err
	}
	txs, err := v.ledger.CurrentTransactions(ctx, account)
	if err != nil {
		return views.TaxedAccount{}, accountErr("loading transactions", err)
	}
	rates, err := v.cashRates(ctx, txs, to)
	if err != nil {
		return views.TaxedAccount{}, err
	}
	sales, err := ledger.Sales(txs)
	if err != nil {
		return views.TaxedAccount{}, fmt.Errorf("sales of %s: %w", account, err)
	}
	acct := valuation.AccountTerms{Pack: v.pack, Wrapper: terms.Wrapper, Fee: terms.Fee}
	gains, missing, err := acct.RealizedGains(sales, rates)
	if err != nil {
		return views.TaxedAccount{}, fmt.Errorf("realized gains of %s: %w", account, err)
	}
	out := views.TaxedAccount{Wrapper: terms.Wrapper, Vals: stored, Cash: cash, Realized: gains, RealizedMissing: missing != 0}
	paid, paidMissing, err := views.PaidIn(txs, rates, v.pack)
	if err != nil {
		return views.TaxedAccount{}, fmt.Errorf("payments into %s: %w", account, err)
	}
	out.PaidIn, out.PaidInMissing = paid, paidMissing
	return out, nil
}
