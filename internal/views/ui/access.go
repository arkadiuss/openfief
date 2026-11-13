package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/durationpb"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// accessPath is the Access page (D-210), and the anchors on it the Inbox links
// to.
const (
	accessPath        = "/access"
	viewPluginsAnchor = "#view-plugins-title"
	tokensAnchor      = "#tokens-title"
	peopleAnchor      = "#people-title"
)

// The kinds of a grant choice's value, "kind:id", as the forms post them.
const (
	choiceHousehold = "household"
	choiceSleeve    = "sleeve"
	choiceAccount   = "account"
	choiceSep       = ":"
	// grantField is the name of every grant checkbox.
	grantField = "grant"
	// nameField is the name of the field that names a token or a viewer.
	nameField = "name"
	// maxChoiceChars caps one posted choice: a target's ID is at most 128
	// characters, and its kind a few more.
	maxChoiceChars = 160
)

// grantChoice is one thing the owner can let a token or a plugin view: a
// checkbox of a grants form.
type grantChoice struct {
	// Value is "kind:id", what the form posts.
	Value   string
	Label   string
	Checked bool
}

// grantChoices are the checkboxes of one grants form: the whole household, each
// sleeve and each account, checked where the holder holds the grant (D-209).
// Prefix makes their element IDs unique on the page.
type grantChoices struct {
	Prefix    string
	Household grantChoice
	Sleeves   []grantChoice
	Accounts  []grantChoice
}

// accessHolder is a script token or a view plugin as the Access page lists it,
// with what it may view in words and the form that changes it.
type accessHolder struct {
	// ID is the token's ID or the plugin's instance name.
	ID    string
	Token *identityv1.ScriptToken
	// Holds are the targets it may view, in words; none when it holds no grant.
	Holds []string
	// Editable reports whether the holder's grants can still change: not a
	// revoked or expired token's.
	Editable bool
	Choices  grantChoices
	// Action is the path the grants form posts to.
	Action string
}

// accessData is what the Access page shows (D-210): the script tokens and the
// view plugins with their grants, what can be granted, and once, right after
// it was made, a new token's secret.
type accessData struct {
	// People are the owner and the viewers (D-210), and Invite the form that
	// invites a viewer.
	People []personRow
	Invite inviteForm
	// Confirm is the viewer the owner asked to revoke, shown with what
	// revoking does until the owner confirms.
	Confirm *personRow
	Tokens  []accessHolder
	// Plugins are the installed view plugins that read with their own
	// identity; PluginsUnavailable says the core can't read them now.
	Plugins            []accessHolder
	PluginsUnavailable bool
	// Enforced is whether grants limit what holders view now: false under the
	// policy, where they are stored with no effect.
	Enforced bool
	// New describes the grants form of a new token.
	New         grantChoices
	DefaultDays int
	NewToken    *identityv1.ScriptToken
	NewSecret   string
	Error       string
	// PasskeyCheck is an action the core refused for want of a recent
	// passkey check (D-183), to post again once the owner confirms.
	PasskeyCheck *pendingAction
	// Posted is a new token's form as it was posted, shown again when the
	// core refused it.
	Posted *postedToken
}

// postedToken is the form of a new token as the owner filled it.
type postedToken struct {
	Name   string
	Days   string
	MCP    bool
	Grants []string
}

// accessRoutes maps the Access page and its forms, every one the owner's.
func (s *server) accessRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+accessPath, s.withOwner(s.accessPage))
	s.peopleRoutes(mux)
	mux.HandleFunc("POST "+accessPath+"/tokens", s.withOwner(s.createToken))
	mux.HandleFunc("POST "+accessPath+"/tokens/{id}/revoke", s.withOwner(s.revokeToken))
	mux.HandleFunc("POST "+accessPath+"/tokens/{id}/grants", s.withOwner(s.updateTokenGrants))
	mux.HandleFunc("POST "+accessPath+"/plugins/{instance}/grants", s.withOwner(s.updatePluginGrants))
}

// accessPage shows the Access page.
func (s *server) accessPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderAccess(w, r, sess, http.StatusOK, &accessData{})
}

