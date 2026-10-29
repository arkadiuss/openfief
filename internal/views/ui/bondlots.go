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

// bondLotForm is the "Add bond lot" form of a bond account's page: D-9's
// series, quantity and purchase date, and the series' terms from its issue
// letter, as the owner typed them.
type bondLotForm struct {
	Series       string
	Quantity     string
	PurchaseDate string
	FirstRate    string
	Margin       string
	Fee          string
}

// parseBondLotForm reads the posted form, each field exactly once and
// trimmed, the series in upper case; the core checks every rule (D-49).
func parseBondLotForm(r *http.Request) (*bondLotForm, bool) {
	if err := r.ParseForm(); err != nil {
		return nil, false
	}
	names := []string{"series", "quantity", "purchase_date", "first_rate", "margin", "fee"}
	v := make([]string, len(names))
	for i, name := range names {
		if len(r.PostForm[name]) != 1 {
			return nil, false
		}
		v[i] = strings.TrimSpace(r.PostForm[name][0])
	}
	return &bondLotForm{
		Series: strings.ToUpper(v[0]), Quantity: v[1], PurchaseDate: v[2], FirstRate: v[3], Margin: v[4], Fee: v[5],
	}, true
}

// message turns the form into the lot the core takes.
func (f *bondLotForm) message() *ingestv1.BondLot {
	return &ingestv1.BondLot{
		Series: f.Series, Quantity: f.Quantity, PurchaseDate: f.PurchaseDate,
		FirstPeriodRatePercent: f.FirstRate, MarginPercent: f.Margin, EarlyRedemptionFeePln: f.Fee,
	}
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
