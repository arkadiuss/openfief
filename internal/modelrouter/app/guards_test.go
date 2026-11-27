package app_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
)

func decimalOf(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func usd(t *testing.T) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency("USD")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// fakeUsage is an in-memory UsageStore, by day, that fails when spentErr or
// recordErr is set.
type fakeUsage struct {
	mu        sync.Mutex
	byDay     map[string]kernel.Decimal
	records   []app.UsageRecord
	spentErr  error
	recordErr error
	reads     int
}

func (f *fakeUsage) Spent(_ context.Context, day kernel.Date, _ kernel.Currency) (kernel.Decimal, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reads++
	return f.byDay[day.String()], f.spentErr
}

func (f *fakeUsage) Record(_ context.Context, u *app.UsageRecord) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.recordErr != nil {
		return f.recordErr
	}
	if f.byDay == nil {
		f.byDay = map[string]kernel.Decimal{}
	}
	f.byDay[u.Day.String()] = f.byDay[u.Day.String()].Add(u.Cost)
	f.records = append(f.records, *u)
	return nil
}

// guardedFixture is a router fixture whose cloud model is priced at USD 1 in
// and 2 out per million tokens (a call of the fakes, 3 tokens in and 2 out,
// costs USD 0.000007), with guards on its router.
func guardedFixture(t *testing.T, g app.Guards) *routerFixture {
	t.Helper()
	f := newRouterFixture(t)
	f.providers.locations = map[string]modelrouter.Location{"ollama": modelrouter.LocationOwnerHardware}
	f.providers.entries["openrouter"] = []modelrouter.ModelEntry{{
		ID: "qwen", ToolCalling: true, StructuredOut: true,
		HasPrice: true, InputPrice: decimalOf(t, "1"), OutputPrice: decimalOf(t, "2"), PriceCurrency: usd(t),
	}}
	f.rebuild(g)
	return f
}

func (f *routerFixture) rebuild(g app.Guards) {
	f.router = app.NewRouter(&app.RouterPorts{
		Settings: f.store, Providers: f.providers, Audit: f.audit, Attention: f.attention, CallIDs: callIDs(), Now: f.clock.now, Guards: g,
	})
}

func budgetOf(t *testing.T, limit string) modelrouter.Budget {
	t.Helper()
	b, err := modelrouter.NewBudget(decimalOf(t, limit), usd(t))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestRouterRefusesBeforeCallingOnceTheDaysBudgetIsSpent(t *testing.T) {
	t.Parallel()
	usage := &fakeUsage{}
	// A call costs USD 0.000007: the second call reaches the cap of 0.000014.
	f := guardedFixture(t, app.Guards{Budget: budgetOf(t, "0.000014"), Usage: usage})
	for i := range 2 {
		if _, err := f.generate(request(modelrouter.DataTierT2)); err != nil {
			t.Fatalf("call %d: %v", i+1, err)
		}
	}
	before := len(f.providers.called)
	_, err := f.generate(request(modelrouter.DataTierT2))
	if !errors.Is(err, modelrouter.ErrBudgetSpent) {
		t.Fatalf("error = %v, want ErrBudgetSpent", err)
	}
	spent, ok := errors.AsType[*modelrouter.BudgetSpentError](err)
	if !ok {
		t.Fatalf("error %v carries no reset time", err)
	}
	// The day is the clock's, 2026-01-02 in UTC, and ends at its midnight.
	if want := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC); !spent.ResetsAt.Equal(want) {
		t.Errorf("resets at %s, want %s", spent.ResetsAt, want)
	}
	if len(f.providers.called) != before {
		t.Errorf("a call reached a provider after the budget was spent: %v", f.providers.called[before:])
	}
	f.wantOneRefusalOfClassBudget(2)
	if got := usage.byDay["2026-01-02"]; !got.Equal(decimalOf(t, "0.000014")) {
		t.Errorf("the day's total = %s, want 0.000014", got)
	}
}

