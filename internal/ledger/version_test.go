package ledger_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestDecideImport(t *testing.T) {
	t.Parallel()
	in := instrument(t, isinA, "TSTA")
	h := header(t, "b-1", "2026-03-02")
	original := must(t)(ledger.NewBuy(h, in, qty(t, "10"), money(t, "1500.50", "USD"), money(t, "4", "PLN")))
	// The same row again, with postings in another order and amounts in another scale.
	reordered := original
	reordered.Postings = []ledger.Posting{original.Postings[3], original.Postings[2], original.Postings[1], original.Postings[0]}
	reordered.Postings[3].Amount = money(t, "1500.500", "USD")
	corrected := must(t)(ledger.NewBuy(h, in, qty(t, "10"), money(t, "1500.75", "USD"), money(t, "4", "PLN")))
	described := original
	described.Description = "corrected by the broker"
	other := must(t)(ledger.NewBuy(header(t, "b-2", "2026-03-02"), in, qty(t, "1"), money(t, "1", "USD"), kernel.Money{}))
	deletion := must(t)(ledger.NewDeletion(h))
	v := func(tx ledger.Transaction, n int) ledger.VersionedTransaction {
		return ledger.VersionedTransaction{Transaction: tx, Version: n}
	}
	tests := []struct {
		name     string
		versions []ledger.VersionedTransaction
		incoming ledger.Transaction
		want     ledger.ImportResult
		wantErr  error
	}{
		{name: "new key", incoming: original, want: ledger.ImportResult{Decision: ledger.ImportNew, Version: 1}},
		{name: "same row again", versions: []ledger.VersionedTransaction{v(original, 1)}, incoming: original, want: ledger.ImportResult{Decision: ledger.ImportUnchanged, Version: 1}},
		{name: "same content in another order and scale", versions: []ledger.VersionedTransaction{v(original, 1)}, incoming: reordered, want: ledger.ImportResult{Decision: ledger.ImportUnchanged, Version: 1}},
		{name: "changed amount", versions: []ledger.VersionedTransaction{v(original, 1)}, incoming: corrected, want: ledger.ImportResult{Decision: ledger.ImportNewVersion, Version: 2, Supersedes: 1}},
		{name: "changed description", versions: []ledger.VersionedTransaction{v(original, 1)}, incoming: described, want: ledger.ImportResult{Decision: ledger.ImportNewVersion, Version: 2, Supersedes: 1}},
		{name: "corrected row again", versions: []ledger.VersionedTransaction{v(corrected, 2), v(original, 1)}, incoming: corrected, want: ledger.ImportResult{Decision: ledger.ImportUnchanged, Version: 2}},
		{name: "back to the first content", versions: []ledger.VersionedTransaction{v(original, 1), v(corrected, 2)}, incoming: original, want: ledger.ImportResult{Decision: ledger.ImportNewVersion, Version: 3, Supersedes: 2}},
		{name: "deletion supersedes the current version", versions: []ledger.VersionedTransaction{v(original, 1)}, incoming: deletion, want: ledger.ImportResult{Decision: ledger.ImportNewVersion, Version: 2, Supersedes: 1}},
		{name: "deleting again changes nothing", versions: []ledger.VersionedTransaction{v(original, 1), v(deletion, 2)}, incoming: deletion, want: ledger.ImportResult{Decision: ledger.ImportUnchanged, Version: 2}},
		{name: "entering a deleted row again brings it back", versions: []ledger.VersionedTransaction{v(original, 1), v(deletion, 2)}, incoming: original, want: ledger.ImportResult{Decision: ledger.ImportNewVersion, Version: 3, Supersedes: 2}},
		{name: "deleting what was never recorded", incoming: deletion, wantErr: ledger.ErrNothingToDelete},
		{name: "a version of another key", versions: []ledger.VersionedTransaction{v(other, 1)}, incoming: original, wantErr: ledger.ErrInvalidLog},
		{name: "a repeated version", versions: []ledger.VersionedTransaction{v(original, 1), v(corrected, 1)}, incoming: original, wantErr: ledger.ErrInvalidLog},
		{name: "a gap in versions", versions: []ledger.VersionedTransaction{v(original, 1), v(corrected, 3)}, incoming: original, wantErr: ledger.ErrInvalidLog},
		{name: "version 0", versions: []ledger.VersionedTransaction{v(original, 0)}, incoming: original, wantErr: ledger.ErrInvalidLog},
		{name: "invalid incoming", incoming: ledger.Transaction{Key: h.Key}, wantErr: ledger.ErrInvalidTransaction},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.DecideImport(tt.versions, &tt.incoming)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("DecideImport error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DecideImport (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCurrentTransactions(t *testing.T) {
	t.Parallel()
	in := instrument(t, isinA, "TSTA")
	first := must(t)(ledger.NewBuy(header(t, "b-1", "2026-03-04"), in, qty(t, "10"), money(t, "100", "USD"), kernel.Money{}))
	fixed := must(t)(ledger.NewBuy(header(t, "b-1", "2026-03-04"), in, qty(t, "12"), money(t, "120", "USD"), kernel.Money{}))
	earlier := must(t)(ledger.NewDeposit(header(t, "d-1", "2026-03-02"), money(t, "500", "USD")))
	sameDay := must(t)(ledger.NewDeposit(header(t, "a-1", "2026-03-04"), money(t, "1", "USD")))
	deleted := must(t)(ledger.NewDeletion(header(t, "b-1", "2026-03-04")))
	tests := []struct {
		name    string
		log     []ledger.VersionedTransaction
		want    []ledger.Transaction
		wantErr error
	}{
		{name: "empty log", want: []ledger.Transaction{}},
		{
			name: "superseded version left out, date then key order",
			log: []ledger.VersionedTransaction{
				{Transaction: first, Version: 1},
				{Transaction: sameDay, Version: 1},
				{Transaction: fixed, Version: 2},
				{Transaction: earlier, Version: 1},
			},
			want: []ledger.Transaction{earlier, sameDay, fixed},
		},
		{
			name: "deleted key left out, its versions kept in the log",
			log: []ledger.VersionedTransaction{
				{Transaction: first, Version: 1},
				{Transaction: deleted, Version: 2},
				{Transaction: earlier, Version: 1},
			},
			want: []ledger.Transaction{earlier},
		},
		{
			name: "entered again after a deletion",
			log: []ledger.VersionedTransaction{
				{Transaction: first, Version: 1},
				{Transaction: deleted, Version: 2},
				{Transaction: fixed, Version: 3},
			},
			want: []ledger.Transaction{fixed},
		},
		{
			name:    "repeated version",
			log:     []ledger.VersionedTransaction{{Transaction: first, Version: 1}, {Transaction: fixed, Version: 1}},
			wantErr: ledger.ErrInvalidLog,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.CurrentTransactions(tt.log)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("CurrentTransactions error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("CurrentTransactions (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFingerprintCoversEveryField(t *testing.T) {
	t.Parallel()
	in, other := instrument(t, isinA, "TSTA"), instrument(t, isinB, "TSTB")
	base := must(t)(ledger.NewOpeningLot(header(t, "o-1", "2026-03-02"), in, qty(t, "10"), money(t, "812.34", "USD"), date(t, "2019-05-17")))
	base.Description = "opening lot"
	otherDay, otherOpen := date(t, "2026-03-03"), date(t, "2019-05-18")
	moreUnits, moreCost, lessCost := qty(t, "10.001"), money(t, "812.35", "USD"), money(t, "-812.35", "USD")
	inEUR, inEURNeg := money(t, "812.34", "EUR"), money(t, "-812.34", "EUR")
	eur := cur(t, "EUR")
	tests := []struct {
		name string
		edit func(tx *ledger.Transaction)
	}{
		{name: "source", edit: func(tx *ledger.Transaction) { tx.Key.Source = "other" }},
		{name: "external ID", edit: func(tx *ledger.Transaction) { tx.Key.ExternalID = "o-2" }},
		{name: "account", edit: func(tx *ledger.Transaction) { tx.Account = "acc-2" }},
		{name: "kind", edit: func(tx *ledger.Transaction) { tx.Kind = ledger.KindBuy }},
		{name: "date", edit: func(tx *ledger.Transaction) { tx.Date = otherDay }},
		{name: "lot open date", edit: func(tx *ledger.Transaction) { tx.LotOpened = otherOpen }},
		{name: "description", edit: func(tx *ledger.Transaction) { tx.Description = "" }},
		{name: "instrument", edit: func(tx *ledger.Transaction) { tx.Postings[0].Book = ledger.HoldingBook(other) }},
		{name: "quantity", edit: func(tx *ledger.Transaction) { tx.Postings[0].Quantity = moreUnits }},
		{name: "amount", edit: func(tx *ledger.Transaction) { tx.Postings[0].Amount, tx.Postings[1].Amount = moreCost, lessCost }},
		{name: "currency", edit: func(tx *ledger.Transaction) {
			tx.Postings[0].Amount, tx.Postings[1] = inEUR, ledger.Posting{Book: ledger.EquityBook(eur), Amount: inEURNeg}
		}},
	}
	want := ledger.Fingerprint(&base)
	if copied := base; ledger.Fingerprint(&copied) != want {
		t.Fatalf("Fingerprint of a copy differs")
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tx := base
			tx.Postings = append([]ledger.Posting(nil), base.Postings...)
			tt.edit(&tx)
			if ledger.Fingerprint(&tx) == want {
				t.Errorf("changing the %s left the fingerprint unchanged", tt.name)
			}
		})
	}
}

func TestImportDecisionNames(t *testing.T) {
	t.Parallel()
	got := []string{}
	for d := ledger.ImportDecision(0); d <= ledger.ImportNewVersion+1; d++ {
		got = append(got, d.String())
	}
	want := []string{"ImportDecision(0)", "unchanged", "new", "new-version", "ImportDecision(4)"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ImportDecision.String (-want +got):\n%s", diff)
	}
}
