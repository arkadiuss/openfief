package delivery

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// ErrMalformedHoldings reports a sync request whose accounts or bond terms
// don't parse (D-117).
var ErrMalformedHoldings = errors.New("malformed holdings in the sync request")

// Held reads what a sync request says the instance's bound accounts hold
// (D-117): each account's go-live date, tax wrapper, payout setting
// (D-135), holdings and open lots, and the bond series terms the core stores for them. Anything
// malformed fails with an error wrapping ErrMalformedHoldings naming the
// field.
func Held(req *connectorv1.SyncRequest) ([]ingest.HeldAccount, []ingest.BondTerms, error) {
	accounts := make([]ingest.HeldAccount, 0, len(req.GetAccounts()))
	for i, a := range req.GetAccounts() {
		held, err := heldAccount(a)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: accounts[%d]: %w", ErrMalformedHoldings, i, err)
		}
		accounts = append(accounts, held)
	}
	terms := make([]ingest.BondTerms, 0, len(req.GetBondTerms()))
	for i, t := range req.GetBondTerms() {
		bt, err := bondTermsFrom(t)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: bond_terms[%d]: %w", ErrMalformedHoldings, i, err)
		}
		terms = append(terms, bt)
	}
	return accounts, terms, nil
}

// heldAccount reads one account's holdings.
func heldAccount(a *connectorv1.AccountHoldings) (ingest.HeldAccount, error) {
	if a.GetAccountId() == "" {
		return ingest.HeldAccount{}, fmt.Errorf("account_id: %w", ledger.ErrInvalidAccount)
	}
	out := ingest.HeldAccount{Account: ledger.AccountID(a.GetAccountId()), Wrapper: a.GetTaxWrapper()}
	if a.GetGoLive() != "" {
		d, err := kernel.ParseDate(a.GetGoLive())
		if err != nil {
			return ingest.HeldAccount{}, fmt.Errorf("go_live: %w", err)
		}
		out.GoLive = d
	}
	payout, err := payoutFrom(a)
	if err != nil {
		return ingest.HeldAccount{}, err
	}
	out.Payout = payout
	for i, h := range a.GetHoldings() {
		in, err := instrumentFrom(h.GetInstrument())
		if err != nil {
			return ingest.HeldAccount{}, fmt.Errorf("holdings[%d].instrument: %w", i, err)
		}
		q, err := kernel.ParseQuantity(h.GetQuantity())
		if err != nil {
			return ingest.HeldAccount{}, fmt.Errorf("holdings[%d].quantity: %w", i, err)
		}
		out.Holdings = append(out.Holdings, ledger.Holding{Account: out.Account, Instrument: in, Quantity: q})
	}
	for i, l := range a.GetLots() {
		lot, err := openLotFrom(l)
		if err != nil {
			return ingest.HeldAccount{}, fmt.Errorf("lots[%d]: %w", i, err)
		}
		out.Lots = append(out.Lots, lot)
	}
	return out, nil
}

// payoutFrom reads an account's payout setting with the settings in force
// before it (D-162): unspecified is ingest.PayoutCash, and the settings
// must have taken effect in order, the first one with no day.
func payoutFrom(a *connectorv1.AccountHoldings) (ingest.PayoutSetting, error) {
	current, err := payoutPeriodFrom(a.GetBondPayout(), a.GetBondPayoutSince())
	if err != nil {
		return ingest.PayoutSetting{}, err
	}
	out := ingest.PayoutSetting{Payout: current.Payout, Since: current.Since}
	for i, e := range a.GetBondPayoutEarlier() {
		p, err := payoutPeriodFrom(e.GetBondPayout(), e.GetSince())
		if err != nil {
			return ingest.PayoutSetting{}, fmt.Errorf("bond_payout_earlier[%d]: %w", i, err)
		}
		out.Earlier = append(out.Earlier, p)
	}
	if err := out.Validate(); err != nil {
		return ingest.PayoutSetting{}, fmt.Errorf("bond_payout_since: %w", err)
	}
	return out, nil
}

