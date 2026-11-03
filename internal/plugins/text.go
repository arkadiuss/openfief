package plugins

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"
)

// ErrUnsafeText reports a manifest string holding a character that can
// change what a terminal shows around it: a C0 or C1 control character,
// such as the escape that starts a cursor movement, or a Unicode
// bidirectional embedding, override or isolate. The CLI shows the owner a
// manifest's text to confirm an install or upgrade (D-46, D-90), so that
// text must read as it is written.
var ErrUnsafeText = errors.New("control or bidirectional formatting character")

// Unicode's bidirectional embeddings and overrides, U+202A to U+202E, and
// isolates, U+2066 to U+2069.
const (
	firstBidiOverride = '\u202a'
	lastBidiOverride  = '\u202e'
	firstBidiIsolate  = '\u2066'
	lastBidiIsolate   = '\u2069'
	// del is DEL, U+007F, and lastC1 the last C1 control character: DEL
	// and the C1 controls are one run.
	del    = '\u007f'
	lastC1 = '\u009f'
)

// UnsafeRune reports whether r is a C0 control character, DEL, a C1
// control character, or a bidirectional embedding, override or isolate.
func UnsafeRune(r rune) bool {
	return r < ' ' || (r >= del && r <= lastC1) ||
		(r >= firstBidiOverride && r <= lastBidiOverride) || (r >= firstBidiIsolate && r <= lastBidiIsolate)
}

// SafeText reports whether s is valid UTF-8 with no unsafe rune.
func SafeText(s string) bool {
	return utf8.ValidString(s) && !strings.ContainsFunc(s, UnsafeRune)
}

// textField is one string of a manifest and the field it is in.
type textField struct {
	field, value string
}

// validateManifestText refuses every string of m that isn't SafeText,
// the config schema's included: its property names, titles,
// descriptions and enum values.
func validateManifestText(m *Manifest) []error {
	var errs []error
	for _, f := range manifestText(m) {
		if !SafeText(f.value) {
			errs = append(errs, fieldErr(f.field, ErrUnsafeText, f.value))
		}
	}
	if s, ok := unsafeSchemaText(m.ConfigSchema); !ok {
		errs = append(errs, fieldErr("configSchema", ErrUnsafeText, s))
	}
	return errs
}

// manifestText lists m's strings other than its config schema.
func manifestText(m *Manifest) []textField {
	out := []textField{{"browserPath", m.BrowserPath}}
	if m.ClusterService != nil {
		out = append(out, textField{"clusterService.description", m.ClusterService.Description})
	}
	for i, c := range m.Capabilities {
		out = append(out, textField{indexed("capabilities", i), string(c)})
	}
	for i, h := range m.Egress {
		out = append(out, textField{indexed("egress", i), h.String()})
	}
	for i, s := range m.Secrets {
		out = append(out, textField{indexed("secrets", i) + ".role", s.Role}, textField{indexed("secrets", i) + ".mountPath", s.MountPath})
	}
	for i, d := range m.DataTiers {
		out = append(out, textField{indexed("dataTiers", i), string(d)})
	}
	for i, f := range m.EgressFromConfig {
		out = append(out, textField{indexed("egressFromConfig", i), f})
	}
	for i, in := range m.Institutions {
		out = append(out, textField{indexed("institutions", i), in})
	}
	return out
}

// unsafeSchemaText returns the first string of a config schema, a key or
// a value at any depth, that isn't SafeText, and false; or true when there
// is none. Invalid UTF-8 anywhere is unsafe too: the decoder would replace
// it, so it is checked on the bytes. A schema that isn't JSON is left to
// Validate's own check.
func unsafeSchemaText(schema []byte) (string, bool) {
	if len(schema) == 0 {
		return "", true
	}
	if !utf8.Valid(schema) {
		return string(schema), false
	}
	d := json.NewDecoder(bytes.NewReader(schema))
	d.UseNumber()
	var doc any
	if d.Decode(&doc) != nil {
		return "", true
	}
	return unsafeJSONText(doc)
}

// unsafeJSONText is unsafeSchemaText over a decoded JSON value.
func unsafeJSONText(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, SafeText(t)
	case []any:
		for _, e := range t {
			if s, ok := unsafeJSONText(e); !ok {
				return s, false
			}
		}
	case map[string]any:
		for _, k := range slices.Sorted(maps.Keys(t)) {
			if !SafeText(k) {
				return k, false
			}
			if s, ok := unsafeJSONText(t[k]); !ok {
				return s, false
			}
		}
	}
	return "", true
}
