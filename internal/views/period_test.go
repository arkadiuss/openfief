package views_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views"
)

// periodTotal is a total of the three values on a day; "" for net or
// liquidation means that value is missing.
type periodTotal struct {
	day, gross, net, liq string
}

func (p periodTotal) build(t *testing.T) views.DateTotal {
	t.Helper()
	out := views.DateTotal{Date: date(t, p.day), Reporting: money(t, p.gross, "PLN")}
	if p.net == "" {
		out.NetMissing = true
	} else {
		out.Net = money(t, p.net, "PLN")
	}
	if p.liq == "" {
		out.LiquidationMissing = true
	} else {
		out.Liquidation = money(t, p.liq, "PLN")
	}
	return out
}

// periodDay is a day of the performance series: its index and its net cash
// flow.
type periodDay struct{ day, index, flow string }

// changeText is a ValueChange as text: start, end, change and percent.
type changeText struct{ start, end, change, percent string }

// periodWant is a PeriodChange as text.
type periodWant struct {
	opening                            string
	gross                              changeText
	net, liq                           *changeText
	ret                                string
	deposits, withdrawals, netDeposits string
}

func textOf(c *views.ValueChange) *changeText {
	if !c.Known {
		return nil
	}
	pct := ""
	if c.HasPercent {
		pct = c.Percent.String()
	}
	return &changeText{c.Start.Amount().String(), c.End.Amount().String(), c.Change.Amount().String(), pct}
}

// normalized writes the decimals of w the way a Decimal prints them, without
// trailing zeros: "3.50" is "3.5".
func normalized(t *testing.T, w *periodWant) periodWant {
	t.Helper()
	norm := func(s string) string {
		if s == "" {
			return s
		}
		return dec(t, s).String()
	}
	change := func(c *changeText) *changeText {
		if c == nil {
			return nil
		}
		return &changeText{norm(c.start), norm(c.end), norm(c.change), norm(c.percent)}
	}
	g := change(&w.gross)
	return periodWant{
		opening: w.opening, gross: *g, net: change(w.net), liq: change(w.liq), ret: norm(w.ret),
		deposits: norm(w.deposits), withdrawals: norm(w.withdrawals), netDeposits: norm(w.netDeposits),
	}
}

