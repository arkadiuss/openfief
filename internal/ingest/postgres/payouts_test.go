package postgres_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/postgres"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestPayoutSettingRoundTrip stores a hand-entered bond account's payout
// setting (D-135) with the setting in force before it (D-162) in the
// ingest schema, keyed by the ledger's account ID, replaces it, and
// forgets it with a deleted account (D-120).
func TestPayoutSettingRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s := postgres.New(newPool(t))
	const account ledger.AccountID = "acc-bonds"
	checkPayoutSetting(t, s, account, ingest.PayoutSetting{})
	since, err := kernel.ParseDate("2026-03-02")
	if err != nil {
		t.Fatal(err)
	}
	paidOut := ingest.PayoutSetting{Payout: ingest.PayoutPaidOut, Since: since, Earlier: []ingest.PayoutPeriod{{Payout: ingest.PayoutCash}}}
	backToCash := paidOut.Changed(ingest.PayoutCash, since.AddDays(7))
	for _, want := range []ingest.PayoutSetting{paidOut, backToCash, {Payout: ingest.PayoutPaidOut}} {
		if err := s.SavePayoutSetting(ctx, account, want); err != nil {
			t.Fatalf("SavePayoutSetting: %v", err)
		}
		checkPayoutSetting(t, s, account, want)
	}
	if err := s.ForgetAccount(ctx, account); err != nil {
		t.Fatalf("ForgetAccount: %v", err)
	}
	checkPayoutSetting(t, s, account, ingest.PayoutSetting{})
}

// checkPayoutSetting fails t unless account's stored payout setting is
// want; the zero setting is the default, as cash.
func checkPayoutSetting(t *testing.T, s *postgres.Store, account ledger.AccountID, want ingest.PayoutSetting) {
	t.Helper()
	got, err := s.PayoutSetting(t.Context(), account)
	if err != nil {
		t.Fatalf("PayoutSetting: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("payout setting of %s (-want +got):\n%s", account, diff)
	}
}
