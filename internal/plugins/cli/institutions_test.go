package cli_test

import (
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/cli"
)

// TestRenderSharedInstitutions checks that the install summary and an
// upgrade that changes the institutions warn about each institution
// another instance already books for, naming that instance, and that an
// upgrade's diff lists the institutions it adds (D-185).
func TestRenderSharedInstitutions(t *testing.T) {
	t.Parallel()
	const warning = "Warning: ibkr-joint already books for ibkr. Both instances will book for it"
	p := &app.Proposal{
		Plugin: plugins.Plugin{Name: "ibkr-main", Image: sampleImage, Manifest: plugins.Manifest{
			Capabilities: []plugins.Capability{plugins.CapabilityConnector}, Institutions: []string{"ibkr"},
		}},
		Decision: plugins.TrustDecision{Trust: plugins.Unverified, Signer: plugins.OpenFiefCI()},
		Shared:   []plugins.SharedInstitution{{Institution: "ibkr", Instance: "ibkr-joint"}},
	}
	if got := cli.RenderProposal(p); !strings.Contains(got, warning) || !strings.Contains(got, "  Institutions: ibkr\n") {
		t.Errorf("install summary without the warning or the institutions:\n%s", got)
	}
	up := &app.UpgradeProposal{
		Proposal: *p, From: p.Plugin,
		Diff: plugins.PermissionsDiff{Added: plugins.Permissions{Institutions: []string{"ibkr"}}},
	}
	got := cli.RenderUpgrade(up)
	if !strings.Contains(got, warning) || !strings.Contains(got, "  + institution   ibkr, which the core accepts its deliveries for (D-185)") {
		t.Errorf("upgrade summary without the warning or the added institution:\n%s", got)
	}
	p.Shared = nil
	if got := cli.RenderProposal(p); strings.Contains(got, "Warning:") {
		t.Errorf("install summary warns with no shared institution:\n%s", got)
	}
}
