package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

// The pages of the audit log (D-211): Settings' own page, which shows every
// entry to the owner and filters them, and the history on the Access page,
// which shows the grants and the links (D-210). Both are the owner's alone.
const (
	auditPath = "/settings/audit"
	// historyAnchor is the history section on the Access page.
	historyAnchor = "#history-title"
	// auditPageSize and historyPageSize are the entries a page of each shows.
	auditPageSize   = 50
	historyPageSize = 15
	// idChars is how many characters of a long ID a row shows; the record's
	// name replaces it where the record is known.
	idChars = 8
	// dayFormat is how the filter's dates are written and read, in UTC.
	dayFormat = "2006-01-02"
	// whenFormat is how an entry's time is shown, in UTC.
	whenFormat = "2006-01-02 15:04:05 MST"
	// beforeParam is the query field of the cursor a page hands out.
	beforeParam = "before"
	// historyBeforeParam is the cursor of the Access page's history.
	historyBeforeParam = "history_before"
	// unknownWord names an action or an actor the page doesn't know; toneNeutral
	// is the tone of an outcome that is neither good nor bad.
	unknownWord = "Unknown"
	toneNeutral = "neutral"
)

// historyActions are the entries the Access page's history lists: every grant
// and revocation, of people, script tokens and view plugins, and the links
// that invite a viewer, with when the viewer was created and revoked.
func historyActions() []identityv1.AuditAction {
	return []identityv1.AuditAction{
		identityv1.AuditAction_AUDIT_ACTION_GRANT_CREATED, identityv1.AuditAction_AUDIT_ACTION_GRANT_REVOKED,
		identityv1.AuditAction_AUDIT_ACTION_PERSON_CREATED, identityv1.AuditAction_AUDIT_ACTION_PERSON_REVOKED,
		identityv1.AuditAction_AUDIT_ACTION_SCRIPT_TOKEN_CREATED, identityv1.AuditAction_AUDIT_ACTION_SCRIPT_TOKEN_REVOKED,
		identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_ISSUED, identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_USED,
		identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_CANCELLED,
	}
}

// actionWords says what an entry's action is, in words.
func actionWords(a identityv1.AuditAction) string {
	words := map[identityv1.AuditAction]string{
		identityv1.AuditAction_AUDIT_ACTION_SIGN_IN:                       "Sign-in",
		identityv1.AuditAction_AUDIT_ACTION_SIGN_OUT:                      "Sign-out",
		identityv1.AuditAction_AUDIT_ACTION_SIGN_OUT_EVERYWHERE:           "Signed out everywhere",
		identityv1.AuditAction_AUDIT_ACTION_PASSKEY_ADDED:                 "Passkey added",
		identityv1.AuditAction_AUDIT_ACTION_PASSKEY_REMOVED:               "Passkey removed",
		identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_ISSUED:        "Enrollment link issued",
		identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_USED:          "Enrollment link used",
		identityv1.AuditAction_AUDIT_ACTION_ENROLLMENT_LINK_CANCELLED:     "Invitation cancelled",
		identityv1.AuditAction_AUDIT_ACTION_SCRIPT_TOKEN_CREATED:          "Script token created",
		identityv1.AuditAction_AUDIT_ACTION_SCRIPT_TOKEN_REVOKED:          "Script token revoked",
		identityv1.AuditAction_AUDIT_ACTION_GRANT_CREATED:                 "Access granted",
		identityv1.AuditAction_AUDIT_ACTION_GRANT_REVOKED:                 "Access ended",
		identityv1.AuditAction_AUDIT_ACTION_MODEL_CALL:                    "Model called",
		identityv1.AuditAction_AUDIT_ACTION_MODEL_CALL_RESULT:             "Model call ended",
		identityv1.AuditAction_AUDIT_ACTION_PERSON_CREATED:                "Viewer created",
		identityv1.AuditAction_AUDIT_ACTION_PERSON_REVOKED:                "Viewer revoked",
		identityv1.AuditAction_AUDIT_ACTION_MAINTAINER_RUN:                "Maintainer run",
		identityv1.AuditAction_AUDIT_ACTION_MAINTAINER_PROPOSAL_DROPPED:   "Maintainer proposal dropped",
		identityv1.AuditAction_AUDIT_ACTION_MAINTAINER_PROPOSAL_APPLIED:   "Maintainer proposal applied",
		identityv1.AuditAction_AUDIT_ACTION_MAINTAINER_PROPOSAL_DISMISSED: "Maintainer proposal dismissed",
		identityv1.AuditAction_AUDIT_ACTION_MODEL_SETTINGS_CHANGED:        "Models settings changed",
	}
	if w, ok := words[a]; ok {
		return w
	}
	return "Unknown action"
}

