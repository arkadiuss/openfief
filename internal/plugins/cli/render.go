package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
)

const (
	// noneText marks an empty list in a summary.
	noneText = "none"
	// tablePadding is the space between the columns of a table.
	tablePadding = 2
)

// RenderProposal describes an install for the owner's confirmation (D-46):
// the instance, the image and why it is trusted, and every permission its
// manifest requests, which the grant will approve exactly.
func RenderProposal(p *app.Proposal) string {
	var b strings.Builder
	m := &p.Plugin.Manifest
	fmt.Fprintf(&b, "Instance:     %s\n", p.Plugin.Name)
	fmt.Fprintf(&b, "Image:        %s\n", p.Plugin.Image)
	fmt.Fprintf(&b, "Trust:        %s\n", trustText(p))
	b.WriteString("\nIt requests, and the PluginGrant will approve, exactly:\n")
	fmt.Fprintf(&b, "  Capabilities: %s\n", joinOr(strsOf(m.Capabilities), noneText))
	hosts := make([]string, 0, len(m.Egress)+len(p.ConfigHosts))
	for _, h := range m.Egress {
		hosts = append(hosts, h.String())
	}
	for _, h := range p.ConfigHosts {
		hosts = append(hosts, h.Host.String()+" (the host of config field "+h.Field+")")
	}
	fmt.Fprintf(&b, "  Egress hosts: %s\n", joinOr(hosts, "none: it cannot reach the internet"))
	for _, f := range m.EgressFromConfig {
		if !slices.ContainsFunc(p.ConfigHosts, func(h plugins.ConfigHost) bool { return h.Field == f }) {
			fmt.Fprintf(&b, "                no host from config field %s: the config doesn't set it\n", f)
		}
	}
	if len(m.Secrets) == 0 {
		fmt.Fprintf(&b, "  Secrets:      %s\n", noneText)
	} else {
		b.WriteString("  Secrets:      this instance's own, mounted read-only; create each in namespace openfief yourself\n")
		for _, s := range m.Secrets {
			fmt.Fprintf(&b, "                %s (role %s) at %s\n", plugins.SecretName(p.Plugin.Name, s.Role), s.Role, s.MountPath)
		}
	}
	fmt.Fprintf(&b, "  Data tiers:   %s\n", joinOr(strsOf(m.DataTiers), noneText))
	if m.BrowserPath != "" {
		fmt.Fprintf(&b, "  Browser path: %s, which the core proxies to it for browsers (D-81)\n", m.BrowserPath)
	}
	if r := m.ClusterService; r != nil {
		writeClusterService(&b, r, p.Plugin.ClusterService)
	}
	b.WriteString("  Config:       ")
	b.WriteString(strings.Join(describeSchema(m.ConfigSchema), "\n                "))
	b.WriteString("\n")
	config := noneText
	if len(p.Plugin.Config) > 0 {
		config = string(p.Plugin.Config)
	}
	fmt.Fprintf(&b, "\nConfig given: %s\n", config)
	return b.String()
}

// writeClusterService describes the in-cluster service the manifest asks
// for and the owner picked, with D-74's warning: that traffic skips the
// egress proxy and may be plain HTTP.
func writeClusterService(b *strings.Builder, r *plugins.ClusterServiceRequest, cs *plugins.ClusterService) {
	what := r.Description
	if what == "" {
		what = "one service the owner picks"
	}
	if cs == nil {
		fmt.Fprintf(b, "  In-cluster:   none picked (it asks for %s)\n", what)
		return
	}
	fmt.Fprintf(b, "  In-cluster:   namespace %s, service %s, port %d (%s),\n", cs.Namespace, cs.Service, cs.Port, what)
	b.WriteString("                reached directly at its ClusterIP, not through the egress proxy\n")
	b.WriteString("  WARNING:      traffic to that service may be plain HTTP, unencrypted inside the cluster\n")
}

// RenderIngressRule tells the owner to add rule, a NetworkPolicy in YAML,
// because the in-cluster service's namespace has NetworkPolicies of its own
// that would otherwise drop the plugin's traffic (D-74).
func RenderIngressRule(namespace string, rule []byte) string {
	return "\nNamespace " + namespace + " has NetworkPolicies of its own. Add this rule there so the plugin's\n" +
		"traffic reaches the service (kubectl apply -f -):\n" + string(rule)
}

