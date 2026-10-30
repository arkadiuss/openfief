package pl_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestPackDefaults(t *testing.T) {
	t.Parallel()
	p := pl.NewPack()
	got := []string{p.Version(), p.ReportingCurrency().String(), p.NumberFormat().String()}
	if diff := cmp.Diff([]string{"pl/" + pl.RulesVersion, "PLN", "space-comma"}, got); diff != "" {
		t.Errorf("version, currency, number format (-want +got):\n%s", diff)
	}
	want := []jurisdiction.Wrapper{"taxable", "ike", "ikze", "oki"}
	if diff := cmp.Diff(want, p.Wrappers()); diff != "" {
		t.Errorf("wrappers (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(dec(t, "0.10"), pl.IKZEExitRate()); diff != "" {
		t.Errorf("IKZE exit rate (-want +got):\n%s", diff)
	}
}

// valuesCase is one wrapper's expected values; liquidation "?" is unknown.
type valuesCase struct {
	name    string
	wrapper string
	pit     string // the owner's marginal rate; empty when not set
	want    [3]string
	wantErr error
}

// settings returns the owner's settings with marginal rate pit, if any.
func settings(t *testing.T, pit string) *jurisdiction.Settings {
	t.Helper()
	if pit == "" {
		return &jurisdiction.Settings{}
	}
	return &jurisdiction.Settings{MarginalRate: dec(t, pit), HasMarginalRate: true}
}

// wantValues builds the expected values of c.
func wantValues(t *testing.T, c *valuesCase) jurisdiction.Values {
	t.Helper()
	v := jurisdiction.Values{Gross: plnMoney(t, c.want[0]), Net: plnMoney(t, c.want[1])}
	if c.want[2] == "?" {
		v.Liquidation, v.LiquidationUnknown = plnMoney(t, "0"), true
	} else {
		v.Liquidation = plnMoney(t, c.want[2])
	}
	return v
}

func checkValues(t *testing.T, c *valuesCase, got *jurisdiction.Values, err error) {
	t.Helper()
	if diff := cmp.Diff(c.wantErr, err, cmpopts.EquateErrors()); diff != "" {
		t.Fatalf("error (-want +got):\n%s", diff)
	}
	if err != nil {
		return
	}
	if diff := cmp.Diff(wantValues(t, c), *got); diff != "" {
		t.Errorf("gross, net, liquidation (-want +got):\n%s", diff)
	}
}

func TestValueListed(t *testing.T) {
	t.Parallel()
	// Synthetic: worth 10 000.00 PLN, bought for 8 000.00, selling costs 29.00.
	tests := []valuesCase{
		{
			// Net: 10 000 - 19% x 2 000 = 9 620. Liquidation: proceeds 9 971,
			// gain 1 971, tax 374.49, 9 596.51.
			name: "taxable: Belka on the gain, the sale cost deducted", wrapper: "taxable",
			want: [3]string{"10000", "9620", "9596.51"},
		},
		{name: "IKE: untaxed at the exit, a return taxed like a sale", wrapper: "ike", want: [3]string{"10000", "10000", "9596.51"}},
		{name: "IKZE: 10% exit tax; liquidation missing without the PIT rate", wrapper: "ikze", want: [3]string{"10000", "9000", "?"}},
		{
			// 9 971 - 32% x 9 971 (3 190.72) = 6 780.28.
			name: "IKZE: an early withdrawal taxed at the marginal rate", wrapper: "ikze", pit: "0.32",
			want: [3]string{"10000", "9000", "6780.28"},
		},
		{name: "OKI: no Belka tax inside", wrapper: "oki", want: [3]string{"10000", "10000", "9971"}},
		{name: "an unknown wrapper", wrapper: "ppk", wantErr: pl.ErrUnknownTaxWrapper},
	}
	p := pl.NewPack()
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := p.ValueListed(&jurisdiction.Listed{
				Wrapper: jurisdiction.Wrapper(tt.wrapper), Gross: plnMoney(t, "10000"), Cost: plnMoney(t, "8000"), SaleCost: plnMoney(t, "29"),
			}, settings(t, tt.pit))
			checkValues(t, tt, &got, err)
		})
	}
	t.Run("a loss is not taxed", func(t *testing.T) {
		t.Parallel()
		got, err := p.ValueListed(&jurisdiction.Listed{Wrapper: "taxable", Gross: plnMoney(t, "7000"), Cost: plnMoney(t, "8000"), SaleCost: plnMoney(t, "14")}, settings(t, ""))
		checkValues(t, &valuesCase{want: [3]string{"7000", "7000", "6986"}}, &got, err)
	})
	t.Run("tax rounds half up to grosze: 19% of 0.05 is 0.01", func(t *testing.T) {
		t.Parallel()
		got, err := p.ValueListed(&jurisdiction.Listed{Wrapper: "taxable", Gross: plnMoney(t, "100.05"), Cost: plnMoney(t, "100"), SaleCost: plnMoney(t, "0")}, settings(t, ""))
		checkValues(t, &valuesCase{want: [3]string{"100.05", "100.04", "100.04"}}, &got, err)
	})
	t.Run("amounts must be PLN", func(t *testing.T) {
		t.Parallel()
		usd, err := kernel.ParseCurrency("USD")
		if err != nil {
			t.Fatal(err)
		}
		_, err = p.ValueListed(&jurisdiction.Listed{Wrapper: "taxable", Gross: kernel.NewMoney(dec(t, "1"), usd)}, settings(t, ""))
		checkValues(t, &valuesCase{wantErr: pl.ErrNotPLN}, &jurisdiction.Values{}, err)
	})
}

