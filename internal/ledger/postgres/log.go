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

// ErrCorruptLog reports a stored row that does not decode to a valid value,
// or a version whose content no longer matches its stored fingerprint.
var ErrCorruptLog = errors.New("corrupt transaction log")

const (
	// uniqueViolation is PostgreSQL's SQLSTATE for a unique constraint.
	uniqueViolation = "23505"
	// keyVersionConstraint is the unique (source, external_id, version).
	keyVersionConstraint = "transaction_versions_key_version"
)

const insertVersion = `
INSERT INTO ledger.transaction_versions
    (source, external_id, version, account_id, kind, date, lot_opened, description, fingerprint)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id`

const insertPosting = `
INSERT INTO ledger.postings (version_id, position, book_kind, isin, symbol, amount, currency, quantity)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

// Append stores v and its postings atomically: in a database transaction of
// its own, or in a savepoint inside InTx, so that a refused append leaves the
// surrounding transaction usable. A version
// its key already has fails with an error wrapping app.ErrVersionConflict;
// the database's unique (source, external_id, version) decides which of two
// racing appends wins.
func (s *Store) Append(ctx context.Context, v *ledger.VersionedTransaction) error {
	t := &v.Transaction
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		var id int64
		err := tx.QueryRow(ctx, insertVersion,
			t.Key.Source, t.Key.ExternalID, v.Version, string(t.Account), t.Kind.String(),
			t.Date.String(), lotOpened(t), t.Description, ledger.Fingerprint(t),
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("inserting the version: %w", err)
		}
		batch := &pgx.Batch{}
		for i := range t.Postings {
			p := &t.Postings[i]
			in := columnsOf(p.Book)
			batch.Queue(insertPosting, id, i, p.Book.Kind.String(), in.isin, in.symbol,
				p.Amount.Amount().String(), p.Amount.Currency().String(), p.Quantity.String())
		}
		if err := tx.SendBatch(ctx, batch).Close(); err != nil {
			return fmt.Errorf("inserting postings: %w", err)
		}
		return nil
	})
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation && pgErr.ConstraintName == keyVersionConstraint {
		return fmt.Errorf("%s version %d: %w", t.Key, v.Version, app.ErrVersionConflict)
	}
	if err != nil {
		return fmt.Errorf("appending %s version %d: %w", t.Key, v.Version, err)
	}
	return nil
}

// lotOpened returns t's lot open date, or nil for none.
func lotOpened(t *ledger.Transaction) *string {
	if t.LotOpened.IsZero() {
		return nil
	}
	d := t.LotOpened.String()
	return &d
}

// instrumentColumns holds the isin and symbol columns of a ledger account:
// a holding's instrument, or NULLs.
type instrumentColumns struct {
	isin, symbol *string
}

// columnsOf returns b's instrument columns.
func columnsOf(b ledger.Book) instrumentColumns {
	if b.Kind != ledger.BookHolding {
		return instrumentColumns{}
	}
	i, sym := b.Instrument.ISIN.String(), b.Instrument.Symbol
	return instrumentColumns{isin: &i, symbol: &sym}
}

const selectVersions = `
SELECT id, source, external_id, version, account_id, kind,
       to_char(date, 'YYYY-MM-DD'), coalesce(to_char(lot_opened, 'YYYY-MM-DD'), ''), description, fingerprint
