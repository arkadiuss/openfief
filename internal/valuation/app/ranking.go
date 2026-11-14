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
	// SeriesSources returns every series with a source, each with the
	// sources that have a value of it.
	SeriesSources(ctx context.Context) ([]valuation.SeriesSource, error)
	// AvailableSources returns the sources with a value of series.
	AvailableSources(ctx context.Context, series valuation.Series) ([]string, error)
	// SaveSeriesSource records the owner's choice of a series' source and,
	// only when the source changed, enqueues revals in the same
	// transaction; it reports whether it changed.
	SaveSeriesSource(ctx context.Context, ss *valuation.SeriesSource, revals []valuation.Revaluation) (bool, error)
	// GapThresholds returns the owner's gap threshold of each kind that
	// has one.
	GapThresholds(ctx context.Context) (valuation.GapThresholds, error)
	// SaveGapThreshold sets kind's gap threshold.
	SaveGapThreshold(ctx context.Context, kind valuation.MarketDataKind, days int) error
}

// Rankings runs the owner's ranking of market-data sources and the source
// of each series (D-94): each instrument's prices, each currency's FX rates
// and each period index's readings take every value from one source,
// chosen from the kind's ranking when the series first needs a value,
// which the owner can change for that one series; every other source's
// values are kept as cross-checks. A change of ranking changes no series'
// source, so it revalues nothing; a change of a series' source revalues
// the accounts it affects (D-36).
type Rankings struct {
	planner *Planner
	store   RankingStore
	owners  Owners
	audit   ChangeAuditor
}

// NewRankings returns the ranking use cases over planner and store, with
// the owner's sessions checked by owners.
func NewRankings(planner *Planner, store RankingStore, owners Owners) *Rankings {
	return &Rankings{planner: planner, store: store, owners: owners}
}

// WithAudit has every change of a ranking, a series' source and a gap
// threshold entered in the audit log through a (D-211), and returns r.
func (r *Rankings) WithAudit(a ChangeAuditor) *Rankings {
	r.audit = a
	return r
}

// SourceRankings returns the ranking of every kind, in the order of
// valuation.MarketDataKinds, to the owner: a session of owner scope on
// origin.
func (r *Rankings) SourceRankings(ctx context.Context, origin, sessionToken string) ([]valuation.Ranking, error) {
	if err := r.checkOwner(ctx, origin, sessionToken); err != nil {
		return nil, err
	}
	gaps, err := r.store.GapThresholds(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading the source rankings: %w", err)
	}
	out := make([]valuation.Ranking, 0, len(valuation.MarketDataKinds()))
	for _, kind := range valuation.MarketDataKinds() {
		ranking, err := r.store.SourceRanking(ctx, kind)
		if err != nil {
			return nil, fmt.Errorf("loading the source rankings: %w", err)
		}
		ranking.GapThreshold = gaps.For(kind)
		out = append(out, ranking)
	}
	return out, nil
}

// RankSources makes placed the owner's order of kind's sources, for the
// owner, and returns the ranking that results. A ranking that breaks
// valuation.ValidateRanking is an error wrapping ErrInvalidInput and
// nothing changes. The ranking only chooses the source of a series that
// first needs a value from then on (D-94): every series keeps the source
// it has, so nothing is revalued.
func (r *Rankings) RankSources(ctx context.Context, origin, sessionToken string, kind valuation.MarketDataKind, placed []string) (valuation.Ranking, error) {
	if err := r.checkOwner(ctx, origin, sessionToken); err != nil {
		return valuation.Ranking{}, err
	}
	if err := valuation.ValidateRanking(kind, placed); err != nil {
		return valuation.Ranking{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	saveErr := auditedSetting(ctx, r.audit, origin, sessionToken, settingRankingPrefix+kind.String(), func(ctx context.Context) error {
		if _, err := r.store.SaveSourceRanking(ctx, kind, placed, nil); err != nil {
			return fmt.Errorf("saving: %w", err)
		}
		return nil
	})
	if saveErr != nil {
		return valuation.Ranking{}, fmt.Errorf("ranking the sources of %s: %w", kind, saveErr)
	}
	ranking, err := r.store.SourceRanking(ctx, kind)
	if err != nil {
		return valuation.Ranking{}, fmt.Errorf("ranking the sources of %s: %w", kind, err)
	}
	gaps, err := r.store.GapThresholds(ctx)
	if err != nil {
		return valuation.Ranking{}, fmt.Errorf("ranking the sources of %s: %w", kind, err)
	}
	ranking.GapThreshold = gaps.For(kind)
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
