package ui

import (
	"errors"
	"net/http"
	"strconv"

	"connectrpc.com/connect"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

const (
	// defaultTokenDays is a new script token's lifetime when the form
	// leaves it out: the core's default, 30 days.
	defaultTokenDays = 30
	// maxTokenDays is the longest a script token lives (D-27).
	maxTokenDays = 365
	// hoursPerDay turns days into a duration.
	hoursPerDay = 24
)

// settingsData is what the settings page shows: the owner's passkeys,
// number format and marginal PIT rate, model portfolio and ranking of
// market-data sources, the installed plugins, and once, right after it was
// made, a new enrollment link. Script tokens are on the Access page.
type settingsData struct {
	Origin     string
	Passkeys   []*identityv1.Passkey
	Enrollment *identityv1.CreateEnrollmentLinkResponse
	Error      string
	// PasskeyCheck is an action the core refused for want of a recent
	// passkey check (D-183), to post again once the owner confirms.
	PasskeyCheck *pendingAction
	// Model is the model portfolio form (D-52): as stored, or as posted
	// when the core refused it.
	Model *modelForm
	// Plugins are the installed plugins, shown read-only (D-46, D-56).
	Plugins pluginList
	// MarketData is the owner's ranking of market-data sources, one per
	// kind, and Series each series' source (D-94).
	MarketData []rankingView
	Series     []seriesView
	// Values are the owner's number format and marginal PIT rate (D-78,
	// D-113).
	Values *valuesSettings
	// Taxes are this year's yearly taxes and the classification form
	// (D-127, D-132).
	Taxes *taxesSettings
}

// settingsPage shows the owner's settings.
func (s *server) settingsPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderSettings(w, r, sess, http.StatusOK, &settingsData{})
}

// renderSettings loads the passkeys and the rest of the settings and shows
// them with whatever the last action left in data.
func (s *server) renderSettings(w http.ResponseWriter, r *http.Request, sess session, status int, data *settingsData) {
	keys, err := s.auth.ListPasskeys(r.Context(), connect.NewRequest(&identityv1.ListPasskeysRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	if err != nil {
		s.coreFailed(w, r, "ListPasskeys", err)
		return
	}
	plugins, err := s.loadPlugins(r, sess)
	if err != nil {
		s.coreFailed(w, r, "ListPlugins", err)
		return
	}
	data.Plugins = plugins
	if data.Values, err = s.loadValuesSettings(r, sess); err != nil {
		s.coreFailed(w, r, "GetOwnerSettings", err)
		return
	}
	if data.Taxes, err = s.loadTaxes(r, sess); err != nil {
		s.coreFailed(w, r, "ListYearlyTaxes", err)
		return
	}
	if data.MarketData, err = s.loadRankings(r, sess); err != nil {
		s.coreFailed(w, r, "ListSourceRankings", err)
		return
	}
	if data.Series, err = s.loadSeries(r, sess); err != nil {
		s.coreFailed(w, r, "ListSeriesSources", err)
		return
	}
	if data.Model == nil {
		form, err := s.loadModel(r, sess)
		if err != nil {
			s.coreFailed(w, r, "GetModelPortfolio", err)
			return
		}
		data.Model = form.withBlankRows()
	}
	data.Origin, data.Passkeys = sess.Origin, keys.Msg.GetPasskeys()
	s.show(w, r, status, "settings", page{Title: "Settings", Session: &sess, Data: data})
}

// actionFailed shows the settings again with the core's reason when it
// refused an action, and otherwise answers like any failed call.
func (s *server) actionFailed(w http.ResponseWriter, r *http.Request, sess session, method string, err error) {
	if status, refused := refusalStatus(connect.CodeOf(err)); refused {
		s.renderSettings(w, r, sess, status, &settingsData{Error: connectMessage(err)})
		return
	}
	s.coreFailed(w, r, method, err)
}

// pendingAction is a settings action the core refused for want of a
// passkey check made in the last five minutes (D-183). The settings page
// asks the owner to confirm with their passkey, runs the check, and posts
// the action again as it was.
type pendingAction struct {
	// What names the action for the owner, such as "Creating a script
	// token".
	What string
	// Action is the path the form posts to again.
	Action string
	// Fields are the form's fields, posted again unchanged.
	Fields []formField
}

// formField is one field of a pendingAction's form.
type formField struct {
	Name  string
	Value string
}

// passkeyCheckRequired reports whether err is the core's refusal of an
// action D-183 protects: FAILED_PRECONDITION with the PasskeyCheckRequired
// detail.
func passkeyCheckRequired(err error) bool {
	ce, ok := errors.AsType[*connect.Error](err)
	if !ok || ce.Code() != connect.CodeFailedPrecondition {
		return false
	}
	for _, d := range ce.Details() {
		msg, valueErr := d.Value()
		if valueErr != nil {
			continue
		}
		if _, isCheck := msg.(*identityv1.PasskeyCheckRequired); isCheck {
			return true
		}
	}
	return false
}

// protectedFailed answers a refused action D-183 protects: for want of a
// recent passkey check, the settings again with pending to confirm and
// post again; anything else as actionFailed does.
func (s *server) protectedFailed(w http.ResponseWriter, r *http.Request, sess session, method string, err error, pending *pendingAction) {
	if passkeyCheckRequired(err) {
		s.renderSettings(w, r, sess, http.StatusForbidden, &settingsData{PasskeyCheck: pending})
		return
	}
	s.actionFailed(w, r, sess, method, err)
}

// removePasskey removes a passkey (RemovePasskey); the core refuses to
// remove the last one, and revokes every session the passkey signed in
// (D-183). When that includes this one, the cookie goes and the login page
// opens.
func (s *server) removePasskey(w http.ResponseWriter, r *http.Request, sess session) {
	resp, err := s.auth.RemovePasskey(r.Context(), connect.NewRequest(&identityv1.RemovePasskeyRequest{
		Origin: sess.Origin, SessionToken: sess.Token, PasskeyId: r.PathValue("id"),
	}))
	if err != nil {
		s.protectedFailed(w, r, sess, "RemovePasskey", err, &pendingAction{What: "Removing a passkey", Action: r.URL.Path})
		return
	}
	if resp.Msg.GetSignedOut() {
		clearSession(w)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

// tokenDays reads a token lifetime in days: empty is the default, anything
// else a whole number from 1 to maxTokenDays.
func tokenDays(raw string) (int, bool) {
	if raw == "" {
		return defaultTokenDays, true
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days < 1 || days > maxTokenDays {
		return 0, false
	}
	return days, true
}

// createEnrollmentLink issues a one-time link for creating a passkey on
// another allowed origin, or on this one when the form names none (D-60),
// and shows it this once.
func (s *server) createEnrollmentLink(w http.ResponseWriter, r *http.Request, sess session) {
	resp, err := s.auth.CreateEnrollmentLink(r.Context(), connect.NewRequest(&identityv1.CreateEnrollmentLinkRequest{
		Origin: sess.Origin, SessionToken: sess.Token, TargetOrigin: r.PostFormValue("target_origin"),
	}))
	if err != nil {
		s.protectedFailed(w, r, sess, "CreateEnrollmentLink", err, &pendingAction{
			What: "Creating an enrollment link", Action: r.URL.Path,
			Fields: []formField{{Name: "target_origin", Value: r.PostFormValue("target_origin")}},
		})
		return
	}
	s.renderSettings(w, r, sess, http.StatusOK, &settingsData{Enrollment: resp.Msg})
}
