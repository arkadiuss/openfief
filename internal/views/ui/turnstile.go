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
	// checkPathField is the hidden form field that says which way the check
	// completed, so the log shows how often the invisible one stalls.
	checkPathField = "ask-check-path"
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
	// turnstileNoToken: no usable token reached the UI: the widget had not
	// finished, did not load (JavaScript off, challenges.cloudflare.com
	// blocked), or sent one longer than Cloudflare issues.
	turnstileNoToken turnstileVerdict = iota
	// turnstileRejected: Cloudflare did not accept the token: expired, used
	// before, forged or made for another widget. Its error codes say which.
	turnstileRejected
	// turnstileWrongHost: Cloudflare accepted the token, but it was made on
	// another host than the one asking.
	turnstileWrongHost
	// turnstilePassed: Cloudflare accepted the token.
	turnstilePassed
	// turnstileUnavailable: the check could not be made: Cloudflare could not be
	// reached, answered nothing usable, or the UI has no secret key. It fails
	// closed: the question is not sent.
	turnstileUnavailable
)

// String is the verdict as a fixed word for the log.
func (v turnstileVerdict) String() string {
	switch v {
	case turnstileNoToken:
		return "no-token"
	case turnstileRejected:
		return "rejected"
	case turnstileWrongHost:
		return "wrong-host"
	case turnstilePassed:
		return "passed"
	case turnstileUnavailable:
		return "unavailable"
	}
	return "unknown"
}

// The ways the check can complete, as the page's script names them in
// checkPathField: "invisible" is the widget's own interaction-only run, and
// "visible-fallback" the visible widget the script renders when that one stalls.
const (
	checkPathInvisible = "invisible"
	checkPathVisible   = "visible-fallback"
	checkPathUnknown   = "unknown"
)

// checkPath returns raw when it is one of the two fixed words, and "unknown" for
// anything else, including no field: typed data from the browser is never kept,
// logged or shown.
func checkPath(raw string) string {
	switch raw {
	case checkPathInvisible, checkPathVisible:
		return raw
	}
	return checkPathUnknown
}

