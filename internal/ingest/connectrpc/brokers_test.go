package connectrpc

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

func TestDeliveryFromByBrokerAccount(t *testing.T) {
	t.Parallel()
	d := testDelivery()
	tx := d.GetTransactions()[0]
	direct := &connectorv1.Transaction{
		Key: &connectorv1.TransactionKey{Source: "acc-1", ExternalId: "T0"}, AccountId: "acc-1",
		Kind: tx.GetKind(), Date: tx.GetDate(), Postings: tx.GetPostings(),
	}
	tx.AccountId, tx.BrokerAccount, tx.Key.Source = "", "U0000002", "ibkr-flex"
	d.Transactions = append(d.Transactions, direct)
	d.BalanceAssertions[0].AccountId, d.BalanceAssertions[0].BrokerAccount = "", "U0000002"
	d.BrokerAccounts = []*connectorv1.BrokerAccount{{Id: "U0000001"}, {Id: "U0000002"}}
	got, err := deliveryFrom(d, "b-1", "")
	if err != nil {
		t.Fatalf("deliveryFrom: %v", err)
	}
	checkSplit(t, &got.Mapping)
}

// checkSplit fails t unless m holds the transaction of acc-1 directly and
// U0000002's transaction and assertion apart, in no account yet, with
// both listed broker accounts reported.
func checkSplit(t *testing.T, m *ingest.Mapping) {
	t.Helper()
	if n := len(m.Transactions); n != 1 || m.Transactions[0].Account != "acc-1" || len(m.Assertions) != 0 {
		t.Errorf("direct entries = %+v, %+v; want only the transaction in acc-1", m.Transactions, m.Assertions)
	}
	if diff := cmp.Diff([]ingest.BrokerAccountID{"U0000001", "U0000002"}, m.BrokerAccounts); diff != "" {
		t.Errorf("BrokerAccounts (-want +got):\n%s", diff)
	}
	// entry is what each broker account's entries come to: its ID, its
	// transactions' keys and accounts, and its assertions' accounts.
	type entry struct {
		Account          ingest.BrokerAccountID
		Keys             []ledger.TransactionKey
		TxAccounts       []ledger.AccountID
		AssertionAccount []ledger.AccountID
	}
	got := make([]entry, 0, len(m.ByBrokerAccount))
	for _, e := range m.ByBrokerAccount {
		g := entry{Account: e.Account}
		for i := range e.Transactions {
			g.Keys = append(g.Keys, e.Transactions[i].Key)
			g.TxAccounts = append(g.TxAccounts, e.Transactions[i].Account)
		}
		for i := range e.Assertions {
			g.AssertionAccount = append(g.AssertionAccount, e.Assertions[i].Account)
		}
		got = append(got, g)
	}
	want := []entry{{
		Account: "U0000002", Keys: []ledger.TransactionKey{{Source: "ibkr-flex", ExternalID: "T1"}},
		TxAccounts: []ledger.AccountID{""}, AssertionAccount: []ledger.AccountID{""},
	}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ByBrokerAccount (-want +got):\n%s", diff)
	}
}
