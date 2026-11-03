package ingest_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestCheckPayoutFrom(t *testing.T) {
	t.Parallel()
	today := goLiveDate(t, "2026-03-02")
	tests := []struct {
		name    string
		from    string
		wantErr error
	}{
		{name: "today", from: "2026-03-02"},
		{name: "an earlier day", from: "2024-01-15"},
		{name: "a later day", from: "2026-03-03", wantErr: ingest.ErrPayoutFromFuture},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ingest.CheckPayoutFrom(goLiveDate(t, tt.from), today)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("CheckPayoutFrom (-want +got):\n%s", diff)
			}
		})
	}
}

// payoutTx returns a synthetic bond account's transaction of kind on day,
// keyed id.
func payoutTx(t *testing.T, kind ledger.TransactionKind, id, day string) ledger.Transaction {
	t.Helper()
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	cash := kernel.NewMoney(kernel.NewDecimal(5305, -2), pln)
	h := ledger.TransactionHeader{
		Key: ledger.TransactionKey{Source: "bonds-main:treasury-bonds:acc-bonds", ExternalID: id}, Account: "acc-bonds", Date: goLiveDate(t, day),
	}
	newTx := ledger.NewDeposit
	if kind == ledger.KindWithdrawal {
		newTx = ledger.NewWithdrawal
	}
	tx, err := newTx(h, cash)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

// TestStalePayouts supersedes the payout withdrawals a payout change from
// a day no longer makes (D-162).
func TestStalePayouts(t *testing.T) {
	t.Parallel()
	from := goLiveDate(t, "2026-01-01")
	paid := ingest.PayoutSetting{Payout: ingest.PayoutPaidOut}
	backToCash := paid.Changed(ingest.PayoutCash, from)
	current := []ledger.Transaction{
		payoutTx(t, ledger.KindWithdrawal, "lot:COI1028:2024-10-10:coupon:2025-10-10:payout", "2025-10-10"),
		payoutTx(t, ledger.KindWithdrawal, "lot:COI1028:2024-10-10:coupon:2026-10-10:payout", "2026-10-10"),
		payoutTx(t, ledger.KindWithdrawal, "lot:COI1028:2024-10-10:maturity:2026-01-01:payout", "2026-01-01"),
		payoutTx(t, ledger.KindWithdrawal, "withdrawal-1", "2026-02-02"),
		payoutTx(t, ledger.KindDeposit, "deposit-1:payout", "2026-02-02"),
	}
	tests := []struct {
		name    string
		setting ingest.PayoutSetting
		want    []string
	}{
		{
			name:    "back to cash supersedes the withdrawals from that day on",
			setting: backToCash,
			want: []string{
				"bonds-main:treasury-bonds:acc-bonds/lot:COI1028:2024-10-10:coupon:2026-10-10:payout 2026-10-10",
				"bonds-main:treasury-bonds:acc-bonds/lot:COI1028:2024-10-10:maturity:2026-01-01:payout 2026-01-01",
			},
		},
		{name: "paid out supersedes none", setting: paid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ingest.StalePayouts(current, tt.setting, from)
			if err != nil {
				t.Fatalf("StalePayouts: %v", err)
			}
			var keys []string
			for i := range got {
				if got[i].Kind != ledger.KindDeletion || got[i].Account != "acc-bonds" {
					t.Errorf("StalePayouts returned %s of %s, want a deletion in acc-bonds", got[i].Kind, got[i].Account)
				}
				keys = append(keys, got[i].Key.String()+" "+got[i].Date.String())
			}
			if diff := cmp.Diff(tt.want, keys); diff != "" {
				t.Errorf("StalePayouts (-want +got):\n%s", diff)
			}
		})
	}
}
