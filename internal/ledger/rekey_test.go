package ledger_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestRekeyed moves a transaction's holding postings from a provisional
// instrument to the registry's (D-214), keeping everything else, so its
// new version supersedes it; one that doesn't post to the instrument, or
// a deletion, is returned as it is, and the original is never changed.
func TestRekeyed(t *testing.T) {
	t.Parallel()
	from, to := instrument(t, "PLEDO0536000", "EDO0536"), instrument(t, "PL9999000018", "EDO0536")
	other := instrument(t, "DE0000000009", "SYNT")
	buy := must(t)(ledger.NewBuy(header(t, "lot:EDO0536:2026-05-04:buy", "2026-05-04"), from, qty(t, "10"), money(t, "1000", "PLN"), money(t, "0", "PLN")))
	interest := must(t)(ledger.NewInterest(header(t, "lot:EDO0536:2026-05-04:coupon:2027-05-04", "2027-05-04"), from, money(t, "61.10", "PLN"), money(t, "11.61", "PLN")))
	unrelated := must(t)(ledger.NewBuy(header(t, "b-other", "2026-05-04"), other, qty(t, "1"), money(t, "10", "PLN"), money(t, "0", "PLN")))
	deletion := must(t)(ledger.NewDeletion(header(t, "lot:EDO0536:2026-05-04:buy", "2026-05-04")))
	tests := []struct {
		name  string
		tx    ledger.Transaction
		moved bool
	}{
		{name: "a buy", tx: buy, moved: true},
		{name: "a coupon booked as interest", tx: interest, moved: true},
		{name: "another instrument's buy", tx: unrelated},
		{name: "a deletion", tx: deletion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			original := ledger.Fingerprint(&tt.tx)
			got, moved := ledger.Rekeyed(&tt.tx, from, to)
			if moved != tt.moved || ledger.Fingerprint(&tt.tx) != original {
				t.Fatalf("Rekeyed moved = %t, want %t, and the original changed: %t", moved, tt.moved, ledger.Fingerprint(&tt.tx) != original)
			}
			if diff := cmp.Diff(byHand(&tt.tx, from, to), got); diff != "" {
				t.Errorf("Rekeyed (-want +got):\n%s", diff)
			}
			if err := got.Validate(); err != nil {
				t.Errorf("Rekeyed is not valid: %v", err)
			}
			if tt.moved == (ledger.Fingerprint(&got) == original) {
				t.Errorf("Rekeyed's fingerprint is the original's: %t, want %t", !tt.moved, tt.moved)
			}
		})
	}
}

// byHand returns tx with its holding postings of from moved to the same
// holding of to.
func byHand(tx *ledger.Transaction, from, to ledger.Instrument) ledger.Transaction {
	out := *tx
	out.Postings = append([]ledger.Posting(nil), tx.Postings...)
	for i := range out.Postings {
		if out.Postings[i].Book == ledger.HoldingBook(from) {
			out.Postings[i].Book = ledger.HoldingBook(to)
		}
	}
	return out
}

// TestRekeyedNames carries a provisional instrument's display names over to
// the registry's ISIN (D-197, D-214): its first delivered name and the
// owner's latest rename, each only while the registry's ISIN has none of
// its kind; carried over once, it carries nothing again.
func TestRekeyedNames(t *testing.T) {
	t.Parallel()
	from, to := instrument(t, "PLEDO0536000", "EDO0536").ISIN, instrument(t, "PL9999000018", "EDO0536").ISIN
	other := instrument(t, "DE0000000009", "SYNT").ISIN
	delivered := ledger.InstrumentName{ISIN: from, Name: "EDO0536 delivered", Source: "bonds-main"}
	later := ledger.InstrumentName{ISIN: from, Name: "EDO0536 later", Source: "bonds-other"}
	renamed := ledger.InstrumentName{ISIN: from, Name: "My first EDO", ByOwner: true}
	renamedAgain := ledger.InstrumentName{ISIN: from, Name: "My EDO", ByOwner: true}
	at := func(n ledger.InstrumentName) ledger.InstrumentName { n.ISIN = to; return n }
	tests := []struct {
		name    string
		records []ledger.InstrumentName
		want    []ledger.InstrumentName
	}{
		{name: "nothing named", records: []ledger.InstrumentName{{ISIN: other, Name: "SYNT"}}},
		{name: "the first delivered and the latest rename", records: []ledger.InstrumentName{delivered, later, renamed, renamedAgain}, want: []ledger.InstrumentName{at(delivered), at(renamedAgain)}},
		{name: "the registry's ISIN delivered already", records: []ledger.InstrumentName{delivered, renamed, at(later)}, want: []ledger.InstrumentName{at(renamed)}},
		{name: "carried over already", records: []ledger.InstrumentName{delivered, renamed, at(delivered), at(renamed)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, ledger.RekeyedNames(tt.records, from, to)); diff != "" {
				t.Errorf("RekeyedNames (-want +got):\n%s", diff)
			}
		})
	}
}
