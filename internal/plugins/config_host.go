package plugins

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrConfigHost reports a config field a manifest names in EgressFromConfig
// whose value is not an https URL of a single hostname (D-75).
var ErrConfigHost = errors.New("not an https URL of one hostname: no wildcard, no IP address, no user")

// maxConfigFieldLen bounds the name of a config field in EgressFromConfig.
const maxConfigFieldLen = 64

// ConfigHost is an egress host that comes from the instance's config instead
// of the manifest: the host of the https URL in config field Field (D-75).
type ConfigHost struct {
	Field string
	Host  EgressHost
}

// ConfigHosts reads, for each config field m.EgressFromConfig names, the host
// of the https URL the config holds there, in the manifest's order. A field
// that is absent, null or empty asks for no host. A config that isn't a JSON
// object, or a field that is not a string or not an https URL of one
// hostname, fails with a *FieldError naming "config.<field>" and wrapping
// ErrConfigHost; IP addresses and wildcards are refused as in EgressHost.
func ConfigHosts(m *Manifest, config []byte) ([]ConfigHost, error) {
	if len(m.EgressFromConfig) == 0 {
		return nil, nil
	}
	fields := map[string]json.RawMessage{}
	if trimmed := bytes.TrimSpace(config); len(trimmed) > 0 && !bytes.Equal(trimmed, []byte("null")) {
		if err := json.Unmarshal(trimmed, &fields); err != nil {
			return nil, &FieldError{Field: "config", Err: fmt.Errorf("%w: %s", ErrInvalidJSON, err.Error())}
		}
	}
	var out []ConfigHost
	var errs []error
	for _, f := range m.EgressFromConfig {
		h, ok, err := configHost(f, fields[f])
		switch {
		case err != nil:
			errs = append(errs, err)
		case ok:
			out = append(out, ConfigHost{Field: f, Host: h})
		}
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return out, nil
}

// configHost reads the host of config field f, whose JSON value is raw, and
// whether it gives one: an absent, null or empty field gives none.
func configHost(f string, raw json.RawMessage) (EgressHost, bool, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return EgressHost{}, false, nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return EgressHost{}, false, fieldErr("config."+f, ErrConfigHost, string(raw))
	}
	if s == "" {
		return EgressHost{}, false, nil
	}
	h, err := hostOfURL(s)
	if err != nil {
		return EgressHost{}, false, &FieldError{Field: "config." + f, Err: err}
	}
	return h, true, nil
}

// hostOfURL returns the host and port of an https URL, such as
// https://ghostfolio.example.com:8443/app: port 443 when none is written.
// The URL may not name a user, and its host must pass EgressHost's rules
// once lower-cased, so an IP address, in brackets or not, and a wildcard
// are refused.
func hostOfURL(raw string) (EgressHost, error) {
	const scheme = "https://"
	if len(raw) < len(scheme) || !strings.EqualFold(raw[:len(scheme)], scheme) {
		return EgressHost{}, fmt.Errorf("%w: %q", ErrConfigHost, raw)
	}
	authority := raw[len(scheme):]
	if i := strings.IndexAny(authority, "/?#"); i >= 0 {
		authority = authority[:i]
	}
	if authority == "" || strings.Contains(authority, "@") {
		return EgressHost{}, fmt.Errorf("%w: %q", ErrConfigHost, raw)
	}
	h, ok := splitAuthority(authority)
	if !ok {
		return EgressHost{}, fmt.Errorf("%w: %q: %w", ErrConfigHost, raw, ErrInvalidPort)
	}
	if err := h.validate(); err != nil {
		return EgressHost{}, fmt.Errorf("%w: %q: %w", ErrConfigHost, raw, err)
	}
	return h, nil
}

// splitAuthority splits a URL's host[:port], lower-casing the host; port
// 443 when none is written. An IPv6 literal stays whole, for validate to
// refuse. It reports false for a port that isn't 1 to 65535.
func splitAuthority(authority string) (EgressHost, bool) {
	i := strings.LastIndexByte(authority, ':')
	if i < 0 || strings.Contains(authority, "]") {
		return EgressHost{Host: strings.ToLower(authority), Port: DefaultEgressPort}, true
	}
	n, err := strconv.ParseUint(authority[i+1:], 10, 16)
	if err != nil || n == 0 {
		return EgressHost{}, false
	}
	return EgressHost{Host: strings.ToLower(authority[:i]), Port: uint16(n)}, true
}

// validateConfigFields checks the config field names of EgressFromConfig:
// each a letter followed by letters, digits, '_' or '-', at most 64 bytes,
// none twice.
func validateConfigFields(fields []string) []error {
	var errs []error
	for i, f := range fields {
		if !validConfigField(f) {
			errs = append(errs, fieldErr(indexed("egressFromConfig", i), ErrInvalidName, f))
		}
	}
	return append(errs, duplicates("egressFromConfig", fields, func(s string) string { return s })...)
}

// validConfigField reports whether f names a top-level config field.
func validConfigField(f string) bool {
	if f == "" || len(f) > maxConfigFieldLen {
		return false
	}
	for i, r := range f {
		if !configFieldRune(i == 0, r) {
			return false
		}
	}
	return true
}

// configFieldRune reports whether r may stand in a config field name: a
// letter anywhere, a digit, '_' or '-' after the first.
func configFieldRune(first bool, r rune) bool {
	letter := ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
	return letter || (!first && (isDigit(r) || r == '_' || r == '-'))
}
