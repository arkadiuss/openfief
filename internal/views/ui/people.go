package ui

import (
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/durationpb"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// inviteMinutes are the lifetimes the owner may give an enrollment link, in
// minutes, the core's default first; the core allows at most 24 hours.
func inviteMinutes() []int { return []int{15, 60, 24 * 60} }

// confirmField is the field of the revoke form that says the owner confirmed.
const (
	confirmField = "confirm"
	confirmYes   = "yes"
)

// personRow is a person as the People section lists them (D-210): the owner,
// or a viewer with what they may view in words, their passkeys and sessions,
// and the form that changes their grants.
type personRow struct {
	ID, Name string
	Owner    bool
	// Revoked reports that the owner revoked the viewer: they stay listed.
	Revoked bool
	// Holds are what the viewer may view, in words; none when they hold no
	// grant. NoAccess is true for an active viewer with none, who sees
	// nothing yet.
	Holds    []string
	NoAccess bool
	Passkeys int
	Sessions int
	// Editable reports whether the grants can change and the viewer be
	// revoked: an active viewer.
	Editable bool
	Choices  grantChoices
	// Action is the path the grants form posts to, RevokeAction the revoke
	// form's.
	Action, RevokeAction string
}

// lifetime is one choice of how long an enrollment link is valid.
type lifetime struct {
	Minutes  int
	Label    string
	Selected bool
}

// inviteForm is the form that invites a viewer: the name, the grants and the
// link's lifetime, as the owner filled them when the core refused them, and,
// once the core issued the link, the link, shown this once.
type inviteForm struct {
	Name    string
	Minutes string
	Grants  []string
	// Choices are the grants form's boxes; Lifetimes the link's choices.
	Choices   grantChoices
	Lifetimes []lifetime
	// Link is the enrollment link just issued, and LinkName the viewer it is
	// for.
	Link     *identityv1.InviteViewerResponse
	LinkName string
}

// peopleRoutes maps the People section's forms, every one the owner's.
func (s *server) peopleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST "+accessPath+"/people", s.withOwner(s.inviteViewer))
	mux.HandleFunc("POST "+accessPath+"/people/{id}/revoke", s.withOwner(s.revokePerson))
	mux.HandleFunc("POST "+accessPath+"/people/{id}/grants", s.withOwner(s.updatePersonGrants))
}

// activeViewer reports whether p is a viewer the owner has not revoked.
func activeViewer(p *identityv1.PersonAccess) bool {
	return p.GetPerson().GetKind() == identityv1.PersonKind_PERSON_KIND_VIEWER && p.GetPerson().GetRevokedAt() == nil
}

// peopleRows returns the people as the section lists them, in the order the
// core gave: the owner first, then the viewers.
func peopleRows(people []*identityv1.PersonAccess, sleeves []*viewsv1.Sleeve, accounts []*viewsv1.Account) []personRow {
	rows := make([]personRow, 0, len(people))
	for _, p := range people {
		id := p.GetPerson().GetId()
		row := personRow{
			ID: id, Name: p.GetPerson().GetName(), Passkeys: int(p.GetPasskeys()), Sessions: int(p.GetActiveSessions()),
			Owner:   p.GetPerson().GetKind() == identityv1.PersonKind_PERSON_KIND_OWNER,
			Revoked: p.GetPerson().GetRevokedAt() != nil,
			Holds:   holdsInWords(p.GetGrants(), sleeves, accounts),
		}
		if activeViewer(p) {
			row.Editable = true
			row.NoAccess = len(row.Holds) == 0
			row.Choices = newChoices("person-"+id, sleeves, accounts, heldValues(p.GetGrants()))
			row.Action = accessPath + "/people/" + url.PathEscape(id) + "/grants"
			row.RevokeAction = accessPath + "/people/" + url.PathEscape(id) + "/revoke"
		}
		rows = append(rows, row)
	}
	return rows
}

// inviteLifetimes returns the lifetimes an enrollment link may have, with the
// default selected.
func inviteLifetimes() []lifetime {
	labels := map[int]string{15: "15 minutes", 60: "1 hour", 24 * 60: "24 hours"}
	out := make([]lifetime, 0, len(inviteMinutes()))
	for i, m := range inviteMinutes() {
		out = append(out, lifetime{Minutes: m, Label: labels[m], Selected: i == 0})
	}
	return out
}

// selectLifetime marks the lifetime named by posted as the selected one, when
// it is one of the choices.
func selectLifetime(choices []lifetime, posted string) []lifetime {
	m, err := strconv.Atoi(posted)
	if err != nil || !slices.ContainsFunc(choices, func(l lifetime) bool { return l.Minutes == m }) {
		return choices
	}
	for i := range choices {
		choices[i].Selected = choices[i].Minutes == m
	}
	return choices
}

