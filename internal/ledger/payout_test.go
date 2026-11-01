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
		payout  ledger.Payout
		since   string
		on      string
		wantOut bool
	}{
		{name: "the default keeps payouts as cash", payout: ledger.PayoutCash, on: "2026-10-10"},
		{name: "paid out on the day it took effect", payout: ledger.PayoutPaidOut, since: "2026-10-10", on: "2026-10-10", wantOut: true},
		{name: "paid out after it took effect", payout: ledger.PayoutPaidOut, since: "2026-10-10", on: "2027-10-10", wantOut: true},
		{name: "what fell due before the change stays as cash", payout: ledger.PayoutPaidOut, since: "2026-10-10", on: "2026-10-09"},
		{name: "back to cash keeps later payouts as cash", payout: ledger.PayoutCash, since: "2026-10-10", on: "2027-10-10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := ledger.PayoutSetting{Payout: tt.payout}
			if tt.since != "" {
				s.Since = date(t, tt.since)
			}
			if got := s.PaidOut(date(t, tt.on)); got != tt.wantOut {
				t.Errorf("PaidOut(%s) = %v, want %v", tt.on, got, tt.wantOut)
			}
		})
	}
}

func TestPayoutSettingChanged(t *testing.T) {
	t.Parallel()
	first, later := date(t, "2026-03-02"), date(t, "2026-09-01")
	paid := ledger.PayoutSetting{}.Changed(ledger.PayoutPaidOut, first)
	tests := []struct {
		name string
		got  ledger.PayoutSetting
		want ledger.PayoutSetting
	}{
		{name: "a change takes effect today", got: paid, want: ledger.PayoutSetting{Payout: ledger.PayoutPaidOut, Since: first}},
		{name: "the same setting keeps its day", got: paid.Changed(ledger.PayoutPaidOut, later), want: paid},
		{name: "back to cash takes effect today", got: paid.Changed(ledger.PayoutCash, later), want: ledger.PayoutSetting{Payout: ledger.PayoutCash, Since: later}},
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
	if err := (ledger.PayoutSetting{Payout: ledger.Payout(9)}).Validate(); err == nil {
		t.Error("Validate accepted an unknown payout")
	}
}
