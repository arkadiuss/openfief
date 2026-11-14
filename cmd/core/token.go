package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
)

// tokenAdmin is the script token use cases of the core's own commands. It
// issues and revokes as the admin actor, so each change is in the audit log
// with the change itself (D-211).
type tokenAdmin struct {
	store  identityapp.AuthStore
	tokens *identityapp.Tokens
	// grants ends a revoked token's grants, and gives a new token the
	// household grant when asked (D-209).
	grants *identityapp.Grants
	// ping asks SpiceDB whether it answers; nil with the policy, which has
	// nothing to wait for. wait is how long Issue waits for it, asking every
	// pollEvery.
	ping      func(context.Context) error
	wait      time.Duration
	pollEvery time.Duration
}

// errSpiceDBSilent reports a SpiceDB that did not answer in time, so token
// issue made nothing.
var errSpiceDBSilent = errors.New("SpiceDB did not answer, so no token was issued")

// spicedbPollEvery is how often token issue asks a SpiceDB that has not
// answered yet.
const spicedbPollEvery = 2 * time.Second

// systemActor is the core, as the audit log names it.
func systemActor() identity.Actor {
	return identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}
}

// Issue creates a token and, when household is set, the token's household
// grant (D-209), atomically as far as the owner can tell. A token is given no
// grant otherwise: under SpiceDB it then views nothing until the owner grants
// it something, and nothing of SpiceDB is needed. With the grant under
// SpiceDB, Issue first waits, up to the wait it was given, for SpiceDB to
// answer, and issues nothing if it doesn't. Then the token, its grant and
// their audit entries are stored in one transaction and the grant is written
// to SpiceDB; if that write still fails, the token is revoked, with an audit
// entry, no secret is returned and the error says so
// (identityapp.Grants.IssueToken).
func (a tokenAdmin) Issue(ctx context.Context, name string, ttl time.Duration, household bool) (identity.ScriptToken, string, error) {
	if household && a.ping != nil {
		if err := waitForSpiceDB(ctx, a.ping, a.wait, a.pollEvery); err != nil {
			return identity.ScriptToken{}, "", err
		}
	}
	tok, secret, err := a.grants.IssueToken(ctx, adminActor(), name, ttl, household)
	if err != nil {
		return identity.ScriptToken{}, "", fmt.Errorf("issuing as the administrator: %w", err)
	}
	return tok, secret, nil
}

// waitForSpiceDB asks ping until it answers, every poll, for up to wait, and
// returns an error wrapping errSpiceDBSilent, with the last answer, when it
// never does. A wait of zero asks once.
func waitForSpiceDB(ctx context.Context, ping func(context.Context) error, wait, poll time.Duration) error {
	expired := time.After(wait)
	for {
		err := ping(ctx)
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %w", errSpiceDBSilent, ctx.Err())
		case <-expired:
			return fmt.Errorf("%w (waited %s): %w", errSpiceDBSilent, wait, err)
		case <-time.After(poll):
		}
	}
}

// Revoke revokes a token, ends its grants and writes the audit entries.
func (a tokenAdmin) Revoke(ctx context.Context, id identity.ScriptTokenID) error {
	if err := a.grants.RevokeToken(ctx, adminActor(), id); err != nil {
		return fmt.Errorf("revoking as the administrator: %w", err)
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
		tok, secret, err := tokens.Issue(ctx, cmd.TokenName, cmd.TTL, cmd.GrantHousehold)
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
