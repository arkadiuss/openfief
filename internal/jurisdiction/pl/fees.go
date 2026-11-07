package pl

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Errors of the reviewed fee table and of the owner's fee.
var (
	// ErrNoReviewedFee reports a series sold before the reviewed fee
	// table's first row for its bond type: the table holds no fee for it,
	// so the owner enters that series' terms (D-97).
	ErrNoReviewedFee = errors.New("no reviewed early-redemption fee for the series")
	// ErrFeeNotAssumed reports the owner confirming a series' fee that is
	// not assumed: there is nothing to confirm.
	ErrFeeNotAssumed = errors.New("the series' early-redemption fee is not assumed")
	// ErrUnknownFeeSource reports a fee source that is none of FeeSource's.
	ErrUnknownFeeSource = errors.New("unknown early-redemption fee source")
)

// FeeSource says where a series' early-redemption fee came from (D-97).
// The zero source is the owner's: terms the owner enters carry their own
// fee.
type FeeSource int

// The fee sources.
const (
	// FeeEntered is a fee the owner entered, or an assumed one the owner
	// confirmed.
	FeeEntered FeeSource = iota
	// FeeFromTable is the reviewed fee table's fee for the series' sale
	// month, read in its issue letter.
	FeeFromTable
	// FeeAssumed is the latest reviewed fee of the series' bond type,
	// carried forward to a series sold after the month the table was
	// reviewed through, whose letter nobody has read yet. It stands until
	// the owner confirms or overrides it.
	FeeAssumed
)

// The fee sources' names, as stored and as the read API writes them.
const (
	feeEnteredName = "entered"
	feeTableName   = "table"
	feeAssumedName = "assumed"
)

// String returns "entered", "table" or "assumed", or a placeholder for an
// invalid source.
func (s FeeSource) String() string {
	switch s {
	case FeeEntered:
		return feeEnteredName
	case FeeFromTable:
		return feeTableName
	case FeeAssumed:
		return feeAssumedName
	}
	return fmt.Sprintf("FeeSource(%d)", int(s))
}