// wantOneRefusalOfClassBudget checks that the last audit entry is a refusal of
// class budget with no model, and that only the calls that were made, sent,
// have a sent entry.
func (f *routerFixture) wantOneRefusalOfClassBudget(sent int) {
	f.t.Helper()
	last := f.audit.results[len(f.audit.results)-1]
	if last.Outcome != app.OutcomeRefused || last.Class != app.ClassBudget || !last.Model.IsZero() {
		f.t.Errorf("the audit entry says %+v, want a refusal of class budget and no model", last)
	}
	if len(f.audit.sent) != sent {
		f.t.Errorf("sent entries = %d, want only the %d calls that were made", len(f.audit.sent), sent)
	}
}

func TestRouterCountsEveryRolesCallsAndStartsTheNextDayOver(t *testing.T) {
	t.Parallel()
	usage := &fakeUsage{}
	f := guardedFixture(t, app.Guards{Budget: budgetOf(t, "0.000007"), Usage: usage})
	if _, err := f.generate(request(modelrouter.DataTierT2)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.generate(request(modelrouter.DataTierT2)); !errors.Is(err, modelrouter.ErrBudgetSpent) {
		t.Fatalf("error = %v, want ErrBudgetSpent", err)
	}
	f.clock.t = f.clock.t.Add(24 * time.Hour)
	if _, err := f.generate(request(modelrouter.DataTierT2)); err != nil {
		t.Errorf("the next day's first call: %v", err)
	}
	if len(usage.records) != 2 || usage.records[1].Day.String() != "2026-01-03" {
		t.Errorf("records = %+v, want the second on 2026-01-03", usage.records)
	}
	if got := usage.records[0]; got.Model != f.cloud || got.Usage != (modelrouter.Usage{InputTokens: 3, OutputTokens: 2}) || got.Currency != usd(t) {
		t.Errorf("record = %+v", got)
	}
}

func TestRouterCountsTheDayInItsZone(t *testing.T) {
	t.Parallel()
	warsaw, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatal(err)
	}
	usage := &fakeUsage{}
	f := guardedFixture(t, app.Guards{Budget: budgetOf(t, "1"), Usage: usage, Zone: warsaw})
	// 23:30 UTC on the 2nd is already the 3rd in Warsaw.
	f.clock.t = time.Date(2026, 1, 2, 23, 30, 0, 0, time.UTC)
	if _, err := f.generate(request(modelrouter.DataTierT2)); err != nil {
		t.Fatal(err)
	}
	if got := usage.records[0].Day.String(); got != "2026-01-03" {
		t.Errorf("day = %s, want 2026-01-03", got)
	}
}

func TestRouterWithoutABudgetReadsAndRecordsNothing(t *testing.T) {
	t.Parallel()
	usage := &fakeUsage{}
	f := guardedFixture(t, app.Guards{Usage: usage})
	for range 3 {
		if _, err := f.generate(request(modelrouter.DataTierT2)); err != nil {
			t.Fatal(err)
		}
	}
	if usage.reads != 0 || len(usage.records) != 0 {
		t.Errorf("reads = %d, records = %d, want none without a budget", usage.reads, len(usage.records))
	}
}

func TestRouterCountsAnEntryWithNoPriceAsFree(t *testing.T) {
	t.Parallel()
	usage := &fakeUsage{}
	f := guardedFixture(t, app.Guards{Budget: budgetOf(t, "0.000001"), Usage: usage})
	// T0 data goes to the local model only, which lists no price.
	for range 3 {
		if _, err := f.generate(request(modelrouter.DataTierT0)); err != nil {
			t.Fatalf("a free call: %v", err)
		}
	}
	if !usage.byDay["2026-01-02"].IsZero() {
		t.Errorf("total = %s, want 0", usage.byDay["2026-01-02"])
	}
}

