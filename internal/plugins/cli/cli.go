// Package cli is the openfief command line (D-46): it installs, uninstalls
// and lists plugin instances with the owner's own kubeconfig. It parses
// arguments, shows what an install or uninstall will do, asks for
// confirmation and calls the plugins context's use cases. cmd/openfief
// wires the adapters it needs through Env.
package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
)

// Usage is the command line's synopsis.
const Usage = `usage:
  openfief plugin install <catalog-name|image@sha256:...> [--name NAME] [--config FILE]
                          [--trusted-signers FILE] [--yes] [--print]
  openfief plugin uninstall <instance> [--yes]
  openfief plugin list
common flags: --kubeconfig PATH, --context NAME`

// Errors of the command line.
var (
	// ErrUsage reports arguments the command line does not accept.
	ErrUsage = errors.New("bad arguments")
	// ErrAborted reports that the owner did not confirm.
	ErrAborted = errors.New("not confirmed; nothing changed")
)

// Cluster is the owner's cluster, reached with their kubeconfig: the
// installed instances, and the owner's trusted signers the operator reads.
type Cluster interface {
	app.InstanceStore
	OwnerSigners(ctx context.Context) ([]plugins.Signer, error)
}

// Env is what the command line needs from outside it.
type Env struct {
	In  io.Reader
	Out io.Writer
	// Err gets messages that must stay out of Out, such as the summary
	// next to --print's YAML.
	Err     io.Writer
	Catalog plugins.Catalog
	// ReadFile reads a file the owner named.
	ReadFile func(path string) ([]byte, error)
	// Cluster connects to the cluster with a kubeconfig path and context,
	// either empty for the defaults.
	Cluster func(kubeconfig, context string) (Cluster, error)
	// Images builds the signature check for a trust policy.
	Images func(policy *plugins.TrustPolicy) (app.ImageInspector, error)
	// ValidateConfig checks a config against a JSON Schema.
	ValidateConfig func(schema, config []byte) error
	// Render renders an install's resources as YAML, for --print.
	Render func(p *plugins.Plugin, g *plugins.Grant) ([]byte, error)
}

// options are the parsed flags of one command.
type options struct {
	kubeconfig, context string
	name, config        string
	trustedSigners      string
	yes, print          bool
}

// Run runs the command line with args, the arguments after the program's
// name.
func Run(ctx context.Context, args []string, env *Env) error {
	if len(args) < 2 || args[0] != "plugin" {
		return fmt.Errorf("%w\n%s", ErrUsage, Usage)
	}
	cmd := args[1]
	o, pos, err := parseFlags(cmd, args[2:])
	if err != nil {
		return err
	}
	switch {
	case cmd == "install" && len(pos) == 1:
		return install(ctx, env, &o, pos[0])
	case cmd == "uninstall" && len(pos) == 1:
		return uninstall(ctx, env, &o, pos[0])
	case cmd == "list" && len(pos) == 0:
		return list(ctx, env, &o)
	}
	return fmt.Errorf("%w\n%s", ErrUsage, Usage)
}

// parseFlags parses one command's flags, which may come before, between or
// after its arguments, and returns the arguments.
func parseFlags(cmd string, args []string) (options, []string, error) {
	var o options
	fs := flag.NewFlagSet("openfief plugin "+cmd, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&o.kubeconfig, "kubeconfig", "", "kubeconfig file; default: $KUBECONFIG or ~/.kube/config")
	fs.StringVar(&o.context, "context", "", "kubeconfig context; default: the current one")
	switch cmd {
	case "install":
		fs.StringVar(&o.name, "name", "", "instance name; default: the catalog name or the image's repository name")
		fs.StringVar(&o.config, "config", "", "JSON file with the instance's config")
		fs.StringVar(&o.trustedSigners, "trusted-signers", "", "the owner's signers file; default: the cluster's openfief-trusted-signers ConfigMap")
		fs.BoolVar(&o.yes, "yes", false, "install without asking")
		fs.BoolVar(&o.print, "print", false, "print the Plugin and PluginGrant as YAML for GitOps instead of applying them")
	case "uninstall":
		fs.BoolVar(&o.yes, "yes", false, "uninstall without asking")
	}
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			return options{}, nil, fmt.Errorf("%w: %w\n%s", ErrUsage, err, Usage)
		}
		if fs.NArg() == 0 {
			return o, pos, nil
		}
		pos = append(pos, fs.Arg(0))
		args = fs.Args()[1:]
	}
}

