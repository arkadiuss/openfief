package operator_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// Synthetic image digests; not of any real image.
const (
	imageA = "registry.example.com/openfief/sample-connector@sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"
	imageB = "registry.example.com/openfief/sample-connector@sha256:0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f"
)

// signedImages is a test ImageVerifier: the listed images are verified,
// every other one is rejected.
type signedImages []string

func (s signedImages) VerifyImage(_ context.Context, image plugins.ImageRef, _ *plugins.Manifest) operator.Verification {
	for _, i := range s {
		if i == image.String() {
			return operator.Verification{Status: metav1.ConditionTrue, Reason: "Verified", Message: "test signature"}
		}
	}
	return operator.Verification{Status: metav1.ConditionFalse, Reason: "Unsigned", Message: "no test signature"}
}

// cluster is one envtest API server shared by the subtests of TestReconcile.
type testCluster struct {
	client client.Client
	scheme *runtime.Scheme
	cfg    *envtest.Environment
}

// startCluster runs an API server and etcd from KUBEBUILDER_ASSETS (set by
// `task test`) with the CRDs from deploy/crds, and creates the openfief
// namespace and the core and egress proxy Services.
func startCluster(t *testing.T) *testCluster {
	t.Helper()
	if os.Getenv("KUBEBUILDER_ASSETS") == "" {
		t.Fatal("KUBEBUILDER_ASSETS is not set: run the tests with `task test`, which installs the envtest binaries")
	}
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("client-go scheme: %v", err)
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("v1alpha1 scheme: %v", err)
	}
	env := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "deploy", "crds")},
		ErrorIfCRDPathMissing: true,
		Scheme:                scheme,
	}
	cfg, err := env.Start()
	if err != nil {
		t.Fatalf("start envtest: %v", err)
	}
	t.Cleanup(func() {
		if stopErr := env.Stop(); stopErr != nil {
			t.Errorf("stop envtest: %v", stopErr)
		}
	})
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	ctx := t.Context()
	create(ctx, t, c, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "openfief"}})
	create(ctx, t, c, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "openfief-core", Namespace: "openfief"},
		Spec: corev1.ServiceSpec{Selector: coreSelector(), Ports: []corev1.ServicePort{
			{Name: "connect", Port: 80, TargetPort: intstr.FromInt32(8080)},
		}},
	})
	create(ctx, t, c, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "openfief-egress-proxy", Namespace: "openfief"},
		Spec:       corev1.ServiceSpec{Selector: proxySelector(), Ports: []corev1.ServicePort{{Name: "proxy", Port: 4750}}},
	})
	return &testCluster{client: c, scheme: scheme, cfg: env}
}

func create(ctx context.Context, t *testing.T, c client.Client, obj client.Object) {
	t.Helper()
	if err := c.Create(ctx, obj); err != nil {
		t.Fatalf("create %s: %v", obj.GetName(), err)
	}
}

func (tc *testCluster) reconciler(v operator.ImageVerifier) *operator.PluginReconciler {
	return &operator.PluginReconciler{
		Client: tc.client, Scheme: tc.scheme, Verifier: v,
		Settings: operator.Settings{
			CoreService: "openfief-core", CorePort: "connect",
			EgressProxyService: "openfief-egress-proxy", EgressProxyPort: "proxy",
			PluginPort: 9090, CoreServiceAccount: "openfief-core",
		},
	}
}

// reconcile runs one reconcile of the named Plugin.
func (tc *testCluster) reconcile(ctx context.Context, t *testing.T, r *operator.PluginReconciler, name string) {
	t.Helper()
	req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "openfief", Name: name}}
	if _, err := r.Reconcile(ctx, req); err != nil {
		t.Fatalf("Reconcile(%s): %v", name, err)
	}
}

func samplePlugin(name, image string, egress ...string) *v1alpha1.Plugin {
	return &v1alpha1.Plugin{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "openfief"},
		Spec: v1alpha1.PluginSpec{
			Image: image,
			Manifest: v1alpha1.Manifest{
				Capabilities: []string{"connector"},
				Egress:       egress,
				Secrets:      []v1alpha1.SecretMount{{Role: "token", MountPath: "/etc/openfief/token"}},
				ConfigSchema: &runtime.RawExtension{Raw: []byte(`{"type":"object","properties":{"interval":{"type":"string"}}}`)},
			},
			Config: &runtime.RawExtension{Raw: []byte(`{"interval":"24h"}`)},
		},
	}
}

