package ui

import (
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
)

// modelsPath is the owner's page of the agents-and-models settings (D-212):
// the one switch and the feature toggles, the model providers, which models
// may receive which data tier (D-206) and each role's route (D-205).
const modelsPath = "/settings/models"

// The roles of the router, which are also the names of the toggles of the
// features that use them; the MCP server has a toggle and no role.
const (
	roleAsk        = "ask"
	roleMaintainer = "maintainer"
	featureMCP     = "mcp"
)

// The ways a route entry can be wrong, as routeProblem says them.
const (
	problemBroken   = "broken"
	problemDown     = "down"
	problemUnlisted = "unlisted"
	// reasonMissing and reasonNotProvider are the core's reasons a route
	// entry is broken (BrokenEntry).
	reasonMissing     = "missing"
	reasonNotProvider = "not-model-provider"
)

// errBadForm reports a posted form of the Models page that is malformed.
var errBadForm = errors.New("the form is malformed")

// roleInfo describes a role of the router.
type roleInfo struct {
	Role, Name, Does, Feature string
	Needs                     []string
}

// featureInfo describes a feature's toggle.
type featureInfo struct {
	Feature, Name, Help string
	NotBuilt            bool
}

// modelTiers are the data tiers, in the order the owner reads them, with what
// each holds (D-206).
func modelTiers() []tierRow {
	return []tierRow{
		{"T0", "Whole portfolio", "Your holdings, amounts, transactions and tax figures."},
		{"TF", "One fief", "One fief's own lease, cash, positions and working orders. No fief exists yet."},
		{"T1", "A thesis", "A thesis's text and the instruments it names. No thesis exists yet."},
		{"T2", "Public data", "Prices, filings, news and the instructions OpenFief writes: nothing of yours."},
	}
}

// modelRoles describes each role: its name, what it does, the feature whose
// toggle turns it on, and the tiers a request of it holds, so the page can say
// when no model on its route may receive them.
func modelRoles() []roleInfo {
	return []roleInfo{
		{roleAsk, "The Ask box", "Answers the questions that are not one of its commands (D-203).", roleAsk, []string{"T0", "T2"}},
		{roleMaintainer, "The maintainer", "Writes the weekly note on the model portfolio's drift (D-204).", roleMaintainer, []string{"T0", "T2"}},
	}
}

// modelFeatures describes each feature's toggle. NotBuilt is true for the one
// that is stored and does nothing yet.
func modelFeatures() []featureInfo {
	return []featureInfo{
		{roleAsk, "Ask box model", "Sends the Ask box's questions that are not one of its commands to the model on its route (D-203).", false},
		{roleMaintainer, "Maintainer", "A model writes a note about the model portfolio every Monday and when you ask, and may propose target changes for you to apply or dismiss (D-204).", false},
		{featureMCP, "MCP server", "Not built yet: the toggle is stored, and no MCP endpoint exists to answer (D-213).", true},
	}
}

// modelsData is what the Models page shows.
type modelsData struct {
	Error, Notice string
	Enabled       bool
	Features      []featureRow
	Providers     []providerRow
	Tiers         []tierRow
	Models        []modelRow
	Routes        []routeRow
}

// featureRow is one feature's toggle.
type featureRow struct {
	Feature, Name, Help string
	On, NotBuilt        bool
	Action              string
}

// providerRow is a model-provider instance with its declared terms and models.
type providerRow struct {
	Instance  string
	Available bool
	// Location and Retention are the provider's own declarations, in words.
	Location  string
	Retention string
	Models    []providerModel
}

// providerModel is one model entry of a provider.
type providerModel struct {
	ID, Ref, Input, Output, Context, MaxOutput string
	Tools, Structured                          bool
}

// tierRow explains one data tier.
type tierRow struct{ Tier, Name, Holds string }

// modelRow is a row of the tier grid: a model with a checkbox per tier.
type modelRow struct {
	Ref string
	// Note says why the model is listed although no provider offers it now.
	Note  string
	Cells []tierCell
}

// tierCell is the checkbox of one tier for one model.
type tierCell struct {
	Tier, Value, Label string
	Checked            bool
}

// routeRow is a role's route: its models in order, and the form that adds one.
type routeRow struct {
	Role, Name, Does string
	// On is true when the role's feature is on under the switch: the route is in use.
	On      bool
	Entries []routeEntry
	// Choices are the models a provider offers that the route lacks.
	Choices []string
	Action  string
	// Warning is set when no model of the route may receive what the role sends.
	Warning string
}

