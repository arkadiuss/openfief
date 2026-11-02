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

// statementRoutes maps the owner's statement uploads (D-49) and the
// tickers page and its form, where the rows of an upload wait for their
// tickers' ISINs (D-116).
func (s *server) statementRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /accounts/{id}/statements", s.withOwner(s.uploadStatement))
	mux.HandleFunc("GET /tickers", s.withOwner(s.tickersPage))
	mux.HandleFunc("POST /tickers/confirm", s.withOwner(s.confirmTicker))
	mux.HandleFunc("POST /tickers/correct", s.withOwner(s.correctTicker))
}

// tickersView is what the tickers page shows: the tickers waiting for the
// owner's confirmation, those confirmed, which the owner can correct
// (D-153), the rows the last confirmation or correction booked, and the
// core's reason for refusing one.
type tickersView struct {
	Tickers   []*ingestv1.PendingTicker
	Confirmed []*ingestv1.ConfirmedTicker
	Booked    string
	Corrected string
	Error     string
}

// tickersPage shows the owner the tickers whose ISIN isn't confirmed yet
// (D-116): each with the ISIN the market-data connector proposed, to
// confirm, or a field to type it when the lookup found none, and a choice
// of trading currency when its exchange quotes in more than one, required
// when the proposal names none (D-170).
func (s *server) tickersPage(w http.ResponseWriter, r *http.Request, sess session) {
	var view tickersView
	if n, err := strconv.Atoi(r.URL.Query().Get("booked")); err == nil && n >= 0 {
		view.Booked = strconv.Itoa(n)
	}
	if n, err := strconv.Atoi(r.URL.Query().Get("corrected")); err == nil && n >= 0 {
		view.Corrected = strconv.Itoa(n)
	}
	s.renderTickers(w, r, sess, http.StatusOK, &view)
}

// renderTickers shows the tickers page with status and view's messages.
func (s *server) renderTickers(w http.ResponseWriter, r *http.Request, sess session, status int, view *tickersView) {
	resp, err := s.tickers.ListPendingTickers(r.Context(), connect.NewRequest(&ingestv1.ListPendingTickersRequest{
		Origin: sess.Origin, SessionToken: sess.Token,
	}))
	if err != nil {
		s.coreFailed(w, r, "ListPendingTickers", err)
		return
	}
	view.Tickers = resp.Msg.GetTickers()
	confirmed, err := s.tickers.ListConfirmedTickers(r.Context(), connect.NewRequest(&ingestv1.ListConfirmedTickersRequest{
		Origin: sess.Origin, SessionToken: sess.Token,
	}))
	if err != nil {
		s.coreFailed(w, r, "ListConfirmedTickers", err)
		return
	}
	view.Confirmed = confirmed.Msg.GetTickers()
	s.show(w, r, status, "tickers", page{Title: "Tickers to confirm", Session: &sess, Data: view})
}

// Caps of the ticker form's fields.
const (
	maxISINChars     = 12
	maxCurrencyChars = 3
)

// confirmTicker hands the owner's ISIN for a ticker, the proposal or one
// typed, to the core (ConfirmTicker, D-116), with the trading currency the
// owner picked, if any (D-170), which books the rows waiting for it, and
// shows the page again with how many it booked, or the core's reason for
// refusing it, such as the currencies to pick from.
func (s *server) confirmTicker(w http.ResponseWriter, r *http.Request, sess session) {
	f, ok := s.tickerForm(w, r, sess)
	if !ok {
		return
	}
	resp, err := s.tickers.ConfirmTicker(r.Context(), connect.NewRequest(&ingestv1.ConfirmTickerRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Instance: f.instance, Symbol: f.symbol, Isin: f.isin, Currency: f.currency,
	}))
	if err == nil {
		http.Redirect(w, r, "/tickers?booked="+strconv.Itoa(int(resp.Msg.GetBooked())), http.StatusSeeOther)
		return
	}
	s.tickerRefused(w, r, sess, "ConfirmTicker", err)
}

// correctTicker hands the owner's corrected ISIN for a confirmed ticker to
// the core (CorrectTicker, D-153), with the trading currency the owner
// picked, if any (D-170), which books every row of the ticker anew,
// superseding what was booked under the old ISIN, and shows the page again
// with how many it booked, or the core's reason for refusing it.
func (s *server) correctTicker(w http.ResponseWriter, r *http.Request, sess session) {
	f, ok := s.tickerForm(w, r, sess)
	if !ok {
		return
	}
	resp, err := s.tickers.CorrectTicker(r.Context(), connect.NewRequest(&ingestv1.CorrectTickerRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Instance: f.instance, Symbol: f.symbol, Isin: f.isin, Currency: f.currency,
	}))
	if err == nil {
		http.Redirect(w, r, "/tickers?corrected="+strconv.Itoa(int(resp.Msg.GetBooked())), http.StatusSeeOther)
		return
	}
	s.tickerRefused(w, r, sess, "CorrectTicker", err)
}

// tickerFormValues are a ticker form's fields as posted, the ISIN upper-cased.
type tickerFormValues struct {
	instance, symbol, isin, currency string
}

// tickerForm reads a ticker form, or shows why it is refused and returns
// false: a form that doesn't parse, an ISIN longer than 12 characters or a
// trading currency longer than 3.
func (s *server) tickerForm(w http.ResponseWriter, r *http.Request, sess session) (tickerFormValues, bool) {
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return tickerFormValues{}, false
	}
	f := tickerFormValues{
		instance: r.PostForm.Get("instance"), symbol: r.PostForm.Get("symbol"),
		isin: strings.ToUpper(strings.TrimSpace(r.PostForm.Get("isin"))), currency: strings.TrimSpace(r.PostForm.Get("currency")),
	}
	switch {
	case len(f.isin) > maxISINChars:
		s.renderTickers(w, r, sess, http.StatusBadRequest, &tickersView{Error: "An ISIN has 12 characters, such as PLTESTA00002."})
		return tickerFormValues{}, false
	case len(f.currency) > maxCurrencyChars:
		s.renderTickers(w, r, sess, http.StatusBadRequest, &tickersView{Error: "A trading currency has 3 letters, such as GBX."})
		return tickerFormValues{}, false
	}
	return f, true
}

// tickerRefused shows the core's refusal of a confirmation or correction
// on the page, or an error page when the core failed.
func (s *server) tickerRefused(w http.ResponseWriter, r *http.Request, sess session, method string, err error) {
	code := connect.CodeOf(err)
	if status, refused := refusalStatus(code); refused && code != connect.CodePermissionDenied {
		s.renderTickers(w, r, sess, status, &tickersView{Error: "The core refused this: " + connectMessage(err)})
		return
	}
	s.coreFailed(w, r, method, err)
}
