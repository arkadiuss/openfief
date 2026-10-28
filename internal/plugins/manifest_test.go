package plugins_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

func TestManifestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		change func(m *plugins.Manifest)
		want   []fieldError
	}{
		{name: "valid", change: func(*plugins.Manifest) {}},
		{
			name:   "no egress, no secrets, no schema",
			change: func(m *plugins.Manifest) { m.Egress, m.Secrets, m.ConfigSchema = nil, nil, nil },
		},
		{
			name:   "no capability",
			change: func(m *plugins.Manifest) { m.Capabilities = nil },
			want:   []fieldError{{"capabilities", plugins.ErrNoCapability}},
		},
		{
			name:   "unknown capability and duplicate",
			change: func(m *plugins.Manifest) { m.Capabilities = []plugins.Capability{"view", "trader", "view"} },
			want: []fieldError{
				{"capabilities[1]", plugins.ErrUnknownCapability},
				{"capabilities[2]", plugins.ErrDuplicate},
			},
		},
		{
			name: "bad egress hosts",
			change: func(m *plugins.Manifest) {
				m.Egress = []plugins.EgressHost{
					host("api.example.com", 443), host("*.example.com", 443), host("192.0.2.1", 443),
					host("api.example.com", 0), host("api.example.com", 443),
				}
			},
			want: []fieldError{
				{"egress[1]", plugins.ErrWildcardHost},
				{"egress[2]", plugins.ErrIPLiteral},
				{"egress[3]", plugins.ErrInvalidPort},
				{"egress[4]", plugins.ErrDuplicate},
			},
		},
		{
			name: "bad secrets",
			change: func(m *plugins.Manifest) {
				m.Secrets = []plugins.SecretMount{
					{Name: "Bad_Name", MountPath: "/etc/a"},
					{Name: "b", MountPath: "etc/b"},
					{Name: "c", MountPath: "/etc/c/../d"},
					{Name: "d", MountPath: "/var/run/secrets/kubernetes.io"},
					{Name: "e", MountPath: "/var"},
					{Name: "b", MountPath: "/etc/a/nested"},
					{Name: "g", MountPath: "/"},
				}
			},
			want: []fieldError{
				{"secrets[0]", plugins.ErrInvalidName},
				{"secrets[5]", plugins.ErrDuplicate},
				{"secrets[1].mountPath", plugins.ErrInvalidMountPath},
				{"secrets[2].mountPath", plugins.ErrInvalidMountPath},
				{"secrets[3].mountPath", plugins.ErrInvalidMountPath},
				{"secrets[4].mountPath", plugins.ErrInvalidMountPath},
				{"secrets[6].mountPath", plugins.ErrInvalidMountPath},
				{"secrets[4].mountPath", plugins.ErrOverlappingMountPaths},
				{"secrets[5].mountPath", plugins.ErrOverlappingMountPaths},
			},
		},
		{
			name:   "bad data tiers",
			change: func(m *plugins.Manifest) { m.DataTiers = []plugins.DataTier{"T1", "", "T 2", "T1"} },
			want: []fieldError{
				{"dataTiers[1]", plugins.ErrInvalidDataTier},
				{"dataTiers[2]", plugins.ErrInvalidDataTier},
				{"dataTiers[3]", plugins.ErrDuplicate},
			},
		},
		{
			name:   "config schema not JSON",
			change: func(m *plugins.Manifest) { m.ConfigSchema = []byte("{type: object") },
			want:   []fieldError{{"configSchema", plugins.ErrInvalidJSON}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := samplePlugin().Manifest
			tt.change(&m)
			if diff := cmp.Diff(tt.want, fieldErrors(m.Validate()), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestManifestRequested(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		manifest plugins.Manifest
		want     plugins.Permissions
	}{
		{
			name:     "sorted, secrets by name",
			manifest: samplePlugin().Manifest,
			want: plugins.Permissions{
				Capabilities: []plugins.Capability{plugins.CapabilityConnector},
				Egress:       []plugins.EgressHost{host("api.example.com", 443), host("data.example.org", 8443)},
				Secrets:      []string{"sample-cert", "sample-token"},
				DataTiers:    []plugins.DataTier{"T1"},
			},
		},
		{
			name:     "view with nothing else",
			manifest: plugins.Manifest{Capabilities: []plugins.Capability{plugins.CapabilityView}},
			want:     plugins.Permissions{Capabilities: []plugins.Capability{plugins.CapabilityView}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.manifest.Requested()); diff != "" {
				t.Errorf("Requested() (-want +got):\n%s", diff)
			}
		})
	}
}
