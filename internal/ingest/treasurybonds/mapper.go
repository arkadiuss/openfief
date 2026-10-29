package treasurybonds

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Institution names the treasury bond registry on every delivery.
const Institution ingest.InstitutionID = "treasury-bonds"

// SourcePrefix starts the source of every transaction key this connector
// writes; the ledger account ID follows it.
const SourcePrefix = "treasury-bonds:"

// isinCountry is the country code of the ISINs Instrument derives: Polish,
// so the views count the holding as a treasury bond.
const isinCountry = "PL"

// isinFiller pads a seven-character series code to an ISIN's nine-character
// national number.
const isinFiller = "00"

// checkDigits is how many check digits Instrument tries.
const checkDigits = 10

// Errors mapping lots.
var (
	// ErrNoAccount reports a mapper without the ledger account its lots go
	// into.
	ErrNoAccount = errors.New("no ledger account for the bond lots")
	// ErrDuplicateLot reports two lots of one series bought on one day in
	// one batch: they would share a key, so the owner enters them as one.
	ErrDuplicateLot = errors.New("two lots of one series bought on one day")
	// ErrConflictingTerms reports two lots of one series in one batch with
	// different series terms.
	ErrConflictingTerms = errors.New("one series with two different terms")
)

// Instrument returns the instrument of bond series s: its venue symbol is
// the series code, such as EDO0536, and its ISIN is derived from the code
// as "PL", the code, "00" and the check digit that makes it valid, such as
// PLEDO0536000. The owner enters only the series code (D-9), and the
// issuer's offer and interest tables name series by code, so the ISIN is
// derived, not the one the registry may have assigned; the same series
// always gets the same instrument.
func Instrument(s pl.Series) (ledger.Instrument, error) {
	body := isinCountry + s.String() + isinFiller
	for d := range checkDigits {
		if isin, err := ledger.ParseISIN(body + strconv.Itoa(d)); err == nil {
			return ledger.Instrument{ISIN: isin, Symbol: s.String()}, nil
		}
	}
	return ledger.Instrument{}, fmt.Errorf("%w: no check digit for %s", ledger.ErrInvalidISIN, body)
}

// Mapper maps bond lots the owner entered into the ledger account Account.
type Mapper struct {
	Account ledger.AccountID
}

// Map validates each lot as of today (ingest.BondLot.Validate) and maps it
// to two transactions on its purchase date, so that the account's cash
// stays at zero and the money counts as having entered the portfolio then
// (D-61): a deposit of its cost against equity, and the buy of its bonds for
// that cost. It also returns each series' terms once, in the order the
// series first appear. Any invalid lot, two lots of one series bought on one
// day, or one series with two different terms refuses the whole batch.
//
// The keys are "lot:<series>:<purchase date>:deposit" and ":buy", under the
// source SourcePrefix plus the account ID. Entering a lot of the same series
// and day again with another quantity makes a new version of both, which
// supersedes the old (D-18).
func (m Mapper) Map(lots []ingest.BondLot, today kernel.Date) (ingest.Mapping, []ingest.BondTerms, error) {
	if m.Account == "" {
		return ingest.Mapping{}, nil, ErrNoAccount
	}
	var (
		out   ingest.Mapping
		terms []ingest.BondTerms
		seen  = map[string]bool{}
		bySer = map[pl.Series]int{}
	)
	for i := range lots {
		l := &lots[i]
		if err := l.Validate(today); err != nil {
			return ingest.Mapping{}, nil, fmt.Errorf("lot %d: %w", i+1, err)
		}
		id := "lot:" + l.Series().String() + ":" + l.Purchased.String()
		if seen[id] {
			return ingest.Mapping{}, nil, fmt.Errorf("lot %d: %w: %s on %s", i+1, ErrDuplicateLot, l.Series(), l.Purchased)
		}
		seen[id] = true
		in, err := Instrument(l.Series())
		if err != nil {
			return ingest.Mapping{}, nil, fmt.Errorf("lot %d: %w", i+1, err)
		}
		if terms, err = addTerms(terms, bySer, in, &l.Terms); err != nil {
			return ingest.Mapping{}, nil, fmt.Errorf("lot %d: %w", i+1, err)
		}
		txs, err := m.lotTransactions(id, in, l)
		if err != nil {
			return ingest.Mapping{}, nil, fmt.Errorf("lot %d: %w", i+1, err)
		}
		out.Transactions = append(out.Transactions, txs...)
	}
	return out, terms, nil
}

// addTerms appends t for in to terms unless its series is already there,
// in which case the terms must be equal.
func addTerms(terms []ingest.BondTerms, bySeries map[pl.Series]int, in ledger.Instrument, t *pl.Terms) ([]ingest.BondTerms, error) {
	if j, ok := bySeries[t.Series]; ok {
		if !sameTerms(&terms[j].Terms, t) {
			return nil, fmt.Errorf("%w: %s", ErrConflictingTerms, t.Series)
		}
		return terms, nil
	}
	bySeries[t.Series] = len(terms)
	return append(terms, ingest.BondTerms{Instrument: in, Terms: *t}), nil
}

// sameTerms reports whether a and b are the same terms of the same series.
func sameTerms(a, b *pl.Terms) bool {
	return a.Series == b.Series && a.FirstPeriodRate.Equal(b.FirstPeriodRate) &&
		a.Margin.Equal(b.Margin) && a.EarlyRedemptionFee.Equal(b.EarlyRedemptionFee)
}

// lotTransactions returns the deposit and the buy of lot l keyed under id.
func (m Mapper) lotTransactions(id string, in ledger.Instrument, l *ingest.BondLot) ([]ledger.Transaction, error) {
	header := func(suffix string) ledger.TransactionHeader {
		return ledger.TransactionHeader{
			Key:     ledger.TransactionKey{Source: SourcePrefix + string(m.Account), ExternalID: id + ":" + suffix},
			Account: m.Account, Date: l.Purchased,
		}
	}
	cost := l.Cost()
	deposit, err := ledger.NewDeposit(header("deposit"), cost)
	if err != nil {
		return nil, fmt.Errorf("deposit for %s: %w", in.Symbol, err)
	}
	deposit.Description = "Cash for treasury bonds " + in.Symbol
	buy, err := ledger.NewBuy(header("buy"), in, l.Quantity, cost, kernel.Money{})
	if err != nil {
		return nil, fmt.Errorf("buy of %s: %w", in.Symbol, err)
	}
	buy.Description = "Treasury bonds " + in.Symbol
	return []ledger.Transaction{deposit, buy}, nil
}
