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

// key returns the key of the lot of series bought on purchased.
func key(t *testing.T, series, purchased string) ingest.BondLotKey {
	t.Helper()
	s, err := pl.ParseSeries(series)
	if err != nil {
		t.Fatalf("ParseSeries: %v", err)
	}
	return ingest.BondLotKey{Series: s, Purchased: mustDate(t, purchased)}
}

// mapRow is one mapped transaction in comparable form.
type mapRow struct {
	Key         ledger.TransactionKey
	Kind        ledger.TransactionKind
	Date        string
	LotOpened   string
	Description string
	Postings    []string
}

// rows returns m's transactions as rows, checking each is valid and in
// account.
func rows(t *testing.T, m *ingest.Mapping) []mapRow {
	t.Helper()
	got := make([]mapRow, 0, len(m.Transactions))
	for i := range m.Transactions {
		tx := &m.Transactions[i]
		if tx.Account != account {
			t.Errorf("%s is in account %s, want %s", tx.Key, tx.Account, account)
		}
		if err := tx.Validate(); err != nil {
			t.Errorf("%s: %v", tx.Key, err)
		}
		r := mapRow{Key: tx.Key, Kind: tx.Kind, Date: tx.Date.String(), Description: tx.Description}
		if !tx.LotOpened.IsZero() {
			r.LotOpened = tx.LotOpened.String()
		}
		for _, p := range tx.Postings {
			r.Postings = append(r.Postings, strings.TrimSpace(p.Book.String()+" "+p.Amount.String()+" "+qty(p.Quantity)))
		}
		got = append(got, r)
	}
	return got
}

