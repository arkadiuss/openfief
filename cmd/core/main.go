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
//
// `token issue` prints the new token's secret once, on stdout, and nowhere
// else. The service logs internal IDs and counts only: never an amount, a
// holding, a token or its hash.
//
// Not built yet: the core calling connectors (D-49), with their schedules,
// retries and import batches, and the per-instance tokens it presents to
// them (D-59). No connector plugin exists yet; data enters through the
// demo loader (cmd/demo-seed) until then.
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
	}
	return fmt.Errorf("%w: %q", errUsage, cmd.Name)
}
