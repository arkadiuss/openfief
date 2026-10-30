package connectapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// transferBody is a Transfer message's JSON.
type transferBody struct {
	WithdrawalAccountID  string `json:"withdrawalAccountId"`
	WithdrawalDate       string `json:"withdrawalDate"`
	DepositAccountID     string `json:"depositAccountId"`
	DepositDate          string `json:"depositDate"`
	Amount               amount `json:"amount"`
	WithdrawalSource     string `json:"withdrawalSource"`
	WithdrawalExternalID string `json:"withdrawalExternalId"`
	DepositSource        string `json:"depositSource"`
	DepositExternalID    string `json:"depositExternalId"`
}

// transfersBody is a ListTransfersResponse message's JSON.
type transfersBody struct {
	Suggested []transferBody `json:"suggested"`
	Confirmed []transferBody `json:"confirmed"`
	Rejected  []transferBody `json:"rejected"`
}

func TestRESTTransfers(t *testing.T) {
	t.Parallel()
	l := newLedger(t)
	leg := func(acc ledger.AccountID, id, day string, deposit bool) ledger.Transaction {
		h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: id}, Account: acc, Date: date(t, day)}
		build := ledger.NewWithdrawal
		if deposit {
			build = ledger.NewDeposit
		}
		tx, err := build(h, money(t, "250", "PLN"))
		if err != nil {
			t.Fatalf("building %s: %v", id, err)
		}
		return tx
	}
	// The fake answers every account with all of txs, and lists acc-1 only,
	// so legs in acc-2 are read once.
	l.txs = append(l.txs,
		leg("acc-1", "w1", "2026-03-02", false), leg("acc-2", "d1", "2026-03-03", true),
		leg("acc-1", "w2", "2026-03-05", false), leg("acc-2", "d2", "2026-03-06", true))
	l.links = []ledger.TransferLink{{
		Pair: ledger.TransferPair{
			Withdrawal: ledger.TransactionKey{Source: "test", ExternalID: "w1"},
			Deposit:    ledger.TransactionKey{Source: "test", ExternalID: "d1"},
		},
		Decision: ledger.TransferConfirmed, Version: 1,
	}}
	req, rec := newAPI(t, l).serve(t, http.MethodGet, "/v1/transfers", goodToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/transfers = %d\n%s", rec.Code, rec.Body)
	}
	loadContract(t).check(t, req, rec)
	var got transfersBody
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decoding %s: %v", rec.Body, err)
	}
	pair := func(w, wDay, d, dDay string) transferBody {
		return transferBody{
			WithdrawalAccountID: "acc-1", WithdrawalDate: wDay, DepositAccountID: "acc-2", DepositDate: dDay,
			Amount: amount{Amount: "250", Currency: "PLN"}, WithdrawalSource: "test", WithdrawalExternalID: w,
			DepositSource: "test", DepositExternalID: d,
		}
	}
	want := transfersBody{
		Suggested: []transferBody{pair("w2", "2026-03-05", "d2", "2026-03-06")},
		Confirmed: []transferBody{pair("w1", "2026-03-02", "d1", "2026-03-03")},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GET /v1/transfers mismatch (-want +got):\n%s", diff)
	}
}