func sampleGrant(name string, egress ...string) *v1alpha1.PluginGrant {
	return &v1alpha1.PluginGrant{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "openfief"},
		Spec:       v1alpha1.Permissions{Capabilities: []string{"connector"}, Egress: egress, Secrets: []string{name + "-token"}},
	}
}

// state is what a test checks after a reconcile.
type state struct {
	Image          string // the Deployment's image; "" when there is none
	ServiceAccount bool
	ProxyRule      bool // the NetworkPolicy allows egress to the proxy
	// TokenAudiences are the pod's projected tokens, each as "<audience>
	// <variable naming its file>".
	TokenAudiences  []string
	DenyAll         bool // the NetworkPolicy allows nothing
	AllowList       string
	Conditions      map[string]string // type -> "Status/Reason"
	DeployedImage   string
	EffectiveEgress []string
	Pending         *v1alpha1.Permissions
	Signature       *v1alpha1.SignatureStatus
}

func (tc *testCluster) state(ctx context.Context, t *testing.T, name string) state {
	t.Helper()
	s := tc.objectState(ctx, t, name)
	var cm corev1.ConfigMap
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: "openfief-egress-allowlist"}, &cm); err != nil {
		t.Fatalf("get allow-list: %v", err)
	}
	s.AllowList = cm.Data[name]
	var p v1alpha1.Plugin
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: name}, &p); err != nil {
		t.Fatalf("get plugin: %v", err)
	}
	s.Conditions = map[string]string{}
	for _, c := range p.Status.Conditions {
		s.Conditions[c.Type] = string(c.Status) + "/" + c.Reason
	}
	if p.Status.Deployed != nil {
		s.DeployedImage = p.Status.Deployed.Image
	}
	if p.Status.Effective != nil {
		s.EffectiveEgress = p.Status.Effective.Egress
	}
	s.Pending = p.Status.PendingApproval
	s.Signature = p.Status.Signature
	return s
}

// objectState reads the plugin's Deployment, ServiceAccount and
// NetworkPolicy.
func (tc *testCluster) objectState(ctx context.Context, t *testing.T, name string) state {
	t.Helper()
	var s state
	key := types.NamespacedName{Namespace: "openfief", Name: "plugin-" + name}
	var d appsv1.Deployment
	switch err := tc.client.Get(ctx, key, &d); {
	case err == nil:
		s.Image = d.Spec.Template.Spec.Containers[0].Image
		s.TokenAudiences = tokenAudiences(&d)
	case !apierrors.IsNotFound(err):
		t.Fatalf("get deployment: %v", err)
	}
	var sa corev1.ServiceAccount
	s.ServiceAccount = tc.client.Get(ctx, key, &sa) == nil
	var np networkingv1.NetworkPolicy
	if err := tc.client.Get(ctx, key, &np); err != nil {
		t.Fatalf("get network policy: %v", err)
	}
	s.DenyAll = len(np.Spec.Ingress) == 0 && len(np.Spec.Egress) == 0 && len(np.Spec.PolicyTypes) == 2
	for _, e := range np.Spec.Egress {
		s.ProxyRule = s.ProxyRule || cmp.Equal(e.To[0].PodSelector.MatchLabels, proxySelector())
	}
	return s
}

// tokenAudiences lists "<audience> <variable>" for each projected token
// volume of d's pod, paired with the variable whose value is its file.
func tokenAudiences(d *appsv1.Deployment) []string {
	spec := &d.Spec.Template.Spec
	c := &spec.Containers[0]
	mountOf := map[string]string{}
	for _, m := range c.VolumeMounts {
		mountOf[m.Name] = m.MountPath
	}
	varOf := map[string]string{}
	for _, e := range c.Env {
		varOf[e.Value] = e.Name
	}
	var out []string
	for i := range spec.Volumes {
		v := &spec.Volumes[i]
		if v.Projected == nil {
			continue
		}
		for _, src := range v.Projected.Sources {
			if t := src.ServiceAccountToken; t != nil {
				out = append(out, t.Audience+" "+varOf[mountOf[v.Name]+"/"+t.Path])
			}
		}
	}
	slices.Sort(out)
	return out
}

func conditions(ready, approval, signature string) map[string]string {
	return map[string]string{"Ready": ready, "ApprovalRequired": approval, "SignatureVerified": signature}
}

