package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/cli"
)

// Synthetic image and signer; neither is real.
const (
	sampleImage   = "registry.example.com/example/sample@sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"
	publisherSubj = "https://github.com/example/sample/.github/workflows/release.yml@refs/heads/main"
)

// proposal is an install of sampleImage as name, with schema as its config
// schema, trusted as trust.
func proposal(name, schema string, trust plugins.Trust) *app.Proposal {
	p := &app.Proposal{
		Plugin: plugins.Plugin{Name: name, Image: sampleImage, Manifest: plugins.Manifest{
			Capabilities: []plugins.Capability{plugins.CapabilityConnector},
			Egress:       []plugins.EgressHost{{Host: "api.example.com", Port: 443}, {Host: "data.example.org", Port: 8443}},
			Secrets:      []plugins.SecretMount{{Role: "credentials", MountPath: "/etc/openfief/sample"}},
			DataTiers:    []plugins.DataTier{"T1"},
		}},
		Decision: plugins.TrustDecision{Trust: trust, Signer: plugins.OpenFiefCI()},
	}
	if schema != "" {
		p.Plugin.Manifest.ConfigSchema = []byte(schema)
	}
	if trust == plugins.Verified {
		p.Plugin.Manifest.Institutions = []string{"example-bank", "example-broker"}
		pub := plugins.Signer{
			Name: "example", Kind: plugins.SignerPublisher,
			Identity: plugins.SignerIdentity{
				Issuer: plugins.GitHubActionsIssuer, Subject: publisherSubj,
				Source: &plugins.SignerSource{Repository: "https://github.com/example/sample", Ref: "refs/heads/main"},
			},
		}
		p.Decision.Signer = pub
		p.Entry = &plugins.CatalogEntry{Name: "sample", Publisher: pub}
	}
	return p
}

