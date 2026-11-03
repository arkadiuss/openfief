package kube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// The owner's trusted signers (D-45), where the operator reads them: the
// optional ConfigMap deploy/operator mounts, and its key.
const (
	TrustedSignersConfigMap = "openfief-trusted-signers"
	TrustedSignersKey       = "signers.json"
)

// Instances is the openfief CLI's InstanceStore: the Plugin and PluginGrant
// resources and the owner's Secrets in the openfief namespace, read and
// written with the owner's own kubeconfig (D-46). The client's scheme must
// hold v1alpha1 and the core types.
type Instances struct {
	Client client.Client
}

var _ app.InstanceStore = (*Instances)(nil)

// List returns every instance with a Plugin, a PluginGrant or both, sorted
// by name.
func (s *Instances) List(ctx context.Context) ([]app.Instance, error) {
	var ps v1alpha1.PluginList
	if err := s.Client.List(ctx, &ps, client.InNamespace(plugins.Namespace)); err != nil {
		return nil, fmt.Errorf("list plugins: %w", err)
	}
	var gs v1alpha1.PluginGrantList
	if err := s.Client.List(ctx, &gs, client.InNamespace(plugins.Namespace)); err != nil {
		return nil, fmt.Errorf("list plugin grants: %w", err)
	}
	byName := map[string]*app.Instance{}
	get := func(name string) *app.Instance {
		if byName[name] == nil {
			byName[name] = &app.Instance{Name: name}
		}
		return byName[name]
	}
	for i := range ps.Items {
		describePlugin(get(ps.Items[i].Name), &ps.Items[i])
	}
	for i := range gs.Items {
		g := &gs.Items[i]
		in := get(g.Name)
		in.HasGrant = true
		in.Secrets = append(in.Secrets, g.Spec.Secrets...)
	}
	owners, err := s.secretOwners(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]app.Instance, 0, len(byName))
	for _, in := range byName {
		slices.Sort(in.Secrets)
		in.Secrets = slices.Compact(in.Secrets)
		in.SecretOwners = ownersOf(in.Secrets, owners)
		out = append(out, *in)
	}
	slices.SortFunc(out, func(a, b app.Instance) int { return strings.Compare(a.Name, b.Name) })
	return out, nil
}

// describePlugin copies what a Plugin says about its instance: its image,
// browser path, the Secret each role names, its in-cluster service and
// what the operator reported.
func describePlugin(in *app.Instance, p *v1alpha1.Plugin) {
	in.HasPlugin, in.Image, in.BrowserPath = true, p.Spec.Image, p.Spec.Manifest.BrowserPath
	in.Institutions = slices.Clone(p.Spec.Manifest.Institutions)
	for _, sm := range p.Spec.Manifest.Secrets {
		in.Secrets = append(in.Secrets, plugins.SecretName(p.Name, sm.Role))
	}
	if r := p.Spec.ClusterService; r != nil {
		if cs, err := clusterServiceOf(r); err == nil {
			in.ClusterService = &cs
		}
	}
	describeStatus(in, &p.Status)
}

// describeStatus copies what the operator reported about a Plugin: an
// image refused for its untrusted signers names them (D-169).
func describeStatus(in *app.Instance, st *v1alpha1.PluginStatus) {
	if c := meta.FindStatusCondition(st.Conditions, v1alpha1.ConditionSignatureVerified); c != nil {
		in.Signature = c.Reason
		if c.Status == metav1.ConditionFalse && len(st.UntrustedSigners) > 0 {
			in.Signature += " (" + strings.Join(st.UntrustedSigners, ", ") + ")"
		}
	}
	if c := meta.FindStatusCondition(st.Conditions, v1alpha1.ConditionReady); c != nil {
		in.Ready = string(c.Status)
		if c.Status != metav1.ConditionTrue && c.Reason != "" {
			in.Ready += " (" + c.Reason + ")"
		}
	}
	in.Pending = st.PendingApproval != nil
}

// Create applies a new instance: with an in-cluster service (D-74), the Role
// and RoleBinding that let the operator read that Service; then its Plugin,
// which runs nothing without a grant; then its PluginGrant. It fails with
// app.ErrInstanceExists when either already exists.
func (s *Instances) Create(ctx context.Context, p *plugins.Plugin, g *plugins.Grant) error {
	plugin, grant := Resources(p, g)
	for _, obj := range []client.Object{&v1alpha1.Plugin{}, &v1alpha1.PluginGrant{}} {
		err := s.Client.Get(ctx, client.ObjectKey{Namespace: plugins.Namespace, Name: p.Name}, obj)
		switch {
		case err == nil:
			return fmt.Errorf("%w: %q", app.ErrInstanceExists, p.Name)
		case !apierrors.IsNotFound(err):
			return fmt.Errorf("look up %q: %w", p.Name, err)
		}
	}
	if p.ClusterService != nil {
		if err := s.createServiceAccess(ctx, p.Name, *p.ClusterService); err != nil {
			return err
		}
	}
	if err := s.Client.Create(ctx, plugin); err != nil {
		return fmt.Errorf("create plugin %q: %w", p.Name, err)
	}
	if err := s.Client.Create(ctx, grant); err != nil {
		return fmt.Errorf("create plugin grant %q: %w", p.Name, err)
	}
	return nil
}

