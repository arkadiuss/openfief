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
