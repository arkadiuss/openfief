package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/app"
)

// errBacklog stands for the job queue failing.
var errBacklog = errors.New("job queue unavailable")

// fakeBacklog answers how many audit entries are stuck, or err.
type fakeBacklog struct {
	n   int
	err error
}

func (f fakeBacklog) StuckEntries(context.Context) (int, error) { return f.n, f.err }

// TestAuditLogBehind checks the check of the audit outbox (D-211, D-196): no
// entry stuck, no condition, so the item clears; any stuck, one condition of
// the one audit log whose detail says how many changes wait, that OpenFief keeps
// retrying and where to look; and a queue that can't be read is an error, so no
// item is cleared on a guess.
func TestAuditLogBehind(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		backlog fakeBacklog
		want    []attention.Condition
		wantErr error
	}{
		{name: "the log keeps up"},
		{
			name: "one entry waits", backlog: fakeBacklog{n: 1},
			want: []attention.Condition{{
				Kind: attention.KindAuditLogBehind, Subject: "audit-log",
				Detail: "1 change is waiting to enter the audit log; OpenFief keeps retrying. Check the database's health.",
			}},
		},
		{
			name: "several wait", backlog: fakeBacklog{n: 12},
			want: []attention.Condition{{
				Kind: attention.KindAuditLogBehind, Subject: "audit-log",
				Detail: "12 changes are waiting to enter the audit log; OpenFief keeps retrying. Check the database's health.",
			}},
		},
		{name: "the queue can't be read", backlog: fakeBacklog{err: errBacklog}, wantErr: errBacklog},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			check := app.AuditLogBehind(tt.backlog)
			if check.Kind != attention.KindAuditLogBehind {
				t.Errorf("Kind = %q, want %q", check.Kind, attention.KindAuditLogBehind)
			}
			got, err := check.Find(t.Context())
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Find error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("conditions (-want +got):\n%s", diff)
			}
			checkValid(t, got)
		})
	}
}

// checkValid fails t for each of conds that is not a valid condition.
func checkValid(t *testing.T, conds []attention.Condition) {
	t.Helper()
	for i := range conds {
		if err := conds[i].Validate(); err != nil {
			t.Errorf("the condition is not valid: %v", err)
		}
	}
}
