package ingest_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// isinOf parses a synthetic ISIN or fails the test.
func isinOf(t *testing.T, s string) ledger.ISIN {
	t.Helper()
	i, err := ledger.ParseISIN(s)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", s, err)
	}
	return i
}

// TestRegistryISINValidate: the core re-keys only a treasury bond series
// under its provisional ISIN, and only to a Polish ISIN that is no
// provisional one (D-214).
func TestRegistryISINValidate(t *testing.T) {
	t.Parallel()
	edo := ledger.Instrument{ISIN: isinOf(t, "PLEDO0536000"), Symbol: "EDO0536"}
	tests := []struct {
		name  string
		r     ingest.RegistryISIN
		valid bool
	}{
		{name: "a provisional series to its registry ISIN", r: ingest.RegistryISIN{Instrument: edo, ISIN: isinOf(t, "PL9999000018")}, valid: true},
		{name: "a registry ISIN already", r: ingest.RegistryISIN{Instrument: ledger.Instrument{ISIN: isinOf(t, "PL9999000026"), Symbol: "EDO0536"}, ISIN: isinOf(t, "PL9999000018")}},
		{name: "another series' provisional ISIN", r: ingest.RegistryISIN{Instrument: ledger.Instrument{ISIN: edo.ISIN, Symbol: "EDO1136"}, ISIN: isinOf(t, "PL9999000018")}},
		{name: "a listed instrument", r: ingest.RegistryISIN{Instrument: ledger.Instrument{ISIN: isinOf(t, "PL9999000026"), Symbol: "SYNT"}, ISIN: isinOf(t, "PL9999000018")}},
		{name: "to a non-Polish ISIN", r: ingest.RegistryISIN{Instrument: edo, ISIN: isinOf(t, "DE0000000009")}},
		{name: "to a provisional ISIN", r: ingest.RegistryISIN{Instrument: edo, ISIN: isinOf(t, "PLEDO1136008")}},
		{name: "to no ISIN", r: ingest.RegistryISIN{Instrument: edo}},
		{name: "no instrument", r: ingest.RegistryISIN{ISIN: isinOf(t, "PL9999000018")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.r.Validate()
			if (err == nil) != tt.valid || (err != nil && !errors.Is(err, ingest.ErrInvalidRegistryISIN)) {
				t.Errorf("Validate() = %v, want valid %t", err, tt.valid)
			}
		})
	}
}

// TestCheckRegistryISINs orders registry ISINs by instrument and refuses
// an instrument named twice or an invalid one.
func TestCheckRegistryISINs(t *testing.T) {
	t.Parallel()
	edo := ingest.RegistryISIN{Instrument: ledger.Instrument{ISIN: isinOf(t, "PLEDO0536000"), Symbol: "EDO0536"}, ISIN: isinOf(t, "PL9999000018")}
	coi := ingest.RegistryISIN{Instrument: ledger.Instrument{ISIN: isinOf(t, "PLCOI1130004"), Symbol: "COI1130"}, ISIN: isinOf(t, "PL9999000034")}
	got, err := ingest.CheckRegistryISINs([]ingest.RegistryISIN{edo, coi})
	if err != nil {
		t.Fatalf("CheckRegistryISINs: %v", err)
	}
	if diff := cmp.Diff([]ingest.RegistryISIN{coi, edo}, got, cmp.Comparer(ledger.ISIN.Equal)); diff != "" {
		t.Errorf("CheckRegistryISINs (-want +got):\n%s", diff)
	}
	again := edo
	again.ISIN = isinOf(t, "PL9999000026")
	for _, isins := range [][]ingest.RegistryISIN{{edo, again}, {edo, {Instrument: edo.Instrument}}} {
		if _, err := ingest.CheckRegistryISINs(isins); !errors.Is(err, ingest.ErrInvalidRegistryISIN) {
			t.Errorf("CheckRegistryISINs(%v) = %v, want ErrInvalidRegistryISIN", isins, err)
		}
	}
	if got, err := ingest.CheckRegistryISINs(nil); err != nil || len(got) != 0 {
		t.Errorf("CheckRegistryISINs(none) = %v, %v, want none", got, err)
	}
}

// TestProvisionalInstruments lists the instruments transactions post to a
// holding of under a provisional ISIN, held or not, each once in order,
// and none under a registry ISIN or another kind of instrument (D-214).
func TestProvisionalInstruments(t *testing.T) {
	t.Parallel()
	edo := ledger.Instrument{ISIN: isinOf(t, "PLEDO0536000"), Symbol: "EDO0536"}
	coi := ledger.Instrument{ISIN: isinOf(t, "PLCOI1130004"), Symbol: "COI1130"}
	listed := ledger.Instrument{ISIN: isinOf(t, "PL9999000018"), Symbol: "EDO1136"}
	etf := ledger.Instrument{ISIN: isinOf(t, "DE0000000009"), Symbol: "SYNT"}
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	hold := func(in ledger.Instrument) ledger.Posting {
		return ledger.Posting{Book: ledger.HoldingBook(in), Amount: kernel.NewMoney(kernel.NewDecimal(100, 0), pln)}
	}
	txs := []ledger.Transaction{
		{Postings: []ledger.Posting{hold(edo), hold(listed), {Book: ledger.CashBook(pln)}}},
		{Postings: []ledger.Posting{hold(coi), hold(etf), hold(edo)}},
	}
	if diff := cmp.Diff([]ledger.Instrument{coi, edo}, ingest.ProvisionalInstruments(txs)); diff != "" {
		t.Errorf("ProvisionalInstruments (-want +got):\n%s", diff)
	}
}
