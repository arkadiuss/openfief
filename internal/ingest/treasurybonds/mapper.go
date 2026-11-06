package treasurybonds

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Institution names the treasury bond registry on every delivery.
const Institution = ingest.TreasuryBonds

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
	// ErrNoAccount reports entries without the account they were entered
	// on (D-118).
	ErrNoAccount = errors.New("no ledger account for the bond lots")
	// ErrDuplicateLot reports two lots of one series bought on one day in
	// one batch, added or deleted: they would share a key, so the owner
	// enters them as one.
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

// Mapper maps bond lots the owner entered into the ledger account the
// entries name (D-118), so one connector instance serves every bond
// account. The zero Mapper is ready to use.
type Mapper struct {
	// account is the account of the entries being mapped.
	account ledger.AccountID
}

// Map validates each lot of e, entered on e.Account, as of today
// (ingest.BondLot.Validate) and books it on its real purchase date (D-98):
// two transactions, a deposit of its cost against equity and the buy of
// its bonds for that cost, so the account's cash stays at zero, its
// history starts at its first purchase, and the money counts as having
// entered the portfolio then (D-61). Only an account seeded with opening
// lots before that rule has a go-live date in e: a lot bought on or before
// it is booked as the opening lot it was, on the go-live date and keeping
// its own purchase date and cost, so it keeps its key. Each deleted lot of
// e (D-100) becomes a deletion of each of its transactions, keyed and
// dated as adding it was, and of everything e.Booked says was booked for
// it since: its coupons, its redemption and their payout withdrawals, so
// their cash goes with it (D-154). Map also returns each series' terms once, in the
// order the series first appear. Any invalid lot, one series bought on one
// day twice in the batch, or one series with two different terms refuses
// the whole batch.
//
// The keys are "lot:<series>:<purchase date>" with ":deposit" and ":buy",
// or without them for an opening lot, under the source
// SourcePrefix plus the account ID, so each lot is idempotent per series and
// purchase date: entering it again unchanged changes nothing, and with
// another quantity, price or terms makes a new version that supersedes the
// old (D-18).
func (m Mapper) Map(e *ingest.BondLotEntries, today kernel.Date) (ingest.Mapping, []ingest.BondTerms, error) {
	m.account = e.Account
	if m.account == "" {
		return ingest.Mapping{}, nil, ErrNoAccount
	}
	var (
		out   ingest.Mapping
		terms []ingest.BondTerms
		seen  = map[ingest.BondLotKey]bool{}
		bySer = map[pl.Series]int{}
	)
	for i := range e.Lots {
		l := &e.Lots[i]
		if err := l.Validate(today); err != nil {
			return ingest.Mapping{}, nil, fmt.Errorf("lot %d: %w", i+1, err)
		}
		in, err := m.unique(seen, l.Key())
		if err != nil {
			return ingest.Mapping{}, nil, fmt.Errorf("lot %d: %w", i+1, err)
		}
		if terms, err = addTerms(terms, bySer, in, &l.Terms); err != nil {
			return ingest.Mapping{}, nil, fmt.Errorf("lot %d: %w", i+1, err)
		}
		txs, err := m.lotTransactions(in, l, e.GoLive)
		if err != nil {
			return ingest.Mapping{}, nil, fmt.Errorf("lot %d: %w", i+1, err)
		}
		out.Transactions = append(out.Transactions, txs...)
	}
	deleted, err := m.mapDeleted(e, seen)
	if err != nil {
		return ingest.Mapping{}, nil, err
	}
	out.Transactions = append(out.Transactions, deleted...)
	return out, terms, nil
}

// mapDeleted returns the deletions of e's deleted lots, recording each in
// seen.
func (m Mapper) mapDeleted(e *ingest.BondLotEntries, seen map[ingest.BondLotKey]bool) ([]ledger.Transaction, error) {
	var out []ledger.Transaction
	for i, k := range e.Deleted {
		if err := k.Validate(); err != nil {
			return nil, fmt.Errorf("deleted lot %d: %w", i+1, err)
		}
		in, err := m.unique(seen, k)
		if err != nil {
			return nil, fmt.Errorf("deleted lot %d: %w", i+1, err)
		}
		txs, err := m.deletions(in, k, e.GoLive, e.Booked)
		if err != nil {
			return nil, fmt.Errorf("deleted lot %d: %w", i+1, err)
		}
		out = append(out, txs...)
	}
	return out, nil
}

