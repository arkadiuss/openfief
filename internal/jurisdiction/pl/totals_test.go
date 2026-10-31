package pl_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// rowsOf builds an account's rows: rows and listed are gross, net and
// liquidation; then the listed cost basis, sale cost, realized gains and
// what was paid in.
func rowsOf(t *testing.T, wrapper jurisdiction.Wrapper, rows, listed [3]string, cost, saleCost, realized, paidIn string) jurisdiction.AccountRows {
	t.Helper()
	v := func(a [3]string) jurisdiction.Values {
		return jurisdiction.Values{Gross: plnMoney(t, a[0]), Net: plnMoney(t, a[1]), Liquidation: plnMoney(t, a[2])}
	}
	return jurisdiction.AccountRows{
		Wrapper: wrapper, Rows: v(rows), Listed: v(listed), ListedCost: plnMoney(t, cost), ListedSaleCost: plnMoney(t, saleCost),
		Realized: plnMoney(t, realized), PaidIn: plnMoney(t, paidIn),
	}
}

// t1 is a taxable account: holding A worth 10 000 at a cost of 8 000 (own
// tax 380, net 9 620; sold for 9 990 after a 10.00 sale cost, 1 990 gain,
// 378.10 tax, liquidation 9 611.90), holding B worth 5 000 at a cost of
// 6 000 (a loss: net 5 000; liquidation 4 995) and 1 000 of cash. Its rows
// sum to 16 000, 15 620 and 15 606.90.
func t1(t *testing.T, realized string) jurisdiction.AccountRows {
	t.Helper()
	return rowsOf(t, "taxable", [3]string{"16000", "15620", "15606.90"}, [3]string{"15000", "14620", "14606.90"}, "14000", "15", realized, "0")
}

func TestTotals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		accounts []jurisdiction.AccountRows
		pit      string
		want     [3]string // gross, net, liquidation; "?" unknown
	}{
		{
			// B's loss offsets A's gain: net base 15 000 − 14 000 = 1 000, tax
			// 190, net 16 000 − 190 = 15 810 (rows 15 620: offset +190).
			// Liquidation base 1 000 − 15 = 985, tax 187.15 → 187; proceeds
			// 14 985 + 1 000 cash − 187 = 15 798 (rows 15 606.90: offset +191.10).
			name:     "a taxable account's holdings offset each other",
			accounts: []jurisdiction.AccountRows{t1(t, "0")},
			want:     [3]string{"16000", "15810", "15798"},
		},
		{
			// A realized loss of 1 500 this year: bases −500 and −515, no tax.
			name:     "the year's realized loss offsets unrealized gains",
			accounts: []jurisdiction.AccountRows{t1(t, "-1500")},
			want:     [3]string{"16000", "16000", "15985"},
		},
		{
			// A realized gain of 500: bases 1 500 (285) and 1 485 (282.15 → 282).
			name:     "the year's realized gain is taxed with them",
			accounts: []jurisdiction.AccountRows{t1(t, "500")},
			want:     [3]string{"16000", "15715", "15703"},
		},
		{
			// Two taxable accounts: 10 000 at a cost of 8 000 and 3 000 at a
			// cost of 4 000. Base 1 000, tax 190: net 13 000 − 190 = 12 810.
			name: "taxable accounts offset each other",
			accounts: []jurisdiction.AccountRows{
				rowsOf(t, "taxable", [3]string{"10000", "9620", "9620"}, [3]string{"10000", "9620", "9620"}, "8000", "0", "0", "0"),
				rowsOf(t, "taxable", [3]string{"3000", "3000", "3000"}, [3]string{"3000", "3000", "3000"}, "4000", "0", "0", "0"),
			},
			want: [3]string{"13000", "12810", "12810"},
		},
		{
			// Base 2 234.50 − 1 000 = 1 234.50 → 1 235; 19% is 234.65 → 235.
			name:     "base and tax rounded to whole złoty",
			accounts: []jurisdiction.AccountRows{rowsOf(t, "taxable", [3]string{"2234.50", "1999.95", "1999.95"}, [3]string{"2234.50", "1999.95", "1999.95"}, "1000", "0", "0", "0")},
			want:     [3]string{"2234.50", "1999.50", "1999.50"},
		},
		{
			// A bond keeps its own row's tax: 1 000 gross, 980 net, 960 if
			// redeemed today; nothing listed, so no PIT-38 tax.
			name:     "bonds keep their own tax",
			accounts: []jurisdiction.AccountRows{rowsOf(t, "taxable", [3]string{"1000", "980", "960"}, [3]string{"0", "0", "0"}, "0", "0", "0", "0")},
			want:     [3]string{"1000", "980", "960"},
		},
		{
			// IKE: 11 000 listed + 1 000 cash; a return sells for 11 989 after
			// an 11.00 sale cost, 1 989 above the 10 000 paid in: 377.91 tax.
			name:     "an IKE's early return is taxed above the payments",
			accounts: []jurisdiction.AccountRows{rowsOf(t, "ike", [3]string{"12000", "12000", "11800"}, [3]string{"11000", "11000", "10800"}, "9000", "11", "0", "10000")},
			want:     [3]string{"12000", "12000", "11611.09"},
		},
		{
			name:     "an IKE's return below the payments is not taxed",
			accounts: []jurisdiction.AccountRows{rowsOf(t, "ike", [3]string{"12000", "12000", "11800"}, [3]string{"11000", "11000", "10800"}, "9000", "11", "0", "13000")},
			want:     [3]string{"12000", "12000", "11989"},
		},
		{
			// IKZE: 10% of 10 000 at the exit; 32% of the 9 990 proceeds,
			// 3 196.80, before it.
			name: "IKZE taxes the whole amount", pit: "0.32",
			accounts: []jurisdiction.AccountRows{rowsOf(t, "ikze", [3]string{"10000", "9000", "6793.20"}, [3]string{"10000", "9000", "6793.20"}, "8000", "10", "0", "0")},
			want:     [3]string{"10000", "9000", "6793.20"},
		},
		{
			name:     "IKZE without the marginal rate",
			accounts: []jurisdiction.AccountRows{rowsOf(t, "ikze", [3]string{"10000", "9000", "0"}, [3]string{"10000", "9000", "0"}, "8000", "10", "0", "0")},
			want:     [3]string{"10000", "9000", "?"},
		},
		{
			// OKI's rows add up as they are; with the taxable account of the
			// first case, 16 000 + 5 000 and so on.
			name:     "OKI adds its rows, other accounts their own rules",
			accounts: []jurisdiction.AccountRows{t1(t, "0"), rowsOf(t, "oki", [3]string{"5000", "5000", "4990"}, [3]string{"5000", "5000", "4990"}, "6000", "10", "0", "0")},
			want:     [3]string{"21000", "20810", "20788"},
		},
		{
			name:     "nothing",
			accounts: nil,
			want:     [3]string{"0", "0", "0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pl.NewPack().Totals(tt.accounts, settings(t, tt.pit))
			c := valuesCase{want: tt.want}
			checkValues(t, &c, &got, err)
		})
	}
}

