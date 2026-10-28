package plugins

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// DefaultEgressPort is the port of an egress host written without one.
	DefaultEgressPort uint16 = 443

	maxHostnameLen = 253
	maxLabelLen    = 63
	minLabels      = 2
	idnaReservedAt = 2 // labels with "--" at index 2 are reserved, except A-labels (xn--)
)

// EgressHost is one destination a plugin may reach through the egress proxy:
// a DNS hostname and a TCP port.
//
// The hostname is canonical: lower-case ASCII letters, digits and hyphens in
// at least two dot-separated labels, with no trailing dot. Internationalized
// names are written in their IDNA A-label form (xn--...). IP literals and
// wildcards are refused: an allow-list entry names exactly one host, and the
// proxy, not the plugin, resolves it (D-32).
type EgressHost struct {
	Host string
	Port uint16
}

// ParseEgressHost reads "host" or "host:port"; a missing port means
// DefaultEgressPort.
func ParseEgressHost(s string) (EgressHost, error) {
	host, port := s, DefaultEgressPort
	if i := strings.LastIndexByte(s, ':'); i >= 0 {
		if strings.Count(s, ":") > 1 || strings.HasPrefix(s, "[") {
			return EgressHost{}, fmt.Errorf("%w: %q", ErrIPLiteral, s)
		}
		p, err := strconv.ParseUint(s[i+1:], 10, 16)
		if err != nil || p == 0 {
			return EgressHost{}, fmt.Errorf("%w: %q", ErrInvalidPort, s)
		}
		host, port = s[:i], uint16(p)
	}
	h := EgressHost{Host: host, Port: port}
	if err := h.validate(); err != nil {
		return EgressHost{}, err
	}
	return h, nil
}

// String returns "host:port".
func (h EgressHost) String() string {
	return h.Host + ":" + strconv.FormatUint(uint64(h.Port), 10)
}

// compareEgressHosts orders by host, then port.
func compareEgressHosts(a, b EgressHost) int {
	return cmp.Or(strings.Compare(a.Host, b.Host), cmp.Compare(a.Port, b.Port))
}

func (h EgressHost) validate() error {
	if h.Port == 0 {
		return fmt.Errorf("%w: %q", ErrInvalidPort, h.String())
	}
	return validateHostname(h.Host)
}

// validateHostname checks the canonical hostname form described on EgressHost.
func validateHostname(host string) error {
	switch {
	case strings.Contains(host, "*"):
		return fmt.Errorf("%w: %q", ErrWildcardHost, host)
	case strings.ContainsAny(host, ":[]"):
		return fmt.Errorf("%w: %q", ErrIPLiteral, host)
	case host == "" || len(host) > maxHostnameLen:
		return fmt.Errorf("%w: %q: must be 1 to %d characters", ErrInvalidHostname, host, maxHostnameLen)
	}
	labels := strings.Split(host, ".")
	if len(labels) < minLabels {
		return fmt.Errorf("%w: %q: must have at least %d labels", ErrInvalidHostname, host, minLabels)
	}
	for _, label := range labels {
		if err := validateLabel(label); err != nil {
			return fmt.Errorf("%w: %q: %w", ErrInvalidHostname, host, err)
		}
	}
	if isDigits(labels[len(labels)-1]) {
		return fmt.Errorf("%w: %q", ErrIPLiteral, host)
	}
	return nil
}

var (
	errLabelLength   = errors.New("each label must be 1 to 63 characters")
	errLabelChars    = errors.New("labels may hold only a-z, 0-9 and '-' (use the xn-- form for other scripts)")
	errLabelHyphen   = errors.New("a label must not start or end with '-'")
	errLabelReserved = errors.New("labels with '--' in the third and fourth places are reserved for xn--")
)

func validateLabel(label string) error {
	if label == "" || len(label) > maxLabelLen {
		return errLabelLength
	}
	for _, r := range label {
		if !isLowerAlnum(r) && r != '-' {
			return errLabelChars
		}
	}
	if label[0] == '-' || label[len(label)-1] == '-' {
		return errLabelHyphen
	}
	if strings.HasPrefix(label[min(idnaReservedAt, len(label)):], "--") && !strings.HasPrefix(label, "xn--") {
		return errLabelReserved
	}
	return nil
}

func isLowerAlnum(r rune) bool { return ('a' <= r && r <= 'z') || isDigit(r) }

func isDigit(r rune) bool { return '0' <= r && r <= '9' }

func isDigits(s string) bool {
	for _, r := range s {
		if !isDigit(r) {
			return false
		}
	}
	return s != ""
}
