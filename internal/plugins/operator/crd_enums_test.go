package operator_test

import (
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// names returns the string forms of the values of a Go enum.
func names[T ~string](vs []T) []string {
	out := make([]string, 0, len(vs))
	for _, v := range vs {
		out = append(out, string(v))
	}
	return out
}

// obj returns v as a map, failing the test when it is not one.
func obj(t *testing.T, v any, where string) map[string]any {
	t.Helper()
	out, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("%s is %T, want an object", where, v)
	}
	return out
}

// arr returns v as a list, failing the test when it is not one.
func arr(t *testing.T, v any, where string) []any {
	t.Helper()
	out, ok := v.([]any)
	if !ok {
		t.Fatalf("%s is %T, want a list", where, v)
	}
	return out
}

// child returns node[key] as a map, failing the test when it is not one.
func child(t *testing.T, node map[string]any, key, where string) map[string]any {
	t.Helper()
	return obj(t, node[key], where+"."+key)
}

// schemaAt returns the schema at path in the CRD's v1alpha1 version. A
// segment "[]" steps into an array's items, any other into a property.
func schemaAt(t *testing.T, crd map[string]any, path []string) map[string]any {
	t.Helper()
	versions := arr(t, child(t, crd, "spec", "crd")["versions"], "spec.versions")
	for _, v := range versions {
		version := obj(t, v, "version")
		if version["name"] != "v1alpha1" {
			continue
		}
		node := child(t, child(t, version, "schema", "version"), "openAPIV3Schema", "schema")
		for _, seg := range path {
			if seg == "[]" {
				node = child(t, node, "items", "array")
				continue
			}
			node = child(t, child(t, node, "properties", "object"), seg, strings.Join(path, "."))
		}
		return node
	}
	t.Fatal("no v1alpha1 version")
	return nil
}

// crdEnum returns the enum of the schema at path in the CRD file.
func crdEnum(t *testing.T, file string, path []string) []string {
	t.Helper()
	raw, err := fs.ReadFile(os.DirFS("../../../deploy/crds"), file)
	if err != nil {
		t.Fatalf("reading CRD: %v", err)
	}
	var crd map[string]any
	if err := yaml.Unmarshal(raw, &crd); err != nil {
		t.Fatalf("parsing %s: %v", file, err)
	}
	list := arr(t, schemaAt(t, crd, path)["enum"], file+" enum at "+strings.Join(path, "."))
	out := make([]string, 0, len(list))
	for _, e := range list {
		out = append(out, fmt.Sprint(e))
	}
	return out
}

// TestCRDEnumsMatchGo checks that every Enum marker of the v1alpha1 types,
// as the generated CRD says it, lists exactly the values Go accepts: a value
// the parser takes but the CRD lacks makes the API server refuse a valid
// Plugin, and the reverse admits one the operator then rejects.
func TestCRDEnumsMatchGo(t *testing.T) {
	t.Parallel()
	capabilities := names(plugins.Capabilities())
	signerKinds := names(plugins.SignerKinds())
	trusts := names(plugins.Trusts())
	tests := []struct {
		name string
		file string
		path []string
		want []string
	}{
		{"Plugin spec.manifest.capabilities", "openfief.io_plugins.yaml", []string{"spec", "manifest", "capabilities", "[]"}, capabilities},
		{"Plugin status.deployed.manifest.capabilities", "openfief.io_plugins.yaml", []string{"status", "deployed", "manifest", "capabilities", "[]"}, capabilities},
		{"Plugin status.effective.capabilities", "openfief.io_plugins.yaml", []string{"status", "effective", "capabilities", "[]"}, capabilities},
		{"Plugin status.pendingApproval.capabilities", "openfief.io_plugins.yaml", []string{"status", "pendingApproval", "capabilities", "[]"}, capabilities},
		{"PluginGrant spec.capabilities", "openfief.io_plugingrants.yaml", []string{"spec", "capabilities", "[]"}, capabilities},
		{"Plugin status.signature.trust", "openfief.io_plugins.yaml", []string{"status", "signature", "trust"}, trusts},
		{"Plugin status.signature.signerKind", "openfief.io_plugins.yaml", []string{"status", "signature", "signerKind"}, signerKinds},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := crdEnum(t, tt.file, tt.path)
			slices.Sort(got)
			want := slices.Clone(tt.want)
			slices.Sort(want)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("CRD enum differs from the Go set (-go +crd); fix the kubebuilder Enum marker and run task gen:\n%s", diff)
			}
		})
	}
}
