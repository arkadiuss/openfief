// Package apicontract walks the read API's contract for the end-to-end
// suite's check of D-200's criterion 7: the operations its OpenAPI
// document lists, the HTTP routes the protobuf service's google.api.http
// annotations declare (D-50), the REST target and the Connect request of
// one call built from the same request fields, and the Connect error code
// of a refused call. It only reads descriptors, documents and bytes; the
// suite (test/e2e, readapi_test.go) makes the calls. It is never released.
package apicontract

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Connect's streaming envelope (connectrpc.com/docs/protocol): a flags
// byte, a big-endian 32-bit length, then the message.
const (
	envelopeHeader = 5
	endStreamFlag  = 0x02
)

var (
	// ErrNoHTTPRule reports a method without a google.api.http annotation.
	ErrNoHTTPRule = errors.New("no google.api.http rule")
	// ErrUnsupportedRule reports an HTTP rule this package can't map: a
	// custom verb or a path variable that is not a plain top-level field.
	ErrUnsupportedRule = errors.New("unsupported google.api.http rule")
	// ErrUnknownField reports a request field the method's input lacks.
	ErrUnknownField = errors.New("no such request field")
	// ErrMissingPathField reports a path variable without a value.
	ErrMissingPathField = errors.New("path variable without a value")
	// ErrWrongMessage reports a message of another type than a method's input.
	ErrWrongMessage = errors.New("not the method's input message")
	// ErrMessageTooLong reports a message no envelope can carry.
	ErrMessageTooLong = errors.New("message too long for an envelope")
	// ErrNoConnectError reports a body that holds no Connect error.
	ErrNoConnectError = errors.New("no Connect error in the body")
)

// Operation is one HTTP operation: its method, its path template with
// OpenAPI's {jsonName} variables, and its operation ID.
type Operation struct {
	Method, Path, ID string
}

// Operations lists every operation doc describes, by path then method.
func Operations(doc *openapi3.T) []Operation {
	var ops []Operation
	for path, item := range doc.Paths.Map() {
		for method, op := range item.Operations() {
			ops = append(ops, Operation{Method: method, Path: path, ID: op.OperationID})
		}
	}
	sortOperations(ops)
	return ops
}

// sortOperations sorts ops by path, then method.
func sortOperations(ops []Operation) {
	slices.SortFunc(ops, func(a, b Operation) int {
		if c := strings.Compare(a.Path, b.Path); c != 0 {
			return c
		}
		return strings.Compare(a.Method, b.Method)
	})
}

// Route is one method of a protobuf service with its HTTP mapping.
type Route struct {
	// Name is the method's name, such as ListAccounts.
	Name string
	// Procedure is its Connect path, such as
	// /openfief.views.v1.ReadService/ListAccounts.
	Procedure string
	// Operation is its HTTP mapping as protoc-gen-connect-openapi writes it
	// into the OpenAPI document: path variables by their JSON names and
	// the method's full name, dot-separated, as the operation ID.
	Operation Operation
	// ReadOnly reports a unary method that declares no side effects.
	ReadOnly bool

	template []string
	input    protoreflect.MessageDescriptor
}

// Routes returns sd's methods in declaration order with their HTTP
// mappings. A method without a google.api.http rule, with a custom verb
// or with a path variable other than a plain top-level field of its input
// is an error.
func Routes(sd protoreflect.ServiceDescriptor) ([]*Route, error) {
	routes := make([]*Route, 0, sd.Methods().Len())
	for i := range sd.Methods().Len() {
		r, err := routeOf(sd.Methods().Get(i))
		if err != nil {
			return nil, err
		}
		routes = append(routes, r)
	}
	return routes, nil
}

// RouteOperations returns the operations of routes, by path then method.
func RouteOperations(routes []*Route) []Operation {
	ops := make([]Operation, 0, len(routes))
	for _, r := range routes {
		ops = append(ops, r.Operation)
	}
	sortOperations(ops)
	return ops
}