func TestReconcile(t *testing.T) {
	t.Parallel()
	tc := startCluster(t)
	signed := tc.reconciler(signedImages{imageA, imageB})
	t.Run("plugin without a grant runs nothing", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		create(ctx, t, tc.client, samplePlugin("nogrant", imageA, "api.example.com"))
		tc.reconcile(ctx, t, signed, "nogrant")
		want := state{
			DenyAll:    true,
			Conditions: conditions("False/NoGrant", "True/PermissionsPending", "True/Verified"),
			Pending: &v1alpha1.Permissions{
				Capabilities: []string{"connector"}, Egress: []string{"api.example.com:443"}, Secrets: []string{"nogrant-token"},
			},
		}
		if diff := cmp.Diff(want, tc.state(ctx, t, "nogrant"), cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("state mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("granted plugin runs, and an upgrade waits for approval", func(t *testing.T) {
		t.Parallel()
		testUpgrade(t, tc, signed)
	})
	t.Run("granted plugin without egress gets only the core token", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		create(ctx, t, tc.client, samplePlugin("noegress", imageA))
		create(ctx, t, tc.client, sampleGrant("noegress"))
		tc.reconcile(ctx, t, signed, "noegress")
		want := state{
			Image: imageA, ServiceAccount: true,
			TokenAudiences: []string{"openfief-core OPENFIEF_TOKEN_FILE"},
			Conditions:     conditions("False/Progressing", "False/Approved", "True/Verified"),
			DeployedImage:  imageA,
		}
		if diff := cmp.Diff(want, tc.state(ctx, t, "noegress"), cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("state mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("revoking the grant removes the plugin", func(t *testing.T) {
		t.Parallel()
		testRevoke(t, tc, signed)
	})
	t.Run("unverified image is never deployed", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		create(ctx, t, tc.client, samplePlugin("unsigned", imageA))
		create(ctx, t, tc.client, sampleGrant("unsigned"))
		tc.reconcile(ctx, t, tc.reconciler(operator.UnimplementedVerifier{}), "unsigned")
		want := state{
			DenyAll:    true,
			Conditions: conditions("False/SignatureNotVerified", "False/Approved", "Unknown/NotImplemented"),
		}
		if diff := cmp.Diff(want, tc.state(ctx, t, "unsigned"), cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("state mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("signature checks run through the trust policy", func(t *testing.T) {
		t.Parallel()
		testPolicyVerifier(t, tc)
	})
	t.Run("an image that could not be checked is checked again", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		create(ctx, t, tc.client, samplePlugin("unchecked", imageA))
		req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "openfief", Name: "unchecked"}}
		res, err := tc.reconciler(operator.UnimplementedVerifier{}).Reconcile(ctx, req)
		if err != nil {
			t.Fatalf("Reconcile: %v", err)
		}
		if res.RequeueAfter <= 0 {
			t.Errorf("Reconcile result %+v, want a requeue", res)
		}
	})
	t.Run("config its schema rejects is never deployed", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		p := samplePlugin("badconfig", imageA)
		p.Spec.Config = &runtime.RawExtension{Raw: []byte(`{"interval":24}`)}
		create(ctx, t, tc.client, p)
		create(ctx, t, tc.client, sampleGrant("badconfig"))
		tc.reconcile(ctx, t, signed, "badconfig")
		want := state{
			DenyAll:    true,
			Conditions: conditions("False/InvalidConfig", "False/Approved", "True/Verified"),
		}
		if diff := cmp.Diff(want, tc.state(ctx, t, "badconfig"), cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("state mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("invalid grant is marked and enforced as no grant", func(t *testing.T) {
		t.Parallel()
		testInvalidGrant(t, tc, signed)
	})
}

func testUpgrade(t *testing.T, tc *testCluster, r *operator.PluginReconciler) {
	t.Helper()
	ctx := t.Context()
	create(ctx, t, tc.client, samplePlugin("upgrade", imageA, "api.example.com"))
	create(ctx, t, tc.client, sampleGrant("upgrade", "api.example.com"))
	tc.reconcile(ctx, t, r, "upgrade")
	running := state{
		Image: imageA, ServiceAccount: true, ProxyRule: true, AllowList: "api.example.com:443\n",
		TokenAudiences: []string{"openfief-core OPENFIEF_TOKEN_FILE", "openfief-egress OPENFIEF_EGRESS_TOKEN_FILE"},
		Conditions:     conditions("False/Progressing", "False/Approved", "True/Verified"),
		DeployedImage:  imageA, EffectiveEgress: []string{"api.example.com:443"},
	}
	if diff := cmp.Diff(running, tc.state(ctx, t, "upgrade"), cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("after grant (-want +got):\n%s", diff)
	}

	var p v1alpha1.Plugin
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: "upgrade"}, &p); err != nil {
		t.Fatalf("get plugin: %v", err)
	}
	p.Spec = samplePlugin("upgrade", imageB, "api.example.com", "new.example.net:8443").Spec
	if err := tc.client.Update(ctx, &p); err != nil {
		t.Fatalf("update plugin: %v", err)
	}
	tc.reconcile(ctx, t, r, "upgrade")
	waiting := running
	waiting.Conditions = conditions("False/Progressing", "True/PermissionsPending", "True/Verified")
	waiting.Pending = &v1alpha1.Permissions{Egress: []string{"new.example.net:8443"}}
	if diff := cmp.Diff(waiting, tc.state(ctx, t, "upgrade"), cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("after upgrade requesting a new host (-want +got):\n%s", diff)
	}

	var g v1alpha1.PluginGrant
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: "upgrade"}, &g); err != nil {
		t.Fatalf("get grant: %v", err)
	}
	g.Spec.Egress = append(g.Spec.Egress, "new.example.net:8443")
	if err := tc.client.Update(ctx, &g); err != nil {
		t.Fatalf("update grant: %v", err)
	}
	tc.reconcile(ctx, t, r, "upgrade")
	approved := running
	approved.Image, approved.DeployedImage = imageB, imageB
	approved.AllowList = "api.example.com:443\nnew.example.net:8443\n"
	approved.EffectiveEgress = []string{"api.example.com:443", "new.example.net:8443"}
	if diff := cmp.Diff(approved, tc.state(ctx, t, "upgrade"), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("after approval (-want +got):\n%s", diff)
	}
}