func TestRenderProposal(t *testing.T) {
	t.Parallel()
	const schema = `{"type":"object","properties":{"start":{"type":"string"},"mode":{"enum":["fast","slow"]},` +
		`"accounts":{"type":"array"},"limit":{"type":["integer","null"]},"other":{}},"required":["accounts"],"additionalProperties":false}`
	tests := []struct {
		name   string
		p      *app.Proposal
		config string
		want   string
	}{
		{
			name: "verified, with a schema and a config", p: proposal("sample", schema, plugins.Verified), config: `{"accounts":[]}`,
			want: `Instance:     sample
Image:        ` + sampleImage + `
Trust:        verified plugin "sample": the catalog pins this image, signed by its publisher example (` + publisherSubj + ` (issuer https://token.actions.githubusercontent.com, source https://github.com/example/sample, ref refs/heads/main))

It requests, and the PluginGrant will approve, exactly:
  Capabilities: connector
  Egress hosts: api.example.com:443, data.example.org:8443
  Secrets:      this instance's own, mounted read-only; create each in namespace openfief yourself
                and label it openfief.io/instance=sample, or uninstall leaves it in place (D-182)
                sample-credentials (role credentials) at /etc/openfief/sample
  Data tiers:   T1
  Institutions: example-bank, example-broker
  Config:       accounts: array, required
                limit: integer or null
                mode: one of fast, slow
                other: any
                start: string
                no other keys

Config given: {"accounts":[]}
`,
		},
		{
			name: "unverified, no schema", p: proposal("second", "", plugins.Unverified),
			want: `Instance:     second
Image:        ` + sampleImage + `
Trust:        unverified: signed by trusted signer openfief-ci (OpenFiefCI, ` + plugins.OpenFiefCISubject + ` (issuer https://token.actions.githubusercontent.com, source https://github.com/arkadiuss/openfief, ref refs/heads/main)); the catalog does not pin this image

It requests, and the PluginGrant will approve, exactly:
  Capabilities: connector
  Egress hosts: api.example.com:443, data.example.org:8443
  Secrets:      this instance's own, mounted read-only; create each in namespace openfief yourself
                and label it openfief.io/instance=second, or uninstall leaves it in place (D-182)
                second-credentials (role credentials) at /etc/openfief/sample
  Data tiers:   T1
  Institutions: none: the core refuses everything it delivers (D-182)
  Config:       no config schema: any config

Config given: none
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.config != "" {
				tt.p.Plugin.Config = []byte(tt.config)
			}
			if diff := cmp.Diff(tt.want, cli.RenderProposal(tt.p)); diff != "" {
				t.Errorf("summary (-want +got):\n%s", diff)
			}
		})
	}
}

// TestRenderProposalMarketData checks that the install shows the owner that
// a market-data connector will learn what they hold (D-123), and that a
// connector that doesn't declare it says nothing of it.
func TestRenderProposalMarketData(t *testing.T) {
	t.Parallel()
	const line = "  Market data:  it is a market-data connector: each sync tells it every instrument you hold\n" +
		"                and the model portfolio's targets, so it learns what you hold (D-123)\n"
	for _, declared := range []bool{true, false} {
		p := &app.Proposal{
			Plugin: plugins.Plugin{Name: "prices", Image: sampleImage, Manifest: plugins.Manifest{
				Capabilities: []plugins.Capability{plugins.CapabilityConnector}, MarketData: declared,
			}},
			Decision: plugins.TrustDecision{Trust: plugins.Unverified, Signer: plugins.OpenFiefCI()},
		}
		if got := cli.RenderProposal(p); strings.Contains(got, line) != declared {
			t.Errorf("declared %v: summary\n%s", declared, got)
		}
	}
}

func TestRenderProposalWithoutPermissions(t *testing.T) {
	t.Parallel()
	p := &app.Proposal{
		Plugin: plugins.Plugin{Name: "parser", Image: sampleImage, Manifest: plugins.Manifest{
			Capabilities: []plugins.Capability{plugins.CapabilityConnector},
			ConfigSchema: []byte(`{"type":"object"}`),
		}},
		Decision: plugins.TrustDecision{Trust: plugins.Unverified, Signer: plugins.OpenFiefCI()},
	}
	want := `  Egress hosts: none: it cannot reach the internet
  Secrets:      none
  Data tiers:   none
  Institutions: none: the core refuses everything it delivers (D-182)
  Config:       a JSON Schema; --print shows it in full
`
	if got := cli.RenderProposal(p); !strings.Contains(got, want) {
		t.Errorf("summary\n%s\ndoes not contain\n%s", got, want)
	}
}

func TestRenderUninstallPlan(t *testing.T) {
	t.Parallel()
	plan := &app.UninstallPlan{
		Instance: app.Instance{Name: "sample", HasPlugin: true, HasGrant: true},
		Secrets:  plugins.UninstallSecrets{Delete: []string{"a-cert", "sample-credentials"}, Keep: []string{"shared"}, NotOwned: []string{"sample-key"}},
	}
	want := `Instance:     sample
Deletes:
  its PluginGrant
  its Plugin, and with it its workload and egress allow-list entry
  Secret a-cert
  Secret sample-credentials
Keeps Secrets another instance uses: shared
Leaves Secrets not labelled openfief.io/instance=sample: sample-key
Keeps the data it brought into the ledger; reinstalling under the same name continues from it.
`
	if diff := cmp.Diff(want, cli.RenderUninstallPlan(plan)); diff != "" {
		t.Errorf("plan (-want +got):\n%s", diff)
	}
}

func TestWriteInstances(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	err := cli.WriteInstances(&out, []app.Instance{
		{Name: "orphan", HasGrant: true},
		{Name: "sample", Image: "registry.example.com/s@sha256:00", HasPlugin: true, HasGrant: true, Signature: "Verified", Ready: "True"},
		{Name: "waiting", Image: "registry.example.com/w@sha256:00", HasPlugin: true, Signature: "SignedByTrustedSigner", Ready: "False (NoGrant)", Pending: true},
	})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	want := `NAME     IMAGE                             SIGNATURE              READY            GRANT  PENDING APPROVAL
orphan   -                                 -                      -                yes    no
sample   registry.example.com/s@sha256:00  Verified               True             yes    no
waiting  registry.example.com/w@sha256:00  SignedByTrustedSigner  False (NoGrant)  no     yes
`
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("table (-want +got):\n%s", diff)
	}
}

// TestRenderProposalDeliveries checks that the install shows the owner a
// connector that declares bond terms or index levels (D-179), or index
// readings (D-186), which the core stores only once the grant approves
// them, that a connector that declares none says nothing of them, and that
// an upgrade adding index readings lists them.
func TestRenderProposalDeliveries(t *testing.T) {
	t.Parallel()
	const (
		terms    = "  Declares:     it delivers treasury bond series terms, which the core stores once approved (D-179)\n"
		levels   = "  Declares:     it delivers market index levels, which the core stores once approved (D-179)\n"
		readings = "  Declares:     it delivers index readings such as CPI, which the core stores once approved (D-186)\n"
	)
	for _, declared := range []bool{true, false} {
		p := &app.Proposal{
			Plugin: plugins.Plugin{Name: "bonds", Image: sampleImage, Manifest: plugins.Manifest{
				Capabilities: []plugins.Capability{plugins.CapabilityConnector}, BondTerms: declared, IndexLevels: declared,
				IndexReadings: declared,
			}},
			Decision: plugins.TrustDecision{Trust: plugins.Unverified, Signer: plugins.OpenFiefCI()},
		}
		got := cli.RenderProposal(p)
		if strings.Contains(got, terms) != declared || strings.Contains(got, levels) != declared || strings.Contains(got, readings) != declared {
			t.Errorf("declared %v: summary\n%s", declared, got)
		}
		up := &app.UpgradeProposal{Proposal: *p, From: p.Plugin, Diff: plugins.PermissionsDiff{Added: plugins.Permissions{IndexReadings: declared}}}
		const added = "  + readings      the core stores the index readings it delivers, such as CPI, which value your bonds (D-186)\n"
		if got := cli.RenderUpgrade(up); strings.Contains(got, added) != declared {
			t.Errorf("declared %v: upgrade summary\n%s", declared, got)
		}
	}
}
