package postgres_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/postgres"
)

// dsnVar names the server the tests create their databases on. task test:pg
// sets it to a disposable server; plain go test leaves it unset, and the
// tests skip.
const dsnVar = "OPENFIEF_PG_DSN"

// newPool returns a pool on a fresh database with the identity migrations
// applied, dropped when the test ends.
func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv(dsnVar)
	if dsn == "" {
		t.Skip(dsnVar + " is not set: task test:pg runs this test on a disposable PostgreSQL")
	}
	ctx := t.Context()
	sum := sha256.Sum256([]byte(t.Name()))
	name := "identity_" + hex.EncodeToString(sum[:8]) + "_" + strconv.Itoa(os.Getpid())
	adminExec(ctx, t, dsn, "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
	adminExec(ctx, t, dsn, "CREATE DATABASE "+name)
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parsing %s: %v", dsnVar, err)
	}
	cfg.ConnConfig.Database = name
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connecting to %s: %v", name, err)
	}
	t.Cleanup(func() {
		pool.Close()
		adminExec(context.WithoutCancel(ctx), t, dsn, "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
	})
	for range 2 { // a second run applies nothing
		if err := postgres.Migrate(ctx, pool); err != nil {
			t.Fatalf("Migrate: %v", err)
		}
	}
	return pool
}

// adminExec runs sql on the server's own database.
func adminExec(ctx context.Context, t *testing.T, dsn, sql string) {
	t.Helper()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connecting to %s: %v", dsnVar, err)
	}
	defer func() {
		if err := conn.Close(ctx); err != nil {
			t.Errorf("closing: %v", err)
		}
	}()
	if _, err := conn.Exec(ctx, sql); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

// issuedAt is when the tests issue tokens.
func issuedAt() time.Time { return time.Date(2026, 10, 28, 9, 30, 0, 0, time.UTC) }

func TestStoreRoundTrip(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := t.Context()
	store := postgres.New(pool)
	now := issuedAt()
	random := bytes.NewReader(bytes.Repeat([]byte{0x5a}, 2*(identity.IDBytes+identity.KeyBytes)))
	tokens := app.NewTokens(store, func() time.Time { return now }, random)
	first, secret, err := tokens.Issue(ctx, "first", time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, _, dupErr := tokens.Issue(ctx, "same-random", time.Hour); dupErr == nil {
		t.Error("Issue stored a second token with the same ID")
	}
	_, key, err := identity.ParseSecret(secret)
	if err != nil {
		t.Fatalf("ParseSecret: %v", err)
	}
	got, hash, err := store.ScriptToken(ctx, first.ID)
	if err != nil {
		t.Fatalf("ScriptToken: %v", err)
	}
	if diff := cmp.Diff(first, got); diff != "" {
		t.Errorf("ScriptToken (-want +got):\n%s", diff)
	}
	if !bytes.Equal(hash, identity.HashKey(key)) {
		t.Errorf("stored hash = %x, want the key's SHA-256", hash)
	}
	checkNoSecretStored(t, pool, key, secret)
	if verifyErr := tokens.Verify(ctx, secret, identity.ScopeRead); verifyErr != nil {
		t.Errorf("Verify = %v", verifyErr)
	}
	checkRevoked(t, tokens, &now, &first, secret)
}

// checkRevoked revokes first twice, a minute apart, and checks that its
// secret no longer verifies and that the first revocation's time is kept.
func checkRevoked(t *testing.T, tokens *app.Tokens, now *time.Time, first *identity.ScriptToken, secret string) {
	t.Helper()
	ctx := t.Context()
	*now = now.Add(time.Minute)
	revokedAt := *now
	for range 2 {
		if err := tokens.Revoke(ctx, first.ID); err != nil {
			t.Fatalf("Revoke: %v", err)
		}
		*now = now.Add(time.Minute)
	}
	if err := tokens.Verify(ctx, secret, identity.ScopeRead); !errors.Is(err, identity.ErrTokenRevoked) {
		t.Errorf("Verify after Revoke = %v, want %v", err, identity.ErrTokenRevoked)
	}
	listed, err := tokens.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	want := *first
	want.RevokedAt = revokedAt
	if diff := cmp.Diff([]identity.ScriptToken{want}, listed); diff != "" {
		t.Errorf("List (-want +got):\n%s", diff)
	}
}

// checkNoSecretStored fails if any column of identity.script_tokens holds the
// key or the secret.
func checkNoSecretStored(t *testing.T, pool *pgxpool.Pool, key []byte, secret string) {
	t.Helper()
	var dump string
	if err := pool.QueryRow(t.Context(), `SELECT string_agg(t::text, ' ') FROM identity.script_tokens t`).Scan(&dump); err != nil {
		t.Fatalf("dumping tokens: %v", err)
	}
	keyHex := hex.EncodeToString(key)
	_, keyText, _ := strings.Cut(strings.TrimPrefix(secret, identity.SecretPrefix), "_")
	if strings.Contains(dump, keyHex) || strings.Contains(dump, keyText) {
		t.Error("identity.script_tokens holds a token's key")
	}
}

func TestStoreUnknownToken(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	store := postgres.New(pool)
	const id identity.ScriptTokenID = "0123456789abcdef"
	if _, _, err := store.ScriptToken(t.Context(), id); !errors.Is(err, app.ErrUnknownToken) {
		t.Errorf("ScriptToken = %v, want %v", err, app.ErrUnknownToken)
	}
	if err := store.RevokeScriptToken(t.Context(), id, issuedAt()); !errors.Is(err, app.ErrUnknownToken) {
		t.Errorf("RevokeScriptToken = %v, want %v", err, app.ErrUnknownToken)
	}
	listed, err := store.ScriptTokens(t.Context())
	if err != nil || len(listed) != 0 {
		t.Errorf("ScriptTokens = %v, %v; want none", listed, err)
	}
}

func TestStoreRefusesBadRows(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	store := postgres.New(pool)
	good := identity.ScriptToken{
		ID: "0123456789abcdef", Name: "x", Scope: identity.ScopeRead,
		CreatedAt: issuedAt(), ExpiresAt: issuedAt().Add(time.Hour),
	}
	hash := identity.HashKey(make([]byte, identity.KeyBytes))
	tests := []struct {
		name string
		edit func(tok *identity.ScriptToken)
		hash []byte
	}{
		{name: "short hash", edit: func(*identity.ScriptToken) {}, hash: hash[:31]},
		{name: "expires before it is created", edit: func(tok *identity.ScriptToken) { tok.ExpiresAt = tok.CreatedAt }, hash: hash},
		{name: "upper-case ID", edit: func(tok *identity.ScriptToken) { tok.ID = "0123456789ABCDEF" }, hash: hash},
		{name: "name with a space", edit: func(tok *identity.ScriptToken) { tok.Name = "a b" }, hash: hash},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tok := good
			tt.edit(&tok)
			if err := store.SaveScriptToken(t.Context(), &tok, tt.hash); err == nil {
				t.Error("SaveScriptToken stored a bad row")
			}
		})
	}
}
