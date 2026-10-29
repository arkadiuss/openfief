package ui

import (
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// Where the sleeve and account forms return to: their own pages, or a
// step of the first-run wizard (D-73).
const (
	sleevesPath   = "/sleeves"
	accountsPath  = "/accounts"
	setupSleeves  = "/setup?step=sleeves"
	setupAccounts = "/setup?step=accounts"
)

// returnTo is the page a form posted next asks to return to, when it is
// one of the pages the forms are on; any other value is fallback, so a
// form can never send the browser elsewhere.
func returnTo(next, fallback string) string {
	switch next {
	case sleevesPath:
		return sleevesPath
	case accountsPath:
		return accountsPath
	case setupSleeves:
		return setupSleeves
	case setupAccounts:
		return setupAccounts
	}
	return fallback
}

// sleevesData is what the sleeve forms show: the kinds and sleeves, the
// page the forms return to, and the core's reason when it refused one.
type sleevesData struct {
	Owner   bool
	Kinds   []*viewsv1.SleeveKind
	Sleeves []*viewsv1.Sleeve
	Next    string
	Error   string
}

// KindName returns the name of the kind key, or the key when it is not
// listed.
func (d *sleevesData) KindName(key string) string {
	for _, k := range d.Kinds {
		if k.GetKey() == key {
			return k.GetName()
		}
	}
	return key
}

// KindRole returns the default role of the kind key, as a word.
func (d *sleevesData) KindRole(key string) string {
	for _, k := range d.Kinds {
		if k.GetKey() == key {
			return roleWord(k.GetDefaultRole())
		}
	}
	return ""
}

// roleWord names a role in lower case: defensive or offensive.
func roleWord(r viewsv1.SleeveRole) string {
	return strings.ToLower(enumLabel(r.String(), "SLEEVE_ROLE_"))
}

// loadSleeves reads the kinds and sleeves for the sleeve forms.
func (s *server) loadSleeves(r *http.Request, sess session, next string) (*sleevesData, error) {
	kinds, err := s.read.ListSleeveKinds(r.Context(), bearer(&viewsv1.ListSleeveKindsRequest{}, sess))
	if err != nil {
		return nil, fmt.Errorf("ListSleeveKinds: %w", err)
	}
	sleeves, err := s.read.ListSleeves(r.Context(), bearer(&viewsv1.ListSleevesRequest{}, sess))
	if err != nil {
		return nil, fmt.Errorf("ListSleeves: %w", err)
	}
	return &sleevesData{Owner: sess.Owner(), Kinds: kinds.Msg.GetKinds(), Sleeves: sleeves.Msg.GetSleeves(), Next: next}, nil
}

// sleevesPage shows the sleeve kinds and sleeves (D-71), with the owner's
// forms to add and change them.
func (s *server) sleevesPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderSleeves(w, r, sess, http.StatusOK, sleevesPath, "")
}

// renderSleeves shows the sleeve forms on the page next names, the sleeves
// page or the wizard's step, with message as the core's refusal.
func (s *server) renderSleeves(w http.ResponseWriter, r *http.Request, sess session, status int, next, message string) {
	if strings.HasPrefix(next, "/setup") {
		s.renderSetup(w, r, sess, status, stepSleeves, message)
		return
	}
	data, err := s.loadSleeves(r, sess, sleevesPath)
	if err != nil {
		s.coreFailed(w, r, "ListSleeves", err)
		return
	}
	data.Error = message
	s.show(w, r, status, "sleeves", page{Title: "Sleeves", Session: &sess, Data: data})
}

// sleeveEntryDone sends the browser back where the form came from, or, when
// the core refused the entry, shows the form's page again with the reason.
func (s *server) sleeveEntryDone(w http.ResponseWriter, r *http.Request, sess session, method string, err error) {
	next := returnTo(r.PostFormValue("next"), sleevesPath)
	if err == nil {
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}
	if status, refused := refusalStatus(connect.CodeOf(err)); refused {
		s.renderSleeves(w, r, sess, status, next, connectMessage(err))
		return
	}
	s.coreFailed(w, r, method, err)
}

// formRole reads a role field: defensive, offensive, or anything else for
// the kind's default role.
func formRole(v string) portfoliov1.SleeveRole {
	switch v {
	case "defensive":
		return portfoliov1.SleeveRole_SLEEVE_ROLE_DEFENSIVE
	case "offensive":
		return portfoliov1.SleeveRole_SLEEVE_ROLE_OFFENSIVE
	}
	return portfoliov1.SleeveRole_SLEEVE_ROLE_UNSPECIFIED
}

// roleField is what a sleeve form's role fields show: the form's ID prefix,
// the role chosen explicitly, "" for the kind's default, and the override.
type roleField struct {
	ID       string
	Role     string
	Override bool
}

// roleFields returns the role fields of a form with id, for sleeve, or for
// a new sleeve when it is nil: an explicit role only where the sleeve's
// role differs from its kind's.
func roleFields(id string, sleeve *viewsv1.Sleeve) roleField {
	f := roleField{ID: id}
	if sleeve != nil && sleeve.GetRoleOverride() {
		f.Role, f.Override = roleWord(sleeve.GetRole()), true
	}
	return f
}

// addKind hands a new sleeve kind to the core (AddSleeveKind).
func (s *server) addKind(w http.ResponseWriter, r *http.Request, sess session) {
	_, err := s.sleeves.AddSleeveKind(r.Context(), connect.NewRequest(&portfoliov1.AddSleeveKindRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Name: strings.TrimSpace(r.PostFormValue("name")),
		DefaultRole: formRole(r.PostFormValue("default_role")),
	}))
	s.sleeveEntryDone(w, r, sess, "AddSleeveKind", err)
}

// renameKind hands a kind's new name to the core (RenameSleeveKind).
func (s *server) renameKind(w http.ResponseWriter, r *http.Request, sess session) {
	_, err := s.sleeves.RenameSleeveKind(r.Context(), connect.NewRequest(&portfoliov1.RenameSleeveKindRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Key: r.PathValue("key"), Name: strings.TrimSpace(r.PostFormValue("name")),
	}))
	s.sleeveEntryDone(w, r, sess, "RenameSleeveKind", err)
}

// addSleeve hands a new sleeve to the core (AddSleeve). A role other than
// the kind's default needs the override box ticked: the owner's explicit
// choice (D-71).
func (s *server) addSleeve(w http.ResponseWriter, r *http.Request, sess session) {
	_, err := s.sleeves.AddSleeve(r.Context(), connect.NewRequest(&portfoliov1.AddSleeveRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Name: strings.TrimSpace(r.PostFormValue("name")),
		Kind: r.PostFormValue("kind"), Role: formRole(r.PostFormValue("role")), RoleOverride: r.PostFormValue("role_override") == "on",
	}))
	s.sleeveEntryDone(w, r, sess, "AddSleeve", err)
}

// updateSleeve hands a sleeve's new name, kind or role to the core
// (UpdateSleeve).
func (s *server) updateSleeve(w http.ResponseWriter, r *http.Request, sess session) {
	_, err := s.sleeves.UpdateSleeve(r.Context(), connect.NewRequest(&portfoliov1.UpdateSleeveRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Id: r.PathValue("id"), Name: strings.TrimSpace(r.PostFormValue("name")),
		Kind: r.PostFormValue("kind"), Role: formRole(r.PostFormValue("role")), RoleOverride: r.PostFormValue("role_override") == "on",
	}))
	s.sleeveEntryDone(w, r, sess, "UpdateSleeve", err)
}