// Delete removes an instance: its PluginGrant first, so the operator stops
// it at once, then its Plugin, whose owned workload objects the garbage
// collector deletes and whose egress allow-list entry the operator drops
// (D-58), then the named Secrets that are labelled as the instance's
// (D-182) and, when the Plugin names an in-cluster service, the Role and
// RoleBinding that let the operator read it (D-74). Objects already gone are
// skipped; a Secret not labelled as the instance's is left alone and
// reported.
func (s *Instances) Delete(ctx context.Context, name string, secrets []string) error {
	access, hasAccess, err := s.clusterServiceOfPlugin(ctx, name)
	if err != nil {
		return err
	}
	own := []client.Object{
		&v1alpha1.PluginGrant{ObjectMeta: metav1.ObjectMeta{Namespace: plugins.Namespace, Name: name}},
		&v1alpha1.Plugin{ObjectMeta: metav1.ObjectMeta{Namespace: plugins.Namespace, Name: name}},
	}
	var errs []error
	for _, obj := range own {
		if err := s.Client.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			errs = append(errs, fmt.Errorf("delete %T %q: %w", obj, obj.GetName(), err))
		}
	}
	for _, sec := range secrets {
		if err := s.deleteOwnedSecret(ctx, name, sec); err != nil {
			errs = append(errs, err)
		}
	}
	if hasAccess {
		errs = append(errs, s.deleteServiceAccess(ctx, name, access)...)
	}
	return errors.Join(errs...)
}

// secretOwners maps every Secret in the openfief namespace to the instance
// its plugins.OwnerInstanceLabel names, empty when it has none (D-182). It
// reads only the Secrets' metadata, never their data.
func (s *Instances) secretOwners(ctx context.Context) (map[string]string, error) {
	list := &metav1.PartialObjectMetadataList{}
	list.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("SecretList"))
	if err := s.Client.List(ctx, list, client.InNamespace(plugins.Namespace)); err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	out := make(map[string]string, len(list.Items))
	for i := range list.Items {
		out[list.Items[i].Name] = list.Items[i].Labels[plugins.OwnerInstanceLabel]
	}
	return out, nil
}

// ownersOf returns, for each of names that owners holds, the instance its
// label names; nil when owners holds none of them.
func ownersOf(names []string, owners map[string]string) map[string]string {
	var out map[string]string
	for _, name := range names {
		if owner, ok := owners[name]; ok {
			if out == nil {
				out = map[string]string{}
			}
			out[name] = owner
		}
	}
	return out
}

// deleteOwnedSecret deletes the Secret name when its
// plugins.OwnerInstanceLabel names instance, and only the version it read, so
// a Secret relabelled in between is not deleted (D-182). One already gone
// is not an error; one not labelled as instance's is left alone and
// reported with app.ErrSecretNotOwned.
func (s *Instances) deleteOwnedSecret(ctx context.Context, instance, name string) error {
	md := &metav1.PartialObjectMetadata{}
	md.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))
	err := s.Client.Get(ctx, client.ObjectKey{Namespace: plugins.Namespace, Name: name}, md)
	switch {
	case apierrors.IsNotFound(err):
		return nil
	case err != nil:
		return fmt.Errorf("read Secret %q: %w", name, err)
	case md.Labels[plugins.OwnerInstanceLabel] != instance:
		return fmt.Errorf("%w: %s, label %s=%s to delete it", app.ErrSecretNotOwned, name, plugins.OwnerInstanceLabel, instance)
	}
	sec := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: plugins.Namespace, Name: name}}
	pre := client.Preconditions{UID: &md.UID, ResourceVersion: &md.ResourceVersion}
	if err := s.Client.Delete(ctx, sec, pre); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete Secret %q: %w", name, err)
	}
	return nil
}

// OwnerSigners reads the owner's trusted signers from the ConfigMap the
// operator mounts, so the CLI trusts exactly whom the operator trusts
// (D-45, D-56). A missing ConfigMap or key adds no signer; an invalid
// document fails, as it stops the operator.
func OwnerSigners(ctx context.Context, c client.Reader) ([]plugins.Signer, error) {
	var cm corev1.ConfigMap
	err := c.Get(ctx, client.ObjectKey{Namespace: plugins.Namespace, Name: TrustedSignersConfigMap}, &cm)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", TrustedSignersConfigMap, err)
	}
	data, ok := cm.Data[TrustedSignersKey]
	if !ok {
		return nil, nil
	}
	signers, err := plugins.ParseOwnerSigners([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", TrustedSignersConfigMap, err)
	}
	return signers, nil
}

