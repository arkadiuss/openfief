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
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
)

// Usage is the command line's synopsis.
const Usage = `usage:
  openfief plugin install <catalog-name|image@sha256:...> [--name NAME] [--config FILE]
                          [--cluster-service NAMESPACE/SERVICE:PORT] [--extra-ca]
                          [--trusted-signers FILE] [--yes] [--print]
  openfief plugin upgrade <instance> <catalog-version|image@sha256:...> [--config FILE]
                          [--extra-ca | --no-extra-ca] [--trusted-signers FILE] [--yes]
  openfief plugin ca-bundle set <instance> <pem-file> [--print]
  openfief plugin ca-bundle unset <instance> [--yes]
  openfief plugin uninstall <instance> [--yes]
  openfief plugin list
  openfief enroll [--port 8080] [--ttl 15m]
  openfief version
common flags: --kubeconfig PATH, --context NAME`

// Errors of the command line.
var (
	// ErrUsage reports arguments the command line does not accept.
	ErrUsage = errors.New("bad arguments")
	// ErrAborted reports that the owner did not confirm.
	ErrAborted = errors.New("not confirmed; nothing changed")
)

// Cluster is the owner's cluster, reached with their kubeconfig: the
// installed instances, their extra CA bundles, and the owner's trusted
// signers the operator reads.
type Cluster interface {
	app.InstanceStore
	app.CABundleStore
	OwnerSigners(ctx context.Context) ([]plugins.Signer, error)
	// ClusterService reads an in-cluster service the owner picked (D-74).
	ClusterService(ctx context.Context, cs plugins.ClusterService) (app.ClusterServiceInfo, error)
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
	// IngressRule renders the NetworkPolicy the owner adds in an in-cluster
	// service's namespace that has NetworkPolicies of its own (D-74).
	IngressRule func(instance string, cs plugins.ClusterService, info *app.ClusterServiceInfo) ([]byte, error)
	// RenderCABundle renders, for `ca-bundle set --print`, the CA bundle
	// Secret and the Plugin stamp of its checksum as YAML.
	RenderCABundle func(instance string, pem []byte, checksum string) ([]byte, error)
}

// options are the parsed flags of one command, and the owner's input.
type options struct {
	kubeconfig, context string
	name, config        string
	clusterService      string
	extraCA, noExtraCA  bool
	trustedSigners      string
	yes, print          bool
	// in reads the owner's answers, line by line.
	in *bufio.Reader
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
	o.in = bufio.NewReader(env.In)
	if !argsFit(cmd, pos) {
		return fmt.Errorf("%w\n%s", ErrUsage, Usage)
	}
	switch cmd {
	case cmdCABundle:
		return caBundle(ctx, env, &o, pos)
	case cmdInstall:
		return install(ctx, env, &o, pos[0])
	case cmdUpgrade:
		return upgrade(ctx, env, &o, pos[0], pos[1])
	case cmdUninstall:
		return uninstall(ctx, env, &o, pos[0])
	}
	return list(ctx, env, &o)
}

// The commands of `openfief plugin`.
const (
	cmdInstall   = "install"
	cmdUpgrade   = "upgrade"
	cmdUninstall = "uninstall"
	cmdList      = "list"
	cmdCABundle  = "ca-bundle"
	// upgradeArgs are an instance and a version.
	upgradeArgs = 2
)