// actorKindWords names a kind of actor in words, for the filter.
func actorKindWords(k identityv1.AuditActorKind) string {
	switch k {
	case identityv1.AuditActorKind_AUDIT_ACTOR_KIND_OWNER:
		return "Owner"
	case identityv1.AuditActorKind_AUDIT_ACTOR_KIND_VIEWER:
		return "A viewer"
	case identityv1.AuditActorKind_AUDIT_ACTOR_KIND_ENROLLMENT_CODE:
		return "An enrollment link"
	case identityv1.AuditActorKind_AUDIT_ACTOR_KIND_SYSTEM:
		return "The core"
	case identityv1.AuditActorKind_AUDIT_ACTOR_KIND_ANONYMOUS:
		return "Nobody signed in"
	case identityv1.AuditActorKind_AUDIT_ACTOR_KIND_UNSPECIFIED:
	}
	return unknownWord
}

// outcomeWord names an outcome in words.
func outcomeWord(o identityv1.AuditOutcome) string {
	switch o {
	case identityv1.AuditOutcome_AUDIT_OUTCOME_OK:
		return "ok"
	case identityv1.AuditOutcome_AUDIT_OUTCOME_REFUSED:
		return "refused"
	case identityv1.AuditOutcome_AUDIT_OUTCOME_FAILED:
		return "failed"
	case identityv1.AuditOutcome_AUDIT_OUTCOME_SENT:
		return "sent"
	case identityv1.AuditOutcome_AUDIT_OUTCOME_UNSPECIFIED:
	}
	return "unknown"
}

// outcomeTone is the tone of an outcome's pill: ok, warn, danger or neutral.
func outcomeTone(o identityv1.AuditOutcome) string {
	switch o {
	case identityv1.AuditOutcome_AUDIT_OUTCOME_OK:
		return "ok"
	case identityv1.AuditOutcome_AUDIT_OUTCOME_REFUSED:
		return "warn"
	case identityv1.AuditOutcome_AUDIT_OUTCOME_FAILED:
		return "danger"
	case identityv1.AuditOutcome_AUDIT_OUTCOME_SENT, identityv1.AuditOutcome_AUDIT_OUTCOME_UNSPECIFIED:
	}
	return toneNeutral
}

// enumKey is the query value of an enum value: its name without the prefix, in
// lower case, as the domain spells it ("sign_in").
func enumKey(name, prefix string) string {
	return strings.ToLower(strings.TrimPrefix(name, prefix))
}

// auditNames resolves the IDs an entry holds to the names the owner knows,
// from the people, tokens, sleeves, accounts and invitations the Access page
// reads. A record that is gone, or never known to those lists, shows its ID.
type auditNames struct {
	people, tokens, sleeves, accounts, invites map[string]string
}

// newAuditNames returns the names the Access inputs hold.
func newAuditNames(in *accessInputs) *auditNames {
	n := &auditNames{
		people: map[string]string{}, tokens: map[string]string{}, sleeves: map[string]string{},
		accounts: map[string]string{}, invites: map[string]string{},
	}
	for _, p := range in.People.GetPeople() {
		n.people[p.GetPerson().GetId()] = p.GetPerson().GetName()
	}
	for _, t := range in.Tokens.GetTokens() {
		n.tokens[t.GetId()] = t.GetName()
	}
	for _, sl := range in.Sleeves {
		n.sleeves[sl.GetId()] = sl.GetName()
	}
	for _, a := range in.Accounts {
		n.accounts[a.GetId()] = a.GetName()
	}
	for _, inv := range in.Invites.GetInvites() {
		n.invites[inv.GetId()] = inv.GetName()
	}
	return n
}

