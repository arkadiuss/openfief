package pl_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// span is n consecutive days from a date with the same value in each
// class: savings, investment and other ("" leaves the class out).
type span struct {
	from                       string
	n                          int
	savings, investment, other string
}

// flow is a day's payments in and out.
type flow struct{ date, in, out string }

// okiAccount builds an account's days from spans, with flows on their days.
func okiAccount(t *testing.T, name string, spans []span, flows ...flow) jurisdiction.TaxAccount {
	t.Helper()
	a := jurisdiction.TaxAccount{Account: name}
	for i := range spans {
		s := spans[i]
		d := date(t, s.from)
		for range s.n {
			a.Days = append(a.Days, okiDay(t, d, &s, flows))
			d = d.AddDays(1)
		}
	}
	return a
}

// okiDay builds day d of span s, with its flow if one is on it.
func okiDay(t *testing.T, d kernel.Date, s *span, flows []flow) jurisdiction.TaxDay {
	t.Helper()
	day := jurisdiction.TaxDay{Date: d, In: plnMoney(t, "0"), Out: plnMoney(t, "0")}
	for _, cv := range []struct {
		class jurisdiction.ExemptionClass
		value string
	}{{pl.Savings, s.savings}, {pl.Investment, s.investment}, {pl.Other, s.other}} {
		if cv.value != "" {
			day.Values = append(day.Values, jurisdiction.ClassValue{Class: cv.class, Value: plnMoney(t, cv.value)})
		}
	}
	for _, f := range flows {
		if f.date == d.String() {
			day.In, day.Out = plnMoney(t, f.in), plnMoney(t, f.out)
		}
	}
	return day
}

// classes returns savings, investment and other amounts as class values.
func classes(t *testing.T, savings, investment, other string) []jurisdiction.ClassValue {
	t.Helper()
	return []jurisdiction.ClassValue{
		{Class: pl.Savings, Value: plnMoney(t, savings)},
		{Class: pl.Investment, Value: plnMoney(t, investment)},
		{Class: pl.Other, Value: plnMoney(t, other)},
	}
}

// okiWant is an expected tax: days, through, averages per class, exempt
// savings and investment, taxed, rate and tax ("" for unknown).
type okiWant struct {
	days                       int
	through                    string
	avg                        [3]string
	exemptS, exemptI, taxed    string
	rate, tax                  string
	exemptionsKnown, rateKnown bool
}

func (w *okiWant) build(t *testing.T, year int) jurisdiction.YearlyTax {
	t.Helper()
	out := jurisdiction.YearlyTax{Wrapper: "oki", Year: year, InForce: true, Days: w.days, Averages: classes(t, w.avg[0], w.avg[1], w.avg[2])}
	if w.through != "" {
		out.Through = date(t, w.through)
	}
	total := dec(t, w.avg[0]).Add(dec(t, w.avg[1])).Add(dec(t, w.avg[2]))
	out.Average = kernel.NewMoney(total, plnMoney(t, "0").Currency())
	if w.exemptionsKnown {
		out.ExemptionsKnown, out.Exempt, out.Taxed = true, classes(t, w.exemptS, w.exemptI, "0"), plnMoney(t, w.taxed)
	}
	if w.rateKnown {
		out.RateKnown, out.Rate, out.Tax = true, dec(t, w.rate), plnMoney(t, w.tax)
	}
	return out
}

// laterAct is the real act with a reference rate in force on 31 October
// 2027 and 2029 and a CPI index for 2030, so later years can be computed.
func laterAct(t *testing.T) pl.OKIAct {
	t.Helper()
	act := pl.CurrentOKIAct()
	act.ReferenceRates = []pl.RateChange{
		{From: date(t, "2027-10-01"), Rate: dec(t, "0.0575")},
		{From: date(t, "2029-06-01"), Rate: dec(t, "0.0500")},
	}
	act.CPI = map[int]kernel.Decimal{2030: dec(t, "1.035")}
	return act
}

