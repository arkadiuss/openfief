package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
)

// ExemptionClasses returns the owner's class of each instrument they
// classified for a yearly tax (D-132), by instrument.
func (s *Store) ExemptionClasses(ctx context.Context) (map[string]jurisdiction.ExemptionClass, error) {
	rows, err := s.pool.Query(ctx, `SELECT instrument, class FROM valuation.exemption_classes`)
	if err != nil {
		return nil, fmt.Errorf("loading exemption classes: %w", err)
	}
	out := map[string]jurisdiction.ExemptionClass{}
	var instrument, class string
	_, err = pgx.ForEachRow(rows, []any{&instrument, &class}, func() error {
		out[instrument] = jurisdiction.ExemptionClass(class)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading exemption classes: %w", err)
	}
	return out, nil
}

// SaveExemptionClass sets instrument's class, replacing any earlier one.
func (s *Store) SaveExemptionClass(ctx context.Context, instrument string, class jurisdiction.ExemptionClass) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO valuation.exemption_classes (instrument, class) VALUES ($1, $2)
		ON CONFLICT (instrument) DO UPDATE SET class = EXCLUDED.class, set_at = now()`,
		instrument, string(class))
	if err != nil {
		return fmt.Errorf("saving the exemption class of %s: %w", instrument, err)
	}
	return nil
}