// renderAccess loads the people, the tokens, the view plugins and what can be
// granted, and shows them with whatever the last action left in data.
func (s *server) renderAccess(w http.ResponseWriter, r *http.Request, sess session, status int, data *accessData) {
	in, err := s.loadAccessInputs(r, sess)
	if err != nil {
		s.coreFailed(w, r, "access", err)
		return
	}
	var posted []string
	if data.Posted != nil {
		posted = data.Posted.Grants
	}
	data.Enforced = in.Tokens.GetGrantsEnforced()
	data.DefaultDays = defaultTokenDays
	data.New = newChoices("new-token", in.Sleeves, in.Accounts, posted)
	data.PluginsUnavailable = in.Plugins == nil
	for _, t := range in.Tokens.GetTokens() {
		data.Tokens = append(data.Tokens, tokenHolder(t, in.Sleeves, in.Accounts))
	}
	for _, p := range in.Plugins.GetPlugins() {
		data.Plugins = append(data.Plugins, pluginHolder(p, in.Sleeves, in.Accounts))
	}
	data.People = peopleRows(in.People.GetPeople(), in.Sleeves, in.Accounts)
	data.Invite.Choices = newChoices("new-viewer", in.Sleeves, in.Accounts, data.Invite.Grants)
	data.Invite.Lifetimes = selectLifetime(inviteLifetimes(), data.Invite.Minutes)
	s.show(w, r, status, "access", page{Title: "Access", Session: &sess, Data: data})
}

// accessInputs is what the Access page reads: the script tokens, the
// installed view plugins (nil when the core can't read them now), the people,
// and the sleeves and accounts that can be granted.
type accessInputs struct {
	Tokens   *identityv1.ListScriptTokensResponse
	Plugins  *identityv1.ListViewPluginsResponse
	People   *identityv1.ListPeopleResponse
	Sleeves  []*viewsv1.Sleeve
	Accounts []*viewsv1.Account
}