// ParseFeeSource reads a fee source's name (FeeSource.String); any other
// text fails with ErrUnknownFeeSource.
func ParseFeeSource(name string) (FeeSource, error) {
	for _, s := range []FeeSource{FeeEntered, FeeFromTable, FeeAssumed} {
		if s.String() == name {
			return s, nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrUnknownFeeSource, name)
}

// FeeRow is one row of the reviewed early-redemption fee table (D-97): from
// the series first sold in the month starting From, every series of Type
// carries Fee per bond, until the next row of the same type. Letter names
// the issue letter of the first series with that fee.
type FeeRow struct {
	Type   BondType
	From   kernel.Date
	Fee    kernel.Money
	Letter string
}

// feesReviewedYear and feesReviewedMonth are the sale month of the last
// series whose issue letter was checked: November 2026.
const (
	feesReviewedYear  = 2026
	feesReviewedMonth = time.November
)

// FeesReviewedThrough returns the first day of the last sale month the fee
// table was reviewed through: the issue letter of every series of every
// bond type sold from the type's first row up to that month was read.
func FeesReviewedThrough() kernel.Date {
	return firstOfMonth(feesReviewedYear, feesReviewedMonth)
}

// EarlyRedemptionFees returns the reviewed fee table (D-97), per bond type
// in order of From. Each fee is the amount "b" by which an early
// redemption's payment per bond is reduced, "należność wypłacona z tytułu
// przedterminowego wykupu jednej obligacji jest pomniejszana o kwotę ...",
// as the Minister of Finance's issue letters (listy emisyjne) set it. The
// letters are official documents, which carry no copyright (R-27); they were
// read as published at
// https://www.finanse.mf.gov.pl/dlug-publiczny/obligacje-detaliczne/wyszukiwarka-listow-emisyjnych1,
// every series' letter from each type's first row through
// FeesReviewedThrough (R-28):
//
//   - OTS: no fee from OTS0118 (sold from October 2017, list emisyjny nr
//     55/2017) through OTS0227 (nr 103/2026): an early redemption pays the
//     nominal and no interest ("odsetki od obligacji nie przysługują").
//   - ROR: 0.50 PLN from ROR0623 (June 2022, nr 64/2022) through ROR1127
//     (nr 104/2026).
//   - DOR: 0.70 PLN from DOR0624 (June 2022, nr 65/2022) through DOR1128
//     (nr 105/2026).
//   - TOS: 0.70 PLN from TOS0825 (August 2022, nr 87/2022); 1.00 PLN from
//     TOS0927 (September 2024, nr 94/2024) through TOS1129 (nr 106/2026).
//   - COI: 1.00 PLN from COI0118 (January 2014, nr 61/2013); 0.70 PLN from
//     COI0419 (April 2015, nr 14/2015); 2.00 PLN from COI0928 (September
//     2024, nr 95/2024) through COI1130 (nr 107/2026).
//   - EDO: 2.00 PLN from EDO0124 (January 2014, nr 62/2013); 3.00 PLN from
//     EDO0934 (September 2024, nr 96/2024) through EDO1136 (nr 108/2026).
//   - ROS: 0.70 PLN from ROS1022 (October 2016, nr 47/2016); 2.00 PLN from
//     ROS0930 (September 2024, nr 97/2024) through ROS1132 (nr 109/2026).
//   - ROD: 2.00 PLN from ROD1028 (October 2016, nr 48/2016); 3.00 PLN from
//     ROD0936 (September 2024, nr 98/2024) through ROD1138 (nr 110/2026).
//
// Each type's first row is the first series of that type in the ministry's
// retail bond spreadsheet, or January 2014 for COI and EDO; series sold
// earlier are matured or nearly so, and their letters were not read.
func EarlyRedemptionFees() []FeeRow {
	return []FeeRow{
		feeRow(OTS, "2017-10", "0.00", "list emisyjny nr 55/2017 (OTS0118)"),
		feeRow(ROR, "2022-06", "0.50", "list emisyjny nr 64/2022 (ROR0623)"),
		feeRow(DOR, "2022-06", "0.70", "list emisyjny nr 65/2022 (DOR0624)"),
		feeRow(TOS, "2022-08", "0.70", "list emisyjny nr 87/2022 (TOS0825)"),
		feeRow(TOS, "2024-09", "1.00", "list emisyjny nr 94/2024 (TOS0927)"),
		feeRow(COI, "2014-01", "1.00", "list emisyjny nr 61/2013 (COI0118)"),
		feeRow(COI, "2015-04", "0.70", "list emisyjny nr 14/2015 (COI0419)"),
		feeRow(COI, "2024-09", "2.00", "list emisyjny nr 95/2024 (COI0928)"),
		feeRow(EDO, "2014-01", "2.00", "list emisyjny nr 62/2013 (EDO0124)"),
		feeRow(EDO, "2024-09", "3.00", "list emisyjny nr 96/2024 (EDO0934)"),
		feeRow(ROS, "2016-10", "0.70", "list emisyjny nr 47/2016 (ROS1022)"),
		feeRow(ROS, "2024-09", "2.00", "list emisyjny nr 97/2024 (ROS0930)"),
		feeRow(ROD, "2016-10", "2.00", "list emisyjny nr 48/2016 (ROD1028)"),
		feeRow(ROD, "2024-09", "3.00", "list emisyjny nr 98/2024 (ROD0936)"),
	}
}

// feeRow is the row of type t from the first day of month, "YYYY-MM",
// with a fee of amount PLN. The table is checked in, and its test reads
// every row, so a row that doesn't parse is a zero row its test fails on.
func feeRow(t BondType, month, amount, letter string) FeeRow {
	from, err1 := kernel.ParseDate(month + "-01")
	fee, err2 := kernel.ParseMoney(amount, pln())
	if err1 != nil || err2 != nil {
		return FeeRow{Type: t, Letter: letter}
	}
	return FeeRow{Type: t, From: from, Fee: fee, Letter: letter}
}

// firstOfMonth is the first day of month in year, which is always a date.
func firstOfMonth(year int, month time.Month) kernel.Date {
	d, err := kernel.NewDate(year, month, 1)
	if err != nil {
		return kernel.Date{}
	}
	return d
}

// SaleMonth returns the first day of the calendar month in which series s
// is sold: its maturity month less its bond type's term, since each bond's
// term runs from its own purchase date in that month. EDO1136 is sold in
// November 2026.
func SaleMonth(s Series) (kernel.Date, error) {
	st, err := s.Type.Structure()
	if err != nil {
		return kernel.Date{}, err
	}
	return addMonths(firstOfMonth(s.MaturityYear, s.MaturityMonth), -st.TenorMonths)
}

// TableFee is a series' early-redemption fee from the reviewed table: the
// fee, whether it is the table's own for the series or assumed, and the
// row it comes from, which names its issue letter.
type TableFee struct {
	Fee    kernel.Money
	Source FeeSource
	Row    FeeRow
}

// EarlyRedemptionFee returns the early-redemption fee per bond of series s
// from the reviewed table (EarlyRedemptionFees, D-97): the fee of the
// latest row of its bond type from on or before its sale month
// (SaleMonth), as FeeFromTable. A series sold after FeesReviewedThrough
// carries that latest row's fee forward as FeeAssumed, until the owner
// confirms or overrides it. A series sold before the type's first row
// fails with ErrNoReviewedFee: the table has no fee to carry back.
func EarlyRedemptionFee(s Series) (TableFee, error) {
	return feeFrom(EarlyRedemptionFees(), FeesReviewedThrough(), s)
}

// feeFrom looks series s up in the fee table rows, reviewed through the
// sale month through.
func feeFrom(rows []FeeRow, through kernel.Date, s Series) (TableFee, error) {
	sold, err := SaleMonth(s)
	if err != nil {
		return TableFee{}, err
	}
	var (
		out   TableFee
		found bool
	)
	for _, r := range rows {
		if r.Type == s.Type && !r.From.After(sold) {
			out, found = TableFee{Fee: r.Fee, Source: FeeFromTable, Row: r}, true
		}
	}
	if !found {
		return TableFee{}, fmt.Errorf("%w: %s is sold in %s, before the table's first %s row", ErrNoReviewedFee, s, sold, s.Type)
	}
	if sold.After(through) {
		out.Source = FeeAssumed
	}
	return out, nil
}

// AssumedFeeDetail is what the owner is asked about a series whose fee is
// assumed (FeeAssumed): "confirm the early-redemption fee for EDO1137:
// 3.00 PLN assumed from the latest reviewed letter, list emisyjny nr
// 96/2024 (EDO0934)". It names the letter when the table still has the
// row the fee comes from.
func AssumedFeeDetail(t *Terms) string {
	out := fmt.Sprintf("confirm the early-redemption fee for %s: %s assumed from the latest reviewed letter", t.Series, inGrosze(t.EarlyRedemptionFee))
	if f, err := EarlyRedemptionFee(t.Series); err == nil && f.Fee.Equal(t.EarlyRedemptionFee) {
		out += ", " + f.Row.Letter
	}
	return out
}

// inGrosze writes m with exactly two decimal places, half up: "3.00 PLN".
func inGrosze(m kernel.Money) string {
	const places = 2
	r, err := m.Amount().Round(places, kernel.RoundHalfUp)
	if err != nil {
		return m.String()
	}
	whole, frac, _ := strings.Cut(r.String(), ".")
	frac += strings.Repeat("0", places-len(frac))
	return whole + "." + frac + " " + m.Currency().String()
}

// FeesToConfirm returns the terms among held whose early-redemption fee is
// assumed (FeeAssumed), in series order: each one the owner is asked to
// confirm or override (D-97).
func FeesToConfirm(held []Terms) []Terms {
	var out []Terms
	for i := range held {
		if held[i].FeeSource == FeeAssumed {
			out = append(out, held[i])
		}
	}
	slices.SortFunc(out, func(a, b Terms) int { return strings.Compare(a.Series.String(), b.Series.String()) })
	return out
}

// ConfirmFee returns t with its assumed fee confirmed by the owner: the
// same fee, now the owner's (FeeEntered). A fee that is not assumed fails
// with ErrFeeNotAssumed.
func (t *Terms) ConfirmFee() (Terms, error) {
	if t.FeeSource != FeeAssumed {
		return Terms{}, fmt.Errorf("%w: %s's fee is %s", ErrFeeNotAssumed, t.Series, t.FeeSource)
	}
	out := *t
	out.FeeSource = FeeEntered
	return out, nil
}

// OverrideFee returns t with fee as the owner's early-redemption fee
// (FeeEntered), whatever the fee came from before (D-97). A fee that is not
// a PLN amount of zero or more fails with ErrInvalidTerms.
func (t *Terms) OverrideFee(fee kernel.Money) (Terms, error) {
	if !fee.Currency().Equal(pln()) || fee.Sign() < 0 {
		return Terms{}, fmt.Errorf("%w: %s's early-redemption fee must be PLN, zero or more", ErrInvalidTerms, t.Series)
	}
	out := *t
	out.EarlyRedemptionFee, out.FeeSource = fee, FeeEntered
	return out, nil
}
