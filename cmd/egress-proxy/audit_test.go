package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins/smokescreen"
)

func TestReport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		audit   smokescreen.Audit
		want    string
		wantErr error
	}{
		{
			name:  "a clean audit names each destination and passes",
			audit: smokescreen.Audit{Decisions: 5, Refused: 2, Allowed: []smokescreen.Destination{{Instance: "nbp", Host: "api.nbp.pl:443", Count: 3}}},
			want: "egress audit: 5 decision lines, 2 refused, 3 allowed\n" +
				"egress audit: allowed nbp to api.nbp.pl:443, 3 times\n" +
				"egress audit: every allowed request went to a host its instance's grant approves\n",
		},
		{
			name:  "a problem fails the audit after it is printed",
			audit: smokescreen.Audit{Decisions: 1, Problems: []string{"decision line a1: no decision (allow)"}},
			want: "egress audit: 1 decision lines, 0 refused, 1 allowed\n" +
				"egress audit: decision line a1: no decision (allow)\n",
			wantErr: errAuditFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			err := report(&out, tt.audit)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("report() error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, out.String()); diff != "" {
				t.Errorf("report() output mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRunAuditRefusesBadFlags(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	if err := runAudit(t.Context(), []string{"-no-such-flag"}, &out); err == nil {
		t.Error("runAudit accepted an unknown flag, want an error")
	}
}
