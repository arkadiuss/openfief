package app_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// errOwnerCheck is what fakeOwners answers for any token but "owner".
var errOwnerCheck = app.ErrNotOwner

// fakeOwners accepts the session token "owner" only, and refuses it a
// change to the accounts in denied, as an authorizer that denies some
// accounts would (D-163).
type fakeOwners struct {
	denied []ledger.AccountID
}

func (fakeOwners) CheckOwner(_ context.Context, _, token string) error {
	if token != "owner" {
		return errOwnerCheck
	}
	return nil
}

func (f fakeOwners) Changeable(ctx context.Context, origin, token string, accounts []ledger.AccountID) ([]ledger.AccountID, error) {
	if err := f.CheckOwner(ctx, origin, token); err != nil {
		return nil, err
	}
	return slices.DeleteFunc(slices.Clone(accounts), func(a ledger.AccountID) bool { return slices.Contains(f.denied, a) }), nil
}

// transferLedger returns the use cases over a memStore holding, in acc-1,
// a withdrawal w of 100 PLN on 2026-03-02 and in acc-2 the deposits d1 of
// 100 PLN on 2026-03-03 and d2 of 100 PLN on 2026-03-20.
func transferLedger(t *testing.T) (*app.Ledger, *memStore) {
	t.Helper()
	l, store := seeded(t)
	txs := []ledger.Transaction{
		must(t)(ledger.NewWithdrawal(header(t, acc1, "w", "2026-03-02"), money(t, "100", "PLN"))),
		must(t)(ledger.NewDeposit(header(t, acc2, "d1", "2026-03-03"), money(t, "100", "PLN"))),
		must(t)(ledger.NewDeposit(header(t, acc2, "d2", "2026-03-20"), money(t, "100", "PLN"))),
		must(t)(ledger.NewWithdrawal(header(t, acc1, "w2", "2026-03-04"), money(t, "100", "PLN"))),
	}
	if _, err := l.Import(t.Context(), txs, nil); err != nil {
		t.Fatalf("Import: %v", err)
	}
	return l, store
}

// tpair returns the pair of withdrawal w and deposit d of source "test".
func tpair(w, d string) ledger.TransferPair {
	return ledger.TransferPair{
		Withdrawal: ledger.TransactionKey{Source: "test", ExternalID: w},
		Deposit:    ledger.TransactionKey{Source: "test", ExternalID: d},
	}
}

func TestDecideTransfer(t *testing.T) {
	t.Parallel()
	confirmed, rejected := ledger.TransferConfirmed, ledger.TransferRejected
	type step struct {
		pair     ledger.TransferPair
		decision ledger.TransferDecision
		want     ledger.TransferLink
		wantErr  error
	}
	tests := []struct {
		name      string
		steps     []step
		wantLinks []ledger.TransferLink
		wantAll   int
	}{
		{
			"confirm",
			[]step{{tpair("w", "d1"), confirmed, ledger.TransferLink{Pair: tpair("w", "d1"), Decision: confirmed, Version: 1}, nil}},
			[]ledger.TransferLink{{Pair: tpair("w", "d1"), Decision: confirmed, Version: 1}},
			1,
		},
		{
			"confirming twice appends once",
			[]step{
				{tpair("w", "d1"), confirmed, ledger.TransferLink{Pair: tpair("w", "d1"), Decision: confirmed, Version: 1}, nil},
				{tpair("w", "d1"), confirmed, ledger.TransferLink{Pair: tpair("w", "d1"), Decision: confirmed, Version: 1}, nil},
			},
			[]ledger.TransferLink{{Pair: tpair("w", "d1"), Decision: confirmed, Version: 1}},
			1,
		},
		{
			"a correction supersedes and keeps the earlier version",
			[]step{
				{tpair("w", "d1"), confirmed, ledger.TransferLink{Pair: tpair("w", "d1"), Decision: confirmed, Version: 1}, nil},
				{tpair("w", "d1"), rejected, ledger.TransferLink{Pair: tpair("w", "d1"), Decision: rejected, Version: 2}, nil},
			},
			[]ledger.TransferLink{{Pair: tpair("w", "d1"), Decision: rejected, Version: 2}},
			2,
		},
		{
			"a taken deposit is refused",
			[]step{
				{tpair("w", "d1"), confirmed, ledger.TransferLink{Pair: tpair("w", "d1"), Decision: confirmed, Version: 1}, nil},
				{tpair("w2", "d1"), confirmed, ledger.TransferLink{}, ledger.ErrTransferTaken},
			},
			[]ledger.TransferLink{{Pair: tpair("w", "d1"), Decision: confirmed, Version: 1}},
			1,
		},
		{
			"too far apart",
			[]step{{tpair("w", "d2"), confirmed, ledger.TransferLink{}, ledger.ErrInvalidTransfer}},
			nil, 0,
		},
		{
			"a deposit as the withdrawal",
			[]step{{tpair("d1", "d2"), rejected, ledger.TransferLink{}, ledger.ErrInvalidTransfer}},
			nil, 0,
		},
		{
			"unknown leg",
			[]step{{tpair("w", "nope"), confirmed, ledger.TransferLink{}, app.ErrUnknownTransaction}},
			nil, 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, store := transferLedger(t)
			for i, s := range tt.steps {
				got, err := l.DecideTransfer(t.Context(), s.pair, s.decision)
				checkDecision(t, i, &s.want, &got, s.wantErr, err)
			}
			links, err := l.TransferLinks(t.Context())
			checkResult(t, "TransferLinks", tt.wantLinks, links, err)
			all, err := store.TransferLinks(t.Context())
			if err != nil || len(all) != tt.wantAll {
				t.Errorf("stored versions = %d (%v), want %d", len(all), err, tt.wantAll)
			}
		})
	}
}

