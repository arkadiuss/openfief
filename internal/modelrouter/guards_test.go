package modelrouter_test

import (
	"errors"
	"testing"
	"time"
	_ "time/tzdata" // the zone the budget's day is counted in

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/modelrouter"
)

func decimal(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

func currency(t *testing.T, code string) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", code, err)
	}
	return c
}

func TestCallCostPricesInputAndOutputTokensPerMillion(t *testing.T) {
	t.Parallel()
	usd := currency(t, "USD")
	priced := modelrouter.ModelEntry{
		ID: "vendor/synthetic", HasPrice: true, InputPrice: decimal(t, "0.0675"), OutputPrice: decimal(t, "0.225"), PriceCurrency: usd,
	}
	tests := []struct {
		name    string
		entry   modelrouter.ModelEntry
		usage   modelrouter.Usage
		cur     kernel.Currency
		want    string
		wantErr error
	}{
		{name: "a million of each", entry: priced, usage: modelrouter.Usage{InputTokens: 1_000_000, OutputTokens: 1_000_000}, cur: usd, want: "0.2925"},
		{name: "a question's calls", entry: priced, usage: modelrouter.Usage{InputTokens: 12_000, OutputTokens: 800}, cur: usd, want: "0.00099"},
		{name: "input alone", entry: priced, usage: modelrouter.Usage{InputTokens: 1_000_000}, cur: usd, want: "0.0675"},
		{name: "output alone", entry: priced, usage: modelrouter.Usage{OutputTokens: 1}, cur: usd, want: "0.000000225"},
		{name: "no tokens cost nothing", entry: priced, cur: usd, want: "0"},
		{
			name: "an entry with no price costs nothing", entry: modelrouter.ModelEntry{ID: "local"},
			usage: modelrouter.Usage{InputTokens: 5_000_000, OutputTokens: 5_000_000}, cur: usd, want: "0",
		},
		{
			name: "another currency can't be counted", entry: priced, usage: modelrouter.Usage{InputTokens: 1}, cur: currency(t, "EUR"),
			wantErr: modelrouter.ErrUnpriceable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := modelrouter.CallCost(&tt.entry, tt.usage, tt.cur)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CallCost error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !got.Equal(decimal(t, tt.want)) {
				t.Errorf("CallCost = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestBudgetExceededAtTheCap(t *testing.T) {
	t.Parallel()
	usd := currency(t, "USD")
	b, err := modelrouter.NewBudget(decimal(t, "0.50"), usd)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		budget modelrouter.Budget
		spent  string
		want   bool
	}{
		{name: "under the cap", budget: b, spent: "0.4999", want: false},
		{name: "exactly the cap is spent", budget: b, spent: "0.50", want: true},
		{name: "over the cap", budget: b, spent: "12", want: true},
		{name: "nothing spent", budget: b, spent: "0", want: false},
		{name: "no cap, nothing is exceeded", budget: modelrouter.Budget{}, spent: "9999", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.budget.Exceeded(decimal(t, tt.spent)); got != tt.want {
				t.Errorf("Exceeded(%s) = %v, want %v", tt.spent, got, tt.want)
			}
		})
	}
}

func TestNewBudget(t *testing.T) {
	t.Parallel()
	usd := currency(t, "USD")
	tests := []struct {
		name        string
		cap         string
		currency    kernel.Currency
		wantEnabled bool
		wantErr     error
	}{
		{name: "a cap", cap: "1", currency: usd, wantEnabled: true},
		{name: "zero is no budget", cap: "0", currency: usd},
		{name: "zero needs no currency", cap: "0"},
		{name: "a cap needs a currency", cap: "1", wantErr: modelrouter.ErrInvalidBudget},
		{name: "a negative cap", cap: "-1", currency: usd, wantErr: modelrouter.ErrInvalidBudget},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b, err := modelrouter.NewBudget(decimal(t, tt.cap), tt.currency)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewBudget error = %v, want %v", err, tt.wantErr)
			}
			if b.Enabled() != tt.wantEnabled {
				t.Errorf("Enabled = %v, want %v", b.Enabled(), tt.wantEnabled)
			}
		})
	}
}

