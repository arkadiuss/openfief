package ui

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// Price check form limits and the precision of a threshold in percent.
const (
	// maxPercentChars caps a threshold field: "12.345" is plenty.
	maxPercentChars = 12
	// thresholdPlaces is how many decimal places a threshold fraction
	// keeps: a percent with up to four places.
	thresholdPlaces = 6
)

// errBadPercent reports a threshold field that isn't a percent.
var errBadPercent = errors.New("the threshold must be a percent greater than 0 and less than 100, such as 2.5")

// checkRow is one price check as a table row shows it: Acknowledge offers
// the owner its acknowledgement, on a flagged check not acknowledged yet.
type checkRow struct {
	Check       *viewsv1.PriceCheck
	Acknowledge bool
	// Next is the page the acknowledgement returns to, the Inbox when
	// empty.
	Next string
}

// instrumentChecks are one instrument's price checks: the latest, and the
// ones before it, newest first (D-112).
type instrumentChecks struct {
	Instrument *viewsv1.Instrument
	Latest     checkRow
	History    []checkRow
	// Threshold is the latest check's, in percent; Custom is the owner's
	// own for the instrument, in percent, empty for the default.
	Threshold string
	Custom    string
	// Owner shows the owner's threshold form.
	Owner bool
	// Next is the page the owner's forms return to, the Inbox when empty.
	Next string
}

// Open reports whether the latest check is flagged and not acknowledged:
// one for the owner to look at.
func (c *instrumentChecks) Open() bool { return openCheck(c.Latest.Check) }

// openCheck reports whether c is flagged and not acknowledged.
func openCheck(c *viewsv1.PriceCheck) bool {
	return c.GetStatus() == viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_FLAGGED && !c.GetAcknowledged()
}

// latestPerInstrument groups checks by instrument, in instrument order:
// each instrument's latest check by date, then by source, and the others
// newest first (D-112). def is the default threshold, to tell an
// instrument's own threshold apart; owner offers the owner's actions.
func latestPerInstrument(checks []*viewsv1.PriceCheck, def string, owner bool) []instrumentChecks {
	byKey := map[string][]*viewsv1.PriceCheck{}
	var keys []string
	for _, c := range checks {
		k := instrumentKey(c.GetInstrument())
		if _, ok := byKey[k]; !ok {
			keys = append(keys, k)
		}
		byKey[k] = append(byKey[k], c)
	}
	slices.Sort(keys)
	out := make([]instrumentChecks, 0, len(keys))
	for _, k := range keys {
		cs := byKey[k]
		slices.SortStableFunc(cs, func(a, b *viewsv1.PriceCheck) int {
			return cmp.Or(-strings.Compare(a.GetDate(), b.GetDate()), -strings.Compare(a.GetCrossCheckSource(), b.GetCrossCheckSource()))
		})
		rows := make([]checkRow, len(cs))
		for i, c := range cs {
			rows[i] = checkRow{Check: c, Acknowledge: owner && openCheck(c)}
		}
		ic := instrumentChecks{
			Instrument: cs[0].GetInstrument(), Latest: rows[0], History: rows[1:],
			Threshold: percent(cs[0].GetThreshold()), Owner: owner,
		}
		if t := cs[0].GetThreshold(); t != def {
			ic.Custom = strings.TrimSuffix(percent(t), "%")
		}
		out = append(out, ic)
	}
	return out
}

// acknowledgePriceCheck records the owner's acknowledgement of the
// flagged check the form names (D-112) and shows the Inbox again.
func (s *server) acknowledgePriceCheck(w http.ResponseWriter, r *http.Request, sess session) {
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	_, err := s.priceChecks.AcknowledgePriceCheck(r.Context(), connect.NewRequest(&valuationv1.AcknowledgePriceCheckRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Isin: r.PostForm.Get("isin"), Symbol: r.PostForm.Get("symbol"),
		Date: r.PostForm.Get("date"), Source: r.PostForm.Get("source"),
	}))
	s.afterPriceCheckForm(w, r, sess, "AcknowledgePriceCheck", err)
}

