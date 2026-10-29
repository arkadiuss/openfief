//go:build e2e

// Package e2e runs OpenFief the way the demo machine does, on a local kind or
// k3d cluster, and checks it from the outside (docs/dev/testing.md). Run it
// with `task e2e`, after `task e2e:images` has loaded this checkout's images.
//
// TestMain installs, in order: CloudNativePG from its release manifest at the
// version deploy/demo/cnpg pins, the namespace and CRDs (deploy/crds), the
// demo's one-instance cluster openfief-db (deploy/demo/database), the
// suite's trusted signer (a key made for this run, which signs the plugin
// image in the suite's registry, see registry_test.go), the demo's
// application layer (deploy/demo/openfief through overlays/openfief, which
// only swaps the images, runs the built-in UI from a stand-in and admits the
// suite's client pod), and two client pods (testdata/clients.yaml). Each test then waits for its own condition,
// with a deadline, and on failure logs the cluster's pods and events, the
// pods that aren't ready, the plugins and their allow-lists, and the logs of
// the core, the operator, the plugins, the demo loader, the database and
// CloudNativePG.
package e2e

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// cnpgManifestEnv names the CloudNativePG release manifest that task e2e
	// downloaded and checked against its pinned sha256.
	cnpgManifestEnv = "OPENFIEF_E2E_CNPG_MANIFEST"
	// installTimeout bounds the whole installation.
	installTimeout = 15 * time.Minute
	// rolloutTimeout bounds the CloudNativePG operator's rollout and the
	// client pods' start.
	rolloutTimeout = 5 * time.Minute
	// databaseTimeout bounds the wait for the openfief-db cluster to be ready.
	databaseTimeout = 8 * time.Minute
	// webhookTimeout bounds retries of an apply that CloudNativePG's webhook
	// refuses while it starts.
	webhookTimeout = 2 * time.Minute
	// localPath is the storage class the demo's database asks for.
	localPath = "local-path"
)

// The repository's manifests, relative to this package.
const (
	crdsDir     = "../../deploy/crds"
	databaseDir = "../../deploy/demo/database"
	openfiefDir = "overlays/openfief"
	routesDir   = "overlays/routes"
	clientsFile = "testdata/clients.yaml"
)

var (
	// errNoManifest reports a run without the CloudNativePG manifest.
	errNoManifest = errors.New(cnpgManifestEnv + " is not set; run the suite with task e2e")
	// errNoStorageClass reports a cluster with neither local-path nor a
	// default storage class.
	errNoStorageClass = errors.New("no local-path and no default storage class (D-12 needs a storage class)")
)

// cnpgCluster is CloudNativePG's Cluster resource.
func cnpgCluster() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "postgresql.cnpg.io", Version: "v1", Resource: "clusters"}
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), installTimeout)
	k, err := connect()
	if err == nil {
		err = install(ctx, k, logStderrf)
	}
	cancel()
	if err != nil {
		logStderrf("e2e: installing the stack failed: %v", err)
		if k != nil {
			dctx, dcancel := context.WithTimeout(context.Background(), diagnosticsTimeout)
			k.dumpDiagnostics(dctx, logStderrf)
			dcancel()
		}
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// logStderrf writes one line to stderr, which go test shows as it runs.
func logStderrf(format string, args ...any) {
	if _, err := fmt.Fprintf(os.Stderr, format+"\n", args...); err != nil {
		return // nowhere left to report it
	}
}

// install puts the demo's stack on the cluster in the order of the demo's
// Flux layers, waiting where a later step needs an earlier one running. The
// trusted signer goes in before the application layer, since the operator
// reads it when it starts.
func install(ctx context.Context, k *kube, logf func(string, ...any)) error {
	reg, err := registryFromEnv()
	if err != nil {
		return err
	}
	var signers string
	steps := []func() error{
		func() error { return installCNPG(ctx, k, logf) },
		func() error { return installDatabase(ctx, k, logf) },
		func() error {
			hash, trustErr := trustSigner(ctx, k, reg, logf)
			signers = hash
			return trustErr
		},
		func() error { return installApp(ctx, k, logf, editApp(signers)) },
		func() error { return installClients(ctx, k, logf) },
	}
	for _, step := range steps {
		if stepErr := step(); stepErr != nil {
			return stepErr
		}
	}
	return nil
}

// installCNPG installs CloudNativePG from the release manifest that task e2e
// checked, and waits for its operator.
func installCNPG(ctx context.Context, k *kube, logf func(string, ...any)) error {
	manifestPath := os.Getenv(cnpgManifestEnv)
	if manifestPath == "" {
		return errNoManifest
	}
	logf("e2e: installing CloudNativePG from %s", manifestPath)
	root, err := os.OpenRoot(filepath.Dir(manifestPath))
	if err != nil {
		return fmt.Errorf("opening the CloudNativePG manifest's directory: %w", err)
	}
	manifest, err := root.ReadFile(filepath.Base(manifestPath))
	if closeErr := root.Close(); err == nil && closeErr != nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("reading the CloudNativePG manifest: %w", err)
	}
	if applyErr := k.apply(ctx, manifest); applyErr != nil {
		return applyErr
	}
	if waitErr := waitFor(ctx, rolloutTimeout, k.deploymentAvailable("cnpg-system", "cnpg-controller-manager")); waitErr != nil {
		return fmt.Errorf("waiting for the CloudNativePG operator: %w", waitErr)
	}
	return k.ensureLocalPath(ctx, logf)
}