// TestRouterRefusesACallItCannotPriceUnderABudget checks that a budget never
// lets a call through that it couldn't count: a model the catalog has no entry
// for, which is every model after a restart until ListModels answers, and a
// third party's model whose entry carries no price. A model on the owner's
// hardware stays free.
func TestRouterRefusesACallItCannotPriceUnderABudget(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mod  func(f *routerFixture)
		tier modelrouter.DataTier
	}{
		{"an empty catalog", func(f *routerFixture) { f.providers.entries = map[string][]modelrouter.ModelEntry{} }, modelrouter.DataTierT2},
		{"a provider that can't list its models yet", func(f *routerFixture) {
			f.providers.listErr = errBoom
		}, modelrouter.DataTierT2},
		{"a cloud model's entry with no price", func(f *routerFixture) {
			f.providers.entries["openrouter"][0].HasPrice = false
		}, modelrouter.DataTierT2},
		{"a local model the catalog doesn't list", func(f *routerFixture) {
			f.providers.entries["ollama"] = nil
		}, modelrouter.DataTierT0},
		{"an unpriced model in a third party's cloud", func(f *routerFixture) {
			f.providers.locations = nil
		}, modelrouter.DataTierT0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			usage := &fakeUsage{}
			f := guardedFixture(t, app.Guards{Budget: budgetOf(t, "1"), Usage: usage})
			tt.mod(f)
			f.rebuild(app.Guards{Budget: budgetOf(t, "1"), Usage: usage})
			_, err := f.generate(request(tt.tier))
			if !errors.Is(err, modelrouter.ErrNoPrice) {
				t.Fatalf("error = %v, want ErrNoPrice", err)
			}
			if len(f.providers.called) != 0 {
				t.Errorf("a call that can't be counted was sent: %v", f.providers.called)
			}
			if got := f.audit.results[0]; got.Outcome != app.OutcomeRefused || got.Class != app.ClassBudget {
				t.Errorf("audit entry = %+v, want a refusal of class budget", got)
			}
			if len(usage.records) != 0 {
				t.Errorf("records = %v, want none", usage.records)
			}
		})
	}
}

func TestRouterRefusesACandidatePricedInAnotherCurrency(t *testing.T) {
	t.Parallel()
	usage := &fakeUsage{}
	f := guardedFixture(t, app.Guards{Budget: budgetOf(t, "1"), Usage: usage})
	eur, err := kernel.ParseCurrency("EUR")
	if err != nil {
		t.Fatal(err)
	}
	f.providers.entries["openrouter"][0].PriceCurrency = eur
	f.router = app.NewRouter(&app.RouterPorts{
		Settings: f.store, Providers: f.providers, Audit: f.audit, Attention: f.attention, CallIDs: callIDs(), Now: f.clock.now,
		Guards: app.Guards{Budget: budgetOf(t, "1"), Usage: usage},
	})
	_, err = f.generate(request(modelrouter.DataTierT2))
	if !errors.Is(err, modelrouter.ErrUnpriceable) {
		t.Fatalf("error = %v, want ErrUnpriceable", err)
	}
	if len(f.providers.called) != 0 {
		t.Errorf("a call that can't be counted was sent: %v", f.providers.called)
	}
	if got := f.audit.results[0]; got.Outcome != app.OutcomeRefused || got.Class != app.ClassBudget {
		t.Errorf("audit entry = %+v", got)
	}
}

func TestRouterFailsClosedWhenTheDaysUsageCannotBeRead(t *testing.T) {
	t.Parallel()
	usage := &fakeUsage{spentErr: errBoom}
	f := guardedFixture(t, app.Guards{Budget: budgetOf(t, "1"), Usage: usage})
	_, err := f.generate(request(modelrouter.DataTierT2))
	if !errors.Is(err, app.ErrUsageUnavailable) {
		t.Fatalf("error = %v, want ErrUsageUnavailable", err)
	}
	if len(f.providers.called) != 0 {
		t.Errorf("a call was sent without knowing the budget: %v", f.providers.called)
	}
}