func TestChangeOver(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		opening, end periodTotal
		days         []periodDay
		want         periodWant
	}{
		{
			name:    "value rose partly from a deposit while the return fell",
			opening: periodTotal{"2025-10-31", "1000.00", "900.00", "880.00"},
			end:     periodTotal{"2025-11-03", "1100.00", "985.00", "960.00"},
			days: []periodDay{
				{"2025-10-31", "100", "0.00"}, {"2025-11-01", "99", "0.00"}, {"2025-11-02", "98.5", "200.00"}, {"2025-11-03", "98", "0.00"},
			},
			want: periodWant{
				opening: "2025-10-31",
				gross:   changeText{"1000.00", "1100.00", "100.00", "10.00"},
				net:     &changeText{"900.00", "985.00", "85.00", "9.44"},
				liq:     &changeText{"880.00", "960.00", "80.00", "9.09"},
				ret:     "-2.00", deposits: "200.00", withdrawals: "0.00", netDeposits: "200.00",
			},
		},
		{
			name:    "withdrawals and deposits are told apart, by day",
			opening: periodTotal{"2026-01-01", "5000.00", "4500.00", "4400.00"},
			end:     periodTotal{"2026-01-04", "4600.00", "4150.00", "4050.00"},
			days: []periodDay{
				{"2026-01-01", "100", "0.00"}, {"2026-01-02", "101", "-1000.00"}, {"2026-01-03", "102", "300.00"}, {"2026-01-04", "103.5", "-50.50"},
			},
			want: periodWant{
				opening: "2026-01-01",
				gross:   changeText{"5000.00", "4600.00", "-400.00", "-8.00"},
				net:     &changeText{"4500.00", "4150.00", "-350.00", "-7.78"},
				liq:     &changeText{"4400.00", "4050.00", "-350.00", "-7.95"},
				ret:     "3.50", deposits: "300.00", withdrawals: "1050.50", netDeposits: "-750.50",
			},
		},
		{
			name:    "the opening day's own flow is already in the start",
			opening: periodTotal{"2026-01-01", "1000.00", "1000.00", "1000.00"},
			end:     periodTotal{"2026-01-02", "1010.00", "1010.00", "1010.00"},
			days:    []periodDay{{"2026-01-01", "100", "1000.00"}, {"2026-01-02", "101", "0.00"}},
			want: periodWant{
				opening: "2026-01-01",
				gross:   changeText{"1000.00", "1010.00", "10.00", "1.00"},
				net:     &changeText{"1000.00", "1010.00", "10.00", "1.00"},
				liq:     &changeText{"1000.00", "1010.00", "10.00", "1.00"},
				ret:     "1.00", deposits: "0.00", withdrawals: "0.00", netDeposits: "0.00",
			},
		},
		{
			name:    "a value missing at either end is not given",
			opening: periodTotal{"2026-01-01", "1000.00", "", "950.00"},
			end:     periodTotal{"2026-01-02", "1000.00", "980.00", ""},
			days:    []periodDay{{"2026-01-01", "100", "0.00"}, {"2026-01-02", "100", "0.00"}},
			want: periodWant{
				opening: "2026-01-01",
				gross:   changeText{"1000.00", "1000.00", "0.00", "0.00"},
				ret:     "0.00", deposits: "0.00", withdrawals: "0.00", netDeposits: "0.00",
			},
		},
		{
			name:    "a start of nothing has no percentage",
			opening: periodTotal{"2026-01-01", "0.00", "0.00", "0.00"},
			end:     periodTotal{"2026-01-02", "500.00", "500.00", "500.00"},
			days:    []periodDay{{"2026-01-01", "100", "0.00"}, {"2026-01-02", "100", "500.00"}},
			want: periodWant{
				opening: "2026-01-01",
				gross:   changeText{"0.00", "500.00", "500.00", ""},
				net:     &changeText{"0.00", "500.00", "500.00", ""},
				liq:     &changeText{"0.00", "500.00", "500.00", ""},
				ret:     "0.00", deposits: "500.00", withdrawals: "0.00", netDeposits: "500.00",
			},
		},
		{
			name:    "a tie rounds half-even to two places (0.125 is 0.12)",
			opening: periodTotal{"2026-01-01", "800.00", "800.00", "800.00"},
			end:     periodTotal{"2026-01-02", "801.00", "801.00", "801.00"},
			days:    []periodDay{{"2026-01-01", "100", "0.00"}, {"2026-01-02", "100.125", "0.00"}},
			want: periodWant{
				opening: "2026-01-01",
				gross:   changeText{"800.00", "801.00", "1.00", "0.12"},
				net:     &changeText{"800.00", "801.00", "1.00", "0.12"},
				liq:     &changeText{"800.00", "801.00", "1.00", "0.12"},
				ret:     "0.12", deposits: "0.00", withdrawals: "0.00", netDeposits: "0.00",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in := &views.PeriodInput{
				From: date(t, tt.opening.day), To: date(t, tt.end.day), Opening: tt.opening.build(t), Closing: tt.end.build(t), Reporting: pln(t),
			}
			for _, d := range tt.days {
				in.Days = append(in.Days, views.PerformanceDay{Date: date(t, d.day), Index: dec(t, d.index), Flow: money(t, d.flow, "PLN")})
			}
			got, err := views.ChangeOver(in)
			if err != nil {
				t.Fatalf("ChangeOver() error = %v", err)
			}
			have := periodWant{
				opening: got.Opening.String(), gross: *textOf(&got.Gross), net: textOf(&got.Net), liq: textOf(&got.Liquidation),
				ret: got.Return.String(), deposits: got.Deposits.Amount().String(), withdrawals: got.Withdrawals.Amount().String(),
				netDeposits: got.NetDeposits.Amount().String(),
			}
			if diff := cmp.Diff(normalized(t, &tt.want), have, cmp.AllowUnexported(periodWant{}, changeText{})); diff != "" {
				t.Errorf("ChangeOver() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestChangeOverRefusesAnUnvaluedPeriod(t *testing.T) {
	t.Parallel()
	open := periodTotal{"2026-01-01", "1000.00", "1000.00", "1000.00"}
	end := periodTotal{"2026-01-03", "1000.00", "1000.00", "1000.00"}
	tests := []struct {
		name string
		days []periodDay
		to   string
	}{
		{name: "no days", to: "2026-01-03"},
		{name: "the series stops before the last day", days: []periodDay{{"2026-01-01", "100", "0"}, {"2026-01-02", "100", "0"}}, to: "2026-01-03"},
		{name: "the series starts after the opening day", days: []periodDay{{"2026-01-02", "100", "0"}, {"2026-01-03", "100", "0"}}, to: "2026-01-03"},
		{name: "the total is not the last day's", days: []periodDay{{"2026-01-01", "100", "0"}, {"2026-01-03", "100", "0"}}, to: "2026-01-04"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in := &views.PeriodInput{From: date(t, "2026-01-02"), To: date(t, tt.to), Opening: open.build(t), Closing: end.build(t), Reporting: pln(t)}
			for _, d := range tt.days {
				in.Days = append(in.Days, views.PerformanceDay{Date: date(t, d.day), Index: dec(t, d.index), Flow: money(t, d.flow, "PLN")})
			}
			if _, err := views.ChangeOver(in); !errors.Is(err, views.ErrPeriodUnvalued) {
				t.Errorf("ChangeOver() error = %v, want %v", err, views.ErrPeriodUnvalued)
			}
		})
	}
}

func TestChangeOverRefusesAnotherCurrency(t *testing.T) {
	t.Parallel()
	open := periodTotal{"2026-01-01", "1000.00", "1000.00", "1000.00"}.build(t)
	end := periodTotal{"2026-01-02", "1000.00", "1000.00", "1000.00"}.build(t)
	end.Reporting = money(t, "1000.00", "EUR")
	in := &views.PeriodInput{
		From: date(t, "2026-01-02"), To: date(t, "2026-01-02"), Opening: open, Closing: end, Reporting: pln(t),
		Days: []views.PerformanceDay{
			{Date: date(t, "2026-01-01"), Index: dec(t, "100"), Flow: kernel.NewMoney(kernel.Decimal{}, pln(t))},
			{Date: date(t, "2026-01-02"), Index: dec(t, "100"), Flow: kernel.NewMoney(kernel.Decimal{}, pln(t))},
		},
	}
	var mismatch *kernel.CurrencyMismatchError
	if _, err := views.ChangeOver(in); !errors.As(err, &mismatch) {
		t.Errorf("ChangeOver() error = %v, want a currency mismatch", err)
	}
}
