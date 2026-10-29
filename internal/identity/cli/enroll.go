// Package cli is the identity context's part of the owner's openfief
// command line: `openfief enroll`, the first passkey (D-80). It forwards a
// local port to the core, the only entry point for browsers (D-81), has
// the core issue D-60's one-time link for that origin, prints it, and keeps
// the port-forward open until the owner stops it. cmd/openfief wires the
// Kubernetes adapter, internal/identity/kube's CoreAccess.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
)

// Usage is the command's synopsis.
const Usage = `usage:
  openfief enroll [--port 8080] [--ttl 15m] [--kubeconfig PATH] [--context NAME]`

// Defaults of the flags.
const (
	// DefaultPort is the local port, the one the core's default allowed
	// origin, http://localhost:8080, names.
	DefaultPort = 8080
	maxPort     = 65535
)

// Errors of the command.
var (
	// ErrUsage reports arguments the command does not accept.
	ErrUsage = errors.New("bad arguments")
	// errForwardEnded reports a port-forward that stopped on its own.
	errForwardEnded = errors.New("the port-forward to the core stopped")
)

// Core is the core's pod, reached with the owner's kubeconfig.
type Core interface {
	// Forward forwards localhost:localPort to the core until ctx ends,
	// closing ready once it listens.
	Forward(ctx context.Context, localPort int, ready chan struct{}) error
	// EnrollLink runs the core's enroll-link command for host and returns
	// the one-time link it prints.
	EnrollLink(ctx context.Context, host string, ttl time.Duration) (string, error)
}

// Env is what the command needs from outside it.
type Env struct {
	Out io.Writer
	// Connect reaches the core with a kubeconfig path and context, either
	// empty for the defaults.
	Connect func(kubeconfig, context string) (Core, error)
}

// options are the parsed flags.
type options struct {
	kubeconfig, context string
	port                int
	ttl                 time.Duration
}

// parse parses enroll's flags.
func parse(args []string) (options, error) {
	fs := flag.NewFlagSet("openfief enroll", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o options
	fs.StringVar(&o.kubeconfig, "kubeconfig", "", "kubeconfig file")
	fs.StringVar(&o.context, "context", "", "kubeconfig context")
	fs.IntVar(&o.port, "port", DefaultPort, "local port to forward to the core")
	fs.DurationVar(&o.ttl, "ttl", identity.DefaultEnrollmentTTL, "how long the link works")
	if err := fs.Parse(args); err != nil {
		return options{}, fmt.Errorf("%w: %w\n%s", ErrUsage, err, Usage)
	}
	if fs.NArg() > 0 {
		return options{}, fmt.Errorf("%w: unexpected %q\n%s", ErrUsage, fs.Arg(0), Usage)
	}
	if o.port < 1 || o.port > maxPort {
		return options{}, fmt.Errorf("%w: --port %d is not a port\n%s", ErrUsage, o.port, Usage)
	}
	return o, nil
}

// Run runs `openfief enroll` with args, the arguments after "enroll": it
// starts the port-forward to the core, prints the one-time link for
// http://localhost:<port> once the forward listens, and keeps forwarding
// until ctx ends, which is a normal stop.
func Run(ctx context.Context, args []string, env *Env) error {
	o, err := parse(args)
	if err != nil {
		return err
	}
	core, err := env.Connect(o.kubeconfig, o.context)
	if err != nil {
		return fmt.Errorf("connecting to the cluster: %w", err)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ready := make(chan struct{})
	forwarded := make(chan error, 1)
	go func() { forwarded <- core.Forward(ctx, o.port, ready) }()
	select {
	case <-ready:
	case ferr := <-forwarded:
		return forwardEnded(ferr)
	case <-ctx.Done():
		return nil
	}
	host := "localhost:" + strconv.Itoa(o.port)
	link, err := core.EnrollLink(ctx, host, o.ttl)
	if err != nil {
		return fmt.Errorf("issuing the link: %w", err)
	}
	if err := printLink(env.Out, link, host, &o); err != nil {
		return err
	}
	return untilStopped(ctx, forwarded)
}

// untilStopped waits until the owner stops the command (ctx ends) and the
// port-forward with it, or the port-forward ends on its own, an error.
func untilStopped(ctx context.Context, forwarded <-chan error) error {
	select {
	case ferr := <-forwarded:
		return forwardEnded(ferr)
	case <-ctx.Done():
		if ferr := <-forwarded; ferr != nil {
			return fmt.Errorf("stopping the port-forward: %w", ferr)
		}
		return nil
	}
}

// forwardEnded reports a port-forward that ended before the owner stopped
// it.
func forwardEnded(err error) error {
	if err != nil {
		return fmt.Errorf("%w: %w", errForwardEnded, err)
	}
	return errForwardEnded
}

// printLink tells the owner what to open and what follows.
func printLink(out io.Writer, link, host string, o *options) error {
	var warn string
	if o.port != DefaultPort {
		warn = fmt.Sprintf("\nThe core accepts http://%s only if it runs with --auth-origin=http://%s.\n", host, host)
	}
	_, err := fmt.Fprintf(out, `Forwarding http://%[1]s to the OpenFief core (kubectl port-forward).
Open this link in your browser within %[2]s to create your first passkey on %[1]s; it works once:

  %[3]s
%[4]s
After you create the passkey, the built-in UI's first-run wizard walks you through creating accounts and
sleeves and binding the connector instances you installed.

Keep this command running while you use OpenFief on http://%[1]s; press Ctrl-C to stop forwarding.
`, host, o.ttl, link, warn)
	if err != nil {
		return fmt.Errorf("writing: %w", err)
	}
	return nil
}
