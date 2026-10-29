package views_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// testImage is a synthetic image reference; not of any real image.
const testImage = "registry.example.com/openfief/sample@sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"

// signed is a running plugin record whose image a trusted signer of kind
// signed, with the operator's status values.
func signed(trust, kind, signer string) views.PluginRecord {
	return views.PluginRecord{
		Instance: "sample", Image: testImage,
		Requested: []string{"connector", "view"}, Effective: []string{"connector"},
		Ready:     views.Condition{Status: "True", Reason: "Running"},
		Signature: views.Condition{Status: "True", Reason: "SignedByTrustedSigner"},
		Trust:     trust, SignerKind: kind, Signer: signer,
	}
}

func TestPluginOf(t *testing.T) {
	t.Parallel()
	running := func(trust views.SignatureTrust, signer string) views.Plugin {
		return views.Plugin{
			Instance: "sample", Image: testImage, Digest: "sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1",
			Capabilities: []string{"connector"}, Requested: []string{"connector", "view"},
			Trust: trust, Signer: signer, TrustReason: "SignedByTrustedSigner", Phase: "Running", Ready: true,
		}
	}
	tests := []struct {
		name   string
		record func() views.PluginRecord
		want   func() views.Plugin
	}{
		{
			name: "verified plugin",
			record: func() views.PluginRecord {
				r := signed("Verified", "Publisher", "example-publisher")
				r.Signature.Reason = "Verified"
				return r
			},
			want: func() views.Plugin {
				p := running(views.TrustVerified, "example-publisher")
				p.TrustReason = "Verified"
				return p
			},
		},
		{
			name:   "signed by OpenFief's CI",
			record: func() views.PluginRecord { return signed("Unverified", "OpenFiefCI", "openfief-ci") },
			want:   func() views.Plugin { return running(views.TrustOpenFiefCI, "openfief-ci") },
		},
		{
			name:   "signed by a catalog publisher, not pinned",
			record: func() views.PluginRecord { return signed("Unverified", "Publisher", "example-publisher") },
			want:   func() views.Plugin { return running(views.TrustPublisher, "example-publisher") },
		},
		{
			name:   "signed by the owner's signer",
			record: func() views.PluginRecord { return signed("Unverified", "Owner", "owner-laptop") },
			want:   func() views.Plugin { return running(views.TrustOwner, "owner-laptop") },
		},
		{
			name:   "signed, from an operator that names no signer",
			record: func() views.PluginRecord { return signed("", "", "") },
			want:   func() views.Plugin { return running(views.TrustSigned, "") },
		},
		{
			name: "verified, from an operator that names no signer",
			record: func() views.PluginRecord {
				r := signed("", "", "")
				r.Signature.Reason = "Verified"
				return r
			},
			want: func() views.Plugin {
				p := running(views.TrustVerified, "")
				p.TrustReason = "Verified"
				return p
			},
		},
		{
			name: "refused, with a stale signer that is not shown",
			record: func() views.PluginRecord {
				r := signed("Unverified", "Owner", "owner-laptop")
				r.Signature = views.Condition{Status: "False", Reason: "UntrustedSigner"}
				r.Ready = views.Condition{Status: "False", Reason: "SignatureNotVerified"}
				r.Effective = nil
				return r
			},
			want: func() views.Plugin {
				p := running(views.TrustRefused, "")
				p.TrustReason, p.Phase, p.Ready, p.Capabilities = "UntrustedSigner", "SignatureNotVerified", false, nil
				return p
			},
		},
		{
			name: "registry unreadable",
			record: func() views.PluginRecord {
				r := signed("", "", "")
				r.Signature = views.Condition{Status: "Unknown", Reason: "RegistryUnavailable"}
				return r
			},
			want: func() views.Plugin {
				p := running(views.TrustNotChecked, "")
				p.TrustReason = "RegistryUnavailable"
				return p
			},
		},
		{
			name: "not reconciled yet",
			record: func() views.PluginRecord {
				return views.PluginRecord{Instance: "sample", Image: testImage, Requested: []string{"view"}}
			},
			want: func() views.Plugin {
				return views.Plugin{
					Instance: "sample", Image: testImage, Digest: "sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1",
					Requested: []string{"view"}, Trust: views.TrustNotChecked, Phase: "Pending",
				}
			},
		},
		{
			name: "being uninstalled",
			record: func() views.PluginRecord {
				r := signed("Unverified", "OpenFiefCI", "openfief-ci")
				r.Deleting = true
				return r
			},
			want: func() views.Plugin {
				p := running(views.TrustOpenFiefCI, "openfief-ci")
				p.Phase, p.Ready = "Removing", false
				return p
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := tt.record()
			if diff := cmp.Diff(tt.want(), views.PluginOf(&r)); diff != "" {
				t.Errorf("PluginOf (-want +got):\n%s", diff)
			}
		})
	}
}

