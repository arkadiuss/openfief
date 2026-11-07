package cli_test

import (
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/cli"
)

// TestRenderProposalOptionalSecret checks that the install tells the owner
// which Secrets the plugin runs without, such as a connector's extra CA
// bundle, and marks no required Secret that way.
func TestRenderProposalOptionalSecret(t *testing.T) {
	t.Parallel()
	p := &app.Proposal{
		Plugin: plugins.Plugin{Name: "rates", Image: sampleImage, Manifest: plugins.Manifest{
			Capabilities: []plugins.Capability{plugins.CapabilityConnector},
			Secrets: []plugins.SecretMount{
				{Role: "api-token", MountPath: "/etc/openfief/rates"},
				{Role: "ca-bundle", MountPath: "/etc/openfief/ca-bundle", Optional: true},
			},
		}},
		Decision: plugins.TrustDecision{Trust: plugins.Unverified, Signer: plugins.OpenFiefCI()},
	}
	got := cli.RenderProposal(p)
	for line, want := range map[string]bool{
		"                rates--ca-bundle (role ca-bundle) at /etc/openfief/ca-bundle, optional: the plugin runs without it\n": true,
		"                rates--api-token (role api-token) at /etc/openfief/rates\n":                                           true,
		"rates--api-token (role api-token) at /etc/openfief/rates, optional":                                                   false,
	} {
		if strings.Contains(got, line) != want {
			t.Errorf("summary contains %q = %v, want %v; summary:\n%s", line, !want, want, got)
		}
	}
}
