package postgres_test

import (
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
	"github.com/arkadiuss/openfief/internal/modelrouter/postgres"
)

func guardsDay(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func guardsDecimal(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func guardsCurrency(t *testing.T, code string) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestUsageSumsTheDaysCostsPerCurrencyAndSurvivesANewStore(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	store := postgres.New(pool)
	ctx := t.Context()
	usd, eur := guardsCurrency(t, "USD"), guardsCurrency(t, "EUR")
	rec := func(day, model, cost string, c kernel.Currency) {
		t.Helper()
		err := store.Record(ctx, &app.UsageRecord{
			Day: guardsDay(t, day), Model: refs(t, model)[0], Usage: modelrouter.Usage{InputTokens: 1000, OutputTokens: 100},
			Cost: guardsDecimal(t, cost), Currency: c,
		})
		if err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
	rec("2026-01-02", "openrouter/google/gemma", "0.000345", usd)
	rec("2026-01-02", "openrouter/google/gemma", "0.000345", usd)
	rec("2026-01-02", "openrouter/other", "0.1", usd)
	rec("2026-01-02", "openrouter/other", "7", eur)
	rec("2026-01-03", "openrouter/other", "1", usd)
	// A new store over the same database, as after a restart, sees the same totals.
	again := postgres.New(pool)
	tests := []struct {
		name string
		day  string
		cur  kernel.Currency
		want string
	}{
		{"the day's calls of every model add up", "2026-01-02", usd, "0.10069"},
		{"another currency is counted apart", "2026-01-02", eur, "7"},
		{"the next day starts over", "2026-01-03", usd, "1"},
		{"a day with no call", "2026-01-04", usd, "0"},
	}
	for _, tt := range tests {
		got, err := again.Spent(ctx, guardsDay(t, tt.day), tt.cur)
		if err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		if !got.Equal(guardsDecimal(t, tt.want)) {
			t.Errorf("%s: Spent = %s, want %s", tt.name, got, tt.want)
		}
	}
	var calls, input, output int64
	err := pool.QueryRow(ctx, `SELECT calls, input_tokens, output_tokens FROM modelrouter.daily_usage
		WHERE day = '2026-01-02' AND model = 'google/gemma'`).Scan(&calls, &input, &output)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || input != 2000 || output != 200 {
		t.Errorf("gemma's row = %d calls, %d in, %d out, want 2, 2000, 200", calls, input, output)
	}
}

func TestUsageRecordsCallsAtTheSameMomentWithoutLosingAny(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	usd := guardsCurrency(t, "USD")
	var wg sync.WaitGroup
	for range 25 {
		wg.Go(func() {
			err := store.Record(ctx, &app.UsageRecord{
				Day: guardsDay(t, "2026-01-02"), Model: refs(t, "openrouter/m")[0], Cost: guardsDecimal(t, "0.01"), Currency: usd,
			})
			if err != nil {
				t.Errorf("Record: %v", err)
			}
		})
	}
	wg.Wait()
	got, err := store.Spent(ctx, guardsDay(t, "2026-01-02"), usd)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(guardsDecimal(t, "0.25")) {
		t.Errorf("Spent = %s, want 0.25 from 25 calls of 0.01", got)
	}
}

func TestAdmitCountsPerSessionAndPerAddressAndSaysWhenItResets(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	limits := modelrouter.QuestionLimits{PerHour: 2, PerDay: 3}
	allowed := app.QuestionRuling{Allowed: true}
	// The steps run in order, each question after the ones before it.
	steps := []struct {
		name, session, address string
		after                  time.Duration
		want                   app.QuestionRuling
	}{
		{name: "the first question", session: "s1", address: "198.51.100.7", want: allowed},
		{name: "the second", session: "s1", address: "198.51.100.7", after: 10 * time.Minute, want: allowed},
		{
			name: "the third of one session in the hour", session: "s1", address: "198.51.100.7", after: 20 * time.Minute,
			want: app.QuestionRuling{Subject: app.SubjectSession, ResetsAt: now.Add(time.Hour)},
		},
		{
			name: "a second session at the address, which counted the first two", session: "s2", address: "198.51.100.7", after: 20 * time.Minute,
			want: app.QuestionRuling{Subject: app.SubjectAddress, ResetsAt: now.Add(time.Hour)},
		},
		{name: "another address, another session: free", session: "s3", address: "203.0.113.9", after: 20 * time.Minute, want: allowed},
		{name: "an unknown address leaves only the session's limit", session: "s4", after: 20 * time.Minute, want: allowed},
		{
			name: "a stopped question was not counted: after the hour it is the day's third", session: "s1", address: "198.51.100.7",
			after: time.Hour + time.Minute, want: allowed,
		},
		{
			name: "the day's limit of 3 resets when the oldest of the three leaves the day", session: "s1", address: "198.51.100.7",
			after: 2 * time.Hour, want: app.QuestionRuling{Subject: app.SubjectSession, ResetsAt: now.Add(24 * time.Hour)},
		},
	}
	for _, st := range steps {
		got, err := store.Admit(ctx, &app.QuestionAsked{At: now.Add(st.after), Session: st.session, Address: st.address, Limits: limits})
		if err != nil {
			t.Fatalf("%s: Admit: %v", st.name, err)
		}
		if diff := cmp.Diff(st.want, got); diff != "" {
			t.Errorf("%s (-want +got):\n%s", st.name, diff)
		}
	}
}

func TestAdmitLetsExactlyTheLimitThroughWhenQuestionsArriveAtOnce(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	limits := modelrouter.QuestionLimits{PerHour: 5}
	var mu sync.Mutex
	allowed := 0
	var wg sync.WaitGroup
	for range 30 {
		wg.Go(func() {
			r, err := store.Admit(ctx, &app.QuestionAsked{At: now, Session: "one-session", Address: "198.51.100.7", Limits: limits})
			if err != nil {
				t.Errorf("Admit: %v", err)
				return
			}
			if r.Allowed {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if allowed != 5 {
		t.Errorf("%d questions of 30 at once went through, want exactly the limit, 5", allowed)
	}
}

func TestAdmitKeepsOnlyASaltedHashOfTheAddressAndForgetsOldQuestions(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	store := postgres.New(pool)
	ctx := t.Context()
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	limits := modelrouter.QuestionLimits{PerDay: 100}
	for _, q := range []app.QuestionAsked{
		{At: now.Add(-30 * time.Hour), Session: "old", Address: "198.51.100.7", Limits: limits},
		{At: now, Session: "a", Address: "198.51.100.7", Limits: limits},
		{At: now, Session: "b", Address: "198.51.100.7", Limits: limits},
		{At: now, Session: "c", Address: "2001:db8:1:2::/64", Limits: limits},
		{At: now, Session: "d", Limits: limits},
	} {
		if _, err := store.Admit(ctx, &q); err != nil {
			t.Fatalf("Admit: %v", err)
		}
	}
	keys := addressKeysBySession(t, pool)
	// The question 30 hours old is forgotten as soon as a newer one arrives.
	if _, kept := keys["old"]; kept {
		t.Errorf("a question older than the longest window was kept: %v", keys)
	}
	wantAddressesKeyedAsHashes(t, keys)
}

// wantAddressesKeyedAsHashes checks the address keys of the sessions a, b (one
// address), c (another) and d (none): the same address keys alike, as 64 hex
// digits that are not the address, another address differently, and none not at
// all.
func wantAddressesKeyedAsHashes(t *testing.T, keys map[string]string) {
	t.Helper()
	if keys["a"] == "" || keys["a"] != keys["b"] {
		t.Errorf("one address must key alike: %q %q", keys["a"], keys["b"])
	}
	if k := keys["a"]; len(k) != 64 || k == "198.51.100.7" {
		t.Errorf("the address key = %q, want 64 hex digits that are not the address", k)
	}
	if keys["c"] == "" || keys["c"] == keys["a"] {
		t.Errorf("another address must key differently: %q", keys["c"])
	}
	if keys["d"] != "" {
		t.Errorf("an unknown address was keyed: %q", keys["d"])
	}
}

// TestTheMigrationMakesOneSaltHoweverOftenItRuns checks the salt of the address
// keys: one random value per database.
func TestTheMigrationMakesOneSaltHoweverOftenItRuns(t *testing.T) {
	t.Parallel()
	pool := newPool(t) // runs the migrations twice
	var salts int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM modelrouter.guard_salt").Scan(&salts); err != nil || salts != 1 {
		t.Errorf("salts = %d, %v, want the one the migration made", salts, err)
	}
}

// addressKeysBySession returns the stored address key of each session's
// question, "" for a question with none.
func addressKeysBySession(t *testing.T, pool *pgxpool.Pool) map[string]string {
	t.Helper()
	rows, err := pool.Query(t.Context(), "SELECT session_id, COALESCE(address_key, '') FROM modelrouter.ask_questions")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var session, key string
		if err := rows.Scan(&session, &key); err != nil {
			t.Fatal(err)
		}
		out[session] = key
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}
