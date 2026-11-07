package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
)

// ledgerReport is what `core check ledger` writes as JSON: the whole-log
// check (ledger.CheckLog, D-200 criterion 2), and with --account that
// account's stored versions.
type ledgerReport struct {
	// Balanced is true when every version of every transaction sums to
	// zero in every currency, and so the whole log does.
	Balanced     bool                 `json:"balanced"`
	Transactions int                  `json:"transactions"`
	Versions     int                  `json:"versions"`
	Superseded   int                  `json:"superseded"`
	Accounts     []ledgerAccountCount `json:"accounts"`
	Unbalanced   []ledgerUnbalanced   `json:"unbalanced"`
	Sums         []ledgerAmount       `json:"sums"`
	Logged       []ledgerVersion      `json:"logged,omitempty"`
}

// ledgerAccountCount is one account's part of the log.
type ledgerAccountCount struct {
	Account      string `json:"account"`
	Transactions int    `json:"transactions"`
	Versions     int    `json:"versions"`
	Superseded   int    `json:"superseded"`
}

// ledgerUnbalanced is a version whose postings sum to its amount, not
// zero, in its currency.
type ledgerUnbalanced struct {
	Account    string `json:"account"`
	Source     string `json:"source"`
	ExternalID string `json:"externalId"`
	Version    int    `json:"version"`
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
}

// ledgerAmount is an exact decimal amount in a currency.
type ledgerAmount struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// ledgerVersion is one stored version of a transaction, without its
// postings.
type ledgerVersion struct {
	Source     string `json:"source"`
	ExternalID string `json:"externalId"`
	Version    int    `json:"version"`
	Kind       string `json:"kind"`
	Superseded bool   `json:"superseded"`
}

// checkLedger writes the whole-log check over pool to out as JSON, with
// the stored versions of account when it isn't empty. Whoever can run the
// core's binary in its pod holds the owner's kubeconfig, as for `core check
// balances`. It logs that the check ran, with counts only, never a key or
// an amount (GR-12).
func checkLedger(ctx context.Context, pool *pgxpool.Pool, account ledger.AccountID, out io.Writer, log *slog.Logger) error {
	c, err := ledgerapp.New(ledgerpg.New(pool)).CheckLog(ctx)
	if err != nil {
		return fmt.Errorf("checking the ledger: %w", err)
	}
	report := ledgerReportOf(&c, account)
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&report); err != nil {
		return fmt.Errorf("check ledger: %w", err)
	}
	log.Info("ledger check ran", "balanced", report.Balanced, "accounts", len(report.Accounts),
		"transactions", report.Transactions, "versions", report.Versions, "superseded", report.Superseded,
		"unbalanced", len(report.Unbalanced))
	return nil
}

// ledgerReportOf writes c as the report, with account's stored versions
// when account isn't empty.
func ledgerReportOf(c *ledger.LogCheck, account ledger.AccountID) ledgerReport {
	r := ledgerReport{
		Balanced: c.Balanced(), Transactions: c.Transactions, Versions: c.Versions, Superseded: c.Superseded,
		Accounts: []ledgerAccountCount{}, Unbalanced: []ledgerUnbalanced{}, Sums: []ledgerAmount{},
	}
	for _, a := range c.Accounts {
		r.Accounts = append(r.Accounts, ledgerAccountCount{
			Account: string(a.Account), Transactions: a.Transactions, Versions: a.Versions, Superseded: a.Superseded,
		})
	}
	for _, u := range c.Unbalanced {
		r.Unbalanced = append(r.Unbalanced, ledgerUnbalanced{
			Account: string(u.Account), Source: u.Key.Source, ExternalID: u.Key.ExternalID, Version: u.Version,
			Amount: u.Sum.Amount().String(), Currency: u.Sum.Currency().String(),
		})
	}
	for _, s := range c.Sums {
		r.Sums = append(r.Sums, ledgerAmount{Amount: s.Amount().String(), Currency: s.Currency().String()})
	}
	for _, v := range c.Logged {
		if account != "" && v.Account == account {
			r.Logged = append(r.Logged, ledgerVersion{
				Source: v.Key.Source, ExternalID: v.Key.ExternalID, Version: v.Version, Kind: v.Kind.String(), Superseded: v.Superseded,
			})
		}
	}
	return r
}