func TestBudgetDayEndsAtLocalMidnight(t *testing.T) {
	t.Parallel()
	warsaw, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}
	tests := []struct {
		name     string
		at       time.Time
		loc      *time.Location
		wantDay  string
		wantEnds time.Time
	}{
		{
			name: "afternoon in Warsaw", at: time.Date(2026, time.November, 14, 13, 0, 0, 0, time.UTC), loc: warsaw,
			wantDay: "2026-11-14", wantEnds: time.Date(2026, time.November, 14, 23, 0, 0, 0, time.UTC),
		},
		{
			name: "23:30 UTC is already the next day in Warsaw", at: time.Date(2026, time.November, 14, 23, 30, 0, 0, time.UTC), loc: warsaw,
			wantDay: "2026-11-15", wantEnds: time.Date(2026, time.November, 15, 23, 0, 0, 0, time.UTC),
		},
		{
			name: "UTC", at: time.Date(2026, time.November, 14, 23, 30, 0, 0, time.UTC), loc: time.UTC,
			wantDay: "2026-11-14", wantEnds: time.Date(2026, time.November, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "the end of the year", at: time.Date(2026, time.December, 31, 12, 0, 0, 0, time.UTC), loc: time.UTC,
			wantDay: "2026-12-31", wantEnds: time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			day, ends := modelrouter.BudgetDay(tt.at, tt.loc)
			if day.String() != tt.wantDay {
				t.Errorf("day = %s, want %s", day, tt.wantDay)
			}
			if !ends.Equal(tt.wantEnds) {
				t.Errorf("ends = %s, want %s", ends.UTC(), tt.wantEnds)
			}
		})
	}
}

func TestBudgetSpentErrorIsErrBudgetSpent(t *testing.T) {
	t.Parallel()
	var err error = &modelrouter.BudgetSpentError{ResetsAt: time.Unix(1, 0)}
	if !errors.Is(err, modelrouter.ErrBudgetSpent) {
		t.Errorf("%v is not ErrBudgetSpent", err)
	}
	var spent *modelrouter.BudgetSpentError
	if !errors.As(err, &spent) || !spent.ResetsAt.Equal(time.Unix(1, 0)) {
		t.Errorf("errors.As lost the reset time: %+v", spent)
	}
}

func TestQuestionLimitsValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		limits  modelrouter.QuestionLimits
		wantErr bool
		enabled bool
	}{
		{name: "none", limits: modelrouter.QuestionLimits{}},
		{name: "D-208's", limits: modelrouter.QuestionLimits{PerHour: 10, PerDay: 30}, enabled: true},
		{name: "an hourly limit alone", limits: modelrouter.QuestionLimits{PerHour: 5}, enabled: true},
		{name: "a daily limit alone", limits: modelrouter.QuestionLimits{PerDay: 5}, enabled: true},
		{name: "negative", limits: modelrouter.QuestionLimits{PerHour: -1}, wantErr: true},
		{name: "more an hour than a day", limits: modelrouter.QuestionLimits{PerHour: 31, PerDay: 30}, wantErr: true, enabled: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.limits.Validate()
			if (err != nil) != tt.wantErr || (err != nil && !errors.Is(err, modelrouter.ErrInvalidQuestionLimits)) {
				t.Errorf("Validate error = %v, want error: %v", err, tt.wantErr)
			}
			if tt.limits.Enabled() != tt.enabled {
				t.Errorf("Enabled = %v, want %v", tt.limits.Enabled(), tt.enabled)
			}
		})
	}
}

