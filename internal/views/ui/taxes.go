package ui

import (
	"fmt"
	"net/http"
	"slices"

	"connectrpc.com/connect"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// instrumentClass is one instrument on the classification form (D-132):
// its ISIN, and the owner's class, empty until the owner classifies it.
type instrumentClass struct {
	ISIN       string
	Class      string
	Classified bool
}

// taxesSettings is the settings page's section on yearly taxes (D-127,
// D-132): this year's tax of each wrapper that has one, the classes an
// instrument can be in, and the instruments to classify: those a tax
// counted without a class first, then the owner's, in ISIN order.
type taxesSettings struct {
	Year        int
	Taxes       []*viewsv1.YearlyTax
	Classes     []string
	Instruments []instrumentClass
}

// loadTaxes reads this year's yearly taxes (ListYearlyTaxes) and the
// owner's classifications (GetOwnerSettings); with no classes, when no
// wrapper has a yearly tax, nothing else.
func (s *server) loadTaxes(r *http.Request, sess session) (*taxesSettings, error) {
	settings, err := s.settings.GetOwnerSettings(r.Context(), connect.NewRequest(&valuationv1.GetOwnerSettingsRequest{
		Origin: sess.Origin, SessionToken: sess.Token,
	}))
	if err != nil {
		return nil, fmt.Errorf("GetOwnerSettings: %w", err)
	}
	if len(settings.Msg.GetExemptionClasses()) == 0 {
		return &taxesSettings{}, nil
	}
	year := s.today().Year()
	taxes, err := s.read.ListYearlyTaxes(r.Context(), bearer(&viewsv1.ListYearlyTaxesRequest{Year: clampYear(year)}, sess))
	if err != nil {
		return nil, fmt.Errorf("ListYearlyTaxes: %w", err)
	}
	out := &taxesSettings{Year: year, Taxes: taxes.Msg.GetTaxes(), Classes: settings.Msg.GetExemptionClasses()}
	var unclassified []string
	for _, t := range out.Taxes {
		unclassified = append(unclassified, t.GetUnclassifiedInstruments()...)
	}
	slices.Sort(unclassified)
	for _, isin := range slices.Compact(unclassified) {
		out.Instruments = append(out.Instruments, instrumentClass{ISIN: isin})
	}
	for _, c := range settings.Msg.GetInstrumentClasses() {
		out.Instruments = append(out.Instruments, instrumentClass{ISIN: c.GetIsin(), Class: c.GetExemptionClass(), Classified: true})
	}
	return out, nil
}

// classLabel returns an exemption class's label: the Polish pack's three for
// OKI (D-127), any other key as it is.
func classLabel(key string) string {
	labels := map[string]string{"savings": "Savings assets", "investment": "Investment assets", "other": "Other assets"}
	if l, ok := labels[key]; ok {
		return l
	}
	return key
}

// clampYear returns year as an int32 for the request, within the years a
// date can have.
func clampYear(year int) int32 {
	const firstYear, lastYear = 1, 9999
	switch {
	case year < firstYear:
		return firstYear
	case year > lastYear:
		return lastYear
	}
	return int32(year)
}

// setExemptionClass classifies the posted instrument for a yearly tax's
// exemptions (SetExemptionClass, D-132); it changes no stored valuation.
func (s *server) setExemptionClass(w http.ResponseWriter, r *http.Request, sess session) {
	_, err := s.settings.SetExemptionClass(r.Context(), connect.NewRequest(&valuationv1.SetExemptionClassRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Isin: r.PostFormValue("isin"), ExemptionClass: r.PostFormValue("exemption_class"),
	}))
	if err != nil {
		s.actionFailed(w, r, sess, "SetExemptionClass", err)
		return
	}
	http.Redirect(w, r, "/settings#taxes-title", http.StatusSeeOther)
}
