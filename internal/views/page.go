package views

import (
	"cmp"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrInvalidCursor reports a transaction cursor that TransactionCursor.Encode
// did not write.
var ErrInvalidCursor = errors.New("invalid transaction cursor")

// ErrInvalidPageSize reports a page size outside 1..MaxPageSize.
var ErrInvalidPageSize = errors.New("invalid page size")

// The page sizes of a transaction listing.
const (
	// DefaultPageSize is the page size when none is asked for.
	DefaultPageSize = 100
	// MaxPageSize is the largest page size that may be asked for.
	MaxPageSize = 500
)

// TransactionCursor marks the last transaction of a page: the next page
// starts after it in date then key order.
type TransactionCursor struct {
	Date kernel.Date
	Key  ledger.TransactionKey
}

// cursorJSON is the encoded form of a TransactionCursor.
type cursorJSON struct {
	Date       string `json:"d"`
	Source     string `json:"s"`
	ExternalID string `json:"e"`
}

// Encode returns c as an opaque, URL-safe string that ParseTransactionCursor
// reads back.
func (c TransactionCursor) Encode() string {
	b, err := json.Marshal(cursorJSON{Date: c.Date.String(), Source: c.Key.Source, ExternalID: c.Key.ExternalID})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// ParseTransactionCursor reads a cursor that Encode wrote.
func ParseTransactionCursor(s string) (TransactionCursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return TransactionCursor{}, fmt.Errorf("%w: %w", ErrInvalidCursor, err)
	}
	var j cursorJSON
	if jsonErr := json.Unmarshal(b, &j); jsonErr != nil {
		return TransactionCursor{}, fmt.Errorf("%w: %w", ErrInvalidCursor, jsonErr)
	}
	d, err := kernel.ParseDate(j.Date)
	if err != nil {
		return TransactionCursor{}, fmt.Errorf("%w: %w", ErrInvalidCursor, err)
	}
	if j.Source == "" || j.ExternalID == "" {
		return TransactionCursor{}, fmt.Errorf("%w: no transaction key", ErrInvalidCursor)
	}
	return TransactionCursor{Date: d, Key: ledger.TransactionKey{Source: j.Source, ExternalID: j.ExternalID}}, nil
}

// TransactionPage is one page of transactions. Next is set when more follow:
// pass it to PageTransactions to get the next page.
type TransactionPage struct {
	Transactions []ledger.Transaction
	Next         *TransactionCursor
	// Origins holds each transaction's origin, in the same order, once
	// Installed.Origins set it (D-58); nil before.
	Origins []Origin
}

// PageTransactions returns at most size of txs in date then key order,
// starting after the cursor after, or from the first when after is nil. A
// size outside 1..MaxPageSize is an error. txs is not changed.
func PageTransactions(txs []ledger.Transaction, after *TransactionCursor, size int) (TransactionPage, error) {
	if size < 1 || size > MaxPageSize {
		return TransactionPage{}, fmt.Errorf("%w: %d, want 1 to %d", ErrInvalidPageSize, size, MaxPageSize)
	}
	sorted := slices.Clone(txs)
	slices.SortStableFunc(sorted, func(a, b ledger.Transaction) int { return compareAt(a.Date, a.Key, b.Date, b.Key) })
	start := 0
	if after != nil {
		start = len(sorted)
		for i := range sorted {
			if compareAt(sorted[i].Date, sorted[i].Key, after.Date, after.Key) > 0 {
				start = i
				break
			}
		}
	}
	end := min(start+size, len(sorted))
	page := TransactionPage{Transactions: sorted[start:end]}
	if end < len(sorted) {
		last := &sorted[end-1]
		page.Next = &TransactionCursor{Date: last.Date, Key: last.Key}
	}
	return page, nil
}

// compareAt orders transactions by date, then source, then external ID.
func compareAt(ad kernel.Date, ak ledger.TransactionKey, bd kernel.Date, bk ledger.TransactionKey) int {
	if c := ad.Compare(bd); c != 0 {
		return c
	}
	if c := cmp.Compare(ak.Source, bk.Source); c != 0 {
		return c
	}
	return cmp.Compare(ak.ExternalID, bk.ExternalID)
}
