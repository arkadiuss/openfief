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

// instrument returns the provisional instrument of series (D-214), what a
// lot is booked under while the published data lists no ISIN of it, or
// fails the test.
func instrument(t *testing.T, series string) ledger.Instrument {
	t.Helper()
	s, err := pl.ParseSeries(series)
	if err != nil {
		t.Fatalf("ParseSeries: %v", err)
	}
	in, err := treasurybonds.ProvisionalInstrument(s)
	if err != nil {
		t.Fatalf("ProvisionalInstrument: %v", err)
	}
	return in
}

func TestProvisionalInstrument(t *testing.T) {
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
				t.Errorf("ProvisionalInstrument(%s) = %s, want %s/%s", tt.series, in, tt.wantISIN, tt.series)
			}
			if _, err := ledger.ParseISIN(in.ISIN.String()); err != nil || in.Validate() != nil {
				t.Errorf("ProvisionalInstrument(%s) = %s is not a valid instrument: %v", tt.series, in, err)
			}
			if again := instrument(t, tt.series); again != in {
				t.Errorf("ProvisionalInstrument(%s) differs between calls: %s, %s", tt.series, in, again)
			}
			if !pl.IsProvisional(in.ISIN.String(), in.Symbol) {
				t.Errorf("ProvisionalInstrument(%s) = %s is not provisional", tt.series, in)
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
		Account: account, GoLive: mustDate(t, "2026-05-15"), Lots: lots,
		Deleted: []ingest.BondLotKey{key(t, "EDO0536", "2026-05-10"), key(t, "EDO0536", "2026-05-20")},
	}
	m, terms, err := treasurybonds.Mapper{}.Map(&e, nil, today)
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

	// D-98: without a go-live, as in every account not seeded with opening
	// lots, each lot is a deposit and a buy on its real purchase date, in
	// any order, and a deleted lot deletes both.
	e.GoLive = kernel.Date{}
	e.Lots, e.Deleted = []ingest.BondLot{lots[1], lots[0]}, []ingest.BondLotKey{key(t, "EDO0536", "2026-05-10")}
	onDates, _, err := treasurybonds.Mapper{}.Map(&e, nil, today)
	if err != nil {
		t.Fatalf("Map without go-live: %v", err)
	}
	wantReal := []mapRow{
		{
			Key: k("lot:EDO0536:2026-05-15:deposit"), Kind: ledger.KindDeposit, Date: "2026-05-15",
			Description: "Cash for treasury bonds EDO0536", Postings: []string{"cash:PLN 500 PLN", "equity:PLN -500 PLN"},
		},
		{
			Key: k("lot:EDO0536:2026-05-15:buy"), Kind: ledger.KindBuy, Date: "2026-05-15",
			Description: "Treasury bonds EDO0536", Postings: []string{"holding:" + edo.String() + " 500 PLN 5", "cash:PLN -500 PLN"},
		},
		{
			Key: k("lot:EDO0536:2026-05-04:deposit"), Kind: ledger.KindDeposit, Date: "2026-05-04",
			Description: "Cash for treasury bonds EDO0536", Postings: []string{"cash:PLN 3996 PLN", "equity:PLN -3996 PLN"},
		},
		{
			Key: k("lot:EDO0536:2026-05-04:buy"), Kind: ledger.KindBuy, Date: "2026-05-04",
			Description: "Treasury bonds EDO0536", Postings: []string{"holding:" + edo.String() + " 3996 PLN 40", "cash:PLN -3996 PLN"},
		},
		{Key: k("lot:EDO0536:2026-05-10:deposit"), Kind: ledger.KindDeletion, Date: "2026-05-10", Description: "Deleted treasury bonds EDO0536"},
		{Key: k("lot:EDO0536:2026-05-10:buy"), Kind: ledger.KindDeletion, Date: "2026-05-10", Description: "Deleted treasury bonds EDO0536"},
	}
	if diff := cmp.Diff(wantReal, rows(t, &onDates)); diff != "" {
		t.Errorf("Map without go-live (-want +got):\n%s", diff)
	}
	wantTerms := []ingest.BondTerms{{Instrument: edo, Terms: lots[0].Terms}, {Instrument: tos, Terms: lots[2].Terms}}
	if diff := cmp.Diff(wantTerms, terms); diff != "" {
		t.Errorf("Map terms, each series once (-want +got):\n%s", diff)
	}
	if len(m.Problems) != 0 || len(m.Assertions) != 0 {
		t.Errorf("Map = %d problems, %d assertions; want none", len(m.Problems), len(m.Assertions))
	}
}

