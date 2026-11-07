package treasurybonds

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// PublishedSource supplies the Ministry of Finance's published retail bond
// data (D-97); MinistryClient implements it.
type PublishedSource interface {
	Published(ctx context.Context) (*pl.Published, error)
}

// The processing CompleteTerms adds to the published data's attribution,
// naming where the early-redemption fee came from, since the data holds
// none.
const (
	tableFeeProcessing   = "; the early-redemption fee is not in the data: OpenFief took it from its reviewed table of the issue letters"
	assumedFeeProcessing = "; the early-redemption fee is not in the data: OpenFief assumed the latest fee of its reviewed table of the issue letters, for the owner to confirm"
	enteredFeeProcessing = "; the early-redemption fee is not in the data: the owner entered it"
)

// CompleteTerms reads the published data from src once, when e has lots,
// for the registry's ISIN of each lot's series (D-214) and the terms the
// owner left to the connector on each (D-97), fills those terms in, and
// returns the data read for Mapper.Map. With no lot, or data that can't be
// read while no lot needs its rates, it returns data that lists nothing:
// the lots are then booked under their provisional instrument, which a
// later sync re-keys; when a lot needs the rates, the failure is returned
// as is. A lot whose every term was
// left to the connector and whose series the data doesn't list yet stays
// pending (BondLot.Pending): it is booked without terms, and a later sync
// fills them in (FillMissing). A series the data doesn't list whose fee the
// owner entered fails with an error wrapping pl.ErrNotPublished, and one
// sold before its type's first fee row with pl.ErrNoReviewedFee: the owner
// then enters that series' terms.
func CompleteTerms(ctx context.Context, src PublishedSource, e *ingest.BondLotEntries) (*pl.Published, error) {
	if len(e.Lots) == 0 {
		return &pl.Published{}, nil
	}
	published, err := src.Published(ctx)
	if err != nil {
		if slices.ContainsFunc(e.Lots, func(l ingest.BondLot) bool { return l.PublishedRates }) {
			return nil, fmt.Errorf("reading the published retail bond data: %w", err)
		}
		published = &pl.Published{}
	}
	for i := range e.Lots {
		if err := Complete(&e.Lots[i], published); err != nil {
			return nil, fmt.Errorf("lot %d: %w", i+1, err)
		}
	}
	return published, nil
}

// Complete fills in the terms the owner left to the connector on l: the
// first-period rate and margin from p, which must not be nil then, and the
// early-redemption fee from the reviewed table (pl.EarlyRedemptionFee),
// with where it came from: the table's own fee, or, for a series sold
// after the table was reviewed through, its type's latest fee carried
// forward as assumed, for the owner to confirm (D-97). Terms with published
// rates carry p's attribution, whose processing also says where the fee
// came from; terms the owner entered in full carry none.
//
// A lot that left every term to the connector, of a series p doesn't list
// yet, is left as it is, pending, and Complete returns nil. A series p
// doesn't list whose fee the owner entered fails with an error wrapping
// pl.ErrNotPublished; one the fee table doesn't cover with
// pl.ErrNoReviewedFee.
func Complete(l *ingest.BondLot, p *pl.Published) error {
	s := l.Series()
	var rates pl.PublishedRates
	if l.PublishedRates {
		if p == nil {
			return fmt.Errorf("%w: %s: no data read", pl.ErrNotPublished, s)
		}
		r, err := p.RatesOf(s)
		switch {
		case errors.Is(err, pl.ErrNotPublished) && l.ReviewedFee:
			return nil
		case err != nil:
			return fmt.Errorf("first-period rate and margin: %w", err)
		}
		rates = r
	}
	feeNote := enteredFeeProcessing
	if l.ReviewedFee {
		fee, err := pl.EarlyRedemptionFee(s)
		if err != nil {
			return fmt.Errorf("early-redemption fee: %w", err)
		}
		l.Terms.EarlyRedemptionFee, l.Terms.FeeSource, l.ReviewedFee = fee.Fee, fee.Source, false
		feeNote = tableFeeProcessing
		if fee.Source == pl.FeeAssumed {
			feeNote = assumedFeeProcessing
		}
	}
	if !l.PublishedRates {
		return nil
	}
	att := p.Attribution
	att.Processing += feeNote
	l.Terms.FirstPeriodRate, l.Terms.Margin, l.Terms.Attribution = rates.FirstPeriodRate, rates.Margin, &att
	l.PublishedRates = false
	return nil
}

