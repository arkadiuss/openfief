package ingest

import (
	"strings"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// connectorKeyPrefix starts the transaction key source of everything a
// connector plugin delivers; the instance's name follows it.
const connectorKeyPrefix = "plugin:"

// ConnectorKeySource returns the transaction key source a transaction from a
// connector instance (D-48) gets for the source the connector gave it:
// "plugin:<instance>:<source>". The prefix keeps each instance's keys apart
// from every other's, so no connector can write or supersede another's
// transactions (D-18), and an instance reinstalled under the same name
// continues from its own (D-58).
func ConnectorKeySource(instance, source string) string {
	return connectorKeyPrefix + instance + ":" + source
}

// ScopeToConnector returns copies of txs whose key sources are scoped to the
// connector instance with ConnectorKeySource. A transaction whose source is
// empty stays empty, so the ledger still rejects it. txs is not changed.
func ScopeToConnector(instance string, txs []ledger.Transaction) []ledger.Transaction {
	out := make([]ledger.Transaction, len(txs))
	for i := range txs {
		out[i] = txs[i]
		if out[i].Key.Source != "" {
			out[i].Key.Source = ConnectorKeySource(instance, out[i].Key.Source)
		}
	}
	return out
}

// BookedBy returns the headers of the transactions of current that
// connector instance delivered, in order, each keyed as the connector keyed
// it: the source without the prefix ConnectorKeySource added. A connector
// that stores nothing learns from them what it booked, such as the coupons
// of a bond lot the owner deletes (D-154). Transactions of any other source
// are left out.
func BookedBy(instance string, current []ledger.Transaction) []ledger.TransactionHeader {
	prefix := ConnectorKeySource(instance, "")
	var out []ledger.TransactionHeader
	for i := range current {
		source, ok := strings.CutPrefix(current[i].Key.Source, prefix)
		if !ok || source == "" {
			continue
		}
		out = append(out, ledger.TransactionHeader{
			Key:     ledger.TransactionKey{Source: source, ExternalID: current[i].Key.ExternalID},
			Account: current[i].Account, Date: current[i].Date,
		})
	}
	return out
}
