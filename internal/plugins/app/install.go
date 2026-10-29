package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// Errors of the install use cases.
var (
	// ErrNoInstanceName reports an install with no instance name given and
	// none to derive from the catalog entry or the image (D-48).
	ErrNoInstanceName = errors.New("name the instance with --name")
	// ErrInstanceExists reports an install under a name that already has a
	// Plugin or a PluginGrant: an installed instance changes version only
	// through `openfief plugin upgrade` (D-47, D-90).
	ErrInstanceExists = errors.New("an instance of that name is already installed")
	// ErrNotInstalled reports an uninstall of a name with neither a Plugin
	// nor a PluginGrant.
	ErrNotInstalled = errors.New("no instance of that name is installed")
	// ErrNotTrusted reports an image the trust policy refuses: the operator
	// would never deploy it (D-57), so the CLI does not install it.
	ErrNotTrusted = errors.New("image refused by the trust policy")
	// ErrConfig reports a config the manifest's config schema rejects.
	ErrConfig = errors.New("config rejected")
	// ErrSecretInUse reports an install whose instance would mount a
	// Secret another installed instance already names: two instances never
	// share a Secret (D-89).
	ErrSecretInUse = errors.New("another instance already names that Secret")
	// ErrClusterService reports an in-cluster service the owner picked that
	// can't be reached as D-74 needs: no such port, or no pods selected.
	ErrClusterService = errors.New("in-cluster service unusable")
)

// ImageInspector reads an image from its registry: the trusted signatures
// over its digest that verify, and its signed manifest label. It is the same
// check the operator runs (D-56); cosign.Verifier implements it.
type ImageInspector interface {
	Inspect(ctx context.Context, image plugins.ImageRef) (plugins.SignedImage, error)
}

// Instance is one installed plugin instance as the cluster holds it: its
// Plugin, its PluginGrant, or both.
type Instance struct {
	Name string
	// Image is the Plugin's image; empty without a Plugin.
	Image string
	// HasPlugin and HasGrant say which resources exist.
	HasPlugin, HasGrant bool
	// Secrets are the Secret names its manifest or its grant names.
	Secrets []string
	// Signature is the reason of the Plugin's SignatureVerified condition,
	// such as Verified or SignedByTrustedSigner; empty before the operator
	// checked it.
	Signature string
	// Ready is the status of its Ready condition, with the reason when it is
	// not True, such as "False (NoGrant)"; empty before the operator saw it.
	Ready string
	// Pending reports permissions waiting for the owner's approval.
	Pending bool
	// ClusterService is the in-cluster service the Plugin names (D-74);
	// nil when it names none.
	ClusterService *plugins.ClusterService
	// BrowserPath is the browser path the Plugin's manifest declares
	// (D-81); empty when it declares none.
	BrowserPath string
}

// CheckBrowserView refuses p, about to be installed or to replace its
// instance's version, when it declares a browser path while another
// installed instance declares one: at most one instance holds a browser
// path (D-114). The error wraps plugins.ErrBrowserPathHeld.
func CheckBrowserView(ctx context.Context, store InstanceStore, p *plugins.Plugin) error {
	if p.Manifest.BrowserPath == "" {
		return nil
	}
	all, err := store.List(ctx)
	if err != nil {
		return fmt.Errorf("list instances: %w", err)
	}
	return checkBrowserView(all, p)
}

// checkBrowserView is CheckBrowserView over the installed instances all.
func checkBrowserView(all []Instance, p *plugins.Plugin) error {
	var held []plugins.BrowserView
	for i := range all {
		if all[i].BrowserPath != "" {
			held = append(held, plugins.BrowserView{Instance: all[i].Name, Path: all[i].BrowserPath})
		}
	}
	if err := plugins.RefuseSecondBrowserView(p.Name, p.Manifest.BrowserPath, held); err != nil {
		return fmt.Errorf("%s declares browser path %s: %w", p.Name, p.Manifest.BrowserPath, err)
	}
	return nil
}

// ClusterServiceInfo is what the CLI reads about an in-cluster service the
// owner picked (D-74), with the owner's kubeconfig: the pod port behind the
// Service port, as written (a number or a port name), the Service's pod
// selector, and whether its namespace has NetworkPolicies of its own, in
// which case the owner must add an ingress rule there.
type ClusterServiceInfo struct {
	TargetPort                  string
	Selector                    map[string]string
	NamespaceHasNetworkPolicies bool
}