func testRevoke(t *testing.T, tc *testCluster, r *operator.PluginReconciler) {
	t.Helper()
	ctx := t.Context()
	create(ctx, t, tc.client, samplePlugin("revoke", imageA, "api.example.com"))
	g := sampleGrant("revoke", "api.example.com")
	create(ctx, t, tc.client, g)
	tc.reconcile(ctx, t, r, "revoke")
	if got := tc.state(ctx, t, "revoke"); got.Image != imageA || got.AllowList == "" {
		t.Fatalf("granted plugin not running: %+v", got)
	}
	if err := tc.client.Delete(ctx, g); err != nil {
		t.Fatalf("delete grant: %v", err)
	}
	tc.reconcile(ctx, t, r, "revoke")
	want := state{
		DenyAll:    true,
		Conditions: conditions("False/NoGrant", "True/PermissionsPending", "True/Verified"),
		Pending: &v1alpha1.Permissions{
			Capabilities: []string{"connector"}, Egress: []string{"api.example.com:443"}, Secrets: []string{"revoke-token"},
		},
	}
	if diff := cmp.Diff(want, tc.state(ctx, t, "revoke"), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("after revoking (-want +got):\n%s", diff)
	}
}

func testInvalidGrant(t *testing.T, tc *testCluster, r *operator.PluginReconciler) {
	t.Helper()
	ctx := t.Context()
	create(ctx, t, tc.client, samplePlugin("badgrant", imageA, "api.example.com"))
	create(ctx, t, tc.client, sampleGrant("badgrant", "*.example.com"))
	tc.reconcile(ctx, t, r, "badgrant")
	want := state{
		DenyAll:    true,
		Conditions: conditions("False/InvalidGrant", "Unknown/InvalidGrant", "True/Verified"),
	}
	if diff := cmp.Diff(want, tc.state(ctx, t, "badgrant"), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("plugin state mismatch (-want +got):\n%s", diff)
	}
	gr := &operator.GrantReconciler{Client: tc.client}
	req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "openfief", Name: "badgrant"}}
	if _, err := gr.Reconcile(ctx, req); err != nil {
		t.Fatalf("GrantReconciler.Reconcile: %v", err)
	}
	var g v1alpha1.PluginGrant
	if err := tc.client.Get(ctx, req.NamespacedName, &g); err != nil {
		t.Fatalf("get grant: %v", err)
	}
	if c := meta.FindStatusCondition(g.Status.Conditions, "Valid"); c == nil || c.Status != metav1.ConditionFalse {
		t.Errorf("grant Valid condition = %+v, want False", c)
	}
}

