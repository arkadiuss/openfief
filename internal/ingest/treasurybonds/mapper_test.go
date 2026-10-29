package treasurybonds_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// account is the synthetic ledger account the lots go into.
const account ledger.AccountID = "acc-bonds"

// mustDate parses a date or fails the test.
func mustDate(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	return d
}

// pln returns amount in PLN or fails the test.
func pln(t *testing.T, amount string) kernel.Money {
	t.Helper()
	m, err := kernel.ParseMoney(amount, pl.Nominal().Currency())
	if err != nil {
		t.Fatalf("ParseMoney: %v", err)
	}
	return m
}

// lot returns a synthetic lot of quantity bonds of series bought on
// purchased, with a first-period rate of first, a 2% margin and a 2.00 PLN
// fee.
func lot(t *testing.T, series, quantity, purchased, first string) ingest.BondLot {
	t.Helper()
	s, err := pl.ParseSeries(series)
	if err != nil {
		t.Fatalf("ParseSeries: %v", err)
	}
	q, err := kernel.ParseQuantity(quantity)
	if err != nil {
		t.Fatalf("ParseQuantity: %v", err)
	}
	r, err := kernel.ParseDecimal(first)
	if err != nil {
		t.Fatalf("ParseDecimal: %v", err)
	}
	return ingest.BondLot{
		Terms:    pl.Terms{Series: s, FirstPeriodRate: r, Margin: kernel.NewDecimal(2, -2), EarlyRedemptionFee: pln(t, "2.00")},
		Quantity: q, Purchased: mustDate(t, purchased),
	}
}

// instrument returns the derived instrument of series or fails the test.
func instrument(t *testing.T, series string) ledger.Instrument {
	t.Helper()
	s, err := pl.ParseSeries(series)
	if err != nil {
		t.Fatalf("ParseSeries: %v", err)
	}
	in, err := treasurybonds.Instrument(s)
	if err != nil {
		t.Fatalf("Instrument: %v", err)
	}
	return in
}

func TestInstrument(t *testing.T) {
	t.Parallel()
	tests := []struct {
		series, wantISIN string
	}{
		{"EDO0536", "PLEDO0536000"},
		{"COI1030", "PLCOI1030006"},
		{"TOS1029", "PLTOS1029001"},
		{"ROR1127", "PLROR1127009"},
	}
	for _, tt := range tests {
		t.Run(tt.series, func(t *testing.T) {
			t.Parallel()
			in := instrument(t, tt.series)
			if in.ISIN.String() != tt.wantISIN || in.Symbol != tt.series {
				t.Errorf("Instrument(%s) = %s, want %s/%s", tt.series, in, tt.wantISIN, tt.series)
			}
			if _, err := ledger.ParseISIN(in.ISIN.String()); err != nil || in.Validate() != nil {
				t.Errorf("Instrument(%s) = %s is not a valid instrument: %v", tt.series, in, err)
			}
			if again := instrument(t, tt.series); again != in {
				t.Errorf("Instrument(%s) differs between calls: %s, %s", tt.series, in, again)
			}
		})
	}
}

