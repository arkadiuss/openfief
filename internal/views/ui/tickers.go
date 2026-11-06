package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
)

// waitingReason starts the reason of an ingest problem that holds a row
// until the owner confirms its ticker's ISIN (ingest.WaitingForISIN).
const waitingReason = "waits for the owner to confirm the ISIN of "

// waitingRows counts the problems of an upload that wait for a ticker's
// ISIN (D-116), for the upload result's link to the tickers page.
func waitingRows(problems []*ingestv1.Problem) int {
	n := 0
	for _, p := range problems {
		if strings.HasPrefix(p.GetReason(), waitingReason) {
			n++
		}
	}
	return n
}

// statementRoutes maps the owner's statement uploads (D-49) and the
// owner's corrections of a confirmed ticker on its instrument's page
// (D-153); the rows of an upload that wait for their tickers' ISINs
// (D-116) are confirmed in the Inbox.
func (s *server) statementRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /accounts/{id}/statements", s.withOwner(s.uploadStatement))
	mux.HandleFunc("POST /instruments/{isin}/tickers/correct", s.withOwner(s.correctTicker))
}

// Caps of the ticker form's fields.
const (
	maxISINChars     = 12
	maxCurrencyChars = 3
)

// confirmTicker hands the owner's ISIN for a ticker, the proposal or one
// typed, to the core (ConfirmTicker, D-116), with the trading currency the
// owner picked, if any (D-170), which books the rows waiting for it, and
// shows the Inbox again with how many it booked, or the core's reason for
// refusing it, such as the currencies to pick from.
func (s *server) confirmTicker(w http.ResponseWriter, r *http.Request, sess session) {
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	isin := strings.ToUpper(strings.TrimSpace(r.PostForm.Get("isin")))
	if len(isin) > maxISINChars {
		s.renderInbox(w, r, sess, http.StatusBadRequest, "", "An ISIN has 12 characters, such as PLTESTA00002.")
		return
	}
	currency := strings.TrimSpace(r.PostForm.Get("currency"))
	if len(currency) > maxCurrencyChars {
		s.renderInbox(w, r, sess, http.StatusBadRequest, "", "A trading currency has 3 letters, such as GBX.")
		return
	}
	resp, err := s.tickers.ConfirmTicker(r.Context(), connect.NewRequest(&ingestv1.ConfirmTickerRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Instance: r.PostForm.Get("instance"),
		Symbol: r.PostForm.Get("symbol"), Isin: isin, Currency: currency,
	}))
	if err == nil {
		q := url.Values{"booked": {strconv.Itoa(int(resp.Msg.GetBooked()))}, "isin": {isin}}
		http.Redirect(w, r, inboxPath+"?"+q.Encode(), http.StatusSeeOther)
		return
	}
	s.afterInboxForm(w, r, sess, "ConfirmTicker", err)
}

// confirmedTickers returns the tickers whose confirmed ISIN is isin
// (D-153), for the owner's corrections on the instrument's page; none for
// another session.
func (s *server) confirmedTickers(r *http.Request, sess session, isin string) ([]*ingestv1.ConfirmedTicker, error) {
	if !sess.Owner() {
		return nil, nil
	}
	resp, err := s.tickers.ListConfirmedTickers(r.Context(), connect.NewRequest(&ingestv1.ListConfirmedTickersRequest{
		Origin: sess.Origin, SessionToken: sess.Token,
	}))
	if err != nil {
		return nil, fmt.Errorf("ListConfirmedTickers: %w", err)
	}
	var out []*ingestv1.ConfirmedTicker
	for _, t := range resp.Msg.GetTickers() {
		if t.GetIsin() == isin {
			out = append(out, t)
		}
	}
	return out, nil
}

// correctTicker hands the owner's corrected ISIN for a ticker confirmed
// with the instrument's ISIN to the core (CorrectTicker, D-153), with the
// trading currency the owner picked, if any (D-170), which books every
// row of the ticker anew, superseding what was booked under the old ISIN,
// then shows the corrected instrument's page with how many rows it booked;
// a refusal shows the instrument's page again with the core's reason.
func (s *server) correctTicker(w http.ResponseWriter, r *http.Request, sess session) {
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	v, ok := postedOnce(r, "instance", "symbol", "isin")
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	from := r.PathValue("isin")
	isin, currency := strings.ToUpper(strings.TrimSpace(v[2])), strings.TrimSpace(r.PostForm.Get("currency"))
	switch {
	case len(isin) > maxISINChars:
		s.renderInstrument(w, r, sess, from, http.StatusBadRequest, instrumentForm{Error: "An ISIN has 12 characters, such as PLTESTA00002."})
		return
	case len(currency) > maxCurrencyChars:
		s.renderInstrument(w, r, sess, from, http.StatusBadRequest, instrumentForm{Error: "A trading currency has 3 letters, such as GBX."})
		return
	}
	resp, err := s.tickers.CorrectTicker(r.Context(), connect.NewRequest(&ingestv1.CorrectTickerRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Instance: v[0], Symbol: v[1], Isin: isin, Currency: currency,
	}))
	if err == nil {
		q := url.Values{"corrected": {strconv.Itoa(int(resp.Msg.GetBooked()))}}
		http.Redirect(w, r, "/instruments/"+url.PathEscape(isin)+"?"+q.Encode(), http.StatusSeeOther)
		return
	}
	s.afterInstrumentForm(w, r, sess, from, "CorrectTicker", err)
}