func TestRouterWithholdsAnAnswerItCannotCount(t *testing.T) {
	t.Parallel()
	usage := &fakeUsage{recordErr: errBoom}
	f := guardedFixture(t, app.Guards{Budget: budgetOf(t, "1"), Usage: usage})
	res, err := f.generate(request(modelrouter.DataTierT2))
	if !errors.Is(err, app.ErrUsageUnavailable) || res.Content != "" {
		t.Fatalf("result = %+v, error = %v, want it withheld with ErrUsageUnavailable", res, err)
	}
	last := f.audit.results[len(f.audit.results)-1]
	if last.Outcome != app.OutcomeFailed || last.Class != app.ClassInternal {
		t.Errorf("audit entry = %+v, want a failed call of class internal", last)
	}
}

func TestRouterBudgetKeepsTextOutOfRecordsAndErrors(t *testing.T) {
	t.Parallel()
	usage := &fakeUsage{}
	f := guardedFixture(t, app.Guards{Budget: budgetOf(t, "0.000007"), Usage: usage})
	if _, err := f.generate(request(modelrouter.DataTierT2)); err != nil {
		t.Fatal(err)
	}
	_, err := f.generate(request(modelrouter.DataTierT2))
	if err == nil || strings.Contains(err.Error(), "SECRET-QUESTION-TEXT") {
		t.Fatalf("error = %v, want one that quotes no question", err)
	}
}

// blockingProviders is fakeProviders whose Generate waits to be released, and
// says when a call has started.
type blockingProviders struct {
	*fakeProviders

	started chan struct{}
	release chan struct{}
}

func (b *blockingProviders) Generate(ctx context.Context, provider, model string, req *modelrouter.Request) (modelrouter.Result, error) {
	b.started <- struct{}{}
	<-b.release
	return b.fakeProviders.Generate(ctx, provider, model, req)
}

// gatedFixture is a router fixture with two calls at once and a short wait, over
// providers that block until released.
func gatedFixture(t *testing.T, wait time.Duration) (*routerFixture, *blockingProviders) {
	t.Helper()
	f := newRouterFixture(t)
	bp := &blockingProviders{fakeProviders: f.providers, started: make(chan struct{}, 8), release: make(chan struct{}, 8)}
	f.router = app.NewRouter(&app.RouterPorts{
		Settings: f.store, Providers: bp, Audit: f.audit, Attention: f.attention, CallIDs: callIDs(), Now: f.clock.now,
		Guards: app.Guards{AskCalls: 2, QueueWait: wait},
	})
	return f, bp
}

// generateAsync asks in a goroutine and returns what comes of it.
func (f *routerFixture) generateAsync(ctx context.Context) <-chan error {
	done := make(chan error, 1)
	go func() {
		_, err := f.router.Generate(ctx, modelrouter.RoleAsk, owner(), request(modelrouter.DataTierT2))
		done <- err
	}()
	return done
}

func TestRouterRefusesAThirdCallAtOnceAfterTheQueueWait(t *testing.T) {
	t.Parallel()
	f, bp := gatedFixture(t, 20*time.Millisecond)
	first, second := f.generateAsync(t.Context()), f.generateAsync(t.Context())
	<-bp.started
	<-bp.started
	_, err := f.router.Generate(t.Context(), modelrouter.RoleAsk, owner(), request(modelrouter.DataTierT2))
	if !errors.Is(err, app.ErrBusy) {
		t.Fatalf("third call error = %v, want ErrBusy", err)
	}
	if got := len(f.providers.called); got != 0 {
		// Both calls are still blocked in the provider: none has recorded a call yet.
		t.Errorf("provider calls recorded = %d while blocked", got)
	}
	f.audit.mu.Lock()
	var refused []app.CallResult
	for _, r := range f.audit.results {
		if r.Outcome == app.OutcomeRefused {
			refused = append(refused, r)
		}
	}
	f.audit.mu.Unlock()
	if len(refused) != 1 || refused[0].Class != app.ClassBusy {
		t.Errorf("refusals = %+v, want one of class busy", refused)
	}
	bp.release <- struct{}{}
	bp.release <- struct{}{}
	for _, c := range []<-chan error{first, second} {
		if err := <-c; err != nil {
			t.Errorf("a call within the limit failed: %v", err)
		}
	}
}

