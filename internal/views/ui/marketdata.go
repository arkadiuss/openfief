package ui

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"connectrpc.com/connect"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
)

// maxRankedSources bounds the sources a ranking form may post.
const maxRankedSources = 64

// kindName is how the settings page names a kind of market data: its
// label and its form value.
type kindName struct{ Label, Value string }

// rankingKind returns the name of a kind of market data for the settings
// page, and false for an unspecified kind.
func rankingKind(k valuationv1.MarketDataKind) (kindName, bool) {
	switch k {
	case valuationv1.MarketDataKind_MARKET_DATA_KIND_PRICES:
		return kindName{"Prices", "prices"}, true
	case valuationv1.MarketDataKind_MARKET_DATA_KIND_FX_RATES:
		return kindName{"Exchange rates", "fx_rates"}, true
	case valuationv1.MarketDataKind_MARKET_DATA_KIND_INDEX_READINGS:
		return kindName{"Index readings (CPI, NBP reference rate)", "index_readings"}, true
	case valuationv1.MarketDataKind_MARKET_DATA_KIND_UNSPECIFIED:
	}
	return kindName{}, false
}

// rankingView is one kind's ranking on the settings page, with its gap
// threshold in days (D-94).
type rankingView struct {
	Label        string
	Kind         string
	Sources      []*valuationv1.RankedSource
	GapThreshold int32
}

// seriesView is one series' source on the settings page (D-94): the
// kind's label and form value, the series' key, its source, whether the
// owner chose it, and every source with a value of it.
type seriesView struct {
	Label, Kind, Key, Source string
	ByOwner                  bool
	Available                []string
}

// Names returns the sources' names in their order, for the form.
func (v *rankingView) Names() []string {
	out := make([]string, len(v.Sources))
	for i, s := range v.Sources {
		out[i] = s.GetName()
	}
	return out
}

// loadRankings reads every kind's ranking of market-data sources.
func (s *server) loadRankings(r *http.Request, sess session) ([]rankingView, error) {
	resp, err := s.marketData.ListSourceRankings(r.Context(), connect.NewRequest(&valuationv1.ListSourceRankingsRequest{
		Origin: sess.Origin, SessionToken: sess.Token,
	}))
	if err != nil {
		return nil, fmt.Errorf("ListSourceRankings: %w", err)
	}
	out := make([]rankingView, 0, len(resp.Msg.GetRankings()))
	for _, rk := range resp.Msg.GetRankings() {
		name, ok := rankingKind(rk.GetKind())
		if !ok {
			continue
		}
		out = append(out, rankingView{Label: name.Label, Kind: name.Value, Sources: rk.GetSources(), GapThreshold: rk.GetGapThresholdDays()})
	}
	return out, nil
}

// loadSeries reads every series that has a source.
func (s *server) loadSeries(r *http.Request, sess session) ([]seriesView, error) {
	resp, err := s.marketData.ListSeriesSources(r.Context(), connect.NewRequest(&valuationv1.ListSeriesSourcesRequest{
		Origin: sess.Origin, SessionToken: sess.Token,
	}))
	if err != nil {
		return nil, fmt.Errorf("ListSeriesSources: %w", err)
	}
	out := make([]seriesView, 0, len(resp.Msg.GetSeries()))
	for _, ss := range resp.Msg.GetSeries() {
		name, ok := rankingKind(ss.GetKind())
		if !ok {
			continue
		}
		out = append(out, seriesView{
			Label: name.Label, Kind: name.Value, Key: ss.GetKey(), Source: ss.GetSource(), ByOwner: ss.GetByOwner(), Available: ss.GetAvailable(),
		})
	}
	return out, nil
}

// setSeriesSource makes the posted source the source of the posted series
// with the core's MarketDataService (D-94); the core revalues what it
// affects.
func (s *server) setSeriesSource(w http.ResponseWriter, r *http.Request, sess session) {
	kind, ok := kindOf(r.PostFormValue("kind"))
	if !ok {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	_, err := s.marketData.SetSeriesSource(r.Context(), connect.NewRequest(&valuationv1.SetSeriesSourceRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Kind: kind, Key: r.PostFormValue("key"), Source: r.PostFormValue("source"),
	}))
	if err != nil {
		s.actionFailed(w, r, sess, "SetSeriesSource", err)
		return
	}
	http.Redirect(w, r, "/settings#series-sources-title", http.StatusSeeOther)
}

// setGapThreshold sets the posted kind's gap threshold, a whole number of
// days, with the core's MarketDataService (D-94).
func (s *server) setGapThreshold(w http.ResponseWriter, r *http.Request, sess session) {
	kind, ok := kindOf(r.PostFormValue("kind"))
	days, err := strconv.ParseInt(r.PostFormValue("days"), 10, 32)
	if !ok || err != nil {
		s.renderSettings(w, r, sess, http.StatusBadRequest, &settingsData{Error: "The gap threshold must be a whole number of days."})
		return
	}
	_, err = s.marketData.SetGapThreshold(r.Context(), connect.NewRequest(&valuationv1.SetGapThresholdRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Kind: kind, Days: int32(days),
	}))
	if err != nil {
		s.actionFailed(w, r, sess, "SetGapThreshold", err)
		return
	}
	http.Redirect(w, r, "/settings#market-data-title", http.StatusSeeOther)
}

// rankSources moves one source of a kind up or down, or puts the kind's
// sources back in their default places, and saves the order with the
// core's MarketDataService (D-94). The form posts the kind, the sources in
// their current order, and the button pressed: up or down with a source's
// name, or reset.
func (s *server) rankSources(w http.ResponseWriter, r *http.Request, sess session) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	kind, ok := kindOf(r.PostFormValue("kind"))
	order := r.PostForm["source"]
	if !ok || len(order) > maxRankedSources {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	var placed []string
	switch {
	case r.PostFormValue("reset") != "":
	case r.PostFormValue("up") != "":
		placed = moveSource(order, r.PostFormValue("up"), -1)
	case r.PostFormValue("down") != "":
		placed = moveSource(order, r.PostFormValue("down"), 1)
	default:
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	_, err := s.marketData.SetSourceRanking(r.Context(), connect.NewRequest(&valuationv1.SetSourceRankingRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Kind: kind, Sources: placed,
	}))
	if err != nil {
		s.actionFailed(w, r, sess, "SetSourceRanking", err)
		return
	}
	http.Redirect(w, r, "/settings#market-data-title", http.StatusSeeOther)
}

// kindOf is the kind a form names by its value.
func kindOf(value string) (valuationv1.MarketDataKind, bool) {
	for _, k := range []valuationv1.MarketDataKind{
		valuationv1.MarketDataKind_MARKET_DATA_KIND_PRICES, valuationv1.MarketDataKind_MARKET_DATA_KIND_FX_RATES,
		valuationv1.MarketDataKind_MARKET_DATA_KIND_INDEX_READINGS,
	} {
		if name, _ := rankingKind(k); name.Value == value {
			return k, true
		}
	}
	return valuationv1.MarketDataKind_MARKET_DATA_KIND_UNSPECIFIED, false
}

// moveSource returns order with name moved by step places, -1 up or 1
// down, every source then placed; a source at the end it moves toward, or
// not in order, stays where it is.
func moveSource(order []string, name string, step int) []string {
	out := slices.Clone(order)
	i := slices.Index(out, name)
	j := i + step
	if i < 0 || j < 0 || j >= len(out) {
		return out
	}
	out[i], out[j] = out[j], out[i]
	return out
}
