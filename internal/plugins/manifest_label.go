package plugins

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// ManifestLabel is the image config label that holds a plugin's manifest
// (D-42). The image's signature covers its config, so it covers the label.
const ManifestLabel = "io.openfief.plugin.manifest"

// manifestDoc is the label's JSON form; it has the same fields as the Plugin
// resource's spec.manifest.
type manifestDoc struct {
	Capabilities []Capability `json:"capabilities"`
	Egress       []string     `json:"egress,omitempty"`
	Secrets      []struct {
		Name      string `json:"name"`
		MountPath string `json:"mountPath"`
	} `json:"secrets,omitempty"`
	DataTiers    []DataTier      `json:"dataTiers,omitempty"`
	ConfigSchema json.RawMessage `json:"configSchema,omitempty"`
	BrowserPath  string          `json:"browserPath,omitempty"`
}

// ParseManifestLabel reads a manifest from the value of an image's
// ManifestLabel: a JSON object with the fields of spec.manifest
// (capabilities, egress, secrets, dataTiers, configSchema, browserPath). Unknown fields
// are refused, so a manifest is never half-read. The manifest is validated.
func ParseManifestLabel(value string) (Manifest, error) {
	var doc manifestDoc
	if err := decodeStrict([]byte(value), &doc); err != nil {
		return Manifest{}, fmt.Errorf("manifest label: %w", err)
	}
	m := Manifest{Capabilities: doc.Capabilities, DataTiers: doc.DataTiers, BrowserPath: doc.BrowserPath}
	if len(doc.ConfigSchema) > 0 && !bytes.Equal(doc.ConfigSchema, []byte("null")) {
		m.ConfigSchema = doc.ConfigSchema
	}
	var errs []error
	for i, s := range doc.Egress {
		h, err := ParseEgressHost(s)
		if err != nil {
			errs = append(errs, &FieldError{Field: indexed("egress", i), Err: err})
			continue
		}
		m.Egress = append(m.Egress, h)
	}
	for _, s := range doc.Secrets {
		m.Secrets = append(m.Secrets, SecretMount(s))
	}
	if err := errors.Join(errs...); err != nil {
		return Manifest{}, fmt.Errorf("manifest label: %w", err)
	}
	if err := m.Validate(); err != nil {
		return Manifest{}, fmt.Errorf("manifest label: %w", err)
	}
	return m, nil
}

// Equal reports whether m and o declare the same thing: the same sets of
// capabilities, egress hosts, secret mounts and data tiers, in any order, the
// same browser path, and config schemas that are the same JSON value.
func (m *Manifest) Equal(o *Manifest) bool {
	a, b := m.Requested(), o.Requested()
	return slices.Equal(a.Capabilities, b.Capabilities) &&
		slices.Equal(a.Egress, b.Egress) &&
		slices.Equal(a.DataTiers, b.DataTiers) &&
		slices.Equal(sortedMounts(m.Secrets), sortedMounts(o.Secrets)) &&
		m.BrowserPath == o.BrowserPath &&
		sameJSON(m.ConfigSchema, o.ConfigSchema)
}

// sortedMounts returns the mounts sorted by name, then mount path.
func sortedMounts(ms []SecretMount) []SecretMount {
	return sortedCopy(ms, func(a, b SecretMount) int {
		if c := strings.Compare(a.Name, b.Name); c != 0 {
			return c
		}
		return strings.Compare(a.MountPath, b.MountPath)
	})
}

// sameJSON reports whether two JSON documents hold the same value; an absent
// document equals only another absent one, and invalid JSON equals nothing.
func sameJSON(a, b []byte) bool {
	if len(a) == 0 || len(b) == 0 {
		return len(a) == len(b)
	}
	var va, vb any
	if decodeNumbers(a, &va) != nil || decodeNumbers(b, &vb) != nil {
		return false
	}
	return reflect.DeepEqual(va, vb)
}

// decodeNumbers decodes JSON keeping numbers exact, as json.Number.
func decodeNumbers(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}
