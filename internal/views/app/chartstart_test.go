package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/app"
)

var errStartsDown = errors.New("settings unavailable")

// fakeStarts holds the charts' starts the owner stored (D-167), by
// account, the portfolio's under the empty ID; it records the accounts it
// was asked about, and fails every read with fail.
type fakeStarts struct {
	stored map[ledger.AccountID]kernel.Date
	asked  []ledger.AccountID
	fail   error
}

func (s *fakeStarts) ChartStart(_ context.Context, account ledger.AccountID) (kernel.Date, error) {
	s.asked = append(s.asked, account)
	return s.stored[account], s.fail
}

// withStarts returns f's views reading the stored charts' starts from s,
// asking a whether the caller may view each account.
func withStarts(t *testing.T, f *fixture, s *fakeStarts, a *denyingAuthorizer) *app.Views {
	t.Helper()
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), app.WithChartStarts(s), app.WithAuthorizer(a))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	return v
}

// markOpened marks the fixture's acc-2 as opened on its seeding day
// (D-168).
func markOpened(f *fixture) {
	for i := range f.ledger.accounts {
		if f.ledger.accounts[i].ID == "acc-2" {
			f.ledger.accounts[i].OpenedOnSeedingDay = true
		}
	}
}

// TestPerformanceChartStart checks where the portfolio's and an account's
// time charts start (D-138) with the owner's stored setting (D-167) and
// the mark of an account opened on its seeding day (D-168): a stored start
// is used when the query names none, a named one previews over it, each
// account keeps its own, and only a seeded account not so marked moves the
// portfolio's default start.
func TestPerformanceChartStart(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		account ledger.AccountID
		stored  map[ledger.AccountID]string
		// seedACC2 seeds acc-2 from opening cash on 03-02, and opened marks
		// it as opened on that day.
		seedACC2, opened bool
		from             string
		wantFrom, start  string
		chosen           string
	}{
		{name: "no stored start: the default", wantFrom: "2026-03-01", start: "2026-03-01"},
		{
			name: "the portfolio's stored start when the query names none", stored: map[ledger.AccountID]string{"": "2026-03-02"},
			wantFrom: "2026-03-02", start: "2026-03-01", chosen: "2026-03-02",
		},
		{
			name: "a start the query names previews over the stored one", stored: map[ledger.AccountID]string{"": "2026-03-02"},
			from: "2026-03-03", wantFrom: "2026-03-03", start: "2026-03-01", chosen: "2026-03-02",
		},
		{
			name: "a stored start before the first day starts on it", stored: map[ledger.AccountID]string{"": "2025-12-31"},
			wantFrom: "2026-03-01", start: "2026-03-01", chosen: "2025-12-31",
		},
		{
			name: "an unmarked account seeded from a starting balance moves the default start", seedACC2: true,
			wantFrom: "2026-03-02", start: "2026-03-02",
		},
		{
			name: "an account marked as opened on its seeding day doesn't move it", seedACC2: true, opened: true,
			wantFrom: "2026-03-01", start: "2026-03-01",
		},
		{
			name: "an account's own stored start, not the portfolio's", account: "acc-1",
			stored:   map[ledger.AccountID]string{"": "2026-03-03", "acc-1": "2026-03-02"},
			wantFrom: "2026-03-02", start: "2026-03-01", chosen: "2026-03-02",
		},
		{
			name: "an account without one starts on its first day", account: "acc-1", stored: map[ledger.AccountID]string{"": "2026-03-03"},
			wantFrom: "2026-03-01", start: "2026-03-01",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := seededFixture(t, tt.seedACC2, tt.opened)
			starts := storedStarts(t, tt.stored)
			got, err := performanceOf(t, withStarts(t, f, starts, &denyingAuthorizer{}), tt.account, tt.from)
			if err != nil {
				t.Fatalf("performance: %v", err)
			}
			want := []kernel.Date{date(t, tt.wantFrom), date(t, tt.start), optionalDate(t, tt.chosen)}
			if diff := cmp.Diff(want, []kernel.Date{got.From, got.Start, got.Chosen}); diff != "" {
				t.Errorf("from, default start, stored start (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff([]ledger.AccountID{tt.account}, starts.asked); diff != "" {
				t.Errorf("stored starts read (-want +got):\n%s", diff)
			}
		})
	}
}

