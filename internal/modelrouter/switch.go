package modelrouter

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

// ErrUnknownFeature reports a feature the switch has no toggle for.
var ErrUnknownFeature = errors.New("unknown feature")

// Feature is something that uses models or serves agents, with a toggle of
// its own under the switch (D-212).
type Feature string

// The features, a closed set (D-212).
const (
	// FeatureAsk is the Ask box's model (D-203).
	FeatureAsk Feature = "ask"
	// FeatureMaintainer is the model-portfolio maintainer (D-204).
	FeatureMaintainer Feature = "maintainer"
	// FeatureMCP is the MCP server for external agents (D-213).
	FeatureMCP Feature = "mcp"
)

// Features returns every feature, in the order Settings lists them.
func Features() []Feature {
	return []Feature{FeatureAsk, FeatureMaintainer, FeatureMCP}
}

// ParseFeature parses a feature as its constant spells it.
func ParseFeature(s string) (Feature, error) {
	f := Feature(s)
	if !slices.Contains(Features(), f) {
		return "", fmt.Errorf("%w: %q", ErrUnknownFeature, s)
	}
	return f, nil
}

// Switch is the one switch for agents and models and the toggle of each
// feature under it (D-212). The zero value is off, with every toggle off:
// agents and models are off by default.
type Switch struct {
	// Enabled is the one switch. Off, nothing calls a model and no MCP
	// endpoint answers, whatever the toggles say.
	Enabled bool
	// Toggles holds the features that are on. A feature absent is off.
	Toggles map[Feature]bool
}

// Allows reports whether f may run: the switch is on and f's toggle is on.
// An unknown feature is never allowed.
func (s Switch) Allows(f Feature) bool {
	return s.Enabled && s.Toggles[f] && slices.Contains(Features(), f)
}

// WithEnabled returns s with the switch set to on.
func (s Switch) WithEnabled(on bool) Switch {
	s.Toggles = cloneToggles(s.Toggles)
	s.Enabled = on
	return s
}

// WithToggle returns s with f's toggle set to on, or an error wrapping
// ErrUnknownFeature when f is no feature.
func (s Switch) WithToggle(f Feature, on bool) (Switch, error) {
	if !slices.Contains(Features(), f) {
		return Switch{}, fmt.Errorf("%w: %q", ErrUnknownFeature, string(f))
	}
	s.Toggles = cloneToggles(s.Toggles)
	if on {
		s.Toggles[f] = true
	} else {
		delete(s.Toggles, f)
	}
	return s, nil
}

// cloneToggles returns a copy of t that can be written to.
func cloneToggles(t map[Feature]bool) map[Feature]bool {
	out := make(map[Feature]bool, len(t)+1)
	maps.Copy(out, t)
	return out
}
