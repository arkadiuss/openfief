package smokescreen

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// DecisionMessage is the msg of the line Smokescreen logs for each request
// it decides, its canonical decision line.
const DecisionMessage = "CANONICAL-PROXY-DECISION"

// bareHost is what requested_host may hold: a host and an optional port,
// and nothing of a path, query or credentials.
var bareHost = regexp.MustCompile(`^[A-Za-z0-9.-]+(:\d+)?$`)

// DecisionLine is one decision line of the proxy's log, decoded from JSON.
type DecisionLine map[string]any

// Text returns field key of l if it is a string, and "" otherwise.
func (l DecisionLine) Text(key string) string {
	if s, ok := l[key].(string); ok {
		return s
	}
	return ""
}

// DecisionLines returns the decision lines of log, a stream of JSON lines
// from one or more of the proxy's pods; every other line is skipped.
func DecisionLines(log []byte) []DecisionLine {
	var lines []DecisionLine
	for raw := range bytes.SplitSeq(log, []byte("\n")) {
		var l DecisionLine
		if json.Unmarshal(raw, &l) != nil || l["msg"] != DecisionMessage {
			continue
		}
		lines = append(lines, l)
	}
	return lines
}

// decisionKeys are the fields a decision line may hold: who asked (role,
// the plugin instance, and its address), what (proxy type and host:port),
// the decision and its timing. Nothing else, so no path, query, header,
// body or token is logged (GR-12).
func decisionKeys() []string {
	return []string{
		"allow", "conn_establish_time_ms", "content_length", "decision_reason", "dns_lookup_time_ms",
		"enforce_would_deny", "error", "id", "inbound_remote_addr", "level", "msg", "outbound_local_addr",
		"outbound_remote_addr", "project", "proxy_type", "requested_host", "role", "start_time", "status_code",
		"time", "trace_id",
	}
}

// Destination is a host:port the proxy let an instance reach, and how many
// times it did.
type Destination struct {
	Instance string
	Host     string
	Count    int
}

// Audit is what AuditLog found in the proxy's log.
type Audit struct {
	// Decisions counts the decision lines, Refused those that refused.
	Decisions, Refused int
	// Allowed are the destinations the proxy allowed, by instance and host.
	Allowed []Destination
	// Problems describe each line that breaks a rule, by its id, instance
	// and field names; never by a value that could hold a request's
	// contents.
	Problems []string
}

// AuditLog checks every decision line of log against the rules of D-200's
// criterion 16 and GR-12: a line holds only decisionKeys, a bare host:port
// in requested_host, its instance in role and its decision in allow; and
// every request the proxy allowed went to a host the instance's grant
// approves. granted holds each instance's approved hosts.
func AuditLog(log []byte, granted map[string][]plugins.EgressHost) Audit {
	var a Audit
	allowed := map[[2]string]int{}
	for _, l := range DecisionLines(log) {
		a.Decisions++
		problems := lineProblems(l)
		a.Problems = append(a.Problems, problems...)
		if allow, ok := l["allow"].(bool); !ok || !allow {
			a.Refused++
			continue
		}
		role, host := l.Text("role"), l.Text("requested_host")
		if len(problems) > 0 {
			continue // its instance or host can't be trusted; already reported
		}
		if !grants(granted[role], host) {
			a.Problems = append(a.Problems, fmt.Sprintf("decision line %v: allowed %s to %s, which its grant doesn't approve", l["id"], role, host))
		}
		allowed[[2]string{role, host}]++
	}
	for k, n := range allowed {
		a.Allowed = append(a.Allowed, Destination{Instance: k[0], Host: k[1], Count: n})
	}
	slices.SortFunc(a.Allowed, func(x, y Destination) int {
		return cmp.Or(cmp.Compare(x.Instance, y.Instance), cmp.Compare(x.Host, y.Host))
	})
	return a
}

// lineProblems returns what is wrong with one decision line: a field
// beyond decisionKeys, a requested_host that isn't a bare host:port, no
// instance or no decision.
func lineProblems(l DecisionLine) []string {
	var problems []string
	keys := make([]string, 0, len(l))
	for key := range l {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if !slices.Contains(decisionKeys(), key) {
			problems = append(problems, fmt.Sprintf("decision line %v: unexpected field %q", l["id"], key))
		}
	}
	if !bareHost.MatchString(l.Text("requested_host")) {
		problems = append(problems, fmt.Sprintf("decision line %v: requested_host is not a bare host and port", l["id"]))
	}
	if _, ok := l["role"].(string); !ok {
		problems = append(problems, fmt.Sprintf("decision line %v: no instance (role)", l["id"]))
	}
	if _, ok := l["allow"].(bool); !ok {
		problems = append(problems, fmt.Sprintf("decision line %v: no decision (allow)", l["id"]))
	}
	return problems
}

// grants reports whether hosts holds host, a host:port.
func grants(hosts []plugins.EgressHost, host string) bool {
	return slices.ContainsFunc(hosts, func(h plugins.EgressHost) bool { return h.String() == host })
}
