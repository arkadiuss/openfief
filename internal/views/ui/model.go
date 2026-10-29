package ui

import (
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
)

const (
	// blankTargetRows is how many empty target rows the settings form
	// offers below the stored targets.
	blankTargetRows = 3
	// maxTargetRows bounds the rows a form may post: the core's limit of
	// 50 targets, plus the blank ones.
	maxTargetRows = 50 + blankTargetRows
	// kindIndex is the form's value for an index target.
	kindIndex = "index"
	// kindInstrument is the form's value for an instrument target.
	kindInstrument = "instrument"
)

// targetRow is one target row of the model portfolio form. ID is an
// instrument's "ISIN/symbol" or an index's name.
type targetRow struct {
	Kind     string
	ID       string
	Currency string
	Weight   string
}

// modelForm is the model portfolio section of the settings page (D-52).
type modelForm struct {
	Enabled bool
	Rows    []targetRow
}

// withBlankRows returns f with blankTargetRows empty rows added.
func (f *modelForm) withBlankRows() *modelForm {
	for range blankTargetRows {
		f.Rows = append(f.Rows, targetRow{Kind: kindInstrument})
	}
	return f
}

// formOf turns a stored model portfolio into the form.
func formOf(m *portfoliov1.ModelPortfolio) *modelForm {
	f := &modelForm{Enabled: m.GetEnabled()}
	for _, t := range m.GetTargets() {
		row := targetRow{Kind: kindInstrument, ID: t.GetIsin() + "/" + t.GetSymbol(), Weight: t.GetWeightPercent()}
		if t.GetKind() == portfoliov1.TargetKind_TARGET_KIND_INDEX {
			row = targetRow{Kind: kindIndex, ID: t.GetIndex(), Currency: t.GetCurrency(), Weight: t.GetWeightPercent()}
		}
		f.Rows = append(f.Rows, row)
	}
	return f
}

// parseModelForm reads the posted form: the enabled checkbox and parallel
// kind, id, currency and weight fields, one per row. Rows with neither an
// ID nor a weight are left out; the core checks the rest (D-52).
func parseModelForm(r *http.Request) (*modelForm, bool) {
	if err := r.ParseForm(); err != nil {
		return nil, false
	}
	kinds, ids, currencies, weights := r.PostForm["kind"], r.PostForm["id"], r.PostForm["currency"], r.PostForm["weight"]
	n := len(ids)
	if n > maxTargetRows || len(kinds) != n || len(currencies) != n || len(weights) != n {
		return nil, false
	}
	f := &modelForm{Enabled: r.PostFormValue("enabled") == "on"}
	for i := range n {
		row := targetRow{Kind: kinds[i], ID: strings.TrimSpace(ids[i]), Currency: strings.ToUpper(strings.TrimSpace(currencies[i])), Weight: strings.TrimSpace(weights[i])}
		if row.ID == "" && row.Weight == "" {
			continue
		}
		f.Rows = append(f.Rows, row)
	}
	return f, true
}

// message turns the form into the model portfolio the core stores.
func (f *modelForm) message() *portfoliov1.ModelPortfolio {
	m := &portfoliov1.ModelPortfolio{Enabled: f.Enabled}
	for _, row := range f.Rows {
		t := &portfoliov1.Target{Kind: portfoliov1.TargetKind_TARGET_KIND_INSTRUMENT, WeightPercent: row.Weight}
		switch row.Kind {
		case kindIndex:
			t.Kind, t.Index, t.Currency = portfoliov1.TargetKind_TARGET_KIND_INDEX, row.ID, row.Currency
		case kindInstrument:
			t.Isin, t.Symbol, _ = strings.Cut(row.ID, "/")
		default:
			t.Kind = portfoliov1.TargetKind_TARGET_KIND_UNSPECIFIED
		}
		m.Targets = append(m.Targets, t)
	}
	return m
}

// loadModel reads the stored model portfolio, on or off, for the form.
func (s *server) loadModel(r *http.Request, sess session) (*modelForm, error) {
	resp, err := s.model.GetModelPortfolio(r.Context(), connect.NewRequest(&portfoliov1.GetModelPortfolioRequest{
		Origin: sess.Origin, SessionToken: sess.Token,
	}))
	if err != nil {
		return nil, fmt.Errorf("GetModelPortfolio: %w", err)
	}
	return formOf(resp.Msg.GetModel()), nil
}

// saveModel hands the posted model portfolio to the core
// (SaveModelPortfolio, D-52). A refusal, such as weights not summing to
// 100%, shows the form again as posted, with the core's reason.
func (s *server) saveModel(w http.ResponseWriter, r *http.Request, sess session) {
	f, ok := parseModelForm(r)
	if !ok {
		s.renderSettings(w, r, sess, http.StatusBadRequest, &settingsData{Error: "The model portfolio form is malformed."})
		return
	}
	_, err := s.model.SaveModelPortfolio(r.Context(), connect.NewRequest(&portfoliov1.SaveModelPortfolioRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Model: f.message(),
	}))
	if err != nil {
		if status, refused := refusalStatus(connect.CodeOf(err)); refused {
			s.renderSettings(w, r, sess, status, &settingsData{Error: "The model portfolio was not saved: " + connectMessage(err), Model: f.withBlankRows()})
			return
		}
		s.coreFailed(w, r, "SaveModelPortfolio", err)
		return
	}
	http.Redirect(w, r, "/settings#model-title", http.StatusSeeOther)
}
