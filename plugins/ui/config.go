package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/views/ui"
)

// errBadConfig reports a config the manifest's schema would refuse. It names the
// field and never a value.
var errBadConfig = errors.New("invalid config")

// The most the config lets one question take: well above D-208's defaults, and
// below what would let a question run a tool loop without end.
const (
	maxToolRounds   = 10
	maxToolCalls    = 50
	maxOutputTokens = 4000
	minDeadline     = 10 * time.Second
	maxDeadline     = 5 * time.Minute
)

// rawConfig is the plugin's configuration as OPENFIEF_CONFIG holds it. Every
// field is optional: a zero or missing one keeps D-208's default for the Ask
// box, and a plugin with no config at all runs as it always did.
type rawConfig struct {
	// AskToolRounds, AskToolCalls, AskOutputTokens and AskDeadline are the limits
	// of one question the Ask box sends a model (D-208, ui.AskLimits): rounds that
	// may call tools, tool calls in all, output tokens in all, and a duration such
	// as "90s" for the whole question.
	AskToolRounds   int    `json:"askToolRounds"`
	AskToolCalls    int    `json:"askToolCalls"`
	AskOutputTokens int64  `json:"askOutputTokens"`
	AskDeadline     string `json:"askDeadline"`
}

// askLimits returns the Ask box's limits the config sets, each zero field left for
// ui.AskLimits to default, or an error wrapping errBadConfig for a value out of
// range: negative, above the most a question may be given, or a deadline that is
// no duration between ten seconds and five minutes.
func (r *rawConfig) askLimits() (ui.AskLimits, error) {
	for _, c := range []struct {
		name     string
		got, max int64
	}{
		{"askToolRounds", int64(r.AskToolRounds), maxToolRounds},
		{"askToolCalls", int64(r.AskToolCalls), maxToolCalls},
		{"askOutputTokens", r.AskOutputTokens, maxOutputTokens},
	} {
		if c.got < 0 || c.got > c.max {
			return ui.AskLimits{}, fmt.Errorf("%w: %s is 0 to %d", errBadConfig, c.name, c.max)
		}
	}
	deadline, err := r.deadline()
	if err != nil {
		return ui.AskLimits{}, err
	}
	return ui.AskLimits{ToolRounds: r.AskToolRounds, ToolCalls: r.AskToolCalls, OutputTokens: r.AskOutputTokens, Deadline: deadline}, nil
}

// deadline returns the configured deadline, zero when none is set.
func (r *rawConfig) deadline() (time.Duration, error) {
	if r.AskDeadline == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(r.AskDeadline)
	if err != nil || d < minDeadline || d > maxDeadline {
		return 0, fmt.Errorf("%w: askDeadline is a duration from %s to %s", errBadConfig, minDeadline, maxDeadline)
	}
	return d, nil
}