// InstanceStore holds the installed instances: in a cluster, the Plugin and
// PluginGrant resources and the owner's Secrets, written with the owner's
// own kubeconfig (D-46).
type InstanceStore interface {
	// List returns every installed instance, sorted by name.
	List(ctx context.Context) ([]Instance, error)
	// Create applies a new instance's Plugin and grant. It fails with
	// ErrInstanceExists when either already exists.
	Create(ctx context.Context, p *plugins.Plugin, g *plugins.Grant) error
	// Get reads an installed instance's Plugin and grant. It fails with
	// ErrNotInstalled when either is missing.
	Get(ctx context.Context, name string) (plugins.Plugin, plugins.Grant, error)
	// Update replaces an installed instance's Plugin and grant, for a
	// confirmed upgrade (D-90). It fails with ErrNotInstalled when either is
	// missing.
	Update(ctx context.Context, p *plugins.Plugin, g *plugins.Grant) error
	// Delete removes an instance's grant, its Plugin and the named Secrets;
	// one already gone is not an error.
	Delete(ctx context.Context, name string, secrets []string) error
}

// InstallRequest is what the owner asked to install: the argument (a
// catalog name or an image reference), the instance name, which may be
// empty, the config, a JSON document or nothing, and the in-cluster service
// the owner picked for a manifest that asks for one (D-74), or nil.
type InstallRequest struct {
	Arg            string
	Name           string
	Config         []byte
	ClusterService *plugins.ClusterService
}

// Proposal is an install ready for the owner's confirmation: what will be
// applied, and why the image is trusted.
type Proposal struct {
	Plugin   plugins.Plugin
	Grant    plugins.Grant
	Decision plugins.TrustDecision
	// Entry is the catalog entry that pins the image; nil when unverified.
	Entry *plugins.CatalogEntry
	// ConfigHosts are the egress hosts the config gives, by field (D-75).
	ConfigHosts []plugins.ConfigHost
}

// PickClusterService sets the in-cluster service the owner picked after the
// proposal was prepared (D-74): the Plugin requests it and the grant
// approves exactly what the Plugin now requests.
func (p *Proposal) PickClusterService(cs plugins.ClusterService) error {
	plugin := p.Plugin
	plugin.ClusterService = &cs
	if err := plugin.Validate(); err != nil {
		return fmt.Errorf("in-cluster service: %w", err)
	}
	p.Plugin = plugin
	p.Grant.Permissions = plugin.Requested()
	return nil
}

// Installer prepares installs: it resolves the argument against the
// catalog, verifies the image with the same trust policy as the operator
// (D-45, D-56) and checks the config against the manifest's schema.
type Installer struct {
	Images  ImageInspector
	Policy  *plugins.TrustPolicy
	Catalog *plugins.Catalog
	// ValidateConfig checks a config against a JSON Schema; nil checks
	// nothing beyond the config being JSON.
	ValidateConfig func(schema, config []byte) error
}

// Prepare resolves, verifies and plans an install without changing
// anything. An image the trust policy refuses fails with ErrNotTrusted.
func (in *Installer) Prepare(ctx context.Context, req InstallRequest) (Proposal, error) {
	src, err := plugins.ResolveInstallSource(in.Catalog, req.Arg)
	if err != nil {
		return Proposal{}, fmt.Errorf("resolve %q: %w", req.Arg, err)
	}
	name := req.Name
	if name == "" {
		if name = plugins.DefaultInstanceName(&src); name == "" {
			return Proposal{}, fmt.Errorf("%w: %s", ErrNoInstanceName, src.Image)
		}
	}
	signed, decision, err := in.verify(ctx, src.Image, req.Config)
	if err != nil {
		return Proposal{}, err
	}
	p, g, err := plugins.PlanInstall(name, &signed, req.Config, req.ClusterService)
	if err != nil {
		return Proposal{}, fmt.Errorf("plan install: %w", err)
	}
	out := Proposal{Plugin: p, Grant: g, Decision: decision, ConfigHosts: configHosts(&p)}
	if decision.Trust == plugins.Verified {
		out.Entry = src.Entry
	}
	return out, nil
}