// short is the first characters of a long ID.
func short(id string) string {
	if len(id) <= idChars {
		return id
	}
	return id[:idChars]
}

// named returns "<kind> <name>" for the record id when names knows it, and
// "<kind> <ID> (<gone>)" otherwise.
func named(names map[string]string, kind, id, gone string) string {
	if name, ok := names[id]; ok {
		return kind + " " + name
	}
	return kind + " " + id + " (" + gone + ")"
}

// ref names the record an entry refers to.
func (n *auditNames) ref(r *identityv1.AuditRef) string {
	id := r.GetId()
	switch r.GetKind() {
	case identityv1.AuditRefKind_AUDIT_REF_KIND_PERSON, identityv1.AuditRefKind_AUDIT_REF_KIND_SCRIPT_TOKEN,
		identityv1.AuditRefKind_AUDIT_REF_KIND_VIEW_PLUGIN, identityv1.AuditRefKind_AUDIT_REF_KIND_ACCOUNT,
		identityv1.AuditRefKind_AUDIT_REF_KIND_SLEEVE, identityv1.AuditRefKind_AUDIT_REF_KIND_HOUSEHOLD:
		return n.holderOrTarget(r)
	case identityv1.AuditRefKind_AUDIT_REF_KIND_ENROLLMENT_CODE:
		if name, ok := n.invites[id]; ok {
			return "Invitation for " + name
		}
		return "Link " + short(id)
	case identityv1.AuditRefKind_AUDIT_REF_KIND_PASSKEY:
		return "Passkey " + short(id)
	case identityv1.AuditRefKind_AUDIT_REF_KIND_SESSION:
		return "Session " + short(id)
	case identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_ROLE, identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL,
		identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_CALL, identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_ERROR,
		identityv1.AuditRefKind_AUDIT_REF_KIND_MAINTAINER_RUN, identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_SETTING:
		return modelRef(r)
	case identityv1.AuditRefKind_AUDIT_REF_KIND_UNSPECIFIED:
	}
	return "Record " + short(id)
}

// modelRef names what a model call's entry names: the role, the model, the call
// and the class of an error, by the names the owner configured or the IDs the
// entries hold, never text.
func modelRef(r *identityv1.AuditRef) string {
	prefixes := map[identityv1.AuditRefKind]string{
		identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_ROLE: "Role ", identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL: "Model ",
		identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_CALL: "Call ", identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_ERROR: "Error ",
		identityv1.AuditRefKind_AUDIT_REF_KIND_MAINTAINER_RUN: "Maintainer run ",
		identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_SETTING:  "Setting ",
	}
	id := r.GetId()
	if k := r.GetKind(); k == identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_CALL || k == identityv1.AuditRefKind_AUDIT_REF_KIND_MAINTAINER_RUN {
		id = short(id)
	}
	return prefixes[r.GetKind()] + id
}

