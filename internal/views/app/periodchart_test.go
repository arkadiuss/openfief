package app_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// newYearFixture is one account valued every day from 2024-12-20 through
// 2025-12-31: 1000 PLN paid in and invested at the start and 500 more on
// 2025-06-10, its value moving up and down by a fixed pattern, so a period's
// return is no round number and a flow falls inside the year.
func newYearFixture(t *testing.T) *fixture {
	t.Helper()
	f := newPerformanceFixture(t)
	etf := instrument(t, isinIE, "ETF1")
	header := func(id, day string) ledger.TransactionHeader {
		return ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: id}, Account: "acc-1", Date: date(t, day)}
	}
	txs := make([]ledger.Transaction, 0, 4)
	for _, in := range []struct{ id, day, pln string }{{"1", "2024-12-20", "1000"}, {"2", "2025-06-10", "500"}} {
		deposit, err := ledger.NewDeposit(header("dep-"+in.id, in.day), money(t, in.pln, "PLN"))
		if err != nil {
			t.Fatalf("NewDeposit: %v", err)
		}
		buy, err := ledger.NewBuy(header("buy-"+in.id, in.day), etf, kernel.NewQuantity(dec(t, "10")), money(t, in.pln, "PLN"), kernel.Money{})
		if err != nil {
			t.Fatalf("NewBuy: %v", err)
		}
		txs = append(txs, deposit, buy)
	}
	f.ledger.txs = map[ledger.AccountID][]ledger.Transaction{"acc-1": txs}
	f.valuations.vals = nil
	topUp := date(t, "2025-06-10")
	for i, day := 0, date(t, "2024-12-20"); !day.After(date(t, "2025-12-31")); i, day = i+1, day.AddDays(1) {
		value, qty := int64(1000+(i*37)%101-50), "10"
		if !day.Before(topUp) {
			value, qty = value+500, "20"
		}
		v := (&valRow{"acc-1", etf, day.String(), qty, kernel.NewDecimal(value, 0).String()}).build(t)
		v.Net, v.Liquidation = v.Reporting, v.Reporting
		f.valuations.vals = append(f.valuations.vals, v)
	}
	through := date(t, "2025-12-31")
	f.statuses = fakeStatuses{"acc-1": {ValuedThrough: through}, "acc-2": {ValuedThrough: through}}
	return f
}

// TestAPeriodsCardAndChartShowTheSameReturn checks D-223 on the core's
// figures: a period's return, which the Ask box's card shows, runs from the
// close of the day before its first day, and the performance chart opened on
// that day, from a preset or from the card's link, keeps 100 on it and ends on
// an index whose distance from 100 is that return, to the two places the card
// shows. A chart opened on the period's own first day would leave that day's
// move out of it.
func TestAPeriodsCardAndChartShowTheSameReturn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		from, to    string
		wantOpening string
	}{
		{name: "2025", from: "2025-01-01", to: "2025-12-31", wantOpening: "2024-12-31"},
		{name: "November 2025", from: "2025-11-01", to: "2025-11-30", wantOpening: "2025-10-31"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			views := newYearFixture(t).views(t)
			card, err := views.PeriodChange(reader(t), date(t, tt.from), date(t, tt.to))
			if err != nil {
				t.Fatalf("PeriodChange() error = %v", err)
			}
			if got := card.Opening.String(); got != tt.wantOpening {
				t.Fatalf("PeriodChange() opening = %s, want %s: the close of the day before %s", got, tt.wantOpening, tt.from)
			}
			wantSameReturn(t, views, &card, date(t, tt.from), date(t, tt.to))
		})
	}
}

// wantSameReturn fails unless the chart opened on the period's opening close,
// from the preset's date and from the card link's, returns what card does,
// and the chart opened on the first day itself does not.
func wantSameReturn(t *testing.T, views *app.Views, card *app.PeriodChange, from, to kernel.Date) {
	t.Helper()
	want := twoPlaces(t, card.Return)
	// The preset sets the charts' start to the day before the period's first
	// day; the card's link sets it to the card's opening day.
	for name, start := range map[string]kernel.Date{"preset": from.AddDays(-1), "card link": card.Opening} {
		if got := chartReturn(t, views, start, to); got != want {
			t.Errorf("%s: the chart returns %s%% from %s, the card %s%%", name, got, start, want)
		}
	}
	// Opened on the first day itself, the chart leaves that day's move out.
	if got := chartReturn(t, views, from, to); got == want {
		t.Errorf("a chart opened on %s returns %s%%, the card's too: the data doesn't tell the two openings apart", from, got)
	}
}

// chartReturn opens the performance chart on start and returns its last
// index's distance from 100 as the card shows it, after checking that the
// chart starts on start at 100.
func chartReturn(t *testing.T, views *app.Views, start, to kernel.Date) string {
	t.Helper()
	chart, err := views.Performance(reader(t), start, to)
	if err != nil {
		t.Fatalf("Performance(%s) error = %v", start, err)
	}
	days := chart.Series.Days
	first, last := days[0], days[len(days)-1]
	if first.Date != start || !first.Index.Equal(dec(t, "100")) {
		t.Fatalf("the chart starts on %s at %s, want %s at 100", first.Date, first.Index, start)
	}
	return twoPlaces(t, last.Index.Sub(first.Index))
}

// twoPlaces writes d in percent as the Ask box's card shows it, two decimal
// places, so the comparison is exact at the displayed precision.
func twoPlaces(t *testing.T, d kernel.Decimal) string {
	t.Helper()
	r, err := d.Round(2, kernel.RoundHalfEven)
	if err != nil {
		t.Fatalf("Round: %v", err)
	}
	return r.String()
}