// unique records k in seen and returns its series' instrument, or fails
// with ErrDuplicateLot when the batch named k already.
func (m Mapper) unique(seen map[ingest.BondLotKey]bool, k ingest.BondLotKey) (ledger.Instrument, error) {
	if seen[k] {
		return ledger.Instrument{}, fmt.Errorf("%w: %s on %s", ErrDuplicateLot, k.Series, k.Purchased)
	}
	seen[k] = true
	return Instrument(k.Series)
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

// lotID returns the base external ID of the lot k: "lot:<series>:<date>".
func lotID(k ingest.BondLotKey) string {
	return "lot:" + k.Series.String() + ":" + k.Purchased.String()
}

// header returns the header of the transaction with external ID id on day.
func (m Mapper) header(id string, day kernel.Date) ledger.TransactionHeader {
	return ledger.TransactionHeader{
		Key:     ledger.TransactionKey{Source: SourcePrefix + string(m.account), ExternalID: id},
		Account: m.account, Date: day,
	}
}

// opening reports whether a lot bought on purchased is an opening lot: only
// in an account seeded with opening lots, when bought on or before its
// go-live date goLive (D-98).
func opening(purchased, goLive kernel.Date) bool {
	return !goLive.IsZero() && !purchased.After(goLive)
}

// lotTransactions returns the transactions of lot l: a deposit and a buy
// on its purchase date, or the opening lot on goLive it was booked as
// (opening).
func (m Mapper) lotTransactions(in ledger.Instrument, l *ingest.BondLot, goLive kernel.Date) ([]ledger.Transaction, error) {
	id, cost := lotID(l.Key()), l.Cost()
	if opening(l.Purchased, goLive) {
		lot, err := ledger.NewOpeningLot(m.header(id, goLive), in, l.Quantity, cost, l.Purchased)
		if err != nil {
			return nil, fmt.Errorf("opening lot of %s: %w", in.Symbol, err)
		}
		lot.Description = "Treasury bonds " + in.Symbol + " bought before go-live"
		return []ledger.Transaction{lot}, nil
	}
	deposit, err := ledger.NewDeposit(m.header(id+":deposit", l.Purchased), cost)
	if err != nil {
		return nil, fmt.Errorf("deposit for %s: %w", in.Symbol, err)
	}
	deposit.Description = "Cash for treasury bonds " + in.Symbol
	buy, err := ledger.NewBuy(m.header(id+":buy", l.Purchased), in, l.Quantity, cost, kernel.Money{})
	if err != nil {
		return nil, fmt.Errorf("buy of %s: %w", in.Symbol, err)
	}
	buy.Description = "Treasury bonds " + in.Symbol
	return []ledger.Transaction{deposit, buy}, nil
}

// bookedFor returns the headers of booked that were booked for lot k in
// m's account under SourcePrefix, other than those of lot itself: every
// key that extends the lot's own, "lot:<series>:<purchase date>:", such as
// a coupon ("...:coupon:<due date>"), a redemption ("...:redemption") and
// their payout withdrawals (":payout"), each with its own date.
func (m Mapper) bookedFor(k ingest.BondLotKey, booked, lot []ledger.TransactionHeader) []ledger.TransactionHeader {
	source, prefix := SourcePrefix+string(m.account), lotID(k)+":"
	var out []ledger.TransactionHeader
	for _, b := range booked {
		if b.Key.Source != source || !strings.HasPrefix(b.Key.ExternalID, prefix) ||
			slices.ContainsFunc(lot, func(h ledger.TransactionHeader) bool { return h.Key == b.Key }) {
			continue
		}
		out = append(out, m.header(b.Key.ExternalID, b.Date))
	}
	return out
}

// deletions returns the deletion of each transaction lot k was booked as,
// keyed and dated as lotTransactions books it, then of each of booked, in
// order, that was booked for the lot (bookedFor).
func (m Mapper) deletions(in ledger.Instrument, k ingest.BondLotKey, goLive kernel.Date, booked []ledger.TransactionHeader) ([]ledger.Transaction, error) {
	id := lotID(k)
	lot := []ledger.TransactionHeader{m.header(id+":deposit", k.Purchased), m.header(id+":buy", k.Purchased)}
	if opening(k.Purchased, goLive) {
		lot = []ledger.TransactionHeader{m.header(id, goLive)}
	}
	headers := slices.Concat(lot, m.bookedFor(k, booked, lot))
	out := make([]ledger.Transaction, 0, len(headers))
	for _, h := range headers {
		t, err := ledger.NewDeletion(h)
		if err != nil {
			return nil, fmt.Errorf("deletion of %s: %w", in.Symbol, err)
		}
		t.Description = "Deleted treasury bonds " + in.Symbol
		out = append(out, t)
	}
	return out, nil
}