// holderOrTarget names what a grant names: a person, a script token or a view
// plugin, an account, a sleeve or the household, by name or, when the record is
// gone, by ID.
func (n *auditNames) holderOrTarget(r *identityv1.AuditRef) string {
	id := r.GetId()
	switch r.GetKind() {
	case identityv1.AuditRefKind_AUDIT_REF_KIND_PERSON:
		if name, ok := n.people[id]; ok {
			return name
		}
		return "Person " + id
	case identityv1.AuditRefKind_AUDIT_REF_KIND_SCRIPT_TOKEN:
		return named(n.tokens, "Script token", id, "not listed")
	case identityv1.AuditRefKind_AUDIT_REF_KIND_VIEW_PLUGIN:
		return "View plugin " + id
	case identityv1.AuditRefKind_AUDIT_REF_KIND_ACCOUNT:
		return named(n.accounts, "Account", id, "removed")
	case identityv1.AuditRefKind_AUDIT_REF_KIND_SLEEVE:
		return named(n.sleeves, "Sleeve", id, "removed")
	case identityv1.AuditRefKind_AUDIT_REF_KIND_HOUSEHOLD:
		return "Whole household"
	case identityv1.AuditRefKind_AUDIT_REF_KIND_UNSPECIFIED, identityv1.AuditRefKind_AUDIT_REF_KIND_SESSION,
		identityv1.AuditRefKind_AUDIT_REF_KIND_PASSKEY, identityv1.AuditRefKind_AUDIT_REF_KIND_ENROLLMENT_CODE,
		identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_ROLE, identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL,
		identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_CALL, identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_ERROR,
		identityv1.AuditRefKind_AUDIT_REF_KIND_MAINTAINER_RUN, identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_SETTING:
	}
	return "Record " + short(id)
}

// actor names who made the change.
func (n *auditNames) actor(a *identityv1.AuditActor) string {
	switch a.GetKind() {
	case identityv1.AuditActorKind_AUDIT_ACTOR_KIND_OWNER:
		return "Owner"
	case identityv1.AuditActorKind_AUDIT_ACTOR_KIND_VIEWER:
		if name, ok := n.people[a.GetPersonId()]; ok && a.GetPersonId() != "" {
			return name
		}
		return "A viewer (session " + short(a.GetId()) + ")"
	case identityv1.AuditActorKind_AUDIT_ACTOR_KIND_ENROLLMENT_CODE:
		if name, ok := n.invites[a.GetId()]; ok {
			return "Invitation link for " + name
		}
		return "Enrollment link " + short(a.GetId())
	case identityv1.AuditActorKind_AUDIT_ACTOR_KIND_SYSTEM:
		return "The core (command line)"
	case identityv1.AuditActorKind_AUDIT_ACTOR_KIND_ANONYMOUS:
		return "Nobody signed in"
	case identityv1.AuditActorKind_AUDIT_ACTOR_KIND_UNSPECIFIED:
	}
	return unknownWord
}

// auditRow is one entry as a table shows it.
type auditRow struct {
	When    string
	Who     string
	What    string
	Records string
	Outcome string
	Tone    string
}

// row renders an entry: a grant names its holder, an arrow and its target, any
// other entry the records it names.
func (n *auditNames) row(e *identityv1.AuditEntry) auditRow {
	words := make([]string, 0, len(e.GetRefs()))
	for _, r := range e.GetRefs() {
		words = append(words, n.ref(r))
	}
	records := strings.Join(words, ", ")
	isGrant := e.GetAction() == identityv1.AuditAction_AUDIT_ACTION_GRANT_CREATED || e.GetAction() == identityv1.AuditAction_AUDIT_ACTION_GRANT_REVOKED
	if isGrant && len(words) == 2 {
		records = words[0] + " → " + words[1]
	}

	return auditRow{
		When: e.GetAt().AsTime().UTC().Format(whenFormat), Who: n.actor(e.GetActor()), What: actionWords(e.GetAction()),
		Records: records, Outcome: outcomeWord(e.GetOutcome()), Tone: outcomeTone(e.GetOutcome()),
	}
}

// auditOption is one choice of a filter's select.
type auditOption struct {
	Value, Label string
	Selected     bool
}

// auditFilterForm is the filter as the owner set it, with the choices.
type auditFilterForm struct {
	Actor, Action, Outcome, From, To string
	Actors, Actions, Outcomes        []auditOption
	// Active reports that some filter is set.
	Active bool
}

// auditView is what the Audit log page shows.
type auditView struct {
	Filter auditFilterForm
	Rows   []auditRow
	// Older is the link to the next page of older entries, empty on the last.
	Older string
	// Unavailable says the core couldn't be read for the entries.
	Unavailable bool
}

// historyView is what the Access page's History section shows.
type historyView struct {
	Rows        []auditRow
	Older       string
	SeeAll      string
	Unavailable bool
}