// TestChartStartReadAccess checks that an account's stored charts' start is
// read only for a caller who may view the account (D-163), and that a
// stored start that can't be read fails the performance.
func TestChartStartReadAccess(t *testing.T) {
	t.Parallel()
	f := newPerformanceFixture(t)
	starts := &fakeStarts{stored: map[ledger.AccountID]kernel.Date{"acc-1": date(t, "2026-03-02")}}
	v := withStarts(t, f, starts, &denyingAuthorizer{denied: "acc-1"})
	_, err := v.AccountPerformance(reader(t), "acc-1", kernel.Date{}, kernel.Date{})
	if !cmp.Equal(app.ErrUnknownAccount, err, cmpopts.EquateErrors()) {
		t.Errorf("an account the caller may not view: error = %v, want ErrUnknownAccount", err)
	}
	if len(starts.asked) != 0 {
		t.Errorf("stored starts read = %v, want none", starts.asked)
	}

	failing := &fakeStarts{fail: errStartsDown}
	v = withStarts(t, f, failing, &denyingAuthorizer{})
	if _, err := v.Performance(reader(t), kernel.Date{}, kernel.Date{}); !cmp.Equal(errStartsDown, err, cmpopts.EquateErrors()) {
		t.Errorf("Performance with the settings down: error = %v, want errStartsDown", err)
	}
	if _, err := v.AccountPerformance(reader(t), "acc-1", kernel.Date{}, kernel.Date{}); !cmp.Equal(errStartsDown, err, cmpopts.EquateErrors()) {
		t.Errorf("AccountPerformance with the settings down: error = %v, want errStartsDown", err)
	}
}

// TestModelPortfolioStoredStart checks that the model portfolio's line
// starts on the portfolio's stored charts' start when the query names
// none, as the time-weighted return does (D-53, D-167).
func TestModelPortfolioStoredStart(t *testing.T) {
	t.Parallel()
	f := newPerformanceFixture(t)
	starts := &fakeStarts{stored: map[ledger.AccountID]kernel.Date{"": date(t, "2026-03-02")}}
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"),
		app.WithModel(newFakeModels(t, true)), app.WithChartStarts(starts))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	got, err := v.ModelPortfolio(reader(t), kernel.Date{}, kernel.Date{})
	if err != nil {
		t.Fatalf("ModelPortfolio: %v", err)
	}
	if diff := cmp.Diff(points(t, []string{"2026-03-02 100", "2026-03-03 122.222222222222"}), got.Line); diff != "" {
		t.Errorf("line (-want +got):\n%s", diff)
	}
	starts.fail = errStartsDown
	if _, err := v.ModelPortfolio(reader(t), kernel.Date{}, kernel.Date{}); !cmp.Equal(errStartsDown, err, cmpopts.EquateErrors()) {
		t.Errorf("with the settings down: error = %v, want errStartsDown", err)
	}
}

// storedStarts is fakeStarts holding days by account.
func storedStarts(t *testing.T, days map[ledger.AccountID]string) *fakeStarts {
	t.Helper()
	s := &fakeStarts{stored: map[ledger.AccountID]kernel.Date{}}
	for acc, d := range days {
		s.stored[acc] = date(t, d)
	}
	return s
}

// performanceOf returns v's performance of account, the portfolio's when
// it is empty, from the day from names, the default when empty.
func performanceOf(t *testing.T, v *app.Views, account ledger.AccountID, from string) (app.Performance, error) {
	t.Helper()
	var (
		p   app.Performance
		err error
	)
	if account == "" {
		p, err = v.Performance(reader(t), optionalDate(t, from), kernel.Date{})
	} else {
		p, err = v.AccountPerformance(reader(t), account, optionalDate(t, from), kernel.Date{})
	}
	if err != nil {
		return app.Performance{}, fmt.Errorf("performance of %q: %w", account, err)
	}
	return p, nil
}

// seededFixture is the performance fixture with acc-2 seeded from opening
// cash on 03-02 when seed says, and marked as opened on its seeding day
// when opened says (D-168).
func seededFixture(t *testing.T, seed, opened bool) *fixture {
	t.Helper()
	f := newPerformanceFixture(t)
	if seed {
		f.ledger.txs["acc-2"] = []ledger.Transaction{openingUSD(t)}
	}
	if opened {
		markOpened(f)
	}
	return f
}
