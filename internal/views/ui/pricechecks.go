package ui

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
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

// priceChecksView is what the price checks page shows: the default
// threshold as a percentage, empty when the core compares no prices, the
// checks grouped by instrument, how many instruments' latest check is
// open, and the owner's refused form's reason.
type priceChecksView struct {
	Threshold   string
	Instruments []instrumentChecks
	Open        int
	Error       string
}

// priceChecksPage shows the latest price check of every instrument with
// its history (D-112): a day on which an institution's own price, such as
// IBKR's mark price, is compared with the latest price holdings are valued
// at on or before it (D-111). The owner can acknowledge a flagged check
// and set an instrument's threshold (D-110); nothing changes a price here.
func (s *server) priceChecksPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderPriceChecks(w, r, sess, http.StatusOK, "")
}

// renderPriceChecks shows the price checks page with status, and with the
// reason the core refused what the owner posted.
func (s *server) renderPriceChecks(w http.ResponseWriter, r *http.Request, sess session, status int, refusal string) {
	resp, err := s.read.ListPriceChecks(r.Context(), bearer(&viewsv1.ListPriceChecksRequest{}, sess))
	if err != nil {
		s.coreFailed(w, r, "ListPriceChecks", err)
		return
	}
	view := priceChecksView{
		Threshold:   percent(resp.Msg.GetThreshold()),
		Instruments: latestPerInstrument(resp.Msg.GetChecks(), resp.Msg.GetThreshold(), sess.Owner()), Error: refusal,
	}
	for i := range view.Instruments {
		if view.Instruments[i].Open() {
			view.Open++
		}
	}
	s.show(w, r, status, "price-checks", page{Title: "Price checks", Session: &sess, Data: view})
}

// acknowledgePriceCheck records the owner's acknowledgement of the
// flagged check the form names (D-112) and shows the page again.
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
		s.renderPriceChecks(w, r, sess, http.StatusBadRequest, err.Error())
		return
	}
	_, err = s.priceChecks.SetPriceCheckThreshold(r.Context(), connect.NewRequest(&valuationv1.SetPriceCheckThresholdRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Isin: r.PostForm.Get("isin"), Symbol: r.PostForm.Get("symbol"),
		Threshold: threshold,
	}))
	s.afterPriceCheckForm(w, r, sess, "SetPriceCheckThreshold", err)
}

// afterPriceCheckForm answers a posted price check form: back to the page
// when the core took it, the page with the core's reason when it refused
// it, and any other failure like any failed call to the core.
func (s *server) afterPriceCheckForm(w http.ResponseWriter, r *http.Request, sess session, name string, err error) {
	if err == nil {
		http.Redirect(w, r, "/price-checks", http.StatusSeeOther)
		return
	}
	code := connect.CodeOf(err)
	if status, refused := refusalStatus(code); refused && code != connect.CodePermissionDenied {
		s.renderPriceChecks(w, r, sess, status, "The core refused this: "+connectMessage(err))
		return
	}
	s.coreFailed(w, r, name, err)
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
