package ledger_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// seedFixture holds synthetic transactions of testAccount around a go-live
// on 2026-03-31.
type seedFixture struct {
	lot, lotCorrected, cash, cashCorrected, deposit, buy, otherLot ledger.Transaction
}

func newSeedFixture(t *testing.T) seedFixture {
	t.Helper()
	in := instrument(t, isinA, "TSTA")
	lotHeader := header(t, "opening-lot:"+isinA+":2025-06-02:n1", "2026-03-31")
	cashHeader := header(t, ledger.OpeningCashID(cur(t, "USD"), date(t, "2026-03-31")), "2026-03-31")
	other := header(t, "opening-lot:"+isinA+":2025-06-02:n1", "2026-03-31")
	other.Account = "acc-2"
	return seedFixture{
		lot:           must(t)(ledger.NewOpeningLot(lotHeader, in, qty(t, "10"), money(t, "1500", "USD"), date(t, "2025-06-02"))),
		lotCorrected:  must(t)(ledger.NewOpeningLot(lotHeader, in, qty(t, "10"), money(t, "1510", "USD"), date(t, "2025-06-02"))),
		cash:          must(t)(ledger.NewOpeningCash(cashHeader, money(t, "250", "USD"))),
		cashCorrected: must(t)(ledger.NewOpeningCash(cashHeader, money(t, "-20", "USD"))),
		deposit:       must(t)(ledger.NewDeposit(header(t, "d-1", "2026-04-02"), money(t, "100", "USD"))),
		buy:           must(t)(ledger.NewBuy(header(t, "b-1", "2026-04-03"), in, qty(t, "1"), money(t, "50", "USD"), kernel.Money{})),
		otherLot:      must(t)(ledger.NewOpeningLot(other, in, qty(t, "1"), money(t, "150", "USD"), date(t, "2025-06-02"))),
	}
}

// v1 returns ts as version 1 each.
func v1(ts ...ledger.Transaction) []ledger.VersionedTransaction {
	out := make([]ledger.VersionedTransaction, len(ts))
	for i := range ts {
		out[i] = ledger.VersionedTransaction{Transaction: ts[i], Version: 1}
	}
	return out
}

func TestOpeningCashID(t *testing.T) {
	t.Parallel()
	got := ledger.OpeningCashID(cur(t, "EUR"), date(t, "2026-03-31"))
	if want := "opening-cash:EUR:2026-03-31"; got != want {
		t.Errorf("OpeningCashID = %q, want %q", got, want)
	}
}

