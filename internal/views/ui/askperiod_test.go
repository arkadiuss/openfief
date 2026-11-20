package ui

import (
	"net/http"
	"net/url"
	"slices"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/views/app/answer"
)

// TestAskPeriodCards checks the Ask box's answers for a bounded period over the
// fake core (D-194): the card of each command with the core's own figures, and
// what the read API was asked.
func TestAskPeriodCards(t *testing.T) {
	t.Parallel()
	tests := []struct {
		q    string
		want []string
		from string
		to   string
	}{
		{
			q: "Performance in November 2025", from: "2025-11-01", to: "2025-11-30",
			want: []string{
				"Time-weighted return in November 2025: -1,20%", "From the end of 2025-10-31 to the end of 2025-11-30, gross.",
				"Start", "56\u202f719,52 PLN", "End", "58\u202f686,37 PLN", "Deposits", "2\u202f000 PLN", "Withdrawals", "Show the performance chart",
				"start=2025-10-31",
			},
		},
		{q: "return from 2025-11-01 to 2025-11-30", from: "2025-11-01", to: "2025-11-30", want: []string{"Time-weighted return from 2025-11-01 to 2025-11-30: -1,20%"}},
		{q: "twr 2025-11", from: "2025-11-01", to: "2025-11-30", want: []string{"Time-weighted return in November 2025: -1,20%"}},
		{
			q: "Value change in November 2025", from: "2025-11-01", to: "2025-11-30",
			want: []string{
				"Value rose by 1\u202f966,85 PLN in November 2025, gross", "56\u202f719,52 PLN to 58\u202f686,37 PLN", "&#43;1\u202f966,85 PLN (&#43;3,47%)",
				"Net", "Liquidation", "Net deposits", "paid in 2\u202f000 PLN, taken out 0 PLN", "&#43;2\u202f000 PLN", "Open the overview on 2025-11-30",
			},
		},
		{q: "value in November 2025", from: "2025-11-01", to: "2025-11-30", want: []string{"Value rose by 1\u202f966,85 PLN in November 2025, gross"}},
		{q: "value from 2025-11-01 to 2025-11-30", from: "2025-11-01", to: "2025-11-30", want: []string{"Value rose by 1\u202f966,85 PLN from 2025-11-01 to 2025-11-30, gross"}},
		{q: "change in value between 2025-11-01 and 2025-11-30", from: "2025-11-01", to: "2025-11-30", want: []string{"Value rose by 1\u202f966,85 PLN"}},
		{q: "value for 2025", from: "2025-01-01", to: "2025-12-31", want: []string{"Value rose by 1\u202f966,85 PLN in 2025, gross"}},
	}
	for _, tt := range tests {
		t.Run(tt.q, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			rec := h.serve(t, htmx(withSession(request(http.MethodGet, "/ask?"+url.Values{"q": {tt.q}}.Encode(), ""), readSession)))
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec, append([]string{`class="bubble"`, "Answered from your data · no AI model"}, tt.want...)...)
			var asked []string
			h.read.locked(func() {
				for _, r := range h.read.periods {
					asked = append(asked, r.GetFrom()+" "+r.GetTo())
				}
			})
			if diff := cmp.Diff([]string{tt.from + " " + tt.to}, asked); diff != "" {
				t.Errorf("GetPeriodChange asked (-want +got):\n%s", diff)
			}
		})
	}
}

// TestAskPeriodRefusals checks that a period the box can't answer is a clear
// card and never a guess (D-194): a reversed one, one past the latest valued
// day (2026-03-31) and one the core refuses are told apart, and none of the
// first two asks the core.
func TestAskPeriodRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		q      string
		fail   error
		want   []string
		asked  int
		wantNo []string
	}{
		{
			q:    "performance from 2025-11-30 to 2025-11-01",
			want: []string{"That period can&#39;t be answered", "The period ends on 2025-11-01, before it starts on 2025-11-30."}, wantNo: []string{"Time-weighted return"},
		},
		{
			q:    "value change from 2026-03-30 to 2026-04-02",
			want: []string{"That period can&#39;t be answered", "valued only through 2026-03-31."}, wantNo: []string{"Value rose", "Value fell"},
		},
		{
			q:    "performance in December 2026",
			want: []string{"The period ends on 2026-12-31, but your accounts are valued only through 2026-03-31."},
		},
		{
			q:    "value for 2026",
			want: []string{"The period ends on 2026-12-31, but your accounts are valued only through 2026-03-31."},
		},
		{
			q: "performance in November 2025", fail: codeErr(connect.CodeFailedPrecondition, nil), asked: 1,
			want: []string{"Your accounts are not valued for all of 2025-11-01 to 2025-11-30 yet"},
		},
		{
			q: "value change in November 2025", fail: codeErr(connect.CodeInvalidArgument, nil), asked: 1,
			want: []string{"The read API doesn&#39;t accept 2025-11-01 to 2025-11-30 as a period"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.q, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.read.locked(func() { h.read.periodErr = tt.fail })
			rec := h.serve(t, htmx(withSession(request(http.MethodGet, "/ask?"+url.Values{"q": {tt.q}}.Encode(), ""), readSession)))
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec, tt.want...)
			wantNoBody(t, rec, tt.wantNo...)
			var asked int
			h.read.locked(func() { asked = len(h.read.periods) })
			if asked != tt.asked {
				t.Errorf("the core was asked %d times, want %d", asked, tt.asked)
			}
		})
	}
}

