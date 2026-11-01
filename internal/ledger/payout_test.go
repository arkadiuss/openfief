package ledger_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestPayoutSettingPaidOut(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setting func(t *testing.T) ledger.PayoutSetting
		on      string
		wantOut bool
	}{
		{name: "the default keeps payouts as cash", setting: settingOf(ledger.PayoutCash, ""), on: "2026-10-10"},
		{name: "a first paid-out setting covers the whole history", setting: settingOf(ledger.PayoutPaidOut, ""), on: "2020-01-02", wantOut: true},
		{name: "paid out on the day it took effect", setting: changedSetting(ledger.PayoutPaidOut, "2026-10-10"), on: "2026-10-10", wantOut: true},
		{name: "paid out after it took effect", setting: changedSetting(ledger.PayoutPaidOut, "2026-10-10"), on: "2027-10-10", wantOut: true},
		{name: "what fell due before the change follows the first setting", setting: changedSetting(ledger.PayoutPaidOut, "2026-10-10"), on: "2026-10-09"},
		{name: "back to cash keeps later payouts as cash", setting: changedSetting(ledger.PayoutCash, "2026-10-10"), on: "2027-10-10"},
		{name: "back to cash keeps earlier payouts paid out", setting: changedSetting(ledger.PayoutCash, "2026-10-10"), on: "2026-05-04", wantOut: true},
		{name: "a change from an earlier day covers what fell due since", setting: changedSetting(ledger.PayoutPaidOut, "2025-01-02"), on: "2025-06-01", wantOut: true},
		{name: "a setting without its earlier ones covers the days before it", setting: settingOf(ledger.PayoutPaidOut, "2026-10-10"), on: "2026-05-04", wantOut: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.setting(t).PaidOut(date(t, tt.on)); got != tt.wantOut {
				t.Errorf("PaidOut(%s) = %v, want %v", tt.on, got, tt.wantOut)
			}
		})
	}
}

// settingOf returns a setting of p in force since the day since, or for
// the whole history when since is empty.
func settingOf(p ledger.Payout, since string) func(t *testing.T) ledger.PayoutSetting {
	return func(t *testing.T) ledger.PayoutSetting {
		t.Helper()
		s := ledger.PayoutSetting{Payout: p}
		if since != "" {
			s.Since = date(t, since)
		}
		return s
	}
}

// changedSetting returns the setting of an account first set to the other
// payout than p, changed to p from the day from.
func changedSetting(p ledger.Payout, from string) func(t *testing.T) ledger.PayoutSetting {
	return func(t *testing.T) ledger.PayoutSetting {
		t.Helper()
		first := ledger.PayoutPaidOut
		if p == ledger.PayoutPaidOut {
			first = ledger.PayoutCash
		}
		return ledger.PayoutSetting{Payout: first}.Changed(p, date(t, from))
	}
}

func TestPayoutSettingChanged(t *testing.T) {
	t.Parallel()
	first, later, earlier := date(t, "2026-03-02"), date(t, "2026-09-01"), date(t, "2025-11-03")
	paid := ledger.PayoutSetting{}.Changed(ledger.PayoutPaidOut, first)
	cashFirst := ledger.PayoutPeriod{Payout: ledger.PayoutCash}
	tests := []struct {
		name string
		got  ledger.PayoutSetting
		want ledger.PayoutSetting
	}{
		{
			name: "a change takes effect from the day given",
			got:  paid,
			want: ledger.PayoutSetting{Payout: ledger.PayoutPaidOut, Since: first, Earlier: []ledger.PayoutPeriod{cashFirst}},
		},
		{name: "the same setting keeps its day", got: paid.Changed(ledger.PayoutPaidOut, later), want: paid},
		{
			name: "the same setting from an earlier day takes effect from then",
			got:  paid.Changed(ledger.PayoutPaidOut, earlier),
			want: ledger.PayoutSetting{Payout: ledger.PayoutPaidOut, Since: earlier, Earlier: []ledger.PayoutPeriod{cashFirst}},
		},
		{
			name: "back to cash keeps the paid-out days before it",
			got:  paid.Changed(ledger.PayoutCash, later),
			want: ledger.PayoutSetting{Payout: ledger.PayoutCash, Since: later, Earlier: []ledger.PayoutPeriod{cashFirst, {Payout: ledger.PayoutPaidOut, Since: first}}},
		},
		{name: "back to cash before the change replaces it", got: paid.Changed(ledger.PayoutCash, earlier), want: ledger.PayoutSetting{}},
		{name: "the first setting stays when nothing changes", got: ledger.PayoutSetting{}.Changed(ledger.PayoutCash, first), want: ledger.PayoutSetting{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.got); diff != "" {
				t.Errorf("Changed (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPayoutKeys(t *testing.T) {
	t.Parallel()
	tests := []struct {
		key     string
		want    ledger.Payout
		wantErr error
	}{
		{key: "cash", want: ledger.PayoutCash},
		{key: "paid-out", want: ledger.PayoutPaidOut},
		{key: "monthly", wantErr: ledger.ErrUnknownPayout},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.ParsePayout(tt.key)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParsePayout error (-want +got):\n%s", diff)
			}
			if err == nil && (got != tt.want || got.String() != tt.key) {
				t.Errorf("ParsePayout(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
	if got := ledger.Payout(9).String(); got != "Payout(9)" {
		t.Errorf("String of an unknown payout = %q", got)
	}
}

func TestPayoutSettingValidate(t *testing.T) {
	t.Parallel()
	first, later := date(t, "2026-03-02"), date(t, "2026-09-01")
	tests := []struct {
		name    string
		setting ledger.PayoutSetting
		wantErr error
	}{
		{name: "a first paid-out setting needs no day", setting: ledger.PayoutSetting{Payout: ledger.PayoutPaidOut}},
		{name: "a changed setting", setting: ledger.PayoutSetting{}.Changed(ledger.PayoutPaidOut, first).Changed(ledger.PayoutCash, later)},
		{name: "an unknown payout", setting: ledger.PayoutSetting{Payout: ledger.Payout(9)}, wantErr: ledger.ErrUnknownPayout},
		{
			name:    "an unknown earlier payout",
			setting: ledger.PayoutSetting{Since: first, Earlier: []ledger.PayoutPeriod{{Payout: ledger.Payout(9)}}},
			wantErr: ledger.ErrUnknownPayout,
		},
		{
			name:    "a later setting without a day",
			setting: ledger.PayoutSetting{Payout: ledger.PayoutPaidOut, Earlier: []ledger.PayoutPeriod{{}}},
			wantErr: ledger.ErrUnknownPayout,
		},
		{
			name:    "settings out of order",
			setting: ledger.PayoutSetting{Since: first, Earlier: []ledger.PayoutPeriod{{}, {Payout: ledger.PayoutPaidOut, Since: later}}},
			wantErr: ledger.ErrUnknownPayout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.wantErr, tt.setting.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate (-want +got):\n%s", diff)
			}
		})
	}
}
