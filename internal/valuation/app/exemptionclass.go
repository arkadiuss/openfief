package app

import (
	"context"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// UnclassifiedKind is the kind of the attention item that asks the owner to
// classify an instrument for a yearly tax's exemptions (D-132); its subject
// is the instrument's ISIN.
const UnclassifiedKind = "instrument-unclassified"

// AttentionRaiser raises and clears attention items for the owner (D-130),
// the shape of the attention store's port. Both are idempotent: raising an
// item that is open, or clearing one that isn't, changes nothing.
type AttentionRaiser interface {
	Raise(ctx context.Context, kind, subject, detail string) error
	Clear(ctx context.Context, kind, subject string) error
}

// WithAttention makes r raise an attention item through a for each
// instrument a revaluation finds that the pack can't classify and the owner
// hasn't, and clear it once the owner has (D-132), and returns r.
func (r *Revaluer) WithAttention(a AttentionRaiser) *Revaluer {
	r.attention = a
	return r
}

// raiseUnclassified raises an attention item for each instrument among
// rows, an account's valued holdings, that its wrapper's yearly tax needs
// a class for (Pack.ExemptionClasses), the pack can't classify and the owner
// hasn't classified, in ISIN order, and clears it for each such holding the
// owner has classified.
func (r *Revaluer) raiseUnclassified(ctx context.Context, account ledger.AccountID, rows []valuation.HoldingValuation) error {
	if r.attention == nil || len(rows) == 0 {
		return nil
	}
	w, recorded, err := r.store.TaxWrapper(ctx, account)
	if err != nil {
		return fmt.Errorf("loading the tax wrapper of %s: %w", account, err)
	}
	if !recorded {
		w = jurisdiction.DefaultWrapper(r.pack)
	}
	if len(r.pack.ExemptionClasses(w)) == 0 {
		return nil
	}
	owner, err := r.store.ExemptionClasses(ctx)
	if err != nil {
		return fmt.Errorf("loading exemption classes: %w", err)
	}
	return r.syncUnclassified(ctx, w, rows, owner)
}

// syncUnclassified raises the item for each of rows' instruments no one
// classified in w and clears it for those the owner classified.
func (r *Revaluer) syncUnclassified(ctx context.Context, w jurisdiction.Wrapper, rows []valuation.HoldingValuation, owner map[string]jurisdiction.ExemptionClass) error {
	for _, isin := range Unclassified(r.pack, w, rows, owner) {
		detail := "the jurisdiction pack can't tell this instrument's exemption class for " + string(w) + "'s yearly tax"
		if err := r.attention.Raise(ctx, UnclassifiedKind, isin, detail); err != nil {
			return fmt.Errorf("raising an attention item for %s: %w", isin, err)
		}
	}
	for _, isin := range Unclassified(r.pack, w, rows, nil) {
		if _, set := owner[isin]; !set {
			continue
		}
		if err := r.attention.Clear(ctx, UnclassifiedKind, isin); err != nil {
			return fmt.Errorf("clearing the attention item for %s: %w", isin, err)
		}
	}
	return nil
}

// Unclassified returns the ISINs of the holdings priced at market among
// rows that pack can't classify in wrapper w and that owner, the owner's
// classes by ISIN, doesn't hold, in ISIN order without repeats.
func Unclassified(pack jurisdiction.Pack, w jurisdiction.Wrapper, rows []valuation.HoldingValuation, owner map[string]jurisdiction.ExemptionClass) []string {
	var out []string
	for i := range rows {
		v := &rows[i]
		if v.Basis != valuation.Market || !v.Valued() {
			continue
		}
		isin := v.Instrument.ISIN.String()
		asset := jurisdiction.Asset{Kind: jurisdiction.AssetListed, Instrument: isin, Currency: v.Value.Currency()}
		if _, known := pack.ExemptionClass(w, &asset); known {
			continue
		}
		if _, set := owner[isin]; set || slices.Contains(out, isin) {
			continue
		}
		out = append(out, isin)
	}
	slices.Sort(out)
	return out
}

// RecordExemptionClass records the owner's class of the instrument with isin
// for a yearly tax (D-132). No stored value depends on it, so nothing is
// revalued. An ISIN that doesn't parse, or a class no wrapper of the pack
// offers (Pack.ExemptionClasses), is an error wrapping ErrInvalidInput.
func (r *InputRecorder) RecordExemptionClass(ctx context.Context, isin string, class jurisdiction.ExemptionClass) error {
	parsed, err := ledger.ParseISIN(isin)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if !slices.Contains(OfferedClasses(r.pack), class) {
		return fmt.Errorf("%w: exemption class %q is not one the pack offers", ErrInvalidInput, class)
	}
	if err := r.recorder.SaveExemptionClass(ctx, parsed.String(), class); err != nil {
		return fmt.Errorf("recording the exemption class of %s: %w", isin, err)
	}
	return nil
}

// OfferedClasses returns the exemption classes of every wrapper of pack, in
// wrapper then class order, without repeats.
func OfferedClasses(pack jurisdiction.Pack) []jurisdiction.ExemptionClass {
	var out []jurisdiction.ExemptionClass
	for _, w := range pack.Wrappers() {
		for _, c := range pack.ExemptionClasses(w) {
			if !slices.Contains(out, c) {
				out = append(out, c)
			}
		}
	}
	return out
}

// SetExemptionClass makes class the owner's class of the instrument with isin,
// for the owner (RecordExemptionClass).
func (s *Settings) SetExemptionClass(ctx context.Context, origin, sessionToken, isin string, class jurisdiction.ExemptionClass) error {
	if err := s.checkOwner(ctx, origin, sessionToken); err != nil {
		return err
	}
	if err := s.recorder.RecordExemptionClass(ctx, isin, class); err != nil {
		return fmt.Errorf("setting the exemption class: %w", err)
	}
	return nil
}
