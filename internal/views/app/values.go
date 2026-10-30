package app

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
)

// Terms reads what the three values depend on besides the record (D-76,
// D-78, D-113): each account's tax wrapper and fee rule, and the owner's
// settings. *valuation/postgres.Store implements it.
type Terms interface {
	// TaxWrapper returns account's recorded tax wrapper, and false when
	// none was recorded.
	TaxWrapper(ctx context.Context, account ledger.AccountID) (jurisdiction.Wrapper, bool, error)
	// FeeRule returns account's fee rule, the zero rule unless recorded.
	FeeRule(ctx context.Context, account ledger.AccountID) (jurisdiction.FeeRule, error)
	// ValueSettings returns the owner's settings the values need.
	ValueSettings(ctx context.Context) (jurisdiction.Settings, error)
	// NumberFormat returns the owner's number format, and false when the
	// owner has not chosen one.
	NumberFormat(ctx context.Context) (jurisdiction.NumberFormat, bool, error)
}

// WithJurisdiction values cash and computes yearly taxes with pack (D-70),
// reading the accounts' terms and the owner's settings from terms. Without
// it every account is in no wrapper, cash is worth its amount in all three
// values, and no yearly tax is computed.
func WithJurisdiction(pack jurisdiction.Pack, terms Terms) Option {
	return func(v *Views) { v.pack, v.terms = pack, terms }
}

// AccountTerms is an account's tax wrapper, the pack's default unless one
// was recorded, and its fee rule.
type AccountTerms struct {
	Wrapper jurisdiction.Wrapper
	Fee     jurisdiction.FeeRule
}

// AccountTerms returns account's tax wrapper and fee rule; both are zero
// without a jurisdiction (WithJurisdiction).
func (v *Views) AccountTerms(ctx context.Context, account ledger.AccountID) (AccountTerms, error) {
	if v.pack == nil {
		return AccountTerms{}, nil
	}
	w, recorded, err := v.terms.TaxWrapper(ctx, account)
	if err != nil {
		return AccountTerms{}, fmt.Errorf("loading the tax wrapper of %s: %w", account, err)
	}
	if !recorded {
		w = jurisdiction.DefaultWrapper(v.pack)
	}
	fee, err := v.terms.FeeRule(ctx, account)
	if err != nil {
		return AccountTerms{}, fmt.Errorf("loading the fee rule of %s: %w", account, err)
	}
	return AccountTerms{Wrapper: w, Fee: fee}, nil
}

// Settings are what every reader writes the record with: the reporting
// currency (D-62), the number format (D-113), whether it is the pack's
// default, and the tax wrappers an account can have (D-76).
type Settings struct {
	ReportingCurrency   kernel.Currency
	NumberFormat        jurisdiction.NumberFormat
	NumberFormatDefault bool
	Wrappers            []jurisdiction.Wrapper
}

// defaultNumberFormat is the number format without a jurisdiction: grouped
// with a space, a decimal point.
func defaultNumberFormat() jurisdiction.NumberFormat {
	return jurisdiction.NumberFormat{Group: jurisdiction.SeparatorSpace, Decimal: jurisdiction.SeparatorDot}
}

// Settings returns the settings: the owner's number format, or the pack's
// default until the owner chooses one.
func (v *Views) Settings(ctx context.Context) (Settings, error) {
	out := Settings{ReportingCurrency: v.reporting, NumberFormat: defaultNumberFormat(), NumberFormatDefault: true}
	if v.pack == nil {
		return out, nil
	}
	out.NumberFormat, out.Wrappers = v.pack.NumberFormat(), v.pack.Wrappers()
	f, chosen, err := v.terms.NumberFormat(ctx)
	if err != nil {
		return Settings{}, fmt.Errorf("loading the number format: %w", err)
	}
	if chosen {
		out.NumberFormat, out.NumberFormatDefault = f, false
	}
	return out, nil
}

// cashDays returns account a's cash at the end of each day of the range
// from..to with its three values in the reporting currency, in a's tax
// wrapper under settings: converted at the FX rate for the day, or
// unconverted without one.
func (v *Views) cashDays(ctx context.Context, a ledger.AccountID, from, to kernel.Date, settings *jurisdiction.Settings) ([]views.DayCash, error) {
	txs, err := v.ledger.CurrentTransactions(ctx, a)
	if err != nil {
		return nil, accountErr("loading transactions", err)
	}
	terms, err := v.AccountTerms(ctx, a)
	if err != nil {
		return nil, err
	}
	balances := views.CashOnDays(txs, from, to)
	rates, err := v.cashRates(ctx, txs, to)
	if err != nil {
		return nil, err
	}
	out := make([]views.DayCash, 0, len(balances))
	for _, b := range balances {
		day := views.DayCash{Date: b.Date, Amount: b.Amount}
		converted, convErr := valuation.ConvertToPLN(b.Amount, rates, b.Date, cashRounding)
		if errors.Is(convErr, valuation.ErrNoRate) {
			day.Unconverted = true
			out = append(out, day)
			continue
		}
		if convErr != nil {
			return nil, fmt.Errorf("converting cash: %w", convErr)
		}
		if day.Values, err = v.cashValues(converted, terms.Wrapper, settings); err != nil {
			return nil, err
		}
		out = append(out, day)
	}
	return out, nil
}

// cashValues returns the three values of converted cash in wrapper; all
// three are the amount without a jurisdiction.
func (v *Views) cashValues(converted kernel.Money, wrapper jurisdiction.Wrapper, settings *jurisdiction.Settings) (jurisdiction.Values, error) {
	if v.pack == nil {
		return jurisdiction.Values{Gross: converted, Net: converted, Liquidation: converted}, nil
	}
	out, err := v.pack.ValueCash(&jurisdiction.Cash{Wrapper: wrapper, Amount: converted}, settings)
	if err != nil {
		return jurisdiction.Values{}, fmt.Errorf("valuing cash: %w", err)
	}
	return out, nil
}

