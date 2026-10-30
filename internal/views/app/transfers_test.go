package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// transferKey returns the test source's key id.
func transferKey(id string) ledger.TransactionKey {
	return ledger.TransactionKey{Source: "test", ExternalID: id}
}

// newTransferFixture is the base fixture whose acc-1 withdraws 200 PLN on
// 03-01 (w1) and 300 PLN on 03-02 (w2), and whose acc-2 gets deposits of
// 200 PLN on 03-02 (d1) and 300 PLN on 03-03 (d2), with links.
func newTransferFixture(t *testing.T, links ...ledger.TransferLink) *fixture {
	t.Helper()
	f := newFixture(t)
	leg := func(acc ledger.AccountID, id, day, amount string, deposit bool) ledger.Transaction {
		h := ledger.TransactionHeader{Key: transferKey(id), Account: acc, Date: date(t, day)}
		build := ledger.NewWithdrawal
		if deposit {
			build = ledger.NewDeposit
		}
		tx, err := build(h, money(t, amount, "PLN"))
		if err != nil {
			t.Fatalf("building %s: %v", id, err)
		}
		return tx
	}
	f.ledger.txs = map[ledger.AccountID][]ledger.Transaction{
		"acc-1": {leg("acc-1", "w1", "2026-03-01", "200", false), leg("acc-1", "w2", "2026-03-02", "300", false)},
		"acc-2": {leg("acc-2", "d1", "2026-03-02", "200", true), leg("acc-2", "d2", "2026-03-03", "300", true)},
	}
	f.ledger.links = links
	return f
}

// legKeys names a transfer by its legs' external IDs.
func legKeys(ts []views.Transfer) [][2]string {
	out := make([][2]string, 0, len(ts))
	for i := range ts {
		out = append(out, [2]string{ts[i].Withdrawal.Key.ExternalID, ts[i].Deposit.Key.ExternalID})
	}
	return out
}

func TestTransfers(t *testing.T) {
	t.Parallel()
	link := func(w, d string, decision ledger.TransferDecision) ledger.TransferLink {
		return ledger.TransferLink{Pair: ledger.TransferPair{Withdrawal: transferKey(w), Deposit: transferKey(d)}, Decision: decision, Version: 1}
	}
	tests := []struct {
		name                           string
		links                          []ledger.TransferLink
		suggested, confirmed, rejected [][2]string
	}{
		{"nothing decided: both suggested", nil, [][2]string{{"w1", "d1"}, {"w2", "d2"}}, nil, nil},
		{
			"one confirmed, one rejected",
			[]ledger.TransferLink{link("w1", "d1", ledger.TransferConfirmed), link("w2", "d2", ledger.TransferRejected)},
			nil,
			[][2]string{{"w1", "d1"}},
			[][2]string{{"w2", "d2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := newTransferFixture(t, tt.links...).views(t).Transfers(t.Context())
			if err != nil {
				t.Fatalf("Transfers: %v", err)
			}
			for _, c := range []struct {
				name      string
				want, got [][2]string
			}{
				{"suggested", tt.suggested, legKeys(got.Suggested)},
				{"confirmed", tt.confirmed, legKeys(got.Confirmed)},
				{"rejected", tt.rejected, legKeys(got.Rejected)},
			} {
				if diff := cmp.Diff(c.want, c.got, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("%s (-want +got):\n%s", c.name, diff)
				}
			}
		})
	}
}

func TestTransfersFailWithTheLedger(t *testing.T) {
	t.Parallel()
	f := newTransferFixture(t)
	f.ledger.fail = errStore
	if _, err := f.views(t).Transfers(t.Context()); err == nil {
		t.Fatal("Transfers on a failing ledger succeeded")
	}
}
