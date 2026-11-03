package plugins

import (
	"bytes"
	"cmp"
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
		Role      string `json:"role"`
		MountPath string `json:"mountPath"`
	} `json:"secrets,omitempty"`
	DataTiers        []DataTier      `json:"dataTiers,omitempty"`
	ConfigSchema     json.RawMessage `json:"configSchema,omitempty"`
	BrowserPath      string          `json:"browserPath,omitempty"`
	EgressFromConfig []string        `json:"egressFromConfig,omitempty"`
	ClusterService   *struct {
		Description string `json:"description"`
	} `json:"clusterService,omitempty"`
	CrossCheckPrices bool     `json:"crossCheckPrices,omitempty"`
	StatementFiles   bool     `json:"statementFiles,omitempty"`
	MarketData       bool     `json:"marketData,omitempty"`
	Institutions     []string `json:"institutions,omitempty"`
	BondTerms        bool     `json:"bondTerms,omitempty"`
	IndexLevels      bool     `json:"indexLevels,omitempty"`
	IndexReadings    bool     `json:"indexReadings,omitempty"`
}

// ParseManifestLabel reads a manifest from the value of an image's
// ManifestLabel: a JSON object with the fields of spec.manifest
// (capabilities, egress, secrets, dataTiers, configSchema, browserPath,
// egressFromConfig, clusterService, crossCheckPrices, statementFiles,
// marketData, institutions, bondTerms, indexLevels, indexReadings).
// Unknown fields are refused, so a
// manifest is never half-read. The manifest is validated.
func ParseManifestLabel(value string) (Manifest, error) {
	var doc manifestDoc
	if err := decodeStrict([]byte(value), &doc); err != nil {
		return Manifest{}, fmt.Errorf("manifest label: %w", err)
	}
	m := Manifest{
		Capabilities: doc.Capabilities, DataTiers: doc.DataTiers, BrowserPath: doc.BrowserPath,
		EgressFromConfig: doc.EgressFromConfig, CrossCheckPrices: doc.CrossCheckPrices, StatementFiles: doc.StatementFiles,
		MarketData: doc.MarketData, Institutions: doc.Institutions, BondTerms: doc.BondTerms, IndexLevels: doc.IndexLevels,
		IndexReadings: doc.IndexReadings,
	}
	if doc.ClusterService != nil {
		m.ClusterService = &ClusterServiceRequest{Description: doc.ClusterService.Description}
	}
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
// capabilities, egress hosts, secret mounts, data tiers and config fields
// giving egress hosts, in any order, the same browser path, in-cluster
// service request and connector declarations, and config schemas that are
// the same JSON value.
func (m *Manifest) Equal(o *Manifest) bool {
	return slices.Equal(sortedCopy(m.Capabilities, cmp.Compare[Capability]), sortedCopy(o.Capabilities, cmp.Compare[Capability])) &&
		slices.Equal(sortedCopy(m.Egress, compareEgressHosts), sortedCopy(o.Egress, compareEgressHosts)) &&
		slices.Equal(sortedCopy(m.DataTiers, cmp.Compare[DataTier]), sortedCopy(o.DataTiers, cmp.Compare[DataTier])) &&
		slices.Equal(sortedMounts(m.Secrets), sortedMounts(o.Secrets)) &&
		m.BrowserPath == o.BrowserPath &&
		slices.Equal(sortedCopy(m.EgressFromConfig, strings.Compare), sortedCopy(o.EgressFromConfig, strings.Compare)) &&
		equalPtr(m.ClusterService, o.ClusterService) &&
		m.sameDeclarations(o) && sameJSON(m.ConfigSchema, o.ConfigSchema)
}

// sameDeclarations reports whether m and o make the same connector
// declarations (D-112, D-109, D-123, D-179, D-186) and declare the same
// institutions, in any order (D-182).
func (m *Manifest) sameDeclarations(o *Manifest) bool {
	return m.CrossCheckPrices == o.CrossCheckPrices && m.StatementFiles == o.StatementFiles && m.MarketData == o.MarketData &&
		m.BondTerms == o.BondTerms && m.IndexLevels == o.IndexLevels && m.IndexReadings == o.IndexReadings &&
		slices.Equal(sortedCopy(m.Institutions, strings.Compare), sortedCopy(o.Institutions, strings.Compare))
}

// sortedMounts returns the mounts sorted by role, then mount path.
func sortedMounts(ms []SecretMount) []SecretMount {
	return sortedCopy(ms, func(a, b SecretMount) int {
		if c := strings.Compare(a.Role, b.Role); c != 0 {
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
