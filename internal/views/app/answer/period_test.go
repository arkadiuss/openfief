package answer

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

func TestParseBounded(t *testing.T) {
	t.Parallel()
	today := mustDay(t, askToday)
	month := func(year int, m time.Month) Period { return Period{Kind: PeriodMonthOf, Year: year, Month: m} }
	between := func(from, to string) Period {
		return Period{Kind: PeriodBetween, Since: mustDay(t, from), To: mustDay(t, to)}
	}
	tests := []struct {
		words string
		want  Period
		ok    bool
	}{
		{words: "november 2025", want: month(2025, time.November), ok: true},
		{words: "in november 2025", want: month(2025, time.November), ok: true},
		{words: "for november 2025", want: month(2025, time.November), ok: true},
		{words: "during march 2024", want: month(2024, time.March), ok: true},
		{words: "nov 2025", want: month(2025, time.November), ok: true},
		{words: "sept 2025", want: month(2025, time.September), ok: true},
		{words: "may 2025", want: month(2025, time.May), ok: true},
		{words: "december 2025", want: month(2025, time.December), ok: true},
		{words: "2025-11", want: month(2025, time.November), ok: true},
		{words: "in 2025-01", want: month(2025, time.January), ok: true},
		{words: "2025-13"},
		{words: "2025-00"},
		{words: "smarch 2025"},
		{words: "2025"},
		{words: "in 2025"},
		{words: "from 2025-11-01 to 2025-11-30", want: between("2025-11-01", "2025-11-30"), ok: true},
		{words: "between 2025-11-01 and 2025-11-30", want: between("2025-11-01", "2025-11-30"), ok: true},
		{words: "2025-11-01 to 2025-11-30", want: between("2025-11-01", "2025-11-30"), ok: true},
		{words: "from 1.11.2025 until 30.11.2025", want: between("2025-11-01", "2025-11-30"), ok: true},
		{words: "from 2025-03-01 to today", want: between("2025-03-01", askToday), ok: true},
		{words: "from 2025-11-30 to 2025-11-01", want: between("2025-11-30", "2025-11-01"), ok: true},
		{words: "from 2025-11-31 to 2025-12-01"},
		{words: "from 2025-11-01 to someday"},
		{words: "from 2025-11-01"},
		{words: "since 2025-11-01"},
		{words: "last 12 months"},
		{words: ""},
	}
	for _, tt := range tests {
		t.Run(tt.words, func(t *testing.T) {
			t.Parallel()
			got, ok := parseBounded(tt.words, today)
			if ok != tt.ok {
				t.Fatalf("parseBounded(%q) ok = %v, want %v", tt.words, ok, tt.ok)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseBounded(%q) (-want +got):\n%s", tt.words, diff)
			}
		})
	}
}

func TestParseChange(t *testing.T) {
	t.Parallel()
	today := mustDay(t, askToday)
	tests := []struct {
		words string
		want  Period
		ok    bool
	}{
		{words: "in november 2025", want: Period{Kind: PeriodMonthOf, Year: 2025, Month: time.November}, ok: true},
		{words: "in 2025", want: Period{Kind: PeriodCalendarYear, Year: 2025}, ok: true},
		{words: "2025", want: Period{Kind: PeriodCalendarYear, Year: 2025}, ok: true},
		{words: "  for 2024 ", want: Period{Kind: PeriodCalendarYear, Year: 2024}, ok: true},
		{words: "", ok: false},
		{words: "this year", ok: false},
		{words: "last month", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.words, func(t *testing.T) {
			t.Parallel()
			got, ok := parseChange(tt.words, today)
			if ok != tt.ok {
				t.Fatalf("parseChange(%q) ok = %v, want %v", tt.words, ok, tt.ok)
			}
			if diff := cmp.Diff(tt.want, got); ok && diff != "" {
				t.Errorf("parseChange(%q) (-want +got):\n%s", tt.words, diff)
			}
		})
	}
}