func TestValueAccrued(t *testing.T) {
	t.Parallel()
	// The EDO lot of TestValueLot: 10 bonds worth 1 030.20 with 30.20 of
	// interest on 2024-09-15; redeeming early charges 20.00 where a fee is due.
	tests := []struct {
		valuesCase

		discount, fee string
	}{
		{
			// Net: 1 030.20 - 19% x 30.20 (5.738 -> 5.74) = 1 024.46. Liquidation as
			// TestValueLot: 1 030.20 - 20.00 - 19% x 10.20 (1.94) = 1 008.26.
			valuesCase: valuesCase{name: "taxable: Belka at maturity, fee and Belka if redeemed today", wrapper: "taxable", want: [3]string{"1030.20", "1024.46", "1008.26"}},
			discount:   "0", fee: "20.00",
		},
		{
			// Bought by exchange at 99.90: income 31.20, net tax 5.928 -> 5.93; liquidation
			// tax 19% x 11.20 = 2.128 -> 2.13.
			valuesCase: valuesCase{name: "taxable: the exchange discount is income", wrapper: "taxable", want: [3]string{"1030.20", "1024.27", "1008.07"}},
			discount:   "1.00", fee: "20.00",
		},
		{
			valuesCase: valuesCase{name: "IKE: untaxed at the exit, Belka on a return, no fee", wrapper: "ike", want: [3]string{"1030.20", "1030.20", "1024.46"}},
			discount:   "0", fee: "0",
		},
		{
			// 10% of 1 030.20 = 103.02.
			valuesCase: valuesCase{name: "IKZE: 10% exit tax, liquidation missing without the PIT rate", wrapper: "ikze", want: [3]string{"1030.20", "927.18", "?"}},
			discount:   "0", fee: "0",
		},
		{
			// 12% of 1 030.20 = 123.624 -> 123.62.
			valuesCase: valuesCase{name: "IKZE at a 12% marginal rate", wrapper: "ikze", pit: "0.12", want: [3]string{"1030.20", "927.18", "906.58"}},
			discount:   "0", fee: "0",
		},
		{
			valuesCase: valuesCase{name: "OKI: no Belka tax, the fee charged", wrapper: "oki", want: [3]string{"1030.20", "1030.20", "1010.20"}},
			discount:   "0", fee: "20.00",
		},
		{
			valuesCase: valuesCase{name: "an unknown wrapper", wrapper: "", wantErr: pl.ErrUnknownTaxWrapper},
			discount:   "0", fee: "0",
		},
	}
	p := pl.NewPack()
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			income := dec(t, "30.20").Add(dec(t, tt.discount))
			got, err := p.ValueAccrued(&jurisdiction.Accrued{
				Wrapper: jurisdiction.Wrapper(tt.wrapper), Gross: plnMoney(t, "1030.20"),
				Income: kernel.NewMoney(income, pl.Nominal().Currency()), EarlyFee: plnMoney(t, tt.fee),
			}, settings(t, tt.pit))
			checkValues(t, &tt.valuesCase, &got, err)
		})
	}
}

