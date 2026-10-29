package ingest_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestBrokerAccountIDValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		id      ingest.BrokerAccountID
		wantErr error
	}{
		{name: "synthetic Flex account", id: "U0000001"},
		{name: "64 characters", id: ingest.BrokerAccountID(strings.Repeat("A", 64))},
		{name: "empty", id: "", wantErr: ingest.ErrInvalidBrokerAccount},
		{name: "65 characters", id: ingest.BrokerAccountID(strings.Repeat("A", 65)), wantErr: ingest.ErrInvalidBrokerAccount},
		{name: "space", id: "U000 0001", wantErr: ingest.ErrInvalidBrokerAccount},
		{name: "control character", id: "U0000001\n", wantErr: ingest.ErrInvalidBrokerAccount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.wantErr, tt.id.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate (-want +got):\n%s", diff)
			}
		})
	}
}

// brokerTx is a transaction of kind with external ID id, keyed by the
// connector's source that names no account, in no account yet.
func brokerTx(kind ledger.TransactionKind, id string) ledger.Transaction {
	return ledger.Transaction{Key: ledger.TransactionKey{Source: "ibkr-flex", ExternalID: id}, Kind: kind}
}

// booked is brokerTx as resolved into acc-main.
func booked(kind ledger.TransactionKind, id string) ledger.Transaction {
	t := brokerTx(kind, id)
	t.Account = "acc-main"
	t.Key.Source = "ibkr-flex:acc-main"
	return t
}

func TestResolveBrokerAccounts(t *testing.T) {
	t.Parallel()
	const main, side ingest.BrokerAccountID = "U0000001", "U0000002"
	m := ingest.Mapping{
		Transactions: []ledger.Transaction{{Account: "acc-direct"}},
		Problems:     []ingest.Problem{ingest.Skipped("row#1", "summary row")},
		ByBrokerAccount: []ingest.BrokerEntries{
			{
				Account:      main,
				Transactions: []ledger.Transaction{brokerTx(ledger.KindOpeningLot, "lot-1"), brokerTx(ledger.KindBuy, "t-1")},
				Assertions:   []ledger.BalanceAssertion{{Book: ledger.Book{Kind: ledger.BookCash}}},
			},
			{Account: side, Transactions: []ledger.Transaction{brokerTx(ledger.KindBuy, "t-2")}},
		},
	}
	mainAssertion := ledger.BalanceAssertion{Account: "acc-main", Book: ledger.Book{Kind: ledger.BookCash}}
	mapped := func(seeded bool) []ingest.BrokerAccount {
		return []ingest.BrokerAccount{{Instance: "ibkr-main", ID: main, Account: "acc-main", Seeded: seeded}, {Instance: "ibkr-main", ID: side}}
	}
	direct := m.Transactions[0]
	tests := []struct {
		name  string
		known []ingest.BrokerAccount
		mode  ingest.Mode
		want  ingest.Resolution
	}{
		{
			name: "nothing known: every broker account unmapped, nothing of theirs booked",
			mode: ingest.ModeTransactions,
			want: ingest.Resolution{
				Mapping:  ingest.Mapping{Transactions: []ledger.Transaction{direct}, Problems: m.Problems},
				Unmapped: []ingest.BrokerAccountID{main, side},
			},
		},
		{
			name:  "a mapped account is booked with its account in the key; the other stays unmapped",
			known: mapped(false), mode: ingest.ModeTransactions,
			want: ingest.Resolution{
				Mapping: ingest.Mapping{
					Transactions: []ledger.Transaction{direct, booked(ledger.KindOpeningLot, "lot-1"), booked(ledger.KindBuy, "t-1")},
					Assertions:   []ledger.BalanceAssertion{mainAssertion}, Problems: m.Problems,
				},
				Unmapped: []ingest.BrokerAccountID{side},
			},
		},
		{
			name:  "an opening batch seeds a mapped account not seeded yet",
			known: mapped(false), mode: ingest.ModeOpening,
			want: ingest.Resolution{
				Mapping: ingest.Mapping{
					Transactions: []ledger.Transaction{direct, booked(ledger.KindOpeningLot, "lot-1"), booked(ledger.KindBuy, "t-1")},
					Assertions:   []ledger.BalanceAssertion{mainAssertion}, Problems: m.Problems,
				},
				Unmapped: []ingest.BrokerAccountID{side},
				Seeding:  []ingest.BrokerAccountID{main},
			},
		},
		{
			name:  "an opening batch leaves out the seeding of an account already seeded, keeps its assertions",
			known: mapped(true), mode: ingest.ModeOpening,
			want: ingest.Resolution{
				Mapping: ingest.Mapping{
					Transactions: []ledger.Transaction{direct, booked(ledger.KindBuy, "t-1")},
					Assertions:   []ledger.BalanceAssertion{mainAssertion}, Problems: m.Problems,
				},
				Unmapped: []ingest.BrokerAccountID{side},
			},
		},
		{
			name:  "a mapped account the delivery doesn't name is still seeded by an opening batch",
			known: append(mapped(false), ingest.BrokerAccount{ID: "U0000003", Account: "acc-closed"}), mode: ingest.ModeOpening,
			want: ingest.Resolution{
				Mapping: ingest.Mapping{
					Transactions: []ledger.Transaction{direct, booked(ledger.KindOpeningLot, "lot-1"), booked(ledger.KindBuy, "t-1")},
					Assertions:   []ledger.BalanceAssertion{mainAssertion}, Problems: m.Problems,
				},
				Unmapped: []ingest.BrokerAccountID{side},
				Seeding:  []ingest.BrokerAccountID{main, "U0000003"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ingest.ResolveBrokerAccounts(&m, tt.known, tt.mode)
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ResolveBrokerAccounts (-want +got):\n%s", diff)
			}
		})
	}
	if m.ByBrokerAccount[0].Transactions[0].Account != "" {
		t.Error("ResolveBrokerAccounts changed the mapping it was given")
	}
}

func TestReportedBrokerAccounts(t *testing.T) {
	t.Parallel()
	m := ingest.Mapping{
		BrokerAccounts:  []ingest.BrokerAccountID{"U0000002", "U0000001"},
		ByBrokerAccount: []ingest.BrokerEntries{{Account: "U0000001"}, {Account: "U0000003"}},
	}
	want := []ingest.BrokerAccountID{"U0000002", "U0000001", "U0000003"}
	if diff := cmp.Diff(want, ingest.ReportedBrokerAccounts(&m)); diff != "" {
		t.Errorf("ReportedBrokerAccounts (-want +got):\n%s", diff)
	}
}
