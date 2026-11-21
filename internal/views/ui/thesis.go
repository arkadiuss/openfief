package ui

import (
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect"

	thesisv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/thesis/v1"
)

// thesesPath is the owner's page of draft theses (D-229): the drafts saved from
// the Ask box, each with its versions. They are the owner's alone and T1 data
// (D-206); no viewer, demo session or script token reaches them.
const thesesPath = "/settings/theses"

// listPreviewRunes is how much of a draft's statement the list shows; the
// draft's own page shows all of it.
const listPreviewRunes = 160

// The words of a draft thesis the page and its notices share.
const (
	thesisHint = "Start an agent: arrives with the control plane."
	// thesisNote says what a draft is and is not (D-229).
	thesisNote = "A draft has no agent, no fief and no lease, and no lease size: it is only text, which only you can read. " +
		"An agent starts from one later, as a proposal that always needs your approval."
)

// thesesData is what the Theses page shows.
type thesesData struct {
	Theses []thesisRow
	Notice string
}

// thesisRow is a draft thesis as the list shows it: its latest version.
type thesisRow struct {
	Href, Statement string
	Horizon         string
	ShortTerm       bool
	Instruments     []string
	Version         int
	Updated         string
}

// thesisData is what a draft thesis's page shows: its latest version, the edit
// form and every version, the newest first.
type thesisData struct {
	Error, Notice string
	ID            string
	Version       int
	Latest        thesisVersionView
	Versions      []thesisVersionView
	Form          draftForm
	Action        string
	Hint          string
	Note          string
}

// thesisVersionView is one saved text of a draft thesis, in words.
type thesisVersionView struct {
	Number       int
	Saved        string
	Current      bool
	Statement    string
	Horizon      string
	ShortTerm    bool
	Instruments  []string
	Invalidation []string
	Sources      []string
}

// draftForm is the edit form's fields: the draft as the owner sees it, one item
// to a line in the lists, as stored or as posted when the core refused it.
type draftForm struct {
	Statement, Horizon                 string
	Instruments, Invalidation, Sources string
	Base                               int
}

// thesesRoutes maps the Theses pages and their forms, every one the owner's.
func (s *server) thesesRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+thesesPath, s.withOwner(s.thesesPage))
	mux.HandleFunc("POST "+thesesPath, s.withOwner(s.saveThesis))
	mux.HandleFunc("GET "+thesesPath+"/{id}", s.withOwner(s.thesisPage))
	mux.HandleFunc("POST "+thesesPath+"/{id}", s.withOwner(s.editThesis))
}

