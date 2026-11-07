package treasurybonds

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
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
	enteredFeeProcessing = "; the early-redemption fee is not in the data: the owner entered it"
)

// CompleteTerms fills in the terms the owner left to the connector on each
// lot of e (D-97), reading the published data from src only when a lot
// needs its rates. A failure to read the data is returned as is; a series
// the data doesn't list, or the fee table doesn't cover, fails with an
// error wrapping pl.ErrNotPublished or pl.ErrNoReviewedFee: the owner then
// enters that series' terms.
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
// early-redemption fee from the reviewed table (pl.EarlyRedemptionFee).
// Terms with published rates carry p's attribution, whose processing also
// says where the fee came from; terms the owner entered in full carry none.
// A series p doesn't list fails with an error wrapping pl.ErrNotPublished,
// one the fee table doesn't cover with pl.ErrNoReviewedFee.
func Complete(l *ingest.BondLot, p *pl.Published) error {
	s := l.Series()
	feeNote := enteredFeeProcessing
	if l.ReviewedFee {
		fee, err := pl.EarlyRedemptionFee(s)
		if err != nil {
			return fmt.Errorf("early-redemption fee: %w", err)
		}
		l.Terms.EarlyRedemptionFee, l.ReviewedFee, feeNote = fee, false, tableFeeProcessing
	}
	if !l.PublishedRates {
		return nil
	}
	if p == nil {
		return fmt.Errorf("%w: %s: no data read", pl.ErrNotPublished, s)
	}
	r, err := p.RatesOf(s)
	if err != nil {
		return fmt.Errorf("first-period rate and margin: %w", err)
	}
	att := p.Attribution
	att.Processing += feeNote
	l.Terms.FirstPeriodRate, l.Terms.Margin, l.Terms.Attribution = r.FirstPeriodRate, r.Margin, &att
	l.PublishedRates = false
	return nil
}
