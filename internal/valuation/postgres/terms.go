package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// The keys of valuation.settings.
const (
	rulesVersionKey = "rules_version"
	marginalRateKey = "marginal_rate"
	numberFormatKey = "number_format"
)

// TaxWrapper returns account's recorded tax wrapper, and false when none is
// stored.
func (s *Store) TaxWrapper(ctx context.Context, account ledger.AccountID) (jurisdiction.Wrapper, bool, error) {
	var name string
	err := s.pool.QueryRow(ctx, `SELECT wrapper FROM valuation.tax_wrappers WHERE account_id = $1`, string(account)).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("loading the tax wrapper of %s: %w", account, err)
	}
	return jurisdiction.Wrapper(name), true, nil
}

// SaveTaxWrapper records account's tax wrapper and enqueues revals in one
// transaction.
func (s *Store) SaveTaxWrapper(ctx context.Context, account ledger.AccountID, wrapper jurisdiction.Wrapper, revals []valuation.Revaluation) error {
	err := s.inTx(ctx, revals, func(tx pgx.Tx) error {
		if _, execErr := tx.Exec(ctx, `
			INSERT INTO valuation.tax_wrappers (account_id, wrapper) VALUES ($1, $2)
			ON CONFLICT (account_id) DO UPDATE SET wrapper = EXCLUDED.wrapper`, string(account), string(wrapper)); execErr != nil {
			return fmt.Errorf("upserting: %w", execErr)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("saving the tax wrapper of %s: %w", account, err)
	}
	return nil
}

// FeeRule returns account's fee rule, the zero rule unless one is stored.
func (s *Store) FeeRule(ctx context.Context, account ledger.AccountID) (jurisdiction.FeeRule, error) {
	var rate, minimum, currency, conversion string
	err := s.pool.QueryRow(ctx, `
		SELECT rate::text, minimum::text, coalesce(minimum_currency, ''), conversion_rate::text
		FROM valuation.fee_rules WHERE account_id = $1`, string(account)).Scan(&rate, &minimum, &currency, &conversion)
	if errors.Is(err, pgx.ErrNoRows) {
		return jurisdiction.FeeRule{}, nil
	}
	if err != nil {
		return jurisdiction.FeeRule{}, fmt.Errorf("loading the fee rule of %s: %w", account, err)
	}
	r, err1 := kernel.ParseDecimal(rate)
	m, err2 := kernel.ParseDecimal(minimum)
	c, err3 := kernel.ParseDecimal(conversion)
	var cur kernel.Currency
	var err4 error
	if currency != "" {
		cur, err4 = kernel.ParseCurrency(currency)
	}
	if err := errors.Join(err1, err2, err3, err4); err != nil {
		return jurisdiction.FeeRule{}, fmt.Errorf("%w: fee rule of %s: %w", ErrCorrupt, account, err)
	}
	return jurisdiction.FeeRule{Rate: r, Minimum: kernel.NewMoney(m, cur), ConversionRate: c}, nil
}

// SaveFeeRule records account's fee rule and enqueues revals in one
// transaction.
func (s *Store) SaveFeeRule(ctx context.Context, account ledger.AccountID, rule *jurisdiction.FeeRule, revals []valuation.Revaluation) error {
	var currency *string
	if c := rule.Minimum.Currency(); !c.IsZero() {
		code := c.String()
		currency = &code
	}
	err := s.inTx(ctx, revals, func(tx pgx.Tx) error {
		if _, execErr := tx.Exec(ctx, `
			INSERT INTO valuation.fee_rules (account_id, rate, minimum, minimum_currency, conversion_rate)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (account_id) DO UPDATE SET rate = EXCLUDED.rate, minimum = EXCLUDED.minimum,
			    minimum_currency = EXCLUDED.minimum_currency, conversion_rate = EXCLUDED.conversion_rate`,
			string(account), rule.Rate.String(), rule.Minimum.Amount().String(), currency, rule.ConversionRate.String()); execErr != nil {
			return fmt.Errorf("upserting: %w", execErr)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("saving the fee rule of %s: %w", account, err)
	}
	return nil
}

// ValueSettings returns the owner's marginal income tax rate, unset unless
// one is stored.
func (s *Store) ValueSettings(ctx context.Context) (jurisdiction.Settings, error) {
	text, err := s.setting(ctx, marginalRateKey)
	if err != nil || text == "" {
		return jurisdiction.Settings{}, err
	}
	rate, err := kernel.ParseDecimal(text)
	if err != nil {
		return jurisdiction.Settings{}, fmt.Errorf("%w: marginal rate: %w", ErrCorrupt, err)
	}
	return jurisdiction.Settings{MarginalRate: rate, HasMarginalRate: true}, nil
}

// SaveMarginalRate records the owner's marginal income tax rate and
// enqueues revals in one transaction.
func (s *Store) SaveMarginalRate(ctx context.Context, rate kernel.Decimal, revals []valuation.Revaluation) error {
	err := s.inTx(ctx, revals, func(tx pgx.Tx) error {
		return putSetting(ctx, tx, marginalRateKey, rate.String())
	})
	if err != nil {
		return fmt.Errorf("saving the marginal rate: %w", err)
	}
	return nil
}

// NumberFormat returns the owner's number format, and false unless one is
// stored.
func (s *Store) NumberFormat(ctx context.Context) (jurisdiction.NumberFormat, bool, error) {
	key, err := s.setting(ctx, numberFormatKey)
	if err != nil || key == "" {
		return jurisdiction.NumberFormat{}, false, err
	}
	f, err := jurisdiction.ParseNumberFormat(key)
	if err != nil {
		return jurisdiction.NumberFormat{}, false, fmt.Errorf("%w: number format: %w", ErrCorrupt, err)
	}
	return f, true, nil
}

// SaveNumberFormat records the owner's number format.
func (s *Store) SaveNumberFormat(ctx context.Context, format jurisdiction.NumberFormat) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		return putSetting(ctx, tx, numberFormatKey, format.String())
	})
	if err != nil {
		return fmt.Errorf("saving the number format: %w", err)
	}
	return nil
}

// SaveRulesVersion records version and, only when the recorded version
// differs or none was recorded, enqueues revals, in one transaction. It
// reports whether it enqueued them.
func (s *Store) SaveRulesVersion(ctx context.Context, version string, revals []valuation.Revaluation) (bool, error) {
	changed := false
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var old string
		err := tx.QueryRow(ctx, `SELECT value FROM valuation.settings WHERE key = $1 FOR UPDATE`, rulesVersionKey).Scan(&old)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("reading the recorded rules version: %w", err)
		}
		if old == version {
			return nil
		}
		changed = true
		if err := putSetting(ctx, tx, rulesVersionKey, version); err != nil {
			return err
		}
		return s.enqueueTx(ctx, tx, revals)
	})
	if err != nil {
		return false, fmt.Errorf("saving the rules version: %w", err)
	}
	return changed, nil
}

// setting returns the stored value of key, empty when there is none: no
// setting is stored empty.
func (s *Store) setting(ctx context.Context, key string) (string, error) {
	var value string
	err := s.pool.QueryRow(ctx, `SELECT value FROM valuation.settings WHERE key = $1`, key).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("loading the setting %s: %w", key, err)
	}
	return value, nil
}

// putSetting upserts key's value in tx.
func putSetting(ctx context.Context, tx pgx.Tx, key, value string) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO valuation.settings (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`, key, value); err != nil {
		return fmt.Errorf("recording the setting %s: %w", key, err)
	}
	return nil
}
