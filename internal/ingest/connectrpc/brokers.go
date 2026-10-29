package connectrpc

import (
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// brokerAccountFrom translates a broker account ID; it must be valid.
func brokerAccountFrom(field, id string) (ingest.BrokerAccountID, error) {
	b := ingest.BrokerAccountID(id)
	if err := b.Validate(); err != nil {
		return "", invalid(field, err)
	}
	return b, nil
}

// brokerEntries collects the entries of a mapping a connector named by
// broker account (D-72), in the order the delivery first names each
// account.
type brokerEntries struct {
	order []ingest.BrokerAccountID
	by    map[ingest.BrokerAccountID]*ingest.BrokerEntries
}

// entries returns the entries of broker account id, adding them.
func (b *brokerEntries) entries(id ingest.BrokerAccountID) *ingest.BrokerEntries {
	if b.by == nil {
		b.by = map[ingest.BrokerAccountID]*ingest.BrokerEntries{}
	}
	e, ok := b.by[id]
	if !ok {
		e = &ingest.BrokerEntries{Account: id}
		b.by[id] = e
		b.order = append(b.order, id)
	}
	return e
}

// list returns the collected entries, nil when there are none.
func (b *brokerEntries) list() []ingest.BrokerEntries {
	if len(b.order) == 0 {
		return nil
	}
	out := make([]ingest.BrokerEntries, 0, len(b.order))
	for _, id := range b.order {
		out = append(out, *b.by[id])
	}
	return out
}

// owner says which account a transaction or balance assertion names: its
// OpenFief account, or the broker account it names instead, which it
// returns. Naming both is refused.
func owner(field, accountID, brokerAccount string) (ingest.BrokerAccountID, error) {
	if brokerAccount == "" {
		return "", nil
	}
	if accountID != "" {
		return "", invalid(field+" names both account_id and broker_account", nil)
	}
	return brokerAccountFrom(field+".broker_account", brokerAccount)
}

// splitByBrokerAccount moves the transactions and balance assertions of m
// that d names by broker account into m.ByBrokerAccount, and sets
// m.BrokerAccounts to the broker accounts d lists. m's transactions and
// assertions are d's, translated in the same order.
func splitByBrokerAccount(d *connectorv1.Delivery, m *ingest.Mapping) error {
	var b brokerEntries
	var txs []ledger.Transaction
	for i, t := range d.GetTransactions() {
		id, err := owner(fmt.Sprintf("transactions[%d]", i), t.GetAccountId(), t.GetBrokerAccount())
		if err != nil {
			return err
		}
		if id == "" {
			txs = append(txs, m.Transactions[i])
			continue
		}
		e := b.entries(id)
		e.Transactions = append(e.Transactions, m.Transactions[i])
	}
	var assertions []ledger.BalanceAssertion
	for i, a := range d.GetBalanceAssertions() {
		id, err := owner(fmt.Sprintf("balance_assertions[%d]", i), a.GetAccountId(), a.GetBrokerAccount())
		if err != nil {
			return err
		}
		if id == "" {
			assertions = append(assertions, m.Assertions[i])
			continue
		}
		e := b.entries(id)
		e.Assertions = append(e.Assertions, m.Assertions[i])
	}
	for i, a := range d.GetBrokerAccounts() {
		id, err := brokerAccountFrom(fmt.Sprintf("broker_accounts[%d].id", i), a.GetId())
		if err != nil {
			return err
		}
		m.BrokerAccounts = append(m.BrokerAccounts, id)
	}
	m.Transactions, m.Assertions, m.ByBrokerAccount = txs, assertions, b.list()
	return nil
}
