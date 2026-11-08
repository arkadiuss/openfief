// Command openapierrors adds the read API's error responses to its generated OpenAPI document
// (docs/dev/pipeline.md#contracts). protoc-gen-connect-openapi declares only an operation's success and its
// default error; this tool adds, by HTTP status, the responses a rules file assigns to every operation, to
// every operation with a path parameter, and to single operations by operationId. The responses themselves
// are defined in the document's components (the generator's base file). It is a development tool, not part
// of the product: task proto:gen runs it after buf generate, and scripts/proto-check.sh does the same before
// comparing, so the committed document always carries the responses.
//
//	openapierrors -rules FILE -spec FILE
//
// It rewrites -spec in place, keeping the generator's formatting.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

const (
	// specMode is the document's file mode when the tool creates it; an existing file keeps its own.
	specMode = 0o600
	// indent is the generator's indent, in spaces.
	indent = 2
)

// errUsage reports a missing or unknown flag.
var errUsage = errors.New("usage: openapierrors -rules FILE -spec FILE")

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "openapierrors:", err)
		os.Exit(1)
	}
}

// run reads the rules and the document its flags name and writes the document back with the error
// responses added.
func run(args []string) error {
	fs := flag.NewFlagSet("openapierrors", flag.ContinueOnError)
	rulesPath := fs.String("rules", "", "the error responses by status, as read_service.openapi-errors.yaml")
	specPath := fs.String("spec", "", "the generated OpenAPI document, rewritten in place")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w: %w", errUsage, err)
	}
	if *rulesPath == "" || *specPath == "" || fs.NArg() != 0 {
		return errUsage
	}
	rawRules, err := os.ReadFile(*rulesPath)
	if err != nil {
		return fmt.Errorf("reading the rules: %w", err)
	}
	r, err := parseRules(rawRules)
	if err != nil {
		return fmt.Errorf("%s: %w", *rulesPath, err)
	}
	spec, err := os.ReadFile(*specPath)
	if err != nil {
		return fmt.Errorf("reading the document: %w", err)
	}
	out, err := addErrors(spec, r)
	if err != nil {
		return fmt.Errorf("%s: %w", *specPath, err)
	}
	if err := os.WriteFile(*specPath, out, specMode); err != nil {
		return fmt.Errorf("writing the document: %w", err)
	}
	return nil
}

// encode writes doc as YAML the way protoc-gen-connect-openapi does: two-space indents, sequences
// indented under their key.
func encode(doc *yaml.Node) ([]byte, error) {
	var out bytes.Buffer
	enc := yaml.NewEncoder(&out)
	enc.SetIndent(indent)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("encoding the document: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("encoding the document: %w", err)
	}
	return out.Bytes(), nil
}
