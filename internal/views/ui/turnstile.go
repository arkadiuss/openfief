package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// The Turnstile check in front of the Ask box's model (D-208). Cloudflare's
// widget runs in the visitor's browser and puts a one-time token in the Ask
// form; before a question goes to a model the UI hands the token to Cloudflare's
// siteverify and asks again only with a fresh one. It is configured, not built
// in: without a TurnstileConfig nothing of it exists on the page.
const (
	// turnstileField is the form field the widget puts its token in.
	turnstileField = "cf-turnstile-response"
	// maxTurnstileToken is the longest token Cloudflare issues.
	maxTurnstileToken = 2048
	// turnstileTimeout bounds the call to Cloudflare: the visitor waits for it.
	turnstileTimeout = 5 * time.Second
	// maxTurnstileReply bounds what is read of Cloudflare's answer.
	maxTurnstileReply = 16 << 10
	// turnstileScript and turnstileOrigin are where the widget's script comes from
	// and the one origin the page's policy then allows beyond its own.
	turnstileScript = "https://challenges.cloudflare.com/turnstile/v0/api.js"
	turnstileOrigin = "https://challenges.cloudflare.com"
)

// ErrBadTurnstile reports a TurnstileConfig that is incomplete or unusable.
var ErrBadTurnstile = errors.New("invalid Turnstile config")

// TurnstileConfig turns the check on. The UI shows the widget on the Ask box
// when a model answers it, and a question that goes to a model passes only with a
// token Cloudflare accepts.
type TurnstileConfig struct {
	// SiteKey is the public key of the widget Cloudflare made for the Ask box's
	// host; it goes into the page.
	SiteKey string
	// VerifyURL is Cloudflare's siteverify endpoint, an https URL whose host the
	// plugin's grant approves for egress (D-75).
	VerifyURL string
	// Secret returns the widget's secret key, read on every call from the
	// instance's Secret so that a rotated key works at once. It is never logged
	// or shown.
	Secret func() (string, error)
	// HTTP calls VerifyURL: the plugin's egress client, which reaches only the
	// granted host.
	HTTP *http.Client
}

// validate returns nil, or an error wrapping ErrBadTurnstile for a config without
// a site key, a secret or a client, or whose endpoint is not an https URL.
func (c *TurnstileConfig) validate() error {
	u, err := url.Parse(c.VerifyURL)
	switch {
	case c.SiteKey == "" || c.Secret == nil || c.HTTP == nil:
		return fmt.Errorf("%w: a site key, a secret and a client", ErrBadTurnstile)
	case err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil:
		return fmt.Errorf("%w: the verify URL is an https URL", ErrBadTurnstile)
	}
	return nil
}

// turnstileVerdict is what the check made of a question's token.
type turnstileVerdict int

// The verdicts. The zero one passes nothing.
const (
	// turnstileRefused: no token, or one Cloudflare did not accept: expired, used
	// before, made for another host or for another widget.
	turnstileRefused turnstileVerdict = iota
	// turnstilePassed: Cloudflare accepted the token.
	turnstilePassed
	// turnstileUnavailable: the check could not be made: Cloudflare could not be
	// reached, answered nothing usable, or the UI has no secret key. It fails
	// closed: the question is not sent.
	turnstileUnavailable
)

// turnstileCheck verifies tokens with Cloudflare for the Ask box.
type turnstileCheck struct {
	cfg TurnstileConfig
	log *slog.Logger
}

// siteverifyReply is the part of Cloudflare's answer the check reads.
type siteverifyReply struct {
	Success    bool
	Hostname   string
	ErrorCodes []string
}

// decodeSiteverify reads the members of Cloudflare's JSON answer the check needs,
// by key; one that is missing or of another type is left at its zero value, which
// refuses the token.
func decodeSiteverify(body []byte) (siteverifyReply, error) {
	var members map[string]json.RawMessage
	if err := json.Unmarshal(body, &members); err != nil {
		return siteverifyReply{}, fmt.Errorf("decoding Cloudflare's answer: %w", err)
	}
	var reply siteverifyReply
	for key, into := range map[string]any{"success": &reply.Success, "hostname": &reply.Hostname, "error-codes": &reply.ErrorCodes} {
		if raw, ok := members[key]; ok {
			if err := json.Unmarshal(raw, into); err != nil {
				return siteverifyReply{}, fmt.Errorf("decoding Cloudflare's %s: %w", key, err)
			}
		}
	}
	return reply, nil
}

