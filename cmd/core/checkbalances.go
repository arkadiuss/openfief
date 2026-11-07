package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
	viewscli "github.com/arkadiuss/openfief/internal/views/cli"
)

// checkCallerID names the owner as `core check balances` acts for them:
// whoever can run the core's binary in its pod holds the owner's
// kubeconfig, the same authority `openfief enroll` runs enroll-link with
// (D-80).
const checkCallerID = "core-check-balances"

// checkBalances writes the owner's balance check (D-200, criterion 1) over
// pool to out as JSON (viewscli.WriteJSON): the latest balance assertion of
// each position and cash balance, or every one with all, of the accounts
// the authorizer lets the owner view (D-163). It logs that the check ran,
// with its counts only, never a value (GR-12).
func checkBalances(ctx context.Context, pool *pgxpool.Pool, all bool, out io.Writer, log *slog.Logger) error {
	vstore, err := valuationpg.NewStore(pool)
	if err != nil {
		return fmt.Errorf("valuation store: %w", err)
	}
	views, err := viewsapp.New(ledgerapp.New(ledgerpg.New(pool)), vstore, vstore, vstore,
		jurisdictionPack().ReportingCurrency(), jurisdictionPack().ReportingCurrency(),
		viewsapp.WithAuthorizer(identityapp.Policy{}), viewsapp.WithSyncHistory(ingestpg.New(pool)))
	if err != nil {
		return fmt.Errorf("views: %w", err)
	}
	report, err := views.BalanceReport(identityapp.WithCaller(ctx, identity.Caller{Holder: identity.HolderOwner, ID: checkCallerID}), all)
	if err != nil {
		return fmt.Errorf("checking balances: %w", err)
	}
	if err := viewscli.WriteJSON(out, &report); err != nil {
		return fmt.Errorf("check balances: %w", err)
	}
	log.Info("balance check ran", "all", all, "checked", len(report.Checks), "failed", report.Failed(),
		"unsupportedInstances", len(report.Unsupported))
	return nil
}
