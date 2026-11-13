package openfief

import (
	"errors"
	"fmt"
	"strings"
)

// LogSafe describes err for a log line without any text it may carry
// (GR-12): the messages of the errors among known in its chain, which the
// caller vouches name a rule and never a request, an answer, an amount or a
// key, joined by "; ", or else only the innermost error's Go type, such as
// "*url.Error", whose text may quote a URL or a reply. Nil is "". A plugin
// passes its own known errors; none is assumed safe. It is internal/kernel's
// LogSafe for plugins, which may not import internal/.
func LogSafe(err error, known ...error) string {
	if err == nil {
		return ""
	}
	var named []string
	for _, k := range known {
		if errors.Is(err, k) {
			named = append(named, k.Error())
		}
	}
	if len(named) > 0 {
		return strings.Join(named, "; ")
	}
	return fmt.Sprintf("unclassified error of type %T", innermost(err))
}

// innermost follows err's chain, the first of joined errors at each step,
// to the error that wraps nothing.
func innermost(err error) error {
	for {
		if next := errors.Unwrap(err); next != nil {
			err = next
			continue
		}
		var joined interface{ Unwrap() []error }
		if errors.As(err, &joined) {
			if errs := joined.Unwrap(); len(errs) > 0 && errs[0] != nil {
				err = errs[0]
				continue
			}
		}
		return err
	}
}
