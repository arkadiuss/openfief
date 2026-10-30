package ui

import (
	"net/http"
	"net/url"
	"strings"

	"connectrpc.com/connect"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
)

// bondInstitution is the institution of accounts at the treasury bond
// registry, whose lots the owner enters by hand (D-9). Only their pages
// offer the "Add bond lot" form.
const bondInstitution = "treasury-bonds"

// Purchases on the "Add bond lot" form (D-100).
const (
	purchaseCash     = "cash"
	purchaseExchange = "exchange"
)

// bondLotForm is the "Add bond lot" form of a bond account's page: D-9's
// series, quantity and purchase date, how the bonds were bought (D-100),
// and the series' terms from its issue letter, as the owner typed them.
type bondLotForm struct {
	Series       string
	Quantity     string
	PurchaseDate string
	Purchase     string
	FirstRate    string
	Margin       string
	Fee          string
}

// Exchange reports whether the form says the bonds were bought by
// exchange, so the page keeps that choice when it shows the form again.
func (f *bondLotForm) Exchange() bool { return f.Purchase == purchaseExchange }

// parseBondLotForm reads the posted form, each field exactly once and
// trimmed, the series in upper case; the core checks every rule (D-49).
func parseBondLotForm(r *http.Request) (*bondLotForm, bool) {
	if err := r.ParseForm(); err != nil {
		return nil, false
	}
	v, ok := postedOnce(r, "series", "quantity", "purchase_date", "purchase", "first_rate", "margin", "fee")
	if !ok || (v[3] != purchaseCash && v[3] != purchaseExchange) {
		return nil, false
	}
	return &bondLotForm{
		Series: strings.ToUpper(v[0]), Quantity: v[1], PurchaseDate: v[2], Purchase: v[3], FirstRate: v[4], Margin: v[5], Fee: v[6],
	}, true
}

// postedOnce returns the posted form's fields names, each given exactly
// once, trimmed, or false.
func postedOnce(r *http.Request, names ...string) ([]string, bool) {
	v := make([]string, len(names))
	for i, name := range names {
		if len(r.PostForm[name]) != 1 {
			return nil, false
		}
		v[i] = strings.TrimSpace(r.PostForm[name][0])
	}
	return v, true
}

// message turns the form into the lot the core takes.
func (f *bondLotForm) message() *ingestv1.BondLot {
	return &ingestv1.BondLot{
		Series: f.Series, Quantity: f.Quantity, PurchaseDate: f.PurchaseDate,
		FirstPeriodRatePercent: f.FirstRate, MarginPercent: f.Margin, EarlyRedemptionFeePln: f.Fee,
		Purchase: f.purchase(),
	}
}

// purchase returns the form's purchase as the core takes it.
func (f *bondLotForm) purchase() ingestv1.BondPurchase {
	if f.Exchange() {
		return ingestv1.BondPurchase_BOND_PURCHASE_EXCHANGE
	}
	return ingestv1.BondPurchase_BOND_PURCHASE_CASH
}

// addBondLot hands the posted lot to the core (AddBondLot, D-49), which
// hands it to the connector bound to the account. A refusal, such as a
// purchase date that doesn't match the series, shows the account page
// again with the core's reason and the form as posted; so does a connector
// that can't be reached, so the owner can send it again.
func (s *server) addBondLot(w http.ResponseWriter, r *http.Request, sess session) {
	id := r.PathValue("id")
	f, ok := parseBondLotForm(r)
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The bond lot form is malformed.")
		return
	}
	_, err := s.bondLots.AddBondLot(r.Context(), connect.NewRequest(&ingestv1.AddBondLotRequest{
		Origin: sess.Origin, SessionToken: sess.Token, AccountId: id, Lot: f.message(),
	}))
	if err == nil {
		http.Redirect(w, r, "/accounts/"+url.PathEscape(id)+"#lots-title", http.StatusSeeOther)
		return
	}
	code := connect.CodeOf(err)
	if code == connect.CodeUnavailable {
		s.renderAccount(w, r, sess, http.StatusServiceUnavailable, accountForm{
			Lot: f, Error: "The bond lot was not added: its connector can't be reached now. Try again in a moment.",
		})
		return
	}
	if status, refused := refusalStatus(code); refused {
		s.renderAccount(w, r, sess, status, accountForm{Lot: f, Error: "The bond lot was not added: " + connectMessage(err)})
		return
	}
	s.coreFailed(w, r, "AddBondLot", err)
}

// deleteBondLot hands the lot a delete button names, by its series and
// purchase date, to the core (DeleteBondLot, D-100), which deletes it as a
// correction: its transactions are superseded and kept. A refusal, or a
// connector that can't be reached, shows the account page again with the
// reason.
func (s *server) deleteBondLot(w http.ResponseWriter, r *http.Request, sess session) {
	id := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The bond lot form is malformed.")
		return
	}
	v, ok := postedOnce(r, "series", "purchase_date")
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The bond lot form is malformed.")
		return
	}
	_, err := s.bondLots.DeleteBondLot(r.Context(), connect.NewRequest(&ingestv1.DeleteBondLotRequest{
		Origin: sess.Origin, SessionToken: sess.Token, AccountId: id, Series: strings.ToUpper(v[0]), PurchaseDate: v[1],
	}))
	if err == nil {
		http.Redirect(w, r, "/accounts/"+url.PathEscape(id)+"#lots-title", http.StatusSeeOther)
		return
	}
	code := connect.CodeOf(err)
	if code == connect.CodeUnavailable {
		s.renderAccount(w, r, sess, http.StatusServiceUnavailable, accountForm{
			Error: "The bond lot was not deleted: its connector can't be reached now. Try again in a moment.",
		})
		return
	}
	if status, refused := refusalStatus(code); refused {
		s.renderAccount(w, r, sess, status, accountForm{Error: "The bond lot was not deleted: " + connectMessage(err)})
		return
	}
	s.coreFailed(w, r, "DeleteBondLot", err)
}
