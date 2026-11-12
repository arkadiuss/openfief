package chartcheck_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/test/e2e/chartcheck"
)

func TestChartFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, perfFrom, date, want string
		wantErr                    error
	}{
		{name: "the performance's first day", perfFrom: "2026-03-06", date: "2026-03-13", want: "2026-03-06"},
		{name: "no performance", perfFrom: "", date: "2026-03-13", want: "2026-03-13"},
		{name: "a first day after the date", perfFrom: "2026-03-20", date: "2026-03-13", want: "2026-03-13"},
		{name: "the same day", perfFrom: "2026-03-13", date: "2026-03-13", want: "2026-03-13"},
		{name: "at most ten years back", perfFrom: "2000-01-01", date: "2026-03-13", want: "2016-03-15"},
		{name: "a bad date", perfFrom: "2026-03-06", date: "13.03.2026", wantErr: chartcheck.ErrBadDate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := chartcheck.ChartFrom(tt.perfFrom, tt.date)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ChartFrom(%q, %q) error (-want +got):\n%s", tt.perfFrom, tt.date, diff)
			}
			if got != tt.want {
				t.Errorf("ChartFrom(%q, %q) = %q, want %q", tt.perfFrom, tt.date, got, tt.want)
			}
		})
	}
}

func TestDaysBefore(t *testing.T) {
	t.Parallel()
	got, err := chartcheck.DaysBefore("2026-03-13", chartcheck.InstrumentPriceDays)
	if err != nil || got != "2025-03-13" {
		t.Errorf("DaysBefore(2026-03-13, 365) = %q, %v; want 2025-03-13", got, err)
	}
	if _, err := chartcheck.DaysBefore("tomorrow", 1); !cmp.Equal(err, chartcheck.ErrBadDate, cmpopts.EquateErrors()) {
		t.Errorf("DaysBefore(tomorrow) error = %v, want ErrBadDate", err)
	}
}

// money is a Money in PLN.
func money(amount string) *viewsv1.Money { return &viewsv1.Money{Amount: amount, Currency: "PLN"} }

