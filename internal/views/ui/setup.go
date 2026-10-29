package ui

import (
	"net/http"
)

// The first-run wizard's steps (D-73), in order.
const (
	stepSleeves  = "sleeves"
	stepAccounts = "accounts"
	stepDone     = "done"
)

// setupData is what a step of the first-run wizard shows: the step, and
// the sleeve or account forms it holds.
type setupData struct {
	Step     string
	Sleeves  *sleevesData
	Accounts *accountsData
}

// Steps lists the wizard's steps for its progress list, each with whether
// it is the current one.
func (d *setupData) Steps() []setupStep {
	steps := []setupStep{
		{Key: stepSleeves, Label: "Sleeves"},
		{Key: stepAccounts, Label: "Accounts and connectors"},
		{Key: stepDone, Label: "Done"},
	}
	for i := range steps {
		steps[i].Current = steps[i].Key == d.Step
	}
	return steps
}

// setupStep is one step in the wizard's progress list.
type setupStep struct {
	Key, Label string
	Current    bool
}

// setupPage shows a step of the first-run wizard (D-73): the sleeves the
// install started with, which the owner may rename, change or add to; then
// the accounts, each with the connector instance already installed that
// feeds it and the broker account it is; then where to go next. Installing
// plugins stays in the CLI (D-46).
func (s *server) setupPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderSetup(w, r, sess, http.StatusOK, r.URL.Query().Get("step"), "")
}

// renderSetup shows the wizard's step, with message as the core's refusal
// of the step's last form. An unknown step is the first.
func (s *server) renderSetup(w http.ResponseWriter, r *http.Request, sess session, status int, step, message string) {
	data := &setupData{Step: step}
	var err error
	switch step {
	case stepAccounts:
		if data.Accounts, err = s.loadAccounts(r, sess, setupAccounts); err != nil {
			s.coreFailed(w, r, "ListAccounts", err)
			return
		}
		data.Accounts.Error = message
	case stepDone:
	default:
		data.Step = stepSleeves
		if data.Sleeves, err = s.loadSleeves(r, sess, setupSleeves); err != nil {
			s.coreFailed(w, r, "ListSleeves", err)
			return
		}
		data.Sleeves.Error = message
	}
	s.show(w, r, status, "setup", page{Title: "Set up OpenFief", Session: &sess, Data: data})
}

// onboardingRoutes maps the paths of the first-run wizard (D-73) and of the
// sleeve and account pages (D-71, D-72): anyone with a session may look,
// only the owner may change.
func (s *server) onboardingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /setup", s.withOwner(s.setupPage))
	mux.HandleFunc("GET /sleeves", s.withSession(s.sleevesPage))
	mux.HandleFunc("POST /sleeves", s.withOwner(s.addSleeve))
	mux.HandleFunc("POST /sleeves/{id}", s.withOwner(s.updateSleeve))
	mux.HandleFunc("POST /sleeve-kinds", s.withOwner(s.addKind))
	mux.HandleFunc("POST /sleeve-kinds/{key}", s.withOwner(s.renameKind))
	mux.HandleFunc("GET /accounts", s.withSession(s.accountsPage))
	mux.HandleFunc("POST /accounts", s.withOwner(s.createAccount))
	mux.HandleFunc("POST /accounts/bind", s.withOwner(s.bindAccount))
	mux.HandleFunc("POST /accounts/{id}/edit", s.withOwner(s.updateAccount))
}

// firstRun reports whether an owner session sees an install that has no
// account yet: the overview then sends the owner to the first-run wizard,
// which is where the owner lands after creating the first passkey (D-73).
func firstRun(sess session, accounts int) bool {
	return sess.Owner() && accounts == 0
}