func TestTotalsErrors(t *testing.T) {
	t.Parallel()
	bad := rowsOf(t, "taxable", [3]string{"1", "1", "1"}, [3]string{"0", "0", "0"}, "0", "0", "0", "0")
	bad.Wrapper = "isa"
	foreign := rowsOf(t, "taxable", [3]string{"1", "1", "1"}, [3]string{"0", "0", "0"}, "0", "0", "0", "0")
	usd, err := kernel.ParseCurrency("USD")
	if err != nil {
		t.Fatal(err)
	}
	foreign.Realized = kernel.NewMoney(dec(t, "1"), usd)
	for name, tt := range map[string]struct {
		rows jurisdiction.AccountRows
		want error
	}{
		"an unknown wrapper": {bad, pl.ErrUnknownTaxWrapper},
		"not in PLN":         {foreign, pl.ErrNotPLN},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := pl.NewPack().Totals([]jurisdiction.AccountRows{tt.rows}, nil)
			if diff := cmp.Diff(tt.want, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestPIT38Tax(t *testing.T) {
	t.Parallel()
	tests := []struct{ base, want string }{
		{"1234.50", "235"}, // base 1 235, 234.65 → 235
		{"1234.49", "234"}, // base 1 234, 234.46 → 234
		{"2.63", "1"},      // base 3, 0.57 → 1
		{"0.49", "0"},      // base 0
		{"0", "0"},
		{"-100", "0"},
	}
	for _, tt := range tests {
		t.Run(tt.base, func(t *testing.T) {
			t.Parallel()
			got, err := pl.PIT38Tax(dec(t, tt.base))
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(dec(t, tt.want), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

// TestTotalsProperties checks, over generated taxable accounts of two
// holdings, that offsetting only lowers the tax: the net total is at most
// the gross total and at least the rows' net values summed, less the under
// one złoty that rounding to whole złoty can add.
func TestTotalsProperties(t *testing.T) {
	t.Parallel()
	r := &rng{state: 128}
	thousand, rate := dec(t, "1000"), pl.BelkaRate()
	for range 300 {
		gross, net, cost := kernel.Decimal{}, kernel.Decimal{}, kernel.Decimal{}
		for range 2 {
			g := thousand.Mul(kernel.NewDecimal(int64(1+r.intn(50)), 0)).Add(kernel.NewDecimal(int64(r.intn(100)), -2))
			c := thousand.Mul(kernel.NewDecimal(int64(1+r.intn(50)), 0))
			own := kernel.Decimal{}
			if gain := g.Sub(c); gain.Sign() > 0 {
				var err error
				if own, err = gain.Mul(rate).Round(2, kernel.RoundHalfUp); err != nil {
					t.Fatal(err)
				}
			}
			gross, net, cost = gross.Add(g), net.Add(g.Sub(own)), cost.Add(c)
		}
		v := [3]string{gross.String(), net.String(), net.String()}
		a := rowsOf(t, "taxable", v, v, cost.String(), "0", "0", "0")
		got, err := pl.NewPack().Totals([]jurisdiction.AccountRows{a}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got.Net.Amount().Cmp(gross) > 0 || got.Net.Amount().Cmp(net.Sub(kernel.NewDecimal(1, 0))) < 0 {
			t.Fatalf("gross %s, rows' net %s, cost %s: net total %s", gross, net, cost, got.Net)
		}
	}
}