func TestImageDigest(t *testing.T) {
	t.Parallel()
	tests := []struct{ image, want string }{
		{testImage, "sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"},
		{"registry.example.com/openfief/sample:1.0", ""},
		{"registry.example.com/openfief/sample@latest", ""},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.image, func(t *testing.T) {
			t.Parallel()
			if got := views.ImageDigest(tt.image); got != tt.want {
				t.Errorf("ImageDigest(%q) = %q, want %q", tt.image, got, tt.want)
			}
		})
	}
}

func TestConnectorInstance(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		want   string
		ok     bool
	}{
		{ingest.ConnectorKeySource("ibkr-main", "ibkr:acc-1"), "ibkr-main", true},
		{ingest.ConnectorKeySource("bonds", "treasury-bonds:acc-bonds"), "bonds", true},
		{"plugin:ibkr-main", "", false},
		{"plugin::ibkr", "", false},
		{"demo", "", false},
		{"ibkr:plugin:x", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			t.Parallel()
			got, ok := views.ConnectorInstance(tt.source)
			if got != tt.want || ok != tt.ok {
				t.Errorf("ConnectorInstance(%q) = %q, %v; want %q, %v", tt.source, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestOrigins(t *testing.T) {
	t.Parallel()
	txs := []ledger.Transaction{
		{Key: ledger.TransactionKey{Source: ingest.ConnectorKeySource("ibkr-main", "ibkr"), ExternalID: "1"}},
		{Key: ledger.TransactionKey{Source: ingest.ConnectorKeySource("ibkr-old", "ibkr"), ExternalID: "2"}},
		{Key: ledger.TransactionKey{Source: "demo", ExternalID: "3"}},
	}
	tests := []struct {
		name      string
		installed views.Installed
		want      []views.Origin
	}{
		{
			name:      "an instance no longer installed is marked removed",
			installed: views.InstalledOf([]views.Plugin{{Instance: "ibkr-main"}, {Instance: "ui"}}),
			want:      []views.Origin{{Instance: "ibkr-main"}, {Instance: "ibkr-old", Removed: true}, {}},
		},
		{
			name:      "reinstalled under the same name, it is current again",
			installed: views.InstalledOf([]views.Plugin{{Instance: "ibkr-main"}, {Instance: "ui"}, {Instance: "ibkr-old"}}),
			want:      []views.Origin{{Instance: "ibkr-main"}, {Instance: "ibkr-old"}, {}},
		},
		{
			name:      "nothing is marked while the installed plugins are unknown",
			installed: views.Installed{},
			want:      []views.Origin{{Instance: "ibkr-main"}, {Instance: "ibkr-old"}, {}},
		},
		{
			name:      "with no plugin installed, every instance is removed",
			installed: views.InstalledOf(nil),
			want:      []views.Origin{{Instance: "ibkr-main", Removed: true}, {Instance: "ibkr-old", Removed: true}, {}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.installed.Origins(txs)); diff != "" {
				t.Errorf("Origins (-want +got):\n%s", diff)
			}
		})
	}
}
