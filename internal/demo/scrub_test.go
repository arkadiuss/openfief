package demo_test

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.yaml.in/yaml/v3"

	"github.com/arkadiuss/openfief/internal/identity"
)

// collectorConfig is the demo's OpenTelemetry Collector configuration
// (deploy/demo/telemetry), which ships the demo's logs, traces and metrics to
// Grafana Cloud.
const collectorConfig = "../../deploy/demo/telemetry/collector.yaml"

// redacted is what a rewritten value becomes.
const redacted = "[redacted]"

// The collector's processor transform/scrub is OTTL, which a unit test can't
// run without the collector's binary. Its statements are a small, regular
// subset though, and the collector's regular expressions are Go's, so these
// tests read the statements out of the committed configuration, refuse any
// statement of a form they don't understand, and apply the three forms
// (delete_matching_keys, replace_all_patterns, replace_pattern) to sample
// records with Go's regexp. What they can't check, that OTTL accepts the
// statements and runs them as written, scripts/demo/validate.sh's check of the
// build can't either; the real collector (contrib 0.164.0) was run on records
// like these when the layer was written (docs/dev/demo.md#telemetry).

// rule is one statement of transform/scrub, understood.
type rule struct {
	// fn is delete_matching_keys, replace_all_patterns or replace_pattern.
	fn string
	// path is what it works on: <context>.attributes, log.body, span.name or
	// span.status.message.
	path string
	// re is its regular expression.
	re *regexp.Regexp
}

// errUnknownForm is a statement the tests can't run.
var errUnknownForm = errors.New("a form of statement the test does not understand: teach it, so the rule is tested")

var (
	deleteStatement  = regexp.MustCompile(`^delete_matching_keys\(([a-z]+\.attributes), "(.*)"\)$`)
	replaceAllStmt   = regexp.MustCompile(`^replace_all_patterns\(([a-z]+\.attributes), "value", "(.*)", "\[redacted\]"\)$`)
	replaceStatement = regexp.MustCompile(`^replace_pattern\(([a-z.]+), "(.*)", "\[redacted\]"\)( where IsString\(log\.body\))?$`)
)

// loadConfig reads the collector's configuration as generic YAML.
func loadConfig(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(collectorConfig)
	if err != nil {
		t.Fatalf("reading the collector's configuration: %v", err)
	}
	var cfg map[string]any
	if unmarshalErr := yaml.Unmarshal(raw, &cfg); unmarshalErr != nil {
		t.Fatalf("parsing the collector's configuration: %v", unmarshalErr)
	}
	return cfg
}

// lookup returns the value at a path of map keys, or nil when there is none.
func lookup(v any, path ...string) any {
	for _, key := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[key]
	}
	return v
}

// list returns a YAML list, or nil when v is none.
func list(v any) []any {
	if l, ok := v.([]any); ok {
		return l
	}
	return nil
}