// install runs `openfief plugin install`.
func install(ctx context.Context, env *Env, o *options, arg string) error {
	config, err := readConfig(env, o.config)
	if err != nil {
		return err
	}
	var cluster Cluster
	if !o.print {
		c, clusterErr := env.Cluster(o.kubeconfig, o.context)
		if clusterErr != nil {
			return fmt.Errorf("cluster: %w", clusterErr)
		}
		cluster = c
	}
	in, err := installer(ctx, env, o, cluster)
	if err != nil {
		return err
	}
	p, err := in.Prepare(ctx, app.InstallRequest{Arg: arg, Name: o.name, Config: config})
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if o.print {
		return printResources(env, &p)
	}
	if err := say(env.Out, RenderProposal(&p)); err != nil {
		return err
	}
	if err := confirm(env, o, "Install "+p.Plugin.Name); err != nil {
		return err
	}
	if err := app.Install(ctx, cluster, &p); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	return say(env.Out, "Installed "+p.Plugin.Name+". The operator verifies the image again and deploys it once its Secrets exist.\n")
}

// readConfig reads the config file the owner named; nothing without one.
func readConfig(env *Env, path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	data, err := env.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return data, nil
}

// installer builds the install use case with the operator's trust policy:
// OpenFief's CI, the catalog's publishers and the owner's signers (D-45).
func installer(ctx context.Context, env *Env, o *options, cluster Cluster) (*app.Installer, error) {
	owner, err := ownerSigners(ctx, env, o, cluster)
	if err != nil {
		return nil, err
	}
	policy, err := plugins.NewTrustPolicy(env.Catalog, owner)
	if err != nil {
		return nil, fmt.Errorf("trust policy: %w", err)
	}
	images, err := env.Images(&policy)
	if err != nil {
		return nil, fmt.Errorf("signature check: %w", err)
	}
	return &app.Installer{Images: images, Policy: &policy, Catalog: &env.Catalog, ValidateConfig: env.ValidateConfig}, nil
}

// ownerSigners reads the owner's signers from the file the owner named,
// else from the cluster; with --print and no file, there are none.
func ownerSigners(ctx context.Context, env *Env, o *options, cluster Cluster) ([]plugins.Signer, error) {
	switch {
	case o.trustedSigners != "":
		data, err := env.ReadFile(o.trustedSigners)
		if err != nil {
			return nil, fmt.Errorf("trusted signers: %w", err)
		}
		signers, err := plugins.ParseOwnerSigners(data)
		if err != nil {
			return nil, fmt.Errorf("trusted signers %s: %w", o.trustedSigners, err)
		}
		return signers, nil
	case cluster != nil:
		signers, err := cluster.OwnerSigners(ctx)
		if err != nil {
			return nil, fmt.Errorf("trusted signers: %w", err)
		}
		return signers, nil
	}
	return nil, say(env.Err, "Trusting OpenFief's CI and the catalog only; --trusted-signers adds your own signers.\n")
}

// printResources writes the summary to Err and the resources to Out.
func printResources(env *Env, p *app.Proposal) error {
	if err := say(env.Err, RenderProposal(p)); err != nil {
		return err
	}
	data, err := env.Render(&p.Plugin, &p.Grant)
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}
	if _, err := env.Out.Write(data); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// uninstall runs `openfief plugin uninstall`.
func uninstall(ctx context.Context, env *Env, o *options, name string) error {
	cluster, err := env.Cluster(o.kubeconfig, o.context)
	if err != nil {
		return fmt.Errorf("cluster: %w", err)
	}
	plan, err := app.PlanUninstall(ctx, cluster, name)
	if err != nil {
		return fmt.Errorf("uninstall: %w", err)
	}
	if err := say(env.Out, RenderUninstallPlan(&plan)); err != nil {
		return err
	}
	if err := confirm(env, o, "Uninstall "+name); err != nil {
		return err
	}
	if err := app.Uninstall(ctx, cluster, &plan); err != nil {
		return fmt.Errorf("uninstall: %w", err)
	}
	return say(env.Out, "Uninstalled "+name+".\n")
}

// list runs `openfief plugin list`.
func list(ctx context.Context, env *Env, o *options) error {
	cluster, err := env.Cluster(o.kubeconfig, o.context)
	if err != nil {
		return fmt.Errorf("cluster: %w", err)
	}
	all, err := app.Instances(ctx, cluster)
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}
	return WriteInstances(env.Out, all)
}

// confirm asks the owner, unless --yes was given; anything but y or yes is
// a no, and so is no input at all.
func confirm(env *Env, o *options, question string) error {
	if o.yes {
		return nil
	}
	if err := say(env.Out, question+"? [y/N] "); err != nil {
		return err
	}
	line, err := bufio.NewReader(env.In).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("read answer: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return nil
	}
	return ErrAborted
}

// say writes s to w.
func say(w io.Writer, s string) error {
	if _, err := io.WriteString(w, s); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}
