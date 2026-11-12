// Package modelseries holds the two synthetic index series the end-to-end
// suite's model portfolio runs on (D-200's criterion 19): MODA and MODB,
// each a closing level in PLN per weekday, as EODHD's end-of-day API would
// serve them for MODA.INDX and MODB.INDX. The institution fixture
// (test/e2e/fixture) serves them and modelcheck works the backtest out
// from them by hand; both import this package, so the series the fixture
// serves is the series the hand calculation assumes. The levels are round
// numbers that change on a few known days and hold between them, so that
// the peaks and troughs of the series and of a basket of them are known by
// inspection. It is never released.
package modelseries

import (
	"time"
)

// End is the last day either series has a close: the Friday that ends the
// data path's IBKR week.
const End = "2026-03-13"

// isoDate is the layout of the dates in the series.
const isoDate = time.DateOnly

// Step is a level that takes effect on a weekday, and holds until the next
// step.
type Step struct {
	From, Level string
}

// Series is one index: its name in OpenFief, its EODHD ticker and its
// levels.
type Series struct {
	Index, Ticker string
	Steps         []Step
}

// Close is one weekday's closing level as text.
type Close struct {
	Date, Level string
}

// All returns the two series in the order the model portfolio lists them.
// MODA starts on Monday, December 1, 2025 and MODB two days later, so the
// backtest of a basket of both starts on Wednesday, December 3, the first
// day both have a level, and crosses the year end that rebalances it. Every
// step is on a weekday.
func All() []Series {
	return []Series{
		{Index: "MODA", Ticker: "MODA.INDX", Steps: []Step{
			{"2025-12-01", "100"},
			{"2025-12-10", "120"},
			{"2025-12-22", "96"},
			{"2025-12-31", "110"},
			{"2026-01-15", "132"},
			{"2026-02-02", "99"},
			{"2026-03-02", "121"},
		}},
		{Index: "MODB", Ticker: "MODB.INDX", Steps: []Step{
			{"2025-12-03", "50"},
			{"2025-12-12", "60"},
			{"2025-12-29", "45"},
			{"2026-01-20", "54"},
			{"2026-02-16", "27"},
			{"2026-03-09", "45"},
		}},
	}
}

// Find returns the series whose EODHD ticker is ticker.
func Find(ticker string) (Series, bool) {
	for _, s := range All() {
		if s.Ticker == ticker {
			return s, true
		}
	}
	return Series{}, false
}

// dated is a step with its day parsed.
type dated struct {
	at    time.Time
	level string
}

// weekend reports whether d is a Saturday or a Sunday.
func weekend(d time.Time) bool {
	return d.Weekday() == time.Saturday || d.Weekday() == time.Sunday
}

// outside reports whether d is before from or after to, either of which
// may be zero for no bound.
func outside(d, from, to time.Time) bool {
	return (!from.IsZero() && d.Before(from)) || (!to.IsZero() && d.After(to))
}

// Closes returns the closes on each weekday from the series' first step to
// End that falls within from and to, both included; a zero from or to is
// no bound. A level takes effect on its step's day and holds until the next
// step. A series with a step that doesn't parse has no closes.
func (s Series) Closes(from, to time.Time) []Close {
	steps, ok := s.parsed()
	if !ok {
		return nil
	}
	end, err := time.Parse(isoDate, End)
	if err != nil {
		return nil
	}
	var out []Close
	level := ""
	next := 0
	for d := steps[0].at; !d.After(end); d = d.AddDate(0, 0, 1) {
		for next < len(steps) && !steps[next].at.After(d) {
			level = steps[next].level
			next++
		}
		if !weekend(d) && !outside(d, from, to) {
			out = append(out, Close{Date: d.Format(isoDate), Level: level})
		}
	}
	return out
}

// parsed returns the steps with their days parsed, and false when one
// doesn't parse or there are none.
func (s Series) parsed() ([]dated, bool) {
	out := make([]dated, 0, len(s.Steps))
	for _, st := range s.Steps {
		at, err := time.Parse(isoDate, st.From)
		if err != nil {
			return nil, false
		}
		out = append(out, dated{at: at, level: st.Level})
	}
	return out, len(out) > 0
}