// Procedure returns m's Connect path: /<service full name>/<method>.
func Procedure(m protoreflect.MethodDescriptor) string {
	return "/" + string(m.Parent().FullName()) + "/" + string(m.Name())
}

// routeOf maps one method.
func routeOf(m protoreflect.MethodDescriptor) (*Route, error) {
	opts, ok := m.Options().(*descriptorpb.MethodOptions)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNoHTTPRule, m.FullName())
	}
	rule, ok := proto.GetExtension(opts, annotations.E_Http).(*annotations.HttpRule)
	if !ok || rule.GetPattern() == nil {
		return nil, fmt.Errorf("%w: %s", ErrNoHTTPRule, m.FullName())
	}
	p := httpPattern(rule)
	if p.method == "" {
		return nil, fmt.Errorf("%w: %s: custom verb", ErrUnsupportedRule, m.FullName())
	}
	tmpl, err := parseTemplate(p.path, m.Input())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", m.FullName(), err)
	}
	unary := !m.IsStreamingClient() && !m.IsStreamingServer()
	return &Route{
		Name:      string(m.Name()),
		Procedure: Procedure(m),
		Operation: Operation{Method: p.method, Path: tmpl.openAPIPath, ID: string(m.FullName())},
		ReadOnly:  unary && opts.GetIdempotencyLevel() == descriptorpb.MethodOptions_NO_SIDE_EFFECTS,
		template:  tmpl.segments,
		input:     m.Input(),
	}, nil
}

// pattern is an HTTP rule's method and path template.
type pattern struct {
	method, path string
}

// httpPattern returns rule's HTTP method and path template; the method is
// empty for a custom verb.
func httpPattern(rule *annotations.HttpRule) pattern {
	switch p := rule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		return pattern{http.MethodGet, p.Get}
	case *annotations.HttpRule_Put:
		return pattern{http.MethodPut, p.Put}
	case *annotations.HttpRule_Post:
		return pattern{http.MethodPost, p.Post}
	case *annotations.HttpRule_Delete:
		return pattern{http.MethodDelete, p.Delete}
	case *annotations.HttpRule_Patch:
		return pattern{http.MethodPatch, p.Patch}
	}
	return pattern{}
}

// template is a path template split at its variables: the even elements
// of segments are literal text, the odd ones the proto names of the
// fields the variables name; openAPIPath writes each variable by its
// field's JSON name, as the OpenAPI document does.
type template struct {
	segments    []string
	openAPIPath string
}

// parseTemplate splits path at its variables.
func parseTemplate(path string, input protoreflect.MessageDescriptor) (template, error) {
	var segments []string
	var b strings.Builder
	rest := path
	for {
		open := strings.IndexByte(rest, '{')
		if open < 0 {
			break
		}
		end := strings.IndexByte(rest[open:], '}')
		if end < 0 {
			return template{}, fmt.Errorf("%w: unclosed variable in %q", ErrUnsupportedRule, path)
		}
		name := rest[open+1 : open+end]
		field := input.Fields().ByName(protoreflect.Name(name))
		if field == nil || strings.ContainsAny(name, ".=*") {
			return template{}, fmt.Errorf("%w: variable {%s} in %q is not a field of %s", ErrUnsupportedRule, name, path, input.FullName())
		}
		segments = append(segments, rest[:open], name)
		b.WriteString(rest[:open] + "{" + field.JSONName() + "}")
		rest = rest[open+end+1:]
	}
	segments = append(segments, rest)
	b.WriteString(rest)
	return template{segments: segments, openAPIPath: b.String()}, nil
}

