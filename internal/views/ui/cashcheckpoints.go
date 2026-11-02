package ui

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// checkpointsAnchor is the account page's cash checkpoints section, where
// the owner enters one and books a difference (D-158).
const checkpointsAnchor = "#cash-checkpoints"

// checkpointRow is a cash checkpoint the owner entered, with its
// difference to the derived cash and, while they differ, the manual
// transaction that books it: "deposit" when the institution showed more
// cash, "withdrawal" when it showed less (D-158).
type checkpointRow struct {
	Check      *viewsv1.AssertionCheck
	Difference *viewsv1.Money
	Booking    string
}

// checkpointForm is the "Enter cash checkpoint" form, as the owner typed
// it.
type checkpointForm struct {
	Date, Amount, Currency string
}

// checkpointRows returns the checks the owner entered as cash checkpoints,
// in the order given, with their differences.
func checkpointRows(checks []*viewsv1.AssertionCheck) []checkpointRow {
	var out []checkpointRow
	for _, c := range checks {
		if !c.GetEntered() {
			continue
		}
		row := checkpointRow{Check: c}
		shown, err1 := kernel.ParseDecimal(c.GetReported())
		derived, err2 := kernel.ParseDecimal(c.GetDerived())
		if err1 == nil && err2 == nil {
			diff := shown.Sub(derived)
			row.Difference = &viewsv1.Money{Amount: diff.String(), Currency: c.GetBook().GetCurrency()}
			switch {
			case diff.Sign() > 0:
				row.Booking = "deposit"
			case diff.Sign() < 0:
				row.Booking = "withdrawal"
			}
		}
		out = append(out, row)
	}
	return out
}

// checkpointFormFor returns the form the account page offers the owner:
// form as posted, or a fresh one for today in the currency of the
// account's first cash balance.
func checkpointFormFor(form *checkpointForm, today kernel.Date, cash []*viewsv1.Money) *checkpointForm {
	if form != nil {
		return form
	}
	f := &checkpointForm{Date: today.String()}
	if len(cash) > 0 {
		f.Currency = cash[0].GetCurrency()
	}
	return f
}

// recordCashCheckpoint hands the cash the owner entered to the core
// (RecordCashCheckpoint, D-158), which stores it as a balance assertion
// and raises an attention item when the derived cash differs. A refusal,
// such as an amount that isn't a decimal, shows the account page again
// with the core's reason and the form as posted.
func (s *server) recordCashCheckpoint(w http.ResponseWriter, r *http.Request, sess session) {
	id := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The cash checkpoint form is malformed.")
		return
	}
	v, ok := postedOnce(r, "date", "amount", "currency")
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The cash checkpoint form is malformed.")
		return
	}
	f := &checkpointForm{Date: v[0], Amount: v[1], Currency: strings.ToUpper(v[2])}
	_, err := s.checkpoints.RecordCashCheckpoint(r.Context(), connect.NewRequest(&ledgerv1.RecordCashCheckpointRequest{
		Origin: sess.Origin, SessionToken: sess.Token, AccountId: id, Date: f.Date, Currency: f.Currency, Amount: f.Amount,
	}))
	s.afterCheckpoint(w, r, sess, "RecordCashCheckpoint", "The cash checkpoint was not recorded: ", f, err)
}

// bookCashDifference hands the owner's booking of a checkpoint's
// difference to the core (BookCashDifference, D-158): a deposit or a
// withdrawal, or cash income when the form says "as" "cash-income"
// (D-152), dated on the checkpoint, after which it matches. A refusal,
// such as a checkpoint that matches already, shows the account page again
// with the core's reason.
func (s *server) bookCashDifference(w http.ResponseWriter, r *http.Request, sess session) {
	id := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The booking form is malformed.")
		return
	}
	v, ok := postedOnce(r, "date", "currency")
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The booking form is malformed.")
		return
	}
	cashIncome, err := bookedAs(r)
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The booking form is malformed.")
		return
	}
	_, err = s.checkpoints.BookCashDifference(r.Context(), connect.NewRequest(&ledgerv1.BookCashDifferenceRequest{
		Origin: sess.Origin, SessionToken: sess.Token, AccountId: id, Date: v[0], Currency: strings.ToUpper(v[1]),
		CashIncome: cashIncome,
	}))
	s.afterCheckpoint(w, r, sess, "BookCashDifference", "The difference was not booked: ", nil, err)
}

// errMalformedBooking reports a booking form whose "as" field is neither
// absent nor "cash-income".
var errMalformedBooking = errors.New("malformed booking form")

// bookedAs reports whether the booking form asks for cash income: its
// "as" field is "cash-income", or absent for the default deposit or
// withdrawal. Any other value, or more than one, is errMalformedBooking.
func bookedAs(r *http.Request) (bool, error) {
	switch as := r.PostForm["as"]; {
	case len(as) == 0:
		return false, nil
	case len(as) == 1 && as[0] == "cash-income":
		return true, nil
	}
	return false, errMalformedBooking
}

// afterCheckpoint answers a checkpoint form: back to the account's cash
// checkpoints on success, the page again with the core's reason on a
// refusal, an error page otherwise.
func (s *server) afterCheckpoint(w http.ResponseWriter, r *http.Request, sess session, method, refused string, f *checkpointForm, err error) {
	if err == nil {
		http.Redirect(w, r, "/accounts/"+url.PathEscape(r.PathValue("id"))+checkpointsAnchor, http.StatusSeeOther)
		return
	}
	if status, ok := refusalStatus(connect.CodeOf(err)); ok {
		s.renderAccount(w, r, sess, status, accountForm{Checkpoint: f, Error: refused + connectMessage(err)})
		return
	}
	s.coreFailed(w, r, method, err)
}

// checkpointPage returns the account page section where the owner books
// the difference of the cash checkpoint an attention item's subject names,
// "account/date/currency" (ledger.CheckpointSubject), the account
// possibly holding slashes itself; empty when the subject is not one.
func checkpointPage(subject string) string {
	i := strings.LastIndex(subject, "/")
	if i < 0 {
		return ""
	}
	j := strings.LastIndex(subject[:i], "/")
	if j <= 0 {
		return ""
	}
	return "/accounts/" + url.PathEscape(subject[:j]) + checkpointsAnchor
}