// argsFit reports whether cmd is a command and pos its arguments: how many
// each command takes, and for ca-bundle a subcommand with its own.
func argsFit(cmd string, pos []string) bool {
	if cmd == cmdCABundle {
		return caBundleArgs(pos)
	}
	n, ok := map[string]int{cmdInstall: 1, cmdUpgrade: upgradeArgs, cmdUninstall: 1, cmdList: 0}[cmd]
	return ok && len(pos) == n
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
	case cmdInstall:
		fs.StringVar(&o.name, "name", "", "instance name; default: the catalog name or the image's repository name")
		fs.StringVar(&o.config, "config", "", "JSON file with the instance's config")
		fs.StringVar(&o.clusterService, "cluster-service", "",
			"the in-cluster service a plugin that asks for one reaches, as namespace/service:port; asked when not given")
		fs.BoolVar(&o.extraCA, "extra-ca", false,
			"expect an extra CA bundle in the Secret <instance>--ca-bundle, for an endpoint a private CA signs; the manifest must ask for role ca-bundle")
		fs.StringVar(&o.trustedSigners, "trusted-signers", "", "the owner's signers file; default: the cluster's openfief-trusted-signers ConfigMap")
		fs.BoolVar(&o.yes, "yes", false, "install without asking")
		fs.BoolVar(&o.print, "print", false, "print the Plugin and PluginGrant as YAML for GitOps instead of applying them")
	case cmdUpgrade:
		fs.StringVar(&o.config, "config", "", "JSON file with the instance's new config; default: keep its config")
		fs.BoolVar(&o.extraCA, "extra-ca", false,
			"expect an extra CA bundle from now on; the new manifest must ask for role ca-bundle; default: keep what the instance expects")
		fs.BoolVar(&o.noExtraCA, "no-extra-ca", false, "expect no extra CA bundle from now on")
		fs.StringVar(&o.trustedSigners, "trusted-signers", "", "the owner's signers file; default: the cluster's openfief-trusted-signers ConfigMap")
		fs.BoolVar(&o.yes, "yes", false, "upgrade without asking")
	case cmdUninstall:
		fs.BoolVar(&o.yes, "yes", false, "uninstall without asking")
	case cmdCABundle:
		fs.BoolVar(&o.yes, "yes", false, "unset without asking")
		fs.BoolVar(&o.print, "print", false, "print the Secret and the Plugin's checksum stamp as YAML for GitOps instead of applying them")
	}
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			return options{}, nil, fmt.Errorf("%w: %w\n%s", ErrUsage, err, Usage)
		}
		if fs.NArg() == 0 {
			if o.extraCA && o.noExtraCA {
				return options{}, nil, fmt.Errorf("%w: --extra-ca and --no-extra-ca together\n%s", ErrUsage, Usage)
			}
			return o, pos, nil
		}
		pos = append(pos, fs.Arg(0))
		args = fs.Args()[1:]
	}
}

// install runs `openfief plugin install`.
func install(ctx context.Context, env *Env, o *options, arg string) error {
	var cluster Cluster
	if !o.print {
		c, err := env.Cluster(o.kubeconfig, o.context)
		if err != nil {
			return fmt.Errorf("cluster: %w", err)
		}
		cluster = c
	}
	p, err := prepare(ctx, env, o, cluster, arg)
	if err != nil {
		return err
	}
	if o.print {
		return printResources(env, &p)
	}
	if err = checkInstall(ctx, cluster, &p); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	rule, err := checkClusterService(ctx, env, cluster, &p.Plugin)
	if err != nil {
		return err
	}
	if err := say(env.Out, RenderProposal(&p)+rule); err != nil {
		return err
	}
	if err := confirm(env, o, "Install "+p.Plugin.Name); err != nil {
		return err
	}
	if err := app.Install(ctx, cluster, &p); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	return say(env.Out, "Installed "+p.Plugin.Name+". The operator verifies the image again and deploys it once its Secrets exist.\n"+grantText(&p.Plugin))
}

// grantText is what the summary of an installed view plugin adds: with SpiceDB
// authorization a view plugin reads nothing until the owner grants it accounts
// (D-209). The command can't tell which authorization the core runs with, so
// the line says when it matters; it is empty for any other plugin.
func grantText(p *plugins.Plugin) string {
	if !slices.Contains(p.Manifest.Capabilities, plugins.CapabilityView) {
		return ""
	}
	return "If OpenFief runs with SpiceDB authorization, grant it accounts in Settings → Access.\n"
}

// prepare reads the config and the in-cluster service the owner named,
// verifies the image and plans the install, asking for the in-cluster
// service when the manifest wants one and no flag named it.
func prepare(ctx context.Context, env *Env, o *options, cluster Cluster, arg string) (app.Proposal, error) {
	config, err := readConfig(env, o.config)
	if err != nil {
		return app.Proposal{}, err
	}
	var picked *plugins.ClusterService
	if o.clusterService != "" {
		cs, csErr := plugins.ParseClusterService(o.clusterService)
		if csErr != nil {
			return app.Proposal{}, fmt.Errorf("%w: --cluster-service: %w", ErrUsage, csErr)
		}
		picked = &cs
	}
	in, err := installer(ctx, env, o, cluster)
	if err != nil {
		return app.Proposal{}, err
	}
	p, err := in.Prepare(ctx, app.InstallRequest{Arg: arg, Name: o.name, Config: config, ClusterService: picked, ExtraCA: o.extraCA})
	if err != nil {
		return app.Proposal{}, fmt.Errorf("install: %w", err)
	}
	if err := askClusterService(env, o, &p); err != nil {
		return app.Proposal{}, err
	}
	return p, nil
}