func TestValueChart(t *testing.T) {
	t.Parallel()
	totals := []*viewsv1.DateTotal{
		{Date: "2026-03-12", ValueReporting: money("100.00"), NetReporting: money("95.00")},
		{Date: "2026-03-13", ValueReporting: money("110.50"), NetReporting: money("104.10"), LiquidationReporting: money("103.00")},
	}
	tests := []struct {
		name  string
		kinds []viewsv1.ValueKind
		want  chartcheck.Chart
	}{
		{
			name: "gross", kinds: []viewsv1.ValueKind{viewsv1.ValueKind_VALUE_KIND_GROSS},
			want: chartcheck.Chart{Labels: []string{"2026-03-12", "2026-03-13"}, Series: [][]string{{"100.00", "110.50"}}},
		},
		{
			name: "all three, a day without liquidation", kinds: chartcheck.ValueKinds(),
			want: chartcheck.Chart{Labels: []string{"2026-03-12", "2026-03-13"}, Series: [][]string{
				{"100.00", "110.50"}, {"95.00", "104.10"}, {"", "103.00"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, chartcheck.ValueChart(totals, tt.kinds)); diff != "" {
				t.Errorf("ValueChart (-want +got):\n%s", diff)
			}
		})
	}
	empty := chartcheck.Chart{Labels: []string{}, Series: [][]string{{}}}
	if diff := cmp.Diff(empty, chartcheck.ValueChart(nil, []viewsv1.ValueKind{viewsv1.ValueKind_VALUE_KIND_UNSPECIFIED})); diff != "" {
		t.Errorf("ValueChart of no totals (-want +got):\n%s", diff)
	}
}

func TestPerformanceChart(t *testing.T) {
	t.Parallel()
	p := &viewsv1.GetPerformanceResponse{Days: []*viewsv1.PerformanceDay{
		{Date: "2026-03-06", Index: "100"}, {Date: "2026-03-09", Index: "101.25"},
	}}
	want := chartcheck.Chart{Labels: []string{"2026-03-06", "2026-03-09"}, Series: [][]string{{"100", "101.25"}}}
	if diff := cmp.Diff(want, chartcheck.PerformanceChart(p)); diff != "" {
		t.Errorf("PerformanceChart (-want +got):\n%s", diff)
	}
}

// price is a close of listing symbol from source on date.
func price(symbol, source, date, amount string, primary bool) *viewsv1.InstrumentPrice {
	return &viewsv1.InstrumentPrice{
		Instrument: &viewsv1.Instrument{Symbol: symbol}, Source: source, Date: date,
		Close: &viewsv1.Money{Amount: amount, Currency: "GBP"}, Primary: primary,
	}
}

func TestPriceCharts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		prices []*viewsv1.InstrumentPrice
		want   []chartcheck.Chart
	}{
		{
			name: "the primary first, others by name, gaps empty, listings in order",
			prices: []*viewsv1.InstrumentPrice{
				price("TSTG", "zeta", "2026-03-12", "9.80", false),
				price("TSTG", "eodhd", "2026-03-13", "10.00", true),
				price("TSTG", "eodhd", "2026-03-12", "9.90", true),
				price("TSTG", "alpha", "2026-03-13", "10.10", false),
				price("TSTG.L", "eodhd", "2026-03-13", "11.00", true),
			},
			want: []chartcheck.Chart{
				{Labels: []string{"2026-03-12", "2026-03-13"}, Series: [][]string{
					{"9.90", "10.00"}, {"", "10.10"}, {"9.80", ""},
				}},
				{Labels: []string{"2026-03-13"}, Series: [][]string{{"11.00"}}},
			},
		},
		{
			name: "no primary: the first source by name",
			prices: []*viewsv1.InstrumentPrice{
				price("TSTH", "ibkr", "2026-03-13", "4.00", false),
				price("TSTH", "eodhd", "2026-03-13", "4.10", false),
			},
			want: []chartcheck.Chart{{Labels: []string{"2026-03-13"}, Series: [][]string{{"4.10"}, {"4.00"}}}},
		},
		{name: "no prices", want: []chartcheck.Chart{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, chartcheck.PriceCharts(tt.prices)); diff != "" {
				t.Errorf("PriceCharts (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAllocation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		resp    *viewsv1.GetAllocationResponse
		want    []chartcheck.Bar
		wantErr error
	}{
		{
			name: "shares rounded half-even",
			resp: &viewsv1.GetAllocationResponse{TotalReporting: money("800.00"), Groups: []*viewsv1.AllocationGroup{
				{Key: "equity", ValueReporting: money("500.40")},
				{Key: "cash", ValueReporting: money("299.60")},
			}},
			want: []chartcheck.Bar{
				{Key: "equity", Share: "62.6%", Width: "62.6", Value: "500.40 PLN"},
				{Key: "cash", Share: "37.4%", Width: "37.4", Value: "299.60 PLN"},
			},
		},
		{
			name: "a zero total has no share",
			resp: &viewsv1.GetAllocationResponse{TotalReporting: money("0"), Groups: []*viewsv1.AllocationGroup{
				{Key: "cash", ValueReporting: money("0")},
			}},
			want: []chartcheck.Bar{{Key: "cash", Value: "0 PLN"}},
		},
		{
			name: "a value that is no decimal",
			resp: &viewsv1.GetAllocationResponse{TotalReporting: money("1"), Groups: []*viewsv1.AllocationGroup{
				{Key: "cash", ValueReporting: money("one")},
			}},
			wantErr: chartcheck.ErrBadAmount,
		},
		{
			name: "a total that is no decimal",
			resp: &viewsv1.GetAllocationResponse{TotalReporting: money(""), Groups: []*viewsv1.AllocationGroup{
				{Key: "cash", ValueReporting: money("1")},
			}},
			wantErr: chartcheck.ErrBadAmount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := chartcheck.Allocation(tt.resp)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Allocation error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Allocation (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlainMoney(t *testing.T) {
	t.Parallel()
	polish := &viewsv1.NumberFormat{GroupSeparator: " ", DecimalSeparator: ","}
	tests := []struct {
		name, shown string
		format      *viewsv1.NumberFormat
		want        string
		wantErr     error
	}{
		{name: "space group, decimal comma", shown: "-1 234 567,80 PLN", format: polish, want: "-1234567.80 PLN"},
		{
			name: "comma group, decimal point", shown: "12,345.6789 GBP",
			format: &viewsv1.NumberFormat{GroupSeparator: ",", DecimalSeparator: "."}, want: "12345.6789 GBP",
		},
		{
			name: "point group, decimal comma", shown: "12.345,67 EUR",
			format: &viewsv1.NumberFormat{GroupSeparator: ".", DecimalSeparator: ","}, want: "12345.67 EUR",
		},
		{name: "no format: the plain one", shown: "5 000.00 PLN", format: nil, want: "5000.00 PLN"},
		{name: "a whole amount", shown: "0 PLN", format: polish, want: "0 PLN"},
		{name: "no currency", shown: "5 000,00", format: polish, wantErr: chartcheck.ErrBadAmount},
		{name: "a word", shown: "lots PLN", format: polish, wantErr: chartcheck.ErrBadAmount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := chartcheck.PlainMoney(tt.shown, tt.format)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("PlainMoney(%q) error (-want +got):\n%s", tt.shown, diff)
			}
			if got != tt.want {
				t.Errorf("PlainMoney(%q) = %q, want %q", tt.shown, got, tt.want)
			}
		})
	}
}

// coreLog is a stretch of the core's log: read API requests, the UI's
// calls, another plugin's, a line of another kind and one that is no JSON.
const coreLog = `{"time":"2026-11-08T10:00:00.5Z","level":"INFO","msg":"read API request","request_id":"a","method":"POST","status":200}
{"time":"2026-11-08T10:00:01Z","level":"INFO","msg":"auth call","procedure":"/openfief.identity.v1.AuthService/WhoAmI","plugin":"ui","code":"ok"}
{"time":"2026-11-08T10:00:01.2Z","level":"INFO","msg":"auth call","procedure":"/openfief.identity.v1.AuthService/WhoAmI","plugin":"ui","code":"ok"}
{"time":"2026-11-08T10:00:01.5Z","level":"INFO","msg":"auth call","procedure":"/openfief.ingest.v1.StatementService/ListStatementAccounts","plugin":"ui","code":"ok"}
{"time":"2026-11-08T10:00:02Z","level":"INFO","msg":"auth call","procedure":"/openfief.valuation.v1.SettingsService/GetOwnerSettings","plugin":"other","code":"ok"}
{"time":"2026-11-08T10:00:02.5Z","level":"WARN","msg":"auth call refused","procedure":"/openfief.identity.v1.AuthService/WhoAmI"}
panic: not JSON
{"time":"2026-11-08T10:00:03Z","level":"INFO","msg":"auth call","procedure":"/openfief.valuation.v1.SettingsService/GetOwnerSettings","plugin":"ui","code":"ok"}
{"time":"2026-11-08T10:00:03.5Z","level":"INFO","msg":"read API request","request_id":"b","method":"GET","status":200}
`

// at is a time on 2026-11-08 at 10:00 and ms milliseconds.
func at(ms int) time.Time {
	return time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC).Add(time.Duration(ms) * time.Millisecond)
}

func TestCalls(t *testing.T) {
	t.Parallel()
	calls, err := chartcheck.Calls([]byte(coreLog))
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 7 {
		t.Fatalf("Calls read %d lines, want 7: %+v", len(calls), calls)
	}
	tests := []struct {
		name      string
		window    chartcheck.Window
		wantProcs []string
		wantReads int
	}{
		{
			name: "the first three seconds", window: chartcheck.Window{From: at(0), To: at(2000)},
			wantProcs: []string{"/openfief.identity.v1.AuthService/WhoAmI", "/openfief.ingest.v1.StatementService/ListStatementAccounts"},
			wantReads: 1,
		},
		{
			name: "both ends included", window: chartcheck.Window{From: at(1500), To: at(3000)},
			wantProcs: []string{"/openfief.ingest.v1.StatementService/ListStatementAccounts", "/openfief.valuation.v1.SettingsService/GetOwnerSettings"},
		},
		{name: "nothing in it", window: chartcheck.Window{From: at(4000), To: at(5000)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.wantProcs, chartcheck.PluginCalls(calls, "ui", tt.window), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("PluginCalls (-want +got):\n%s", diff)
			}
			if got := chartcheck.ReadRequests(calls, tt.window); got != tt.wantReads {
				t.Errorf("ReadRequests = %d, want %d", got, tt.wantReads)
			}
		})
	}
	if _, err := chartcheck.Calls([]byte(`{"time":"yesterday","msg":"auth call"}`)); !cmp.Equal(err, chartcheck.ErrBadLog, cmpopts.EquateErrors()) {
		t.Errorf("Calls of a line without a time: error %v, want ErrBadLog", err)
	}
}

func TestDisallowed(t *testing.T) {
	t.Parallel()
	allowed := []string{"/openfief.identity.v1.AuthService/", "/openfief.ingest.v1.StatementService/ListStatementAccounts"}
	tests := []struct {
		name  string
		procs []string
		want  []string
	}{
		{name: "a whole service and one procedure", procs: []string{
			"/openfief.identity.v1.AuthService/WhoAmI", "/openfief.ingest.v1.StatementService/ListStatementAccounts",
		}},
		{name: "another procedure of the same service", procs: []string{
			"/openfief.ingest.v1.StatementService/UploadStatement", "/openfief.identity.v1.AuthService/Logout",
		}, want: []string{"/openfief.ingest.v1.StatementService/UploadStatement"}},
		{
			name: "a service named like an allowed one", procs: []string{"/openfief.identity.v1.AuthServiceX/WhoAmI"},
			want: []string{"/openfief.identity.v1.AuthServiceX/WhoAmI"},
		},
		{name: "none"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, chartcheck.Disallowed(tt.procs, allowed), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Disallowed (-want +got):\n%s", diff)
			}
		})
	}
}
