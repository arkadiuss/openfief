package smokescreen_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/smokescreen"
)

// decision is a decision line as Smokescreen logs it; fields are the
// JSON members after msg, already encoded.
func decision(fields string) string {
	return `{"level":"info","msg":"CANONICAL-PROXY-DECISION",` + fields + `}`
}

// auditGrants are the synthetic instances' approved hosts: nbp may reach
// api.nbp.pl on 443, ui nothing.
func auditGrants() map[string][]plugins.EgressHost {
	return map[string][]plugins.EgressHost{
		"nbp": {{Host: "api.nbp.pl", Port: plugins.DefaultEgressPort}},
		"ui":  {},
	}
}

func TestAuditLog(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		log  []string
		want smokescreen.Audit
	}{
		{
			name: "an empty log has nothing to check",
			want: smokescreen.Audit{},
		},
		{
			name: "allowed requests to granted hosts and refusals pass, counted by instance and host",
			log: []string{
				`{"level":"info","msg":"egress proxy listening"}`,
				decision(`"id":"a1","role":"nbp","proxy_type":"connect","requested_host":"api.nbp.pl:443","allow":true,"decision_reason":"host matched allowed domain in rule"`),
				decision(`"id":"a2","role":"nbp","proxy_type":"connect","requested_host":"api.nbp.pl:443","allow":true`),
				decision(`"id":"r1","role":"ui","proxy_type":"connect","requested_host":"example.com:443","allow":false,"decision_reason":"host not in allow-list"`),
				decision(`"id":"r2","role":"","proxy_type":"connect","requested_host":"example.com:443","allow":false`),
				`not json at all`,
			},
			want: smokescreen.Audit{
				Decisions: 4, Refused: 2,
				Allowed: []smokescreen.Destination{{Instance: "nbp", Host: "api.nbp.pl:443", Count: 2}},
			},
		},
		{
			name: "an allowed request to a host outside the instance's grant is a problem",
			log: []string{
				decision(`"id":"a3","role":"nbp","proxy_type":"connect","requested_host":"example.org:443","allow":true`),
				decision(`"id":"a4","role":"ui","proxy_type":"connect","requested_host":"api.nbp.pl:443","allow":true`),
				decision(`"id":"a5","role":"nbp","proxy_type":"connect","requested_host":"api.nbp.pl:8443","allow":true`),
			},
			want: smokescreen.Audit{
				Decisions: 3,
				Allowed: []smokescreen.Destination{
					{Instance: "nbp", Host: "api.nbp.pl:8443", Count: 1},
					{Instance: "nbp", Host: "example.org:443", Count: 1},
					{Instance: "ui", Host: "api.nbp.pl:443", Count: 1},
				},
				Problems: []string{
					"decision line a3: allowed nbp to example.org:443, which its grant doesn't approve",
					"decision line a4: allowed ui to api.nbp.pl:443, which its grant doesn't approve",
					"decision line a5: allowed nbp to api.nbp.pl:8443, which its grant doesn't approve",
				},
			},
		},
		{
			name: "an allowed request by an instance without a grant is a problem",
			log: []string{
				decision(`"id":"a6","role":"ghost","proxy_type":"connect","requested_host":"api.nbp.pl:443","allow":true`),
			},
			want: smokescreen.Audit{
				Decisions: 1,
				Allowed:   []smokescreen.Destination{{Instance: "ghost", Host: "api.nbp.pl:443", Count: 1}},
				Problems:  []string{"decision line a6: allowed ghost to api.nbp.pl:443, which its grant doesn't approve"},
			},
		},
		{
			name: "a line holding more than who, where and the decision is a problem, never quoting the value",
			log: []string{
				decision(`"id":"p1","role":"nbp","proxy_type":"connect","requested_host":"api.nbp.pl:443","allow":true,"proxy_authorization":"Basic c3ludGhldGlj"`),
				decision(`"id":"p2","role":"nbp","proxy_type":"http","requested_host":"api.nbp.pl/api/exchangerates?format=json","allow":false`),
			},
			want: smokescreen.Audit{
				Decisions: 2, Refused: 1,
				Problems: []string{
					`decision line p1: unexpected field "proxy_authorization"`,
					"decision line p2: requested_host is not a bare host and port",
				},
			},
		},
		{
			name: "a line without its instance or decision is a problem",
			log: []string{
				decision(`"id":"m1","proxy_type":"connect","requested_host":"api.nbp.pl:443","allow":true`),
				decision(`"id":"m2","role":"nbp","proxy_type":"connect","requested_host":"api.nbp.pl:443"`),
			},
			want: smokescreen.Audit{
				Decisions: 2, Refused: 1,
				Problems: []string{
					"decision line m1: no instance (role)",
					"decision line m2: no decision (allow)",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := smokescreen.AuditLog([]byte(strings.Join(tt.log, "\n")), auditGrants())
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("AuditLog() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDecisionLines(t *testing.T) {
	t.Parallel()
	log := strings.Join([]string{
		`{"level":"info","msg":"egress proxy listening"}`,
		decision(`"id":"d1","role":"nbp","requested_host":"api.nbp.pl:443","allow":true`),
		`{"msg":`,
		decision(`"id":"d2","role":"","requested_host":"example.com:443","allow":false`),
	}, "\n")
	lines := smokescreen.DecisionLines([]byte(log))
	got := make([]string, 0, len(lines))
	for _, l := range lines {
		got = append(got, l.Text("id")+" "+l.Text("role")+" "+l.Text("requested_host")+" "+l.Text("allow"))
	}
	want := []string{"d1 nbp api.nbp.pl:443 ", "d2  example.com:443 "}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("DecisionLines() mismatch (-want +got):\n%s", diff)
	}
}
