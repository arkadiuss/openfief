package pl

import (
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrNotPublished reports a series the ministry's published data doesn't
// list, such as one too new to be in the file yet.
var ErrNotPublished = errors.New("series not in the published retail bond data")

// ErrInvalidAttribution reports an attribution that lacks what the
// publisher's re-use terms ask to be named.
var ErrInvalidAttribution = errors.New("invalid attribution")

// Attribution says where series terms taken from published data came from,
// as the Ministry of Finance's re-use terms ask of anyone who re-uses its
// data (CC BY 4.0, R-27): the publisher named as the source, the dataset,
// where it was taken from, when the publisher made it and when it was taken,
// and how it was processed (D-97). Terms the owner entered have none.
type Attribution struct {
	// Publisher is the source, "Ministerstwo Finansów".
	Publisher string
	// Dataset is the publication's title, such as "Dane dotyczące obligacji
	// detalicznych".
	Dataset string
	// License is the publication's licence, such as "CC BY 4.0".
	License string
	// URL is the address of the file the terms were read from.
	URL string
	// Made is when the publisher made the data: the file's or the
	// resource's date.
	Made kernel.Date
	// Fetched is when the file was taken.
	Fetched time.Time
	// Processing says what was done to the data.
	Processing string
}

// Validate reports whether a names what the re-use terms ask: the
// publisher, the dataset and its licence, the file, when it was made and
// taken, and the processing.
// Otherwise it returns an error wrapping ErrInvalidAttribution.
func (a *Attribution) Validate() error {
	switch {
	case a.Publisher == "":
		return fmt.Errorf("%w: no publisher", ErrInvalidAttribution)
	case a.Dataset == "":
		return fmt.Errorf("%w: no dataset", ErrInvalidAttribution)
	case a.License == "":
		return fmt.Errorf("%w: no licence", ErrInvalidAttribution)
	case a.URL == "":
		return fmt.Errorf("%w: no address of the data", ErrInvalidAttribution)
	case a.Made.IsZero():
		return fmt.Errorf("%w: no date the data was made", ErrInvalidAttribution)
	case a.Fetched.IsZero():
		return fmt.Errorf("%w: no time the data was taken", ErrInvalidAttribution)
	case a.Processing == "":
		return fmt.Errorf("%w: no processing stated", ErrInvalidAttribution)
	}
	return nil
}

// PublishedRates are one series' first-period rate and margin as the
// ministry publishes them, as yearly fractions; the margin is zero for a
// fixed-rate series.
type PublishedRates struct {
	Series          Series
	FirstPeriodRate kernel.Decimal
	Margin          kernel.Decimal
}

// Published is the ministry's published retail bond data (D-97): every
// series' rates it lists, and where they came from. It carries no
// early-redemption fee; the reviewed fee table supplies that.
type Published struct {
	Rates       []PublishedRates
	Attribution Attribution
}

// RatesOf returns the published first-period rate and margin of series s. A
// series p doesn't list, such as one too new to be in the file yet, fails
// with ErrNotPublished. The data carries no early-redemption fee: the
// reviewed table (EarlyRedemptionFee) supplies it.
func (p *Published) RatesOf(s Series) (PublishedRates, error) {
	for _, r := range p.Rates {
		if r.Series == s {
			return r, nil
		}
	}
	return PublishedRates{}, fmt.Errorf("%w: %s", ErrNotPublished, s)
}