// MissingTerms returns the treasury bond instruments the held accounts
// hold, as a holding or an open lot, whose terms the core stores none of,
// in symbol order and each once: lots entered before the ministry's data
// listed their series, which a sync fills in (D-97). An instrument is a
// bond when its symbol is a series code and its ISIN a Polish one, the
// registry's or a provisional one (D-214).
func MissingTerms(held []ingest.HeldAccount, stored []ingest.BondTerms) []ledger.Instrument {
	have := map[ledger.ISIN]bool{}
	for i := range stored {
		have[stored[i].Instrument.ISIN] = true
	}
	seen := map[ledger.Instrument]bool{}
	var out []ledger.Instrument
	add := func(in ledger.Instrument) {
		if _, err := pl.ParseSeries(in.Symbol); err != nil || seen[in] || have[in.ISIN] || pl.CheckBondISIN(in.ISIN.String()) != nil {
			return
		}
		seen[in] = true
		out = append(out, in)
	}
	for i := range held {
		for _, h := range held[i].Holdings {
			add(h.Instrument)
		}
		for j := range held[i].Lots {
			add(held[i].Lots[j].Instrument)
		}
	}
	slices.SortFunc(out, func(a, b ledger.Instrument) int { return strings.Compare(a.String(), b.String()) })
	return out
}

// SyncPublished returns what a sync delivers from the ministry's published
// data, read from src once and only when missing or provisional isn't
// empty: the terms FillMissing fills in for missing (D-97) and the registry
// ISINs RegistryISINs finds for provisional (D-214). A failure to read the
// data is returned as is.
func SyncPublished(ctx context.Context, src PublishedSource, missing, provisional []ledger.Instrument) ([]ingest.BondTerms, []ingest.RegistryISIN, error) {
	if len(missing) == 0 && len(provisional) == 0 {
		return nil, nil, nil
	}
	p, err := src.Published(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("reading the published retail bond data: %w", err)
	}
	terms, err := FillMissing(p, missing)
	if err != nil {
		return nil, nil, err
	}
	return terms, RegistryISINs(p, provisional), nil
}

// FillMissing returns the terms of each bond instrument of missing whose
// series the published data p now lists, filled in as Complete fills a lot
// that left every term to the connector. Terms of an instrument under its
// provisional ISIN are keyed by the series' instrument in p
// (SeriesInstrument): the registry's once p lists its ISIN, which the
// core's re-key then brings the holding to (RegistryISINs). A series p
// still doesn't list, or one sold before its type's first fee row, is left
// out, so a later sync tries again; it never returns terms for an
// instrument not in missing, so terms the core stores, the owner's
// included, are never replaced.
func FillMissing(p *pl.Published, missing []ledger.Instrument) ([]ingest.BondTerms, error) {
	var out []ingest.BondTerms
	for _, in := range missing {
		s, err := pl.ParseSeries(in.Symbol)
		if err != nil {
			continue
		}
		l := ingest.BondLot{Terms: pl.Terms{Series: s}, PublishedRates: true, ReviewedFee: true}
		if completeErr := Complete(&l, p); errors.Is(completeErr, pl.ErrNoReviewedFee) {
			continue
		} else if completeErr != nil {
			return nil, fmt.Errorf("%s: %w", s, completeErr)
		}
		if l.Pending() {
			continue
		}
		if pl.IsProvisional(in.ISIN.String(), in.Symbol) {
			if in, err = SeriesInstrument(p, s); err != nil {
				return nil, err
			}
		}
		out = append(out, ingest.BondTerms{Instrument: in, Terms: l.Terms})
	}
	return out, nil
}

// RegistryISINs returns the registry's ISIN of each instrument of
// provisional, a treasury bond series under its provisional ISIN, whose
// ISIN the published data p now lists (D-214), in the order given. One p
// lists none of yet is left out, so a later sync tries again; an
// instrument that isn't under its series' provisional ISIN is never
// re-keyed.
func RegistryISINs(p *pl.Published, provisional []ledger.Instrument) []ingest.RegistryISIN {
	var out []ingest.RegistryISIN
	for _, in := range provisional {
		s, ok := pl.ProvisionalSeries(in.ISIN.String())
		if !ok || s.String() != in.Symbol {
			continue
		}
		listed, err := SeriesInstrument(p, s)
		if err != nil || listed.ISIN == in.ISIN {
			continue
		}
		out = append(out, ingest.RegistryISIN{Instrument: in, ISIN: listed.ISIN})
	}
	return out
}