// checkDecision fails t unless step i decided want, or failed with wantErr.
func checkDecision(t *testing.T, i int, want, got *ledger.TransferLink, wantErr, err error) {
	t.Helper()
	if diff := cmp.Diff(wantErr, err, cmpopts.EquateErrors()); diff != "" {
		t.Fatalf("step %d: DecideTransfer error mismatch (-want +got):\n%s", i, diff)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("step %d: DecideTransfer mismatch (-want +got):\n%s", i, diff)
	}
}

func TestTransfersNeedTheOwner(t *testing.T) {
	t.Parallel()
	l, _ := transferLedger(t)
	tr := &app.Transfers{Owners: fakeOwners{}, Ledger: l}
	if _, err := tr.Decide(t.Context(), "https://ui.example.com", "read", tpair("w", "d1"), ledger.TransferConfirmed); !errors.Is(err, app.ErrNotOwner) {
		t.Fatalf("Decide by a reader: error = %v, want ErrNotOwner", err)
	}
	if links, err := l.TransferLinks(t.Context()); err != nil || len(links) != 0 {
		t.Fatalf("a refused decision stored %v (%v)", links, err)
	}
	got, err := tr.Decide(t.Context(), "https://ui.example.com", "owner", tpair("w", "d1"), ledger.TransferConfirmed)
	checkResult(t, "Decide", ledger.TransferLink{Pair: tpair("w", "d1"), Decision: ledger.TransferConfirmed, Version: 1}, got, err)
}

func TestTransfersNeedTheChangeOfBothAccounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		denied  []ledger.AccountID
		pair    ledger.TransferPair
		wantErr error
	}{
		{name: "both accounts changeable", pair: tpair("w", "d1")},
		{name: "the withdrawal's account denied", denied: []ledger.AccountID{acc1}, pair: tpair("w", "d1"), wantErr: app.ErrNotOwner},
		{name: "the deposit's account denied", denied: []ledger.AccountID{acc2}, pair: tpair("w", "d1"), wantErr: app.ErrNotOwner},
		{name: "an unstored leg asks only for the stored one", denied: []ledger.AccountID{acc2}, pair: tpair("w", "nope"), wantErr: app.ErrUnknownTransaction},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, _ := transferLedger(t)
			tr := &app.Transfers{Owners: fakeOwners{denied: tt.denied}, Ledger: l}
			_, err := tr.Decide(t.Context(), "https://ui.example.com", "owner", tt.pair, ledger.TransferConfirmed)
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Fatalf("Decide error = %v, want %v", err, tt.wantErr)
			}
			links, err := l.TransferLinks(t.Context())
			if err != nil {
				t.Fatalf("TransferLinks: %v", err)
			}
			if stored := len(links) == 1; stored != (tt.wantErr == nil) {
				t.Errorf("links = %v after error %v", links, tt.wantErr)
			}
		})
	}
}

func TestTransferAccounts(t *testing.T) {
	t.Parallel()
	l, _ := transferLedger(t)
	tests := []struct {
		name string
		pair ledger.TransferPair
		want []ledger.AccountID
	}{
		{name: "both legs", pair: tpair("w", "d1"), want: []ledger.AccountID{acc1, acc2}},
		{name: "one leg stored", pair: tpair("w", "nope"), want: []ledger.AccountID{acc1}},
		{name: "both legs in one account", pair: tpair("w", "w2"), want: []ledger.AccountID{acc1}},
		{name: "none stored", pair: tpair("x", "y")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := l.TransferAccounts(t.Context(), tt.pair)
			if err != nil {
				t.Fatalf("TransferAccounts: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("TransferAccounts mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConcurrentConfirmationsTakeALegOnce(t *testing.T) {
	t.Parallel()
	l, _ := transferLedger(t)
	errs := make(chan error, 2)
	for _, w := range []string{"w", "w2"} {
		go func() {
			_, err := l.DecideTransfer(context.Background(), tpair(w, "d1"), ledger.TransferConfirmed)
			errs <- err
		}()
	}
	var taken int
	for range 2 {
		err := <-errs
		switch {
		case errors.Is(err, ledger.ErrTransferTaken):
			taken++
		case err != nil:
			t.Fatalf("DecideTransfer: %v", err)
		}
	}
	links, err := l.TransferLinks(t.Context())
	if err != nil || len(links) != 1 || taken != 1 {
		t.Errorf("links = %v (%v), refused %d; want one link and one refusal", links, err, taken)
	}
}