func TestMap(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-10-29")
	lots := []ingest.BondLot{
		lot(t, "EDO0536", "40", "2026-05-04", "0.06"),
		lot(t, "EDO0536", "5", "2026-05-15", "0.06"),
		lot(t, "TOS1029", "12", "2026-10-01", "0.05"),
	}
	lots[0].Purchase = ingest.PurchaseExchange // bought by exchange of maturing bonds (D-100)
	e := ingest.BondLotEntries{
		GoLive: mustDate(t, "2026-05-15"), Lots: lots,
		Deleted: []ingest.BondLotKey{key(t, "EDO0536", "2026-05-10"), key(t, "EDO0536", "2026-05-20")},
	}
	m, terms, err := treasurybonds.Mapper{Account: account}.Map(&e, today)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	edo, tos := instrument(t, "EDO0536"), instrument(t, "TOS1029")
	k := func(id string) ledger.TransactionKey {
		return ledger.TransactionKey{Source: "treasury-bonds:acc-bonds", ExternalID: id}
	}
	want := []mapRow{
		// Bought before go-live, by exchange at 99.90 PLN a bond: an opening lot on go-live.
		{
			Key: k("lot:EDO0536:2026-05-04"), Kind: ledger.KindOpeningLot, Date: "2026-05-15", LotOpened: "2026-05-04",
			Description: "Treasury bonds EDO0536 bought before go-live", Postings: []string{"holding:" + edo.String() + " 3996 PLN 40", "equity:PLN -3996 PLN"},
		},
		// Bought on the go-live date: an opening lot too.
		{
			Key: k("lot:EDO0536:2026-05-15"), Kind: ledger.KindOpeningLot, Date: "2026-05-15", LotOpened: "2026-05-15",
			Description: "Treasury bonds EDO0536 bought before go-live", Postings: []string{"holding:" + edo.String() + " 500 PLN 5", "equity:PLN -500 PLN"},
		},
		// Bought after go-live, for cash: a deposit and a buy on its purchase date.
		{
			Key: k("lot:TOS1029:2026-10-01:deposit"), Kind: ledger.KindDeposit, Date: "2026-10-01",
			Description: "Cash for treasury bonds TOS1029", Postings: []string{"cash:PLN 1200 PLN", "equity:PLN -1200 PLN"},
		},
		{
			Key: k("lot:TOS1029:2026-10-01:buy"), Kind: ledger.KindBuy, Date: "2026-10-01",
			Description: "Treasury bonds TOS1029", Postings: []string{"holding:" + tos.String() + " 1200 PLN 12", "cash:PLN -1200 PLN"},
		},
		// Deleted lots: each transaction the lot was booked as, keyed and dated alike.
		{Key: k("lot:EDO0536:2026-05-10"), Kind: ledger.KindDeletion, Date: "2026-05-15", Description: "Deleted treasury bonds EDO0536"},
		{Key: k("lot:EDO0536:2026-05-20:deposit"), Kind: ledger.KindDeletion, Date: "2026-05-20", Description: "Deleted treasury bonds EDO0536"},
		{Key: k("lot:EDO0536:2026-05-20:buy"), Kind: ledger.KindDeletion, Date: "2026-05-20", Description: "Deleted treasury bonds EDO0536"},
	}
	if diff := cmp.Diff(want, rows(t, &m)); diff != "" {
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
	for _, goLive := range []string{"2026-05-04", "2026-05-03"} {
		first, _, err := mapper.Map(&ingest.BondLotEntries{GoLive: mustDate(t, goLive), Lots: []ingest.BondLot{lot(t, "EDO0536", "40", "2026-05-04", "0.06")}}, today)
		if err != nil {
			t.Fatalf("Map: %v", err)
		}
		fixed, _, err := mapper.Map(&ingest.BondLotEntries{GoLive: mustDate(t, goLive), Lots: []ingest.BondLot{lot(t, "EDO0536", "45", "2026-05-04", "0.06")}}, today)
		if err != nil {
			t.Fatalf("Map: %v", err)
		}
		for i := range first.Transactions {
			if first.Transactions[i].Key != fixed.Transactions[i].Key {
				t.Errorf("go-live %s: re-entering the lot keyed %s as %s; want the same key, so it supersedes (D-18)",
					goLive, first.Transactions[i].Key, fixed.Transactions[i].Key)
			}
			if first.Transactions[i].Postings[0].Amount.Equal(fixed.Transactions[i].Postings[0].Amount) {
				t.Errorf("go-live %s: re-entering the lot with 45 bonds kept the amount of 40", goLive)
			}
		}
	}
}

func TestMapRefuses(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-10-29")
	good := lot(t, "EDO0536", "40", "2026-05-04", "0.06")
	unknown := good
	unknown.Purchase = ingest.BondPurchase(7)
	goLive := mustDate(t, "2026-06-01")
	tests := []struct {
		name    string
		account ledger.AccountID
		entries ingest.BondLotEntries
		wantErr error
	}{
		{name: "no account", entries: ingest.BondLotEntries{GoLive: goLive, Lots: []ingest.BondLot{good}}, wantErr: treasurybonds.ErrNoAccount},
		{name: "no go-live", account: account, entries: ingest.BondLotEntries{Lots: []ingest.BondLot{good}}, wantErr: treasurybonds.ErrNoGoLive},
		{name: "a lot under the wrong series", account: account, entries: ingest.BondLotEntries{GoLive: goLive, Lots: []ingest.BondLot{good, lot(t, "EDO0536", "1", "2026-06-02", "0.06")}}, wantErr: ingest.ErrInvalidBondLot},
		{name: "a lot bought after today", account: account, entries: ingest.BondLotEntries{GoLive: goLive, Lots: []ingest.BondLot{lot(t, "EDO1136", "1", "2026-11-02", "0.06")}}, wantErr: ingest.ErrInvalidBondLot},
		{name: "a purchase neither cash nor exchange", account: account, entries: ingest.BondLotEntries{GoLive: goLive, Lots: []ingest.BondLot{unknown}}, wantErr: ingest.ErrInvalidBondLot},
		{name: "one series and day twice", account: account, entries: ingest.BondLotEntries{GoLive: goLive, Lots: []ingest.BondLot{good, good}}, wantErr: treasurybonds.ErrDuplicateLot},
		{name: "a lot added and deleted at once", account: account, entries: ingest.BondLotEntries{GoLive: goLive, Lots: []ingest.BondLot{good}, Deleted: []ingest.BondLotKey{good.Key()}}, wantErr: treasurybonds.ErrDuplicateLot},
		{name: "a deleted lot without a date", account: account, entries: ingest.BondLotEntries{GoLive: goLive, Deleted: []ingest.BondLotKey{{Series: good.Series()}}}, wantErr: ingest.ErrInvalidBondLot},
		{name: "one series with two first-period rates", account: account, entries: ingest.BondLotEntries{GoLive: goLive, Lots: []ingest.BondLot{good, lot(t, "EDO0536", "1", "2026-05-05", "0.07")}}, wantErr: treasurybonds.ErrConflictingTerms},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, terms, err := treasurybonds.Mapper{Account: tt.account}.Map(&tt.entries, today)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Map error (-want +got):\n%s", diff)
			}
			if len(m.Transactions) != 0 || len(terms) != 0 {
				t.Errorf("a refused batch mapped %d transactions and %d terms; want nothing", len(m.Transactions), len(terms))
			}
		})
	}
}
