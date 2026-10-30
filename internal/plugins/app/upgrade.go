package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// UpgradeRequest is what the owner asked to upgrade (D-90): the instance,
// the version (a catalog version or a digest-pinned image reference) and a
// new config, or nil to keep the instance's.
type UpgradeRequest struct {
	Instance string
	Version  string
	Config   []byte
}

// UpgradeProposal is an upgrade ready for the owner's confirmation: the new
// version as a Proposal, the version it replaces, and how the permissions
// they request differ.
type UpgradeProposal struct {
	Proposal

	// From is the installed version.
	From plugins.Plugin
	// Diff is what the new version requests beyond the installed one, and
	// what it no longer requests.
	Diff plugins.PermissionsDiff
}

// PrepareUpgrade resolves the version against the catalog and the
// instance's current image, verifies the new image with the operator's
// trust policy (D-56) and checks the config, without changing anything.
// An instance that isn't installed fails with ErrNotInstalled.
func (in *Installer) PrepareUpgrade(ctx context.Context, store InstanceStore, req UpgradeRequest) (UpgradeProposal, error) {
	current, _, err := store.Get(ctx, req.Instance)
	if err != nil {
		return UpgradeProposal{}, fmt.Errorf("upgrade %q: %w", req.Instance, err)
	}
	image, err := plugins.ParseImageRef(current.Image)
	if err != nil {
		return UpgradeProposal{}, fmt.Errorf("upgrade %q: installed image: %w", req.Instance, err)
	}
	src, err := plugins.ResolveUpgradeSource(in.Catalog, image, req.Version)
	if err != nil {
		return UpgradeProposal{}, fmt.Errorf("resolve %q: %w", req.Version, err)
	}
	config := req.Config
	if config == nil {
		config = current.Config
	}
	signed, decision, err := in.verify(ctx, src.Image, config)
	if err != nil {
		return UpgradeProposal{}, err
	}
	p, g, err := plugins.PlanUpgrade(&current, &signed, config)
	if err != nil {
		return UpgradeProposal{}, fmt.Errorf("plan upgrade: %w", err)
	}
	if err := CheckBrowserView(ctx, store, &p); err != nil {
		return UpgradeProposal{}, fmt.Errorf("upgrade %q: %w", req.Instance, err)
	}
	from, to := current.Requested(), p.Requested()
	out := UpgradeProposal{
		Proposal: Proposal{Plugin: p, Grant: g, Decision: decision, ConfigHosts: configHosts(&p)},
		From:     current,
		Diff:     plugins.DiffPermissions(&from, &to),
	}
	if decision.Trust == plugins.Verified {
		out.Entry = src.Entry
	}
	return out, nil
}

// Upgrade applies a confirmed upgrade: the new Plugin and a grant of exactly
// what it requests. Like PrepareUpgrade, it refuses a version that declares
// a browser path while another instance declares one (D-114).
func Upgrade(ctx context.Context, store InstanceStore, up *UpgradeProposal) error {
	if err := CheckBrowserView(ctx, store, &up.Plugin); err != nil {
		return fmt.Errorf("upgrade %q: %w", up.Plugin.Name, err)
	}
	if err := store.Update(ctx, &up.Plugin, &up.Grant); err != nil {
		return fmt.Errorf("upgrade %q: %w", up.Plugin.Name, err)
	}
	return nil
}
