// Command egress-proxy is OpenFief's egress proxy (D-32, D-63): Smokescreen,
// embedded in OpenFief's own HTTP server, with plugin identity from projected
// ServiceAccount tokens (audience openfief-egress) checked by TokenReview and
// cached, and per-plugin egress allow-lists from the operator's ConfigMap. It
// runs in the openfief namespace with in-cluster credentials and stops on
// SIGTERM, letting open tunnels drain.
//
// `egress-proxy audit` serves nothing: it reads the proxy's logs and the
// PluginGrants and checks that every request the proxy allowed went to a
// host its instance's grant approves, and that no log line holds more than
// who asked, where to and the decision (audit.go). The demo runs it after
// each deploy, and the release smoke test after its install.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/egress"
	"github.com/arkadiuss/openfief/internal/plugins/kube"
	"github.com/arkadiuss/openfief/internal/plugins/smokescreen"
)

// nodeIPEnv names the environment variable that holds the addresses of the
// node the proxy runs on, comma-separated; the Deployment sets it from the
// pod's status.hostIP.
const nodeIPEnv = "OPENFIEF_NODE_IP"

// exemptHostEnv names the variable that holds the test-only exemption when
// -refused-range-exempt-host is not given (D-215).
const exemptHostEnv = "OPENFIEF_REFUSED_RANGE_EXEMPT_HOST"

func main() {
	if len(os.Args) > 1 && os.Args[1] == auditCommand {
		if err := runAudit(context.Background(), os.Args[2:], os.Stdout); err != nil {
			log := logrus.New()
			log.SetFormatter(&logrus.JSONFormatter{})
			log.WithError(err).Error("egress audit failed")
			os.Exit(1)
		}
		return
	}
	addr := flag.String("listen", ":4750", "address plugins connect to")
	denyCIDRs := flag.String("deny-cidr", "", "comma-separated ranges to refuse beyond the private and internal ones, "+
		"such as a cluster's public or global IPv6 pod, Service and node ranges")
	exemptHost := flag.String("refused-range-exempt-host", "", "TEST ONLY (D-215): host=address[,address...], exempting that one host, "+
		"resolving to exactly those addresses, from the refused address ranges; "+exemptHostEnv+" is the same setting. "+
		"The proxy warns at start when it is set, and a release never sets it")
	flag.Parse()
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	deny, err := egress.ParseDenyList(*denyCIDRs, os.Getenv(nodeIPEnv))
	if err != nil {
		log.WithError(err).Error("egress proxy stopped")
		os.Exit(1)
	}
	exemptText := *exemptHost
	if exemptText == "" {
		exemptText = os.Getenv(exemptHostEnv)
	}
	exempt, err := egress.ParseExemption(exemptText)
	if err != nil {
		log.WithError(err).Error("egress proxy stopped")
		os.Exit(1)
	}
	if err := run(*addr, deny, exempt, log); err != nil {
		log.WithError(err).Error("egress proxy stopped")
		os.Exit(1)
	}
}

func run(addr string, deny []netip.Prefix, exempt egress.Exemption, log *logrus.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("in-cluster config: %w", err)
	}
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("kubernetes client: %w", err)
	}
	lists, err := kube.WatchAllowLists(ctx, client, func(err error) {
		log.WithError(err).Error("egress allow-lists refused; denying all egress until fixed")
	})
	if err != nil {
		return fmt.Errorf("egress allow-lists: %w", err)
	}
	tokens, err := app.NewCachingTokenReviewer(kube.NewTokenReviewer(client, plugins.EgressAudience), app.DefaultTokenCacheTTL, time.Now)
	if err != nil {
		return fmt.Errorf("token cache: %w", err)
	}
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	proxy, err := smokescreen.New(app.NewEgress(tokens, lists), &smokescreen.Options{
		Listener:    listener,
		Deny:        deny,
		Exempt:      exempt,
		Healthcheck: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }),
		Log:         log,
	})
	if err != nil {
		return errors.Join(fmt.Errorf("egress proxy: %w", err), listener.Close())
	}
	if err := proxy.Serve(ctx); err != nil {
		return fmt.Errorf("egress proxy: %w", err)
	}
	return nil
}