// routeEntry is one model of a route.
type routeEntry struct {
	Position string
	Ref      string
	Index    int
	First    bool
	Last     bool
	// Problem is one of the problem constants or empty; ProblemWords says it.
	Problem      string
	ProblemWords string
	// Missing are the tiers the role sends that this model may not receive.
	Missing []string
}

// routeFacts is what the page knows of the models, for judging a route.
type routeFacts struct {
	offered   map[string]bool
	available map[string]bool
	allowed   map[string]map[string]bool
	broken    map[string]string
}

// modelsRoutes maps the Models page and its forms, every one the owner's.
func (s *server) modelsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+modelsPath, s.withOwner(s.modelsPage))
	mux.HandleFunc("POST "+modelsPath+"/switch", s.withOwner(s.setModelsSwitch))
	mux.HandleFunc("POST "+modelsPath+"/features/{feature}", s.withOwner(s.setModelsFeature))
	mux.HandleFunc("POST "+modelsPath+"/tiers", s.withOwner(s.setModelTiers))
	mux.HandleFunc("POST "+modelsPath+"/routes/{role}", s.withOwner(s.changeModelRoute))
}

// modelsPage shows the settings of agents and models.
func (s *server) modelsPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderModels(w, r, sess, http.StatusOK, modelsNotice(r.URL.Query()), "")
}

// modelsNotice is what the owner's last change did, read from the redirect's query.
func modelsNotice(q url.Values) string {
	switch q.Get("saved") {
	case "switch":
		return "The agents-and-models switch is saved."
	case "feature":
		return "The toggle is saved."
	case "tiers":
		return "The models allowed for each data tier are saved."
	case "route":
		return "The route is saved."
	}
	return ""
}

// renderModels reads the settings and the providers from the core and shows the
// page with status, a notice and the core's reason for refusing the owner's
// last change, if any.
func (s *server) renderModels(w http.ResponseWriter, r *http.Request, sess session, status int, notice, refusal string) {
	if s.modelSettings == nil {
		s.errorPage(w, r, http.StatusNotFound, "This core serves no models settings.")
		return
	}
	got, ok := s.loadModels(w, r, sess)
	if !ok {
		return
	}
	data := buildModels(got.settings, got.providers)
	data.Notice, data.Error = notice, refusal
	s.show(w, r, status, "models", page{Title: "Models", Session: &sess, Data: data})
}

// modelsLoad is what the core says of the settings and of the providers.
type modelsLoad struct {
	settings  *modelsv1.GetModelSettingsResponse
	providers *modelsv1.ListModelProvidersResponse
}

