package views_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
)

// perfBase is the first day of every performance case; days are offsets
// from it.
const perfBase = "2026-03-02"

func perfDay(t *testing.T, offset int) kernel.Date {
	t.Helper()
	return date(t, perfBase).AddDays(offset)
}

func dec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

// event is one transaction of a performance case: a deposit, withdrawal,
// opening cash, buy of ETF1 or opening lot of ETF1, in account acc, on day
// offset day, of amount in currency (the cost for a buy or a lot), qty units.
type event struct {
	kind, acc   string
	day         int
	amount, cur string
	qty         string
}

func (e *event) build(t *testing.T, n int) ledger.Transaction {
	t.Helper()
	h := ledger.TransactionHeader{
		Key:     ledger.TransactionKey{Source: "test", ExternalID: e.kind + "-" + strconv.Itoa(n)},
		Account: ledger.AccountID(e.acc), Date: perfDay(t, e.day),
	}
	amount := money(t, e.amount, e.cur)
	var tx ledger.Transaction
	var err error
	switch e.kind {
	case "deposit":
		tx, err = ledger.NewDeposit(h, amount)
	case "withdrawal":
		tx, err = ledger.NewWithdrawal(h, amount)
	case "opening-cash":
		h.Key.ExternalID = ledger.OpeningCashID(amount.Currency(), h.Date)
		tx, err = ledger.NewOpeningCash(h, amount)
	case "buy":
		tx, err = ledger.NewBuy(h, instrument(t, isinIE, "ETF1"), kernel.NewQuantity(dec(t, e.qty)), amount, kernel.Money{})
	case "opening-lot":
		tx, err = ledger.NewOpeningLot(h, instrument(t, isinIE, "ETF1"), kernel.NewQuantity(dec(t, e.qty)), amount, h.Date.AddDays(-30))
	default:
		t.Fatalf("unknown event kind %q", e.kind)
	}
	if err != nil {
		t.Fatalf("building %s: %v", e.kind, err)
	}
	return tx
}

// row is a stored valuation of acc's ETF1 holding of qty units on day offset
// day worth pln PLN, or lacking a price when pln is "missing".
type row struct {
	acc      string
	day      int
	qty, pln string
}

func (r *row) build(t *testing.T) valuation.HoldingValuation {
	t.Helper()
	v := valuation.HoldingValuation{
		Account: ledger.AccountID(r.acc), Instrument: instrument(t, isinIE, "ETF1"), Date: perfDay(t, r.day),
		Basis: valuation.Market, Quantity: kernel.NewQuantity(dec(t, r.qty)), Source: "test",
	}
	if r.pln == "missing" {
		v.Missing = valuation.MissingPrice
		return v
	}
	v.PLN = money(t, r.pln, "PLN")
	v.Value = v.PLN
	return v
}

// rate is an FX rate to PLN of cur on day offset day.
type rate struct {
	cur  string
	day  int
	rate string
}

// perfCase is the input of one performance case, for the days 0 to last,
// valued through day through.
type perfCase struct {
	events        []event
	rows          []row
	rates         []rate
	last, through int
}

func (c *perfCase) input(t *testing.T) *views.PerformanceInput {
	t.Helper()
	in := &views.PerformanceInput{From: perfDay(t, 0), To: perfDay(t, c.last), ValuedThrough: perfDay(t, c.through), Reporting: pln(t)}
	for i := range c.events {
		in.Transactions = append(in.Transactions, c.events[i].build(t, i))
	}
	for i := range c.rows {
		in.Valuations = append(in.Valuations, c.rows[i].build(t))
	}
	for _, r := range c.rates {
		in.Rates = append(in.Rates, valuation.FXRate{Currency: money(t, "0", r.cur).Currency(), Date: perfDay(t, r.day), Rate: dec(t, r.rate), Source: "test"})
	}
	return in
}

// point is what a case checks of each day: its value, in-transit money and
// net external flow in PLN, and the index.
type point struct {
	Date                   kernel.Date
	Value, InTransit, Flow kernel.Decimal
	Index                  kernel.Decimal
}

func points(days []views.PerformanceDay) []point {
	out := make([]point, 0, len(days))
	for i := range days {
		d := &days[i]
		out = append(out, point{Date: d.Date, Value: d.Value.Amount(), InTransit: d.InTransit.Amount(), Flow: d.Flow.Amount(), Index: d.Index})
	}
	return out
}

// pt is a point on day offset day: value, in transit, flow and index.
type pt struct {
	day                           int
	value, inTransit, flow, index string
}

