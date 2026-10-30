package kube

import (
	"context"
	"fmt"
	"strings"
)

// DefaultAuthOrigin is the one origin the core allows when it runs without
// --auth-origin, as cmd/core's flag default says: kubectl port-forward to
// localhost:8080.
const DefaultAuthOrigin = "http://localhost:8080"

// authOriginFlag is the core's flag that names an allowed origin.
const authOriginFlag = "auth-origin"

// AllowedOrigins returns the origins the running core accepts the owner's
// browser from (D-60): the values of the --auth-origin flags its container
// in a running, ready core pod was started with, or DefaultAuthOrigin when
// there are none. The owner needs get and list on pods, which CorePod
// needs too.
func (a *CoreAccess) AllowedOrigins(ctx context.Context) ([]string, error) {
	pod, err := a.readyPod(ctx)
	if err != nil {
		return nil, err
	}
	for i := range pod.Spec.Containers {
		if c := &pod.Spec.Containers[i]; c.Name == CoreContainer {
			return AuthOrigins(append(append([]string(nil), c.Command...), c.Args...)), nil
		}
	}
	return nil, fmt.Errorf("%w: pod %s has no container %s", ErrNoCorePod, pod.Name, CoreContainer)
}

// AuthOrigins returns the values of the --auth-origin flags in a core's
// command line, written --auth-origin=V or --auth-origin V, with one dash or
// two as Go's flag package reads them, in order; DefaultAuthOrigin when
// there are none.
func AuthOrigins(args []string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		name, value, hasValue := strings.Cut(strings.TrimPrefix(strings.TrimPrefix(args[i], "-"), "-"), "=")
		if !strings.HasPrefix(args[i], "-") || name != authOriginFlag {
			continue
		}
		if !hasValue {
			if i+1 >= len(args) {
				break
			}
			i++
			value = args[i]
		}
		out = append(out, value)
	}
	if len(out) == 0 {
		return []string{DefaultAuthOrigin}
	}
	return out
}
