package connectrpc

import (
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// heldTo writes what each bound account holds (D-117): its go-live date,
// empty while not seeded, tax wrapper, payout setting (D-135), holdings and
// open lots.
func heldTo(accounts []ingest.HeldAccount) []*connectorv1.AccountHoldings {
	if len(accounts) == 0 {
		return nil
	}
	out := make([]*connectorv1.AccountHoldings, 0, len(accounts))
	for i := range accounts {
		a := &accounts[i]
		pb := &connectorv1.AccountHoldings{AccountId: string(a.Account), TaxWrapper: a.Wrapper}
		if !a.GoLive.IsZero() {
			pb.GoLive = a.GoLive.String()
		}
		pb.BondPayout, pb.BondPayoutSince = payoutTo(a.Payout.Payout), dateTo(a.Payout.Since)
		for _, e := range a.Payout.Earlier {
			pb.BondPayoutEarlier = append(pb.BondPayoutEarlier, &connectorv1.BondPayoutPeriod{
				BondPayout: payoutTo(e.Payout), Since: dateTo(e.Since),
			})
		}
		for j := range a.Holdings {
			h := &a.Holdings[j]
			pb.Holdings = append(pb.Holdings, &connectorv1.Holding{
				Instrument: instrumentTo(h.Instrument), Quantity: h.Quantity.String(),
			})
		}
		for j := range a.Lots {
			l := &a.Lots[j]
			pb.Lots = append(pb.Lots, &connectorv1.OpenLot{
				Instrument: instrumentTo(l.Instrument), Quantity: l.Quantity.String(), Opened: l.Opened.String(),
				Cost: moneyTo(l.Cost), Opening: l.Opening,
			})
		}
		out = append(out, pb)
	}
	return out
}

// dateTo writes a date as "YYYY-MM-DD", or empty for the zero date.
func dateTo(d kernel.Date) string {
	if d.IsZero() {
		return ""
	}
	return d.String()
}

// payoutTo writes a payout setting.
func payoutTo(p ingest.Payout) connectorv1.BondPayout {
	switch p {
	case ingest.PayoutCash:
		return connectorv1.BondPayout_BOND_PAYOUT_CASH
	case ingest.PayoutPaidOut:
		return connectorv1.BondPayout_BOND_PAYOUT_PAID_OUT
	}
	return connectorv1.BondPayout_BOND_PAYOUT_UNSPECIFIED
}

// bondTermsTo writes treasury bond series terms: rates as fractions, the
// fee as an exact amount with where it came from, and the attribution of
// terms read from published data (D-97).
func bondTermsTo(terms []ingest.BondTerms) []*connectorv1.BondTerms {
	if len(terms) == 0 {
		return nil
	}
	out := make([]*connectorv1.BondTerms, 0, len(terms))
	for i := range terms {
		t := &terms[i]
		out = append(out, &connectorv1.BondTerms{
			Instrument:         instrumentTo(t.Instrument),
			FirstPeriodRate:    t.Terms.FirstPeriodRate.String(),
			Margin:             t.Terms.Margin.String(),
			EarlyRedemptionFee: moneyTo(t.Terms.EarlyRedemptionFee),
			Attribution:        attributionTo(t.Terms.Attribution),
			FeeSource:          feeSourceTo(t.Terms.FeeSource),
		})
	}
	return out
}

// feeSourceTo writes where a series' early-redemption fee came from
// (D-97); an invalid source is written unspecified, which reads as
// entered.
func feeSourceTo(s pl.FeeSource) connectorv1.BondFeeSource {
	switch s {
	case pl.FeeEntered:
		return connectorv1.BondFeeSource_BOND_FEE_SOURCE_ENTERED
	case pl.FeeFromTable:
		return connectorv1.BondFeeSource_BOND_FEE_SOURCE_TABLE
	case pl.FeeAssumed:
		return connectorv1.BondFeeSource_BOND_FEE_SOURCE_ASSUMED
	}
	return connectorv1.BondFeeSource_BOND_FEE_SOURCE_UNSPECIFIED
}

// attributionTo writes the attribution of terms read from published data,
// the time it was taken in RFC 3339 in UTC; nil for none.
func attributionTo(a *pl.Attribution) *connectorv1.BondTermsAttribution {
	if a == nil {
		return nil
	}
	return &connectorv1.BondTermsAttribution{
		Publisher: a.Publisher, Dataset: a.Dataset, License: a.License, Url: a.URL,
		Made: a.Made.String(), Fetched: a.Fetched.UTC().Format(time.RFC3339), Processing: a.Processing,
	}
}

// instrumentTo writes one instrument.
func instrumentTo(in ledger.Instrument) *connectorv1.Instrument {
	return &connectorv1.Instrument{Isin: in.ISIN.String(), Symbol: in.Symbol}
}

// moneyTo writes m as an exact amount and its currency.
func moneyTo(m kernel.Money) *connectorv1.Money {
	return &connectorv1.Money{Amount: m.Amount().String(), Currency: m.Currency().String()}
}
