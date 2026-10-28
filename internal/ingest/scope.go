package ingest

import "github.com/arkadiuss/openfief/internal/ledger"

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