func TestMapDeletesWhatWasBookedForTheLot(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-10-29")
	k := func(id string) ledger.TransactionKey {
		return ledger.TransactionKey{Source: "treasury-bonds:acc-bonds", ExternalID: id}
	}
	booked := func(source, id, day string) ledger.TransactionHeader {
		return ledger.TransactionHeader{Key: ledger.TransactionKey{Source: source, ExternalID: id}, Account: account, Date: mustDate(t, day)}
	}
	own := "treasury-bonds:acc-bonds"
	e := ingest.BondLotEntries{
		Account: account, Deleted: []ingest.BondLotKey{key(t, "COI1028", "2024-10-10")},
		Booked: []ledger.TransactionHeader{
			booked(own, "lot:COI1028:2024-10-10:deposit", "2024-10-10"),
			booked(own, "lot:COI1028:2024-10-10:buy", "2024-10-10"),
			booked(own, "lot:COI1028:2024-10-10:coupon:2025-10-10", "2025-10-10"),
			booked(own, "lot:COI1028:2024-10-10:coupon:2025-10-10:payout", "2025-10-10"),
			booked(own, "lot:COI1028:2024-10-11:coupon:2025-10-11", "2025-10-11"),                        // another lot of the series
			booked("treasury-bonds:acc-other", "lot:COI1028:2024-10-10:coupon:2025-10-10", "2025-10-10"), // another account's source
			booked(own, "lot:COI1028:2024-10-10:coupon:2026-10-10", "2026-10-10"),
			booked(own, "lot:COI1028:2024-10-10:redemption", "2028-10-10"),
		},
	}
	m, _, err := treasurybonds.Mapper{}.Map(&e, nil, today)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	deleted := func(id, day string) mapRow {
		return mapRow{Key: k(id), Kind: ledger.KindDeletion, Date: day, Description: "Deleted treasury bonds COI1028"}
	}
	// D-154: the lot's deposit and buy, then its coupons, their payout withdrawal and its redemption, each as dated
	// when booked; nothing of another lot or another account.
	want := []mapRow{
		deleted("lot:COI1028:2024-10-10:deposit", "2024-10-10"),
		deleted("lot:COI1028:2024-10-10:buy", "2024-10-10"),
		deleted("lot:COI1028:2024-10-10:coupon:2025-10-10", "2025-10-10"),
		deleted("lot:COI1028:2024-10-10:coupon:2025-10-10:payout", "2025-10-10"),
		deleted("lot:COI1028:2024-10-10:coupon:2026-10-10", "2026-10-10"),
		deleted("lot:COI1028:2024-10-10:redemption", "2028-10-10"),
	}
	if diff := cmp.Diff(want, rows(t, &m)); diff != "" {
		t.Errorf("Map deletions (-want +got):\n%s", diff)
	}

	// An opening lot in an account seeded before D-98 was rewritten: its own key and its coupons after go-live.
	e.GoLive = mustDate(t, "2025-01-02")
	e.Booked = []ledger.TransactionHeader{
		booked(own, "lot:COI1028:2024-10-10", "2025-01-02"),
		booked(own, "lot:COI1028:2024-10-10:coupon:2025-10-10", "2025-10-10"),
	}
	opening, _, err := treasurybonds.Mapper{}.Map(&e, nil, today)
	if err != nil {
		t.Fatalf("Map with go-live: %v", err)
	}
	wantOpening := []mapRow{
		deleted("lot:COI1028:2024-10-10", "2025-01-02"),
		deleted("lot:COI1028:2024-10-10:coupon:2025-10-10", "2025-10-10"),
	}
	if diff := cmp.Diff(wantOpening, rows(t, &opening)); diff != "" {
		t.Errorf("Map deletions with go-live (-want +got):\n%s", diff)
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
	mapper := treasurybonds.Mapper{}
	for _, goLive := range []string{"", "2026-05-04", "2026-05-03"} {
		var live kernel.Date
		if goLive != "" {
			live = mustDate(t, goLive)
		}
		first, _, err := mapper.Map(&ingest.BondLotEntries{Account: account, GoLive: live, Lots: []ingest.BondLot{lot(t, "EDO0536", "40", "2026-05-04", "0.06")}}, nil, today)
		if err != nil {
			t.Fatalf("Map: %v", err)
		}
		fixed, _, err := mapper.Map(&ingest.BondLotEntries{Account: account, GoLive: live, Lots: []ingest.BondLot{lot(t, "EDO0536", "45", "2026-05-04", "0.06")}}, nil, today)
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

// TestMapPendingLot books a lot still waiting for its series' terms, as
// one of a series the ministry's data doesn't list yet, and delivers no
// terms for it, while another lot's terms go out (D-97).
func TestMapPendingLot(t *testing.T) {
	t.Parallel()
	pending := lot(t, "COI1130", "10", "2026-11-03", "0.05")
	pending.PublishedRates, pending.ReviewedFee = true, true
	entered := lot(t, "EDO1136", "5", "2026-11-03", "0.06")
	m, terms, err := treasurybonds.Mapper{}.Map(&ingest.BondLotEntries{Account: account, Lots: []ingest.BondLot{pending, entered}}, nil, mustDate(t, "2026-11-07"))
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if len(m.Transactions) != 4 {
		t.Errorf("Map booked %d transactions, want a deposit and a buy for each lot", len(m.Transactions))
	}
	if len(terms) != 1 || terms[0].Terms.Series != entered.Series() {
		t.Errorf("Map terms = %+v, want only EDO1136's", terms)
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
			e := tt.entries
			e.Account = tt.account
			m, terms, err := treasurybonds.Mapper{}.Map(&e, nil, today)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Map error (-want +got):\n%s", diff)
			}
			if len(m.Transactions) != 0 || len(terms) != 0 {
				t.Errorf("a refused batch mapped %d transactions and %d terms; want nothing", len(m.Transactions), len(terms))
			}
		})
	}
}

