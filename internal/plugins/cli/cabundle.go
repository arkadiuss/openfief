package cli

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
)

// The subcommands of `openfief plugin ca-bundle`.
const (
	caBundleSet   = "set"
	caBundleUnset = "unset"
	// caBundleSetArgs are set, an instance and a PEM file; unset takes an
	// instance.
	caBundleSetArgs   = 3
	caBundleUnsetArgs = 2
)

// caBundleArgs checks the arguments of `openfief plugin ca-bundle`: set
// with an instance and a PEM file, or unset with an instance.
func caBundleArgs(pos []string) bool {
	switch {
	case len(pos) == caBundleSetArgs && pos[0] == caBundleSet:
		return true
	case len(pos) == caBundleUnsetArgs && pos[0] == caBundleUnset:
		return true
	}
	return false
}

// caBundle runs `openfief plugin ca-bundle set|unset`.
func caBundle(ctx context.Context, env *Env, o *options, pos []string) error {
	if pos[0] == caBundleUnset {
		return unsetCABundle(ctx, env, o, pos[1])
	}
	return setCABundle(ctx, env, o, pos[1], pos[2])
}

// setCABundle runs `openfief plugin ca-bundle set <instance> <pem-file>`:
// it checks the file is PEM certificates, creates or replaces the Secret
// <instance>--ca-bundle and stamps its checksum on the Plugin, so the
// operator, which never reads the Secret, restarts the plugin with it. With
// --print it writes both for GitOps instead, and says that the owner keeps
// them in step.
func setCABundle(ctx context.Context, env *Env, o *options, instance, path string) error {
	data, err := env.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ca bundle: %w", err)
	}
	if o.print {
		return printCABundle(env, instance, data)
	}
	cluster, err := env.Cluster(o.kubeconfig, o.context)
	if err != nil {
		return fmt.Errorf("cluster: %w", err)
	}
	bundle, err := app.SetCABundle(ctx, cluster, instance, data)
	if err != nil {
		return fmt.Errorf("ca bundle set: %w", err)
	}
	return say(env.Out, fmt.Sprintf("Set %s: %d certificate(s), sha256 %s, stamped on Plugin %s as spec.caBundleSha256.\n"+
		"The operator restarts the plugin with the new bundle.\n",
		plugins.SecretName(instance, plugins.CABundleRole), bundle.Certificates, bundle.Checksum, shown(instance)))
}

// printCABundle writes, for GitOps, the Secret and the Plugin stamp that
// set would apply, and to Err why they must stay in step.
func printCABundle(env *Env, instance string, data []byte) error {
	bundle, err := plugins.ParseCABundle(data)
	if err != nil {
		return fmt.Errorf("ca bundle: %w", err)
	}
	out, err := env.RenderCABundle(instance, bundle.PEM, bundle.Checksum)
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}
	if err := say(env.Err, RenderCABundleNote(instance, &bundle)); err != nil {
		return err
	}
	if _, err := env.Out.Write(out); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// RenderCABundleNote is what --print says next to the YAML: the operator
// never reads the Secret, so the Plugin's spec.caBundleSha256 must stay the
// SHA-256 of the Secret's ca.crt, or a rotated bundle doesn't restart the
// plugin.
func RenderCABundleNote(instance string, b *plugins.CABundle) string {
	return fmt.Sprintf("%d certificate(s), sha256 %s.\n"+
		"Apply both documents together: the Secret %s and spec.caBundleSha256 on Plugin %s.\n"+
		"Keep the checksum in step with the Secret (sha256sum of its %s): the operator never reads the Secret,\n"+
		"so it restarts the plugin only when the checksum changes, and a bundle that doesn't match shows only\n"+
		"as the plugin's TLS failures. The Plugin must expect a bundle (spec.extraCa: true).\n",
		b.Certificates, b.Checksum, plugins.SecretName(instance, plugins.CABundleRole), shown(instance), plugins.CABundleKey)
}

// unsetCABundle runs `openfief plugin ca-bundle unset <instance>`: it
// deletes the Secret and clears the stamp, once the owner confirms.
func unsetCABundle(ctx context.Context, env *Env, o *options, instance string) error {
	cluster, err := env.Cluster(o.kubeconfig, o.context)
	if err != nil {
		return fmt.Errorf("cluster: %w", err)
	}
	secret := plugins.SecretName(instance, plugins.CABundleRole)
	if err := say(env.Out, fmt.Sprintf("Deletes Secret %s and clears spec.caBundleSha256 on Plugin %s; the plugin restarts without the bundle.\n",
		secret, shown(instance))); err != nil {
		return err
	}
	if err := confirm(env, o, "Unset the CA bundle of "+instance); err != nil {
		return err
	}
	if err := app.UnsetCABundle(ctx, cluster, instance); err != nil {
		return fmt.Errorf("ca bundle unset: %w", err)
	}
	return say(env.Out, "Unset the CA bundle of "+instance+".\n")
}
