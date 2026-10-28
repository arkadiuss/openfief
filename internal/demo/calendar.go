package demo

import (
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// period is one calendar month of the demo's history, clipped to the days
// after go-live and up to today: the period of one IBKR activity statement.
type period struct {
	// index counts the months from go-live's, which is 0.
	index int
	year  int
	month time.Month
	from  kernel.Date
	to    kernel.Date
}

// periods returns the months from go-live's to today's, each clipped to
// (goLive, today]; a go-live on a month's last day gives that month no days,
// so it is left out.
func periods(goLive, today kernel.Date) []period {
	var out []period
	first := goLive.AddDays(1)
	for i, d := 0, first; !d.After(today); i++ {
		end := lastOfMonth(d)
		if end.After(today) {
			end = today
		}
		out = append(out, period{index: i, year: d.Year(), month: d.Month(), from: d, to: end})
		d = end.AddDays(1)
	}
	return out
}

// lastOfMonth returns the last day of d's month.
func lastOfMonth(d kernel.Date) kernel.Date {
	t := time.Date(d.Year(), d.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	return dateOf(t)
}

// dateOf returns t's calendar date.
func dateOf(t time.Time) kernel.Date {
	d, err := kernel.NewDate(t.Year(), t.Month(), t.Day())
	if err != nil {
		return kernel.Date{}
	}
	return d
}

// businessDay returns the n-th business day (1-based) of p's calendar month,
// and false when it falls outside p's days.
func (p *period) businessDay(n int) (kernel.Date, bool) {
	d := dateOf(time.Date(p.year, p.month, 1, 0, 0, 0, 0, time.UTC))
	for seen := 0; d.Month() == p.month; d = d.AddDays(1) {
		if !isBusinessDay(d) {
			continue
		}
		seen++
		if seen == n {
			return d, !d.Before(p.from) && !d.After(p.to)
		}
	}
	return kernel.Date{}, false
}
