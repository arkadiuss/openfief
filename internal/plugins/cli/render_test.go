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
			Secrets:      []plugins.SecretMount{{Name: "sample-credentials", MountPath: "/etc/openfief/sample"}},
			DataTiers:    []plugins.DataTier{"T1"},
		}},
		Decision: plugins.TrustDecision{Trust: trust, Signer: plugins.OpenFiefCI()},
	}
	if schema != "" {
		p.Plugin.Manifest.ConfigSchema = []byte(schema)
	}
	if trust == plugins.Verified {
		pub := plugins.Signer{
			Name: "example", Kind: plugins.SignerPublisher,
			Identity: plugins.SignerIdentity{Issuer: plugins.GitHubActionsIssuer, Subject: publisherSubj},
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
Trust:        verified plugin "sample": the catalog pins this image, signed by its publisher example (` + publisherSubj + ` (issuer https://token.actions.githubusercontent.com))

It requests, and the PluginGrant will approve, exactly:
  Capabilities: connector
  Egress hosts: api.example.com:443, data.example.org:8443
  Secrets:      mounted read-only; create each in namespace openfief yourself
                sample-credentials at /etc/openfief/sample
  Data tiers:   T1
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
Trust:        unverified: signed by trusted signer openfief-ci (OpenFiefCI, ` + plugins.OpenFiefCISubject + ` (issuer https://token.actions.githubusercontent.com)); the catalog does not pin this image

It requests, and the PluginGrant will approve, exactly:
  Capabilities: connector
  Egress hosts: api.example.com:443, data.example.org:8443
  Secrets:      mounted read-only; create each in namespace openfief yourself
                sample-credentials at /etc/openfief/sample
  Data tiers:   T1
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
		Secrets:  plugins.UninstallSecrets{Delete: []string{"a-cert", "sample-credentials"}, Keep: []string{"shared"}},
	}
	want := `Instance:     sample
Deletes:
  its PluginGrant
  its Plugin, and with it its workload and egress allow-list entry
  Secret a-cert
  Secret sample-credentials
Keeps Secrets another instance uses: shared
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
