package answer

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// askToday is the synthetic day the tests ask on.
const askToday = "2026-03-31"

// mustDay returns the date s.
func mustDay(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func TestParseAsk(t *testing.T) {
	t.Parallel()
	day := func(s string) kernel.Date {
		d, err := kernel.ParseDate(s)
		if err != nil {
			t.Fatalf("ParseDate(%q): %v", s, err)
		}
		return d
	}
	tests := []struct {
		q    string
		want Query
		ok   bool
	}{
		{q: "What needs me?", want: Query{Command: CommandNeeds}, ok: true},
		{q: "  INBOX ", want: Query{Command: CommandNeeds}, ok: true},
		{q: "what needs my attention", want: Query{Command: CommandNeeds}, ok: true},
		{q: "Show bonds", want: Query{Command: CommandShow, Term: "bonds"}, ok: true},
		{q: "show me the account Synthetic brokerage", want: Query{Command: CommandShow, Term: "synthetic brokerage", Show: ShowAccount}, ok: true},
		{q: "sleeve core", want: Query{Command: CommandShow, Term: "core", Show: ShowSleeve}, ok: true},
		{q: "open instrument PLTEST000010", want: Query{Command: CommandShow, Term: "pltest000010", Show: ShowInstrument}, ok: true},
		{q: "Why is xtb failing?", want: Query{Command: CommandSync, Term: "xtb"}, ok: true},
		{q: "why is the flex-main sync failing", want: Query{Command: CommandSync, Term: "flex-main"}, ok: true},
		{q: "why does nbp-fx keep failing", want: Query{Command: CommandSync, Term: "nbp-fx"}, ok: true},
		{q: "sync of xtb-main", want: Query{Command: CommandSync, Term: "xtb-main"}, ok: true},
		{q: "Allocation", want: Query{Command: CommandAllocation, Dimension: "asset-class"}, ok: true},
		{q: "allocation by currency", want: Query{Command: CommandAllocation, Dimension: "currency"}, ok: true},
		{q: "by broker", want: Query{Command: CommandAllocation, Dimension: "institution"}, ok: true},
		{q: "allocation by colour", want: Query{Command: CommandAllocation}},
		{q: "Performance", want: Query{Command: CommandPerformance}, ok: true},
		{q: "performance this month", want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodMonth}}, ok: true},
		{q: "return last 12 months", want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodTwelveMonths}}, ok: true},
		{q: "performance in 2025", want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodCalendarYear, Year: 2025}}, ok: true},
		{q: "performance since 2026-01-15", want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodSince, Since: day("2026-01-15")}}, ok: true},
		{q: "performance since tomorrow", want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodSince}}},
		{q: "performance in November 2025", want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodMonthOf, Year: 2025, Month: time.November}}, ok: true},
		{q: "performance november 2025", want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodMonthOf, Year: 2025, Month: time.November}}, ok: true},
		{q: "return for nov 2025", want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodMonthOf, Year: 2025, Month: time.November}}, ok: true},
		{q: "twr 2025-11", want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodMonthOf, Year: 2025, Month: time.November}}, ok: true},
		{
			q:    "performance from 2025-11-01 to 2025-11-30",
			want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodBetween, Since: day("2025-11-01"), To: day("2025-11-30")}}, ok: true,
		},
		{
			q:    "Performance between 1.11.2025 and 30.11.2025?",
			want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodBetween, Since: day("2025-11-01"), To: day("2025-11-30")}}, ok: true,
		},
		{
			q:    "performance from 2025-11-30 to 2025-11-01",
			want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodBetween, Since: day("2025-11-30"), To: day("2025-11-01")}}, ok: true,
		},
		{q: "performance in 2025-13", want: Query{Command: CommandPerformance}},
		{q: "performance in smarch 2025", want: Query{Command: CommandPerformance}},
		{q: "performance from 2025-11-31 to 2025-12-01", want: Query{Command: CommandPerformance}},
		{q: "performance from 2025-11-01", want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodSince, Since: day("2025-11-01")}}, ok: true},
		{q: "value", want: Query{Command: CommandValue}, ok: true},
		{
			q:    "value change in November 2025",
			want: Query{Command: CommandValue, Change: &Period{Kind: PeriodMonthOf, Year: 2025, Month: time.November}}, ok: true,
		},
		{
			q:    "change in value in November 2025",
			want: Query{Command: CommandValue, Change: &Period{Kind: PeriodMonthOf, Year: 2025, Month: time.November}}, ok: true,
		},
		{
			q:    "value change from 2025-11-01 to 2025-11-30",
			want: Query{Command: CommandValue, Change: &Period{Kind: PeriodBetween, Since: day("2025-11-01"), To: day("2025-11-30")}}, ok: true,
		},
		{
			q:    "value from 2025-11-01 to 2025-11-30",
			want: Query{Command: CommandValue, Change: &Period{Kind: PeriodBetween, Since: day("2025-11-01"), To: day("2025-11-30")}}, ok: true,
		},
		{
			q:    "value between 2025-11-01 and 2025-11-30",
			want: Query{Command: CommandValue, Change: &Period{Kind: PeriodBetween, Since: day("2025-11-01"), To: day("2025-11-30")}}, ok: true,
		},
		{q: "value in november 2025", want: Query{Command: CommandValue, Change: &Period{Kind: PeriodMonthOf, Year: 2025, Month: time.November}}, ok: true},
		{q: "value for 2025", want: Query{Command: CommandValue, Change: &Period{Kind: PeriodCalendarYear, Year: 2025}}, ok: true},
		{q: "value change", want: Query{Command: CommandValue}},
		{q: "value change last autumn", want: Query{Command: CommandValue}},
		{q: "value in the autumn", want: Query{Command: CommandValue}},
		{q: "value from 2025-11-01", want: Query{Command: CommandValue}},
		{q: "value on 2026-02-27", want: Query{Command: CommandValue, Date: day("2026-02-27")}, ok: true},
		{q: "value on 27.2.2026", want: Query{Command: CommandValue, Date: day("2026-02-27")}, ok: true},
		{q: "value yesterday", want: Query{Command: CommandValue, Date: day("2026-03-30")}, ok: true},
		{q: "value at end of last month", want: Query{Command: CommandValue, Date: day("2026-02-28")}, ok: true},
		{q: "value on 2026-02-30", want: Query{Command: CommandValue}},
		{q: "buy more bonds", want: Query{}},
		{q: "", want: Query{}},
	}
	for _, tt := range tests {
		t.Run(tt.q, func(t *testing.T) {
			t.Parallel()
			got, ok := Parse(tt.q, day(askToday))
			if ok != tt.ok {
				t.Fatalf("Parse(%q) ok = %v, want %v", tt.q, ok, tt.ok)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" && ok {
				t.Errorf("Parse(%q) (-want +got):\n%s", tt.q, diff)
			}
		})
	}
}