func TestValueCash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		valuesCase

		amount string
	}{
		{valuesCase: valuesCase{name: "taxable cash is cash", wrapper: "taxable", want: [3]string{"500", "500", "500"}}, amount: "500"},
		{valuesCase: valuesCase{name: "IKE cash is cash", wrapper: "ike", want: [3]string{"500", "500", "500"}}, amount: "500"},
		{valuesCase: valuesCase{name: "OKI cash is cash", wrapper: "oki", want: [3]string{"500", "500", "500"}}, amount: "500"},
		{valuesCase: valuesCase{name: "IKZE cash pays the exit tax; liquidation missing", wrapper: "ikze", want: [3]string{"500", "450", "?"}}, amount: "500"},
		{valuesCase: valuesCase{name: "IKZE cash at a 32% marginal rate", wrapper: "ikze", pit: "0.32", want: [3]string{"500", "450", "340"}}, amount: "500"},
		{valuesCase: valuesCase{name: "negative IKZE cash is not taxed", wrapper: "ikze", pit: "0.32", want: [3]string{"-100", "-100", "-100"}}, amount: "-100"},
		{valuesCase: valuesCase{name: "a marginal rate of 0 is refused", wrapper: "ikze", pit: "0", wantErr: jurisdiction.ErrInvalidMarginalRate}, amount: "500"},
		{valuesCase: valuesCase{name: "an unknown wrapper", wrapper: "TAXABLE", wantErr: pl.ErrUnknownTaxWrapper}, amount: "500"},
	}
	p := pl.NewPack()
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := p.ValueCash(&jurisdiction.Cash{Wrapper: jurisdiction.Wrapper(tt.wrapper), Amount: plnMoney(t, tt.amount)}, settings(t, tt.pit))
			checkValues(t, &tt.valuesCase, &got, err)
		})
	}
}

// TestValuesProperties draws holdings, bond lots and cash in every wrapper
// from a seeded generator: no value ever exceeds the gross value, gross
// stays as given, and liquidation never exceeds net at Polish marginal
// rates (12% and 32%, both above IKZE's 10% exit tax).
func TestValuesProperties(t *testing.T) {
	t.Parallel()
	p := pl.NewPack()
	r := &rng{state: 20261029}
	grosze := func(below int) kernel.Money {
		return kernel.NewMoney(kernel.NewDecimal(int64(r.intn(below)), -2), pl.Nominal().Currency())
	}
	rates := []string{"", "0.12", "0.32"}
	for i := range 600 {
		w := p.Wrappers()[r.intn(len(p.Wrappers()))]
		s := settings(t, rates[r.intn(len(rates))])
		gross := grosze(10_000_000)
		part := func() kernel.Money { return grosze(10_000_000) }
		var got jurisdiction.Values
		var err error
		switch i % 3 {
		case 0:
			got, err = p.ValueListed(&jurisdiction.Listed{Wrapper: w, Gross: gross, Cost: grosze(10_000_000), SaleCost: minMoney(gross, part())}, s)
		case 1:
			got, err = p.ValueAccrued(&jurisdiction.Accrued{Wrapper: w, Gross: gross, Income: minMoney(gross, part()), EarlyFee: minMoney(gross, part())}, s)
		default:
			got, err = p.ValueCash(&jurisdiction.Cash{Wrapper: w, Amount: gross}, s)
		}
		if err != nil {
			t.Fatalf("case %d (%s): %v", i, w, err)
		}
		if !got.Gross.Equal(gross) {
			t.Errorf("case %d (%s): gross %s, want %s", i, w, got.Gross, gross)
		}
		if got.Net.Amount().Cmp(gross.Amount()) > 0 {
			t.Errorf("case %d (%s): net %s above gross %s", i, w, got.Net, gross)
		}
		if !got.LiquidationUnknown && got.Liquidation.Amount().Cmp(got.Net.Amount()) > 0 {
			t.Errorf("case %d (%s): liquidation %s above net %s", i, w, got.Liquidation, got.Net)
		}
	}
}

