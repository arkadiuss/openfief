package main

import (
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

const crdDir = "deploy/crds"

// field is one row of a custom resource's schema, flattened to a dotted path.
type field struct {
	path        string
	typ         string
	required    bool
	description string
}

// loadCRDs reads every CustomResourceDefinition in deploy/crds, in file-name order.
func loadCRDs(repo fs.FS) ([]apiextv1.CustomResourceDefinition, error) {
	entries, err := fs.ReadDir(repo, crdDir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", crdDir, err)
	}
	var crds []apiextv1.CustomResourceDefinition
	for _, e := range entries {
		data, err := fs.ReadFile(repo, path.Join(crdDir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		var crd apiextv1.CustomResourceDefinition
		if err := yaml.Unmarshal(data, &crd); err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		if crd.Kind == "CustomResourceDefinition" {
			crds = append(crds, crd)
		}
	}
	return crds, nil
}

// schemaFields flattens a schema's properties into rows, depth first, in property-name order.
func schemaFields(prefix string, s *apiextv1.JSONSchemaProps) []field {
	if s == nil {
		return nil
	}
	if s.Type == "array" && s.Items != nil && s.Items.Schema != nil {
		return schemaFields(prefix+"[]", s.Items.Schema)
	}
	var fields []field
	for _, name := range slices.Sorted(maps.Keys(s.Properties)) {
		prop := s.Properties[name]
		p := name
		if prefix != "" {
			p = prefix + "." + name
		}
		fields = append(fields, field{path: p, typ: typeName(&prop), required: slices.Contains(s.Required, name), description: firstSentence(prop.Description)})
		fields = append(fields, schemaFields(p, &prop)...)
	}
	return fields
}

func typeName(s *apiextv1.JSONSchemaProps) string {
	t := s.Type
	if t == "array" && s.Items != nil && s.Items.Schema != nil {
		t = "[]" + s.Items.Schema.Type
	}
	if len(s.Enum) > 0 {
		values := make([]string, 0, len(s.Enum))
		for _, v := range s.Enum {
			values = append(values, strings.Trim(string(v.Raw), `"`))
		}
		t += ": " + strings.Join(values, ", ")
	}
	if s.Pattern != "" {
		t += " matching `" + strings.ReplaceAll(s.Pattern, "|", "\\|") + "`"
	}
	return t
}

// crdPage renders the custom-resource reference from the generated CRDs.
func crdPage(crds []apiextv1.CustomResourceDefinition) string {
	var b strings.Builder
	b.WriteString("---\ntitle: Custom resources\ndescription: The fields of every OpenFief custom resource, read from the CRDs the operator's Go types generate.\nweight: 20\n---\n\n")
	b.WriteString("Generated from [deploy/crds](../../../../deploy/crds/) on every build; controller-gen writes those files from the Go types (`task gen`). The [operator](../../../../docs/concepts/operator.md) page explains how the resources are used.\n")
	for i := range crds {
		crd := &crds[i]
		fmt.Fprintf(&b, "\n## %s\n\n`%s/%s`, %s-scoped, plural `%s`.\n", crd.Spec.Names.Kind, crd.Spec.Group, versionNames(crd), strings.ToLower(string(crd.Spec.Scope)), crd.Spec.Names.Plural)
		for j := range crd.Spec.Versions {
			writeVersionTable(&b, crd.Spec.Names.Kind, &crd.Spec.Versions[j])
		}
	}
	return b.String()
}

// writeVersionTable writes one version's spec and status fields; apiVersion, kind and metadata are standard.
func writeVersionTable(b *strings.Builder, kind string, v *apiextv1.CustomResourceDefinitionVersion) {
	if v.Schema == nil || v.Schema.OpenAPIV3Schema == nil {
		return
	}
	fmt.Fprintf(b, "\n### %s %s\n\n| Field | Type | Required | Description |\n|---|---|---|---|\n", kind, v.Name)
	for _, f := range schemaFields("", v.Schema.OpenAPIV3Schema) {
		if f.path == "apiVersion" || f.path == "kind" || strings.HasPrefix(f.path, "metadata") {
			continue
		}
		req := ""
		if f.required {
			req = "yes"
		}
		fmt.Fprintf(b, "| `%s` | %s | %s | %s |\n", f.path, f.typ, req, f.description)
	}
}

func versionNames(crd *apiextv1.CustomResourceDefinition) string {
	names := make([]string, 0, len(crd.Spec.Versions))
	for _, v := range crd.Spec.Versions {
		names = append(names, v.Name)
	}
	return strings.Join(names, ", ")
}
