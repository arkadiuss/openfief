package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestRejectingABrokenTransfer(t *testing.T) {
	t.Parallel()
	l, _ := transferLedger(t)
	if _, err := l.DecideTransfer(t.Context(), tpair("w", "d1"), ledger.TransferConfirmed); err != nil {
		t.Fatalf("confirming: %v", err)
	}
	corrected := must(t)(ledger.NewDeposit(header(t, acc2, "d1", "2026-03-03"), money(t, "99.50", "PLN")))
	if _, err := l.Import(t.Context(), []ledger.Transaction{corrected}, nil); err != nil {
		t.Fatalf("importing the correction: %v", err)
	}
	_, err := l.DecideTransfer(t.Context(), tpair("w", "d1"), ledger.TransferConfirmed)
	if diff := cmp.Diff(ledger.ErrInvalidTransfer, err, cmpopts.EquateErrors()); diff != "" {
		t.Fatalf("confirming the corrected pair again: error (-want +got):\n%s", diff)
	}
	got, err := l.DecideTransfer(t.Context(), tpair("w", "d1"), ledger.TransferRejected)
	want := ledger.TransferLink{Pair: tpair("w", "d1"), Decision: ledger.TransferRejected, Version: 2}
	checkDecision(t, 0, &want, &got, nil, err)
	_, err = l.DecideTransfer(t.Context(), tpair("w2", "d2"), ledger.TransferRejected)
	if diff := cmp.Diff(ledger.ErrInvalidTransfer, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("rejecting a pair that was never confirmed still checks its legs (-want +got):\n%s", diff)
	}
}
