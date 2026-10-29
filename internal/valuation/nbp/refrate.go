package nbp

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

const (
	// DefaultArchiveURL is where NBP publishes its interest rate files. The
	// Web API at api.nbp.pl serves exchange rates and gold prices only.
	DefaultArchiveURL = "https://static.nbp.pl"

	// archivePath is NBP's archive of its base rates: every change of the
	// reference, lombard, deposit, rediscount and discount rates since
	// 1998 (obowiazuje_od), the one in force now included, at
	// https://static.nbp.pl/dane/stopy/stopy_procentowe_archiwum.xml.
	archivePath = "/dane/stopy/stopy_procentowe_archiwum.xml"
	// referenceRateID marks the reference rate among a change's rates.
	referenceRateID = "ref"
	// maxArchiveBytes bounds a response: the archive is about 40 KB.
	maxArchiveBytes = 1 << 20
)

// ErrMalformedArchive reports a response that is not NBP's interest rate
// archive.
var ErrMalformedArchive = errors.New("malformed NBP interest rate archive")

// Archive reads NBP's archive of base interest rates.
type Archive struct {
	http    *http.Client
	baseURL string
}

// NewArchive returns an archive reader that sends its requests with hc to
// baseURL, normally DefaultArchiveURL. The caller owns hc, as with New.
func NewArchive(hc *http.Client, baseURL string) *Archive {
	return &Archive{http: hc, baseURL: baseURL}
}

// ReferenceRates returns every change of the NBP reference rate in the
// archive, in the order they took effect, each as a yearly fraction: 5.75%
// is 0.0575. A change that leaves the reference rate as it was, listing
// only other rates, is not one.
func (a *Archive) ReferenceRates(ctx context.Context) ([]pl.RateChange, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+archivePath, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build NBP archive request: %w", err)
	}
	req.Header.Set("Accept", "text/xml")
	req.Header.Set("User-Agent", userAgent)
	resp, err := a.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("NBP interest rate archive: %w", err)
	}
	changes, err := readArchive(resp)
	closeErr := resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("NBP interest rate archive: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close NBP archive response: %w", closeErr)
	}
	return changes, nil
}

// readArchive maps NBP's answer: 200 carries the archive, anything else is
// an error.
func readArchive(resp *http.Response) ([]pl.RateChange, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxArchiveBytes))
	if err != nil {
		return nil, fmt.Errorf("read NBP archive: %w", err)
	}
	return parseArchive(body)
}

// archive is NBP's XML shape: one pozycje element per change, dated by the
// day it took effect.
type archive struct {
	XMLName xml.Name        `xml:"stopy_procentowe_archiwum"`
	Changes []archiveChange `xml:"pozycje"`
}

// archiveChange is one change: each rate it sets, in percent with a decimal
// comma, by its id.
type archiveChange struct {
	From  string `xml:"obowiazuje_od,attr"`
	Rates []struct {
		ID   string `xml:"id,attr"`
		Rate string `xml:"oprocentowanie,attr"`
	} `xml:"pozycja"`
}

// parseArchive reads the reference rate changes out of the archive's XML,
// which NBP starts with a UTF-8 byte order mark.
func parseArchive(body []byte) ([]pl.RateChange, error) {
	var a archive
	if err := xml.Unmarshal(bytes.TrimPrefix(body, []byte("\uFEFF")), &a); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformedArchive, err)
	}
	out := make([]pl.RateChange, 0, len(a.Changes))
	for i := range a.Changes {
		c, ok, err := a.Changes[i].referenceRate()
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, c)
		}
	}
	slices.SortFunc(out, func(x, y pl.RateChange) int { return x.From.Compare(y.From) })
	for i := 1; i < len(out); i++ {
		if out[i].From.Equal(out[i-1].From) {
			return nil, fmt.Errorf("%w: two reference rates from %s", ErrMalformedArchive, out[i].From)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: no reference rate", ErrMalformedArchive)
	}
	return out, nil
}

// referenceRate returns the reference rate c sets, as a fraction, or false
// when c sets other rates only.
func (c *archiveChange) referenceRate() (pl.RateChange, bool, error) {
	from, err := kernel.ParseDate(c.From)
	if err != nil {
		return pl.RateChange{}, false, fmt.Errorf("%w: %w", ErrMalformedArchive, err)
	}
	for _, r := range c.Rates {
		if r.ID != referenceRateID {
			continue
		}
		rate, err := kernel.ParseDecimal(strings.Replace(r.Rate, ",", ".", 1))
		if err != nil {
			return pl.RateChange{}, false, fmt.Errorf("%w: reference rate from %s: %w", ErrMalformedArchive, from, err)
		}
		percent := kernel.NewDecimal(1, -2)
		return pl.RateChange{From: from, Rate: rate.Mul(percent)}, true, nil
	}
	return pl.RateChange{}, false, nil
}
