package main

import (
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// contracts indexes the base's contracts: each file's package, each method's request and response, and
// the messages each message's fields hold, all by fully qualified name without the leading dot.
type contracts struct {
	packages map[string]string
	methods  map[string][]string
	fields   map[string][]string
}

// indexContracts indexes set, the base's contracts as buf build writes them.
func indexContracts(set *descriptorpb.FileDescriptorSet) *contracts {
	c := &contracts{packages: map[string]string{}, methods: map[string][]string{}, fields: map[string][]string{}}
	for _, f := range set.GetFile() {
		pkg := f.GetPackage()
		c.packages[f.GetName()] = pkg
		for _, s := range f.GetService() {
			for _, m := range s.GetMethod() {
				name := pkg + "." + s.GetName() + "." + m.GetName()
				c.methods[name] = []string{trimDot(m.GetInputType()), trimDot(m.GetOutputType())}
			}
		}
		for _, m := range f.GetMessageType() {
			c.indexMessage(pkg+"."+m.GetName(), m)
		}
	}
	return c
}

// indexMessage records the message types m's fields hold, and its nested messages'.
func (c *contracts) indexMessage(name string, m *descriptorpb.DescriptorProto) {
	c.fields[name] = []string{}
	for _, f := range m.GetField() {
		if f.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
			c.fields[name] = append(c.fields[name], trimDot(f.GetTypeName()))
		}
	}
	for _, n := range m.GetNestedType() {
		c.indexMessage(name+"."+n.GetName(), n)
	}
}

// trimDot drops the leading dot of a fully qualified type name.
func trimDot(name string) string { return strings.TrimPrefix(name, ".") }

// packageOf returns the package of the file at path, a path in the repository such as
// "proto/openfief/views/v1/read_service.proto", which buf names relative to its module.
func (c *contracts) packageOf(path string) (string, bool) {
	for name, pkg := range c.packages {
		if path == name || strings.HasSuffix(path, "/"+name) {
			return pkg, true
		}
	}
	return "", false
}

// hasMessage reports whether the base has the message name.
func (c *contracts) hasMessage(name string) bool {
	_, ok := c.fields[name]
	return ok
}

// reachable returns every message the methods hold: their requests and responses, and every message
// their fields hold, nested ones included.
func (c *contracts) reachable(methods []string) map[string]bool {
	out := map[string]bool{}
	var queue []string
	for _, m := range methods {
		queue = append(queue, c.methods[m]...)
	}
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if out[name] {
			continue
		}
		out[name] = true
		queue = append(queue, c.fields[name]...)
	}
	return out
}
