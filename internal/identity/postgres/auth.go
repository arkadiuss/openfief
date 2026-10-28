package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

var _ app.AuthStore = (*Store)(nil)

// errBadRow reports a stored row the domain can't hold.
var errBadRow = errors.New("bad row")

// Owner returns the owner, or an error wrapping app.ErrNoOwner.
func (s *Store) Owner(ctx context.Context) (identity.Owner, error) {
	var o identity.Owner
	err := s.pool.QueryRow(ctx, `SELECT handle, created_at FROM identity.owner WHERE id = 1`).Scan(&o.Handle, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.Owner{}, app.ErrNoOwner
	}
	if err != nil {
		return identity.Owner{}, fmt.Errorf("loading the owner: %w", err)
	}
	o.CreatedAt = o.CreatedAt.UTC()
	return o, nil
}

// EnsureOwner stores candidate as the owner unless there is one, and
// returns the stored owner.
func (s *Store) EnsureOwner(ctx context.Context, candidate identity.Owner) (identity.Owner, error) {
	_, err := s.pool.Exec(ctx, `INSERT INTO identity.owner (id, handle, created_at) VALUES (1, $1, $2) ON CONFLICT (id) DO NOTHING`,
		candidate.Handle, candidate.CreatedAt)
	if err != nil {
		return identity.Owner{}, fmt.Errorf("storing the owner: %w", err)
	}
	return s.Owner(ctx)
}

// webauthnColumns are the columns scanPasskey reads, in its order.
const webauthnColumns = `credential_id, rp_id, name, public_key, attestation_type, attestation_format, aaguid,
	sign_count, transports, user_verified, backup_eligible, backup_state, created_at, last_used_at`

