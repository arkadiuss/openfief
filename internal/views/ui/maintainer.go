package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"

	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// modelPath is the model portfolio's page (D-52, D-204).
const modelPath = "/model"

// The kinds of attention item the model-portfolio maintainer raises (D-196,
// D-204): a proposal of target changes for the owner to apply or dismiss, and
// a weekly run that keeps failing.
const (
	kindTargetProposal    = "model-target-proposal"
	kindMaintainerFailing = "maintainer-failing"
)

// proposalIDChars is how many hex digits a proposal's ID has: its run's.
const proposalIDChars = 16

// maintainerRoutes maps the model portfolio's page and the owner's ask.
func (s *server) maintainerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+modelPath, s.withSession(s.modelPage))
	mux.HandleFunc("POST "+modelPath+"/maintainer/ask", s.withOwner(s.askMaintainer))
}

// modelTargetRow is one target of the model portfolio as the page lists it.
type modelTargetRow struct {
	Label, Weight string
}

// noteView is the maintainer's latest note as the page shows it.
type noteView struct {
	Text, When, Model, How string
}

// maintainerView is the maintainer's section of the model portfolio page.
type maintainerView struct {
	// Available is true when the maintainer can run: the models switch, its
	// toggle and the model portfolio are on.
	Available bool
	Running   bool
	Note      *noteView
	// Failure says in words how the latest run ended without a note, for the
	// owner.
	Failure string
	// Pending is how many proposals wait in the Inbox, for the owner.
	Pending int
	// CanAsk offers the owner "Ask the maintainer".
	CanAsk bool
}

// modelPageData is what the model portfolio's page shows.
type modelPageData struct {
	Targets  []modelTargetRow
	Currency string
	// Maintainer is nil when the core serves no maintainer or this session
	// may not read its note, which speaks of the whole portfolio.
	Maintainer *maintainerView
	// Unreachable is true when the core could not say how the maintainer
	// stands: the section says so, and the targets still show.
	Unreachable   bool
	Notice, Error string
}

// modelPage shows the model portfolio and the maintainer's latest note. It
// shows nothing while the model portfolio is off (D-52).
func (s *server) modelPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderModel(w, r, sess, http.StatusOK, modelNotice(r.URL.Query()), "")
}

// modelNotice is what the owner's last action did, read from the redirect's
// query.
func modelNotice(q url.Values) string {
	switch q.Get("asked") {
	case "1":
		return "The maintainer was asked. Its note shows here when the run is done."
	case "busy":
		return "A run is already queued or under way."
	}
	return ""
}

// renderModel shows the page with status, a notice and the core's reason for
// refusing the owner's last action, if any.
func (s *server) renderModel(w http.ResponseWriter, r *http.Request, sess session, status int, notice, refusal string) {
	resp, err := s.read.GetModelPortfolio(r.Context(), bearer(&viewsv1.GetModelPortfolioRequest{}, sess))
	switch {
	case connect.CodeOf(err) == connect.CodeNotFound:
		s.errorPage(w, r, http.StatusNotFound, "There is no model portfolio.")
		return
	case err != nil:
		s.coreFailed(w, r, "GetModelPortfolio", err)
		return
	}
	data := modelPageData{Currency: resp.Msg.GetReportingCurrency(), Notice: notice, Error: refusal}
	for _, t := range resp.Msg.GetTargets() {
		data.Targets = append(data.Targets, modelTargetRow{Label: targetLabel(t), Weight: t.GetWeightPercent()})
	}
	data.Maintainer, data.Unreachable = s.maintainerView(r, sess)
	s.show(w, r, status, "model", page{Title: "Model portfolio", Session: &sess, Data: data})
}

