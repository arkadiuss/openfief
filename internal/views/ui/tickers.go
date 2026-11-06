package ui

import (
	"net/http"
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

// statementRoutes maps the owner's statement uploads (D-49); the rows of
// an upload that wait for their tickers' ISINs (D-116) are confirmed in the
// Inbox.
func (s *server) statementRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /accounts/{id}/statements", s.withOwner(s.uploadStatement))
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
		http.Redirect(w, r, inboxPath+"?booked="+strconv.Itoa(int(resp.Msg.GetBooked())), http.StatusSeeOther)
		return
	}
	s.afterInboxForm(w, r, sess, "ConfirmTicker", err)
}
