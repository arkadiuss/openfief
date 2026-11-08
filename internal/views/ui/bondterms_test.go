package ui

import (
	"cmp"
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"connectrpc.com/connect"
	gocmp "github.com/google/go-cmp/cmp"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The synthetic treasury bonds: one whose fee is assumed, and one whose
// terms aren't stored yet.
const (
	assumedBond = "PLEDO1137006"
	pendingBond = "PLCOI1130004"
)

// errNotAssumedFake is the fake core's refusal to confirm a fee that isn't
// assumed.
var errNotAssumedFake = errors.New("the series' early-redemption fee is not assumed")

// bondDetail is the synthetic bond with ISIN isin: EDO1137, with published
// rates and an assumed fee, or COI1130, without terms and under its
// provisional ISIN (D-214).
func bondDetail(isin string) *viewsv1.InstrumentDetail {
	in := &viewsv1.Instrument{Isin: isin, Symbol: "EDO1137"}
	out := &viewsv1.InstrumentDetail{Isin: isin, AssetClass: assetTreasuryBond, Listings: []*viewsv1.InstrumentListing{{Instrument: in}}}
	if isin == pendingBond {
		in.Symbol, in.IsinProvisional, out.IsinProvisional = "COI1130", true, true
		return out
	}
	out.BondTerms = &viewsv1.BondSeriesTerms{
		Series: "EDO1137", FirstPeriodRate: "0.0611", Margin: "0.02", RatesSource: termsPublished,
		EarlyRedemptionFee: pln("3.00"), FeeSource: feeAssumed,
		Attribution: &viewsv1.BondTermsAttribution{
			Publisher: "Ministerstwo Finansów", Dataset: "Dane dotyczące obligacji detalicznych", License: "CC BY 4.0",
			Url: "https://www.example.org/attachment/x", Made: "2026-10-26", Fetched: "2026-11-07T06:00:00Z", Processing: "read",
		},
	}
	return out
}

// SetEarlyRedemptionFee records the owner's fee by ISIN, "confirmed" for
// none, and refuses to confirm the fee of a bond without an assumed one.
func (f *fakeSettings) SetEarlyRedemptionFee(_ context.Context, r *connect.Request[valuationv1.SetEarlyRedemptionFeeRequest]) (*connect.Response[valuationv1.SetEarlyRedemptionFeeResponse], error) {
	if r.Msg.GetSessionToken() != ownerSession {
		return nil, codeErr(connect.CodePermissionDenied, errNotOwnerFake)
	}
	if r.Msg.GetIsin() != assumedBond && r.Msg.GetFee() == "" {
		return nil, codeErr(connect.CodeFailedPrecondition, errNotAssumedFake)
	}
	f.locked(func() {
		if f.fees == nil {
			f.fees = map[string]string{}
		}
		f.fees[r.Msg.GetIsin()] = cmp.Or(r.Msg.GetFee(), "confirmed")
	})
	return connect.NewResponse(&valuationv1.SetEarlyRedemptionFeeResponse{}), nil
}

// TestBondTermsOnInstrumentPage shows a bond's stored terms in the
// reader's number format with where each came from and the published
// data's attribution line, the assumed fee marked to confirm, and the fee
// form to the owner only; a bond without terms says the next sync fills
// them in (D-97).
func TestBondTermsOnInstrumentPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/instruments/"+assumedBond, ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `id="terms-title">Series terms`, "Term of EDO1137", "6,11%", "2%", "3,00 PLN",
		"Published: Ministry of Finance data", "Assumed from the latest reviewed letter", `<span class="pill pill-warn">to confirm</span>`,
		"Source: Ministerstwo Finansów, Dane dotyczące obligacji detalicznych, CC BY 4.0, made 2026-10-26, fetched 2026-11-07, processed by OpenFief")
	wantNoBody(t, rec, `action="/instruments/`+assumedBond+`/fee"`)
	rec = h.serve(t, withSession(request(http.MethodGet, "/instruments/"+assumedBond, ""), ownerSession))
	wantBody(t, rec, `action="/instruments/`+assumedBond+`/fee"`, "Leave it empty to confirm 3,00 PLN", "Confirm or save the fee")
	rec = h.serve(t, withSession(request(http.MethodGet, "/instruments/"+pendingBond, ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "No terms are stored for this series yet", "the next sync")
	rec = h.serve(t, withSession(request(http.MethodGet, "/instruments/XX0000000001", ""), readSession))
	wantNoBody(t, rec, `id="terms-title"`)
}

// TestSetBondFee confirms an assumed fee with an empty field and overrides
// one with a decimal comma; a fee that isn't a number, a refusal, a
// malformed form and a read session change nothing.
func TestSetBondFee(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	post := func(isin, body, token string) *http.Request {
		return withSession(form("/instruments/"+isin+"/fee", body), token)
	}
	rec := h.serve(t, post(assumedBond, "fee=", ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if loc := rec.Header().Get("Location"); loc != "/instruments/"+assumedBond+"?fee=1#terms-title" {
		t.Errorf("Location = %q", loc)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/instruments/"+assumedBond+"?fee=1", ""), ownerSession))
	wantBody(t, rec, "The early-redemption fee is saved as yours.")
	rec = h.serve(t, post(pendingBond, url.Values{"fee": {" 2,50 "}}.Encode(), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	rec = h.serve(t, post(pendingBond, "fee=", ownerSession))
	wantStatus(t, rec, http.StatusConflict)
	wantBody(t, rec, "The core refused this: the series&#39; early-redemption fee is not assumed")
	rec = h.serve(t, post(assumedBond, "fee=three", ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "A fee is an amount in PLN")
	for _, body := range []string{"fee=1&fee=2", "other=x"} {
		wantStatus(t, h.serve(t, post(assumedBond, body, ownerSession)), http.StatusBadRequest)
	}
	wantStatus(t, h.serve(t, post(assumedBond, "fee=", readSession)), http.StatusForbidden)
	var got map[string]string
	h.settings.locked(func() { got = h.settings.fees })
	if diff := gocmp.Diff(map[string]string{assumedBond: "confirmed", pendingBond: "2.50"}, got); diff != "" {
		t.Errorf("fees (-want +got):\n%s", diff)
	}
}

// TestBondFeeItem titles an assumed fee's item by its ISIN and links its
// instrument.
func TestBondFeeItem(t *testing.T) {
	t.Parallel()
	it := &viewsv1.AttentionItem{Kind: kindBondFeeAssumed, Subject: assumedBond, Isin: assumedBond}
	if got := itemTitle(it); got != "Confirm the early-redemption fee of "+assumedBond {
		t.Errorf("itemTitle = %q", got)
	}
	if got := attentionKindOf(kindBondFeeAssumed).Label; got != "Bond fee to confirm" {
		t.Errorf("label = %q", got)
	}
	links := itemLinks(it)
	if len(links) != 1 || links[0].Href != instrumentPath(assumedBond) {
		t.Errorf("links = %+v, want the instrument's page", links)
	}
}
