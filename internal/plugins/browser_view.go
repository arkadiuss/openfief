package plugins

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// ErrBrowserPathHeld reports a plugin that declares a browser path while
// another instance holds one: in M0 at most one may (D-114), because the
// core passes the owner's session cookie to the view that serves a path.
var ErrBrowserPathHeld = errors.New("another plugin instance holds a browser path; at most one may (D-114)")

// BrowserView is a plugin instance that holds a browser path (D-81): the
// path, and when the instance was created, which decides between two that
// both came to hold one at once.
type BrowserView struct {
	Instance string
	Path     string
	Created  time.Time
}

// senior reports whether v came before w: created earlier, or at the same
// time with a name that sorts first.
func (v BrowserView) senior(w BrowserView) bool {
	if !v.Created.Equal(w.Created) {
		return v.Created.Before(w.Created)
	}
	return v.Instance < w.Instance
}

// RefuseSecondBrowserView checks a plugin instance about to be installed,
// or upgraded to a version whose manifest declares the browser path given
// (empty for none), against the views other instances hold, as the
// openfief CLI does (D-114): it fails with an error wrapping
// ErrBrowserPathHeld, naming the holder, when path is not empty and another
// instance holds one.
func RefuseSecondBrowserView(instance, path string, held []BrowserView) error {
	if path == "" {
		return nil
	}
	others := slices.DeleteFunc(slices.Clone(held), func(v BrowserView) bool { return v.Instance == instance })
	if len(others) == 0 {
		return nil
	}
	return heldBy(others)
}

// MayHoldBrowserPath decides, as the operator does (D-114), whether self, a
// plugin whose running version would hold a browser path, may: when
// holding, self already holds it. held are the other instances that hold
// one. A plugin that doesn't hold one yet yields to any holder; of two that
// both hold one, the senior keeps it. A refusal wraps ErrBrowserPathHeld
// and names the holder.
func MayHoldBrowserPath(self BrowserView, holding bool, held []BrowserView) error {
	others := slices.DeleteFunc(slices.Clone(held), func(v BrowserView) bool { return v.Instance == self.Instance })
	if len(others) == 0 {
		return nil
	}
	if holding && !slices.ContainsFunc(others, func(v BrowserView) bool { return v.senior(self) }) {
		return nil
	}
	return heldBy(others)
}

// heldBy is the refusal naming the senior of views, which is not empty.
func heldBy(views []BrowserView) error {
	first := slices.MinFunc(views, func(a, b BrowserView) int {
		switch {
		case a.senior(b):
			return -1
		case b.senior(a):
			return 1
		}
		return strings.Compare(a.Instance, b.Instance)
	})
	return fmt.Errorf("%w: %s holds %s", ErrBrowserPathHeld, first.Instance, first.Path)
}
