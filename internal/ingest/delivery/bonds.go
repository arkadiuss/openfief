package delivery

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// ErrMalformedLot reports a bond lot message whose fields don't parse: a
// series code, quantity, date, rate or fee that isn't what the contract
// writes, or a missing fee.
var ErrMalformedLot = errors.New("malformed bond lot")

// errNoFee reports a lot without its early-redemption fee.
var errNoFee = errors.New("early_redemption_fee is missing")

// BondLots reads the lots of an ImportBondLots request, in order. It
// checks only that each field parses; whether a lot can be booked is
// ingest.BondLot.Validate's to say.
func BondLots(lots []*connectorv1.BondLot) ([]ingest.BondLot, error) {
	out := make([]ingest.BondLot, 0, len(lots))
	for i, l := range lots {
		lot, err := bondLot(l)
		if err != nil {
			return nil, fmt.Errorf("%w: lots[%d]: %w", ErrMalformedLot, i, err)
		}
		out = append(out, lot)
	}
	return out, nil
}

// bondLot reads one lot.
func bondLot(l *connectorv1.BondLot) (ingest.BondLot, error) {
	series, err := pl.ParseSeries(l.GetSeries())
	if err != nil {
		return ingest.BondLot{}, fmt.Errorf("series: %w", err)
	}
	q, err := kernel.ParseQuantity(l.GetQuantity())
	if err != nil {
		return ingest.BondLot{}, fmt.Errorf("quantity: %w", err)
	}
	purchased, err := kernel.ParseDate(l.GetPurchaseDate())
	if err != nil {
		return ingest.BondLot{}, fmt.Errorf("purchase_date: %w", err)
	}
	first, err := kernel.ParseDecimal(l.GetFirstPeriodRate())
	if err != nil {
		return ingest.BondLot{}, fmt.Errorf("first_period_rate: %w", err)
	}
	margin, err := kernel.ParseDecimal(l.GetMargin())
	if err != nil {
		return ingest.BondLot{}, fmt.Errorf("margin: %w", err)
	}
	fee, err := feeFrom(l.GetEarlyRedemptionFee())
	if err != nil {
		return ingest.BondLot{}, err
	}
	return ingest.BondLot{
		Terms:    pl.Terms{Series: series, FirstPeriodRate: first, Margin: margin, EarlyRedemptionFee: fee},
		Quantity: q, Purchased: purchased,
	}, nil
}

// feeFrom reads an early-redemption fee; a missing one is an error.
func feeFrom(m *connectorv1.Money) (kernel.Money, error) {
	if m == nil {
		return kernel.Money{}, errNoFee
	}
	c, err := kernel.ParseCurrency(m.GetCurrency())
	if err != nil {
		return kernel.Money{}, fmt.Errorf("early_redemption_fee.currency: %w", err)
	}
	d, err := kernel.ParseDecimal(m.GetAmount())
	if err != nil {
		return kernel.Money{}, fmt.Errorf("early_redemption_fee.amount: %w", err)
	}
	return kernel.NewMoney(d, c), nil
}

// BondTerms writes treasury bond series terms, in order: rates as
// fractions, the fee as an exact PLN amount.
func BondTerms(terms []ingest.BondTerms) []*connectorv1.BondTerms {
	out := make([]*connectorv1.BondTerms, 0, len(terms))
	for i := range terms {
		t := &terms[i]
		out = append(out, &connectorv1.BondTerms{
			Instrument:         &connectorv1.Instrument{Isin: t.Instrument.ISIN.String(), Symbol: t.Instrument.Symbol},
			FirstPeriodRate:    t.Terms.FirstPeriodRate.String(),
			Margin:             t.Terms.Margin.String(),
			EarlyRedemptionFee: money(t.Terms.EarlyRedemptionFee),
		})
	}
	return out
}