// investedEvents is acc-1 depositing 1000 PLN on day 0 and buying 10 ETF1
// with it, followed by more.
func investedEvents(more ...event) []event {
	return append([]event{
		{kind: "deposit", acc: "acc-1", day: 0, amount: "1000", cur: "PLN"},
		{kind: "buy", acc: "acc-1", day: 0, amount: "1000", cur: "PLN", qty: "10"},
	}, more...)
}

// investedRows values the 10 ETF1 of investedEvents at 1000, 1100 and 1050
// PLN on days 0 to 2, followed by more.
func investedRows(more ...row) []row {
	return append([]row{{"acc-1", 0, "10", "1000.00"}, {"acc-1", 1, "10", "1100.00"}, {"acc-1", 2, "10", "1050.00"}}, more...)
}

func TestPerformance(t *testing.T) {
	t.Parallel()
	usd := []rate{{"USD", 0, "4.00"}, {"USD", 1, "4.10"}, {"USD", 2, "4.20"}}
	tests := []struct {
		name string
		c    perfCase
		want []pt
		gap  views.PerformanceGap
	}{
		{
			name: "a price move is a return",
			c:    perfCase{events: investedEvents(), rows: investedRows(), last: 2, through: 2},
			want: []pt{{0, "1000", "0", "1000", "100"}, {1, "1100", "0", "0", "110"}, {2, "1050", "0", "0", "105"}},
		},
		{
			name: "a deposit invested at the day's price leaves the index as it was",
			c: perfCase{
				events: investedEvents(
					event{kind: "deposit", acc: "acc-1", day: 1, amount: "1100", cur: "PLN"},
					event{kind: "buy", acc: "acc-1", day: 1, amount: "1100", cur: "PLN", qty: "10"},
				),
				rows: []row{{"acc-1", 0, "10", "1000.00"}, {"acc-1", 1, "20", "2200.00"}, {"acc-1", 2, "20", "2100.00"}},
				last: 2, through: 2,
			},
			// 1.1 × round12(2100/2200) = 1.0499999999995, rounded half-even to 1.05.
			want: []pt{{0, "1000", "0", "1000", "100"}, {1, "2200", "0", "1100", "110"}, {2, "2100", "0", "0", "105"}},
		},
		{
			name: "a deposit kept in cash is not a return but earns nothing after",
			c: perfCase{
				events: investedEvents(event{kind: "deposit", acc: "acc-1", day: 1, amount: "500", cur: "PLN"}),
				rows:   investedRows(), last: 2, through: 2,
			},
			// (1600 - 500) / 1000 = 1.1; then 1550 / 1600 = 0.96875.
			want: []pt{{0, "1000", "0", "1000", "100"}, {1, "1600", "0", "500", "110"}, {2, "1550", "0", "0", "106.5625"}},
		},
		{
			name: "a withdrawal is not a loss",
			c: perfCase{
				events: []event{
					{kind: "deposit", acc: "acc-1", day: 0, amount: "1100", cur: "PLN"},
					{kind: "buy", acc: "acc-1", day: 0, amount: "1000", cur: "PLN", qty: "10"},
					{kind: "withdrawal", acc: "acc-1", day: 1, amount: "100", cur: "PLN"},
				},
				rows: investedRows(), last: 2, through: 2,
			},
			// round12(1200/1100) = 1.090909090909, round12(1050/1100) =
			// 0.954545454545, and their product rounds to 1.041322314049.
			want: []pt{{0, "1100", "0", "1100", "100"}, {1, "1100", "0", "-100", "109.0909090909"}, {2, "1050", "0", "0", "104.1322314049"}},
		},
		{
			name: "a transfer between own accounts is not a cash flow: it is in transit at the day's rate",
			c: perfCase{
				events: []event{
					{kind: "deposit", acc: "acc-1", day: 0, amount: "100", cur: "USD"},
					{kind: "withdrawal", acc: "acc-1", day: 1, amount: "100", cur: "USD"},
					{kind: "deposit", acc: "acc-2", day: 2, amount: "100", cur: "USD"},
				},
				rates: usd, last: 2, through: 2,
			},
			want: []pt{{0, "400", "0", "400", "100"}, {1, "410", "410", "0", "102.5"}, {2, "420", "0", "0", "105"}},
		},
		{
			name: "a transfer whose deposit is booked first is counted once",
			c: perfCase{
				events: []event{
					{kind: "deposit", acc: "acc-1", day: 0, amount: "100", cur: "USD"},
					{kind: "deposit", acc: "acc-2", day: 1, amount: "100", cur: "USD"},
					{kind: "withdrawal", acc: "acc-1", day: 2, amount: "100", cur: "USD"},
				},
				rates: usd, last: 2, through: 2,
			},
			want: []pt{{0, "400", "0", "400", "100"}, {1, "410", "-410", "0", "102.5"}, {2, "420", "0", "0", "105"}},
		},
		{
			name: "a withdrawal and a deposit further apart than the window are cash flows",
			c: perfCase{
				events: []event{
					{kind: "deposit", acc: "acc-1", day: 0, amount: "100", cur: "USD"},
					{kind: "withdrawal", acc: "acc-1", day: 1, amount: "100", cur: "USD"},
					{kind: "deposit", acc: "acc-2", day: 1 + views.TransferWindowDays + 1, amount: "100", cur: "USD"},
				},
				rates: usd, last: 7, through: 7,
			},
			// Day 1 ends with nothing, so days 2 to 7 have no return.
			want: []pt{
				{0, "400", "0", "400", "100"},
				{1, "0", "0", "-410", "102.5"},
				{2, "0", "0", "0", "102.5"},
				{3, "0", "0", "0", "102.5"},
				{4, "0", "0", "0", "102.5"},
				{5, "0", "0", "0", "102.5"},
				{6, "0", "0", "0", "102.5"},
				{7, "420", "0", "420", "102.5"},
			},
		},
		{
			name: "an account seeded later brings its opening lot and cash in as flows",
			c: perfCase{
				events: investedEvents(
					event{kind: "opening-lot", acc: "acc-2", day: 1, amount: "400", cur: "PLN", qty: "5"},
					event{kind: "opening-cash", acc: "acc-2", day: 1, amount: "50", cur: "PLN"},
				),
				rows: investedRows(row{"acc-2", 1, "5", "550.00"}, row{"acc-2", 2, "5", "525.00"}),
				last: 2, through: 2,
			},
			// The lot counts at its valuation, 550, not its cost:
			// (1700 - 600) / 1000 = 1.1, then 1625 / 1700.
			want: []pt{{0, "1000", "0", "1000", "100"}, {1, "1700", "0", "600", "110"}, {2, "1625", "0", "0", "105.1470588235"}},
		},
		{
			name: "a valuation lacking a price stops the series the day before",
			c:    perfCase{events: investedEvents(), rows: append(investedRows()[:2], row{"acc-1", 2, "10", "missing"}), last: 2, through: 2},
			want: []pt{{0, "1000", "0", "1000", "100"}, {1, "1100", "0", "0", "110"}},
			gap:  views.GapUnvalued,
		},
		{
			name: "a holding with no valuation stops the series",
			c:    perfCase{events: investedEvents(), rows: investedRows()[:2], last: 2, through: 2},
			want: []pt{{0, "1000", "0", "1000", "100"}, {1, "1100", "0", "0", "110"}},
			gap:  views.GapUnvalued,
		},
		{
			name: "a valuation of a stale quantity stops the series",
			c:    perfCase{events: investedEvents(), rows: append(investedRows()[:2], row{"acc-1", 2, "9", "945.00"}), last: 2, through: 2},
			want: []pt{{0, "1000", "0", "1000", "100"}, {1, "1100", "0", "0", "110"}},
			gap:  views.GapUnvalued,
		},
		{
			name: "the valued-through date stops the series",
			c:    perfCase{events: investedEvents(), rows: investedRows(), last: 2, through: 1},
			want: []pt{{0, "1000", "0", "1000", "100"}, {1, "1100", "0", "0", "110"}},
			gap:  views.GapUnvalued,
		},
		{
			name: "cash in a currency with no rate stops the series",
			c: perfCase{
				events: investedEvents(event{kind: "deposit", acc: "acc-2", day: 1, amount: "10", cur: "EUR"}),
				rows:   investedRows(), last: 2, through: 2,
			},
			want: []pt{{0, "1000", "0", "1000", "100"}},
			gap:  views.GapNoFXRate,
		},
		{
			name: "days before go-live are worth nothing and have no return",
			c: perfCase{
				events: []event{
					{kind: "deposit", acc: "acc-1", day: 1, amount: "1000", cur: "PLN"},
					{kind: "buy", acc: "acc-1", day: 1, amount: "1000", cur: "PLN", qty: "10"},
				},
				rows: investedRows()[1:], last: 2, through: 2,
			},
			want: []pt{{0, "0", "0", "0", "100"}, {1, "1100", "0", "1000", "100"}, {2, "1050", "0", "0", "95.4545454545"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := views.Performance(tt.c.input(t))
			if err != nil {
				t.Fatalf("Performance() error = %v", err)
			}
			want := make([]point, 0, len(tt.want))
			for _, p := range tt.want {
				want = append(want, point{Date: perfDay(t, p.day), Value: dec(t, p.value), InTransit: dec(t, p.inTransit), Flow: dec(t, p.flow), Index: dec(t, p.index)})
			}
			if diff := cmp.Diff(want, points(got.Days)); diff != "" {
				t.Errorf("Performance() days mismatch (-want +got):\n%s", diff)
			}
			if got.Gap != tt.gap {
				t.Errorf("Performance() gap = %v, want %v", got.Gap, tt.gap)
			}
		})
	}
}

func TestPerformanceParts(t *testing.T) {
	t.Parallel()
	c := perfCase{
		events: investedEvents(event{kind: "deposit", acc: "acc-2", day: 0, amount: "10", cur: "USD"}),
		rows:   investedRows()[:1], rates: []rate{{"USD", -3, "4.00"}}, last: 0, through: 0,
	}
	got, err := views.Performance(c.input(t))
	if err != nil {
		t.Fatalf("Performance() error = %v", err)
	}
	want := []views.PerformanceDay{{
		Date: perfDay(t, 0), Value: money(t, "1040", "PLN"), Holdings: money(t, "1000", "PLN"), Cash: money(t, "40", "PLN"),
		InTransit: money(t, "0", "PLN"), Flow: money(t, "1040", "PLN"), Index: dec(t, "100"),
	}}
	if diff := cmp.Diff(want, got.Days); diff != "" {
		t.Errorf("Performance() mismatch (-want +got):\n%s", diff)
	}
}

func TestPerformanceRejects(t *testing.T) {
	t.Parallel()
	t.Run("a reporting currency other than PLN", func(t *testing.T) {
		t.Parallel()
		in := (&perfCase{events: investedEvents(), rows: investedRows(), last: 2, through: 2}).input(t)
		in.Reporting = money(t, "0", "EUR").Currency()
		if _, err := views.Performance(in); !errors.Is(err, views.ErrNotReportingCurrency) {
			t.Errorf("Performance(EUR) error = %v, want %v", err, views.ErrNotReportingCurrency)
		}
	})
	t.Run("a reversed range", func(t *testing.T) {
		t.Parallel()
		in := (&perfCase{events: investedEvents(), rows: investedRows(), last: 2, through: 2}).input(t)
		in.From, in.To = in.To, in.From
		if _, err := views.Performance(in); !errors.Is(err, views.ErrInvalidRange) {
			t.Errorf("Performance(reversed) error = %v, want %v", err, views.ErrInvalidRange)
		}
	})
}

func TestPerformanceGapString(t *testing.T) {
	t.Parallel()
	for g, want := range map[views.PerformanceGap]string{views.GapUnvalued: "unvalued", views.GapNoFXRate: "no_fx_rate", 0: "PerformanceGap(0)"} {
		if got := g.String(); got != want {
			t.Errorf("PerformanceGap(%d).String() = %q, want %q", int(g), got, want)
		}
	}
}

// eventTxs builds the transactions of events.
func eventTxs(t *testing.T, events []event) []ledger.Transaction {
	t.Helper()
	txs := make([]ledger.Transaction, 0, len(events))
	for i := range events {
		txs = append(txs, events[i].build(t, i))
	}
	return txs
}

// flowsOf returns the cash flow of each of events, which are all deposits
// and withdrawals, in the order of events.
func flowsOf(t *testing.T, events []event) []views.CashFlow {
	t.Helper()
	txs := eventTxs(t, events)
	byKey := map[ledger.TransactionKey]views.CashFlow{}
	for _, f := range views.ExternalFlows(txs) {
		byKey[f.Key] = f
	}
	out := make([]views.CashFlow, 0, len(txs))
	for i := range txs {
		out = append(out, byKey[txs[i].Key])
	}
	return out
}

// matched returns the transfers between the flows at the index pairs of
// transfers, and the flows at the indexes of external.
func matched(flows []views.CashFlow, transfers [][2]int, external []int) ([]views.Transfer, []views.CashFlow) {
	ts := make([]views.Transfer, 0, len(transfers))
	for _, p := range transfers {
		ts = append(ts, views.Transfer{Withdrawal: flows[p[0]], Deposit: flows[p[1]]})
	}
	ext := make([]views.CashFlow, 0, len(external))
	for _, i := range external {
		ext = append(ext, flows[i])
	}
	return ts, ext
}

func TestMatchTransfers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		events    []event
		transfers [][2]int // indexes of the withdrawal and the deposit in events
		external  []int
	}{
		{
			name:      "same day, same amount, another account",
			events:    []event{{kind: "withdrawal", acc: "acc-1", day: 0, amount: "500", cur: "PLN"}, {kind: "deposit", acc: "acc-2", day: 0, amount: "500.00", cur: "PLN"}},
			transfers: [][2]int{{0, 1}},
		},
		{
			name:      "the deposit up to the window after",
			events:    []event{{kind: "withdrawal", acc: "acc-1", day: 0, amount: "500", cur: "PLN"}, {kind: "deposit", acc: "acc-2", day: views.TransferWindowDays, amount: "500", cur: "PLN"}},
			transfers: [][2]int{{0, 1}},
		},
		{
			name:      "the deposit up to the window before",
			events:    []event{{kind: "deposit", acc: "acc-2", day: 0, amount: "500", cur: "PLN"}, {kind: "withdrawal", acc: "acc-1", day: views.TransferWindowDays, amount: "500", cur: "PLN"}},
			transfers: [][2]int{{1, 0}},
		},
		{
			name:     "beyond the window",
			events:   []event{{kind: "withdrawal", acc: "acc-1", day: 0, amount: "500", cur: "PLN"}, {kind: "deposit", acc: "acc-2", day: views.TransferWindowDays + 1, amount: "500", cur: "PLN"}},
			external: []int{0, 1},
		},
		{
			name:     "the same account",
			events:   []event{{kind: "withdrawal", acc: "acc-1", day: 0, amount: "500", cur: "PLN"}, {kind: "deposit", acc: "acc-1", day: 1, amount: "500", cur: "PLN"}},
			external: []int{0, 1},
		},
		{
			name:     "another amount",
			events:   []event{{kind: "withdrawal", acc: "acc-1", day: 0, amount: "500", cur: "PLN"}, {kind: "deposit", acc: "acc-2", day: 1, amount: "499.99", cur: "PLN"}},
			external: []int{0, 1},
		},
		{
			name:     "another currency",
			events:   []event{{kind: "withdrawal", acc: "acc-1", day: 0, amount: "500", cur: "PLN"}, {kind: "deposit", acc: "acc-2", day: 1, amount: "500", cur: "EUR"}},
			external: []int{0, 1},
		},
		{
			name:     "opening cash is never a leg",
			events:   []event{{kind: "withdrawal", acc: "acc-1", day: 0, amount: "500", cur: "PLN"}, {kind: "opening-cash", acc: "acc-2", day: 0, amount: "500", cur: "PLN"}},
			external: []int{0, 1},
		},
		{
			name: "the nearest deposit, each used once",
			events: []event{
				{kind: "withdrawal", acc: "acc-1", day: 2, amount: "500", cur: "PLN"},
				{kind: "deposit", acc: "acc-2", day: 0, amount: "500", cur: "PLN"},
				{kind: "deposit", acc: "acc-3", day: 3, amount: "500", cur: "PLN"},
				{kind: "withdrawal", acc: "acc-1", day: 4, amount: "500", cur: "PLN"},
			},
			transfers: [][2]int{{0, 2}, {3, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flows := flowsOf(t, tt.events)
			wantTransfers, wantExternal := matched(flows, tt.transfers, tt.external)
			gotTransfers, gotExternal := views.MatchTransfers(views.ExternalFlows(eventTxs(t, tt.events)), views.TransferWindowDays)
			if diff := cmp.Diff(wantTransfers, gotTransfers, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("MatchTransfers() transfers mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(wantExternal, gotExternal); diff != "" {
				t.Errorf("MatchTransfers() external mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExternalFlows(t *testing.T) {
	t.Parallel()
	events := []event{
		{kind: "deposit", acc: "acc-2", day: 1, amount: "10", cur: "PLN"},
		{kind: "buy", acc: "acc-1", day: 0, amount: "5", cur: "PLN", qty: "1"},
		{kind: "withdrawal", acc: "acc-1", day: 1, amount: "3", cur: "PLN"},
		{kind: "opening-cash", acc: "acc-1", day: 0, amount: "-7", cur: "PLN"},
	}
	txs := eventTxs(t, events)
	want := []views.CashFlow{
		{Account: "acc-1", Key: txs[3].Key, Date: perfDay(t, 0), Amount: money(t, "-7", "PLN"), Seeding: true},
		{Account: "acc-1", Key: txs[2].Key, Date: perfDay(t, 1), Amount: money(t, "-3", "PLN")},
		{Account: "acc-2", Key: txs[0].Key, Date: perfDay(t, 1), Amount: money(t, "10", "PLN")},
	}
	if diff := cmp.Diff(want, views.ExternalFlows(txs)); diff != "" {
		t.Errorf("ExternalFlows() mismatch (-want +got):\n%s", diff)
	}
}
