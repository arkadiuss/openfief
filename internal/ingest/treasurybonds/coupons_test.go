package treasurybonds_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// heldLot is an open lot of quantity bonds of series bought on purchased,
// for 100 PLN a bond.
func heldLot(t *testing.T, series, quantity, purchased string) ledger.Lot {
	t.Helper()
	l := lot(t, series, quantity, purchased, "0.06")
	return ledger.Lot{Instrument: instrument(t, series), Opened: l.Purchased, Quantity: l.Quantity, Cost: l.Cost(), Opening: true}
}

// termsOf returns the series terms lot uses: first-period rate first, a
// 2% margin, a 2.00 PLN fee.
func termsOf(t *testing.T, series, first string) ingest.BondTerms {
	t.Helper()
	l := lot(t, series, "1", "2000-01-01", first)
	return ingest.BondTerms{Instrument: instrument(t, series), Terms: l.Terms}
}

// coupons names each due coupon as "account series due period".
func coupons(due []treasurybonds.DueCoupon) []string {
	out := make([]string, 0, len(due))
	for i := range due {
		c := &due[i]
		out = append(out, string(c.Account)+" "+c.Terms.Series.String()+" "+c.Due.String()+" "+strconv.Itoa(c.Period))
	}
	return out
}

// problemRefs names problems by their ref.
func problemRefs(problems []ingest.Problem) []string {
	out := make([]string, 0, len(problems))
	for _, p := range problems {
		out = append(out, p.Ref)
	}
	return out
}

