package main

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Errors in the rules or the document.
var (
	errBadRules    = errors.New("the rules are malformed")
	errBadDocument = errors.New("the document is not an OpenAPI document this tool can read")
	errBadStatus   = errors.New("a rule's status is not an HTTP error status")
	errNoResponse  = errors.New("a rule names a response the document's components don't define")
	errNoOperation = errors.New("a rule names an operation the document doesn't have")
	errDeclared    = errors.New("the generator already declares the status")
)

// rules are the error responses, by HTTP status, the operations declare: every operation, every operation
// whose path has a parameter, and single operations by operationId. Each response is the name of one in
// the document's components.responses. A later group's status replaces an earlier one's.
type rules struct {
	Every          map[string]string            `yaml:"every"`
	PathParameters map[string]string            `yaml:"pathParameters"`
	Operations     map[string]map[string]string `yaml:"operations"`
}

// errorStatus is an HTTP client or server error status, as OpenAPI keys a response.
var errorStatus = regexp.MustCompile(`^[45]\d{2}$`)

const (
	// strTag is YAML's tag of a string scalar.
	strTag = "!!str"
	// pairWidth is how many nodes a mapping's Content holds per entry: the key, then the value.
	pairWidth = 2
)

// httpMethods returns the keys of a path item that hold an operation.
func httpMethods() []string {
	return []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"}
}

// parseRules reads rules, refusing an unknown key or a status that is not an error status.
func parseRules(raw []byte) (rules, error) {
	var r rules
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&r); err != nil {
		return rules{}, fmt.Errorf("%w: %w", errBadRules, err)
	}
	groups := []map[string]string{r.Every, r.PathParameters}
	for _, id := range slices.Sorted(maps.Keys(r.Operations)) {
		groups = append(groups, r.Operations[id])
	}
	for _, g := range groups {
		for status := range g {
			if !errorStatus.MatchString(status) {
				return rules{}, fmt.Errorf("%w: %q", errBadStatus, status)
			}
		}
	}
	return r, nil
}

// statuses returns the responses, by status, of the operation id at a path with a parameter or without.
func (r rules) statuses(id string, pathParameter bool) map[string]string {
	out := maps.Clone(r.Every)
	if out == nil {
		out = map[string]string{}
	}
	if pathParameter {
		maps.Copy(out, r.PathParameters)
	}
	maps.Copy(out, r.Operations[id])
	return out
}

// addErrors returns the OpenAPI document spec with each operation's error responses under r added as
// references to components.responses, its responses in status order with default last. It fails on a
// response the components don't define, an operation r names that the document lacks, and a status the
// generator already declares.
func addErrors(spec []byte, r rules) ([]byte, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(spec, &root); err != nil {
		return nil, fmt.Errorf("%w: %w", errBadDocument, err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) != 1 {
		return nil, errBadDocument
	}
	doc := root.Content[0]
	defined := lookup(lookup(doc, "components"), "responses")
	paths := lookup(doc, "paths")
	if paths == nil || paths.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%w: no paths", errBadDocument)
	}
	seen := map[string]bool{}
	for i := 0; i+1 < len(paths.Content); i += pairWidth {
		if err := addToPath(paths.Content[i].Value, paths.Content[i+1], r, defined, seen); err != nil {
			return nil, err
		}
	}
	for _, id := range slices.Sorted(maps.Keys(r.Operations)) {
		if !seen[id] {
			return nil, fmt.Errorf("%w: %s", errNoOperation, id)
		}
	}
	return encode(&root)
}

// addToPath adds the error responses under r to each operation of item, the path item at path, and
// marks each operation's ID in seen.
func addToPath(path string, item *yaml.Node, r rules, defined *yaml.Node, seen map[string]bool) error {
	for _, method := range httpMethods() {
		op := lookup(item, method)
		if op == nil {
			continue
		}
		id := scalar(lookup(op, "operationId"))
		seen[id] = true
		if err := addResponses(op, r.statuses(id, strings.Contains(path, "{")), defined); err != nil {
			return fmt.Errorf("%s %s: %w", strings.ToUpper(method), path, err)
		}
	}
	return nil
}

// addResponses adds a reference to each of statuses' responses to op's responses and sorts them by
// status, default last.
func addResponses(op *yaml.Node, statuses map[string]string, defined *yaml.Node) error {
	responses := lookup(op, "responses")
	if responses == nil || responses.Kind != yaml.MappingNode {
		return fmt.Errorf("%w: an operation without responses", errBadDocument)
	}
	for _, status := range slices.Sorted(maps.Keys(statuses)) {
		name := statuses[status]
		if lookup(defined, name) == nil {
			return fmt.Errorf("%w: %s", errNoResponse, name)
		}
		if lookup(responses, status) != nil {
			return fmt.Errorf("%w: %s", errDeclared, status)
		}
		ref := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: strTag, Value: "$ref"},
			{Kind: yaml.ScalarNode, Tag: strTag, Value: "#/components/responses/" + name, Style: yaml.SingleQuotedStyle},
		}}
		key := &yaml.Node{Kind: yaml.ScalarNode, Tag: strTag, Value: status, Style: yaml.DoubleQuotedStyle}
		responses.Content = append(responses.Content, key, ref)
	}
	sortResponses(responses)
	return nil
}

// sortResponses orders a responses mapping by status, default last.
func sortResponses(responses *yaml.Node) {
	// keys holds each entry's key index in Content; its value follows it.
	keys := make([]int, 0, len(responses.Content)/pairWidth)
	for i := 0; i+1 < len(responses.Content); i += pairWidth {
		keys = append(keys, i)
	}
	rank := func(i int) string {
		if key := responses.Content[i].Value; key != "default" {
			return key
		}
		return "~" // after every status
	}
	slices.SortStableFunc(keys, func(a, b int) int { return strings.Compare(rank(a), rank(b)) })
	sorted := make([]*yaml.Node, 0, len(responses.Content))
	for _, i := range keys {
		sorted = append(sorted, responses.Content[i], responses.Content[i+1])
	}
	responses.Content = sorted
}

// lookup returns the value of key in the mapping m, or nil when m is no mapping or lacks key.
func lookup(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += pairWidth {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// scalar returns n's value, or "" when n is no scalar.
func scalar(n *yaml.Node) string {
	if n == nil || n.Kind != yaml.ScalarNode {
		return ""
	}
	return n.Value
}