// listing returns synthetic published data that lists the registry ISIN
// of each "series=ISIN" pair, and no rates.
func listing(t *testing.T, pairs ...string) *pl.Published {
	t.Helper()
	p := &pl.Published{}
	for _, pair := range pairs {
		code, isin, _ := strings.Cut(pair, "=")
		s, err := pl.ParseSeries(code)
		if err != nil {
			t.Fatalf("ParseSeries: %v", err)
		}
		p.ISINs = append(p.ISINs, pl.PublishedISIN{Series: s, ISIN: isin})
	}
	return p
}

// registry returns the instrument of series with the synthetic registry
// ISIN isin, or fails the test.
func registry(t *testing.T, series, isin string) ledger.Instrument {
	t.Helper()
	i, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	return ledger.Instrument{ISIN: i, Symbol: series}
}

// TestSeriesInstrument: a series is its registry ISIN as the published
// data lists it, and its provisional instrument while the data lists none
// or wasn't read (D-214).
func TestSeriesInstrument(t *testing.T) {
	t.Parallel()
	p := listing(t, "EDO0536=PL9999000018")
	tests := []struct {
		name, series string
		p            *pl.Published
		want         ledger.Instrument
	}{
		{name: "listed: the registry's ISIN", series: "EDO0536", p: p, want: registry(t, "EDO0536", "PL9999000018")},
		{name: "not listed yet: provisional", series: "EDO1136", p: p, want: instrument(t, "EDO1136")},
		{name: "data not read: provisional", series: "EDO0536", want: instrument(t, "EDO0536")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := treasurybonds.SeriesInstrument(tt.p, series(t, tt.series))
			if err != nil {
				t.Fatalf("SeriesInstrument: %v", err)
			}
			if got != tt.want {
				t.Errorf("SeriesInstrument(%s) = %s, want %s", tt.series, got, tt.want)
			}
		})
	}
}

// TestMapRegistryISIN books a lot of a series the published data lists
// under its registry ISIN, with its terms, and one of a series it doesn't
// list yet under its provisional ISIN; the keys name the series either
// way, so a later re-key keeps them (D-214).
func TestMapRegistryISIN(t *testing.T) {
	t.Parallel()
	e := ingest.BondLotEntries{Account: account, Lots: []ingest.BondLot{
		lot(t, "EDO0536", "10", "2026-05-04", "0.06"), lot(t, "EDO1136", "3", "2026-11-02", "0.06"),
	}}
	m, terms, err := treasurybonds.Mapper{}.Map(&e, listing(t, "EDO0536=PL9999000018"), mustDate(t, "2026-11-07"))
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	edo, waiting := registry(t, "EDO0536", "PL9999000018"), instrument(t, "EDO1136")
	got := map[string]string{}
	for _, r := range rows(t, &m) {
		got[r.Key.ExternalID] = strings.Join(r.Postings, ", ")
	}
	want := map[string]string{
		"lot:EDO0536:2026-05-04:deposit": "cash:PLN 1000 PLN, equity:PLN -1000 PLN",
		"lot:EDO0536:2026-05-04:buy":     "holding:" + edo.String() + " 1000 PLN 10, cash:PLN -1000 PLN",
		"lot:EDO1136:2026-11-02:deposit": "cash:PLN 300 PLN, equity:PLN -300 PLN",
		"lot:EDO1136:2026-11-02:buy":     "holding:" + waiting.String() + " 300 PLN 3, cash:PLN -300 PLN",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Map postings (-want +got):\n%s", diff)
	}
	if len(terms) != 2 || terms[0].Instrument != edo || terms[1].Instrument != waiting {
		t.Errorf("Map terms = %v, want the terms of %s and %s", terms, edo, waiting)
	}
}