// installDatabase applies deploy/crds and the demo's database, and waits
// until the database is ready.
func installDatabase(ctx context.Context, k *kube, logf func(string, ...any)) error {
	logf("e2e: applying deploy/crds and deploy/demo/database")
	if err := k.applyBuild(ctx, exec.CommandContext(ctx, "kustomize", "build", crdsDir)); err != nil {
		return err
	}
	database, err := build(exec.CommandContext(ctx, "kustomize", "build", databaseDir))
	if err != nil {
		return err
	}
	if err = retry(ctx, webhookTimeout, func() error { return k.apply(ctx, database) }); err != nil {
		return fmt.Errorf("applying the database: %w", err)
	}
	logf("e2e: waiting for the openfief-db cluster")
	err = waitFor(ctx, databaseTimeout, func(ctx context.Context) (string, bool, error) {
		ready, condErr := k.conditionTrue(ctx, cnpgCluster(), namespace, "openfief-db", "Ready")
		return "not ready", ready, condErr
	})
	if err != nil {
		return fmt.Errorf("waiting for the openfief-db cluster: %w", err)
	}
	return nil
}

// installApp applies the demo's application layer, each object changed by
// edit first, and, on k3s, the demo's route through Traefik.
func installApp(ctx context.Context, k *kube, logf func(string, ...any), edit func(*unstructured.Unstructured) error) error {
	logf("e2e: applying the demo's application layer")
	app, err := build(exec.CommandContext(ctx, "kustomize", "build", openfiefDir))
	if err != nil {
		return err
	}
	if applyErr := k.applyEdited(ctx, app, edit); applyErr != nil {
		return applyErr
	}
	isK3s, err := k.isK3s(ctx)
	if err != nil || !isK3s {
		return err
	}
	logf("e2e: k3s: applying the demo's route through Traefik")
	return k.applyBuild(ctx, exec.CommandContext(ctx, "kustomize", "build", routesDir))
}

// installClients applies the client pods and waits until they run.
func installClients(ctx context.Context, k *kube, logf func(string, ...any)) error {
	logf("e2e: starting the client pods")
	clients, err := os.ReadFile(clientsFile)
	if err != nil {
		return fmt.Errorf("reading the client pods: %w", err)
	}
	if applyErr := k.apply(ctx, clients); applyErr != nil {
		return applyErr
	}
	for _, pod := range []string{clientPod, outsiderPod} {
		err = waitFor(ctx, rolloutTimeout, func(ctx context.Context) (string, bool, error) {
			p, getErr := k.clients.CoreV1().Pods(clientNamespace).Get(ctx, pod, metav1.GetOptions{})
			if getErr != nil {
				return "", false, fmt.Errorf("%w: %w", errRetry, getErr)
			}
			state, stuck := podState(p)
			return state, podReady(p), stuck
		})
		if err != nil {
			return fmt.Errorf("waiting for client pod %s: %w", pod, err)
		}
	}
	return nil
}

// isK3s reports whether the cluster is k3s (k3d), the distribution the demo
// machine runs, by its nodes' kubelet version.
func (k *kube) isK3s(ctx context.Context) (bool, error) {
	nodes, err := k.clients.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("listing nodes: %w", err)
	}
	for i := range nodes.Items {
		if strings.Contains(nodes.Items[i].Status.NodeInfo.KubeletVersion, "+k3s") {
			return true, nil
		}
	}
	return false, nil
}

// deploymentAvailable checks that Deployment ns/name is available.
func (k *kube) deploymentAvailable(ns, name string) check {
	return func(ctx context.Context) (string, bool, error) {
		d, err := k.clients.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		for _, c := range d.Status.Conditions {
			if c.Type == appsv1.DeploymentAvailable && c.Status == corev1.ConditionTrue && d.Status.AvailableReplicas > 0 {
				return "available", true, nil
			}
		}
		return fmt.Sprintf("%d of %d replicas available", d.Status.AvailableReplicas, d.Status.Replicas), false, nil
	}
}

// ensureLocalPath makes sure the storage class the demo's database asks for
// exists. k3s ships local-path; kind ships the same provisioner as its
// default class under another name, so a copy named local-path is added.
func (k *kube) ensureLocalPath(ctx context.Context, logf func(string, ...any)) error {
	classes := k.clients.StorageV1().StorageClasses()
	_, err := classes.Get(ctx, localPath, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("getting storage class %s: %w", localPath, err)
	}
	list, err := classes.List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("listing storage classes: %w", err)
	}
	for i := range list.Items {
		sc := &list.Items[i]
		if sc.Annotations["storageclass.kubernetes.io/is-default-class"] != "true" {
			continue
		}
		copied := &storagev1.StorageClass{
			ObjectMeta:        metav1.ObjectMeta{Name: localPath},
			Provisioner:       sc.Provisioner,
			Parameters:        sc.Parameters,
			ReclaimPolicy:     sc.ReclaimPolicy,
			VolumeBindingMode: sc.VolumeBindingMode,
		}
		logf("e2e: adding storage class %s, a copy of the default class %s (%s)", localPath, sc.Name, sc.Provisioner)
		if _, err := classes.Create(ctx, copied, metav1.CreateOptions{FieldManager: fieldManager}); err != nil {
			return fmt.Errorf("creating storage class %s: %w", localPath, err)
		}
		return nil
	}
	return errNoStorageClass
}
