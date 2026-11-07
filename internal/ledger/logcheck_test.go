package ledger_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// storedVersion is one stored version of a transaction in a test log: its
// account, external ID and version, and its postings as "cash 100 PLN" or
// "equity -100 PLN"; none for a deletion. Its postings are stored as given,
// balanced or not, as a corrupted log could hold them.
type storedVersion struct {
	account  ledger.AccountID
	id       string
	version  int
	postings []string
}

// TestCheckLog checks the whole-log check: counts per account and in all,
// superseded versions, a version listed twice counted once, every
// unbalanced version named per currency, and the per-currency sums.
func TestCheckLog(t *testing.T) {
	t.Parallel()
	deposit := []string{"cash 100 PLN", "equity -100 PLN"}
	tests := []struct {
		name         string
		log          []storedVersion
		want         []string
		wantBalanced bool
	}{
		{name: "an empty log", want: []string{"all: 0 transactions, 0 versions, 0 superseded"}, wantBalanced: true},
		{
			name: "balanced versions, one superseded and one deletion",
			log: []storedVersion{
				{account: "acc-a", id: "dep-1", version: 1, postings: deposit},
				{account: "acc-a", id: "dep-1", version: 2, postings: []string{"cash 110 PLN", "equity -110 PLN"}},
				{account: "acc-b", id: "fx-1", version: 1, postings: []string{"cash 10 USD", "fx -10 USD", "cash -40 PLN", "fx 40 PLN"}},
				{account: "acc-b", id: "fee-1", version: 1, postings: []string{"fee 1 PLN", "cash -1 PLN"}},
				{account: "acc-b", id: "fee-1", version: 2},
			},
			want: []string{
				"acc-a: 1 transactions, 2 versions, 1 superseded",
				"acc-b: 2 transactions, 3 versions, 1 superseded",
				"all: 3 transactions, 5 versions, 2 superseded",
				"sum 0 PLN", "sum 0 USD",
				"logged acc-a test/dep-1 v1 deposit superseded=true",
				"logged acc-a test/dep-1 v2 deposit superseded=false",
				"logged acc-b test/fee-1 v1 deposit superseded=true",
				"logged acc-b test/fee-1 v2 deletion superseded=false",
				"logged acc-b test/fx-1 v1 deposit superseded=false",
			},
			wantBalanced: true,
		},
		{
			name: "a version listed twice counts once",
			log: []storedVersion{
				{account: "acc-a", id: "dep-1", version: 1, postings: deposit},
				{account: "acc-a", id: "dep-1", version: 1, postings: deposit},
			},
			want: []string{
				"acc-a: 1 transactions, 1 versions, 0 superseded", "all: 1 transactions, 1 versions, 0 superseded", "sum 0 PLN",
				"logged acc-a test/dep-1 v1 deposit superseded=false",
			},
			wantBalanced: true,
		},
		{
			name: "an unbalanced superseded version is named, per currency",
			log: []storedVersion{
				{account: "acc-a", id: "dep-1", version: 1, postings: []string{"cash 100 PLN", "equity -99.99 PLN", "cash 1 USD"}},
				{account: "acc-a", id: "dep-1", version: 2, postings: deposit},
			},
			want: []string{
				"acc-a: 1 transactions, 2 versions, 1 superseded",
				"all: 1 transactions, 2 versions, 1 superseded",
				"unbalanced acc-a test/dep-1 v1: 0.01 PLN", "unbalanced acc-a test/dep-1 v1: 1 USD",
				"sum 0.01 PLN", "sum 1 USD",
				"logged acc-a test/dep-1 v1 deposit superseded=true",
				"logged acc-a test/dep-1 v2 deposit superseded=false",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			log := make([]ledger.VersionedTransaction, 0, len(tt.log))
			for _, v := range tt.log {
				log = append(log, versionOf(t, v))
			}
			got := ledger.CheckLog(log)
			if diff := cmp.Diff(tt.want, describeLogCheck(&got)); diff != "" {
				t.Errorf("CheckLog (-want +got):\n%s", diff)
			}
			if got.Balanced() != tt.wantBalanced {
				t.Errorf("Balanced() = %t, want %t", got.Balanced(), tt.wantBalanced)
			}
		})
	}
}

// versionOf builds the stored version v: a deposit, or a deletion without
// postings, keyed under source "test".
func versionOf(t *testing.T, v storedVersion) ledger.VersionedTransaction {
	t.Helper()
	tx := ledger.Transaction{
		Key: ledger.TransactionKey{Source: "test", ExternalID: v.id}, Account: v.account,
		Kind: ledger.KindDeposit, Date: date(t, "2026-03-02"),
	}
	if len(v.postings) == 0 {
		tx.Kind = ledger.KindDeletion
	}
	for _, p := range v.postings {
		f := strings.Fields(p)
		m := money(t, f[1], f[2])
		var book ledger.Book
		switch f[0] {
		case "cash":
			book = ledger.CashBook(m.Currency())
		case "equity":
			book = ledger.EquityBook(m.Currency())
		case "fee":
			book = ledger.FeeBook(m.Currency())
		case "fx":
			book = ledger.Book{Kind: ledger.BookFX, Currency: m.Currency()}
		default:
			t.Fatalf("unknown ledger account %q", f[0])
		}
		tx.Postings = append(tx.Postings, ledger.Posting{Book: book, Amount: m})
	}
	return ledger.VersionedTransaction{Transaction: tx, Version: v.version}
}

// describeLogCheck writes c as lines: each account's counts, the whole
// log's, each unbalanced version, each currency's sum and each logged
// version.
func describeLogCheck(c *ledger.LogCheck) []string {
	counts := func(name string, tx, versions, superseded int) string {
		return fmt.Sprintf("%s: %d transactions, %d versions, %d superseded", name, tx, versions, superseded)
	}
	out := make([]string, 0, len(c.Accounts)+1+len(c.Unbalanced)+len(c.Sums)+len(c.Logged))
	for _, a := range c.Accounts {
		out = append(out, counts(string(a.Account), a.Transactions, a.Versions, a.Superseded))
	}
	out = append(out, counts("all", c.Transactions, c.Versions, c.Superseded))
	for _, u := range c.Unbalanced {
		out = append(out, fmt.Sprintf("unbalanced %s %s v%d: %s", u.Account, u.Key, u.Version, u.Sum))
	}
	for _, s := range c.Sums {
		out = append(out, "sum "+s.String())
	}
	for _, v := range c.Logged {
		out = append(out, fmt.Sprintf("logged %s %s v%d %s superseded=%t", v.Account, v.Key, v.Version, v.Kind, v.Superseded))
	}
	return out
}