// Resources returns an install's Plugin and PluginGrant resources in the
// openfief namespace.
func Resources(p *plugins.Plugin, g *plugins.Grant) (*v1alpha1.Plugin, *v1alpha1.PluginGrant) {
	m := &p.Manifest
	plugin := &v1alpha1.Plugin{
		TypeMeta:   metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion().String(), Kind: "Plugin"},
		ObjectMeta: metav1.ObjectMeta{Namespace: plugins.Namespace, Name: p.Name},
		Spec: v1alpha1.PluginSpec{
			Image: p.Image,
			Manifest: v1alpha1.Manifest{
				Capabilities:     strs(m.Capabilities),
				Egress:           hosts(m.Egress),
				DataTiers:        strs(m.DataTiers),
				ConfigSchema:     raw(m.ConfigSchema),
				BrowserPath:      m.BrowserPath,
				EgressFromConfig: slices.Clone(m.EgressFromConfig),
				CrossCheckPrices: m.CrossCheckPrices,
				StatementFiles:   m.StatementFiles,
				MarketData:       m.MarketData,
				BondTerms:        m.BondTerms,
				IndexLevels:      m.IndexLevels,
				IndexReadings:    m.IndexReadings,
				Institutions:     slices.Clone(m.Institutions),
			},
			Config: raw(p.Config),
		},
	}
	if m.ClusterService != nil {
		plugin.Spec.Manifest.ClusterService = &v1alpha1.ClusterServiceRequest{Description: m.ClusterService.Description}
	}
	if p.ClusterService != nil {
		ref := clusterServiceRef(*p.ClusterService)
		plugin.Spec.ClusterService = &ref
	}
	for _, s := range m.Secrets {
		plugin.Spec.Manifest.Secrets = append(plugin.Spec.Manifest.Secrets, v1alpha1.SecretMount{Role: s.Role, MountPath: s.MountPath})
	}
	grant := &v1alpha1.PluginGrant{
		TypeMeta:   metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion().String(), Kind: "PluginGrant"},
		ObjectMeta: metav1.ObjectMeta{Namespace: plugins.Namespace, Name: g.Plugin},
		Spec: v1alpha1.Permissions{
			Capabilities: strs(g.Permissions.Capabilities),
			Egress:       hosts(g.Permissions.Egress),
			Secrets:      slices.Clone(g.Permissions.Secrets),
			DataTiers:    strs(g.Permissions.DataTiers),
			BrowserPaths: slices.Clone(g.Permissions.BrowserPaths),
			MarketData:   g.Permissions.MarketData,
			BondTerms:    g.Permissions.BondTerms,
			IndexLevels:  g.Permissions.IndexLevels,
			// The index readings bond terms depend on (D-186).
			IndexReadings: g.Permissions.IndexReadings,
		},
	}
	// A connector's grant always names its institutions, even none, so
	// what it approves is plain to read (D-185).
	if slices.Contains(m.Capabilities, plugins.CapabilityConnector) {
		institutions := slices.Clone(g.Permissions.Institutions)
		if institutions == nil {
			institutions = []string{}
		}
		grant.Spec.Institutions = &institutions
	}
	for _, cs := range g.Permissions.ClusterServices {
		grant.Spec.ClusterServices = append(grant.Spec.ClusterServices, clusterServiceRef(cs))
	}
	return plugin, grant
}

// ResourcesYAML renders an install's Plugin and PluginGrant as one YAML
// stream, for GitOps (D-46): no status, no server-set fields. With an
// in-cluster service (D-74), the Role and RoleBinding that let the operator
// read it come first.
func ResourcesYAML(p *plugins.Plugin, g *plugins.Grant) ([]byte, error) {
	plugin, grant := Resources(p, g)
	objs := []any{plugin, grant}
	if p.ClusterService != nil {
		role, binding := ServiceAccess(p.Name, *p.ClusterService)
		objs = append([]any{role, binding}, objs...)
	}
	var out []byte
	for _, obj := range objs {
		doc, err := cleanYAML(obj)
		if err != nil {
			return nil, err
		}
		out = append(append(out, "---\n"...), doc...)
	}
	return out, nil
}

// cleanYAML renders obj as YAML without status and creationTimestamp, which
// only the API server sets.
func cleanYAML(obj any) ([]byte, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}
	var doc map[string]any
	if jsonErr := json.Unmarshal(data, &doc); jsonErr != nil {
		return nil, fmt.Errorf("render: %w", jsonErr)
	}
	delete(doc, "status")
	if md, ok := doc["metadata"].(map[string]any); ok {
		delete(md, "creationTimestamp")
	}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}
	return out, nil
}

// strs converts named string values to strings; nil for none.
func strs[T ~string](xs []T) []string {
	if len(xs) == 0 {
		return nil
	}
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		out = append(out, string(x))
	}
	return out
}

// hosts writes egress hosts as "host:port"; nil for none.
func hosts(hs []plugins.EgressHost) []string {
	if len(hs) == 0 {
		return nil
	}
	out := make([]string, 0, len(hs))
	for _, h := range hs {
		out = append(out, h.String())
	}
	return out
}

// raw wraps a JSON document; nil for none.
func raw(doc []byte) *runtime.RawExtension {
	if len(doc) == 0 {
		return nil
	}
	return &runtime.RawExtension{Raw: slices.Clone(doc)}
}
