// Command operator runs OpenFief's plugin operator (D-13): it reconciles the
// Plugin and PluginGrant resources of the openfief namespace, and only that
// namespace (D-15), into running, isolated plugins.
package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/catalog"
	"github.com/arkadiuss/openfief/internal/plugins/cosign"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

const (
	defaultPluginPort = 8080
	// defaultTrustedSigners is where deploy/operator mounts the optional
	// openfief-trusted-signers ConfigMap.
	defaultTrustedSigners = "/etc/openfief/trust/signers.json"
)

// Flag errors.
var (
	// errPluginPort reports a plugin port outside 1 to 65535.
	errPluginPort = errors.New("plugin port must be 1 to 65535")
	// errKeysRefresh reports a signing keys refresh interval that is not
	// positive.
	errKeysRefresh = errors.New("signing keys refresh must be positive")
)

type options struct {
	settings       operator.Settings
	keysRefresh    time.Duration
	pluginPort     uint
	trustedSigners string
	leaderElect    bool
	probeAddr      string
	metricsAddr    string
	zap            zap.Options
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "operator:", err)
		os.Exit(1)
	}
}

func parseFlags(args []string) (*options, error) {
	var o options
	fs := flag.NewFlagSet("operator", flag.ContinueOnError)
	fs.StringVar(&o.settings.CoreService, "core-service", "openfief-core", "the core's Service in the openfief namespace")
	fs.StringVar(&o.settings.CorePort, "core-port", "connect", "the name of the core Service's Connect RPC port")
	fs.StringVar(&o.settings.EgressProxyService, "egress-proxy-service", "egress-proxy", "the egress proxy's Service in the openfief namespace")
	fs.StringVar(&o.settings.EgressProxyPort, "egress-proxy-port", "proxy", "the name of the egress proxy Service's port")
	fs.UintVar(&o.pluginPort, "plugin-port", defaultPluginPort, "the port every plugin serves Connect RPC on")
	fs.StringVar(&o.settings.CoreServiceAccount, "core-service-account", "openfief-core",
		"the core's ServiceAccount in the openfief namespace, whose tokens plugins accept (D-59)")
	fs.DurationVar(&o.keysRefresh, "signing-keys-refresh", operator.DefaultSigningKeysRefresh,
		"how often to publish the cluster's ServiceAccount signing keys for plugins again")
	fs.StringVar(&o.trustedSigners, "trusted-signers", defaultTrustedSigners,
		"JSON file of the signers the owner trusts besides OpenFief's CI and the catalog (D-45); a missing file adds none")
	fs.BoolVar(&o.leaderElect, "leader-elect", true, "elect one active operator with a Lease in the openfief namespace")
	fs.StringVar(&o.probeAddr, "health-probe-bind-address", ":8081", "address of the liveness and readiness probes")
	fs.StringVar(&o.metricsAddr, "metrics-bind-address", "0", `address of the metrics endpoint; "0" disables it`)
	o.zap.BindFlags(fs)
	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("flags: %w", err)
	}
	if o.pluginPort == 0 || o.pluginPort > math.MaxUint16 {
		return nil, fmt.Errorf("%w: %d", errPluginPort, o.pluginPort)
	}
	if o.keysRefresh <= 0 {
		return nil, fmt.Errorf("%w: %s", errKeysRefresh, o.keysRefresh)
	}
	o.settings.PluginPort = uint16(o.pluginPort)
	return &o, nil
}

func run(args []string) error {
	o, err := parseFlags(args)
	if err != nil {
		return err
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&o.zap)))
	scheme := runtime.NewScheme()
	if schemeErr := errors.Join(clientgoscheme.AddToScheme(scheme), v1alpha1.AddToScheme(scheme)); schemeErr != nil {
		return fmt.Errorf("scheme: %w", schemeErr)
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// D-15: the operator sees, and holds a Role in, the openfief namespace only.
		Cache:                   cache.Options{DefaultNamespaces: map[string]cache.Config{plugins.Namespace: {}}},
		LeaderElection:          o.leaderElect,
		LeaderElectionID:        "openfief-operator",
		LeaderElectionNamespace: plugins.Namespace,
		HealthProbeBindAddress:  o.probeAddr,
		Metrics:                 metricsserver.Options{BindAddress: o.metricsAddr},
	})
	if err != nil {
		return fmt.Errorf("manager: %w", err)
	}
	verifier, err := imageVerifier(o.trustedSigners)
	if err != nil {
		return err
	}
	pr := &operator.PluginReconciler{
		Client: mgr.GetClient(), Scheme: scheme, Settings: o.settings, Verifier: verifier,
	}
	dc, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		return fmt.Errorf("discovery client: %w", err)
	}
	keys := &operator.SigningKeysPublisher{
		Client: mgr.GetClient(), Source: operator.IssuerDiscovery{Client: dc.RESTClient()}, Refresh: o.keysRefresh,
	}
	if setupErr := errors.Join(
		pr.SetupWithManager(mgr),
		(&operator.GrantReconciler{Client: mgr.GetClient()}).SetupWithManager(mgr),
		keys.SetupWithManager(mgr),
	); setupErr != nil {
		return fmt.Errorf("controllers: %w", setupErr)
	}
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("health check: %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("ready check: %w", err)
	}
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return nil
}

// imageVerifier builds the operator's signature check (D-45, D-56, D-57):
// OpenFief's CI, the embedded catalog and the owner's signers, verified
// against the bundled Sigstore trust root. There is no flag that skips it.
func imageVerifier(trustedSignersPath string) (*operator.PolicyVerifier, error) {
	policy, err := trustPolicy(trustedSignersPath)
	if err != nil {
		return nil, err
	}
	tr, err := cosign.TrustedRoot()
	if err != nil {
		return nil, fmt.Errorf("trust root: %w", err)
	}
	source, err := cosign.New(&policy, tr)
	if err != nil {
		return nil, fmt.Errorf("signature verifier: %w", err)
	}
	return operator.NewPolicyVerifier(source, &policy), nil
}

// trustPolicy reads the embedded catalog and the owner's signers. A missing
// signers file means the owner added none; an unreadable or invalid one stops
// the operator, so a typo never silently drops or widens trust.
func trustPolicy(trustedSignersPath string) (plugins.TrustPolicy, error) {
	c, err := catalog.Load()
	if err != nil {
		return plugins.TrustPolicy{}, fmt.Errorf("catalog: %w", err)
	}
	var owner []plugins.Signer
	data, err := os.ReadFile(filepath.Clean(trustedSignersPath))
	switch {
	case errors.Is(err, os.ErrNotExist):
	case err != nil:
		return plugins.TrustPolicy{}, fmt.Errorf("trusted signers: %w", err)
	default:
		if owner, err = plugins.ParseOwnerSigners(data); err != nil {
			return plugins.TrustPolicy{}, fmt.Errorf("trusted signers %s: %w", trustedSignersPath, err)
		}
	}
	policy, err := plugins.NewTrustPolicy(c, owner)
	if err != nil {
		return plugins.TrustPolicy{}, fmt.Errorf("trust policy: %w", err)
	}
	return policy, nil
}
