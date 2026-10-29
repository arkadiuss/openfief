package kube

import (
	"context"
	"errors"
	"fmt"
	"slices"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// Get reads an installed instance's Plugin and PluginGrant. It fails with
// app.ErrNotInstalled when either is missing, and when the Plugin can't be
// read as a valid plugin.
func (s *Instances) Get(ctx context.Context, name string) (plugins.Plugin, plugins.Grant, error) {
	var p v1alpha1.Plugin
	var g v1alpha1.PluginGrant
	if err := s.getBoth(ctx, name, &p, &g); err != nil {
		return plugins.Plugin{}, plugins.Grant{}, err
	}
	plugin, err := pluginOf(&p)
	if err != nil {
		return plugins.Plugin{}, plugins.Grant{}, err
	}
	return plugin, grantOf(&g), nil
}

// Update replaces an installed instance's Plugin spec, then its grant's,
// for an upgrade the owner confirmed (D-90). With a new in-cluster service
// it first lets the operator read it, and afterwards takes back the access
// to a service the instance no longer names (D-74).
func (s *Instances) Update(ctx context.Context, p *plugins.Plugin, g *plugins.Grant) error {
	var plugin v1alpha1.Plugin
	var grant v1alpha1.PluginGrant
	if err := s.getBoth(ctx, p.Name, &plugin, &grant); err != nil {
		return err
	}
	old, hadAccess, err := s.clusterServiceOfPlugin(ctx, p.Name)
	if err != nil {
		return err
	}
	if p.ClusterService != nil {
		if err := s.createServiceAccess(ctx, p.Name, *p.ClusterService); err != nil {
			return err
		}
	}
	wantPlugin, wantGrant := Resources(p, g)
	plugin.Spec, grant.Spec = wantPlugin.Spec, wantGrant.Spec
	if err := s.Client.Update(ctx, &plugin); err != nil {
		return fmt.Errorf("update plugin %q: %w", p.Name, err)
	}
	if err := s.Client.Update(ctx, &grant); err != nil {
		return fmt.Errorf("update plugin grant %q: %w", p.Name, err)
	}
	if hadAccess && (p.ClusterService == nil || *p.ClusterService != old) {
		return errors.Join(s.deleteServiceAccess(ctx, p.Name, old)...)
	}
	return nil
}

// getBoth reads the named instance's Plugin and PluginGrant; either missing
// is app.ErrNotInstalled.
func (s *Instances) getBoth(ctx context.Context, name string, p *v1alpha1.Plugin, g *v1alpha1.PluginGrant) error {
	key := client.ObjectKey{Namespace: plugins.Namespace, Name: name}
	for _, obj := range []client.Object{p, g} {
		err := s.Client.Get(ctx, key, obj)
		switch {
		case apierrors.IsNotFound(err):
			return fmt.Errorf("%w: %q", app.ErrNotInstalled, name)
		case err != nil:
			return fmt.Errorf("look up %q: %w", name, err)
		}
	}
	return nil
}

// pluginOf reads a Plugin resource as the domain's plugin.
func pluginOf(r *v1alpha1.Plugin) (plugins.Plugin, error) {
	m := &r.Spec.Manifest
	p := plugins.Plugin{
		Name:  r.Name,
		Image: r.Spec.Image,
		Manifest: plugins.Manifest{
			BrowserPath:      m.BrowserPath,
			EgressFromConfig: m.EgressFromConfig,
			ConfigSchema:     rawOf(m.ConfigSchema),
		},
		Config: rawOf(r.Spec.Config),
	}
	for _, c := range m.Capabilities {
		p.Manifest.Capabilities = append(p.Manifest.Capabilities, plugins.Capability(c))
	}
	for _, d := range m.DataTiers {
		p.Manifest.DataTiers = append(p.Manifest.DataTiers, plugins.DataTier(d))
	}
	for _, sm := range m.Secrets {
		p.Manifest.Secrets = append(p.Manifest.Secrets, plugins.SecretMount{Role: sm.Role, MountPath: sm.MountPath})
	}
	var errs []error
	for _, h := range m.Egress {
		eh, err := plugins.ParseEgressHost(h)
		errs = append(errs, err)
		p.Manifest.Egress = append(p.Manifest.Egress, eh)
	}
	if m.ClusterService != nil {
		p.Manifest.ClusterService = &plugins.ClusterServiceRequest{Description: m.ClusterService.Description}
	}
	if r.Spec.ClusterService != nil {
		cs, err := clusterServiceOf(r.Spec.ClusterService)
		errs = append(errs, err)
		p.ClusterService = &cs
	}
	errs = append(errs, p.Validate())
	if err := errors.Join(errs...); err != nil {
		return plugins.Plugin{}, fmt.Errorf("%w: %q can't be read: %w", app.ErrNotInstalled, r.Name, err)
	}
	return p, nil
}

// grantOf reads a PluginGrant resource as the domain's grant; entries that
// don't parse are left out, which approves less.
func grantOf(r *v1alpha1.PluginGrant) plugins.Grant {
	g := plugins.Grant{Plugin: r.Name, Permissions: plugins.Permissions{
		Secrets: r.Spec.Secrets, BrowserPaths: r.Spec.BrowserPaths,
	}}
	for _, c := range r.Spec.Capabilities {
		g.Permissions.Capabilities = append(g.Permissions.Capabilities, plugins.Capability(c))
	}
	for _, d := range r.Spec.DataTiers {
		g.Permissions.DataTiers = append(g.Permissions.DataTiers, plugins.DataTier(d))
	}
	for _, h := range r.Spec.Egress {
		if eh, err := plugins.ParseEgressHost(h); err == nil {
			g.Permissions.Egress = append(g.Permissions.Egress, eh)
		}
	}
	for i := range r.Spec.ClusterServices {
		if cs, err := clusterServiceOf(&r.Spec.ClusterServices[i]); err == nil {
			g.Permissions.ClusterServices = append(g.Permissions.ClusterServices, cs)
		}
	}
	return g
}

// rawOf returns the JSON document in r, or nil when there is none.
func rawOf(r *runtime.RawExtension) []byte {
	if r == nil || len(r.Raw) == 0 {
		return nil
	}
	return slices.Clone(r.Raw)
}
