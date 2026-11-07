package plugins_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// caConnector is a connector whose manifest asks for the optional CA bundle
// Secret beside a required token, its base URL from config field baseUrl
// (D-75), with config.
func caConnector(config string) plugins.Plugin {
	return plugins.Plugin{
		Name:  "rates",
		Image: imageA,
		Manifest: plugins.Manifest{
			Capabilities:     []plugins.Capability{plugins.CapabilityConnector},
			Egress:           []plugins.EgressHost{host("api.example.com", 443)},
			EgressFromConfig: []string{"baseUrl"},
			Secrets: []plugins.SecretMount{
				{Role: "token", MountPath: "/etc/openfief/token"},
				{Role: plugins.CABundleRole, MountPath: "/etc/openfief/ca-bundle", Optional: true},
			},
		},
		Config: []byte(config),
	}
}

// caGrant approves everything caConnector requests with egress hosts.
func caGrant(egress ...plugins.EgressHost) plugins.Grant {
	return plugins.Grant{Plugin: "rates", Permissions: plugins.Permissions{
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Egress:       egress,
		Secrets:      []string{"rates--ca-bundle", "rates--token"},
	}}
}

func TestCABundleStateOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		expected, readable, exists bool
		want                       plugins.CABundleState
	}{
		{name: "not expected, not read", want: plugins.CABundleNotExpected},
		{name: "not expected, Secret exists anyway", readable: true, exists: true, want: plugins.CABundleNotExpected},
		{name: "expected and present", expected: true, readable: true, exists: true, want: plugins.CABundlePresent},
		{name: "expected and missing", expected: true, readable: true, want: plugins.CABundleMissing},
		{name: "expected, the operator may not read it", expected: true, want: plugins.CABundleUnreadable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := plugins.CABundleStateOf(tt.expected, tt.readable, tt.exists); got != tt.want {
				t.Errorf("CABundleStateOf(%v, %v, %v) = %s, want %s", tt.expected, tt.readable, tt.exists, got, tt.want)
			}
		})
	}
}

func TestSecretDataChecksum(t *testing.T) {
	t.Parallel()
	// The SHA-256 of no input: no keys hash nothing.
	const empty = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	ca := map[string][]byte{"ca.crt": []byte("-----BEGIN CERTIFICATE-----\nAA==\n-----END CERTIFICATE-----\n")}
	tests := []struct {
		name string
		a, b map[string][]byte
		same bool
	}{
		{name: "nil and empty", a: nil, b: map[string][]byte{}, same: true},
		{
			name: "same data built in another order",
			a:    map[string][]byte{"ca.crt": []byte("x"), "extra.crt": []byte("y")},
			b:    map[string][]byte{"extra.crt": []byte("y"), "ca.crt": []byte("x")}, same: true,
		},
		{name: "a changed value", a: ca, b: map[string][]byte{"ca.crt": []byte("other")}},
		{name: "a key moved into the value", a: map[string][]byte{"a": []byte("bc")}, b: map[string][]byte{"ab": []byte("c")}},
		{
			name: "a value moved into the next key",
			a:    map[string][]byte{"a": []byte("b"), "c": nil}, b: map[string][]byte{"a": nil, "bc": nil},
		},
		{name: "an empty Secret and a Secret with an empty key", a: nil, b: map[string][]byte{"": nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := plugins.SecretDataChecksum(tt.a), plugins.SecretDataChecksum(tt.b)
			if (a == b) != tt.same {
				t.Errorf("checksums %s and %s: equal %v, want %v", a, b, a == b, tt.same)
			}
			if a != plugins.SecretDataChecksum(tt.a) {
				t.Errorf("checksum of the same data changed")
			}
		})
	}
	if got := plugins.SecretDataChecksum(nil); got != empty {
		t.Errorf("SecretDataChecksum(nil) = %s, want %s", got, empty)
	}
}

func TestCABundleAccessName(t *testing.T) {
	t.Parallel()
	if got, want := plugins.CABundleAccessName("rates"), "openfief-plugin-rates-ca-bundle"; got != want {
		t.Errorf("CABundleAccessName = %q, want %q", got, want)
	}
}

// caMounts is what an instance of caConnector mounts: the roles of its
// Secrets, its CA bundle Secret, "" for none, and how many Secrets it
// requests.
type caMounts struct {
	Roles     []string
	Bundle    string
	Requested int
}

// mountsOf is what caConnector, expecting a CA bundle when extraCA, mounts
// under a grant that approves the Secrets approved.
func mountsOf(t *testing.T, extraCA bool, approved []string) caMounts {
	t.Helper()
	p := caConnector("")
	if err := p.SetExtraCA(extraCA); err != nil {
		t.Fatalf("SetExtraCA: %v", err)
	}
	g := caGrant(host("api.example.com", 443))
	g.Permissions.Secrets = approved
	e, err := plugins.Effective(&p, &g)
	if err != nil {
		t.Fatalf("Effective: %v", err)
	}
	m := caMounts{Requested: len(p.Requested().Secrets)}
	for _, sm := range e.Secrets() {
		m.Roles = append(m.Roles, sm.Role)
	}
	m.Bundle, _ = e.CABundleSecret()
	return m
}

