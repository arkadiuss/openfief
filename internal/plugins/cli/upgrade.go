package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
)

// upgrade runs `openfief plugin upgrade` (D-90): it verifies the new image
// as an install does, shows how its requested permissions differ from the
// installed version's, and applies it only when the owner confirms.
func upgrade(ctx context.Context, env *Env, o *options, instance, version string) error {
	cluster, err := env.Cluster(o.kubeconfig, o.context)
	if err != nil {
		return fmt.Errorf("cluster: %w", err)
	}
	var config []byte
	if o.config != "" {
		if config, err = readConfig(env, o.config); err != nil {
			return err
		}
	}
	in, err := installer(ctx, env, o, cluster)
	if err != nil {
		return err
	}
	up, err := in.PrepareUpgrade(ctx, cluster, app.UpgradeRequest{Instance: instance, Version: version, Config: config, ExtraCA: o.extraCAChoice()})
	if err != nil {
		return fmt.Errorf("upgrade: %w", err)
	}
	rule, err := checkClusterService(ctx, env, cluster, &up.Plugin)
	if err != nil {
		return err
	}
	if err := say(env.Out, RenderUpgrade(&up)+rule); err != nil {
		return err
	}
	if err := confirm(env, o, "Upgrade "+instance); err != nil {
		return err
	}
	if err := app.Upgrade(ctx, cluster, &up); err != nil {
		return fmt.Errorf("upgrade: %w", err)
	}
	return say(env.Out, "Upgraded "+instance+". The operator verifies the new image again and replaces the running version.\n")
}

// extraCAChoice is what --extra-ca or --no-extra-ca asked of an upgrade;
// nil keeps what the instance expects.
func (o *options) extraCAChoice() *bool {
	switch {
	case o.extraCA:
		on := true
		return &on
	case o.noExtraCA:
		off := false
		return &off
	}
	return nil
}

// RenderUpgrade describes an upgrade for the owner's confirmation (D-90): the
// instance, the installed and the new image and why the new one is trusted,
// and how the permissions it requests differ; the PluginGrant will approve
// exactly the new version's.
func RenderUpgrade(up *app.UpgradeProposal) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Instance:     %s\n", shown(up.Plugin.Name))
	fmt.Fprintf(&b, "From:         %s\n", shown(up.From.Image))
	fmt.Fprintf(&b, "To:           %s\n", shown(up.Plugin.Image))
	fmt.Fprintf(&b, "Trust:        %s\n", trustText(&up.Proposal))
	if up.Diff.IsEmpty() {
		b.WriteString("\nIt requests the same permissions as the installed version.\n")
	} else {
		b.WriteString("\nIts requested permissions change; the PluginGrant will approve exactly the new ones:\n")
		writePermissions(&b, "  + ", &up.Diff.Added)
		writePermissions(&b, "  - ", &up.Diff.Removed)
	}
	writeShared(&b, up.Shared)
	writeExtraCA(&b, up)
	if r := up.Plugin.Manifest.ClusterService; r != nil {
		writeClusterService(&b, r, up.Plugin.ClusterService)
	}
	config := noneText
	if len(up.Plugin.Config) > 0 {
		config = shown(string(up.Plugin.Config))
	}
	fmt.Fprintf(&b, "\nConfig:       %s\n", config)
	return b.String()
}

// writePermissions writes one line per permission in p, each after prefix.
func writePermissions(b *strings.Builder, prefix string, p *plugins.Permissions) {
	line := func(kind, value string) { fmt.Fprintf(b, "%s%-13s %s\n", prefix, kind, shown(value)) }
	for _, c := range p.Capabilities {
		line("capability", string(c))
	}
	for _, h := range p.Egress {
		line("egress host", h.String())
	}
	for _, s := range p.Secrets {
		line("Secret", s)
	}
	for _, d := range p.DataTiers {
		line("data tier", string(d))
	}
	for _, bp := range p.BrowserPaths {
		line("browser path", bp)
	}
	for _, cs := range p.ClusterServices {
		line("in-cluster", cs.String()+", outside the egress proxy")
	}
	writeDeliveries(line, p)
}

// writeDeliveries writes, with line, what a connector is told and what of
// its deliveries the core accepts: market data (D-123), bond terms and
// index levels (D-179), index readings (D-186) and the institutions it
// books for (D-185).
func writeDeliveries(line func(kind, value string), p *plugins.Permissions) {
	if p.MarketData {
		line("market data", "each sync tells it every instrument you hold and the model's targets (D-123)")
	}
	if p.BondTerms {
		line("bond terms", "the core stores the treasury bond series terms it delivers, which value your bonds (D-179)")
	}
	if p.IndexLevels {
		line("index levels", "the core stores the market index levels it delivers (D-179)")
	}
	if p.IndexReadings {
		line("readings", "the core stores the index readings it delivers, such as CPI, which value your bonds (D-186)")
	}
	for _, in := range p.Institutions {
		line("institution", in+", which the core accepts its deliveries for (D-185)")
	}
}

// writeShared warns about each institution another installed instance
// already books for, naming that instance (D-185). Both may book for it, as
// two accounts at one broker do; the owner decides.
func writeShared(b *strings.Builder, shared []plugins.SharedInstitution) {
	for _, s := range shared {
		fmt.Fprintf(b, "\nWarning: %s already books for %s. Both instances will book for it; make sure that is what you want (D-185).\n",
			shown(s.Instance), shown(s.Institution))
	}
}

// writeExtraCA says, for a new version whose manifest asks for the CA bundle
// role, whether the instance expects an extra CA bundle, and what a change
// of it means; nothing for one that doesn't ask for the role.
func writeExtraCA(b *strings.Builder, up *app.UpgradeProposal) {
	p := &up.Plugin
	if !p.Manifest.AsksForCABundle() {
		return
	}
	switch {
	case p.ExtraCA && !up.From.ExtraCA:
		fmt.Fprintf(b, "\nExtra CA:     expected from now on: set it with `openfief plugin ca-bundle set %s <pem-file>`\n", shown(p.Name))
	case p.ExtraCA:
		fmt.Fprintf(b, "\nExtra CA:     expected, %s\n", stampText(p.CABundleSHA256))
	case up.From.ExtraCA:
		b.WriteString("\nExtra CA:     no longer expected: the bundle isn't mounted and its stamp is cleared\n")
	default:
		b.WriteString("\nExtra CA:     not expected; --extra-ca expects one\n")
	}
}

// stampText says whether a CA bundle checksum is stamped.
func stampText(checksum string) string {
	if checksum == "" {
		return "none set yet"
	}
	return "set (sha256 " + checksum + ")"
}
