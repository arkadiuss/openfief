package plugins

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// EgressAllowListConfigMap is the ConfigMap, in Namespace, through which the
// operator hands every plugin's egress allow-list to the egress proxy (D-32).
// Its data is EncodeAllowLists' output.
const EgressAllowListConfigMap = "openfief-egress-allowlist"

// EncodeAllowLists renders egress allow-lists as ConfigMap data: one key per
// plugin, named after the plugin, whose value lists its egress hosts as
// "host:port", one per line, sorted, each line ending in a newline. A plugin
// with no egress hosts gets an empty value, so it is known to have none.
// Every plugin name and egress host is validated, and a plugin may appear
// only once.
func EncodeAllowLists(lists []EgressAllowList) (map[string]string, error) {
	data := make(map[string]string, len(lists))
	for i, l := range lists {
		field := indexed("allowLists", i)
		if !validPluginName(l.Plugin) {
			return nil, fieldErr(field+".plugin", ErrInvalidName, l.Plugin)
		}
		if _, dup := data[l.Plugin]; dup {
			return nil, fieldErr(field+".plugin", ErrDuplicate, l.Plugin)
		}
		hosts := sortedCopy(l.Hosts, compareEgressHosts)
		var b strings.Builder
		for j, h := range hosts {
			if err := h.validate(); err != nil {
				return nil, &FieldError{Field: field + ".hosts", Err: err}
			}
			if j > 0 && hosts[j-1] == h {
				return nil, fieldErr(field+".hosts", ErrDuplicate, h.String())
			}
			b.WriteString(h.String())
			b.WriteByte('\n')
		}
		data[l.Plugin] = b.String()
	}
	return data, nil
}

// DecodeAllowLists reads ConfigMap data written by EncodeAllowLists, sorted by
// plugin name. It is strict, because the proxy refuses every destination it
// cannot read: each key must be a plugin name, and each line a canonical
// "host:port" (lower case, port always written) that appears once.
func DecodeAllowLists(data map[string]string) ([]EgressAllowList, error) {
	lists := make([]EgressAllowList, 0, len(data))
	for _, plugin := range slices.Sorted(maps.Keys(data)) {
		field := fmt.Sprintf("data[%q]", plugin)
		if !validPluginName(plugin) {
			return nil, fieldErr(field, ErrInvalidName, plugin)
		}
		hosts, err := decodeHosts(field, data[plugin])
		if err != nil {
			return nil, err
		}
		lists = append(lists, EgressAllowList{Plugin: plugin, Hosts: hosts})
	}
	return lists, nil
}

// decodeHosts reads one plugin's value: "host:port\n" lines.
func decodeHosts(field, value string) ([]EgressHost, error) {
	if value == "" {
		return nil, nil
	}
	body, ok := strings.CutSuffix(value, "\n")
	if !ok {
		return nil, fieldErr(field, ErrInvalidHostname, value) // every line ends in "\n"
	}
	lines := strings.Split(body, "\n")
	hosts := make([]EgressHost, 0, len(lines))
	for i, line := range lines {
		h, err := ParseEgressHost(line)
		if err != nil {
			return nil, &FieldError{Field: indexed(field, i), Err: err}
		}
		if h.String() != line {
			return nil, fieldErr(indexed(field, i), ErrInvalidHostname, line)
		}
		hosts = append(hosts, h)
	}
	hosts = sortedCopy(hosts, compareEgressHosts)
	for i := 1; i < len(hosts); i++ {
		if hosts[i] == hosts[i-1] {
			return nil, fieldErr(field, ErrDuplicate, hosts[i].String())
		}
	}
	return hosts, nil
}
