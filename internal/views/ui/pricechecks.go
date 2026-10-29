package ui

import (
	"fmt"
	"net/http"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// priceChecksView is what the price checks page shows: the threshold as a
// percentage, empty when the core compares no prices, and the flagged
// checks.
type priceChecksView struct {
	Threshold string
	Checks    []*viewsv1.PriceCheck
}

// priceChecksPage lists every flagged price check: a day on which an
// institution's own price of an instrument, such as IBKR's mark price,
// differs from the price holdings are valued at by more than the threshold
// (D-22). The page only shows them; nothing changes a price from here.
func (s *server) priceChecksPage(w http.ResponseWriter, r *http.Request, sess session) {
	resp, err := s.read.ListPriceChecks(r.Context(), bearer(&viewsv1.ListPriceChecksRequest{
		Status: viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_FLAGGED,
	}, sess))
	if err != nil {
		s.coreFailed(w, r, "ListPriceChecks", err)
		return
	}
	view := priceChecksView{Threshold: percent(resp.Msg.GetThreshold()), Checks: resp.Msg.GetChecks()}
	s.show(w, r, http.StatusOK, "price-checks", page{Title: "Flagged price checks", Session: &sess, Data: view})
}

// flaggedPriceChecks counts the flagged price checks, for the overview.
func (s *server) flaggedPriceChecks(r *http.Request, sess session) (int, error) {
	resp, err := s.read.ListPriceChecks(r.Context(), bearer(&viewsv1.ListPriceChecksRequest{
		Status: viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_FLAGGED,
	}, sess))
	if err != nil {
		return 0, fmt.Errorf("ListPriceChecks: %w", err)
	}
	return len(resp.Msg.GetChecks()), nil
}