// object returns a YAML mapping, or nil when v is none.
func object(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

// stringList returns the strings of a YAML list.
func stringList(v any) []string {
	var out []string
	for _, item := range list(v) {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// parseStatement understands one statement, or says why not.
func parseStatement(s string) (rule, error) {
	for _, form := range []struct {
		fn string
		re *regexp.Regexp
	}{{"delete_matching_keys", deleteStatement}, {"replace_all_patterns", replaceAllStmt}, {"replace_pattern", replaceStatement}} {
		m := form.re.FindStringSubmatch(s)
		if m == nil {
			continue
		}
		re, err := regexp.Compile(m[2])
		if err != nil {
			return rule{}, fmt.Errorf("compiling %q: %w", m[2], err)
		}
		return rule{fn: form.fn, path: m[1], re: re}, nil
	}
	return rule{}, errUnknownForm
}

// rulesOf returns the rules of one OTTL context in a processor's statement
// list (log_statements, trace_statements or metric_statements), failing the
// test on a statement it does not understand or that works outside its
// context, so no rule goes untested.
func rulesOf(t *testing.T, cfg map[string]any, kind, context string) []rule {
	t.Helper()
	var out []rule
	for _, block := range list(lookup(cfg, "processors", "transform/scrub", kind)) {
		if lookup(block, "context") != context {
			continue
		}
		for _, s := range stringList(lookup(block, "statements")) {
			r, err := parseStatement(s)
			if err != nil {
				t.Fatalf("statement %q in %s: %v", s, context, err)
			}
			if !strings.HasPrefix(r.path, context+".") {
				t.Fatalf("statement %q works on %s, outside its context %s", s, r.path, context)
			}
			out = append(out, r)
		}
	}
	return out
}

// scrubAttributes applies a context's rules to attributes the way the
// collector does: delete_matching_keys removes each key that matches, and
// replace_all_patterns rewrites each match in each string value; other values
// are left as they are.
func scrubAttributes(rules []rule, attrs map[string]any) map[string]any {
	out := maps.Clone(attrs)
	for _, r := range rules {
		for k, v := range out {
			switch s, isString := v.(string); {
			case r.fn == "delete_matching_keys" && strings.HasSuffix(r.path, ".attributes") && r.re.MatchString(k):
				delete(out, k)
			case r.fn == "replace_all_patterns" && isString:
				out[k] = r.re.ReplaceAllString(s, redacted)
			}
		}
	}
	return out
}

// scrubField applies the replace_pattern rules of one path, such as log.body.
func scrubField(rules []rule, path, value string) string {
	for _, r := range rules {
		if r.fn == "replace_pattern" && r.path == path {
			value = r.re.ReplaceAllString(value, redacted)
		}
	}
	return value
}

// scrubContext is an OTTL context whose attributes a signal carries, in the
// processor's statement list of that signal.
type scrubContext struct{ kind, context string }

// attributeContexts are every context a signal's attributes live in.
func attributeContexts() []scrubContext {
	var out []scrubContext
	for kind, contexts := range map[string][]string{
		"log_statements":    {"resource", "scope", "log"},
		"trace_statements":  {"resource", "scope", "span", "spanevent"},
		"metric_statements": {"resource", "scope", "datapoint"},
	} {
		for _, c := range contexts {
			out = append(out, scrubContext{kind: kind, context: c})
		}
	}
	return out
}

// TestScrubDropsPersonalKeys checks that, in every context a signal's
// attributes live in, the keys that carry a client address, a session or a
// token ID, a credential, a header or a question are dropped, and the keys the
// demo's telemetry needs are kept (TEST-demo-telemetry-scrub).
func TestScrubDropsPersonalKeys(t *testing.T) {
	t.Parallel()
	dropped := []string{
		"client.address", "client.ip", "client_ip", "clientIP", "client.socket.address", "net.peer.ip", "net.sock.peer.addr",
		"network.peer.address", "remote_addr", "remote.ip", "source.address", "http.client_ip", "x-real-ip", "real_ip",
		"cf-connecting-ip", "http.request.header.x-forwarded-for", "x-forwarded-for", "forwarded", "k8s.pod.ip", "host.ip", "ip",
		"session.id", "session_id", "sessionId", "session", "sid", "http.request.header.cookie", "http.response.header.set-cookie",
		"cookie", "token", "token.id", "token_id", "tokenId", "access_token", "api_key", "apiKey", "secret", "password",
		"http.request.header.authorization", "authorization", "Authorization", "bearer", "jwt", "csrf",
		"http.request.header.user-agent", "user_agent.original", "headers", "header",
		"question", "ui.question", "q", "app.q", "prompt", "model.prompt", "query", "answer", "completion",
	}
	kept := []string{
		"kind", "browser", "major", "mobile", "ms", "line", "status", "code", "path", "script", "message", "instance",
		"http.route", "http.request.method", "http.response.status_code", "service.name", "k8s.pod.name", "k8s.namespace.name",
		"k8s.container.name", "deployment.environment.name", "log.iostream", "to", "os.type",
		"skip", "skipped", "tip", "zip", "shipping", "request", "quantile", "sequence", "uniq", "recipient", "description",
	}
	in := map[string]any{}
	for _, k := range slices.Concat(dropped, kept) {
		in[k] = "x"
	}
	cfg := loadConfig(t)
	for _, sc := range attributeContexts() {
		t.Run(sc.kind+"/"+sc.context, func(t *testing.T) {
			t.Parallel()
			assertKeys(t, scrubAttributes(rulesOf(t, cfg, sc.kind, sc.context), in), dropped, kept)
		})
	}
}

// assertKeys fails the test for each of dropped that got still has, and each
// of kept that it lacks.
func assertKeys(t *testing.T, got map[string]any, dropped, kept []string) {
	t.Helper()
	for _, k := range dropped {
		if _, ok := got[k]; ok {
			t.Errorf("key %q was kept; want it dropped", k)
		}
	}
	for _, k := range kept {
		if _, ok := got[k]; !ok {
			t.Errorf("key %q was dropped; want it kept", k)
		}
	}
}

// valueCases are strings and what the scrub makes of them: addresses, tokens
// and credentials rewritten, ordinary text, times and versions untouched.
func valueCases() []struct{ name, in, want string } {
	return []struct{ name, in, want string }{
		{"an IPv4 address", "203.0.113.9", redacted},
		{"an IPv4 address in text", "client 203.0.113.9 refused", "client " + redacted + " refused"},
		{"an IPv4 address with a port", "192.0.2.55:4433", redacted + ":4433"},
		{"a pod address", "10.42.0.17", redacted},
		{"a compressed IPv6 address", "2001:db8::1", redacted},
		{"the IPv6 loopback", "::1", redacted},
		{"a link-local IPv6 address", "fe80::1ff:fe23:4567:890a", redacted},
		{"a full IPv6 address", "2001:0db8:0000:0000:0000:ff00:0042:8329", redacted},
		{"an IPv6 address ending in ::", "1:2:3:4:5:6:7::", redacted},
		{"an IPv6 address in brackets with a port", "[2001:db8::1]:8080", "[" + redacted + "]:8080"},
		{"an IPv4-mapped IPv6 address", "::ffff:192.0.2.1", redacted},
		{"an IPv6 address in text", "from 2001:db8:85a3::8a2e:370:7334 to", "from " + redacted + " to"},
		{"a script token", "ofs_deadbeef01234567_cafebabecafebabecafebabe", redacted},
		{"a session token", "ofsess_AbCdEf123456_zzzzzzzzzzzzzzzz", redacted},
		{"an enrollment code", "ofenroll_QWERTY123456", redacted},
		{"a token in a URL", "https://x.example/v1?token=ofs_aaaa_bbbb&a=1", "https://x.example/v1?token=" + redacted + "&a=1"},
		{"a bearer credential", "Bearer abc.def-ghi_123", redacted},
		{"a bearer credential in lower case", "authorization: bearer abc123xyz", "authorization: " + redacted},
		{"a basic credential", "Basic dXNlcjpwYXNzd29yZA==", redacted},
		{"a JSON web token", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.c2lnbmF0dXJl", redacted},
		{"a long random string", "Zm9vYmFyYmF6cXV4Zm9vYmFyYmF6cXV4Zm9vYmFyYmF6", redacted},
		{"a sha256 digest", "sha256:" + strings.Repeat("ab", 32), "sha256:" + redacted},
		{"a time of day", "12:30:45", "12:30:45"},
		{"a timestamp", "2026-11-19T10:00:00.123456789Z", "2026-11-19T10:00:00.123456789Z"},
		{"a version of three numbers", "1.25.0", "1.25.0"},
		{"a version with a prefix", "v0.164.0", "v0.164.0"},
		{"a route", "GET /v1/accounts", "GET /v1/accounts"},
		{"an event kind", "turnstile-success", "turnstile-success"},
		{"a status", "503", "503"},
		{"a field name that starts like a prefix", "offset_ms", "offset_ms"},
		{"a pod's UUID", "0a1b2c3d-0000-4000-8000-000000000001", "0a1b2c3d-0000-4000-8000-000000000001"},
		{"a pod's name", "openfief-core-5c6b7d9f8d-xk2lp", "openfief-core-5c6b7d9f8d-xk2lp"},
		{"a log message", "sync failed", "sync failed"},
	}
}

// TestScrubRewritesPersonalValues checks that every string value of every
// attribute, in every context, passes through the same rewriting of
// addresses, token prefixes, credentials and long random strings, and that a
// number is left alone (TEST-demo-telemetry-scrub).
func TestScrubRewritesPersonalValues(t *testing.T) {
	t.Parallel()
	cfg := loadConfig(t)
	for _, sc := range attributeContexts() {
		rules := rulesOf(t, cfg, sc.kind, sc.context)
		for _, tc := range valueCases() {
			t.Run(sc.kind+"/"+sc.context+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				got := scrubAttributes(rules, map[string]any{"detail": tc.in, "count": 7})
				if got["detail"] != tc.want {
					t.Errorf("value %q became %q; want %q", tc.in, got["detail"], tc.want)
				}
				if got["count"] != 7 {
					t.Errorf("a number became %v; want it left as it was", got["count"])
				}
			})
		}
	}
}

// TestScrubRewritesTheLogBodyAndTheSpanText checks that the log body, the
// span's name and its status message are rewritten by the same patterns as
// the attributes' values.
func TestScrubRewritesTheLogBodyAndTheSpanText(t *testing.T) {
	t.Parallel()
	cfg := loadConfig(t)
	fields := []struct{ kind, context, path string }{
		{"log_statements", "log", "log.body"},
		{"trace_statements", "span", "span.name"},
		{"trace_statements", "span", "span.status.message"},
	}
	for _, f := range fields {
		rules := rulesOf(t, cfg, f.kind, f.context)
		for _, tc := range valueCases() {
			t.Run(f.path+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				if got := scrubField(rules, f.path, tc.in); got != tc.want {
					t.Errorf("%s %q became %q; want %q", f.path, tc.in, got, tc.want)
				}
			})
		}
	}
}

// TestScrubKnowsOpenfiefTokenPrefixes checks that a token of each kind the
// identity context issues, whatever follows its prefix, is rewritten, in an
// attribute and in a log body.
func TestScrubKnowsOpenfiefTokenPrefixes(t *testing.T) {
	t.Parallel()
	cfg := loadConfig(t)
	rules := rulesOf(t, cfg, "log_statements", "log")
	for _, prefix := range []string{identity.SecretPrefix, identity.SessionPrefix, identity.EnrollmentPrefix} {
		got := scrubAttributes(rules, map[string]any{"v": "see " + prefix + "abcd1234_x"})
		if want := "see " + redacted; got["v"] != want {
			t.Errorf("a token starting %q became %q; want %q", prefix, got["v"], want)
		}
		if body := scrubField(rules, "log.body", prefix+"abcd1234"); body != redacted {
			t.Errorf("a log body of a token starting %q became %q; want %q", prefix, body, redacted)
		}
	}
}

// patterns returns the regular expressions of the rules that use fn.
func patterns(rules []rule, fn string) []string {
	var out []string
	for _, r := range rules {
		if r.fn == fn {
			out = append(out, r.re.String())
		}
	}
	return out
}

// TestScrubStatementsAgree checks that every context drops the same keys and
// rewrites the same values, and that the log body and the span's name and
// status message use the very same patterns, so the contexts cannot drift
// apart.
func TestScrubStatementsAgree(t *testing.T) {
	t.Parallel()
	cfg := loadConfig(t)
	base := rulesOf(t, cfg, "log_statements", "resource")
	for _, sc := range attributeContexts() {
		rules := rulesOf(t, cfg, sc.kind, sc.context)
		for _, fn := range []string{"delete_matching_keys", "replace_all_patterns"} {
			got := patterns(rules, fn)
			if diff := cmp.Diff(patterns(base, fn), got); diff != "" || len(got) == 0 {
				t.Errorf("%s in %s's %s differs from the logs' resource (-resource +%s):\n%s", fn, sc.context, sc.kind, sc.context, diff)
			}
		}
	}
	values := patterns(base, "replace_all_patterns")
	body := patterns(rulesOf(t, cfg, "log_statements", "log"), "replace_pattern")
	span := patterns(rulesOf(t, cfg, "trace_statements", "span"), "replace_pattern")
	if diff := cmp.Diff(values, body); diff != "" {
		t.Errorf("the log body doesn't use the values' patterns (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(slices.Concat(values, values), span); diff != "" {
		t.Errorf("the span's name and status message don't use the values' patterns (-want +got):\n%s", diff)
	}
}

// TestCollectorPipelinesScrubBeforeTheExporter checks the configuration's
// structure: there is a pipeline for logs, traces and metrics, each runs
// transform/scrub with nothing but the memory limiter before it, and each
// exports to the one exporter, Grafana's; the scrub fails closed
// (TEST-demo-telemetry-scrub).
func TestCollectorPipelinesScrubBeforeTheExporter(t *testing.T) {
	t.Parallel()
	cfg := loadConfig(t)
	if got := lookup(cfg, "processors", "transform/scrub", "error_mode"); got != "propagate" {
		t.Errorf("transform/scrub's error_mode is %v; want propagate, so a failing statement drops the batch", got)
	}
	pipelines := object(lookup(cfg, "service", "pipelines"))
	for _, signal := range []string{"logs", "traces", "metrics"} {
		if _, ok := pipelines[signal]; !ok {
			t.Errorf("no %s pipeline", signal)
		}
	}
	for name, p := range pipelines {
		processors := stringList(lookup(p, "processors"))
		i := slices.Index(processors, "transform/scrub")
		if i < 0 {
			t.Errorf("the pipeline %s doesn't run transform/scrub: %v", name, processors)
			continue
		}
		if before := slices.DeleteFunc(slices.Clone(processors[:i]), func(s string) bool { return s == "memory_limiter" }); len(before) > 0 {
			t.Errorf("the pipeline %s runs %v before transform/scrub", name, before)
		}
		if diff := cmp.Diff([]string{"otlphttp/grafana"}, stringList(lookup(p, "exporters"))); diff != "" {
			t.Errorf("the pipeline %s's exporters (-want +got):\n%s", name, diff)
		}
	}
	if exporters := object(lookup(cfg, "exporters")); len(exporters) != 1 {
		t.Errorf("the configuration has %d exporters; want only otlphttp/grafana", len(exporters))
	}
}
