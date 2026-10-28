package connectapi

import (
	"fmt"
	"math"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// enumNumber returns a domain enum constant's number as a protobuf enum
// number. The domain enums and the protobuf enums share their order, so the
// numbers name the same values (TestEnumsMatch checks every one); a number
// out of range gives the unspecified value, 0.
func enumNumber(d int) int32 {
	if d < 0 || d > math.MaxInt32 {
		return 0
	}
	return int32(d)
}

// parseDate reads a request's YYYY-MM-DD date.
func parseDate(s string) (kernel.Date, error) {
	d, err := kernel.ParseDate(s)
	if err != nil {
		return kernel.Date{}, fmt.Errorf("%w: date: %w", errBadParameter, err)
	}
	return d, nil
}

// pbDate writes a date as YYYY-MM-DD, or "" for the zero date.
func pbDate(d kernel.Date) string {
	if d.IsZero() {
		return ""
	}
	return d.String()
}

// pbMoney converts money, keeping its currency and exact amount.
func pbMoney(m kernel.Money) *viewsv1.Money {
	return &viewsv1.Money{Amount: m.Amount().String(), Currency: m.Currency().String()}
}

// pbMonies converts each of ms.
func pbMonies(ms []kernel.Money) []*viewsv1.Money {
	out := make([]*viewsv1.Money, 0, len(ms))
	for _, m := range ms {
		out = append(out, pbMoney(m))
	}
	return out
}

// pbInstrument converts an instrument.
func pbInstrument(in ledger.Instrument) *viewsv1.Instrument {
	return &viewsv1.Instrument{Isin: in.ISIN.String(), Symbol: in.Symbol}
}

// pbSleeve converts a sleeve.
func pbSleeve(s *portfolio.Sleeve) *viewsv1.Sleeve {
	return &viewsv1.Sleeve{Id: string(s.ID), Name: s.Name, Kind: string(s.Kind), Role: viewsv1.SleeveRole(enumNumber(int(s.Role)))}
}

// pbAccount converts an account.
func pbAccount(a *ledger.Account) *viewsv1.Account {
	return &viewsv1.Account{Id: string(a.ID), Institution: a.Institution, Name: a.Name, Sleeve: string(a.Sleeve)}
}

// pbBook converts a ledger account: a holding names its instrument, any other
// kind its currency.
func pbBook(b ledger.Book) *viewsv1.Book {
	out := &viewsv1.Book{Kind: viewsv1.BookKind(enumNumber(int(b.Kind)))}
	if b.Kind == ledger.BookHolding {
		out.Instrument = pbInstrument(b.Instrument)
		return out
	}
	out.Currency = b.Currency.String()
	return out
}

// pbTransaction converts a transaction with its postings.
func pbTransaction(t *ledger.Transaction) *viewsv1.Transaction {
	out := &viewsv1.Transaction{
		Source: t.Key.Source, ExternalId: t.Key.ExternalID, Kind: viewsv1.TransactionKind(enumNumber(int(t.Kind))),
		Date: pbDate(t.Date), LotOpened: pbDate(t.LotOpened), Description: t.Description,
		Postings: make([]*viewsv1.Posting, 0, len(t.Postings)),
	}
	for i := range t.Postings {
		p := &t.Postings[i]
		pp := &viewsv1.Posting{Book: pbBook(p.Book), Amount: pbMoney(p.Amount)}
		if p.Book.Kind == ledger.BookHolding {
			pp.Quantity = p.Quantity.String()
		}
		out.Postings = append(out.Postings, pp)
	}
	return out
}

// pbLot converts an open lot.
func pbLot(l *ledger.Lot) *viewsv1.Lot {
	return &viewsv1.Lot{
		Instrument: pbInstrument(l.Instrument), Opened: pbDate(l.Opened), Quantity: l.Quantity.String(),
		Cost: pbMoney(l.Cost), Opening: l.Opening, OriginSource: l.Origin.Source, OriginExternalId: l.Origin.ExternalID,
	}
}

// pbCheck converts a balance assertion check.
func pbCheck(c *ledgerapp.AssertionCheck) *viewsv1.AssertionCheck {
	return &viewsv1.AssertionCheck{
		Date: pbDate(c.Assertion.Date), Book: pbBook(c.Assertion.Book), Reported: c.Assertion.Balance.String(),
		Derived: c.Derived.String(), Status: viewsv1.AssertionStatus(enumNumber(int(c.Status))),
	}
}

// pbValuation converts a stored valuation of a holding.
func pbValuation(v *valuation.HoldingValuation) *viewsv1.Valuation {
	return &viewsv1.Valuation{
		Date: pbDate(v.Date), Instrument: pbInstrument(v.Instrument), Basis: viewsv1.ValuationBasis(enumNumber(int(v.Basis))),
		Source: v.Source, Value: pbMoney(v.Value), ValueReporting: pbMoney(v.PLN),
	}
}

// pbTotals converts totals per date.
func pbTotals(ts []views.DateTotal) []*viewsv1.DateTotal {
	out := make([]*viewsv1.DateTotal, 0, len(ts))
	for i := range ts {
		out = append(out, &viewsv1.DateTotal{Date: pbDate(ts[i].Date), ValueReporting: pbMoney(ts[i].Reporting), ByCurrency: pbMonies(ts[i].ByCurrency)})
	}
	return out
}

// pbStatus converts an account's valuation status; a never-valued account
// has no valued-through date.
func pbStatus(s *valuationapp.ValuationStatus) *viewsv1.ValuationStatus {
	return &viewsv1.ValuationStatus{AccountId: string(s.Account), ValuedThrough: pbDate(s.ValuedThrough), Pending: s.Pending, Running: s.Running}
}

// pbAllocation converts an allocation on a date with its unvalued holdings.
func pbAllocation(a *app.AllocationOn) *viewsv1.GetAllocationResponse {
	out := &viewsv1.GetAllocationResponse{
		Date: pbDate(a.Date), By: viewsv1.AllocationDimension(enumNumber(int(a.Allocation.By))), TotalReporting: pbMoney(a.Allocation.Total),
		Groups: make([]*viewsv1.AllocationGroup, 0, len(a.Allocation.Groups)), Unvalued: make([]*viewsv1.UnvaluedHolding, 0, len(a.Unvalued)),
	}
	for _, g := range a.Allocation.Groups {
		out.Groups = append(out.Groups, &viewsv1.AllocationGroup{Key: g.Key, ValueReporting: pbMoney(g.Reporting)})
	}
	for i := range a.Unvalued {
		h := &a.Unvalued[i]
		out.Unvalued = append(out.Unvalued, &viewsv1.UnvaluedHolding{AccountId: string(h.Account), Instrument: pbInstrument(h.Instrument), Quantity: h.Quantity.String()})
	}
	return out
}