func TestDeletedPluginLeavesTheAllowList(t *testing.T) {
	t.Parallel()
	tc := startCluster(t)
	ctx := t.Context()
	r := tc.reconciler(signedImages{imageA})
	p := samplePlugin("gone", imageA, "api.example.com")
	create(ctx, t, tc.client, p)
	create(ctx, t, tc.client, sampleGrant("gone", "api.example.com"))
	tc.reconcile(ctx, t, r, "gone")
	if err := tc.client.Delete(ctx, p); err != nil {
		t.Fatalf("delete plugin: %v", err)
	}
	tc.reconcile(ctx, t, r, "gone")
	var cm corev1.ConfigMap
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: "openfief-egress-allowlist"}, &cm); err != nil {
		t.Fatalf("get allow-list: %v", err)
	}
	if diff := cmp.Diff(map[string]string{}, cm.Data, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("allow-list after deletion (-want +got):\n%s", diff)
	}
}

func TestSetupWithManager(t *testing.T) {
	t.Parallel()
	tc := startCluster(t)
	mgr, err := ctrl.NewManager(tc.cfg.Config, ctrl.Options{
		Scheme:  tc.scheme,
		Cache:   cache.Options{DefaultNamespaces: map[string]cache.Config{"openfief": {}}},
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	r := tc.reconciler(operator.UnimplementedVerifier{})
	r.Client = mgr.GetClient()
	if err := r.SetupWithManager(mgr); err != nil {
		t.Errorf("PluginReconciler.SetupWithManager: %v", err)
	}
	if err := (&operator.GrantReconciler{Client: mgr.GetClient()}).SetupWithManager(mgr); err != nil {
		t.Errorf("GrantReconciler.SetupWithManager: %v", err)
	}
}

// testPolicyVerifier runs the reconciler with the operator's PolicyVerifier:
// an image signed by OpenFief's CI whose signed manifest label is the
// Plugin's manifest runs; one whose label differs never does (D-42).
func testPolicyVerifier(t *testing.T, tc *testCluster) {
	t.Helper()
	ctx := t.Context()
	label, err := plugins.ParseManifestLabel(`{"capabilities":["connector"],"egress":["api.example.com"],` +
		`"secrets":[{"role":"token","mountPath":"/etc/openfief/token"}],` +
		`"configSchema":{"type":"object","properties":{"interval":{"type":"string"}}}}`)
	if err != nil {
		t.Fatalf("label: %v", err)
	}
	src := &fakeSource{images: map[string]plugins.SignedImage{
		imageB: {Image: mustImage(t, imageB), Signatures: []plugins.SignerIdentity{ciSignature()}, Manifest: label},
	}}
	r := tc.reconciler(operator.NewPolicyVerifier(src, testPolicy(t)))
	for _, name := range []string{"signedci", "relabelled"} {
		p := samplePlugin(name, imageB, "api.example.com")
		if name == "relabelled" { // declares a mount the signed label doesn't
			p.Spec.Manifest.Secrets[0].MountPath = "/etc/openfief/elsewhere"
		}
		create(ctx, t, tc.client, p)
		create(ctx, t, tc.client, sampleGrant(name, "api.example.com"))
		tc.reconcile(ctx, t, r, name)
	}
	running := state{
		Image: imageB, ServiceAccount: true, ProxyRule: true, AllowList: "api.example.com:443\n",
		TokenAudiences: []string{"openfief-core OPENFIEF_TOKEN_FILE", "openfief-egress OPENFIEF_EGRESS_TOKEN_FILE"},
		Conditions:     conditions("False/Progressing", "False/Approved", "True/SignedByTrustedSigner"),
		DeployedImage:  imageB, EffectiveEgress: []string{"api.example.com:443"},
		Signature: &v1alpha1.SignatureStatus{Trust: "Unverified", SignerKind: "OpenFiefCI", Signer: "openfief-ci"},
	}
	if diff := cmp.Diff(running, tc.state(ctx, t, "signedci"), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("label matches: state mismatch (-want +got):\n%s", diff)
	}
	refused := state{
		DenyAll:    true,
		Conditions: conditions("False/SignatureNotVerified", "False/Approved", "False/ManifestMismatch"),
	}
	if diff := cmp.Diff(refused, tc.state(ctx, t, "relabelled"), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("label differs: state mismatch (-want +got):\n%s", diff)
	}
}
