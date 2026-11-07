//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	authorizationv1 "k8s.io/api/authorization/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authv1client "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

// The operator's identity and the Role its install grants it (D-15).
const (
	// operatorRoleFile is the operator's Role, generated from the RBAC
	// markers in internal/plugins/operator, which the bundle installs.
	operatorRoleFile = "../../deploy/operator/role.yaml"
	// operatorUser is the operator's ServiceAccount as the API server names
	// it.
	operatorUser = "system:serviceaccount:" + namespace + ":openfief-operator"
	// reviewsQPS is how many SubjectAccessReviews a second the test sends.
	reviewsQPS = 50
)

// resourceKind is one resource the scope test asks about.
type resourceKind struct {
	group, resource string
}

// namespacedKinds are the namespaced resources the scope test asks about:
// every kind the operator's Role names, and the core kinds an operator that
// escaped its namespace would reach for, Secrets above all.
func namespacedKinds() []resourceKind {
	return []resourceKind{
		{"", "secrets"},
		{"", "pods"},
		{"", "pods/exec"},
		{"", "configmaps"},
		{"", "services"},
		{"", "serviceaccounts"},
		{"", "serviceaccounts/token"},
		{"", "persistentvolumeclaims"},
		{"", "events"},
		{"apps", "deployments"},
		{"apps", "daemonsets"},
		{"apps", "statefulsets"},
		{"batch", "jobs"},
		{"networking.k8s.io", "networkpolicies"},
		{"networking.k8s.io", "ingresses"},
		{"rbac.authorization.k8s.io", "roles"},
		{"rbac.authorization.k8s.io", "rolebindings"},
		{"coordination.k8s.io", "leases"},
		{"events.k8s.io", "events"},
		{"openfief.io", "plugins"},
		{"openfief.io", "plugins/status"},
		{"openfief.io", "plugins/finalizers"},
		{"openfief.io", "plugingrants"},
		{"openfief.io", "plugingrants/status"},
		{"postgresql.cnpg.io", "clusters"},
	}
}

// clusterKinds are cluster-scoped resources the operator must not touch.
func clusterKinds() []resourceKind {
	return []resourceKind{
		{"", "namespaces"},
		{"", "nodes"},
		{"", "persistentvolumes"},
		{"apiextensions.k8s.io", "customresourcedefinitions"},
		{"rbac.authorization.k8s.io", "clusterroles"},
		{"rbac.authorization.k8s.io", "clusterrolebindings"},
		{"admissionregistration.k8s.io", "validatingwebhookconfigurations"},
		{"admissionregistration.k8s.io", "mutatingwebhookconfigurations"},
		{"storage.k8s.io", "storageclasses"},
		{"certificates.k8s.io", "certificatesigningrequests"},
		{"authentication.k8s.io", "tokenreviews"},
		{"authorization.k8s.io", "subjectaccessreviews"},
		{"gateway.networking.k8s.io", "gatewayclasses"},
	}
}

// verbs are every verb the scope test asks about.
func verbs() []string {
	return []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"}
}

// scopeNamespaces are the namespaces the scope test asks about: OpenFief's
// own, and others where an operator that escaped its namespace would act.
func scopeNamespaces() []string {
	return []string{namespace, metav1.NamespaceDefault, metav1.NamespaceSystem, "cnpg-system"}
}

