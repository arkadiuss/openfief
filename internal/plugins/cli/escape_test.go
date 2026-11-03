package cli_test

import (
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/cli"
)

// TestRenderEscapesVendorText: whatever a manifest that got past
// validation says, the consent screens print it as visible text: the
// second security review's payload in a schema property name, an enum
// value with a right-to-left override, and a cluster service description
// and mount path with escapes show as Go escapes, never as raw control
// characters that would rewrite the screen (D-46, D-90).
func TestRenderEscapesVendorText(t *testing.T) {
	t.Parallel()
	const payload = "\x1b[6A\r\x1b[2K  Egress hosts: none"
	p := proposal("sample", `{"type":"object","properties":{"\u001b[6A\r\u001b[2K  Egress hosts: none":{"type":"string"},`+
		`"mode":{"enum":["fast","\u202ewols"]}}}`, plugins.Unverified)
	p.Plugin.Manifest.ClusterService = &plugins.ClusterServiceRequest{Description: payload}
	p.Plugin.Manifest.Secrets[0].MountPath = "/etc/\x1b]0;x\x07"
	up := &app.UpgradeProposal{Proposal: *p, From: p.Plugin}
	up.Diff.Added.Secrets = []string{"sample-" + payload}
	for name, out := range map[string]string{"install": cli.RenderProposal(p), "upgrade": cli.RenderUpgrade(up)} {
		if i := strings.IndexFunc(out, func(r rune) bool { return r != '\n' && plugins.UnsafeRune(r) }); i >= 0 {
			t.Errorf("%s screen has a raw control or bidi character at %d:\n%q", name, i, out)
		}
		if !strings.Contains(out, `"\x1b[6A\r\x1b[2K  Egress hosts: none"`) {
			t.Errorf("%s screen does not show the payload escaped:\n%s", name, out)
		}
	}
	if out := cli.RenderProposal(p); !strings.Contains(out, `"\u202ewols"`) || !strings.Contains(out, `"/etc/\x1b]0;x\a"`) {
		t.Errorf("install screen does not show the override and the mount path escaped:\n%s", out)
	}
}