// auditRoutes maps the Audit log page, the owner's alone.
func (s *server) auditRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+auditPath, s.withOwner(s.auditPage))
}

// parsedAudit is the filter a request names, and the form it is shown in.
type parsedAudit struct {
	req  *identityv1.ListAuditEntriesRequest
	form auditFilterForm
}

// parseAuditQuery reads the filter of the Audit log page from q: the actor
// kind, action and outcome by the names the selects post, and the dates as
// days in UTC, the last day included. ok is false for a value that isn't one of
// the choices or a date out of order, so a malformed link never reaches the core.
func parseAuditQuery(q url.Values) (parsedAudit, bool) {
	form := auditFilterForm{Actor: q.Get("actor"), Action: q.Get("action"), Outcome: q.Get("outcome"), From: q.Get("from"), To: q.Get("to")}
	req := &identityv1.ListAuditEntriesRequest{PageSize: auditPageSize}
	ok := parseAuditWhat(&form, req) && parseAuditDays(&form, req)
	if before := q.Get(beforeParam); before != "" {
		n, err := strconv.ParseInt(before, 10, 64)
		ok = ok && err == nil && n > 0
		req.PageToken = before
	}
	form.Active = form.Actor != "" || form.Action != "" || form.Outcome != "" || form.From != "" || form.To != ""
	return parsedAudit{req: req, form: form}, ok
}

// parseAuditWhat sets the actor kind, action and outcome of req from form, and
// reports whether each one set is one of the choices.
func parseAuditWhat(form *auditFilterForm, req *identityv1.ListAuditEntriesRequest) bool {
	ok := true
	if form.Actor != "" {
		var found bool
		req.ActorKind, found = actorKindOf(form.Actor)
		ok = ok && found
	}
	if form.Action != "" {
		var found bool
		req.Action, found = actionOf(form.Action)
		ok = ok && found
	}
	if form.Outcome != "" {
		var found bool
		req.Outcome, found = outcomeOf(form.Outcome)
		ok = ok && found
	}
	return ok
}

// parseAuditDays sets the range of req from the days of form, in UTC and the
// last day included, and reports whether they are dates and in order.
func parseAuditDays(form *auditFilterForm, req *identityv1.ListAuditEntriesRequest) bool {
	ok := true
	if form.From != "" {
		t, err := time.Parse(dayFormat, form.From)
		ok = ok && err == nil
		req.From = timestamppb.New(t)
	}
	if form.To != "" {
		t, err := time.Parse(dayFormat, form.To)
		ok = ok && err == nil
		req.To = timestamppb.New(t.AddDate(0, 0, 1))
	}
	return ok && (form.From == "" || form.To == "" || req.GetFrom().AsTime().Before(req.GetTo().AsTime()))
}

// actorKindOf returns the actor kind a select value names.
func actorKindOf(key string) (identityv1.AuditActorKind, bool) {
	for n, name := range identityv1.AuditActorKind_name {
		if n != 0 && enumKey(name, "AUDIT_ACTOR_KIND_") == key {
			return identityv1.AuditActorKind(n), true
		}
	}
	return 0, false
}

// actionOf returns the action a select value names.
func actionOf(key string) (identityv1.AuditAction, bool) {
	for n, name := range identityv1.AuditAction_name {
		if n != 0 && enumKey(name, "AUDIT_ACTION_") == key {
			return identityv1.AuditAction(n), true
		}
	}
	return 0, false
}

// outcomeOf returns the outcome a select value names.
func outcomeOf(key string) (identityv1.AuditOutcome, bool) {
	for n, name := range identityv1.AuditOutcome_name {
		if n != 0 && enumKey(name, "AUDIT_OUTCOME_") == key {
			return identityv1.AuditOutcome(n), true
		}
	}
	return 0, false
}