// askClusterService asks the owner for the in-cluster service a manifest
// asks for, unless --cluster-service named one, --yes or --print was given
// (then there is none), or the manifest asks for none (D-74). An empty
// namespace picks none.
func askClusterService(env *Env, o *options, p *app.Proposal) error {
	r := p.Plugin.Manifest.ClusterService
	if r == nil || p.Plugin.ClusterService != nil || o.yes || o.print {
		return nil
	}
	what := shown(r.Description)
	if r.Description == "" {
		what = "a service in this cluster"
	}
	if err := say(env.Out, "The plugin asks to reach "+what+" directly, outside the egress proxy.\n"); err != nil {
		return err
	}
	cs, err := askService(env, o)
	if err != nil || cs == nil {
		return err
	}
	if err := p.PickClusterService(*cs); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	return nil
}

// askService asks for an in-cluster service's namespace, service and port;
// nil when the owner gives no namespace.
func askService(env *Env, o *options) (*plugins.ClusterService, error) {
	ns, err := ask(env, o, "  Namespace (empty for none): ")
	if err != nil || ns == "" {
		return nil, err
	}
	svc, err := ask(env, o, "  Service: ")
	if err != nil {
		return nil, err
	}
	port, err := ask(env, o, "  Port: ")
	if err != nil {
		return nil, err
	}
	cs, err := plugins.ParseClusterService(ns + "/" + svc + ":" + port)
	if err != nil {
		return nil, fmt.Errorf("in-cluster service: %w", err)
	}
	return &cs, nil
}

// checkClusterService reads the in-cluster service the Plugin names, which
// must exist, and returns the ingress rule the owner must add when its
// namespace has NetworkPolicies of its own (D-74); nothing otherwise.
func checkClusterService(ctx context.Context, env *Env, cluster Cluster, p *plugins.Plugin) (string, error) {
	if p.ClusterService == nil {
		return "", nil
	}
	cs := *p.ClusterService
	info, err := cluster.ClusterService(ctx, cs)
	if err != nil {
		return "", fmt.Errorf("in-cluster service: %w", err)
	}
	if !info.NamespaceHasNetworkPolicies {
		return "", nil
	}
	rule, err := env.IngressRule(p.Name, cs, &info)
	if err != nil {
		return "", fmt.Errorf("ingress rule: %w", err)
	}
	return RenderIngressRule(cs.Namespace, rule), nil
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

// printResources writes the summary to Err and the resources to Out. With
// an in-cluster service it can't read the cluster, so it says what an
// ingress rule in that namespace must admit (D-74).
func printResources(env *Env, p *app.Proposal) error {
	summary := RenderProposal(p)
	if cs := p.Plugin.ClusterService; cs != nil {
		summary += fmt.Sprintf("\nIf namespace %s has NetworkPolicies of its own, add one there that admits pods labelled\n"+
			"openfief.io/plugin=%s in namespace %s to the service's pods on its target port.\n",
			cs.Namespace, p.Plugin.Name, plugins.Namespace)
	}
	if err := say(env.Err, summary); err != nil {
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
	answer, err := ask(env, o, question+"? [y/N] ")
	if err != nil {
		return err
	}
	switch strings.ToLower(answer) {
	case "y", "yes":
		return nil
	}
	return ErrAborted
}

// ask writes the prompt and reads one line of the owner's answer, trimmed;
// no input at all is an empty answer.
func ask(env *Env, o *options, prompt string) (string, error) {
	if err := say(env.Out, prompt); err != nil {
		return "", err
	}
	line, err := o.in.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read answer: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// say writes s to w.
func say(w io.Writer, s string) error {
	if _, err := io.WriteString(w, terminalText(s)); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// checkInstall refuses a second browser view (D-114) and a Secret another
// instance names (D-89, D-182), and finds the institutions another
// instance already books for, which the summary warns about (D-185).
func checkInstall(ctx context.Context, cluster Cluster, p *app.Proposal) error {
	if err := app.CheckBrowserView(ctx, cluster, &p.Plugin); err != nil {
		return fmt.Errorf("browser view: %w", err)
	}
	if err := app.CheckSecrets(ctx, cluster, &p.Plugin, &p.Grant); err != nil {
		return fmt.Errorf("secrets: %w", err)
	}
	shared, err := app.SharedInstitutions(ctx, cluster, &p.Plugin)
	if err != nil {
		return fmt.Errorf("institutions: %w", err)
	}
	p.Shared = shared
	return nil
}
