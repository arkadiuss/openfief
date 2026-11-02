package connectapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
)

// fakeItems is app.AttentionItems over fixed items.
type fakeItems []attention.Item

func (f fakeItems) List(context.Context) ([]attention.Item, error) { return f, nil }

// attentionBody is a ListAttentionItemsResponse message's JSON.
type attentionBody struct {
	Items []struct {
		ID            string `json:"id"`
		Kind          string `json:"kind"`
		Subject       string `json:"subject"`
		State         string `json:"state"`
		Detail        string `json:"detail"`
		OwnerResolves bool   `json:"ownerResolves"`
		History       []struct {
			Seq    int    `json:"seq"`
			State  string `json:"state"`
			Reason string `json:"reason"`
			At     string `json:"at"`
		} `json:"history"`
	} `json:"items"`
	Open int `json:"open"`
}

func TestRESTAttentionItems(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	items := fakeItems{
		{ID: 1, Kind: attention.KindImportProblems, Subject: "b-7", Changes: []attention.Change{
			{Seq: 1, State: attention.StateOpen, Reason: attention.ReasonRaised, Detail: "flex-main: 2 rows left out", At: at},
		}},
		{ID: 2, Kind: attention.KindSettingMissing, Subject: "marginal-rate", Changes: []attention.Change{
			{Seq: 1, State: attention.StateOpen, Reason: attention.ReasonRaised, At: at},
			{Seq: 2, State: attention.StateResolved, Reason: attention.ReasonCleared, At: at.Add(time.Hour)},
		}},
	}
	var log bytes.Buffer
	h, err := connectapi.New(newViews(t, newLedger(t), app.WithJurisdiction(pl.NewPack(), fakeTerms{}), app.WithAttention(items)),
		fakeVerifier{}, connectapi.WithLogger(slog.New(slog.NewJSONHandler(&log, nil))))
	if err != nil {
		t.Fatalf("connectapi.New: %v", err)
	}
	req, rec := api{handler: h, log: &log}.serve(t, http.MethodGet, "/v1/attention-items", goodToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/attention-items = %d\n%s", rec.Code, rec.Body)
	}
	loadContract(t).check(t, req, rec)
	checkAttentionBody(t, rec.Body.Bytes())
	req, rec = newAPI(t, newLedger(t)).serve(t, http.MethodGet, "/v1/attention-items", goodToken)
	loadContract(t).check(t, req, rec)
	if body := rec.Body.String(); rec.Code != http.StatusOK || body != "{}" {
		t.Errorf("without attention items: %d %s, want 200 and none", rec.Code, body)
	}
}

// checkAttentionBody fails t unless body lists the open import problems
// and the cleared missing setting.
func checkAttentionBody(t *testing.T, body []byte) {
	t.Helper()
	var got attentionBody
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decoding %s: %v", body, err)
	}
	summary := make([]string, 0, len(got.Items))
	for _, it := range got.Items {
		last := it.History[len(it.History)-1]
		summary = append(summary, fmt.Sprintf("%s %s %s %q resolves=%t history=%d last=%s at %s",
			it.ID, it.Kind, it.State, it.Detail, it.OwnerResolves, len(it.History), last.Reason, last.At))
	}
	want := []string{
		`1 import-problems ATTENTION_ITEM_STATE_OPEN "flex-main: 2 rows left out" resolves=true history=1 last=ATTENTION_ITEM_REASON_RAISED at 2026-03-02T10:00:00Z`,
		`2 setting-missing ATTENTION_ITEM_STATE_RESOLVED "" resolves=false history=2 last=ATTENTION_ITEM_REASON_CLEARED at 2026-03-02T11:00:00Z`,
	}
	if diff := cmp.Diff(want, summary); diff != "" || got.Open != 1 {
		t.Errorf("items (-want +got):\n%s\nopen: %d, want 1", diff, got.Open)
	}
}

func TestRESTBrokenTransfers(t *testing.T) {
	t.Parallel()
	l := newLedger(t)
	key := func(id string) ledger.TransactionKey { return ledger.TransactionKey{Source: "test", ExternalID: id} }
	w, err := ledger.NewWithdrawal(ledger.TransactionHeader{Key: key("w1"), Account: "acc-1", Date: date(t, "2026-03-02")}, money(t, "250", "PLN"))
	if err != nil {
		t.Fatal(err)
	}
	d, err := ledger.NewDeposit(ledger.TransactionHeader{Key: key("d1"), Account: "acc-2", Date: date(t, "2026-03-03")}, money(t, "249.50", "PLN"))
	if err != nil {
		t.Fatal(err)
	}
	l.txs = append(l.txs, w, d)
	l.accounts = append(l.accounts, ledger.Account{ID: "acc-2", Institution: "broker-b", Name: "Synthetic B", Sleeve: "core"})
	l.links = []ledger.TransferLink{{Pair: ledger.TransferPair{Withdrawal: key("w1"), Deposit: key("d1")}, Decision: ledger.TransferConfirmed, Version: 1}}
	req, rec := newAPI(t, l).serve(t, http.MethodGet, "/v1/transfers", goodToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/transfers = %d\n%s", rec.Code, rec.Body)
	}
	loadContract(t).check(t, req, rec)
	var got struct {
		Confirmed []transferBody `json:"confirmed"`
		Broken    []transferBody `json:"broken"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decoding %s: %v", rec.Body, err)
	}
	want := []transferBody{{
		WithdrawalAccountID: "acc-1", WithdrawalDate: "2026-03-02", DepositAccountID: "acc-2", DepositDate: "2026-03-03",
		Amount: amount{Amount: "249.5", Currency: "PLN"}, WithdrawalSource: "test", WithdrawalExternalID: "w1",
		DepositSource: "test", DepositExternalID: "d1",
	}}
	if diff := cmp.Diff(want, got.Broken); diff != "" || len(got.Confirmed) != 0 {
		t.Errorf("broken (-want +got):\n%s\nconfirmed: %v", diff, got.Confirmed)
	}
}