func TestBoundedSpan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		p              Period
		from, to, text string
	}{
		{"a month", Period{Kind: PeriodMonthOf, Year: 2025, Month: time.November}, "2025-11-01", "2025-11-30", "in November 2025"},
		{"a 31-day month", Period{Kind: PeriodMonthOf, Year: 2025, Month: time.October}, "2025-10-01", "2025-10-31", "in October 2025"},
		{"December", Period{Kind: PeriodMonthOf, Year: 2025, Month: time.December}, "2025-12-01", "2025-12-31", "in December 2025"},
		{"February in a leap year", Period{Kind: PeriodMonthOf, Year: 2024, Month: time.February}, "2024-02-01", "2024-02-29", "in February 2024"},
		{"February in another", Period{Kind: PeriodMonthOf, Year: 2025, Month: time.February}, "2025-02-01", "2025-02-28", "in February 2025"},
		{
			"two days",
			Period{Kind: PeriodBetween, Since: mustDay(t, "2025-11-01"), To: mustDay(t, "2025-11-30")},
			"2025-11-01", "2025-11-30", "from 2025-11-01 to 2025-11-30",
		},
		{"a year, whole, never cut", Period{Kind: PeriodCalendarYear, Year: 2026}, "2026-01-01", "2026-12-31", "in 2026"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := boundedSpan(tt.p)
			if diff := cmp.Diff([]string{tt.from, tt.to, tt.text}, []string{s.From.String(), s.To.String(), s.Label}); diff != "" {
				t.Errorf("boundedSpan (-want +got):\n%s", diff)
			}
		})
	}
	if s := boundedSpan(Period{Kind: PeriodYear}); !s.From.IsZero() || !s.To.IsZero() {
		t.Errorf("boundedSpan(this year) = %+v, want none: only a month, a year and two days are bounded", s)
	}
}

func TestPeriodProblem(t *testing.T) {
	t.Parallel()
	latest := mustDay(t, "2026-03-31")
	tests := []struct {
		name     string
		from, to string
		want     string
	}{
		{name: "a period inside the valued days", from: "2025-11-01", to: "2025-11-30"},
		{name: "ending on the latest valued day", from: "2026-03-01", to: "2026-03-31"},
		{name: "one day", from: "2026-03-31", to: "2026-03-31"},
		{
			name: "ending before it starts", from: "2025-11-30", to: "2025-11-01",
			want: "The period ends on 2025-11-01, before it starts on 2025-11-30.",
		},
		{
			name: "ending after the latest valued day", from: "2026-03-01", to: "2026-04-01",
			want: "The period ends on 2026-04-01, but your accounts are valued only through 2026-03-31.",
		},
		{
			name: "entirely after it", from: "2026-05-01", to: "2026-05-31",
			want: "The period ends on 2026-05-31, but your accounts are valued only through 2026-03-31.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := periodProblem(span{From: mustDay(t, tt.from), To: mustDay(t, tt.to)}, latest)
			if got != tt.want {
				t.Errorf("periodProblem(%s to %s) = %q, want %q", tt.from, tt.to, got, tt.want)
			}
		})
	}
	if got := periodProblem(span{}, latest); got == "" {
		t.Error("periodProblem of no days = \"\", want the problem: a period needs a first and a last day")
	}
}

