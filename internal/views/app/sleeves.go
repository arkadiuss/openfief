package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// Portfolio is what the views read from the portfolio context, which keeps
// the sleeve kinds, the sleeves and which account is in which sleeve: the
// ledger knows nothing of sleeves. *portfolio/app.Registry implements it.
type Portfolio interface {
	// Sleeves returns every sleeve, in ID order.
	Sleeves(ctx context.Context) ([]portfolio.Sleeve, error)
	// Kinds returns every sleeve kind, in key order.
	Kinds(ctx context.Context) ([]portfolio.Kind, error)
	// AccountSleeves returns the sleeve of every account in one.
	AccountSleeves(ctx context.Context) (map[ledger.AccountID]portfolio.SleeveID, error)
}

// PayoutSettings reads an account's payout setting (D-135, D-162) from the
// ingest context, the default, as cash, when none is stored.
// *ingest/postgres.Store implements it.
type PayoutSettings interface {
	PayoutSetting(ctx context.Context, account ledger.AccountID) (ingest.PayoutSetting, error)
}

// WithPortfolio reads the sleeves, their kinds and which account is in
// which sleeve from p. Without it there are no sleeves and no account is
// in one.
func WithPortfolio(p Portfolio) Option {
	return func(v *Views) { v.portfolio = p }
}

// WithPayoutSettings reads the accounts' payout settings from p. Without
// it every account keeps its payouts as cash.
func WithPayoutSettings(p PayoutSettings) Option {
	return func(v *Views) { v.payouts = p }
}

// AccountSettings is what the read API shows of an account beyond the
// ledger's account and its terms: the sleeve the portfolio context puts it
// in and the payout setting the ingest context keeps for it.
type AccountSettings struct {
	// Sleeve is the account's sleeve; empty for none.
	Sleeve portfolio.SleeveID
	// Payout is the account's payout setting; the zero setting keeps its
	// payouts as cash.
	Payout ingest.PayoutSetting
}

// AccountSettings returns account's sleeve and payout setting.
func (v *Views) AccountSettings(ctx context.Context, account ledger.AccountID) (AccountSettings, error) {
	sleeves, err := v.accountSleeves(ctx)
	if err != nil {
		return AccountSettings{}, err
	}
	out := AccountSettings{Sleeve: sleeves[account]}
	if v.payouts == nil {
		return out, nil
	}
	if out.Payout, err = v.payouts.PayoutSetting(ctx, account); err != nil {
		return AccountSettings{}, fmt.Errorf("loading the payout setting of %s: %w", account, err)
	}
	return out, nil
}

// Sleeves returns every sleeve, in ID order.
func (v *Views) Sleeves(ctx context.Context) ([]portfolio.Sleeve, error) {
	if v.portfolio == nil {
		return nil, nil
	}
	s, err := v.portfolio.Sleeves(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading sleeves: %w", err)
	}
	return s, nil
}

// SleeveKinds returns every sleeve kind, in key order (D-71).
func (v *Views) SleeveKinds(ctx context.Context) ([]portfolio.Kind, error) {
	if v.portfolio == nil {
		return nil, nil
	}
	k, err := v.portfolio.Kinds(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading sleeve kinds: %w", err)
	}
	return k, nil
}

// accountSleeves returns the sleeve of every account in one.
func (v *Views) accountSleeves(ctx context.Context) (map[ledger.AccountID]portfolio.SleeveID, error) {
	if v.portfolio == nil {
		return map[ledger.AccountID]portfolio.SleeveID{}, nil
	}
	m, err := v.portfolio.AccountSleeves(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading the accounts' sleeves: %w", err)
	}
	return m, nil
}
