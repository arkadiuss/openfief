package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ForgetAccount removes what the ingest schema holds of a deleted account
// (D-120), in one transaction: its connector bindings, and the mappings of
// the broker accounts mapped to it, which are left unmapped as when first
// reported, the go-live an opening upload recorded for it (D-160), and the
// ticker entries and holdings named in it (D-116, D-155).
// Forgetting it again changes nothing.
func (s *Store) ForgetAccount(ctx context.Context, account ledger.AccountID) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM ingest.connector_bindings WHERE account_id = $1`, string(account)); err != nil {
			return fmt.Errorf("bindings: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE ingest.broker_accounts SET account_id = NULL, mapped_at = NULL, seeded = false
			WHERE account_id = $1`, string(account)); err != nil {
			return fmt.Errorf("broker accounts: %w", err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM ingest.account_starts WHERE account_id = $1`, string(account)); err != nil {
			return fmt.Errorf("start: %w", err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM ingest.ticker_entries WHERE account_id = $1`, string(account)); err != nil {
			return fmt.Errorf("ticker entries: %w", err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM ingest.ticker_assertions WHERE account_id = $1`, string(account)); err != nil {
			return fmt.Errorf("ticker assertions: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("forgetting account %s: %w", account, err)
	}
	return nil
}

// RemapBrokerAccount maps instance's broker account id to account at at,
// or leaves it unmapped when account is empty, whatever it was mapped to
// (D-120): the caller checks that the old account holds nothing from it.
// It is no longer seeded, so the next opening batch seeds the new account
// from it. An unknown broker account fails with app.ErrUnknownBrokerAccount.
func (s *Store) RemapBrokerAccount(ctx context.Context, instance string, id ingest.BrokerAccountID, account ledger.AccountID, at time.Time) error {
	var mapped, mappedAt any
	if account != "" {
		mapped, mappedAt = string(account), at
	}
	var found string
	err := s.pool.QueryRow(ctx, `
		UPDATE ingest.broker_accounts SET account_id = $3, mapped_at = $4, seeded = false
		WHERE instance = $1 AND broker_account = $2 RETURNING broker_account`,
		instance, string(id), mapped, mappedAt).Scan(&found)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return fmt.Errorf("remapping a broker account of %s: %w", instance, app.ErrUnknownBrokerAccount)
	case err != nil:
		return fmt.Errorf("remapping a broker account of %s: %w", instance, err)
	}
	return nil
}
