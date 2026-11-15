package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// demoRouteFile and demoProviderKind name what the demo route's config is made
// of in deploy/demo/model: settings.json holds the route of the Ask box, each
// entry "<provider instance>/<model id>", and the provider's Plugin resource
// lists the instance's model entries, with the hosts each is pinned to.
const (
	demoSettingsFile = "settings.json"
	demoProviderKind = "Plugin"
	demoRoleAsk      = "ask"
)

// demoSettings is the part of the demo's model settings the bench reads.
type demoSettings struct {
	Routes map[string][]string `json:"routes"`
}

// demoPlugin is the part of a Plugin resource the bench reads: its name and the
// model entries of its config.
type demoPlugin struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		Config struct {
			Models []demoEntry `yaml:"models"`
		} `yaml:"config"`
	} `yaml:"spec"`
}

// demoEntry is a model entry of the provider instance (D-205).
type demoEntry struct {
	ID               string   `yaml:"id"`
	InputPrice       string   `yaml:"inputPrice"`
	OutputPrice      string   `yaml:"outputPrice"`
	StructuredOutput bool     `yaml:"structuredOutput"`
	Reasoning        string   `yaml:"reasoning"`
	Hosts            []string `yaml:"hosts"`
}

// LoadDemoRoute reads the Ask box's route of the demo from dir (deploy/demo/model),
// read-only, and returns each member as the bench would ask it: the entry's
// model id, hosts, reasoning, structured-output flag and prices. Its label is
// the label the bench's report keys the run by, so a run of a member is found
// in the history and the gate by it.
func LoadDemoRoute(dir string) ([]Model, error) {
	raw, err := os.ReadFile(filepath.Clean(filepath.Join(dir, demoSettingsFile)))
	if err != nil {
		return nil, fmt.Errorf("%w: reading the demo route: %w", ErrUsage, err)
	}
	var s demoSettings
	if jsonErr := json.Unmarshal(raw, &s); jsonErr != nil {
		return nil, fmt.Errorf("%w: %s is no model settings: %w", ErrUsage, demoSettingsFile, jsonErr)
	}
	refs := s.Routes[demoRoleAsk]
	if len(refs) == 0 {
		return nil, fmt.Errorf("%w: %s has no route for %q", ErrUsage, demoSettingsFile, demoRoleAsk)
	}
	entries, err := demoEntries(dir)
	if err != nil {
		return nil, err
	}
	var out []Model
	for _, ref := range refs {
		m, err := memberOf(ref, entries)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

// memberOf returns the model a route entry, "<instance>/<model id>", names, as the
// bench asks it.
func memberOf(ref string, entries map[string]map[string]demoEntry) (Model, error) {
	instance, id, ok := strings.Cut(ref, "/")
	if !ok {
		return Model{}, fmt.Errorf("%w: the route entry %q is not <instance>/<model id>", ErrUsage, ref)
	}
	e, found := entries[instance][id]
	if !found {
		return Model{}, fmt.Errorf("%w: the route entry %q names no model entry of the instance %q", ErrUsage, ref, instance)
	}
	m := priced(e.ID, e.InputPrice, e.OutputPrice)
	if m.ID == "" || (m.InputPrice.Sign() == 0 && m.OutputPrice.Sign() == 0) {
		return Model{}, fmt.Errorf("%w: the model entry %q has no readable prices", ErrUsage, ref)
	}
	m.Route, m.Structured, m.Reasoning = e.Hosts, e.StructuredOutput, e.Reasoning
	return m, nil
}

// demoEntries returns the model entries of each provider instance (Plugin
// resource) in dir's YAML files, by instance name and model id.
func demoEntries(dir string) (map[string]map[string]demoEntry, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUsage, err)
	}
	slices.Sort(files)
	out := map[string]map[string]demoEntry{}
	for _, f := range files {
		if err := readProviders(f, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// readProviders adds the model entries of each provider resource in the YAML
// file to out, by instance name and model id.
func readProviders(file string, out map[string]map[string]demoEntry) error {
	raw, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return fmt.Errorf("%w: reading the demo route: %w", ErrUsage, err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	for {
		var p demoPlugin
		if err := dec.Decode(&p); errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return fmt.Errorf("%w: %s is no YAML: %w", ErrUsage, filepath.Base(file), err)
		}
		if p.Kind != demoProviderKind || len(p.Spec.Config.Models) == 0 {
			continue
		}
		byID := map[string]demoEntry{}
		for _, e := range p.Spec.Config.Models {
			byID[e.ID] = e
		}
		out[p.Metadata.Name] = byID
	}
}

// memberLabels are the labels of the route's models, in the route's order.
func memberLabels(models []Model) []string {
	out := make([]string, len(models))
	for i := range models {
		out[i] = modelLabel(&models[i])
	}
	return out
}
