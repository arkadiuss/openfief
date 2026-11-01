// Command xtb is OpenFief's first-party XTB connector plugin (D-9, D-49):
// the core calls its ImportFile with an xStation "Historia operacji" export
// the owner uploaded (R-24), one account per file, and it maps the file in
// the transactions mode the call names. It has no egress and no Secret: it
// reads only the uploaded file. The parser and mapping are
// internal/ingest/xtb; this binary only wires them to the SDK.
//
// The export names instruments by ticker only, so the plugin delivers
// stock and ETF rows as ticker transactions, which the core books once the
// owner confirms each ticker's ISIN (D-116). It names every entry by the
// XTB account of the file and reports that account, so the core maps it to
// the OpenFief account the owner picked (D-72). It never logs a file's
// contents, and never reads a deposit's or withdrawal's comment, which name
// the sending bank and the XTB account.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

func main() {
	os.Exit(serve())
}

// serve runs the plugin until SIGTERM or an interrupt and returns its exit
// code.
func serve() int {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	if err := run(ctx, log); err != nil {
		log.Error("xtb stopped", "error", err.Error())
		return 1
	}
	return 0
}

// run serves the connector until ctx is done.
func run(ctx context.Context, log *slog.Logger) error {
	log.Info("xtb serving")
	if err := openfief.ServeConnector(ctx, &connector{}); err != nil {
		return fmt.Errorf("xtb: %w", err)
	}
	return nil
}