// loadAccessInputs reads what the Access page shows.
func (s *server) loadAccessInputs(r *http.Request, sess session) (*accessInputs, error) {
	ctx := r.Context()
	tokens, err := s.auth.ListScriptTokens(ctx, connect.NewRequest(&identityv1.ListScriptTokensRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	if err != nil {
		return nil, fmt.Errorf("ListScriptTokens: %w", err)
	}
	in := &accessInputs{Tokens: tokens.Msg}
	plugins, err := s.auth.ListViewPlugins(ctx, connect.NewRequest(&identityv1.ListViewPluginsRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	switch {
	case connect.CodeOf(err) == connect.CodeUnavailable:
		// The core can't read its cluster now; the page says so.
	case err != nil:
		return nil, fmt.Errorf("ListViewPlugins: %w", err)
	default:
		in.Plugins = plugins.Msg
	}
	people, err := s.auth.ListPeople(ctx, connect.NewRequest(&identityv1.ListPeopleRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	if err != nil {
		return nil, fmt.Errorf("ListPeople: %w", err)
	}
	in.People = people.Msg
	sleeves, err := s.read.ListSleeves(ctx, bearer(&viewsv1.ListSleevesRequest{}, sess))
	if err != nil {
		return nil, fmt.Errorf("ListSleeves: %w", err)
	}
	accounts, err := s.read.ListAccounts(ctx, bearer(&viewsv1.ListAccountsRequest{}, sess))
	if err != nil {
		return nil, fmt.Errorf("ListAccounts: %w", err)
	}
	in.Sleeves, in.Accounts = sleeves.Msg.GetSleeves(), accounts.Msg.GetAccounts()
	return in, nil
}

// choiceValue is the value a grants form posts for a target: "kind:id".
func choiceValue(kind, id string) string { return kind + choiceSep + id }

// targetChoiceKinds names each grant target kind as the forms post it.
func targetChoiceKinds() map[identityv1.GrantTargetKind]string {
	return map[identityv1.GrantTargetKind]string{
		identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD: choiceHousehold,
		identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE:    choiceSleeve,
		identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT:   choiceAccount,
	}
}

// heldValues returns the form values of grants, "kind:id" each.
func heldValues(grants []*identityv1.Grant) []string {
	kinds := targetChoiceKinds()
	out := make([]string, 0, len(grants))
	for _, g := range grants {
		kind, ok := kinds[g.GetTargetKind()]
		if !ok {
			continue
		}
		id := g.GetTargetId()
		if kind == choiceHousehold {
			id = "home"
		}
		out = append(out, choiceValue(kind, id))
	}
	return out
}

// newChoices returns the checkboxes of a grants form with the values in held
// checked.
func newChoices(prefix string, sleeves []*viewsv1.Sleeve, accounts []*viewsv1.Account, held []string) grantChoices {
	c := grantChoices{
		Prefix:    prefix,
		Household: grantChoice{Value: choiceValue(choiceHousehold, "home"), Label: "Whole household"},
	}
	c.Household.Checked = slices.Contains(held, c.Household.Value)
	for _, sl := range sleeves {
		v := choiceValue(choiceSleeve, sl.GetId())
		c.Sleeves = append(c.Sleeves, grantChoice{Value: v, Label: sl.GetName(), Checked: slices.Contains(held, v)})
	}
	for _, a := range accounts {
		v := choiceValue(choiceAccount, a.GetId())
		c.Accounts = append(c.Accounts, grantChoice{Value: v, Label: accountLabel(a), Checked: slices.Contains(held, v)})
	}
	return c
}

// accountLabel names an account for a grant: its name, and its institution when
// it has one.
func accountLabel(a *viewsv1.Account) string {
	if a.GetInstitution() == "" {
		return a.GetName()
	}
	return a.GetName() + " (" + a.GetInstitution() + ")"
}

// holdsInWords returns what grants let their holder view, in words: the
// household, then sleeves, then accounts, each by name; a target the lists no
// longer have is named by its ID.
func holdsInWords(grants []*identityv1.Grant, sleeves []*viewsv1.Sleeve, accounts []*viewsv1.Account) []string {
	var household, sleeveWords, accountWords []string
	for _, g := range grants {
		switch g.GetTargetKind() {
		case identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD:
			household = append(household, "Whole household")
		case identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE:
			name := g.GetTargetId()
			if i := slices.IndexFunc(sleeves, func(sl *viewsv1.Sleeve) bool { return sl.GetId() == name }); i >= 0 {
				name = sleeves[i].GetName()
			}
			sleeveWords = append(sleeveWords, "Sleeve "+name)
		case identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT:
			name := g.GetTargetId()
			if i := slices.IndexFunc(accounts, func(a *viewsv1.Account) bool { return a.GetId() == name }); i >= 0 {
				name = accountLabel(accounts[i])
			}
			accountWords = append(accountWords, "Account "+name)
		case identityv1.GrantTargetKind_GRANT_TARGET_KIND_UNSPECIFIED:
		}
	}
	return slices.Concat(household, sleeveWords, accountWords)
}

// tokenHolder returns a script token as the page lists it.
func tokenHolder(t *identityv1.ScriptToken, sleeves []*viewsv1.Sleeve, accounts []*viewsv1.Account) accessHolder {
	active := t.GetStatus() == identityv1.ScriptTokenStatus_SCRIPT_TOKEN_STATUS_ACTIVE
	return accessHolder{
		ID: t.GetId(), Token: t, Editable: active, Holds: holdsInWords(t.GetGrants(), sleeves, accounts),
		Choices: newChoices("token-"+t.GetId(), sleeves, accounts, heldValues(t.GetGrants())),
		Action:  accessPath + "/tokens/" + url.PathEscape(t.GetId()) + "/grants",
	}
}

// pluginHolder returns a view plugin as the page lists it.
func pluginHolder(p *identityv1.ViewPlugin, sleeves []*viewsv1.Sleeve, accounts []*viewsv1.Account) accessHolder {
	return accessHolder{
		ID: p.GetInstance(), Editable: true, Holds: holdsInWords(p.GetGrants(), sleeves, accounts),
		Choices: newChoices("plugin-"+p.GetInstance(), sleeves, accounts, heldValues(p.GetGrants())),
		Action:  accessPath + "/plugins/" + url.PathEscape(p.GetInstance()) + "/grants",
	}
}

// parseChoices turns the posted grant values into targets: each must be
// "household:home", "sleeve:<id>" or "account:<id>". ok is false for anything
// else, so a malformed form never reaches the core.
func parseChoices(values []string) ([]*identityv1.GrantTarget, bool) {
	kinds := map[string]identityv1.GrantTargetKind{
		choiceHousehold: identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD,
		choiceSleeve:    identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE,
		choiceAccount:   identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT,
	}
	var out []*identityv1.GrantTarget
	seen := map[string]bool{}
	for _, v := range values {
		if seen[v] {
			continue
		}
		if len(v) > maxChoiceChars {
			return nil, false
		}
		seen[v] = true
		kind, id, found := strings.Cut(v, choiceSep)
		k, known := kinds[kind]
		if !found || !known || id == "" || (kind == choiceHousehold && id != "home") {
			return nil, false
		}
		out = append(out, &identityv1.GrantTarget{Kind: k, Id: id})
	}
	return out, true
}

// postedFields returns the fields of a form to post again after a passkey
// check: each of names with each value it was posted with.
func postedFields(r *http.Request, names ...string) []formField {
	var out []formField
	for _, name := range names {
		for _, v := range r.PostForm[name] {
			out = append(out, formField{Name: name, Value: v})
		}
	}
	return out
}

// createToken issues a script token with the grants chosen on the form
// (CreateScriptToken) and shows its secret this once; the core never shows it
// again. If the core can't pass its grants on to the authorization model it
// takes the token back, and no secret is shown.
func (s *server) createToken(w http.ResponseWriter, r *http.Request, sess session) {
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	posted := &postedToken{
		Name: r.PostFormValue("name"), Days: r.PostFormValue("days"), MCP: r.PostFormValue("mcp") != "", Grants: r.PostForm[grantField],
	}
	days, ok := tokenDays(posted.Days)
	if !ok {
		s.renderAccess(w, r, sess, http.StatusBadRequest, &accessData{Error: "Days must be a whole number from 1 to 365.", Posted: posted})
		return
	}
	targets, ok := parseChoices(posted.Grants)
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	resp, err := s.auth.CreateScriptToken(r.Context(), connect.NewRequest(&identityv1.CreateScriptTokenRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Name: posted.Name,
		Ttl:        durationpb.New(time.Duration(days) * hoursPerDay * time.Hour),
		McpAllowed: posted.MCP, Grants: targets,
	}))
	if err != nil {
		s.createTokenFailed(w, r, sess, err, posted)
		return
	}
	s.renderAccess(w, r, sess, http.StatusOK, &accessData{NewToken: resp.Msg.GetToken(), NewSecret: resp.Msg.GetSecret()})
}

// createTokenFailed answers a token the core did not create: for want of a
// recent passkey check (D-183), the page with the action to confirm and post
// again; for an authorization model that did not take the grants, the page
// saying nothing was kept; any other refusal with the core's reason.
func (s *server) createTokenFailed(w http.ResponseWriter, r *http.Request, sess session, err error, posted *postedToken) {
	switch {
	case passkeyCheckRequired(err):
		fields := slices.Concat([]formField{{Name: nameField, Value: posted.Name}, {Name: "days", Value: posted.Days}}, postedFields(r, "mcp", grantField))
		s.renderAccess(w, r, sess, http.StatusForbidden, &accessData{
			PasskeyCheck: &pendingAction{What: "Creating a script token", Action: r.URL.Path, Fields: fields}, Posted: posted,
		})
	case connect.CodeOf(err) == connect.CodeUnavailable:
		s.renderAccess(w, r, sess, http.StatusServiceUnavailable, &accessData{
			Error:  "The token was not created: SpiceDB did not take its grants, so nothing was kept and no secret exists. Try again in a moment.",
			Posted: posted,
		})
	default:
		s.accessFailed(w, r, sess, "CreateScriptToken", err, posted)
	}
}

// accessFailed shows the Access page again with the core's reason when it
// refused an action, and otherwise answers like any failed call.
func (s *server) accessFailed(w http.ResponseWriter, r *http.Request, sess session, method string, err error, posted *postedToken) {
	if status, refused := refusalStatus(connect.CodeOf(err)); refused {
		s.renderAccess(w, r, sess, status, &accessData{Error: connectMessage(err), Posted: posted})
		return
	}
	s.coreFailed(w, r, method, err)
}

// revokeToken revokes a script token (RevokeScriptToken), which ends its
// grants.
func (s *server) revokeToken(w http.ResponseWriter, r *http.Request, sess session) {
	_, err := s.auth.RevokeScriptToken(r.Context(), connect.NewRequest(&identityv1.RevokeScriptTokenRequest{
		Origin: sess.Origin, SessionToken: sess.Token, TokenId: r.PathValue("id"),
	}))
	if err != nil {
		s.accessFailed(w, r, sess, "RevokeScriptToken", err, nil)
		return
	}
	http.Redirect(w, r, accessPath+tokensAnchor, http.StatusSeeOther)
}

// updateTokenGrants makes a script token's grants the ones posted.
func (s *server) updateTokenGrants(w http.ResponseWriter, r *http.Request, sess session) {
	s.updateGrants(w, r, sess, identityv1.GrantHolderKind_GRANT_HOLDER_KIND_SCRIPT_TOKEN, r.PathValue("id"), tokensAnchor)
}

// updatePluginGrants makes a view plugin's grants the ones posted.
func (s *server) updatePluginGrants(w http.ResponseWriter, r *http.Request, sess session) {
	s.updateGrants(w, r, sess, identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN, r.PathValue("instance"), viewPluginsAnchor)
}

// updateGrants makes the grants of a holder the ones the form posted, among
// the targets the form offers (UpdateGrants): the grants it newly checks are
// added and those it unchecks are ended, and a grant to a target the form
// doesn't offer, such as a deleted sleeve's, is left to the core.
func (s *server) updateGrants(w http.ResponseWriter, r *http.Request, sess session, kind identityv1.GrantHolderKind, id, anchor string) {
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	want, ok := parseChoices(r.PostForm[grantField])
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	in, err := s.loadAccessInputs(r, sess)
	if err != nil {
		s.coreFailed(w, r, "access", err)
		return
	}
	held, found := heldBy(kind, id, in)
	if !found {
		s.errorPage(w, r, http.StatusNotFound, "No such viewer, script token or view plugin.")
		return
	}
	offered := newChoices("", in.Sleeves, in.Accounts, nil)
	diff := diffGrants(want, held, &offered)
	if len(diff.Add) > 0 || len(diff.Remove) > 0 {
		_, err = s.auth.UpdateGrants(r.Context(), connect.NewRequest(&identityv1.UpdateGrantsRequest{
			Origin: sess.Origin, SessionToken: sess.Token, HolderKind: kind, HolderId: id, Add: diff.Add, Remove: diff.Remove,
		}))
		if err != nil {
			s.updateGrantsFailed(w, r, sess, err)
			return
		}
	}
	http.Redirect(w, r, accessPath+anchor, http.StatusSeeOther)
}

// updateGrantsFailed answers a grants change the core did not make: when the
// authorization model did not take it, the page saying it is stored and
// saving again passes it on; otherwise as accessFailed does.
func (s *server) updateGrantsFailed(w http.ResponseWriter, r *http.Request, sess session, err error) {
	if connect.CodeOf(err) == connect.CodeUnavailable {
		s.renderAccess(w, r, sess, http.StatusServiceUnavailable, &accessData{
			Error: "The change is stored, but SpiceDB did not take it yet. Save it again to pass it on; the core also repairs it at its next start.",
		})
		return
	}
	s.accessFailed(w, r, sess, "UpdateGrants", err, nil)
}

// heldBy returns the grants, as form values, of the viewer, script token or
// view plugin id among those listed, and false when it isn't listed or can't
// change any more: a revoked viewer or token.
func heldBy(kind identityv1.GrantHolderKind, id string, in *accessInputs) ([]string, bool) {
	switch kind {
	case identityv1.GrantHolderKind_GRANT_HOLDER_KIND_PERSON:
		people := in.People.GetPeople()
		if i := slices.IndexFunc(people, func(p *identityv1.PersonAccess) bool { return activeViewer(p) && p.GetPerson().GetId() == id }); i >= 0 {
			return heldValues(people[i].GetGrants()), true
		}
	case identityv1.GrantHolderKind_GRANT_HOLDER_KIND_SCRIPT_TOKEN:
		tokens := in.Tokens.GetTokens()
		active := func(t *identityv1.ScriptToken) bool {
			return t.GetId() == id && t.GetStatus() == identityv1.ScriptTokenStatus_SCRIPT_TOKEN_STATUS_ACTIVE
		}
		if i := slices.IndexFunc(tokens, active); i >= 0 {
			return heldValues(tokens[i].GetGrants()), true
		}
	case identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN:
		plugins := in.Plugins.GetPlugins()
		if i := slices.IndexFunc(plugins, func(p *identityv1.ViewPlugin) bool { return p.GetInstance() == id }); i >= 0 {
			return heldValues(plugins[i].GetGrants()), true
		}
	case identityv1.GrantHolderKind_GRANT_HOLDER_KIND_UNSPECIFIED:
	}
	return nil, false
}

// grantDiff is the targets to add and to end so a holder's grants match a form.
type grantDiff struct {
	Add, Remove []*identityv1.GrantTarget
}

// diffGrants returns the targets to add and to remove so that a holder holding
// held holds want among the choices the form offers: a target held and not
// offered stays as it is.
func diffGrants(want []*identityv1.GrantTarget, held []string, offered *grantChoices) grantDiff {
	var add, remove []*identityv1.GrantTarget
	wanted := make([]string, 0, len(want))
	for _, t := range want {
		v := choiceValue(targetChoiceKinds()[t.GetKind()], t.GetId())
		wanted = append(wanted, v)
		if !slices.Contains(held, v) {
			add = append(add, t)
		}
	}
	all := slices.Concat([]grantChoice{offered.Household}, offered.Sleeves, offered.Accounts)
	for _, c := range all {
		if slices.Contains(held, c.Value) && !slices.Contains(wanted, c.Value) {
			if targets, ok := parseChoices([]string{c.Value}); ok {
				remove = append(remove, targets...)
			}
		}
	}
	return grantDiff{Add: add, Remove: remove}
}