func TestEffectiveCABundle(t *testing.T) {
	t.Parallel()
	both := []string{"rates--ca-bundle", "rates--token"}
	tests := []struct {
		name     string
		extraCA  bool
		approved []string
		want     caMounts
	}{
		{name: "no bundle expected: not mounted though approved", approved: both, want: caMounts{Roles: []string{"token"}, Requested: 2}},
		{
			name: "bundle expected and approved: mounted", extraCA: true, approved: both,
			want: caMounts{Roles: []string{plugins.CABundleRole, "token"}, Bundle: "rates--ca-bundle", Requested: 2},
		},
		{
			name: "bundle expected, not approved: not mounted", extraCA: true, approved: []string{"rates--token"},
			want: caMounts{Roles: []string{"token"}, Requested: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, mountsOf(t, tt.extraCA, tt.approved)); diff != "" {
				t.Errorf("mounts (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSetExtraCA(t *testing.T) {
	t.Parallel()
	asks := caConnector("")
	if err := asks.SetExtraCA(true); err != nil || !asks.ExtraCA {
		t.Errorf("SetExtraCA(true) on a manifest asking for the role: %v, ExtraCA %v; want it set", err, asks.ExtraCA)
	}
	if err := asks.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}
	other := samplePlugin()
	err := other.SetExtraCA(true)
	if !errors.Is(err, plugins.ErrExtraCANotRequested) || other.ExtraCA {
		t.Errorf("SetExtraCA(true) without the role: %v, ExtraCA %v; want ErrExtraCANotRequested and unchanged", err, other.ExtraCA)
	}
	other.ExtraCA = true
	if diff := cmp.Diff([]fieldError{{"extraCa", plugins.ErrExtraCANotRequested}}, fieldErrors(other.Validate()), cmpopts.EquateErrors()); diff != "" {
		t.Errorf("Validate() (-want +got):\n%s", diff)
	}
}

func TestPlanUpgradeKeepsExtraCA(t *testing.T) {
	t.Parallel()
	current := caConnector("")
	current.ExtraCA = true
	img, err := plugins.ParseImageRef(imageB)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	kept, _, err := plugins.PlanUpgrade(&current, &plugins.SignedImage{Image: img, Manifest: current.Manifest}, nil)
	if err != nil || !kept.ExtraCA {
		t.Errorf("upgrade still asking for the role: ExtraCA %v, %v; want it kept", kept.ExtraCA, err)
	}
	m := current.Manifest
	m.Secrets = m.Secrets[:1]
	dropped, _, err := plugins.PlanUpgrade(&current, &plugins.SignedImage{Image: img, Manifest: m}, nil)
	if err != nil || dropped.ExtraCA {
		t.Errorf("upgrade no longer asking for the role: ExtraCA %v, %v; want it dropped", dropped.ExtraCA, err)
	}
}

// TestConfigHostChangeKeepsApprovedAllowList checks that pointing a
// connector's base URL at another host, while the grant still approves only
// the old one, keeps the running version on the old approved allow-list
// with the new host pending (D-16, D-75), and that the upgrade diff shows
// the host change (D-90).
func TestConfigHostChangeKeepsApprovedAllowList(t *testing.T) {
	t.Parallel()
	running := caConnector(`{"baseUrl":"https://mirror.example.org/rates"}`)
	moved := caConnector(`{"baseUrl":"https://other.example.net/rates"}`)
	grant := caGrant(host("api.example.com", 443), host("mirror.example.org", 443))
	r, err := plugins.PlanRollout(plugins.RolloutInput{Current: &moved, CurrentDeployable: true, Previous: &running, Grant: &grant})
	if err != nil {
		t.Fatalf("PlanRollout: %v", err)
	}
	if r.Outcome != plugins.RolloutPrevious {
		t.Errorf("outcome %s, want %s", r.Outcome, plugins.RolloutPrevious)
	}
	want := []plugins.EgressHost{host("api.example.com", 443), host("mirror.example.org", 443)}
	if diff := cmp.Diff(want, r.Run.Egress()); diff != "" {
		t.Errorf("enforced egress (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(plugins.Permissions{Egress: []plugins.EgressHost{host("other.example.net", 443)}}, r.Pending); diff != "" {
		t.Errorf("pending (-want +got):\n%s", diff)
	}
	from, to := running.Requested(), moved.Requested()
	diff := plugins.PermissionsDiff{
		Added:   plugins.Permissions{Egress: []plugins.EgressHost{host("other.example.net", 443)}},
		Removed: plugins.Permissions{Egress: []plugins.EgressHost{host("mirror.example.org", 443)}},
	}
	if d := cmp.Diff(diff, plugins.DiffPermissions(&from, &to)); d != "" {
		t.Errorf("upgrade diff (-want +got):\n%s", d)
	}
}
