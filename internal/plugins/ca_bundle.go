package plugins

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"maps"
	"slices"
)

// CABundleRole is the Secret role of an instance's extra CA bundle: the
// manifest of a connector that trusts extra CAs for its endpoint asks for
// <instance>--ca-bundle (D-89), marked optional. The instance mounts it only
// when the owner says one is expected (Plugin.ExtraCA).
const CABundleRole = "ca-bundle"

// CABundleAccessName names the Role and RoleBinding, in the openfief
// namespace, that let the operator get one instance's CA bundle Secret and
// no other: "openfief-plugin-<instance>-ca-bundle". The owner's install
// creates them for an instance that expects a bundle; the operator's own
// Role reads no Secret.
func CABundleAccessName(instance string) string {
	return "openfief-" + ServiceAccountName(instance) + "-" + CABundleRole
}

// CABundleState is what the owner sees about an instance's extra CA bundle.
type CABundleState string

// The CA bundle states.
const (
	// CABundleNotExpected: the instance doesn't expect a bundle, so its
	// Secret is not mounted and nothing is reported.
	CABundleNotExpected CABundleState = "NotExpected"
	// CABundlePresent: the instance expects a bundle and its Secret exists.
	CABundlePresent CABundleState = "Present"
	// CABundleMissing: the instance expects a bundle and its Secret doesn't
	// exist, so the plugin trusts only the system's roots and TLS to an
	// endpoint a private CA signs fails. It is reported, never waited for.
	CABundleMissing CABundleState = "Missing"
	// CABundleUnreadable: the instance expects a bundle, but the operator
	// may not read its Secret, because the Role CABundleAccessName names
	// is missing, so whether it exists is unknown.
	CABundleUnreadable CABundleState = "Unreadable"
)

// CABundleStateOf decides what to report about an instance's extra CA
// bundle: nothing unless the instance expects one; then, when the operator
// may read its Secret, whether that exists.
func CABundleStateOf(expected, readable, exists bool) CABundleState {
	switch {
	case !expected:
		return CABundleNotExpected
	case !readable:
		return CABundleUnreadable
	case exists:
		return CABundlePresent
	default:
		return CABundleMissing
	}
}

// SecretDataChecksum is the hex SHA-256 of a Secret's data, the same for
// the same keys and values whatever the map's order: each key and value in
// key order, each preceded by its length, so no two different maps hash the
// same input. It is a hash only, never the data; nil and an empty map have
// the same checksum.
func SecretDataChecksum(data map[string][]byte) string {
	h := sha256.New()
	var size [8]byte
	write := func(b []byte) {
		binary.BigEndian.PutUint64(size[:], uint64(len(b)))
		h.Write(size[:])
		h.Write(b)
	}
	for _, k := range slices.Sorted(maps.Keys(data)) {
		write([]byte(k))
		write(data[k])
	}
	return hex.EncodeToString(h.Sum(nil))
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

// asksForCABundle reports whether m asks for the CA bundle Secret role.
func (m *Manifest) asksForCABundle() bool {
	return slices.ContainsFunc(m.Secrets, func(s SecretMount) bool { return s.Role == CABundleRole })
}

// validateExtraCA checks that a Plugin expects an extra CA bundle only when
// its manifest asks for the CA bundle Secret role.
func (p *Plugin) validateExtraCA() []error {
	if p.ExtraCA && !p.Manifest.asksForCABundle() {
		return []error{fieldErr("extraCa", ErrExtraCANotRequested, CABundleRole)}
	}
	return nil
}

// SetExtraCA sets whether p expects an extra CA bundle. Expecting one needs
// a manifest that asks for the CA bundle Secret role; otherwise p is left as
// it was and the error says why. What p requests doesn't change: its grant
// approves the Secret either way.
func (p *Plugin) SetExtraCA(on bool) error {
	q := *p
	q.ExtraCA = on
	if errs := q.validateExtraCA(); len(errs) > 0 {
		return errs[0]
	}
	p.ExtraCA = on
	return nil
}