// maintainerView reads how the maintainer stands. A core that serves none, and
// a session that may not read the note, give no view; a core that can't say
// gives none and true.
func (s *server) maintainerView(r *http.Request, sess session) (*maintainerView, bool) {
	if s.maintainer == nil {
		return nil, false
	}
	resp, err := s.maintainer.GetMaintainer(r.Context(), connect.NewRequest(&portfoliov1.GetMaintainerRequest{
		Origin: sess.Origin, SessionToken: sess.Token,
	}))
	switch {
	case connect.CodeOf(err) == connect.CodePermissionDenied:
		return nil, false
	case err != nil:
		s.log.WarnContext(r.Context(), "reading the maintainer", "code", connect.CodeOf(err).String())
		return nil, true
	}
	m := resp.Msg
	v := &maintainerView{Available: m.GetAvailable(), Running: m.GetRunning(), Pending: int(m.GetPendingProposals())}
	if n := m.GetNote(); n != nil {
		v.Note = &noteView{Text: n.GetText(), When: whenOf(n.GetCreatedAt()), Model: n.GetModel(), How: triggerWords(n.GetTrigger())}
	}
	if f := m.GetLastFailure(); f != nil {
		v.Failure = "The last run, on " + whenOf(f.GetAt()) + ", ended without a note: " + failureWords(f.GetErrorClass()) + "."
	}
	v.CanAsk = sess.Owner() && v.Available && !v.Running
	return v, false
}

// whenOf writes an RFC 3339 time as a date and a time in UTC; text that isn't
// one is left as it is.
func whenOf(rfc3339 string) string {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return rfc3339
	}
	return t.UTC().Format("2006-01-02 15:04") + " UTC"
}

// triggerWords says what started a run.
func triggerWords(t portfoliov1.RunTrigger) string {
	switch t {
	case portfoliov1.RunTrigger_RUN_TRIGGER_WEEKLY:
		return "the weekly run"
	case portfoliov1.RunTrigger_RUN_TRIGGER_OWNER:
		return "asked by the owner"
	case portfoliov1.RunTrigger_RUN_TRIGGER_UNSPECIFIED:
	}
	return "a run"
}

// failureWords says in words the class a run ended in.
func failureWords(class string) string {
	words := map[string]string{
		"unavailable":    "no model of the route answered in time",
		"rejected":       "the provider rejected the request",
		"canceled":       "the run was cancelled",
		"misconfigured":  "the route names a provider instance that is missing or is not a model provider",
		"internal":       "something broke inside the core",
		"denied":         "the models settings did not allow it: the switch, the maintainer's toggle, a model on its route, or a model allowed for whole-portfolio data (T0) and public data (T2)",
		"invalid_output": "the model's answer was not a usable note",
		"no_figures":     "there were no complete figures to send, since something the portfolio holds is not valued yet",
	}
	if w, ok := words[class]; ok {
		return w
	}
	return "an unknown reason"
}

// askMaintainer hands the owner's "Ask the maintainer" to the core, which
// queues a run, then back to the page saying so, or with the core's reason
// for refusing it, such as a maintainer that is off.
func (s *server) askMaintainer(w http.ResponseWriter, r *http.Request, sess session) {
	if s.maintainer == nil {
		s.errorPage(w, r, http.StatusNotFound, "This core has no maintainer.")
		return
	}
	resp, err := s.maintainer.AskMaintainer(r.Context(), connect.NewRequest(&portfoliov1.AskMaintainerRequest{
		Origin: sess.Origin, SessionToken: sess.Token,
	}))
	if err == nil {
		asked := "1"
		if !resp.Msg.GetStarted() {
			asked = "busy"
		}
		http.Redirect(w, r, modelPath+"?"+url.Values{"asked": {asked}}.Encode(), http.StatusSeeOther)
		return
	}
	if status, refused := refusalStatus(connect.CodeOf(err)); refused && connect.CodeOf(err) != connect.CodePermissionDenied {
		s.renderModel(w, r, sess, status, "", "The core refused this: "+connectMessage(err))
		return
	}
	s.coreFailed(w, r, "AskMaintainer", err)
}