func TestDueCoupons(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-10-29")
	coi := termsOf(t, "COI1028", "0.0655")
	tests := []struct {
		name         string
		account      ingest.HeldAccount
		terms        []ingest.BondTerms
		want         []string
		wantProblems []string
	}{
		{
			name:    "coupons after go-live up to today, none on go-live itself",
			account: ingest.HeldAccount{Account: "a", GoLive: mustDate(t, "2025-10-10"), Lots: []ledger.Lot{heldLot(t, "COI1028", "10", "2024-10-10")}},
			terms:   []ingest.BondTerms{coi}, want: []string{"a COI1028 2026-10-10 2"},
		},
		{
			name:    "the coupon falling due today counts",
			account: ingest.HeldAccount{Account: "a", GoLive: mustDate(t, "2025-01-01"), Lots: []ledger.Lot{heldLot(t, "COI1029", "1", "2025-10-29")}},
			terms:   []ingest.BondTerms{termsOf(t, "COI1029", "0.06")}, want: []string{"a COI1029 2026-10-29 1"},
		},
		{
			name:    "an account not seeded holds its whole history: every coupon since purchase (D-98)",
			account: ingest.HeldAccount{Account: "a", Lots: []ledger.Lot{heldLot(t, "COI1028", "10", "2024-10-10")}},
			terms:   []ingest.BondTerms{coi}, want: []string{"a COI1028 2025-10-10 1", "a COI1028 2026-10-10 2"},
		},
		{
			name:    "capitalized bonds and instruments that are no bond pay none",
			account: ingest.HeldAccount{Account: "a", GoLive: mustDate(t, "2024-01-01"), Lots: []ledger.Lot{heldLot(t, "EDO0535", "3", "2025-05-05"), {Instrument: ledger.Instrument{Symbol: "SPY"}}}},
			terms:   []ingest.BondTerms{termsOf(t, "EDO0535", "0.0625")},
		},
		{
			name:         "a bond without terms is a problem",
			account:      ingest.HeldAccount{Account: "a", GoLive: mustDate(t, "2024-01-01"), Lots: []ledger.Lot{heldLot(t, "COI1028", "10", "2024-10-10")}},
			wantProblems: []string{"a COI1028 bought 2024-10-10"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			due, problems := treasurybonds.DueCoupons([]ingest.HeldAccount{tt.account}, tt.terms, today)
			if diff := cmp.Diff(tt.want, coupons(due), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("DueCoupons (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantProblems, problemRefs(problems), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("DueCoupons problems (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCouponSyncCoupons(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-10-29")
	accounts := []ingest.HeldAccount{
		{
			Account: "acc-bonds", GoLive: mustDate(t, "2024-01-01"), Wrapper: "taxable",
			Lots: []ledger.Lot{heldLot(t, "COI1028", "10", "2024-10-10"), heldLot(t, "COI1126", "2", "2022-11-10")},
		},
		{Account: "acc-odd", GoLive: mustDate(t, "2024-01-01"), Wrapper: "pension", Lots: []ledger.Lot{heldLot(t, "COI1028", "1", "2024-10-10")}},
	}
	terms := []ingest.BondTerms{termsOf(t, "COI1028", "0.0655"), termsOf(t, "COI1126", "0.07")}
	cpi := &fakeCPI{}
	m, err := (&treasurybonds.CouponSync{CPI: cpi, Rates: &fakeRates{}}).Coupons(t.Context(), accounts, terms, today)
	if err != nil {
		t.Fatalf("Coupons: %v", err)
	}
	type row struct{ Key, Date, Gross string }
	got := make([]row, 0, len(m.Transactions))
	for i := range m.Transactions {
		tx := &m.Transactions[i]
		if tx.Kind != ledger.KindInterest {
			t.Errorf("%s is %v, want interest (D-101)", tx.Key, tx.Kind)
		}
		got = append(got, row{tx.Key.String(), tx.Date.String(), tx.Postings[0].Amount.Neg().Amount().String()})
	}
	src := "treasury-bonds:acc-bonds/"
	want := []row{
		// COI1028 years 1 and 2: 6.55% then CPI 3% + 2% margin, ten bonds.
		{src + "lot:COI1028:2024-10-10:coupon:2025-10-10", "2025-10-10", "65.5"},
		{src + "lot:COI1028:2024-10-10:coupon:2026-10-10", "2026-10-10", "50"},
		// COI1126 bought 2022-11-10: years 2 and 3 fell due after go-live, two bonds at 3% + 2%.
		{src + "lot:COI1126:2022-11-10:coupon:2024-11-10", "2024-11-10", "10"},
		{src + "lot:COI1126:2022-11-10:coupon:2025-11-10", "2025-11-10", "10"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Coupons (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]int{2023, 2024, 2025}, cpi.years); diff != "" {
		t.Errorf("CPI years asked (-want +got):\n%s", diff)
	}
	wantProblems := []string{
		"acc-odd COI1028 bought 2024-10-10 coupon 2025-10-10", "acc-odd COI1028 bought 2024-10-10 coupon 2026-10-10",
	}
	if diff := cmp.Diff(wantProblems, problemRefs(m.Problems)); diff != "" {
		t.Errorf("Coupons problems (-want +got):\n%s", diff)
	}
}

func TestAccountsBookedIn(t *testing.T) {
	t.Parallel()
	accounts := []ingest.HeldAccount{{Account: "seeded", GoLive: mustDate(t, "2025-01-01")}, {Account: "real-dates"}}
	tests := []struct {
		mode ingest.Mode
		want []string
	}{
		{mode: ingest.ModeTransactions, want: []string{"seeded"}},
		{mode: ingest.ModeOpening, want: []string{"real-dates"}},
	}
	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			t.Parallel()
			var got []string
			for _, a := range treasurybonds.AccountsBookedIn(accounts, tt.mode) {
				got = append(got, string(a.Account))
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("AccountsBookedIn (-want +got):\n%s", diff)
			}
		})
	}
}

// errGUSDown is a failed request to GUS.
var errGUSDown = errors.New("gus down")

func TestCouponSyncWaitsAndFails(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-10-29")
	// COI1027 bought 2023-10-10 needs CPI of August 2024 and 2025; the fake has none after September 2026.
	held := []ingest.HeldAccount{{Account: "a", GoLive: mustDate(t, "2023-12-01"), Wrapper: "ike", Lots: []ledger.Lot{heldLot(t, "COI1027", "1", "2023-10-10")}}}
	terms := []ingest.BondTerms{termsOf(t, "COI1027", "0.07")}

	gusDown := errGUSDown
	if _, err := (&treasurybonds.CouponSync{CPI: &fakeCPI{err: gusDown}, Rates: &fakeRates{}}).Coupons(t.Context(), held, terms, today); !errors.Is(err, gusDown) {
		t.Errorf("Coupons with GUS down = %v, want %v", err, gusDown)
	}

	// No readings for the period: the coupon waits as a problem and nothing is booked.
	m, err := treasurybonds.BookCoupons([]treasurybonds.DueCoupon{{
		Account: "a", Wrapper: "ike", Lot: held[0].Lots[0], Terms: terms[0].Terms, Period: 2,
		Due: mustDate(t, "2025-10-10"), Purchased: mustDate(t, "2023-10-10"),
	}}, nil)
	if err != nil {
		t.Fatalf("BookCoupons: %v", err)
	}
	if len(m.Transactions) != 0 || len(m.Problems) != 1 {
		t.Errorf("BookCoupons without readings = %d transactions, %v; want one problem", len(m.Transactions), m.Problems)
	}
}

func TestCouponSyncKeepsCPIYears(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-10-29")
	// Five COI lots bought a year apart need more years of CPI than one sync asks GUS for.
	many := make([]ingest.HeldAccount, 0, 5)
	manyTerms := make([]ingest.BondTerms, 0, 5)
	for _, s := range []struct{ series, bought string }{
		{"COI0124", "2020-01-10"}, {"COI0125", "2021-01-10"}, {"COI0126", "2022-01-10"}, {"COI0127", "2023-01-10"}, {"COI0128", "2024-01-10"},
	} {
		many = append(many, ingest.HeldAccount{Account: ledger.AccountID("a-" + s.series), GoLive: mustDate(t, "2019-01-01"), Wrapper: "ike", Lots: []ledger.Lot{heldLot(t, s.series, "1", s.bought)}})
		manyTerms = append(manyTerms, termsOf(t, s.series, "0.05"))
	}
	cpi := &fakeCPI{}
	sync := &treasurybonds.CouponSync{CPI: cpi, Rates: &fakeRates{}}
	type syncResult struct {
		Asked    []int
		Booked   int
		Problems []string
	}
	got := make([]syncResult, 0, 3)
	for range 3 {
		cpi.years = nil
		m, err := sync.Coupons(t.Context(), many, manyTerms, today)
		if err != nil {
			t.Fatalf("Coupons: %v", err)
		}
		got = append(got, syncResult{Asked: cpi.years, Booked: len(m.Transactions), Problems: problemRefs(m.Problems)})
	}
	// The first sync asks for four years and books what they pay; the coupons and maturities of later years wait.
	// The second asks only for the years it lacks, so every coupon and maturity since purchase is booked (D-143),
	// and the third asks GUS for nothing it was served in full.
	want := []syncResult{
		{Asked: []int{2020, 2021, 2022, 2023}, Booked: 14, Problems: []string{
			"a-COI0127 COI0127 bought 2023-01-10 coupon 2026-01-10",
			"a-COI0128 COI0128 bought 2024-01-10 coupon 2026-01-10",
			"a-COI0126 COI0126 bought 2022-01-10 maturity 2026-01-10",
		}},
		{Asked: []int{2024}, Booked: 17},
		{Booked: 17},
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("three syncs (-want +got):\n%s", diff)
	}
}
