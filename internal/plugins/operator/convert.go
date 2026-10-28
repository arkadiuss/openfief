package operator

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// pluginFromSpec translates one version of a Plugin resource into the
// domain's Plugin and validates it. Every problem is returned, joined.
func pluginFromSpec(name string, spec *v1alpha1.PluginSpec) (*plugins.Plugin, error) {
	egress, errs := parseEgress("manifest.egress", spec.Manifest.Egress)
	p := &plugins.Plugin{
		Name:  name,
		Image: spec.Image,
		Manifest: plugins.Manifest{
			Capabilities: capabilities(spec.Manifest.Capabilities),
			Egress:       egress,
			DataTiers:    dataTiers(spec.Manifest.DataTiers),
			ConfigSchema: rawJSON(spec.Manifest.ConfigSchema),
		},
		Config: rawJSON(spec.Config),
	}
	for _, s := range spec.Manifest.Secrets {
		p.Manifest.Secrets = append(p.Manifest.Secrets, plugins.SecretMount{Name: s.Name, MountPath: s.MountPath})
	}
	if err := p.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("plugin %q: %w", name, err)
	}
	return p, nil
}

// grantFromResource translates a PluginGrant into the domain's Grant for the
// plugin of the same name and validates it.
func grantFromResource(g *v1alpha1.PluginGrant) (*plugins.Grant, error) {
	egress, errs := parseEgress("egress", g.Spec.Egress)
	grant := &plugins.Grant{
		Plugin: g.Name,
		Permissions: plugins.Permissions{
			Capabilities: capabilities(g.Spec.Capabilities),
			Egress:       egress,
			Secrets:      g.Spec.Secrets,
			DataTiers:    dataTiers(g.Spec.DataTiers),
		},
	}
	if err := grant.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("grant %q: %w", g.Name, err)
	}
	return grant, nil
}

// permissionsToAPI translates domain permissions into the API's shape; nil
// when p holds nothing.
func permissionsToAPI(p *plugins.Permissions) *v1alpha1.Permissions {
	if p.IsEmpty() {
		return nil
	}
	out := &v1alpha1.Permissions{Secrets: p.Secrets}
	for _, c := range p.Capabilities {
		out.Capabilities = append(out.Capabilities, string(c))
	}
	for _, h := range p.Egress {
		out.Egress = append(out.Egress, h.String())
	}
	for _, d := range p.DataTiers {
		out.DataTiers = append(out.DataTiers, string(d))
	}
	return out
}

// effectiveToAPI translates effective permissions into the API's shape.
func effectiveToAPI(e *plugins.EffectivePermissions) *v1alpha1.Permissions {
	p := plugins.Permissions{
		Capabilities: e.Capabilities(),
		Egress:       e.Egress(),
		DataTiers:    e.DataTiers(),
	}
	for _, s := range e.Secrets() {
		p.Secrets = append(p.Secrets, s.Name)
	}
	if api := permissionsToAPI(&p); api != nil {
		return api
	}
	return &v1alpha1.Permissions{}
}

// parseEgress parses each egress host, naming a bad one by its index.
func parseEgress(field string, hosts []string) ([]plugins.EgressHost, []error) {
	var errs []error
	out := make([]plugins.EgressHost, 0, len(hosts))
	for i, s := range hosts {
		h, err := plugins.ParseEgressHost(s)
		if err != nil {
			errs = append(errs, &plugins.FieldError{Field: fmt.Sprintf("%s[%d]", field, i), Err: err})
			continue
		}
		out = append(out, h)
	}
	return out, errs
}

func capabilities(cs []string) []plugins.Capability {
	out := make([]plugins.Capability, 0, len(cs))
	for _, c := range cs {
		out = append(out, plugins.Capability(c))
	}
	return out
}

func dataTiers(ds []string) []plugins.DataTier {
	out := make([]plugins.DataTier, 0, len(ds))
	for _, d := range ds {
		out = append(out, plugins.DataTier(d))
	}
	return out
}

// rawJSON returns the JSON document in r, or nil when there is none.
func rawJSON(r *runtime.RawExtension) []byte {
	if r == nil || len(r.Raw) == 0 {
		return nil
	}
	return r.Raw
}