func TestRouterQueuesAThirdCallForAFreedSlot(t *testing.T) {
	t.Parallel()
	f, bp := gatedFixture(t, time.Minute)
	first, second := f.generateAsync(t.Context()), f.generateAsync(t.Context())
	<-bp.started
	<-bp.started
	third := f.generateAsync(t.Context())
	bp.release <- struct{}{} // one of the two finishes and frees its slot: the third takes it
	<-bp.started             // the third call reached the provider, so it did not wait out its minute
	bp.release <- struct{}{}
	bp.release <- struct{}{}
	for name, c := range map[string]<-chan error{"first": first, "second": second, "third": third} {
		if err := <-c; err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestRouterReleasesTheSlotWhenACallFails(t *testing.T) {
	t.Parallel()
	f, bp := gatedFixture(t, 20*time.Millisecond)
	f.providers.answers["openrouter/qwen"] = app.ErrProviderRejected
	for range 3 {
		bp.release <- struct{}{}
		if _, err := f.router.Generate(t.Context(), modelrouter.RoleAsk, owner(), request(modelrouter.DataTierT2)); !errors.Is(err, app.ErrProviderRejected) {
			t.Fatalf("error = %v, want the provider's rejection, never ErrBusy", err)
		}
		<-bp.started
	}
}

func TestRouterWaitEndsWithTheCallersContext(t *testing.T) {
	t.Parallel()
	f, bp := gatedFixture(t, time.Minute)
	first, second := f.generateAsync(t.Context()), f.generateAsync(t.Context())
	<-bp.started
	<-bp.started
	ctx, cancel := context.WithCancel(t.Context())
	third := f.generateAsync(ctx)
	cancel()
	if err := <-third; !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want the context's", err)
	}
	bp.release <- struct{}{}
	bp.release <- struct{}{}
	for _, c := range []<-chan error{first, second} {
		if err := <-c; err != nil {
			t.Errorf("a call within the limit failed: %v", err)
		}
	}
	for _, r := range f.audit.results {
		if r.Outcome == app.OutcomeRefused {
			t.Errorf("a caller that gave up while waiting was audited as refused: %+v", r)
		}
	}
}

func TestRouterWithNoLimitOnCallsAtOnceNeverWaits(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	for range 10 {
		if _, err := f.generate(request(modelrouter.DataTierT2)); err != nil {
			t.Fatal(err)
		}
	}
}

// fakeQuestions is a QuestionStore that applies modelrouter.DecideQuestion to
// the questions it holds, per subject, as the database's store does.
type fakeQuestions struct {
	mu       sync.Mutex
	sessions map[string][]time.Time
	address  map[string][]time.Time
	seen     []app.QuestionAsked
	err      error
}

func (f *fakeQuestions) Admit(_ context.Context, q *app.QuestionAsked) (app.QuestionRuling, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seen = append(f.seen, *q)
	if f.err != nil {
		return app.QuestionRuling{}, f.err
	}
	if f.sessions == nil {
		f.sessions, f.address = map[string][]time.Time{}, map[string][]time.Time{}
	}
	if v := modelrouter.DecideQuestion(f.sessions[q.Session], q.At, q.Limits); !v.Allowed {
		return app.QuestionRuling{Subject: app.SubjectSession, ResetsAt: v.ResetsAt}, nil
	}
	if q.Address != "" {
		if v := modelrouter.DecideQuestion(f.address[q.Address], q.At, q.Limits); !v.Allowed {
			return app.QuestionRuling{Subject: app.SubjectAddress, ResetsAt: v.ResetsAt}, nil
		}
		f.address[q.Address] = append(f.address[q.Address], q.At)
	}
	f.sessions[q.Session] = append(f.sessions[q.Session], q.At)
	return app.QuestionRuling{Allowed: true}, nil
}

func newAdmitFixture(t *testing.T, g app.Guards, questions app.QuestionStore, limits modelrouter.QuestionLimits) (*routerFixture, *app.Ask) {
	t.Helper()
	f := guardedFixture(t, g)
	askers := fakeAskers{
		askerToken: {Kind: app.ActorOwner, ID: "session-owner"}, readerToken: {Kind: app.ActorViewer, ID: "session-read"},
	}
	return f, app.NewAsk(askers, f.store, f.router).WithQuestionLimits(questions, limits)
}

func TestAskAdmitCountsQuestionsPerSessionAndTellsWhenItResets(t *testing.T) {
	t.Parallel()
	qs := &fakeQuestions{}
	f, ask := newAdmitFixture(t, app.Guards{}, qs, modelrouter.QuestionLimits{PerHour: 2, PerDay: 30})
	for i := range 2 {
		got, err := ask.Admit(t.Context(), "https://demo", askerToken, "198.51.100.7")
		if err != nil || got.Decision != app.AdmitAllowed {
			t.Fatalf("question %d: %+v, %v", i+1, got, err)
		}
		f.clock.t = f.clock.t.Add(10 * time.Minute)
	}
	got, err := ask.Admit(t.Context(), "https://demo", askerToken, "198.51.100.7")
	if err != nil {
		t.Fatal(err)
	}
	// The first of the two questions was asked 20 minutes ago and leaves the hour in 40.
	want := app.Admission{Decision: app.AdmitSessionLimit, ResetsAt: f.clock.t.Add(40 * time.Minute)}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("admission (-want +got):\n%s", diff)
	}
	// Another session of the same address is stopped by the address.
	other, err := ask.Admit(t.Context(), "https://demo", readerToken, "198.51.100.7")
	if err != nil {
		t.Fatal(err)
	}
	if other.Decision != app.AdmitAddressLimit {
		t.Errorf("a second session at the same address: %+v, want the address limit", other)
	}
	// And a question stopped is not counted: the same session at a new address is still stopped by its own count.
	third, err := ask.Admit(t.Context(), "https://demo", askerToken, "203.0.113.9")
	if err != nil || third.Decision != app.AdmitSessionLimit {
		t.Errorf("same session, new address: %+v, %v, want the session limit", third, err)
	}
}