// filterChoices fills the choices of the filter's selects, marking the ones set.
func filterChoices(form *auditFilterForm) {
	for _, n := range sortedKeys(identityv1.AuditActorKind_name) {
		name := identityv1.AuditActorKind_name[n]
		v := enumKey(name, "AUDIT_ACTOR_KIND_")
		form.Actors = append(form.Actors, auditOption{Value: v, Label: actorKindWords(identityv1.AuditActorKind(n)), Selected: v == form.Actor})
	}
	for _, n := range sortedKeys(identityv1.AuditAction_name) {
		name := identityv1.AuditAction_name[n]
		v := enumKey(name, "AUDIT_ACTION_")
		form.Actions = append(form.Actions, auditOption{Value: v, Label: actionWords(identityv1.AuditAction(n)), Selected: v == form.Action})
	}
	for _, n := range sortedKeys(identityv1.AuditOutcome_name) {
		name := identityv1.AuditOutcome_name[n]
		v := enumKey(name, "AUDIT_OUTCOME_")
		word := outcomeWord(identityv1.AuditOutcome(n))
		form.Outcomes = append(form.Outcomes, auditOption{Value: v, Label: word, Selected: v == form.Outcome})
	}
}

// sortedKeys returns the non-zero numbers of an enum's name map, ascending.
func sortedKeys(m map[int32]string) []int32 {
	out := make([]int32, 0, len(m))
	for n := range m {
		if n != 0 {
			out = append(out, n)
		}
	}
	slices.Sort(out)
	return out
}

// auditPage shows every entry of the audit log (D-211) to the owner, newest
// first and paged, with the filters of D-211: the kind of actor, the action, the
// outcome and a range of days. Reading the log is itself no entry.
func (s *server) auditPage(w http.ResponseWriter, r *http.Request, sess session) {
	parsed, ok := parseAuditQuery(r.URL.Query())
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The filter is malformed.")
		return
	}
	parsed.req.Origin, parsed.req.SessionToken = sess.Origin, sess.Token
	filterChoices(&parsed.form)
	view := auditView{Filter: parsed.form}
	in, err := s.loadAccessInputs(r, sess)
	if err != nil {
		s.coreFailed(w, r, "audit log", err)
		return
	}
	resp, err := s.auth.ListAuditEntries(r.Context(), connect.NewRequest(parsed.req))
	switch {
	case connect.CodeOf(err) == connect.CodeUnavailable:
		view.Unavailable = true
	case err != nil:
		s.coreFailed(w, r, "ListAuditEntries", err)
		return
	default:
		names := newAuditNames(in)
		for _, e := range resp.Msg.GetEntries() {
			view.Rows = append(view.Rows, names.row(e))
		}
		if next := resp.Msg.GetNextPageToken(); next != "" {
			q := r.URL.Query()
			q.Set(beforeParam, next)
			view.Older = auditPath + "?" + q.Encode()
		}
	}
	s.show(w, r, http.StatusOK, "audit", page{Title: "Audit log", Session: &sess, Data: view})
}

// loadHistory reads the page of the Access page's history the request names:
// the grants, revocations and links, newest first (D-210). It never fails the
// page: a core that can't answer is said so in the section.
func (s *server) loadHistory(r *http.Request, sess session, in *accessInputs) historyView {
	view := historyView{SeeAll: auditPath}
	req := &identityv1.ListAuditEntriesRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Actions: historyActions(), PageSize: historyPageSize,
	}
	if before := r.URL.Query().Get(historyBeforeParam); before != "" {
		if n, err := strconv.ParseInt(before, 10, 64); err == nil && n > 0 {
			req.PageToken = before
		}
	}
	resp, err := s.auth.ListAuditEntries(r.Context(), connect.NewRequest(req))
	if err != nil {
		s.log.WarnContext(r.Context(), "call to the core failed", "method", "ListAuditEntries", "code", connect.CodeOf(err).String())
		view.Unavailable = true
		return view
	}
	names := newAuditNames(in)
	for _, e := range resp.Msg.GetEntries() {
		view.Rows = append(view.Rows, names.row(e))
	}
	if next := resp.Msg.GetNextPageToken(); next != "" {
		view.Older = fmt.Sprintf("%s?%s=%s%s", accessPath, historyBeforeParam, url.QueryEscape(next), historyAnchor)
	}
	return view
}
