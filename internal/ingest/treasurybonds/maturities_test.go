package treasurybonds_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// maturities names each matured lot as "account series maturity".
func maturities(due []treasurybonds.DueMaturity) []string {
	out := make([]string, 0, len(due))
	for i := range due {
		m := &due[i]
		out = append(out, string(m.Account)+" "+m.Terms.Series.String()+" "+m.Due.String())
	}
	return out
}

func TestDueMaturities(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-10-29")
	ots := termsOf(t, "OTS0426", "0.03")
	tests := []struct {
		name    string
		account ingest.HeldAccount
		terms   []ingest.BondTerms
		want    []string
	}{
		{
			name:    "a lot that matured after go-live and by today",
			account: ingest.HeldAccount{Account: "a", GoLive: mustDate(t, "2026-02-01"), Lots: []ledger.Lot{heldLot(t, "OTS0426", "10", "2026-01-05")}},
			terms:   []ingest.BondTerms{ots}, want: []string{"a OTS0426 2026-04-05"},
		},
		{
			name:    "a lot maturing today counts",
			account: ingest.HeldAccount{Account: "a", GoLive: mustDate(t, "2026-08-01"), Lots: []ledger.Lot{heldLot(t, "OTS1026", "1", "2026-07-29")}},
			terms:   []ingest.BondTerms{termsOf(t, "OTS1026", "0.03")}, want: []string{"a OTS1026 2026-10-29"},
		},
		{
			name:    "a lot not matured yet",
			account: ingest.HeldAccount{Account: "a", GoLive: mustDate(t, "2026-02-01"), Lots: []ledger.Lot{heldLot(t, "COI1028", "1", "2024-10-10")}},
			terms:   []ingest.BondTerms{termsOf(t, "COI1028", "0.0655")},
		},
		{
			name:    "a lot that matured on go-live is history",
			account: ingest.HeldAccount{Account: "a", GoLive: mustDate(t, "2026-04-05"), Lots: []ledger.Lot{heldLot(t, "OTS0426", "10", "2026-01-05")}},
			terms:   []ingest.BondTerms{ots},
		},
		{
			name:    "an account not seeded holds its whole history (D-98)",
			account: ingest.HeldAccount{Account: "a", Lots: []ledger.Lot{heldLot(t, "OTS0426", "10", "2026-01-05")}},
			terms:   []ingest.BondTerms{ots}, want: []string{"a OTS0426 2026-04-05"},
		},
		{
			name:    "a lot without terms or that is no bond is left to DueCoupons",
			account: ingest.HeldAccount{Account: "a", GoLive: mustDate(t, "2026-02-01"), Lots: []ledger.Lot{heldLot(t, "OTS0426", "10", "2026-01-05"), {Instrument: ledger.Instrument{Symbol: "SPY"}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := treasurybonds.DueMaturities([]ingest.HeldAccount{tt.account}, tt.terms, today)
			if diff := cmp.Diff(tt.want, maturities(got), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("DueMaturities (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBookMaturities(t *testing.T) {
	t.Parallel()
	ots, tos := termsOf(t, "OTS0426", "0.03"), termsOf(t, "TOS0526", "0.05")
	exchanged := heldLot(t, "TOS0526", "10", "2023-05-10")
	exchanged.Cost = pln(t, "999.00")
	paidOut := ledger.PayoutSetting{Payout: ledger.PayoutPaidOut, Since: mustDate(t, "2026-01-01")}
	src := "treasury-bonds:" + string(account)
	key := func(id string) ledger.TransactionKey { return ledger.TransactionKey{Source: src, ExternalID: id} }
	tests := []struct {
		name string
		due  treasurybonds.DueMaturity
		want []mapRow
	}{
		{
			name: "redeemed at nominal plus interest, Belka tax withheld, kept as cash",
			due: treasurybonds.DueMaturity{
				Account: account, Wrapper: "taxable", Lot: heldLot(t, "OTS0426", "10", "2026-01-05"), Terms: ots.Terms, Due: mustDate(t, "2026-04-05"),
			},
			// 3% for three months is 0.75 PLN a bond; 19% of 7.50 PLN is 1.425, half up 1.43.
			want: []mapRow{{
				Key: key("lot:OTS0426:2026-01-05:redemption"), Kind: ledger.KindRedemption, Date: "2026-04-05",
				Description: "Treasury bonds OTS0426 redeemed at maturity",
				Postings:    []string{"holding:PLOTS042600" + checkDigit(t, "OTS0426") + "/OTS0426 -1007.5 PLN -10", "withholding-tax:PLN 1.43 PLN", "cash:PLN 1006.07 PLN"},
			}},
		},
		{
			name: "bought by exchange: the discount is taxed with the interest, then paid out",
			due: treasurybonds.DueMaturity{
				Account: account, Wrapper: "taxable", Payout: paidOut, Lot: exchanged, Terms: tos.Terms, Due: mustDate(t, "2026-05-10"),
			},
			// 5% a year capitalized for three years: 115.76 PLN a bond; 19% of 157.60 interest plus 1.00 discount.
			want: []mapRow{
				{
					Key: key("lot:TOS0526:2023-05-10:redemption"), Kind: ledger.KindRedemption, Date: "2026-05-10",
					Description: "Treasury bonds TOS0526 redeemed at maturity",
					Postings:    []string{"holding:PLTOS052600" + checkDigit(t, "TOS0526") + "/TOS0526 -1157.6 PLN -10", "withholding-tax:PLN 30.13 PLN", "cash:PLN 1127.47 PLN"},
				},
				{
					Key: key("lot:TOS0526:2023-05-10:redemption:payout"), Kind: ledger.KindWithdrawal, Date: "2026-05-10",
					Description: "Paid out: Treasury bonds TOS0526 redeemed at maturity",
					Postings:    []string{"cash:PLN -1127.47 PLN", "equity:PLN 1127.47 PLN"},
				},
			},
		},
		{
			name: "inside IKE nothing is withheld",
			due: treasurybonds.DueMaturity{
				Account: account, Wrapper: "ike", Lot: heldLot(t, "OTS0426", "10", "2026-01-05"), Terms: ots.Terms, Due: mustDate(t, "2026-04-05"),
			},
			want: []mapRow{{
				Key: key("lot:OTS0426:2026-01-05:redemption"), Kind: ledger.KindRedemption, Date: "2026-04-05",
				Description: "Treasury bonds OTS0426 redeemed at maturity",
				Postings:    []string{"holding:PLOTS042600" + checkDigit(t, "OTS0426") + "/OTS0426 -1007.5 PLN -10", "cash:PLN 1007.5 PLN"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := treasurybonds.BookMaturities([]treasurybonds.DueMaturity{tt.due}, nil)
			if err != nil {
				t.Fatalf("BookMaturities: %v", err)
			}
			if diff := cmp.Diff(tt.want, rows(t, &m)); diff != "" {
				t.Errorf("BookMaturities (-want +got):\n%s", diff)
			}
		})
	}
}

// checkDigit returns the last character of the ISIN derived from series.
func checkDigit(t *testing.T, series string) string {
	t.Helper()
	isin := instrument(t, series).ISIN.String()
	return isin[len(isin)-1:]
}

func TestBookMaturitiesWaits(t *testing.T) {
	t.Parallel()
	// EDO0535 bought 2025-05-05 needs CPI for years 2..10; without readings it waits.
	edo := termsOf(t, "EDO0535", "0.0625")
	due := []treasurybonds.DueMaturity{
		{Account: "a", Wrapper: "taxable", Lot: heldLot(t, "EDO0535", "1", "2025-05-05"), Terms: edo.Terms, Due: mustDate(t, "2035-05-05")},
		{Account: "b", Wrapper: "pension", Lot: heldLot(t, "OTS0426", "1", "2026-01-05"), Terms: termsOf(t, "OTS0426", "0.03").Terms, Due: mustDate(t, "2026-04-05")},
	}
	m, err := treasurybonds.BookMaturities(due, nil)
	if err != nil {
		t.Fatalf("BookMaturities: %v", err)
	}
	want := []string{"a EDO0535 bought 2025-05-05 maturity 2035-05-05", "b OTS0426 bought 2026-01-05 maturity 2026-04-05"}
	if diff := cmp.Diff(want, problemRefs(m.Problems)); diff != "" || len(m.Transactions) != 0 {
		t.Errorf("BookMaturities problems (-want +got):\n%s\ntransactions: %d", diff, len(m.Transactions))
	}
	years, rate := treasurybonds.MaturityCPIYears(due)
	if diff := cmp.Diff([]int{2026, 2027, 2028, 2029, 2030, 2031, 2032, 2033, 2034}, years); diff != "" || rate {
		t.Errorf("MaturityCPIYears (-want +got):\n%s\nrate: %v", diff, rate)
	}
}

// booked writes each transaction of m as "kind external ID date cash".
func booked(m *ingest.Mapping) []string {
	out := make([]string, 0, len(m.Transactions))
	for i := range m.Transactions {
		tx := &m.Transactions[i]
		cash := ""
		for _, p := range tx.Postings {
			if p.Book.Kind == ledger.BookCash {
				cash = p.Amount.String()
			}
		}
		out = append(out, tx.Kind.String()+" "+tx.Key.ExternalID+" "+tx.Date.String()+" "+cash)
	}
	return out
}

func TestPayoutWithdrawal(t *testing.T) {
	t.Parallel()
	coi := termsOf(t, "COI1028", "0.0655")
	coupon := treasurybonds.DueCoupon{
		Account: account, Wrapper: "taxable", Lot: heldLot(t, "COI1028", "10", "2024-10-10"), Terms: coi.Terms, Period: 1,
		Due: mustDate(t, "2025-10-10"), Purchased: mustDate(t, "2024-10-10"),
	}
	// 6.55% of 1000 PLN is 65.50; Belka 12.45 leaves 53.05 in cash.
	interest := "interest lot:COI1028:2024-10-10:coupon:2025-10-10 2025-10-10 53.05 PLN"
	tests := []struct {
		name   string
		payout ledger.PayoutSetting
		want   []string
	}{
		{name: "kept as cash: interest only", want: []string{interest}},
		{
			name:   "paid out: interest, then a withdrawal of the cash it brought on the same day",
			payout: ledger.PayoutSetting{Payout: ledger.PayoutPaidOut, Since: mustDate(t, "2025-10-10")},
			want:   []string{interest, "withdrawal lot:COI1028:2024-10-10:coupon:2025-10-10:payout 2025-10-10 -53.05 PLN"},
		},
		{
			name:   "paid out only from a later day: what fell due before follows the earlier setting",
			payout: ledger.PayoutSetting{Payout: ledger.PayoutPaidOut, Since: mustDate(t, "2025-10-11"), Earlier: []ledger.PayoutPeriod{{Payout: ledger.PayoutCash}}},
			want:   []string{interest},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := coupon
			c.Payout = tt.payout
			m, err := treasurybonds.BookCoupons([]treasurybonds.DueCoupon{c}, nil)
			if err != nil {
				t.Fatalf("BookCoupons: %v", err)
			}
			if diff := cmp.Diff(tt.want, booked(&m)); diff != "" {
				t.Errorf("BookCoupons (-want +got):\n%s", diff)
			}
		})
	}
}
