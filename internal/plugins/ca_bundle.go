package plugins

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"slices"
)

// CABundleRole is the Secret role of an instance's extra CA bundle: the
// manifest of a connector that trusts extra CAs for its endpoint asks for
// <instance>--ca-bundle (D-89), marked optional. The instance mounts it only
// when the owner says one is expected (Plugin.ExtraCA).
const CABundleRole = "ca-bundle"

// CABundleKey is the key of the CA bundle Secret that holds the PEM
// certificates, which the plugin reads from <mount path>/ca.crt.
const CABundleKey = "ca.crt"

// Errors of a CA bundle file.
var (
	// ErrCABundleEmpty: the file holds no PEM block at all.
	ErrCABundleEmpty = errors.New("no PEM certificate in the CA bundle")
	// ErrCABundleNotPEM: the file holds something besides PEM blocks.
	ErrCABundleNotPEM = errors.New("the CA bundle holds something that isn't a PEM block")
	// ErrCABundleNotCertificate: a PEM block isn't a CERTIFICATE, such as a
	// private key, which never belongs in a CA bundle.
	ErrCABundleNotCertificate = errors.New("the CA bundle holds a PEM block that isn't a CERTIFICATE")
	// ErrCABundleBadCertificate: a CERTIFICATE block doesn't parse.
	ErrCABundleBadCertificate = errors.New("the CA bundle holds a certificate that doesn't parse")
	// ErrExtraCANotExpected: a CA bundle is set for an instance that doesn't
	// expect one (spec.extraCa), which would mount nothing.
	ErrExtraCANotExpected = errors.New("the instance doesn't expect an extra CA bundle; upgrade it with --extra-ca first")
)

// CABundle is a CA bundle file the owner gave, checked: the PEM as given,
// how many certificates it holds, and its checksum, the hex SHA-256 of the
// PEM, which is exactly what the Secret holds under CABundleKey. The CLI
// stamps the checksum on the Plugin (spec.caBundleSha256), so the operator,
// which never reads the Secret, rolls the pod when it changes.
type CABundle struct {
	PEM          []byte
	Certificates int
	Checksum     string
}

// ParseCABundle checks that data is PEM certificates and nothing else:
// every block a CERTIFICATE that parses, at least one, and nothing but
// whitespace around them. It returns the bundle with its checksum.
func ParseCABundle(data []byte) (CABundle, error) {
	rest := data
	n := 0
	for {
		block, next := pem.Decode(rest)
		if block == nil {
			break
		}
		n++
		if block.Type != "CERTIFICATE" {
			return CABundle{}, fmt.Errorf("%w: block %d is %q", ErrCABundleNotCertificate, n, block.Type)
		}
		if _, err := x509.ParseCertificate(block.Bytes); err != nil {
			return CABundle{}, fmt.Errorf("%w: block %d: %w", ErrCABundleBadCertificate, n, err)
		}
		rest = next
	}
	if len(bytes.TrimSpace(rest)) > 0 {
		return CABundle{}, fmt.Errorf("%w: after %d certificates", ErrCABundleNotPEM, n)
	}
	if n == 0 {
		return CABundle{}, ErrCABundleEmpty
	}
	return CABundle{PEM: bytes.Clone(data), Certificates: n, Checksum: CABundleChecksum(data)}, nil
}

// CABundleChecksum is the hex SHA-256 of a CA bundle's PEM, what
// `sha256sum ca.crt` prints for the same file: a hash, never the data.
func CABundleChecksum(pemData []byte) string {
	sum := sha256.Sum256(pemData)
	return hex.EncodeToString(sum[:])
}