// loadModels asks the core for the settings and the providers; ok is false
// after it answered the failure.
func (s *server) loadModels(w http.ResponseWriter, r *http.Request, sess session) (modelsLoad, bool) {
	got, err := s.modelSettings.GetModelSettings(r.Context(), connect.NewRequest(&modelsv1.GetModelSettingsRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	if err != nil {
		s.coreFailed(w, r, "GetModelSettings", err)
		return modelsLoad{}, false
	}
	providers, err := s.modelSettings.ListModelProviders(r.Context(), connect.NewRequest(&modelsv1.ListModelProvidersRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	if err != nil {
		s.coreFailed(w, r, "ListModelProviders", err)
		return modelsLoad{}, false
	}
	return modelsLoad{settings: got.Msg, providers: providers.Msg}, true
}

// buildModels turns the core's answers into the page's data.
func buildModels(set *modelsv1.GetModelSettingsResponse, list *modelsv1.ListModelProvidersResponse) *modelsData {
	data := &modelsData{Enabled: set.GetAgentsEnabled(), Tiers: modelTiers()}
	toggles := map[string]bool{}
	for _, f := range set.GetFeatures() {
		toggles[f.GetFeature()] = f.GetEnabled()
	}
	for _, f := range modelFeatures() {
		data.Features = append(data.Features, featureRow{
			Feature: f.Feature, Name: f.Name, Help: f.Help, On: toggles[f.Feature], NotBuilt: f.NotBuilt,
			Action: modelsPath + "/features/" + f.Feature,
		})
	}
	facts := &routeFacts{offered: map[string]bool{}, available: map[string]bool{}, allowed: map[string]map[string]bool{}, broken: map[string]string{}}
	for _, p := range list.GetProviders() {
		data.Providers = append(data.Providers, providerOf(p))
		facts.available[p.GetInstance()] = p.GetAvailable()
		for _, m := range p.GetModels() {
			facts.offered[p.GetInstance()+"/"+m.GetId()] = true
		}
	}
	for _, t := range set.GetTiers() {
		facts.allowed[t.GetTier()] = map[string]bool{}
		for _, m := range t.GetModels() {
			facts.allowed[t.GetTier()][m] = true
		}
	}
	for _, b := range set.GetBroken() {
		facts.broken[b.GetRole()+" "+b.GetModel()] = b.GetReason()
	}
	data.Models = gridRows(facts)
	for _, role := range modelRoles() {
		data.Routes = append(data.Routes, routeOf(&role, set, facts, set.GetAgentsEnabled() && toggles[role.Feature]))
	}
	return data
}

// providerOf describes a provider instance in words.
func providerOf(p *modelsv1.ModelProvider) providerRow {
	row := providerRow{Instance: p.GetInstance(), Available: p.GetAvailable()}
	if !p.GetAvailable() {
		return row
	}
	switch p.GetLocation() {
	case modelsv1.Location_LOCATION_OWNER_HARDWARE:
		row.Location = "Its models run on hardware you run."
	case modelsv1.Location_LOCATION_THIRD_PARTY_CLOUD, modelsv1.Location_LOCATION_UNSPECIFIED:
		row.Location = "Its models run in a third party's cloud."
	}
	row.Retention = retentionWords(p.GetRetention())
	for _, m := range p.GetModels() {
		row.Models = append(row.Models, modelOf(p.GetInstance(), m))
	}
	return row
}

// retentionWords says what the provider declares it keeps of a request.
func retentionWords(r *modelsv1.Retention) string {
	var kept string
	switch days := r.GetDays(); {
	case days <= 0:
		kept = "It declares it keeps nothing of a request or its answer"
	case days == 1:
		kept = "It declares it keeps a request and its answer for 1 day"
	default:
		kept = "It declares it keeps a request and its answer for " + strconv.Itoa(int(days)) + " days"
	}
	if r.GetMayTrain() {
		return kept + ", and that it may train models on them."
	}
	return kept + ", and that it does not train models on them."
}

// modelOf describes a model entry: prices per million tokens with their currency.
func modelOf(instance string, m *modelsv1.ModelEntry) providerModel {
	out := providerModel{
		ID: m.GetId(), Ref: instance + "/" + m.GetId(), Input: "no price", Output: "no price",
		Context: tokens(m.GetContextTokens()), MaxOutput: tokens(m.GetMaxOutputTokens()),
		Tools: m.GetToolCalling(), Structured: m.GetStructuredOutput(),
	}
	if m.GetInputPrice() != "" || m.GetOutputPrice() != "" {
		out.Input = m.GetInputPrice() + " " + m.GetPriceCurrency()
		out.Output = m.GetOutputPrice() + " " + m.GetPriceCurrency()
	}
	return out
}

// tokens spells a token limit, "not stated" for zero.
func tokens(n int64) string {
	if n <= 0 {
		return "not stated"
	}
	return strconv.FormatInt(n, 10)
}

// gridRows lists the models the tier grid has a row for: every model a
// provider offers, and every model a tier allows that none offers any more, so
// the owner can untick it.
func gridRows(facts *routeFacts) []modelRow {
	refs := map[string]bool{}
	for ref := range facts.offered {
		refs[ref] = true
	}
	for _, set := range facts.allowed {
		for ref := range set {
			refs[ref] = true
		}
	}
	sorted := slices.Sorted(func(yield func(string) bool) {
		for r := range refs {
			if !yield(r) {
				return
			}
		}
	})
	rows := make([]modelRow, 0, len(sorted))
	for _, ref := range sorted {
		row := modelRow{Ref: ref}
		if !facts.offered[ref] {
			row.Note = unofferedWords(ref, facts.available)
		}
		for _, t := range modelTiers() {
			row.Cells = append(row.Cells, tierCell{
				Tier: t.Tier, Value: t.Tier + "|" + ref, Label: "Allow " + ref + " to receive " + t.Tier, Checked: facts.allowed[t.Tier][ref],
			})
		}
		rows = append(rows, row)
	}
	return rows
}

// unofferedWords says why a listed model is not among a provider's.
func unofferedWords(ref string, available map[string]bool) string {
	instance, _, _ := strings.Cut(ref, "/")
	up, known := available[instance]
	switch {
	case known && !up:
		return "Its provider is not answering now."
	case known:
		return "Its provider does not list it."
	}
	return "No model provider has that name."
}

// routeOf builds a role's route card; inUse says the role may run now.
func routeOf(role *roleInfo, set *modelsv1.GetModelSettingsResponse, facts *routeFacts, inUse bool) routeRow {
	row := routeRow{Role: role.Role, Name: role.Name, Does: role.Does, On: inUse, Action: modelsPath + "/routes/" + role.Role}
	var refs []string
	for _, r := range set.GetRoutes() {
		if r.GetRole() == role.Role {
			refs = r.GetModels()
		}
	}
	ready := false
	for i, ref := range refs {
		e := routeEntryOf(role, ref, i, len(refs), facts)
		ready = ready || (len(e.Missing) == 0 && e.Problem != problemBroken)
		row.Entries = append(row.Entries, e)
	}
	if len(refs) > 0 && !ready {
		row.Warning = "No model of this route may receive " + strings.Join(role.Needs, " and ") + ", which this role sends: its requests are refused until you allow one below."
	}
	for ref := range facts.offered {
		if !slices.Contains(refs, ref) {
			row.Choices = append(row.Choices, ref)
		}
	}
	slices.Sort(row.Choices)
	return row
}

// routeEntryOf describes the model at index of a route of n models.
func routeEntryOf(role *roleInfo, ref string, index, n int, facts *routeFacts) routeEntry {
	e := routeEntry{Position: "Fallback " + strconv.Itoa(index), Ref: ref, Index: index, First: index == 0, Last: index == n-1}
	if index == 0 {
		e.Position = "Primary"
	}
	p := routeProblem(role.Role, ref, facts)
	e.Problem, e.ProblemWords = p.Kind, p.Words
	for _, tier := range role.Needs {
		if !facts.allowed[tier][ref] {
			e.Missing = append(e.Missing, tier)
		}
	}
	return e
}

// routeProblem says what is wrong with a route entry: problemBroken for the
// router's model-route-broken state (no plugin by that name, or one that is no
// model provider), problemDown for a provider that does not answer now,
// problemUnlisted for a model its provider does not list; empty when nothing
// is, with its words.
func routeProblem(role, ref string, facts *routeFacts) problem {
	switch facts.broken[role+" "+ref] {
	case reasonMissing:
		return problem{problemBroken, "Broken: no plugin has that name."}
	case reasonNotProvider:
		return problem{problemBroken, "Broken: that plugin is not a model provider."}
	}
	instance, _, _ := strings.Cut(ref, "/")
	up, known := facts.available[instance]
	switch {
	case known && !up:
		return problem{problemDown, "Its provider is not answering now."}
	case known && !facts.offered[ref]:
		return problem{problemUnlisted, "Its provider does not list this model."}
	}
	return problem{}
}

// problem is what is wrong with a route entry: one of the problem constants,
// and the words that say it.
type problem struct{ Kind, Words string }

// setModelsSwitch turns the one switch on or off.
func (s *server) setModelsSwitch(w http.ResponseWriter, r *http.Request, sess session) {
	on, err := postedOn(r)
	if err != nil || s.modelSettings == nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	_, err = s.modelSettings.SetAgentsEnabled(r.Context(), connect.NewRequest(&modelsv1.SetAgentsEnabledRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Enabled: on,
	}))
	s.afterModelsForm(w, r, sess, "SetAgentsEnabled", "switch", err)
}

// setModelsFeature turns one feature's toggle on or off.
func (s *server) setModelsFeature(w http.ResponseWriter, r *http.Request, sess session) {
	on, err := postedOn(r)
	feature := r.PathValue("feature")
	if err != nil || s.modelSettings == nil || !slices.ContainsFunc(modelFeatures(), func(f featureInfo) bool { return f.Feature == feature }) {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	_, err = s.modelSettings.SetFeatureEnabled(r.Context(), connect.NewRequest(&modelsv1.SetFeatureEnabledRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Feature: feature, Enabled: on,
	}))
	s.afterModelsForm(w, r, sess, "SetFeatureEnabled", "feature", err)
}

// postedOn reads the posted state of a switch: "1" for on, "0" for off.
func postedOn(r *http.Request) (bool, error) {
	if err := r.ParseForm(); err != nil {
		return false, errBadForm
	}
	switch r.PostForm.Get("enabled") {
	case "1":
		return true, nil
	case "0":
		return false, nil
	}
	return false, errBadForm
}

// postedTiers reads the grid's ticked boxes, by tier, in the order posted.
func postedTiers(r *http.Request) (map[string][]string, error) {
	if err := r.ParseForm(); err != nil {
		return nil, errBadForm
	}
	posted := map[string][]string{}
	for _, v := range r.PostForm["allow"] {
		tier, ref, found := strings.Cut(v, "|")
		if !found || !slices.ContainsFunc(modelTiers(), func(t tierRow) bool { return t.Tier == tier }) {
			return nil, errBadForm
		}
		posted[tier] = append(posted[tier], ref)
	}
	return posted, nil
}

// setModelTiers saves the tier grid: for each tier whose ticked models differ
// from the core's, one SetTierModels. A tier nothing is ticked for is cleared,
// which denies it to every model (D-206).
func (s *server) setModelTiers(w http.ResponseWriter, r *http.Request, sess session) {
	posted, err := postedTiers(r)
	if err != nil || s.modelSettings == nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	current, err := s.modelSettings.GetModelSettings(r.Context(), connect.NewRequest(&modelsv1.GetModelSettingsRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	if err != nil {
		s.coreFailed(w, r, "GetModelSettings", err)
		return
	}
	have := map[string][]string{}
	for _, t := range current.Msg.GetTiers() {
		have[t.GetTier()] = t.GetModels()
	}
	for _, t := range modelTiers() {
		want := posted[t.Tier]
		if sameSet(want, have[t.Tier]) {
			continue
		}
		_, err := s.modelSettings.SetTierModels(r.Context(), connect.NewRequest(&modelsv1.SetTierModelsRequest{
			Origin: sess.Origin, SessionToken: sess.Token, Tier: t.Tier, Models: want,
		}))
		if err != nil {
			s.afterModelsForm(w, r, sess, "SetTierModels", "tiers", err)
			return
		}
	}
	http.Redirect(w, r, modelsPath+"?"+url.Values{"saved": {"tiers"}}.Encode()+"#tiers-title", http.StatusSeeOther)
}

// sameSet reports whether a and b hold the same strings, in any order.
func sameSet(a, b []string) bool {
	return len(a) == len(b) && slices.Equal(slices.Sorted(slices.Values(a)), slices.Sorted(slices.Values(b)))
}

// changeModelRoute adds a model to a role's route, removes one or moves one up
// or down, and hands the whole route to the core.
func (s *server) changeModelRoute(w http.ResponseWriter, r *http.Request, sess session) {
	role := r.PathValue("role")
	if err := r.ParseForm(); err != nil || s.modelSettings == nil || !slices.ContainsFunc(modelRoles(), func(x roleInfo) bool { return x.Role == role }) {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	current, err := s.modelSettings.GetModelSettings(r.Context(), connect.NewRequest(&modelsv1.GetModelSettingsRequest{Origin: sess.Origin, SessionToken: sess.Token}))
	if err != nil {
		s.coreFailed(w, r, "GetModelSettings", err)
		return
	}
	var route []string
	for _, rt := range current.Msg.GetRoutes() {
		if rt.GetRole() == role {
			route = slices.Clone(rt.GetModels())
		}
	}
	next, ok := editRoute(route, r.PostForm)
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	_, err = s.modelSettings.SetRoute(r.Context(), connect.NewRequest(&modelsv1.SetRouteRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Role: role, Models: next,
	}))
	s.afterModelsForm(w, r, sess, "SetRoute", "route", err)
}

// editRoute applies the posted operation to route: "add" the model of the
// "model" field, or of "ref" when "model" is empty, "remove" the entry
// "index", "up" or "down" to move it. ok is false for anything else.
func editRoute(route []string, f url.Values) ([]string, bool) {
	if f.Get("op") == "add" {
		ref := strings.TrimSpace(f.Get("model"))
		if ref == "" {
			ref = strings.TrimSpace(f.Get("ref"))
		}
		return append(route, ref), ref != ""
	}
	index, err := strconv.Atoi(f.Get("index"))
	if err != nil || index < 0 || index >= len(route) {
		return nil, false
	}
	switch f.Get("op") {
	case "remove":
		return slices.Delete(route, index, index+1), true
	case "up":
		return swapEntries(route, index-1, index)
	case "down":
		return swapEntries(route, index, index+1)
	}
	return nil, false
}

// swapEntries swaps the entries at i and j, which must both be in route.
func swapEntries(route []string, i, j int) ([]string, bool) {
	if i < 0 || j >= len(route) {
		return nil, false
	}
	route[i], route[j] = route[j], route[i]
	return route, true
}

// afterModelsForm answers a change of the settings: back to the page saying it
// was saved, the page again with the core's reason when it refused the change,
// and anything else as a failed call.
func (s *server) afterModelsForm(w http.ResponseWriter, r *http.Request, sess session, method, saved string, err error) {
	if err == nil {
		http.Redirect(w, r, modelsPath+"?"+url.Values{"saved": {saved}}.Encode(), http.StatusSeeOther)
		return
	}
	if status, refused := refusalStatus(connect.CodeOf(err)); refused && connect.CodeOf(err) != connect.CodePermissionDenied {
		s.renderModels(w, r, sess, status, "", "The core refused this: "+connectMessage(err))
		return
	}
	s.coreFailed(w, r, method, err)
}