// payoutPeriodFrom reads one payout setting and the day it took effect,
// empty for none; unspecified is ingest.PayoutCash.
func payoutPeriodFrom(p connectorv1.BondPayout, since string) (ingest.PayoutPeriod, error) {
	var out ingest.PayoutPeriod
	switch p {
	case connectorv1.BondPayout_BOND_PAYOUT_UNSPECIFIED, connectorv1.BondPayout_BOND_PAYOUT_CASH:
		out.Payout = ingest.PayoutCash
	case connectorv1.BondPayout_BOND_PAYOUT_PAID_OUT:
		out.Payout = ingest.PayoutPaidOut
	default:
		return ingest.PayoutPeriod{}, fmt.Errorf("bond_payout: %w: %d", ingest.ErrUnknownPayout, int32(p))
	}
	if since != "" {
		d, err := kernel.ParseDate(since)
		if err != nil {
			return ingest.PayoutPeriod{}, fmt.Errorf("bond_payout_since: %w", err)
		}
		out.Since = d
	}
	return out, nil
}

// openLotFrom reads one open lot.
func openLotFrom(l *connectorv1.OpenLot) (ledger.Lot, error) {
	in, err := instrumentFrom(l.GetInstrument())
	if err != nil {
		return ledger.Lot{}, fmt.Errorf("instrument: %w", err)
	}
	q, err := kernel.ParseQuantity(l.GetQuantity())
	if err != nil {
		return ledger.Lot{}, fmt.Errorf("quantity: %w", err)
	}
	opened, err := kernel.ParseDate(l.GetOpened())
	if err != nil {
		return ledger.Lot{}, fmt.Errorf("opened: %w", err)
	}
	cost, err := moneyFrom(l.GetCost())
	if err != nil {
		return ledger.Lot{}, fmt.Errorf("cost: %w", err)
	}
	return ledger.Lot{Instrument: in, Opened: opened, Quantity: q, Cost: cost, Opening: l.GetOpening()}, nil
}

// bondTermsFrom reads one series' terms, with where its fee came from;
// the series is the instrument's symbol.
func bondTermsFrom(t *connectorv1.BondTerms) (ingest.BondTerms, error) {
	in, err := instrumentFrom(t.GetInstrument())
	if err != nil {
		return ingest.BondTerms{}, fmt.Errorf("instrument: %w", err)
	}
	s, err := pl.ParseSeries(in.Symbol)
	if err != nil {
		return ingest.BondTerms{}, fmt.Errorf("instrument.symbol: %w", err)
	}
	first, err := kernel.ParseDecimal(t.GetFirstPeriodRate())
	if err != nil {
		return ingest.BondTerms{}, fmt.Errorf("first_period_rate: %w", err)
	}
	margin, err := kernel.ParseDecimal(t.GetMargin())
	if err != nil {
		return ingest.BondTerms{}, fmt.Errorf("margin: %w", err)
	}
	fee, err := feeFrom(t.GetEarlyRedemptionFee())
	if err != nil {
		return ingest.BondTerms{}, err
	}
	src, err := FeeSourceFrom(t.GetFeeSource())
	if err != nil {
		return ingest.BondTerms{}, fmt.Errorf("fee_source: %w", err)
	}
	return ingest.BondTerms{
		Instrument: in, Terms: pl.Terms{Series: s, FirstPeriodRate: first, Margin: margin, EarlyRedemptionFee: fee, FeeSource: src},
	}, nil
}

// instrumentFrom reads an instrument: a valid ISIN and its symbol.
func instrumentFrom(in *connectorv1.Instrument) (ledger.Instrument, error) {
	isin, err := ledger.ParseISIN(in.GetIsin())
	if err != nil {
		return ledger.Instrument{}, fmt.Errorf("isin: %w", err)
	}
	return ledger.Instrument{ISIN: isin, Symbol: in.GetSymbol()}, nil
}

// moneyFrom reads an amount in a currency.
func moneyFrom(m *connectorv1.Money) (kernel.Money, error) {
	c, err := kernel.ParseCurrency(m.GetCurrency())
	if err != nil {
		return kernel.Money{}, fmt.Errorf("currency: %w", err)
	}
	d, err := kernel.ParseDecimal(m.GetAmount())
	if err != nil {
		return kernel.Money{}, fmt.Errorf("amount: %w", err)
	}
	return kernel.NewMoney(d, c), nil
}
