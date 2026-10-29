package portfolio

import (
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Drawdown is the worst fall of a series from a peak to a later trough
// (glossary: Max drawdown): the peak's and the trough's dates and values,
// and the fall as a fraction of the peak, 0.25 for 25%. A series that never
// falls has a zero Depth, with its first day as both peak and trough.
type Drawdown struct {
	Peak, Trough           kernel.Date
	PeakValue, TroughValue kernel.Decimal
	Depth                  kernel.Decimal
}

// MaxDrawdown returns the max drawdown of points, which are in date order
// and positive (D-54): over every pair of a day and a later one, the
// largest fall from the first's value to the second's as a fraction of the
// first's, rounded half-even to BacktestPlaces. Of equal falls it keeps the
// earliest. No points have the zero Drawdown.
func MaxDrawdown(points []Point) (Drawdown, error) {
	if len(points) == 0 {
		return Drawdown{}, nil
	}
	peak := points[0]
	best := Drawdown{Peak: peak.Date, Trough: peak.Date, PeakValue: peak.Value, TroughValue: peak.Value}
	for i, p := range points {
		if p.Value.Sign() <= 0 {
			return Drawdown{}, fmt.Errorf("%w: a value of %s on %s is not positive", ErrInvalidHistory, p.Value, p.Date)
		}
		if i > 0 && !p.Date.After(points[i-1].Date) {
			return Drawdown{}, fmt.Errorf("%w: %s follows %s", ErrInvalidHistory, p.Date, points[i-1].Date)
		}
		if p.Value.Cmp(peak.Value) > 0 {
			peak = p
			continue
		}
		// (peak - p) / peak > (best peak - best trough) / best peak, without
		// dividing: both peaks are positive.
		fall := peak.Value.Sub(p.Value).Mul(best.PeakValue)
		if fall.Cmp(best.PeakValue.Sub(best.TroughValue).Mul(peak.Value)) > 0 {
			best = Drawdown{Peak: peak.Date, Trough: p.Date, PeakValue: peak.Value, TroughValue: p.Value}
		}
	}
	depth, err := best.PeakValue.Sub(best.TroughValue).Quo(best.PeakValue, BacktestPlaces, BacktestRounding)
	if err != nil {
		return Drawdown{}, fmt.Errorf("max drawdown: %w", err)
	}
	best.Depth = depth
	return best, nil
}

// Since returns points from the first one dated on or after from, rebased
// so that it is 100: each value times 100 divided by that point's value,
// rounded half-even to BacktestPlaces. The model's line in the
// value-over-time chart starts at 100 on go-live this way (D-53). Points in
// date order with none on or after from, or a first one that is not
// positive, give none.
func Since(points []Point, from kernel.Date) ([]Point, error) {
	i := 0
	for i < len(points) && points[i].Date.Before(from) {
		i++
	}
	if i == len(points) || points[i].Value.Sign() <= 0 {
		return nil, nil
	}
	base := points[i].Value
	out := make([]Point, 0, len(points)-i)
	for _, p := range points[i:] {
		v, err := p.Value.Mul(kernel.NewDecimal(indexBase, 0)).Quo(base, BacktestPlaces, BacktestRounding)
		if err != nil {
			return nil, fmt.Errorf("rebasing: %w", err)
		}
		out = append(out, Point{Date: p.Date, Value: v})
	}
	return out, nil
}
