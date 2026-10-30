package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Errors about broker accounts (D-72).
var (
	// ErrNoBrokerAccountStore reports a delivery that names broker accounts
	// to a Connectors that has nowhere to resolve them.
	ErrNoBrokerAccountStore = errors.New("no store for broker accounts")
	// ErrUnknownBrokerAccount reports a broker account the connector
	// instance never reported.
	ErrUnknownBrokerAccount = errors.New("the connector instance never reported the broker account")
	// ErrBrokerAccountMapped reports a broker account already mapped to
	// another OpenFief account.
	ErrBrokerAccountMapped = errors.New("the broker account is mapped to another account")
)

// BrokerAccountStore keeps the broker accounts each connector instance
// reported and the OpenFief account each is mapped to (D-72).
type BrokerAccountStore interface {
	// RecordBrokerAccounts records that a delivery of instance from
	// institution reported ids on on: each is stored, unmapped when new,
	// with on as the day it was last seen, and of instance's broker
	// accounts only these are in its last delivery.
	RecordBrokerAccounts(ctx context.Context, instance string, institution ingest.InstitutionID, ids []ingest.BrokerAccountID, on kernel.Date) error
	// BrokerAccounts returns instance's broker accounts, in ID order.
	BrokerAccounts(ctx context.Context, instance string) ([]ingest.BrokerAccount, error)
	// MarkSeeded records that an opening batch of instance completed after
	// ids were mapped.
	MarkSeeded(ctx context.Context, instance string, ids []ingest.BrokerAccountID) error
	// MapBrokerAccount maps instance's broker account id to account at at.
	// Mapping it again to the same account changes nothing; a broker
	// account instance never reported fails with ErrUnknownBrokerAccount,
	// one mapped to another account with ErrBrokerAccountMapped.
	MapBrokerAccount(ctx context.Context, instance string, id ingest.BrokerAccountID, account ledger.AccountID, at time.Time) error
}

// resolveBrokerAccounts records the broker accounts d reports for instance
// and returns d's mapping with the entries it names by broker account
// resolved into the OpenFief accounts they are mapped to, the broker
// accounts left unmapped and those an opening batch of it seeds
// (ingest.ResolveBrokerAccounts). A delivery that reports none keeps its
// mapping.
func (c *Connectors) resolveBrokerAccounts(ctx context.Context, instance string, mode ingest.Mode, d *Delivery) (ingest.Resolution, error) {
	reported := ingest.ReportedBrokerAccounts(&d.Mapping)
	if len(reported) == 0 {
		return ingest.Resolution{Mapping: d.Mapping}, nil
	}
	if c.Brokers == nil {
		return ingest.Resolution{}, ErrNoBrokerAccountStore
	}
	if err := c.Brokers.RecordBrokerAccounts(ctx, instance, d.Institution, reported, c.Today()); err != nil {
		return ingest.Resolution{}, fmt.Errorf("recording broker accounts: %w", err)
	}
	known, err := c.Brokers.BrokerAccounts(ctx, instance)
	if err != nil {
		return ingest.Resolution{}, fmt.Errorf("loading broker accounts: %w", err)
	}
	return ingest.ResolveBrokerAccounts(&d.Mapping, known, mode), nil
}

// markSeeded records that an opening batch of instance seeded the broker
// accounts seeding; none is skipped.
func (c *Connectors) markSeeded(ctx context.Context, instance string, seeding []ingest.BrokerAccountID) error {
	if len(seeding) == 0 {
		return nil
	}
	if err := c.Brokers.MarkSeeded(ctx, instance, seeding); err != nil {
		return fmt.Errorf("marking broker accounts seeded: %w", err)
	}
	return nil
}