// TestAskPeriodCardWithAMissingBasis checks that a basis the core leaves out is
// missing on the card and not filled in.
func TestAskPeriodCardWithAMissingBasis(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.read.locked(func() { h.read.periodMissing = true })
	rec := h.serve(t, htmx(withSession(request(http.MethodGet, "/ask?"+url.Values{"q": {"value change in November 2025"}}.Encode(), ""), readSession)))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Liquidation", "missing", "A revaluation is pending, so these figures may change.")
}

// TestTheModelsPeriodCallsAskWhatTheTypedCommandsAsk checks that the model's
// value and performance tools, called for a period, run the very query a typed
// command of the same period makes (D-203): the box's tools are its commands'
// answerers and nothing more.
func TestTheModelsPeriodCallsAskWhatTheTypedCommandsAsk(t *testing.T) {
	t.Parallel()
	f := newLoopFixture(t, tierT0, tierT2)
	var mu sync.Mutex
	var ran []askQuery
	run := f.a.run
	f.a.run = func(r *http.Request, s session, q *askQuery) (answerCard, error) {
		mu.Lock()
		ran = append(ran, *q)
		mu.Unlock()
		return run(r, s, q)
	}
	f.port.replies = []askReply{
		calling(
			askCall{ID: "1", Name: "value", Args: `{"period":{"from":"2025-11-01","to":"2025-11-30"}}`},
			askCall{ID: "2", Name: "performance", Args: `{"month":"2025-11"}`},
			askCall{ID: "3", Name: "performance", Args: `{"from":"2025-11-01","to":"2025-11-30"}`},
		),
		say("done", "c1", "c2", "c3"),
	}
	if _, err := f.ask("how did November go?"); err != nil {
		t.Fatal(err)
	}
	today := mustDate(t, askToday)
	want := make([]askQuery, 0, 3)
	for _, typed := range []string{"value change from 2025-11-01 to 2025-11-30", "performance in november 2025", "performance from 2025-11-01 to 2025-11-30"} {
		q, ok := answer.Parse(typed, today)
		if !ok {
			t.Fatalf("answer.Parse(%q) is no command", typed)
		}
		want = append(want, q)
	}
	order := func(qs []askQuery) {
		slices.SortFunc(qs, func(a, b askQuery) int {
			return int(a.Command)*100 + int(a.Period.Kind) - int(b.Command)*100 - int(b.Period.Kind)
		})
	}
	order(want)
	order(ran)
	if diff := cmp.Diff(want, ran); diff != "" {
		t.Errorf("queries run by the tools against the typed commands' (-typed +tools):\n%s", diff)
	}
}

