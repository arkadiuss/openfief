// Command openfief is the owner's command line (D-46): it installs,
// uninstalls and lists plugin instances with the owner's own kubeconfig,
// and gives the owner the link for the first passkey (D-80).
//
// Usage:
//
//	openfief enroll [--port 8080] [--ttl 15m]
//	openfief plugin install <catalog-name|image@sha256:...> [--name NAME] [--config FILE]
//	                        [--trusted-signers FILE] [--yes] [--print]
//	openfief plugin uninstall <instance> [--yes]
//	openfief plugin list
//
// install reads the catalog entry or the image, verifies the image's
// signature with the operator's trust policy and signature check, shows the
// permissions its signed manifest requests and, once the owner confirms,
// creates the Plugin and a PluginGrant approving exactly those; --print
// writes them as YAML for GitOps instead. uninstall deletes an instance's
// PluginGrant, Plugin and the Secrets no other instance names (D-58).
// enroll forwards localhost:8080 to the core, the only entry point for
// browsers (D-81), has the core issue D-60's one-time link for it by running
// core enroll-link in its pod, prints the link and keeps forwarding until
// Ctrl-C. Every command takes --kubeconfig and --context.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	identitycli "github.com/arkadiuss/openfief/internal/identity/cli"
	identitykube "github.com/arkadiuss/openfief/internal/identity/kube"
	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/catalog"
	"github.com/arkadiuss/openfief/internal/plugins/cli"
	"github.com/arkadiuss/openfief/internal/plugins/cosign"
	"github.com/arkadiuss/openfief/internal/plugins/kube"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// exitUsage is the exit status for bad arguments.
const exitUsage = 2

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	err := run(ctx, os.Args[1:])
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "openfief:", err)
		if errors.Is(err, cli.ErrUsage) || errors.Is(err, identitycli.ErrUsage) {
			os.Exit(exitUsage)
		}
		os.Exit(1)
	}
}

// run builds the real adapters and runs the command line.
func run(ctx context.Context, args []string) error {
	if len(args) > 0 && args[0] == "enroll" {
		connectTo := func(kubeconfig, kubeContext string) (identitycli.Core, error) {
			core, err := connectCore(kubeconfig, kubeContext)
			if err != nil {
				return nil, err
			}
			return core, nil
		}
		if err := identitycli.Run(ctx, args[1:], &identitycli.Env{Out: os.Stdout, Connect: connectTo}); err != nil {
			return fmt.Errorf("enroll: %w", err)
		}
		return nil
	}
	env, err := newEnv()
	if err != nil {
		return err
	}
	if err := cli.Run(ctx, args, env); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// newEnv wires the command line to the cluster, the registry and the files
// the owner names.
func newEnv() (*cli.Env, error) {
	c, err := catalog.Load()
	if err != nil {
		return nil, fmt.Errorf("catalog: %w", err)
	}
	return &cli.Env{
		In: os.Stdin, Out: os.Stdout, Err: os.Stderr,
		Catalog:  c,
		ReadFile: func(path string) ([]byte, error) { return os.ReadFile(filepath.Clean(path)) },
		Cluster: func(kubeconfig, kubeContext string) (cli.Cluster, error) {
			c, err := connect(kubeconfig, kubeContext)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		Images: func(policy *plugins.TrustPolicy) (app.ImageInspector, error) {
			v, err := images(policy)
			if err != nil {
				return nil, err
			}
			return v, nil
		},
		ValidateConfig: operator.ValidateConfig,
		Render:         kube.ResourcesYAML,
		IngressRule:    kube.IngressRuleYAML,
	}, nil
}

// images is the operator's own signature check (D-56): sigstore-go against
// the bundled Sigstore trust root, reading the registry anonymously.
func images(policy *plugins.TrustPolicy) (*cosign.Verifier, error) {
	tr, err := cosign.TrustedRoot()
	if err != nil {
		return nil, fmt.Errorf("trust root: %w", err)
	}
	v, err := cosign.New(policy, tr)
	if err != nil {
		return nil, fmt.Errorf("signature verifier: %w", err)
	}
	return v, nil
}

// cluster is the owner's cluster: the instance store and the owner's
// trusted signers.
type cluster struct {
	*kube.Instances
}

// OwnerSigners reads the openfief-trusted-signers ConfigMap.
func (c cluster) OwnerSigners(ctx context.Context) ([]plugins.Signer, error) {
	signers, err := kube.OwnerSigners(ctx, c.Client)
	if err != nil {
		return nil, fmt.Errorf("cluster: %w", err)
	}
	return signers, nil
}

// restConfig reads the owner's kubeconfig: the file given, or $KUBECONFIG,
// or ~/.kube/config, and the context given or the current one.
func restConfig(kubeconfig, kubeContext string) (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = kubeconfig
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules, &clientcmd.ConfigOverrides{CurrentContext: kubeContext},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("kubeconfig: %w", err)
	}
	return cfg, nil
}

// connectCore reaches the core's pod in the openfief namespace with the
// owner's kubeconfig, for exec and port-forward (D-80).
func connectCore(kubeconfig, kubeContext string) (*identitykube.CoreAccess, error) {
	cfg, err := restConfig(kubeconfig, kubeContext)
	if err != nil {
		return nil, err
	}
	clients, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}
	return &identitykube.CoreAccess{Config: cfg, Client: clients, Namespace: plugins.Namespace}, nil
}

// connect builds a client from the owner's kubeconfig (restConfig).
func connect(kubeconfig, kubeContext string) (cluster, error) {
	cfg, err := restConfig(kubeconfig, kubeContext)
	if err != nil {
		return cluster{}, err
	}
	scheme, err := newScheme()
	if err != nil {
		return cluster{}, err
	}
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return cluster{}, fmt.Errorf("client: %w", err)
	}
	return cluster{&kube.Instances{Client: c}}, nil
}

// newScheme holds the core types and the operator's.
func newScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := errors.Join(clientgoscheme.AddToScheme(scheme), v1alpha1.AddToScheme(scheme)); err != nil {
		return nil, fmt.Errorf("scheme: %w", err)
	}
	return scheme, nil
}
