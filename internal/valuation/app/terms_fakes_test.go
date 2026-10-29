package app_test

import (
	"context"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// fakeTerms are the accounts' terms and the owner's settings in memory,
// for fakeStore.
type fakeTerms struct {
	wrappers map[ledger.AccountID]jurisdiction.Wrapper
	fees     map[ledger.AccountID]jurisdiction.FeeRule
	settings jurisdiction.Settings
	format   *jurisdiction.NumberFormat
}

func (f *fakeStore) TaxWrapper(_ context.Context, account ledger.AccountID) (jurisdiction.Wrapper, bool, error) {
	w, ok := f.wrappers[account]
	return w, ok, f.err
}

func (f *fakeStore) FeeRule(_ context.Context, account ledger.AccountID) (jurisdiction.FeeRule, error) {
	return f.fees[account], f.err
}

func (f *fakeStore) ValueSettings(context.Context) (jurisdiction.Settings, error) {
	return f.settings, f.err
}

func (f *fakeStore) NumberFormat(context.Context) (jurisdiction.NumberFormat, bool, error) {
	if f.format == nil {
		return jurisdiction.NumberFormat{}, false, f.err
	}
	return *f.format, true, f.err
}

func (f *fakeRecorder) SaveTaxWrapper(_ context.Context, _ ledger.AccountID, w jurisdiction.Wrapper, revals []valuation.Revaluation) error {
	return f.record("wrapper "+string(w), revals)
}

func (f *fakeRecorder) SaveFeeRule(_ context.Context, _ ledger.AccountID, r *jurisdiction.FeeRule, revals []valuation.Revaluation) error {
	return f.record("fee "+r.Rate.String(), revals)
}

func (f *fakeRecorder) SaveMarginalRate(_ context.Context, rate kernel.Decimal, revals []valuation.Revaluation) error {
	return f.record("marginal rate "+rate.String(), revals)
}

func (f *fakeRecorder) SaveNumberFormat(_ context.Context, format jurisdiction.NumberFormat) error {
	return f.record("number format "+format.String(), nil)
}

func (f *fakeRecorder) SaveRulesVersion(_ context.Context, version string, revals []valuation.Revaluation) (bool, error) {
	return f.changed, f.record("rules "+version, revals)
}