// SavePasskey stores a new passkey.
func (s *Store) SavePasskey(ctx context.Context, p *identity.Passkey) error {
	var lastUsed *time.Time
	if !p.LastUsedAt.IsZero() {
		lastUsed = &p.LastUsedAt
	}
	transports := p.Transports
	if transports == nil {
		transports = []string{}
	}
	aaguid := p.AAGUID
	if aaguid == nil {
		aaguid = []byte{}
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO identity.passkeys (`+webauthnColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		p.CredentialID, p.RPID, p.Name, p.PublicKey, p.AttestationType, p.AttestationFormat, aaguid,
		int64(p.SignCount), transports, p.UserVerified, p.BackupEligible, p.BackupState, p.CreatedAt, lastUsed)
	if err != nil {
		return fmt.Errorf("storing passkey %s: %w", p.ID(), err)
	}
	return nil
}

// Passkeys returns every passkey, in creation then ID order.
func (s *Store) Passkeys(ctx context.Context) ([]identity.Passkey, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+webauthnColumns+` FROM identity.passkeys ORDER BY created_at, credential_id`)
	if err != nil {
		return nil, fmt.Errorf("listing passkeys: %w", err)
	}
	keys, err := pgx.CollectRows(rows, scanPasskey)
	if err != nil {
		return nil, fmt.Errorf("listing passkeys: %w", err)
	}
	return keys, nil
}

// scanPasskey reads a row of webauthnColumns.
func scanPasskey(row pgx.CollectableRow) (identity.Passkey, error) {
	var (
		p        identity.Passkey
		count    int64
		lastUsed *time.Time
	)
	err := row.Scan(&p.CredentialID, &p.RPID, &p.Name, &p.PublicKey, &p.AttestationType, &p.AttestationFormat, &p.AAGUID,
		&count, &p.Transports, &p.UserVerified, &p.BackupEligible, &p.BackupState, &p.CreatedAt, &lastUsed)
	if err != nil {
		return identity.Passkey{}, fmt.Errorf("scanning a passkey: %w", err)
	}
	if count < 0 || count > math.MaxUint32 {
		return identity.Passkey{}, fmt.Errorf("%w: passkey %s: sign count %d", errBadRow, p.ID(), count)
	}
	p.SignCount = uint32(count)
	p.CreatedAt = p.CreatedAt.UTC()
	if lastUsed != nil {
		p.LastUsedAt = lastUsed.UTC()
	}
	return p, nil
}

// UpdatePasskeyUse stores p's sign count, backup state and last use.
func (s *Store) UpdatePasskeyUse(ctx context.Context, p *identity.Passkey) error {
	tag, err := s.pool.Exec(ctx, `UPDATE identity.passkeys SET sign_count = $2, backup_state = $3, last_used_at = $4
		WHERE credential_id = $1`, p.CredentialID, int64(p.SignCount), p.BackupState, p.LastUsedAt)
	if err != nil {
		return fmt.Errorf("recording the use of passkey %s: %w", p.ID(), err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", app.ErrUnknownPasskey, p.ID())
	}
	return nil
}

// DeletePasskey deletes the passkey id unless it is the last one, under a
// lock on the table so two removals can't both pass the check.
func (s *Store) DeletePasskey(ctx context.Context, id []byte) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `LOCK TABLE identity.passkeys IN SHARE ROW EXCLUSIVE MODE`); err != nil {
			return fmt.Errorf("removing a passkey: %w", err)
		}
		var exists bool
		var n int
		err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT FROM identity.passkeys WHERE credential_id = $1), count(*) FROM identity.passkeys`, id).
			Scan(&exists, &n)
		switch {
		case err != nil:
			return fmt.Errorf("removing a passkey: %w", err)
		case !exists:
			return app.ErrUnknownPasskey
		case n <= 1:
			return app.ErrLastPasskey
		}
		if _, err := tx.Exec(ctx, `DELETE FROM identity.passkeys WHERE credential_id = $1`, id); err != nil {
			return fmt.Errorf("removing a passkey: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("removing a passkey: %w", err)
	}
	return nil
}

// SaveCeremony stores a ceremony, deleting those expired at its start.
func (s *Store) SaveCeremony(ctx context.Context, c *identity.Ceremony) error {
	var code *string
	if c.EnrollmentCode != "" {
		code = (*string)(&c.EnrollmentCode)
	}
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM identity.ceremonies WHERE expires_at <= $1`, c.CreatedAt); err != nil {
			return fmt.Errorf("deleting expired ceremonies: %w", err)
		}
		_, err := tx.Exec(ctx, `INSERT INTO identity.ceremonies (id, kind, origin, state, enrollment_code_id, created_at, expires_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`, c.ID, c.Kind.String(), c.Origin.String(), c.State, code, c.CreatedAt, c.ExpiresAt)
		if err != nil {
			return fmt.Errorf("storing a ceremony: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("beginning a ceremony: %w", err)
	}
	return nil
}

// TakeCeremony deletes and returns the ceremony id, or fails with an error
// wrapping app.ErrUnknownCeremony.
func (s *Store) TakeCeremony(ctx context.Context, id string) (identity.Ceremony, error) {
	var (
		c            identity.Ceremony
		kind, origin string
		code         *string
	)
	err := s.pool.QueryRow(ctx, `DELETE FROM identity.ceremonies WHERE id = $1
		RETURNING id, kind, origin, state, enrollment_code_id, created_at, expires_at`, id).
		Scan(&c.ID, &kind, &origin, &c.State, &code, &c.CreatedAt, &c.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.Ceremony{}, app.ErrUnknownCeremony
	}
	if err != nil {
		return identity.Ceremony{}, fmt.Errorf("taking a ceremony: %w", err)
	}
	if c.Kind, err = identity.ParseCeremonyKind(kind); err != nil {
		return identity.Ceremony{}, fmt.Errorf("ceremony %s: %w", c.ID, err)
	}
	if c.Origin, err = identity.ParseOrigin(origin); err != nil {
		return identity.Ceremony{}, fmt.Errorf("ceremony %s: %w", c.ID, err)
	}
	if code != nil {
		c.EnrollmentCode = identity.EnrollmentCodeID(*code)
	}
	c.CreatedAt, c.ExpiresAt = c.CreatedAt.UTC(), c.ExpiresAt.UTC()
	return c, nil
}

// SaveEnrollmentCode stores a new code with its key's hash.
func (s *Store) SaveEnrollmentCode(ctx context.Context, c *identity.EnrollmentCode, keyHash []byte) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO identity.enrollment_codes (id, origin, key_sha256, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)`, string(c.ID), c.Origin.String(), keyHash, c.CreatedAt, c.ExpiresAt)
	if err != nil {
		return fmt.Errorf("storing enrollment code %s: %w", c.ID, err)
	}
	return nil
}

// EnrollmentCode returns the code id and its key's hash, or an error
// wrapping app.ErrUnknownEnrollmentCode.
func (s *Store) EnrollmentCode(ctx context.Context, id identity.EnrollmentCodeID) (identity.EnrollmentCode, []byte, error) {
	var (
		c      identity.EnrollmentCode
		origin string
		hash   []byte
		used   *time.Time
	)
	err := s.pool.QueryRow(ctx, `SELECT id, origin, key_sha256, created_at, expires_at, used_at
		FROM identity.enrollment_codes WHERE id = $1`, string(id)).Scan(&c.ID, &origin, &hash, &c.CreatedAt, &c.ExpiresAt, &used)
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.EnrollmentCode{}, nil, fmt.Errorf("%w: %s", app.ErrUnknownEnrollmentCode, id)
	}
	if err != nil {
		return identity.EnrollmentCode{}, nil, fmt.Errorf("loading enrollment code %s: %w", id, err)
	}
	if c.Origin, err = identity.ParseOrigin(origin); err != nil {
		return identity.EnrollmentCode{}, nil, fmt.Errorf("enrollment code %s: %w", id, err)
	}
	c.CreatedAt, c.ExpiresAt = c.CreatedAt.UTC(), c.ExpiresAt.UTC()
	if used != nil {
		c.UsedAt = used.UTC()
	}
	return c, hash, nil
}

// UseEnrollmentCode marks the code id used at at, failing with an error
// wrapping identity.ErrEnrollmentCodeUsed when it already was.
func (s *Store) UseEnrollmentCode(ctx context.Context, id identity.EnrollmentCodeID, at time.Time) error {
	tag, err := s.pool.Exec(ctx, `UPDATE identity.enrollment_codes SET used_at = GREATEST($2, created_at)
		WHERE id = $1 AND used_at IS NULL`, string(id), at)
	if err != nil {
		return fmt.Errorf("using enrollment code %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", identity.ErrEnrollmentCodeUsed, id)
	}
	return nil
}

// sessionColumns are the columns scanSession reads, in its order.
const sessionColumns = `id, scope, authenticator, rp_id, created_at, last_seen_at, expires_at, idle_seconds, revoked_at`

// SaveSession stores a new session with its key's hash.
func (s *Store) SaveSession(ctx context.Context, ss *identity.Session, keyHash []byte) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO identity.sessions
		(id, scope, authenticator, rp_id, key_sha256, created_at, last_seen_at, expires_at, idle_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		string(ss.ID), ss.Scope.String(), ss.Authenticator.String(), ss.RPID, keyHash,
		ss.CreatedAt, ss.LastSeenAt, ss.ExpiresAt, int64(ss.IdleTimeout/time.Second))
	if err != nil {
		return fmt.Errorf("storing session %s: %w", ss.ID, err)
	}
	return nil
}

