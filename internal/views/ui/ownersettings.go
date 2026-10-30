package ui

import (
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
)

// formatSample is the amount the number format choices are shown with.
const formatSample = "5000.00"

// formatChoice is one number format the owner can pick, written out.
type formatChoice struct {
	Key     string
	Example string
	Chosen  bool
}

// valuesSettings is the settings page's section on values (D-78, D-113):
// the reporting currency, the number formats with the owner's or the
// pack's default chosen, and the owner's marginal PIT rate in percent,
// empty until set.
type valuesSettings struct {
	ReportingCurrency string
	Formats           []formatChoice
	FormatDefault     bool
	MarginalRate      string
}

// loadValuesSettings reads the owner's settings (GetOwnerSettings).
func (s *server) loadValuesSettings(r *http.Request, sess session) (*valuesSettings, error) {
	resp, err := s.settings.GetOwnerSettings(r.Context(), connect.NewRequest(&valuationv1.GetOwnerSettingsRequest{
		Origin: sess.Origin, SessionToken: sess.Token,
	}))
	if err != nil {
		return nil, fmt.Errorf("GetOwnerSettings: %w", err)
	}
	m := resp.Msg
	v := &valuesSettings{
		ReportingCurrency: m.GetReportingCurrency(), FormatDefault: m.GetNumberFormatDefault(),
		MarginalRate: toPercent(m.GetMarginalRate()),
	}
	for _, nf := range m.GetNumberFormats() {
		f := separators(nf.GetGroupSeparator(), nf.GetDecimalSeparator())
		v.Formats = append(v.Formats, formatChoice{
			Key: nf.GetKey(), Example: f.amount(formatSample) + " " + m.GetReportingCurrency(),
			Chosen: nf.GetKey() == m.GetNumberFormat().GetKey(),
		})
	}
	return v, nil
}

// setNumberFormat makes the posted number format the owner's
// (SetNumberFormat); every page then writes amounts in it (D-113).
func (s *server) setNumberFormat(w http.ResponseWriter, r *http.Request, sess session) {
	_, err := s.settings.SetNumberFormat(r.Context(), connect.NewRequest(&valuationv1.SetNumberFormatRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Key: r.PostFormValue("number_format"),
	}))
	if err != nil {
		s.actionFailed(w, r, sess, "SetNumberFormat", err)
		return
	}
	http.Redirect(w, r, "/settings#values-title", http.StatusSeeOther)
}

// setMarginalRate sets the owner's marginal PIT rate, typed in percent,
// which an IKZE's liquidation value needs (D-78); the core revalues every
// account.
func (s *server) setMarginalRate(w http.ResponseWriter, r *http.Request, sess session) {
	_, err := s.settings.SetMarginalRate(r.Context(), connect.NewRequest(&valuationv1.SetMarginalRateRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Rate: fromPercent(strings.TrimSpace(r.PostFormValue("marginal_rate"))),
	}))
	if err != nil {
		s.actionFailed(w, r, sess, "SetMarginalRate", err)
		return
	}
	http.Redirect(w, r, "/settings#values-title", http.StatusSeeOther)
}