func TestDecideQuestion(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.November, 14, 12, 0, 0, 0, time.UTC)
	ago := func(d time.Duration) time.Time { return now.Add(-d) }
	// n questions, the i-th asked i minutes ago.
	minutesAgo := func(n int) []time.Time {
		var out []time.Time
		for i := 1; i <= n; i++ {
			out = append(out, ago(time.Duration(i)*time.Minute))
		}
		return out
	}
	limits := modelrouter.QuestionLimits{PerHour: 10, PerDay: 30}
	tests := []struct {
		name       string
		times      []time.Time
		limits     modelrouter.QuestionLimits
		want       bool
		wantResets time.Time
	}{
		{name: "the first question", limits: limits, want: true},
		{name: "nine in the hour leaves one", times: minutesAgo(9), limits: limits, want: true},
		{
			// The tenth newest is 10 minutes old, and leaves the hour in 50.
			name: "ten in the hour is the limit", times: minutesAgo(10), limits: limits, want: false,
			wantResets: ago(10 * time.Minute).Add(time.Hour),
		},
		{
			name: "the hour rolls: the limit-th newest decides, not the oldest", limits: limits, want: false,
			times:      append(minutesAgo(9), ago(59*time.Minute), ago(2*time.Hour)),
			wantResets: ago(59 * time.Minute).Add(time.Hour),
		},
		{
			name: "ten questions an hour and a minute ago are not counted", limits: limits, want: true,
			times: []time.Time{
				ago(61 * time.Minute), ago(62 * time.Minute), ago(63 * time.Minute), ago(64 * time.Minute), ago(65 * time.Minute),
				ago(66 * time.Minute), ago(67 * time.Minute), ago(68 * time.Minute), ago(69 * time.Minute), ago(70 * time.Minute),
			},
		},
		{
			name: "the day: 30 in the day with the hour free", limits: limits, want: false,
			times: func() []time.Time {
				// Asked 2h, 2h30m, ... 16h30m ago: none in the hour, the oldest the 30th newest.
				out := make([]time.Time, 0, 30)
				for i := range 30 {
					out = append(out, ago(2*time.Hour+time.Duration(i)*30*time.Minute))
				}
				return out
			}(),
			wantResets: ago(2*time.Hour + 29*30*time.Minute).Add(24 * time.Hour),
		},
		{
			name: "both limits reached: it resets when both are free", limits: modelrouter.QuestionLimits{PerHour: 2, PerDay: 3}, want: false,
			times:      []time.Time{ago(10 * time.Minute), ago(20 * time.Minute), ago(23 * time.Hour)},
			wantResets: ago(23 * time.Hour).Add(24 * time.Hour),
		},
		{
			name: "only the hour reached while the day has room", limits: modelrouter.QuestionLimits{PerHour: 2, PerDay: 30}, want: false,
			times:      []time.Time{ago(10 * time.Minute), ago(20 * time.Minute)},
			wantResets: ago(20 * time.Minute).Add(time.Hour),
		},
		{name: "a question older than a day is forgotten", limits: modelrouter.QuestionLimits{PerDay: 1}, times: []time.Time{ago(25 * time.Hour)}, want: true},
		{
			name: "exactly a day old has left the window", limits: modelrouter.QuestionLimits{PerDay: 1}, times: []time.Time{ago(24 * time.Hour)}, want: true,
		},
		{name: "a question in the future is ignored", limits: modelrouter.QuestionLimits{PerHour: 1}, times: []time.Time{now.Add(time.Minute)}, want: true},
		{name: "no limit allows anything", times: minutesAgo(500), limits: modelrouter.QuestionLimits{}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := modelrouter.DecideQuestion(tt.times, now, tt.limits)
			if got.Allowed != tt.want {
				t.Fatalf("Allowed = %v, want %v", got.Allowed, tt.want)
			}
			if !got.ResetsAt.Equal(tt.wantResets) {
				t.Errorf("ResetsAt = %s, want %s", got.ResetsAt, tt.wantResets)
			}
		})
	}
}
