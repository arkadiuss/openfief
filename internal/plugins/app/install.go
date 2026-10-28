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
	// Plugin or a PluginGrant. Upgrades are the owner's own edit (D-47).
	ErrInstanceExists = errors.New("an instance of that name is already installed")
	// ErrNotInstalled reports an uninstall of a name with neither a Plugin
	// nor a PluginGrant.
	ErrNotInstalled = errors.New("no instance of that name is installed")
	// ErrNotTrusted reports an image the trust policy refuses: the operator
	// would never deploy it (D-57), so the CLI does not install it.
	ErrNotTrusted = errors.New("image refused by the trust policy")
	// ErrConfig reports a config the manifest's config schema rejects.
	ErrConfig = errors.New("config rejected")
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
	// Delete removes an instance's grant, its Plugin and the named Secrets;
	// one already gone is not an error.
	Delete(ctx context.Context, name string, secrets []string) error
}

// InstallRequest is what the owner asked to install: the argument (a
// catalog name or an image reference), the instance name, which may be
// empty, and the config, a JSON document or nothing.
type InstallRequest struct {
	Arg    string
	Name   string
	Config []byte
}

// Proposal is an install ready for the owner's confirmation: what will be
// applied, and why the image is trusted.
type Proposal struct {
	Plugin   plugins.Plugin
	Grant    plugins.Grant
	Decision plugins.TrustDecision
	// Entry is the catalog entry that pins the image; nil when unverified.
	Entry *plugins.CatalogEntry
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
	signed, err := in.Images.Inspect(ctx, src.Image)
	if err != nil {
		return Proposal{}, fmt.Errorf("check image %s: %w", src.Image, err)
	}
	decision, err := in.Policy.Decide(&signed, nil)
	if err != nil {
		return Proposal{}, fmt.Errorf("%w: %w", ErrNotTrusted, err)
	}
	if in.ValidateConfig != nil {
		if cfgErr := in.ValidateConfig(signed.Manifest.ConfigSchema, req.Config); cfgErr != nil {
			return Proposal{}, fmt.Errorf("%w: %w", ErrConfig, cfgErr)
		}
	}
	p, g, err := plugins.PlanInstall(name, &signed, req.Config)
	if err != nil {
		return Proposal{}, fmt.Errorf("plan install: %w", err)
	}
	out := Proposal{Plugin: p, Grant: g, Decision: decision}
	if decision.Trust == plugins.Verified {
		out.Entry = src.Entry
	}
	return out, nil
}

// Install applies a confirmed proposal.
func Install(ctx context.Context, store InstanceStore, p *Proposal) error {
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