// proposalForm reads the posted proposal ID: the hex ID of a run. ok is false
// for a form that is malformed.
func proposalForm(r *http.Request) (string, bool) {
	if err := r.ParseForm(); err != nil {
		return "", false
	}
	id := r.PostForm.Get("id")
	if len(id) != proposalIDChars || strings.TrimFunc(id, isLowerHex) != "" {
		return "", false
	}
	return id, true
}

// isLowerHex reports whether r is a lower-case hex digit.
func isLowerHex(r rune) bool { return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') }

// applyProposal hands the owner's "Apply" of a proposal to the core, which
// changes the model portfolio's targets, audited, and trades nothing.
func (s *server) applyProposal(w http.ResponseWriter, r *http.Request, sess session) {
	id, ok := proposalForm(r)
	if !ok || s.maintainer == nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	_, err := s.maintainer.ApplyProposal(r.Context(), connect.NewRequest(&portfoliov1.ApplyProposalRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Id: id,
	}))
	if err == nil {
		http.Redirect(w, r, inboxPath+"?"+url.Values{"applied": {"1"}}.Encode(), http.StatusSeeOther)
		return
	}
	s.afterInboxForm(w, r, sess, "ApplyProposal", err)
}

// dismissProposal hands the owner's "Dismiss" of a proposal to the core.
func (s *server) dismissProposal(w http.ResponseWriter, r *http.Request, sess session) {
	id, ok := proposalForm(r)
	if !ok || s.maintainer == nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	_, err := s.maintainer.DismissProposal(r.Context(), connect.NewRequest(&portfoliov1.DismissProposalRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Id: id,
	}))
	if err == nil {
		http.Redirect(w, r, inboxPath+"?"+url.Values{"dismissed": {"1"}}.Encode(), http.StatusSeeOther)
		return
	}
	s.afterInboxForm(w, r, sess, "DismissProposal", err)
}

// proposalEvidence reads the proposal an item's subject, its ID, names, for
// the owner; a core that has no maintainer, or no such proposal, gives none.
func (s *server) proposalEvidence(r *http.Request, sess session, id string, ev *inboxEvidence) error {
	if s.maintainer == nil || !sess.Owner() || id == "" {
		return nil
	}
	if _, ok := ev.proposals[id]; ok {
		return nil
	}
	resp, err := s.maintainer.GetProposal(r.Context(), connect.NewRequest(&portfoliov1.GetProposalRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Id: id,
	}))
	if code := connect.CodeOf(err); err != nil && (code == connect.CodeNotFound || code == connect.CodePermissionDenied) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("GetProposal: %w", err)
	}
	ev.proposals[id] = resp.Msg.GetProposal()
	return nil
}

// proposalTarget names a change's target for the table: an instrument by its
// ISIN and symbol, an index by its name.
func proposalTarget(key string) string {
	if name, ok := strings.CutPrefix(key, "instrument:"); ok {
		return name
	}
	if name, ok := strings.CutPrefix(key, "index:"); ok {
		return name + " (index)"
	}
	return key
}

// driftPoints writes a drift in percentage points with its sign.
func driftPoints(d string) string {
	if d == "" || strings.HasPrefix(d, "-") || d == "0" {
		return d
	}
	return "+" + d
}

// proposalStatusWords says where a proposal stands.
func proposalStatusWords(p *portfoliov1.Proposal) string {
	switch p.GetStatus() {
	case portfoliov1.ProposalStatus_PROPOSAL_STATUS_APPLIED:
		return "Applied on " + whenOf(p.GetDecidedAt()) + ": the model portfolio's targets changed. Nothing was bought or sold."
	case portfoliov1.ProposalStatus_PROPOSAL_STATUS_DISMISSED:
		return "Dismissed on " + whenOf(p.GetDecidedAt()) + ": the model portfolio is as it was."
	case portfoliov1.ProposalStatus_PROPOSAL_STATUS_SUPERSEDED:
		return "Replaced by a later proposal before you decided it."
	case portfoliov1.ProposalStatus_PROPOSAL_STATUS_PENDING, portfoliov1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED:
	}
	return ""
}