// inviteViewer invites a viewer with the name and the grants chosen on the form
// (InviteViewer) and shows the one-time enrollment link with its expiry, this
// once: the core keeps only a hash of its code. The core needs a recent passkey
// check for it (D-183), which the page then asks for and posts the form again.
func (s *server) inviteViewer(w http.ResponseWriter, r *http.Request, sess session) {
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	posted := inviteForm{Name: r.PostFormValue("name"), Minutes: r.PostFormValue("minutes"), Grants: r.PostForm[grantField]}
	minutes := inviteMinutes()[0]
	if posted.Minutes != "" {
		m, err := strconv.Atoi(posted.Minutes)
		if err != nil || !slices.Contains(inviteMinutes(), m) {
			s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
			return
		}
		minutes = m
	}
	targets, ok := parseChoices(posted.Grants)
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	resp, err := s.auth.InviteViewer(r.Context(), connect.NewRequest(&identityv1.InviteViewerRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Name: posted.Name, Grants: targets,
		Ttl: durationpb.New(time.Duration(minutes) * time.Minute),
	}))
	if err != nil {
		s.inviteFailed(w, r, sess, err, &posted)
		return
	}
	s.renderAccess(w, r, sess, http.StatusOK, &accessData{Invite: inviteForm{Link: resp.Msg, LinkName: posted.Name}})
}

// inviteFailed answers an invitation the core did not make: for want of a
// recent passkey check (D-183), the page with the action to confirm and post
// again; for any other refusal, the page with the core's reason and the form as
// the owner filled it.
func (s *server) inviteFailed(w http.ResponseWriter, r *http.Request, sess session, err error, posted *inviteForm) {
	switch {
	case passkeyCheckRequired(err):
		fields := slices.Concat([]formField{{Name: nameField, Value: posted.Name}, {Name: "minutes", Value: posted.Minutes}}, postedFields(r, grantField))
		s.renderAccess(w, r, sess, http.StatusForbidden, &accessData{
			PasskeyCheck: &pendingAction{What: "Inviting a viewer", Action: r.URL.Path, Fields: fields}, Invite: *posted,
		})
	default:
		if status, refused := refusalStatus(connect.CodeOf(err)); refused {
			s.renderAccess(w, r, sess, status, &accessData{Error: connectMessage(err), Invite: *posted})
			return
		}
		s.coreFailed(w, r, "InviteViewer", err)
	}
}

// revokePerson revokes a viewer (RevokePerson) after the owner confirmed: the
// form that asks has no confirmation yet, so it shows the viewer again with
// what revoking does, and nothing reaches the core until the owner confirms. A
// viewer's sessions end, their passkeys stop working and their grants end; the
// person stays listed as revoked. It needs no passkey check: it only takes
// access away.
func (s *server) revokePerson(w http.ResponseWriter, r *http.Request, sess session) {
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	id := r.PathValue("id")
	if r.PostFormValue(confirmField) != confirmYes {
		in, err := s.loadAccessInputs(r, sess)
		if err != nil {
			s.coreFailed(w, r, "access", err)
			return
		}
		rows := peopleRows(in.People.GetPeople(), in.Sleeves, in.Accounts)
		i := slices.IndexFunc(rows, func(p personRow) bool { return p.ID == id && p.Editable })
		if i < 0 {
			s.errorPage(w, r, http.StatusNotFound, "No such viewer, or already revoked.")
			return
		}
		s.renderAccess(w, r, sess, http.StatusOK, &accessData{Confirm: &rows[i]})
		return
	}
	_, err := s.auth.RevokePerson(r.Context(), connect.NewRequest(&identityv1.RevokePersonRequest{
		Origin: sess.Origin, SessionToken: sess.Token, PersonId: id,
	}))
	if err != nil {
		s.accessFailed(w, r, sess, "RevokePerson", err, nil)
		return
	}
	http.Redirect(w, r, accessPath+peopleAnchor, http.StatusSeeOther)
}

// updatePersonGrants makes a viewer's grants the ones posted.
func (s *server) updatePersonGrants(w http.ResponseWriter, r *http.Request, sess session) {
	s.updateGrants(w, r, sess, identityv1.GrantHolderKind_GRANT_HOLDER_KIND_PERSON, r.PathValue("id"), peopleAnchor)
}

// emptyForViewer reports whether sess is a viewer who can view nothing: the
// core lists them no account, because the owner granted them none yet (D-210).
func emptyForViewer(sess session, accounts int) bool { return sess.Viewer() && accounts == 0 }

// nothingShared answers a viewer who can view no account with the page that
// says so, in place of the title page's charts and lists, which would be empty
// or fail: no error, no empty chart.
func (s *server) nothingShared(w http.ResponseWriter, r *http.Request, sess session, title string) {
	s.show(w, r, http.StatusOK, "shared", page{Title: title, Session: &sess})
}
