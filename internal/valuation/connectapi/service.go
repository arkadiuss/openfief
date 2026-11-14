// Package connectapi serves the valuation context's MarketDataService
// (proto/openfief/valuation/v1/market_data.proto) to view plugins over
// Connect: the built-in UI hands the owner's ranking of market-data
// sources, each series' source and the gap thresholds to the core through
// it (D-94, D-79). Callers are admitted by the
// interceptor the caller passes, cmd/core's view-plugin check; the owner's
// session travels in the request and the use cases check its scope. Errors
// map to Connect codes: no valid session is unauthenticated, a session
// that is not the owner's is permission_denied, a ranking breaking a rule
// is invalid_argument with the rule it broke, anything else internal with
// nothing of the error.
package connectapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
)

// maxRequestBytes caps one request message before the caller is checked
// (D-184): the owner services take a few fields, never a file.
const maxRequestBytes = 1 << 20

var (
	// errInternal is what a caller sees of an internal failure.
	errInternal = errors.New("internal error")
	// errUnknownKind reports an unspecified kind of market data.
	errUnknownKind = errors.New("unknown kind of market data")
)

// New returns the MarketDataService's path and HTTP handler over rankings,
// with callers in front of every call to admit view plugins only, logging
// to log.
func New(rankings *app.Rankings, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return valuationv1connect.NewMarketDataServiceHandler(&service{rankings: rankings, log: log},
		connect.WithInterceptors(callers), connect.WithReadMaxBytes(maxRequestBytes))
}

// service implements valuationv1connect.MarketDataServiceHandler.
type service struct {
	rankings *app.Rankings
	log      *slog.Logger
}

// ListSourceRankings returns every kind's ranking to the owner.
func (s *service) ListSourceRankings(ctx context.Context, req *connect.Request[valuationv1.ListSourceRankingsRequest]) (*connect.Response[valuationv1.ListSourceRankingsResponse], error) {
	rankings, err := s.rankings.SourceRankings(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &valuationv1.ListSourceRankingsResponse{Rankings: make([]*valuationv1.SourceRanking, 0, len(rankings))}
	for i := range rankings {
		out.Rankings = append(out.Rankings, pbRanking(&rankings[i]))
	}
	return connect.NewResponse(out), nil
}

// SetSourceRanking replaces the owner's order of one kind's sources.
func (s *service) SetSourceRanking(ctx context.Context, req *connect.Request[valuationv1.SetSourceRankingRequest]) (*connect.Response[valuationv1.SetSourceRankingResponse], error) {
	kind, err := kindFrom(req.Msg.GetKind())
	if err != nil {
		return nil, s.fail(ctx, fmt.Errorf("%w: %w", app.ErrInvalidInput, err))
	}
	ranking, err := s.rankings.RankSources(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), kind, req.Msg.GetSources())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&valuationv1.SetSourceRankingResponse{Ranking: pbRanking(&ranking)}), nil
}

// ListSeriesSources returns every series that has a source to the owner.
func (s *service) ListSeriesSources(ctx context.Context, req *connect.Request[valuationv1.ListSeriesSourcesRequest]) (*connect.Response[valuationv1.ListSeriesSourcesResponse], error) {
	all, err := s.rankings.SeriesSources(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &valuationv1.ListSeriesSourcesResponse{Series: make([]*valuationv1.SeriesSource, 0, len(all))}
	for i := range all {
		ss := &all[i]
		out.Series = append(out.Series, &valuationv1.SeriesSource{
			Kind: kindTo(ss.Series.Kind), Key: ss.Series.Key, Source: ss.Source, ByOwner: ss.ByOwner, Available: ss.Available,
		})
	}
	return connect.NewResponse(out), nil
}

// SetSeriesSource makes one source the owner's choice for one series.
func (s *service) SetSeriesSource(ctx context.Context, req *connect.Request[valuationv1.SetSeriesSourceRequest]) (*connect.Response[valuationv1.SetSeriesSourceResponse], error) {
	kind, err := kindFrom(req.Msg.GetKind())
	if err != nil {
		return nil, s.fail(ctx, fmt.Errorf("%w: %w", app.ErrInvalidInput, err))
	}
	series := valuation.Series{Kind: kind, Key: req.Msg.GetKey()}
	if err := s.rankings.SetSeriesSource(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), series, req.Msg.GetSource()); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&valuationv1.SetSeriesSourceResponse{}), nil
}

// SetGapThreshold sets one kind's gap threshold.
func (s *service) SetGapThreshold(ctx context.Context, req *connect.Request[valuationv1.SetGapThresholdRequest]) (*connect.Response[valuationv1.SetGapThresholdResponse], error) {
	kind, err := kindFrom(req.Msg.GetKind())
	if err != nil {
		return nil, s.fail(ctx, fmt.Errorf("%w: %w", app.ErrInvalidInput, err))
	}
	if err := s.rankings.SetGapThreshold(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), kind, int(req.Msg.GetDays())); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&valuationv1.SetGapThresholdResponse{}), nil
}

// fail maps err to its Connect error. Only a broken rule carries its text,
// which names the rule and the source, so the owner can fix it.
func (s *service) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, kernel.ErrAuditUnavailable):
		return connect.NewError(connect.CodeUnavailable, kernel.ErrAuditUnavailable)
	case errors.Is(err, app.ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "market data call failed", slog.String("error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded)))
	return connect.NewError(connect.CodeInternal, errInternal)
}

// pbRanking converts a ranking to its message.
func pbRanking(r *valuation.Ranking) *valuationv1.SourceRanking {
	out := &valuationv1.SourceRanking{
		Kind: kindTo(r.Kind), Sources: make([]*valuationv1.RankedSource, 0, len(r.Sources)), GapThresholdDays: days32(r.GapThreshold),
	}
	for _, s := range r.Sources {
		out.Sources = append(out.Sources, &valuationv1.RankedSource{Name: s.Name, Institution: s.Institution, Placed: s.Placed})
	}
	return out
}

// days32 is days as the message's int32: a gap threshold is at most 3660
// days (valuation.ValidateGapThreshold), and one out of range is 0.
func days32(days int) int32 {
	if days < 0 || days > math.MaxInt32 {
		return 0
	}
	return int32(days)
}

// kindTo is the message's kind of k.
func kindTo(k valuation.MarketDataKind) valuationv1.MarketDataKind {
	switch k {
	case valuation.KindPrices:
		return valuationv1.MarketDataKind_MARKET_DATA_KIND_PRICES
	case valuation.KindFXRates:
		return valuationv1.MarketDataKind_MARKET_DATA_KIND_FX_RATES
	case valuation.KindIndexReadings:
		return valuationv1.MarketDataKind_MARKET_DATA_KIND_INDEX_READINGS
	}
	return valuationv1.MarketDataKind_MARKET_DATA_KIND_UNSPECIFIED
}

// kindFrom reads a message's kind; unspecified is errUnknownKind.
func kindFrom(k valuationv1.MarketDataKind) (valuation.MarketDataKind, error) {
	switch k {
	case valuationv1.MarketDataKind_MARKET_DATA_KIND_PRICES:
		return valuation.KindPrices, nil
	case valuationv1.MarketDataKind_MARKET_DATA_KIND_FX_RATES:
		return valuation.KindFXRates, nil
	case valuationv1.MarketDataKind_MARKET_DATA_KIND_INDEX_READINGS:
		return valuation.KindIndexReadings, nil
	case valuationv1.MarketDataKind_MARKET_DATA_KIND_UNSPECIFIED:
	}
	return 0, fmt.Errorf("%w: %s", errUnknownKind, k)
}
