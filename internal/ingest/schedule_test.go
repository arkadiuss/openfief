package ingest_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
)

// policy is a sync policy of six-hourly syncs with up to 30 minutes of
// jitter, retried after 5 minutes, doubling up to an hour.
func policy() ingest.SyncPolicy {
	return ingest.SyncPolicy{Interval: 6 * time.Hour, Jitter: 30 * time.Minute, RetryFirst: 5 * time.Minute, RetryMax: time.Hour}
}

func TestSyncPolicyValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		edit    func(p *ingest.SyncPolicy)
		wantErr error
	}{
		{name: "valid", edit: func(*ingest.SyncPolicy) {}},
		{name: "no jitter", edit: func(p *ingest.SyncPolicy) { p.Jitter = 0 }},
		{name: "retry max equals first", edit: func(p *ingest.SyncPolicy) { p.RetryMax = p.RetryFirst }},
		{name: "zero interval", edit: func(p *ingest.SyncPolicy) { p.Interval = 0 }, wantErr: ingest.ErrInvalidSyncPolicy},
		{name: "negative jitter", edit: func(p *ingest.SyncPolicy) { p.Jitter = -time.Second }, wantErr: ingest.ErrInvalidSyncPolicy},
		{name: "zero first retry", edit: func(p *ingest.SyncPolicy) { p.RetryFirst = 0 }, wantErr: ingest.ErrInvalidSyncPolicy},
		{name: "retry max below first", edit: func(p *ingest.SyncPolicy) { p.RetryMax = time.Minute }, wantErr: ingest.ErrInvalidSyncPolicy},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := policy()
			tt.edit(&p)
			if diff := cmp.Diff(tt.wantErr, p.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSyncPolicyNextSync(t *testing.T) {
	t.Parallel()
	ended := time.Date(2026, time.March, 2, 6, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		failures int
		jitter   time.Duration
		want     time.Duration
	}{
		{name: "success waits the interval", want: 6 * time.Hour},
		{name: "success adds the jitter", jitter: 10 * time.Minute, want: 6*time.Hour + 10*time.Minute},
		{name: "jitter above the policy's is clamped", jitter: 2 * time.Hour, want: 6*time.Hour + 30*time.Minute},
		{name: "negative jitter is clamped to zero", jitter: -time.Hour, want: 6 * time.Hour},
		{name: "first failure retries after the first delay", failures: 1, jitter: 10 * time.Minute, want: 5 * time.Minute},
		{name: "second failure doubles it", failures: 2, want: 10 * time.Minute},
		{name: "fourth failure", failures: 4, want: 40 * time.Minute},
		{name: "fifth failure reaches the maximum", failures: 5, want: time.Hour},
		{name: "many failures stay at the maximum", failures: 1000, want: time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := policy()
			if got := p.NextSync(ended, tt.failures, tt.jitter); !got.Equal(ended.Add(tt.want)) {
				t.Errorf("NextSync = %s, want %s", got, ended.Add(tt.want))
			}
		})
	}
}