// validCABundleChecksum reports whether s is a hex SHA-256 as
// CABundleChecksum writes it: 64 lower-case hex digits.
func validCABundleChecksum(s string) bool {
	const sha256Hex = 2 * sha256.Size
	if len(s) != sha256Hex {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// CABundleState is what the owner sees about an instance's extra CA bundle,
// read from the Plugin's spec alone: the operator never reads the Secret.
type CABundleState string

// The CA bundle states.
const (
	// CABundleNotExpected: the instance doesn't expect a bundle, so its
	// Secret is not mounted and nothing is reported.
	CABundleNotExpected CABundleState = "NotExpected"
	// CABundleNotStamped: the instance expects a bundle and no checksum is
	// stamped on its spec, so no bundle was set with the CLI (or the
	// GitOps equivalent): the plugin likely trusts only the system's roots.
	// It is reported, never waited for.
	CABundleNotStamped CABundleState = "NotStamped"
	// CABundleStamped: the instance expects a bundle and its spec carries
	// the bundle's checksum. Whether the Secret matches it is not checked;
	// a mismatch shows only as the plugin's TLS failures.
	CABundleStamped CABundleState = "Stamped"
)

// CABundleStateOf decides what to report about an instance's extra CA
// bundle from its spec: whether it expects one (extraCa) and the checksum
// stamped on it (caBundleSha256), empty when none is.
func CABundleStateOf(extraCA bool, checksum string) CABundleState {
	switch {
	case !extraCA:
		return CABundleNotExpected
	case checksum == "":
		return CABundleNotStamped
	default:
		return CABundleStamped
	}
}

// CABundleSecret returns the name of the instance's CA bundle Secret when
// e mounts it: the instance expects a bundle, its manifest asks for the
// role and its grant approves the Secret.
func (e *EffectivePermissions) CABundleSecret() (string, bool) {
	for _, s := range e.secrets {
		if s.Role == CABundleRole {
			return SecretName(e.plugin, s.Role), true
		}
	}
	return "", false
}

// AsksForCABundle reports whether m asks for the CA bundle Secret role.
func (m *Manifest) AsksForCABundle() bool {
	return slices.ContainsFunc(m.Secrets, func(s SecretMount) bool { return s.Role == CABundleRole })
}

// validateExtraCA checks that a Plugin expects an extra CA bundle only when
// its manifest asks for the CA bundle Secret role, and that a stamped
// checksum is a hex SHA-256.
func (p *Plugin) validateExtraCA() []error {
	var errs []error
	if p.ExtraCA && !p.Manifest.AsksForCABundle() {
		errs = append(errs, fieldErr("extraCa", ErrExtraCANotRequested, CABundleRole))
	}
	if p.CABundleSHA256 != "" && !validCABundleChecksum(p.CABundleSHA256) {
		errs = append(errs, fieldErr("caBundleSha256", ErrCABundleChecksum, p.CABundleSHA256))
	}
	return errs
}

// SetExtraCA sets whether p expects an extra CA bundle. Expecting one needs
// a manifest that asks for the CA bundle Secret role; otherwise p is left as
// it was and the error says why. What p requests doesn't change: its grant
// approves the Secret either way. Turning it off clears the stamped
// checksum, so turning it on again asks for the bundle to be set again.
func (p *Plugin) SetExtraCA(on bool) error {
	q := *p
	q.ExtraCA = on
	if errs := q.validateExtraCA(); len(errs) > 0 {
		return errs[0]
	}
	p.ExtraCA = on
	if !on {
		p.CABundleSHA256 = ""
	}
	return nil
}

// StampCABundle stamps the checksum of the CA bundle the owner set on p;
// an empty checksum clears it. Stamping one needs p to expect a bundle
// (ErrExtraCANotExpected) and a hex SHA-256 (ErrCABundleChecksum);
// otherwise p is left as it was.
func (p *Plugin) StampCABundle(checksum string) error {
	if checksum == "" {
		p.CABundleSHA256 = ""
		return nil
	}
	if !p.ExtraCA {
		return fmt.Errorf("%w: %s", ErrExtraCANotExpected, p.Name)
	}
	if !validCABundleChecksum(checksum) {
		return fmt.Errorf("%w: %q", ErrCABundleChecksum, checksum)
	}
	p.CABundleSHA256 = checksum
	return nil
}
