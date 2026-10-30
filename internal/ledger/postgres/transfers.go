package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// linkVersionConstraint is the primary key (pair, version) of
// ledger.transfer_links.
const linkVersionConstraint = "transfer_links_pair_version"

// AppendTransferLink stores l as version l.Version of its pair (D-105). A
// version the pair already has fails with an error wrapping
// app.ErrVersionConflict; triggers refuse a version that skips a number and
// any change to a stored one (D-18).
func (s *Store) AppendTransferLink(ctx context.Context, l *ledger.TransferLink) error {
	w, d := l.Pair.Withdrawal, l.Pair.Deposit
	_, err := s.db.Exec(ctx, `
		INSERT INTO ledger.transfer_links
			(withdrawal_source, withdrawal_external_id, deposit_source, deposit_external_id, version, decision)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		w.Source, w.ExternalID, d.Source, d.ExternalID, l.Version, l.Decision.String())
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation && pgErr.ConstraintName == linkVersionConstraint {
		return fmt.Errorf("transfer %s version %d: %w", l.Pair, l.Version, app.ErrVersionConflict)
	}
	if err != nil {
		return fmt.Errorf("appending transfer %s version %d: %w", l.Pair, l.Version, err)
	}
	return nil
}

// TransferLinks returns every stored version of every pair's link, in pair
// then version order.
func (s *Store) TransferLinks(ctx context.Context) ([]ledger.TransferLink, error) {
	rows, err := s.db.Query(ctx, `
		SELECT withdrawal_source, withdrawal_external_id, deposit_source, deposit_external_id, version, decision
		FROM ledger.transfer_links
		ORDER BY withdrawal_source COLLATE "C", withdrawal_external_id COLLATE "C",
			deposit_source COLLATE "C", deposit_external_id COLLATE "C", version`)
	if err != nil {
		return nil, fmt.Errorf("loading transfer links: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanLink)
	if err != nil {
		return nil, fmt.Errorf("loading transfer links: %w", err)
	}
	return out, nil
}

// LockTransferLinks locks ledger.transfer_links against every other
// LockTransferLinks and every write until the transaction ends; reads go
// on. Outside InTx it fails with ErrNoTx.
func (s *Store) LockTransferLinks(ctx context.Context) error {
	if s.tx == nil {
		return fmt.Errorf("locking transfer links: %w", ErrNoTx)
	}
	if _, err := s.db.Exec(ctx, `LOCK TABLE ledger.transfer_links IN EXCLUSIVE MODE`); err != nil {
		return fmt.Errorf("locking transfer links: %w", err)
	}
	return nil
}

// scanLink reads one transfer link row.
func scanLink(row pgx.CollectableRow) (ledger.TransferLink, error) {
	var l ledger.TransferLink
	var decision string
	w, d := &l.Pair.Withdrawal, &l.Pair.Deposit
	if err := row.Scan(&w.Source, &w.ExternalID, &d.Source, &d.ExternalID, &l.Version, &decision); err != nil {
		return ledger.TransferLink{}, fmt.Errorf("reading a transfer link: %w", err)
	}
	var err error
	if l.Decision, err = parseDecision(decision); err != nil {
		return ledger.TransferLink{}, err
	}
	return l, nil
}

// parseDecision returns the transfer decision whose String is s.
func parseDecision(s string) (ledger.TransferDecision, error) {
	for d := ledger.TransferConfirmed; d <= ledger.TransferRejected; d++ {
		if d.String() == s {
			return d, nil
		}
	}
	return 0, fmt.Errorf("%w: transfer decision %q", ErrCorruptLog, s)
}
