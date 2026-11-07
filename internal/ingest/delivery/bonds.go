package delivery

import (
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// ErrMalformedLot reports a bond lot message whose fields don't parse: a
// series code, quantity, date, rate or fee that isn't what the contract
// writes, or a first-period rate without its margin or the other way round.
var ErrMalformedLot = errors.New("malformed bond lot")

// errHalfRates reports a lot with only one of first_period_rate and margin.
var errHalfRates = errors.New("first_period_rate and margin must both be given or both be empty")

// errNoKey reports a booked transaction without its source or external ID.
var errNoKey = errors.New("source and external_id are required")

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
	purchase, err := purchaseFrom(l.GetPurchase())
	if err != nil {
		return ingest.BondLot{}, err
	}
	lot := ingest.BondLot{Terms: pl.Terms{Series: series}, Quantity: q, Purchased: purchased, Purchase: purchase}
	if err := lotTerms(l, &lot); err != nil {
		return ingest.BondLot{}, err
	}
	return lot, nil
}

// lotTerms reads the terms the owner entered with l into lot: empty rates
// leave them to the connector (PublishedRates), and so does an unset fee
// (ReviewedFee, D-97).
func lotTerms(l *connectorv1.BondLot, lot *ingest.BondLot) error {
	switch first, margin := l.GetFirstPeriodRate(), l.GetMargin(); {
	case first == "" && margin == "":
		lot.PublishedRates = true
	case first == "" || margin == "":
		return errHalfRates
	default:
		var err error
		if lot.Terms.FirstPeriodRate, err = kernel.ParseDecimal(first); err != nil {
			return fmt.Errorf("first_period_rate: %w", err)
		}
		if lot.Terms.Margin, err = kernel.ParseDecimal(margin); err != nil {
			return fmt.Errorf("margin: %w", err)
		}
	}
	if l.GetEarlyRedemptionFee() == nil {
		lot.ReviewedFee = true
		return nil
	}
	fee, err := feeFrom(l.GetEarlyRedemptionFee())
	if err != nil {
		return err
	}
	lot.Terms.EarlyRedemptionFee = fee
	return nil
}

// errBadPurchase reports a purchase the contract doesn't define.
var errBadPurchase = errors.New("purchase is not a known bond purchase")

// purchaseFrom reads a lot's purchase; unspecified is a cash purchase.
func purchaseFrom(p connectorv1.BondPurchase) (ingest.BondPurchase, error) {
	switch p {
	case connectorv1.BondPurchase_BOND_PURCHASE_UNSPECIFIED, connectorv1.BondPurchase_BOND_PURCHASE_CASH:
		return ingest.PurchaseCash, nil
	case connectorv1.BondPurchase_BOND_PURCHASE_EXCHANGE:
		return ingest.PurchaseExchange, nil
	}
	return 0, fmt.Errorf("%w: %d", errBadPurchase, p)
}

// BondLotEntries reads an ImportBondLots request: its account (D-118), its
// go-live date when it has one (empty for every account not seeded with
// opening lots, D-98), its lots in order (BondLots) and the lots it
// deletes. Anything malformed fails
// with an error wrapping ErrMalformedLot naming the field.
func BondLotEntries(req *connectorv1.ImportBondLotsRequest) (ingest.BondLotEntries, error) {
	var goLive kernel.Date
	if req.GetGoLive() != "" {
		d, err := kernel.ParseDate(req.GetGoLive())
		if err != nil {
			return ingest.BondLotEntries{}, fmt.Errorf("%w: go_live: %w", ErrMalformedLot, err)
		}
		goLive = d
	}
	lots, err := BondLots(req.GetLots())
	if err != nil {
		return ingest.BondLotEntries{}, err
	}
	out := ingest.BondLotEntries{Account: ledger.AccountID(req.GetAccountId()), GoLive: goLive, Lots: lots}
	for i, k := range req.GetDeleted() {
		series, err := pl.ParseSeries(k.GetSeries())
		if err != nil {
			return ingest.BondLotEntries{}, fmt.Errorf("%w: deleted[%d].series: %w", ErrMalformedLot, i, err)
		}
		purchased, err := kernel.ParseDate(k.GetPurchaseDate())
		if err != nil {
			return ingest.BondLotEntries{}, fmt.Errorf("%w: deleted[%d].purchase_date: %w", ErrMalformedLot, i, err)
		}
		out.Deleted = append(out.Deleted, ingest.BondLotKey{Series: series, Purchased: purchased})
	}
	booked, bookedErr := bookedFrom(out.Account, req.GetBooked())
	if bookedErr != nil {
		return ingest.BondLotEntries{}, bookedErr
	}
	out.Booked = booked
	return out, nil
}

// bookedFrom reads the transactions of account a request says the
// connector booked (D-154): each needs a key and a date.
func bookedFrom(account ledger.AccountID, booked []*connectorv1.BookedTransaction) ([]ledger.TransactionHeader, error) {
	out := make([]ledger.TransactionHeader, 0, len(booked))
	for i, b := range booked {
		k := b.GetKey()
		if k.GetSource() == "" || k.GetExternalId() == "" {
			return nil, fmt.Errorf("%w: booked[%d].key: %w", ErrMalformedLot, i, errNoKey)
		}
		day, err := kernel.ParseDate(b.GetDate())
		if err != nil {
			return nil, fmt.Errorf("%w: booked[%d].date: %w", ErrMalformedLot, i, err)
		}
		out = append(out, ledger.TransactionHeader{
			Key:     ledger.TransactionKey{Source: k.GetSource(), ExternalID: k.GetExternalId()},
			Account: account, Date: day,
		})
	}
	return out, nil
}

// feeFrom reads an early-redemption fee.
func feeFrom(m *connectorv1.Money) (kernel.Money, error) {
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
// fractions, the fee as an exact PLN amount, and the attribution of terms
// read from published data (D-97).
func BondTerms(terms []ingest.BondTerms) []*connectorv1.BondTerms {
	out := make([]*connectorv1.BondTerms, 0, len(terms))
	for i := range terms {
		t := &terms[i]
		out = append(out, &connectorv1.BondTerms{
			Instrument:         &connectorv1.Instrument{Isin: t.Instrument.ISIN.String(), Symbol: t.Instrument.Symbol},
			FirstPeriodRate:    t.Terms.FirstPeriodRate.String(),
			Margin:             t.Terms.Margin.String(),
			EarlyRedemptionFee: money(t.Terms.EarlyRedemptionFee),
			Attribution:        Attribution(t.Terms.Attribution),
		})
	}
	return out
}

// Attribution writes the attribution of terms read from published data,
// the time it was taken in RFC 3339 in UTC; nil for none.
func Attribution(a *pl.Attribution) *connectorv1.BondTermsAttribution {
	if a == nil {
		return nil
	}
	return &connectorv1.BondTermsAttribution{
		Publisher: a.Publisher, Dataset: a.Dataset, License: a.License, Url: a.URL,
		Made: a.Made.String(), Fetched: a.Fetched.UTC().Format(time.RFC3339), Processing: a.Processing,
	}
}