// turnstileOutcome is a verdict with what the visitor may be told of it. The
// detail is only ever Cloudflare's fixed error codes and two host names: no
// token, no secret, no question.
type turnstileOutcome struct {
	verdict turnstileVerdict
	// codes are Cloudflare's error codes for turnstileRejected, cleaned by
	// safeCodes.
	codes []string
	// reported is the host Cloudflare says the token was made on, and asked the
	// host that is asking, both for turnstileWrongHost.
	reported, asked string
}

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
// along with the visitor's address when the core knew it. Only the verdict and
// what may be shown of it come back; nothing of the exchange but Cloudflare's error codes, which are
// fixed words, is logged, and never a token, the secret or a question.
func (t *turnstileCheck) verify(ctx context.Context, token, remoteIP, host string) turnstileOutcome {
	if token == "" || len(token) > maxTurnstileToken {
		return turnstileOutcome{verdict: turnstileNoToken}
	}
	reply, ok := t.ask(ctx, token, remoteIP)
	if !ok {
		return turnstileOutcome{verdict: turnstileUnavailable}
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
func (t *turnstileCheck) judge(ctx context.Context, reply *siteverifyReply, host string) turnstileOutcome {
	if !reply.Success {
		for _, code := range reply.ErrorCodes {
			if code == "missing-input-secret" || code == "invalid-input-secret" {
				t.log.WarnContext(ctx, "turnstile refused the secret key", "codes", strings.Join(reply.ErrorCodes, ","))
				return turnstileOutcome{verdict: turnstileUnavailable}
			}
		}
		codes := safeCodes(reply.ErrorCodes)
		t.log.InfoContext(ctx, "turnstile refused a token", "codes", strings.Join(codes, ","))
		return turnstileOutcome{verdict: turnstileRejected, codes: codes}
	}
	// A token made on another site with the same key would pass Cloudflare's
	// check; it must have been made on the host that is asking.
	if reply.Hostname != "" && !strings.EqualFold(reply.Hostname, hostOnly(host)) {
		out := turnstileOutcome{verdict: turnstileWrongHost, reported: safeHost(reply.Hostname), asked: safeHost(hostOnly(host))}
		t.log.InfoContext(ctx, "turnstile token was made for another host", "reported", out.reported, "asked", out.asked)
		return out
	}
	return turnstileOutcome{verdict: turnstilePassed}
}

const (
	// maxShownCodes and maxCodeLen bound the error codes shown to a visitor; a
	// real one is a short fixed word such as timeout-or-duplicate.
	maxShownCodes = 8
	maxCodeLen    = 48
	// maxHostLen is the longest DNS name.
	maxHostLen = 253
)

// safeCodes keeps the error codes that look like Cloudflare's fixed words
// (lower-case letters, digits and hyphens), so that nothing else a reply holds
// reaches a page or a log.
func safeCodes(codes []string) []string {
	var kept []string
	for _, c := range codes {
		if len(kept) == maxShownCodes {
			break
		}
		if c != "" && len(c) <= maxCodeLen && strings.Trim(c, "abcdefghijklmnopqrstuvwxyz0123456789-") == "" {
			kept = append(kept, c)
		}
	}
	return kept
}

// safeHost returns host cut to a DNS name's length; a host name is no secret,
// and the template escapes it.
func safeHost(host string) string {
	if len(host) > maxHostLen {
		return host[:maxHostLen]
	}
	return host
}

// hostOnly returns host without its port.
func hostOnly(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}

// checkNeededAnswer is the answer to a question that came with no token: the
// widget had not finished or did not load. The check runs in the browser, so
// nothing was sent to Cloudflare.
func checkNeededAnswer(question string) answerCard {
	return modelCard(question, "A check comes before the model answers",
		"No result of the check reached OpenFief with your question: it hadn't finished, or it didn't load. "+
			"Wait until the line under the Ask field says it is ready, then ask again, or reload the page. "+
			"The check needs JavaScript and a connection to challenges.cloudflare.com, which some browsers or blockers stop. "+
			"If your browser blocked it, allow challenges.cloudflare.com for this site (in Brave: turn Shields off for it), then reload. "+
			"The commands answer without it.")
}

// checkRefusedAnswer is the answer to a question whose token Cloudflare refused,
// with Cloudflare's error codes: fixed words such as timeout-or-duplicate.
func checkRefusedAnswer(question string, codes []string) answerCard {
	shown := "it gave no error code"
	if len(codes) > 0 {
		shown = "error codes: " + strings.Join(codes, ", ")
	}
	return modelCard(question, "Cloudflare refused the check",
		"Cloudflare didn't accept the result of the check ("+shown+"). "+
			"A new check is being made; ask again in a moment. "+
			"The commands answer without it.")
}

// checkWrongHostAnswer is the answer to a question whose token Cloudflare
// accepted for another host than the one asking. Host names are no secret.
func checkWrongHostAnswer(question, reported, asked string) answerCard {
	return modelCard(question, "The check was made for another host",
		"Cloudflare says the check was made for "+reported+", but the question came from "+asked+". "+
			"Reload the page from the address you are using, then ask again. "+
			"The commands answer without it.")
}

// checkUnavailableAnswer is the answer when the check could not be made.
func checkUnavailableAnswer(question string) answerCard {
	return modelCard(question, "The check couldn't run",
		"The box couldn't reach the service that checks a person is asking, so nothing was sent to the model. "+
			"Try again in a moment; the commands answer from your own record.")
}

// turnstileAnswer returns the card of an outcome that stops the question, or
// false for one that lets it go on.
func turnstileAnswer(question string, o turnstileOutcome) (answerCard, bool) {
	switch o.verdict {
	case turnstilePassed:
		return answerCard{}, false
	case turnstileUnavailable:
		return checkUnavailableAnswer(question), true
	case turnstileRejected:
		return checkRefusedAnswer(question, o.codes), true
	case turnstileWrongHost:
		return checkWrongHostAnswer(question, o.reported, o.asked), true
	case turnstileNoToken:
	}
	return checkNeededAnswer(question), true
}
