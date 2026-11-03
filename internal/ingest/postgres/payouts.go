package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrCorruptPayout reports a stored payout setting the ingest context can't
// read back.
var ErrCorruptPayout = errors.New("corrupt payout setting")

var _ app.PayoutSettings = (*Store)(nil)

// storedPeriod is one earlier payout setting as payout_earlier holds it
// (D-162): its key and the day it took effect, empty for the first.
type storedPeriod struct {
	Payout string `json:"payout"`
	Since  string `json:"since"`
}

// SavePayoutSetting stores account's payout setting (D-135) with the
// settings in force before it (D-162), replacing the stored one.
func (s *Store) SavePayoutSetting(ctx context.Context, account ledger.AccountID, p ingest.PayoutSetting) error {
	since := ""
	if !p.Since.IsZero() {
		since = p.Since.String()
	}
	earlier, err := encodeEarlier(&p)
	if err != nil {
		return fmt.Errorf("saving the payout setting of %s: %w", account, err)
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO ingest.payout_settings (account_id, payout, payout_since, payout_earlier)
		VALUES ($1, $2, nullif($3, '')::date, $4::jsonb)
		ON CONFLICT (account_id) DO UPDATE SET payout = EXCLUDED.payout, payout_since = EXCLUDED.payout_since,
		    payout_earlier = EXCLUDED.payout_earlier`,
		string(account), p.Payout.String(), since, string(earlier))
	if err != nil {
		return fmt.Errorf("saving the payout setting of %s: %w", account, err)
	}
	return nil
}

// PayoutSetting returns account's stored payout setting, or the default,
// as cash, when none is stored.
func (s *Store) PayoutSetting(ctx context.Context, account ledger.AccountID) (ingest.PayoutSetting, error) {
	var payout, since, earlier string
	err := s.pool.QueryRow(ctx, `
		SELECT payout, coalesce(payout_since::text, ''), payout_earlier::text
		FROM ingest.payout_settings WHERE account_id = $1`, string(account)).Scan(&payout, &since, &earlier)
	if errors.Is(err, pgx.ErrNoRows) {
		return ingest.PayoutSetting{}, nil
	}
	if err != nil {
		return ingest.PayoutSetting{}, fmt.Errorf("loading the payout setting of %s: %w", account, err)
	}
	out, err := decodeSetting(payout, since, earlier)
	if err != nil {
		return ingest.PayoutSetting{}, fmt.Errorf("the payout setting of %s: %w", account, err)
	}
	return out, nil
}

// decodeSetting reads one payout_settings row.
func decodeSetting(payout, since, earlier string) (ingest.PayoutSetting, error) {
	p, err := ingest.ParsePayout(payout)
	if err != nil {
		return ingest.PayoutSetting{}, fmt.Errorf("%w: %w", ErrCorruptPayout, err)
	}
	out := ingest.PayoutSetting{Payout: p}
	if since != "" {
		if out.Since, err = kernel.ParseDate(since); err != nil {
			return ingest.PayoutSetting{}, fmt.Errorf("%w: %w", ErrCorruptPayout, err)
		}
	}
	if out.Earlier, err = decodeEarlier([]byte(earlier)); err != nil {
		return ingest.PayoutSetting{}, err
	}
	return out, nil
}

// encodeEarlier returns the payout_earlier JSON of s's earlier settings.
func encodeEarlier(s *ingest.PayoutSetting) ([]byte, error) {
	out := make([]storedPeriod, len(s.Earlier))
	for i, p := range s.Earlier {
		out[i].Payout = p.Payout.String()
		if !p.Since.IsZero() {
			out[i].Since = p.Since.String()
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("encoding earlier payout settings: %w", err)
	}
	return b, nil
}

// decodeEarlier reads payout_earlier, nil for none.
func decodeEarlier(b []byte) ([]ingest.PayoutPeriod, error) {
	var stored []storedPeriod
	if err := json.Unmarshal(b, &stored); err != nil {
		return nil, fmt.Errorf("%w: earlier payout settings: %w", ErrCorruptPayout, err)
	}
	if len(stored) == 0 {
		return nil, nil
	}
	out := make([]ingest.PayoutPeriod, len(stored))
	for i, p := range stored {
		payout, err := ingest.ParsePayout(p.Payout)
		if err != nil {
			return nil, fmt.Errorf("%w: earlier payout settings: %w", ErrCorruptPayout, err)
		}
		var since kernel.Date
		if p.Since != "" {
			if since, err = kernel.ParseDate(p.Since); err != nil {
				return nil, fmt.Errorf("%w: earlier payout settings: %w", ErrCorruptPayout, err)
			}
		}
		out[i] = ingest.PayoutPeriod{Payout: payout, Since: since}
	}
	return out, nil
}
