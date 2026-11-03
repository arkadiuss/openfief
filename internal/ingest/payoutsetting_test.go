package ingest_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
)

func TestPayoutSettingPaidOut(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setting func(t *testing.T) ingest.PayoutSetting
		on      string
		wantOut bool
	}{
		{name: "the default keeps payouts as cash", setting: settingOf(ingest.PayoutCash, ""), on: "2026-10-10"},
		{name: "a first paid-out setting covers the whole history", setting: settingOf(ingest.PayoutPaidOut, ""), on: "2020-01-02", wantOut: true},
		{name: "paid out on the day it took effect", setting: changedSetting(ingest.PayoutPaidOut, "2026-10-10"), on: "2026-10-10", wantOut: true},
		{name: "paid out after it took effect", setting: changedSetting(ingest.PayoutPaidOut, "2026-10-10"), on: "2027-10-10", wantOut: true},
		{name: "what fell due before the change follows the first setting", setting: changedSetting(ingest.PayoutPaidOut, "2026-10-10"), on: "2026-10-09"},
		{name: "back to cash keeps later payouts as cash", setting: changedSetting(ingest.PayoutCash, "2026-10-10"), on: "2027-10-10"},
		{name: "back to cash keeps earlier payouts paid out", setting: changedSetting(ingest.PayoutCash, "2026-10-10"), on: "2026-05-04", wantOut: true},
		{name: "a change from an earlier day covers what fell due since", setting: changedSetting(ingest.PayoutPaidOut, "2025-01-02"), on: "2025-06-01", wantOut: true},
		{name: "a setting without its earlier ones covers the days before it", setting: settingOf(ingest.PayoutPaidOut, "2026-10-10"), on: "2026-05-04", wantOut: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.setting(t).PaidOut(goLiveDate(t, tt.on)); got != tt.wantOut {
				t.Errorf("PaidOut(%s) = %v, want %v", tt.on, got, tt.wantOut)
			}
		})
	}
}

// settingOf returns a setting of p in force since the day since, or for
// the whole history when since is empty.
func settingOf(p ingest.Payout, since string) func(t *testing.T) ingest.PayoutSetting {
	return func(t *testing.T) ingest.PayoutSetting {
		t.Helper()
		s := ingest.PayoutSetting{Payout: p}
		if since != "" {
			s.Since = goLiveDate(t, since)
		}
		return s
	}
}

// changedSetting returns the setting of an account first set to the other
// payout than p, changed to p from the day from.
func changedSetting(p ingest.Payout, from string) func(t *testing.T) ingest.PayoutSetting {
	return func(t *testing.T) ingest.PayoutSetting {
		t.Helper()
		first := ingest.PayoutPaidOut
		if p == ingest.PayoutPaidOut {
			first = ingest.PayoutCash
		}
		return ingest.PayoutSetting{Payout: first}.Changed(p, goLiveDate(t, from))
	}
}

func TestPayoutSettingChanged(t *testing.T) {
	t.Parallel()
	first, later, earlier := goLiveDate(t, "2026-03-02"), goLiveDate(t, "2026-09-01"), goLiveDate(t, "2025-11-03")
	paid := ingest.PayoutSetting{}.Changed(ingest.PayoutPaidOut, first)
	cashFirst := ingest.PayoutPeriod{Payout: ingest.PayoutCash}
	tests := []struct {
		name string
		got  ingest.PayoutSetting
		want ingest.PayoutSetting
	}{
		{
			name: "a change takes effect from the day given",
			got:  paid,
			want: ingest.PayoutSetting{Payout: ingest.PayoutPaidOut, Since: first, Earlier: []ingest.PayoutPeriod{cashFirst}},
		},
		{name: "the same setting keeps its day", got: paid.Changed(ingest.PayoutPaidOut, later), want: paid},
		{
			name: "the same setting from an earlier day takes effect from then",
			got:  paid.Changed(ingest.PayoutPaidOut, earlier),
			want: ingest.PayoutSetting{Payout: ingest.PayoutPaidOut, Since: earlier, Earlier: []ingest.PayoutPeriod{cashFirst}},
		},
		{
			name: "back to cash keeps the paid-out days before it",
			got:  paid.Changed(ingest.PayoutCash, later),
			want: ingest.PayoutSetting{Payout: ingest.PayoutCash, Since: later, Earlier: []ingest.PayoutPeriod{cashFirst, {Payout: ingest.PayoutPaidOut, Since: first}}},
		},
		{name: "back to cash before the change replaces it", got: paid.Changed(ingest.PayoutCash, earlier), want: ingest.PayoutSetting{}},
		{name: "the first setting stays when nothing changes", got: ingest.PayoutSetting{}.Changed(ingest.PayoutCash, first), want: ingest.PayoutSetting{}},
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
		want    ingest.Payout
		wantErr error
	}{
		{key: "cash", want: ingest.PayoutCash},
		{key: "paid-out", want: ingest.PayoutPaidOut},
		{key: "monthly", wantErr: ingest.ErrUnknownPayout},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			got, err := ingest.ParsePayout(tt.key)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParsePayout error (-want +got):\n%s", diff)
			}
			if err == nil && (got != tt.want || got.String() != tt.key) {
				t.Errorf("ParsePayout(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
	if got := ingest.Payout(9).String(); got != "Payout(9)" {
		t.Errorf("String of an unknown payout = %q", got)
	}
}

func TestPayoutSettingValidate(t *testing.T) {
	t.Parallel()
	first, later := goLiveDate(t, "2026-03-02"), goLiveDate(t, "2026-09-01")
	tests := []struct {
		name    string
		setting ingest.PayoutSetting
		wantErr error
	}{
		{name: "a first paid-out setting needs no day", setting: ingest.PayoutSetting{Payout: ingest.PayoutPaidOut}},
		{name: "a changed setting", setting: ingest.PayoutSetting{}.Changed(ingest.PayoutPaidOut, first).Changed(ingest.PayoutCash, later)},
		{name: "an unknown payout", setting: ingest.PayoutSetting{Payout: ingest.Payout(9)}, wantErr: ingest.ErrUnknownPayout},
		{
			name:    "an unknown earlier payout",
			setting: ingest.PayoutSetting{Since: first, Earlier: []ingest.PayoutPeriod{{Payout: ingest.Payout(9)}}},
			wantErr: ingest.ErrUnknownPayout,
		},
		{
			name:    "a later setting without a day",
			setting: ingest.PayoutSetting{Payout: ingest.PayoutPaidOut, Earlier: []ingest.PayoutPeriod{{}}},
			wantErr: ingest.ErrUnknownPayout,
		},
		{
			name:    "settings out of order",
			setting: ingest.PayoutSetting{Since: first, Earlier: []ingest.PayoutPeriod{{}, {Payout: ingest.PayoutPaidOut, Since: later}}},
			wantErr: ingest.ErrUnknownPayout,
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
