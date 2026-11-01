package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// storedPeriod is one earlier payout setting as payout_earlier holds it
// (D-162): its key and the day it took effect, empty for the first.
type storedPeriod struct {
	Payout string `json:"payout"`
	Since  string `json:"since"`
}

// encodeEarlier returns the payout_earlier JSON of s's earlier settings.
func encodeEarlier(s *ledger.PayoutSetting) ([]byte, error) {
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
func decodeEarlier(b []byte) ([]ledger.PayoutPeriod, error) {
	var stored []storedPeriod
	if err := json.Unmarshal(b, &stored); err != nil {
		return nil, fmt.Errorf("%w: earlier payout settings: %w", ErrCorruptLog, err)
	}
	if len(stored) == 0 {
		return nil, nil
	}
	out := make([]ledger.PayoutPeriod, len(stored))
	for i, p := range stored {
		payout, err := ledger.ParsePayout(p.Payout)
		if err != nil {
			return nil, fmt.Errorf("%w: earlier payout settings: %w", ErrCorruptLog, err)
		}
		var since kernel.Date
		if p.Since != "" {
			if since, err = parseDate(p.Since); err != nil {
				return nil, err
			}
		}
		out[i] = ledger.PayoutPeriod{Payout: payout, Since: since}
	}
	return out, nil
}
