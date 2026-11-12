package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// spec is a generated document in miniature: an operation without a path parameter, one with, and the
// responses the base file defines.
const spec = `openapi: 3.1.0
components:
  responses:
    BadRequest:
      description: Invalid.
    NotFound:
      description: Not found.
    Off:
      description: Off.
paths:
  /v1/things:
    get:
      operationId: example.v1.Service.ListThings
      responses:
        "200":
          description: Success
        default:
          description: An unexpected error response.
  /v1/things/{thingId}:
    get:
      operationId: example.v1.Service.GetThing
      responses:
        "200":
          description: Success
        default:
          description: An unexpected error response.
`

// want is spec with every operation's 400, the path parameter's 404 and ListThings's own 404.
const want = `openapi: 3.1.0
components:
  responses:
    BadRequest:
      description: Invalid.
    NotFound:
      description: Not found.
    Off:
      description: Off.
paths:
  /v1/things:
    get:
      operationId: example.v1.Service.ListThings
      responses:
        "200":
          description: Success
        "400":
          $ref: '#/components/responses/BadRequest'
        "404":
          $ref: '#/components/responses/Off'
        default:
          description: An unexpected error response.
  /v1/things/{thingId}:
    get:
      operationId: example.v1.Service.GetThing
      responses:
        "200":
          description: Success
        "400":
          $ref: '#/components/responses/BadRequest'
        "404":
          $ref: '#/components/responses/NotFound'
        default:
          description: An unexpected error response.
`

// replaced returns spec with GetThing's own 404 in place of the path parameter's; ListThings has none.
func replaced() string {
	return strings.Replace(spec, `      operationId: example.v1.Service.GetThing
      responses:
        "200":
          description: Success
`, `      operationId: example.v1.Service.GetThing
      responses:
        "200":
          description: Success
        "404":
          $ref: '#/components/responses/Off'
`, 1)
}

func TestAddErrors(t *testing.T) {
	t.Parallel()
	all := rules{
		Every:          map[string]string{"400": "BadRequest"},
		PathParameters: map[string]string{"404": "NotFound"},
		Operations:     map[string]map[string]string{"example.v1.Service.ListThings": {"404": "Off"}},
	}
	tests := []struct {
		name    string
		spec    string
		rules   rules
		want    string
		wantErr error
	}{
		{"adds each group's responses in status order", spec, all, want, nil},
		{"no rules leave the document as it is", spec, rules{}, spec, nil},
		{"an operation's rule replaces a path parameter's", spec, rules{
			PathParameters: map[string]string{"404": "NotFound"},
			Operations:     map[string]map[string]string{"example.v1.Service.GetThing": {"404": "Off"}},
		}, replaced(), nil},
		{"a response the components lack", spec, rules{Every: map[string]string{"401": "Unauthorized"}}, "", errNoResponse},
		{"an operation the document lacks", spec, rules{Operations: map[string]map[string]string{"example.v1.Service.Gone": {"404": "Off"}}}, "", errNoOperation},
		{"a status the generator declares", spec, rules{Every: map[string]string{"200": "BadRequest"}}, "", errDeclared},
		{"no paths", "openapi: 3.1.0\n", all, "", errBadDocument},
		{"an operation without responses", "paths:\n  /v1/x:\n    get:\n      operationId: x\n", all, "", errBadDocument},
		{"not YAML", "paths: [", all, "", errBadDocument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := addErrors([]byte(tt.spec), tt.rules)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("addErrors() error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, string(got)); err == nil && diff != "" {
				t.Errorf("addErrors() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseRules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		raw     string
		want    rules
		wantErr error
	}{
		{"every group", "every:\n  \"400\": BadRequest\npathParameters:\n  \"404\": NotFound\noperations:\n  a.B:\n    \"404\": Off\n", rules{
			Every: map[string]string{"400": "BadRequest"}, PathParameters: map[string]string{"404": "NotFound"},
			Operations: map[string]map[string]string{"a.B": {"404": "Off"}},
		}, nil},
		{"an unknown key", "everything:\n  \"400\": BadRequest\n", rules{}, errBadRules},
		{"a success status", "every:\n  \"200\": BadRequest\n", rules{}, errBadStatus},
		{"default is no status", "pathParameters:\n  default: NotFound\n", rules{}, errBadStatus},
		{"an operation's status", "operations:\n  a.B:\n    \"4xx\": Off\n", rules{}, errBadStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseRules([]byte(tt.raw))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("parseRules() error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseRules() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestRun runs the command on files: the document is rewritten in place, and bad flags or files fail.
func TestRun(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write := func(name, text string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	rulesFile := write("rules.yaml", "every:\n  \"400\": BadRequest\npathParameters:\n  \"404\": NotFound\noperations:\n  example.v1.Service.ListThings:\n    \"404\": Off\n")
	specFile := write("spec.yaml", spec)
	if err := run([]string{"-rules", rulesFile, "-spec", specFile}); err != nil {
		t.Fatalf("run() = %v", err)
	}
	got, err := os.ReadFile(filepath.Clean(specFile))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("rewritten document mismatch (-want +got):\n%s", diff)
	}
	for _, args := range [][]string{
		{}, {"-rules", rulesFile}, {"-spec", specFile}, {"-rules", rulesFile, "-spec", specFile, "extra"}, {"-bogus"},
	} {
		if err := run(args); !errors.Is(err, errUsage) {
			t.Errorf("run(%q) = %v, want %v", args, err, errUsage)
		}
	}
	if err := run([]string{"-rules", filepath.Join(dir, "missing.yaml"), "-spec", specFile}); err == nil {
		t.Error("run() with a missing rules file succeeded")
	}
	if err := run([]string{"-rules", write("bad.yaml", "every: [\n"), "-spec", specFile}); !errors.Is(err, errBadRules) {
		t.Errorf("run() with malformed rules = %v, want %v", err, errBadRules)
	}
	if err := run([]string{"-rules", rulesFile, "-spec", filepath.Join(dir, "missing.yaml")}); err == nil {
		t.Error("run() with a missing document succeeded")
	}
}