// trustText says why the image may run (D-44, D-45, D-56).
func trustText(p *app.Proposal) string {
	s := p.Decision.Signer
	if p.Decision.Trust == plugins.Verified && p.Entry != nil {
		return fmt.Sprintf("verified plugin %q: the catalog pins this image, signed by its publisher %s (%s)",
			p.Entry.Name, s.Name, s.Identity)
	}
	return fmt.Sprintf("unverified: signed by trusted signer %s (%s, %s); the catalog does not pin this image",
		s.Name, s.Kind, s.Identity)
}

// describeSchema summarizes a config JSON Schema: each top-level property
// with its type and whether it is required.
func describeSchema(schema []byte) []string {
	if len(schema) == 0 {
		return []string{"no config schema: any config"}
	}
	var doc struct {
		Properties           map[string]json.RawMessage `json:"properties"`
		Required             []string                   `json:"required"`
		AdditionalProperties *bool                      `json:"additionalProperties"`
	}
	if json.Unmarshal(schema, &doc) != nil || len(doc.Properties) == 0 {
		return []string{"a JSON Schema; --print shows it in full"}
	}
	names := make([]string, 0, len(doc.Properties))
	for n := range doc.Properties {
		names = append(names, n)
	}
	slices.Sort(names)
	out := make([]string, 0, len(names)+1)
	for _, n := range names {
		line := n + ": " + propertyType(doc.Properties[n])
		if slices.Contains(doc.Required, n) {
			line += ", required"
		}
		out = append(out, line)
	}
	if doc.AdditionalProperties != nil && !*doc.AdditionalProperties {
		out = append(out, "no other keys")
	}
	return out
}

// propertyType names a schema property's type, or its allowed values.
func propertyType(raw json.RawMessage) string {
	var p struct {
		Type any   `json:"type"`
		Enum []any `json:"enum"`
	}
	if json.Unmarshal(raw, &p) != nil {
		return "any"
	}
	if len(p.Enum) > 0 {
		vals := make([]string, 0, len(p.Enum))
		for _, v := range p.Enum {
			vals = append(vals, fmt.Sprint(v))
		}
		return "one of " + strings.Join(vals, ", ")
	}
	switch t := p.Type.(type) {
	case string:
		return t
	case []any:
		ts := make([]string, 0, len(t))
		for _, x := range t {
			ts = append(ts, fmt.Sprint(x))
		}
		return strings.Join(ts, " or ")
	}
	return "any"
}

// RenderUninstallPlan describes an uninstall for the owner's confirmation
// (D-58).
func RenderUninstallPlan(plan *app.UninstallPlan) string {
	var b strings.Builder
	in := &plan.Instance
	fmt.Fprintf(&b, "Instance:     %s\n", in.Name)
	b.WriteString("Deletes:\n")
	if in.HasGrant {
		b.WriteString("  its PluginGrant\n")
	}
	if in.HasPlugin {
		b.WriteString("  its Plugin, and with it its workload and egress allow-list entry\n")
	}
	for _, s := range plan.Secrets.Delete {
		fmt.Fprintf(&b, "  Secret %s\n", s)
	}
	if len(plan.Secrets.Keep) > 0 {
		fmt.Fprintf(&b, "Keeps Secrets another instance uses: %s\n", strings.Join(plan.Secrets.Keep, ", "))
	}
	b.WriteString("Keeps the data it brought into the ledger; reinstalling under the same name continues from it.\n")
	return b.String()
}

// WriteInstances writes the installed instances as a table.
func WriteInstances(w io.Writer, instances []app.Instance) error {
	var b strings.Builder
	b.WriteString("NAME\tIMAGE\tSIGNATURE\tREADY\tGRANT\tPENDING APPROVAL\n")
	for i := range instances {
		in := &instances[i]
		b.WriteString(strings.Join([]string{
			in.Name, orDash(in.Image), orDash(in.Signature), orDash(in.Ready), yesNo(in.HasGrant), yesNo(in.Pending),
		}, "\t") + "\n")
	}
	tw := tabwriter.NewWriter(w, 0, 0, tablePadding, ' ', 0)
	if _, err := io.WriteString(tw, b.String()); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

func strsOf[T ~string](xs []T) []string {
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		out = append(out, string(x))
	}
	return out
}

func joinOr(xs []string, empty string) string {
	if len(xs) == 0 {
		return empty
	}
	return strings.Join(xs, ", ")
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
