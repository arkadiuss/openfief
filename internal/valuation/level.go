package valuation

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrInvalidIndexLevel reports an index level without an index name of 1 to
// MaxIndexNameLen characters, a date or a positive level.
var ErrInvalidIndexLevel = errors.New("invalid index level")

// MaxIndexNameLen bounds an index's name.
const MaxIndexNameLen = 64

// IndexLevel is one end-of-day closing level of a market index, such as
// WIG20 (glossary: Index level), named as the market-data connector's config
// names the index. A level has no currency: whoever uses it knows the
// index's. No holding is valued from it; the model portfolio's backtest
// reads it for index targets (D-52, D-53). There is at most one level per
// index and date.
type IndexLevel struct {
	Index  string
	Date   kernel.Date
	Level  kernel.Decimal
	Source string
}

// Validate reports whether l names an index, has a date and a positive
// level.
func (l *IndexLevel) Validate() error {
	if l.Index == "" || len(l.Index) > MaxIndexNameLen {
		return fmt.Errorf("%w: an index name of 1 to %d characters is needed", ErrInvalidIndexLevel, MaxIndexNameLen)
	}
	if l.Date.IsZero() {
		return fmt.Errorf("%w: %s has no date", ErrInvalidIndexLevel, l.Index)
	}
	if l.Level.Sign() <= 0 {
		return fmt.Errorf("%w: %s on %s: level %s is not positive", ErrInvalidIndexLevel, l.Index, l.Date, l.Level)
	}
	return nil
}
