package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/valuation"
)

// Errors of the owner's checks, which the Owners port wraps.
var (
	// ErrNoSession reports a call without a valid session.
	ErrNoSession = errors.New("no valid session")
	// ErrNotOwner reports a session that is not the owner's, owner-scoped,
	// on an allowed origin.
	ErrNotOwner = errors.New("not the owner")
)

// Owners checks that a session is the owner's (D-66, D-79); the identity
// context answers it in cmd/core, with errors wrapping ErrNoSession or
// ErrNotOwner.
type Owners interface {
	CheckOwner(ctx context.Context, origin, sessionToken string) error
}

// RankingStore keeps the market-data sources and the owner's rankings of
// them (D-94). *postgres.Store implements it.
type RankingStore interface {
	// SaveSources records sources, each with the institution it named.
	SaveSources(ctx context.Context, sources []valuation.Source) error
	// SourceRanking returns the ranking of kind's sources.
	SourceRanking(ctx context.Context, kind valuation.MarketDataKind) (valuation.Ranking, error)
	// SaveSourceRanking replaces the owner's order of kind's sources and,
	// only when it changed, enqueues revals in the same transaction; it
	// reports whether it changed.
	SaveSourceRanking(ctx context.Context, kind valuation.MarketDataKind, placed []string, revals []valuation.Revaluation) (bool, error)
}

// Rankings runs the owner's ranking of market-data sources (D-94): for
// each kind of market data, the highest-ranked source with a value wins,
// every source's values are stored, and a change of ranking revalues
// every account (D-36).
type Rankings struct {
	planner *Planner
	store   RankingStore
	owners  Owners
}

// NewRankings returns the ranking use cases over planner and store, with
// the owner's sessions checked by owners.
func NewRankings(planner *Planner, store RankingStore, owners Owners) *Rankings {
	return &Rankings{planner: planner, store: store, owners: owners}
}

// SourceRankings returns the ranking of every kind, in the order of
// valuation.MarketDataKinds, to the owner: a session of owner scope on
// origin.
func (r *Rankings) SourceRankings(ctx context.Context, origin, sessionToken string) ([]valuation.Ranking, error) {
	if err := r.checkOwner(ctx, origin, sessionToken); err != nil {
		return nil, err
	}
	out := make([]valuation.Ranking, 0, len(valuation.MarketDataKinds()))
	for _, kind := range valuation.MarketDataKinds() {
		ranking, err := r.store.SourceRanking(ctx, kind)
		if err != nil {
			return nil, fmt.Errorf("loading the source rankings: %w", err)
		}
		out = append(out, ranking)
	}
	return out, nil
}

// RankSources makes placed the owner's order of kind's sources, for the
// owner, and returns the ranking that results. A ranking that breaks
// valuation.ValidateRanking is an error wrapping ErrInvalidInput and
// nothing changes. When the order changed, every account is revalued from
// its first transaction in the same transaction (D-36), since the winning
// value of any date may have changed.
func (r *Rankings) RankSources(ctx context.Context, origin, sessionToken string, kind valuation.MarketDataKind, placed []string) (valuation.Ranking, error) {
	if err := r.checkOwner(ctx, origin, sessionToken); err != nil {
		return valuation.Ranking{}, err
	}
	if err := valuation.ValidateRanking(kind, placed); err != nil {
		return valuation.Ranking{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	revals, err := r.planner.FromStart(ctx, "")
	if err != nil {
		return valuation.Ranking{}, fmt.Errorf("ranking the sources of %s: %w", kind, err)
	}
	if _, saveErr := r.store.SaveSourceRanking(ctx, kind, placed, revals); saveErr != nil {
		return valuation.Ranking{}, fmt.Errorf("ranking the sources of %s: %w", kind, saveErr)
	}
	ranking, err := r.store.SourceRanking(ctx, kind)
	if err != nil {
		return valuation.Ranking{}, fmt.Errorf("ranking the sources of %s: %w", kind, err)
	}
	return ranking, nil
}

// RecordSources records the sources a delivery's values come from, each
// with the institution it named, which places a source by default. It
// needs no session: the core calls it for connectors' deliveries. A source
// of an unknown kind or with an invalid name is an error wrapping
// ErrInvalidInput, and nothing is recorded.
func (r *Rankings) RecordSources(ctx context.Context, sources []valuation.Source) error {
	for _, s := range sources {
		if err := valuation.ValidateRanking(s.Kind, []string{s.Name}); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
	}
	if len(sources) == 0 {
		return nil
	}
	if err := r.store.SaveSources(ctx, sources); err != nil {
		return fmt.Errorf("recording market-data sources: %w", err)
	}
	return nil
}

// checkOwner checks that sessionToken is the owner's on origin.
func (r *Rankings) checkOwner(ctx context.Context, origin, sessionToken string) error {
	if err := r.owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return fmt.Errorf("checking the session: %w", err)
	}
	return nil
}