FROM ledger.transaction_versions`

const selectPostings = `
SELECT version_id, book_kind, coalesce(isin, ''), coalesce(symbol, ''), amount::text, currency, quantity::text
FROM ledger.postings
WHERE version_id = ANY($1)
ORDER BY version_id, position`

// Versions returns every stored version of key, in version order.
func (s *Store) Versions(ctx context.Context, key ledger.TransactionKey) ([]ledger.VersionedTransaction, error) {
	out, err := s.load(ctx, selectVersions+` WHERE source = $1 AND external_id = $2 ORDER BY version`,
		key.Source, key.ExternalID)
	if err != nil {
		return nil, fmt.Errorf("loading versions of %s: %w", key, err)
	}
	return out, nil
}

// Log returns every version of every key that has a version in account, in
// key then version order.
func (s *Store) Log(ctx context.Context, account ledger.AccountID) ([]ledger.VersionedTransaction, error) {
	out, err := s.load(ctx, selectVersions+`
		WHERE (source, external_id) IN (
			SELECT source, external_id FROM ledger.transaction_versions WHERE account_id = $1)
		ORDER BY source COLLATE "C", external_id COLLATE "C", version`, string(account))
	if err != nil {
		return nil, fmt.Errorf("loading the log of %s: %w", account, err)
	}
	return out, nil
}

// storedVersion is one row of selectVersions.
type storedVersion struct {
	id          int64
	fingerprint string
	version     ledger.VersionedTransaction
}

// load runs a selectVersions query, attaches each version's postings and
// checks every version against its stored fingerprint. A version and its
// postings are committed together, so the second query sees all postings of
// every version the first one returned.
func (s *Store) load(ctx context.Context, query string, args ...any) ([]ledger.VersionedTransaction, error) {
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying versions: %w", err)
	}
	stored, err := pgx.CollectRows(rows, scanVersion)
	if err != nil {
		return nil, fmt.Errorf("reading versions: %w", err)
	}
	byID := make(map[int64]*ledger.Transaction, len(stored))
	ids := make([]int64, len(stored))
	for i := range stored {
		ids[i] = stored[i].id
		byID[stored[i].id] = &stored[i].version.Transaction
	}
	if err := s.attachPostings(ctx, ids, byID); err != nil {
		return nil, err
	}
	out := make([]ledger.VersionedTransaction, len(stored))
	for i := range stored {
		v := &stored[i]
		if got := ledger.Fingerprint(&v.version.Transaction); got != v.fingerprint {
			return nil, fmt.Errorf("%w: %s version %d does not match its fingerprint",
				ErrCorruptLog, v.version.Transaction.Key, v.version.Version)
		}
		out[i] = v.version
	}
	return out, nil
}

// attachPostings loads the postings of the versions ids and appends each to
// its transaction in byID, in stored order.
func (s *Store) attachPostings(ctx context.Context, ids []int64, byID map[int64]*ledger.Transaction) error {
	if len(ids) == 0 {
		return nil
	}
	rows, err := s.db.Query(ctx, selectPostings, ids)
	if err != nil {
		return fmt.Errorf("querying postings: %w", err)
	}
	var versionID int64
	var kind, isin, symbol, amount, currency, quantity string
	_, err = pgx.ForEachRow(rows, []any{&versionID, &kind, &isin, &symbol, &amount, &currency, &quantity}, func() error {
		p, decodeErr := decodePosting(kind, isin, symbol, amount, currency, quantity)
		if decodeErr != nil {
			return fmt.Errorf("version row %d: %w", versionID, decodeErr)
		}
		t := byID[versionID]
		t.Postings = append(t.Postings, p)
		return nil
	})
	if err != nil {
		return fmt.Errorf("reading postings: %w", err)
	}
	return nil
}

// scanVersion reads one row of selectVersions, without postings.
func scanVersion(row pgx.CollectableRow) (storedVersion, error) {
	var sv storedVersion
	var account, kind, date, lotOpened string
	t := &sv.version.Transaction
	if err := row.Scan(&sv.id, &t.Key.Source, &t.Key.ExternalID, &sv.version.Version, &account, &kind,
		&date, &lotOpened, &t.Description, &sv.fingerprint); err != nil {
		return storedVersion{}, fmt.Errorf("reading a version: %w", err)
	}
	t.Account = ledger.AccountID(account)
	var err error
	if t.Kind, err = parseKind(kind); err != nil {
		return storedVersion{}, err
	}
	if t.Date, err = parseDate(date); err != nil {
		return storedVersion{}, err
	}
	if lotOpened != "" {
		if t.LotOpened, err = parseDate(lotOpened); err != nil {
			return storedVersion{}, err
		}
	}
	return sv, nil
}
