package plugins_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"slices"
	"testing"
	"time"

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

// testCert is a self-signed certificate made for the test, as PEM: no
// key or certificate is checked in.
func testCert(t *testing.T, cn string) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: cn},
		NotBefore: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), NotAfter: time.Date(2036, 1, 1, 0, 0, 0, 0, time.UTC),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("certificate: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestParseCABundle(t *testing.T) {
	t.Parallel()
	one, two := testCert(t, "ca.example.com"), testCert(t, "other.example.org")
	notCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not DER")})
	key := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("synthetic")})
	tests := []struct {
		name    string
		data    []byte
		want    int
		wantErr error
	}{
		{name: "one certificate", data: one, want: 1},
		{name: "two certificates with blank lines around", data: slices.Concat([]byte("\n"), one, []byte("\n\n"), two, []byte("\n")), want: 2},
		{name: "empty", data: nil, wantErr: plugins.ErrCABundleEmpty},
		{name: "only whitespace", data: []byte(" \n\t\n"), wantErr: plugins.ErrCABundleEmpty},
		{name: "not PEM at all", data: []byte("hello"), wantErr: plugins.ErrCABundleNotPEM},
		{name: "text after a certificate", data: slices.Concat(one, []byte("trailing")), wantErr: plugins.ErrCABundleNotPEM},
		{name: "a private key", data: key, wantErr: plugins.ErrCABundleNotCertificate},
		{name: "a key after a certificate", data: slices.Concat(one, key), wantErr: plugins.ErrCABundleNotCertificate},
		{name: "a CERTIFICATE block that doesn't parse", data: notCert, wantErr: plugins.ErrCABundleBadCertificate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := plugins.ParseCABundle(tt.data)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseCABundle: %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			want := plugins.CABundle{PEM: tt.data, Certificates: tt.want, Checksum: plugins.CABundleChecksum(tt.data)}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("ParseCABundle (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCABundleChecksum(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		data string
		want string
	}{
		// What sha256sum prints for the same bytes.
		{name: "nothing", data: "", want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{name: "abc", data: "abc", want: "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := plugins.CABundleChecksum([]byte(tt.data)); got != tt.want {
				t.Errorf("CABundleChecksum(%q) = %s, want %s", tt.data, got, tt.want)
			}
		})
	}
}

func TestCABundleStateOf(t *testing.T) {
	t.Parallel()
	const sum = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	tests := []struct {
		name     string
		extraCA  bool
		checksum string
		want     plugins.CABundleState
	}{
		{name: "not expected", want: plugins.CABundleNotExpected},
		{name: "not expected, a stamp left over", checksum: sum, want: plugins.CABundleNotExpected},
		{name: "expected, nothing stamped", extraCA: true, want: plugins.CABundleNotStamped},
		{name: "expected and stamped", extraCA: true, checksum: sum, want: plugins.CABundleStamped},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := plugins.CABundleStateOf(tt.extraCA, tt.checksum); got != tt.want {
				t.Errorf("CABundleStateOf(%v, %q) = %s, want %s", tt.extraCA, tt.checksum, got, tt.want)
			}
		})
	}
}

func TestPluginStampCABundle(t *testing.T) {
	t.Parallel()
	const sum = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	tests := []struct {
		name     string
		extraCA  bool
		before   string
		checksum string
		want     string
		wantErr  error
	}{
		{name: "stamps an expected bundle", extraCA: true, checksum: sum, want: sum},
		{name: "an empty checksum clears it", extraCA: true, before: sum, want: ""},
		{name: "clears it even when not expected", before: sum, want: ""},
		{name: "refused when no bundle is expected", checksum: sum, wantErr: plugins.ErrExtraCANotExpected},
		{name: "refused when not hex SHA-256", extraCA: true, before: sum, checksum: "ABC", want: sum, wantErr: plugins.ErrCABundleChecksum},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := caConnector("")
			p.ExtraCA, p.CABundleSHA256 = tt.extraCA, tt.before
			if err := p.StampCABundle(tt.checksum); !errors.Is(err, tt.wantErr) {
				t.Errorf("StampCABundle: %v, want %v", err, tt.wantErr)
			}
			if p.CABundleSHA256 != tt.want {
				t.Errorf("CABundleSHA256 = %q, want %q", p.CABundleSHA256, tt.want)
			}
		})
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
	other.ExtraCA, other.CABundleSHA256 = true, "not-a-checksum"
	want := []fieldError{{"extraCa", plugins.ErrExtraCANotRequested}, {"caBundleSha256", plugins.ErrCABundleChecksum}}
	if diff := cmp.Diff(want, fieldErrors(other.Validate()), cmpopts.EquateErrors()); diff != "" {
		t.Errorf("Validate() (-want +got):\n%s", diff)
	}
	asks.CABundleSHA256 = plugins.CABundleChecksum([]byte("bundle"))
	if err := asks.SetExtraCA(false); err != nil || asks.ExtraCA || asks.CABundleSHA256 != "" {
		t.Errorf("SetExtraCA(false): %v, ExtraCA %v, stamp %q; want both cleared", err, asks.ExtraCA, asks.CABundleSHA256)
	}
}

func TestPlanUpgradeKeepsExtraCA(t *testing.T) {
	t.Parallel()
	current := caConnector("")
	current.ExtraCA, current.CABundleSHA256 = true, plugins.CABundleChecksum([]byte("bundle"))
	img, err := plugins.ParseImageRef(imageB)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	kept, _, err := plugins.PlanUpgrade(&current, &plugins.SignedImage{Image: img, Manifest: current.Manifest}, nil)
	if err != nil || !kept.ExtraCA || kept.CABundleSHA256 != current.CABundleSHA256 {
		t.Errorf("upgrade still asking for the role: ExtraCA %v, stamp %q, %v; want both kept", kept.ExtraCA, kept.CABundleSHA256, err)
	}
	m := current.Manifest
	m.Secrets = m.Secrets[:1]
	dropped, _, err := plugins.PlanUpgrade(&current, &plugins.SignedImage{Image: img, Manifest: m}, nil)
	if err != nil || dropped.ExtraCA || dropped.CABundleSHA256 != "" {
		t.Errorf("upgrade no longer asking for the role: ExtraCA %v, stamp %q, %v; want both dropped", dropped.ExtraCA, dropped.CABundleSHA256, err)
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
