package plugins

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// ErrSecretHeld reports a plugin instance that would mount a Secret
// another instance mounts. Secret names join the instance and the role
// with "--", which neither may hold (D-89), so no two valid instances name
// one Secret; this check stays as a second layer behind that rule: the
// operator refuses the second (D-182), as the CLI's app.ErrSecretInUse
// does at install and upgrade.
var ErrSecretHeld = errors.New("another plugin instance names that Secret")

// SecretHolder is a plugin instance and the Secrets it names, with when
// it was created, which decides between two that came to mount one at
// once.
type SecretHolder struct {
	Instance string
	Created  time.Time
	Secrets  []string
}

// MayMountSecrets decides, as the operator does, whether self may mount
// the Secrets it names: mounting are those it already mounts, and held the
// Secrets the other instances mount now. A Secret no other instance mounts
// is self's to mount. One another mounts is refused, unless self mounts it
// too and is the senior, the older Plugin by creation time and then by
// name, as when two came to mount it at once; then the other stops. A
// refusal wraps ErrSecretHeld and names the Secret and its holder, which
// the operator puts in the stopped plugin's status.
func MayMountSecrets(self SecretHolder, mounting []string, held []SecretHolder) error {
	for _, s := range self.Secrets {
		for _, h := range held {
			if h.Instance == self.Instance || !slices.Contains(h.Secrets, s) {
				continue
			}
			if slices.Contains(mounting, s) && self.senior(h) {
				continue
			}
			return fmt.Errorf("%w: %s, mounted by %s", ErrSecretHeld, s, h.Instance)
		}
	}
	return nil
}

// senior reports whether h came before o, as BrowserView.senior decides.
func (h SecretHolder) senior(o SecretHolder) bool {
	return BrowserView{Instance: h.Instance, Created: h.Created}.senior(BrowserView{Instance: o.Instance, Created: o.Created})
}
