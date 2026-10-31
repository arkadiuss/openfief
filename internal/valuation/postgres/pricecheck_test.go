package postgres_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// TestPriceCheckSettings checks the owner's price check thresholds (D-110)
// and acknowledgements (D-112) in PostgreSQL: a threshold is set, replaced
// and cleared per instrument; acknowledgements come back in the order
// recorded, and the table refuses to change or delete one.
func TestPriceCheckSettings(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	const etf, other = etfISIN + "/TSTB", "PLTEST000010/TSTA"

	five, two := dec(t, "0.05"), dec(t, "0.02")
	for _, step := range []struct {
		instrument string
		threshold  *kernel.Decimal
	}{{etf, &five}, {other, &two}, {etf, &two}, {other, nil}} {
		if err := store.SavePriceCheckThreshold(ctx, step.instrument, step.threshold); err != nil {
			t.Fatalf("SavePriceCheckThreshold(%s): %v", step.instrument, err)
		}
	}
	got, err := store.PriceCheckThresholds(ctx)
	check(t, "thresholds", valuation.Thresholds{ByInstrument: map[string]kernel.Decimal{etf: two}}, got, err)
	one := dec(t, "1")
	if saveErr := store.SavePriceCheckThreshold(ctx, etf, &one); saveErr == nil {
		t.Error("saving a threshold of 1 succeeded; the table check must refuse it")
	}

	acks := []valuation.PriceCheckAcknowledgement{
		{
			Instrument: etf, Date: date(t, "2026-03-08"), Source: "ibkr-main", CrossCheck: money(t, "103", "USD"),
			PriceDate: date(t, "2026-03-06"), Price: money(t, "100", "USD"),
		},
		{
			Instrument: etf, Date: date(t, "2026-03-08"), Source: "ibkr-main", CrossCheck: money(t, "104.5", "USD"),
			PriceDate: date(t, "2026-03-06"), Price: money(t, "100", "USD"),
		},
	}
	for i := range acks {
		if appendErr := store.AppendPriceCheckAcknowledgement(ctx, &acks[i]); appendErr != nil {
			t.Fatalf("AppendPriceCheckAcknowledgement: %v", appendErr)
		}
	}
	gotAcks, err := store.PriceCheckAcknowledgements(ctx)
	check(t, "acknowledgements", acks, gotAcks, err)
	for _, sql := range []string{
		`UPDATE valuation.price_check_acknowledgements SET source = 'other'`,
		`DELETE FROM valuation.price_check_acknowledgements`,
		`TRUNCATE valuation.price_check_acknowledgements`,
	} {
		if _, err := pool.Exec(ctx, sql); err == nil {
			t.Errorf("%s succeeded; acknowledgements are append-only", sql)
		}
	}
}
