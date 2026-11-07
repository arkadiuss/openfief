package ui

import (
	"net/http"
	"net/url"
	"strings"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// assetTreasuryBond is the read API's asset class of a treasury bond
// series.
const assetTreasuryBond = "treasury_bond"

// The read API's sources of a series' terms (D-97).
const (
	termsPublished = "published"
	termsEntered   = "entered"
	feeTable       = "table"
	feeAssumed     = "assumed"
)

// seriesTerms is what a treasury bond's page shows of its series' stored
// terms (D-97): each term, written in the reader's number format, with
// where it came from, and the published data's attribution line, which its
// licence asks for (CC BY 4.0).
type seriesTerms struct {
	Series                string
	FirstRate, Margin     string
	Fee                   string
	RatesFrom, FeeFrom    string
	FeeAssumed, Published bool
	Attribution           string
}

// termSourceLabel names where a term came from: the ministry's data, the
// fee table, the table's latest fee carried forward, or the owner.
func termSourceLabel(source string) string {
	switch source {
	case termsPublished:
		return "Published: Ministry of Finance data"
	case feeTable:
		return "From the fee table of the issue letters"
	case feeAssumed:
		return "Assumed from the latest reviewed letter"
	case termsEntered:
		return "Entered by the owner"
	}
	return source
}

// seriesTermsView is the view of t, in f; nil for none.
func seriesTermsView(t *viewsv1.BondSeriesTerms, f numberFormat) *seriesTerms {
	if t == nil {
		return nil
	}
	return &seriesTerms{
		Series: t.GetSeries(), FirstRate: f.percent(t.GetFirstPeriodRate()), Margin: f.percent(t.GetMargin()),
		Fee: f.money(t.GetEarlyRedemptionFee()), RatesFrom: termSourceLabel(t.GetRatesSource()), FeeFrom: termSourceLabel(t.GetFeeSource()),
		FeeAssumed: t.GetFeeSource() == feeAssumed, Published: t.GetRatesSource() == termsPublished,
		Attribution: attributionLine(t.GetAttribution()),
	}
}

// percent writes a decimal fraction as a percentage in f: "0.0611" is
// "6.11%", or "6,11%" with a decimal comma; text that isn't a decimal
// gives "".
func (f numberFormat) percent(fraction string) string {
	p := percent(fraction)
	if p == "" {
		return ""
	}
	return f.amount(strings.TrimSuffix(p, "%")) + "%"
}

// attributionLine is the line the ministry's licence asks for under terms
// read from its data (D-97): "Source: Ministerstwo Finansów, <dataset>,
// CC BY 4.0, made <date>, fetched <date>, processed by OpenFief"; "" for
// none.
func attributionLine(a *viewsv1.BondTermsAttribution) string {
	if a == nil {
		return ""
	}
	fetched, _, _ := strings.Cut(a.GetFetched(), "T")
	return "Source: " + a.GetPublisher() + ", " + a.GetDataset() + ", " + a.GetLicense() + ", made " + a.GetMade() +
		", fetched " + fetched + ", processed by OpenFief"
}

// setBondFee hands the owner's confirmation of a series' assumed
// early-redemption fee, or the fee that overrides it, to the core
// (SetEarlyRedemptionFee, D-97), then shows the instrument's page again; a
// refusal shows the core's reason. The form posts "fee", empty to confirm.
func (s *server) setBondFee(w http.ResponseWriter, r *http.Request, sess session) {
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	fields, ok := postedOnce(r, "fee")
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	fee, isin := fields[0], r.PathValue("isin")
	if fee != "" {
		fee = strings.ReplaceAll(fee, ",", ".")
		if _, err := kernel.ParseDecimal(fee); err != nil {
			s.renderInstrument(w, r, sess, isin, http.StatusBadRequest, instrumentForm{Error: "A fee is an amount in PLN, such as 3.00."})
			return
		}
	}
	_, err := s.settings.SetEarlyRedemptionFee(r.Context(), connect.NewRequest(&valuationv1.SetEarlyRedemptionFeeRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Isin: isin, Fee: fee,
	}))
	if err == nil {
		http.Redirect(w, r, "/instruments/"+url.PathEscape(isin)+"?fee=1#terms-title", http.StatusSeeOther)
		return
	}
	code := connect.CodeOf(err)
	if code == connect.CodeFailedPrecondition {
		s.renderInstrument(w, r, sess, isin, http.StatusConflict, instrumentForm{Error: "The core refused this: " + connectMessage(err)})
		return
	}
	s.afterInstrumentForm(w, r, sess, isin, "SetEarlyRedemptionFee", err)
}
