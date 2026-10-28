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

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

const defaultPluginPort = 8080

// errPluginPort reports a plugin port outside 1 to 65535.
var errPluginPort = errors.New("plugin port must be 1 to 65535")

type options struct {
	settings    operator.Settings
	pluginPort  uint
	leaderElect bool
	probeAddr   string
	metricsAddr string
	zap         zap.Options
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
	pr := &operator.PluginReconciler{
		Client: mgr.GetClient(), Scheme: scheme, Settings: o.settings,
		// D-17: no plugin is deployed until cosign verification replaces this seam.
		Verifier: operator.UnimplementedVerifier{},
	}
	if setupErr := errors.Join(
		pr.SetupWithManager(mgr),
		(&operator.GrantReconciler{Client: mgr.GetClient()}).SetupWithManager(mgr),
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
