package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// CABundleStore holds an instance's extra CA bundle, written with the
// owner's own kubeconfig: the Secret <instance>--ca-bundle, which the
// operator never reads, and the checksum stamped on the instance's Plugin,
// which the operator copies to the pod template so a new bundle rolls the
// pod.
type CABundleStore interface {
	// Get reads an installed instance's Plugin and grant. It fails with
	// ErrNotInstalled when either is missing.
	Get(ctx context.Context, name string) (plugins.Plugin, plugins.Grant, error)
	// PutCABundle creates or replaces the instance's CA bundle Secret,
	// holding pem under plugins.CABundleKey and labelled as the
	// instance's (D-182). A Secret of that name labelled as another
	// instance's is left alone and fails with ErrSecretNotOwned.
	PutCABundle(ctx context.Context, instance string, pem []byte) error
	// DeleteCABundle deletes the instance's CA bundle Secret; one already
	// gone is not an error, and one not labelled as the instance's is left
	// alone and fails with ErrSecretNotOwned.
	DeleteCABundle(ctx context.Context, instance string) error
	// StampCABundle sets the checksum on the instance's Plugin
	// (spec.caBundleSha256); an empty one clears it.
	StampCABundle(ctx context.Context, instance, checksum string) error
}

// SetCABundle sets an installed instance's extra CA bundle: it checks the
// file is PEM certificates, that the instance expects a bundle
// (plugins.ErrExtraCANotExpected otherwise, before anything changes), then
// creates or replaces the Secret and stamps its checksum on the Plugin, in
// that order, so the pod the stamp rolls starts with the new bundle.
func SetCABundle(ctx context.Context, store CABundleStore, instance string, pemData []byte) (plugins.CABundle, error) {
	bundle, err := plugins.ParseCABundle(pemData)
	if err != nil {
		return plugins.CABundle{}, fmt.Errorf("ca bundle: %w", err)
	}
	p, _, err := store.Get(ctx, instance)
	if err != nil {
		return plugins.CABundle{}, fmt.Errorf("ca bundle %q: %w", instance, err)
	}
	if err := p.StampCABundle(bundle.Checksum); err != nil {
		return plugins.CABundle{}, fmt.Errorf("ca bundle %q: %w", instance, err)
	}
	if err := store.PutCABundle(ctx, instance, bundle.PEM); err != nil {
		return plugins.CABundle{}, fmt.Errorf("ca bundle %q: %w", instance, err)
	}
	if err := store.StampCABundle(ctx, instance, bundle.Checksum); err != nil {
		return plugins.CABundle{}, fmt.Errorf("ca bundle %q: %w", instance, err)
	}
	return bundle, nil
}

// UnsetCABundle removes an installed instance's extra CA bundle: it deletes
// the Secret, then clears the checksum on the Plugin, which rolls the pod
// without it. While the instance still expects a bundle, its
// CABundleMissing condition says none is set.
func UnsetCABundle(ctx context.Context, store CABundleStore, instance string) error {
	if _, _, err := store.Get(ctx, instance); err != nil {
		return fmt.Errorf("ca bundle %q: %w", instance, err)
	}
	if err := store.DeleteCABundle(ctx, instance); err != nil {
		return fmt.Errorf("ca bundle %q: %w", instance, err)
	}
	if err := store.StampCABundle(ctx, instance, ""); err != nil {
		return fmt.Errorf("ca bundle %q: %w", instance, err)
	}
	return nil
}