func TestNewOpeningCash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		id       string
		cash     string
		wantKind ledger.TransactionKind
		// wantCash and wantEquity are the amounts posted to cash and equity.
		wantCash, wantEquity string
		wantErr              error
	}{
		{name: "positive balance is a deposit", id: "opening-cash:USD:2026-03-31", cash: "250.10", wantKind: ledger.KindDeposit, wantCash: "250.10", wantEquity: "-250.10"},
		{name: "negative balance is a withdrawal", id: "opening-cash:USD:2026-03-31", cash: "-20", wantKind: ledger.KindWithdrawal, wantCash: "-20", wantEquity: "20"},
		{name: "zero balance", id: "opening-cash:USD:2026-03-31", cash: "0", wantErr: ledger.ErrInvalidTransaction},
		{name: "key without the opening cash prefix", id: "d-1", cash: "250", wantErr: ledger.ErrInvalidTransaction},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.NewOpeningCash(header(t, tt.id, "2026-03-31"), money(t, tt.cash, "USD"))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("NewOpeningCash error (-want +got):\n%s", diff)
			}
			if tt.wantErr != nil {
				return
			}
			usd := cur(t, "USD")
			want := ledger.Transaction{
				Key: ledger.TransactionKey{Source: "test", ExternalID: tt.id}, Account: testAccount, Kind: tt.wantKind, Date: date(t, "2026-03-31"),
				Postings: []ledger.Posting{
					{Book: ledger.CashBook(usd), Amount: money(t, tt.wantCash, "USD")},
					{Book: ledger.EquityBook(usd), Amount: money(t, tt.wantEquity, "USD")},
				},
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("NewOpeningCash (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIsSeeding(t *testing.T) {
	t.Parallel()
	f := newSeedFixture(t)
	fee := must(t)(ledger.NewFee(header(t, "opening-cash:USD:2026-03-31", "2026-03-31"), money(t, "1", "USD")))
	tests := []struct {
		name string
		tx   ledger.Transaction
		want bool
	}{
		{name: "opening lot", tx: f.lot, want: true},
		{name: "opening cash deposit", tx: f.cash, want: true},
		{name: "opening cash withdrawal", tx: f.cashCorrected, want: true},
		{name: "ordinary deposit", tx: f.deposit},
		{name: "buy", tx: f.buy},
		{name: "fee keyed like opening cash", tx: fee},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ledger.IsSeeding(&tt.tx); got != tt.want {
				t.Errorf("IsSeeding = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSeeded(t *testing.T) {
	t.Parallel()
	f := newSeedFixture(t)
	movedLot := f.lot
	movedLot.Account = "acc-2"
	tests := []struct {
		name    string
		log     []ledger.VersionedTransaction
		want    bool
		wantErr error
	}{
		{name: "empty log"},
		{name: "transactions only", log: v1(f.deposit, f.buy)},
		{name: "an opening lot", log: v1(f.deposit, f.lot), want: true},
		{name: "opening cash only", log: v1(f.cash), want: true},
		{name: "another account's opening lot", log: v1(f.otherLot)},
		{name: "opening lot moved to another account by a correction", log: []ledger.VersionedTransaction{{Transaction: f.lot, Version: 1}, {Transaction: movedLot, Version: 2}}},
		{name: "invalid log", log: []ledger.VersionedTransaction{{Transaction: f.lot, Version: 2}}, wantErr: ledger.ErrInvalidLog},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.IsSeeded(tt.log, testAccount)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("IsSeeded error (-want +got):\n%s", diff)
			}
			if got != tt.want {
				t.Errorf("IsSeeded = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecideSeed(t *testing.T) {
	t.Parallel()
	f := newSeedFixture(t)
	tests := []struct {
		name     string
		log      []ledger.VersionedTransaction
		incoming []ledger.Transaction
		want     ledger.SeedDecision
		wantErr  error
	}{
		{name: "first seeding", incoming: []ledger.Transaction{f.lot, f.cash}, want: ledger.SeedFirst},
		{name: "first seeding after transactions", log: v1(f.deposit), incoming: []ledger.Transaction{f.lot}, want: ledger.SeedFirst},
		{name: "another account is seeded", log: v1(f.otherLot), incoming: []ledger.Transaction{f.lot}, want: ledger.SeedFirst},
		{name: "same seeding again", log: v1(f.lot, f.cash), incoming: []ledger.Transaction{f.cash, f.lot}, want: ledger.SeedUnchanged},
		{name: "part of the same seeding again", log: v1(f.lot, f.cash), incoming: []ledger.Transaction{f.cash}, want: ledger.SeedUnchanged},
		{name: "same seeding with another account's rows", log: v1(f.lot), incoming: []ledger.Transaction{f.lot, f.otherLot}, want: ledger.SeedUnchanged},
		{name: "changed opening lot", log: v1(f.lot, f.cash), incoming: []ledger.Transaction{f.lotCorrected, f.cash}, wantErr: ledger.ErrAlreadySeeded},
		{name: "changed opening cash", log: v1(f.lot, f.cash), incoming: []ledger.Transaction{f.lot, f.cashCorrected}, wantErr: ledger.ErrAlreadySeeded},
		{name: "new row in a seeded account", log: v1(f.cash), incoming: []ledger.Transaction{f.cash, f.lot}, wantErr: ledger.ErrAlreadySeeded},
		{name: "invalid log", log: []ledger.VersionedTransaction{{Transaction: f.lot, Version: 2}}, incoming: []ledger.Transaction{f.lot}, wantErr: ledger.ErrInvalidLog},
		{name: "invalid incoming", log: v1(f.lot), incoming: []ledger.Transaction{{Key: f.cash.Key, Account: testAccount}}, wantErr: ledger.ErrInvalidTransaction},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.DecideSeed(tt.log, testAccount, tt.incoming)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("DecideSeed error (-want +got):\n%s", diff)
			}
			if got != tt.want {
				t.Errorf("DecideSeed = %d, want %d", got, tt.want)
			}
			if errors.Is(tt.wantErr, ledger.ErrAlreadySeeded) {
				checkAlreadySeeded(t, err)
			}
		})
	}
}

// checkAlreadySeeded fails t unless err is an *AlreadySeededError for
// testAccount with its message.
func checkAlreadySeeded(t *testing.T, err error) {
	t.Helper()
	seededErr, ok := errors.AsType[*ledger.AlreadySeededError](err)
	if !ok || seededErr.Account != testAccount {
		t.Fatalf("DecideSeed error = %v, want an *AlreadySeededError for %s", err, testAccount)
	}
	if want := "account acc-1 is already seeded: seeding it again is refused"; err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}
