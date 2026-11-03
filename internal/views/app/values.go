package app

import (
	"context"
	"errors"
	"fmt"
	"maps"

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
	// ExemptionClasses returns the owner's class of each instrument they
	// classified for a yearly tax (D-132), by ISIN.
	ExemptionClasses(ctx context.Context) (map[string]jurisdiction.ExemptionClass, error)
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
		converted, convErr := valuation.ConvertAt(b.Amount, v.reporting, rates, b.Date, cashRounding)
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