// verify reads an image's signatures and signed manifest label, lets the
// trust policy decide (D-56) and checks config against the manifest's
// schema.
func (in *Installer) verify(ctx context.Context, image plugins.ImageRef, config []byte) (plugins.SignedImage, plugins.TrustDecision, error) {
	signed, err := in.Images.Inspect(ctx, image)
	if err != nil {
		return plugins.SignedImage{}, plugins.TrustDecision{}, fmt.Errorf("check image %s: %w", image, err)
	}
	decision, err := in.Policy.Decide(&signed, nil)
	if err != nil {
		return plugins.SignedImage{}, plugins.TrustDecision{}, fmt.Errorf("%w: %w", ErrNotTrusted, err)
	}
	if in.ValidateConfig != nil {
		if cfgErr := in.ValidateConfig(signed.Manifest.ConfigSchema, config); cfgErr != nil {
			return plugins.SignedImage{}, plugins.TrustDecision{}, fmt.Errorf("%w: %w", ErrConfig, cfgErr)
		}
	}
	return signed, decision, nil
}

// configHosts lists the hosts p's config gives (D-75); none for a config
// Validate refuses, which never gets this far.
func configHosts(p *plugins.Plugin) []plugins.ConfigHost {
	hosts, err := plugins.ConfigHosts(&p.Manifest, p.Config)
	if err != nil {
		return nil
	}
	return hosts
}

// Install applies a confirmed proposal. It refuses one whose Secrets another
// installed instance already names (D-89), and one that declares a browser
// path while another instance declares one (D-114).
func Install(ctx context.Context, store InstanceStore, p *Proposal) error {
	all, err := store.List(ctx)
	if err != nil {
		return fmt.Errorf("install %q: list instances: %w", p.Plugin.Name, err)
	}
	if err := checkBrowserView(all, &p.Plugin); err != nil {
		return fmt.Errorf("install %q: %w", p.Plugin.Name, err)
	}
	for i := range all {
		if all[i].Name == p.Plugin.Name {
			continue // Create refuses it with ErrInstanceExists
		}
		for _, s := range p.Grant.Permissions.Secrets {
			if slices.Contains(all[i].Secrets, s) {
				return fmt.Errorf("install %q: %w: %s, named by %s", p.Plugin.Name, ErrSecretInUse, s, all[i].Name)
			}
		}
	}
	if err := store.Create(ctx, &p.Plugin, &p.Grant); err != nil {
		return fmt.Errorf("install %q: %w", p.Plugin.Name, err)
	}
	return nil
}

// UninstallPlan is what uninstalling an instance removes (D-58): its
// PluginGrant and its Plugin, whose owned workload the cluster's garbage
// collector deletes and whose allow-list entry the operator drops, and the
// Secrets no other instance names. Data the instance brought into the
// ledger is never touched.
type UninstallPlan struct {
	Instance Instance
	Secrets  plugins.UninstallSecrets
}

// PlanUninstall finds the instance and the Secrets its uninstall deletes,
// without changing anything.
func PlanUninstall(ctx context.Context, store InstanceStore, name string) (UninstallPlan, error) {
	all, err := store.List(ctx)
	if err != nil {
		return UninstallPlan{}, fmt.Errorf("list instances: %w", err)
	}
	i := slices.IndexFunc(all, func(in Instance) bool { return in.Name == name })
	if i < 0 {
		return UninstallPlan{}, fmt.Errorf("%w: %q", ErrNotInstalled, name)
	}
	var others []string
	for j := range all {
		if j != i {
			others = append(others, all[j].Secrets...)
		}
	}
	return UninstallPlan{Instance: all[i], Secrets: plugins.SecretsToDelete(all[i].Secrets, others)}, nil
}

// Uninstall carries out a confirmed plan.
func Uninstall(ctx context.Context, store InstanceStore, plan *UninstallPlan) error {
	if err := store.Delete(ctx, plan.Instance.Name, plan.Secrets.Delete); err != nil {
		return fmt.Errorf("uninstall %q: %w", plan.Instance.Name, err)
	}
	return nil
}

// Instances lists the installed instances, sorted by name.
func Instances(ctx context.Context, store InstanceStore) ([]Instance, error) {
	all, err := store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	slices.SortFunc(all, func(a, b Instance) int { return strings.Compare(a.Name, b.Name) })
	return all, nil
}