func TestPeriodRange(t *testing.T) {
	t.Parallel()
	to, err := kernel.ParseDate(askToday)
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	since, err := kernel.ParseDate("2026-01-15")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	tests := []struct {
		p              Period
		from, last, lb string
	}{
		{Period{Kind: PeriodYear}, "2026-01-01", askToday, "this year"},
		{Period{Kind: PeriodMonth}, "2026-03-01", askToday, "this month"},
		{Period{Kind: PeriodTwelveMonths}, "2025-04-01", askToday, "over the last 12 months"},
		{Period{Kind: PeriodLastYear}, "2025-01-01", "2025-12-31", "last year"},
		{Period{Kind: PeriodAll}, "0001-01-01", askToday, "since tracking began"},
		{Period{Kind: PeriodSince, Since: since}, "2026-01-15", askToday, "since 2026-01-15"},
		{Period{Kind: PeriodCalendarYear, Year: 2026}, "2026-01-01", askToday, "in 2026"},
		{Period{Kind: PeriodMonthOf, Year: 2025, Month: time.November}, "2025-11-01", "2025-11-30", "in November 2025"},
		{Period{Kind: PeriodMonthOf, Year: 2026, Month: time.December}, "2026-12-01", "2026-12-31", "in December 2026"},
		{Period{Kind: PeriodBetween, Since: since, To: to}, "2026-01-15", askToday, "from 2026-01-15 to 2026-03-31"},
	}
	for _, tt := range tests {
		t.Run(tt.lb, func(t *testing.T) {
			t.Parallel()
			s := periodRange(tt.p, to)
			got := []string{s.From.String(), s.To.String(), s.Label}
			if s.From.IsZero() {
				got[0] = "0001-01-01"
			}
			if diff := cmp.Diff([]string{tt.from, tt.last, tt.lb}, got); diff != "" {
				t.Errorf("periodRange (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMatchTarget(t *testing.T) {
	t.Parallel()
	targets := []askTarget{
		{Kind: ShowAccount, ID: "acc-1", Label: "Synthetic brokerage", Words: []string{"Synthetic brokerage", "acc-1", "test-broker"}},
		{Kind: ShowAccount, ID: "acc-bonds", Label: "Treasury bonds", Words: []string{"Treasury bonds", "acc-bonds"}},
		{Kind: ShowSleeve, ID: "core", Label: "Core", Words: []string{"Core", "core"}},
		{Kind: ShowInstrument, ID: "PLTEST000010", Label: "Synthetic Fund · TESTA (PLTEST000010)", Words: []string{"Synthetic Fund", "TESTA", "PLTEST000010"}},
	}
	tests := []struct {
		term string
		kind ShowKind
		want string
		ok   bool
	}{
		{term: "acc-1", want: "acc-1", ok: true},
		{term: "bonds", want: "acc-bonds", ok: true},
		{term: "core", want: "core", ok: true},
		{term: "core", kind: ShowAccount},
		{term: "testa", want: "PLTEST000010", ok: true},
		{term: "pltest000010", kind: ShowInstrument, want: "PLTEST000010", ok: true},
		{term: "synth", want: "acc-1", ok: true},
		{term: "synth", kind: ShowInstrument, want: "PLTEST000010", ok: true},
		{term: "nothing like it"},
		{term: " "},
	}
	for _, tt := range tests {
		t.Run(tt.term, func(t *testing.T) {
			t.Parallel()
			got, ok := matchTarget(tt.term, tt.kind, targets)
			if ok != tt.ok || got.ID != tt.want {
				t.Errorf("matchTarget(%q, %v) = %q, %v; want %q, %v", tt.term, tt.kind, got.ID, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestMatchInstance(t *testing.T) {
	t.Parallel()
	instances := []string{"eodhd", "nbp-fx", "xtb-main"}
	tests := []struct{ term, want string }{
		{"xtb", "xtb-main"},
		{"the xtb-main", "xtb-main"},
		{"NBP-FX", "nbp-fx"},
		{"ibkr", ""},
	}
	for _, tt := range tests {
		t.Run(tt.term, func(t *testing.T) {
			t.Parallel()
			if got, ok := matchInstance(tt.term, instances); got != tt.want || ok != (tt.want != "") {
				t.Errorf("matchInstance(%q) = %q, %v; want %q", tt.term, got, ok, tt.want)
			}
		})
	}
}

func TestKnownCommandsAreUnderstood(t *testing.T) {
	t.Parallel()
	for _, c := range KnownCommands() {
		if _, ok := Parse(c.Example, mustDay(t, askToday)); !ok {
			t.Errorf("example %q of %q is not understood", c.Example, c.Command)
		}
	}
}

// TestSyncAnswerSaysWhySyncFails checks the card of "why is … failing" (D-194,
// D-198): the latest attempts each with its time and its error class, never its
// text, and, for a sync that failed, how long it has been failing, how many
// attempts in a row failed and the latest error class, in words that name it as
// the error class so that nothing reads it as missing data.
func TestSyncAnswerSaysWhySyncFails(t *testing.T) {
	t.Parallel()
	today := mustDay(t, askToday)
	failed := func(at string, class viewsv1.SyncErrorClass) *viewsv1.SyncAttempt {
		return &viewsv1.SyncAttempt{StartedAt: at, Mode: "scheduled", Failed: true, ErrorClass: class}
	}
	ok := &viewsv1.SyncAttempt{StartedAt: "2026-03-27T06:00:00Z", Mode: "scheduled"}
	timeout, unavailable := viewsv1.SyncErrorClass_SYNC_ERROR_CLASS_TIMEOUT, viewsv1.SyncErrorClass_SYNC_ERROR_CLASS_UNAVAILABLE
	tests := []struct {
		name     string
		status   *viewsv1.SyncStatus
		attempts []*viewsv1.SyncAttempt
		title    string
		text     string
		rows     []Row
	}{
		{
			name: "keeps failing for days",
			status: &viewsv1.SyncStatus{
				Failures: 4, FailingSince: "2026-03-28T06:00:00Z", KeepsFailing: true, LastErrorClass: unavailable,
				LastSuccessAt: "2026-03-27T06:00:00Z", NextSyncAt: "2026-04-01T06:00:00Z",
			},
			attempts: []*viewsv1.SyncAttempt{failed("2026-03-31T06:00:00Z", unavailable), failed("2026-03-30T06:00:00Z", unavailable), ok},
			title:    "The x sync keeps failing since 2026-03-28T06:00:00Z",
			text: "Failing for 3 days, since 2026-03-28: 4 failed attempts in a row, the latest with the error class unavailable. " +
				"Last good sync: 2026-03-27T06:00:00Z. Next try: 2026-04-01T06:00:00Z. Every attempt keeps its error class, never its text.",
			rows: []Row{
				{Label: "2026-03-31T06:00:00Z", Detail: "scheduled sync failed, error class", Figure: "unavailable"},
				{Label: "2026-03-30T06:00:00Z", Detail: "scheduled sync failed, error class", Figure: "unavailable"},
				{Label: "2026-03-27T06:00:00Z", Detail: "scheduled sync", Figure: "went through"},
			},
		},
		{
			name:     "failing since today",
			status:   &viewsv1.SyncStatus{Failures: 3, FailingSince: "2026-03-31T01:00:00Z", KeepsFailing: true, LastErrorClass: timeout},
			attempts: []*viewsv1.SyncAttempt{failed("2026-03-31T07:00:00Z", timeout)},
			title:    "The x sync keeps failing since 2026-03-31T01:00:00Z",
			text: "Failing for less than a day, since 2026-03-31: 3 failed attempts in a row, the latest with the error class timeout. " +
				"Last good sync: none yet. Next try: none yet. Every attempt keeps its error class, never its text.",
			rows: []Row{{Label: "2026-03-31T07:00:00Z", Detail: "scheduled sync failed, error class", Figure: "timeout"}},
		},
		{
			name:     "one day, one failure, not yet keeping failing",
			status:   &viewsv1.SyncStatus{Failures: 1, FailingSince: "2026-03-30", LastErrorClass: timeout},
			attempts: []*viewsv1.SyncAttempt{failed("2026-03-30T06:00:00Z", timeout)},
			title:    "The x sync failed 1 time in a row",
			text: "Failing for 1 day, since 2026-03-30: 1 failed attempt in a row, the latest with the error class timeout. " +
				"Last good sync: none yet. Next try: none yet. Every attempt keeps its error class, never its text.",
			rows: []Row{{Label: "2026-03-30T06:00:00Z", Detail: "scheduled sync failed, error class", Figure: "timeout"}},
		},
		{
			name:     "no class and no start",
			status:   &viewsv1.SyncStatus{Failures: 2},
			attempts: []*viewsv1.SyncAttempt{failed("2026-03-30T06:00:00Z", viewsv1.SyncErrorClass_SYNC_ERROR_CLASS_UNSPECIFIED)},
			title:    "The x sync failed 2 times in a row",
			text: "Failing: 2 failed attempts in a row. Last good sync: none yet. Next try: none yet. " +
				"Every attempt keeps its error class, never its text.",
			rows: []Row{{Label: "2026-03-30T06:00:00Z", Detail: "scheduled sync failed, error class", Figure: "unspecified"}},
		},
		{
			name:     "a sync that works says nothing of failing",
			status:   &viewsv1.SyncStatus{LastSuccessAt: "2026-03-31T06:00:00Z", NextSyncAt: "2026-04-01T06:00:00Z"},
			attempts: []*viewsv1.SyncAttempt{{StartedAt: "2026-03-31T06:00:00Z", Mode: "scheduled"}},
			title:    "The x sync is working",
			text:     "Last good sync: 2026-03-31T06:00:00Z. Next try: 2026-04-01T06:00:00Z. Every attempt keeps its error class, never its text.",
			rows:     []Row{{Label: "2026-03-31T06:00:00Z", Detail: "scheduled sync", Figure: "went through"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card := syncAnswer("x", &viewsv1.ListSyncAttemptsResponse{Status: tt.status, Attempts: tt.attempts}, today, NoLinks{})
			if card.Title != tt.title || card.Text != tt.text {
				t.Errorf("card = %q / %q, want %q / %q", card.Title, card.Text, tt.title, tt.text)
			}
			if diff := cmp.Diff(tt.rows, card.Rows); diff != "" {
				t.Errorf("rows (-want +got):\n%s", diff)
			}
		})
	}
}
