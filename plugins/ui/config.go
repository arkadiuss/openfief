package main

import (
	"errors"
	"fmt"
	"regexp"
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
	// TurnstileSiteKey and TurnstileVerifyURL turn Cloudflare Turnstile on in
	// front of the Ask box's model (D-208): the public key of the widget made for
	// the Ask box's host, and Cloudflare's siteverify URL, whose host the grant
	// approves for egress (egressFromConfig, D-75). Both or neither; the secret
	// key is the instance's Secret ui--turnstile-secret, never config.
	TurnstileSiteKey   string `json:"turnstileSiteKey"`
	TurnstileVerifyURL string `json:"turnstileVerifyUrl"`
}

// siteKeyPattern is what Cloudflare's site keys look like: a digit and an x, then
// letters, digits, underscores and dashes ("0x4AAAA..." for a real widget,
// "1x00000000000000000000AA" for the always-passing test key). A placeholder
// fails it, so an unset key never reaches a page.
var siteKeyPattern = regexp.MustCompile(`^[0-3]x[A-Za-z0-9_-]{8,64}$`)

// turnstileOn reports whether the config turns Turnstile on, and an error
// wrapping errBadConfig for one that sets only half of it or a site key that is
// not one.
func (r *rawConfig) turnstileOn() (bool, error) {
	switch {
	case r.TurnstileSiteKey == "" && r.TurnstileVerifyURL == "":
		return false, nil
	case r.TurnstileSiteKey == "" || r.TurnstileVerifyURL == "":
		return false, fmt.Errorf("%w: turnstileSiteKey and turnstileVerifyUrl are set together", errBadConfig)
	case !siteKeyPattern.MatchString(r.TurnstileSiteKey):
		return false, fmt.Errorf("%w: turnstileSiteKey is a Turnstile site key", errBadConfig)
	}
	return true, nil
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
