package ledger_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// leg returns a deposit (positive amount) or a withdrawal (negative) keyed
// id in account on day, or opening cash when id is "opening".
func leg(t *testing.T, account, id, day, amount, code string) ledger.Transaction {
	t.Helper()
	h := ledger.TransactionHeader{
		Key: ledger.TransactionKey{Source: "test", ExternalID: id}, Account: ledger.AccountID(account), Date: date(t, day),
	}
	m := money(t, amount, code)
	if id == "opening" {
		h.Key.ExternalID = ledger.OpeningCashID(m.Currency(), h.Date)
		return must(t)(ledger.NewOpeningCash(h, m))
	}
	if m.Sign() < 0 {
		return must(t)(ledger.NewWithdrawal(h, m.Neg()))
	}
	return must(t)(ledger.NewDeposit(h, m))
}

// key returns the test source's key id.
func key(id string) ledger.TransactionKey {
	return ledger.TransactionKey{Source: "test", ExternalID: id}
}

// pair returns the pair of withdrawal w and deposit d.
func pair(w, d string) ledger.TransferPair {
	return ledger.TransferPair{Withdrawal: key(w), Deposit: key(d)}
}

func TestTransferAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, id, amount string
		want             string
		wantOK           bool
	}{
		{"deposit", "d", "100", "100", true},
		{"withdrawal", "w", "-100", "-100", true},
		{"opening cash", "opening", "100", "", false},
		{"fee", "fee", "1", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tx := leg(t, "a", tt.id, "2026-03-02", tt.amount, "PLN")
			if tt.id == "fee" {
				h := ledger.TransactionHeader{Key: key("f"), Account: "a", Date: date(t, "2026-03-02")}
				tx = must(t)(ledger.NewFee(h, money(t, tt.amount, "PLN")))
			}
			got, ok := ledger.TransferAmount(&tx)
			if ok != tt.wantOK {
				t.Fatalf("TransferAmount ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && !got.Equal(money(t, tt.want, "PLN")) {
				t.Errorf("TransferAmount = %s, want %s PLN", got, tt.want)
			}
		})
	}
}

func TestCheckTransfer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                         string
		wAccount, wID, wDay, wAmount string
		dAccount, dID, dDay, dAmount string
		dCurrency                    string
		wantErr                      bool
	}{
		{"same day", "a", "w", "2026-03-02", "-100", "b", "d", "2026-03-02", "100", "PLN", false},
		{"deposit five days later", "a", "w", "2026-03-02", "-100", "b", "d", "2026-03-07", "100", "PLN", false},
		{"deposit five days earlier", "a", "w", "2026-03-07", "-100", "b", "d", "2026-03-02", "100", "PLN", false},
		{"six days apart", "a", "w", "2026-03-02", "-100", "b", "d", "2026-03-08", "100", "PLN", true},
		{"same account", "a", "w", "2026-03-02", "-100", "a", "d", "2026-03-02", "100", "PLN", true},
		{"different amount", "a", "w", "2026-03-02", "-100", "b", "d", "2026-03-02", "100.01", "PLN", true},
		{"different currency", "a", "w", "2026-03-02", "-100", "b", "d", "2026-03-02", "100", "EUR", true},
		{"two deposits", "a", "w", "2026-03-02", "100", "b", "d", "2026-03-02", "100", "PLN", true},
		{"two withdrawals", "a", "w", "2026-03-02", "-100", "b", "d", "2026-03-02", "-100", "PLN", true},
		{"opening cash deposit", "a", "w", "2026-03-02", "-100", "b", "opening", "2026-03-02", "100", "PLN", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := leg(t, tt.wAccount, tt.wID, tt.wDay, tt.wAmount, "PLN")
			d := leg(t, tt.dAccount, tt.dID, tt.dDay, tt.dAmount, tt.dCurrency)
			err := ledger.CheckTransfer(&w, &d)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckTransfer error = %v, want error %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ledger.ErrInvalidTransfer) {
				t.Errorf("CheckTransfer error = %v, want ErrInvalidTransfer", err)
			}
		})
	}
}

