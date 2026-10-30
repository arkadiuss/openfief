package plugins_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// upgraded is samplePlugin at a new image that also requests a new egress host.
func upgraded() plugins.Plugin {
	p := samplePlugin()
	p.Image = imageB
	p.Manifest.Egress = append(p.Manifest.Egress, host("new.example.net", 443))
	return p
}

// noEgressGrant is fullGrant without any egress host.
func noEgressGrant() plugins.Grant {
	g := fullGrant()
	g.Permissions.Egress = nil
	return g
}

func TestPlanRollout(t *testing.T) {
	t.Parallel()
	current, next, previous := samplePlugin(), upgraded(), samplePlugin()
	full, narrowed := fullGrant(), noEgressGrant()
	viewOnly := fullGrant()
	viewOnly.Permissions.Capabilities = []plugins.Capability{plugins.CapabilityView}
	tests := []struct {
		name        string
		in          plugins.RolloutInput
		wantOutcome plugins.RolloutOutcome
		runs        *plugins.Plugin // the version expected to run under in.Grant, or nil
		wantPending plugins.Permissions
	}{
		{
			name:        "no grant runs nothing and everything is pending",
			in:          plugins.RolloutInput{Current: &current, CurrentDeployable: true, Previous: &previous},
			wantOutcome: plugins.RolloutNoGrant,
			wantPending: current.Requested(),
		},
		{
			name:        "fully granted current version runs",
			in:          plugins.RolloutInput{Current: &current, CurrentDeployable: true, Grant: &full},
			wantOutcome: plugins.RolloutCurrent,
			runs:        &current,
		},
		{
			name:        "upgrade requesting a new host keeps the previous version",
			in:          plugins.RolloutInput{Current: &next, CurrentDeployable: true, Previous: &previous, Grant: &full},
			wantOutcome: plugins.RolloutPrevious,
			runs:        &previous,
			wantPending: plugins.Permissions{Egress: []plugins.EgressHost{host("new.example.net", 443)}},
		},
		{
			name:        "first install requesting more than the grant waits",
			in:          plugins.RolloutInput{Current: &next, CurrentDeployable: true, Grant: &full},
			wantOutcome: plugins.RolloutAwaitingApproval,
			wantPending: plugins.Permissions{Egress: []plugins.EgressHost{host("new.example.net", 443)}},
		},
		{
			name:        "undeployable upgrade keeps the previous version",
			in:          plugins.RolloutInput{Current: &next, Previous: &previous, Grant: &full},
			wantOutcome: plugins.RolloutPrevious,
			runs:        &previous,
			wantPending: plugins.Permissions{Egress: []plugins.EgressHost{host("new.example.net", 443)}},
		},
		{
			name:        "undeployable first install runs nothing",
			in:          plugins.RolloutInput{Current: &current, Grant: &full},
			wantOutcome: plugins.RolloutNotDeployable,
		},
		{
			name:        "invalid current version keeps the previous one",
			in:          plugins.RolloutInput{Previous: &previous, Grant: &full},
			wantOutcome: plugins.RolloutPrevious,
			runs:        &previous,
		},
		{
			name:        "invalid current version and nothing running",
			in:          plugins.RolloutInput{Grant: &full},
			wantOutcome: plugins.RolloutNotDeployable,
		},
		{
			name:        "narrowed grant applies to the running version at once",
			in:          plugins.RolloutInput{Current: &current, CurrentDeployable: true, Previous: &previous, Grant: &narrowed},
			wantOutcome: plugins.RolloutPrevious,
			runs:        &previous,
			wantPending: plugins.Permissions{Egress: []plugins.EgressHost{host("api.example.com", 443), host("data.example.org", 8443)}},
		},
		{
			name:        "grant without a requested capability runs nothing",
			in:          plugins.RolloutInput{Current: &current, CurrentDeployable: true, Previous: &previous, Grant: &viewOnly},
			wantOutcome: plugins.RolloutNothingGranted,
			wantPending: plugins.Permissions{Capabilities: []plugins.Capability{plugins.CapabilityConnector}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := plugins.PlanRollout(tt.in)
			if err != nil {
				t.Fatalf("PlanRollout: %v", err)
			}
			var wantRun *plugins.EffectivePermissions
			if tt.runs != nil {
				if wantRun, err = plugins.Effective(tt.runs, tt.in.Grant); err != nil {
					t.Fatalf("Effective: %v", err)
				}
			}
			want := plugins.Rollout{Outcome: tt.wantOutcome, Run: wantRun, Pending: tt.wantPending}
			if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("PlanRollout mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlanRolloutRefuses(t *testing.T) {
	t.Parallel()
	current, previous := samplePlugin(), samplePlugin()
	invalid := fullGrant()
	invalid.Permissions.Egress = []plugins.EgressHost{host("*.example.com", 443)}
	other := fullGrant()
	other.Plugin = "other"
	badPrevious := samplePlugin()
	badPrevious.Image = "registry.example.com/openfief/sample-connector:1.0.0"
	full := fullGrant()
	tests := []struct {
		name string
		in   plugins.RolloutInput
		want error
	}{
		{
			name: "invalid grant",
			in:   plugins.RolloutInput{Current: &current, CurrentDeployable: true, Previous: &previous, Grant: &invalid},
			want: plugins.ErrWildcardHost,
		},
		{
			name: "grant for another plugin",
			in:   plugins.RolloutInput{Current: &current, CurrentDeployable: true, Grant: &other},
			want: plugins.ErrGrantMismatch,
		},
		{
			name: "invalid previous version",
			in:   plugins.RolloutInput{Previous: &badPrevious, Grant: &full},
			want: plugins.ErrImageNotPinned,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := plugins.PlanRollout(tt.in)
			if diff := cmp.Diff(tt.want, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("PlanRollout error mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(plugins.Rollout{}, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("PlanRollout returned a rollout with its error:\n%s", diff)
			}
		})
	}
}
