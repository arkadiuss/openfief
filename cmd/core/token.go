package main

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
)

// tokenAdmin is the script token use cases of the core's own commands. It
// issues and revokes as the system actor, so each change is in the audit log
// with the change itself (D-211).
type tokenAdmin struct {
	store  identityapp.AuthStore
	tokens *identityapp.Tokens
}

// systemActor is the core, as the audit log names it.
func systemActor() identity.Actor {
	return identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}
}

// Issue creates a token and its audit entry.
func (a tokenAdmin) Issue(ctx context.Context, name string, ttl time.Duration) (identity.ScriptToken, string, error) {
	tok, secret, err := identityapp.IssueToken(ctx, a.store, a.tokens, systemActor(), name, ttl)
	if err != nil {
		return identity.ScriptToken{}, "", fmt.Errorf("issuing as the core: %w", err)
	}
	return tok, secret, nil
}

// Revoke revokes a token and writes its audit entry.
func (a tokenAdmin) Revoke(ctx context.Context, id identity.ScriptTokenID) error {
	if err := identityapp.RevokeToken(ctx, a.store, a.tokens, systemActor(), id); err != nil {
		return fmt.Errorf("revoking as the core: %w", err)
	}
	return nil
}

// List returns every token.
func (a tokenAdmin) List(ctx context.Context) ([]identity.ScriptToken, error) {
	toks, err := a.tokens.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing: %w", err)
	}
	return toks, nil
}

// tokenCommand runs a token command over tokens and writes what it shows to
// out: for token issue, the secret alone on the first line, then the
// token's ID and expiry; nothing else ever shows a secret.
func tokenCommand(ctx context.Context, cmd *command, tokens tokenAdmin, out io.Writer, now time.Time) error {
	switch cmd.Name {
	case cmdTokenIssue:
		tok, secret, err := tokens.Issue(ctx, cmd.TokenName, cmd.TTL)
		if err != nil {
			return fmt.Errorf("token issue: %w", err)
		}
		_, err = fmt.Fprintf(out, "%s\nid %s, read-only, expires %s; the secret above is shown only once\n",
			secret, tok.ID, tok.ExpiresAt.Format(time.RFC3339))
		return writeErr(err)
	case cmdTokenRevoke:
		if err := tokens.Revoke(ctx, identity.ScriptTokenID(cmd.TokenID)); err != nil {
			return fmt.Errorf("token revoke: %w", err)
		}
		_, err := fmt.Fprintf(out, "revoked %s\n", cmd.TokenID)
		return writeErr(err)
	case cmdTokenList:
		toks, err := tokens.List(ctx)
		if err != nil {
			return fmt.Errorf("token list: %w", err)
		}
		return writeErr(listTokens(out, toks, now))
	}
	return fmt.Errorf("%w: %q", errUsage, cmd.Name)
}

// columnGap is the spaces between the columns of token list.
const columnGap = 2

// listTokens writes toks as a table, with each one's status at now.
func listTokens(out io.Writer, toks []identity.ScriptToken, now time.Time) error {
	tw := tabwriter.NewWriter(out, 0, 0, columnGap, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tNAME\tSCOPE\tCREATED\tEXPIRES\tSTATUS"); err != nil {
		return fmt.Errorf("listing: %w", err)
	}
	for i := range toks {
		t := &toks[i]
		_, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", t.ID, t.Name, t.Scope,
			t.CreatedAt.Format(time.RFC3339), t.ExpiresAt.Format(time.RFC3339), t.Status(now))
		if err != nil {
			return fmt.Errorf("listing: %w", err)
		}
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("listing: %w", err)
	}
	return nil
}

// writeErr wraps an error writing a command's output.
func writeErr(err error) error {
	if err != nil {
		return fmt.Errorf("writing the output: %w", err)
	}
	return nil
}
