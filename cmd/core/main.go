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
//
// `token issue` prints the new token's secret once, on stdout, and nowhere
// else. The service logs internal IDs and counts only: never an amount, a
// holding, a token or its hash.
//
// In a Kubernetes cluster, serve also syncs the running connector
// instances on a schedule (D-49), presenting a TokenRequest token of its own
// ServiceAccount whose audience is the instance (D-59). connector bind lets
// an instance write to an account; connector sync syncs one instance now
// and needs the cluster too. Outside a cluster serve doesn't sync.
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

// run runs cmd, writing what a token command shows to out.
func run(ctx context.Context, cmd *command, out io.Writer, log *slog.Logger) error {
	switch cmd.Name {
	case cmdServe:
		return serve(ctx, cmd, log, time.Now)
	case cmdMigrate:
		return withPool(ctx, cmd.DSN, func(d *deps) error {
			if err := migrate(ctx, d.pool); err != nil {
				return err
			}
			log.Info("database migrated")
			return nil
		})
	case cmdTokenIssue, cmdTokenList, cmdTokenRevoke:
		return withPool(ctx, cmd.DSN, func(d *deps) error {
			if err := migrate(ctx, d.pool); err != nil {
				return err
			}
			return tokenCommand(ctx, cmd, newTokens(d.pool, time.Now), out, time.Now())
		})
	case cmdBind, cmdSync:
		cl := clusterOrNil(log)
		return withConnectorDeps(ctx, cmd.DSN, log, func(d *connectorDeps) error {
			return connectorCommand(ctx, cmd, d, cl, out)
		})
	}
	return fmt.Errorf("%w: %q", errUsage, cmd.Name)
}
