package pl_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
)

// TestValueLotMatchesInterestTables checks ValueLot against the per-bond
// daily values the issuer publishes at
// https://www.obligacjeskarbowe.pl/tabela-odsetkowa/ ("Zestawienie
// zaakceptowane przez Ministerstwo Finansów"). Each table is for one bond
// bought on the first day of the series' sale month and lists the accrued
// interest (gross minus 100 PLN) on every day of one interest period. The
// lots sit in IKE, so no fee or tax hides the gross value.
func TestValueLotMatchesInterestTables(t *testing.T) {
	t.Parallel()
	// EDO1133 (margin 1.50%): 7.00% in year 1, then 6.40% and 4.40%, which is
	// September CPI of 4.9% and 2.9% plus the margin.
	edo1133 := lotCase{
		series: "EDO1133", first: "0.07", margin: "0.015", fee: "2.00", idx: map[int]string{2: "0.049", 3: "0.029"},
		quantity: "1", purchased: "2023-11-01", wrapper: pl.IKE,
	}
	// TOS0128: 5.95% fixed for three years.
	tos0128 := lotCase{
		series: "TOS0128", first: "0.0595", margin: "0", fee: "0.70", quantity: "1", purchased: "2025-01-01", wrapper: pl.IKE,
	}
	tests := []lotCase{
		// Year 1 runs 2023-11-01..2024-11-01: ACT 366 (it holds 2024-02-29), a = 121.
		// 7 × 121/366 = 2.3142 -> 2.31 (ACT 365 would give 2.32). Table: 2,31.
		at(&edo1133, "EDO1133 table, year 1, 2024-03-01: ACT counts the leap day", "2024-03-01", "102.31", "2.31"),
		// Year 3: 100 × 1.07 × 1.064 × (1 + 0.044 × 120/365) = 113.848 + 1.6469 = 115.4949 -> 115.49.
		// Rounding the base to 113.85 first would give 115.50. Table: 15,49.
		at(&edo1133, "EDO1133 table, year 3, 2026-03-01: compounded once, rounded once", "2026-03-01", "115.49", "15.49"),
		// a = 15: 113.848 × (1 + 0.044 × 15/365) = 114.0538 -> 114.05; a rounded base gives 114.06. Table: 14,05.
		at(&edo1133, "EDO1133 table, year 3, 2025-11-16", "2025-11-16", "114.05", "14.05"),
		// Year 2 of TOS: 105.95 × (1 + 0.0595 × 59/365) = 106.9690 -> 106.97. Table: 6,97.
		at(&tos0128, "TOS0128 table, year 2, 2026-03-01", "2026-03-01", "106.97", "6.97"),
		// Year 3: 112.254025 × (1 + 0.0595 × 3/365) = 112.3089 -> 112.31. The letter's rounded
		// base 112.25 would give 112.30. Table: 12,31.
		at(&tos0128, "TOS0128 table, year 3, 2027-01-04: the table's unrounded base", "2027-01-04", "112.31", "12.31"),
		{
			// COI0128 (margin 1.25%), year 2 at 4.70% CPI + 1.25% = 5.95%; the year-1 coupon
			// was paid out. a = 59 of 365: 5.95 × 59/365 = 0.9618 -> 0.96. Table: 0,96.
			name: "COI0128 table, year 2, 2025-03-01: only the current coupon accrues", series: "COI0128",
			first: "0.0665", margin: "0.0125", fee: "0.70", idx: map[int]string{2: "0.047"},
			quantity: "1", purchased: "2024-01-01", wrapper: pl.IKE, on: "2025-03-01",
			gross: "100.96", interest: "0.96", wantFee: "0", tax: "0", net: "100.96",
		},
		{
			// ROR1026 bought 2025-10-31 (its own table): period 5 runs 2026-02-28..2026-03-31,
			// ACT 31, at 4.00%. a = 15: 100 × 0.04 × 15 / (31 × 12) = 0.1613 -> 0.16. Table: 0,16.
			// Periods 2 to 4 were paid out; their readings don't change the value.
			name: "ROR1026 table, bought on the 31st, period 5 back to the 31st", series: "ROR1026",
			first: "0.0475", margin: "0", fee: "0.50", idx: map[int]string{2: "0.045", 3: "0.04", 4: "0.04", 5: "0.04"},
			quantity: "1", purchased: "2025-10-31", wrapper: pl.IKE, on: "2026-03-15",
			gross: "100.16", interest: "0.16", wantFee: "0", tax: "0", net: "100.16",
		},
		{
			// Bought 2025-10-29: period 5 runs 2026-02-28..2026-03-29, ACT 29.
			// 100 × 0.04 × 15 / (29 × 12) = 0.1724 -> 0.17. Table: 0,17.
			name: "ROR1026 table, bought on the 29th, period 5 back to the 29th", series: "ROR1026",
			first: "0.0475", margin: "0", fee: "0.50", idx: map[int]string{2: "0.045", 3: "0.04", 4: "0.04", 5: "0.04"},
			quantity: "1", purchased: "2025-10-29", wrapper: pl.IKE, on: "2026-03-15",
			gross: "100.17", interest: "0.17", wantFee: "0", tax: "0", net: "100.17",
		},
	}
	runLotCases(t, tests)
}

