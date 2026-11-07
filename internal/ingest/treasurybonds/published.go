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

// CompleteTerms fills in the terms the owner left to the connector on each
// lot of e (D-97), reading the published data from src only when a lot
// needs its rates. A failure to read the data is returned as is. A lot
// whose every term was left to the connector and whose series the data
// doesn't list yet stays pending (BondLot.Pending): it is booked without
// terms, and a later sync fills them in (FillMissing). A series the data
// doesn't list whose fee the owner entered fails with an error wrapping
// pl.ErrNotPublished, and one sold before its type's first fee row with
// pl.ErrNoReviewedFee: the owner then enters that series' terms.
func CompleteTerms(ctx context.Context, src PublishedSource, e *ingest.BondLotEntries) error {
	var published *pl.Published
	for i := range e.Lots {
		l := &e.Lots[i]
		if l.PublishedRates && published == nil {
			p, err := src.Published(ctx)
			if err != nil {
				return fmt.Errorf("reading the published retail bond data: %w", err)
			}
			published = p
		}
		if err := Complete(l, published); err != nil {
			return fmt.Errorf("lot %d: %w", i+1, err)
		}
	}
	return nil
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

// MissingTerms returns the series of the treasury bonds the held accounts
// hold, as a holding or an open lot, whose terms the core stores none of,
// in series order and each once: lots entered before the ministry's data
// listed their series, which a sync fills in (D-97). An instrument is a
// bond of this connector when its symbol is a series code and its ISIN the
// one Instrument derives from it.
func MissingTerms(held []ingest.HeldAccount, stored []ingest.BondTerms) []pl.Series {
	have := map[ledger.ISIN]bool{}
	for i := range stored {
		have[stored[i].Instrument.ISIN] = true
	}
	seen := map[pl.Series]bool{}
	var out []pl.Series
	add := func(in ledger.Instrument) {
		s, err := pl.ParseSeries(in.Symbol)
		if err != nil || seen[s] || have[in.ISIN] {
			return
		}
		if derived, err := Instrument(s); err != nil || derived != in {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	for i := range held {
		for _, h := range held[i].Holdings {
			add(h.Instrument)
		}
		for j := range held[i].Lots {
			add(held[i].Lots[j].Instrument)
		}
	}
	slices.SortFunc(out, func(a, b pl.Series) int { return strings.Compare(a.String(), b.String()) })
	return out
}

// FillMissing returns the terms of each series of missing that the
// published data now lists, filled in as Complete fills a lot that left
// every term to the connector, reading the data from src only when missing
// isn't empty. A series the data still doesn't list, or one sold before its
// type's first fee row, is left out, so a later sync tries again; it never
// returns terms for a series not in missing, so terms the core stores, the
// owner's included, are never replaced. A failure to read the data is
// returned as is.
func FillMissing(ctx context.Context, src PublishedSource, missing []pl.Series) ([]ingest.BondTerms, error) {
	if len(missing) == 0 {
		return nil, nil
	}
	p, err := src.Published(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading the published retail bond data: %w", err)
	}
	var out []ingest.BondTerms
	for _, s := range missing {
		l := ingest.BondLot{Terms: pl.Terms{Series: s}, PublishedRates: true, ReviewedFee: true}
		if err := Complete(&l, p); errors.Is(err, pl.ErrNoReviewedFee) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("%s: %w", s, err)
		}
		if l.Pending() {
			continue
		}
		in, err := Instrument(s)
		if err != nil {
			return nil, err
		}
		out = append(out, ingest.BondTerms{Instrument: in, Terms: l.Terms})
	}
	return out, nil
}