func TestMap(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-10-29")
	lots := []ingest.BondLot{
		lot(t, "EDO0536", "40", "2026-05-04", "0.06"),
		lot(t, "EDO0536", "5", "2026-05-20", "0.06"),
		lot(t, "TOS1029", "12", "2026-10-01", "0.05"),
	}
	m, terms, err := treasurybonds.Mapper{Account: account}.Map(lots, today)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	edo, tos := instrument(t, "EDO0536"), instrument(t, "TOS1029")
	source := "treasury-bonds:acc-bonds"
	type row struct {
		Key         ledger.TransactionKey
		Kind        ledger.TransactionKind
		Date        string
		Description string
		Postings    []string
	}
	got := make([]row, 0, len(m.Transactions))
	for i := range m.Transactions {
		tx := &m.Transactions[i]
		if tx.Account != account {
			t.Errorf("%s is in account %s, want %s", tx.Key, tx.Account, account)
		}
		if err := tx.Validate(); err != nil {
			t.Errorf("%s: %v", tx.Key, err)
		}
		r := row{Key: tx.Key, Kind: tx.Kind, Date: tx.Date.String(), Description: tx.Description}
		for _, p := range tx.Postings {
			r.Postings = append(r.Postings, strings.TrimSpace(p.Book.String()+" "+p.Amount.String()+" "+qty(p.Quantity)))
		}
		got = append(got, r)
	}
	want := []row{
		{
			Key: ledger.TransactionKey{Source: source, ExternalID: "lot:EDO0536:2026-05-04:deposit"}, Kind: ledger.KindDeposit, Date: "2026-05-04",
			Description: "Cash for treasury bonds EDO0536", Postings: []string{"cash:PLN 4000 PLN", "equity:PLN -4000 PLN"},
		},
		{
			Key: ledger.TransactionKey{Source: source, ExternalID: "lot:EDO0536:2026-05-04:buy"}, Kind: ledger.KindBuy, Date: "2026-05-04",
			Description: "Treasury bonds EDO0536", Postings: []string{"holding:" + edo.String() + " 4000 PLN 40", "cash:PLN -4000 PLN"},
		},
		{
			Key: ledger.TransactionKey{Source: source, ExternalID: "lot:EDO0536:2026-05-20:deposit"}, Kind: ledger.KindDeposit, Date: "2026-05-20",
			Description: "Cash for treasury bonds EDO0536", Postings: []string{"cash:PLN 500 PLN", "equity:PLN -500 PLN"},
		},
		{
			Key: ledger.TransactionKey{Source: source, ExternalID: "lot:EDO0536:2026-05-20:buy"}, Kind: ledger.KindBuy, Date: "2026-05-20",
			Description: "Treasury bonds EDO0536", Postings: []string{"holding:" + edo.String() + " 500 PLN 5", "cash:PLN -500 PLN"},
		},
		{
			Key: ledger.TransactionKey{Source: source, ExternalID: "lot:TOS1029:2026-10-01:deposit"}, Kind: ledger.KindDeposit, Date: "2026-10-01",
			Description: "Cash for treasury bonds TOS1029", Postings: []string{"cash:PLN 1200 PLN", "equity:PLN -1200 PLN"},
		},
		{
			Key: ledger.TransactionKey{Source: source, ExternalID: "lot:TOS1029:2026-10-01:buy"}, Kind: ledger.KindBuy, Date: "2026-10-01",
			Description: "Treasury bonds TOS1029", Postings: []string{"holding:" + tos.String() + " 1200 PLN 12", "cash:PLN -1200 PLN"},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Map transactions (-want +got):\n%s", diff)
	}
	wantTerms := []ingest.BondTerms{{Instrument: edo, Terms: lots[0].Terms}, {Instrument: tos, Terms: lots[2].Terms}}
	if diff := cmp.Diff(wantTerms, terms); diff != "" {
		t.Errorf("Map terms, each series once (-want +got):\n%s", diff)
	}
	if len(m.Problems) != 0 || len(m.Assertions) != 0 {
		t.Errorf("Map = %d problems, %d assertions; want none", len(m.Problems), len(m.Assertions))
	}
}

// qty writes a holding posting's quantity, or nothing for any other.
func qty(q kernel.Quantity) string {
	if q.IsZero() {
		return ""
	}
	return q.String()
}

func TestMapKeepsKeysAcrossReEntry(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-10-29")
	mapper := treasurybonds.Mapper{Account: account}
	first, _, err := mapper.Map([]ingest.BondLot{lot(t, "EDO0536", "40", "2026-05-04", "0.06")}, today)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	fixed, _, err := mapper.Map([]ingest.BondLot{lot(t, "EDO0536", "45", "2026-05-04", "0.06")}, today)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	for i := range first.Transactions {
		if first.Transactions[i].Key != fixed.Transactions[i].Key {
			t.Errorf("re-entering the lot keyed %s as %s; want the same key, so it supersedes (D-18)",
				first.Transactions[i].Key, fixed.Transactions[i].Key)
		}
		if first.Transactions[i].Postings[0].Amount.Equal(fixed.Transactions[i].Postings[0].Amount) {
			t.Errorf("re-entering the lot with 45 bonds kept the amount of 40")
		}
	}
}

func TestMapRefuses(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-10-29")
	good := lot(t, "EDO0536", "40", "2026-05-04", "0.06")
	tests := []struct {
		name    string
		account ledger.AccountID
		lots    []ingest.BondLot
		wantErr error
	}{
		{name: "no account", lots: []ingest.BondLot{good}, wantErr: treasurybonds.ErrNoAccount},
		{name: "a lot under the wrong series", account: account, lots: []ingest.BondLot{good, lot(t, "EDO0536", "1", "2026-06-02", "0.06")}, wantErr: ingest.ErrInvalidBondLot},
		{name: "a lot bought after today", account: account, lots: []ingest.BondLot{lot(t, "EDO1136", "1", "2026-11-02", "0.06")}, wantErr: ingest.ErrInvalidBondLot},
		{name: "one series and day twice", account: account, lots: []ingest.BondLot{good, good}, wantErr: treasurybonds.ErrDuplicateLot},
		{name: "one series with two first-period rates", account: account, lots: []ingest.BondLot{good, lot(t, "EDO0536", "1", "2026-05-05", "0.07")}, wantErr: treasurybonds.ErrConflictingTerms},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, terms, err := treasurybonds.Mapper{Account: tt.account}.Map(tt.lots, today)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Map error (-want +got):\n%s", diff)
			}
			if len(m.Transactions) != 0 || len(terms) != 0 {
				t.Errorf("a refused batch mapped %d transactions and %d terms; want nothing", len(m.Transactions), len(terms))
			}
		})
	}
}