// TestBundleOperatorScope asks the API server, with SubjectAccessReviews for
// the operator's ServiceAccount and its groups, what the operator installed
// from the bundle may do, and compares every answer with what the Role in
// deploy/operator grants: in the openfief namespace exactly the Role's
// verbs on the Role's resources, and nothing at all on any other resource,
// in any other namespace or on any cluster-scoped resource. So it can't
// list or read Secrets, but for the CA bundle Secret an owner's --extra-ca
// install lets it get by name, nor create or change anything outside
// openfief (D-15; SCEN-e2e-operator-namespace-scoped).
func TestBundleOperatorScope(t *testing.T) {
	holdsCore(t)
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	role := readOperatorRole(t)
	reviews := k.reviews(t)
	want := map[string]bool{}
	got := map[string]bool{}
	ask := func(ns string, r resourceKind, verb string) {
		key := fmt.Sprintf("%s %s/%s in %q", verb, r.group, r.resource, ns)
		want[key] = ns == namespace && roleAllows(role, r, verb)
		allowed, err := operatorMay(t.Context(), reviews, ns, r, verb)
		if err != nil {
			t.Fatal(err)
		}
		got[key] = allowed
	}
	for _, ns := range scopeNamespaces() {
		for _, r := range namespacedKinds() {
			for _, verb := range verbs() {
				ask(ns, r, verb)
			}
		}
	}
	for _, r := range clusterKinds() {
		for _, verb := range verbs() {
			ask("", r, verb)
		}
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("what the operator's ServiceAccount may do, by deploy/operator's Role (-want +got):\n%s", diff)
	}
	if !slices.Contains(slices.Collect(maps.Values(got)), true) {
		t.Error("the operator may do nothing at all; the SubjectAccessReviews asked about the wrong identity")
	}
}

// reviews returns an authorization client allowed reviewsQPS requests a
// second: the scope test sends about a thousand SubjectAccessReviews, which
// client-go's default of five a second would take minutes over.
func (k *kube) reviews(t *testing.T) *authv1client.AuthorizationV1Client {
	t.Helper()
	cfg := rest.CopyConfig(k.config)
	cfg.QPS, cfg.Burst = reviewsQPS, 2*reviewsQPS
	client, err := authv1client.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("authorization client: %v", err)
	}
	return client
}

// readOperatorRole reads the operator's Role from deploy/operator.
func readOperatorRole(t *testing.T) *rbacv1.Role {
	t.Helper()
	data, err := os.ReadFile(operatorRoleFile)
	if err != nil {
		t.Fatalf("reading the operator's Role: %v", err)
	}
	var role rbacv1.Role
	if err := yaml.UnmarshalStrict(data, &role); err != nil {
		t.Fatalf("decoding the operator's Role: %v", err)
	}
	if role.Namespace != namespace || len(role.Rules) == 0 {
		t.Fatalf("the operator's Role is %s/%s with %d rules; want one in %s with rules", role.Namespace, role.Name, len(role.Rules), namespace)
	}
	return &role
}

// roleAllows reports whether role grants verb on r: a rule naming its group,
// resource and verb, with no resource names. The Role uses no wildcards.
func roleAllows(role *rbacv1.Role, r resourceKind, verb string) bool {
	for i := range role.Rules {
		rule := &role.Rules[i]
		if len(rule.ResourceNames) == 0 && slices.Contains(rule.APIGroups, r.group) &&
			slices.Contains(rule.Resources, r.resource) && slices.Contains(rule.Verbs, verb) {
			return true
		}
	}
	return false
}

// operatorMay asks the API server whether the operator's ServiceAccount may
// do verb on r in ns, or cluster-wide for an empty ns, with the groups the
// API server gives every ServiceAccount of the namespace.
func operatorMay(ctx context.Context, reviews *authv1client.AuthorizationV1Client, ns string, r resourceKind, verb string) (bool, error) {
	resource, sub, _ := strings.Cut(r.resource, "/")
	review := &authorizationv1.SubjectAccessReview{Spec: authorizationv1.SubjectAccessReviewSpec{
		User:   operatorUser,
		Groups: []string{"system:serviceaccounts", "system:serviceaccounts:" + namespace, "system:authenticated"},
		ResourceAttributes: &authorizationv1.ResourceAttributes{
			Namespace: ns, Verb: verb, Group: r.group, Resource: resource, Subresource: sub,
		},
	}}
	got, err := reviews.SubjectAccessReviews().Create(ctx, review, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("SubjectAccessReview for %s %s/%s in %q: %w", verb, r.group, r.resource, ns, err)
	}
	return got.Status.Allowed, nil
}