// verify asks Cloudflare whether token is one it issued for host, and passes it
// along with the visitor's address when the core knew it. Only the verdict
// comes back; nothing of the exchange but Cloudflare's error codes, which are
// fixed words, is logged, and never a token, the secret or a question.
func (t *turnstileCheck) verify(ctx context.Context, token, remoteIP, host string) turnstileVerdict {
	if token == "" || len(token) > maxTurnstileToken {
		return turnstileRefused
	}
	reply, ok := t.ask(ctx, token, remoteIP)
	if !ok {
		return turnstileUnavailable
	}
	return t.judge(ctx, &reply, host)
}

// ask posts the secret key, the token and the address to Cloudflare's siteverify
// and returns its answer, or false when the check could not be made: no secret,
// no answer or one that is no JSON.
func (t *turnstileCheck) ask(ctx context.Context, token, remoteIP string) (siteverifyReply, bool) {
	secret, err := t.cfg.Secret()
	if err != nil || secret == "" {
		t.log.WarnContext(ctx, "turnstile has no secret key")
		return siteverifyReply{}, false
	}
	form := url.Values{"secret": {secret}, "response": {token}}
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}
	ctx, cancel := context.WithTimeout(ctx, turnstileTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.cfg.VerifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return siteverifyReply{}, false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := t.cfg.HTTP.Do(req)
	if err != nil {
		t.log.WarnContext(ctx, "turnstile verification could not be made")
		return siteverifyReply{}, false
	}
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, maxTurnstileReply))
	closeErr := resp.Body.Close()
	reply, decodeErr := decodeSiteverify(raw)
	if readErr != nil || closeErr != nil || decodeErr != nil {
		t.log.WarnContext(ctx, "turnstile verification answered nothing usable", "status", resp.StatusCode)
		return siteverifyReply{}, false
	}
	return reply, true
}

// judge turns Cloudflare's answer into a verdict. A token that fails is the
// visitor's to retry; a secret Cloudflare doesn't accept is the installation's,
// and fails closed as unavailable.
func (t *turnstileCheck) judge(ctx context.Context, reply *siteverifyReply, host string) turnstileVerdict {
	if !reply.Success {
		for _, code := range reply.ErrorCodes {
			if code == "missing-input-secret" || code == "invalid-input-secret" {
				t.log.WarnContext(ctx, "turnstile refused the secret key", "codes", strings.Join(reply.ErrorCodes, ","))
				return turnstileUnavailable
			}
		}
		t.log.InfoContext(ctx, "turnstile refused a token", "codes", strings.Join(reply.ErrorCodes, ","))
		return turnstileRefused
	}
	// A token made on another site with the same key would pass Cloudflare's
	// check; it must have been made on the host that is asking.
	if reply.Hostname != "" && !strings.EqualFold(reply.Hostname, hostOnly(host)) {
		t.log.InfoContext(ctx, "turnstile token was made for another host")
		return turnstileRefused
	}
	return turnstilePassed
}

// hostOnly returns host without its port.
func hostOnly(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}

// checkNeededAnswer is the answer to a question whose token was missing or not
// accepted: the model answers only after the check, which runs in the browser.
func checkNeededAnswer(question string) answerCard {
	return modelCard(question, "A check comes before the model answers",
		"The model answers only after a quick check that a person is asking, and it didn't pass or didn't run. "+
			"If you see no check next to the Ask field, reload the page (the check needs JavaScript and a connection to Cloudflare), then ask again. "+
			"The commands answer without it.")
}

// checkUnavailableAnswer is the answer when the check could not be made.
func checkUnavailableAnswer(question string) answerCard {
	return modelCard(question, "The check couldn't run",
		"The box couldn't reach the service that checks a person is asking, so nothing was sent to the model. "+
			"Try again in a moment; the commands answer from your own record.")
}

// turnstileAnswer returns the card of a verdict that stops the question, or
// false for one that lets it go on.
func turnstileAnswer(question string, v turnstileVerdict) (answerCard, bool) {
	switch v {
	case turnstilePassed:
		return answerCard{}, false
	case turnstileUnavailable:
		return checkUnavailableAnswer(question), true
	case turnstileRefused:
	}
	return checkNeededAnswer(question), true
}
