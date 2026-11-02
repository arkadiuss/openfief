package operator_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"
)

// kinds lists the kind of every YAML document under dir.
func kinds(t *testing.T, dir string) map[string][]string {
	t.Helper()
	kindLine := regexp.MustCompile(`(?m)^kind:\s*(\S+)\s*$`)
	out := map[string][]string{}
	fsys := os.DirFS(dir)
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return err
		}
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		for _, m := range kindLine.FindAllStringSubmatch(string(data), -1) {
			out[path] = append(out[path], m[1])
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return out
}

// declared lists "path declares a Kind" for each of kinds declared under dir.
func declared(all map[string][]string, dir string, kinds []string) []string {
	var out []string
	for path, ks := range all {
		for _, k := range ks {
			if strings.HasPrefix(path, dir) && slices.Contains(kinds, k) {
				out = append(out, path+" declares a "+k)
			}
		}
	}
	return out
}

func TestManifests(t *testing.T) {
	t.Parallel()
	deploy := filepath.Join("..", "..", "..", "deploy")
	all := kinds(t, deploy)
	if len(all) == 0 {
		t.Fatalf("no manifests under %s", deploy)
	}
	// The product's install bases. deploy/demo is not one: the demo installation exposes them over HTTP itself.
	product := []string{"crds/", "operator/", "egress-proxy/", "core/"}
	tests := []struct {
		name      string
		dirs      []string // under deploy/
		forbidden []string
	}{
		// D-28: no Ingress and no tunnel; the core and UI are reached by port-forward.
		{name: "no Ingress or Gateway", dirs: product, forbidden: []string{"Ingress", "IngressClass", "Gateway", "HTTPRoute"}},
		// D-15: the operator holds a Role in the openfief namespace only.
		{name: "no cluster-wide RBAC for the operator", dirs: []string{"operator/"}, forbidden: []string{"ClusterRole", "ClusterRoleBinding"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, dir := range tt.dirs {
				for _, found := range declared(all, dir, tt.forbidden) {
					t.Errorf("%s", found)
				}
			}
		})
	}
	role := "operator/role.yaml"
	if got := all[role]; !slices.Equal(got, []string{"Role"}) {
		t.Errorf("%s kinds = %v, want [Role]", role, got)
	}
}

// TestNamespaceLimitRange checks the openfief namespace's LimitRange (D-182):
// the install's cluster-admin base applies it, and it gives every container
// without its own resources a request and a memory limit, so nothing in the
// namespace can grow without a bound.
func TestNamespaceLimitRange(t *testing.T) {
	t.Parallel()
	crds := os.DirFS(filepath.Join("..", "..", "..", "deploy", "crds"))
	kustomization, err := fs.ReadFile(crds, "kustomization.yaml")
	if err != nil {
		t.Fatalf("read kustomization: %v", err)
	}
	var k struct {
		Resources []string `json:"resources"`
	}
	if parseErr := yaml.Unmarshal(kustomization, &k); parseErr != nil {
		t.Fatalf("parse kustomization: %v", parseErr)
	}
	if !slices.Contains(k.Resources, "limitrange.yaml") {
		t.Errorf("deploy/crds resources = %v, want limitrange.yaml among them", k.Resources)
	}
	data, err := fs.ReadFile(crds, "limitrange.yaml")
	if err != nil {
		t.Fatalf("read LimitRange: %v", err)
	}
	var got corev1.LimitRange
	if parseErr := yaml.UnmarshalStrict(data, &got); parseErr != nil {
		t.Fatalf("parse LimitRange: %v", parseErr)
	}
	if got.Kind != "LimitRange" || got.Namespace != "openfief" {
		t.Errorf("LimitRange is %s in namespace %q, want a LimitRange in openfief", got.Kind, got.Namespace)
	}
	want := []corev1.LimitRangeItem{{
		Type:           corev1.LimitTypeContainer,
		DefaultRequest: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("50m"), corev1.ResourceMemory: resource.MustParse("64Mi")},
		Default:        corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
	}}
	if diff := cmp.Diff(want, got.Spec.Limits); diff != "" {
		t.Errorf("LimitRange limits (-want +got):\n%s", diff)
	}
}
