// Command core is OpenFief's core service (docs/DESIGN.md D-2): the system of
// record. It migrates the database, stores the built-in sleeve kinds and
// sleeves, runs the River revaluation workers (D-36) and serves the read
// API (D-6, D-50) to the owner's scripts, which present script tokens
// (D-27). It also issues, lists and revokes those tokens.
//
// Usage:
//
//	OPENFIEF_DSN=postgres://... core [serve] [flags]
//	OPENFIEF_DSN=postgres://... core migrate
//	OPENFIEF_DSN=postgres://... core token issue --name NAME [--ttl 720h]
//	OPENFIEF_DSN=postgres://... core token list
//	OPENFIEF_DSN=postgres://... core token revoke ID
//	OPENFIEF_DSN=postgres://... core connector bind --instance NAME --account ID
//	OPENFIEF_DSN=postgres://... core connector sync --instance NAME [--mode opening|transactions]
//	OPENFIEF_DSN=postgres://... core enroll-link [--host localhost:8080] [--ttl 15m]
//	OPENFIEF_DSN=postgres://... core check balances [--all]
//	core init-credentials
//
// `token issue` prints the new token's secret once, on stdout, and nowhere
// else; it is the break-glass way to a script token, which the owner
// otherwise creates from a logged-in session. `enroll-link` prints a
// one-time link for creating a passkey on a host (D-60), localhost by
// default: the way in for the first passkey. `check balances` writes the
// owner's balance check (D-200, criterion 1) as JSON on stdout, for
// `openfief check balances`, which runs it through Kubernetes exec with
// the owner's kubeconfig; it logs only that the check ran, with counts.
//
// `init-credentials` is the install's one-shot Job (D-216), run from this
// image with its own ServiceAccount, which may only create Secrets: it makes
// SpiceDB's preshared key and its database role's password, each only if its
// Secret is missing, and never reads, changes or rotates one. It needs no
// database and logs Secret names and "created" or "kept", never a value.
//
// Next to the read API, serve offers view plugins the AuthService
// (proto/openfief/identity/v1): passkey login and registration, sessions,
// script tokens and enrollment links. OPENFIEF_AUTH_ORIGINS lists the
// origins the owner's browser may use, separated by commas, and enroll-link
// refuses a link for any other (D-126); --session-idle and --session-max bound sessions;
// --demo-sessions issues read-only sessions without credentials on a demo
// database. The service logs internal IDs and counts only: never an amount, a
// holding, a token or its hash.
//
// In a Kubernetes cluster, serve also syncs the running connector
// instances on a schedule (D-49), presenting a TokenRequest token of its own
// ServiceAccount whose audience is the instance (D-59). connector bind lets
// an instance write to an account; connector sync syncs one instance now
// and needs the cluster too. Outside a cluster serve doesn't sync.
//
// serve and the connector commands trace their requests, connector calls,
// queries, revaluation jobs and syncs with OpenTelemetry when the standard
// variables name an OTLP endpoint (OTEL_EXPORTER_OTLP_ENDPOINT and its
// kin, docs/guides/tracing.md); without one nothing is recorded or sent.
// Spans name operations only: never an amount, a holding, an account
// number, a token or a query's arguments.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

const (
	// exitUsage is the exit status for bad arguments.
	exitUsage = 2
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	cmd, err := parseArgs(os.Args[1:], os.Getenv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "core:", err)
		os.Exit(exitUsage)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	err = run(ctx, &cmd, os.Stdout, log)
	stop()
	if err != nil {
		log.Error("core failed", "command", cmd.Name, "error", err)
		os.Exit(1)
	}
}

// runCheck runs the check command cmd names over d's database: check
// balances or check ledger.
func runCheck(ctx context.Context, d *deps, cmd *command, out io.Writer, log *slog.Logger) error {
	if cmd.Name == cmdCheckLedger {
		return checkLedger(ctx, d.pool, cmd.LedgerAccount, out, log)
	}
	return checkBalances(ctx, d.pool, cmd.AllChecks, out, log)
}

// withMigratedPool connects to dsn, migrates, and runs fn with the pool.
func withMigratedPool(ctx context.Context, dsn string, fn func(d *deps) error) error {
	return withPool(ctx, dsn, func(d *deps) error {
		if err := migrate(ctx, d.pool); err != nil {
			return err
		}
		return fn(d)
	})
}

// runTokenCommand runs a token command over a migrated database, with the
// connection to SpiceDB it asked for, if any, which it closes at the end.
func runTokenCommand(ctx context.Context, cmd *command, out io.Writer, log *slog.Logger) error {
	return withMigratedPool(ctx, cmd.DSN, func(d *deps) error {
		tokens, closer, err := newTokens(d.pool, &cmd.Authz, cmd.SpiceDBWait, time.Now)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := closer.Close(); closeErr != nil {
				log.Warn("closing the connection to SpiceDB failed", "error", kernel.LogSafe(closeErr))
			}
		}()
		return tokenCommand(ctx, cmd, tokens, out, time.Now())
	})
}

// run runs cmd, writing what a token command shows to out.
func run(ctx context.Context, cmd *command, out io.Writer, log *slog.Logger) error {
	switch cmd.Name {
	case cmdServe:
		return withTracing(ctx, func() error { return serve(ctx, cmd, log, time.Now) })
	case cmdMigrate:
		return withPool(ctx, cmd.DSN, func(d *deps) error {
			if err := migrate(ctx, d.pool); err != nil {
				return err
			}
			log.Info("database migrated")
			return nil
		})
	case cmdTokenIssue, cmdTokenList, cmdTokenRevoke:
		return runTokenCommand(ctx, cmd, out, log)
	case cmdInitCredentials:
		return initCredentialsInCluster(ctx, log)
	case cmdEnrollLink:
		return withMigratedPool(ctx, cmd.DSN, func(d *deps) error {
			return enrollLink(ctx, d.pool, &cmd.Enroll, out, time.Now())
		})
	case cmdCheck, cmdCheckLedger:
		return withPool(ctx, cmd.DSN, func(d *deps) error { return runCheck(ctx, d, cmd, out, log) })
	case cmdBind, cmdSync:
		cl := clusterOrNil(log)
		return withTracing(ctx, func() error {
			return withConnectorDeps(ctx, cmd.DSN, log, func(d *connectorDeps) error {
				return connectorCommand(ctx, cmd, d, cl, out)
			})
		})
	}
	return fmt.Errorf("%w: %q", errUsage, cmd.Name)
}
