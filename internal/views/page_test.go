package views_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// txAt is a transaction keyed test/id, dated day.
func txAt(t *testing.T, day, id string) ledger.Transaction {
	t.Helper()
	return ledger.Transaction{Key: ledger.TransactionKey{Source: "test", ExternalID: id}, Account: "acc-1", Kind: ledger.KindDeposit, Date: date(t, day)}
}

// keys returns the external IDs of txs.
func keys(txs []ledger.Transaction) []string {
	out := make([]string, 0, len(txs))
	for i := range txs {
		out = append(out, txs[i].Key.ExternalID)
	}
	return out
}

// pageFixture returns five transactions out of order: the page is in date
// then key order.
func pageFixture(t *testing.T) []ledger.Transaction {
	t.Helper()
	return []ledger.Transaction{
		txAt(t, "2026-01-03", "e"), txAt(t, "2026-01-01", "b"), txAt(t, "2026-01-01", "a"),
		txAt(t, "2026-01-02", "d"), txAt(t, "2026-01-02", "c"),
	}
}

// cursorAt returns a cursor at the transaction of txs with external ID id,
// or nil when none has it.
func cursorAt(txs []ledger.Transaction, id string) *views.TransactionCursor {
	for i := range txs {
		if txs[i].Key.ExternalID == id {
			return &views.TransactionCursor{Date: txs[i].Date, Key: txs[i].Key}
		}
	}
	return nil
}

// nextID returns the external ID the page's cursor points at, or "" for none.
func nextID(p *views.TransactionPage) string {
	if p.Next == nil {
		return ""
	}
	return p.Next.Key.ExternalID
}

func TestPageTransactions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		size     int
		after    string // the external ID of the cursor's transaction; "" for the first page
		want     []string
		wantNext string // "" when this is the last page
	}{
		{name: "first page", size: 2, want: []string{"a", "b"}, wantNext: "b"},
		{name: "page after a cursor", size: 2, after: "b", want: []string{"c", "d"}, wantNext: "d"},
		{name: "last page has no cursor", size: 2, after: "d", want: []string{"e"}},
		{name: "a page that fits everything", size: 5, want: []string{"a", "b", "c", "d", "e"}},
		{name: "after the last transaction is empty", size: 2, after: "e", want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			txs := pageFixture(t)
			got, err := views.PageTransactions(txs, cursorAt(txs, tt.after), tt.size)
			if err != nil {
				t.Fatalf("PageTransactions() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, keys(got.Transactions)); diff != "" {
				t.Errorf("PageTransactions() mismatch (-want +got):\n%s", diff)
			}
			if n := nextID(&got); n != tt.wantNext {
				t.Errorf("PageTransactions() next = %q, want %q", n, tt.wantNext)
			}
			if txs[0].Key.ExternalID != "e" {
				t.Error("PageTransactions() reordered its input")
			}
		})
	}
}

func TestPageTransactionsSize(t *testing.T) {
	t.Parallel()
	for _, size := range []int{0, -1, views.MaxPageSize + 1} {
		if _, err := views.PageTransactions(nil, nil, size); !errors.Is(err, views.ErrInvalidPageSize) {
			t.Errorf("PageTransactions(size %d) error = %v, want %v", size, err, views.ErrInvalidPageSize)
		}
	}
	if _, err := views.PageTransactions(nil, nil, views.MaxPageSize); err != nil {
		t.Errorf("PageTransactions(size %d) error = %v", views.MaxPageSize, err)
	}
}

func TestTransactionCursor(t *testing.T) {
	t.Parallel()
	c := views.TransactionCursor{Date: date(t, "2026-01-02"), Key: ledger.TransactionKey{Source: "ibkr-flex", ExternalID: "T/1?&"}}
	got, err := views.ParseTransactionCursor(c.Encode())
	if err != nil {
		t.Fatalf("ParseTransactionCursor(Encode()) error = %v", err)
	}
	if diff := cmp.Diff(c, got); diff != "" {
		t.Errorf("cursor round trip mismatch (-want +got):\n%s", diff)
	}
	for _, bad := range []string{"", "not base64!", "bm90IGpzb24", "eyJkIjoiMjAyNi0xMy0wMSIsInMiOiJ4IiwiZSI6InkifQ", "eyJkIjoiMjAyNi0wMS0wMSJ9"} {
		if _, err := views.ParseTransactionCursor(bad); !errors.Is(err, views.ErrInvalidCursor) {
			t.Errorf("ParseTransactionCursor(%q) error = %v, want %v", bad, err, views.ErrInvalidCursor)
		}
	}
}
