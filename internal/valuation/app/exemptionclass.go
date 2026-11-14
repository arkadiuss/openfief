package app

import (
	"context"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/ledger"
)

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
	parsed, err := ledger.ParseISIN(isin)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	err = auditedSetting(ctx, s.audit, origin, sessionToken, settingExemptionClass, func(ctx context.Context) error {
		return s.recorder.RecordExemptionClass(ctx, isin, class)
	}, auditInstrumentRef+parsed.String())
	if err != nil {
		return fmt.Errorf("setting the exemption class: %w", err)
	}
	return nil
}
