package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

var _ app.BrokerAccountStore = (*Store)(nil)

// RecordBrokerAccounts stores the broker accounts a delivery of instance
// reported on on, new ones unmapped, and marks them, and only them, as in
// the instance's last delivery, in one transaction.
func (s *Store) RecordBrokerAccounts(ctx context.Context, instance string, institution ingest.InstitutionID, ids []ingest.BrokerAccountID, on kernel.Date) error {
	reported := make([]string, len(ids))
	for i, id := range ids {
		reported[i] = string(id)
	}
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			UPDATE ingest.broker_accounts SET in_last_delivery = false
			WHERE instance = $1 AND NOT broker_account = ANY($2)`, instance, reported); err != nil {
			return fmt.Errorf("clearing the last delivery's broker accounts: %w", err)
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO ingest.broker_accounts AS b (instance, broker_account, institution, last_seen, in_last_delivery)
			SELECT $1, r, $3, $4, true FROM unnest($2::text[]) AS r
			ON CONFLICT (instance, broker_account) DO UPDATE
				SET institution = EXCLUDED.institution, last_seen = EXCLUDED.last_seen, in_last_delivery = true`,
			instance, reported, string(institution), on.String())
		if err != nil {
			return fmt.Errorf("storing broker accounts: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("recording the broker accounts of %s: %w", instance, err)
	}
	return nil
}

// brokerColumns are the columns scanBrokerAccount reads.
const brokerColumns = `instance, broker_account, institution, coalesce(account_id, ''), last_seen::text, in_last_delivery, seeded`

// BrokerAccounts returns instance's broker accounts, in ID order.
func (s *Store) BrokerAccounts(ctx context.Context, instance string) ([]ingest.BrokerAccount, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+brokerColumns+` FROM ingest.broker_accounts
		WHERE instance = $1 ORDER BY broker_account COLLATE "C"`, instance)
	if err != nil {
		return nil, fmt.Errorf("loading the broker accounts of %s: %w", instance, err)
	}
	out, err := pgx.CollectRows(rows, scanBrokerAccount)
	if err != nil {
		return nil, fmt.Errorf("loading the broker accounts of %s: %w", instance, err)
	}
	return out, nil
}

// AllBrokerAccounts returns every instance's broker accounts, in instance
// then ID order.
func (s *Store) AllBrokerAccounts(ctx context.Context) ([]ingest.BrokerAccount, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+brokerColumns+` FROM ingest.broker_accounts
		ORDER BY instance COLLATE "C", broker_account COLLATE "C"`)
	if err != nil {
		return nil, fmt.Errorf("loading broker accounts: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanBrokerAccount)
	if err != nil {
		return nil, fmt.Errorf("loading broker accounts: %w", err)
	}
	return out, nil
}

// scanBrokerAccount reads one row of brokerColumns.
func scanBrokerAccount(row pgx.CollectableRow) (ingest.BrokerAccount, error) {
	var (
		b                   ingest.BrokerAccount
		id, inst, acc, seen string
	)
	if err := row.Scan(&b.Instance, &id, &inst, &acc, &seen, &b.InLastDelivery, &b.Seeded); err != nil {
		return ingest.BrokerAccount{}, fmt.Errorf("scanning: %w", err)
	}
	on, err := kernel.ParseDate(seen)
	if err != nil {
		return ingest.BrokerAccount{}, fmt.Errorf("broker account last seen: %w", err)
	}
	b.ID, b.Institution, b.Account, b.LastSeen = ingest.BrokerAccountID(id), ingest.InstitutionID(inst), ledger.AccountID(acc), on
	return b, nil
}

// MarkSeeded records that an opening batch of instance completed after ids
// were mapped.
func (s *Store) MarkSeeded(ctx context.Context, instance string, ids []ingest.BrokerAccountID) error {
	seeded := make([]string, len(ids))
	for i, id := range ids {
		seeded[i] = string(id)
	}
	if _, err := s.pool.Exec(ctx, `
		UPDATE ingest.broker_accounts SET seeded = true
		WHERE instance = $1 AND broker_account = ANY($2) AND account_id IS NOT NULL`, instance, seeded); err != nil {
		return fmt.Errorf("marking the broker accounts of %s seeded: %w", instance, err)
	}
	return nil
}

// MapBrokerAccount maps instance's broker account id to account at at,
// unless it is mapped to another account (app.ErrBrokerAccountMapped) or
// instance never reported it (app.ErrUnknownBrokerAccount). Mapping it to
// the account it is mapped to changes nothing.
func (s *Store) MapBrokerAccount(ctx context.Context, instance string, id ingest.BrokerAccountID, account ledger.AccountID, at time.Time) error {
	var mapped *string
	err := s.pool.QueryRow(ctx, `
		WITH m AS (
			UPDATE ingest.broker_accounts SET account_id = $3, mapped_at = $4
			WHERE instance = $1 AND broker_account = $2 AND account_id IS NULL
			RETURNING account_id)
		SELECT coalesce((SELECT account_id FROM m), b.account_id)
		FROM ingest.broker_accounts b WHERE b.instance = $1 AND b.broker_account = $2`,
		instance, string(id), string(account), at).Scan(&mapped)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return fmt.Errorf("mapping a broker account of %s: %w", instance, app.ErrUnknownBrokerAccount)
	case err != nil:
		return fmt.Errorf("mapping a broker account of %s: %w", instance, err)
	case mapped == nil || *mapped != string(account):
		return fmt.Errorf("mapping a broker account of %s: %w", instance, app.ErrBrokerAccountMapped)
	}
	return nil
}