// Target returns the REST path and query of one call of r with fields,
// the request's fields by their JSON names: each path variable takes its
// field's value, path-escaped, and every other field becomes a query
// parameter of the same name, in name order.
func (r *Route) Target(fields map[string]string) (string, error) {
	if err := r.checkFields(fields); err != nil {
		return "", err
	}
	used := map[string]bool{}
	var b strings.Builder
	for i, s := range r.template {
		if i%2 == 0 {
			b.WriteString(s)
			continue
		}
		name := r.input.Fields().ByName(protoreflect.Name(s)).JSONName()
		v, ok := fields[name]
		if !ok || v == "" {
			return "", fmt.Errorf("%w: %s of %s", ErrMissingPathField, name, r.Name)
		}
		used[name] = true
		b.WriteString(url.PathEscape(v))
	}
	query := url.Values{}
	for name, v := range fields {
		if !used[name] {
			query.Set(name, v)
		}
	}
	if len(query) > 0 {
		b.WriteString("?" + query.Encode())
	}
	return b.String(), nil
}

// Fill sets fields on m, r's input message, as protojson reads them:
// integers and enum names as strings. A message of another type is an
// error.
func (r *Route) Fill(m proto.Message, fields map[string]string) error {
	if got := m.ProtoReflect().Descriptor().FullName(); got != r.input.FullName() {
		return fmt.Errorf("%w: %s, want %s", ErrWrongMessage, got, r.input.FullName())
	}
	if err := r.checkFields(fields); err != nil {
		return err
	}
	if fields == nil {
		fields = map[string]string{}
	}
	raw, err := json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("request fields of %s: %w", r.Name, err)
	}
	if err := protojson.Unmarshal(raw, m); err != nil {
		return fmt.Errorf("request fields of %s: %w", r.Name, err)
	}
	return nil
}

// checkFields refuses a field name r's input has no field for.
func (r *Route) checkFields(fields map[string]string) error {
	for name := range fields {
		if r.input.Fields().ByJSONName(name) == nil {
			return fmt.Errorf("%w: %s in %s", ErrUnknownField, name, r.input.FullName())
		}
	}
	return nil
}

// Envelope wraps msg in one Connect streaming envelope, as a client
// stream's only message. A message longer than an envelope can say is an
// error.
func Envelope(msg []byte) ([]byte, error) {
	n := len(msg)
	if n < 0 || n > math.MaxUint32 {
		return nil, fmt.Errorf("%w: %d bytes", ErrMessageTooLong, n)
	}
	b := make([]byte, envelopeHeader, envelopeHeader+len(msg))
	binary.BigEndian.PutUint32(b[1:], uint32(n))
	return append(b, msg...), nil
}

// ConnectErrorCode returns the Connect error code of a response: a unary
// call's JSON error body, or the error in a streaming response's
// end-of-stream message when contentType is a Connect streaming type
// (application/connect+...). A streaming response that ends without an
// error answers ""; a body that is neither is ErrNoConnectError.
func ConnectErrorCode(contentType string, body []byte) (string, error) {
	if !strings.HasPrefix(contentType, "application/connect+") {
		var e struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(body, &e); err != nil || e.Code == "" {
			return "", fmt.Errorf("%w: %q", ErrNoConnectError, excerpt(body))
		}
		return e.Code, nil
	}
	rest := body
	for len(rest) >= envelopeHeader {
		flags, n := rest[0], binary.BigEndian.Uint32(rest[1:envelopeHeader])
		if int64(len(rest)-envelopeHeader) < int64(n) {
			break
		}
		msg := rest[envelopeHeader : envelopeHeader+int(n)]
		rest = rest[envelopeHeader+int(n):]
		if flags&endStreamFlag == 0 {
			continue
		}
		var end struct {
			Error *struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(msg, &end); err != nil {
			return "", fmt.Errorf("%w: end of stream %q: %w", ErrNoConnectError, excerpt(msg), err)
		}
		if end.Error == nil {
			return "", nil
		}
		return end.Error.Code, nil
	}
	return "", fmt.Errorf("%w: no end of stream in %q", ErrNoConnectError, excerpt(body))
}

// excerpt is at most the first 200 bytes of b, for an error message.
func excerpt(b []byte) []byte {
	const limit = 200
	if len(b) > limit {
		return append(bytes.Clone(b[:limit]), "..."...)
	}
	return b
}
