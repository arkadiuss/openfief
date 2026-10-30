package postgres_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// linkPair returns the pair of withdrawal w and deposit d of source "test".
func linkPair(w, d string) ledger.TransferPair {
	return ledger.TransferPair{
		Withdrawal: ledger.TransactionKey{Source: "test", ExternalID: w},
		Deposit:    ledger.TransactionKey{Source: "test", ExternalID: d},
	}
}

func TestTransferLinksRoundTrip(t *testing.T) {
	t.Parallel()
	store, _ := seededStore(t)
	ctx := t.Context()
	links := []ledger.TransferLink{
		{Pair: linkPair("w2", "d2"), Decision: ledger.TransferRejected, Version: 1},
		{Pair: linkPair("w1", "d1"), Decision: ledger.TransferConfirmed, Version: 1},
		{Pair: linkPair("w1", "d1"), Decision: ledger.TransferRejected, Version: 2},
	}
	for i := range links {
		if err := store.AppendTransferLink(ctx, &links[i]); err != nil {
			t.Fatalf("AppendTransferLink(%v): %v", links[i], err)
		}
	}
	got, err := store.TransferLinks(ctx)
	checkResult(t, "TransferLinks", []ledger.TransferLink{links[1], links[2], links[0]}, got, err)

	err = store.AppendTransferLink(ctx, &ledger.TransferLink{Pair: linkPair("w1", "d1"), Decision: ledger.TransferConfirmed, Version: 2})
	if !errors.Is(err, app.ErrVersionConflict) {
		t.Errorf("appending a stored version: error = %v, want ErrVersionConflict", err)
	}
	err = store.AppendTransferLink(ctx, &ledger.TransferLink{Pair: linkPair("w3", "d3"), Decision: ledger.TransferConfirmed, Version: 2})
	if sqlState(err) != versionGap {
		t.Errorf("appending version 2 of a new pair: error = %v, want SQLSTATE %s", err, versionGap)
	}
	if err := store.LockTransferLinks(ctx); err == nil {
		t.Error("LockTransferLinks outside InTx succeeded")
	}
}

func TestTransferLinksAreAppendOnly(t *testing.T) {
	t.Parallel()
	store, pool := seededStore(t)
	link := ledger.TransferLink{Pair: linkPair("w", "d"), Decision: ledger.TransferConfirmed, Version: 1}
	if err := store.AppendTransferLink(t.Context(), &link); err != nil {
		t.Fatalf("AppendTransferLink: %v", err)
	}
	for _, sql := range []string{
		"UPDATE ledger.transfer_links SET decision = 'rejected'",
		"DELETE FROM ledger.transfer_links",
		"TRUNCATE ledger.transfer_links",
	} {
		if _, err := pool.Exec(t.Context(), sql); sqlState(err) != appendOnlyViolation {
			t.Errorf("%s: error %v, want SQLSTATE %s", sql, err, appendOnlyViolation)
		}
	}
	got, err := store.TransferLinks(t.Context())
	checkResult(t, "TransferLinks after refused changes", []ledger.TransferLink{link}, got, err)
}

func TestDecideTransferOverPostgres(t *testing.T) {
	t.Parallel()
	store, _ := seededStore(t)
	l := app.New(store)
	txs := []ledger.Transaction{
		must(t)(ledger.NewWithdrawal(header(t, acc1, "w1", "2026-03-02"), money(t, "100", "PLN"))),
		must(t)(ledger.NewWithdrawal(header(t, acc1, "w2", "2026-03-02"), money(t, "100", "PLN"))),
		must(t)(ledger.NewDeposit(header(t, acc2, "d", "2026-03-04"), money(t, "100", "PLN"))),
	}
	if _, err := l.Import(t.Context(), txs, nil); err != nil {
		t.Fatalf("Import: %v", err)
	}
	// Two confirmations race for the one deposit: the lock lets one check
	// and append at a time, so exactly one wins.
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i, w := range []string{"w1", "w2"} {
		wg.Go(func() {
			_, errs[i] = l.DecideTransfer(t.Context(), linkPair(w, "d"), ledger.TransferConfirmed)
		})
	}
	wg.Wait()
	taken := 0
	for _, err := range errs {
		switch {
		case errors.Is(err, ledger.ErrTransferTaken):
			taken++
		case err != nil:
			t.Fatalf("DecideTransfer: %v", err)
		}
	}
	links, err := l.TransferLinks(t.Context())
	if err != nil || len(links) != 1 || taken != 1 {
		t.Fatalf("links = %v (%v), refused %d; want one link and one refusal", links, err, taken)
	}
	corrected, err := l.DecideTransfer(t.Context(), links[0].Pair, ledger.TransferRejected)
	checkResult(t, "DecideTransfer (correction)",
		ledger.TransferLink{Pair: links[0].Pair, Decision: ledger.TransferRejected, Version: 2}, corrected, err)
}
