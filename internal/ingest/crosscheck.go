package ingest

import (
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// CrossCheckPrice is an institution's own end-of-day price of one unit of an
// instrument as a connector maps it (glossary: Cross-check price), such as
// the mark price of a position an IBKR Flex statement reports, in the
// currency the institution reports it in. The core stores it only to
// cross-check the market-data connector's prices (D-22), never as a price.
type CrossCheckPrice struct {
	Instrument ledger.Instrument
	Date       kernel.Date
	Close      kernel.Money
}