// thesesPage lists the draft theses, the most recently changed first.
func (s *server) thesesPage(w http.ResponseWriter, r *http.Request, sess session) {
	if s.theses == nil {
		s.errorPage(w, r, http.StatusNotFound, "This core serves no draft theses.")
		return
	}
	got, err := s.theses.ListTheses(r.Context(), connect.NewRequest(&thesisv1.ListThesesRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	if err != nil {
		s.coreFailed(w, r, "ListTheses", err)
		return
	}
	data := &thesesData{}
	for _, t := range got.Msg.GetTheses() {
		d := t.GetDraft()
		data.Theses = append(data.Theses, thesisRow{
			Href: thesesPath + "/" + t.GetId(), Statement: preview(d.GetStatement()), Horizon: horizonWords(d.GetHorizon()),
			ShortTerm: d.GetHorizon() == thesisv1.Horizon_HORIZON_SHORT_TERM, Instruments: d.GetInstruments(),
			Version: int(t.GetVersion()), Updated: when(t.GetUpdatedAt().AsTime()),
		})
	}
	s.show(w, r, http.StatusOK, "theses", page{Title: "Theses", Session: &sess, Data: data})
}

// preview cuts text to listPreviewRunes characters, marking the cut.
func preview(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if utf8.RuneCountInString(text) <= listPreviewRunes {
		return text
	}
	return strings.TrimSpace(string([]rune(text)[:listPreviewRunes])) + "…"
}

// when writes a time as the pages write it: the day and the minute, in UTC.
func when(t time.Time) string { return t.UTC().Format("2006-01-02 15:04") + " UTC" }

// horizonWords says a horizon as the page does.
func horizonWords(h thesisv1.Horizon) string {
	if h == thesisv1.Horizon_HORIZON_SHORT_TERM {
		return "Short-term"
	}
	return "Long-term"
}

// thesisPage shows a draft thesis, its edit form and its versions.
func (s *server) thesisPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderThesis(w, r, sess, http.StatusOK, r.PathValue("id"), nil, "")
}

// renderThesis reads the thesis id and shows it with status; form, when it is
// not nil, is what the edit form holds instead of the latest version, and
// refusal is the core's reason for refusing the owner's edit.
func (s *server) renderThesis(w http.ResponseWriter, r *http.Request, sess session, status int, id string, form *draftForm, refusal string) {
	if s.theses == nil {
		s.errorPage(w, r, http.StatusNotFound, "This core serves no draft theses.")
		return
	}
	got, err := s.theses.GetThesis(r.Context(), connect.NewRequest(&thesisv1.GetThesisRequest{Origin: sess.Origin, SessionToken: sess.Token, ThesisId: id}))
	if err != nil {
		s.thesisFailed(w, r, "GetThesis", err)
		return
	}
	t := got.Msg.GetThesis()
	data := &thesisData{ID: t.GetId(), Action: thesesPath + "/" + t.GetId(), Hint: thesisHint, Note: thesisNote, Error: refusal}
	vs := t.GetVersions()
	for i, v := range slices.Backward(vs) {
		data.Versions = append(data.Versions, versionView(v, i == len(vs)-1))
	}
	if len(data.Versions) > 0 {
		data.Latest, data.Version = data.Versions[0], data.Versions[0].Number
	}
	data.Form = draftFormOf(&data.Latest)
	if form != nil {
		data.Form = *form
	}
	data.Form.Base = data.Version
	if saved := r.URL.Query().Get("saved"); saved == "1" {
		data.Notice = "Saved as a draft thesis, version 1. It is entered in the audit log, never its text."
	} else if saved == "edit" {
		data.Notice = "Saved as version " + strconv.Itoa(data.Version) + ". The earlier versions are kept below, and the change is entered in the audit log, never its text."
	}
	s.show(w, r, status, "thesis", page{Title: "Draft thesis", Session: &sess, Data: data})
}

// versionView describes a version in words.
func versionView(v *thesisv1.ThesisVersion, current bool) thesisVersionView {
	d := v.GetDraft()
	return thesisVersionView{
		Number: int(v.GetVersion()), Saved: when(v.GetSavedAt().AsTime()), Current: current, Statement: d.GetStatement(),
		Horizon: horizonWords(d.GetHorizon()), ShortTerm: d.GetHorizon() == thesisv1.Horizon_HORIZON_SHORT_TERM,
		Instruments: d.GetInstruments(), Invalidation: d.GetInvalidation(), Sources: d.GetSources(),
	}
}

// draftFormOf fills the edit form with a version's text, a line to an item.
func draftFormOf(v *thesisVersionView) draftForm {
	horizon := horizonLongTerm
	if v.ShortTerm {
		horizon = horizonShortTerm
	}
	return draftForm{
		Statement: v.Statement, Horizon: horizon, Instruments: strings.Join(v.Instruments, "\n"),
		Invalidation: strings.Join(v.Invalidation, "\n"), Sources: strings.Join(v.Sources, "\n"),
	}
}

// postedDraft reads a draft from a posted form: the statement, the horizon, and
// each list as the form's values, each of which may itself hold several lines.
func postedDraft(r *http.Request) (*thesisv1.ThesisDraft, error) {
	if err := r.ParseForm(); err != nil {
		return nil, errBadForm
	}
	f := r.PostForm
	d := &thesisv1.ThesisDraft{
		Statement: f.Get("statement"), Instruments: postedLines(f["instruments"]), Invalidation: postedLines(f["invalidation"]),
		Sources: postedLines(f["sources"]),
	}
	switch f.Get("horizon") {
	case "long_term":
		d.Horizon = thesisv1.Horizon_HORIZON_LONG_TERM
	case "short_term":
		d.Horizon = thesisv1.Horizon_HORIZON_SHORT_TERM
	}
	return d, nil
}

// postedLines splits each value into its lines and drops the empty ones.
func postedLines(values []string) []string {
	var out []string
	for _, v := range values {
		for line := range strings.SplitSeq(strings.ReplaceAll(v, "\r", "\n"), "\n") {
			if line = strings.TrimSpace(line); line != "" {
				out = append(out, line)
			}
		}
	}
	return out
}

// draftFormOfPosted is the edit form as the owner posted d.
func draftFormOfPosted(d *thesisv1.ThesisDraft) draftForm {
	horizon := ""
	switch d.GetHorizon() {
	case thesisv1.Horizon_HORIZON_LONG_TERM:
		horizon = "long_term"
	case thesisv1.Horizon_HORIZON_SHORT_TERM:
		horizon = "short_term"
	case thesisv1.Horizon_HORIZON_UNSPECIFIED:
	}
	return draftForm{
		Statement: d.GetStatement(), Horizon: horizon, Instruments: strings.Join(d.GetInstruments(), "\n"),
		Invalidation: strings.Join(d.GetInvalidation(), "\n"), Sources: strings.Join(d.GetSources(), "\n"),
	}
}

// thesisSaved is what the page says after the Ask box's card saved a draft.
type thesisSaved struct {
	Href  string
	Error string
}

// saveThesis saves the draft the Ask box's card posts as a new thesis. The card
// posts with htmx, so the conversation under it stays, and the page answers with
// the card's new content; without htmx it redirects to the draft.
func (s *server) saveThesis(w http.ResponseWriter, r *http.Request, sess session) {
	draft, err := postedDraft(r)
	if err != nil || s.theses == nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	got, err := s.theses.SaveThesis(r.Context(), connect.NewRequest(&thesisv1.SaveThesisRequest{Origin: sess.Origin, SessionToken: sess.Token, Draft: draft}))
	if err != nil {
		if status, refused := refusalStatus(connect.CodeOf(err)); refused && isHTMX(r) {
			s.showTemplate(w, r, status, "overview", "thesis-saved", thesisSaved{Error: "The core refused this draft: " + connectMessage(err)})
			return
		}
		s.thesisFailed(w, r, "SaveThesis", err)
		return
	}
	href := thesesPath + "/" + got.Msg.GetThesis().GetId()
	if isHTMX(r) {
		s.showTemplate(w, r, http.StatusOK, "overview", "thesis-saved", thesisSaved{Href: href})
		return
	}
	http.Redirect(w, r, href+"?saved=1", http.StatusSeeOther)
}

// editThesis saves the owner's edit as the next version of a draft thesis,
// keeping the earlier ones. An edit of a version that is no longer the latest,
// or one the core refuses, shows the page again with the owner's text and the
// reason.
func (s *server) editThesis(w http.ResponseWriter, r *http.Request, sess session) {
	draft, err := postedDraft(r)
	base, baseErr := strconv.Atoi(r.PostForm.Get("base_version"))
	if err != nil || baseErr != nil || s.theses == nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	id := r.PathValue("id")
	_, err = s.theses.EditThesis(r.Context(), connect.NewRequest(&thesisv1.EditThesisRequest{
		Origin: sess.Origin, SessionToken: sess.Token, ThesisId: id, BaseVersion: int32(base), Draft: draft, //nolint:gosec // a version the form carried, which the core checks
	}))
	if err == nil {
		http.Redirect(w, r, thesesPath+"/"+url.PathEscape(id)+"?saved=edit", http.StatusSeeOther)
		return
	}
	code := connect.CodeOf(err)
	if code == connect.CodeInvalidArgument || code == connect.CodeFailedPrecondition {
		form := draftFormOfPosted(draft)
		reason := "The core refused this edit: " + connectMessage(err)
		if code == connect.CodeFailedPrecondition {
			reason = "This draft has a newer version, saved since you opened it. Your text is below; open the draft again to edit the latest."
		}
		s.renderThesis(w, r, sess, http.StatusBadRequest, id, &form, reason)
		return
	}
	s.thesisFailed(w, r, "EditThesis", err)
}

// thesisFailed answers a failed call about a draft thesis: a draft that does
// not exist is a 404, anything else as every failed call is answered.
func (s *server) thesisFailed(w http.ResponseWriter, r *http.Request, method string, err error) {
	if connect.CodeOf(err) == connect.CodeNotFound {
		s.errorPage(w, r, http.StatusNotFound, "There is no such draft thesis.")
		return
	}
	s.coreFailed(w, r, method, err)
}
