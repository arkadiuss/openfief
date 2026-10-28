package pl_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// lotCase describes one synthetic lot. Rates are yearly fractions; the
// expected amounts are worked out by hand in each case's comment.
type lotCase struct {
	name      string
	series    string
	first     string
	margin    string
	fee       string
	idx       map[int]string
	quantity  string
	purchased string
	wrapper   pl.TaxWrapper
	on        string
	// want: gross, interest, fee, tax, net (PLN, whole lot), matured
	gross, interest, wantFee, tax, net string
	matured                            bool
}

func TestValueLot(t *testing.T) {
	t.Parallel()
	every5 := map[int]string{2: "0.03", 3: "0.03", 4: "0.03", 5: "0.03", 6: "0.03", 7: "0.03", 8: "0.03", 9: "0.03", 10: "0.03"}
	tests := []lotCase{
		{
			// a = 0 days: W = 100.00, no interest, fee capped at 0.
			name: "EDO on its purchase date is worth its nominal", series: "EDO0334", first: "0.06", margin: "0.02", fee: "2.00",
			quantity: "10", purchased: "2024-03-15", wrapper: pl.Taxable, on: "2024-03-15",
			gross: "1000", interest: "0", wantFee: "0", tax: "0", net: "1000",
		},
		{
			// ACT = 365 (2024-03-15..2025-03-15), a = 184.
			// W = 100 × (1 + 0.06 × 184/365) = 100 + 1104/365 = 103.0247 -> 103.02; interest 3.02.
			// Fee 2.00 < 3.02, not capped. Lot of 10: gross 1030.20, interest 30.20, fee 20.00.
			// Tax = 19% × (30.20 - 20.00) = 1.938 -> 1.94. Net = 1030.20 - 20.00 - 1.94 = 1008.26.
			name: "EDO mid first year: 6% x 184/365, fee, Belka on interest minus fee", series: "EDO0334", first: "0.06", margin: "0.02", fee: "2.00",
			quantity: "10", purchased: "2024-03-15", wrapper: pl.Taxable, on: "2024-09-15",
			gross: "1030.20", interest: "30.20", wantFee: "20.00", tax: "1.94", net: "1008.26",
		},
		{
			// Same lot in IKE: no tax, net = 1030.20 - 20.00 = 1010.20.
			name: "same EDO lot inside IKE pays no Belka tax", series: "EDO0334", first: "0.06", margin: "0.02", fee: "2.00",
			quantity: "10", purchased: "2024-03-15", wrapper: pl.IKE, on: "2024-09-15",
			gross: "1030.20", interest: "30.20", wantFee: "20.00", tax: "0", net: "1010.20",
		},
		{
			// a = 31: W = 100 + 600 × 31/36500 = 100.5096 -> 100.51; interest 0.51 < fee 2.00,
			// so the fee is capped at 0.51 and the bond redeems at its nominal, untaxed.
			name: "EDO early-redemption fee capped at accrued interest", series: "EDO0334", first: "0.06", margin: "0.02", fee: "2.00",
			quantity: "1", purchased: "2024-03-15", wrapper: pl.Taxable, on: "2024-04-15",
			gross: "100.51", interest: "0.51", wantFee: "0.51", tax: "0", net: "100",
		},
		{
			// Year 2: r2 = CPI 4.9% + margin 2% = 6.9%. ACT = 365 (2025-03-15..2026-03-15), a = 184.
			// W = 100 × 1.06 × (1 + 0.069 × 184/365) = 106 + 1345.776/365 = 109.6871 -> 109.69.
			// Interest 9.69, fee 2.00, tax = 19% × 7.69 = 1.4611 -> 1.46, net = 109.69 - 2.00 - 1.46 = 106.23.
			name: "EDO year 2 compounds year 1 and indexes to CPI plus margin", series: "EDO0334", first: "0.06", margin: "0.02", fee: "2.00",
			idx: map[int]string{2: "0.049"}, quantity: "1", purchased: "2024-03-15", wrapper: pl.Taxable, on: "2025-09-15",
			gross: "109.69", interest: "9.69", wantFee: "2.00", tax: "1.46", net: "106.23",
		},
		{
			// Deflation: CPI -1% counts as 0, r2 = margin 2%. a = 364 of ACT 365.
			// W = 106 × (1 + 0.02 × 364/365) = 106 + 771.68/365 = 108.1142 -> 108.11.
			// Fee 2.00, tax = 19% × 6.11 = 1.1609 -> 1.16, net = 108.11 - 2.00 - 1.16 = 104.95.
			name: "EDO in deflation earns only the margin", series: "EDO0334", first: "0.06", margin: "0.02", fee: "2.00",
			idx: map[int]string{2: "-0.01"}, quantity: "1", purchased: "2024-03-15", wrapper: pl.Taxable, on: "2026-03-14",
			gross: "108.11", interest: "8.11", wantFee: "2.00", tax: "1.16", net: "104.95",
		},
		{
			// r1 = 6%, r2..r10 = CPI 3% + 2% = 5%. W = 100 × 1.06 × 1.05^9 = 106 × 1.551328216 = 164.4408 -> 164.44.
			// Matured: no fee, valued at maturity even after it. Tax = 19% × 64.44 = 12.2436 -> 12.24, net 152.20.
			name: "EDO after maturity: redemption value, no fee", series: "EDO0334", first: "0.06", margin: "0.02", fee: "2.00",
			idx: every5, quantity: "1", purchased: "2024-03-15", wrapper: pl.Taxable, on: "2034-06-01",
			gross: "164.44", interest: "64.44", wantFee: "0", tax: "12.24", net: "152.20", matured: true,
		},
		{
			// COI pays coupons. ACT = 365 (2025-01-10..2026-01-10), a = 181.
			// Interest = 100 × 0.065 × 181/365 = 1176.5/365 = 3.2233 -> 3.22; gross 103.22.
			// Fee 2.00 < 3.22. Tax = 19% × 1.22 = 0.2318 -> 0.23. Net = 103.22 - 2.00 - 0.23 = 100.99.
			name: "COI first year accrues the current coupon", series: "COI0129", first: "0.065", margin: "0.0125", fee: "2.00",
			quantity: "1", purchased: "2025-01-10", wrapper: pl.Taxable, on: "2025-07-10",
			gross: "103.22", interest: "3.22", wantFee: "2.00", tax: "0.23", net: "100.99",
		},
		{
			// a = 10: interest = 650/36500 = 0.1781 -> 0.18 < fee 2.00, capped at 0.18 in period 1.
			name: "COI fee capped at interest in the first period", series: "COI0129", first: "0.065", margin: "0.0125", fee: "2.00",
			quantity: "1", purchased: "2025-01-10", wrapper: pl.Taxable, on: "2025-01-20",
			gross: "100.18", interest: "0.18", wantFee: "0.18", tax: "0", net: "100",
		},
		{
			// Year 1 coupon already paid out. r2 = 4% + 1.25% = 5.25%, a = 10 of ACT 365:
			// interest = 525/36500 = 0.1438 -> 0.14 per bond. The full fee 2.00 applies after period 1.
			// Lot of 5: gross 500.70, interest 0.70, fee 10.00, interest - fee < 0 so no tax, net 490.70.
			name: "COI after the first period charges the full fee, cutting into nominal", series: "COI0129", first: "0.065", margin: "0.0125", fee: "2.00",
			idx: map[int]string{2: "0.04"}, quantity: "5", purchased: "2025-01-10", wrapper: pl.Taxable, on: "2026-01-20",
			gross: "500.70", interest: "0.70", wantFee: "10.00", tax: "0", net: "490.70",
		},
		{
			// Monthly period 2025-01-31..2025-02-28: ACT 28, a = 14.
			// Interest = 100 × 0.0575 × 14 / (12 × 28) = 80.5/336 = 0.2396 -> 0.24; fee 0.50 capped at 0.24.
			name: "ROR first month: interest r/12 x a/ACT, fee capped", series: "ROR0126", first: "0.0575", margin: "0", fee: "0.50",
			quantity: "1", purchased: "2025-01-31", wrapper: pl.Taxable, on: "2025-02-14",
			gross: "100.24", interest: "0.24", wantFee: "0.24", tax: "0", net: "100",
		},
		{
			// Period 2 is 2025-02-28..2025-03-31 (ACT 31), a = 15, r2 = reference 5.75% + 0.
			// Interest = 100 × 0.0575 × 15 / (12 × 31) = 86.25/372 = 0.2319 -> 0.23. Full fee 0.50, no tax.
			// Net = 100.23 - 0.50 = 99.73.
			name: "ROR second month: reference rate, full fee", series: "ROR0126", first: "0.0575", margin: "0", fee: "0.50",
			idx: map[int]string{2: "0.0575"}, quantity: "1", purchased: "2025-01-31", wrapper: pl.Taxable, on: "2025-03-15",
			gross: "100.23", interest: "0.23", wantFee: "0.50", tax: "0", net: "99.73",
		},
		{
			// One 3-month period at 3% a year: W = 100 × (1 + 0.03 × 3/12) = 100.75.
			// Matured: no fee. Tax = 19% × 0.75 = 0.1425 -> 0.14. Net 100.61.
			name: "OTS at maturity: 3 months of 3% a year", series: "OTS0825", first: "0.03", margin: "0", fee: "0",
			quantity: "1", purchased: "2025-05-05", wrapper: pl.Taxable, on: "2025-08-05",
			gross: "100.75", interest: "0.75", wantFee: "0", tax: "0.14", net: "100.61", matured: true,
		},
		{
			// Leap-day purchase: period 2 is 2025-02-28..2026-02-28 (ACT 365), a = 30, fixed 5.95%.
			// W = 105.95 × (1 + 0.0595 × 30/365) = 105.95 + 189.12075/365 = 106.4681 -> 106.47.
			// Fee 1.00, tax = 19% × 5.47 = 1.0393 -> 1.04, net = 106.47 - 1.00 - 1.04 = 104.43.
			name: "TOS bought on a leap day compounds its fixed rate", series: "TOS0227", first: "0.0595", margin: "0", fee: "1.00",
			quantity: "1", purchased: "2024-02-29", wrapper: pl.Taxable, on: "2025-03-30",
			gross: "106.47", interest: "6.47", wantFee: "1.00", tax: "1.04", net: "104.43",
		},
		{
			// IKZE: same as the TOS case without tax: net = 106.47 - 1.00 = 105.47.
			name: "TOS inside IKZE withholds no tax", series: "TOS0227", first: "0.0595", margin: "0", fee: "1.00",
			quantity: "1", purchased: "2024-02-29", wrapper: pl.IKZE, on: "2025-03-30",
			gross: "106.47", interest: "6.47", wantFee: "1.00", tax: "0", net: "105.47",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := parseLot(t, &tt)
			got, err := pl.ValueLot(&a.terms, a.idx, a.quantity, a.purchased, a.wrapper, a.on)
			if err != nil {
				t.Fatalf("ValueLot: %v", err)
			}
			want := pl.LotValue{
				Gross: plnMoney(t, tt.gross), Interest: plnMoney(t, tt.interest), Fee: plnMoney(t, tt.wantFee),
				Tax: plnMoney(t, tt.tax), Net: plnMoney(t, tt.net), Matured: tt.matured,
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestValueLotRejectsBadInput(t *testing.T) {
	t.Parallel()
	ok := lotCase{series: "EDO0334", first: "0.06", margin: "0.02", fee: "2.00", quantity: "1", purchased: "2024-03-15", wrapper: pl.Taxable, on: "2024-09-15"}
	tests := []struct {
		name    string
		edit    func(c lotCase) lotCase
		wantErr error
	}{
		{name: "valued before purchase", edit: func(c lotCase) lotCase { c.on = "2024-03-14"; return c }, wantErr: pl.ErrBeforePurchase},
		{name: "purchase outside the series month", edit: func(c lotCase) lotCase { c.purchased = "2024-04-02"; c.on = "2024-09-15"; return c }, wantErr: pl.ErrSeriesMismatch},
		{name: "fractional quantity", edit: func(c lotCase) lotCase { c.quantity = "1.5"; return c }, wantErr: pl.ErrInvalidQuantity},
		{name: "zero quantity", edit: func(c lotCase) lotCase { c.quantity = "0"; return c }, wantErr: pl.ErrInvalidQuantity},
		{name: "negative fee", edit: func(c lotCase) lotCase { c.fee = "-1"; return c }, wantErr: pl.ErrInvalidTerms},
		{name: "negative first rate", edit: func(c lotCase) lotCase { c.first = "-0.01"; return c }, wantErr: pl.ErrInvalidTerms},
		{name: "unknown wrapper", edit: func(c lotCase) lotCase { c.wrapper = pl.TaxWrapper(7); return c }, wantErr: pl.ErrUnknownTaxWrapper},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := tt.edit(ok)
			a := parseLot(t, &c)
			_, err := pl.ValueLot(&a.terms, a.idx, a.quantity, a.purchased, a.wrapper, a.on)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("error (-want +got):\n%s", diff)
			}
		})
	}
	t.Run("fee in another currency", func(t *testing.T) {
		t.Parallel()
		usd, err := kernel.ParseCurrency("USD")
		if err != nil {
			t.Fatal(err)
		}
		terms := pl.Terms{Series: series(t, "EDO0334"), FirstPeriodRate: dec(t, "0.06"), EarlyRedemptionFee: kernel.NewMoney(dec(t, "2"), usd)}
		_, err = pl.ValueLot(&terms, nil, kernel.NewQuantity(dec(t, "1")), date(t, "2024-03-15"), pl.Taxable, date(t, "2024-09-15"))
		if diff := cmp.Diff(pl.ErrInvalidTerms, err, cmpopts.EquateErrors()); diff != "" {
			t.Errorf("error (-want +got):\n%s", diff)
		}
	})
	t.Run("missing CPI for the current period", func(t *testing.T) {
		t.Parallel()
		c := ok
		c.on = "2025-09-15"
		a := parseLot(t, &c)
		_, err := pl.ValueLot(&a.terms, a.idx, a.quantity, a.purchased, a.wrapper, a.on)
		want := &pl.MissingIndexError{Series: series(t, "EDO0334"), Period: 2, Start: date(t, "2025-03-15")}
		if diff := cmp.Diff(error(want), err); diff != "" {
			t.Errorf("error (-want +got):\n%s", diff)
		}
	})
}

// lotArgs are a lotCase's inputs, parsed.
type lotArgs struct {
	terms         pl.Terms
	idx           pl.Indices
	quantity      kernel.Quantity
	purchased, on kernel.Date
	wrapper       pl.TaxWrapper
}

func parseLot(t *testing.T, c *lotCase) lotArgs {
	t.Helper()
	return lotArgs{
		terms: pl.Terms{
			Series: series(t, c.series), FirstPeriodRate: dec(t, c.first), Margin: dec(t, c.margin),
			EarlyRedemptionFee: plnMoney(t, c.fee),
		},
		idx: indices(t, c.idx), quantity: kernel.NewQuantity(dec(t, c.quantity)),
		purchased: date(t, c.purchased), on: date(t, c.on), wrapper: c.wrapper,
	}
}