func TestOKIYearlyTax(t *testing.T) {
	t.Parallel()
	known := func(w okiWant) okiWant { w.exemptionsKnown, w.rateKnown = true, true; return w }
	tests := []struct {
		name     string
		act      pl.OKIAct
		year     int
		accounts []jurisdiction.TaxAccount
		want     okiWant
	}{
		{
			// 20 000 of savings all year: all of it exempt (up to 25 000).
			name: "savings under the exemption", act: pl.CurrentOKIAct(), year: 2027,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2027-01-01", 365, "20000", "", ""}})},
			want:     known(okiWant{days: 365, through: "2027-12-31", avg: [3]string{"20000", "0", "0"}, exemptS: "20000", exemptI: "0", taxed: "0", rate: "0.0085", tax: "0"}),
		},
		{
			// 150 000 of investment assets: 100 000 exempt, 50 000 × 0.85% = 425.
			name: "investment above the exemption", act: pl.CurrentOKIAct(), year: 2027,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2027-01-01", 365, "", "150000", ""}})},
			want:     known(okiWant{days: 365, through: "2027-12-31", avg: [3]string{"0", "150000", "0"}, exemptS: "0", exemptI: "100000", taxed: "50000", rate: "0.0085", tax: "425.00"}),
		},
		{
			// 30 000 savings + 90 000 investment: savings first, 25 000; then
			// investment up to 100 000 − 25 000 = 75 000. Taxed 120 000 − 100 000
			// = 20 000; 0.85% is 170.
			name: "exemptions straddle savings and investment", act: pl.CurrentOKIAct(), year: 2027,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2027-01-01", 365, "30000", "90000", ""}})},
			want:     known(okiWant{days: 365, through: "2027-12-31", avg: [3]string{"30000", "90000", "0"}, exemptS: "25000", exemptI: "75000", taxed: "20000", rate: "0.0085", tax: "170.00"}),
		},
		{
			// 2028 is a leap year of 366 days: 200 000 of other assets for the
			// 183 days through 1 July, then nothing. 36 600 000 / 366 = 100 000,
			// none exempt. Rate: 19% of the 5.75% in force on 2027-10-31 is
			// 1.0925%, down to 1.09%; 100 000 × 1.09% = 1 090.
			name: "leap year", act: laterAct(t), year: 2028,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2028-01-01", 183, "", "", "200000"}, {"2028-07-02", 183, "", "", "0"}})},
			want:     known(okiWant{days: 366, through: "2028-12-31", avg: [3]string{"0", "0", "100000"}, exemptS: "0", exemptI: "0", taxed: "100000", rate: "0.0109", tax: "1090.00"}),
		},
		{
			// Opened 1 July: held 184 days. 92 days at 150 000 and 92 at
			// 218 000: 33 856 000 / 184 = 184 000; 84 000 taxed, 714.
			name: "account opened mid-year", act: pl.CurrentOKIAct(), year: 2027,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2027-07-01", 92, "", "150000", ""}, {"2027-10-01", 92, "", "218000", ""}})},
			want:     known(okiWant{days: 184, through: "2027-12-31", avg: [3]string{"0", "184000", "0"}, exemptS: "0", exemptI: "100000", taxed: "84000", rate: "0.0085", tax: "714.00"}),
		},
		{
			// Payments count up to the day's withdrawals: 10 000 in and 4 000
			// out counts 4 000; 5 000 in alone counts nothing; 2 000 in and
			// 7 000 out (withdrawals exceeding payments) counts 2 000. Sum
			// 36 500 000 + 6 000 = 36 506 000, / 365 = 100 016.438… → 100 016.44;
			// 16.44 taxed, × 0.85% = 0.13974 → 0.14.
			name: "payments up to withdrawals", act: pl.CurrentOKIAct(), year: 2027,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2027-01-01", 365, "", "100000", ""}},
				flow{"2027-02-01", "10000", "4000"}, flow{"2027-03-01", "5000", "0"}, flow{"2027-04-01", "2000", "7000"})},
			want: known(okiWant{days: 365, through: "2027-12-31", avg: [3]string{"0", "100016.44", "0"}, exemptS: "0", exemptI: "100000", taxed: "16.44", rate: "0.0085", tax: "0.14"}),
		},
		{
			// Two classes share the counted payments in proportion to their
			// sums (art. 26 ust. 3): 3 650 000 and 10 950 000 of 14 600 000,
			// so 1/4 and 3/4 of 3 650: (3 650 000 + 912.5) / 365 = 10 002.50
			// and (10 950 000 + 2 737.5) / 365 = 30 007.50.
			name: "payments split by class", act: pl.CurrentOKIAct(), year: 2027,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2027-01-01", 365, "10000", "30000", ""}}, flow{"2027-05-05", "3650", "3650"})},
			want:     known(okiWant{days: 365, through: "2027-12-31", avg: [3]string{"10002.50", "30007.50", "0"}, exemptS: "10002.50", exemptI: "30007.50", taxed: "0", rate: "0.0085", tax: "0"}),
		},
		{
			// Each account's average is over its own days: 80 000 all year and
			// 73 000 from 1 October (92 days) sum to 153 000; 53 000 taxed, 450.50.
			name: "two accounts summed", act: pl.CurrentOKIAct(), year: 2027,
			accounts: []jurisdiction.TaxAccount{
				okiAccount(t, "a", []span{{"2027-01-01", 365, "", "80000", ""}}),
				okiAccount(t, "b", []span{{"2027-10-01", 92, "", "73000", ""}}),
			},
			want: known(okiWant{days: 365, through: "2027-12-31", avg: [3]string{"0", "153000", "0"}, exemptS: "0", exemptI: "100000", taxed: "53000", rate: "0.0085", tax: "450.50"}),
		},
		{
			// 300 001 / 3 = 100 000.333… → 100 000.33; × 0.85% = 850.0028 → 850.00.
			name: "the average rounds half up to grosze", act: pl.CurrentOKIAct(), year: 2027,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2027-12-29", 2, "", "", "100000"}, {"2027-12-31", 1, "", "", "100001"}})},
			want:     known(okiWant{days: 3, through: "2027-12-31", avg: [3]string{"0", "0", "100000.33"}, exemptS: "0", exemptI: "0", taxed: "100000.33", rate: "0.0085", tax: "850.00"}),
		},
		{
			// No value on either day but a 1 000 round trip: the payments go
			// to other assets, 1 000 / 2 = 500; 4.25.
			name: "payments on an empty account", act: pl.CurrentOKIAct(), year: 2027,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2027-12-30", 2, "", "", ""}}, flow{"2027-12-30", "1000", "1000"})},
			want:     known(okiWant{days: 2, through: "2027-12-31", avg: [3]string{"0", "0", "500.00"}, exemptS: "0", exemptI: "0", taxed: "500.00", rate: "0.0085", tax: "4.25"}),
		},
		{
			// Days outside the year don't count.
			name: "days of another year are left out", act: pl.CurrentOKIAct(), year: 2027,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2026-12-31", 2, "", "", "1000"}})},
			want:     known(okiWant{days: 1, through: "2027-01-01", avg: [3]string{"0", "0", "1000"}, exemptS: "0", exemptI: "0", taxed: "1000", rate: "0.0085", tax: "8.50"}),
		},
		{
			name: "no day held yet", act: pl.CurrentOKIAct(), year: 2027,
			want: known(okiWant{avg: [3]string{"0", "0", "0"}, exemptS: "0", exemptI: "0", taxed: "0", rate: "0.0085", tax: "0"}),
		},
		{
			// No reference rate known for 31 October 2027: averages and
			// exemptions, no rate or tax.
			name: "a later year's rate not known yet", act: pl.CurrentOKIAct(), year: 2028,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2028-01-01", 1, "", "150000", ""}})},
			want:     okiWant{days: 1, through: "2028-01-01", avg: [3]string{"0", "150000", "0"}, exemptS: "0", exemptI: "100000", taxed: "50000", exemptionsKnown: true},
		},
		{
			name: "2030's exemptions need the CPI index", act: pl.CurrentOKIAct(), year: 2030,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2030-01-01", 1, "", "150000", ""}})},
			want:     okiWant{days: 1, through: "2030-01-01", avg: [3]string{"0", "150000", "0"}},
		},
		{
			// 2030: exemptions × 1.035, down to hundreds: 25 875 → 25 800 and
			// 103 500. Savings 25 800 exempt, investment up to 103 500 − 25 800
			// = 77 700; taxed 130 000 − 103 500 = 26 500. Rate: 19% of 5% is
			// 0.95%; 251.75.
			name: "exemptions indexed with CPI from 2030", act: laterAct(t), year: 2030,
			accounts: []jurisdiction.TaxAccount{okiAccount(t, "a", []span{{"2030-01-01", 365, "30000", "100000", ""}})},
			want:     known(okiWant{days: 365, through: "2030-12-31", avg: [3]string{"30000", "100000", "0"}, exemptS: "25800", exemptI: "77700", taxed: "26500", rate: "0.0095", tax: "251.75"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &pl.Pack{OKI: tt.act}
			got, err := p.YearlyTax("oki", tt.year, tt.accounts)
			if err != nil {
				t.Fatalf("YearlyTax: %v", err)
			}
			if diff := cmp.Diff(tt.want.build(t, tt.year), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestOKIYearlyTaxNotDue(t *testing.T) {
	t.Parallel()
	usd, err := kernel.ParseCurrency("USD")
	if err != nil {
		t.Fatal(err)
	}
	dup := okiAccount(t, "a", []span{{"2027-01-01", 1, "", "", "1"}})
	dup.Days = append(dup.Days, dup.Days[0])
	foreign := okiAccount(t, "a", []span{{"2027-01-01", 1, "", "", "1"}})
	foreign.Days[0].Values[0].Value = kernel.NewMoney(dec(t, "1"), usd)
	unknown := okiAccount(t, "a", []span{{"2027-01-01", 1, "", "", "1"}})
	unknown.Days[0].Values[0].Class = "bonds"
	negative := okiAccount(t, "a", []span{{"2027-01-01", 1, "", "", "1"}}, flow{"2027-01-01", "-1", "0"})
	tests := []struct {
		name     string
		act      pl.OKIAct
		wrapper  jurisdiction.Wrapper
		year     int
		accounts []jurisdiction.TaxAccount
		want     jurisdiction.YearlyTax
		wantErr  error
	}{
		{name: "no act: nothing is computed", wrapper: "oki", year: 2027, want: jurisdiction.YearlyTax{Wrapper: "oki", Year: 2027}},
		{name: "a year before the act", act: pl.CurrentOKIAct(), wrapper: "oki", year: 2026, want: jurisdiction.YearlyTax{Wrapper: "oki", Year: 2026}},
		{name: "IKE has no yearly tax", act: pl.CurrentOKIAct(), wrapper: "ike", year: 2027, wantErr: jurisdiction.ErrNoYearlyTax},
		{name: "a day given twice", act: pl.CurrentOKIAct(), wrapper: "oki", year: 2027, accounts: []jurisdiction.TaxAccount{dup}, wantErr: pl.ErrTaxDay},
		{name: "a value not in PLN", act: pl.CurrentOKIAct(), wrapper: "oki", year: 2027, accounts: []jurisdiction.TaxAccount{foreign}, wantErr: pl.ErrTaxDay},
		{name: "a class the pack lacks", act: pl.CurrentOKIAct(), wrapper: "oki", year: 2027, accounts: []jurisdiction.TaxAccount{unknown}, wantErr: pl.ErrTaxDay},
		{name: "a negative payment", act: pl.CurrentOKIAct(), wrapper: "oki", year: 2027, accounts: []jurisdiction.TaxAccount{negative}, wantErr: pl.ErrTaxDay},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &pl.Pack{OKI: tt.act}
			got, err := p.YearlyTax(tt.wrapper, tt.year, tt.accounts)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestOKIRate(t *testing.T) {
	t.Parallel()
	tests := []struct{ reference, want string }{
		{"0.0575", "0.0109"}, // 19% is 1.0925%, down to 1.09%
		{"0.0525", "0.0099"}, // 0.9975% down to 0.99%
		{"0.0675", "0.0128"}, // 1.2825% down to 1.28%
		{"0.0040", "0.001"},  // 0.076%: at least 0.1%
		{"0", "0.001"},
	}
	for _, tt := range tests {
		t.Run(tt.reference, func(t *testing.T) {
			t.Parallel()
			got, err := pl.OKIRate(dec(t, tt.reference))
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(dec(t, tt.want), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestIndexOKIExemption(t *testing.T) {
	t.Parallel()
	tests := []struct{ amount, cpi, want string }{
		{"25000", "1.035", "25800"},   // 25 875 down to hundreds
		{"100000", "1.035", "103500"}, // already whole hundreds
		{"25800", "1.041", "26800"},   // 26 857.8, compounding from the year before
		{"25000", "1", "25000"},
	}
	for _, tt := range tests {
		t.Run(tt.amount+"x"+tt.cpi, func(t *testing.T) {
			t.Parallel()
			got, err := pl.IndexOKIExemption(plnMoney(t, tt.amount), dec(t, tt.cpi))
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(plnMoney(t, tt.want), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestCurrentOKIAct(t *testing.T) {
	t.Parallel()
	act := pl.CurrentOKIAct()
	want := pl.OKIAct{
		InForceFrom: date(t, "2027-01-01"),
		Rates:       map[int]kernel.Decimal{2027: dec(t, "0.0085")},
		Exemptions:  pl.OKIExemptions{Savings: plnMoney(t, "25000"), Investment: plnMoney(t, "100000"), Total: plnMoney(t, "100000")},
		IndexedFrom: 2030,
	}
	if diff := cmp.Diff(want, act); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(act, pl.NewPack().OKI); diff != "" {
		t.Errorf("NewPack's act (-want +got):\n%s", diff)
	}
}

func TestOKIExemptionClass(t *testing.T) {
	t.Parallel()
	pln := plnMoney(t, "0").Currency()
	usd, err := kernel.ParseCurrency("USD")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		wrapper jurisdiction.Wrapper
		asset   jurisdiction.Asset
		want    jurisdiction.ExemptionClass
		known   bool
	}{
		{"a treasury savings bond", "oki", jurisdiction.Asset{Kind: jurisdiction.AssetAccrued, Instrument: "PLEDO0536000", Currency: pln}, pl.Savings, true},
		{"PLN cash on a bond account", "oki", jurisdiction.Asset{Kind: jurisdiction.AssetCash, Currency: pln, AccountHoldsAccrued: true}, pl.Savings, true},
		{"PLN cash on a brokerage account", "oki", jurisdiction.Asset{Kind: jurisdiction.AssetCash, Currency: pln}, pl.Investment, true},
		{"foreign cash", "oki", jurisdiction.Asset{Kind: jurisdiction.AssetCash, Currency: usd}, pl.Other, true},
		{"a listed holding: the pack can't tell", "oki", jurisdiction.Asset{Kind: jurisdiction.AssetListed, Instrument: "PLXXXXXXXXX1", Currency: pln}, pl.Other, false},
		{"no class outside OKI", "taxable", jurisdiction.Asset{Kind: jurisdiction.AssetListed, Currency: pln}, "", true},
	}
	p := pl.NewPack()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, known := p.ExemptionClass(tt.wrapper, &tt.asset)
			if got != tt.want || known != tt.known {
				t.Errorf("ExemptionClass = %q, %v; want %q, %v", got, known, tt.want, tt.known)
			}
		})
	}
	if diff := cmp.Diff([]jurisdiction.ExemptionClass{"savings", "investment", "other"}, p.ExemptionClasses("oki")); diff != "" {
		t.Errorf("OKI's classes (-want +got):\n%s", diff)
	}
	if got := p.ExemptionClasses("ike"); got != nil {
		t.Errorf("IKE's classes = %v, want none", got)
	}
}
