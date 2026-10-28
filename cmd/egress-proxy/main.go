// Command egress-proxy is OpenFief's egress proxy (D-32): Smokescreen with
// plugin identity from projected ServiceAccount tokens (D-27) and per-plugin
// egress allow-lists from the operator's ConfigMap. It runs in the openfief
// namespace with in-cluster credentials and stops on SIGTERM.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	ss "github.com/stripe/smokescreen/pkg/smokescreen"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/kube"
	"github.com/arkadiuss/openfief/internal/plugins/smokescreen"
)

func main() {
	addr := flag.String("listen", ":4750", "address plugins connect to")
	flag.Parse()
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	if err := run(*addr, log); err != nil {
		log.WithError(err).Error("egress proxy stopped")
		os.Exit(1)
	}
}

func run(addr string, log *logrus.Logger) error {
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
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	cfg, err := smokescreen.NewConfig(app.NewEgress(kube.NewTokenReviewer(client), lists), smokescreen.Options{
		Listener:    listener,
		Healthcheck: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }),
		Log:         log,
	})
	if err != nil {
		return errors.Join(fmt.Errorf("egress proxy config: %w", err), listener.Close())
	}
	// Smokescreen handles SIGTERM itself, draining open tunnels; quit stops
	// it at once when the allow-list watch's context ends first.
	quit := make(chan any)
	go func() {
		<-ctx.Done()
		close(quit)
	}()
	ss.StartWithConfig(cfg, quit)
	return nil
}