// minMoney returns the smaller of a and b, both in one currency.
func minMoney(a, b kernel.Money) kernel.Money {
	if b.Amount().Cmp(a.Amount()) < 0 {
		return b
	}
	return a
}

func TestYearlyTax(t *testing.T) {
	t.Parallel()
	// A synthetic act: the real one's limit and rate are not known to the code.
	act := func(from string) pl.OKIAct {
		return pl.OKIAct{InForceFrom: date(t, from), Limit: plnMoney(t, "100000"), Rate: dec(t, "0.008")}
	}
	day := func(d, balance string) jurisdiction.DayBalance {
		return jurisdiction.DayBalance{Date: date(t, d), Balance: plnMoney(t, balance)}
	}
	year := []jurisdiction.DayBalance{
		day("2026-12-31", "1000000"), day("2027-01-01", "90000"), day("2027-06-30", "120000"), day("2027-07-01", "150000"),
	}
	inForce := func(days int, avg, excess, tax string) jurisdiction.YearlyTax {
		return jurisdiction.YearlyTax{
			Wrapper: "oki", Year: 2027, InForce: true, Days: days, Average: plnMoney(t, avg),
			Limit: plnMoney(t, "100000"), Rate: dec(t, "0.008"), Excess: plnMoney(t, excess), Tax: plnMoney(t, tax),
		}
	}
	tests := []struct {
		name    string
		act     pl.OKIAct
		wrapper jurisdiction.Wrapper
		year    int
		days    []jurisdiction.DayBalance
		want    jurisdiction.YearlyTax
		wantErr error
	}{
		{
			name: "no act in force: nothing is computed", act: pl.CurrentOKIAct(), wrapper: "oki", year: 2027, days: year,
			want: jurisdiction.YearlyTax{Wrapper: "oki", Year: 2027},
		},
		{
			// Average of 90 000, 120 000 and 150 000 is 120 000; 20 000 above the limit, 0.8% is 160.
			name: "in force all year: the average above the limit is taxed", act: act("2027-01-01"), wrapper: "oki", year: 2027, days: year,
			want: inForce(3, "120000", "20000", "160.00"),
		},
		{
			name: "in force from July: earlier days don't count", act: act("2027-07-01"), wrapper: "oki", year: 2027, days: year,
			want: inForce(1, "150000", "50000", "400.00"),
		},
		{
			name: "a year before the act", act: act("2027-07-01"), wrapper: "oki", year: 2026, days: year,
			want: jurisdiction.YearlyTax{Wrapper: "oki", Year: 2026},
		},
		{
			name: "under the limit, no tax", act: act("2027-01-01"), wrapper: "oki", year: 2027, days: year[:2],
			want: inForce(1, "90000", "0", "0"),
		},
		{
			// 300 001 / 3 = 100 000.3333 -> 100 000.33; 0.33 above, 0.8% is 0.00264 -> 0.00.
			name: "the average rounds half up to grosze", act: act("2027-01-01"), wrapper: "oki", year: 2027,
			days: []jurisdiction.DayBalance{day("2027-01-01", "100000"), day("2027-01-02", "100000"), day("2027-01-03", "100001")},
			want: inForce(3, "100000.33", "0.33", "0.00"),
		},
		{
			name: "no valued day yet", act: act("2027-01-01"), wrapper: "oki", year: 2027,
			want: inForce(0, "0", "0", "0"),
		},
		{name: "IKE has no yearly tax", act: act("2027-01-01"), wrapper: "ike", year: 2027, wantErr: jurisdiction.ErrNoYearlyTax},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &pl.Pack{OKI: tt.act}
			got, err := p.YearlyTax(tt.wrapper, tt.year, tt.days)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
