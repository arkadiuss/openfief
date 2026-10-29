package kube

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
)

// The core's workload, as deploy/core names it.
const (
	// CoreLabel selects the core's pods: app.kubernetes.io/name=openfief-core.
	CoreLabel = "app.kubernetes.io/name=openfief-core"
	// CoreContainer is the core's container in its pod.
	CoreContainer = "core"
	// CoreBinary is the core's binary in its image.
	CoreBinary = "/core"
	// CorePort is the port the core serves browsers, the read API and its
	// health endpoints on, in its pod.
	CorePort = 8080
)

// ErrNoCorePod reports a namespace without a running, ready core pod.
var ErrNoCorePod = errors.New("no running and ready OpenFief core pod")

// CoreAccess reaches the core's pod with the owner's own kubeconfig (D-80):
// it runs the core's commands in the pod through Kubernetes exec, and
// forwards a local port to it, as kubectl exec and kubectl port-forward
// do. The owner needs get and list on pods, and create on pods/exec and
// pods/portforward, in the namespace.
type CoreAccess struct {
	Config    *rest.Config
	Client    kubernetes.Interface
	Namespace string
}

// CorePod returns the name of a running core pod whose containers are all
// ready, the first in name order, or ErrNoCorePod.
func (a *CoreAccess) CorePod(ctx context.Context) (string, error) {
	p, err := a.readyPod(ctx)
	if err != nil {
		return "", err
	}
	return p.Name, nil
}

// podReady reports whether a pod runs, is not being deleted, and has its
// Ready condition True.
func podReady(p *corev1.Pod) bool {
	if p.DeletionTimestamp != nil || p.Status.Phase != corev1.PodRunning {
		return false
	}
	for _, c := range p.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

// EnrollLink runs `core enroll-link --host <host> --ttl <ttl>` in the core's
// pod (D-60) and returns the one-time link it prints, the first line of its
// output. The link is the only place the code is shown; it never goes into
// an error.
func (a *CoreAccess) EnrollLink(ctx context.Context, host string, ttl time.Duration) (string, error) {
	pod, err := a.CorePod(ctx)
	if err != nil {
		return "", err
	}
	req := a.Client.CoreV1().RESTClient().Post().Resource("pods").Namespace(a.Namespace).Name(pod).SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: CoreContainer,
			Command:   []string{CoreBinary, "enroll-link", "--host", host, "--ttl", ttl.String()},
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(a.Config, http.MethodPost, req.URL())
	if err != nil {
		return "", fmt.Errorf("exec in %s: %w", pod, err)
	}
	var stdout, stderr bytes.Buffer
	if err := exec.StreamWithContext(ctx, remotecommand.StreamOptions{Stdout: &stdout, Stderr: &stderr}); err != nil {
		return "", fmt.Errorf("core enroll-link in %s: %w: %s", pod, err, strings.TrimSpace(stderr.String()))
	}
	link, _, _ := strings.Cut(stdout.String(), "\n")
	if !strings.Contains(link, "/enroll#code=") {
		return "", fmt.Errorf("core enroll-link in %s printed no link: %w", pod, errNoLink)
	}
	return link, nil
}

// errNoLink reports output of core enroll-link without a link.
var errNoLink = errors.New("unexpected output")

// Forward forwards localhost:localPort, on 127.0.0.1 and ::1, to the core pod's CorePort until ctx
// ends, as kubectl port-forward does, closing ready once it listens. It
// returns nil when ctx ends.
func (a *CoreAccess) Forward(ctx context.Context, localPort int, ready chan struct{}) error {
	pod, err := a.CorePod(ctx)
	if err != nil {
		return err
	}
	transport, upgrader, err := spdy.RoundTripperFor(a.Config)
	if err != nil {
		return fmt.Errorf("port-forward: %w", err)
	}
	u := a.Client.CoreV1().RESTClient().Post().Resource("pods").Namespace(a.Namespace).Name(pod).SubResource("portforward").URL()
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, u)
	stop := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(stop)
	}()
	ports := []string{strconv.Itoa(localPort) + ":" + strconv.Itoa(CorePort)}
	fw, err := portforward.NewOnAddresses(dialer, []string{"localhost"}, ports, stop, ready, io.Discard, io.Discard)
	if err != nil {
		return fmt.Errorf("port-forward to %s: %w", pod, err)
	}
	if err := fw.ForwardPorts(); err != nil {
		return fmt.Errorf("port-forward to %s: %w", pod, err)
	}
	return nil
}

// readyPod returns the pod CorePod names.
func (a *CoreAccess) readyPod(ctx context.Context) (*corev1.Pod, error) {
	pods, err := a.Client.CoreV1().Pods(a.Namespace).List(ctx, metav1.ListOptions{LabelSelector: CoreLabel})
	if err != nil {
		return nil, fmt.Errorf("listing the core's pods: %w", err)
	}
	var ready []*corev1.Pod
	for i := range pods.Items {
		if podReady(&pods.Items[i]) {
			ready = append(ready, &pods.Items[i])
		}
	}
	if len(ready) == 0 {
		return nil, fmt.Errorf("%w in namespace %s", ErrNoCorePod, a.Namespace)
	}
	return slices.MinFunc(ready, func(p, q *corev1.Pod) int { return strings.Compare(p.Name, q.Name) }), nil
}