// valueSettings returns the owner's settings the values need; none without
// a jurisdiction.
func (v *Views) valueSettings(ctx context.Context) (jurisdiction.Settings, error) {
	if v.pack == nil {
		return jurisdiction.Settings{}, nil
	}
	s, err := v.terms.ValueSettings(ctx)
	if err != nil {
		return jurisdiction.Settings{}, fmt.Errorf("loading the owner's settings: %w", err)
	}
	return s, nil
}

// YearlyTaxes returns, for year, the yearly tax of every tax wrapper the
// pack taxes on its balance (D-76), such as OKI's: under a law in force
// that year, from the daily gross totals, cash included, of every account
// in the wrapper, summed per day from the first day of the year through
// the last day every account is valued for and the year's last day; a law
// not in force gives the tax with InForce false. None without a
// jurisdiction.
func (v *Views) YearlyTaxes(ctx context.Context, year int) ([]jurisdiction.YearlyTax, error) {
	if v.pack == nil {
		return nil, nil
	}
	var out []jurisdiction.YearlyTax
	for _, w := range v.pack.Wrappers() {
		t, err := v.pack.YearlyTax(w, year, nil)
		if errors.Is(err, jurisdiction.ErrNoYearlyTax) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("yearly tax of %s: %w", w, err)
		}
		if t.InForce {
			days, dayErr := v.wrapperBalances(ctx, w, year)
			if dayErr != nil {
				return nil, dayErr
			}
			if t, err = v.pack.YearlyTax(w, year, days); err != nil {
				return nil, fmt.Errorf("yearly tax of %s: %w", w, err)
			}
		}
		out = append(out, t)
	}
	return out, nil
}

// wrapperBalances returns the gross totals, cash included, of every account
// in wrapper w, summed per day of year up to the earliest date any account
// is valued through.
func (v *Views) wrapperBalances(ctx context.Context, w jurisdiction.Wrapper, year int) ([]jurisdiction.DayBalance, error) {
	from, err1 := kernel.NewDate(year, time.January, 1)
	next, err2 := kernel.NewDate(year+1, time.January, 1)
	if err := errors.Join(err1, err2); err != nil {
		return nil, fmt.Errorf("year %d: %w", year, err)
	}
	to := next.AddDays(-1)
	in, err := v.accountsIn(ctx, w)
	if err != nil {
		return nil, err
	}
	rec, err := v.recorded(ctx, in)
	if err != nil {
		return nil, err
	}
	if len(rec.txs) == 0 || rec.through.Before(from) {
		return nil, nil
	}
	totals, err := v.totals(ctx, in, from, earlier(to, rec.through))
	if err != nil {
		return nil, err
	}
	out := make([]jurisdiction.DayBalance, 0, len(totals))
	for i := range totals {
		out = append(out, jurisdiction.DayBalance{Date: totals[i].Date, Balance: totals[i].Reporting})
	}
	return out, nil
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
// stored valuations and cash.
func (v *Views) totals(ctx context.Context, accounts []ledger.Account, from, to kernel.Date) ([]views.DateTotal, error) {
	settings, err := v.valueSettings(ctx)
	if err != nil {
		return nil, err
	}
	var vals []valuation.HoldingValuation
	var cash []views.DayCash
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
	}
	totals, err := views.TotalsByDate(vals, cash, v.reporting)
	if err != nil {
		return nil, fmt.Errorf("totalling valuations: %w", err)
	}
	return totals, nil
}

// MarkedHoldings are an account's holdings on a date with the removed
// connector instances behind each (D-91), by instrument; an instrument
// with none is not in Removed.
type MarkedHoldings struct {
	Holdings []ledger.Holding
	Removed  map[ledger.Instrument][]string
}

// MarkedLots are an account's open lots on a date with the removed
// connector instances behind each (D-91), by the lot's origin; a lot with
// none is not in Removed.
type MarkedLots struct {
	Lots    []ledger.Lot
	Removed map[ledger.TransactionKey][]string
}

// markHoldings works out, when read (D-91), which of account's holdings at
// the end of on have a transaction from a removed instance behind them.
func (v *Views) markHoldings(ctx context.Context, account ledger.AccountID, on kernel.Date, holdings []ledger.Holding) (MarkedHoldings, error) {
	txs, err := v.ledger.CurrentTransactions(ctx, account)
	if err != nil {
		return MarkedHoldings{}, accountErr("loading transactions", err)
	}
	return MarkedHoldings{Holdings: holdings, Removed: views.RemovedBehindHoldings(txs, v.installed(ctx), on)}, nil
}

// markLots works out, when read (D-91), which of account's open lots at the
// end of on have a transaction from a removed instance behind them.
func (v *Views) markLots(ctx context.Context, account ledger.AccountID, on kernel.Date, lots []ledger.Lot) (MarkedLots, error) {
	txs, err := v.ledger.CurrentTransactions(ctx, account)
	if err != nil {
		return MarkedLots{}, accountErr("loading transactions", err)
	}
	installed := v.installed(ctx)
	out := MarkedLots{Lots: lots, Removed: map[ledger.TransactionKey][]string{}}
	seen := map[ledger.Instrument]bool{}
	for i := range lots {
		in := lots[i].Instrument
		if seen[in] {
			continue
		}
		seen[in] = true
		removed, lotErr := views.RemovedBehindLots(txs, in, installed, on)
		if lotErr != nil {
			return MarkedLots{}, fmt.Errorf("marking lots: %w", lotErr)
		}
		maps.Copy(out.Removed, removed)
	}
	return out, nil
}
