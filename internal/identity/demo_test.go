package identity_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/identity"
)

func TestDemoDeployment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		sessions, marked bool
		wantSessions     bool
		wantLoading      bool
	}{
		{name: "demo sessions on, database marked: ready", sessions: true, marked: true, wantSessions: true},
		{name: "demo sessions on, database not marked yet: loading", sessions: true, wantLoading: true},
		{name: "demo sessions off, database marked: no demo", marked: true},
		{name: "demo sessions off, database not marked: no demo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := identity.DemoSessionsAvailable(tt.sessions, tt.marked); got != tt.wantSessions {
				t.Errorf("DemoSessionsAvailable(%v, %v) = %v, want %v", tt.sessions, tt.marked, got, tt.wantSessions)
			}
			if got := identity.DemoLoading(tt.sessions, tt.marked); got != tt.wantLoading {
				t.Errorf("DemoLoading(%v, %v) = %v, want %v", tt.sessions, tt.marked, got, tt.wantLoading)
			}
		})
	}
}