// at returns c valued on date on in IKE, where the expected net equals the
// gross value.
func at(base *lotCase, name, on, gross, interest string) lotCase {
	c := *base
	c.name, c.on = name, on
	c.gross, c.interest, c.net, c.wantFee, c.tax = gross, interest, gross, "0", "0"
	return c
}

// TestValueLotMatchesWorkedExamples checks ValueLot against the early
// redemption examples on the offer pages under
// https://www.obligacjeskarbowe.pl/oferta-obligacji/ (section "Wcześniejsze
// zakończenie oszczędzania", "Przykład"). Each example values one bond on the
// fifth business day after the redemption order, the day its interest is
// counted to.
func TestValueLotMatchesWorkedExamples(t *testing.T) {
	t.Parallel()
	tests := []lotCase{
		{
			// EDO0524 bought 10.05.2014 at 4.00%, 1.50% in year 2; "101,75 zł = 100 zł + 4,16 zł
			// (odsetki narosłe na 16.06.2015 r.) – 2 zł [...] – 0,41 zł (wartość podatku od kwoty
			// 2,16 zł". Year 2 runs 2015-05-10..2016-05-10 (ACT 366), a = 37:
			// 104 × (1 + 0.015 × 37/366) = 104.1577 -> 104.16. The example gives only the year-2
			// rate, so the index and margin are set to make 1.50%.
			name: "EDO0524 example: 101.75 PLN", series: "EDO0524", first: "0.04", margin: "0.015", fee: "2.00",
			idx: map[int]string{2: "0"}, quantity: "1", purchased: "2014-05-10", wrapper: pl.Taxable, on: "2015-06-16",
			gross: "104.16", interest: "4.16", wantFee: "2.00", tax: "0.41", net: "101.75",
		},
		{
			// TOS0825 bought 1.08.2022 at 6.50%; "105,88 zł = 100 zł + 7,96 zł (odsetki narosłe na
			// 17.10.2023 r.) – 0,70 zł [...] – 1,38 zł". Year 2 runs 2023-08-01..2024-08-01
			// (ACT 366), a = 77: 106.5 × (1 + 0.065 × 77/366) = 107.9564 -> 107.96.
			name: "TOS0825 example: 105.88 PLN", series: "TOS0825", first: "0.065", margin: "0", fee: "0.70",
			quantity: "1", purchased: "2022-08-01", wrapper: pl.Taxable, on: "2023-10-17",
			gross: "107.96", interest: "7.96", wantFee: "0.70", tax: "1.38", net: "105.88",
		},
		{
			// ROS1022 bought 10.10.2016 at 2.60%, 3.00% in year 2: "narosłe odsetki wynoszą 2,91 zł",
			// fee 0.70, tax 0.42, "101,79 zł". a = 37 of 365: 102.6 × (1 + 0.03 × 37/365) = 102.9120.
			// The index and margin are set to make the example's 3.00%.
			name: "ROS1022 example: 101.79 PLN", series: "ROS1022", first: "0.026", margin: "0.03", fee: "0.70",
			idx: map[int]string{2: "0"}, quantity: "1", purchased: "2016-10-10", wrapper: pl.Taxable, on: "2017-11-16",
			gross: "102.91", interest: "2.91", wantFee: "0.70", tax: "0.42", net: "101.79",
		},
		{
			// DOR0624 bought 1.06.2022 at 5.50%, redeemed in period 1: "100,00 zł = 100 zł + 0,43 zł
			// (odsetki narosłe na 29.06.2022 r. [...]) – 0,43 zł (wysokość opłaty [...] pobrana do
			// wysokości narosłych odsetek)". 100 × 0.055 × 28 / (30 × 12) = 0.4278 -> 0.43.
			name: "DOR0624 example, period 1: fee capped, 100.00 PLN", series: "DOR0624", first: "0.055", margin: "0", fee: "0.70",
			quantity: "1", purchased: "2022-06-01", wrapper: pl.Taxable, on: "2022-06-29",
			gross: "100.43", interest: "0.43", wantFee: "0.43", tax: "0", net: "100.00",
		},
		{
			// Period 2 at 5.50%: "99,70 zł = 100 zł + 0,40 zł (odsetki narosłe na 28.07.2022 r.) –
			// 0,70 zł". 100 × 0.055 × 27 / (31 × 12) = 0.3992 -> 0.40; the full fee applies.
			name: "DOR0624 example, period 2: full fee, 99.70 PLN", series: "DOR0624", first: "0.055", margin: "0", fee: "0.70",
			idx: map[int]string{2: "0.055"}, quantity: "1", purchased: "2022-06-01", wrapper: pl.Taxable, on: "2022-07-28",
			gross: "100.40", interest: "0.40", wantFee: "0.70", tax: "0", net: "99.70",
		},
		{
			// ROR0623 bought 1.06.2022 at 5.25%: "100,00 zł = 100 zł + 0,41 zł [...] – 0,41 zł".
			// 100 × 0.0525 × 28 / (30 × 12) = 0.4083 -> 0.41.
			name: "ROR0623 example, period 1: fee capped, 100.00 PLN", series: "ROR0623", first: "0.0525", margin: "0", fee: "0.50",
			quantity: "1", purchased: "2022-06-01", wrapper: pl.Taxable, on: "2022-06-29",
			gross: "100.41", interest: "0.41", wantFee: "0.41", tax: "0", net: "100.00",
		},
		{
			// Period 2 at 5.25%: "99,68 zł = 100 zł + 0,18 zł (odsetki narosłe na 14.07.2022 r.) –
			// 0,50 zł". 100 × 0.0525 × 13 / (31 × 12) = 0.1835 -> 0.18.
			name: "ROR0623 example, period 2: full fee, 99.68 PLN", series: "ROR0623", first: "0.0525", margin: "0", fee: "0.50",
			idx: map[int]string{2: "0.0525"}, quantity: "1", purchased: "2022-06-01", wrapper: pl.Taxable, on: "2022-07-14",
			gross: "100.18", interest: "0.18", wantFee: "0.50", tax: "0", net: "99.68",
		},
		{
			// OTS0118 bought 10.10.2017 at 1.5%, redeemed early: "otrzymuje: 100,00 zł – kwotę,
			// którą wpłacił w dniu zakupu obligacji".
			name: "OTS0118 example: early redemption returns 100.00 PLN", series: "OTS0118", first: "0.015", margin: "0", fee: "0",
			quantity: "1", purchased: "2017-10-10", wrapper: pl.Taxable, on: "2017-11-28",
			gross: "100", interest: "0", wantFee: "0", tax: "0", net: "100",
		},
	}
	runLotCases(t, tests)
}