// setPriceCheckThreshold sets the instrument's threshold from the form's
// percent, or puts it back on the default when the field is empty (D-110).
func (s *server) setPriceCheckThreshold(w http.ResponseWriter, r *http.Request, sess session) {
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	threshold, err := fractionOf(r.PostForm.Get("threshold"))
	if err != nil {
		s.refusePriceCheckForm(w, r, sess, err.Error())
		return
	}
	_, err = s.priceChecks.SetPriceCheckThreshold(r.Context(), connect.NewRequest(&valuationv1.SetPriceCheckThresholdRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Isin: r.PostForm.Get("isin"), Symbol: r.PostForm.Get("symbol"),
		Threshold: threshold,
	}))
	s.afterPriceCheckForm(w, r, sess, "SetPriceCheckThreshold", err)
}

// afterPriceCheckForm answers a posted price check form: back to the page
// it came from when the core took it, the Inbox or the instrument's page
// (priceCheckReturn), that page with the core's reason when it refused
// it, and any other failure like any failed call to the core.
func (s *server) afterPriceCheckForm(w http.ResponseWriter, r *http.Request, sess session, name string, err error) {
	isin, toInstrument := priceCheckReturn(r)
	switch {
	case err == nil && toInstrument:
		http.Redirect(w, r, "/instruments/"+url.PathEscape(isin), http.StatusSeeOther)
	case err == nil:
		http.Redirect(w, r, backToInbox(r), http.StatusSeeOther)
	case toInstrument:
		s.afterInstrumentForm(w, r, sess, isin, name, err)
	default:
		s.afterInboxForm(w, r, sess, name, err)
	}
}

// priceCheckReturn returns the ISIN of the instrument whose page a posted
// price check form came from, and false for the Inbox: its next field
// must be the page of the instrument the form names, nothing else.
func priceCheckReturn(r *http.Request) (string, bool) {
	isin := r.PostFormValue("isin")
	next := r.PostFormValue("next")
	return isin, isin != "" && next == instrumentPath(isin)
}

// refusePriceCheckForm shows the page a price check form came from with
// reason, refused before the core was called.
func (s *server) refusePriceCheckForm(w http.ResponseWriter, r *http.Request, sess session, reason string) {
	if isin, ok := priceCheckReturn(r); ok {
		s.renderInstrument(w, r, sess, isin, http.StatusBadRequest, instrumentForm{Error: reason})
		return
	}
	s.renderInbox(w, r, sess, http.StatusBadRequest, "", reason)
}

// fractionOf turns a threshold the owner typed in percent, such as "2.5",
// into the fraction the core takes, "0.025"; empty stays empty, which puts
// the instrument back on the default. Anything but a percent greater than
// 0 and less than 100 is errBadPercent.
func fractionOf(p string) (string, error) {
	p = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(p), "%"))
	if p == "" {
		return "", nil
	}
	full := kernel.NewDecimal(hundred, 0)
	d, err := kernel.ParseDecimal(p)
	if err != nil || len(p) > maxPercentChars || d.Sign() <= 0 || d.Cmp(full) >= 0 {
		return "", errBadPercent
	}
	f, err := d.Quo(full, thresholdPlaces, kernel.RoundHalfEven)
	if err != nil || f.Sign() <= 0 {
		return "", errBadPercent
	}
	return f.String(), nil
}

// flaggedPriceChecks counts, for the overview, the instruments whose
// latest price check is flagged and not acknowledged.
func (s *server) flaggedPriceChecks(r *http.Request, sess session) (int, error) {
	resp, err := s.read.ListPriceChecks(r.Context(), bearer(&viewsv1.ListPriceChecksRequest{}, sess))
	if err != nil {
		return 0, fmt.Errorf("ListPriceChecks: %w", err)
	}
	n := 0
	for _, c := range latestPerInstrument(resp.Msg.GetChecks(), resp.Msg.GetThreshold(), false) {
		if c.Open() {
			n++
		}
	}
	return n, nil
}