// TestEveryPerformancePeriodIsOnePeriodCard checks that each way of naming a
// performance period, in a typed command and as the model's tool arguments,
// asks the read API's GetPeriodChange for the days the period resolves to and
// answers with the same card, which opens the period at the end of the day
// before its first (D-61, D-194, D-203). The latest valued day is 2026-03-31 and
// the tracking began on 2026-03-30 (the fake core's default start).
func TestEveryPerformancePeriodIsOnePeriodCard(t *testing.T) {
	t.Parallel()
	tests := []struct {
		q, args  string
		from, to string
		title    string
		// same is false for a command whose tool arguments name the period by
		// default, which the typed command words differently.
		same bool
	}{
		{"performance", `{}`, "2026-01-01", "2026-03-31", "Time-weighted return this year: -1,20%", false},
		{"performance this year", `{"period":"this_year"}`, "2026-01-01", "2026-03-31", "Time-weighted return this year: -1,20%", true},
		{"performance this month", `{"period":"this_month"}`, "2026-03-01", "2026-03-31", "Time-weighted return this month: -1,20%", true},
		{"performance last 12 months", `{"period":"last_12_months"}`, "2025-04-01", "2026-03-31", "Time-weighted return over the last 12 months: -1,20%", true},
		{"performance last year", `{"period":"last_year"}`, "2025-01-01", "2025-12-31", "Time-weighted return last year: -1,20%", true},
		{"performance in 2025", `{"year":2025}`, "2025-01-01", "2025-12-31", "Time-weighted return in 2025: -1,20%", true},
		{"performance in 2026", `{"year":2026}`, "2026-01-01", "2026-03-31", "Time-weighted return in 2026: -1,20%", true},
		{"performance since 2026-01-15", `{"since":"2026-01-15"}`, "2026-01-15", "2026-03-31", "Time-weighted return since 2026-01-15: -1,20%", true},
		{"performance all", `{"period":"all"}`, "2026-03-30", "2026-03-31", "Time-weighted return since tracking began: -1,20%", true},
		{"performance in November 2025", `{"month":"2025-11"}`, "2025-11-01", "2025-11-30", "Time-weighted return in November 2025: -1,20%", true},
		{
			"performance from 2025-11-01 to 2025-11-30", `{"from":"2025-11-01","to":"2025-11-30"}`,
			"2025-11-01", "2025-11-30", "Time-weighted return from 2025-11-01 to 2025-11-30: -1,20%", true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.q, func(t *testing.T) {
			t.Parallel()
			if tt.same {
				wantToolAsksWhatTheCommandAsks(t, tt.q, tt.args)
			}
			h := newHarness(t)
			rec := h.serve(t, htmx(withSession(request(http.MethodGet, "/ask?"+url.Values{"q": {tt.q}}.Encode(), ""), readSession)))
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec, tt.title, "From the end of ", "Start", "56\u202f719,52 PLN", "End", "58\u202f686,37 PLN", "Deposits", "Withdrawals", "Show the performance chart")
			// The card's link opens the chart on the period's opening close, the
			// day before its first day, so both show the same return (D-223).
			opening := mustDate(t, tt.from).AddDays(-1).String()
			wantBody(t, rec, `href="/?over=performance&amp;start=`+opening+`#over-time">Show the performance chart`)
			if diff := cmp.Diff([]string{tt.from + " " + tt.to}, periodsAsked(h)); diff != "" {
				t.Errorf("GetPeriodChange asked (-want +got):\n%s", diff)
			}
		})
	}
}

// wantToolAsksWhatTheCommandAsks fails unless the performance tool, given args,
// runs the query the typed command q makes.
func wantToolAsksWhatTheCommandAsks(t *testing.T, q, args string) {
	t.Helper()
	today := mustDate(t, askToday)
	typed, ok := answer.Parse(q, today)
	if !ok {
		t.Fatalf("answer.Parse(%q) is no command", q)
	}
	var performance answer.Tool
	for _, tool := range answer.Tools() {
		if tool.Name == "performance" {
			performance = tool
		}
	}
	called, err := performance.Query(args, today)
	if err != nil {
		t.Fatalf("the performance tool refused %s: %v", args, err)
	}
	if diff := cmp.Diff(typed, called); diff != "" {
		t.Errorf("tool against typed command (-typed +tool):\n%s", diff)
	}
}

// periodsAsked returns the periods the fake core was asked for, as "from to".
func periodsAsked(h *harness) []string {
	var asked []string
	h.read.locked(func() {
		for _, r := range h.read.periods {
			asked = append(asked, r.GetFrom()+" "+r.GetTo())
		}
	})
	return asked
}

// TestAPerformanceWithNothingRecordedHasNoStart checks that the period of
// everything, with nothing recorded, says so and asks for no period.
func TestAPerformanceWithNothingRecordedHasNoStart(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.read.locked(func() { h.read.noPerformanceStart = true })
	rec := h.serve(t, htmx(withSession(request(http.MethodGet, "/ask?"+url.Values{"q": {"performance since tracking began"}}.Encode(), ""), readSession)))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "No valued days since tracking began")
	var asked int
	h.read.locked(func() { asked = len(h.read.periods) })
	if asked != 0 {
		t.Errorf("the core was asked %d periods, want none", asked)
	}
}
