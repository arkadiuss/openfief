package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ChartStarts reads the charts' start the owner stored as a setting
// (D-167). *valuation/postgres.Store implements it.
type ChartStarts interface {
	// ChartStart returns the start stored for account's charts, or for the
	// portfolio's when account is empty, and the zero date when none is
	// stored: the charts start on their default then (D-138).
	ChartStart(ctx context.Context, account ledger.AccountID) (kernel.Date, error)
}

// WithChartStarts has the time charts start on the start the owner stored
// in starts when a query names none (D-167); without it they start on
// their default (D-138).
func WithChartStarts(starts ChartStarts) Option {
	return func(v *Views) { v.starts = starts }
}

// storedStart returns the charts' start stored for account, or for the
// portfolio when account is empty, and the zero date when none is stored
// or the views read no settings. The caller has checked that it may view
// account.
func (v *Views) storedStart(ctx context.Context, account ledger.AccountID) (kernel.Date, error) {
	if v.starts == nil {
		return kernel.Date{}, nil
	}
	d, err := v.starts.ChartStart(ctx, account)
	if err != nil {
		return kernel.Date{}, fmt.Errorf("loading the charts' start: %w", err)
	}
	return d, nil
}
