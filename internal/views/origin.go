package views

import (
	"strings"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// connectorSourcePrefix starts the key source of every transaction a
// connector instance delivered: "plugin:<instance>:<source>", as
// ingest.ConnectorKeySource writes it.
const connectorSourcePrefix = "plugin:"

// ConnectorInstance returns the connector instance a transaction key source
// names, such as "ibkr-main" from "plugin:ibkr-main:ibkr", and false for a
// source no connector instance wrote. Instance names are DNS labels, so the
// first colon after the prefix ends the name.
func ConnectorInstance(source string) (string, bool) {
	rest, ok := strings.CutPrefix(source, connectorSourcePrefix)
	if !ok {
		return "", false
	}
	instance, _, ok := strings.Cut(rest, ":")
	if !ok || instance == "" {
		return "", false
	}
	return instance, true
}

// Installed is the set of installed plugin instances, or nothing known
// about them: its zero value knows nothing.
type Installed struct {
	known bool
	names map[string]struct{}
}

// InstalledOf returns the instances of plugins as known to be installed.
// A Plugin being deleted still counts until it is gone.
func InstalledOf(plugins []Plugin) Installed {
	names := make(map[string]struct{}, len(plugins))
	for i := range plugins {
		names[plugins[i].Instance] = struct{}{}
	}
	return Installed{known: true, names: names}
}

// Origin is where a transaction came from: the connector instance whose
// delivery put it in the ledger, empty for any other source, and whether
// that instance is removed, so no Plugin of its name is installed any more
// (D-58). Reinstalling it under the same name makes it current again.
type Origin struct {
	Instance string
	Removed  bool
}

// OriginOf returns the origin of a transaction with key source source.
// Nothing is marked removed while the installed instances are unknown.
func (in Installed) OriginOf(source string) Origin {
	instance, ok := ConnectorInstance(source)
	if !ok {
		return Origin{}
	}
	_, installed := in.names[instance]
	return Origin{Instance: instance, Removed: in.known && !installed}
}

// Origins returns the origin of each of txs, in the same order.
func (in Installed) Origins(txs []ledger.Transaction) []Origin {
	out := make([]Origin, len(txs))
	for i := range txs {
		out[i] = in.OriginOf(txs[i].Key.Source)
	}
	return out
}
