package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arkadiuss/openfief/internal/plugins/kube"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
	"github.com/arkadiuss/openfief/internal/plugins/smokescreen"
)

// auditCommand is the subcommand that audits the proxy's log instead of
// serving.
const auditCommand = "audit"

// errAuditFailed reports an audit that found a line breaking a rule.
var errAuditFailed = errors.New("egress audit failed")

// runAudit is `egress-proxy audit [-kubeconfig PATH] [-context NAME]`: it
// reads the logs of the egress proxy's pods and every PluginGrant in the
// openfief namespace, with in-cluster credentials when it runs in a pod and
// the kubeconfig otherwise, and checks the decision lines
// (smokescreen.AuditLog): every allowed request went to a host its
// instance's grant approves, and no line holds more than who asked, where
// to and the decision (D-200 criterion 16, GR-12). It only reads, and it
// prints counts, instances and hosts; never a line itself.
func runAudit(ctx context.Context, args []string, out io.Writer) error {
	fs := flag.NewFlagSet(auditCommand, flag.ContinueOnError)
	fs.SetOutput(out)
	kubeconfig := fs.String("kubeconfig", "", "kubeconfig file; in a pod, in-cluster credentials are used without one")
	kubeContext := fs.String("context", "", "kubeconfig context; the current one by default")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%s: %w", auditCommand, err)
	}
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = *kubeconfig
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules, &clientcmd.ConfigOverrides{CurrentContext: *kubeContext},
	).ClientConfig()
	if err != nil {
		return fmt.Errorf("kubeconfig: %w", err)
	}
	clients, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("kubernetes client: %w", err)
	}
	scheme := runtime.NewScheme()
	if schemeErr := v1alpha1.AddToScheme(scheme); schemeErr != nil {
		return fmt.Errorf("scheme: %w", schemeErr)
	}
	grants, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("plugin grant client: %w", err)
	}
	log, err := kube.EgressProxyLog(ctx, clients)
	if err != nil {
		return fmt.Errorf("egress audit: %w", err)
	}
	granted, err := kube.GrantedEgress(ctx, grants)
	if err != nil {
		return fmt.Errorf("egress audit: %w", err)
	}
	return report(out, smokescreen.AuditLog(log, granted))
}

// report prints what an audit found, and fails when it found a problem.
func report(out io.Writer, a smokescreen.Audit) error {
	var err error
	write := func(format string, args ...any) {
		if err == nil {
			_, err = fmt.Fprintf(out, format, args...)
		}
	}
	write("egress audit: %d decision lines, %d refused, %d allowed\n", a.Decisions, a.Refused, a.Decisions-a.Refused)
	for _, d := range a.Allowed {
		write("egress audit: allowed %s to %s, %d times\n", d.Instance, d.Host, d.Count)
	}
	for _, p := range a.Problems {
		write("egress audit: %s\n", p)
	}
	if err != nil {
		return fmt.Errorf("egress audit: %w", err)
	}
	if len(a.Problems) > 0 {
		return fmt.Errorf("%w: %d problems", errAuditFailed, len(a.Problems))
	}
	write("egress audit: every allowed request went to a host its instance's grant approves\n")
	if err != nil {
		return fmt.Errorf("egress audit: %w", err)
	}
	return nil
}