func TestPercentText(t *testing.T) {
	t.Parallel()
	f := NumberFormat{Group: " ", Decimal: ","}
	tests := []struct{ in, want string }{
		{"3.47", "+3,47%"},
		{"-1.2", "-1,20%"},
		{"-1.20", "-1,20%"},
		{"0", "+0,00%"},
		{"12", "+12,00%"},
		{"1234.5", "+1 234,50%"},
		{"-0.05", "-0,05%"},
		{"", ""},
		{"n/a", "n/a"},
	}
	for _, tt := range tests {
		if got := percentText(tt.in, f); got != tt.want {
			t.Errorf("percentText(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestMoved(t *testing.T) {
	t.Parallel()
	f := NumberFormat{Group: " ", Decimal: ","}
	tests := []struct{ amount, want string }{
		{"1966.85", "rose by 1 966,85 PLN"},
		{"-50", "fell by 50 PLN"},
		{"0", "did not change"},
		{"0.00", "did not change"},
		{"", "did not change"},
	}
	for _, tt := range tests {
		if got := moved(&viewsv1.Money{Amount: tt.amount, Currency: "PLN"}, f); got != tt.want {
			t.Errorf("moved(%q) = %q, want %q", tt.amount, got, tt.want)
		}
	}
}

func TestChangeRow(t *testing.T) {
	t.Parallel()
	f := NumberFormat{Group: " ", Decimal: ","}
	pln := func(a string) *viewsv1.Money { return &viewsv1.Money{Amount: a, Currency: "PLN"} }
	tests := []struct {
		name string
		c    *viewsv1.ValueChange
		what string
		want Row
	}{
		{
			name: "a rise", what: "",
			c:    &viewsv1.ValueChange{StartReporting: pln("56719.52"), EndReporting: pln("58686.37"), ChangeReporting: pln("1966.85"), ChangePercent: "3.47"},
			want: Row{Label: "Gross", Detail: "56 719,52 PLN to 58 686,37 PLN", Figure: "+1 966,85 PLN (+3,47%)"},
		},
		{
			name: "a fall with what it is", what: "if everything were sold",
			c:    &viewsv1.ValueChange{StartReporting: pln("100"), EndReporting: pln("90"), ChangeReporting: pln("-10"), ChangePercent: "-10"},
			want: Row{Label: "Gross", Detail: "if everything were sold: 100 PLN to 90 PLN", Figure: "-10 PLN (-10,00%)"},
		},
		{
			name: "no percentage from nothing", what: "",
			c:    &viewsv1.ValueChange{StartReporting: pln("0"), EndReporting: pln("5"), ChangeReporting: pln("5")},
			want: Row{Label: "Gross", Detail: "0 PLN to 5 PLN", Figure: "+5 PLN"},
		},
		{name: "a basis the core leaves out", what: "after the taxes due", want: Row{Label: "Gross", Detail: "after the taxes due", Figure: "missing"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, changeRow("Gross", tt.what, tt.c, f)); diff != "" {
				t.Errorf("changeRow (-want +got):\n%s", diff)
			}
		})
	}
}

// TestPeriodOpening checks the day a period opens on the charts (D-223): the
// close of the day before its first day, so the first day counts in full,
// and none for the period with no first day.
func TestPeriodOpening(t *testing.T) {
	t.Parallel()
	to := mustDay(t, "2026-03-31")
	tests := []struct {
		name string
		p    Period
		want string
	}{
		{"this year opens on 31 December", Period{Kind: PeriodYear}, "2025-12-31"},
		{"this month opens on the last day of February", Period{Kind: PeriodMonth}, "2026-02-28"},
		{"the last 12 months open 365 days before", Period{Kind: PeriodTwelveMonths}, "2025-03-31"},
		{"last year opens on 31 December 2024", Period{Kind: PeriodLastYear}, "2024-12-31"},
		{"2025 opens on 31 December 2024", Period{Kind: PeriodCalendarYear, Year: 2025}, "2024-12-31"},
		{"November 2025 opens on 31 October", Period{Kind: PeriodMonthOf, Year: 2025, Month: time.November}, "2025-10-31"},
		{"since a day opens on the day before", Period{Kind: PeriodSince, Since: mustDay(t, "2026-01-15")}, "2026-01-14"},
		{"everything has no opening: the charts' default start", Period{Kind: PeriodAll}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := OptionalString(PeriodOpening(tt.p, to)); got != tt.want {
				t.Errorf("PeriodOpening = %q, want %q", got, tt.want)
			}
		})
	}
}