func TestAskAdmitAddressKeys(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, in, want string
	}{
		{"an IPv4 address", "198.51.100.7", "198.51.100.7"},
		{"with space around it", " 198.51.100.7 ", "198.51.100.7"},
		{"an IPv4-mapped IPv6 address is the IPv4 one", "::ffff:198.51.100.7", "198.51.100.7"},
		{"an IPv6 address is its /64", "2001:db8:1:2:3:4:5:6", "2001:db8:1:2::/64"},
		{"two addresses of one /64 are one", "2001:db8:1:2:ffff::1", "2001:db8:1:2::/64"},
		{"a zone is dropped", "fe80::1%eth0", "fe80::/64"},
		{"not an address", "198.51.100.7, 10.0.0.1", ""},
		{"words", "unknown", ""},
		{"nothing", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			qs := &fakeQuestions{}
			_, ask := newAdmitFixture(t, app.Guards{}, qs, modelrouter.QuestionLimits{PerHour: 5})
			if _, err := ask.Admit(t.Context(), "https://demo", askerToken, tt.in); err != nil {
				t.Fatal(err)
			}
			if got := qs.seen[0].Address; got != tt.want {
				t.Errorf("counted under %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAskAdmitWithAnUnknownAddressStillLimitsTheSession(t *testing.T) {
	t.Parallel()
	qs := &fakeQuestions{}
	_, ask := newAdmitFixture(t, app.Guards{}, qs, modelrouter.QuestionLimits{PerHour: 1})
	if got, err := ask.Admit(t.Context(), "https://demo", askerToken, ""); err != nil || got.Decision != app.AdmitAllowed {
		t.Fatalf("first: %+v, %v", got, err)
	}
	if got, err := ask.Admit(t.Context(), "https://demo", askerToken, ""); err != nil || got.Decision != app.AdmitSessionLimit {
		t.Fatalf("second: %+v, %v, want the session limit", got, err)
	}
}

func TestAskAdmitSaysTheBudgetIsSpentBeforeCountingAQuestion(t *testing.T) {
	t.Parallel()
	usage := &fakeUsage{byDay: map[string]kernel.Decimal{"2026-01-02": decimalOf(t, "5")}}
	qs := &fakeQuestions{}
	_, ask := newAdmitFixture(t, app.Guards{Budget: budgetOf(t, "5"), Usage: usage}, qs, modelrouter.QuestionLimits{PerHour: 10})
	got, err := ask.Admit(t.Context(), "https://demo", askerToken, "198.51.100.7")
	if err != nil {
		t.Fatal(err)
	}
	if want := (app.Admission{Decision: app.AdmitBudgetSpent, ResetsAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)}); got != want {
		t.Errorf("admission = %+v, want %+v", got, want)
	}
	if len(qs.seen) != 0 {
		t.Errorf("a question was counted while the budget is spent: %+v", qs.seen)
	}
}

func TestAskAdmitWithNoLimitsAllowsAndCountsNothing(t *testing.T) {
	t.Parallel()
	qs := &fakeQuestions{}
	_, ask := newAdmitFixture(t, app.Guards{}, qs, modelrouter.QuestionLimits{})
	for range 50 {
		if got, err := ask.Admit(t.Context(), "https://demo", askerToken, "198.51.100.7"); err != nil || got.Decision != app.AdmitAllowed {
			t.Fatalf("%+v, %v", got, err)
		}
	}
	if len(qs.seen) != 0 {
		t.Errorf("questions counted with no limit: %d", len(qs.seen))
	}
	// An Ask door that never had a store allows too.
	f := newRouterFixture(t)
	plain := app.NewAsk(fakeAskers{askerToken: {Kind: app.ActorOwner, ID: "s"}}, f.store, f.router)
	if got, err := plain.Admit(t.Context(), "https://demo", askerToken, "198.51.100.7"); err != nil || got.Decision != app.AdmitAllowed {
		t.Errorf("no store: %+v, %v", got, err)
	}
}

func TestAskAdmitRefusesAnInvalidSessionAndFailsClosedOnStoreErrors(t *testing.T) {
	t.Parallel()
	qs := &fakeQuestions{}
	_, ask := newAdmitFixture(t, app.Guards{}, qs, modelrouter.QuestionLimits{PerHour: 1})
	if _, err := ask.Admit(t.Context(), "https://demo", "nobody", "198.51.100.7"); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("error = %v, want ErrNoSession", err)
	}
	if len(qs.seen) != 0 {
		t.Error("a question of no session was counted")
	}
	qs.err = errBoom
	if _, err := ask.Admit(t.Context(), "https://demo", askerToken, "198.51.100.7"); !errors.Is(err, errBoom) {
		t.Errorf("a failing store: error = %v, want it to wrap the store's", err)
	}
	usage := &fakeUsage{spentErr: errBoom}
	_, ask = newAdmitFixture(t, app.Guards{Budget: budgetOf(t, "1"), Usage: usage}, &fakeQuestions{}, modelrouter.QuestionLimits{PerHour: 1})
	if _, err := ask.Admit(t.Context(), "https://demo", askerToken, ""); !errors.Is(err, app.ErrUsageUnavailable) {
		t.Errorf("an unreadable budget: error = %v, want ErrUsageUnavailable", err)
	}
}
