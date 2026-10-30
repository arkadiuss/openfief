package postgres_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// TestFeeRulesRoundTrip stores fee rules: nothing stored reads as the zero
// rule, a rule with a USD minimum and one without a minimum read back as
// stored, and saving one enqueues its revaluations in its own transaction.
func TestFeeRulesRoundTrip(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	rule, err := store.FeeRule(ctx, acc1)
	check(t, "fee rule before any", jurisdiction.FeeRule{}, rule, err)

	want := jurisdiction.FeeRule{Rate: dec(t, "0.0029"), Minimum: money(t, "5", "USD"), ConversionRate: dec(t, "0.005")}
	if saveErr := store.SaveFeeRule(ctx, acc1, &want, []valuation.Revaluation{reval(t, acc1, "2026-03-02")}); saveErr != nil {
		t.Fatalf("SaveFeeRule: %v", saveErr)
	}
	rule, err = store.FeeRule(ctx, acc1)
	check(t, "fee rule", want, rule, err)
	zeroMinimum := jurisdiction.FeeRule{Rate: dec(t, "0.001"), Minimum: kernel.NewMoney(kernel.Decimal{}, kernel.Currency{})}
	if saveErr := store.SaveFeeRule(ctx, acc2, &zeroMinimum, nil); saveErr != nil {
		t.Fatalf("SaveFeeRule without a minimum: %v", saveErr)
	}
	rule, err = store.FeeRule(ctx, acc2)
	check(t, "fee rule without a minimum", zeroMinimum, rule, err)
	if diff := cmp.Diff([]string{"acc-1 available"}, liveJobs(t, pool)); diff != "" {
		t.Errorf("jobs (-want +got):\n%s", diff)
	}
}

// TestOwnerSettingsRoundTrip stores the owner's marginal rate and number
// format: unset until stored, then read back as stored; the rate enqueues
// its revaluations, the format none.
func TestOwnerSettingsRoundTrip(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	settings, err := store.ValueSettings(ctx)
	check(t, "settings before any", jurisdiction.Settings{}, settings, err)
	_, chosen, err := store.NumberFormat(ctx)
	check(t, "number format chosen before any", false, chosen, err)

	if saveErr := store.SaveMarginalRate(ctx, dec(t, "0.32"), []valuation.Revaluation{reval(t, acc2, "2026-03-03")}); saveErr != nil {
		t.Fatalf("SaveMarginalRate: %v", saveErr)
	}
	settings, err = store.ValueSettings(ctx)
	check(t, "settings", jurisdiction.Settings{MarginalRate: dec(t, "0.32"), HasMarginalRate: true}, settings, err)

	format := jurisdiction.NumberFormat{Group: jurisdiction.SeparatorDot, Decimal: jurisdiction.SeparatorComma}
	if saveErr := store.SaveNumberFormat(ctx, format); saveErr != nil {
		t.Fatalf("SaveNumberFormat: %v", saveErr)
	}
	got, chosen, err := store.NumberFormat(ctx)
	check(t, "number format", format, got, err)
	check(t, "number format chosen", true, chosen, nil)
	if diff := cmp.Diff([]string{"acc-2 available"}, liveJobs(t, pool)); diff != "" {
		t.Errorf("jobs (-want +got):\n%s", diff)
	}
}