// Session returns the session id and its key's hash, or an error wrapping
// app.ErrUnknownSession.
func (s *Store) Session(ctx context.Context, id identity.SessionID) (identity.Session, []byte, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+sessionColumns+`, key_sha256 FROM identity.sessions WHERE id = $1`, string(id))
	if err != nil {
		return identity.Session{}, nil, fmt.Errorf("loading session %s: %w", id, err)
	}
	var hash []byte
	ss, err := pgx.CollectExactlyOneRow(rows, func(row pgx.CollectableRow) (identity.Session, error) {
		return scanSession(row, &hash)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.Session{}, nil, fmt.Errorf("%w: %s", app.ErrUnknownSession, id)
	}
	if err != nil {
		return identity.Session{}, nil, fmt.Errorf("loading session %s: %w", id, err)
	}
	return ss, hash, nil
}

// scanSession reads a row of sessionColumns, then into extra any further
// columns.
func scanSession(row pgx.CollectableRow, extra ...any) (identity.Session, error) {
	var (
		ss                         identity.Session
		id, scope, auth            string
		idle                       int64
		revoked                    *time.Time
		created, lastSeen, expires time.Time
	)
	dest := append([]any{&id, &scope, &auth, &ss.RPID, &created, &lastSeen, &expires, &idle, &revoked}, extra...)
	if err := row.Scan(dest...); err != nil {
		return identity.Session{}, fmt.Errorf("scanning a session: %w", err)
	}
	var err error
	if ss.Scope, err = identity.ParseSessionScope(scope); err != nil {
		return identity.Session{}, fmt.Errorf("session %s: %w", id, err)
	}
	if ss.Authenticator, err = identity.ParseAuthenticator(auth); err != nil {
		return identity.Session{}, fmt.Errorf("session %s: %w", id, err)
	}
	ss.ID = identity.SessionID(id)
	ss.CreatedAt, ss.LastSeenAt, ss.ExpiresAt = created.UTC(), lastSeen.UTC(), expires.UTC()
	ss.IdleTimeout = time.Duration(idle) * time.Second
	if revoked != nil {
		ss.RevokedAt = revoked.UTC()
	}
	return ss, nil
}

// TouchSession stores at as the session's last use, never moving it back.
func (s *Store) TouchSession(ctx context.Context, id identity.SessionID, at time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE identity.sessions SET last_seen_at = GREATEST(last_seen_at, $2) WHERE id = $1`, string(id), at)
	if err != nil {
		return fmt.Errorf("recording the use of session %s: %w", id, err)
	}
	return nil
}

// RevokeSession marks the session revoked at at, keeping an earlier
// revocation, or fails with an error wrapping app.ErrUnknownSession.
func (s *Store) RevokeSession(ctx context.Context, id identity.SessionID, at time.Time) error {
	tag, err := s.pool.Exec(ctx, `UPDATE identity.sessions SET revoked_at = COALESCE(revoked_at, GREATEST($2, created_at))
		WHERE id = $1`, string(id), at)
	if err != nil {
		return fmt.Errorf("revoking session %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", app.ErrUnknownSession, id)
	}
	return nil
}