func TestCurrentTransferLinks(t *testing.T) {
	t.Parallel()
	confirmed, rejected := ledger.TransferConfirmed, ledger.TransferRejected
	tests := []struct {
		name    string
		all     []ledger.TransferLink
		want    []ledger.TransferLink
		wantErr error
	}{
		{"none", nil, nil, nil},
		{
			"latest version wins, pairs in key order",
			[]ledger.TransferLink{
				{Pair: pair("w2", "d2"), Decision: confirmed, Version: 1},
				{Pair: pair("w1", "d1"), Decision: rejected, Version: 2},
				{Pair: pair("w1", "d1"), Decision: confirmed, Version: 1},
			},
			[]ledger.TransferLink{
				{Pair: pair("w1", "d1"), Decision: rejected, Version: 2},
				{Pair: pair("w2", "d2"), Decision: confirmed, Version: 1},
			},
			nil,
		},
		{"missing first version", []ledger.TransferLink{{Pair: pair("w", "d"), Decision: confirmed, Version: 2}}, nil, ledger.ErrInvalidLog},
		{
			"repeated version",
			[]ledger.TransferLink{
				{Pair: pair("w", "d"), Decision: confirmed, Version: 1},
				{Pair: pair("w", "d"), Decision: rejected, Version: 1},
			},
			nil, ledger.ErrInvalidLog,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.CurrentTransferLinks(tt.all)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("CurrentTransferLinks error mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("CurrentTransferLinks mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDecideTransferLink(t *testing.T) {
	t.Parallel()
	confirmed, rejected := ledger.TransferConfirmed, ledger.TransferRejected
	current := []ledger.TransferLink{
		{Pair: pair("w1", "d1"), Decision: confirmed, Version: 1},
		{Pair: pair("w2", "d2"), Decision: rejected, Version: 3},
	}
	tests := []struct {
		name     string
		pair     ledger.TransferPair
		decision ledger.TransferDecision
		want     ledger.TransferLink
		wantOK   bool
		wantErr  error
	}{
		{"new pair", pair("w3", "d3"), confirmed, ledger.TransferLink{Pair: pair("w3", "d3"), Decision: confirmed, Version: 1}, true, nil},
		{"new rejection", pair("w3", "d3"), rejected, ledger.TransferLink{Pair: pair("w3", "d3"), Decision: rejected, Version: 1}, true, nil},
		{"same decision again", pair("w1", "d1"), confirmed, current[0], false, nil},
		{"correction supersedes", pair("w1", "d1"), rejected, ledger.TransferLink{Pair: pair("w1", "d1"), Decision: rejected, Version: 2}, true, nil},
		{"confirming a rejected pair", pair("w2", "d2"), confirmed, ledger.TransferLink{Pair: pair("w2", "d2"), Decision: confirmed, Version: 4}, true, nil},
		{"withdrawal taken", pair("w1", "d3"), confirmed, ledger.TransferLink{}, false, ledger.ErrTransferTaken},
		{"deposit taken", pair("w3", "d1"), confirmed, ledger.TransferLink{}, false, ledger.ErrTransferTaken},
		{"rejecting with a taken leg", pair("w1", "d3"), rejected, ledger.TransferLink{Pair: pair("w1", "d3"), Decision: rejected, Version: 1}, true, nil},
		{"leg of a rejected pair is free", pair("w2", "d3"), confirmed, ledger.TransferLink{Pair: pair("w2", "d3"), Decision: confirmed, Version: 1}, true, nil},
		{"no decision", pair("w3", "d3"), 0, ledger.TransferLink{}, false, ledger.ErrInvalidTransfer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok, err := ledger.DecideTransferLink(current, tt.pair, tt.decision)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("DecideTransferLink error mismatch (-want +got):\n%s", diff)
			}
			if ok != tt.wantOK {
				t.Errorf("DecideTransferLink ok = %v, want %v", ok, tt.wantOK)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DecideTransferLink mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTransferStrings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		got, want string
	}{
		{ledger.TransferConfirmed.String(), "confirmed"},
		{ledger.TransferRejected.String(), "rejected"},
		{ledger.TransferDecision(9).String(), "TransferDecision(9)"},
		{pair("w", "d").String(), "test/w -> test/d"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("String() = %q, want %q", tt.got, tt.want)
		}
	}
}
