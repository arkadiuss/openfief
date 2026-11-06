package ui

import (
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// suggestedTransfers counts the suggested transfers, for the overview.
func (s *server) suggestedTransfers(r *http.Request, sess session) (int, error) {
	resp, err := s.read.ListTransfers(r.Context(), bearer(&viewsv1.ListTransfersRequest{}, sess))
	if err != nil {
		return 0, fmt.Errorf("ListTransfers: %w", err)
	}
	return len(resp.Msg.GetSuggested()), nil
}

// confirmTransfer hands the owner's confirmation of the posted pair to the
// core (ConfirmTransfer) and shows the Inbox again.
func (s *server) confirmTransfer(w http.ResponseWriter, r *http.Request, sess session) {
	s.decideTransfer(w, r, sess, true)
}

// rejectTransfer hands the owner's rejection of the posted pair to the
// core (RejectTransfer) and shows the Inbox again.
func (s *server) rejectTransfer(w http.ResponseWriter, r *http.Request, sess session) {
	s.decideTransfer(w, r, sess, false)
}

// decideTransfer posts the decision on the pair the form names. A refusal,
// such as a deposit another confirmed transfer took, shows the Inbox again
// with the core's reason.
func (s *server) decideTransfer(w http.ResponseWriter, r *http.Request, sess session, confirm bool) {
	f, ok := parseTransferForm(r)
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The transfer form is malformed.")
		return
	}
	var err error
	method := "ConfirmTransfer"
	if confirm {
		_, err = s.transfers.ConfirmTransfer(r.Context(), connect.NewRequest(&ledgerv1.ConfirmTransferRequest{
			Origin: sess.Origin, SessionToken: sess.Token, Withdrawal: f.withdrawal, Deposit: f.deposit,
		}))
	} else {
		method = "RejectTransfer"
		_, err = s.transfers.RejectTransfer(r.Context(), connect.NewRequest(&ledgerv1.RejectTransferRequest{
			Origin: sess.Origin, SessionToken: sess.Token, Withdrawal: f.withdrawal, Deposit: f.deposit,
		}))
	}
	if err == nil {
		http.Redirect(w, r, backToInbox(r), http.StatusSeeOther)
		return
	}
	if status, refused := refusalStatus(connect.CodeOf(err)); refused {
		s.renderInbox(w, r, sess, status, "", "The decision was not recorded: "+connectMessage(err))
		return
	}
	s.coreFailed(w, r, method, err)
}

// transferForm is the pair a confirm or reject button posts: its legs'
// transaction keys.
type transferForm struct {
	withdrawal, deposit *ledgerv1.TransactionKey
}

// parseTransferForm reads the posted pair's keys, each field exactly once,
// trimmed and not empty; the core checks every rule.
func parseTransferForm(r *http.Request) (transferForm, bool) {
	if err := r.ParseForm(); err != nil {
		return transferForm{}, false
	}
	names := []string{"withdrawal_source", "withdrawal_id", "deposit_source", "deposit_id"}
	v := make([]string, len(names))
	for i, name := range names {
		if len(r.PostForm[name]) != 1 {
			return transferForm{}, false
		}
		if v[i] = strings.TrimSpace(r.PostForm[name][0]); v[i] == "" {
			return transferForm{}, false
		}
	}
	return transferForm{
		withdrawal: &ledgerv1.TransactionKey{Source: v[0], ExternalId: v[1]},
		deposit:    &ledgerv1.TransactionKey{Source: v[2], ExternalId: v[3]},
	}, true
}
