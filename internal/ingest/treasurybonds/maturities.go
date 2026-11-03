package treasurybonds

import (
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// purchasePricePlaces is how many decimal places a lot's price per bond,
// its cost over its quantity, is taken to: grosze, as the valuation takes
// it.
const purchasePricePlaces = 2

// DueMaturity is an open lot of a held account that reached maturity: the
// account, its tax wrapper and payout setting, the lot and its series
// terms, and the day it matured.
type DueMaturity struct {
	Account ledger.AccountID
	Wrapper string
	Payout  ingest.PayoutSetting
	Lot     ledger.Lot
	Terms   pl.Terms
	Due     kernel.Date
}

// DueMaturities returns every open lot of a bond series held in accounts
// whose maturity (pl.Maturity) fell on or before today (D-135), in account
// and lot order. In an account seeded with opening lots, a lot that
// matured on or before its go-live is history the opening lots hold. A lot whose instrument is no bond series is
// left out, and so is one without terms or whose purchase date doesn't
// match its series: DueCoupons reports those as problems.
func DueMaturities(accounts []ingest.HeldAccount, terms []ingest.BondTerms, today kernel.Date) []DueMaturity {
	var out []DueMaturity
	for i := range accounts {
		a := &accounts[i]
		for j := range a.Lots {
			lot := &a.Lots[j]
			s, err := pl.ParseSeries(lot.Instrument.Symbol)
			if err != nil {
				continue
			}
			t, ok := termsOf(terms, lot.Instrument)
			if !ok {
				continue
			}
			due, err := pl.Maturity(s, lot.Opened)
			if err != nil || (!a.GoLive.IsZero() && !due.After(a.GoLive)) || due.After(today) {
				continue
			}
			out = append(out, DueMaturity{Account: a.Account, Wrapper: a.Wrapper, Payout: a.Payout, Lot: *lot, Terms: t, Due: due})
		}
	}
	return out
}

// MaturityCPIYears returns the years of CPI the matured lots' values need,
// in order, and whether any needs the NBP reference rate: the reading of
// every indexed period after the first (valuation.LotIndices).
func MaturityCPIYears(due []DueMaturity) ([]int, bool) {
	var (
		years []int
		rate  bool
	)
	for i := range due {
		m := &due[i]
		st, err := m.Terms.Series.Type.Structure()
		if err != nil {
			continue
		}
		switch st.Indexation {
		case pl.Inflation:
			for n := 2; n <= st.TenorMonths/st.PeriodMonths; n++ {
				start, err := pl.PeriodStart(m.Terms.Series.Type, m.Lot.Opened, n)
				if err != nil {
					continue
				}
				y, _ := pl.CPIMonth(start)
				years = append(years, y)
			}
		case pl.ReferenceRate:
			rate = true
		case pl.Fixed:
		}
	}
	slices.Sort(years)
	return slices.Compact(years), rate
}

// BookMaturities books each matured lot as a redemption (D-135) in its
// account: its bonds leave at the redemption value the Polish pack gives
// them at maturity (pl.ValueLot), and the Belka tax its wrapper withholds
// from their interest and purchase discount (pl.BelkaTax, D-100) goes to
// withholding tax, on the day it matured. When the account's payout
// setting pays out what falls due that day, a withdrawal of the cash
// received follows on the same day (PayoutWithdrawal). Each is keyed
// "lot:<series>:<purchase date>:redemption" under the source SourcePrefix
// plus the account, as the lot itself, so it is idempotent per lot:
// delivering it again on every sync until the lot is gone changes nothing.
// A redemption whose index readings are missing waits as a problem, and so
// does one in an account with a wrapper the pack doesn't know.
func BookMaturities(due []DueMaturity, readings []valuation.IndexReading) (ingest.Mapping, error) {
	return bookPayouts(due, func(m *DueMaturity) payout {
		t, err := bookMaturity(m, readings)
		return payout{tx: t, err: err, setting: m.Payout, ref: maturityRef(m), waiting: "its periods' index readings are not published yet"}
	})
}

// maturityRef names a matured lot in a problem.
func maturityRef(m *DueMaturity) string {
	return lotRef(m.Account, m.Terms.Series, m.Lot.Opened) + " maturity " + m.Due.String()
}

// bookMaturity returns the redemption transaction of m.
func bookMaturity(m *DueMaturity, readings []valuation.IndexReading) (ledger.Transaction, error) {
	s, purchased := m.Terms.Series, m.Lot.Opened
	idx, err := valuation.LotIndices(s, purchased, m.Due, readings)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("maturity of %s: %w", s, err)
	}
	wrapper, err := pl.ParseTaxWrapper(m.Wrapper)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("maturity of %s: %w", s, err)
	}
	price, err := m.Lot.Cost.Amount().Quo(m.Lot.Quantity.Decimal(), purchasePricePlaces, kernel.RoundHalfEven)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("maturity of %s: price per bond: %w", s, err)
	}
	p := pl.Purchase{Date: purchased, Price: kernel.NewMoney(price, m.Lot.Cost.Currency())}
	v, err := pl.ValueLot(&m.Terms, idx, m.Lot.Quantity, p, wrapper, m.Due)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("maturity of %s: %w", s, err)
	}
	taxed, err := v.Interest.Add(v.Discount)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("maturity of %s: %w", s, err)
	}
	tax, err := pl.BelkaTax(taxed, wrapper)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("maturity of %s: %w", s, err)
	}
	mp := Mapper{account: m.Account}
	key := ingest.BondLotKey{Series: s, Purchased: purchased}
	t, err := ledger.NewRedemption(mp.header(lotID(key)+":redemption", m.Due), m.Lot.Instrument, m.Lot.Quantity, v.Gross, tax)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("maturity of %s: %w", s, err)
	}
	t.Description = fmt.Sprintf("Treasury bonds %s redeemed at maturity", s)
	return t, nil
}

// PayoutWithdrawal returns the withdrawal that pays out t, a coupon's
// interest or a maturity's redemption, when payout pays out what falls due
// on t's date (D-135): the cash t brought into the account, withdrawn on
// the same day and keyed as t plus ":payout", so the account's cash
// matches the registry while t still counts as return. It reports false
// when the payout stays as cash, or t brought no cash.
func PayoutWithdrawal(t *ledger.Transaction, payout ingest.PayoutSetting) (ledger.Transaction, bool, error) {
	if !payout.PaidOut(t.Date) {
		return ledger.Transaction{}, false, nil
	}
	var cash kernel.Money
	for _, p := range t.Postings {
		if p.Book.Kind != ledger.BookCash {
			continue
		}
		if cash.Currency().IsZero() {
			cash = p.Amount
			continue
		}
		sum, err := cash.Add(p.Amount)
		if err != nil {
			return ledger.Transaction{}, false, fmt.Errorf("payout of %s: %w", t.Key, err)
		}
		cash = sum
	}
	if cash.Sign() <= 0 {
		return ledger.Transaction{}, false, nil
	}
	h := ledger.TransactionHeader{
		Key:     ledger.TransactionKey{Source: t.Key.Source, ExternalID: t.Key.ExternalID + ingest.PayoutSuffix},
		Account: t.Account, Date: t.Date,
	}
	w, err := ledger.NewWithdrawal(h, cash)
	if err != nil {
		return ledger.Transaction{}, false, fmt.Errorf("payout of %s: %w", t.Key, err)
	}
	w.Description = "Paid out: " + t.Description
	return w, true, nil
}
